import styles from './FormPageTemplate.module.css'

interface CheckboxOptionProps {
  checked: boolean
  onChange: (checked: boolean) => void
  title: string
  description?: string
  disabled?: boolean
}

export function CheckboxOption({ checked, onChange, title, description, disabled }: CheckboxOptionProps) {
  return (
    <label className={`${styles.checkboxLabel} ${disabled ? styles.checkboxDisabled : ''}`}>
      <input
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
      />
      <div>
        <span className={styles.checkboxTitle}>{title}</span>
        {description && <span className={styles.checkboxDesc}>{description}</span>}
      </div>
    </label>
  )
}
