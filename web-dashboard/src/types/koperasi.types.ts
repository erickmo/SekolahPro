import type { BaseEntity } from './entity.types'

// ── ProdukAkad ───────────────────────────────────────────────────────────────

export type ProdukAkadType = 'pembiayaan' | 'tabungan' | 'deposito'
export type AkadType = 'murabahah' | 'mudharabah' | 'musyarakah' | 'istishna' | 'salam' | 'ijarah' | 'wadiyah' | 'mudharabah_muthlaqoh'
export type ProdukAkadStatus = 'draft' | 'active' | 'inactive' | 'archived'

export interface ProdukAkad extends BaseEntity {
  kode: string
  namaProduk: string
  type: ProdukAkadType
  akadType: AkadType
  deskripsi: string
  status: ProdukAkadStatus
}

export const PRODUK_AKAD_TYPE_LABELS: Record<ProdukAkadType, string> = {
  pembiayaan: 'Pembiayaan',
  tabungan: 'Tabungan',
  deposito: 'Deposito',
}

export const AKAD_TYPE_LABELS: Record<AkadType, string> = {
  murabahah: 'Murabahah',
  mudharabah: 'Mudharabah',
  musyarakah: 'Musyarakah',
  istishna: 'Istishna',
  salam: 'Salam',
  ijarah: 'Ijarah',
  wadiyah: 'Wadiyah',
  mudharabah_muthlaqoh: 'Mudharabah Muthlaqoh',
}

export const PRODUK_AKAD_STATUS_LABELS: Record<ProdukAkadStatus, string> = {
  draft: 'Draft',
  active: 'Aktif',
  inactive: 'Nonaktif',
  archived: 'Diarsipkan',
}

// ── Nasabah ──────────────────────────────────────────────────────────────────

export type NasabahType = 'perorangan' | 'badan_usaha' | 'kelompok'
export type NasabahStatus = 'pending' | 'active' | 'frozen' | 'closed'

export interface Nasabah extends BaseEntity {
  noNasabah: string
  namaLengkap: string
  nik: string
  type: NasabahType
  status: NasabahStatus
  tanggalDaftar: string
}

export const NASABAH_TYPE_LABELS: Record<NasabahType, string> = {
  perorangan: 'Perorangan',
  badan_usaha: 'Badan Usaha',
  kelompok: 'Kelompok',
}

export const NASABAH_STATUS_LABELS: Record<NasabahStatus, string> = {
  pending: 'Pending',
  active: 'Aktif',
  frozen: 'Dibekukan',
  closed: 'Ditutup',
}

// ── Rekening ─────────────────────────────────────────────────────────────────

export type RekeningStatus = 'pending_open' | 'active' | 'dormant' | 'frozen' | 'pending_close' | 'closed'

export interface Rekening extends BaseEntity {
  noRekening: string
  nasabahId: string
  nasabahNama: string
  produkAkadId: string
  produkAkadNama: string
  saldo: number
  status: RekeningStatus
}

export const REKENING_STATUS_LABELS: Record<RekeningStatus, string> = {
  pending_open: 'Pending Buka',
  active: 'Aktif',
  dormant: 'Dorman',
  frozen: 'Dibekukan',
  pending_close: 'Pending Tutup',
  closed: 'Ditutup',
}

// ── SimpananPokokWajib ─────────────────────────────────────────────────────

export type SimpananPokokWajibJenis = 'pokok' | 'wajib'
export type SimpananPokokWajibStatus = 'pending' | 'paid' | 'overdue' | 'refunded'

export interface SimpananPokokWajib extends BaseEntity {
  nasabahId: string
  nasabahNama: string
  noRekening: string
  jenis: SimpananPokokWajibJenis
  nominal: number
  periode: string
  jatuhTempo: string
  status: SimpananPokokWajibStatus
}

export const SIMPANAN_POKOK_WAJIB_JENIS_LABELS: Record<SimpananPokokWajibJenis, string> = {
  pokok: 'Pokok',
  wajib: 'Wajib',
}

export const SIMPANAN_POKOK_WAJIB_STATUS_LABELS: Record<SimpananPokokWajibStatus, string> = {
  pending: 'Pending',
  paid: 'Lunas',
  overdue: 'Tunggakan',
  refunded: 'Dikembalikan',
}

// ── Tabungan ────────────────────────────────────────────────────────────────

export type TabunganProduk = 'regular' | 'education' | 'holiday' | 'qurban' | 'goal'
export type TabunganStatus = 'active' | 'dormant' | 'frozen' | 'closed'

export interface Tabungan extends BaseEntity {
  noRekening: string
  nasabahId: string
  nasabahNama: string
  produk: TabunganProduk
  saldo: number
  targetGoal: number | null
  status: TabunganStatus
}

export const TABUNGAN_PRODUK_LABELS: Record<TabunganProduk, string> = {
  regular: 'Regular',
  education: 'Pendidikan',
  holiday: 'Liburan',
  qurban: 'Qurban',
  goal: 'Goal',
}

export const TABUNGAN_STATUS_LABELS: Record<TabunganStatus, string> = {
  active: 'Aktif',
  dormant: 'Dorman',
  frozen: 'Dibekukan',
  closed: 'Ditutup',
}

// ── Deposito ────────────────────────────────────────────────────────────────

export type DepositoStatus = 'active' | 'matured' | 'rolled_over' | 'early_withdrawn' | 'closed'

export interface Deposito extends BaseEntity {
  noRekening: string
  nasabahId: string
  nasabahNama: string
  nominal: number
  tenor: number
  bunga: number
  jatuhTempo: string
  status: DepositoStatus
  autoRoll: boolean
  onHold: boolean
}

export const DEPOSITO_STATUS_LABELS: Record<DepositoStatus, string> = {
  active: 'Aktif',
  matured: 'Jatuh Tempo',
  rolled_over: 'Roll Over',
  early_withdrawn: 'Pencairan Awal',
  closed: 'Ditutup',
}
