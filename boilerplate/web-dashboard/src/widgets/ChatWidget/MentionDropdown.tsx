import { useEffect, useRef, useState } from 'react'
import type { ChatMember } from './chat.types'
import styles from './MentionDropdown.module.css'

const MAX_RESULTS = 6

interface MentionDropdownProps {
  query: string
  members: ChatMember[]
  onSelect: (member: ChatMember) => void
  onClose: () => void
}

export function MentionDropdown({ query, members, onSelect, onClose }: MentionDropdownProps) {
  const [activeIndex, setActiveIndex] = useState(0)
  const [prevQuery, setPrevQuery] = useState(query)
  const listRef = useRef<HTMLDivElement>(null)

  const filtered = members
    .filter((m) => m.name.toLowerCase().includes(query.toLowerCase()))
    .slice(0, MAX_RESULTS)

  // Reset active index when query changes (derived state pattern — avoids useEffect)
  if (prevQuery !== query) {
    setPrevQuery(query)
    setActiveIndex(0)
  }

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setActiveIndex((i) => Math.min(i + 1, filtered.length - 1))
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        setActiveIndex((i) => Math.max(i - 1, 0))
      } else if (e.key === 'Enter') {
        e.preventDefault()
        if (filtered[activeIndex]) onSelect(filtered[activeIndex])
      } else if (e.key === 'Escape') {
        onClose()
      }
    }
    window.addEventListener('keydown', handleKey, true)
    return () => window.removeEventListener('keydown', handleKey, true)
  }, [filtered, activeIndex, onSelect, onClose])

  if (filtered.length === 0) return null

  return (
    <div className={styles.dropdown} ref={listRef} role="listbox" aria-label="Pilih member">
      <p className={styles.hint}>Member di channel</p>
      {filtered.map((member, idx) => (
        <button
          key={member.id}
          className={`${styles.item} ${idx === activeIndex ? styles.itemActive : ''}`}
          onMouseDown={(e) => {
            e.preventDefault()
            onSelect(member)
          }}
          onMouseEnter={() => setActiveIndex(idx)}
          role="option"
          aria-selected={idx === activeIndex}
        >
          <span className={styles.avatar}>{member.initials}</span>
          <span className={styles.info}>
            <span className={styles.name}>{member.name}</span>
            <span className={styles.role}>{member.role}</span>
          </span>
          <span className={`${styles.dot} ${member.isOnline ? styles.dotOnline : styles.dotOffline}`} />
        </button>
      ))}
    </div>
  )
}
