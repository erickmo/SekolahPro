/**
 * Reusable field validators for use with `useForm`.
 *
 * Each validator returns an error string or null.
 * Compose them with `composeValidators` for multiple rules per field.
 *
 * @example
 * validate: (values) => ({
 *   email: composeValidators(values.email, [required(), email()]),
 *   name: composeValidators(values.name, [required(), minLength(3)]),
 * })
 */

type Validator = (value: unknown) => string | null

// ─── Primitive Validators ─────────────────────────────────────────────────────

export const required =
  (message = 'Wajib diisi'): Validator =>
  (value) => {
    if (value === null || value === undefined) return message
    if (typeof value === 'string' && value.trim() === '') return message
    if (Array.isArray(value) && value.length === 0) return message
    return null
  }

export const minLength =
  (min: number, message?: string): Validator =>
  (value) => {
    const str = String(value ?? '')
    return str.length < min
      ? (message ?? `Minimal ${min} karakter`)
      : null
  }

export const maxLength =
  (max: number, message?: string): Validator =>
  (value) => {
    const str = String(value ?? '')
    return str.length > max
      ? (message ?? `Maksimal ${max} karakter`)
      : null
  }

export const minValue =
  (min: number, message?: string): Validator =>
  (value) => {
    const n = Number(value)
    return isNaN(n) || n < min
      ? (message ?? `Nilai minimal ${min}`)
      : null
  }

export const maxValue =
  (max: number, message?: string): Validator =>
  (value) => {
    const n = Number(value)
    return isNaN(n) || n > max
      ? (message ?? `Nilai maksimal ${max}`)
      : null
  }

// ─── Format Validators ────────────────────────────────────────────────────────

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export const email =
  (message = 'Format email tidak valid'): Validator =>
  (value) => {
    if (!value) return null // Let required() handle empty
    return EMAIL_RE.test(String(value)) ? null : message
  }

const PHONE_RE = /^(\+62|62|0)[0-9]{8,13}$/

export const phone =
  (message = 'Format nomor telepon tidak valid'): Validator =>
  (value) => {
    if (!value) return null
    return PHONE_RE.test(String(value).replace(/[\s\-().]/g, '')) ? null : message
  }

const URL_RE = /^https?:\/\/.+/

export const url =
  (message = 'Format URL tidak valid'): Validator =>
  (value) => {
    if (!value) return null
    return URL_RE.test(String(value)) ? null : message
  }

export const numeric =
  (message = 'Hanya boleh angka'): Validator =>
  (value) => {
    if (!value && value !== 0) return null
    return isNaN(Number(value)) ? message : null
  }

// ─── Match Validator ──────────────────────────────────────────────────────────

export const matches =
  (pattern: RegExp, message = 'Format tidak valid'): Validator =>
  (value) => {
    if (!value) return null
    return pattern.test(String(value)) ? null : message
  }

// ─── Composer ────────────────────────────────────────────────────────────────

/**
 * Run validators in order; return the first error or null.
 */
export function composeValidators(value: unknown, validators: Validator[]): string | null {
  for (const validate of validators) {
    const error = validate(value)
    if (error) return error
  }
  return null
}
