import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { subjectsService } from '@/services/phase5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Subject, SubjectGroup, SubjectStatus } from '@/types/phase5.types'
import { SUBJECT_GROUP_LABELS, SUBJECT_STATUS_LABELS } from '@/types/phase5.types'
import styles from './SubjectPage.module.css'

const PAGE_SIZE = 20

const GROUP_OPTIONS: { value: SubjectGroup | ''; label: string }[] = [
  { value: '', label: 'Semua Kelompok' },
  { value: 'science', label: 'IPA' },
  { value: 'social', label: 'IPS' },
  { value: 'language', label: 'Bahasa' },
  { value: 'mathematics', label: 'Matematika' },
  { value: 'religion', label: 'Pendidikan Agama' },
  { value: 'arts', label: 'Seni' },
  { value: 'physical', label: 'PJOK' },
  { value: 'technology', label: 'Teknologi' },
  { value: 'other', label: 'Lainnya' },
]

const STATUS_OPTIONS: { value: SubjectStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'draft', label: 'Draft' },
  { value: 'active', label: 'Aktif' },
  { value: 'inactive', label: 'Nonaktif' },
  { value: 'archived', label: 'Diarsipkan' },
]

function getGroupClass(group: SubjectGroup): string {
  switch (group) {
    case 'science': return styles.badgeScience
    case 'social': return styles.badgeSocial
    case 'language': return styles.badgeLanguage
    case 'mathematics': return styles.badgeMathematics
    case 'religion': return styles.badgeReligion
    case 'arts': return styles.badgeArts
    case 'physical': return styles.badgePhysical
    case 'technology': return styles.badgeTechnology
    default: return styles.badgeOther
  }
}

function getStatusClass(status: SubjectStatus): string {
  switch (status) {
    case 'active': return styles.badgeActive
    case 'draft': return styles.badgeDraft
    case 'inactive': return styles.badgeInactive
    case 'archived': return styles.badgeArchived
    default: return styles.badgeDefault
  }
}

export default function SubjectPage() {
  const [search, setSearch] = useState('')
  const [group, setGroup] = useState<SubjectGroup | ''>('')
  const [status, setStatus] = useState<SubjectStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(group ? { group } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.subjects, filters],
    queryFn: () => subjectsService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => subjectsService.delete(id),
    onSuccess: () => {
      toast.success('Mata pelajaran berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.subjects] })
    },
    onError: () => {
      toast.error('Gagal menghapus mata pelajaran')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || group !== '' || status !== ''

  const totalActive = items.filter((item) => item.status === 'active').length
  const totalWeight = items.reduce((sum, item) => sum + item.weightKnowledge + item.weightSkill, 0)

  function handleReset() {
    setSearch('')
    setGroup('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Subject) {
    if (!confirm(`Hapus mata pelajaran "${item.name}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Mata Pelajaran</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Mapel
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Aktif</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{totalActive}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Bobot</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{totalWeight}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Data</div>
          <div className={styles.summaryValue}>{total}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari kode atau nama mata pelajaran..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={group}
          onChange={(e) => { setGroup(e.target.value as SubjectGroup | ''); setPage(1) }}
          aria-label="Filter kelompok"
        >
          {GROUP_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as SubjectStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Kode</th>
                <th className={styles.th}>Nama</th>
                <th className={styles.th}>Kelompok</th>
                <th className={styles.th}>Bobot Pengetahuan</th>
                <th className={styles.th}>Bobot Keterampilan</th>
                <th className={styles.th}>Status</th>
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
                      <p className={styles.emptySubtitle}>Belum ada mata pelajaran yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.code || '-'}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.name}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getGroupClass(item.group)}`}>
                        {SUBJECT_GROUP_LABELS[item.group] ?? item.group}
                      </span>
                    </td>
                    <td className={styles.td}>{item.weightKnowledge}</td>
                    <td className={styles.td}>{item.weightSkill}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {SUBJECT_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus mata pelajaran ${item.name}`}
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
