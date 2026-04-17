import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { teachingJournalService } from '@/services/phase5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { TeachingJournal, TeachingJournalStatus } from '@/types/phase5.types'
import { TEACHING_JOURNAL_STATUS_LABELS } from '@/types/phase5.types'
import styles from './TeachingJournalPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: TeachingJournalStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'planned', label: 'Direncanakan' },
  { value: 'completed', label: 'Selesai' },
  { value: 'cancelled', label: 'Dibatalkan' },
  { value: 'rescheduled', label: 'Dijadwalkan Ulang' },
]

const SEMESTER_OPTIONS = [
  { value: '', label: 'Semua Semester' },
  { value: 'ganjil', label: 'Ganjil' },
  { value: 'genap', label: 'Genap' },
]

function getStatusClass(status: TeachingJournalStatus): string {
  switch (status) {
    case 'planned': return styles.badgePlanned
    case 'completed': return styles.badgeCompleted
    case 'cancelled': return styles.badgeCancelled
    case 'rescheduled': return styles.badgeRescheduled
    default: return styles.badgeDefault
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}

export default function TeachingJournalPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<TeachingJournalStatus | ''>('')
  const [semester, setSemester] = useState('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(status ? { status } : {}),
    ...(semester ? { semester } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.teachingJournal, filters],
    queryFn: () => teachingJournalService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => teachingJournalService.delete(id),
    onSuccess: () => {
      toast.success('Jurnal mengajar berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.teachingJournal] })
    },
    onError: () => {
      toast.error('Gagal menghapus jurnal mengajar')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || status !== '' || semester !== ''

  const totalCompleted = items.filter((item) => item.status === 'completed').length
  const totalPlanned = items.filter((item) => item.status === 'planned').length

  function handleReset() {
    setSearch('')
    setStatus('')
    setSemester('')
    setPage(1)
  }

  function handleDelete(item: TeachingJournal) {
    if (!confirm(`Hapus jurnal "${item.subjectName}" tanggal ${item.date}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Jurnal Mengajar</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Jurnal
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Selesai</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{totalCompleted}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Direncanakan</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{totalPlanned}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Jurnal</div>
          <div className={styles.summaryValue}>{total}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari guru, mapel, atau topik..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as TeachingJournalStatus | ''); setPage(1) }}
          aria-label="Filter status"
        >
          {STATUS_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Tanggal</th>
                <th className={styles.th}>Mata Pelajaran</th>
                <th className={styles.th}>Guru</th>
                <th className={styles.th}>Kelas</th>
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
                      <p className={styles.emptySubtitle}>Belum ada jurnal mengajar yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{formatDate(item.date)}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.subjectName || '-'}</span></td>
                    <td className={styles.td}>{item.teacherName || '-'}</td>
                    <td className={styles.td}>{item.classRoomName || '-'}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {TEACHING_JOURNAL_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>{item.semester || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus jurnal ${item.subjectName}`}
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
