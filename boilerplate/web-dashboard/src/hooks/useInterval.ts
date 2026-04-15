import { useEffect, useRef } from 'react'

interface UseIntervalOptions {
  immediate?: boolean
}

export function useInterval(
  callback: () => void,
  delay: number | null,
  options: UseIntervalOptions = {},
): void {
  const { immediate = false } = options
  const callbackRef = useRef<() => void>(callback)

  useEffect(() => {
    callbackRef.current = callback
  }, [callback])

  useEffect(() => {
    if (delay === null) return

    if (immediate) {
      callbackRef.current()
    }

    const id = setInterval(() => callbackRef.current(), delay)

    return () => clearInterval(id)
  }, [delay, immediate])
}
