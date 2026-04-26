# ADR-010: Academic Years (Tahun Ajaran)

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Tahun ajaran adalah **unit waktu fundamental** dalam sistem pendidikan Indonesia. Hampir semua domain student memiliki foreign key ke `academic_year_id`:

- S004 Academic Record
- S008 Daily Attendance
- S009 Finance/SPP (Invoices)
- S011 Subject Grades
- S012 Discipline
- S013 Achievement
- S014 Class Placement
- S015 Extracurricular
- S016 PPDB
- S018 Rapor

Tanpa tabel `academic_years`, **tidak ada domain student yang bisa diimplementasi**.

Tahun ajaran di Indonesia:
- Dimulai **Juli** dan berakhir **Juni** tahun berikutnya.
- Terdiri dari 2 semester: **Ganjil** (Juli-Desember) dan **Genap** (Januari-Juni).
- Satu sekolah hanya punya **satu tahun ajaran aktif** pada satu waktu.
- Tahun ajaran sebelumnya tetap bisa diakses (historical) tapi tidak bisa diedit setelah ditutup.

### Mengapa Vernon Pattern?

- Referenced by 10+ domain sebagai belongs_to.
- Read-heavy: setiap query student memerlukan konteks tahun ajaran.
- Jumlah record sedikit (1 per tahun = ~10 records seumur hidup sekolah).
- SyncEngine: perubahan nama tahun ajaran harus propagate ke semua `_data`.

## Decision

Menggunakan **Vernon Pattern** untuk domain `academic_year`.

### Table Schema

```sql
CREATE TABLE academic_years (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(20) NOT NULL,
    code            VARCHAR(10) NOT NULL,

    -- Periode
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,

    -- Semester dates
    semester1_start DATE NOT NULL,
    semester1_end   DATE NOT NULL,
    semester2_start DATE NOT NULL,
    semester2_end   DATE NOT NULL,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT false,
    status          VARCHAR(20) NOT NULL DEFAULT 'planning',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_academic_year_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT uq_academic_year_name UNIQUE (tenant_id, company_id, name),
    CONSTRAINT chk_academic_year_status CHECK (status IN ('planning', 'active', 'closed')),
    CONSTRAINT chk_academic_year_dates CHECK (start_date < end_date),
    CONSTRAINT chk_semester1_dates CHECK (semester1_start < semester1_end),
    CONSTRAINT chk_semester2_dates CHECK (semester2_start < semester2_end),
    CONSTRAINT chk_semester_order CHECK (semester1_end <= semester2_start)
);

-- Hanya satu tahun ajaran aktif per company
CREATE UNIQUE INDEX uq_academic_year_active
    ON academic_years (tenant_id, company_id)
    WHERE is_active = true;

-- Indexes
CREATE INDEX idx_academic_year_tenant_company ON academic_years (tenant_id, company_id);
CREATE INDEX idx_academic_year_active ON academic_years (is_active) WHERE is_active = true;
CREATE INDEX idx_academic_year_rels ON academic_years USING GIN (_rels);
CREATE INDEX idx_academic_year_data ON academic_years USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `name` | VARCHAR(20), e.g. "2025/2026" | Format standar tahun ajaran Indonesia |
| `code` | VARCHAR(10), e.g. "2526" | Short code untuk referensi (invoice numbering, dll) |
| `start_date` / `end_date` | DATE | Periode penuh tahun ajaran (Juli-Juni) |
| `semester*_start/end` | DATE | Tanggal eksplisit per semester — bisa beda antar sekolah |
| `is_active` | BOOLEAN, unique partial | Hanya satu aktif per company — enforced di database |
| `status` | 3 stage lifecycle | planning (setup) → active (berjalan) → closed (arsip) |

### Lifecycle

```
1. Admin buat tahun ajaran baru → status = 'planning'
   - Setup: tanggal semester, kelas, wali kelas
   - Belum ada data operasional

2. Admin aktivasi → status = 'active', is_active = true
   - Tahun ajaran sebelumnya otomatis: is_active = false, status = 'closed'
   - Semua operasional (absensi, nilai, SPP) mengacu ke tahun ajaran ini
   - student._data.academic_year diupdate via SyncEngine

3. Di akhir tahun → admin tutup → status = 'closed'
   - Data tetap bisa dibaca (historical)
   - Tidak bisa diedit (immutable after close)
```

### Vernon Relationships

`academic_years` tidak punya belongs_to — ini entity root. Tapi ia menjadi **sumber _data** untuk 10+ domain:

| Domain yang reference | Field di _data |
|---|---|
| students | `_data.academic_year.name`, `_data.academic_year.is_active` |
| student_academics | `_data.academic_year.name` |
| student_attendances | `_data.academic_year.name` |
| student_grades | `_data.academic_year.name` |
| student_invoices | `_data.academic_year.name` |
| ... (10+ domains) | ... |

### SyncEngine Impact

Ketika `academic_years.name` diupdate, SyncEngine harus propagate ke **semua domain** yang punya `academic_year_id`. Ini adalah sync terbesar di system — tapi jarang terjadi (nama tahun ajaran hampir tidak pernah berubah).

### API Endpoints

```
GET    /api/v1/academic-years                        — List tahun ajaran
GET    /api/v1/academic-years/active                 — Tahun ajaran aktif saat ini
POST   /api/v1/academic-years                        — Buat tahun ajaran baru
PUT    /api/v1/academic-years/{id}                   — Update (planning stage only)
POST   /api/v1/academic-years/{id}/activate          — Aktivasi (deactivate yang lama)
POST   /api/v1/academic-years/{id}/close             — Tutup tahun ajaran
```

## Consequences

### Positive

- **Single source of truth**: Semua domain reference satu tabel academic_years.
- **Active enforcement**: Database-level unique constraint menjamin hanya 1 aktif per company.
- **Semester dates explicit**: Tidak perlu kalkulasi — tanggal semester jelas per sekolah.
- **Lifecycle clear**: planning → active → closed dengan aturan imutabilitas.

### Negative / Trade-offs

- **SyncEngine heavy**: Update nama tahun ajaran memicu sync ke 10+ domain — harus async.
- **Single active limitation**: Beberapa sekolah mungkin dalam transisi 2 tahun ajaran bersamaan (kelas XII belum selesai, kelas VII sudah mulai). Ini di-handle dengan `status = 'active'` di tahun baru sementara data kelas XII masih bisa diakses di tahun lama.
- **No semester entity**: Semester bukan tabel terpisah — hanya field VARCHAR + tanggal di academic_year. Jika perlu semester-level config nanti, butuh tabel terpisah.

## Alternatives Considered

### 1. Semester sebagai tabel terpisah
- Deferred: untuk MVP, semester sebagai field di academic_year cukup. Bisa ditambah nanti jika perlu config per semester.

### 2. Tahun ajaran global (tanpa company scope)
- Ditolak: dalam multi-tenant, setiap sekolah bisa punya kalender berbeda (pesantren mulai Muharram, dll).

### 3. Tanpa partial unique index (enforce di application)
- Ditolak: database-level enforcement lebih aman — tidak mungkin punya 2 active year meski ada race condition.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` | — | `"academic_years"` |
| U02 | Validate rejects `start_date >= end_date` | start > end | Error |
| U03 | Validate rejects `semester1_end > semester2_start` | overlap | Error |
| U04 | Validate rejects invalid `status` | `"archived"` | Error |
| U05 | Validate accepts valid academic year | All fields valid | No error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create academic year | POST with valid data | 201 |
| I02 | Unique name per company | Create 2 with same name | 409/422 |
| I03 | Unique code per company | Create 2 with same code | 409/422 |
| I04 | Activate academic year | POST /activate | 200, `is_active = true` |
| I05 | Only one active per company | Activate second year | Previous year `is_active = false` |
| I06 | Cannot activate 2 simultaneously | Direct INSERT 2 active rows | DB error (partial unique) |
| I07 | Close academic year | POST /close | 200, `status = 'closed'` |
| I08 | Cannot edit closed year | PUT on closed year | 422, immutable |
| I09 | Get active year | GET /academic-years/active | 200, returns current active |
| I10 | Date CHECK constraints | INSERT with start > end | DB error |
| I11 | SyncEngine propagates name change | Update name on planning year | All referencing domains' `_data` updated |
| I12 | Tenant isolation | Access other tenant's year | 404 |
