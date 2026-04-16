import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { classRoomService } from '@/services/class-room.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { ClassRoom } from '@/types/class-room.types'
import { MAJOR_LABELS } from '@/types/class-room.types'
import styles from '../AcademicYears/AcademicYearsPage.module.css'

const PAGE_SIZE = 20

const GRADE_OPTIONS = [
  { value: '', label: 'Semua Tingkat' },
  ...Array.from({ length: 12 }, (_, i) => ({
    value: String(i + 1),
    label: `Kelas ${i + 1}`,
  })),
]

export default function ClassRoomsPage() {
  const [search, setSearch] = useState('')
  const [gradeLevel, setGradeLevel] = useState('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { name: debouncedSearch } : {}),
    ...(gradeLevel ? { grade_level: gradeLevel } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.classRooms, filters],
    queryFn: () => classRoomService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => classRoomService.delete(id),
    onSuccess: () => {
      toast.success('Kelas berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.classRooms] })
    },
    onError: () => toast.error('Gagal menghapus kelas'),
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || gradeLevel !== ''

  function handleReset() { setSearch(''); setGradeLevel(''); setPage(1) }
  function handleDelete(item: ClassRoom) {
    if (!confirm(`Hapus kelas "${item.name}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Kelas & Rombel</h1>
          {!isLoading && <p className={styles.totalInfo}>{total.toLocaleString('id-ID')} data ditemukan</p>}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Kelas
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input type="text" className={styles.searchInput} placeholder="Cari nama kelas..." value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }} />
        </div>
        <select className={styles.select} value={gradeLevel} onChange={(e) => { setGradeLevel(e.target.value); setPage(1) }}>
          {GRADE_OPTIONS.map((opt) => <option key={opt.value} value={opt.value}>{opt.label}</option>)}
        </select>
        {hasFilters && (
          <button type="button" className={styles.resetBtn} onClick={handleReset}><RotateCcw size={14} /> Reset Filter</button>
        )}
      </div>

      <div className={styles.tableCard}>
        <div className={styles.tableScroll}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.th}>Nama Kelas</th>
                <th className={styles.th}>Tingkat</th>
                <th className={styles.th}>Jurusan</th>
                <th className={styles.th}>Kapasitas</th>
                <th className={styles.th}>Tahun Ajaran</th>
                <th className={styles.th}>Wali Kelas</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 7 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '70px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr><td colSpan={7} className={styles.emptyCell}>
                  <div className={styles.emptyState}>
                    <p className={styles.emptyTitle}>Tidak ada data</p>
                    <p className={styles.emptySubtitle}>Belum ada kelas yang terdaftar.</p>
                  </div>
                </td></tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.name}</span></td>
                    <td className={styles.td}>{item.gradeLevel}</td>
                    <td className={styles.td}>{item.major ? (MAJOR_LABELS[item.major] ?? item.major) : '—'}</td>
                    <td className={styles.td}>{item.currentCount}/{item.capacity || '—'}</td>
                    <td className={styles.td}>{item.academicYear?.name ?? '—'}</td>
                    <td className={styles.td}>{item.homeroomTeacher?.fullName ?? '—'}</td>
                    <td className={styles.td}>
                      <button type="button" className={styles.deleteBtn} onClick={() => handleDelete(item)} disabled={deleteMutation.isPending}>
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
            <button type="button" className={styles.pageBtn} disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>Sebelumnya</button>
            <span className={styles.pageInfo}>Halaman {page} dari {totalPages}</span>
            <button type="button" className={styles.pageBtn} disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>Berikutnya</button>
          </div>
        )}
      </div>
    </div>
  )
}
