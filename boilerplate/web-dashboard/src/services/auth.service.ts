import { apiClient } from './api.client'
import type { LoginRequest, LoginResponse } from '@/types/auth.types'

export const authService = {
  login: (body: LoginRequest): Promise<LoginResponse> =>
    apiClient.post<LoginResponse>('/api/auth/login', body),

  logout: (): Promise<void> =>
    apiClient.post<void>('/api/auth/logout', {}),

  me: (): Promise<LoginResponse['user']> =>
    apiClient.get('/api/auth/me'),
}
