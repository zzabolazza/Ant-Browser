package database

import (
	"path/filepath"
	"testing"
)

func openBackupTestDB(t *testing.T, name string) *DB {
	t.Helper()
	db, err := NewDB(filepath.Join(t.TempDir(), name))
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

func TestBackupAndRestoreReplaceCompleteDatabase(t *testing.T) {
	source := openBackupTestDB(t, "source.db")
	if _, err := source.GetConn().Exec(`
		CREATE TABLE future_module (id INTEGER PRIMARY KEY, value TEXT NOT NULL);
		INSERT INTO future_module (id, value) VALUES (7, 'from-backup');
		INSERT INTO browser_bookmarks (id, name, url, open_on_start, sort_order)
		VALUES (42, 'Saved', 'https://saved.example/', 1, 9);
	`); err != nil {
		t.Fatal(err)
	}

	snapshot := filepath.Join(t.TempDir(), "snapshot.db")
	if err := source.BackupTo(snapshot); err != nil {
		t.Fatal(err)
	}

	target := openBackupTestDB(t, "target.db")
	if _, err := target.GetConn().Exec(`
		CREATE TABLE target_only (value TEXT);
		INSERT INTO target_only VALUES ('must disappear');
		INSERT INTO browser_bookmarks (name, url, sort_order)
		VALUES ('Current', 'https://current.example/', 0);
	`); err != nil {
		t.Fatal(err)
	}
	if err := target.RestoreFrom(snapshot); err != nil {
		t.Fatal(err)
	}

	var value string
	if err := target.GetConn().QueryRow(`SELECT value FROM future_module WHERE id=7`).Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != "from-backup" {
		t.Fatalf("unexpected restored value: %q", value)
	}
	var bookmarkCount int
	if err := target.GetConn().QueryRow(`SELECT COUNT(*) FROM browser_bookmarks WHERE id=42 AND open_on_start=1 AND sort_order=9`).Scan(&bookmarkCount); err != nil {
		t.Fatal(err)
	}
	if bookmarkCount != 1 {
		t.Fatalf("bookmark row was not restored exactly: %d", bookmarkCount)
	}
	var targetOnlyCount int
	if err := target.GetConn().QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='target_only'`).Scan(&targetOnlyCount); err != nil {
		t.Fatal(err)
	}
	if targetOnlyCount != 0 {
		t.Fatal("restore retained a table that was not present in the backup")
	}
	var integrity string
	if err := target.GetConn().QueryRow(`PRAGMA integrity_check`).Scan(&integrity); err != nil {
		t.Fatal(err)
	}
	if integrity != "ok" {
		t.Fatalf("restored database integrity check failed: %s", integrity)
	}
}
