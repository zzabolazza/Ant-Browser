import { useRef, useState } from 'react'
import { ConfirmModal, toast } from '../../shared/components'
import {
  initializeSystemData,
  exportSystemConfig,
  importSystemConfig,
} from './api'
import { BackupImportModal, BackupSettingsCard } from './components/BackupSettingsCard'
import { ThemeSettingsCard } from './components/ThemeSettingsCard'
import type { BackupExportLogItem, BackupExportProgress } from './progress'
import { useSettingsProgressEffects } from './hooks/useSettingsProgressEffects'
import { getStoredTheme, setThemeMode, type ThemeMode } from '../../shared/theme/theme'

export function SettingsPage() {
  const [theme, setTheme] = useState<ThemeMode>(() => getStoredTheme())
  const [importModalOpen, setImportModalOpen] = useState(false)
  const [initializeConfirmOpen, setInitializeConfirmOpen] = useState(false)
  const [actionLoading, setActionLoading] = useState<'none' | 'init' | 'export' | 'import-reset' | 'import-merge'>('none')
  const [exportProgress, setExportProgress] = useState<BackupExportProgress | null>(null)
  const [importProgress, setImportProgress] = useState<BackupExportProgress | null>(null)
  const [exportLogs, setExportLogs] = useState<BackupExportLogItem[]>([])
  const exportLogsRef = useRef<HTMLDivElement | null>(null)

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
      toast.success(res.message || '已恢复出厂设置')
    } catch (error: any) {
      toast.error(error?.message || '恢复出厂设置失败')
    } finally {
      setActionLoading('none')
    }
  }

  const handleExportSystem = async () => {
    setActionLoading('export')
    setExportLogs([])
    setExportProgress({ phase: 'starting', progress: 0, message: '准备导出...' })
    try {
      const res = await exportSystemConfig()
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

  const handleImportSystem = async (resetFirst: boolean) => {
    setActionLoading(resetFirst ? 'import-reset' : 'import-merge')
    setImportProgress({
      phase: 'starting',
      progress: 0,
      message: resetFirst ? '等待选择 ZIP 配置（清空现有数据后加载）...' : '等待选择 ZIP 配置（判重合并）...',
    })
    try {
      const res = await importSystemConfig(resetFirst)
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
      setImportModalOpen(false)
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
        onInitialize={() => setInitializeConfirmOpen(true)}
        onExport={() => { void handleExportSystem() }}
        onOpenImport={() => {
          setImportProgress(null)
          setImportModalOpen(true)
        }}
      />

      <BackupImportModal
        open={importModalOpen}
        actionLoading={actionLoading}
        importProgress={importProgress}
        onClose={() => {
          setImportModalOpen(false)
          setImportProgress(null)
        }}
        onImport={(resetFirst) => { void handleImportSystem(resetFirst) }}
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
