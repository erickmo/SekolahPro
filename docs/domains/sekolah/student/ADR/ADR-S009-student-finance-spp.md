# ADR-S009: Student Finance / SPP

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Pembayaran SPP (Sumbangan Pembinaan Pendidikan) dan biaya sekolah lainnya adalah tulang punggung operasional keuangan sekolah swasta. Data keuangan siswa diperlukan untuk:

1. **Tagihan bulanan**: SPP, uang makan, uang kegiatan — recurring per bulan.
2. **Tagihan tahunan**: Uang pangkal, seragam, buku — one-time per tahun ajaran.
3. **Pelacakan pembayaran**: Status bayar/belum/cicilan per tagihan.
4. **Tunggakan**: Monitoring siswa yang menunggak dan follow-up.
5. **Laporan keuangan**: Rekap pemasukan per bulan, per kelas, per jenis tagihan.
6. **Dashboard orang tua**: Tampilan status pembayaran untuk orang tua (future).

Arsitektur keuangan terdiri dari 3 layer:

- **Fee Type**: Jenis tagihan (template) — SPP, uang makan, uang pangkal.
- **Invoice**: Tagihan yang di-generate per siswa per periode.
- **Payment**: Pembayaran yang dilakukan terhadap invoice.

### Mengapa Vernon Pattern?

- has_many dari student (banyak invoices per siswa).
- Read-heavy: dashboard keuangan, laporan tunggakan, cetak kwitansi.
- Relasi ke student, academic_year, fee_type.
- Business logic moderate: generate tagihan, record pembayaran, hitung sisa.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `fee_types` (master jenis tagihan), `student_invoices` (tagihan per siswa), dan `student_payments` (pembayaran).

### Table Schema

```sql
-- Master jenis tagihan
CREATE TABLE fee_types (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    description     TEXT,

    -- Konfigurasi
    fee_category    VARCHAR(20) NOT NULL,
    default_amount  BIGINT NOT NULL,
    is_recurring    BOOLEAN NOT NULL DEFAULT false,
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_fee_type_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_fee_category CHECK (fee_category IN ('monthly', 'annual', 'one_time', 'incidental'))
);

-- Tagihan per siswa
CREATE TABLE student_invoices (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    fee_type_id     UUID NOT NULL,

    -- Invoice detail
    invoice_no      VARCHAR(50) NOT NULL,
    period_month    INT,
    period_year     INT NOT NULL,
    amount          BIGINT NOT NULL,
    discount        BIGINT NOT NULL DEFAULT 0,
    total_amount    BIGINT NOT NULL,
    paid_amount     BIGINT NOT NULL DEFAULT 0,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'unpaid',
    due_date        DATE NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_invoice_no UNIQUE (tenant_id, company_id, invoice_no),
    CONSTRAINT uq_invoice_student_period UNIQUE (student_id, fee_type_id, period_year, period_month),
    CONSTRAINT chk_invoice_status CHECK (status IN ('unpaid', 'partial', 'paid', 'overdue', 'waived')),
    CONSTRAINT chk_invoice_month CHECK (period_month IS NULL OR (period_month >= 1 AND period_month <= 12)),
    CONSTRAINT chk_invoice_amount CHECK (amount > 0),
    CONSTRAINT chk_invoice_total CHECK (total_amount >= 0),
    CONSTRAINT chk_invoice_paid CHECK (paid_amount >= 0 AND paid_amount <= total_amount)
);

-- Indexes
CREATE INDEX idx_invoice_tenant_company ON student_invoices (tenant_id, company_id);
CREATE INDEX idx_invoice_student ON student_invoices (student_id);
CREATE INDEX idx_invoice_status ON student_invoices (status) WHERE status != 'paid';
CREATE INDEX idx_invoice_due_date ON student_invoices (due_date) WHERE status IN ('unpaid', 'partial', 'overdue');
CREATE INDEX idx_invoice_year_month ON student_invoices (academic_year_id, period_year, period_month);
CREATE INDEX idx_invoice_fee_type ON student_invoices (fee_type_id);
CREATE INDEX idx_invoice_rels ON student_invoices USING GIN (_rels);
CREATE INDEX idx_invoice_data ON student_invoices USING GIN (_data);

-- Pembayaran
CREATE TABLE student_payments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    invoice_id      UUID NOT NULL,
    student_id      UUID NOT NULL,

    -- Payment detail
    receipt_no      VARCHAR(50) NOT NULL,
    amount          BIGINT NOT NULL,
    payment_method  VARCHAR(20) NOT NULL,
    payment_date    DATE NOT NULL,
    note            TEXT,

    -- Pencatat
    received_by     UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_receipt_no UNIQUE (tenant_id, company_id, receipt_no),
    CONSTRAINT chk_payment_amount CHECK (amount > 0),
    CONSTRAINT chk_payment_method CHECK (payment_method IN ('cash', 'transfer', 'debit', 'qris', 'va'))
);

-- Indexes
CREATE INDEX idx_payment_tenant_company ON student_payments (tenant_id, company_id);
CREATE INDEX idx_payment_invoice ON student_payments (invoice_id);
CREATE INDEX idx_payment_student ON student_payments (student_id);
CREATE INDEX idx_payment_date ON student_payments (payment_date);
CREATE INDEX idx_payment_rels ON student_payments USING GIN (_rels);
CREATE INDEX idx_payment_data ON student_payments USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `amount` | BIGINT (bukan NUMERIC) | Rupiah tidak punya desimal, BIGINT lebih efisien |
| `default_amount` | di fee_types | Template amount — bisa di-override per invoice (beasiswa) |
| `discount` | BIGINT di invoice | Potongan per siswa (beasiswa, anak guru, dll) |
| `total_amount` | Computed: `amount - discount` | Denormalisasi untuk query cepat |
| `paid_amount` | BIGINT di invoice | Running total — diupdate setiap payment masuk |
| `status` | VARCHAR(20), CHECK | 5 status: unpaid → partial → paid, atau overdue, waived |
| `period_month` | INT nullable | NULL untuk tagihan tahunan/one-time, 1-12 untuk bulanan |
| `payment_method` | VARCHAR(20), CHECK | 5 metode: cash, transfer, debit, QRIS, virtual account |
| `received_by` | UUID, NOT NULL | Audit trail: siapa yang menerima pembayaran |

### Vernon Relationships

**student_invoices:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Selalu perlu tahu pemilik tagihan |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |
| `fee_type` | belongs_to | **Ya** | Nama dan kategori tagihan |

**student_payments:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `invoice` | belongs_to | **Ya** | Selalu perlu tahu tagihan yang dibayar |
| `student` | belongs_to | **Ya** | Pemilik pembayaran |

### _rels / _data Structure

```json
// student_invoices
{
  "_rels": {
    "student_id": "018f...",
    "academic_year_id": "018f...",
    "fee_type_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" },
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "fee_type": { "id": "018f...", "name": "SPP", "code": "SPP", "fee_category": "monthly" }
  }
}

// student_payments
{
  "_rels": {
    "invoice_id": "018f...",
    "student_id": "018f..."
  },
  "_data": {
    "invoice": { "id": "018f...", "invoice_no": "INV-2026-001", "total_amount": 500000 },
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" }
  }
}
```

### API Endpoints

```
# Fee Types (Master)
GET    /api/v1/fee-types                            — List jenis tagihan
POST   /api/v1/fee-types                            — Buat jenis tagihan
PUT    /api/v1/fee-types/{id}                       — Update jenis tagihan

# Invoices
GET    /api/v1/students/{id}/invoices               — Tagihan siswa
GET    /api/v1/student-invoices                     — List semua tagihan (admin)
POST   /api/v1/student-invoices/generate            — Generate tagihan bulk (per kelas/angkatan)
PUT    /api/v1/student-invoices/{id}                — Update tagihan (discount, waive)

# Payments
GET    /api/v1/students/{id}/payments               — Riwayat pembayaran siswa
GET    /api/v1/student-invoices/{id}/payments       — Pembayaran per tagihan
POST   /api/v1/student-payments                     — Record pembayaran
GET    /api/v1/student-payments/{id}/receipt         — Cetak kwitansi

# Reports
GET    /api/v1/finance/summary                      — Rekap keuangan per bulan
GET    /api/v1/finance/overdue                      — Daftar tunggakan
```

### Invoice Generation Flow

```
1. Admin pilih: fee_type + academic_year + target (kelas/angkatan/semua)
2. System query students berdasarkan target
3. Untuk setiap siswa: buat invoice dengan amount dari fee_type.default_amount
4. Admin bisa adjust: discount per siswa (beasiswa), waive tagihan
5. Invoice status = 'unpaid', due_date = tanggal jatuh tempo
```

### Payment Flow

```
1. Admin pilih invoice yang akan dibayar
2. Input: amount, payment_method, payment_date
3. System:
   a. Create student_payment record
   b. Update invoice.paid_amount += payment.amount
   c. Update invoice.status:
      - paid_amount == 0: 'unpaid'
      - paid_amount < total_amount: 'partial'
      - paid_amount >= total_amount: 'paid'
4. Generate receipt_no otomatis
```

## Consequences

### Positive

- **3-layer clean**: Fee type (template) → Invoice (tagihan) → Payment (bayar) memisahkan concern dengan jelas.
- **Flexible billing**: Mendukung bulanan, tahunan, one-time, dan incidental.
- **Partial payment**: Cicilan didukung secara native lewat `paid_amount`.
- **Bulk generate**: Tagihan bisa di-generate massal per kelas/angkatan.
- **Audit trail**: Setiap pembayaran tercatat siapa yang menerima.
- **QRIS/VA ready**: Payment method mendukung metode digital.

### Negative / Trade-offs

- **Tidak double-entry**: Ini bukan akuntansi penuh — hanya pencatatan tagihan dan pembayaran siswa. Jika perlu jurnal umum, butuh domain accounting terpisah.
- **Invoice generation complexity**: Bulk generate perlu handle edge case (siswa pindah tengah semester, siswa baru masuk tengah tahun).
- **Overdue detection**: Perlu cron job untuk update status `unpaid` → `overdue` berdasarkan `due_date`.
- **No refund**: Belum ada mekanisme pengembalian pembayaran. Bisa ditambah sebagai enhancement.

## Alternatives Considered

### 1. Satu tabel payment tanpa invoice
- Ditolak: tidak bisa track tunggakan — perlu tahu "berapa yang harus dibayar" vs "berapa yang sudah dibayar".

### 2. Invoice tanpa fee_type master
- Ditolak: tanpa template, admin harus input jenis + amount manual setiap kali generate — error prone dan tidak konsisten.

### 3. NUMERIC(15,2) untuk amount
- Ditolak: Rupiah tidak punya pecahan desimal dalam konteks SPP. BIGINT lebih hemat storage dan menghindari floating point issues.

### 4. Pisah domain keuangan dari student
- Ditolak: finance siswa sangat erat dengan student entity — invoice di-scope per student_id. Memisahkan ke microservice terpisah terlalu premature untuk MVP.

## Test Cases

### Unit Tests — Descriptor & Validation

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `FeeTypeDescriptor.TableName()` | — | `"fee_types"` |
| U02 | `InvoiceDescriptor.TableName()` | — | `"student_invoices"` |
| U03 | `PaymentDescriptor.TableName()` | — | `"student_payments"` |
| U04 | Validate rejects invalid `fee_category` | `"quarterly"` | Error: invalid fee_category |
| U05 | Validate rejects `amount <= 0` for invoice | `amount = 0` | Error: amount must be positive |
| U06 | Validate rejects invalid `payment_method` | `"bitcoin"` | Error: invalid payment_method |
| U07 | Validate rejects `paid_amount > total_amount` | `paid = 600000, total = 500000` | Error: paid exceeds total |
| U08 | Validate rejects invalid `status` | `"cancelled"` | Error: invalid status |
| U09 | Validate accepts valid invoice | All fields valid | No error |

### Integration Tests — Fee Types

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create fee type | POST with name + code + category + amount | 201, created |
| I02 | Unique code per company | Create 2 fee types with same code | 409/422, unique constraint |
| I03 | List fee types | GET /fee-types | 200, filtered by tenant + company |
| I04 | Category CHECK enforced | INSERT with `fee_category = 'weekly'` | DB error |

### Integration Tests — Invoices

| # | Test Case | Action | Expected |
|---|---|---|---|
| I05 | Generate invoices for class | POST /student-invoices/generate for class VII-A | 201, one invoice per student |
| I06 | Unique per student+type+period | Generate same invoice twice | 409/422, unique constraint |
| I07 | Get student invoices | GET /students/{id}/invoices | 200, with `_data` |
| I08 | Apply discount | PUT invoice with `discount = 100000` | 200, `total_amount` updated |
| I09 | Waive invoice | PUT with `status = 'waived'` | 200, status changed |
| I10 | Invoice status CHECK | INSERT with `status = 'cancelled'` | DB error |
| I11 | Period month nullable for annual | Create annual invoice with `period_month = NULL` | 201, allowed |
| I12 | Period month range check | Create with `period_month = 13` | DB error |

### Integration Tests — Payments

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | Record full payment | POST payment with amount = total_amount | 201, invoice status → `paid` |
| I14 | Record partial payment | POST payment with amount < total_amount | 201, invoice status → `partial` |
| I15 | Multiple partial payments | 2 payments summing to total | Invoice status → `paid` |
| I16 | Payment exceeds remaining | POST payment > (total - paid) | 422, rejected |
| I17 | Unique receipt number | Create 2 payments with same receipt_no | 409/422 |
| I18 | Payment method CHECK | INSERT with `payment_method = 'bitcoin'` | DB error |
| I19 | Get payment receipt | GET /student-payments/{id}/receipt | 200, receipt data |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I20 | StudentUpdated syncs to invoices | Update student name | `_data.student.full_name` updated |
| I21 | FeeTypeUpdated syncs to invoices | Update fee_type name | `_data.fee_type.name` updated |
| I22 | InvoiceUpdated syncs to payments | Update invoice number | `_data.invoice.invoice_no` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I23 | Cannot access other tenant's invoices | GET with wrong tenant | 404 |
| I24 | Cannot record payment cross-tenant | POST payment for other tenant's invoice | Error |
