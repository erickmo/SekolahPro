import { MessageSquare, X } from 'lucide-react'
import { useChatStore } from '@/stores/chat.store'
import styles from './ChatButton.module.css'

export function ChatButton() {
  const { isOpen, toggleOpen, unreadCount } = useChatStore()

  return (
    <button
      className={`${styles.fab} ${isOpen ? styles.open : ''}`}
      onClick={toggleOpen}
      aria-label={isOpen ? 'Tutup obrolan' : 'Buka obrolan'}
    >
      <span className={styles.iconWrap}>
        {isOpen ? <X size={22} strokeWidth={2.5} /> : <MessageSquare size={22} strokeWidth={2} />}
      </span>
      {!isOpen && unreadCount > 0 && (
        <span className={styles.badge}>{unreadCount > 99 ? '99+' : unreadCount}</span>
      )}
    </button>
  )
}
