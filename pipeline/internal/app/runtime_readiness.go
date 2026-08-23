package app

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type RuntimeReadinessCheck struct {
	Tool             string `json:"tool"`
	ExecutionClass   string `json:"execution_class"`
	Status           string `json:"status"`
	Image            string `json:"image,omitempty"`
	ImageID          string `json:"image_id,omitempty"`
	ArtifactDigest   string `json:"artifact_digest,omitempty"`
	ObservedOutput   string `json:"observed_output,omitempty"`
	ExitCode         int    `json:"exit_code,omitempty"`
	DurationMS       int64  `json:"duration_ms,omitempty"`
	Reason           string `json:"reason,omitempty"`
	HardeningProfile string `json:"hardening_profile,omitempty"`
}

type RuntimeReadinessReport struct {
	SchemaVersion string                  `json:"schema_version"`
	CheckedAt     string                  `json:"checked_at"`
	DockerVersion string                  `json:"docker_version"`
	PullRequested bool                    `json:"pull_requested"`
	Approved      bool                    `json:"approved"`
	Checks        []RuntimeReadinessCheck `json:"checks"`
}

func CheckRuntimeReadiness(ctx context.Context, lock ToolLockFile, sbomPath string, pull bool) (RuntimeReadinessReport, error) {
	report := RuntimeReadinessReport{SchemaVersion: "1.0.0", CheckedAt: now(), PullRequested: pull, Approved: true, Checks: []RuntimeReadinessCheck{}}
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return report, fmt.Errorf("Docker não instalado")
	}
	versionCtx, versionCancel := context.WithTimeout(ctx, 15*time.Second)
	versionOutput, versionErr := exec.CommandContext(versionCtx, dockerPath, "version", "--format", "{{.Server.Version}}").CombinedOutput()
	versionCancel()
	if versionErr != nil {
		return report, fmt.Errorf("Docker Engine indisponível: %w", versionErr)
	}
	report.DockerVersion = strings.TrimSpace(string(versionOutput))

	supply := ValidateToolLock(lock, sbomPath)
	if !supply.Approved {
		return report, fmt.Errorf("cadeia de ferramentas não aprovada")
	}
	pulled := map[string]error{}
	for _, tool := range lock.Tools {
		check := RuntimeReadinessCheck{Tool: tool.Name, ExecutionClass: tool.ExecutionClass, ExitCode: -1}
		if tool.ExecutionClass != "automatic" {
			check.Status = "restricted"
			check.Reason = "não executável por decisão de adoção"
			report.Checks = append(report.Checks, check)
			continue
		}
		if tool.Container == nil || tool.Liveness == nil {
			check.Status, check.Reason = "failed", "contrato de runtime incompleto"
			report.Approved = false
			report.Checks = append(report.Checks, check)
			continue
		}
		image := tool.Container.Image + "@" + normalizeDigest(tool.Container.Digest)
		check.Image = image
		check.ArtifactDigest = toolExecutionDigest(tool)
		if pull {
			pullErr, seen := pulled[image]
			if !seen {
				pullCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
				pullErr = exec.CommandContext(pullCtx, dockerPath, "pull", image).Run()
				cancel()
				pulled[image] = pullErr
			}
			if pullErr != nil {
				check.Status, check.Reason = "failed", "pull por digest falhou: "+pullErr.Error()
				report.Approved = false
				report.Checks = append(report.Checks, check)
				continue
			}
		}
		inspectCtx, inspectCancel := context.WithTimeout(ctx, 30*time.Second)
		imageOutput, inspectErr := exec.CommandContext(inspectCtx, dockerPath, "image", "inspect", image, "--format", "{{.Id}}").CombinedOutput()
		inspectCancel()
		if inspectErr != nil {
			check.Status, check.Reason = "failed", "imagem exata ausente: "+strings.TrimSpace(string(imageOutput))
			report.Approved = false
			report.Checks = append(report.Checks, check)
			continue
		}
		check.ImageID = strings.TrimSpace(string(imageOutput))

		started := time.Now()
		name := containerNameForRun("liveness_" + normalizeToolName(tool.Name) + "_" + HashBytes([]byte(started.String()))[:8])
		args, argsErr := runtimeCheckArguments(tool, name)
		if argsErr != nil {
			check.Status, check.Reason = "failed", argsErr.Error()
			report.Approved = false
			report.Checks = append(report.Checks, check)
			continue
		}
		runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		command := exec.CommandContext(runCtx, dockerPath, args...)
		stdout := cappedBuffer{limit: 1 << 20}
		stderr := cappedBuffer{limit: 1 << 20}
		command.Stdout, command.Stderr = &stdout, &stderr
		runErr := command.Run()
		timedOut := runCtx.Err() == context.DeadlineExceeded
		cancel()
		if timedOut {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
			_ = exec.CommandContext(cleanupCtx, dockerPath, "rm", "-f", name).Run()
			cleanupCancel()
		}
		check.DurationMS = time.Since(started).Milliseconds()
		combined := strings.TrimSpace(strings.ReplaceAll(string(append(stdout.Bytes(), stderr.Bytes()...)), "\r", ""))
		if len(combined) > 4096 {
			combined = combined[:4096]
		}
		check.ObservedOutput = combined
		exitCode := 0
		if runErr != nil {
			exitCode = -1
			if exitErr, ok := runErr.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}
		check.ExitCode = exitCode
		allowed := exitCode == 0 || containsInt(tool.Liveness.AllowedExitCodes, exitCode)
		pattern, patternErr := regexp.Compile(tool.Liveness.ExpectedPattern)
		switch {
		case timedOut:
			check.Status, check.Reason = "failed", "liveness excedeu dois minutos"
		case runErr != nil && !allowed:
			check.Status, check.Reason = "failed", runErr.Error()
		case patternErr != nil || !pattern.MatchString(combined):
			check.Status, check.Reason = "failed", "saída não corresponde ao expected_pattern pinado"
		default:
			check.Status = "passed"
			check.HardeningProfile = "network=none,cap-drop=ALL,no-new-privileges,read-only,tmpfs,resource-limits"
		}
		if check.Status != "passed" {
			report.Approved = false
		}
		report.Checks = append(report.Checks, check)
	}
	sort.Slice(report.Checks, func(i, j int) bool { return report.Checks[i].Tool < report.Checks[j].Tool })
	return report, nil
}

func runtimeCheckArguments(tool ToolLock, containerName string) ([]string, error) {
	if tool.Container == nil || tool.Liveness == nil {
		return nil, fmt.Errorf("runtime incompleto")
	}
	args := []string{
		"run", "--rm", "--pull=never", "--name", containerName,
		"--hostname", containerName, "--add-host", containerName + ":127.0.0.1",
		"--network", "none", "--cap-drop", "ALL", "--security-opt", "no-new-privileges=true",
		"--read-only", "--init", "--pids-limit", "256", "--memory", "1g", "--memory-swap", "1g", "--cpus", "1",
		"--ulimit", "nofile=2048:2048", "--stop-timeout", "10",
		"--log-driver", "local", "--log-opt", "max-size=1m", "--log-opt", "max-file=2",
		"--tmpfs", "/tmp:rw,noexec,nosuid,nodev,size=512m,mode=1777", "--env", "HOME=/tmp", "--env", "TMPDIR=/tmp",
	}
	if normalizeToolName(tool.Name) == "zap" {
		args = append(args, "--tmpfs", "/home/zap:rw,noexec,nosuid,nodev,size=512m,mode=1777")
	}
	environmentNames := make([]string, 0, len(tool.Container.Environment))
	for name := range tool.Container.Environment {
		environmentNames = append(environmentNames, name)
	}
	sort.Strings(environmentNames)
	for _, name := range environmentNames {
		args = append(args, "--env", name+"="+tool.Container.Environment[name])
	}
	for _, artifact := range tool.Container.RuntimeArtifacts {
		path, err := resolveSupplyAsset(artifact.Path)
		if err != nil {
			return nil, err
		}
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		args = append(args, "--mount", "type=bind,src="+absolute+",dst="+artifact.MountPath+",readonly")
	}
	if tool.Container.Entrypoint != "" {
		args = append(args, "--entrypoint", tool.Container.Entrypoint)
	}
	args = append(args, tool.Container.Image+"@"+normalizeDigest(tool.Container.Digest))
	args = append(args, tool.Container.CommandPrefix...)
	args = append(args, tool.Liveness.Command...)
	return args, nil
}
