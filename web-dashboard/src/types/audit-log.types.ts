export type AuditAction =
  | 'create'
  | 'update'
  | 'delete'
  | 'login'
  | 'logout'
  | 'export'
  | 'view'

export interface AuditLog {
  id: string
  timestamp: string       // ISO datetime
  userId: string
  userName: string
  userEmail: string
  action: AuditAction
  entity: string          // e.g. 'User', 'Product', 'Order'
  entityId?: string
  description: string
  ipAddress?: string
  metadata?: Record<string, unknown>
}

export interface AuditLogFilters {
  search?: string
  action?: AuditAction | ''
  entity?: string
  dateFrom?: string
  dateTo?: string
  page: number
  pageSize: number
}
