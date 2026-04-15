import { usePermission } from '@/hooks/usePermission'

// ─── Types ────────────────────────────────────────────────────────────────────

export interface PermissionGateProps {
  children: React.ReactNode
  permission?: string
  anyOf?: string[]
  allOf?: string[]
  role?: string | string[]
  fallback?: React.ReactNode
}

// ─── Component ────────────────────────────────────────────────────────────────

export function PermissionGate({
  children,
  permission,
  anyOf,
  allOf,
  role,
  fallback = null,
}: PermissionGateProps) {
  const { can, canAny, canAll, hasRole } = usePermission()

  const checks: boolean[] = []

  if (permission !== undefined) checks.push(can(permission))
  if (anyOf !== undefined) checks.push(canAny(anyOf))
  if (allOf !== undefined) checks.push(canAll(allOf))
  if (role !== undefined) checks.push(hasRole(role))

  // No props given → always render children
  if (checks.length === 0) return <>{children}</>

  // All provided conditions must pass (AND logic)
  const allowed = checks.every(Boolean)

  return allowed ? <>{children}</> : <>{fallback}</>
}
