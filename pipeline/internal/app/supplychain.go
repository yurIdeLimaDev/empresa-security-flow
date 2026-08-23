package app

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const ToolLockSchemaVersion = "1.1.0"

var sha256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var commitPattern = regexp.MustCompile(`^[a-f0-9]{40}$`)
var environmentNamePattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

type ToolLockFile struct {
	SchemaVersion string            `json:"schema_version"`
	GeneratedAt   string            `json:"generated_at"`
	Policy        SupplyChainPolicy `json:"policy"`
	Tools         []ToolLock        `json:"tools"`
}

type SupplyChainPolicy struct {
	DenyLatest             bool `json:"deny_latest"`
	RequirePinnedVersion   bool `json:"require_pinned_version"`
	RequireCommit          bool `json:"require_commit"`
	RequireArtifactSHA256  bool `json:"require_artifact_sha256"`
	RequireContainerDigest bool `json:"require_container_digest"`
	RequireCodeReview      bool `json:"require_code_review"`
	RequireSBOM            bool `json:"require_sbom"`
}

type ReleaseLock struct {
	Tag               string `json:"tag"`
	ReleaseID         int64  `json:"release_id"`
	PublishedAt       string `json:"published_at"`
	Commit            string `json:"commit"`
	TagType           string `json:"tag_type"`
	SignatureStatus   string `json:"signature_status"`
	Artifact          string `json:"artifact,omitempty"`
	ArtifactSHA256    string `json:"artifact_sha256,omitempty"`
	VerificationNotes string `json:"verification_notes"`
}

type ContainerLock struct {
	Image            string                `json:"image"`
	Digest           string                `json:"digest"`
	Entrypoint       string                `json:"entrypoint,omitempty"`
	CommandPrefix    []string              `json:"command_prefix,omitempty"`
	Environment      map[string]string     `json:"environment,omitempty"`
	RuntimeArtifacts []RuntimeArtifactLock `json:"runtime_artifacts,omitempty"`
}

type RuntimeArtifactLock struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	MountPath string `json:"mount_path"`
}

type CodeReviewLock struct {
	Status          string `json:"status"`
	Reviewer        string `json:"reviewer,omitempty"`
	ReviewedAt      string `json:"reviewed_at,omitempty"`
	ReviewedCommit  string `json:"reviewed_commit,omitempty"`
	Reference       string `json:"reference,omitempty"`
	ReferenceSHA256 string `json:"reference_sha256,omitempty"`
	Scope           string `json:"scope,omitempty"`
}

type LivenessLock struct {
	Command          []string `json:"command"`
	ExpectedPattern  string   `json:"expected_pattern"`
	AllowedExitCodes []int    `json:"allowed_exit_codes,omitempty"`
}

type ValidatedClaim struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	ValidatedAt  string `json:"validated_at,omitempty"`
	EvidencePath string `json:"evidence_path,omitempty"`
}

type ToolLock struct {
	Name              string           `json:"name"`
	ExecutionClass    string           `json:"execution_class"`
	Repository        string           `json:"repository"`
	Organization      string           `json:"organization"`
	License           string           `json:"license"`
	Role              string           `json:"role"`
	Release           ReleaseLock      `json:"release"`
	Container         *ContainerLock   `json:"container,omitempty"`
	LocalBinarySHA256 string           `json:"local_binary_sha256,omitempty"`
	APIContractPath   string           `json:"api_contract_path,omitempty"`
	APIContractSHA256 string           `json:"api_contract_sha256,omitempty"`
	DependencyLock    string           `json:"dependency_lock,omitempty"`
	SBOMComponentRef  string           `json:"sbom_component_ref"`
	CodeReview        CodeReviewLock   `json:"code_review"`
	Liveness          *LivenessLock    `json:"liveness,omitempty"`
	ValidatedClaims   []ValidatedClaim `json:"validated_claims,omitempty"`
}

type SupplyChainCheck struct {
	Tool           string   `json:"tool"`
	ExecutionClass string   `json:"execution_class,omitempty"`
	ReviewStatus   string   `json:"review_status,omitempty"`
	Approved       bool     `json:"approved"`
	Blockers       []string `json:"blockers"`
}

type SupplyChainReport struct {
	SchemaVersion string             `json:"schema_version"`
	CheckedAt     string             `json:"checked_at"`
	LockPath      string             `json:"lock_path"`
	SBOMPath      string             `json:"sbom_path"`
	Approved      bool               `json:"approved"`
	Checks        []SupplyChainCheck `json:"checks"`
}

func ReadToolLock(path string) (ToolLockFile, error) {
	var lock ToolLockFile
	if err := ReadJSON(path, &lock); err != nil {
		return lock, err
	}
	if lock.SchemaVersion != ToolLockSchemaVersion {
		return lock, fmt.Errorf("schema de tools.lock incompatível: %s", lock.SchemaVersion)
	}
	return lock, nil
}

func ValidateToolLock(lock ToolLockFile, sbomPath string) SupplyChainReport {
	report := SupplyChainReport{SchemaVersion: ToolLockSchemaVersion, CheckedAt: now(), SBOMPath: sbomPath, Approved: true, Checks: []SupplyChainCheck{}}
	sbomRefs := map[string]struct{}{}
	if lock.Policy.RequireSBOM {
		data, err := os.ReadFile(sbomPath)
		if err != nil {
			report.Approved = false
			report.Checks = append(report.Checks, SupplyChainCheck{Tool: "_sbom", Blockers: []string{"SBOM ausente ou ilegível: " + err.Error()}})
		} else {
			var sbom struct {
				Components []struct {
					BOMRef string `json:"bom-ref"`
				} `json:"components"`
			}
			if json.Unmarshal(data, &sbom) != nil {
				report.Approved = false
				report.Checks = append(report.Checks, SupplyChainCheck{Tool: "_sbom", Blockers: []string{"SBOM JSON inválido"}})
			}
			for _, component := range sbom.Components {
				sbomRefs[component.BOMRef] = struct{}{}
			}
		}
	}
	seen := map[string]struct{}{}
	for _, tool := range lock.Tools {
		check := SupplyChainCheck{Tool: tool.Name, ExecutionClass: tool.ExecutionClass, ReviewStatus: tool.CodeReview.Status, Approved: true, Blockers: []string{}}
		name := strings.ToLower(strings.TrimSpace(tool.Name))
		if name == "" {
			check.Blockers = append(check.Blockers, "nome ausente")
		} else if _, ok := seen[name]; ok {
			check.Blockers = append(check.Blockers, "entrada duplicada")
		} else {
			seen[name] = struct{}{}
		}
		version := strings.ToLower(strings.TrimSpace(tool.Release.Tag))
		if lock.Policy.RequirePinnedVersion && version == "" {
			check.Blockers = append(check.Blockers, "versão não pinada")
		}
		if lock.Policy.DenyLatest && (version == "latest" || strings.HasSuffix(version, ":latest")) {
			check.Blockers = append(check.Blockers, "latest é proibido")
		}
		isAPIContract := strings.TrimSpace(tool.APIContractSHA256) != ""
		isSourceBuild := strings.EqualFold(tool.Release.TagType, "commit") && strings.TrimSpace(tool.DependencyLock) != ""
		hasPinnedContainer := tool.Container != nil && sha256Pattern.MatchString(strings.TrimPrefix(strings.ToLower(tool.Container.Digest), "sha256:"))
		executionClass := strings.ToLower(strings.TrimSpace(tool.ExecutionClass))
		if executionClass != "automatic" && executionClass != "manual" && executionClass != "disabled" && executionClass != "infrastructure" {
			check.Blockers = append(check.Blockers, "execution_class inválida")
		}
		if executionClass == "automatic" {
			if tool.Liveness == nil || len(tool.Liveness.Command) == 0 || strings.TrimSpace(tool.Liveness.ExpectedPattern) == "" {
				check.Blockers = append(check.Blockers, "contrato de liveness ausente")
			} else if _, err := regexp.Compile(tool.Liveness.ExpectedPattern); err != nil {
				check.Blockers = append(check.Blockers, "expected_pattern de liveness inválido")
			}
		} else if tool.Liveness != nil {
			check.Blockers = append(check.Blockers, "ferramenta restrita não pode declarar liveness executável")
		}
		if lock.Policy.RequireCommit && !isAPIContract && !commitPattern.MatchString(strings.ToLower(tool.Release.Commit)) {
			check.Blockers = append(check.Blockers, "commit exato ausente ou inválido")
		}
		if !isAPIContract && !strings.EqualFold(tool.Release.TagType, "commit") && tool.Release.ReleaseID <= 0 {
			check.Blockers = append(check.Blockers, "release_id ausente")
		}
		if executionClass == "automatic" && !isAPIContract && !isSourceBuild && !hasPinnedContainer && tool.Release.SignatureStatus != "verified" && !sha256Pattern.MatchString(strings.ToLower(tool.Release.ArtifactSHA256)) {
			check.Blockers = append(check.Blockers, "tag não assinada/verificada e artefato sem SHA-256 validado")
		}
		if executionClass == "automatic" && lock.Policy.RequireArtifactSHA256 && !isAPIContract && !isSourceBuild && tool.Container == nil && !sha256Pattern.MatchString(strings.ToLower(tool.Release.ArtifactSHA256)) {
			check.Blockers = append(check.Blockers, "artefato sem SHA-256 pinado")
		}
		if isAPIContract {
			if tool.APIContractPath == "" || !sha256Pattern.MatchString(strings.ToLower(tool.APIContractSHA256)) {
				check.Blockers = append(check.Blockers, "contrato de API sem caminho/hash válido")
			} else if contractPath, resolveErr := resolveSupplyAsset(tool.APIContractPath); resolveErr != nil {
				check.Blockers = append(check.Blockers, "contrato de API ausente: "+resolveErr.Error())
			} else if actual, err := HashFile(contractPath); err != nil || !strings.EqualFold(actual, tool.APIContractSHA256) {
				check.Blockers = append(check.Blockers, "contrato de API ausente ou hash divergente")
			}
		}
		if tool.Container != nil {
			if strings.Contains(strings.ToLower(tool.Container.Image), ":latest") {
				check.Blockers = append(check.Blockers, "imagem latest proibida")
			}
			digest := strings.TrimPrefix(strings.ToLower(tool.Container.Digest), "sha256:")
			if lock.Policy.RequireContainerDigest && !sha256Pattern.MatchString(digest) {
				check.Blockers = append(check.Blockers, "digest de container inválido")
			}
			if strings.TrimSpace(tool.Container.Entrypoint) != "" && !strings.HasPrefix(tool.Container.Entrypoint, "/") {
				check.Blockers = append(check.Blockers, "entrypoint de container precisa ser absoluto")
			}
			for key := range tool.Container.Environment {
				if !environmentNamePattern.MatchString(key) {
					check.Blockers = append(check.Blockers, "nome inválido em environment do container")
				}
			}
			seenMounts := map[string]struct{}{}
			for _, artifact := range tool.Container.RuntimeArtifacts {
				if artifact.Path == "" || artifact.MountPath == "" || !strings.HasPrefix(artifact.MountPath, "/") || strings.Contains(artifact.MountPath, "..") {
					check.Blockers = append(check.Blockers, "runtime artifact sem caminho/mount seguro")
					continue
				}
				if _, duplicate := seenMounts[artifact.MountPath]; duplicate {
					check.Blockers = append(check.Blockers, "runtime artifact com mount duplicado")
				}
				seenMounts[artifact.MountPath] = struct{}{}
				if !sha256Pattern.MatchString(strings.ToLower(strings.TrimPrefix(artifact.SHA256, "sha256:"))) {
					check.Blockers = append(check.Blockers, "runtime artifact sem SHA-256 válido")
					continue
				}
				artifactPath, resolveErr := resolveSupplyAsset(artifact.Path)
				if resolveErr != nil {
					check.Blockers = append(check.Blockers, "runtime artifact ausente: "+resolveErr.Error())
					continue
				}
				actual, hashErr := HashPath(artifactPath)
				if hashErr != nil || !strings.EqualFold(actual, strings.TrimPrefix(artifact.SHA256, "sha256:")) {
					check.Blockers = append(check.Blockers, "runtime artifact ausente ou hash divergente")
				}
			}
		} else if lock.Policy.RequireContainerDigest && executionClass == "automatic" {
			check.Blockers = append(check.Blockers, "container por digest ausente para ferramenta automática")
		}
		if lock.Policy.RequireCodeReview {
			reviewStatus := strings.ToLower(strings.TrimSpace(tool.CodeReview.Status))
			if reviewStatus != "approved" && reviewStatus != "restricted" {
				check.Blockers = append(check.Blockers, "revisão de adoção pendente")
			} else if executionClass == "automatic" && reviewStatus != "approved" {
				check.Blockers = append(check.Blockers, "ferramenta automática sem revisão aprovada")
			} else if executionClass != "automatic" && reviewStatus != "restricted" {
				check.Blockers = append(check.Blockers, "ferramenta não automática precisa permanecer restricted")
			}
			if reviewStatus == "approved" || reviewStatus == "restricted" {
				if strings.TrimSpace(tool.CodeReview.Reviewer) == "" || strings.TrimSpace(tool.CodeReview.Reference) == "" || strings.TrimSpace(tool.CodeReview.Scope) == "" {
					check.Blockers = append(check.Blockers, "revisão sem reviewer, referência ou escopo")
				} else if reviewedAt, err := time.Parse(time.RFC3339, tool.CodeReview.ReviewedAt); err != nil || reviewedAt.IsZero() {
					check.Blockers = append(check.Blockers, "revisão sem reviewed_at RFC3339 válido")
				}
				if !strings.EqualFold(tool.CodeReview.ReviewedCommit, tool.Release.Commit) {
					check.Blockers = append(check.Blockers, "commit revisado diverge do commit pinado")
				}
				if !sha256Pattern.MatchString(strings.ToLower(strings.TrimPrefix(tool.CodeReview.ReferenceSHA256, "sha256:"))) {
					check.Blockers = append(check.Blockers, "revisão sem SHA-256 da referência")
				} else if reviewPath, resolveErr := resolveSupplyAsset(tool.CodeReview.Reference); resolveErr != nil {
					check.Blockers = append(check.Blockers, "referência da revisão ausente: "+resolveErr.Error())
				} else if actual, hashErr := HashFile(reviewPath); hashErr != nil || !strings.EqualFold(actual, strings.TrimPrefix(tool.CodeReview.ReferenceSHA256, "sha256:")) {
					check.Blockers = append(check.Blockers, "referência da revisão com hash divergente")
				}
			}
		}
		if normalizeToolName(tool.Name) == "keyleak-detector" && executionClass == "automatic" {
			rollbackVerified := false
			for _, claim := range tool.ValidatedClaims {
				if strings.EqualFold(claim.Name, "supabase-rollback-zero-persistence") && claim.Status == "verified" && claim.ValidatedAt != "" && claim.EvidencePath != "" {
					rollbackVerified = true
				}
			}
			if !rollbackVerified {
				check.Blockers = append(check.Blockers, "alegação de rollback do keyleak ainda não validada empiricamente")
			}
		}
		if lock.Policy.RequireSBOM {
			if _, ok := sbomRefs[tool.SBOMComponentRef]; !ok {
				check.Blockers = append(check.Blockers, "componente ausente no SBOM")
			}
		}
		check.Approved = len(check.Blockers) == 0
		if !check.Approved {
			report.Approved = false
		}
		report.Checks = append(report.Checks, check)
	}
	sort.Slice(report.Checks, func(i, j int) bool { return report.Checks[i].Tool < report.Checks[j].Tool })
	return report
}

func resolveSupplyAsset(path string) (string, error) {
	candidates := []string{}
	if filepath.IsAbs(path) {
		candidates = append(candidates, path)
	}
	prefix := "."
	for depth := 0; depth < 6; depth++ {
		candidates = append(candidates, filepath.Join(prefix, path), filepath.Join(prefix, "pipeline", path))
		prefix = filepath.Join(prefix, "..")
	}
	if strings.HasPrefix(filepath.ToSlash(path), "pipeline/") {
		candidates = append(candidates, strings.TrimPrefix(filepath.ToSlash(path), "pipeline/"))
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%s não encontrado", path)
}

func ApproveLocalExecutable(lock ToolLockFile, tool, executable, sbomPath string) error {
	report := ValidateToolLock(lock, sbomPath)
	for _, check := range report.Checks {
		if (check.Tool == "_sbom" || strings.EqualFold(check.Tool, tool)) && !check.Approved {
			return fmt.Errorf("%s bloqueado pela política da cadeia: %s", tool, strings.Join(check.Blockers, "; "))
		}
	}
	var entry *ToolLock
	for index := range lock.Tools {
		if strings.EqualFold(lock.Tools[index].Name, tool) {
			entry = &lock.Tools[index]
			break
		}
	}
	if entry == nil {
		return fmt.Errorf("%s não existe no tools.lock", tool)
	}
	if entry.ExecutionClass != "automatic" {
		return fmt.Errorf("%s bloqueado: execution_class=%s", tool, entry.ExecutionClass)
	}
	if entry.CodeReview.Status != "approved" {
		return fmt.Errorf("%s bloqueado: revisão de código pendente", tool)
	}
	if lock.Policy.RequireSBOM {
		data, err := os.ReadFile(sbomPath)
		if err != nil {
			return fmt.Errorf("%s bloqueado: SBOM ausente ou ilegível", tool)
		}
		var sbom struct {
			Components []struct {
				BOMRef string `json:"bom-ref"`
			} `json:"components"`
		}
		if json.Unmarshal(data, &sbom) != nil {
			return fmt.Errorf("%s bloqueado: SBOM inválido", tool)
		}
		found := false
		for _, component := range sbom.Components {
			if component.BOMRef == entry.SBOMComponentRef {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s bloqueado: componente ausente no SBOM", tool)
		}
	}
	if !sha256Pattern.MatchString(strings.ToLower(entry.LocalBinarySHA256)) {
		return fmt.Errorf("%s bloqueado: hash do binário local não pinado", tool)
	}
	actual, err := HashFile(executable)
	if err != nil {
		return err
	}
	if subtle.ConstantTimeCompare([]byte(strings.ToLower(actual)), []byte(strings.ToLower(entry.LocalBinarySHA256))) != 1 {
		return fmt.Errorf("%s bloqueado: hash do binário diverge do lock", tool)
	}
	if entry.Release.SignatureStatus != "verified" && !sha256Pattern.MatchString(strings.ToLower(entry.Release.ArtifactSHA256)) {
		return fmt.Errorf("%s bloqueado: origem da release não verificada", tool)
	}
	return nil
}

func ApproveContainer(lock ToolLockFile, tool, sbomPath string) (*ToolLock, error) {
	report := ValidateToolLock(lock, sbomPath)
	for _, check := range report.Checks {
		if (check.Tool == "_sbom" || strings.EqualFold(normalizeToolName(check.Tool), normalizeToolName(tool))) && !check.Approved {
			return nil, fmt.Errorf("%s bloqueado pela política da cadeia: %s", tool, strings.Join(check.Blockers, "; "))
		}
	}
	for index := range lock.Tools {
		entry := &lock.Tools[index]
		if normalizeToolName(entry.Name) != normalizeToolName(tool) {
			continue
		}
		if entry.Container == nil {
			return nil, fmt.Errorf("%s bloqueado: container por digest não está pinado", tool)
		}
		if entry.ExecutionClass != "automatic" {
			return nil, fmt.Errorf("%s bloqueado: execution_class=%s", tool, entry.ExecutionClass)
		}
		digest := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(entry.Container.Digest)), "sha256:")
		if !sha256Pattern.MatchString(digest) {
			return nil, fmt.Errorf("%s bloqueado: digest de container inválido", tool)
		}
		if entry.CodeReview.Status != "approved" {
			return nil, fmt.Errorf("%s bloqueado: revisão de código pendente", tool)
		}
		return entry, nil
	}
	return nil, fmt.Errorf("%s não existe no tools.lock", tool)
}

func ResolveToolLock(path string) (ToolLockFile, string, error) {
	candidates := []string{path}
	if path == "" {
		candidates = []string{"tools.lock.json", filepath.Join("pipeline", "tools.lock.json")}
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			lock, readErr := ReadToolLock(candidate)
			return lock, candidate, readErr
		}
	}
	return ToolLockFile{}, "", fmt.Errorf("tools.lock.json não encontrado")
}

func FindExecutable(name string) (string, error) {
	path, ok := executableFor(name)
	if !ok {
		return "", exec.ErrNotFound
	}
	return path, nil
}
