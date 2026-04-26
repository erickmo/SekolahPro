# ADR-S057: Scholarship Management (Manajemen Beasiswa)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Beasiswa adalah komponen penting di sekolah Indonesia, baik dari pemerintah maupun swasta. Program beasiswa utama:

1. **PIP (Program Indonesia Pintar)**: Beasiswa dari pemerintah untuk siswa tidak mampu — dicairkan via rekening siswa/orang tua.
2. **KIP (Kartu Indonesia Pintar)**: Kartu identitas penerima PIP — siswa yang punya KIP otomatis eligible.
3. **BOS Siswa**: Sebagian dana BOS dialokasikan untuk beasiswa siswa berprestasi atau kurang mampu.
4. **Yayasan**: Beasiswa internal dari yayasan penyelenggara sekolah — sering untuk anak yatim, hafidz, atau berprestasi.
5. **Donatur external**: Beasiswa dari alumni (S056), perusahaan CSR, atau lembaga filantropi.

Kebutuhan manajemen beasiswa:

- **Application workflow**: Siswa/orang tua mengajukan → TU verifikasi kelengkapan → kepsek approve.
- **Selection criteria**: Nilai akademik (S011), kehadiran (S008), kondisi ekonomi, prestasi (S013).
- **Finance integration**: Beasiswa mengurangi tagihan SPP (S009) — bisa partial atau full.
- **Disbursement tracking**: Berapa yang dicairkan, kapan, ke siapa, metode apa.
- **Renewal**: Beasiswa bisa per semester atau per tahun — perlu renewal management.
- **Reporting**: Laporan penerima beasiswa untuk Dapodik (S055), akreditasi, dan donatur.

### Mengapa Vernon Pattern?

- Read-heavy: admin, kepsek, dan donatur sering melihat daftar penerima, status, dan laporan beasiswa.
- Relasi ke student, academic_year, fee_type (S009).
- Write periodik: application dan disbursement terjadi per semester.
- Business logic moderate: eligibility check, workflow approval, SPP offset calculation.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `scholarship_programs` (program beasiswa), `scholarship_applications` (pengajuan), dan `scholarship_disbursements` (pencairan).

### Table Schema

```sql
-- Program beasiswa
CREATE TABLE scholarship_programs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(200) NOT NULL,
    code            VARCHAR(30) NOT NULL,
    description     TEXT,
    scholarship_type VARCHAR(30) NOT NULL,

    -- Sumber dana
    funding_source  VARCHAR(30) NOT NULL,
    donor_name      VARCHAR(255),
    donor_contact   VARCHAR(255),

    -- Konfigurasi
    amount_per_period BIGINT NOT NULL,
    period_type     VARCHAR(20) NOT NULL,
    max_recipients  INT,
    is_renewable    BOOLEAN NOT NULL DEFAULT true,
    max_renewal     INT,

    -- Kriteria seleksi
    min_gpa         NUMERIC(4,2),
    min_attendance_rate NUMERIC(5,2),
    requires_financial_need BOOLEAN NOT NULL DEFAULT false,
    additional_criteria JSONB NOT NULL DEFAULT '{}',

    -- Periode aktif
    academic_year_id UUID NOT NULL,
    start_date      DATE NOT NULL,
    end_date        DATE,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    current_recipients INT NOT NULL DEFAULT 0,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_scholarship_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_scholarship_type CHECK (scholarship_type IN ('pip', 'kip', 'bos_siswa', 'yayasan', 'donatur_external', 'prestasi', 'hafidz', 'yatim', 'other')),
    CONSTRAINT chk_funding_source CHECK (funding_source IN ('pemerintah_pusat', 'pemerintah_daerah', 'bos', 'yayasan', 'alumni', 'perusahaan', 'lembaga', 'perorangan', 'other')),
    CONSTRAINT chk_period_type CHECK (period_type IN ('monthly', 'semester', 'annual', 'one_time')),
    CONSTRAINT chk_program_status CHECK (status IN ('active', 'inactive', 'closed', 'suspended')),
    CONSTRAINT chk_amount_positive CHECK (amount_per_period > 0)
);

CREATE INDEX idx_scholarship_program_tenant ON scholarship_programs (tenant_id, company_id);
CREATE INDEX idx_scholarship_program_type ON scholarship_programs (scholarship_type);
CREATE INDEX idx_scholarship_program_year ON scholarship_programs (academic_year_id);
CREATE INDEX idx_scholarship_program_status ON scholarship_programs (status) WHERE status = 'active';
CREATE INDEX idx_scholarship_program_rels ON scholarship_programs USING GIN (_rels);
CREATE INDEX idx_scholarship_program_data ON scholarship_programs USING GIN (_data);

-- Pengajuan beasiswa
CREATE TABLE scholarship_applications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    program_id      UUID NOT NULL,
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Identitas pengajuan
    application_no  VARCHAR(30) NOT NULL,
    semester        INT NOT NULL,

    -- Data pendukung
    gpa             NUMERIC(4,2),
    attendance_rate NUMERIC(5,2),
    financial_need_proof TEXT,
    supporting_docs JSONB NOT NULL DEFAULT '[]',
    applicant_note  TEXT,

    -- KIP/PIP specific
    kip_number      VARCHAR(20),
    pip_recipient_id VARCHAR(30),

    -- Workflow
    status          VARCHAR(20) NOT NULL DEFAULT 'submitted',
    verified_by     UUID,
    verified_at     TIMESTAMPTZ,
    verification_note TEXT,
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,
    approval_note   TEXT,
    rejected_reason TEXT,

    -- Approval details
    approved_amount BIGINT,
    approved_period_start DATE,
    approved_period_end DATE,

    -- Renewal
    is_renewal      BOOLEAN NOT NULL DEFAULT false,
    previous_application_id UUID,
    renewal_count   INT NOT NULL DEFAULT 0,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_scholarship_application UNIQUE (tenant_id, company_id, application_no),
    CONSTRAINT uq_scholarship_student_program_sem UNIQUE (program_id, student_id, academic_year_id, semester),
    CONSTRAINT chk_application_semester CHECK (semester IN (1, 2)),
    CONSTRAINT chk_application_status CHECK (status IN ('submitted', 'verified', 'approved', 'rejected', 'withdrawn', 'expired', 'renewed')),
    CONSTRAINT chk_approved_amount CHECK (approved_amount IS NULL OR approved_amount > 0)
);

CREATE INDEX idx_scholarship_app_tenant ON scholarship_applications (tenant_id, company_id);
CREATE INDEX idx_scholarship_app_program ON scholarship_applications (program_id);
CREATE INDEX idx_scholarship_app_student ON scholarship_applications (student_id);
CREATE INDEX idx_scholarship_app_status ON scholarship_applications (status) WHERE status IN ('submitted', 'verified');
CREATE INDEX idx_scholarship_app_year ON scholarship_applications (academic_year_id, semester);
CREATE INDEX idx_scholarship_app_kip ON scholarship_applications (kip_number) WHERE kip_number IS NOT NULL;
CREATE INDEX idx_scholarship_app_rels ON scholarship_applications USING GIN (_rels);
CREATE INDEX idx_scholarship_app_data ON scholarship_applications USING GIN (_data);

-- Pencairan beasiswa
CREATE TABLE scholarship_disbursements (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    application_id  UUID NOT NULL,
    student_id      UUID NOT NULL,
    program_id      UUID NOT NULL,

    -- Invoice offset (integrasi S009)
    invoice_id      UUID,

    -- Detail pencairan
    disbursement_no VARCHAR(30) NOT NULL,
    amount          BIGINT NOT NULL,
    period_month    INT,
    period_year     INT NOT NULL,

    -- Metode pencairan
    disbursement_method VARCHAR(20) NOT NULL,
    bank_name       VARCHAR(100),
    account_number  VARCHAR(30),
    account_holder  VARCHAR(255),

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    disbursed_at    TIMESTAMPTZ,
    disbursed_by    UUID,
    receipt_url     TEXT,
    note            TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_disbursement_no UNIQUE (tenant_id, company_id, disbursement_no),
    CONSTRAINT uq_disbursement_period UNIQUE (application_id, period_year, period_month),
    CONSTRAINT chk_disbursement_method CHECK (disbursement_method IN ('spp_offset', 'bank_transfer', 'cash', 'ewallet', 'other')),
    CONSTRAINT chk_disbursement_status CHECK (status IN ('pending', 'processing', 'disbursed', 'failed', 'returned')),
    CONSTRAINT chk_disbursement_amount CHECK (amount > 0),
    CONSTRAINT chk_disbursement_month CHECK (period_month IS NULL OR (period_month >= 1 AND period_month <= 12))
);

CREATE INDEX idx_disbursement_tenant ON scholarship_disbursements (tenant_id, company_id);
CREATE INDEX idx_disbursement_application ON scholarship_disbursements (application_id);
CREATE INDEX idx_disbursement_student ON scholarship_disbursements (student_id);
CREATE INDEX idx_disbursement_program ON scholarship_disbursements (program_id);
CREATE INDEX idx_disbursement_invoice ON scholarship_disbursements (invoice_id) WHERE invoice_id IS NOT NULL;
CREATE INDEX idx_disbursement_status ON scholarship_disbursements (status) WHERE status IN ('pending', 'processing');
CREATE INDEX idx_disbursement_period ON scholarship_disbursements (period_year, period_month);
CREATE INDEX idx_disbursement_rels ON scholarship_disbursements USING GIN (_rels);
CREATE INDEX idx_disbursement_data ON scholarship_disbursements USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `scholarship_type` | VARCHAR(30), 9 types | Mencakup semua jenis beasiswa umum di Indonesia: PIP, KIP, BOS, yayasan, donatur, prestasi, hafidz, yatim |
| `funding_source` | VARCHAR(30) | Sumber dana terpisah dari tipe — PIP sumbernya pemerintah pusat, yayasan sumbernya yayasan |
| `amount_per_period` | BIGINT | Rupiah per periode (bulan/semester/tahun) — konsisten dengan S009 |
| `period_type` | VARCHAR(20) | Fleksibel: bulanan (SPP offset), semester (PIP), tahunan, atau one-time |
| `min_gpa` / `min_attendance_rate` | NUMERIC | Kriteria otomatis — di-check dari S011 (grade) dan S008 (attendance) |
| `kip_number` | VARCHAR(20) | Nomor KIP — identitas penerima PIP dari pemerintah |
| `pip_recipient_id` | VARCHAR(30) | ID penerima PIP dari database PIP — untuk reconciliation |
| `is_renewal` / `previous_application_id` | BOOLEAN + UUID | Tracking renewal chain — beasiswa semester ini continuation dari semester lalu |
| `disbursement_method` | VARCHAR(20), 5 metode | `spp_offset` = langsung kurangi tagihan SPP (paling umum), bank_transfer untuk PIP |
| `invoice_id` | UUID, nullable | Jika metode = spp_offset, link ke invoice S009 yang di-discount |

### Scholarship → SPP Integration Flow

```
┌─────────────────────────────────────────────────────────────┐
│ Scholarship ↔ SPP (S009) Integration                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ 1. Beasiswa approved (status = 'approved')                   │
│    approved_amount = Rp 500.000/bulan                       │
│                                                             │
│ 2. Setiap bulan, system generate disbursement:              │
│    disbursement_method = 'spp_offset'                       │
│                                                             │
│ 3. System find invoice SPP bulan ini (S009):                │
│    student_invoices WHERE student_id + period_month          │
│                                                             │
│ 4. Apply offset:                                            │
│    invoice.discount += disbursement.amount                  │
│    invoice.total_amount = invoice.amount - invoice.discount │
│                                                             │
│ 5. Disbursement linked:                                     │
│    disbursement.invoice_id = invoice.id                     │
│    disbursement.status = 'disbursed'                        │
│                                                             │
│ 6. Jika beasiswa full (amount >= SPP):                      │
│    invoice.status = 'waived' atau 'paid'                    │
│                                                             │
│ Contoh:                                                     │
│ SPP = Rp 800.000                                            │
│ Beasiswa yayasan = Rp 500.000                               │
│ Orang tua bayar = Rp 300.000                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Vernon Relationships

**scholarship_programs:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran program |

**scholarship_applications:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `program` | belongs_to | **Ya** | Nama program, tipe, dan amount |
| `student` | belongs_to | **Ya** | Nama dan data siswa penerima |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

**scholarship_disbursements:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `application` | belongs_to | **Ya** | Konteks pengajuan yang di-cairkan |
| `student` | belongs_to | **Ya** | Penerima pencairan |
| `program` | belongs_to | **Ya** | Program beasiswa |
| `invoice` | belongs_to | Tidak | Opsional — hanya untuk spp_offset method |

### _rels / _data Structure

```json
// scholarship_programs
{
  "_rels": {
    "academic_year_id": "018f..."
  },
  "_data": {
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}

// scholarship_applications
{
  "_rels": {
    "program_id": "018f...",
    "student_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "program": { "id": "018f...", "name": "Beasiswa Yayasan Al-Hikmah", "scholarship_type": "yayasan", "amount_per_period": 500000 },
    "student": { "id": "018f...", "full_name": "Ahmad Fauzi", "nis": "12345", "class_name": "VII-A" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}

// scholarship_disbursements
{
  "_rels": {
    "application_id": "018f...",
    "student_id": "018f...",
    "program_id": "018f...",
    "invoice_id": "018f..."
  },
  "_data": {
    "application": { "id": "018f...", "application_no": "BSW-2026-001", "status": "approved" },
    "student": { "id": "018f...", "full_name": "Ahmad Fauzi", "nis": "12345" },
    "program": { "id": "018f...", "name": "Beasiswa Yayasan Al-Hikmah", "scholarship_type": "yayasan" },
    "invoice": { "id": "018f...", "invoice_no": "INV-2026-001", "total_amount": 800000 }
  }
}
```

### API Endpoints

```
# Scholarship Programs
GET    /api/v1/scholarship-programs                      — List program beasiswa
POST   /api/v1/scholarship-programs                      — Buat program baru
PUT    /api/v1/scholarship-programs/{id}                 — Update program
GET    /api/v1/scholarship-programs/{id}                 — Detail program + stats

# Applications
GET    /api/v1/scholarship-applications                  — List semua pengajuan (admin)
POST   /api/v1/scholarship-applications                  — Ajukan beasiswa (siswa/orang tua)
GET    /api/v1/scholarship-applications/{id}             — Detail pengajuan
PUT    /api/v1/scholarship-applications/{id}/verify      — Verifikasi oleh TU
PUT    /api/v1/scholarship-applications/{id}/approve     — Approve oleh kepsek
PUT    /api/v1/scholarship-applications/{id}/reject      — Reject dengan alasan
POST   /api/v1/scholarship-applications/{id}/renew       — Perpanjang beasiswa
GET    /api/v1/students/{id}/scholarship-applications    — Pengajuan per siswa

# Eligibility Check
GET    /api/v1/scholarship-programs/{id}/check-eligibility/{student_id}  — Cek eligibility siswa

# Disbursements
GET    /api/v1/scholarship-disbursements                 — List pencairan
POST   /api/v1/scholarship-disbursements                 — Catat pencairan
POST   /api/v1/scholarship-disbursements/generate-batch  — Generate pencairan bulk per bulan
GET    /api/v1/scholarship-disbursements/{id}            — Detail pencairan
GET    /api/v1/students/{id}/scholarship-disbursements   — Pencairan per siswa

# Reports
GET    /api/v1/scholarship/summary                       — Rekap beasiswa per program
GET    /api/v1/scholarship/recipients                    — Daftar penerima aktif
GET    /api/v1/scholarship/disbursement-report           — Laporan pencairan per periode
```

## Consequences

### Positive

- **Indonesia-specific**: Mendukung PIP, KIP, BOS — program beasiswa standar pemerintah Indonesia.
- **SPP integration**: Beasiswa langsung offset tagihan SPP (S009) — mengurangi manual accounting.
- **Workflow approval**: Pengajuan → verifikasi → approval memastikan proper governance.
- **Eligibility automation**: Cek GPA dan kehadiran otomatis dari S011 dan S008 — mengurangi subjektivitas.
- **Renewal tracking**: Chain of applications memudahkan tracking history beasiswa per siswa.
- **Disbursement audit**: Setiap pencairan tercatat dengan metode, tanggal, dan penerima.

### Negative / Trade-offs

- **SPP coupling**: Integrasi erat dengan S009 — perubahan di domain keuangan bisa impact beasiswa.
- **Complex eligibility**: Kriteria seleksi yang configurable per program menambah complexity.
- **PIP/KIP external data**: Data penerima PIP/KIP dari pemerintah perlu di-reconcile manual — tidak ada API resmi.
- **Partial offset calculation**: Jika siswa punya multiple beasiswa, perhitungan offset SPP bisa complex.
- **Renewal decision**: Keputusan renewal masih manual — belum ada auto-renewal berdasarkan performa.

## Alternatives Considered

### 1. Beasiswa sebagai discount di fee_types (S009)
- Ditolak: beasiswa punya lifecycle sendiri (application, approval, renewal) yang tidak fit di domain keuangan murni.

### 2. Satu tabel scholarship_recipients tanpa application workflow
- Ditolak: tanpa workflow, tidak ada audit trail pengajuan dan approval — penting untuk governance dan transparansi.

### 3. External scholarship management system
- Ditolak: beasiswa erat dengan data internal sekolah (nilai, kehadiran, keuangan) — integrasi external terlalu fragile.

### 4. Beasiswa hanya sebagai JSONB di students
- Ditolak: siswa bisa punya multiple beasiswa dengan lifecycle berbeda — perlu tabel proper untuk tracking.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `ProgramDescriptor.TableName()` | — | `"scholarship_programs"` |
| U02 | `ApplicationDescriptor.TableName()` | — | `"scholarship_applications"` |
| U03 | `DisbursementDescriptor.TableName()` | — | `"scholarship_disbursements"` |
| U04 | Validate rejects invalid `scholarship_type` | `"grant"` | Error: invalid scholarship_type |
| U05 | Validate rejects invalid `funding_source` | `"government"` | Error: invalid funding_source |
| U06 | Validate rejects `amount <= 0` | `amount_per_period = 0` | Error: must be positive |
| U07 | Validate rejects invalid `period_type` | `"weekly"` | Error: invalid period_type |
| U08 | Validate rejects invalid application `status` | `"pending"` | Error: invalid status |
| U09 | Validate rejects invalid `disbursement_method` | `"bitcoin"` | Error: invalid method |
| U10 | Validate rejects semester > 2 | `semester = 3` | Error: must be 1 or 2 |
| U11 | Eligibility check — GPA below minimum | GPA=3.0, min_gpa=3.5 | Not eligible |
| U12 | Eligibility check — all criteria met | GPA=3.8, attendance=95% | Eligible |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create scholarship program | POST with type=yayasan | 201, created |
| I02 | Unique code per company | Create 2 programs with same code | 409/422, unique constraint |
| I03 | Submit application | POST application for student | 201, status = 'submitted' |
| I04 | Duplicate application rejected | Same student+program+semester | 409/422, unique constraint |
| I05 | Verify application | PUT /verify by TU | 200, status → 'verified' |
| I06 | Approve application | PUT /approve by kepsek | 200, status → 'approved', program.current_recipients++ |
| I07 | Reject application | PUT /reject with reason | 200, status → 'rejected', reason stored |
| I08 | Check eligibility — eligible | Student with GPA > min | 200, eligible = true |
| I09 | Check eligibility — not eligible | Student with GPA < min | 200, eligible = false, reasons listed |
| I10 | Generate monthly disbursement | POST /generate-batch for month | Disbursements created for all active recipients |
| I11 | SPP offset disbursement | Disburse with method=spp_offset | Invoice discount updated, disbursement.invoice_id set |
| I12 | Bank transfer disbursement | Disburse with method=bank_transfer | No invoice link, bank details recorded |
| I13 | Renew application | POST /renew | New application with is_renewal=true, previous_application_id set |
| I14 | Max recipients exceeded | Approve beyond max_recipients | 422, quota full |
| I15 | Scholarship type CHECK enforced | INSERT with `scholarship_type = 'grant'` | DB error |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | StudentUpdated syncs to applications | Update student name | `_data.student.full_name` updated |
| I17 | ProgramUpdated syncs to applications | Update program name | `_data.program.name` updated |
| I18 | ApplicationUpdated syncs to disbursements | Update application status | `_data.application.status` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I19 | Cannot access other tenant's programs | GET with wrong tenant | 404 |
| I20 | Cannot apply cross-tenant | POST application for other tenant's student | Error |
| I21 | Disbursement tenant-isolated | List disbursements from other tenant | Empty result |
