import { useEffect, useRef, useState, type RefObject } from 'react'
import { ChevronDown, Cloud, Download, RotateCcw, Save, Upload } from 'lucide-react'

import { Button, Card, FormItem, Input, Modal, Progress } from '../../../shared/components'

import type { BackupExportLogItem, BackupExportProgress } from '../progress'
import type { BackupWebDAVSettings } from '../api'

type BackupActionLoading = 'none' | 'init' | 'export' | 'import-reset' | 'import-merge'

interface BackupProgressPanelProps {
  progress: BackupExportProgress
  loadingLabel: string
  logs?: BackupExportLogItem[]
  logsRef?: RefObject<HTMLDivElement>
}

interface BackupSettingsCardProps {
  actionLoading: BackupActionLoading
  exportProgress: BackupExportProgress | null
  exportLogs: BackupExportLogItem[]
  exportLogsRef: RefObject<HTMLDivElement>
  webDAVSettings: BackupWebDAVSettings
  webDAVSaving: boolean
  onInitialize: () => void
  onExport: (target: 'local' | 'webdav', password: string) => void
  onSaveWebDAV: (settings: BackupWebDAVSettings) => Promise<void>
  onOpenImport: () => void
}

interface BackupImportModalProps {
  open: boolean
  filePath: string
  actionLoading: BackupActionLoading
  importProgress: BackupExportProgress | null
  onClose: () => void
  onImport: (resetFirst: boolean, password: string) => void
}

function BackupProgressPanel({ progress, loadingLabel, logs = [], logsRef }: BackupProgressPanelProps) {
  return (
    <div className="rounded-md border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)] px-3 py-2 space-y-2">
      <div className="flex items-center justify-between text-xs">
        <span className="text-[var(--color-text-secondary)]">{progress.message}</span>
        {progress.phase === 'error' && <span className="text-[var(--color-error)]">失败</span>}
        {progress.phase === 'done' && <span className="text-[var(--color-success)]">完成</span>}
        {progress.phase !== 'done' && progress.phase !== 'error' && (
          <span className="text-[var(--color-text-muted)]">{loadingLabel}</span>
        )}
      </div>
      {(progress.componentName || progress.componentId || logsRef) && (
        <div className="text-xs text-[var(--color-text-muted)]">
          当前组件：
          {' '}
          {progress.componentName || progress.componentId || '准备中'}
          {progress.entryIndex && progress.entryTotal
            ? `（${progress.entryIndex}/${progress.entryTotal}）`
            : ''}
        </div>
      )}
      <Progress
        percent={progress.progress}
        size="sm"
        status={progress.phase === 'error' ? 'error' : progress.phase === 'done' ? 'success' : 'normal'}
      />
      {logsRef && (
        <div className="rounded border border-[var(--color-border-muted)] bg-[var(--color-bg-primary)] px-2 py-2">
          <div className="flex items-center justify-between text-xs mb-1">
            <span className="text-[var(--color-text-secondary)]">导出日志</span>
            <span className="text-[var(--color-text-muted)]">{logs.length} 条</span>
          </div>
          <div ref={logsRef} className="max-h-36 overflow-y-auto pr-1 space-y-1">
            {logs.length === 0 && (
              <p className="text-xs text-[var(--color-text-muted)]">等待导出日志...</p>
            )}
            {logs.map(item => (
              <div key={item.id} className="text-xs leading-5 font-mono">
                <span className="text-[var(--color-text-muted)] mr-2">{item.time}</span>
                <span className={item.phase === 'error' ? 'text-[var(--color-error)]' : item.phase === 'done' ? 'text-[var(--color-success)]' : 'text-[var(--color-text-secondary)]'}>
                  {item.text}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

export function BackupSettingsCard({
  actionLoading,
  exportProgress,
  exportLogs,
  exportLogsRef,
  webDAVSettings,
  webDAVSaving,
  onInitialize,
  onExport,
  onSaveWebDAV,
  onOpenImport,
}: BackupSettingsCardProps) {
  const actionRunning = actionLoading !== 'none'
  const [exportTarget, setExportTarget] = useState<'local' | 'webdav'>('local')
  const [exportMenuOpen, setExportMenuOpen] = useState(false)
  const [exportPasswordOpen, setExportPasswordOpen] = useState(false)
  const [exportPassword, setExportPassword] = useState('')
  const [exportPasswordConfirm, setExportPasswordConfirm] = useState('')
  const menuRef = useRef<HTMLDivElement | null>(null)
  const [webDAVForm, setWebDAVForm] = useState<BackupWebDAVSettings>(webDAVSettings)

  useEffect(() => setWebDAVForm({ ...webDAVSettings, password: '' }), [webDAVSettings])
  useEffect(() => {
    const close = (event: MouseEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) setExportMenuOpen(false)
    }
    document.addEventListener('mousedown', close)
    return () => document.removeEventListener('mousedown', close)
  }, [])

  const openExportPassword = (target: 'local' | 'webdav') => {
    setExportTarget(target)
    setExportMenuOpen(false)
    setExportPassword('')
    setExportPasswordConfirm('')
    setExportPasswordOpen(true)
  }

  const selectExportTarget = (target: 'local' | 'webdav') => {
    setExportTarget(target)
    setExportMenuOpen(false)
    if (target === 'local') {
      openExportPassword('local')
      return
    }
    if (webDAVSettings.url) {
      openExportPassword('webdav')
    }
  }

  const passwordValid = exportPassword.length >= 8
    && exportPassword === exportPasswordConfirm
    && (exportTarget === 'local' || Boolean(webDAVSettings.url))

  return (
    <>
    <Card
      title="系统备份"
      subtitle="备份包含应用配置、实例、代理及相关本机数据"
      padding="md"
    >
      <div className="space-y-4">
        <section className="flex flex-col gap-3 rounded-[10px] border border-[var(--color-border-default)] bg-[var(--color-bg-subtle)] p-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-start gap-3">
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[var(--color-accent-muted)] text-[var(--color-accent)]">
              <Download className="h-4 w-4" />
            </div>
            <div className="min-w-0">
              <h3 className="text-[13.5px] font-semibold text-[var(--color-text-primary)]">导出完整备份</h3>
              <p className="mt-1 text-xs leading-5 text-[var(--color-text-muted)]">将当前系统数据打包为强制密码保护的加密备份。</p>
            </div>
          </div>
          <div ref={menuRef} className="relative w-full sm:w-auto">
            <Button
              className="w-full sm:w-auto sm:min-w-[112px]"
              variant="primary"
              size="sm"
              onClick={() => setExportMenuOpen(value => !value)}
              loading={actionLoading === 'export'}
              disabled={actionRunning && actionLoading !== 'export'}
              aria-haspopup="menu"
              aria-expanded={exportMenuOpen}
            >
              导出
              <ChevronDown className="h-4 w-4" />
            </Button>
            {exportMenuOpen && (
              <div className="absolute right-0 top-9 z-20 min-w-[168px] overflow-hidden rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] py-1 shadow-lg" role="menu">
                <button type="button" role="menuitem" className="flex w-full items-center gap-2 px-3 py-2 text-left text-xs text-[var(--color-text-primary)] hover:bg-[var(--color-bg-muted)]" onClick={() => selectExportTarget('local')}>
                  <Download className="h-4 w-4" />保存到本地
                </button>
                <button type="button" role="menuitem" className="flex w-full items-center gap-2 px-3 py-2 text-left text-xs text-[var(--color-text-primary)] hover:bg-[var(--color-bg-muted)]" onClick={() => selectExportTarget('webdav')}>
                  <Cloud className="h-4 w-4" />上传到 WebDAV
                </button>
              </div>
            )}
          </div>
        </section>

        {exportTarget === 'webdav' && (
          <section className="rounded-[10px] border border-[var(--color-border-default)] bg-[var(--color-bg-subtle)] p-4">
            <div className="mb-3 flex items-center gap-2">
              <Cloud className="h-4 w-4 text-[var(--color-accent)]" />
              <div>
                <h3 className="text-[13.5px] font-semibold text-[var(--color-text-primary)]">WebDAV 设置</h3>
                <p className="mt-0.5 text-xs text-[var(--color-text-muted)]">连接密码保存在本机配置中，读取设置时不会回传到界面。</p>
              </div>
            </div>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <FormItem label="服务器地址" required className="sm:col-span-2">
                <Input value={webDAVForm.url} placeholder="https://dav.example.com/remote.php/dav/files/user" onChange={event => setWebDAVForm(value => ({ ...value, url: event.target.value }))} />
              </FormItem>
              <FormItem label="用户名">
                <Input value={webDAVForm.username} autoComplete="username" onChange={event => setWebDAVForm(value => ({ ...value, username: event.target.value }))} />
              </FormItem>
              <FormItem label="连接密码" hint={webDAVForm.hasPassword ? '留空则保留已保存密码' : undefined}>
                <Input type="password" value={webDAVForm.password || ''} autoComplete="new-password" onChange={event => setWebDAVForm(value => ({ ...value, password: event.target.value }))} />
              </FormItem>
              <FormItem label="远程目录" hint="可选" className="sm:col-span-2">
                <Input value={webDAVForm.remoteDir} placeholder="Facade/Backups" onChange={event => setWebDAVForm(value => ({ ...value, remoteDir: event.target.value }))} />
              </FormItem>
            </div>
            <div className="mt-3 flex justify-end">
              <Button size="sm" variant="secondary" loading={webDAVSaving} disabled={!webDAVForm.url.trim()} onClick={() => { void onSaveWebDAV(webDAVForm) }}>
                <Save className="h-4 w-4" />保存 WebDAV 设置
              </Button>
            </div>
          </section>
        )}

        {exportProgress && (
          <BackupProgressPanel
            progress={exportProgress}
            loadingLabel="处理中"
            logs={exportLogs}
            logsRef={exportLogsRef}
          />
        )}

        <section className="flex flex-col gap-3 rounded-[10px] border border-[var(--color-border-default)] bg-[var(--color-bg-subtle)] p-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-start gap-3">
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[rgb(22_199_132_/_0.1)] text-[var(--color-success)]">
              <Upload className="h-4 w-4" />
            </div>
            <div className="min-w-0">
              <h3 className="text-[13.5px] font-semibold text-[var(--color-text-primary)]">导入系统备份</h3>
              <p className="mt-1 text-xs leading-5 text-[var(--color-text-muted)]">从加密备份恢复数据，可选择合并导入或完整恢复。</p>
            </div>
          </div>
          <Button
            className="w-full sm:w-auto sm:min-w-[112px]"
            variant="success"
            size="sm"
            onClick={onOpenImport}
            disabled={actionRunning}
          >
            导入
          </Button>
        </section>

        <section className="flex flex-col gap-3 rounded-[10px] border border-[rgb(239_71_87_/_0.18)] bg-[rgb(239_71_87_/_0.025)] p-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-start gap-3">
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[rgb(239_71_87_/_0.1)] text-[var(--color-error)]">
              <RotateCcw className="h-4 w-4" />
            </div>
            <div>
              <h3 className="text-[13.5px] font-semibold text-[var(--color-text-primary)]">恢复出厂设置</h3>
              <p className="mt-1 text-xs leading-5 text-[var(--color-text-muted)]">清空所有业务数据并恢复默认配置，此操作无法撤销。</p>
            </div>
          </div>
          <Button
            className="w-full sm:w-auto sm:min-w-[112px] sm:self-center"
            variant="danger"
            size="sm"
            onClick={onInitialize}
            loading={actionLoading === 'init'}
            disabled={actionRunning && actionLoading !== 'init'}
          >
            恢复
          </Button>
        </section>
      </div>
    </Card>
    <Modal
      open={exportPasswordOpen}
      onClose={() => { if (!actionRunning) setExportPasswordOpen(false) }}
      title={exportTarget === 'webdav' ? '加密并上传到 WebDAV' : '导出本地加密备份'}
      width="440px"
      closable={!actionRunning}
      footer={(
        <>
          <Button variant="secondary" disabled={actionRunning} onClick={() => setExportPasswordOpen(false)}>取消</Button>
          <Button disabled={!passwordValid || actionRunning} loading={actionLoading === 'export'} onClick={() => {
            const password = exportPassword
            setExportPasswordOpen(false)
            setExportPassword('')
            setExportPasswordConfirm('')
            onExport(exportTarget, password)
          }}>开始导出</Button>
        </>
      )}
    >
      <div className="space-y-3">
        <p className="text-xs leading-5 text-[var(--color-text-muted)]">备份将使用密码加密。密码不会保存，丢失后无法恢复备份。</p>
        {exportTarget === 'webdav' && !webDAVSettings.url && (
          <p className="text-xs text-[var(--color-error)]">请先保存 WebDAV 设置。</p>
        )}
        <FormItem label="备份密码" required hint="至少 8 个字符">
          <Input type="password" autoFocus value={exportPassword} autoComplete="new-password" onChange={event => setExportPassword(event.target.value)} />
        </FormItem>
        <FormItem label="确认密码" required error={exportPasswordConfirm.length > 0 && exportPassword !== exportPasswordConfirm ? '两次输入的密码不一致' : undefined}>
          <Input type="password" value={exportPasswordConfirm} autoComplete="new-password" onChange={event => setExportPasswordConfirm(event.target.value)} />
        </FormItem>
      </div>
    </Modal>
    </>
  )
}

export function BackupImportModal({
  open,
  filePath,
  actionLoading,
  importProgress,
  onClose,
  onImport,
}: BackupImportModalProps) {
  const importRunning = actionLoading === 'import-reset' || actionLoading === 'import-merge'
  const [password, setPassword] = useState('')
  const fileName = filePath.split(/[/\\]/).pop() || filePath

  useEffect(() => {
    if (!open) setPassword('')
  }, [open])

  return (
    <Modal
      open={open}
      onClose={() => {
        if (actionLoading !== 'none') {
          return
        }
        onClose()
      }}
      title="加载配置"
      width="520px"
      closable={!importRunning}
      footer={!importRunning ? (
        <Button variant="secondary" onClick={onClose}>取消</Button>
      ) : undefined}
    >
      <div className="space-y-4 text-sm text-[var(--color-text-secondary)]">
        {!importRunning && (
          <>
            <FormItem label="备份文件">
              <Input value={fileName} readOnly title={filePath} />
            </FormItem>
            <FormItem label="备份密码" required hint="仅支持加密备份">
              <Input type="password" autoFocus value={password} autoComplete="current-password" onChange={event => setPassword(event.target.value)} />
            </FormItem>
            <p className="text-[13px] leading-5 text-[var(--color-text-muted)]">选择备份文件的导入方式：</p>
            <div className="space-y-3">
              <section className="flex items-center gap-3 rounded-[10px] border border-[var(--color-border-default)] p-3.5">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[var(--color-accent-muted)] text-[var(--color-accent)]">
                  <Upload className="h-4 w-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <h3 className="text-[13.5px] font-semibold text-[var(--color-text-primary)]">合并导入</h3>
                    <span className="rounded bg-[var(--color-accent-muted)] px-1.5 py-0.5 text-[10px] font-medium text-[var(--color-accent)]">推荐</span>
                  </div>
                  <p className="mt-1 text-xs leading-5 text-[var(--color-text-muted)]">保留现有数据，对备份内容判重后合并。</p>
                </div>
                <Button className="min-w-[104px]" size="sm" variant="success" disabled={password.length < 8} onClick={() => onImport(false, password)}>
                  <Upload className="h-4 w-4" />
                  合并导入
                </Button>
              </section>

              <section className="flex items-center gap-3 rounded-[10px] border border-[rgb(239_71_87_/_0.2)] bg-[rgb(239_71_87_/_0.035)] p-3.5">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[rgb(239_71_87_/_0.1)] text-[var(--color-error)]">
                  <RotateCcw className="h-4 w-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <h3 className="text-[13.5px] font-semibold text-[var(--color-text-primary)]">完整恢复</h3>
                  <p className="mt-1 text-xs leading-5 text-[var(--color-text-muted)]">以备份内容完整替换当前系统数据，已有内容不会保留。</p>
                </div>
                <Button className="min-w-[104px]" size="sm" variant="danger" disabled={password.length < 8} onClick={() => onImport(true, password)}>
                  <Upload className="h-4 w-4" />
                  完整恢复
                </Button>
              </section>
            </div>
          </>
        )}
        {importProgress && (
          <BackupProgressPanel progress={importProgress} loadingLabel="加载中" />
        )}
        {importRunning && (
          <p className="text-xs text-[var(--color-warning)]">
            当前正在加载配置，弹窗不可关闭。若需中断，请直接关闭应用。
          </p>
        )}
      </div>
    </Modal>
  )
}
