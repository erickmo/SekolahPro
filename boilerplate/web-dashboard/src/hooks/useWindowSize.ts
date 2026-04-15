import { useEffect, useState } from 'react'

interface WindowSize {
  width: number
  height: number
}

const SSR_SIZE: WindowSize = { width: 0, height: 0 }

function getWindowSize(): WindowSize {
  if (typeof window === 'undefined') return SSR_SIZE
  return { width: window.innerWidth, height: window.innerHeight }
}

export function useWindowSize(): WindowSize {
  const [size, setSize] = useState<WindowSize>(getWindowSize)

  useEffect(() => {
    if (typeof window === 'undefined') return

    let rafId: ReturnType<typeof setTimeout>

    const handleResize = () => {
      clearTimeout(rafId)
      rafId = setTimeout(() => setSize(getWindowSize()), 16)
    }

    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      clearTimeout(rafId)
    }
  }, [])

  return size
}
