import styles from './Timeline.module.css'

export interface TimelineEvent {
  id: string
  title: string
  description?: string
  timestamp: string
  icon?: React.ReactNode
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info'
  actor?: string
}

interface TimelineProps {
  events: TimelineEvent[]
  emptyText?: string
}

export function Timeline({ events, emptyText = 'Belum ada aktivitas.' }: TimelineProps) {
  if (events.length === 0) {
    return <p className={styles.empty}>{emptyText}</p>
  }

  return (
    <ol className={styles.timeline} aria-label="Riwayat aktivitas">
      {events.map((event, index) => {
        const isLast = index === events.length - 1
        const variant = event.variant ?? 'default'

        return (
          <li key={event.id} className={`${styles.item} ${styles[variant]}`}>
            {/* Connector line */}
            {!isLast && <div className={styles.connector} aria-hidden="true" />}

            {/* Dot / Icon */}
            <div className={styles.dotWrap} aria-hidden="true">
              {event.icon ? (
                <div className={styles.iconDot}>{event.icon}</div>
              ) : (
                <div className={styles.dot} />
              )}
            </div>

            {/* Content */}
            <div className={styles.content}>
              <div className={styles.header}>
                <span className={styles.title}>{event.title}</span>
                <time className={styles.timestamp} dateTime={event.timestamp}>
                  {event.timestamp}
                </time>
              </div>
              {event.description && (
                <p className={styles.description}>{event.description}</p>
              )}
              {event.actor && (
                <span className={styles.actor}>oleh {event.actor}</span>
              )}
            </div>
          </li>
        )
      })}
    </ol>
  )
}
