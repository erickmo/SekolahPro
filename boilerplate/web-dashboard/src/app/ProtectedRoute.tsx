import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth.store'
import { appConfig } from '@/config/app.config'

interface RouteProps {
  children: React.ReactNode
}

// ─── Root redirect ────────────────────────────────────────────────────────────
// Determines where to send the user based on auth state and tenant mode.

export function RootRedirect() {
  const { isAuthenticated, user, selectedCompany } = useAuthStore()

  if (!isAuthenticated) return <Navigate to="/login" replace />

  if (!appConfig.isMultiTenant) {
    return <Navigate to="/dashboard" replace />
  }

  // Multi-tenant routing based on role
  if (user?.role === 'superuser') return <Navigate to="/su/dashboard" replace />
  if (user?.role === 'tenant_owner') return <Navigate to="/choose-company" replace />
  if (selectedCompany) return <Navigate to={`/c/${selectedCompany.code}/dashboard`} replace />
  return <Navigate to="/choose-company" replace />
}

// ─── Auth guard — requires any authenticated user ─────────────────────────────

export function AuthRoute({ children }: RouteProps) {
  const { isAuthenticated } = useAuthStore()
  const location = useLocation()
  if (!isAuthenticated) return <Navigate to="/login" state={{ from: location }} replace />
  return <>{children}</>
}

// ─── Guest guard — redirect authenticated users away from login ───────────────

export function GuestRoute({ children }: RouteProps) {
  const { isAuthenticated } = useAuthStore()
  if (isAuthenticated) return <Navigate to="/" replace />
  return <>{children}</>
}

// ─── Role guard — requires specific role(s) ───────────────────────────────────

export function RoleRoute({ children, role }: RouteProps & { role: string | string[] }) {
  const { isAuthenticated, user } = useAuthStore()
  const location = useLocation()

  if (!isAuthenticated) return <Navigate to="/login" state={{ from: location }} replace />

  const roles = Array.isArray(role) ? role : [role]
  if (!roles.includes(user?.role ?? '')) return <Navigate to="/403" replace />

  return <>{children}</>
}

// ─── Multi-tenant guards ──────────────────────────────────────────────────────
// These are only meaningful when appConfig.isMultiTenant === true.

/** Only platform superusers can access /su/* routes */
export function SuperuserRoute({ children }: RouteProps) {
  const { isAuthenticated, user } = useAuthStore()
  const location = useLocation()
  if (!isAuthenticated) return <Navigate to="/login" state={{ from: location }} replace />
  if (user?.role !== 'superuser') return <Navigate to="/" replace />
  return <>{children}</>
}

/** Only tenant_owners can access /g/* (HQ/Group) routes */
export function GroupRoute({ children }: RouteProps) {
  const { isAuthenticated, user } = useAuthStore()
  const location = useLocation()
  if (!isAuthenticated) return <Navigate to="/login" state={{ from: location }} replace />
  if (user?.role === 'superuser') return <Navigate to="/su/dashboard" replace />
  if (user?.role !== 'tenant_owner') return <Navigate to="/" replace />
  return <>{children}</>
}

/** Authenticated users with a selected company can access /c/:companyCode/* routes */
export function CompanyRoute({ children }: RouteProps) {
  const { isAuthenticated, user, selectedCompany } = useAuthStore()
  const location = useLocation()
  if (!isAuthenticated) return <Navigate to="/login" state={{ from: location }} replace />
  if (user?.role === 'superuser') return <Navigate to="/su/dashboard" replace />
  if (!selectedCompany) return <Navigate to="/choose-company" replace />
  return <>{children}</>
}
