import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { teacherEvaluationService } from '@/services/sprint7.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { TeacherEvaluation, Semester, WorkflowStatus, EvaluationGrade } from '@/types/sprint7.types'
import {
  SEMESTER_LABELS,
  WORKFLOW_STATUS_LABELS,
  EVAL_GRADE_LABELS,
} from '@/types/sprint7.types'
import styles from './TeacherEvaluationPage.module.css'

const PAGE_SIZE = 20

const SEMESTER_OPTIONS: { value: Semester | ''; label: string }[] = [
  { value: '', label: 'Semua Semester' },
  ...Object.entries(SEMESTER_LABELS).map(([value, label]) => ({
    value: value as Semester,
    label,
  })),
]

const WORKFLOW_OPTIONS: { value: WorkflowStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(WORKFLOW_STATUS_LABELS).map(([value, label]) => ({
    value: value as WorkflowStatus,
    label,
  })),
]

const GRADE_OPTIONS: { value: EvaluationGrade | ''; label: string }[] = [
  { value: '', label: 'Semua Grade' },
  ...Object.entries(EVAL_GRADE_LABELS).map(([value, label]) => ({
    value: value as EvaluationGrade,
    label,
  })),
]

function getWorkflowBadgeClass(status: WorkflowStatus): string {
  switch (status) {
    case 'draft': return styles.badgeGray
    case 'self_assessment': return styles.badgeYellow
    case 'peer_review': return styles.badgeBlue
    case 'supervisor_review': return styles.badgeBlue
    case 'final': return styles.badgeGreen
    case 'approved': return styles.badgeGreen
    default: return styles.badgeDefault
  }
}

function getGradeBadgeClass(grade: EvaluationGrade): string {
  switch (grade) {
    case 'A': return styles.badgeGreen
    case 'B': return styles.badgeBlue
    case 'C': return styles.badgeYellow
    case 'D': return styles.badgeRed
    case 'E': return styles.badgeRed
    default: return styles.badgeDefault
  }
}

function getSemesterBadgeClass(semester: Semester): string {
  switch (semester) {
    case 'ganjil': return styles.badgeBlue
    case 'genap': return styles.badgeGreen
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

export default function TeacherEvaluationPage() {
  const [search, setSearch] = useState('')
  const [semester, setSemester] = useState<Semester | ''>('')
  const [workflowStatus, setWorkflowStatus] = useState<WorkflowStatus | ''>('')
  const [grade, setGrade] = useState<EvaluationGrade | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { teacher_name: debouncedSearch } : {}),
    ...(semester ? { semester } : {}),
    ...(workflowStatus ? { workflow_status: workflowStatus } : {}),
    ...(grade ? { grade } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.teacherEvaluations, filters],
    queryFn: () => teacherEvaluationService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => teacherEvaluationService.delete(id),
    onSuccess: () => {
      toast.success('Data evaluasi berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.teacherEvaluations] })
    },
    onError: () => {
      toast.error('Gagal menghapus data evaluasi')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || semester !== '' || workflowStatus !== '' || grade !== ''

  function handleReset() {
    setSearch('')
    setSemester('')
    setWorkflowStatus('')
    setGrade('')
    setPage(1)
  }

  function handleDelete(item: TeacherEvaluation) {
    if (!confirm(`Hapus evaluasi ${item.teacherName} (${item.academicYearName})?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Evaluasi Kinerja Guru (PKG)</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Buat Evaluasi
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nama guru..."
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
          value={workflowStatus}
          onChange={(e) => { setWorkflowStatus(e.target.value as WorkflowStatus | ''); setPage(1) }}
          aria-label="Filter status"
        >
          {WORKFLOW_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={grade}
          onChange={(e) => { setGrade(e.target.value as EvaluationGrade | ''); setPage(1) }}
          aria-label="Filter grade"
        >
          {GRADE_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Guru</th>
                <th className={styles.th}>NIP</th>
                <th className={styles.th}>Tahun Ajaran</th>
                <th className={styles.th}>Semester</th>
                <th className={styles.th}>Skor</th>
                <th className={styles.th}>Grade</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Tanggal Evaluasi</th>
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
                    <div>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada data evaluasi guru.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.teacherName || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.teacherNip || '-'}</span></td>
                    <td className={styles.td}>{item.academicYearName || '-'}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getSemesterBadgeClass(item.semester)}`}>
                        {SEMESTER_LABELS[item.semester] ?? item.semester}
                      </span>
                    </td>
                    <td className={styles.td}>{item.totalScore != null ? item.totalScore : '-'}</td>
                    <td className={styles.td}>
                      {item.grade ? (
                        <span className={`${styles.badge} ${getGradeBadgeClass(item.grade)}`}>
                          {item.grade}
                        </span>
                      ) : '-'}
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getWorkflowBadgeClass(item.workflowStatus)}`}>
                        {WORKFLOW_STATUS_LABELS[item.workflowStatus] ?? item.workflowStatus}
                      </span>
                    </td>
                    <td className={styles.td}>{formatDate(item.evaluationDate)}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus evaluasi ${item.teacherName}`}
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
