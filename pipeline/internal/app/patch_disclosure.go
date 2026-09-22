package app

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// A conservative last check of the actual outbound JSON, not a DLP guarantee.
// Reject rather than redact: changing source would invalidate hashes and could
// change the proposed correction. Do not report the match or its source value.
var patchDisclosurePatterns = []*regexp.Regexp{
	regexp.MustCompile(`-----BEGIN (?:[A-Z ]+ )?PRIVATE KEY-----`),
	regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{16,}\b`),
	regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{20,}\b`),
	regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{20,}\b`),
	regexp.MustCompile(`\b(?:AKIA|ASIA)[A-Z0-9]{16}\b`),
	regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`),
	regexp.MustCompile(`(?i)\b(?:password|passwd|api_?key|access_?token|refresh_?token|client_?secret|private_?key|service_?role_?key)["']?\s*[:=]\s*["'][^"'\r\n]{4,}["']`),
	regexp.MustCompile(`(?i)\b(?:https?|postgres(?:ql)?|mysql|mongodb(?:\+srv)?|redis)://[^\s/"'<>:@]+:[^\s/"'<>@]+@`),
	regexp.MustCompile(`(?i)\b(?:authorization\s*[:=]\s*["']?\s*)?(?:bearer|basic)\s+[A-Za-z0-9_+/.=-]{8,}`),
}

func validatePatchDisclosure(data []byte, credential string) error {
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("requisição de geração inválida antes do envio")
	}
	var inspect func(any) bool
	inspect = func(value any) bool {
		switch value := value.(type) {
		case string:
			if credential != "" && strings.Contains(value, credential) {
				return true
			}
			for _, pattern := range patchDisclosurePatterns {
				if pattern.MatchString(value) {
					return true
				}
			}
		case []any:
			for _, item := range value {
				if inspect(item) {
					return true
				}
			}
		case map[string]any:
			for key, item := range value {
				if inspect(key) || inspect(item) {
					return true
				}
			}
		}
		return false
	}
	if inspect(payload) {
		return fmt.Errorf("envio bloqueado: possível credencial no contexto; sanear fontes e metadados antes de tentar novamente")
	}
	return nil
}
