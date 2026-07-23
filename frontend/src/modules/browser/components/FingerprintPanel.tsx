import { useEffect, useState } from 'react'
import { ChevronDown, ChevronUp, RefreshCw, Wand2 } from 'lucide-react'
import { ConfirmModal, FormItem, Input, Select, Textarea } from '../../../shared/components'
import {
  type DisableSpoofingValue,
  type FingerprintConfig,
  DISABLE_SPOOFING_VALUES,
  FINGERPRINT_PRESETS,
  PRESET_RESOLUTIONS,
  deserialize,
  randomFingerprintSeed,
  serialize,
} from '../utils/fingerprintSerializer'

interface FingerprintPanelProps {
  value: string[]
  onChange: (args: string[]) => void
}

const BRAND_OPTIONS = [
  { value: '', label: '不设置' },
  { value: 'Chrome', label: 'Chrome' },
  { value: 'Edge', label: 'Edge' },
  { value: 'Opera', label: 'Opera' },
  { value: 'Vivaldi', label: 'Vivaldi' },
]

const PLATFORM_OPTIONS = [
  { value: '', label: '不设置' },
  { value: 'windows', label: 'Windows' },
  { value: 'macos', label: 'macOS' },
  { value: 'linux', label: 'Linux' },
]

const RESOLUTION_OPTIONS = [
  { value: '', label: '不设置' },
  ...PRESET_RESOLUTIONS.map(r => ({ value: r, label: r })),
  { value: 'custom', label: '自定义...' },
]

const HARDWARE_CONCURRENCY_OPTIONS = [
  { value: '', label: '不设置（由种子推导）' },
  { value: '2', label: '2 核' },
  { value: '4', label: '4 核' },
  { value: '6', label: '6 核' },
  { value: '8', label: '8 核' },
  { value: '10', label: '10 核' },
  { value: '12', label: '12 核' },
  { value: '16', label: '16 核' },
]

const WEBRTC_OPTIONS = [
  { value: '', label: '不设置' },
  { value: 'disable_non_proxied_udp', label: '禁用非代理 UDP（推荐）' },
  { value: 'default_public_interface_only', label: '仅公网接口' },
  { value: 'default_public_and_private_interfaces', label: '公网+私网接口' },
]

const DISABLE_SPOOFING_LABELS: Record<DisableSpoofingValue, string> = {
  font: '字体',
  audio: 'Audio',
  canvas: 'Canvas',
  clientrects: 'ClientRects',
  gpu: 'GPU',
}

const PRESET_OPTIONS = [
  { value: '', label: '选择预设...' },
  ...FINGERPRINT_PRESETS.map(p => ({ value: p.id, label: p.name })),
]

export function FingerprintPanel({ value, onChange }: FingerprintPanelProps) {
  const [config, setConfig] = useState<FingerprintConfig>(() => deserialize(value))
  const [advancedOpen, setAdvancedOpen] = useState(false)
  const [confirmSeedOpen, setConfirmSeedOpen] = useState(false)

  useEffect(() => {
    setConfig(deserialize(value))
  }, [value.join('\n')])

  const update = (patch: Partial<FingerprintConfig>) => {
    const next = {
      ...config,
      ...patch,
      lang: config.lang,
      timezone: config.timezone,
    }
    setConfig(next)
    onChange(serialize(next))
  }

  const toggleDisableSpoofing = (item: DisableSpoofingValue) => {
    const current = config.disableSpoofing ?? []
    const next = current.includes(item)
      ? current.filter(v => v !== item)
      : [...current, item]
    update({ disableSpoofing: next.length > 0 ? next : undefined })
  }

  const handlePresetChange = (presetId: string) => {
    if (!presetId) return
    const preset = FINGERPRINT_PRESETS.find(p => p.id === presetId)
    if (!preset) return
    const next: FingerprintConfig = {
      ...preset.config,
      seed: randomFingerprintSeed(),
      unknownArgs: config.unknownArgs,
      lang: config.lang,
      timezone: config.timezone,
    }
    setConfig(next)
    onChange(serialize(next))
  }

  const handleAdvancedChange = (text: string) => {
    const args = text.split('\n').map(s => s.trim()).filter(Boolean)
    const parsed = deserialize(args)
    parsed.lang = config.lang
    parsed.timezone = config.timezone
    setConfig(parsed)
    onChange(serialize(parsed))
  }

  const advancedText = serialize(config).join('\n')
  const disabledSpoofing = new Set(config.disableSpoofing ?? [])

  return (
    <div className="space-y-4">
      <div className="p-3 rounded-lg bg-[var(--color-bg-hover)] border border-[var(--color-border)] space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wide">指纹种子（Fingerprint Seed）</span>
          <span className="text-xs text-[var(--color-text-muted)]">决定噪声与硬件特征；deviceMemory 由种子在 8/16/32 中选取</span>
        </div>
        <div className="flex items-center gap-2">
          <Input
            value={config.seed ?? ''}
            onChange={e => update({ seed: e.target.value || undefined })}
            placeholder="留空则由系统按 ProfileId 自动生成"
            className="flex-1 font-mono text-sm"
          />
          <button
            type="button"
            title="随机生成新种子"
            onClick={() => {
              if (config.seed) {
                setConfirmSeedOpen(true)
              } else {
                update({ seed: randomFingerprintSeed() })
              }
            }}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs bg-[var(--color-primary)] text-white hover:opacity-90 transition-opacity shrink-0"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            随机
          </button>
        </div>
      </div>

      <ConfirmModal
        open={confirmSeedOpen}
        onClose={() => setConfirmSeedOpen(false)}
        onConfirm={() => update({ seed: randomFingerprintSeed() })}
        title="重新生成指纹种子"
        content="重新生成后，当前指纹将完全改变，Canvas、Audio、GPU、deviceMemory 等特征都会随之变化。确定继续？"
        confirmText="确定重新生成"
        danger
      />

      <div className="flex items-center gap-3 p-3 rounded-lg bg-[var(--color-bg-hover)] border border-[var(--color-border)]">
        <Wand2 className="w-4 h-4 text-[var(--color-text-muted)] shrink-0" />
        <div className="flex-1 min-w-0">
          <Select
            value=""
            onChange={e => handlePresetChange(e.target.value)}
            options={PRESET_OPTIONS}
          />
        </div>
        <span className="text-xs text-[var(--color-text-muted)] shrink-0">选择后覆盖当前配置</span>
      </div>

      <div>
        <p className="text-xs font-medium text-[var(--color-text-muted)] mb-2 uppercase tracking-wide">基础身份</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormItem label="浏览器品牌">
            <Select value={config.brand ?? ''} onChange={e => update({ brand: e.target.value || undefined })} options={BRAND_OPTIONS} />
          </FormItem>
          <FormItem label="品牌版本" hint="可选，对应 --fingerprint-brand-version">
            <Input
              value={config.brandVersion ?? ''}
              onChange={e => update({ brandVersion: e.target.value || undefined })}
              placeholder="留空则使用内核默认"
            />
          </FormItem>
          <FormItem label="操作系统">
            <Select value={config.platform ?? ''} onChange={e => update({ platform: e.target.value || undefined })} options={PLATFORM_OPTIONS} />
          </FormItem>
          <FormItem label="系统版本" hint="可选，对应 --fingerprint-platform-version">
            <Input
              value={config.platformVersion ?? ''}
              onChange={e => update({ platformVersion: e.target.value || undefined })}
              placeholder="例如 15.2.0"
            />
          </FormItem>
          <FormItem label="语言" hint="由代理出口自动匹配">
            <Input value={config.lang || '未匹配'} disabled className="opacity-80" />
          </FormItem>
          <FormItem label="时区" hint="由代理出口自动匹配">
            <Input value={config.timezone || '未匹配'} disabled className="opacity-80" />
          </FormItem>
        </div>
      </div>

      <div>
        <p className="text-xs font-medium text-[var(--color-text-muted)] mb-2 uppercase tracking-wide">屏幕与硬件</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormItem label="窗口分辨率" hint="对应 --window-size">
            <Select
              value={config.resolution ?? ''}
              onChange={e => update({ resolution: e.target.value || undefined, customResolution: undefined })}
              options={RESOLUTION_OPTIONS}
            />
          </FormItem>
          {config.resolution === 'custom' && (
            <FormItem label="自定义分辨率">
              <Input value={config.customResolution ?? ''} onChange={e => update({ customResolution: e.target.value || undefined })} placeholder="1600,900" />
            </FormItem>
          )}
          <FormItem label="CPU 核心数">
            <Select value={config.hardwareConcurrency ?? ''} onChange={e => update({ hardwareConcurrency: e.target.value || undefined })} options={HARDWARE_CONCURRENCY_OPTIONS} />
          </FormItem>
        </div>
      </div>

      <div>
        <p className="text-xs font-medium text-[var(--color-text-muted)] mb-2 uppercase tracking-wide">伪装开关</p>
        <p className="text-xs text-[var(--color-text-muted)] mb-2">开启指纹种子后以下伪装默认生效；勾选表示通过 --disable-spoofing 关闭该项</p>
        <div className="flex flex-wrap gap-2">
          {DISABLE_SPOOFING_VALUES.map(item => {
            const checked = disabledSpoofing.has(item)
            return (
              <label
                key={item}
                className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-md border text-xs cursor-pointer transition-colors ${
                  checked
                    ? 'border-[var(--color-border-strong)] bg-[var(--color-bg-hover)]'
                    : 'border-[var(--color-border)] hover:border-[var(--color-border-strong)]'
                }`}
              >
                <input
                  type="checkbox"
                  checked={checked}
                  onChange={() => toggleDisableSpoofing(item)}
                />
                关闭 {DISABLE_SPOOFING_LABELS[item]}
              </label>
            )
          })}
        </div>
      </div>

      <div>
        <p className="text-xs font-medium text-[var(--color-text-muted)] mb-2 uppercase tracking-wide">网络与隐私</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormItem label="WebRTC 策略">
            <Select value={config.webrtcPolicy ?? ''} onChange={e => update({ webrtcPolicy: e.target.value || undefined })} options={WEBRTC_OPTIONS} />
          </FormItem>
        </div>
      </div>

      <div className="border border-[var(--color-border)] rounded-lg overflow-hidden">
        <button
          type="button"
          className="w-full flex items-center justify-between px-4 py-2.5 text-sm text-[var(--color-text-muted)] hover:bg-[var(--color-bg-hover)] transition-colors"
          onClick={() => setAdvancedOpen(v => !v)}
        >
          <span>高级模式（原始参数）</span>
          {advancedOpen ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
        </button>
        {advancedOpen && (
          <div className="px-4 pb-4 pt-2 border-t border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] mb-2">每行一个参数，修改后自动同步到上方控件</p>
            <Textarea
              value={advancedText}
              onChange={e => handleAdvancedChange(e.target.value)}
              rows={6}
              placeholder="--fingerprint-brand=Chrome"
            />
          </div>
        )}
      </div>
    </div>
  )
}
