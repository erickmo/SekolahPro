import styles from './RadioGroup.module.css'

export interface RadioOption {
  value: string
  label: string
  description?: string
  disabled?: boolean
}

interface RadioGroupProps {
  options: RadioOption[]
  value: string
  onChange: (value: string) => void
  name: string
  label?: string
  error?: string
  orientation?: 'vertical' | 'horizontal'
  variant?: 'default' | 'card'
  disabled?: boolean
}

export function RadioGroup({
  options,
  value,
  onChange,
  name,
  label,
  error,
  orientation = 'vertical',
  variant = 'default',
  disabled = false,
}: RadioGroupProps) {
  return (
    <fieldset className={styles.fieldset}>
      {label && <legend className={styles.legend}>{label}</legend>}

      <div className={`${styles.group} ${styles[orientation]} ${styles[`variant_${variant}`]}`}>
        {options.map((opt) => {
          const isChecked = opt.value === value
          const isDisabled = disabled || opt.disabled

          return (
            <label
              key={opt.value}
              className={`${styles.option} ${isChecked ? styles.checked : ''} ${isDisabled ? styles.optionDisabled : ''}`}
            >
              <input
                type="radio"
                name={name}
                value={opt.value}
                checked={isChecked}
                disabled={isDisabled}
                onChange={() => onChange(opt.value)}
                className={styles.hiddenInput}
              />

              <span className={`${styles.radio} ${isChecked ? styles.radioChecked : ''}`} aria-hidden="true">
                {isChecked && <span className={styles.radioDot} />}
              </span>

              <span className={styles.labelGroup}>
                <span className={styles.optionLabel}>{opt.label}</span>
                {opt.description && (
                  <span className={styles.optionDescription}>{opt.description}</span>
                )}
              </span>
            </label>
          )
        })}
      </div>

      {error && <p className={styles.error}>{error}</p>}
    </fieldset>
  )
}
