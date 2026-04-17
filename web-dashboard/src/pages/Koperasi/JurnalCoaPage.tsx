import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2, BookOpen, FileText } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { coaService, jurnalService } from '@/services/koperasi-ops.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import type {
  COA,
  COAAccountType,
  Jurnal,
  JurnalSourceType,
  JurnalStatus,
} from '@/types/koperasi-ops.types'
import {
  COA_ACCOUNT_TYPE_LABELS,
  JURNAL_SOURCE_TYPE_LABELS,
  JURNAL_STATUS_LABELS,
} from '@/types/koperasi-ops.types'
import styles from './JurnalCoaPage.module.css'

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

type TabKey = 'coa' | 'jurnal'

const COA_TYPE_OPTIONS: { value: COAAccountType | ''; label: string }[] = [
  { value: '', label: 'Semua Tipe Akun' },
  ...Object.entries(COA_ACCOUNT_TYPE_LABELS).map(([value, label]) => ({
    value: value as COAAccountType,
    label,
  })),
]

const JURNAL_STATUS_OPTIONS: { value: JurnalStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  ...Object.entries(JURNAL_STATUS_LABELS).map(([value, label]) => ({
    value: value as JurnalStatus,
    label,
  })),
]

const SOURCE_TYPE_OPTIONS: { value: JurnalSourceType | ''; label: string }[] = [
  { value: '', label: 'Semua Sumber' },
  ...Object.entries(JURNAL_SOURCE_TYPE_LABELS).map(([value, label]) => ({
    value: value as JurnalSourceType,
    label,
  })),
]

function getCOATypeClass(type: COAAccountType): string {
  switch (type) {
    case 'asset': return styles.badgeAsset
    case 'liability': return styles.badgeLiability
    case 'equity': return styles.badgeEquity
    case 'revenue': return styles.badgeRevenue
    case 'expense': return styles.badgeExpense
    default: return styles.badgeDefault
  }
}

function getJurnalStatusClass(status: JurnalStatus): string {
  switch (status) {
    case 'unposted': return styles.badgePending
    case 'posted': return styles.badgePosted
    case 'reversed': return styles.badgeReversed
    default: return styles.badgeDefault
  }
}

export default function JurnalCoaPage() {
  const [tab, setTab] = useState<TabKey>('coa')
  const queryClient = useQueryClient()

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Jurnal & Akuntansi</h1>
          <p className={styles.subtitle}>Chart of Accounts dan Jurnal Akuntansi Koperasi</p>
        </div>
      </div>

      <div className={styles.tabBar}>
        <button
          type="button"
          className={`${styles.tab} ${tab === 'coa' ? styles.tabActive : ''}`}
          onClick={() => setTab('coa')}
        >
          <BookOpen size={15} /> Chart of Accounts
        </button>
        <button
          type="button"
          className={`${styles.tab} ${tab === 'jurnal' ? styles.tabActive : ''}`}
          onClick={() => setTab('jurnal')}
        >
          <FileText size={15} /> Jurnal
        </button>
      </div>

      {tab === 'coa' ? (
        <COATab queryClient={queryClient} />
      ) : (
        <JurnalTab queryClient={queryClient} />
      )}
    </div>
  )
}

// ── COA Tab ──────────────────────────────────────────────────────────────────

function COATab({ queryClient }: { queryClient: ReturnType<typeof useQueryClient> }) {
  const [search, setSearch] = useState('')
  const [accountType, setAccountType] = useState<COAAccountType | ''>('')
  const [page, setPage] = useState(1)

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(accountType ? { account_type: accountType } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.coa, filters],
    queryFn: () => coaService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => coaService.delete(id),
    onSuccess: () => {
      toast.success('Akun berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.coa] })
    },
    onError: () => {
      toast.error('Gagal menghapus akun')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || accountType !== ''

  function handleReset() {
    setSearch('')
    setAccountType('')
    setPage(1)
  }

  function handleDelete(item: COA) {
    if (item.isSystem) {
      toast.error('Akun sistem tidak dapat dihapus')
      return
    }
    if (!confirm(`Hapus akun "${item.accountCode} - ${item.accountName}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <>
      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari kode atau nama akun..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={accountType}
          onChange={(e) => { setAccountType(e.target.value as COAAccountType | ''); setPage(1) }}
          aria-label="Filter tipe akun"
        >
          {COA_TYPE_OPTIONS.map((opt) => (
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
        <p className={styles.totalInfo}>{total.toLocaleString('id-ID')} akun ditemukan</p>
      )}

      <div className={styles.tableCard}>
        <div className={styles.tableScroll}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.th}>Kode</th>
                <th className={styles.th}>Nama Akun</th>
                <th className={styles.th}>Level</th>
                <th className={styles.th}>Tipe</th>
                <th className={styles.th}>Saldo Normal</th>
                <th className={styles.th}>Mode Koperasi</th>
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
                      <p className={styles.emptySubtitle}>Belum ada akun yang terdaftar.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.accountCode}</span></td>
                    <td className={styles.td}>{item.accountName}</td>
                    <td className={styles.td}>{item.level}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getCOATypeClass(item.accountType)}`}>
                        {COA_ACCOUNT_TYPE_LABELS[item.accountType] ?? item.accountType}
                      </span>
                    </td>
                    <td className={styles.td}>{item.normalBalance === 'debit' ? 'Debit' : 'Kredit'}</td>
                    <td className={styles.td}>
                      {item.coopTypeRequired === 'both' ? 'Semua' : item.coopTypeRequired === 'islamic' ? 'Syariah' : 'Konvensional'}
                    </td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${item.isActive ? styles.badgePosted : styles.badgeReversed}`}>
                        {item.isActive ? 'Aktif' : 'Nonaktif'}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending || item.isSystem}
                        aria-label={`Hapus akun ${item.accountCode}`}
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

// ── Jurnal Tab ───────────────────────────────────────────────────────────────

function JurnalTab({ queryClient }: { queryClient: ReturnType<typeof useQueryClient> }) {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<JurnalStatus | ''>('')
  const [sourceType, setSourceType] = useState<JurnalSourceType | ''>('')
  const [page, setPage] = useState(1)

  const debouncedSearch = useDebounce(search, 300)

  const filters: Record<string, unknown> = {
    _page: page,
    _limit: PAGE_SIZE,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
    ...(status ? { status } : {}),
    ...(sourceType ? { source_type: sourceType } : {}),
  }

  const { data, isLoading } = useQuery({
    queryKey: [QK.jurnal, filters],
    queryFn: () => jurnalService.list(filters),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => jurnalService.delete(id),
    onSuccess: () => {
      toast.success('Jurnal berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.jurnal] })
    },
    onError: () => {
      toast.error('Gagal menghapus jurnal')
    },
  })

  const items = data?.items ?? []
  const totalPages = data?.totalPages ?? 1
  const total = data?.total ?? 0
  const hasFilters = search !== '' || status !== '' || sourceType !== ''

  const totalDebit = items.reduce((s, i) => s + i.totalDebit, 0)
  const totalCredit = items.reduce((s, i) => s + i.totalCredit, 0)

  function handleReset() {
    setSearch('')
    setStatus('')
    setSourceType('')
    setPage(1)
  }

  function handleDelete(item: Jurnal) {
    if (item.status === 'posted') {
      toast.error('Jurnal yang sudah di-posting tidak dapat dihapus')
      return
    }
    if (!confirm(`Hapus jurnal "${item.journalNumber}"?`)) return
    deleteMutation.mutate(item.id)
  }

  return (
    <>
      <div className={styles.summaryRow}>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Debit</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueBlue}`}>{formatRupiah(totalDebit)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Kredit</div>
          <div className={`${styles.summaryValue} ${styles.summaryValueGreen}`}>{formatRupiah(totalCredit)}</div>
        </div>
        <div className={styles.summaryCard}>
          <div className={styles.summaryLabel}>Total Jurnal</div>
          <div className={styles.summaryValue}>{total}</div>
        </div>
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Cari nomor jurnal atau deskripsi..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <select
          className={styles.select}
          value={sourceType}
          onChange={(e) => { setSourceType(e.target.value as JurnalSourceType | ''); setPage(1) }}
          aria-label="Filter sumber jurnal"
        >
          {SOURCE_TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          className={styles.select}
          value={status}
          onChange={(e) => { setStatus(e.target.value as JurnalStatus | ''); setPage(1) }}
          aria-label="Filter status jurnal"
        >
          {JURNAL_STATUS_OPTIONS.map((opt) => (
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
                <th className={styles.th}>No. Jurnal</th>
                <th className={styles.th}>Tanggal</th>
                <th className={styles.th}>Deskripsi</th>
                <th className={styles.th}>Sumber</th>
                <th className={styles.th}>Periode</th>
                <th className={styles.th}>Debit</th>
                <th className={styles.th}>Kredit</th>
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
                      <p className={styles.emptySubtitle}>Belum ada jurnal yang tercatat.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className={styles.row}>
                    <td className={styles.td}><span className={styles.code}>{item.journalNumber}</span></td>
                    <td className={styles.td}>{formatDate(item.journalDate)}</td>
                    <td className={styles.td}>{item.description}</td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${styles.badgeDefault}`}>
                        {JURNAL_SOURCE_TYPE_LABELS[item.sourceType] ?? item.sourceType}
                      </span>
                    </td>
                    <td className={styles.td}>{item.periodName || '-'}</td>
                    <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.totalDebit)}</span></td>
                    <td className={styles.td}><span className={styles.amount}>{formatRupiah(item.totalCredit)}</span></td>
                    <td className={styles.td}>
                      <span className={`${styles.badge} ${getJurnalStatusClass(item.status)}`}>
                        {JURNAL_STATUS_LABELS[item.status] ?? item.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.deleteBtn}
                        onClick={() => handleDelete(item)}
                        disabled={deleteMutation.isPending}
                        aria-label={`Hapus jurnal ${item.journalNumber}`}
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
