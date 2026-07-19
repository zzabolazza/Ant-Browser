package backend

import (
	"facade/backend/internal/backup"
	"facade/backend/internal/database"
	"fmt"
	"os"
	"path/filepath"
)

func (a *App) backupPrepareDatabaseSnapshot(scope backup.Scope) (backup.Scope, func(), error) {
	entryIndex := -1
	for i := range scope.Entries {
		if scope.Entries[i].ID == "database_sqlite_main" {
			entryIndex = i
			break
		}
	}
	if entryIndex < 0 {
		return scope, func() {}, nil
	}
	if a.db == nil {
		return scope, func() {}, fmt.Errorf("数据库未初始化，无法创建一致性快照")
	}
	tmpDir, err := os.MkdirTemp("", "facade-backup-db-*")
	if err != nil {
		return scope, func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }
	snapshotPath := filepath.Join(tmpDir, "app.db")
	if err := a.db.BackupTo(snapshotPath); err != nil {
		cleanup()
		return scope, func() {}, fmt.Errorf("创建 SQLite 一致性快照失败: %w", err)
	}
	if err := database.ValidateSnapshot(snapshotPath); err != nil {
		cleanup()
		return scope, func() {}, fmt.Errorf("校验 SQLite 一致性快照失败: %w", err)
	}
	scope.Entries[entryIndex].SourcePath = snapshotPath
	scope.Entries[entryIndex].Exists = true
	return scope, cleanup, nil
}

func (a *App) backupRestoreDatabaseSnapshot(srcPath string) error {
	if a.db == nil {
		return fmt.Errorf("数据库未初始化，无法完整恢复")
	}
	if err := a.db.RestoreFrom(srcPath); err != nil {
		return fmt.Errorf("完整恢复 SQLite 数据库失败: %w", err)
	}
	if err := a.db.Migrate(); err != nil {
		return fmt.Errorf("恢复后升级 SQLite 数据库失败: %w", err)
	}
	return nil
}

func (a *App) backupResetDatabaseToFactory() error {
	if a.db == nil {
		return fmt.Errorf("数据库未初始化，无法恢复出厂设置")
	}
	tmpDir, err := os.MkdirTemp("", "facade-factory-db-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	blankPath := filepath.Join(tmpDir, "app.db")
	blankDB, err := database.NewDB(blankPath)
	if err != nil {
		return err
	}
	if err := blankDB.Migrate(); err != nil {
		_ = blankDB.Close()
		return err
	}
	if err := blankDB.Close(); err != nil {
		return err
	}
	if err := a.db.RestoreFrom(blankPath); err != nil {
		return fmt.Errorf("重建出厂数据库失败: %w", err)
	}
	return a.db.Migrate()
}
