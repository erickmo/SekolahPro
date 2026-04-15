import { http, HttpResponse } from 'msw'

const BASE_URL = 'http://localhost:8080'

export const handlers = [
  // Auth
  http.post(`${BASE_URL}/api/auth/login`, () => {
    return HttpResponse.json({
      token: 'mock-token',
      refreshToken: 'mock-refresh-token',
      user: {
        id: '1',
        name: 'Test User',
        email: 'test@example.com',
        role: 'admin',
        permissions: ['*'],
      },
    })
  }),

  http.post(`${BASE_URL}/api/auth/logout`, () => {
    return new HttpResponse(null, { status: 204 })
  }),

  http.get(`${BASE_URL}/api/auth/me`, () => {
    return HttpResponse.json({
      id: '1',
      name: 'Test User',
      email: 'test@example.com',
      role: 'admin',
      permissions: ['*'],
    })
  }),
]
