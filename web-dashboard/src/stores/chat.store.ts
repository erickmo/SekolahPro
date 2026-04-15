import { create } from 'zustand'
import { chatService } from '@/services/chat.service'
import type {
  ChannelCategory,
  ChatMessage,
  ChatMember,
  MessageStatus,
  SendMessageOptions,
  ChatChannel,
} from '@/widgets/ChatWidget/chat.types'
import type { ChatChannelAPI, ChatMessageAPI, ChatMemberAPI } from '@/services/chat.service'
import { useAuthStore } from './auth.store'

// ─── Helpers ────────────────────────────────────────────────────────────────────

const CURRENT_USER_ID = 'current-user'

function totalUnread(cats: ChannelCategory[]): number {
  return cats.flatMap((c) => c.channels).reduce((acc, ch) => acc + ch.unreadCount, 0)
}

function groupChannelsByCategory(channels: ChatChannelAPI[]): ChannelCategory[] {
  const groups = new Map<string, ChatChannel[]>()
  for (const ch of channels) {
    const cat = ch.kategori || 'Umum'
    if (!groups.has(cat)) groups.set(cat, [])
    groups.get(cat)!.push({
      id: ch.id,
      name: ch.nama,
      description: ch.deskripsi,
      categoryId: cat,
      unreadCount: ch.unread_count ?? 0,
      memberCount: 0,
      lastMessageAt: ch.last_message?.dikirim_pada,
    })
  }
  return Array.from(groups.entries()).map(([label, channels]) => ({
    id: label,
    label,
    channels,
    isCollapsed: false,
  }))
}

function mapMessage(api: ChatMessageAPI): ChatMessage {
  return {
    id: api.id,
    channelId: api.channel_id,
    senderId: api.is_current_user ? CURRENT_USER_ID : api.pengirim_id,
    body: api.isi,
    status: 'sent' as MessageStatus,
    timestamp: api.dikirim_pada,
    mentionIds: api.mentions,
    attachments: api.attachment_url
      ? [{ id: api.id, type: 'image', url: api.attachment_url, name: '', size: 0, mimeType: '' }]
      : undefined,
  }
}

function mapMember(api: ChatMemberAPI): ChatMember {
  return {
    id: api.user_id,
    name: api.nama,
    initials: api.inisial,
    isOnline: false,
    role: api.peran,
  }
}

// ─── State ───────────────────────────────────────────────────────────────────────

interface ChatState {
  categories: ChannelCategory[]
  messagesByChannel: Record<string, ChatMessage[]>
  members: ChatMember[]
  channelMembers: Record<string, ChatMember[]>
  currentMemberRole: string | null
  isOpen: boolean
  activeChannelId: string | null
  unreadCount: number
  isLoadingChannels: boolean
  isLoadingMessages: boolean
  showSettings: boolean
}

interface ChatActions {
  toggleOpen: () => void
  close: () => void
  setActiveChannel: (id: string) => void
  sendMessage: (channelId: string, body: string, options?: SendMessageOptions) => void
  markChannelRead: (id: string) => void
  toggleCategory: (categoryId: string) => void
  upsertMessage: (message: ChatMessage) => void
  updateMessageStatus: (channelId: string, messageId: string, status: MessageStatus) => void
  fetchChannels: () => Promise<void>
  fetchMessages: (channelId: string) => Promise<void>
  fetchMembers: (channelId: string) => Promise<void>
  setShowSettings: (show: boolean) => void
}

export const useChatStore = create<ChatState & ChatActions>()((set, get) => ({
  categories: [],
  messagesByChannel: {},
  members: [],
  channelMembers: {},
  currentMemberRole: null,
  isOpen: false,
  activeChannelId: null,
  unreadCount: 0,
  isLoadingChannels: false,
  isLoadingMessages: false,
  showSettings: false,

  toggleOpen: () => set((s) => ({ isOpen: !s.isOpen })),

  close: () => set({ isOpen: false }),

  setShowSettings: (show) => set({ showSettings: show }),

  setActiveChannel: (id) => {
    set({ activeChannelId: id, showSettings: false })
    get().markChannelRead(id)
    get().fetchMessages(id)
    get().fetchMembers(id)
  },

  markChannelRead: (id) => {
    set((s) => {
      const cats = s.categories.map((cat) => ({
        ...cat,
        channels: cat.channels.map((ch) =>
          ch.id === id ? { ...ch, unreadCount: 0 } : ch,
        ),
      }))
      return { categories: cats, unreadCount: totalUnread(cats) }
    })
    chatService.markRead(id).catch(() => {})
  },

  toggleCategory: (categoryId) =>
    set((s) => ({
      categories: s.categories.map((cat) =>
        cat.id === categoryId ? { ...cat, isCollapsed: !cat.isCollapsed } : cat,
      ),
    })),

  upsertMessage: (message) =>
    set((s) => {
      const existing = s.messagesByChannel[message.channelId] ?? []
      const index = existing.findIndex((m) => m.id === message.id)
      const updated =
        index >= 0 ? existing.map((m, i) => (i === index ? message : m)) : [...existing, message]
      return {
        messagesByChannel: { ...s.messagesByChannel, [message.channelId]: updated },
      }
    }),

  updateMessageStatus: (channelId, messageId, status) =>
    set((s) => ({
      messagesByChannel: {
        ...s.messagesByChannel,
        [channelId]: (s.messagesByChannel[channelId] ?? []).map((m) =>
          m.id === messageId ? { ...m, status } : m,
        ),
      },
    })),

  // ─── API Methods ──────────────────────────────────────────────────────────────

  fetchChannels: async () => {
    set({ isLoadingChannels: true })
    try {
      const channels = await chatService.listChannels()
      const categories = groupChannelsByCategory(channels)
      set({ categories, unreadCount: totalUnread(categories) })
    } catch {
      // Silent fail — keep existing state
    } finally {
      set({ isLoadingChannels: false })
    }
  },

  fetchMessages: async (channelId) => {
    set({ isLoadingMessages: true })
    try {
      const result = await chatService.listMessages(channelId, { limit: 50 })
      const messages = result.items.map(mapMessage)
      set((s) => ({
        messagesByChannel: { ...s.messagesByChannel, [channelId]: messages },
      }))
    } catch {
      // Silent fail
    } finally {
      set({ isLoadingMessages: false })
    }
  },

  fetchMembers: async (channelId) => {
    try {
      const result = await chatService.listMembers(channelId, { limit: 100 })
      const members = result.items.map(mapMember)
      const auth = useAuthStore.getState()
      const me = members.find((m) => m.id === auth.user?.id)
      set((s) => ({
        channelMembers: { ...s.channelMembers, [channelId]: members },
        currentMemberRole: me?.role ?? null,
      }))
    } catch {
      // Silent fail
    }
  },

  sendMessage: (channelId, body, options = {}) => {
    const messageId = Math.random().toString(36).slice(2)
    const now = new Date().toISOString()

    get().upsertMessage({
      id: messageId,
      channelId,
      senderId: CURRENT_USER_ID,
      body,
      status: 'sending',
      timestamp: now,
      mentionIds: options.mentionIds,
      attachments: options.attachments,
      linkedDocuments: options.linkedDocuments,
    })

    set((s) => ({
      categories: s.categories.map((cat) => ({
        ...cat,
        channels: cat.channels.map((ch) =>
          ch.id === channelId ? { ...ch, lastMessageAt: now } : ch,
        ),
      })),
    }))

    // Send to API
    chatService
      .sendMessage(channelId, {
        isi: body,
        mentions: options.mentionIds,
      })
      .then(() => {
        get().updateMessageStatus(channelId, messageId, 'sent')
      })
      .catch(() => {
        get().updateMessageStatus(channelId, messageId, 'error')
      })
  },
}))
