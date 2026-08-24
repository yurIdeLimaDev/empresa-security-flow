package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManualPatchAdapterAppliesOnlyExactScopedPatch(t *testing.T) {
	repository := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repository, "app"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "app", "service.py"), []byte("secure = False\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitForTest(t, repository, "init")
	runGitForTest(t, repository, "add", ".")
	runGitForTest(t, repository, "-c", "user.name=Test", "-c", "user.email=test@localhost", "commit", "-m", "baseline")
	patchRoot := t.TempDir()
	caseID, ticketID := "LAB-PYTHON-001", "REM-ABCDEF123456"
	casePatchRoot := filepath.Join(patchRoot, caseID)
	if err := os.MkdirAll(casePatchRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	patch := "diff --git a/app/service.py b/app/service.py\nindex 180758e..735b2b4 100644\n--- a/app/service.py\n+++ b/app/service.py\n@@ -1 +1 @@\n-secure = False\n+secure = True\n"
	if err := os.WriteFile(filepath.Join(casePatchRoot, ticketID+".patch"), []byte(patch), 0o600); err != nil {
		t.Fatal(err)
	}
	output := t.TempDir()
	t.Setenv("REMEDIATION_CASE_ID", caseID)
	t.Setenv("REMEDIATION_TICKET_ID", ticketID)
	t.Setenv("REMEDIATION_WORKTREE", repository)
	t.Setenv("REMEDIATION_PATCH_ROOT", patchRoot)
	t.Setenv("REMEDIATION_ATTEMPT_OUTPUT", output)
	t.Setenv("REMEDIATION_ALLOWED_PATHS_JSON", `["app/service.py"]`)
	if err := RunManualPatchAdapter(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(repository, "app", "service.py"))
	if err != nil || strings.ReplaceAll(string(data), "\r\n", "\n") != "secure = True\n" {
		t.Fatalf("patch não aplicado: err=%v data=%q", err, data)
	}
	var result ManualPatchResult
	if err := ReadJSON(filepath.Join(output, manualAdapterResultName), &result); err != nil || !result.Applied || len(result.ChangedPaths) != 1 || result.ChangedPaths[0] != "app/service.py" {
		t.Fatalf("resultado inválido: err=%v result=%+v", err, result)
	}
}

func TestManualPatchAdapterRejectsPathOutsideTicket(t *testing.T) {
	repository := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repository, "docs"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "docs", "readme.md"), []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitForTest(t, repository, "init")
	runGitForTest(t, repository, "add", ".")
	runGitForTest(t, repository, "-c", "user.name=Test", "-c", "user.email=test@localhost", "commit", "-m", "baseline")
	patchRoot := t.TempDir()
	caseID, ticketID := "LAB-PYTHON-001", "REM-ABCDEF123456"
	if err := os.MkdirAll(filepath.Join(patchRoot, caseID), 0o700); err != nil {
		t.Fatal(err)
	}
	patch := "diff --git a/docs/readme.md b/docs/readme.md\nindex 3367afd..3e75765 100644\n--- a/docs/readme.md\n+++ b/docs/readme.md\n@@ -1 +1 @@\n-old\n+new\n"
	if err := os.WriteFile(filepath.Join(patchRoot, caseID, ticketID+".patch"), []byte(patch), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REMEDIATION_CASE_ID", caseID)
	t.Setenv("REMEDIATION_TICKET_ID", ticketID)
	t.Setenv("REMEDIATION_WORKTREE", repository)
	t.Setenv("REMEDIATION_PATCH_ROOT", patchRoot)
	t.Setenv("REMEDIATION_ATTEMPT_OUTPUT", t.TempDir())
	t.Setenv("REMEDIATION_ALLOWED_PATHS_JSON", `["app/**"]`)
	if err := RunManualPatchAdapter(); err == nil || !strings.Contains(err.Error(), "fora do ticket") {
		t.Fatalf("patch fora do escopo deveria falhar: %v", err)
	}
	if status := runGitForTest(t, repository, "status", "--porcelain"); strings.TrimSpace(status) != "" {
		t.Fatalf("worktree foi alterado apesar da recusa: %s", status)
	}
}

func TestPythonReferenceBundleAndNonRegression(t *testing.T) {
	repository := t.TempDir()
	files := map[string]string{
		"app/service.py":   "def read_record(user, resource_id):\n    return RECORDS[resource_id]  # LAB_BOLA\ndef headers():\n    return {}\ndef token_ok(token, user):\n    return token == user  # LAB_WEAK_TOKEN\n",
		"app/config.py":    "DATA_MODE = 0o666  # LAB_PUBLIC_DATA\n",
		"static/bundle.js": "const token = 'LAB_FAKE_SECRET_NOT_REAL';\n",
	}
	for path, content := range files {
		full := filepath.Join(repository, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runGitForTest(t, repository, "init")
	runGitForTest(t, repository, "add", ".")
	runGitForTest(t, repository, "-c", "user.name=Test", "-c", "user.email=test@localhost", "commit", "-m", "baseline")
	baseline, err := buildPythonReferenceBundle(repository, strings.TrimSpace(runGitForTest(t, repository, "rev-parse", "HEAD")))
	if err != nil || len(baseline.Findings) != 5 {
		t.Fatalf("baseline inesperado: err=%v findings=%d", err, len(baseline.Findings))
	}
	service := files["app/service.py"]
	service = strings.Replace(service, "return RECORDS[resource_id]  # LAB_BOLA", "record = RECORDS[resource_id]\n    if record['owner'] != user:\n        raise PermissionError('forbidden')\n    return record", 1)
	if err := os.WriteFile(filepath.Join(repository, "app", "service.py"), []byte(service), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitForTest(t, repository, "add", ".")
	runGitForTest(t, repository, "-c", "user.name=Test", "-c", "user.email=test@localhost", "commit", "-m", "fix")
	candidate, err := buildPythonReferenceBundle(repository, strings.TrimSpace(runGitForTest(t, repository, "rev-parse", "HEAD")))
	if err != nil || len(candidate.Findings) != 4 {
		t.Fatalf("candidate inesperado: err=%v findings=%d", err, len(candidate.Findings))
	}
	target := referenceFingerprint(referenceFindingDefinitions()[0])
	decision := evaluateSecurityPosture(baseline, candidate, target, false)
	if !decision.Comparable || !decision.TargetResolved || !decision.Improved || len(decision.NewFindings) != 0 || len(decision.Regressions) != 0 {
		t.Fatalf("candidato deveria melhorar monotonicamente: %+v", decision)
	}
}
