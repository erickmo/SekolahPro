import { useState } from 'react'
import { Search, RotateCcw, Plus, Trash2, Eye } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDebounce } from '@/hooks/useDebounce'
import { createVernonService } from '@/services/vernon.service'
import { QK } from '@/services/query-keys'
import { toast } from '@/widgets/Toast/Toast'
import styles from './ZakatInfaqPage.module.css'

const PAGE_SIZE = 20

// ── Types ──────────────────────────────────────────────────────────────────────

type ZakatType = 'zakat_mal' | 'zakat_fitrah' | 'zakat_institusi'
type PaymentMethod = 'cash' | 'auto_debit' | 'transfer'
type ZakatCollectionStatus = 'pending' | 'confirmed' | 'cancelled'

type SourceFund = 'zakat' | 'infaq' | 'tazir'
type DistributionStatus = 'draft' | 'pending' | 'approved' | 'rejected' | 'completed'
type DistributionMethod = 'cash' | 'goods' | 'account_transfer'
type AsnafCategory = 'fakir' | 'miskin' | 'amil' | 'muallaf' | 'riqab' | 'gharimin' | 'fisabilillah' | 'ibnu_sabil'

type MustahikStatus = 'active' | 'inactive' | 'graduated'

type TabKey = 'zakat_collection' | 'zakat_distribution' | 'infaq' | 'mustahik' | 'tazir_fund'

interface ZakatCollection {
  id: string
  muzakkiName: string
  zakatType: ZakatType
  amount: number
  paymentMethod: PaymentMethod
  status: ZakatCollectionStatus
  nasabahName: string
  branchName: string
  collectionDate: string
  fitrahHeadCount: number | null
  notes: string
}

interface ZakatDistribution {
  id: string
  mustahikName: string
  primaryCategory: string
  sourceFund: SourceFund
  amount: number
  purpose: string
  distributionMethod: DistributionMethod
  status: DistributionStatus
  branchName: string
  asnafCategory: AsnafCategory | null
}

interface InfaqRecord {
  id: string
  donaturName: string
  infaqType: 'one_time' | 'recurring'
  amount: number
  designation: string
  paymentMethod: PaymentMethod
  nasabahName: string
  branchName: string
  infaqDate: string
}

interface Mustahik {
  id: string
  fullName: string
  asnafCategories: string[]
  primaryCategory: AsnafCategory
  needsAssessment: string
  assessmentDate: string
  status: MustahikStatus
  branchName: string
  phone: string
  nextReviewDate: string
}

interface TazirFund {
  id: string
  dendaId: string
  pinjamanId: string
  nasabahName: string
  amount: number
  status: 'collected' | 'distributed'
  collectionDate: string
  branchName: string
}

// ── Constants ──────────────────────────────────────────────────────────────────

const TABS: { key: TabKey; label: string }[] = [
  { key: 'zakat_collection', label: 'Koleksi Zakat' },
  { key: 'zakat_distribution', label: 'Distribusi Zakat' },
  { key: 'infaq', label: 'Infaq' },
  { key: 'mustahik', label: 'Mustahik' },
  { key: 'tazir_fund', label: 'Dana Ta\'zir' },
]

const ZAKAT_TYPE_LABELS: Record<ZakatType, string> = {
  zakat_mal: 'Zakat Mal',
  zakat_fitrah: 'Zakat Fitrah',
  zakat_institusi: 'Zakat Institusi',
}

const COLLECTION_STATUS_LABELS: Record<ZakatCollectionStatus, string> = {
  pending: 'Pending',
  confirmed: 'Dikonfirmasi',
  cancelled: 'Dibatalkan',
}

const SOURCE_FUND_LABELS: Record<SourceFund, string> = {
  zakat: 'Zakat',
  infaq: 'Infaq',
  tazir: "Ta'zir",
}

const DISTRIBUTION_STATUS_LABELS: Record<DistributionStatus, string> = {
  draft: 'Draft',
  pending: 'Menunggu',
  approved: 'Disetujui',
  rejected: 'Ditolak',
  completed: 'Selesai',
}

const ASNAF_LABELS: Record<AsnafCategory, string> = {
  fakir: 'Fakir',
  miskin: 'Miskin',
  amil: 'Amil',
  muallaf: 'Muallaf',
  riqab: 'Riqab',
  gharimin: 'Gharimin',
  fisabilillah: 'Fisabilillah',
  ibnu_sabil: 'Ibnu Sabil',
}

const MUSTAHIK_STATUS_LABELS: Record<MustahikStatus, string> = {
  active: 'Aktif',
  inactive: 'Nonaktif',
  graduated: 'Mandiri',
}

const ZAKAT_TYPE_OPTIONS: { value: ZakatType | ''; label: string }[] = [
  { value: '', label: 'Semua Jenis' },
  { value: 'zakat_mal', label: 'Zakat Mal' },
  { value: 'zakat_fitrah', label: 'Zakat Fitrah' },
  { value: 'zakat_institusi', label: 'Zakat Institusi' },
]

const COLLECTION_STATUS_OPTIONS: { value: ZakatCollectionStatus | ''; label: string }[] = [
  { value: '', label: 'Semua Status' },
  { value: 'pending', label: 'Pending' },
  { value: 'confirmed', label: 'Dikonfirmasi' },
  { value: 'cancelled', label: 'Dibatalkan' },
]

// ── Services ───────────────────────────────────────────────────────────────────

function transformZC(raw: Record<string, unknown>): ZakatCollection {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const nasabahData = (data.nasabah ?? {}) as Record<string, unknown>
  const branchData = (data.branch ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    muzakkiName: (raw.muzakki_name ?? '') as string,
    zakatType: (raw.zakat_type ?? 'zakat_mal') as ZakatType,
    amount: Number(raw.amount ?? 0),
    paymentMethod: (raw.payment_method ?? 'cash') as PaymentMethod,
    status: (raw.status ?? 'pending') as ZakatCollectionStatus,
    nasabahName: (nasabahData.nama_lengkap ?? '-') as string,
    branchName: (branchData.name ?? '-') as string,
    collectionDate: (raw.collection_date ?? '') as string,
    fitrahHeadCount: (raw.fitrah_head_count ?? null) as number | null,
    notes: (raw.notes ?? '') as string,
  }
}

function transformZD(raw: Record<string, unknown>): ZakatDistribution {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const mustahikData = (data.mustahik ?? {}) as Record<string, unknown>
  const branchData = (data.branch ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    mustahikName: (mustahikData.full_name ?? '-') as string,
    primaryCategory: (mustahikData.primary_category ?? '') as string,
    sourceFund: (raw.source_fund ?? 'zakat') as SourceFund,
    amount: Number(raw.amount ?? 0),
    purpose: (raw.purpose ?? '') as string,
    distributionMethod: (raw.distribution_method ?? 'cash') as DistributionMethod,
    status: (raw.status ?? 'draft') as DistributionStatus,
    branchName: (branchData.name ?? '-') as string,
    asnafCategory: (raw.asnaf_category ?? null) as AsnafCategory | null,
  }
}

function transformInfaq(raw: Record<string, unknown>): InfaqRecord {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const nasabahData = (data.nasabah ?? {}) as Record<string, unknown>
  const branchData = (data.branch ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    donaturName: (raw.donatur_name ?? '') as string,
    infaqType: (raw.infaq_type ?? 'one_time') as 'one_time' | 'recurring',
    amount: Number(raw.amount ?? 0),
    designation: (raw.designation ?? '') as string,
    paymentMethod: (raw.payment_method ?? 'cash') as PaymentMethod,
    nasabahName: (nasabahData.nama_lengkap ?? '-') as string,
    branchName: (branchData.name ?? '-') as string,
    infaqDate: (raw.infaq_date ?? '') as string,
  }
}

function transformMustahik(raw: Record<string, unknown>): Mustahik {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const branchData = (data.branch ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    fullName: (raw.full_name ?? '') as string,
    asnafCategories: (raw.asnaf_categories ?? []) as string[],
    primaryCategory: (raw.primary_category ?? 'fakir') as AsnafCategory,
    needsAssessment: (raw.needs_assessment ?? '') as string,
    assessmentDate: (raw.assessment_date ?? '') as string,
    status: (raw.status ?? 'active') as MustahikStatus,
    branchName: (branchData.name ?? '-') as string,
    phone: (raw.phone ?? '') as string,
    nextReviewDate: (raw.next_review_date ?? '') as string,
  }
}

function transformTF(raw: Record<string, unknown>): TazirFund {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const nasabahData = (data.nasabah ?? {}) as Record<string, unknown>
  const branchData = (data.branch ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    dendaId: (raw.denda_id ?? '') as string,
    pinjamanId: (raw.pinjaman_id ?? '') as string,
    nasabahName: (nasabahData.nama_lengkap ?? '-') as string,
    amount: Number(raw.amount ?? 0),
    status: (raw.status ?? 'collected') as 'collected' | 'distributed',
    collectionDate: (raw.collection_date ?? '') as string,
    branchName: (branchData.name ?? '-') as string,
  }
}

const zakatCollectionService = createVernonService<ZakatCollection, Record<string, unknown>>(
  '/zakat_collection',
  transformZC,
)

const zakatDistributionService = createVernonService<ZakatDistribution, Record<string, unknown>>(
  '/zakat_distribution',
  transformZD,
)

const infaqService = createVernonService<InfaqRecord, Record<string, unknown>>(
  '/infaq',
  transformInfaq,
)

const mustahikService = createVernonService<Mustahik, Record<string, unknown>>(
  '/mustahik',
  transformMustahik,
)

const tazirFundService = createVernonService<TazirFund, Record<string, unknown>>(
  '/tazir_fund',
  transformTF,
)

// ── Helpers ────────────────────────────────────────────────────────────────────

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency', currency: 'IDR',
    minimumFractionDigits: 0, maximumFractionDigits: 0,
  }).format(amount)
}

function getZCStatusClass(status: ZakatCollectionStatus): string {
  switch (status) {
    case 'pending': return styles.badgePending
    case 'confirmed': return styles.badgeConfirmed
    case 'cancelled': return styles.badgeCancelled
    default: return styles.badgeDefault
  }
}

function getZDStatusClass(status: DistributionStatus): string {
  switch (status) {
    case 'draft': return styles.badgeDraft
    case 'pending': return styles.badgePending
    case 'approved': return styles.badgeApproved
    case 'rejected': return styles.badgeRejected
    case 'completed': return styles.badgeCompleted
    default: return styles.badgeDefault
  }
}

function getMustahikStatusClass(status: MustahikStatus): string {
  switch (status) {
    case 'active': return styles.badgeConfirmed
    case 'inactive': return styles.badgeCancelled
    case 'graduated': return styles.badgeApproved
    default: return styles.badgeDefault
  }
}

// ── Component ──────────────────────────────────────────────────────────────────

export default function ZakatInfaqPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('zakat_collection')
  const [search, setSearch] = useState('')
  const [zakatType, setZakatType] = useState<ZakatType | ''>('')
  const [zcStatus, setZcStatus] = useState<ZakatCollectionStatus | ''>('')
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const debouncedSearch = useDebounce(search, 300)

  // Reset filters on tab change
  function switchTab(tab: TabKey) {
    setActiveTab(tab)
    setSearch('')
    setZakatType('')
    setZcStatus('')
    setPage(1)
  }

  // ── Zakat Collection ────────────────────────────────────────────────────────

  const zcFilters: Record<string, unknown> = {
    _page: page, _limit: PAGE_SIZE,
    ...(debouncedSearch ? { muzakki_name: debouncedSearch } : {}),
    ...(zakatType ? { zakat_type: zakatType } : {}),
    ...(zcStatus ? { status: zcStatus } : {}),
  }

  const { data: zcData, isLoading: zcLoading } = useQuery({
    queryKey: [QK.zakatCollection, zcFilters],
    queryFn: () => zakatCollectionService.list(zcFilters),
    enabled: activeTab === 'zakat_collection',
  })

  const zcDelete = useMutation({
    mutationFn: (id: string) => zakatCollectionService.delete(id),
    onSuccess: () => {
      toast.success('Data zakat berhasil dihapus')
      queryClient.invalidateQueries({ queryKey: [QK.zakatCollection] })
    },
    onError: () => toast.error('Gagal menghapus data zakat'),
  })

  // ── Zakat Distribution ──────────────────────────────────────────────────────

  const zdFilters: Record<string, unknown> = {
    _page: page, _limit: PAGE_SIZE,
    ...(debouncedSearch ? { purpose: debouncedSearch } : {}),
  }

  const { data: zdData, isLoading: zdLoading } = useQuery({
    queryKey: [QK.zakatDistribution, zdFilters],
    queryFn: () => zakatDistributionService.list(zdFilters),
    enabled: activeTab === 'zakat_distribution',
  })

  // ── Infaq ───────────────────────────────────────────────────────────────────

  const infaqFilters: Record<string, unknown> = {
    _page: page, _limit: PAGE_SIZE,
    ...(debouncedSearch ? { donatur_name: debouncedSearch } : {}),
  }

  const { data: infaqData, isLoading: infaqLoading } = useQuery({
    queryKey: [QK.infaq, infaqFilters],
    queryFn: () => infaqService.list(infaqFilters),
    enabled: activeTab === 'infaq',
  })

  // ── Mustahik ────────────────────────────────────────────────────────────────

  const mustahikFilters: Record<string, unknown> = {
    _page: page, _limit: PAGE_SIZE,
    ...(debouncedSearch ? { full_name: debouncedSearch } : {}),
  }

  const { data: mustahikData, isLoading: mustahikLoading } = useQuery({
    queryKey: [QK.mustahik, mustahikFilters],
    queryFn: () => mustahikService.list(mustahikFilters),
    enabled: activeTab === 'mustahik',
  })

  // ── Tazir Fund ──────────────────────────────────────────────────────────────

  const tfFilters: Record<string, unknown> = {
    _page: page, _limit: PAGE_SIZE,
  }

  const { data: tfData, isLoading: tfLoading } = useQuery({
    queryKey: [QK.tazirFund, tfFilters],
    queryFn: () => tazirFundService.list(tfFilters),
    enabled: activeTab === 'tazir_fund',
  })

  // ── Current data ────────────────────────────────────────────────────────────

  const currentData = (() => {
    switch (activeTab) {
      case 'zakat_collection': return { items: zcData?.items ?? [], total: zcData?.total ?? 0, totalPages: zcData?.totalPages ?? 1, loading: zcLoading }
      case 'zakat_distribution': return { items: zdData?.items ?? [], total: zdData?.total ?? 0, totalPages: zdData?.totalPages ?? 1, loading: zdLoading }
      case 'infaq': return { items: infaqData?.items ?? [], total: infaqData?.total ?? 0, totalPages: infaqData?.totalPages ?? 1, loading: infaqLoading }
      case 'mustahik': return { items: mustahikData?.items ?? [], total: mustahikData?.total ?? 0, totalPages: mustahikData?.totalPages ?? 1, loading: mustahikLoading }
      case 'tazir_fund': return { items: tfData?.items ?? [], total: tfData?.total ?? 0, totalPages: tfData?.totalPages ?? 1, loading: tfLoading }
    }
  })()

  const hasFilters = search !== '' || zakatType !== '' || zcStatus !== ''

  function handleReset() {
    setSearch('')
    setZakatType('')
    setZcStatus('')
    setPage(1)
  }

  // ── Render ──────────────────────────────────────────────────────────────────

  return (
    <div className={styles.root}>
      <div className={styles.topBar}>
        <div>
          <h1 className={styles.pageTitle}>Zakat & Infaq</h1>
          {!currentData.loading && (
            <p className={styles.totalInfo}>
              {currentData.total.toLocaleString('id-ID')} data ditemukan
            </p>
          )}
        </div>
        <button type="button" className={styles.addBtn}>
          <Plus size={16} /> Tambah Data
        </button>
      </div>

      <div className={styles.tabs}>
        {TABS.map((tab) => (
          <button
            key={tab.key}
            type="button"
            className={`${styles.tab} ${activeTab === tab.key ? styles.tabActive : ''}`}
            onClick={() => switchTab(tab.key)}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className={styles.filterRow}>
        <div className={styles.searchWrap}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder={
              activeTab === 'mustahik' ? 'Cari nama mustahik...'
                : activeTab === 'zakat_collection' ? 'Cari nama muzakki...'
                  : activeTab === 'infaq' ? 'Cari nama donatur...'
                    : 'Cari...'
            }
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        {activeTab === 'zakat_collection' && (
          <>
            <select
              className={styles.select}
              value={zakatType}
              onChange={(e) => { setZakatType(e.target.value as ZakatType | ''); setPage(1) }}
              aria-label="Filter jenis zakat"
            >
              {ZAKAT_TYPE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
            <select
              className={styles.select}
              value={zcStatus}
              onChange={(e) => { setZcStatus(e.target.value as ZakatCollectionStatus | ''); setPage(1) }}
              aria-label="Filter status"
            >
              {COLLECTION_STATUS_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </>
        )}
        {hasFilters && (
          <button type="button" className={styles.resetBtn} onClick={handleReset}>
            <RotateCcw size={14} /> Reset Filter
          </button>
        )}
      </div>

      <div className={styles.tableCard}>
        <div className={styles.tableScroll}>
          {activeTab === 'zakat_collection' && (
            <ZakatCollectionTable
              items={currentData.items as ZakatCollection[]}
              loading={currentData.loading}
              onDelete={(item) => {
                if (!confirm(`Hapus data zakat "${item.muzakkiName}"?`)) return
                zcDelete.mutate(item.id)
              }}
              deleting={zcDelete.isPending}
            />
          )}
          {activeTab === 'zakat_distribution' && (
            <ZakatDistributionTable
              items={currentData.items as ZakatDistribution[]}
              loading={currentData.loading}
            />
          )}
          {activeTab === 'infaq' && (
            <InfaqTable
              items={currentData.items as InfaqRecord[]}
              loading={currentData.loading}
            />
          )}
          {activeTab === 'mustahik' && (
            <MustahikTable
              items={currentData.items as Mustahik[]}
              loading={currentData.loading}
            />
          )}
          {activeTab === 'tazir_fund' && (
            <TazirFundTable
              items={currentData.items as TazirFund[]}
              loading={currentData.loading}
            />
          )}
        </div>

        {!currentData.loading && currentData.totalPages > 1 && (
          <div className={styles.pagination}>
            <button type="button" className={styles.pageBtn} disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              Sebelumnya
            </button>
            <span className={styles.pageInfo}>Halaman {page} dari {currentData.totalPages}</span>
            <button type="button" className={styles.pageBtn} disabled={page >= currentData.totalPages} onClick={() => setPage((p) => p + 1)}>
              Berikutnya
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

// ── Sub-tables ─────────────────────────────────────────────────────────────────

function ZakatCollectionTable({ items, loading, onDelete, deleting }: {
  items: ZakatCollection[]
  loading: boolean
  onDelete: (item: ZakatCollection) => void
  deleting: boolean
}) {
  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th className={styles.th}>Muzakki</th>
          <th className={styles.th}>Jenis Zakat</th>
          <th className={styles.th}>Jumlah</th>
          <th className={styles.th}>Metode Bayar</th>
          <th className={styles.th}>Cabang</th>
          <th className={styles.th}>Status</th>
          <th className={styles.th}>Aksi</th>
        </tr>
      </thead>
      <tbody>
        {loading ? (
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
                <p className={styles.emptySubtitle}>Belum ada koleksi zakat yang tercatat.</p>
              </div>
            </td>
          </tr>
        ) : (
          items.map((item) => (
            <tr key={item.id} className={styles.row}>
              <td className={styles.td}>
                <span className={styles.name}>{item.muzakkiName}</span>
                {item.nasabahName !== '-' && (
                  <span className={styles.subtext}>{item.nasabahName}</span>
                )}
              </td>
              <td className={styles.td}>
                <span className={`${styles.badge} ${styles.badgeZakatType}`}>
                  {ZAKAT_TYPE_LABELS[item.zakatType] ?? item.zakatType}
                </span>
              </td>
              <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.amount)}</span></td>
              <td className={styles.td}>{item.paymentMethod}</td>
              <td className={styles.td}>{item.branchName}</td>
              <td className={styles.td}>
                <span className={`${styles.badge} ${getZCStatusClass(item.status)}`}>
                  {COLLECTION_STATUS_LABELS[item.status] ?? item.status}
                </span>
              </td>
              <td className={styles.td}>
                <button
                  type="button"
                  className={styles.deleteBtn}
                  onClick={() => onDelete(item)}
                  disabled={deleting}
                  aria-label={`Hapus zakat ${item.muzakkiName}`}
                >
                  <Trash2 size={14} />
                </button>
              </td>
            </tr>
          ))
        )}
      </tbody>
    </table>
  )
}

function ZakatDistributionTable({ items, loading }: { items: ZakatDistribution[]; loading: boolean }) {
  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th className={styles.th}>Mustahik</th>
          <th className={styles.th}>Sumber Dana</th>
          <th className={styles.th}>Jumlah</th>
          <th className={styles.th}>Tujuan</th>
          <th className={styles.th}>Metode</th>
          <th className={styles.th}>Status</th>
          <th className={styles.th}>Aksi</th>
        </tr>
      </thead>
      <tbody>
        {loading ? (
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
                <p className={styles.emptySubtitle}>Belum ada distribusi zakat yang tercatat.</p>
              </div>
            </td>
          </tr>
        ) : (
          items.map((item) => (
            <tr key={item.id} className={styles.row}>
              <td className={styles.td}>
                <span className={styles.name}>{item.mustahikName}</span>
                {item.asnafCategory && (
                  <span className={styles.subtext}>{ASNAF_LABELS[item.asnafCategory] ?? item.asnafCategory}</span>
                )}
              </td>
              <td className={styles.td}>{SOURCE_FUND_LABELS[item.sourceFund] ?? item.sourceFund}</td>
              <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.amount)}</span></td>
              <td className={styles.td}>{item.purpose}</td>
              <td className={styles.td}>{item.distributionMethod}</td>
              <td className={styles.td}>
                <span className={`${styles.badge} ${getZDStatusClass(item.status)}`}>
                  {DISTRIBUTION_STATUS_LABELS[item.status] ?? item.status}
                </span>
              </td>
              <td className={styles.td}>
                <button type="button" className={styles.viewBtn} aria-label="Lihat detail">
                  <Eye size={14} />
                </button>
              </td>
            </tr>
          ))
        )}
      </tbody>
    </table>
  )
}

function InfaqTable({ items, loading }: { items: InfaqRecord[]; loading: boolean }) {
  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th className={styles.th}>Donatur</th>
          <th className={styles.th}>Jenis</th>
          <th className={styles.th}>Jumlah</th>
          <th className={styles.th}>Peruntukan</th>
          <th className={styles.th}>Metode Bayar</th>
          <th className={styles.th}>Cabang</th>
        </tr>
      </thead>
      <tbody>
        {loading ? (
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
                <p className={styles.emptySubtitle}>Belum ada infaq yang tercatat.</p>
              </div>
            </td>
          </tr>
        ) : (
          items.map((item) => (
            <tr key={item.id} className={styles.row}>
              <td className={styles.td}><span className={styles.name}>{item.donaturName}</span></td>
              <td className={styles.td}>
                <span className={`${styles.badge} ${item.infaqType === 'recurring' ? styles.badgeRecurring : styles.badgeOneTime}`}>
                  {item.infaqType === 'recurring' ? 'Berkala' : 'Sekali'}
                </span>
              </td>
              <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.amount)}</span></td>
              <td className={styles.td}>{item.designation}</td>
              <td className={styles.td}>{item.paymentMethod}</td>
              <td className={styles.td}>{item.branchName}</td>
            </tr>
          ))
        )}
      </tbody>
    </table>
  )
}

function MustahikTable({ items, loading }: { items: Mustahik[]; loading: boolean }) {
  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th className={styles.th}>Nama</th>
          <th className={styles.th}>Kategori Utama</th>
          <th className={styles.th}>Assessment</th>
          <th className={styles.th}>Review Berikutnya</th>
          <th className={styles.th}>Cabang</th>
          <th className={styles.th}>Status</th>
        </tr>
      </thead>
      <tbody>
        {loading ? (
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
                <p className={styles.emptySubtitle}>Belum ada mustahik yang terdaftar.</p>
              </div>
            </td>
          </tr>
        ) : (
          items.map((item) => (
            <tr key={item.id} className={styles.row}>
              <td className={styles.td}>
                <span className={styles.name}>{item.fullName}</span>
                {item.phone && <span className={styles.subtext}>{item.phone}</span>}
              </td>
              <td className={styles.td}>
                <span className={`${styles.badge} ${styles.badgeAsnaf}`}>
                  {ASNAF_LABELS[item.primaryCategory] ?? item.primaryCategory}
                </span>
              </td>
              <td className={styles.td}><span className={styles.assessment}>{item.needsAssessment}</span></td>
              <td className={styles.td}>{item.nextReviewDate || '-'}</td>
              <td className={styles.td}>{item.branchName}</td>
              <td className={styles.td}>
                <span className={`${styles.badge} ${getMustahikStatusClass(item.status)}`}>
                  {MUSTAHIK_STATUS_LABELS[item.status] ?? item.status}
                </span>
              </td>
            </tr>
          ))
        )}
      </tbody>
    </table>
  )
}

function TazirFundTable({ items, loading }: { items: TazirFund[]; loading: boolean }) {
  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th className={styles.th}>Nasabah</th>
          <th className={styles.th}>Jumlah Denda</th>
          <th className={styles.th}>Cabang</th>
          <th className={styles.th}>Tanggal</th>
          <th className={styles.th}>Status</th>
        </tr>
      </thead>
      <tbody>
        {loading ? (
          Array.from({ length: 5 }).map((_, i) => (
            <tr key={i} className={styles.skeletonRow}>
              {Array.from({ length: 5 }).map((_, j) => (
                <td key={j}><span className={styles.skeleton} style={{ width: '80px' }} /></td>
              ))}
            </tr>
          ))
        ) : items.length === 0 ? (
          <tr>
            <td colSpan={5} className={styles.emptyCell}>
              <div className={styles.emptyState}>
                <p className={styles.emptyTitle}>Tidak ada data</p>
                <p className={styles.emptySubtitle}>Belum ada dana ta'zir yang tercatat.</p>
              </div>
            </td>
          </tr>
        ) : (
          items.map((item) => (
            <tr key={item.id} className={styles.row}>
              <td className={styles.td}><span className={styles.name}>{item.nasabahName}</span></td>
              <td className={styles.td}><span className={styles.saldo}>{formatRupiah(item.amount)}</span></td>
              <td className={styles.td}>{item.branchName}</td>
              <td className={styles.td}>{item.collectionDate || '-'}</td>
              <td className={styles.td}>
                <span className={`${styles.badge} ${item.status === 'collected' ? styles.badgeConfirmed : styles.badgeDistributed}`}>
                  {item.status === 'collected' ? 'Terkumpul' : 'Disalurkan'}
                </span>
              </td>
            </tr>
          ))
        )}
      </tbody>
    </table>
  )
}
