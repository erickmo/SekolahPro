# ADR-S055: Dapodik Integration (Integrasi Dapodik)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

**Dapodik** (Data Pokok Pendidikan) adalah sistem database nasional dari Kementerian Pendidikan, Kebudayaan, Riset, dan Teknologi (Kemendikbud Ristek) yang **wajib** diisi oleh seluruh sekolah di Indonesia. Dapodik menjadi dasar untuk:

1. **BOS (Bantuan Operasional Sekolah)**: Dana BOS dihitung berdasarkan jumlah siswa di Dapodik — jika data Dapodik tidak akurat, dana BOS bisa berkurang atau tertunda.
2. **Tunjangan guru**: Sertifikasi guru dan tunjangan profesi berdasarkan data NUPTK/NIP di Dapodik.
3. **Akreditasi sekolah**: Data Dapodik menjadi input BAN-S/M (Badan Akreditasi Nasional Sekolah/Madrasah).
4. **NISN (Nomor Induk Siswa Nasional)**: Identitas unik siswa se-Indonesia — wajib ada di Dapodik.
5. **NPSN (Nomor Pokok Sekolah Nasional)**: Identitas unik sekolah — semua laporan merujuk ke NPSN.
6. **NUPTK (Nomor Unik Pendidik dan Tenaga Kependidikan)**: Identitas unik guru/pegawai.

Masalah operator Dapodik saat ini:
- **Double entry**: Operator input data ke SekolahPro (atau sistem internal) DAN ke Dapodik secara manual — sangat memakan waktu.
- **Data inconsistency**: Data di SekolahPro dan Dapodik sering tidak sinkron.
- **Validasi ketat**: Dapodik punya ratusan aturan validasi — data yang tidak valid ditolak, operator harus debugging.
- **Deadline pressure**: Dapodik ada cutoff tanggal per semester — keterlambatan berdampak ke dana BOS.

### Mengapa Vernon Pattern?

- Read-heavy: admin dan operator sering cek status sync, history, dan error logs.
- Relasi ke student, teacher, school_profile.
- Write periodik: sync terjadi per semester (2x setahun) dengan batch processing.
- Business logic complex: field mapping, validation rules, conflict resolution.
- Eventual consistency acceptable — Dapodik sendiri process data secara batch.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `dapodik_field_mappings` (mapping field SekolahPro ↔ Dapodik), `dapodik_sync_batches` (batch sync per semester), dan `dapodik_sync_records` (status per entity yang di-sync).

> **C-Suite CTO Review Note (2026-04-15):**
> Dapodik **tidak memiliki public REST API yang stabil.** Aplikasi Dapodik berjalan sebagai
> aplikasi desktop lokal (Django-based) yang konek ke server sekolah dan sync ke server pusat
> Kemendikbud via mekanisme internal. Integrasi realistis per 2026:
>
> **Mode utama (v1.0): `manual_export`**
> - SekolahPro menghasilkan file export (CSV/Excel/JSON) dalam format yang siap diimport ke Dapodik
> - Operator Dapodik memvalidasi data di SekolahPro, download file, lalu import manual ke Dapodik
> - Ini menghilangkan **double entry** (masalah utama operator) tanpa butuh API integration
>
> **Mode aspirational (v2.0+): `api_sync`**
> - Jika Kemendikbud membuka API atau ada solusi community (scraping/automation Dapodik lokal)
> - Ini bersifat **deferred** — jangan investasi engineering time sebelum ada API yang stabil
>
> **Rekomendasi:** Ubah `sync_mode` default ke `manual_export`. Arsitektur sudah mendukung
> kedua mode (`sync_mode IN ('manual_export', 'api_sync')`), tapi v1.0 harus fokus pada
> export yang akurat, validasi pre-export yang ketat, dan field mapping yang lengkap.
> Fitur `api_sync` tetap ada di schema tapi **tidak diimplementasikan di Phase 1**.

### Table Schema

```sql
-- Mapping field SekolahPro → Dapodik
CREATE TABLE dapodik_field_mappings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas mapping
    entity_type     VARCHAR(30) NOT NULL,
    sp_field        VARCHAR(100) NOT NULL,
    dapodik_field   VARCHAR(100) NOT NULL,
    dapodik_table   VARCHAR(100) NOT NULL,

    -- Transformasi
    transform_type  VARCHAR(20) NOT NULL DEFAULT 'direct',
    transform_config JSONB,

    -- Validasi Dapodik
    is_required     BOOLEAN NOT NULL DEFAULT false,
    validation_rule VARCHAR(255),
    dapodik_lookup  JSONB,

    -- Status
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

    CONSTRAINT uq_dapodik_mapping UNIQUE (tenant_id, company_id, entity_type, sp_field),
    CONSTRAINT chk_entity_type CHECK (entity_type IN ('student', 'teacher', 'school_profile', 'class', 'subject', 'attendance', 'grade', 'infrastructure')),
    CONSTRAINT chk_transform_type CHECK (transform_type IN ('direct', 'lookup', 'format', 'concat', 'split', 'custom'))
);

CREATE INDEX idx_dapodik_mapping_tenant ON dapodik_field_mappings (tenant_id, company_id);
CREATE INDEX idx_dapodik_mapping_entity ON dapodik_field_mappings (entity_type);
CREATE INDEX idx_dapodik_mapping_rels ON dapodik_field_mappings USING GIN (_rels);
CREATE INDEX idx_dapodik_mapping_data ON dapodik_field_mappings USING GIN (_data);

-- Batch sync per semester
CREATE TABLE dapodik_sync_batches (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas batch
    academic_year_id UUID NOT NULL,
    semester        INT NOT NULL,
    batch_no        VARCHAR(30) NOT NULL,
    description     TEXT,

    -- Konfigurasi
    sync_mode       VARCHAR(20) NOT NULL,
    entity_types    JSONB NOT NULL DEFAULT '[]',

    -- School identity (Dapodik)
    npsn            VARCHAR(8) NOT NULL,
    dapodik_server_url TEXT,

    -- Stats
    total_records   INT NOT NULL DEFAULT 0,
    pending_count   INT NOT NULL DEFAULT 0,
    synced_count    INT NOT NULL DEFAULT 0,
    error_count     INT NOT NULL DEFAULT 0,
    skipped_count   INT NOT NULL DEFAULT 0,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    initiated_by    UUID NOT NULL,
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,

    -- Timing
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    error_message   TEXT,

    -- Export file (for manual mode)
    export_file_url TEXT,
    export_format   VARCHAR(10),

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_dapodik_batch UNIQUE (tenant_id, company_id, academic_year_id, semester, batch_no),
    CONSTRAINT chk_dapodik_semester CHECK (semester IN (1, 2)),
    CONSTRAINT chk_dapodik_sync_mode CHECK (sync_mode IN ('manual_export', 'api_sync')),
    CONSTRAINT chk_dapodik_batch_status CHECK (status IN ('draft', 'validating', 'validated', 'validation_error', 'approved', 'syncing', 'completed', 'partial', 'error')),
    CONSTRAINT chk_dapodik_export_format CHECK (export_format IS NULL OR export_format IN ('json', 'csv', 'xlsx'))
);

CREATE INDEX idx_dapodik_batch_tenant ON dapodik_sync_batches (tenant_id, company_id);
CREATE INDEX idx_dapodik_batch_year ON dapodik_sync_batches (academic_year_id, semester);
CREATE INDEX idx_dapodik_batch_status ON dapodik_sync_batches (status) WHERE status NOT IN ('completed');
CREATE INDEX idx_dapodik_batch_rels ON dapodik_sync_batches USING GIN (_rels);
CREATE INDEX idx_dapodik_batch_data ON dapodik_sync_batches USING GIN (_data);

-- Status per entity yang di-sync
CREATE TABLE dapodik_sync_records (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    batch_id        UUID NOT NULL,
    entity_type     VARCHAR(30) NOT NULL,
    entity_id       UUID NOT NULL,

    -- Dapodik identifiers
    nisn            VARCHAR(10),
    nuptk           VARCHAR(16),
    npsn            VARCHAR(8),

    -- Data snapshot
    sp_data         JSONB NOT NULL DEFAULT '{}',
    dapodik_data    JSONB NOT NULL DEFAULT '{}',
    diff_fields     JSONB NOT NULL DEFAULT '[]',

    -- Validation
    validation_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    validation_errors JSONB NOT NULL DEFAULT '[]',

    -- Sync status
    sync_status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    sync_response   JSONB,
    synced_at       TIMESTAMPTZ,
    retry_count     INT NOT NULL DEFAULT 0,
    max_retries     INT NOT NULL DEFAULT 3,
    next_retry_at   TIMESTAMPTZ,

    -- Error detail
    error_code      VARCHAR(50),
    error_message   TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_dapodik_record UNIQUE (batch_id, entity_type, entity_id),
    CONSTRAINT chk_dapodik_rec_entity CHECK (entity_type IN ('student', 'teacher', 'school_profile', 'class', 'subject', 'attendance', 'grade', 'infrastructure')),
    CONSTRAINT chk_dapodik_validation CHECK (validation_status IN ('pending', 'valid', 'invalid', 'warning')),
    CONSTRAINT chk_dapodik_rec_status CHECK (sync_status IN ('pending', 'validating', 'validated', 'syncing', 'synced', 'error', 'retry', 'skipped'))
);

CREATE INDEX idx_dapodik_record_tenant ON dapodik_sync_records (tenant_id, company_id);
CREATE INDEX idx_dapodik_record_batch ON dapodik_sync_records (batch_id);
CREATE INDEX idx_dapodik_record_entity ON dapodik_sync_records (entity_type, entity_id);
CREATE INDEX idx_dapodik_record_nisn ON dapodik_sync_records (nisn) WHERE nisn IS NOT NULL;
CREATE INDEX idx_dapodik_record_nuptk ON dapodik_sync_records (nuptk) WHERE nuptk IS NOT NULL;
CREATE INDEX idx_dapodik_record_sync ON dapodik_sync_records (sync_status) WHERE sync_status IN ('pending', 'error', 'retry');
CREATE INDEX idx_dapodik_record_validation ON dapodik_sync_records (validation_status) WHERE validation_status = 'invalid';
CREATE INDEX idx_dapodik_record_rels ON dapodik_sync_records USING GIN (_rels);
CREATE INDEX idx_dapodik_record_data ON dapodik_sync_records USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `npsn` | VARCHAR(8) di batch | NPSN selalu 8 digit — identitas sekolah di Dapodik |
| `nisn` | VARCHAR(10) di records | NISN selalu 10 digit — identitas siswa nasional |
| `nuptk` | VARCHAR(16) di records | NUPTK 16 digit — identitas guru/pegawai nasional |
| `sp_data` / `dapodik_data` | JSONB snapshots | Snapshot data saat sync — untuk audit dan debugging jika ada perbedaan |
| `diff_fields` | JSONB array | Field-field yang berbeda antara SekolahPro dan Dapodik — memudahkan review |
| `validation_errors` | JSONB array | List error validasi Dapodik per record — operator bisa fix satu-satu |
| `transform_type` | VARCHAR(20) | Jenis transformasi: direct (1:1), lookup (kode → kode), format (date format), concat, split |
| `dapodik_lookup` | JSONB | Tabel referensi Dapodik — misal kode agama: 1=Islam, 2=Kristen, dst |
| `sync_mode` | 2 mode | manual_export: generate file untuk upload manual. api_sync: langsung kirim ke server Dapodik |
| `retry_count` / `max_retries` | INT | Auto-retry untuk transient error — max 3x sebelum mark as error |
| `export_file_url` | TEXT, nullable | Hanya diisi untuk mode manual_export — URL file yang di-generate |

### Dapodik Sync Flow

```
┌─────────────────────────────────────────────────────────────┐
│ Dapodik Sync Flow                                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ 1. PREPARE                                                  │
│    Operator buat batch → pilih semester + entity types       │
│    System generate dapodik_sync_records dari data SekolahPro │
│                                                             │
│ 2. VALIDATE                                                 │
│    Batch status: draft → validating                          │
│    Setiap record di-validasi against Dapodik rules:         │
│    - NISN format (10 digit)                                 │
│    - NUPTK format (16 digit)                                │
│    - Tanggal lahir tidak kosong                              │
│    - Agama harus kode Dapodik (1-6)                          │
│    - Jenis kelamin: L/P                                      │
│    - Alamat lengkap wajib ada                                │
│    Result: valid / invalid / warning per record              │
│                                                             │
│ 3. REVIEW & FIX                                              │
│    Operator review validation_errors                         │
│    Fix data di SekolahPro → re-validate                      │
│    Batch status: validated / validation_error                │
│                                                             │
│ 4. APPROVE                                                   │
│    Kepsek approve batch → status: approved                   │
│                                                             │
│ 5. SYNC (2 modes)                                           │
│    A) Manual Export:                                         │
│       Generate file (JSON/CSV/XLSX) → operator upload        │
│       ke aplikasi Dapodik desktop                            │
│    B) API Sync:                                              │
│       Kirim langsung ke server Dapodik via API               │
│       (jika sekolah punya akses API Dapodik)                 │
│                                                             │
│ 6. TRACK                                                     │
│    Per record: pending → syncing → synced / error / retry    │
│    Batch: syncing → completed / partial / error              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Vernon Relationships

**dapodik_sync_batches:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran dan semester |
| `initiated_by` (user) | belongs_to | **Ya** | Siapa yang membuat batch |
| `approved_by` (user) | belongs_to | Tidak | Hanya perlu saat review detail |

**dapodik_sync_records:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `batch` | belongs_to | **Ya** | Selalu perlu konteks batch |
| `entity` (polymorphic) | belongs_to | Tidak | Entity bisa student/teacher/dll — di-resolve on demand |

### _rels / _data Structure

```json
// dapodik_sync_batches
{
  "_rels": {
    "academic_year_id": "018f...",
    "initiated_by": "018f..."
  },
  "_data": {
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "initiated_by_user": { "id": "018f...", "full_name": "Operator Dapodik", "role": "tata_usaha" }
  }
}

// dapodik_sync_records
{
  "_rels": {
    "batch_id": "018f...",
    "entity_id": "018f..."
  },
  "_data": {
    "batch": { "id": "018f...", "batch_no": "BATCH-2026-S1-001", "npsn": "12345678" },
    "entity_summary": { "name": "Ahmad Fauzi", "nisn": "0012345678", "type": "student" }
  }
}
```

### API Endpoints

```
# Field Mappings
GET    /api/v1/dapodik/field-mappings                    — List mapping per entity type
POST   /api/v1/dapodik/field-mappings                    — Buat/update mapping
PUT    /api/v1/dapodik/field-mappings/{id}               — Update mapping

# Sync Batches
GET    /api/v1/dapodik/sync-batches                      — List batch per tahun ajaran
POST   /api/v1/dapodik/sync-batches                      — Buat batch baru
GET    /api/v1/dapodik/sync-batches/{id}                 — Detail batch + stats
POST   /api/v1/dapodik/sync-batches/{id}/validate        — Jalankan validasi
POST   /api/v1/dapodik/sync-batches/{id}/approve         — Approve batch (kepsek)
POST   /api/v1/dapodik/sync-batches/{id}/sync            — Jalankan sync (API mode)
POST   /api/v1/dapodik/sync-batches/{id}/export          — Generate export file (manual mode)
GET    /api/v1/dapodik/sync-batches/{id}/download        — Download export file

# Sync Records
GET    /api/v1/dapodik/sync-records                      — List records per batch
GET    /api/v1/dapodik/sync-records/{id}                 — Detail record + validation errors
POST   /api/v1/dapodik/sync-records/{id}/revalidate      — Re-validate setelah fix
POST   /api/v1/dapodik/sync-records/{id}/retry           — Retry failed sync
POST   /api/v1/dapodik/sync-records/{id}/skip            — Skip record dari sync

# Dashboard
GET    /api/v1/dapodik/dashboard                         — Stats: total, synced, error, pending
GET    /api/v1/dapodik/validation-summary                — Summary validation errors per field
```

## Consequences

### Positive

- **Eliminasi double entry**: Data di SekolahPro otomatis di-mapping ke format Dapodik — operator tidak perlu input ulang.
- **Validation before sync**: Error Dapodik terdeteksi sebelum submit — mengurangi reject rate.
- **Audit trail lengkap**: Snapshot data + diff + error per record — memudahkan debugging.
- **Two sync modes**: Manual export untuk sekolah tanpa akses API, API sync untuk yang sudah terintegrasi.
- **Retry mechanism**: Transient error otomatis di-retry — mengurangi beban operator.
- **NPSN/NISN/NUPTK tracking**: Identifier nasional tersimpan dan tervalidasi.

### Negative / Trade-offs

- **Dapodik API instability**: API Dapodik sering berubah tanpa notice — perlu adapter layer yang mudah di-update.
- **Validation rules maintenance**: Ratusan aturan validasi Dapodik harus di-maintain sesuai versi terbaru.
- **Complex field mapping**: Beberapa field perlu transformasi non-trivial (lookup kode, format tanggal, concatenation).
- **Seasonal urgency**: Deadline Dapodik per semester menciptakan pressure — system harus reliable saat cutoff.
- **Security concern**: Data NISN, NUPTK, biodata lengkap adalah PII — perlu encryption dan access control ketat.

## Alternatives Considered

### 1. Direct database sync ke Dapodik
- Ditolak: Dapodik tidak expose database — hanya API dan file upload yang available.

### 2. Hanya export file tanpa validasi
- Ditolak: tanpa pre-validation, error rate tinggi — operator harus bolak-balik fix dan re-upload.

### 3. Real-time sync (setiap data berubah di SekolahPro)
- Ditolak: Dapodik tidak support real-time — data diproses batch per semester. Real-time sync juga wasteful karena data bisa berubah berkali-kali sebelum cutoff.

### 4. Embed Dapodik validation rules di client-side
- Ditolak: validation rules terlalu banyak dan sering berubah — server-side validation lebih maintainable.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `FieldMappingDescriptor.TableName()` | — | `"dapodik_field_mappings"` |
| U02 | `SyncBatchDescriptor.TableName()` | — | `"dapodik_sync_batches"` |
| U03 | `SyncRecordDescriptor.TableName()` | — | `"dapodik_sync_records"` |
| U04 | Validate NISN format | `"001234567"` (9 digit) | Error: NISN must be 10 digits |
| U05 | Validate NISN format valid | `"0012345678"` (10 digit) | No error |
| U06 | Validate NUPTK format | `"12345"` (5 digit) | Error: NUPTK must be 16 digits |
| U07 | Validate NPSN format | `"123456789"` (9 digit) | Error: NPSN must be 8 digits |
| U08 | Transform lookup: agama | `"islam"` | `1` (Dapodik code) |
| U09 | Transform lookup: gender | `"L"` | `1` (Dapodik code for laki-laki) |
| U10 | Validate rejects invalid `entity_type` | `"parent"` | Error: invalid entity_type |
| U11 | Validate rejects invalid `sync_mode` | `"realtime"` | Error: invalid sync_mode |
| U12 | Validate rejects semester > 2 | `semester = 3` | Error: semester must be 1 or 2 |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create sync batch | POST with academic_year + semester | 201, batch_no generated |
| I02 | Generate sync records from students | Create batch for entity_type=student | Records created for all active students |
| I03 | Validate batch — all valid | POST /validate | All records validation_status = 'valid' |
| I04 | Validate batch — some invalid | Student without NISN | Record validation_status = 'invalid', errors listed |
| I05 | Fix and revalidate | Update student NISN, POST /revalidate | Record validation_status → 'valid' |
| I06 | Approve batch | POST /approve as kepsek | Status → 'approved', approved_by set |
| I07 | Export file generated | POST /export for manual_export mode | File URL generated, export_format set |
| I08 | API sync — success | POST /sync with valid Dapodik server | Records sync_status → 'synced' |
| I09 | API sync — partial failure | Some records fail | Batch status = 'partial', failed records have error details |
| I10 | Retry failed record | POST /retry for error record | retry_count incremented, re-synced |
| I11 | Max retries exceeded | Retry 3 times, still fails | sync_status stays 'error', no more retries |
| I12 | Skip record | POST /skip | sync_status → 'skipped', batch counts updated |
| I13 | Unique batch per semester | Create duplicate batch_no | 409/422, unique constraint |
| I14 | Semester CHECK enforced | INSERT with `semester = 3` | DB error |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I15 | AcademicYearUpdated syncs to batches | Update academic year name | `_data.academic_year.name` updated |
| I16 | BatchUpdated syncs to records | Update batch batch_no | `_data.batch.batch_no` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Cannot access other tenant's batches | GET with wrong tenant | 404 |
| I18 | Cannot sync other tenant's students | Create record for other tenant entity | Error |
| I19 | Export file tenant-isolated | Download other tenant's export | 404 |
