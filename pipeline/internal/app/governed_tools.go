package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type PathHash struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

type ExecutionNetworkLog struct {
	Bridge   string `json:"bridge"`
	Chain    string `json:"chain"`
	ProxyLog string `json:"proxy_log"`
}

type ExecutionCounts struct {
	Requests int `json:"requests"`
}

type ExecutionLog struct {
	RunID            string              `json:"run_id"`
	EngagementID     string              `json:"engagement_id"`
	Tool             string              `json:"tool"`
	ToolDigest       string              `json:"tool_digest"`
	ContainerDigest  string              `json:"container_digest"`
	PolicyFilesHash  PolicyHashes        `json:"policy_files_hash"`
	InputHashes      []PathHash          `json:"input_hashes"`
	RuntimeArtifacts []PathHash          `json:"runtime_artifacts"`
	Network          ExecutionNetworkLog `json:"network"`
	StartedAt        string              `json:"started_at"`
	EndedAt          string              `json:"ended_at"`
	ExitCode         int                 `json:"exit_code"`
	TimedOut         bool                `json:"timed_out"`
	StdoutTruncated  bool                `json:"stdout_truncated"`
	StderrTruncated  bool                `json:"stderr_truncated"`
	Counts           ExecutionCounts     `json:"counts"`
	Artifacts        []string            `json:"artifacts"`
	ArtifactHashes   []PathHash          `json:"artifact_hashes"`
	ReviewedBy       *string             `json:"reviewed_by"`
	ReviewedAt       *string             `json:"reviewed_at"`
}

const maxExecutionLogBytes = 16 << 20

type cappedBuffer struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func (b *cappedBuffer) Write(data []byte) (int, error) {
	written := len(data)
	remaining := b.limit - b.buffer.Len()
	if remaining > 0 {
		if remaining > len(data) {
			remaining = len(data)
		}
		_, _ = b.buffer.Write(data[:remaining])
	}
	if remaining < len(data) {
		b.truncated = true
	}
	return written, nil
}

func (b *cappedBuffer) Bytes() []byte { return b.buffer.Bytes() }

type GovernedRunOptions struct {
	Pipeline int
	Policies ValidatedPolicies
	Lock     ToolLockFile
	SBOMPath string
	Network  NetworkState
	CaseDir  string
}

type GateError struct{ Reason string }

func (e *GateError) Error() string { return e.Reason }

func gateBlock(action *PlannedAction, reason string) error {
	action.State, action.Reason = "blocked", reason
	return &GateError{Reason: reason}
}

func IsGateError(err error) bool {
	_, ok := err.(*GateError)
	return ok
}

func RunGovernedTool(ctx context.Context, options GovernedRunOptions, action *PlannedAction) error {
	policyName := action.PolicyTool
	if policyName == "" {
		policyName = normalizeToolName(action.Tool)
	}
	toolPolicy, err := options.Policies.Tool(policyName, options.Pipeline)
	if err != nil {
		return gateBlock(action, err.Error())
	}
	if toolPolicy.Mode != "automatic" {
		return gateBlock(action, fmt.Sprintf("tool-policy define modo %s, não automatic", toolPolicy.Mode))
	}
	if toolPolicy.MaxDurationMinutes < 1 {
		return gateBlock(action, "max_duration_minutes ausente")
	}
	if err := enforceActionGates(options, *action, toolPolicy); err != nil {
		return gateBlock(action, err.Error())
	}
	lockName := normalizeToolName(policyName)
	lockEntry, err := ApproveContainer(options.Lock, lockName, options.SBOMPath)
	if err != nil {
		return gateBlock(action, err.Error())
	}
	if len(action.Command) == 0 {
		return fmt.Errorf("ação %s sem argumentos", action.ID)
	}
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return gateBlock(action, "Docker não instalado no host Linux")
	}
	inputHashes, err := hashInputs(action.InputPaths)
	if err != nil {
		return gateBlock(action, "hash de insumo: "+err.Error())
	}
	started := time.Now().UTC()
	runID := "run_" + started.Format("20060102t150405000000000") + "_" + HashBytes([]byte(action.ID + started.String()))[:8]
	containerName := containerNameForRun(runID)
	args, err := buildDockerArguments(options, *action, *lockEntry, runID)
	if err != nil {
		return gateBlock(action, err.Error())
	}
	toolCtx, cancel := context.WithTimeout(ctx, time.Duration(toolPolicy.MaxDurationMinutes)*time.Minute)
	defer cancel()
	command := exec.CommandContext(toolCtx, dockerPath, args...)
	command.Dir = options.CaseDir
	command.Env = os.Environ()
	stdout := cappedBuffer{limit: maxExecutionLogBytes}
	stderr := cappedBuffer{limit: maxExecutionLogBytes}
	command.Stdout = &stdout
	command.Stderr = &stderr
	runErr := command.Run()
	timedOut := toolCtx.Err() == context.DeadlineExceeded
	if timedOut {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		_ = exec.CommandContext(cleanupCtx, dockerPath, "rm", "-f", containerName).Run()
		cleanupCancel()
	}
	ended := time.Now().UTC()
	stdoutPath := filepath.Join(options.CaseDir, action.ID+".stdout.log")
	stderrPath := filepath.Join(options.CaseDir, action.ID+".stderr.log")
	_ = os.WriteFile(stdoutPath, stdout.Bytes(), 0o600)
	_ = os.WriteFile(stderrPath, stderr.Bytes(), 0o600)
	exitCode := 0
	if runErr != nil {
		exitCode = -1
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	action.ExitCode = exitCode
	allowed := exitCode == 0 || containsInt(action.AllowedExitCodes, exitCode)
	if timedOut {
		action.State = "failed"
		action.Reason = fmt.Sprintf("tempo limite de %d minuto(s) excedido", toolPolicy.MaxDurationMinutes)
	} else if runErr != nil && !allowed {
		action.State = "failed"
		action.Reason = runErr.Error()
	} else {
		action.State = "completed"
		if exitCode != 0 {
			action.Reason = fmt.Sprintf("exit code %d significa candidato, não falha de execução", exitCode)
		}
	}
	artifactHashes := []PathHash{}
	if action.State == "completed" {
		artifactHashes, err = validateArtifactOutputs(options.CaseDir, action.ArtifactPaths)
		if err != nil {
			action.State = "failed"
			action.Reason = "contrato de artefatos: " + err.Error()
			runErr = fmt.Errorf("%s", action.Reason)
			allowed = false
		}
	}
	runtimeHashes, err := runtimeArtifactHashes(*lockEntry)
	if err != nil {
		return gateBlock(action, "runtime artifact: "+err.Error())
	}
	artifacts := append([]string{}, action.ArtifactPaths...)
	artifacts = append(artifacts, stdoutPath, stderrPath)
	logHashes, logHashErr := validateArtifactOutputs(options.CaseDir, []string{stdoutPath, stderrPath})
	if logHashErr == nil {
		artifactHashes = append(artifactHashes, logHashes...)
	}
	log := ExecutionLog{
		RunID: runID, EngagementID: options.Policies.Engagement.EngagementID,
		Tool: policyName, ToolDigest: toolExecutionDigest(*lockEntry), ContainerDigest: normalizeDigest(lockEntry.Container.Digest),
		PolicyFilesHash: options.Policies.Hashes, InputHashes: inputHashes, RuntimeArtifacts: runtimeHashes,
		Network:   ExecutionNetworkLog{Bridge: options.Network.NetworkName, Chain: options.Network.Chain, ProxyLog: options.Network.ProxyLog},
		StartedAt: started.Format(time.RFC3339Nano), EndedAt: ended.Format(time.RFC3339Nano), ExitCode: exitCode,
		TimedOut: timedOut, StdoutTruncated: stdout.truncated, StderrTruncated: stderr.truncated,
		Counts:    ExecutionCounts{Requests: countRequestLogLines(options.Network.ProxyLog, runID)},
		Artifacts: relativeArtifacts(options.CaseDir, artifacts), ArtifactHashes: artifactHashes, ReviewedBy: nil, ReviewedAt: nil,
	}
	logDir := filepath.Join(options.CaseDir, "execution-logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return err
	}
	logPath := filepath.Join(logDir, runID+".json")
	if err := WriteJSON(logPath, log); err != nil {
		return err
	}
	if err := validateJSONSchemaFile("execution-log.schema.json", logPath); err != nil {
		return fmt.Errorf("execution log interno inválido: %w", err)
	}
	for _, path := range action.EphemeralPaths {
		root, rootErr := filepath.Abs(options.CaseDir)
		if abs, absErr := filepath.Abs(path); absErr == nil && rootErr == nil && pathWithin(root, abs) {
			_ = os.Remove(abs)
		}
	}
	if runErr != nil && (!allowed || timedOut) {
		return fmt.Errorf("%s: %w", action.ID, runErr)
	}
	return nil
}

func enforceActionGates(options GovernedRunOptions, action PlannedAction, policy ToolPolicyEntry) error {
	if options.Pipeline == 2 && (options.Policies.Engagement.Authorization == nil || !options.Policies.Engagement.Authorization.Confirmed) {
		return fmt.Errorf("Pipeline 2 sem autorização confirmada")
	}
	if action.PolicyTool == "feroxbuster" || action.Tool == "feroxbuster" {
		target, ok := findPolicyTarget(options.Policies.Engagement, firstURLArgument(action.Command))
		if !ok || target.RatePerSecond <= 0 {
			return fmt.Errorf("feroxbuster exige rate_per_second no alvo")
		}
	}
	if (action.PolicyTool == "amass_active" || action.ID == "01-amass-active") && (options.Policies.Engagement.Authorization == nil || !options.Policies.Engagement.Authorization.Confirmed) {
		return fmt.Errorf("amass ativo exige authorization.confirmed")
	}
	if options.Pipeline == 1 && action.Tool == "trufflehog" && !containsFold(action.Command, "--no-verification") {
		return fmt.Errorf("runner recusou TruffleHog P1 sem --no-verification")
	}
	if action.PolicyTool == "hibp" && policy.Flags.DomainSearch {
		return fmt.Errorf("HIBP domain_search é proibido")
	}
	if policy.Flags.ClaimResource || policy.Flags.AutoRegisterAccounts || policy.Flags.ApplyChanges {
		return fmt.Errorf("proibição absoluta habilitada")
	}
	return nil
}

func buildDockerArguments(options GovernedRunOptions, action PlannedAction, lock ToolLock, runID string) ([]string, error) {
	caseAbs, err := filepath.Abs(options.CaseDir)
	if err != nil {
		return nil, err
	}
	args := []string{
		"run", "--rm", "--pull=never", "--name", containerNameForRun(runID),
		"--label", "empresa-security.managed=true", "--label", "empresa-security.engagement=" + options.Policies.Engagement.EngagementID,
		"--label", "empresa-security.run-id=" + runID,
		"--hostname", containerNameForRun(runID), "--add-host", containerNameForRun(runID) + ":127.0.0.1",
		"--network", options.Network.NetworkName, "--dns", options.Policies.Engagement.Environment.DNSResolverIP,
		"--sysctl", "net.ipv6.conf.all.disable_ipv6=1",
		"--cap-drop", "ALL", "--security-opt", "no-new-privileges=true", "--read-only", "--init",
		"--pids-limit", "512", "--memory", "2g", "--memory-swap", "2g", "--cpus", "2",
		"--ulimit", "nofile=4096:4096", "--stop-timeout", "10",
		"--log-driver", "local", "--log-opt", "max-size=16m", "--log-opt", "max-file=2",
		"--tmpfs", "/tmp:rw,noexec,nosuid,nodev,size=1g,mode=1777",
		"--env", "HOME=/tmp", "--env", "TMPDIR=/tmp",
		"--mount", "type=bind,src=" + caseAbs + ",dst=/work",
		"-w", "/work",
	}
	if normalizeToolName(lock.Name) == "zap" {
		args = append(args, "--tmpfs", "/home/zap:rw,noexec,nosuid,nodev,size=512m,mode=1777")
	}
	environmentNames := make([]string, 0, len(lock.Container.Environment))
	for name := range lock.Container.Environment {
		environmentNames = append(environmentNames, name)
	}
	sort.Strings(environmentNames)
	for _, name := range environmentNames {
		args = append(args, "--env", name+"="+lock.Container.Environment[name])
	}
	if options.Network.HTTPProxyURL != "" {
		proxyURL := proxyURLForRun(options.Network.HTTPProxyURL, runID)
		args = append(args, "--env", "HTTP_PROXY="+proxyURL, "--env", "HTTPS_PROXY="+proxyURL, "--env", "NO_PROXY=")
	}
	mappings := map[string]string{}
	for index, input := range action.InputPaths {
		inputAbs, absErr := filepath.Abs(input)
		if absErr != nil {
			return nil, absErr
		}
		if pathWithin(caseAbs, inputAbs) {
			rel, _ := filepath.Rel(caseAbs, inputAbs)
			mappings[input] = "/work/" + filepath.ToSlash(rel)
			mappings[inputAbs] = mappings[input]
			continue
		}
		containerPath := fmt.Sprintf("/inputs/%d", index)
		args = append(args, "--mount", "type=bind,src="+inputAbs+",dst="+containerPath+",readonly")
		mappings[input] = containerPath
		mappings[inputAbs] = containerPath
	}
	for _, secret := range action.SecretArgsFromEnv {
		value, exists := os.LookupEnv(secret.Env)
		if !exists || strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("variável de ambiente ausente: %s", secret.Env)
		}
		args = append(args, "--env", secret.Env)
	}
	for _, artifact := range lock.Container.RuntimeArtifacts {
		artifactPath, resolveErr := resolveSupplyAsset(artifact.Path)
		if resolveErr != nil {
			return nil, resolveErr
		}
		artifactAbs, absErr := filepath.Abs(artifactPath)
		if absErr != nil {
			return nil, absErr
		}
		args = append(args, "--mount", "type=bind,src="+artifactAbs+",dst="+artifact.MountPath+",readonly")
	}
	if lock.Container.Entrypoint != "" {
		args = append(args, "--entrypoint", lock.Container.Entrypoint)
	}
	image := lock.Container.Image + "@" + normalizeDigest(lock.Container.Digest)
	args = append(args, image)
	args = append(args, lock.Container.CommandPrefix...)
	for _, secret := range action.SecretArgsFromEnv {
		if secret.Position != "prepend" {
			continue
		}
		if secret.Flag != "" {
			args = append(args, secret.Flag)
		}
		args = append(args, os.Getenv(secret.Env))
	}
	for _, argument := range action.Command {
		translated := translateActionPath(argument, caseAbs, mappings)
		args = append(args, translated)
	}
	for _, secret := range action.SecretArgsFromEnv {
		if secret.Position == "prepend" || secret.Position == "env_only" {
			continue
		}
		value := os.Getenv(secret.Env)
		if secret.Flag != "" {
			args = append(args, secret.Flag)
		}
		args = append(args, value)
	}
	return args, nil
}

func containerNameForRun(runID string) string {
	value := strings.ToLower(strings.ReplaceAll(runID, "_", "-"))
	if len(value) > 63 {
		value = value[:63]
	}
	return "es-" + value
}

func runtimeArtifactHashes(lock ToolLock) ([]PathHash, error) {
	result := []PathHash{}
	if lock.Container == nil {
		return result, nil
	}
	for _, artifact := range lock.Container.RuntimeArtifacts {
		path, err := resolveSupplyAsset(artifact.Path)
		if err != nil {
			return nil, err
		}
		value, err := hashPathWithSize(path)
		if err != nil {
			return nil, err
		}
		value.Path = artifact.Path
		result = append(result, value)
	}
	return result, nil
}

func toolExecutionDigest(lock ToolLock) string {
	parts := []string{normalizeDigest(lock.Container.Digest)}
	for _, artifact := range lock.Container.RuntimeArtifacts {
		parts = append(parts, artifact.MountPath+"="+normalizeDigest(artifact.SHA256))
	}
	sort.Strings(parts[1:])
	return "sha256:" + HashBytes([]byte(strings.Join(parts, "\n")))
}

func validateArtifactOutputs(caseDir string, paths []string) ([]PathHash, error) {
	root, err := filepath.Abs(caseDir)
	if err != nil {
		return nil, err
	}
	result := make([]PathHash, 0, len(paths))
	for _, path := range paths {
		abs, absErr := filepath.Abs(path)
		if absErr != nil || !pathWithin(root, abs) {
			return nil, fmt.Errorf("artefato fora do diretório do caso: %s", path)
		}
		value, hashErr := hashPathWithSize(abs)
		if hashErr != nil {
			return nil, fmt.Errorf("artefato ausente ou ilegível %s: %w", path, hashErr)
		}
		rel, _ := filepath.Rel(root, abs)
		value.Path = filepath.ToSlash(rel)
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

func hashPathWithSize(path string) (PathHash, error) {
	hash, err := HashPath(path)
	if err != nil {
		return PathHash{}, err
	}
	var size int64
	err = filepath.Walk(path, func(_ string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return PathHash{}, err
	}
	return PathHash{Path: path, SHA256: "sha256:" + hash, SizeBytes: size}, nil
}

func proxyURLForRun(raw, runID string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	parsed.User = url.User(runID)
	return parsed.String()
}

func translateActionPath(argument, caseAbs string, mappings map[string]string) string {
	if replacement, ok := mappings[argument]; ok {
		return replacement
	}
	if abs, err := filepath.Abs(argument); err == nil {
		if replacement, ok := mappings[abs]; ok {
			return replacement
		}
		if pathWithin(caseAbs, abs) {
			rel, _ := filepath.Rel(caseAbs, abs)
			return "/work/" + filepath.ToSlash(rel)
		}
	}
	return argument
}

func pathWithin(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func hashInputs(paths []string) ([]PathHash, error) {
	result := make([]PathHash, 0, len(paths))
	for _, path := range paths {
		hash, err := HashPath(path)
		if err != nil {
			return nil, err
		}
		result = append(result, PathHash{Path: path, SHA256: "sha256:" + hash})
	}
	return result, nil
}

func HashPath(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("symlink não é aceito como insumo: %s", path)
	}
	if !info.IsDir() {
		return HashFile(path)
	}
	root, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	files := []string{}
	err = filepath.Walk(root, func(current string, entry os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink não é aceito como insumo: %s", current)
		}
		if !entry.IsDir() {
			files = append(files, current)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	h := sha256.New()
	for _, file := range files {
		rel, _ := filepath.Rel(root, file)
		_, _ = io.WriteString(h, filepath.ToSlash(rel)+"\x00")
		f, openErr := os.Open(file)
		if openErr != nil {
			return "", openErr
		}
		_, copyErr := io.Copy(h, f)
		closeErr := f.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		_, _ = io.WriteString(h, "\x00")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func relativeArtifacts(caseDir string, paths []string) []string {
	result := []string{}
	seen := map[string]struct{}{}
	for _, path := range paths {
		if path == "" {
			continue
		}
		value := path
		if abs, err := filepath.Abs(path); err == nil {
			if root, rootErr := filepath.Abs(caseDir); rootErr == nil && pathWithin(root, abs) {
				if rel, relErr := filepath.Rel(root, abs); relErr == nil {
					value = filepath.ToSlash(rel)
				}
			}
		}
		if _, ok := seen[value]; !ok {
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}

func countRequestLogLines(path, runID string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	count := 0
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		if bytes.Contains(line, []byte(`"tool_run_id":"`+runID+`"`)) {
			count++
		}
	}
	return count
}

func firstURLArgument(args []string) string {
	for _, item := range args {
		if strings.HasPrefix(item, "http://") || strings.HasPrefix(item, "https://") {
			return item
		}
	}
	return ""
}

func findPolicyTarget(policy EngagementPolicy, raw string) (NetworkTarget, bool) {
	for _, target := range policy.Environment.Targets {
		if raw == "" {
			return target, true
		}
		if u, err := urlParse(raw); err == nil && strings.EqualFold(u.Hostname(), target.Host) {
			return target, true
		}
	}
	return NetworkTarget{}, false
}

func urlParse(raw string) (*url.URL, error) { return url.Parse(raw) }

func normalizeDigest(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if !strings.HasPrefix(value, "sha256:") {
		value = "sha256:" + value
	}
	return value
}
