import { Info } from 'lucide-react'
import { create } from 'zustand'
import styles from './InfoAlertDialog.module.css'

// ─── Store ────────────────────────────────────────────────────────────────────

interface ModalState {
  isOpen: boolean
  title: string
  message: string
  open: (title: string, message: string) => void
  close: () => void
}

export const useInfoAlertDialogStore = create<ModalState>((set) => ({
  isOpen: false,
  title: '',
  message: '',
  open: (title, message) => set({ isOpen: true, title, message }),
  close: () => set({ isOpen: false }),
}))

// ─── Hook ─────────────────────────────────────────────────────────────────────

export function useInfoAlertDialog() {
  const open = useInfoAlertDialogStore((s) => s.open)
  return (title: string, message: string) => {
    open(title, message)
  }
}

// ─── Modal Component ──────────────────────────────────────────────────────────

function InfoAlertDialogContent() {
  const { isOpen, title, message, close } = useInfoAlertDialogStore()

  if (!isOpen) return null

  return (
    <div className={styles.overlay} onClick={close}>
      <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
        <div className={styles.icon}>
          <Info size={24} />
        </div>
        <h3 className={styles.title}>{title}</h3>
        <p className={styles.message}>{message}</p>
        <div className={styles.actions}>
          <button className={styles.btnPrimary} onClick={close}>
            Mengerti
          </button>
        </div>
      </div>
    </div>
  )
}

export function InfoAlertDialogProvider() {
  return <InfoAlertDialogContent />
}
