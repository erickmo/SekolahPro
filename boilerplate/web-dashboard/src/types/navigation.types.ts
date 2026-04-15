import type { LucideIcon } from 'lucide-react'

export interface NavItem {
  key: string
  label: string
  icon: LucideIcon
  path: string
  badge?: string | number
}

export interface SubNavItem {
  key: string
  label?: string
  href?: string
  badge?: string | number
  children?: DropdownNavItem[]
  divider?: true
  columns?: number
  exact?: boolean
}

export interface DropdownNavItem {
  key: string
  label: string
  href?: string
  badge?: string | number
  disabled?: boolean
  comingSoon?: boolean
  caption?: true
  divider?: true
}

export interface Breadcrumb {
  label: string
  href?: string
}
