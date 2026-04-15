import styles from './FormPageTemplate.module.css'

export interface RadioOption {
  value: string
  label: string
}

interface RadioGroupProps {
  name: string
  value: string
  onChange: (value: string) => void
  options: RadioOption[]
}

export function RadioGroup({ name, value, onChange, options }: RadioGroupProps) {
  return (
    <div className={styles.radioGroup}>
      {options.map((opt) => (
        <label key={opt.value} className={styles.radioLabel}>
          <input
            type="radio"
            name={name}
            value={opt.value}
            checked={value === opt.value}
            onChange={() => onChange(opt.value)}
          />
          {opt.label}
        </label>
      ))}
    </div>
  )
}
