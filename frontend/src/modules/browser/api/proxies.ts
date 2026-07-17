import type { BrowserProxy, ProxyIPHealthResult, ProxyLocationResolveResult, ProxySpeedTestResult } from '../types'
import { getBindings, getMockProxies, nowISOString, setMockProxies } from './runtime'

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

export async function fetchBrowserProxies(): Promise<BrowserProxy[]> {
  const bindings: any = await getBindings()
  if (bindings?.BrowserProxyList) {
    return (await bindings.BrowserProxyList()) || []
  }
  return getMockProxies()
}

export async function fetchBrowserProxyGroups(): Promise<string[]> {
  const bindings: any = await getBindings()
  if (bindings?.BrowserProxyListGroups) {
    return (await bindings.BrowserProxyListGroups()) || []
  }
  return []
}

export async function saveBrowserProxies(proxies: BrowserProxy[]): Promise<boolean> {
  const bindings: any = await getBindings()
  if (bindings?.SaveBrowserProxies) {
    await bindings.SaveBrowserProxies(proxies)
    return true
  }
  setMockProxies(proxies)
  return true
}

export async function validateProxyConfig(proxyConfig: string, proxyId: string): Promise<{ supported: boolean; errorMsg: string }> {
  const bindings: any = await getBindings()
  if (bindings?.ValidateProxyConfig) {
    return (await bindings.ValidateProxyConfig(proxyConfig, proxyId)) || { supported: true, errorMsg: '' }
  }
  return { supported: true, errorMsg: '' }
}

export async function testProxyConnectivity(proxyId: string, proxyConfig: string): Promise<ProxySpeedTestResult> {
  const bindings: any = await getBindings()
  if (bindings?.TestProxyConnectivity) {
    return (await bindings.TestProxyConnectivity(proxyId, proxyConfig)) || { proxyId, ok: false, latencyMs: 0, engine: 'unknown', error: '调用失败' }
  }
  await sleep(300 + Math.random() * 500)
  return { proxyId, ok: true, latencyMs: Math.floor(100 + Math.random() * 200), engine: 'mock', error: '' }
}

export async function browserProxyTestSpeed(proxyId: string): Promise<ProxySpeedTestResult> {
  const bindings: any = await getBindings()
  if (bindings?.BrowserProxyTestSpeed) {
    return (await bindings.BrowserProxyTestSpeed(proxyId)) || { proxyId, ok: false, latencyMs: 0, engine: 'unknown', error: '调用失败' }
  }
  await sleep(300 + Math.random() * 500)
  return { proxyId, ok: true, latencyMs: Math.floor(100 + Math.random() * 400), engine: 'mock', error: '' }
}

export async function browserProxyBatchTestSpeed(proxyIds: string[], concurrency: number = 20): Promise<ProxySpeedTestResult[]> {
  const bindings: any = await getBindings()
  if (bindings?.BrowserProxyBatchTestSpeed) {
    return (await bindings.BrowserProxyBatchTestSpeed(proxyIds, concurrency)) || []
  }
  await sleep(1000)
  return proxyIds.map((proxyId) => ({ proxyId, ok: true, latencyMs: Math.floor(100 + Math.random() * 400), engine: 'mock', error: '' }))
}

export async function browserProxyCheckIPHealth(proxyId: string): Promise<ProxyIPHealthResult> {
  const bindings: any = await getBindings()
  if (bindings?.BrowserProxyCheckIPHealth) {
    return (
      (await bindings.BrowserProxyCheckIPHealth(proxyId)) || {
        proxyId,
        ok: false,
        source: 'ip_health',
        error: '调用失败',
        ip: '',
        fraudScore: 0,
        isResidential: false,
        isBroadcast: false,
        country: '',
        countryCode: '',
        region: '',
        city: '',
        asOrganization: '',
        rawData: {},
        updatedAt: nowISOString(),
      }
    )
  }

  await sleep(600)
  return {
    proxyId,
    ok: true,
    source: 'ip_health',
    error: '',
    ip: '127.0.0.1',
    fraudScore: Math.floor(Math.random() * 100),
    isResidential: Math.random() > 0.5,
    isBroadcast: false,
    country: 'US',
    countryCode: 'US',
    region: 'Mock',
    city: 'Mock',
    asOrganization: 'Mock ISP',
    rawData: {},
    updatedAt: nowISOString(),
  }
}

export async function browserProxyResolveLocation(proxyId: string, proxyConfig: string = ''): Promise<ProxyLocationResolveResult> {
  const bindings: any = await getBindings()
  if (bindings?.BrowserProxyResolveLocation) {
    return (await bindings.BrowserProxyResolveLocation(proxyId, proxyConfig)) || {
      proxyId,
      ok: false,
      auto: false,
      source: 'location',
      error: '调用失败',
      ip: '',
      country: '',
      region: '',
      city: '',
      timezone: '',
      lang: '',
      resolvedAt: nowISOString(),
    }
  }

  await sleep(400)
  return {
    proxyId: proxyId || '__direct__',
    ok: true,
    auto: true,
    source: 'mock',
    error: '',
    ip: '127.0.0.1',
    country: 'US',
    region: 'New York',
    city: 'New York',
    timezone: 'America/New_York',
    lang: 'en-US',
    resolvedAt: nowISOString(),
  }
}

export async function browserProxyBatchCheckIPHealth(proxyIds: string[], concurrency: number = 10): Promise<ProxyIPHealthResult[]> {
  const bindings: any = await getBindings()
  if (bindings?.BrowserProxyBatchCheckIPHealth) {
    return (await bindings.BrowserProxyBatchCheckIPHealth(proxyIds, concurrency)) || []
  }

  await sleep(1200)
  return proxyIds.map((proxyId) => ({
    proxyId,
    ok: true,
    source: 'ip_health',
    error: '',
    ip: '127.0.0.1',
    fraudScore: Math.floor(Math.random() * 100),
    isResidential: Math.random() > 0.5,
    isBroadcast: false,
    country: 'US',
    countryCode: 'US',
    region: 'Mock',
    city: 'Mock',
    asOrganization: 'Mock ISP',
    rawData: {},
    updatedAt: nowISOString(),
  }))
}
