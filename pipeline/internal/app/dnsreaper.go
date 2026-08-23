package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var hostnamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)+$`)

func CurateDNSReaperCandidates(domain string, paths ...string) ([]string, error) {
	domain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
	seen := map[string]struct{}{}
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 64<<10), 1<<20)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			candidate := line
			if strings.HasPrefix(line, "{") {
				var row map[string]any
				if json.Unmarshal([]byte(line), &row) == nil {
					for _, key := range []string{"host", "name", "domain", "input"} {
						if value, ok := row[key].(string); ok && value != "" {
							candidate = value
							break
						}
					}
				}
			}
			candidate = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(candidate), "."))
			if candidate == domain || !strings.HasSuffix(candidate, "."+domain) || strings.ContainsAny(candidate, "*/:@ ") || !hostnamePattern.MatchString(candidate) {
				continue
			}
			seen[candidate] = struct{}{}
		}
		scanErr := scanner.Err()
		_ = file.Close()
		if scanErr != nil {
			return nil, scanErr
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result, nil
}

func WriteLines(path string, lines []string) error {
	content := ""
	if len(lines) > 0 {
		content = strings.Join(lines, "\n") + "\n"
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

type DNSReaperCandidate struct {
	Domain       string   `json:"domain"`
	Signature    string   `json:"signature"`
	Info         string   `json:"info"`
	Confidence   string   `json:"confidence"`
	ARecords     []string `json:"a_records"`
	AAAARecords  []string `json:"aaaa_records"`
	CNAMERecords []string `json:"cname_records"`
	NSRecords    []string `json:"ns_records"`
	MoreInfoURL  string   `json:"more_info_url"`
	Status       string   `json:"status"`
	Limitation   string   `json:"limitation"`
}

func PostprocessDNSReaper(inputPath, candidatePath, unlikelyPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	var raw []struct {
		Domain       string   `json:"domain"`
		Signature    string   `json:"signature"`
		Info         string   `json:"info"`
		Confidence   string   `json:"confidence"`
		ARecords     []string `json:"a_records"`
		AAAARecords  []string `json:"aaaa_records"`
		CNAMERecords []string `json:"cname_records"`
		NSRecords    []string `json:"ns_records"`
		MoreInfoURL  string   `json:"more_info_url"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("resultado dnsReaper inválido: %w", err)
	}
	candidates, unlikely := []DNSReaperCandidate{}, []DNSReaperCandidate{}
	for _, row := range raw {
		item := DNSReaperCandidate{Domain: row.Domain, Signature: row.Signature, Info: row.Info, Confidence: row.Confidence, ARecords: row.ARecords, AAAARecords: row.AAAARecords, CNAMERecords: row.CNAMERecords, NSRecords: row.NSRecords, MoreInfoURL: row.MoreInfoURL, Status: "candidate", Limitation: "dnsReaper gera candidato; takeover não foi confirmado e nenhum recurso foi reivindicado."}
		if strings.EqualFold(row.Confidence, "UNLIKELY") {
			unlikely = append(unlikely, item)
		} else {
			candidates = append(candidates, item)
		}
	}
	if err := WriteJSON(candidatePath, candidates); err != nil {
		return err
	}
	return WriteJSON(unlikelyPath, unlikely)
}
