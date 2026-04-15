import type { LucideIcon } from 'lucide-react'
import { SearchX } from 'lucide-react'
import styles from './EmptyState.module.css'

interface EmptyStateProps {
  icon?: LucideIcon
  title?: string
  description?: string
  action?: React.ReactNode
}

export function EmptyState({
  icon: Icon = SearchX,
  title = 'Tidak ada data',
  description = 'Belum ada data yang tersedia.',
  action,
}: EmptyStateProps) {
  return (
    <div className={styles.root}>
      <div className={styles.iconWrap}>
        <Icon size={40} />
      </div>
      <p className={styles.title}>{title}</p>
      <p className={styles.desc}>{description}</p>
      {action && <div className={styles.action}>{action}</div>}
    </div>
  )
}
