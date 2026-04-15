import { useCallback, useMemo, useState } from 'react'

type FieldErrors<T> = Partial<Record<keyof T, string>>
type ValidateFn<T> = (values: T) => FieldErrors<T>

interface UseFormConfig<T> {
  initialValues: T
  validate?: ValidateFn<T>
}

interface FieldProps {
  value: unknown
  onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => void
  onBlur: () => void
}

interface UseFormReturn<T> {
  values: T
  errors: FieldErrors<T>
  touched: Partial<Record<keyof T, boolean>>
  isDirty: boolean
  setFieldValue: (key: keyof T, value: unknown) => void
  field: (key: keyof T) => FieldProps
  handleSubmit: (onSubmit: (values: T) => void | Promise<void>) => (e?: React.FormEvent) => Promise<void>
  reset: () => void
  setServerErrors: (errors: Record<string, string>) => void
}

export function useForm<T extends Record<string, unknown>>({
  initialValues,
  validate,
}: UseFormConfig<T>): UseFormReturn<T> {
  const [initialSnapshot] = useState(() => JSON.stringify(initialValues))
  const [values, setValues] = useState<T>(initialValues)
  const [errors, setErrors] = useState<FieldErrors<T>>({})
  const [touched, setTouched] = useState<Partial<Record<keyof T, boolean>>>({})

  const isDirty = useMemo(
    () => JSON.stringify(values) !== initialSnapshot,
    [values, initialSnapshot],
  )

  const setFieldValue = useCallback((key: keyof T, value: unknown) => {
    setValues((prev) => ({ ...prev, [key]: value }))
    setErrors((prev) => ({ ...prev, [key]: undefined }))
  }, [])

  const field = useCallback(
    (key: keyof T): FieldProps => ({
      value: values[key],
      onChange: (e) => setFieldValue(key, e.target.value),
      onBlur: () => {
        setTouched((prev) => ({ ...prev, [key]: true }))
        if (validate) {
          const result = validate(values)
          setErrors((prev) => ({ ...prev, [key]: result[key] }))
        }
      },
    }),
    [values, setFieldValue, validate],
  )

  const handleSubmit = useCallback(
    (onSubmit: (values: T) => void | Promise<void>) =>
      async (e?: React.FormEvent) => {
        e?.preventDefault()
        if (validate) {
          const result = validate(values)
          const hasErrors = Object.values(result).some(Boolean)
          if (hasErrors) {
            setErrors(result)
            const allTouched = Object.fromEntries(
              Object.keys(values).map((k) => [k, true]),
            ) as Partial<Record<keyof T, boolean>>
            setTouched(allTouched)
            return
          }
        }
        await onSubmit(values)
      },
    [values, validate],
  )

  const reset = useCallback(() => {
    setValues(JSON.parse(initialSnapshot) as T)
    setErrors({})
    setTouched({})
  }, [initialSnapshot])

  const setServerErrors = useCallback((errs: Record<string, string>) => {
    setErrors(errs as FieldErrors<T>)
  }, [])

  return { values, errors, touched, isDirty, setFieldValue, field, handleSubmit, reset, setServerErrors }
}
