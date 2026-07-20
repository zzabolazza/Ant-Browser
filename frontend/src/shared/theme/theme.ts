export type ThemeMode = 'dark' | 'light' | 'cream' | 'mint' | 'ocean'

const THEME_STORAGE_KEY = 'app-theme'
export const DEFAULT_THEME: ThemeMode = 'light'

function isThemeMode(value: unknown): value is ThemeMode {
  return value === 'dark' || value === 'light' || value === 'cream' || value === 'mint' || value === 'ocean'
}

async function getBindings(): Promise<any | null> {
  const direct = (globalThis as any).go?.main?.App
  if (direct) return direct
  try {
    return await import('../../wailsjs/go/main/App')
  } catch {
    return null
  }
}

export function getStoredTheme(): ThemeMode {
  const stored = localStorage.getItem(THEME_STORAGE_KEY)
  if (isThemeMode(stored)) return stored
  return DEFAULT_THEME
}

export function applyTheme(mode: ThemeMode) {
  document.documentElement.dataset.theme = mode
  document.documentElement.style.colorScheme = mode === 'dark' ? 'dark' : 'light'
}

async function persistThemeToBackend(mode: ThemeMode | null): Promise<void> {
  const bindings = await getBindings()
  if (!bindings?.SaveUIPreferences) return
  await bindings.SaveUIPreferences(mode ? { theme: mode } : {})
}

export function setThemeMode(mode: ThemeMode) {
  localStorage.setItem(THEME_STORAGE_KEY, mode)
  applyTheme(mode)
  void persistThemeToBackend(mode)
}

export function initializeTheme() {
  applyTheme(getStoredTheme())
  void hydrateThemeFromBackend()
}

/** 从本机 data/ui-preferences.json 同步主题（备份/恢复后会走这里）。 */
export async function hydrateThemeFromBackend(): Promise<ThemeMode> {
  const bindings = await getBindings()
  if (!bindings?.GetUIPreferences) {
    return getStoredTheme()
  }
  try {
    const prefs = (await bindings.GetUIPreferences()) || {}
    if (isThemeMode(prefs.theme)) {
      localStorage.setItem(THEME_STORAGE_KEY, prefs.theme)
      applyTheme(prefs.theme)
      return prefs.theme
    }
  } catch {
    // ignore and fall back to localStorage
  }
  // 迁移：本地有主题但后端还没有时，写入 data/ 以便纳入备份
  const local = getStoredTheme()
  void persistThemeToBackend(local)
  return local
}

/** 导出前确保当前主题已写入可备份文件。 */
export async function flushThemeForBackup(): Promise<void> {
  await persistThemeToBackend(getStoredTheme())
}

export function resetThemeMode() {
  localStorage.removeItem(THEME_STORAGE_KEY)
  applyTheme(DEFAULT_THEME)
  void persistThemeToBackend(DEFAULT_THEME)
}
