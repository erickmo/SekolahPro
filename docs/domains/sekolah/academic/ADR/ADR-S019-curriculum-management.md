# ADR-S019: Curriculum Management (Manajemen Kurikulum)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Kurikulum adalah **fondasi dari seluruh aktivitas akademik**. Tanpa kurikulum yang terdefinisi, tidak ada dasar untuk menentukan mata pelajaran (S020), jadwal (S021), penilaian (S022), RPP (S024), atau rapor (S018).

Indonesia saat ini berada dalam **masa transisi kurikulum**:

1. **Kurikulum Merdeka** (2022-sekarang): Pendekatan baru dari Kemendikbud dengan Capaian Pembelajaran (CP), Tujuan Pembelajaran (TP), Alur Tujuan Pembelajaran (ATP), Modul Ajar, dan Profil Pelajar Pancasila (P5).
2. **KTSP / K13** (legacy): Masih digunakan sebagian sekolah — berbasis Kompetensi Inti (KI) dan Kompetensi Dasar (KD).
3. **Kurikulum Pesantren**: Sekolah Islam/pesantren menambahkan kurikulum diniyah di atas kurikulum nasional (ADR-009).

Struktur hierarki kurikulum:
```
Kurikulum (e.g. Merdeka 2024/2025 Kelas 7)
  └─ Capaian Pembelajaran (CP) per fase
       └─ Tujuan Pembelajaran (TP)
            └─ Alur Tujuan Pembelajaran (ATP) — urutan TP per semester
                 └─ Modul Ajar — implementasi TP di kelas
```

Elemen Profil Pelajar Pancasila (P5):
- Beriman, Bertakwa kepada Tuhan YME, dan Berakhlak Mulia
- Berkebinekaan Global
- Bergotong Royong
- Mandiri
- Bernalar Kritis
- Kreatif

### Mengapa Vernon Pattern?

- Referenced by subjects (S020), lesson plans (S024), grades (S011/S022), dan rapor (S018).
- Read-heavy: kurikulum jarang berubah, tapi dibaca setiap kali setup mapel, buat RPP, atau generate rapor.
- Relasi ke academic_year (ADR-010) — kurikulum berlaku per tahun ajaran per jenjang.
- Jumlah record moderate: ~3-5 kurikulum per sekolah (Merdeka + legacy + diniyah).
- SyncEngine: perubahan nama kurikulum propagate ke domain yang reference.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `curricula` (master kurikulum), `learning_outcomes` (CP/TP/ATP), dan `p5_projects` (proyek Profil Pelajar Pancasila).

### Table Schema

```sql
-- Master kurikulum
CREATE TABLE curricula (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    curriculum_type VARCHAR(20) NOT NULL,

    -- Scope
    academic_year_id UUID NOT NULL,
    grade_level     VARCHAR(5) NOT NULL,
    phase           VARCHAR(10),

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,
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

    CONSTRAINT uq_curriculum_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT uq_curriculum_year_grade UNIQUE (tenant_id, company_id, academic_year_id, grade_level, curriculum_type),
    CONSTRAINT chk_curriculum_type CHECK (curriculum_type IN ('merdeka', 'k13', 'ktsp', 'diniyah', 'custom')),
    CONSTRAINT chk_curriculum_grade CHECK (grade_level IN ('1','2','3','4','5','6','7','8','9','10','11','12')),
    CONSTRAINT chk_curriculum_phase CHECK (phase IS NULL OR phase IN ('A','B','C','D','E','F'))
);

-- Indexes
CREATE INDEX idx_curriculum_tenant_company ON curricula (tenant_id, company_id);
CREATE INDEX idx_curriculum_year ON curricula (academic_year_id);
CREATE INDEX idx_curriculum_type ON curricula (curriculum_type);
CREATE INDEX idx_curriculum_active ON curricula (is_active) WHERE is_active = true;
CREATE INDEX idx_curriculum_rels ON curricula USING GIN (_rels);
CREATE INDEX idx_curriculum_data ON curricula USING GIN (_data);

-- Capaian Pembelajaran (CP), Tujuan Pembelajaran (TP), Alur Tujuan Pembelajaran (ATP)
CREATE TABLE learning_outcomes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    curriculum_id   UUID NOT NULL,
    subject_id      UUID,
    parent_id       UUID,

    -- Hierarki
    outcome_type    VARCHAR(10) NOT NULL,
    code            VARCHAR(30) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,

    -- Ordering
    semester        VARCHAR(10),
    sequence_order  INT NOT NULL DEFAULT 0,

    -- K13 compatibility
    ki_number       INT,
    kd_code         VARCHAR(20),

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

    CONSTRAINT uq_learning_outcome_code UNIQUE (tenant_id, company_id, curriculum_id, code),
    CONSTRAINT chk_outcome_type CHECK (outcome_type IN ('cp', 'tp', 'atp')),
    CONSTRAINT chk_outcome_semester CHECK (semester IS NULL OR semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_ki_number CHECK (ki_number IS NULL OR ki_number IN (1, 2, 3, 4))
);

-- Indexes
CREATE INDEX idx_lo_tenant_company ON learning_outcomes (tenant_id, company_id);
CREATE INDEX idx_lo_curriculum ON learning_outcomes (curriculum_id);
CREATE INDEX idx_lo_subject ON learning_outcomes (subject_id);
CREATE INDEX idx_lo_parent ON learning_outcomes (parent_id);
CREATE INDEX idx_lo_type ON learning_outcomes (outcome_type);
CREATE INDEX idx_lo_rels ON learning_outcomes USING GIN (_rels);
CREATE INDEX idx_lo_data ON learning_outcomes USING GIN (_data);

-- Proyek Profil Pelajar Pancasila (P5)
CREATE TABLE p5_projects (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    curriculum_id   UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Identitas
    name            VARCHAR(200) NOT NULL,
    theme           VARCHAR(50) NOT NULL,
    description     TEXT,

    -- Scope
    grade_level     VARCHAR(5) NOT NULL,
    semester        VARCHAR(10) NOT NULL,

    -- Dimensi P5 yang diukur (multi-select)
    p5_dimensions   JSONB NOT NULL DEFAULT '[]',

    -- Timeline
    start_date      DATE,
    end_date        DATE,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_p5_theme CHECK (theme IN (
        'gaya_hidup_berkelanjutan', 'kearifan_lokal',
        'bhinneka_tunggal_ika', 'bangunlah_jiwa_raganya',
        'suara_demokrasi', 'berekayasa_dan_berteknologi',
        'kewirausahaan'
    )),
    CONSTRAINT chk_p5_grade CHECK (grade_level IN ('1','2','3','4','5','6','7','8','9','10','11','12')),
    CONSTRAINT chk_p5_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_p5_status CHECK (status IN ('draft', 'active', 'completed', 'archived')),
    CONSTRAINT chk_p5_dates CHECK (start_date IS NULL OR end_date IS NULL OR start_date <= end_date)
);

-- Indexes
CREATE INDEX idx_p5_tenant_company ON p5_projects (tenant_id, company_id);
CREATE INDEX idx_p5_curriculum ON p5_projects (curriculum_id);
CREATE INDEX idx_p5_year ON p5_projects (academic_year_id);
CREATE INDEX idx_p5_status ON p5_projects (status);
CREATE INDEX idx_p5_rels ON p5_projects USING GIN (_rels);
CREATE INDEX idx_p5_data ON p5_projects USING GIN (_data);
```

### Field Design Rationale

**curricula:**

| Field | Keputusan | Alasan |
|---|---|---|
| `curriculum_type` | 5 tipe | Merdeka, K13, KTSP, Diniyah (pesantren per ADR-009), Custom |
| `grade_level` | VARCHAR(5), CHECK 1-12 | Kurikulum berlaku per jenjang kelas — SD (1-6), SMP (7-9), SMA (10-12) |
| `phase` | VARCHAR(10), nullable | Fase Kurikulum Merdeka: A (kelas 1-2), B (3-4), C (5-6), D (7-9), E (10), F (11-12) |
| `academic_year_id` | UUID, NOT NULL | Kurikulum terikat tahun ajaran — bisa berubah tiap tahun |
| `code` | VARCHAR(20), unique per company | Kode singkat: "MRD-2526-7", "K13-2526-8" |

**learning_outcomes:**

| Field | Keputusan | Alasan |
|---|---|---|
| `outcome_type` | 3 tipe: cp, tp, atp | Hierarki Kurikulum Merdeka: CP > TP > ATP |
| `parent_id` | UUID, nullable, self-reference | CP tidak punya parent; TP parent = CP; ATP parent = TP |
| `subject_id` | UUID, nullable | CP/TP terikat mapel; beberapa CP bersifat lintas mapel |
| `ki_number` | INT, nullable | Kompatibilitas K13: KI-1 (Spiritual), KI-2 (Sosial), KI-3 (Pengetahuan), KI-4 (Keterampilan) |
| `kd_code` | VARCHAR(20), nullable | Kompatibilitas K13: kode Kompetensi Dasar (e.g. "3.1", "4.2") |
| `semester` | VARCHAR(10), nullable | ATP bersifat per semester; CP/TP bisa lintas semester |

**p5_projects:**

| Field | Keputusan | Alasan |
|---|---|---|
| `theme` | 7 tema standar Kemendikbud | Tema P5 sudah ditetapkan nasional |
| `p5_dimensions` | JSONB array | Satu proyek bisa mengukur beberapa dimensi P5 sekaligus |
| `status` | 4 stage lifecycle | draft > active > completed > archived |

### Vernon Relationships

**curricula:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

**learning_outcomes:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `curriculum` | belongs_to | **Ya** | Kurikulum induk |
| `subject` | belongs_to | **Ya** | Mata pelajaran terkait |
| `parent` | belongs_to | **Tidak** | Self-reference — load on demand untuk tree |

**p5_projects:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `curriculum` | belongs_to | **Ya** | Kurikulum induk |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |

### _rels / _data Structure

**curricula:**
```json
{
  "_rels": {
    "academic_year_id": "018f..."
  },
  "_data": {
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

**learning_outcomes:**
```json
{
  "_rels": {
    "curriculum_id": "018f...",
    "subject_id": "018f...",
    "parent_id": "018f..."
  },
  "_data": {
    "curriculum": { "id": "018f...", "name": "Kurikulum Merdeka 2025/2026 Kelas 7", "curriculum_type": "merdeka" },
    "subject": { "id": "018f...", "name": "Matematika", "code": "MTK" }
  }
}
```

**p5_projects:**
```json
{
  "_rels": {
    "curriculum_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "curriculum": { "id": "018f...", "name": "Kurikulum Merdeka 2025/2026 Kelas 7" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

### Fase Kurikulum Merdeka

| Fase | Kelas | Setara |
|------|-------|--------|
| A | 1-2 | SD Awal |
| B | 3-4 | SD Tengah |
| C | 5-6 | SD Akhir |
| D | 7-9 | SMP |
| E | 10 | SMA Kelas 10 |
| F | 11-12 | SMA Kelas 11-12 |

### API Endpoints

```
# Curricula
GET    /api/v1/curricula                                — List kurikulum
GET    /api/v1/curricula?year_id={id}&grade_level=7     — Filter per tahun + jenjang
POST   /api/v1/curricula                                — Buat kurikulum
PUT    /api/v1/curricula/{id}                           — Update kurikulum
GET    /api/v1/curricula/{id}                           — Detail kurikulum

# Learning Outcomes (CP/TP/ATP)
GET    /api/v1/curricula/{id}/learning-outcomes          — Tree CP > TP > ATP per kurikulum
GET    /api/v1/learning-outcomes?type=cp&subject_id={id} — Filter by type + subject
POST   /api/v1/learning-outcomes                         — Buat CP/TP/ATP
PUT    /api/v1/learning-outcomes/{id}                    — Update
DELETE /api/v1/learning-outcomes/{id}                    — Soft delete

# P5 Projects
GET    /api/v1/p5-projects                               — List proyek P5
GET    /api/v1/p5-projects?year_id={id}&grade_level=7    — Filter
POST   /api/v1/p5-projects                               — Buat proyek P5
PUT    /api/v1/p5-projects/{id}                          — Update
POST   /api/v1/p5-projects/{id}/activate                 — Aktivasi proyek
POST   /api/v1/p5-projects/{id}/complete                 — Tandai selesai
```

## Consequences

### Positive

- **Multi-kurikulum**: Mendukung Kurikulum Merdeka, K13/KTSP, dan Diniyah pesantren secara bersamaan.
- **Hierarki lengkap**: CP > TP > ATP terdefinisi dengan self-referencing parent_id.
- **P5 terintegrasi**: Proyek P5 sebagai tabel terpisah memungkinkan tracking per tema dan dimensi.
- **Backward compatible**: Field `ki_number` dan `kd_code` mendukung sekolah yang masih K13.
- **Per tahun ajaran**: Kurikulum terikat academic_year sehingga bisa berevolusi tiap tahun.
- **Pesantren support**: `curriculum_type = 'diniyah'` untuk kurikulum keagamaan (ADR-009).

### Negative / Trade-offs

- **Kompleksitas setup**: Admin harus setup CP/TP/ATP per kurikulum per tahun — butuh template import atau carry-forward.
- **Tree query**: Self-referencing learning_outcomes memerlukan recursive query (WITH RECURSIVE) untuk render hierarki lengkap.
- **P5 dimensions as JSONB**: Tidak bisa di-query seefisien kolom terpisah — trade-off untuk fleksibilitas multi-select.
- **Volume data**: 12 grade levels x 10 subjects x ~5 TP per subject = ~600 learning_outcomes per kurikulum.
- **Tema P5 hardcoded**: 7 tema nasional di CHECK constraint — jika Kemendikbud menambah tema, ALTER TABLE diperlukan.

## Alternatives Considered

### 1. Kurikulum sebagai JSONB di academic_years
- Ditolak: kurikulum punya lifecycle sendiri, bisa beberapa kurikulum per tahun ajaran (nasional + diniyah), dan perlu di-query per subject.

### 2. Satu tabel flat untuk semua learning outcomes tanpa hierarki
- Ditolak: hierarki CP > TP > ATP adalah struktur fundamental Kurikulum Merdeka — tanpa parent_id, tidak bisa render tree atau validasi kelengkapan.

### 3. P5 sebagai bagian dari learning_outcomes
- Ditolak: P5 punya struktur berbeda (tema, dimensi, timeline proyek) — tidak cocok di tabel learning_outcomes yang fokus pada kompetensi mapel.

### 4. Import langsung dari data Kemendikbud
- Deferred: API Kemendikbud tidak stabil dan tidak semua sekolah terhubung. Untuk MVP, manual input + template import lebih reliable.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `CurriculumDescriptor.TableName()` | — | `"curricula"` |
| U02 | `LearningOutcomeDescriptor.TableName()` | — | `"learning_outcomes"` |
| U03 | `P5ProjectDescriptor.TableName()` | — | `"p5_projects"` |
| U04 | Validate rejects invalid `curriculum_type` | `"cbsa"` | Error: invalid curriculum_type |
| U05 | Validate rejects invalid `grade_level` | `"13"` | Error: invalid grade_level |
| U06 | Validate rejects invalid `phase` | `"G"` | Error: invalid phase |
| U07 | Validate rejects invalid `outcome_type` | `"ki"` | Error: must be cp/tp/atp |
| U08 | Validate rejects invalid `p5_theme` | `"random"` | Error: invalid theme |
| U09 | Validate rejects `start_date > end_date` on P5 | start > end | Error |
| U10 | Validate accepts valid curriculum | All fields valid | No error |

### Integration Tests — Curricula

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create curriculum | POST with valid data | 201 |
| I02 | Unique code per company | Create 2 with same code | 409/422 |
| I03 | Unique per year+grade+type | Create duplicate | 409/422 |
| I04 | Curriculum type CHECK | INSERT with `curriculum_type = 'cbsa'` | DB error |
| I05 | List by year + grade | GET /curricula?year_id=x&grade_level=7 | 200, filtered |

### Integration Tests — Learning Outcomes

| # | Test Case | Action | Expected |
|---|---|---|---|
| I06 | Create CP | POST with outcome_type=cp | 201 |
| I07 | Create TP under CP | POST with parent_id = CP.id | 201 |
| I08 | Create ATP under TP | POST with parent_id = TP.id, semester=ganjil | 201 |
| I09 | Get tree per curriculum | GET /curricula/{id}/learning-outcomes | 200, hierarchical structure |
| I10 | Unique code per curriculum | Create 2 with same code | 409/422 |
| I11 | K13 compatibility | Create with ki_number=3, kd_code="3.1" | 201 |

### Integration Tests — P5 Projects

| # | Test Case | Action | Expected |
|---|---|---|---|
| I12 | Create P5 project | POST with theme + dimensions | 201 |
| I13 | Theme CHECK | INSERT with invalid theme | DB error |
| I14 | Activate project | POST /p5-projects/{id}/activate | 200, status=active |
| I15 | Complete project | POST /p5-projects/{id}/complete | 200, status=completed |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | AcademicYearUpdated syncs to curricula | Update year name | `_data.academic_year.name` updated |
| I17 | CurriculumUpdated syncs to learning_outcomes | Update curriculum name | `_data.curriculum.name` updated |
| I18 | SubjectUpdated syncs to learning_outcomes | Update subject name | `_data.subject.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I19 | Cannot access other tenant's curriculum | GET with wrong tenant | 404 |
| I20 | Cannot access other tenant's learning outcomes | GET with wrong tenant | 404 |
