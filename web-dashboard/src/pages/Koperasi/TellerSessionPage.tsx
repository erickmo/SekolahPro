import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { tellerSessionService } from '@/services/koperasi-ops.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { TellerSession, TellerSessionStatus } from '@/types/koperasi-ops.types'
import { TELLER_SESSION_STATUS_LABELS } from '@/types/koperasi-ops.types'
import styles from './TellerSessionPage.module.css'

const PAGE_SIZE = 20

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    })
  } catch {
    return dateStr
  }
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleTimeString('id-ID', {
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return dateStr
  }
}

const STATUS_OPTIONS: { value: TellerSessionStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(TELLER_SESSION_STATUS_LABELS).map(([value, label]) => ({
    value: value as TellerSessionStatus,
    label,
  })),
]

function getStatusClass(status: TellerSessionStatus): string {
  switch (status) {
    case 'open': return styles.badgeOpen
    case 'closed': return styles.badgeClosed
    case 'reconciled': return styles.badgeReconciled
    default: return styles.badgeDefault
  }
}

export default function TellerSessionPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<TellerSessionStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.tellerSession, filters],
    queryFn: () => tellerSessionService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => tellerSessionService.delete(id),
    onSuccess: () => {
      toast.success('Sesi teller berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.tellerSession] })
    },
    onError: () => {
      toast.error('Gagal menghapus sesi teller')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || status !== ''

  const totalOpeningBalance = items.reduce((sum, item) => sum + item.openingBalance, 0)
  const totalCashTotal = items.reduce((sum, item) => sum + item.cashTotal, 0)
  const openSessions = items.filter((item) => item.status === 'open').length

  function handleReset() {
    setSearch('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: TellerSession) {
    if (!confirm(`Hapus sesi teller tanggal "${item.sessionDate}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Sesi Teller</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Buka Sesi Baru
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Saldo Awal</div>
          <div className={styles.summaryValue}>{formatRupiah(totalOpeningBalance)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Kas</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{formatRupiah(totalCashTotal)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Sesi Aktif</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{openSessions}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari sesi teller..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as TellerSessionStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Tanggal Sesi</th>
                <th className={styles.th}>Jam Mulai</th>
                <th className={styles.th}>Jam Selesai</th>
                <th className={styles.th}>Saldo Awal</th>
                <th className={styles.th}>Saldo Akhir</th>
                <th className={styles.th}>Total Kas</th>
                <th className={styles.th}>Jml. Transaksi</th>
                <th className={styles.th}>Status</th>
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
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada sesi teller yang tercatat.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}>{formatDate(item.sessionDate)}</td>
                    <td className={styles.td}>{formatTime(item.startTime)}</td>
                    <td className={styles.td}>{item.endTime ? formatTime(item.endTime) : <span className={styles.dash}>&mdash;</span>}</td>
                    <td className={styles.td}><span className={styles.amount}>{formatRupiah(item.openingBalance)}</span></td>
                    <td className={styles.td}><span className={styles.amount}>{formatRupiah(item.closingBalance)}</span></td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.cashTotal)}</span></td>
                    <td className={styles.td}>{item.transactionCount}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {TELLER_SESSION_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus sesi ${item.sessionDate}`}
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
