import type { ProxyIPHealthResult } from '../types'

const COUNTRY_NAME_TO_CODE: Record<string, string> = {
  china: 'CN',
  'mainland china': 'CN',
  中国: 'CN',
  'hong kong': 'HK',
  'hong kong sar': 'HK',
  'hong kong sar china': 'HK',
  'hong kong, china': 'HK',
  'china hong kong': 'HK',
  香港: 'HK',
  中国香港: 'HK',
  taiwan: 'TW',
  台湾: 'TW',
  'united states': 'US',
  usa: 'US',
  us: 'US',
  美国: 'US',
  'united kingdom': 'GB',
  uk: 'GB',
  'great britain': 'GB',
  英国: 'GB',
  japan: 'JP',
  日本: 'JP',
  'south korea': 'KR',
  korea: 'KR',
  韩国: 'KR',
  singapore: 'SG',
  新加坡: 'SG',
  germany: 'DE',
  德国: 'DE',
  france: 'FR',
  法国: 'FR',
  netherlands: 'NL',
  荷兰: 'NL',
  canada: 'CA',
  加拿大: 'CA',
  australia: 'AU',
  澳大利亚: 'AU',
  russia: 'RU',
  俄罗斯: 'RU',
  brazil: 'BR',
  巴西: 'BR',
  india: 'IN',
  印度: 'IN',
}

/** Normalize country name / code to ISO 3166-1 alpha-2 (same coverage as backend). */
export function normalizeCountryCode(country: string): string {
  const value = (country || '').trim()
  if (!value) return ''
  const upper = value.toUpperCase()
  if (upper.length === 2 && /^[A-Z]{2}$/.test(upper)) return upper
  return COUNTRY_NAME_TO_CODE[value.toLowerCase()] || ''
}

/** Convert ISO 3166-1 alpha-2 to regional-indicator flag emoji (e.g. US → 🇺🇸). */
export function countryCodeToFlagEmoji(code: string): string {
  const normalized = (code || '').trim().toUpperCase()
  if (!/^[A-Z]{2}$/.test(normalized)) return ''
  const A = 0x1f1e6
  const base = 'A'.charCodeAt(0)
  return String.fromCodePoint(
    A + (normalized.charCodeAt(0) - base),
    A + (normalized.charCodeAt(1) - base),
  )
}

export function resolveProxyCountryDisplay(
  health: Pick<ProxyIPHealthResult, 'ok' | 'country' | 'countryCode'> | null | undefined,
): { code: string; flag: string } | null {
  if (!health?.ok) return null
  const code = normalizeCountryCode(health.countryCode || '') || normalizeCountryCode(health.country || '')
  if (!code) return null
  const flag = countryCodeToFlagEmoji(code)
  if (!flag) return null
  return { code, flag }
}

export function parseProxyIPHealthResult(raw?: string): ProxyIPHealthResult | null {
  const value = (raw || '').trim()
  if (!value) return null
  try {
    const parsed = JSON.parse(value) as ProxyIPHealthResult
    return parsed && typeof parsed === 'object' ? parsed : null
  } catch {
    return null
  }
}

export function resolveProfileProxyCountryDisplay(
  proxyId: string,
  lastIPHealthJson: string | undefined,
  ipHealthMap: Record<string, ProxyIPHealthResult>,
): { code: string; flag: string } | null {
  const cached = proxyId ? ipHealthMap[proxyId] : undefined
  const persisted = parseProxyIPHealthResult(lastIPHealthJson)
  return resolveProxyCountryDisplay(cached) || resolveProxyCountryDisplay(persisted)
}
