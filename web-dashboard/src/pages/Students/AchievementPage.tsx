import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { achievementService } from '@/services/sprint5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Achievement, AchievementType, AchievementLevel } from '@/types/sprint5.types'
import { ACHIEVEMENT_TYPE_LABELS, ACHIEVEMENT_LEVEL_LABELS } from '@/types/sprint5.types'
import styles from './AchievementPage.module.css'

const PAGE_SIZE = 20

const TYPE_OPTIONS: { value: AchievementType | ''; label: string }[] = [
  { value: '', label: 'Semua Jenis Prestasi' },
  ...Object.entries(ACHIEVEMENT_TYPE_LABELS).map(([value, label]) => ({
    value: value as AchievementType,
    label,
  })),
]

const LEVEL_OPTIONS: { value: AchievementLevel | ''; label: string }[] = [
  { value: '', label: 'Semua Tingkat' },
  ...Object.entries(ACHIEVEMENT_LEVEL_LABELS).map(([value, label]) => ({
    value: value as AchievementLevel,
    label,
  })),
]

function getLevelClass(level: AchievementLevel): string {
  switch (level) {
    case 'school': return styles.badgeSchool
    case 'district': return styles.badgeDistrict
    case 'province': return styles.badgeProvince
    case 'national': return styles.badgeNational
    case 'international': return styles.badgeInternational
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

export default function AchievementPage() {
  const [search, setSearch] = useState('')
  const [achievementType, setAchievementType] = useState<AchievementType | ''>('')
  const [level, setLevel] = useState<AchievementLevel | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { student_name: debouncedSearch } : {}),
    ...(achievementType ? { achievement_type: achievementType } : {}),
    ...(level ? { level } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.achievement, filters],
    queryFn: () => achievementService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => achievementService.delete(id),
    onSuccess: () => {
      toast.success('Data prestasi berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.achievement] })
    },
    onError: () => {
      toast.error('Gagal menghapus data prestasi')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || achievementType !== '' || level !== ''

  function handleReset() {
    setSearch('')
    setAchievementType('')
    setLevel('')
    setPage(1)
  }

  function handleDelete(item: Achievement) {
    if (!confirm(`Hapus prestasi "${item.title}" - ${item.studentName}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Data Prestasi</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Prestasi
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nama siswa / prestasi..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={achievementType}
          onChange={(e) => { setAchievementType(e.target.value as AchievementType | ''); setPage(1) }}
          aria-label="Filter jenis prestasi"
        >
          {TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={level}
          onChange={(e) => { setLevel(e.target.value as AchievementLevel | ''); setPage(1) }}
          aria-label="Filter tingkat"
        >
          {LEVEL_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Siswa</th>
                <th className={styles.th}>Jenis</th>
                <th className={styles.th}>Tingkat</th>
                <th className={styles.th}>Tanggal</th>
                <th className={styles.th}>Penyelenggara</th>
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
                      <p className={styles.emptySubtitle}>Belum ada data prestasi.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.title}</span></td>
                    <td className={styles.td}>{item.studentName || '-'}</td>
                    <td className={styles.td}>
                      <span className={styles.badgeType}>
                        {ACHIEVEMENT_TYPE_LABELS[item.achievementType] ?? item.achievementType}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getLevelClass(item.level)}`}>
                        {ACHIEVEMENT_LEVEL_LABELS[item.level] ?? item.level}
                      </span>
                    </td>
                    <td className={styles.td}>{formatDate(item.date)}</td>
                    <td className={styles.td}>{item.organizer || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus prestasi ${item.title}`}
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
