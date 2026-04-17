import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { academicCalendarService } from '@/services/phase5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { AcademicCalendar, AcademicCalendarEventType, AcademicCalendarType } from '@/types/phase5.types'
import { ACADEMIC_CALENDAR_EVENT_TYPE_LABELS, ACADEMIC_CALENDAR_TYPE_LABELS } from '@/types/phase5.types'
import styles from './AcademicCalendarPage.module.css'

const PAGE_SIZE = 20

const EVENT_TYPE_OPTIONS: { value: AcademicCalendarEventType | ''; label: string }[] = [
  { value: '', label: 'Semua Tipe Acara' },
  { value: 'semester_start', label: 'Mulai Semester' },
  { value: 'semester_end', label: 'Akhir Semester' },
  { value: 'holiday', label: 'Libur' },
  { value: 'exam', label: 'Ujian' },
  { value: 'event', label: 'Acara' },
  { value: 'break', label: 'Jeda' },
  { value: 'other', label: 'Lainnya' },
]

const CALENDAR_TYPE_OPTIONS: { value: AcademicCalendarType | ''; label: string }[] = [
  { value: '', label: 'Semua Kalender' },
  { value: 'national', label: 'Nasional' },
  { value: 'school', label: 'Sekolah' },
  { value: 'academic', label: 'Akademik' },
]

function getEventTypeClass(eventType: AcademicCalendarEventType): string {
  switch (eventType) {
    case 'semester_start': return styles.badgeSemesterStart
    case 'semester_end': return styles.badgeSemesterEnd
    case 'holiday': return styles.badgeHoliday
    case 'exam': return styles.badgeExam
    case 'event': return styles.badgeEvent
    case 'break': return styles.badgeBreak
    default: return styles.badgeOther
  }
}

function getCalendarTypeClass(calendarType: AcademicCalendarType): string {
  switch (calendarType) {
    case 'national': return styles.badgeNational
    case 'school': return styles.badgeSchool
    case 'academic': return styles.badgeAcademic
    default: return styles.badgeDefault
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}

export default function AcademicCalendarPage() {
  const [search, setSearch] = useState('')
  const [eventType, setEventType] = useState<AcademicCalendarEventType | ''>('')
  const [calendarType, setCalendarType] = useState<AcademicCalendarType | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(eventType ? { event_type: eventType } : {}),
    ...(calendarType ? { calendar_type: calendarType } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.academicCalendar, filters],
    queryFn: () => academicCalendarService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => academicCalendarService.delete(id),
    onSuccess: () => {
      toast.success('Kalender akademik berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.academicCalendar] })
    },
    onError: () => {
      toast.error('Gagal menghapus kalender akademik')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || eventType !== '' || calendarType !== ''

  const totalHolidays = items.filter((item) => item.eventType === 'holiday').length
  const totalExams = items.filter((item) => item.eventType === 'exam').length

  function handleReset() {
    setSearch('')
    setEventType('')
    setCalendarType('')
    setPage(1)
  }

  function handleDelete(item: AcademicCalendar) {
    if (!confirm(`Hapus event "${item.title}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Kalender Akademik</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Event
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Libur</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{totalHolidays}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Ujian</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{totalExams}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Event</div>
          <div className={styles.summaryValue}>{total}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari judul event..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={eventType}
          onChange={(e) => { setEventType(e.target.value as AcademicCalendarEventType | ''); setPage(1) }}
          aria-label="Filter tipe acara"
        >
          {EVENT_TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={calendarType}
          onChange={(e) => { setCalendarType(e.target.value as AcademicCalendarType | ''); setPage(1) }}
          aria-label="Filter kalender"
        >
          {CALENDAR_TYPE_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Tipe Acara</th>
                <th className={styles.th}>Tanggal Mulai</th>
                <th className={styles.th}>Tanggal Selesai</th>
                <th className={styles.th}>Jenis Kalender</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 6 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={6} className={styles.emptyCell}>
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada kalender akademik yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.title}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getEventTypeClass(item.eventType)}`}>
                        {ACADEMIC_CALENDAR_EVENT_TYPE_LABELS[item.eventType] ?? item.eventType}
                      </span>
                    </td>
                    <td className={styles.td}>{formatDate(item.startDate)}</td>
                    <td className={styles.td}>{formatDate(item.endDate)}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getCalendarTypeClass(item.calendarType)}`}>
                        {ACADEMIC_CALENDAR_TYPE_LABELS[item.calendarType] ?? item.calendarType}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus event ${item.title}`}
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
