package backend

import (
	"errors"
	"facade/backend/internal/backup"
	"facade/backend/internal/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) backupCreateRollbackPackage() (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "facade-restore-rollback-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }
	scope, err := a.BackupGetScopeDefinition()
	if err != nil {
		cleanup()
		return "", func() {}, err
	}
	scope, cleanupSnapshot, err := a.backupPrepareDatabaseSnapshot(scope)
	if err != nil {
		cleanup()
		return "", func() {}, err
	}
	defer cleanupSnapshot()
	zipPath := filepath.Join(tmpDir, "rollback.zip")
	manifest := backup.BuildManifest(scope, a.appName(), a.appVersion(), time.Now())
	if _, _, _, err := backupWritePackageZip(zipPath, scope, manifest, nil); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("创建恢复前回滚快照失败: %w", err)
	}
	return zipPath, cleanup, nil
}

func (a *App) backupRestoreRollbackPackage(zipPath string) error {
	extractRoot, _, err := backupExtractAndValidate(zipPath)
	if err != nil {
		return err
	}
	defer os.RemoveAll(extractRoot)
	payloadRoot := filepath.Join(extractRoot, "payload")
	cfg, hasCfg, err := backupLoadIncomingConfig(payloadRoot)
	if err != nil || !hasCfg {
		if err == nil {
			err = fmt.Errorf("回滚包缺少主配置")
		}
		return err
	}
	if err := a.backupApplyIncomingConfig(cfg, true); err != nil {
		return err
	}
	if err := a.backupMergeProxiesFile(payloadRoot, true, &backupMergeStats{}); err != nil {
		return err
	}
	dbSrc := backupFindDatabaseFile(payloadRoot)
	if dbSrc == "" {
		return fmt.Errorf("回滚包缺少 SQLite 数据库快照")
	}
	if err := a.backupRestoreDatabaseSnapshot(dbSrc); err != nil {
		return err
	}
	fileTreeCfg := cfg
	if a.browserMgr != nil && a.browserMgr.CoreDAO != nil {
		if cores, err := a.browserMgr.CoreDAO.List(); err == nil {
			cloned := *cfg
			cloned.Browser = cfg.Browser
			cloned.Browser.Cores = cores
			fileTreeCfg = &cloned
		}
	}
	preservePaths := []string{a.backupResolveLogDir(cfg)}
	var restoreIssues []error
	a.backupImportFileTrees(payloadRoot, fileTreeCfg, true, &backupMergeStats{}, preservePaths,
		func(_ string, componentName string, issue error) {
			if issue != nil {
				restoreIssues = append(restoreIssues, fmt.Errorf("%s: %w", componentName, issue))
			}
		})
	if len(restoreIssues) > 0 {
		return errors.Join(restoreIssues...)
	}
	return a.backupReloadAfterMutation(false)
}

func backupValidateFullRestorePayload(payloadRoot string, cfg *config.Config) error {
	required := []struct {
		path string
		dir  bool
		name string
	}{
		{path: filepath.Join(payloadRoot, "system", "config.yaml"), name: "主配置"},
		{path: filepath.Join(payloadRoot, "app", "data"), dir: true, name: "应用数据目录"},
	}
	for _, item := range required {
		info, err := os.Stat(item.path)
		if err != nil {
			return fmt.Errorf("备份包缺少%s: %w", item.name, err)
		}
		if item.dir != info.IsDir() {
			return fmt.Errorf("备份包中%s类型不正确", item.name)
		}
	}
	if backupFindDatabaseFile(payloadRoot) == "" {
		return fmt.Errorf("备份包缺少 SQLite 数据库快照")
	}
	if backupUserDataStoredSeparately(cfg) {
		path := filepath.Join(payloadRoot, "browser", "user-data")
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			if err == nil {
				err = fmt.Errorf("不是目录")
			}
			return fmt.Errorf("备份包缺少独立浏览器用户数据目录: %w", err)
		}
	}
	return nil
}

func backupUserDataStoredSeparately(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	root := strings.TrimSpace(filepath.ToSlash(cfg.Browser.UserDataRoot))
	if root == "" || root == "." || strings.EqualFold(root, "data") {
		return false
	}
	if filepath.IsAbs(cfg.Browser.UserDataRoot) {
		return true
	}
	clean := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(cfg.Browser.UserDataRoot)), "./")
	return !strings.EqualFold(clean, "data") && !strings.HasPrefix(strings.ToLower(clean), "data/")
}
