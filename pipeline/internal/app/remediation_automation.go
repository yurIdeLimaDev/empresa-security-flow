package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type RemediationAutomationSummary struct {
	SchemaVersion       string `json:"schema_version"`
	CaseID              string `json:"case_id"`
	Status              string `json:"status"`
	Reason              string `json:"reason"`
	BestCommit          string `json:"best_commit"`
	BestBundleSHA256    string `json:"best_bundle_sha256"`
	TotalAttempts       int    `json:"total_attempts"`
	GlobalGatePassed    bool   `json:"global_gate_passed"`
	FinalApprovalStatus string `json:"final_approval_status"`
	DeliveryAuthorized  bool   `json:"delivery_authorized"`
	FinishedAt          string `json:"finished_at"`
}

// Operation locking also applies to the individual CLI commands. A second
// orchestrator or a manual command cannot mutate this case during automation.
func acquireRemediationOperation(dir string, cfg RemediationConfig) (func(), error) {
	release, err := acquireRemediationLock(dir)
	if err != nil {
		return nil, err
	}
	owner, err := os.ReadFile(filepath.Join(dir, "automation.lock"))
	if err != nil && !os.IsNotExist(err) {
		release()
		return nil, err
	}
	if err == nil && (cfg.automationToken == "" || string(owner) != cfg.automationToken) {
		release()
		return nil, fmt.Errorf("caso reservado por execução automática; não remova o lock sem confirmar que ela terminou")
	}
	return release, nil
}

func reserveRemediationAutomation(dir string) (string, func(), error) {
	release, err := acquireRemediationLock(dir)
	if err != nil {
		return "", nil, err
	}
	defer release()
	token, err := randomPatchID()
	if err != nil {
		return "", nil, err
	}
	path := filepath.Join(dir, "automation.lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", nil, fmt.Errorf("já existe automação ou lock pendente neste caso")
	}
	_, writeErr := f.WriteString(token)
	closeErr := f.Close()
	if writeErr != nil || closeErr != nil {
		return "", nil, fmt.Errorf("não foi possível persistir lock da automação")
	}
	return token, func() { _ = os.Remove(path) }, nil
}

// RunRemediationAutomation stops at the final human review. It never creates an
// approval, finalizes delivery, edits gates or falls back to the manual adapter.
func RunRemediationAutomation(ctx context.Context, configPath string, cfg RemediationConfig, planPath string) (string, error) {
	if cfg.Agent.Generation == nil {
		return "", fmt.Errorf("automação exige geração configurada; adaptador manual não é fallback")
	}
	readiness, err := CheckRemediationReadiness(configPath, cfg)
	if err != nil {
		return "", err
	}
	if !readiness.Approved {
		return "", fmt.Errorf("preflight bloqueou automação; configure o gateway e os gates antes de executar")
	}
	if planPath == "" {
		planPath, err = CreateRemediationPlan(configPath, cfg)
		if err != nil {
			return "", err
		}
	}
	plan, dir, err := readRemediationPlan(planPath)
	if err != nil {
		return "", err
	}
	if plan.ConfigSHA256 != HashJSON(cfg) || plan.CaseID != cfg.CaseID || plan.GateProfileSHA256 != remediationGateProfile(cfg) {
		return "", fmt.Errorf("configuração diverge do plano")
	}
	token, release, err := reserveRemediationAutomation(dir)
	if err != nil {
		return "", err
	}
	defer release()
	cfg.automationToken = token
	statePath := filepath.Join(dir, "state.json")
	finish := func(status, reason string, runErr error) (string, error) {
		var state RemediationState
		if err := ReadJSON(statePath, &state); err != nil {
			return "", err
		}
		summary := RemediationAutomationSummary{SchemaVersion: RemediationSchemaVersion, CaseID: cfg.CaseID, Status: status, Reason: reason, BestCommit: state.BestRef, BestBundleSHA256: state.BestBundleSHA256, TotalAttempts: state.TotalAttempts, GlobalGatePassed: state.GlobalGatePassed, FinalApprovalStatus: state.FinalApprovalStatus, DeliveryAuthorized: false, FinishedAt: now()}
		path := filepath.Join(dir, "automation-summary.json")
		if err := writeJSONAtomic(path, summary); err != nil {
			return "", err
		}
		return path, runErr
	}
	for {
		if err := ctx.Err(); err != nil {
			return finish("interrupted", "execução cancelada; BEST preservado", err)
		}
		var state RemediationState
		if err := ReadJSON(statePath, &state); err != nil {
			return "", err
		}
		if state.FinalApprovalStatus == "approved" || state.FinalApprovalStatus == "rejected" {
			return finish("closed", "caso já encerrado; nenhuma ação executada", nil)
		}
		if state.GlobalGatePassed {
			return finish("awaiting_final_review", "gates globais aprovados; revisão humana final obrigatória", nil)
		}
		var next *RemediationTicket
		allDone := true
		for _, ticket := range plan.Tickets {
			item := ticketState(&state, ticket.ID)
			if item == nil {
				return finish("blocked", "estado de ticket ausente", fmt.Errorf("estado incompleto"))
			}
			if item.Status == "accepted" || item.Status == "no_code_final_review" {
				continue
			}
			allDone = false
			if item.Status != "pending" && item.Status != "retry" && item.Status != "prepared" && item.Status != "candidate_ready" {
				continue
			}
			ready := true
			for _, id := range ticket.DependsOn {
				dep := ticketState(&state, id)
				if dep == nil || dep.Status != "accepted" {
					ready = false
				}
			}
			if ready {
				copy := ticket
				next = &copy
				break
			}
		}
		if allDone {
			if state.TotalAttempts >= cfg.Limits.MaxTotalAttempts {
				return finish("blocked", "sem orçamento para validação global", fmt.Errorf("limite global atingido"))
			}
			worktree := filepath.Join(dir, "worktrees", fmt.Sprintf("global-%d", time.Now().UnixNano()))
			unlock, err := acquireRemediationOperation(dir, cfg)
			if err != nil {
				return finish("blocked", "falha de lock", err)
			}
			_, err = gitOutput(plan.RepositoryPath, "worktree", "add", "--detach", worktree, state.BestRef)
			unlock()
			if err != nil {
				return finish("blocked", "não foi possível preparar BEST para os gates globais", err)
			}
			runPath, err := RunRemediationGates(ctx, cfg, planPath, "", "global", state.BestRef, worktree)
			if err != nil {
				return finish("blocked", "falha ao executar gates globais", err)
			}
			bundle, err := automationCandidateBundle(cfg, runPath)
			if err != nil {
				if recordErr := rejectUnverifiableAutomationCandidate(cfg, planPath, "", runPath); recordErr != nil {
					return finish("blocked", "falha ao registrar rejeição global", recordErr)
				}
				return finish("blocked", "bundle global ausente ou inválido", err)
			}
			evalPath, err := EvaluateRemediationCandidate(cfg, planPath, "", state.BestRef, bundle, runPath, true)
			if err != nil {
				return finish("blocked", "avaliação global não concluída", err)
			}
			var evaluation RemediationEvaluation
			if err := ReadJSON(evalPath, &evaluation); err != nil {
				return "", err
			}
			if evaluation.Decision != "promote" {
				return finish("blocked", "gates globais rejeitaram a entrega; BEST preservado", fmt.Errorf("validação global rejeitada"))
			}
			continue
		}
		// Reserve one global evaluation instead of exhausting the budget on patches.
		if next == nil || state.TotalAttempts >= cfg.Limits.MaxTotalAttempts-1 {
			return finish("blocked", "tickets pendentes sem tentativa/dependência disponível; BEST preservado", fmt.Errorf("correção não concluída dentro dos limites"))
		}
		item := ticketState(&state, next.ID)
		if item.Status == "pending" || item.Status == "retry" {
			if _, err := PrepareRemediationWorktree(cfg, planPath, next.ID); err != nil {
				return finish("blocked", "preparação de tentativa falhou", err)
			}
		}
		if err := ReadJSON(statePath, &state); err != nil {
			return "", err
		}
		item = ticketState(&state, next.ID)
		if item.Status == "prepared" {
			beforeAttempts := state.TotalAttempts
			if _, err := RunRemediationAgent(ctx, cfg, planPath, next.ID); err != nil {
				if readErr := ReadJSON(statePath, &state); readErr != nil {
					return "", readErr
				}
				if current := ticketState(&state, next.ID); current != nil && current.Status == "blocked_provider_uncertain" {
					return finish("blocked", "resultado do gateway desconhecido; verificar a tentativa antes de repetir", err)
				}
				if state.TotalAttempts <= beforeAttempts {
					return finish("blocked", "tentativa não avançou; recuperação operacional necessária", err)
				}
				continue
			}
		}
		if err := ReadJSON(statePath, &state); err != nil {
			return "", err
		}
		item = ticketState(&state, next.ID)
		if item.Status != "candidate_ready" || item.PreparedCandidateRef == "" {
			return finish("blocked", "agente não deixou candidato rastreável", fmt.Errorf("estado de candidato inválido"))
		}
		runPath, err := RunRemediationGates(ctx, cfg, planPath, next.ID, "candidate", item.PreparedCandidateRef, item.PreparedWorktree)
		if err != nil {
			return finish("blocked", "execução dos gates não concluída", err)
		}
		bundle, err := automationCandidateBundle(cfg, runPath)
		if err != nil {
			if err := rejectUnverifiableAutomationCandidate(cfg, planPath, next.ID, runPath); err != nil {
				return finish("blocked", "não foi possível registrar rejeição", err)
			}
			continue
		}
		if _, err := EvaluateRemediationCandidate(cfg, planPath, next.ID, item.PreparedCandidateRef, bundle, runPath, false); err != nil {
			return finish("blocked", "avaliação do candidato não concluída; BEST preservado", err)
		}
	}
}

func automationCandidateBundle(cfg RemediationConfig, runPath string) (string, error) {
	var run RemediationGateRun
	if err := ReadJSON(runPath, &run); err != nil {
		return "", err
	}
	var candidates []string
	for _, result := range run.Results {
		if result.ID != cfg.CandidateBundleGateID || result.Status != "passed" {
			continue
		}
		for _, artifact := range result.Artifacts {
			if !pathWithin(filepath.Dir(runPath), artifact.Path) {
				continue
			}
			hash, err := HashFile(artifact.Path)
			if err != nil || hash != artifact.SHA256 {
				continue
			}
			if _, err := ReadBundle(artifact.Path); err == nil {
				candidates = append(candidates, artifact.Path)
			}
		}
	}
	if len(candidates) != 1 {
		return "", fmt.Errorf("gate security precisa produzir exatamente um bundle íntegro")
	}
	return candidates[0], nil
}

func rejectUnverifiableAutomationCandidate(cfg RemediationConfig, planPath, ticketID, gateRun string) error {
	_, dir, err := readRemediationPlan(planPath)
	if err != nil {
		return err
	}
	unlock, err := acquireRemediationOperation(dir, cfg)
	if err != nil {
		return err
	}
	defer unlock()
	path := filepath.Join(dir, "state.json")
	var state RemediationState
	if err := ReadJSON(path, &state); err != nil {
		return err
	}
	item := ticketState(&state, ticketID)
	if ticketID != "" && (item == nil || item.Status != "candidate_ready") {
		return fmt.Errorf("candidato não está pronto para rejeição")
	}
	gateHash, err := HashFile(gateRun)
	if err != nil {
		return err
	}
	candidate := state.BestRef
	if item != nil {
		candidate = item.PreparedCandidateRef
	}
	receipt := map[string]any{"ticket_id": ticketID, "candidate_commit": candidate, "decision": "reject", "reason": "security bundle unavailable or invalid", "gate_run_sha256": gateHash, "best_commit_preserved": state.BestRef}
	if err := WriteJSON(filepath.Join(filepath.Dir(gateRun), "automation-rejection.json"), receipt); err != nil {
		return err
	}
	state.TotalAttempts++
	if item != nil {
		item.Attempts++
		item.LastFeedback = "independent security verifier failed or produced no valid canonical bundle; candidate rejected; correct the property without changing verifiers"
		item.PreparedBaseRef = ""
		item.PreparedWorktree = ""
		item.PreparedAt = ""
		item.PreparedCandidateRef = ""
		item.Status = "retry"
		if item.Attempts >= cfg.Limits.MaxAttemptsPerTicket {
			item.Status = "blocked_attempt_limit"
		}
	}
	state.Revision++
	state.UpdatedAt = now()
	return writeJSONAtomic(path, state)
}
