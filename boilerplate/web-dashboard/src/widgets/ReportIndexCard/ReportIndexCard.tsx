import { Link } from 'react-router-dom'
import styles from './ReportIndexCard.module.css'

// ─── ReportIndexCard ──────────────────────────────────────────────────────────

interface ReportIndexCardProps {
  to: string
  icon: React.ReactNode
  iconColor: string
  iconBg: string
  title: string
  description: string
}

export function ReportIndexCard({
  to,
  icon,
  iconColor,
  iconBg,
  title,
  description,
}: ReportIndexCardProps) {
  return (
    <Link to={to} className={styles.card}>
      <div className={styles.iconWrap} style={{ background: iconBg, color: iconColor }}>
        {icon}
      </div>
      <div className={styles.body}>
        <span className={styles.title}>{title}</span>
        <span className={styles.desc}>{description}</span>
      </div>
    </Link>
  )
}

// ─── ReportGroup ──────────────────────────────────────────────────────────────

interface ReportGroupProps {
  label: string
  children: React.ReactNode
}

export function ReportGroup({ label, children }: ReportGroupProps) {
  return (
    <div className={styles.group}>
      <h3 className={styles.groupLabel}>{label}</h3>
      <div className={styles.grid}>
        {children}
      </div>
    </div>
  )
}
