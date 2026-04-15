import { useEffect, useState } from 'react'

interface Breakpoint {
  isMobile: boolean
  isTablet: boolean
  isDesktop: boolean
  width: number
}

function getBreakpoint(width: number): Breakpoint {
  return {
    isMobile: width < 768,
    isTablet: width >= 768 && width < 1024,
    isDesktop: width >= 1024,
    width,
  }
}

export function useBreakpoint(): Breakpoint {
  const [bp, setBp] = useState<Breakpoint>(() =>
    getBreakpoint(typeof window !== 'undefined' ? window.innerWidth : 1024),
  )

  useEffect(() => {
    const observer = new ResizeObserver(() => {
      setBp(getBreakpoint(window.innerWidth))
    })
    observer.observe(document.documentElement)
    return () => observer.disconnect()
  }, [])

  return bp
}
