import type { BaseEntity } from './entity.types'

// ── Transaksi ────────────────────────────────────────────────────────────────

export type TransaksiTransactionType = 'credit' | 'debit' | 'transfer'
export type TransaksiStatus = 'pending' | 'posted' | 'reversed'

export interface Transaksi extends BaseEntity {
  rekeningId: string
  rekeningNo: string
  transactionType: TransaksiTransactionType
  amount: number
  description: string
  referenceNo: string
  balanceBefore: number
  balanceAfter: number
  postedDate: string
  status: TransaksiStatus
}

export const TRANSAKSI_TRANSACTION_TYPE_LABELS: Record<TransaksiTransactionType, string> = {
  credit: 'Kredit',
  debit: 'Debit',
  transfer: 'Transfer',
}

export const TRANSAKSI_STATUS_LABELS: Record<TransaksiStatus, string> = {
  pending: 'Pending',
  posted: 'Posted',
  reversed: 'Reversed',
}

// ── TellerSession ────────────────────────────────────────────────────────────

export type TellerSessionStatus = 'open' | 'closed' | 'reconciled'

export interface TellerSession extends BaseEntity {
  userId: string
  sessionDate: string
  startTime: string
  endTime: string
  openingBalance: number
  closingBalance: number
  cashTotal: number
  transactionCount: number
  status: TellerSessionStatus
  notes: string
}

export const TELLER_SESSION_STATUS_LABELS: Record<TellerSessionStatus, string> = {
  open: 'Buka',
  closed: 'Tutup',
  reconciled: 'Direkonsiliasi',
}

// ── MoneyDenomination ────────────────────────────────────────────────────────

export type MoneyDenominationType = 'opening' | 'closing'

export interface MoneyDenomination extends BaseEntity {
  tellerSessionId: string
  sessionDate: string
  sessionStatus: string
  denomination: number
  count: number
  total: number
  type: MoneyDenominationType
}

export const MONEY_DENOMINATION_TYPE_LABELS: Record<MoneyDenominationType, string> = {
  opening: 'Pembukaan',
  closing: 'Penutupan',
}

// ── Kas ──────────────────────────────────────────────────────────────────────

export type KasType = 'masuk' | 'keluar'
export type KasStatus = 'pending' | 'posted' | 'reversed'

export interface Kas extends BaseEntity {
  tellerSessionId: string
  sessionDate: string
  sessionStatus: string
  kasType: KasType
  amount: number
  description: string
  referenceNo: string
  postedDate: string
  status: KasStatus
}

export const KAS_TYPE_LABELS: Record<KasType, string> = {
  masuk: 'Masuk',
  keluar: 'Keluar',
}

export const KAS_STATUS_LABELS: Record<KasStatus, string> = {
  pending: 'Pending',
  posted: 'Posted',
  reversed: 'Reversed',
}

// ── Jurnal & COA (ADR-K015) ─────────────────────────────────────────────────

// COA
export type COAAccountType = 'asset' | 'liability' | 'equity' | 'revenue' | 'expense' | 'zakat' | 'kebajikan' | 'tazir'
export type COANormalBalance = 'debit' | 'credit'
export type COACoopTypeRequired = 'general' | 'islamic' | 'both'

export interface COA extends BaseEntity {
  accountCode: string
  accountName: string
  parentId: string
  level: number
  accountType: COAAccountType
  normalBalance: COANormalBalance
  description: string
  isSystem: boolean
  isActive: boolean
  coopTypeRequired: COACoopTypeRequired
}

export const COA_ACCOUNT_TYPE_LABELS: Record<COAAccountType, string> = {
  asset: 'Aset',
  liability: 'Kewajiban',
  equity: 'Ekuitas',
  revenue: 'Pendapatan',
  expense: 'Beban',
  zakat: 'Zakat',
  kebajikan: 'Kebajikan',
  tazir: "Ta'zir",
}

export const COA_NORMAL_BALANCE_LABELS: Record<COANormalBalance, string> = {
  debit: 'Debit',
  credit: 'Kredit',
}

export const COA_COOP_TYPE_LABELS: Record<COACoopTypeRequired, string> = {
  general: 'Konvensional',
  islamic: 'Syariah',
  both: 'Semua',
}

// Jurnal
export type JurnalSourceType = 'transaction' | 'manual' | 'closing' | 'adjustment' | 'shu_distribution'
export type JurnalStatus = 'unposted' | 'posted' | 'reversed'

export interface Jurnal extends BaseEntity {
  journalNumber: string
  journalDate: string
  description: string
  sourceType: JurnalSourceType
  sourceId: string
  reference: string
  periodId: string
  branchId: string
  branchName: string
  periodName: string
  totalDebit: number
  totalCredit: number
  status: JurnalStatus
  isAutoPost: boolean
}

export const JURNAL_SOURCE_TYPE_LABELS: Record<JurnalSourceType, string> = {
  transaction: 'Transaksi',
  manual: 'Manual',
  closing: 'Penutupan',
  adjustment: 'Penyesuaian',
  shu_distribution: 'Distribusi SHU',
}

export const JURNAL_STATUS_LABELS: Record<JurnalStatus, string> = {
  unposted: 'Belum Di-posting',
  posted: 'Di-posting',
  reversed: 'Dibalik',
}

// Accounting Period
export type APPeriodType = 'monthly' | 'annual'
export type APStatus = 'open' | 'closed' | 'locked'

export interface AccountingPeriod extends BaseEntity {
  periodType: APPeriodType
  year: number
  month: number | null
  periodName: string
  startDate: string
  endDate: string
  status: APStatus
}

export const AP_PERIOD_TYPE_LABELS: Record<APPeriodType, string> = {
  monthly: 'Bulanan',
  annual: 'Tahunan',
}

export const AP_STATUS_LABELS: Record<APStatus, string> = {
  open: 'Terbuka',
  closed: 'Ditutup',
  locked: 'Terkunci',
}

// Journal Mapping
export type JMapCoopType = 'general' | 'islamic' | 'both'

export interface JournalMapping extends BaseEntity {
  transactionType: string
  coopType: JMapCoopType
  rules: Record<string, unknown>[]
  description: string
  isActive: boolean
}

export const JMAP_COOP_TYPE_LABELS: Record<JMapCoopType, string> = {
  general: 'Konvensional',
  islamic: 'Syariah',
  both: 'Semua',
}

// ── SHU (ADR-K016) ───────────────────────────────────────────────────────────

// SHU Periode
export type SHUPeriodeStatus = 'calculated' | 'reviewed' | 'approved' | 'distributed'

export interface SHUPeriode extends BaseEntity {
  tahunBuku: number
  periodStart: string
  periodEnd: string
  totalPendapatan: number
  totalBeban: number
  shuBruto: number
  shuNeto: number
  distributionConfig: Record<string, unknown>
  status: SHUPeriodeStatus
}

export const SHU_PERIODE_STATUS_LABELS: Record<SHUPeriodeStatus, string> = {
  calculated: 'Dihitung',
  reviewed: 'Direview',
  approved: 'Disetujui',
  distributed: 'Didistribusikan',
}

// SHU Anggota
export type SHUAnggotaDistMethod = 'credit_tabungan' | 'separate_payout' | 'pending'

export interface SHUAnggota extends BaseEntity {
  shuPeriodeId: string
  nasabahId: string
  nasabahNama: string
  nasabahNo: string
  avgSimpanan: number
  totalTransaksi: number
  activeDays: number
  jasaModal: number
  jasaUsaha: number
  totalSHU: number
  distributionMethod: SHUAnggotaDistMethod
  targetRekeningId: string
  targetRekeningNo: string
}

export const SHU_ANGGOTA_DIST_METHOD_LABELS: Record<SHUAnggotaDistMethod, string> = {
  credit_tabungan: 'Kredit Tabungan',
  separate_payout: 'Pembayaran Terpisah',
  pending: 'Menunggu',
}
