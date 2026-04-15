import { useState, useRef, useEffect, useCallback } from 'react'
import { Pencil, Check, X } from 'lucide-react'
import styles from './InlineEditField.module.css'

interface InlineEditFieldProps {
  value: string
  onSave: (value: string) => Promise<void> | void
  placeholder?: string
  validate?: (value: string) => string | null
  disabled?: boolean
  multiline?: boolean
  label?: string
  emptyText?: string
}

export function InlineEditField({
  value,
  onSave,
  placeholder = 'Klik untuk mengisi...',
  validate,
  disabled = false,
  multiline = false,
  label,
  emptyText = '—',
}: InlineEditFieldProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [draft, setDraft] = useState(value)
  const [error, setError] = useState<string | null>(null)
  const [isSaving, setIsSaving] = useState(false)
  const inputRef = useRef<HTMLInputElement & HTMLTextAreaElement>(null)

  useEffect(() => {
    if (isEditing) {
      inputRef.current?.focus()
      inputRef.current?.select()
    }
  }, [isEditing])

  // Sync value prop changes while not editing
  useEffect(() => {
    if (!isEditing) setDraft(value)
  }, [value, isEditing])

  const startEdit = () => {
    if (disabled) return
    setDraft(value)
    setError(null)
    setIsEditing(true)
  }

  const cancel = useCallback(() => {
    setDraft(value)
    setError(null)
    setIsEditing(false)
  }, [value])

  const save = useCallback(async () => {
    const trimmed = draft.trim()
    if (validate) {
      const msg = validate(trimmed)
      if (msg) { setError(msg); return }
    }
    setIsSaving(true)
    try {
      await onSave(trimmed)
      setIsEditing(false)
      setError(null)
    } catch {
      setError('Gagal menyimpan, coba lagi')
    } finally {
      setIsSaving(false)
    }
  }, [draft, validate, onSave])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !multiline) { e.preventDefault(); save() }
    if (e.key === 'Escape') cancel()
  }

  const displayValue = value || emptyText
  const isEmpty = !value

  return (
    <div className={styles.wrapper}>
      {label && <span className={styles.label}>{label}</span>}

      {isEditing ? (
        <div className={styles.editBox}>
          {multiline ? (
            <textarea
              ref={inputRef as React.RefObject<HTMLTextAreaElement>}
              className={`${styles.input} ${styles.textarea} ${error ? styles.inputError : ''}`}
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={placeholder}
              disabled={isSaving}
              rows={3}
            />
          ) : (
            <input
              ref={inputRef as React.RefObject<HTMLInputElement>}
              className={`${styles.input} ${error ? styles.inputError : ''}`}
              type="text"
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={placeholder}
              disabled={isSaving}
            />
          )}
          <div className={styles.actions}>
            <button
              className={styles.btnSave}
              onClick={save}
              disabled={isSaving}
              title="Simpan (Enter)"
              type="button"
            >
              <Check size={14} strokeWidth={3} />
            </button>
            <button
              className={styles.btnCancel}
              onClick={cancel}
              disabled={isSaving}
              title="Batal (Esc)"
              type="button"
            >
              <X size={14} />
            </button>
          </div>
          {error && <p className={styles.error}>{error}</p>}
        </div>
      ) : (
        <button
          className={`${styles.display} ${isEmpty ? styles.empty : ''} ${disabled ? styles.disabled : ''}`}
          onClick={startEdit}
          disabled={disabled}
          type="button"
          title={disabled ? undefined : 'Klik untuk mengedit'}
        >
          <span className={styles.displayText}>{displayValue}</span>
          {!disabled && <Pencil size={12} className={styles.pencilIcon} />}
        </button>
      )}
    </div>
  )
}
