import { useEffect, useCallback } from 'react'

type KeyboardHandler = (event: KeyboardEvent) => void

interface ShortcutDef {
  key: string
  ctrl?: boolean
  meta?: boolean
  shift?: boolean
  alt?: boolean
  handler: KeyboardHandler
  /** Prevent default browser action */
  preventDefault?: boolean
}

/**
 * Register one or more global keyboard shortcuts.
 *
 * @example
 * useKeyboard([
 *   { key: 'k', ctrl: true, handler: openSearch, preventDefault: true },
 *   { key: 'Escape', handler: closeModal },
 * ])
 */
export function useKeyboard(shortcuts: ShortcutDef[], enabled = true): void {
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    for (const sc of shortcuts) {
      const keyMatch = e.key === sc.key || e.code === sc.key
      const ctrlMatch = sc.ctrl ? (e.ctrlKey || e.metaKey) : !sc.ctrl || false
      const metaMatch = sc.meta ? e.metaKey : true
      const shiftMatch = sc.shift ? e.shiftKey : !sc.shift || true
      const altMatch = sc.alt ? e.altKey : !sc.alt || true

      // Simple match: key + optional modifiers
      if (
        keyMatch &&
        (!sc.ctrl || e.ctrlKey || e.metaKey) &&
        (!sc.meta || e.metaKey) &&
        (!sc.shift || e.shiftKey) &&
        (!sc.alt || e.altKey)
      ) {
        if (sc.preventDefault) e.preventDefault()
        sc.handler(e)
        break
      }

      void ctrlMatch; void metaMatch; void shiftMatch; void altMatch
    }
  }, [shortcuts]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!enabled) return
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown, enabled])
}
