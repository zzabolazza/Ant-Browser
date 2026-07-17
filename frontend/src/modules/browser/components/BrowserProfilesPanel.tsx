import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { Link } from 'react-router-dom'
import { Copy, Download, Key, Loader2, MoreHorizontal, Play, Puzzle, Repeat2, RotateCcw, Settings, Square, Trash2, Wifi } from 'lucide-react'

import { Badge, Button, Table } from '../../../shared/components'
import type { TableColumn } from '../../../shared/components/Table'

import type { BrowserCore, BrowserProfile, BrowserProxy, ProxyIPHealthResult, ProxySpeedTestResult } from '../types'
import { browserProxyTestSpeed, testProxyConnectivity } from '../api'
import { readIPHealthCache } from '../pages/proxyPool/storage'
import { resolveProfileProxyCountryDisplay } from '../utils/countryFlag'
import type { BrowserViewMode } from './BrowserListLayout'
import { CdpUrlCell, KeywordInlineRow, LaunchCodeCell } from './BrowserListWidgets'

type ProfileStatusVariant = 'default' | 'success' | 'error' | 'warning' | 'info'

interface ProfileStatus {
  variant: ProfileStatusVariant
  label: string
}

interface BrowserProfilesPanelProps {
  loading: boolean
  viewMode: BrowserViewMode
  profiles: BrowserProfile[]
  proxies: BrowserProxy[]
  selectedIds: Set<string>
  resolveProfileCore: (profile: BrowserProfile) => BrowserCore | null
  getProfileCoreLabel: (profile: BrowserProfile) => string
  getProfileStatus: (profile: BrowserProfile) => ProfileStatus
  isProfileStarting: (profileId: string) => boolean
  isProfileStopping: (profileId: string) => boolean
  isProfileBusy: (profileId: string) => boolean
  onToggleSelect: (profileId: string) => void
  onSelectAll: () => void
  onDeselectAll: () => void
  onRefreshProfiles: () => void
  onStart: (profileId: string) => void
  onStop: (profileId: string) => void
  onRestart: (profileId: string) => void
  onOpenKeywords: (profile: BrowserProfile) => void
  onOpenExtensions: (profile: BrowserProfile) => void
  onExport: (profile: BrowserProfile) => void
  onOpenCopy: (profile: BrowserProfile) => void
  onOpenProxyPicker: (profile: BrowserProfile) => void
  onDelete: (profileId: string) => void
}

const formatTime = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString('zh-CN')
}

function formatProxyLabel(profile: BrowserProfile, proxy?: BrowserProxy): string {
  if (proxy?.proxyName) {
    return proxy.proxyName
  }
  if (profile.proxyId) {
    return profile.proxyId
  }
  const customProxy = (profile.proxyConfig || '').trim()
  if (customProxy) {
    return `自定义: ${customProxy}`
  }
  return '-'
}

function ProxyLatency({ result }: { result?: ProxySpeedTestResult | null }) {
  if (!result) return null
  if (!result.ok) return <span className="text-xs text-red-500">失败</span>
  const color = result.latencyMs < 200 ? 'text-green-500' : result.latencyMs < 500 ? 'text-yellow-500' : 'text-red-500'
  return <span className={`text-xs font-medium ${color}`}>{result.latencyMs}ms</span>
}

function ProxyInlineActions({
  profile,
  proxy,
  countryDisplay,
  isBusy,
  onOpenProxyPicker,
  maxWidthClass = 'max-w-[220px]',
}: {
  profile: BrowserProfile
  proxy?: BrowserProxy
  countryDisplay?: { code: string; flag: string } | null
  isBusy: boolean
  onOpenProxyPicker: (profile: BrowserProfile) => void
  maxWidthClass?: string
}) {
  const [testing, setTesting] = useState(false)
  const [speedResult, setSpeedResult] = useState<ProxySpeedTestResult | null>(null)
  const historyResult = proxy?.lastTestedAt
    ? {
        proxyId: proxy.proxyId,
        ok: proxy.lastTestOk ?? false,
        latencyMs: proxy.lastLatencyMs ?? -1,
        error: '',
      }
    : null
  const displayResult = speedResult || historyResult
  const canTest = !!profile.proxyId || !!profile.proxyConfig.trim()
  const proxyLabel = formatProxyLabel(profile, proxy)
  const title = countryDisplay ? `${countryDisplay.flag} ${countryDisplay.code} · ${proxyLabel}` : proxyLabel

  const handleTest = async () => {
    if (testing || !canTest) return
    setTesting(true)
    try {
      const result = profile.proxyId
        ? await browserProxyTestSpeed(profile.proxyId)
        : await testProxyConnectivity(profile.profileId, profile.proxyConfig)
      setSpeedResult(result)
    } catch (error: any) {
      setSpeedResult({
        proxyId: profile.proxyId || profile.profileId,
        ok: false,
        latencyMs: -1,
        error: error?.message || '测速失败',
      })
    } finally {
      setTesting(false)
    }
  }

  return (
    <div className={`inline-flex ${maxWidthClass} items-center gap-1.5 text-xs`} title={title}>
      {countryDisplay && (
        <span className="shrink-0 whitespace-nowrap text-[var(--color-text-primary)]" title={countryDisplay.code}>
          {countryDisplay.flag} {countryDisplay.code}
        </span>
      )}
      {countryDisplay && <span className="shrink-0 text-[var(--color-text-muted)]">-</span>}
      <span className="min-w-0 truncate text-[var(--color-text-primary)]">{proxyLabel}</span>
      <button
        type="button"
        className="shrink-0 rounded p-0.5 text-[var(--color-text-muted)] transition-colors hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-40"
        title={isBusy ? '实例操作中，暂不可切换代理' : '切换代理'}
        disabled={isBusy}
        onClick={() => onOpenProxyPicker(profile)}
      >
        <Repeat2 className="h-3.5 w-3.5" />
      </button>
      <button
        type="button"
        className="shrink-0 rounded p-0.5 text-[var(--color-text-muted)] transition-colors hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-40"
        title={canTest ? '测速' : '无可测速代理'}
        disabled={testing || !canTest}
        onClick={handleTest}
      >
        {testing ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Wifi className="h-3.5 w-3.5" />}
      </button>
      <ProxyLatency result={displayResult} />
    </div>
  )
}

function ProfileMoreActions({
  open,
  disabled,
  onToggle,
  onClose,
  onRestart,
  onOpenKeywords,
  onOpenExtensions,
  onExport,
}: {
  open: boolean
  disabled: boolean
  onToggle: () => void
  onClose: () => void
  onRestart: () => void
  onOpenKeywords: () => void
  onOpenExtensions: () => void
  onExport: () => void
}) {
  const triggerRef = useRef<HTMLDivElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)
  const [menuPosition, setMenuPosition] = useState({ top: 0, left: 0 })

  useEffect(() => {
    if (!open) return
    const updateMenuPosition = () => {
      const rect = triggerRef.current?.getBoundingClientRect()
      if (!rect) return
      const menuWidth = 128
      const menuHeight = 168
      const gap = 8
      const left = Math.max(8, Math.min(rect.right - menuWidth, window.innerWidth - menuWidth - 8))
      const belowTop = rect.bottom + gap
      const top = belowTop + menuHeight > window.innerHeight
        ? Math.max(8, rect.top - menuHeight - gap)
        : belowTop
      setMenuPosition({ top, left })
    }
    const handlePointerDown = (event: MouseEvent) => {
      const target = event.target as Node
      if (!triggerRef.current?.contains(target) && !menuRef.current?.contains(target)) {
        onClose()
      }
    }
    updateMenuPosition()
    document.addEventListener('mousedown', handlePointerDown)
    window.addEventListener('resize', updateMenuPosition)
    window.addEventListener('scroll', updateMenuPosition, true)
    return () => {
      document.removeEventListener('mousedown', handlePointerDown)
      window.removeEventListener('resize', updateMenuPosition)
      window.removeEventListener('scroll', updateMenuPosition, true)
    }
  }, [open, onClose])

  const runAndClose = (handler: () => void) => {
    handler()
    onClose()
  }

  return (
    <>
    <div ref={triggerRef} className="inline-flex">
      <Button
        size="icon"
        variant="ghost"
        onClick={onToggle}
        title="更多"
        disabled={disabled}
      >
        <MoreHorizontal className="h-3.5 w-3.5 shrink-0" />
      </Button>
    </div>
      {open && createPortal(
        <div
          ref={menuRef}
          className="fixed z-[9999] w-32 rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-1.5 shadow-xl"
          style={{ top: menuPosition.top, left: menuPosition.left }}
        >
          <button
            type="button"
            className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-xs text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-text-primary)]"
            onClick={() => runAndClose(onRestart)}
          >
            <RotateCcw className="w-3.5 h-3.5" />
            重启
          </button>
          <button
            type="button"
            className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-xs text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-text-primary)]"
            onClick={() => runAndClose(onOpenKeywords)}
          >
            <Key className="w-3.5 h-3.5" />
            关键字
          </button>
          <button
            type="button"
            className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-xs text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-text-primary)]"
            onClick={() => runAndClose(onOpenExtensions)}
          >
            <Puzzle className="w-3.5 h-3.5" />
            插件
          </button>
          <button
            type="button"
            className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-xs text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-text-primary)]"
            onClick={() => runAndClose(onExport)}
          >
            <Download className="w-3.5 h-3.5" />
            导出
          </button>
        </div>,
        document.body
      )}
    </>
  )
}

function EditProfileAction({
  profileId,
  disabled,
  compact = false,
}: {
  profileId: string
  disabled: boolean
  compact?: boolean
}) {
  const iconClassName = 'h-3.5 w-3.5 shrink-0'
  const linkClassName = compact
    ? 'inline-flex h-7 w-7 items-center justify-center rounded-lg px-0 text-xs font-semibold text-[var(--color-text-secondary)] transition-all duration-150 hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-text-primary)] focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-[var(--color-border-strong)]'
    : 'inline-flex h-8 items-center justify-center gap-1.5 rounded-lg px-3 text-xs font-semibold text-[var(--color-text-secondary)] transition-all duration-150 hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-text-primary)] focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-[var(--color-border-strong)]'

  if (disabled) {
    return (
      <Button
        size={compact ? 'icon' : 'sm'}
        variant="ghost"
        title="配置"
        disabled
      >
        <Settings className={iconClassName} />
        {!compact && '配置'}
      </Button>
    )
  }

  return (
    <Link
      to={`/browser/edit/${profileId}`}
      title="配置"
      className={linkClassName}
    >
      <Settings className={iconClassName} />
      {!compact && '配置'}
    </Link>
  )
}

function BrowserProfileCard({
  profile,
  proxy,
  countryDisplay,
  isSelected,
  status,
  coreLabel,
  isStarting,
  isStopping,
  isBusy,
  onToggleSelect,
  onRefreshProfiles,
  onStart,
  onStop,
  onRestart,
  onOpenKeywords,
  onOpenExtensions,
  onOpenCopy,
  onOpenProxyPicker,
  onDelete,
}: {
  profile: BrowserProfile
  proxy: BrowserProxy | undefined
  countryDisplay: { code: string; flag: string } | null
  isSelected: boolean
  status: ProfileStatus
  coreLabel: string
  isStarting: boolean
  isStopping: boolean
  isBusy: boolean
  onToggleSelect: (profileId: string) => void
  onRefreshProfiles: () => void
  onStart: (profileId: string) => void
  onStop: (profileId: string) => void
  onRestart: (profileId: string) => void
  onOpenKeywords: (profile: BrowserProfile) => void
  onOpenExtensions: (profile: BrowserProfile) => void
  onOpenCopy: (profile: BrowserProfile) => void
  onOpenProxyPicker: (profile: BrowserProfile) => void
  onDelete: (profileId: string) => void
}) {
  return (
    <div
      className={`flex h-[300px] flex-col overflow-hidden rounded-[10px] border bg-[var(--color-bg-surface)] p-3 transition-colors duration-150
        ${isSelected ? 'border-[var(--color-accent)] ring-1 ring-[rgb(75_110_255_/_0.2)]' : 'border-[var(--color-border-default)] hover:border-[var(--color-border-strong)]'}
      `}
    >
      <div className="flex flex-col gap-3 pb-3 border-b border-[var(--color-border-muted)] shrink-0">
        <div className="flex justify-between items-start gap-2">
          <div className="flex items-center gap-2 flex-wrap">
            <input
              type="checkbox"
              className="w-4 h-4 rounded cursor-pointer accent-[var(--color-accent)] mt-0.5 shrink-0"
              checked={isSelected}
              onChange={() => onToggleSelect(profile.profileId)}
            />
            <Link className="max-w-[200px] truncate text-sm font-semibold text-[var(--color-text-primary)] transition-colors hover:text-[var(--color-accent)] hover:underline" to={`/browser/detail/${profile.profileId}`}>
              {profile.profileName}
            </Link>
            {profile.tags && profile.tags.length > 0 && (
              <div className="flex gap-1 ml-1">
                {profile.tags.map(tag => <Badge variant="default" key={tag}>{tag}</Badge>)}
              </div>
            )}
          </div>

          <Badge variant={status.variant} dot dotClassName="w-2 h-2 shrink-0">
            {status.label}
          </Badge>
        </div>

        <div className="flex items-center gap-1 flex-wrap">
          {profile.running ? (
            <Button size="sm" variant="secondary" onClick={() => onStop(profile.profileId)} title={isStopping ? '停止中' : '停止'} loading={isStopping}>
              {!isStopping && <Square className="h-3.5 w-3.5 shrink-0" />}
              {isStopping ? '停止中' : '停止'}
            </Button>
          ) : (
            <Button
              size="sm"
              variant="ghost"
              onClick={() => onStart(profile.profileId)}
              title={isStarting ? '启动中' : '启动'}
              loading={isStarting}
              className="group hover:!bg-transparent hover:!text-[var(--color-text-secondary)]"
            >
              {!isStarting && <Play className="h-3.5 w-3.5 shrink-0 transition-colors group-hover:text-[var(--color-success)]" />}
              {isStarting ? '启动中' : '启动'}
            </Button>
          )}
          <span className="w-px h-4 bg-[var(--color-border-muted)] mx-1"></span>
          <Button size="sm" variant="ghost" onClick={() => onRestart(profile.profileId)} title="重启" disabled={isBusy}><RotateCcw className="h-3.5 w-3.5 shrink-0" />重启</Button>
          <Button size="sm" variant="ghost" onClick={() => onOpenKeywords(profile)} title="关键字管理" disabled={isBusy}><Key className="h-3.5 w-3.5 shrink-0" />关键字</Button>
          <Button size="sm" variant="ghost" onClick={() => onOpenExtensions(profile)} title="插件配置" disabled={isBusy}><Puzzle className="h-3.5 w-3.5 shrink-0" />插件</Button>
          <EditProfileAction profileId={profile.profileId} disabled={isBusy} />
          <Button size="sm" variant="ghost" onClick={() => onOpenCopy(profile)} title="克隆" disabled={isBusy}><Copy className="h-3.5 w-3.5 shrink-0" />克隆</Button>
          <Button size="sm" variant="ghost" onClick={() => onDelete(profile.profileId)} title="删除" className="text-red-500 hover:text-red-600 hover:bg-red-50" disabled={isBusy}><Trash2 className="h-3.5 w-3.5 shrink-0" />删除</Button>
        </div>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 py-2 shrink-0">
        <div className="flex flex-col gap-0.5">
          <span className="text-xs text-[var(--color-text-muted)] font-medium">内核版本</span>
          <span className="text-xs text-[var(--color-text-primary)]">{coreLabel}</span>
        </div>
        <div className="flex flex-col gap-0.5">
          <span className="text-xs text-[var(--color-text-muted)] font-medium">代理配置</span>
          <ProxyInlineActions
            profile={profile}
            proxy={proxy}
            countryDisplay={countryDisplay}
            isBusy={isBusy}
            onOpenProxyPicker={onOpenProxyPicker}
            maxWidthClass="max-w-full"
          />
        </div>
        <div className="flex flex-col gap-0.5">
          <span className="text-xs text-[var(--color-text-muted)] font-medium">快捷配置码</span>
          <div className="mt-0.5"><LaunchCodeCell profileId={profile.profileId} code={profile.launchCode || ''} onRefresh={onRefreshProfiles} /></div>
        </div>
        <div className="flex flex-col gap-0.5">
          <span className="text-xs text-[var(--color-text-muted)] font-medium">CDP</span>
          <div className="mt-0.5"><CdpUrlCell debugReady={profile.debugReady} debugPort={profile.debugPort} /></div>
        </div>
        <div className="flex flex-col gap-0.5">
          <span className="text-xs text-[var(--color-text-muted)] font-medium">上次更新时间</span>
          <span className="text-xs text-[var(--color-text-primary)]">{formatTime(profile.updatedAt)}</span>
        </div>
      </div>

      <div className="border-t border-[var(--color-border-muted)] pt-2 flex items-start gap-2 flex-1 min-h-0">
        <span className="text-xs font-medium text-[var(--color-text-primary)] shrink-0 pt-0.5">系统关键字</span>
        <div className="flex-1 min-h-0 overflow-y-auto pr-1">
          <KeywordInlineRow keywords={profile.keywords || []} />
        </div>
      </div>
    </div>
  )
}

export function BrowserProfilesPanel({
  loading,
  viewMode,
  profiles,
  proxies,
  selectedIds,
  resolveProfileCore,
  getProfileCoreLabel,
  getProfileStatus,
  isProfileStarting,
  isProfileStopping,
  isProfileBusy,
  onToggleSelect,
  onSelectAll,
  onDeselectAll,
  onRefreshProfiles,
  onStart,
  onStop,
  onRestart,
  onOpenKeywords,
  onOpenExtensions,
  onExport,
  onOpenCopy,
  onOpenProxyPicker,
  onDelete,
}: BrowserProfilesPanelProps) {
  const allSelected = profiles.length > 0 && selectedIds.size === profiles.length
  const partiallySelected = selectedIds.size > 0 && selectedIds.size < profiles.length
  const [openMoreProfileId, setOpenMoreProfileId] = useState<string | null>(null)
  const [ipHealthMap, setIPHealthMap] = useState<Record<string, ProxyIPHealthResult>>({})

  useEffect(() => {
    setIPHealthMap(readIPHealthCache())
  }, [profiles, proxies])

  const getProfileCountryDisplay = (profile: BrowserProfile, proxy?: BrowserProxy) => (
    resolveProfileProxyCountryDisplay(profile.proxyId || '', proxy?.lastIPHealthJson, ipHealthMap)
  )

  const columns: TableColumn<BrowserProfile>[] = [
    {
      key: 'selection',
      title: (
        <input
          type="checkbox"
          className="w-4 h-4 rounded cursor-pointer accent-[var(--color-accent)]"
          checked={allSelected}
          ref={(input) => {
            if (input) {
              input.indeterminate = partiallySelected
            }
          }}
          onChange={(event) => {
            if (event.target.checked) {
              onSelectAll()
            } else {
              onDeselectAll()
            }
          }}
        />
      ),
      width: 40,
      render: (_, record) => (
        <input
          type="checkbox"
          className="w-4 h-4 rounded cursor-pointer accent-[var(--color-accent)]"
          checked={selectedIds.has(record.profileId)}
          onChange={() => onToggleSelect(record.profileId)}
        />
      ),
    },
    {
      key: 'profileName',
      title: '实例名称',
      width: 180,
      render: (value, record) => (
        <div className="flex min-w-0 flex-col gap-1">
          <Link className="block truncate whitespace-nowrap text-sm font-semibold text-[var(--color-text-primary)] transition-colors hover:text-[var(--color-accent)] hover:underline" to={`/browser/detail/${record.profileId}`} title={String(value || '')}>
            {value}
          </Link>
          {record.tags && record.tags.length > 0 && (
            <div className="flex gap-1 flex-wrap">
              {record.tags.map(tag => <Badge variant="default" key={tag}>{tag}</Badge>)}
            </div>
          )}
        </div>
      ),
    },
    {
      key: 'running',
      title: '状态',
      width: 100,
      render: (_, record) => {
        const status = getProfileStatus(record)
        return <Badge variant={status.variant} dot>{status.label}</Badge>
      },
    },
    {
      key: 'coreId',
      title: '核心',
      width: 160,
      render: (_, record) => <span className="text-xs">{getProfileCoreLabel(record)}</span>,
    },
    {
      key: 'proxyId',
      title: '代理',
      width: 260,
      render: (value, record) => {
        const proxy = proxies.find(item => item.proxyId === value)
        const isBusy = isProfileBusy(record.profileId)
        return (
          <ProxyInlineActions
            profile={record}
            proxy={proxy}
            countryDisplay={getProfileCountryDisplay(record, proxy)}
            isBusy={isBusy}
            onOpenProxyPicker={onOpenProxyPicker}
          />
        )
      },
    },
    {
      key: 'launchCode',
      title: '快捷打开码',
      render: (value, record) => <LaunchCodeCell profileId={record.profileId} code={value || ''} onRefresh={onRefreshProfiles} />,
    },
    {
      key: 'cdpUrl',
      title: 'CDP',
      width: 240,
      render: (_, record) => <CdpUrlCell debugReady={record.debugReady} debugPort={record.debugPort} />,
    },
    {
      key: 'keywords',
      title: '关键字',
      width: 200,
      render: (value) => <KeywordInlineRow keywords={value || []} />,
    },
    {
      key: 'updatedAt',
      title: '上次更新',
      render: formatTime,
    },
    {
      key: 'actions',
      title: '操作',
      width: 248,
      align: 'right',
      render: (_, record) => {
        const isStarting = isProfileStarting(record.profileId)
        const isStopping = isProfileStopping(record.profileId)
        const isBusy = isProfileBusy(record.profileId)
        const isMoreOpen = openMoreProfileId === record.profileId

        return (
          <div className="flex justify-end gap-1.5 whitespace-nowrap">
            {record.running ? (
              <Button size="icon" variant="secondary" onClick={() => onStop(record.profileId)} title="停止" loading={isStopping}>
                {!isStopping && <Square className="h-3.5 w-3.5 shrink-0" />}
              </Button>
            ) : (
              <Button
                size="icon"
                variant="ghost"
                onClick={() => onStart(record.profileId)}
                title="启动"
                loading={isStarting}
                className="group hover:!bg-transparent hover:!text-[var(--color-text-secondary)]"
              >
                {!isStarting && <Play className="h-3.5 w-3.5 shrink-0 transition-colors group-hover:text-[var(--color-success)]" />}
              </Button>
            )}
            <EditProfileAction profileId={record.profileId} disabled={isBusy} compact />
            <Button size="icon" variant="ghost" onClick={() => onOpenCopy(record)} title="克隆" disabled={isBusy}><Copy className="h-3.5 w-3.5 shrink-0" /></Button>
            <ProfileMoreActions
              open={isMoreOpen}
              disabled={isBusy}
              onToggle={() => setOpenMoreProfileId(isMoreOpen ? null : record.profileId)}
              onClose={() => setOpenMoreProfileId(null)}
              onRestart={() => onRestart(record.profileId)}
              onOpenKeywords={() => onOpenKeywords(record)}
              onOpenExtensions={() => onOpenExtensions(record)}
              onExport={() => onExport(record)}
            />
            <Button size="icon" variant="ghost" onClick={() => onDelete(record.profileId)} title="删除" disabled={isBusy}><Trash2 className="h-3.5 w-3.5 shrink-0 text-red-500" /></Button>
          </div>
        )
      },
    },
  ]

  return (
    <div className="rounded-[10px] bg-[var(--color-bg-surface)]">
      <div className="overflow-auto" style={{ maxHeight: 'calc(100vh - 340px)' }}>
        {loading ? (
          <div className="py-16 flex items-center justify-center text-sm text-[var(--color-text-muted)]">加载中...</div>
        ) : profiles.length === 0 ? (
          <div className="py-16 flex items-center justify-center text-sm text-[var(--color-text-muted)]">暂无数据</div>
        ) : viewMode === 'table' ? (
          <Table
            columns={columns}
            data={profiles}
            rowKey="profileId"
          />
        ) : (
          <div className="flex flex-wrap gap-4 min-h-[500px] p-4 items-start content-start">
            {profiles.map((profile) => {
              const proxy = proxies.find(item => item.proxyId === profile.proxyId)
              return (
                <div key={profile.profileId} className="min-w-[360px] max-w-[560px] flex-[1_1_440px]">
                  <BrowserProfileCard
                    profile={profile}
                    proxy={proxy}
                    countryDisplay={getProfileCountryDisplay(profile, proxy)}
                    isSelected={selectedIds.has(profile.profileId)}
                    status={getProfileStatus(profile)}
                    coreLabel={resolveProfileCore(profile)?.coreName || getProfileCoreLabel(profile)}
                    isStarting={isProfileStarting(profile.profileId)}
                    isStopping={isProfileStopping(profile.profileId)}
                    isBusy={isProfileBusy(profile.profileId)}
                    onToggleSelect={onToggleSelect}
                    onRefreshProfiles={onRefreshProfiles}
                    onStart={onStart}
                    onStop={onStop}
                    onRestart={onRestart}
                    onOpenKeywords={onOpenKeywords}
                    onOpenExtensions={onOpenExtensions}
                    onOpenCopy={onOpenCopy}
                    onOpenProxyPicker={onOpenProxyPicker}
                    onDelete={onDelete}
                  />
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
