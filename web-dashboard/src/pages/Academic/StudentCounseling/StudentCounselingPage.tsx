import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { counselingCaseService } from '@/services/sprint6.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { CounselingCase, CaseCategory, CaseSeverity, CaseStatus } from '@/types/sprint6.types'
import { CASE_CATEGORY_LABELS, CASE_SEVERITY_LABELS, CASE_STATUS_LABELS } from '@/types/sprint6.types'
import styles from './StudentCounselingPage.module.css'

const PAGE_SIZE = 20

const CATEGORY_OPTIONS: { value: CaseCategory | ''; label: string }[] = [
  { value: '', label: 'Semua Kategori' },
  ...Object.entries(CASE_CATEGORY_LABELS).map(([value, label]) => ({
    value: value as CaseCategory,
    label,
  })),
]

const STATUS_OPTIONS: { value: CaseStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(CASE_STATUS_LABELS).map(([value, label]) => ({
    value: value as CaseStatus,
    label,
  })),
]

function getSeverityClass(severity: CaseSeverity): string {
  switch (severity) {
    case 'low': return styles.badgeLow
    case 'medium': return styles.badgeMedium
    case 'high': return styles.badgeHigh
    case 'critical': return styles.badgeCritical
    default: return styles.badgeDefault
  }
}

function getStatusClass(status: CaseStatus): string {
  switch (status) {
    case 'open': return styles.statusOpen
    case 'in_progress': return styles.statusInProgress
    case 'referred': return styles.statusReferred
    case 'resolved': return styles.statusResolved
    case 'closed': return styles.statusClosed
    default: return styles.badgeDefault
  }
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

export default function StudentCounselingPage() {
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState<CaseCategory | ''>('')
  const [status, setStatus] = useState<CaseStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { student_name: debouncedSearch } : {}),
    ...(category ? { category } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.counselingCase, filters],
    queryFn: () => counselingCaseService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => counselingCaseService.delete(id),
    onSuccess: () => {
      toast.success('Kasus BK berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.counselingCase] })
    },
    onError: () => {
      toast.error('Gagal menghapus kasus BK')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || category !== '' || status !== ''

  function handleReset() {
    setSearch('')
    setCategory('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: CounselingCase) {
    if (!confirm(`Hapus kasus BK "${item.caseNo}" milik ${item.studentName}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Bimbingan Konseling</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} kasus ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Kasus
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nama siswa..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={category}
          onChange={(e) => { setCategory(e.target.value as CaseCategory | ''); setPage(1) }}
          aria-label="Filter kategori"
        >
          {CATEGORY_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as CaseStatus | ''); setPage(1) }}
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
                <th className={styles.th}>No. Kasus</th>
                <th className={styles.th}>Siswa</th>
                <th className={styles.th}>Judul</th>
                <th className={styles.th}>Kategori</th>
                <th className={styles.th}>Tingkat</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Tanggal Buka</th>
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
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada kasus bimbingan konseling.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.caseNo}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.studentName || '-'}</span></td>
                    <td className={styles.td}>{item.title}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${styles.badgeDefault}`}>
                        {CASE_CATEGORY_LABELS[item.category] ?? item.category}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getSeverityClass(item.severity)}`}>
                        {CASE_SEVERITY_LABELS[item.severity] ?? item.severity}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {CASE_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>{formatDate(item.openedDate)}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus kasus ${item.caseNo}`}
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
