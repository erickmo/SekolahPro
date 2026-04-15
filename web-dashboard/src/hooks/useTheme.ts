import { useUiStore } from '@/stores/ui.store'

export function useTheme() {
  const { theme, setTheme } = useUiStore()
  const isDark = theme === 'dark' ||
    (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)

  return { theme, setTheme, isDark }
}
