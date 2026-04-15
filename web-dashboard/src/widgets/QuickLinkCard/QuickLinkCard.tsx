import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import styles from './QuickLinkCard.module.css'

interface QuickLinkCardProps {
  href: string
  icon: ReactNode
  iconBg: string
  iconColor: string
  label: string
  desc: string
}

export function QuickLinkCard({ href, icon, iconBg, iconColor, label, desc }: QuickLinkCardProps) {
  return (
    <Link to={href} className={styles.card}>
      <span className={styles.icon} style={{ background: iconBg, color: iconColor }}>
        {icon}
      </span>
      <span className={styles.text}>
        <span className={styles.label}>{label}</span>
        <span className={styles.desc}>{desc}</span>
      </span>
    </Link>
  )
}
