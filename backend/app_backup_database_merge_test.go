package backend

import (
	"path/filepath"
	"testing"

	"facade/backend/internal/database"
)

func TestBackupDatabaseMergePreservesAllCurrentFieldsAndRemapsReferences(t *testing.T) {
	source := openMigratedTestDB(t, "source.db")
	target := openMigratedTestDB(t, "target.db")

	if _, err := target.GetConn().Exec(`INSERT INTO browser_groups (group_id, group_name, parent_id) VALUES ('target-group', 'Shared', '')`); err != nil {
		t.Fatal(err)
	}
	if _, err := target.GetConn().Exec(`INSERT INTO browser_cores (core_id, core_name, core_path) VALUES ('target-core', 'Current', 'C:/shared/core')`); err != nil {
		t.Fatal(err)
	}
	if _, err := target.GetConn().Exec(`INSERT INTO browser_proxies (proxy_id, proxy_name, proxy_config) VALUES ('target-proxy', 'Current', 'socks5://shared')`); err != nil {
		t.Fatal(err)
	}

	if _, err := source.GetConn().Exec(`INSERT INTO browser_groups (group_id, group_name, parent_id) VALUES ('source-group', 'Shared', '')`); err != nil {
		t.Fatal(err)
	}
	if _, err := source.GetConn().Exec(`INSERT INTO browser_cores (core_id, core_name, core_path) VALUES ('source-core', 'Backup', 'C:/shared/core')`); err != nil {
		t.Fatal(err)
	}
	if _, err := source.GetConn().Exec(`INSERT INTO browser_proxies (proxy_id, proxy_name, proxy_config) VALUES ('source-proxy', 'Backup', 'socks5://shared')`); err != nil {
		t.Fatal(err)
	}
	if _, err := source.GetConn().Exec(`
		INSERT INTO browser_proxies (
			proxy_id, proxy_name, proxy_config, group_name,
			last_latency_ms, last_test_ok, last_tested_at, last_ip_health_json, sort_order
		) VALUES ('new-proxy', 'New', 'http://new', 'Imported', 88, 1,
			'tested', '{"ok":true}', 7)`); err != nil {
		t.Fatal(err)
	}
	if _, err := source.GetConn().Exec(`
		INSERT INTO browser_profiles (
			profile_id, profile_name, user_data_dir, core_id, fingerprint_args,
			proxy_id, proxy_config, proxy_bind_name, proxy_bind_updated_at, launch_args, tags, keywords,
			group_id, created_at, updated_at
		) VALUES ('new-profile', 'Imported', 'profiles/new', 'source-core', '[]',
			'source-proxy', 'socks5://shared', 'Bound Proxy', 'bind-time', '[]', '["tag"]', '["keyword"]',
			'source-group', 'created', 'updated')`); err != nil {
		t.Fatal(err)
	}
	if _, err := source.GetConn().Exec(`
		INSERT INTO browser_extensions (
			extension_id, name, version, description, icon_data_url, manifest_json,
			source_url, install_dir, enabled, installed_at, updated_at
		) VALUES ('extension-x', 'Extension', '1.0', 'desc', 'data:image/png;base64,AA', '{}',
			'https://extension.example/', 'extensions/x', 1, 'installed', 'updated')`); err != nil {
		t.Fatal(err)
	}

	app := &App{db: target}
	stats := &backupMergeStats{}
	if err := app.backupMergeDatabaseFromSource(sourcePath(t, source), false, stats); err != nil {
		t.Fatal(err)
	}

	var groupName, testedAt, health string
	var latency, testOK, sortOrder int
	if err := target.GetConn().QueryRow(`
		SELECT group_name, last_latency_ms, last_test_ok, last_tested_at, last_ip_health_json, sort_order
		FROM browser_proxies WHERE proxy_id='new-proxy'`).Scan(
		&groupName, &latency, &testOK, &testedAt, &health, &sortOrder,
	); err != nil {
		t.Fatal(err)
	}
	if groupName != "Imported" || latency != 88 || testOK != 1 || testedAt != "tested" || health != `{"ok":true}` || sortOrder != 7 {
		t.Fatalf("proxy fields were not fully merged: %q %d %d %q %q %d", groupName, latency, testOK, testedAt, health, sortOrder)
	}

	var coreID, proxyID, groupID, bindName, bindUpdatedAt string
	if err := target.GetConn().QueryRow(`
		SELECT core_id, proxy_id, group_id, proxy_bind_name, proxy_bind_updated_at
		FROM browser_profiles WHERE profile_id='new-profile'`).Scan(
		&coreID, &proxyID, &groupID, &bindName, &bindUpdatedAt,
	); err != nil {
		t.Fatal(err)
	}
	if coreID != "target-core" || proxyID != "target-proxy" || groupID != "target-group" {
		t.Fatalf("profile references were not remapped: core=%q proxy=%q group=%q", coreID, proxyID, groupID)
	}
	if bindName != "Bound Proxy" || bindUpdatedAt != "bind-time" {
		t.Fatalf("profile fields were not fully merged: %q %q", bindName, bindUpdatedAt)
	}

	var icon string
	if err := target.GetConn().QueryRow(`SELECT icon_data_url FROM browser_extensions WHERE extension_id='extension-x'`).Scan(&icon); err != nil {
		t.Fatal(err)
	}
	if icon != "data:image/png;base64,AA" {
		t.Fatalf("extension icon was not merged: %q", icon)
	}
}

func TestBackupDatabaseMergeSkipsDependentsForProfileThatWasNotImported(t *testing.T) {
	source := openMigratedTestDB(t, "source.db")
	target := openMigratedTestDB(t, "target.db")
	if _, err := target.GetConn().Exec(`
		INSERT INTO browser_profiles (profile_id, profile_name, user_data_dir, fingerprint_args, launch_args, tags, keywords, created_at, updated_at)
		VALUES ('target-profile', 'Current', 'profiles/shared', '[]', '[]', '[]', '[]', 'created', 'updated')`); err != nil {
		t.Fatal(err)
	}
	if _, err := source.GetConn().Exec(`
		INSERT INTO browser_profiles (profile_id, profile_name, user_data_dir, fingerprint_args, launch_args, tags, keywords, created_at, updated_at)
		VALUES ('source-profile', 'Backup', 'profiles/shared', '[]', '[]', '[]', '[]', 'created', 'updated')`); err != nil {
		t.Fatal(err)
	}
	if _, err := source.GetConn().Exec(`INSERT INTO launch_codes (profile_id, code) VALUES ('source-profile', 'orphan-code')`); err != nil {
		t.Fatal(err)
	}
	if _, err := source.GetConn().Exec(`INSERT INTO browser_profile_extension_settings (profile_id, configured) VALUES ('source-profile', 1)`); err != nil {
		t.Fatal(err)
	}

	app := &App{db: target}
	if err := app.backupMergeDatabaseFromSource(sourcePath(t, source), false, &backupMergeStats{}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := target.GetConn().QueryRow(`
		SELECT (SELECT COUNT(*) FROM launch_codes WHERE profile_id='source-profile') +
		       (SELECT COUNT(*) FROM browser_profile_extension_settings WHERE profile_id='source-profile')`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected dependent rows for a skipped profile to be skipped, got %d", count)
	}
}

func TestBackupDatabaseMergeAcceptsOlderSchemaWithoutOptionalColumns(t *testing.T) {
	source, err := database.NewDB(filepath.Join(t.TempDir(), "old-source.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = source.Close() })
	if _, err := source.GetConn().Exec(`
		CREATE TABLE browser_proxies (
			proxy_id TEXT PRIMARY KEY, proxy_name TEXT NOT NULL, proxy_config TEXT NOT NULL,
			dns_servers TEXT NOT NULL DEFAULT '', sort_order INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO browser_proxies (proxy_id, proxy_name, proxy_config, dns_servers)
		VALUES ('old-proxy', 'Old', 'http://old', '8.8.8.8');
		CREATE TABLE browser_profiles (
			profile_id TEXT PRIMARY KEY, profile_name TEXT NOT NULL, user_data_dir TEXT NOT NULL DEFAULT '',
			fingerprint_args TEXT NOT NULL DEFAULT '[]', proxy_config TEXT NOT NULL DEFAULT '',
			launch_args TEXT NOT NULL DEFAULT '[]', tags TEXT NOT NULL DEFAULT '[]',
			keywords TEXT NOT NULL DEFAULT '[]', created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
			deleted_at TEXT NOT NULL DEFAULT ''
		);
		INSERT INTO browser_profiles (profile_id, profile_name, user_data_dir, created_at, updated_at, deleted_at)
		VALUES ('old-active', 'Active', 'old-active', 'created', 'updated', ''),
		       ('old-trashed', 'Trashed', 'old-trashed', 'created', 'updated', 'deleted');`); err != nil {
		t.Fatal(err)
	}

	target := openMigratedTestDB(t, "target.db")
	app := &App{db: target}
	if err := app.backupMergeDatabaseFromSource(sourcePath(t, source), false, &backupMergeStats{}); err != nil {
		t.Fatalf("old schema merge failed: %v", err)
	}
	var proxyName, proxyConfig string
	if err := target.GetConn().QueryRow(`
		SELECT proxy_name, proxy_config FROM browser_proxies WHERE proxy_id='old-proxy'`,
	).Scan(&proxyName, &proxyConfig); err != nil {
		t.Fatal(err)
	}
	if proxyName != "Old" || proxyConfig != "http://old" {
		t.Fatalf("unexpected old-schema merge: name=%q config=%q", proxyName, proxyConfig)
	}
	var activeCount, trashedCount int
	if err := target.GetConn().QueryRow(`SELECT COUNT(*) FROM browser_profiles WHERE profile_id='old-active'`).Scan(&activeCount); err != nil {
		t.Fatal(err)
	}
	if err := target.GetConn().QueryRow(`SELECT COUNT(*) FROM browser_profiles WHERE profile_id='old-trashed'`).Scan(&trashedCount); err != nil {
		t.Fatal(err)
	}
	if activeCount != 1 || trashedCount != 0 {
		t.Fatalf("obsolete recycle-bin rows were merged: active=%d trashed=%d", activeCount, trashedCount)
	}
}

func TestBackupDatabaseMergeRemapsNestedGroupParents(t *testing.T) {
	source := openMigratedTestDB(t, "source.db")
	target := openMigratedTestDB(t, "target.db")
	if _, err := target.GetConn().Exec(`
		INSERT INTO browser_groups (group_id, group_name, parent_id) VALUES
		  ('target-parent', 'Parent', ''),
		  ('target-child', 'Child', 'target-parent')`); err != nil {
		t.Fatal(err)
	}
	if _, err := source.GetConn().Exec(`
		INSERT INTO browser_groups (group_id, group_name, parent_id) VALUES
		  ('source-parent', 'Parent', ''),
		  ('source-child', 'Child', 'source-parent');
		INSERT INTO browser_profiles
		  (profile_id, profile_name, user_data_dir, fingerprint_args, launch_args, tags, keywords, group_id, created_at, updated_at)
		VALUES ('nested-profile', 'Nested', 'nested-profile', '[]', '[]', '[]', '[]', 'source-child', 'created', 'updated')`); err != nil {
		t.Fatal(err)
	}

	app := &App{db: target}
	if err := app.backupMergeDatabaseFromSource(sourcePath(t, source), false, &backupMergeStats{}); err != nil {
		t.Fatal(err)
	}
	var groupCount int
	if err := target.GetConn().QueryRow(`SELECT COUNT(*) FROM browser_groups`).Scan(&groupCount); err != nil {
		t.Fatal(err)
	}
	if groupCount != 2 {
		t.Fatalf("nested group collision created duplicates: %d", groupCount)
	}
	var profileGroup string
	if err := target.GetConn().QueryRow(`SELECT group_id FROM browser_profiles WHERE profile_id='nested-profile'`).Scan(&profileGroup); err != nil {
		t.Fatal(err)
	}
	if profileGroup != "target-child" {
		t.Fatalf("profile group was not remapped to existing nested group: %q", profileGroup)
	}
}

func TestFactoryDatabaseResetRemovesUnknownTables(t *testing.T) {
	db := openMigratedTestDB(t, "factory.db")
	if _, err := db.GetConn().Exec(`
		CREATE TABLE obsolete_module (id INTEGER PRIMARY KEY, value TEXT);
		INSERT INTO obsolete_module VALUES (1, 'stale');
		INSERT INTO browser_bookmarks (name, url) VALUES ('Stale', 'https://stale.example/')`); err != nil {
		t.Fatal(err)
	}
	app := &App{db: db}
	if err := app.backupResetDatabaseToFactory(); err != nil {
		t.Fatal(err)
	}
	var unknownTable, bookmarks int
	if err := db.GetConn().QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='obsolete_module'`).Scan(&unknownTable); err != nil {
		t.Fatal(err)
	}
	if err := db.GetConn().QueryRow(`SELECT COUNT(*) FROM browser_bookmarks`).Scan(&bookmarks); err != nil {
		t.Fatal(err)
	}
	if unknownTable != 0 || bookmarks != 0 {
		t.Fatalf("factory database reset left data behind: unknownTable=%d bookmarks=%d", unknownTable, bookmarks)
	}
}
