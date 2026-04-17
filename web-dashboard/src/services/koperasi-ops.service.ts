import { createVernonService } from './vernon.service'
import type {
  Transaksi,
  TellerSession,
  MoneyDenomination,
  Kas,
  COA,
  Jurnal,
  AccountingPeriod,
  JournalMapping,
  SHUPeriode,
  SHUAnggota,
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

// ── COA (ADR-K015) ────────────────────────────────────────────────────────────

function transformCOA(raw: Record<string, unknown>): COA {
  return {
    id: raw.id as string,
    accountCode: (raw.account_code ?? '') as string,
    accountName: (raw.account_name ?? '') as string,
    parentId: (raw.parent_id ?? '') as string,
    level: (raw.level ?? 1) as number,
    accountType: (raw.account_type ?? 'asset') as COA['accountType'],
    normalBalance: (raw.normal_balance ?? 'debit') as COA['normalBalance'],
    description: (raw.description ?? '') as string,
    isSystem: (raw.is_system ?? false) as boolean,
    isActive: (raw.is_active ?? true) as boolean,
    coopTypeRequired: (raw.coop_type_required ?? 'both') as COA['coopTypeRequired'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const coaService = createVernonService<COA, Record<string, unknown>>(
  '/coa',
  transformCOA,
)

// ── Jurnal (ADR-K015) ─────────────────────────────────────────────────────────

function transformJurnal(raw: Record<string, unknown>): Jurnal {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const branch = (data.branch ?? {}) as Record<string, unknown>
  const period = (data.period ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    journalNumber: (raw.journal_number ?? '') as string,
    journalDate: (raw.journal_date ?? '') as string,
    description: (raw.description ?? '') as string,
    sourceType: (raw.source_type ?? 'manual') as Jurnal['sourceType'],
    sourceId: (raw.source_id ?? '') as string,
    reference: (raw.reference ?? '') as string,
    periodId: (raw.period_id ?? '') as string,
    branchId: (raw.branch_id ?? '') as string,
    branchName: (branch.name ?? '') as string,
    periodName: (period.period_name ?? '') as string,
    totalDebit: (raw.total_debit ?? 0) as number,
    totalCredit: (raw.total_credit ?? 0) as number,
    status: (raw.status ?? 'unposted') as Jurnal['status'],
    isAutoPost: (raw.is_auto_post ?? false) as boolean,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const jurnalService = createVernonService<Jurnal, Record<string, unknown>>(
  '/jurnal',
  transformJurnal,
)

// ── Accounting Period (ADR-K015) ──────────────────────────────────────────────

function transformAccountingPeriod(raw: Record<string, unknown>): AccountingPeriod {
  return {
    id: raw.id as string,
    periodType: (raw.period_type ?? 'monthly') as AccountingPeriod['periodType'],
    year: (raw.year ?? 2026) as number,
    month: (raw.month ?? null) as number | null,
    periodName: (raw.period_name ?? '') as string,
    startDate: (raw.start_date ?? '') as string,
    endDate: (raw.end_date ?? '') as string,
    status: (raw.status ?? 'open') as AccountingPeriod['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const accountingPeriodService = createVernonService<AccountingPeriod, Record<string, unknown>>(
  '/accounting_period',
  transformAccountingPeriod,
)

// ── Journal Mapping (ADR-K015) ────────────────────────────────────────────────

function transformJournalMapping(raw: Record<string, unknown>): JournalMapping {
  return {
    id: raw.id as string,
    transactionType: (raw.transaction_type ?? '') as string,
    coopType: (raw.coop_type ?? 'both') as JournalMapping['coopType'],
    rules: (raw.rules ?? []) as Record<string, unknown>[],
    description: (raw.description ?? '') as string,
    isActive: (raw.is_active ?? true) as boolean,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const journalMappingService = createVernonService<JournalMapping, Record<string, unknown>>(
  '/journal_mapping',
  transformJournalMapping,
)

// ── SHU Periode (ADR-K016) ────────────────────────────────────────────────────

function transformSHUPeriode(raw: Record<string, unknown>): SHUPeriode {
  return {
    id: raw.id as string,
    tahunBuku: (raw.tahun_buku ?? 0) as number,
    periodStart: (raw.period_start ?? '') as string,
    periodEnd: (raw.period_end ?? '') as string,
    totalPendapatan: (raw.total_pendapatan ?? 0) as number,
    totalBeban: (raw.total_beban ?? 0) as number,
    shuBruto: (raw.shu_bruto ?? 0) as number,
    shuNeto: (raw.shu_neto ?? 0) as number,
    distributionConfig: (raw.distribution_config ?? {}) as Record<string, unknown>,
    status: (raw.status ?? 'calculated') as SHUPeriode['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const shuPeriodeService = createVernonService<SHUPeriode, Record<string, unknown>>(
  '/shu_periode',
  transformSHUPeriode,
)

// ── SHU Anggota (ADR-K016) ────────────────────────────────────────────────────

function transformSHUAnggota(raw: Record<string, unknown>): SHUAnggota {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const nasabah = (data.nasabah ?? {}) as Record<string, unknown>
  const rekening = (data.target_rekening ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    shuPeriodeId: (raw.shu_periode_id ?? '') as string,
    nasabahId: (raw.nasabah_id ?? '') as string,
    nasabahNama: (nasabah.nama_lengkap ?? '') as string,
    nasabahNo: (nasabah.no_nasabah ?? '') as string,
    avgSimpanan: (raw.avg_simpanan ?? 0) as number,
    totalTransaksi: (raw.total_transaksi ?? 0) as number,
    activeDays: (raw.active_days ?? 0) as number,
    jasaModal: (raw.jasa_modal ?? 0) as number,
    jasaUsaha: (raw.jasa_usaha ?? 0) as number,
    totalSHU: (raw.total_shu ?? 0) as number,
    distributionMethod: (raw.distribution_method ?? 'pending') as SHUAnggota['distributionMethod'],
    targetRekeningId: (raw.target_rekening_id ?? '') as string,
    targetRekeningNo: (rekening.no_rekening ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const shuAnggotaService = createVernonService<SHUAnggota, Record<string, unknown>>(
  '/shu_anggota',
  transformSHUAnggota,
)
