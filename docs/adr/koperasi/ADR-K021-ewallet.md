# ADR-K021: E-Wallet / Uang Saku Digital

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Siswa/santri di lingkungan sekolah membutuhkan cara pembayaran yang aman, terkontrol, dan cashless — terutama di kantin dan toko koperasi ([ADR-K019](./ADR-K019-toko-kantin.md)). Orang tua ingin:

- Mengontrol berapa yang bisa dibelanjakan anak per hari
- Memantau pengeluaran anak secara real-time
- Top-up saldo anak tanpa harus datang ke sekolah
- Membatasi kategori belanja anak

Sistem ini merupakan **spending interface** di atas rekening tabungan yang sudah ada ([ADR-K002](./ADR-K002-rekening.md), [ADR-K005](./ADR-K005-tabungan.md)), bukan balance terpisah. Ini menghindari kompleksitas mengelola dua saldo berbeda untuk satu nasabah.

## Decision

### 1. Arsitektur — Spending Interface, Bukan Balance Terpisah

E-wallet **bukan** saldo terpisah. E-wallet adalah **layer kontrol belanja** yang berada di atas rekening tabungan nasabah.

```
┌─────────────────────────────────────┐
│         E-Wallet Layer              │  ← Spending controls, kartu, QR
│  (daily limit, category restrict)   │
├─────────────────────────────────────┤
│      Rekening Tabungan (K002/K005)  │  ← Saldo aktual
│      balance: Rp 500.000           │
│      available_balance: Rp 500.000 │
└─────────────────────────────────────┘
```

**Keputusan ini berarti:**
- Saldo yang dibelanjakan melalui e-wallet **langsung mengurangi saldo tabungan**
- Tidak ada saldo e-wallet terpisah yang perlu di-reconcile
- Top-up e-wallet = setoran ke tabungan (transaksi standar K011)
- Semua aturan tabungan tetap berlaku (minimum balance, freeze, dll)

**Alasan memilih pendekatan ini:**
- Menghindari double-balance problem (dua saldo yang harus dijaga konsisten)
- Reuse semua infrastructure tabungan yang sudah ada (K002, K005, K011)
- Nasabah yang belanja via e-wallet atau via teller menggunakan saldo yang sama
- Laporan keuangan tetap sederhana — satu saldo per rekening

### 2. Parent Controls — Konfigurasi Belanja Anak

Orang tua mengontrol spending anak melalui konfigurasi:

```
ewallet_config
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah, siswa)
├── rekening_id           UUID (FK → rekening, tabungan yang digunakan)
│
├── ── Spending Limits ──
├── daily_limit           NUMERIC(15,2) NOT NULL (maks belanja per hari)
├── per_transaction_limit NUMERIC(15,2) (nullable, maks per transaksi)
├── weekly_limit          NUMERIC(15,2) (nullable, maks per minggu)
├── monthly_limit         NUMERIC(15,2) (nullable, maks per bulan)
│
├── ── Category Restrictions ──
├── allowed_categories    JSONB (nullable, daftar category_id yang dibolehkan)
├── blocked_categories    JSONB (nullable, daftar category_id yang diblokir)
│
├── ── Time Restrictions ──
├── allowed_hours_start   TIME (nullable, jam mulai boleh belanja)
├── allowed_hours_end     TIME (nullable, jam akhir boleh belanja)
├── allowed_days          JSONB (nullable, hari yang dibolehkan: ["mon","tue",...])
│
├── ── Pengaturan ──
├── is_active             BOOLEAN DEFAULT true
├── managed_by            ENUM (parent, admin, teacher)
├── parent_nasabah_id     UUID (nullable, FK → nasabah, orang tua)
│
├── ── PIN ──
├── pin_hash              VARCHAR (nullable, PIN untuk transaksi di atas threshold)
├── pin_threshold         NUMERIC(15,2) (nullable, nominal yang memerlukan PIN)
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

**Aturan parent controls:**
- Orang tua bisa set limit via teller atau self-service portal ([ADR-K023](./ADR-K023-dashboard-portal.md))
- Semua limit bersifat **opsional** — jika tidak di-set, menggunakan default tenant config
- `allowed_categories` dan `blocked_categories` bersifat **mutually exclusive** — gunakan salah satu
- Time restriction untuk membatasi belanja hanya di jam sekolah
- PIN diperlukan untuk transaksi di atas `pin_threshold` — melindungi dari penyalahgunaan

### 3. Kartu & Identifikasi

Siswa mengidentifikasi diri saat pembayaran via kartu fisik atau QR code:

```
ewallet_card
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
├── ewallet_config_id     UUID (FK → ewallet_config)
│
├── ── Card Info ──
├── card_type             ENUM (nfc, qr_static, qr_dynamic, barcode)
├── card_number           VARCHAR UNIQUE (nomor kartu fisik / identifier)
├── card_uid              VARCHAR (nullable, NFC UID)
│
├── ── Status ──
├── status                ENUM (active, frozen, deactivated, lost)
├── activated_at          TIMESTAMPTZ
├── frozen_at             TIMESTAMPTZ (nullable)
├── frozen_by             UUID (nullable)
├── deactivated_at        TIMESTAMPTZ (nullable)
│
├── ── Expiry ──
├── expires_at            DATE (nullable)
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

**Jenis kartu:**

| Type | Deskripsi | Use Case |
|------|-----------|----------|
| `nfc` | Kartu NFC/RFID, tap di reader | Paling cepat, butuh hardware NFC |
| `qr_static` | QR code tercetak di kartu | Murah, cukup scan kamera |
| `qr_dynamic` | QR code di-generate per transaksi | Paling aman, butuh device siswa |
| `barcode` | Barcode tercetak di kartu | Kompatibel scanner existing |

**Aturan kartu:**
- Satu nasabah bisa memiliki **multiple cards** (misal: NFC + QR backup)
- Hanya kartu dengan status **ACTIVE** yang bisa digunakan untuk transaksi
- Kartu hilang bisa **di-freeze** oleh orang tua/admin — langsung efektif
- Kartu yang di-deactivate tidak bisa di-reactivate — terbitkan kartu baru
- **Tidak ada saldo di kartu** — kartu hanya identifier, saldo di tabungan

### 4. Payment Flow via E-Wallet

```
Siswa tap kartu / scan QR di POS kantin
        │
        v
┌─────────────────────────────┐
│ STEP 1: Identifikasi         │
│ Lookup card → nasabah        │
│ Cek card status = ACTIVE     │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│ STEP 2: Validasi E-Wallet    │
│ Cek ewallet_config aktif     │
│ Cek daily_limit              │
│ Cek per_transaction_limit    │
│ Cek time restriction         │
│ Cek category restriction     │
│ Cek available_balance        │
└────────┬────────────────────┘
         │ ALL PASS
         v
┌─────────────────────────────┐
│ STEP 3: PIN (jika required)  │
│ amount > pin_threshold?      │
│ → minta PIN                  │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│ STEP 4: Execute              │
│ 1. Create transaksi K011     │
│    (penarikan dari tabungan) │
│ 2. Create penjualan POS K019 │
│ 3. Update saldo tabungan     │
│ All in 1 DB transaction      │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│ STEP 5: Receipt & Log        │
│ Print receipt POS             │
│ Log spending untuk parent     │
└─────────────────────────────┘
```

**Validasi berurutan — gagal di step manapun langsung reject:**

| Check | Pesan Reject |
|-------|-------------|
| Card not found | "Kartu tidak terdaftar" |
| Card frozen/deactivated | "Kartu dibekukan/nonaktif" |
| E-wallet config inactive | "E-wallet tidak aktif" |
| Daily limit exceeded | "Batas belanja harian tercapai" |
| Per-transaction limit exceeded | "Melebihi batas per transaksi" |
| Outside allowed hours | "Di luar jam belanja yang diizinkan" |
| Category not allowed | "Kategori barang tidak diizinkan" |
| Insufficient balance | "Saldo tabungan tidak mencukupi" |
| Wrong PIN | "PIN salah" (max 3 attempts → freeze card 30 min) |

### 5. Top-up Channels

Cara orang tua mengisi saldo tabungan anak:

```
top_up_channel:
├── TELLER              Orang tua datang ke koperasi, setoran via teller
├── TRANSFER            Transfer bank ke rekening koperasi (via K024 payment gateway)
├── PAYROLL_DEDUCT      Potong gaji orang tua (jika orang tua juga karyawan, via K020)
└── PARENT_TABUNGAN     Transfer dari tabungan orang tua ke tabungan anak (internal transfer)
```

**Aturan:**
- Semua top-up tercatat sebagai **setoran tabungan** di [ADR-K011](./ADR-K011-transaksi.md)
- Transfer internal antar nasabah (parent → child) tanpa fee (configurable per tenant)
- Top-up via transfer bank memerlukan integrasi payment gateway ([ADR-K024](./ADR-K024-integrasi-api.md))

### 6. Spending Tracking & Analytics

Tracking pengeluaran siswa untuk visibility orang tua:

```
Spending data derived from:
├── toko_penjualan (K019) WHERE nasabah_id = student
│   ├── Per item: produk, kategori, jumlah, harga
│   ├── Per transaksi: total, waktu, lokasi
│   └── Aggregasi: harian, mingguan, bulanan
│
└── Tersedia di:
    ├── Parent dashboard (K023)
    ├── Notifikasi harian (K022)
    └── Export CSV/PDF (K023)
```

**Data yang tersedia untuk orang tua:**

| Data | Granularity | Delivery |
|------|------------|----------|
| Detail transaksi | Per transaksi | Dashboard, notifikasi |
| Spending harian | Per hari | Notifikasi WhatsApp (opsional) |
| Spending mingguan | Per minggu | Dashboard |
| Spending bulanan | Per bulan | Dashboard, export |
| Category breakdown | Per bulan | Dashboard |
| Top items purchased | Per bulan | Dashboard |
| Remaining daily limit | Real-time | Dashboard |

### 7. Konteks Asrama (Boarding)

Untuk siswa boarding/asrama, e-wallet menjadi **satu-satunya** cara belanja karena tidak ada akses ke uang tunai:

**Aturan khusus boarding:**
- Admin bisa set flag `is_boarding = true` di ewallet_config — mengaktifkan fitur boarding-specific
- Meal plan integration ([ADR-K019](./ADR-K019-toko-kantin.md) Section 9) — makan otomatis dari tabungan
- Semua pengeluaran siswa boarding terlacak 100% — transparency penuh untuk orang tua
- Emergency spending bisa di-approve oleh wali kelas/wali asrama tanpa limit check

### 8. Teacher/Wali Kelas sebagai Proxy

Untuk siswa muda (SD/MI), wali kelas bisa bertindak sebagai proxy:

```
ewallet_proxy
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── student_nasabah_id    UUID (FK → nasabah, siswa)
├── proxy_user_id         UUID (FK → user, wali kelas)
│
├── ── Permission ──
├── can_make_purchase     BOOLEAN DEFAULT true
├── can_view_balance      BOOLEAN DEFAULT true
├── can_freeze_card       BOOLEAN DEFAULT false
├── max_transaction       NUMERIC(15,2) (nullable, limit per transaksi untuk proxy)
│
├── ── Periode ──
├── effective_date        DATE NOT NULL
├── end_date              DATE (nullable)
├── is_active             BOOLEAN DEFAULT true
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Aturan proxy:**
- Proxy harus di-assign oleh Admin atau orang tua (via teller)
- Proxy tetap terikat spending limit yang di-set orang tua
- Semua transaksi oleh proxy tercatat dengan `created_by = proxy_user_id`
- Proxy bisa dicabut kapan saja oleh orang tua atau admin

### 9. Lost Card Handling

```
Orang tua / siswa lapor kartu hilang
        │
        v
┌─────────────────────────────┐
│ STEP 1: Freeze kartu lama    │
│ card.status = FROZEN         │
│ Bisa oleh: parent, admin,    │
│ teacher (via proxy)          │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│ STEP 2: Deactivate           │
│ card.status = DEACTIVATED    │
│ Oleh: Teller / Admin         │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│ STEP 3: Issue new card       │
│ Create ewallet_card baru     │
│ Link ke nasabah & config     │
│ yang sama                    │
└─────────────────────────────┘
```

**Aturan:**
- **Tidak ada saldo yang hilang** — saldo ada di tabungan, bukan di kartu
- Kartu lama langsung di-freeze (mencegah penyalahgunaan) → kemudian di-deactivate
- Kartu baru bisa langsung diterbitkan — menggunakan ewallet_config yang sama
- Fee penggantian kartu configurable per tenant (default: Rp 10.000)
- Fee penggantian bisa dipotong dari tabungan atau bayar cash

### 10. Security

**PIN management:**
- PIN 6 digit, di-hash (bcrypt/argon2) — tidak disimpan plain
- PIN di-set oleh orang tua saat aktivasi e-wallet
- Reset PIN oleh orang tua via teller atau self-service portal
- 3x salah PIN → kartu di-freeze otomatis selama 30 menit
- 5x salah PIN dalam 24 jam → kartu di-freeze hingga manual unfreeze oleh admin

**Card freeze:**
- Orang tua bisa freeze kartu **kapan saja** — via teller, portal, atau notifikasi emergency ke admin
- Freeze langsung efektif (real-time)
- Unfreeze oleh orang tua atau admin

**Transaction limits:**
- Semua limit di-enforce real-time di setiap transaksi
- Limit aggregation (daily/weekly/monthly) dihitung dari SUM transaksi dalam periode

### 11. Vernon _rels dan _data Structure

**ewallet_config _rels/_data:**

```json
// _rels
{
  "tenant_id": "018f...",
  "nasabah_id": "018f...",
  "rekening_id": "018f...",
  "parent_nasabah_id": "018f..."
}

// _data
{
  "nasabah": {
    "id": "018f...",
    "full_name": "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000150"
  },
  "rekening": {
    "id": "018f...",
    "account_number": "TB-2026-JKT-00000150",
    "category": "tabungan"
  },
  "parent": {
    "id": "018f...",
    "full_name": "Haji Fauzi",
    "member_number": "KOP-2026-JKT-000042"
  }
}
```

**ewallet_card _rels/_data:**

```json
// _rels
{
  "tenant_id": "018f...",
  "nasabah_id": "018f...",
  "ewallet_config_id": "018f..."
}

// _data
{
  "nasabah": {
    "id": "018f...",
    "full_name": "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000150"
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.nasabah` di ewallet_config dan ewallet_card
- `RekeningUpdatedEvent` → update `_data.rekening` di ewallet_config

### 12. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin | Parent |
|------------|--------|------------|---------|-------|--------|
| Issue card | v | v | v | v | - |
| Freeze card | - | v | v | v | v |
| Unfreeze card | - | v | v | v | v |
| Deactivate card | v | v | v | v | - |
| Set spending limits | - | - | - | v | v |
| View child spending | - | - | - | - | v |
| Set PIN | - | - | - | - | v |
| Reset PIN | v | v | v | v | v |
| Configure ewallet | - | v | v | v | - |
| Assign proxy | - | - | - | v | v* |
| View all e-wallet data | - | v | v | v | - |
| System config (defaults) | - | - | - | v | - |

*Parent bisa request proxy assignment via teller — bukan self-service langsung.

**Catatan:**
- **Parent** memiliki akses khusus ke data anak — ini satu-satunya konteks di mana nasabah bisa mengelola data nasabah lain
- Parent access di-enforce berdasarkan `parent_nasabah_id` di ewallet_config — bukan akses bebas ke semua siswa
- Teller bisa issue dan deactivate card, tapi **tidak bisa set limits** — itu hak orang tua

### 13. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Module name | Uang Saku Digital | Uang Saku Digital |
| Balance label | Saldo Tabungan | Saldo Tabungan |
| Top-up label | Setoran | Setoran |

E-wallet bersifat **non-interest-bearing** — sama di kedua mode. Tidak ada perbedaan substansial.

## Consequences

### Positif

- **Cashless campus** — siswa tidak perlu bawa uang tunai, mengurangi risiko kehilangan/pemalakan
- **Parental control** — orang tua punya kontrol penuh atas belanja anak
- **Transparency** — semua pengeluaran tercatat dan bisa dipantau real-time
- **No double balance** — reuse tabungan existing, tidak ada saldo terpisah yang harus di-reconcile
- **Boarding-friendly** — siswa asrama bisa beroperasi fully cashless
- **Secure** — PIN, freeze, limit berlapis melindungi dari penyalahgunaan

### Negatif

- **Hardware dependency** — butuh NFC reader atau QR scanner di setiap POS
- **Card management overhead** — issue, replace, deactivate kartu untuk ratusan siswa
- **Parent onboarding** — orang tua harus memahami cara set limit dan monitor
- **Network dependency** — POS harus online untuk validasi real-time (offline mode TBD)

### Mitigasi

- QR static sebagai **fallback murah** — cetak QR di kartu, scan pakai kamera HP (tanpa hardware khusus)
- Batch card issuance untuk awal tahun ajaran — template import data siswa
- Panduan orang tua disediakan saat pendaftaran — simple, visual, step-by-step
- Offline mode bisa dipertimbangkan di fase 2 — dengan risk-based limit (offline hanya untuk transaksi kecil)

## Alternatives Considered

### A. Saldo E-Wallet Terpisah (Separate Balance)

E-wallet memiliki saldo sendiri, terpisah dari tabungan.

**Ditolak** karena: menambah kompleksitas — dua saldo per nasabah yang harus di-reconcile, dua set transaksi, dan bingung bagi nasabah ("saldo tabungan saya berapa?" vs "saldo e-wallet saya berapa?"). Top-up juga menjadi transaksi internal yang tidak perlu.

### B. Physical Cash Card (Stored Value Card)

Kartu prabayar dengan saldo tersimpan di kartu fisik.

**Ditolak** karena: saldo di kartu tidak bisa di-reconcile real-time, risiko kehilangan saldo jika kartu hilang/rusak, dan memerlukan hardware khusus (card writer) yang mahal.

### C. Tanpa Kartu — Manual Input Nasabah ID

Kasir input nomor anggota siswa secara manual di POS.

**Ditolak** karena: terlalu lambat untuk high-volume canteen (antrian jam istirahat). Juga rawan human error (salah input nomor). Kartu/QR jauh lebih efisien dan akurat.
