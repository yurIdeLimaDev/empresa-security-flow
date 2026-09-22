package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatchDisclosureBlocksBeforeTransport(t *testing.T) {
	credential := "synthetic-gateway-credential-not-valid"
	t.Setenv("PATCH_DISCLOSURE_TEST_KEY", credential)
	cfg := RemediationConfig{Agent: RemediationAgent{Generation: &PatchGenerationConfig{
		APIKeyEnv: "PATCH_DISCLOSURE_TEST_KEY", Endpoint: "https://example.invalid/patch", MaxResponseBytes: 1024,
	}}}
	calls := 0
	cfg.generationTransport = patchRoundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	})
	for name, secret := range map[string]string{
		"private key":    "-----BEGIN " + "PRIVATE KEY-----",
		"api token":      "sk-" + strings.Repeat("synthetic", 4),
		"github token":   "ghp_" + strings.Repeat("a", 24),
		"fine grained":   "github_pat_" + strings.Repeat("a", 24),
		"aws key":        "AKIA" + strings.Repeat("A", 16),
		"jwt":            "eyJ" + strings.Repeat("A", 12) + "." + strings.Repeat("B", 12) + "." + strings.Repeat("C", 12),
		"password":       `const password = "synthetic-password";`,
		"client secret":  `const clientSecret = "synthetic-value";`,
		"dsn":            "postgres://fixture:synthetic-password@example.invalid/db",
		"authorization":  "Bearer synthetic-fixture-value",
		"basic":          "Basic " + strings.Repeat("A", 20),
		"own credential": credential,
	} {
		t.Run(name, func(t *testing.T) {
			// Exercise nested metadata as well as source, including JSON escapes.
			for _, field := range []string{"files_untrusted", "acceptance_criteria", "finding_description_untrusted", "previous_outcome"} {
				data, _ := json.Marshal(map[string]any{field: []any{map[string]string{"content": secret}}})
				before := calls
				_, err := callPatchGateway(context.Background(), cfg, data)
				if err == nil || calls != before || strings.Contains(err.Error(), secret) {
					t.Fatal("sensitive request reached transport or disclosed a value in its error")
				}
			}
		})
	}
	// The guard must not turn runtime variable lookup or an ordinary hash into
	// a detected credential; it also must leave the transmitted bytes intact.
	safe := []byte(`{"files_untrusted":[{"content":"const password = process.env.PASSWORD; const hash = 'abcdef0123456789';"}]}`)
	cfg.generationTransport = patchRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		actual, _ := io.ReadAll(r.Body)
		if !bytes.Equal(actual, safe) {
			t.Fatal("guard rewrote source or JSON")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	})
	if _, err := callPatchGateway(context.Background(), cfg, safe); err != nil || calls != 1 {
		t.Fatal("ordinary source should pass unchanged")
	}
	if _, err := callPatchGateway(context.Background(), cfg, []byte(`{"broken":`)); err == nil || calls != 1 {
		t.Fatal("malformed JSON reached transport")
	}
}

func TestPatchAutomationBlocksCredentialDisclosure(t *testing.T) {
	for _, location := range []string{"source", "finding"} {
		t.Run(location, func(t *testing.T) {
			cfg, config := setupPatchGeneration(t)
			secret := "synthetic-password-never-valid"
			if location == "source" {
				path := filepath.Join(cfg.RepositoryPath, "src", "app.go")
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				data = append(data, []byte(fmt.Sprintf("const password = %q\n", secret))...)
				if err := os.WriteFile(path, data, 0o600); err != nil {
					t.Fatal(err)
				}
				runGitForTest(t, cfg.RepositoryPath, "add", "src/app.go")
				runGitForTest(t, cfg.RepositoryPath, "commit", "-m", "synthetic credential fixture")
			} else {
				bundle, err := ReadBundle(cfg.BaselineBundlePath)
				if err != nil {
					t.Fatal(err)
				}
				bundle.Findings[0].Description = fmt.Sprintf("observed password = %q", secret)
				if err := WriteBundle(cfg.BaselineBundlePath, bundle); err != nil {
					t.Fatal(err)
				}
			}
			baseline, _ := resolveCommit(cfg.RepositoryPath, "HEAD")
			calls := 0
			cfg.generationTransport = patchRoundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, fmt.Errorf("transport should never be reached")
			})
			_, err := RunRemediationAutomation(context.Background(), config, cfg, "")
			if err == nil || calls != 0 {
				t.Fatal("automation did not block before disclosure")
			}
			state := generationState(t, cfg)
			if state.BestRef != baseline || state.GlobalGatePassed || state.TotalAttempts > cfg.Limits.MaxTotalAttempts {
				t.Fatal("blocked context changed BEST or bypassed the attempt limit")
			}
			paths, _ := filepath.Glob(filepath.Join(cfg.OutputRoot, cfg.CaseID, "remediation", "agents", "*", "*", "agent.log"))
			if len(paths) == 0 {
				t.Fatal("missing rejection log")
			}
			for _, path := range paths {
				data, err := os.ReadFile(path)
				if err != nil || bytes.Contains(data, []byte(secret)) {
					t.Fatal("rejection log missing or contains the synthetic credential")
				}
			}
		})
	}
}
