import styles from './ChartCard.module.css'

interface ChartCardProps {
  title: string
  subtitle?: string
  children: React.ReactNode
  actions?: React.ReactNode
  className?: string
}

export function ChartCard({ title, subtitle, children, actions, className }: ChartCardProps) {
  return (
    <div className={`${styles.card} ${className ?? ''}`}>
      <div className={styles.header}>
        <div className={styles.titleGroup}>
          <h3 className={styles.title}>{title}</h3>
          {subtitle && <p className={styles.subtitle}>{subtitle}</p>}
        </div>
        {actions && <div className={styles.actions}>{actions}</div>}
      </div>
      <div className={styles.body}>{children}</div>
    </div>
  )
}
