import { useState, useRef, useEffect } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import {
  LayoutDashboard, Users, Settings, Bell,
  ChevronDown, LogOut,
  CheckCircle, AlertCircle, AlertTriangle, Info,
  Building2, Shield, Globe,
} from 'lucide-react'
import { useAuthStore } from '@/stores/auth.store'
import { useNotificationStore } from '@/stores/notification.store'
import { getInitials, formatRelative } from '@/utils/format'
import { appConfig } from '@/config/app.config'
import { cn } from '@/utils/cn'
import type { AppContext } from '@/layouts/AppShell/AppShell'
import styles from './AppNavbar.module.css'

// ─── Nav item definitions per context ────────────────────────────────────────

const NAV_ITEMS_DEFAULT = [
  { key: 'dashboard', label: 'Dashboard', icon: LayoutDashboard, path: 'dashboard' },
  { key: 'users', label: 'Pengguna', icon: Users, path: 'users' },
  { key: 'settings', label: 'Pengaturan', icon: Settings, path: 'settings' },
]

const NAV_ITEMS_SUPERUSER = [
  { key: 'dashboard', label: 'Dashboard', icon: LayoutDashboard, path: 'dashboard' },
  { key: 'tenants', label: 'Tenants', icon: Globe, path: 'tenants' },
  { key: 'companies', label: 'Perusahaan', icon: Building2, path: 'companies' },
  { key: 'settings', label: 'Pengaturan', icon: Settings, path: 'settings' },
]

const NAV_ITEMS_HQ = [
  { key: 'dashboard', label: 'HQ Dashboard', icon: LayoutDashboard, path: 'dashboard' },
  { key: 'users', label: 'Pengguna', icon: Users, path: 'users' },
  { key: 'settings', label: 'Pengaturan', icon: Settings, path: 'settings' },
]

const NAV_ITEMS_COMPANY = [
  { key: 'dashboard', label: 'Dashboard', icon: LayoutDashboard, path: 'dashboard' },
  { key: 'users', label: 'Pengguna', icon: Users, path: 'users' },
  { key: 'settings', label: 'Pengaturan', icon: Settings, path: 'settings' },
]

const NAV_ITEMS_BY_CONTEXT = {
  default: NAV_ITEMS_DEFAULT,
  superuser: NAV_ITEMS_SUPERUSER,
  hq: NAV_ITEMS_HQ,
  company: NAV_ITEMS_COMPANY,
} as const

const BASE_PATH_BY_CONTEXT = {
  default: '',
  superuser: '/su',
  hq: '/g',
  company: '', // resolved dynamically from selectedCompany.code
} as const

// ─── Notification icons ───────────────────────────────────────────────────────

const NOTIF_ICONS = {
  success: <CheckCircle size={14} />,
  error: <AlertCircle size={14} />,
  warning: <AlertTriangle size={14} />,
  info: <Info size={14} />,
}

// ─── Component ────────────────────────────────────────────────────────────────

interface AppNavbarProps {
  context?: AppContext
}

export function AppNavbar({ context = 'default' }: AppNavbarProps) {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, selectedCompany, selectedGroup, logout } = useAuthStore()
  const { items: notifications, unreadCount, markAllRead, markRead } = useNotificationStore()

  const [showNotif, setShowNotif] = useState(false)
  const [showProfile, setShowProfile] = useState(false)

  const notifRef = useRef<HTMLDivElement>(null)
  const profileRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (notifRef.current && !notifRef.current.contains(e.target as Node)) setShowNotif(false)
      if (profileRef.current && !profileRef.current.contains(e.target as Node)) setShowProfile(false)
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  const navItems = NAV_ITEMS_BY_CONTEXT[context]
  const basePath = context === 'company'
    ? `/c/${selectedCompany?.code ?? ''}`
    : BASE_PATH_BY_CONTEXT[context]

  const isActive = (path: string) => {
    const full = `${basePath}/${path}`
    if (path === 'dashboard') return location.pathname === full || location.pathname === basePath
    return location.pathname.startsWith(full)
  }

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  // Workspace label for multi-tenant company switcher
  const workspaceLabel = context === 'hq'
    ? selectedGroup?.name
    : context === 'company'
      ? selectedCompany?.name
      : null

  return (
    <nav className={cn(
      styles.navbar,
      context === 'superuser' && styles.navbarSuperuser,
      context === 'hq' && styles.navbarHq,
    )}>
      {/* Logo */}
      <Link to={basePath ? `${basePath}/dashboard` : '/dashboard'} className={styles.logo}>
        {context === 'superuser' && <Shield size={16} className={styles.logoContextIcon} />}
        {context === 'hq' && <Globe size={16} className={styles.logoContextIcon} />}
        {(context === 'default' || context === 'company') && (
          <span className={styles.logoIcon}>{appConfig.appName[0]}</span>
        )}
        <span className={styles.logoText}>{appConfig.appName}</span>
      </Link>

      {/* Main nav */}
      <ul className={styles.navList}>
        {navItems.map((item) => (
          <li key={item.key}>
            <Link
              to={`${basePath}/${item.path}`}
              className={cn(styles.navItem, isActive(item.path) && styles.navItemActive)}
            >
              <item.icon size={16} />
              <span className={styles.navLabel}>{item.label}</span>
            </Link>
          </li>
        ))}
      </ul>

      {/* Right zone */}
      <div className={styles.right}>
        {/* Company / workspace switcher (multi-tenant only) */}
        {appConfig.isMultiTenant && workspaceLabel && (
          <Link to="/choose-company" className={styles.workspaceBtn}>
            <Building2 size={14} />
            <span className={styles.workspaceName}>{workspaceLabel}</span>
            <ChevronDown size={12} />
          </Link>
        )}

        {/* Notifications */}
        <div className={styles.iconWrap} ref={notifRef}>
          <button
            className={styles.iconBtn}
            onClick={() => setShowNotif(!showNotif)}
            aria-label="Notifikasi"
          >
            <Bell size={18} />
            {unreadCount > 0 && (
              <span className={styles.badge}>{unreadCount > 9 ? '9+' : unreadCount}</span>
            )}
          </button>
          {showNotif && (
            <div className={styles.notifDropdown}>
              <div className={styles.notifHeader}>
                <span>Notifikasi</span>
                {unreadCount > 0 && (
                  <button className={styles.markAllBtn} onClick={markAllRead}>
                    Tandai semua dibaca
                  </button>
                )}
              </div>
              <div className={styles.notifList}>
                {notifications.length === 0 && (
                  <p className={styles.notifEmpty}>Tidak ada notifikasi</p>
                )}
                {notifications.slice(0, 6).map((n) => (
                  <button
                    key={n.id}
                    className={cn(styles.notifItem, !n.isRead && styles.notifUnread)}
                    onClick={() => { markRead(n.id); setShowNotif(false) }}
                  >
                    <span className={cn(styles.notifIcon, styles[`notifIcon_${n.type}`])}>
                      {NOTIF_ICONS[n.type]}
                    </span>
                    <div className={styles.notifContent}>
                      <span className={styles.notifTitle}>{n.title}</span>
                      <span className={styles.notifMessage}>{n.message}</span>
                      <span className={styles.notifTime}>{formatRelative(n.createdAt)}</span>
                    </div>
                    {!n.isRead && <span className={styles.unreadDot} />}
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Profile */}
        <div className={styles.iconWrap} ref={profileRef}>
          <button className={styles.avatarBtn} onClick={() => setShowProfile(!showProfile)}>
            <span className={styles.avatar}>{getInitials(user?.name ?? 'U')}</span>
            <ChevronDown size={12} className={styles.avatarChevron} />
          </button>
          {showProfile && (
            <div className={styles.profileDropdown}>
              <div className={styles.profileInfo}>
                <span className={styles.profileAvatar}>{getInitials(user?.name ?? 'U')}</span>
                <div>
                  <p className={styles.profileName}>{user?.name}</p>
                  <p className={styles.profileEmail}>{user?.email}</p>
                  <span className={styles.roleBadge}>{user?.role}</span>
                </div>
              </div>
              <hr className={styles.divider} />
              <Link
                to={`${basePath}/settings`}
                className={styles.dropdownItem}
                onClick={() => setShowProfile(false)}
              >
                <Settings size={14} />
                Pengaturan
              </Link>
              <button className={cn(styles.dropdownItem, styles.logoutItem)} onClick={handleLogout}>
                <LogOut size={14} />
                Keluar
              </button>
            </div>
          )}
        </div>
      </div>
    </nav>
  )
}
