import { Trash2 } from 'lucide-react'
import { create } from 'zustand'
import styles from './DeleteConfirmModal.module.css'

// ─── Store ────────────────────────────────────────────────────────────────────

interface ModalState {
  isOpen: boolean
  title: string
  message: string
  isDeleting: boolean
  onConfirm: (() => Promise<void>) | null
  open: (title: string, message: string, onConfirm: () => Promise<void>) => void
  close: () => void
  setDeleting: (isDeleting: boolean) => void
}

export const useDeleteConfirmModalStore = create<ModalState>((set) => ({
  isOpen: false,
  title: '',
  message: '',
  isDeleting: false,
  onConfirm: null,
  open: (title, message, onConfirm) => set({ isOpen: true, title, message, onConfirm }),
  close: () => set({ isOpen: false, onConfirm: null }),
  setDeleting: (isDeleting) => set({ isDeleting }),
}))

// ─── Hook ─────────────────────────────────────────────────────────────────────

export function useDeleteConfirmModal() {
  const open = useDeleteConfirmModalStore((s) => s.open)
  return (title: string, message: string, onConfirm: () => Promise<void>) => {
    open(title, message, onConfirm)
  }
}

// ─── Modal Component ──────────────────────────────────────────────────────────

function DeleteConfirmModalContent() {
  const { isOpen, title, message, isDeleting, onConfirm, close, setDeleting } = useDeleteConfirmModalStore()

  const handleConfirm = async () => {
    if (!onConfirm) return
    setDeleting(true)
    try {
      await onConfirm()
    } finally {
      setDeleting(false)
      close()
    }
  }

  if (!isOpen) return null

  return (
    <div className={styles.overlay} onClick={() => !isDeleting && close()}>
      <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
        <div className={styles.icon}>
          <Trash2 size={24} />
        </div>
        <h3 className={styles.title}>{title}</h3>
        <p className={styles.message}>{message}</p>
        <div className={styles.actions}>
          <button
            className={styles.btnSecondary}
            onClick={close}
            disabled={isDeleting}
          >
            Batal
          </button>
          <button
            className={styles.btnDanger}
            onClick={handleConfirm}
            disabled={isDeleting}
          >
            {isDeleting ? 'Menghapus...' : 'Ya, Hapus'}
          </button>
        </div>
      </div>
    </div>
  )
}

export function DeleteConfirmModalProvider() {
  return <DeleteConfirmModalContent />
}
