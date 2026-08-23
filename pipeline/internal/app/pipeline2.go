package app

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func RunPipeline2(ctx context.Context, cfg ScopeConfig, execute bool) error {
	base, err := validateScopeConfig(cfg, execute)
	if err != nil {
		return err
	}
	lock, lockPath, err := ResolveToolLock(cfg.ToolLockPath)
	if err != nil {
		return fmt.Errorf("cadeia de ferramentas: %w", err)
	}
	policies, err := LoadPolicies(cfg.Policies, 2, &lock)
	if err != nil {
		return err
	}
	if _, err := ValidateConfigAgainstEngagement(cfg.BaseURL, cfg.MaxRequests, cfg.RequestsPerSecond, cfg.MaxConcurrency, policies.Engagement); err != nil {
		return err
	}
	if cfg.StackProfilePath != "" {
		cfg.StackProfilePath, _ = filepath.Abs(cfg.StackProfilePath)
	}
	if cfg.RepositoryPath != "" {
		cfg.RepositoryPath, _ = filepath.Abs(cfg.RepositoryPath)
	}
	if cfg.SemgrepConfigPath != "" {
		cfg.SemgrepConfigPath, _ = filepath.Abs(cfg.SemgrepConfigPath)
	}
	if cfg.WordlistPath != "" {
		cfg.WordlistPath, _ = filepath.Abs(cfg.WordlistPath)
	}
	if cfg.ScenarioPath != "" {
		cfg.ScenarioPath, _ = filepath.Abs(cfg.ScenarioPath)
	}
	if cfg.HadrianRolesPath != "" {
		cfg.HadrianRolesPath, _ = filepath.Abs(cfg.HadrianRolesPath)
	}
	if cfg.HadrianAuthPath != "" {
		cfg.HadrianAuthPath, _ = filepath.Abs(cfg.HadrianAuthPath)
	}
	if cfg.Isolation.EvidencePath != "" {
		cfg.Isolation.EvidencePath, _ = filepath.Abs(cfg.Isolation.EvidencePath)
	}
	for index, template := range cfg.NucleiTemplates {
		cfg.NucleiTemplates[index], _ = filepath.Abs(template)
	}
	absolutizeAutomation(&cfg)
	if execute {
		if err := validateExecutionFiles(cfg); err != nil {
			return err
		}
	}
	dir, err := caseDir(cfg.OutputRoot, cfg.CaseID, "pipeline-2")
	if err != nil {
		return err
	}
	network, err := GenerateIsolationScripts(dir, policies.Engagement)
	if err != nil {
		return err
	}

	profileStatus := "unconfirmed_plan"
	if execute {
		profileStatus = "confirmed_by_client"
	}
	profile := StackProfile{Subject: base.Hostname(), Status: profileStatus, CollectedAt: now(), Category: "unknown"}
	if cfg.StackProfilePath != "" {
		if _, statErr := os.Stat(cfg.StackProfilePath); statErr == nil {
			if err := ReadJSON(cfg.StackProfilePath, &profile); err != nil {
				return fmt.Errorf("stack profile: %w", err)
			}
		} else if execute {
			return fmt.Errorf("stack profile: %w", statErr)
		}
	}
	if !profileSubjectMatchesHost(profile.Subject, base.Hostname()) {
		return fmt.Errorf("stack profile pertence a outro domínio: %s", profile.Subject)
	}
	profile.Status = profileStatus
	modules := routeModules(profile, cfg.EnabledModules)
	profile.ApprovedModules = modules
	manifest := Manifest{CaseID: cfg.CaseID, Pipeline: 2, Mode: mode(execute), StartedAt: now(), ConfigSHA256: HashJSON(cfg)}
	manifest.Actions = pipeline2Plan(cfg, modules, dir, policies)
	if err := WriteJSON(filepath.Join(dir, "routed-stack-profile.json"), profile); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		return err
	}
	if !execute {
		manifest.FinishedAt = now()
		if err := WriteJSON(filepath.Join(dir, "manifest.json"), manifest); err != nil {
			return err
		}
		if err := writePipeline2Report(dir, cfg, profile, manifest, nil); err != nil {
			return err
		}
		return writeChecksums(dir)
	}
	if err := SetupAndTestIsolation(ctx, network, dir); err != nil {
		return err
	}
	defer func() { _ = TeardownIsolation(context.Background(), network) }()
	auditProxy, err := StartEngagementAuditProxy(ctx, &network, policies.Engagement, cfg.MaxRequests, cfg.RequestsPerSecond)
	if err != nil {
		return fmt.Errorf("proxy de auditoria: %w", err)
	}
	defer func() { _ = auditProxy.Close(context.Background()) }()
	_ = WriteJSON(filepath.Join(dir, "isolation", "network-state.json"), network)
	runOptions := GovernedRunOptions{Pipeline: 2, Policies: policies, Lock: lock, SBOMPath: filepath.Join(filepath.Dir(lockPath), "sbom.cdx.json"), Network: network, CaseDir: dir}
	for index := range manifest.Actions {
		if manifest.Actions[index].ID != "10-zap-authenticated" || manifest.Actions[index].State != "planned" {
			continue
		}
		planPath, loginPath, renderErr := RenderZAPPlan(cfg, dir)
		if renderErr != nil {
			return fmt.Errorf("render ZAP: %w", renderErr)
		}
		manifest.Actions[index].InputPaths = []string{planPath, loginPath}
		manifest.Actions[index].EphemeralPaths = []string{planPath, loginPath}
		defer func() { _ = os.Remove(planPath); _ = os.Remove(loginPath) }()
	}

	for index := range manifest.Actions {
		action := &manifest.Actions[index]
		if action.ID != "03-openapi-fetch" {
			continue
		}
		if err := downloadOpenAPISchema(ctx, cfg, policies.Engagement, filepath.Join(dir, "openapi-schema.yaml")); err != nil {
			action.State = "failed"
			action.Reason = err.Error()
			manifest.Warnings = append(manifest.Warnings, "OpenAPI: "+err.Error())
			for candidate := range manifest.Actions {
				if containsFold([]string{"03-schemathesis", "04-hadrian-dry-run", "04-hadrian-authorization"}, manifest.Actions[candidate].ID) {
					manifest.Actions[candidate].State = "skipped"
					manifest.Actions[candidate].Reason = "schema OpenAPI não pôde ser obtido dentro do escopo"
				}
			}
		} else {
			action.State = "completed"
		}
	}

	for i := range manifest.Actions {
		action := &manifest.Actions[i]
		if action.Tool == "native" || action.State == "manual" || action.State == "skipped" || action.State == "failed" {
			continue
		}
		err := RunGovernedTool(ctx, runOptions, action)
		if err != nil {
			if IsGateError(err) {
				_ = WriteJSON(filepath.Join(dir, "manifest.json"), manifest)
				return fmt.Errorf("gate %s: %w", action.ID, err)
			}
			manifest.Warnings = append(manifest.Warnings, err.Error())
		}
	}

	var scenarioResults []ScenarioResult
	if cfg.ScenarioPath != "" && actionHasState(manifest.Actions, "05-credit-billing-scenarios", "planned") {
		results, err := runScenarios(ctx, cfg, base, modules, policies.Engagement)
		if err != nil {
			manifest.Warnings = append(manifest.Warnings, "cenários: "+err.Error())
		} else {
			scenarioResults = results
		}
		for index := range manifest.Actions {
			if manifest.Actions[index].ID != "05-credit-billing-scenarios" {
				continue
			}
			if err != nil {
				manifest.Actions[index].State = "failed"
				manifest.Actions[index].Reason = err.Error()
			} else {
				manifest.Actions[index].State = "completed"
			}
		}
		if err := WriteJSON(filepath.Join(dir, "scenario-results.json"), scenarioResults); err != nil {
			return err
		}
	}
	manifest.FinishedAt = now()
	scopeID := "authorized:" + HashJSON(struct {
		Schemes []string
		Hosts   []string
		Ports   []int
		SOW     string
	}{cfg.AllowedSchemes, cfg.AllowedHosts, cfg.AllowedPorts, cfg.Authorization.SOWSHA256})[:24]
	_, normalizeErr := NormalizeKnownOutputs(dir, cfg.CaseID+"-pipeline-2", scopeID, []OutputAdapter{
		{Tool: "schemathesis", File: "schemathesis-junit.xml"}, {Tool: "nuclei", File: "nuclei.jsonl"}, {Tool: "gitleaks", File: "gitleaks-repository.json"},
		{Tool: "osv-scanner", File: "osv-scanner.json"}, {Tool: "semgrep", File: "semgrep.json"}, {Tool: "trivy", File: "trivy.json"},
		{Tool: "zap", File: "zap.json"}, {Tool: "sqlmap", File: "sqlmap-report.json"}, {Tool: "dalfox", File: "dalfox.jsonl"}, {Tool: "jwt_tool", File: "jwt-tool.log"},
		{Tool: "hadrian", File: "hadrian.json"},
	})
	if normalizeErr != nil {
		manifest.Warnings = append(manifest.Warnings, normalizeErr.Error())
	}
	if err := WriteJSON(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		return err
	}
	if err := writePipeline2Report(dir, cfg, profile, manifest, scenarioResults); err != nil {
		return err
	}
	return writeChecksums(dir)
}

func validateScopeConfig(cfg ScopeConfig, execute bool) (*url.URL, error) {
	if !caseIDPattern.MatchString(cfg.CaseID) {
		return nil, fmt.Errorf("case_id inválido")
	}
	base, err := url.Parse(cfg.BaseURL)
	if err != nil || base.Scheme == "" || base.Hostname() == "" {
		return nil, fmt.Errorf("base_url inválida")
	}
	if base.User != nil || base.Fragment != "" {
		return nil, fmt.Errorf("base_url não pode conter credencial ou fragmento")
	}
	if execute && !cfg.Authorization.Confirmed {
		return nil, fmt.Errorf("authorization.confirmed precisa ser true")
	}
	if execute && cfg.Authorization.PaymentStatus != "paid" {
		return nil, fmt.Errorf("payment_status precisa ser paid")
	}
	if execute && strings.TrimSpace(cfg.Authorization.EmergencyContact) == "" {
		return nil, fmt.Errorf("emergency_contact é obrigatório")
	}
	if execute && cfg.Isolation.Method != "container-egress-policy" {
		return nil, fmt.Errorf("execution_isolation.method precisa ser container-egress-policy; proxy não é barreira")
	}
	if execute {
		if cfg.Authorization.SOWPath == "" || cfg.Authorization.SOWSHA256 == "" {
			return nil, fmt.Errorf("SOW e SHA-256 são obrigatórios")
		}
		hash, err := HashFile(cfg.Authorization.SOWPath)
		if err != nil {
			return nil, fmt.Errorf("SOW: %w", err)
		}
		if subtle.ConstantTimeCompare([]byte(strings.ToLower(hash)), []byte(strings.ToLower(cfg.Authorization.SOWSHA256))) != 1 {
			return nil, fmt.Errorf("hash do SOW não confere")
		}
	}
	if !containsFold(cfg.AllowedSchemes, base.Scheme) {
		return nil, fmt.Errorf("base_url usa scheme fora do escopo")
	}
	if len(cfg.AllowedSchemes) == 0 || len(cfg.AllowedHosts) == 0 || len(cfg.AllowedPorts) == 0 {
		return nil, fmt.Errorf("allowlists de scheme, host e porta são obrigatórias")
	}
	for _, scheme := range cfg.AllowedSchemes {
		if !strings.EqualFold(scheme, "http") && !strings.EqualFold(scheme, "https") {
			return nil, fmt.Errorf("allowed_schemes aceita somente http/https")
		}
	}
	if !containsFold(cfg.AllowedHosts, base.Hostname()) {
		return nil, fmt.Errorf("base_url usa host fora do escopo")
	}
	port := 80
	if base.Scheme == "https" {
		port = 443
	}
	if base.Port() != "" {
		port, _ = strconv.Atoi(base.Port())
	}
	if !containsInt(cfg.AllowedPorts, port) {
		return nil, fmt.Errorf("base_url usa porta fora do escopo")
	}
	if cfg.OpenAPIURL != "" {
		openapi, err := url.Parse(cfg.OpenAPIURL)
		if err != nil || openapi.Hostname() == "" || openapi.User != nil || openapi.Fragment != "" || !containsFold(cfg.AllowedHosts, openapi.Hostname()) || !containsFold(cfg.AllowedSchemes, openapi.Scheme) {
			return nil, fmt.Errorf("openapi_url está fora do escopo")
		}
		openapiPort := 80
		if openapi.Scheme == "https" {
			openapiPort = 443
		}
		if openapi.Port() != "" {
			openapiPort, err = strconv.Atoi(openapi.Port())
			if err != nil {
				return nil, fmt.Errorf("openapi_url contém porta inválida")
			}
		}
		if !containsInt(cfg.AllowedPorts, openapiPort) {
			return nil, fmt.Errorf("openapi_url usa porta fora do escopo")
		}
	}
	if cfg.HadrianMutations && (cfg.OpenAPIURL == "" || cfg.HadrianRolesPath == "" || cfg.HadrianAuthPath == "") {
		return nil, fmt.Errorf("hadrian_allow_mutations exige openapi_url, hadrian_roles_path e hadrian_auth_path")
	}
	if cfg.InteractshServer != "" {
		server, parseErr := url.Parse(cfg.InteractshServer)
		if parseErr != nil || server.Scheme != "https" || server.Hostname() == "" || server.User != nil || server.RawQuery != "" || server.Fragment != "" {
			return nil, fmt.Errorf("interactsh_server deve ser uma URL HTTPS própria e sem credencial")
		}
		if strings.TrimSpace(cfg.InteractshTokenEnv) == "" {
			return nil, fmt.Errorf("interactsh_token_env é obrigatório para servidor próprio")
		}
	} else if cfg.InteractshTokenEnv != "" {
		return nil, fmt.Errorf("interactsh_token_env não pode existir sem interactsh_server")
	}
	if cfg.MaxRequests <= 0 || cfg.MaxRequests > 10000 {
		return nil, fmt.Errorf("max_requests deve estar entre 1 e 10000")
	}
	if cfg.RequestsPerSecond <= 0 || cfg.RequestsPerSecond > 50 {
		return nil, fmt.Errorf("requests_per_second deve estar entre 0 e 50")
	}
	if cfg.MaxConcurrency <= 0 || cfg.MaxConcurrency > 20 {
		return nil, fmt.Errorf("max_concurrency deve estar entre 1 e 20")
	}
	if cfg.TimeoutSeconds <= 0 || cfg.TimeoutSeconds > 120 {
		return nil, fmt.Errorf("timeout_seconds deve estar entre 1 e 120")
	}
	allowedModules := []string{"active-recon", "api-mapping", "authentication", "authorization", "business-logic", "dependencies", "reporting", "baas", "storage", "custom-backend", "race-conditions"}
	for _, module := range cfg.EnabledModules {
		if !containsFold(allowedModules, module) {
			return nil, fmt.Errorf("módulo desconhecido: %s", module)
		}
	}
	return base, nil
}

func validateExecutionFiles(cfg ScopeConfig) error {
	type requiredPath struct {
		name     string
		path     string
		wantDir  bool
		optional bool
	}
	paths := []requiredPath{
		{name: "repository_path", path: cfg.RepositoryPath, wantDir: true, optional: true},
		{name: "semgrep_config_path", path: cfg.SemgrepConfigPath, optional: true},
		{name: "wordlist_path", path: cfg.WordlistPath, optional: true},
		{name: "scenario_path", path: cfg.ScenarioPath, optional: true},
		{name: "hadrian_roles_path", path: cfg.HadrianRolesPath, optional: true},
		{name: "hadrian_auth_path", path: cfg.HadrianAuthPath, optional: true},
		{name: "execution_isolation.evidence_path", path: cfg.Isolation.EvidencePath, optional: true},
	}
	if cfg.Automation.AmassActive != nil {
		paths = append(paths, requiredPath{name: "automation.amass_active.datasources_path", path: cfg.Automation.AmassActive.DatasourcesPath})
	}
	if cfg.Automation.ZAP != nil {
		paths = append(paths, requiredPath{name: "automation.zap.template_path", path: cfg.Automation.ZAP.TemplatePath}, requiredPath{name: "automation.zap.login_script_path", path: cfg.Automation.ZAP.LoginScriptPath})
	}
	if cfg.Automation.JWT != nil && cfg.Automation.JWT.PublicKeyPath != "" {
		paths = append(paths, requiredPath{name: "automation.jwt_tool.public_key_path", path: cfg.Automation.JWT.PublicKeyPath})
	}
	if cfg.Automation.Dalfox != nil {
		paths = append(paths, requiredPath{name: "automation.dalfox.targets_path", path: cfg.Automation.Dalfox.TargetsPath}, requiredPath{name: "automation.dalfox.payloads_path", path: cfg.Automation.Dalfox.PayloadsPath})
	}
	if cfg.Automation.SQLMap != nil {
		paths = append(paths, requiredPath{name: "automation.sqlmap.request_path", path: cfg.Automation.SQLMap.RequestPath})
	}
	for index, template := range cfg.NucleiTemplates {
		paths = append(paths, requiredPath{name: fmt.Sprintf("nuclei_templates[%d]", index), path: template})
	}
	for _, item := range paths {
		if item.path == "" && item.optional {
			continue
		}
		info, err := os.Stat(item.path)
		if err != nil {
			return fmt.Errorf("%s: %w", item.name, err)
		}
		if item.wantDir != info.IsDir() {
			kind := "arquivo"
			if item.wantDir {
				kind = "diretório"
			}
			return fmt.Errorf("%s precisa ser %s", item.name, kind)
		}
	}
	if cfg.ScenarioPath != "" {
		if err := validateScenarioDefinitions(cfg.ScenarioPath); err != nil {
			return err
		}
	}
	return nil
}

func absolutizeAutomation(cfg *ScopeConfig) {
	makeAbs := func(value *string) {
		if *value != "" {
			*value, _ = filepath.Abs(*value)
		}
	}
	if cfg.Automation.AmassActive != nil {
		makeAbs(&cfg.Automation.AmassActive.DatasourcesPath)
	}
	if cfg.Automation.ZAP != nil {
		makeAbs(&cfg.Automation.ZAP.TemplatePath)
		makeAbs(&cfg.Automation.ZAP.LoginScriptPath)
	}
	if cfg.Automation.JWT != nil {
		makeAbs(&cfg.Automation.JWT.PublicKeyPath)
	}
	if cfg.Automation.Dalfox != nil {
		makeAbs(&cfg.Automation.Dalfox.TargetsPath)
		makeAbs(&cfg.Automation.Dalfox.PayloadsPath)
	}
	if cfg.Automation.SQLMap != nil {
		makeAbs(&cfg.Automation.SQLMap.RequestPath)
	}
}

func routeModules(profile StackProfile, explicit []string) []string {
	if len(explicit) > 0 {
		return dedupeSorted(explicit)
	}
	values := strings.ToLower(profile.Category)
	for _, observation := range profile.Observations {
		values += " " + strings.ToLower(observation.Value)
	}
	modules := []string{"active-recon", "api-mapping", "authentication", "authorization", "business-logic", "dependencies", "reporting"}
	if strings.Contains(values, "supabase") || strings.Contains(values, "firebase") || strings.Contains(values, "baas") {
		modules = append(modules, "baas", "storage")
	} else {
		modules = append(modules, "custom-backend")
	}
	if strings.Contains(values, "billing") || strings.Contains(values, "commerce") || strings.Contains(values, "credit") {
		modules = append(modules, "race-conditions")
	}
	return dedupeSorted(modules)
}

func pipeline2Plan(cfg ScopeConfig, modules []string, dir string, policyArgs ...ValidatedPolicies) []PlannedAction {
	policies := ValidatedPolicies{Tools: ToolPolicyFile{Tools: map[string]ToolPolicyEntry{}}}
	if len(policyArgs) > 0 {
		policies = policyArgs[0]
	}
	has := func(name string) bool { return containsFold(modules, name) }
	toolMode := func(name string) string {
		if tool, ok := policies.Tools.Tools[name]; ok {
			return tool.Mode
		}
		return "disabled"
	}
	actions := []PlannedAction{}
	if has("active-recon") {
		if toolMode("feroxbuster") == "automatic" && cfg.WordlistPath != "" {
			target, _ := ValidateConfigAgainstEngagement(cfg.BaseURL, cfg.MaxRequests, cfg.RequestsPerSecond, cfg.MaxConcurrency, policies.Engagement)
			if rate, ok := exactIntegerRate(target.RatePerSecond); ok {
				args := []string{"-u", cfg.BaseURL, "-w", cfg.WordlistPath, "--rate-limit", strconv.Itoa(rate), "-t", "10", "--scan-limit", "2", "-d", "2", "-T", "7", "--time-limit", "20m", "-s", "200", "204", "301", "302", "307", "308", "401", "403", "--json", "-o", filepath.Join(dir, "ferox-out.json")}
				actions = append(actions, PlannedAction{ID: "01-feroxbuster", Module: "active-recon", Tool: "feroxbuster", PolicyTool: "feroxbuster", State: "planned", Command: args, InputPaths: []string{cfg.WordlistPath}, ArtifactPaths: []string{filepath.Join(dir, "ferox-out.json")}})
			} else {
				actions = append(actions, PlannedAction{ID: "01-feroxbuster", Module: "active-recon", Tool: "feroxbuster", State: "skipped", Reason: "feroxbuster exige rate_per_second inteiro; policy fracionária não é truncada"})
			}
		} else if toolMode("ffuf") == "automatic" && cfg.WordlistPath != "" {
			args := []string{"-u", strings.TrimRight(cfg.BaseURL, "/") + "/FUZZ", "-w", cfg.WordlistPath, "-timeout", strconv.Itoa(cfg.TimeoutSeconds), "-maxtime", "600", "-noninteractive", "-of", "json", "-o", filepath.Join(dir, "ffuf.json")}
			if rate, ok := exactIntegerRate(cfg.RequestsPerSecond); ok {
				args = append(args, "-t", strconv.Itoa(cfg.MaxConcurrency), "-rate", strconv.Itoa(rate))
			} else {
				args = append(args, "-t", "1", "-p", strconv.FormatFloat(1/cfg.RequestsPerSecond, 'f', 6, 64))
			}
			actions = append(actions, PlannedAction{ID: "01-ffuf", Module: "active-recon", Tool: "ffuf", State: "planned", Command: args, InputPaths: []string{cfg.WordlistPath}, ArtifactPaths: []string{filepath.Join(dir, "ffuf.json")}})
		} else {
			actions = append(actions, PlannedAction{ID: "01-content-discovery", Module: "active-recon", Tool: "feroxbuster", State: "skipped", Reason: "nenhum content discovery automático habilitado com wordlist_path"})
		}
		if toolMode("amass_active") == "automatic" && cfg.Automation.AmassActive != nil {
			actions = append(actions,
				PlannedAction{ID: "01-amass-sources", Module: "active-recon", Tool: "amass", PolicyTool: "amass_active", State: "planned", Command: []string{"enum", "-list"}, InputPaths: []string{cfg.Automation.AmassActive.DatasourcesPath}},
				PlannedAction{ID: "01-amass-active", Module: "active-recon", Tool: "amass", PolicyTool: "amass_active", State: "planned", Command: []string{"enum", "-active", "-brute", "-d", strings.ToLower(baseHost(cfg.BaseURL)), "-config", cfg.Automation.AmassActive.DatasourcesPath, "-max-dns-queries", "500", "-timeout", "15", "-o", filepath.Join(dir, "amass-active.txt")}, InputPaths: []string{cfg.Automation.AmassActive.DatasourcesPath}, ArtifactPaths: []string{filepath.Join(dir, "amass-active.txt")}},
			)
		} else {
			actions = append(actions, PlannedAction{ID: "01-amass-active", Module: "active-recon", Tool: "amass", PolicyTool: "amass_active", State: "skipped", Reason: "amass_active não está automático ou datasources_path não foi configurado"})
		}
	}
	if has("api-mapping") && cfg.OpenAPIURL != "" {
		actions = append(actions,
			PlannedAction{ID: "03-openapi-fetch", Module: "api-mapping", Tool: "native", State: "planned", Reason: "download pela allowlist e IP pinado antes de entregar o schema ao scanner"},
			PlannedAction{ID: "03-schemathesis", Module: "api-mapping", Tool: "schemathesis", State: "planned", Command: []string{"run", filepath.Join(dir, "openapi-schema.yaml"), "--url", cfg.BaseURL, "--workers", strconv.Itoa(cfg.MaxConcurrency), "--max-examples", "50", "--report", "junit", "--report-junit-path", filepath.Join(dir, "schemathesis-junit.xml"), "--output-sanitize", "true"}, InputPaths: []string{filepath.Join(dir, "openapi-schema.yaml")}, ArtifactPaths: []string{filepath.Join(dir, "schemathesis-junit.xml")}},
		)
	}
	if has("authorization") {
		if cfg.OpenAPIURL != "" && cfg.HadrianRolesPath != "" && cfg.HadrianAuthPath != "" {
			baseArgs := []string{"test", "rest", "--api", filepath.Join(dir, "openapi-schema.yaml"), "--roles", cfg.HadrianRolesPath, "--auth", cfg.HadrianAuthPath, "--category", "all", "--rate-limit", strconv.FormatFloat(cfg.RequestsPerSecond, 'f', -1, 64), "--timeout", strconv.Itoa(cfg.TimeoutSeconds)}
			actions = append(actions, PlannedAction{ID: "04-hadrian-dry-run", Module: "authorization", Tool: "hadrian", State: "planned", Command: append(append([]string{}, baseArgs...), "--dry-run"), Reason: "prévia obrigatória antes de testes que podem criar, alterar ou excluir fixtures", InputPaths: []string{filepath.Join(dir, "openapi-schema.yaml"), cfg.HadrianRolesPath, cfg.HadrianAuthPath}})
			state := "skipped"
			reason := "hadrian_allow_mutations está false; a prévia não autoriza mutações"
			if cfg.HadrianMutations {
				state = "planned"
				reason = "contas A/B e mutações explicitamente habilitadas no escopo autorizado"
			}
			actions = append(actions, PlannedAction{ID: "04-hadrian-authorization", Module: "authorization", Tool: "hadrian", State: state, Reason: reason, Command: append(append([]string{}, baseArgs...), "--output", "json", "--output-file", filepath.Join(dir, "hadrian.json")), InputPaths: []string{filepath.Join(dir, "openapi-schema.yaml"), cfg.HadrianRolesPath, cfg.HadrianAuthPath}, ArtifactPaths: []string{filepath.Join(dir, "hadrian.json")}})
		} else {
			actions = append(actions, PlannedAction{ID: "04-hadrian-authorization", Module: "authorization", Tool: "hadrian", State: "skipped", Reason: "OpenAPI, roles e auth do Hadrian precisam estar configurados"})
		}
		actions = append(actions, PlannedAction{ID: "04-authprobe-optional", Module: "authorization", Tool: "authprobe", State: "skipped", Reason: "opcional e mutuamente exclusivo com Hadrian; repositório canônico da ferramenta BOLA citada ainda não foi verificável"})
	}
	if has("business-logic") || has("race-conditions") {
		state, reason := "planned", "somente cenários de crédito/billing com validação numérica de estado antes e depois"
		if cfg.ScenarioPath == "" {
			state, reason = "skipped", "scenario_path de crédito/billing não configurado"
		}
		actions = append(actions, PlannedAction{ID: "05-credit-billing-scenarios", Module: "credit-billing", Tool: "native", State: state, Reason: reason})
	}
	if has("baas") {
		actions = append(actions, PlannedAction{ID: "06-supabase-rls-checker", Module: "baas", Tool: "supabase-rls-checker", PolicyTool: "supabase_rls_checker", State: "manual", Reason: "GUI local do repositório sahilahluwalia, commit pinado e pnpm --frozen-lockfile; matriz A/B obrigatória; anon key não prova falha"})
		actions = append(actions, PlannedAction{ID: "06-supashield", Module: "baas", Tool: "supashield", State: "skipped", Reason: "restrito: tag v0.3.0 publica package 0.2.1 e não oferece o contrato JSON esperado pelo normalizador"})
		actions = append(actions,
			PlannedAction{ID: "06-firepwn", Module: "baas", Tool: "firepwn", PolicyTool: "firepwn", State: "manual", Reason: "GUI desligada por padrão; usa somente contas/recursos da test-matrix e exige limpeza _pentest_"},
			PlannedAction{ID: "06-keyleak-active", Module: "baas", Tool: "keyleak-detector", PolicyTool: "keyleak", State: "skipped", Reason: "validação ativa continua desligada até a alegação de rollback ser comprovada empiricamente"},
		)
	}
	if len(cfg.NucleiTemplates) > 0 {
		args := []string{"-u", cfg.BaseURL, "-c", strconv.Itoa(cfg.MaxConcurrency), "-disable-update-check", "-jsonl-export", filepath.Join(dir, "nuclei.jsonl")}
		for _, template := range cfg.NucleiTemplates {
			args = append(args, "-t", template)
		}
		state, reason := "planned", "somente templates locais revisados"
		if rate, ok := exactIntegerRate(cfg.RequestsPerSecond); ok {
			args = append(args, "-rl", strconv.Itoa(rate))
		} else {
			state, reason = "skipped", "Nuclei aceita limite inteiro; taxa fracionária não é truncada silenciosamente"
		}
		action := PlannedAction{ID: "08-nuclei", Module: "classic-injection", Tool: "nuclei", State: state, Reason: reason, Command: args, InputPaths: append([]string{}, cfg.NucleiTemplates...), ArtifactPaths: []string{filepath.Join(dir, "nuclei.jsonl")}}
		if cfg.InteractshServer == "" {
			action.Command = append(action.Command, "-no-interactsh")
		} else {
			action.Command = append(action.Command, "-iserver", cfg.InteractshServer)
			action.SecretArgsFromEnv = []SecretArg{{Flag: "-itoken", Env: cfg.InteractshTokenEnv}}
		}
		actions = append(actions, action)
	}
	if cfg.RepositoryPath != "" && has("dependencies") {
		actions = append(actions,
			PlannedAction{ID: "09-gitleaks-repository", Module: "secrets", Tool: "gitleaks", State: "planned", Command: []string{"dir", cfg.RepositoryPath, "--no-banner", "--redact=100", "--report-format", "json", "--report-path", filepath.Join(dir, "gitleaks-repository.json")}, InputPaths: []string{cfg.RepositoryPath}, ArtifactPaths: []string{filepath.Join(dir, "gitleaks-repository.json")}},
			PlannedAction{ID: "09-osv-scanner", Module: "dependencies", Tool: "osv-scanner", State: "planned", Command: []string{"scan", "source", "--recursive", "--format=json", "--output-file", filepath.Join(dir, "osv-scanner.json"), cfg.RepositoryPath}, InputPaths: []string{cfg.RepositoryPath}, ArtifactPaths: []string{filepath.Join(dir, "osv-scanner.json")}},
		)
		if cfg.SemgrepConfigPath != "" {
			actions = append(actions, PlannedAction{ID: "09-semgrep", Module: "sast", Tool: "semgrep", State: "planned", Command: []string{"scan", "--config", cfg.SemgrepConfigPath, "--metrics=off", "--disable-version-check", "--json", "--output", filepath.Join(dir, "semgrep.json"), cfg.RepositoryPath}, InputPaths: []string{cfg.SemgrepConfigPath, cfg.RepositoryPath}, ArtifactPaths: []string{filepath.Join(dir, "semgrep.json")}})
		} else {
			actions = append(actions, PlannedAction{ID: "09-semgrep", Module: "sast", Tool: "semgrep", State: "skipped", Reason: "semgrep_config_path local não configurado; o runner não usa regras remotas automaticamente"})
		}
		actions = append(actions,
			PlannedAction{ID: "09-trivy", Module: "dependencies", Tool: "trivy", State: "planned", Command: []string{"fs", "--scanners", "vuln,misconfig", "--format", "json", "--output", filepath.Join(dir, "trivy.json"), cfg.RepositoryPath}, InputPaths: []string{cfg.RepositoryPath}, ArtifactPaths: []string{filepath.Join(dir, "trivy.json")}},
			PlannedAction{ID: "09-rlsgate-future", Module: "sast", Tool: "rlsgate", State: "skipped", Reason: "etapa futura; somente revisão estática de repositório autorizado sob NDA"},
		)
	}
	if toolMode("zap") == "automatic" && cfg.Automation.ZAP != nil {
		actions = append(actions, PlannedAction{ID: "10-zap-authenticated", Module: "dast", Tool: "zap", PolicyTool: "zap", State: "planned", Reason: "plano final renderizado pelo runner com autenticação verificada e durações finitas", Command: []string{"-cmd", "-silent", "-autorun", filepath.Join(dir, "zap-plan.rendered.yaml")}, InputPaths: []string{cfg.Automation.ZAP.TemplatePath, cfg.Automation.ZAP.LoginScriptPath}, ArtifactPaths: []string{filepath.Join(dir, "zap-report.json"), filepath.Join(dir, "zap-plan.rendered.yaml")}})
	} else {
		actions = append(actions, PlannedAction{ID: "10-zap-authenticated", Module: "dast", Tool: "zap", State: "skipped", Reason: "ZAP automático exige configuração de autenticação completa"})
	}
	if toolMode("jwt_tool") == "automatic" && cfg.Automation.JWT != nil {
		baseArgs := []string{"-t", cfg.Automation.JWT.TargetURL, "-rh", "Authorization: Bearer "}
		actions = append(actions,
			PlannedAction{ID: "11-jwt-baseline", Module: "authentication", Tool: "jwt_tool", PolicyTool: "jwt_tool", State: "planned", Command: append([]string{}, baseArgs...), SecretArgsFromEnv: []SecretArg{{Env: cfg.Automation.JWT.TokenEnv, Position: "prepend"}}},
			PlannedAction{ID: "11-jwt-alg-none", Module: "authentication", Tool: "jwt_tool", PolicyTool: "jwt_tool", State: "planned", Command: append([]string{"-X", "a"}, baseArgs...), SecretArgsFromEnv: []SecretArg{{Env: cfg.Automation.JWT.TokenEnv, Position: "prepend"}}},
		)
		if cfg.Automation.JWT.ExpectedAlgorithm == "RS256" && cfg.Automation.JWT.PublicKeyPath != "" {
			actions = append(actions, PlannedAction{ID: "11-jwt-rs-hs-confusion", Module: "authentication", Tool: "jwt_tool", PolicyTool: "jwt_tool", State: "planned", Command: append([]string{"-X", "k", "-pk", cfg.Automation.JWT.PublicKeyPath}, baseArgs...), SecretArgsFromEnv: []SecretArg{{Env: cfg.Automation.JWT.TokenEnv, Position: "prepend"}}, InputPaths: []string{cfg.Automation.JWT.PublicKeyPath}})
		}
	} else {
		actions = append(actions, PlannedAction{ID: "11-jwt-tool", Module: "authentication", Tool: "jwt_tool", State: "skipped", Reason: "jwt_tool automático exige token de teste e issuer/audience/algoritmo documentados"})
	}
	if toolMode("dalfox") == "automatic" && cfg.Automation.Dalfox != nil {
		actions = append(actions, PlannedAction{ID: "12-dalfox", Module: "xss", Tool: "dalfox", PolicyTool: "dalfox", State: "planned", Reason: "somente alvos de reflexão pré-listados; stored XSS não é executado por este wrapper", Command: []string{"scan", cfg.Automation.Dalfox.TargetsPath, "--custom-payload", cfg.Automation.Dalfox.PayloadsPath, "--skip-bav", "--skip-mining", "--workers", "5", "--delay", "200", "--format", "jsonl", "--output", filepath.Join(dir, "dalfox.jsonl")}, InputPaths: []string{cfg.Automation.Dalfox.TargetsPath, cfg.Automation.Dalfox.PayloadsPath}, ArtifactPaths: []string{filepath.Join(dir, "dalfox.jsonl")}})
	} else {
		actions = append(actions, PlannedAction{ID: "12-dalfox", Module: "xss", Tool: "dalfox", State: "skipped", Reason: "Dalfox exige lista explícita de reflexões e payloads locais conservadores"})
	}
	if toolMode("sqlmap") == "automatic" && cfg.Automation.SQLMap != nil {
		technique := "BEU"
		flags := policies.Tools.Tools["sqlmap"].Flags
		if flags.AllowTimeBasedSQLi {
			technique += "T"
		}
		if flags.AllowStackedQueries {
			technique += "S"
		}
		actions = append(actions, PlannedAction{ID: "12-sqlmap", Module: "injection", Tool: "sqlmap", PolicyTool: "sqlmap", State: "planned", Reason: "somente requisição capturada e parâmetro explícito contra RPC/backend customizado", Command: []string{"-r", cfg.Automation.SQLMap.RequestPath, "-p", cfg.Automation.SQLMap.Parameter, "--method=POST", "--risk=1", "--level=2", "--technique=" + technique, "--delay=1", "--time-sec=5", "--safe-url=" + cfg.Automation.SQLMap.SafeURL, "--safe-freq=5", "--batch", "-o", "--output-dir", filepath.Join(dir, "sqlmap-output")}, InputPaths: []string{cfg.Automation.SQLMap.RequestPath}, ArtifactPaths: []string{filepath.Join(dir, "sqlmap-output")}})
	} else {
		actions = append(actions, PlannedAction{ID: "12-sqlmap", Module: "injection", Tool: "sqlmap", State: "skipped", Reason: "sqlmap exige requisição real capturada, parâmetro e target_kind explícitos"})
	}
	return actions
}

func downloadOpenAPISchema(ctx context.Context, cfg ScopeConfig, policy EngagementPolicy, destination string) error {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	client := NewScopedClientWithPins(cfg.AllowedSchemes, cfg.AllowedHosts, cfg.AllowedPorts, cfg.AllowPrivateIPs, 4, cfg.RequestsPerSecond, timeout, engagementPins(policy))
	observation, data := doObservedRequest(ctx, client, http.MethodGet, cfg.OpenAPIURL, nil, 4<<20)
	if observation.Error != "" {
		return fmt.Errorf("download: %s", observation.Error)
	}
	if observation.Status != http.StatusOK {
		return fmt.Errorf("download retornou status %d", observation.Status)
	}
	if len(data) == 0 || len(data) >= 4<<20 {
		return fmt.Errorf("schema vazio ou maior que 4 MiB")
	}
	return os.WriteFile(destination, data, 0o600)
}

func profileSubjectMatchesHost(subject, host string) bool {
	subject = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(subject), "."))
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	return subject != "" && (host == subject || strings.HasSuffix(host, "."+subject))
}

func validateScenarioDefinitions(path string) error {
	var file ScenarioFile
	if err := ReadJSON(path, &file); err != nil {
		return fmt.Errorf("scenario_path: %w", err)
	}
	if len(file.Scenarios) == 0 {
		return fmt.Errorf("scenario_path não contém cenários")
	}
	for _, scenario := range file.Scenarios {
		if scenario.Name == "" || !containsFold([]string{"credit", "billing"}, scenario.Domain) {
			return fmt.Errorf("cenário %q: domain deve ser credit ou billing", scenario.Name)
		}
		if !containsFold([]string{"business-logic", "race-conditions"}, scenario.Module) {
			return fmt.Errorf("cenário %s: scenario runner aceita somente business-logic ou race-conditions", scenario.Name)
		}
		if !strings.HasPrefix(scenario.Path, "/") || strings.HasPrefix(scenario.Path, "//") {
			return fmt.Errorf("cenário %s: path precisa ser relativo ao host autorizado", scenario.Name)
		}
		if scenario.Concurrency < 1 {
			return fmt.Errorf("cenário %s: concurrency é obrigatório", scenario.Name)
		}
		if len(scenario.ExpectedStatuses) == 0 {
			return fmt.Errorf("cenário %s: expected_statuses é obrigatório", scenario.Name)
		}
		if scenario.StateCheck == nil || !containsFold([]string{"GET", "HEAD"}, scenario.StateCheck.Method) || !strings.HasPrefix(scenario.StateCheck.Path, "/") || strings.HasPrefix(scenario.StateCheck.Path, "//") || !strings.HasPrefix(scenario.StateCheck.JSONPointer, "/") {
			return fmt.Errorf("cenário %s: state_check GET/HEAD, path relativo e json_pointer são obrigatórios", scenario.Name)
		}
	}
	return nil
}

func runScenarios(ctx context.Context, cfg ScopeConfig, base *url.URL, routedModules []string, policyArgs ...EngagementPolicy) ([]ScenarioResult, error) {
	if err := validateScenarioDefinitions(cfg.ScenarioPath); err != nil {
		return nil, err
	}
	var file ScenarioFile
	if err := ReadJSON(cfg.ScenarioPath, &file); err != nil {
		return nil, err
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	client := NewScopedClient(cfg.AllowedSchemes, cfg.AllowedHosts, cfg.AllowedPorts, cfg.AllowPrivateIPs, cfg.MaxRequests, cfg.RequestsPerSecond, timeout)
	if len(policyArgs) > 0 {
		client = NewScopedClientWithPins(cfg.AllowedSchemes, cfg.AllowedHosts, cfg.AllowedPorts, cfg.AllowPrivateIPs, cfg.MaxRequests, cfg.RequestsPerSecond, timeout, engagementPins(policyArgs[0]))
	}
	results := []ScenarioResult{}
	for _, scenario := range file.Scenarios {
		if !containsFold(routedModules, scenario.Module) {
			continue
		}
		if scenario.Concurrency > cfg.MaxConcurrency {
			return nil, fmt.Errorf("cenário %s excede max_concurrency", scenario.Name)
		}
		target, err := base.Parse(scenario.Path)
		if err != nil {
			return nil, err
		}
		if !containsFold(cfg.AllowedHosts, target.Hostname()) {
			return nil, fmt.Errorf("cenário %s saiu do host permitido", scenario.Name)
		}
		before, err := readScenarioState(ctx, client, base, *scenario.StateCheck)
		if err != nil {
			return nil, fmt.Errorf("cenário %s, estado inicial: %w", scenario.Name, err)
		}
		resultChannel := make(chan ScenarioResult, scenario.Concurrency)
		var wait sync.WaitGroup
		for index := 0; index < scenario.Concurrency; index++ {
			wait.Add(1)
			go func() { defer wait.Done(); resultChannel <- executeScenario(ctx, client, target.String(), scenario) }()
		}
		wait.Wait()
		close(resultChannel)
		after, err := readScenarioState(ctx, client, base, *scenario.StateCheck)
		if err != nil {
			return nil, fmt.Errorf("cenário %s, estado final: %w", scenario.Name, err)
		}
		actualDelta := after - before
		statePassed := math.Abs(actualDelta-scenario.StateCheck.ExpectedDelta) <= 1e-9
		for result := range resultChannel {
			result.StateBefore = before
			result.StateAfter = after
			result.ExpectedDelta = scenario.StateCheck.ExpectedDelta
			result.ActualDelta = actualDelta
			result.Passed = result.Passed && statePassed
			if !statePassed && result.Error == "" {
				result.Error = fmt.Sprintf("delta de estado %.6f; esperado %.6f", actualDelta, scenario.StateCheck.ExpectedDelta)
			}
			results = append(results, result)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Name == results[j].Name {
			return results[i].DurationMS < results[j].DurationMS
		}
		return results[i].Name < results[j].Name
	})
	return results, nil
}

func engagementPins(policy EngagementPolicy) map[string][]string {
	result := map[string][]string{}
	for _, target := range policy.Environment.Targets {
		result[target.Host] = append(result[target.Host], target.IP)
	}
	for _, api := range policy.Environment.ExternalAPIs {
		result[api.Host] = append(result[api.Host], api.IP)
	}
	return result
}

func readScenarioState(ctx context.Context, client *http.Client, base *url.URL, check ScenarioStateCheck) (float64, error) {
	target, err := base.Parse(check.Path)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(check.Method), target.String(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "EmpresaSecurity-AuthorizedAssessment/0.2")
	for header, envName := range check.HeadersFromEnv {
		value, ok := os.LookupEnv(envName)
		if !ok || value == "" {
			return 0, fmt.Errorf("variável de ambiente ausente: %s", envName)
		}
		req.Header.Set(header, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("state_check retornou status %d", resp.StatusCode)
	}
	var document any
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return 0, fmt.Errorf("JSON inválido: %w", err)
	}
	value, err := valueAtJSONPointer(document, check.JSONPointer)
	if err != nil {
		return 0, err
	}
	switch typed := value.(type) {
	case json.Number:
		return typed.Float64()
	case float64:
		return typed, nil
	default:
		return 0, fmt.Errorf("json_pointer não aponta para número")
	}
}

func valueAtJSONPointer(document any, pointer string) (any, error) {
	current := document
	for _, rawToken := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		token := strings.ReplaceAll(strings.ReplaceAll(rawToken, "~1", "/"), "~0", "~")
		switch typed := current.(type) {
		case map[string]any:
			value, ok := typed[token]
			if !ok {
				return nil, fmt.Errorf("json_pointer ausente: %s", pointer)
			}
			current = value
		case []any:
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(typed) {
				return nil, fmt.Errorf("índice inválido em json_pointer: %s", pointer)
			}
			current = typed[index]
		default:
			return nil, fmt.Errorf("json_pointer atravessa valor escalar: %s", pointer)
		}
	}
	return current, nil
}

func executeScenario(ctx context.Context, client *http.Client, target string, scenario Scenario) ScenarioResult {
	started := time.Now()
	result := ScenarioResult{Name: scenario.Name, Module: scenario.Module}
	body := bytes.NewReader(scenario.Body)
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(scenario.Method), target, body)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	req.Header.Set("User-Agent", "EmpresaSecurity-AuthorizedAssessment/0.1")
	if len(scenario.Body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	for header, envName := range scenario.HeadersFromEnv {
		value, ok := os.LookupEnv(envName)
		if !ok || value == "" {
			result.Error = "variável de ambiente ausente: " + envName
			return result
		}
		req.Header.Set(header, value)
	}
	resp, err := client.Do(req)
	result.DurationMS = time.Since(started).Milliseconds()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	result.Status = resp.StatusCode
	result.BodyBytes = len(data)
	result.BodySHA256 = HashBytes(data)
	result.Passed = len(scenario.ExpectedStatuses) == 0 || containsInt(scenario.ExpectedStatuses, resp.StatusCode)
	return result
}

func writePipeline2Report(dir string, cfg ScopeConfig, profile StackProfile, manifest Manifest, results []ScenarioResult) error {
	var builder strings.Builder
	builder.WriteString("# Entrega técnica — " + cfg.CaseID + "\n\n")
	builder.WriteString("Modo: `" + manifest.Mode + "`  \nPerfil: `" + profile.Category + "`\n\n")
	builder.WriteString("## Módulos roteados\n\n")
	for _, module := range profile.ApprovedModules {
		builder.WriteString("- " + module + "\n")
	}
	builder.WriteString("\n## Execução\n\n")
	for _, action := range manifest.Actions {
		builder.WriteString(fmt.Sprintf("- `%s`: %s (%s)\n", action.ID, action.State, action.Tool))
	}
	if len(results) > 0 {
		builder.WriteString("\n## Cenários de crédito/billing\n\n")
		for _, result := range results {
			builder.WriteString(fmt.Sprintf("- %s: status %d, passed=%t, delta=%.6f (esperado %.6f), body_sha256=%s\n", result.Name, result.Status, result.Passed, result.ActualDelta, result.ExpectedDelta, result.BodySHA256))
		}
	}
	builder.WriteString("\n## Limitações\n\nSaídas automáticas são candidatas. A confirmação, severidade, reprodução e recomendação exigem revisão humana antes da entrega ao cliente.\n")
	return os.WriteFile(filepath.Join(dir, "final-report.md"), []byte(builder.String()), 0o600)
}

func writeChecksums(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	lines := []string{}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "SHA256SUMS.txt" {
			continue
		}
		hash, err := HashFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		lines = append(lines, hash+"  "+entry.Name())
	}
	sort.Strings(lines)
	return os.WriteFile(filepath.Join(dir, "SHA256SUMS.txt"), []byte(strings.Join(lines, "\n")+"\n"), 0o600)
}

func VerifyCase(dir string) error {
	data, err := os.ReadFile(filepath.Join(dir, "SHA256SUMS.txt"))
	if err != nil {
		return err
	}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("linha de checksum inválida")
		}
		hash, err := HashFile(filepath.Join(dir, parts[1]))
		if err != nil {
			return err
		}
		if subtle.ConstantTimeCompare([]byte(hash), []byte(parts[0])) != 1 {
			return fmt.Errorf("checksum divergente: %s", parts[1])
		}
	}
	return nil
}

func containsFold(items []string, wanted string) bool {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(wanted)) {
			return true
		}
	}
	return false
}
func containsInt(items []int, wanted int) bool {
	for _, item := range items {
		if item == wanted {
			return true
		}
	}
	return false
}

func baseHost(raw string) string {
	if parsed, err := url.Parse(raw); err == nil {
		return parsed.Hostname()
	}
	return ""
}

func actionHasState(actions []PlannedAction, id, state string) bool {
	for _, action := range actions {
		if action.ID == id && action.State == state {
			return true
		}
	}
	return false
}
