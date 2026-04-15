import styles from './Badge.module.css'

export type BadgeVariant =
  | 'default'
  | 'primary'
  | 'success'
  | 'warning'
  | 'danger'
  | 'info'
  | 'secondary'

export type BadgeSize = 'sm' | 'md'

interface BadgeProps {
  children: React.ReactNode
  variant?: BadgeVariant
  size?: BadgeSize
  dot?: boolean
  onRemove?: () => void
  className?: string
}

export function Badge({
  children,
  variant = 'default',
  size = 'md',
  dot = false,
  onRemove,
  className,
}: BadgeProps) {
  return (
    <span className={`${styles.badge} ${styles[variant]} ${styles[size]} ${className ?? ''}`}>
      {dot && <span className={styles.dot} aria-hidden="true" />}
      {children}
      {onRemove && (
        <button
          type="button"
          className={styles.remove}
          onClick={onRemove}
          aria-label="Hapus tag"
        >
          ×
        </button>
      )}
    </span>
  )
}

// ─── Convenience presets ──────────────────────────────────────────────────────

interface StatusBadgeProps {
  status: string
  map: Record<string, BadgeVariant>
  size?: BadgeSize
}

/** Maps a status string to a variant using a provided lookup map */
export function StatusBadge({ status, map, size }: StatusBadgeProps) {
  const variant = map[status] ?? 'default'
  return <Badge variant={variant} size={size}>{status}</Badge>
}
