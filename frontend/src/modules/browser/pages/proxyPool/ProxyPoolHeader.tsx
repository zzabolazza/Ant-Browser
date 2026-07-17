import { Plus } from 'lucide-react'
import { Button } from '../../../../shared/components'

interface ProxyPoolHeaderProps {
  onOpenImport: () => void
  onOpenSettings: () => void
  totalCount: number
  availableCount?: number
  avgLatencyMs?: number | null
}

export function ProxyPoolHeader({
  onOpenImport,
  onOpenSettings,
  totalCount,
  availableCount,
  avgLatencyMs = null,
}: ProxyPoolHeaderProps) {
  const stats = [
    { label: '代理地址总数', value: String(totalCount), tone: 'default' as const },
    {
      label: '可用地址',
      value: availableCount === undefined ? '-' : String(availableCount),
      tone: 'success' as const,
    },
    {
      label: '平均延迟',
      value: avgLatencyMs === null ? '-' : `${avgLatencyMs} ms`,
      tone: 'default' as const,
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="max-w-2xl text-[12.5px] leading-5 text-[var(--color-text-muted)]">
          管理代理地址：协议、主机、端口与认证信息，支持批量导入与健康检测。
        </p>
        <div className="flex flex-wrap justify-end gap-2">
          <Button size="sm" variant="secondary" onClick={onOpenSettings}>
            检测设置
          </Button>
          <Button size="sm" onClick={onOpenImport}>
            <Plus className="h-4 w-4" />
            新建代理
          </Button>
        </div>
      </div>

      <div className="grid gap-3 sm:grid-cols-3">
        {stats.map((item) => (
          <div
            key={item.label}
            className="rounded-[10px] border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-4 py-3"
          >
            <div className="text-[11.5px] font-semibold text-[var(--color-text-muted)]">
              {item.label}
            </div>
            <div className={`mt-1.5 text-2xl font-extrabold tracking-tight ${item.tone === 'success' ? 'text-[var(--color-success)]' : 'text-[var(--color-text-primary)]'}`}>
              {item.value}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
