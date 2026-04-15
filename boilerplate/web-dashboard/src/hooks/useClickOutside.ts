import { useEffect, RefObject } from 'react'

/**
 * Fires `handler` when a click occurs outside all provided refs.
 * Accepts a single ref or an array of refs (useful for trigger + panel patterns).
 */
export function useClickOutside<T extends HTMLElement>(
  refs: RefObject<T> | RefObject<T>[],
  handler: (event: MouseEvent | TouchEvent) => void,
  enabled = true,
): void {
  useEffect(() => {
    if (!enabled) return

    const refList = Array.isArray(refs) ? refs : [refs]

    const listener = (event: MouseEvent | TouchEvent) => {
      const target = event.target as Node
      const isInside = refList.some((ref) => ref.current?.contains(target))
      if (!isInside) handler(event)
    }

    document.addEventListener('mousedown', listener)
    document.addEventListener('touchstart', listener)

    return () => {
      document.removeEventListener('mousedown', listener)
      document.removeEventListener('touchstart', listener)
    }
  }, [refs, handler, enabled])
}
