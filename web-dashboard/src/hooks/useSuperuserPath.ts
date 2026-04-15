/**
 * Returns a helper to build paths scoped to the Superuser context.
 * Usage: const path = useSuperuserPath(); navigate(path('tenants'))  →  /su/tenants
 */
export function useSuperuserPath() {
  const base = '/su'
  return (segment: string) => `${base}/${segment}`
}
