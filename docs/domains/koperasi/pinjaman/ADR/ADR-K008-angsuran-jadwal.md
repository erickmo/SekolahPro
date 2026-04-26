# ADR-K008: Angsuran & Jadwal

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Angsuran (installment) dan jadwal pembayaran adalah jantung operasional pinjaman/pembiayaan. Setiap pinjaman yang di-approve menghasilkan jadwal angsuran yang menentukan kapan, berapa, dan bagaimana nasabah harus membayar. Sistem harus mengakomodasi:

- **Kalkulasi akurat**: Berbeda metode perhitungan untuk konvensional (flat, declining, annuity) dan Islamic (Murabahah flat, Musyarakah projected bagi hasil)
- **Pencatatan pembayaran**: Terima pembayaran penuh, partial, overpayment, dan match ke jadwal yang tepat
- **Multi-channel**: Teller (cash), auto-debit tabungan, potong gaji (payroll deduction)
- **Aging & DPD**: Tracking keterlambatan per angsuran untuk NPL classification (K007)
- **Reconciliation**: Perbandingan jadwal vs aktual untuk audit dan pelaporan
- **Dual-mode**: Terminologi dan perhitungan berbeda antara konvensional dan BMT

## Decision

### 1. Installment Schedule Generation

Jadwal angsuran di-generate **otomatis** saat pinjaman di-approve dan sebelum disbursement:

```
PinjamanApprovedEvent
        |
        v
+-----------------------------------+
| Generate Jadwal Angsuran:         |
| 1. Hitung per metode/akad        |
| 2. Tentukan due date per bulan   |
| 3. Buat N rows (N = tenor)       |
| 4. Simpan ke tabel angsuran      |
+-----------------------------------+
        |
        v
+-----------------------------------+
| Jadwal siap, lanjut disbursement  |
+-----------------------------------+
```

**Aturan:**
- Jadwal di-generate **sebelum disbursement** — nasabah bisa review jadwal sebelum dana cair
- Jumlah baris jadwal = tenor (bulan)
- Grace period mempengaruhi due date angsuran pertama (bukan jumlah angsuran)
- Jadwal bersifat **immutable** setelah disbursement — perubahan hanya via restrukturisasi (K007)
- Jika ada grace period dengan tipe `principal_only`, tambahkan baris khusus grace period (bunga/margin saja)

### 2. Metode Perhitungan — General Mode

**A. Flat (Bunga Tetap)**

Bunga dihitung dari pokok awal, angsuran tetap setiap bulan.

```
Contoh: Pokok = 12.000.000, Rate = 12%/tahun, Tenor = 12 bulan

Total Bunga   = Pokok x Rate x (Tenor/12)
              = 12.000.000 x 12% x (12/12) = 1.440.000

Angsuran/bln  = (Pokok + Total Bunga) / Tenor
              = (12.000.000 + 1.440.000) / 12 = 1.120.000

Porsi Pokok   = 12.000.000 / 12 = 1.000.000 (tetap)
Porsi Bunga   = 1.440.000 / 12  = 120.000   (tetap)
```

| Bln | Pokok | Bunga | Total | Sisa Pokok |
|-----|-------|-------|-------|------------|
| 1 | 1.000.000 | 120.000 | 1.120.000 | 11.000.000 |
| 2 | 1.000.000 | 120.000 | 1.120.000 | 10.000.000 |
| ... | ... | ... | ... | ... |
| 12 | 1.000.000 | 120.000 | 1.120.000 | 0 |

**B. Declining / Efektif (Bunga Menurun)**

Bunga dihitung dari sisa pokok, angsuran menurun setiap bulan.

```
Contoh: Pokok = 12.000.000, Rate = 12%/tahun, Tenor = 12 bulan

Porsi Pokok   = 12.000.000 / 12 = 1.000.000 (tetap)
Bunga bln-N   = Sisa Pokok x (Rate / 12)
```

| Bln | Pokok | Bunga | Total | Sisa Pokok |
|-----|-------|-------|-------|------------|
| 1 | 1.000.000 | 120.000 | 1.120.000 | 11.000.000 |
| 2 | 1.000.000 | 110.000 | 1.110.000 | 10.000.000 |
| 3 | 1.000.000 | 100.000 | 1.100.000 | 9.000.000 |
| ... | ... | ... | ... | ... |
| 12 | 1.000.000 | 10.000 | 1.010.000 | 0 |

**C. Anuitas (Annuity)**

Angsuran tetap setiap bulan, porsi pokok dan bunga berubah.

```
Contoh: Pokok = 12.000.000, Rate = 12%/tahun, Tenor = 12 bulan

Monthly Rate (r) = 12% / 12 = 1% = 0.01

Angsuran/bln  = Pokok x r x (1+r)^n / ((1+r)^n - 1)
              = 12.000.000 x 0.01 x (1.01)^12 / ((1.01)^12 - 1)
              = 12.000.000 x 0.01 x 1.12682503 / (1.12682503 - 1)
              = 12.000.000 x 0.01126825 / 0.12682503
              = 1.066.185 (pembulatan)

Bunga bln-N   = Sisa Pokok x r
Pokok bln-N   = Angsuran - Bunga bln-N
```

| Bln | Pokok | Bunga | Total | Sisa Pokok |
|-----|-------|-------|-------|------------|
| 1 | 946.185 | 120.000 | 1.066.185 | 11.053.815 |
| 2 | 955.647 | 110.538 | 1.066.185 | 10.098.168 |
| 3 | 965.203 | 100.982 | 1.066.185 | 9.132.965 |
| ... | ... | ... | ... | ... |
| 12 | 1.055.629 | 10.556 | 1.066.185 | 0 |

### 3. Metode Perhitungan — Islamic Mode

**A. Murabahah (Margin Tetap)**

Sama seperti flat — margin dihitung di awal, angsuran tetap. Ini akad **paling umum** di BMT.

```
Contoh: Harga Barang = 12.000.000, Margin = 12%, Tenor = 12 bulan

Total Margin  = 12.000.000 x 12% = 1.440.000
Harga Jual    = 12.000.000 + 1.440.000 = 13.440.000

Angsuran/bln  = 13.440.000 / 12 = 1.120.000

Porsi Pokok   = 12.000.000 / 12 = 1.000.000 (tetap)
Porsi Margin  = 1.440.000 / 12  = 120.000   (tetap)
```

**Perbedaan dengan Flat konvensional**: Secara kalkulasi sama, tapi secara **akad** berbeda. Murabahah adalah jual beli — BMT membeli barang lalu menjual ke nasabah dengan margin. Total harga jual ditetapkan di awal dan **tidak berubah** (fixed price).

**B. Musyarakah / Mudharabah (Bagi Hasil)**

Angsuran pokok tetap, bagi hasil mengikuti **projected profit** yang di-review berkala.

```
Contoh: Modal BMT = 10.000.000, Nisbah: Nasabah 60%, BMT 40%, Tenor = 12 bulan

Porsi Pokok   = 10.000.000 / 12 = 833.333 (tetap)
Bagi Hasil    = Projected profit x Nisbah BMT

Projected profit per bulan = Rp 500.000 (estimasi)
Bagi Hasil BMT = 500.000 x 40% = 200.000

Angsuran/bln  = 833.333 + 200.000 = 1.033.333 (bisa berubah)
```

**Aturan khusus:**
- Jadwal Musyarakah/Mudharabah menggunakan **projected** bagi hasil saat generate
- Review bagi hasil dilakukan **per 3 bulan** (configurable) berdasarkan laporan profit aktual
- Jika profit aktual berbeda dari projected: selisih di-adjust di angsuran berikutnya
- Laporan profit nasabah harus di-submit dan di-verify oleh Supervisor

**C. Ijarah (Ujrah/Sewa)**

Ujrah (biaya sewa) tetap per bulan, tidak ada pokok yang berkurang.

```
Contoh: Nilai Aset = 12.000.000, Ujrah = 1% per bulan, Tenor = 12 bulan

Ujrah/bln     = 12.000.000 x 1% = 120.000 (tetap)
Tidak ada porsi pokok — nasabah membayar sewa
Di akhir: opsi beli (jika Ijarah Muntahiyah Bittamlik) atau kembalikan aset
```

**D. Qardh (Tanpa Profit)**

Angsuran hanya pokok, tidak ada bunga/margin/bagi hasil.

```
Contoh: Pokok = 1.000.000, Tenor = 10 bulan

Angsuran/bln  = 1.000.000 / 10 = 100.000 (pokok saja)
Bunga/Margin  = 0
```

### 4. Jadwal Angsuran — Data Model

```
angsuran
+-- id                    UUID v7 (PK)
+-- tenant_id             UUID (FK -> tenant)
+-- pinjaman_id           UUID (FK -> pinjaman) NOT NULL
+-- rekening_id           UUID (FK -> rekening pinjaman) NOT NULL
|
+-- -- Jadwal (diisi saat generate) --
+-- installment_number    INTEGER NOT NULL (1, 2, 3, ..., N)
+-- due_date              DATE NOT NULL
+-- principal_amount      NUMERIC(15,2) NOT NULL (porsi pokok jadwal)
+-- interest_amount       NUMERIC(15,2) DEFAULT 0 (porsi bunga, konvensional)
+-- margin_amount         NUMERIC(15,2) DEFAULT 0 (porsi margin, BMT murabahah)
+-- profit_share_amount   NUMERIC(15,2) DEFAULT 0 (bagi hasil, BMT musyarakah/mudharabah)
+-- ujrah_amount          NUMERIC(15,2) DEFAULT 0 (sewa, BMT ijarah)
+-- total_amount          NUMERIC(15,2) NOT NULL (total angsuran jadwal)
+-- remaining_principal   NUMERIC(15,2) NOT NULL (sisa pokok setelah angsuran ini)
|
+-- -- Pembayaran Aktual --
+-- paid_principal        NUMERIC(15,2) DEFAULT 0
+-- paid_interest         NUMERIC(15,2) DEFAULT 0
+-- paid_margin           NUMERIC(15,2) DEFAULT 0
+-- paid_profit_share     NUMERIC(15,2) DEFAULT 0
+-- paid_ujrah            NUMERIC(15,2) DEFAULT 0
+-- paid_penalty          NUMERIC(15,2) DEFAULT 0 (denda yang dibayar di angsuran ini)
+-- total_paid            NUMERIC(15,2) DEFAULT 0
|
+-- -- Status Pembayaran --
+-- payment_status        ENUM (scheduled, partial, paid, overdue, waived, voided)
|   # Status `voided` digunakan saat restrukturisasi (K007) — jadwal lama di-void, jadwal baru di-generate
+-- paid_date             DATE (nullable, tanggal pembayaran aktual)
+-- paid_via              ENUM (teller_cash, auto_debit, payroll) (nullable)
+-- transaction_id        UUID (nullable, FK -> transaksi, link ke transaksi pembayaran)
|
+-- -- Keterlambatan --
+-- dpd                   INTEGER DEFAULT 0 (Days Past Due — auto-calculate)
+-- is_overdue            BOOLEAN DEFAULT false
+-- overdue_since         DATE (nullable)
|
+-- -- Bagi Hasil Adjustment (khusus Musyarakah/Mudharabah) --
+-- projected_profit      NUMERIC(15,2) DEFAULT 0
+-- actual_profit         NUMERIC(15,2) (nullable)
+-- profit_adjustment     NUMERIC(15,2) DEFAULT 0 (selisih projected vs actual)
|
+-- -- Vernon Fields --
+-- _rels                 JSONB NOT NULL DEFAULT '{}'
+-- _data                 JSONB NOT NULL DEFAULT '{}'
|
+-- -- Audit --
    +-- created_at        TIMESTAMPTZ
    +-- created_by        UUID
    +-- updated_at        TIMESTAMPTZ
    +-- updated_by        UUID
```

**Index penting:**
- `(pinjaman_id, installment_number)` — UNIQUE, lookup jadwal per pinjaman
- `(due_date, payment_status)` — untuk query angsuran jatuh tempo hari ini
- `(pinjaman_id, payment_status)` — untuk query angsuran tertunggak per pinjaman
- `(tenant_id, due_date)` — untuk laporan angsuran jatuh tempo per tenant

### 5. Payment Recording — Pencatatan Pembayaran

```
Nasabah bayar angsuran
        |
        v
+-----------------------------------+
| 1. Identifikasi angsuran target   |
|    - Auto: angsuran terlama yang  |
|      belum lunas (FIFO)           |
|    - Manual: Teller pilih         |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 2. Alokasi pembayaran:           |
|    Prioritas:                     |
|    a. Denda tertunggak (K009)     |
|    b. Bunga/margin tertunggak     |
|    c. Pokok tertunggak            |
|    d. Bunga/margin berjalan       |
|    e. Pokok berjalan              |
|    f. Sisa -> overpayment         |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 3. Update angsuran record:        |
|    - paid_* fields                |
|    - payment_status               |
|    - paid_date, paid_via          |
|    - Link transaction_id          |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 4. Update pinjaman outstanding    |
| 5. Update rekening balance        |
| 6. Cek: semua angsuran paid?      |
|    Ya -> status pinjaman COMPLETED|
+-----------------------------------+
```

**Prioritas alokasi pembayaran:**
1. **Denda outstanding** (jika ada) — harus dilunasi dulu
2. **Bunga/margin tertunggak** — bunga/margin dari angsuran yang sudah overdue
3. **Pokok tertunggak** — pokok dari angsuran yang sudah overdue
4. **Bunga/margin berjalan** — bunga/margin angsuran saat ini
5. **Pokok berjalan** — pokok angsuran saat ini
6. **Sisa** -> overpayment (lihat Section 6)

**Concurrency Control — Mandatory Locking:**

Pembayaran angsuran WAJIB menggunakan row-level locking untuk mencegah race condition antara:
- Auto-debit scheduled job dan pembayaran manual teller secara bersamaan
- Dua teller memproses pembayaran nasabah yang sama

Mekanisme:
1. `SELECT ... FOR UPDATE` pada row angsuran yang akan dibayar
2. Jika row sudah di-lock oleh proses lain, tunggu (dengan timeout configurable, default 5 detik)
3. Jika timeout, return error "Pembayaran sedang diproses, coba lagi"
4. Setelah lock acquired: validasi status (masih SCHEDULED/OVERDUE/PARTIAL?), proses pembayaran, update status, release lock

Auto-debit job: proses secara sequential per nasabah, bukan parallel, untuk menghindari self-deadlock.

**Aturan:**
- Pembayaran selalu di-match ke angsuran **paling lama yang belum lunas** (FIFO) — kecuali Teller override manual
- Satu pembayaran bisa menutup **multiple angsuran** jika nominal cukup
- Semua update dalam **satu database transaction** — atomik
- Setiap pembayaran menghasilkan **transaksi** (K011) yang di-link ke angsuran
- Pembulatan: selalu bulatkan ke bawah untuk pokok, sisa masuk ke angsuran terakhir

### 6. Overpayment Handling

Jika nasabah membayar lebih dari angsuran yang jatuh tempo:

```
overpayment_config:
  handling: "next_installment" | "reduce_principal"

  # next_installment: kelebihan diaplikasikan ke angsuran berikutnya
  # reduce_principal: kelebihan mengurangi pokok (tenor tetap, angsuran berkurang)
```

**Mode "next_installment" (default):**
- Kelebihan langsung diaplikasikan ke angsuran berikutnya
- Jika cukup, angsuran berikutnya juga ter-mark `paid`
- Proses berlanjut sampai kelebihan habis

**Mode "reduce_principal":**
- Kelebihan mengurangi `outstanding_principal`
- Jadwal angsuran **di-recalculate** — tenor tetap, angsuran turun
- ATAU: tenor berkurang, angsuran tetap (configurable)

**Aturan:**
- Mode handling **configurable per produk**
- Untuk Murabahah: overpayment selalu **next_installment** — karena total harga jual sudah fixed
- Untuk Musyarakah/Mudharabah: overpayment bisa **reduce_principal** — mengurangi modal BMT
- Overpayment amount dan handling tercatat di audit trail

### 7. Underpayment — Pembayaran Sebagian

Jika nasabah membayar kurang dari angsuran yang jatuh tempo:

```
Angsuran jatuh tempo: 1.120.000
Nasabah bayar:          800.000
Kurang:                 320.000
        |
        v
+-----------------------------------+
| 1. Alokasi sesuai prioritas:     |
|    - Bunga/margin dulu: 120.000   |
|    - Sisa ke pokok:    680.000    |
|    (pokok jadwal 1.000.000)       |
|    Kurang pokok: 320.000          |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 2. Status: PARTIAL               |
|    paid_interest: 120.000         |
|    paid_principal: 680.000        |
|    Sisa tertunggak: 320.000      |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 3. Sisa 320.000 carry forward    |
|    ke angsuran berikutnya         |
|    (ditambahkan ke total_amount)  |
+-----------------------------------+
```

**Aturan:**
- Underpayment diterima — sistem **tidak menolak** pembayaran parsial
- Angsuran di-mark `PARTIAL` dengan detail berapa yang sudah terbayar
- Sisa yang belum terbayar di-carry forward ke bulan berikutnya
- Angsuran `PARTIAL` tetap dihitung sebagai **overdue** untuk DPD jika sudah lewat due date
- Minimum payment threshold: **tidak ada** — berapapun diterima

### 8. Payment Channels

**A. Teller (Cash / Transfer)**

```
Teller menerima pembayaran cash
        |
        v
+-----------------------------------+
| 1. Input nominal pembayaran       |
| 2. Sistem auto-match ke angsuran  |
| 3. Teller confirm                 |
| 4. Record transaksi               |
| 5. Cetak bukti pembayaran         |
+-----------------------------------+
```

**B. Auto-debit dari Tabungan**

```
auto_debit_config:
  enabled: true
  source_account_id: "018f..."     # Rekening tabungan sumber
  debit_day: "due_date"            # Hari debit = hari jatuh tempo
  debit_time: "08:00"              # Jam eksekusi
  retry_days: 3                    # Jika saldo kurang, retry selama 3 hari
  partial_debit: false             # Jika saldo kurang, skip (jangan partial)
```

```
Scheduled Job (harian, pagi):
+-----------------------------------+
| 1. Query angsuran jatuh tempo     |
|    hari ini dengan auto_debit     |
|    enabled                        |
| 2. Cek available_balance rekening |
|    tabungan                       |
| 3. Jika cukup:                    |
|    - Debit tabungan               |
|    - Credit rekening pinjaman     |
|    - Record pembayaran angsuran   |
| 4. Jika tidak cukup:              |
|    - Skip (coba lagi besok sampai |
|      retry habis)                 |
|    - Notifikasi nasabah           |
+-----------------------------------+
```

**C. Payroll Deduction (Potong Gaji)**

Khusus untuk **guru/staff** yang gajinya diproses melalui sistem sekolah:

```
payroll_deduction_config:
  enabled: true
  school_entity_id: "018f..."      # Link ke entitas sekolah
  deduction_type: "fixed_amount"   # Potongan tetap per bulan
  deduction_amount: 1.120.000      # = angsuran bulanan
  effective_from: "2026-05-01"
  effective_until: "2027-04-01"    # = maturity date
```

```
Payroll Processing (bulanan):
+-----------------------------------+
| 1. Sekolah proses gaji bulanan    |
| 2. Sistem kirim daftar potongan   |
|    per guru/staff                 |
| 3. Sekolah potong gaji & transfer |
|    ke koperasi                    |
| 4. Koperasi terima & match ke     |
|    angsuran masing-masing         |
+-----------------------------------+
```

**Aturan:**
- Payroll deduction membutuhkan **persetujuan tertulis** nasabah + sekolah
- Jika nasabah resign/pindah: auto-debit dari tabungan sebagai fallback
- Potongan gaji **prioritas di atas** kewajiban lain (sesuai perjanjian)
- Konfirmasi penerimaan dana dari sekolah harus di-reconcile

### 9. Due Date Configuration

```
due_date_config:
  mode: "fixed_day" | "disbursement_relative"

  # fixed_day: semua angsuran jatuh tempo di tanggal yang sama setiap bulan
  fixed_day: 10                    # Tanggal 10 setiap bulan

  # disbursement_relative: jatuh tempo N hari/bulan setelah disbursement
  # (otomatis berdasarkan tanggal disbursement)

  # Holiday adjustment
  holiday_adjustment: "next_business_day" | "previous_business_day" | "none"
```

**Aturan:**
- Mode **configurable per tenant** — default: `fixed_day`
- `fixed_day`: semua pinjaman di tenant tersebut jatuh tempo di tanggal yang sama (memudahkan operasional)
- `disbursement_relative`: jatuh tempo mengikuti tanggal disbursement (misal: disbursement tgl 15, angsuran tiap tgl 15)
- Jika tanggal fixed > 28 (misal: 30, 31) dan bulan tidak punya tanggal tersebut -> geser ke akhir bulan
- Holiday adjustment: jika jatuh tempo jatuh di hari libur nasional/weekend:
  - `next_business_day`: geser ke hari kerja berikutnya (default)
  - `previous_business_day`: geser ke hari kerja sebelumnya
  - `none`: tetap di tanggal asli (DPD tetap dihitung dari tanggal asli)
- Kalender hari libur **configurable per tenant per tahun**

### 10. Aging Report — DPD Tracking

DPD (Days Past Due) dihitung **per angsuran** dan per pinjaman:

```
DPD Calculation (scheduled job, harian):
+-----------------------------------+
| Untuk setiap angsuran yang        |
| payment_status != 'paid':         |
|                                   |
| IF today > due_date:              |
|   dpd = today - due_date          |
|   is_overdue = true               |
|   overdue_since = due_date        |
| ELSE:                             |
|   dpd = 0                         |
|   is_overdue = false              |
+-----------------------------------+
        |
        v
+-----------------------------------+
| DPD Pinjaman = MAX(dpd) dari      |
| semua angsuran yang overdue       |
|                                   |
| Update pinjaman.dpd               |
| Update pinjaman.npl_classification|
| (berdasarkan tabel NPL di K007)  |
+-----------------------------------+
```

**Aging Buckets (untuk laporan):**

| Bucket | DPD Range | Label | Warna |
|--------|-----------|-------|-------|
| Current | 0 | Lancar | Hijau |
| 1-30 | 1-30 | Perhatian Khusus | Kuning |
| 31-60 | 31-60 | Perlu Tindakan | Oranye |
| 61-90 | 61-90 | Eskalasi | Merah Muda |
| 91-120 | 91-120 | Kurang Lancar | Merah |
| 121-180 | 121-180 | Diragukan | Merah Tua |
| >180 | 181+ | Macet | Hitam |

**Aturan:**
- DPD di-update **setiap hari** via scheduled job (off-peak hours)
- DPD pinjaman = DPD **tertinggi** dari semua angsuran yang overdue
- DPD di-reset ke 0 saat **semua** angsuran tertunggak dilunasi
- Pembayaran partial **tidak** me-reset DPD — harus lunas penuh per angsuran
- Aging report tersedia per cabang, per tenant, dan agregat

### 11. Prepayment / Extra Payment

Nasabah membayar lebih dari angsuran bulanan, atau membayar beberapa angsuran sekaligus:

```
prepayment_config:
  allowed: true
  effect: "shorten_tenor" | "reduce_installment" | "nasabah_choice"

  # shorten_tenor: tenor berkurang, angsuran tetap
  # reduce_installment: angsuran turun, tenor tetap
  # nasabah_choice: nasabah pilih saat bayar
```

**Flow prepayment:**

```
Nasabah bayar 3x angsuran sekaligus
        |
        v
+-----------------------------------+
| 1. Tutup angsuran bulan ini       |
| 2. Tutup angsuran bulan depan     |
| 3. Tutup angsuran bulan depannya  |
| 4. Semua di-mark PAID             |
+-----------------------------------+
        |
        v
+-----------------------------------+
| [Jika effect = shorten_tenor]     |
| Tenor efektif berkurang 2 bulan   |
|                                   |
| [Jika effect = reduce_installment]|
| Recalculate sisa angsuran         |
+-----------------------------------+
```

**Aturan:**
- Prepayment **selalu diperbolehkan** (configurable — bisa dimatikan per produk)
- Untuk Murabahah: prepayment menutup angsuran berikutnya (total harga jual tetap)
- Untuk Declining/Anuitas: prepayment bisa mengurangi total bunga karena pokok berkurang lebih cepat
- Extra payment di-record sebagai transaksi terpisah dengan kode `PREPAYMENT`

### 12. Rescheduling Impact

Saat pinjaman di-restrukturisasi (K007 Section 8), jadwal angsuran berubah:

```
Restructuring Approved
        |
        v
+-----------------------------------+
| 1. Void semua angsuran lama       |
|    yang belum dibayar             |
|    (status -> VOIDED)             |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 2. Generate jadwal baru:          |
|    - Sisa pokok = outstanding     |
|    - Terms baru (tenor, rate)     |
|    - Due date baru                |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 3. Angsuran yang sudah dibayar    |
|    TIDAK berubah — tetap tercatat |
+-----------------------------------+
```

**Aturan:**
- Angsuran yang sudah PAID **tidak berubah** — data historis preserved
- Angsuran yang belum dibayar (SCHEDULED, OVERDUE, PARTIAL) → status VOIDED
- Jadwal baru dimulai dengan sisa pokok outstanding
- Jadwal baru bisa menggunakan terms berbeda (rate, tenor, metode) sesuai approval restrukturisasi
- Nomor installment di jadwal baru dimulai dari 1 (bukan melanjutkan nomor lama)
- Link ke `restructured_from` di pinjaman menghubungkan jadwal lama dan baru

### 13. Reconciliation — Jadwal vs Aktual

Scheduled job untuk membandingkan angsuran terjadwal vs pembayaran aktual:

```
Reconciliation Job (bulanan):
+-----------------------------------+
| Per pinjaman:                     |
| 1. SUM(total_amount) all scheduled|
|    = expected                     |
| 2. SUM(total_paid) all angsuran   |
|    = actual                       |
| 3. Outstanding pinjaman           |
|    = principal - SUM(paid_principal)|
| 4. Compare: rekening.balance      |
|    vs calculated outstanding      |
+-----------------------------------+
        |
        v
+-----------------------------------+
| Match?                            |
| Ya -> OK, log reconciled          |
| Tidak -> ALERT, flag for review   |
+-----------------------------------+
```

**Aturan:**
- Reconciliation berjalan **bulanan** (setelah cut-off tanggal)
- Membandingkan:
  - SUM `paid_principal` di angsuran vs perubahan `outstanding_principal` di pinjaman
  - `rekening.balance` vs calculated outstanding
- Selisih yang ditemukan di-flag untuk review manual oleh Supervisor
- Reconciliation report tersedia per cabang dan per tenant
- Selisih > threshold (configurable, default: Rp 1) otomatis create **alert** ke Manager

### 14. Angsuran Pembayaran — Data Model

Setiap pembayaran dicatat di tabel terpisah untuk audit trail detail:

```
angsuran_pembayaran
+-- id                    UUID v7 (PK)
+-- tenant_id             UUID (FK -> tenant)
+-- angsuran_id           UUID (FK -> angsuran) NOT NULL
+-- pinjaman_id           UUID (FK -> pinjaman) NOT NULL
+-- transaction_id        UUID (FK -> transaksi) NOT NULL
|
+-- -- Detail Pembayaran --
+-- payment_date          DATE NOT NULL
+-- payment_amount        NUMERIC(15,2) NOT NULL (total yang dibayar)
+-- allocated_principal   NUMERIC(15,2) DEFAULT 0
+-- allocated_interest    NUMERIC(15,2) DEFAULT 0
+-- allocated_margin      NUMERIC(15,2) DEFAULT 0
+-- allocated_profit_share NUMERIC(15,2) DEFAULT 0
+-- allocated_ujrah       NUMERIC(15,2) DEFAULT 0
+-- allocated_penalty     NUMERIC(15,2) DEFAULT 0
|
+-- -- Channel --
+-- payment_channel       ENUM (teller_cash, auto_debit, payroll)
+-- payment_reference     VARCHAR (nullable, nomor referensi)
|
+-- -- Overpayment --
+-- overpayment_amount    NUMERIC(15,2) DEFAULT 0
+-- overpayment_handling  ENUM (next_installment, reduce_principal) nullable
|
+-- -- Vernon Fields --
+-- _rels                 JSONB NOT NULL DEFAULT '{}'
+-- _data                 JSONB NOT NULL DEFAULT '{}'
|
+-- -- Audit --
    +-- created_at        TIMESTAMPTZ
    +-- created_by        UUID
    +-- updated_at        TIMESTAMPTZ
    +-- updated_by        UUID
```

### 15. Vernon _rels dan _data Structure

**Angsuran _rels:**
```json
{
  "pinjaman_id":  "018f...",
  "rekening_id":  "018f...",
  "tenant_id":    "018f..."
}
```

**Angsuran _data:**
```json
{
  "pinjaman": {
    "id":          "018f...",
    "loan_number": "PJ-2026-JKT-00000001",
    "akad_type":   "murabahah"
  },
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  }
}
```

**SyncEngine triggers:**
- `PinjamanUpdatedEvent` -> update `_data.pinjaman` di semua angsuran pinjaman tersebut
- `NasabahUpdatedEvent` -> update `_data.nasabah` di semua angsuran nasabah tersebut

### 16. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View jadwal angsuran | v | v | v | v |
| Record pembayaran angsuran | v | v | v | v |
| View aging report | v | v | v | v |
| Setup auto-debit | v | v | v | v |
| Cancel auto-debit | - | v | v | v |
| Setup payroll deduction | - | v | v | v |
| Override payment allocation (manual match) | - | v | v | v |
| Waive angsuran (mark as waived) | - | - | v | v |
| Adjust profit share (Musyarakah/Mudharabah) | - | v | v | v |
| View reconciliation report | - | v | v | v |
| Resolve reconciliation discrepancy | - | - | v | v |

**Catatan:**
- Teller bisa record pembayaran dan setup auto-debit — operasi rutin harian
- Override payment allocation (pilih angsuran mana yang dibayar, bukan FIFO) membutuhkan Supervisor+
- Waive angsuran = menghapuskan kewajiban (biasanya bagian dari restrukturisasi) — Manager+ only
- Reconciliation discrepancy resolution membutuhkan Manager+ karena impact ke laporan keuangan

### 17. Dual-Mode Terminology

| Konsep | `coop_type = "general"` | `coop_type = "islamic"` |
|--------|------------------------|------------------------|
| Angsuran | Angsuran | Angsuran |
| Jadwal | Jadwal Angsuran | Jadwal Angsuran |
| Porsi bunga/margin | Bunga | Margin (Murabahah) / Bagi Hasil (Musyarakah) / Ujrah (Ijarah) |
| Porsi pokok | Pokok | Pokok |
| Sisa pokok | Sisa Pinjaman | Sisa Pembiayaan |
| Pembayaran | Pembayaran Angsuran | Pembayaran Angsuran |
| Keterlambatan | Tunggakan | Tunggakan |
| Pelunasan dini | Pelunasan Dipercepat | Pelunasan Dipercepat + Ibra |
| Potong gaji | Potongan Gaji | Potongan Gaji |
| Auto-debit | Pendebitan Otomatis | Pendebitan Otomatis |
| Overpayment | Kelebihan Bayar | Kelebihan Bayar |
| Underpayment | Kurang Bayar | Kurang Bayar |

## Consequences

### Positif

- **Calculation engine komprehensif** — mendukung 3 metode konvensional + 4 akad syariah dalam satu sistem
- **Payment flexible** — mendukung full, partial, overpayment, dan prepayment
- **Multi-channel** — Teller, auto-debit, payroll deduction memudahkan nasabah
- **Auditable** — setiap pembayaran di-link ke angsuran spesifik dan transaksi
- **DPD tracking akurat** — per angsuran, bukan per pinjaman — granularity tinggi untuk NPL
- **Reconciliation built-in** — deteksi inkonsistensi otomatis antara jadwal dan aktual
- **School-integrated** — payroll deduction untuk guru/staff terintegrasi dengan sistem sekolah

### Negatif

- **Calculation complexity** — 7 metode perhitungan berbeda membutuhkan testing ekstensif
- **Scheduled jobs** — DPD update harian + reconciliation bulanan menambah beban operasional
- **Payroll integration** — ketergantungan pada sistem sekolah untuk potong gaji
- **Bagi hasil variability** — Musyarakah/Mudharabah membutuhkan review berkala profit aktual
- **Rounding differences** — pembulatan di kalkulasi angsuran bisa menghasilkan selisih kecil di angsuran terakhir

### Mitigasi

- Calculation engine di-implement sebagai **strategy pattern** — satu class per metode, unit test per class
- DPD update job ringan (hanya UPDATE rows yang berubah) dan berjalan di off-peak hours
- Payroll integration menggunakan **file exchange** (CSV/API) — tidak hard-dependency
- Bagi hasil review di-schedule otomatis dengan reminder ke Supervisor
- Rounding adjustment diterapkan di **angsuran terakhir** — selisih ditambahkan/dikurangi dari angsuran terakhir agar total tepat

## Alternatives Considered

### A. Single Calculation Method

Hanya mendukung satu metode perhitungan (misal: flat saja).

**Ditolak** karena: koperasi membutuhkan fleksibilitas metode perhitungan sesuai jenis produk. Anuitas umum untuk pinjaman besar, flat untuk pinjaman kecil. BMT wajib menggunakan akad syariah dengan perhitungan yang berbeda.

### B. No Auto-debit

Semua pembayaran hanya via Teller.

**Ditolak** karena: auto-debit dan payroll deduction signifikan mengurangi tunggakan. Guru/staff yang sibuk mengajar sering lupa bayar angsuran — potong gaji otomatis menyelesaikan masalah ini.

### C. DPD dari Payment Date (bukan Due Date)

Menghitung DPD berdasarkan tanggal pembayaran terakhir, bukan tanggal jatuh tempo.

**Ditolak** karena: standar OJK menghitung DPD berdasarkan **due date** angsuran yang tertunggak. Payment date hanya relevan untuk reset DPD setelah pelunasan. Menggunakan payment date bisa menghasilkan DPD yang lebih rendah dari seharusnya.

### D. Real-time Reconciliation (tanpa scheduled job)

Reconciliation dilakukan setiap kali ada pembayaran.

**Ditolak** karena: menambah latency di setiap transaksi pembayaran. Reconciliation bulanan cukup untuk deteksi anomali — selisih intraday tidak material untuk koperasi. Kasus kritis (balance mismatch) sudah dicegah oleh atomic transaction saat pembayaran.

### E. Separate Table untuk Pembayaran (payment vs schedule)

Jadwal dan pembayaran di tabel terpisah, di-join saat query.

**Ditolak** karena: menambah complexity JOIN dan menyulitkan query "angsuran mana yang sudah dibayar". Menyimpan `paid_*` fields di tabel angsuran lebih sederhana dan performant. Audit trail detail sudah tercatat di tabel transaksi (K011) dan `angsuran_pembayaran`.
