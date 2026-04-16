import { createVernonService } from './vernon.service'
import type {
  ProdukAkad,
  Nasabah,
  Rekening,
  SimpananPokokWajib,
  Tabungan,
  Deposito,
} from '@/types/koperasi.types'

// ── ProdukAkad ───────────────────────────────────────────────────────────────

function transformProdukAkad(raw: Record<string, unknown>): ProdukAkad {
  return {
    id: raw.id as string,
    kode: (raw.kode ?? '') as string,
    namaProduk: (raw.nama_produk ?? '') as string,
    type: (raw.type ?? 'pembiayaan') as ProdukAkad['type'],
    akadType: (raw.akad_type ?? 'murabahah') as ProdukAkad['akadType'],
    deskripsi: (raw.deskripsi ?? '') as string,
    status: (raw.status ?? 'draft') as ProdukAkad['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const produkAkadService = createVernonService<ProdukAkad, Record<string, unknown>>(
  '/produk_akad',
  transformProdukAkad,
)

// ── Nasabah ──────────────────────────────────────────────────────────────────

function transformNasabah(raw: Record<string, unknown>): Nasabah {
  return {
    id: raw.id as string,
    noNasabah: (raw.no_nasabah ?? '') as string,
    namaLengkap: (raw.nama_lengkap ?? '') as string,
    nik: (raw.nik ?? '') as string,
    type: (raw.type ?? 'perorangan') as Nasabah['type'],
    status: (raw.status ?? 'pending') as Nasabah['status'],
    tanggalDaftar: (raw.tanggal_daftar ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const nasabahService = createVernonService<Nasabah, Record<string, unknown>>(
  '/nasabah',
  transformNasabah,
)

// ── Rekening ─────────────────────────────────────────────────────────────────

function transformRekening(raw: Record<string, unknown>): Rekening {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    noRekening: (raw.no_rekening ?? '') as string,
    nasabahId: (raw.nasabah_id ?? '') as string,
    nasabahNama: (data.nasabah_nama ?? '') as string,
    produkAkadId: (raw.produk_akad_id ?? '') as string,
    produkAkadNama: (data.produk_akad_nama ?? '') as string,
    saldo: (raw.saldo ?? 0) as number,
    status: (raw.status ?? 'pending_open') as Rekening['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const rekeningService = createVernonService<Rekening, Record<string, unknown>>(
  '/rekening',
  transformRekening,
)

// ── SimpananPokokWajib ─────────────────────────────────────────────────────

function transformSimpananPokokWajib(raw: Record<string, unknown>): SimpananPokokWajib {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    nasabahId: (raw.nasabah_id ?? '') as string,
    nasabahNama: (data.nasabah_nama ?? '') as string,
    noRekening: (raw.no_rekening ?? '') as string,
    jenis: (raw.jenis ?? 'pokok') as SimpananPokokWajib['jenis'],
    nominal: (raw.nominal ?? 0) as number,
    periode: (raw.periode ?? '') as string,
    jatuhTempo: (raw.jatuh_tempo ?? '') as string,
    status: (raw.status ?? 'pending') as SimpananPokokWajib['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const simpananPokokWajibService = createVernonService<SimpananPokokWajib, Record<string, unknown>>(
  '/simpanan_pokok_wajib',
  transformSimpananPokokWajib,
)

// ── Tabungan ────────────────────────────────────────────────────────────────

function transformTabungan(raw: Record<string, unknown>): Tabungan {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    noRekening: (raw.no_rekening ?? '') as string,
    nasabahId: (raw.nasabah_id ?? '') as string,
    nasabahNama: (data.nasabah_nama ?? '') as string,
    produk: (raw.produk ?? 'regular') as Tabungan['produk'],
    saldo: (raw.saldo ?? 0) as number,
    targetGoal: (raw.target_goal ?? null) as number | null,
    status: (raw.status ?? 'active') as Tabungan['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const tabunganService = createVernonService<Tabungan, Record<string, unknown>>(
  '/tabungan',
  transformTabungan,
)

// ── Deposito ────────────────────────────────────────────────────────────────

function transformDeposito(raw: Record<string, unknown>): Deposito {
  const data = (raw._data ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    noRekening: (raw.no_rekening ?? '') as string,
    nasabahId: (raw.nasabah_id ?? '') as string,
    nasabahNama: (data.nasabah_nama ?? '') as string,
    nominal: (raw.nominal ?? 0) as number,
    tenor: (raw.tenor ?? 1) as number,
    bunga: (raw.bunga ?? 0) as number,
    jatuhTempo: (raw.jatuh_tempo ?? '') as string,
    status: (raw.status ?? 'active') as Deposito['status'],
    autoRoll: (raw.auto_roll ?? false) as boolean,
    onHold: (raw.on_hold ?? false) as boolean,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const depositoService = createVernonService<Deposito, Record<string, unknown>>(
  '/deposito',
  transformDeposito,
)
