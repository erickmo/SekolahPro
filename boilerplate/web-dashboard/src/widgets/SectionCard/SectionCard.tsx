import { useState } from 'react'
import { ChevronDown } from 'lucide-react'
import { AnimatePresence, motion } from 'framer-motion'
import styles from './SectionCard.module.css'

interface SectionCardProps {
  title: string
  description?: string
  children: React.ReactNode
  headerRight?: React.ReactNode
  /** Makes the section collapsible */
  collapsible?: boolean
  /** Initial expanded state (when collapsible) */
  defaultExpanded?: boolean
  /** Controlled expanded state */
  expanded?: boolean
  onExpandChange?: (expanded: boolean) => void
  noPadding?: boolean
  className?: string
}

export function SectionCard({
  title,
  description,
  children,
  headerRight,
  collapsible = false,
  defaultExpanded = true,
  expanded: controlledExpanded,
  onExpandChange,
  noPadding = false,
  className,
}: SectionCardProps) {
  const [internalExpanded, setInternalExpanded] = useState(defaultExpanded)

  const isControlled = controlledExpanded !== undefined
  const isExpanded = isControlled ? controlledExpanded : internalExpanded

  const toggle = () => {
    if (!collapsible) return
    const next = !isExpanded
    if (!isControlled) setInternalExpanded(next)
    onExpandChange?.(next)
  }

  return (
    <section className={`${styles.card} ${className ?? ''}`}>
      <div
        className={`${styles.header} ${collapsible ? styles.clickable : ''}`}
        onClick={collapsible ? toggle : undefined}
        role={collapsible ? 'button' : undefined}
        aria-expanded={collapsible ? isExpanded : undefined}
        tabIndex={collapsible ? 0 : undefined}
        onKeyDown={collapsible ? (e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggle() } } : undefined}
      >
        <div className={styles.headerLeft}>
          <h3 className={styles.title}>{title}</h3>
          {description && <p className={styles.description}>{description}</p>}
        </div>

        <div className={styles.headerRight} onClick={(e) => e.stopPropagation()}>
          {headerRight}
          {collapsible && (
            <ChevronDown
              size={16}
              className={`${styles.chevron} ${isExpanded ? styles.chevronUp : ''}`}
              aria-hidden="true"
            />
          )}
        </div>
      </div>

      <AnimatePresence initial={false}>
        {isExpanded && (
          <motion.div
            key="body"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2, ease: [0.4, 0, 0.2, 1] }}
            style={{ overflow: 'hidden' }}
          >
            <div className={`${styles.body} ${noPadding ? styles.noPadding : ''}`}>
              {children}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </section>
  )
}
