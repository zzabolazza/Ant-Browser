package backup

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptedBackupRoundTrip(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "source.zip")
	encrypted := filepath.Join(root, "backup.facade")
	decrypted := filepath.Join(root, "restored.zip")
	want := bytes.Repeat([]byte("Facade backup data\n"), 70000)
	if err := os.WriteFile(src, want, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EncryptFile(src, encrypted, "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	if err := DecryptFile(encrypted, decrypted, "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(decrypted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("decrypted content differs from source")
	}
}

func TestEncryptedBackupRejectsWrongPasswordAndPlainZip(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "source.zip")
	encrypted := filepath.Join(root, "backup.facade")
	if err := os.WriteFile(src, []byte("PK\x03\x04plain zip"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EncryptFile(src, encrypted, "right-password"); err != nil {
		t.Fatal(err)
	}
	if err := DecryptFile(encrypted, filepath.Join(root, "wrong.zip"), "wrong-password"); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected invalid password error, got %v", err)
	}
	if err := DecryptFile(src, filepath.Join(root, "plain-output.zip"), "any-password"); !errors.Is(err, ErrNotEncryptedBackup) {
		t.Fatalf("expected unencrypted backup rejection, got %v", err)
	}
}

func TestEncryptedBackupRejectsTruncation(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "source.zip")
	encrypted := filepath.Join(root, "backup.facade")
	if err := os.WriteFile(src, bytes.Repeat([]byte("x"), encryptedChunkSize+20), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EncryptFile(src, encrypted, "password"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(encrypted, data[:len(data)-8], 0o600); err != nil {
		t.Fatal(err)
	}
	if err := DecryptFile(encrypted, filepath.Join(root, "output.zip"), "password"); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected truncation rejection, got %v", err)
	}
}

func TestEncryptedBackupSupportsExactChunkBoundary(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "source.zip")
	encrypted := filepath.Join(root, "backup.facade")
	decrypted := filepath.Join(root, "restored.zip")
	want := bytes.Repeat([]byte("z"), encryptedChunkSize)
	if err := os.WriteFile(src, want, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EncryptFile(src, encrypted, "boundary-password"); err != nil {
		t.Fatal(err)
	}
	if err := DecryptFile(encrypted, decrypted, "boundary-password"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(decrypted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("exact-boundary content differs")
	}
}
