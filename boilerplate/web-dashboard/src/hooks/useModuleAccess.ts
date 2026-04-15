import { useLocation } from 'react-router-dom'
import { useDashboardContext } from './useDashboardContext'

type ModuleAccessLevel = 'full' | 'readonly'

/** HQ: transaksi hanya bisa dilihat */
const HQ_READONLY_MODULES = new Set(['sales', 'purchasing', 'inventory'])

/** Company: master data & regulasi hanya bisa dilihat */
const COMPANY_READONLY_MODULES = new Set(['master-data', 'regulasi'])

export interface ModuleAccess {
  readonly: boolean
  managedByHQ: boolean
}

/**
 * Hook untuk mendeteksi apakah modul saat ini bersifat readonly.
 *
 * Aturan:
 * - HQ context  → sales, purchasing, inventory = readonly ("Hanya Baca")
 * - Company ctx → master-data, regulasi = managedByHQ ("Dikelola HQ")
 * - Semua modul lain = full access
 */
export function useModuleAccess(): ModuleAccess {
  const location = useLocation()
  const ctx = useDashboardContext()

  const segments = location.pathname.split('/').filter(Boolean)
  // /g/<module>/...        → segments[1]
  // /c/:kode/<module>/...  → segments[2]
  const moduleSegment = ctx === 'hq' ? segments[1] : segments[2]

  if (ctx === 'hq') {
    return {
      readonly: HQ_READONLY_MODULES.has(moduleSegment ?? ''),
      managedByHQ: false,
    }
  }

  return {
    readonly: false,
    managedByHQ: COMPANY_READONLY_MODULES.has(moduleSegment ?? ''),
  }
}
