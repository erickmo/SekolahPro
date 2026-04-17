import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { teacherSubstitutionService, dutyScheduleService } from '@/services/sprint6.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { TeacherSubstitution, SubstitutionStatus, DutySchedule } from '@/types/sprint6.types'
import {
  SUBSTITUTION_STATUS_LABELS,
  SUBSTITUTION_REASON_LABELS,
  DUTY_TYPE_LABELS,
  DAY_OF_WEEK_LABELS,
} from '@/types/sprint6.types'
import styles from './TeacherSubstitutionPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: SubstitutionStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(SUBSTITUTION_STATUS_LABELS).map(([value, label]) => ({
    value: value as SubstitutionStatus,
    label,
  })),
]

function getStatusClass(status: SubstitutionStatus): string {
  switch (status) {
    case 'accepted': return styles.badgeAccepted
    case 'declined': return styles.badgeDeclined
    case 'pending': return styles.badgePending
    case 'notified': return styles.badgeNotified
    case 'in_progress': return styles.badgeInProgress
    case 'completed': return styles.badgeCompleted
    case 'cancelled': return styles.badgeCancelled
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

export default function TeacherSubstitutionPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<SubstitutionStatus | ''>('')
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
    queryKey: [QK.teacherSubstitution, filters],
    queryFn: () => teacherSubstitutionService.list(filters),
  })

  const { data: dutyData, isLoading: dutyLoading } = useQuery({
    queryKey: [QK.dutySchedule],
    queryFn: () => dutyScheduleService.list({ _limit: 100 }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => teacherSubstitutionService.delete(id),
    onSuccess: () => {
      toast.success('Data substitusi berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.teacherSubstitution] })
    },
    onError: () => {
      toast.error('Gagal menghapus data substitusi')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const dutyItems: DutySchedule[] = dutyData?.items ?? []
  const hasFilters = search !== '' || status !== ''

  function handleReset() {
    setSearch('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: TeacherSubstitution) {
    if (!confirm(`Hapus substitusi guru "${item.substituteTeacherName}" menggantikan "${item.originalTeacherName}" tanggal ${formatDate(item.substitutionDate)}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Penggantian Guru & Jadwal Piket</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} substitusi ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Buat Substitusi
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
          onChange={(e) => { setStatus(e.target.value as SubstitutionStatus | ''); setPage(1) }}
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
        <div className={styles.sectionTitle}>Daftar Substitusi Guru</div>
        <div className={styles.tableScroll}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.th}>Tanggal</th>
                <th className={styles.th}>Guru Asli</th>
                <th className={styles.th}>Guru Pengganti</th>
                <th className={styles.th}>Kelas</th>
                <th className={styles.th}>Alasan</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Catatan</th>
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
                      <p className={styles.emptySubtitle}>Belum ada data substitusi guru.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}>{formatDate(item.substitutionDate)}</td>
                    <td className={styles.td}><span className={styles.name}>{item.originalTeacherName || '-'}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.substituteTeacherName || '-'}</span></td>
                    <td className={styles.td}>{item.classRoomName || '-'}</td>
                    <td className={styles.td}>{SUBSTITUTION_REASON_LABELS[item.reasonType] ?? item.reasonType}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {SUBSTITUTION_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>{item.notes || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus substitusi ${item.substituteTeacherName}`}
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

      <div className={styles.tableCard}>
        <div className={styles.sectionTitle}>Jadwal Piket Hari Ini</div>
        <div className={styles.tableScroll}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.th}>Guru</th>
                <th className={styles.th}>NIP</th>
                <th className={styles.th}>Hari</th>
                <th className={styles.th}>Tipe Piket</th>
                <th className={styles.th}>Waktu Mulai</th>
                <th className={styles.th}>Waktu Selesai</th>
                <th className={styles.th}>Catatan</th>
              </tr>
            </thead>
            <tbody>
              {dutyLoading ? (
                Array.from({ length: 3 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 7 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '60px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : dutyItems.length === 0 ? (
                <tr>
                  <td colSpan={7} className={styles.emptyCell}>
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada jadwal piket yang diatur.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                dutyItems.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.teacherName || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.teacherNip || '-'}</span></td>
                    <td className={styles.td}>{DAY_OF_WEEK_LABELS[item.dayOfWeek] ?? item.dayOfWeek}</td>
                    <td className={styles.td}>{DUTY_TYPE_LABELS[item.dutyType] ?? item.dutyType}</td>
                    <td className={styles.td}>{item.startTime || '-'}</td>
                    <td className={styles.td}>{item.endTime || '-'}</td>
                    <td className={styles.td}>{item.notes || '-'}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
