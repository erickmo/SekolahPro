import React, { useState, useRef, useEffect, useCallback } from 'react'
import { createPortal } from 'react-dom'
import { AnimatePresence, motion } from 'framer-motion'
import styles from './Popover.module.css'

type PopoverPlacement = 'top' | 'bottom' | 'left' | 'right'

interface PopoverProps {
  trigger: React.ReactElement
  children: React.ReactNode
  placement?: PopoverPlacement
  open?: boolean
  onOpenChange?: (open: boolean) => void
  offset?: number
  closeOnOutsideClick?: boolean
}

const DEFAULT_OFFSET = 8

function getPosition(
  trigger: DOMRect,
  popover: DOMRect,
  placement: PopoverPlacement,
  offset: number,
): { top: number; left: number } {
  const scroll = { x: window.scrollX, y: window.scrollY }

  switch (placement) {
    case 'top':
      return {
        top: trigger.top + scroll.y - popover.height - offset,
        left: trigger.left + scroll.x + trigger.width / 2 - popover.width / 2,
      }
    case 'bottom':
      return {
        top: trigger.bottom + scroll.y + offset,
        left: trigger.left + scroll.x + trigger.width / 2 - popover.width / 2,
      }
    case 'left':
      return {
        top: trigger.top + scroll.y + trigger.height / 2 - popover.height / 2,
        left: trigger.left + scroll.x - popover.width - offset,
      }
    case 'right':
      return {
        top: trigger.top + scroll.y + trigger.height / 2 - popover.height / 2,
        left: trigger.right + scroll.x + offset,
      }
  }
}

const MOTION_VARIANTS: Record<PopoverPlacement, { y?: number; x?: number }> = {
  top:    { y: 4 },
  bottom: { y: -4 },
  left:   { x: 4 },
  right:  { x: -4 },
}

export function Popover({
  trigger,
  children,
  placement = 'bottom',
  open: controlledOpen,
  onOpenChange,
  offset = DEFAULT_OFFSET,
  closeOnOutsideClick = true,
}: PopoverProps) {
  const isControlled = controlledOpen !== undefined
  const [internalOpen, setInternalOpen] = useState(false)
  const isOpen = isControlled ? controlledOpen : internalOpen

  const [coords, setCoords] = useState({ top: 0, left: 0 })
  const triggerRef = useRef<HTMLElement>(null)
  const popoverRef = useRef<HTMLDivElement>(null)

  const setOpen = useCallback(
    (value: boolean) => {
      if (!isControlled) setInternalOpen(value)
      onOpenChange?.(value)
    },
    [isControlled, onOpenChange],
  )

  const handleTriggerClick = useCallback(() => {
    setOpen(!isOpen)
  }, [isOpen, setOpen])

  // Position calculation after open
  useEffect(() => {
    if (!isOpen || !triggerRef.current || !popoverRef.current) return
    const triggerRect = triggerRef.current.getBoundingClientRect()
    const popoverRect = popoverRef.current.getBoundingClientRect()
    setCoords(getPosition(triggerRect, popoverRect, placement, offset))
  }, [isOpen, placement, offset])

  // Close on Escape
  useEffect(() => {
    if (!isOpen) return
    const handle = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false) }
    document.addEventListener('keydown', handle)
    return () => document.removeEventListener('keydown', handle)
  }, [isOpen, setOpen])

  // Close on outside click
  useEffect(() => {
    if (!isOpen || !closeOnOutsideClick) return
    const handle = (e: MouseEvent) => {
      const target = e.target as Node
      if (
        triggerRef.current?.contains(target) ||
        popoverRef.current?.contains(target)
      ) return
      setOpen(false)
    }
    document.addEventListener('mousedown', handle)
    return () => document.removeEventListener('mousedown', handle)
  }, [isOpen, closeOnOutsideClick, setOpen])

  const motionInitial = MOTION_VARIANTS[placement]

  return (
    <>
      {React.cloneElement(trigger, {
        ref: triggerRef,
        onClick: handleTriggerClick,
      } as React.HTMLAttributes<HTMLElement> & { ref: React.Ref<HTMLElement> })}

      {createPortal(
        <AnimatePresence>
          {isOpen && (
            <motion.div
              ref={popoverRef}
              className={styles.popover}
              style={{ top: coords.top, left: coords.left }}
              initial={{ opacity: 0, ...motionInitial }}
              animate={{ opacity: 1, y: 0, x: 0 }}
              exit={{ opacity: 0, ...motionInitial }}
              transition={{ duration: 0.15, ease: 'easeOut' }}
            >
              {children}
            </motion.div>
          )}
        </AnimatePresence>,
        document.body,
      )}
    </>
  )
}
