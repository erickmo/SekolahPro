import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type Theme = 'light' | 'dark' | 'system'
type Density = 'compact' | 'comfortable' | 'spacious'

interface UiState {
  theme: Theme
  sidebarOpen: boolean
  density: Density
}

interface UiActions {
  setTheme: (theme: Theme) => void
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  setDensity: (density: Density) => void
}

function applyTheme(theme: Theme) {
  const root = document.documentElement
  if (theme === 'system') {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    root.setAttribute('data-theme', prefersDark ? 'dark' : 'light')
  } else {
    root.setAttribute('data-theme', theme)
  }
}

export const useUiStore = create<UiState & UiActions>()(
  persist(
    (set) => ({
      theme: 'light',
      sidebarOpen: true,
      density: 'comfortable',

      setTheme: (theme: Theme) => {
        applyTheme(theme)
        set({ theme })
      },

      toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
      setSidebarOpen: (open: boolean) => set({ sidebarOpen: open }),
      setDensity: (density: Density) => set({ density }),
    }),
    {
      name: 'dashboard-ui',
      partialize: (state) => ({ theme: state.theme, density: state.density }),
      onRehydrateStorage: () => (state) => {
        if (state) applyTheme(state.theme)
      },
    },
  ),
)
