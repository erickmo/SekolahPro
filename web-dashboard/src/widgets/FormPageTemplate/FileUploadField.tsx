import { useState, useRef } from 'react'
import { Upload, X } from 'lucide-react'
import { mediaService } from '@/services/media.service'
import { toast } from '@/widgets/Toast/Toast'
import styles from './FormPageTemplate.module.css'

interface FileUploadFieldProps {
  value?: string
  onChange?: (url: string) => void
  /** Direct file handler — bypasses mediaService upload */
  onFile?: (file: File) => Promise<void>
  label: string
  hint?: string
  required?: boolean
  accept?: string
  maxSizeMB?: number
  previewImageUrl?: string
  onClear?: () => void
}

export function FileUploadField({
  value = '',
  onChange,
  onFile,
  label,
  hint,
  required = false,
  accept = 'image/*',
  maxSizeMB = 10,
  previewImageUrl,
  onClear,
}: FileUploadFieldProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const dropzoneRef = useRef<HTMLDivElement>(null)
  const [isUploading, setIsUploading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)
  const [previewUrl, setPreviewUrl] = useState<string | null>(previewImageUrl ?? null)
  const [isDragOver, setIsDragOver] = useState(false)

  async function handleFileSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    await uploadFile(file)
  }

  async function uploadFile(file: File) {
    // Validate file size
    const fileSizeMB = file.size / (1024 * 1024)
    if (fileSizeMB > maxSizeMB) {
      toast.error(`Ukuran file tidak boleh lebih dari ${maxSizeMB}MB`)
      return
    }

    setIsUploading(true)
    setUploadProgress(0)

    // Direct file mode — bypass mediaService
    if (onFile) {
      try {
        await onFile(file)
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Gagal mengupload file')
      } finally {
        setIsUploading(false)
        setUploadProgress(0)
        if (fileInputRef.current) fileInputRef.current.value = ''
      }
      return
    }

    // Media service mode
    // Create preview
    const reader = new FileReader()
    reader.onload = (event) => {
      setPreviewUrl(event.target?.result as string)
    }
    reader.readAsDataURL(file)

    try {
      const result = await mediaService.uploadFile(file, (progress) => {
        setUploadProgress(progress)
      })
      onChange?.(result.url)
      toast.success('File berhasil diupload')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal mengupload file')
      setPreviewUrl(null)
    } finally {
      setIsUploading(false)
      setUploadProgress(0)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  function handleDragOver(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault()
    e.stopPropagation()
    setIsDragOver(true)
  }

  function handleDragLeave(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault()
    e.stopPropagation()
    setIsDragOver(false)
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault()
    e.stopPropagation()
    setIsDragOver(false)

    const files = e.dataTransfer.files
    if (files.length > 0) {
      uploadFile(files[0])
    }
  }

  return (
    <div className={styles.field}>
      <label className={styles.fieldLabel}>
        {label}
        {required && <span className={styles.required}>*</span>}
        {hint && <span className={styles.hint}> ({hint})</span>}
      </label>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
        {/* Preview */}
        {previewUrl && (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 'var(--space-2)',
              padding: 'var(--space-2)',
              backgroundColor: 'var(--color-surface-raised)',
              border: '1px solid var(--color-border)',
              borderRadius: 'var(--radius-md)',
            }}
          >
            <img
              src={previewUrl}
              alt="Preview"
              style={{
                width: 48,
                height: 48,
                objectFit: 'cover',
                borderRadius: 'var(--radius-sm)',
                flexShrink: 0,
              }}
            />
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontSize: 'var(--font-xs)', fontWeight: 500, color: 'var(--color-text-primary)' }}>
                Pratinjau
              </div>
              <div
                style={{
                  fontSize: 'var(--font-xs)',
                  color: 'var(--color-text-tertiary)',
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                }}
              >
                {value}
              </div>
            </div>
            {onClear && (
              <button
                type="button"
                onClick={onClear}
                style={{
                  padding: '4px 8px',
                  backgroundColor: 'transparent',
                  border: '1px solid var(--color-border)',
                  borderRadius: 'var(--radius-sm)',
                  cursor: 'pointer',
                  fontSize: 'var(--font-xs)',
                  color: 'var(--color-text-tertiary)',
                  transition: 'all var(--duration-fast)',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.backgroundColor = 'var(--color-surface)'
                  e.currentTarget.style.borderColor = 'var(--color-border-dark)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.backgroundColor = 'transparent'
                  e.currentTarget.style.borderColor = 'var(--color-border)'
                }}
              >
                Hapus
              </button>
            )}
          </div>
        )}

        {/* Dropzone */}
        <div
          ref={dropzoneRef}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          onClick={() => fileInputRef.current?.click()}
          style={{
            padding: 'var(--space-4)',
            backgroundColor: isDragOver ? 'var(--color-primary-light, rgba(59, 130, 246, 0.05))' : 'var(--color-surface-raised)',
            border: `2px dashed ${isDragOver ? 'var(--color-primary)' : 'var(--color-border)'}`,
            borderRadius: 'var(--radius-md)',
            cursor: isUploading ? 'not-allowed' : 'pointer',
            transition: 'all var(--duration-fast)',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 'var(--space-2)',
            minHeight: 120,
            opacity: isUploading ? 0.6 : 1,
          }}
        >
          <Upload
            size={32}
            style={{
              color: isDragOver ? 'var(--color-primary)' : 'var(--color-text-tertiary)',
              transition: 'color var(--duration-fast)',
            }}
          />
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 'var(--font-sm)', fontWeight: 500, color: 'var(--color-text-primary)' }}>
              {isUploading ? `Mengunggah... ${uploadProgress > 0 ? `${uploadProgress}%` : ''}` : onFile ? 'Drag & drop file di sini' : 'Drag & drop gambar di sini'}
            </div>
            <div style={{ fontSize: 'var(--font-xs)', color: 'var(--color-text-tertiary)', marginTop: 'var(--space-1)' }}>
              atau klik untuk memilih file
            </div>
          </div>
        </div>

        {/* Hidden file input */}
        <input
          ref={fileInputRef}
          type="file"
          accept={accept}
          onChange={handleFileSelect}
          style={{ display: 'none' }}
          disabled={isUploading}
        />

        {/* URL display — only in media service mode */}
        {value && !isUploading && !onFile && (
          <div
            style={{
              padding: 'var(--space-2)',
              backgroundColor: 'var(--color-surface-raised)',
              border: '1px solid var(--color-border)',
              borderRadius: 'var(--radius-md)',
              fontSize: 'var(--font-xs)',
              color: 'var(--color-text-tertiary)',
              wordBreak: 'break-all',
              maxHeight: 60,
              overflow: 'auto',
            }}
          >
            <strong style={{ color: 'var(--color-text-secondary)' }}>URL:</strong> {value}
          </div>
        )}
      </div>
    </div>
  )
}
