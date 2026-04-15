import { useCallback, useEffect, useRef } from 'react'

interface UseTimeoutReturn {
  reset: () => void
  clear: () => void
}

export function useTimeout(callback: () => void, delay: number): UseTimeoutReturn {
  const callbackRef = useRef<() => void>(callback)
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    callbackRef.current = callback
  }, [callback])

  const clear = useCallback(() => {
    if (timeoutRef.current !== null) {
      clearTimeout(timeoutRef.current)
      timeoutRef.current = null
    }
  }, [])

  const reset = useCallback(() => {
    clear()
    timeoutRef.current = setTimeout(() => callbackRef.current(), delay)
  }, [clear, delay])

  useEffect(() => {
    reset()
    return clear
  }, [reset, clear])

  return { reset, clear }
}
