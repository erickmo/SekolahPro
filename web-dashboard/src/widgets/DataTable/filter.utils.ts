import type { ActiveFilter, FilterDef, FilterFieldType, FilterOperator, FilterTuple } from './filter.types'

const DEFAULT_OPERATORS: Record<FilterFieldType, FilterOperator[]> = {
  text: ['like', '=', '!='],
  number: ['=', '>', '>=', '<', '<=', 'between'],
  select: ['=', '!=', 'in'],
  'select-async': ['=', '!=', 'in'],
  date: ['=', '>=', '<=', '>', '<', 'between'],
  'date-range': ['between'],
  boolean: ['='],
}

export const OPERATOR_LABELS: Record<FilterOperator, string> = {
  '=': 'sama dengan',
  '!=': '≠ tidak sama',
  'like': 'mengandung',
  'in': 'dalam list',
  '>': 'lebih dari',
  '>=': '≥ lebih atau sama',
  '<': 'kurang dari',
  '<=': '≤ kurang atau sama',
  'between': 'antara',
  'is': 'adalah (null)',
}

export function getOperatorsForDef(def: FilterDef): FilterOperator[] {
  return def.operators ?? DEFAULT_OPERATORS[def.type] ?? ['=']
}

export function buildNewFilter(def: FilterDef, id: string): ActiveFilter {
  const ops = getOperatorsForDef(def)
  const defaultVal = def.type === 'date-range' || ops[0] === 'between' ? ['', ''] : ''
  return { id, key: def.key, operator: ops[0], value: defaultVal }
}

function isComplete(f: ActiveFilter): boolean {
  const v = f.value
  if (v === null || v === undefined || v === '') return false
  if (typeof v === 'boolean') return true
  if (Array.isArray(v)) return (v as string[]).some((x) => x !== '')
  if (typeof v === 'object' && 'value' in (v as object)) {
    return !!(v as { value: string }).value
  }
  return true
}

function serializeOne(f: ActiveFilter): Record<string, unknown> {
  if (!isComplete(f)) return {}

  if (Array.isArray(f.value)) {
    const [from, to] = f.value as [unknown, unknown]
    const r: Record<string, unknown> = {}
    if (from !== undefined && from !== '') r[`${f.key}_from`] = from
    if (to !== undefined && to !== '') r[`${f.key}_to`] = to
    return r
  }

  const suffixMap: Partial<Record<FilterOperator, string>> = {
    '!=': '_neq',
    'like': '_contains',
    '>': '_gt',
    '>=': '_gte',
    '<': '_lt',
    '<=': '_lte',
  }
  const paramKey = `${f.key}${suffixMap[f.operator] ?? ''}`

  // select-async: value is { value, label } — serialize only the id
  const raw = f.value
  const serialized =
    typeof raw === 'object' && raw !== null && 'value' in (raw as object)
      ? (raw as { value: unknown }).value
      : raw

  return { [paramKey]: serialized }
}

/**
 * Format lama: spread ke query params (status=active, name_contains=John).
 * Dipertahankan untuk backward compatibility — handler lama tetap menerima format ini.
 */
export function serializeFiltersLegacy(filters: ActiveFilter[]): Record<string, unknown> {
  return filters.reduce<Record<string, unknown>>((acc, f) => ({ ...acc, ...serializeOne(f) }), {})
}

/**
 * Format baru: JSON array [["field","operator","value"], ...]
 * Dikirim sebagai ?filters=JSON.stringify(array)
 *
 * Contoh: [["status","=","active"],["name","like","John"]]
 */
export function serializeFilters(filters: ActiveFilter[]): Record<string, unknown> {
  const complete = filters.filter((f) => isComplete(f))
  if (complete.length === 0) return {}

  const arr = complete.map((f) => {
    // select-async: ambil .value dari object {value, label}
    const raw = f.value
    const value =
      typeof raw === 'object' && raw !== null && !Array.isArray(raw) && 'value' in (raw as object)
        ? (raw as { value: unknown }).value
        : raw
    return [f.key, f.operator, value]
  })

  // Kirim hanya format baru (legacy params dihapus)
  return {
    filters: JSON.stringify(arr),
  }
}

/**
 * Parse filter dari URL query string ?filters=[["field","op","value"],...]
 * Mengembalikan { matched, adhoc } untuk dipakai sebagai initial state.
 * - matched: filters yang punya definisi di filterDefs
 * - adhoc: filters yang tidak punya definisi (legacy atau custom)
 */
export function parseFiltersFromURL(
  searchParams: URLSearchParams,
  defs?: FilterDef[],
): { matched: ActiveFilter[]; adhoc: ActiveFilter[] } {
  const raw = searchParams.get('filters')
  if (!raw || !defs) return { matched: [], adhoc: [] }

  try {
    const arr = JSON.parse(raw) as unknown[][]
    if (!Array.isArray(arr)) return { matched: [], adhoc: [] }

    const filters = arr
      .filter((item) => Array.isArray(item) && item.length === 3)
      .map((item, i) => {
        const [field, operator, value] = item as [string, string, unknown]
        return {
          id: `url_${i}`,
          key: String(field),
          operator: operator as FilterOperator,
          value,
        }
      })

    const matched = filters.filter((f) => defs.some((d) => d.key === f.key))
    const adhoc = filters.filter((f) => !defs.some((d) => d.key === f.key))

    return { matched, adhoc }
  } catch {
    return { matched: [], adhoc: [] }
  }
}

/**
 * Build URL query string dari FilterTuple array.
 * Contoh: buildFilterQueryString('master-data/items', [['status', '=', 'active']])
 * Hasil: 'master-data/items?filters=[["status","=","active"]]'
 */
export function buildFilterQueryString(path: string, filters: FilterTuple[]): string {
  if (filters.length === 0) return path

  const arr = filters.map((f) => {
    const [key, operator, value] = f
    return [key, operator, value]
  })

  const queryString = `filters=${encodeURIComponent(JSON.stringify(arr))}`
  return `${path}?${queryString}`
}
