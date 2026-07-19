package backend

import (
	"facade/backend/internal/config"
	"facade/backend/internal/logger"
	"fmt"
	"os"
	"strings"
)

func (a *App) backupInitializeLocked(applyReload bool) (map[string]interface{}, error) {
	log := logger.New("Backup")
	a.backupStopRuntimeForMaintenance()

	defaultCfg := config.DefaultConfig()
	oldCfg := a.config
	if oldCfg == nil {
		oldCfg = config.DefaultConfig()
	}
	activeDBPath := a.backupResolveDBPath(oldCfg)
	keepFiles := map[string]struct{}{
		backupNormalizePath(activeDBPath):          {},
		backupNormalizePath(activeDBPath + "-wal"): {},
		backupNormalizePath(activeDBPath + "-shm"): {},
	}

	if err := defaultCfg.Save(a.resolveAppPath("config.yaml")); err != nil {
		return nil, fmt.Errorf("写入默认配置失败: %w", err)
	}
	a.config = defaultCfg
	a.applyRuntimeConfig(defaultCfg.Runtime)
	if err := os.Remove(a.resolveAppPath("proxies.yaml")); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("删除代理配置失败: %w", err)
	}

	if applyReload {
		if err := a.backupResetDatabaseToFactory(); err != nil {
			return nil, err
		}
	}
	if applyReload {
		if err := a.backupSeedFactoryBookmarks(); err != nil {
			return nil, fmt.Errorf("恢复出厂书签失败: %w", err)
		}
	}

	cleared := make([]string, 0, 3)
	dataRoot := a.resolveAppPath("data")
	if logDir := a.backupResolveLogDir(oldCfg); logDir != "" && backupPathWithin(logDir, dataRoot) {
		keepFiles[backupNormalizePath(logDir)] = struct{}{}
	}
	if err := backupRemoveContentsExcept(dataRoot, keepFiles); err != nil {
		return nil, fmt.Errorf("清理应用数据目录失败: %w", err)
	}
	cleared = append(cleared, dataRoot)
	oldUserRoot := a.backupResolveUserDataRoot(oldCfg)
	newUserRoot := a.backupResolveUserDataRoot(defaultCfg)
	for _, p := range backupUniqueNonEmpty([]string{oldUserRoot, newUserRoot}) {
		if backupSamePath(p, dataRoot) {
			continue
		}
		if err := backupRemoveContentsExcept(p, keepFiles); err != nil {
			return nil, fmt.Errorf("清理浏览器用户数据目录失败(%s): %w", p, err)
		}
		cleared = append(cleared, p)
	}

	if applyReload {
		if err := a.backupReloadAfterMutation(true); err != nil {
			return nil, err
		}
	}

	log.Info("系统已恢复出厂设置", logger.F("cleared_dirs", strings.Join(cleared, ";")))
	return map[string]interface{}{
		"cancelled":   false,
		"resetDone":   true,
		"clearedDirs": cleared,
		"message":     "系统已恢复出厂设置",
	}, nil
}

func (a *App) backupSeedFactoryBookmarks() error {
	items := append([]BrowserBookmark{}, defaultBookmarkList...)
	if a.browserMgr != nil && a.browserMgr.BookmarkDAO != nil {
		return a.browserMgr.BookmarkDAO.ReplaceAll(items)
	}
	if a.config == nil {
		return nil
	}
	a.config.Browser.DefaultBookmarks = items
	return a.config.Save(a.resolveAppPath("config.yaml"))
}
