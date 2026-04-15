import styles from './Switch.module.css'

export type SwitchSize = 'sm' | 'md' | 'lg'

interface SwitchProps {
  checked: boolean
  onChange: (checked: boolean) => void
  label?: string
  description?: string
  size?: SwitchSize
  disabled?: boolean
  id?: string
}

export function Switch({
  checked,
  onChange,
  label,
  description,
  size = 'md',
  disabled = false,
  id,
}: SwitchProps) {
  const inputId = id ?? `switch-${Math.random().toString(36).slice(2)}`

  return (
    <label
      htmlFor={inputId}
      className={`${styles.wrapper} ${disabled ? styles.disabled : ''}`}
    >
      <span
        className={`${styles.track} ${styles[size]} ${checked ? styles.on : styles.off}`}
        aria-hidden="true"
      >
        <span className={styles.thumb} />
      </span>

      <input
        id={inputId}
        type="checkbox"
        role="switch"
        className={styles.hiddenInput}
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
        aria-checked={checked}
      />

      {(label || description) && (
        <span className={styles.labelGroup}>
          {label && <span className={styles.label}>{label}</span>}
          {description && <span className={styles.description}>{description}</span>}
        </span>
      )}
    </label>
  )
}
