import { useState } from 'react'
import { ChevronDown } from 'lucide-react'
import { AnimatePresence, motion } from 'framer-motion'
import styles from './Accordion.module.css'

type AccordionVariant = 'default' | 'bordered' | 'filled'

interface AccordionItem {
  key: string
  label: string
  icon?: React.ReactNode
  content: React.ReactNode
  disabled?: boolean
}

interface AccordionProps {
  items: AccordionItem[]
  defaultOpenKeys?: string[]
  multiple?: boolean
  variant?: AccordionVariant
}

export function Accordion({
  items,
  defaultOpenKeys = [],
  multiple = false,
  variant = 'default',
}: AccordionProps) {
  const [openKeys, setOpenKeys] = useState<string[]>(defaultOpenKeys)

  const toggle = (key: string) => {
    setOpenKeys((prev) => {
      const isOpen = prev.includes(key)
      if (isOpen) return prev.filter((k) => k !== key)
      if (multiple) return [...prev, key]
      return [key]
    })
  }

  return (
    <div className={`${styles.root} ${styles[variant]}`}>
      {items.map((item) => {
        const isOpen = openKeys.includes(item.key)
        const isDisabled = item.disabled === true

        return (
          <div
            key={item.key}
            className={`${styles.item} ${isDisabled ? styles.itemDisabled : ''}`}
          >
            <button
              type="button"
              className={styles.trigger}
              onClick={() => { if (!isDisabled) toggle(item.key) }}
              aria-expanded={isOpen}
              aria-disabled={isDisabled}
              disabled={isDisabled}
            >
              {item.icon && (
                <span className={styles.icon} aria-hidden="true">
                  {item.icon}
                </span>
              )}
              <span className={styles.label}>{item.label}</span>
              <span
                className={styles.chevron}
                aria-hidden="true"
                data-open={isOpen}
              >
                <ChevronDown size={16} />
              </span>
            </button>

            <AnimatePresence initial={false}>
              {isOpen && (
                <motion.div
                  className={styles.contentWrapper}
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: 'auto', opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
                  style={{ overflow: 'hidden' }}
                >
                  <div className={styles.content}>{item.content}</div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        )
      })}
    </div>
  )
}
