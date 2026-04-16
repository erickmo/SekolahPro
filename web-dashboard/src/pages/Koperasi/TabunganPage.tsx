import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { tabunganService } from '@/services/koperasi.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Tabungan, TabunganProduk, TabunganStatus } from '@/types/koperasi.types'
import { TABUNGAN_PRODUK_LABELS, TABUNGAN_STATUS_LABELS } from '@/types/koperasi.types'
import styles from './TabunganPage.module.css'

const PAGE_SIZE = 20

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

const PRODUK_OPTIONS: { value: TabunganProduk | ''; label: string }[] = [
  { value: '', label: 'Semua Produk' },
  { value: 'regular', label: 'Regular' },
  { value: 'education', label: 'Pendidikan' },
  { value: 'holiday', label: 'Liburan' },
  { value: 'qurban', label: 'Qurban' },
  { value: 'goal', label: 'Goal' },
]

const STATUS_OPTIONS: { value: TabunganStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'active', label: 'Aktif' },
  { value: 'dormant', label: 'Dorman' },
  { value: 'frozen', label: 'Dibekukan' },
  { value: 'closed', label: 'Ditutup' },
]

function getStatusClass(status: TabunganStatus): string {
  switch (status) {
    case 'active': return styles.badgeActive
    case 'dormant': return styles.badgeDormant
    case 'frozen': return styles.badgeFrozen
    case 'closed': return styles.badgeClosed
    default: return styles.badgeDefault
  }
}

function getProdukClass(produk: TabunganProduk): string {
  switch (produk) {
    case 'regular': return styles.badgeRegular
    case 'education': return styles.badgeEducation
    case 'holiday': return styles.badgeHoliday
    case 'qurban': return styles.badgeQurban
    case 'goal': return styles.badgeGoal
    default: return styles.badgeDefault
  }
}

function getGoalPercentage(saldo: number, target: number): number {
  if (target <= 0) return 0
  return Math.min(Math.round((saldo / target) * 100), 100)
}

export default function TabunganPage() {
  const [search, setSearch] = useState('')
  const [produk, setProduk] = useState<TabunganProduk | ''>('')
  const [status, setStatus] = useState<TabunganStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(produk ? { produk } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.tabungan, filters],
    queryFn: () => tabunganService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => tabunganService.delete(id),
    onSuccess: () => {
      toast.success('Tabungan berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.tabungan] })
    },
    onError: () => {
      toast.error('Gagal menghapus tabungan')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || produk !== '' || status !== ''

  const totalSaldo = items.reduce((sum, item) => sum + item.saldo, 0)
  const totalRekeningAktif = items.filter((item) => item.status === 'active').length
  const totalTargetGoal = items
    .filter((item) => item.targetGoal && item.targetGoal > 0)
    .reduce((sum, item) => sum + (item.targetGoal ?? 0), 0)

  function handleReset() {
    setSearch('')
    setProduk('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Tabungan) {
    if (!confirm(`Hapus tabungan "${item.noRekening}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Tabungan</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Tabungan
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Saldo</div>
          <div className={styles.summaryValue}>{formatRupiah(totalSaldo)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Rekening Aktif</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{totalRekeningAktif}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Target Goal</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{formatRupiah(totalTargetGoal)}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nasabah atau no. rekening..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={produk}
          onChange={(e) => { setProduk(e.target.value as TabunganProduk | ''); setPage(1) }}
          aria-label="Filter produk"
        >
          {PRODUK_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as TabunganStatus | ''); setPage(1) }}
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
                <th className={styles.th}>No. Rekening</th>
                <th className={styles.th}>Produk</th>
                <th className={styles.th}>Saldo</th>
                <th className={styles.th}>Target</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 7 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={7} className={styles.emptyCell}>
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada tabungan yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.nasabahNama || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.noRekening}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getProdukClass(item.produk)}`}>
                        {TABUNGAN_PRODUK_LABELS[item.produk] ?? item.produk}
                      </span>
                    </td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.saldo)}</span></td>
                    <td className={styles.td}>
                      {item.targetGoal && item.targetGoal > 0 ? (
                        <div className={styles.goalCell}>
                          <span className={styles.goalAmount}>{formatRupiah(item.targetGoal)}</span>
                          <div className={styles.progressTrack}>
                            <div
                              className={styles.progressFill}
                              style={{ width: `${getGoalPercentage(item.saldo, item.targetGoal)}%` }}
                            />
                          </div>
                          <span className={styles.progressLabel}>{getGoalPercentage(item.saldo, item.targetGoal)}%</span>
                        </div>
                      ) : (
                        <span className={styles.dash}>&mdash;</span>
                      )}
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {TABUNGAN_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus tabungan ${item.noRekening}`}
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
