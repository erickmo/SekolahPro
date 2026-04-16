import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { nasabahService } from '@/services/koperasi.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Nasabah, NasabahType, NasabahStatus } from '@/types/koperasi.types'
import { NASABAH_TYPE_LABELS, NASABAH_STATUS_LABELS } from '@/types/koperasi.types'
import styles from './NasabahPage.module.css'

const PAGE_SIZE = 20

const NIK_MASK_LENGTH = 4

function maskNik(nik: string): string {
  if (!nik || nik.length <= NIK_MASK_LENGTH) return nik
  return '*'.repeat(nik.length - NIK_MASK_LENGTH) + nik.slice(-NIK_MASK_LENGTH)
}

const TYPE_OPTIONS: { value: NasabahType | ''; label: string }[] = [
  { value: '', label: 'Semua Type' },
  { value: 'perorangan', label: 'Perorangan' },
  { value: 'badan_usaha', label: 'Badan Usaha' },
  { value: 'kelompok', label: 'Kelompok' },
]

const STATUS_OPTIONS: { value: NasabahStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'pending', label: 'Pending' },
  { value: 'active', label: 'Aktif' },
  { value: 'frozen', label: 'Dibekukan' },
  { value: 'closed', label: 'Ditutup' },
]

function getStatusClass(status: NasabahStatus): string {
  switch (status) {
    case 'active': return styles.badgeActive
    case 'pending': return styles.badgePending
    case 'frozen': return styles.badgeFrozen
    case 'closed': return styles.badgeClosed
    default: return styles.badgeDefault
  }
}

function getTypeClass(type: NasabahType): string {
  switch (type) {
    case 'perorangan': return styles.badgePerorangan
    case 'badan_usaha': return styles.badgeBadanUsaha
    case 'kelompok': return styles.badgeKelompok
    default: return styles.badgeDefault
  }
}

export default function NasabahPage() {
  const [search, setSearch] = useState('')
  const [type, setType] = useState<NasabahType | ''>('')
  const [status, setStatus] = useState<NasabahStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { nama_lengkap: debouncedSearch } : {}),
    ...(type ? { type } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.nasabah, filters],
    queryFn: () => nasabahService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => nasabahService.delete(id),
    onSuccess: () => {
      toast.success('Nasabah berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.nasabah] })
    },
    onError: () => {
      toast.error('Gagal menghapus nasabah')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || type !== '' || status !== ''

  function handleReset() {
    setSearch('')
    setType('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Nasabah) {
    if (!confirm(`Hapus nasabah "${item.namaLengkap}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Nasabah</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Nasabah
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nasabah..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={type}
          onChange={(e) => { setType(e.target.value as NasabahType | ''); setPage(1) }}
          aria-label="Filter type"
        >
          {TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as NasabahStatus | ''); setPage(1) }}
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
                <th className={styles.th}>No. Nasabah</th>
                <th className={styles.th}>Nama Lengkap</th>
                <th className={styles.th}>NIK</th>
                <th className={styles.th}>Type</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Tgl Daftar</th>
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
                      <p className={styles.emptySubtitle}>Belum ada nasabah yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.noNasabah}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.namaLengkap}</span></td>
                    <td className={styles.td}><span className={styles.nik}>{maskNik(item.nik)}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getTypeClass(item.type)}`}>
                        {NASABAH_TYPE_LABELS[item.type] ?? item.type}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {NASABAH_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>{item.tanggalDaftar || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus ${item.namaLengkap}`}
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
