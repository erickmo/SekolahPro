import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2, FileText } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { raporRecordService } from '@/services/sprint6.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { RaporRecord, RecordStatus, Semester } from '@/types/sprint6.types'
import { RECORD_STATUS_LABELS, SEMESTER_LABELS, CURRICULUM_TYPE_LABELS } from '@/types/sprint6.types'
import styles from './RaporGenerationPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: RecordStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(RECORD_STATUS_LABELS).map(([value, label]) => ({
    value: value as RecordStatus,
    label,
  })),
]

const SEMESTER_OPTIONS: { value: Semester | ''; label: string }[] = [
  { value: '', label: 'Semua Semester' },
  ...Object.entries(SEMESTER_LABELS).map(([value, label]) => ({
    value: value as Semester,
    label,
  })),
]

function getStatusClass(status: RecordStatus): string {
  switch (status) {
    case 'draft': return styles.badgeDraft
    case 'generated': return styles.badgeGenerated
    case 'reviewed': return styles.badgeReviewed
    case 'finalized': return styles.badgeFinalized
    default: return styles.badgeDefault
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
  } catch {
    return dateStr
  }
}

export default function RaporGenerationPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<RecordStatus | ''>('')
  const [semester, setSemester] = useState<Semester | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { student_name: debouncedSearch } : {}),
    ...(status ? { status } : {}),
    ...(semester ? { semester } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.raporRecords, filters],
    queryFn: () => raporRecordService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => raporRecordService.delete(id),
    onSuccess: () => {
      toast.success('Rapor berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.raporRecords] })
    },
    onError: () => {
      toast.error('Gagal menghapus rapor')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || status !== '' || semester !== ''

  function handleReset() {
    setSearch('')
    setStatus('')
    setSemester('')
    setPage(1)
  }

  function handleDelete(item: RaporRecord) {
    if (item.status === 'finalized') {
      toast.error('Rapor yang sudah difinalisasi tidak dapat dihapus')
      return
    }
    if (!confirm(`Hapus rapor ${item.studentName} semester ${SEMESTER_LABELS[item.semester]}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Generasi Rapor</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} rapor ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <FileText size={16} /> Generate Rapor
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
          value={semester}
          onChange={(e) => { setSemester(e.target.value as Semester | ''); setPage(1) }}
          aria-label="Filter semester"
        >
          {SEMESTER_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as RecordStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Siswa</th>
                <th className={styles.th}>NIS</th>
                <th className={styles.th}>Kelas</th>
                <th className={styles.th}>Tahun Ajaran</th>
                <th className={styles.th}>Semester</th>
                <th className={styles.th}>Template</th>
                <th className={styles.th}>Kurikulum</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Dihasilkan</th>
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
                      <p className={styles.emptySubtitle}>Belum ada rapor yang dihasilkan.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.studentName || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.studentNis || '-'}</span></td>
                    <td className={styles.td}>{item.classRoomName || '-'}</td>
                    <td className={styles.td}>{item.academicYearName || '-'}</td>
                    <td className={styles.td}>{SEMESTER_LABELS[item.semester] ?? item.semester}</td>
                    <td className={styles.td}>{item.templateName || '-'}</td>
                    <td className={styles.td}>{CURRICULUM_TYPE_LABELS[item.curriculumType] ?? item.curriculumType}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {RECORD_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>{item.generatedAt ? formatDate(item.generatedAt) : '-'}</td>
                    <td className={styles.td}>
                      <div className={styles.actionGroup}>
                        {item.pdfUrl && (
                          <a
                            href={item.pdfUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            className={styles.pdfBtn}
                            aria-label={`Unduh PDF rapor ${item.studentName}`}
                          >
                            <FileText size={14} />
                          </a>
                        )}
                        <button
                          type="button"
                          className={styles.deleteBtn}
                          onClick={() => handleDelete(item)}
                          disabled={deleteMutation.isPending || item.status === 'finalized'}
                          aria-label={`Hapus rapor ${item.studentName}`}
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
