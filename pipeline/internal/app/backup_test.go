package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
)

func TestEncryptedBackupRestoreRoundTrip(t *testing.T) {
	root := t.TempDir()
	caseDir := filepath.Join(root, "CASE-001")
	if err := os.MkdirAll(filepath.Join(caseDir, "evidence"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(caseDir, "state.json"), []byte("{\"ok\":true}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(caseDir, "evidence", "result.txt"), []byte("sanitized\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(t.TempDir(), "case.age")
	receipt, err := CreateEncryptedBackup(context.Background(), root, caseDir, backup, identity.Recipient().String())
	if err != nil || receipt.EntryCount != 2 || receipt.CiphertextSHA256 == "" {
		t.Fatalf("backup inválido: err=%v receipt=%+v", err, receipt)
	}
	identityPath := filepath.Join(t.TempDir(), "identity.txt")
	if err := os.WriteFile(identityPath, []byte(identity.String()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	restored := filepath.Join(t.TempDir(), "restored")
	restore, err := RestoreEncryptedBackup(context.Background(), backup, identityPath, restored)
	if err != nil || !restore.Verified || restore.EntryCount != 2 {
		t.Fatalf("restauração inválida: err=%v receipt=%+v", err, restore)
	}
	data, err := os.ReadFile(filepath.Join(restored, "evidence", "result.txt"))
	if err != nil || string(data) != "sanitized\n" {
		t.Fatalf("conteúdo restaurado diverge: err=%v data=%q", err, data)
	}
}

func TestRestoreAcceptsStandardAgeIdentityFile(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(t.TempDir(), "cases")
	caseDir := filepath.Join(root, "CASE-002")
	if err := os.MkdirAll(caseDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(caseDir, "result.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(t.TempDir(), "backup.age")
	if _, err := CreateEncryptedBackup(context.Background(), root, caseDir, backup, identity.Recipient().String()); err != nil {
		t.Fatal(err)
	}
	identityPath := filepath.Join(t.TempDir(), "identity.txt")
	standard := "# created for controlled test\n# public key: " + identity.Recipient().String() + "\n" + identity.String() + "\n"
	if err := os.WriteFile(identityPath, []byte(standard), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := RestoreEncryptedBackup(context.Background(), backup, identityPath, filepath.Join(t.TempDir(), "restore")); err != nil {
		t.Fatal(err)
	}
}

func TestBackupRejectsSourceRootAndPlainOutput(t *testing.T) {
	root := t.TempDir()
	identity, _ := age.GenerateX25519Identity()
	if _, err := CreateEncryptedBackup(context.Background(), root, root, filepath.Join(t.TempDir(), "plain.tar"), identity.Recipient().String()); err == nil {
		t.Fatal("backup da raiz inteira ou sem .age deveria falhar")
	}
}
