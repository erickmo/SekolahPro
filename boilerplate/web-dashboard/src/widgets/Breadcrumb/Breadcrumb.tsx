import { ChevronRight, Home } from 'lucide-react'
import { Link } from 'react-router-dom'
import styles from './Breadcrumb.module.css'

export interface BreadcrumbItem {
  label: string
  href?: string
}

interface BreadcrumbProps {
  items: BreadcrumbItem[]
  showHome?: boolean
  homeHref?: string
}

export function Breadcrumb({ items, showHome = true, homeHref = '/' }: BreadcrumbProps) {
  const all: BreadcrumbItem[] = showHome
    ? [{ label: 'Beranda', href: homeHref }, ...items]
    : items

  return (
    <nav aria-label="Breadcrumb">
      <ol className={styles.breadcrumb}>
        {all.map((item, index) => {
          const isLast = index === all.length - 1
          const isFirst = index === 0 && showHome

          return (
            <li key={index} className={styles.item}>
              {index > 0 && (
                <ChevronRight size={12} className={styles.separator} aria-hidden="true" />
              )}

              {isLast || !item.href ? (
                <span
                  className={`${styles.label} ${isLast ? styles.current : ''}`}
                  aria-current={isLast ? 'page' : undefined}
                >
                  {isFirst ? <Home size={13} aria-hidden="true" /> : item.label}
                  {isFirst && <span className={styles.srOnly}>{item.label}</span>}
                </span>
              ) : (
                <Link
                  to={item.href}
                  className={`${styles.label} ${styles.link} ${isFirst ? styles.homeLink : ''}`}
                >
                  {isFirst ? <Home size={13} aria-hidden="true" /> : item.label}
                  {isFirst && <span className={styles.srOnly}>{item.label}</span>}
                </Link>
              )}
            </li>
          )
        })}
      </ol>
    </nav>
  )
}
