import { useState, useCallback } from 'react'

const RESET_DELAY_MS = 2000

interface UseClipboardReturn {
  copied: boolean
  copy: (text: string) => Promise<void>
  reset: () => void
}

export function useClipboard(): UseClipboardReturn {
  const [copied, setCopied] = useState(false)
  const [timer, setTimer] = useState<ReturnType<typeof setTimeout> | null>(null)

  const copy = useCallback(async (text: string) => {
    if (timer) clearTimeout(timer)

    try {
      await navigator.clipboard.writeText(text)
    } catch {
      // Fallback
      const el = document.createElement('textarea')
      el.value = text
      el.style.position = 'fixed'
      el.style.opacity = '0'
      document.body.appendChild(el)
      el.focus()
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
    }

    setCopied(true)
    const t = setTimeout(() => setCopied(false), RESET_DELAY_MS)
    setTimer(t)
  }, [timer])

  const reset = useCallback(() => {
    if (timer) clearTimeout(timer)
    setCopied(false)
  }, [timer])

  return { copied, copy, reset }
}
