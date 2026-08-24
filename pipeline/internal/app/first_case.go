package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	PythonReferenceProfileVersion = "1.0.0"
	PythonReferenceImage          = "python:3.13-slim-bookworm@sha256:00faa2debb87529f9f0764e9491d8ba400a3678976616c3bd7cb193745ac20d1"
	manualAdapterResultName       = "manual-adapter-result.json"
)

type ManualPatchResult struct {
	SchemaVersion string   `json:"schema_version"`
	Profile       string   `json:"profile"`
	CaseID        string   `json:"case_id"`
	TicketID      string   `json:"ticket_id"`
	PatchPath     string   `json:"patch_path"`
	PatchSHA256   string   `json:"patch_sha256"`
	ChangedPaths  []string `json:"changed_paths"`
	Applied       bool     `json:"applied"`
}

type ReferenceGateSummary struct {
	SchemaVersion   string `json:"schema_version"`
	Profile         string `json:"profile"`
	ProfileVersion  string `json:"profile_version"`
	Gate            string `json:"gate"`
	CandidateCommit string `json:"candidate_commit"`
	ContainerImage  string `json:"container_image"`
	Passed          bool   `json:"passed"`
}

type referenceFindingDefinition struct {
	RuleID      string
	Title       string
	Description string
	Category    string
	Severity    string
	Location    string
	TestName    string
	Active      func(string) bool
}

// RunManualPatchAdapter is the only code-changing adapter in the first-case
// profile. Its contract is intentionally environment-only because the parent
// remediation runner supplies and hashes the immutable command template.
func RunManualPatchAdapter() error {
	caseID := strings.TrimSpace(os.Getenv("REMEDIATION_CASE_ID"))
	ticketID := strings.TrimSpace(os.Getenv("REMEDIATION_TICKET_ID"))
	worktreeInput := strings.TrimSpace(os.Getenv("REMEDIATION_WORKTREE"))
	patchRootInput := strings.TrimSpace(os.Getenv("REMEDIATION_PATCH_ROOT"))
	outputInput := strings.TrimSpace(os.Getenv("REMEDIATION_ATTEMPT_OUTPUT"))
	allowedJSON := os.Getenv("REMEDIATION_ALLOWED_PATHS_JSON")
	if !caseIDPattern.MatchString(caseID) || !strings.HasPrefix(ticketID, "REM-") || !caseIDPattern.MatchString(ticketID) {
		return fmt.Errorf("case_id ou ticket_id inválido")
	}
	if worktreeInput == "" || patchRootInput == "" || outputInput == "" || allowedJSON == "" {
		return fmt.Errorf("contrato do adaptador manual incompleto")
	}
	var allowed []string
	if err := json.Unmarshal([]byte(allowedJSON), &allowed); err != nil || len(allowed) == 0 {
		return fmt.Errorf("REMEDIATION_ALLOWED_PATHS_JSON inválido")
	}
	for _, pattern := range allowed {
		if !safeRelativePattern(pattern) {
			return fmt.Errorf("allowed path inseguro: %s", pattern)
		}
	}
	worktree, err := canonicalRepository(worktreeInput)
	if err != nil {
		return err
	}
	patchRoot, err := secureExistingDirectory(patchRootInput)
	if err != nil {
		return fmt.Errorf("patch root: %w", err)
	}
	outputRoot, err := secureExistingDirectory(outputInput)
	if err != nil {
		return fmt.Errorf("attempt output: %w", err)
	}
	expected := filepath.Join(patchRoot, caseID, ticketID+".patch")
	patchPath, err := filepath.Abs(expected)
	if err != nil || !pathWithin(patchRoot, patchPath) {
		return fmt.Errorf("patch fora do patch root")
	}
	resolvedPatch, err := filepath.EvalSymlinks(patchPath)
	if err != nil {
		return fmt.Errorf("patch obrigatório ausente: %w", err)
	}
	if !samePath(resolvedPatch, patchPath) || !pathWithin(patchRoot, resolvedPatch) {
		return fmt.Errorf("patch simbólico ou fora do patch root é proibido")
	}
	info, err := os.Stat(resolvedPatch)
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("patch precisa ser arquivo regular")
	}
	status, err := gitOutput(worktree, "status", "--porcelain")
	if err != nil || strings.TrimSpace(status) != "" {
		return fmt.Errorf("worktree precisa estar limpo antes do patch")
	}
	paths, err := patchChangedPaths(worktree, resolvedPatch)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return fmt.Errorf("patch não altera arquivo")
	}
	for _, item := range paths {
		if !safeRelativePath(item) || !matchesAnyPath(item, allowed) {
			return fmt.Errorf("patch altera caminho fora do ticket: %s", item)
		}
	}
	if _, err := gitOutput(worktree, "apply", "--check", "--whitespace=error-all", "--", resolvedPatch); err != nil {
		return fmt.Errorf("git apply --check recusou o patch: %w", err)
	}
	if _, err := gitOutput(worktree, "apply", "--whitespace=error-all", "--", resolvedPatch); err != nil {
		return fmt.Errorf("git apply recusou o patch: %w", err)
	}
	changed, err := gitOutput(worktree, "status", "--porcelain", "-z")
	if err != nil {
		return err
	}
	actual := porcelainPaths(changed)
	if strings.Join(actual, "\n") != strings.Join(paths, "\n") {
		return fmt.Errorf("arquivos alterados divergem do patch validado")
	}
	patchHash, err := HashFile(resolvedPatch)
	if err != nil {
		return err
	}
	relativePatch, err := filepath.Rel(patchRoot, resolvedPatch)
	if err != nil || !safeRelativePath(relativePatch) {
		return fmt.Errorf("não foi possível registrar caminho relativo do patch")
	}
	result := ManualPatchResult{
		SchemaVersion: RemediationSchemaVersion, Profile: "python-3.13-stdlib",
		CaseID: caseID, TicketID: ticketID, PatchPath: filepath.ToSlash(relativePatch),
		PatchSHA256: patchHash, ChangedPaths: actual, Applied: true,
	}
	return WriteJSON(filepath.Join(outputRoot, manualAdapterResultName), result)
}

func safeRelativePattern(value string) bool {
	value = filepath.ToSlash(strings.TrimSpace(value))
	if value == "" || strings.HasPrefix(value, "/") || filepath.IsAbs(value) {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(value))
	return clean != ".." && !strings.HasPrefix(clean, "../") && !strings.Contains(clean, "/../")
}

func secureExistingDirectory(value string) (string, error) {
	abs, err := filepath.Abs(value)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	if !samePath(abs, resolved) {
		return "", fmt.Errorf("diretório simbólico não é permitido")
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("diretório ausente")
	}
	return resolved, nil
}

func patchChangedPaths(repository, patch string) ([]string, error) {
	output, err := gitOutput(repository, "apply", "--numstat", "-z", "--", patch)
	if err != nil {
		return nil, fmt.Errorf("não foi possível inspecionar patch: %w", err)
	}
	items := []string{}
	for _, record := range strings.Split(output, "\x00") {
		fields := strings.SplitN(record, "\t", 3)
		if len(fields) != 3 {
			continue
		}
		path := canonicalRepoPath(fields[2])
		if !safeRelativePath(path) {
			return nil, fmt.Errorf("patch contém caminho inseguro: %s", path)
		}
		items = append(items, path)
	}
	return dedupeSortedCase(items), nil
}

func porcelainPaths(output string) []string {
	items := []string{}
	for _, record := range strings.Split(output, "\x00") {
		if len(record) < 4 {
			continue
		}
		path := strings.TrimSpace(record[3:])
		if strings.Contains(path, " -> ") {
			parts := strings.SplitN(path, " -> ", 2)
			path = parts[1]
		}
		items = append(items, canonicalRepoPath(path))
	}
	return dedupeSortedCase(items)
}

// RunPythonReferenceGate executes the first-case profile without downloading
// anything. The image must already exist by exact digest; deploy preflight is
// responsible for provisioning it.
func RunPythonReferenceGate(ctx context.Context, gate, repository, outputDir, ticketID, candidateRef, toolLockPath, sbomPath string) error {
	if gate != "quality" && gate != "retest" && gate != "security-full" {
		return fmt.Errorf("gate desconhecido: %s", gate)
	}
	repository, err := canonicalRepository(repository)
	if err != nil {
		return err
	}
	outputDir, err = secureExistingDirectory(outputDir)
	if err != nil {
		return err
	}
	commit, err := resolveCommit(repository, candidateRef)
	if err != nil {
		return err
	}
	head, err := resolveCommit(repository, "HEAD")
	if err != nil || head != commit {
		return fmt.Errorf("candidate_ref precisa ser o HEAD do worktree")
	}
	status, err := gitOutput(repository, "status", "--porcelain")
	if err != nil || strings.TrimSpace(status) != "" {
		return fmt.Errorf("gate exige worktree limpo")
	}
	approvedRuntime, err := approvePythonReferenceRuntime(toolLockPath, sbomPath)
	if err != nil {
		return fmt.Errorf("runtime Python não foi aprovado pela cadeia de ferramentas: %w", err)
	}
	lockedImage := approvedRuntime.Container.Image + "@" + normalizeDigest(approvedRuntime.Container.Digest)
	if lockedImage != PythonReferenceImage {
		return fmt.Errorf("runtime Python do perfil diverge do lock aprovado")
	}
	if gate == "quality" {
		if err := runPythonProfileTests(ctx, repository, []string{"discover", "-s", "tests", "-p", "test_quality.py"}); err != nil {
			return err
		}
		return WriteJSON(filepath.Join(outputDir, "quality-result.json"), ReferenceGateSummary{RemediationSchemaVersion, "python-3.13-stdlib", PythonReferenceProfileVersion, gate, commit, PythonReferenceImage, true})
	}
	if gate == "retest" {
		testName, err := referenceRetestName(ticketID)
		if err != nil {
			return err
		}
		if err := runPythonProfileTests(ctx, repository, []string{testName}); err != nil {
			return err
		}
		return WriteJSON(filepath.Join(outputDir, "retest-result.json"), ReferenceGateSummary{RemediationSchemaVersion, "python-3.13-stdlib", PythonReferenceProfileVersion, gate, commit, PythonReferenceImage, true})
	}
	bundle, err := buildPythonReferenceBundle(repository, commit)
	if err != nil {
		return err
	}
	return WriteBundle(filepath.Join(outputDir, "candidate-bundle.json"), bundle)
}

func approvePythonReferenceRuntime(toolLockPath, sbomPath string) (*ToolLock, error) {
	lock, lockAbs, err := ResolveToolLock(toolLockPath)
	if err != nil {
		return nil, err
	}
	if lock.SchemaVersion != ToolLockSchemaVersion || !lock.Policy.RequireSBOM || !lock.Policy.RequireContainerDigest || !lock.Policy.RequireCodeReview {
		return nil, fmt.Errorf("política de supply chain insuficiente")
	}
	var selected *ToolLock
	for index := range lock.Tools {
		if normalizeToolName(lock.Tools[index].Name) == "sqlmap" {
			selected = &lock.Tools[index]
			break
		}
	}
	if selected == nil || selected.ExecutionClass != "automatic" || selected.CodeReview.Status != "approved" || selected.Container == nil || selected.SBOMComponentRef == "" {
		return nil, fmt.Errorf("entrada aprovada do runtime ausente")
	}
	if !validSHA256Hex(strings.TrimPrefix(selected.Container.Digest, "sha256:")) || strings.Contains(strings.ToLower(selected.Container.Image), ":latest") {
		return nil, fmt.Errorf("imagem sem digest exato")
	}
	var sbom struct {
		Components []struct {
			BOMRef string `json:"bom-ref"`
		} `json:"components"`
	}
	data, err := os.ReadFile(sbomPath)
	if err != nil || json.Unmarshal(data, &sbom) != nil {
		return nil, fmt.Errorf("SBOM ausente ou inválido")
	}
	found := false
	for _, component := range sbom.Components {
		if component.BOMRef == selected.SBOMComponentRef {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("runtime ausente do SBOM")
	}
	lockRoot := filepath.Dir(lockAbs)
	reference := filepath.Clean(filepath.Join(lockRoot, filepath.FromSlash(selected.CodeReview.Reference)))
	if hash, hashErr := HashFile(reference); hashErr != nil || !strings.EqualFold(hash, strings.TrimPrefix(selected.CodeReview.ReferenceSHA256, "sha256:")) {
		return nil, fmt.Errorf("revisão de adoção sem integridade")
	}
	for _, artifact := range selected.Container.RuntimeArtifacts {
		path := filepath.Clean(filepath.Join(lockRoot, filepath.FromSlash(artifact.Path)))
		if !pathWithin(lockRoot, path) {
			return nil, fmt.Errorf("runtime artifact fora da raiz do pipeline")
		}
		hash, hashErr := HashPath(path)
		if hashErr != nil || !strings.EqualFold(hash, strings.TrimPrefix(artifact.SHA256, "sha256:")) {
			return nil, fmt.Errorf("runtime artifact sem integridade")
		}
	}
	return selected, nil
}

func runPythonProfileTests(ctx context.Context, repository string, testArgs []string) error {
	docker, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("Docker ausente")
	}
	if output, err := exec.CommandContext(ctx, docker, "image", "inspect", PythonReferenceImage, "--format", "{{.Id}}").CombinedOutput(); err != nil {
		return fmt.Errorf("imagem de referência exata ausente; download em gate é proibido: %s", strings.TrimSpace(string(output)))
	}
	args := []string{
		"run", "--rm", "--network", "none", "--cap-drop", "ALL", "--security-opt", "no-new-privileges=true",
		"--read-only", "--pids-limit", "128", "--memory", "512m", "--cpus", "1",
		"--tmpfs", "/tmp:rw,noexec,nosuid,nodev,size=64m", "-e", "PYTHONDONTWRITEBYTECODE=1",
		"--mount", "type=bind,src=" + repository + ",dst=/repo,readonly", "-w", "/repo", PythonReferenceImage,
		"python", "-B", "-m", "unittest",
	}
	args = append(args, testArgs...)
	command := exec.CommandContext(ctx, docker, args...)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		return fmt.Errorf("gate interrompido: %w", ctx.Err())
	}
	if err != nil {
		return fmt.Errorf("testes do perfil falharam: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func referenceFindingDefinitions() []referenceFindingDefinition {
	contains := func(needle string) func(string) bool {
		return func(value string) bool { return strings.Contains(value, needle) }
	}
	missing := func(needle string) func(string) bool {
		return func(value string) bool { return !strings.Contains(value, needle) }
	}
	return []referenceFindingDefinition{
		{"LAB-BOLA-001", "Autorização entre contas ausente", "A identidade A consegue ler o recurso da identidade B.", "authorization", "high", "app/service.py", "tests.test_security.SecurityProperties.test_account_b_cannot_read_account_a", contains("return RECORDS[resource_id]  # LAB_BOLA")},
		{"LAB-HEADER-001", "Header de proteção ausente", "A resposta não define X-Content-Type-Options.", "configuration", "low", "app/service.py", "tests.test_security.SecurityProperties.test_security_header", missing("X-Content-Type-Options")},
		{"LAB-TOKEN-001", "Regra de token fraca", "O token é aceito quando é igual ao identificador do usuário.", "authentication", "high", "app/service.py", "tests.test_security.SecurityProperties.test_forged_token_is_rejected", contains("token == user  # LAB_WEAK_TOKEN")},
		{"LAB-DATA-001", "Configuração de dados permissiva", "O modo de armazenamento concede escrita além do proprietário.", "configuration", "medium", "app/config.py", "tests.test_security.SecurityProperties.test_data_mode_is_private", contains("0o666  # LAB_PUBLIC_DATA")},
		{"LAB-SECRET-001", "Segredo fictício no bundle", "Um marcador de segredo fictício foi incluído no bundle estático.", "secrets", "medium", "static/bundle.js", "tests.test_security.SecurityProperties.test_bundle_has_no_fake_secret", contains("LAB_FAKE_SECRET_")},
	}
}

func referenceFingerprint(def referenceFindingDefinition) string {
	return HashBytes([]byte(strings.Join([]string{def.RuleID, def.Category, def.Location}, "|")))
}

func referenceRetestName(ticketID string) (string, error) {
	if ticketID == "" {
		return "tests.test_security.SecurityProperties", nil
	}
	for _, def := range referenceFindingDefinitions() {
		if ticketID == "REM-"+strings.ToUpper(referenceFingerprint(def)[:12]) {
			return def.TestName, nil
		}
	}
	return "", fmt.Errorf("ticket não pertence ao perfil python de referência")
}

func buildPythonReferenceBundle(repository, commit string) (NormalizedBundle, error) {
	runID := "run-reference-security"
	scope := Scope{ID: "scope_python_reference_lab", BaseURL: "http://python-reference-lab.invalid", AllowedSchemes: []string{"http"}, AllowedHosts: []string{"python-reference-lab.invalid"}, AllowedPorts: []int{8080}}
	scope.Fingerprint = scopeFingerprint(scope)
	policyHash := HashBytes([]byte("python-reference-profile|" + PythonReferenceProfileVersion))
	bundle := newNormalizedBundle("eng_python_reference_lab", scope.ID)
	bundle.GeneratedAt = "2026-08-23T00:00:00Z"
	bundle.Scope = scope
	bundle.Engagement = Engagement{ID: "eng_python_reference_lab", Name: "laboratório Python controlado", Pipeline: 2, ScopeID: scope.ID, Status: "completed", ToolRunIDs: []string{runID}}
	bundle.ToolRuns = []ToolRun{{ID: runID, Tool: "empresa-security-reference-gate", Version: PythonReferenceProfileVersion, Status: "success", InputPath: "git:" + commit, InputSHA256: HashBytes([]byte(commit)), Parser: "python-reference-profile", ParserVersion: PythonReferenceProfileVersion, ToolDigest: "sha256:" + strings.TrimPrefix(PythonReferenceImage[strings.Index(PythonReferenceImage, "@")+1:], "sha256:"), PolicySHA256: policyHash, CommandSHA256: HashBytes([]byte("security-full")), SupplyChainStatus: "runner-gated"}}
	definitions := referenceFindingDefinitions()
	for _, def := range definitions {
		data, err := os.ReadFile(filepath.Join(repository, filepath.FromSlash(def.Location)))
		if err != nil {
			return bundle, err
		}
		if !def.Active(string(data)) {
			continue
		}
		fingerprint := referenceFingerprint(def)
		evidenceHash := HashBytes([]byte(def.RuleID + "|active"))
		evidenceID := stableID("evidence", fingerprint)
		bundle.Evidence = append(bundle.Evidence, Evidence{ID: evidenceID, Kind: "security-property", Path: "sanitized/" + def.RuleID + ".json", SHA256: evidenceHash, Locator: def.Location, CollectedAt: "2026-08-23T00:00:00Z", ToolRunID: runID, IntegrityStatus: "verified", ReviewStatus: "validated", Sensitive: false})
		severityScore := map[string]int{"low": 20, "medium": 40, "high": 70, "critical": 90}[def.Severity]
		bundle.Findings = append(bundle.Findings, Finding{ID: stableID("finding", fingerprint), Fingerprint: fingerprint, Title: def.Title, Description: def.Description, Category: def.Category, Severity: def.Severity, Confidence: "confirmed", Status: "validated", RuleID: def.RuleID, EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: def.Location, Exposure: "internal-lab", Exploitability: "demonstrated", Impact: "security property violated", Remediation: "Aplicar somente o patch governado do ticket e repetir o teste de propriedade.", FirstObserved: "2026-08-23T00:00:00Z", LastObserved: "2026-08-23T00:00:00Z", Occurrences: 1, Rank: RankBreakdown{Severity: severityScore, Confidence: 20, Evidence: 20, Exposure: 5, Exploitability: 15, BusinessImpact: 10, Total: severityScore + 70}, Commercial: CommercialValue{Relevance: "security", ProofLevel: "validated", Statement: "Falha reproduzida no laboratório controlado.", Limitation: "Evidência válida somente para o commit indicado."}})
	}
	sort.Slice(bundle.Findings, func(i, j int) bool { return bundle.Findings[i].Fingerprint < bundle.Findings[j].Fingerprint })
	sort.Slice(bundle.Evidence, func(i, j int) bool { return bundle.Evidence[i].ID < bundle.Evidence[j].ID })
	return bundle, nil
}

func BuildPythonReferenceBaseline(repository, output string) error {
	repository, err := canonicalRepository(repository)
	if err != nil {
		return err
	}
	commit, err := resolveCommit(repository, "HEAD")
	if err != nil {
		return err
	}
	bundle, err := buildPythonReferenceBundle(repository, commit)
	if err != nil {
		return err
	}
	if len(bundle.Findings) != len(referenceFindingDefinitions()) {
		return fmt.Errorf("laboratório baseline não contém todos os achados intencionais")
	}
	return WriteBundle(output, bundle)
}

func PythonReferenceProfileConfig(repository, baselineBundle, outputRoot, runnerPath string) (RemediationConfig, error) {
	repository, err := canonicalRepository(repository)
	if err != nil {
		return RemediationConfig{}, err
	}
	runnerPath, err = filepath.Abs(runnerPath)
	if err != nil {
		return RemediationConfig{}, err
	}
	runnerHash, err := HashFile(runnerPath)
	if err != nil {
		return RemediationConfig{}, err
	}
	baselineRef, err := resolveCommit(repository, "HEAD")
	if err != nil {
		return RemediationConfig{}, err
	}
	pipelineRoot := filepath.Dir(filepath.Dir(runnerPath))
	toolLockPath := filepath.Join(pipelineRoot, "tools.lock.json")
	sbomPath := filepath.Join(pipelineRoot, "sbom.cdx.json")
	if _, err := os.Stat(toolLockPath); err != nil {
		return RemediationConfig{}, fmt.Errorf("tools.lock.json não encontrado ao lado do runner")
	}
	if _, err := os.Stat(sbomPath); err != nil {
		return RemediationConfig{}, fmt.Errorf("sbom.cdx.json não encontrado ao lado do runner")
	}
	defs := referenceFindingDefinitions()
	overrides := make([]FindingRemediationOverride, 0, len(defs))
	for index, def := range defs {
		overrides = append(overrides, FindingRemediationOverride{FindingFingerprint: referenceFingerprint(def), RequiresCodeChange: true, AllowedPaths: []string{def.Location}, DependsOnFingerprints: []string{}, RetestGateIDs: []string{"retest-property"}, Priority: index + 1})
	}
	gate := func(id, class, artifact string) RemediationGate {
		return RemediationGate{ID: id, Class: class, Command: []string{runnerPath, "reference-gate", "--kind", map[string]string{"quality-project": "quality", "security-full": "security-full", "retest-property": "retest"}[id], "--repository", "{repository}", "--output", "{gate_output}", "--ticket", "{ticket_id}", "--candidate-ref", "{candidate_ref}", "--tool-lock", toolLockPath, "--sbom", sbomPath}, ExecutableSHA256: runnerHash, TimeoutSeconds: 300, Required: true, ArtifactPaths: []string{artifact}}
	}
	return RemediationConfig{SchemaVersion: RemediationSchemaVersion, CaseID: "LAB-PYTHON-001", RepositoryPath: repository, BaselineRef: baselineRef, BaselineBundlePath: baselineBundle, OutputRoot: outputRoot, SecurityChangesOnly: true, HumanReviewStage: "final", Limits: RemediationLimits{MaxAttemptsPerTicket: 2, MaxTotalAttempts: 12}, ChangePolicy: RemediationChangePolicy{DefaultAllowedPaths: []string{"app/**", "static/**"}, TestPaths: []string{"tests/**"}, ProtectedPaths: []string{"security/**", "tests/**", ".github/**", "pyproject.toml"}, MaxFilesChanged: 3, MaxChangedLines: 120, AllowDependencyChange: false, AllowBinaryChanges: false}, CandidateBundleGateID: "security-full", DefaultRetestGateIDs: []string{"retest-property"}, EnvironmentAllowlist: []string{"REMEDIATION_PATCH_ROOT"}, Agent: RemediationAgent{Command: []string{runnerPath, "remediation-apply-patch"}, ExecutableSHA256: runnerHash, TimeoutSeconds: 60, AutoCommit: true}, Gates: []RemediationGate{gate("quality-project", "quality", "quality-result.json"), gate("retest-property", "retest", "retest-result.json"), gate("security-full", "security", "candidate-bundle.json")}, FindingOverrides: overrides}, nil
}
