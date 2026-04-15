import { AlertTriangle, CheckCircle, Info, Trash2 } from 'lucide-react'
import { create } from 'zustand'
import styles from './ConfirmDialog.module.css'

// ─── Types ────────────────────────────────────────────────────────────────────

export type ConfirmDialogVariant = 'danger' | 'warning' | 'info' | 'success'

interface ConfirmDialogOptions {
  title: string
  message: string
  variant?: ConfirmDialogVariant
  confirmLabel?: string
  cancelLabel?: string
  onConfirm: () => Promise<void> | void
}

interface ConfirmDialogState extends ConfirmDialogOptions {
  isOpen: boolean
  isLoading: boolean
  open: (opts: ConfirmDialogOptions) => void
  close: () => void
  setLoading: (v: boolean) => void
}

// ─── Store ────────────────────────────────────────────────────────────────────

const DEFAULT_STATE: Omit<ConfirmDialogState, 'open' | 'close' | 'setLoading'> = {
  isOpen: false,
  isLoading: false,
  title: '',
  message: '',
  variant: 'danger',
  confirmLabel: 'Ya, Lanjutkan',
  cancelLabel: 'Batal',
  onConfirm: async () => {},
}

export const useConfirmDialogStore = create<ConfirmDialogState>((set) => ({
  ...DEFAULT_STATE,
  open: (opts) => set({ ...DEFAULT_STATE, ...opts, isOpen: true }),
  close: () => set({ isOpen: false, isLoading: false }),
  setLoading: (isLoading) => set({ isLoading }),
}))

// ─── Hook ─────────────────────────────────────────────────────────────────────

export function useConfirmDialog() {
  const open = useConfirmDialogStore((s) => s.open)
  return (opts: ConfirmDialogOptions) => open(opts)
}

// ─── Icons ────────────────────────────────────────────────────────────────────

const VARIANT_ICONS: Record<ConfirmDialogVariant, React.ReactNode> = {
  danger: <Trash2 size={22} />,
  warning: <AlertTriangle size={22} />,
  info: <Info size={22} />,
  success: <CheckCircle size={22} />,
}

// ─── Component ────────────────────────────────────────────────────────────────

function ConfirmDialogContent() {
  const {
    isOpen, isLoading, title, message,
    variant = 'danger', confirmLabel, cancelLabel,
    onConfirm, close, setLoading,
  } = useConfirmDialogStore()

  const handleConfirm = async () => {
    setLoading(true)
    try {
      await onConfirm()
      close()
    } catch {
      setLoading(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className={styles.overlay} onClick={() => !isLoading && close()}>
      <div
        className={`${styles.dialog} ${styles[variant]}`}
        onClick={(e) => e.stopPropagation()}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="confirm-title"
      >
        <div className={styles.iconWrap}>{VARIANT_ICONS[variant]}</div>
        <h3 id="confirm-title" className={styles.title}>{title}</h3>
        <p className={styles.message}>{message}</p>
        <div className={styles.actions}>
          <button
            className={styles.btnCancel}
            onClick={close}
            disabled={isLoading}
          >
            {cancelLabel ?? 'Batal'}
          </button>
          <button
            className={`${styles.btnConfirm} ${styles[`btn_${variant}`]}`}
            onClick={handleConfirm}
            disabled={isLoading}
          >
            {isLoading ? 'Memproses...' : (confirmLabel ?? 'Ya, Lanjutkan')}
          </button>
        </div>
      </div>
    </div>
  )
}

export function ConfirmDialogProvider() {
  return <ConfirmDialogContent />
}
