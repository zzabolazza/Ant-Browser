package backup

import (
	"os"
	"path/filepath"
	"testing"

	"facade/backend/internal/config"
)

func TestBuildScopeExportsDatabaseSeparatelyFromDataTree(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "data", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(root, "data", "nested", "custom.db")
	if err := os.WriteFile(dbPath, []byte("db"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte("app: {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.DefaultConfig()
	cfg.Database.SQLite.Path = filepath.Join("data", "nested", "custom.db")

	scope, err := BuildScope(BuildOptions{AppRoot: root, Config: cfg})
	if err != nil {
		t.Fatal(err)
	}
	var databaseEntry, dataEntry *ScopeEntry
	for i := range scope.Entries {
		switch scope.Entries[i].ID {
		case "database_sqlite_main":
			databaseEntry = &scope.Entries[i]
		case "app_data_root":
			dataEntry = &scope.Entries[i]
		}
	}
	if databaseEntry == nil {
		t.Fatal("database snapshot is missing from scope")
	}
	if databaseEntry.ArchivePath != "payload/app/database/app.db" {
		t.Fatalf("unexpected database archive path: %s", databaseEntry.ArchivePath)
	}
	if dataEntry == nil {
		t.Fatal("app data entry is missing from scope")
	}
	wantExcluded := map[string]bool{
		filepath.Clean(dbPath):          false,
		filepath.Clean(dbPath + "-wal"): false,
		filepath.Clean(dbPath + "-shm"): false,
	}
	for _, path := range dataEntry.ExcludeSourcePaths {
		if _, ok := wantExcluded[filepath.Clean(path)]; ok {
			wantExcluded[filepath.Clean(path)] = true
		}
	}
	for path, found := range wantExcluded {
		if !found {
			t.Fatalf("database sidecar was not excluded from data tree: %s", path)
		}
	}
}
