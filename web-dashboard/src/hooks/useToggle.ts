import { useCallback, useState } from 'react'

export function useToggle(
  initialValue = false,
): [boolean, () => void, () => void, () => void, (v: boolean) => void] {
  const [value, setValue] = useState<boolean>(initialValue)

  const toggle = useCallback(() => setValue((prev) => !prev), [])
  const setTrue = useCallback(() => setValue(true), [])
  const setFalse = useCallback(() => setValue(false), [])
  const set = useCallback((v: boolean) => setValue(v), [])

  return [value, toggle, setTrue, setFalse, set]
}
