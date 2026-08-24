package app

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const OnboardingSchemaVersion = "1.0.0"

type OnboardingAuthorizationInput struct {
	Confirmed         bool   `json:"confirmed"`
	ConfirmedAt       string `json:"confirmed_at"`
	ConfirmedBy       string `json:"confirmed_by"`
	DocumentReference string `json:"document_reference"`
	SOWPath           string `json:"sow_path"`
	SOWSHA256         string `json:"sow_sha256"`
	EmergencyContact  string `json:"emergency_contact"`
	PaymentStatus     string `json:"payment_status"`
}

type OnboardingAdapterInput struct {
	RunnerPath            string `json:"runner_path"`
	RunnerSHA256          string `json:"runner_sha256"`
	RepositoryPath        string `json:"repository_path"`
	BaselineRef           string `json:"baseline_ref"`
	BaselineBundlePath    string `json:"baseline_bundle_path"`
	RemediationOutputRoot string `json:"remediation_output_root"`
}

type OnboardingInput struct {
	SchemaVersion      string                       `json:"schema_version"`
	CaseID             string                       `json:"case_id"`
	EngagementID       string                       `json:"engagement_id"`
	Domain             string                       `json:"domain"`
	BaseURL            string                       `json:"base_url"`
	OutputRoot         string                       `json:"output_root"`
	Target             NetworkTarget                `json:"target"`
	MaxDownloadBytes   int                          `json:"max_download_bytes"`
	ToolLockPath       string                       `json:"tool_lock_path"`
	PublicContactPaths []string                     `json:"public_contact_paths"`
	PublicObservation  PublicObservationPolicy      `json:"public_observation"`
	ExternalAPIs       []ExternalAPI                `json:"external_apis"`
	AuthoritativeNS    []AuthoritativeNS            `json:"authoritative_ns"`
	DNSResolverIP      string                       `json:"dns_resolver_ip"`
	CanaryImage        string                       `json:"canary_image"`
	NegativeCanary     CanaryEndpointInput          `json:"negative_canary"`
	EmergencyOwner     string                       `json:"emergency_owner"`
	Authorization      OnboardingAuthorizationInput `json:"authorization"`
	ExecutionIsolation ExecutionIsolation           `json:"execution_isolation"`
	EnabledModules     []string                     `json:"enabled_modules"`
	EnabledToolsP1     []string                     `json:"enabled_tools_p1"`
	EnabledToolsP2     []string                     `json:"enabled_tools_p2"`
	ToolPolicyP1Source string                       `json:"tool_policy_p1_source"`
	ToolPolicyP2Source string                       `json:"tool_policy_p2_source"`
	Tests              []MatrixTest                 `json:"tests"`
	Adapter            OnboardingAdapterInput       `json:"adapter"`
}

type CanaryEndpointInput struct {
	URL  string `json:"url"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type OnboardingStatus struct {
	SchemaVersion string   `json:"schema_version"`
	CaseID        string   `json:"case_id"`
	Authorized    bool     `json:"authorized"`
	Executable    bool     `json:"executable"`
	Blockers      []string `json:"blockers"`
	Generated     []string `json:"generated"`
}

func GenerateOnboardingDrafts(inputPath, outputDir string) (OnboardingStatus, error) {
	var input OnboardingInput
	if err := ReadJSONWithSchema(inputPath, "onboarding-input.schema.json", &input); err != nil {
		return OnboardingStatus{}, err
	}
	if err := validateOnboardingInput(input); err != nil {
		return OnboardingStatus{}, err
	}
	if entries, err := os.ReadDir(outputDir); err == nil && len(entries) != 0 {
		return OnboardingStatus{}, fmt.Errorf("output-dir do onboarding precisa estar vazio")
	} else if err != nil && !os.IsNotExist(err) {
		return OnboardingStatus{}, err
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return OnboardingStatus{}, err
	}
	paths := PolicyPaths{Engagement: "engagement-policy-p1.json", Tool: "tool-policy-p1.json", Matrix: "test-matrix-p1.json"}
	lead := LeadConfig{CaseID: input.CaseID, Domain: input.Domain, BaseURL: input.BaseURL, ApplicationCategory: "unknown", LeadOrigin: "onboarding-schema", Tags: []string{"explicit-input-only"}, OutputRoot: input.OutputRoot, MaxRequests: minInt(input.Target.MaxRequests, 500), RequestsPerSecond: minFloat(input.Target.RatePerSecond, 5), MaxDownloadBytes: int64(input.MaxDownloadBytes), ToolLockPath: input.ToolLockPath, Policies: paths, HIBPAPIKeyEnv: "HIBP_DISABLED", HIBPUserAgent: "EmpresaSecurity-Onboarding/1.0", PublicContactPaths: dedupeSortedCase(input.PublicContactPaths)}
	p1Engagement := EngagementPolicy{EngagementID: input.EngagementID + "_p1", PipelineMode: "pipeline1_public_low_impact", PublicObservation: &input.PublicObservation, Environment: EngagementEnvironment{Targets: []NetworkTarget{input.Target}, ExternalAPIs: input.ExternalAPIs, AuthoritativeNS: input.AuthoritativeNS, DNSResolverIP: input.DNSResolverIP, CanaryImage: input.CanaryImage}, EmergencyStop: EmergencyStopPolicy{Script: "eng-stop.sh", Owner: input.EmergencyOwner}, Canary: CanaryPolicy{NegativeURL: input.NegativeCanary.URL, NegativeIP: input.NegativeCanary.IP, NegativePort: input.NegativeCanary.Port, PositiveURL: input.BaseURL, PositiveIP: input.Target.IP, PositivePort: input.Target.Port}}
	p1Tools, err := selectedToolPolicy(input.ToolPolicyP1Source, input.EnabledToolsP1, 1)
	if err != nil {
		return OnboardingStatus{}, err
	}
	p2Tools, err := selectedToolPolicy(input.ToolPolicyP2Source, input.EnabledToolsP2, 2)
	if err != nil {
		return OnboardingStatus{}, err
	}
	p1Matrix := TestMatrix{EngagementID: input.EngagementID + "_p1", Tests: []MatrixTest{}}
	p2Matrix := TestMatrix{EngagementID: input.EngagementID, Tests: append([]MatrixTest{}, input.Tests...)}
	scope := ScopeConfig{CaseID: input.CaseID, BaseURL: input.BaseURL, OutputRoot: input.OutputRoot, Authorization: Authorization{Confirmed: input.Authorization.Confirmed, SOWPath: input.Authorization.SOWPath, SOWSHA256: input.Authorization.SOWSHA256, EmergencyContact: input.Authorization.EmergencyContact, PaymentStatus: input.Authorization.PaymentStatus}, Isolation: input.ExecutionIsolation, AllowedSchemes: []string{mustURLScheme(input.BaseURL)}, AllowedHosts: []string{input.Target.Host}, AllowedPorts: []int{input.Target.Port}, AllowPrivateIPs: isPrivateAddress(input.Target.IP), MaxRequests: input.Target.MaxRequests, RequestsPerSecond: input.Target.RatePerSecond, MaxConcurrency: input.Target.MaxConcurrency, TimeoutSeconds: 30, RepositoryPath: input.Adapter.RepositoryPath, EnabledModules: dedupeSortedCase(input.EnabledModules), NucleiTemplates: []string{}, ToolLockPath: input.ToolLockPath, Policies: PolicyPaths{Engagement: "engagement-policy-p2.json", Tool: "tool-policy-p2.json", Matrix: "test-matrix-p2.json"}, Automation: AutomationConfig{}}
	// A generic draft is deliberately non-executable. A reviewed stack profile
	// must replace both sentinel commands and hashes before preflight can pass.
	blockedHash := strings.Repeat("0", 64)
	remediation := RemediationConfig{SchemaVersion: RemediationSchemaVersion, CaseID: input.CaseID, RepositoryPath: input.Adapter.RepositoryPath, BaselineRef: input.Adapter.BaselineRef, BaselineBundlePath: input.Adapter.BaselineBundlePath, OutputRoot: input.Adapter.RemediationOutputRoot, SecurityChangesOnly: true, HumanReviewStage: "final", Limits: RemediationLimits{MaxAttemptsPerTicket: 2, MaxTotalAttempts: 20}, ChangePolicy: RemediationChangePolicy{DefaultAllowedPaths: []string{"src/**", "app/**"}, TestPaths: []string{"tests/**"}, ProtectedPaths: []string{".github/**", "docs/**", "public/**"}, MaxFilesChanged: 12, MaxChangedLines: 800, AllowDependencyChange: false, AllowBinaryChanges: false}, CandidateBundleGateID: "security-full", DefaultRetestGateIDs: []string{"retest-finding"}, EnvironmentAllowlist: []string{"REMEDIATION_PATCH_ROOT"}, Agent: RemediationAgent{Command: []string{input.Adapter.RunnerPath, "PROFILE_ADAPTER_REQUIRED"}, ExecutableSHA256: blockedHash, TimeoutSeconds: 300, AutoCommit: true}, Gates: []RemediationGate{{ID: "quality-project", Class: "quality", Command: []string{input.Adapter.RunnerPath, "PROJECT_QUALITY_PROFILE_REQUIRED"}, ExecutableSHA256: blockedHash, TimeoutSeconds: 900, Required: true, ArtifactPaths: []string{"quality-result.json"}}, {ID: "retest-finding", Class: "retest", Command: []string{input.Adapter.RunnerPath, "PROJECT_RETEST_PROFILE_REQUIRED"}, ExecutableSHA256: blockedHash, TimeoutSeconds: 900, Required: true, ArtifactPaths: []string{"retest-result.json"}}, {ID: "security-full", Class: "security", Command: []string{input.Adapter.RunnerPath, "PROJECT_SECURITY_PROFILE_REQUIRED"}, ExecutableSHA256: blockedHash, TimeoutSeconds: 1800, Required: true, ArtifactPaths: []string{"candidate-bundle.json"}}}, FindingOverrides: []FindingRemediationOverride{}}
	objects := []struct {
		name   string
		schema string
		value  any
	}{{"lead.json", "lead.schema.json", lead}, {"engagement-policy-p1.json", "engagement-policy.schema.json", p1Engagement}, {"tool-policy-p1.json", "tool-policy.schema.json", p1Tools}, {"test-matrix-p1.json", "test-matrix.schema.json", p1Matrix}, {"scope.json", "scope.schema.json", scope}, {"tool-policy-p2.json", "tool-policy.schema.json", p2Tools}, {"test-matrix-p2.json", "test-matrix.schema.json", p2Matrix}, {"remediation.json", "remediation-config.schema.json", remediation}}
	if input.Authorization.Confirmed {
		p2 := EngagementPolicy{EngagementID: input.EngagementID, PipelineMode: "pipeline2_active", Authorization: &EngagementAuthorization{Confirmed: true, ConfirmedAt: input.Authorization.ConfirmedAt, ConfirmedBy: input.Authorization.ConfirmedBy, DocumentReference: input.Authorization.DocumentReference}, Environment: p1Engagement.Environment, EmergencyStop: p1Engagement.EmergencyStop, Canary: p1Engagement.Canary}
		objects = append(objects, struct {
			name   string
			schema string
			value  any
		}{"engagement-policy-p2.json", "engagement-policy.schema.json", p2})
	}
	status := OnboardingStatus{SchemaVersion: OnboardingSchemaVersion, CaseID: input.CaseID, Authorized: input.Authorization.Confirmed, Executable: false, Blockers: []string{}, Generated: []string{}}
	for _, object := range objects {
		path := filepath.Join(outputDir, object.name)
		if err := WriteJSON(path, object.value); err != nil {
			return status, err
		}
		if err := validateJSONSchemaFile(object.schema, path); err != nil {
			return status, fmt.Errorf("rascunho %s não respeita schema: %w", object.name, err)
		}
		status.Generated = append(status.Generated, object.name)
	}
	if !input.Authorization.Confirmed {
		status.Blockers = append(status.Blockers, "autorização explícita ausente; engagement-policy-p2.json não foi gerado")
	}
	if !input.ExecutionIsolation.Confirmed {
		status.Blockers = append(status.Blockers, "isolamento ainda não confirmado por evidência")
	}
	status.Blockers = append(status.Blockers, "gates de projeto permanecem marcados como PROJECT_*_PROFILE_REQUIRED até seleção do perfil")
	sort.Strings(status.Generated)
	if err := WriteJSON(filepath.Join(outputDir, "onboarding-status.json"), status); err != nil {
		return status, err
	}
	return status, nil
}

func validateOnboardingInput(input OnboardingInput) error {
	if input.SchemaVersion != OnboardingSchemaVersion || !caseIDPattern.MatchString(input.CaseID) || !strings.HasPrefix(input.EngagementID, "eng_") {
		return fmt.Errorf("identificadores do onboarding inválidos")
	}
	parsed, err := url.Parse(input.BaseURL)
	if err != nil || parsed.Hostname() != input.Target.Host || parsed.Port() != optionalPort(input.Target.Port, parsed.Scheme) {
		return fmt.Errorf("base_url deve corresponder exatamente a target host/port")
	}
	if input.PublicObservation.AcknowledgedAt == "" || input.PublicObservation.AcknowledgedBy == "" || !input.PublicObservation.Acknowledged {
		return fmt.Errorf("observação pública exige reconhecimento explícito fornecido pelo operador")
	}
	if containsFold(input.EnabledToolsP1, "hibp") {
		return fmt.Errorf("HIBP permanece desabilitado no primeiro caso")
	}
	if input.Authorization.Confirmed {
		if input.Authorization.ConfirmedAt == "" || input.Authorization.ConfirmedBy == "" || input.Authorization.DocumentReference == "" || !validSHA256Hex(input.Authorization.SOWSHA256) || input.Authorization.SOWPath == "" || input.Authorization.PaymentStatus != "paid" {
			return fmt.Errorf("autorização confirmada exige origem, SOW com SHA-256 e pagamento explícitos")
		}
	}
	if input.Adapter.RunnerPath == "" || !validSHA256Hex(input.Adapter.RunnerSHA256) || input.Adapter.RepositoryPath == "" || input.Adapter.BaselineRef == "" || input.Adapter.BaselineBundlePath == "" || input.Adapter.RemediationOutputRoot == "" {
		return fmt.Errorf("dados explícitos do adaptador/remediação são obrigatórios")
	}
	return nil
}

func selectedToolPolicy(source string, selected []string, pipeline int) (ToolPolicyFile, error) {
	var sourcePolicy ToolPolicyFile
	if err := ReadJSONWithSchema(source, "tool-policy.schema.json", &sourcePolicy); err != nil {
		return ToolPolicyFile{}, err
	}
	result := ToolPolicyFile{Version: 1, Tools: map[string]ToolPolicyEntry{}}
	for _, name := range dedupeSortedCase(selected) {
		entry, exists := sourcePolicy.Tools[name]
		if !exists || entry.Pipeline != pipeline {
			return result, fmt.Errorf("ferramenta %s não existe no perfil P%d", name, pipeline)
		}
		result.Tools[name] = entry
	}
	return result, nil
}

func mustURLScheme(raw string) string {
	parsed, _ := url.Parse(raw)
	return parsed.Scheme
}

func optionalPort(port int, scheme string) string {
	if (scheme == "https" && port == 443) || (scheme == "http" && port == 80) {
		return ""
	}
	return fmt.Sprintf("%d", port)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func isPrivateAddress(raw string) bool {
	ip := net.ParseIP(raw)
	return ip != nil && (ip.IsPrivate() || ip.IsLoopback())
}
