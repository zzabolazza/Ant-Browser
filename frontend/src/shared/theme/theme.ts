export type ThemeMode = 'dark' | 'light' | 'cream' | 'mint' | 'ocean'

const THEME_STORAGE_KEY = 'app-theme'

function isThemeMode(value: unknown): value is ThemeMode {
  return value === 'dark' || value === 'light' || value === 'cream' || value === 'mint' || value === 'ocean'
}

export function getStoredTheme(): ThemeMode {
  const stored = localStorage.getItem(THEME_STORAGE_KEY)
  if (isThemeMode(stored)) return stored
  return 'light'
}

export function applyTheme(mode: ThemeMode) {
  document.documentElement.dataset.theme = mode
  document.documentElement.style.colorScheme = mode === 'dark' ? 'dark' : 'light'
}

export function setThemeMode(mode: ThemeMode) {
  localStorage.setItem(THEME_STORAGE_KEY, mode)
  applyTheme(mode)
}

export function initializeTheme() {
  applyTheme(getStoredTheme())
}
