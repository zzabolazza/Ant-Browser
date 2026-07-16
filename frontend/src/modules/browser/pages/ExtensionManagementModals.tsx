import { useEffect, useMemo, useState } from 'react'
import { ExternalLink, Search } from 'lucide-react'
import { Button, Modal, toast } from '../../../shared/components'
import type { BrowserExtension, BrowserGroupWithCount, BrowserProfile, BrowserProfileExtensionSettings } from '../types'
import { fetchBrowserProfileExtensionSettings, saveBrowserProfileExtensionSettings } from '../api/extensions'
import { fetchGroups } from '../api/groups'
import { fetchBrowserProfiles } from '../api/profiles'
import { extensionHistoryActionLabel, formatExtensionTime, sameStringSet, type ExtensionHistoryRecord } from './extensionManagementUtils'

const UNGROUPED_PROFILE_GROUP_ID = '__ungrouped__'

export interface ExtensionProfileLimitModalProps {
  open: boolean
  extension: BrowserExtension | null
  allExtensions: BrowserExtension[]
  onClose: () => void
}

export function ExtensionProfileLimitModal({ open, extension, allExtensions, onClose }: ExtensionProfileLimitModalProps) {
  const [profiles, setProfiles] = useState<BrowserProfile[]>([])
  const [groups, setGroups] = useState<BrowserGroupWithCount[]>([])
  const [settingsByProfile, setSettingsByProfile] = useState<Record<string, BrowserProfileExtensionSettings>>({})
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)

  const selectedSet = useMemo(() => new Set(selectedIds), [selectedIds])
  const enabledExtensionIds = useMemo(
    () => allExtensions.filter((item) => item.enabled).map((item) => item.extensionId),
    [allExtensions],
  )
  const groupNameMap = useMemo(() => {
    const map = new Map<string, string>()
    groups.forEach((group) => map.set(group.groupId, group.groupName))
    return map
  }, [groups])

  const profileGroups = useMemo(() => {
    const buckets = new Map<string, BrowserProfile[]>()
    profiles.forEach((profile) => {
      const groupId = (profile.groupId || '').trim() || UNGROUPED_PROFILE_GROUP_ID
      if (!buckets.has(groupId)) buckets.set(groupId, [])
      buckets.get(groupId)!.push(profile)
    })

    const sections = groups
      .filter((group) => buckets.has(group.groupId))
      .sort((a, b) => a.sortOrder - b.sortOrder || a.groupName.localeCompare(b.groupName))
      .map((group) => ({
        groupId: group.groupId,
        groupName: group.groupName,
        profiles: buckets.get(group.groupId) || [],
      }))

    for (const [groupId, items] of buckets.entries()) {
      if (groupId === UNGROUPED_PROFILE_GROUP_ID) continue
      if (!sections.some((section) => section.groupId === groupId)) {
        sections.push({
          groupId,
          groupName: groupNameMap.get(groupId) || `分组 ${groupId}`,
          profiles: items,
        })
      }
    }

    if (buckets.has(UNGROUPED_PROFILE_GROUP_ID)) {
      sections.push({
        groupId: UNGROUPED_PROFILE_GROUP_ID,
        groupName: '未分组',
        profiles: buckets.get(UNGROUPED_PROFILE_GROUP_ID) || [],
      })
    }

    return sections
  }, [profiles, groups, groupNameMap])

  useEffect(() => {
    if (!open || !extension) return
    setLoading(true)
    Promise.all([fetchBrowserProfiles(), fetchGroups()]).then(async ([profileItems, groupItems]) => {
      const profileSettings = await Promise.all(profileItems.map(async (profile) => ({
        profile,
        settings: await fetchBrowserProfileExtensionSettings(profile.profileId),
      })))
      const settingsMap: Record<string, BrowserProfileExtensionSettings> = {}
      profileSettings.forEach(({ profile, settings }) => {
        settingsMap[profile.profileId] = settings
      })
      setProfiles(profileItems)
      setGroups(groupItems)
      setSettingsByProfile(settingsMap)
      setSelectedIds(profileItems
        .filter((profile) => {
          const settings = settingsMap[profile.profileId]
          return settings?.configured ? settings.extensionIds.includes(extension.extensionId) : extension.enabled
        })
        .map((profile) => profile.profileId))
    }).catch((error: any) => {
      toast.error(error?.message || '加载实例限制失败')
    }).finally(() => setLoading(false))
  }, [open, extension])

  const toggleProfile = (profileId: string, checked: boolean) => {
    setSelectedIds((current) => {
      if (checked) return current.includes(profileId) ? current : [...current, profileId]
      return current.filter((item) => item !== profileId)
    })
  }

  const toggleGroup = (profileIds: string[], checked: boolean) => {
    setSelectedIds((current) => {
      const next = new Set(current)
      profileIds.forEach((profileId) => {
        if (checked) next.add(profileId)
        else next.delete(profileId)
      })
      return Array.from(next)
    })
  }

  const handleSave = async () => {
    if (!extension) return
    setSaving(true)
    try {
      const selected = new Set(selectedIds)
      const saveTasks = profiles.map((profile) => {
        const current = settingsByProfile[profile.profileId]
        const baseIds = current?.configured ? current.extensionIds : enabledExtensionIds
        const nextIds = selected.has(profile.profileId)
          ? Array.from(new Set([...baseIds, extension.extensionId]))
          : baseIds.filter((extensionId) => extensionId !== extension.extensionId)

        if (!current?.configured && sameStringSet(baseIds, nextIds)) return null
        if (current?.configured && sameStringSet(current.extensionIds, nextIds)) return null
        return saveBrowserProfileExtensionSettings(profile.profileId, nextIds, true)
      }).filter((task): task is Promise<BrowserProfileExtensionSettings> => task !== null)

      if (saveTasks.length > 0) await Promise.all(saveTasks)
      toast.success('实例限制已保存')
      onClose()
    } catch (error: any) {
      toast.error(error?.message || '保存实例限制失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={extension ? `限制实例：${extension.name || extension.extensionId}` : '限制实例'}
      width="680px"
      footer={(
        <>
          <Button variant="secondary" onClick={onClose}>取消</Button>
          <Button onClick={handleSave} loading={saving} disabled={loading || !extension}>保存</Button>
        </>
      )}
    >
      {loading ? (
        <div className="py-8 text-center text-sm text-[var(--color-text-muted)]">正在加载实例...</div>
      ) : (
        <div className="space-y-3">
          <div className="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-muted)] px-3 py-2 text-sm text-[var(--color-text-secondary)]">
            勾选的实例会加载此插件；未勾选的实例会排除此插件。
          </div>
          <div className="max-h-[420px] space-y-3 overflow-auto pr-1">
            {profileGroups.map((group) => {
              const groupProfileIds = group.profiles.map((profile) => profile.profileId)
              const selectedCount = groupProfileIds.filter((profileId) => selectedSet.has(profileId)).length
              const allSelected = groupProfileIds.length > 0 && selectedCount === groupProfileIds.length

              return (
                <section key={group.groupId} className="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)]">
                  <div className="flex items-center justify-between gap-3 border-b border-[var(--color-border-muted)] px-3 py-2">
                    <div className="min-w-0">
                      <div className="truncate text-sm font-medium text-[var(--color-text-primary)]">{group.groupName}</div>
                      <div className="text-xs text-[var(--color-text-muted)]">已选 {selectedCount} / {groupProfileIds.length}</div>
                    </div>
                    <Button size="sm" variant="ghost" onClick={() => toggleGroup(groupProfileIds, !allSelected)}>
                      {allSelected ? '取消本组' : '选择本组'}
                    </Button>
                  </div>
                  <div className="divide-y divide-[var(--color-border-muted)]">
                    {group.profiles.map((profile) => (
                      <label key={profile.profileId} className="flex items-start gap-3 px-3 py-2">
                        <input
                          type="checkbox"
                          checked={selectedSet.has(profile.profileId)}
                          onChange={(event) => toggleProfile(profile.profileId, event.target.checked)}
                          className="mt-1 h-4 w-4 shrink-0 rounded accent-[var(--color-accent)]"
                        />
                        <div className="min-w-0 flex-1">
                          <div className="flex flex-wrap items-center gap-2 text-sm font-medium text-[var(--color-text-primary)]">
                            <span>{profile.profileName || profile.profileId}</span>
                            {profile.running ? <span className="rounded bg-green-50 px-1.5 py-0.5 text-xs text-green-700">运行中</span> : null}
                            {settingsByProfile[profile.profileId]?.configured ? <span className="rounded bg-[var(--color-bg-muted)] px-1.5 py-0.5 text-xs font-normal text-[var(--color-text-muted)]">已单独配置</span> : null}
                          </div>
                          <div className="mt-1 break-all font-mono text-xs text-[var(--color-text-muted)]">{profile.profileId}</div>
                        </div>
                      </label>
                    ))}
                  </div>
                </section>
              )
            })}

            {profiles.length === 0 ? (
              <div className="rounded-xl border border-dashed border-[var(--color-border-default)] bg-[var(--color-bg-muted)] px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">
                暂无实例
              </div>
            ) : null}
          </div>
        </div>
      )}
    </Modal>
  )
}

export interface ExtensionHistoryModalProps {
  open: boolean
  records: ExtensionHistoryRecord[]
  onClose: () => void
  onPick: (record: ExtensionHistoryRecord) => void
  onClear: () => void
}

export function ExtensionHistoryModal({ open, records, onClose, onPick, onClear }: ExtensionHistoryModalProps) {
  return (
    <Modal
      open={open}
      onClose={onClose}
      title="插件历史"
      width="760px"
      footer={(
        <>
          <Button variant="secondary" onClick={onClear} disabled={records.length === 0}>清空历史</Button>
          <Button onClick={onClose}>关闭</Button>
        </>
      )}
    >
      <div className="max-h-[520px] overflow-y-auto">
        {records.length === 0 ? (
          <div className="rounded-xl border border-dashed border-[var(--color-border-default)] bg-[var(--color-bg-muted)] px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">
            暂无历史记录
          </div>
        ) : (
          <div className="space-y-2">
            {records.map((record) => (
              <div key={record.id} className="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-3">
                <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="rounded-full bg-[var(--color-bg-muted)] px-2 py-0.5 text-xs text-[var(--color-text-muted)]">{extensionHistoryActionLabel(record.action)}</span>
                      <span className={record.ok ? 'text-xs text-green-600' : 'text-xs text-red-500'}>{record.ok ? '成功' : '失败'}</span>
                      <span className="text-xs text-[var(--color-text-muted)]">{formatExtensionTime(record.createdAt)}</span>
                    </div>
                    <div className="mt-1 truncate text-sm font-medium text-[var(--color-text-primary)]">{record.name || record.extensionId || record.query}</div>
                    {record.extensionId ? <div className="mt-1 break-all font-mono text-xs text-[var(--color-text-muted)]">{record.extensionId}</div> : null}
                    {record.proxyLabel ? <div className="mt-1 text-xs text-[var(--color-text-muted)]">{record.proxyLabel}</div> : null}
                    {record.message ? <div className="mt-1 line-clamp-2 text-xs text-[var(--color-text-muted)]">{record.message}</div> : null}
                  </div>
                  <div className="flex shrink-0 gap-2">
                    <Button type="button" size="sm" variant="secondary" onClick={() => onPick(record)}>
                      <Search className="h-4 w-4" />
                      使用
                    </Button>
                    {record.storeUrl ? (
                      <Button type="button" size="sm" variant="secondary" onClick={() => window.open(record.storeUrl, '_blank', 'noopener,noreferrer')}>
                        <ExternalLink className="h-4 w-4" />
                        商店页
                      </Button>
                    ) : null}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </Modal>
  )
}
