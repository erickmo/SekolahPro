# ADR-S003: Student Guardian (Orang Tua/Wali)

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Data orang tua/wali siswa adalah komponen wajib dalam administrasi sekolah Indonesia. Dibutuhkan untuk:

1. **Rapor**: Nama ayah/ibu wajib tercantum di rapor.
2. **Komunikasi**: Nomor HP dan email untuk notifikasi dan pengumuman.
3. **Data Dapodik**: Pelaporan ke Kemendikbud memerlukan data orang tua lengkap.
4. **Beasiswa**: Penghasilan orang tua menentukan eligibilitas bantuan.
5. **Darurat**: Kontak darurat saat siswa sakit atau kecelakaan.

Satu siswa bisa memiliki **lebih dari satu guardian** (ayah, ibu, wali), dan satu guardian bisa menjadi wali dari **lebih dari satu siswa** (kakak-adik). Ini adalah relasi many-to-many.

### Mengapa Vernon Pattern?

- Relasi ke student (many-to-many melalui junction table).
- Read-heavy: data guardian diakses setiap kali melihat profil siswa lengkap.
- Business logic sederhana: CRUD.
- Eventual consistency acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk domain `student_guardian` dengan junction table `student_guardian_map` untuk relasi many-to-many.

### Table Schema

```sql
-- Data guardian (orang tua/wali)
CREATE TABLE student_guardians (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    full_name       VARCHAR(255) NOT NULL,
    nik             VARCHAR(16),
    gender          VARCHAR(1) NOT NULL,
    birth_place     VARCHAR(100),
    birth_date      DATE,
    religion        VARCHAR(20),
    phone           VARCHAR(20),
    email           VARCHAR(255),

    -- Pekerjaan & ekonomi
    occupation      VARCHAR(100),
    income_range    VARCHAR(20),
    education       VARCHAR(30),

    -- Status
    is_alive        BOOLEAN NOT NULL DEFAULT true,

    -- Alamat
    address         TEXT,
    rt              VARCHAR(3),
    rw              VARCHAR(3),
    village         VARCHAR(100),
    district        VARCHAR(100),
    city            VARCHAR(100),
    province        VARCHAR(100),
    postal_code     VARCHAR(5),

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_guardians_gender CHECK (gender IN ('L', 'P')),
    CONSTRAINT chk_guardians_income CHECK (income_range IN ('lt_1m', '1m_3m', '3m_5m', '5m_10m', 'gt_10m')),
    CONSTRAINT chk_guardians_education CHECK (education IN ('tidak_sekolah', 'sd', 'smp', 'sma', 'd1', 'd2', 'd3', 's1', 's2', 's3'))
);

-- Junction table: student <-> guardian (many-to-many)
CREATE TABLE student_guardian_map (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,
    student_id      UUID NOT NULL,
    guardian_id     UUID NOT NULL,
    relationship    VARCHAR(20) NOT NULL,
    is_primary      BOOLEAN NOT NULL DEFAULT false,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_student_guardian UNIQUE (student_id, guardian_id),
    CONSTRAINT chk_relationship CHECK (relationship IN ('father', 'mother', 'guardian'))
);

-- Indexes
CREATE INDEX idx_guardians_tenant_company ON student_guardians (tenant_id, company_id);
CREATE INDEX idx_guardians_rels ON student_guardians USING GIN (_rels);
CREATE INDEX idx_guardians_data ON student_guardians USING GIN (_data);
CREATE INDEX idx_guardian_map_student ON student_guardian_map (student_id);
CREATE INDEX idx_guardian_map_guardian ON student_guardian_map (guardian_id);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `nik` | VARCHAR(16), nullable | NIK KTP 16 digit, nullable karena tidak semua wali punya KTP |
| `income_range` | VARCHAR(20), CHECK | Range diskrit untuk privasi, bukan angka eksak |
| `education` | VARCHAR(30), CHECK | Jenjang pendidikan standar Indonesia |
| `is_alive` | BOOLEAN | Diperlukan Dapodik — mempengaruhi status wali |
| `address` fields | Terpisah per level | Mengikuti format alamat Indonesia (RT/RW/Kel/Kec/Kota/Prov) |
| `is_primary` | di junction table | Menandai guardian utama untuk kontak prioritas |

### API Endpoints

```
GET    /api/v1/students/{id}/guardians          — List guardians siswa
POST   /api/v1/students/{id}/guardians          — Tambah guardian ke siswa
DELETE /api/v1/students/{id}/guardians/{gid}     — Lepas guardian dari siswa
GET    /api/v1/student-guardians                 — List semua guardians (admin)
POST   /api/v1/student-guardians                 — Buat guardian baru
PUT    /api/v1/student-guardians/{id}            — Update guardian
```

## Consequences

### Positive

- **Reusable guardian**: Satu guardian bisa di-link ke beberapa siswa (kakak-adik) tanpa duplikasi data.
- **Dapodik compliant**: Field set mencakup semua data yang dibutuhkan pelaporan Kemendikbud.
- **Privasi**: Income sebagai range, bukan angka eksak.
- **Alamat terstruktur**: Mendukung filter/reporting per wilayah.

### Negative / Trade-offs

- **Junction table complexity**: Many-to-many membutuhkan tabel tambahan dan query lebih kompleks.
- **Guardian lookup**: Mencari guardian berdasarkan student memerlukan JOIN melalui junction table.
- **Orphan guardians**: Guardian tanpa student link bisa terjadi jika siswa dihapus tanpa cleanup.

## Alternatives Considered

### 1. Guardian sebagai kolom di Student Table
- Ditolak: maksimal 1 guardian per kolom set, tidak bisa handle ayah + ibu + wali.

### 2. Guardian sebagai array JSONB di Student
- Ditolak: tidak bisa reuse guardian untuk kakak-adik, sulit di-query dan di-index.

### 3. Simple has_many (tanpa junction)
- Ditolak: satu orang tua dengan 3 anak harus diinput 3 kali — duplikasi data.

## Test Cases

### Unit Tests — Descriptor & Validation

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` returns correct name | — | `"student_guardians"` |
| U02 | Validate rejects empty `full_name` | `{ "full_name": "" }` | Error: full_name required |
| U03 | Validate rejects invalid `gender` | `{ ..., "gender": "X" }` | Error: gender must be L or P |
| U04 | Validate rejects invalid `income_range` | `{ ..., "income_range": "rich" }` | Error: invalid income_range |
| U05 | Validate rejects invalid `education` | `{ ..., "education": "phd" }` | Error: invalid education |
| U06 | Validate accepts valid guardian with minimal fields | `full_name` + `gender` only | No error |
| U07 | Validate accepts all nullable fields as nil | `nik`, `phone`, `email`, `occupation` = nil | No error |

### Integration Tests — CRUD Guardian

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create guardian | `POST /api/v1/student-guardians` with valid payload | 201, guardian created |
| I02 | Update guardian | `PUT /api/v1/student-guardians/{id}` | 200, fields updated |
| I03 | List all guardians (admin) | `GET /api/v1/student-guardians` | 200, paginated list scoped by tenant + company |
| I04 | NIK format accepted (16 digits) | Create guardian with `nik = "3201234567890001"` | 201, created |
| I05 | Gender CHECK enforced | INSERT with `gender = 'M'` | DB error, CHECK violation |
| I06 | Income range CHECK enforced | INSERT with `income_range = 'invalid'` | DB error, CHECK violation |
| I07 | Education CHECK enforced | INSERT with `education = 'phd'` | DB error, CHECK violation |

### Integration Tests — Many-to-Many (Junction Table)

| # | Test Case | Action | Expected |
|---|---|---|---|
| I08 | Link guardian to student | `POST /api/v1/students/{id}/guardians` with `guardian_id` + `relationship` | 201, mapping created in `student_guardian_map` |
| I09 | List guardians for student | `GET /api/v1/students/{id}/guardians` | 200, returns guardians linked to this student |
| I10 | Unlink guardian from student | `DELETE /api/v1/students/{id}/guardians/{gid}` | 200, mapping removed, guardian data preserved |
| I11 | Same guardian linked to 2 students (siblings) | Link guardian to student A and student B | Both mappings exist, guardian data shared |
| I12 | Duplicate link rejected | Link same guardian to same student twice | 409 or 422, unique constraint on `(student_id, guardian_id)` |
| I13 | Relationship CHECK enforced | Create mapping with `relationship = "uncle"` | DB error, CHECK violation (only father/mother/guardian) |
| I14 | Mark primary guardian | Create mapping with `is_primary = true` | 201, `is_primary` set correctly |
| I15 | List returns relationship type | `GET /students/{id}/guardians` | Each item includes `relationship` (father/mother/guardian) and `is_primary` |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Cannot access other tenant's guardian | GET guardian from tenant A using tenant B scope | 404 |
| I17 | Cannot link cross-tenant student-guardian | Link guardian from tenant A to student from tenant B | Error, foreign key or scope violation |
