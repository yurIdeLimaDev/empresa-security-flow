package app

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"filippo.io/age"
)

const backupManifestName = "_BACKUP_MANIFEST.json"

type BackupEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size_bytes"`
}

type BackupManifest struct {
	SchemaVersion string        `json:"schema_version"`
	Entries       []BackupEntry `json:"entries"`
}

type BackupReceipt struct {
	SchemaVersion    string `json:"schema_version"`
	CreatedAt        string `json:"created_at"`
	CiphertextPath   string `json:"ciphertext_path"`
	CiphertextSHA256 string `json:"ciphertext_sha256"`
	Recipient        string `json:"recipient"`
	EntryCount       int    `json:"entry_count"`
}

type RestoreReceipt struct {
	SchemaVersion string `json:"schema_version"`
	RestoredAt    string `json:"restored_at"`
	BackupSHA256  string `json:"backup_sha256"`
	EntryCount    int    `json:"entry_count"`
	Verified      bool   `json:"verified"`
}

func CreateEncryptedBackup(ctx context.Context, caseRoot, caseDir, destination, recipientText string) (BackupReceipt, error) {
	receipt := BackupReceipt{SchemaVersion: DeliverySchemaVersion, CreatedAt: now(), Recipient: strings.TrimSpace(recipientText)}
	root, err := secureExistingDirectory(caseRoot)
	if err != nil {
		return receipt, err
	}
	source, err := secureExistingDirectory(caseDir)
	if err != nil || !pathWithin(root, source) || samePath(root, source) {
		return receipt, fmt.Errorf("case-dir precisa ser filho direto ou indireto da raiz exclusiva")
	}
	destination, err = filepath.Abs(destination)
	if err != nil {
		return receipt, err
	}
	if pathWithin(root, destination) {
		return receipt, fmt.Errorf("backup criptografado precisa ficar fora da raiz de casos")
	}
	if filepath.Ext(destination) != ".age" {
		return receipt, fmt.Errorf("backup precisa usar extensão .age")
	}
	recipient, err := age.ParseX25519Recipient(receipt.Recipient)
	if err != nil {
		return receipt, err
	}
	entries, err := collectBackupEntries(source)
	if err != nil {
		return receipt, err
	}
	manifest := BackupManifest{SchemaVersion: DeliverySchemaVersion, Entries: entries}
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return receipt, err
	}
	manifestData = append(manifestData, '\n')
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return receipt, err
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return receipt, err
	}
	complete := false
	defer func() {
		if !complete {
			output.Close()
			_ = os.Remove(destination)
		}
	}()
	ageWriter, err := age.Encrypt(output, recipient)
	if err != nil {
		return receipt, err
	}
	tarWriter := tar.NewWriter(ageWriter)
	if err := writeTarBytes(tarWriter, backupManifestName, manifestData); err != nil {
		return receipt, err
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return receipt, err
		}
		file, err := os.Open(filepath.Join(source, filepath.FromSlash(entry.Path)))
		if err != nil {
			return receipt, err
		}
		header := &tar.Header{Name: entry.Path, Mode: 0o600, Size: entry.Size, ModTime: time.Unix(0, 0).UTC(), AccessTime: time.Unix(0, 0).UTC(), ChangeTime: time.Unix(0, 0).UTC(), Typeflag: tar.TypeReg, Uid: 0, Gid: 0}
		if err := tarWriter.WriteHeader(header); err != nil {
			file.Close()
			return receipt, err
		}
		if _, err := io.Copy(tarWriter, file); err != nil {
			file.Close()
			return receipt, err
		}
		if err := file.Close(); err != nil {
			return receipt, err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return receipt, err
	}
	if err := ageWriter.Close(); err != nil {
		return receipt, err
	}
	if err := output.Sync(); err != nil {
		return receipt, err
	}
	if err := output.Close(); err != nil {
		return receipt, err
	}
	complete = true
	receipt.CiphertextPath = destination
	receipt.CiphertextSHA256, err = HashFile(destination)
	if err != nil {
		return receipt, err
	}
	receipt.EntryCount = len(entries)
	if err := WriteJSON(destination+".receipt.json", receipt); err != nil {
		return receipt, err
	}
	return receipt, nil
}

func RestoreEncryptedBackup(ctx context.Context, backupPath, identityPath, outputDir string) (RestoreReceipt, error) {
	receipt := RestoreReceipt{SchemaVersion: DeliverySchemaVersion, RestoredAt: now()}
	backupPath, err := filepath.Abs(backupPath)
	if err != nil {
		return receipt, err
	}
	identityData, err := os.ReadFile(identityPath)
	if err != nil {
		return receipt, err
	}
	identities, err := age.ParseIdentities(strings.NewReader(string(identityData)))
	for index := range identityData {
		identityData[index] = 0
	}
	if err != nil {
		return receipt, fmt.Errorf("identidade age inválida: %w", err)
	}
	if len(identities) != 1 {
		return receipt, fmt.Errorf("arquivo de identidade deve conter exatamente uma identidade age")
	}
	if entries, readErr := os.ReadDir(outputDir); readErr == nil && len(entries) != 0 {
		return receipt, fmt.Errorf("diretório de restauração precisa estar vazio")
	} else if readErr != nil && !os.IsNotExist(readErr) {
		return receipt, readErr
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return receipt, err
	}
	input, err := os.Open(backupPath)
	if err != nil {
		return receipt, err
	}
	defer input.Close()
	reader, err := age.Decrypt(input, identities...)
	if err != nil {
		return receipt, err
	}
	tarReader := tar.NewReader(reader)
	var manifest BackupManifest
	actual := []BackupEntry{}
	seenManifest := false
	for {
		if err := ctx.Err(); err != nil {
			return receipt, err
		}
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return receipt, err
		}
		if header.Typeflag != tar.TypeReg || !safeRelativePath(header.Name) {
			return receipt, fmt.Errorf("entrada insegura no backup: %s", header.Name)
		}
		if header.Name == backupManifestName {
			if seenManifest || len(actual) != 0 {
				return receipt, fmt.Errorf("manifesto de backup ausente da primeira entrada")
			}
			data, readErr := io.ReadAll(io.LimitReader(tarReader, 8<<20))
			if readErr != nil || json.Unmarshal(data, &manifest) != nil || manifest.SchemaVersion != DeliverySchemaVersion {
				return receipt, fmt.Errorf("manifesto interno inválido")
			}
			seenManifest = true
			continue
		}
		if !seenManifest {
			return receipt, fmt.Errorf("manifesto interno precisa ser a primeira entrada")
		}
		destination := filepath.Join(outputDir, filepath.FromSlash(header.Name))
		if !pathWithin(outputDir, destination) {
			return receipt, fmt.Errorf("path traversal no backup")
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			return receipt, err
		}
		file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return receipt, err
		}
		if _, err := io.CopyN(file, tarReader, header.Size); err != nil {
			file.Close()
			return receipt, err
		}
		if err := file.Close(); err != nil {
			return receipt, err
		}
		hash, err := HashFile(destination)
		if err != nil {
			return receipt, err
		}
		actual = append(actual, BackupEntry{Path: filepath.ToSlash(header.Name), SHA256: hash, Size: header.Size})
	}
	if !seenManifest || !sameBackupEntries(manifest.Entries, actual) {
		return receipt, fmt.Errorf("restauração não corresponde ao manifesto interno")
	}
	receipt.BackupSHA256, err = HashFile(backupPath)
	if err != nil {
		return receipt, err
	}
	receipt.EntryCount, receipt.Verified = len(actual), true
	if err := WriteJSON(filepath.Join(outputDir, "RESTORE-VERIFIED.json"), receipt); err != nil {
		return receipt, err
	}
	return receipt, nil
}

func collectBackupEntries(root string) ([]BackupEntry, error) {
	items := []BackupEntry{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("links simbólicos são proibidos no backup")
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("arquivo não regular no caso")
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || !safeRelativePath(relative) {
			return fmt.Errorf("caminho inseguro no caso")
		}
		hash, err := HashFile(path)
		if err != nil {
			return err
		}
		items = append(items, BackupEntry{Path: filepath.ToSlash(relative), SHA256: hash, Size: info.Size()})
		return nil
	})
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return items, err
}

func writeTarBytes(writer *tar.Writer, name string, data []byte) error {
	header := &tar.Header{Name: name, Mode: 0o600, Size: int64(len(data)), ModTime: time.Unix(0, 0).UTC(), AccessTime: time.Unix(0, 0).UTC(), ChangeTime: time.Unix(0, 0).UTC(), Typeflag: tar.TypeReg, Uid: 0, Gid: 0}
	if err := writer.WriteHeader(header); err != nil {
		return err
	}
	_, err := writer.Write(data)
	return err
}

func sameBackupEntries(expected, actual []BackupEntry) bool {
	if len(expected) != len(actual) {
		return false
	}
	sort.Slice(expected, func(i, j int) bool { return expected[i].Path < expected[j].Path })
	sort.Slice(actual, func(i, j int) bool { return actual[i].Path < actual[j].Path })
	for index := range expected {
		if expected[index] != actual[index] {
			return false
		}
	}
	return true
}
