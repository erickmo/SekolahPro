import { useRef, useState, useCallback } from 'react'
import { Upload, X, File, Image, FileText } from 'lucide-react'
import styles from './FileUpload.module.css'

export interface UploadedFile {
  file: File
  previewUrl?: string
}

interface FileUploadProps {
  accept?: string
  multiple?: boolean
  maxSizeMB?: number
  label?: string
  hint?: string
  onChange: (files: UploadedFile[]) => void
  value?: UploadedFile[]
  disabled?: boolean
  error?: string
}

const IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml']

function getFileIcon(file: File) {
  if (IMAGE_TYPES.includes(file.type)) return <Image size={16} />
  if (file.type === 'application/pdf') return <FileText size={16} />
  return <File size={16} />
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function FileUpload({
  accept,
  multiple = false,
  maxSizeMB = 10,
  label = 'Unggah File',
  hint,
  onChange,
  value = [],
  disabled = false,
  error,
}: FileUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [isDragging, setIsDragging] = useState(false)
  const [sizeError, setSizeError] = useState<string | null>(null)
  const maxBytes = maxSizeMB * 1024 * 1024

  const processFiles = useCallback((incoming: FileList | File[]) => {
    setSizeError(null)
    const valid: UploadedFile[] = []

    for (const file of Array.from(incoming)) {
      if (file.size > maxBytes) {
        setSizeError(`"${file.name}" melebihi batas ${maxSizeMB} MB`)
        continue
      }
      const previewUrl = IMAGE_TYPES.includes(file.type)
        ? URL.createObjectURL(file)
        : undefined
      valid.push({ file, previewUrl })
    }

    if (valid.length === 0) return
    onChange(multiple ? [...value, ...valid] : valid.slice(0, 1))
  }, [maxBytes, maxSizeMB, multiple, value, onChange])

  const handleDrop = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragging(false)
    if (disabled) return
    processFiles(e.dataTransfer.files)
  }, [disabled, processFiles])

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    if (!disabled) setIsDragging(true)
  }

  const handleRemove = (index: number) => {
    const next = value.filter((_, i) => i !== index)
    if (value[index].previewUrl) URL.revokeObjectURL(value[index].previewUrl!)
    onChange(next)
  }

  const displayError = error ?? sizeError

  return (
    <div className={styles.wrapper}>
      <div
        className={`${styles.dropzone} ${isDragging ? styles.dragging : ''} ${disabled ? styles.disabled : ''} ${displayError ? styles.hasError : ''}`}
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={() => setIsDragging(false)}
        onClick={() => !disabled && inputRef.current?.click()}
        role="button"
        tabIndex={disabled ? -1 : 0}
        onKeyDown={(e) => e.key === 'Enter' && !disabled && inputRef.current?.click()}
        aria-label={label}
      >
        <input
          ref={inputRef}
          type="file"
          accept={accept}
          multiple={multiple}
          className={styles.hiddenInput}
          onChange={(e) => e.target.files && processFiles(e.target.files)}
          disabled={disabled}
        />
        <Upload size={20} className={styles.uploadIcon} />
        <p className={styles.primaryText}>{label}</p>
        <p className={styles.secondaryText}>
          {hint ?? `Seret & lepas file di sini, atau klik untuk memilih`}
        </p>
        <p className={styles.limit}>Maks. {maxSizeMB} MB{accept ? ` · ${accept}` : ''}</p>
      </div>

      {displayError && <p className={styles.error}>{displayError}</p>}

      {value.length > 0 && (
        <ul className={styles.fileList}>
          {value.map((item, i) => (
            <li key={i} className={styles.fileItem}>
              {item.previewUrl ? (
                <img src={item.previewUrl} alt={item.file.name} className={styles.preview} />
              ) : (
                <span className={styles.fileIconWrap}>{getFileIcon(item.file)}</span>
              )}
              <div className={styles.fileMeta}>
                <span className={styles.fileName}>{item.file.name}</span>
                <span className={styles.fileSize}>{formatBytes(item.file.size)}</span>
              </div>
              {!disabled && (
                <button
                  className={styles.removeBtn}
                  onClick={() => handleRemove(i)}
                  type="button"
                  aria-label={`Hapus ${item.file.name}`}
                >
                  <X size={14} />
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
