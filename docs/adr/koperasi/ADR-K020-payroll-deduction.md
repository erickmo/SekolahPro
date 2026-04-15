# ADR-K020: Payroll Integration & Auto-Deduction

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Guru, ustadz, dan staf sekolah yang menjadi anggota koperasi seringkali memiliki kewajiban finansial rutin:

- **Simpanan Wajib** — bulanan, wajib dibayar selama menjadi anggota ([ADR-K004](./ADR-K004-simpanan-pokok-wajib.md))
- **Angsuran Pinjaman** — cicilan bulanan ([ADR-K008](./ADR-K008-angsuran-jadwal.md))
- **Tabungan Berencana** — setoran rutin yang disepakati ([ADR-K005](./ADR-K005-tabungan.md))

Proses pembayaran manual setiap bulan tidak efisien dan rawan keterlambatan. Banyak koperasi sekolah di Indonesia menerapkan **potong gaji otomatis** melalui koordinasi dengan TU (Tata Usaha) atau bagian keuangan sekolah.

Sistem ini **tidak membangun payroll** — hanya mengonsumsi data gaji yang dikirim oleh sistem HR sekolah atau di-upload manual oleh TU, lalu mengeksekusi potongan yang sudah diotorisasi.

## Decision

### 1. Integration Model — Consumer, Bukan Producer

Koperasi menerima data payroll dari sistem HR sekolah atau upload manual. Tidak membangun payroll engine.

```
School HR System / TU
        │
        │ CSV upload / API push
        v
┌─────────────────────────────┐
│ Koperasi: Payroll Processor │
│ 1. Parse data payroll       │
│ 2. Match ke nasabah         │
│ 3. Calculate deductions     │
│ 4. Generate report          │
│ 5. Execute (after approval) │
└─────────────────────────────┘
```

**Aturan:**
- Sistem koperasi **tidak menyimpan data gaji secara permanen** — hanya memproses untuk kalkulasi potongan
- Data gaji di batch dihapus setelah periode retensi yang dikonfigurasi (default: 12 bulan)
- Integrasi bisa via CSV upload (manual) atau API push (otomatis) — configurable per tenant

### 2. Jenis Potongan (Deduction Types)

Potongan yang bisa dikonfigurasi per nasabah:

```
deduction_type:
├── SIMPANAN_WAJIB          Simpanan wajib bulanan (K004)
├── ANGSURAN_PINJAMAN       Angsuran/cicilan pinjaman (K008)
├── TABUNGAN_BERENCANA      Setoran rutin tabungan berencana (K005)
├── TABUNGAN_REGULER        Setoran rutin tabungan biasa (K005)
├── IURAN_ANGGOTA           Iuran keanggotaan tahunan (jika berlaku)
└── CUSTOM                  Potongan custom yang dikonfigurasi admin
```

### 3. Prioritas Potongan

Jika gaji bersih (net salary) tidak mencukupi untuk semua potongan, sistem mengikuti prioritas:

```
priority_config (configurable per tenant):
  1. ANGSURAN_PINJAMAN        ← Paling prioritas (kewajiban hukum)
  2. SIMPANAN_WAJIB           ← Kewajiban keanggotaan
  3. TABUNGAN_BERENCANA       ← Komitmen sukarela
  4. TABUNGAN_REGULER         ← Sukarela
  5. IURAN_ANGGOTA            ← Tahunan
  6. CUSTOM                   ← Terendah
```

**Aturan prioritas:**
- Urutan prioritas **configurable per tenant** — default di atas bisa diubah oleh Admin
- Sistem memproses potongan dari prioritas tertinggi ke terendah
- Jika sisa gaji tidak cukup untuk potongan prioritas berikutnya, potongan tersebut di-**skip** (bukan partial)
- **Opsi partial deduction** bisa diaktifkan per tenant — jika aktif, sisa gaji dipotong sebagian untuk prioritas berikutnya
- Setiap potongan yang di-skip dicatat dengan alasan `INSUFFICIENT_SALARY`

### 4. Process Flow

```
STEP 1: Data Payroll Diterima
        │
        │ CSV upload oleh TU / API push dari HR system
        v
┌─────────────────────────────┐
│ payroll_batch                │
│ Status: UPLOADED             │
│ period: 2026-04 (April)     │
│ total_employees: 85          │
└────────┬────────────────────┘
         │ auto-process
         v
STEP 2: Matching & Calculation
┌─────────────────────────────┐
│ Untuk setiap employee:       │
│ 1. Match ke nasabah          │
│    (identity_number/emp_id)  │
│ 2. Cek otorisasi aktif       │
│ 3. Hitung semua potongan     │
│ 4. Apply priority jika perlu │
│ Status: CALCULATED           │
└────────┬────────────────────┘
         │
         v
STEP 3: Deduction Report
┌─────────────────────────────┐
│ Generate report untuk review:│
│ - Per employee: daftar potong│
│ - Total per jenis potongan   │
│ - Skipped deductions         │
│ - Unmatched employees        │
│ Status: PENDING_APPROVAL     │
└────────┬────────────────────┘
         │ Manager approve
         v
STEP 4: Execution
┌─────────────────────────────┐
│ Untuk setiap potongan:       │
│ 1. Create transaksi (K011)   │
│ 2. Update saldo rekening     │
│ 3. Update jadwal angsuran    │
│ Status: EXECUTED             │
└────────┬────────────────────┘
         │
         v
STEP 5: Notification
┌─────────────────────────────┐
│ 1. Notifikasi ke nasabah     │
│    (detail potongan)         │
│ 2. Konfirmasi ke TU/HR      │
│    (summary batch)           │
│ Status: COMPLETED            │
└─────────────────────────────┘
```

### 5. Payroll Data Format

Format import yang configurable:

```
payroll_import_config (per tenant):
  format: "csv" | "xlsx" | "api"
  delimiter: ","                    # untuk CSV
  encoding: "UTF-8"
  column_mapping:
    employee_id: "NIP"             # Kolom di file → field di sistem
    employee_name: "Nama"
    identity_number: "NIK"
    gross_salary: "Gaji Bruto"
    net_salary: "Gaji Bersih"
    department: "Unit"              # Opsional
  header_row: 1                    # Baris header
  data_start_row: 2               # Baris data mulai
```

**Contoh CSV:**
```csv
NIP,Nama,NIK,Gaji Bruto,Gaji Bersih,Unit
198501,Budi Santoso,3201...,8000000,6500000,TU
198502,Siti Rahmah,3202...,7500000,6000000,Guru
```

**Aturan:**
- Column mapping **configurable per tenant** — setiap sekolah mungkin punya format berbeda
- Validasi saat upload: semua required columns harus ada, format angka valid
- Preview data sebelum proses — TU bisa review dan cancel jika ada kesalahan

### 6. Employee-Nasabah Matching

Mencocokkan data karyawan dari payroll dengan nasabah koperasi:

```
matching_strategy:
  primary_key: "identity_number"     # NIK — paling reliable
  fallback_key: "employee_id"        # NIP — jika NIK tidak ada
  auto_create: false                 # Jangan auto-create nasabah dari payroll
```

**Matching flow:**
```
Untuk setiap employee di payroll:
        │
        v
┌─────────────────────────────┐
│ Match by identity_number     │
│ (employee.NIK == nasabah.    │
│  identity_number)            │
└────────┬────────────────────┘
         │
    ┌────┴────┐
    v         v
  FOUND     NOT FOUND
    │         │
    │    ┌────┴────┐
    │    v         v
    │  Match by   NOT FOUND
    │  employee_id    │
    │    │             v
    │    v         UNMATCHED
    │  FOUND      (skip, report)
    │    │
    v    v
  MATCHED → process deductions
```

**Aturan:**
- Employee yang tidak match dilaporkan di deduction report sebagai **UNMATCHED** — bukan error, tapi informasi
- Nasabah yang sudah match bisa disimpan di mapping table untuk mempercepat proses bulan berikutnya
- Mapping bisa dibuat manual oleh TU/Supervisor untuk kasus yang tidak bisa auto-match

**Mapping table:**

```
payroll_mapping
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── employee_id           VARCHAR NOT NULL (NIP dari payroll)
├── nasabah_id            UUID (FK → nasabah)
├── identity_number       VARCHAR (nullable, NIK)
├── is_verified           BOOLEAN DEFAULT false
├── verified_by           UUID (nullable)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

### 7. Otorisasi Potongan (Authorization)

Nasabah harus memberikan persetujuan tertulis untuk setiap jenis potongan:

```
payroll_authorization
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Jenis Otorisasi ──
├── deduction_type        ENUM (sesuai Section 2)
├── rekening_id           UUID (FK → rekening, target setoran)
├── amount                NUMERIC(15,2) NOT NULL (nominal potongan)
│
├── ── Periode ──
├── effective_date        DATE NOT NULL (mulai berlaku)
├── end_date              DATE (nullable, berlaku sampai dicabut jika null)
│
├── ── Status ──
├── status                ENUM (active, revoked, expired)
├── revoked_at            TIMESTAMPTZ (nullable)
├── revoked_by            UUID (nullable)
├── revoke_reason         TEXT (nullable)
│
├── ── Dokumen ──
├── authorization_doc_url VARCHAR (nullable, scan surat kuasa)
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

**Aturan otorisasi:**
- Nasabah **wajib** menandatangani surat kuasa untuk setiap jenis potongan
- Scan surat kuasa disimpan di `authorization_doc_url`
- Nasabah bisa **mencabut (revoke)** otorisasi kapan saja — efektif mulai bulan berikutnya
- Revoke simpanan wajib memerlukan persetujuan Manager (karena simpanan wajib bersifat mandatory)
- Otorisasi angsuran pinjaman **tidak bisa dicabut** selama pinjaman masih aktif — ini kewajiban hukum
- Sistem hanya memproses potongan yang memiliki otorisasi aktif

### 8. Payroll Batch Data Model

```
payroll_batch
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
│
├── ── Identitas ──
├── batch_number          VARCHAR UNIQUE per tenant
├── period_month          INTEGER NOT NULL (1-12)
├── period_year           INTEGER NOT NULL
├── upload_source         ENUM (csv, xlsx, api)
│
├── ── Statistik ──
├── total_employees       INTEGER NOT NULL
├── matched_employees     INTEGER DEFAULT 0
├── unmatched_employees   INTEGER DEFAULT 0
├── total_deduction_amount NUMERIC(15,2) DEFAULT 0
├── total_deductions      INTEGER DEFAULT 0
├── skipped_deductions    INTEGER DEFAULT 0
│
├── ── Workflow ──
├── status                ENUM (uploaded, calculated, pending_approval, approved, executing, executed, completed, failed)
├── calculated_at         TIMESTAMPTZ (nullable)
├── approved_at           TIMESTAMPTZ (nullable)
├── approved_by           UUID (nullable)
├── executed_at           TIMESTAMPTZ (nullable)
├── completed_at          TIMESTAMPTZ (nullable)
├── error_message         TEXT (nullable, jika failed)
│
├── ── Raw Data ──
├── raw_data_url          VARCHAR (nullable, URL ke file upload asli)
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

**Deduction per employee:**

```
payroll_deduction
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── batch_id              UUID (FK → payroll_batch)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Payroll Data ──
├── employee_id           VARCHAR NOT NULL (dari file payroll)
├── employee_name         VARCHAR NOT NULL
├── net_salary            NUMERIC(15,2) NOT NULL
│
├── ── Deduction ──
├── deduction_type        ENUM (sesuai Section 2)
├── authorization_id      UUID (FK → payroll_authorization)
├── rekening_id           UUID (FK → rekening, target)
├── amount                NUMERIC(15,2) NOT NULL
├── priority_order        INTEGER NOT NULL
│
├── ── Status ──
├── status                ENUM (calculated, approved, executed, skipped, failed)
├── skip_reason           TEXT (nullable, misal: INSUFFICIENT_SALARY)
├── transaction_id        UUID (nullable, FK → transaksi K011, setelah executed)
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

### 9. Failed Deduction Handling

Ketika gaji tidak mencukupi untuk semua potongan:

```
Contoh:
  Net salary:  Rp 3.000.000
  Angsuran:    Rp 2.500.000 (priority 1) → EXECUTED (sisa: 500.000)
  Simp. Wajib: Rp 200.000   (priority 2) → EXECUTED (sisa: 300.000)
  Tab. Rencana: Rp 500.000  (priority 3) → SKIPPED (sisa < amount)
```

**Aturan:**
- Proses dari prioritas tertinggi ke terendah
- Jika sisa tidak cukup dan `partial_deduction = false` (default) → **skip** seluruhnya
- Jika `partial_deduction = true` → potong sebanyak sisa, catat sebagai partial
- Semua potongan yang di-skip masuk ke deduction report dengan alasan
- Notifikasi ke nasabah untuk potongan yang di-skip ([ADR-K022](./ADR-K022-notifikasi.md))
- Potongan yang di-skip **tidak otomatis diakumulasi** ke bulan berikutnya — ini keputusan manual Manager

### 10. Reporting

| Laporan | Deskripsi | Audience |
|---------|-----------|----------|
| Deduction Report per Batch | Detail potongan per employee per bulan | Manager, TU |
| Summary per Deduction Type | Total per jenis potongan | Manager |
| Unmatched Employees | Karyawan yang tidak match ke nasabah | TU, Supervisor |
| Skipped Deductions | Potongan yang gagal/di-skip | Manager, Nasabah |
| Monthly Deduction History | Riwayat potongan per nasabah | Nasabah (via K023) |
| Annual Summary | Ringkasan potongan setahun per nasabah | Nasabah, Manager |

### 11. Vernon _rels dan _data Structure

**payroll_batch _rels/_data:**

```json
// _rels
{
  "tenant_id": "018f...",
  "branch_id": "018f..."
}

// _data
{
  "branch": {
    "id": "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  }
}
```

**payroll_deduction _rels/_data:**

```json
// _rels
{
  "tenant_id": "018f...",
  "batch_id": "018f...",
  "nasabah_id": "018f...",
  "rekening_id": "018f...",
  "authorization_id": "018f..."
}

// _data
{
  "nasabah": {
    "id": "018f...",
    "full_name": "Budi Santoso",
    "member_number": "KOP-2026-JKT-000042"
  },
  "rekening": {
    "id": "018f...",
    "account_number": "SW-2026-JKT-00000042",
    "category": "simpanan_wajib"
  },
  "batch": {
    "id": "018f...",
    "batch_number": "PAY-2026-04-001",
    "period_month": 4,
    "period_year": 2026
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.nasabah` di semua deduction nasabah tersebut
- `RekeningUpdatedEvent` → update `_data.rekening` di semua deduction yang merujuk rekening tersebut

### 12. Authorization — RBAC

| Permission | TU/Operator | Supervisor | Manager | Admin |
|------------|-------------|------------|---------|-------|
| Upload data payroll | v | v | v | v |
| Configure import format | - | v | v | v |
| View deduction report | v | v | v | v |
| Approve deduction batch | - | - | v | v |
| Execute deduction batch | - | - | - | System |
| Create nasabah-employee mapping | v | v | v | v |
| Verify mapping | - | v | v | v |
| Manage authorization | - | v | v | v |
| Revoke authorization (simpanan wajib) | - | - | v | v |
| Configure priority order | - | - | - | v |
| Configure import template | - | - | v | v |
| View batch history | v | v | v | v |
| Retry failed batch | - | - | v | v |

**Catatan:**
- TU/Operator bisa upload data tapi **tidak bisa approve** — separation of duties
- Execution dilakukan oleh **System** setelah Manager approve — otomatis, bukan manual
- Revoke otorisasi simpanan wajib butuh Manager karena bersifat mandatory

### 13. Idempotency & Safety

Batch processing harus idempotent dan aman:

**Aturan:**
- Satu batch per `period_month + period_year + branch_id` — tidak bisa double-execute untuk periode yang sama
- Jika batch gagal di tengah jalan, bisa di-retry — sistem cek mana yang sudah executed dan skip
- Setiap deduction yang sudah executed punya `transaction_id` — link ke transaksi K011
- Rollback seluruh batch dimungkinkan oleh Admin — reverse semua transaksi yang sudah dibuat (harus dalam 24 jam)

### 14. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Module name | Potongan Gaji | Potongan Gaji |
| Deduction report | Laporan Potongan Gaji | Laporan Potongan Gaji |
| Loan installment label | Angsuran Pinjaman | Angsuran Pembiayaan |

Proses potong gaji bersifat **operasional** — sama di kedua mode. Perbedaan hanya pada label jenis potongan yang mengikuti terminologi masing-masing mode.

## Consequences

### Positif

- **Efisiensi** — potongan otomatis mengurangi kerja manual TU dan teller
- **Kepatuhan** — simpanan wajib dan angsuran pasti terpotong (selama gaji cukup)
- **Transparansi** — nasabah mendapat notifikasi detail potongan setiap bulan
- **Audit trail** — setiap batch, deduction, dan authorization tercatat lengkap
- **Flexible** — format import, prioritas, dan jenis potongan configurable per tenant
- **Safe** — approval flow mencegah potongan salah, idempotency mencegah double-deduction

### Negatif

- **Dependency pada HR** — jika data payroll telat atau salah, proses terganggu
- **Matching complexity** — tidak semua karyawan bisa auto-match ke nasabah
- **Privacy concern** — sistem koperasi menerima data gaji (meskipun tidak menyimpan permanen)
- **Batch execution time** — untuk koperasi besar (ratusan karyawan), execution bisa memakan waktu

### Mitigasi

- Deadline upload payroll bisa di-set di tenant config — reminder otomatis ke TU jika belum upload
- Manual mapping sebagai fallback untuk karyawan yang tidak bisa auto-match
- Data gaji dihapus setelah retensi — retention period configurable
- Batch execution menggunakan background job dengan progress tracking — tidak blocking UI

## Alternatives Considered

### A. Direct Payroll Integration (Build Payroll Module)

Membangun modul payroll lengkap di dalam sistem koperasi.

**Ditolak** karena: payroll adalah domain yang sangat kompleks (pajak, BPJS, tunjangan, lembur) dan sudah banyak solusi dedicated. Koperasi hanya butuh data net salary untuk kalkulasi potongan, bukan seluruh payroll engine.

### B. Manual Entry per Employee per Month

TU input potongan satu per satu untuk setiap karyawan setiap bulan.

**Ditolak** karena: tidak scalable untuk koperasi dengan puluhan-ratusan karyawan. Error-prone dan memakan waktu. Batch processing jauh lebih efisien.

### C. Auto-Accumulate Skipped Deductions

Potongan yang di-skip otomatis diakumulasi ke bulan berikutnya.

**Ditolak** karena: bisa menyebabkan gaji bulan berikutnya terpotong terlalu banyak tanpa persetujuan nasabah. Keputusan akumulasi harus manual oleh Manager dengan persetujuan nasabah.
