import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { examAssessmentService } from '@/services/sprint5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { ExamAssessment, ExamType, ExamStatus } from '@/types/sprint5.types'
import { EXAM_TYPE_LABELS, EXAM_STATUS_LABELS } from '@/types/sprint5.types'
import styles from './ExamAssessmentPage.module.css'

const PAGE_SIZE = 20

const EXAM_TYPE_OPTIONS: { value: ExamType | ''; label: string }[] = [
  { value: '', label: 'Semua Jenis Ujian' },
  ...Object.entries(EXAM_TYPE_LABELS).map(([value, label]) => ({
    value: value as ExamType,
    label,
  })),
]

const EXAM_STATUS_OPTIONS: { value: ExamStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(EXAM_STATUS_LABELS).map(([value, label]) => ({
    value: value as ExamStatus,
    label,
  })),
]

function getStatusClass(status: ExamStatus): string {
  switch (status) {
    case 'draft': return styles.badgeDraft
    case 'published': return styles.badgePublished
    case 'completed': return styles.badgeCompleted
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

export default function ExamAssessmentPage() {
  const [search, setSearch] = useState('')
  const [examType, setExamType] = useState<ExamType | ''>('')
  const [status, setStatus] = useState<ExamStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { title: debouncedSearch } : {}),
    ...(examType ? { exam_type: examType } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.examAssessment, filters],
    queryFn: () => examAssessmentService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => examAssessmentService.delete(id),
    onSuccess: () => {
      toast.success('Ujian/penilaian berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.examAssessment] })
    },
    onError: () => {
      toast.error('Gagal menghapus ujian/penilaian')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || examType !== '' || status !== ''

  function handleReset() {
    setSearch('')
    setExamType('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: ExamAssessment) {
    if (!confirm(`Hapus ujian "${item.title}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Ujian & Penilaian</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Ujian
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari judul ujian..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={examType}
          onChange={(e) => { setExamType(e.target.value as ExamType | ''); setPage(1) }}
          aria-label="Filter jenis ujian"
        >
          {EXAM_TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as ExamStatus | ''); setPage(1) }}
          aria-label="Filter status"
        >
          {EXAM_STATUS_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Judul</th>
                <th className={styles.th}>Jenis</th>
                <th className={styles.th}>Mata Pelajaran</th>
                <th className={styles.th}>Guru</th>
                <th className={styles.th}>Kelas</th>
                <th className={styles.th}>Tanggal</th>
                <th className={styles.th}>Nilai Maks</th>
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
                      <p className={styles.emptySubtitle}>Belum ada data ujian/penilaian.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.title}</span></td>
                    <td className={styles.td}>
                      <span className={styles.badgeType}>
                        {EXAM_TYPE_LABELS[item.examType] ?? item.examType}
                      </span>
                    </td>
                    <td className={styles.td}>{item.subjectName || '-'}</td>
                    <td className={styles.td}>{item.teacherName || '-'}</td>
                    <td className={styles.td}>{item.classRoomName || '-'}</td>
                    <td className={styles.td}>{formatDate(item.date)}</td>
                    <td className={styles.td}>{item.maxScore}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {EXAM_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus ujian ${item.title}`}
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
