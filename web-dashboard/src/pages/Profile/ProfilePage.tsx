import { useState } from 'react'
import { Camera, Moon, Sun, Monitor } from 'lucide-react'
import { useAuthStore } from '@/stores/auth.store'
import { useUiStore } from '@/stores/ui.store'
import { apiClient } from '@/services/api.client'
import { useForm } from '@/hooks/useForm'
import { AppApiError } from '@/types/api.types'
import { toast } from '@/widgets/Toast/Toast'
import styles from './ProfilePage.module.css'

// ─── Constants ────────────────────────────────────────────────────────────────

const THEME_OPTIONS = [
  { value: 'light', label: 'Light', Icon: Sun },
  { value: 'dark', label: 'Dark', Icon: Moon },
  { value: 'system', label: 'System', Icon: Monitor },
] as const

type ThemeValue = 'light' | 'dark' | 'system'

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getInitials(name: string): string {
  return name
    .split(' ')
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? '')
    .join('')
}

// ─── Profile form types ───────────────────────────────────────────────────────

interface ProfileFormValues {
  name: string
  phone: string
}

// ─── Profile page ─────────────────────────────────────────────────────────────

export default function ProfilePage() {
  const user = useAuthStore((s) => s.user)
  const theme = useUiStore((s) => s.theme)
  const setTheme = useUiStore((s) => s.setTheme)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const { values, errors, field, handleSubmit } = useForm<ProfileFormValues>({
    initialValues: {
      name: user?.name ?? '',
      phone: '',
    },
    validate: (v) => ({
      name: !v.name.trim() ? 'Nama lengkap wajib diisi' : undefined,
      phone: undefined,
    }),
  })

  const onSubmit = handleSubmit(async (v) => {
    setIsSubmitting(true)
    try {
      await apiClient.put('/api/profile', { name: v.name, phone: v.phone })
      toast.success('Profil berhasil diperbarui')
    } catch (err) {
      if (err instanceof AppApiError) {
        toast.error(err.message)
      } else {
        toast.error('Terjadi kesalahan, coba lagi')
      }
    } finally {
      setIsSubmitting(false)
    }
  })

  const initials = getInitials(user?.name ?? 'U')

  return (
    <div className={styles.root}>
      <div className={styles.container}>
        <h1 className={styles.pageTitle}>Profil Saya</h1>

        {/* Section 1: Avatar & Identitas */}
        <section className={styles.section}>
          <div className={styles.avatarRow}>
            <div className={styles.avatarWrap}>
              {user?.avatar ? (
                <img src={user.avatar} alt={user.name} className={styles.avatarImg} />
              ) : (
                <div className={styles.avatarInitials}>{initials}</div>
              )}
              <div className={styles.avatarOverlay} title="Coming soon">
                <Camera size={18} />
              </div>
            </div>

            <div className={styles.avatarInfo}>
              <p className={styles.userName}>{user?.name ?? '—'}</p>
              <p className={styles.userEmail}>{user?.email ?? '—'}</p>
              <button
                type="button"
                className={styles.changePhotoBtn}
                disabled
                title="Coming soon"
              >
                Ganti Foto
              </button>
            </div>
          </div>
        </section>

        {/* Section 2: Edit Profil */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Edit Profil</h2>
          <form onSubmit={onSubmit} noValidate>
            <div className={styles.formGrid}>
              {/* Nama Lengkap */}
              <div className={styles.field}>
                <label htmlFor="name" className={styles.label}>
                  Nama Lengkap <span className={styles.required}>*</span>
                </label>
                <input
                  id="name"
                  type="text"
                  className={styles.input}
                  placeholder="Masukkan nama lengkap"
                  aria-invalid={!!errors.name}
                  {...field('name')}
                  value={String(values.name)}
                />
                {errors.name && <p className={styles.error}>{errors.name}</p>}
              </div>

              {/* Nomor Telepon */}
              <div className={styles.field}>
                <label htmlFor="phone" className={styles.label}>
                  Nomor Telepon
                  <span className={styles.optional}>(opsional)</span>
                </label>
                <input
                  id="phone"
                  type="tel"
                  className={styles.input}
                  placeholder="Contoh: 08123456789"
                  {...field('phone')}
                  value={String(values.phone)}
                />
              </div>
            </div>

            <div className={styles.formActions}>
              <button
                type="submit"
                className={styles.submitBtn}
                disabled={isSubmitting}
              >
                {isSubmitting ? <span className={styles.spinner} /> : 'Simpan Perubahan'}
              </button>
            </div>
          </form>
        </section>

        {/* Section 3: Preferensi */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Preferensi</h2>
          <p className={styles.prefLabel}>Tema Tampilan</p>
          <div className={styles.themeOptions}>
            {THEME_OPTIONS.map(({ value, label, Icon }) => (
              <button
                key={value}
                type="button"
                className={`${styles.themeOption} ${theme === value ? styles.themeOptionActive : ''}`}
                onClick={() => setTheme(value as ThemeValue)}
                aria-pressed={theme === value}
              >
                <Icon size={16} />
                <span>{label}</span>
              </button>
            ))}
          </div>
        </section>
      </div>
    </div>
  )
}
