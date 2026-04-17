import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { lessonPlanService } from '@/services/phase5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { LessonPlan, LessonPlanType, LessonPlanStatus } from '@/types/phase5.types'
import { LESSON_PLAN_TYPE_LABELS, LESSON_PLAN_STATUS_LABELS } from '@/types/phase5.types'
import styles from './LessonPlanPage.module.css'

const PAGE_SIZE = 20

const PLAN_TYPE_OPTIONS: { value: LessonPlanType | ''; label: string }[] = [
  { value: '', label: 'Semua Tipe' },
  { value: 'daily', label: 'Harian' },
  { value: 'weekly', label: 'Mingguan' },
  { value: 'unit', label: 'Unit' },
  { value: 'semester', label: 'Semester' },
]

const STATUS_OPTIONS: { value: LessonPlanStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'draft', label: 'Draft' },
  { value: 'submitted', label: 'Dikirim' },
  { value: 'approved', label: 'Disetujui' },
  { value: 'revision', label: 'Revisi' },
  { value: 'archived', label: 'Diarsipkan' },
]

function getPlanTypeClass(planType: LessonPlanType): string {
  switch (planType) {
    case 'daily': return styles.badgeDaily
    case 'weekly': return styles.badgeWeekly
    case 'unit': return styles.badgeUnit
    case 'semester': return styles.badgeSemester
    default: return styles.badgeDefault
  }
}

function getStatusClass(status: LessonPlanStatus): string {
  switch (status) {
    case 'draft': return styles.badgeDraft
    case 'submitted': return styles.badgeSubmitted
    case 'approved': return styles.badgeApproved
    case 'revision': return styles.badgeRevision
    case 'archived': return styles.badgeArchived
    default: return styles.badgeDefault
  }
}

export default function LessonPlanPage() {
  const [search, setSearch] = useState('')
  const [planType, setPlanType] = useState<LessonPlanType | ''>('')
  const [status, setStatus] = useState<LessonPlanStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(planType ? { plan_type: planType } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.lessonPlan, filters],
    queryFn: () => lessonPlanService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => lessonPlanService.delete(id),
    onSuccess: () => {
      toast.success('RPP berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.lessonPlan] })
    },
    onError: () => {
      toast.error('Gagal menghapus RPP')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || planType !== '' || status !== ''

  const totalApproved = items.filter((item) => item.status === 'approved').length
  const totalDraft = items.filter((item) => item.status === 'draft').length

  function handleReset() {
    setSearch('')
    setPlanType('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: LessonPlan) {
    if (!confirm(`Hapus RPP "${item.title}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Rencana Pelaksanaan Pembelajaran</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah RPP
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Disetujui</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{totalApproved}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Draft</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{totalDraft}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total RPP</div>
          <div className={styles.summaryValue}>{total}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari judul, mapel, atau guru..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={planType}
          onChange={(e) => { setPlanType(e.target.value as LessonPlanType | ''); setPage(1) }}
          aria-label="Filter tipe rencana"
        >
          {PLAN_TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as LessonPlanStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Judul</th>
                <th className={styles.th}>Tipe</th>
                <th className={styles.th}>Mata Pelajaran</th>
                <th className={styles.th}>Guru</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Semester</th>
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
                      <p className={styles.emptySubtitle}>Belum ada RPP yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.title}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getPlanTypeClass(item.planType)}`}>
                        {LESSON_PLAN_TYPE_LABELS[item.planType] ?? item.planType}
                      </span>
                    </td>
                    <td className={styles.td}>{item.subjectName || '-'}</td>
                    <td className={styles.td}>{item.teacherName || '-'}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {LESSON_PLAN_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>{item.semester || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus RPP ${item.title}`}
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
