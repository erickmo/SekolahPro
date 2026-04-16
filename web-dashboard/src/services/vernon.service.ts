/**
 * Vernon-specific service factory.
 *
 * Vernon handler responses differ from standard API:
 * - List:  { data: [...], meta: { total, page, per_page, total_pages } }
 * - Detail: { data: {...}, meta: { _sync_status, _sync_version } }
 * - Create/Update: { data: {...} }
 * - Delete: 204 No Content
 *
 * This factory transforms Vernon responses into the app's internal format.
 */
import { apiClient } from './api.client'
import type { ListParams } from './createEntityService'

// ── Vernon API response shapes ──────────────────────────────────────────────

interface VernonListResponse<T> {
  data: T[]
  meta: {
    total: number
    page: number
    per_page: number
    total_pages: number
  }
}

interface VernonSingleResponse<T> {
  data: T
  meta?: {
    _sync_status: string
    _sync_version: number
  }
}

// ── Vernon paginated result ─────────────────────────────────────────────────

export interface VernonPaginatedResult<T> {
  items: T[]
  total: number
  page: number
  perPage: number
  totalPages: number
}

// ── Query string builder ────────────────────────────────────────────────────

function buildVernonQueryString(params?: ListParams): string {
  if (!params) return ''
  const q = new URLSearchParams()

  if (params._page) q.set('_page', String(params._page))
  if (params._limit) q.set('_limit', String(params._limit))
  if (params._sort) q.set('_sort', params._sort)

  // Remaining params become filters
  const skip = new Set(['_page', '_limit', '_sort', 'limit', 'offset', 'sort', 'order', 'search'])
  for (const [k, v] of Object.entries(params ?? {})) {
    if (skip.has(k) || v === undefined || v === null || v === '') continue
    q.set(k, String(v))
  }

  const str = q.toString()
  return str ? `?${str}` : ''
}

// ── Factory ─────────────────────────────────────────────────────────────────

export function createVernonService<T, TApi = T>(
  basePath: string,
  transform?: (raw: TApi) => T,
) {
  function mapItem(raw: TApi): T {
    return transform ? transform(raw) : (raw as unknown as T)
  }

  return {
    list: async (params?: ListParams): Promise<VernonPaginatedResult<T>> => {
      const response = await apiClient.get<VernonListResponse<TApi>>(
        `${basePath}${buildVernonQueryString(params)}`,
      )
      return {
        items: response.data.map(mapItem),
        total: response.meta.total,
        page: response.meta.page,
        perPage: response.meta.per_page,
        totalPages: response.meta.total_pages,
      }
    },

    getById: async (id: string): Promise<T> => {
      const response = await apiClient.get<VernonSingleResponse<TApi>>(
        `${basePath}/${id}`,
      )
      return mapItem(response.data)
    },

    create: (data: Record<string, unknown>): Promise<T> =>
      apiClient.post<VernonSingleResponse<TApi>>(basePath, data).then((r) => mapItem(r.data)),

    update: (id: string, data: Record<string, unknown>): Promise<T> =>
      apiClient.put<VernonSingleResponse<TApi>>(`${basePath}/${id}`, data).then((r) => mapItem(r.data)),

    patch: (id: string, data: Record<string, unknown>): Promise<T> =>
      apiClient.patch<VernonSingleResponse<TApi>>(`${basePath}/${id}`, data).then((r) => mapItem(r.data)),

    delete: (id: string): Promise<void> =>
      apiClient.delete(`${basePath}/${id}`),
  }
}
