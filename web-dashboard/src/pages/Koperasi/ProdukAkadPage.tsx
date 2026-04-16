import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { produkAkadService } from '@/services/koperasi.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { ProdukAkad, ProdukAkadType, ProdukAkadStatus } from '@/types/koperasi.types'
import {
  PRODUK_AKAD_TYPE_LABELS,
  AKAD_TYPE_LABELS,
  PRODUK_AKAD_STATUS_LABELS,
} from '@/types/koperasi.types'
import styles from './ProdukAkadPage.module.css'

const PAGE_SIZE = 20

const TYPE_OPTIONS: { value: ProdukAkadType | ''; label: string }[] = [
  { value: '', label: 'Semua Type' },
  { value: 'pembiayaan', label: 'Pembiayaan' },
  { value: 'tabungan', label: 'Tabungan' },
  { value: 'deposito', label: 'Deposito' },
]

const STATUS_OPTIONS: { value: ProdukAkadStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'draft', label: 'Draft' },
  { value: 'active', label: 'Aktif' },
  { value: 'inactive', label: 'Nonaktif' },
  { value: 'archived', label: 'Diarsipkan' },
]

function getStatusClass(status: ProdukAkadStatus): string {
  switch (status) {
    case 'active': return styles.badgeActive
    case 'draft': return styles.badgeDraft
    case 'inactive': return styles.badgeInactive
    case 'archived': return styles.badgeArchived
    default: return styles.badgeDefault
  }
}

function getTypeClass(type: ProdukAkadType): string {
  switch (type) {
    case 'pembiayaan': return styles.badgePembiayaan
    case 'tabungan': return styles.badgeTabungan
    case 'deposito': return styles.badgeDeposito
    default: return styles.badgeDefault
  }
}

export default function ProdukAkadPage() {
  const [search, setSearch] = useState('')
  const [type, setType] = useState<ProdukAkadType | ''>('')
  const [status, setStatus] = useState<ProdukAkadStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { nama_produk: debouncedSearch } : {}),
    ...(type ? { type } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.produkAkad, filters],
    queryFn: () => produkAkadService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => produkAkadService.delete(id),
    onSuccess: () => {
      toast.success('Produk akad berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.produkAkad] })
    },
    onError: () => {
      toast.error('Gagal menghapus produk akad')
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

  function handleDelete(item: ProdukAkad) {
    if (!confirm(`Hapus produk akad "${item.namaProduk}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Produk Akad</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Produk Akad
        </button>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari produk akad..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={type}
          onChange={(e) => { setType(e.target.value as ProdukAkadType | ''); setPage(1) }}
          aria-label="Filter type"
        >
          {TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as ProdukAkadStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Nama Produk</th>
                <th className={styles.th}>Type</th>
                <th className={styles.th}>Akad Type</th>
                <th className={styles.th}>Status</th>
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
                      <p className={styles.emptySubtitle}>Belum ada produk akad yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.kode}</span></td>
                    <td className={styles.td}><span className={styles.name}>{item.namaProduk}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getTypeClass(item.type)}`}>
                        {PRODUK_AKAD_TYPE_LABELS[item.type] ?? item.type}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${styles.badgeAkad}`}>
                        {AKAD_TYPE_LABELS[item.akadType] ?? item.akadType}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {PRODUK_AKAD_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus ${item.namaProduk}`}
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
