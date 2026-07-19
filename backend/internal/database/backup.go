package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sqlite "modernc.org/sqlite"
)

type onlineBackuper interface {
	NewBackup(string) (*sqlite.Backup, error)
	NewRestore(string) (*sqlite.Backup, error)
}

// ValidateSnapshot checks that a database snapshot is readable and internally consistent.
func ValidateSnapshot(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("snapshot path is empty")
	}
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetMaxOpenConns(1)
	var result string
	if err := conn.QueryRow(`PRAGMA integrity_check`).Scan(&result); err != nil {
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(result), "ok") {
		return fmt.Errorf("integrity check failed: %s", result)
	}
	return nil
}

// BackupTo creates a transactionally consistent snapshot of the live database.
func (db *DB) BackupTo(dstPath string) error {
	if db == nil || db.conn == nil {
		return fmt.Errorf("database is not initialized")
	}
	dstPath = strings.TrimSpace(dstPath)
	if dstPath == "" {
		return fmt.Errorf("backup destination is empty")
	}
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}
	if err := os.Remove(dstPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return db.runOnlineCopy(dstPath, false)
}

// RestoreFrom replaces the complete live database with the supplied snapshot.
func (db *DB) RestoreFrom(srcPath string) error {
	if db == nil || db.conn == nil {
		return fmt.Errorf("database is not initialized")
	}
	srcPath = strings.TrimSpace(srcPath)
	if srcPath == "" {
		return fmt.Errorf("restore source is empty")
	}
	if info, err := os.Stat(srcPath); err != nil {
		return err
	} else if info.IsDir() {
		return fmt.Errorf("restore source is a directory: %s", srcPath)
	}
	return db.runOnlineCopy(srcPath, true)
}

func (db *DB) runOnlineCopy(path string, restore bool) error {
	ctx := context.Background()
	conn, err := db.conn.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.Raw(func(driverConn any) error {
		backuper, ok := driverConn.(onlineBackuper)
		if !ok {
			return fmt.Errorf("sqlite driver does not support online backup")
		}
		var backup *sqlite.Backup
		if restore {
			backup, err = backuper.NewRestore(path)
		} else {
			backup, err = backuper.NewBackup(path)
		}
		if err != nil {
			return err
		}
		finished := false
		defer func() {
			if !finished {
				_ = backup.Finish()
			}
		}()
		for more := true; more; {
			more, err = backup.Step(-1)
			if err != nil {
				return err
			}
		}
		if err := backup.Finish(); err != nil {
			return err
		}
		finished = true
		return nil
	})
}
