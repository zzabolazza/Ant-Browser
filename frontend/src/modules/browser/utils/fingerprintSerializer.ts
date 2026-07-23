// 指纹参数序列化/反序列化工具（对齐 fingerprint-chromium 148+）

/**
 * 获取系统当前时区
 * @returns IANA 时区标识符，如 "Asia/Shanghai"
 */
export function getSystemTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone
  } catch {
    return 'UTC'
  }
}

export function getSystemLanguage(): string {
  try {
    const lang = (typeof navigator !== 'undefined' && navigator.language) ? navigator.language.trim() : ''
    return lang || 'zh-CN'
  } catch {
    return 'zh-CN'
  }
}

/** --disable-spoofing 允许的取值（Chrome 144+ / 148+） */
export const DISABLE_SPOOFING_VALUES = ['font', 'audio', 'canvas', 'clientrects', 'gpu'] as const
export type DisableSpoofingValue = (typeof DISABLE_SPOOFING_VALUES)[number]

export interface FingerprintConfig {
  // 指纹种子（核心）
  seed?: string            // --fingerprint=<seed>  控制所有随机噪声的根种子

  // 基础身份
  brand?: string            // --fingerprint-brand=
  brandVersion?: string     // --fingerprint-brand-version=
  platform?: string         // --fingerprint-platform=  windows|linux|macos
  platformVersion?: string  // --fingerprint-platform-version=
  lang?: string             // --lang=
  timezone?: string         // --timezone=

  // 屏幕与窗口
  resolution?: string       // --window-size=（预设值或 'custom'）
  customResolution?: string // 当 resolution === 'custom' 时使用

  // 硬件信息
  hardwareConcurrency?: string  // --fingerprint-hardware-concurrency=

  // 选择性关闭伪装（随 --fingerprint 默认开启的子集）
  disableSpoofing?: DisableSpoofingValue[]  // --disable-spoofing=

  // 网络与隐私
  webrtcPolicy?: string     // --webrtc-ip-handling-policy=

  unknownArgs?: string[]    // 无法识别的原始参数，原样保留
}

export const PRESET_RESOLUTIONS = ['1920,1080', '1440,900', '1366,768', '2560,1440', '1280,800', '1600,900']

// CLI 参数前缀 → FingerprintConfig 字段映射
export const KEY_MAP: Record<string, keyof FingerprintConfig> = {
  '--fingerprint': 'seed',
  '--fingerprint-brand': 'brand',
  '--fingerprint-brand-version': 'brandVersion',
  '--fingerprint-platform': 'platform',
  '--fingerprint-platform-version': 'platformVersion',
  '--lang': 'lang',
  '--timezone': 'timezone',
  '--window-size': 'resolution',
  '--fingerprint-hardware-concurrency': 'hardwareConcurrency',
  '--webrtc-ip-handling-policy': 'webrtcPolicy',
  '--disable-spoofing': 'disableSpoofing',
}

function parseDisableSpoofing(value: string): DisableSpoofingValue[] {
  const allowed = new Set<string>(DISABLE_SPOOFING_VALUES)
  const out: DisableSpoofingValue[] = []
  const seen = new Set<string>()
  for (const part of value.split(',')) {
    const item = part.trim().toLowerCase()
    if (!item || !allowed.has(item) || seen.has(item)) continue
    seen.add(item)
    out.push(item as DisableSpoofingValue)
  }
  return out
}

// FingerprintConfig → string[]
export function serialize(config: FingerprintConfig): string[] {
  const args: string[] = []
  if (config.seed) args.push(`--fingerprint=${config.seed}`)
  if (config.brand) args.push(`--fingerprint-brand=${config.brand}`)
  if (config.brandVersion) args.push(`--fingerprint-brand-version=${config.brandVersion}`)
  if (config.platform) args.push(`--fingerprint-platform=${config.platform}`)
  if (config.platformVersion) args.push(`--fingerprint-platform-version=${config.platformVersion}`)
  if (config.lang) args.push(`--lang=${config.lang}`)
  if (config.timezone) {
    const tz = config.timezone === 'system' ? getSystemTimezone() : config.timezone
    args.push(`--timezone=${tz}`)
  }

  const res = config.resolution === 'custom' ? config.customResolution : config.resolution
  if (res) args.push(`--window-size=${res}`)

  if (config.hardwareConcurrency) {
    args.push(`--fingerprint-hardware-concurrency=${config.hardwareConcurrency}`)
  }

  if (config.disableSpoofing && config.disableSpoofing.length > 0) {
    args.push(`--disable-spoofing=${config.disableSpoofing.join(',')}`)
  }

  if (config.webrtcPolicy) args.push(`--webrtc-ip-handling-policy=${config.webrtcPolicy}`)

  return [...args, ...(config.unknownArgs ?? [])]
}

// string[] → FingerprintConfig
export function deserialize(args: string[]): FingerprintConfig {
  const config: FingerprintConfig = { unknownArgs: [] }

  for (const arg of args) {
    const eqIdx = arg.indexOf('=')
    if (eqIdx === -1) {
      config.unknownArgs!.push(arg)
      continue
    }
    const key = arg.slice(0, eqIdx)
    const val = arg.slice(eqIdx + 1)
    const field = KEY_MAP[key]

    if (!field) {
      if (key.startsWith('--fingerprint-') || key === '--fingerprint') {
        continue
      }
      config.unknownArgs!.push(arg)
      continue
    }

    if (field === 'disableSpoofing') {
      config.disableSpoofing = parseDisableSpoofing(val)
    } else if (field === 'resolution') {
      if (PRESET_RESOLUTIONS.includes(val)) {
        config.resolution = val
      } else {
        config.resolution = 'custom'
        config.customResolution = val
      }
    } else {
      (config as Record<string, unknown>)[field] = val
    }
  }

  return config
}

// 生成随机指纹种子（32位正整数）
export function randomFingerprintSeed(): string {
  return String(Math.floor(Math.random() * 2147483647) + 1)
}

// ─── 预设指纹配置 ────────────────────────────────────────────────────────────

export interface FingerprintPreset {
  id: string
  name: string
  description: string
  config: Partial<FingerprintConfig>
}

export const FINGERPRINT_PRESETS: FingerprintPreset[] = [
  {
    id: 'win-chrome-office',
    name: 'Windows / Chrome / 办公',
    description: '模拟国内办公室 Windows 用户，中文环境，1920x1080',
    config: {
      brand: 'Chrome',
      platform: 'windows',
      lang: 'zh-CN',
      timezone: 'Asia/Shanghai',
      resolution: '1920,1080',
      hardwareConcurrency: '8',
      webrtcPolicy: 'disable_non_proxied_udp',
    },
  },
  {
    id: 'win-chrome-gaming',
    name: 'Windows / Chrome / 游戏主机',
    description: '模拟高配游戏 PC，2560x1440',
    config: {
      brand: 'Chrome',
      platform: 'windows',
      lang: 'en-US',
      timezone: 'America/New_York',
      resolution: '2560,1440',
      hardwareConcurrency: '16',
      webrtcPolicy: 'disable_non_proxied_udp',
    },
  },
  {
    id: 'mac-chrome-designer',
    name: 'macOS / Chrome / 设计师',
    description: '模拟 Mac 设计师用户，Retina 分辨率',
    config: {
      brand: 'Chrome',
      platform: 'macos',
      lang: 'zh-CN',
      timezone: 'Asia/Shanghai',
      resolution: '2560,1440',
      hardwareConcurrency: '10',
      webrtcPolicy: 'disable_non_proxied_udp',
    },
  },
  {
    id: 'win-edge-enterprise',
    name: 'Windows / Edge / 企业',
    description: '模拟企业 Windows 用户，Edge 浏览器，标准配置',
    config: {
      brand: 'Edge',
      platform: 'windows',
      lang: 'zh-CN',
      timezone: 'Asia/Shanghai',
      resolution: '1366,768',
      hardwareConcurrency: '4',
      webrtcPolicy: 'default_public_interface_only',
    },
  },
  {
    id: 'win-chrome-us-user',
    name: 'Windows / Chrome / 美国用户',
    description: '模拟美国普通用户，英文环境',
    config: {
      brand: 'Chrome',
      platform: 'windows',
      lang: 'en-US',
      timezone: 'America/Los_Angeles',
      resolution: '1920,1080',
      hardwareConcurrency: '8',
      webrtcPolicy: 'disable_non_proxied_udp',
    },
  },
  {
    id: 'mac-chrome-jp',
    name: 'macOS / Chrome / 日本用户',
    description: '模拟日本 Mac 用户，日语环境',
    config: {
      brand: 'Chrome',
      platform: 'macos',
      lang: 'ja-JP',
      timezone: 'Asia/Tokyo',
      resolution: '1440,900',
      hardwareConcurrency: '8',
      webrtcPolicy: 'disable_non_proxied_udp',
    },
  },
  {
    id: 'win-chrome-uk-office',
    name: 'Windows / Chrome / 英国-办公',
    description: '模拟英国办公室 Windows 用户，英文环境 (en-GB)',
    config: {
      brand: 'Chrome',
      platform: 'windows',
      lang: 'en-GB',
      timezone: 'Europe/London',
      resolution: '1920,1080',
      hardwareConcurrency: '8',
      webrtcPolicy: 'disable_non_proxied_udp',
    },
  },
  {
    id: 'mac-chrome-us-edu',
    name: 'macOS / Chrome / 美国-教育',
    description: '模拟美国大学教育网 Mac 用户，英文环境 (en-US)',
    config: {
      brand: 'Chrome',
      platform: 'macos',
      lang: 'en-US',
      timezone: 'America/New_York',
      resolution: '1440,900',
      hardwareConcurrency: '8',
      webrtcPolicy: 'disable_non_proxied_udp',
    },
  },
]

export function applyLocaleToFingerprintArgs(args: string[], lang: string, timezone: string): string[] {
  const nextConfig = deserialize(args || [])
  if (lang) nextConfig.lang = lang
  if (timezone) nextConfig.timezone = timezone
  return serialize(nextConfig)
}
