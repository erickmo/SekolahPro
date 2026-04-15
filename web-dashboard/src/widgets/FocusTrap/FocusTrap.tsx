import { useEffect, useRef } from 'react'

const FOCUSABLE_SELECTORS = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
  '[contenteditable="true"]',
].join(', ')

interface FocusTrapProps {
  children: React.ReactNode
  active?: boolean
  /** Whether to return focus to the previously focused element on unmount */
  restoreFocus?: boolean
}

/**
 * Traps keyboard focus within its children while `active` is true.
 * Required for accessible modals, drawers, and dialogs.
 *
 * @example
 * <FocusTrap active={isOpen}>
 *   <div role="dialog">...</div>
 * </FocusTrap>
 */
export function FocusTrap({ children, active = true, restoreFocus = true }: FocusTrapProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const previousFocusRef = useRef<Element | null>(null)

  useEffect(() => {
    if (!active) return

    previousFocusRef.current = document.activeElement

    // Focus the first focusable element
    const first = containerRef.current?.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTORS)[0]
    first?.focus()

    return () => {
      if (restoreFocus && previousFocusRef.current instanceof HTMLElement) {
        previousFocusRef.current.focus()
      }
    }
  }, [active, restoreFocus])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    if (!active || e.key !== 'Tab') return

    const container = containerRef.current
    if (!container) return

    const focusable = Array.from(
      container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTORS),
    ).filter((el) => !el.closest('[aria-hidden="true"]'))

    if (focusable.length === 0) { e.preventDefault(); return }

    const first = focusable[0]
    const last = focusable[focusable.length - 1]

    if (e.shiftKey) {
      if (document.activeElement === first) { e.preventDefault(); last.focus() }
    } else {
      if (document.activeElement === last) { e.preventDefault(); first.focus() }
    }
  }

  return (
    <div ref={containerRef} onKeyDown={handleKeyDown}>
      {children}
    </div>
  )
}
