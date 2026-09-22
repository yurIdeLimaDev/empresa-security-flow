package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type patchRoundTripFunc func(*http.Request) (*http.Response, error)

func (f patchRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func mockPatchResponse(t *testing.T, r *http.Request) PatchGenerationResponse {
	t.Helper()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	var request PatchGenerationRequest
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}
	requestPath := filepath.Join(t.TempDir(), "request.json")
	if err := os.WriteFile(requestPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ReadJSONWithSchema(requestPath, "patch-generation-request.schema.json", &request); err != nil {
		t.Fatal(err)
	}
	if r.Header.Get("X-Patch-Request-SHA256") != bytesSHA256(data) {
		t.Fatal("request hash mismatch")
	}
	if len(request.Files) != 1 || request.Files[0].Path != "src/app.go" {
		t.Fatal("unexpected source disclosure")
	}
	if request.Attempt > 1 && (request.PreviousOutcome == "none" || request.PreviousOutcome == "") {
		t.Fatal("retry omitted verifier feedback")
	}
	if strings.Contains(string(data), "fixture-token-do-not-log") || strings.Contains(string(data), "private.txt") {
		t.Fatal("credential or unauthorized file sent")
	}
	file := request.Files[0]
	return PatchGenerationResponse{SchemaVersion: PatchProtocolVersion, RequestID: request.RequestID, RequestSHA256: bytesSHA256(data), BaseCommit: request.BaseCommit, Provider: request.Provider, Model: request.Model, ReasoningEffort: request.ReasoningEffort, Status: "patch", Replacements: []PatchReplacement{{Path: file.Path, BeforeSHA256: file.SHA256, Content: strings.Replace(file.Content, "secure = false", "secure = true", 1)}}}
}

func mockPatchTransport(t *testing.T, mutate func(*PatchGenerationResponse)) patchTransport {
	return patchRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		response := mockPatchResponse(t, r)
		responsePath := filepath.Join(t.TempDir(), "response.json")
		if err := WriteJSON(responsePath, response); err != nil {
			t.Fatal(err)
		}
		if err := ReadJSONWithSchema(responsePath, "patch-generation-response.schema.json", &response); err != nil {
			t.Fatal(err)
		}
		if mutate != nil {
			mutate(&response)
		}
		data, _ := json.Marshal(response)
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(bytes.NewReader(data))}, nil
	})
}

func setupPatchGeneration(t *testing.T) (RemediationConfig, string) {
	t.Helper()
	root := t.TempDir()
	repo := filepath.Join(root, "client")
	for _, dir := range []string{"src", "tests"} {
		if err := os.MkdirAll(filepath.Join(repo, dir), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for path, content := range map[string]string{"src/app.go": "package sample\n\nconst secure = false\nconst behavior = \"original\"\n", "tests/contract.txt": "protected verifier\n", "private.txt": "not authorized for model\n"} {
		if err := os.WriteFile(filepath.Join(repo, filepath.FromSlash(path)), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runGitForTest(t, repo, "init")
	runGitForTest(t, repo, "config", "core.autocrlf", "false")
	runGitForTest(t, repo, "config", "user.name", "Test Operator")
	runGitForTest(t, repo, "config", "user.email", "test@example.invalid")
	runGitForTest(t, repo, "add", ".")
	runGitForTest(t, repo, "commit", "-m", "test baseline")
	baseline := filepath.Join(root, "baseline.json")
	candidate := filepath.Join(root, "candidate.json")
	if err := WriteBundle(baseline, remediationTestBundle(t, true)); err != nil {
		t.Fatal(err)
	}
	if err := WriteBundle(candidate, remediationTestBundle(t, false)); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATCH_FIXTURE_KEY", "fixture-token-do-not-log")
	t.Setenv("PATCH_GATE_HELPER", "1")
	t.Setenv("REMEDIATION_TEST_BUNDLE", candidate)
	hash, err := HashFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	command := []string{os.Args[0], "-test.run=^TestPatchIndependentGateHelper$"}
	cfg := RemediationConfig{SchemaVersion: RemediationSchemaVersion, CaseID: "CASE-GENERATION", RepositoryPath: repo, BaselineRef: "HEAD", BaselineBundlePath: baseline, OutputRoot: filepath.Join(root, "runtime"), SecurityChangesOnly: true, HumanReviewStage: "final", Limits: RemediationLimits{MaxAttemptsPerTicket: 2, MaxTotalAttempts: 5}, ChangePolicy: RemediationChangePolicy{DefaultAllowedPaths: []string{"src/**"}, TestPaths: []string{"tests/**"}, ProtectedPaths: []string{".github/**"}, MaxFilesChanged: 2, MaxChangedLines: 30}, CandidateBundleGateID: "security-full", DefaultRetestGateIDs: []string{"retest-property"}, EnvironmentAllowlist: []string{"PATCH_GATE_HELPER", "REMEDIATION_TEST_BUNDLE", "PATCH_GATE_FAIL"}, Agent: RemediationAgent{Command: []string{patchGenerationCommand}, ExecutableSHA256: hash, TimeoutSeconds: 5, AutoCommit: true, Generation: &PatchGenerationConfig{Enabled: true, Endpoint: "https://gateway.example.invalid/patch", Provider: "fixture-not-an-ai", Model: "fixture-model", ReasoningEffort: "fixture-effort", APIKeyEnv: "PATCH_FIXTURE_KEY", SourceTransferApproved: true, ContextFiles: []string{"src/app.go"}, MaxContextBytes: 8192, MaxResponseBytes: 16384}}, FindingOverrides: []FindingRemediationOverride{}}
	for _, gate := range []struct{ id, class string }{{"quality-project", "quality"}, {"security-full", "security"}, {"retest-property", "retest"}} {
		artifacts := []string{}
		if gate.class == "security" {
			artifacts = []string{"candidate-bundle.json"}
		}
		cfg.Gates = append(cfg.Gates, RemediationGate{ID: gate.id, Class: gate.class, Command: command, ExecutableSHA256: hash, TimeoutSeconds: 10, Required: true, ArtifactPaths: artifacts})
	}
	config := filepath.Join(root, "remediation.json")
	if err := WriteJSON(config, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadRemediationConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	loaded.generationTransport = mockPatchTransport(t, nil)
	return loaded, config
}

// This subprocess is a synthetic, independent verifier, never the generator.
func TestPatchIndependentGateHelper(t *testing.T) {
	if os.Getenv("PATCH_GATE_HELPER") != "1" {
		return
	}
	if os.Getenv("PATCH_FIXTURE_KEY") != "" {
		t.Fatal("generator credential leaked into gate")
	}
	output := os.Getenv("REMEDIATION_GATE_OUTPUT")
	if output == "" {
		t.Fatal("missing controlled output")
	}
	fail := os.Getenv("PATCH_GATE_FAIL")
	if fail == "all" || fail == "global" && strings.Contains(output, string(filepath.Separator)+"GLOBAL"+string(filepath.Separator)) {
		t.Fatal("deliberate gate failure")
	}
	data, err := os.ReadFile("src/app.go")
	if err != nil {
		t.Fatal(err)
	}
	if !validSyntheticControlSource(data) {
		t.Fatal("security or behavior property failed")
	}
	protected, err := os.ReadFile("tests/contract.txt")
	if err != nil || string(protected) != "protected verifier\n" {
		t.Fatal("verifier changed")
	}
	if filepath.Base(output) == "security-full" {
		data, err := os.ReadFile(os.Getenv("REMEDIATION_TEST_BUNDLE"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "candidate-bundle.json"), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// Parse, do not execute provider-authored code. This deliberately narrow
// reference grammar prevents comments or an unused second constant passing.
func validSyntheticControlSource(data []byte) bool {
	file, err := parser.ParseFile(token.NewFileSet(), "fixture.go", data, 0)
	if err != nil || file.Name.Name != "sample" || len(file.Imports) != 0 {
		return false
	}
	seen := map[string]bool{}
	for _, decl := range file.Decls {
		group, ok := decl.(*ast.GenDecl)
		if !ok || group.Tok != token.CONST {
			return false
		}
		for _, spec := range group.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok || len(value.Names) != 1 || len(value.Values) != 1 {
				return false
			}
			name := value.Names[0].Name
			if seen[name] {
				return false
			}
			seen[name] = true
			switch name {
			case "secure":
				v, ok := value.Values[0].(*ast.Ident)
				if !ok || v.Name != "true" {
					return false
				}
			case "behavior":
				v, ok := value.Values[0].(*ast.BasicLit)
				if !ok || v.Kind != token.STRING || v.Value != "\"original\"" {
					return false
				}
			default:
				return false
			}
		}
	}
	return len(seen) == 2
}

func generationState(t *testing.T, cfg RemediationConfig) RemediationState {
	t.Helper()
	var state RemediationState
	if err := ReadJSON(filepath.Join(cfg.OutputRoot, cfg.CaseID, "remediation", "state.json"), &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func TestPatchAutomationReachesOnlyFinalHumanReview(t *testing.T) {
	cfg, config := setupPatchGeneration(t)
	baseline, _ := resolveCommit(cfg.RepositoryPath, "HEAD")
	calls := 0
	normal := cfg.generationTransport
	cfg.generationTransport = patchRoundTripFunc(func(r *http.Request) (*http.Response, error) { calls++; return normal.RoundTrip(r) })
	path, err := RunRemediationAutomation(context.Background(), config, cfg, "")
	if err != nil {
		for _, pattern := range []string{"agents/*/*/agent.log", "gates/*/*/*/gate.log"} {
			paths, _ := filepath.Glob(filepath.Join(cfg.OutputRoot, cfg.CaseID, "remediation", filepath.FromSlash(pattern)))
			for _, path := range paths {
				data, _ := os.ReadFile(path)
				t.Logf("%s: %s", path, data)
			}
		}
		t.Fatal(err)
	}
	var summary RemediationAutomationSummary
	if err := ReadJSON(path, &summary); err != nil {
		t.Fatal(err)
	}
	if summary.Status != "awaiting_final_review" || !summary.GlobalGatePassed || summary.FinalApprovalStatus != "pending" || summary.DeliveryAuthorized || calls != 1 {
		t.Fatalf("unexpected summary: %+v calls=%d", summary, calls)
	}
	if summary.BestCommit == baseline || summary.TotalAttempts != 2 {
		t.Fatal("candidate/global not evaluated")
	}
	if got, _ := resolveCommit(cfg.RepositoryPath, "HEAD"); got != baseline {
		t.Fatal("original checkout changed")
	}
	planPath := filepath.Join(filepath.Dir(path), "plan.json")
	if _, err := RunRemediationAutomation(context.Background(), config, cfg, planPath); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("completed automation generated another patch")
	}
	// No approval or delivery authorization can be manufactured by the loop.
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "delivery-authorization.json")); !os.IsNotExist(err) {
		t.Fatal("unexpected delivery authorization")
	}
}

func TestPatchAutomationRejectsResponsesAndKeepsBEST(t *testing.T) {
	for name, mutate := range map[string]func(*PatchGenerationResponse){
		"stale request":      func(r *PatchGenerationResponse) { r.RequestID = "replayed" },
		"stale source":       func(r *PatchGenerationResponse) { r.Replacements[0].BeforeSHA256 = strings.Repeat("0", 64) },
		"wrong model":        func(r *PatchGenerationResponse) { r.Model = "not-selected" },
		"wrong effort":       func(r *PatchGenerationResponse) { r.ReasoningEffort = "not-selected" },
		"wrong provider":     func(r *PatchGenerationResponse) { r.Provider = "not-selected" },
		"protected verifier": func(r *PatchGenerationResponse) { r.Replacements[0].Path = "tests/contract.txt" },
		"traversal":          func(r *PatchGenerationResponse) { r.Replacements[0].Path = "../outside" },
		"no code":            func(r *PatchGenerationResponse) { r.Status = "blocked"; r.Replacements = nil },
		"binary":             func(r *PatchGenerationResponse) { r.Replacements[0].Content = "\x00binary" },
		"duplicate edit":     func(r *PatchGenerationResponse) { r.Replacements = append(r.Replacements, r.Replacements[0]) },
		"behavior regression": func(r *PatchGenerationResponse) {
			r.Replacements[0].Content = strings.ReplaceAll(r.Replacements[0].Content, "original", "changed")
		},
		"line budget": func(r *PatchGenerationResponse) { r.Replacements[0].Content += strings.Repeat("// unrelated\n", 40) },
	} {
		t.Run(name, func(t *testing.T) {
			cfg, config := setupPatchGeneration(t)
			baseline, _ := resolveCommit(cfg.RepositoryPath, "HEAD")
			cfg.generationTransport = mockPatchTransport(t, mutate)
			path, err := RunRemediationAutomation(context.Background(), config, cfg, "")
			if err == nil || path == "" {
				t.Fatal("invalid proposal should block after bounded retries")
			}
			state := generationState(t, cfg)
			if state.BestRef != baseline || state.GlobalGatePassed || state.TotalAttempts != 2 || state.Tickets[0].Status != "blocked_attempt_limit" {
				t.Fatalf("BEST or limits failed: %+v", state)
			}
		})
	}
}

func TestPatchGenerationDisabledFailsBeforeWorktreeOrNetwork(t *testing.T) {
	cfg, config := setupPatchGeneration(t)
	cfg.Agent.Generation.Enabled = false
	cfg.Agent.Generation.Endpoint = ""
	cfg.Agent.Generation.Provider = ""
	cfg.Agent.Generation.Model = ""
	cfg.Agent.Generation.ReasoningEffort = ""
	cfg.generationTransport = patchRoundTripFunc(func(*http.Request) (*http.Response, error) { t.Fatal("network call while disabled"); return nil, nil })
	if err := WriteJSON(config, cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRemediationConfig(config); err != nil {
		t.Fatal(err)
	}
	readiness, err := CheckRemediationReadiness(config, cfg)
	if err != nil || readiness.Approved {
		t.Fatal("disabled provider should block readiness")
	}
	if _, err := RunRemediationAutomation(context.Background(), config, cfg, ""); err == nil {
		t.Fatal("disabled automation accepted")
	}
	if _, err := os.Stat(cfg.OutputRoot); !os.IsNotExist(err) {
		t.Fatal("disabled generation created runtime artifacts")
	}
}

func TestPatchConfigRejectsUnsafeSettings(t *testing.T) {
	for name, change := range map[string]func(*RemediationConfig){
		"HTTP":                  func(c *RemediationConfig) { c.Agent.Generation.Endpoint = "http://example.invalid" },
		"embedded credential":   func(c *RemediationConfig) { c.Agent.Generation.Endpoint = "https://user:password@example.invalid" },
		"query credential":      func(c *RemediationConfig) { c.Agent.Generation.Endpoint = "https://example.invalid/?key=test" },
		"transfer not approved": func(c *RemediationConfig) { c.Agent.Generation.SourceTransferApproved = false },
		"credential shared with gates": func(c *RemediationConfig) {
			c.EnvironmentAllowlist = append(c.EnvironmentAllowlist, "PATCH_FIXTURE_KEY")
		},
		"protected context":  func(c *RemediationConfig) { c.Agent.Generation.ContextFiles = []string{"tests/contract.txt"} },
		"dot context":        func(c *RemediationConfig) { c.Agent.Generation.ContextFiles = []string{".env"} },
		"dependency context": func(c *RemediationConfig) { c.Agent.Generation.ContextFiles = []string{"package.json"} },
		"wildcard context":   func(c *RemediationConfig) { c.Agent.Generation.ContextFiles = []string{"src/**"} },
	} {
		t.Run(name, func(t *testing.T) {
			cfg, _ := setupPatchGeneration(t)
			change(&cfg)
			if err := validatePatchGenerationConfig(cfg, true); err == nil {
				t.Fatal("unsafe configuration accepted")
			}
		})
	}
}

func TestPatchResponseStrictJSON(t *testing.T) {
	for _, data := range []string{`{"status":"patch","status":"blocked"}`, `{"Status":"patch"}`, `{"tools":["execute"]}`, `{} {}`, `null`, strings.Repeat("[", 10) + strings.Repeat("]", 10)} {
		response, err := decodePatchResponse([]byte(data))
		if err == nil && response.SchemaVersion != "" {
			t.Fatalf("bad contract accepted: %s", data)
		}
		if err == nil {
			t.Fatalf("bad JSON accepted: %s", data)
		}
	}
}

func TestPatchGatewayRejectsRedirectAndDoesNotLeakToken(t *testing.T) {
	cfg, _ := setupPatchGeneration(t)
	hits := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Location", "https://not-authorized.invalid")
		w.WriteHeader(302)
		fmt.Fprint(w, "fixture-token-do-not-log")
	}))
	defer server.Close()
	cfg.Agent.Generation.Endpoint = server.URL
	cfg.generationTransport = server.Client().Transport
	_, err := callPatchGateway(context.Background(), cfg, []byte(`{}`))
	if err == nil || hits != 1 || strings.Contains(err.Error(), "fixture-token") {
		t.Fatal("redirect/error secrecy failure")
	}
}

func TestPatchGatewayBoundsAndCancellation(t *testing.T) {
	cfg, _ := setupPatchGeneration(t)
	cfg.generationTransport = patchRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", cfg.Agent.Generation.MaxResponseBytes+1)))}, nil
	})
	if _, err := callPatchGateway(context.Background(), cfg, []byte(`{}`)); err == nil {
		t.Fatal("oversize response accepted")
	}
	cfg.generationTransport = patchRoundTripFunc(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := callPatchGateway(ctx, cfg, []byte(`{}`)); err == nil {
		t.Fatal("timeout accepted")
	}
}

func TestPatchAutomationGlobalFailureNeverAuthorizes(t *testing.T) {
	cfg, config := setupPatchGeneration(t)
	t.Setenv("PATCH_GATE_FAIL", "global")
	path, err := RunRemediationAutomation(context.Background(), config, cfg, "")
	if err == nil || path == "" {
		t.Fatal("global failure accepted")
	}
	state := generationState(t, cfg)
	if state.GlobalGatePassed || state.FinalApprovalStatus != "pending" || state.Tickets[0].Status != "accepted" {
		t.Fatalf("unexpected state: %+v", state)
	}
}

func TestPatchAutomationLockBlocksManualOperations(t *testing.T) {
	cfg, config := setupPatchGeneration(t)
	planPath, err := CreateRemediationPlan(config, cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, release, err := reserveRemediationAutomation(filepath.Dir(planPath))
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	var plan RemediationPlan
	if err := ReadJSON(planPath, &plan); err != nil {
		t.Fatal(err)
	}
	if _, err := PrepareRemediationWorktree(cfg, planPath, plan.Tickets[0].ID); err == nil {
		t.Fatal("manual mutation bypassed automation lock")
	}
	if _, err := RunRemediationAutomation(context.Background(), config, cfg, planPath); err == nil {
		t.Fatal("concurrent automation accepted")
	}
}

func TestPatchSourceRejectsUntrackedSymlinkAndOversize(t *testing.T) {
	cfg, _ := setupPatchGeneration(t)
	if _, err := newPatchRequest(cfg, RemediationTicket{AllowedPaths: []string{"other/**"}}, cfg.RepositoryPath, "HEAD", 1); err == nil {
		t.Fatal("generation with no editable context accepted")
	}
	if err := os.WriteFile(filepath.Join(cfg.RepositoryPath, "src", "untracked.go"), []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readPatchSource(cfg.RepositoryPath, "src/untracked.go", 1024); err == nil {
		t.Fatal("untracked source accepted")
	}
	if _, err := readPatchSource(cfg.RepositoryPath, "src/app.go", 1); err == nil {
		t.Fatal("oversize source accepted")
	}
	link := filepath.Join(cfg.RepositoryPath, "src", "linked.go")
	if err := os.Symlink(filepath.Join(cfg.RepositoryPath, "private.txt"), link); err == nil {
		if _, err := readPatchSource(cfg.RepositoryPath, "src/linked.go", 1024); err == nil {
			t.Fatal("symlink source accepted")
		}
	}
}

func TestPatchAutomationRetriesWithoutHumanAndReservesGlobalBudget(t *testing.T) {
	cfg, config := setupPatchGeneration(t)
	calls := 0
	cfg.generationTransport = mockPatchTransport(t, func(response *PatchGenerationResponse) {
		calls++
		if calls == 1 {
			response.Model = "wrong-model"
		}
	})
	path, err := RunRemediationAutomation(context.Background(), config, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	var summary RemediationAutomationSummary
	if err := ReadJSON(path, &summary); err != nil {
		t.Fatal(err)
	}
	if calls != 2 || summary.TotalAttempts != 3 || summary.Status != "awaiting_final_review" {
		t.Fatalf("retry/global failure: %+v calls=%d", summary, calls)
	}
}

func TestPatchAutomationDoesNotReplayUnknownGeneration(t *testing.T) {
	cfg, config := setupPatchGeneration(t)
	planPath, err := CreateRemediationPlan(config, cfg)
	if err != nil {
		t.Fatal(err)
	}
	state := generationState(t, cfg)
	state.Tickets[0].Status = "generating"
	if err := writeJSONAtomic(filepath.Join(filepath.Dir(planPath), "state.json"), state); err != nil {
		t.Fatal(err)
	}
	cfg.generationTransport = patchRoundTripFunc(func(*http.Request) (*http.Response, error) { t.Fatal("unknown generation replayed"); return nil, nil })
	if _, err := RunRemediationAutomation(context.Background(), config, cfg, planPath); err == nil {
		t.Fatal("uncertain interrupted state accepted")
	}
}

func TestPatchGenerationChecksSourceAgainAfterResponse(t *testing.T) {
	cfg, config := setupPatchGeneration(t)
	planPath, err := CreateRemediationPlan(config, cfg)
	if err != nil {
		t.Fatal(err)
	}
	var plan RemediationPlan
	if err := ReadJSON(planPath, &plan); err != nil {
		t.Fatal(err)
	}
	metadataPath, err := PrepareRemediationWorktree(cfg, planPath, plan.Tickets[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	var worktree RemediationWorktree
	if err := ReadJSON(metadataPath, &worktree); err != nil {
		t.Fatal(err)
	}
	original := cfg.generationTransport
	cfg.generationTransport = patchRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		response, err := original.RoundTrip(r)
		if err := os.WriteFile(filepath.Join(worktree.Path, "src", "app.go"), []byte("concurrent edit\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		return response, err
	})
	if _, err := RunRemediationAgent(context.Background(), cfg, planPath, plan.Tickets[0].ID); err == nil {
		t.Fatal("stale patch applied")
	}
	state := generationState(t, cfg)
	if state.BestRef != plan.BaselineCommit {
		t.Fatal("BEST changed after concurrent edit")
	}
}

func TestPatchAutomationDoesNotRetryUncertainGateway(t *testing.T) {
	cfg, config := setupPatchGeneration(t)
	calls := 0
	cfg.generationTransport = patchRoundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, fmt.Errorf("fixture transport interruption")
	})
	if _, err := RunRemediationAutomation(context.Background(), config, cfg, ""); err == nil {
		t.Fatal("uncertain gateway should block")
	}
	state := generationState(t, cfg)
	if calls != 1 || state.TotalAttempts != 1 || state.Tickets[0].Status != "blocked_provider_uncertain" {
		t.Fatalf("unknown result was retried: calls=%d state=%+v", calls, state)
	}
}
