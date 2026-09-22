package app

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"filippo.io/age"
)

const DeliverySchemaVersion = "1.0.0"

type DeliveryManifestEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size_bytes"`
}

type DeliveryManifest struct {
	SchemaVersion  string                  `json:"schema_version"`
	CaseID         string                  `json:"case_id"`
	Commit         string                  `json:"commit"`
	BundleSHA256   string                  `json:"bundle_sha256"`
	PlanSHA256     string                  `json:"plan_sha256"`
	ApprovalSHA256 string                  `json:"approval_sha256"`
	NonRegression  string                  `json:"security_non_regression"`
	Entries        []DeliveryManifestEntry `json:"entries"`
}

type DeliveryPackageResult struct {
	SchemaVersion     string `json:"schema_version"`
	CaseID            string `json:"case_id"`
	ManifestPath      string `json:"manifest_path"`
	ManifestSHA256    string `json:"manifest_sha256"`
	ArchivePath       string `json:"archive_path"`
	ArchiveSHA256     string `json:"archive_sha256"`
	EncryptedPath     string `json:"encrypted_path"`
	EncryptedSHA256   string `json:"encrypted_sha256"`
	Recipient         string `json:"recipient"`
	ContainerRequired string `json:"container_required"`
}

func BuildDeliveryPackage(ctx context.Context, authorizationPath, planPath, statePath, approvalPath, bundlePath, patchRoot, outputDir, recipientText string) (DeliveryPackageResult, error) {
	result := DeliveryPackageResult{SchemaVersion: DeliverySchemaVersion, Recipient: recipientText, ContainerRequired: "gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab"}
	if patchRoot != "" {
		return result, fmt.Errorf("patch-root não é aceito na entrega: os patches são derivados do baseline e BEST aprovados; omita esse parâmetro")
	}
	var authorization DeliveryAuthorization
	if err := ReadJSON(authorizationPath, &authorization); err != nil {
		return result, err
	}
	if authorization.Status != "approved_for_delivery" || authorization.SecurityNonRegression != "passed" {
		return result, fmt.Errorf("entrega exige autorização final e não regressão aprovada")
	}
	plan, _, err := readRemediationPlan(planPath)
	if err != nil {
		return result, err
	}
	var state RemediationState
	if err := ReadJSON(statePath, &state); err != nil {
		return result, err
	}
	var approval RemediationApproval
	if err := ReadJSONWithSchema(approvalPath, "remediation-approval.schema.json", &approval); err != nil {
		return result, err
	}
	bundle, err := ReadBundle(bundlePath)
	if err != nil {
		return result, err
	}
	for _, evidence := range bundle.Evidence {
		if evidence.Sensitive {
			return result, fmt.Errorf("bundle permitido não pode conter evidência marcada como sensível")
		}
	}
	if authorization.CaseID != plan.CaseID || authorization.CaseID != state.CaseID || authorization.CaseID != approval.CaseID {
		return result, fmt.Errorf("artefatos pertencem a casos diferentes")
	}
	if state.BestRef != authorization.Commit || approval.ReviewedCommit != authorization.Commit || approval.Decision != "approved" || !state.GlobalGatePassed || state.FinalApprovalStatus != "approved" || authorization.HumanReviewStage != "final" {
		return result, fmt.Errorf("commit da entrega não corresponde ao BEST aprovado")
	}
	checks := []struct{ path, expected, name string }{{planPath, authorization.PlanSHA256, "plan"}, {statePath, authorization.StateSHA256, "state"}, {approvalPath, authorization.ApprovalSHA256, "approval"}, {bundlePath, authorization.BundleSHA256, "bundle"}}
	for _, check := range checks {
		hash, hashErr := HashFile(check.path)
		if hashErr != nil || !strings.EqualFold(hash, check.expected) {
			return result, fmt.Errorf("hash de %s diverge da autorização", check.name)
		}
	}
	if !strings.EqualFold(state.BestBundleSHA256, authorization.BundleSHA256) || !strings.EqualFold(approval.ReviewedBundleSHA256, authorization.BundleSHA256) {
		return result, fmt.Errorf("bundle não corresponde ao BEST/revisão")
	}
	patchPath := filepath.Join(filepath.Dir(authorizationPath), "approved-security.patch")
	patchInfo, err := os.Lstat(patchPath)
	if err != nil || !patchInfo.Mode().IsRegular() || patchInfo.Size() >= 16<<20 || !validSHA256Hex(authorization.PatchSHA256) {
		return result, fmt.Errorf("patch canônico aprovado ausente ou inválido; refaça a finalização em diretório novo")
	}
	patch, err := os.ReadFile(patchPath)
	if err != nil || bytesSHA256(patch) != authorization.PatchSHA256 {
		return result, fmt.Errorf("patch diverge da autorização final")
	}
	recipient, err := age.ParseX25519Recipient(strings.TrimSpace(recipientText))
	if err != nil {
		return result, fmt.Errorf("destinatário age inválido: %w", err)
	}
	outputAbs, err := filepath.Abs(outputDir)
	if err != nil {
		return result, err
	}
	if pathWithin(plan.RepositoryPath, outputAbs) {
		return result, fmt.Errorf("pacote de entrega precisa ficar fora do repositório avaliado")
	}
	if entries, readErr := os.ReadDir(outputAbs); readErr == nil && len(entries) != 0 {
		return result, fmt.Errorf("output-dir da entrega precisa estar vazio")
	} else if readErr != nil && !os.IsNotExist(readErr) {
		return result, readErr
	}
	packageDir := filepath.Join(outputAbs, "package")
	if err := os.MkdirAll(filepath.Join(packageDir, "patches"), 0o700); err != nil {
		return result, err
	}
	result.CaseID = authorization.CaseID
	htmlData := renderDeliveryHTML(authorization, plan, bundle)
	pdfData := renderDeterministicPDF(deliveryReportLines(authorization, plan, bundle))
	if err := os.WriteFile(filepath.Join(packageDir, "report.html"), htmlData, 0o600); err != nil {
		return result, err
	}
	if err := os.WriteFile(filepath.Join(packageDir, "report.pdf"), pdfData, 0o600); err != nil {
		return result, err
	}
	for source, destination := range map[string]string{bundlePath: "bundle.json", authorizationPath: "delivery-authorization.json"} {
		if err := copyFileExact(source, filepath.Join(packageDir, destination)); err != nil {
			return result, err
		}
	}
	if len(patch) > 0 {
		if err := os.WriteFile(filepath.Join(packageDir, "patches", "security.patch"), patch, 0o600); err != nil {
			return result, err
		}
	}
	manifest := DeliveryManifest{SchemaVersion: DeliverySchemaVersion, CaseID: authorization.CaseID, Commit: authorization.Commit, BundleSHA256: authorization.BundleSHA256, PlanSHA256: authorization.PlanSHA256, ApprovalSHA256: authorization.ApprovalSHA256, NonRegression: authorization.SecurityNonRegression, Entries: []DeliveryManifestEntry{}}
	files, err := packageFiles(packageDir, map[string]bool{"manifest.json": true, "SHA256SUMS": true})
	if err != nil {
		return result, err
	}
	for _, file := range files {
		hash, hashErr := HashFile(filepath.Join(packageDir, filepath.FromSlash(file)))
		info, statErr := os.Stat(filepath.Join(packageDir, filepath.FromSlash(file)))
		if hashErr != nil || statErr != nil {
			return result, fmt.Errorf("manifesto: arquivo ilegível %s", file)
		}
		manifest.Entries = append(manifest.Entries, DeliveryManifestEntry{Path: file, SHA256: hash, Size: info.Size()})
	}
	manifestPath := filepath.Join(packageDir, "manifest.json")
	if err := WriteJSON(manifestPath, manifest); err != nil {
		return result, err
	}
	manifestHash, err := HashFile(manifestPath)
	if err != nil {
		return result, err
	}
	checksumFiles := append(append([]string{}, files...), "manifest.json")
	sort.Strings(checksumFiles)
	var sums strings.Builder
	for _, file := range checksumFiles {
		hash, hashErr := HashFile(filepath.Join(packageDir, filepath.FromSlash(file)))
		if hashErr != nil {
			return result, hashErr
		}
		sums.WriteString(hash + "  " + file + "\n")
	}
	if err := os.WriteFile(filepath.Join(packageDir, "SHA256SUMS"), []byte(sums.String()), 0o600); err != nil {
		return result, err
	}
	archivePath := filepath.Join(outputAbs, "delivery.zip")
	if err := writeDeterministicZip(packageDir, archivePath); err != nil {
		return result, err
	}
	archiveHash, err := HashFile(archivePath)
	if err != nil {
		return result, err
	}
	encryptedPath := archivePath + ".age"
	if err := encryptAge(ctx, archivePath, encryptedPath, recipient); err != nil {
		return result, err
	}
	encryptedHash, err := HashFile(encryptedPath)
	if err != nil {
		return result, err
	}
	result.ManifestPath, result.ManifestSHA256 = manifestPath, manifestHash
	result.ArchivePath, result.ArchiveSHA256 = archivePath, archiveHash
	result.EncryptedPath, result.EncryptedSHA256 = encryptedPath, encryptedHash
	if err := WriteJSON(filepath.Join(outputAbs, "delivery-package-result.json"), result); err != nil {
		return result, err
	}
	return result, nil
}

func renderDeliveryHTML(authorization DeliveryAuthorization, plan RemediationPlan, bundle NormalizedBundle) []byte {
	var builder strings.Builder
	builder.WriteString("<!doctype html><html lang=\"pt-BR\"><head><meta charset=\"utf-8\"><title>Relatório de segurança</title><style>body{font:14px system-ui;max-width:900px;margin:40px auto;color:#17202a}code{overflow-wrap:anywhere}table{border-collapse:collapse;width:100%}th,td{border:1px solid #ccd1d1;padding:8px;text-align:left}</style></head><body>")
	builder.WriteString("<h1>Relatório de segurança — " + html.EscapeString(authorization.CaseID) + "</h1>")
	builder.WriteString("<p>Versão do formato: " + DeliverySchemaVersion + "</p><p>Commit: <code>" + html.EscapeString(authorization.Commit) + "</code><br>Bundle SHA-256: <code>" + html.EscapeString(authorization.BundleSHA256) + "</code><br>Não regressão: <strong>" + html.EscapeString(authorization.SecurityNonRegression) + "</strong></p>")
	builder.WriteString("<h2>Escopo e cobertura</h2><p>Scope: <code>" + html.EscapeString(bundle.Scope.ID) + "</code>; fingerprint: <code>" + html.EscapeString(bundle.Scope.Fingerprint) + "</code>.</p><ul>")
	for _, run := range bundle.ToolRuns {
		builder.WriteString("<li>" + html.EscapeString(run.Tool) + " " + html.EscapeString(run.Version) + " — " + html.EscapeString(run.Status) + "</li>")
	}
	builder.WriteString("</ul><h2>Achados após reteste</h2><table><thead><tr><th>Severidade</th><th>Título</th><th>Status</th><th>Local</th></tr></thead><tbody>")
	for _, finding := range bundle.Findings {
		builder.WriteString("<tr><td>" + html.EscapeString(finding.Severity) + "</td><td>" + html.EscapeString(finding.Title) + "</td><td>" + html.EscapeString(finding.Status) + "</td><td><code>" + html.EscapeString(finding.Location) + "</code></td></tr>")
	}
	builder.WriteString("</tbody></table><h2>Limitações</h2><p>O resultado vale somente para o escopo, o commit, as ferramentas e os hashes registrados. Evidência sensível e logs brutos não integram este pacote.</p><h2>Gates</h2><p>Estratégia: " + html.EscapeString(plan.ExecutionStrategy) + "; revisão humana: final.</p></body></html>\n")
	return []byte(builder.String())
}

func deliveryReportLines(authorization DeliveryAuthorization, plan RemediationPlan, bundle NormalizedBundle) []string {
	lines := []string{"RELATORIO DE SEGURANCA", "Caso: " + authorization.CaseID, "Commit: " + authorization.Commit, "Bundle SHA-256: " + authorization.BundleSHA256, "Nao regressao: " + authorization.SecurityNonRegression, "Escopo: " + bundle.Scope.ID, "Cobertura:"}
	for _, run := range bundle.ToolRuns {
		lines = append(lines, "- "+run.Tool+" "+run.Version+" ["+run.Status+"]")
	}
	lines = append(lines, "Achados apos reteste:")
	for _, finding := range bundle.Findings {
		lines = append(lines, "- "+finding.Severity+" | "+finding.Title+" | "+finding.Status)
	}
	lines = append(lines, "Evidencia sensivel e logs brutos foram excluidos.", "Estrategia: "+plan.ExecutionStrategy, "Revisao humana: final")
	return lines
}

func renderDeterministicPDF(lines []string) []byte {
	type row struct {
		text    string
		style   string
		spacing int
	}
	rows := []row{}
	for index, line := range lines {
		if index == 0 {
			continue
		}
		style, spacing, width := "body", 16, 86
		if line == "Cobertura:" || line == "Achados apos reteste:" {
			style, spacing, width = "heading", 26, 72
		}
		for partIndex, part := range wrapASCII(line, width) {
			partSpacing := spacing
			if partIndex > 0 {
				partSpacing = 16
			}
			if partIndex == 0 && len(rows) > 0 && rows[len(rows)-1].style == "heading" {
				partSpacing += 8
			}
			rows = append(rows, row{text: part, style: style, spacing: partSpacing})
		}
	}
	pages := [][]row{{}}
	y := 710
	for _, item := range rows {
		if y-item.spacing < 72 {
			pages = append(pages, []row{})
			y = 710
		}
		pages[len(pages)-1] = append(pages[len(pages)-1], item)
		y -= item.spacing
	}
	fontRegular := 3 + len(pages)*2
	fontBold := fontRegular + 1
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", ""}
	kids := make([]string, 0, len(pages))
	for pageIndex, pageRows := range pages {
		pageObject := 3 + pageIndex*2
		contentObject := pageObject + 1
		kids = append(kids, fmt.Sprintf("%d 0 R", pageObject))
		var content strings.Builder
		content.WriteString("q 0.055 0.118 0.204 rg 0 742 595 100 re f Q\n")
		content.WriteString(fmt.Sprintf("BT /F2 20 Tf 1 1 1 rg 48 792 Td (%s) Tj ET\n", pdfEscape("RELATORIO DE SEGURANCA")))
		content.WriteString(fmt.Sprintf("BT /F1 9 Tf 0.80 0.86 0.94 rg 48 770 Td (%s) Tj ET\n", pdfEscape("Entrega tecnica controlada - formato "+DeliverySchemaVersion)))
		content.WriteString("q 0.12 0.42 0.64 rg 48 728 499 2 re f Q\n")
		cursor := 710
		for _, item := range pageRows {
			cursor -= item.spacing
			font, size, red, green, blue, x := "F1", 10, 0.10, 0.14, 0.18, 48
			if item.style == "heading" {
				font, size, red, green, blue = "F2", 12, 0.07, 0.32, 0.50
			}
			if strings.HasPrefix(item.text, "- ") {
				x = 60
			}
			content.WriteString(fmt.Sprintf("BT /%s %d Tf %.2f %.2f %.2f rg %d %d Td (%s) Tj ET\n", font, size, red, green, blue, x, cursor, pdfEscape(item.text)))
		}
		content.WriteString("q 0.78 0.82 0.86 RG 48 54 m 547 54 l S Q\n")
		content.WriteString(fmt.Sprintf("BT /F1 8 Tf 0.35 0.40 0.46 rg 48 38 Td (%s) Tj ET\n", pdfEscape(fmt.Sprintf("Empresa Security | pagina %d de %d | evidencia saneada", pageIndex+1, len(pages)))))
		objects = append(objects,
			fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 %d 0 R /F2 %d 0 R >> >> /Contents %d 0 R >>", fontRegular, fontBold, contentObject),
			fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", content.Len(), content.String()),
		)
	}
	objects[1] = fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(pages))
	objects = append(objects, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>", "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>")
	var output bytes.Buffer
	output.WriteString("%PDF-1.4\n% deterministic empresa-security\n")
	offsets := []int{0}
	for index, object := range objects {
		offsets = append(offsets, output.Len())
		fmt.Fprintf(&output, "%d 0 obj\n%s\nendobj\n", index+1, object)
	}
	xref := output.Len()
	fmt.Fprintf(&output, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&output, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&output, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return output.Bytes()
}

func wrapASCII(value string, limit int) []string {
	value = asciiOnly(value, 4096)
	if len(value) <= limit {
		return []string{value}
	}
	indent := ""
	if strings.HasPrefix(value, "- ") {
		indent = "  "
	}
	words := strings.Fields(value)
	result := []string{}
	current := ""
	for _, word := range words {
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}
		if len(candidate) <= limit {
			current = candidate
			continue
		}
		if current != "" {
			result = append(result, current)
		}
		for len(word) > limit {
			result = append(result, indent+word[:limit-len(indent)])
			word = word[limit-len(indent):]
		}
		current = indent + word
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func asciiOnly(value string, limit int) string {
	var builder strings.Builder
	for _, character := range value {
		if builder.Len() >= limit {
			break
		}
		if character >= 32 && character <= 126 {
			builder.WriteRune(character)
		} else {
			builder.WriteByte('?')
		}
	}
	return builder.String()
}

func pdfEscape(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "(", "\\(")
	return strings.ReplaceAll(value, ")", "\\)")
}

func packageFiles(root string, excluded map[string]bool) ([]string, error) {
	items := []string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("link simbólico proibido no pacote")
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if !excluded[relative] {
			items = append(items, relative)
		}
		return nil
	})
	sort.Strings(items)
	return items, err
}

func writeDeterministicZip(root, destination string) error {
	files, err := packageFiles(root, map[string]bool{})
	if err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	archive := zip.NewWriter(output)
	fixed := time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, relative := range files {
		data, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if readErr != nil {
			archive.Close()
			output.Close()
			return readErr
		}
		header := &zip.FileHeader{Name: relative, Method: zip.Deflate}
		header.SetModTime(fixed)
		header.SetMode(0o600)
		writer, createErr := archive.CreateHeader(header)
		if createErr != nil {
			archive.Close()
			output.Close()
			return createErr
		}
		if _, writeErr := writer.Write(data); writeErr != nil {
			archive.Close()
			output.Close()
			return writeErr
		}
	}
	if err := archive.Close(); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

func encryptAge(ctx context.Context, source, destination string, recipient age.Recipient) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	writer, err := age.Encrypt(output, recipient)
	if err != nil {
		output.Close()
		return err
	}
	if _, err := io.Copy(writer, input); err != nil {
		writer.Close()
		output.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

func DecodeDeliveryManifest(path string) (DeliveryManifest, error) {
	var manifest DeliveryManifest
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest, err
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, err
	}
	return manifest, nil
}
