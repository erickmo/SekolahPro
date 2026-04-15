import { useState, useEffect, useRef, useMemo, useCallback } from 'react'
import { createPortal } from 'react-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { Search } from 'lucide-react'
import styles from './CommandPalette.module.css'

// ─── Constants ────────────────────────────────────────────────────────────────

const DEFAULT_PLACEHOLDER = 'Cari perintah atau halaman...'
const DEFAULT_EMPTY_TEXT = 'Tidak ada hasil ditemukan'
const OVERLAY_Z_INDEX = 'calc(var(--z-modal) + 10)'

// ─── Types ────────────────────────────────────────────────────────────────────

export interface CommandItem {
  id: string
  label: string
  description?: string
  icon?: React.ReactNode
  group?: string
  shortcut?: string[]
  action: () => void
}

export interface CommandPaletteProps {
  commands: CommandItem[]
  isOpen: boolean
  onClose: () => void
  placeholder?: string
  emptyText?: string
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function matchesSearch(item: CommandItem, query: string): boolean {
  if (!query) return true
  const q = query.toLowerCase()
  return (
    item.label.toLowerCase().includes(q) ||
    (item.description?.toLowerCase().includes(q) ?? false)
  )
}

type GroupedCommands = Array<{ group: string | null; items: CommandItem[] }>

function groupCommands(filtered: CommandItem[]): GroupedCommands {
  const map = new Map<string, CommandItem[]>()
  const NO_GROUP = '__nogroup__'

  for (const item of filtered) {
    const key = item.group ?? NO_GROUP
    if (!map.has(key)) map.set(key, [])
    map.get(key)!.push(item)
  }

  const result: GroupedCommands = []
  for (const [key, items] of map.entries()) {
    result.push({ group: key === NO_GROUP ? null : key, items })
  }
  return result
}

// ─── Component ────────────────────────────────────────────────────────────────

export function CommandPalette({
  commands,
  isOpen,
  onClose,
  placeholder = DEFAULT_PLACEHOLDER,
  emptyText = DEFAULT_EMPTY_TEXT,
}: CommandPaletteProps) {
  const [query, setQuery] = useState('')
  const [activeIndex, setActiveIndex] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  const filtered = useMemo(
    () => commands.filter(cmd => matchesSearch(cmd, query)),
    [commands, query],
  )

  const grouped = useMemo(() => groupCommands(filtered), [filtered])

  // Flat list for keyboard navigation
  const flatItems = useMemo(() => filtered, [filtered])

  // Reset state when opened
  useEffect(() => {
    if (isOpen) {
      setQuery('')
      setActiveIndex(0)
      requestAnimationFrame(() => inputRef.current?.focus())
    }
  }, [isOpen])

  // Reset active index when filtered list changes
  useEffect(() => {
    setActiveIndex(0)
  }, [query])

  // Scroll active item into view
  useEffect(() => {
    const activeEl = listRef.current?.querySelector(`[data-active="true"]`) as HTMLElement | null
    activeEl?.scrollIntoView({ block: 'nearest' })
  }, [activeIndex])

  const executeItem = useCallback(
    (item: CommandItem) => {
      item.action()
      onClose()
    },
    [onClose],
  )

  // Keyboard handling
  useEffect(() => {
    if (!isOpen) return

    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        onClose()
        return
      }

      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setActiveIndex(i => (i + 1) % (flatItems.length || 1))
        return
      }

      if (e.key === 'ArrowUp') {
        e.preventDefault()
        setActiveIndex(i => (i - 1 + (flatItems.length || 1)) % (flatItems.length || 1))
        return
      }

      if (e.key === 'Enter') {
        e.preventDefault()
        const item = flatItems[activeIndex]
        if (item) executeItem(item)
      }
    }

    document.addEventListener('keydown', handleKey)
    return () => document.removeEventListener('keydown', handleKey)
  }, [isOpen, flatItems, activeIndex, onClose, executeItem])

  let globalIdx = -1

  const content = (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Overlay */}
          <motion.div
            className={styles.overlay}
            style={{ zIndex: OVERLAY_Z_INDEX as unknown as number }}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.15 }}
            onClick={onClose}
            aria-hidden="true"
          />

          {/* Card */}
          <motion.div
            className={styles.card}
            style={{ zIndex: OVERLAY_Z_INDEX as unknown as number }}
            initial={{ opacity: 0, scale: 0.96, y: -8 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.96, y: -8 }}
            transition={{ duration: 0.18, ease: [0.16, 1, 0.3, 1] }}
            role="dialog"
            aria-modal="true"
            aria-label="Command palette"
          >
            {/* Search input */}
            <div className={styles.searchRow}>
              <Search size={16} className={styles.searchIcon} />
              <input
                ref={inputRef}
                type="text"
                className={styles.searchInput}
                placeholder={placeholder}
                value={query}
                onChange={e => setQuery(e.target.value)}
                autoComplete="off"
                spellCheck={false}
              />
            </div>

            <div className={styles.divider} />

            {/* Results list */}
            <div className={styles.list} ref={listRef}>
              {filtered.length === 0 ? (
                <div className={styles.empty}>{emptyText}</div>
              ) : (
                grouped.map(({ group, items }) => (
                  <div key={group ?? 'default'} className={styles.group}>
                    {group && (
                      <div className={styles.groupHeader}>{group}</div>
                    )}

                    {items.map(item => {
                      globalIdx++
                      const idx = globalIdx
                      const isActive = activeIndex === idx

                      return (
                        <button
                          key={item.id}
                          type="button"
                          className={[
                            styles.item,
                            isActive ? styles.itemActive : '',
                          ].filter(Boolean).join(' ')}
                          data-active={isActive ? 'true' : undefined}
                          onMouseEnter={() => setActiveIndex(idx)}
                          onClick={() => executeItem(item)}
                        >
                          {item.icon && (
                            <span className={styles.itemIcon}>{item.icon}</span>
                          )}

                          <span className={styles.itemContent}>
                            <span className={styles.itemLabel}>{item.label}</span>
                            {item.description && (
                              <span className={styles.itemDescription}>
                                {item.description}
                              </span>
                            )}
                          </span>

                          {item.shortcut && (
                            <span className={styles.shortcut}>
                              {item.shortcut.map((key, i) => (
                                <kbd key={i} className={styles.kbd}>{key}</kbd>
                              ))}
                            </span>
                          )}
                        </button>
                      )
                    })}
                  </div>
                ))
              )}
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )

  return createPortal(content, document.body)
}
