import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { kasService } from '@/services/koperasi-ops.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Kas, KasType, KasStatus } from '@/types/koperasi-ops.types'
import { KAS_TYPE_LABELS, KAS_STATUS_LABELS } from '@/types/koperasi-ops.types'
import styles from './KasPage.module.css'

const PAGE_SIZE = 20

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
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

const KAS_TYPE_OPTIONS: { value: KasType | ''; label: string }[] = [
  { value: '', label: 'Semua Tipe' },
  { value: 'masuk', label: 'Masuk' },
  { value: 'keluar', label: 'Keluar' },
]

const STATUS_OPTIONS: { value: KasStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(KAS_STATUS_LABELS).map(([value, label]) => ({
    value: value as KasStatus,
    label,
  })),
]

function getKasTypeClass(kasType: KasType): string {
  switch (kasType) {
    case 'masuk': return styles.badgeMasuk
    case 'keluar': return styles.badgeKeluar
    default: return styles.badgeDefault
  }
}

function getStatusClass(status: KasStatus): string {
  switch (status) {
    case 'pending': return styles.badgePending
    case 'posted': return styles.badgePosted
    case 'reversed': return styles.badgeReversed
    default: return styles.badgeDefault
  }
}

export default function KasPage() {
  const [search, setSearch] = useState('')
  const [kasType, setKasType] = useState<KasType | ''>('')
  const [status, setStatus] = useState<KasStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(kasType ? { kas_type: kasType } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.kas, filters],
    queryFn: () => kasService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => kasService.delete(id),
    onSuccess: () => {
      toast.success('Mutasi kas berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.kas] })
    },
    onError: () => {
      toast.error('Gagal menghapus mutasi kas')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || kasType !== '' || status !== ''

  const totalMasuk = items.filter((i) => i.kasType === 'masuk').reduce((s, i) => s + i.amount, 0)
  const totalKeluar = items.filter((i) => i.kasType === 'keluar').reduce((s, i) => s + i.amount, 0)
  const netto = totalMasuk - totalKeluar

  function handleReset() {
    setSearch('')
    setKasType('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Kas) {
    if (!confirm(`Hapus mutasi kas "${item.referenceNo || item.id}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Kas</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Mutasi
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Kas Masuk</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{formatRupiah(totalMasuk)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Kas Keluar</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueRed}`}>{formatRupiah(totalKeluar)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Netto</div>
          <div className={`${styles.summaryValue} ${netto >= 0 ? styles.summaryValueGreen : styles.summaryValueRed}`}>
            {formatRupiah(netto)}
          </div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari referensi atau deskripsi..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={kasType}
          onChange={(e) => { setKasType(e.target.value as KasType | ''); setPage(1) }}
          aria-label="Filter tipe kas"
        >
          {KAS_TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as KasStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Tipe</th>
                <th className={styles.th}>Jumlah</th>
                <th className={styles.th}>Deskripsi</th>
                <th className={styles.th}>Referensi</th>
                <th className={styles.th}>Tanggal Posting</th>
                <th className={styles.th}>Tanggal Sesi</th>
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
                      <p className={styles.emptySubtitle}>Belum ada mutasi kas yang tercatat.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getKasTypeClass(item.kasType)}`}>
                        {KAS_TYPE_LABELS[item.kasType] ?? item.kasType}
                      </span>
                    </td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.amount)}</span></td>
                    <td className={styles.td}>{item.description || '-'}</td>
                    <td className={styles.td}><span className={styles.code}>{item.referenceNo || '-'}</span></td>
                    <td className={styles.td}>{formatDate(item.postedDate)}</td>
                    <td className={styles.td}>{item.sessionDate || '-'}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {KAS_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus mutasi kas ${item.referenceNo}`}
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
