import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { Company, CompanyGroup, LoginResponse, UserProfile } from '@/types/auth.types'

// ─── State ────────────────────────────────────────────────────────────────────

interface AuthState {
  user: UserProfile | null
  token: string | null
  refreshToken: string | null
  isAuthenticated: boolean

  // Multi-tenant fields — null/empty in single-tenant mode
  selectedCompany: Company | null
  selectedGroup: CompanyGroup | null
  availableGroups: CompanyGroup[]
}

interface AuthActions {
  login: (response: LoginResponse) => void
  logout: () => void
  updateToken: (token: string, refreshToken: string) => void

  // Multi-tenant actions
  selectGroup: (group: CompanyGroup) => void
  selectCompany: (company: Company) => void
  setAvailableGroups: (groups: CompanyGroup[]) => void
}

const initialState: AuthState = {
  user: null,
  token: null,
  refreshToken: null,
  isAuthenticated: false,
  selectedCompany: null,
  selectedGroup: null,
  availableGroups: [],
}

// ─── Store ────────────────────────────────────────────────────────────────────

export const useAuthStore = create<AuthState & AuthActions>()(
  persist(
    (set) => ({
      ...initialState,

      login: (response: LoginResponse) => {
        const groups = response.companyGroups ?? []
        const defaultGroup = groups.length === 1 ? groups[0] : null
        const defaultCompany = defaultGroup?.companies.length === 1
          ? defaultGroup.companies[0]
          : null

        set({
          user: response.user,
          token: response.token,
          refreshToken: response.refreshToken,
          isAuthenticated: true,
          availableGroups: groups,
          selectedGroup: defaultGroup,
          selectedCompany: defaultCompany,
        })
      },

      logout: () => set(initialState),

      updateToken: (token: string, refreshToken: string) => {
        set({ token, refreshToken })
      },

      // Multi-tenant actions
      selectGroup: (group: CompanyGroup) => {
        set({
          selectedGroup: group,
          selectedCompany: group.companies.length === 1 ? group.companies[0] : null,
        })
      },

      selectCompany: (company: Company) => {
        set({ selectedCompany: company })
      },

      setAvailableGroups: (groups: CompanyGroup[]) => {
        set({ availableGroups: groups })
      },
    }),
    {
      name: 'dashboard-auth',
      partialize: (state) => ({
        token: state.token,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
        user: state.user,
        selectedCompany: state.selectedCompany,
        selectedGroup: state.selectedGroup,
        availableGroups: state.availableGroups,
      }),
    },
  ),
)
