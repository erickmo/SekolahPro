import { useState } from 'react'
import { Eye, EyeOff, Lock, ShieldCheck } from 'lucide-react'
import { apiClient } from '@/services/api.client'
import { useForm } from '@/hooks/useForm'
import { AppApiError } from '@/types/api.types'
import { toast } from '@/widgets/Toast/Toast'
import styles from './ChangePasswordPage.module.css'

// ─── Constants ────────────────────────────────────────────────────────────────

const MIN_PASSWORD_LENGTH = 8

const STRENGTH_LABELS = ['', 'Lemah', 'Sedang', 'Kuat'] as const

// ─── Types ────────────────────────────────────────────────────────────────────

interface ChangePasswordFormValues {
  currentPassword: string
  newPassword: string
  confirmPassword: string
}

type PasswordStrength = 0 | 1 | 2 | 3

// ─── Helpers ──────────────────────────────────────────────────────────────────

function calcStrength(password: string): PasswordStrength {
  if (!password) return 0
  const hasLetter = /[a-zA-Z]/.test(password)
  const hasDigit = /\d/.test(password)
  const hasSymbol = /[^a-zA-Z0-9]/.test(password)
  const mixCount = [hasLetter, hasDigit, hasSymbol].filter(Boolean).length

  if (password.length >= 10 && mixCount === 3) return 3
  if (password.length >= 6 && mixCount >= 2) return 2
  return 1
}

// ─── Change password page ─────────────────────────────────────────────────────

export default function ChangePasswordPage() {
  const [showCurrent, setShowCurrent] = useState(false)
  const [showNew, setShowNew] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [serverError, setServerError] = useState('')

  const { values, errors, field, handleSubmit, reset } = useForm<ChangePasswordFormValues>({
    initialValues: {
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
    },
    validate: (v) => ({
      currentPassword: !v.currentPassword ? 'Kata sandi saat ini wajib diisi' : undefined,
      newPassword: !v.newPassword
        ? 'Kata sandi baru wajib diisi'
        : v.newPassword.length < MIN_PASSWORD_LENGTH
        ? `Minimal ${MIN_PASSWORD_LENGTH} karakter`
        : undefined,
      confirmPassword: !v.confirmPassword
        ? 'Konfirmasi kata sandi wajib diisi'
        : v.confirmPassword !== v.newPassword
        ? 'Kata sandi tidak cocok'
        : undefined,
    }),
  })

  const strength = calcStrength(String(values.newPassword))
  const strengthLabel = STRENGTH_LABELS[strength]

  const onSubmit = handleSubmit(async (v) => {
    setIsSubmitting(true)
    setServerError('')
    try {
      await apiClient.post('/api/auth/change-password', {
        currentPassword: v.currentPassword,
        newPassword: v.newPassword,
      })
      toast.success('Kata sandi berhasil diubah')
      reset()
    } catch (err) {
      if (err instanceof AppApiError) {
        setServerError(err.message)
      } else {
        setServerError('Terjadi kesalahan, coba lagi')
      }
    } finally {
      setIsSubmitting(false)
    }
  })

  const hasValidationErrors = Object.values(errors).some(Boolean)
  const isFormEmpty =
    !values.currentPassword && !values.newPassword && !values.confirmPassword

  return (
    <div className={styles.root}>
      <div className={styles.card}>
        {/* Header */}
        <div className={styles.header}>
          <div className={styles.headerIcon}>
            <ShieldCheck size={24} />
          </div>
          <div>
            <h1 className={styles.title}>Ubah Kata Sandi</h1>
            <p className={styles.subtitle}>
              Pastikan kata sandi baru Anda kuat dan mudah diingat.
            </p>
          </div>
        </div>

        {/* Form */}
        <form onSubmit={onSubmit} noValidate className={styles.form}>
          {/* Kata Sandi Saat Ini */}
          <div className={styles.field}>
            <label htmlFor="currentPassword" className={styles.label}>
              Kata Sandi Saat Ini <span className={styles.required}>*</span>
            </label>
            <div className={styles.inputWrap}>
              <Lock size={16} className={styles.inputIcon} />
              <input
                id="currentPassword"
                type={showCurrent ? 'text' : 'password'}
                className={styles.input}
                placeholder="Masukkan kata sandi saat ini"
                autoComplete="current-password"
                aria-invalid={!!errors.currentPassword}
                {...field('currentPassword')}
                value={String(values.currentPassword)}
              />
              <button
                type="button"
                className={styles.eyeBtn}
                onClick={() => setShowCurrent((v) => !v)}
                aria-label={showCurrent ? 'Sembunyikan' : 'Tampilkan'}
              >
                {showCurrent ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
            {errors.currentPassword && (
              <p className={styles.error}>{errors.currentPassword}</p>
            )}
          </div>

          {/* Kata Sandi Baru */}
          <div className={styles.field}>
            <label htmlFor="newPassword" className={styles.label}>
              Kata Sandi Baru <span className={styles.required}>*</span>
            </label>
            <div className={styles.inputWrap}>
              <Lock size={16} className={styles.inputIcon} />
              <input
                id="newPassword"
                type={showNew ? 'text' : 'password'}
                className={styles.input}
                placeholder={`Minimal ${MIN_PASSWORD_LENGTH} karakter`}
                autoComplete="new-password"
                aria-invalid={!!errors.newPassword}
                {...field('newPassword')}
                value={String(values.newPassword)}
              />
              <button
                type="button"
                className={styles.eyeBtn}
                onClick={() => setShowNew((v) => !v)}
                aria-label={showNew ? 'Sembunyikan' : 'Tampilkan'}
              >
                {showNew ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
            {errors.newPassword && (
              <p className={styles.error}>{errors.newPassword}</p>
            )}

            {/* Strength indicator */}
            {values.newPassword && (
              <div className={styles.strengthWrap}>
                <div className={styles.strengthBars}>
                  {([1, 2, 3] as const).map((level) => (
                    <div
                      key={level}
                      className={`${styles.strengthBar} ${styles[`strengthBar_${strength >= level ? getStrengthClass(strength) : 'empty'}`]}`}
                    />
                  ))}
                </div>
                {strengthLabel && (
                  <span className={`${styles.strengthText} ${styles[`strengthText_${getStrengthClass(strength)}`]}`}>
                    {strengthLabel}
                  </span>
                )}
              </div>
            )}
          </div>

          {/* Konfirmasi Kata Sandi */}
          <div className={styles.field}>
            <label htmlFor="confirmPassword" className={styles.label}>
              Konfirmasi Kata Sandi Baru <span className={styles.required}>*</span>
            </label>
            <div className={styles.inputWrap}>
              <Lock size={16} className={styles.inputIcon} />
              <input
                id="confirmPassword"
                type={showConfirm ? 'text' : 'password'}
                className={styles.input}
                placeholder="Ulangi kata sandi baru"
                autoComplete="new-password"
                aria-invalid={!!errors.confirmPassword}
                {...field('confirmPassword')}
                value={String(values.confirmPassword)}
              />
              <button
                type="button"
                className={styles.eyeBtn}
                onClick={() => setShowConfirm((v) => !v)}
                aria-label={showConfirm ? 'Sembunyikan' : 'Tampilkan'}
              >
                {showConfirm ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
            {errors.confirmPassword && (
              <p className={styles.error}>{errors.confirmPassword}</p>
            )}
          </div>

          {/* Server error */}
          {serverError && (
            <div className={styles.serverError} role="alert">
              {serverError}
            </div>
          )}

          <button
            type="submit"
            className={styles.submitBtn}
            disabled={isSubmitting || hasValidationErrors || isFormEmpty}
          >
            {isSubmitting ? <span className={styles.spinner} /> : 'Ubah Kata Sandi'}
          </button>
        </form>
      </div>
    </div>
  )
}

// ─── Strength class helper (kept outside component for zero re-render overhead)

function getStrengthClass(strength: PasswordStrength): string {
  if (strength === 3) return 'strong'
  if (strength === 2) return 'medium'
  return 'weak'
}
