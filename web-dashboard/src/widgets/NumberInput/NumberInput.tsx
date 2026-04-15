import { useCallback } from 'react'
import { Minus, Plus } from 'lucide-react'
import styles from './NumberInput.module.css'

interface NumberInputProps {
  value: number | ''
  onChange: (value: number | '') => void
  min?: number
  max?: number
  step?: number
  placeholder?: string
  disabled?: boolean
  error?: string
  label?: string
  suffix?: string
  prefix?: string
  id?: string
}

export function NumberInput({
  value,
  onChange,
  min,
  max,
  step = 1,
  placeholder = '0',
  disabled = false,
  error,
  label,
  suffix,
  prefix,
  id,
}: NumberInputProps) {
  const inputId = id ?? `number-${Math.random().toString(36).slice(2)}`
  const num = value === '' ? NaN : Number(value)

  const clamp = useCallback((v: number): number => {
    if (min !== undefined && v < min) return min
    if (max !== undefined && v > max) return max
    return v
  }, [min, max])

  const increment = () => {
    const next = isNaN(num) ? (min ?? 0) + step : clamp(num + step)
    onChange(next)
  }

  const decrement = () => {
    const next = isNaN(num) ? (min ?? 0) : clamp(num - step)
    onChange(next)
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value
    if (raw === '' || raw === '-') { onChange(''); return }
    const parsed = parseFloat(raw)
    if (!isNaN(parsed)) onChange(parsed)
  }

  const handleBlur = () => {
    if (!isNaN(num)) onChange(clamp(num))
  }

  const canDecrement = !disabled && (isNaN(num) || min === undefined || num > min)
  const canIncrement = !disabled && (isNaN(num) || max === undefined || num < max)

  return (
    <div className={styles.wrapper}>
      {label && <label htmlFor={inputId} className={styles.label}>{label}</label>}

      <div className={`${styles.inputRow} ${error ? styles.hasError : ''} ${disabled ? styles.disabled : ''}`}>
        <button
          type="button"
          className={styles.stepBtn}
          onClick={decrement}
          disabled={!canDecrement}
          tabIndex={-1}
          aria-label="Kurangi"
        >
          <Minus size={12} />
        </button>

        {prefix && <span className={styles.affix}>{prefix}</span>}

        <input
          id={inputId}
          type="number"
          className={styles.input}
          value={value}
          onChange={handleChange}
          onBlur={handleBlur}
          placeholder={placeholder}
          disabled={disabled}
          min={min}
          max={max}
          step={step}
          aria-invalid={!!error}
          aria-describedby={error ? `${inputId}-error` : undefined}
        />

        {suffix && <span className={styles.affix}>{suffix}</span>}

        <button
          type="button"
          className={styles.stepBtn}
          onClick={increment}
          disabled={!canIncrement}
          tabIndex={-1}
          aria-label="Tambah"
        >
          <Plus size={12} />
        </button>
      </div>

      {error && (
        <p id={`${inputId}-error`} className={styles.error}>{error}</p>
      )}
    </div>
  )
}
