# ADR-K006: Deposito / Simpanan Berjangka

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Deposito (Simpanan Berjangka) adalah produk simpanan dengan **jangka waktu tetap** (tenor) dan **rate yang lebih tinggi** dibandingkan tabungan biasa. Dana ditempatkan untuk periode tertentu dan tidak bisa ditarik sebelum jatuh tempo tanpa penalti.

Di koperasi sekolah, deposito biasanya digunakan oleh:
- **Guru/staff**: punya dana lebih, ingin return yang lebih tinggi dari tabungan biasa
- **Orang tua/wali**: investasi jangka menengah untuk biaya pendidikan anak
- **External member**: masyarakat yang ingin mendapatkan return kompetitif dari koperasi

Tantangan yang harus diakomodasi:
- **Multi-tenor**: 1, 3, 6, 12, 24 bulan — masing-masing dengan rate berbeda
- **Tiered rate**: rate lebih tinggi untuk tenor lebih panjang dan nominal lebih besar
- **Maturity handling**: apa yang terjadi saat jatuh tempo — perpanjangan otomatis, cairkan, atau pindah ke tabungan
- **Dual-mode**: general menggunakan bunga tetap, BMT menggunakan nisbah bagi hasil indikatif
- **Deposito as collateral**: deposito bisa dijadikan jaminan pinjaman (K010)

Setiap deposito dibuka berdasarkan **produk** dari K003 dengan kategori DEPOSITO, dan menggunakan **rekening** dari K002.

## Decision

### 1. Tenor Options

Tenor deposito configurable per produk:

```
tenor_options (dari produk K003):
├── 1 bulan      Jangka sangat pendek
├── 3 bulan      Jangka pendek
├── 6 bulan      Jangka menengah
├── 12 bulan     Jangka panjang (paling populer)
└── 24 bulan     Jangka sangat panjang
```

**Aturan:**
- Tenor tersimpan di produk sebagai array: `tenor_options: [1, 3, 6, 12, 24]`
- Tenant bisa membatasi pilihan tenor per produk — misal hanya `[3, 6, 12]`
- Nasabah memilih tenor saat pengajuan deposito (K002 Section 3)
- `maturity_date` dihitung otomatis: `placement_date + tenor_months`
- Tenor minimum dan maksimum di-enforce dari produk (K003 Section 8)
- Satu produk deposito bisa mencakup **beberapa tenor** — atau tenant bisa buat produk terpisah per tenor

### 2. Rate Tiers — Higher Rate for Longer Tenor & Larger Amount

Rate deposito menggunakan **tiered structure** berdasarkan tenor dan nominal:

#### 2a. General Mode — Suku Bunga Berjenjang

```
rate_tiers (JSONB di produk_version):
[
  { "tenor": 1,  "min_amount": 1000000,   "max_amount": 9999999,   "rate": 4.0000 },
  { "tenor": 1,  "min_amount": 10000000,  "max_amount": 49999999,  "rate": 4.5000 },
  { "tenor": 1,  "min_amount": 50000000,  "max_amount": null,      "rate": 5.0000 },
  { "tenor": 3,  "min_amount": 1000000,   "max_amount": 9999999,   "rate": 5.0000 },
  { "tenor": 3,  "min_amount": 10000000,  "max_amount": 49999999,  "rate": 5.5000 },
  { "tenor": 3,  "min_amount": 50000000,  "max_amount": null,      "rate": 6.0000 },
  { "tenor": 6,  "min_amount": 1000000,   "max_amount": 9999999,   "rate": 6.0000 },
  { "tenor": 6,  "min_amount": 10000000,  "max_amount": 49999999,  "rate": 6.5000 },
  { "tenor": 6,  "min_amount": 50000000,  "max_amount": null,      "rate": 7.0000 },
  { "tenor": 12, "min_amount": 1000000,   "max_amount": 9999999,   "rate": 7.0000 },
  { "tenor": 12, "min_amount": 10000000,  "max_amount": 49999999,  "rate": 7.5000 },
  { "tenor": 12, "min_amount": 50000000,  "max_amount": null,      "rate": 8.0000 },
  { "tenor": 24, "min_amount": 1000000,   "max_amount": 9999999,   "rate": 8.0000 },
  { "tenor": 24, "min_amount": 10000000,  "max_amount": 49999999,  "rate": 8.5000 },
  { "tenor": 24, "min_amount": 50000000,  "max_amount": null,      "rate": 9.0000 }
]
```

**Tabel ringkasan:**

| Tenor | < 10 jt | 10 - 50 jt | > 50 jt |
|-------|---------|------------|---------|
| 1 bulan | 4.0% | 4.5% | 5.0% |
| 3 bulan | 5.0% | 5.5% | 6.0% |
| 6 bulan | 6.0% | 6.5% | 7.0% |
| 12 bulan | 7.0% | 7.5% | 8.0% |
| 24 bulan | 8.0% | 8.5% | 9.0% |

#### 2b. Islamic Mode — Nisbah Berjenjang

```
nisbah_tiers (JSONB di produk_version):
[
  { "tenor": 1,  "min_amount": 1000000,   "nisbah_nasabah": 30.00, "nisbah_koperasi": 70.00 },
  { "tenor": 3,  "min_amount": 1000000,   "nisbah_nasabah": 35.00, "nisbah_koperasi": 65.00 },
  { "tenor": 6,  "min_amount": 1000000,   "nisbah_nasabah": 40.00, "nisbah_koperasi": 60.00 },
  { "tenor": 12, "min_amount": 1000000,   "nisbah_nasabah": 45.00, "nisbah_koperasi": 55.00 },
  { "tenor": 12, "min_amount": 50000000,  "nisbah_nasabah": 50.00, "nisbah_koperasi": 50.00 },
  { "tenor": 24, "min_amount": 1000000,   "nisbah_nasabah": 50.00, "nisbah_koperasi": 50.00 },
  { "tenor": 24, "min_amount": 50000000,  "nisbah_nasabah": 55.00, "nisbah_koperasi": 45.00 }
]
```

**Aturan rate tiers:**
- Rate/nisbah ditentukan saat **placement** berdasarkan tenor dan nominal yang dipilih
- Rate **locked** selama tenor deposito — tidak berubah meskipun produk update rate baru
- Lookup tier: match tenor + nominal → ambil rate dari tier yang sesuai
- Jika nominal tepat di batas `min_amount`, gunakan tier yang lebih tinggi
- `max_amount: null` berarti unlimited — catch-all untuk nominal besar
- Rate tiers mengikuti **product versioning** (K003 Section 6) — perubahan tier menghasilkan version baru

### 3. Minimum Placement Amount

Nominal minimum penempatan per produk:

```
min_placement (dari produk K003):
  default: 1000000     # Rp 1.000.000 (minimum standar)
  per_product:
    Deposito_3_Bulan:  1000000
    Deposito_12_Bulan: 5000000     # Lebih tinggi untuk tenor panjang (opsional)
    Deposito_24_Bulan: 10000000
```

**Aturan:**
- `min_placement` di-enforce saat **pengajuan deposito** (K002 Section 3, submit pre-check)
- Tidak ada `max_placement` standar — tapi bisa dikonfigurasi jika tenant membatasinya
- Nominal harus **bulat** (kelipatan Rp 100.000 — configurable `placement_rounding`)
- Nominal tidak bisa diubah setelah placement — untuk menambah, buka deposito baru

### 4. Interest / Profit Calculation

#### 4a. General Mode — Bunga Tetap (Fixed Rate)

```
Contoh: Deposito Rp 50.000.000, Tenor 12 bulan, Rate 8% p.a.

Opsi 1: Bunga dibayar BULANAN
  Bunga/bulan = 50.000.000 × (8/100) × (1/12) = Rp 333.333
  PPh 10% = Rp 33.333
  Bunga bersih/bulan = Rp 300.000
  → Diposting ke tabungan nasabah setiap bulan

Opsi 2: Bunga dibayar SAAT JATUH TEMPO
  Total bunga = 50.000.000 × (8/100) × (12/12) = Rp 4.000.000
  PPh 10% = Rp 400.000
  Bunga bersih = Rp 3.600.000
  → Dibayarkan bersama pokok saat jatuh tempo
```

**Interest payment options:**
```
interest_payment_method:
├── monthly             Bunga dibayar setiap bulan ke tabungan nasabah
├── at_maturity         Bunga dibayar sekaligus saat jatuh tempo
└── capitalize          Bunga ditambahkan ke pokok (compound) — dibayar saat jatuh tempo
```

#### 4b. Islamic Mode — Bagi Hasil Deposito Mudharabah

```
Contoh: Deposito Mudharabah Rp 50.000.000, Tenor 12 bulan
Nisbah: 45:55 (Nasabah:Koperasi)
Indicative rate: 7.5% p.a.

Bulan ke-1:
1. Total dana deposito Mudharabah koperasi = Rp 2.000.000.000
2. Porsi nasabah A = 50.000.000 / 2.000.000.000 = 2.5%
3. Profit koperasi dari penyaluran bulan ini = Rp 30.000.000
4. Bagian nasabah total = 30.000.000 × 45% = Rp 13.500.000
5. Bagi hasil nasabah A = 13.500.000 × 2.5% = Rp 337.500
   PPh 10% = Rp 33.750
   Bagi hasil bersih = Rp 303.750

Catatan: Bagi hasil aktual bisa LEBIH atau KURANG dari indicative rate
         tergantung performa koperasi bulan tersebut.
```

**Aturan:**
- General: rate **locked** saat placement — tidak berubah selama tenor
- Islamic: nisbah **locked** saat placement, tapi bagi hasil aktual **bervariasi** per bulan tergantung profit koperasi
- `indicative_rate` pada BMT mode hanyalah **estimasi** — disclaimer wajib ditampilkan
- PPh atas bunga/bagi hasil: rate configurable per tenant (default: 10%)
- Scheduled job menghitung dan memposting bunga/bagi hasil sesuai `interest_payment_method`
- Bunga/bagi hasil monthly diposting ke **tabungan nasabah** (membutuhkan rekening tabungan aktif)
- Jika nasabah tidak punya tabungan → bunga/bagi hasil di-capitalize (ditambah ke pokok)

### 5. Maturity Handling

Saat deposito mendekati dan mencapai jatuh tempo:

```
Notification schedule (configurable per tenant):
├── maturity_date - 30 hari   Reminder pertama
├── maturity_date - 7 hari    Reminder kedua
├── maturity_date - 1 hari    Reminder terakhir
└── maturity_date              Jatuh tempo → proses sesuai instruksi
```

**Rollover options (dipilih saat pengajuan, bisa diubah sebelum jatuh tempo):**

```
rollover_instruction:
├── NONE                 Cairkan semua (pokok + bunga) ke tabungan
├── PRINCIPAL_ONLY       Perpanjang pokok, bunga ke tabungan
├── PRINCIPAL_AND_PROFIT Perpanjang pokok + bunga (compound)
└── TO_TABUNGAN          Cairkan semua ke tabungan tertentu
```

**Maturity processing flow:**
```
Maturity date (scheduled job cek harian)
        │
        v
┌───────────────────────────────────────┐
│ Untuk setiap deposito jatuh tempo:    │
│                                       │
│ NONE / TO_TABUNGAN:                   │
│ 1. Hitung bunga/bagi hasil terakhir   │
│ 2. Total = pokok + bunga outstanding  │
│ 3. Transfer total ke tabungan nasabah │
│ 4. Saldo deposito → 0                │
│ 5. Status rekening deposito → CLOSED  │
│ 6. Kirim notifikasi pencairan         │
│                                       │
│ PRINCIPAL_ONLY:                       │
│ 1. Hitung bunga/bagi hasil terakhir   │
│ 2. Transfer bunga ke tabungan         │
│ 3. Buat deposito baru (auto-rollover) │
│    - Nominal = pokok asli             │
│    - Tenor = tenor yang sama          │
│    - Rate = rate TERBARU (bukan lama) │
│ 4. Close deposito lama                │
│ 5. Kirim notifikasi perpanjangan      │
│                                       │
│ PRINCIPAL_AND_PROFIT:                 │
│ 1. Hitung bunga/bagi hasil terakhir   │
│ 2. Buat deposito baru (auto-rollover) │
│    - Nominal = pokok + semua bunga    │
│    - Tenor = tenor yang sama          │
│    - Rate = rate TERBARU              │
│ 3. Close deposito lama                │
│ 4. Kirim notifikasi perpanjangan      │
└───────────────────────────────────────┘
```

**Aturan:**
- Rollover instruction di-set saat pengajuan — **bisa diubah** sebelum jatuh tempo oleh nasabah (via teller)
- Auto-rollover menggunakan **rate terbaru** dari produk (bukan rate deposito lama) — rate terbaru dari product version yang berlaku
- Deposito baru hasil rollover mendapat **nomor rekening baru** dan **bilyet baru**
- Jika rollover dan produk sudah DISCONTINUED → cairkan ke tabungan (fallback to NONE)
- Jika nasabah tidak punya tabungan saat maturity → hold dana, kirim notifikasi, teller proses manual
- Notifikasi maturity dikirim via channel yang dikonfigurasi (in_app, SMS)
- Maturity processing di-run oleh **scheduled job harian** (pagi hari, sebelum jam operasional)

### 6. Early Withdrawal — Pencairan Sebelum Jatuh Tempo

```
Nasabah mengajukan pencairan dini
        │
        v
┌───────────────────────────────────────┐
│ Pre-check:                            │
│ 1. Deposito status = ACTIVE           │
│ 2. Tidak sedang di-hold sebagai       │
│    collateral (K010)                  │
│ 3. Approval: Manager+                 │
└───────────────────────────────────────┘
        │ approved
        v
┌───────────────────────────────────────┐
│ Penalty calculation:                  │
│                                       │
│ [General mode]                        │
│ penalty = earned_interest × penalty%  │
│ contoh: bunga yang sudah didapat      │
│   Rp 2.000.000 × 50% = Rp 1.000.000 │
│ net_payout = pokok + bunga - penalty  │
│                                       │
│ [Islamic mode]                        │
│ penalty = 0 (bagi hasil yang sudah    │
│   dibayar tidak ditarik kembali)      │
│ Tapi: bagi hasil bulan berjalan      │
│   tidak dibayarkan (forfeit)          │
└───────────────────────────────────────┘
        │
        v
┌───────────────────────────────────────┐
│ Processing:                           │
│ 1. Hitung net payout                  │
│ 2. Transfer ke tabungan nasabah       │
│ 3. Catat transaksi penalty (jika ada) │
│ 4. Saldo deposito → 0                │
│ 5. Status rekening → CLOSED           │
│ 6. Bilyet di-mark CANCELLED           │
│ 7. Kirim notifikasi                   │
└───────────────────────────────────────┘
```

**Aturan:**
- Early withdrawal membutuhkan approval **Manager+** — keputusan yang berdampak pada likuiditas koperasi
- Penalty rate dari produk (`early_withdrawal_penalty_pct` di K003 fee_config)
- General: penalty dihitung dari **bunga yang sudah earned** (bukan dari pokok)
- Islamic: tidak ada penalty atas bagi hasil yang sudah dibayar (prinsip syariah) — hanya bagi hasil bulan berjalan yang di-forfeit
- **Cooling-off period**: configurable per tenant (default: 3 hari kerja setelah approval sebelum dana dicairkan) — memberi waktu untuk pembatalan
- Deposito yang di-hold sebagai collateral **tidak bisa dicairkan dini** — harus release hold dulu (K010)
- Seluruh penalty dicatat sebagai **transaksi terpisah** untuk audit trail

### 7. Bilyet / Certificate

Setiap deposito mendapatkan **bilyet deposito** (sertifikat):

```
bilyet_deposito
├── certificate_number     VARCHAR UNIQUE per tenant
│   format: "BYT-{YEAR}-{BRANCH}-{SEQ}"
│   contoh: "BYT-2026-JKT-000001"
│
├── Isi bilyet:
│   ├── Nama koperasi & alamat
│   ├── Nama nasabah & nomor anggota
│   ├── Nomor rekening deposito
│   ├── Nominal penempatan
│   ├── Tenor & tanggal jatuh tempo
│   ├── Rate/nisbah yang berlaku
│   ├── Rollover instruction
│   ├── [BMT] Jenis akad & nisbah
│   ├── Tanggal penempatan
│   ├── Tanda tangan Manager
│   └── Tanda tangan Nasabah
│
├── Format: PDF (auto-generate)
├── Signed by: Manager branch
└── Status: ACTIVE | MATURED | CANCELLED | REPLACED
```

**Flow generate bilyet:**
```
Deposito approved & rekening dibuat
        │
        v
┌───────────────────────────────────────┐
│ Auto-generate bilyet PDF:             │
│ 1. Ambil template bilyet tenant       │
│ 2. Isi data nasabah, deposito, rate   │
│ 3. Generate certificate_number        │
│ 4. Store PDF ke object storage        │
│ 5. Link ke rekening deposito          │
└───────────────────────────────────────┘
```

**Aturan:**
- Bilyet di-generate otomatis saat rekening deposito dibuat (setelah approval)
- Certificate number **unique per tenant**, auto-increment, format configurable
- Bilyet membutuhkan **tanda tangan Manager** — digital signature atau approval flag
- Saat rollover, bilyet lama di-mark `REPLACED`, bilyet baru di-generate untuk deposito baru
- Saat early withdrawal, bilyet di-mark `CANCELLED`
- Template bilyet customizable per tenant — logo, layout, disclaimer
- BMT mode: bilyet menyertakan **surat akad Mudharabah** sebagai lampiran
- Bilyet bisa di-download nasabah dan di-print oleh teller

### 8. Deposito sebagai Collateral

Deposito bisa dijadikan jaminan untuk pinjaman (referensi K010):

```
Deposito Rp 50.000.000
        │
        v (nasabah ajukan pinjaman dengan jaminan deposito)
┌───────────────────────────────────────┐
│ Hold mechanism (K002 Section 11):     │
│ - hold_amount += plafon pinjaman      │
│   (atau % dari deposito, configurable)│
│ - available_balance berkurang         │
│ - Deposito tetap ACTIVE              │
│ - Bunga/bagi hasil tetap dihitung     │
└───────────────────────────────────────┘
        │
        v (pinjaman lunas)
┌───────────────────────────────────────┐
│ Release hold:                         │
│ - hold_amount -= plafon pinjaman      │
│ - available_balance kembali normal    │
│ - Deposito bebas dari pledge          │
└───────────────────────────────────────┘
```

**Aturan:**
- Hold menggunakan mekanisme `hold_amount` di rekening (K002 Section 11)
- Deposito yang di-hold **tetap berjalan** — bunga/bagi hasil tetap dihitung dan dibayar
- Hold amount biasanya = **100% atau 110%** dari plafon pinjaman (configurable)
- Deposito yang di-hold **tidak bisa dicairkan** (early withdrawal diblokir) dan **rollover instruction locked** (tidak bisa diubah ke NONE)
- Saat deposito jatuh tempo dan masih di-hold → **auto-rollover PRINCIPAL_AND_PROFIT** (override instruction) sampai pinjaman lunas
- Release hold saat pinjaman **lunas** (K007 event) atau saat Manager manual release (kasus khusus)
- Detail jaminan didokumentasikan di [ADR-K010](./ADR-K010-jaminan-agunan.md)

### 9. Partial Withdrawal

```
┌───────────────────────────────────────┐
│ PARTIAL WITHDRAWAL: TIDAK DIIZINKAN  │
│                                       │
│ Deposito bersifat all-or-nothing:     │
│ - Tidak bisa tarik sebagian           │
│ - Harus break seluruh deposito        │
│ - Jika ingin nominal lebih kecil:     │
│   → Break deposito lama (+ penalty)   │
│   → Buka deposito baru dengan         │
│     nominal yang diinginkan           │
└───────────────────────────────────────┘
```

**Aturan:**
- Partial withdrawal **tidak diperbolehkan** — sesuai praktik standar deposito koperasi
- Alasan: bilyet dan rate terikat pada nominal penempatan awal
- Jika nasabah butuh sebagian dana:
  1. Ajukan early withdrawal (break seluruh deposito)
  2. Terima pokok + bunga - penalty
  3. Buka deposito baru dengan nominal yang lebih kecil (jika masih ingin deposito)
  4. Sisa dana masuk ke tabungan
- Proses ini membutuhkan **dua transaksi** terpisah — pencairan lama + penempatan baru

### 10. Teacher/Staff Considerations

Deposito paling cocok untuk guru/staff dengan penghasilan tetap:

```
teacher_staff_deposito_config:
  ── Bonus Rate (opsional, configurable per tenant) ──
  bonus_rate_enabled:        true
  bonus_rate_conditions:
    - years_of_service: 5    bonus_bps: 25    # +0.25% untuk masa kerja > 5 tahun
    - years_of_service: 10   bonus_bps: 50    # +0.50% untuk masa kerja > 10 tahun
    - years_of_service: 20   bonus_bps: 75    # +0.75% untuk masa kerja > 20 tahun
  
  ── Payroll Integration ──
  payroll_auto_placement:    true    # Potong gaji otomatis untuk deposito bulanan
  payroll_deposito_amount:   configurable per nasabah
```

**Aturan:**
- Bonus rate **opsional** — hanya jika tenant mengaktifkan
- Bonus rate ditambahkan di atas rate tier standar (Section 2)
- Masa kerja diambil dari data sekolah (jika terintegrasi) atau diinput manual
- Bonus rate hanya untuk `school_relation_type = teacher` atau `staff`
- Payroll auto-placement: potong gaji → langsung ditempatkan sebagai deposito baru
- Detail integrasi payroll mengikuti mekanisme yang sama dengan K004 (payroll deduction)

### 11. Maturity Tracking — Scheduled Job

```
Daily Maturity Check (scheduled job, pagi hari):
        │
        v
┌───────────────────────────────────────┐
│ 1. Query semua deposito dengan        │
│    maturity_date = TODAY              │
│                                       │
│ 2. Per deposito:                      │
│    a. Cek rollover_instruction        │
│    b. Cek apakah ada hold (collateral)│
│    c. Process sesuai instruksi:       │
│       - NONE → cairkan               │
│       - PRINCIPAL_ONLY → rollover     │
│       - PRINCIPAL_AND_PROFIT → roll   │
│       - TO_TABUNGAN → transfer        │
│    d. Generate bilyet baru (rollover) │
│    e. Kirim notifikasi ke nasabah     │
│                                       │
│ 3. Log semua processing ke audit      │
│                                       │
│ 4. Generate laporan maturity harian   │
│    untuk Manager branch               │
└───────────────────────────────────────┘
```

**Notification schedule:**
```
notification_schedule:
  - days_before: 30    channel: in_app    message: "Deposito Anda jatuh tempo 30 hari lagi"
  - days_before: 7     channel: in_app    message: "Deposito Anda jatuh tempo 7 hari lagi"
  - days_before: 1     channel: in_app    message: "Deposito Anda jatuh tempo besok"
  - days_before: 0     channel: in_app    message: "Deposito Anda telah jatuh tempo"
```

**Aturan:**
- Job berjalan **harian** di pagi hari sebelum jam operasional
- Maturity yang jatuh di **hari libur** diproses di hari kerja berikutnya
- Jika processing gagal (misal: tabungan tujuan sudah CLOSED), deposito di-hold dan teller diberi notifikasi untuk proses manual
- Laporan maturity harian dikirim ke Manager branch — daftar deposito yang jatuh tempo hari ini
- Upcoming maturity report: deposito yang akan jatuh tempo 7 dan 30 hari ke depan (untuk planning)

### 12. Data Model

Deposito menggunakan tabel **rekening** dari K002 dengan `category = deposito`. Field tambahan khusus deposito:

#### 12a. Field Deposito di Tabel `rekening`

```
rekening (field tambahan untuk deposito):
├── ── Deposito Specific ──
├── deposito_config           JSONB (nullable, hanya untuk deposito)
│   {
│     "placement_amount":      50000000,
│     "tenor_months":          12,
│     "placement_date":        "2026-04-15",
│     "maturity_date":         "2027-04-15",
│     "rollover_instruction":  "principal_only",
│     "interest_payment_method": "monthly",
│     "rate_locked":           8.0000,
│     "rate_tier_applied":     "12m_50jt_above",
│     "certificate_number":    "BYT-2026-JKT-000001",
│     "certificate_url":       "https://storage/.../byt-001.pdf",
│     "certificate_status":    "active",
│     "is_collateral":         false,
│     "collateral_loan_id":    null,
│     "bonus_rate_bps":        25,
│     "total_rate":            8.2500
│   }
│
│   [Islamic mode tambahan]:
│   {
│     "nisbah_nasabah_locked":  45.00,
│     "nisbah_koperasi_locked": 55.00,
│     "indicative_rate_locked": 7.5000,
│     "akad_document_url":      "https://storage/.../akad-mudharabah-001.pdf"
│   }
```

#### 12b. Tabel `deposito_interest_schedule`

```
deposito_interest_schedule
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── rekening_id           UUID (FK → rekening deposito)
│
├── ── Schedule ──
├── period                VARCHAR (format: "2026-04")
├── posting_date          DATE NOT NULL
├── status                ENUM (scheduled, posted, skipped)
│
├── ── Calculation ──
├── principal_amount      NUMERIC(15,2) NOT NULL
├── rate_applied          NUMERIC(7,4) NOT NULL
├── days_in_period        INTEGER NOT NULL
├── gross_interest        NUMERIC(15,2) NOT NULL
├── tax_amount            NUMERIC(15,2) NOT NULL
├── net_interest          NUMERIC(15,2) NOT NULL
│
├── ── Payment ──
├── paid_to_rekening_id   UUID (nullable, FK → rekening tabungan, jika monthly payment)
├── transaction_id        UUID (nullable, FK → transaksi, filled on posting)
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

#### 12c. Tabel `deposito_bilyet`

```
deposito_bilyet
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── rekening_id           UUID (FK → rekening deposito)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Certificate ──
├── certificate_number    VARCHAR UNIQUE per tenant
├── status                ENUM (active, matured, cancelled, replaced)
│
├── ── Content ──
├── placement_amount      NUMERIC(15,2) NOT NULL
├── tenor_months          INTEGER NOT NULL
├── placement_date        DATE NOT NULL
├── maturity_date         DATE NOT NULL
├── rate_or_nisbah        VARCHAR NOT NULL (display: "8% p.a." atau "Nisbah 45:55")
├── rollover_instruction  VARCHAR NOT NULL
│
├── ── Document ──
├── document_url          VARCHAR (URL ke PDF)
├── signed_by             UUID (FK → user, Manager yang menandatangani)
├── signed_at             TIMESTAMPTZ
│
├── ── Lifecycle ──
├── replaced_by_id        UUID (nullable, FK → deposito_bilyet, jika rollover)
├── cancelled_at          TIMESTAMPTZ (nullable)
├── cancel_reason         VARCHAR (nullable)
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

#### 12d. Vernon _rels dan _data Structure

**deposito_interest_schedule _rels:**
```json
{
  "tenant_id":           "018f...",
  "rekening_id":         "018f...",
  "paid_to_rekening_id": "018f..."
}
```

**deposito_interest_schedule _data:**
```json
{
  "rekening": {
    "id":             "018f...",
    "account_number": "DP-2026-JKT-00000001",
    "nasabah_name":   "Ahmad Fauzi"
  },
  "paid_to_rekening": {
    "id":             "018f...",
    "account_number": "TB-2026-JKT-00000001"
  }
}
```

**deposito_bilyet _rels:**
```json
{
  "tenant_id":    "018f...",
  "rekening_id":  "018f...",
  "nasabah_id":   "018f...",
  "signed_by":    "018f..."
}
```

**deposito_bilyet _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "rekening": {
    "id":             "018f...",
    "account_number": "DP-2026-JKT-00000001",
    "product_name":   "Deposito Mudharabah 12 Bulan"
  },
  "signer": {
    "id":   "018f...",
    "name": "Hj. Fatimah, S.Pd"
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.nasabah` di deposito_bilyet
- `RekeningUpdatedEvent` → update `_data.rekening` di interest_schedule & bilyet
- `UserUpdatedEvent` (signer) → update `_data.signer` di deposito_bilyet

### 13. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View deposito & saldo | v | v | v | v |
| Buat pengajuan deposito | v | v | v | v |
| Approve pengajuan deposito | - | v | v | v |
| View bilyet deposito | v | v | v | v |
| Generate/print bilyet | v | v | v | v |
| Ubah rollover instruction | v | v | v | v |
| Early withdrawal (ajukan) | v | v | v | v |
| Approve early withdrawal | - | - | v | v |
| Sign bilyet (tanda tangan digital) | - | - | v | v |
| View interest/profit schedule | v | v | v | v |
| View maturity report | v | v | v | v |
| Process maturity (manual override) | - | v | v | v |
| Set bonus rate guru/staff | - | - | v | v |
| Configure rate tiers | - | - | - | v |

**Catatan:**
- Teller bisa buat pengajuan dan view semua informasi deposito — operasi day-to-day
- Approval pengajuan oleh **Supervisor+** — konsisten dengan K002
- Early withdrawal approval hanya **Manager+** — keputusan berdampak pada likuiditas
- Bilyet signing hanya **Manager+** — otoritas resmi koperasi
- Rate tiers configuration hanya **Admin** — berdampak pada semua deposito baru

### 14. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Entity name | Deposito / Simpanan Berjangka | Deposito Mudharabah |
| Rate label | Suku Bunga | Nisbah Bagi Hasil |
| Rate display | 8% p.a. | Nisbah 45:55 (Ind. Rate: 7.5%) |
| Yield label | Bunga Deposito | Bagi Hasil Deposito |
| Yield posting | Posting Bunga | Posting Bagi Hasil |
| Certificate | Bilyet Deposito | Bilyet Deposito + Akad Mudharabah |
| Early withdrawal penalty | Penalti Pencairan Dini | Forfeit Bagi Hasil Bulan Berjalan |
| Tax label | PPh atas Bunga | PPh atas Bagi Hasil |
| Maturity notice | Pemberitahuan Jatuh Tempo | Pemberitahuan Jatuh Tempo |
| Rollover | Perpanjangan Otomatis | Perpanjangan Otomatis |
| Collateral label | Jaminan Deposito | Jaminan Deposito |

## Consequences

### Positif

- **Higher yield** — nasabah mendapat return lebih tinggi dari tabungan, mendorong penghimpunan dana jangka panjang
- **Predictable cash flow** — tenor tetap membantu koperasi merencanakan penyaluran dana
- **Tiered rates** — mendorong penempatan nominal besar dan tenor panjang
- **Auto-maturity** — rollover otomatis mengurangi beban operasional dan risiko idle cash
- **Collateral-ready** — deposito bisa dijaminkan untuk pinjaman tanpa likuidasi (hold mechanism)
- **Bilyet formal** — sertifikat deposito memenuhi aspek legal dan meningkatkan kepercayaan nasabah
- **Dual-mode** — bunga tetap (general) dan bagi hasil indikatif (BMT) sama-sama terdefinisi jelas

### Negatif

- **Liquidity risk** — dana terkunci selama tenor, koperasi harus manage maturity mismatch
- **Rate tier complexity** — banyak kombinasi tenor × nominal, perlu UI yang jelas untuk nasabah
- **Maturity processing** — scheduled job yang gagal bisa menyebabkan deposito tidak diproses tepat waktu
- **Bilyet management** — generate, sign, store, replace — lifecycle bilyet menambah complexity
- **Islamic profit variance** — bagi hasil aktual bisa jauh dari indicative rate, memerlukan komunikasi yang jelas ke nasabah

### Mitigasi

- Maturity report harian dan upcoming maturity alert ke Manager — proaktif monitoring
- Rate tiers ditampilkan sebagai **tabel sederhana** di UI — nasabah langsung lihat rate untuk tenor & nominal mereka
- Scheduled job maturity dilengkapi **retry mechanism** dan **manual fallback** — jika auto gagal, teller bisa proses manual
- Bilyet di-generate async — tidak memblokir proses pembukaan rekening
- BMT: disclaimer **wajib** ditampilkan: "Bagi hasil yang tercantum adalah indikatif dan dapat berubah sesuai kinerja koperasi"

## Alternatives Considered

### A. Single Rate (tanpa tiered structure)

Satu rate untuk semua tenor dan nominal.

**Ditolak karena:** tidak kompetitif dan tidak mendorong nasabah untuk menempatkan dana lebih besar atau lebih lama. Tiered rate adalah praktik standar di perbankan dan koperasi untuk optimasi likuiditas.

### B. Partial Withdrawal Diizinkan

Nasabah bisa tarik sebagian dari deposito tanpa break seluruh placement.

**Ditolak karena:** memperumit perhitungan bunga (rate berubah jika nominal turun ke tier yang lebih rendah) dan manajemen bilyet (harus re-issue untuk nominal yang berubah). Break + open new deposito lebih clean dan auditable.

### C. Tanpa Bilyet (digital only)

Tidak menerbitkan sertifikat deposito, hanya record digital.

**Ditolak karena:** regulasi koperasi dan kebiasaan nasabah masih mengharapkan bukti fisik/cetak untuk deposito. Bilyet juga menjadi bukti hukum jika ada sengketa. Digital-first dengan opsi cetak adalah kompromi yang tepat.

### D. Bagi Hasil Fixed (bukan aktual) untuk BMT

BMT menggunakan fixed rate seperti konvensional, hanya beda label.

**Ditolak karena:** melanggar prinsip syariah Mudharabah yang mengharuskan bagi hasil berdasarkan profit riil (bukan dijanjikan di muka). Fixed rate pada BMT bisa dikategorikan sebagai riba. Indicative rate + actual profit sharing adalah mekanisme yang sesuai fatwa DSN-MUI.
