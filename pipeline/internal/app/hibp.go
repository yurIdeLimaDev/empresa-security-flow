package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var publicEmailPattern = regexp.MustCompile(`(?i)\b[a-z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)+\b`)

type HIBPResult struct {
	AccountSHA256 string   `json:"account_sha256"`
	HTTPStatus    int      `json:"http_status"`
	Breaches      []string `json:"breaches"`
	CheckedAt     string   `json:"checked_at"`
}

func DiscoverPublicEmails(ctx context.Context, client *http.Client, base *url.URL, paths []string, maxBytes int64) ([]string, []HTTPObservation) {
	emails := map[string]struct{}{}
	observations := []HTTPObservation{}
	for _, path := range paths {
		if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
			continue
		}
		target := base.ResolveReference(&url.URL{Path: path})
		observation, body := doObservedRequest(ctx, client, http.MethodGet, target.String(), nil, min64(maxBytes, 1<<20))
		observations = append(observations, observation)
		if observation.Status != http.StatusOK {
			continue
		}
		for _, match := range publicEmailPattern.FindAllString(string(body), -1) {
			email := strings.ToLower(strings.TrimSpace(match))
			if strings.HasSuffix(email, "@"+strings.ToLower(base.Hostname())) {
				emails[email] = struct{}{}
			}
			if len(emails) >= 20 {
				break
			}
		}
		if len(emails) >= 20 {
			break
		}
	}
	result := make([]string, 0, len(emails))
	for email := range emails {
		result = append(result, email)
	}
	sort.Strings(result)
	return result, observations
}

func RunHIBPContainer(ctx context.Context, options GovernedRunOptions, cfg LeadConfig, emails []string, action *PlannedAction) error {
	policy, err := options.Policies.Tool("hibp", 1)
	if err != nil {
		return gateBlock(action, err.Error())
	}
	if policy.Mode != "automatic" {
		return gateBlock(action, "HIBP não está automatic no tool-policy")
	}
	if policy.Flags.DomainSearch {
		return gateBlock(action, "HIBP domain_search é proibido")
	}
	if len(emails) == 0 {
		action.State, action.Reason = "skipped", "nenhum e-mail corporativo publicado foi encontrado nas páginas configuradas"
		return nil
	}
	apiKey, exists := os.LookupEnv(cfg.HIBPAPIKeyEnv)
	if !exists || strings.TrimSpace(apiKey) == "" {
		return gateBlock(action, "variável HIBP ausente: "+cfg.HIBPAPIKeyEnv)
	}
	if strings.ContainsAny(apiKey, "\r\n") || strings.ContainsAny(cfg.HIBPUserAgent, "\r\n") {
		return gateBlock(action, "credencial ou user-agent HIBP contém quebra de linha")
	}
	lockEntry, err := ApproveContainer(options.Lock, "hibp", options.SBOMPath)
	if err != nil {
		return gateBlock(action, err.Error())
	}
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return gateBlock(action, "Docker não instalado")
	}
	started := time.Now().UTC()
	runID := "run_" + started.Format("20060102t150405000000000") + "_" + HashBytes([]byte(action.ID + started.String()))[:8]
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(policy.MaxDurationMinutes)*time.Minute)
	defer cancel()
	results := make([]HIBPResult, 0, len(emails))
	exitCode := 0
	for index, email := range emails {
		if index > 0 {
			timer := time.NewTimer(2100 * time.Millisecond)
			select {
			case <-timer.C:
			case <-runCtx.Done():
				timer.Stop()
				exitCode = -1
				break
			}
		}
		endpoint := "https://haveibeenpwned.com/api/v3/breachedaccount/" + url.PathEscape(email) + "?truncateResponse=true&includeUnverified=true"
		curlConfig := fmt.Sprintf("url = %s\nheader = %s\nuser-agent = %s\n", curlConfigQuote(endpoint), curlConfigQuote("hibp-api-key: "+apiKey), curlConfigQuote(cfg.HIBPUserAgent))
		args := []string{"run", "--rm", "-i", "--network", options.Network.NetworkName, "--dns", options.Policies.Engagement.Environment.DNSResolverIP, "--sysctl", "net.ipv6.conf.all.disable_ipv6=1", lockEntry.Container.Image + "@" + normalizeDigest(lockEntry.Container.Digest), "--config", "-", "--silent", "--show-error", "--output", "-", "--write-out", "\n%{http_code}"}
		if options.Network.HTTPProxyURL != "" {
			args = append(args, "--proxy", proxyURLForRun(options.Network.HTTPProxyURL, runID))
		}
		command := exec.CommandContext(runCtx, dockerPath, args...)
		command.Stdin = strings.NewReader(curlConfig)
		var output, stderr bytes.Buffer
		command.Stdout, command.Stderr = &output, &stderr
		if err := command.Run(); err != nil {
			exitCode = -1
			action.Reason = "HIBP: " + err.Error()
			break
		}
		body, status, parseErr := splitCurlResponse(output.Bytes())
		if parseErr != nil {
			exitCode = -1
			action.Reason = parseErr.Error()
			break
		}
		item := HIBPResult{AccountSHA256: HashBytes([]byte(email)), HTTPStatus: status, Breaches: []string{}, CheckedAt: now()}
		switch status {
		case 200:
			var rows []struct {
				Name string `json:"Name"`
			}
			if err := json.Unmarshal(body, &rows); err != nil {
				exitCode = -1
				action.Reason = "resposta HIBP inválida"
				break
			}
			for _, row := range rows {
				if row.Name != "" {
					item.Breaches = append(item.Breaches, row.Name)
				}
			}
			sort.Strings(item.Breaches)
		case 404:
		default:
			exitCode = status
			action.Reason = fmt.Sprintf("HIBP retornou status %d; nenhuma inferência foi gerada", status)
		}
		results = append(results, item)
		if exitCode != 0 {
			break
		}
	}
	artifact := filepath.Join(options.CaseDir, "hibp-email-results.json")
	if err := WriteJSON(artifact, results); err != nil {
		return err
	}
	ended := time.Now().UTC()
	action.ExitCode = exitCode
	if exitCode == 0 {
		action.State = "completed"
	} else {
		action.State = "failed"
	}
	inputHashes, _ := hashInputs(action.InputPaths)
	log := ExecutionLog{RunID: runID, EngagementID: options.Policies.Engagement.EngagementID, Tool: "hibp", ToolDigest: normalizeDigest(lockEntry.Container.Digest), PolicyFilesHash: options.Policies.Hashes, InputHashes: inputHashes, Network: ExecutionNetworkLog{Bridge: options.Network.NetworkName, Chain: options.Network.Chain, ProxyLog: options.Network.ProxyLog}, StartedAt: started.Format(time.RFC3339Nano), EndedAt: ended.Format(time.RFC3339Nano), ExitCode: exitCode, Counts: ExecutionCounts{Requests: len(results)}, Artifacts: []string{filepath.Base(artifact)}, ReviewedBy: nil, ReviewedAt: nil}
	logDir := filepath.Join(options.CaseDir, "execution-logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return err
	}
	logPath := filepath.Join(logDir, runID+".json")
	if err := WriteJSON(logPath, log); err != nil {
		return err
	}
	if err := validateJSONSchemaFile("execution-log.schema.json", logPath); err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("HIBP falhou com status/exit %d", exitCode)
	}
	return nil
}

func curlConfigQuote(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return "\"" + value + "\""
}

func splitCurlResponse(data []byte) ([]byte, int, error) {
	trimmed := bytes.TrimRight(data, "\r\n")
	index := bytes.LastIndexByte(trimmed, '\n')
	if index < 0 {
		return nil, 0, fmt.Errorf("curl não retornou status HTTP")
	}
	statusText := strings.TrimSpace(string(trimmed[index+1:]))
	var status int
	if _, err := fmt.Sscanf(statusText, "%d", &status); err != nil {
		return nil, 0, fmt.Errorf("status HIBP inválido")
	}
	return bytes.TrimSpace(trimmed[:index]), status, nil
}
