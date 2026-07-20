import { useEffect, useRef, useState } from 'react'
import { ConfirmModal, toast } from '../../shared/components'
import {
  initializeSystemData,
  exportSystemConfig,
  getBackupWebDAVSettings,
  importSystemConfig,
  pickImportBackupFile,
  saveBackupWebDAVSettings,
  type BackupWebDAVSettings,
} from './api'
import { BackupImportModal, BackupSettingsCard } from './components/BackupSettingsCard'
import { ThemeSettingsCard } from './components/ThemeSettingsCard'
import type { BackupExportLogItem, BackupExportProgress } from './progress'
import { useSettingsProgressEffects } from './hooks/useSettingsProgressEffects'
import {
  DEFAULT_THEME,
  flushThemeForBackup,
  getStoredTheme,
  hydrateThemeFromBackend,
  resetThemeMode,
  setThemeMode,
  type ThemeMode,
} from '../../shared/theme/theme'

export function SettingsPage() {
  const [theme, setTheme] = useState<ThemeMode>(() => getStoredTheme())
  const [importModalOpen, setImportModalOpen] = useState(false)
  const [importFilePath, setImportFilePath] = useState('')
  const [initializeConfirmOpen, setInitializeConfirmOpen] = useState(false)
  const [actionLoading, setActionLoading] = useState<'none' | 'init' | 'export' | 'import-reset' | 'import-merge'>('none')
  const [exportProgress, setExportProgress] = useState<BackupExportProgress | null>(null)
  const [importProgress, setImportProgress] = useState<BackupExportProgress | null>(null)
  const [exportLogs, setExportLogs] = useState<BackupExportLogItem[]>([])
  const exportLogsRef = useRef<HTMLDivElement | null>(null)
  const [webDAVSettings, setWebDAVSettings] = useState<BackupWebDAVSettings>({
    url: '', username: '', remoteDir: '', hasPassword: false,
  })
  const [webDAVSaving, setWebDAVSaving] = useState(false)

  useEffect(() => {
    void getBackupWebDAVSettings().then(setWebDAVSettings).catch(() => undefined)
  }, [])

  useEffect(() => {
    void hydrateThemeFromBackend().then(setTheme).catch(() => undefined)
  }, [])

  const handleThemeChange = (nextTheme: ThemeMode) => {
    setTheme(nextTheme)
    setThemeMode(nextTheme)
  }

  useSettingsProgressEffects({
    actionLoading,
    exportLogs,
    exportLogsRef,
    importProgress,
    setExportLogs,
    setExportProgress,
    setImportProgress,
  })

  const handleInitializeSystem = async () => {
    setActionLoading('init')
    try {
      const res = await initializeSystemData()
      if (res.cancelled) {
        toast.info('已取消恢复出厂设置')
        return
      }
      resetThemeMode()
      setTheme(DEFAULT_THEME)
      toast.success(res.message || '已恢复出厂设置')
    } catch (error: any) {
      toast.error(error?.message || '恢复出厂设置失败')
    } finally {
      setActionLoading('none')
    }
  }

  const handleExportSystem = async (target: 'local' | 'webdav', password: string) => {
    setActionLoading('export')
    setExportLogs([])
    setExportProgress({ phase: 'starting', progress: 0, message: '准备导出...' })
    try {
      await flushThemeForBackup()
      const res = await exportSystemConfig(password, target)
      if (res.cancelled) {
        setExportProgress(null)
        setExportLogs([])
        toast.info('已取消导出')
        return
      }
      setExportProgress(prev => prev?.phase === 'done'
        ? prev
        : { phase: 'done', progress: 100, message: res.message || '导出完成' })
      toast.success(res.message || '导出完成')
    } catch (error: any) {
      setExportProgress(prev => ({
        phase: 'error',
        progress: prev?.progress ?? 0,
        message: error?.message || '导出失败',
      }))
      setExportLogs(prev => {
        const timestamp = new Date().toLocaleTimeString('zh-CN', { hour12: false })
        const text = error?.message || '导出失败'
        const next = [...prev, { id: Date.now() + Math.floor(Math.random() * 1000), phase: 'error', time: timestamp, text }]
        return next.length > 120 ? next.slice(next.length - 120) : next
      })
      toast.error(error?.message || '导出失败')
    } finally {
      setActionLoading('none')
    }
  }

  const handleSaveWebDAV = async (settings: BackupWebDAVSettings) => {
    setWebDAVSaving(true)
    try {
      await saveBackupWebDAVSettings(settings)
      const loaded = await getBackupWebDAVSettings()
      setWebDAVSettings(loaded)
      toast.success('WebDAV 设置已保存')
    } catch (error: any) {
      toast.error(error?.message || 'WebDAV 设置保存失败')
    } finally {
      setWebDAVSaving(false)
    }
  }

  const handleOpenImport = async () => {
    if (actionLoading !== 'none') return
    try {
      const picked = await pickImportBackupFile()
      if (picked.cancelled || !picked.path) {
        toast.info('已取消选择备份文件')
        return
      }
      setImportFilePath(picked.path)
      setImportProgress(null)
      setImportModalOpen(true)
    } catch (error: any) {
      toast.error(error?.message || '选择备份文件失败')
    }
  }

  const handleImportSystem = async (resetFirst: boolean, password: string) => {
    if (!importFilePath) {
      toast.error('请先选择备份文件')
      return
    }
    setActionLoading(resetFirst ? 'import-reset' : 'import-merge')
    setImportProgress({
      phase: 'starting',
      progress: 0,
      message: resetFirst ? '正在完整恢复...' : '正在合并导入...',
    })
    try {
      const res = await importSystemConfig(resetFirst, password, importFilePath)
      if (res.cancelled) {
        setImportProgress(null)
        toast.info('已取消加载')
        return
      }
      const imported = res.imported ?? 0
      const skipped = res.skipped ?? 0
      const conflicts = res.conflicts ?? 0
      const componentFailed = Number.isFinite(res.componentFailed) ? Math.max(0, Math.round(res.componentFailed || 0)) : 0
      const componentTotal = Number.isFinite(res.componentTotal) ? Math.max(0, Math.round(res.componentTotal || 0)) : 0
      const failedComponents = Array.isArray(res.failedComponents) ? res.failedComponents : []

      if (res.partial || componentFailed > 0) {
        const moduleNames = failedComponents
          .map(item => (item?.componentName || item?.componentId || '').trim())
          .filter(Boolean)
        const moduleHint = moduleNames.length > 0
          ? `：${moduleNames.slice(0, 3).join('、')}${moduleNames.length > 3 ? ` 等 ${moduleNames.length} 个模块` : ''}`
          : ''
        if (componentTotal > 0) {
          const componentSuccess = Math.max(0, componentTotal - componentFailed)
          toast.warning(`加载完成（部分成功）：模块成功 ${componentSuccess}/${componentTotal}，异常 ${componentFailed}${moduleHint}`)
        } else {
          toast.warning(`加载完成（部分成功）：异常模块 ${componentFailed}${moduleHint}`)
        }
      } else {
        toast.success(`加载完成：导入 ${imported}，跳过 ${skipped}，冲突 ${conflicts}`)
      }
      const restoredTheme = await hydrateThemeFromBackend()
      setTheme(restoredTheme)
      setImportModalOpen(false)
      setImportFilePath('')
      setImportProgress(null)
    } catch (error: any) {
      setImportProgress(prev => ({
        phase: 'error',
        progress: prev?.progress ?? 0,
        message: error?.message || '加载失败',
      }))
      toast.error(error?.message || '加载失败')
    } finally {
      setActionLoading('none')
    }
  }

  return (
    <div className="w-full space-y-5 animate-fade-in">
      <p className="max-w-2xl text-[12.5px] leading-5 text-[var(--color-text-muted)]">
        管理界面主题与本机数据。执行恢复操作前，建议先创建一份完整备份。
      </p>

      <ThemeSettingsCard value={theme} onChange={handleThemeChange} />

      <BackupSettingsCard
        actionLoading={actionLoading}
        exportProgress={exportProgress}
        exportLogs={exportLogs}
        exportLogsRef={exportLogsRef}
        webDAVSettings={webDAVSettings}
        webDAVSaving={webDAVSaving}
        onInitialize={() => setInitializeConfirmOpen(true)}
        onExport={(target, password) => { void handleExportSystem(target, password) }}
        onSaveWebDAV={handleSaveWebDAV}
        onOpenImport={() => { void handleOpenImport() }}
      />

      <BackupImportModal
        open={importModalOpen}
        filePath={importFilePath}
        actionLoading={actionLoading}
        importProgress={importProgress}
        onClose={() => {
          setImportModalOpen(false)
          setImportFilePath('')
          setImportProgress(null)
        }}
        onImport={(resetFirst, password) => { void handleImportSystem(resetFirst, password) }}
      />

      <ConfirmModal
        open={initializeConfirmOpen}
        onClose={() => setInitializeConfirmOpen(false)}
        onConfirm={() => { void handleInitializeSystem() }}
        title="恢复出厂设置"
        content="此操作会清空所有实例、代理、分组及浏览器数据，且无法撤销。建议先导出配置。"
        confirmText="确认恢复"
        danger
      />

    </div>
  )
}
