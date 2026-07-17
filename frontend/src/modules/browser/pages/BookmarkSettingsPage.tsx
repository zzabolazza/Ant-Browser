import { useEffect, useState } from 'react'
import clsx from 'clsx'
import { Bookmark, GripVertical, LayoutGrid, List, Plus, RefreshCw, RotateCcw, Trash2 } from 'lucide-react'
import { Button, Card, ConfirmModal, Input, toast } from '../../../shared/components'
import type { BrowserBookmark } from '../types'
import { fetchBookmarks, resetBookmarks, saveBookmarks, syncBookmarksToProfiles } from '../api'

type BookmarkViewMode = 'list' | 'grid'

function extractDomain(url: string): string {
  const raw = url.trim()
  if (!raw) return ''
  try {
    const parsed = new URL(raw.includes('://') ? raw : `https://${raw}`)
    return parsed.hostname.replace(/^www\./, '')
  } catch {
    return ''
  }
}

function faviconURL(url: string): string | null {
  const domain = extractDomain(url)
  return domain ? `https://favicon.im/${domain}` : null
}

function BookmarkFavicon({ url }: { url: string }) {
  const src = faviconURL(url)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    setFailed(false)
  }, [src])

  if (!src || failed) {
    return (
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-muted)]">
        <Bookmark className="h-4 w-4 text-[var(--color-text-muted)]" />
      </div>
    )
  }

  return (
    <div className="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-muted)]">
      <img
        src={src}
        alt=""
        className="h-5 w-5 object-contain"
        loading="lazy"
        referrerPolicy="no-referrer"
        onError={() => setFailed(true)}
      />
    </div>
  )
}

export function BookmarkSettingsPage() {
  const [items, setItems] = useState<BrowserBookmark[]>([])
  const [saving, setSaving] = useState(false)
  const [syncing, setSyncing] = useState(false)
  const [resetOpen, setResetOpen] = useState(false)
  const [syncOpen, setSyncOpen] = useState(false)
  const [dragIndex, setDragIndex] = useState<number | null>(null)
  const [viewMode, setViewMode] = useState<BookmarkViewMode>('list')

  useEffect(() => {
    fetchBookmarks().then(setItems)
  }, [])

  const handleChange = (index: number, field: keyof BrowserBookmark, value: string) => {
    setItems(prev => prev.map((item, i) => i === index ? { ...item, [field]: value } : item))
  }

  const handleAdd = () => {
    setItems(prev => [...prev, { name: '', url: '', openOnStart: false }])
  }

  const handleDelete = (index: number) => {
    setItems(prev => prev.filter((_, i) => i !== index))
  }

  const handleOpenOnStartChange = (index: number, checked: boolean) => {
    setItems(prev => prev.map((item, i) => i === index ? { ...item, openOnStart: checked } : item))
  }

  const handleSave = async () => {
    const valid = items.filter(i => i.name.trim() && i.url.trim())
    if (valid.length !== items.length) {
      toast.error('存在空的名称或 URL，请填写完整后保存')
      return
    }
    setSaving(true)
    try {
      await saveBookmarks(items)
      const result = await syncBookmarksToProfiles()
      const parts = ['书签已保存']
      if (result.synced > 0) parts.push(`已同步 ${result.synced} 个已有实例`)
      if (result.skipped > 0) parts.push(`跳过运行中 ${result.skipped} 个，停止后再同步`)
      if (result.failed > 0) parts.push(`失败 ${result.failed} 个`)
      const message = parts.join('，')
      if (result.failed > 0 || result.skipped > 0) {
        toast.warning(message)
      } else {
        toast.success(message)
      }
    } finally {
      setSaving(false)
    }
  }

  const handleReset = async () => {
    await resetBookmarks()
    const fresh = await fetchBookmarks()
    setItems(fresh)
    toast.success('已恢复默认书签')
  }

  const handleSync = async () => {
    setSyncing(true)
    try {
      const result = await syncBookmarksToProfiles()
      const parts = [`已同步 ${result.synced} 个实例`]
      if (result.skipped > 0) parts.push(`跳过运行中 ${result.skipped} 个，停止后再同步`)
      if (result.failed > 0) parts.push(`失败 ${result.failed} 个`)
      const message = parts.join('，')
      if (result.failed > 0 || result.skipped > 0) {
        toast.warning(message)
      } else {
        toast.success(message)
      }
      setSyncOpen(false)
    } catch (error: any) {
      toast.error(error?.message || '同步失败')
    } finally {
      setSyncing(false)
    }
  }

  const handleDragStart = (index: number) => setDragIndex(index)
  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault()
    if (dragIndex === null || dragIndex === index) return
    setItems(prev => {
      const next = [...prev]
      const [moved] = next.splice(dragIndex, 1)
      next.splice(index, 0, moved)
      return next
    })
    setDragIndex(index)
  }
  const handleDragEnd = () => setDragIndex(null)

  const itemShellClass = (index: number) => clsx(
    'rounded-[10px] border border-[var(--color-border-default)] p-2.5 transition-all duration-150',
    dragIndex === index
      ? 'bg-[var(--color-accent-muted)] ring-1 ring-[var(--color-border-strong)]'
      : 'bg-[var(--color-bg-surface)] hover:bg-[var(--color-bg-subtle)]',
  )

  const addShellClass = clsx(
    'flex w-full items-center justify-center gap-2 rounded-[10px] border border-dashed border-[var(--color-border-default)] p-2.5',
    'bg-[var(--color-bg-surface)] text-[12.5px] font-medium text-[var(--color-text-secondary)]',
    'transition-colors hover:border-[var(--color-border-strong)] hover:bg-[var(--color-bg-subtle)] hover:text-[var(--color-text-primary)]',
  )

  return (
    <div className="space-y-5 animate-fade-in">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="max-w-2xl text-[12.5px] leading-5 text-[var(--color-text-muted)]">
          维护新建实例自动携带的默认书签列表；保存后会增量同步到未运行实例。
        </p>
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="secondary" size="sm" onClick={() => setSyncOpen(true)} loading={syncing}>
            <RefreshCw className="w-4 h-4" />
            手动同步
          </Button>
          <Button variant="secondary" size="sm" onClick={() => setResetOpen(true)}>
            <RotateCcw className="w-4 h-4" />
            恢复默认
          </Button>
          <Button size="sm" onClick={handleSave} loading={saving}>保存更改</Button>
        </div>
      </div>

      <Card padding="sm">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <div className="text-[13.5px] font-bold text-[var(--color-text-primary)]">
            书签列表（{items.length}）
          </div>
          <div className="flex overflow-hidden rounded-md border border-[var(--color-border-default)]">
            <button
              type="button"
              className={clsx(
                'flex h-8 w-8 items-center justify-center text-[var(--color-text-muted)] transition-colors hover:text-[var(--color-text-primary)]',
                viewMode === 'grid' && 'bg-[var(--color-bg-muted)] text-[var(--color-accent)]',
              )}
              onClick={() => setViewMode('grid')}
              title="多列网格"
            >
              <LayoutGrid className="h-4 w-4" />
            </button>
            <button
              type="button"
              className={clsx(
                'flex h-8 w-8 items-center justify-center text-[var(--color-text-muted)] transition-colors hover:text-[var(--color-text-primary)]',
                viewMode === 'list' && 'bg-[var(--color-bg-muted)] text-[var(--color-accent)]',
              )}
              onClick={() => setViewMode('list')}
              title="纵向列表"
            >
              <List className="h-4 w-4" />
            </button>
          </div>
        </div>

        {items.length === 0 ? (
          <div className="space-y-2">
            <p className="rounded-[10px] border border-dashed border-[var(--color-border-default)] px-4 py-8 text-center text-[12.5px] text-[var(--color-text-muted)]">
              暂无书签，点击下方按钮添加
            </p>
            <button type="button" onClick={handleAdd} className={clsx(addShellClass, 'min-h-[52px]')}>
              <Plus className="h-4 w-4" />
              添加书签
            </button>
          </div>
        ) : viewMode === 'list' ? (
          <div className="space-y-2">
            {items.map((item, index) => (
              <div
                key={index}
                draggable
                onDragStart={() => handleDragStart(index)}
                onDragOver={e => handleDragOver(e, index)}
                onDragEnd={handleDragEnd}
                className={clsx(itemShellClass(index), 'flex min-h-[52px] w-full items-center gap-2')}
              >
                <GripVertical className="h-4 w-4 shrink-0 cursor-grab text-[var(--color-text-muted)]" />
                <BookmarkFavicon url={item.url} />
                <Input
                  value={item.name}
                  onChange={e => handleChange(index, 'name', e.target.value)}
                  placeholder="名称，如 Google"
                  className="w-36 shrink-0"
                />
                <Input
                  value={item.url}
                  onChange={e => handleChange(index, 'url', e.target.value)}
                  placeholder="https://..."
                  className="min-w-0 flex-1 font-mono text-[12.5px]"
                />
                <label className="flex select-none items-center gap-1.5 whitespace-nowrap px-2 text-[12px] text-[var(--color-text-secondary)]">
                  <input
                    type="checkbox"
                    checked={Boolean(item.openOnStart)}
                    onChange={e => handleOpenOnStartChange(index, e.target.checked)}
                    className="h-4 w-4 rounded border-[var(--color-border-default)] accent-[var(--color-accent)]"
                  />
                  启动打开
                </label>
                <button
                  type="button"
                  onClick={() => handleDelete(index)}
                  className="shrink-0 rounded p-1.5 text-[var(--color-text-muted)] transition-colors hover:bg-red-50 hover:text-red-500"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            ))}
            <button type="button" onClick={handleAdd} className={clsx(addShellClass, 'min-h-[52px]')}>
              <Plus className="h-4 w-4" />
              添加书签
            </button>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            {items.map((item, index) => (
              <div
                key={index}
                draggable
                onDragStart={() => handleDragStart(index)}
                onDragOver={e => handleDragOver(e, index)}
                onDragEnd={handleDragEnd}
                className={clsx(itemShellClass(index), 'flex min-h-[148px] flex-col gap-2.5')}
              >
                <div className="flex items-center gap-2">
                  <GripVertical className="h-4 w-4 shrink-0 cursor-grab text-[var(--color-text-muted)]" />
                  <BookmarkFavicon url={item.url} />
                  <Input
                    value={item.name}
                    onChange={e => handleChange(index, 'name', e.target.value)}
                    placeholder="名称，如 Google"
                    className="min-w-0 flex-1"
                  />
                  <button
                    type="button"
                    onClick={() => handleDelete(index)}
                    className="shrink-0 rounded p-1.5 text-[var(--color-text-muted)] transition-colors hover:bg-red-50 hover:text-red-500"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
                <Input
                  value={item.url}
                  onChange={e => handleChange(index, 'url', e.target.value)}
                  placeholder="https://..."
                  className="font-mono text-[12.5px]"
                />
                <label className="mt-auto flex select-none items-center gap-1.5 text-[12px] text-[var(--color-text-secondary)]">
                  <input
                    type="checkbox"
                    checked={Boolean(item.openOnStart)}
                    onChange={e => handleOpenOnStartChange(index, e.target.checked)}
                    className="h-4 w-4 rounded border-[var(--color-border-default)] accent-[var(--color-accent)]"
                  />
                  启动打开
                </label>
              </div>
            ))}
            <button type="button" onClick={handleAdd} className={clsx(addShellClass, 'min-h-[148px] flex-col')}>
              <Plus className="h-5 w-5" />
              添加书签
            </button>
          </div>
        )}
      </Card>

      <ConfirmModal
        open={resetOpen}
        onClose={() => setResetOpen(false)}
        onConfirm={handleReset}
        title="恢复默认书签"
        content="将清除当前所有自定义书签，恢复为内置默认列表。确定继续？"
        confirmText="确定恢复"
        danger
      />

      <ConfirmModal
        open={syncOpen}
        onClose={() => setSyncOpen(false)}
        onConfirm={handleSync}
        title="手动同步已有实例"
        content="只会增量追加缺失的默认书签，不会删除、改名或移动用户已有书签。运行中的实例会跳过。"
        confirmText="开始同步"
      />
    </div>
  )
}
