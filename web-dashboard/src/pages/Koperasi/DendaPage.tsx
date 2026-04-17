import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { dendaService } from '@/services/phase5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Denda, DendaPenaltyType, DendaStatus } from '@/types/phase5.types'
import { DENDA_PENALTY_TYPE_LABELS, DENDA_STATUS_LABELS } from '@/types/phase5.types'
import styles from './DendaPage.module.css'

const PAGE_SIZE = 20

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

const PENALTY_TYPE_OPTIONS: { value: DendaPenaltyType | ''; label: string }[] = [
  { value: '', label: 'Semua Tipe' },
  { value: 'late_payment', label: 'Keterlambatan' },
  { value: 'early_withdrawal', label: 'Pencairan Awal' },
  { value: 'overdue', label: 'Tunggakan' },
  { value: 'administrative', label: 'Administratif' },
  { value: 'other', label: 'Lainnya' },
]

const STATUS_OPTIONS: { value: DendaStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'pending', label: 'Pending' },
  { value: 'paid', label: 'Lunas' },
  { value: 'waived', label: 'Dibebaskan' },
  { value: 'cancelled', label: 'Dibatalkan' },
]

function getPenaltyTypeClass(penaltyType: DendaPenaltyType): string {
  switch (penaltyType) {
    case 'late_payment': return styles.badgeLatePayment
    case 'early_withdrawal': return styles.badgeEarlyWithdrawal
    case 'overdue': return styles.badgeOverdue
    case 'administrative': return styles.badgeAdministrative
    default: return styles.badgeOther
  }
}

function getStatusClass(status: DendaStatus): string {
  switch (status) {
    case 'pending': return styles.badgePending
    case 'paid': return styles.badgePaid
    case 'waived': return styles.badgeWaived
    case 'cancelled': return styles.badgeCancelled
    default: return styles.badgeDefault
  }
}

export default function DendaPage() {
  const [search, setSearch] = useState('')
  const [penaltyType, setPenaltyType] = useState<DendaPenaltyType | ''>('')
  const [status, setStatus] = useState<DendaStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(penaltyType ? { penalty_type: penaltyType } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.denda, filters],
    queryFn: () => dendaService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => dendaService.delete(id),
    onSuccess: () => {
      toast.success('Denda berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.denda] })
    },
    onError: () => {
      toast.error('Gagal menghapus denda')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || penaltyType !== '' || status !== ''

  const totalAmount = items.reduce((sum, item) => sum + item.amount, 0)
  const totalPending = items.filter((item) => item.status === 'pending').reduce((sum, item) => sum + item.amount, 0)

  function handleReset() {
    setSearch('')
    setPenaltyType('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Denda) {
    if (!confirm(`Hapus denda untuk "${item.nasabahNama}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Denda</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Denda
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Denda</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueRed}`}>{formatRupiah(totalAmount)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Pending</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{formatRupiah(totalPending)}</div>
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
            placeholder="Cari nasabah atau pinjaman..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={penaltyType}
          onChange={(e) => { setPenaltyType(e.target.value as DendaPenaltyType | ''); setPage(1) }}
          aria-label="Filter tipe denda"
        >
          {PENALTY_TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as DendaStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Pinjaman</th>
                <th className={styles.th}>Tipe Denda</th>
                <th className={styles.th}>Jumlah</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Periode</th>
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
                      <p className={styles.emptySubtitle}>Belum ada denda yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.nasabahNama || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.pinjamanId || '-'}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getPenaltyTypeClass(item.penaltyType)}`}>
                        {DENDA_PENALTY_TYPE_LABELS[item.penaltyType] ?? item.penaltyType}
                      </span>
                    </td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.amount)}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {DENDA_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>{item.period || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus denda ${item.nasabahNama}`}
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
