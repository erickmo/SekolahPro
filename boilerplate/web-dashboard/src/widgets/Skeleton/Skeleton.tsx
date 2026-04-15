import styles from './Skeleton.module.css'

interface SkeletonProps {
  width?: string | number
  height?: string | number
  borderRadius?: string
  className?: string
}

/** Single skeleton block */
export function Skeleton({ width, height, borderRadius, className }: SkeletonProps) {
  return (
    <span
      className={`${styles.skeleton} ${className ?? ''}`}
      style={{
        width: width ?? '100%',
        height: height ?? '1em',
        borderRadius,
        display: 'block',
      }}
      aria-hidden="true"
    />
  )
}

interface SkeletonTextProps {
  lines?: number
  lastLineWidth?: string
}

/** Multiple lines of text skeleton */
export function SkeletonText({ lines = 3, lastLineWidth = '60%' }: SkeletonTextProps) {
  return (
    <div className={styles.textGroup}>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton
          key={i}
          width={i === lines - 1 ? lastLineWidth : '100%'}
          height="0.85em"
        />
      ))}
    </div>
  )
}

interface SkeletonCardProps {
  hasAvatar?: boolean
  lines?: number
}

/** Preset: card with optional avatar + text lines */
export function SkeletonCard({ hasAvatar = false, lines = 3 }: SkeletonCardProps) {
  return (
    <div className={styles.card}>
      {hasAvatar && <Skeleton width={40} height={40} borderRadius="50%" />}
      <div className={styles.cardBody}>
        <Skeleton width="55%" height="0.9em" />
        <SkeletonText lines={lines} />
      </div>
    </div>
  )
}

interface SkeletonTableProps {
  rows?: number
  cols?: number
}

/** Preset: data table skeleton */
export function SkeletonTable({ rows = 5, cols = 4 }: SkeletonTableProps) {
  return (
    <div className={styles.table}>
      {/* Header */}
      <div className={styles.row}>
        {Array.from({ length: cols }).map((_, i) => (
          <Skeleton key={i} height="0.75em" width="70%" />
        ))}
      </div>
      {/* Body rows */}
      {Array.from({ length: rows }).map((_, r) => (
        <div key={r} className={styles.row}>
          {Array.from({ length: cols }).map((_, c) => (
            <Skeleton key={c} height="0.8em" width={c === 0 ? '85%' : '60%'} />
          ))}
        </div>
      ))}
    </div>
  )
}
