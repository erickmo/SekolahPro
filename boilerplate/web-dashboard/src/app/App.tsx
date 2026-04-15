import { RouterProvider } from 'react-router-dom'
import { useEffect } from 'react'
import { Providers } from './providers'
import { router } from './routes'
import { useUiStore } from '@/stores/ui.store'

function ThemeInitializer() {
  const { theme, setTheme } = useUiStore()
  useEffect(() => {
    setTheme(theme)
  }, [theme, setTheme])
  return null
}

export function App() {
  return (
    <Providers>
      <ThemeInitializer />
      <RouterProvider router={router} />
    </Providers>
  )
}
