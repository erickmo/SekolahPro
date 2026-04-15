import { useState, useRef } from 'react'
import { Upload, X, Image, Film } from 'lucide-react'
import { mediaService } from '@/services/media.service'
import { toast } from '@/widgets/Toast/Toast'
import styles from './FormPageTemplate.module.css'

interface MultiFileUploadFieldProps {
  values: string[]
  onChange: (urls: string[]) => void
  label: string
  hint?: string
  accept?: string
  maxSizeMB?: number
  maxFiles?: number
}

interface UploadingFile {
  id: string
  name: string
  progress: number
}

export function MultiFileUploadField({
  values,
  onChange,
  label,
  hint,
  accept = 'image/*,video/*',
  maxSizeMB = 10,
  maxFiles = 10,
}: MultiFileUploadFieldProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState<UploadingFile[]>([])
  const [isDragOver, setIsDragOver] = useState(false)

  const isUploading = uploading.length > 0
  const canAddMore = values.length + uploading.length < maxFiles

  async function handleFiles(files: FileList | File[]) {
    const fileArray = Array.from(files)
    const remaining = maxFiles - values.length - uploading.length
    if (remaining <= 0) {
      toast.error(`Maksimal ${maxFiles} file`)
      return
    }
    const toUpload = fileArray.slice(0, remaining)

    for (const file of toUpload) {
      if (file.size / (1024 * 1024) > maxSizeMB) {
        toast.error(`${file.name}: ukuran melebihi ${maxSizeMB}MB`)
        continue
      }

      const uploadId = crypto.randomUUID()
      setUploading((prev) => [...prev, { id: uploadId, name: file.name, progress: 0 }])

      try {
        const result = await mediaService.uploadFile(file, (progress) => {
          setUploading((prev) =>
            prev.map((u) => (u.id === uploadId ? { ...u, progress } : u)),
          )
        })
        onChange([...values, result.url])
      } catch (err) {
        toast.error(err instanceof Error ? err.message : `Gagal upload ${file.name}`)
      } finally {
        setUploading((prev) => prev.filter((u) => u.id !== uploadId))
      }
    }
  }

  function handleFileSelect(e: React.ChangeEvent<HTMLInputElement>) {
    if (e.target.files && e.target.files.length > 0) {
      handleFiles(e.target.files)
      e.target.value = ''
    }
  }

  function handleRemove(index: number) {
    onChange(values.filter((_, i) => i !== index))
  }

  function handleDragOver(e: React.DragEvent) {
    e.preventDefault()
    e.stopPropagation()
    setIsDragOver(true)
  }

  function handleDragLeave(e: React.DragEvent) {
    e.preventDefault()
    e.stopPropagation()
    setIsDragOver(false)
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    e.stopPropagation()
    setIsDragOver(false)
    if (e.dataTransfer.files.length > 0) {
      handleFiles(e.dataTransfer.files)
    }
  }

  const isVideo = (url: string) => /\.(mp4|webm|mov|avi)($|\?)/i.test(url)

  return (
    <div className={styles.field}>
      <label className={styles.fieldLabel}>
        {label}
        {hint && <span className={styles.hint}> ({hint})</span>}
      </label>

      {/* Thumbnail grid */}
      {(values.length > 0 || uploading.length > 0) && (
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(96px, 1fr))',
          gap: '8px',
          marginBottom: '8px',
        }}>
          {values.map((url, i) => (
            <div
              key={i}
              style={{
                position: 'relative',
                aspectRatio: '1',
                borderRadius: 'var(--radius-md)',
                border: '1px solid var(--color-border)',
                overflow: 'hidden',
                backgroundColor: 'var(--color-surface-raised)',
              }}
            >
              {isVideo(url) ? (
                <div style={{
                  width: '100%',
                  height: '100%',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  backgroundColor: '#1a1a2e',
                }}>
                  <Film size={28} style={{ color: '#8b8ba7' }} />
                </div>
              ) : (
                <img
                  src={url}
                  alt={`File ${i + 1}`}
                  style={{ width: '100%', height: '100%', objectFit: 'cover' }}
                />
              )}
              <button
                type="button"
                onClick={() => handleRemove(i)}
                style={{
                  position: 'absolute',
                  top: 4,
                  right: 4,
                  width: 20,
                  height: 20,
                  borderRadius: '50%',
                  border: 'none',
                  background: 'rgba(0,0,0,0.6)',
                  color: '#fff',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  padding: 0,
                }}
                title="Hapus"
              >
                <X size={12} />
              </button>
            </div>
          ))}

          {uploading.map((u) => (
            <div
              key={u.id}
              style={{
                aspectRatio: '1',
                borderRadius: 'var(--radius-md)',
                border: '1px solid var(--color-border)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                backgroundColor: 'var(--color-surface-raised)',
                gap: 4,
              }}
            >
              <Image size={20} style={{ color: 'var(--color-text-tertiary)' }} />
              <span style={{ fontSize: '0.6875rem', color: 'var(--color-text-tertiary)' }}>
                {u.progress}%
              </span>
            </div>
          ))}
        </div>
      )}

      {/* Dropzone */}
      {canAddMore && (
        <div
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          onClick={() => fileInputRef.current?.click()}
          style={{
            padding: 'var(--space-3)',
            backgroundColor: isDragOver ? 'rgba(59, 130, 246, 0.05)' : 'var(--color-surface-raised)',
            border: `2px dashed ${isDragOver ? 'var(--color-primary)' : 'var(--color-border)'}`,
            borderRadius: 'var(--radius-md)',
            cursor: isUploading ? 'not-allowed' : 'pointer',
            transition: 'all var(--duration-fast)',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 'var(--space-1)',
            minHeight: 80,
            opacity: isUploading ? 0.6 : 1,
          }}
        >
          <Upload
            size={24}
            style={{ color: isDragOver ? 'var(--color-primary)' : 'var(--color-text-tertiary)' }}
          />
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 'var(--font-xs)', fontWeight: 500, color: 'var(--color-text-primary)' }}>
              Drag & drop file di sini
            </div>
            <div style={{ fontSize: '0.6875rem', color: 'var(--color-text-tertiary)' }}>
              atau klik untuk memilih ({values.length}/{maxFiles})
            </div>
          </div>
        </div>
      )}

      <input
        ref={fileInputRef}
        type="file"
        accept={accept}
        multiple
        onChange={handleFileSelect}
        style={{ display: 'none' }}
        disabled={isUploading}
      />
    </div>
  )
}
