# ADR-S050: School Budget Management (RKAS)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

RKAS (Rencana Kegiatan dan Anggaran Sekolah) adalah dokumen wajib bagi setiap sekolah di Indonesia. RKAS mengatur perencanaan anggaran tahunan yang mencakup seluruh kegiatan sekolah dan menjadi dasar penggunaan dana BOS, APBD, yayasan, dan sumber lainnya.

Manajemen anggaran sekolah diperlukan untuk:

1. **Perencanaan tahunan**: Setiap tahun ajaran, sekolah menyusun RKAS yang harus disetujui oleh komite sekolah dan dinas pendidikan.
2. **Multi-sumber dana**: Sekolah menerima dana dari berbagai sumber — BOS (Bantuan Operasional Sekolah), APBD, Yayasan, Komite, dan Dana Mandiri — masing-masing punya aturan penggunaan berbeda.
3. **Alokasi per kegiatan**: Setiap kegiatan sekolah harus punya alokasi anggaran yang jelas dari sumber dana tertentu.
4. **Realisasi & monitoring**: Pelacakan realisasi anggaran vs rencana — berapa yang sudah terpakai, berapa sisa.
5. **Pelaporan BOS**: Dana BOS wajib dilaporkan ke Kemendikbud dalam format standar (8 standar / 13 komponen BOS Reguler).
6. **Audit trail**: Setiap transaksi anggaran harus tercatat untuk audit internal dan eksternal.
7. **Pesantren/Yayasan**: Lembaga pesantren memiliki sumber dana tambahan (infaq, wakaf, donatur) yang perlu dikelola bersamaan.

### Mengapa Vernon Pattern?

- Read-heavy: dashboard anggaran, laporan realisasi, perbandingan rencana vs aktual.
- Relasi ke academic_year, kegiatan sekolah (S058), dan sumber dana.
- Business logic moderate: alokasi, validasi sumber dana, perhitungan sisa anggaran.
- Write terjadi periodik (saat penyusunan RKAS dan pencatatan realisasi), bukan transaksional tinggi.
- Eventual consistency acceptable untuk reporting.

## Decision

Menggunakan **Vernon Pattern** untuk 4 tabel: `budget_sources` (sumber dana), `budget_plans` (RKAS per tahun ajaran), `budget_items` (alokasi per kegiatan), dan `budget_realizations` (realisasi pengeluaran).

### Table Schema

```sql
-- Master sumber dana
CREATE TABLE budget_sources (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    source_type     VARCHAR(30) NOT NULL,
    description     TEXT,

    -- Aturan penggunaan
    usage_rules     TEXT,
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

    CONSTRAINT uq_budget_source_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_source_type CHECK (source_type IN ('bos_reguler', 'bos_kinerja', 'bos_afirmasi', 'apbd', 'yayasan', 'komite', 'dana_mandiri', 'infaq', 'wakaf', 'donatur', 'other'))
);

CREATE INDEX idx_budget_source_tenant ON budget_sources (tenant_id, company_id);

-- RKAS per tahun ajaran
CREATE TABLE budget_plans (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    academic_year_id UUID NOT NULL,

    -- Identitas
    plan_name       VARCHAR(200) NOT NULL,
    fiscal_year     INT NOT NULL,

    -- Total anggaran
    total_planned   BIGINT NOT NULL DEFAULT 0,
    total_realized  BIGINT NOT NULL DEFAULT 0,

    -- Status workflow
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_budget_plan_year UNIQUE (tenant_id, company_id, academic_year_id),
    CONSTRAINT chk_plan_status CHECK (status IN ('draft', 'submitted', 'approved', 'revised', 'closed')),
    CONSTRAINT chk_plan_total CHECK (total_planned >= 0),
    CONSTRAINT chk_plan_realized CHECK (total_realized >= 0)
);

CREATE INDEX idx_budget_plan_tenant ON budget_plans (tenant_id, company_id);
CREATE INDEX idx_budget_plan_year ON budget_plans (academic_year_id);
CREATE INDEX idx_budget_plan_status ON budget_plans (status);
CREATE INDEX idx_budget_plan_rels ON budget_plans USING GIN (_rels);
CREATE INDEX idx_budget_plan_data ON budget_plans USING GIN (_data);

-- Alokasi anggaran per kegiatan
CREATE TABLE budget_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    budget_plan_id  UUID NOT NULL,
    budget_source_id UUID NOT NULL,

    -- Kategori BOS (8 standar SNP)
    bos_component   VARCHAR(50),

    -- Detail kegiatan
    activity_name   VARCHAR(200) NOT NULL,
    activity_code   VARCHAR(30),
    description     TEXT,

    -- Anggaran
    planned_amount  BIGINT NOT NULL,
    realized_amount BIGINT NOT NULL DEFAULT 0,
    remaining_amount BIGINT NOT NULL,

    -- Periode (triwulan)
    quarter         INT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_planned_amount CHECK (planned_amount > 0),
    CONSTRAINT chk_realized_amount CHECK (realized_amount >= 0),
    CONSTRAINT chk_remaining CHECK (remaining_amount >= 0),
    CONSTRAINT chk_quarter CHECK (quarter IS NULL OR (quarter >= 1 AND quarter <= 4)),
    CONSTRAINT chk_bos_component CHECK (bos_component IS NULL OR bos_component IN (
        'standar_kompetensi_lulusan', 'standar_isi', 'standar_proses',
        'standar_penilaian', 'standar_ptk', 'standar_sarpras',
        'standar_pengelolaan', 'standar_pembiayaan'
    ))
);

CREATE INDEX idx_budget_item_plan ON budget_items (budget_plan_id);
CREATE INDEX idx_budget_item_source ON budget_items (budget_source_id);
CREATE INDEX idx_budget_item_component ON budget_items (bos_component) WHERE bos_component IS NOT NULL;
CREATE INDEX idx_budget_item_rels ON budget_items USING GIN (_rels);
CREATE INDEX idx_budget_item_data ON budget_items USING GIN (_data);

-- Realisasi pengeluaran
CREATE TABLE budget_realizations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    budget_item_id  UUID NOT NULL,

    -- Detail transaksi
    transaction_date DATE NOT NULL,
    description     VARCHAR(500) NOT NULL,
    amount          BIGINT NOT NULL,
    receipt_no      VARCHAR(50),
    vendor_name     VARCHAR(200),

    -- Bukti
    attachment_url  TEXT,

    -- Pencatat
    recorded_by     UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_realization_amount CHECK (amount > 0)
);

CREATE INDEX idx_realization_item ON budget_realizations (budget_item_id);
CREATE INDEX idx_realization_date ON budget_realizations (transaction_date);
CREATE INDEX idx_realization_tenant ON budget_realizations (tenant_id, company_id);
CREATE INDEX idx_realization_rels ON budget_realizations USING GIN (_rels);
CREATE INDEX idx_realization_data ON budget_realizations USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `source_type` | VARCHAR(30), CHECK 11 values | Mencakup semua sumber dana sekolah termasuk pesantren (infaq, wakaf, donatur) |
| `bos_component` | VARCHAR(50), nullable | 8 Standar Nasional Pendidikan — wajib untuk pelaporan BOS, NULL untuk dana non-BOS |
| `planned_amount` / `realized_amount` | BIGINT | Rupiah tanpa desimal, konsisten dengan S009 |
| `remaining_amount` | Denormalisasi computed | `planned_amount - realized_amount` — untuk query cepat sisa anggaran |
| `quarter` | INT 1-4, nullable | Triwulan untuk pelaporan BOS (Triwulan I-IV), NULL jika tidak terikat triwulan |
| `fiscal_year` | INT di budget_plans | Tahun anggaran (bisa beda dengan tahun ajaran, tapi biasanya sama) |
| `status` workflow | 5 status | draft → submitted → approved → (revised) → closed |
| `recorded_by` | UUID NOT NULL | Audit trail: siapa yang mencatat realisasi |

### Vernon Relationships

**budget_plans:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `academic_year` | belongs_to | **Ya** | Selalu perlu konteks tahun ajaran |

**budget_items:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `budget_plan` | belongs_to | **Ya** | Konteks RKAS induk |
| `budget_source` | belongs_to | **Ya** | Informasi sumber dana |

**budget_realizations:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `budget_item` | belongs_to | **Ya** | Konteks kegiatan dan alokasi |

### _rels / _data Structure

```json
// budget_plans
{
  "_rels": {
    "academic_year_id": "018f..."
  },
  "_data": {
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}

// budget_items
{
  "_rels": {
    "budget_plan_id": "018f...",
    "budget_source_id": "018f..."
  },
  "_data": {
    "budget_plan": { "id": "018f...", "plan_name": "RKAS 2025/2026", "fiscal_year": 2025 },
    "budget_source": { "id": "018f...", "name": "BOS Reguler", "code": "BOS-REG", "source_type": "bos_reguler" }
  }
}

// budget_realizations
{
  "_rels": {
    "budget_item_id": "018f..."
  },
  "_data": {
    "budget_item": {
      "id": "018f...",
      "activity_name": "Pengadaan Buku Pelajaran",
      "planned_amount": 50000000,
      "budget_source": { "name": "BOS Reguler", "code": "BOS-REG" }
    }
  }
}
```

### API Endpoints

```
# Budget Sources (Master)
GET    /api/v1/budget-sources                          — List sumber dana
POST   /api/v1/budget-sources                          — Buat sumber dana
PUT    /api/v1/budget-sources/{id}                     — Update sumber dana

# Budget Plans (RKAS)
GET    /api/v1/budget-plans                            — List RKAS
POST   /api/v1/budget-plans                            — Buat RKAS baru
GET    /api/v1/budget-plans/{id}                       — Detail RKAS dengan semua item
PUT    /api/v1/budget-plans/{id}                       — Update RKAS
PUT    /api/v1/budget-plans/{id}/submit                — Submit untuk approval
PUT    /api/v1/budget-plans/{id}/approve               — Approve RKAS
PUT    /api/v1/budget-plans/{id}/close                 — Tutup RKAS akhir tahun

# Budget Items (Alokasi per Kegiatan)
GET    /api/v1/budget-plans/{id}/items                 — List alokasi dalam RKAS
POST   /api/v1/budget-items                            — Buat alokasi kegiatan
PUT    /api/v1/budget-items/{id}                       — Update alokasi
DELETE /api/v1/budget-items/{id}                       — Hapus alokasi (soft delete)

# Realizations (Realisasi)
GET    /api/v1/budget-items/{id}/realizations           — List realisasi per kegiatan
POST   /api/v1/budget-realizations                      — Catat realisasi pengeluaran
PUT    /api/v1/budget-realizations/{id}                 — Update realisasi
GET    /api/v1/budget-realizations/{id}                 — Detail realisasi

# Reports
GET    /api/v1/budget-plans/{id}/summary               — Ringkasan RKAS (rencana vs realisasi)
GET    /api/v1/budget-plans/{id}/bos-report             — Laporan BOS format standar
GET    /api/v1/budget-plans/{id}/by-source              — Rekapitulasi per sumber dana
GET    /api/v1/budget-plans/{id}/by-quarter             — Realisasi per triwulan
```

## Consequences

### Positive

- **Compliance BOS**: Format 8 standar SNP terpenuhi untuk pelaporan ke Kemendikbud.
- **Multi-sumber dana**: Mendukung semua jenis sumber dana termasuk konteks pesantren.
- **Real-time monitoring**: Dashboard sisa anggaran per kegiatan, per sumber dana.
- **Audit trail**: Setiap realisasi tercatat siapa pencatatnya dan bukti transaksi.
- **Linked to events**: Budget item bisa dikaitkan dengan kegiatan sekolah (S058) untuk perencanaan terpadu.
- **Triwulan tracking**: Mendukung pelaporan BOS per triwulan sesuai ketentuan Kemendikbud.

### Negative / Trade-offs

- **Bukan akuntansi penuh**: Ini budget management, bukan double-entry accounting. Untuk jurnal umum dan neraca, perlu domain accounting terpisah.
- **Remaining amount denormalisasi**: Perlu dijaga sinkron setiap kali realisasi dicatat — risiko inkonsistensi jika update gagal.
- **BOS component mapping**: 8 standar SNP bisa berubah jika kebijakan Kemendikbud berubah — perlu flexibility.
- **Approval workflow sederhana**: Hanya 1 level approval — sekolah besar mungkin perlu multi-level (Kepala Sekolah → Yayasan).

## Alternatives Considered

### 1. Spreadsheet-based (Excel export/import)
- Ditolak: tidak bisa tracking real-time, rawan human error, tidak ada audit trail.

### 2. Satu tabel budget tanpa pemisahan plan/item/realization
- Ditolak: terlalu flat — tidak bisa query per kegiatan, per sumber dana, per triwulan.

### 3. Full accounting system (double-entry)
- Ditolak: overkill untuk MVP. Sekolah butuh budget tracking, bukan general ledger. Bisa ditambah nanti.

### 4. BOS component sebagai separate table
- Ditolak: 8 standar SNP relatif stabil — CHECK constraint cukup. Jika berubah, migration saja.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `BudgetSourceDescriptor.TableName()` | — | `"budget_sources"` |
| U02 | `BudgetPlanDescriptor.TableName()` | — | `"budget_plans"` |
| U03 | `BudgetItemDescriptor.TableName()` | — | `"budget_items"` |
| U04 | `BudgetRealizationDescriptor.TableName()` | — | `"budget_realizations"` |
| U05 | Validate rejects invalid `source_type` | `"hibah"` | Error: invalid source_type |
| U06 | Validate rejects `planned_amount <= 0` | `planned_amount = 0` | Error: must be positive |
| U07 | Validate rejects invalid `bos_component` | `"standar_xyz"` | Error: invalid component |
| U08 | Validate rejects invalid `status` | `"cancelled"` | Error: invalid status |
| U09 | Validate rejects quarter out of range | `quarter = 5` | Error: must be 1-4 |
| U10 | Validate accepts valid budget item | All fields valid | No error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create budget source | POST with valid data | 201, created |
| I02 | Unique source code per company | Create 2 with same code | 409/422, unique constraint |
| I03 | Create RKAS | POST budget plan for academic year | 201, created |
| I04 | Unique RKAS per academic year | Create 2 for same year | 409/422, unique constraint |
| I05 | Add budget item | POST item with source + amount | 201, plan total_planned updated |
| I06 | Record realization | POST realization for item | 201, item realized_amount updated |
| I07 | Realization exceeds planned | POST realization > remaining | 422, rejected |
| I08 | Submit RKAS for approval | PUT /submit | 200, status → submitted |
| I09 | Approve RKAS | PUT /approve with approver | 200, status → approved, approved_at set |
| I10 | BOS report format | GET /bos-report | 200, grouped by 8 standar SNP |
| I11 | Summary by source | GET /by-source | 200, totals per source type |
| I12 | Quarter filter | GET /by-quarter?quarter=1 | 200, only Q1 items |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | AcademicYearUpdated syncs to plans | Update year name | `_data.academic_year.name` updated |
| I14 | BudgetSourceUpdated syncs to items | Update source name | `_data.budget_source.name` updated |
| I15 | BudgetItemUpdated syncs to realizations | Update activity name | `_data.budget_item.activity_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Cannot access other tenant's RKAS | GET with wrong tenant | 404 |
| I17 | Cannot record realization cross-tenant | POST realization for other tenant | Error |
