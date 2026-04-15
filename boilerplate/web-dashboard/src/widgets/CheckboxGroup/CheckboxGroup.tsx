import { Check, Minus } from 'lucide-react'
import styles from './CheckboxGroup.module.css'

export interface CheckboxOption {
  value: string
  label: string
  description?: string
  disabled?: boolean
}

interface CheckboxGroupProps {
  options: CheckboxOption[]
  value: string[]
  onChange: (value: string[]) => void
  label?: string
  error?: string
  orientation?: 'vertical' | 'horizontal'
  disabled?: boolean
  /** Show "select all" header checkbox */
  selectAll?: boolean
}

export function CheckboxGroup({
  options,
  value,
  onChange,
  label,
  error,
  orientation = 'vertical',
  disabled = false,
  selectAll = false,
}: CheckboxGroupProps) {
  const enabledOptions = options.filter((o) => !o.disabled)
  const allSelected = enabledOptions.every((o) => value.includes(o.value))
  const someSelected = enabledOptions.some((o) => value.includes(o.value))
  const isIndeterminate = someSelected && !allSelected

  const toggleAll = () => {
    if (allSelected) {
      onChange(value.filter((v) => !enabledOptions.some((o) => o.value === v)))
    } else {
      const toAdd = enabledOptions.map((o) => o.value).filter((v) => !value.includes(v))
      onChange([...value, ...toAdd])
    }
  }

  const toggle = (optValue: string) => {
    if (value.includes(optValue)) {
      onChange(value.filter((v) => v !== optValue))
    } else {
      onChange([...value, optValue])
    }
  }

  return (
    <fieldset className={styles.fieldset}>
      {(label || selectAll) && (
        <div className={styles.header}>
          {selectAll && (
            <CheckboxItem
              checked={allSelected}
              indeterminate={isIndeterminate}
              onChange={toggleAll}
              disabled={disabled}
              label={label ?? 'Pilih Semua'}
              bold
            />
          )}
          {label && !selectAll && <legend className={styles.legend}>{label}</legend>}
        </div>
      )}

      <div className={`${styles.group} ${styles[orientation]} ${selectAll ? styles.indented : ''}`}>
        {options.map((opt) => (
          <CheckboxItem
            key={opt.value}
            checked={value.includes(opt.value)}
            onChange={() => toggle(opt.value)}
            disabled={disabled || opt.disabled}
            label={opt.label}
            description={opt.description}
          />
        ))}
      </div>

      {error && <p className={styles.error}>{error}</p>}
    </fieldset>
  )
}

// ─── Internal CheckboxItem ────────────────────────────────────────────────────

interface CheckboxItemProps {
  checked: boolean
  indeterminate?: boolean
  onChange: () => void
  disabled?: boolean
  label: string
  description?: string
  bold?: boolean
}

function CheckboxItem({
  checked,
  indeterminate = false,
  onChange,
  disabled = false,
  label,
  description,
  bold = false,
}: CheckboxItemProps) {
  return (
    <label className={`${styles.item} ${disabled ? styles.itemDisabled : ''}`}>
      <input
        type="checkbox"
        className={styles.hiddenInput}
        checked={checked}
        onChange={onChange}
        disabled={disabled}
        ref={(el) => { if (el) el.indeterminate = indeterminate }}
      />
      <span
        className={`${styles.checkbox} ${checked || indeterminate ? styles.checkboxOn : ''}`}
        aria-hidden="true"
      >
        {indeterminate ? (
          <Minus size={10} strokeWidth={3} />
        ) : checked ? (
          <Check size={10} strokeWidth={3} />
        ) : null}
      </span>
      <span className={styles.labelGroup}>
        <span className={`${styles.itemLabel} ${bold ? styles.bold : ''}`}>{label}</span>
        {description && <span className={styles.itemDescription}>{description}</span>}
      </span>
    </label>
  )
}
