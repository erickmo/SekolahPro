import { useState, useCallback } from 'react'

/**
 * Like useState but persists to localStorage.
 * SSR-safe: falls back gracefully if localStorage is unavailable.
 */
export function useLocalStorage<T>(key: string, initialValue: T) {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key)
      return item ? (JSON.parse(item) as T) : initialValue
    } catch {
      return initialValue
    }
  })

  const setValue = useCallback((value: T | ((prev: T) => T)) => {
    setStoredValue((prev) => {
      const next = typeof value === 'function' ? (value as (prev: T) => T)(prev) : value
      try {
        window.localStorage.setItem(key, JSON.stringify(next))
      } catch {
        // Quota exceeded or private mode — silently fail
      }
      return next
    })
  }, [key])

  const removeValue = useCallback(() => {
    try {
      window.localStorage.removeItem(key)
    } catch { /* ignore */ }
    setStoredValue(initialValue)
  }, [key, initialValue])

  return [storedValue, setValue, removeValue] as const
}
