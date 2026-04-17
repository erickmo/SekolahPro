import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { subjectGradeService } from '@/services/sprint5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { SubjectGrade, GradeType } from '@/types/sprint5.types'
import { GRADE_TYPE_LABELS } from '@/types/sprint5.types'
import styles from './SubjectGradePage.module.css'

const PAGE_SIZE = 20

const GRADE_TYPE_OPTIONS: { value: GradeType | ''; label: string }[] = [
  { value: '', label: 'Semua Jenis Nilai' },
  ...Object.entries(GRADE_TYPE_LABELS).map(([value, label]) => ({
    value: value as GradeType,
    label,
  })),
]

export default function SubjectGradePage() {
  const [search, setSearch] = useState('')
  const [gradeType, setGradeType] = useState<GradeType | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { student_name: debouncedSearch } : {}),
    ...(gradeType ? { grade_type: gradeType } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.subjectGrade, filters],
    queryFn: () => subjectGradeService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => subjectGradeService.delete(id),
    onSuccess: () => {
      toast.success('Nilai berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.subjectGrade] })
    },
    onError: () => {
      toast.error('Gagal menghapus nilai')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || gradeType !== ''

  function handleReset() {
    setSearch('')
    setGradeType('')
    setPage(1)
  }

  function handleDelete(item: SubjectGrade) {
    if (!confirm(`Hapus nilai "${item.studentName}" - ${item.subjectName}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Nilai Mata Pelajaran</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Nilai
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
          value={gradeType}
          onChange={(e) => { setGradeType(e.target.value as GradeType | ''); setPage(1) }}
          aria-label="Filter jenis nilai"
        >
          {GRADE_TYPE_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Mata Pelajaran</th>
                <th className={styles.th}>Guru</th>
                <th className={styles.th}>Jenis Nilai</th>
                <th className={styles.th}>Nilai</th>
                <th className={styles.th}>Bobot</th>
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
                      <p className={styles.emptySubtitle}>Belum ada data nilai mata pelajaran.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.studentName || '-'}</span></td>
                    <td className={styles.td}>{item.subjectName || '-'}</td>
                    <td className={styles.td}>{item.teacherName || '-'}</td>
                    <td className={styles.td}>
                      <span className={styles.badgeType}>
                        {GRADE_TYPE_LABELS[item.gradeType] ?? item.gradeType}
                      </span>
                    </td>
                    <td className={styles.td}><span className={styles.score}>{item.grade > 0 ? item.grade : '-'}</span></td>
                    <td className={styles.td}>{item.weight > 0 ? `${item.weight}%` : '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus nilai ${item.studentName}`}
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
