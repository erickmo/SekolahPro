# ADR-K002: Rekening (Akun Nasabah)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Rekening adalah **central nexus** antara nasabah dan seluruh aktivitas keuangan di koperasi/BMT. Setiap transaksi — setoran, penarikan, angsuran, bagi hasil — terjadi melalui rekening. Sistem harus mengakomodasi:

- **Multi-rekening per nasabah**: satu nasabah bisa memiliki beberapa rekening dengan jenis berbeda (tabungan, deposito, pinjaman)
- **Relasi ke produk**: setiap rekening terikat pada satu produk koperasi (lihat [ADR-K003](./ADR-K003-produk-akad.md))
- **Saldo real-time**: saldo harus selalu akurat, derivasi dari riwayat transaksi dengan cache di rekening
- **Dual-mode terminologi**: Rekening (konvensional) vs Rekening/Hisab (BMT) — lihat [ADR-009](../core/ADR-009-dual-mode-institution-type.md)
- **Nomor rekening unik**: auto-generate dengan format yang bisa dikonfigurasi per tenant

## Decision

### 1. Rekening Categories

Rekening dikelompokkan berdasarkan **kategori** yang menentukan perilaku dan aturan bisnisnya:

```
rekening_category:
├── SIMPANAN_POKOK      Wajib, 1x seumur keanggotaan, tidak bisa ditarik selama aktif
├── SIMPANAN_WAJIB      Wajib, bulanan, tidak bisa ditarik selama aktif
├── TABUNGAN            Sukarela, bisa multiple per nasabah, bisa tarik kapan saja
├── DEPOSITO            Berjangka, ada jatuh tempo, penalti jika tarik sebelum tempo
└── PINJAMAN            Kewajiban nasabah ke koperasi (saldo = outstanding pokok)
```

**Aturan per kategori:**

| Kategori | Max per Nasabah | Setoran | Penarikan | Auto-create |
|----------|-----------------|---------|-----------|-------------|
| SIMPANAN_POKOK | 1 | 1x saat join | Saat keluar saja | Ya, saat approval nasabah |
| SIMPANAN_WAJIB | 1 | Bulanan (wajib) | Saat keluar saja | Ya, saat approval nasabah |
| TABUNGAN | Unlimited (per produk) | Kapan saja | Kapan saja (min. saldo) | Tidak, manual |
| DEPOSITO | Unlimited | 1x saat buka | Saat jatuh tempo | Tidak, manual |
| PINJAMAN | Sesuai kebijakan | N/A (pencairan) | N/A (angsuran) | Tidak, via approval pinjaman |

### 2. Auto-create Rekening Wajib

Saat nasabah di-approve (status → ACTIVE), sistem **otomatis** membuat dua rekening:

```
NasabahApprovedEvent
        │
        v
┌───────────────────────────────┐
│ Auto-create:                  │
│ 1. Rekening SIMPANAN_POKOK    │
│ 2. Rekening SIMPANAN_WAJIB    │
└───────────────────────────────┘
```

Kedua rekening ini:
- Menggunakan **produk default** yang sudah dikonfigurasi tenant
- Nomor rekening di-generate otomatis
- Status langsung **ACTIVE**
- Saldo awal = 0 (setoran dicatat sebagai transaksi terpisah)
- **Tanpa form pengajuan** — langsung dibuat oleh sistem sebagai bagian dari proses keanggotaan

### 3. Pengajuan Rekening — Application Form + Approval

Untuk rekening **non-wajib** (tabungan, deposito, pinjaman), nasabah harus mengajukan pembukaan rekening melalui formulir pengajuan yang di-approve sebelum rekening aktif.

```
Nasabah (via Teller) mengisi form pengajuan
        │
        v
┌─────────────────┐
│  Status: DRAFT  │  Teller bisa edit sebelum submit
└────────┬────────┘
         │ submit
         v
┌─────────────────┐
│ Status: PENDING │  Menunggu approval
└────────┬────────┘
         │
    ┌────┴────┐
    v         v
┌────────┐ ┌──────────┐
│APPROVED│ │ REJECTED │  Reviewer bisa beri catatan penolakan
└───┬────┘ └──────────┘
    │ auto-create rekening
    v
┌─────────────────┐
│ Rekening ACTIVE │  Nomor rekening di-generate, siap transaksi
└─────────────────┘
```

**Pengajuan per kategori:**

| Kategori | Perlu Pengajuan? | Approval Level | Catatan |
|----------|------------------|----------------|---------|
| SIMPANAN_POKOK | Tidak — auto-create saat nasabah di-approve | - | Bagian dari proses keanggotaan |
| SIMPANAN_WAJIB | Tidak — auto-create saat nasabah di-approve | - | Bagian dari proses keanggotaan |
| TABUNGAN | **Ya** | Supervisor+ | Pilih produk tabungan, setoran awal opsional |
| DEPOSITO | **Ya** | Supervisor+ | Pilih produk, nominal, tenor, jatuh tempo |
| PINJAMAN | **Ya** | Manager+ | Lebih ketat — detail di [ADR-K007](./ADR-K007-pinjaman.md) |

**Aturan pengajuan:**
- Hanya nasabah **ACTIVE** yang bisa mengajukan rekening baru
- Form yang sudah di-submit **tidak bisa di-edit** oleh Teller — harus ditolak lalu ajukan ulang
- Teller bisa buat pengajuan tapi **tidak bisa approve sendiri** — separation of duties (konsisten dengan K001)
- Pengajuan pinjaman memiliki **approval level lebih tinggi** (Manager+) karena risiko finansial
- Rejection wajib menyertakan alasan (`rejection_reason`)
- Pengajuan yang di-approve otomatis membuat rekening baru — Teller tidak perlu buat rekening manual
- Riwayat approval tersimpan di audit log

**Validasi saat submit:**

```
Submit Pre-check:
├── nasabah.status = ACTIVE                     ✓
├── Produk valid & aktif untuk tenant ini        ✓
├── Kategori sesuai dengan produk yang dipilih   ✓
├── Tidak melebihi max rekening per kategori     ✓
│   (misal: tabungan unlimited, tapi cek per produk)
├── KYC level mencukupi (jika produk butuh full) ✓
└── [DEPOSITO] nominal ≥ minimum penempatan      ✓
    │
    ALL PASS → Submit allowed
    ANY FAIL → Submit blocked, return error list
```

**Customizable Form Template:**

Mirip dengan nasabah application form (K001 Section 2), admin tenant dapat meng-customize template formulir pengajuan rekening:

```
rekening_form_template
├── category              ← Template per kategori rekening
├── sections[]
│   ├── product_selection   ← System-defined, pilih produk
│   ├── account_details     ← System-defined, detail sesuai kategori
│   │   ├── [TABUNGAN]  → setoran awal (opsional)
│   │   ├── [DEPOSITO]  → nominal, tenor, jatuh tempo, rollover preference
│   │   └── [PINJAMAN]  → plafon, tujuan, tenor — detail di K007
│   ├── custom_section_*    ← Tenant-defined
│   └── terms_conditions    ← System-defined, konten bisa di-edit
│       └── [BMT mode]  → termasuk Akad syariah sesuai produk
├── version               ← Increment setiap perubahan
└── coop_type_variant     ← Template bisa berbeda per mode
```

**Aturan template:**
- Template dikelompokkan **per kategori** — form tabungan berbeda dari deposito
- Section `product_selection`, `account_details`, dan `terms_conditions` adalah **system-defined** — tidak bisa dihapus
- Tenant bisa menambahkan **custom sections** untuk kebutuhan spesifik
- Setiap perubahan template menghasilkan **version baru**
- Pengajuan yang di-submit menyimpan `template_version` (immutable snapshot)
- BMT mode: T&C otomatis menyertakan **Akad** sesuai jenis produk (Mudharabah, Murabahah, dll)

### 4. Rekening Application Data Model

```
rekening_application
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
├── nasabah_id            UUID (FK → nasabah) NOT NULL
│
├── ── Pilihan Rekening ──
├── product_id            UUID (FK → produk) NOT NULL
├── category              ENUM (tabungan, deposito, pinjaman)
│
├── ── Detail Pengajuan ──
├── form_data             JSONB (snapshot seluruh isian form)
├── initial_deposit       NUMERIC(15,2) (nullable, setoran awal yang direncanakan)
│
├── ── Khusus Deposito ──
├── deposit_amount        NUMERIC(15,2) (nullable, nominal penempatan)
├── tenor_months          INTEGER (nullable, jangka waktu dalam bulan)
├── maturity_date         DATE (nullable, auto-calculate dari tenor)
├── rollover_instruction  ENUM (none, principal_only, principal_and_profit) DEFAULT 'none'
│
├── ── Khusus Pinjaman (ringkasan — detail di K007) ──
├── loan_amount           NUMERIC(15,2) (nullable, plafon yang diajukan)
├── loan_tenor_months     INTEGER (nullable)
├── loan_purpose          TEXT (nullable, tujuan pinjaman)
│
├── ── Template ──
├── template_version      INTEGER (versi template yang digunakan)
├── terms_accepted        BOOLEAN NOT NULL
├── terms_accepted_at     TIMESTAMPTZ
│
├── ── Workflow ──
├── status                ENUM (draft, pending, approved, rejected)
├── submitted_at          TIMESTAMPTZ (nullable)
├── reviewed_at           TIMESTAMPTZ (nullable)
├── reviewed_by           UUID (nullable, FK → user)
├── rejection_reason      TEXT (nullable)
│
├── ── Result ──
├── rekening_id           UUID (nullable, FK → rekening, filled on approval)
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

**Vernon _rels/_data untuk rekening_application:**

```json
// _rels
{
  "nasabah_id": "018f...",
  "product_id": "018f...",
  "branch_id":  "018f...",
  "tenant_id":  "018f..."
}

// _data
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "product": {
    "id":   "018f...",
    "name": "Tabungan Berkah",
    "code": "TB-BERKAH"
  },
  "branch": {
    "id":   "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  }
}
```

### 5. Approval on Create Event

Saat pengajuan rekening di-approve, sistem men-trigger event yang menjalankan proses pembuatan rekening secara otomatis:

```
RekeningApplicationApprovedEvent
        │
        v
┌───────────────────────────────────────┐
│ 1. Generate account_number (auto)     │
│ 2. Create rekening record             │
│    - status = ACTIVE                  │
│    - balance = 0                      │
│    - link ke application              │
│ 3. Update application.rekening_id     │
│ 4. [DEPOSITO] Set maturity_date       │
│ 5. [BMT] Catat akad di rekening       │
└───────────────────────────────────────┘
```

**Aturan:**
- Semua dalam **satu database transaction** — jika gagal, seluruhnya rollback
- Rekening langsung ACTIVE — tidak ada status pending di level rekening
- Setoran awal (jika ada) dicatat sebagai **transaksi terpisah** setelah rekening dibuat — bukan bagian dari proses approval
- Application tetap tersimpan sebagai audit trail — `rekening_id` menjadi link antara pengajuan dan rekening yang dihasilkan

### 6. Account Number — Auto-generate with Configuration

Nomor rekening di-generate otomatis menggunakan format yang bisa dikonfigurasi per tenant:

```
account_number_config:
  mode: "auto" | "manual"          # Tenant config, default: "auto"
  prefix_by_category:
    SIMPANAN_POKOK: "SP"
    SIMPANAN_WAJIB: "SW"
    TABUNGAN:       "TB"
    DEPOSITO:       "DP"
    PINJAMAN:       "PJ"
  separator: "-"                    # Configurable separator
  sequence_digits: 8                # Jumlah digit sequence
  include_year: true                # Sertakan tahun
  include_branch: true              # Sertakan kode cabang
```

**Contoh hasil:**
```
auto + all options:    TB-2026-JKT-00000001    (Tabungan, tahun 2026, cabang JKT)
auto + minimal:        TB-00000001
manual:                Operator input manual (harus unique, validated)
```

**Aturan:**
- Mode `auto` adalah default — `manual` hanya tersedia jika diaktifkan di tenant config
- Nomor rekening **immutable** setelah assigned — tidak bisa diubah
- Uniqueness di-enforce di database level (unique constraint per tenant)
- Sequence auto-increment per kategori, tidak ada reuse nomor yang sudah dipakai
- Prefix per kategori memudahkan identifikasi jenis rekening secara visual

### 7. Account Status Lifecycle

```
┌─────────────────┐
│ Status: ACTIVE  │  Rekening baru langsung aktif
└────────┬────────┘
         │
    ┌────┴────────────┐
    v                 v
┌────────┐     ┌──────────┐
│ FROZEN │     │  CLOSED  │
└───┬────┘     └──────────┘
    │
    v
┌─────────────────┐
│ Status: ACTIVE  │  Unfreeze mengembalikan ke ACTIVE
└─────────────────┘
```

**Status definitions:**

| Status | Deskripsi | Transaksi Masuk | Transaksi Keluar |
|--------|-----------|-----------------|------------------|
| ACTIVE | Normal, bisa bertransaksi | Diizinkan | Diizinkan |
| FROZEN | Dibekukan sementara (investigasi, sengketa, dll) | Diizinkan | Diblokir |
| CLOSED | Ditutup permanen, saldo = 0 | Diblokir | Diblokir |

**Aturan transisi:**
- `ACTIVE → FROZEN`: oleh Supervisor+ dengan alasan wajib (`freeze_reason`)
- `FROZEN → ACTIVE`: oleh Manager+ (unfreeze)
- `ACTIVE → CLOSED` atau `FROZEN → CLOSED`: hanya jika saldo = 0, oleh Manager+
- `CLOSED → *`: tidak bisa di-reactivate — nasabah harus buka rekening baru
- FROZEN membolehkan transaksi masuk (setoran, angsuran) tapi memblokir transaksi keluar (penarikan) — melindungi dana tanpa menghambat pembayaran kewajiban

### 8. Balance Management

Saldo rekening menggunakan pendekatan **cached balance + transaction-derived verification**:

```
                    ┌─────────────────────┐
                    │     rekening        │
                    │  balance = 1.500.000│  ← Cached, updated setiap transaksi
                    └──────────┬──────────┘
                               │
             ┌─────────────────┼─────────────────┐
             v                 v                 v
        ┌─────────┐     ┌─────────┐       ┌─────────┐
        │ TRX-001 │     │ TRX-002 │       │ TRX-003 │
        │ +500.000│     │ +1.200.0│       │ -200.000│
        └─────────┘     └─────────┘       └─────────┘
                    SUM = 1.500.000  ← Source of truth
```

**Aturan:**
- `balance` di tabel rekening adalah **cache** — di-update secara atomik bersamaan dengan insert transaksi (dalam satu database transaction)
- **Source of truth** adalah SUM dari seluruh transaksi rekening tersebut
- Scheduled job melakukan **reconciliation** harian: membandingkan cached balance vs SUM transaksi, alert jika ada selisih
- `balance` tidak boleh < 0 untuk rekening simpanan (kecuali ada konfigurasi `allow_overdraft` di produk)
- Untuk rekening PINJAMAN, `balance` = sisa pokok outstanding (berkurang setiap angsuran)

### 9. Minimum Balance (Saldo Minimum)

Setiap produk bisa mendefinisikan **saldo minimum** yang harus dijaga:

```
minimum_balance rules:
├── SIMPANAN_POKOK    → Tidak ada min (saldo tetap sampai keluar)
├── SIMPANAN_WAJIB    → Tidak ada min (saldo tetap sampai keluar)
├── TABUNGAN          → Configurable per produk (misal: Rp 10.000)
├── DEPOSITO          → N/A (saldo tetap sampai jatuh tempo)
└── PINJAMAN          → N/A
```

**Aturan:**
- Penarikan ditolak jika `(balance - amount) < minimum_balance`
- Minimum balance di-set di level **produk**, bukan di level rekening individual
- Admin bisa override minimum balance per rekening untuk kasus khusus (documented, requires Manager+ approval)

### 10. Rekening Data Model

```
rekening
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
├── nasabah_id            UUID (FK → nasabah) NOT NULL
├── product_id            UUID (FK → produk) NOT NULL
│
├── ── Identitas ──
├── account_number        VARCHAR UNIQUE per tenant
├── category              ENUM (simpanan_pokok, simpanan_wajib, tabungan, deposito, pinjaman)
├── account_name          VARCHAR (default: nama nasabah + jenis produk)
│
├── ── Saldo ──
├── balance               NUMERIC(15,2) NOT NULL DEFAULT 0
├── available_balance     NUMERIC(15,2) NOT NULL DEFAULT 0
├── hold_amount           NUMERIC(15,2) NOT NULL DEFAULT 0
│
├── ── Status ──
├── status                ENUM (active, frozen, closed)
├── frozen_at             TIMESTAMPTZ (nullable)
├── frozen_by             UUID (nullable, FK → user)
├── freeze_reason         TEXT (nullable)
├── closed_at             TIMESTAMPTZ (nullable)
├── closed_by             UUID (nullable, FK → user)
├── close_reason          TEXT (nullable)
│
├── ── Konfigurasi ──
├── minimum_balance       NUMERIC(15,2) DEFAULT 0 (override dari produk)
├── allow_overdraft       BOOLEAN DEFAULT false
│
├── ── Tanggal Penting ──
├── opened_at             TIMESTAMPTZ NOT NULL
├── last_transaction_at   TIMESTAMPTZ (nullable)
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

### 11. Available Balance vs Balance

Sistem membedakan **balance** dan **available_balance** untuk mengakomodasi hold/block saldo:

```
balance           = Total saldo aktual (dari transaksi)
hold_amount       = Saldo yang di-hold (misal: proses kliring, jaminan)
available_balance = balance - hold_amount  ← Yang bisa ditarik nasabah
```

**Use case hold:**
- Deposito yang belum jatuh tempo tapi nasabah mengajukan pencairan dini → saldo di-hold selama proses approval
- Pinjaman dijaminkan dengan saldo tabungan → sebagian saldo di-hold
- Proses kliring transfer antar cabang

**Aturan:**
- Penarikan menggunakan `available_balance`, bukan `balance`
- Hold hanya bisa dilakukan oleh Supervisor+ dengan alasan
- Hold otomatis di-release saat proses selesai (timeout configurable)

### 12. Vernon _rels dan _data Structure

Rekening menggunakan Vernon pattern karena listing rekening membutuhkan data nasabah, produk, dan cabang (≥ 3 JOIN).

**_rels:**
```json
{
  "nasabah_id":  "018f...",
  "product_id":  "018f...",
  "branch_id":   "018f...",
  "tenant_id":   "018f..."
}
```

**_data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001",
    "identity_number": "3201..."
  },
  "product": {
    "id":            "018f...",
    "name":          "Tabungan Berkah",
    "code":          "TB-BERKAH",
    "category":      "tabungan"
  },
  "branch": {
    "id":   "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.nasabah` di semua rekening nasabah tersebut
- `ProductUpdatedEvent` → update `_data.product` di semua rekening dengan produk tersebut
- `BranchUpdatedEvent` → update `_data.branch` di semua rekening cabang tersebut

### 13. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| Buat pengajuan rekening (tabungan/deposito) | v | v | v | v |
| Buat pengajuan rekening (pinjaman) | v | v | v | v |
| Approve/Reject pengajuan (tabungan/deposito) | - | v | v | v |
| Approve/Reject pengajuan (pinjaman) | - | - | v | v |
| Edit form template pengajuan | - | - | v | v |
| View rekening & saldo | v | v | v | v |
| View daftar pengajuan | v | v | v | v |
| Freeze rekening | - | v | v | v |
| Unfreeze rekening | - | - | v | v |
| Tutup rekening | - | - | v | v |
| Override minimum balance | - | - | v | v |
| Hold/release saldo | - | v | v | v |

**Catatan:**
- Teller bisa buat pengajuan tapi **tidak bisa approve sendiri** — separation of duties (konsisten dengan K001)
- Pengajuan pinjaman butuh **Manager+** untuk approve — risiko finansial lebih tinggi
- Override minimum balance membutuhkan Manager+ dan di-log ke audit trail
- Edit form template pengajuan membutuhkan Manager+ — template mempengaruhi semua pengajuan baru

### 14. Closing Rules

Rekening hanya bisa ditutup jika memenuhi **semua** kondisi:

```
Closing Pre-check:
├── balance           = 0  ✓
├── hold_amount       = 0  ✓
├── No pending transactions   ✓
└── Not referenced by active loan collateral  ✓
    │
    ALL MET  → Closing allowed
    ANY FAIL → Closing blocked, return list of blocking items
```

**Aturan:**
- Rekening SIMPANAN_POKOK dan SIMPANAN_WAJIB **tidak bisa ditutup individual** — hanya ditutup saat nasabah di-deactivate (lihat [ADR-K001](./ADR-K001-nasabah.md) bagian Deactivation)
- Penutupan meng-set `status = closed`, `closed_at = NOW()`, wajib isi `close_reason`
- Data rekening **tidak dihapus** (soft close) — tetap bisa diakses untuk audit, pelaporan, dan statement historis
- Nomor rekening yang sudah ditutup **tidak di-reuse**

### 15. Dormant Account Handling

Rekening tabungan yang tidak ada transaksi selama periode tertentu dianggap **dormant**:

```
dormant_config:
  inactive_period_months: 12       # Configurable per tenant
  notification_before_days: 30     # Kirim notifikasi sebelum dormant
  admin_fee_on_dormant: true       # Boleh potong biaya admin
  admin_fee_amount: 5000           # Nominal biaya admin bulanan
  auto_close_after_months: 36     # Auto-close jika saldo habis karena admin fee
```

**Flow:**
```
Tidak ada transaksi selama 12 bulan
        │
        v
┌────────────────────┐
│ Flag: DORMANT      │  (bukan status baru, tapi flag di rekening)
│ dormant_since: ... │
└────────┬───────────┘
         │
    ┌────┴────┐
    v         v
 Transaksi   Saldo habis (dari admin fee)
 masuk        │
    │         v
    v     Auto-close (jika dikonfigurasi)
 Flag dihapus
 (kembali normal)
```

**Aturan:**
- Dormant adalah **flag** (`is_dormant`, `dormant_since`), bukan status terpisah — rekening tetap ACTIVE
- Nasabah bisa mengaktifkan kembali kapan saja dengan melakukan transaksi
- Admin fee dormant hanya berlaku jika dikonfigurasi di tenant config
- Notifikasi dikirim 30 hari sebelum rekening di-flag dormant

### 16. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Entity name | Rekening | Rekening |
| Account categories | Simpanan Pokok, Simpanan Wajib, Tabungan, Deposito, Pinjaman | Simpanan Pokok, Simpanan Wajib, Tabungan, Deposito, Pembiayaan |
| Interest/profit label | Bunga | Bagi Hasil / Margin |
| Loan account label | Rekening Pinjaman | Rekening Pembiayaan |
| Application form title | Formulir Pengajuan Rekening | Formulir Pengajuan Rekening + Akad |
| Application T&C | Syarat & Ketentuan Rekening | Syarat & Ketentuan Rekening + Akad (sesuai produk) |

BMT mode menambahkan **Akad** (kontrak syariah) pada setiap formulir pengajuan rekening — jenis akad ditentukan oleh produk yang dipilih (Mudharabah untuk tabungan, Murabahah untuk pembiayaan, dll).

## Consequences

### Positif

- **Auditable** — setiap pembukaan rekening punya jejak dari pengajuan hingga approval (konsisten dengan K001)
- **Compliant** — formulir formal dengan T&C/akad memenuhi regulasi koperasi/OJK
- **Flexible** — multi-rekening per nasabah mengakomodasi berbagai jenis produk koperasi
- **Safe** — balance management dengan cached balance + reconciliation mencegah inkonsistensi saldo
- **Fraud prevention** — Teller tidak bisa approve sendiri, separation of duties
- **Performant** — Vernon pattern memastikan listing rekening < 100ms tanpa JOIN
- **Hold mechanism** — available_balance mencegah penarikan dana yang sedang dalam proses
- **Template customizable** — form pengajuan bisa disesuaikan per tenant dan per kategori

### Negatif

- **Pembukaan lebih lambat** — butuh approval untuk tabungan/deposito (bukan instant)
- **Complexity** — tiga field saldo (balance, available_balance, hold_amount) membutuhkan pengelolaan yang cermat
- **Reconciliation overhead** — scheduled job harian untuk verifikasi saldo menambah beban sistem
- **Vernon sync** — perubahan data nasabah/produk memerlukan propagasi ke semua rekening terkait
- **Template versioning** — harus maintain snapshot per pengajuan (konsisten dengan K001, shared complexity)

### Mitigasi

- Supervisor+ bisa approve secara **batch** untuk efisiensi — beberapa pengajuan tabungan sekaligus
- Balance update selalu dalam satu database transaction dengan insert transaksi — atomicity dijamin oleh database
- Reconciliation berjalan di off-peak hours dan hanya melakukan SELECT (tidak write) — impact minimal
- SyncEngine menggunakan batching untuk propagasi massal — efisien untuk update nama nasabah yang memiliki banyak rekening
- Template versioning otomatis (increment on save), reuse pattern dari K001

## Alternatives Considered

### A. Direct Opening tanpa Application Form

Teller langsung membuat rekening tanpa proses pengajuan dan approval.

**Ditolak** karena: tidak ada audit trail pembukaan rekening, tidak ada separation of duties (Teller bisa buka rekening sendiri tanpa kontrol), dan tidak memenuhi regulasi koperasi yang mewajibkan dokumentasi pembukaan rekening. Juga tidak konsisten dengan pola K001 yang sudah menggunakan application + approval.

### B. Single Account per Nasabah (Omnibus)

Satu rekening untuk semua jenis simpanan/pinjaman, dibedakan oleh sub-account.

**Ditolak** karena: koperasi memiliki regulasi yang mengharuskan pemisahan jelas antara simpanan pokok, simpanan wajib, tabungan sukarela, dan pinjaman. Omnibus account menyulitkan pelaporan per-jenis dan audit OJK/Dinas Koperasi.

### B. Balance dari SUM Transaksi (tanpa cache)

Menghitung saldo real-time dari SUM seluruh transaksi setiap kali dibutuhkan.

**Ditolak** karena: performa tidak acceptable untuk nasabah dengan ribuan transaksi. Query SUM pada tabel transaksi besar bisa > 500ms. Cached balance dengan reconciliation memberikan keseimbangan antara performa dan akurasi.

### C. Separate Status untuk Dormant

Menambah `DORMANT` sebagai status rekening terpisah di enum.

**Ditolak** karena: dormant bukan status operasional — rekening dormant masih bisa menerima transaksi. Menambahnya sebagai status memperumit state machine dan memerlukan transisi tambahan. Flag boolean lebih sesuai karena dormant bersifat informasional dan auto-reversible.

### D. Balance Hanya di Rekening (tanpa reconciliation)

Mempercayai cached balance sepenuhnya tanpa verifikasi periodik.

**Ditolak** karena: risiko inkonsistensi saldo terlalu tinggi untuk sistem finansial. Bug, race condition, atau failed transaction bisa menyebabkan saldo tidak akurat. Reconciliation adalah safety net yang wajib ada di sistem keuangan.

### E. Hold sebagai Transaksi Terpisah

Mencatat hold sebagai transaksi dengan tipe khusus, bukan field di rekening.

**Ditolak** karena: hold bersifat temporer dan reversible — bukan transaksi finansial yang sebenarnya. Mencatatnya sebagai transaksi memperumit laporan keuangan dan rekonsiliasi. Field `hold_amount` di rekening lebih sederhana dan langsung ter-reflect di `available_balance`.
