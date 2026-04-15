import styles from './FormPageTemplate.module.css'

interface FieldRowProps {
  children: React.ReactNode
}

export function FieldRow({ children }: FieldRowProps) {
  return <div className={styles.fieldRow}>{children}</div>
}
