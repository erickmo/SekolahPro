import { useEffect, useRef } from 'react'
import {
  CheckCheck,
  Check,
  AlertCircle,
  Clock,
  ShoppingCart,
  Package,
  FileText,
  Receipt,
  Truck,
  ClipboardList,
  FileSearch,
  CreditCard,
  ExternalLink,
} from 'lucide-react'
import { router } from '@/app/routes'
import { useChatStore } from '@/stores/chat.store'
import { formatRelative, formatCurrency } from '@/utils/format'
import type { ChatMessage, ChatMember, DocumentModule, LinkedDocument } from './chat.types'
import styles from './ChatMessageList.module.css'

const CURRENT_USER_ID = 'current-user' // matches store convention

const MODULE_ROUTE: Record<DocumentModule, string> = {
  PO: 'purchasing/orders',
  SO: 'sales/pesanan',
  INV: 'sales/faktur',
  PINV: 'purchasing/invoices',
  DO: 'sales/surat-jalan',
  PR: 'purchasing/requests',
  RFQ: 'purchasing/quotations',
  PAYMENT: 'accounting/payments',
}

function buildDocUrl(module: DocumentModule, docId: string): string {
  const match = router.state.location.pathname.match(/^\/c\/([^/]+)/)
  const kode = match?.[1]
  if (!kode) return ''
  return `/c/${kode}/${MODULE_ROUTE[module]}/${docId}`
}

const STATUS_ICON: Record<string, React.ReactNode> = {
  sending: <Clock size={12} />,
  sent: <CheckCheck size={12} />,
  error: <AlertCircle size={12} />,
}

const MODULE_ICON: Record<DocumentModule, React.ReactNode> = {
  PO: <ShoppingCart size={12} />,
  SO: <Package size={12} />,
  INV: <FileText size={12} />,
  PINV: <Receipt size={12} />,
  DO: <Truck size={12} />,
  PR: <ClipboardList size={12} />,
  RFQ: <FileSearch size={12} />,
  PAYMENT: <CreditCard size={12} />,
}

// Parse body and highlight @Name mentions
function renderBodyWithMentions(body: string, mentionIds: string[], members: ChatMember[]) {
  if (!mentionIds || mentionIds.length === 0) return body

  const mentionedMembers = mentionIds
    .map((id) => members.find((m) => m.id === id))
    .filter(Boolean) as ChatMember[]

  if (mentionedMembers.length === 0) return body

  const parts = body.split(/(@[\w ]+)/)
  return parts.map((part, i) => {
    const matched = mentionedMembers.find((m) => part === `@${m.name}`)
    if (matched) {
      return (
        <span key={i} className={styles.mention}>
          {part}
        </span>
      )
    }
    return part
  })
}

function DocCard({ doc }: { doc: LinkedDocument }) {
  const { close } = useChatStore()
  const url = buildDocUrl(doc.module, doc.id)

  const handleClick = () => {
    if (!url) return
    close()
    router.navigate(url)
  }

  return (
    <div
      className={`${styles.docCard} ${url ? styles.docCardClickable : ''}`}
      onClick={url ? handleClick : undefined}
      role={url ? 'button' : undefined}
      tabIndex={url ? 0 : undefined}
      onKeyDown={url ? (e) => e.key === 'Enter' && handleClick() : undefined}
      title={url ? `Buka ${doc.code}` : undefined}
    >
      <span className={styles.docCardIcon}>{MODULE_ICON[doc.module]}</span>
      <div className={styles.docCardInfo}>
        <span className={styles.docCardCode}>{doc.code}</span>
        {doc.counterpartyName && (
          <span className={styles.docCardParty}>{doc.counterpartyName}</span>
        )}
      </div>
      <div className={styles.docCardRight}>
        {doc.total !== undefined && (
          <span className={styles.docCardTotal}>{formatCurrency(doc.total, true)}</span>
        )}
        {url && <ExternalLink size={11} className={styles.docCardLink} />}
      </div>
    </div>
  )
}

function MessageBubble({ message, showSender }: { message: ChatMessage; showSender: boolean }) {
  const isMe = message.senderId === CURRENT_USER_ID
  const { members } = useChatStore()
  const sender = members.find((m) => m.id === message.senderId)

  return (
    <div className={`${styles.row} ${isMe ? styles.rowMe : styles.rowOther}`}>
      {!isMe && (
        <div
          className={styles.avatar}
          title={sender?.name}
          style={{ visibility: showSender ? 'visible' : 'hidden' }}
        >
          {sender?.initials ?? '?'}
        </div>
      )}
      <div className={styles.bubble}>
        {!isMe && showSender && sender && (
          <p className={styles.senderName}>{sender.name}</p>
        )}

        {/* Image attachments */}
        {message.attachments && message.attachments.length > 0 && (
          <div className={styles.attachments}>
            {message.attachments.map((att) =>
              att.type === 'image' ? (
                <img
                  key={att.id}
                  src={att.url}
                  alt={att.name}
                  className={styles.attachImage}
                  loading="lazy"
                />
              ) : (
                <video
                  key={att.id}
                  src={att.url}
                  controls
                  className={styles.attachVideo}
                  preload="metadata"
                />
              ),
            )}
          </div>
        )}

        {/* Message body */}
        {message.isDeleted ? (
          <p className={styles.deleted}>Pesan dihapus</p>
        ) : message.body ? (
          <p className={styles.body}>
            {renderBodyWithMentions(
              message.body,
              message.mentionIds ?? [],
              members,
            )}
          </p>
        ) : null}

        {/* Linked documents */}
        {message.linkedDocuments && message.linkedDocuments.length > 0 && (
          <div className={styles.docCards}>
            {message.linkedDocuments.map((doc) => (
              <DocCard key={doc.id} doc={doc} />
            ))}
          </div>
        )}

        <div className={`${styles.meta} ${isMe ? styles.metaMe : ''}`}>
          <span className={styles.time}>{formatRelative(message.timestamp)}</span>
          {isMe && (
            <span
              className={`${styles.status} ${message.status === 'error' ? styles.statusError : ''}`}
            >
              {STATUS_ICON[message.status] ?? <Check size={12} />}
            </span>
          )}
        </div>
      </div>
    </div>
  )
}

function EmptyChannel() {
  return (
    <div className={styles.emptyThread}>
      <p className={styles.emptyText}>Belum ada pesan di channel ini</p>
    </div>
  )
}

interface ChatMessageListProps {
  channelId: string
}

export function ChatMessageList({ channelId }: ChatMessageListProps) {
  const { messagesByChannel } = useChatStore()
  const messages = messagesByChannel[channelId] ?? []
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length])

  if (messages.length === 0) return <EmptyChannel />

  return (
    <div className={styles.list} role="log" aria-live="polite" aria-label="Pesan">
      {messages.map((msg, idx) => {
        const prev = messages[idx - 1]
        const showSender = !prev || prev.senderId !== msg.senderId
        return <MessageBubble key={msg.id} message={msg} showSender={showSender} />
      })}
      <div ref={bottomRef} />
    </div>
  )
}
