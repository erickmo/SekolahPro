# 17 — Vernon Accounting Integration

> **Terakhir diperbarui**: 20 Maret 2026
> **Status**: Placeholder — modul akan diisi saat `module-vernon-accounting` tersedia di `/mnt/skills/`

`module-vernon-accounting` digunakan sebagai engine pembukuan untuk seluruh transaksi keuangan EDS. Modul ini menggantikan logika akuntansi ad-hoc di **M03 Koperasi**, **M18 SPP Online**, **M51 RKAS**, dan **M52 Dana Komite**.

---

## Arsitektur Integrasi

```
EDS Financial Modules
  ├── M03  Koperasi & Tabungan  ─────┐
  ├── M18  Pembayaran SPP       ─────┤
  ├── M51  Anggaran & RKAS      ─────┤──► module-vernon-accounting
  └── M52  Dana Komite          ─────┘         │
                                               ├── General Ledger
                                               ├── Chart of Accounts
                                               ├── Journal Entries
                                               ├── Financial Reports
                                               └── Multi-entity (per schoolId)
```

---

## Chart of Accounts Default Sekolah

Vernon Accounting menggunakan CoA (Chart of Accounts) yang disesuaikan untuk sekolah Indonesia:

```
1xxx  ASET
  1100  Kas & Bank
    1101  Kas Tunai
    1102  Rekening BOS
    1103  Rekening Komite
    1104  EDS Wallet (M19 Kantin)
  1200  Piutang
    1201  Piutang SPP
    1202  Piutang Tabungan Siswa

2xxx  KEWAJIBAN
  2100  Hutang Jangka Pendek
    2101  Tabungan Siswa (liability koperasi)
    2102  Titipan Dana Kegiatan

3xxx  EKUITAS
  3100  Saldo Dana BOS
  3200  Saldo Dana Komite

4xxx  PENDAPATAN
  4100  SPP
  4200  Dana BOS
  4300  Sumbangan & Donasi
  4400  Pendapatan Koperasi
  4500  Pendapatan Kantin

5xxx  BEBAN
  5100  Beban Gaji & Honor
  5200  Beban Operasional
  5300  Beban Sarana Prasarana
  5400  Beban Kegiatan Siswa
  5500  Beban Administrasi
```

---

## Mapping Transaksi EDS → Vernon Accounting

### M03 — Koperasi & Tabungan

| Transaksi EDS | Journal Entry Vernon |
|---------------|----------------------|
| Setoran tabungan siswa | Dr 1101 Kas Tunai / Cr 2101 Tabungan Siswa |
| Penarikan tabungan | Dr 2101 Tabungan Siswa / Cr 1101 Kas Tunai |
| Penjualan produk koperasi | Dr 1101 Kas / Cr 4400 Pendapatan Koperasi |

### M18 — Pembayaran SPP

| Transaksi EDS | Journal Entry Vernon |
|---------------|----------------------|
| Pembayaran SPP diterima | Dr 1102 Rek. SPP / Cr 4100 Pendapatan SPP |
| Refund SPP | Dr 4100 Pendapatan SPP / Cr 1102 Rek. SPP |
| Piutang SPP (belum bayar) | Dr 1201 Piutang SPP / Cr 4100 Pendapatan SPP |

### M51 — RKAS & Realisasi Anggaran

| Aksi EDS | Aksi Vernon |
|----------|-------------|
| Item RKAS diapprove | Budget allocation per CoA item |
| Realisasi pengeluaran di-input | Journal entry beban + update budget utilization |
| Dana BOS masuk | Dr 1102 Rek. BOS / Cr 3100 Saldo Dana BOS |

### M52 — Dana Komite

| Transaksi EDS | Journal Entry Vernon |
|---------------|----------------------|
| Donasi/sumbangan masuk | Dr 1103 Rek. Komite / Cr 4300 Sumbangan |
| Pengeluaran dari dana komite | Dr 5xxx Beban / Cr 1103 Rek. Komite |

---

## Laporan Keuangan dari Vernon Accounting

Semua laporan dihasilkan oleh Vernon Accounting engine, bukan di-generate manual di EDS:

- [ ] **Neraca** (Balance Sheet) — per sekolah, per periode
- [ ] **Laporan Laba Rugi** — pendapatan vs. beban operasional
- [ ] **Laporan Arus Kas** — cash flow bulanan
- [ ] **Laporan Realisasi RKAS** — rencana vs. realisasi per item
- [ ] **Rekap Tabungan Siswa** — saldo per siswa, per kelas
- [ ] **Laporan Tunggakan SPP** — aging receivables
- [ ] **Laporan BOS** — penerimaan & penggunaan dana BOS (format Kemendikbud)
- [ ] Export ke PDF (skill `pdf`) dan Excel (skill `xlsx`)

---

## Konfigurasi Tenant Vernon Accounting

```typescript
// packages/accounting/client.ts
export async function initVernonAccountingTenant(schoolId: string, config: SchoolFinanceConfig) {
  return await vernonAccounting.entities.create({
    entityId: `eds_${schoolId}`,
    name: config.schoolName,
    fiscalYearStart: 'JANUARY',   // atau JULY untuk sekolah dengan tahun ajaran Juli
    currency: 'IDR',
    coa: config.coaTemplate ?? 'SEKOLAH_INDONESIA_DEFAULT',
    taxId: config.npwp,
  });
}
```

---

## Event Bridge EDS ↔ Vernon Accounting

```typescript
type AccountingEvent =
  | { type: 'payment.received';    module: 'M18'; amount: number; studentId: string; period: string }
  | { type: 'savings.deposit';     module: 'M03'; amount: number; studentId: string }
  | { type: 'savings.withdrawal';  module: 'M03'; amount: number; studentId: string }
  | { type: 'budget.realized';     module: 'M51'; itemId: string; amount: number; description: string }
  | { type: 'donation.received';   module: 'M52'; amount: number; donorName?: string }
  | { type: 'canteen.sale';        module: 'M19'; amount: number; studentId: string }
```

Setiap event di atas secara otomatis men-generate journal entry di Vernon Accounting. EDS tidak perlu tahu detail debit/kredit — itu urusan Vernon.

---

## Audit Trail Keuangan

Kombinasi **M45 Audit Log EDS** + **Vernon Accounting audit trail** memberikan dua lapis jejak:

```
Transaksi SPP masuk
  ├── M45 Audit Log: userId=kasir, action=CREATE, entity=Payment, timestamp, IP
  └── Vernon Accounting: journal entry #JNL-2026-003421, debit/kredit, diposting oleh sistem
```

Kedua log ini immutable dan tersimpan > 5 tahun.

---

## Instalasi & Konfigurasi

> Baca `module-vernon-accounting/SKILL.md` terlebih dahulu sebelum implementasi.
> Path skill: `/mnt/skills/user/module-vernon-accounting/SKILL.md` (install terlebih dahulu)

```bash
# Start Vernon Accounting service
docker-compose -f infrastructure/docker-compose.yml up vernon-accounting -d

# Inisialisasi chart of accounts untuk semua tenant
pnpm --filter @eds/accounting run init-coa

# Migrasi data transaksi historis
pnpm --filter @eds/accounting run migrate-transactions --from=2024-01-01
```

---

## Environment Variables

```env
VERNON_ACC_API_URL="http://vernon-accounting:5200"
VERNON_ACC_API_KEY="..."
VERNON_ACC_TENANT_PREFIX="eds_"
VERNON_ACC_DEFAULT_COA="SEKOLAH_INDONESIA_DEFAULT"
VERNON_ACC_WEBHOOK_SECRET="..."
VERNON_ACC_FISCAL_YEAR_START="JANUARY"   # JANUARY atau JULY
```

---

## Catatan Penting

1. **Vernon Accounting adalah source of truth untuk semua angka keuangan** — data di tabel `Transaction`, `Invoice` EDS hanya untuk UI; nilai buku resmi ada di Vernon.
2. **Double-entry bookkeeping wajib** — setiap transaksi harus balance. Jika Vernon menolak entry karena tidak balance, transaksi EDS harus di-rollback.
3. **Multi-entity per schoolId** — satu instance Vernon Accounting melayani semua sekolah dengan isolasi per entity.
4. **Fiscal year konfigurasi per sekolah** — beberapa sekolah menggunakan tahun anggaran Januari–Desember, yang lain Juli–Juni.
5. **Reconciliation harian otomatis** — cron job jam 23.00 WIB mencocokkan saldo EDS vs. Vernon Accounting. Jika ada gap, buat alert ke bendahara dan tim teknis.
