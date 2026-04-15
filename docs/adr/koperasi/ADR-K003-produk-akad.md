# ADR-K003: Produk & Akad

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Setiap rekening di koperasi/BMT harus terikat pada **produk** — produk mendefinisikan aturan bisnis (rate, tenor, biaya, limit) yang berlaku untuk rekening tersebut. Tanpa katalog produk yang terstruktur, konfigurasi bisnis tersebar di banyak tempat dan sulit di-maintain.

Sistem harus mendukung **dual-mode**:
- **General (Konvensional)**: produk berbasis bunga (suku bunga tetap/menurun/anuitas)
- **Islamic (BMT)**: produk berbasis akad syariah — setiap produk wajib memiliki jenis akad yang menentukan skema bagi hasil, margin, atau ujrah

Produk adalah **fondasi** — ADR ini harus selesai sebelum K004 (Simpanan Pokok & Wajib), K005 (Tabungan), K006 (Deposito), dan K007 (Pinjaman) bisa diimplementasikan, karena semua merujuk ke produk.

## Decision

### 1. Product Catalog Structure

Produk adalah entitas yang mendefinisikan **aturan bisnis** untuk setiap kategori rekening. Satu tenant bisa memiliki banyak produk per kategori.

```
produk
├── Identitas
│   ├── code: "TB-BERKAH"        ← Kode unik per tenant
│   ├── name: "Tabungan Berkah"  ← Nama tampilan
│   └── description              ← Deskripsi lengkap
│
├── Klasifikasi
│   ├── category: TABUNGAN       ← Mapping ke rekening_category
│   └── coop_type: islamic       ← Mode koperasi (general | islamic)
│
├── Rate / Yield
│   ├── [General] interest_rate, calculation_method
│   └── [Islamic] akad_type, nisbah, margin_rate, ujrah_rate
│
├── Limit & Threshold
│   ├── min_balance, max_balance
│   ├── min_deposit, max_deposit
│   ├── min_withdrawal, max_withdrawal_daily
│   └── max_withdrawal_monthly
│
├── Biaya (Fees)
│   ├── admin_fee_monthly
│   ├── opening_fee
│   ├── closing_fee
│   └── penalty_rate
│
├── Eligibility
│   └── eligible_relations[]     ← school_relation_type yang boleh akses
│
└── Lifecycle
    ├── status: ACTIVE
    └── version: 3
```

### 2. Product Types Mapping ke Rekening Categories

Setiap produk wajib berada dalam satu kategori rekening. Mapping ini menentukan perilaku dasar produk:

| Kategori Rekening | Contoh Produk (General) | Contoh Produk (Islamic/BMT) | Akad Default |
|-------------------|------------------------|-----------------------------|--------------|
| SIMPANAN_POKOK | Simpanan Pokok Anggota | Simpanan Pokok Anggota | Wadiah Yad Dhamanah |
| SIMPANAN_WAJIB | Simpanan Wajib Bulanan | Simpanan Wajib Bulanan | Wadiah Yad Dhamanah |
| TABUNGAN | Tabungan Reguler, Tabungan Pendidikan | Tabungan Mudharabah, Tabungan Wadiah | Mudharabah / Wadiah |
| DEPOSITO | Deposito Berjangka 3 Bulan | Deposito Mudharabah 3 Bulan | Mudharabah |
| PINJAMAN | Pinjaman Reguler, Pinjaman Darurat | Pembiayaan Murabahah, Pembiayaan Musyarakah | Murabahah / Musyarakah / Ijarah / Qardh |

**Aturan:**
- Satu produk **tepat satu** kategori — tidak bisa lintas kategori
- Satu kategori bisa memiliki **banyak produk** (kecuali SIMPANAN_POKOK dan SIMPANAN_WAJIB yang biasanya 1 produk default per tenant)
- Produk pinjaman di BMT mode **wajib** memiliki `akad_type` — tidak boleh null
- Tenant bisa membuat produk custom selama sesuai dengan kategori yang ada

### 3. Rate / Yield Configuration

#### 3a. General Mode — Suku Bunga

```
interest_config (JSONB):
  rate:                  NUMERIC(7,4)    # Persentase per tahun (misal: 6.0000 = 6% p.a.)
  calculation_method:    ENUM
    ├── FLAT             # Bunga dihitung dari pokok awal (pinjaman)
    ├── DECLINING        # Bunga dihitung dari sisa pokok (pinjaman)
    ├── ANNUITY          # Angsuran tetap, komposisi pokok-bunga berubah (pinjaman)
    ├── DAILY_BALANCE    # Bunga harian berdasarkan saldo harian (tabungan)
    └── MONTHLY_BALANCE  # Bunga bulanan berdasarkan saldo rata-rata (tabungan)
  compounding:           ENUM (none, monthly, quarterly, yearly)
  tax_rate:              NUMERIC(5,2)    # PPh atas bunga (default: 10% sesuai regulasi)
```

**Contoh perhitungan tabungan (DAILY_BALANCE):**
```
Saldo harian rata-rata bulan Januari: Rp 5.000.000
Rate: 3% p.a.
Bunga = 5.000.000 x (3/100) x (31/365) = Rp 12.740
PPh 10% = Rp 1.274
Bunga bersih = Rp 11.466
```

**Contoh perhitungan pinjaman (DECLINING):**
```
Pokok: Rp 10.000.000, Tenor: 12 bulan, Rate: 12% p.a.
Bulan 1: Bunga = 10.000.000 x (12/100) x (1/12) = Rp 100.000
Bulan 2: Bunga =  9.166.667 x (12/100) x (1/12) = Rp  91.667  (pokok berkurang)
... dst
```

#### 3b. Islamic Mode — Nisbah, Margin, Ujrah

```
islamic_config (JSONB):
  akad_type:             ENUM (lihat Section 4)
  ├── [MUDHARABAH]
  │   nisbah_nasabah:    NUMERIC(5,2)   # Porsi bagi hasil nasabah (misal: 40.00 = 40%)
  │   nisbah_koperasi:   NUMERIC(5,2)   # Porsi bagi hasil koperasi (misal: 60.00 = 60%)
  │   projected_return:  NUMERIC(7,4)   # Indikasi return (bukan janji, untuk simulasi)
  │
  ├── [WADIAH]
  │   bonus_policy:      ENUM (none, discretionary)  # Bonus atas kebijakan koperasi
  │
  ├── [MURABAHAH]
  │   margin_rate:       NUMERIC(7,4)   # Margin jual beli (% dari harga pokok)
  │   margin_amount:     NUMERIC(15,2)  # Margin nominal (auto-calculate atau manual)
  │   payment_method:    ENUM (flat, declining)
  │
  ├── [MUSYARAKAH]
  │   nisbah_nasabah:    NUMERIC(5,2)
  │   nisbah_koperasi:   NUMERIC(5,2)
  │   capital_share:     NUMERIC(5,2)   # Porsi modal nasabah (%)
  │
  ├── [IJARAH]
  │   ujrah_rate:        NUMERIC(7,4)   # Biaya sewa per periode
  │   ujrah_amount:      NUMERIC(15,2)  # Nominal sewa per bulan
  │
  └── [QARDH]
      admin_fee_only:    BOOLEAN         # Hanya biaya admin, tanpa margin/bunga
      max_admin_fee_pct: NUMERIC(5,2)    # Batas % biaya admin dari pokok
```

**Aturan nisbah:**
- `nisbah_nasabah + nisbah_koperasi` **harus = 100%** — di-enforce di application layer
- `projected_return` adalah **indikasi**, bukan janji — disclaimer wajib ditampilkan di UI
- Bagi hasil aktual dihitung dari **profit riil** koperasi di akhir periode (bulanan/tahunan)

### 4. Jenis Akad (Islamic Contracts)

Akad adalah kontrak syariah yang mendasari setiap produk di BMT mode. Berikut jenis akad yang didukung:

| Akad | Prinsip | Kategori Rekening | Penjelasan |
|------|---------|-------------------|------------|
| **Wadiah** (Yad Dhamanah) | Titipan dengan jaminan | SIMPANAN_POKOK, SIMPANAN_WAJIB, TABUNGAN | Nasabah menitipkan dana, koperasi boleh menggunakan, bisa beri bonus (bukan wajib) |
| **Mudharabah** | Bagi hasil | TABUNGAN, DEPOSITO | Nasabah = shahibul maal (pemilik dana), koperasi = mudharib (pengelola). Profit dibagi sesuai nisbah |
| **Murabahah** | Jual beli + margin | PINJAMAN | Koperasi beli barang, jual ke nasabah dengan margin yang disepakati. Angsuran tetap |
| **Musyarakah** | Kemitraan/partnership | PINJAMAN | Modal bersama, profit & loss sharing sesuai nisbah. Cocok untuk modal usaha |
| **Ijarah** | Sewa | PINJAMAN | Koperasi menyewakan aset/jasa, nasabah bayar ujrah (sewa) berkala |
| **Qardh** | Pinjaman kebajikan | PINJAMAN | Pinjaman tanpa bunga/margin, hanya biaya admin. Untuk kebutuhan darurat/sosial |

**Aturan:**
- Setiap produk BMT **wajib** memiliki tepat satu `akad_type`
- Produk general **tidak boleh** memiliki `akad_type` — field harus null
- Mapping akad ke kategori di-enforce di application layer (misal: Murabahah tidak bisa di-assign ke produk tabungan)
- Dokumen akad (template PDF) di-generate otomatis berdasarkan `akad_type` saat pembukaan rekening
- Nasabah wajib menyetujui akad (digital consent) — tercatat di `rekening_application.terms_accepted`

### 5. Product Lifecycle

```
┌─────────────────┐
│  Status: DRAFT  │  Produk baru, belum bisa digunakan
└────────┬────────┘
         │ activate
         v
┌─────────────────┐
│ Status: ACTIVE  │  Produk tersedia untuk pembukaan rekening
└────────┬────────┘
         │ discontinue
         v
┌──────────────────────┐
│ Status: DISCONTINUED │  Tidak bisa digunakan untuk rekening baru
└──────────────────────┘
         Rekening existing tetap aktif dengan rate lama
```

**Aturan:**
- `DRAFT → ACTIVE`: oleh Manager+ — validasi kelengkapan field wajib
- `ACTIVE → DISCONTINUED`: oleh Manager+ — hanya jika ada produk pengganti di kategori yang sama (untuk mandatory categories)
- **DISCONTINUED bukan DELETE** — produk tidak bisa dihapus jika ada rekening yang mereferensikannya
- Rekening yang sudah menggunakan produk DISCONTINUED **tetap aktif** dengan rate/config dari versi terakhir saat pembukaan
- Produk DRAFT bisa dihapus (hard delete) karena belum ada referensi

### 6. Product Versioning

Rate dan konfigurasi produk bisa berubah dari waktu ke waktu. Sistem menggunakan **versioning** untuk menjaga konsistensi:

```
produk (current)
├── version: 3
├── interest_rate: 5.0%
└── effective_since: 2026-04-01

produk_version_history
├── version: 1 │ rate: 3.0% │ effective: 2025-01-01 │ superseded: 2025-07-01
├── version: 2 │ rate: 4.0% │ effective: 2025-07-01 │ superseded: 2026-04-01
└── version: 3 │ rate: 5.0% │ effective: 2026-04-01 │ superseded: null (current)
```

**Aturan:**
- Setiap perubahan rate/config menghasilkan **version baru** (immutable history)
- `produk` tabel menyimpan **versi terkini** — query cepat untuk produk aktif
- `produk_version_history` menyimpan **semua versi** — untuk audit dan lookup rekening lama
- Rekening menyimpan `product_version` di momen pembukaan — rate yang berlaku saat buka
- **Existing rekening tetap pakai rate versi lama** kecuali:
  - Tabungan: rate otomatis update ke versi terbaru (floating rate)
  - Deposito: rate locked selama tenor (fixed rate at placement)
  - Pinjaman: tergantung tipe — fixed rate locked, floating rate update
- Field `rate_type` di produk: `FIXED` (locked saat buka) atau `FLOATING` (ikut versi terbaru)

### 7. Product Eligibility Rules

Tidak semua produk tersedia untuk semua jenis anggota. Eligibility diatur per produk berdasarkan `school_relation_type`:

```
produk_eligibility:
  eligible_relations:
    - student        # Siswa/Santri
    - teacher        # Guru/Ustadz
    - staff          # TU/Staf
    - parent         # Orang Tua/Wali
    - external       # Masyarakat Umum
```

**Contoh konfigurasi:**

| Produk | student | teacher | staff | parent | external |
|--------|---------|---------|-------|--------|----------|
| Simpanan Pokok | v | v | v | v | v |
| Simpanan Wajib | v | v | v | v | v |
| Tabungan Reguler | v | v | v | v | v |
| Tabungan Pendidikan | v | - | - | v | - |
| Deposito 12 Bulan | - | v | v | v | v |
| Pinjaman Reguler | - | v | v | - | v |
| Pinjaman Darurat | - | v | v | - | - |
| Pembiayaan Qurban (BMT) | v | v | v | v | v |

**Aturan khusus per relasi:**
- **Student**: limit pinjaman lebih kecil (configurable), wajib ada wali sebagai penjamin
- **Teacher/Staff**: eligible semua produk, limit standar, bisa payroll deduction
- **Parent**: bisa setor ke rekening anak, eligible deposito dan beberapa tabungan
- **External**: eligible produk standar, limit mungkin berbeda (configurable per produk)

**Enforcement:**
- Eligibility di-cek saat **pengajuan rekening** (submit pre-check di K002)
- Jika nasabah tidak eligible, pengajuan di-reject otomatis dengan pesan "Produk tidak tersedia untuk jenis keanggotaan Anda"
- Admin bisa mengubah eligibility produk kapan saja — perubahan hanya berlaku untuk pengajuan baru, rekening existing tidak terpengaruh

### 8. Minimum / Maximum Amounts per Product

Setiap produk mendefinisikan batas nominal yang berlaku:

```
amount_config:
  ── Saldo ──
  min_balance:              NUMERIC(15,2)   # Saldo minimum yang harus dijaga
  max_balance:              NUMERIC(15,2)   # Saldo maksimum (null = unlimited)

  ── Setoran ──
  min_deposit:              NUMERIC(15,2)   # Setoran minimum per transaksi
  max_deposit_per_trx:      NUMERIC(15,2)   # Setoran maksimum per transaksi

  ── Penarikan ──
  min_withdrawal:           NUMERIC(15,2)   # Penarikan minimum per transaksi
  max_withdrawal_daily:     NUMERIC(15,2)   # Limit penarikan harian
  max_withdrawal_monthly:   NUMERIC(15,2)   # Limit penarikan bulanan

  ── Pinjaman ──
  min_loan_amount:          NUMERIC(15,2)   # Plafon minimum
  max_loan_amount:          NUMERIC(15,2)   # Plafon maksimum
  min_tenor_months:         INTEGER         # Tenor minimum
  max_tenor_months:         INTEGER         # Tenor maksimum

  ── Deposito ──
  min_placement:            NUMERIC(15,2)   # Nominal penempatan minimum
  tenor_options:            INTEGER[]       # Pilihan tenor: [1, 3, 6, 12, 24]
```

**Contoh konfigurasi produk Tabungan Siswa:**
```
Tabungan Siswa:
  min_balance:            Rp     5.000
  min_deposit:            Rp     1.000
  max_withdrawal_daily:   Rp   500.000
  max_withdrawal_monthly: Rp 2.000.000
```

**Contoh konfigurasi produk Pinjaman Guru:**
```
Pinjaman Guru:
  min_loan_amount:        Rp   1.000.000
  max_loan_amount:        Rp  50.000.000
  min_tenor_months:       3
  max_tenor_months:       36
```

### 9. Product-Level Fee Configuration

Setiap produk bisa mendefinisikan biaya-biaya yang berlaku:

```
fee_config (JSONB):
  ── Biaya Pembukaan ──
  opening_fee:              NUMERIC(15,2)   # Biaya buka rekening (null = gratis)
  opening_fee_type:         ENUM (fixed, percentage)
  opening_fee_percentage:   NUMERIC(5,2)    # Jika type = percentage

  ── Biaya Administrasi ──
  admin_fee_monthly:        NUMERIC(15,2)   # Biaya admin bulanan (null = gratis)
  admin_fee_deduct_from:    ENUM (balance, separate)  # Potong dari saldo atau tagih terpisah

  ── Biaya Penutupan ──
  closing_fee:              NUMERIC(15,2)   # Biaya tutup rekening (null = gratis)

  ── Penalti ──
  early_withdrawal_penalty_pct: NUMERIC(5,2)  # % penalti pencairan deposito sebelum jatuh tempo
  late_payment_penalty_pct:     NUMERIC(5,2)  # % denda keterlambatan angsuran (general mode)
  late_payment_penalty_fixed:   NUMERIC(15,2) # Denda tetap per hari keterlambatan

  ── Khusus BMT ──
  tazir_rate:               NUMERIC(5,2)    # Ta'zir (denda syariah) — masuk dana sosial, bukan pendapatan koperasi
  tazir_destination:        VARCHAR          # Kode akun tujuan ta'zir (dana sosial/ZIS)
```

**Aturan:**
- Fee configuration bersifat **opsional** — null berarti gratis/tidak berlaku
- Perubahan fee mengikuti **versioning** produk — rekening existing tetap pakai fee versi saat buka (kecuali floating)
- BMT mode: denda (ta'zir) **tidak masuk pendapatan** koperasi — wajib dialokasikan ke dana sosial/ZIS sesuai fatwa DSN-MUI
- Semua potongan fee dicatat sebagai **transaksi terpisah** dengan referensi ke produk dan jenis fee

### 10. Tenant-Configurable Default Products

Setiap tenant wajib memiliki **produk default** untuk kategori mandatory (SIMPANAN_POKOK dan SIMPANAN_WAJIB) yang digunakan saat auto-create rekening pada approval nasabah baru (lihat [ADR-K002](./ADR-K002-rekening.md) Section 2):

```
tenant_default_products:
  SIMPANAN_POKOK:  "produk-uuid-pokok"     # Wajib ada, 1 produk
  SIMPANAN_WAJIB:  "produk-uuid-wajib"     # Wajib ada, 1 produk
```

**Aturan:**
- Saat tenant onboarding, sistem membuat **2 produk default** (Simpanan Pokok dan Simpanan Wajib) dengan konfigurasi standar
- Admin tenant bisa mengubah konfigurasi produk default (rate, jumlah, dll) tapi **tidak bisa menghapusnya**
- Jika produk default di-DISCONTINUED, harus ada **pengganti** yang di-assign sebagai default baru sebelum discontinue diizinkan
- Produk default untuk BMT mode otomatis menyertakan akad Wadiah Yad Dhamanah

**Onboarding auto-seed:**
```
Tenant Created
      │
      v
┌───────────────────────────────────┐
│ Auto-create 2 produk default:     │
│ 1. Simpanan Pokok [category=SP]   │
│    - amount: Rp 100.000 (default) │
│    - status: ACTIVE               │
│ 2. Simpanan Wajib [category=SW]   │
│    - amount: Rp 50.000 (default)  │
│    - status: ACTIVE               │
│ [BMT] akad: Wadiah Yad Dhamanah  │
└───────────────────────────────────┘
      │
      v
Admin bisa adjust konfigurasi setelah onboarding
```

### 11. Data Model

```
produk
├── id                      UUID v7 (PK)
├── tenant_id               UUID (FK → tenant)
├── code                    VARCHAR NOT NULL UNIQUE per tenant
├── name                    VARCHAR NOT NULL
├── description             TEXT
│
├── ── Klasifikasi ──
├── category                ENUM (simpanan_pokok, simpanan_wajib, tabungan, deposito, pinjaman)
├── coop_type               ENUM (general, islamic) NOT NULL
│
├── ── Rate / Yield (General) ──
├── interest_config         JSONB (nullable, hanya untuk general mode)
│   # { rate, calculation_method, compounding, tax_rate }
│
├── ── Akad & Rate (Islamic) ──
├── akad_type               ENUM (wadiah, mudharabah, murabahah, musyarakah, ijarah, qardh) nullable
├── islamic_config          JSONB (nullable, hanya untuk islamic mode)
│   # { nisbah_nasabah, nisbah_koperasi, margin_rate, ujrah_rate, ... }
│
├── ── Rate Type ──
├── rate_type               ENUM (fixed, floating) DEFAULT 'floating'
│
├── ── Amount Limits ──
├── min_balance             NUMERIC(15,2) DEFAULT 0
├── max_balance             NUMERIC(15,2) (nullable, null = unlimited)
├── min_deposit             NUMERIC(15,2) DEFAULT 0
├── max_deposit_per_trx     NUMERIC(15,2) (nullable)
├── min_withdrawal          NUMERIC(15,2) DEFAULT 0
├── max_withdrawal_daily    NUMERIC(15,2) (nullable)
├── max_withdrawal_monthly  NUMERIC(15,2) (nullable)
│
├── ── Pinjaman Limits ──
├── min_loan_amount         NUMERIC(15,2) (nullable)
├── max_loan_amount         NUMERIC(15,2) (nullable)
├── min_tenor_months        INTEGER (nullable)
├── max_tenor_months        INTEGER (nullable)
│
├── ── Deposito Config ──
├── min_placement           NUMERIC(15,2) (nullable)
├── tenor_options           JSONB (nullable, array of integers: [1,3,6,12,24])
│
├── ── Fee Config ──
├── fee_config              JSONB (nullable)
│   # { opening_fee, admin_fee_monthly, closing_fee, penalties... }
│
├── ── Eligibility ──
├── eligible_relations      JSONB NOT NULL DEFAULT '["student","teacher","staff","parent","external"]'
│   # Array of school_relation_type
│
├── ── Lifecycle ──
├── status                  ENUM (draft, active, discontinued) DEFAULT 'draft'
├── version                 INTEGER NOT NULL DEFAULT 1
├── effective_since         DATE NOT NULL
│
├── ── Tenant Default ──
├── is_default              BOOLEAN DEFAULT false
│   # true = produk default untuk kategori ini di tenant ini
│
├── ── Vernon Fields ──
├── _rels                   JSONB NOT NULL DEFAULT '{}'
├── _data                   JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at          TIMESTAMPTZ
    ├── created_by          UUID
    ├── updated_at          TIMESTAMPTZ
    └── updated_by          UUID
```

```
produk_version_history
├── id                      UUID v7 (PK)
├── tenant_id               UUID (FK → tenant)
├── produk_id               UUID (FK → produk) NOT NULL
├── version                 INTEGER NOT NULL
│
├── ── Snapshot Config ──
├── interest_config         JSONB (nullable)
├── islamic_config          JSONB (nullable)
├── fee_config              JSONB (nullable)
├── min_balance             NUMERIC(15,2)
├── max_balance             NUMERIC(15,2) (nullable)
│   # ... semua field config lainnya yang bisa berubah
│
├── ── Periode Berlaku ──
├── effective_since         DATE NOT NULL
├── superseded_at           DATE (nullable, null = current version)
│
└── ── Audit ──
    ├── created_at          TIMESTAMPTZ
    └── created_by          UUID
```

### 12. Vernon _rels dan _data Structure

Produk menggunakan Vernon pattern karena listing produk membutuhkan data tenant (minimal), tapi lebih penting: **rekening, transaksi, dan laporan** perlu snapshot data produk tanpa JOIN.

**_rels:**
```json
{
  "tenant_id": "018f..."
}
```

**_data:**
```json
{
  "tenant": {
    "id":   "018f...",
    "name": "Koperasi Al-Ikhlas",
    "coop_type": "islamic"
  }
}
```

**Catatan:**
- Produk sendiri jarang di-list lintas entity — Vernon di produk lebih bermanfaat sebagai **sumber _data untuk entity lain** (rekening, transaksi)
- Saat rekening dibuat, `_data.product` di rekening di-populate dari produk (lihat K002 Section 12)
- Saat produk di-update (nama, rate), SyncEngine propagasi ke `_data.product` di semua rekening terkait

**SyncEngine triggers:**
- `ProductUpdatedEvent` → update `_data.product` di semua `rekening` yang menggunakan produk tersebut
- `ProductUpdatedEvent` → update `_data.product` di semua `rekening_application` yang pending dengan produk tersebut

### 13. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View katalog produk | v | v | v | v |
| View detail produk (termasuk rate) | v | v | v | v |
| Buat produk baru (DRAFT) | - | - | v | v |
| Edit produk DRAFT | - | - | v | v |
| Activate produk (DRAFT → ACTIVE) | - | - | v | v |
| Edit produk ACTIVE (trigger new version) | - | - | - | v |
| Discontinue produk | - | - | - | v |
| Delete produk DRAFT | - | - | v | v |
| Set produk sebagai default tenant | - | - | - | v |
| Edit eligibility rules | - | - | v | v |
| View version history | - | v | v | v |

**Catatan:**
- Manager bisa buat dan activate produk, tapi **edit produk ACTIVE** (yang memicu version baru) hanya Admin — karena berdampak pada semua rekening
- Discontinue hanya Admin — operasi berisiko tinggi yang mempengaruhi operasional koperasi
- Teller hanya view — mereka perlu tahu produk apa saja yang tersedia untuk ditawarkan ke nasabah

### 14. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Product entity | Produk | Produk |
| Rate config label | Suku Bunga | Nisbah Bagi Hasil / Margin |
| Rate display | Bunga: 6% p.a. | Nisbah: 40:60 (Nasabah:Koperasi) |
| Loan product category | Pinjaman | Pembiayaan |
| Contract | Perjanjian | Akad |
| Interest income | Pendapatan Bunga | Pendapatan Margin / Bagi Hasil |
| Penalty | Denda | Ta'zir |
| Penalty destination | Pendapatan koperasi | Dana sosial (ZIS) |
| Deposit product | Deposito Berjangka | Deposito Mudharabah |
| Savings product | Tabungan | Tabungan Mudharabah / Tabungan Wadiah |
| Certificate | Bilyet Deposito | Bilyet Deposito + Surat Akad |

**Aturan dual-mode:**
- Satu tenant hanya bisa memiliki **satu coop_type** — produk harus sesuai dengan tipe tenant
- Produk dengan `coop_type = general` **tidak bisa** digunakan di tenant BMT dan sebaliknya
- Validasi `coop_type` produk vs tenant di-enforce saat create/activate produk
- UI menampilkan label sesuai `coop_type` tenant yang sedang login

## Consequences

### Positif

- **Single source of truth** — semua aturan bisnis rekening terdefinisi di satu tempat (produk), bukan tersebar di kode
- **Flexible** — tenant bisa membuat produk custom sesuai kebutuhan koperasi masing-masing
- **Versioned** — perubahan rate tidak merusak rekening existing, audit trail lengkap
- **Dual-mode compliant** — mendukung koperasi konvensional dan BMT dalam satu codebase
- **Eligibility** — kontrol akses produk per jenis anggota mencegah kesalahan operasional
- **Fee transparency** — semua biaya terdefinisi di level produk, tidak ada hidden charges
- **Regulasi syariah** — akad, nisbah, dan ta'zir destination memenuhi ketentuan DSN-MUI

### Negatif

- **Complexity** — banyak field dan konfigurasi per produk, error-prone saat setup
- **Version management** — harus maintain history dan tentukan kapan rate lama vs baru berlaku
- **JSONB fields** — `interest_config`, `islamic_config`, `fee_config` tidak bisa di-enforce schema di DB level
- **Dual-mode duplication** — beberapa logic harus di-branch berdasarkan `coop_type`

### Mitigasi

- Sistem menyediakan **produk template** saat tenant onboarding — Admin tinggal adjust, tidak mulai dari nol
- Version increment otomatis saat save — tidak perlu manage version manual
- JSONB schema di-validate di **application layer** dengan strict validation (JSON Schema atau Go struct)
- Dual-mode branching di-isolasi di **service layer** — handler/controller agnostik terhadap mode

## Alternatives Considered

### A. Hard-coded Product Types (tanpa katalog dinamis)

Produk didefinisikan di kode (enum tetap), tenant hanya bisa mengubah rate.

**Ditolak** karena: setiap koperasi memiliki produk yang berbeda-beda. Koperasi A mungkin punya "Tabungan Qurban" sementara koperasi B punya "Tabungan Wisata Akhir Tahun". Katalog dinamis memungkinkan fleksibilitas tanpa deploy kode baru.

### B. Rate di Level Rekening (bukan di Produk)

Setiap rekening menyimpan rate sendiri, tanpa referensi ke produk.

**Ditolak** karena: menyulitkan bulk update rate, tidak ada konsistensi antara rekening sejenis, dan sulit untuk reporting per produk. Produk sebagai single source of truth lebih maintainable.

### C. Separate Table per Akad Type

Tabel berbeda untuk konfigurasi Mudharabah, Murabahah, dll.

**Ditolak** karena: menambah complexity JOIN dan migrasi setiap kali ada akad baru. JSONB `islamic_config` dengan schema validation di application layer lebih fleksibel dan extensible.

### D. Produk Tanpa Versioning (overwrite langsung)

Perubahan rate langsung overwrite data produk, semua rekening otomatis ikut rate baru.

**Ditolak** karena: deposito dengan fixed rate harus dijaga sesuai kontrak awal. Tanpa versioning, tidak ada cara untuk membedakan rate saat buka vs rate terkini. Juga melanggar prinsip audit trail untuk perubahan konfigurasi finansial.
