import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { academicRecordService } from '@/services/sprint5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { AcademicRecord, AcademicRecordStatus } from '@/types/sprint5.types'
import { ACADEMIC_RECORD_STATUS_LABELS } from '@/types/sprint5.types'
import styles from './AcademicRecordPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: AcademicRecordStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(ACADEMIC_RECORD_STATUS_LABELS).map(([value, label]) => ({
    value: value as AcademicRecordStatus,
    label,
  })),
]

function getStatusClass(status: AcademicRecordStatus): string {
  switch (status) {
    case 'pass': return styles.badgePass
    case 'fail': return styles.badgeFail
    case 'remedial': return styles.badgeRemedial
    default: return styles.badgeDefault
  }
}

export default function AcademicRecordPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<AcademicRecordStatus | ''>('')
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
    queryKey: [QK.academicRecord, filters],
    queryFn: () => academicRecordService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => academicRecordService.delete(id),
    onSuccess: () => {
      toast.success('Catatan akademik berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.academicRecord] })
    },
    onError: () => {
      toast.error('Gagal menghapus catatan akademik')
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

  function handleDelete(item: AcademicRecord) {
    if (!confirm(`Hapus catatan akademik "${item.studentName}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Catatan Akademik</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Catatan
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
          value={status}
          onChange={(e) => { setStatus(e.target.value as AcademicRecordStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Tahun Ajaran</th>
                <th className={styles.th}>Semester</th>
                <th className={styles.th}>Rata-rata</th>
                <th className={styles.th}>Peringkat</th>
                <th className={styles.th}>Status</th>
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
                      <p className={styles.emptySubtitle}>Belum ada catatan akademik.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.studentName || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.studentNis || '-'}</span></td>
                    <td className={styles.td}>{item.academicYearName || '-'}</td>
                    <td className={styles.td}>Semester {item.semester}</td>
                    <td className={styles.td}>{item.gpa > 0 ? item.gpa.toFixed(2) : '-'}</td>
                    <td className={styles.td}>{item.rank > 0 ? item.rank : '-'}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {ACADEMIC_RECORD_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus catatan ${item.studentName}`}
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
