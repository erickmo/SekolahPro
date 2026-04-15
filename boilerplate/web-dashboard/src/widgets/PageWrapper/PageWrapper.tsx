import { AlertCircle, RefreshCw } from 'lucide-react'
import styles from './PageWrapper.module.css'

interface PageWrapperProps {
  children: React.ReactNode
  isLoading?: boolean
  error?: Error | null
  onRetry?: () => void
  className?: string
}

function PageSkeleton() {
  return (
    <div className={styles.skeleton}>
      {[1, 2, 3].map((i) => (
        <div key={i} className={styles.skeletonBlock} style={{ animationDelay: `${i * 0.1}s` }} />
      ))}
    </div>
  )
}

export function PageWrapper({ children, isLoading, error, onRetry, className }: PageWrapperProps) {
  if (isLoading) return <PageSkeleton />

  if (error) {
    return (
      <div className={styles.error}>
        <AlertCircle size={40} className={styles.errorIcon} />
        <h3 className={styles.errorTitle}>Gagal memuat data</h3>
        <p className={styles.errorMessage}>{error.message}</p>
        {onRetry && (
          <button className={`${styles.retryBtn}`} onClick={onRetry}>
            <RefreshCw size={14} />
            Coba lagi
          </button>
        )}
      </div>
    )
  }

  return (
    <div className={`${styles.root} ${className ?? ''}`}>
      {children}
    </div>
  )
}
