import { Outlet } from 'react-router-dom'
import { AppNavbar } from '@/layouts/AppNavbar/AppNavbar'
import styles from './AppShell.module.css'

export type AppContext = 'default' | 'superuser' | 'hq' | 'company'

interface AppShellProps {
  /** Multi-tenant context — controls navbar color and nav items. */
  context?: AppContext
}

export function AppShell({ context = 'default' }: AppShellProps) {
  return (
    <div className={styles.root}>
      <AppNavbar context={context} />
      <main className={styles.main}>
        <Outlet />
      </main>
    </div>
  )
}
