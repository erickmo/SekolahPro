import { createVernonService } from './vernon.service'
import type {
  Transaksi,
  TellerSession,
  MoneyDenomination,
  Kas,
} from '@/types/koperasi-ops.types'

// ── Transaksi ─────────────────────────────────────────────────────────────────

function transformTransaksi(raw: Record<string, unknown>): Transaksi {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    rekeningId: (raw.rekening_id ?? '') as string,
    rekeningNo: (data.no_rekening ?? '') as string,
    transactionType: (raw.transaction_type ?? 'credit') as Transaksi['transactionType'],
    amount: (raw.amount ?? 0) as number,
    description: (raw.description ?? '') as string,
    referenceNo: (raw.reference_no ?? '') as string,
    balanceBefore: (raw.balance_before ?? 0) as number,
    balanceAfter: (raw.balance_after ?? 0) as number,
    postedDate: (raw.posted_date ?? '') as string,
    status: (raw.status ?? 'pending') as Transaksi['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const transaksiService = createVernonService<Transaksi, Record<string, unknown>>(
  '/transaksi',
  transformTransaksi,
)

// ── TellerSession ──────────────────────────────────────────────────────────────

function transformTellerSession(raw: Record<string, unknown>): TellerSession {
  return {
    id: raw.id as string,
    userId: (raw.user_id ?? '') as string,
    sessionDate: (raw.session_date ?? '') as string,
    startTime: (raw.start_time ?? '') as string,
    endTime: (raw.end_time ?? '') as string,
    openingBalance: (raw.opening_balance ?? 0) as number,
    closingBalance: (raw.closing_balance ?? 0) as number,
    cashTotal: (raw.cash_total ?? 0) as number,
    transactionCount: (raw.transaction_count ?? 0) as number,
    status: (raw.status ?? 'open') as TellerSession['status'],
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const tellerSessionService = createVernonService<TellerSession, Record<string, unknown>>(
  '/teller_sessions',
  transformTellerSession,
)

// ── MoneyDenomination ──────────────────────────────────────────────────────────

function transformMoneyDenomination(raw: Record<string, unknown>): MoneyDenomination {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    tellerSessionId: (raw.teller_session_id ?? '') as string,
    sessionDate: (data.session_date ?? '') as string,
    sessionStatus: (data.status ?? '') as string,
    denomination: (raw.denomination ?? 0) as number,
    count: (raw.count ?? 0) as number,
    total: (raw.total ?? 0) as number,
    type: (raw.type ?? 'opening') as MoneyDenomination['type'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const moneyDenominationService = createVernonService<MoneyDenomination, Record<string, unknown>>(
  '/money_denominations',
  transformMoneyDenomination,
)

// ── Kas ────────────────────────────────────────────────────────────────────────

function transformKas(raw: Record<string, unknown>): Kas {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    tellerSessionId: (raw.teller_session_id ?? '') as string,
    sessionDate: (data.session_date ?? '') as string,
    sessionStatus: (data.status ?? '') as string,
    kasType: (raw.kas_type ?? 'masuk') as Kas['kasType'],
    amount: (raw.amount ?? 0) as number,
    description: (raw.description ?? '') as string,
    referenceNo: (raw.reference_no ?? '') as string,
    postedDate: (raw.posted_date ?? '') as string,
    status: (raw.status ?? 'pending') as Kas['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const kasService = createVernonService<Kas, Record<string, unknown>>(
  '/kas',
  transformKas,
)
