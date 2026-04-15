import { useState, useRef, useEffect } from 'react'
import { Bell, Check, CheckCheck, X } from 'lucide-react'
import { create } from 'zustand'
import { formatRelative } from '@/utils/format'
import styles from './NotificationBell.module.css'

// ─── Types ────────────────────────────────────────────────────────────────────

export interface NotificationItem {
  id: string
  title: string
  message?: string
  timestamp: string | Date
  read: boolean
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info'
  onClick?: () => void
}

interface NotificationStore {
  items: NotificationItem[]
  add: (item: Omit<NotificationItem, 'id' | 'read'>) => void
  markRead: (id: string) => void
  markAllRead: () => void
  remove: (id: string) => void
  clear: () => void
}

// ─── Store ────────────────────────────────────────────────────────────────────

export const useNotificationStore = create<NotificationStore>((set) => ({
  items: [],
  add: (item) => {
    const id = Math.random().toString(36).slice(2)
    set((s) => ({ items: [{ ...item, id, read: false }, ...s.items] }))
  },
  markRead: (id) =>
    set((s) => ({ items: s.items.map((n) => (n.id === id ? { ...n, read: true } : n)) })),
  markAllRead: () =>
    set((s) => ({ items: s.items.map((n) => ({ ...n, read: true })) })),
  remove: (id) => set((s) => ({ items: s.items.filter((n) => n.id !== id) })),
  clear: () => set({ items: [] }),
}))

// ─── Component ────────────────────────────────────────────────────────────────

const VARIANT_DOT: Record<string, string> = {
  default: styles.dotDefault,
  success: styles.dotSuccess,
  warning: styles.dotWarning,
  danger: styles.dotDanger,
  info: styles.dotInfo,
}

export function NotificationBell() {
  const [open, setOpen] = useState(false)
  const bellRef = useRef<HTMLButtonElement>(null)
  const panelRef = useRef<HTMLDivElement>(null)
  const { items, markRead, markAllRead, remove } = useNotificationStore()

  const unread = items.filter((n) => !n.read).length

  // Close on outside click
  useEffect(() => {
    if (!open) return
    const handle = (e: MouseEvent) => {
      if (
        !bellRef.current?.contains(e.target as Node) &&
        !panelRef.current?.contains(e.target as Node)
      ) setOpen(false)
    }
    document.addEventListener('mousedown', handle)
    return () => document.removeEventListener('mousedown', handle)
  }, [open])

  const handleItemClick = (item: NotificationItem) => {
    markRead(item.id)
    item.onClick?.()
  }

  return (
    <div className={styles.wrapper}>
      <button
        ref={bellRef}
        type="button"
        className={`${styles.bell} ${open ? styles.bellActive : ''}`}
        onClick={() => setOpen((v) => !v)}
        aria-label={`Notifikasi${unread > 0 ? ` (${unread} belum dibaca)` : ''}`}
        aria-haspopup="true"
        aria-expanded={open}
      >
        <Bell size={18} />
        {unread > 0 && (
          <span className={styles.badge} aria-hidden="true">
            {unread > 99 ? '99+' : unread}
          </span>
        )}
      </button>

      {open && (
        <div ref={panelRef} className={styles.panel} role="dialog" aria-label="Notifikasi">
          <div className={styles.header}>
            <span className={styles.title}>Notifikasi</span>
            {unread > 0 && (
              <button
                type="button"
                className={styles.markAllBtn}
                onClick={markAllRead}
                title="Tandai semua sudah dibaca"
              >
                <CheckCheck size={14} />
                Baca semua
              </button>
            )}
          </div>

          <div className={styles.list}>
            {items.length === 0 ? (
              <p className={styles.empty}>Tidak ada notifikasi</p>
            ) : (
              items.map((item) => (
                <div
                  key={item.id}
                  className={`${styles.item} ${!item.read ? styles.unread : ''}`}
                  onClick={() => handleItemClick(item)}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(e) => e.key === 'Enter' && handleItemClick(item)}
                >
                  <div className={`${styles.dot} ${VARIANT_DOT[item.variant ?? 'default']}`} />
                  <div className={styles.itemBody}>
                    <span className={styles.itemTitle}>{item.title}</span>
                    {item.message && <p className={styles.itemMessage}>{item.message}</p>}
                    <time className={styles.itemTime}>
                      {formatRelative(item.timestamp instanceof Date ? item.timestamp : new Date(item.timestamp))}
                    </time>
                  </div>
                  <div className={styles.itemActions}>
                    {!item.read && (
                      <button
                        type="button"
                        className={styles.actionBtn}
                        onClick={(e) => { e.stopPropagation(); markRead(item.id) }}
                        title="Tandai sudah dibaca"
                      >
                        <Check size={12} />
                      </button>
                    )}
                    <button
                      type="button"
                      className={styles.actionBtn}
                      onClick={(e) => { e.stopPropagation(); remove(item.id) }}
                      title="Hapus"
                    >
                      <X size={12} />
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  )
}
