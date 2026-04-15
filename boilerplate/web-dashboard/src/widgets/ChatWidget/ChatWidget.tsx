import { AnimatePresence } from 'framer-motion'
import { useChatStore } from '@/stores/chat.store'
import { ChatPanel } from './ChatPanel'

export function ChatWidget() {
  const { isOpen } = useChatStore()

  return (
    <AnimatePresence>{isOpen && <ChatPanel />}</AnimatePresence>
  )
}
