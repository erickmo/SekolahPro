import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2, Lock, Check, Minus } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { depositoService } from '@/services/koperasi.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { Deposito, DepositoStatus } from '@/types/koperasi.types'
import { DEPOSITO_STATUS_LABELS } from '@/types/koperasi.types'
import styles from './DepositoPage.module.css'

const PAGE_SIZE = 20

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

const TENOR_OPTIONS: { value: number | ''; label: string }[] = [
  { value: '', label: 'Semua Tenor' },
  { value: 1, label: '1 Bulan' },
  { value: 3, label: '3 Bulan' },
  { value: 6, label: '6 Bulan' },
  { value: 12, label: '12 Bulan' },
  { value: 24, label: '24 Bulan' },
]

const STATUS_OPTIONS: { value: DepositoStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'active', label: 'Aktif' },
  { value: 'matured', label: 'Jatuh Tempo' },
  { value: 'rolled_over', label: 'Roll Over' },
  { value: 'early_withdrawn', label: 'Pencairan Awal' },
  { value: 'closed', label: 'Ditutup' },
]

function getStatusClass(status: DepositoStatus): string {
  switch (status) {
    case 'active': return styles.badgeActive
    case 'matured': return styles.badgeMatured
    case 'rolled_over': return styles.badgeRolledOver
    case 'early_withdrawn': return styles.badgeEarlyWithdrawn
    case 'closed': return styles.badgeClosed
    default: return styles.badgeDefault
  }
}

export default function DepositoPage() {
  const [search, setSearch] = useState('')
  const [tenor, setTenor] = useState<number | ''>('')
  const [status, setStatus] = useState<DepositoStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { nasabah_nama: debouncedSearch } : {}),
    ...(tenor !== '' ? { tenor } : {}),
    ...(status ? { status } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.deposito, filters],
    queryFn: () => depositoService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => depositoService.delete(id),
    onSuccess: () => {
      toast.success('Deposito berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.deposito] })
    },
    onError: () => {
      toast.error('Gagal menghapus deposito')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || tenor !== '' || status !== ''

  const totalDepositoAktif = items.filter((item) => item.status === 'active').length
  const totalJatuhTempoBulanIni = items.filter((item) => item.status === 'matured').length
  const totalNominal = items.reduce((sum, item) => sum + item.nominal, 0)

  function handleReset() {
    setSearch('')
    setTenor('')
    setStatus('')
    setPage(1)
  }

  function handleDelete(item: Deposito) {
    if (!confirm(`Hapus deposito "${item.noRekening}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Deposito</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Deposito
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Deposito Aktif</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{totalDepositoAktif}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Jatuh Tempo Bulan Ini</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueRed}`}>{totalJatuhTempoBulanIni}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Nominal</div>
          <div className={styles.summaryValue}>{formatRupiah(totalNominal)}</div>
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
          value={tenor}
          onChange={(e) => { setTenor(e.target.value === '' ? '' : Number(e.target.value)); setPage(1) }}
          aria-label="Filter tenor"
        >
          {TENOR_OPTIONS.map((opt) => (
            <option key={String(opt.value)} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as DepositoStatus | ''); setPage(1) }}
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
                <th className={styles.th}>Nominal</th>
                <th className={styles.th}>Tenor</th>
                <th className={styles.th}>Bunga (%)</th>
                <th className={styles.th}>Jatuh Tempo</th>
                <th className={styles.th}>Status</th>
                <th className={styles.th}>Auto Roll</th>
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
                      <p className={styles.emptySubtitle}>Belum ada deposito yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}>
                      <span className={styles.name}>
                        {item.onHold && <Lock size={12} className={styles.lockIcon} />}
                        {item.nasabahNama || '-'}
                      </span>
                    </td>
                    <td className={styles.td}><span className={styles.code}>{item.noRekening}</span></td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.nominal)}</span></td>
                    <td className={styles.td}>{item.tenor} bln</td>
                    <td className={styles.td}>{item.bunga}%</td>
                    <td className={styles.td}>{item.jatuhTempo || '-'}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getStatusClass(item.status)}`}>
                        {DEPOSITO_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      {item.autoRoll ? (
                        <Check size={16} className={styles.iconCheck} />
                      ) : (
                        <Minus size={16} className={styles.iconDash} />
                      )}
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus deposito ${item.noRekening}`}
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
