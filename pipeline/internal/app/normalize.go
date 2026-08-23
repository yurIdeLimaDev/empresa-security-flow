package app

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type NormalizeConfig struct {
	Tool              string
	InputPath         string
	EngagementID      string
	ScopeID           string
	ToolRunID         string
	ToolVersion       string
	ToolDigest        string
	PolicySHA256      string
	CommandSHA256     string
	SupplyChainStatus string
	ObservedAt        string
}

var supportedNormalizers = map[string]struct{}{
	"subfinder": {}, "amass": {}, "httpx": {}, "wappalyzer": {}, "wappalyzergo": {},
	"gitleaks": {}, "trufflehog": {}, "testssl": {}, "testssl.sh": {}, "nuclei": {},
	"zap": {}, "schemathesis": {}, "jwt_tool": {}, "jwt-tool": {}, "sqlmap": {},
	"dalfox": {}, "osv-scanner": {}, "trivy": {}, "semgrep": {},
	"jsluice": {}, "hadrian": {}, "dnsreaper": {}, "hibp": {},
}

func NormalizeFile(cfg NormalizeConfig) (NormalizedBundle, error) {
	tool := strings.ToLower(strings.TrimSpace(cfg.Tool))
	if _, ok := supportedNormalizers[tool]; !ok {
		return NormalizedBundle{}, fmt.Errorf("normalizador não suportado: %s", cfg.Tool)
	}
	if cfg.EngagementID == "" || cfg.ScopeID == "" || cfg.InputPath == "" {
		return NormalizedBundle{}, fmt.Errorf("engagement, scope e input são obrigatórios")
	}
	data, err := os.ReadFile(cfg.InputPath)
	if err != nil {
		return NormalizedBundle{}, fmt.Errorf("ler saída de %s: %w", tool, err)
	}
	absPath, _ := filepath.Abs(cfg.InputPath)
	stamp := cfg.ObservedAt
	if stamp == "" {
		if info, statErr := os.Stat(cfg.InputPath); statErr == nil {
			stamp = info.ModTime().UTC().Format(time.RFC3339)
		} else {
			stamp = now()
		}
	}
	fileHash := HashBytes(data)
	runID := cfg.ToolRunID
	if runID == "" {
		runID = stableID("toolrun", cfg.EngagementID+"|"+tool+"|"+fileHash)
	}
	bundle := newNormalizedBundle(cfg.EngagementID, cfg.ScopeID)
	bundle.Engagement.Status = "normalized"
	bundle.ToolRuns = append(bundle.ToolRuns, ToolRun{
		ID: runID, Tool: tool, Version: cfg.ToolVersion, Status: "normalized", InputPath: absPath,
		InputSHA256: fileHash, Parser: "empresa-security/" + tool, ParserVersion: CanonicalSchemaVersion,
		ToolDigest: cfg.ToolDigest, PolicySHA256: cfg.PolicySHA256, CommandSHA256: cfg.CommandSHA256,
		SupplyChainStatus: defaultString(cfg.SupplyChainStatus, "unknown"), FinishedAt: stamp,
	})

	var parseErr error
	if looksLikeSARIF(data) {
		parseErr = parseSARIF(&bundle, data, absPath, fileHash, runID, stamp, tool)
	} else {
		switch tool {
		case "subfinder":
			parseErr = parseSubfinder(&bundle, data, runID, stamp)
		case "amass":
			parseErr = parseAmass(&bundle, data, runID, stamp)
		case "httpx":
			parseErr = parseHTTPX(&bundle, data, absPath, fileHash, runID, stamp)
		case "wappalyzer", "wappalyzergo":
			parseErr = parseWappalyzer(&bundle, data, runID, stamp)
		case "gitleaks":
			parseErr = parseGitleaks(&bundle, data, absPath, fileHash, runID, stamp)
		case "trufflehog":
			parseErr = parseTruffleHog(&bundle, data, absPath, fileHash, runID, stamp)
		case "testssl", "testssl.sh":
			parseErr = parseTestSSL(&bundle, data, absPath, fileHash, runID, stamp)
		case "nuclei":
			parseErr = parseNuclei(&bundle, data, absPath, fileHash, runID, stamp)
		case "zap":
			parseErr = parseZAP(&bundle, data, absPath, fileHash, runID, stamp)
		case "schemathesis":
			parseErr = parseSchemathesis(&bundle, data, absPath, fileHash, runID, stamp)
		case "jwt_tool", "jwt-tool":
			bundle.Exceptions = append(bundle.Exceptions, newException(runID, "normalization", "unsupported_stable_machine_format", "jwt_tool não oferece um contrato estável de relatório; a saída foi preservada e não virou achado automaticamente", false, stamp))
		case "sqlmap":
			parseErr = parseSQLMap(&bundle, data, absPath, fileHash, runID, stamp)
		case "dalfox":
			parseErr = parseDalfox(&bundle, data, absPath, fileHash, runID, stamp)
		case "osv-scanner":
			parseErr = parseOSV(&bundle, data, absPath, fileHash, runID, stamp)
		case "trivy":
			parseErr = parseTrivy(&bundle, data, absPath, fileHash, runID, stamp)
		case "semgrep":
			parseErr = parseSemgrep(&bundle, data, absPath, fileHash, runID, stamp)
		case "jsluice":
			parseErr = parseJSLuice(&bundle, data, absPath, fileHash, runID, stamp)
		case "hadrian":
			parseErr = parseHadrian(&bundle, data, absPath, fileHash, runID, stamp)
		case "dnsreaper":
			parseErr = parseDNSReaper(&bundle, data, absPath, fileHash, runID, stamp)
		case "hibp":
			parseErr = parseHIBP(&bundle, data, absPath, fileHash, runID, stamp)
		}
	}
	if parseErr != nil {
		bundle.ToolRuns[0].Status = "failed"
		bundle.Exceptions = append(bundle.Exceptions, newException(runID, "normalization", "parse_error", parseErr.Error(), true, stamp))
		finalizeBundle(&bundle)
		return bundle, fmt.Errorf("normalizar %s: %w", tool, parseErr)
	}
	finalizeBundle(&bundle)
	return bundle, ValidateBundle(bundle)
}

func parseDNSReaper(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	var rows []DNSReaperCandidate
	if err := json.Unmarshal(data, &rows); err != nil {
		return err
	}
	for index, row := range rows {
		assetID := addAsset(bundle, "hostname", row.Domain, runID, stamp, nil)
		evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("/%d", index), stamp, false)
		confidence := "medium"
		if strings.EqualFold(row.Confidence, "UNLIKELY") {
			confidence = "low"
		}
		addFinding(bundle, Finding{Title: "Possível subdomain takeover", Description: row.Info, Category: "subdomain-takeover-candidate", Severity: "medium", Confidence: confidence, Status: "candidate", RuleID: row.Signature, AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: row.Domain, Exploitability: "not-verified", Impact: "Candidato apenas; nenhum recurso foi reivindicado.", FirstObserved: stamp, CanonicalKey: "dnsreaper|" + strings.ToLower(row.Domain) + "|" + strings.ToLower(row.Signature)})
	}
	return nil
}

func parseHIBP(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	var rows []HIBPResult
	if err := json.Unmarshal(data, &rows); err != nil {
		return err
	}
	for index, row := range rows {
		for _, breach := range row.Breaches {
			evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("/%d", index), stamp, true)
			addFinding(bundle, Finding{Title: "E-mail corporativo aparece em violação conhecida", Description: "O identificador do e-mail foi pseudonimizado no artefato canônico; a ocorrência precisa de revisão antes de qualquer contato.", Category: "breach-exposure", Severity: "medium", Confidence: "high", Status: "candidate", RuleID: breach, EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: "email-sha256:" + row.AccountSHA256, Exposure: "public-email-to-hibp", Exploitability: "not-applicable", FirstObserved: stamp, CanonicalKey: "hibp|" + row.AccountSHA256 + "|" + strings.ToLower(breach)})
		}
	}
	return nil
}

func newException(runID, stage, kind, reason string, blocking bool, stamp string) Exception {
	id := stableID("exception", runID+"|"+stage+"|"+kind+"|"+reason)
	return Exception{ID: id, ToolRunID: runID, Stage: stage, Type: kind, Reason: reason, Blocking: blocking, CreatedAt: stamp}
}

func parseJSONRecords(data []byte) ([]map[string]any, error) {
	var root any
	if err := json.Unmarshal(data, &root); err == nil {
		switch value := root.(type) {
		case []any:
			return mapsFrom(value), nil
		case map[string]any:
			return []map[string]any{value}, nil
		}
	}
	result := []map[string]any{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 16<<20)
	line := 0
	for scanner.Scan() {
		line++
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return nil, fmt.Errorf("JSONL linha %d: %w", line, err)
		}
		result = append(result, row)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("saída não contém JSON/JSONL reconhecível")
	}
	return result, nil
}

func mapsFrom(values []any) []map[string]any {
	result := []map[string]any{}
	for _, value := range values {
		if row, ok := value.(map[string]any); ok {
			result = append(result, row)
		}
	}
	return result
}

func mapAt(row map[string]any, key string) map[string]any {
	if value, ok := row[key].(map[string]any); ok {
		return value
	}
	return map[string]any{}
}

func arrayAt(row map[string]any, key string) []any {
	if value, ok := row[key].([]any); ok {
		return value
	}
	return nil
}

func stringAt(row map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := row[key]; ok && value != nil {
			switch typed := value.(type) {
			case string:
				if strings.TrimSpace(typed) != "" {
					return typed
				}
			case json.Number:
				return typed.String()
			case float64:
				return strconv.FormatFloat(typed, 'f', -1, 64)
			}
		}
	}
	return ""
}

func intAt(row map[string]any, keys ...string) int {
	value := stringAt(row, keys...)
	parsed, _ := strconv.Atoi(strings.Split(value, ".")[0])
	return parsed
}

func boolAt(row map[string]any, keys ...string) bool {
	for _, key := range keys {
		if value, ok := row[key].(bool); ok {
			return value
		}
	}
	return false
}

func stringSlice(value any) []string {
	result := []string{}
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
	case []string:
		result = append(result, typed...)
	case string:
		for _, item := range strings.Split(typed, ",") {
			result = append(result, strings.TrimSpace(item))
		}
	}
	return dedupeSortedCase(result)
}

func parseSubfinder(bundle *NormalizedBundle, data []byte, runID, stamp string) error {
	rows, err := parseJSONRecords(data)
	if err != nil {
		for _, line := range strings.Fields(string(data)) {
			addAsset(bundle, "domain", line, runID, stamp, nil)
		}
		if len(bundle.Assets) == 0 {
			return err
		}
		return nil
	}
	for _, row := range rows {
		addAsset(bundle, "domain", stringAt(row, "host", "name", "input"), runID, stamp, nil)
	}
	return nil
}

func parseAmass(bundle *NormalizedBundle, data []byte, runID, stamp string) error {
	rows, err := parseJSONRecords(data)
	if err != nil {
		for _, line := range strings.Fields(string(data)) {
			addAsset(bundle, "domain", line, runID, stamp, nil)
		}
		if len(bundle.Assets) == 0 {
			return err
		}
		return nil
	}
	for _, row := range rows {
		addAsset(bundle, "domain", stringAt(row, "name", "domain"), runID, stamp, nil)
		for _, addr := range arrayAt(row, "addresses") {
			if item, ok := addr.(map[string]any); ok {
				addAsset(bundle, "ip", stringAt(item, "ip", "address"), runID, stamp, nil)
			}
		}
	}
	return nil
}

func parseHTTPX(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	rows, err := parseJSONRecords(data)
	if err != nil {
		return err
	}
	for index, row := range rows {
		target := redactURL(stringAt(row, "url", "input"))
		tech := stringSlice(row["tech"])
		addAsset(bundle, "url", target, runID, stamp, tech)
		method := strings.ToUpper(defaultString(stringAt(row, "method"), "GET"))
		status := intAt(row, "status_code", "status-code")
		requestHash := HashBytes([]byte(method + " " + target))
		bundle.RequestLogs = append(bundle.RequestLogs, RequestLog{ID: stableID("request", runID+"|"+strconv.Itoa(index)+"|"+requestHash), ToolRunID: runID, ObservedAt: stamp, Method: method, URL: target, Status: status, RequestSHA256: requestHash, ScopeDecision: "imported_unverified"})
		addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("jsonl:%d", index+1), stamp, false)
	}
	return nil
}

func parseJSLuice(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	rows, err := parseJSONRecords(data)
	if err != nil {
		return err
	}
	for index, row := range rows {
		target := redactURL(stringAt(row, "url"))
		if target == "" {
			continue
		}
		addAsset(bundle, "url", target, runID, stamp, nil)
		method := strings.ToUpper(defaultString(stringAt(row, "method"), "GET"))
		requestHash := HashBytes([]byte(method + " " + target))
		bundle.RequestLogs = append(bundle.RequestLogs, RequestLog{
			ID: stableID("request", runID+"|"+strconv.Itoa(index)+"|"+requestHash), ToolRunID: runID,
			ObservedAt: stamp, Method: method, URL: target, RequestSHA256: requestHash, ScopeDecision: "static_bundle_candidate",
		})
		addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("jsonl:%d", index+1), stamp, false)
	}
	return nil
}

// parseHadrian consumes the stable JSON report emitted by Hadrian v1.0.0.
// The released v1.0.0 does not expose SARIF, so the pinned release contract is
// deliberately used instead of relying on unreleased code from the main branch.
func parseHadrian(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	var report struct {
		Metadata struct {
			Tool    string `json:"tool"`
			Version string `json:"version"`
		} `json:"metadata"`
		Findings []struct {
			ID              string         `json:"id"`
			Category        string         `json:"category"`
			Name            string         `json:"name"`
			Description     string         `json:"description"`
			Severity        string         `json:"severity"`
			Confidence      float64        `json:"confidence"`
			IsVulnerability bool           `json:"is_vulnerability"`
			Endpoint        string         `json:"endpoint"`
			Method          string         `json:"method"`
			AttackerRole    string         `json:"attacker_role"`
			VictimRole      string         `json:"victim_role"`
			RequestIDs      []string       `json:"request_ids"`
			Evidence        map[string]any `json:"evidence"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("esperado relatório JSON do Hadrian v1.0.0: %w", err)
	}
	if !strings.EqualFold(report.Metadata.Tool, "hadrian") {
		return fmt.Errorf("metadata.tool do relatório Hadrian é inválido")
	}
	if report.Metadata.Version != "" {
		bundle.ToolRuns[0].Version = report.Metadata.Version
	}
	for index, item := range report.Findings {
		request := mapAt(item.Evidence, "request")
		response := mapAt(item.Evidence, "response")
		setupResponse := mapAt(item.Evidence, "setup_response")
		attackResponse := mapAt(item.Evidence, "attack_response")
		verifyResponse := mapAt(item.Evidence, "verify_response")
		target := redactURL(stringAt(request, "url"))
		if target == "" {
			target = item.Endpoint
		}
		method := strings.ToUpper(defaultString(item.Method, stringAt(request, "method")))
		assetID := ""
		if strings.Contains(target, "://") {
			assetID = addAsset(bundle, "url", target, runID, stamp, nil)
		}
		requestHash := HashBytes([]byte(method + " " + target + "|" + stringAt(request, "body")))
		responseHash := stringAt(response, "body_hash")
		if responseHash == "" && stringAt(response, "body") != "" {
			responseHash = HashBytes([]byte(stringAt(response, "body")))
		}
		bundle.RequestLogs = append(bundle.RequestLogs, RequestLog{
			ID: stableID("request", runID+"|"+strconv.Itoa(index)+"|"+requestHash), ToolRunID: runID,
			ObservedAt: stamp, Method: method, URL: target, Status: intAt(response, "status_code"),
			RequestSHA256: requestHash, ResponseSHA256: responseHash, ScopeDecision: "hadrian_identity_comparison",
		})
		evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("findings[%d]:%s", index, item.ID), stamp, true)
		if !item.IsVulnerability {
			continue
		}
		confidence := "low"
		if item.Confidence >= 0.8 {
			confidence = "high"
		} else if item.Confidence >= 0.5 {
			confidence = "medium"
		}
		status, exploitability := "candidate", "identity-compared"
		// Three-phase mutation evidence proves the action effect only when the
		// released report contains setup, attack and verification responses.
		if len(setupResponse) > 0 && len(attackResponse) > 0 && len(verifyResponse) > 0 {
			status, exploitability = "validated", "effect-verified"
		}
		roles := strings.TrimSpace(item.AttackerRole)
		if item.VictimRole != "" {
			roles += " -> " + item.VictimRole
		}
		addFinding(bundle, Finding{
			Title: defaultString(item.Name, "Falha de autorização indicada pelo Hadrian"), Description: item.Description,
			Category: item.Category + " authorization", Severity: item.Severity, Confidence: confidence,
			Status: status, RuleID: item.Category, Identifiers: item.RequestIDs, AssetIDs: nonEmpty(assetID),
			EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: target,
			Parameter: roles, Exploitability: exploitability, FirstObserved: stamp,
			CanonicalKey: "hadrian|" + item.Category + "|" + method + "|" + normalizedLocation(target) + "|" + roles,
		})
	}
	return nil
}

func parseWappalyzer(bundle *NormalizedBundle, data []byte, runID, stamp string) error {
	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	add := func(target string, value any) {
		tech := []string{}
		switch items := value.(type) {
		case []any:
			for _, item := range items {
				if row, ok := item.(map[string]any); ok {
					tech = append(tech, stringAt(row, "name", "technology"))
				} else if text, ok := item.(string); ok {
					tech = append(tech, text)
				}
			}
		case map[string]any:
			for name := range items {
				tech = append(tech, name)
			}
		}
		addAsset(bundle, "url", target, runID, stamp, tech)
	}
	switch typed := root.(type) {
	case map[string]any:
		if target := stringAt(typed, "url", "target"); target != "" {
			add(target, typed["technologies"])
		} else {
			for target, value := range typed {
				add(target, value)
			}
		}
	case []any:
		for _, item := range typed {
			if row, ok := item.(map[string]any); ok {
				add(stringAt(row, "url", "target"), row["technologies"])
			}
		}
	}
	return nil
}

func parseGitleaks(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	rows, err := parseJSONRecords(data)
	if err != nil {
		return err
	}
	for index, row := range rows {
		file := stringAt(row, "File", "file")
		line := intAt(row, "StartLine", "start_line", "line")
		locator := fmt.Sprintf("json:%d:%s:%d", index, normalizedLocation(file), line)
		evidenceID := addEvidence(bundle, runID, path, fileHash, locator, stamp, true)
		rule := stringAt(row, "RuleID", "rule_id")
		addFinding(bundle, Finding{Title: defaultString(stringAt(row, "Description", "description"), "Possível segredo exposto"), Category: "secret-exposure", Severity: "high", Confidence: "medium", RuleID: rule, Location: fmt.Sprintf("%s:%d", file, line), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, FirstObserved: stamp, CanonicalKey: "secret|" + normalizedLocation(file) + "|" + strconv.Itoa(line)})
	}
	return nil
}

func parseTruffleHog(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	rows, err := parseJSONRecords(data)
	if err != nil {
		return err
	}
	for index, row := range rows {
		meta := mapAt(mapAt(mapAt(row, "SourceMetadata"), "Data"), "Filesystem")
		file := stringAt(meta, "file", "path")
		line := intAt(meta, "line")
		if file == "" {
			git := mapAt(mapAt(mapAt(row, "SourceMetadata"), "Data"), "Git")
			file, line = stringAt(git, "file"), intAt(git, "line")
		}
		detector := stringAt(row, "DetectorName", "DetectorType", "detector_name")
		evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("jsonl:%d:%s:%d", index+1, normalizedLocation(file), line), stamp, true)
		confidence := "medium"
		exploitability := "candidate"
		if boolAt(row, "Verified", "verified") {
			confidence, exploitability = "high", "tool-verified"
		}
		addFinding(bundle, Finding{Title: defaultString(detector, "Possível segredo exposto"), Category: "secret-exposure", Severity: "high", Confidence: confidence, RuleID: detector, Location: fmt.Sprintf("%s:%d", file, line), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Exploitability: exploitability, FirstObserved: stamp, CanonicalKey: "secret|" + normalizedLocation(file) + "|" + strconv.Itoa(line)})
	}
	return nil
}

func flattenTestSSL(value any, result *[]map[string]any) {
	switch typed := value.(type) {
	case map[string]any:
		if stringAt(typed, "id") != "" && (stringAt(typed, "severity") != "" || stringAt(typed, "finding") != "") {
			*result = append(*result, typed)
		}
		for _, child := range typed {
			flattenTestSSL(child, result)
		}
	case []any:
		for _, child := range typed {
			flattenTestSSL(child, result)
		}
	}
}

func parseTestSSL(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	rows := []map[string]any{}
	flattenTestSSL(root, &rows)
	for index, row := range rows {
		severity := stringAt(row, "severity")
		if normalizeSeverity(severity) == "info" {
			continue
		}
		id, finding := stringAt(row, "id"), stringAt(row, "finding")
		target := stringAt(row, "ip", "fqdn", "targetHost")
		assetID := addAsset(bundle, "host", target, runID, stamp, nil)
		evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("json:%d:%s", index, id), stamp, false)
		addFinding(bundle, Finding{Title: defaultString(id, "Configuração TLS"), Description: finding, Category: "tls-misconfiguration", Severity: severity, Confidence: "medium", RuleID: id, AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: target, FirstObserved: stamp})
	}
	return nil
}

func parseNuclei(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	rows, err := parseJSONRecords(data)
	if err != nil {
		return err
	}
	for index, row := range rows {
		info := mapAt(row, "info")
		target := redactURL(stringAt(row, "matched-at", "matched", "url", "host"))
		assetID := addAsset(bundle, "url", target, runID, stamp, nil)
		rule := stringAt(row, "template-id", "templateID")
		evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("jsonl:%d:%s", index+1, rule), stamp, false)
		addFinding(bundle, Finding{Title: defaultString(stringAt(info, "name"), rule), Description: stringAt(info, "description"), Category: strings.Join(stringSlice(info["tags"]), ","), Severity: stringAt(info, "severity"), Confidence: "medium", RuleID: rule, AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: target, Parameter: stringAt(row, "matcher-name"), FirstObserved: stamp})
	}
	return nil
}

func parseZAP(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	for siteIndex, siteValue := range arrayAt(root, "site") {
		site, _ := siteValue.(map[string]any)
		for alertIndex, alertValue := range arrayAt(site, "alerts") {
			alert, _ := alertValue.(map[string]any)
			instances := arrayAt(alert, "instances")
			if len(instances) == 0 {
				instances = []any{map[string]any{"uri": stringAt(site, "@name")}}
			}
			for instanceIndex, instanceValue := range instances {
				instance, _ := instanceValue.(map[string]any)
				target, parameter := redactURL(stringAt(instance, "uri", "url")), stringAt(instance, "param", "parameter")
				assetID := addAsset(bundle, "url", target, runID, stamp, nil)
				rule := stringAt(alert, "pluginid", "alertRef")
				evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("site:%d/alert:%d/instance:%d", siteIndex, alertIndex, instanceIndex), stamp, false)
				addFinding(bundle, Finding{Title: defaultString(stringAt(alert, "name", "alert"), rule), Description: stringAt(alert, "desc"), Severity: stringAt(alert, "riskdesc", "riskcode"), Confidence: stringAt(alert, "confidence"), RuleID: rule, AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: target, Parameter: parameter, Remediation: stringAt(alert, "solution"), FirstObserved: stamp})
			}
		}
	}
	return nil
}

type junitSuite struct {
	TestCases []junitCase  `xml:"testcase"`
	Suites    []junitSuite `xml:"testsuite"`
}
type junitCase struct {
	Name    string        `xml:"name,attr"`
	Class   string        `xml:"classname,attr"`
	Failure *junitFailure `xml:"failure"`
	Error   *junitFailure `xml:"error"`
}
type junitFailure struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

func parseSchemathesis(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	var suite junitSuite
	if err := xml.Unmarshal(data, &suite); err != nil {
		return fmt.Errorf("esperado JUnit XML ou SARIF: %w", err)
	}
	var walk func(junitSuite)
	walk = func(current junitSuite) {
		for index, test := range current.TestCases {
			failure := test.Failure
			if failure == nil {
				failure = test.Error
			}
			if failure == nil {
				continue
			}
			evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("junit:%s:%d", test.Name, index), stamp, false)
			addFinding(bundle, Finding{Title: defaultString(failure.Message, test.Name), Description: strings.TrimSpace(failure.Text), Category: "api-contract", Severity: "medium", Confidence: "medium", RuleID: test.Class, Location: test.Name, EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, FirstObserved: stamp})
		}
		for _, child := range current.Suites {
			walk(child)
		}
	}
	walk(suite)
	return nil
}

func parseSQLMap(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	if rows, err := parseJSONRecords(data); err == nil {
		for index, row := range rows {
			target, parameter := redactURL(stringAt(row, "url", "target", "place")), stringAt(row, "parameter", "param")
			if parameter == "" && stringAt(row, "title", "payload", "technique") == "" {
				continue
			}
			assetID := addAsset(bundle, "url", target, runID, stamp, nil)
			evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("json:%d", index), stamp, true)
			addFinding(bundle, Finding{Title: defaultString(stringAt(row, "title", "technique"), "Possível injeção SQL"), Description: "Saída do sqlmap importada; payloads não são copiados para o modelo canônico.", Category: "sql-injection", Severity: "high", Confidence: "high", RuleID: stringAt(row, "dbms"), AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: target, Parameter: parameter, Exploitability: "tool-verified", FirstObserved: stamp})
		}
		return nil
	}
	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("esperado --report-json ou CSV: %w", err)
	}
	if len(records) < 2 {
		return nil
	}
	headers := map[string]int{}
	for i, value := range records[0] {
		headers[strings.ToLower(value)] = i
	}
	for index, record := range records[1:] {
		get := func(name string) string {
			if position, ok := headers[name]; ok && position < len(record) {
				return record[position]
			}
			return ""
		}
		target, parameter := get("target url"), get("parameter")
		assetID := addAsset(bundle, "url", target, runID, stamp, nil)
		evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("csv:%d", index+2), stamp, true)
		addFinding(bundle, Finding{Title: "Possível injeção SQL", Category: "sql-injection", Severity: "high", Confidence: "high", AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: target, Parameter: parameter, Exploitability: "tool-verified", FirstObserved: stamp})
	}
	return nil
}

func parseDalfox(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	rows, err := parseJSONRecords(data)
	if err != nil {
		return err
	}
	for index, row := range rows {
		target := redactURL(stringAt(row, "url", "target"))
		if target == "" {
			target = safeURLFromPOC(stringAt(row, "poc"))
		}
		assetID := addAsset(bundle, "url", target, runID, stamp, nil)
		evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("jsonl:%d", index+1), stamp, false)
		addFinding(bundle, Finding{Title: defaultString(stringAt(row, "type"), "Possível XSS"), Category: "xss", Severity: "high", Confidence: "high", RuleID: stringAt(row, "type"), AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: target, Parameter: stringAt(row, "param", "parameter"), Exploitability: "tool-verified", FirstObserved: stamp})
	}
	return nil
}

func parseOSV(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	for resultIndex, resultValue := range arrayAt(root, "results") {
		result, _ := resultValue.(map[string]any)
		for packageIndex, packageValue := range arrayAt(result, "packages") {
			pkg, _ := packageValue.(map[string]any)
			info := mapAt(pkg, "package")
			name, ecosystem, version := stringAt(info, "name"), canonicalEcosystem(stringAt(info, "ecosystem")), stringAt(info, "version")
			if version == "" {
				version = stringAt(pkg, "version")
			}
			assetID := addAsset(bundle, "package", strings.ToLower(ecosystem)+":"+name+"@"+version, runID, stamp, nil)
			for vulnIndex, vulnValue := range arrayAt(pkg, "vulnerabilities") {
				vuln, _ := vulnValue.(map[string]any)
				id := stringAt(vuln, "id")
				evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("results:%d/packages:%d/vulnerabilities:%d", resultIndex, packageIndex, vulnIndex), stamp, false)
				addFinding(bundle, Finding{Title: defaultString(stringAt(vuln, "summary"), id), Description: stringAt(vuln, "details"), Category: "dependency-vulnerability", Severity: osvSeverity(vuln), Confidence: "high", RuleID: id, Identifiers: append([]string{id}, stringSlice(vuln["aliases"])...), AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: name + "@" + version, FirstObserved: stamp, CanonicalKey: "dependency|" + strings.ToLower(ecosystem) + ":" + strings.ToLower(name) + "@" + version + "|" + id})
			}
		}
	}
	return nil
}

func parseTrivy(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	for resultIndex, resultValue := range arrayAt(root, "Results") {
		result, _ := resultValue.(map[string]any)
		target := stringAt(result, "Target")
		for vulnIndex, vulnValue := range arrayAt(result, "Vulnerabilities") {
			vuln, _ := vulnValue.(map[string]any)
			id, name, version := stringAt(vuln, "VulnerabilityID"), stringAt(vuln, "PkgName"), stringAt(vuln, "InstalledVersion")
			ecosystem := canonicalEcosystem(stringAt(result, "Type", "Class"))
			assetID := addAsset(bundle, "package", ecosystem+":"+name+"@"+version, runID, stamp, nil)
			evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("Results:%d/Vulnerabilities:%d", resultIndex, vulnIndex), stamp, false)
			addFinding(bundle, Finding{Title: defaultString(stringAt(vuln, "Title"), id), Description: stringAt(vuln, "Description"), Category: "dependency-vulnerability", Severity: stringAt(vuln, "Severity"), Confidence: "high", RuleID: id, Identifiers: []string{id}, AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: target + ":" + name + "@" + version, Remediation: stringAt(vuln, "FixedVersion"), FirstObserved: stamp, CanonicalKey: "dependency|" + ecosystem + ":" + strings.ToLower(name) + "@" + version + "|" + id})
		}
		for misIndex, misValue := range arrayAt(result, "Misconfigurations") {
			mis, _ := misValue.(map[string]any)
			id := stringAt(mis, "ID", "AVDID")
			assetID := addAsset(bundle, "file", target, runID, stamp, nil)
			evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("Results:%d/Misconfigurations:%d", resultIndex, misIndex), stamp, false)
			addFinding(bundle, Finding{Title: defaultString(stringAt(mis, "Title"), id), Description: stringAt(mis, "Description"), Category: "misconfiguration", Severity: stringAt(mis, "Severity"), Confidence: "medium", RuleID: id, AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: target, Remediation: stringAt(mis, "Resolution"), FirstObserved: stamp})
		}
	}
	return nil
}

func parseSemgrep(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp string) error {
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	for index, resultValue := range arrayAt(root, "results") {
		row, _ := resultValue.(map[string]any)
		extra := mapAt(row, "extra")
		start := mapAt(row, "start")
		file, line, rule := stringAt(row, "path"), intAt(start, "line"), stringAt(row, "check_id")
		assetID := addAsset(bundle, "file", file, runID, stamp, nil)
		evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("results:%d:%s:%d", index, normalizedLocation(file), line), stamp, false)
		addFinding(bundle, Finding{Title: defaultString(stringAt(extra, "message"), rule), Category: "static-analysis", Severity: stringAt(extra, "severity"), Confidence: "medium", RuleID: rule, AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: fmt.Sprintf("%s:%d", file, line), FirstObserved: stamp})
	}
	return nil
}

func looksLikeSARIF(data []byte) bool {
	var root map[string]any
	return json.Unmarshal(data, &root) == nil && strings.Contains(strings.ToLower(stringAt(root, "version")), "2.1") && arrayAt(root, "runs") != nil
}

func parseSARIF(bundle *NormalizedBundle, data []byte, path, fileHash, runID, stamp, tool string) error {
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	for runIndex, runValue := range arrayAt(root, "runs") {
		run, _ := runValue.(map[string]any)
		for resultIndex, resultValue := range arrayAt(run, "results") {
			row, _ := resultValue.(map[string]any)
			message := mapAt(row, "message")
			location, line := "", 0
			if locations := arrayAt(row, "locations"); len(locations) > 0 {
				loc, _ := locations[0].(map[string]any)
				physical := mapAt(loc, "physicalLocation")
				artifact := mapAt(physical, "artifactLocation")
				region := mapAt(physical, "region")
				location, line = stringAt(artifact, "uri"), intAt(region, "startLine")
			}
			assetID := addAsset(bundle, choose(strings.Contains(location, "://"), "url", "file"), location, runID, stamp, nil)
			evidenceID := addEvidence(bundle, runID, path, fileHash, fmt.Sprintf("runs:%d/results:%d", runIndex, resultIndex), stamp, false)
			rule := stringAt(row, "ruleId")
			addFinding(bundle, Finding{Title: defaultString(stringAt(message, "text", "markdown"), rule), Category: tool + " " + rule, Severity: stringAt(row, "level"), Confidence: "medium", RuleID: rule, AssetIDs: nonEmpty(assetID), EvidenceIDs: []string{evidenceID}, ToolRunIDs: []string{runID}, Location: fmt.Sprintf("%s:%d", location, line), FirstObserved: stamp})
		}
	}
	return nil
}

func nonEmpty(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}

func safeURLFromPOC(poc string) string {
	for _, field := range strings.Fields(poc) {
		candidate := strings.Trim(field, "'\"()[]{}")
		if parsed, err := url.Parse(candidate); err == nil && parsed.Scheme != "" && parsed.Hostname() != "" {
			parsed.RawQuery = ""
			parsed.Fragment = ""
			return parsed.String()
		}
	}
	return ""
}

func redactURL(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme == "" || parsed.Hostname() == "" {
		return value
	}
	query := parsed.Query()
	for key := range query {
		query.Set(key, "[redacted]")
	}
	parsed.RawQuery = query.Encode()
	parsed.Fragment = ""
	return parsed.String()
}

func canonicalEcosystem(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "gomod", "go":
		return "go"
	case "npm", "node-pkg", "yarn", "pnpm":
		return "npm"
	case "pip", "pipenv", "poetry", "python", "pypi":
		return "pypi"
	case "maven", "gradle", "jar":
		return "maven"
	case "nuget", "dotnet-core":
		return "nuget"
	case "cargo", "rust":
		return "cargo"
	case "bundler", "ruby", "gem":
		return "rubygems"
	default:
		return value
	}
}

func osvSeverity(vuln map[string]any) string {
	if value := stringAt(mapAt(vuln, "database_specific"), "severity"); value != "" {
		return value
	}
	for _, value := range arrayAt(vuln, "severity") {
		if row, ok := value.(map[string]any); ok {
			score := stringAt(row, "score")
			if strings.Contains(score, "/AV:N") {
				return "high"
			}
		}
	}
	return "medium"
}
