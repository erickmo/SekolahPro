import { create } from 'zustand'

export interface NotificationItem {
  id: string
  title: string
  message: string
  type: 'info' | 'success' | 'warning' | 'error'
  isRead: boolean
  createdAt: string
  href?: string
}

interface NotificationState {
  items: NotificationItem[]
  unreadCount: number
}

interface NotificationActions {
  add: (item: NotificationItem) => void
  markRead: (id: string) => void
  markAllRead: () => void
  remove: (id: string) => void
}

export const useNotificationStore = create<NotificationState & NotificationActions>()((set) => ({
  items: [],
  unreadCount: 0,

  add: (item: NotificationItem) =>
    set((s) => ({
      items: [item, ...s.items],
      unreadCount: s.unreadCount + (item.isRead ? 0 : 1),
    })),

  markRead: (id: string) =>
    set((s) => ({
      items: s.items.map((n) => (n.id === id ? { ...n, isRead: true } : n)),
      unreadCount: Math.max(0, s.unreadCount - (s.items.find((n) => n.id === id)?.isRead ? 0 : 1)),
    })),

  markAllRead: () =>
    set((s) => ({
      items: s.items.map((n) => ({ ...n, isRead: true })),
      unreadCount: 0,
    })),

  remove: (id: string) =>
    set((s) => {
      const item = s.items.find((n) => n.id === id)
      return {
        items: s.items.filter((n) => n.id !== id),
        unreadCount: Math.max(0, s.unreadCount - (item?.isRead ? 0 : 1)),
      }
    }),
}))
