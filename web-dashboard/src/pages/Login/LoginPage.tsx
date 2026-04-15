import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Eye, EyeOff, Lock, Mail } from 'lucide-react'
import { useAuthStore } from '@/stores/auth.store'
import { authService } from '@/services/auth.service'
import { useForm } from '@/hooks/useForm'
import { AppApiError } from '@/types/api.types'
import styles from './LoginPage.module.css'

interface LoginFormValues {
  email: string
  password: string
}

export default function LoginPage() {
  const navigate = useNavigate()
  const login = useAuthStore((s) => s.login)
  const [showPassword, setShowPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [serverError, setServerError] = useState('')

  const { values, errors, field, handleSubmit } = useForm<LoginFormValues>({
    initialValues: { email: '', password: '' },
    validate: (v) => ({
      email: !v.email ? 'Email wajib diisi' : undefined,
      password: !v.password ? 'Kata sandi wajib diisi' : undefined,
    }),
  })

  const onSubmit = handleSubmit(async (v) => {
    setIsLoading(true)
    setServerError('')
    try {
      const response = await authService.login(v)
      login(response)
      navigate('/dashboard', { replace: true })
    } catch (err) {
      if (err instanceof AppApiError) {
        setServerError(err.message)
      } else {
        setServerError('Terjadi kesalahan, coba lagi')
      }
    } finally {
      setIsLoading(false)
    }
  })

  return (
    <div className={styles.root}>
      <div className={styles.card}>
        {/* Logo */}
        <div className={styles.logoWrap}>
          <span className={styles.logoIcon}>D</span>
          <span className={styles.logoText}>Dashboard</span>
        </div>

        <h1 className={styles.title}>Masuk ke Akun Anda</h1>
        <p className={styles.subtitle}>Masukkan email dan kata sandi untuk melanjutkan</p>

        <form onSubmit={onSubmit} className={styles.form} noValidate>
          {/* Email */}
          <div className={styles.field}>
            <label htmlFor="email" className={styles.label}>Email</label>
            <div className={styles.inputWrap}>
              <Mail size={16} className={styles.inputIcon} />
              <input
                id="email"
                type="email"
                autoComplete="email"
                placeholder="nama@perusahaan.com"
                className={styles.input}
                aria-invalid={!!errors.email}
                {...field('email')}
                value={String(values.email)}
              />
            </div>
            {errors.email && <p className={styles.error}>{errors.email}</p>}
          </div>

          {/* Password */}
          <div className={styles.field}>
            <label htmlFor="password" className={styles.label}>Kata Sandi</label>
            <div className={styles.inputWrap}>
              <Lock size={16} className={styles.inputIcon} />
              <input
                id="password"
                type={showPassword ? 'text' : 'password'}
                autoComplete="current-password"
                placeholder="Masukkan kata sandi"
                className={styles.input}
                aria-invalid={!!errors.password}
                {...field('password')}
                value={String(values.password)}
              />
              <button
                type="button"
                className={styles.eyeBtn}
                onClick={() => setShowPassword(!showPassword)}
                aria-label={showPassword ? 'Sembunyikan kata sandi' : 'Tampilkan kata sandi'}
              >
                {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
            {errors.password && <p className={styles.error}>{errors.password}</p>}
          </div>

          {serverError && (
            <div className={styles.serverError} role="alert">
              {serverError}
            </div>
          )}

          <button
            type="submit"
            className={styles.submitBtn}
            disabled={isLoading}
          >
            {isLoading ? (
              <span className={styles.spinner} />
            ) : (
              'Masuk'
            )}
          </button>
        </form>
      </div>
    </div>
  )
}
