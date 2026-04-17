import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { teacherAttendanceService } from '@/services/sprint6.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { TeacherAttendance, TeacherAttendanceStatus } from '@/types/sprint6.types'
import { TEACHER_ATT_STATUS_LABELS } from '@/types/sprint6.types'
import styles from './TeacherAttendancePage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: TeacherAttendanceStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(TEACHER_ATT_STATUS_LABELS).map(([value, label]) => ({
    value: value as TeacherAttendanceStatus,
    label,
  })),
]

function getStatusClass(status: TeacherAttendanceStatus): string {
  switch (status) {
    case 'present': return styles.badgePresent
    case 'sick': return styles.badgeSick
    case 'permitted': return styles.badgePermitted
    case 'absent': return styles.badgeAbsent
    case 'dinas_luar': return styles.badgeDinasLuar
    case 'cuti': return styles.badgeCuti
    case 'libur': return styles.badgeLibur
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

function formatTime(dateTimeStr: string): string {
  if (!dateTimeStr) return '-'
  try {
    return new Date(dateTimeStr).toLocaleTimeString('id-ID', {
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return '-'
  }
}

export default function TeacherAttendancePage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<TeacherAttendanceStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { teacher_name: debouncedSearch } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.teacherAttendance, filters],
    queryFn: () => teacherAttendanceService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => teacherAttendanceService.delete(id),
    onSuccess: () => {
      toast.success('Data absensi berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.teacherAttendance] })
    },
    onError: () => {
      toast.error('Gagal menghapus data absensi')
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

  function handleDelete(item: TeacherAttendance) {
    if (!confirm(`Hapus data absensi ${item.teacherName} tanggal ${formatDate(item.attendanceDate)}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Absensi Guru & Staff</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Input Manual
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
          value={status}
          onChange={(e) => { setStatus(e.target.value as TeacherAttendanceStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Guru</th>
                <th className={styles.th}>NIP</th>
                <th className={styles.th}>Tanggal</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Jam Masuk</th>
                <th className={styles.th}>Jam Keluar</th>
                <th className={styles.th}>Keterlambatan</th>
                <th className={styles.th}>Pulang Awal</th>
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
                      <p className={styles.emptySubtitle}>Belum ada data absensi guru.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.teacherName || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.teacherNip || '-'}</span></td>
                    <td className={styles.td}>{formatDate(item.attendanceDate)}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {TEACHER_ATT_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>{formatTime(item.clockIn)}</td>
                    <td className={styles.td}>{formatTime(item.clockOut)}</td>
                    <td className={styles.td}>
                      {item.lateMinutes > 0
                        ? <span className={styles.lateIndicator}>{item.lateMinutes} menit</span>
                        : <span className={styles.onTime}>Tepat waktu</span>
                      }
                    </td>
                    <td className={styles.td}>
                      {item.earlyLeaveMinutes > 0
                        ? <span className={styles.lateIndicator}>{item.earlyLeaveMinutes} menit</span>
                        : '-'
                      }
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus absensi ${item.teacherName}`}
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
