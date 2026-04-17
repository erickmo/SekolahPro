import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { moneyDenominationService } from '@/services/koperasi-ops.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type { MoneyDenomination, MoneyDenominationType } from '@/types/koperasi-ops.types'
import { MONEY_DENOMINATION_TYPE_LABELS } from '@/types/koperasi-ops.types'
import styles from './MoneyDenominationPage.module.css'

const PAGE_SIZE = 20

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

const TYPE_OPTIONS: { value: MoneyDenominationType | ''; label: string }[] = [
  { value: '', label: 'Semua Tipe' },
  { value: 'opening', label: 'Pembukaan' },
  { value: 'closing', label: 'Penutupan' },
]

function getTypeClass(type: MoneyDenominationType): string {
  switch (type) {
    case 'opening': return styles.badgeOpening
    case 'closing': return styles.badgeClosing
    default: return styles.badgeDefault
  }
}

export default function MoneyDenominationPage() {
  const [search, setSearch] = useState('')
  const [type, setType] = useState<MoneyDenominationType | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(type ? { type } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.moneyDenomination, filters],
    queryFn: () => moneyDenominationService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => moneyDenominationService.delete(id),
    onSuccess: () => {
      toast.success('Pecahan uang berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.moneyDenomination] })
    },
    onError: () => {
      toast.error('Gagal menghapus pecahan uang')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || type !== ''

  const totalOpening = items.filter((i) => i.type === 'opening').reduce((s, i) => s + i.total, 0)
  const totalClosing = items.filter((i) => i.type === 'closing').reduce((s, i) => s + i.total, 0)
  const totalCount = items.reduce((s, i) => s + i.count, 0)

  function handleReset() {
    setSearch('')
    setType('')
    setPage(1)
  }

  function handleDelete(item: MoneyDenomination) {
    if (!confirm(`Hapus pecahan ${formatRupiah(item.denomination)}?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Pecahan Uang</h1>
          {!isLoading && (
            <p className={styles.totalInfo}>
              {total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Pecahan
        </button>
      </div>

      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Pembukaan</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{formatRupiah(totalOpening)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Penutupan</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{formatRupiah(totalClosing)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Lembar</div>
          <div className={styles.summaryValue}>{totalCount}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari pecahan atau sesi..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={type}
          onChange={(e) => { setType(e.target.value as MoneyDenominationType | ''); setPage(1) }}
          aria-label="Filter tipe"
        >
          {TYPE_OPTIONS.map((opt) => (
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
                <th className={styles.th}>Nominal</th>
                <th className={styles.th}>Jumlah</th>
                <th className={styles.th}>Total</th>
                <th className={styles.th}>Tipe</th>
                <th className={styles.th}>Tanggal Sesi</th>
                <th className={styles.th}>Status Sesi</th>
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
                      <p className={styles.emptySubtitle}>Belum ada pecahan uang yang tercatat.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.denomination)}</span></td>
                    <td className={styles.td}>{item.count} lembar</td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.total)}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getTypeClass(item.type)}`}>
                        {MONEY_DENOMINATION_TYPE_LABELS[item.type] ?? item.type}
                      </span>
                    </td>
                    <td className={styles.td}>{item.sessionDate || '-'}</td>
                    <td className={styles.td}>{item.sessionStatus || '-'}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus pecahan ${item.denomination}`}
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
