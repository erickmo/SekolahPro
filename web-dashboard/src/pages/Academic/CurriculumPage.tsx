import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { curriculaService } from '@/services/phase5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Curriculum, CurriculumType, CurriculumStatus } from '@/types/phase5.types'
import { CURRICULUM_TYPE_LABELS, CURRICULUM_STATUS_LABELS } from '@/types/phase5.types'
import styles from './CurriculumPage.module.css'

const PAGE_SIZE = 20

const TYPE_OPTIONS: { value: CurriculumType | ''; label: string }[] = [
  { value: '', label: 'Semua Tipe' },
  { value: 'national', label: 'Nasional' },
  { value: 'school', label: 'Sekolah' },
  { value: 'international', label: 'Internasional' },
  { value: 'hybrid', label: 'Hibrida' },
]

const STATUS_OPTIONS: { value: CurriculumStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'draft', label: 'Draft' },
  { value: 'active', label: 'Aktif' },
  { value: 'archived', label: 'Diarsipkan' },
]

function getTypeClass(type: CurriculumType): string {
  switch (type) {
    case 'national': return styles.badgeNational
    case 'school': return styles.badgeSchool
    case 'international': return styles.badgeInternational
    case 'hybrid': return styles.badgeHybrid
    default: return styles.badgeDefault
  }
}

function getStatusClass(status: CurriculumStatus): string {
  switch (status) {
    case 'active': return styles.badgeActive
    case 'draft': return styles.badgeDraft
    case 'archived': return styles.badgeArchived
    default: return styles.badgeDefault
  }
}

export default function CurriculumPage() {
  const [search, setSearch] = useState('')
  const [type, setType] = useState<CurriculumType | ''>('')
  const [status, setStatus] = useState<CurriculumStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(type ? { type } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.curricula, filters],
    queryFn: () => curriculaService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => curriculaService.delete(id),
    onSuccess: () => {
      toast.success('Kurikulum berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.curricula] })
    },
    onError: () => {
      toast.error('Gagal menghapus kurikulum')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || type !== '' || status !== ''

  const totalActive = items.filter((item) => item.status === 'active').length
  const totalDraft = items.filter((item) => item.status === 'draft').length

  function handleReset() {
    setSearch('')
    setType('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Curriculum) {
    if (!confirm(`Hapus kurikulum "${item.name}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Kurikulum</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Kurikulum
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Aktif</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{totalActive}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Draft</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{totalDraft}</div>
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
            placeholder="Cari kode atau nama kurikulum..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={type}
          onChange={(e) => { setType(e.target.value as CurriculumType | ''); setPage(1) }}
          aria-label="Filter tipe"
        >
          {TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as CurriculumStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Tipe</th>
                <th className={styles.th}>Jenjang</th>
                <th className={styles.th}>Fase</th>
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
                      <p className={styles.emptySubtitle}>Belum ada kurikulum yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.code || '-'}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.name}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getTypeClass(item.type)}`}>
                        {CURRICULUM_TYPE_LABELS[item.type] ?? item.type}
                      </span>
                    </td>
                    <td className={styles.td}>{item.gradeLevel.toUpperCase()}</td>
                    <td className={styles.td}>{item.phase.toUpperCase()}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {CURRICULUM_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus kurikulum ${item.name}`}
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
