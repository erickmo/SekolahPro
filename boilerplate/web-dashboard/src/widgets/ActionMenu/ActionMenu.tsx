import { useState, useRef, useEffect } from 'react'
import { createPortal } from 'react-dom'
import { MoreHorizontal } from 'lucide-react'
import styles from './ActionMenu.module.css'

export interface ActionMenuItem {
  key: string
  label: string
  icon?: React.ReactNode
  variant?: 'default' | 'danger'
  disabled?: boolean
  hidden?: boolean
  onClick: () => void
}

interface ActionMenuProps {
  items: ActionMenuItem[]
  /** Custom trigger — defaults to MoreHorizontal (3-dot) button */
  trigger?: React.ReactNode
  align?: 'left' | 'right'
  disabled?: boolean
}

const MENU_OFFSET = 4

export function ActionMenu({ items, trigger, align = 'right', disabled = false }: ActionMenuProps) {
  const [open, setOpen] = useState(false)
  const [coords, setCoords] = useState({ top: 0, left: 0 })
  const triggerRef = useRef<HTMLButtonElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)

  const visible = items.filter((i) => !i.hidden)

  const toggle = () => {
    if (disabled) return
    if (!open) {
      const rect = triggerRef.current?.getBoundingClientRect()
      if (rect) {
        setCoords({
          top: rect.bottom + window.scrollY + MENU_OFFSET,
          left: align === 'right'
            ? rect.right + window.scrollX
            : rect.left + window.scrollX,
        })
      }
    }
    setOpen((v) => !v)
  }

  // Close on outside click
  useEffect(() => {
    if (!open) return
    const handle = (e: MouseEvent) => {
      if (
        !triggerRef.current?.contains(e.target as Node) &&
        !menuRef.current?.contains(e.target as Node)
      ) setOpen(false)
    }
    document.addEventListener('mousedown', handle)
    return () => document.removeEventListener('mousedown', handle)
  }, [open])

  // Close on Escape
  useEffect(() => {
    if (!open) return
    const handle = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false) }
    document.addEventListener('keydown', handle)
    return () => document.removeEventListener('keydown', handle)
  }, [open])

  return (
    <>
      <button
        ref={triggerRef}
        type="button"
        className={`${styles.trigger} ${open ? styles.triggerActive : ''}`}
        onClick={toggle}
        disabled={disabled}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label="Aksi"
      >
        {trigger ?? <MoreHorizontal size={16} />}
      </button>

      {open && createPortal(
        <div
          ref={menuRef}
          className={styles.menu}
          style={{
            top: coords.top,
            ...(align === 'right'
              ? { right: `calc(100vw - ${coords.left}px)` }
              : { left: coords.left }),
          }}
          role="menu"
        >
          {visible.map((item) => (
            <button
              key={item.key}
              type="button"
              role="menuitem"
              disabled={item.disabled}
              className={`${styles.item} ${item.variant === 'danger' ? styles.itemDanger : ''}`}
              onClick={() => { item.onClick(); setOpen(false) }}
            >
              {item.icon && <span className={styles.itemIcon}>{item.icon}</span>}
              {item.label}
            </button>
          ))}
        </div>,
        document.body,
      )}
    </>
  )
}
