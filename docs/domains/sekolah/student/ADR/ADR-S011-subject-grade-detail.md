# ADR-S011: Subject Grade Detail

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

ADR-S004 menyimpan `grade_average` sebagai agregat per semester, namun belum ada data **nilai per mata pelajaran** yang menjadi sumber data rata-rata tersebut. Nilai detail per mapel diperlukan untuk:

1. **Rapor**: Rapor menampilkan nilai per mata pelajaran, bukan hanya rata-rata.
2. **Kurikulum Merdeka**: Penilaian menggunakan aspek P (Pengetahuan), K (Keterampilan), dan Sikap — bukan hanya angka tunggal.
3. **Analisis akademik**: Identifikasi mata pelajaran yang perlu perhatian per siswa.
4. **Pelaporan Dapodik**: Nilai per mapel diperlukan untuk e-rapor Kemendikbud.
5. **Kenaikan kelas**: Keputusan promosi berdasarkan nilai per mapel, bukan hanya rata-rata.

Kurikulum Indonesia memiliki beberapa model penilaian:
- **KTSP/K13**: KI-1 (Spiritual), KI-2 (Sosial), KI-3 (Pengetahuan), KI-4 (Keterampilan)
- **Kurikulum Merdeka**: Asesmen Formatif + Asesmen Sumatif → Nilai Akhir
- **Pesantren**: Bisa menambahkan mata pelajaran keagamaan (Tahfidz, Fiqh, dll)

### Mengapa Vernon Pattern?

- has_many dari student (banyak nilai per siswa per semester).
- Read-heavy: rapor, dashboard, dan laporan.
- Relasi ke student, subject, academic_year, class_room.
- Business logic moderate: input nilai, hitung rata-rata, generate rapor.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `subjects` (master mata pelajaran) dan `student_grades` (nilai per siswa per mapel per semester).

### Table Schema

```sql
-- Master mata pelajaran
CREATE TABLE subjects (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    subject_group   VARCHAR(30) NOT NULL,

    -- Konfigurasi
    is_national     BOOLEAN NOT NULL DEFAULT false,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    sort_order      INT NOT NULL DEFAULT 0,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_subject_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_subject_group CHECK (subject_group IN (
        'agama', 'pkn', 'bahasa', 'matematika', 'ipa', 'ips',
        'seni_budaya', 'pjok', 'prakarya', 'muatan_lokal',
        'keagamaan', 'tahfidz', 'other'
    ))
);

-- Indexes
CREATE INDEX idx_subject_tenant_company ON subjects (tenant_id, company_id);

-- Nilai per siswa per mata pelajaran per semester
CREATE TABLE student_grades (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    subject_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    class_room_id   UUID NOT NULL,

    -- Periode
    semester        VARCHAR(10) NOT NULL,

    -- Nilai (skala 0-100)
    score_knowledge  NUMERIC(5,2),
    score_skill      NUMERIC(5,2),
    score_attitude   VARCHAR(2),
    final_score      NUMERIC(5,2),

    -- Deskripsi (untuk rapor naratif Kurikulum Merdeka)
    description_knowledge TEXT,
    description_skill     TEXT,

    -- Predikat
    grade_letter    VARCHAR(2),

    -- Guru pengajar
    teacher_id      UUID,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_grade_student_subject_semester UNIQUE (student_id, subject_id, academic_year_id, semester),
    CONSTRAINT chk_grade_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_score_knowledge CHECK (score_knowledge IS NULL OR (score_knowledge >= 0 AND score_knowledge <= 100)),
    CONSTRAINT chk_score_skill CHECK (score_skill IS NULL OR (score_skill >= 0 AND score_skill <= 100)),
    CONSTRAINT chk_score_attitude CHECK (score_attitude IS NULL OR score_attitude IN ('SB', 'B', 'C', 'K')),
    CONSTRAINT chk_final_score CHECK (final_score IS NULL OR (final_score >= 0 AND final_score <= 100)),
    CONSTRAINT chk_grade_letter CHECK (grade_letter IS NULL OR grade_letter IN ('A', 'B', 'C', 'D', 'E'))
);

-- Indexes
CREATE INDEX idx_grade_tenant_company ON student_grades (tenant_id, company_id);
CREATE INDEX idx_grade_student ON student_grades (student_id);
CREATE INDEX idx_grade_subject ON student_grades (subject_id);
CREATE INDEX idx_grade_student_semester ON student_grades (student_id, academic_year_id, semester);
CREATE INDEX idx_grade_class_semester ON student_grades (class_room_id, academic_year_id, semester);
CREATE INDEX idx_grade_teacher ON student_grades (teacher_id);
CREATE INDEX idx_grade_rels ON student_grades USING GIN (_rels);
CREATE INDEX idx_grade_data ON student_grades USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `score_knowledge` | NUMERIC(5,2), nullable | Nilai Pengetahuan (KI-3 / Asesmen Sumatif), 0-100 |
| `score_skill` | NUMERIC(5,2), nullable | Nilai Keterampilan (KI-4 / Asesmen Formatif), 0-100 |
| `score_attitude` | VARCHAR(2), CHECK | Nilai Sikap: SB (Sangat Baik), B, C, K (Kurang) — bukan angka |
| `final_score` | NUMERIC(5,2), nullable | Nilai akhir yang masuk rapor — bisa dihitung atau manual |
| `description_*` | TEXT, nullable | Deskripsi naratif untuk Kurikulum Merdeka |
| `grade_letter` | VARCHAR(2), CHECK | Predikat huruf A-E |
| `teacher_id` | UUID, nullable | Guru pengajar mapel — untuk atribusi |
| `subject_group` | VARCHAR(30), CHECK | Kelompok mapel termasuk `keagamaan` dan `tahfidz` untuk pesantren (ADR-009) |

### Vernon Relationships

**student_grades:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Pemilik nilai |
| `subject` | belongs_to | **Ya** | Nama mata pelajaran |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |
| `class_room` | belongs_to | **Ya** | Kelas |

### _rels / _data Structure

```json
{
  "_rels": {
    "student_id": "018f...",
    "subject_id": "018f...",
    "academic_year_id": "018f...",
    "class_room_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" },
    "subject": { "id": "018f...", "name": "Matematika", "code": "MTK", "subject_group": "matematika" },
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "class_room": { "id": "018f...", "name": "VII-A" }
  }
}
```

### API Endpoints

```
# Subjects (Master)
GET    /api/v1/subjects                             — List mata pelajaran
POST   /api/v1/subjects                             — Buat mata pelajaran
PUT    /api/v1/subjects/{id}                        — Update mata pelajaran

# Grades
GET    /api/v1/students/{id}/grades                 — Nilai siswa (all semesters)
GET    /api/v1/students/{id}/grades?semester=ganjil&year_id={id} — Nilai per semester
POST   /api/v1/student-grades/bulk                  — Bulk input nilai (per mapel per kelas)
PUT    /api/v1/student-grades/{id}                  — Update nilai individual
GET    /api/v1/class-rooms/{id}/grades/{semester}   — Nilai seluruh kelas per semester (rekap guru)
```

### Aggregation to S004

Ketika nilai per mapel sudah lengkap, sistem menghitung `grade_average` untuk S004:

```sql
SELECT
    student_id,
    AVG(final_score) AS grade_average
FROM student_grades
WHERE academic_year_id = $1 AND semester = $2 AND final_score IS NOT NULL
GROUP BY student_id;
```

Hasil di-update ke `student_academics.grade_average` via event `GradeAggregated`.

## Consequences

### Positive

- **Rapor ready**: Nilai per mapel lengkap dengan knowledge, skill, attitude, dan deskripsi naratif.
- **Multi-kurikulum**: Mendukung K13 (KI-1 s/d KI-4) dan Kurikulum Merdeka (formatif + sumatif).
- **Pesantren support**: Subject group `keagamaan` dan `tahfidz` mendukung ADR-009.
- **Bulk input**: Guru bisa input nilai seluruh kelas per mapel sekaligus.
- **Sumber kebenaran**: S004 `grade_average` dihitung dari data ini, bukan manual.

### Negative / Trade-offs

- **Tidak ada assignment tracking**: ADR ini menyimpan nilai akhir per semester, bukan nilai per tugas/ulangan. Assignment tracking bisa jadi domain terpisah.
- **Attitude non-numeric**: `score_attitude` sebagai huruf (SB/B/C/K) tidak bisa di-average — harus ditampilkan terpisah.
- **Volume data**: 30 siswa × 12 mapel × 2 semester = 720 rows per kelas per tahun.
- **Grade letter mapping**: A/B/C/D/E mapping ke range nilai (A >= 90, dll) harus di-configure per sekolah.

## Alternatives Considered

### 1. Semua nilai sebagai JSONB di S004
- Ditolak: tidak bisa query per mapel, tidak bisa filter guru, tidak bisa bulk input per mapel.

### 2. Satu kolom `score` saja
- Ditolak: kurikulum Indonesia mengharuskan pemisahan knowledge/skill/attitude.

### 3. Denormalisasi score components ke JSONB
- Ditolak: score_knowledge dan score_skill perlu di-query dan di-aggregate — JSONB path query kurang efisien.

## Test Cases

### Unit Tests — Descriptor & Validation

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `SubjectDescriptor.TableName()` | — | `"subjects"` |
| U02 | `GradeDescriptor.TableName()` | — | `"student_grades"` |
| U03 | Validate rejects invalid `subject_group` | `"science"` | Error: invalid subject_group |
| U04 | Validate rejects score > 100 | `score_knowledge = 101` | Error: score out of range |
| U05 | Validate rejects score < 0 | `score_knowledge = -1` | Error: score out of range |
| U06 | Validate rejects invalid `score_attitude` | `"A"` | Error: must be SB/B/C/K |
| U07 | Validate rejects invalid `grade_letter` | `"F"` | Error: must be A-E |
| U08 | Validate accepts all nullable scores | All scores null | No error |
| U09 | Validate accepts valid grade | All fields valid | No error |

### Integration Tests — Subjects

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create subject | POST with name + code + group | 201 |
| I02 | Unique code per company | Create 2 with same code | 409/422 |
| I03 | Subject group CHECK | INSERT with `subject_group = 'science'` | DB error |

### Integration Tests — Grades

| # | Test Case | Action | Expected |
|---|---|---|---|
| I04 | Bulk input grades | POST /student-grades/bulk for 30 students | 201, 30 records |
| I05 | Unique per student+subject+year+semester | Insert same grade twice | 409/422 |
| I06 | Get student grades per semester | GET /students/{id}/grades?semester=ganjil | 200, all subjects |
| I07 | Get class grades | GET /class-rooms/{id}/grades/ganjil | 200, matrix (students × subjects) |
| I08 | Update individual grade | PUT /student-grades/{id} | 200, updated |
| I09 | Score range CHECK | INSERT with `score_knowledge = 150` | DB error |
| I10 | Attitude CHECK | INSERT with `score_attitude = 'A'` | DB error |
| I11 | Grade letter CHECK | INSERT with `grade_letter = 'F'` | DB error |
| I12 | Nullable scores allowed | INSERT all scores NULL | 201 |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | StudentUpdated syncs to grades | Update student name | `_data.student.full_name` updated |
| I14 | SubjectUpdated syncs to grades | Update subject name | `_data.subject.name` updated |

### Integration Tests — Aggregation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I15 | Grade average calculated | Insert grades for 5 subjects | Average matches manual calculation |
| I16 | Null scores excluded from average | 3 subjects with score, 2 null | Average from 3 only |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Cannot access other tenant's grades | GET with wrong tenant | 404 |
