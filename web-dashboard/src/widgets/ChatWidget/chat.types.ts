export type MessageStatus = 'sending' | 'sent' | 'error'

export type AttachmentType = 'image' | 'video'

export type DocumentModule = 'PO' | 'SO' | 'INV' | 'PINV' | 'DO' | 'PR' | 'RFQ' | 'PAYMENT'

export interface MessageAttachment {
  id: string
  type: AttachmentType
  url: string
  name: string
  size: number
  mimeType: string
}

export interface LinkedDocument {
  id: string
  module: DocumentModule
  code: string
  label: string
  counterpartyName?: string
  total?: number
  status?: string
}

export interface ChatMember {
  id: string
  name: string
  initials: string
  avatar?: string
  isOnline: boolean
  role: string
}

export interface ChatMessage {
  id: string
  channelId: string
  senderId: string
  body: string
  status: MessageStatus
  timestamp: string
  isDeleted?: boolean
  mentionIds?: string[]
  attachments?: MessageAttachment[]
  linkedDocuments?: LinkedDocument[]
}

export interface ChatChannel {
  id: string
  name: string
  description?: string
  categoryId: string
  unreadCount: number
  memberCount: number
  lastMessageAt?: string
}

export interface ChannelCategory {
  id: string
  label: string
  channels: ChatChannel[]
  isCollapsed: boolean
}

export interface SendMessageOptions {
  mentionIds?: string[]
  attachments?: MessageAttachment[]
  linkedDocuments?: LinkedDocument[]
}
