import { useEffect } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'
import { AnimatePresence, motion } from 'framer-motion'
import styles from './Modal.module.css'

type ModalSize = 'sm' | 'md' | 'lg' | 'xl'

interface ModalProps {
  isOpen: boolean
  onClose: () => void
  title?: string
  children: React.ReactNode
  footer?: React.ReactNode
  size?: ModalSize
  closeOnOverlay?: boolean
  hideCloseButton?: boolean
}

const SIZE_WIDTH: Record<ModalSize, string> = {
  sm: '400px',
  md: '560px',
  lg: '720px',
  xl: '900px',
}

export function Modal({
  isOpen,
  onClose,
  title,
  children,
  footer,
  size = 'md',
  closeOnOverlay = true,
  hideCloseButton = false,
}: ModalProps) {
  useEffect(() => {
    if (isOpen) document.body.style.overflow = 'hidden'
    else document.body.style.overflow = ''
    return () => { document.body.style.overflow = '' }
  }, [isOpen])

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
          <motion.div
            className={styles.overlay}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            onClick={closeOnOverlay ? onClose : undefined}
            aria-hidden="true"
          />

          <motion.div
            className={styles.panel}
            style={{ width: SIZE_WIDTH[size] }}
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
            role="dialog"
            aria-modal="true"
            aria-label={title}
          >
            <div className={styles.header}>
              {title && <h2 className={styles.title}>{title}</h2>}
              {!hideCloseButton && (
                <button
                  type="button"
                  className={styles.closeBtn}
                  onClick={onClose}
                  aria-label="Tutup"
                >
                  <X size={18} />
                </button>
              )}
            </div>

            <div className={styles.body}>{children}</div>

            {footer && <div className={styles.footer}>{footer}</div>}
          </motion.div>
        </>
      )}
    </AnimatePresence>,
    document.body,
  )
}
