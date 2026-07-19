package backend

import (
	"os"
	"path/filepath"
	"testing"

	"facade/backend/internal/snapshot"
)

func TestRestoreSnapshotArchiveKeepsCurrentDataWhenArchiveIsInvalid(t *testing.T) {
	root := t.TempDir()
	userDataDir := filepath.Join(root, "profile")
	if err := os.MkdirAll(userDataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	currentFile := filepath.Join(userDataDir, "current.txt")
	if err := os.WriteFile(currentFile, []byte("current"), 0o644); err != nil {
		t.Fatal(err)
	}
	badZip := filepath.Join(root, "bad.zip")
	if err := os.WriteFile(badZip, []byte("not a zip"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := restoreSnapshotArchive(badZip, userDataDir); err == nil {
		t.Fatal("expected invalid snapshot restore to fail")
	}
	content, err := os.ReadFile(currentFile)
	if err != nil {
		t.Fatalf("current data was removed after failed restore: %v", err)
	}
	if string(content) != "current" {
		t.Fatalf("current data changed after failed restore: %q", content)
	}
}

func TestRestoreSnapshotArchiveReplacesDataAfterValidation(t *testing.T) {
	root := t.TempDir()
	userDataDir := filepath.Join(root, "profile")
	if err := os.MkdirAll(userDataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userDataDir, "stale.txt"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	sourceDir := filepath.Join(root, "snapshot-source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "restored.txt"), []byte("restored"), 0o644); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(root, "snapshot.zip")
	if err := snapshot.ZipDir(sourceDir, zipPath); err != nil {
		t.Fatal(err)
	}

	if err := restoreSnapshotArchive(zipPath, userDataDir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(userDataDir, "stale.txt")); !os.IsNotExist(err) {
		t.Fatal("stale data remained after snapshot restore")
	}
	content, err := os.ReadFile(filepath.Join(userDataDir, "restored.txt"))
	if err != nil || string(content) != "restored" {
		t.Fatalf("snapshot data was not restored: content=%q err=%v", content, err)
	}
}
