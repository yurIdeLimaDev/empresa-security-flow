package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var caseIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{2,63}$`)

type LeadConfig struct {
	CaseID              string      `json:"case_id"`
	Domain              string      `json:"domain"`
	BaseURL             string      `json:"base_url"`
	CORSOrigin          string      `json:"cors_origin,omitempty"`
	ApplicationCategory string      `json:"application_category,omitempty"`
	LeadOrigin          string      `json:"lead_origin,omitempty"`
	Tags                []string    `json:"tags,omitempty"`
	OutputRoot          string      `json:"output_root"`
	MaxRequests         int         `json:"max_requests"`
	RequestsPerSecond   float64     `json:"requests_per_second"`
	MaxDownloadBytes    int64       `json:"max_download_bytes"`
	ToolLockPath        string      `json:"tool_lock_path,omitempty"`
	Policies            PolicyPaths `json:"policies"`
	HIBPAPIKeyEnv       string      `json:"hibp_api_key_env"`
	HIBPUserAgent       string      `json:"hibp_user_agent"`
	PublicContactPaths  []string    `json:"public_contact_paths"`
}

type Authorization struct {
	Confirmed        bool   `json:"confirmed"`
	SOWPath          string `json:"sow_path"`
	SOWSHA256        string `json:"sow_sha256"`
	EmergencyContact string `json:"emergency_contact"`
	PaymentStatus    string `json:"payment_status"`
}

type ScopeConfig struct {
	CaseID             string             `json:"case_id"`
	BaseURL            string             `json:"base_url"`
	OutputRoot         string             `json:"output_root"`
	Authorization      Authorization      `json:"authorization"`
	AllowedSchemes     []string           `json:"allowed_schemes"`
	AllowedHosts       []string           `json:"allowed_hosts"`
	AllowedPorts       []int              `json:"allowed_ports"`
	AllowPrivateIPs    bool               `json:"allow_private_ips"`
	MaxRequests        int                `json:"max_requests"`
	RequestsPerSecond  float64            `json:"requests_per_second"`
	MaxConcurrency     int                `json:"max_concurrency"`
	TimeoutSeconds     int                `json:"timeout_seconds"`
	StackProfilePath   string             `json:"stack_profile_path"`
	RepositoryPath     string             `json:"repository_path,omitempty"`
	SemgrepConfigPath  string             `json:"semgrep_config_path,omitempty"`
	WordlistPath       string             `json:"wordlist_path,omitempty"`
	OpenAPIURL         string             `json:"openapi_url,omitempty"`
	ScenarioPath       string             `json:"scenario_path,omitempty"`
	HadrianRolesPath   string             `json:"hadrian_roles_path,omitempty"`
	HadrianAuthPath    string             `json:"hadrian_auth_path,omitempty"`
	HadrianMutations   bool               `json:"hadrian_allow_mutations,omitempty"`
	InteractshServer   string             `json:"interactsh_server,omitempty"`
	InteractshTokenEnv string             `json:"interactsh_token_env,omitempty"`
	Isolation          ExecutionIsolation `json:"execution_isolation"`
	EnabledModules     []string           `json:"enabled_modules,omitempty"`
	NucleiTemplates    []string           `json:"nuclei_templates,omitempty"`
	ToolLockPath       string             `json:"tool_lock_path,omitempty"`
	Policies           PolicyPaths        `json:"policies"`
	Automation         AutomationConfig   `json:"automation"`
}

type AutomationConfig struct {
	AmassActive *AmassActiveConfig `json:"amass_active,omitempty"`
	Supashield  *SupashieldConfig  `json:"supashield,omitempty"`
	ZAP         *ZAPConfig         `json:"zap,omitempty"`
	JWT         *JWTConfig         `json:"jwt_tool,omitempty"`
	Dalfox      *DalfoxConfig      `json:"dalfox,omitempty"`
	SQLMap      *SQLMapConfig      `json:"sqlmap,omitempty"`
}

type AmassActiveConfig struct {
	DatasourcesPath string `json:"datasources_path"`
}

type SupashieldConfig struct {
	Environment     string `json:"environment"`
	BackupReference string `json:"backup_reference"`
	DatabaseURLEnv  string `json:"database_url_env"`
}

type ZAPConfig struct {
	TemplatePath    string   `json:"template_path"`
	LoginScriptPath string   `json:"login_script_path"`
	UserEmailEnv    string   `json:"user_email_env"`
	UserPasswordEnv string   `json:"user_password_env"`
	IncludePaths    []string `json:"include_paths"`
	ExcludePaths    []string `json:"exclude_paths"`
	LoggedInRegex   string   `json:"logged_in_regex"`
	LoggedOutRegex  string   `json:"logged_out_regex"`
	SpiderMinutes   int      `json:"spider_minutes"`
	RuleMinutes     int      `json:"rule_minutes"`
}

type JWTConfig struct {
	TokenEnv          string `json:"token_env"`
	TargetURL         string `json:"target_url"`
	ExpectedIssuer    string `json:"expected_issuer"`
	ExpectedAudience  string `json:"expected_audience"`
	ExpectedAlgorithm string `json:"expected_algorithm"`
	PublicKeyPath     string `json:"public_key_path"`
}

type DalfoxConfig struct {
	TargetsPath  string `json:"targets_path"`
	PayloadsPath string `json:"payloads_path"`
}

type SQLMapConfig struct {
	RequestPath string `json:"request_path"`
	Parameter   string `json:"parameter"`
	SafeURL     string `json:"safe_url"`
	TargetKind  string `json:"target_kind"`
}

type ExecutionIsolation struct {
	Confirmed      bool   `json:"confirmed"`
	Method         string `json:"method,omitempty"`
	EvidencePath   string `json:"evidence_path,omitempty"`
	EvidenceSHA256 string `json:"evidence_sha256,omitempty"`
}

type Observation struct {
	Category    string `json:"category"`
	Value       string `json:"value"`
	SourceType  string `json:"source_type"`
	SourceURL   string `json:"source_url"`
	ObservedAt  string `json:"observed_at"`
	Confidence  string `json:"confidence"`
	Limitations string `json:"limitations,omitempty"`
}

type StackProfile struct {
	Subject         string        `json:"subject"`
	Status          string        `json:"status"`
	CollectedAt     string        `json:"collected_at"`
	Category        string        `json:"application_category,omitempty"`
	Observations    []Observation `json:"observations"`
	ApprovedModules []string      `json:"approved_modules,omitempty"`
}

type PlannedAction struct {
	ID                string      `json:"id"`
	Module            string      `json:"module"`
	Tool              string      `json:"tool"`
	Command           []string    `json:"command,omitempty"`
	SecretArgsFromEnv []SecretArg `json:"secret_args_from_env,omitempty"`
	State             string      `json:"state"`
	Reason            string      `json:"reason,omitempty"`
	ExitCode          int         `json:"exit_code,omitempty"`
	PolicyTool        string      `json:"policy_tool,omitempty"`
	InputPaths        []string    `json:"input_paths,omitempty"`
	ArtifactPaths     []string    `json:"artifact_paths,omitempty"`
	AllowedExitCodes  []int       `json:"allowed_exit_codes,omitempty"`
	EphemeralPaths    []string    `json:"ephemeral_paths,omitempty"`
}

type SecretArg struct {
	Flag     string `json:"flag"`
	Env      string `json:"env"`
	Position string `json:"position,omitempty"`
}

type Manifest struct {
	CaseID       string          `json:"case_id"`
	Pipeline     int             `json:"pipeline"`
	Mode         string          `json:"mode"`
	StartedAt    string          `json:"started_at"`
	FinishedAt   string          `json:"finished_at,omitempty"`
	ConfigSHA256 string          `json:"config_sha256"`
	Actions      []PlannedAction `json:"actions"`
	Warnings     []string        `json:"warnings,omitempty"`
}

type HTTPObservation struct {
	URL         string              `json:"url"`
	Method      string              `json:"method"`
	Status      int                 `json:"status"`
	Headers     map[string][]string `json:"headers,omitempty"`
	BodyBytes   int                 `json:"body_bytes"`
	BodySHA256  string              `json:"body_sha256,omitempty"`
	ContentType string              `json:"content_type,omitempty"`
	Error       string              `json:"error,omitempty"`
}

type ScenarioFile struct {
	Scenarios []Scenario `json:"scenarios"`
}

type Scenario struct {
	Name             string              `json:"name"`
	Domain           string              `json:"domain"`
	Module           string              `json:"module"`
	Method           string              `json:"method"`
	Path             string              `json:"path"`
	HeadersFromEnv   map[string]string   `json:"headers_from_env,omitempty"`
	Body             json.RawMessage     `json:"body,omitempty"`
	Concurrency      int                 `json:"concurrency,omitempty"`
	ExpectedStatuses []int               `json:"expected_statuses,omitempty"`
	StateCheck       *ScenarioStateCheck `json:"state_check"`
}

type ScenarioStateCheck struct {
	Method         string            `json:"method"`
	Path           string            `json:"path"`
	HeadersFromEnv map[string]string `json:"headers_from_env,omitempty"`
	JSONPointer    string            `json:"json_pointer"`
	ExpectedDelta  float64           `json:"expected_delta"`
}

type ScenarioResult struct {
	Name          string  `json:"name"`
	Module        string  `json:"module"`
	Status        int     `json:"status"`
	BodyBytes     int     `json:"body_bytes"`
	BodySHA256    string  `json:"body_sha256,omitempty"`
	DurationMS    int64   `json:"duration_ms"`
	Passed        bool    `json:"passed"`
	StateBefore   float64 `json:"state_before"`
	StateAfter    float64 `json:"state_after"`
	ExpectedDelta float64 `json:"expected_delta"`
	ActualDelta   float64 `json:"actual_delta"`
	Error         string  `json:"error,omitempty"`
}

func ReadJSON(path string, target any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("abrir %s: %w", path, err)
	}
	defer f.Close()
	decoder := json.NewDecoder(io.LimitReader(f, 4<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("JSON inválido em %s: %w", path, err)
	}
	return nil
}

func WriteJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func HashJSON(value any) string {
	data, _ := json.Marshal(value)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func HashBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func caseDir(root, caseID, pipeline string) (string, error) {
	if !caseIDPattern.MatchString(caseID) {
		return "", errors.New("case_id deve ter 3-64 caracteres seguros")
	}
	if root == "" {
		root = "artifacts"
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(absRoot, caseID, pipeline)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func validateBaseURL(raw, domain string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Hostname() == "" {
		return nil, errors.New("base_url deve ser uma URL absoluta")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("base_url aceita somente http/https")
	}
	if u.User != nil || u.Fragment != "" {
		return nil, errors.New("base_url não pode conter credencial ou fragmento")
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	domain = strings.ToLower(strings.TrimSuffix(domain, "."))
	if host != domain && !strings.HasSuffix(host, "."+domain) {
		return nil, errors.New("base_url não pertence ao domínio informado")
	}
	return u, nil
}

func normalizeLimits(maxRequests *int, rps *float64) {
	if *maxRequests <= 0 {
		*maxRequests = 50
	}
	if *maxRequests > 500 {
		*maxRequests = 500
	}
	if *rps <= 0 {
		*rps = 1
	}
	if *rps > 5 {
		*rps = 5
	}
}

func dedupeSorted(items []string) []string {
	seen := make(map[string]struct{})
	for _, item := range items {
		item = strings.TrimSpace(strings.ToLower(item))
		if item != "" {
			seen[item] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for item := range seen {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func exactIntegerRate(value float64) (int, bool) {
	rounded := math.Round(value)
	if value <= 0 || math.Abs(value-rounded) > 1e-9 || rounded > float64(^uint(0)>>1) {
		return 0, false
	}
	return int(rounded), true
}
