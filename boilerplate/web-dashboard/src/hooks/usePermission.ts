import { useAuthStore } from '@/stores/auth.store'

export function usePermission() {
  const { user } = useAuthStore()
  const permissions = user?.permissions ?? []
  const isSuperAdmin = permissions.includes('*')

  return {
    can: (permission: string): boolean =>
      isSuperAdmin || permissions.includes(permission),

    canAny: (perms: string[]): boolean =>
      isSuperAdmin || perms.some((p) => permissions.includes(p)),

    canAll: (perms: string[]): boolean =>
      isSuperAdmin || perms.every((p) => permissions.includes(p)),

    hasRole: (role: string | string[]): boolean => {
      const currentRole = user?.role ?? ''
      if (currentRole === 'admin') return true
      const roles = Array.isArray(role) ? role : [role]
      return roles.includes(currentRole)
    },
  }
}
