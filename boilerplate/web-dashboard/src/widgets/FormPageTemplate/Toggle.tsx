import styles from './FormPageTemplate.module.css'

interface ToggleProps {
  checked: boolean
  onChange: (checked: boolean) => void
  label?: string
  disabled?: boolean
  id?: string
}

export function Toggle({ checked, onChange, label, disabled, id }: ToggleProps) {
  return (
    <label
      className={[
        styles.toggle,
        checked ? styles.toggleChecked : '',
        disabled ? styles.toggleDisabled : '',
      ].join(' ')}
    >
      <input
        type="checkbox"
        id={id}
        className={styles.toggleInput}
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
      />
      <span className={styles.toggleTrack}>
        <span className={styles.toggleThumb} />
      </span>
      {label && <span className={styles.toggleLabel}>{label}</span>}
    </label>
  )
}
