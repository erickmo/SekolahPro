import { useState } from 'react'
import { ImageOff } from 'lucide-react'
import { useIntersectionObserver } from '@/hooks/useIntersectionObserver'
import styles from './LazyImage.module.css'

interface LazyImageProps {
  src: string
  alt: string
  width?: number | string
  height?: number | string
  objectFit?: 'cover' | 'contain' | 'fill' | 'none'
  borderRadius?: string
  className?: string
  /** Shown while loading */
  placeholder?: React.ReactNode
  /** Shown on error */
  fallback?: React.ReactNode
  /** IntersectionObserver rootMargin — how far before viewport to start loading */
  rootMargin?: string
}

export function LazyImage({
  src,
  alt,
  width = '100%',
  height,
  objectFit = 'cover',
  borderRadius,
  className,
  placeholder,
  fallback,
  rootMargin = '200px',
}: LazyImageProps) {
  const [loaded, setLoaded] = useState(false)
  const [error, setError] = useState(false)
  const { ref, isIntersecting } = useIntersectionObserver<HTMLDivElement>({
    rootMargin,
    once: true,
  })

  return (
    <div
      ref={ref}
      className={`${styles.wrapper} ${className ?? ''}`}
      style={{ width, height, borderRadius }}
    >
      {/* Placeholder / shimmer */}
      {!loaded && !error && (
        <div className={styles.placeholder}>
          {placeholder ?? <div className={styles.shimmer} />}
        </div>
      )}

      {/* Error fallback */}
      {error && (
        <div className={styles.errorState}>
          {fallback ?? (
            <>
              <ImageOff size={20} className={styles.errorIcon} />
              <span className={styles.errorText}>Gambar tidak tersedia</span>
            </>
          )}
        </div>
      )}

      {/* Actual image — only fetched after entering viewport */}
      {isIntersecting && !error && (
        <img
          src={src}
          alt={alt}
          className={`${styles.img} ${loaded ? styles.visible : styles.hidden}`}
          style={{ objectFit, borderRadius }}
          onLoad={() => setLoaded(true)}
          onError={() => setError(true)}
          loading="lazy"
          decoding="async"
        />
      )}
    </div>
  )
}
