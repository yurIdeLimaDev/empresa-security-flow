package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type staticResolver struct{ addresses []net.IPAddr }

func (r staticResolver) LookupIPAddr(context.Context, string) ([]net.IPAddr, error) {
	return r.addresses, nil
}

func TestPipeline1PlanHasNoNetwork(t *testing.T) {
	root := t.TempDir()
	cfg := LeadConfig{
		CaseID: "LEAD-TEST", Domain: "example.com", BaseURL: "https://example.com",
		OutputRoot: root, MaxRequests: 10, RequestsPerSecond: 1, MaxDownloadBytes: 1024,
		ToolLockPath: testRepoPath("tools.lock.json"), Policies: PolicyPaths{Engagement: testRepoPath("config/examples/engagement-policy-p1.json"), Tool: testRepoPath("config/examples/tool-policy-p1.json"), Matrix: testRepoPath("config/examples/test-matrix-p1.json")},
		HIBPAPIKeyEnv: "HIBP_API_KEY", HIBPUserAgent: "test-agent", PublicContactPaths: []string{"/about"},
	}
	if err := RunPipeline1(context.Background(), cfg, false); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "LEAD-TEST", "pipeline-1")
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"mode": "plan"`) {
		t.Fatalf("manifesto inesperado: %s", data)
	}
	if _, err := os.Stat(filepath.Join(dir, "prospect-report.md")); err != nil {
		t.Fatal(err)
	}
	if err := VerifyCase(dir); err != nil {
		t.Fatal(err)
	}
}

func TestPipeline1ExecuteRequiresFrontLoadedPolicies(t *testing.T) {
	cfg := LeadConfig{CaseID: "LEAD-TEST", Domain: "example.com", BaseURL: "https://example.com", OutputRoot: t.TempDir()}
	if err := RunPipeline1(context.Background(), cfg, true); err == nil {
		t.Fatal("esperava bloqueio")
	}
}

func TestPipeline1RejectsInvalidCategoryAndCORSOrigin(t *testing.T) {
	base := LeadConfig{CaseID: "LEAD-TEST", Domain: "example.com", BaseURL: "https://example.com", OutputRoot: t.TempDir()}
	invalidCategory := base
	invalidCategory.ApplicationCategory = "marketing-label"
	if err := RunPipeline1(context.Background(), invalidCategory, false); err == nil {
		t.Fatal("esperava rejeição da categoria")
	}
	invalidCORS := base
	invalidCORS.CORSOrigin = "file:///tmp/origin"
	if err := RunPipeline1(context.Background(), invalidCORS, false); err == nil {
		t.Fatal("esperava rejeição da origem CORS")
	}
}

func TestScopedClientRejectsOutOfScope(t *testing.T) {
	client := NewScopedClient([]string{"https"}, []string{"allowed.example"}, []int{443}, false, 5, 5, time.Second)
	req, _ := http.NewRequest(http.MethodGet, "https://outside.example/", nil)
	if _, err := client.Do(req); err == nil || !strings.Contains(err.Error(), "host fora do escopo") {
		t.Fatalf("erro inesperado: %v", err)
	}
}

func TestScopedClientEnforcesBudget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer server.Close()
	host := strings.TrimPrefix(server.URL, "http://")
	host = strings.Split(host, ":")[0]
	port := strings.TrimPrefix(strings.TrimPrefix(server.URL, "http://"), host+":")
	portNumber := 0
	for _, c := range port {
		portNumber = portNumber*10 + int(c-'0')
	}
	client := NewScopedClient([]string{"http"}, []string{host}, []int{portNumber}, true, 1, 100, time.Second)
	if _, err := client.Get(server.URL); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Get(server.URL); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Fatalf("esperava budget, recebeu %v", err)
	}
}

func TestScopedClientDialsOnlyResolvedValidatedIP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer server.Close()
	parsed, _ := url.Parse(server.URL)
	port := 0
	for _, c := range parsed.Port() {
		port = port*10 + int(c-'0')
	}
	client := NewScopedClient([]string{"http"}, []string{"allowed.example"}, []int{port}, true, 1, 100, time.Second)
	transport := client.Transport.(*ScopeTransport)
	transport.Resolver = staticResolver{addresses: []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}}
	response, err := client.Get("http://allowed.example:" + parsed.Port() + "/")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status inesperado: %d", response.StatusCode)
	}
}

func TestRouteModulesForBaaS(t *testing.T) {
	profile := StackProfile{Category: "baas", Observations: []Observation{{Value: "Supabase"}}}
	modules := routeModules(profile, nil)
	if !containsFold(modules, "baas") || !containsFold(modules, "storage") {
		t.Fatalf("roteamento: %v", modules)
	}
	if containsFold(modules, "race-conditions") {
		t.Fatalf("módulo indevido: %v", modules)
	}
}

func TestPipeline2PlanDoesNotRequireAuthorization(t *testing.T) {
	root := t.TempDir()
	cfg := ScopeConfig{
		CaseID: "CLIENT-TEST", BaseURL: "https://staging.example.com", OutputRoot: root,
		AllowedSchemes: []string{"https"}, AllowedHosts: []string{"staging.example.com"}, AllowedPorts: []int{443},
		MaxRequests: 100, RequestsPerSecond: 1, MaxConcurrency: 2, TimeoutSeconds: 5,
		ToolLockPath: testRepoPath("tools.lock.json"), Policies: PolicyPaths{Engagement: testRepoPath("config/examples/engagement-policy-p2.json"), Tool: testRepoPath("config/examples/tool-policy-p2.json"), Matrix: testRepoPath("config/examples/test-matrix-p2.json")},
	}
	if err := RunPipeline2(context.Background(), cfg, false); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "CLIENT-TEST", "pipeline-2")
	if err := VerifyCase(dir); err != nil {
		t.Fatal(err)
	}
	profileData, err := os.ReadFile(filepath.Join(dir, "routed-stack-profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(profileData), `"status": "unconfirmed_plan"`) {
		t.Fatalf("plano não pode confirmar o cliente: %s", profileData)
	}
}

func TestPipeline2ExecuteRejectsMissingGate(t *testing.T) {
	cfg := ScopeConfig{
		CaseID: "CLIENT-TEST", BaseURL: "https://staging.example.com", OutputRoot: t.TempDir(),
		AllowedSchemes: []string{"https"}, AllowedHosts: []string{"staging.example.com"}, AllowedPorts: []int{443},
		MaxRequests: 100, RequestsPerSecond: 1, MaxConcurrency: 2,
	}
	if err := RunPipeline2(context.Background(), cfg, true); err == nil {
		t.Fatal("esperava bloqueio de autorização")
	}
}

func TestScenarioRunnerUsesRoutedModules(t *testing.T) {
	balance := 10
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/state" {
			_, _ = w.Write([]byte(fmt.Sprintf(`{"balance":%d}`, balance)))
			return
		}
		balance--
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	parsedHost := strings.TrimPrefix(server.URL, "http://")
	host := strings.Split(parsedHost, ":")[0]
	port := 0
	for _, c := range strings.TrimPrefix(parsedHost, host+":") {
		port = port*10 + int(c-'0')
	}
	scenarioPath := filepath.Join(t.TempDir(), "scenarios.json")
	if err := WriteJSON(scenarioPath, ScenarioFile{Scenarios: []Scenario{
		{Name: "allowed", Domain: "credit", Module: "race-conditions", Method: "POST", Path: "/action", Concurrency: 1, ExpectedStatuses: []int{http.StatusNoContent}, StateCheck: &ScenarioStateCheck{Method: "GET", Path: "/state", JSONPointer: "/balance", ExpectedDelta: -1}},
		{Name: "not-routed", Domain: "billing", Module: "business-logic", Method: "POST", Path: "/action", Concurrency: 1, ExpectedStatuses: []int{http.StatusNoContent}, StateCheck: &ScenarioStateCheck{Method: "GET", Path: "/state", JSONPointer: "/balance", ExpectedDelta: -1}},
	}}); err != nil {
		t.Fatal(err)
	}
	cfg := ScopeConfig{
		ScenarioPath: scenarioPath, AllowedSchemes: []string{"http"}, AllowedHosts: []string{host}, AllowedPorts: []int{port},
		AllowPrivateIPs: true, MaxRequests: 5, RequestsPerSecond: 100, MaxConcurrency: 2, TimeoutSeconds: 2,
	}
	base, _ := url.Parse(server.URL)
	results, err := runScenarios(context.Background(), cfg, base, []string{"race-conditions"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != "allowed" || !results[0].Passed {
		t.Fatalf("roteamento de cenários inesperado: %#v", results)
	}
}

func TestFractionalRatesAreNotTruncated(t *testing.T) {
	cfg := ScopeConfig{BaseURL: "https://example.test", WordlistPath: "words.txt", RequestsPerSecond: 0.5, MaxConcurrency: 4, TimeoutSeconds: 5, NucleiTemplates: []string{"template.yaml"}}
	policies := ValidatedPolicies{Engagement: EngagementPolicy{Environment: EngagementEnvironment{Targets: []NetworkTarget{{Host: "example.test", Port: 443, RatePerSecond: 0.5, MaxConcurrency: 4, MaxRequests: 100}}}}, Tools: ToolPolicyFile{Tools: map[string]ToolPolicyEntry{"ffuf": {Mode: "automatic"}, "nuclei": {Mode: "automatic"}}}}
	actions := pipeline2Plan(cfg, []string{"active-recon"}, t.TempDir(), policies)
	var ffuf, nuclei PlannedAction
	for _, action := range actions {
		if action.Tool == "ffuf" {
			ffuf = action
		}
		if action.Tool == "nuclei" {
			nuclei = action
		}
	}
	if strings.Contains(strings.Join(ffuf.Command, " "), "-rate 0") || !strings.Contains(strings.Join(ffuf.Command, " "), "-p 2.000000") {
		t.Fatalf("taxa ffuf truncada ou ausente: %v", ffuf.Command)
	}
	if nuclei.State != "skipped" || !strings.Contains(nuclei.Reason, "fracionária") {
		t.Fatalf("Nuclei deveria bloquear taxa fracionária: %+v", nuclei)
	}
	leadActions := pipeline1Plan(LeadConfig{Domain: "example.test", BaseURL: "https://example.test", RequestsPerSecond: 0.5}, t.TempDir())
	if leadActions[2].State != "skipped" {
		t.Fatalf("httpx deveria bloquear taxa fracionária: %+v", leadActions[2])
	}
}

func TestHadrianUsesPinnedReleaseJSONContractAndKeepsAuthProbeOff(t *testing.T) {
	cfg := ScopeConfig{
		BaseURL: "https://example.test", OpenAPIURL: "https://example.test/openapi.json",
		HadrianRolesPath: "roles.yaml", HadrianAuthPath: "auth.yaml", HadrianMutations: true,
		RequestsPerSecond: 0.5, TimeoutSeconds: 7,
	}
	actions := pipeline2Plan(cfg, []string{"authorization"}, t.TempDir())
	var dryRun, authorization, authProbe PlannedAction
	for _, action := range actions {
		switch action.ID {
		case "04-hadrian-dry-run":
			dryRun = action
		case "04-hadrian-authorization":
			authorization = action
		case "04-authprobe-optional":
			authProbe = action
		}
	}
	if dryRun.State != "planned" || !containsFold(dryRun.Command, "--dry-run") {
		t.Fatalf("dry-run Hadrian ausente: %+v", dryRun)
	}
	command := strings.Join(authorization.Command, " ")
	if authorization.State != "planned" || !strings.Contains(command, "--output json") || !strings.Contains(command, "hadrian.json") || strings.Contains(command, "sarif") {
		t.Fatalf("contrato JSON v1.0.0 não respeitado: %+v", authorization)
	}
	if !strings.Contains(command, "--category all") || !strings.Contains(command, "--rate-limit 0.5") || !strings.Contains(command, "--timeout 7") {
		t.Fatalf("limites Hadrian ausentes: %v", authorization.Command)
	}
	if authProbe.State != "skipped" || !strings.Contains(authProbe.Reason, "mutuamente exclusivo") {
		t.Fatalf("AuthProbe não deveria executar com Hadrian: %+v", authProbe)
	}
}

func TestPipeline2ExecutionRequiresIsolationEvidence(t *testing.T) {
	sow := filepath.Join(t.TempDir(), "sow.txt")
	if err := os.WriteFile(sow, []byte("authorized"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := ScopeConfig{
		CaseID: "CLIENT-TEST", BaseURL: "https://staging.example.com", OutputRoot: t.TempDir(),
		AllowedSchemes: []string{"https"}, AllowedHosts: []string{"staging.example.com"}, AllowedPorts: []int{443},
		MaxRequests: 100, RequestsPerSecond: 1, MaxConcurrency: 2, TimeoutSeconds: 5,
		Authorization: Authorization{Confirmed: true, PaymentStatus: "paid", EmergencyContact: "security@example.test", SOWPath: sow, SOWSHA256: HashBytes([]byte("authorized"))},
	}
	if _, err := validateScopeConfig(cfg, true); err == nil || !strings.Contains(err.Error(), "container-egress-policy") {
		t.Fatalf("esperava bloqueio de isolamento, recebeu %v", err)
	}
}

func TestProfileSubjectMustMatchScopedHost(t *testing.T) {
	root := t.TempDir()
	profilePath := filepath.Join(root, "profile.json")
	if err := WriteJSON(profilePath, StackProfile{Subject: "other.example", Status: "hypothesis", Category: "unknown"}); err != nil {
		t.Fatal(err)
	}
	cfg := ScopeConfig{
		CaseID: "CLIENT-TEST", BaseURL: "https://staging.example.com", OutputRoot: root, StackProfilePath: profilePath,
		AllowedSchemes: []string{"https"}, AllowedHosts: []string{"staging.example.com"}, AllowedPorts: []int{443},
		MaxRequests: 100, RequestsPerSecond: 1, MaxConcurrency: 2, TimeoutSeconds: 5,
		ToolLockPath: testRepoPath("tools.lock.json"), Policies: PolicyPaths{Engagement: testRepoPath("config/examples/engagement-policy-p2.json"), Tool: testRepoPath("config/examples/tool-policy-p2.json"), Matrix: testRepoPath("config/examples/test-matrix-p2.json")},
	}
	if err := RunPipeline2(context.Background(), cfg, false); err == nil {
		t.Fatal("esperava rejeição de perfil pertencente a outro domínio")
	}
}

func testRepoPath(relative string) string {
	candidates := []string{relative, filepath.Join("pipeline", relative), filepath.Join("..", relative), filepath.Join("..", "..", relative)}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return relative
}
