import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@/theme/variables.css'
import '@/theme/reset.css'
import '@/theme/typography.css'
import '@/theme/motion.css'
import { App } from '@/app/App'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
