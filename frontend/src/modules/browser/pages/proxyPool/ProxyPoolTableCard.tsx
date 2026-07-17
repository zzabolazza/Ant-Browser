import { useMemo } from 'react'

import { Button, Card, Input, Table } from '../../../../shared/components'
import type { SortOrder, TableColumn } from '../../../../shared/components/Table'
import type { ProxyIPHealthResult } from '../../types'

import { BUILTIN_PROXY_IDS, type ProxyDisplayInfo } from './helpers'
import { resolveProxyCountryDisplay } from '../../utils/countryFlag'

interface ProxyPoolTableCardProps {
  allFilteredSelected: boolean
  checkingIPHealthIds: Set<string>
  data: ProxyDisplayInfo[]
  filterGroup: string
  filterKeyword: string
  filterProtocol: string
  filterAvailableOnly: boolean
  groups: string[]
  ipHealthMap: Record<string, ProxyIPHealthResult>
  loading: boolean
  checkingAllIPHealth: boolean
  testingAll: boolean
  onBatchCheckIPHealth: () => void
  onBatchTestSpeed: () => void
  onCheckOneIPHealth: (record: ProxyDisplayInfo) => void
  onClearFilters: () => void
  onDelete: (proxyId: string) => void
  onEdit: (record: ProxyDisplayInfo) => void
  onFilterGroupChange: (nextValue: string) => void
  onFilterKeywordChange: (nextValue: string) => void
  onFilterProtocolChange: (nextValue: string) => void
  onFilterAvailableOnlyChange: (checked: boolean) => void
  onOpenBatchDelete: () => void
  onOpenIPHealthDetail: (proxyId: string) => void
  onSort: (next: { column: string; order: SortOrder }) => void
  onTestOne: (record: ProxyDisplayInfo) => void
  onToggleAll: () => void
  onToggleOne: (proxyId: string) => void
  protocolOptions: string[]
  selectedCount: number
  selectedIds: Set<string>
  someFilteredSelected: boolean
  sortColumn: string
  sortOrder: SortOrder
  latencyMap: Record<string, number>
  latencyEngineMap: Record<string, string>
  latencyErrorMap: Record<string, string>
}

export function ProxyPoolTableCard({
  allFilteredSelected,
  checkingIPHealthIds,
  data,
  filterGroup,
  filterKeyword,
  filterProtocol,
  filterAvailableOnly,
  groups,
  ipHealthMap,
  loading,
  onCheckOneIPHealth,
  onClearFilters,
  onDelete,
  onEdit,
  onFilterGroupChange,
  onFilterKeywordChange,
  onFilterProtocolChange,
  onFilterAvailableOnlyChange,
  onOpenBatchDelete,
  onOpenIPHealthDetail,
  onSort,
  onTestOne,
  onToggleAll,
  onToggleOne,
  protocolOptions,
  selectedCount,
  selectedIds,
  someFilteredSelected,
  sortColumn,
  sortOrder,
  latencyMap,
  latencyEngineMap,
  latencyErrorMap,
  checkingAllIPHealth,
  testingAll,
  onBatchCheckIPHealth,
  onBatchTestSpeed,
}: ProxyPoolTableCardProps) {
  const hasActiveFilters = filterProtocol !== 'all' || !!filterKeyword || filterGroup !== 'all' || filterAvailableOnly

  const renderLatency = (record: ProxyDisplayInfo) => {
    const value = latencyMap[record.proxyId]
    if (value === undefined) return <span className="text-[var(--color-text-muted)] text-xs">-</span>
    if (value === -1) return <span className="text-[var(--color-text-muted)] text-xs animate-pulse">测试中...</span>
    const error = latencyErrorMap[record.proxyId] || ''
    if (value === -2) return <span className="text-red-500 text-xs" title={error || '测速超时'}>超时</span>
    if (value === -3) return <span className="text-gray-400 text-xs" title={error || '协议不支持'}>不支持</span>
    if (value === -4) return <span className="text-red-500 text-xs" title={error || '测速失败'}>失败</span>
    const color = value < 200 ? 'text-green-500' : value < 500 ? 'text-yellow-500' : 'text-red-500'
    return <span className={`text-xs font-medium ${color}`}>{value} ms</span>
  }

  const renderLatencyEngine = (record: ProxyDisplayInfo) => {
    const value = latencyMap[record.proxyId]
    if (value === undefined || value === -1) return <span className="text-[var(--color-text-muted)] text-xs">-</span>
    return latencyEngineMap[record.proxyId]
      ? <span className="text-xs text-[var(--color-text-secondary)] whitespace-nowrap">{latencyEngineMap[record.proxyId]}</span>
      : <span className="text-[var(--color-text-muted)] text-xs">-</span>
  }

  const renderIPHealth = (record: ProxyDisplayInfo) => {
    if (checkingIPHealthIds.has(record.proxyId)) {
      return <span className="text-[var(--color-text-muted)] text-xs animate-pulse">检测中...</span>
    }

    const result = ipHealthMap[record.proxyId]
    if (!result) return <span className="text-[var(--color-text-muted)] text-xs">-</span>
    if (!result.ok) {
      return (
        <div className="flex items-center gap-2">
          <span className="text-xs text-red-500 truncate max-w-[120px]" title={result.error || '检测失败'}>失败</span>
          <Button size="sm" variant="ghost" onClick={(event) => { event.stopPropagation(); onOpenIPHealthDetail(record.proxyId) }}>原始</Button>
        </div>
      )
    }

    const location = [result.country, result.region, result.city].filter(Boolean).join(' / ')
    return (
      <div className="flex items-center gap-2 min-w-0">
        <div className="min-w-0">
          <div className="text-xs text-[var(--color-text-primary)] truncate">{result.ip || '-'}</div>
          <div className="text-[11px] text-[var(--color-text-muted)] truncate">
            {`fraud ${result.fraudScore} | ${result.isResidential ? '住宅' : '机房'}${location ? ` | ${location}` : ''}`}
          </div>
        </div>
        <Button size="sm" variant="ghost" onClick={(event) => { event.stopPropagation(); onOpenIPHealthDetail(record.proxyId) }}>原始</Button>
      </div>
    )
  }

  const renderCountry = (record: ProxyDisplayInfo) => {
    if (checkingIPHealthIds.has(record.proxyId)) {
      return <span className="text-[var(--color-text-muted)] text-xs animate-pulse">检测中...</span>
    }
    const display = resolveProxyCountryDisplay(ipHealthMap[record.proxyId])
    if (!display) return <span className="text-[var(--color-text-muted)] text-xs">-</span>
    return (
      <span className="text-xs text-[var(--color-text-primary)] whitespace-nowrap" title={display.code}>
        {display.flag} {display.code}
      </span>
    )
  }

  const columns = useMemo<TableColumn<ProxyDisplayInfo>[]>(() => [
    {
      key: 'checkbox',
      title: '',
      width: '40px',
      render: (_, record) => (
        <input
          type="checkbox"
          checked={selectedIds.has(record.proxyId)}
          disabled={BUILTIN_PROXY_IDS.has(record.proxyId)}
          onChange={() => onToggleOne(record.proxyId)}
          onClick={(event) => event.stopPropagation()}
          className="w-4 h-4 rounded border-[var(--color-border-default)] accent-[var(--color-accent)] cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
        />
      ),
    },
    { key: 'proxyName', title: '代理名称', width: '180px', sortable: true },
    {
      key: 'groupName',
      title: '分组',
      width: '100px',
      sortable: true,
      render: (value) => value ? <span className="px-1.5 py-0.5 text-xs rounded bg-[var(--color-accent)]/10 text-[var(--color-accent)]">{value}</span> : '-',
    },
    { key: 'type', title: '类型', width: '90px', sortable: true },
    { key: 'server', title: '服务器', width: '180px', sortable: true },
    {
      key: 'country',
      title: '国家/地区',
      width: '100px',
      render: (_, record) => renderCountry(record),
    },
    { key: 'port', title: '端口', width: '80px', sortable: true, render: (value) => value || '-' },
    {
      key: 'latency',
      title: '延迟',
      width: '90px',
      sortable: true,
      render: (_, record) => renderLatency(record),
    },
    {
      key: 'latencyEngine',
      title: '协议',
      width: '90px',
      render: (_, record) => renderLatencyEngine(record),
    },
    {
      key: 'ipHealth',
      title: (
        <div className="leading-tight">
          <div>IP健康</div>
          <div className="mt-0.5 text-[10px] font-normal text-[var(--color-text-muted)]">仅供参考</div>
        </div>
      ),
      width: '280px',
      render: (_, record) => renderIPHealth(record),
    },
    {
      key: 'actions',
      title: '操作',
      width: '280px',
      render: (_, record) => {
        const isBuiltin = BUILTIN_PROXY_IDS.has(record.proxyId)
        return (
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="ghost"
              onClick={(event) => { event.stopPropagation(); onTestOne(record) }}
              loading={latencyMap[record.proxyId] === -1}
            >
              测速
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={(event) => { event.stopPropagation(); onCheckOneIPHealth(record) }}
              loading={checkingIPHealthIds.has(record.proxyId)}
            >
              IP健康
            </Button>
            <Button
              size="sm"
              variant="ghost"
              disabled={isBuiltin}
              onClick={(event) => {
                event.stopPropagation()
                if (!isBuiltin) onEdit(record)
              }}
            >
              编辑
            </Button>
            <Button
              size="sm"
              variant="danger"
              disabled={isBuiltin}
              onClick={(event) => {
                event.stopPropagation()
                if (!isBuiltin) onDelete(record.proxyId)
              }}
            >
              删除
            </Button>
          </div>
        )
      },
    },
  ], [
    checkingIPHealthIds,
    ipHealthMap,
    latencyMap,
    latencyEngineMap,
    onCheckOneIPHealth,
    onDelete,
    onEdit,
    onOpenIPHealthDetail,
    onTestOne,
    onToggleOne,
    selectedIds,
  ])

  return (
    <Card padding="sm">
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <Input
          value={filterKeyword}
          onChange={(event) => onFilterKeywordChange(event.target.value)}
          placeholder="搜索名称或服务器..."
          style={{ width: '220px' }}
        />
        <select
          value={filterProtocol}
          onChange={(event) => onFilterProtocolChange(event.target.value)}
          className="h-8 rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-muted)] px-2.5 text-[12.5px] text-[var(--color-text-primary)]"
        >
          {protocolOptions.map((protocol) => (
            <option key={protocol} value={protocol}>{protocol === 'all' ? '全部协议' : protocol.toUpperCase()}</option>
          ))}
        </select>
        <select
          value={filterGroup}
          onChange={(event) => onFilterGroupChange(event.target.value)}
          className="h-8 rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-muted)] px-2.5 text-[12.5px] text-[var(--color-text-primary)]"
        >
          <option value="all">全部分组</option>
          {groups.map((group) => <option key={group} value={group}>{group}</option>)}
        </select>
        {hasActiveFilters && (
          <Button size="sm" variant="ghost" onClick={onClearFilters}>清除筛选</Button>
        )}
        <label className="flex cursor-pointer select-none items-center gap-1.5 text-[12.5px] text-[var(--color-text-secondary)]">
          <input
            type="checkbox"
            checked={filterAvailableOnly}
            onChange={(event) => onFilterAvailableOnlyChange(event.target.checked)}
            className="h-4 w-4 cursor-pointer rounded border-[var(--color-border-default)] accent-[var(--color-accent)]"
          />
          只展示可用
        </label>
        <div className="flex-1" />
        {data.length > 0 && (
          <label className="flex cursor-pointer select-none items-center gap-1.5 text-[12.5px] text-[var(--color-text-muted)]">
            <input
              type="checkbox"
              checked={allFilteredSelected}
              ref={(element) => {
                if (element) {
                  element.indeterminate = someFilteredSelected && !allFilteredSelected
                }
              }}
              onChange={onToggleAll}
              className="h-4 w-4 cursor-pointer rounded border-[var(--color-border-default)] accent-[var(--color-accent)]"
            />
            全选
          </label>
        )}
        {selectedCount > 0 && (
          <>
            <Button size="sm" variant="secondary" onClick={onBatchCheckIPHealth} loading={checkingAllIPHealth}>
              批量健康检测 ({selectedCount})
            </Button>
            <Button size="sm" variant="secondary" onClick={onBatchTestSpeed} loading={testingAll}>
              批量测速 ({selectedCount})
            </Button>
            <Button size="sm" variant="danger" onClick={onOpenBatchDelete}>
              删除所选 ({selectedCount})
            </Button>
          </>
        )}
      </div>
      <Table
        columns={columns}
        data={data}
        rowKey="proxyId"
        loading={loading}
        emptyText="暂无代理配置，点击上方按钮添加"
        sortColumn={sortColumn}
        sortOrder={sortOrder}
        onSort={onSort}
        className="rounded-[10px] border-0"
      />
    </Card>
  )
}
