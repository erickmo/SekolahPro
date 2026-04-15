import { useEffect, useRef, useState, RefObject } from 'react'

interface UseIntersectionObserverOptions extends IntersectionObserverInit {
  /** Fire only once, then disconnect */
  once?: boolean
}

interface UseIntersectionObserverReturn<T extends Element> {
  ref: RefObject<T>
  isIntersecting: boolean
  entry: IntersectionObserverEntry | null
}

/**
 * Observes whether an element is visible in the viewport.
 * Useful for lazy loading, infinite scroll, and scroll-triggered animations.
 *
 * @example
 * const { ref, isIntersecting } = useIntersectionObserver<HTMLDivElement>({ threshold: 0.1, once: true })
 * return <div ref={ref} className={isIntersecting ? 'visible' : ''} />
 */
export function useIntersectionObserver<T extends Element = Element>({
  root = null,
  rootMargin = '0px',
  threshold = 0,
  once = false,
}: UseIntersectionObserverOptions = {}): UseIntersectionObserverReturn<T> {
  const ref = useRef<T>(null)
  const [entry, setEntry] = useState<IntersectionObserverEntry | null>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return

    const observer = new IntersectionObserver(
      ([obs]) => {
        setEntry(obs)
        if (once && obs.isIntersecting) observer.disconnect()
      },
      { root, rootMargin, threshold },
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [root, rootMargin, threshold, once])

  return { ref, isIntersecting: entry?.isIntersecting ?? false, entry }
}
