// ─── Application Configuration ───────────────────────────────────────────────
// Set VITE_MULTI_TENANT=true in .env.local to enable multi-tenant mode.
// Set VITE_MULTI_TENANT=false (or omit) for single-tenant mode.

export const appConfig = {
  /**
   * When true, enables multi-tenant routing:
   * - Superuser context: /su/*
   * - HQ/Group context: /g/*
   * - Company context: /c/:companyCode/*
   *
   * When false, uses flat single-tenant routing: /dashboard, /settings, etc.
   */
  isMultiTenant: import.meta.env.VITE_MULTI_TENANT === 'true',

  /**
   * App display name — used in Navbar logo and page title.
   */
  appName: import.meta.env.VITE_APP_NAME ?? 'Dashboard',
} as const
