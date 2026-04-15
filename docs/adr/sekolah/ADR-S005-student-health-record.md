# ADR-S005: Student Health Record

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Data kesehatan siswa dibutuhkan untuk:

1. **UKS (Usaha Kesehatan Sekolah)**: Pemantauan kesehatan berkala.
2. **Kegiatan olahraga**: Identifikasi siswa dengan kondisi khusus (asma, jantung) sebelum kegiatan fisik.
3. **Alergi**: Informasi kritis untuk kantin sekolah dan kegiatan outdoor.
4. **Asuransi**: Data BPJS atau asuransi swasta untuk klaim.
5. **Dapodik**: Pelaporan kondisi kesehatan dan disabilitas ke Kemendikbud.
6. **Inklusi**: Identifikasi siswa berkebutuhan khusus untuk program pendampingan.

Data kesehatan bersifat **berkala** — diukur setiap semester atau setiap tahun ajaran. Sehingga satu siswa memiliki **banyak health records** yang menunjukkan perkembangan fisik dari waktu ke waktu.

### Mengapa Vernon Pattern?

- has_many dari student.
- Read-heavy: data diakses oleh guru olahraga, UKS, dan dashboard.
- Business logic minimal: CRUD + riwayat.
- Eventual consistency acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk domain `student_health`.

### Table Schema

```sql
CREATE TABLE student_health (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign key
    student_id      UUID NOT NULL,

    -- Pengukuran fisik
    height_cm       NUMERIC(5,1),
    weight_kg       NUMERIC(5,1),

    -- Kondisi
    eye_condition   VARCHAR(50),
    hearing_condition VARCHAR(50),
    dental_condition VARCHAR(50),

    -- Riwayat medis
    allergies       TEXT[],
    chronic_diseases TEXT[],
    disability_type VARCHAR(50),
    disability_note TEXT,

    -- Asuransi
    insurance_type  VARCHAR(20),
    insurance_number VARCHAR(50),

    -- Tanggal pengukuran
    measured_at     DATE NOT NULL,
    measured_by     VARCHAR(255),

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_health_insurance CHECK (insurance_type IN ('bpjs', 'swasta', 'none') OR insurance_type IS NULL)
);

-- Indexes
CREATE INDEX idx_health_tenant_company ON student_health (tenant_id, company_id);
CREATE INDEX idx_health_student ON student_health (student_id);
CREATE INDEX idx_health_measured_at ON student_health (student_id, measured_at DESC);
CREATE INDEX idx_health_rels ON student_health USING GIN (_rels);
CREATE INDEX idx_health_data ON student_health USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `height_cm`, `weight_kg` | NUMERIC(5,1) | Presisi 1 desimal cukup untuk pemantauan tumbuh kembang |
| `allergies`, `chronic_diseases` | TEXT[] | Array karena bisa lebih dari satu, dan jarang di-query secara individual |
| `disability_type` | VARCHAR(50) | Kategori umum (tuna netra, tuna rungu, dll) untuk pelaporan Dapodik |
| `measured_at` | DATE, NOT NULL | Setiap record HARUS punya tanggal pengukuran untuk tracking historis |
| `measured_by` | VARCHAR(255) | Nama petugas UKS atau dokter yang mengukur |
| `insurance_type` | VARCHAR(20) | BPJS / swasta / none — menentukan proses klaim |

### Vernon Relationships

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Selalu perlu tahu pemilik record |

### API Endpoints

```
GET    /api/v1/students/{id}/health              — Riwayat kesehatan siswa (ordered by measured_at DESC)
GET    /api/v1/students/{id}/health/latest       — Data kesehatan terbaru
POST   /api/v1/student-health                    — Tambah record kesehatan
PUT    /api/v1/student-health/{id}               — Update record
```

## Consequences

### Positive

- **Historis berkala**: Setiap pengukuran tersimpan sebagai record terpisah — bisa lihat tren pertumbuhan.
- **Alergi sebagai array**: Mudah ditambah/dikurangi tanpa mengubah schema.
- **Inklusi**: `disability_type` mendukung identifikasi ABK (Anak Berkebutuhan Khusus).
- **Audit trail**: `measured_by` + `measured_at` memberikan akuntabilitas data.

### Negative / Trade-offs

- **Data sensitif**: Data kesehatan memerlukan akses kontrol ketat — belum di-define di ADR ini.
- **Array querying**: PostgreSQL array (`TEXT[]`) kurang efisien untuk full-text search dibanding tabel terpisah.
- **Tidak ada BMI auto-calc**: BMI harus dihitung di application layer dari height dan weight.

## Alternatives Considered

### 1. Satu row per siswa (latest only)
- Ditolak: tidak bisa tracking pertumbuhan dan perubahan kondisi dari waktu ke waktu.

### 2. Alergi dan penyakit sebagai tabel terpisah
- Ditolak: over-normalized untuk kebutuhan saat ini. Array cukup — jarang di-query secara individual.

### 3. JSONB untuk semua data kesehatan
- Ditolak: field seperti height dan weight perlu di-aggregate (rata-rata, min, max) — kolom eksplisit lebih efisien.

## Test Cases

### Unit Tests — Descriptor & Validation

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` returns correct name | — | `"student_health"` |
| U02 | `DefaultRels()` returns 1 autoloaded rel | — | `student` (autoload: true) |
| U03 | Validate rejects missing `measured_at` | `{ ..., "measured_at": null }` | Error: measured_at required |
| U04 | Validate rejects invalid `insurance_type` | `{ ..., "insurance_type": "unknown" }` | Error: invalid insurance_type |
| U05 | Validate accepts valid health record | All required fields valid | No error |
| U06 | Validate accepts empty arrays for allergies | `{ ..., "allergies": [] }` | No error |

### Integration Tests — CRUD

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create health record | `POST /api/v1/student-health` with valid payload | 201, record created |
| I02 | Get student's health history | `GET /api/v1/students/{id}/health` | 200, list ordered by `measured_at DESC` |
| I03 | Get latest health record | `GET /api/v1/students/{id}/health/latest` | 200, returns most recent record |
| I04 | Update health record | `PUT /api/v1/student-health/{id}` | 200, fields updated |
| I05 | `_data` includes student snapshot | Create record with valid student_id | `_data.student` contains `{id, full_name, nis}` |

### Integration Tests — Physical Measurements

| # | Test Case | Action | Expected |
|---|---|---|---|
| I06 | Height stored with 1 decimal | Create with `height_cm = 155.5` | Stored as `155.5` |
| I07 | Weight stored with 1 decimal | Create with `weight_kg = 45.3` | Stored as `45.3` |
| I08 | Height and weight nullable | Create without height and weight | 201, both NULL |
| I09 | Multiple records show growth trend | Create 3 records with increasing height at different dates | GET returns chronological growth data |

### Integration Tests — Arrays & Medical

| # | Test Case | Action | Expected |
|---|---|---|---|
| I10 | Allergies stored as array | Create with `allergies = ["kacang", "udang"]` | Stored as PostgreSQL TEXT array |
| I11 | Chronic diseases stored as array | Create with `chronic_diseases = ["asma"]` | Stored as PostgreSQL TEXT array |
| I12 | Empty arrays allowed | Create with `allergies = []`, `chronic_diseases = []` | 201, empty arrays stored |
| I13 | Disability fields nullable | Create without `disability_type` and `disability_note` | 201, both NULL |

### Integration Tests — Insurance & Constraints

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | Insurance type CHECK enforced | INSERT with `insurance_type = 'unknown'` | DB error, CHECK violation |
| I15 | Insurance type accepts valid values | Create with `insurance_type = 'bpjs'` | 201, stored correctly |
| I16 | Insurance type nullable | Create without `insurance_type` | 201, NULL allowed |
| I17 | `measured_by` stored correctly | Create with `measured_by = "Dr. Sari"` | 200, stored for audit |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I18 | Cannot access other tenant's health records | GET with wrong tenant scope | 404 |
| I19 | Cannot create health record for student in different company | POST with mismatched company_id | Error, scope violation |
