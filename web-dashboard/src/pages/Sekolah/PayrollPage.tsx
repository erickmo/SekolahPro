import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { payrollPeriodService } from '@/services/sprint7.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { PayrollPeriod, PayrollWorkflowStatus } from '@/types/sprint7.types'
import { PAYROLL_WORKFLOW_LABELS, MONTH_LABELS } from '@/types/sprint7.types'
import styles from './PayrollPage.module.css'

const PAGE_SIZE = 20

const CURRENT_YEAR = new Date().getFullYear()
const YEAR_OPTIONS = Array.from({ length: 5 }, (_, i) => CURRENT_YEAR - 2 + i)

const WORKFLOW_OPTIONS: { value: PayrollWorkflowStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(PAYROLL_WORKFLOW_LABELS).map(([value, label]) => ({
    value: value as PayrollWorkflowStatus,
    label,
  })),
]

function getWorkflowBadgeClass(status: PayrollWorkflowStatus): string {
  switch (status) {
    case 'draft': return styles.badgeGray
    case 'calculating': return styles.badgeYellow
    case 'calculated': return styles.badgeBlue
    case 'approved': return styles.badgeGreen
    case 'processing': return styles.badgeBlue
    case 'paid': return styles.badgeGreen
    case 'cancelled': return styles.badgeRed
    case 'failed': return styles.badgeRed
    default: return styles.badgeDefault
  }
}

function formatCurrency(amount: number): string {
  return amount.toLocaleString('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 })
}

export default function PayrollPage() {
  const [search, setSearch] = useState('')
  const [workflowStatus, setWorkflowStatus] = useState<PayrollWorkflowStatus | ''>('')
  const [year, setYear] = useState<number | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(workflowStatus ? { workflow_status: workflowStatus } : {}),
    ...(year ? { year } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.payrollPeriods, filters],
    queryFn: () => payrollPeriodService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => payrollPeriodService.delete(id),
    onSuccess: () => {
      toast.success('Data payroll berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.payrollPeriods] })
    },
    onError: () => {
      toast.error('Gagal menghapus data payroll')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || workflowStatus !== '' || year !== ''

  function handleReset() {
    setSearch('')
    setWorkflowStatus('')
    setYear('')
    setPage(1)
  }

  function handleDelete(item: PayrollPeriod) {
    if (!confirm(`Hapus periode payroll "${item.periodName}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Penggajian</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Buat Periode
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari periode payroll..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={workflowStatus}
          onChange={(e) => { setWorkflowStatus(e.target.value as PayrollWorkflowStatus | ''); setPage(1) }}
          aria-label="Filter status"
        >
          {WORKFLOW_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={year}
          onChange={(e) => { setYear(e.target.value ? Number(e.target.value) : ''); setPage(1) }}
          aria-label="Filter tahun"
        >
          <option value="">Semua Tahun</option>
          {YEAR_OPTIONS.map((y) => (
            <option key={y} value={y}>{y}</option>
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
                <th className={styles.th}>Nama Periode</th>
                <th className={styles.th}>Tahun Ajaran</th>
                <th className={styles.th}>Bulan</th>
                <th className={styles.th}>Tahun</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Jumlah Karyawan</th>
                <th className={styles.th}>Total Netto</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 8 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={8} className={styles.emptyCell}>
                    <div>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada data periode payroll.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.periodName || '-'}</span></td>
                    <td className={styles.td}>{item.academicYearName || '-'}</td>
                    <td className={styles.td}>{MONTH_LABELS[item.month] ?? item.month}</td>
                    <td className={styles.td}>{item.year}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getWorkflowBadgeClass(item.workflowStatus)}`}>
                        {PAYROLL_WORKFLOW_LABELS[item.workflowStatus] ?? item.workflowStatus}
                      </span>
                    </td>
                    <td className={styles.td}>{item.totalEmployees}</td>
                    <td className={styles.td}><span className={styles.amount}>{formatCurrency(item.totalNet)}</span></td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus periode ${item.periodName}`}
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
