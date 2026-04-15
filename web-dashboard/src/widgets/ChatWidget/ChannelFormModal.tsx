import { useState, useEffect } from 'react'
import { X } from 'lucide-react'
import { create } from 'zustand'
import { chatService } from '@/services/chat.service'
import { toast } from '@/widgets/Toast/Toast'
import { useChatStore } from '@/stores/chat.store'
import styles from './ChannelFormModal.module.css'

// ─── Types ──────────────────────────────────────────────────────────────────────

const CHANNEL_CATEGORIES = [
  { value: 'Perusahaan', label: 'Perusahaan' },
  { value: 'Departemen', label: 'Departemen' },
  { value: 'Santai', label: 'Santai' },
  { value: 'Umum', label: 'Umum' },
]

interface FormState {
  nama: string
  kategori: string
  deskripsi: string
}

interface FormErrors {
  nama?: string
}

const EMPTY_FORM: FormState = {
  nama: '',
  kategori: 'Umum',
  deskripsi: '',
}

// ─── Store ───────────────────────────────────────────────────────────────────────

interface ModalState {
  isOpen: boolean
  open: () => void
  close: () => void
}

const useChannelFormModalStore = create<ModalState>((set) => ({
  isOpen: false,
  open: () => set({ isOpen: true }),
  close: () => set({ isOpen: false }),
}))

// ─── Hook ────────────────────────────────────────────────────────────────────────

export function useChannelFormModal() {
  const open = useChannelFormModalStore((s) => s.open)
  return () => open()
}

// ─── Modal Component ─────────────────────────────────────────────────────────────

function ChannelFormModalContent() {
  const { isOpen, close } = useChannelFormModalStore()
  const fetchChannels = useChatStore((s) => s.fetchChannels)
  const [form, setForm] = useState<FormState>(EMPTY_FORM)
  const [errors, setErrors] = useState<FormErrors>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (isOpen) {
      setForm(EMPTY_FORM)
      setErrors({})
    }
  }, [isOpen])

  function setField<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [key]: value }))
    if (key === 'nama' && errors.nama) {
      setErrors((prev) => ({ ...prev, nama: undefined }))
    }
  }

  function validateForm(): boolean {
    const next: FormErrors = {}
    if (!form.nama.trim()) next.nama = 'Nama channel wajib diisi'
    setErrors(next)
    return Object.keys(next).length === 0
  }

  async function handleSubmit() {
    if (!validateForm()) return

    setIsSubmitting(true)
    try {
      await chatService.createChannel({
        nama: form.nama.trim(),
        kategori: form.kategori,
        deskripsi: form.deskripsi.trim() || undefined,
      })
      toast.success('Channel berhasil dibuat')
      await fetchChannels()
      close()
    } catch {
      toast.error('Gagal membuat channel')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className={styles.overlay} onClick={() => !isSubmitting && close()}>
      <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>Buat Channel Baru</h2>
          <button className={styles.closeBtn} onClick={close} disabled={isSubmitting} title="Tutup">
            <X size={18} />
          </button>
        </div>

        <div className={styles.body}>
          <div className={styles.field}>
            <label className={styles.label}>Nama Channel *</label>
            <input
              type="text"
              className={`${styles.input} ${errors.nama ? styles.inputError : ''}`}
              value={form.nama}
              onChange={(e) => setField('nama', e.target.value)}
              placeholder="contoh: diskusi-marketing"
              disabled={isSubmitting}
              maxLength={100}
            />
            {errors.nama && <div className={styles.error}>{errors.nama}</div>}
          </div>

          <div className={styles.field}>
            <label className={styles.label}>Kategori</label>
            <select
              className={styles.select}
              value={form.kategori}
              onChange={(e) => setField('kategori', e.target.value)}
              disabled={isSubmitting}
            >
              {CHANNEL_CATEGORIES.map((cat) => (
                <option key={cat.value} value={cat.value}>
                  {cat.label}
                </option>
              ))}
            </select>
          </div>

          <div className={styles.field}>
            <label className={styles.label}>Deskripsi</label>
            <textarea
              className={styles.textarea}
              value={form.deskripsi}
              onChange={(e) => setField('deskripsi', e.target.value)}
              placeholder="Deskripsi singkat tentang channel ini..."
              disabled={isSubmitting}
              rows={3}
              maxLength={500}
            />
          </div>
        </div>

        <div className={styles.footer}>
          <button className={styles.btnSecondary} onClick={close} disabled={isSubmitting}>
            Batal
          </button>
          <button className={styles.btnPrimary} onClick={handleSubmit} disabled={isSubmitting}>
            {isSubmitting ? 'Menyimpan...' : 'Buat Channel'}
          </button>
        </div>
      </div>
    </div>
  )
}

export function ChannelFormModalProvider() {
  return <ChannelFormModalContent />
}
