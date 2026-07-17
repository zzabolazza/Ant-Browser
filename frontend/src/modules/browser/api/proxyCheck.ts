import type { ProxyCheckSettings } from '../types'
import { getBindings } from './runtime'

export function createDefaultProxyCheckSettings(): ProxyCheckSettings {
  return {
    prepareTimeoutMs: 15000,
    speedTargetId: 'gstatic_generate_204',
    ipHealthTargetId: 'ippure_info',
    targets: [
      {
        id: 'gstatic_generate_204',
        name: 'Google generate_204',
        type: 'speed',
        url: 'http://www.gstatic.com/generate_204',
        timeoutMs: 3000,
        expectedStatus: [204],
      },
      {
        id: 'ippure_info',
        name: 'IPPure 出口信息',
        type: 'ip_health',
        url: 'https://my.ippure.com/v1/info',
        parser: 'json',
        timeoutMs: 10000,
      },
    ],
  }
}

export async function fetchProxyCheckSettings(): Promise<ProxyCheckSettings> {
  const bindings: any = await getBindings()
  if (bindings?.GetProxyCheckSettings) {
    return (await bindings.GetProxyCheckSettings()) || createDefaultProxyCheckSettings()
  }
  return createDefaultProxyCheckSettings()
}

export async function saveProxyCheckSettings(settings: ProxyCheckSettings): Promise<boolean> {
  const bindings: any = await getBindings()
  if (bindings?.SaveProxyCheckSettings) {
    await bindings.SaveProxyCheckSettings(settings)
    return true
  }
  return true
}
