# ADR-K005: Tabungan (Simpanan Sukarela)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Tabungan adalah **produk paling aktif** di koperasi sekolah — nasabah (guru, siswa, orang tua) paling sering bertransaksi melalui tabungan. Berbeda dari simpanan pokok dan wajib (K004) yang bersifat mandatory dan locked, tabungan bersifat **sukarela** dengan fleksibilitas setoran dan penarikan.

Tantangan di lingkungan sekolah:
- **Multi-produk**: satu koperasi bisa punya Tabungan Reguler, Tabungan Pendidikan, Tabungan Qurban — masing-masing dengan aturan berbeda
- **Siswa**: nominal kecil, setoran mingguan via wali kelas, penarikan dibatasi
- **Goal-based saving**: tabungan dengan target (wisata, pendidikan, qurban) punya aturan khusus
- **Dual-mode**: general menggunakan bunga harian, BMT menggunakan bagi hasil berdasarkan nisbah
- **Parent-child**: orang tua bisa setor ke tabungan anak
- **High volume**: tabungan reguler bisa menghasilkan ratusan transaksi per hari di koperasi aktif

Setiap tabungan dibuka berdasarkan **produk** dari K003 dengan kategori TABUNGAN, dan menggunakan **rekening** dari K002.

## Decision

### 1. Multiple Tabungan Products

Koperasi bisa menawarkan beberapa jenis tabungan, masing-masing sebagai produk terpisah di K003:

```
┌─────────────────────────────────────────────────────────────┐
│                   JENIS TABUNGAN                             │
├───────────────────┬─────────────────────────────────────────┤
│ Produk            │ Karakteristik                           │
├───────────────────┼─────────────────────────────────────────┤
│ Tabungan Reguler  │ Umum, semua nasabah, tarik kapan saja. │
│                   │ Bunga/bagi hasil standar.               │
├───────────────────┼─────────────────────────────────────────┤
│ Tabungan          │ Khusus siswa/santri. Nominal kecil.     │
│ Pendidikan        │ Target: biaya sekolah. Penarikan        │
│                   │ terbatas (hanya awal semester).         │
├───────────────────┼─────────────────────────────────────────┤
│ Tabungan          │ Target: Idul Fitri/Natal. Penarikan    │
│ Hari Raya         │ hanya menjelang hari raya               │
│                   │ (configurable window).                  │
├───────────────────┼─────────────────────────────────────────┤
│ Tabungan Qurban   │ BMT mode. Target: pembelian hewan       │
│                   │ qurban. Penarikan hanya menjelang       │
│                   │ Idul Adha.                              │
├───────────────────┼─────────────────────────────────────────┤
│ Tabungan Wisata   │ Goal-based. Target tanggal & nominal.   │
│                   │ Auto-debit bulanan. Penalty jika        │
│                   │ tarik sebelum target.                   │
└───────────────────┴─────────────────────────────────────────┘
```

**Aturan:**
- Setiap jenis tabungan adalah **produk** di K003 dengan `category = TABUNGAN`
- Satu nasabah bisa punya **multiple rekening tabungan** — bahkan beberapa dari produk yang sama
- Nasabah membuka tabungan melalui **pengajuan rekening** (K002 Section 3) — Supervisor+ approve
- Setiap produk tabungan punya rate, fee, dan limit sendiri (didefinisikan di K003)
- Tenant memilih produk tabungan mana yang ditawarkan — tidak wajib semua jenis

### 2. Setoran (Deposit) Rules

```
Setoran Tabungan
        │
        v
┌───────────────────────────────────────┐
│ Validasi:                             │
│ 1. Rekening status = ACTIVE           │
│ 2. amount ≥ min_deposit (produk)      │
│ 3. amount ≤ max_deposit_per_trx       │
│ 4. Rekening bukan FROZEN? → OK,       │
│    FROZEN masih boleh setoran (K002)  │
└───────────────────────────────────────┘
        │ valid
        v
┌───────────────────────────────────────┐
│ Record transaksi (K011):              │
│ - type: DEPOSIT                       │
│ - amount: nominal setoran             │
│ - balance_after: balance + amount     │
│ - channel: teller/auto_debit/batch    │
└───────────────────────────────────────┘
```

**Channels setoran:**
- **Teller cash**: nasabah setor langsung di counter
- **Transfer**: dari rekening bank atau tabungan lain di koperasi
- **Batch wali kelas**: guru kumpulkan dari siswa, setor batch
- **Auto-debit**: pindah buku dari tabungan lain (misal: tabungan reguler → tabungan qurban)
- **Parent deposit**: orang tua setor ke tabungan anak

**Aturan:**
- `min_deposit` dan `max_deposit_per_trx` dari produk (K003 Section 8)
- Setoran ke rekening FROZEN **diizinkan** (K002 Section 7) — dana masuk tapi tidak bisa ditarik
- Setoran oleh orang tua ke tabungan anak: `deposited_by_nasabah_id` dicatat di transaksi
- Tidak ada batas frekuensi setoran per hari (kecuali tenant mengonfigurasi)

### 3. Penarikan (Withdrawal) Rules

```
Penarikan Tabungan
        │
        v
┌───────────────────────────────────────┐
│ Validasi:                             │
│ 1. Rekening status = ACTIVE           │
│    (FROZEN → ditolak untuk penarikan) │
│ 2. amount ≤ available_balance         │
│ 3. (balance - amount) ≥ min_balance   │
│ 4. amount ≤ max_withdrawal_daily      │
│    (cek akumulasi hari ini)           │
│ 5. sum_month + amount ≤              │
│    max_withdrawal_monthly             │
│ 6. KYC limit check (K001 Section 11) │
│ 7. [> threshold] → verifikasi         │
│    specimen tanda tangan (K001)       │
│ 8. [> manager_threshold] → approval   │
│    Manager                            │
└───────────────────────────────────────┘
        │ valid
        v
┌───────────────────────────────────────┐
│ Record transaksi (K011):              │
│ - type: WITHDRAWAL                    │
│ - amount: nominal penarikan           │
│ - balance_after: balance - amount     │
└───────────────────────────────────────┘
```

**Threshold verification:**
```
withdrawal_verification:
  specimen_check_threshold:   1000000     # Rp 1jt → teller verifikasi tanda tangan
  manager_approval_threshold: 5000000     # Rp 5jt → perlu approval Manager
  # Kedua threshold configurable per tenant (K001 Section 10)
```

**Aturan:**
- `min_balance` dari produk — penarikan tidak boleh menyebabkan saldo di bawah minimum
- `max_withdrawal_daily` dan `max_withdrawal_monthly` dari produk (K003 Section 8)
- Daily/monthly limit dihitung dari **akumulasi** penarikan hari/bulan ini (bukan per transaksi saja)
- KYC level `basic` memiliki **limit transaksi bulanan** (K001 Section 11) — penarikan termasuk
- Penarikan di atas `specimen_check_threshold` → teller wajib verifikasi tanda tangan (K001 Section 10)
- Penarikan di atas `manager_approval_threshold` → perlu approval Manager real-time
- Rekening FROZEN → **penarikan diblokir** (K002 Section 7)

### 4. Interest / Profit Calculation

#### 4a. General Mode — Bunga Harian (Daily Balance Method)

```
Perhitungan bunga tabungan (daily balance):

Hari ke-1 s/d 10:  Saldo = Rp 5.000.000
Hari ke-11 s/d 20: Saldo = Rp 7.000.000 (setor Rp 2jt)
Hari ke-21 s/d 31: Saldo = Rp 6.000.000 (tarik Rp 1jt)

Rate: 3% p.a.

Bunga hari 1-10:  5.000.000 × (3/100) × (10/365) = Rp 4.110
Bunga hari 11-20: 7.000.000 × (3/100) × (10/365) = Rp 5.753
Bunga hari 21-31: 6.000.000 × (3/100) × (11/365) = Rp 5.425
─────────────────────────────────────────────────────────
Total bunga bulan ini:                        Rp 15.288
PPh 10%:                                      Rp  1.529
Bunga bersih (diposting ke rekening):         Rp 13.759
```

**Aturan:**
- Bunga dihitung **harian** berdasarkan saldo akhir hari
- Posting bunga ke rekening: tanggal configurable per produk (`interest_posting_day`, default: akhir bulan)
- PPh (Pajak Penghasilan) atas bunga: rate configurable per tenant (default: 10% sesuai regulasi)
- Scheduled job menjalankan perhitungan di tanggal posting
- Bunga yang diposting menjadi **transaksi masuk** di rekening tabungan (tipe: `INTEREST_CREDIT`)

#### 4b. Islamic Mode — Bagi Hasil Bulanan

```
Perhitungan bagi hasil (Mudharabah):

1. Hitung rata-rata saldo harian nasabah bulan ini
   Avg saldo nasabah A = Rp 5.500.000

2. Hitung total rata-rata saldo seluruh nasabah tabungan Mudharabah
   Total avg = Rp 500.000.000

3. Porsi nasabah A = 5.500.000 / 500.000.000 = 1.1%

4. Profit koperasi bulan ini (dari penyaluran pembiayaan)
   Profit = Rp 25.000.000

5. Nisbah nasabah = 40% (dari produk)
   Bagian nasabah total = 25.000.000 × 40% = Rp 10.000.000

6. Bagi hasil nasabah A = 10.000.000 × 1.1% = Rp 110.000
   PPh 10% = Rp 11.000
   Bagi hasil bersih = Rp 99.000
```

**Aturan:**
- Bagi hasil dihitung **bulanan** berdasarkan rata-rata saldo harian nasabah
- **Profit pool**: total keuntungan koperasi dari penyaluran dana tabungan Mudharabah
- Nisbah dari produk (K003 Section 3b) — misal 40:60 (nasabah:koperasi)
- `indicative_rate` di produk hanyalah **estimasi** — bagi hasil aktual tergantung profit riil
- Bagi hasil diposting sebagai transaksi `PROFIT_SHARE_CREDIT`
- Untuk akad **Wadiah**: tidak ada bagi hasil wajib — koperasi bisa memberikan bonus (athaya) atas kebijakan sendiri

**Fallback & Negative Profit Handling:**

Bagi hasil Mudharabah bergantung pada profit pool dari K015 (Akuntansi). Penanganan kasus khusus:

1. **Profit pool belum final**: Jika perhitungan profit bulan ini dari K015 belum selesai saat jadwal posting bagi hasil:
   - Gunakan **indicative rate** dari produk sebagai estimasi sementara
   - Posting bagi hasil berdasarkan indicative rate
   - Saat profit pool final, hitung **adjustment** (selisih estimasi vs aktual)
   - Adjustment di-posting di bulan berikutnya sebagai koreksi

2. **Profit pool negatif** (koperasi rugi): Per prinsip Mudharabah, nasabah ikut menanggung kerugian:
   - Bagi hasil bulan tersebut = Rp 0 (tidak ada distribusi)
   - Kerugian **tidak mengurangi saldo pokok** nasabah (kerugian ditanggung dari potensi keuntungan, bukan modal)
   - Nasabah diinformasikan via notifikasi bahwa bagi hasil bulan ini nihil
   - Jika rugi berturut-turut > 3 bulan, Manager wajib evaluasi dan informasikan ke nasabah

3. **Profit pool sangat kecil**: Jika bagi hasil per nasabah < Rp 1 setelah perhitungan:
   - Dibulatkan ke Rp 0 untuk nasabah tersebut
   - Akumulasi pembulatan dialokasikan ke cadangan koperasi

### 5. Tabungan Berencana / Goal-Based Savings

Tabungan dengan target nominal dan/atau tanggal:

```
goal_config (JSONB di rekening):
├── target_amount          NUMERIC(15,2)   Rp 5.000.000
├── target_date            DATE            2026-12-15
├── monthly_auto_debit     NUMERIC(15,2)   Rp 500.000
├── auto_debit_day         INTEGER         5 (tanggal 5 setiap bulan)
├── source_rekening_id     UUID            (tabungan reguler untuk auto-debit)
├── early_withdrawal_penalty_pct  NUMERIC(5,2)  2.00 (2% dari saldo)
└── allow_partial_withdrawal BOOLEAN       false
```

**Flow:**
```
Buka Tabungan Berencana
        │
        v
┌───────────────────────────────────────┐
│ Set target: Rp 5jt, 15 Des 2026      │
│ Auto-debit: Rp 500rb/bulan           │
│ dari Tabungan Reguler                 │
└───────────────────────────────────────┘
        │
        v (setiap tanggal 5)
┌───────────────────────────────────────┐
│ Auto-debit Rp 500rb dari Tab. Reguler │
│ → credit ke Tabungan Berencana        │
└───────────────────────────────────────┘
        │
        v (target tercapai / tanggal tiba)
┌───────────────────────────────────────┐
│ Notifikasi: "Target tercapai!"        │
│ Nasabah bisa cairkan penuh            │
│ Atau lanjutkan menabung               │
└───────────────────────────────────────┘

Jika tarik sebelum target:
┌───────────────────────────────────────┐
│ Penalty 2% dari saldo                 │
│ + Bunga/bagi hasil yang sudah         │
│   diposting tetap milik nasabah       │
│ Membutuhkan approval Supervisor+      │
└───────────────────────────────────────┘
```

**Aturan:**
- Goal config opsional — tidak semua tabungan punya target
- Auto-debit mengikuti rules dari K004 Section 10 (shared mechanism)
- Early withdrawal penalty dari produk (`early_withdrawal_pct` di K003 fee_config)
- Partial withdrawal **tidak diizinkan** untuk tabungan berencana (default) — harus full withdrawal
- Target date bersifat informasional — tidak auto-close saat target tercapai
- Progress tracking: `(current_balance / target_amount) × 100%` ditampilkan di dashboard nasabah

### 6. Student Savings — Tabungan Siswa

Fitur khusus untuk nasabah siswa/santri:

```
student_savings_config:
  min_deposit:               1000      # Rp 1.000 (sangat kecil)
  min_balance:               5000      # Rp 5.000
  max_withdrawal_daily:     200000     # Rp 200.000
  max_withdrawal_monthly:   500000     # Rp 500.000
  collection_mode:          "weekly"   # Wali kelas kumpulkan mingguan
  educational_goals:
    - "Tabungan Pendidikan"            # Target: biaya sekolah
    - "Tabungan Wisata Sekolah"        # Target: dana wisata
```

**Weekly collection flow:**
```
Senin (wali kelas di kelas)
        │
        v
┌───────────────────────────────────────┐
│ Wali kelas catat:                     │
│ - Ahmad: Rp 5.000                     │
│ - Budi:  Rp 10.000                    │
│ - Citra: Rp 3.000                     │
│ - Dani:  (tidak setor)               │
└───────────────────────────────────────┘
        │
        v (akhir minggu / akhir bulan)
┌───────────────────────────────────────┐
│ Wali kelas setor batch ke koperasi    │
│ Total: Rp 18.000                      │
│ Daftar: 3 siswa, masing-masing       │
│ dengan nominal yang dicatat           │
└───────────────────────────────────────┘
        │
        v
┌───────────────────────────────────────┐
│ Teller proses batch:                  │
│ - Validasi total = sum per siswa      │
│ - Generate transaksi per siswa        │
│ - Update saldo per rekening           │
└───────────────────────────────────────┘
```

**Aturan:**
- Nominal minimum setoran **sangat kecil** — mengakomodasi siswa SD/MI
- Limit penarikan **lebih ketat** — mencegah penyalahgunaan
- Wali kelas menggunakan fitur **batch deposit** — input daftar siswa + nominal
- Orang tua bisa setor langsung ke tabungan anak via teller atau transfer
- Educational goal products (Tabungan Pendidikan) bisa membatasi penarikan hanya di periode tertentu (awal semester)

### 7. Parent-Child Deposit Linking

Orang tua bisa menyetor ke tabungan anak:

```
Parent Deposit to Child Account
        │
        v
┌───────────────────────────────────────┐
│ Teller:                               │
│ 1. Pilih rekening anak (nasabah)      │
│ 2. Input nominal setoran              │
│ 3. Pilih sumber: "Setoran oleh pihak │
│    lain" → pilih nasabah orang tua   │
│ 4. Proses setoran                     │
└───────────────────────────────────────┘
        │
        v
┌───────────────────────────────────────┐
│ Transaksi:                            │
│ - rekening_id: tabungan anak          │
│ - amount: nominal setoran             │
│ - deposited_by_nasabah_id: ID ortu    │
│ - channel: teller_cash                │
│ - notes: "Setoran oleh Bp. Ahmad"     │
└───────────────────────────────────────┘
```

**Aturan:**
- Setoran oleh pihak lain dicatat di **rekening penerima** (anak)
- `deposited_by_nasabah_id` mencatat siapa yang menyetor — untuk audit trail
- **Tidak perlu family link formal** — cukup referensi nasabah_id penyetor
- Siapa saja bisa menyetor ke rekening siapa saja (guru setor untuk siswa, orang tua untuk anak) — yang di-track adalah siapa yang melakukan setoran
- Penarikan **hanya bisa dilakukan oleh pemilik rekening** atau wali (jika siswa di bawah umur, wali kelas/orang tua dengan kuasa)

### 8. Auto-Debit untuk Simpanan Wajib

Tabungan reguler sering menjadi **sumber dana** untuk auto-debit simpanan wajib (K004 Section 10):

```
Priority auto-debit dari tabungan:
┌──────────────────────────────────────┐
│ Priority 1: Simpanan Wajib bulanan   │
│ Priority 2: Tabungan Berencana       │
│ Priority 3: Lainnya (jika ada)       │
└──────────────────────────────────────┘
```

**Aturan:**
- Auto-debit simpanan wajib **diprioritaskan** di atas auto-debit lainnya
- Jika `available_balance` tidak cukup untuk semua auto-debit, yang priority lebih tinggi dieksekusi duluan
- Saldo setelah semua auto-debit tidak boleh < `min_balance` produk tabungan
- Nasabah bisa memilih tabungan mana yang dijadikan sumber auto-debit (jika punya multiple)
- Detail mekanisme auto-debit di K004 Section 10

### 9. Statement Generation

Nasabah bisa mendapatkan rekening koran (statement) tabungan:

```
statement_config:
  auto_generate:          monthly        # Otomatis setiap akhir bulan
  on_demand:              true           # Bisa request kapan saja
  format:                 PDF
  delivery:               in_app         # Download di aplikasi
  retention_months:       24             # Simpan 24 bulan terakhir
```

**Isi statement:**
```
┌─────────────────────────────────────────────────────────────┐
│           REKENING KORAN TABUNGAN                           │
│  Koperasi Al-Ikhlas                                         │
├─────────────────────────────────────────────────────────────┤
│  Nasabah: Ahmad Fauzi (KOP-2026-JKT-000001)               │
│  Rekening: TB-2026-JKT-00000001                            │
│  Produk: Tabungan Berkah (Mudharabah)                      │
│  Periode: 01 April 2026 - 30 April 2026                   │
├─────────────────────────────────────────────────────────────┤
│  Saldo Awal:                              Rp  5.000.000   │
├──────┬───────────┬──────────┬──────────┬───────────────────┤
│ Tgl  │ Keterangan│   Debit  │  Kredit  │     Saldo        │
├──────┼───────────┼──────────┼──────────┼───────────────────┤
│ 05/04│ Setoran   │          │  500.000 │  5.500.000       │
│ 10/04│ Penarikan │  200.000 │          │  5.300.000       │
│ 15/04│ Auto-debit│   50.000 │          │  5.250.000       │
│      │ S.Wajib   │          │          │                   │
│ 30/04│ Bagi Hasil│          │   99.000 │  5.349.000       │
├──────┴───────────┴──────────┴──────────┴───────────────────┤
│  Saldo Akhir:                             Rp  5.349.000   │
│  Bagi Hasil Bulan Ini:                    Rp     99.000   │
│  (Nisbah Nasabah: 40%, Eq. Rate: 2.4%)                    │
└─────────────────────────────────────────────────────────────┘
```

**Aturan:**
- Statement otomatis di-generate akhir bulan oleh scheduled job
- On-demand request melalui teller atau self-service (jika ada portal nasabah)
- Format PDF — header dan footer sesuai branding tenant
- Retention configurable per tenant (default: 24 bulan)
- BMT mode: menampilkan nisbah dan equivalent rate sebagai tambahan informasi

### 10. Transaction Limits per Product

Setiap produk tabungan bisa membatasi transaksi harian dan bulanan:

```
transaction_limits (dari produk K003):
├── max_deposit_per_trx          NUMERIC(15,2)   Per transaksi setoran
├── max_withdrawal_daily         NUMERIC(15,2)   Akumulasi penarikan per hari
├── max_withdrawal_monthly       NUMERIC(15,2)   Akumulasi penarikan per bulan
├── max_transactions_daily       INTEGER         Max jumlah transaksi per hari (nullable)
└── kyc_monthly_limit            NUMERIC(15,2)   Limit bulanan berdasarkan KYC (K001)
```

**Enforcement flow:**
```
Setiap transaksi
        │
        v
┌───────────────────────────────────────┐
│ 1. Cek per-transaction limit          │
│ 2. Cek daily accumulation             │
│ 3. Cek monthly accumulation           │
│ 4. Cek KYC monthly limit (K001)       │
│ 5. Cek max transactions daily count   │
│    │                                  │
│    ALL PASS → Proceed                 │
│    ANY FAIL → Reject with reason      │
└───────────────────────────────────────┘
```

**Aturan:**
- Limit di-enforce di **transaction layer** (K011)
- Daily accumulation di-reset setiap jam 00:00 (timezone tenant)
- Monthly accumulation di-reset setiap tanggal 1
- KYC limit dari K001 Section 11 — limit nasabah basic vs full
- `max_transactions_daily` nullable — null berarti unlimited frequency
- Limit **per rekening**, bukan per nasabah — nasabah dengan multiple tabungan punya limit terpisah per rekening

### 11. Dormant Handling

Tabungan yang tidak ada transaksi selama periode tertentu di-flag **dormant** (referensi K002 Section 15):

```
Dormant flow (dari K002):
├── Tidak ada transaksi selama dormant_period (default: 12 bulan)
│   → Flag: is_dormant = true, dormant_since = date
│
├── Jika dormant_fee_enabled (dari produk K003):
│   → Potong dormant_fee_monthly dari saldo setiap bulan
│   → Transaksi tipe: DORMANT_FEE
│
├── Jika saldo habis karena dormant fee:
│   → auto_close_after_months (configurable, default: 36 bulan)
│   → Rekening ditutup otomatis
│
└── Jika ada transaksi masuk:
    → Flag dihapus: is_dormant = false
    → Kembali normal, dormant fee berhenti
```

**Aturan:**
- Dormant logic didefinisikan di K002 Section 15 — di sini hanya referensi
- Tabungan Pendidikan / Hari Raya / Qurban bisa **dikecualikan** dari dormant check (configurable per produk) karena memang jarang ditransaksikan
- Dormant fee dari produk (K003 `fee_config.dormant_fee_monthly`)
- Notifikasi dikirim 30 hari sebelum flag dormant — nasabah bisa setor untuk mencegah

### 12. Data Model

Tabungan menggunakan tabel **rekening** dari K002 dengan `category = tabungan`. Tidak ada tabel terpisah untuk tabungan — yang spesifik hanya konfigurasi goal-based dan statement.

#### 12a. Goal-based Config (di tabel rekening)

Goal-based savings disimpan di kolom JSONB di tabel rekening:

```
rekening (tambahan field untuk goal-based):
├── goal_config               JSONB (nullable)
│   {
│     "target_amount":           5000000,
│     "target_date":            "2026-12-15",
│     "monthly_auto_debit":      500000,
│     "auto_debit_day":          5,
│     "source_rekening_id":     "018f...",
│     "early_withdrawal_penalty_pct": 2.00,
│     "allow_partial_withdrawal": false
│   }
```

Catatan: field `goal_config` sudah di-accommodate di tabel `rekening` (K002) sebagai JSONB — tidak perlu ALTER TABLE karena JSONB bersifat schemaless.

#### 12b. Tabel `tabungan_statement`

```
tabungan_statement
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── rekening_id           UUID (FK → rekening)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Periode ──
├── period_start          DATE NOT NULL
├── period_end            DATE NOT NULL
├── period_label          VARCHAR (misal: "April 2026")
│
├── ── Ringkasan ──
├── opening_balance       NUMERIC(15,2) NOT NULL
├── closing_balance       NUMERIC(15,2) NOT NULL
├── total_deposits        NUMERIC(15,2) NOT NULL
├── total_withdrawals     NUMERIC(15,2) NOT NULL
├── total_interest        NUMERIC(15,2) NOT NULL (bunga/bagi hasil)
├── total_fees            NUMERIC(15,2) NOT NULL (admin fee, dll)
├── transaction_count     INTEGER NOT NULL
│
├── ── Document ──
├── document_url          VARCHAR (URL ke PDF yang di-generate)
├── generated_at          TIMESTAMPTZ
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

#### 12c. Vernon _rels dan _data Structure

**tabungan_statement _rels:**
```json
{
  "tenant_id":    "018f...",
  "rekening_id":  "018f...",
  "nasabah_id":   "018f..."
}
```

**tabungan_statement _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "rekening": {
    "id":             "018f...",
    "account_number": "TB-2026-JKT-00000001",
    "product_name":   "Tabungan Berkah"
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.nasabah` di tabungan_statement
- `RekeningUpdatedEvent` → update `_data.rekening` di tabungan_statement
- Statement yang sudah di-generate **tidak** di-update — snapshot immutable, hanya _data untuk listing

### 13. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View saldo & transaksi tabungan | v | v | v | v |
| Setoran tabungan (teller cash) | v | v | v | v |
| Penarikan tabungan (dalam limit) | v | v | v | v |
| Penarikan di atas specimen threshold | v + verifikasi | v | v | v |
| Penarikan di atas manager threshold | - | - | v | v |
| Batch deposit (wali kelas) | v | v | v | v |
| Parent deposit ke anak | v | v | v | v |
| Setup goal-based config | v | v | v | v |
| Setup auto-debit dari tabungan | v | v | v | v |
| Generate statement on-demand | v | v | v | v |
| Early withdrawal tabungan berencana | - | v | v | v |
| Override transaction limits | - | - | v | v |
| View dormant accounts list | - | v | v | v |
| Reactivate dormant account | v | v | v | v |

**Catatan:**
- Teller bisa handle semua operasi day-to-day (setoran, penarikan dalam limit, batch)
- Penarikan besar membutuhkan **verifikasi specimen** (Teller verifikasi visual) atau **approval Manager**
- Early withdrawal tabungan berencana butuh **Supervisor+** karena ada penalty
- Override limit hanya **Manager+** — harus ada alasan dan di-log

### 14. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Entity name | Tabungan | Tabungan |
| Product example | Tabungan Reguler | Tabungan Mudharabah |
| Yield label | Bunga | Bagi Hasil |
| Yield calculation | Bunga harian (daily balance) | Bagi hasil bulanan (nisbah × profit) |
| Yield posting | Bunga diposting bulanan | Bagi hasil diposting bulanan |
| Tax label | PPh atas Bunga | PPh atas Bagi Hasil |
| Wadiah product | N/A | Tabungan Wadiah (bonus sukarela) |
| Statement yield line | Bunga Bulan Ini | Bagi Hasil Bulan Ini |
| Goal-based | Tabungan Berencana | Tabungan Berencana / Tabungan Haji |
| Qurban product | N/A | Tabungan Qurban (akad Mudharabah) |

## Consequences

### Positif

- **Most-used product** — tabungan reguler melayani semua nasabah dengan fleksibilitas tinggi
- **Student-friendly** — weekly collection, nominal kecil, wali kelas batch deposit
- **Goal-oriented** — tabungan berencana memotivasi nasabah menabung dengan target jelas
- **Parent-child** — orang tua bisa menyetor langsung ke tabungan anak tanpa family link formal
- **Dual-mode native** — bunga dan bagi hasil terdefinisi jelas dengan formula masing-masing
- **Statement** — nasabah punya bukti transaksi formal, koperasi punya audit trail
- **Auto-debit integrated** — satu tabungan bisa menjadi sumber dana untuk simpanan wajib dan tabungan berencana

### Negatif

- **High transaction volume** — tabungan reguler bisa menghasilkan ratusan transaksi/hari, perlu optimasi database
- **Bagi hasil complexity** — kalkulasi bagi hasil BMT membutuhkan profit pool yang akurat (dependency ke K015 akuntansi)
- **Multi-limit enforcement** — per-transaction, daily, monthly, KYC, specimen — banyak layer validasi
- **Statement storage** — PDF statement per nasabah per bulan membutuhkan storage management
- **Dormant fee edge case** — auto-deduction dormant fee bisa menyebabkan saldo negatif jika timing salah

### Mitigasi

- Transaction table di-partitioned per bulan — query performance tetap optimal
- Bagi hasil dihitung dari **profit yang sudah di-posting** di jurnal (K015) — bukan estimasi
- Limit enforcement di-centralize di satu middleware/service — tidak tersebar di banyak handler
- Statement PDF di-generate async dan disimpan di object storage — tidak membebani database
- Dormant fee di-skip jika saldo < fee amount — tidak pernah menghasilkan saldo negatif

## Alternatives Considered

### A. Single Tabungan Product (tanpa multi-produk)

Satu jenis tabungan saja untuk semua nasabah.

**Ditolak karena:** koperasi sekolah membutuhkan tabungan dengan tujuan berbeda (pendidikan, qurban, wisata). Single product tidak bisa mengakomodasi perbedaan aturan penarikan, rate, dan limit per jenis tabungan.

### B. Monthly Balance Method (bukan Daily Balance)

Bunga dihitung dari saldo rata-rata bulanan, bukan saldo harian.

**Ditolak karena:** kurang adil untuk nasabah — tidak memperhitungkan fluktuasi saldo dalam bulan. Daily balance lebih akurat dan merupakan standar industri perbankan/koperasi.

### C. Goal-based sebagai Tabel Terpisah

Tabel `tabungan_goal` terpisah dari rekening.

**Ditolak karena:** over-engineering. Goal config hanya beberapa field — cukup disimpan sebagai JSONB di rekening. Tabel terpisah menambah JOIN tanpa benefit signifikan. JSONB sudah di-validate di application layer.

### D. Parent-Child via Family Link Table

Setoran orang tua ke anak mewajibkan relasi `nasabah_family_link`.

**Ditolak karena:** family link ditunda ke Phase 2 (K001 Section 16). Untuk setoran, cukup mencatat `deposited_by_nasabah_id` — siapa saja bisa menyetor ke rekening siapa saja. Family link hanya menambah complexity tanpa mengubah mekanisme setoran.
