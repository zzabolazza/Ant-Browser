package backend

import (
	"path/filepath"
	"testing"

	"facade/backend/internal/browser"
	"facade/backend/internal/config"
	"facade/backend/internal/database"
)

func openMigratedTestDB(t *testing.T, name string) *database.DB {
	t.Helper()
	db, err := database.NewDB(filepath.Join(t.TempDir(), name))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestBookmarkSaveAllowsEmptyList(t *testing.T) {
	db := openMigratedTestDB(t, "bookmarks.db")
	cfg := config.DefaultConfig()
	app := &App{
		config: cfg,
		browserMgr: &browser.Manager{
			Config:      cfg,
			BookmarkDAO: browser.NewSQLiteBookmarkDAO(db.GetConn()),
		},
	}

	if err := app.BookmarkSave([]BrowserBookmark{}); err != nil {
		t.Fatal(err)
	}
	if got := app.BookmarkList(); len(got) != 0 {
		t.Fatalf("expected an empty bookmark list, got %d entries", len(got))
	}
}

func TestBackupDatabaseCanBeImportedRepeatedly(t *testing.T) {
	source := openMigratedTestDB(t, "source.db")
	if _, err := source.GetConn().Exec(
		`INSERT INTO browser_bookmarks (name, url, open_on_start, sort_order) VALUES (?, ?, ?, ?)`,
		"Example", "https://example.com/", 1, 0,
	); err != nil {
		t.Fatal(err)
	}

	target := openMigratedTestDB(t, "target.db")
	app := &App{db: target}
	first := &backupMergeStats{}
	if err := app.backupMergeDatabaseFromSource(sourcePath(t, source), false, first); err != nil {
		t.Fatalf("first import failed: %v", err)
	}
	second := &backupMergeStats{}
	if err := app.backupMergeDatabaseFromSource(sourcePath(t, source), false, second); err != nil {
		t.Fatalf("second import failed: %v", err)
	}
	if second.Skipped == 0 {
		t.Fatal("expected the repeated import to skip existing data")
	}
}

func TestBackupDatabaseMergeKeepsExistingBookmark(t *testing.T) {
	source := openMigratedTestDB(t, "source.db")
	if _, err := source.GetConn().Exec(
		`INSERT INTO browser_bookmarks (name, url, open_on_start, sort_order) VALUES (?, ?, ?, ?), (?, ?, ?, ?)`,
		"Backup Name", "https://same.example/", 1, 1,
		"New Bookmark", "https://new.example/", 1, 2,
	); err != nil {
		t.Fatal(err)
	}

	target := openMigratedTestDB(t, "target.db")
	if _, err := target.GetConn().Exec(
		`INSERT INTO browser_bookmarks (name, url, open_on_start, sort_order) VALUES (?, ?, ?, ?)`,
		"Current Name", "https://same.example/", 0, 9,
	); err != nil {
		t.Fatal(err)
	}

	app := &App{db: target}
	stats := &backupMergeStats{}
	if err := app.backupMergeDatabaseFromSource(sourcePath(t, source), false, stats); err != nil {
		t.Fatalf("merge import failed: %v", err)
	}

	var name string
	var openOnStart, sortOrder int
	if err := target.GetConn().QueryRow(
		`SELECT name, open_on_start, sort_order FROM browser_bookmarks WHERE url = ?`,
		"https://same.example/",
	).Scan(&name, &openOnStart, &sortOrder); err != nil {
		t.Fatal(err)
	}
	if name != "Current Name" || openOnStart != 0 || sortOrder != 9 {
		t.Fatalf("merge overwrote existing bookmark: name=%q open=%d order=%d", name, openOnStart, sortOrder)
	}

	var newCount int
	if err := target.GetConn().QueryRow(
		`SELECT COUNT(*) FROM browser_bookmarks WHERE url = ?`,
		"https://new.example/",
	).Scan(&newCount); err != nil {
		t.Fatal(err)
	}
	if newCount != 1 {
		t.Fatalf("expected the missing bookmark to be merged, got %d rows", newCount)
	}
	if stats.Skipped == 0 || stats.Imported == 0 {
		t.Fatalf("expected merge to report both skipped and imported rows, got %+v", stats)
	}
}

func TestBackupProxyFileMergeDoesNotTreatEmptyIDsAsDuplicates(t *testing.T) {
	appRoot := t.TempDir()
	stateRoot := t.TempDir()
	t.Setenv("LOCALAPPDATA", stateRoot)
	payloadRoot := filepath.Join(t.TempDir(), "payload")
	proxyPath := filepath.Join(stateRoot, "facade", "proxies.yaml")
	if err := config.SaveProxies(proxyPath, []config.BrowserProxy{
		{ProxyName: "Current", ProxyConfig: "http://current"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProxies(filepath.Join(payloadRoot, "system", "proxies.yaml"), []config.BrowserProxy{
		{ProxyName: "First", ProxyConfig: "http://first"},
		{ProxyName: "Second", ProxyConfig: "http://second"},
	}); err != nil {
		t.Fatal(err)
	}

	app := NewApp(appRoot)
	stats := &backupMergeStats{}
	if err := app.backupMergeProxiesFile(payloadRoot, false, stats); err != nil {
		t.Fatal(err)
	}
	merged, err := config.LoadProxies(proxyPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(merged) != 3 || stats.Imported != 2 {
		t.Fatalf("expected both proxies with empty IDs to merge, got len=%d stats=%+v", len(merged), stats)
	}
}

func TestMigrateToSQLiteDoesNotSeedProfilesOrTagsWhenDatabaseIsEmpty(t *testing.T) {
	db := openMigratedTestDB(t, "profiles.db")
	cfg := config.DefaultConfig()
	cfg.Browser.Profiles = []config.BrowserProfileConfig{
		{
			ProfileId:   "legacy-default",
			ProfileName: "默认实例",
			UserDataDir: "default",
			Tags:        []string{"默认"},
		},
	}
	mgr := &browser.Manager{
		Config:      cfg,
		ProfileDAO:  browser.NewSQLiteProfileDAO(db.GetConn()),
		ProxyDAO:    browser.NewSQLiteProxyDAO(db.GetConn()),
		CoreDAO:     browser.NewSQLiteCoreDAO(db.GetConn()),
		BookmarkDAO: browser.NewSQLiteBookmarkDAO(db.GetConn()),
	}
	app := &App{config: cfg, db: db, browserMgr: mgr}

	app.migrateToSQLite()
	profiles, err := mgr.ProfileDAO.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 0 {
		t.Fatalf("expected an empty profile database to remain empty, got %d profiles", len(profiles))
	}
	if tags := mgr.GetAllTags(); len(tags) != 0 {
		t.Fatalf("expected no tags without profiles, got %v", tags)
	}
}

func TestFactoryResetSeedsBookmarksExplicitly(t *testing.T) {
	db := openMigratedTestDB(t, "factory-bookmarks.db")
	cfg := config.DefaultConfig()
	app := &App{
		config: cfg,
		browserMgr: &browser.Manager{
			Config:      cfg,
			BookmarkDAO: browser.NewSQLiteBookmarkDAO(db.GetConn()),
		},
	}

	if err := app.backupSeedFactoryBookmarks(); err != nil {
		t.Fatal(err)
	}
	bookmarks := app.BookmarkList()
	if len(bookmarks) != len(defaultBookmarkList) {
		t.Fatalf("expected %d factory bookmarks, got %d", len(defaultBookmarkList), len(bookmarks))
	}
	for i := range defaultBookmarkList {
		if bookmarks[i].Name != defaultBookmarkList[i].Name || bookmarks[i].URL != defaultBookmarkList[i].URL {
			t.Fatalf("factory bookmark %d mismatch: got %+v want %+v", i, bookmarks[i], defaultBookmarkList[i])
		}
	}
}

func sourcePath(t *testing.T, db *database.DB) string {
	t.Helper()
	var path string
	if err := db.GetConn().QueryRow(`PRAGMA database_list`).Scan(new(int), new(string), &path); err != nil {
		t.Fatal(err)
	}
	return path
}
