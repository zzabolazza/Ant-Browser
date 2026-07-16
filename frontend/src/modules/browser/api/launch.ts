import { getBindings, getGoApp } from './runtime'

export interface LaunchServerInfo {
  host: string
  port: number
  preferredPort: number
  baseUrl: string
  ready: boolean
  apiAuth: {
    requested: boolean
    configured: boolean
    enabled: boolean
    header: string
  }
}

function normalizeLaunchServerInfo(payload: any): LaunchServerInfo {
  const host = String(payload?.host || '127.0.0.1')
  const port = Number(payload?.port) || 0
  const preferredPort = Number(payload?.preferredPort) || 0
  const fallbackPort = preferredPort > 0 ? preferredPort : 19876
  const effectivePort = port > 0 ? port : fallbackPort
  const baseUrl = String(payload?.baseUrl || (effectivePort > 0 ? `http://${host}:${effectivePort}` : ''))
  const apiAuthPayload = payload?.apiAuth || {}

  return {
    host,
    port: effectivePort,
    preferredPort,
    baseUrl,
    ready: !!payload?.ready && port > 0,
    apiAuth: {
      requested: !!apiAuthPayload?.requested,
      configured: !!apiAuthPayload?.configured,
      enabled: !!apiAuthPayload?.enabled,
      header: String(apiAuthPayload?.header || 'X-Ant-Api-Key'),
    },
  }
}

export async function fetchLaunchServerInfo(): Promise<LaunchServerInfo> {
  const bindings: any = await getBindings()
  if (bindings?.GetLaunchServerInfo) {
    return normalizeLaunchServerInfo(await bindings.GetLaunchServerInfo())
  }

  const goApp = getGoApp()
  if (goApp?.GetLaunchServerInfo) {
    return normalizeLaunchServerInfo(await goApp.GetLaunchServerInfo())
  }

  return {
    host: '127.0.0.1',
    port: 19876,
    preferredPort: 19876,
    baseUrl: 'http://127.0.0.1:19876',
    ready: false,
    apiAuth: {
      requested: false,
      configured: false,
      enabled: false,
      header: 'X-Ant-Api-Key',
    },
  }
}
