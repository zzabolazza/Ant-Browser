package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	metaPath := filepath.Join(dir, "snap-1_demo.meta.json")
	zipPath := filepath.Join(dir, "snap-1_demo.zip")

	if err := os.WriteFile(metaPath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}
	if err := os.WriteFile(zipPath, []byte("zip"), 0o644); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	gotMeta, gotZip, err := FindFiles(dir, "snap-1")
	if err != nil {
		t.Fatalf("FindFiles failed: %v", err)
	}
	if gotMeta != metaPath {
		t.Fatalf("meta path = %q, want %q", gotMeta, metaPath)
	}
	if gotZip != zipPath {
		t.Fatalf("zip path = %q, want %q", gotZip, zipPath)
	}
}

func TestEnsureDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir, err := EnsureDir(root, "profile-1")
	if err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}
	expected := filepath.Join(root, "snapshots", "profile-1")
	if dir != expected {
		t.Fatalf("dir = %q, want %q", dir, expected)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("dir was not created: info=%v err=%v", info, err)
	}
}
