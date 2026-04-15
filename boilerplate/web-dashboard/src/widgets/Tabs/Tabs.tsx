import { useRef } from 'react'
import { motion } from 'framer-motion'
import styles from './Tabs.module.css'

export interface TabItem {
  key: string
  label: string
  icon?: React.ReactNode
  badge?: number
  disabled?: boolean
}

interface TabsProps {
  items: TabItem[]
  activeKey: string
  onChange: (key: string) => void
  variant?: 'line' | 'pill'
  size?: 'sm' | 'md'
}

export function Tabs({ items, activeKey, onChange, variant = 'line', size = 'md' }: TabsProps) {
  const listRef = useRef<HTMLDivElement>(null)

  return (
    <div
      ref={listRef}
      className={`${styles.tabs} ${styles[variant]} ${styles[size]}`}
      role="tablist"
    >
      {items.map((tab) => {
        const isActive = tab.key === activeKey
        return (
          <button
            key={tab.key}
            role="tab"
            aria-selected={isActive}
            aria-disabled={tab.disabled}
            disabled={tab.disabled}
            className={`${styles.tab} ${isActive ? styles.active : ''}`}
            onClick={() => !tab.disabled && onChange(tab.key)}
          >
            {tab.icon && <span className={styles.icon}>{tab.icon}</span>}
            <span>{tab.label}</span>
            {tab.badge !== undefined && tab.badge > 0 && (
              <span className={styles.badge}>{tab.badge > 99 ? '99+' : tab.badge}</span>
            )}
            {isActive && variant === 'line' && (
              <motion.span
                className={styles.indicator}
                layoutId="tabs-indicator"
                transition={{ type: 'spring', stiffness: 500, damping: 40 }}
              />
            )}
          </button>
        )
      })}
    </div>
  )
}
