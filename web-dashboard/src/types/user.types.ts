import type { UserRole } from './auth.types'

// ─── User entity ────────────────────────────────────────────────────────────

export interface User {
  id: string
  email: string
  fullName: string
  isActive: boolean
  isSuperadmin: boolean
  phone: string
  roles: UserRole[]
  createdAt: string
  updatedAt: string
}

// ─── API request types ──────────────────────────────────────────────────────

export interface CreateUserRequest {
  email: string
  password: string
  fullName: string
  phone?: string
  roles: UserRole[]
  isActive?: boolean
}

export interface UpdateUserRequest {
  email?: string
  fullName?: string
  phone?: string
  roles?: UserRole[]
  isActive?: boolean
}

// ─── Role display labels ────────────────────────────────────────────────────

export const ROLE_LABELS: Record<string, string> = {
  superuser: 'Superuser',
  tenant_owner: 'Pemilik Tenant',
  employee: 'Karyawan',
  admin: 'Admin',
  user: 'Pengguna',
  viewer: 'Peninjau',
}

// ─── Status labels ──────────────────────────────────────────────────────────

export const USER_STATUS_LABELS: Record<'active' | 'inactive', string> = {
  active: 'Aktif',
  inactive: 'Nonaktif',
}
