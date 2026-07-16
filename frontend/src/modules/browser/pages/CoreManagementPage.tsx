import { useEffect, useState, useCallback } from 'react'
import { FolderOpen } from 'lucide-react'
import { Badge, Button, Card, ConfirmModal, Table, toast } from '../../../shared/components'
import type { TableColumn } from '../../../shared/components/Table'
import type { BrowserCore, BrowserCoreInput, BrowserCoreValidateResult, BrowserSettings } from '../types'
import {
  fetchBrowserCores,
  saveBrowserCore,
  deleteBrowserCore,
  setDefaultBrowserCore,
  validateBrowserCorePath,
  openCorePath,
  fetchBrowserSettings,
  saveBrowserSettings,
  fetchCoreExtendedInfo,
  pickBrowserCoreDirectory,
} from '../api'
import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'
import { CoreEditModal } from './coreManagement/CoreEditModal'
import { CoreSettingsCard } from './coreManagement/CoreSettingsCard'
import { CoreSettingsModal } from './coreManagement/CoreSettingsModal'
import type { CoreDisplayInfo, CoreEditForm, CoreSettingsForm } from './coreManagement.types'

export function CoreManagementPage() {
  const [cores, setCores] = useState<BrowserCore[]>([])
  const [displayList, setDisplayList] = useState<CoreDisplayInfo[]>([])
  const [loading, setLoading] = useState(true)

  const [settings, setSettings] = useState<BrowserSettings>({
    userDataRoot: '',
    defaultFingerprintArgs: [],
    defaultLaunchArgs: [],
    defaultStartUrls: [],
    lightStartEnabled: true,
    restoreLastSession: false,
    startReadyTimeoutMs: 3000,
    startStableWindowMs: 1200,
    defaultConnectorType: 'xray',
  })
  const [settingsModalOpen, setSettingsModalOpen] = useState(false)
  const [settingsForm, setSettingsForm] = useState<CoreSettingsForm>({
    userDataRoot: '',
    defaultFingerprintArgs: '',
    defaultLaunchArgs: '',
    defaultStartUrls: '',
    lightStartEnabled: true,
    restoreLastSession: false,
    startReadyTimeoutMs: 3000,
    startStableWindowMs: 1200,
  })
  const [savingSettings, setSavingSettings] = useState(false)

  const [editModalOpen, setEditModalOpen] = useState(false)
  const [editingCore, setEditingCore] = useState<BrowserCore | null>(null)
  const [editForm, setEditForm] = useState<CoreEditForm>({ coreName: '', corePath: '' })
  const [saving, setSaving] = useState(false)
  const [pickingPath, setPickingPath] = useState(false)
  const [pathValidating, setPathValidating] = useState(false)
  const [pathValidResult, setPathValidResult] = useState<BrowserCoreValidateResult | null>(null)

  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [deletingCore, setDeletingCore] = useState<CoreDisplayInfo | null>(null)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [settingsData, coreList, extendedInfo] = await Promise.all([
        fetchBrowserSettings(),
        fetchBrowserCores(),
        fetchCoreExtendedInfo(),
      ])

      setSettings(settingsData)
      setCores(coreList)

      const extendedMap = new Map<string, { chromeVersion: string; instanceCount: number }>()
      extendedInfo.forEach(info => extendedMap.set(info.coreId, info))

      const displayInfoList: CoreDisplayInfo[] = await Promise.all(
        coreList.map(async (core) => {
          const result = await validateBrowserCorePath(core.corePath)
          const extended = extendedMap.get(core.coreId)
          return {
            coreId: core.coreId,
            coreName: core.coreName,
            corePath: core.corePath,
            isDefault: core.isDefault,
            pathValid: result.valid,
            pathMessage: result.message,
            chromeVersion: extended?.chromeVersion || '',
            instanceCount: extended?.instanceCount || 0,
          }
        })
      )
      setDisplayList(displayInfoList)
    } finally {
      setLoading(false)
    }
  }

  const validatePath = useCallback(async (path: string) => {
    if (!path.trim()) {
      setPathValidResult(null)
      return
    }
    setPathValidating(true)
    try {
      const result = await validateBrowserCorePath(path)
      setPathValidResult(result)
    } finally {
      setPathValidating(false)
    }
  }, [])

  useEffect(() => {
    const timer = setTimeout(() => {
      if (editModalOpen && editForm.corePath) {
        validatePath(editForm.corePath)
      }
    }, 500)
    return () => clearTimeout(timer)
  }, [editForm.corePath, editModalOpen, validatePath])

  const columns: TableColumn<CoreDisplayInfo>[] = [
    { key: 'coreName', title: '内核名称', width: '150px' },
    { key: 'corePath', title: '内核路径', width: '180px' },
    {
      key: 'chromeVersion',
      title: 'Chrome 版本',
      width: '130px',
      render: (val) => val || '-',
    },
    {
      key: 'instanceCount',
      title: '使用实例',
      width: '90px',
      render: (val) => <Badge variant="default">{val}</Badge>,
    },
    {
      key: 'isDefault',
      title: '默认',
      width: '70px',
      render: (val) => val ? <Badge variant="info">默认</Badge> : null,
    },
    {
      key: 'pathValid',
      title: '状态',
      width: '80px',
      render: (val) => (
        <Badge variant={val ? 'success' : 'error'}>
          {val ? '有效' : '无效'}
        </Badge>
      ),
    },
    {
      key: 'actions',
      title: '操作',
      width: '180px',
      render: (_, record) => (
        <div className="flex gap-2">
          <Button size="sm" variant="ghost" onClick={(e) => { e.stopPropagation(); handleOpenPath(record.corePath) }} title="打开目录">
            <FolderOpen className="w-4 h-4" />
          </Button>
          <Button size="sm" variant="ghost" onClick={(e) => { e.stopPropagation(); handleEdit(record) }}>
            编辑
          </Button>
          {!record.isDefault && (
            <Button size="sm" variant="ghost" onClick={(e) => { e.stopPropagation(); handleSetDefault(record.coreId) }}>
              设为默认
            </Button>
          )}
          <Button size="sm" variant="danger" onClick={(e) => { e.stopPropagation(); handleDeleteClick(record) }}>
            删除
          </Button>
        </div>
      ),
    },
  ]

  const handleOpenPath = async (corePath: string) => {
    try {
      await openCorePath(corePath)
    } catch (error: any) {
      toast.error(error?.message || '打开目录失败')
    }
  }

  const handleAdd = () => {
    setEditingCore(null)
    setEditForm({ coreName: '', corePath: '' })
    setPathValidResult(null)
    setEditModalOpen(true)
  }

  const handleEdit = (record: CoreDisplayInfo) => {
    const core = cores.find(c => c.coreId === record.coreId)
    if (core) {
      setEditingCore(core)
      setEditForm({ coreName: core.coreName, corePath: core.corePath })
      setPathValidResult({ valid: record.pathValid, message: record.pathMessage })
      setEditModalOpen(true)
    }
  }

  const handlePickDirectory = async () => {
    setPickingPath(true)
    try {
      const picked = await pickBrowserCoreDirectory()
      if (!picked) {
        return
      }
      setEditForm(prev => ({
        corePath: picked.corePath,
        coreName: prev.coreName.trim() || picked.suggestedName,
      }))
    } catch (error: any) {
      toast.error(error?.message || '选择目录失败')
    } finally {
      setPickingPath(false)
    }
  }

  const handleSaveCore = async () => {
    if (!editForm.coreName.trim()) {
      toast.error('请输入内核名称')
      return
    }
    if (!editForm.corePath.trim()) {
      toast.error('请选择内核目录')
      return
    }
    setSaving(true)
    try {
      const input: BrowserCoreInput = {
        coreId: editingCore?.coreId || `core-${Date.now()}`,
        coreName: editForm.coreName.trim(),
        corePath: editForm.corePath.trim(),
        isDefault: editingCore?.isDefault || false,
      }
      await saveBrowserCore(input)
      await loadData()
      setEditModalOpen(false)
      toast.success(editingCore ? '内核已更新' : '内核已添加')
    } catch (error: any) {
      toast.error(error?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleDeleteClick = (record: CoreDisplayInfo) => {
    if (record.isDefault) {
      toast.warning('默认内核不能删除')
      return
    }
    setDeletingCore(record)
    setDeleteConfirmOpen(true)
  }

  const handleDeleteConfirm = async () => {
    if (!deletingCore) return
    try {
      await deleteBrowserCore(deletingCore.coreId)
      await loadData()
      toast.success('内核已删除')
    } catch (error: any) {
      toast.error(error?.message || '删除失败')
    }
    setDeletingCore(null)
  }

  const handleSetDefault = async (coreId: string) => {
    try {
      await setDefaultBrowserCore(coreId)
      await loadData()
      toast.success('已设为默认内核')
    } catch (error: any) {
      toast.error(error?.message || '设置失败')
    }
  }

  const handleEditSettings = () => {
    setSettingsForm({
      userDataRoot: settings.userDataRoot,
      defaultFingerprintArgs: settings.defaultFingerprintArgs.join('\n'),
      defaultLaunchArgs: settings.defaultLaunchArgs.join('\n'),
      defaultStartUrls: settings.defaultStartUrls.join('\n'),
      lightStartEnabled: settings.lightStartEnabled,
      restoreLastSession: settings.restoreLastSession,
      startReadyTimeoutMs: settings.startReadyTimeoutMs,
      startStableWindowMs: settings.startStableWindowMs,
    })
    setSettingsModalOpen(true)
  }

  const handleSaveSettings = async () => {
    setSavingSettings(true)
    try {
      const newSettings: BrowserSettings = {
        userDataRoot: settingsForm.userDataRoot.trim(),
        defaultFingerprintArgs: settingsForm.defaultFingerprintArgs.split('\n').map(s => s.trim()).filter(Boolean),
        defaultLaunchArgs: settingsForm.defaultLaunchArgs.split('\n').map(s => s.trim()).filter(Boolean),
        defaultStartUrls: settingsForm.defaultStartUrls.split('\n').map(s => s.trim()).filter(Boolean),
        lightStartEnabled: settingsForm.lightStartEnabled,
        restoreLastSession: settingsForm.restoreLastSession,
        startReadyTimeoutMs: Math.max(1000, Number(settingsForm.startReadyTimeoutMs) || 3000),
        startStableWindowMs: Math.max(0, Number(settingsForm.startStableWindowMs) || 1200),
        defaultConnectorType: settings.defaultConnectorType || 'xray',
      }
      await saveBrowserSettings(newSettings)
      setSettings(newSettings)
      setSettingsModalOpen(false)
      toast.success('设置已保存')
    } catch (error: any) {
      toast.error(error?.message || '保存失败')
    } finally {
      setSavingSettings(false)
    }
  }

  return (
    <div className="space-y-5 animate-fade-in">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-[var(--color-text-primary)]">内核管理</h1>
          <p className="text-sm text-[var(--color-text-muted)] mt-1">
            管理{' '}
            <button
              type="button"
              onClick={() => BrowserOpenURL('https://github.com/adryfish/fingerprint-chromium/releases')}
              className="text-[var(--color-accent)] hover:underline cursor-pointer font-medium"
            >
              fingerprint-chromium
            </button>
            {' '}内核版本和全局设置
          </p>
        </div>
        <Button size="sm" onClick={handleAdd}>新增内核</Button>
      </div>

      <CoreSettingsCard settings={settings} onEdit={handleEditSettings} />

      <Card title="内核列表" subtitle="已配置的 Chrome 内核">
        <Table
          columns={columns}
          data={displayList}
          rowKey="coreId"
          loading={loading}
          emptyText="暂无内核，请添加内核"
        />
      </Card>

      <CoreSettingsModal
        open={settingsModalOpen}
        form={settingsForm}
        saving={savingSettings}
        setForm={setSettingsForm}
        onClose={() => setSettingsModalOpen(false)}
        onSave={handleSaveSettings}
      />

      <CoreEditModal
        open={editModalOpen}
        isEditing={Boolean(editingCore)}
        form={editForm}
        saving={saving}
        pathValidating={pathValidating}
        pickingPath={pickingPath}
        pathValidResult={pathValidResult}
        setForm={setEditForm}
        onClose={() => setEditModalOpen(false)}
        onSave={handleSaveCore}
        onPickDirectory={handlePickDirectory}
      />

      <ConfirmModal
        open={deleteConfirmOpen}
        onClose={() => setDeleteConfirmOpen(false)}
        onConfirm={handleDeleteConfirm}
        title="确认删除"
        content={`确定要删除内核"${deletingCore?.coreName}"吗？此操作不可恢复。`}
        confirmText="删除"
        danger
      />
    </div>
  )
}
