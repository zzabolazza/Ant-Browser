const getBindings = async () => {
  const direct = (globalThis as any).go?.main?.App
  if (direct) return direct
  try {
    return await import('../../wailsjs/go/main/App')
  } catch {
    return null
  }
}

export interface BackupActionResult {
  cancelled?: boolean
  message?: string
  zipPath?: string
  path?: string
  remotePath?: string
  resetFirst?: boolean
  imported?: number
  skipped?: number
  conflicts?: number
  partial?: boolean
  componentTotal?: number
  componentSuccess?: number
  componentFailed?: number
  failedComponents?: Array<{
    componentId?: string
    componentName?: string
    error?: string
  }>
}

export interface BackupWebDAVSettings {
  url: string
  username: string
  password?: string
  remoteDir: string
  hasPassword: boolean
}

export async function initializeSystemData(): Promise<BackupActionResult> {
  const bindings: any = await getBindings()
  if (!bindings?.BackupInitializeSystem) {
    return { cancelled: false, message: '当前环境不支持后端初始化接口' }
  }
  return (await bindings.BackupInitializeSystem()) || {}
}

export async function exportSystemConfig(password: string, target: 'local' | 'webdav'): Promise<BackupActionResult> {
  const bindings: any = await getBindings()
  if (target === 'webdav') {
    if (!bindings?.BackupExportPackageToWebDAV) {
      return { cancelled: false, message: '当前环境不支持 WebDAV 导出接口' }
    }
    return (await bindings.BackupExportPackageToWebDAV(password)) || {}
  }
  if (!bindings?.BackupExportPackage) {
    return { cancelled: false, message: '当前环境不支持后端导出接口' }
  }
  return (await bindings.BackupExportPackage(password)) || {}
}

export async function pickImportBackupFile(): Promise<BackupActionResult> {
  const bindings: any = await getBindings()
  if (!bindings?.BackupPickImportFile) {
    return { cancelled: true, message: '当前环境不支持选择备份文件' }
  }
  return (await bindings.BackupPickImportFile()) || { cancelled: true }
}

export async function importSystemConfig(
  resetFirst: boolean,
  password: string,
  zipPath: string,
): Promise<BackupActionResult> {
  const bindings: any = await getBindings()
  if (!bindings?.BackupImportPackage) {
    return { cancelled: false, message: '当前环境不支持后端加载接口' }
  }
  return (await bindings.BackupImportPackage(resetFirst, password, zipPath)) || {}
}

export async function getBackupWebDAVSettings(): Promise<BackupWebDAVSettings> {
  const bindings: any = await getBindings()
  if (!bindings?.BackupGetWebDAVSettings) {
    return { url: '', username: '', remoteDir: '', hasPassword: false }
  }
  return (await bindings.BackupGetWebDAVSettings()) || { url: '', username: '', remoteDir: '', hasPassword: false }
}

export async function saveBackupWebDAVSettings(settings: BackupWebDAVSettings): Promise<void> {
  const bindings: any = await getBindings()
  if (!bindings?.BackupSaveWebDAVSettings) {
    throw new Error('当前环境不支持 WebDAV 设置')
  }
  await bindings.BackupSaveWebDAVSettings(settings)
}
