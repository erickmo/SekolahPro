import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { academicYearService } from '@/services/academic.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { AcademicYear, AcademicYearStatus } from '@/types/academic.types'
import { STATUS_LABELS } from '@/types/academic.types'
import styles from './AcademicYearsPage.module.css'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: AcademicYearStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'planning', label: 'Perencanaan' },
  { value: 'active', label: 'Aktif' },
  { value: 'closed', label: 'Ditutup' },
]

function getStatusClass(status: AcademicYearStatus): string {
  switch (status) {
    case 'active': return styles.badgeActive
    case 'planning': return styles.badgePlanning
    case 'closed': return styles.badgeClosed
    default: return styles.badgeDefault
  }
}

export default function AcademicYearsPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<AcademicYearStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { name: debouncedSearch } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.academicYears, filters],
    queryFn: () => academicYearService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => academicYearService.delete(id),
    onSuccess: () => {
      toast.success('Tahun ajaran berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.academicYears] })
    },
    onError: () => {
      toast.error('Gagal menghapus tahun ajaran')
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

  function handleDelete(item: AcademicYear) {
    if (!confirm(`Hapus tahun ajaran "${item.name}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Tahun Ajaran</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Tahun Ajaran
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari tahun ajaran..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as AcademicYearStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Nama</th>
                <th className={styles.th}>Kode</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Mulai</th>
                <th className={styles.th}>Selesai</th>
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
                      <p className={styles.emptySubtitle}>Belum ada tahun ajaran yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.name}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.code}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>{item.startDate || '—'}</td>
                    <td className={styles.td}>{item.endDate || '—'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus ${item.name}`}
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
