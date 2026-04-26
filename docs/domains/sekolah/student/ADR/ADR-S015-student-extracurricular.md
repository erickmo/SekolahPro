# ADR-S015: Student Extracurricular

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Kegiatan ekstrakurikuler adalah bagian penting dari pendidikan holistik di Indonesia. Data ekskul diperlukan untuk:

1. **Rapor**: Nilai dan predikat ekskul wajib dicantumkan di rapor (Kurikulum Merdeka).
2. **Dapodik**: Pelaporan kegiatan ekskul dan peserta ke Kemendikbud.
3. **Pembinaan minat**: Tracking partisipasi siswa di kegiatan non-akademik.
4. **Absensi ekskul**: Kehadiran per sesi ekskul.
5. **Penilaian ekskul**: Predikat dan deskripsi per semester untuk rapor.
6. **Pesantren**: Ekskul keagamaan (Tahfidz, Kaligrafi) sesuai ADR-009.

Arsitektur terdiri dari 3 layer:
- **Extracurricular**: Master kegiatan ekskul yang tersedia di sekolah.
- **Enrollment**: Siswa mendaftar ke ekskul per tahun ajaran.
- **Assessment**: Penilaian per semester (untuk rapor).

### Mengapa Vernon Pattern?

- Many-to-many: siswa ↔ ekskul (melalui enrollment).
- Read-heavy: rapor, dashboard, dan laporan.
- Business logic sederhana: enroll, assess, report.
- Eventual consistency acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `extracurriculars` (master), `student_extracurricular_enrollments` (pendaftaran), dan `student_extracurricular_assessments` (penilaian).

### Table Schema

```sql
-- Master kegiatan ekskul
CREATE TABLE extracurriculars (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    description     TEXT,
    category        VARCHAR(20) NOT NULL,

    -- Konfigurasi
    day_of_week     VARCHAR(10),
    time_start      TIME,
    time_end        TIME,
    max_capacity    INT,
    coach_id        UUID,
    is_mandatory    BOOLEAN NOT NULL DEFAULT false,
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

    CONSTRAINT uq_extracurricular_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_extracurricular_category CHECK (category IN (
        'sports', 'arts', 'science', 'technology', 'language',
        'religious', 'social', 'scouts', 'other'
    )),
    CONSTRAINT chk_extracurricular_day CHECK (day_of_week IS NULL OR day_of_week IN (
        'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'
    ))
);

-- Indexes
CREATE INDEX idx_extracurricular_tenant_company ON extracurriculars (tenant_id, company_id);

-- Pendaftaran siswa ke ekskul
CREATE TABLE student_extracurricular_enrollments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id          UUID NOT NULL,
    extracurricular_id  UUID NOT NULL,
    academic_year_id    UUID NOT NULL,

    -- Status
    enrollment_status VARCHAR(20) NOT NULL DEFAULT 'active',
    enrolled_date     DATE NOT NULL,
    withdrawn_date    DATE,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_enrollment_student_ekskul_year UNIQUE (student_id, extracurricular_id, academic_year_id),
    CONSTRAINT chk_enrollment_status CHECK (enrollment_status IN ('active', 'withdrawn', 'completed'))
);

-- Indexes
CREATE INDEX idx_enrollment_tenant_company ON student_extracurricular_enrollments (tenant_id, company_id);
CREATE INDEX idx_enrollment_student ON student_extracurricular_enrollments (student_id);
CREATE INDEX idx_enrollment_ekskul ON student_extracurricular_enrollments (extracurricular_id);
CREATE INDEX idx_enrollment_year ON student_extracurricular_enrollments (academic_year_id);
CREATE INDEX idx_enrollment_rels ON student_extracurricular_enrollments USING GIN (_rels);
CREATE INDEX idx_enrollment_data ON student_extracurricular_enrollments USING GIN (_data);

-- Penilaian ekskul per semester
CREATE TABLE student_extracurricular_assessments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    enrollment_id   UUID NOT NULL,
    student_id      UUID NOT NULL,
    extracurricular_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Periode
    semester        VARCHAR(10) NOT NULL,

    -- Penilaian
    grade           VARCHAR(2) NOT NULL,
    description     TEXT,

    -- Assessor
    assessed_by     UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_assessment_enrollment_semester UNIQUE (enrollment_id, semester),
    CONSTRAINT chk_assessment_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_assessment_grade CHECK (grade IN ('SB', 'B', 'C', 'K'))
);

-- Indexes
CREATE INDEX idx_assessment_tenant_company ON student_extracurricular_assessments (tenant_id, company_id);
CREATE INDEX idx_assessment_enrollment ON student_extracurricular_assessments (enrollment_id);
CREATE INDEX idx_assessment_student ON student_extracurricular_assessments (student_id);
CREATE INDEX idx_assessment_rels ON student_extracurricular_assessments USING GIN (_rels);
CREATE INDEX idx_assessment_data ON student_extracurricular_assessments USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `category` | VARCHAR(20), CHECK | 9 kategori termasuk `religious` untuk pesantren (ADR-009) |
| `is_mandatory` | BOOLEAN | Beberapa ekskul wajib (Pramuka di Kurikulum 2013) |
| `coach_id` | UUID, nullable | Guru pembina — nullable karena belum tentu terdaftar di system |
| `max_capacity` | INT, nullable | Batas peserta — null = unlimited |
| `enrollment_status` | VARCHAR(20), CHECK | active/withdrawn/completed lifecycle |
| `grade` | VARCHAR(2), CHECK | SB/B/C/K — standar penilaian ekskul rapor Indonesia |
| `description` | TEXT | Deskripsi naratif untuk rapor (wajib di Kurikulum Merdeka) |

### Vernon Relationships

**student_extracurricular_enrollments:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Pemilik enrollment |
| `extracurricular` | belongs_to | **Ya** | Nama dan kategori ekskul |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |

**student_extracurricular_assessments:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `enrollment` | belongs_to | **Ya** | Enrollment terkait |
| `student` | belongs_to | **Ya** | Pemilik assessment |
| `extracurricular` | belongs_to | **Ya** | Nama ekskul |

### _rels / _data Structure

```json
// enrollment
{
  "_rels": {
    "student_id": "018f...",
    "extracurricular_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" },
    "extracurricular": { "id": "018f...", "name": "Pramuka", "code": "PRM", "category": "scouts" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}

// assessment
{
  "_rels": {
    "enrollment_id": "018f...",
    "student_id": "018f...",
    "extracurricular_id": "018f..."
  },
  "_data": {
    "enrollment": { "id": "018f...", "enrollment_status": "active" },
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" },
    "extracurricular": { "id": "018f...", "name": "Pramuka", "code": "PRM" }
  }
}
```

### API Endpoints

```
# Extracurriculars (Master)
GET    /api/v1/extracurriculars                      — List ekskul
POST   /api/v1/extracurriculars                      — Buat ekskul
PUT    /api/v1/extracurriculars/{id}                 — Update ekskul

# Enrollments
GET    /api/v1/students/{id}/extracurriculars        — Ekskul siswa
POST   /api/v1/student-extracurricular-enrollments    — Enroll siswa ke ekskul
PUT    /api/v1/student-extracurricular-enrollments/{id} — Update status (withdraw)
GET    /api/v1/extracurriculars/{id}/members          — Anggota ekskul

# Assessments
GET    /api/v1/students/{id}/extracurricular-assessments — Penilaian ekskul siswa
POST   /api/v1/student-extracurricular-assessments/bulk — Bulk input penilaian per ekskul
PUT    /api/v1/student-extracurricular-assessments/{id} — Update penilaian
```

## Consequences

### Positive

- **Rapor ready**: Grade + description per semester sesuai format rapor Indonesia.
- **Kurikulum Merdeka compliant**: Penilaian ekskul menggunakan predikat (SB/B/C/K).
- **Pesantren support**: Kategori `religious` mendukung ekskul keagamaan.
- **Capacity control**: `max_capacity` mencegah over-enrollment.
- **Enrollment lifecycle**: Tracking active → withdrawn → completed.

### Negative / Trade-offs

- **Tidak ada absensi per sesi**: MVP hanya menyimpan enrollment dan assessment per semester — absensi per pertemuan ekskul bisa ditambah sebagai enhancement.
- **3 tabel**: Complexity lebih tinggi dibanding single table, tapi separation of concerns lebih baik.
- **Coach management**: `coach_id` hanya reference — belum ada domain teacher/staff.
- **No scheduling**: Jadwal ekskul (day + time) bersifat informational — belum terintegrasi dengan calendar.

## Alternatives Considered

### 1. Satu tabel saja (enrollment + assessment merged)
- Ditolak: enrollment bersifat per tahun ajaran, assessment bersifat per semester — granularity berbeda.

### 2. Assessment sebagai JSONB di enrollment
- Ditolak: sulit query "semua penilaian semester ganjil" cross-enrollments.

### 3. Absensi ekskul di S008 (daily attendance)
- Deferred: S008 dirancang untuk absensi kelas harian. Absensi ekskul bisa ditambah sebagai sub-type di S008 atau tabel terpisah di enhancement.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `ExtracurricularDescriptor.TableName()` | — | `"extracurriculars"` |
| U02 | `EnrollmentDescriptor.TableName()` | — | `"student_extracurricular_enrollments"` |
| U03 | `AssessmentDescriptor.TableName()` | — | `"student_extracurricular_assessments"` |
| U04 | Validate rejects invalid `category` | `"music"` | Error |
| U05 | Validate rejects invalid `day_of_week` | `"sunday"` | Error |
| U06 | Validate rejects invalid `enrollment_status` | `"paused"` | Error |
| U07 | Validate rejects invalid `grade` | `"A"` | Error: must be SB/B/C/K |
| U08 | Validate accepts valid enrollment | All fields valid | No error |

### Integration Tests — Extracurriculars

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create extracurricular | POST with valid data | 201 |
| I02 | Unique code per company | Create 2 with same code | 409/422 |
| I03 | Category CHECK | INSERT with `category = 'music'` | DB error |

### Integration Tests — Enrollments

| # | Test Case | Action | Expected |
|---|---|---|---|
| I04 | Enroll student | POST enrollment | 201 |
| I05 | Unique per student+ekskul+year | Enroll same student twice | 409/422 |
| I06 | Get student extracurriculars | GET /students/{id}/extracurriculars | 200 |
| I07 | Get ekskul members | GET /extracurriculars/{id}/members | 200 |
| I08 | Withdraw student | PUT with `enrollment_status = 'withdrawn'` | 200 |
| I09 | Capacity check | Enroll when members = max_capacity | 422, full |

### Integration Tests — Assessments

| # | Test Case | Action | Expected |
|---|---|---|---|
| I10 | Bulk input assessments | POST /assessments/bulk for ekskul members | 201 |
| I11 | Unique per enrollment+semester | Assess same enrollment same semester | 409/422 |
| I12 | Grade CHECK | INSERT with `grade = 'A'` | DB error |
| I13 | Semester CHECK | INSERT with `semester = 'midterm'` | DB error |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | StudentUpdated syncs to enrollments | Update student name | `_data.student` updated |
| I15 | ExtracurricularUpdated syncs | Update ekskul name | `_data.extracurricular` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Cannot access other tenant's enrollments | GET with wrong tenant | 404 |
