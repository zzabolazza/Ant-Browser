import { useCallback, useEffect, useState } from 'react'
import { toast } from '../../../../shared/components'
import { browserProxyCoreDownload, browserProxyCoreStatus, fetchBrowserSettings, saveBrowserSettings } from '../../api'
import type { BrowserSettings, ProxyCoreDownloadProgress, ProxyCoreStatusResult } from '../../types'
import { EventsOff, EventsOn } from '../../../../wailsjs/runtime/runtime'

function defaultProxyCoreTarget(): { goos: string; goarch: string } {
  const platform = navigator.platform.toLowerCase()
  const userAgent = navigator.userAgent.toLowerCase()
  const goos = platform.includes('mac') ? 'darwin' : platform.includes('linux') ? 'linux' : 'windows'
  const goarch = platform.includes('arm') || userAgent.includes('arm64') || userAgent.includes('aarch64') ? 'arm64' : 'amd64'
  return { goos, goarch }
}

export function useProxyCoreDownload() {
  const defaultTarget = defaultProxyCoreTarget()
  const [browserSettings, setBrowserSettings] = useState<BrowserSettings | null>(null)
  const [connectorSwitching, setConnectorSwitching] = useState(false)
  const [coreDownloadOpen, setCoreDownloadOpen] = useState(false)
  const [coreDownloadType, setCoreDownloadType] = useState('xray')
  const [coreDownloadGOOS, setCoreDownloadGOOS] = useState(defaultTarget.goos)
  const [coreDownloadGOARCH, setCoreDownloadGOARCH] = useState(defaultTarget.goarch)
  const [coreDownloadProxy, setCoreDownloadProxy] = useState('')
  const [coreDownloadProgress, setCoreDownloadProgress] = useState<ProxyCoreDownloadProgress | null>(null)
  const [currentCoreStatus, setCurrentCoreStatus] = useState<ProxyCoreStatusResult | null>(null)
  const [downloadCoreStatus, setDownloadCoreStatus] = useState<ProxyCoreStatusResult | null>(null)
  const [downloadCoreStatusLoading, setDownloadCoreStatusLoading] = useState(false)

  const refreshCurrentCoreStatus = useCallback(async (core: string) => {
    const target = defaultProxyCoreTarget()
    const status = await browserProxyCoreStatus(core || 'xray', target.goos, target.goarch)
    setCurrentCoreStatus(status)
    return status
  }, [])

  const refreshDownloadCoreStatus = useCallback(async (
    core = coreDownloadType,
    goos = coreDownloadGOOS,
    goarch = coreDownloadGOARCH,
  ) => {
    setDownloadCoreStatusLoading(true)
    try {
      const status = await browserProxyCoreStatus(core, goos, goarch)
      setDownloadCoreStatus(status)
      return status
    } finally {
      setDownloadCoreStatusLoading(false)
    }
  }, [coreDownloadType, coreDownloadGOOS, coreDownloadGOARCH])

  const loadBrowserSettings = useCallback(async () => {
    try {
      const settings = await fetchBrowserSettings()
      setBrowserSettings(settings)
      await refreshCurrentCoreStatus(settings.defaultConnectorType || 'xray')
    } catch (error: any) {
      toast.error(error?.message || '读取代理内核失败')
    }
  }, [refreshCurrentCoreStatus])

  const handleSwitchConnector = useCallback(async () => {
    const currentSettings = browserSettings || await fetchBrowserSettings()
    const currentConnector = currentSettings.defaultConnectorType === 'mihomo' ? 'mihomo' : 'xray'
    const nextConnector = currentConnector === 'mihomo' ? 'xray' : 'mihomo'
    const target = defaultProxyCoreTarget()
    setConnectorSwitching(true)
    try {
      const status = await browserProxyCoreStatus(nextConnector, target.goos, target.goarch)
      if (!status.installed) {
        setCoreDownloadType(nextConnector)
        setCoreDownloadGOOS(target.goos)
        setCoreDownloadGOARCH(target.goarch)
        setDownloadCoreStatus(status)
        setCoreDownloadProgress(null)
        setCoreDownloadOpen(true)
        toast.warning(`${nextConnector === 'mihomo' ? 'Mihomo' : 'Xray'} 未下载，先下载内核`)
        return
      }
      const nextSettings = { ...currentSettings, defaultConnectorType: nextConnector }
      await saveBrowserSettings(nextSettings)
      setBrowserSettings(nextSettings)
      setCurrentCoreStatus(status)
      toast.success(`已切换到 ${nextConnector === 'mihomo' ? 'Mihomo' : 'Xray'}`)
    } catch (error: any) {
      toast.error(error?.message || '切换代理内核失败')
    } finally {
      setConnectorSwitching(false)
    }
  }, [browserSettings])

  useEffect(() => {
    const onProgress = (data: ProxyCoreDownloadProgress) => {
      setCoreDownloadProgress(data)
      if (data.phase === 'done') {
        toast.success(data.message || '代理内核已安装')
        void refreshDownloadCoreStatus(data.core, data.goos, data.goarch)
        if ((browserSettings?.defaultConnectorType || 'xray') === data.core) void refreshCurrentCoreStatus(data.core)
      }
      if (data.phase === 'error') toast.error(data.message || '代理内核下载失败')
    }
    EventsOn('proxy-core:download:progress', onProgress)
    return () => EventsOff('proxy-core:download:progress')
  }, [browserSettings?.defaultConnectorType, refreshCurrentCoreStatus, refreshDownloadCoreStatus])

  useEffect(() => {
    if (coreDownloadOpen) void refreshDownloadCoreStatus(coreDownloadType, coreDownloadGOOS, coreDownloadGOARCH)
  }, [coreDownloadOpen, coreDownloadType, coreDownloadGOOS, coreDownloadGOARCH, refreshDownloadCoreStatus])

  const handleStartCoreDownload = useCallback(async () => {
    const downloadProxy = coreDownloadProxy.trim()
    setCoreDownloadProgress({ core: coreDownloadType, goos: coreDownloadGOOS, goarch: coreDownloadGOARCH, phase: 'starting', progress: 0, message: downloadProxy ? `准备下载（指定代理：${downloadProxy}）` : '准备下载（直连）' })
    try {
      await browserProxyCoreDownload(coreDownloadType, coreDownloadGOOS, coreDownloadGOARCH, downloadProxy)
    } catch (error: any) {
      const message = error?.message || '启动下载失败'
      setCoreDownloadProgress({ core: coreDownloadType, goos: coreDownloadGOOS, goarch: coreDownloadGOARCH, phase: 'error', progress: 0, message })
      toast.error(message)
    }
  }, [coreDownloadProxy, coreDownloadType, coreDownloadGOOS, coreDownloadGOARCH])

  const openCoreDownload = useCallback(() => {
    setCoreDownloadProgress(null)
    setCoreDownloadOpen(true)
    void refreshDownloadCoreStatus()
  }, [refreshDownloadCoreStatus])

  const closeCoreDownload = useCallback(() => {
    setCoreDownloadOpen(false)
    setCoreDownloadProgress(null)
  }, [])

  return {
    browserSettings,
    connectorSwitching,
    coreDownloadOpen,
    setCoreDownloadOpen,
    coreDownloadType,
    setCoreDownloadType,
    coreDownloadGOOS,
    setCoreDownloadGOOS,
    coreDownloadGOARCH,
    setCoreDownloadGOARCH,
    coreDownloadProxy,
    setCoreDownloadProxy,
    coreDownloadProgress,
    currentCoreStatus,
    downloadCoreStatus,
    downloadCoreStatusLoading,
    loadBrowserSettings,
    handleSwitchConnector,
    handleStartCoreDownload,
    openCoreDownload,
    closeCoreDownload,
  }
}
