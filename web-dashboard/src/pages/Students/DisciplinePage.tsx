import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { disciplineService } from '@/services/sprint5.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Discipline, ViolationType, DisciplineStatus } from '@/types/sprint5.types'
import { VIOLATION_TYPE_LABELS, DISCIPLINE_STATUS_LABELS } from '@/types/sprint5.types'
import styles from './DisciplinePage.module.css'

const PAGE_SIZE = 20

const VIOLATION_OPTIONS: { value: ViolationType | ''; label: string }[] = [
  { value: '', label: 'Semua Jenis Pelanggaran' },
  ...Object.entries(VIOLATION_TYPE_LABELS).map(([value, label]) => ({
    value: value as ViolationType,
    label,
  })),
]

const STATUS_OPTIONS: { value: DisciplineStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(DISCIPLINE_STATUS_LABELS).map(([value, label]) => ({
    value: value as DisciplineStatus,
    label,
  })),
]

function getViolationClass(type: ViolationType): string {
  switch (type) {
    case 'minor': return styles.badgeMinor
    case 'moderate': return styles.badgeModerate
    case 'major': return styles.badgeMajor
    default: return styles.badgeDefault
  }
}

function getStatusClass(status: DisciplineStatus): string {
  switch (status) {
    case 'reported': return styles.badgeReported
    case 'reviewed': return styles.badgeReviewed
    case 'sanctioned': return styles.badgeSanctioned
    case 'resolved': return styles.badgeResolved
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

export default function DisciplinePage() {
  const [search, setSearch] = useState('')
  const [violationType, setViolationType] = useState<ViolationType | ''>('')
  const [status, setStatus] = useState<DisciplineStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { student_name: debouncedSearch } : {}),
    ...(violationType ? { violation_type: violationType } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.discipline, filters],
    queryFn: () => disciplineService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => disciplineService.delete(id),
    onSuccess: () => {
      toast.success('Catatan pelanggaran berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.discipline] })
    },
    onError: () => {
      toast.error('Gagal menghapus catatan pelanggaran')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || violationType !== '' || status !== ''

  function handleReset() {
    setSearch('')
    setViolationType('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Discipline) {
    if (!confirm(`Hapus catatan pelanggaran "${item.studentName}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Catatan Pelanggaran</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Pelanggaran
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nama siswa..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={violationType}
          onChange={(e) => { setViolationType(e.target.value as ViolationType | ''); setPage(1) }}
          aria-label="Filter jenis pelanggaran"
        >
          {VIOLATION_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as DisciplineStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Siswa</th>
                <th className={styles.th}>Jenis</th>
                <th className={styles.th}>Tanggal Kejadian</th>
                <th className={styles.th}>Deskripsi</th>
                <th className={styles.th}>Poin</th>
                <th className={styles.th}>Sanksi</th>
                <th className={styles.th}>Pelapor</th>
                <th className={styles.th}>Status</th>
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
                      <p className={styles.emptySubtitle}>Belum ada catatan pelanggaran.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.studentName || '-'}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getViolationClass(item.violationType)}`}>
                        {VIOLATION_TYPE_LABELS[item.violationType] ?? item.violationType}
                      </span>
                    </td>
                    <td className={styles.td}>{formatDate(item.incidentDate)}</td>
                    <td className={styles.td}>{item.description || '-'}</td>
                    <td className={styles.td}><span className={styles.points}>{item.points > 0 ? item.points : '-'}</span></td>
                    <td className={styles.td}>{item.sanction || '-'}</td>
                    <td className={styles.td}>{item.teacherName || '-'}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {DISCIPLINE_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus pelanggaran ${item.studentName}`}
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
