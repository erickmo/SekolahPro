// ─── User roles ───────────────────────────────────────────────────────────────
// Single-tenant: 'admin' | 'user' | 'viewer'
// Multi-tenant:  add 'superuser' | 'tenant_owner' | 'employee'

export type UserRole =
  | 'superuser'    // multi-tenant: platform-level admin
  | 'tenant_owner' // multi-tenant: manages company groups
  | 'employee'     // multi-tenant: regular company user
  | 'admin'        // single-tenant: admin
  | 'user'         // single-tenant: regular user
  | 'viewer'       // single-tenant: read-only user
  | string

export interface UserProfile {
  id: string
  name: string
  email: string
  avatar?: string
  role: UserRole
  permissions: string[]
}

// ─── Multi-tenant types ───────────────────────────────────────────────────────
// Always defined — just unused (null) in single-tenant mode.

export interface Company {
  id: string
  code: string
  name: string
  logo?: string
  groupId?: string
}

export interface CompanyGroup {
  id: string
  name: string
  companies: Company[]
}

// ─── Auth request/response ────────────────────────────────────────────────────

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
  refreshToken: string
  user: UserProfile
  // Multi-tenant only — omit or pass empty array in single-tenant
  companyGroups?: CompanyGroup[]
}

// ─── Raw API response types (match your backend) ─────────────────────────────

export interface ApiLoginResult {
  access_token: string
  refresh_token: string
  expires_at: number
  user_id: string
  name: string
  email: string
  role: string
}
