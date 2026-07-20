import clsx from 'clsx'
import { AlertTriangle, Download, ExternalLink, History, LayoutGrid, List, Power, Puzzle, RefreshCw, RotateCw, Search, Settings, Trash2, Users } from 'lucide-react'
import { Button, Card, Input } from '../../../shared/components'
import type { BrowserExtension, BrowserExtensionLookupResult, BrowserProxy } from '../types'
import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'
import { extensionStoreURL, formatExtensionSource, formatExtensionTime, getExtensionManifestMeta, getProxySpeedState } from './extensionManagementUtils'

const CHROME_WEB_STORE_URL = 'https://chromewebstore.google.com/category/extensions'

export type ExtensionViewMode = 'list' | 'grid'

export function ProxyStatePill({ useProxy, proxy }: { useProxy: boolean; proxy?: BrowserProxy }) {
  if (!useProxy) {
    return <span className="rounded-full bg-[var(--color-bg-muted)] px-2 py-0.5 text-xs text-[var(--color-text-muted)]">直连下载</span>
  }
  if (!proxy) {
    return <span className="rounded-full bg-red-50 px-2 py-0.5 text-xs text-red-600">代理未选择</span>
  }
  const state = getProxySpeedState(proxy)
  const status = state?.ok ? `${state.latencyMs}ms` : state ? '不可用' : '未测试'
  return (
    <span className="rounded-full bg-green-50 px-2 py-0.5 text-xs text-green-700">
      使用代理：{proxy.proxyName || proxy.proxyId} · {status}
    </span>
  )
}


export interface ExtensionManagementHeaderProps {
  proxyButtonText: string
  loading: boolean
  importing: 'none' | 'file'
  onOpenProxy: () => void
  onOpenHistory: () => void
  onImportFile: () => void
  onRefresh: () => void
}

export function ExtensionManagementHeader({
  proxyButtonText,
  loading,
  importing,
  onOpenProxy,
  onOpenHistory,
  onImportFile,
  onRefresh,
}: ExtensionManagementHeaderProps) {
  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="max-w-2xl text-[12.5px] leading-5 text-[var(--color-text-muted)]">
          统一管理扩展包，按需挂载到指定实例范围，支持
          <button
            type="button"
            onClick={() => BrowserOpenURL(CHROME_WEB_STORE_URL)}
            className="mx-0.5 inline-flex items-center gap-0.5 font-medium text-[var(--color-primary)] underline underline-offset-2 hover:opacity-80"
          >
            商店
            <ExternalLink className="h-3 w-3" />
          </button>
          链接安装与本地导入。
        </p>
        <div className="flex flex-wrap justify-end gap-2">
          <Button size="sm" variant="secondary" onClick={onOpenProxy}>
            <Settings className="h-4 w-4" />
            {proxyButtonText}
          </Button>
          <Button size="sm" variant="secondary" onClick={onOpenHistory}>
            <History className="h-4 w-4" />
            历史
          </Button>
          <Button size="sm" variant="secondary" onClick={onImportFile} loading={importing === 'file'}>
            <Download className="h-4 w-4" />
            导入
          </Button>
          <Button size="sm" variant="secondary" onClick={onRefresh} loading={loading}>
            <RefreshCw className="h-4 w-4" />
            刷新
          </Button>
        </div>
      </div>
      <div className="flex gap-2 rounded-lg border border-[var(--color-warning)]/40 bg-[var(--color-warning)]/10 p-3 text-xs leading-5 text-[var(--color-warning)]">
        <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
        <p>
          <span className="font-medium">Google 登录限制：</span>
          部分扩展依赖 Chrome Identity API，回调地址为{' '}
          <code className="rounded bg-black/5 px-1 py-0.5 font-mono text-[11px]">{'https://<扩展ID>.chromiumapp.org/'}</code>
          。当前使用的 fingerprint-chromium（Ungoogled Chromium）会改写该域名，导致 Google OAuth 校验失败。此类扩展的商店登录可能无法完成，与插件安装是否成功无关。
        </p>
      </div>
    </div>
  )
}

export interface ExtensionInstallCardProps {
  query: string
  lookup: BrowserExtensionLookupResult | null
  querying: boolean
  installing: boolean
  useProxy: boolean
  selectedProxy?: BrowserProxy
  installedIds: Set<string>
  lastLookupProxyLabel: string
  onQueryChange: (value: string) => void
  onLookup: () => void
  onOpenProxy: () => void
  onInstall: () => void
}

export function ExtensionInstallCard({
  query,
  lookup,
  querying,
  installing,
  useProxy,
  selectedProxy,
  installedIds,
  lastLookupProxyLabel,
  onQueryChange,
  onLookup,
  onOpenProxy,
  onInstall,
}: ExtensionInstallCardProps) {
  return (
    <Card padding="sm">
      <div className="flex flex-col gap-3 md:flex-row md:items-center">
        <Input
          value={query}
          onChange={(event) => onQueryChange(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === 'Enter') onLookup()
          }}
          placeholder="粘贴 Chrome 应用商店 URL 或扩展 ID…"
          className="flex-1 font-mono text-[12.5px]"
        />
        <ProxyStatePill useProxy={useProxy} proxy={selectedProxy} />
        <Button type="button" size="sm" variant="ghost" onClick={onOpenProxy}>
          切换代理
        </Button>
        <Button type="button" size="sm" onClick={onLookup} loading={querying}>
          <Search className="h-4 w-4" />
          解析
        </Button>
      </div>

      {lookup ? (
        <div className="mt-3 flex flex-col gap-3 rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-muted)] p-3 md:flex-row md:items-center md:justify-between">
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2 font-medium text-[var(--color-text-primary)]">
              <span>{lookup.name || lookup.extensionId}</span>
              {lookup.version ? <span className="text-xs font-normal text-[var(--color-text-muted)]">v{lookup.version}</span> : null}
            </div>
            <div className="mt-1 break-all font-mono text-xs text-[var(--color-text-muted)]">{lookup.extensionId}</div>
            {lastLookupProxyLabel ? <div className="mt-1 text-xs text-[var(--color-text-muted)]">本次查询：{lastLookupProxyLabel}</div> : null}
            {lookup.description ? <div className="mt-1 line-clamp-2 text-xs text-[var(--color-text-muted)]">{lookup.description}</div> : null}
            {lookup.message ? <div className="mt-1 text-xs text-[var(--color-text-muted)]">{lookup.message}</div> : null}
          </div>
          <div className="flex shrink-0 gap-2">
            {lookup.storeUrl ? (
              <Button type="button" size="sm" variant="secondary" onClick={() => BrowserOpenURL(lookup.storeUrl)}>
                <ExternalLink className="h-4 w-4" />
                商店
              </Button>
            ) : null}
            <Button
              type="button"
              size="sm"
              onClick={onInstall}
              loading={installing}
              disabled={!lookup.installable || installedIds.has(lookup.extensionId)}
            >
              <Download className="h-4 w-4" />
              {installedIds.has(lookup.extensionId) ? '已安装' : '安装'}
            </Button>
          </div>
        </div>
      ) : null}
    </Card>
  )
}

export interface InstalledExtensionsListProps {
  items: BrowserExtension[]
  busyId: string
  updatingId: string
  viewMode: ExtensionViewMode
  onViewModeChange: (next: ExtensionViewMode) => void
  onRestrictProfiles: (item: BrowserExtension) => void
  onUpdate: (item: BrowserExtension) => void
  onToggle: (item: BrowserExtension) => void
  onDelete: (item: BrowserExtension) => void
}

export function InstalledExtensionsList({
  items,
  busyId,
  updatingId,
  viewMode,
  onViewModeChange,
  onRestrictProfiles,
  onUpdate,
  onToggle,
  onDelete,
}: InstalledExtensionsListProps) {
  return (
    <Card>
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <div className="text-sm font-medium text-[var(--color-text-primary)]">已安装插件（{items.length}）</div>
        <div className="flex overflow-hidden rounded-md border border-[var(--color-border-default)]">
          <button
            type="button"
            className={clsx(
              "flex h-8 w-8 items-center justify-center text-[var(--color-text-muted)] transition-colors hover:text-[var(--color-text-primary)]",
              viewMode === 'grid' && "bg-[var(--color-bg-muted)] text-[var(--color-accent)]",
            )}
            onClick={() => onViewModeChange('grid')}
            title="多列网格"
          >
            <LayoutGrid className="h-4 w-4" />
          </button>
          <button
            type="button"
            className={clsx(
              "flex h-8 w-8 items-center justify-center text-[var(--color-text-muted)] transition-colors hover:text-[var(--color-text-primary)]",
              viewMode === 'list' && "bg-[var(--color-bg-muted)] text-[var(--color-accent)]",
            )}
            onClick={() => onViewModeChange('list')}
            title="纵向列表"
          >
            <List className="h-4 w-4" />
          </button>
        </div>
      </div>
      {items.length === 0 ? (
        <div className="rounded-xl border border-dashed border-[var(--color-border-default)] bg-[var(--color-bg-muted)] px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">
          暂无插件，先通过上方输入插件 ID 或商店链接安装。
        </div>
      ) : (
        <div className={viewMode === 'grid' ? 'grid gap-3 sm:grid-cols-2 xl:grid-cols-3' : 'space-y-2'}>
          {items.map((item) => (
            <InstalledExtensionCard
              key={item.extensionId}
              item={item}
              viewMode={viewMode}
              busy={busyId === item.extensionId}
              updating={updatingId === item.extensionId}
              onRestrictProfiles={onRestrictProfiles}
              onUpdate={onUpdate}
              onToggle={onToggle}
              onDelete={onDelete}
            />
          ))}
        </div>
      )}
    </Card>
  )
}

export interface InstalledExtensionCardProps {
  item: BrowserExtension
  viewMode: ExtensionViewMode
  busy: boolean
  updating: boolean
  onRestrictProfiles: (item: BrowserExtension) => void
  onUpdate: (item: BrowserExtension) => void
  onToggle: (item: BrowserExtension) => void
  onDelete: (item: BrowserExtension) => void
}

export function InstalledExtensionCard({ item, viewMode, busy, updating, onRestrictProfiles, onUpdate, onToggle, onDelete }: InstalledExtensionCardProps) {
  const meta = getExtensionManifestMeta(item)
  const storeUrl = extensionStoreURL(item)
  const isGrid = viewMode === 'grid'
  return (
    <div className={clsx(
      "rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-3 shadow-[var(--shadow-xs)]",
      isGrid && "h-full",
    )}>
      <div className={clsx(
        "flex gap-3",
        isGrid ? "h-full flex-col" : "flex-col md:flex-row md:items-start md:justify-between",
      )}>
        <div className={clsx("flex min-w-0 gap-3", isGrid && "flex-1")}>
          <div className={clsx(
            "flex shrink-0 items-center justify-center overflow-hidden rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-muted)]",
            isGrid ? "h-11 w-11" : "h-12 w-12",
          )}>
            {item.iconDataUrl ? (
              <img src={item.iconDataUrl} alt="" className={isGrid ? "h-8 w-8 object-contain" : "h-9 w-9 object-contain"} />
            ) : (
              <Puzzle className={isGrid ? "h-5 w-5 text-[var(--color-text-muted)]" : "h-6 w-6 text-[var(--color-text-muted)]"} />
            )}
          </div>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <span className="font-medium text-[var(--color-text-primary)]">{item.name || item.extensionId}</span>
              <span className="rounded-full bg-[var(--color-bg-muted)] px-2 py-0.5 text-xs text-[var(--color-text-muted)]">{item.enabled ? '已启用' : '已停用'}</span>
              {item.version ? <span className="text-xs text-[var(--color-text-muted)]">v{item.version}</span> : null}
              {meta.manifestVersion ? <span className="text-xs text-[var(--color-text-muted)]">MV{meta.manifestVersion}</span> : null}
            </div>
            {item.description ? <div className="mt-1 line-clamp-2 text-sm text-[var(--color-text-secondary)]">{item.description}</div> : null}
            <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-[var(--color-text-muted)]">
              <span className="break-all font-mono">{item.extensionId}</span>
              <span>{formatExtensionSource(item.sourceUrl)}</span>
              <span>安装：{formatExtensionTime(item.installedAt)}</span>
              {item.updatedAt ? <span>更新：{formatExtensionTime(item.updatedAt)}</span> : null}
            </div>
            {meta.permissions.length > 0 || meta.hostPermissionCount > 0 ? (
              <div className="mt-2 flex flex-wrap gap-1.5">
                {meta.permissions.map((permission) => (
                  <span key={permission} className="rounded-full bg-[var(--color-bg-muted)] px-2 py-0.5 text-xs text-[var(--color-text-muted)]">{permission}</span>
                ))}
                {meta.hostPermissionCount > 0 ? <span className="rounded-full bg-[var(--color-bg-muted)] px-2 py-0.5 text-xs text-[var(--color-text-muted)]">站点权限 {meta.hostPermissionCount}</span> : null}
              </div>
            ) : null}
          </div>
        </div>
        <div className={clsx(
          "flex shrink-0 flex-wrap gap-2",
          isGrid && "border-t border-[var(--color-border-muted)] pt-3",
        )}>
          {storeUrl ? (
            <Button type="button" size="sm" variant="secondary" onClick={() => BrowserOpenURL(storeUrl)}>
              <ExternalLink className="h-4 w-4" />
              商店
            </Button>
          ) : null}
          <Button type="button" size="sm" variant="secondary" onClick={() => onRestrictProfiles(item)}>
            <Users className="h-4 w-4" />
            限制实例
          </Button>
          <Button type="button" size="sm" variant="secondary" onClick={() => onUpdate(item)} loading={updating}>
            <RotateCw className="h-4 w-4" />
            更新
          </Button>
          <Button type="button" size="sm" variant="secondary" onClick={() => onToggle(item)} loading={busy}>
            <Power className="h-4 w-4" />
            {item.enabled ? '停用' : '启用'}
          </Button>
          <Button type="button" size="sm" variant="secondary" onClick={() => onDelete(item)} loading={busy}>
            <Trash2 className="h-4 w-4" />
            删除
          </Button>
        </div>
      </div>
    </div>
  )
}
