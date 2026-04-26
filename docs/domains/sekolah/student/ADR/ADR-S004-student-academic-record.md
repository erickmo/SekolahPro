# ADR-S004: Student Academic Record

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Riwayat akademik per semester adalah data inti yang menghubungkan siswa dengan kinerja belajarnya. Data ini dibutuhkan untuk:

1. **Rapor**: Rekap nilai dan absensi per semester.
2. **Kenaikan kelas**: Keputusan naik/tinggal/lulus berdasarkan data akademik.
3. **Dashboard siswa**: Menampilkan progress akademik secara visual.
4. **Pelaporan**: Data agregat per kelas, per angkatan, per sekolah.
5. **Historis**: Riwayat lengkap dari masuk hingga lulus.

Satu siswa memiliki **banyak academic records** — satu per semester sepanjang masa sekolah. Untuk SMP 3 tahun = 6 records, SMA 3 tahun = 6 records.

### Mengapa Vernon Pattern?

- has_many dari student (banyak records per siswa).
- Read-heavy: dashboard dan rapor memerlukan akses cepat.
- Relasi ke class_room dan academic_year.
- Business logic sederhana: CRUD + summary calculation.

## Decision

Menggunakan **Vernon Pattern** untuk domain `student_academic`.

### Table Schema

```sql
CREATE TABLE student_academics (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    class_room_id   UUID NOT NULL,

    -- Semester
    semester        VARCHAR(10) NOT NULL,

    -- Rekap absensi
    days_present    INT NOT NULL DEFAULT 0,
    days_sick       INT NOT NULL DEFAULT 0,
    days_permitted  INT NOT NULL DEFAULT 0,
    days_absent     INT NOT NULL DEFAULT 0,

    -- Rekap nilai
    grade_average   NUMERIC(5,2),
    class_rank      INT,
    total_students  INT,

    -- Status
    promotion_status VARCHAR(20),
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

    CONSTRAINT uq_student_semester UNIQUE (student_id, academic_year_id, semester),
    CONSTRAINT chk_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_promotion CHECK (promotion_status IN ('promoted', 'retained', 'graduated') OR promotion_status IS NULL)
);

-- Indexes
CREATE INDEX idx_academics_tenant_company ON student_academics (tenant_id, company_id);
CREATE INDEX idx_academics_student ON student_academics (student_id);
CREATE INDEX idx_academics_year_semester ON student_academics (academic_year_id, semester);
CREATE INDEX idx_academics_class ON student_academics (class_room_id);
CREATE INDEX idx_academics_rels ON student_academics USING GIN (_rels);
CREATE INDEX idx_academics_data ON student_academics USING GIN (_data);
```

### Vernon Relationships

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Selalu perlu tahu siapa pemilik record |
| `academic_year` | belongs_to | **Ya** | Selalu perlu tahu tahun ajaran |
| `class_room` | belongs_to | **Ya** | Selalu perlu tahu kelas |

### _rels / _data Structure

```json
{
  "_rels": {
    "student_id": "018f...",
    "academic_year_id": "018f...",
    "class_room_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" },
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "class_room": { "id": "018f...", "name": "VII-A", "grade_level": "7" }
  }
}
```

### API Endpoints

```
GET    /api/v1/students/{id}/academics              — Riwayat akademik siswa
GET    /api/v1/students/{id}/academics/{semester}    — Detail per semester
POST   /api/v1/student-academics                     — Buat record akademik
PUT    /api/v1/student-academics/{id}                — Update record
```

## Consequences

### Positive

- **Unique per semester**: Constraint mencegah duplikasi record per siswa per semester.
- **Absensi terstruktur**: 4 kolom absensi (hadir/sakit/izin/alpha) memudahkan perhitungan dan laporan.
- **Ranking**: `class_rank` dan `total_students` memungkinkan tampilan ranking di dashboard.
- **Autoload lengkap**: Student + kelas + tahun ajaran tersedia langsung di setiap record.

### Negative / Trade-offs

- **Summary, bukan detail**: ADR ini hanya menyimpan rekap per semester — detail nilai per mata pelajaran membutuhkan domain terpisah (`grades`).
- **Ranking manual**: `class_rank` harus di-calculate dan di-update oleh application — bukan computed column.
- **Batch update**: Kenaikan kelas massal di akhir semester memerlukan bulk operation.

## Alternatives Considered

### 1. Nilai detail per mata pelajaran di tabel ini
- Ditolak: terlalu banyak data per row, lebih baik sebagai domain terpisah (`grades`) yang di-aggregate ke sini.

### 2. JSONB untuk absensi dan nilai
- Ditolak: kolom eksplisit lebih mudah di-query, di-index, dan di-aggregate dibanding JSONB path queries.

### 3. Satu row per tahun (bukan per semester)
- Ditolak: Indonesia menggunakan sistem semester — rapor per semester, kenaikan kelas per tahun tapi evaluasi per semester.

## Test Cases

### Unit Tests — Descriptor & Validation

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` returns correct name | — | `"student_academics"` |
| U02 | `DefaultRels()` returns 3 autoloaded rels | — | `student`, `academic_year`, `class_room` (all autoload: true) |
| U03 | Validate rejects invalid `semester` | `{ ..., "semester": "midterm" }` | Error: semester must be ganjil or genap |
| U04 | Validate rejects negative attendance values | `{ ..., "days_present": -1 }` | Error: days_present cannot be negative |
| U05 | Validate accepts valid academic record | All required fields valid | No error |

### Integration Tests — CRUD

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create academic record | `POST /api/v1/student-academics` with valid payload | 201, record created with `_data` populated |
| I02 | Get student's academic history | `GET /api/v1/students/{id}/academics` | 200, list ordered by academic_year + semester |
| I03 | Update academic record | `PUT /api/v1/student-academics/{id}` | 200, fields updated |
| I04 | `_data` includes student, academic_year, class_room | Create record with valid FKs | `_data` contains snapshots of all 3 relations |

### Integration Tests — Constraints

| # | Test Case | Action | Expected |
|---|---|---|---|
| I05 | Unique per student+year+semester | Create 2 records for same student, same year, same semester | 409 or 422, unique constraint violation |
| I06 | Same student, same year, different semester allowed | Create ganjil + genap for same student + year | 201, both created |
| I07 | Semester CHECK enforced | INSERT with `semester = 'midterm'` | DB error, CHECK violation |
| I08 | Promotion status CHECK enforced | INSERT with `promotion_status = 'failed'` | DB error, CHECK violation |
| I09 | Promotion status nullable | INSERT with `promotion_status = NULL` | 201, allowed (not yet evaluated) |

### Integration Tests — Attendance & Grades

| # | Test Case | Action | Expected |
|---|---|---|---|
| I10 | Attendance defaults to 0 | Create record without attendance fields | All `days_*` = 0 |
| I11 | Grade average nullable | Create record without `grade_average` | 201, `grade_average` = NULL |
| I12 | Class rank and total_students stored | Create record with `class_rank = 5`, `total_students = 32` | 200, values stored correctly |
| I13 | Filter by academic year | `GET /students/{id}/academics?academic_year_id={ayid}` | 200, returns only records for that year |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | StudentUpdated syncs to academic records | Update student `full_name` | `_data.student.full_name` updated in all related academic records |
| I15 | ClassRoomUpdated syncs to academic records | Update class_room name | `_data.class_room.name` updated in matching records |
| I16 | AcademicYearUpdated syncs to academic records | Update academic_year name | `_data.academic_year.name` updated in matching records |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Cannot access other tenant's academic records | GET with wrong tenant scope | 404 |
| I18 | Cannot create record for student in different company | POST with mismatched company_id | Error, scope violation |
