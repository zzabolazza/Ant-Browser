package backend

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"facade/backend/internal/backup"
	"facade/backend/internal/config"
	"facade/backend/internal/database"
)

func TestFullBackupUsesSnapshotAndDoesNotPackItsOwnOutput(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.DefaultConfig()
	cfg.Database.SQLite.Path = filepath.Join("data", "app.db")
	if err := cfg.Save(filepath.Join(root, "config.yaml")); err != nil {
		t.Fatal(err)
	}
	db, err := database.NewDB(filepath.Join(dataDir, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetConn().Exec(`INSERT INTO browser_bookmarks (name, url, sort_order) VALUES ('Saved', 'https://saved.example/', 0)`); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "ordinary.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	scope, err := backup.BuildScope(backup.BuildOptions{AppRoot: root, Config: cfg})
	if err != nil {
		t.Fatal(err)
	}
	app := &App{db: db}
	scope, cleanup, err := app.backupPrepareDatabaseSnapshot(scope)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	zipPath := filepath.Join(dataDir, "facade-backup.zip")
	manifest := backup.BuildManifest(scope, "Facade", "test", time.Now())
	if _, _, _, err := backupWritePackageZip(zipPath, scope, manifest, nil); err != nil {
		t.Fatal(err)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	names := make(map[string]bool, len(reader.File))
	for _, file := range reader.File {
		names[file.Name] = true
	}
	for _, required := range []string{
		"manifest.json",
		"payload/app/database/app.db",
		"payload/app/data/ordinary.txt",
		"payload/app/data/",
	} {
		if !names[required] {
			t.Fatalf("backup is missing %s", required)
		}
	}
	for _, forbidden := range []string{
		"payload/app/data/app.db",
		"payload/app/data/app.db-wal",
		"payload/app/data/app.db-shm",
		"payload/app/data/facade-backup.zip",
		"payload/app/data/facade-backup.zip.tmp",
	} {
		if names[forbidden] {
			t.Fatalf("backup unexpectedly contains %s", forbidden)
		}
	}
}

func TestValidateFullRestorePayloadRejectsMissingAppData(t *testing.T) {
	payloadRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(payloadRoot, "system"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(payloadRoot, "system", "config.yaml"), []byte("app: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := backupValidateFullRestorePayload(payloadRoot, config.DefaultConfig())
	if err == nil {
		t.Fatal("expected missing app data directory to be rejected")
	}
}

func TestValidateFullRestorePayloadRequiresSeparateBrowserData(t *testing.T) {
	payloadRoot := t.TempDir()
	for _, dir := range []string{
		filepath.Join(payloadRoot, "system"),
		filepath.Join(payloadRoot, "app", "data"),
		filepath.Join(payloadRoot, "app", "database"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(payloadRoot, "system", "config.yaml"), []byte("app: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(payloadRoot, "app", "database", "app.db"), []byte("placeholder"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Browser.UserDataRoot = filepath.Join(t.TempDir(), "browser-data")
	if err := backupValidateFullRestorePayload(payloadRoot, cfg); err == nil {
		t.Fatal("expected missing separate browser user data directory to be rejected")
	}

	if err := os.MkdirAll(filepath.Join(payloadRoot, "browser", "user-data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := backupValidateFullRestorePayload(payloadRoot, cfg); err != nil {
		t.Fatalf("expected complete payload to pass validation: %v", err)
	}
}

func TestBackupReplaceFileReplacesExistingOnlyAfterSourceIsReady(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "new.facade")
	dst := filepath.Join(root, "saved.facade")
	if err := os.WriteFile(src, []byte("new encrypted backup"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old encrypted backup"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := backupReplaceFile(src, dst); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new encrypted backup" {
		t.Fatalf("unexpected replaced content: %q", got)
	}
}
