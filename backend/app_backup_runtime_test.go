package backend

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"facade/backend/internal/browser"
	"facade/backend/internal/config"
	"facade/backend/internal/database"
)

func TestBackupReloadAfterMutationDoesNotResurrectStaleDefaultProfile(t *testing.T) {
	root := t.TempDir()
	cfg := config.DefaultConfig()
	if err := cfg.Save(filepath.Join(root, "config.yaml")); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(root, "data", "app.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}

	profileDAO := browser.NewSQLiteProfileDAO(db.GetConn())
	backupOnly := &browser.Profile{
		ProfileId:   "from-backup",
		ProfileName: "备份实例",
		UserDataDir: "from-backup",
		ProxyId:     "__direct__",
		ProxyConfig: "direct://",
	}
	if err := profileDAO.Upsert(backupOnly); err != nil {
		t.Fatal(err)
	}

	staleDefault := &browser.Profile{
		ProfileId:   "stale-default",
		ProfileName: "默认实例",
		UserDataDir: "default",
		ProxyId:     "__direct__",
		ProxyConfig: "direct://",
		Tags:        []string{"默认"},
	}

	app := &App{
		appRoot: root,
		config:  cfg,
		db:      db,
		browserMgr: &browser.Manager{
			Config:     cfg,
			ProfileDAO: profileDAO,
			ProxyDAO:   browser.NewSQLiteProxyDAO(db.GetConn()),
			CoreDAO:    browser.NewSQLiteCoreDAO(db.GetConn()),
			Profiles: map[string]*browser.Profile{
				staleDefault.ProfileId: staleDefault,
				backupOnly.ProfileId: {
					ProfileId:   backupOnly.ProfileId,
					ProfileName: backupOnly.ProfileName,
					UserDataDir: backupOnly.UserDataDir,
					ProxyId:     backupOnly.ProxyId,
					ProxyConfig: backupOnly.ProxyConfig,
				},
			},
			BrowserProcesses: make(map[string]*exec.Cmd),
		},
	}

	if err := app.backupReloadAfterMutation(false); err != nil {
		t.Fatalf("backupReloadAfterMutation: %v", err)
	}

	profiles, err := profileDAO.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected only backup profile after reload, got %d: %+v", len(profiles), profiles)
	}
	if profiles[0].ProfileId != "from-backup" {
		t.Fatalf("expected from-backup, got %q (%s)", profiles[0].ProfileId, profiles[0].ProfileName)
	}
	if _, ok := app.browserMgr.Profiles["stale-default"]; ok {
		t.Fatal("stale default profile still present in memory after reload")
	}
	if _, ok := app.browserMgr.Profiles["from-backup"]; !ok {
		t.Fatal("backup profile missing from memory after reload")
	}
}
