import styles from './FormPageTemplate.module.css'

interface FieldSectionProps {
  title: string
  children: React.ReactNode
}

export function FieldSection({ title, children }: FieldSectionProps) {
  return (
    <div className={styles.card}>
      <h2 className={styles.cardTitle}>{title}</h2>
      {children}
    </div>
  )
}
