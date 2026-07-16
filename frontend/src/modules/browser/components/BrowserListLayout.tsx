import { Link } from 'react-router-dom'
import { Archive, LayoutGrid, List, Play, RefreshCw, Upload } from 'lucide-react'

import { Button } from '../../../shared/components'

import type { BrowserCore, BrowserGroupWithCount, BrowserProxy } from '../types'
import { InstanceFilterBar } from './InstanceFilterBar'
import type { InstanceFilters } from './InstanceFilterBar'

export type BrowserViewMode = 'card' | 'table'

interface BrowserListHeaderProps {
  profileCount: number
  filteredProfileCount: number
  runningCount: number
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
  const statItems = [
    { label: '总数', value: profileCount },
    { label: '运行', value: runningCount },
    { label: '停止', value: Math.max(0, profileCount - runningCount) },
  ]

  return (
    <>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-3 min-w-0">
          <h1 className="text-xl font-semibold text-[var(--color-text-primary)]">实例列表</h1>
          <div className="flex flex-wrap items-center gap-2">
            {statItems.map((item) => (
              <div
                key={item.label}
                className="flex h-8 items-center gap-2 rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-elevated)] px-3 text-sm"
              >
                <span className="text-[var(--color-text-muted)]">{item.label}</span>
                <span className="font-semibold text-[var(--color-text-primary)]">{item.value}</span>
              </div>
            ))}
            {filteredProfileCount !== profileCount && (
              <div className="flex h-8 items-center gap-2 rounded-lg border border-[var(--color-accent)]/30 bg-[var(--color-accent)]/5 px-3 text-sm">
                <span className="text-[var(--color-text-muted)]">筛选</span>
                <span className="font-semibold text-[var(--color-accent)]">{filteredProfileCount}</span>
              </div>
            )}
          </div>
        </div>
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="secondary" size="sm" onClick={onRefresh}>
            <RefreshCw className="w-4 h-4" />刷新
          </Button>
          <Button variant="secondary" size="sm" onClick={onImportProfiles} loading={importingProfiles}>
            <Upload className="w-4 h-4" />导入
          </Button>
          <Button variant="secondary" size="sm" onClick={onOpenBackup}>
            <Archive className="w-4 h-4" />备份
          </Button>
          <div className="flex items-center bg-[var(--color-bg-secondary)] rounded-md border border-[var(--color-border-default)] p-0.5 ml-2">
            <button
              className={`p-1.5 rounded text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] transition-colors ${viewMode === 'card' ? 'bg-[var(--color-bg-surface)] shadow-sm text-[var(--color-accent)]' : ''}`}
              onClick={() => onViewModeChange('card')}
              title="卡片视图"
            >
              <LayoutGrid className="w-4 h-4" />
            </button>
            <button
              className={`p-1.5 rounded text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] transition-colors ${viewMode === 'table' ? 'bg-[var(--color-bg-surface)] shadow-sm text-[var(--color-accent)]' : ''}`}
              onClick={() => onViewModeChange('table')}
              title="表格视图"
            >
              <List className="w-4 h-4" />
            </button>
          </div>
          <span className="w-px h-4 bg-[var(--color-border-muted)] mx-1 self-center"></span>
          <Link to="/browser/edit/new">
            <Button size="sm">
              <Play className="w-4 h-4" />新建配置
            </Button>
          </Link>
        </div>
      </div>
      <InstanceFilterBar
        filters={filters}
        onChange={onFiltersChange}
        proxies={proxies}
        cores={cores}
        allTags={allTags}
        groups={groups}
      />
    </>
  )
}
