import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2, FileText, Eye } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { createVernonService } from '@/services/vernon.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import styles from './LaporanRegulasiPage.module.css'

const PAGE_SIZE = 20

// ── Types ──────────────────────────────────────────────────────────────────────

type ReportCategory = 'dinas_koperasi' | 'ojk' | 'islamic' | 'internal'
type LaporanStatus = 'draft' | 'reviewed' | 'final' | 'submitted' | 'overdue'
type PeriodType = 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'annually'
type ConsolidationLevel = 'branch' | 'company' | 'tenant'

interface Laporan {
  id: string
  laporanConfigId: string
  reportType: string
  reportCategory: ReportCategory
  periodType: PeriodType
  periodStart: string
  periodEnd: string
  periodLabel: string
  status: LaporanStatus
  consolidationLevel: ConsolidationLevel
  generatedAt: string
  generatedBy: string
  branchName: string
  configName: string
  anomalyFlags: unknown[]
}

// ── Constants ──────────────────────────────────────────────────────────────────

const CATEGORY_LABELS: Record<ReportCategory, string> = {
  dinas_koperasi: 'Dinas Koperasi',
  ojk: 'OJK',
  islamic: 'Syariah',
  internal: 'Internal',
}

const STATUS_LABELS: Record<LaporanStatus, string> = {
  draft: 'Draft',
  reviewed: 'Ditinjau',
  final: 'Final',
  submitted: 'Diserahkan',
  overdue: 'Terlambat',
}

const CATEGORY_OPTIONS: { value: ReportCategory | ''; label: string }[] = [
  { value: '', label: 'Semua Kategori' },
  { value: 'dinas_koperasi', label: 'Dinas Koperasi' },
  { value: 'ojk', label: 'OJK' },
  { value: 'islamic', label: 'Syariah' },
  { value: 'internal', label: 'Internal' },
]

const STATUS_OPTIONS: { value: LaporanStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'draft', label: 'Draft' },
  { value: 'reviewed', label: 'Ditinjau' },
  { value: 'final', label: 'Final' },
  { value: 'submitted', label: 'Diserahkan' },
  { value: 'overdue', label: 'Terlambat' },
]

// ── Service ────────────────────────────────────────────────────────────────────

function transformLaporan(raw: Record<string, unknown>): Laporan {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const branchData = (data.branch ?? {}) as Record<string, unknown>
  const configData = (data.laporan_config ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    laporanConfigId: (raw.laporan_config_id ?? '') as string,
    reportType: (raw.report_type ?? '') as string,
    reportCategory: (raw.report_category ?? 'internal') as ReportCategory,
    periodType: (raw.period_type ?? 'monthly') as PeriodType,
    periodStart: (raw.period_start ?? '') as string,
    periodEnd: (raw.period_end ?? '') as string,
    periodLabel: (raw.period_label ?? '') as string,
    status: (raw.status ?? 'draft') as LaporanStatus,
    consolidationLevel: (raw.consolidation_level ?? 'branch') as ConsolidationLevel,
    generatedAt: (raw.generated_at ?? '') as string,
    generatedBy: (raw.generated_by ?? '') as string,
    branchName: (branchData.name ?? '-') as string,
    configName: (configData.report_name ?? '-') as string,
    anomalyFlags: (raw.anomaly_flags ?? []) as unknown[],
  }
}

const laporanService = createVernonService<Laporan, Record<string, unknown>>(
  '/laporan',
  transformLaporan,
)

// ── Helpers ────────────────────────────────────────────────────────────────────

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: '2-digit', month: 'short', year: 'numeric',
  })
}

function getStatusClass(status: LaporanStatus): string {
  switch (status) {
    case 'draft': return styles.badgeDraft
    case 'reviewed': return styles.badgeReviewed
    case 'final': return styles.badgeFinal
    case 'submitted': return styles.badgeSubmitted
    case 'overdue': return styles.badgeOverdue
    default: return styles.badgeDefault
  }
}

function getCategoryClass(cat: ReportCategory): string {
  switch (cat) {
    case 'dinas_koperasi': return styles.badgeDinas
    case 'ojk': return styles.badgeOjk
    case 'islamic': return styles.badgeIslamic
    case 'internal': return styles.badgeInternal
    default: return styles.badgeDefault
  }
}

// ── Component ──────────────────────────────────────────────────────────────────

export default function LaporanRegulasiPage() {
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState<ReportCategory | ''>('')
  const [status, setStatus] = useState<LaporanStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { period_label: debouncedSearch } : {}),
    ...(category ? { report_category: category } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.laporanRegulasi, filters],
    queryFn: () => laporanService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => laporanService.delete(id),
    onSuccess: () => {
      toast.success('Laporan berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.laporanRegulasi] })
    },
    onError: () => {
      toast.error('Gagal menghapus laporan')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || category !== '' || status !== ''

  const totalDraft = items.filter((i) => i.status === 'draft').length
  const totalSubmitted = items.filter((i) => i.status === 'submitted').length
  const totalOverdue = items.filter((i) => i.status === 'overdue').length

  function handleReset() {
    setSearch('')
    setCategory('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Laporan) {
    if (!confirm(`Hapus laporan "${item.periodLabel}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Laporan Regulasi</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} laporan ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Buat Laporan
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Draft</div>
          <div className={styles.summaryValue}>{totalDraft}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Diserahkan</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{totalSubmitted}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Terlambat</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueRed}`}>{totalOverdue}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari periode laporan..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={category}
          onChange={(e) => { setCategory(e.target.value as ReportCategory | ''); setPage(1) }}
          aria-label="Filter kategori"
        >
          {CATEGORY_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as LaporanStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Periode</th>
                <th className={styles.th}>Jenis Laporan</th>
                <th className={styles.th}>Kategori</th>
                <th className={styles.th}>Cabang</th>
                <th className={styles.th}>Tanggal Generate</th>
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
                      <p className={styles.emptySubtitle}>Belum ada laporan regulasi yang di-generate.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}>
                      <span className={styles.name}>{item.periodLabel}</span>
                    </td>
                    <td className={styles.td}>{item.configName}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getCategoryClass(item.reportCategory)}`}>
                        {CATEGORY_LABELS[item.reportCategory] ?? item.reportCategory}
                      </span>
                    </td>
                    <td className={styles.td}>{item.branchName}</td>
                    <td className={styles.td}>{formatDate(item.generatedAt)}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <div className={styles.actionBtns}>
                        <button type="button" className={styles.viewBtn} aria-label="Lihat detail">
                          <Eye size={14} />
                        </button>
                        <button type="button" className={styles.docBtn} aria-label="Unduh dokumen">
                          <FileText size={14} />
                        </button>
                        <button
                          type="button"
                          className={styles.deleteBtn}
                          onClick={() => handleDelete(item)}
                          disabled={deleteMutation.isPending}
                          aria-label={`Hapus laporan ${item.periodLabel}`}
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>
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
