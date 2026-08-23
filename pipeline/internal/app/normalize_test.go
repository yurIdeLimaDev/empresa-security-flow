package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func normalizeFixture(t *testing.T, tool, content string) NormalizedBundle {
	t.Helper()
	path := filepath.Join(t.TempDir(), tool+".out")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	bundle, err := NormalizeFile(NormalizeConfig{Tool: tool, InputPath: path, EngagementID: "eng-001", ScopeID: "scope-001", ToolVersion: "test", SupplyChainStatus: "verified", ObservedAt: "2026-08-19T00:00:00Z"})
	if err != nil {
		t.Fatalf("%s: %v", tool, err)
	}
	return bundle
}

func TestNormalizeAndDeduplicateSecretsAcrossTools(t *testing.T) {
	gitleaks := normalizeFixture(t, "gitleaks", `[{"RuleID":"generic-api-key","Description":"Generic key","File":"src/config.go","StartLine":12,"Secret":"must-not-copy"}]`)
	trufflehog := normalizeFixture(t, "trufflehog", `{"DetectorName":"PrivateKey","Verified":true,"Raw":"must-not-copy","SourceMetadata":{"Data":{"Filesystem":{"file":"src/config.go","line":12}}}}`)
	merged, err := MergeBundles(gitleaks, trufflehog)
	if err != nil {
		t.Fatal(err)
	}
	if len(merged.Findings) != 1 {
		t.Fatalf("esperava 1 achado deduplicado, obteve %d", len(merged.Findings))
	}
	finding := merged.Findings[0]
	if finding.Occurrences != 2 || len(finding.ToolRunIDs) != 2 || len(finding.EvidenceIDs) != 2 {
		t.Fatalf("proveniência perdida: %+v", finding)
	}
	if finding.Status != "candidate" {
		t.Fatalf("scanner não deve validar automaticamente: %s", finding.Status)
	}
	encoded, _ := json.Marshal(merged)
	if strings.Contains(string(encoded), "must-not-copy") {
		t.Fatal("segredo bruto foi copiado para o modelo canônico")
	}
}

func TestNormalizeAndDeduplicateDependenciesAcrossTools(t *testing.T) {
	osv := normalizeFixture(t, "osv-scanner", `{"results":[{"packages":[{"package":{"name":"example.org/lib","ecosystem":"Go"},"version":"1.0.0","vulnerabilities":[{"id":"CVE-2026-0001","summary":"Issue"}]}]}]}`)
	trivy := normalizeFixture(t, "trivy", `{"Results":[{"Target":"go.mod","Type":"gomod","Vulnerabilities":[{"VulnerabilityID":"CVE-2026-0001","PkgName":"example.org/lib","InstalledVersion":"1.0.0","Severity":"HIGH","Title":"Issue"}]}]}`)
	merged, err := MergeBundles(osv, trivy)
	if err != nil {
		t.Fatal(err)
	}
	if len(merged.Findings) != 1 {
		t.Fatalf("esperava 1 CVE deduplicado, obteve %d", len(merged.Findings))
	}
	if merged.Findings[0].Occurrences != 2 {
		t.Fatalf("ocorrências incorretas: %d", merged.Findings[0].Occurrences)
	}
}

func TestHTTPXRedactsQueryAndCreatesRequestLog(t *testing.T) {
	bundle := normalizeFixture(t, "httpx", `{"url":"https://example.test/path?token=secret&id=1","status_code":200,"method":"GET","tech":["Go","nginx"]}`)
	if len(bundle.Assets) != 1 || len(bundle.RequestLogs) != 1 {
		t.Fatalf("inventário incompleto: %+v", bundle.Summary)
	}
	encoded, _ := json.Marshal(bundle)
	if strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "?token=secret") {
		t.Fatal("query sensível não foi sanitizada")
	}
}

func TestJWTToolIsPreservedWithoutInventingFinding(t *testing.T) {
	bundle := normalizeFixture(t, "jwt_tool", `token=header.payload.signature`)
	if len(bundle.Findings) != 0 || len(bundle.Exceptions) != 1 || bundle.Exceptions[0].Blocking {
		t.Fatalf("tratamento jwt_tool incorreto: %+v", bundle)
	}
	if bundle.Exceptions[0].Type != "unsupported_stable_machine_format" {
		t.Fatalf("exceção incorreta: %+v", bundle.Exceptions[0])
	}
}

func TestFindingAdaptersSmoke(t *testing.T) {
	cases := []struct{ tool, input string }{
		{"testssl", `[{"id":"TLS1","severity":"HIGH","finding":"TLS antigo","fqdn":"example.test"}]`},
		{"nuclei", `{"template-id":"exposure","info":{"name":"Exposure","severity":"high"},"matched-at":"https://example.test/a?key=secret"}`},
		{"zap", `{"site":[{"@name":"https://example.test","alerts":[{"pluginid":"1001","name":"Header","riskdesc":"Medium (2)","instances":[{"uri":"https://example.test/a","param":"q"}]}]}]}`},
		{"schemathesis", `<testsuite><testcase name="GET /users" classname="contract"><failure message="500 response">unexpected</failure></testcase></testsuite>`},
		{"sqlmap", `{"url":"https://example.test/u?id=1","parameter":"id","title":"boolean-based blind","dbms":"PostgreSQL"}`},
		{"dalfox", `{"type":"V","url":"https://example.test/search?q=x","param":"q"}`},
		{"semgrep", `{"results":[{"check_id":"go.lang.security","path":"main.go","start":{"line":9},"extra":{"message":"unsafe","severity":"ERROR"}}]}`},
	}
	for _, test := range cases {
		t.Run(test.tool, func(t *testing.T) {
			bundle := normalizeFixture(t, test.tool, test.input)
			if len(bundle.Findings) == 0 {
				t.Fatalf("%s não produziu achado", test.tool)
			}
			for _, finding := range bundle.Findings {
				if finding.Status != "candidate" {
					t.Fatalf("%s validou automaticamente", test.tool)
				}
			}
		})
	}
}

func TestInventoryAdaptersSmoke(t *testing.T) {
	cases := []struct{ tool, input string }{
		{"subfinder", `{"host":"api.example.test","source":"crtsh"}`},
		{"amass", `{"name":"api.example.test","addresses":[{"ip":"192.0.2.10"}]}`},
		{"wappalyzer", `{"url":"https://example.test","technologies":[{"name":"React"}]}`},
	}
	for _, test := range cases {
		t.Run(test.tool, func(t *testing.T) {
			if bundle := normalizeFixture(t, test.tool, test.input); len(bundle.Assets) == 0 {
				t.Fatal("asset ausente")
			}
		})
	}
}

func TestSARIFGenericAdapter(t *testing.T) {
	bundle := normalizeFixture(t, "zap", `{"version":"2.1.0","runs":[{"results":[{"ruleId":"R1","level":"warning","message":{"text":"Issue"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"https://example.test/a"},"region":{"startLine":1}}}]}]}]}`)
	if len(bundle.Findings) != 1 || bundle.Findings[0].Severity != "medium" {
		t.Fatalf("SARIF incorreto: %+v", bundle.Findings)
	}
}

func TestHadrianSARIFAndJSLuiceInventory(t *testing.T) {
	hadrian := normalizeFixture(t, "hadrian", `{"metadata":{"tool":"hadrian","version":"1.0.0","timestamp":"2026-08-21T00:00:00Z"},"summary":{"total_operations":1,"total_templates":1,"total_findings":1,"duration":"1s","by_severity":{"HIGH":1},"by_category":{"API1":1}},"findings":[{"id":"h-1","category":"API1","name":"BOLA entre contas A/B","description":"Conta A acessou recurso da conta B","severity":"HIGH","confidence":0.95,"is_vulnerability":true,"endpoint":"GET /api/items/{id}","method":"GET","attacker_role":"account-a","victim_role":"account-b","evidence":{"request":{"method":"GET","url":"https://example.test/api/items/1","headers":{},"body":""},"response":{"status_code":200,"headers":{},"body":"redacted","body_hash":"abc","size":8,"truncated":false}},"request_ids":["req-1"],"timestamp":"2026-08-21T00:00:00Z"}]}`)
	if len(hadrian.Findings) != 1 || hadrian.Findings[0].Category != "authorization" || hadrian.Findings[0].Status != "candidate" || hadrian.ToolRuns[0].Version != "1.0.0" {
		t.Fatalf("Hadrian JSON v1.0.0 incorreto: %+v", hadrian.Findings)
	}
	validatedHadrian := normalizeFixture(t, "hadrian", `{"metadata":{"tool":"hadrian","version":"1.0.0"},"findings":[{"id":"h-2","category":"API5","name":"BFLA com efeito confirmado","severity":"CRITICAL","confidence":1,"is_vulnerability":true,"endpoint":"POST /api/credits","method":"POST","attacker_role":"user","victim_role":"admin","evidence":{"request":{"method":"POST","url":"https://example.test/api/credits","headers":{},"body":"{}"},"response":{"status_code":200,"headers":{},"body":"","body_hash":"","size":0,"truncated":false},"setup_response":{"status_code":201},"attack_response":{"status_code":200},"verify_response":{"status_code":200}}}]}`)
	if len(validatedHadrian.Findings) != 1 || validatedHadrian.Findings[0].Status != "validated" || validatedHadrian.Findings[0].Exploitability != "effect-verified" {
		t.Fatalf("Hadrian three-phase incorreto: %+v", validatedHadrian.Findings)
	}
	jsluice := normalizeFixture(t, "jsluice", `{"url":"https://example.test/api/items?id=1","method":"GET","filename":"bundle.js"}`)
	if len(jsluice.Assets) != 1 || len(jsluice.RequestLogs) != 1 || jsluice.RequestLogs[0].ScopeDecision != "static_bundle_candidate" {
		t.Fatalf("inventário jsluice incorreto: %+v", jsluice)
	}
}

func TestCompareRequiresCoverageAndSeparatesFixed(t *testing.T) {
	old := normalizeFixture(t, "nuclei", `{"template-id":"x","info":{"name":"X","severity":"high"},"matched-at":"https://example.test"}`)
	current := old
	current.Engagement.ID = "eng-002"
	current.Findings[0].Status = "fixed_verified"
	comparison := CompareBundles(old, current)
	if !comparison.Comparable || len(comparison.FixedVerified) != 1 || len(comparison.Recurring) != 0 {
		t.Fatalf("comparação incorreta: %+v", comparison)
	}
	current.ToolRuns = nil
	comparison = CompareBundles(old, current)
	if comparison.Comparable {
		t.Fatal("comparação sem cobertura foi considerada comparável")
	}
}

func TestSupplyChainFailsClosedAndCanApproveSyntheticBinary(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "tool.bin")
	if err := os.WriteFile(binary, []byte("reviewed-binary"), 0o600); err != nil {
		t.Fatal(err)
	}
	hash, _ := HashFile(binary)
	sbom := filepath.Join(dir, "sbom.json")
	if err := os.WriteFile(sbom, []byte(`{"components":[{"bom-ref":"tool:demo"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	review := filepath.Join(dir, "review.md")
	if err := os.WriteFile(review, []byte("synthetic adoption review"), 0o600); err != nil {
		t.Fatal(err)
	}
	reviewHash, _ := HashFile(review)
	commit := "0123456789abcdef0123456789abcdef01234567"
	lock := ToolLockFile{SchemaVersion: ToolLockSchemaVersion, Policy: SupplyChainPolicy{DenyLatest: true, RequirePinnedVersion: true, RequireCommit: true, RequireArtifactSHA256: true, RequireContainerDigest: false, RequireCodeReview: true, RequireSBOM: true}, Tools: []ToolLock{{Name: "demo", ExecutionClass: "automatic", Release: ReleaseLock{Tag: "v1.0.0", ReleaseID: 1, Commit: commit, SignatureStatus: "verified", ArtifactSHA256: hash}, LocalBinarySHA256: hash, SBOMComponentRef: "tool:demo", CodeReview: CodeReviewLock{Status: "approved", Reviewer: "test", ReviewedAt: "2026-08-19T00:00:00Z", ReviewedCommit: commit, Reference: review, ReferenceSHA256: "sha256:" + reviewHash, Scope: "synthetic test"}, Liveness: &LivenessLock{Command: []string{"--version"}, ExpectedPattern: "1\\.0"}}}}
	report := ValidateToolLock(lock, sbom)
	if !report.Approved {
		t.Fatalf("lock sintético deveria ser aprovado: %+v", report)
	}
	if err := ApproveLocalExecutable(lock, "demo", binary, sbom); err != nil {
		t.Fatal(err)
	}
	lock.Tools[0].Release.Tag = "latest"
	if ValidateToolLock(lock, sbom).Approved {
		t.Fatal("latest deveria ser bloqueado")
	}
}

func TestNormalizeKnownOutputsCarriesRunnerProvenance(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "semgrep.json"), []byte(`{"results":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	logDir := filepath.Join(dir, "execution-logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		t.Fatal(err)
	}
	log := ExecutionLog{
		RunID: "run_provenance", EngagementID: "eng_provenance", Tool: "semgrep",
		ToolDigest: "sha256:" + strings.Repeat("c", 64), ContainerDigest: "sha256:" + strings.Repeat("b", 64),
		PolicyFilesHash: PolicyHashes{
			Engagement: "sha256:" + strings.Repeat("d", 64),
			Tool:       "sha256:" + strings.Repeat("e", 64),
			Matrix:     "not-applicable",
		},
		InputHashes: []PathHash{}, RuntimeArtifacts: []PathHash{}, Network: ExecutionNetworkLog{},
		StartedAt: "2026-08-22T11:59:00Z", EndedAt: "2026-08-22T12:00:00Z",
		Counts: ExecutionCounts{}, Artifacts: []string{"semgrep.json"}, ArtifactHashes: []PathHash{},
	}
	if err := WriteJSON(filepath.Join(logDir, "run_provenance.json"), log); err != nil {
		t.Fatal(err)
	}
	bundle, err := NormalizeKnownOutputs(dir, "eng-provenance", "scope-provenance", []OutputAdapter{{Tool: "semgrep", File: "semgrep.json"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.ToolRuns) != 1 {
		t.Fatalf("ToolRun ausente: %+v", bundle.ToolRuns)
	}
	run := bundle.ToolRuns[0]
	if run.ID != log.RunID || run.ToolDigest != log.ToolDigest || run.PolicySHA256 != HashJSON(log.PolicyFilesHash) || run.SupplyChainStatus != "runner-gated" {
		t.Fatalf("proveniência governada não foi preservada: %+v", run)
	}
}

func TestNormalizerDeduplicatesFiveThousandInputsDeterministically(t *testing.T) {
	base := normalizeFixture(t, "nuclei", `{"template-id":"scale-check","info":{"name":"Scale check","severity":"medium"},"matched-at":"https://example.test/a?id=1"}`)
	bundles := make([]NormalizedBundle, 5000)
	for index := range bundles {
		bundles[index] = base
	}
	merged, err := MergeBundles(bundles...)
	if err != nil {
		t.Fatal(err)
	}
	if len(merged.Findings) != 1 || merged.Findings[0].Occurrences != base.Findings[0].Occurrences*5000 {
		t.Fatalf("deduplicação em escala incorreta: %+v", merged.Summary)
	}
	second, err := MergeBundles(bundles...)
	if err != nil {
		t.Fatal(err)
	}
	merged.GeneratedAt, second.GeneratedAt = "", ""
	if HashJSON(merged) != HashJSON(second) {
		t.Fatal("consolidação em escala não foi determinística")
	}
}

func TestSupplyChainRejectsTamperedRuntimeArtifactHash(t *testing.T) {
	lockPath := testRepoPath("tools.lock.json")
	lock, err := ReadToolLock(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	for index := range lock.Tools {
		if lock.Tools[index].Name == "subfinder" {
			lock.Tools[index].Container.RuntimeArtifacts[0].SHA256 = "sha256:" + strings.Repeat("0", 64)
		}
	}
	report := ValidateToolLock(lock, testRepoPath("sbom.cdx.json"))
	if report.Approved {
		t.Fatal("runtime artifact adulterado foi aprovado")
	}
	found := false
	for _, check := range report.Checks {
		if check.Tool == "subfinder" && strings.Contains(strings.Join(check.Blockers, " "), "hash divergente") {
			found = true
		}
	}
	if !found {
		t.Fatalf("bloqueio de integridade não foi registrado: %+v", report.Checks)
	}
}
