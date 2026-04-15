import { useState, useRef, useCallback } from 'react'
import { createPortal } from 'react-dom'
import { ChevronDown, X, Check, Search } from 'lucide-react'
import { useClickOutside } from '@/hooks/useClickOutside'
import styles from './MultiSelect.module.css'

export interface MultiSelectOption {
  value: string
  label: string
  meta?: string
  disabled?: boolean
}

interface MultiSelectProps {
  options: MultiSelectOption[]
  value: string[]
  onChange: (values: string[]) => void
  placeholder?: string
  searchPlaceholder?: string
  label?: string
  error?: string
  disabled?: boolean
  maxSelected?: number
  clearable?: boolean
  id?: string
}

export function MultiSelect({
  options,
  value,
  onChange,
  placeholder = 'Pilih...',
  searchPlaceholder = 'Cari...',
  label,
  error,
  disabled = false,
  maxSelected,
  clearable = true,
  id,
}: MultiSelectProps) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [dropdownPos, setDropdownPos] = useState({ top: 0, left: 0, width: 0 })
  const triggerRef = useRef<HTMLDivElement>(null)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const inputId = id ?? `multiselect-${Math.random().toString(36).slice(2)}`

  useClickOutside([triggerRef, dropdownRef], () => { setOpen(false); setSearch('') })

  const openDropdown = useCallback(() => {
    if (disabled) return
    const rect = triggerRef.current?.getBoundingClientRect()
    if (rect) {
      setDropdownPos({
        top: rect.bottom + window.scrollY + 4,
        left: rect.left + window.scrollX,
        width: rect.width,
      })
    }
    setOpen(true)
  }, [disabled])

  const toggle = (optValue: string) => {
    if (value.includes(optValue)) {
      onChange(value.filter((v) => v !== optValue))
    } else {
      if (maxSelected && value.length >= maxSelected) return
      onChange([...value, optValue])
    }
  }

  const remove = (optValue: string, e: React.MouseEvent) => {
    e.stopPropagation()
    onChange(value.filter((v) => v !== optValue))
  }

  const clear = (e: React.MouseEvent) => {
    e.stopPropagation()
    onChange([])
  }

  const filtered = options.filter((o) =>
    o.label.toLowerCase().includes(search.toLowerCase()) ||
    o.meta?.toLowerCase().includes(search.toLowerCase()),
  )

  const selectedOptions = value
    .map((v) => options.find((o) => o.value === v))
    .filter(Boolean) as MultiSelectOption[]

  return (
    <div className={styles.wrapper}>
      {label && <label id={`${inputId}-label`} className={styles.label}>{label}</label>}

      <div
        ref={triggerRef}
        className={`${styles.trigger} ${open ? styles.triggerOpen : ''} ${error ? styles.hasError : ''} ${disabled ? styles.disabled : ''}`}
        onClick={openDropdown}
        role="combobox"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-labelledby={label ? `${inputId}-label` : undefined}
        tabIndex={disabled ? -1 : 0}
        onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); openDropdown() } }}
      >
        <div className={styles.selectedWrap}>
          {selectedOptions.length === 0 ? (
            <span className={styles.placeholder}>{placeholder}</span>
          ) : (
            selectedOptions.map((opt) => (
              <span key={opt.value} className={styles.chip}>
                {opt.label}
                <button
                  type="button"
                  className={styles.chipRemove}
                  onClick={(e) => remove(opt.value, e)}
                  aria-label={`Hapus ${opt.label}`}
                >
                  <X size={10} />
                </button>
              </span>
            ))
          )}
        </div>

        <div className={styles.actions}>
          {clearable && value.length > 0 && (
            <button type="button" className={styles.clearBtn} onClick={clear} aria-label="Hapus semua">
              <X size={14} />
            </button>
          )}
          <ChevronDown size={14} className={`${styles.chevron} ${open ? styles.chevronUp : ''}`} />
        </div>
      </div>

      {error && <p className={styles.error}>{error}</p>}

      {open && createPortal(
        <div
          ref={dropdownRef}
          className={styles.dropdown}
          style={{ top: dropdownPos.top, left: dropdownPos.left, width: dropdownPos.width }}
          role="listbox"
          aria-multiselectable="true"
        >
          <div className={styles.searchWrap}>
            <Search size={12} className={styles.searchIcon} />
            <input
              type="text"
              className={styles.searchInput}
              placeholder={searchPlaceholder}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              autoFocus
              onClick={(e) => e.stopPropagation()}
            />
          </div>

          <ul className={styles.list}>
            {filtered.length === 0 ? (
              <li className={styles.empty}>Tidak ada hasil</li>
            ) : (
              filtered.map((opt) => {
                const selected = value.includes(opt.value)
                const isMaxed = !selected && maxSelected !== undefined && value.length >= maxSelected
                return (
                  <li
                    key={opt.value}
                    role="option"
                    aria-selected={selected}
                    aria-disabled={opt.disabled || isMaxed}
                    className={`${styles.option} ${selected ? styles.selected : ''} ${(opt.disabled || isMaxed) ? styles.optionDisabled : ''}`}
                    onClick={() => !opt.disabled && !isMaxed && toggle(opt.value)}
                  >
                    <span className={`${styles.checkmark} ${selected ? styles.checkmarkOn : ''}`}>
                      {selected && <Check size={10} strokeWidth={3} />}
                    </span>
                    <span className={styles.optionLabel}>{opt.label}</span>
                    {opt.meta && <span className={styles.optionMeta}>{opt.meta}</span>}
                  </li>
                )
              })
            )}
          </ul>

          {maxSelected && (
            <div className={styles.footer}>
              {value.length}/{maxSelected} dipilih
            </div>
          )}
        </div>,
        document.body,
      )}
    </div>
  )
}
