import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { simpananPokokWajibService } from '@/services/koperasi.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { SimpananPokokWajib, SimpananPokokWajibJenis, SimpananPokokWajibStatus } from '@/types/koperasi.types'
import { SIMPANAN_POKOK_WAJIB_JENIS_LABELS, SIMPANAN_POKOK_WAJIB_STATUS_LABELS } from '@/types/koperasi.types'
import styles from './SimpananPokokWajibPage.module.css'

const PAGE_SIZE = 20

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

const JENIS_OPTIONS: { value: SimpananPokokWajibJenis | ''; label: string }[] = [
  { value: '', label: 'Semua Jenis' },
  { value: 'pokok', label: 'Pokok' },
  { value: 'wajib', label: 'Wajib' },
]

const STATUS_OPTIONS: { value: SimpananPokokWajibStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'pending', label: 'Pending' },
  { value: 'paid', label: 'Lunas' },
  { value: 'overdue', label: 'Tunggakan' },
  { value: 'refunded', label: 'Dikembalikan' },
]

function getStatusClass(status: SimpananPokokWajibStatus): string {
  switch (status) {
    case 'paid': return styles.badgePaid
    case 'pending': return styles.badgePending
    case 'overdue': return styles.badgeOverdue
    case 'refunded': return styles.badgeRefunded
    default: return styles.badgeDefault
  }
}

function getJenisClass(jenis: SimpananPokokWajibJenis): string {
  switch (jenis) {
    case 'pokok': return styles.badgePokok
    case 'wajib': return styles.badgeWajib
    default: return styles.badgeDefault
  }
}

export default function SimpananPokokWajibPage() {
  const [search, setSearch] = useState('')
  const [jenis, setJenis] = useState<SimpananPokokWajibJenis | ''>('')
  const [status, setStatus] = useState<SimpananPokokWajibStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { nasabah_nama: debouncedSearch } : {}),
    ...(jenis ? { jenis } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.simpananPokokWajib, filters],
    queryFn: () => simpananPokokWajibService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => simpananPokokWajibService.delete(id),
    onSuccess: () => {
      toast.success('Simpanan berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.simpananPokokWajib] })
    },
    onError: () => {
      toast.error('Gagal menghapus simpanan')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || jenis !== '' || status !== ''

  const totalTagihan = items.reduce((sum, item) => sum + item.nominal, 0)
  const totalTerbayar = items
    .filter((item) => item.status === 'paid')
    .reduce((sum, item) => sum + item.nominal, 0)
  const totalTunggakan = items
    .filter((item) => item.status === 'overdue')
    .reduce((sum, item) => sum + item.nominal, 0)

  function handleReset() {
    setSearch('')
    setJenis('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: SimpananPokokWajib) {
    if (!confirm(`Hapus simpanan "${item.nasabahNama}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Simpanan Pokok & Wajib</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Simpanan
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Tagihan</div>
          <div className={styles.summaryValue}>{formatRupiah(totalTagihan)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Terbayar</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{formatRupiah(totalTerbayar)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Tunggakan</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueRed}`}>{formatRupiah(totalTunggakan)}</div>
        </div>
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
          value={jenis}
          onChange={(e) => { setJenis(e.target.value as SimpananPokokWajibJenis | ''); setPage(1) }}
          aria-label="Filter jenis"
        >
          {JENIS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as SimpananPokokWajibStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Nasabah</th>
                <th className={styles.th}>No. Rekening</th>
                <th className={styles.th}>Jenis</th>
                <th className={styles.th}>Nominal</th>
                <th className={styles.th}>Periode</th>
                <th className={styles.th}>Jatuh Tempo</th>
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
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada simpanan pokok/wajib yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.name}>{item.nasabahNama || '-'}</span></td>
                    <td className={styles.td}><span className={styles.code}>{item.noRekening}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getJenisClass(item.jenis)}`}>
                        {SIMPANAN_POKOK_WAJIB_JENIS_LABELS[item.jenis] ?? item.jenis}
                      </span>
                    </td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.nominal)}</span></td>
                    <td className={styles.td}>{item.periode || '-'}</td>
                    <td className={styles.td}>{item.jatuhTempo || '-'}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {SIMPANAN_POKOK_WAJIB_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus simpanan ${item.nasabahNama}`}
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
