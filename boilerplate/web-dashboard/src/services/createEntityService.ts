import { apiClient } from './api.client'
import type { PaginatedResponse } from '@/types/api.types'

export interface ListParams {
  limit?: number
  offset?: number
  sort?: string
  order?: 'asc' | 'desc'
  search?: string
  [key: string]: unknown
}

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

export function createEntityService<T, TApi = T>(
  basePath: string,
  transform?: (raw: TApi) => T,
  responseWrapper?: string,
) {
  return {
    list: async (params?: ListParams): Promise<PaginatedResponse<T>> => {
      const response = await apiClient.get<Record<string, PaginatedResponse<TApi>>>(
        `${basePath}${buildQueryString(params)}`,
      )
      const raw: PaginatedResponse<TApi> = responseWrapper
        ? response[responseWrapper]
        : (response as unknown as PaginatedResponse<TApi>)
      return {
        ...raw,
        items: transform ? raw.items.map(transform) : (raw.items as unknown as T[]),
      }
    },

    getById: async (id: string): Promise<T> => {
      const raw = await apiClient.get<TApi>(`${basePath}/${id}`)
      return transform ? transform(raw) : (raw as unknown as T)
    },

    create: (data: Partial<T>): Promise<T> =>
      apiClient.post<T>(basePath, data),

    update: (id: string, data: Partial<T>): Promise<T> =>
      apiClient.put<T>(`${basePath}/${id}`, data),

    delete: (id: string): Promise<void> =>
      apiClient.delete<void>(`${basePath}/${id}`),
  }
}
