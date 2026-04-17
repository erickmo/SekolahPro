import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { jaminanService } from '@/services/phase5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Jaminan, JaminanCollateralType, JaminanStatus } from '@/types/phase5.types'
import { JAMINAN_COLLATERAL_TYPE_LABELS, JAMINAN_STATUS_LABELS } from '@/types/phase5.types'
import styles from './JaminanPage.module.css'

const PAGE_SIZE = 20

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

const COLLATERAL_TYPE_OPTIONS: { value: JaminanCollateralType | ''; label: string }[] = [
  { value: '', label: 'Semua Jenis' },
  { value: 'property', label: 'Properti' },
  { value: 'vehicle', label: 'Kendaraan' },
  { value: 'gold', label: 'Emas' },
  { value: 'land', label: 'Tanah' },
  { value: 'savings', label: 'Simpanan' },
  { value: 'guarantor', label: 'Penjamin' },
  { value: 'other', label: 'Lainnya' },
]

const STATUS_OPTIONS: { value: JaminanStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'verified', label: 'Terverifikasi' },
  { value: 'pending', label: 'Pending' },
  { value: 'rejected', label: 'Ditolak' },
  { value: 'released', label: 'Dilepas' },
  { value: 'foreclosed', label: 'Dieksekusi' },
]

function getCollateralTypeClass(collateralType: JaminanCollateralType): string {
  switch (collateralType) {
    case 'property': return styles.badgeProperty
    case 'vehicle': return styles.badgeVehicle
    case 'gold': return styles.badgeGold
    case 'land': return styles.badgeLand
    case 'savings': return styles.badgeSavings
    case 'guarantor': return styles.badgeGuarantor
    default: return styles.badgeOther
  }
}

function getStatusClass(status: JaminanStatus): string {
  switch (status) {
    case 'verified': return styles.badgeVerified
    case 'pending': return styles.badgePending
    case 'rejected': return styles.badgeRejected
    case 'released': return styles.badgeReleased
    case 'foreclosed': return styles.badgeForeclosed
    default: return styles.badgeDefault
  }
}

export default function JaminanPage() {
  const [search, setSearch] = useState('')
  const [collateralType, setCollateralType] = useState<JaminanCollateralType | ''>('')
  const [status, setStatus] = useState<JaminanStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(collateralType ? { collateral_type: collateralType } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.jaminan, filters],
    queryFn: () => jaminanService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => jaminanService.delete(id),
    onSuccess: () => {
      toast.success('Jaminan berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.jaminan] })
    },
    onError: () => {
      toast.error('Gagal menghapus jaminan')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || collateralType !== '' || status !== ''

  const totalValue = items.reduce((sum, item) => sum + item.value, 0)
  const totalVerified = items.filter((item) => item.status === 'verified').length

  function handleReset() {
    setSearch('')
    setCollateralType('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Jaminan) {
    if (!confirm(`Hapus jaminan untuk "${item.nasabahNama}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Jaminan</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Jaminan
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Nilai</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{formatRupiah(totalValue)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Terverifikasi</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{totalVerified}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Data</div>
          <div className={styles.summaryValue}>{total}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nasabah atau deskripsi jaminan..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={collateralType}
          onChange={(e) => { setCollateralType(e.target.value as JaminanCollateralType | ''); setPage(1) }}
          aria-label="Filter jenis jaminan"
        >
          {COLLATERAL_TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as JaminanStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Nasabah</th>
                <th className={styles.th}>Jenis Jaminan</th>
                <th className={styles.th}>Deskripsi</th>
                <th className={styles.th}>Nilai</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 6 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={6} className={styles.emptyCell}>
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada jaminan yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.nasabahNama || '-'}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getCollateralTypeClass(item.collateralType)}`}>
                        {JAMINAN_COLLATERAL_TYPE_LABELS[item.collateralType] ?? item.collateralType}
                      </span>
                    </td>
                    <td className={styles.td}>{item.description || '-'}</td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.value)}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {JAMINAN_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus jaminan ${item.nasabahNama}`}
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
