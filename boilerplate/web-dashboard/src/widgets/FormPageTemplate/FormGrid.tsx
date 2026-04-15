import styles from './FormPageTemplate.module.css'

interface FormGridProps {
  children: React.ReactNode
}

export function FormGrid({ children }: FormGridProps) {
  return <div className={styles.formGrid}>{children}</div>
}
