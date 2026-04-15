/**
 * Typed localStorage / sessionStorage wrapper.
 * SSR-safe: all methods gracefully handle missing `window` or storage.
 */

type StorageType = 'local' | 'session'

interface StorageService {
  get<T>(key: string, defaultValue?: T): T | undefined
  set<T>(key: string, value: T): void
  remove(key: string): void
  clear(): void
  has(key: string): boolean
  getAll(): Record<string, unknown>
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

function resolveStorage(type: StorageType): Storage | null {
  try {
    if (typeof window === 'undefined') return null
    return type === 'local' ? window.localStorage : window.sessionStorage
  } catch {
    return null
  }
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

function createStorage(type: StorageType): StorageService {
  function get<T>(key: string, defaultValue?: T): T | undefined {
    const store = resolveStorage(type)
    if (store === null) return defaultValue

    try {
      const raw = store.getItem(key)
      if (raw === null) return defaultValue
      return JSON.parse(raw) as T
    } catch {
      return defaultValue
    }
  }

  function set<T>(key: string, value: T): void {
    const store = resolveStorage(type)
    if (store === null) return

    try {
      store.setItem(key, JSON.stringify(value))
    } catch (error) {
      // Silent fail on QuotaExceededError or serialization error
      if (
        error instanceof DOMException &&
        (error.name === 'QuotaExceededError' ||
          error.name === 'NS_ERROR_DOM_QUOTA_REACHED')
      ) {
        return
      }
      // Silently swallow other storage errors (e.g. private-browsing restrictions)
    }
  }

  function remove(key: string): void {
    const store = resolveStorage(type)
    if (store === null) return

    try {
      store.removeItem(key)
    } catch {
      // no-op
    }
  }

  function clear(): void {
    const store = resolveStorage(type)
    if (store === null) return

    try {
      store.clear()
    } catch {
      // no-op
    }
  }

  function has(key: string): boolean {
    const store = resolveStorage(type)
    if (store === null) return false

    try {
      return store.getItem(key) !== null
    } catch {
      return false
    }
  }

  function getAll(): Record<string, unknown> {
    const store = resolveStorage(type)
    if (store === null) return {}

    const result: Record<string, unknown> = {}

    try {
      for (let i = 0; i < store.length; i++) {
        const key = store.key(i)
        if (key === null) continue

        try {
          const raw = store.getItem(key)
          result[key] = raw !== null ? (JSON.parse(raw) as unknown) : undefined
        } catch {
          // Keep raw string value when parse fails
          result[key] = store.getItem(key) ?? undefined
        }
      }
    } catch {
      // no-op
    }

    return result
  }

  return { get, set, remove, clear, has, getAll }
}

// ---------------------------------------------------------------------------
// Pre-built instances
// ---------------------------------------------------------------------------

export const localStore: StorageService = createStorage('local')
export const sessionStore: StorageService = createStorage('session')

export { createStorage }
export type { StorageType, StorageService }
