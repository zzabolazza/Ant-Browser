import { useEffect, useRef, useState } from 'react'
import { RefreshCw, Trash2 } from 'lucide-react'
import { Badge, Button, Card } from '../../../shared/components'

interface LogEntry {
  time: string
  level: string
  component: string
  message: string
  fields?: Record<string, any>
}

const LEVELS = ['ALL', 'DEBUG', 'INFO', 'WARN', 'ERROR'] as const

const levelVariant = (level: string) => {
  switch (level) {
    case 'ERROR': return 'error'
    case 'WARN': return 'warning'
    case 'DEBUG': return 'default'
    default: return 'info'
  }
}

const levelColor = (level: string) => {
  switch (level) {
    case 'ERROR': return 'text-red-500'
    case 'WARN': return 'text-yellow-500'
    case 'DEBUG': return 'text-[var(--color-text-muted)]'
    default: return 'text-[var(--color-text-secondary)]'
  }
}

const selectClassName = 'px-3 py-1.5 text-[12.5px] rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-muted)] text-[var(--color-text-primary)] focus:outline-none focus:ring-1 focus:ring-[var(--color-accent)]'

async function fetchLogs(): Promise<LogEntry[]> {
  try {
    const bindings: any = await import('../../../wailsjs/go/main/App')
    return (await bindings.GetAppLogs()) || []
  } catch { return [] }
}

async function clearLogs() {
  try {
    const bindings: any = await import('../../../wailsjs/go/main/App')
    await bindings.ClearAppLogs()
  } catch { /* ignore */ }
}

export function BrowserLogsPage() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [levelFilter, setLevelFilter] = useState('ALL')
  const [componentFilter, setComponentFilter] = useState('ALL')
  const [timeFrom, setTimeFrom] = useState('')
  const [timeTo, setTimeTo] = useState('')
  const [autoScroll, setAutoScroll] = useState(false)
  const [loading, setLoading] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  const load = async () => {
    setLoading(true)
    try {
      const data = await fetchLogs()
      setLogs(data)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    const timer = setInterval(load, 3000)
    return () => clearInterval(timer)
  }, [])

  useEffect(() => {
    if (autoScroll) {
      bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
  }, [logs, autoScroll])

  const handleClear = async () => {
    await clearLogs()
    setLogs([])
  }

  const filtered = logs.filter(entry => {
    if (levelFilter !== 'ALL' && entry.level !== levelFilter) return false
    if (componentFilter !== 'ALL' && entry.component !== componentFilter) return false
    if (timeFrom && entry.time < timeFrom.replace('T', ' ')) return false
    if (timeTo && entry.time > timeTo.replace('T', ' ')) return false
    return true
  })
  const components = Array.from(new Set(logs.map(entry => entry.component).filter(Boolean))).sort()
  const resetFilters = () => {
    setLevelFilter('ALL')
    setComponentFilter('ALL')
    setTimeFrom('')
    setTimeTo('')
  }

  return (
    <div className="space-y-4 animate-fade-in">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="max-w-2xl text-[12.5px] leading-5 text-[var(--color-text-muted)]">
          本机应用运行日志，支持实时刷新、级别 / 组件 / 时间过滤。
        </p>
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="secondary" size="sm" onClick={handleClear}>
            <Trash2 className="w-4 h-4" />清空日志
          </Button>
          <Button size="sm" onClick={load} loading={loading}>
            <RefreshCw className="w-4 h-4" />刷新
          </Button>
        </div>
      </div>

      <div className="rounded-[10px] border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-3">
        <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
          <select
            value={levelFilter}
            onChange={e => setLevelFilter(e.target.value)}
            className={selectClassName}
          >
            {LEVELS.map(level => (
              <option key={level} value={level}>{level === 'ALL' ? '全部等级' : level}</option>
            ))}
          </select>
          <select
            value={componentFilter}
            onChange={e => setComponentFilter(e.target.value)}
            className={selectClassName}
          >
            <option value="ALL">全部组件</option>
            {components.map(component => (
              <option key={component} value={component}>{component}</option>
            ))}
          </select>
          <div className="flex flex-wrap items-center gap-x-2 gap-y-2">
            <input
              type="datetime-local"
              value={timeFrom}
              onChange={e => setTimeFrom(e.target.value)}
              className={selectClassName}
            />
            <span className="text-xs text-[var(--color-text-muted)]">到</span>
            <input
              type="datetime-local"
              value={timeTo}
              onChange={e => setTimeTo(e.target.value)}
              className={selectClassName}
            />
          </div>
          <label className="ml-auto flex items-center gap-1.5 text-[11.5px] text-[var(--color-text-muted)] cursor-pointer select-none">
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={e => setAutoScroll(e.target.checked)}
              className="w-3.5 h-3.5"
            />
            自动滚动
          </label>
          <Button variant="secondary" size="sm" onClick={resetFilters}>重置筛选</Button>
          <span className="font-mono text-[11px] text-[var(--color-text-muted)]">
            {filtered.length} / {logs.length}
          </span>
        </div>
      </div>

      <Card padding="none">
        <div
          className="overflow-auto font-mono text-xs"
          style={{ maxHeight: 'calc(100vh - 320px)' }}
        >
          {filtered.length === 0 ? (
            <div className="py-16 text-center text-sm text-[var(--color-text-muted)]">暂无日志</div>
          ) : (
            <table className="min-w-full">
              <thead className="sticky top-0 z-10 bg-[var(--color-bg-subtle)]">
                <tr>
                  <th className="px-3.5 py-2.5 text-left text-[11px] font-bold uppercase tracking-[0.04em] text-[var(--color-text-muted)] border-b border-[var(--color-border-default)] w-40">时间</th>
                  <th className="px-3.5 py-2.5 text-left text-[11px] font-bold uppercase tracking-[0.04em] text-[var(--color-text-muted)] border-b border-[var(--color-border-default)] w-16">级别</th>
                  <th className="px-3.5 py-2.5 text-left text-[11px] font-bold uppercase tracking-[0.04em] text-[var(--color-text-muted)] border-b border-[var(--color-border-default)] w-28">组件</th>
                  <th className="px-3.5 py-2.5 text-left text-[11px] font-bold uppercase tracking-[0.04em] text-[var(--color-text-muted)] border-b border-[var(--color-border-default)]">消息</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--color-border-muted)]">
                {filtered.map((entry, i) => (
                  <tr key={i} className="hover:bg-[var(--color-bg-subtle)]">
                    <td className="px-3.5 py-1.5 text-[var(--color-text-muted)] whitespace-nowrap">{entry.time}</td>
                    <td className="px-3.5 py-1.5">
                      <Badge variant={levelVariant(entry.level)} className="text-[10px]">{entry.level}</Badge>
                    </td>
                    <td className="px-3.5 py-1.5 text-[var(--color-text-muted)] truncate max-w-[112px]" title={entry.component}>
                      {entry.component}
                    </td>
                    <td className={`px-3.5 py-1.5 ${levelColor(entry.level)}`}>
                      <span>{entry.message}</span>
                      {entry.fields && Object.keys(entry.fields).length > 0 && (
                        <span className="ml-2 text-[var(--color-text-muted)]">
                          {Object.entries(entry.fields).map(([k, v]) => `${k}=${JSON.stringify(v)}`).join(' ')}
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <div ref={bottomRef} />
        </div>
      </Card>
    </div>
  )
}
