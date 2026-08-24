package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOnboardingDraftsStayBlockedWithoutAuthorization(t *testing.T) {
	runner, err := filepath.Abs(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	runnerHash, err := HashFile(runner)
	if err != nil {
		t.Fatal(err)
	}
	input := OnboardingInput{
		SchemaVersion: OnboardingSchemaVersion, CaseID: "LAB-ONBOARD-001", EngagementID: "eng_lab_onboard_001",
		Domain: "lab.invalid", BaseURL: "http://lab.invalid:8080", OutputRoot: filepath.Join(t.TempDir(), "pipeline-output"),
		Target:           NetworkTarget{Host: "lab.invalid", IP: "127.0.0.1", Port: 8080, RatePerSecond: 1, MaxConcurrency: 1, MaxRequests: 20},
		MaxDownloadBytes: 1024, ToolLockPath: testRepoPath("tools.lock.json"), PublicContactPaths: []string{"/contact"},
		PublicObservation: PublicObservationPolicy{Acknowledged: true, AcknowledgedAt: "2026-08-23T00:00:00Z", AcknowledgedBy: "operator"},
		ExternalAPIs:      []ExternalAPI{}, AuthoritativeNS: []AuthoritativeNS{}, DNSResolverIP: "1.1.1.1",
		CanaryImage:    "curlimages/curl:8.15.0@sha256:4026b29997dc7c823b51c164b71e2b51e0fd95cce4601f78202c513d97da2922",
		NegativeCanary: CanaryEndpointInput{URL: "http://negative.invalid:8081/health", IP: "127.0.0.2", Port: 8081}, EmergencyOwner: "operator",
		Authorization: OnboardingAuthorizationInput{Confirmed: false, PaymentStatus: "pending"}, ExecutionIsolation: ExecutionIsolation{Confirmed: false},
		EnabledModules: []string{"authorization"}, EnabledToolsP1: []string{"gitleaks"}, EnabledToolsP2: []string{"gitleaks"},
		ToolPolicyP1Source: testRepoPath("config/examples/tool-policy-p1.json"), ToolPolicyP2Source: testRepoPath("config/examples/tool-policy-p2.json"), Tests: []MatrixTest{},
		Adapter: OnboardingAdapterInput{RunnerPath: runner, RunnerSHA256: runnerHash, RepositoryPath: t.TempDir(), BaselineRef: "0123456789abcdef0123456789abcdef01234567", BaselineBundlePath: filepath.Join(t.TempDir(), "baseline.json"), RemediationOutputRoot: t.TempDir()},
	}
	inputPath := filepath.Join(t.TempDir(), "onboarding.json")
	if err := WriteJSON(inputPath, input); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "drafts")
	status, err := GenerateOnboardingDrafts(inputPath, output)
	if err != nil {
		t.Fatal(err)
	}
	if status.Authorized || status.Executable || len(status.Blockers) < 2 {
		t.Fatalf("onboarding sem autorização deveria ficar bloqueado: %+v", status)
	}
	if _, err := os.Stat(filepath.Join(output, "engagement-policy-p2.json")); !os.IsNotExist(err) {
		t.Fatal("policy P2 não deve ser forjada sem autorização")
	}
	remediationBytes, err := os.ReadFile(filepath.Join(output, "remediation.json"))
	if err != nil {
		t.Fatal(err)
	}
	var remediation RemediationConfig
	if err := json.Unmarshal(remediationBytes, &remediation); err != nil {
		t.Fatal(err)
	}
	if remediation.Agent.ExecutableSHA256 != strings.Repeat("0", 64) {
		t.Fatal("adapter genérico deve permanecer fail-closed até existir perfil revisado")
	}
	for _, gate := range remediation.Gates {
		if gate.ExecutableSHA256 != strings.Repeat("0", 64) {
			t.Fatalf("gate genérico %s deve permanecer fail-closed", gate.ID)
		}
	}
	for _, item := range []struct{ file, schema string }{{"lead.json", "lead.schema.json"}, {"scope.json", "scope.schema.json"}, {"engagement-policy-p1.json", "engagement-policy.schema.json"}, {"tool-policy-p1.json", "tool-policy.schema.json"}, {"tool-policy-p2.json", "tool-policy.schema.json"}, {"test-matrix-p1.json", "test-matrix.schema.json"}, {"test-matrix-p2.json", "test-matrix.schema.json"}, {"remediation.json", "remediation-config.schema.json"}} {
		if err := validateJSONSchemaFile(item.schema, filepath.Join(output, item.file)); err != nil {
			t.Fatalf("%s inválido: %v", item.file, err)
		}
	}
}
