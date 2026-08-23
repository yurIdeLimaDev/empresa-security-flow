package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func falseToolFlags() ToolFlags { return ToolFlags{} }

func validP1Policy() EngagementPolicy {
	return EngagementPolicy{
		EngagementID: "eng_test", PipelineMode: "pipeline1_public_low_impact",
		PublicObservation: &PublicObservationPolicy{Acknowledged: true, AcknowledgedAt: "2026-08-22T00:00:00Z", AcknowledgedBy: "tester"},
		Environment: EngagementEnvironment{
			Targets:      []NetworkTarget{{Host: "allowed.test", IP: "203.0.113.10", Port: 443, RatePerSecond: 1, MaxConcurrency: 1, MaxRequests: 10}},
			ExternalAPIs: []ExternalAPI{}, AuthoritativeNS: []AuthoritativeNS{}, DNSResolverIP: "1.1.1.1",
			CanaryImage: "curlimages/curl:8.15.0@sha256:4026b29997dc7c823b51c164b71e2b51e0fd95cce4601f78202c513d97da2922",
		},
		EmergencyStop: EmergencyStopPolicy{Script: "eng-stop.sh", Owner: "tester"},
		Canary:        CanaryPolicy{NegativeURL: "https://negative.operator.test/health", NegativeIP: "203.0.113.99", NegativePort: 443, PositiveURL: "https://allowed.test/health", PositiveIP: "203.0.113.10", PositivePort: 443},
	}
}

func validP2Policy() EngagementPolicy {
	policy := validP1Policy()
	policy.PipelineMode = "pipeline2_active"
	policy.PublicObservation = nil
	policy.Authorization = &EngagementAuthorization{Confirmed: true, ConfirmedAt: "2026-08-22T00:00:00Z", ConfirmedBy: "tester", DocumentReference: "sow"}
	policy.Environment.AuthoritativeNS = []AuthoritativeNS{{Host: "ns.allowed.test", IP: "203.0.113.53"}}
	return policy
}

func TestPolicySchemasRejectIncompleteAndConstFalseViolation(t *testing.T) {
	dir := t.TempDir()
	entry := ToolPolicyEntry{Mode: "automatic", Pipeline: 1, Pin: ToolPin{Source: "git_commit", Value: strings.Repeat("a", 40)}, MaxDurationMinutes: 1, Prohibitions: []string{}, Flags: falseToolFlags()}
	valid := ToolPolicyFile{Version: 1, Tools: map[string]ToolPolicyEntry{"dnsreaper": entry}}
	path := filepath.Join(dir, "tool.json")
	if err := WriteJSON(path, valid); err != nil {
		t.Fatal(err)
	}
	if err := validateJSONSchemaFile("tool-policy.schema.json", path); err != nil {
		t.Fatal(err)
	}
	entry.Flags.ClaimResource = true
	valid.Tools["dnsreaper"] = entry
	if err := WriteJSON(path, valid); err != nil {
		t.Fatal(err)
	}
	if err := validateJSONSchemaFile("tool-policy.schema.json", path); err == nil {
		t.Fatal("claim_resource=true deveria ser recusado pelo schema")
	}
	if err := os.WriteFile(path, []byte(`{"version":1,"tools":{"dnsreaper":{}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateJSONSchemaFile("tool-policy.schema.json", path); err == nil {
		t.Fatal("policy incompleta deveria ser recusada")
	}
}

func TestPolicySemanticGates(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ValidatedPolicies)
		want   string
	}{
		{"p2 authorization", func(p *ValidatedPolicies) { p.Engagement = validP2Policy(); p.Engagement.Authorization = nil }, "authorization.confirmed"},
		{"hibp domain", func(p *ValidatedPolicies) {
			e := p.Tools.Tools["hibp"]
			e.Flags.DomainSearch = true
			p.Tools.Tools["hibp"] = e
		}, "domain_search"},
		{"truffle verify", func(p *ValidatedPolicies) {
			e := p.Tools.Tools["trufflehog"]
			e.Flags.VerifySecrets = true
			p.Tools.Tools["trufflehog"] = e
		}, "verificar segredos"},
		{"active matrix", func(p *ValidatedPolicies) {
			p.Engagement = validP2Policy()
			p.Tools.Tools = map[string]ToolPolicyEntry{"supashield": {Mode: "automatic", Pipeline: 2, MaxDurationMinutes: 1, Flags: falseToolFlags()}}
		}, "test-matrix"},
		{"negative allowlisted", func(p *ValidatedPolicies) { p.Engagement.Canary.NegativeIP = p.Engagement.Environment.Targets[0].IP }, "canary negativo"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policies := ValidatedPolicies{Engagement: validP1Policy(), Tools: ToolPolicyFile{Version: 1, Tools: map[string]ToolPolicyEntry{
				"hibp":       {Mode: "disabled", Pipeline: 1, MaxDurationMinutes: 1, Flags: falseToolFlags()},
				"trufflehog": {Mode: "automatic", Pipeline: 1, MaxDurationMinutes: 1, Flags: falseToolFlags()},
			}}}
			tc.mutate(&policies)
			pipeline := 1
			if policies.Engagement.PipelineMode == "pipeline2_active" {
				pipeline = 2
			}
			if err := validatePolicySemantics(&policies, pipeline); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("erro inesperado: %v", err)
			}
		})
	}
}

func TestRunnerGatesFeroxTruffleHogAndAmass(t *testing.T) {
	p2 := ValidatedPolicies{Engagement: validP2Policy()}
	p2.Engagement.Environment.Targets[0].RatePerSecond = 0
	options := GovernedRunOptions{Pipeline: 2, Policies: p2}
	if err := enforceActionGates(options, PlannedAction{Tool: "feroxbuster", PolicyTool: "feroxbuster", Command: []string{"-u", "https://allowed.test"}}, ToolPolicyEntry{}); err == nil || !strings.Contains(err.Error(), "rate_per_second") {
		t.Fatalf("gate ferox: %v", err)
	}
	p1 := GovernedRunOptions{Pipeline: 1, Policies: ValidatedPolicies{Engagement: validP1Policy()}}
	if err := enforceActionGates(p1, PlannedAction{Tool: "trufflehog", Command: []string{"filesystem", "/work"}}, ToolPolicyEntry{}); err == nil || !strings.Contains(err.Error(), "no-verification") {
		t.Fatalf("gate trufflehog: %v", err)
	}
	options.Policies.Engagement.Authorization = nil
	if err := enforceActionGates(options, PlannedAction{Tool: "amass", PolicyTool: "amass_active"}, ToolPolicyEntry{}); err == nil || !strings.Contains(err.Error(), "autorização") {
		t.Fatalf("gate amass: %v", err)
	}
}

func TestToolPinMismatchIsRejected(t *testing.T) {
	policy := ToolPolicyFile{Tools: map[string]ToolPolicyEntry{"dnsreaper": {Pin: ToolPin{Source: "docker_digest", Value: "sha256:" + strings.Repeat("a", 64)}}}}
	lock := ToolLockFile{Tools: []ToolLock{{Name: "dnsreaper", Container: &ContainerLock{Digest: "sha256:" + strings.Repeat("b", 64)}}}}
	if err := validateToolPolicyPins(policy, lock); err == nil || !strings.Contains(err.Error(), "diverge") {
		t.Fatalf("pin divergente aceito: %v", err)
	}
}

func TestIsolationScriptsAreDeterministicAndFailClosed(t *testing.T) {
	root := t.TempDir()
	policy := validP2Policy()
	first, err := GenerateIsolationScripts(root, policy)
	if err != nil {
		t.Fatal(err)
	}
	setup1, _ := os.ReadFile(first.SetupScript)
	second, err := GenerateIsolationScripts(root, policy)
	if err != nil {
		t.Fatal(err)
	}
	setup2, _ := os.ReadFile(second.SetupScript)
	if string(setup1) != string(setup2) {
		t.Fatal("geração não determinística")
	}
	text := string(setup1)
	for _, required := range []string{"iptables -w -F", "while iptables", "-j DROP", "INPUT_CHAIN", "-I INPUT 1", "--ctstate ESTABLISHED,RELATED -j ACCEPT", "-A \"$INPUT_CHAIN\" -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT", policy.Environment.DNSResolverIP, policy.Environment.AuthoritativeNS[0].IP} {
		if !strings.Contains(text, required) {
			t.Fatalf("setup sem %q", required)
		}
	}
	canary, _ := os.ReadFile(first.CanaryScript)
	if strings.Contains(string(canary), "example.com") || !strings.Contains(string(canary), "--network host") || strings.Contains(string(canary), "--network host --sysctl") {
		t.Fatalf("canary inseguro: %s", canary)
	}
}

func TestSupplyChainRequiresContainerAndReviewEvidence(t *testing.T) {
	lock := ToolLockFile{
		SchemaVersion: ToolLockSchemaVersion,
		Policy:        SupplyChainPolicy{RequireContainerDigest: true, RequireCodeReview: true},
		Tools: []ToolLock{{
			Name: "example-tool", ExecutionClass: "automatic", Release: ReleaseLock{Tag: "v1.0.0", Commit: strings.Repeat("a", 40), ReleaseID: 1},
			SBOMComponentRef: "tool:example", CodeReview: CodeReviewLock{Status: "approved"}, Liveness: &LivenessLock{Command: []string{"--version"}, ExpectedPattern: "1\\.0"},
		}},
	}
	report := ValidateToolLock(lock, filepath.Join(t.TempDir(), "missing-sbom.json"))
	if len(report.Checks) == 0 || report.Checks[0].Approved {
		t.Fatalf("ferramenta sem container/revisão documental foi aprovada: %#v", report)
	}
	blockers := strings.Join(report.Checks[0].Blockers, " ")
	if !strings.Contains(blockers, "container por digest ausente para ferramenta automática") || !strings.Contains(blockers, "revisão sem reviewer, referência ou escopo") {
		t.Fatalf("bloqueios esperados ausentes: %s", blockers)
	}
}

func TestRestrictedCatalogEntryNeedsNoExecutableRuntime(t *testing.T) {
	lock := ToolLockFile{
		SchemaVersion: ToolLockSchemaVersion,
		Policy:        SupplyChainPolicy{RequireContainerDigest: true},
		Tools: []ToolLock{{
			Name: "manual-tool", ExecutionClass: "manual",
			Release:    ReleaseLock{Tag: "commit-" + strings.Repeat("a", 12), TagType: "commit", Commit: strings.Repeat("a", 40)},
			CodeReview: CodeReviewLock{Status: "restricted"},
		}},
	}
	report := ValidateToolLock(lock, filepath.Join(t.TempDir(), "unused.json"))
	if !report.Approved {
		t.Fatalf("entrada restrita não executável deveria fechar a pendência sem inventar container: %+v", report)
	}
	if _, err := ApproveContainer(lock, "manual-tool", filepath.Join(t.TempDir(), "unused.json")); err == nil || !strings.Contains(err.Error(), "container por digest") {
		t.Fatalf("executor aceitou entrada manual sem runtime: %v", err)
	}
}

func TestDockerRunnerAppliesMandatoryHardening(t *testing.T) {
	caseDir := t.TempDir()
	options := GovernedRunOptions{
		CaseDir:  caseDir,
		Network:  NetworkState{NetworkName: "eng-test"},
		Policies: ValidatedPolicies{Engagement: EngagementPolicy{Environment: EngagementEnvironment{DNSResolverIP: "1.1.1.1"}}},
	}
	lock := ToolLock{Container: &ContainerLock{Image: "example/tool:1", Digest: "sha256:" + strings.Repeat("a", 64)}}
	args, err := buildDockerArguments(options, PlannedAction{ID: "test", Command: []string{"--version"}}, lock, "run_test")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, required := range []string{"--pull=never", "--cap-drop ALL", "no-new-privileges=true", "--read-only", "--pids-limit 512", "--memory 2g", "--memory-swap 2g", "--cpus 2", "--tmpfs /tmp:", "HOME=/tmp", "--log-opt max-size=16m"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("runner sem hardening %q: %s", required, joined)
		}
	}
	if strings.Contains(joined, "/var/run/docker.sock") || strings.Contains(joined, "--privileged") {
		t.Fatalf("runner incluiu privilégio indevido: %s", joined)
	}
}

func TestScannerOutputAndArtifactContractsAreBounded(t *testing.T) {
	buffer := cappedBuffer{limit: 1024}
	payload := []byte(strings.Repeat("x", 1<<20))
	if written, err := buffer.Write(payload); err != nil || written != len(payload) {
		t.Fatalf("writer não preservou contrato: written=%d err=%v", written, err)
	}
	if len(buffer.Bytes()) != 1024 || !buffer.truncated {
		t.Fatalf("saída não foi limitada: size=%d truncated=%v", len(buffer.Bytes()), buffer.truncated)
	}
	caseDir := t.TempDir()
	artifact := filepath.Join(caseDir, "result.json")
	if err := os.WriteFile(artifact, []byte("[]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	hashes, err := validateArtifactOutputs(caseDir, []string{artifact})
	if err != nil || len(hashes) != 1 || hashes[0].SizeBytes != 3 {
		t.Fatalf("contrato de artefato válido foi recusado: hashes=%+v err=%v", hashes, err)
	}
	if _, err := validateArtifactOutputs(caseDir, []string{filepath.Join(caseDir, "missing.json")}); err == nil {
		t.Fatal("artefato declarado ausente foi aceito")
	}
	if _, err := validateArtifactOutputs(caseDir, []string{filepath.Join(caseDir, "..", "outside.json")}); err == nil {
		t.Fatal("artefato fora do caso foi aceito")
	}
}

func TestRuntimeReadinessIntegration(t *testing.T) {
	output := os.Getenv("EMPRESA_SECURITY_RUNTIME_CHECK_OUTPUT")
	if output == "" {
		t.Skip("set EMPRESA_SECURITY_RUNTIME_CHECK_OUTPUT for the Docker integration test")
	}
	lockPath := testRepoPath("tools.lock.json")
	lock, err := ReadToolLock(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	sbomPath := testRepoPath("sbom.cdx.json")
	pull := os.Getenv("EMPRESA_SECURITY_RUNTIME_CHECK_PULL") == "1"
	report, err := CheckRuntimeReadiness(context.Background(), lock, sbomPath, pull)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(output, report); err != nil {
		t.Fatal(err)
	}
	if !report.Approved {
		for _, check := range report.Checks {
			if check.Status == "failed" {
				t.Logf("%s: %s (%s)", check.Tool, check.Reason, check.ObservedOutput)
			}
		}
		t.Fatal("runtime readiness possui falhas")
	}
}

func TestAuditProxyLogsCooperatingHTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer server.Close()
	parsed, _ := url.Parse(server.URL)
	port := 0
	_, _ = fmt.Sscanf(parsed.Port(), "%d", &port)
	policy := validP1Policy()
	policy.Environment.Targets = []NetworkTarget{{Host: "allowed.test", IP: "127.0.0.1", Port: port, RatePerSecond: 100, MaxConcurrency: 1, MaxRequests: 10}}
	logPath := filepath.Join(t.TempDir(), "requestlog.ndjson")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	proxy, err := StartAuditProxy(ctx, "127.0.0.1:0", "127.0.0.1:0", logPath, NewPolicyScopeTransport(policy, 10, 100))
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close(context.Background())
	proxyURL, _ := url.Parse("http://" + proxy.HTTPAddress)
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}, Timeout: 2 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, "http://allowed.test:"+parsed.Port()+"/ok?secret=value", nil)
	req.Header.Set("X-Empresa-Tool-Run", "run_test")
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"tool_run_id":"run_test"`) || strings.Contains(string(data), "secret=value") {
		t.Fatalf("RequestLog inválido: %s", data)
	}
}

func TestDNSReaperCuratorAndHIBPResponseParser(t *testing.T) {
	dir := t.TempDir()
	subfinder := filepath.Join(dir, "subfinder.jsonl")
	amass := filepath.Join(dir, "amass.txt")
	_ = os.WriteFile(subfinder, []byte("{\"host\":\"api.example.com\"}\n{\"host\":\"*.example.com\"}\n"), 0o600)
	_ = os.WriteFile(amass, []byte("cdn.example.com\noutside.test\n"), 0o600)
	values, err := CurateDNSReaperCandidates("example.com", subfinder, amass)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(values, ",") != "api.example.com,cdn.example.com" {
		t.Fatalf("curadoria: %v", values)
	}
	body, status, err := splitCurlResponse([]byte("\n404\n"))
	if err != nil || status != 404 || len(body) != 0 {
		t.Fatalf("curl parser: status=%d body=%q err=%v", status, body, err)
	}
}

func TestNormalizeDNSReaperAndHIBP(t *testing.T) {
	dir := t.TempDir()
	dnsPath := filepath.Join(dir, "dnsreaper.json")
	if err := WriteJSON(dnsPath, []DNSReaperCandidate{{Domain: "orphan.example.com", Signature: "demo", Confidence: "POTENTIAL", Status: "candidate"}}); err != nil {
		t.Fatal(err)
	}
	dnsBundle, err := NormalizeFile(NormalizeConfig{Tool: "dnsreaper", InputPath: dnsPath, EngagementID: "eng_test", ScopeID: "scope_test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(dnsBundle.Findings) != 1 || dnsBundle.Findings[0].Status != "candidate" {
		t.Fatalf("dnsReaper deveria gerar somente candidato: %#v", dnsBundle.Findings)
	}
	hibpPath := filepath.Join(dir, "hibp.json")
	email := "security@example.com"
	if err := WriteJSON(hibpPath, []HIBPResult{{AccountSHA256: HashBytes([]byte(email)), HTTPStatus: 200, Breaches: []string{"DemoBreach"}, CheckedAt: now()}}); err != nil {
		t.Fatal(err)
	}
	hibpBundle, err := NormalizeFile(NormalizeConfig{Tool: "hibp", InputPath: hibpPath, EngagementID: "eng_test", ScopeID: "scope_test"})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(hibpBundle)
	if len(hibpBundle.Findings) != 1 || hibpBundle.Findings[0].Status != "candidate" || strings.Contains(string(encoded), email) {
		t.Fatalf("HIBP inválido ou vazou e-mail bruto: %s", encoded)
	}
}
