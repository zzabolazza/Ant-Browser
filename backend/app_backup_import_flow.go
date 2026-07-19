package backend

import (
	"errors"
	"facade/backend/internal/database"
	"fmt"
	"os"
	"path/filepath"
)

func (a *App) backupImportFromPathLocked(zipPath string, resetFirst bool) (result map[string]interface{}, retErr error) {
	a.backupStopRuntimeForMaintenance()
	a.backupEmitImportProgress("preparing", 10, "正在解压并校验备份包...")

	extractRoot, manifest, err := backupExtractAndValidate(zipPath)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(extractRoot)
	a.backupEmitImportProgress("preparing", 20, "备份包校验通过，开始加载数据...")

	componentEntries := backupDetectPresentManifestEntries(extractRoot, manifest)
	issueTracker := newBackupImportTracker(componentEntries)
	stats := &backupMergeStats{}
	payloadRoot := filepath.Join(extractRoot, "payload")
	dbSrc := backupFindDatabaseFile(payloadRoot)
	var preservePaths []string
	if resetFirst {
		preservePaths = []string{a.backupResolveLogDir(a.config)}
	}

	a.backupEmitImportProgress("importing", 24, "正在解析备份配置...")
	incomingCfg, hasIncomingCfg, configLoadErr := backupLoadIncomingConfig(payloadRoot)
	if configLoadErr != nil && resetFirst {
		return nil, fmt.Errorf("完整恢复失败：解析备份配置失败: %w", configLoadErr)
	}
	if configLoadErr != nil {
		incomingCfg = nil
		hasIncomingCfg = false
	}
	if resetFirst {
		for _, entry := range manifest.Entries {
			if !entry.Required {
				continue
			}
			if _, ok := componentEntries[entry.ID]; !ok {
				return nil, fmt.Errorf("完整恢复失败：备份包缺少必需组件 %s", backupResolveManifestComponentName(entry))
			}
		}
		if !hasIncomingCfg {
			return nil, fmt.Errorf("完整恢复失败：备份包缺少 payload/system/config.yaml")
		}
		if dbSrc == "" {
			return nil, fmt.Errorf("完整恢复失败：备份包缺少 SQLite 数据库快照")
		}
		if err := database.ValidateSnapshot(dbSrc); err != nil {
			return nil, fmt.Errorf("完整恢复失败：SQLite 数据库快照校验失败: %w", err)
		}
		if err := backupValidateFullRestorePayload(payloadRoot, incomingCfg); err != nil {
			return nil, fmt.Errorf("完整恢复失败：%w", err)
		}
	}

	var rollbackPath string
	var rollbackCleanup func()
	mutationStarted := false
	defer func() {
		if rollbackCleanup != nil {
			defer rollbackCleanup()
		}
		if resetFirst && mutationStarted && retErr != nil && rollbackPath != "" {
			if rollbackErr := a.backupRestoreRollbackPackage(rollbackPath); rollbackErr != nil {
				retErr = errors.Join(retErr, fmt.Errorf("自动回滚失败: %w", rollbackErr))
			} else {
				retErr = fmt.Errorf("%w；已自动恢复到操作前状态", retErr)
			}
		}
	}()

	var resetIssue error
	recordIssue := func(componentID, componentName string, issue error) {
		issueTracker.RecordIssue(componentID, componentName, issue)
		if resetFirst && resetIssue == nil && issue != nil {
			resetIssue = fmt.Errorf("完整恢复失败（%s）: %w", componentName, issue)
		}
	}
	if configLoadErr != nil {
		recordIssue("system_config_main", "主配置文件", fmt.Errorf("解析配置失败: %w", configLoadErr))
	}

	if resetFirst {
		rollbackPath, rollbackCleanup, err = a.backupCreateRollbackPackage()
		if err != nil {
			return nil, err
		}
		mutationStarted = true
		a.backupEmitImportProgress("preparing", 30, "正在准备完整恢复...")
		if _, err := a.backupInitializeLocked(false); err != nil {
			return nil, err
		}
		a.backupEmitImportProgress("preparing", 40, "恢复环境已准备完成，正在加载备份内容...")
	}

	a.backupEmitImportProgress("importing", 50, "正在解析备份配置...")
	if hasIncomingCfg {
		a.backupEmitImportProgress("importing", 58, "正在应用系统配置...")
		if err := a.backupApplyIncomingConfig(incomingCfg, resetFirst); err != nil {
			recordIssue("system_config_main", "主配置文件", err)
		}
	}
	if resetIssue != nil {
		return nil, resetIssue
	}

	proxyProgressMessage := "正在合并代理配置..."
	if resetFirst {
		proxyProgressMessage = "正在恢复代理配置..."
	}
	a.backupEmitImportProgress("importing", 66, proxyProgressMessage)
	if err := a.backupMergeProxiesFile(payloadRoot, resetFirst, stats); err != nil {
		recordIssue("system_config_proxies", "代理配置文件", err)
	}
	if resetIssue != nil {
		return nil, resetIssue
	}

	if dbSrc != "" {
		if resetFirst {
			a.backupEmitImportProgress("importing", 76, "正在完整恢复数据库快照...")
			if err := a.backupRestoreDatabaseSnapshot(dbSrc); err != nil {
				recordIssue("database_sqlite_main", "SQLite 主数据库", err)
			} else {
				stats.Imported++
			}
		} else {
			a.backupEmitImportProgress("importing", 76, "正在合并数据库数据...")
			if err := a.backupMergeDatabaseFromSource(dbSrc, false, stats); err != nil {
				recordIssue("database_sqlite_main", "SQLite 主数据库", err)
			}
		}
	} else if _, ok := componentEntries["database_sqlite_main"]; ok {
		recordIssue("database_sqlite_main", "SQLite 主数据库", fmt.Errorf("备份包缺少数据库文件"))
	}
	if resetIssue != nil {
		return nil, resetIssue
	}

	a.backupEmitImportProgress("importing", 86, "正在同步文件数据...")
	fileTreeCfg := incomingCfg
	if resetFirst && incomingCfg != nil && a.browserMgr != nil && a.browserMgr.CoreDAO != nil {
		if cores, err := a.browserMgr.CoreDAO.List(); err == nil {
			cloned := *incomingCfg
			cloned.Browser = incomingCfg.Browser
			cloned.Browser.Cores = cores
			fileTreeCfg = &cloned
		}
	}
	if resetFirst {
		preservePaths = append(preservePaths, a.backupResolveLogDir(incomingCfg))
	}
	a.backupImportFileTrees(payloadRoot, fileTreeCfg, resetFirst, stats, preservePaths, recordIssue)
	if resetIssue != nil {
		return nil, resetIssue
	}

	a.backupEmitImportProgress("importing", 94, "正在刷新运行时配置...")
	if err := a.backupReloadAfterMutation(!resetFirst); err != nil {
		return nil, err
	}

	totalComponents, successCount, failedCount, partial := issueTracker.Summary()
	message := "加载完成"
	if partial {
		message = fmt.Sprintf("加载完成（部分成功）：成功 %d 个模块，异常 %d 个模块", successCount, failedCount)
	}
	a.backupEmitImportProgress("done", 100, message)
	mutationStarted = false

	return map[string]interface{}{
		"cancelled":        false,
		"zipPath":          zipPath,
		"resetFirst":       resetFirst,
		"imported":         stats.Imported,
		"skipped":          stats.Skipped,
		"conflicts":        stats.Conflicts,
		"partial":          partial,
		"componentTotal":   totalComponents,
		"componentSuccess": successCount,
		"componentFailed":  failedCount,
		"failedComponents": issueTracker.FailedComponents(),
		"message":          message,
	}, nil
}
