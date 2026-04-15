import { useEffect, useState } from 'react'

export function useDebounce<T>(value: T, ms = 300): T {
  const [debounced, setDebounced] = useState<T>(value)

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), ms)
    return () => clearTimeout(timer)
  }, [value, ms])

  return debounced
}
