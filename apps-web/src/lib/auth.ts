import api from './api';
import type { AuthUser, LoginCredentials, AuthResponse } from '@/types';

const ACCESS_TOKEN_KEY = 'eds_access_token';
const REFRESH_TOKEN_KEY = 'eds_refresh_token';
const USER_KEY = 'eds_user';
const SUBDOMAIN_KEY = 'eds_subdomain';

// ─── Token Management ─────────────────────────────────────────────────────────

export function getAccessToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function getRefreshToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function setTokens(accessToken: string, refreshToken: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
}

export function clearTokens(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
  localStorage.removeItem(SUBDOMAIN_KEY);
}

// ─── User Management ──────────────────────────────────────────────────────────

export function getStoredUser(): AuthUser | null {
  if (typeof window === 'undefined') return null;
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as AuthUser;
  } catch {
    return null;
  }
}

export function setStoredUser(user: AuthUser): void {
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function getSubdomain(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(SUBDOMAIN_KEY);
}

// ─── Auth API Calls ───────────────────────────────────────────────────────────

export async function login(credentials: LoginCredentials): Promise<AuthResponse> {
  if (credentials.subdomain) {
    localStorage.setItem(SUBDOMAIN_KEY, credentials.subdomain);
  }

  const response = await api.post<{ success: boolean; data: AuthResponse }>(
    '/auth/login',
    {
      email: credentials.email,
      password: credentials.password,
    }
  );

  const authData = response.data.data;
  setTokens(authData.accessToken, authData.refreshToken);
  setStoredUser(authData.user);

  return authData;
}

export async function logout(): Promise<void> {
  try {
    await api.post('/auth/logout');
  } finally {
    clearTokens();
  }
}

export async function getCurrentUser(): Promise<AuthUser> {
  const response = await api.get<{ success: boolean; data: AuthUser }>('/auth/me');
  const user = response.data.data;
  setStoredUser(user);
  return user;
}

export async function changePassword(
  currentPassword: string,
  newPassword: string
): Promise<void> {
  await api.post('/auth/change-password', { currentPassword, newPassword });
}

// ─── Auth Checks ──────────────────────────────────────────────────────────────

export function isAuthenticated(): boolean {
  return !!getAccessToken() && !!getStoredUser();
}

export function hasRole(user: AuthUser | null, ...roles: string[]): boolean {
  if (!user) return false;
  return roles.includes(user.role);
}

export function canAccess(user: AuthUser | null, module: string): boolean {
  if (!user) return false;

  const accessMap: Record<string, string[]> = {
    students: [
      'ADMIN_SEKOLAH', 'OPERATOR_SIMS', 'TATA_USAHA', 'KEPALA_SEKOLAH',
      'GURU', 'WALI_KELAS', 'GURU_BK', 'BENDAHARA',
    ],
    academic: [
      'ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'KEPALA_KURIKULUM',
      'GURU', 'WALI_KELAS', 'OPERATOR_SIMS',
    ],
    payments: ['ADMIN_SEKOLAH', 'BENDAHARA', 'KEPALA_SEKOLAH'],
    library: ['ADMIN_SEKOLAH', 'PUSTAKAWAN', 'KEPALA_SEKOLAH'],
    cooperative: ['ADMIN_SEKOLAH', 'KASIR_KOPERASI', 'BENDAHARA', 'KEPALA_SEKOLAH'],
    health: ['ADMIN_SEKOLAH', 'PETUGAS_UKS', 'KEPALA_SEKOLAH'],
    counseling: ['ADMIN_SEKOLAH', 'GURU_BK', 'KEPALA_SEKOLAH'],
    dashboard: [
      'ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'KEPALA_KURIKULUM',
      'BENDAHARA', 'GURU', 'WALI_KELAS', 'GURU_BK',
      'OPERATOR_SIMS', 'KASIR_KOPERASI', 'PUSTAKAWAN',
      'PETUGAS_UKS', 'TATA_USAHA',
    ],
  };

  const allowed = accessMap[module] || [];
  return allowed.includes(user.role);
}
