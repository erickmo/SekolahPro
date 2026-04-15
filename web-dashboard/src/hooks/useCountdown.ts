import { useCallback, useEffect, useRef, useState } from 'react'

import { useInterval } from './useInterval'

interface UseCountdownOptions {
  initialCount: number
  autoStart?: boolean
  onComplete?: () => void
}

interface UseCountdownReturn {
  count: number
  isRunning: boolean
  start: () => void
  pause: () => void
  reset: () => void
  restart: () => void
}

const TICK_MS = 1000

export function useCountdown(options: UseCountdownOptions): UseCountdownReturn {
  const { initialCount, autoStart = true, onComplete } = options

  const [count, setCount] = useState<number>(initialCount)
  const [isRunning, setIsRunning] = useState<boolean>(false)
  const onCompleteRef = useRef<(() => void) | undefined>(onComplete)

  useEffect(() => {
    onCompleteRef.current = onComplete
  }, [onComplete])

  // Tick: decrement count by 1, stop and fire onComplete at 0
  const tick = useCallback(() => {
    setCount((prev) => {
      if (prev <= 1) {
        setIsRunning(false)
        onCompleteRef.current?.()
        return 0
      }
      return prev - 1
    })
  }, [])

  useInterval(tick, isRunning ? TICK_MS : null)

  const start = useCallback(() => {
    setCount((prev) => {
      if (prev > 0) setIsRunning(true)
      return prev
    })
  }, [])

  const pause = useCallback(() => {
    setIsRunning(false)
  }, [])

  const reset = useCallback(() => {
    setIsRunning(false)
    setCount(initialCount)
  }, [initialCount])

  const restart = useCallback(() => {
    setCount(initialCount)
    setIsRunning(true)
  }, [initialCount])

  // Auto-start on mount
  useEffect(() => {
    if (autoStart && initialCount > 0) {
      setIsRunning(true)
    }
    // Only run on mount — intentional single-run
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return { count, isRunning, start, pause, reset, restart }
}
