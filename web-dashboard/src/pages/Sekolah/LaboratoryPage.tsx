import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { labService } from '@/services/sprint7.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Laboratory, LabType, LabStatus, SafetyRating } from '@/types/sprint7.types'
import { LAB_TYPE_LABELS, LAB_STATUS_LABELS, SAFETY_RATING_LABELS } from '@/types/sprint7.types'
import styles from './LaboratoryPage.module.css'

const PAGE_SIZE = 20

const TYPE_OPTIONS: { value: LabType | ''; label: string }[] = [
  { value: '', label: 'Semua Tipe' },
  ...Object.entries(LAB_TYPE_LABELS).map(([value, label]) => ({
    value: value as LabType,
    label,
  })),
]

const STATUS_OPTIONS: { value: LabStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(LAB_STATUS_LABELS).map(([value, label]) => ({
    value: value as LabStatus,
    label,
  })),
]

function getStatusBadgeClass(status: LabStatus): string {
  switch (status) {
    case 'active': return styles.badgeGreen
    case 'inactive': return styles.badgeGray
    case 'maintenance': return styles.badgeYellow
    default: return styles.badgeDefault
  }
}

function getSafetyBadgeClass(rating: SafetyRating): string {
  switch (rating) {
    case 'excellent': return styles.badgeGreen
    case 'good': return styles.badgeBlue
    case 'fair': return styles.badgeYellow
    case 'poor': return styles.badgeRed
    default: return styles.badgeDefault
  }
}

function getTypeBadgeClass(type: LabType): string {
  switch (type) {
    case 'ipa': return styles.badgeBlue
    case 'komputer': return styles.badgeGreen
    case 'bahasa': return styles.badgeYellow
    case 'multimedia': return styles.badgeDefault
    default: return styles.badgeDefault
  }
}

export default function LaboratoryPage() {
  const [search, setSearch] = useState('')
  const [type, setType] = useState<LabType | ''>('')
  const [status, setStatus] = useState<LabStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(type ? { type } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.laboratories, filters],
    queryFn: () => labService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => labService.delete(id),
    onSuccess: () => {
      toast.success('Data laboratorium berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.laboratories] })
    },
    onError: () => {
      toast.error('Gagal menghapus data laboratorium')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || type !== '' || status !== ''

  function handleReset() {
    setSearch('')
    setType('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Laboratory) {
    if (!confirm(`Hapus laboratorium "${item.name}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Laboratorium</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Lab
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari laboratorium..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={type}
          onChange={(e) => { setType(e.target.value as LabType | ''); setPage(1) }}
          aria-label="Filter tipe"
        >
          {TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as LabStatus | ''); setPage(1) }}
          aria-label="Filter status"
        >
          {STATUS_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Nama</th>
                <th className={styles.th}>Tipe</th>
                <th className={styles.th}>Kapasitas</th>
                <th className={styles.th}>Gedung</th>
                <th className={styles.th}>Lantai</th>
                <th className={styles.th}>No. Ruang</th>
                <th className={styles.th}>Jumlah Alat</th>
                <th className={styles.th}>Rating Keselamatan</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 10 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={10} className={styles.emptyCell}>
                    <div>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada data laboratorium.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.name || '-'}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getTypeBadgeClass(item.type)}`}>
                        {LAB_TYPE_LABELS[item.type] ?? item.type}
                      </span>
                    </td>
                    <td className={styles.td}>{item.capacity ?? '-'}</td>
                    <td className={styles.td}>{item.building || '-'}</td>
                    <td className={styles.td}>{item.floor ?? '-'}</td>
                    <td className={styles.td}><span className={styles.code}>{item.roomNumber || '-'}</span></td>
                    <td className={styles.td}>{item.equipmentCount}</td>
                    <td className={styles.td}>
                      {item.safetyRating ? (
                        <span className={`${styles.badge} ${getSafetyBadgeClass(item.safetyRating)}`}>
                          {SAFETY_RATING_LABELS[item.safetyRating]}
                        </span>
                      ) : '-'}
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusBadgeClass(item.status)}`}>
                        {LAB_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus lab ${item.name}`}
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
