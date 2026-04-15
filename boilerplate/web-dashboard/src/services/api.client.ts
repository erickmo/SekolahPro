import { AppApiError } from '@/types/api.types'
import { useAuthStore } from '@/stores/auth.store'

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

async function request<T>(
  path: string,
  config?: RequestInit & { overrideToken?: string },
): Promise<T> {
  const authState = useAuthStore.getState()
  const token = config?.overrideToken ?? authState.token
  const isFormData = config?.body instanceof FormData
  const headers: Record<string, string> = isFormData
    ? {}
    : { 'Content-Type': 'application/json' }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(`${BASE_URL}${path}`, {
    ...config,
    headers: { ...headers, ...config?.headers },
  })

  if (response.status === 401) {
    authState.logout()
    throw new AppApiError(401, 'Sesi telah berakhir, silakan login kembali')
  }

  if (!response.ok) {
    const body = await response.json().catch(() => ({})) as Record<string, unknown>
    const message = typeof body['error'] === 'string' ? body['error'] : 'Terjadi kesalahan'
    const errors = body['errors'] as Record<string, string[]> | undefined
    throw new AppApiError(response.status, message, errors)
  }

  if (response.status === 204) return undefined as T

  const text = await response.text()
  if (!text) return undefined as T

  return JSON.parse(text) as T
}

// ─── Standard API client (uses token from auth store) ────────────────────────

export const apiClient = {
  get: <T>(path: string) =>
    request<T>(path, { method: 'GET' }),

  post: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(body) }),

  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),

  patch: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'PATCH', body: JSON.stringify(body) }),

  delete: <T>(path: string) =>
    request<T>(path, { method: 'DELETE' }),

  postForm: <T>(path: string, body: FormData) =>
    request<T>(path, { method: 'POST', body }),

  download: async (path: string, filename: string) => {
    const authState = useAuthStore.getState()
    const headers: Record<string, string> = {}

    if (authState.token) {
      headers['Authorization'] = `Bearer ${authState.token}`
    }

    const response = await fetch(`${BASE_URL}${path}`, {
      method: 'GET',
      headers,
    })

    if (response.status === 401) {
      authState.logout()
      throw new AppApiError(401, 'Sesi telah berakhir, silakan login kembali')
    }

    if (!response.ok) {
      const body = await response.json().catch(() => ({})) as Record<string, unknown>
      const message = typeof body['error'] === 'string' ? body['error'] : 'Terjadi kesalahan'
      throw new AppApiError(response.status, message)
    }

    const blob = await response.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  },
}

// ─── Pre-store client (used during login before token is in auth store) ───────

export const apiClientWithToken = {
  get: <T>(token: string, path: string) =>
    request<T>(path, { method: 'GET', overrideToken: token }),

  post: <T>(token: string, path: string, body: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(body), overrideToken: token }),
}
