import { useCallback, useEffect, useMemo, useState } from 'react'
import { ConfirmModal, toast } from '../../../shared/components'
import type { SortOrder } from '../../../shared/components/Table'
import type { BrowserProxy, ProxyIPHealthResult } from '../types'
import { fetchBrowserProxies, fetchBrowserProxyGroups, saveBrowserProxies } from '../api'
import {
  buildDirectImportCandidate,
  ensureBuiltinProxies,
  formFromProxyConfig,
  toDisplayList,
  type ProxyDisplayInfo,
} from './proxyPool/helpers'
import {
  ProxyPoolEditModal,
  ProxyPoolIPHealthDetailModal,
  ProxyPoolImportModal,
  ProxyPoolPreviewModal,
  type ProxyEditFormValue,
} from './proxyPool/ProxyPoolModals'
import { ProxyPoolHeader } from './proxyPool/ProxyPoolHeader'
import { ProxyPoolTableCard } from './proxyPool/ProxyPoolTableCard'
import { ProxyPoolCheckSettingsModal } from './proxyPool/ProxyPoolCheckSettingsModal'
import { useProxyImportFlow } from './proxyPool/useProxyImportFlow'
import { useProxyChecks } from './proxyPool/useProxyChecks'
import { useProxySelection } from './proxyPool/useProxySelection'
import { useProxyCheckSettingsModal } from './proxyPool/useProxyCheckSettingsModal'
import { useProxyDeleteFlow } from './proxyPool/useProxyDeleteFlow'
import { useProxyPoolFilter } from './proxyPool/useProxyPoolFilter'

export function ProxyPoolPage() {
  const [proxies, setProxies] = useState<BrowserProxy[]>([])
  const [displayList, setDisplayList] = useState<ProxyDisplayInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [groups, setGroups] = useState<string[]>([])

  const [filterProtocol, setFilterProtocol] = useState<string>('all')
  const [filterKeyword, setFilterKeyword] = useState('')
  const [filterGroup, setFilterGroup] = useState<string>('all')
  const [filterAvailableOnly, setFilterAvailableOnly] = useState(false)
  const [sortColumn, setSortColumn] = useState<string>('')
  const [sortOrder, setSortOrder] = useState<SortOrder>(undefined)

  const {
    checkSettingsOpen,
    setCheckSettingsOpen,
    checkSettings,
    setCheckSettings,
    checkTargetsText,
    setCheckTargetsText,
    savingCheckSettings,
    openCheckSettings,
    saveCheckSettings,
  } = useProxyCheckSettingsModal()

  const [editModalOpen, setEditModalOpen] = useState(false)
  const [editingProxy, setEditingProxy] = useState<BrowserProxy | null>(null)
  const [editForm, setEditForm] = useState<ProxyEditFormValue>({
    proxyName: '',
    protocol: 'http',
    server: '',
    port: '',
    username: '',
    password: '',
    groupName: '',
  })
  const [saving, setSaving] = useState(false)

  const saveProxies = useCallback(async (list: BrowserProxy[]) => {
    await saveBrowserProxies(list)
    setProxies(list)
    setDisplayList(toDisplayList(list))
    const grps = await fetchBrowserProxyGroups()
    setGroups(grps)
  }, [])

  const {
    importModalOpen,
    setImportModalOpen,
    importGroupName,
    setImportGroupName,
    directImportText,
    setDirectImportText,
    directImportForm,
    setDirectImportForm,
    previewModalOpen,
    setPreviewModalOpen,
    previewList,
    removedPreviewProxyNames,
    importing,
    canParseImport,
    handleRemovePreviewProxy,
    handleApplyDirectText,
    handleParseImport,
    handleConfirmImport,
  } = useProxyImportFlow({
    proxies,
    saveProxies,
  })

  const {
    latencyMap,
    latencyEngineMap,
    latencyErrorMap,
    testingAll,
    ipHealthMap,
    checkingIPHealthIds,
    checkingAllIPHealth,
    ipHealthDetailOpen,
    setIPHealthDetailOpen,
    currentIPHealthDetail,
    setLatencyMap,
    setLatencyEngineMap,
    setIPHealthMap,
    handleTestOne,
    handleTestAll,
    handleCheckOneIPHealth,
    handleCheckAllIPHealth,
    openIPHealthDetail,
  } = useProxyChecks({ proxies })

  const loadProxies = useCallback(async () => {
    setLoading(true)
    try {
      const [list, groupList] = await Promise.all([
        fetchBrowserProxies(),
        fetchBrowserProxyGroups(),
      ])
      const finalList = ensureBuiltinProxies(list)
      setProxies(finalList)
      setDisplayList(toDisplayList(finalList))
      setGroups(groupList)

      setLatencyMap((prev) => {
        const validIds = new Set(finalList.map((p) => p.proxyId))
        const next: Record<string, number> = {}
        Object.entries(prev).forEach(([proxyId, latency]) => {
          if (validIds.has(proxyId)) next[proxyId] = latency
        })
        return next
      })

      setLatencyEngineMap((prev) => {
        const validIds = new Set(finalList.map((p) => p.proxyId))
        const next: Record<string, string> = {}
        Object.entries(prev).forEach(([proxyId, engine]) => {
          if (validIds.has(proxyId)) next[proxyId] = engine
        })
        return next
      })

      setIPHealthMap((prev) => {
        const validIds = new Set(finalList.map((p) => p.proxyId))
        const next: Record<string, ProxyIPHealthResult> = {}
        Object.entries(prev).forEach(([proxyId, health]) => {
          if (validIds.has(proxyId)) next[proxyId] = health
        })
        return next
      })
    } catch (error: any) {
      toast.error(error?.message || '加载代理失败')
    } finally {
      setLoading(false)
    }
  }, [setIPHealthMap, setLatencyEngineMap, setLatencyMap])

  useEffect(() => {
    void loadProxies()
  }, [loadProxies])

  const { protocolOptions, filteredList } = useProxyPoolFilter({
    displayList,
    filterProtocol,
    filterKeyword,
    filterGroup,
    filterAvailableOnly,
    sortColumn,
    sortOrder,
    latencyMap,
    ipHealthMap,
  })

  const {
    selectedIds,
    selectedCount,
    allFilteredSelected,
    someFilteredSelected,
    batchDeleteConfirmOpen,
    setBatchDeleteConfirmOpen,
    handleToggleAll,
    handleToggleOne,
    handleBatchDeleteConfirm,
    removeSelectedId,
  } = useProxySelection({ proxies, filteredList, saveProxies })

  const selectedProxyList = useMemo(
    () => filteredList.filter((item) => selectedIds.has(item.proxyId)),
    [filteredList, selectedIds],
  )

  const handleEdit = (record: ProxyDisplayInfo) => {
    const proxy = proxies.find((p) => p.proxyId === record.proxyId)
    if (!proxy) return
    try {
      const form = formFromProxyConfig(proxy.proxyName, proxy.proxyConfig)
      setEditingProxy(proxy)
      setEditForm({
        ...form,
        groupName: proxy.groupName || '',
      })
      setEditModalOpen(true)
    } catch (error: any) {
      toast.error(error?.message || '当前代理配置无法编辑，请删除后重建')
    }
  }

  const handleSaveProxy = async () => {
    if (!editingProxy) return
    setSaving(true)
    try {
      const candidate = buildDirectImportCandidate({
        proxyName: editForm.proxyName,
        protocol: editForm.protocol,
        server: editForm.server,
        port: editForm.port,
        username: editForm.username,
        password: editForm.password,
      })
      const newProxies = proxies.map((p) =>
        p.proxyId === editingProxy.proxyId
          ? {
              ...p,
              proxyName: candidate.proxyName,
              proxyConfig: candidate.proxyConfig,
              groupName: editForm.groupName.trim() || undefined,
            }
          : p,
      )
      await saveProxies(newProxies)
      setEditModalOpen(false)
      toast.success('代理已更新')
    } catch (error: any) {
      toast.error(error?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const {
    deleteConfirmOpen,
    setDeleteConfirmOpen,
    handleDeleteClick,
    handleDeleteConfirm,
  } = useProxyDeleteFlow({ proxies, saveProxies, removeSelectedId })

  const { availableCount, avgLatencyMs } = useMemo(() => {
    const latencyValues = displayList
      .map((item) => latencyMap[item.proxyId])
      .filter((value): value is number => typeof value === 'number' && value >= 0)
    const available = latencyValues.length
    const avg = available === 0
      ? null
      : Math.round(latencyValues.reduce((sum, value) => sum + value, 0) / available)
    return { availableCount: available, avgLatencyMs: avg }
  }, [displayList, latencyMap])

  return (
    <div className="space-y-5 animate-fade-in">
      <ProxyPoolHeader
        onOpenSettings={() => void openCheckSettings()}
        onOpenImport={() => setImportModalOpen(true)}
        totalCount={displayList.length}
        availableCount={availableCount}
        avgLatencyMs={avgLatencyMs}
      />

      <ProxyPoolTableCard
        allFilteredSelected={allFilteredSelected}
        checkingIPHealthIds={checkingIPHealthIds}
        data={filteredList}
        filterGroup={filterGroup}
        filterKeyword={filterKeyword}
        filterProtocol={filterProtocol}
        filterAvailableOnly={filterAvailableOnly}
        groups={groups}
        ipHealthMap={ipHealthMap}
        latencyMap={latencyMap}
        latencyEngineMap={latencyEngineMap}
        latencyErrorMap={latencyErrorMap}
        loading={loading}
        checkingAllIPHealth={checkingAllIPHealth}
        testingAll={testingAll}
        onBatchCheckIPHealth={() => void handleCheckAllIPHealth(selectedProxyList)}
        onBatchTestSpeed={() => void handleTestAll(selectedProxyList)}
        onCheckOneIPHealth={(record) => void handleCheckOneIPHealth(record)}
        onClearFilters={() => {
          setFilterProtocol('all')
          setFilterKeyword('')
          setFilterGroup('all')
          setFilterAvailableOnly(false)
        }}
        onDelete={handleDeleteClick}
        onEdit={handleEdit}
        onFilterGroupChange={setFilterGroup}
        onFilterKeywordChange={setFilterKeyword}
        onFilterProtocolChange={setFilterProtocol}
        onFilterAvailableOnlyChange={setFilterAvailableOnly}
        onOpenBatchDelete={() => setBatchDeleteConfirmOpen(true)}
        onOpenIPHealthDetail={openIPHealthDetail}
        onSort={({ column, order }) => {
          setSortColumn(column)
          setSortOrder(order)
        }}
        onTestOne={(record) => void handleTestOne(record)}
        onToggleAll={handleToggleAll}
        onToggleOne={handleToggleOne}
        protocolOptions={protocolOptions}
        selectedCount={selectedCount}
        selectedIds={selectedIds}
        someFilteredSelected={someFilteredSelected}
        sortColumn={sortColumn}
        sortOrder={sortOrder}
      />

      <ProxyPoolImportModal
        open={importModalOpen}
        groups={groups}
        importGroupName={importGroupName}
        directImportText={directImportText}
        directImportForm={directImportForm}
        canParseImport={canParseImport}
        onClose={() => setImportModalOpen(false)}
        onParse={handleParseImport}
        onImportGroupNameChange={setImportGroupName}
        onDirectImportTextChange={setDirectImportText}
        onApplyDirectText={handleApplyDirectText}
        onDirectImportFormChange={(patch) => setDirectImportForm((prev) => ({ ...prev, ...patch }))}
      />

      <ProxyPoolPreviewModal
        open={previewModalOpen}
        previewList={previewList}
        removedPreviewProxyNames={removedPreviewProxyNames}
        importing={importing}
        onClose={() => setPreviewModalOpen(false)}
        onBack={() => {
          setPreviewModalOpen(false)
          setImportModalOpen(true)
        }}
        onConfirm={handleConfirmImport}
        onRemoveProxy={handleRemovePreviewProxy}
      />

      <ProxyPoolEditModal
        open={editModalOpen}
        saving={saving}
        groups={groups}
        editForm={editForm}
        onClose={() => setEditModalOpen(false)}
        onSave={handleSaveProxy}
        onChange={(patch) => setEditForm((prev) => ({ ...prev, ...patch }))}
      />

      <ProxyPoolIPHealthDetailModal
        open={ipHealthDetailOpen}
        detail={currentIPHealthDetail}
        onClose={() => setIPHealthDetailOpen(false)}
      />

      <ProxyPoolCheckSettingsModal
        open={checkSettingsOpen}
        checkSettings={checkSettings}
        checkTargetsText={checkTargetsText}
        saving={savingCheckSettings}
        onClose={() => setCheckSettingsOpen(false)}
        onSave={saveCheckSettings}
        onCheckSettingsChange={setCheckSettings}
        onCheckTargetsTextChange={setCheckTargetsText}
      />

      <ConfirmModal
        open={deleteConfirmOpen}
        onClose={() => setDeleteConfirmOpen(false)}
        onConfirm={handleDeleteConfirm}
        title="确认删除"
        content="确定要删除这个代理吗？此操作不可恢复。"
        confirmText="删除"
        danger
      />

      <ConfirmModal
        open={batchDeleteConfirmOpen}
        onClose={() => setBatchDeleteConfirmOpen(false)}
        onConfirm={handleBatchDeleteConfirm}
        title="批量删除"
        content={`确定要删除选中的 ${selectedCount} 个代理吗？此操作不可恢复。`}
        confirmText="删除"
        danger
      />
    </div>
  )
}
