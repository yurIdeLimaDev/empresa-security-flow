package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemediationSecurityPostureIsMonotonic(t *testing.T) {
	baseline := remediationTestBundle(t, true)
	candidate := remediationTestBundle(t, false)
	fingerprint := baseline.Findings[0].Fingerprint

	decision := evaluateSecurityPosture(baseline, candidate, fingerprint, false)
	if !decision.Comparable || !decision.TargetResolved || !decision.Improved {
		t.Fatalf("correção comparável deveria melhorar a postura: %+v", decision)
	}

	regressed := candidate
	addFinding(&regressed, Finding{
		Title: "Novo problema", Category: "configuration", Severity: "critical", Confidence: "high",
		Status: "validated", Location: "src/other.go:1", EvidenceIDs: []string{regressed.Evidence[0].ID},
		ToolRunIDs: []string{regressed.ToolRuns[0].ID}, CanonicalKey: "new-regression",
	})
	finalizeBundle(&regressed)
	decision = evaluateSecurityPosture(baseline, regressed, fingerprint, false)
	if decision.Improved || len(decision.NewFindings) != 1 {
		t.Fatalf("novo achado deveria bloquear promoção: %+v", decision)
	}
}

func TestRemediationApprovalSchemaRequiresExplicitReopenTickets(t *testing.T) {
	dir := t.TempDir()
	base := RemediationApproval{
		SchemaVersion: RemediationSchemaVersion, CaseID: "CASE-001", Reviewer: "Human Reviewer",
		ReviewedAt: "2026-08-22T12:00:00Z", ReviewedCommit: strings.Repeat("a", 40),
		ReviewedBundleSHA256: strings.Repeat("b", 64), Notes: "review",
	}
	base.Decision = "changes_requested"
	base.RequestedTicketIDs = []string{}
	path := filepath.Join(dir, "invalid-changes.json")
	if err := WriteJSON(path, base); err != nil {
		t.Fatal(err)
	}
	var parsed RemediationApproval
	if err := ReadJSONWithSchema(path, "remediation-approval.schema.json", &parsed); err == nil {
		t.Fatal("changes_requested sem ticket deveria falhar")
	}
	base.RequestedTicketIDs = []string{"REM-0123456789AB"}
	if err := WriteJSON(path, base); err != nil {
		t.Fatal(err)
	}
	if err := ReadJSONWithSchema(path, "remediation-approval.schema.json", &parsed); err != nil {
		t.Fatalf("changes_requested explícito deveria ser válido: %v", err)
	}
}

func TestRemediationEndToEndHasOneFinalHumanGate(t *testing.T) {
	root := t.TempDir()
	repository := filepath.Join(root, "client-project")
	if err := os.MkdirAll(filepath.Join(repository, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repository, "tests"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "src", "app.go"), []byte("package sample\n\nconst secure = false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "tests", "security.txt"), []byte("baseline\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitForTest(t, repository, "init")
	runGitForTest(t, repository, "config", "user.name", "Remediation Test")
	runGitForTest(t, repository, "config", "user.email", "remediation@example.invalid")
	runGitForTest(t, repository, "add", ".")
	runGitForTest(t, repository, "commit", "-m", "baseline")

	baselinePath := filepath.Join(root, "baseline-bundle.json")
	if err := WriteBundle(baselinePath, remediationTestBundle(t, true)); err != nil {
		t.Fatal(err)
	}
	candidatePath := filepath.Join(root, "candidate-bundle.json")
	if err := WriteBundle(candidatePath, remediationTestBundle(t, false)); err != nil {
		t.Fatal(err)
	}
	badBundle := remediationTestBundle(t, true)
	addFinding(&badBundle, Finding{
		Title: "Regression", Category: "configuration", Severity: "critical", Confidence: "high",
		Status: "validated", Location: "src/regression.go:1", EvidenceIDs: []string{badBundle.Evidence[0].ID},
		ToolRunIDs: []string{badBundle.ToolRuns[0].ID}, CanonicalKey: "remediation-regression",
	})
	badBundlePath := filepath.Join(root, "bad-bundle.json")
	if err := WriteBundle(badBundlePath, badBundle); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REMEDIATION_TEST_BUNDLE", badBundlePath)

	helperCommand := []string{os.Args[0], "-test.run=^TestRemediationGateHelper$"}
	agentCommand := []string{os.Args[0], "-test.run=^TestRemediationAgentHelper$"}
	helperHash, err := HashFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	cfg := RemediationConfig{
		SchemaVersion: RemediationSchemaVersion, CaseID: "CASE-E2E", RepositoryPath: repository,
		BaselineRef: "HEAD", BaselineBundlePath: baselinePath, OutputRoot: filepath.Join(root, "runtime"),
		SecurityChangesOnly: true, HumanReviewStage: "final",
		Limits: RemediationLimits{MaxAttemptsPerTicket: 3, MaxTotalAttempts: 8},
		ChangePolicy: RemediationChangePolicy{
			DefaultAllowedPaths: []string{"src/**"}, TestPaths: []string{"tests/**"},
			ProtectedPaths: []string{".github/**"}, MaxFilesChanged: 5, MaxChangedLines: 100,
			AllowDependencyChange: false, AllowBinaryChanges: false,
		},
		CandidateBundleGateID: "security-full", DefaultRetestGateIDs: []string{"retest-target"},
		EnvironmentAllowlist: []string{"REMEDIATION_TEST_BUNDLE"},
		Agent:                RemediationAgent{Command: agentCommand, ExecutableSHA256: helperHash, TimeoutSeconds: 20, AutoCommit: true},
		Gates: []RemediationGate{
			{ID: "quality-build", Class: "quality", Command: helperCommand, ExecutableSHA256: helperHash, TimeoutSeconds: 20, Required: true, WorkingDirectory: "", ArtifactPaths: []string{}},
			{ID: "security-full", Class: "security", Command: helperCommand, ExecutableSHA256: helperHash, TimeoutSeconds: 20, Required: true, WorkingDirectory: "", ArtifactPaths: []string{"candidate-bundle.json"}},
			{ID: "retest-target", Class: "retest", Command: helperCommand, ExecutableSHA256: helperHash, TimeoutSeconds: 20, Required: true, WorkingDirectory: "", ArtifactPaths: []string{}},
		},
		FindingOverrides: []FindingRemediationOverride{},
	}
	configPath := filepath.Join(root, "remediation.json")
	if err := WriteJSON(configPath, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadRemediationConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	readiness, err := CheckRemediationReadiness(configPath, loaded)
	if err != nil || !readiness.Approved || len(readiness.Checks) != 4 {
		t.Fatalf("adaptadores reais e pinados deveriam passar no preflight: err=%v report=%+v", err, readiness)
	}
	readinessPath := filepath.Join(root, "remediation-readiness.json")
	if err := WriteJSON(readinessPath, readiness); err != nil {
		t.Fatal(err)
	}
	var parsedReadiness RemediationReadinessReport
	if err := ReadJSONWithSchema(readinessPath, "remediation-readiness.schema.json", &parsedReadiness); err != nil {
		t.Fatalf("relatório de prontidão deveria respeitar o schema: %v", err)
	}
	placeholder := loaded
	placeholder.Agent.ExecutableSHA256 = strings.Repeat("0", 64)
	blockedReadiness, err := CheckRemediationReadiness(configPath, placeholder)
	if err != nil || blockedReadiness.Approved || blockedReadiness.Checks[0].Status != "blocked" {
		t.Fatalf("hash placeholder deveria bloquear o preflight: err=%v report=%+v", err, blockedReadiness)
	}
	planPath, err := CreateRemediationPlan(configPath, loaded)
	if err != nil {
		t.Fatal(err)
	}
	var plan RemediationPlan
	if err := ReadJSON(planPath, &plan); err != nil {
		t.Fatal(err)
	}
	if len(plan.HumanReviewStages) != 1 || plan.HumanReviewStages[0] != "final" || plan.ExecutionStrategy != "serial-best-version" {
		t.Fatalf("plano possui gates humanos ou estratégia inesperados: %+v", plan)
	}
	ticketID := plan.Tickets[0].ID
	mutated := loaded
	mutated.Limits.MaxTotalAttempts++
	if _, err := PrepareRemediationWorktree(mutated, planPath, ticketID); err == nil {
		t.Fatal("política alterada após o plano deveria ser rejeitada")
	}
	metadataPath, err := PrepareRemediationWorktree(loaded, planPath, ticketID)
	if err != nil {
		t.Fatal(err)
	}
	var rejectedMetadata RemediationWorktree
	if err := ReadJSON(metadataPath, &rejectedMetadata); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rejectedMetadata.Path, "src", "app.go"), []byte("package sample\n\nconst secure = false // unsafe candidate\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitForTest(t, rejectedMetadata.Path, "add", "src/app.go")
	runGitForTest(t, rejectedMetadata.Path, "commit", "-m", "unsafe candidate")
	rejectedCommit := strings.TrimSpace(runGitForTest(t, rejectedMetadata.Path, "rev-parse", "HEAD"))
	rejectedGateRun, err := RunRemediationGates(context.Background(), loaded, planPath, ticketID, "candidate", rejectedCommit, rejectedMetadata.Path)
	if err != nil {
		t.Fatal(err)
	}
	rejectedEvaluationPath, err := EvaluateRemediationCandidate(loaded, planPath, ticketID, rejectedCommit, badBundlePath, rejectedGateRun, false)
	if err != nil {
		t.Fatal(err)
	}
	var rejectedEvaluation RemediationEvaluation
	if err := ReadJSON(rejectedEvaluationPath, &rejectedEvaluation); err != nil || rejectedEvaluation.Decision != "reject" || rejectedEvaluation.KeepVersion != "best" {
		t.Fatalf("candidate pior deveria ser descartado: err=%v evaluation=%+v", err, rejectedEvaluation)
	}
	statePath := filepath.Join(filepath.Dir(planPath), "state.json")
	var stateAfterRejection RemediationState
	if err := ReadJSON(statePath, &stateAfterRejection); err != nil || stateAfterRejection.BestRef != plan.BaselineCommit {
		t.Fatalf("rejeição substituiu a melhor versão: err=%v state=%+v", err, stateAfterRejection)
	}

	t.Setenv("REMEDIATION_TEST_BUNDLE", candidatePath)
	metadataPath, err = PrepareRemediationWorktree(loaded, planPath, ticketID)
	if err != nil {
		t.Fatal(err)
	}
	var metadata RemediationWorktree
	if err := ReadJSON(metadataPath, &metadata); err != nil {
		t.Fatal(err)
	}
	agentRunPath, err := RunRemediationAgent(context.Background(), loaded, planPath, ticketID)
	if err != nil {
		t.Fatal(err)
	}
	var agentRun RemediationAgentRun
	if err := ReadJSON(agentRunPath, &agentRun); err != nil || agentRun.Status != "candidate-ready" {
		t.Fatalf("agente deveria produzir candidato commitado: err=%v run=%+v", err, agentRun)
	}
	candidateCommit := agentRun.CandidateCommit

	gateRun, err := RunRemediationGates(context.Background(), loaded, planPath, ticketID, "candidate", candidateCommit, metadata.Path)
	if err != nil {
		t.Fatal(err)
	}
	evaluationPath, err := EvaluateRemediationCandidate(loaded, planPath, ticketID, candidateCommit, candidatePath, gateRun, false)
	if err != nil {
		t.Fatal(err)
	}
	var evaluation RemediationEvaluation
	if err := ReadJSON(evaluationPath, &evaluation); err != nil || evaluation.Decision != "promote" {
		t.Fatalf("candidate deveria ser promovido: err=%v evaluation=%+v", err, evaluation)
	}

	globalRun, err := RunRemediationGates(context.Background(), loaded, planPath, "", "global", candidateCommit, metadata.Path)
	if err != nil {
		t.Fatal(err)
	}
	globalEvaluationPath, err := EvaluateRemediationCandidate(loaded, planPath, "", candidateCommit, candidatePath, globalRun, true)
	if err != nil {
		t.Fatal(err)
	}
	var globalEvaluation RemediationEvaluation
	if err := ReadJSON(globalEvaluationPath, &globalEvaluation); err != nil || globalEvaluation.Decision != "promote" {
		t.Fatalf("gate global deveria promover a mesma melhor versão: err=%v evaluation=%+v", err, globalEvaluation)
	}

	var state RemediationState
	if err := ReadJSON(statePath, &state); err != nil {
		t.Fatal(err)
	}
	approval := RemediationApproval{
		SchemaVersion: RemediationSchemaVersion, CaseID: loaded.CaseID, Decision: "approved",
		Reviewer: "Human Reviewer", ReviewedAt: "2026-08-22T12:00:00Z", ReviewedCommit: state.BestRef,
		ReviewedBundleSHA256: state.BestBundleSHA256, RequestedTicketIDs: []string{}, Notes: "final check",
	}
	approvalPath := filepath.Join(root, "approval.json")
	if err := WriteJSON(approvalPath, approval); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(root, "delivery-authorization.json")
	if err := FinalizeRemediation(configPath, loaded, planPath, approvalPath, outputPath); err != nil {
		t.Fatal(err)
	}
	var authorization DeliveryAuthorization
	if err := ReadJSON(outputPath, &authorization); err != nil || authorization.Status != "approved_for_delivery" {
		t.Fatalf("entrega final deveria ser autorizada: err=%v authorization=%+v", err, authorization)
	}
}

func TestRemediationGateHelper(t *testing.T) {
	output := os.Getenv("REMEDIATION_GATE_OUTPUT")
	if output == "" {
		return
	}
	source := os.Getenv("REMEDIATION_TEST_BUNDLE")
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(output, "candidate-bundle.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestRemediationAgentHelper(t *testing.T) {
	worktree := os.Getenv("REMEDIATION_WORKTREE")
	if worktree == "" {
		return
	}
	path := filepath.Join(worktree, "src", "app.go")
	if err := os.WriteFile(path, []byte("package sample\n\nconst secure = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func remediationTestBundle(t *testing.T, withFinding bool) NormalizedBundle {
	t.Helper()
	bundle := newNormalizedBundle("eng-remediation", "scope-remediation")
	bundle.Scope.AllowedHosts = []string{"example.invalid"}
	run := ToolRun{
		ID: "run-security", Tool: "semgrep", Version: "1.140.0", Status: "normalized",
		InputPath: "semgrep.sarif", InputSHA256: strings.Repeat("a", 64), Parser: "sarif",
		ParserVersion: "1.0.0", ToolDigest: "sha256:" + strings.Repeat("c", 64),
		PolicySHA256: strings.Repeat("d", 64), SupplyChainStatus: "runner-gated",
	}
	bundle.ToolRuns = append(bundle.ToolRuns, run)
	evidenceID := addEvidence(&bundle, run.ID, "semgrep.sarif", strings.Repeat("b", 64), "result/0", "2026-08-22T00:00:00Z", false)
	if withFinding {
		addFinding(&bundle, Finding{
			Title: "Security control missing", Description: "Synthetic controlled finding", Category: "configuration",
			Severity: "high", Confidence: "high", Status: "validated", Location: "src/app.go:3",
			EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{run.ID}, Remediation: "Enable the control",
			CanonicalKey: "remediation-test-target",
		})
	}
	finalizeBundle(&bundle)
	return bundle
}

func runGitForTest(t *testing.T, repository string, args ...string) string {
	t.Helper()
	command := append([]string{"-C", repository}, args...)
	cmd := exec.Command("git", command...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, output)
	}
	return string(output)
}
