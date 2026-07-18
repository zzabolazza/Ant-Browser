import type { RefObject } from 'react'
import { Download, RotateCcw, Upload } from 'lucide-react'

import { Button, Card, Modal, Progress } from '../../../shared/components'

import type { BackupExportLogItem, BackupExportProgress } from '../progress'

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
  onInitialize: () => void
  onExport: () => void
  onOpenImport: () => void
}

interface BackupImportModalProps {
  open: boolean
  actionLoading: BackupActionLoading
  importProgress: BackupExportProgress | null
  onClose: () => void
  onImport: (resetFirst: boolean) => void
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
  onInitialize,
  onExport,
  onOpenImport,
}: BackupSettingsCardProps) {
  const actionRunning = actionLoading !== 'none'

  return (
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
              <p className="mt-1 text-xs leading-5 text-[var(--color-text-muted)]">将当前系统数据打包为 ZIP 文件，便于迁移或留存。</p>
            </div>
          </div>
          <Button
            className="w-full sm:w-auto sm:min-w-[112px]"
            variant="primary"
            size="sm"
            onClick={onExport}
            loading={actionLoading === 'export'}
            disabled={actionRunning && actionLoading !== 'export'}
          >
            导出
          </Button>
        </section>

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
              <p className="mt-1 text-xs leading-5 text-[var(--color-text-muted)]">从 ZIP 文件恢复数据，可选择合并导入或清空后恢复。</p>
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
  )
}

export function BackupImportModal({
  open,
  actionLoading,
  importProgress,
  onClose,
  onImport,
}: BackupImportModalProps) {
  const importRunning = actionLoading === 'import-reset' || actionLoading === 'import-merge'

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
                <Button className="min-w-[104px]" size="sm" variant="success" onClick={() => onImport(false)}>
                  <Upload className="h-4 w-4" />
                  合并导入
                </Button>
              </section>

              <section className="flex items-center gap-3 rounded-[10px] border border-[rgb(239_71_87_/_0.2)] bg-[rgb(239_71_87_/_0.035)] p-3.5">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[rgb(239_71_87_/_0.1)] text-[var(--color-error)]">
                  <RotateCcw className="h-4 w-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <h3 className="text-[13.5px] font-semibold text-[var(--color-text-primary)]">清空后恢复</h3>
                  <p className="mt-1 text-xs leading-5 text-[var(--color-text-muted)]">删除当前数据，以备份内容完整恢复系统。</p>
                </div>
                <Button className="min-w-[104px]" size="sm" variant="danger" onClick={() => onImport(true)}>
                  <Upload className="h-4 w-4" />
                  清空恢复
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
