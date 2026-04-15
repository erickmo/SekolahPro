import { SearchableSelect } from '@/widgets/SearchableSelect/SearchableSelect'
import { DatePicker } from '@/widgets/DatePicker/DatePicker'
import type { FilterDef, FilterOperator } from './filter.types'
import styles from './FilterPanel.module.css'

interface BetweenInputProps {
  inputType: 'text' | 'number' | 'date'
  value: unknown
  onChange: (v: unknown) => void
}

export function BetweenInput({ inputType, value, onChange }: BetweenInputProps) {
  const [from, to] = (value as [string, string]) ?? ['', '']
  return (
    <div className={styles.betweenWrap}>
      <input
        type={inputType}
        className={styles.input}
        value={String(from ?? '')}
        onChange={(e) => onChange([e.target.value, to])}
        placeholder="Dari"
      />
      <span className={styles.betweenSep}>–</span>
      <input
        type={inputType}
        className={styles.input}
        value={String(to ?? '')}
        onChange={(e) => onChange([from, e.target.value])}
        placeholder="Sampai"
      />
    </div>
  )
}

export interface FilterValueInputProps {
  def: FilterDef
  operator: FilterOperator
  value: unknown
  onValueChange: (v: unknown) => void
}

export function FilterValueInput({ def, operator, value, onValueChange }: FilterValueInputProps) {
  if (operator === 'between' || def.type === 'date-range') {
    const inputType = def.type === 'number' ? 'number' : 'date'
    return <BetweenInput inputType={inputType} value={value} onChange={onValueChange} />
  }
  if (def.type === 'boolean') {
    return (
      <select
        className={styles.select}
        value={value === true ? 'true' : value === false ? 'false' : ''}
        onChange={(e) =>
          onValueChange(e.target.value === '' ? '' : e.target.value === 'true')
        }
      >
        <option value="">-- Pilih --</option>
        <option value="true">Ya</option>
        <option value="false">Tidak</option>
      </select>
    )
  }
  if (def.type === 'select') {
    const valueStr = String(value ?? '')
    const hasValueInOptions = (def.options ?? []).some((opt) => opt.value === valueStr)
    // Jika nilai dari URL tidak ada di opsi, tampilkan sebagai text input
    if (valueStr && !hasValueInOptions) {
      return (
        <input
          type="text"
          className={styles.input}
          value={valueStr}
          onChange={(e) => onValueChange(e.target.value)}
          placeholder="Nilai kustom..."
          title="Nilai ini tidak ada di opsi filter yang tersedia"
        />
      )
    }
    return (
      <select
        className={styles.select}
        value={valueStr}
        onChange={(e) => onValueChange(e.target.value)}
      >
        <option value="">-- Pilih --</option>
        {(def.options ?? []).map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    )
  }
  if (def.type === 'select-async') {
    if (!def.fetchOptions) return null
    const asyncVal = value as { value: string; label: string } | undefined
    return (
      <div className={styles.asyncSelectWrap}>
        <SearchableSelect
          value={asyncVal?.value ?? ''}
          displayLabel={asyncVal?.label ?? ''}
          placeholder="Pilih..."
          fetchOptions={def.fetchOptions}
          onSelect={(opt) =>
            onValueChange(opt ? { value: opt.value, label: opt.label } : undefined)
          }
        />
      </div>
    )
  }
  if (def.type === 'date') {
    return (
      <DatePicker
        className={styles.input}
        value={String(value ?? '')}
        onChange={onValueChange}
      />
    )
  }
  if (def.type === 'number') {
    return (
      <input
        type="number"
        className={styles.input}
        value={String(value ?? '')}
        onChange={(e) => onValueChange(e.target.value)}
        placeholder="Nilai..."
      />
    )
  }
  return (
    <input
      type="text"
      className={styles.input}
      value={String(value ?? '')}
      onChange={(e) => onValueChange(e.target.value)}
      placeholder="Ketik nilai..."
    />
  )
}
