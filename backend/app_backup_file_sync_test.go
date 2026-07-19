package backend

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupRemoveContentsExceptPreservesOnlyNestedDatabaseFiles(t *testing.T) {
	root := t.TempDir()
	dbDir := filepath.Join(root, "nested", "database")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(dbDir, "app.db")
	for path, content := range map[string]string{
		dbPath:                            "db",
		dbPath + "-wal":                   "wal",
		filepath.Join(dbDir, "stale.txt"): "remove",
		filepath.Join(root, "other.txt"):  "remove",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	keep := map[string]struct{}{
		backupNormalizePath(dbPath):          {},
		backupNormalizePath(dbPath + "-wal"): {},
		backupNormalizePath(dbPath + "-shm"): {},
	}
	if err := backupRemoveContentsExcept(root, keep); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{dbPath, dbPath + "-wal"} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("kept database file is missing: %s: %v", path, err)
		}
	}
	for _, path := range []string{filepath.Join(dbDir, "stale.txt"), filepath.Join(root, "other.txt")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("stale file was not removed: %s", path)
		}
	}
}
