import { useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth.store'

export type DashboardContext = 'hq' | 'company'

/**
 * Hook untuk mendeteksi konteks dashboard saat ini (HQ atau Company).
 * - Jika URL dimulai dengan /g/, konteks adalah HQ
 * - Jika URL dimulai dengan /c/, konteks adalah Company
 */
export function useDashboardContext(): DashboardContext {
  const location = useLocation()
  const { selectedCompany } = useAuthStore()

  // Jika path dimulai dengan /g/, ini adalah HQ context
  if (location.pathname.startsWith('/g/')) {
    return 'hq'
  }

  // Jika path dimulai dengan /c/, ini adalah Company context
  if (location.pathname.startsWith('/c/')) {
    return 'company'
  }

  // Default fallback: jika ada selectedCompany, anggap Company context
  return selectedCompany ? 'company' : 'hq'
}
