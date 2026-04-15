import { useAuthStore } from '@/stores/auth.store'

/**
 * Returns a helper to build paths scoped to the current company context.
 * Usage: const path = useCompanyPath(); navigate(path('users'))  →  /c/CORP-01/users
 */
export function useCompanyPath() {
  const selectedCompany = useAuthStore((s) => s.selectedCompany)
  const base = selectedCompany ? `/c/${selectedCompany.code}` : ''

  return (segment: string) => `${base}/${segment}`
}
