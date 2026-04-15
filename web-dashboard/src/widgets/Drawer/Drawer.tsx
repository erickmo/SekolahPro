import { useEffect } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'
import { AnimatePresence, motion } from 'framer-motion'
import styles from './Drawer.module.css'

type DrawerPlacement = 'right' | 'left'

interface DrawerProps {
  isOpen: boolean
  onClose: () => void
  title?: string
  children: React.ReactNode
  footer?: React.ReactNode
  width?: number | string
  placement?: DrawerPlacement
  closeOnOverlay?: boolean
}

const SLIDE: Record<DrawerPlacement, { x: string | number }> = {
  right: { x: '100%' },
  left: { x: '-100%' },
}

export function Drawer({
  isOpen,
  onClose,
  title,
  children,
  footer,
  width = 420,
  placement = 'right',
  closeOnOverlay = true,
}: DrawerProps) {
  // Lock body scroll when open
  useEffect(() => {
    if (isOpen) document.body.style.overflow = 'hidden'
    else document.body.style.overflow = ''
    return () => { document.body.style.overflow = '' }
  }, [isOpen])

  // Close on Escape
  useEffect(() => {
    if (!isOpen) return
    const handle = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', handle)
    return () => document.removeEventListener('keydown', handle)
  }, [isOpen, onClose])

  return createPortal(
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Overlay */}
          <motion.div
            className={styles.overlay}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            onClick={closeOnOverlay ? onClose : undefined}
            aria-hidden="true"
          />

          {/* Panel */}
          <motion.div
            className={`${styles.panel} ${styles[placement]}`}
            style={{ width }}
            initial={SLIDE[placement]}
            animate={{ x: 0 }}
            exit={SLIDE[placement]}
            transition={{ type: 'spring', stiffness: 380, damping: 36 }}
            role="dialog"
            aria-modal="true"
            aria-label={title}
          >
            {/* Header */}
            <div className={styles.header}>
              {title && <h2 className={styles.title}>{title}</h2>}
              <button
                type="button"
                className={styles.closeBtn}
                onClick={onClose}
                aria-label="Tutup"
              >
                <X size={18} />
              </button>
            </div>

            {/* Body */}
            <div className={styles.body}>{children}</div>

            {/* Footer */}
            {footer && <div className={styles.footer}>{footer}</div>}
          </motion.div>
        </>
      )}
    </AnimatePresence>,
    document.body,
  )
}
