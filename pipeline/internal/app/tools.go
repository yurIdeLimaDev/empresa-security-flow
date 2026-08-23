package app

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

type ToolStatus struct {
	Name    string
	State   string
	Version string
}

type toolDefinition struct {
	Name        string
	Executables []string
	VersionArgs []string
}

var toolDefinitions = []toolDefinition{
	{Name: "subfinder", Executables: []string{"subfinder"}, VersionArgs: []string{"-version"}},
	{Name: "amass", Executables: []string{"amass"}, VersionArgs: []string{"-version"}},
	{Name: "httpx", Executables: []string{"httpx"}, VersionArgs: []string{"-version"}},
	{Name: "wappalyzer", Executables: []string{"wappalyzer"}, VersionArgs: []string{"--version"}},
	{Name: "ffuf", Executables: []string{"ffuf"}, VersionArgs: []string{"-V"}},
	{Name: "feroxbuster", Executables: []string{"feroxbuster"}, VersionArgs: []string{"--version"}},
	{Name: "nuclei", Executables: []string{"nuclei"}, VersionArgs: []string{"-version"}},
	{Name: "zap", Executables: []string{"zap", "zap.sh", "zap.bat"}, VersionArgs: []string{"-version"}},
	{Name: "gitleaks", Executables: []string{"gitleaks"}, VersionArgs: []string{"version"}},
	{Name: "trufflehog", Executables: []string{"trufflehog"}, VersionArgs: []string{"--version"}},
	{Name: "testssl", Executables: []string{"testssl.sh", "testssl"}, VersionArgs: []string{"--version"}},
	{Name: "schemathesis", Executables: []string{"schemathesis"}, VersionArgs: []string{"--version"}},
	{Name: "semgrep", Executables: []string{"semgrep"}, VersionArgs: []string{"--version"}},
	{Name: "trivy", Executables: []string{"trivy"}, VersionArgs: []string{"--version"}},
	{Name: "osv-scanner", Executables: []string{"osv-scanner"}, VersionArgs: []string{"--version"}},
	{Name: "jwt_tool", Executables: []string{"jwt_tool", "jwt_tool.py"}, VersionArgs: []string{"-h"}},
	{Name: "sqlmap", Executables: []string{"sqlmap", "sqlmap.py"}, VersionArgs: []string{"--version"}},
	{Name: "dalfox", Executables: []string{"dalfox"}, VersionArgs: []string{"version"}},
	{Name: "jsluice", Executables: []string{"jsluice"}, VersionArgs: []string{"--help"}},
	{Name: "hadrian", Executables: []string{"hadrian"}, VersionArgs: []string{"version"}},
	{Name: "docker", Executables: []string{"docker"}, VersionArgs: []string{"--version"}},
}

func ToolStatuses(timeout time.Duration) []ToolStatus {
	statuses := make([]ToolStatus, 0, len(toolDefinitions))
	for _, definition := range toolDefinitions {
		path := ""
		for _, executable := range definition.Executables {
			if found, err := exec.LookPath(executable); err == nil {
				path = found
				break
			}
		}
		if path == "" {
			statuses = append(statuses, ToolStatus{Name: definition.Name, State: "missing"})
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		command := exec.CommandContext(ctx, path, definition.VersionArgs...)
		output, _ := command.CombinedOutput()
		cancel()
		version := strings.TrimSpace(string(bytes.ReplaceAll(output, []byte("\r"), nil)))
		if len(version) > 160 {
			version = version[:160]
		}
		statuses = append(statuses, ToolStatus{Name: definition.Name, State: "available", Version: version})
	}
	return statuses
}

func executableFor(name string) (string, bool) {
	for _, definition := range toolDefinitions {
		if definition.Name != name {
			continue
		}
		for _, executable := range definition.Executables {
			if path, err := exec.LookPath(executable); err == nil {
				return path, true
			}
		}
	}
	return "", false
}
