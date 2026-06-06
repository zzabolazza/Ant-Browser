package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestZipDirAndUnzipTo(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	src := filepath.Join(root, "src")
	dstZip := filepath.Join(root, "archive.zip")
	dstDir := filepath.Join(root, "dst")

	if err := os.MkdirAll(filepath.Join(src, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "nested", "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	if err := ZipDir(src, dstZip); err != nil {
		t.Fatalf("ZipDir failed: %v", err)
	}
	if err := UnzipTo(dstZip, dstDir); err != nil {
		t.Fatalf("UnzipTo failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dstDir, "nested", "file.txt"))
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("extracted content = %q, want hello", string(data))
	}
}
