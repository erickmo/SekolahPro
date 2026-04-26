# ADR-S022: Exam & Assessment Management (Ujian & Penilaian)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

ADR-S011 menyimpan **nilai akhir per semester** (final_score, score_knowledge, score_skill), namun belum ada mekanisme pencatatan **penilaian harian dan ujian** yang menjadi sumber perhitungan nilai akhir tersebut.

Sistem penilaian di Indonesia saat ini berada dalam transisi:

1. **Kurikulum Merdeka**:
   - **Asesmen Formatif**: Penilaian proses (tugas, kuis, observasi) — tidak harus berupa angka.
   - **Asesmen Sumatif**: Penilaian akhir per unit/bab — berupa angka/deskripsi.
   - **Sumatif Akhir Semester (SAS)**: Menggantikan UAS/PAS.
   - **Sumatif Tengah Semester (STS)**: Menggantikan UTS/PTS.

2. **K13 / KTSP** (legacy):
   - **Ulangan Harian (UH)**: Per bab/unit.
   - **Penilaian Tengah Semester (PTS/UTS)**: Tengah semester.
   - **Penilaian Akhir Semester (PAS/UAS)**: Akhir semester.
   - **Penilaian Akhir Tahun (PAT)**: Akhir tahun (semester genap).

3. **Pesantren** (ADR-009): Ujian tambahan untuk mapel diniyah — Imtihan, hafalan (setoran tahfidz).

Kebutuhan:
- Jadwal ujian per kelas per semester.
- Input nilai per ujian per siswa (bulk input).
- Perhitungan otomatis ke S011 (final_score = weighted average of assessments).
- Bobot penilaian configurable per mapel (dari S020 subject_configurations).

### Mengapa Vernon Pattern?

- has_many dari student (banyak penilaian per siswa per semester per mapel).
- Relasi ke student, subject, teacher, academic_year, class_room.
- Read-heavy: rekap nilai, dashboard, dan input nilai harian.
- Business logic moderate: weighted average, passing grade check.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `assessments` (definisi ujian/penilaian) dan `assessment_scores` (nilai per siswa per penilaian).

### Table Schema

```sql
-- Definisi ujian/penilaian
CREATE TABLE assessments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    subject_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    class_room_id   UUID NOT NULL,
    teacher_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(200) NOT NULL,
    assessment_type VARCHAR(30) NOT NULL,
    assessment_category VARCHAR(20) NOT NULL,

    -- Periode
    semester        VARCHAR(10) NOT NULL,

    -- Jadwal
    scheduled_date  DATE,
    scheduled_start TIME,
    scheduled_end   TIME,

    -- Konfigurasi
    max_score       NUMERIC(5,2) NOT NULL DEFAULT 100.00,
    weight          NUMERIC(5,2) NOT NULL DEFAULT 1.00,
    description     TEXT,

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

    CONSTRAINT chk_assessment_type CHECK (assessment_type IN (
        -- Kurikulum Merdeka
        'formatif', 'sumatif', 'sumatif_tengah_semester', 'sumatif_akhir_semester',
        -- K13 / KTSP
        'ulangan_harian', 'pts', 'pas', 'pat',
        -- Pesantren
        'imtihan', 'setoran_tahfidz',
        -- Generic
        'tugas', 'praktik', 'proyek', 'portofolio'
    )),
    CONSTRAINT chk_assessment_category CHECK (assessment_category IN (
        'knowledge', 'skill', 'attitude', 'tahfidz'
    )),
    CONSTRAINT chk_assessment_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_assessment_status CHECK (status IN ('draft', 'scheduled', 'in_progress', 'completed', 'finalized')),
    CONSTRAINT chk_max_score CHECK (max_score > 0 AND max_score <= 100),
    CONSTRAINT chk_weight CHECK (weight > 0 AND weight <= 10),
    CONSTRAINT chk_schedule_time CHECK (scheduled_start IS NULL OR scheduled_end IS NULL OR scheduled_start < scheduled_end)
);

-- Indexes
CREATE INDEX idx_assessment_tenant_company ON assessments (tenant_id, company_id);
CREATE INDEX idx_assessment_subject ON assessments (subject_id);
CREATE INDEX idx_assessment_class ON assessments (class_room_id);
CREATE INDEX idx_assessment_teacher ON assessments (teacher_id);
CREATE INDEX idx_assessment_year_semester ON assessments (academic_year_id, semester);
CREATE INDEX idx_assessment_type ON assessments (assessment_type);
CREATE INDEX idx_assessment_status ON assessments (status);
CREATE INDEX idx_assessment_date ON assessments (scheduled_date);
CREATE INDEX idx_assessment_rels ON assessments USING GIN (_rels);
CREATE INDEX idx_assessment_data ON assessments USING GIN (_data);

-- Nilai per siswa per penilaian
CREATE TABLE assessment_scores (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    assessment_id   UUID NOT NULL,
    student_id      UUID NOT NULL,

    -- Nilai
    score           NUMERIC(5,2),
    score_attitude  VARCHAR(2),
    description     TEXT,

    -- Remedial
    is_remedial     BOOLEAN NOT NULL DEFAULT false,
    original_score  NUMERIC(5,2),

    -- Tahfidz specific (ADR-009)
    surah_name      VARCHAR(50),
    ayat_from       INT,
    ayat_to         INT,
    tahfidz_grade   VARCHAR(20),

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_score_assessment_student UNIQUE (assessment_id, student_id),
    CONSTRAINT chk_score_range CHECK (score IS NULL OR (score >= 0 AND score <= 100)),
    CONSTRAINT chk_original_score CHECK (original_score IS NULL OR (original_score >= 0 AND original_score <= 100)),
    CONSTRAINT chk_score_attitude CHECK (score_attitude IS NULL OR score_attitude IN ('SB', 'B', 'C', 'K')),
    CONSTRAINT chk_tahfidz_grade CHECK (tahfidz_grade IS NULL OR tahfidz_grade IN (
        'mumtaz', 'jayyid_jiddan', 'jayyid', 'maqbul', 'rasib'
    ))
);

-- Indexes
CREATE INDEX idx_score_tenant_company ON assessment_scores (tenant_id, company_id);
CREATE INDEX idx_score_assessment ON assessment_scores (assessment_id);
CREATE INDEX idx_score_student ON assessment_scores (student_id);
CREATE INDEX idx_score_remedial ON assessment_scores (is_remedial) WHERE is_remedial = true;
CREATE INDEX idx_score_rels ON assessment_scores USING GIN (_rels);
CREATE INDEX idx_score_data ON assessment_scores USING GIN (_data);
```

### Field Design Rationale

**assessments:**

| Field | Keputusan | Alasan |
|---|---|---|
| `assessment_type` | 14 tipe | Kurikulum Merdeka (4) + K13 (4) + Pesantren (2) + Generic (4) — cover semua model penilaian |
| `assessment_category` | 4 kategori | knowledge (pengetahuan), skill (keterampilan), attitude (sikap), tahfidz (hafalan khusus pesantren) |
| `weight` | NUMERIC(5,2), 0-10 | Bobot relatif penilaian — UAS biasanya weight 3, UH weight 1 |
| `max_score` | NUMERIC(5,2), default 100 | Nilai maksimal — bisa 100 (standar) atau custom |
| `status` | 5 stage | draft → scheduled → in_progress → completed → finalized |
| `scheduled_*` | DATE + TIME, nullable | Jadwal ujian — nullable untuk penilaian yang tidak dijadwalkan (tugas) |

**assessment_scores:**

| Field | Keputusan | Alasan |
|---|---|---|
| `score` | NUMERIC(5,2), nullable | Nilai angka — nullable untuk penilaian attitude (hanya predikat) |
| `score_attitude` | VARCHAR(2), nullable | Predikat sikap: SB/B/C/K — untuk assessment_category = attitude |
| `description` | TEXT, nullable | Deskripsi naratif — wajib untuk Kurikulum Merdeka formatif |
| `is_remedial` | BOOLEAN | True jika nilai ini hasil remedial |
| `original_score` | NUMERIC(5,2), nullable | Nilai sebelum remedial — untuk audit trail |
| `surah_name` + `ayat_*` | Tahfidz fields | Khusus setoran tahfidz: surah apa, ayat berapa sampai berapa |
| `tahfidz_grade` | 5 level predikat Arab | mumtaz (sempurna), jayyid jiddan (sangat baik), jayyid (baik), maqbul (cukup), rasib (kurang) |

### Calculation Flow: Assessment → S011

```
Assessments (weighted)                    → S011 (student_grades)
──────────────────────                      ──────────────────────
UH-1 (weight 1): 75                        score_knowledge = weighted avg of knowledge assessments
UH-2 (weight 1): 80                        score_skill = weighted avg of skill assessments
PTS  (weight 2): 85                        score_attitude = modus of attitude assessments
PAS  (weight 3): 90                        final_score = (knowledge × weight_k + skill × weight_s) / 100
                                            → from S020 subject_configurations
Weighted avg knowledge:
= (75×1 + 80×1 + 85×2 + 90×3) / (1+1+2+3)
= 595 / 7 = 85.00
```

Perhitungan dilakukan via event `AssessmentFinalized` → aggregate → update `student_grades` (S011).

### Vernon Relationships

**assessments:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `subject` | belongs_to | **Ya** | Nama mata pelajaran |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |
| `class_room` | belongs_to | **Ya** | Kelas |
| `teacher` | belongs_to | **Ya** | Guru pengajar |

**assessment_scores:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `assessment` | belongs_to | **Ya** | Konteks penilaian |
| `student` | belongs_to | **Ya** | Pemilik nilai |

### _rels / _data Structure

**assessments:**
```json
{
  "_rels": {
    "subject_id": "018f...",
    "academic_year_id": "018f...",
    "class_room_id": "018f...",
    "teacher_id": "018f..."
  },
  "_data": {
    "subject": { "id": "018f...", "name": "Matematika", "code": "MTK" },
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "class_room": { "id": "018f...", "name": "VII-A" },
    "teacher": { "id": "018f...", "full_name": "Bu Siti" }
  }
}
```

**assessment_scores:**
```json
{
  "_rels": {
    "assessment_id": "018f...",
    "student_id": "018f..."
  },
  "_data": {
    "assessment": {
      "id": "018f...",
      "name": "UTS Matematika Sem 1",
      "assessment_type": "pts",
      "assessment_category": "knowledge",
      "subject_name": "Matematika",
      "max_score": 100
    },
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" }
  }
}
```

### API Endpoints

```
# Assessments
GET    /api/v1/assessments?class_id={id}&subject_id={id}&semester=ganjil  — List penilaian per kelas per mapel
POST   /api/v1/assessments                                                — Buat penilaian
PUT    /api/v1/assessments/{id}                                           — Update penilaian
POST   /api/v1/assessments/{id}/schedule                                  — Set jadwal ujian
POST   /api/v1/assessments/{id}/finalize                                  — Finalize (lock nilai)

# Scores (Bulk Input)
POST   /api/v1/assessments/{id}/scores/bulk                               — Bulk input nilai (1 kelas)
PUT    /api/v1/assessment-scores/{id}                                     — Update nilai individual
GET    /api/v1/assessments/{id}/scores                                    — Rekap nilai per penilaian

# Student View
GET    /api/v1/students/{id}/assessment-scores?subject_id={id}&semester=ganjil — Semua nilai siswa per mapel
GET    /api/v1/students/{id}/assessment-summary?semester=ganjil            — Ringkasan per mapel (avg, min, max)

# Remedial
POST   /api/v1/assessment-scores/{id}/remedial                            — Input nilai remedial
  Body: { "score": 78 } → original_score = old score, score = 78, is_remedial = true

# Calculation
POST   /api/v1/assessments/calculate-final?class_id={id}&subject_id={id}&semester=ganjil
  → Calculate weighted average → Update S011 student_grades

# Exam Schedule View
GET    /api/v1/assessments/schedule?year_id={id}&semester=ganjil           — Jadwal ujian (PTS/PAS)
```

### Bulk Score Input Payload

```json
{
  "scores": [
    { "student_id": "018f...", "score": 85.5 },
    { "student_id": "018f...", "score": 72.0, "description": "Perlu perbaikan pada bab pecahan" },
    { "student_id": "018f...", "score": null, "description": "Tidak mengikuti (sakit)" }
  ]
}
```

## Consequences

### Positive

- **Multi-kurikulum**: Assessment types cover Kurikulum Merdeka, K13, dan Pesantren.
- **Sumber kebenaran**: S011 final_score dihitung dari assessment scores — bukan input manual.
- **Remedial tracking**: Nilai asli tetap tersimpan (`original_score`) — audit trail lengkap.
- **Tahfidz support**: Field khusus hafalan (surah, ayat, grade) untuk pesantren (ADR-009).
- **Bulk input**: Guru bisa input nilai 30+ siswa sekaligus per penilaian.
- **Weighted average**: Bobot configurable per assessment — mendukung berbagai kebijakan sekolah.

### Negative / Trade-offs

- **Volume data tinggi**: 30 siswa × 10 mapel × ~10 penilaian/semester = 3.000 assessment_scores per kelas per semester.
- **Assessment type complexity**: 14 tipe di CHECK constraint — maintenance burden jika kurikulum berubah.
- **Tahfidz fields di generic table**: `surah_name`, `ayat_*`, `tahfidz_grade` nullable di semua records meski hanya dipakai pesantren. Trade-off untuk simplicity vs separate table.
- **Calculation async**: Weighted average ke S011 bersifat eventual consistent — ada delay antara input nilai dan update final_score.
- **No question bank**: ADR ini fokus pada nilai/skor, bukan soal ujian. Question bank bisa jadi domain terpisah jika dibutuhkan.

## Alternatives Considered

### 1. Nilai langsung di S011 tanpa assessment detail
- Ditolak: tidak bisa menampilkan rincian nilai per penilaian, tidak bisa hitung weighted average, guru harus manual hitung final score.

### 2. Tabel terpisah untuk setiap tipe assessment
- Ditolak: terlalu banyak tabel (UH, PTS, PAS, formatif, sumatif, dst.). Satu tabel `assessments` dengan `assessment_type` lebih fleksibel.

### 3. Tahfidz sebagai tabel terpisah
- Deferred: untuk MVP, nullable fields di `assessment_scores` cukup. Jika tahfidz butuh tracking lebih detail (per halaman, per juz, mutqin tracking), bisa jadi domain terpisah.

### 4. Question bank terintegrasi
- Deferred: question bank (bank soal) adalah fitur complex — butuh kategorisasi soal, difficulty level, random selection. Di-defer untuk enhancement setelah MVP.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `AssessmentDescriptor.TableName()` | — | `"assessments"` |
| U02 | `ScoreDescriptor.TableName()` | — | `"assessment_scores"` |
| U03 | Validate rejects invalid `assessment_type` | `"quiz"` | Error |
| U04 | Validate rejects invalid `assessment_category` | `"social"` | Error |
| U05 | Validate rejects `max_score = 0` | `0` | Error |
| U06 | Validate rejects `score > 100` | `101` | Error |
| U07 | Validate rejects invalid `tahfidz_grade` | `"excellent"` | Error |
| U08 | Validate rejects invalid `score_attitude` | `"A"` | Error |
| U09 | Weighted average calculation | 3 scores with weights | Correct weighted avg |
| U10 | Validate accepts valid assessment | All fields valid | No error |

### Integration Tests — Assessments

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create assessment | POST with valid data | 201 |
| I02 | Assessment type CHECK | INSERT with `assessment_type = 'quiz'` | DB error |
| I03 | Category CHECK | INSERT with `assessment_category = 'social'` | DB error |
| I04 | Schedule assessment | POST /assessments/{id}/schedule | 200, dates set |
| I05 | Finalize assessment | POST /assessments/{id}/finalize | 200, status=finalized |
| I06 | Cannot edit finalized | PUT on finalized assessment | 403 |

### Integration Tests — Scores

| # | Test Case | Action | Expected |
|---|---|---|---|
| I07 | Bulk input scores | POST /assessments/{id}/scores/bulk for 30 students | 201, 30 records |
| I08 | Unique per assessment+student | Insert same score twice | 409/422 |
| I09 | Score range CHECK | INSERT with score = 150 | DB error |
| I10 | Attitude CHECK | INSERT with score_attitude = 'A' | DB error |
| I11 | Null score allowed | INSERT with score = null, description = "Sakit" | 201 |

### Integration Tests — Remedial

| # | Test Case | Action | Expected |
|---|---|---|---|
| I12 | Input remedial | POST /assessment-scores/{id}/remedial | 200, original_score preserved |
| I13 | Remedial score replaces current | Check score after remedial | score = new, original_score = old |

### Integration Tests — Tahfidz (ADR-009)

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | Create setoran tahfidz | POST with type=setoran_tahfidz, surah + ayat | 201 |
| I15 | Tahfidz grade CHECK | INSERT with `tahfidz_grade = 'excellent'` | DB error |
| I16 | Valid tahfidz grades | INSERT mumtaz, jayyid, etc. | 201 each |

### Integration Tests — Calculation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Calculate final score | POST /calculate-final | 200, S011 student_grades updated |
| I18 | Weighted average correct | 3 assessments with different weights | Calculated avg matches manual |
| I19 | Null scores excluded | 2 scores + 1 null | Average from 2 only |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I20 | StudentUpdated syncs to scores | Update student name | `_data.student.full_name` updated |
| I21 | SubjectUpdated syncs to assessments | Update subject name | `_data.subject.name` updated |
| I22 | TeacherUpdated syncs to assessments | Update teacher name | `_data.teacher.full_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I23 | Cannot access other tenant's assessments | GET with wrong tenant | 404 |
| I24 | Cannot bulk input to other tenant's assessment | POST bulk with wrong tenant | Error |
