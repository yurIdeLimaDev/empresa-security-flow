package app

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

type PolicyPaths struct {
	Engagement string `json:"engagement"`
	Tool       string `json:"tool"`
	Matrix     string `json:"matrix"`
}

type PublicObservationPolicy struct {
	Acknowledged   bool   `json:"acknowledged"`
	AcknowledgedAt string `json:"acknowledged_at"`
	AcknowledgedBy string `json:"acknowledged_by"`
}

type EngagementAuthorization struct {
	Confirmed         bool   `json:"confirmed"`
	ConfirmedAt       string `json:"confirmed_at"`
	ConfirmedBy       string `json:"confirmed_by"`
	DocumentReference string `json:"document_reference"`
}

type NetworkTarget struct {
	Host           string  `json:"host"`
	IP             string  `json:"ip"`
	Port           int     `json:"port"`
	RatePerSecond  float64 `json:"rate_per_second"`
	MaxConcurrency int     `json:"max_concurrency"`
	MaxRequests    int     `json:"max_requests"`
}

type ExternalAPI struct {
	Host    string `json:"host"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	Purpose string `json:"purpose"`
}

type AuthoritativeNS struct {
	Host string `json:"host"`
	IP   string `json:"ip"`
}

type EngagementEnvironment struct {
	Targets         []NetworkTarget   `json:"targets"`
	ExternalAPIs    []ExternalAPI     `json:"external_apis"`
	AuthoritativeNS []AuthoritativeNS `json:"authoritative_ns"`
	DNSResolverIP   string            `json:"dns_resolver_ip"`
	CanaryImage     string            `json:"canary_image"`
}

type EmergencyStopPolicy struct {
	Script string `json:"script"`
	Owner  string `json:"owner"`
}

type CanaryPolicy struct {
	NegativeURL  string `json:"negative_url"`
	NegativeIP   string `json:"negative_ip"`
	NegativePort int    `json:"negative_port"`
	PositiveURL  string `json:"positive_url"`
	PositiveIP   string `json:"positive_ip"`
	PositivePort int    `json:"positive_port"`
}

type EngagementPolicy struct {
	EngagementID      string                   `json:"engagement_id"`
	PipelineMode      string                   `json:"pipeline_mode"`
	PublicObservation *PublicObservationPolicy `json:"public_observation,omitempty"`
	Authorization     *EngagementAuthorization `json:"authorization,omitempty"`
	Environment       EngagementEnvironment    `json:"environment"`
	EmergencyStop     EmergencyStopPolicy      `json:"emergency_stop"`
	Canary            CanaryPolicy             `json:"canary"`
}

type ToolPin struct {
	Source string `json:"source"`
	Value  string `json:"value"`
}

type ToolFlags struct {
	ClaimResource        bool `json:"claim_resource"`
	VerifySecrets        bool `json:"verify_secrets"`
	BaaSValidate         bool `json:"baas_validate"`
	DomainSearch         bool `json:"domain_search"`
	AutoRegisterAccounts bool `json:"auto_register_accounts"`
	ApplyChanges         bool `json:"apply_changes"`
	AllowStoredXSS       bool `json:"allow_stored_xss"`
	AllowTimeBasedSQLi   bool `json:"allow_time_based_sqli"`
	AllowStackedQueries  bool `json:"allow_stacked_queries"`
	AllowJWTPlaybook     bool `json:"allow_jwt_playbook"`
	AllowJWTBruteforce   bool `json:"allow_jwt_bruteforce"`
}

type ToolPolicyEntry struct {
	Mode               string    `json:"mode"`
	Pipeline           int       `json:"pipeline"`
	Pin                ToolPin   `json:"pin"`
	MaxDurationMinutes int       `json:"max_duration_minutes"`
	Prohibitions       []string  `json:"prohibitions"`
	Flags              ToolFlags `json:"flags"`
}

type ToolPolicyFile struct {
	Version int                        `json:"version"`
	Tools   map[string]ToolPolicyEntry `json:"tools"`
}

type MatrixTest struct {
	ID             string `json:"id"`
	Resource       string `json:"resource"`
	User           string `json:"user"`
	Action         string `json:"action"`
	ExpectedResult string `json:"expected_result"`
	EvidencePath   string `json:"evidence_path"`
}

type TestMatrix struct {
	EngagementID string       `json:"engagement_id"`
	Tests        []MatrixTest `json:"tests"`
}

type PolicyHashes struct {
	Engagement string `json:"engagement"`
	Tool       string `json:"tool"`
	Matrix     string `json:"matrix"`
}

type ValidatedPolicies struct {
	Engagement EngagementPolicy
	Tools      ToolPolicyFile
	Matrix     *TestMatrix
	Hashes     PolicyHashes
	Paths      PolicyPaths
}

func LoadPolicies(paths PolicyPaths, pipeline int, lock *ToolLockFile) (ValidatedPolicies, error) {
	var result ValidatedPolicies
	result.Paths = paths
	if strings.TrimSpace(paths.Engagement) == "" || strings.TrimSpace(paths.Tool) == "" {
		return result, fmt.Errorf("policies.engagement e policies.tool são obrigatórios")
	}
	engagementPath, err := filepath.Abs(paths.Engagement)
	if err != nil {
		return result, err
	}
	toolPath, err := filepath.Abs(paths.Tool)
	if err != nil {
		return result, err
	}
	result.Paths.Engagement = engagementPath
	result.Paths.Tool = toolPath
	if err := validateJSONSchemaFile("engagement-policy.schema.json", engagementPath); err != nil {
		return result, fmt.Errorf("engagement policy: %w", err)
	}
	if err := validateJSONSchemaFile("tool-policy.schema.json", toolPath); err != nil {
		return result, fmt.Errorf("tool policy: %w", err)
	}
	if err := ReadJSON(engagementPath, &result.Engagement); err != nil {
		return result, err
	}
	if err := ReadJSON(toolPath, &result.Tools); err != nil {
		return result, err
	}
	result.Hashes.Engagement, err = hashPolicyFile(engagementPath)
	if err != nil {
		return result, err
	}
	result.Hashes.Tool, err = hashPolicyFile(toolPath)
	if err != nil {
		return result, err
	}
	result.Hashes.Matrix = "not-applicable"
	if strings.TrimSpace(paths.Matrix) != "" {
		matrixPath, absErr := filepath.Abs(paths.Matrix)
		if absErr != nil {
			return result, absErr
		}
		result.Paths.Matrix = matrixPath
		if err := validateJSONSchemaFile("test-matrix.schema.json", matrixPath); err != nil {
			return result, fmt.Errorf("test matrix: %w", err)
		}
		var matrix TestMatrix
		if err := ReadJSON(matrixPath, &matrix); err != nil {
			return result, err
		}
		result.Matrix = &matrix
		result.Hashes.Matrix, err = hashPolicyFile(matrixPath)
		if err != nil {
			return result, err
		}
	}
	if err := validatePolicySemantics(&result, pipeline); err != nil {
		return result, err
	}
	if lock != nil {
		if err := validateToolPolicyPins(result.Tools, *lock); err != nil {
			return result, err
		}
	}
	return result, nil
}

func validateJSONSchemaFile(schemaName, instancePath string) error {
	schemaPath, err := resolveSchemaPath(schemaName)
	if err != nil {
		return err
	}
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	instanceData, err := os.ReadFile(instancePath)
	if err != nil {
		return err
	}
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	const resource = "https://empresa-security.local/schema.json"
	schemaDocument, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaData))
	if err != nil {
		return fmt.Errorf("JSON do schema inválido: %w", err)
	}
	if err := compiler.AddResource(resource, schemaDocument); err != nil {
		return fmt.Errorf("carregar schema: %w", err)
	}
	schema, err := compiler.Compile(resource)
	if err != nil {
		return fmt.Errorf("compilar schema: %w", err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(instanceData))
	if err != nil {
		return fmt.Errorf("JSON inválido: %w", err)
	}
	if err := schema.Validate(instance); err != nil {
		return fmt.Errorf("schema recusou %s: %w", instancePath, err)
	}
	return nil
}

func ReadJSONWithSchema(path, schemaName string, target any) error {
	if err := validateJSONSchemaFile(schemaName, path); err != nil {
		return err
	}
	return ReadJSON(path, target)
}

func ReadEngagementPolicy(path string) (EngagementPolicy, error) {
	var policy EngagementPolicy
	if err := ReadJSONWithSchema(path, "engagement-policy.schema.json", &policy); err != nil {
		return policy, err
	}
	return policy, nil
}

func resolveSchemaPath(name string) (string, error) {
	candidates := []string{
		filepath.Join("config", "schemas", name),
		filepath.Join("pipeline", "config", "schemas", name),
		filepath.Join("..", "..", "config", "schemas", name),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("schema %s não encontrado", name)
}

func hashPolicyFile(path string) (string, error) {
	hash, err := HashFile(path)
	if err != nil {
		return "", err
	}
	return "sha256:" + hash, nil
}

func validatePolicySemantics(policies *ValidatedPolicies, pipeline int) error {
	engagement := policies.Engagement
	wantedMode := "pipeline1_public_low_impact"
	if pipeline == 2 {
		wantedMode = "pipeline2_active"
	}
	if engagement.PipelineMode != wantedMode {
		return fmt.Errorf("pipeline_mode %q não corresponde ao Pipeline %d", engagement.PipelineMode, pipeline)
	}
	if pipeline == 1 {
		if engagement.PublicObservation == nil || !engagement.PublicObservation.Acknowledged {
			return fmt.Errorf("Pipeline 1 exige public_observation.acknowledged=true")
		}
		if engagement.Authorization != nil {
			return fmt.Errorf("Pipeline 1 não usa authorization; use public_observation")
		}
		if len(engagement.Environment.AuthoritativeNS) != 0 {
			return fmt.Errorf("authoritative_ns só é permitido no Pipeline 2 ativo")
		}
	} else if engagement.Authorization == nil || !engagement.Authorization.Confirmed {
		return fmt.Errorf("Pipeline 2 exige authorization.confirmed=true")
	}
	if net.ParseIP(engagement.Environment.DNSResolverIP).To4() == nil {
		return fmt.Errorf("dns_resolver_ip precisa ser IPv4")
	}
	allow := map[string]struct{}{}
	for _, target := range engagement.Environment.Targets {
		allow[destinationKey(target.IP, target.Port)] = struct{}{}
	}
	for _, external := range engagement.Environment.ExternalAPIs {
		allow[destinationKey(external.IP, external.Port)] = struct{}{}
	}
	if _, exists := allow[destinationKey(engagement.Canary.NegativeIP, engagement.Canary.NegativePort)]; exists {
		return fmt.Errorf("canary negativo não pode apontar para destino permitido")
	}
	if _, exists := allow[destinationKey(engagement.Canary.PositiveIP, engagement.Canary.PositivePort)]; !exists {
		return fmt.Errorf("canary positivo precisa apontar para target ou API externa permitida")
	}
	if err := validateCanaryURL(engagement.Canary.NegativeURL, engagement.Canary.NegativeIP, engagement.Canary.NegativePort); err != nil {
		return fmt.Errorf("canary negativo: %w", err)
	}
	if err := validateCanaryURL(engagement.Canary.PositiveURL, engagement.Canary.PositiveIP, engagement.Canary.PositivePort); err != nil {
		return fmt.Errorf("canary positivo: %w", err)
	}
	if policies.Matrix != nil && policies.Matrix.EngagementID != engagement.EngagementID {
		return fmt.Errorf("test-matrix pertence a outro engagement_id")
	}
	for name, tool := range policies.Tools.Tools {
		if tool.Pipeline != pipeline && tool.Mode == "automatic" {
			return fmt.Errorf("ferramenta %s automática está declarada para o Pipeline %d", name, tool.Pipeline)
		}
		if tool.Flags.ClaimResource || tool.Flags.AutoRegisterAccounts || tool.Flags.ApplyChanges {
			return fmt.Errorf("ferramenta %s tentou habilitar proibição absoluta", name)
		}
		if name == "hibp" && tool.Flags.DomainSearch {
			return fmt.Errorf("HIBP domain_search é proibido no runner")
		}
		if pipeline == 1 && name == "trufflehog" && tool.Mode == "automatic" && tool.Flags.VerifySecrets {
			return fmt.Errorf("TruffleHog no Pipeline 1 não pode verificar segredos")
		}
		if tool.Mode == "automatic" && requiresTestMatrix(name) && (policies.Matrix == nil || len(policies.Matrix.Tests) == 0) {
			return fmt.Errorf("ferramenta ativa de dados %s exige test-matrix não vazia", name)
		}
	}
	if hibp, ok := policies.Tools.Tools["hibp"]; ok && hibp.Mode == "automatic" {
		found := false
		for _, api := range engagement.Environment.ExternalAPIs {
			if strings.EqualFold(api.Host, "haveibeenpwned.com") && api.Port == 443 && strings.EqualFold(api.Purpose, "hibp") {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("HIBP automático exige haveibeenpwned.com:443 em external_apis com purpose=hibp")
		}
	}
	return nil
}

func validateCanaryURL(raw, ip string, port int) error {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return fmt.Errorf("URL inválida")
	}
	wantedPort := 80
	if u.Scheme == "https" {
		wantedPort = 443
	}
	if u.Port() != "" {
		fmt.Sscanf(u.Port(), "%d", &wantedPort)
	}
	if wantedPort != port {
		return fmt.Errorf("porta da URL diverge da porta pinada")
	}
	if net.ParseIP(ip).To4() == nil {
		return fmt.Errorf("IP precisa ser IPv4")
	}
	return nil
}

func destinationKey(ip string, port int) string { return strings.ToLower(ip) + ":" + fmt.Sprint(port) }

func requiresTestMatrix(name string) bool {
	switch name {
	case "supabase_rls_checker", "supashield", "firepwn":
		return true
	default:
		return false
	}
}

func validateToolPolicyPins(policy ToolPolicyFile, lock ToolLockFile) error {
	entries := make(map[string]ToolLock, len(lock.Tools))
	for _, entry := range lock.Tools {
		entries[normalizeToolName(entry.Name)] = entry
	}
	names := make([]string, 0, len(policy.Tools))
	for name := range policy.Tools {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		entry, exists := entries[normalizeToolName(name)]
		if !exists {
			return fmt.Errorf("%s não existe no tools.lock.json", name)
		}
		pin := policy.Tools[name].Pin
		var locked string
		switch pin.Source {
		case "docker_digest":
			if entry.Container == nil {
				return fmt.Errorf("%s exige container pinado no tools.lock.json", name)
			}
			locked = entry.Container.Digest
		case "git_commit":
			locked = entry.Release.Commit
		case "binary_sha256":
			locked = entry.LocalBinarySHA256
		case "api_contract_sha256":
			locked = entry.APIContractSHA256
		case "pypi_version", "npm_version":
			locked = strings.TrimPrefix(entry.Release.Tag, "v")
		default:
			return fmt.Errorf("%s usa fonte de pin desconhecida", name)
		}
		wanted := pin.Value
		if pin.Source == "pypi_version" || pin.Source == "npm_version" {
			wanted = strings.TrimPrefix(wanted, "v")
		}
		if subtle.ConstantTimeCompare([]byte(strings.ToLower(strings.TrimSpace(wanted))), []byte(strings.ToLower(strings.TrimSpace(locked)))) != 1 {
			return fmt.Errorf("pin de %s diverge do tools.lock.json", name)
		}
	}
	return nil
}

func normalizeToolName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "amass_active":
		return "amass"
	case "supabase_rls_checker":
		return "supabase-rls-checker"
	case "keyleak":
		return "keyleak-detector"
	default:
		return strings.ToLower(strings.TrimSpace(name))
	}
}

func (p ValidatedPolicies) Tool(name string, pipeline int) (ToolPolicyEntry, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	entry, ok := p.Tools.Tools[key]
	if !ok {
		return ToolPolicyEntry{}, fmt.Errorf("ferramenta %s ausente no tool-policy", name)
	}
	if entry.Pipeline != pipeline {
		return ToolPolicyEntry{}, fmt.Errorf("ferramenta %s declarada para Pipeline %d", name, entry.Pipeline)
	}
	return entry, nil
}

func ValidateConfigAgainstEngagement(baseRaw string, maxRequests int, rps float64, maxConcurrency int, policy EngagementPolicy) (NetworkTarget, error) {
	u, err := url.Parse(baseRaw)
	if err != nil || u.Hostname() == "" {
		return NetworkTarget{}, fmt.Errorf("base_url inválida")
	}
	port := 80
	if u.Scheme == "https" {
		port = 443
	}
	if u.Port() != "" {
		if _, scanErr := fmt.Sscanf(u.Port(), "%d", &port); scanErr != nil {
			return NetworkTarget{}, fmt.Errorf("porta da base_url inválida")
		}
	}
	for _, target := range policy.Environment.Targets {
		if !strings.EqualFold(target.Host, u.Hostname()) || target.Port != port {
			continue
		}
		if maxRequests > target.MaxRequests {
			return NetworkTarget{}, fmt.Errorf("max_requests do config excede engagement-policy")
		}
		if rps > target.RatePerSecond {
			return NetworkTarget{}, fmt.Errorf("requests_per_second do config excede engagement-policy")
		}
		if maxConcurrency > 0 && maxConcurrency > target.MaxConcurrency {
			return NetworkTarget{}, fmt.Errorf("max_concurrency do config excede engagement-policy")
		}
		return target, nil
	}
	return NetworkTarget{}, fmt.Errorf("base_url não existe em environment.targets com host, IP e porta pinados")
}
