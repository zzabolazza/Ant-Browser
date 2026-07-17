import clsx from 'clsx'

type ProgressStatus = 'normal' | 'success' | 'error' | 'warning'

interface ProgressProps {
  percent: number
  status?: ProgressStatus
  showInfo?: boolean
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

const statusColors = {
  normal: 'bg-[var(--color-accent)]',
  success: 'bg-[var(--color-success)]',
  error: 'bg-[var(--color-error)]',
  warning: 'bg-[var(--color-warning)]',
}

const sizeStyles = {
  sm: 'h-1',
  md: 'h-2',
  lg: 'h-3',
}

export function Progress({
  percent,
  status = 'normal',
  showInfo = true,
  size = 'md',
  className,
}: ProgressProps) {
  const validPercent = Math.min(100, Math.max(0, percent))

  return (
    <div className={clsx('flex items-center gap-3', className)}>
      <div className={clsx('flex-1 bg-[var(--color-bg-muted)] rounded-full overflow-hidden', sizeStyles[size])}>
        <div
          className={clsx('h-full transition-all duration-300 rounded-full', statusColors[status])}
          style={{ width: `${validPercent}%` }}
        />
      </div>
      {showInfo && (
        <span className="text-sm text-[var(--color-text-muted)] min-w-[3ch] text-right">
          {validPercent}%
        </span>
      )}
    </div>
  )
}
