# ADR-K015: Jurnal & Akuntansi (COA)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Akuntansi adalah **tulang punggung** pelaporan keuangan koperasi/BMT. Setiap transaksi yang terjadi di sistem — setoran tabungan, pencairan pinjaman, angsuran, denda — harus tercatat dalam jurnal akuntansi dengan prinsip **double-entry bookkeeping** (debit = kredit). Tanpa mesin akuntansi yang terintegrasi, laporan keuangan harus dibuat manual dan rawan kesalahan.

Sistem harus mengakomodasi:

- **Chart of Accounts (COA)** yang berbeda antara koperasi konvensional (PSAK/SAK ETAP) dan BMT (PSAK 101-110 Syariah)
- **Auto-journal**: setiap transaksi (K011) otomatis menghasilkan jurnal — operator tidak perlu input jurnal manual
- **Double-entry enforcement**: total debit harus selalu sama dengan total kredit — dijamin di database level
- **Period management**: pembukaan dan penutupan periode akuntansi (bulanan/tahunan)
- **Multi-branch consolidation**: jurnal per cabang, konsolidasi di level company/tenant
- **Dual-mode**: COA template, mapping jurnal, dan format laporan berbeda per `coop_type`

Referensi terkait:
- [ADR-009](../core/ADR-009-dual-mode-institution-type.md) — Dual-mode institution type
- [ADR-K002](./ADR-K002-rekening.md) — Rekening dan balance management
- [ADR-K011](./ADR-K011-transaksi.md) — Core transaction engine

## Decision

### 1. Chart of Accounts (COA) — Struktur Standar

COA mengikuti struktur standar koperasi Indonesia, menggunakan **kode numerik 4 digit** dengan hirarki:

```
KOPERASI KONVENSIONAL (PSAK / SAK ETAP):
──────────────────────────────────────────
1xxx — ASET
├── 1100  Kas
│   ├── 1101  Kas Teller
│   ├── 1102  Kas Besar (Brankas)
│   └── 1103  Kas Kecil
├── 1200  Bank
│   ├── 1201  Bank BRI
│   └── 1202  Bank Mandiri
├── 1300  Piutang
│   ├── 1301  Piutang Pinjaman Anggota
│   ├── 1302  Piutang Bunga Pinjaman
│   └── 1309  Cadangan Kerugian Piutang (contra)
├── 1400  Penyertaan / Investasi
├── 1500  Aktiva Tetap
│   ├── 1501  Inventaris Kantor
│   └── 1509  Akumulasi Penyusutan (contra)
└── 1600  Aktiva Lain-lain

2xxx — KEWAJIBAN
├── 2100  Simpanan Nasabah
│   ├── 2101  Simpanan Tabungan
│   ├── 2102  Simpanan Deposito
│   └── 2103  Bunga Simpanan yang Masih Harus Dibayar
├── 2200  Hutang
│   ├── 2201  Hutang Bank
│   └── 2202  Hutang Lain-lain
└── 2300  Kewajiban Lain-lain

3xxx — EKUITAS
├── 3100  Simpanan Pokok Anggota
├── 3200  Simpanan Wajib Anggota
├── 3300  Cadangan Umum
├── 3400  Cadangan Risiko
├── 3500  SHU Tahun Berjalan
└── 3600  SHU Tahun Lalu (Belum Dibagi)

4xxx — PENDAPATAN
├── 4100  Pendapatan Bunga Pinjaman
├── 4200  Pendapatan Administrasi
├── 4300  Pendapatan Denda
├── 4400  Pendapatan Provisi
└── 4900  Pendapatan Lain-lain

5xxx — BEBAN
├── 5100  Beban Bunga Simpanan
├── 5200  Beban Bunga Deposito
├── 5300  Beban Operasional
│   ├── 5301  Beban Gaji & Tunjangan
│   ├── 5302  Beban Sewa
│   ├── 5303  Beban ATK
│   └── 5304  Beban Penyusutan
├── 5400  Beban Cadangan Kerugian Piutang
└── 5900  Beban Lain-lain
```

```
BMT / KOPERASI SYARIAH (PSAK 101-110 Syariah):
──────────────────────────────────────────────────
Struktur 1xxx-5xxx SAMA dengan konvensional, dengan penyesuaian terminologi:
- 1302 → Piutang Margin Murabahah (bukan Bunga)
- 4100 → Pendapatan Margin/Bagi Hasil (bukan Bunga)
- 5100 → Beban Bagi Hasil Simpanan (bukan Bunga)

TAMBAHAN AKUN SYARIAH:
6xxx — DANA ZAKAT & INFAQ
├── 6100  Penerimaan Zakat
│   ├── 6101  Zakat Maal Anggota
│   ├── 6102  Zakat Fitrah
│   └── 6103  Zakat Institusi (dari laba BMT)
├── 6200  Penyaluran Zakat
│   ├── 6201  Penyaluran ke Fakir
│   ├── 6202  Penyaluran ke Miskin
│   ├── 6203  Penyaluran ke Amil
│   └── 6208  Penyaluran ke Ibnu Sabil
├── 6300  Penerimaan Infaq/Shadaqah
└── 6400  Penyaluran Infaq/Shadaqah

7xxx — DANA KEBAJIKAN (QARDHUL HASAN)
├── 7100  Penerimaan Dana Kebajikan
│   ├── 7101  Infaq dari Anggota
│   └── 7102  Pengembalian Qardh
├── 7200  Penyaluran Dana Kebajikan
│   ├── 7201  Pinjaman Qardh (tanpa margin)
│   └── 7202  Sumbangan Kebajikan
└── 7300  Saldo Dana Kebajikan

8xxx — DANA TA'ZIR (SOCIAL FUND)
├── 8100  Penerimaan Ta'zir
│   └── 8101  Denda Keterlambatan Anggota
├── 8200  Penyaluran Dana Sosial
│   ├── 8201  Bantuan Sosial Anggota
│   └── 8202  Kegiatan Sosial
└── 8300  Saldo Dana Ta'zir
```

**Catatan:**
- Akun 6xxx, 7xxx, 8xxx **hanya tersedia** jika `coop_type = "islamic"` — hidden entirely di mode general
- Akun syariah ini bersifat **off-balance sheet** terhadap laporan laba rugi koperasi (lihat Section 10)
- Kode 4 digit adalah **level standar** — tenant bisa menambah sub-akun (5+ digit) tapi tidak bisa mengubah struktur 4 digit

### 2. COA per Tenant — Template + Customization

Setiap tenant mendapat **COA template default** berdasarkan `coop_type`, dan bisa melakukan customization terbatas.

```
Tenant Onboarding
        │
        v
┌─────────────────────────────┐
│ Auto-seed COA template      │
│ berdasarkan coop_type       │
│ (general / islamic)         │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│ Tenant customization:       │
│ ✓ Tambah sub-akun (5+ digit)│
│ ✓ Edit nama akun            │
│ ✓ Nonaktifkan akun          │
│   (jika saldo = 0)          │
│ ✗ Hapus system account      │
│ ✗ Ubah kode system account  │
│ ✗ Ubah hirarki 4-digit      │
└─────────────────────────────┘
```

**Aturan:**
- **System accounts** (semua akun 4-digit dari template) tidak bisa dihapus atau diubah kodenya — ini akun yang dibutuhkan untuk auto-journal dan laporan regulasi
- Tenant bisa menambahkan **sub-akun** dengan kode 5+ digit (misal: 11011 untuk "Kas Teller Cabang A") — parent harus akun yang sudah ada
- Tenant bisa **menonaktifkan** akun yang tidak dipakai (is_active = false), tapi hanya jika saldo akun = 0 dan tidak ada jurnal unposted yang mereferensikan akun tersebut
- Tenant bisa **mengedit nama** akun (misal: rename "Bank BRI" jadi "Bank BSI") — kode tetap sama
- COA di-seed **satu kali saat onboarding** — perubahan template global tidak otomatis terapply ke tenant yang sudah ada

### 3. Auto-Journal — Transaction-to-Journal Mapping

Setiap transaksi dari K011 **otomatis** menghasilkan jurnal entry. Tidak ada input jurnal manual untuk transaksi operasional.

```
TransactionCreatedEvent (dari K011)
        │
        v
┌─────────────────────────────────────┐
│ JournalEngine                       │
│ 1. Lookup mapping: tx_type → rules  │
│ 2. Resolve akun debit & kredit      │
│ 3. Create jurnal header + lines     │
│ 4. Validate: SUM debit = SUM kredit │
│ 5. Simpan sebagai UNPOSTED          │
└─────────────────────────────────────┘
```

**Mapping standar (contoh):**

| Transaksi | Debit | Kredit | Catatan |
|-----------|-------|--------|---------|
| Setoran tunai tabungan | 1101 Kas Teller | 2101 Simpanan Tabungan | Kas masuk, kewajiban naik |
| Penarikan tunai tabungan | 2101 Simpanan Tabungan | 1101 Kas Teller | Kewajiban turun, kas keluar |
| Setoran simpanan pokok | 1101 Kas Teller | 3100 Simpanan Pokok | Kas masuk, ekuitas naik |
| Setoran simpanan wajib | 1101 Kas Teller | 3200 Simpanan Wajib | Kas masuk, ekuitas naik |
| Pencairan pinjaman (tunai) | 1301 Piutang Pinjaman | 1101 Kas Teller | Piutang naik, kas keluar |
| Angsuran pokok (tunai) | 1101 Kas Teller | 1301 Piutang Pinjaman | Kas masuk, piutang turun |
| Angsuran bunga/margin | 1101 Kas Teller | 4100 Pendapatan Bunga/Margin | Kas masuk, pendapatan naik |
| Biaya admin bulanan | 2101 Simpanan Tabungan | 4200 Pendapatan Admin | Kewajiban turun, pendapatan naik |
| Pembagian bunga/bagi hasil tabungan | 5100 Beban Bunga/Bagi Hasil | 2101 Simpanan Tabungan | Beban naik, kewajiban naik |
| Denda keterlambatan (general) | 1101 Kas Teller | 4300 Pendapatan Denda | Kas masuk, pendapatan naik |
| Denda keterlambatan (islamic) | 1101 Kas Teller | 8101 Penerimaan Ta'zir | Kas masuk, dana sosial naik |
| Penerimaan zakat (islamic) | 1101 Kas Teller | 6101 Zakat Maal | Kas masuk, dana zakat naik |

**Aturan:**
- Mapping disimpan sebagai **konfigurasi** (`journal_mapping` table), bukan hardcode — tenant bisa menyesuaikan akun tujuan
- Mapping default di-seed saat onboarding berdasarkan `coop_type`
- Setiap mapping wajib menghasilkan **SUM debit = SUM kredit** — mapping yang tidak balance tidak bisa disimpan
- Transaksi yang gagal generate jurnal akan **menggagalkan seluruh transaksi** (dalam satu DB transaction)
- Mapping bisa menghasilkan **lebih dari 2 lines** — misal: angsuran = debit kas + kredit piutang pokok + kredit pendapatan bunga
- Mapping yang melibatkan akun 6xxx-8xxx hanya tersedia jika `coop_type = "islamic"`

### 4. Double-Entry Enforcement

Setiap jurnal entry **wajib** memenuhi prinsip double-entry: total debit = total kredit. Enforcement di 3 level:

```
Level 1: Application Layer
├── JournalEngine memvalidasi sebelum insert
│
Level 2: Database Constraint
├── CHECK constraint pada jurnal header:
│   "SUM(debit) of lines = SUM(credit) of lines"
│   (via trigger atau generated column)
│
Level 3: Scheduled Reconciliation
└── Job harian: scan semua jurnal, alert jika ada
    yang lolos dari Level 1 & 2
```

**Aturan:**
- Jurnal yang tidak balance **tidak bisa disimpan** — database akan reject
- Tidak ada toleransi selisih (zero tolerance) — berbeda dari sistem yang mengizinkan selisih pembulatan
- Semua amount menggunakan **NUMERIC(15,2)** — presisi 2 desimal, cukup untuk IDR
- Amount di jurnal line selalu **positif** — arah ditentukan oleh kolom (debit atau credit), bukan tanda

### 5. Journal Entry Structure — Header + Lines

Jurnal menggunakan struktur **header-detail** (1 header : N lines):

```
jurnal (header)
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
│
├── ── Identitas ──
├── journal_number        VARCHAR UNIQUE per tenant
│                         Format: JRN-{YYYY}-{MM}-{SEQ:8}
├── journal_date          DATE NOT NULL
├── description           TEXT NOT NULL
│
├── ── Referensi Sumber ──
├── source_type           ENUM (transaction, manual, closing, adjustment, shu_distribution)
├── source_id             UUID (nullable, FK → transaksi/shu_periode/etc)
│
├── ── Posting ──
├── status                ENUM (unposted, posted, reversed)
├── posted_at             TIMESTAMPTZ (nullable)
├── posted_by             UUID (nullable, FK → user)
│
├── ── Periode ──
├── period_id             UUID (FK → accounting_period) NOT NULL
│
├── ── Totals (denormalized for quick validation) ──
├── total_debit           NUMERIC(15,2) NOT NULL
├── total_credit          NUMERIC(15,2) NOT NULL
│                         CHECK (total_debit = total_credit)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

```
jurnal_line (detail)
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── jurnal_id             UUID (FK → jurnal) NOT NULL
│
├── ── Akun ──
├── coa_id                UUID (FK → coa) NOT NULL
├── line_number           INTEGER NOT NULL (urutan baris, 1-based)
│
├── ── Amount ──
├── debit                 NUMERIC(15,2) NOT NULL DEFAULT 0
├── credit                NUMERIC(15,2) NOT NULL DEFAULT 0
│                         CHECK (debit >= 0 AND credit >= 0)
│                         CHECK (debit > 0 OR credit > 0)
│                         CHECK (NOT (debit > 0 AND credit > 0))
│
├── ── Keterangan ──
├── description           TEXT (nullable, keterangan per baris)
│
├── ── Referensi ──
├── rekening_id           UUID (nullable, FK → rekening, jika terkait rekening tertentu)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    └── created_by        UUID
```

**Aturan jurnal_line:**
- Setiap baris hanya boleh terisi **debit ATAU credit** — tidak boleh keduanya (CHECK constraint)
- Minimal **2 lines** per jurnal (satu debit, satu credit)
- `line_number` untuk menjaga urutan tampilan
- `rekening_id` opsional — diisi jika line terkait rekening spesifik (misal: tabungan nasabah tertentu). Berguna untuk rekonsiliasi per-rekening.

### 6. Posting — Unposted → Posted (Irreversible)

Jurnal dibuat dalam status **UNPOSTED** dan harus di-post sebelum masuk ke laporan keuangan.

```
Auto-journal dari transaksi
        │
        v
┌──────────────────┐
│ Status: UNPOSTED │  Bisa di-edit / di-hapus
└────────┬─────────┘
         │ post (by Supervisor+)
         v
┌──────────────────┐
│ Status: POSTED   │  Immutable, masuk ke laporan
└────────┬─────────┘
         │ (jika ada koreksi)
         v
┌──────────────────┐
│ Buat jurnal      │  Jurnal pembalik (reversal)
│ REVERSAL baru    │  bukan edit jurnal asli
└──────────────────┘
```

**Aturan posting:**
- Jurnal UNPOSTED **tidak masuk** ke trial balance, neraca, dan laporan lainnya — hanya jurnal POSTED yang dihitung
- Posting bersifat **irreversible** — jurnal yang sudah POSTED tidak bisa di-edit atau di-hapus
- Koreksi dilakukan dengan membuat **jurnal reversal** (jurnal baru dengan amount terbalik) + jurnal koreksi yang benar
- Jurnal reversal otomatis mereferensikan jurnal asli via `source_id`
- **Auto-post option**: tenant bisa mengaktifkan auto-post untuk jurnal dari transaksi operasional — jurnal langsung POSTED tanpa review manual. Default: auto-post **ON** (karena jurnal dari auto-journal sudah tervalidasi)
- Jurnal manual (adjustment) **selalu UNPOSTED** — membutuhkan review dan posting manual

### 7. Period Management — Accounting Periods

Periode akuntansi mengelompokkan jurnal dan menentukan kapan jurnal bisa di-posting.

```
accounting_period
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Identitas ──
├── period_type           ENUM (monthly, annual)
├── year                  INTEGER NOT NULL
├── month                 INTEGER (1-12, nullable untuk annual)
├── period_name           VARCHAR (auto: "Januari 2026", "Tahun Buku 2026")
│
├── ── Status ──
├── status                ENUM (open, closed, locked)
├── opened_at             TIMESTAMPTZ NOT NULL
├── closed_at             TIMESTAMPTZ (nullable)
├── closed_by             UUID (nullable, FK → user)
│
├── ── Periode ──
├── start_date            DATE NOT NULL
├── end_date              DATE NOT NULL
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Status lifecycle:**

```
┌─────────────┐
│ Status: OPEN│  Jurnal bisa di-post ke periode ini
└──────┬──────┘
       │ close period (Manager+)
       v
┌──────────────┐
│Status: CLOSED│  Tidak ada jurnal baru, tapi bisa reopen
└──────┬───────┘
       │ lock period (Admin)
       v
┌──────────────┐
│Status: LOCKED│  Permanen, tidak bisa reopen
└──────────────┘
```

**Aturan:**
- Periode **monthly** di-auto-create saat awal bulan (atau saat transaksi pertama di bulan tersebut)
- Periode **annual** di-create manual saat proses year-end closing
- Posting ke periode **CLOSED** atau **LOCKED** diblokir — jurnal harus masuk ke periode yang OPEN
- **Reopen** periode CLOSED dimungkinkan oleh Admin — untuk koreksi akhir bulan. Periode LOCKED tidak bisa reopen.
- Menutup periode bulanan memerlukan minimal role **Manager**
- Mengunci (lock) periode memerlukan role **Admin** — biasanya setelah audit selesai
- Periode harus ditutup **berurutan** — tidak bisa close Maret jika Februari masih OPEN

### 8. Trial Balance — Auto-generate

Trial balance di-generate otomatis dari jurnal yang sudah POSTED dalam satu periode.

```
Trial Balance generation:
1. Filter: jurnal.status = 'posted' AND jurnal.period_id = {target_period}
2. Group by: coa_id
3. Calculate:
   - SUM(debit) per akun
   - SUM(credit) per akun
   - Saldo = SUM(debit) - SUM(credit)
     (positif untuk akun Aset/Beban, negatif untuk Kewajiban/Ekuitas/Pendapatan)
4. Validate: Total SUM(debit) = Total SUM(credit)
```

**Aturan:**
- Trial balance adalah **view/query**, bukan tabel terpisah — selalu real-time dari data jurnal POSTED
- Jika total debit != total kredit di trial balance, sistem menampilkan **alert** — menandakan ada inkonsistensi (seharusnya tidak terjadi jika enforcement benar)
- Trial balance bisa di-generate untuk **satu periode** atau **kumulatif** (awal tahun sampai periode tertentu)
- Format output: kode akun, nama akun, saldo awal, total debit periode, total kredit periode, saldo akhir

### 9. Financial Statements — Laporan Keuangan

Sistem menghasilkan 3 laporan keuangan utama dari data jurnal POSTED:

**A. Neraca (Balance Sheet) — Posisi Keuangan**

```
NERACA per {tanggal}
──────────────────────────────────────
ASET                        │ KEWAJIBAN & EKUITAS
─────────────────────────── │ ──────────────────────────
Aset Lancar:                │ Kewajiban:
  Kas           xxx         │   Simpanan Nasabah    xxx
  Bank          xxx         │   Hutang              xxx
  Piutang       xxx         │
                            │ Ekuitas:
Aset Tetap:                 │   Simpanan Pokok      xxx
  Inventaris    xxx         │   Simpanan Wajib      xxx
  (Akum. Peny) (xxx)       │   Cadangan            xxx
                            │   SHU                 xxx
──────────────────────────── │ ──────────────────────────
TOTAL ASET      xxx         │ TOTAL K+E             xxx
                            │ (harus = Total Aset)
```

**B. Laporan Laba Rugi (Income Statement) — Hasil Usaha**

```
LAPORAN LABA RUGI
Periode: {bulan/tahun}
──────────────────────────────────────
PENDAPATAN:
  Pendapatan Bunga/Margin       xxx
  Pendapatan Administrasi       xxx
  Pendapatan Denda              xxx
  Pendapatan Lain-lain          xxx
                          ──────────
  Total Pendapatan              xxx

BEBAN:
  Beban Bunga/Bagi Hasil        xxx
  Beban Operasional             xxx
  Beban Cadangan Kerugian       xxx
  Beban Lain-lain               xxx
                          ──────────
  Total Beban                   xxx
                          ══════════
  SHU (Sisa Hasil Usaha)       xxx
```

**C. Laporan Arus Kas (Cash Flow Statement)**

```
LAPORAN ARUS KAS
Periode: {bulan/tahun}
──────────────────────────────────────
Arus Kas dari Aktivitas Operasi:
  SHU                                xxx
  Penyesuaian:
    Penyusutan                       xxx
    Perubahan Piutang               (xxx)
    Perubahan Simpanan Nasabah       xxx
                              ──────────
  Arus Kas Operasi Bersih           xxx

Arus Kas dari Aktivitas Investasi:
  Pembelian Aktiva Tetap            (xxx)
                              ──────────
  Arus Kas Investasi Bersih         (xxx)

Arus Kas dari Aktivitas Pendanaan:
  Perubahan Simpanan Pokok           xxx
  Perubahan Simpanan Wajib           xxx
                              ──────────
  Arus Kas Pendanaan Bersih          xxx
                              ══════════
  Kenaikan/(Penurunan) Kas Bersih   xxx
  Kas Awal Periode                   xxx
  Kas Akhir Periode                  xxx
```

**Aturan:**
- Semua laporan dihasilkan dari **data jurnal POSTED** — bukan dari tabel terpisah
- Laporan bisa di-generate untuk **periode apapun** (bulanan, triwulanan, semesteran, tahunan)
- Format output: PDF (formal), Excel (analisis), on-screen (dashboard)
- Neraca harus selalu **balance** (Aset = Kewajiban + Ekuitas) — jika tidak, sistem menampilkan warning

**Islamic Mode — Neraca Tambahan Section:**

Untuk `coop_type = "islamic"`, Neraca memiliki section tambahan per PSAK 109:

```
ASET = KEWAJIBAN + EKUITAS + DANA ZIS
```

Dana ZIS (Zakat, Infaq, Shadaqah) & Dana Ta'zir disajikan sebagai section terpisah di Neraca, bukan sebagai kewajiban maupun ekuitas. Akun 6xxx-8xxx yang memiliki saldo disajikan di section ini.

General mode: Neraca standar (Aset = Kewajiban + Ekuitas), tanpa section Dana ZIS.

### 10. Islamic Accounting Specifics — Laporan Tambahan Syariah

BMT mode memiliki **laporan tambahan** yang terpisah dari laporan laba rugi koperasi, sesuai PSAK Syariah:

```
LAPORAN KHUSUS BMT (coop_type = "islamic"):
─────────────────────────────────────────────
A. Laporan Sumber dan Penyaluran Dana Zakat
   - Saldo awal dana zakat
   - Penerimaan: zakat maal, zakat fitrah, zakat institusi
   - Penyaluran: per asnaf (8 kategori mustahik)
   - Saldo akhir dana zakat

B. Laporan Dana Kebajikan (Qardhul Hasan)
   - Saldo awal dana kebajikan
   - Penerimaan: infaq anggota, pengembalian qardh
   - Penyaluran: pinjaman qardh, sumbangan
   - Saldo akhir dana kebajikan

C. Laporan Dana Ta'zir (Social Fund)
   - Saldo awal dana ta'zir
   - Penerimaan: denda keterlambatan anggota
   - Penyaluran: bantuan sosial, kegiatan sosial
   - Saldo akhir dana ta'zir
```

**Aturan:**
- Dana zakat, kebajikan, dan ta'zir adalah **dana titipan** — bukan pendapatan koperasi
- Akun 6xxx, 7xxx, 8xxx **tidak masuk** ke Laporan Laba Rugi koperasi — terpisah di laporan khusus
- Saldo dana ini disajikan di Neraca sebagai **section terpisah (Dana ZIS)** — bukan kewajiban, bukan ekuitas (per PSAK 109 paragraf 35-36)
- Setiap penerimaan dan penyaluran dana syariah menghasilkan jurnal entry yang sama ketatnya dengan transaksi biasa (double-entry, posting, dll)
- Laporan ini **wajib** untuk BMT — merupakan bagian dari compliance PSAK Syariah

### 11. Year-End Closing — Penutupan Tahun Buku

Proses penutupan akhir tahun memindahkan saldo akun pendapatan dan beban ke SHU.

```
Year-End Closing Flow:
        │
        v
┌──────────────────────────────────────┐
│ STEP 1: Pre-check                    │
│ - Semua periode bulanan (Jan-Des)    │
│   harus berstatus CLOSED             │
│ - Tidak ada jurnal UNPOSTED          │
│   di tahun tersebut                  │
└────────┬─────────────────────────────┘
         │ pre-check pass
         v
┌──────────────────────────────────────┐
│ STEP 2: Generate closing journals    │
│ - Debit semua akun PENDAPATAN (4xxx) │
│   → saldo jadi 0                     │
│ - Credit semua akun BEBAN (5xxx)     │
│   → saldo jadi 0                     │
│ - Selisih masuk ke 3500 SHU Tahun    │
│   Berjalan                           │
└────────┬─────────────────────────────┘
         │
         v
┌──────────────────────────────────────┐
│ STEP 3: Create annual period         │
│ - Buat periode annual                │
│ - Posting closing journals           │
│ - Lock semua periode bulanan         │
└────────┬─────────────────────────────┘
         │
         v
┌──────────────────────────────────────┐
│ STEP 4: Carry forward                │
│ - Saldo akun 1xxx-3xxx dibawa ke     │
│   tahun buku baru sebagai saldo awal │
│ - Akun 4xxx-5xxx mulai dari 0        │
│ - [Islamic] Saldo 6xxx-8xxx dibawa   │
│   sebagai saldo awal (dana titipan)  │
└──────────────────────────────────────┘
```

**Aturan:**
- Year-end closing hanya bisa dilakukan oleh **Admin**
- Semua 12 periode bulanan harus **CLOSED** sebelum year-end closing bisa dijalankan
- Semua jurnal di tahun tersebut harus sudah **POSTED** — tidak boleh ada unposted journal
- Closing menghasilkan jurnal otomatis yang memindahkan saldo P&L ke akun SHU (3500)
- Setelah closing, semua periode bulanan di-**LOCK** (tidak bisa reopen)
- Saldo neraca (1xxx-3xxx) dibawa sebagai **opening balance** tahun buku baru
- Proses ini menjadi dasar untuk perhitungan SHU di [ADR-K016](./ADR-K016-shu.md)
- **Irreversible** — jika ada kesalahan, harus buat jurnal adjustment di tahun buku baru

### 12. COA Data Model

```
coa (Chart of Accounts)
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Identitas ──
├── account_code          VARCHAR NOT NULL (misal: "1101")
├── account_name          VARCHAR NOT NULL (misal: "Kas Teller")
├── account_name_en       VARCHAR (nullable, nama English untuk referensi)
│
├── ── Hirarki ──
├── parent_id             UUID (nullable, FK → coa, self-reference)
├── level                 INTEGER NOT NULL (1 = top, 2 = child, dst)
│
├── ── Klasifikasi ──
├── account_type          ENUM (asset, liability, equity, revenue, expense,
│                               zakat, kebajikan, tazir)
├── normal_balance        ENUM (debit, credit)
│                         (asset/expense = debit, lain = credit)
├── is_header             BOOLEAN DEFAULT false
│                         (true = akun induk, tidak bisa di-posting)
│
├── ── Status ──
├── is_system             BOOLEAN DEFAULT false
│                         (true = system account, tidak bisa hapus/ubah kode)
├── is_active             BOOLEAN DEFAULT true
│
├── ── Coop Type ──
├── coop_type_required    ENUM (general, islamic, both) DEFAULT 'both'
│                         (6xxx-8xxx = islamic only)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Unique constraint:** `(tenant_id, account_code)` — kode akun unik per tenant.

### 13. Journal Mapping Data Model

```
journal_mapping
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Trigger ──
├── transaction_type      VARCHAR NOT NULL
│                         (misal: "deposit_cash", "withdrawal_cash",
│                          "loan_disbursement", "installment_payment",
│                          "zakat_collection", "tazir_collection")
├── coop_type             ENUM (general, islamic, both) DEFAULT 'both'
│
├── ── Mapping Rules ──
├── rules                 JSONB NOT NULL
│   [
│     { "line": 1, "coa_code": "1101", "side": "debit",  "amount_field": "amount" },
│     { "line": 2, "coa_code": "2101", "side": "credit", "amount_field": "amount" }
│   ]
│
├── ── Status ──
├── is_system             BOOLEAN DEFAULT false
├── is_active             BOOLEAN DEFAULT true
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Aturan mapping:**
- `rules` berisi array of lines — setiap line menunjuk ke `coa_code` dan `side` (debit/credit)
- `amount_field` menunjuk ke field dari transaksi yang menjadi nominal (misal: `amount`, `principal`, `interest`)
- System mapping tidak bisa dihapus — hanya bisa di-override (duplikasi dengan is_system = false)
- Validation saat save: SUM debit fields harus = SUM credit fields secara logis

### 14. Dual-Mode — Perbedaan COA dan Mapping

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|-------|------------------------|------------------------|
| COA range | 1xxx - 5xxx | 1xxx - 8xxx (+ zakat, kebajikan, ta'zir) |
| Akun 1302 | Piutang Bunga Pinjaman | Piutang Margin Murabahah |
| Akun 4100 | Pendapatan Bunga Pinjaman | Pendapatan Margin/Bagi Hasil |
| Akun 5100 | Beban Bunga Simpanan | Beban Bagi Hasil Simpanan |
| Denda mapping | tx → 4300 Pendapatan Denda | tx → 8101 Penerimaan Ta'zir |
| Laporan tambahan | Tidak ada | Lap. Zakat, Kebajikan, Ta'zir |
| Year-end closing | 4xxx-5xxx → 3500 SHU | 4xxx-5xxx → 3500 SHU + carry 6xxx-8xxx |

**Aturan:**
- COA template yang di-seed saat onboarding **berbeda** per `coop_type`
- Journal mapping yang di-seed saat onboarding **berbeda** per `coop_type`
- Mapping dengan `coop_type = "both"` dipakai di kedua mode
- UI harus menyembunyikan akun 6xxx-8xxx di mode general — bukan hanya disable, tapi **hidden**
- Strategy pattern digunakan untuk memilih laporan yang di-generate (general vs islamic)

### 15. Vernon _rels dan _data Structure

Jurnal menggunakan Vernon pattern karena listing jurnal membutuhkan data branch dan source transaction (menghindari JOIN).

**_rels (jurnal):**
```json
{
  "tenant_id":  "018f...",
  "branch_id":  "018f...",
  "period_id":  "018f...",
  "source_id":  "018f..."
}
```

**_data (jurnal):**
```json
{
  "branch": {
    "id":   "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  },
  "period": {
    "id":          "018f...",
    "period_name": "Januari 2026",
    "year":        2026,
    "month":       1
  },
  "source": {
    "type":             "transaction",
    "id":               "018f...",
    "transaction_code": "TRX-2026-01-00000001",
    "description":      "Setoran tunai tabungan"
  }
}
```

**SyncEngine triggers:**
- `BranchUpdatedEvent` → update `_data.branch` di semua jurnal cabang tersebut
- `AccountingPeriodUpdatedEvent` → update `_data.period` di semua jurnal periode tersebut

### 16. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View COA | v | v | v | v |
| Add sub-akun (5+ digit) | - | - | v | v |
| Edit nama akun | - | - | v | v |
| Nonaktifkan akun | - | - | - | v |
| View jurnal | v | v | v | v |
| Create jurnal manual (adjustment) | - | v | v | v |
| Post jurnal | - | v | v | v |
| Reverse jurnal | - | - | v | v |
| Edit journal mapping | - | - | - | v |
| Close periode bulanan | - | - | v | v |
| Reopen periode bulanan | - | - | - | v |
| Lock periode | - | - | - | v |
| Year-end closing | - | - | - | v |
| Generate laporan keuangan | v | v | v | v |
| View trial balance | - | v | v | v |

**Catatan:**
- Teller bisa **view** jurnal dan generate laporan, tapi tidak bisa membuat atau mengubah jurnal
- Jurnal manual (adjustment) membutuhkan minimal **Supervisor** — dan tetap harus di-post terpisah
- Reverse jurnal membutuhkan **Manager+** karena berdampak pada laporan yang sudah terbit
- Year-end closing dan lock periode hanya **Admin** — operasi yang berdampak permanen

### 17. Audit Trail

Semua operasi akuntansi tercatat di audit log:

```
Operasi yang di-audit:
├── COA: create, update nama, nonaktifkan, tambah sub-akun
├── Jurnal: create, post, reverse
├── Jurnal manual: create, edit (sebelum post), post
├── Mapping: create, update, override
├── Periode: open, close, reopen, lock
├── Year-end closing: execute (dengan snapshot SHU)
└── Laporan: generate (siapa, kapan, periode apa)
```

**Aturan:**
- Audit log **immutable** — tidak bisa dihapus atau diubah
- Setiap entry berisi: who (user_id), what (operation), when (timestamp), detail (before/after snapshot)
- Audit log digunakan untuk compliance OJK dan audit internal koperasi

## Consequences

### Positif

- **Terintegrasi** — jurnal otomatis dari transaksi, tidak perlu input manual untuk operasional sehari-hari
- **Akurat** — double-entry enforcement di 3 level mencegah ketidakseimbangan
- **Compliant** — COA mengikuti standar PSAK/SAK ETAP (general) dan PSAK Syariah (islamic)
- **Flexible** — tenant bisa customize COA (sub-akun) dan journal mapping tanpa mengubah sistem
- **Auditable** — semua operasi akuntansi punya audit trail, posting irreversible, periode bisa dikunci
- **Dual-mode ready** — COA template dan mapping otomatis berbeda per `coop_type`
- **Real-time** — trial balance dan laporan keuangan selalu up-to-date dari jurnal POSTED

### Negatif

- **Complexity** — mesin akuntansi double-entry menambah complexity di transaction layer
- **Performance** — auto-journal menambah overhead per transaksi (insert jurnal header + N lines)
- **Storage** — setiap transaksi menghasilkan minimal 3 row (1 header + 2 lines), volume tinggi
- **Year-end coupling** — year-end closing harus selesai sebelum SHU bisa dihitung (K016)
- **Mapping maintenance** — jika tenant menambah jenis transaksi custom, mapping harus di-maintain

### Mitigasi

- Auto-journal dilakukan dalam **satu database transaction** dengan transaksi sumber — jika jurnal gagal, transaksi juga gagal. Tidak ada inkonsistensi.
- Jurnal lines menggunakan **batch insert** — performance acceptable untuk volume koperasi sekolah (ratusan transaksi per hari, bukan jutaan)
- **Partitioning** per periode bisa diterapkan di masa depan jika volume storage menjadi masalah
- Year-end closing memiliki **pre-check** yang jelas — tidak bisa dijalankan jika prerequisites belum terpenuhi
- Default mapping sudah mencakup semua jenis transaksi standar — tenant hanya perlu maintain mapping custom

## Alternatives Considered

### A. Single-Entry Bookkeeping

Mencatat transaksi hanya sebagai debit atau kredit di satu akun, tanpa pasangan entry.

**Ditolak** karena: tidak memenuhi standar akuntansi (PSAK/SAK ETAP), tidak bisa menghasilkan neraca yang balance, dan tidak bisa diaudit oleh Dinas Koperasi/OJK. Double-entry adalah **requirement**, bukan pilihan.

### B. Jurnal Manual (tanpa auto-journal)

Operator harus input jurnal secara manual untuk setiap transaksi.

**Ditolak** karena: rawan human error (salah akun, salah nominal), lambat, dan tidak scalable. Auto-journal dari mapping memastikan konsistensi dan mengurangi beban operator. Jurnal manual tetap tersedia untuk transaksi adjustment yang tidak bisa di-automate.

### C. Langsung POSTED (tanpa status UNPOSTED)

Semua jurnal langsung masuk ke laporan tanpa proses posting.

**Ditolak** karena: tidak ada kesempatan untuk review jurnal manual sebelum masuk ke laporan. Untuk jurnal dari auto-journal, fitur auto-post tersedia sebagai kompromi — operator bisa memilih. Jurnal manual tetap harus melalui proses posting untuk kontrol.

### D. COA Global (sama untuk semua tenant)

Satu COA untuk semua tenant, tanpa customization.

**Ditolak** karena: setiap koperasi memiliki kebutuhan yang sedikit berbeda. Misalnya, koperasi dengan usaha toko perlu akun persediaan barang, sementara koperasi simpan pinjam murni tidak. Template default + customization memberikan keseimbangan antara standarisasi dan fleksibilitas.

### E. Period-less Accounting (tanpa periode)

Tidak menggunakan periode — semua jurnal masuk ke satu "wadah" global.

**Ditolak** karena: koperasi wajib melaporkan keuangan per periode (bulanan ke OJK, tahunan ke Dinas Koperasi). Tanpa periode, penutupan buku dan carry forward tidak bisa dilakukan. Periode juga melindungi data historis dari modifikasi setelah laporan diterbitkan.
