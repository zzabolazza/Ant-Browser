export interface CoreDisplayInfo {
  coreId: string
  coreName: string
  corePath: string
  isDefault: boolean
  pathValid: boolean
  pathMessage: string
  chromeVersion: string
  instanceCount: number
}

export interface CoreSettingsForm {
  userDataRoot: string
  defaultFingerprintArgs: string
  defaultLaunchArgs: string
  defaultStartUrls: string
  lightStartEnabled: boolean
  restoreLastSession: boolean
  startReadyTimeoutMs: number
  startStableWindowMs: number
}

export interface CoreEditForm {
  coreName: string
  corePath: string
}
