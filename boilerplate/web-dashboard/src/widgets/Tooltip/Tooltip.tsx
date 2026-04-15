import React, { useState, useRef, useCallback, useEffect } from 'react'
import { createPortal } from 'react-dom'
import styles from './Tooltip.module.css'

type TooltipPlacement = 'top' | 'bottom' | 'left' | 'right'

interface TooltipProps {
  content: React.ReactNode
  children: React.ReactElement
  placement?: TooltipPlacement
  delay?: number
  disabled?: boolean
  maxWidth?: number
}

const OFFSET = 8

function getPosition(
  trigger: DOMRect,
  tooltip: DOMRect,
  placement: TooltipPlacement,
): { top: number; left: number } {
  const scroll = { x: window.scrollX, y: window.scrollY }

  switch (placement) {
    case 'top':
      return {
        top: trigger.top + scroll.y - tooltip.height - OFFSET,
        left: trigger.left + scroll.x + trigger.width / 2 - tooltip.width / 2,
      }
    case 'bottom':
      return {
        top: trigger.bottom + scroll.y + OFFSET,
        left: trigger.left + scroll.x + trigger.width / 2 - tooltip.width / 2,
      }
    case 'left':
      return {
        top: trigger.top + scroll.y + trigger.height / 2 - tooltip.height / 2,
        left: trigger.left + scroll.x - tooltip.width - OFFSET,
      }
    case 'right':
      return {
        top: trigger.top + scroll.y + trigger.height / 2 - tooltip.height / 2,
        left: trigger.right + scroll.x + OFFSET,
      }
  }
}

export function Tooltip({
  content,
  children,
  placement = 'top',
  delay = 400,
  disabled = false,
  maxWidth = 240,
}: TooltipProps) {
  const [visible, setVisible] = useState(false)
  const [coords, setCoords] = useState({ top: 0, left: 0 })
  const triggerRef = useRef<HTMLElement>(null)
  const tooltipRef = useRef<HTMLDivElement>(null)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const show = useCallback(() => {
    if (disabled || !content) return
    timerRef.current = setTimeout(() => setVisible(true), delay)
  }, [disabled, content, delay])

  const hide = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current)
    setVisible(false)
  }, [])

  useEffect(() => {
    if (!visible || !triggerRef.current || !tooltipRef.current) return
    const triggerRect = triggerRef.current.getBoundingClientRect()
    const tooltipRect = tooltipRef.current.getBoundingClientRect()
    setCoords(getPosition(triggerRect, tooltipRect, placement))
  }, [visible, placement])

  useEffect(() => () => { if (timerRef.current) clearTimeout(timerRef.current) }, [])

  return (
    <>
      {React.cloneElement(children, {
        ref: triggerRef,
        onMouseEnter: show,
        onMouseLeave: hide,
        onFocus: show,
        onBlur: hide,
      } as React.HTMLAttributes<HTMLElement>)}

      {visible && createPortal(
        <div
          ref={tooltipRef}
          className={`${styles.tooltip} ${styles[placement]}`}
          style={{ top: coords.top, left: coords.left, maxWidth }}
          role="tooltip"
        >
          {content}
        </div>,
        document.body,
      )}
    </>
  )
}
