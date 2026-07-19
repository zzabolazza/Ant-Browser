package database

import (
	"path/filepath"
	"testing"
)

func TestMigrateRemovesDeprecatedColumns(t *testing.T) {
	db := openBackupTestDB(t, "cleanup.db")

	deprecated := map[string][]string{
		"browser_proxies": {
			"dns_servers", "source_id", "source_url", "source_name_prefix",
			"source_auto_refresh", "source_refresh_interval_m", "source_last_refresh_at",
			"preferred_kernel",
		},
		"browser_profiles": {"proxy_bind_source_id", "proxy_bind_source_url", "deleted_at"},
	}
	for table, names := range deprecated {
		columns, err := testTableColumns(db, table)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			if _, ok := columns[name]; ok {
				t.Fatalf("deprecated column %s.%s still exists", table, name)
			}
		}
	}

	profileColumns, err := testTableColumns(db, "browser_profiles")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"proxy_bind_name", "proxy_bind_updated_at"} {
		if _, ok := profileColumns[name]; !ok {
			t.Fatalf("active profile column %s was removed", name)
		}
	}
}

func TestRecycleBinRemovalPurgesSoftDeletedRowsAndKeepsActiveRows(t *testing.T) {
	db, err := NewDB(filepath.Join(t.TempDir(), "upgrade.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.GetConn().Exec(`CREATE TABLE schema_migrations (
		version INTEGER PRIMARY KEY, desc TEXT NOT NULL DEFAULT '', applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.version >= 14 {
			break
		}
		if err := db.applyMigration(migration); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.GetConn().Exec(`
		INSERT INTO browser_profiles (profile_id, profile_name, user_data_dir, fingerprint_args, launch_args, tags, keywords, created_at, updated_at, deleted_at)
		VALUES ('active', 'Active', 'active', '[]', '[]', '[]', '[]', 'created', 'updated', ''),
		       ('trashed', 'Trashed', 'trashed', '[]', '[]', '[]', '[]', 'created', 'updated', 'deleted');
		INSERT INTO launch_codes (profile_id, code) VALUES ('trashed', 'trash-code');
		INSERT INTO browser_profile_extension_settings (profile_id, configured) VALUES ('trashed', 1);
	`); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}

	var activeCount, trashedCount, relatedCount int
	if err := db.GetConn().QueryRow(`SELECT COUNT(*) FROM browser_profiles WHERE profile_id='active'`).Scan(&activeCount); err != nil {
		t.Fatal(err)
	}
	if err := db.GetConn().QueryRow(`SELECT COUNT(*) FROM browser_profiles WHERE profile_id='trashed'`).Scan(&trashedCount); err != nil {
		t.Fatal(err)
	}
	if err := db.GetConn().QueryRow(`
		SELECT (SELECT COUNT(*) FROM launch_codes WHERE profile_id='trashed') +
		       (SELECT COUNT(*) FROM browser_profile_extension_settings WHERE profile_id='trashed')`).Scan(&relatedCount); err != nil {
		t.Fatal(err)
	}
	if activeCount != 1 || trashedCount != 0 || relatedCount != 0 {
		t.Fatalf("unexpected recycle-bin cleanup: active=%d trashed=%d related=%d", activeCount, trashedCount, relatedCount)
	}
}

func testTableColumns(db *DB, table string) (map[string]struct{}, error) {
	rows, err := db.GetConn().Query(`PRAGMA table_info("` + table + `")`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := map[string]struct{}{}
	for rows.Next() {
		var cid, notNull, pk int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns[name] = struct{}{}
	}
	return columns, rows.Err()
}
