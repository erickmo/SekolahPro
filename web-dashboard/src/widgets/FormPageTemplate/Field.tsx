import styles from './FormPageTemplate.module.css'

interface FieldProps {
  label: string
  required?: boolean
  error?: string
  hint?: string
  children: React.ReactNode
  className?: string
}

export function Field({ label, required, error, hint, children, className }: FieldProps) {
  return (
    <div className={`${styles.field} ${className ?? ''}`}>
      <label className={styles.fieldLabel}>
        {label}
        {required && <span className={styles.required}> *</span>}
      </label>
      {children}
      {hint && !error && <span className={styles.hint}>{hint}</span>}
      {error && <span className={styles.error}>{error}</span>}
    </div>
  )
}
