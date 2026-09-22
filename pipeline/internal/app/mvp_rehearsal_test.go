package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"
)

// The only exported artifacts are assertions/aggregate metrics. Fixtures,
// generated identities, patches and contractual samples stay in t.TempDir.
func writeMVPTestReport(t *testing.T, name string, value any) {
	t.Helper()
	if directory := os.Getenv("MVP_REPORT_DIR"); directory != "" {
		if info, err := os.Lstat(directory); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			t.Fatal("MVP_REPORT_DIR must be an existing ordinary directory")
		}
		path := filepath.Join(directory, name)
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatal("refusing to replace test evidence")
		}
		if err := WriteJSON(path, value); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMVPFirstCustomerRehearsal(t *testing.T) {
	cfg, config := setupPatchGeneration(t)
	root := t.TempDir()
	steps := []string{}
	// Reuse the canonical templates as samples, never as legally signed acts.
	previous := ""
	for _, name := range []string{"01A_PROPOSTA_COMERCIAL.md", "02_CONTRATO_MESTRE_SERVICOS_SEGURANCA.md", "03_SOW_ESCOPO_REGRAS_ENGAJAMENTO.md", "04_AUTORIZACAO_EXPRESSA_TESTES.md"} {
		source, err := os.ReadFile(testRepoPath("../negocio/juridico/" + name))
		if err != nil {
			t.Fatal(err)
		}
		sample := []byte("# SIMULAÇÃO SEM VALIDADE CONTRATUAL\nCliente e prestadora fictícios; nenhum pagamento ou assinatura real.\nAnterior SHA-256: " + previous + "\nModelo SHA-256: " + bytesSHA256(source) + "\n")
		if err := os.WriteFile(filepath.Join(root, name), sample, 0o600); err != nil {
			t.Fatal(err)
		}
		previous = bytesSHA256(sample)
	}
	steps = append(steps, "proposal_contract_sow_authorization_samples_bound")
	input := syntheticOnboardingInput(t)
	input.CaseID = cfg.CaseID
	input.Adapter.RepositoryPath = cfg.RepositoryPath
	input.Adapter.BaselineBundlePath = cfg.BaselineBundlePath
	input.Adapter.RemediationOutputRoot = cfg.OutputRoot
	input.Authorization.PaymentStatus = "paid"
	inputPath := filepath.Join(root, "onboarding.json")
	if err := WriteJSON(inputPath, input); err != nil {
		t.Fatal(err)
	}
	blocked, err := GenerateOnboardingDrafts(inputPath, filepath.Join(root, "blocked"))
	if err != nil || blocked.Authorized || blocked.Executable {
		t.Fatalf("unauthorized onboarding accepted: %v", err)
	}
	steps = append(steps, "simulated_payment_does_not_authorize")
	sowPath := filepath.Join(root, "03_SOW_ESCOPO_REGRAS_ENGAJAMENTO.md")
	sowHash, err := HashFile(sowPath)
	if err != nil {
		t.Fatal(err)
	}
	input.Authorization = OnboardingAuthorizationInput{Confirmed: true, ConfirmedAt: "2026-09-13T00:00:00Z", ConfirmedBy: "SIMULATED_OWNER", DocumentReference: "SIMULATED-AUTHORIZATION", SOWPath: sowPath, SOWSHA256: sowHash, EmergencyContact: "simulated@example.invalid", PaymentStatus: "paid"}
	input.ExecutionIsolation.Method = "container-egress-policy"
	if err := WriteJSON(inputPath, input); err != nil {
		t.Fatal(err)
	}
	drafts := filepath.Join(root, "authorized")
	status, err := GenerateOnboardingDrafts(inputPath, drafts)
	if err != nil || !status.Authorized || status.Executable {
		t.Fatalf("draft/profile boundary failed: %v", err)
	}
	var scope ScopeConfig
	if err := ReadJSONWithSchema(filepath.Join(drafts, "scope.json"), "scope.schema.json", &scope); err != nil {
		t.Fatal(err)
	}
	if _, err := validateScopeConfig(scope, true); err != nil {
		t.Fatal(err)
	}
	tampered := scope
	tampered.Authorization.SOWSHA256 = strings.Repeat("0", 64)
	if _, err := validateScopeConfig(tampered, true); err == nil {
		t.Fatal("changed SOW accepted")
	}
	scope.Policies = PolicyPaths{Engagement: filepath.Join(drafts, "engagement-policy-p2.json"), Tool: filepath.Join(drafts, "tool-policy-p2.json"), Matrix: filepath.Join(drafts, "test-matrix-p2.json")}
	if err := RunPipeline2(context.Background(), scope, false); err != nil {
		t.Fatal(err)
	}
	steps = append(steps, "authorized_scope_schema_and_sow_hash_checked", "verification_plan_no_scanner_execution")
	// The observed finding is a controlled bundle, not a scanner claim about a real host.
	if _, err := ReadBundle(cfg.BaselineBundlePath); err != nil {
		t.Fatal(err)
	}
	steps = append(steps, "synthetic_verification_bundle_validated")
	summaryPath, err := RunRemediationAutomation(context.Background(), config, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	var summary RemediationAutomationSummary
	if err := ReadJSON(summaryPath, &summary); err != nil {
		t.Fatal(err)
	}
	if summary.Status != "awaiting_final_review" || summary.DeliveryAuthorized {
		t.Fatal("human review bypassed")
	}
	dir := filepath.Dir(summaryPath)
	planPath, statePath := filepath.Join(dir, "plan.json"), filepath.Join(dir, "state.json")
	approvalPath, authPath := filepath.Join(root, "approval.json"), filepath.Join(root, "release", "authorization.json")
	if err := FinalizeRemediation(config, cfg, planPath, approvalPath, authPath); err == nil {
		t.Fatal("missing review accepted")
	}
	state := generationState(t, cfg)
	approval := RemediationApproval{SchemaVersion: RemediationSchemaVersion, CaseID: cfg.CaseID, Decision: "approved", Reviewer: "SIMULATED_FINAL_REVIEWER_NOT_A_REAL_APPROVAL", ReviewedAt: "2026-09-13T00:00:00Z", ReviewedCommit: state.BestRef, ReviewedBundleSHA256: state.BestBundleSHA256, RequestedTicketIDs: []string{}, Notes: "synthetic test only"}
	stale := approval
	stale.ReviewedCommit = strings.Repeat("0", 40)
	if err := WriteJSON(approvalPath, stale); err != nil {
		t.Fatal(err)
	}
	if err := FinalizeRemediation(config, cfg, planPath, approvalPath, authPath); err == nil {
		t.Fatal("stale review accepted")
	}
	if err := WriteJSON(approvalPath, approval); err != nil {
		t.Fatal(err)
	}
	if err := FinalizeRemediation(config, cfg, planPath, approvalPath, authPath); err != nil {
		t.Fatal(err)
	}
	steps = append(steps, "generated_patch_independent_gates_best", "missing_and_stale_review_blocked", "simulated_final_review_bound_to_best")
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	delivery, err := BuildDeliveryPackage(context.Background(), authPath, planPath, statePath, approvalPath, state.BestBundlePath, "", filepath.Join(root, "delivery"), identity.Recipient().String())
	if err != nil {
		t.Fatal(err)
	}
	var plan RemediationPlan
	if err := ReadJSON(planPath, &plan); err != nil {
		t.Fatal(err)
	}
	expected, err := approvedDeliveryPatch(context.Background(), plan, state)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(filepath.Join(root, "delivery", "package", "patches", "security.patch"))
	if err != nil || !bytes.Equal(expected, actual) || !bytes.Contains(actual, []byte("+const secure = true")) {
		t.Fatal("delivery patch differs from reviewed code")
	}
	if delivery.ArchiveSHA256 == "" || delivery.EncryptedSHA256 == "" {
		t.Fatal("delivery integrity absent")
	}
	backupPath := filepath.Join(t.TempDir(), "backup.age")
	backup, err := CreateEncryptedBackup(context.Background(), root, filepath.Join(root, "delivery"), backupPath, identity.Recipient().String())
	if err != nil {
		t.Fatal(err)
	}
	identityPath := filepath.Join(root, "restore-identity")
	if err := os.WriteFile(identityPath, []byte(identity.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreEncryptedBackup(context.Background(), backupPath, identityPath, filepath.Join(root, "restored"))
	if err != nil || !restored.Verified || restored.BackupSHA256 != backup.CiphertextSHA256 {
		t.Fatalf("restore failed: %v", err)
	}
	steps = append(steps, "canonical_patch_and_encrypted_delivery", "encrypted_backup_restore_verified")
	writeMVPTestReport(t, "customer-rehearsal.json", map[string]any{"schema_version": 1, "status": "passed", "simulation_only": true, "real_signatures": false, "real_payments": false, "real_scanners": false, "real_ai": false, "steps": steps})
}

func TestMVPMissingGeneratorInputsBlockBeforeNetwork(t *testing.T) {
	for _, field := range []string{"endpoint", "provider", "model", "effort", "credential", "authorization", "context"} {
		t.Run(field, func(t *testing.T) {
			cfg, config := setupPatchGeneration(t)
			switch field {
			case "endpoint":
				cfg.Agent.Generation.Endpoint = ""
			case "provider":
				cfg.Agent.Generation.Provider = ""
			case "model":
				cfg.Agent.Generation.Model = ""
			case "effort":
				cfg.Agent.Generation.ReasoningEffort = ""
			case "credential":
				t.Setenv("PATCH_FIXTURE_KEY", "")
			case "authorization":
				cfg.Agent.Generation.SourceTransferApproved = false
			case "context":
				cfg.Agent.Generation.ContextFiles = []string{"../outside"}
			}
			if _, err := RunRemediationAutomation(context.Background(), config, cfg, ""); err == nil {
				t.Fatal("incomplete input accepted")
			}
			if _, err := os.Stat(cfg.OutputRoot); !os.IsNotExist(err) {
				t.Fatal("blocked config created runtime")
			}
		})
	}
}
