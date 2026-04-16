import { createVernonService } from './vernon.service'
import type {
  ProdukAkad,
  Nasabah,
  Rekening,
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
