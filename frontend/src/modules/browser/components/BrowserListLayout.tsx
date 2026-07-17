import { Link } from 'react-router-dom'
import clsx from 'clsx'
import { Download, LayoutGrid, List, Play, RefreshCw, Upload } from 'lucide-react'

import { Button } from '../../../shared/components'

import type { BrowserCore, BrowserGroupWithCount, BrowserProxy } from '../types'
import { InstanceFilterBar } from './InstanceFilterBar'
import type { InstanceFilters } from './InstanceFilterBar'

export type BrowserViewMode = 'card' | 'table'

interface BrowserListHeaderProps {
  profileCount: number
  filteredProfileCount: number
  runningCount: number
  errorProfileCount?: number
  viewMode: BrowserViewMode
  proxies: BrowserProxy[]
  cores: BrowserCore[]
  groups: BrowserGroupWithCount[]
  allTags: string[]
  filters: InstanceFilters
  onFiltersChange: (next: InstanceFilters) => void
  onRefresh: () => void
  onImportProfiles: () => void
  onOpenBackup: () => void
  importingProfiles?: boolean
  onViewModeChange: (next: BrowserViewMode) => void
}

export function BrowserListHeader({
  profileCount,
  filteredProfileCount,
  runningCount,
  errorProfileCount = 0,
  viewMode,
  proxies,
  cores,
  groups,
  allTags,
  filters,
  onFiltersChange,
  onRefresh,
  onImportProfiles,
  onOpenBackup,
  importingProfiles = false,
  onViewModeChange,
}: BrowserListHeaderProps) {
  const stoppedCount = Math.max(0, profileCount - runningCount)
  const statItems = [
    { label: '实例总数', value: profileCount, tone: 'default' },
    { label: '运行中', value: runningCount, tone: 'success' },
    { label: '已停止', value: stoppedCount, tone: 'default' },
    { label: '代理异常', value: errorProfileCount, tone: 'error' },
  ] as const

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="max-w-2xl text-[12.5px] leading-5 text-[var(--color-text-muted)]">
            管理全部浏览器配置与实例：启动、停止、代理切换、标签分组、批量操作。
          </p>
        </div>
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="secondary" size="sm" onClick={onRefresh}>
            <RefreshCw className="h-4 w-4" />刷新
          </Button>
          <Button variant="secondary" size="sm" onClick={onImportProfiles} loading={importingProfiles}>
            <Download className="h-4 w-4" />导入
          </Button>
          <Button variant="secondary" size="sm" onClick={onOpenBackup}>
            <Upload className="h-4 w-4" />导出
          </Button>
          <Link to="/browser/edit/new">
            <Button size="sm">
              <Play className="h-4 w-4" />新建配置
            </Button>
          </Link>
        </div>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {statItems.map((item) => (
          <div
            key={item.label}
            className="rounded-[10px] border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-4 py-3"
          >
            <div className="text-[11.5px] font-semibold text-[var(--color-text-muted)]">
              {item.label}
            </div>
            <div className={clsx(
              "mt-1.5 text-2xl font-extrabold tracking-tight",
              item.tone === 'success' && "text-[var(--color-success)]",
              item.tone === 'error' && "text-[var(--color-error)]",
              item.tone === 'default' && "text-[var(--color-text-primary)]",
            )}>
              {item.value}
            </div>
          </div>
        ))}
      </div>

      <div className="rounded-[10px] border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-3">
        <div className="flex flex-wrap items-center gap-2">
          <InstanceFilterBar
            filters={filters}
            onChange={onFiltersChange}
            proxies={proxies}
            cores={cores}
            allTags={allTags}
            groups={groups}
          />
          <div className="ml-auto flex items-center gap-2">
            {filteredProfileCount !== profileCount && (
              <span className="rounded-md bg-[var(--color-accent-muted)] px-2 py-1 text-xs font-semibold text-[var(--color-accent)]">
                筛选 {filteredProfileCount}
              </span>
            )}
            <div className="flex overflow-hidden rounded-md border border-[var(--color-border-default)]">
              <button
                className={clsx(
                  "flex h-8 w-8 items-center justify-center text-[var(--color-text-muted)] transition-colors hover:text-[var(--color-text-primary)]",
                  viewMode === 'card' && "bg-[var(--color-bg-muted)] text-[var(--color-accent)]",
                )}
                onClick={() => onViewModeChange('card')}
                title="卡片视图"
              >
                <LayoutGrid className="h-4 w-4" />
              </button>
              <button
                className={clsx(
                  "flex h-8 w-8 items-center justify-center text-[var(--color-text-muted)] transition-colors hover:text-[var(--color-text-primary)]",
                  viewMode === 'table' && "bg-[var(--color-bg-muted)] text-[var(--color-accent)]",
                )}
                onClick={() => onViewModeChange('table')}
                title="表格视图"
              >
                <List className="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
