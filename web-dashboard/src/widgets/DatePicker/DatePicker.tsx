import { useState, useRef, useEffect } from 'react'
import { Calendar, ChevronLeft, ChevronRight } from 'lucide-react'
import styles from './DatePicker.module.css'

// ─── Constants ────────────────────────────────────────────────────────────────

const MONTHS_ID = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
]

const DAYS_SHORT = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']

// ─── Types ────────────────────────────────────────────────────────────────────

export interface DatePickerProps {
  value: string                        // YYYY-MM-DD or empty string
  onChange: (value: string) => void
  className?: string                   // passed through to trigger (e.g. inputError)
  placeholder?: string
  min?: string                         // YYYY-MM-DD
  max?: string                         // YYYY-MM-DD
  disabled?: boolean
  id?: string
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function parseISO(value: string): Date | null {
  if (!value || value.length < 10) return null
  const [y, m, d] = value.split('-').map(Number)
  if (!y || !m || !d) return null
  return new Date(y, m - 1, d)
}

function toISO(year: number, month: number, day: number): string {
  return `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`
}

function formatDisplay(value: string): string {
  const d = parseISO(value)
  if (!d) return ''
  return new Intl.DateTimeFormat('id-ID', {
    day: '2-digit', month: 'short', year: 'numeric',
  }).format(d)
}

function daysInMonth(year: number, month: number): number {
  return new Date(year, month + 1, 0).getDate()
}

function firstDayOfMonth(year: number, month: number): number {
  return new Date(year, month, 1).getDay() // 0 = Sunday
}

type Cell = { year: number; month: number; day: number; current: boolean }

function buildCalendarCells(year: number, month: number): Cell[] {
  const cells: Cell[] = []
  const firstDay = firstDayOfMonth(year, month)
  const prevMonth = month === 0 ? 11 : month - 1
  const prevYear = month === 0 ? year - 1 : year
  const prevDays = daysInMonth(prevYear, prevMonth)

  for (let i = firstDay - 1; i >= 0; i--) {
    cells.push({ year: prevYear, month: prevMonth, day: prevDays - i, current: false })
  }

  const current = daysInMonth(year, month)
  for (let d = 1; d <= current; d++) {
    cells.push({ year, month, day: d, current: true })
  }

  const nextMonth = month === 11 ? 0 : month + 1
  const nextYear = month === 11 ? year + 1 : year
  for (let d = 1; cells.length < 42; d++) {
    cells.push({ year: nextYear, month: nextMonth, day: d, current: false })
  }

  return cells
}

// ─── Component ────────────────────────────────────────────────────────────────

export function DatePicker({
  value,
  onChange,
  className,
  placeholder = 'Pilih tanggal',
  min,
  max,
  disabled,
  id,
}: DatePickerProps) {
  const today = new Date()
  const todayISO = toISO(today.getFullYear(), today.getMonth(), today.getDate())

  const [open, setOpen] = useState(false)
  const [view, setView] = useState<{ year: number; month: number }>(() => {
    const d = parseISO(value) ?? today
    return { year: d.getFullYear(), month: d.getMonth() }
  })
  const wrapperRef = useRef<HTMLDivElement>(null)

  // Sync calendar view when controlled value changes
  useEffect(() => {
    const d = parseISO(value)
    if (d) setView({ year: d.getFullYear(), month: d.getMonth() })
  }, [value])

  // Close on outside click
  useEffect(() => {
    if (!open) return
    function handle(e: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handle)
    return () => document.removeEventListener('mousedown', handle)
  }, [open])

  const minDate = parseISO(min ?? '')
  const maxDate = parseISO(max ?? '')

  function isOutOfRange(year: number, month: number, day: number): boolean {
    const d = new Date(year, month, day)
    if (minDate && d < minDate) return true
    if (maxDate && d > maxDate) return true
    return false
  }

  function prevMonth() {
    setView(v => v.month === 0
      ? { year: v.year - 1, month: 11 }
      : { ...v, month: v.month - 1 })
  }

  function nextMonth() {
    setView(v => v.month === 11
      ? { year: v.year + 1, month: 0 }
      : { ...v, month: v.month + 1 })
  }

  function selectDay(cell: Cell) {
    if (isOutOfRange(cell.year, cell.month, cell.day)) return
    onChange(toISO(cell.year, cell.month, cell.day))
    setOpen(false)
  }

  const cells = buildCalendarCells(view.year, view.month)

  return (
    <div className={styles.wrapper} ref={wrapperRef}>
      <button
        id={id}
        type="button"
        className={`${styles.trigger} ${className ?? ''}`}
        onClick={() => !disabled && setOpen(o => !o)}
        disabled={disabled}
      >
        <span className={value ? '' : styles.placeholderText}>
          {value ? formatDisplay(value) : placeholder}
        </span>
        <Calendar size={14} className={styles.calIcon} />
      </button>

      {open && (
        <div className={styles.popover}>
          {/* Month navigation */}
          <div className={styles.header}>
            <button type="button" className={styles.navBtn} onClick={prevMonth}>
              <ChevronLeft size={14} />
            </button>
            <span className={styles.monthYear}>
              {MONTHS_ID[view.month]} {view.year}
            </span>
            <button type="button" className={styles.navBtn} onClick={nextMonth}>
              <ChevronRight size={14} />
            </button>
          </div>

          {/* Day-of-week headers */}
          <div className={styles.weekRow}>
            {DAYS_SHORT.map(d => (
              <span key={d} className={styles.weekDay}>{d}</span>
            ))}
          </div>

          {/* Day grid */}
          <div className={styles.dayGrid}>
            {cells.map((cell, i) => {
              const iso = toISO(cell.year, cell.month, cell.day)
              const selected = iso === value
              const isToday = iso === todayISO
              const outOfRange = isOutOfRange(cell.year, cell.month, cell.day)
              return (
                <button
                  key={i}
                  type="button"
                  className={[
                    styles.day,
                    !cell.current && styles.dayOther,
                    selected && styles.daySelected,
                    isToday && !selected && styles.dayToday,
                    outOfRange && styles.dayDisabled,
                  ].filter(Boolean).join(' ')}
                  onClick={() => selectDay(cell)}
                  disabled={outOfRange}
                  tabIndex={outOfRange ? -1 : 0}
                >
                  {cell.day}
                </button>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}
