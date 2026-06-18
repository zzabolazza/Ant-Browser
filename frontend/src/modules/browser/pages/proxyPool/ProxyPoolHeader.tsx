import { Button } from '../../../../shared/components'

interface ProxyPoolHeaderProps {
  checkingAllIPHealth: boolean
  connectorSwitching: boolean
  currentConnectorStatus: string
  currentConnectorType: string
  hasURLImportSources: boolean
  onCheckAllIPHealth: () => void
  onOpenImport: () => void
  onOpenCoreDownload: () => void
  onOpenSettings: () => void
  onSwitchConnector: () => void
  onRefreshAllSources: () => void
  onTestAll: () => void
  refreshingAllSources: boolean
  testingAll: boolean
  totalCount: number
}

export function ProxyPoolHeader({
  checkingAllIPHealth,
  connectorSwitching,
  currentConnectorStatus,
  currentConnectorType,
  hasURLImportSources,
  onCheckAllIPHealth,
  onOpenImport,
  onOpenCoreDownload,
  onOpenSettings,
  onSwitchConnector,
  onRefreshAllSources,
  onTestAll,
  refreshingAllSources,
  testingAll,
  totalCount,
}: ProxyPoolHeaderProps) {
  const normalizedConnector = currentConnectorType === 'mihomo' ? 'mihomo' : 'xray'
  const connectorLabel = normalizedConnector === 'mihomo' ? 'Mihomo' : 'Xray'
  const nextConnectorLabel = normalizedConnector === 'mihomo' ? 'Xray' : 'Mihomo'

  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-xl font-semibold text-[var(--color-text-primary)]">代理池配置</h1>
        <p className="text-sm text-[var(--color-text-muted)] mt-1">管理代理配置，支持 Clash 订阅、HTTP、HTTPS、SOCKS5</p>
      </div>
      <div className="flex items-center gap-2">
        <div className="flex items-center gap-2 rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-primary)] px-2 py-1 shadow-sm">
          <span className="inline-flex items-center gap-1 whitespace-nowrap px-2 text-xs text-[var(--color-text-muted)]">
            当前内核：{connectorLabel}
            {normalizedConnector === 'mihomo' && (
              <span className="rounded-full border border-amber-200 bg-amber-50 px-1.5 py-0.5 text-[10px] font-medium leading-none text-amber-700">
                实验
              </span>
            )}
            <span>· {currentConnectorStatus || '未知'}</span>
          </span>
          <Button size="sm" variant="secondary" onClick={onSwitchConnector} loading={connectorSwitching}>
            切换到{nextConnectorLabel}
          </Button>
          <Button size="sm" variant="secondary" onClick={onOpenCoreDownload}>下载内核</Button>
        </div>
        <div className="flex items-center gap-2 rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-primary)] px-2 py-1 shadow-sm">
          <Button
            size="sm"
            variant="secondary"
            onClick={onOpenSettings}
          >
            检测设置
          </Button>
          <Button
            size="sm"
            variant="secondary"
            onClick={onRefreshAllSources}
            loading={refreshingAllSources}
            disabled={!hasURLImportSources}
          >
            刷新订阅
          </Button>
          <Button
            size="sm"
            variant="secondary"
            onClick={onCheckAllIPHealth}
            loading={checkingAllIPHealth}
            disabled={totalCount === 0}
          >
            检测IP健康
          </Button>
          <Button
            size="sm"
            variant="secondary"
            onClick={onTestAll}
            loading={testingAll}
            disabled={totalCount === 0}
          >
            测试全部
          </Button>
        </div>
        <Button size="sm" onClick={onOpenImport}>导入代理</Button>
      </div>
    </div>
  )
}
