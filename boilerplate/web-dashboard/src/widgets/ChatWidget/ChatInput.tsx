import { useState, useRef, useEffect, type KeyboardEvent } from 'react'
import { Send, Paperclip, FileText, X, AtSign } from 'lucide-react'
import { useChatStore } from '@/stores/chat.store'
import { MentionDropdown } from './MentionDropdown'
import { DocumentPickerModal } from './DocumentPickerModal'
import type { LinkedDocument, MessageAttachment, ChatMember } from './chat.types'
import styles from './ChatInput.module.css'

const MAX_CHARS = 4000
const WARN_THRESHOLD = 200
const MAX_ATTACHMENTS = 4
const MAX_FILE_SIZE = 20 * 1024 * 1024 // 20MB
const MAX_LINKED_DOCS = 3

interface DraftAttachment {
  id: string
  objectUrl: string
  type: 'image' | 'video'
  name: string
  size: number
  mimeType: string
}

interface ChatInputProps {
  channelId: string
  channelName: string
}

function extractMentionIds(body: string, members: ChatMember[]): string[] {
  const matches = body.match(/@([\w ]+?)(?=\s|$|[,.])/g) ?? []
  return matches
    .map((m) => m.slice(1).trim())
    .map((name) => members.find((mb) => mb.name === name)?.id)
    .filter((id): id is string => Boolean(id))
}

export function ChatInput({ channelId, channelName }: ChatInputProps) {
  const [value, setValue] = useState('')
  const [isSending, setIsSending] = useState(false)
  const [mentionQuery, setMentionQuery] = useState<string | null>(null)
  const [mentionAnchorPos, setMentionAnchorPos] = useState(0)
  const [draftAttachments, setDraftAttachments] = useState<DraftAttachment[]>([])
  const [linkedDocs, setLinkedDocs] = useState<LinkedDocument[]>([])
  const [showDocPicker, setShowDocPicker] = useState(false)

  const { sendMessage } = useChatStore()
  const channelMembers = useChatStore((s) => s.channelMembers[s.activeChannelId ?? ''] ?? [])
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const remaining = MAX_CHARS - value.length
  const showCounter = remaining <= WARN_THRESHOLD
  const isOverLimit = remaining < 0
  const isEmpty = value.trim().length === 0 && draftAttachments.length === 0 && linkedDocs.length === 0

  // Cleanup object URLs on unmount
  useEffect(() => {
    return () => {
      draftAttachments.forEach((a) => URL.revokeObjectURL(a.objectUrl))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const val = e.target.value
    const cursor = e.target.selectionStart ?? val.length
    setValue(val)

    const textBeforeCursor = val.slice(0, cursor)
    const atIndex = textBeforeCursor.lastIndexOf('@')

    if (atIndex === -1) {
      setMentionQuery(null)
      return
    }

    const fragment = textBeforeCursor.slice(atIndex + 1)
    if (fragment.includes(' ')) {
      setMentionQuery(null)
    } else {
      setMentionQuery(fragment)
      setMentionAnchorPos(atIndex)
    }
  }

  const handleSelectMention = (member: ChatMember) => {
    const cursor = textareaRef.current?.selectionStart ?? value.length
    const before = value.slice(0, mentionAnchorPos)
    const after = value.slice(cursor)
    const newValue = `${before}@${member.name} ${after}`

    setValue(newValue)
    setMentionQuery(null)

    setTimeout(() => {
      if (textareaRef.current) {
        const newCursor = before.length + member.name.length + 2
        textareaRef.current.setSelectionRange(newCursor, newCursor)
        textareaRef.current.focus()
      }
    }, 0)
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (mentionQuery !== null && ['ArrowUp', 'ArrowDown', 'Enter'].includes(e.key)) {
      // Keys are handled by MentionDropdown via window listener
      if (e.key === 'Enter') e.preventDefault()
      return
    }
    if (e.key === 'Escape' && mentionQuery !== null) {
      setMentionQuery(null)
      return
    }
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files ?? [])
    const available = MAX_ATTACHMENTS - draftAttachments.length
    const valid = files.filter((f) => f.size <= MAX_FILE_SIZE).slice(0, available)

    const newDrafts: DraftAttachment[] = valid.map((file) => ({
      id: Math.random().toString(36).slice(2),
      objectUrl: URL.createObjectURL(file),
      type: file.type.startsWith('image/') ? 'image' : 'video',
      name: file.name,
      size: file.size,
      mimeType: file.type,
    }))

    setDraftAttachments((prev) => [...prev, ...newDrafts])
    e.target.value = ''
  }

  const removeAttachment = (id: string) => {
    setDraftAttachments((prev) => {
      const removed = prev.find((a) => a.id === id)
      if (removed) URL.revokeObjectURL(removed.objectUrl)
      return prev.filter((a) => a.id !== id)
    })
  }

  const addDocument = (doc: LinkedDocument) => {
    if (linkedDocs.find((d) => d.id === doc.id)) return
    setLinkedDocs((prev) => [...prev, doc])
    setShowDocPicker(false)
  }

  const removeDocument = (id: string) => {
    setLinkedDocs((prev) => prev.filter((d) => d.id !== id))
  }

  const handleSend = () => {
    if (isEmpty || isOverLimit || isSending) return
    setIsSending(true)

    const attachments: MessageAttachment[] = draftAttachments.map((a) => ({
      id: a.id,
      type: a.type,
      url: a.objectUrl,
      name: a.name,
      size: a.size,
      mimeType: a.mimeType,
    }))

    sendMessage(channelId, value.trim(), {
      mentionIds: extractMentionIds(value, channelMembers),
      attachments: attachments.length > 0 ? attachments : undefined,
      linkedDocuments: linkedDocs.length > 0 ? linkedDocs : undefined,
    })

    setValue('')
    setDraftAttachments([])
    setLinkedDocs([])
    setMentionQuery(null)
    textareaRef.current?.focus()
    setTimeout(() => setIsSending(false), 700)
  }

  const canAddAttachment = draftAttachments.length < MAX_ATTACHMENTS
  const canAddDoc = linkedDocs.length < MAX_LINKED_DOCS

  return (
    <div className={styles.container}>
      {/* Mention dropdown — rendered above the input */}
      {mentionQuery !== null && (
        <MentionDropdown
          query={mentionQuery}
          members={channelMembers}
          onSelect={handleSelectMention}
          onClose={() => setMentionQuery(null)}
        />
      )}

      {/* Attachment previews */}
      {draftAttachments.length > 0 && (
        <div className={styles.attachmentRow}>
          {draftAttachments.map((att) => (
            <div key={att.id} className={styles.attachThumb}>
              {att.type === 'image' ? (
                <img src={att.objectUrl} alt={att.name} className={styles.thumbImg} />
              ) : (
                <video src={att.objectUrl} className={styles.thumbImg} muted />
              )}
              <button
                className={styles.thumbRemove}
                onClick={() => removeAttachment(att.id)}
                aria-label={`Hapus ${att.name}`}
              >
                <X size={10} />
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Linked document chips */}
      {linkedDocs.length > 0 && (
        <div className={styles.docChips}>
          {linkedDocs.map((doc) => (
            <div key={doc.id} className={styles.docChip}>
              <FileText size={11} className={styles.docChipIcon} />
              <span className={styles.docChipCode}>{doc.code}</span>
              <button
                className={styles.docChipRemove}
                onClick={() => removeDocument(doc.id)}
                aria-label={`Hapus ${doc.code}`}
              >
                <X size={10} />
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Textarea + send */}
      <div className={styles.inputRow}>
        <textarea
          ref={textareaRef}
          className={styles.textarea}
          value={value}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder={`Pesan di #${channelName}`}
          rows={1}
          maxLength={MAX_CHARS + 1}
          aria-label="Input pesan"
        />
        <button
          className={`${styles.sendBtn} ${isEmpty || isOverLimit ? styles.disabled : ''}`}
          onClick={handleSend}
          disabled={isEmpty || isOverLimit || isSending}
          aria-label="Kirim pesan"
        >
          <Send size={16} strokeWidth={2} />
        </button>
      </div>

      {/* Toolbar */}
      <div className={styles.toolbar}>
        <button
          className={`${styles.toolBtn} ${!canAddAttachment ? styles.toolBtnDisabled : ''}`}
          onClick={() => canAddAttachment && fileInputRef.current?.click()}
          disabled={!canAddAttachment}
          aria-label="Lampirkan foto/video"
          title="Lampirkan foto/video"
        >
          <Paperclip size={14} />
        </button>
        <button
          className={`${styles.toolBtn} ${!canAddDoc ? styles.toolBtnDisabled : ''}`}
          onClick={() => canAddDoc && setShowDocPicker(true)}
          disabled={!canAddDoc}
          aria-label="Lampirkan dokumen transaksi"
          title="Lampirkan dokumen"
        >
          <FileText size={14} />
        </button>
        <button
          className={styles.toolBtn}
          onClick={() => {
            const pos = textareaRef.current?.selectionStart ?? value.length
            const before = value.slice(0, pos)
            const after = value.slice(pos)
            setValue(`${before}@${after}`)
            setMentionAnchorPos(pos)
            setMentionQuery('')
            setTimeout(() => {
              if (textareaRef.current) {
                textareaRef.current.setSelectionRange(pos + 1, pos + 1)
                textareaRef.current.focus()
              }
            }, 0)
          }}
          aria-label="Mention anggota"
          title="Mention anggota (@)"
        >
          <AtSign size={14} />
        </button>
      </div>

      {/* Char counter */}
      {showCounter && (
        <p className={`${styles.counter} ${isOverLimit ? styles.counterOver : ''}`}>
          {isOverLimit
            ? `Melebihi batas ${Math.abs(remaining)} karakter`
            : `${remaining} karakter tersisa`}
        </p>
      )}

      {/* Hidden file input */}
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*,video/*"
        multiple
        style={{ display: 'none' }}
        onChange={handleFileChange}
      />

      {/* Document picker modal */}
      {showDocPicker && (
        <DocumentPickerModal
          alreadyLinked={linkedDocs.map((d) => d.id)}
          onSelect={addDocument}
          onClose={() => setShowDocPicker(false)}
        />
      )}
    </div>
  )
}
