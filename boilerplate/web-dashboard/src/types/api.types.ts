export interface PaginatedResponse<T> {
  items: T[]
  total: number
  limit: number
  offset: number
}

export class AppApiError extends Error {
  status: number
  errors?: Record<string, string[]>

  constructor(status: number, message: string, errors?: Record<string, string[]>) {
    super(message)
    this.name = 'AppApiError'
    this.status = status
    this.errors = errors
  }
}

export interface RequestConfig {
  method?: string
  headers?: Record<string, string>
  body?: string
}
