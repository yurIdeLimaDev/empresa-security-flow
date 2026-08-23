package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type NetworkState struct {
	NetworkName    string `json:"network_name"`
	Chain          string `json:"chain"`
	InputChain     string `json:"input_chain"`
	ProxyLog       string `json:"proxy_log"`
	SetupScript    string `json:"setup_script"`
	CanaryScript   string `json:"canary_script"`
	StopScript     string `json:"stop_script"`
	TeardownScript string `json:"teardown_script"`
	HTTPProxyURL   string `json:"http_proxy_url,omitempty"`
	SOCKSProxyURL  string `json:"socks_proxy_url,omitempty"`
}

// IsolationSmokeResult is written by the operator-owned integration check. It
// deliberately contains only control-plane metadata: no target response body
// or credential can enter this artifact.
type IsolationSmokeResult struct {
	EngagementID string `json:"engagement_id"`
	NetworkName  string `json:"network_name"`
	ProxyURL     string `json:"proxy_url"`
	Canary       string `json:"canary"`
	ProxyRequest string `json:"proxy_request"`
}

// ValidateIsolationPolicy validates the portion of an engagement policy used
// before any external tool is selected. This lets an operator prove Docker,
// firewall, canaries and proxy behavior with controlled infrastructure while
// the tool supply chain remains fail-closed.
func ValidateIsolationPolicy(policy EngagementPolicy) error {
	pipeline := 0
	switch policy.PipelineMode {
	case "pipeline1_public_low_impact":
		pipeline = 1
	case "pipeline2_active":
		pipeline = 2
	default:
		return fmt.Errorf("pipeline_mode inválido para teste de isolamento: %q", policy.PipelineMode)
	}
	return validatePolicySemantics(&ValidatedPolicies{Engagement: policy, Tools: ToolPolicyFile{Tools: map[string]ToolPolicyEntry{}}}, pipeline)
}

// RunIsolationSmoke performs no scanner action. It applies the same firewall
// and canary controls used by execution, then sends one HTTP request through
// the audit proxy from a container on the dedicated engagement network.
// It is intentionally for operator-controlled Linux/root test infrastructure.
func RunIsolationSmoke(ctx context.Context, policy EngagementPolicy, caseDirectory string, maxRequests int, rps float64) (IsolationSmokeResult, error) {
	if err := ValidateIsolationPolicy(policy); err != nil {
		return IsolationSmokeResult{}, err
	}
	if maxRequests < 1 || rps <= 0 {
		return IsolationSmokeResult{}, fmt.Errorf("max_requests e requests_per_second precisam ser positivos")
	}
	if entries, err := os.ReadDir(caseDirectory); err == nil && len(entries) != 0 {
		return IsolationSmokeResult{}, fmt.Errorf("case-dir do isolation-smoke precisa estar vazio para preservar a evidência de uma única execução")
	} else if err != nil && !os.IsNotExist(err) {
		return IsolationSmokeResult{}, err
	}
	state, err := GenerateIsolationScripts(caseDirectory, policy)
	if err != nil {
		return IsolationSmokeResult{}, err
	}
	if err := SetupAndTestIsolation(ctx, state, caseDirectory); err != nil {
		return IsolationSmokeResult{}, err
	}
	defer func() { _ = TeardownIsolation(context.Background(), state) }()

	proxy, err := StartEngagementAuditProxy(ctx, &state, policy, maxRequests, rps)
	if err != nil {
		return IsolationSmokeResult{}, fmt.Errorf("proxy de auditoria: %w", err)
	}
	defer func() { _ = proxy.Close(context.Background()) }()
	if err := proxyCanaryRequest(ctx, state, policy); err != nil {
		return IsolationSmokeResult{}, err
	}
	result := IsolationSmokeResult{
		EngagementID: policy.EngagementID,
		NetworkName:  state.NetworkName,
		ProxyURL:     state.HTTPProxyURL,
		Canary:       "negative-blocked-and-positive-allowed",
		ProxyRequest: "allowed-pinned-request-logged",
	}
	if err := WriteJSON(filepath.Join(caseDirectory, "isolation-smoke.json"), result); err != nil {
		return IsolationSmokeResult{}, err
	}
	return result, nil
}

func proxyCanaryRequest(ctx context.Context, state NetworkState, policy EngagementPolicy) error {
	if state.HTTPProxyURL == "" {
		return fmt.Errorf("proxy HTTP não foi iniciado")
	}
	command := exec.CommandContext(ctx, "docker", "run", "--rm", "--network", state.NetworkName,
		"--dns", policy.Environment.DNSResolverIP, "--sysctl", "net.ipv6.conf.all.disable_ipv6=1",
		policy.Environment.CanaryImage, "--silent", "--show-error", "--fail", "--connect-timeout", "5", "--max-time", "15",
		"--proxy", state.HTTPProxyURL, "--proxy-user", "run_isolation_smoke:", policy.Canary.PositiveURL)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("request de controle pelo proxy: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if countRequestLogLines(state.ProxyLog, "run_isolation_smoke") != 1 {
		return fmt.Errorf("proxy não registrou exatamente uma requisição de controle")
	}
	data, err := os.ReadFile(state.ProxyLog)
	if err != nil {
		return err
	}
	var matched bool
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var request RequestLog
		if json.Unmarshal([]byte(line), &request) == nil && request.ToolRunID == "run_isolation_smoke" && request.ScopeDecision == "allowed_pinned" {
			matched = true
		}
	}
	if !matched {
		return fmt.Errorf("proxy não registrou decisão allowed_pinned para o controle")
	}
	return nil
}

func GenerateIsolationScripts(caseDirectory string, policy EngagementPolicy) (NetworkState, error) {
	isolationDir := filepath.Join(caseDirectory, "isolation")
	if err := os.MkdirAll(isolationDir, 0o700); err != nil {
		return NetworkState{}, err
	}
	state := NetworkState{
		NetworkName:    policy.EngagementID + "_net",
		Chain:          isolationChain(policy.EngagementID, "EGR"),
		InputChain:     isolationChain(policy.EngagementID, "INP"),
		ProxyLog:       filepath.Join(caseDirectory, "requestlog.ndjson"),
		SetupScript:    filepath.Join(isolationDir, "network-setup.sh"),
		CanaryScript:   filepath.Join(isolationDir, "network-canary.sh"),
		StopScript:     filepath.Join(isolationDir, policy.EmergencyStop.Script),
		TeardownScript: filepath.Join(isolationDir, "network-teardown.sh"),
	}
	setup := renderSetupScript(state, policy)
	canary, err := renderCanaryScript(state, policy)
	if err != nil {
		return NetworkState{}, err
	}
	stop := renderStopScript(state, caseDirectory)
	teardown := renderTeardownScript(state)
	for path, content := range map[string]string{
		state.SetupScript:    setup,
		state.CanaryScript:   canary,
		state.StopScript:     stop,
		state.TeardownScript: teardown,
	} {
		if err := os.WriteFile(path, []byte(content), 0o700); err != nil {
			return NetworkState{}, err
		}
	}
	if err := WriteJSON(filepath.Join(isolationDir, "network-state.json"), state); err != nil {
		return NetworkState{}, err
	}
	return state, nil
}

func isolationChain(engagementID, suffix string) string {
	prefix := strings.ToUpper(strings.ReplaceAll(engagementID, "-", "_"))
	if len(prefix) > 15 {
		prefix = prefix[:15]
	}
	return prefix + "_" + strings.ToUpper(HashBytes([]byte(engagementID))[:8]) + "_" + suffix
}

func renderSetupScript(state NetworkState, policy EngagementPolicy) string {
	lines := []string{
		"#!/bin/sh",
		"set -eu",
		"NETWORK=" + shellQuote(state.NetworkName),
		"CHAIN=" + shellQuote(state.Chain),
		"INPUT_CHAIN=" + shellQuote(state.InputChain),
		"DNS_IP=" + shellQuote(policy.Environment.DNSResolverIP),
		"if ! docker network inspect \"$NETWORK\" >/dev/null 2>&1; then",
		"  docker network create --driver bridge --ipv6=false --opt com.docker.network.enable_ipv6=false \"$NETWORK\" >/dev/null",
		"fi",
		"if [ \"$(docker network inspect \"$NETWORK\" --format '{{.EnableIPv6}}')\" != \"false\" ]; then echo 'rede existente possui IPv6 habilitado' >&2; exit 1; fi",
		"BRIDGE=$(docker network inspect \"$NETWORK\" --format '{{index .Options \"com.docker.network.bridge.name\"}}')",
		"if [ -z \"$BRIDGE\" ] || [ \"$BRIDGE\" = \"<no value>\" ]; then",
		"  NET_ID=$(docker network inspect \"$NETWORK\" --format '{{.Id}}')",
		"  BRIDGE=br-$(printf '%s' \"$NET_ID\" | cut -c1-12)",
		"fi",
		"iptables -w -N \"$CHAIN\" 2>/dev/null || true",
		"iptables -w -F \"$CHAIN\"",
		"while iptables -w -C DOCKER-USER -i \"$BRIDGE\" -j \"$CHAIN\" 2>/dev/null; do iptables -w -D DOCKER-USER -i \"$BRIDGE\" -j \"$CHAIN\"; done",
		"iptables -w -N \"$INPUT_CHAIN\" 2>/dev/null || true",
		"iptables -w -F \"$INPUT_CHAIN\"",
		"while iptables -w -C INPUT -i \"$BRIDGE\" -j \"$INPUT_CHAIN\" 2>/dev/null; do iptables -w -D INPUT -i \"$BRIDGE\" -j \"$INPUT_CHAIN\"; done",
		"iptables -w -A \"$INPUT_CHAIN\" -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT",
		"iptables -w -A \"$INPUT_CHAIN\" -j DROP",
		"iptables -w -I INPUT 1 -i \"$BRIDGE\" -j \"$INPUT_CHAIN\"",
		"iptables -w -A \"$CHAIN\" -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT",
		"iptables -w -A \"$CHAIN\" -d \"$DNS_IP\" -p udp --dport 53 -j ACCEPT",
		"iptables -w -A \"$CHAIN\" -d \"$DNS_IP\" -p tcp --dport 53 -j ACCEPT",
	}
	type rule struct {
		IP       string
		Port     int
		Protocol string
	}
	rules := []rule{}
	for _, target := range policy.Environment.Targets {
		rules = append(rules, rule{IP: target.IP, Port: target.Port, Protocol: "tcp"})
	}
	for _, api := range policy.Environment.ExternalAPIs {
		rules = append(rules, rule{IP: api.IP, Port: api.Port, Protocol: "tcp"})
	}
	if policy.PipelineMode == "pipeline2_active" {
		for _, ns := range policy.Environment.AuthoritativeNS {
			rules = append(rules, rule{IP: ns.IP, Port: 53, Protocol: "udp"}, rule{IP: ns.IP, Port: 53, Protocol: "tcp"})
		}
	}
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].IP != rules[j].IP {
			return rules[i].IP < rules[j].IP
		}
		if rules[i].Port != rules[j].Port {
			return rules[i].Port < rules[j].Port
		}
		return rules[i].Protocol < rules[j].Protocol
	})
	seen := map[string]struct{}{}
	for _, item := range rules {
		key := fmt.Sprintf("%s/%d/%s", item.IP, item.Port, item.Protocol)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		lines = append(lines, fmt.Sprintf("iptables -w -A \"$CHAIN\" -d %s -p %s --dport %d -j ACCEPT", shellQuote(item.IP), item.Protocol, item.Port))
	}
	lines = append(lines,
		"iptables -w -A \"$CHAIN\" -j DROP",
		"iptables -w -I DOCKER-USER 1 -i \"$BRIDGE\" -j \"$CHAIN\"",
		"printf '%s\\n' \"$BRIDGE\" > "+shellQuote(filepath.ToSlash(filepath.Join(filepath.Dir(state.SetupScript), "bridge.name"))),
	)
	return strings.Join(lines, "\n") + "\n"
}

func renderCanaryScript(state NetworkState, policy EngagementPolicy) (string, error) {
	negative, err := url.Parse(policy.Canary.NegativeURL)
	if err != nil {
		return "", err
	}
	positive, err := url.Parse(policy.Canary.PositiveURL)
	if err != nil {
		return "", err
	}
	negativeResolve := fmt.Sprintf("%s:%d:%s", negative.Hostname(), policy.Canary.NegativePort, policy.Canary.NegativeIP)
	positiveResolve := fmt.Sprintf("%s:%d:%s", positive.Hostname(), policy.Canary.PositivePort, policy.Canary.PositiveIP)
	lines := []string{
		"#!/bin/sh",
		"set -eu",
		"NETWORK=" + shellQuote(state.NetworkName),
		"DNS_IP=" + shellQuote(policy.Environment.DNSResolverIP),
		"IMAGE=" + shellQuote(policy.Environment.CanaryImage),
		"NEGATIVE_URL=" + shellQuote(policy.Canary.NegativeURL),
		"POSITIVE_URL=" + shellQuote(policy.Canary.PositiveURL),
		"if ! docker run --rm --network host \"$IMAGE\" --silent --show-error --fail --connect-timeout 5 --max-time 10 --resolve " + shellQuote(negativeResolve) + " \"$NEGATIVE_URL\" >/dev/null; then",
		"  echo 'canary negativo de controle não está acessível pelo host; não é possível provar o bloqueio' >&2",
		"  exit 1",
		"fi",
		"if docker run --rm --network \"$NETWORK\" --dns \"$DNS_IP\" --sysctl net.ipv6.conf.all.disable_ipv6=1 \"$IMAGE\" --silent --show-error --connect-timeout 5 --max-time 10 --resolve " + shellQuote(negativeResolve) + " \"$NEGATIVE_URL\" >/dev/null 2>&1; then",
		"  echo 'canary negativo alcançou destino fora da allowlist' >&2",
		"  exit 1",
		"fi",
		"docker run --rm --network \"$NETWORK\" --dns \"$DNS_IP\" --sysctl net.ipv6.conf.all.disable_ipv6=1 \"$IMAGE\" --silent --show-error --fail --connect-timeout 5 --max-time 15 --resolve " + shellQuote(positiveResolve) + " \"$POSITIVE_URL\" >/dev/null",
		"echo 'canary de egress aprovado'",
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func renderStopScript(state NetworkState, caseDirectory string) string {
	pidPath := filepath.ToSlash(filepath.Join(caseDirectory, "runner.pid"))
	bridgePath := filepath.ToSlash(filepath.Join(filepath.Dir(state.SetupScript), "bridge.name"))
	return strings.Join([]string{
		"#!/bin/sh", "set -eu",
		"NETWORK=" + shellQuote(state.NetworkName),
		"CHAIN=" + shellQuote(state.Chain),
		"INPUT_CHAIN=" + shellQuote(state.InputChain),
		"BRIDGE_FILE=" + shellQuote(bridgePath),
		"CONTAINERS=$(docker ps -q --filter network=\"$NETWORK\")",
		"if [ -n \"$CONTAINERS\" ]; then docker kill $CONTAINERS >/dev/null; fi",
		"iptables -w -N \"$CHAIN\" 2>/dev/null || true",
		"iptables -w -F \"$CHAIN\"",
		"iptables -w -A \"$CHAIN\" -j DROP",
		"iptables -w -N \"$INPUT_CHAIN\" 2>/dev/null || true",
		"iptables -w -F \"$INPUT_CHAIN\"",
		"iptables -w -A \"$INPUT_CHAIN\" -j DROP",
		"if [ -f \"$BRIDGE_FILE\" ]; then",
		"  BRIDGE=$(cat \"$BRIDGE_FILE\")",
		"  while iptables -w -C DOCKER-USER -i \"$BRIDGE\" -j \"$CHAIN\" 2>/dev/null; do iptables -w -D DOCKER-USER -i \"$BRIDGE\" -j \"$CHAIN\"; done",
		"  iptables -w -I DOCKER-USER 1 -i \"$BRIDGE\" -j \"$CHAIN\"",
		"  while iptables -w -C INPUT -i \"$BRIDGE\" -j \"$INPUT_CHAIN\" 2>/dev/null; do iptables -w -D INPUT -i \"$BRIDGE\" -j \"$INPUT_CHAIN\"; done",
		"  iptables -w -I INPUT 1 -i \"$BRIDGE\" -j \"$INPUT_CHAIN\"",
		"fi",
		"if [ -f " + shellQuote(pidPath) + " ]; then kill -TERM \"$(cat " + shellQuote(pidPath) + ")\" 2>/dev/null || true; fi",
	}, "\n") + "\n"
}

func renderTeardownScript(state NetworkState) string {
	bridgePath := filepath.ToSlash(filepath.Join(filepath.Dir(state.SetupScript), "bridge.name"))
	return strings.Join([]string{
		"#!/bin/sh", "set -eu",
		"NETWORK=" + shellQuote(state.NetworkName),
		"CHAIN=" + shellQuote(state.Chain),
		"INPUT_CHAIN=" + shellQuote(state.InputChain),
		"BRIDGE_FILE=" + shellQuote(bridgePath),
		"if [ -f \"$BRIDGE_FILE\" ]; then",
		"  BRIDGE=$(cat \"$BRIDGE_FILE\")",
		"  while iptables -w -C DOCKER-USER -i \"$BRIDGE\" -j \"$CHAIN\" 2>/dev/null; do iptables -w -D DOCKER-USER -i \"$BRIDGE\" -j \"$CHAIN\"; done",
		"  while iptables -w -C INPUT -i \"$BRIDGE\" -j \"$INPUT_CHAIN\" 2>/dev/null; do iptables -w -D INPUT -i \"$BRIDGE\" -j \"$INPUT_CHAIN\"; done",
		"fi",
		"iptables -w -F \"$CHAIN\" 2>/dev/null || true",
		"iptables -w -X \"$CHAIN\" 2>/dev/null || true",
		"iptables -w -F \"$INPUT_CHAIN\" 2>/dev/null || true",
		"iptables -w -X \"$INPUT_CHAIN\" 2>/dev/null || true",
		"docker network rm \"$NETWORK\" >/dev/null 2>&1 || true",
	}, "\n") + "\n"
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func SetupAndTestIsolation(ctx context.Context, state NetworkState, caseDirectory string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("execução isolada exige host Linux; neste host só é possível gerar e testar os scripts por unidade")
	}
	idOut, err := exec.CommandContext(ctx, "id", "-u").Output()
	if err != nil || strings.TrimSpace(string(idOut)) != "0" {
		return fmt.Errorf("setup de DOCKER-USER exige root no host Linux")
	}
	if err := os.WriteFile(filepath.Join(caseDirectory, "runner.pid"), []byte(fmt.Sprint(os.Getpid())+"\n"), 0o600); err != nil {
		return err
	}
	if err := runIsolationScript(ctx, state.SetupScript); err != nil {
		return fmt.Errorf("setup da rede: %w", err)
	}
	if err := runIsolationScript(ctx, state.CanaryScript); err != nil {
		_ = runIsolationScript(context.Background(), state.TeardownScript)
		return fmt.Errorf("canary de egress: %w", err)
	}
	return nil
}

func TeardownIsolation(ctx context.Context, state NetworkState) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	return runIsolationScript(ctx, state.TeardownScript)
}

func runIsolationScript(ctx context.Context, path string) error {
	command := exec.CommandContext(ctx, "/bin/sh", path)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
