import { Component, ErrorInfo } from 'react'
import { AlertTriangle, RefreshCw } from 'lucide-react'
import styles from './ErrorBoundary.module.css'

interface ErrorBoundaryProps {
  children: React.ReactNode
  /** Custom fallback UI — receives error and retry fn */
  fallback?: (error: Error, retry: () => void) => React.ReactNode
  /** Called when an error is caught — useful for logging (e.g. Sentry) */
  onError?: (error: Error, info: ErrorInfo) => void
}

interface ErrorBoundaryState {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    this.props.onError?.(error, info)
    if (import.meta.env.DEV) {
      console.error('[ErrorBoundary]', error, info.componentStack)
    }
  }

  retry = () => this.setState({ hasError: false, error: null })

  render() {
    const { hasError, error } = this.state
    const { children, fallback } = this.props

    if (!hasError) return children

    if (fallback && error) return fallback(error, this.retry)

    return <DefaultFallback error={error} onRetry={this.retry} />
  }
}

// ─── Default Fallback UI ──────────────────────────────────────────────────────

interface DefaultFallbackProps {
  error: Error | null
  onRetry: () => void
}

function DefaultFallback({ error, onRetry }: DefaultFallbackProps) {
  return (
    <div className={styles.fallback} role="alert">
      <AlertTriangle size={32} className={styles.icon} />
      <h3 className={styles.title}>Terjadi Kesalahan</h3>
      <p className={styles.message}>
        {import.meta.env.DEV && error
          ? error.message
          : 'Bagian ini tidak dapat dimuat. Coba muat ulang halaman.'}
      </p>
      <button type="button" className={styles.retryBtn} onClick={onRetry}>
        <RefreshCw size={14} />
        Coba Lagi
      </button>
    </div>
  )
}

// ─── withErrorBoundary HOC ────────────────────────────────────────────────────

export function withErrorBoundary<P extends object>(
  Component: React.ComponentType<P>,
  options?: Omit<ErrorBoundaryProps, 'children'>,
) {
  const Wrapped = (props: P) => (
    <ErrorBoundary {...options}>
      <Component {...props} />
    </ErrorBoundary>
  )
  Wrapped.displayName = `withErrorBoundary(${Component.displayName ?? Component.name})`
  return Wrapped
}
