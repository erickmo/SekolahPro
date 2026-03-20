import axios, { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001/api/v1';

export const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000,
});

// ─── Request Interceptor ──────────────────────────────────────────────────────
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('eds_access_token');
      if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`;
      }

      // Multi-tenant: attach school subdomain header
      const subdomain = localStorage.getItem('eds_subdomain');
      if (subdomain && config.headers) {
        config.headers['X-School-Subdomain'] = subdomain;
      }
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// ─── Response Interceptor ─────────────────────────────────────────────────────
api.interceptors.response.use(
  (response: AxiosResponse) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = localStorage.getItem('eds_refresh_token');
        if (!refreshToken) {
          throw new Error('No refresh token');
        }

        const response = await axios.post(`${API_URL}/auth/refresh`, {
          refreshToken,
        });

        const { accessToken, refreshToken: newRefreshToken } = response.data.data;
        localStorage.setItem('eds_access_token', accessToken);
        localStorage.setItem('eds_refresh_token', newRefreshToken);

        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${accessToken}`;
        }

        return api(originalRequest);
      } catch {
        // Refresh failed — clear auth and redirect to login
        localStorage.removeItem('eds_access_token');
        localStorage.removeItem('eds_refresh_token');
        localStorage.removeItem('eds_user');
        localStorage.removeItem('eds_subdomain');
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    }

    return Promise.reject(error);
  }
);

// ─── Typed API helpers ────────────────────────────────────────────────────────

export function extractData<T>(response: AxiosResponse): T {
  return response.data.data as T;
}

export function extractPaginated<T>(response: AxiosResponse): {
  data: T;
  meta: { total: number; page: number; limit: number; totalPages: number };
} {
  return {
    data: response.data.data as T,
    meta: response.data.meta,
  };
}

export default api;
