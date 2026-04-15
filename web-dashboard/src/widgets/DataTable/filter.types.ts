import type { SelectOption } from '@/widgets/SearchableSelect/SearchableSelect'

export type FilterOperator =
  | '='
  | '!='
  | 'like'
  | 'in'
  | '>'
  | '>='
  | '<'
  | '<='
  | 'between'
  | 'is'

export type FilterFieldType =
  | 'text'
  | 'number'
  | 'select'
  | 'select-async'
  | 'date'
  | 'date-range'
  | 'boolean'

export interface FilterDef {
  key: string
  label: string
  type: FilterFieldType
  /** Override default operators for this field type. */
  operators?: FilterOperator[]
  /** Static options — required when type is 'select'. */
  options?: { label: string; value: string }[]
  /** Async options fetch — required when type is 'select-async'. */
  fetchOptions?: (search: string) => Promise<SelectOption[]>
}

/**
 * Value shape per type:
 * - text | number | date | select: string
 * - boolean: boolean
 * - between (number/date) or date-range: [from: string, to: string]
 * - select-async: { value: string; label: string }
 */
export interface ActiveFilter {
  /** Stable unique row id (used as React key). */
  id: string
  key: string
  operator: FilterOperator
  value: unknown
}

/**
 * Tuple representation of a filter for query string serialization.
 * Format: [field, operator, value]
 */
export type FilterTuple = [string, FilterOperator, unknown]
