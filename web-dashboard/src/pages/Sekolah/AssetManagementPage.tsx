import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { assetService } from '@/services/sprint7.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Asset, AssetCategory, AssetCondition, OwnershipType, LifecycleStatus } from '@/types/sprint7.types'
import {
  ASSET_CATEGORY_LABELS,
  ASSET_CONDITION_LABELS,
  OWNERSHIP_TYPE_LABELS,
  LIFECYCLE_STATUS_LABELS,
} from '@/types/sprint7.types'
import styles from './AssetManagementPage.module.css'

const PAGE_SIZE = 20

const CATEGORY_OPTIONS: { value: AssetCategory | ''; label: string }[] = [
  { value: '', label: 'Semua Kategori' },
  ...Object.entries(ASSET_CATEGORY_LABELS).map(([value, label]) => ({
    value: value as AssetCategory,
    label,
  })),
]

const LIFECYCLE_OPTIONS: { value: LifecycleStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(LIFECYCLE_STATUS_LABELS).map(([value, label]) => ({
    value: value as LifecycleStatus,
    label,
  })),
]

function getCategoryBadgeClass(category: AssetCategory): string {
  switch (category) {
    case 'tanah': return styles.badgeGreen
    case 'bangunan': return styles.badgeBlue
    case 'mesin': return styles.badgeYellow
    case 'kendaraan': return styles.badgeDefault
    case 'peralatan': return styles.badgeDefault
    default: return styles.badgeGray
  }
}

function getConditionBadgeClass(condition: AssetCondition): string {
  switch (condition) {
    case 'new': return styles.badgeGreen
    case 'good': return styles.badgeBlue
    case 'fair': return styles.badgeYellow
    case 'poor': return styles.badgeRed
    case 'damaged': return styles.badgeRed
    default: return styles.badgeDefault
  }
}

function getOwnershipBadgeClass(ownership: OwnershipType): string {
  switch (ownership) {
    case 'owned': return styles.badgeGreen
    case 'leased': return styles.badgeBlue
    case 'donated': return styles.badgeYellow
    default: return styles.badgeDefault
  }
}

function getLifecycleBadgeClass(status: LifecycleStatus): string {
  switch (status) {
    case 'in_use': return styles.badgeGreen
    case 'idle': return styles.badgeGray
    case 'maintenance': return styles.badgeYellow
    case 'disposed': return styles.badgeRed
    case 'lost': return styles.badgeRed
    case 'transferred': return styles.badgeBlue
    default: return styles.badgeDefault
  }
}

function formatCurrency(amount: number | null): string {
  if (amount == null) return '-'
  return amount.toLocaleString('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 })
}

export default function AssetManagementPage() {
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState<AssetCategory | ''>('')
  const [lifecycleStatus, setLifecycleStatus] = useState<LifecycleStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(category ? { category } : {}),
    ...(lifecycleStatus ? { lifecycle_status: lifecycleStatus } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.assets, filters],
    queryFn: () => assetService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => assetService.delete(id),
    onSuccess: () => {
      toast.success('Data aset berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.assets] })
    },
    onError: () => {
      toast.error('Gagal menghapus data aset')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || category !== '' || lifecycleStatus !== ''

  function handleReset() {
    setSearch('')
    setCategory('')
    setLifecycleStatus('')
    setPage(1)
  }

  function handleDelete(item: Asset) {
    if (!confirm(`Hapus aset "${item.name}" (${item.assetCode})?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Manajemen Aset</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Aset
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari aset..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={category}
          onChange={(e) => { setCategory(e.target.value as AssetCategory | ''); setPage(1) }}
          aria-label="Filter kategori"
        >
          {CATEGORY_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={lifecycleStatus}
          onChange={(e) => { setLifecycleStatus(e.target.value as LifecycleStatus | ''); setPage(1) }}
          aria-label="Filter status"
        >
          {LIFECYCLE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        {hasFilters && (
          <button type="button" className={styles.resetBtn} onClick={handleReset}>
            <RotateCcw size={14} /> Reset Filter
          </button>
        )}
      </div>

      <div className={styles.tableCard}>
        <div className={styles.tableScroll}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.th}>Kode Aset</th>
                <th className={styles.th}>Nama</th>
                <th className={styles.th}>Kategori</th>
                <th className={styles.th}>Kondisi</th>
                <th className={styles.th}>Kepemilikan</th>
                <th className={styles.th}>Nilai Buku</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Lokasi</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 9 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={9} className={styles.emptyCell}>
                    <div>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada data aset.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.assetCode || '-'}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.name || '-'}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getCategoryBadgeClass(item.category)}`}>
                        {ASSET_CATEGORY_LABELS[item.category] ?? item.category}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getConditionBadgeClass(item.condition)}`}>
                        {ASSET_CONDITION_LABELS[item.condition] ?? item.condition}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getOwnershipBadgeClass(item.ownershipType)}`}>
                        {OWNERSHIP_TYPE_LABELS[item.ownershipType] ?? item.ownershipType}
                      </span>
                    </td>
                    <td className={styles.td}>{formatCurrency(item.bookValue)}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getLifecycleBadgeClass(item.lifecycleStatus)}`}>
                        {LIFECYCLE_STATUS_LABELS[item.lifecycleStatus] ?? item.lifecycleStatus}
                      </span>
                    </td>
                    <td className={styles.td}>{item.locationRoomName || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus aset ${item.name}`}
                      >
                        <Trash2 size={14} />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {!isLoading && totalPages > 1 && (
          <div className={styles.pagination}>
            <button type="button" className={styles.pageBtn} disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              Sebelumnya
            </button>
            <span className={styles.pageInfo}>Halaman {page} dari {totalPages}</span>
            <button type="button" className={styles.pageBtn} disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              Berikutnya
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
