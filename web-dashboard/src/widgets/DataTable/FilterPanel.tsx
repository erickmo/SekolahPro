import { Plus, X } from 'lucide-react'
import type { ActiveFilter, FilterDef, FilterOperator } from './filter.types'
import { OPERATOR_LABELS, buildNewFilter, getOperatorsForDef } from './filter.utils'
import { FilterValueInput } from './FilterValueInput'
import { cn } from '@/utils/cn'
import styles from './FilterPanel.module.css'

// ─── Filter row ───────────────────────────────────────────────────────────────

function FilterRow({
  filter,
  defs,
  onChange,
  onRemove,
}: {
  filter: ActiveFilter
  defs: FilterDef[]
  onChange: (f: ActiveFilter) => void
  onRemove: () => void
}) {
  const def = defs.find((d) => d.key === filter.key)
  // Untuk filter tanpa definisi (adhoc dari URL), gunakan label dari key field
  const isAdhoc = !def
  // Default operators untuk filter tanpa definisi (gunakan versi string: '=' bukan 'eq')
  const operators = def ? getOperatorsForDef(def) : (['=', 'like', '!='] as FilterOperator[])

  function handleFieldChange(newKey: string) {
    const newDef = defs.find((d) => d.key === newKey)
    if (!newDef) return
    const ops = getOperatorsForDef(newDef)
    const defaultVal = newDef.type === 'date-range' || ops[0] === 'between' ? ['', ''] : ''
    onChange({ ...filter, key: newKey, operator: ops[0], value: defaultVal })
  }

  function handleOperatorChange(newOp: FilterOperator) {
    const isBetween = newOp === 'between'
    const wasBetween = filter.operator === 'between'
    const newVal = isBetween ? ['', ''] : wasBetween ? '' : filter.value
    onChange({ ...filter, operator: newOp, value: newVal })
  }

  return (
    <div className={styles.row}>
      <div className={styles.rowInputs}>
        <select
          className={styles.fieldSelect}
          value={filter.key}
          onChange={(e) => handleFieldChange(e.target.value)}
        >
          {defs.map((d) => (
            <option key={d.key} value={d.key}>
              {d.label}
            </option>
          ))}
          {/* Untuk filter tanpa definisi, sertakan key sebagai opsi */}
          {isAdhoc && (
            <option key={filter.key} value={filter.key}>
              {filter.key} (dari URL)
            </option>
          )}
        </select>
        <select
          className={styles.operatorSelect}
          value={filter.operator}
          onChange={(e) => handleOperatorChange(e.target.value as FilterOperator)}
        >
          {operators.map((op) => (
            <option key={op} value={op}>
              {OPERATOR_LABELS[op]}
            </option>
          ))}
        </select>
        {def ? (
          <FilterValueInput
            def={def}
            operator={filter.operator}
            value={filter.value}
            onValueChange={(v) => onChange({ ...filter, value: v })}
          />
        ) : (
          /* Untuk filter tanpa definisi, gunakan text input sederhana */
          <input
            type="text"
            className={styles.input}
            value={String(filter.value ?? '')}
            onChange={(e) => onChange({ ...filter, value: e.target.value })}
            placeholder="Nilai..."
          />
        )}
      </div>
      <button className={styles.removeBtn} onClick={onRemove} title="Hapus filter">
        <X size={13} />
      </button>
    </div>
  )
}

// ─── Panel ────────────────────────────────────────────────────────────────────

interface FilterPanelProps {
  defs: FilterDef[]
  activeFilters: ActiveFilter[]
  onChange: (filters: ActiveFilter[]) => void
  className?: string
}

export function FilterPanel({ defs, activeFilters, onChange, className }: FilterPanelProps) {
  function addFilter() {
    if (defs.length === 0) return
    const id = Math.random().toString(36).slice(2)
    onChange([...activeFilters, buildNewFilter(defs[0], id)])
  }

  function removeFilter(id: string) {
    onChange(activeFilters.filter((f) => f.id !== id))
  }

  function updateFilter(updated: ActiveFilter) {
    onChange(activeFilters.map((f) => (f.id === updated.id ? updated : f)))
  }

  return (
    <div className={cn(styles.panel, className)}>
      <div className={styles.panelHeader}>
        <span className={styles.panelTitle}>Filter</span>
        {activeFilters.length > 0 && (
          <button className={styles.clearAllBtn} onClick={() => onChange([])}>
            Hapus semua
          </button>
        )}
      </div>
      {activeFilters.map((f) => (
        <FilterRow
          key={f.id}
          filter={f}
          defs={defs}
          onChange={updateFilter}
          onRemove={() => removeFilter(f.id)}
        />
      ))}
      <button className={styles.addBtn} onClick={addFilter}>
        <Plus size={13} />
        Tambah filter
      </button>
    </div>
  )
}
