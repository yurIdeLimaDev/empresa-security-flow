package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var scriptSourcePattern = regexp.MustCompile(`(?i)<script[^>]+src=["']([^"']+)["']`)

func RunPipeline1(ctx context.Context, cfg LeadConfig, execute bool) error {
	base, err := validateBaseURL(cfg.BaseURL, cfg.Domain)
	if err != nil {
		return err
	}
	if cfg.CORSOrigin != "" {
		origin, parseErr := url.Parse(cfg.CORSOrigin)
		if parseErr != nil || origin.Scheme == "" || origin.Hostname() == "" || origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" || (origin.Scheme != "http" && origin.Scheme != "https") {
			return fmt.Errorf("cors_origin deve ser uma origem http/https absoluta sob controle da operação")
		}
	}
	if cfg.ApplicationCategory != "" && !containsFold([]string{"baas", "web-application", "custom-backend", "static", "unknown"}, cfg.ApplicationCategory) {
		return fmt.Errorf("application_category inválida")
	}
	normalizeLimits(&cfg.MaxRequests, &cfg.RequestsPerSecond)
	if cfg.MaxDownloadBytes <= 0 {
		cfg.MaxDownloadBytes = 5 << 20
	}
	if cfg.MaxDownloadBytes > 20<<20 {
		cfg.MaxDownloadBytes = 20 << 20
	}
	lock, lockPath, err := ResolveToolLock(cfg.ToolLockPath)
	if err != nil {
		return fmt.Errorf("cadeia de ferramentas: %w", err)
	}
	policies, err := LoadPolicies(cfg.Policies, 1, &lock)
	if err != nil {
		return err
	}
	targetPolicy, err := ValidateConfigAgainstEngagement(cfg.BaseURL, cfg.MaxRequests, cfg.RequestsPerSecond, 1, policies.Engagement)
	if err != nil {
		return err
	}

	dir, err := caseDir(cfg.OutputRoot, cfg.CaseID, "pipeline-1")
	if err != nil {
		return err
	}
	network, err := GenerateIsolationScripts(dir, policies.Engagement)
	if err != nil {
		return err
	}
	manifest := Manifest{
		CaseID: cfg.CaseID, Pipeline: 1, Mode: mode(execute), StartedAt: now(),
		ConfigSHA256: HashJSON(cfg),
	}
	manifest.Actions = pipeline1Plan(cfg, dir)
	applyPipeline1ToolPolicy(&manifest, policies.Tools)
	if err := WriteJSON(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		return err
	}

	profile := StackProfile{
		Subject: cfg.Domain, Status: "hypothesis", CollectedAt: now(),
		Category: cfg.ApplicationCategory,
		Observations: []Observation{{
			Category: "lead_origin", Value: cfg.LeadOrigin, SourceType: "lead_origin",
			SourceURL: cfg.BaseURL, ObservedAt: now(), Confidence: "low",
			Limitations: "Etiqueta inicial; não autoritativa.",
		}},
	}
	if !execute {
		profile.Observations = filterEmptyObservations(profile.Observations)
		if err := WriteJSON(filepath.Join(dir, "stack-profile.json"), profile); err != nil {
			return err
		}
		manifest.FinishedAt = now()
		if err := WriteJSON(filepath.Join(dir, "manifest.json"), manifest); err != nil {
			return err
		}
		if err := writePipeline1Report(dir, cfg, profile, manifest); err != nil {
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
	runOptions := GovernedRunOptions{Pipeline: 1, Policies: policies, Lock: lock, SBOMPath: filepath.Join(filepath.Dir(lockPath), "sbom.cdx.json"), Network: network, CaseDir: dir}

	for i := range manifest.Actions {
		action := &manifest.Actions[i]
		if action.Tool == "native" || action.Tool == "gitleaks" || action.Tool == "jsluice" || action.Tool == "trufflehog" || action.Tool == "dnsreaper" || action.PolicyTool == "hibp" || action.State == "manual" || action.State == "skipped" {
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

	hosts := []string{strings.ToLower(base.Hostname())}
	client := NewScopedClientWithPins([]string{"http", "https"}, hosts, []int{targetPolicy.Port}, false, cfg.MaxRequests, cfg.RequestsPerSecond, 15*time.Second, map[string][]string{targetPolicy.Host: []string{targetPolicy.IP}})
	observations, body := observePublicHTTP(ctx, client, base, cfg.CORSOrigin, cfg.MaxDownloadBytes)
	emails, contactObservations := DiscoverPublicEmails(ctx, client, base, cfg.PublicContactPaths, cfg.MaxDownloadBytes)
	observations = append(observations, contactObservations...)
	if err := WriteJSON(filepath.Join(dir, "http-observations.json"), observations); err != nil {
		return err
	}
	dnsResult := observeDNS(ctx, cfg.Domain, policies.Engagement.Environment.DNSResolverIP)
	if err := WriteJSON(filepath.Join(dir, "dns-observations.json"), dnsResult); err != nil {
		return err
	}
	if ctAPI, ok := findExternalAPI(policies.Engagement, "crt.sh", "certificate-transparency"); ok {
		ctClient := NewScopedClientWithPins([]string{"https"}, []string{ctAPI.Host}, []int{ctAPI.Port}, false, 1, 0.2, 20*time.Second, map[string][]string{ctAPI.Host: []string{ctAPI.IP}})
		ctResult, ctErr := observeCT(ctx, ctClient, cfg.Domain)
		if ctErr != nil {
			manifest.Warnings = append(manifest.Warnings, "crt.sh: "+ctErr.Error())
		} else if err := WriteJSON(filepath.Join(dir, "certificate-transparency.json"), ctResult); err != nil {
			return err
		}
	} else {
		manifest.Warnings = append(manifest.Warnings, "crt.sh não consultado: external_apis sem purpose=certificate-transparency")
	}

	bundlesDir := filepath.Join(dir, "bundles")
	_ = os.MkdirAll(bundlesDir, 0o700)
	bundleURLs := downloadBundles(ctx, client, base, body, bundlesDir, cfg.MaxDownloadBytes, &manifest)
	candidatePath := filepath.Join(dir, "dnsreaper-candidates.txt")
	candidates, curateErr := CurateDNSReaperCandidates(cfg.Domain, filepath.Join(dir, "subfinder.jsonl"), filepath.Join(dir, "amass.txt"))
	if curateErr != nil {
		manifest.Warnings = append(manifest.Warnings, "dnsReaper curadoria: "+curateErr.Error())
	} else if err := WriteLines(candidatePath, candidates); err != nil {
		return err
	}
	for i := range manifest.Actions {
		action := &manifest.Actions[i]
		if action.Tool == "native" {
			action.State = "completed"
		}
		if action.Tool == "jsluice" {
			bundleFiles, globErr := filepath.Glob(filepath.Join(bundlesDir, "*.js"))
			if globErr != nil || len(bundleFiles) == 0 {
				action.State = "skipped"
				action.Reason = "nenhum bundle JavaScript local para análise"
				continue
			}
			action.Command = append(action.Command, bundleFiles...)
			action.InputPaths = append([]string{}, bundleFiles...)
		}
		if action.Tool == "dnsreaper" {
			if len(candidates) == 0 {
				action.State, action.Reason = "skipped", "nenhum subdomínio curado foi produzido por subfinder/amass passivo"
				continue
			}
			action.InputPaths = []string{candidatePath}
		}
		if action.Tool == "trufflehog" {
			action.InputPaths = []string{bundlesDir}
		}
		if action.PolicyTool == "hibp" {
			action.InputPaths = []string{filepath.Join(dir, "http-observations.json")}
			if err := RunHIBPContainer(ctx, runOptions, cfg, emails, action); err != nil {
				if IsGateError(err) {
					_ = WriteJSON(filepath.Join(dir, "manifest.json"), manifest)
					return fmt.Errorf("gate %s: %w", action.ID, err)
				}
				manifest.Warnings = append(manifest.Warnings, err.Error())
			}
			continue
		}
		if action.Tool != "gitleaks" && action.Tool != "jsluice" && action.Tool != "trufflehog" && action.Tool != "dnsreaper" {
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
		if action.Tool == "dnsreaper" && action.State == "completed" {
			if postErr := PostprocessDNSReaper(filepath.Join(dir, "dnsreaper-results.json"), filepath.Join(dir, "dnsreaper-candidates-report.json"), filepath.Join(dir, "dnsreaper-unlikely.json")); postErr != nil {
				manifest.Warnings = append(manifest.Warnings, postErr.Error())
			}
		}
	}
	profile.Observations = append(profile.Observations, inferStack(cfg.BaseURL, observations, body, bundleURLs)...)
	profile.Observations = filterEmptyObservations(profile.Observations)
	profile.Category = categorizeProfile(profile)
	if err := WriteJSON(filepath.Join(dir, "stack-profile.json"), profile); err != nil {
		return err
	}

	manifest.FinishedAt = now()
	_, normalizeErr := NormalizeKnownOutputs(dir, cfg.CaseID+"-pipeline-1", "public:"+strings.ToLower(cfg.Domain), []OutputAdapter{
		{Tool: "subfinder", File: "subfinder.jsonl"}, {Tool: "amass", File: "amass.txt"}, {Tool: "httpx", File: "httpx.jsonl"},
		{Tool: "testssl", File: "testssl.json"}, {Tool: "gitleaks", File: "gitleaks-bundles.json"}, {Tool: "trufflehog", File: "10-trufflehog-bundles.stdout.log"},
		{Tool: "jsluice", File: "06-jsluice-urls.stdout.log"}, {Tool: "dnsreaper", File: "dnsreaper-candidates-report.json", ProvenanceFile: "dnsreaper-results.json"}, {Tool: "hibp", File: "hibp-email-results.json"},
	})
	if normalizeErr != nil {
		manifest.Warnings = append(manifest.Warnings, normalizeErr.Error())
	}
	if err := WriteJSON(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		return err
	}
	if err := writePipeline1Report(dir, cfg, profile, manifest); err != nil {
		return err
	}
	return writeChecksums(dir)
}

// applyPipeline1ToolPolicy turns product-specific omissions into explicit
// skips before any runner action is attempted. The default commercial policy
// still enables its original tools; Vexkeep Check intentionally has a smaller
// allowlist and must not fail open into HIBP, TruffleHog or Gitleaks.
func applyPipeline1ToolPolicy(manifest *Manifest, policy ToolPolicyFile) {
	for index := range manifest.Actions {
		action := &manifest.Actions[index]
		if action.Tool == "native" || action.State == "skipped" || action.State == "manual" {
			continue
		}
		key := action.PolicyTool
		if key == "" {
			key = action.Tool
		}
		entry, exists := policy.Tools[key]
		if !exists || entry.Mode != "automatic" || entry.Pipeline != 1 {
			action.State = "skipped"
			action.Reason = "não habilitado pela política específica deste produto"
		}
	}
}

func pipeline1Plan(cfg LeadConfig, dir string) []PlannedAction {
	actions := []PlannedAction{
		{ID: "01-subfinder", Module: "asset-discovery", Tool: "subfinder", State: "planned", Command: []string{"-d", cfg.Domain, "-silent", "-json", "-o", filepath.Join(dir, "subfinder.jsonl")}, ArtifactPaths: []string{filepath.Join(dir, "subfinder.jsonl")}},
		{ID: "02-amass-passive", Module: "asset-discovery", Tool: "amass", State: "planned", Command: []string{"enum", "-passive", "-d", cfg.Domain, "-o", filepath.Join(dir, "amass.txt")}, ArtifactPaths: []string{filepath.Join(dir, "amass.txt")}},
		{ID: "03-httpx", Module: "fingerprint", Tool: "httpx", State: "planned", Command: []string{"-u", cfg.BaseURL, "-json", "-td", "-sc", "-title", "-location", "-o", filepath.Join(dir, "httpx.jsonl")}, ArtifactPaths: []string{filepath.Join(dir, "httpx.jsonl")}},
		{ID: "04-native-public-http", Module: "http-posture", Tool: "native", State: "planned", Reason: "GET/HEAD e, se configurado, CORS com budget, escopo, IP pinado e redirecionamento restritos"},
		{ID: "05-testssl", Module: "tls", Tool: "testssl", State: "planned", Command: []string{"--quiet", "--jsonfile-pretty", filepath.Join(dir, "testssl.json"), cfg.BaseURL}, ArtifactPaths: []string{filepath.Join(dir, "testssl.json")}},
		{ID: "06-jsluice-urls", Module: "bundle-static-analysis", Tool: "jsluice", State: "planned", Command: []string{"urls", "-R", cfg.BaseURL}, Reason: "somente análise estática de bundles já baixados do mesmo host"},
		{ID: "07-gitleaks-bundles", Module: "bundle-static-analysis", Tool: "gitleaks", State: "planned", Command: []string{"dir", filepath.Join(dir, "bundles"), "--no-banner", "--redact=100", "--report-format", "json", "--report-path", filepath.Join(dir, "gitleaks-bundles.json")}, InputPaths: []string{filepath.Join(dir, "bundles")}, ArtifactPaths: []string{filepath.Join(dir, "gitleaks-bundles.json")}},
		{ID: "08-keyleak-detector", Module: "bundle-static-analysis", Tool: "keyleak-detector", State: "skipped", Reason: "desligado no Pipeline 1: os modos remotos fazem crawl e podem validar BaaS ativamente; uso ativo pertence ao Pipeline 2 autorizado"},
		{ID: "09-dnsreaper", Module: "asset-discovery", Tool: "dnsreaper", State: "planned", PolicyTool: "dnsreaper", Reason: "automático somente sobre saída curada; todo resultado permanece candidato e nenhum recurso é reivindicado", Command: []string{"file", "--filename", filepath.Join(dir, "dnsreaper-candidates.txt"), "--out", filepath.Join(dir, "dnsreaper-results.json"), "--out-format", "json", "--parallelism", "10"}, ArtifactPaths: []string{filepath.Join(dir, "dnsreaper-results.json")}, AllowedExitCodes: []int{1}},
		{ID: "10-trufflehog-bundles", Module: "bundle-static-analysis", Tool: "trufflehog", State: "planned", Reason: "automático, filesystem local e sem qualquer verificação externa", Command: []string{"filesystem", filepath.Join(dir, "bundles"), "--no-verification", "--json"}, InputPaths: []string{filepath.Join(dir, "bundles")}},
		{ID: "11-hibp-email", Module: "breach-exposure", Tool: "hibp", PolicyTool: "hibp", State: "planned", Reason: "somente breachedaccount para e-mail corporativo publicado nas páginas configuradas; busca por domínio não existe no runner", ArtifactPaths: []string{filepath.Join(dir, "hibp-email-results.json")}},
	}
	if rate, ok := exactIntegerRate(cfg.RequestsPerSecond); ok {
		actions[2].Command = append(actions[2].Command, "-rl", strconv.Itoa(rate))
	} else {
		actions[2].State = "skipped"
		actions[2].Reason = "httpx aceita limite inteiro; taxa fracionária não é truncada silenciosamente"
	}
	return actions
}

func observePublicHTTP(ctx context.Context, client *http.Client, base *url.URL, corsOrigin string, maxBytes int64) ([]HTTPObservation, []byte) {
	paths := []string{"/", "/robots.txt", "/sitemap.xml", "/.well-known/security.txt", "/swagger.json", "/api/docs", "/openapi.json", "/graphql", "/.git/config"}
	results := make([]HTTPObservation, 0, len(paths)+2)
	var homeBody []byte
	for _, path := range paths {
		target := base.ResolveReference(&url.URL{Path: path})
		result, body := doObservedRequest(ctx, client, http.MethodGet, target.String(), nil, maxBytes)
		results = append(results, result)
		if path == "/" {
			homeBody = body
		}
	}
	head, _ := doObservedRequest(ctx, client, http.MethodHead, base.String(), nil, 0)
	results = append(results, head)
	if corsOrigin != "" {
		corsHeaders := http.Header{"Origin": []string{corsOrigin}}
		cors, _ := doObservedRequest(ctx, client, http.MethodGet, base.String(), corsHeaders, 2048)
		results = append(results, cors)
	}
	return results, homeBody
}

func doObservedRequest(ctx context.Context, client *http.Client, method, target string, headers http.Header, maxBytes int64) (HTTPObservation, []byte) {
	req, err := http.NewRequestWithContext(ctx, method, target, nil)
	if err != nil {
		return HTTPObservation{URL: target, Method: method, Error: err.Error()}, nil
	}
	req.Header.Set("User-Agent", "EmpresaSecurity-PublicObservation/0.1")
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return HTTPObservation{URL: target, Method: method, Error: err.Error()}, nil
	}
	defer resp.Body.Close()
	var body []byte
	if maxBytes > 0 {
		body, _ = io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
		if int64(len(body)) > maxBytes {
			body = body[:maxBytes]
		}
	}
	return HTTPObservation{
		URL: target, Method: method, Status: resp.StatusCode, Headers: safeHeaders(resp.Header),
		BodyBytes: len(body), BodySHA256: HashBytes(body), ContentType: resp.Header.Get("Content-Type"),
	}, body
}

func safeHeaders(headers http.Header) map[string][]string {
	allowed := map[string]struct{}{
		"Content-Type": {}, "Content-Security-Policy": {}, "Strict-Transport-Security": {},
		"X-Frame-Options": {}, "X-Content-Type-Options": {}, "Access-Control-Allow-Origin": {},
		"Access-Control-Allow-Credentials": {}, "Server": {}, "Location": {}, "Set-Cookie": {},
	}
	result := make(map[string][]string)
	for key := range allowed {
		for _, value := range headers.Values(key) {
			if key == "Set-Cookie" {
				parts := strings.Split(value, ";")
				if len(parts) > 0 {
					parts[0] = "[redacted]"
				}
				value = strings.Join(parts, ";")
			}
			result[key] = append(result[key], value)
		}
	}
	return result
}

func observeDNS(ctx context.Context, domain, resolverIP string) map[string]any {
	result := map[string]any{"domain": domain, "observed_at": now()}
	resolver := &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "udp", net.JoinHostPort(resolverIP, "53"))
	}}
	if ips, err := resolver.LookupHost(ctx, domain); err == nil {
		result["addresses"] = ips
	} else {
		result["address_error"] = err.Error()
	}
	if cname, err := resolver.LookupCNAME(ctx, domain); err == nil {
		result["cname"] = cname
	}
	if mx, err := resolver.LookupMX(ctx, domain); err == nil {
		values := make([]string, 0, len(mx))
		for _, item := range mx {
			values = append(values, item.Host)
		}
		result["mx"] = values
	}
	if txt, err := resolver.LookupTXT(ctx, domain); err == nil {
		result["txt_count"] = len(txt)
	}
	return result
}

func observeCT(ctx context.Context, client *http.Client, domain string) ([]map[string]any, error) {
	target := "https://crt.sh/?q=" + url.QueryEscape("%."+domain) + "&output=json"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	req.Header.Set("User-Agent", "EmpresaSecurity-PublicObservation/0.1")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var rows []map[string]any
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 4<<20))
	if err := decoder.Decode(&rows); err != nil {
		return nil, err
	}
	if len(rows) > 500 {
		rows = rows[:500]
	}
	return rows, nil
}

func findExternalAPI(policy EngagementPolicy, host, purpose string) (ExternalAPI, bool) {
	for _, api := range policy.Environment.ExternalAPIs {
		if strings.EqualFold(api.Host, host) && strings.EqualFold(api.Purpose, purpose) {
			return api, true
		}
	}
	return ExternalAPI{}, false
}

func downloadBundles(ctx context.Context, client *http.Client, base *url.URL, body []byte, dir string, maxTotal int64, manifest *Manifest) []string {
	matches := scriptSourcePattern.FindAllSubmatch(body, -1)
	result := make([]string, 0, len(matches))
	remaining := maxTotal
	for index, match := range matches {
		if index >= 20 || remaining <= 0 {
			break
		}
		ref, err := url.Parse(string(match[1]))
		if err != nil {
			continue
		}
		target := base.ResolveReference(ref)
		if !strings.EqualFold(target.Hostname(), base.Hostname()) {
			continue
		}
		observation, data := doObservedRequest(ctx, client, http.MethodGet, target.String(), nil, min64(remaining, 2<<20))
		if observation.Error != "" || observation.Status != http.StatusOK || len(data) == 0 {
			continue
		}
		remaining -= int64(len(data))
		name := fmt.Sprintf("bundle-%02d-%s.js", index+1, HashBytes([]byte(target.String()))[:12])
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			manifest.Warnings = append(manifest.Warnings, err.Error())
			continue
		}
		result = append(result, target.String())
	}
	return result
}

func inferStack(source string, observations []HTTPObservation, body []byte, bundles []string) []Observation {
	text := strings.ToLower(string(body) + " " + strings.Join(bundles, " "))
	for _, observation := range observations {
		for key, values := range observation.Headers {
			text += " " + strings.ToLower(key+" "+strings.Join(values, " "))
		}
	}
	type signal struct{ needle, category, value string }
	signals := []signal{
		{"supabase.co", "backend", "Supabase"}, {"firebase", "backend", "Firebase"},
		{"clerk", "auth_provider", "Clerk"}, {"auth0", "auth_provider", "Auth0"},
		{"next/static", "framework", "Next.js"}, {"__next_data__", "framework", "Next.js"},
		{"react", "framework", "React"}, {"vue", "framework", "Vue"},
		{"vercel", "hosting", "Vercel"}, {"netlify", "hosting", "Netlify"},
		{"cloudflare", "hosting", "Cloudflare"},
	}
	seen := make(map[string]struct{})
	result := []Observation{}
	for _, item := range signals {
		if strings.Contains(text, item.needle) {
			key := item.category + ":" + item.value
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, Observation{Category: item.category, Value: item.value, SourceType: "direct_observation", SourceURL: source, ObservedAt: now(), Confidence: "medium", Limitations: "Fingerprint; requer confirmação do cliente."})
		}
	}
	return result
}

func categorizeProfile(profile StackProfile) string {
	values := ""
	for _, observation := range profile.Observations {
		values += " " + strings.ToLower(observation.Value)
	}
	switch {
	case strings.Contains(values, "supabase") || strings.Contains(values, "firebase"):
		return "baas"
	case strings.Contains(values, "next.js") || strings.Contains(values, "react") || strings.Contains(values, "vue"):
		return "web-application"
	case profile.Category != "":
		return profile.Category
	default:
		return "unknown"
	}
}

func filterEmptyObservations(items []Observation) []Observation {
	result := items[:0]
	for _, item := range items {
		if strings.TrimSpace(item.Value) != "" {
			result = append(result, item)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Category == result[j].Category {
			return result[i].Value < result[j].Value
		}
		return result[i].Category < result[j].Category
	})
	return result
}

func writePipeline1Report(dir string, cfg LeadConfig, profile StackProfile, manifest Manifest) error {
	var builder strings.Builder
	builder.WriteString("# Relatório de prospecção — " + cfg.CaseID + "\n\n")
	builder.WriteString("Domínio: `" + cfg.Domain + "`  \nModo: `" + manifest.Mode + "`  \nCategoria: `" + profile.Category + "`\n\n")
	builder.WriteString("## Evidências de stack\n\n")
	if len(profile.Observations) == 0 {
		builder.WriteString("Nenhum sinal técnico consolidado.\n")
	}
	for _, observation := range profile.Observations {
		builder.WriteString(fmt.Sprintf("- %s: **%s** (%s; %s)\n", observation.Category, observation.Value, observation.Confidence, observation.SourceType))
	}
	builder.WriteString("\n## Limitação\n\nResultado de observação da superfície pública e de baixo impacto, não reconhecimento passivo estrito. Candidatos precisam de revisão humana. Chaves anônimas de BaaS e sinais de takeover não são achados por si sós.\n")
	return os.WriteFile(filepath.Join(dir, "prospect-report.md"), []byte(builder.String()), 0o600)
}

func mode(execute bool) string {
	if execute {
		return "execute"
	}
	return "plan"
}
func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
