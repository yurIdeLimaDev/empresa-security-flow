package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type OutputAdapter struct {
	Tool           string
	File           string
	ProvenanceFile string
}

func NormalizeKnownOutputs(dir, engagementID, scopeID string, adapters []OutputAdapter) (NormalizedBundle, error) {
	bundles := []NormalizedBundle{}
	errorsFound := []string{}
	provenance, provenanceErrors := loadExecutionProvenance(dir)
	errorsFound = append(errorsFound, provenanceErrors...)
	for _, adapter := range adapters {
		path := filepath.Join(dir, adapter.File)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			errorsFound = append(errorsFound, adapter.Tool+": "+err.Error())
			continue
		}
		cfg := NormalizeConfig{Tool: adapter.Tool, InputPath: path, EngagementID: engagementID, ScopeID: scopeID, SupplyChainStatus: "runner-provenance-missing"}
		provenanceFile := adapter.ProvenanceFile
		if provenanceFile == "" {
			provenanceFile = adapter.File
		}
		provenanceKey := normalizeToolName(adapter.Tool) + "|" + canonicalRepoPath(provenanceFile)
		if item, exists := provenance[provenanceKey]; exists {
			cfg.ToolRunID = item.RunID
			cfg.ToolDigest = item.ToolDigest
			cfg.PolicySHA256 = HashJSON(item.PolicyFilesHash)
			cfg.SupplyChainStatus = "runner-gated"
		} else {
			errorsFound = append(errorsFound, adapter.Tool+": execution log correspondente ao artefato não encontrado")
		}
		bundle, err := NormalizeFile(cfg)
		bundles = append(bundles, bundle)
		if err != nil {
			errorsFound = append(errorsFound, err.Error())
		}
	}
	var result NormalizedBundle
	var err error
	if len(bundles) == 0 {
		result = newNormalizedBundle(engagementID, scopeID)
		result.Engagement.Status = "completed_without_normalizable_output"
		finalizeBundle(&result)
	} else {
		result, err = MergeBundles(bundles...)
		if err != nil {
			return result, err
		}
	}
	if writeErr := WriteBundle(filepath.Join(dir, "normalized-bundle.json"), result); writeErr != nil {
		return result, writeErr
	}
	if writeErr := WriteNormalizedReport(filepath.Join(dir, "normalized-report.md"), result); writeErr != nil {
		return result, writeErr
	}
	if len(errorsFound) > 0 {
		return result, fmt.Errorf("normalização parcial: %v", errorsFound)
	}
	return result, nil
}

func loadExecutionProvenance(dir string) (map[string]ExecutionLog, []string) {
	result := make(map[string]ExecutionLog)
	errorsFound := []string{}
	logDir := filepath.Join(dir, "execution-logs")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if !os.IsNotExist(err) {
			errorsFound = append(errorsFound, "ler execution-logs: "+err.Error())
		}
		return result, errorsFound
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		path := filepath.Join(logDir, entry.Name())
		var log ExecutionLog
		if err := validateJSONSchemaFile("execution-log.schema.json", path); err != nil {
			errorsFound = append(errorsFound, "execution log fora do schema "+entry.Name()+": "+err.Error())
			continue
		}
		if err := ReadJSON(path, &log); err != nil {
			errorsFound = append(errorsFound, "execution log inválido "+entry.Name()+": "+err.Error())
			continue
		}
		for _, artifact := range log.Artifacts {
			key := normalizeToolName(log.Tool) + "|" + canonicalRepoPath(artifact)
			if previous, exists := result[key]; !exists || log.EndedAt > previous.EndedAt {
				result[key] = log
			}
		}
	}
	return result, errorsFound
}
