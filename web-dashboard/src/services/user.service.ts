import { apiClient } from './api.client'
import type { ListParams } from './createEntityService'
import type { User, CreateUserRequest, UpdateUserRequest } from '@/types/user.types'

// ─── API response shape (snake_case from backend) ───────────────────────────

interface ApiUser {
  id: string
  email: string
  full_name: string
  is_active: boolean
  is_superadmin: boolean
  phone: string
  roles: string[]
  created_at: string
  updated_at: string
}

interface ApiUserListResponse {
  data: ApiUser[]
  meta: {
    total: number
    page: number
    per_page: number
    total_pages: number
  }
}

// ─── Public response types ──────────────────────────────────────────────────

export interface UserListResult {
  items: User[]
  total: number
  page: number
  perPage: number
  totalPages: number
}

// ─── Transformer: snake_case API → camelCase app ────────────────────────────

function transformUser(raw: ApiUser): User {
  return {
    id: raw.id,
    email: raw.email,
    fullName: raw.full_name,
    isActive: raw.is_active,
    isSuperadmin: raw.is_superadmin,
    phone: raw.phone,
    roles: raw.roles,
    createdAt: raw.created_at,
    updatedAt: raw.updated_at,
  }
}

// ─── Query string builder ───────────────────────────────────────────────────

function buildQueryString(params?: ListParams): string {
  if (!params) return ''
  const q = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== null && v !== '') {
      q.set(k, String(v))
    }
  })
  const str = q.toString()
  return str ? `?${str}` : ''
}

// ─── Service ────────────────────────────────────────────────────────────────

const BASE_PATH = '/api/users'

export const userService = {
  list: async (params?: ListParams): Promise<UserListResult> => {
    const response = await apiClient.get<ApiUserListResponse>(
      `${BASE_PATH}${buildQueryString(params)}`,
    )
    return {
      items: response.data.map(transformUser),
      total: response.meta.total,
      page: response.meta.page,
      perPage: response.meta.per_page,
      totalPages: response.meta.total_pages,
    }
  },

  getById: async (id: string): Promise<User> => {
    const raw = await apiClient.get<ApiUser>(`${BASE_PATH}/${id}`)
    return transformUser(raw)
  },

  create: async (data: CreateUserRequest): Promise<User> => {
    const raw = await apiClient.post<ApiUser>(BASE_PATH, data)
    return transformUser(raw)
  },

  update: async (id: string, data: UpdateUserRequest): Promise<User> => {
    const raw = await apiClient.put<ApiUser>(`${BASE_PATH}/${id}`, data)
    return transformUser(raw)
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete<void>(`${BASE_PATH}/${id}`)
  },
}
