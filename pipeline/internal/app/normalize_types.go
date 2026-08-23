package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const CanonicalSchemaVersion = "1.0.0"

type Engagement struct {
	ID         string   `json:"id"`
	Name       string   `json:"name,omitempty"`
	Pipeline   int      `json:"pipeline,omitempty"`
	ScopeID    string   `json:"scope_id"`
	StartedAt  string   `json:"started_at,omitempty"`
	FinishedAt string   `json:"finished_at,omitempty"`
	Status     string   `json:"status"`
	ToolRunIDs []string `json:"tool_run_ids"`
}

type Scope struct {
	ID             string   `json:"id"`
	BaseURL        string   `json:"base_url,omitempty"`
	AllowedSchemes []string `json:"allowed_schemes,omitempty"`
	AllowedHosts   []string `json:"allowed_hosts,omitempty"`
	AllowedPorts   []int    `json:"allowed_ports,omitempty"`
	ExcludedPaths  []string `json:"excluded_paths,omitempty"`
	SOWSHA256      string   `json:"sow_sha256,omitempty"`
	Fingerprint    string   `json:"fingerprint"`
}

type Asset struct {
	ID               string   `json:"id"`
	Kind             string   `json:"kind"`
	Value            string   `json:"value"`
	URL              string   `json:"url,omitempty"`
	Host             string   `json:"host,omitempty"`
	IP               string   `json:"ip,omitempty"`
	Scheme           string   `json:"scheme,omitempty"`
	Port             int      `json:"port,omitempty"`
	Technologies     []string `json:"technologies,omitempty"`
	SourceToolRunIDs []string `json:"source_tool_run_ids"`
	FirstSeen        string   `json:"first_seen"`
	LastSeen         string   `json:"last_seen"`
}

type RankBreakdown struct {
	Severity       int `json:"severity"`
	Confidence     int `json:"confidence"`
	Evidence       int `json:"evidence"`
	Exposure       int `json:"exposure"`
	Exploitability int `json:"exploitability"`
	BusinessImpact int `json:"business_impact"`
	Total          int `json:"total"`
}

type CommercialValue struct {
	Relevance  string `json:"relevance"`
	ProofLevel string `json:"proof_level"`
	Statement  string `json:"statement"`
	Limitation string `json:"limitation"`
}

type Finding struct {
	ID             string          `json:"id"`
	Fingerprint    string          `json:"fingerprint"`
	Title          string          `json:"title"`
	Description    string          `json:"description,omitempty"`
	Category       string          `json:"category"`
	Severity       string          `json:"severity"`
	Confidence     string          `json:"confidence"`
	Status         string          `json:"status"`
	RuleID         string          `json:"rule_id,omitempty"`
	Identifiers    []string        `json:"identifiers,omitempty"`
	AssetIDs       []string        `json:"asset_ids,omitempty"`
	EvidenceIDs    []string        `json:"evidence_ids"`
	ToolRunIDs     []string        `json:"tool_run_ids"`
	Location       string          `json:"location,omitempty"`
	Parameter      string          `json:"parameter,omitempty"`
	Exposure       string          `json:"exposure,omitempty"`
	Exploitability string          `json:"exploitability,omitempty"`
	Impact         string          `json:"impact,omitempty"`
	Remediation    string          `json:"remediation,omitempty"`
	FirstObserved  string          `json:"first_observed"`
	LastObserved   string          `json:"last_observed"`
	Occurrences    int             `json:"occurrences"`
	Rank           RankBreakdown   `json:"rank"`
	Commercial     CommercialValue `json:"commercial_value"`
	CanonicalKey   string          `json:"-"`
}

type Evidence struct {
	ID              string `json:"id"`
	Kind            string `json:"kind"`
	Path            string `json:"path"`
	SHA256          string `json:"sha256"`
	Locator         string `json:"locator,omitempty"`
	CollectedAt     string `json:"collected_at"`
	ToolRunID       string `json:"tool_run_id"`
	IntegrityStatus string `json:"integrity_status"`
	ReviewStatus    string `json:"review_status"`
	Sensitive       bool   `json:"sensitive"`
}

type ToolRun struct {
	ID                string `json:"id"`
	Tool              string `json:"tool"`
	Version           string `json:"version,omitempty"`
	Status            string `json:"status"`
	StartedAt         string `json:"started_at,omitempty"`
	FinishedAt        string `json:"finished_at,omitempty"`
	InputPath         string `json:"input_path"`
	InputSHA256       string `json:"input_sha256"`
	Parser            string `json:"parser"`
	ParserVersion     string `json:"parser_version"`
	ToolDigest        string `json:"tool_digest,omitempty"`
	PolicySHA256      string `json:"policy_sha256,omitempty"`
	CommandSHA256     string `json:"command_sha256,omitempty"`
	SupplyChainStatus string `json:"supply_chain_status"`
}

type RequestLog struct {
	ID             string `json:"id"`
	ToolRunID      string `json:"tool_run_id"`
	ObservedAt     string `json:"observed_at"`
	Method         string `json:"method"`
	URL            string `json:"url"`
	Status         int    `json:"status,omitempty"`
	DurationMS     int64  `json:"duration_ms,omitempty"`
	RequestSHA256  string `json:"request_sha256,omitempty"`
	ResponseSHA256 string `json:"response_sha256,omitempty"`
	ScopeDecision  string `json:"scope_decision"`
}

type Exception struct {
	ID        string `json:"id"`
	ToolRunID string `json:"tool_run_id,omitempty"`
	Stage     string `json:"stage"`
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Blocking  bool   `json:"blocking"`
	CreatedAt string `json:"created_at"`
}

type BundleSummary struct {
	Assets             int            `json:"assets"`
	Findings           int            `json:"findings"`
	CandidateFindings  int            `json:"candidate_findings"`
	ValidatedFindings  int            `json:"validated_findings"`
	FindingsBySeverity map[string]int `json:"findings_by_severity"`
	Evidence           int            `json:"evidence"`
	ToolRuns           int            `json:"tool_runs"`
	Exceptions         int            `json:"exceptions"`
	EvidenceCoverage   float64        `json:"evidence_coverage"`
}

type NormalizedBundle struct {
	SchemaVersion string        `json:"schema_version"`
	GeneratedAt   string        `json:"generated_at"`
	Engagement    Engagement    `json:"engagement"`
	Scope         Scope         `json:"scope"`
	Assets        []Asset       `json:"assets"`
	Findings      []Finding     `json:"findings"`
	Evidence      []Evidence    `json:"evidence"`
	ToolRuns      []ToolRun     `json:"tool_runs"`
	RequestLogs   []RequestLog  `json:"request_logs"`
	Exceptions    []Exception   `json:"exceptions"`
	Summary       BundleSummary `json:"summary"`
}

type Comparison struct {
	SchemaVersion      string   `json:"schema_version"`
	GeneratedAt        string   `json:"generated_at"`
	PreviousEngagement string   `json:"previous_engagement"`
	CurrentEngagement  string   `json:"current_engagement"`
	Comparable         bool     `json:"comparable"`
	Reasons            []string `json:"reasons,omitempty"`
	New                []string `json:"new"`
	Recurring          []string `json:"recurring"`
	NotObserved        []string `json:"not_observed"`
	FixedVerified      []string `json:"fixed_verified"`
	ScoreDelta         int      `json:"score_delta"`
}

func newNormalizedBundle(engagementID, scopeID string) NormalizedBundle {
	stamp := now()
	return NormalizedBundle{
		SchemaVersion: CanonicalSchemaVersion,
		GeneratedAt:   stamp,
		Engagement:    Engagement{ID: engagementID, ScopeID: scopeID, Status: "in_progress"},
		Scope:         Scope{ID: scopeID},
		Assets:        []Asset{}, Findings: []Finding{}, Evidence: []Evidence{}, ToolRuns: []ToolRun{}, RequestLogs: []RequestLog{}, Exceptions: []Exception{},
	}
}

func stableID(prefix, value string) string {
	hash := sha256.Sum256([]byte(value))
	return prefix + "_" + hex.EncodeToString(hash[:12])
}

func canonicalAsset(kind, value string) (string, string, string, string, int) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	value = strings.TrimSpace(value)
	if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" && parsed.Hostname() != "" {
		parsed.Fragment = ""
		query := parsed.Query()
		for key := range query {
			query.Set(key, "[redacted]")
		}
		parsed.RawQuery = query.Encode()
		parsed.Scheme = strings.ToLower(parsed.Scheme)
		parsed.Host = strings.ToLower(parsed.Host)
		port := 0
		if parsed.Port() != "" {
			port, _ = strconv.Atoi(parsed.Port())
		} else if parsed.Scheme == "https" {
			port = 443
		} else if parsed.Scheme == "http" {
			port = 80
		}
		return parsed.String(), parsed.Hostname(), parsed.Scheme, "", port
	}
	if kind == "host" || kind == "domain" {
		return strings.ToLower(strings.TrimSuffix(value, ".")), strings.ToLower(strings.TrimSuffix(value, ".")), "", "", 0
	}
	if kind == "ip" {
		return value, "", "", value, 0
	}
	return strings.ToLower(value), "", "", "", 0
}

func addAsset(bundle *NormalizedBundle, kind, value, toolRunID, observedAt string, technologies []string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	canonical, host, scheme, ip, port := canonicalAsset(kind, value)
	id := stableID("asset", strings.ToLower(kind)+"|"+canonical)
	bundle.Assets = append(bundle.Assets, Asset{
		ID: id, Kind: strings.ToLower(kind), Value: canonical, URL: choose(strings.Contains(canonical, "://"), canonical, ""),
		Host: host, IP: ip, Scheme: scheme, Port: port, Technologies: dedupeSortedCase(technologies),
		SourceToolRunIDs: []string{toolRunID}, FirstSeen: observedAt, LastSeen: observedAt,
	})
	return id
}

func addEvidence(bundle *NormalizedBundle, toolRunID, path, fileHash, locator, collectedAt string, sensitive bool) string {
	key := fileHash + "|" + locator + "|" + toolRunID
	id := stableID("evidence", key)
	bundle.Evidence = append(bundle.Evidence, Evidence{
		ID: id, Kind: "tool-output", Path: path, SHA256: fileHash, Locator: locator,
		CollectedAt: collectedAt, ToolRunID: toolRunID, IntegrityStatus: "recorded_sha256",
		ReviewStatus: "unreviewed", Sensitive: sensitive,
	})
	return id
}

func addFinding(bundle *NormalizedBundle, finding Finding) {
	stamp := now()
	if finding.FirstObserved == "" {
		finding.FirstObserved = stamp
	}
	if finding.LastObserved == "" {
		finding.LastObserved = finding.FirstObserved
	}
	if finding.Occurrences <= 0 {
		finding.Occurrences = 1
	}
	if finding.Status == "" {
		finding.Status = "candidate"
	}
	finding.Severity = normalizeSeverity(finding.Severity)
	finding.Confidence = normalizeConfidence(finding.Confidence)
	finding.Category = normalizeCategory(finding.Category, finding.Title)
	finding.Identifiers = dedupeSortedCase(finding.Identifiers)
	finding.AssetIDs = dedupeSortedCase(finding.AssetIDs)
	finding.EvidenceIDs = dedupeSortedCase(finding.EvidenceIDs)
	finding.ToolRunIDs = dedupeSortedCase(finding.ToolRunIDs)
	finding.Fingerprint = findingFingerprint(finding)
	finding.ID = stableID("finding", finding.Fingerprint)
	bundle.Findings = append(bundle.Findings, finding)
}

func findingFingerprint(f Finding) string {
	if f.CanonicalKey != "" {
		return HashBytes([]byte(strings.ToLower(strings.TrimSpace(f.CanonicalKey))))
	}
	strong := ""
	for _, identifier := range f.Identifiers {
		upper := strings.ToUpper(identifier)
		if strings.HasPrefix(upper, "CVE-") || strings.HasPrefix(upper, "GHSA-") || strings.HasPrefix(upper, "OSV-") {
			strong = upper
			break
		}
	}
	location := normalizedLocation(f.Location)
	key := f.Category + "|" + strong + "|" + strings.Join(f.AssetIDs, ",") + "|" + location + "|" + strings.ToLower(f.Parameter)
	if strong == "" && location == "" {
		key += "|" + strings.ToLower(f.RuleID) + "|" + strings.ToLower(strings.TrimSpace(f.Title))
	}
	return HashBytes([]byte(key))
}

func normalizedLocation(value string) string {
	value = strings.TrimSpace(value)
	if parsed, err := url.Parse(value); err == nil && parsed.Hostname() != "" {
		parsed.RawQuery = ""
		parsed.Fragment = ""
		parsed.Host = strings.ToLower(parsed.Host)
		parsed.Scheme = strings.ToLower(parsed.Scheme)
		return parsed.String()
	}
	return strings.ToLower(strings.ReplaceAll(value, "\\", "/"))
}

func MergeBundles(bundles ...NormalizedBundle) (NormalizedBundle, error) {
	if len(bundles) == 0 {
		return NormalizedBundle{}, fmt.Errorf("nenhum bundle informado")
	}
	result := newNormalizedBundle(bundles[0].Engagement.ID, bundles[0].Scope.ID)
	result.Engagement = bundles[0].Engagement
	result.Scope = bundles[0].Scope
	for _, bundle := range bundles {
		if bundle.SchemaVersion != CanonicalSchemaVersion {
			return NormalizedBundle{}, fmt.Errorf("schema canônico incompatível: %s", bundle.SchemaVersion)
		}
		if result.Engagement.ID != bundle.Engagement.ID || result.Scope.ID != bundle.Scope.ID {
			return NormalizedBundle{}, fmt.Errorf("bundles pertencem a engagement/scope diferentes")
		}
		result.Assets = append(result.Assets, bundle.Assets...)
		result.Findings = append(result.Findings, bundle.Findings...)
		result.Evidence = append(result.Evidence, bundle.Evidence...)
		result.ToolRuns = append(result.ToolRuns, bundle.ToolRuns...)
		result.RequestLogs = append(result.RequestLogs, bundle.RequestLogs...)
		result.Exceptions = append(result.Exceptions, bundle.Exceptions...)
	}
	finalizeBundle(&result)
	return result, ValidateBundle(result)
}

func finalizeBundle(bundle *NormalizedBundle) {
	bundle.GeneratedAt = now()
	bundle.Assets = mergeAssets(bundle.Assets)
	bundle.Evidence = mergeEvidence(bundle.Evidence)
	bundle.ToolRuns = mergeToolRuns(bundle.ToolRuns)
	bundle.RequestLogs = mergeRequestLogs(bundle.RequestLogs)
	bundle.Exceptions = mergeExceptions(bundle.Exceptions)
	bundle.Findings = mergeFindings(bundle.Findings, bundle.Evidence)
	bundle.Engagement.ToolRunIDs = make([]string, 0, len(bundle.ToolRuns))
	for _, run := range bundle.ToolRuns {
		bundle.Engagement.ToolRunIDs = append(bundle.Engagement.ToolRunIDs, run.ID)
	}
	sort.Strings(bundle.Engagement.ToolRunIDs)
	if bundle.Scope.Fingerprint == "" {
		bundle.Scope.Fingerprint = scopeFingerprint(bundle.Scope)
	}
	bundle.Summary = summarizeBundle(*bundle)
}

func mergeAssets(items []Asset) []Asset {
	byID := make(map[string]Asset)
	for _, item := range items {
		current, ok := byID[item.ID]
		if !ok {
			byID[item.ID] = item
			continue
		}
		current.Technologies = dedupeSortedCase(append(current.Technologies, item.Technologies...))
		current.SourceToolRunIDs = dedupeSortedCase(append(current.SourceToolRunIDs, item.SourceToolRunIDs...))
		current.FirstSeen = earliest(current.FirstSeen, item.FirstSeen)
		current.LastSeen = latest(current.LastSeen, item.LastSeen)
		byID[item.ID] = current
	}
	result := make([]Asset, 0, len(byID))
	for _, item := range byID {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func mergeEvidence(items []Evidence) []Evidence {
	byID := make(map[string]Evidence)
	for _, item := range items {
		if existing, ok := byID[item.ID]; ok {
			existing.Sensitive = existing.Sensitive || item.Sensitive
			if existing.ReviewStatus == "unreviewed" && item.ReviewStatus != "" {
				existing.ReviewStatus = item.ReviewStatus
			}
			byID[item.ID] = existing
		} else {
			byID[item.ID] = item
		}
	}
	result := make([]Evidence, 0, len(byID))
	for _, item := range byID {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func mergeToolRuns(items []ToolRun) []ToolRun {
	byID := make(map[string]ToolRun)
	for _, item := range items {
		byID[item.ID] = item
	}
	result := make([]ToolRun, 0, len(byID))
	for _, item := range byID {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func mergeRequestLogs(items []RequestLog) []RequestLog {
	byID := make(map[string]RequestLog)
	for _, item := range items {
		byID[item.ID] = item
	}
	result := make([]RequestLog, 0, len(byID))
	for _, item := range byID {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func mergeExceptions(items []Exception) []Exception {
	byID := make(map[string]Exception)
	for _, item := range items {
		if item.ID == "" {
			item.ID = stableID("exception", item.ToolRunID+"|"+item.Stage+"|"+item.Type+"|"+item.Reason)
		}
		byID[item.ID] = item
	}
	result := make([]Exception, 0, len(byID))
	for _, item := range byID {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func mergeFindings(items []Finding, evidence []Evidence) []Finding {
	evidenceMap := make(map[string]Evidence)
	for _, item := range evidence {
		evidenceMap[item.ID] = item
	}
	byFingerprint := make(map[string]Finding)
	for _, item := range items {
		if item.Fingerprint == "" {
			item.Fingerprint = findingFingerprint(item)
			item.ID = stableID("finding", item.Fingerprint)
		}
		current, ok := byFingerprint[item.Fingerprint]
		if !ok {
			byFingerprint[item.Fingerprint] = item
			continue
		}
		current.Severity = maxSeverity(current.Severity, item.Severity)
		current.Confidence = maxConfidence(current.Confidence, item.Confidence)
		current.AssetIDs = dedupeSortedCase(append(current.AssetIDs, item.AssetIDs...))
		current.EvidenceIDs = dedupeSortedCase(append(current.EvidenceIDs, item.EvidenceIDs...))
		current.ToolRunIDs = dedupeSortedCase(append(current.ToolRunIDs, item.ToolRunIDs...))
		current.Identifiers = dedupeSortedCase(append(current.Identifiers, item.Identifiers...))
		current.Occurrences += item.Occurrences
		current.FirstObserved = earliest(current.FirstObserved, item.FirstObserved)
		current.LastObserved = latest(current.LastObserved, item.LastObserved)
		current.Status = strongestStatus(current.Status, item.Status)
		byFingerprint[item.Fingerprint] = current
	}
	result := make([]Finding, 0, len(byFingerprint))
	for _, item := range byFingerprint {
		item.Rank = rankFinding(item, evidenceMap)
		item.Commercial = commercialValue(item)
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Rank.Total == result[j].Rank.Total {
			return result[i].Fingerprint < result[j].Fingerprint
		}
		return result[i].Rank.Total > result[j].Rank.Total
	})
	return result
}

func rankFinding(f Finding, evidence map[string]Evidence) RankBreakdown {
	severityPoints := map[string]int{"info": 0, "low": 8, "medium": 20, "high": 30, "critical": 40}[normalizeSeverity(f.Severity)]
	confidencePoints := map[string]int{"low": 4, "medium": 10, "high": 15}[normalizeConfidence(f.Confidence)]
	evidencePoints := 0
	reviewed := false
	for _, id := range f.EvidenceIDs {
		if item, ok := evidence[id]; ok && item.SHA256 != "" {
			evidencePoints = 8
			if item.ReviewStatus == "reviewed" || item.ReviewStatus == "reproduced" {
				reviewed = true
			}
		}
	}
	if len(f.EvidenceIDs) > 1 || len(f.ToolRunIDs) > 1 {
		evidencePoints = maxInt(evidencePoints, 12)
	}
	if reviewed {
		evidencePoints = 15
	}
	exposurePoints := map[string]int{"public": 10, "authenticated": 7, "local": 4, "unknown": 2}[defaultString(f.Exposure, "unknown")]
	exploitabilityPoints := map[string]int{"tool-verified": 10, "reproduced": 10, "probable": 7, "candidate": 4, "unknown": 2}[defaultString(f.Exploitability, "candidate")]
	businessPoints := 3
	switch f.Category {
	case "secret-exposure", "authorization", "authentication", "sql-injection", "remote-code-execution":
		businessPoints = 10
	case "xss", "dependency-vulnerability", "tls-misconfiguration", "api-contract":
		businessPoints = 6
	}
	result := RankBreakdown{Severity: severityPoints, Confidence: confidencePoints, Evidence: evidencePoints, Exposure: exposurePoints, Exploitability: exploitabilityPoints, BusinessImpact: businessPoints}
	result.Total = result.Severity + result.Confidence + result.Evidence + result.Exposure + result.Exploitability + result.BusinessImpact
	if f.Status == "false_positive" || f.Status == "fixed_verified" {
		result.Total = 0
	} else if f.Status == "candidate" && result.Total > 79 {
		result.Total = 79
	}
	return result
}

func commercialValue(f Finding) CommercialValue {
	proof := "candidate"
	if len(f.ToolRunIDs) > 1 || len(f.EvidenceIDs) > 1 {
		proof = "corroborated"
	}
	if f.Status == "validated" {
		proof = "validated"
	}
	relevance := "low"
	if f.Rank.Total >= 70 {
		relevance = "high"
	} else if f.Rank.Total >= 40 {
		relevance = "medium"
	}
	statement := "Sinal técnico registrado; útil para priorização, ainda sem impacto comercial quantificado."
	if proof == "corroborated" {
		statement = "Sinal corroborado por mais de uma origem/evidência; sustenta uma conversa técnica, não uma promessa financeira."
	}
	if proof == "validated" {
		statement = "Achado tecnicamente validado e rastreável; pode sustentar priorização executiva dentro do escopo confirmado."
	}
	return CommercialValue{Relevance: relevance, ProofLevel: proof, Statement: statement, Limitation: "O normalizador não calcula perda evitada, ROI ou impacto financeiro sem dados do cliente."}
}

func summarizeBundle(bundle NormalizedBundle) BundleSummary {
	summary := BundleSummary{Assets: len(bundle.Assets), Findings: len(bundle.Findings), Evidence: len(bundle.Evidence), ToolRuns: len(bundle.ToolRuns), Exceptions: len(bundle.Exceptions), FindingsBySeverity: map[string]int{}}
	withEvidence := 0
	for _, finding := range bundle.Findings {
		summary.FindingsBySeverity[finding.Severity]++
		if finding.Status == "validated" {
			summary.ValidatedFindings++
		} else if finding.Status == "candidate" {
			summary.CandidateFindings++
		}
		if len(finding.EvidenceIDs) > 0 {
			withEvidence++
		}
	}
	if len(bundle.Findings) > 0 {
		summary.EvidenceCoverage = float64(withEvidence) / float64(len(bundle.Findings))
	}
	return summary
}

func CompareBundles(previous, current NormalizedBundle) Comparison {
	result := Comparison{SchemaVersion: CanonicalSchemaVersion, GeneratedAt: now(), PreviousEngagement: previous.Engagement.ID, CurrentEngagement: current.Engagement.ID, New: []string{}, Recurring: []string{}, NotObserved: []string{}, FixedVerified: []string{}}
	if previous.Scope.Fingerprint == "" || current.Scope.Fingerprint == "" || previous.Scope.Fingerprint != current.Scope.Fingerprint {
		result.Reasons = append(result.Reasons, "escopos não possuem o mesmo fingerprint")
	}
	previousTools := successfulTools(previous)
	currentTools := successfulTools(current)
	for tool := range previousTools {
		if _, ok := currentTools[tool]; !ok {
			result.Reasons = append(result.Reasons, "cobertura atual não inclui "+tool)
		}
	}
	result.Comparable = len(result.Reasons) == 0
	prevMap := make(map[string]Finding)
	currMap := make(map[string]Finding)
	for _, item := range previous.Findings {
		prevMap[item.Fingerprint] = item
	}
	for _, item := range current.Findings {
		currMap[item.Fingerprint] = item
		if item.Status == "fixed_verified" {
			if _, ok := prevMap[item.Fingerprint]; ok {
				result.FixedVerified = append(result.FixedVerified, item.Fingerprint)
			}
		} else if _, ok := prevMap[item.Fingerprint]; ok {
			result.Recurring = append(result.Recurring, item.Fingerprint)
		} else {
			result.New = append(result.New, item.Fingerprint)
		}
	}
	for fingerprint := range prevMap {
		if _, ok := currMap[fingerprint]; !ok {
			result.NotObserved = append(result.NotObserved, fingerprint)
		}
	}
	sort.Strings(result.New)
	sort.Strings(result.Recurring)
	sort.Strings(result.NotObserved)
	sort.Strings(result.FixedVerified)
	result.ScoreDelta = totalActiveScore(current) - totalActiveScore(previous)
	return result
}

func successfulTools(bundle NormalizedBundle) map[string]struct{} {
	result := make(map[string]struct{})
	for _, run := range bundle.ToolRuns {
		if run.Status == "completed" || run.Status == "normalized" {
			result[run.Tool] = struct{}{}
		}
	}
	return result
}

func totalActiveScore(bundle NormalizedBundle) int {
	total := 0
	for _, item := range bundle.Findings {
		if item.Status != "false_positive" && item.Status != "fixed_verified" {
			total += item.Rank.Total
		}
	}
	return total
}

func ValidateBundle(bundle NormalizedBundle) error {
	if bundle.SchemaVersion != CanonicalSchemaVersion {
		return fmt.Errorf("schema_version inválida")
	}
	if bundle.Engagement.ID == "" || bundle.Scope.ID == "" || bundle.Engagement.ScopeID != bundle.Scope.ID {
		return fmt.Errorf("engagement/scope inválidos")
	}
	if !validSHA256Hex(bundle.Scope.Fingerprint) {
		return fmt.Errorf("scope sem fingerprint SHA-256 válido")
	}
	assets, evidence, runs := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}
	for _, item := range bundle.Assets {
		if _, ok := assets[item.ID]; ok {
			return fmt.Errorf("asset duplicado: %s", item.ID)
		}
		assets[item.ID] = struct{}{}
	}
	for _, item := range bundle.Evidence {
		if !validSHA256Hex(item.SHA256) {
			return fmt.Errorf("evidence sem SHA-256 válido: %s", item.ID)
		}
		evidence[item.ID] = struct{}{}
	}
	for _, item := range bundle.ToolRuns {
		if item.ID == "" || !validSHA256Hex(item.InputSHA256) {
			return fmt.Errorf("tool run sem ID/input SHA-256 válido: %s", item.ID)
		}
		if item.CommandSHA256 != "" && !validSHA256Hex(item.CommandSHA256) {
			return fmt.Errorf("tool run sem command SHA-256 válido: %s", item.ID)
		}
		runs[item.ID] = struct{}{}
	}
	seenFinding := make(map[string]struct{})
	for _, finding := range bundle.Findings {
		if !validSHA256Hex(finding.Fingerprint) {
			return fmt.Errorf("finding sem fingerprint SHA-256 válido: %s", finding.ID)
		}
		if _, ok := seenFinding[finding.Fingerprint]; ok {
			return fmt.Errorf("finding não deduplicado: %s", finding.Fingerprint)
		}
		seenFinding[finding.Fingerprint] = struct{}{}
		for _, id := range finding.AssetIDs {
			if _, ok := assets[id]; !ok {
				return fmt.Errorf("finding referencia asset ausente: %s", id)
			}
		}
		for _, id := range finding.EvidenceIDs {
			if _, ok := evidence[id]; !ok {
				return fmt.Errorf("finding referencia evidence ausente: %s", id)
			}
		}
		for _, id := range finding.ToolRunIDs {
			if _, ok := runs[id]; !ok {
				return fmt.Errorf("finding referencia tool run ausente: %s", id)
			}
		}
	}
	return nil
}

func validSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') && (character < 'A' || character > 'F') {
			return false
		}
	}
	return true
}

func scopeFingerprint(scope Scope) string {
	copyScope := scope
	copyScope.Fingerprint = ""
	copyScope.AllowedSchemes = dedupeSortedCase(copyScope.AllowedSchemes)
	copyScope.AllowedHosts = dedupeSortedCase(copyScope.AllowedHosts)
	sort.Ints(copyScope.AllowedPorts)
	copyScope.ExcludedPaths = dedupeSortedCase(copyScope.ExcludedPaths)
	data, _ := json.Marshal(copyScope)
	return HashBytes(data)
}

func WriteBundle(path string, bundle NormalizedBundle) error {
	finalizeBundle(&bundle)
	if err := ValidateBundle(bundle); err != nil {
		return err
	}
	return WriteJSON(path, bundle)
}

func ReadBundle(path string) (NormalizedBundle, error) {
	var bundle NormalizedBundle
	if err := ReadJSON(path, &bundle); err != nil {
		return bundle, err
	}
	return bundle, ValidateBundle(bundle)
}

func WriteComparison(path string, comparison Comparison) error { return WriteJSON(path, comparison) }

func WriteNormalizedReport(path string, bundle NormalizedBundle) error {
	var builder strings.Builder
	builder.WriteString("# Relatório normalizado — " + bundle.Engagement.ID + "\n\n")
	builder.WriteString(fmt.Sprintf("Assets: **%d**  \nAchados: **%d** (%d validados, %d candidatos)  \nCobertura de evidência: **%.0f%%**  \nExecuções: **%d**  \nExceções: **%d**\n\n", bundle.Summary.Assets, bundle.Summary.Findings, bundle.Summary.ValidatedFindings, bundle.Summary.CandidateFindings, bundle.Summary.EvidenceCoverage*100, bundle.Summary.ToolRuns, bundle.Summary.Exceptions))
	builder.WriteString("## Achados priorizados\n\n")
	if len(bundle.Findings) == 0 {
		builder.WriteString("Nenhum achado normalizado. Isso não prova ausência de vulnerabilidades.\n")
	}
	for _, finding := range bundle.Findings {
		builder.WriteString(fmt.Sprintf("- **%d/100 — %s** (`%s`, %s, %s): %s\n", finding.Rank.Total, finding.Title, finding.Severity, finding.Status, finding.Commercial.ProofLevel, finding.Commercial.Statement))
	}
	builder.WriteString("\n## Limites de prova\n\n")
	builder.WriteString("O SHA-256 prova integridade do artefato registrado, não a veracidade automática do achado. Saídas de scanner permanecem candidatas até revisão/reprodução. `not_observed` em comparação não significa corrigido sem cobertura comparável e reteste explícito. Não há cálculo automático de ROI ou perda evitada.\n")
	return os.WriteFile(path, []byte(builder.String()), 0o600)
}

func choose(condition bool, yes, no string) string {
	if condition {
		return yes
	}
	return no
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func normalizeSeverity(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if strings.Contains(value, "critical") {
		return "critical"
	}
	if strings.Contains(value, "high") {
		return "high"
	}
	if strings.Contains(value, "medium") || strings.Contains(value, "moderate") {
		return "medium"
	}
	if strings.Contains(value, "low") {
		return "low"
	}
	switch value {
	case "critical", "crit", "4":
		return "critical"
	case "high", "3", "error":
		return "high"
	case "medium", "moderate", "2", "warning", "warn":
		return "medium"
	case "low", "1", "note":
		return "low"
	case "info", "informational", "ok", "0", "none", "unknown", "":
		return "info"
	default:
		return "info"
	}
}

func normalizeConfidence(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "high", "confirmed", "certain", "3":
		return "high"
	case "medium", "moderate", "2":
		return "medium"
	default:
		return "low"
	}
}

func normalizeCategory(category, title string) string {
	value := strings.ToLower(category + " " + title)
	switch {
	case strings.Contains(value, "secret") || strings.Contains(value, "credential") || strings.Contains(value, "token"):
		return "secret-exposure"
	case strings.Contains(value, "sql") && strings.Contains(value, "inject"):
		return "sql-injection"
	case strings.Contains(value, "xss") || strings.Contains(value, "cross-site scripting"):
		return "xss"
	case strings.Contains(value, "authorization") || strings.Contains(value, "bola") || strings.Contains(value, "access control"):
		return "authorization"
	case strings.Contains(value, "authentication") || strings.Contains(value, "jwt") || strings.Contains(value, "session"):
		return "authentication"
	case strings.Contains(value, "tls") || strings.Contains(value, "ssl") || strings.Contains(value, "cipher"):
		return "tls-misconfiguration"
	case strings.Contains(value, "depend") || strings.Contains(value, "vulnerab") || strings.Contains(value, "cve") || strings.Contains(value, "ghsa"):
		return "dependency-vulnerability"
	case strings.Contains(value, "schema") || strings.Contains(value, "api contract") || strings.Contains(value, "schemathesis"):
		return "api-contract"
	case strings.TrimSpace(category) != "":
		return strings.ToLower(strings.TrimSpace(category))
	default:
		return "other"
	}
}

func maxSeverity(a, b string) string {
	order := map[string]int{"info": 0, "low": 1, "medium": 2, "high": 3, "critical": 4}
	a, b = normalizeSeverity(a), normalizeSeverity(b)
	if order[b] > order[a] {
		return b
	}
	return a
}

func maxConfidence(a, b string) string {
	order := map[string]int{"low": 0, "medium": 1, "high": 2}
	a, b = normalizeConfidence(a), normalizeConfidence(b)
	if order[b] > order[a] {
		return b
	}
	return a
}

func strongestStatus(a, b string) string {
	order := map[string]int{"candidate": 0, "accepted_risk": 1, "validated": 2, "false_positive": 3, "fixed_verified": 4}
	if order[b] > order[a] {
		return b
	}
	return a
}

func dedupeSortedCase(items []string) []string {
	seen := make(map[string]string)
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			seen[strings.ToLower(item)] = item
		}
	}
	result := make([]string, 0, len(seen))
	for _, item := range seen {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return strings.ToLower(result[i]) < strings.ToLower(result[j]) })
	return result
}

func earliest(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	if ta, err := time.Parse(time.RFC3339, a); err == nil {
		if tb, err := time.Parse(time.RFC3339, b); err == nil && tb.Before(ta) {
			return b
		}
	}
	return a
}

func latest(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	if ta, err := time.Parse(time.RFC3339, a); err == nil {
		if tb, err := time.Parse(time.RFC3339, b); err == nil && tb.After(ta) {
			return b
		}
	}
	return a
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
