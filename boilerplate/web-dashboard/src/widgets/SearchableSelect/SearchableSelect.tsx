import { useState, useRef, useEffect, useCallback } from 'react'
import { createPortal } from 'react-dom'
import { Search, X, ChevronDown, Loader2 } from 'lucide-react'
import styles from './SearchableSelect.module.css'

export interface SelectOption {
  /** ID yang disimpan ke form state */
  value: string
  /** Label utama yang ditampilkan */
  label: string
  /** Info sekunder (kode, UoM, dll.) */
  meta?: string
  /** Data tambahan untuk keperluan form (code, uom, dll.) */
  data?: Record<string, unknown>
}

interface Props {
  /** ID item yang sedang dipilih */
  value: string
  /** Teks yang ditampilkan di trigger (biasanya nama item terpilih) */
  displayLabel: string
  placeholder?: string
  error?: string
  disabled?: boolean
  /** Fungsi fetch opsi — dipanggil dengan string pencarian */
  fetchOptions: (search: string) => Promise<SelectOption[]>
  onSelect: (option: SelectOption | null) => void
}

const DEBOUNCE_MS = 300

export function SearchableSelect({
  value,
  displayLabel,
  placeholder = 'Pilih...',
  error,
  disabled,
  fetchOptions,
  onSelect,
}: Props) {
  const [isOpen, setIsOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [options, setOptions] = useState<SelectOption[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [dropdownStyle, setDropdownStyle] = useState<React.CSSProperties>({})
  const containerRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const searchRef = useRef<HTMLInputElement>(null)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Tutup dropdown saat klik di luar
  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      const target = e.target as Node
      const inContainer = containerRef.current?.contains(target) ?? false
      const inDropdown = dropdownRef.current?.contains(target) ?? false
      if (!inContainer && !inDropdown) setIsOpen(false)
    }
    document.addEventListener('mousedown', onClickOutside)
    return () => document.removeEventListener('mousedown', onClickOutside)
  }, [])

  // Hitung posisi dropdown menggunakan fixed positioning agar tidak terpotong overflow
  useEffect(() => {
    if (!isOpen || !triggerRef.current) return

    function updatePosition() {
      if (!triggerRef.current) return
      const rect = triggerRef.current.getBoundingClientRect()
      setDropdownStyle({
        position: 'fixed',
        top: rect.bottom + 4,
        left: rect.left,
        width: rect.width,
      })
    }

    updatePosition()
    window.addEventListener('scroll', updatePosition, true)
    window.addEventListener('resize', updatePosition)
    return () => {
      window.removeEventListener('scroll', updatePosition, true)
      window.removeEventListener('resize', updatePosition)
    }
  }, [isOpen])

  // Tutup dropdown saat tekan Escape
  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') setIsOpen(false)
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [])

  const doFetch = useCallback(async (q: string) => {
    setIsLoading(true)
    try {
      const results = await fetchOptions(q)
      setOptions(results)
    } catch {
      setOptions([])
    } finally {
      setIsLoading(false)
    }
  }, [fetchOptions])

  // Fetch awal saat dropdown dibuka
  useEffect(() => {
    if (isOpen) {
      searchRef.current?.focus({ preventScroll: true })
      doFetch('')
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen])

  function handleSearchChange(e: React.ChangeEvent<HTMLInputElement>) {
    const q = e.target.value
    setSearch(q)
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => doFetch(q), DEBOUNCE_MS)
  }

  function handleSelect(option: SelectOption) {
    onSelect(option)
    setIsOpen(false)
    setSearch('')
  }

  function handleClear(e: React.MouseEvent) {
    e.stopPropagation()
    onSelect(null)
  }

  function handleTriggerClick() {
    if (!disabled) setIsOpen((prev) => !prev)
  }

  const hasValue = !!value

  return (
    <div
      ref={containerRef}
      className={[
        styles.container,
        error ? styles.hasError : '',
        disabled ? styles.disabled : '',
      ].join(' ')}
    >
      {/* Trigger button */}
      <button
        ref={triggerRef}
        type="button"
        className={`${styles.trigger} ${isOpen ? styles.triggerOpen : ''}`}
        onClick={handleTriggerClick}
        disabled={disabled}
      >
        <span className={`${styles.triggerLabel} ${hasValue ? '' : styles.triggerPlaceholder}`}>
          {hasValue ? displayLabel : placeholder}
        </span>
        <span className={styles.triggerIcons}>
          {hasValue && !disabled && (
            <span
              className={styles.clearBtn}
              onClick={handleClear}
              role="button"
              aria-label="Hapus pilihan"
            >
              <X size={12} />
            </span>
          )}
          <ChevronDown size={14} className={`${styles.chevron} ${isOpen ? styles.chevronOpen : ''}`} />
        </span>
      </button>

      {/* Dropdown — dirender ke document.body via portal agar tidak terpotong overflow parent */}
      {isOpen && createPortal(
        <div ref={dropdownRef} className={styles.dropdown} style={dropdownStyle}>
          <div className={styles.searchWrap}>
            <Search size={13} className={styles.searchIcon} />
            <input
              ref={searchRef}
              type="text"
              className={styles.searchInput}
              placeholder="Ketik untuk mencari..."
              value={search}
              onChange={handleSearchChange}
            />
          </div>
          <div className={styles.optionsList}>
            {isLoading ? (
              <div className={styles.statusRow}>
                <Loader2 size={14} className={styles.spinner} />
                <span>Memuat...</span>
              </div>
            ) : options.length === 0 ? (
              <div className={styles.statusRow}>Tidak ada hasil</div>
            ) : (
              options.map((opt) => (
                <button
                  key={opt.value}
                  type="button"
                  className={`${styles.option} ${opt.value === value ? styles.optionSelected : ''}`}
                  onClick={() => handleSelect(opt)}
                >
                  <span className={styles.optionLabel}>{opt.label}</span>
                  {opt.meta && <span className={styles.optionMeta}>{opt.meta}</span>}
                </button>
              ))
            )}
          </div>
        </div>,
        document.body,
      )}
    </div>
  )
}
