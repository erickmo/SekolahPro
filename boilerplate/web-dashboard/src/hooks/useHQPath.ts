/**
 * Returns a helper to build paths scoped to the HQ/Group context.
 * Usage: const path = useHQPath(); navigate(path('reports'))  →  /g/reports
 */
export function useHQPath() {
  const base = '/g'
  return (segment: string) => `${base}/${segment}`
}
