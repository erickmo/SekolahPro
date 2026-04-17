import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2, PieChart, Users } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { shuPeriodeService, shuAnggotaService } from '@/services/koperasi-ops.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type {
  SHUPeriode,
  SHUPeriodeStatus,
  SHUAnggota,
  SHUAnggotaDistMethod,
} from '@/types/koperasi-ops.types'
import {
  SHU_PERIODE_STATUS_LABELS,
  SHU_ANGGOTA_DIST_METHOD_LABELS,
} from '@/types/koperasi-ops.types'
import styles from './ShuPage.module.css'

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

type TabKey = 'periode' | 'anggota'

const SHU_STATUS_OPTIONS: { value: SHUPeriodeStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(SHU_PERIODE_STATUS_LABELS).map(([value, label]) => ({
    value: value as SHUPeriodeStatus,
    label,
  })),
]

const DIST_METHOD_OPTIONS: { value: SHUAnggotaDistMethod | ''; label: string }[] = [
  { value: '', label: 'Semua Metode' },
  ...Object.entries(SHU_ANGGOTA_DIST_METHOD_LABELS).map(([value, label]) => ({
    value: value as SHUAnggotaDistMethod,
    label,
  })),
]

function getSHUStatusClass(status: SHUPeriodeStatus): string {
  switch (status) {
    case 'calculated': return styles.badgePending
    case 'reviewed': return styles.badgeReviewed
    case 'approved': return styles.badgeApproved
    case 'distributed': return styles.badgePosted
    default: return styles.badgeDefault
  }
}

function getDistMethodClass(method: SHUAnggotaDistMethod): string {
  switch (method) {
    case 'credit_tabungan': return styles.badgePosted
    case 'separate_payout': return styles.badgeReviewed
    case 'pending': return styles.badgePending
    default: return styles.badgeDefault
  }
}

export default function ShuPage() {
  const [tab, setTab] = useState<TabKey>('periode')
  const queryClient = useQueryClient()

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>SHU (Sisa Hasil Usaha)</h1>
          <p className={styles.subtitle}>Perhitungan dan distribusi SHU anggota koperasi</p>
        </div>
      </div>

      <div className={styles.tabBar}>
        <button
          type="button"
          className={`${styles.tab} ${tab === 'periode' ? styles.tabActive : ''}`}
          onClick={() => setTab('periode')}
        >
          <PieChart size={15} /> Periode SHU
        </button>
        <button
          type="button"
          className={`${styles.tab} ${tab === 'anggota' ? styles.tabActive : ''}`}
          onClick={() => setTab('anggota')}
        >
          <Users size={15} /> SHU Anggota
        </button>
      </div>

      {tab === 'periode' ? (
        <PeriodeTab queryClient={queryClient} />
      ) : (
        <AnggotaTab queryClient={queryClient} />
      )}
    </div>
  )
}

// ── Periode Tab ──────────────────────────────────────────────────────────────

function PeriodeTab({ queryClient }: { queryClient: ReturnType<typeof useQueryClient> }) {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<SHUPeriodeStatus | ''>('')
  const [page, setPage] = useState(1)

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.shuPeriode, filters],
    queryFn: () => shuPeriodeService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => shuPeriodeService.delete(id),
    onSuccess: () => {
      toast.success('Periode SHU berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.shuPeriode] })
    },
    onError: () => {
      toast.error('Gagal menghapus periode SHU')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || status !== ''

  const totalSHUNeto = items.reduce((s, i) => s + i.shuNeto, 0)

  function handleReset() {
    setSearch('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: SHUPeriode) {
    if (item.status === 'distributed') {
      toast.error('SHU yang sudah didistribusikan tidak dapat dihapus')
      return
    }
    if (!confirm(`Hapus periode SHU tahun ${item.tahunBuku}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <>
      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total SHU Neto</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{formatRupiah(totalSHUNeto)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Periode</div>
          <div className={styles.summaryValue}>{total}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari tahun buku..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as SHUPeriodeStatus | ''); setPage(1) }}
          aria-label="Filter status SHU"
        >
          {SHU_STATUS_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Tahun Buku</th>
                <th className={styles.th}>Periode</th>
                <th className={styles.th}>Total Pendapatan</th>
                <th className={styles.th}>Total Beban</th>
                <th className={styles.th}>SHU Bruto</th>
                <th className={styles.th}>SHU Neto</th>
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
                      <p className={styles.emptySubtitle}>Belum ada periode SHU yang dihitung.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.tahunBuku}</span></td>
                    <td className={styles.td}>
                      {formatDate(item.periodStart)} — {formatDate(item.periodEnd)}
                    </td>
                    <td className={styles.td}><span className={styles.amount}>{formatRupiah(item.totalPendapatan)}</span></td>
                    <td className={styles.td}><span className={styles.amount}>{formatRupiah(item.totalBeban)}</span></td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.shuBruto)}</span></td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.shuNeto)}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getSHUStatusClass(item.status)}`}>
                        {SHU_PERIODE_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus periode SHU ${item.tahunBuku}`}
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
    </>
  )
}

// ── Anggota Tab ──────────────────────────────────────────────────────────────

function AnggotaTab({ queryClient }: { queryClient: ReturnType<typeof useQueryClient> }) {
  const [search, setSearch] = useState('')
  const [distMethod, setDistMethod] = useState<SHUAnggotaDistMethod | ''>('')
  const [page, setPage] = useState(1)

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(distMethod ? { distribution_method: distMethod } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.shuAnggota, filters],
    queryFn: () => shuAnggotaService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => shuAnggotaService.delete(id),
    onSuccess: () => {
      toast.success('Data SHU anggota berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.shuAnggota] })
    },
    onError: () => {
      toast.error('Gagal menghapus data SHU anggota')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || distMethod !== ''

  const totalJasaModal = items.reduce((s, i) => s + i.jasaModal, 0)
  const totalJasaUsaha = items.reduce((s, i) => s + i.jasaUsaha, 0)
  const totalSHU = items.reduce((s, i) => s + i.totalSHU, 0)

  function handleReset() {
    setSearch('')
    setDistMethod('')
    setPage(1)
  }

  function handleDelete(item: SHUAnggota) {
    if (!confirm(`Hapus data SHU anggota "${item.nasabahNama}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <>
      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Jasa Modal</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{formatRupiah(totalJasaModal)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Jasa Usaha</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{formatRupiah(totalJasaUsaha)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total SHU Anggota</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{formatRupiah(totalSHU)}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nama atau nomor anggota..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={distMethod}
          onChange={(e) => { setDistMethod(e.target.value as SHUAnggotaDistMethod | ''); setPage(1) }}
          aria-label="Filter metode distribusi"
        >
          {DIST_METHOD_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        {hasFilters && (
          <button type="button" className={styles.resetBtn} onClick={handleReset}>
            <RotateCcw size={14} /> Reset Filter
          </button>
        )}
      </div>

      {!isLoading && (
        <p className={styles.totalInfo}>{total.toLocaleString('id-ID')} anggota ditemukan</p>
      )}

      <div className={styles.tableCard}>
        <div className={styles.tableScroll}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.th}>No. Anggota</th>
                <th className={styles.th}>Nama</th>
                <th className={styles.th}>Rata-rata Simpanan</th>
                <th className={styles.th}>Total Transaksi</th>
                <th className={styles.th}>Hari Aktif</th>
                <th className={styles.th}>Jasa Modal</th>
                <th className={styles.th}>Jasa Usaha</th>
                <th className={styles.th}>Total SHU</th>
                <th className={styles.th}>Metode</th>
                <th className={styles.th}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className={styles.skeletonRow}>
                    {Array.from({ length: 10 }).map((_, j) => (
                      <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
                    ))}
                  </tr>
                ))
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={10} className={styles.emptyCell}>
                    <div className={styles.emptyState}>
                      <p className={styles.emptyTitle}>Tidak ada data</p>
                      <p className={styles.emptySubtitle}>Belum ada data SHU anggota.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.nasabahNo || '-'}</span></td>
                    <td className={styles.td}>{item.nasabahNama || '-'}</td>
                    <td className={styles.td}><span className={styles.amount}>{formatRupiah(item.avgSimpanan)}</span></td>
                    <td className={styles.td}><span className={styles.amount}>{formatRupiah(item.totalTransaksi)}</span></td>
                    <td className={styles.td}>{item.activeDays}</td>
                    <td className={styles.td}><span className={styles.amount}>{formatRupiah(item.jasaModal)}</span></td>
                    <td className={styles.td}><span className={styles.amount}>{formatRupiah(item.jasaUsaha)}</span></td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.totalSHU)}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getDistMethodClass(item.distributionMethod)}`}>
                        {SHU_ANGGOTA_DIST_METHOD_LABELS[item.distributionMethod] ?? item.distributionMethod}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus SHU anggota ${item.nasabahNama}`}
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
    </>
  )
}
