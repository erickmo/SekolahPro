import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { studentClassPlacementService } from '@/services/student.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { StudentClassPlacement, PlacementStatus } from '@/types/student.types'
import { PLACEMENT_STATUS_LABELS } from '@/types/student.types'
import styles from './StudentClassPlacementsPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: PlacementStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'active', label: 'Aktif' },
  { value: 'moved', label: 'Pindah' },
  { value: 'graduated', label: 'Naik Kelas' },
]

function getStatusClass(status: PlacementStatus): string {
  switch (status) {
    case 'active': return styles.badgeActive
    case 'moved': return styles.badgeMoved
    case 'graduated': return styles.badgeGraduated
    default: return styles.badgeDefault
  }
}

export default function StudentClassPlacementsPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<PlacementStatus | ''>('')
  const [semester, setSemester] = useState<string>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { full_name: debouncedSearch } : {}),
    ...(status ? { status } : {}),
    ...(semester ? { semester } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.studentClassPlacements, filters],
    queryFn: () => studentClassPlacementService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => studentClassPlacementService.delete(id),
    onSuccess: () => {
      toast.success('Penempatan kelas berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.studentClassPlacements] })
    },
    onError: () => {
      toast.error('Gagal menghapus penempatan kelas')
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

  function handleDelete(item: StudentClassPlacement) {
    if (!confirm(`Hapus penempatan "${item.studentName}" di ${item.classRoomName}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Penempatan Kelas</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Penempatan
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
          onChange={(e) => { setSemester(e.target.value); setPage(1) }}
          aria-label="Filter semester"
        >
          <option value="">Semua Semester</option>
          <option value="1">Semester 1</option>
          <option value="2">Semester 2</option>
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as PlacementStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Kelas</th>
                <th className={styles.th}>Tahun Ajaran</th>
                <th className={styles.th}>Semester</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Tgl Masuk</th>
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
                      <p className={styles.emptySubtitle}>Belum ada penempatan kelas.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.studentName || '-'}</span></td>
                    <td className={styles.td}>{item.classRoomName || '-'}</td>
                    <td className={styles.td}>{item.academicYearName || '-'}</td>
                    <td className={styles.td}>{item.semester}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {PLACEMENT_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>{item.enrollmentDate || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus penempatan ${item.studentName}`}
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
