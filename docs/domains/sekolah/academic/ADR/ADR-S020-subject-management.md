# ADR-S020: Subject Management (Manajemen Mata Pelajaran)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

ADR-S011 mendefinisikan tabel `subjects` secara minimal sebagai bagian dari domain nilai. Namun seiring bertambahnya domain yang bergantung pada mata pelajaran (S019 Curriculum, S021 Schedule, S022 Assessment, S024 Lesson Plan, S025 Teaching Journal), diperlukan **domain mata pelajaran yang lebih komprehensif**.

Mata pelajaran di Indonesia memiliki karakteristik:

1. **Kelompok mapel nasional**: Ditetapkan Kemendikbud — Agama, PKn, Bahasa Indonesia, Matematika, IPA, IPS, Seni Budaya, PJOK, Prakarya.
2. **Muatan Lokal**: Ditentukan per sekolah/daerah — Bahasa Jawa, Bahasa Sunda, dll.
3. **Mapel Pesantren** (ADR-009): Fiqh, Aqidah Akhlak, Quran Hadits, SKI, Bahasa Arab, Tahfidz, Kitab Kuning, dll.
4. **Jam pelajaran**: Setiap mapel punya alokasi jam per minggu yang berbeda per jenjang.
5. **Per tahun ajaran**: Konfigurasi mapel bisa berubah tiap tahun ajaran (mapel baru, jam berubah).
6. **Per kurikulum**: Kurikulum Merdeka dan K13 memiliki struktur mapel yang sedikit berbeda.

Tabel `subjects` di S011 perlu diperluas dengan: subject groups, credit hours per grade, kurikulum linkage, dan konfigurasi per tahun ajaran.

### Mengapa Vernon Pattern?

- Referenced by 8+ domain: grades (S011), schedule (S021), assessment (S022), lesson plan (S024), teaching journal (S025), teacher mapping (ADR-012), curriculum (S019), rapor (S018).
- Read-heavy: nama dan konfigurasi mapel dibaca di hampir semua operasi akademik.
- Master data stabil — jarang berubah dalam satu tahun ajaran.
- SyncEngine: perubahan nama mapel harus propagate ke semua `_data`.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `subjects` (master mata pelajaran — evolusi dari S011) dan `subject_configurations` (konfigurasi per tahun ajaran per jenjang).

### Table Schema

```sql
-- Master mata pelajaran (evolusi dari S011)
CREATE TABLE subjects (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    name_en         VARCHAR(100),

    -- Klasifikasi
    subject_group   VARCHAR(30) NOT NULL,
    category        VARCHAR(20) NOT NULL,

    -- Konfigurasi umum
    is_national     BOOLEAN NOT NULL DEFAULT false,
    is_scored       BOOLEAN NOT NULL DEFAULT true,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    sort_order      INT NOT NULL DEFAULT 0,
    description     TEXT,

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
        'seni_budaya', 'pjok', 'prakarya', 'informatika',
        'muatan_lokal', 'keagamaan', 'tahfidz',
        'lintas_minat', 'peminatan', 'other'
    )),
    CONSTRAINT chk_subject_category CHECK (category IN (
        'wajib_nasional', 'wajib_lokal', 'pilihan',
        'diniyah', 'ekstra_kurikuler'
    ))
);

-- Indexes
CREATE INDEX idx_subject_tenant_company ON subjects (tenant_id, company_id);
CREATE INDEX idx_subject_group ON subjects (subject_group);
CREATE INDEX idx_subject_category ON subjects (category);
CREATE INDEX idx_subject_active ON subjects (is_active) WHERE is_active = true;
CREATE INDEX idx_subject_rels ON subjects USING GIN (_rels);
CREATE INDEX idx_subject_data ON subjects USING GIN (_data);

-- Konfigurasi mapel per tahun ajaran per jenjang
CREATE TABLE subject_configurations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    subject_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    curriculum_id   UUID,

    -- Scope
    grade_level     VARCHAR(5) NOT NULL,

    -- Jam pelajaran
    credit_hours_per_week INT NOT NULL DEFAULT 2,

    -- Bobot penilaian (persentase)
    weight_knowledge INT NOT NULL DEFAULT 50,
    weight_skill     INT NOT NULL DEFAULT 50,

    -- KKM / KKTP (Kriteria Ketuntasan)
    passing_grade   NUMERIC(5,2) NOT NULL DEFAULT 70.00,

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

    CONSTRAINT uq_subject_config UNIQUE (subject_id, academic_year_id, grade_level),
    CONSTRAINT chk_config_grade CHECK (grade_level IN ('1','2','3','4','5','6','7','8','9','10','11','12')),
    CONSTRAINT chk_credit_hours CHECK (credit_hours_per_week >= 1 AND credit_hours_per_week <= 12),
    CONSTRAINT chk_weight_total CHECK (weight_knowledge + weight_skill = 100),
    CONSTRAINT chk_passing_grade CHECK (passing_grade >= 0 AND passing_grade <= 100)
);

-- Indexes
CREATE INDEX idx_subj_cfg_tenant_company ON subject_configurations (tenant_id, company_id);
CREATE INDEX idx_subj_cfg_subject ON subject_configurations (subject_id);
CREATE INDEX idx_subj_cfg_year ON subject_configurations (academic_year_id);
CREATE INDEX idx_subj_cfg_curriculum ON subject_configurations (curriculum_id);
CREATE INDEX idx_subj_cfg_grade ON subject_configurations (grade_level);
CREATE INDEX idx_subj_cfg_rels ON subject_configurations USING GIN (_rels);
CREATE INDEX idx_subj_cfg_data ON subject_configurations USING GIN (_data);
```

### Field Design Rationale

**subjects:**

| Field | Keputusan | Alasan |
|---|---|---|
| `subject_group` | 16 groups | Kelompok standar Kemendikbud + keagamaan/tahfidz untuk pesantren (ADR-009) + informatika (mapel baru) |
| `category` | 5 kategori | wajib_nasional (dari Kemendikbud), wajib_lokal (muatan lokal), pilihan (lintas minat SMA), diniyah (pesantren), ekstra_kurikuler |
| `is_national` | BOOLEAN | True jika mapel ditetapkan Kemendikbud — false untuk muatan lokal dan diniyah |
| `is_scored` | BOOLEAN | Beberapa mapel non-scored (e.g. P5 project) — hanya deskripsi naratif |
| `name_en` | VARCHAR(100), nullable | Nama dalam bahasa Inggris untuk rapor bilingual |

**subject_configurations:**

| Field | Keputusan | Alasan |
|---|---|---|
| `credit_hours_per_week` | INT, 1-12 | Jam pelajaran per minggu — input utama untuk jadwal (S021) |
| `weight_knowledge` / `weight_skill` | INT, total 100% | Bobot penilaian pengetahuan vs keterampilan — Kurikulum Merdeka: 50/50, K13 bisa beda |
| `passing_grade` | NUMERIC(5,2) | KKM (K13) atau KKTP (Merdeka) — batas ketuntasan per mapel |
| `curriculum_id` | UUID, nullable | Opsional link ke kurikulum (S019) — nullable untuk mapel yang tidak terikat kurikulum spesifik |

### Daftar Mapel Pesantren (ADR-009)

| Kode | Nama | Group | Category |
|------|------|-------|----------|
| FQH | Fiqh | keagamaan | diniyah |
| AQD | Aqidah Akhlak | keagamaan | diniyah |
| QHD | Quran Hadits | keagamaan | diniyah |
| SKI | Sejarah Kebudayaan Islam | keagamaan | diniyah |
| ARB | Bahasa Arab | keagamaan | diniyah |
| THF | Tahfidz Al-Quran | tahfidz | diniyah |
| KTB | Kitab Kuning | keagamaan | diniyah |
| NHW | Nahwu Shorof | keagamaan | diniyah |

### Vernon Relationships

**subjects:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| (none) | — | — | Subjects adalah entity root — tidak punya FK ke entity lain |

**subject_configurations:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `subject` | belongs_to | **Ya** | Nama dan kode mapel |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |
| `curriculum` | belongs_to | **Tidak** | Opsional — load on demand |

### _rels / _data Structure

**subjects:**
```json
{
  "_rels": {},
  "_data": {}
}
```

**subject_configurations:**
```json
{
  "_rels": {
    "subject_id": "018f...",
    "academic_year_id": "018f...",
    "curriculum_id": "018f..."
  },
  "_data": {
    "subject": { "id": "018f...", "name": "Matematika", "code": "MTK", "subject_group": "matematika" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

### API Endpoints

```
# Subjects (Master)
GET    /api/v1/subjects                                  — List mata pelajaran
GET    /api/v1/subjects?group=keagamaan&category=diniyah  — Filter by group + category
GET    /api/v1/subjects/{id}                             — Detail mapel
POST   /api/v1/subjects                                  — Buat mapel
PUT    /api/v1/subjects/{id}                             — Update mapel

# Subject Configurations
GET    /api/v1/subject-configurations?year_id={id}&grade_level=7  — Konfigurasi per tahun + jenjang
POST   /api/v1/subject-configurations                    — Buat konfigurasi
PUT    /api/v1/subject-configurations/{id}               — Update konfigurasi
POST   /api/v1/subject-configurations/carry-forward      — Salin konfigurasi dari tahun sebelumnya
GET    /api/v1/subject-configurations/{id}/teachers      — Guru yang mengajar (via ADR-012 teacher_subject_map)
```

### Carry-Forward Flow

Di awal tahun ajaran baru, admin bisa copy konfigurasi mapel dari tahun sebelumnya:

```
POST /api/v1/subject-configurations/carry-forward
Body: {
  "source_academic_year_id": "018f...",
  "target_academic_year_id": "018f...",
  "grade_levels": ["7", "8", "9"]
}
→ Copy all subject_configurations from source to target
→ Admin bisa edit setelah carry-forward
```

## Consequences

### Positive

- **Evolusi dari S011**: Backward compatible dengan `subjects` di S011 — hanya ditambah field dan tabel konfigurasi.
- **Multi-kurikulum**: Konfigurasi per kurikulum per jenjang per tahun — mendukung transisi Merdeka/K13.
- **Pesantren lengkap**: 8 mapel keagamaan standar pesantren terdefinisi (ADR-009).
- **Jam pelajaran**: `credit_hours_per_week` menjadi input utama untuk scheduling (S021).
- **KKM/KKTP configurable**: Batas ketuntasan per mapel per jenjang — sesuai kebijakan sekolah.
- **Carry-forward**: Mengurangi beban admin di awal tahun ajaran.

### Negative / Trade-offs

- **Dua tabel**: Master + konfigurasi menambah complexity dibanding single table S011. Trade-off untuk fleksibilitas per tahun/jenjang.
- **Subject group hardcoded**: 16 groups di CHECK constraint — perlu ALTER TABLE jika ada group baru dari Kemendikbud.
- **Weight constraint**: `weight_knowledge + weight_skill = 100` mengasumsikan hanya 2 komponen. Jika kurikulum baru menambah komponen, schema perlu diubah.
- **Carry-forward complexity**: Bulk copy harus handle existing data (skip vs overwrite).

## Alternatives Considered

### 1. Single table tanpa subject_configurations
- Ditolak: jam pelajaran dan KKM berbeda per jenjang per tahun — tanpa tabel konfigurasi, harus duplikasi seluruh record mapel.

### 2. Subject groups sebagai tabel terpisah
- Deferred: untuk MVP, CHECK constraint cukup. Jika perlu user-defined groups, bisa migrasi ke tabel.

### 3. Konfigurasi sebagai JSONB di subjects
- Ditolak: konfigurasi per tahun per jenjang memerlukan query yang efisien — JSONB nested tidak cocok untuk filtering dan aggregation.

### 4. Mapel pesantren sebagai module terpisah
- Ditolak: mapel pesantren sama strukturnya dengan mapel umum (nama, kode, jam, nilai) — hanya beda group dan category. Satu tabel dengan filter lebih sederhana.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `SubjectDescriptor.TableName()` | — | `"subjects"` |
| U02 | `SubjectConfigDescriptor.TableName()` | — | `"subject_configurations"` |
| U03 | Validate rejects invalid `subject_group` | `"science"` | Error |
| U04 | Validate rejects invalid `category` | `"mandatory"` | Error |
| U05 | Validate rejects `credit_hours > 12` | `13` | Error |
| U06 | Validate rejects `weight total != 100` | `60 + 50` | Error |
| U07 | Validate rejects `passing_grade > 100` | `101` | Error |
| U08 | Validate accepts valid subject | All fields valid | No error |

### Integration Tests — Subjects

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create subject | POST with valid data | 201 |
| I02 | Unique code per company | Create 2 with same code | 409/422 |
| I03 | Subject group CHECK | INSERT with `subject_group = 'science'` | DB error |
| I04 | Category CHECK | INSERT with `category = 'mandatory'` | DB error |
| I05 | List by group + category | GET /subjects?group=keagamaan&category=diniyah | 200, pesantren mapel only |
| I06 | Create pesantren subjects | POST Fiqh, Tahfidz, etc. | 201, correct group/category |

### Integration Tests — Configurations

| # | Test Case | Action | Expected |
|---|---|---|---|
| I07 | Create configuration | POST with subject_id + year + grade | 201 |
| I08 | Unique per subject+year+grade | Create duplicate | 409/422 |
| I09 | Weight total CHECK | INSERT with 60+50 | DB error |
| I10 | Credit hours CHECK | INSERT with 15 | DB error |
| I11 | Carry-forward | POST carry-forward from 2024/2025 to 2025/2026 | 201, all configs copied |
| I12 | Carry-forward skip existing | Carry-forward when target has data | Skip existing, create new only |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | SubjectUpdated syncs to configs | Update subject name | `_data.subject.name` updated in configurations |
| I14 | SubjectUpdated syncs to grades (S011) | Update subject name | `_data.subject.name` updated in student_grades |
| I15 | AcademicYearUpdated syncs to configs | Update year name | `_data.academic_year.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Cannot access other tenant's subjects | GET with wrong tenant | 404 |
| I17 | Cannot access other tenant's configs | GET with wrong tenant | 404 |
