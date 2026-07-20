package backend

import (
	"facade/backend/internal/backup"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// BackupInitializeSystem 初始化系统到最开始状态。
func (a *App) BackupInitializeSystem() (map[string]interface{}, error) {
	a.maintenanceMu.Lock()
	defer a.maintenanceMu.Unlock()

	a.backupStopRuntimeForMaintenance()
	rollbackPath, cleanup, err := a.backupCreateRollbackPackage()
	if err != nil {
		return nil, err
	}
	defer cleanup()
	result, err := a.backupInitializeLocked(true)
	if err == nil {
		return result, nil
	}
	if rollbackErr := a.backupRestoreRollbackPackage(rollbackPath); rollbackErr != nil {
		return nil, fmt.Errorf("%w；自动回滚失败: %v", err, rollbackErr)
	}
	return nil, fmt.Errorf("%w；已自动恢复到操作前状态", err)
}

// BackupExportPackage 将全量配置与数据导出为本地加密备份。
func (a *App) BackupExportPackage(password string) (map[string]interface{}, error) {
	a.maintenanceMu.Lock()
	defer a.maintenanceMu.Unlock()

	if a.ctx == nil {
		return nil, fmt.Errorf("应用上下文未初始化")
	}
	if err := backupValidateEncryptionPassword(password); err != nil {
		return nil, err
	}
	a.backupEmitExportProgress("starting", 0, "等待选择导出路径...")

	defaultName := fmt.Sprintf("facade-backup-%s.facade", time.Now().Format("20060102-150405"))
	savePath, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "导出加密备份",
		DefaultFilename: defaultName,
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "Facade 加密备份 (*.facade)", Pattern: "*.facade"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("打开保存对话框失败: %w", err)
	}
	if strings.TrimSpace(savePath) == "" {
		a.backupEmitExportProgress("cancelled", 0, "已取消导出")
		return map[string]interface{}{
			"cancelled": true,
			"message":   "已取消导出",
		}, nil
	}
	savePath = backupEnsureEncryptedSuffix(savePath)
	result, encryptedPath, cleanup, err := a.backupCreateEncryptedExport(password)
	if err != nil {
		a.backupEmitExportProgress("error", 100, fmt.Sprintf("导出失败: %v", err))
		return nil, err
	}
	defer cleanup()
	if err := backupReplaceFile(encryptedPath, savePath); err != nil {
		a.backupEmitExportProgress("error", 100, fmt.Sprintf("导出失败: %v", err))
		return nil, err
	}
	a.backupEmitExportProgress("done", 100, "加密备份导出完成")
	result["zipPath"] = savePath
	result["message"] = "加密备份导出完成"
	return result, nil
}

// BackupExportPackageToWebDAV 将全量加密备份直接上传到已配置的 WebDAV。
func (a *App) BackupExportPackageToWebDAV(password string) (map[string]interface{}, error) {
	a.maintenanceMu.Lock()
	defer a.maintenanceMu.Unlock()
	if a.ctx == nil {
		return nil, fmt.Errorf("应用上下文未初始化")
	}
	if err := backupValidateEncryptionPassword(password); err != nil {
		return nil, err
	}
	if a.config == nil || strings.TrimSpace(a.config.Backup.WebDAV.URL) == "" {
		return nil, fmt.Errorf("请先配置 WebDAV")
	}
	a.backupEmitExportProgress("starting", 0, "正在准备 WebDAV 加密备份...")
	result, encryptedPath, cleanup, err := a.backupCreateEncryptedExport(password)
	if err != nil {
		a.backupEmitExportProgress("error", 100, fmt.Sprintf("导出失败: %v", err))
		return nil, err
	}
	defer cleanup()
	fileName := fmt.Sprintf("facade-backup-%s.facade", time.Now().Format("20060102-150405"))
	a.backupEmitExportProgress("uploading", 96, "正在上传到 WebDAV...")
	remoteURL, err := a.backupUploadWebDAV(a.ctx, encryptedPath, fileName)
	if err != nil {
		a.backupEmitExportProgress("error", 100, fmt.Sprintf("上传失败: %v", err))
		return nil, err
	}
	a.backupEmitExportProgress("done", 100, "加密备份已上传到 WebDAV")
	result["remotePath"] = remoteURL
	result["message"] = "加密备份已上传到 WebDAV"
	return result, nil
}

// BackupPickImportFile 打开文件选择对话框，返回待导入的加密备份路径。
func (a *App) BackupPickImportFile() (map[string]interface{}, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("应用上下文未初始化")
	}
	zipPath, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择加密备份",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "Facade 加密备份 (*.facade)", Pattern: "*.facade"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("打开文件对话框失败: %w", err)
	}
	if strings.TrimSpace(zipPath) == "" {
		return map[string]interface{}{
			"cancelled": true,
			"message":   "已取消选择",
		}, nil
	}
	return map[string]interface{}{
		"cancelled": false,
		"path":      zipPath,
	}, nil
}

// BackupImportPackage 从已选择的加密备份加载配置与数据。
// resetFirst=true: 按备份内容执行完整恢复。
// resetFirst=false: 保留现有数据并执行判重合并。
func (a *App) BackupImportPackage(resetFirst bool, password string, zipPath string) (map[string]interface{}, error) {
	a.maintenanceMu.Lock()
	defer a.maintenanceMu.Unlock()

	if a.ctx == nil {
		return nil, fmt.Errorf("应用上下文未初始化")
	}
	if err := backupValidateEncryptionPassword(password); err != nil {
		return nil, err
	}
	zipPath = strings.TrimSpace(zipPath)
	if zipPath == "" {
		return nil, fmt.Errorf("请先选择备份文件")
	}
	a.backupEmitImportProgress("preparing", 5, "正在解密并校验备份包...")
	tmpDir, err := os.MkdirTemp("", "facade-backup-decrypt-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	decryptedPath := filepath.Join(tmpDir, "package.zip")
	if err := backup.DecryptFile(zipPath, decryptedPath, password); err != nil {
		return nil, fmt.Errorf("无法解密备份: %w", err)
	}

	result, importErr := a.backupImportFromPathLocked(decryptedPath, resetFirst)
	if importErr != nil {
		a.backupEmitImportProgress("error", 100, fmt.Sprintf("加载失败: %v", importErr))
		return nil, importErr
	}
	result["zipPath"] = zipPath
	return result, nil
}

func (a *App) backupCreateEncryptedExport(password string) (map[string]interface{}, string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "facade-encrypted-export-*")
	if err != nil {
		return nil, "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }
	a.backupEmitExportProgress("preparing", 8, "正在收集导出范围...")
	scope, err := a.BackupGetScopeDefinition()
	if err != nil {
		cleanup()
		return nil, "", func() {}, err
	}
	scope, cleanupSnapshot, err := a.backupPrepareDatabaseSnapshot(scope)
	if err != nil {
		cleanup()
		return nil, "", func() {}, err
	}
	defer cleanupSnapshot()
	manifest := backup.BuildManifest(scope, a.appName(), a.appVersion(), time.Now())
	plainPath := filepath.Join(tmpDir, "package.zip")
	emitPackageProgress := func(phase string, progress int, message string, meta *backupProgressMeta) {
		if phase == "done" {
			return
		}
		if progress > 90 {
			progress = 90
		}
		a.backupEmitExportProgressMeta(phase, progress, message, meta)
	}
	includedEntries, skippedEntries, fileCount, err := backupWritePackageZip(plainPath, scope, manifest, emitPackageProgress)
	if err != nil {
		cleanup()
		return nil, "", func() {}, err
	}
	a.backupEmitExportProgress("encrypting", 92, "正在加密备份内容...")
	encryptedPath := filepath.Join(tmpDir, "package.facade")
	if err := backup.EncryptFile(plainPath, encryptedPath, password); err != nil {
		cleanup()
		return nil, "", func() {}, err
	}
	_ = os.Remove(plainPath)
	return map[string]interface{}{
		"cancelled":       false,
		"includedEntries": includedEntries,
		"skippedEntries":  skippedEntries,
		"fileCount":       fileCount,
		"encrypted":       true,
	}, encryptedPath, cleanup, nil
}

func backupValidateEncryptionPassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("备份密码不能为空")
	}
	if len([]rune(password)) < 8 {
		return fmt.Errorf("备份密码至少需要 8 个字符")
	}
	return nil
}

func backupEnsureEncryptedSuffix(path string) string {
	if strings.EqualFold(filepath.Ext(path), ".facade") {
		return path
	}
	return path + ".facade"
}

func backupReplaceFile(srcPath, dstPath string) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}
	tmpPath := dstPath + ".tmp"
	_ = os.Remove(tmpPath)
	if err := backupCopyFile(srcPath, tmpPath); err != nil {
		return err
	}
	rollbackPath := dstPath + ".rollback-" + uuid.NewString()
	hadExisting := false
	if _, err := os.Stat(dstPath); err == nil {
		if err := os.Rename(dstPath, rollbackPath); err != nil {
			_ = os.Remove(tmpPath)
			return err
		}
		hadExisting = true
	} else if !os.IsNotExist(err) {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, dstPath); err != nil {
		_ = os.Remove(tmpPath)
		if hadExisting {
			if rollbackErr := os.Rename(rollbackPath, dstPath); rollbackErr != nil {
				return fmt.Errorf("写入备份失败: %v；恢复原文件失败: %w", err, rollbackErr)
			}
		}
		return err
	}
	if hadExisting {
		if err := os.Remove(rollbackPath); err != nil {
			return fmt.Errorf("备份已写入，但清理原文件失败: %w", err)
		}
	}
	return nil
}
