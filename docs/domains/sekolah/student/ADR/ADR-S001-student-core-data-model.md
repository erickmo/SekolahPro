# ADR-S001: Student Core Data Model

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

SekolahPro membutuhkan pengelolaan data siswa sebagai entitas utama sistem. Data siswa bersifat **read-heavy** — dashboard admin, laporan, pencarian siswa, dan tampilan profil diakses jauh lebih sering dibanding operasi write (pendaftaran, update biodata).

Karakteristik domain `student`:

1. **Banyak relasi**: Siswa terhubung ke kelas, tahun ajaran, orang tua/wali, alamat, riwayat akademik, kesehatan — minimal 5+ relasi.
2. **Read:Write ratio tinggi**: Estimasi 20:1. Dashboard siswa, absensi harian, dan laporan memerlukan read cepat dengan data relasi lengkap.
3. **Data referensi nasional**: NIS (per sekolah) dan NISN (nasional) menjadi identifier penting selain UUID internal.
4. **Lifecycle tracking**: Siswa memiliki status lifecycle (aktif, lulus, pindah, keluar, cuti) yang harus dilacak dengan tanggal dan alasan.
5. **Multi-tenant**: Data siswa harus di-scope per `tenant_id` + `company_id` sesuai ADR-004.

### Mengapa Vernon Pattern?

Berdasarkan decision tree di ADR-003:

- JOIN >= 3: Ya (kelas, tahun ajaran, orang tua, alamat, akademik)
- Read:Write >= 10:1: Ya (dashboard dan laporan dominan)
- Business logic: Sederhana (CRUD + status transition)
- Eventual consistency: Acceptable (data profil tidak memerlukan real-time consistency)

## Decision

Menggunakan **Vernon Pattern** untuk domain `student` dengan schema 10-kolom standar ditambah kolom bisnis spesifik.

### Table Schema

```sql
CREATE TABLE students (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas sekolah & nasional
    nis             VARCHAR(20) NOT NULL,
    nisn            VARCHAR(10),

    -- Biodata
    full_name       VARCHAR(255) NOT NULL,
    nickname        VARCHAR(100),
    gender          VARCHAR(1) NOT NULL,
    birth_place     VARCHAR(100) NOT NULL,
    birth_date      DATE NOT NULL,
    religion        VARCHAR(20) NOT NULL,
    blood_type      VARCHAR(2),
    photo_url       TEXT,

    -- Status & lifecycle
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    entry_date      DATE NOT NULL,
    entry_type      VARCHAR(20) NOT NULL DEFAULT 'new',
    exit_date       DATE,
    exit_reason     TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_students_nis UNIQUE (tenant_id, company_id, nis),
    CONSTRAINT chk_students_gender CHECK (gender IN ('L', 'P')),
    CONSTRAINT chk_students_status CHECK (status IN ('active', 'graduated', 'transferred', 'expelled', 'on_leave')),
    CONSTRAINT chk_students_entry_type CHECK (entry_type IN ('new', 'transfer', 'returning')),
    CONSTRAINT chk_students_religion CHECK (religion IN ('islam', 'kristen', 'katolik', 'hindu', 'buddha', 'konghucu'))
);

CREATE UNIQUE INDEX uq_students_nisn ON students (nisn) WHERE nisn IS NOT NULL;
CREATE INDEX idx_students_tenant_company ON students (tenant_id, company_id);
CREATE INDEX idx_students_status ON students (tenant_id, company_id, status);
CREATE INDEX idx_students_name ON students (tenant_id, company_id, full_name);
CREATE INDEX idx_students_rels ON students USING GIN (_rels);
CREATE INDEX idx_students_data ON students USING GIN (_data);
CREATE INDEX idx_students_deleted_at ON students (deleted_at) WHERE deleted_at IS NOT NULL;
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `nis` | VARCHAR(20), NOT NULL | Wajib — diberikan sekolah saat pendaftaran |
| `nisn` | VARCHAR(10), nullable | Tidak semua siswa punya NISN (siswa baru, sekolah informal) |
| `gender` | VARCHAR(1), CHECK | Standar Kemendikbud: L/P |
| `religion` | VARCHAR(20), CHECK | 6 agama resmi Indonesia — dibutuhkan untuk rapor |
| `status` | VARCHAR(20), CHECK | 5 status lifecycle sekolah Indonesia |
| `entry_type` | VARCHAR(20), CHECK | Membedakan siswa baru, pindahan, dan kembali |
| `photo_url` | TEXT, nullable | URL ke object storage, bukan blob di database |
| `exit_date` + `exit_reason` | nullable | Hanya diisi saat siswa tidak aktif |

### Vernon Descriptor

```go
package student

type Descriptor struct{}

func (d *Descriptor) TableName() string { return "students" }

func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
    return map[string]vernon.RelDef{
        "class_room": {
            Table:    "class_rooms",
            Type:     vernon.BelongsTo,
            Autoload: true,
            Fields:   []string{"id", "name", "grade_level"},
        },
        "academic_year": {
            Table:    "academic_years",
            Type:     vernon.BelongsTo,
            Autoload: true,
            Fields:   []string{"id", "name", "is_active"},
        },
    }
}
```

## Consequences

### Positive

- **Dashboard cepat**: Satu query tanpa JOIN menampilkan data siswa + kelas + tahun ajaran dari `_data`.
- **Identifier ganda**: NIS (lokal) dan NISN (nasional) mendukung pelaporan Kemendikbud.
- **Lifecycle jelas**: Status + entry/exit tracking memberikan audit trail lengkap.
- **Soft delete**: `deleted_at` memungkinkan pemulihan data siswa yang terhapus.

### Negative / Trade-offs

- **Status transition**: Tidak ada state machine enforcement di database — harus di-enforce di application layer.
- **NISN uniqueness**: Partial unique index bergantung pada data input yang benar.
- **Photo storage**: Mengasumsikan ada object storage terpisah yang belum di-define.

## Alternatives Considered

### 1. CQRS Pattern (tanpa Vernon)
- Ditolak: student punya 5+ relasi, dashboard membutuhkan query tanpa JOIN.

### 2. Embedded Guardian di Student Table
- Ditolak: satu siswa bisa punya 2-3 guardian, flat columns melanggar 1NF.

### 3. NISN sebagai Primary Key
- Ditolak: tidak semua siswa punya NISN, dan NISN bisa berubah. UUID v7 lebih stabil.

## Test Cases

### Unit Tests — Descriptor & Validation

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` returns correct name | — | `"students"` |
| U02 | `Descriptor.DefaultRels()` returns 2 autoloaded rels | — | `class_room` (autoload: true), `academic_year` (autoload: true) |
| U03 | Validate rejects empty `nis` | `{ "full_name": "Ahmad", "nis": "" }` | Error: nis required |
| U04 | Validate rejects empty `full_name` | `{ "nis": "12345", "full_name": "" }` | Error: full_name required |
| U05 | Validate rejects invalid `gender` | `{ ..., "gender": "X" }` | Error: gender must be L or P |
| U06 | Validate rejects invalid `status` | `{ ..., "status": "unknown" }` | Error: invalid status |
| U07 | Validate rejects invalid `entry_type` | `{ ..., "entry_type": "invalid" }` | Error: invalid entry_type |
| U08 | Validate rejects invalid `religion` | `{ ..., "religion": "other" }` | Error: invalid religion |
| U09 | Validate accepts valid complete student | All required fields valid | No error |
| U10 | Validate accepts nullable fields as nil | `nisn`, `nickname`, `blood_type`, `photo_url` = nil | No error |

### Integration Tests — Database & API

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create student with all required fields | `POST /api/v1/students` with valid payload | 201, student created with UUID v7 |
| I02 | Create student — NIS unique per company | Create 2 students with same NIS in same company | 409 or 422, unique constraint violation |
| I03 | Create student — NIS unique across companies | Create 2 students with same NIS in different companies | 201, both created successfully |
| I04 | Create student — NISN globally unique | Create 2 students with same NISN | 409 or 422, unique constraint violation |
| I05 | Create student — NISN null allowed for multiple | Create 2 students without NISN | 201, both created (partial index allows multiple nulls) |
| I06 | Get student by ID | `GET /api/v1/students/{id}` | 200, includes `_data` with class_room and academic_year |
| I07 | List students with pagination | `GET /api/v1/students?page=1&per_page=10` | 200, paginated response with `_data` (zero JOIN) |
| I08 | Update student biodata | `PUT /api/v1/students/{id}` with updated fields | 200, fields updated, `updated_at` changed |
| I09 | Update student status lifecycle | Update status from `active` to `graduated` with `exit_date` | 200, status + exit_date updated |
| I10 | Soft delete student | `DELETE /api/v1/students/{id}` | 200, `deleted_at` set, student excluded from list queries |
| I11 | Tenant isolation — cannot access other tenant's student | GET student from tenant A using tenant B scope | 404 |
| I12 | Company isolation — cannot access other company's student | GET student from company A using company B scope | 404 |
| I13 | Search student by name | `GET /api/v1/students?search=Ahmad` | 200, filtered results matching name |
| I14 | Filter students by status | `GET /api/v1/students?status=active` | 200, only active students returned |
| I15 | Gender CHECK constraint enforced | INSERT with `gender = 'X'` | DB error, CHECK violation |
| I16 | Religion CHECK constraint enforced | INSERT with `religion = 'other'` | DB error, CHECK violation |
