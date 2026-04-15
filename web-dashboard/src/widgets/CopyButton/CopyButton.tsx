import { useState } from 'react'
import { Copy, Check } from 'lucide-react'
import styles from './CopyButton.module.css'

interface CopyButtonProps {
  value: string
  label?: string
  successLabel?: string
  size?: 'sm' | 'md'
  variant?: 'icon' | 'text' | 'inline'
}

const RESET_DELAY_MS = 2000

export function CopyButton({
  value,
  label = 'Salin',
  successLabel = 'Tersalin!',
  size = 'md',
  variant = 'icon',
}: CopyButtonProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    if (copied) return
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      setTimeout(() => setCopied(false), RESET_DELAY_MS)
    } catch {
      // Fallback for older browsers
      const el = document.createElement('textarea')
      el.value = value
      document.body.appendChild(el)
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
      setCopied(true)
      setTimeout(() => setCopied(false), RESET_DELAY_MS)
    }
  }

  return (
    <button
      type="button"
      className={`${styles.btn} ${styles[size]} ${styles[variant]} ${copied ? styles.copied : ''}`}
      onClick={handleCopy}
      aria-label={copied ? successLabel : label}
      title={copied ? successLabel : label}
    >
      {copied ? <Check size={size === 'sm' ? 12 : 14} /> : <Copy size={size === 'sm' ? 12 : 14} />}
      {variant !== 'icon' && (
        <span>{copied ? successLabel : label}</span>
      )}
    </button>
  )
}

// ─── Inline variant: shows value + copy button side by side ──────────────────

interface CopyFieldProps {
  value: string
  label?: string
  truncate?: boolean
}

export function CopyField({ value, label, truncate = true }: CopyFieldProps) {
  return (
    <div className={styles.field}>
      {label && <span className={styles.fieldLabel}>{label}</span>}
      <div className={styles.fieldRow}>
        <code className={`${styles.fieldValue} ${truncate ? styles.truncate : ''}`} title={value}>
          {value}
        </code>
        <CopyButton value={value} size="sm" />
      </div>
    </div>
  )
}
