# ADR-K013: Money Denomination (Pecahan Uang)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Pengelolaan kas fisik di koperasi/BMT membutuhkan pencatatan uang berdasarkan **pecahan (denominasi)**. Hal ini diperlukan untuk:

- **Cash counting**: penghitungan kas teller saat buka dan tutup session ([ADR-K012](./ADR-K012-teller-session.md))
- **Vault management**: mengetahui komposisi kas di brankas cabang
- **Cash transfer**: serah terima kas antar teller/vault dengan detail per denominasi
- **Audit & compliance**: laporan kas detail untuk pengawas koperasi
- **Dual-mode**: terminologi denominasi **identik** di kedua mode — lihat [ADR-009](../core/ADR-009-dual-mode-institution-type.md)

Denominasi Rupiah terdiri dari uang kertas dan uang logam dengan nilai tertentu yang bisa berubah seiring waktu (Bank Indonesia bisa menerbitkan denominasi baru atau menarik yang lama).

## Decision

### 1. Indonesian Money Denominations

Denominasi Rupiah yang berlaku saat ini:

```
UANG KERTAS (Banknote):
┌───────────────┬───────────────────────┐
│ Nilai (Rp)    │ Keterangan            │
├───────────────┼───────────────────────┤
│ 100.000       │ Pecahan terbesar      │
│  50.000       │                       │
│  20.000       │                       │
│  10.000       │                       │
│   5.000       │                       │
│   2.000       │                       │
│   1.000       │ Pecahan kertas terkecil│
└───────────────┴───────────────────────┘

UANG LOGAM (Coin):
┌───────────────┬───────────────────────┐
│ Nilai (Rp)    │ Keterangan            │
├───────────────┼───────────────────────┤
│   1.000       │ Ada versi kertas & logam│
│     500       │                       │
│     200       │                       │
│     100       │ Pecahan terkecil      │
└───────────────┴───────────────────────┘
```

**Catatan:**
- Denominasi Rp 1.000 tersedia dalam bentuk kertas dan logam — sistem mencatat keduanya sebagai entry terpisah (type: `banknote` vs `coin`)
- Denominasi di bawah Rp 100 (misal: Rp 50, Rp 25) sudah tidak umum digunakan — tidak disertakan di default tapi bisa ditambahkan jika diperlukan

### 2. Denomination Master

Katalog denominasi yang tersedia di-maintain sebagai master data:

```
denominasi_master
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant) NOT NULL
│
├── ── Denomination Info ──
├── value                 INTEGER NOT NULL (nilai nominal: 100000, 50000, dll)
├── currency              VARCHAR NOT NULL DEFAULT 'IDR'
├── type                  ENUM (banknote, coin) NOT NULL
├── label                 VARCHAR NOT NULL (display: "Rp 100.000 Kertas")
├── sort_order            INTEGER NOT NULL (urutan tampil, dari besar ke kecil)
│
├── ── Status ──
├── is_active             BOOLEAN NOT NULL DEFAULT true
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Aturan:**
- Master denominasi **per tenant** — setiap tenant punya katalog sendiri
- Default set di-seed saat tenant onboarding (seluruh denominasi Rupiah yang berlaku)
- Admin bisa **menambah** denominasi baru (future-proof untuk pecahan baru dari BI)
- Admin bisa **menonaktifkan** denominasi (`is_active = false`) — denominasi yang tidak aktif tidak muncul di form counting tapi tetap tersimpan di data historis
- Denominasi **tidak bisa dihapus** (soft deactivate) — data historis mereferensikan denominasi tersebut
- `sort_order` menentukan urutan tampil di form counting — default dari nilai terbesar ke terkecil
- `currency` default `IDR` — field ini disiapkan untuk kebutuhan multi-currency di masa depan (tidak digunakan saat ini)

### 3. Usage Points

Denominasi digunakan di beberapa titik operasional:

```
Usage Points:
├── 1. Teller Session — Opening (K012)
│      Hitung denominasi kas awal yang diterima dari vault
│
├── 2. Teller Session — Closing (K012)
│      Hitung denominasi kas akhir di drawer teller
│
├── 3. Vault Count
│      Penghitungan berkala kas di brankas cabang
│
├── 4. Cash Transfer — Teller ↔ Vault
│      Serah terima kas mid-session (top-up atau setor sebagian)
│
├── 5. Cash Transfer — Branch ↔ Branch (K014)
│      Transfer kas fisik antar cabang
│
└── 6. Cash Transfer — Branch ↔ HQ
       Setor ke pusat atau terima dari pusat
```

Setiap usage point mencatat denominasi dengan struktur yang sama — jumlah per pecahan dan total.

### 4. Denomination Record Structure

Setiap pencatatan denominasi menggunakan format yang konsisten:

```
Contoh: Kas Awal Teller (Opening)

┌───────────────┬──────────┬──────────────┐
│ Denominasi    │ Jumlah   │ Subtotal     │
├───────────────┼──────────┼──────────────┤
│ Rp 100.000    │ 20 lbr   │ 2.000.000    │
│ Rp  50.000    │ 30 lbr   │ 1.500.000    │
│ Rp  20.000    │ 25 lbr   │   500.000    │
│ Rp  10.000    │ 50 lbr   │   500.000    │
│ Rp   5.000    │ 40 lbr   │   200.000    │
│ Rp   2.000    │ 50 lbr   │   100.000    │
│ Rp   1.000    │ 50 lbr   │    50.000    │
│ Rp   1.000 *  │ 50 kpg   │    50.000    │
│ Rp     500    │ 100 kpg  │    50.000    │
│ Rp     200    │ 50 kpg   │    10.000    │
│ Rp     100    │ 100 kpg  │    10.000    │
├───────────────┼──────────┼──────────────┤
│ TOTAL         │          │ 4.970.000    │
└───────────────┴──────────┴──────────────┘
* logam
lbr = lembar, kpg = keping
```

**Aturan kalkulasi:**
- `subtotal = denomination_value * quantity`
- `total = SUM(subtotal)` untuk semua denominasi dalam satu record set
- Total harus **match** dengan nominal yang diharapkan (misal: kas_awal, actual_cash)
- Jika total denominasi tidak match dengan nominal yang di-input, sistem **menolak** dan meminta recount
- Quantity bisa = 0 untuk denominasi yang tidak ada di set tersebut

### 5. Denomination Configurability

Denominasi bisa dikonfigurasi untuk mengakomodasi perubahan di masa depan:

```
Denomination Management Flow:

Admin menambah denominasi baru
(misal: BI menerbitkan Rp 75.000)
        │
        v
┌───────────────────────────────────────┐
│ 1. Admin buat entry di denominasi_master │
│    value: 75000                          │
│    type: banknote                        │
│    label: "Rp 75.000 Kertas"            │
│    sort_order: (antara 50000 dan 100000) │
│    is_active: true                       │
│ 2. Denominasi baru muncul di form counting│
│ 3. Data historis tidak terpengaruh       │
└───────────────────────────────────────┘

Admin menonaktifkan denominasi lama
(misal: Rp 200 logam jarang digunakan)
        │
        v
┌───────────────────────────────────────┐
│ 1. Admin set is_active = false        │
│ 2. Denominasi tidak muncul di form    │
│    counting baru                      │
│ 3. Data historis tetap valid          │
│ 4. Bisa diaktifkan kembali           │
└───────────────────────────────────────┘
```

**Aturan:**
- Hanya **Admin** yang bisa manage denominasi master
- Denominasi baru langsung tersedia di semua form counting setelah dibuat
- Deactivation tidak mempengaruhi data historis — denominasi lama tetap tampil di laporan lama
- Tidak perlu migrasi data saat menambah/menonaktifkan denominasi
- Seed data (default denominasi Rupiah) di-insert saat tenant onboarding

### 6. Cash Count Verification

Sistem memverifikasi konsistensi antara total denominasi dan nominal yang diharapkan:

```
Verification Flow:
        │
        v
┌───────────────────────────────────────┐
│ 1. User input quantity per denominasi │
│ 2. Sistem hitung:                     │
│    subtotal = value * quantity         │
│    total = SUM(subtotal)              │
│ 3. Bandingkan total vs expected        │
│                                       │
│ Match     → proceed                   │
│ Not match → warning + recount option  │
└───────────────────────────────────────┘
```

**Context-specific verification:**

| Context | Expected Amount | Source |
|---------|----------------|--------|
| Teller Opening | Nominal kas awal yang diserahkan Supervisor | Manual input by Supervisor |
| Teller Closing | expected_cash dari running total session | System calculated |
| Vault Count | Saldo vault terakhir | System tracked |
| Cash Transfer | Nominal transfer yang disepakati | Manual input |

**Aturan:**
- Pada **teller opening**: SUM denominasi harus = `opening_amount` yang diinput — jika tidak match, recount
- Pada **teller closing**: SUM denominasi menjadi `actual_cash` — selisih dengan `expected_cash` ditangani sesuai [K012 §5](./ADR-K012-teller-session.md)
- Pada **vault count**: SUM denominasi dibandingkan dengan saldo vault terakhir — selisih membutuhkan investigation
- UI menampilkan **real-time calculation** saat user mengisi quantity — total update otomatis

### 7. Vault Management

Brankas cabang melacak total kas per denominasi:

```
Branch Vault:
┌──────────────────────────────────────────────┐
│ Vault: Cabang Jakarta Pusat                  │
│ Last Count: 15 April 2026, 17:30             │
│                                              │
│ ┌───────────────┬──────────┬──────────────┐  │
│ │ Denominasi    │ Stok     │ Subtotal     │  │
│ ├───────────────┼──────────┼──────────────┤  │
│ │ Rp 100.000    │ 200 lbr  │ 20.000.000   │  │
│ │ Rp  50.000    │ 100 lbr  │  5.000.000   │  │
│ │ Rp  20.000    │ 150 lbr  │  3.000.000   │  │
│ │ ... (lanjut)  │          │              │  │
│ ├───────────────┼──────────┼──────────────┤  │
│ │ TOTAL         │          │ 35.000.000   │  │
│ └───────────────┴──────────┴──────────────┘  │
│                                              │
│ Movement Today:                              │
│ - Serah ke Teller Ahmad:  -5.000.000         │
│ - Serah ke Teller Budi:   -5.000.000         │
│ - Terima dari Teller Ahmad: +10.000.000      │
│ - Terima dari Teller Budi:  +7.000.000       │
│ ─────────────────────────────────             │
│ Net movement: +7.000.000                     │
└──────────────────────────────────────────────┘
```

**Vault balance tracking:**
```
vault_balance = last_counted_total
             + SUM(cash_in from teller closings)
             - SUM(cash_out to teller openings)
             + SUM(cash_in from branch transfers)
             - SUM(cash_out to branch transfers)
```

**Aturan:**
- Vault **tidak punya tabel dedicated** untuk real-time tracking — vault balance dihitung dari movements
- Vault count (penghitungan fisik brankas) dilakukan **periodik** oleh Supervisor/Manager
- Vault count mencatat denominasi detail ke `vault_count` table
- Selisih antara vault count dan calculated balance membutuhkan investigation (sama seperti teller variance)
- Minimum vault cash **configurable per branch** — alert jika mendekati minimum (lihat [K014 §6](./ADR-K014-kas-cashflow.md))

### 8. Data Model — Vault Count

```
vault_count
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant) NOT NULL
├── branch_id             UUID (FK → branch) NOT NULL
│
├── ── Count Info ──
├── count_date            DATE NOT NULL
├── counted_by            UUID (FK → user) NOT NULL
├── verified_by           UUID (nullable, FK → user, Supervisor/Manager)
│
├── ── Result ──
├── total_counted         NUMERIC(15,2) NOT NULL (SUM denominasi)
├── total_expected        NUMERIC(15,2) NOT NULL (calculated vault balance)
├── variance              NUMERIC(15,2) NOT NULL (expected - counted)
├── variance_explanation  TEXT (nullable, wajib jika variance ≠ 0)
│
├── ── Status ──
├── status                ENUM (draft, verified, investigated) NOT NULL
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

### 9. Data Model — Vault Count Denomination

```
vault_count_denomination
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant) NOT NULL
├── vault_count_id        UUID (FK → vault_count) NOT NULL
│
├── ── Denomination ──
├── denomination_value    INTEGER NOT NULL
├── denomination_type     ENUM (banknote, coin) NOT NULL
├── quantity              INTEGER NOT NULL DEFAULT 0
├── subtotal              NUMERIC(15,2) NOT NULL
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
```

### 10. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View denominasi master | v | v | v | v |
| Manage denominasi master (add/deactivate) | - | - | - | v |
| Input denominasi (teller session) | v | v | - | - |
| Vault count — execute | - | v | v | v |
| Vault count — verify | - | - | v | v |
| View vault balance | - | v | v | v |
| View vault count history | - | v | v | v |

**Catatan:**
- Teller hanya input denominasi untuk **session sendiri** — tidak bisa akses vault
- Supervisor bisa melakukan vault count dan melihat balance vault
- Manager bisa verify vault count dan investigate variance
- Hanya Admin yang bisa manage denominasi master (menambah/menonaktifkan pecahan)

### 11. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Denominasi | Pecahan Uang | Pecahan Uang |
| Uang Kertas | Uang Kertas | Uang Kertas |
| Uang Logam | Uang Logam | Uang Logam |
| Vault | Brankas | Brankas |
| Cash Count | Penghitungan Kas | Penghitungan Kas |

Terminologi denominasi **identik** di kedua mode — tidak ada perbedaan syariah pada uang fisik.

## Consequences

### Positif

- **Granular tracking** — kas terlacak per denominasi, bukan hanya total nominal
- **Fraud detection** — komposisi denominasi bisa mengungkap manipulasi (misal: mengklaim jumlah sama dengan pecahan berbeda)
- **Audit ready** — laporan denominasi detail memenuhi kebutuhan audit koperasi dan OJK
- **Future-proof** — denominasi baru bisa ditambahkan tanpa perubahan schema
- **Vault visibility** — komposisi kas di brankas diketahui secara real-time
- **Consistency** — format denominasi seragam di semua usage points

### Negatif

- **Input overhead** — teller harus menghitung dan input per denominasi (bukan hanya total)
- **Storage** — setiap pencatatan denominasi menghasilkan 11+ records (satu per denominasi)
- **Master data management** — perlu maintain katalog denominasi per tenant

### Mitigasi

- UI form counting dioptimasi: **keypad entry** per denominasi dengan auto-calculate subtotal dan total
- Form bisa **pre-fill** denominasi yang sering digunakan (misal: skip logam < 500 jika jarang)
- Storage minimal — record denominasi sangat kecil (denominasi_value + quantity + subtotal)
- Denominasi master di-seed otomatis saat onboarding — Admin hanya perlu manage jika ada perubahan BI

## Alternatives Considered

### A. Total Amount Only (tanpa denominasi detail)

Hanya mencatat total nominal kas tanpa breakdown per denominasi.

**Ditolak** karena: tidak cukup untuk audit dan vault management. Total saja tidak bisa mendeteksi manipulasi komposisi kas. Standar operasional perbankan/koperasi mewajibkan pencatatan per denominasi untuk setiap serah terima kas.

### B. Global Denomination Master (bukan per tenant)

Satu katalog denominasi untuk semua tenant.

**Ditolak** karena: meskipun saat ini semua tenant menggunakan Rupiah, tenant-level master memberikan fleksibilitas untuk:
- Tenant yang ingin menonaktifkan denominasi tertentu (misal: logam kecil)
- Future support multi-currency jika diperlukan
- Independensi konfigurasi antar tenant

### C. Denomination sebagai JSONB (bukan tabel terpisah)

Menyimpan detail denominasi sebagai JSONB array di tabel teller_session.

**Ditolak** karena: JSONB tidak mendukung constraint dan aggregation yang efisien. Query seperti "total Rp 100.000 di semua session hari ini" membutuhkan JSON parsing yang lambat. Tabel terpisah memungkinkan indexing dan aggregation SQL standar.

### D. Denomination Template per Teller

Setiap teller punya template denominasi default yang bisa di-reuse.

**Dipertimbangkan untuk Phase 2** — bukan ditolak, tapi ditunda. Template bisa mempercepat input tapi menambah complexity. MVP cukup dengan form kosong dan auto-calculate. Jika feedback dari operasional menunjukkan kebutuhan, bisa ditambahkan.
