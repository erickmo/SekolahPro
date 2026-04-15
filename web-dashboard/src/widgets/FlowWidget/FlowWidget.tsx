import { useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { ArrowRight, HelpCircle, X } from 'lucide-react'
import styles from './FlowWidget.module.css'

export interface FlowStep {
  id: string
  label: string
  count: number
  href?: string
  urgent?: boolean
  urgentCount?: number
  description?: string
}

interface FlowWidgetProps {
  title: string
  subtitle?: string
  icon: React.ReactNode
  accentColor: string
  steps: FlowStep[]
  onInfoClick?: (step: FlowStep) => void
}

interface InfoDialogProps {
  step: FlowStep
  accentColor: string
  onClose: () => void
}

export function InfoDialog({ step, accentColor, onClose }: InfoDialogProps) {
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    },
    [onClose],
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  return (
    <div
      className={styles.backdrop}
      role="presentation"
      onClick={onClose}
    >
      <div
        className={styles.dialog}
        role="dialog"
        aria-modal="true"
        aria-label={`Info: ${step.label}`}
        onClick={(e) => e.stopPropagation()}
      >
        <div className={styles.dialogHeader} style={{ borderLeftColor: accentColor }}>
          <div className={styles.dialogTitleGroup}>
            <span
              className={styles.dialogIcon}
              style={{ background: `${accentColor}18`, color: accentColor }}
              aria-hidden="true"
            >
              <HelpCircle size={16} />
            </span>
            <h4 className={styles.dialogTitle}>{step.label}</h4>
          </div>
          <button
            className={styles.dialogClose}
            onClick={onClose}
            aria-label="Tutup"
          >
            <X size={16} />
          </button>
        </div>
        <p className={styles.dialogBody}>
          {step.description ?? 'Tidak ada keterangan tersedia untuk langkah ini.'}
        </p>
        {step.count > 0 && (
          <div className={styles.dialogFooter} style={{ borderTopColor: `${accentColor}22` }}>
            <span className={styles.dialogCount} style={{ color: accentColor }}>
              {step.count.toLocaleString('id-ID')}
            </span>
            <span className={styles.dialogCountLabel}>item aktif saat ini</span>
          </div>
        )}
      </div>
    </div>
  )
}

export function FlowWidget({ title, subtitle, icon, accentColor, steps, onInfoClick }: FlowWidgetProps) {
  const handleInfoClick = useCallback((step: FlowStep, e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    onInfoClick?.(step)
  }, [onInfoClick])

  return (
    <div className={styles.card} role="region" aria-label={title}>
      <div className={styles.header}>
        <span
          className={styles.iconWrap}
          style={{ background: `${accentColor}18`, color: accentColor }}
          aria-hidden="true"
        >
          {icon}
        </span>
        <div className={styles.titleGroup}>
          <h3 className={styles.title}>{title}</h3>
          {subtitle && <p className={styles.subtitle}>{subtitle}</p>}
        </div>
      </div>

      <div className={styles.flow} role="list">
        {steps.map((step, idx) => {
          const isLast = idx === steps.length - 1
          const cssVars = { '--accent': accentColor } as React.CSSProperties

          const stepBox = (
            <div
              className={`${styles.step} ${step.urgent ? styles.stepUrgent : ''}`}
              style={cssVars}
              role="listitem"
            >
              <span className={styles.accentBar} aria-hidden="true" />
              {step.urgent && step.urgentCount !== undefined && (
                <span
                  className={styles.urgentBadge}
                  aria-label={`${step.urgentCount} item urgent`}
                  title={`${step.urgentCount} item membutuhkan perhatian`}
                >
                  {step.urgentCount}
                </span>
              )}
              <span className={styles.stepCount} aria-label={`${step.count} item`}>
                {step.count.toLocaleString('id-ID')}
              </span>
              <span className={styles.stepLabel}>{step.label}</span>
              <button
                className={styles.helpBtn}
                onClick={(e) => handleInfoClick(step, e)}
                aria-label={`Info tentang ${step.label}`}
                title={`Apa itu ${step.label}?`}
                style={cssVars}
              >
                <HelpCircle size={10} />
              </button>
            </div>
          )

          return (
            <div key={step.id} className={styles.stepSlot}>
              {step.href ? (
                <Link
                  to={step.href}
                  className={styles.stepLink}
                  aria-label={`${step.label}: ${step.count} item${step.urgent ? `, ${step.urgentCount} urgent` : ''}`}
                >
                  {stepBox}
                </Link>
              ) : (
                stepBox
              )}
              {!isLast && (
                <span className={styles.arrow} aria-hidden="true">
                  <ArrowRight size={13} />
                </span>
              )}
            </div>
          )
        })}
      </div>

    </div>
  )
}
