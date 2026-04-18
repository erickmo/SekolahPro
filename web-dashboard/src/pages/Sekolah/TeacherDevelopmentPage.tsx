import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { developmentActivityService } from '@/services/sprint7.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { DevelopmentActivity, ActivityStatus } from '@/types/sprint7.types'
import { ACTIVITY_STATUS_LABELS } from '@/types/sprint7.types'
import styles from './TeacherDevelopmentPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: ActivityStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(ACTIVITY_STATUS_LABELS).map(([value, label]) => ({
    value: value as ActivityStatus,
    label,
  })),
]

function getStatusBadgeClass(status: ActivityStatus): string {
  switch (status) {
    case 'registered': return styles.badgeYellow
    case 'attended': return styles.badgeBlue
    case 'completed': return styles.badgeGreen
    case 'cancelled': return styles.badgeRed
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

export default function TeacherDevelopmentPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<ActivityStatus | ''>('')
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
    queryKey: [QK.developmentActivities, filters],
    queryFn: () => developmentActivityService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => developmentActivityService.delete(id),
    onSuccess: () => {
      toast.success('Data pengembangan berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.developmentActivities] })
    },
    onError: () => {
      toast.error('Gagal menghapus data pengembangan')
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

  function handleDelete(item: DevelopmentActivity) {
    if (!confirm(`Hapus aktivitas "${item.title}" (${item.teacherName})?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Pengembangan Profesional (PKB)</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Aktivitas
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari guru atau aktivitas..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as ActivityStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Judul</th>
                <th className={styles.th}>Tipe Aktivitas</th>
                <th className={styles.th}>Tanggal Mulai</th>
                <th className={styles.th}>Angka Kredit</th>
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
                    <div>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada data pengembangan profesional.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.teacherName || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.teacherNip || '-'}</span></td>
                    <td className={styles.td}>{item.title || '-'}</td>
                    <td className={styles.td}>{item.activityType || '-'}</td>
                    <td className={styles.td}>{formatDate(item.startDate)}</td>
                    <td className={styles.td}>{item.creditPoints}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusBadgeClass(item.status)}`}>
                        {ACTIVITY_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus aktivitas ${item.title}`}
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
