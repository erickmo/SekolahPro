import { useState, useRef, useEffect, useCallback } from 'react'
import { createPortal } from 'react-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { Calendar, ChevronLeft, ChevronRight } from 'lucide-react'
import styles from './DateRangePicker.module.css'

// ─── Constants ────────────────────────────────────────────────────────────────

const MONTHS_ID = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
]

const DAYS_SHORT = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']

const DEFAULT_PLACEHOLDER = 'Pilih rentang tanggal'

// ─── Types ────────────────────────────────────────────────────────────────────

export interface DateRange {
  start: string | null
  end: string | null
}

export interface DateRangePickerProps {
  value: DateRange
  onChange: (range: DateRange) => void
  placeholder?: string
  disabled?: boolean
  error?: string
  label?: string
  minDate?: string
  maxDate?: string
}

type CalendarCell = {
  year: number
  month: number
  day: number
  current: boolean
}

type MonthView = { year: number; month: number }

// ─── Helpers ──────────────────────────────────────────────────────────────────

function parseISO(value: string | null | undefined): Date | null {
  if (!value || value.length < 10) return null
  const [y, m, d] = value.split('-').map(Number)
  if (!y || !m || !d) return null
  return new Date(y, m - 1, d)
}

function toISO(year: number, month: number, day: number): string {
  return `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`
}

function formatDisplayDate(iso: string | null): string {
  if (!iso) return ''
  const d = parseISO(iso)
  if (!d) return ''
  const dd = String(d.getDate()).padStart(2, '0')
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const yyyy = d.getFullYear()
  return `${dd}/${mm}/${yyyy}`
}

function daysInMonth(year: number, month: number): number {
  return new Date(year, month + 1, 0).getDate()
}

function buildCalendarCells(year: number, month: number): CalendarCell[] {
  const cells: CalendarCell[] = []
  const firstDay = new Date(year, month, 1).getDay()
  const prevMonth = month === 0 ? 11 : month - 1
  const prevYear = month === 0 ? year - 1 : year
  const prevDays = daysInMonth(prevYear, prevMonth)

  for (let i = firstDay - 1; i >= 0; i--) {
    cells.push({ year: prevYear, month: prevMonth, day: prevDays - i, current: false })
  }

  const count = daysInMonth(year, month)
  for (let d = 1; d <= count; d++) {
    cells.push({ year, month, day: d, current: true })
  }

  const nextMonth = month === 11 ? 0 : month + 1
  const nextYear = month === 11 ? year + 1 : year
  for (let d = 1; cells.length < 42; d++) {
    cells.push({ year: nextYear, month: nextMonth, day: d, current: false })
  }

  return cells
}

function nextMonthView(view: MonthView): MonthView {
  return view.month === 11
    ? { year: view.year + 1, month: 0 }
    : { ...view, month: view.month + 1 }
}

function prevMonthView(view: MonthView): MonthView {
  return view.month === 0
    ? { year: view.year - 1, month: 11 }
    : { ...view, month: view.month - 1 }
}

function isoFromCell(cell: CalendarCell): string {
  return toISO(cell.year, cell.month, cell.day)
}

function compareDates(a: string | null, b: string | null): number {
  if (!a || !b) return 0
  return a < b ? -1 : a > b ? 1 : 0
}

// ─── Sub-component: SingleCalendar ────────────────────────────────────────────

interface SingleCalendarProps {
  view: MonthView
  startISO: string | null
  endISO: string | null
  hoverISO: string | null
  minDate: string | null
  maxDate: string | null
  onDayClick: (iso: string) => void
  onDayHover: (iso: string | null) => void
}

function SingleCalendar({
  view,
  startISO,
  endISO,
  hoverISO,
  minDate,
  maxDate,
  onDayClick,
  onDayHover,
}: SingleCalendarProps) {
  const cells = buildCalendarCells(view.year, view.month)

  function isDisabled(iso: string): boolean {
    if (minDate && iso < minDate) return true
    if (maxDate && iso > maxDate) return true
    return false
  }

  function isSelected(iso: string): boolean {
    return iso === startISO || iso === endISO
  }

  function isRangeStart(iso: string): boolean {
    return iso === startISO && startISO !== null
  }

  function isRangeEnd(iso: string): boolean {
    return iso === endISO && endISO !== null
  }

  function isInRange(iso: string): boolean {
    const rangeEnd = endISO ?? hoverISO
    if (!startISO || !rangeEnd) return false
    const lo = compareDates(startISO, rangeEnd) <= 0 ? startISO : rangeEnd
    const hi = compareDates(startISO, rangeEnd) <= 0 ? rangeEnd : startISO
    return iso > lo && iso < hi
  }

  const today = toISO(new Date().getFullYear(), new Date().getMonth(), new Date().getDate())

  return (
    <div className={styles.calendar}>
      <div className={styles.weekRow}>
        {DAYS_SHORT.map(d => (
          <span key={d} className={styles.weekDay}>{d}</span>
        ))}
      </div>

      <div className={styles.dayGrid}>
        {cells.map((cell, i) => {
          const iso = isoFromCell(cell)
          const disabled = isDisabled(iso)
          const selected = isSelected(iso)
          const inRange = isInRange(iso)
          const isStart = isRangeStart(iso)
          const isEnd = isRangeEnd(iso)
          const isToday = iso === today

          return (
            <button
              key={i}
              type="button"
              className={[
                styles.day,
                !cell.current && styles.dayOther,
                selected && styles.daySelected,
                isStart && styles.dayStart,
                isEnd && styles.dayEnd,
                inRange && !selected && styles.dayInRange,
                isToday && !selected && styles.dayToday,
                disabled && styles.dayDisabled,
              ].filter(Boolean).join(' ')}
              onClick={() => !disabled && onDayClick(iso)}
              onMouseEnter={() => !disabled && onDayHover(iso)}
              onMouseLeave={() => onDayHover(null)}
              disabled={disabled}
              tabIndex={disabled ? -1 : 0}
              aria-label={iso}
            >
              {cell.day}
            </button>
          )
        })}
      </div>
    </div>
  )
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function DateRangePicker({
  value,
  onChange,
  placeholder = DEFAULT_PLACEHOLDER,
  disabled,
  error,
  label,
  minDate,
  maxDate,
}: DateRangePickerProps) {
  const today = new Date()
  const [open, setOpen] = useState(false)
  const [leftView, setLeftView] = useState<MonthView>(() => {
    const d = parseISO(value.start) ?? today
    return { year: d.getFullYear(), month: d.getMonth() }
  })
  const [hoverISO, setHoverISO] = useState<string | null>(null)
  const [dropdownStyle, setDropdownStyle] = useState<React.CSSProperties>({})

  const triggerRef = useRef<HTMLButtonElement>(null)
  const dropdownRef = useRef<HTMLDivElement>(null)

  const rightView = nextMonthView(leftView)
  const minISO = minDate ?? null
  const maxISO = maxDate ?? null

  const displayText =
    value.start || value.end
      ? [formatDisplayDate(value.start), formatDisplayDate(value.end)]
          .filter(Boolean)
          .join(' — ')
      : ''

  // Position dropdown below trigger
  const updatePosition = useCallback(() => {
    if (!triggerRef.current) return
    const rect = triggerRef.current.getBoundingClientRect()
    setDropdownStyle({
      position: 'fixed',
      top: rect.bottom + 4,
      left: rect.left,
      zIndex: 'var(--z-dropdown)' as unknown as number,
    })
  }, [])

  useEffect(() => {
    if (!open) return
    updatePosition()
    window.addEventListener('resize', updatePosition)
    window.addEventListener('scroll', updatePosition, true)
    return () => {
      window.removeEventListener('resize', updatePosition)
      window.removeEventListener('scroll', updatePosition, true)
    }
  }, [open, updatePosition])

  // Close on outside click or Escape
  useEffect(() => {
    if (!open) return
    function handleMouse(e: MouseEvent) {
      if (
        triggerRef.current?.contains(e.target as Node) ||
        dropdownRef.current?.contains(e.target as Node)
      ) return
      setOpen(false)
    }
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') setOpen(false)
    }
    document.addEventListener('mousedown', handleMouse)
    document.addEventListener('keydown', handleKey)
    return () => {
      document.removeEventListener('mousedown', handleMouse)
      document.removeEventListener('keydown', handleKey)
    }
  }, [open])

  function handleDayClick(iso: string) {
    const { start, end } = value

    // No selection yet, or both already set → start fresh
    if (!start || (start && end)) {
      onChange({ start: iso, end: null })
      return
    }

    // start is set, end is not → set end (or swap)
    if (iso >= start) {
      onChange({ start, end: iso })
    } else {
      onChange({ start: iso, end: start })
    }
    setOpen(false)
  }

  function handleToday() {
    const iso = toISO(today.getFullYear(), today.getMonth(), today.getDate())
    onChange({ start: iso, end: iso })
    setOpen(false)
  }

  function handleClear() {
    onChange({ start: null, end: null })
    setOpen(false)
  }

  const dropdown = (
    <AnimatePresence>
      {open && (
        <motion.div
          ref={dropdownRef}
          className={styles.dropdown}
          style={dropdownStyle}
          initial={{ opacity: 0, y: -6 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -6 }}
          transition={{ duration: 0.15, ease: 'easeOut' }}
        >
          {/* Navigation header */}
          <div className={styles.navRow}>
            <button
              type="button"
              className={styles.navBtn}
              onClick={() => setLeftView(prevMonthView)}
              aria-label="Bulan sebelumnya"
            >
              <ChevronLeft size={14} />
            </button>

            <div className={styles.monthHeaders}>
              <span className={styles.monthTitle}>
                {MONTHS_ID[leftView.month]} {leftView.year}
              </span>
              <span className={styles.monthTitle}>
                {MONTHS_ID[rightView.month]} {rightView.year}
              </span>
            </div>

            <button
              type="button"
              className={styles.navBtn}
              onClick={() => setLeftView(nextMonthView)}
              aria-label="Bulan berikutnya"
            >
              <ChevronRight size={14} />
            </button>
          </div>

          {/* Two calendars side by side */}
          <div className={styles.calendarsRow}>
            <SingleCalendar
              view={leftView}
              startISO={value.start}
              endISO={value.end}
              hoverISO={hoverISO}
              minDate={minISO}
              maxDate={maxISO}
              onDayClick={handleDayClick}
              onDayHover={setHoverISO}
            />
            <div className={styles.calendarDivider} />
            <SingleCalendar
              view={rightView}
              startISO={value.start}
              endISO={value.end}
              hoverISO={hoverISO}
              minDate={minISO}
              maxDate={maxISO}
              onDayClick={handleDayClick}
              onDayHover={setHoverISO}
            />
          </div>

          {/* Quick actions */}
          <div className={styles.actions}>
            <button type="button" className={styles.actionBtn} onClick={handleToday}>
              Hari ini
            </button>
            <button
              type="button"
              className={`${styles.actionBtn} ${styles.actionBtnDanger}`}
              onClick={handleClear}
            >
              Hapus
            </button>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  )

  return (
    <div className={styles.wrapper}>
      {label && <label className={styles.label}>{label}</label>}

      <button
        ref={triggerRef}
        type="button"
        className={[
          styles.trigger,
          error ? styles.triggerError : '',
          disabled ? styles.triggerDisabled : '',
        ].filter(Boolean).join(' ')}
        onClick={() => !disabled && setOpen(o => !o)}
        disabled={disabled}
        aria-haspopup="dialog"
        aria-expanded={open}
      >
        <span className={displayText ? styles.displayText : styles.placeholder}>
          {displayText || placeholder}
        </span>
        <Calendar size={14} className={styles.calIcon} />
      </button>

      {error && <span className={styles.errorMsg}>{error}</span>}

      {createPortal(dropdown, document.body)}
    </div>
  )
}
