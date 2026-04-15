import { useState, useRef, KeyboardEvent } from 'react'
import { X } from 'lucide-react'
import styles from './TagInput.module.css'

interface TagInputProps {
  value: string[]
  onChange: (tags: string[]) => void
  placeholder?: string
  maxTags?: number
  allowDuplicates?: boolean
  disabled?: boolean
  error?: string
  label?: string
  /** Characters that trigger tag creation. Defaults to Enter and comma. */
  separators?: string[]
  id?: string
}

const DEFAULT_SEPARATORS = ['Enter', ',']

export function TagInput({
  value,
  onChange,
  placeholder = 'Ketik lalu tekan Enter...',
  maxTags,
  allowDuplicates = false,
  disabled = false,
  error,
  label,
  separators = DEFAULT_SEPARATORS,
  id,
}: TagInputProps) {
  const [input, setInput] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const inputId = id ?? `tag-input-${Math.random().toString(36).slice(2)}`

  const addTag = (raw: string) => {
    const tag = raw.trim()
    if (!tag) return
    if (!allowDuplicates && value.includes(tag)) { setInput(''); return }
    if (maxTags && value.length >= maxTags) return
    onChange([...value, tag])
    setInput('')
  }

  const removeTag = (index: number) => {
    onChange(value.filter((_, i) => i !== index))
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (separators.includes(e.key)) {
      e.preventDefault()
      addTag(input)
      return
    }
    if (e.key === 'Backspace' && input === '' && value.length > 0) {
      removeTag(value.length - 1)
    }
  }

  const handleBlur = () => {
    if (input.trim()) addTag(input)
  }

  const isFull = maxTags !== undefined && value.length >= maxTags

  return (
    <div className={styles.wrapper}>
      {label && <label htmlFor={inputId} className={styles.label}>{label}</label>}

      <div
        className={`${styles.container} ${error ? styles.hasError : ''} ${disabled ? styles.disabled : ''}`}
        onClick={() => inputRef.current?.focus()}
      >
        {value.map((tag, i) => (
          <span key={i} className={styles.tag}>
            {tag}
            {!disabled && (
              <button
                type="button"
                className={styles.removeBtn}
                onClick={() => removeTag(i)}
                aria-label={`Hapus tag "${tag}"`}
              >
                <X size={10} />
              </button>
            )}
          </span>
        ))}

        {!isFull && !disabled && (
          <input
            ref={inputRef}
            id={inputId}
            type="text"
            className={styles.input}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            onBlur={handleBlur}
            placeholder={value.length === 0 ? placeholder : ''}
            disabled={disabled}
            aria-invalid={!!error}
          />
        )}
      </div>

      <div className={styles.footer}>
        {error && <p className={styles.error}>{error}</p>}
        {maxTags && (
          <span className={styles.counter}>{value.length}/{maxTags}</span>
        )}
      </div>
    </div>
  )
}
