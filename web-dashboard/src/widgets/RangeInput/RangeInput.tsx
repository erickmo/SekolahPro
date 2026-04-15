import { useCallback } from 'react'
import styles from './RangeInput.module.css'

interface RangeInputProps {
  min: number
  max: number
  step?: number
  value: [number, number]
  onChange: (value: [number, number]) => void
  label?: string
  formatValue?: (v: number) => string
  disabled?: boolean
}

const DEFAULT_FORMAT = (v: number) => String(v)

export function RangeInput({
  min,
  max,
  step = 1,
  value,
  onChange,
  label,
  formatValue = DEFAULT_FORMAT,
  disabled = false,
}: RangeInputProps) {
  const [low, high] = value
  const range = max - min

  const lowPct = range === 0 ? 0 : ((low - min) / range) * 100
  const highPct = range === 0 ? 100 : ((high - min) / range) * 100

  const handleLow = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const next = Math.min(Number(e.target.value), high - step)
    onChange([next, high])
  }, [high, step, onChange])

  const handleHigh = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const next = Math.max(Number(e.target.value), low + step)
    onChange([low, next])
  }, [low, step, onChange])

  return (
    <div className={`${styles.wrapper} ${disabled ? styles.disabled : ''}`}>
      {label && <span className={styles.label}>{label}</span>}

      <div className={styles.values}>
        <span className={styles.value}>{formatValue(low)}</span>
        <span className={styles.valueSep}>–</span>
        <span className={styles.value}>{formatValue(high)}</span>
      </div>

      <div className={styles.track}>
        {/* Filled range */}
        <div
          className={styles.fill}
          style={{ left: `${lowPct}%`, width: `${highPct - lowPct}%` }}
          aria-hidden="true"
        />

        {/* Low thumb */}
        <input
          type="range"
          className={`${styles.thumb} ${styles.thumbLow}`}
          min={min}
          max={max}
          step={step}
          value={low}
          onChange={handleLow}
          disabled={disabled}
          aria-label={label ? `${label} — nilai minimum` : 'Nilai minimum'}
        />

        {/* High thumb */}
        <input
          type="range"
          className={`${styles.thumb} ${styles.thumbHigh}`}
          min={min}
          max={max}
          step={step}
          value={high}
          onChange={handleHigh}
          disabled={disabled}
          aria-label={label ? `${label} — nilai maksimum` : 'Nilai maksimum'}
        />
      </div>

      <div className={styles.limits}>
        <span>{formatValue(min)}</span>
        <span>{formatValue(max)}</span>
      </div>
    </div>
  )
}
