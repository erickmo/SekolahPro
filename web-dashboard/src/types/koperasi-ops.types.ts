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
