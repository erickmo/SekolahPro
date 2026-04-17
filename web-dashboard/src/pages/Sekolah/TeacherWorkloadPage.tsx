import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { teacherWorkloadService } from '@/services/sprint6.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { TeacherWorkload, FulfillmentStatus, Semester } from '@/types/sprint6.types'
import { FULFILLMENT_STATUS_LABELS, SEMESTER_LABELS } from '@/types/sprint6.types'
import styles from './TeacherWorkloadPage.module.css'

const PAGE_SIZE = 20
const MAX_HOURS_DISPLAY = 40

const SEMESTER_OPTIONS: { value: Semester | ''; label: string }[] = [
  { value: '', label: 'Semua Semester' },
  ...Object.entries(SEMESTER_LABELS).map(([value, label]) => ({
    value: value as Semester,
    label,
  })),
]

const FULFILLMENT_OPTIONS: { value: FulfillmentStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(FULFILLMENT_STATUS_LABELS).map(([value, label]) => ({
    value: value as FulfillmentStatus,
    label,
  })),
]

function getFulfillmentClass(status: FulfillmentStatus): string {
  switch (status) {
    case 'kurang': return styles.badgeKurang
    case 'terpenuhi': return styles.badgeTerpenuhi
    case 'lebih': return styles.badgeLebih
    default: return styles.badgeDefault
  }
}

function getHoursFillClass(status: FulfillmentStatus): string {
  switch (status) {
    case 'kurang': return styles.hoursFillKurang
    case 'terpenuhi': return styles.hoursFillTerpenuhi
    case 'lebih': return styles.hoursFillLebih
    default: return styles.hoursFillTerpenuhi
  }
}

export default function TeacherWorkloadPage() {
  const [search, setSearch] = useState('')
  const [semester, setSemester] = useState<Semester | ''>('')
  const [fulfillment, setFulfillment] = useState<FulfillmentStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { teacher_name: debouncedSearch } : {}),
    ...(semester ? { semester } : {}),
    ...(fulfillment ? { fulfillment_status: fulfillment } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.teacherWorkload, filters],
    queryFn: () => teacherWorkloadService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => teacherWorkloadService.delete(id),
    onSuccess: () => {
      toast.success('Data beban mengajar berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.teacherWorkload] })
    },
    onError: () => {
      toast.error('Gagal menghapus data beban mengajar')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || semester !== '' || fulfillment !== ''

  function handleReset() {
    setSearch('')
    setSemester('')
    setFulfillment('')
    setPage(1)
  }

  function handleDelete(item: TeacherWorkload) {
    if (!confirm(`Hapus data beban mengajar ${item.teacherName} semester ${SEMESTER_LABELS[item.semester]}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Beban Mengajar Guru</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Hitung Ulang
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
          value={semester}
          onChange={(e) => { setSemester(e.target.value as Semester | ''); setPage(1) }}
          aria-label="Filter semester"
        >
          {SEMESTER_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={fulfillment}
          onChange={(e) => { setFulfillment(e.target.value as FulfillmentStatus | ''); setPage(1) }}
          aria-label="Filter pemenuhan"
        >
          {FULFILLMENT_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Semester</th>
                <th className={styles.th}>Jam Mengajar</th>
                <th className={styles.th}>Jam Tambahan</th>
                <th className={styles.th}>Total JP</th>
                <th className={styles.th}>Minimum</th>
                <th className={styles.th}>Progres</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 10 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '70px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={10} className={styles.emptyCell}>
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada data beban mengajar guru.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => {
                  const pct = Math.min((item.totalHours / MAX_HOURS_DISPLAY) * 100, 100)
                  return (
                    <tr key={item.id} className={styles.row}>
                      <td className={styles.td}><span className={styles.name}>{item.teacherName || '-'}</span></td>
                      <td className={styles.td}><span className={styles.code}>{item.teacherNip || '-'}</span></td>
                      <td className={styles.td}>{SEMESTER_LABELS[item.semester] ?? item.semester}</td>
                      <td className={styles.td}>{item.teachingHours} JP</td>
                      <td className={styles.td}>{item.additionalHours} JP</td>
                      <td className={styles.td}><strong>{item.totalHours} JP</strong></td>
                      <td className={styles.td}>{item.minimumRequired} JP</td>
                      <td className={styles.td}>
                        <div className={styles.hoursBar}>
                          <div className={styles.hoursTrack}>
                            <div
                              className={`${styles.hoursFill} ${getHoursFillClass(item.fulfillmentStatus)}`}
                              style={{ width: `${pct}%` }}
                            />
                          </div>
                          <span>{Math.round(pct)}%</span>
                        </div>
                      </td>
                      <td className={styles.td}>
                        <span className={`${styles.badge} ${getFulfillmentClass(item.fulfillmentStatus)}`}>
                          {FULFILLMENT_STATUS_LABELS[item.fulfillmentStatus] ?? item.fulfillmentStatus}
                        </span>
                      </td>
                      <td className={styles.td}>
                        <button
                          type="button"
                          className={styles.deleteBtn}
                          onClick={() => handleDelete(item)}
                          disabled={deleteMutation.isPending}
                          aria-label={`Hapus beban mengajar ${item.teacherName}`}
                        >
                          <Trash2 size={14} />
                        </button>
                      </td>
                    </tr>
                  )
                })
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
