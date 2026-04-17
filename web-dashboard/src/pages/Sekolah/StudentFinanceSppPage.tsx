import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { studentInvoiceService } from '@/services/sprint6.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { StudentInvoice, InvoiceStatus } from '@/types/sprint6.types'
import { INVOICE_STATUS_LABELS } from '@/types/sprint6.types'
import styles from './StudentFinanceSppPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: InvoiceStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(INVOICE_STATUS_LABELS).map(([value, label]) => ({
    value: value as InvoiceStatus,
    label,
  })),
]

function getStatusClass(status: InvoiceStatus): string {
  switch (status) {
    case 'paid': return styles.badgePaid
    case 'unpaid': return styles.badgeUnpaid
    case 'partial': return styles.badgePartial
    case 'overdue': return styles.badgeOverdue
    case 'waived': return styles.badgeWaived
    default: return styles.badgeDefault
  }
}

function formatCurrency(amount: number): string {
  return amount.toLocaleString('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 })
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
  } catch {
    return dateStr
  }
}

function getPeriodLabel(item: StudentInvoice): string {
  if (item.periodMonth != null) {
    const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Ags', 'Sep', 'Okt', 'Nov', 'Des']
    return `${monthNames[item.periodMonth - 1]} ${item.periodYear}`
  }
  return String(item.periodYear)
}

export default function StudentFinanceSppPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<InvoiceStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { student_name: debouncedSearch } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.studentInvoices, filters],
    queryFn: () => studentInvoiceService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => studentInvoiceService.delete(id),
    onSuccess: () => {
      toast.success('Tagihan berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.studentInvoices] })
    },
    onError: () => {
      toast.error('Gagal menghapus tagihan')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || status !== ''

  function handleReset() {
    setSearch('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: StudentInvoice) {
    if (!confirm(`Hapus tagihan "${item.invoiceNo}" milik ${item.studentName}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Keuangan Siswa (SPP)</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} tagihan ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Buat Tagihan
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nama siswa atau no. invoice..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as InvoiceStatus | ''); setPage(1) }}
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
                <th className={styles.th}>No. Invoice</th>
                <th className={styles.th}>Siswa</th>
                <th className={styles.th}>NIS</th>
                <th className={styles.th}>Jenis</th>
                <th className={styles.th}>Periode</th>
                <th className={styles.th}>Total</th>
                <th className={styles.th}>Dibayar</th>
                <th className={styles.th}>Jatuh Tempo</th>
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
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada tagihan siswa.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.invoiceNo}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.studentName || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.studentNis || '-'}</span></td>
                    <td className={styles.td}>{item.feeTypeName || '-'}</td>
                    <td className={styles.td}>{getPeriodLabel(item)}</td>
                    <td className={styles.td}>{formatCurrency(item.totalAmount)}</td>
                    <td className={styles.td}>{formatCurrency(item.paidAmount)}</td>
                    <td className={styles.td}>{formatDate(item.dueDate)}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {INVOICE_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus tagihan ${item.invoiceNo}`}
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
