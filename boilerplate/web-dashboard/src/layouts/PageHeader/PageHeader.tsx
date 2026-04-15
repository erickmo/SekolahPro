import { cn } from '@/utils/cn'
import type { Breadcrumb } from '@/types/navigation.types'
import styles from './PageHeader.module.css'

interface PageHeaderProps {
  title: string
  subtitle?: string
  breadcrumbs?: Breadcrumb[]
  actions?: React.ReactNode
  className?: string
}

export function PageHeader({ title, subtitle, breadcrumbs, actions, className }: PageHeaderProps) {
  return (
    <div className={cn(styles.root, className)}>
      {breadcrumbs && breadcrumbs.length > 0 && (
        <nav className={styles.breadcrumbs} aria-label="Breadcrumb">
          {breadcrumbs.map((crumb, i) => (
            <span key={i} className={styles.crumbItem}>
              {i > 0 && <span className={styles.separator}>/</span>}
              {crumb.href ? (
                <a href={crumb.href} className={styles.crumbLink}>{crumb.label}</a>
              ) : (
                <span className={styles.crumbCurrent}>{crumb.label}</span>
              )}
            </span>
          ))}
        </nav>
      )}
      <div className={styles.titleRow}>
        <div>
          <h1 className={styles.title}>{title}</h1>
          {subtitle && <p className={styles.subtitle}>{subtitle}</p>}
        </div>
        {actions && <div className={styles.actions}>{actions}</div>}
      </div>
    </div>
  )
}
