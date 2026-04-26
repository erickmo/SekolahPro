# ADR-012: Teachers & Staff (Guru dan Tenaga Kependidikan)

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Data guru dan staff adalah **dependency blocker** untuk implementasi student domains. Banyak domain student yang reference teacher/staff:

| Domain | Field | Konteks |
|--------|-------|---------|
| ADR-011 Class Rooms | `homeroom_teacher_id` | Wali kelas |
| S008 Attendance | `recorded_by` | Guru yang input absensi |
| S011 Subject Grades | `teacher_id` | Guru mapel pengajar |
| S012 Discipline | `reported_by`, `approved_by` | Pelapor dan approver |
| S015 Extracurricular | `coach_id` | Pembina ekskul |
| S017 Counseling | `counselor_id` | Guru BK |
| S018 Rapor | `teacher_signature_url` | Tanda tangan wali kelas |

Di Indonesia, personel sekolah terdiri dari:
- **Guru (GTK)**: Guru mata pelajaran, guru BK, guru ekskul.
- **Tenaga Kependidikan**: Admin TU, bendahara, pustakawan, satpam.
- **Kepala Sekolah**: Bisa merangkap guru — punya wewenang approval.
- **Wakil Kepala Sekolah**: Wakasek kurikulum, kesiswaan, sarana, humas.

### Mengapa Vernon Pattern?

- Referenced by 7+ domain student sebagai belongs_to.
- Read-heavy: nama guru muncul di rapor, absensi, nilai, dll.
- Relasi ke subjects (guru mapel), class_rooms (wali kelas).
- SyncEngine: perubahan nama guru harus propagate ke semua `_data`.

## Decision

Menggunakan **Vernon Pattern** untuk domain `teacher`. Satu tabel `teachers` menampung semua jenis personel sekolah (guru + staff) dengan `role` yang membedakan.

### Table Schema

```sql
CREATE TABLE teachers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Link ke user account (jika ada)
    user_id         UUID,

    -- Identitas
    nip             VARCHAR(18),
    nuptk           VARCHAR(16),
    full_name       VARCHAR(255) NOT NULL,
    gender          VARCHAR(1) NOT NULL,
    birth_place     VARCHAR(100),
    birth_date      DATE,
    religion        VARCHAR(20),
    phone           VARCHAR(20),
    email           VARCHAR(255),
    photo_url       TEXT,

    -- Kepegawaian
    employee_type   VARCHAR(20) NOT NULL,
    role            VARCHAR(30) NOT NULL,
    join_date       DATE NOT NULL,
    resign_date     DATE,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'active',

    -- Tanda tangan (untuk rapor)
    signature_url   TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_teacher_gender CHECK (gender IN ('L', 'P')),
    CONSTRAINT chk_teacher_religion CHECK (religion IS NULL OR religion IN ('islam', 'kristen', 'katolik', 'hindu', 'buddha', 'konghucu')),
    CONSTRAINT chk_employee_type CHECK (employee_type IN ('pns', 'p3k', 'honorer', 'yayasan', 'kontrak')),
    CONSTRAINT chk_teacher_role CHECK (role IN (
        'kepala_sekolah', 'wakasek_kurikulum', 'wakasek_kesiswaan',
        'wakasek_sarana', 'wakasek_humas',
        'guru_mapel', 'guru_bk', 'guru_piket',
        'admin_tu', 'bendahara', 'pustakawan',
        'staff_umum'
    )),
    CONSTRAINT chk_teacher_status CHECK (status IN ('active', 'on_leave', 'resigned', 'retired'))
);

-- Indexes
CREATE INDEX idx_teacher_tenant_company ON teachers (tenant_id, company_id);
CREATE INDEX idx_teacher_user ON teachers (user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_teacher_role ON teachers (role);
CREATE INDEX idx_teacher_status ON teachers (status);
CREATE INDEX idx_teacher_rels ON teachers USING GIN (_rels);
CREATE INDEX idx_teacher_data ON teachers USING GIN (_data);
CREATE UNIQUE INDEX uq_teacher_nip ON teachers (nip) WHERE nip IS NOT NULL;
CREATE UNIQUE INDEX uq_teacher_nuptk ON teachers (nuptk) WHERE nuptk IS NOT NULL;

-- Junction: Guru ↔ Mata Pelajaran (many-to-many)
CREATE TABLE teacher_subject_map (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,
    teacher_id      UUID NOT NULL,
    subject_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_teacher_subject_year UNIQUE (teacher_id, subject_id, academic_year_id)
);

-- Indexes
CREATE INDEX idx_teacher_subject_teacher ON teacher_subject_map (teacher_id);
CREATE INDEX idx_teacher_subject_subject ON teacher_subject_map (subject_id);
CREATE INDEX idx_teacher_subject_year ON teacher_subject_map (academic_year_id);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `nip` | VARCHAR(18), nullable, partial unique | Nomor Induk Pegawai — hanya PNS/P3K yang punya |
| `nuptk` | VARCHAR(16), nullable, partial unique | Nomor Unik Pendidik Tenaga Kependidikan — dari Kemendikbud |
| `user_id` | UUID, nullable | Link ke tabel users (ADR-013) — nullable jika staff belum punya akun |
| `employee_type` | 5 tipe | Status kepegawaian Indonesia: PNS, P3K, Honorer, Yayasan, Kontrak |
| `role` | 12 roles | Peran fungsional di sekolah — menentukan permissions |
| `signature_url` | TEXT, nullable | Gambar tanda tangan untuk rapor digital |
| `teacher_subject_map` | Junction table | Guru bisa mengajar beberapa mapel, satu mapel bisa diajar beberapa guru |

### Pesantren Terminology (ADR-009)

| General | Islamic | Mapping |
|---------|---------|---------|
| Guru | Ustadz/Ustadzah | Presentation layer |
| Kepala Sekolah | Mudir | `role = 'kepala_sekolah'` |
| Wali Kelas | Musyrif | `homeroom_teacher_id` di class_rooms |
| Guru BK | Murshid | `role = 'guru_bk'` |

### Vernon Relationships

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| (none) | — | — | Teachers adalah entity root — tidak punya FK ke entity lain |

Teachers menjadi **sumber _data** untuk banyak domain:

```json
// Contoh _data di domain lain
"_data": {
  "homeroom_teacher": { "id": "018f...", "full_name": "Bu Siti", "nip": "198501012010012001" }
}
```

### API Endpoints

```
GET    /api/v1/teachers                              — List guru/staff
GET    /api/v1/teachers?role=guru_mapel              — Filter by role
GET    /api/v1/teachers/{id}                         — Detail guru
POST   /api/v1/teachers                              — Tambah guru/staff
PUT    /api/v1/teachers/{id}                         — Update data
PUT    /api/v1/teachers/{id}/signature               — Upload tanda tangan

# Mapping guru ↔ mapel
GET    /api/v1/teachers/{id}/subjects                — Mapel yang diajar guru ini
POST   /api/v1/teacher-subject-map                   — Assign guru ke mapel
DELETE /api/v1/teacher-subject-map/{id}              — Remove assignment
GET    /api/v1/subjects/{id}/teachers                — Guru yang mengajar mapel ini
```

## Consequences

### Positive

- **Satu tabel**: Guru dan staff dalam satu tabel — query sederhana, role membedakan.
- **User-Teacher separation**: `user_id` terpisah dari teacher record — staff bisa ada tanpa akun.
- **Subject mapping per tahun**: Guru bisa mengajar mapel berbeda tiap tahun ajaran.
- **Signature**: Tanda tangan digital untuk rapor tanpa cetak-scan.
- **Dapodik aligned**: NIP + NUPTK + employee_type sesuai standar Kemendikbud.

### Negative / Trade-offs

- **Role static**: 12 roles hardcoded di CHECK constraint. Jika perlu role baru, ALTER TABLE diperlukan.
- **Tidak ada payroll**: ADR ini hanya data master guru — gaji, tunjangan, dll butuh domain terpisah.
- **Subject map per tahun**: Harus disetup ulang tiap tahun ajaran (bisa carry-forward seperti class rooms).
- **SyncEngine heavy**: Update nama guru propagate ke semua domain yang reference teacher.

## Alternatives Considered

### 1. Tabel terpisah untuk guru dan staff
- Ditolak: 80% field sama (identitas, kepegawaian) — satu tabel + role column lebih efisien.

### 2. Teacher sebagai extension dari users
- Ditolak: tidak semua teacher punya akun (staff honorer, guru part-time). Teacher = data HR, User = data auth.

### 3. Role sebagai tabel terpisah (RBAC)
- Deferred: untuk MVP, CHECK constraint cukup. Full RBAC di ADR-013 (Users & Roles).

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` | — | `"teachers"` |
| U02 | Validate rejects invalid `role` | `"janitor"` | Error |
| U03 | Validate rejects invalid `employee_type` | `"freelance"` | Error |
| U04 | Validate rejects invalid `status` | `"fired"` | Error |
| U05 | Validate rejects empty `full_name` | `""` | Error |
| U06 | Validate accepts valid teacher | All fields valid | No error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create teacher | POST with valid data | 201 |
| I02 | NIP partial unique | Create 2 teachers same NIP | 409/422 |
| I03 | NIP null allowed multiple | Create 2 teachers without NIP | 201 both |
| I04 | NUPTK partial unique | Create 2 teachers same NUPTK | 409/422 |
| I05 | Assign teacher to subject | POST teacher-subject-map | 201 |
| I06 | Unique per teacher+subject+year | Assign same mapping | 409/422 |
| I07 | Get teacher's subjects | GET /teachers/{id}/subjects | 200 |
| I08 | Get subject's teachers | GET /subjects/{id}/teachers | 200 |
| I09 | Role CHECK | INSERT with `role = 'janitor'` | DB error |
| I10 | Employee type CHECK | INSERT with `employee_type = 'freelance'` | DB error |
| I11 | SyncEngine: name change propagates | Update teacher name | All referencing `_data` updated |
| I12 | Tenant isolation | Access other tenant's teacher | 404 |
