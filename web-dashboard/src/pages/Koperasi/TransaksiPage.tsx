import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { transaksiService } from '@/services/koperasi-ops.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Transaksi, TransaksiTransactionType, TransaksiStatus } from '@/types/koperasi-ops.types'
import { TRANSAKSI_TRANSACTION_TYPE_LABELS, TRANSAKSI_STATUS_LABELS } from '@/types/koperasi-ops.types'
import styles from './TransaksiPage.module.css'

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

const TRANSACTION_TYPE_OPTIONS: { value: TransaksiTransactionType | ''; label: string }[] = [
  { value: '', label: 'Semua Tipe' },
  ...Object.entries(TRANSAKSI_TRANSACTION_TYPE_LABELS).map(([value, label]) => ({
    value: value as TransaksiTransactionType,
    label,
  })),
]

const STATUS_OPTIONS: { value: TransaksiStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(TRANSAKSI_STATUS_LABELS).map(([value, label]) => ({
    value: value as TransaksiStatus,
    label,
  })),
]

function getTransactionTypeClass(type: TransaksiTransactionType): string {
  switch (type) {
    case 'credit': return styles.badgeCredit
    case 'debit': return styles.badgeDebit
    case 'transfer': return styles.badgeTransfer
    default: return styles.badgeDefault
  }
}

function getStatusClass(status: TransaksiStatus): string {
  switch (status) {
    case 'pending': return styles.badgePending
    case 'posted': return styles.badgePosted
    case 'reversed': return styles.badgeReversed
    default: return styles.badgeDefault
  }
}

export default function TransaksiPage() {
  const [search, setSearch] = useState('')
  const [transactionType, setTransactionType] = useState<TransaksiTransactionType | ''>('')
  const [status, setStatus] = useState<TransaksiStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(transactionType ? { transaction_type: transactionType } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.transaksi, filters],
    queryFn: () => transaksiService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => transaksiService.delete(id),
    onSuccess: () => {
      toast.success('Transaksi berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.transaksi] })
    },
    onError: () => {
      toast.error('Gagal menghapus transaksi')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || transactionType !== '' || status !== ''

  const totalKredit = items.filter((i) => i.transactionType === 'credit').reduce((s, i) => s + i.amount, 0)
  const totalDebit = items.filter((i) => i.transactionType === 'debit').reduce((s, i) => s + i.amount, 0)

  function handleReset() {
    setSearch('')
    setTransactionType('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Transaksi) {
    if (!confirm(`Hapus transaksi "${item.referenceNo || item.id}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Transaksi</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Transaksi
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Kredit</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{formatRupiah(totalKredit)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Debit</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueRed}`}>{formatRupiah(totalDebit)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Transaksi</div>
          <div className={styles.summaryValue}>{total}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari referensi atau deskripsi..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={transactionType}
          onChange={(e) => { setTransactionType(e.target.value as TransaksiTransactionType | ''); setPage(1) }}
          aria-label="Filter tipe transaksi"
        >
          {TRANSACTION_TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as TransaksiStatus | ''); setPage(1) }}
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
                <th className={styles.th}>No. Rekening</th>
                <th className={styles.th}>Tipe</th>
                <th className={styles.th}>Jumlah</th>
                <th className={styles.th}>Saldo Sebelum</th>
                <th className={styles.th}>Saldo Sesudah</th>
                <th className={styles.th}>Referensi</th>
                <th className={styles.th}>Tanggal Posting</th>
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
                      <p className={styles.emptySubtitle}>Belum ada transaksi yang tercatat.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.rekeningNo || '-'}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getTransactionTypeClass(item.transactionType)}`}>
                        {TRANSAKSI_TRANSACTION_TYPE_LABELS[item.transactionType] ?? item.transactionType}
                      </span>
                    </td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.amount)}</span></td>
                    <td className={styles.td}><span className={styles.amount}>{formatRupiah(item.balanceBefore)}</span></td>
                    <td className={styles.td}><span className={styles.amount}>{formatRupiah(item.balanceAfter)}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.referenceNo || '-'}</span></td>
                    <td className={styles.td}>{formatDate(item.postedDate)}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {TRANSAKSI_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus transaksi ${item.referenceNo}`}
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
