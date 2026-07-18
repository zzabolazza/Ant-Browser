import { Check } from 'lucide-react'
import clsx from 'clsx'

import { Card } from '../../../shared/components'
import type { ThemeMode } from '../../../shared/theme/theme'

interface ThemeSettingsCardProps {
  value: ThemeMode
  onChange: (value: ThemeMode) => void
}

const THEME_OPTIONS = [
  { value: 'dark', label: '深色', description: '沉稳专业的深色风格' },
  { value: 'light', label: '浅色', description: '简洁明亮的浅色风格' },
  { value: 'cream', label: '奶油', description: '温暖柔和的奶油色调' },
  { value: 'mint', label: '薄荷', description: '清新自然的浅绿风格' },
  { value: 'ocean', label: '海洋', description: '深邃宁静的蓝色风格' },
] as const

const THEME_PREVIEWS: Record<ThemeMode, { background: string; surface: string; accent: string }> = {
  dark: { background: '#0c0c0e', surface: '#18181b', accent: '#fafafa' },
  light: { background: '#f5f6fa', surface: '#ffffff', accent: '#4b6eff' },
  cream: { background: '#faf7f2', surface: '#fffdf8', accent: '#8b7355' },
  mint: { background: '#f6f9f8', surface: '#fbfdfc', accent: '#3d5a4c' },
  ocean: { background: '#f5f8fa', surface: '#fafcfd', accent: '#3a5068' },
}

export function ThemeSettingsCard({ value, onChange }: ThemeSettingsCardProps) {
  return (
    <Card title="主题" subtitle="选择界面配色，主题偏好会保存在本机" padding="md">
      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
          {THEME_OPTIONS.map((option) => {
            const active = value === option.value
            const preview = THEME_PREVIEWS[option.value]
            return (
              <button
                key={option.value}
                type="button"
                onClick={() => onChange(option.value)}
                aria-pressed={active}
                title={option.description}
                className={clsx(
                  'group relative flex flex-col items-center gap-2.5 rounded-[10px] border-2 p-3 transition-all duration-150',
                  active
                    ? 'border-[var(--color-accent)] bg-[var(--color-accent-muted)]'
                    : 'border-[var(--color-border-default)] bg-[var(--color-bg-surface)] hover:border-[var(--color-border-strong)]',
                )}
              >
                <div
                  className="aspect-[4/3] w-full overflow-hidden rounded-lg border border-black/10"
                  style={{ backgroundColor: preview.background }}
                >
                  <div className="h-full w-1/4 float-left" style={{ backgroundColor: preview.surface }}>
                    <div className="mx-auto mt-2 h-1 w-2/3 rounded-full" style={{ backgroundColor: preview.accent }} />
                    <div className="mx-1 mt-2 space-y-1">
                      <div className="h-0.5 rounded-full bg-black/10" />
                      <div className="h-0.5 rounded-full bg-black/10" />
                    </div>
                  </div>
                  <div className="p-1">
                    <div className="mb-1 h-1 w-1/2 rounded-full bg-black/10" />
                    <div className="grid grid-cols-2 gap-0.5">
                      <div className="h-2 rounded-sm" style={{ backgroundColor: preview.surface }} />
                      <div className="h-2 rounded-sm" style={{ backgroundColor: preview.surface }} />
                    </div>
                  </div>
                </div>

                <span className={clsx(
                  'text-xs font-medium transition-colors',
                  active ? 'text-[var(--color-text-primary)]' : 'text-[var(--color-text-secondary)]',
                )}>
                  {option.label}
                </span>

                {active && (
                  <span className="absolute -right-1.5 -top-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-[var(--color-accent)] shadow-sm">
                    <Check className="h-3 w-3 text-[var(--color-text-inverse)]" />
                  </span>
                )}
              </button>
            )
          })}
        </div>
        <p className="text-center text-xs text-[var(--color-text-muted)]">
          {THEME_OPTIONS.find(option => option.value === value)?.description}
        </p>
      </div>
    </Card>
  )
}
