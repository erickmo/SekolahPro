import { useState, useEffect, useCallback } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import styles from './LoadingBar.module.css'

// ─── Constants ────────────────────────────────────────────────────────────────

const DEFAULT_HEIGHT = 3
const BAR_Z_INDEX = 301 // --z-navbar (300) + 1

// ─── Types ────────────────────────────────────────────────────────────────────

export interface LoadingBarProps {
  isLoading?: boolean
  color?: string
  height?: number
}

// ─── useLoadingBar hook ───────────────────────────────────────────────────────

export function useLoadingBar() {
  const [isLoading, setIsLoading] = useState(false)
  const start = useCallback(() => setIsLoading(true), [])
  const complete = useCallback(() => setIsLoading(false), [])
  return { isLoading, start, complete }
}

// ─── Internal: ProgressBar ───────────────────────────────────────────────────

interface ProgressBarProps {
  color: string
  height: number
  onExitComplete: () => void
}

type Phase = 'running' | 'completing'

function ProgressBar({ color, height }: ProgressBarProps) {
  const [phase, setPhase] = useState<Phase>('running')

  useEffect(() => {
    // Give the running animation time to show, then complete
    const timer = setTimeout(() => setPhase('completing'), 2400)
    return () => clearTimeout(timer)
  }, [])

  return (
    <div
      className={styles.track}
      style={{ height, zIndex: BAR_Z_INDEX }}
    >
      <div
        className={[
          styles.bar,
          phase === 'running' ? styles.barRunning : styles.barCompleting,
        ].join(' ')}
        style={{ background: color }}
      />
    </div>
  )
}

// ─── Component ────────────────────────────────────────────────────────────────

export function LoadingBar({
  isLoading = false,
  color = 'var(--color-primary)',
  height = DEFAULT_HEIGHT,
}: LoadingBarProps) {
  const [mounted, setMounted] = useState(false)
  const [completing, setCompleting] = useState(false)

  useEffect(() => {
    if (isLoading) {
      setMounted(true)
      setCompleting(false)
    } else {
      // Trigger completing animation, then unmount after fade out
      setCompleting(true)
      const timer = setTimeout(() => setMounted(false), 500)
      return () => clearTimeout(timer)
    }
  }, [isLoading])

  return (
    <AnimatePresence>
      {mounted && (
        <motion.div
          className={styles.track}
          style={{ height, zIndex: BAR_Z_INDEX }}
          initial={{ opacity: 1 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.3, ease: 'easeOut' }}
        >
          <div
            className={[
              styles.bar,
              completing ? styles.barCompleting : styles.barRunning,
            ].join(' ')}
            style={{ background: color }}
          />
        </motion.div>
      )}
    </AnimatePresence>
  )
}
