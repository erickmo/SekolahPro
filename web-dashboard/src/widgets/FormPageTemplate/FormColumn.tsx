import styles from './FormPageTemplate.module.css'

interface FormColumnProps {
  children: React.ReactNode
}

export function FormColumn({ children }: FormColumnProps) {
  return <div className={styles.column}>{children}</div>
}
