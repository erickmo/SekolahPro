/* eslint-disable react-refresh/only-export-components */
import { useEffect } from 'react'
import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from 'lucide-react'
import { AnimatePresence, motion } from 'framer-motion'
import { create } from 'zustand'
import styles from './Toast.module.css'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

interface ToastItem {
  id: string
  message: string
  type: ToastType
  duration?: number
}

interface ToastStore {
  toasts: ToastItem[]
  add: (item: Omit<ToastItem, 'id'>) => void
  remove: (id: string) => void
}

export const useToastStore = create<ToastStore>()((set) => ({
  toasts: [],
  add: (item) => {
    const id = Math.random().toString(36).slice(2)
    set((s) => ({ toasts: [...s.toasts, { ...item, id }] }))
  },
  remove: (id) => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })),
}))

export const toast = {
  success: (message: string) => useToastStore.getState().add({ message, type: 'success' }),
  error: (message: string) => useToastStore.getState().add({ message, type: 'error', duration: 6000 }),
  warning: (message: string) => useToastStore.getState().add({ message, type: 'warning' }),
  info: (message: string) => useToastStore.getState().add({ message, type: 'info' }),
}

const ICONS: Record<ToastType, React.ReactNode> = {
  success: <CheckCircle size={16} />,
  error: <AlertCircle size={16} />,
  warning: <AlertTriangle size={16} />,
  info: <Info size={16} />,
}

function ToastItem({ item }: { item: ToastItem }) {
  const { remove } = useToastStore()
  const duration = item.duration ?? (item.type === 'error' ? 6000 : 4000)

  useEffect(() => {
    const timer = setTimeout(() => remove(item.id), duration)
    return () => clearTimeout(timer)
  }, [item.id, duration, remove])

  return (
    <motion.div
      className={`${styles.toast} ${styles[item.type]}`}
      initial={{ opacity: 0, x: 48, scale: 0.9 }}
      animate={{ opacity: 1, x: 0, scale: 1 }}
      exit={{ opacity: 0, x: 48, scale: 0.9 }}
      transition={{ type: 'spring', stiffness: 400, damping: 30 }}
      onClick={() => remove(item.id)}
    >
      <span className={styles.icon}>{ICONS[item.type]}</span>
      <span className={styles.message}>{item.message}</span>
      <button className={styles.close} onClick={() => remove(item.id)} aria-label="Tutup">
        <X size={14} />
      </button>
    </motion.div>
  )
}

export function ToastProvider() {
  const { toasts } = useToastStore()

  return (
    <div className={styles.container}>
      <AnimatePresence mode="sync">
        {toasts.map((t) => (
          <ToastItem key={t.id} item={t} />
        ))}
      </AnimatePresence>
    </div>
  )
}
