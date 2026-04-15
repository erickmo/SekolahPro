# ADR-K019: Koperasi Toko & Kantin

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Koperasi sekolah umumnya mengoperasikan **toko** (alat tulis, seragam, buku) dan **kantin** (makanan, minuman) sebagai salah satu sumber pendapatan utama. Keuntungan dari toko dan kantin berkontribusi langsung ke **SHU** ([ADR-K016](./ADR-K016-shu.md)).

Sistem harus mengakomodasi:

- **Penjualan barang fisik** — ATK, seragam, buku, perlengkapan sekolah
- **Penjualan makanan & minuman** — kantin/canteen harian
- **Pembayaran cashless** — siswa membayar dari tabungan koperasi tanpa uang tunai
- **Manajemen stok** — tracking inventory, reorder point, stock opname
- **Meal plan asrama** — untuk sekolah boarding/pesantren
- **Spending limit** — orang tua mengatur batas belanja harian anak
- **Multi-lokasi** — beberapa toko/kantin per cabang atau lintas cabang
- **Dual-mode** — toko dan kantin bersifat non-interest-bearing, sama di kedua mode (general/islamic)

## Decision

### 1. Scope — Toko dan Kantin sebagai Unit Operasional Koperasi

Toko dan kantin adalah **unit bisnis** yang dikelola koperasi. Produk yang dijual di sini **berbeda** dari produk keuangan di [ADR-K003](./ADR-K003-produk-akad.md) — ini adalah barang fisik dan makanan, bukan tabungan/pinjaman/deposito.

```
Koperasi Sekolah
├── Layanan Keuangan (K001-K018)
│   ├── Tabungan, Deposito, Pinjaman
│   └── SHU, Laporan Regulasi
│
└── Unit Usaha — Toko & Kantin (K019)
    ├── Toko (ATK, seragam, buku, perlengkapan)
    └── Kantin (makanan, minuman, snack)
```

**Aturan:**
- Toko dan kantin beroperasi di bawah koperasi — keuntungan masuk ke pendapatan koperasi
- Setiap tenant bisa mengaktifkan/menonaktifkan modul toko dan kantin secara independen via **feature flag**
- Produk toko/kantin memiliki katalog terpisah dari produk keuangan

### 2. Katalog Produk — Inventory Items

Setiap barang yang dijual memiliki entry di katalog produk toko:

```
toko_produk
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
│
├── ── Identitas ──
├── name                  VARCHAR NOT NULL
├── sku                   VARCHAR UNIQUE per tenant+branch
├── barcode               VARCHAR (nullable, untuk scanner)
├── description           TEXT (nullable)
├── category_id           UUID (FK → toko_kategori)
│
├── ── Harga ──
├── cost_price            NUMERIC(15,2) NOT NULL (harga beli/HPP)
├── sell_price            NUMERIC(15,2) NOT NULL (harga jual)
├── margin_percentage     NUMERIC(5,2) (calculated: (sell - cost) / cost * 100)
│
├── ── Stok ──
├── current_stock         INTEGER NOT NULL DEFAULT 0
├── minimum_stock         INTEGER DEFAULT 0 (reorder point alert)
├── unit                  VARCHAR NOT NULL (pcs, pack, box, kg, liter)
│
├── ── Klasifikasi ──
├── product_type          ENUM (store_item, canteen_item)
├── is_active             BOOLEAN DEFAULT true
├── is_daily_menu         BOOLEAN DEFAULT false (khusus kantin — item menu harian)
│
├── ── Visual ──
├── image_url             VARCHAR (nullable)
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

**Kategori produk:**

```
toko_kategori
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── name                  VARCHAR NOT NULL
├── parent_id             UUID (nullable, FK → toko_kategori, untuk sub-kategori)
├── product_type          ENUM (store_item, canteen_item)
├── sort_order            INTEGER DEFAULT 0
├── is_active             BOOLEAN DEFAULT true
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Contoh kategori:**
```
store_item:
├── ATK (Alat Tulis Kantor)
│   ├── Buku Tulis
│   ├── Pensil & Pulpen
│   └── Penggaris & Penghapus
├── Seragam
│   ├── Seragam Harian
│   ├── Seragam Olahraga
│   └── Seragam Pramuka
└── Perlengkapan

canteen_item:
├── Makanan Berat
├── Snack
├── Minuman
└── Paket Meal Plan
```

### 3. Point of Sale (POS)

Proses penjualan di toko/kantin menggunakan sistem POS:

```
Kasir scan/pilih item
        │
        v
┌─────────────────┐
│   CART (Draft)   │  Kasir menambahkan item ke keranjang
│  item 1: Rp 5k  │
│  item 2: Rp 3k  │
│  Total: Rp 8k   │
└────────┬────────┘
         │ checkout
         v
┌─────────────────┐
│  PAYMENT         │  Pilih metode bayar
│  Cash / Tabungan │
│  / E-Wallet      │
│  / Mixed         │
└────────┬────────┘
         │ confirm
         v
┌─────────────────┐
│  COMPLETED       │  Receipt generated
│  No. POS-001     │
└─────────────────┘
```

**Aturan POS:**
- Kasir harus dalam **sesi aktif** — mirip Teller Session ([ADR-K012](./ADR-K012-teller-session.md)) tapi untuk POS
- Setiap transaksi penjualan menghasilkan **receipt** dengan nomor unik
- Item di keranjang bisa di-edit sebelum checkout
- Stok dikurangi secara **atomik** bersamaan dengan insert penjualan (satu database transaction)
- Jika stok < quantity yang diminta, transaksi ditolak
- Pembatalan penjualan (void) hanya oleh **Supervisor+** dalam waktu yang dikonfigurasi (default: 24 jam)

### 4. Metode Pembayaran

POS mendukung beberapa metode pembayaran:

```
payment_method:
├── CASH                  Uang tunai (kasir terima & kembalikan)
├── TABUNGAN_DEBIT        Potong saldo tabungan nasabah (cashless)
├── EWALLET_DEBIT         Potong saldo e-wallet (K021, jika diaktifkan)
└── MIXED                 Kombinasi di atas (misal: sebagian tunai, sebagian tabungan)
```

**TABUNGAN_DEBIT flow:**
```
Siswa tap kartu / scan QR
        │
        v
┌─────────────────────────┐
│ Identifikasi nasabah    │
│ Cek spending limit      │
│ Cek available_balance   │
└────────┬────────────────┘
         │ semua OK
         v
┌─────────────────────────┐
│ Create transaksi K011   │  Penarikan dari tabungan
│ + Create penjualan POS  │  Dalam 1 DB transaction
└─────────────────────────┘
```

**Aturan metode bayar:**
- `TABUNGAN_DEBIT` memerlukan identifikasi nasabah (kartu/QR code)
- Cek `available_balance >= total_amount` sebelum proses
- Cek **spending limit** harian sebelum proses (lihat Section 7)
- Transaksi debit tabungan tercatat sebagai transaksi resmi di [ADR-K011](./ADR-K011-transaksi.md) dengan reference ke penjualan POS
- `MIXED` payment: sistem mencatat masing-masing metode dan nominalnya

### 5. Manajemen Stok

Tracking pergerakan stok barang:

```
toko_stok_movement
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
├── produk_id             UUID (FK → toko_produk)
│
├── ── Movement ──
├── movement_type         ENUM (stock_in, stock_out, adjustment, opname, transfer)
├── quantity              INTEGER NOT NULL (positif untuk masuk, negatif untuk keluar)
├── stock_before          INTEGER NOT NULL
├── stock_after           INTEGER NOT NULL
│
├── ── Referensi ──
├── reference_type        ENUM (purchase, sale, adjustment, opname, transfer)
├── reference_id          UUID (FK → tabel referensi sesuai type)
├── notes                 TEXT (nullable)
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

**Jenis movement:**

| Type | Trigger | Quantity |
|------|---------|----------|
| `stock_in` | Penerimaan barang dari supplier | + |
| `stock_out` | Penjualan via POS | - |
| `adjustment` | Koreksi manual (rusak, expired, hilang) | +/- |
| `opname` | Stock opname (penyesuaian ke fisik) | +/- |
| `transfer` | Transfer antar lokasi toko | +/- |

**Aturan stok:**
- `current_stock` di tabel produk adalah **cache** — source of truth adalah SUM dari stok_movement
- Reconciliation harian: cached stock vs SUM movements, alert jika selisih
- Penjualan yang membuat stok < 0 **ditolak** — tidak boleh stok negatif
- **Reorder point alert**: jika `current_stock <= minimum_stock`, kirim notifikasi ke Supervisor

### 6. Stock Opname (Physical Count)

Proses verifikasi stok fisik terhadap stok di sistem:

```
Supervisor buat session opname
        │
        v
┌─────────────────────────┐
│  Status: OPEN            │  Kasir/staff hitung fisik
│  produk A: sistem=50     │
│            fisik=48      │  selisih: -2
│  produk B: sistem=30     │
│            fisik=30      │  selisih: 0
└────────┬────────────────┘
         │ submit
         v
┌─────────────────────────┐
│  Status: PENDING         │  Menunggu approval Manager
└────────┬────────────────┘
         │ approve
         v
┌─────────────────────────┐
│  Status: APPROVED        │  Stok di-adjust sesuai fisik
│  Auto-create adjustment  │
│  movements untuk selisih │
└─────────────────────────┘
```

**Data model stock opname:**

```
toko_opname
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
├── opname_number         VARCHAR UNIQUE per tenant
├── opname_date           DATE NOT NULL
│
├── ── Workflow ──
├── status                ENUM (open, pending, approved, rejected)
├── submitted_at          TIMESTAMPTZ (nullable)
├── reviewed_at           TIMESTAMPTZ (nullable)
├── reviewed_by           UUID (nullable, FK → user)
├── notes                 TEXT (nullable)
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

toko_opname_item
├── id                    UUID v7 (PK)
├── opname_id             UUID (FK → toko_opname)
├── produk_id             UUID (FK → toko_produk)
├── system_stock          INTEGER NOT NULL
├── physical_stock        INTEGER NOT NULL
├── difference            INTEGER NOT NULL (physical - system)
├── notes                 TEXT (nullable, alasan selisih)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    └── created_by        UUID
```

### 7. Student Daily Spending Limit

Batas belanja harian untuk siswa — dikonfigurasi oleh orang tua atau admin:

```
spending_limit_config
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Limit ──
├── daily_limit           NUMERIC(15,2) NOT NULL (maksimal belanja per hari)
├── per_transaction_limit NUMERIC(15,2) (nullable, maksimal per transaksi)
├── allowed_categories    JSONB (nullable, restrict ke kategori tertentu)
│
├── ── Pengaturan ──
├── is_active             BOOLEAN DEFAULT true
├── set_by                ENUM (parent, admin)
├── parent_nasabah_id     UUID (nullable, FK → nasabah, orang tua yang set)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Aturan spending limit:**
- Default spending limit bisa di-set per **school_relation_type** di tenant config:
  ```
  default_spending_limit:
    student: 30000         # Rp 30.000/hari
    teacher: null          # Tidak ada limit
    staff: null            # Tidak ada limit
    parent: null           # Tidak ada limit
  ```
- Orang tua bisa meng-override default limit untuk anaknya — via teller atau self-service portal ([ADR-K023](./ADR-K023-dashboard-portal.md))
- Limit di-enforce saat checkout POS:
  ```
  Pre-checkout check (jika bayar dari tabungan):
  ├── SUM belanja hari ini + transaksi ini <= daily_limit    ✓
  ├── Transaksi ini <= per_transaction_limit                 ✓
  ├── Kategori item dalam allowed_categories (jika di-set)   ✓
  └── available_balance >= total_amount                      ✓
  ```
- Jika limit terlampaui, transaksi **ditolak** dengan pesan jelas ("Batas belanja harian tercapai")
- Attempt yang ditolak dicatat di log untuk visibility orang tua

### 8. Supplier Management

Tracking supplier dan pembelian barang:

```
toko_supplier
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── name                  VARCHAR NOT NULL
├── contact_person        VARCHAR (nullable)
├── phone                 VARCHAR (nullable)
├── email                 VARCHAR (nullable)
├── address               TEXT (nullable)
├── is_active             BOOLEAN DEFAULT true
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

**Purchase Order:**

```
toko_pembelian
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
├── supplier_id           UUID (FK → toko_supplier)
├── po_number             VARCHAR UNIQUE per tenant
├── po_date               DATE NOT NULL
│
├── ── Total ──
├── total_amount          NUMERIC(15,2) NOT NULL
├── total_items           INTEGER NOT NULL
│
├── ── Workflow ──
├── status                ENUM (draft, pending, approved, received, cancelled)
├── approved_at           TIMESTAMPTZ (nullable)
├── approved_by           UUID (nullable)
├── received_at           TIMESTAMPTZ (nullable)
├── received_by           UUID (nullable)
├── notes                 TEXT (nullable)
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

toko_pembelian_item
├── id                    UUID v7 (PK)
├── pembelian_id          UUID (FK → toko_pembelian)
├── produk_id             UUID (FK → toko_produk)
├── quantity              INTEGER NOT NULL
├── unit_price            NUMERIC(15,2) NOT NULL
├── subtotal              NUMERIC(15,2) NOT NULL
├── received_quantity     INTEGER DEFAULT 0
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    └── created_by        UUID
```

**Flow pembelian:**
```
Supervisor buat PO
        │
        v
┌─────────────────┐
│  Status: DRAFT  │
└────────┬────────┘
         │ submit
         v
┌─────────────────┐
│ Status: PENDING │  Menunggu approval Manager
└────────┬────────┘
         │ approve
         v
┌─────────────────┐
│ Status: APPROVED│  PO dikirim ke supplier
└────────┬────────┘
         │ barang datang
         v
┌─────────────────┐
│ Status: RECEIVED│  Stok otomatis bertambah (stock_in movement)
└─────────────────┘
```

### 9. Kantin-Specific — Menu Harian dan Meal Plan

Fitur khusus kantin:

**Menu harian:**
- Admin kantin mengatur menu per hari (`is_daily_menu = true` pada produk yang tersedia hari itu)
- Menu bisa dirotasi mingguan — configurable per tenant
- Siswa/wali bisa melihat menu hari ini di portal ([ADR-K023](./ADR-K023-dashboard-portal.md))

**Meal plan untuk asrama:**

```
toko_meal_plan
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah, siswa boarding)
│
├── ── Plan ──
├── plan_type             ENUM (full_board, half_board, custom)
├── meals_included        JSONB (["breakfast", "lunch", "dinner"])
├── monthly_fee           NUMERIC(15,2) NOT NULL
├── start_date            DATE NOT NULL
├── end_date              DATE (nullable, ongoing jika null)
│
├── ── Deduction ──
├── auto_deduct           BOOLEAN DEFAULT true
├── deduct_from_rekening  UUID (nullable, FK → rekening, tabungan untuk auto-deduct)
├── deduct_day            INTEGER DEFAULT 1 (tanggal potong bulanan, 1-28)
│
├── ── Status ──
├── is_active             BOOLEAN DEFAULT true
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

**Aturan meal plan:**
- `full_board` = 3x makan (breakfast, lunch, dinner)
- `half_board` = 2x makan (pilih kombinasi)
- `custom` = konfigurasi bebas
- Auto-deduct bulanan dari tabungan siswa — tercatat sebagai transaksi di [ADR-K011](./ADR-K011-transaksi.md)
- Jika saldo tabungan tidak cukup, notifikasi ke orang tua ([ADR-K022](./ADR-K022-notifikasi.md))

### 10. Penjualan Data Model

```
toko_penjualan
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
├── location_id           UUID (FK → toko_lokasi)
│
├── ── Identitas ──
├── receipt_number        VARCHAR UNIQUE per tenant
├── sale_date             TIMESTAMPTZ NOT NULL
├── sale_type             ENUM (store, canteen)
│
├── ── Pembeli ──
├── nasabah_id            UUID (nullable, FK → nasabah, jika bayar dari tabungan)
├── customer_name         VARCHAR (nullable, jika bukan nasabah / cash)
│
├── ── Total ──
├── subtotal              NUMERIC(15,2) NOT NULL
├── discount_amount       NUMERIC(15,2) DEFAULT 0
├── total_amount          NUMERIC(15,2) NOT NULL
│
├── ── Payment ──
├── payment_method        ENUM (cash, tabungan_debit, ewallet_debit, mixed)
├── cash_amount           NUMERIC(15,2) DEFAULT 0
├── tabungan_debit_amount NUMERIC(15,2) DEFAULT 0
├── ewallet_debit_amount  NUMERIC(15,2) DEFAULT 0
├── change_amount         NUMERIC(15,2) DEFAULT 0 (kembalian untuk cash)
│
├── ── Referensi Transaksi ──
├── transaction_id        UUID (nullable, FK → transaksi K011, jika debit tabungan)
│
├── ── Status ──
├── status                ENUM (completed, voided)
├── voided_at             TIMESTAMPTZ (nullable)
├── voided_by             UUID (nullable)
├── void_reason           TEXT (nullable)
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

toko_penjualan_item
├── id                    UUID v7 (PK)
├── penjualan_id          UUID (FK → toko_penjualan)
├── produk_id             UUID (FK → toko_produk)
├── quantity              INTEGER NOT NULL
├── unit_price            NUMERIC(15,2) NOT NULL
├── discount              NUMERIC(15,2) DEFAULT 0
├── subtotal              NUMERIC(15,2) NOT NULL
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    └── created_by        UUID
```

### 11. Multi-Lokasi

Satu branch bisa memiliki beberapa lokasi toko/kantin:

```
toko_lokasi
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
├── name                  VARCHAR NOT NULL (misal: "Kantin Putra", "Toko Utama")
├── location_type         ENUM (store, canteen)
├── is_active             BOOLEAN DEFAULT true
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

**Aturan multi-lokasi:**
- Setiap penjualan tercatat dengan `location_id` — untuk reporting per lokasi
- Stok bisa di-transfer antar lokasi (movement type `transfer`)
- Katalog produk bisa berbeda per lokasi (produk punya `branch_id`)

### 12. Margin/Profit dan Kontribusi ke SHU

Keuntungan toko/kantin berkontribusi ke pendapatan koperasi:

```
Penjualan toko/kantin
        │
        v
┌─────────────────────────┐
│ Margin per item:         │
│ sell_price - cost_price  │
│ = gross profit           │
└────────┬────────────────┘
         │
         v
┌─────────────────────────┐
│ Total gross profit       │
│ - biaya operasional      │
│ = net profit             │
└────────┬────────────────┘
         │
         v
┌─────────────────────────┐
│ Net profit masuk ke      │
│ pendapatan koperasi      │
│ → SHU (K016)             │
└─────────────────────────┘
```

**Aturan:**
- Setiap penjualan item mencatat `cost_price` dan `sell_price` — margin bisa dihitung per item
- Laporan profit toko/kantin di-generate harian, mingguan, dan bulanan
- Kontribusi ke SHU dihitung berdasarkan net profit unit usaha toko/kantin
- Pencatatan akuntansi menggunakan jurnal sesuai [ADR-K015](./ADR-K015-jurnal-coa.md)

### 13. Reporting

Laporan yang tersedia untuk toko/kantin:

| Laporan | Frekuensi | Audience |
|---------|-----------|----------|
| Penjualan harian | Harian | Kasir, Supervisor |
| Stok movement | Harian | Supervisor |
| Profit margin per item | Mingguan | Manager |
| Best sellers | Mingguan/Bulanan | Manager |
| Slow moving items | Bulanan | Manager |
| Stock opname variance | Per opname | Manager |
| Supplier purchase summary | Bulanan | Manager |
| Spending per siswa | Harian/Bulanan | Parent (via K023) |
| Meal plan status | Bulanan | Admin, Parent |

### 14. Vernon _rels dan _data Structure

**toko_produk _rels/_data:**

```json
// _rels
{
  "tenant_id": "018f...",
  "branch_id": "018f...",
  "category_id": "018f..."
}

// _data
{
  "branch": {
    "id": "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  },
  "category": {
    "id": "018f...",
    "name": "ATK",
    "product_type": "store_item"
  }
}
```

**toko_penjualan _rels/_data:**

```json
// _rels
{
  "tenant_id": "018f...",
  "branch_id": "018f...",
  "location_id": "018f...",
  "nasabah_id": "018f..."
}

// _data
{
  "branch": {
    "id": "018f...",
    "name": "Cabang Jakarta Pusat"
  },
  "location": {
    "id": "018f...",
    "name": "Kantin Putra"
  },
  "nasabah": {
    "id": "018f...",
    "full_name": "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  }
}
```

**SyncEngine triggers:**
- `BranchUpdatedEvent` → update `_data.branch` di semua produk dan penjualan cabang tersebut
- `NasabahUpdatedEvent` → update `_data.nasabah` di semua penjualan nasabah tersebut
- `TokoKategoriUpdatedEvent` → update `_data.category` di semua produk kategori tersebut
- `TokoLokasiUpdatedEvent` → update `_data.location` di semua penjualan lokasi tersebut

### 15. Authorization — RBAC

| Permission | Kasir | Supervisor | Manager | Admin |
|------------|-------|------------|---------|-------|
| Proses penjualan (POS) | v | v | v | v |
| Void penjualan (dalam batas waktu) | - | v | v | v |
| Manage katalog produk | - | v | v | v |
| Set harga jual | - | - | v | v |
| Manage kategori | - | - | v | v |
| Manage supplier | - | v | v | v |
| Buat PO | - | v | v | v |
| Approve PO | - | - | v | v |
| Receive PO (terima barang) | - | v | v | v |
| Stock adjustment manual | - | v | v | v |
| Buat session opname | - | v | v | v |
| Approve opname | - | - | v | v |
| Set spending limit | - | - | - | v |
| View laporan penjualan | v | v | v | v |
| View laporan stok | - | v | v | v |
| View laporan profit | - | - | v | v |
| Manage lokasi toko/kantin | - | - | v | v |
| Manage meal plan | - | v | v | v |

**Catatan:**
- **Kasir** adalah role khusus untuk POS — bisa juga di-assign ke Teller yang sudah ada
- Kasir hanya bisa proses penjualan dan lihat laporan penjualan — tidak bisa manage stok atau harga
- Void penjualan membutuhkan Supervisor+ dan alasan wajib
- Set harga dan approve PO membutuhkan Manager+ — kontrol finansial

### 16. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Module name | Toko Koperasi | Toko Koperasi |
| Receipt header | Koperasi [nama] | BMT [nama] |
| Profit label | Laba/Margin | Laba/Margin |

Toko dan kantin bersifat **non-interest-bearing** — tidak ada perbedaan substansial antara mode general dan islamic. Perbedaan hanya pada terminologi header/branding.

## Consequences

### Positif

- **Revenue stream** — toko dan kantin menjadi sumber pendapatan koperasi yang terlacak
- **Cashless** — siswa bisa belanja tanpa uang tunai, meningkatkan keamanan
- **Parental control** — orang tua bisa mengatur dan memantau belanja anak
- **Inventory control** — stok terlacak, reorder point mencegah kehabisan barang
- **Audit trail** — setiap penjualan, pembelian, dan movement stok tercatat
- **Multi-lokasi** — mendukung sekolah dengan beberapa kantin/toko
- **Meal plan** — otomasi makan untuk siswa boarding/asrama
- **SHU contribution** — profit toko/kantin otomatis berkontribusi ke pendapatan koperasi

### Negatif

- **Complexity** — POS system menambah signifikan ke scope sistem
- **Stok management overhead** — stock opname, reorder, supplier management butuh disiplin
- **Performance** — POS harus cepat (< 2 detik per transaksi) — membutuhkan optimasi khusus
- **Hardware dependency** — scanner barcode, printer receipt, NFC reader

### Mitigasi

- POS module bersifat **opsional** — tenant bisa mengaktifkan/menonaktifkan via feature flag
- Stok management bisa disederhanakan untuk tenant kecil (tanpa supplier/PO, hanya adjustment manual)
- Caching agresif untuk katalog produk — jarang berubah, sering di-query
- Hardware-agnostic — bisa manual input tanpa scanner/NFC (fallback)

## Alternatives Considered

### A. Sistem POS Terpisah (Third-party)

Menggunakan software POS terpisah dan integrasi via API.

**Ditolak** karena: kehilangan integrasi native dengan tabungan nasabah (cashless payment), spending limit, dan kontribusi SHU. Data silo mempersulit reporting terpadu.

### B. Tanpa Inventory Management

Hanya mencatat penjualan tanpa tracking stok.

**Ditolak** karena: koperasi membutuhkan kontrol atas aset (stok adalah aset). Tanpa inventory management, tidak bisa menghitung profit margin yang akurat dan risiko shrinkage tidak terdeteksi.

### C. Single Location per Branch

Membatasi satu toko/kantin per branch.

**Ditolak** karena: banyak sekolah memiliki beberapa kantin (putra/putri) atau toko terpisah dari kantin. Pembatasan ini terlalu kaku untuk realitas di lapangan.
