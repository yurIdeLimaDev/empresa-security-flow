package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func RenderZAPPlan(cfg ScopeConfig, caseDir string) (string, string, error) {
	if cfg.Automation.ZAP == nil {
		return "", "", fmt.Errorf("configuração ZAP ausente")
	}
	zap := cfg.Automation.ZAP
	email, emailOK := os.LookupEnv(zap.UserEmailEnv)
	password, passwordOK := os.LookupEnv(zap.UserPasswordEnv)
	if !emailOK || strings.TrimSpace(email) == "" || !passwordOK || password == "" {
		return "", "", fmt.Errorf("credenciais de teste ZAP ausentes")
	}
	if strings.ContainsAny(email, "\r\n") || strings.ContainsAny(password, "\r\n") {
		return "", "", fmt.Errorf("credencial ZAP contém quebra de linha")
	}
	template, err := os.ReadFile(zap.TemplatePath)
	if err != nil {
		return "", "", err
	}
	login, err := os.ReadFile(zap.LoginScriptPath)
	if err != nil {
		return "", "", err
	}
	include := append([]string{}, zap.IncludePaths...)
	exclude := append([]string{}, zap.ExcludePaths...)
	exclude = append(exclude, strings.TrimRight(cfg.BaseURL, "/")+"/logout.*", strings.TrimRight(cfg.BaseURL, "/")+"/account/delete.*")
	values := map[string]string{
		"{{BASE_URL}}":         yamlString(cfg.BaseURL),
		"{{INCLUDE_PATHS}}":    yamlStringList(include, 10),
		"{{EXCLUDE_PATHS}}":    yamlStringList(dedupeSorted(exclude), 10),
		"{{LOGIN_SCRIPT}}":     yamlString("/work/zap-login.rendered.js"),
		"{{LOGGED_IN_REGEX}}":  yamlString(zap.LoggedInRegex),
		"{{LOGGED_OUT_REGEX}}": yamlString(zap.LoggedOutRegex),
		"{{USER_EMAIL}}":       yamlString(email),
		"{{USER_PASSWORD}}":    yamlString(password),
		"{{SPIDER_MINUTES}}":   strconv.Itoa(zap.SpiderMinutes),
		"{{RULE_MINUTES}}":     strconv.Itoa(zap.RuleMinutes),
		"{{REPORT_FILE}}":      yamlString("/work/zap-report.json"),
	}
	rendered := string(template)
	for token, value := range values {
		if !strings.Contains(rendered, token) {
			return "", "", fmt.Errorf("template ZAP não contém token obrigatório %s", token)
		}
		rendered = strings.ReplaceAll(rendered, token, value)
	}
	if strings.Contains(rendered, "{{") {
		return "", "", fmt.Errorf("template ZAP contém token não resolvido")
	}
	planPath := filepath.Join(caseDir, "zap-plan.rendered.yaml")
	loginPath := filepath.Join(caseDir, "zap-login.rendered.js")
	if err := os.WriteFile(planPath, []byte(rendered), 0o600); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(loginPath, login, 0o600); err != nil {
		_ = os.Remove(planPath)
		return "", "", err
	}
	return planPath, loginPath, nil
}

func yamlString(value string) string { return strconv.Quote(value) }

func yamlStringList(values []string, indent int) string {
	if len(values) == 0 {
		return "[]"
	}
	padding := strings.Repeat(" ", indent)
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, "- "+yamlString(value))
	}
	return "\n" + padding + strings.Join(parts, "\n"+padding)
}
