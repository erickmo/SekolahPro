import { useLocation } from 'react-router-dom'

// ─── Route label map ──────────────────────────────────────────────────────────
// Add your route labels here to display human-readable breadcrumbs

export const ROUTE_LABELS: Record<string, string> = {
  dashboard: 'Dashboard',
  settings: 'Pengaturan',
  profile: 'Profil',
  users: 'Pengguna',
  new: 'Tambah Baru',
  edit: 'Edit',
  // Master data
  'master-data': 'Master Data',
  'item-groups': 'Kelompok Item',
  items: 'Item',
  customers: 'Pelanggan',
  suppliers: 'Pemasok',
  uom: 'Satuan',
  brands: 'Merek',
  // Sales
  sales: 'Penjualan',
  orders: 'Pesanan',
  invoices: 'Faktur',
  // Purchasing
  purchasing: 'Pembelian',
  requests: 'Permintaan',
  // Inventory
  inventory: 'Inventori',
  stock: 'Stok',
  warehouses: 'Gudang',
  transfers: 'Transfer',
  // Accounting
  accounting: 'Akuntansi',
  journal: 'Entri Jurnal',
  coa: 'Bagan Akun',
  payments: 'Pembayaran',
  // HRM
  hrm: 'SDM',
  employees: 'Karyawan',
  payroll: 'Penggajian',
  attendance: 'Kehadiran',
  // HQ routes
  g: 'HQ Dashboard',
  keuangan: 'Keuangan',
  regulasi: 'Regulasi',
  perusahaan: 'Perusahaan',
  harga: 'Daftar Harga',
  voucher: 'Voucher Otomasi',
  'template-pajak': 'Template Pajak',
  'tenant-owners': 'Pemilik Perusahaan',
  'company-groups': 'Grup Perusahaan',
}

export interface Breadcrumb {
  label: string
  href?: string
}

// Multi-tenant path prefixes to skip when building breadcrumbs
const SKIP_PREFIXES = ['c', 'g'] as const

/**
 * Generates breadcrumbs from the current pathname.
 * Handles multi-tenant prefixes (/c/:companyCode and /g) automatically.
 *
 * @param lastLabel - Optional override for the last crumb's label (e.g. the record title)
 */
export function useAutoBreadcrumbs(lastLabel?: string): Breadcrumb[] {
  const location = useLocation()
  let parts = location.pathname.split('/').filter(Boolean)

  // Strip multi-tenant prefix (/c/:code or /g/:groupId) and remember it
  // so breadcrumb links stay within the same context
  let pathPrefix = ''
  if (parts.length > 0 && (SKIP_PREFIXES as readonly string[]).includes(parts[0])) {
    pathPrefix = `/${parts[0]}/${parts[1]}`
    parts = parts.slice(2)
  }

  const crumbs: Breadcrumb[] = [{ label: 'Dashboard', href: `${pathPrefix}/` }]
  let path = pathPrefix

  parts.forEach((part, i) => {
    path += `/${part}`
    const isLast = i === parts.length - 1
    const label = isLast && lastLabel
      ? lastLabel
      : (ROUTE_LABELS[part] ?? part.charAt(0).toUpperCase() + part.slice(1))

    if (!isLast) {
      crumbs.push({ label, href: path })
    } else {
      crumbs.push({ label })
    }
  })

  return crumbs
}
