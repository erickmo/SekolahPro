import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { teachingScheduleService } from '@/services/phase5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { TeachingSchedule, TeachingScheduleDay } from '@/types/phase5.types'
import { TEACHING_SCHEDULE_DAY_LABELS } from '@/types/phase5.types'
import styles from './TeachingSchedulePage.module.css'

const PAGE_SIZE = 20

const DAY_OPTIONS: { value: TeachingScheduleDay | ''; label: string }[] = [
  { value: '', label: 'Semua Hari' },
  { value: 'monday', label: 'Senin' },
  { value: 'tuesday', label: 'Selasa' },
  { value: 'wednesday', label: 'Rabu' },
  { value: 'thursday', label: 'Kamis' },
  { value: 'friday', label: 'Jumat' },
  { value: 'saturday', label: 'Sabtu' },
]

const SEMESTER_OPTIONS = [
  { value: '', label: 'Semua Semester' },
  { value: 'ganjil', label: 'Ganjil' },
  { value: 'genap', label: 'Genap' },
]

function getDayClass(day: TeachingScheduleDay): string {
  switch (day) {
    case 'monday': return styles.badgeMonday
    case 'tuesday': return styles.badgeTuesday
    case 'wednesday': return styles.badgeWednesday
    case 'thursday': return styles.badgeThursday
    case 'friday': return styles.badgeFriday
    case 'saturday': return styles.badgeSaturday
    default: return styles.badgeDefault
  }
}

export default function TeachingSchedulePage() {
  const [search, setSearch] = useState('')
  const [dayOfWeek, setDayOfWeek] = useState<TeachingScheduleDay | ''>('')
  const [semester, setSemester] = useState('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(dayOfWeek ? { day_of_week: dayOfWeek } : {}),
    ...(semester ? { semester } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.teachingSchedule, filters],
    queryFn: () => teachingScheduleService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => teachingScheduleService.delete(id),
    onSuccess: () => {
      toast.success('Jadwal mengajar berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.teachingSchedule] })
    },
    onError: () => {
      toast.error('Gagal menghapus jadwal mengajar')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || dayOfWeek !== '' || semester !== ''

  const uniqueTeachers = new Set(items.map((item) => item.teacherId)).size
  const uniqueSubjects = new Set(items.map((item) => item.subjectId)).size

  function handleReset() {
    setSearch('')
    setDayOfWeek('')
    setSemester('')
    setPage(1)
  }

  function handleDelete(item: TeachingSchedule) {
    if (!confirm(`Hapus jadwal "${item.subjectName}" pada ${TEACHING_SCHEDULE_DAY_LABELS[item.dayOfWeek]}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Jadwal Mengajar</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Jadwal
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Guru</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{uniqueTeachers}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Mapel</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{uniqueSubjects}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Jadwal</div>
          <div className={styles.summaryValue}>{total}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari guru, mapel, atau kelas..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={dayOfWeek}
          onChange={(e) => { setDayOfWeek(e.target.value as TeachingScheduleDay | ''); setPage(1) }}
          aria-label="Filter hari"
        >
          {DAY_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={semester}
          onChange={(e) => { setSemester(e.target.value); setPage(1) }}
          aria-label="Filter semester"
        >
          {SEMESTER_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Hari</th>
                <th className={styles.th}>Slot</th>
                <th className={styles.th}>Mata Pelajaran</th>
                <th className={styles.th}>Guru</th>
                <th className={styles.th}>Kelas</th>
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
                      <p className={styles.emptySubtitle}>Belum ada jadwal mengajar yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getDayClass(item.dayOfWeek)}`}>
                        {TEACHING_SCHEDULE_DAY_LABELS[item.dayOfWeek] ?? item.dayOfWeek}
                      </span>
                    </td>
                    <td className={styles.td}><span className={styles.code}>{item.slot}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.subjectName || '-'}</span></td>
                    <td className={styles.td}>{item.teacherName || '-'}</td>
                    <td className={styles.td}>{item.classRoomName || '-'}</td>
                    <td className={styles.td}>{item.semester || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus jadwal ${item.subjectName}`}
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
