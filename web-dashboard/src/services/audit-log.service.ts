import { apiClient } from '@/services/api.client'
import type { AuditLog, AuditLogFilters } from '@/types/audit-log.types'

// ─── Paginated response (inline — not yet in api.types) ───────────────────────

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

// ─── Service ──────────────────────────────────────────────────────────────────

export const auditLogService = {
  list: (filters: AuditLogFilters) =>
    apiClient.get<PaginatedResponse<AuditLog>>('/api/audit-logs', { params: filters }),
}
