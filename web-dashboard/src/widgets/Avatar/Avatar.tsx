import { useState } from 'react'
import { getInitials } from '@/utils/format'
import styles from './Avatar.module.css'

export type AvatarSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl'

const SIZE_PX: Record<AvatarSize, number> = {
  xs: 20, sm: 28, md: 36, lg: 48, xl: 64,
}

// Deterministic color from name string
const PALETTE = [
  '#4D2975', '#E9A800', '#26B8B0', '#2563EB',
  '#16A34A', '#DC2626', '#D97706', '#7C3AED',
]
function nameColor(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return PALETTE[Math.abs(hash) % PALETTE.length]
}

interface AvatarProps {
  name: string
  src?: string
  size?: AvatarSize
  shape?: 'circle' | 'square'
  className?: string
}

export function Avatar({ name, src, size = 'md', shape = 'circle', className }: AvatarProps) {
  const [imgError, setImgError] = useState(false)
  const px = SIZE_PX[size]
  const initials = getInitials(name)
  const bg = nameColor(name)
  const showImg = !!src && !imgError

  return (
    <span
      className={`${styles.avatar} ${styles[size]} ${styles[shape]} ${className ?? ''}`}
      style={{ width: px, height: px, background: showImg ? undefined : bg }}
      title={name}
      aria-label={name}
    >
      {showImg ? (
        <img
          src={src}
          alt={name}
          className={styles.img}
          onError={() => setImgError(true)}
        />
      ) : (
        <span className={styles.initials}>{initials}</span>
      )}
    </span>
  )
}

// ─── AvatarGroup ──────────────────────────────────────────────────────────────

interface AvatarGroupItem {
  name: string
  src?: string
}

interface AvatarGroupProps {
  items: AvatarGroupItem[]
  max?: number
  size?: AvatarSize
}

export function AvatarGroup({ items, max = 4, size = 'sm' }: AvatarGroupProps) {
  const shown = items.slice(0, max)
  const rest = items.length - max

  return (
    <div className={styles.group}>
      {shown.map((item, i) => (
        <Avatar
          key={i}
          name={item.name}
          src={item.src}
          size={size}
          className={styles.groupItem}
        />
      ))}
      {rest > 0 && (
        <span className={`${styles.avatar} ${styles[size]} ${styles.circle} ${styles.groupItem} ${styles.overflow}`}>
          <span className={styles.initials}>+{rest}</span>
        </span>
      )}
    </div>
  )
}
