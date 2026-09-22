package app

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
)

func TestDeliveryPackageIsBoundAndDeterministic(t *testing.T) {
	root := t.TempDir()
	caseID := "LAB-DELIVERY-001"
	commit := "0123456789abcdef0123456789abcdef01234567"
	bundle := newNormalizedBundle("eng_delivery", "scope_delivery")
	bundle.GeneratedAt = "2026-08-23T00:00:00Z"
	bundle.Scope = Scope{ID: "scope_delivery", Fingerprint: HashBytes([]byte("scope"))}
	bundle.Engagement = Engagement{ID: "eng_delivery", ScopeID: "scope_delivery", Status: "completed", ToolRunIDs: []string{"run-delivery"}}
	bundle.ToolRuns = []ToolRun{{ID: "run-delivery", Tool: "reference", Version: "1.0.0", Status: "success", InputPath: "git:" + commit, InputSHA256: HashBytes([]byte(commit)), Parser: "reference", ParserVersion: "1.0.0", ToolDigest: "sha256:" + HashBytes([]byte("tool")), PolicySHA256: HashBytes([]byte("policy")), CommandSHA256: HashBytes([]byte("command")), SupplyChainStatus: "runner-gated"}}
	bundlePath := filepath.Join(root, "bundle.json")
	if err := WriteBundle(bundlePath, bundle); err != nil {
		t.Fatal(err)
	}
	bundleHash, _ := HashFile(bundlePath)
	ticket := RemediationTicket{ID: "REM-ABCDEF123456", FindingFingerprint: HashBytes([]byte("finding")), RequiresCodeChange: true, AllowedPaths: []string{"app/**"}, RetestGateIDs: []string{"retest"}}
	plan := RemediationPlan{SchemaVersion: RemediationSchemaVersion, CaseID: caseID, RepositoryPath: filepath.Join(root, "repository"), BaselineCommit: commit, SecurityChangesOnly: true, HumanReviewStages: []string{"final"}, ExecutionStrategy: "serial-best-version", Tickets: []RemediationTicket{ticket}}
	planPath := filepath.Join(root, "plan.json")
	if err := WriteJSON(planPath, plan); err != nil {
		t.Fatal(err)
	}
	planHash, _ := HashFile(planPath)
	state := RemediationState{SchemaVersion: RemediationSchemaVersion, CaseID: caseID, BestRef: commit, BestBundlePath: bundlePath, BestBundleSHA256: bundleHash, GlobalGatePassed: true, FinalApprovalStatus: "approved", Tickets: []TicketState{{TicketID: ticket.ID, Status: "accepted"}}}
	statePath := filepath.Join(root, "state.json")
	if err := WriteJSON(statePath, state); err != nil {
		t.Fatal(err)
	}
	stateHash, _ := HashFile(statePath)
	approval := RemediationApproval{SchemaVersion: RemediationSchemaVersion, CaseID: caseID, Decision: "approved", Reviewer: "Human Reviewer", ReviewedAt: "2026-08-23T00:00:00Z", ReviewedCommit: commit, ReviewedBundleSHA256: bundleHash, RequestedTicketIDs: []string{}, Notes: "laboratory"}
	approvalPath := filepath.Join(root, "approval.json")
	if err := WriteJSON(approvalPath, approval); err != nil {
		t.Fatal(err)
	}
	approvalHash, _ := HashFile(approvalPath)
	authorization := DeliveryAuthorization{SchemaVersion: RemediationSchemaVersion, CaseID: caseID, Status: "approved_for_delivery", GeneratedAt: "2026-08-23T00:00:00Z", Commit: commit, BundleSHA256: bundleHash, PlanSHA256: planHash, StateSHA256: stateHash, ApprovalSHA256: approvalHash, HumanReviewStage: "final", SecurityNonRegression: "passed"}
	patchData := []byte("synthetic approved patch\n")
	authorization.PatchSHA256 = bytesSHA256(patchData)
	if err := os.WriteFile(filepath.Join(root, "approved-security.patch"), patchData, 0o600); err != nil {
		t.Fatal(err)
	}
	authorizationPath := filepath.Join(root, "authorization.json")
	if err := WriteJSON(authorizationPath, authorization); err != nil {
		t.Fatal(err)
	}
	patchRoot := filepath.Join(root, "patches", caseID)
	if err := os.MkdirAll(patchRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(patchRoot, ticket.ID+".patch"), []byte("sanitized patch\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	first, err := BuildDeliveryPackage(context.Background(), authorizationPath, planPath, statePath, approvalPath, bundlePath, "", filepath.Join(root, "out-1"), identity.Recipient().String())
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildDeliveryPackage(context.Background(), authorizationPath, planPath, statePath, approvalPath, bundlePath, "", filepath.Join(root, "out-2"), identity.Recipient().String())
	if err != nil {
		t.Fatal(err)
	}
	if first.ManifestSHA256 != second.ManifestSHA256 || first.ArchiveSHA256 != second.ArchiveSHA256 {
		t.Fatalf("manifesto/arquivo precisam ser determinísticos: first=%+v second=%+v", first, second)
	}
	pdf, err := os.ReadFile(filepath.Join(root, "out-1", "package", "report.pdf"))
	if err != nil || len(pdf) < 100 || string(pdf[:5]) != "%PDF-" {
		t.Fatalf("PDF inválido: err=%v", err)
	}
	if first.EncryptedSHA256 == second.EncryptedSHA256 {
		t.Fatal("age deve usar encapsulamento aleatório autenticado")
	}
	archive, err := os.Open(first.EncryptedPath)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := age.Decrypt(archive, identity)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := os.ReadFile(first.ArchivePath)
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	archive.Close()
	if !bytes.Equal(buffer, decrypted) {
		t.Fatal("conteúdo age descriptografado diverge do arquivo determinístico")
	}
	if _, err := BuildDeliveryPackage(context.Background(), authorizationPath, planPath, statePath, approvalPath, bundlePath, filepath.Dir(patchRoot), filepath.Join(root, "external"), identity.Recipient().String()); err == nil {
		t.Fatal("external patches accepted")
	}
	if err := os.WriteFile(filepath.Join(root, "approved-security.patch"), []byte("substituted patch"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildDeliveryPackage(context.Background(), authorizationPath, planPath, statePath, approvalPath, bundlePath, "", filepath.Join(root, "tampered"), identity.Recipient().String()); err == nil {
		t.Fatal("tampered approved patch accepted")
	}
}

func TestDeliveryRejectsHashMismatch(t *testing.T) {
	if _, err := BuildDeliveryPackage(context.Background(), "missing", "missing", "missing", "missing", "missing", "missing", t.TempDir(), "invalid"); err == nil {
		t.Fatal("entrega sem artefatos vinculados deveria falhar")
	}
}
