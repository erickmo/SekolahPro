import { useEffect } from 'react'
import { ChevronDown, ChevronRight, Hash, Plus, Settings, Users, X } from 'lucide-react'
import { motion } from 'framer-motion'
import { useChatStore } from '@/stores/chat.store'
import { useAuthStore } from '@/stores/auth.store'
import { ChatMessageList } from './ChatMessageList'
import { ChatInput } from './ChatInput'
import { ChannelSettingsPanel } from './ChannelSettingsPanel'
import { useChannelFormModal } from './ChannelFormModal'
import type { ChannelCategory, ChatChannel } from './chat.types'
import styles from './ChatPanel.module.css'

// ─── RBAC Constants ──────────────────────────────────────────────────────────

const CHANNEL_CREATOR_ROLES = ['superuser', 'tenant_owner']

// ─── Channel Item ────────────────────────────────────────────────────────────

function ChannelItem({ channel }: { channel: ChatChannel }) {
  const { activeChannelId, setActiveChannel } = useChatStore()
  const isActive = channel.id === activeChannelId

  return (
    <button
      className={`${styles.channelItem} ${isActive ? styles.channelItemActive : ''}`}
      onClick={() => setActiveChannel(channel.id)}
      aria-label={`Channel ${channel.name}`}
    >
      <Hash size={14} className={styles.channelHash} />
      <span className={styles.channelName}>{channel.name}</span>
      {channel.unreadCount > 0 && (
        <span className={styles.channelBadge}>
          {channel.unreadCount > 99 ? '99+' : channel.unreadCount}
        </span>
      )}
    </button>
  )
}

// ─── Category Section ────────────────────────────────────────────────────────

function CategorySection({ category }: { category: ChannelCategory }) {
  const { toggleCategory } = useChatStore()

  return (
    <div className={styles.category}>
      <button
        className={styles.categoryHeader}
        onClick={() => toggleCategory(category.id)}
        aria-expanded={!category.isCollapsed}
      >
        {category.isCollapsed ? (
          <ChevronRight size={12} className={styles.categoryChevron} />
        ) : (
          <ChevronDown size={12} className={styles.categoryChevron} />
        )}
        <span className={styles.categoryLabel}>{category.label}</span>
      </button>
      {!category.isCollapsed && (
        <div className={styles.channelList}>
          {category.channels.map((ch) => (
            <ChannelItem key={ch.id} channel={ch} />
          ))}
        </div>
      )}
    </div>
  )
}

// ─── Sidebar ──────────────────────────────────────────────────────────────────

function ChannelSidebar() {
  const { categories, isLoadingChannels } = useChatStore()
  const userRole = useAuthStore((s) => s.user?.role ?? '')
  const canCreate = CHANNEL_CREATOR_ROLES.includes(userRole)
  const openChannelForm = useChannelFormModal()

  return (
    <div className={styles.sidebar}>
      {/* Sidebar header with add button */}
      {canCreate && (
        <div className={styles.sidebarHeader}>
          <button
            className={styles.addChannelBtn}
            onClick={openChannelForm}
            title="Buat channel baru"
          >
            <Plus size={14} />
          </button>
        </div>
      )}
      <div className={styles.sidebarScroll}>
        {isLoadingChannels ? (
          <div className={styles.sidebarLoading}>Memuat...</div>
        ) : categories.length === 0 ? (
          <div className={styles.sidebarLoading}>Belum ada channel</div>
        ) : (
          categories.map((cat) => (
            <CategorySection key={cat.id} category={cat} />
          ))
        )}
      </div>
    </div>
  )
}

// ─── Channel Content ─────────────────────────────────────────────────────────

function ChannelContent() {
  const { activeChannelId, categories, showSettings, setShowSettings, currentMemberRole } =
    useChatStore()

  const activeChannel = categories
    .flatMap((c) => c.channels)
    .find((ch) => ch.id === activeChannelId)

  if (!activeChannel) {
    return (
      <div className={styles.contentEmpty}>
        <Hash size={32} className={styles.contentEmptyIcon} />
        <p className={styles.contentEmptyText}>Pilih channel untuk mulai chatting</p>
      </div>
    )
  }

  // Show settings panel instead of chat
  if (showSettings) {
    return <ChannelSettingsPanel channel={activeChannel} />
  }

  const isModerator = currentMemberRole === 'moderator'

  return (
    <div className={styles.content}>
      {/* Channel Header */}
      <div className={styles.contentHeader}>
        <div className={styles.contentHeaderLeft}>
          <Hash size={16} className={styles.contentHeaderHash} />
          <span className={styles.contentHeaderName}>{activeChannel.name}</span>
          {activeChannel.description && (
            <>
              <span className={styles.contentHeaderDivider}>|</span>
              <span className={styles.contentHeaderDesc}>{activeChannel.description}</span>
            </>
          )}
        </div>
        <div className={styles.contentHeaderRight}>
          <span className={styles.memberCount}>
            <Users size={12} />
            {/* memberCount loaded from channel */}
          </span>
          {isModerator && (
            <button
              className={styles.settingsBtn}
              onClick={() => setShowSettings(true)}
              title="Pengaturan channel"
            >
              <Settings size={14} />
            </button>
          )}
        </div>
      </div>

      {/* Messages */}
      <div className={styles.messages}>
        <ChatMessageList channelId={activeChannel.id} />
      </div>

      {/* Input */}
      <ChatInput channelId={activeChannel.id} channelName={activeChannel.name} />
    </div>
  )
}

// ─── Panel Shell ──────────────────────────────────────────────────────────────

export function ChatPanel() {
  const { close, fetchChannels } = useChatStore()

  useEffect(() => {
    fetchChannels()
  }, [])

  return (
    <motion.div
      className={styles.panel}
      initial={{ opacity: 0, y: 24, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, y: 16, scale: 0.97 }}
      transition={{ type: 'spring', stiffness: 400, damping: 32 }}
    >
      {/* Top Header */}
      <div className={styles.header}>
        <div className={styles.headerLeft}>
          <Hash size={16} className={styles.headerIcon} />
          <span className={styles.headerTitle}>FlashERP Chat</span>
        </div>
        <button className={styles.headerBtn} onClick={close} aria-label="Tutup panel chat">
          <X size={16} />
        </button>
      </div>

      {/* Two-column body */}
      <div className={styles.body}>
        <ChannelSidebar />
        <ChannelContent />
      </div>
    </motion.div>
  )
}
