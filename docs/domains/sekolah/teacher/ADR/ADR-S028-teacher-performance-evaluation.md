# ADR-S028: Teacher Performance Evaluation (Penilaian Kinerja Guru - PKG)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Penilaian Kinerja Guru (PKG) adalah kewajiban regulasi berdasarkan **Permenneg PAN & RB No. 16 Tahun 2009** dan **Permendiknas No. 35 Tahun 2010**. PKG menilai guru berdasarkan 4 kompetensi:

1. **Kompetensi Pedagogik** (7 indikator): Kemampuan merencanakan, melaksanakan, dan mengevaluasi pembelajaran.
2. **Kompetensi Kepribadian** (3 indikator): Integritas, kedewasaan, keteladanan.
3. **Kompetensi Sosial** (2 indikator): Komunikasi dengan sesama guru, orang tua, masyarakat.
4. **Kompetensi Profesional** (2 indikator): Penguasaan materi, pengembangan profesi.

PKG digunakan untuk:

- **SKP (Sasaran Kinerja Pegawai)**: Bagian dari penilaian tahunan ASN — dilaporkan ke BKN.
- **Kenaikan pangkat**: PKG minimal "Baik" diperlukan untuk kenaikan pangkat PNS.
- **Tunjangan profesi**: PKG "Amat Baik" atau "Baik" untuk pencairan tunjangan sertifikasi.
- **Pengembangan profesi**: Identifikasi area yang perlu PKB (S029).
- **Rapor mutu sekolah**: Agregasi PKG seluruh guru menjadi indikator mutu sekolah.

Mekanisme penilaian:

| Penilai | Bobot | Konteks |
|---------|-------|---------|
| Self-assessment | Referensi | Guru menilai diri sendiri — sebagai pembanding |
| Peer assessment | 20-30% | Rekan sejawat sesama guru |
| Supervisor (Kepsek/Wakasek) | 70-80% | Atasan langsung sebagai penilai utama |
| Observasi kelas | Kualitatif | Kunjungan kelas oleh penilai |

### Mengapa Vernon Pattern?

- has_many dari teacher: satu PKG per guru per semester.
- Multi-assessor: satu evaluasi bisa dinilai oleh 3+ penilai.
- Read-heavy: laporan PKG untuk BKN, dashboard kepsek, ringkasan tahunan.
- Business logic moderate: hitung skor per kompetensi, predikat, aggregation.
- Eventually consistent acceptable — evaluasi bersifat periodik (per semester).

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `teacher_evaluations` (evaluasi per guru per semester), `teacher_evaluation_scores` (skor per kompetensi per penilai), dan `teacher_evaluation_competencies` (master indikator kompetensi).

### Table Schema

```sql
-- Master kompetensi dan indikator PKG
CREATE TABLE teacher_evaluation_competencies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    competency_area VARCHAR(30) NOT NULL,
    indicator_code  VARCHAR(20) NOT NULL,
    indicator_name  VARCHAR(255) NOT NULL,
    description     TEXT,
    max_score       NUMERIC(5,2) NOT NULL DEFAULT 4,
    sort_order      INT NOT NULL DEFAULT 0,
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

    CONSTRAINT uq_competency_code UNIQUE (tenant_id, company_id, indicator_code),
    CONSTRAINT chk_competency_area CHECK (competency_area IN (
        'pedagogik', 'kepribadian', 'sosial', 'profesional'
    )),
    CONSTRAINT chk_max_score CHECK (max_score > 0)
);

-- Indexes
CREATE INDEX idx_eval_comp_tenant_company ON teacher_evaluation_competencies (tenant_id, company_id);
CREATE INDEX idx_eval_comp_area ON teacher_evaluation_competencies (competency_area);

-- Evaluasi PKG per guru per semester
CREATE TABLE teacher_evaluations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Periode
    semester        VARCHAR(10) NOT NULL,
    evaluation_date DATE NOT NULL,

    -- Skor agregat (dihitung dari scores)
    score_pedagogik    NUMERIC(5,2),
    score_kepribadian  NUMERIC(5,2),
    score_sosial       NUMERIC(5,2),
    score_profesional  NUMERIC(5,2),
    total_score        NUMERIC(5,2),

    -- Predikat
    grade           VARCHAR(20),

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',

    -- Catatan & rekomendasi
    recommendation  TEXT,
    improvement_plan TEXT,

    -- Approval
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_evaluation_teacher_semester UNIQUE (teacher_id, academic_year_id, semester),
    CONSTRAINT chk_eval_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_eval_grade CHECK (grade IS NULL OR grade IN (
        'amat_baik', 'baik', 'cukup', 'sedang', 'kurang'
    )),
    CONSTRAINT chk_eval_status CHECK (status IN (
        'draft', 'self_assessment', 'peer_review', 'supervisor_review',
        'completed', 'approved'
    )),
    CONSTRAINT chk_scores CHECK (
        (total_score IS NULL) OR (total_score >= 0 AND total_score <= 100)
    )
);

-- Indexes
CREATE INDEX idx_eval_tenant_company ON teacher_evaluations (tenant_id, company_id);
CREATE INDEX idx_eval_teacher ON teacher_evaluations (teacher_id);
CREATE INDEX idx_eval_year_semester ON teacher_evaluations (academic_year_id, semester);
CREATE INDEX idx_eval_grade ON teacher_evaluations (grade);
CREATE INDEX idx_eval_status ON teacher_evaluations (status);
CREATE INDEX idx_eval_rels ON teacher_evaluations USING GIN (_rels);
CREATE INDEX idx_eval_data ON teacher_evaluations USING GIN (_data);

-- Skor detail per kompetensi per penilai
CREATE TABLE teacher_evaluation_scores (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    evaluation_id   UUID NOT NULL,
    competency_id   UUID NOT NULL,
    assessor_id     UUID NOT NULL,

    -- Penilaian
    assessor_type   VARCHAR(20) NOT NULL,
    score           NUMERIC(5,2) NOT NULL,
    evidence        TEXT,
    note            TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_score_eval_comp_assessor UNIQUE (evaluation_id, competency_id, assessor_id),
    CONSTRAINT chk_assessor_type CHECK (assessor_type IN ('self', 'peer', 'supervisor')),
    CONSTRAINT chk_score_range CHECK (score >= 0 AND score <= 4)
);

-- Indexes
CREATE INDEX idx_eval_score_tenant_company ON teacher_evaluation_scores (tenant_id, company_id);
CREATE INDEX idx_eval_score_evaluation ON teacher_evaluation_scores (evaluation_id);
CREATE INDEX idx_eval_score_competency ON teacher_evaluation_scores (competency_id);
CREATE INDEX idx_eval_score_assessor ON teacher_evaluation_scores (assessor_id);
CREATE INDEX idx_eval_score_type ON teacher_evaluation_scores (assessor_type);
CREATE INDEX idx_eval_score_rels ON teacher_evaluation_scores USING GIN (_rels);
CREATE INDEX idx_eval_score_data ON teacher_evaluation_scores USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `competency_area` | 4 area | Sesuai standar PKG Kemendikbud: pedagogik, kepribadian, sosial, profesional |
| `indicator_code` | VARCHAR(20), unique per company | Kode indikator: P1, P2, ..., K1, K2, ..., S1, S2, ..., PR1, PR2 |
| `max_score` | NUMERIC(5,2), DEFAULT 4 | Skala PKG standar: 1-4 per indikator |
| `score_pedagogik` s/d `score_profesional` | NUMERIC(5,2), nullable | Denormalisasi skor rata-rata per area — dihitung dari detail scores |
| `total_score` | NUMERIC(5,2), 0-100 | Konversi ke skala 100 untuk grade: Amat Baik >= 91, Baik >= 76, Cukup >= 61, Sedang >= 51, Kurang < 51 |
| `grade` | VARCHAR(20), 5 predikat | Sesuai standar BKN: amat_baik, baik, cukup, sedang, kurang |
| `status` | VARCHAR(20), 6 status | Workflow: draft -> self_assessment -> peer_review -> supervisor_review -> completed -> approved |
| `assessor_type` | 3 tipe | self (guru sendiri), peer (rekan), supervisor (kepsek/wakasek) |
| `score` | NUMERIC(5,2), 0-4 | Skala per indikator: 1=Kurang, 2=Cukup, 3=Baik, 4=Amat Baik |
| `evidence` | TEXT, nullable | Bukti penilaian: deskripsi observasi, dokumen pendukung |
| `improvement_plan` | TEXT, nullable | Rencana perbaikan — menjadi input untuk PKB (S029) |

### Vernon Relationships

**teacher_evaluations:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `teacher` | belongs_to | **Ya** | Guru yang dievaluasi |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |

**teacher_evaluation_scores:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `evaluation` | belongs_to | **Ya** | Parent evaluation |
| `competency` | belongs_to | **Ya** | Indikator yang dinilai |
| `assessor` (teacher) | belongs_to | **Ya** | Penilai — bisa self/peer/supervisor |

### _rels / _data Structure

```json
// teacher_evaluations
{
  "_rels": {
    "teacher_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "teacher": {
      "id": "018f...",
      "full_name": "Bu Siti Aminah",
      "nip": "198501012010012001",
      "employee_type": "pns",
      "role": "guru_mapel"
    },
    "academic_year": {
      "id": "018f...",
      "name": "2025/2026"
    }
  }
}

// teacher_evaluation_scores
{
  "_rels": {
    "evaluation_id": "018f...",
    "competency_id": "018f...",
    "assessor_id": "018f..."
  },
  "_data": {
    "evaluation": {
      "id": "018f...",
      "semester": "ganjil",
      "status": "supervisor_review"
    },
    "competency": {
      "id": "018f...",
      "competency_area": "pedagogik",
      "indicator_code": "P1",
      "indicator_name": "Mengenal karakteristik peserta didik"
    },
    "assessor": {
      "id": "018f...",
      "full_name": "Pak Ahmad",
      "role": "kepala_sekolah"
    }
  }
}
```

### API Endpoints

```
# Competencies (Master)
GET    /api/v1/teacher-evaluation-competencies          — List indikator PKG
POST   /api/v1/teacher-evaluation-competencies          — Buat indikator
PUT    /api/v1/teacher-evaluation-competencies/{id}     — Update indikator

# Evaluations
GET    /api/v1/teachers/{id}/evaluations                — Riwayat PKG guru
GET    /api/v1/teacher-evaluations                      — Rekap PKG semua guru per semester
POST   /api/v1/teacher-evaluations                      — Buat evaluasi baru (init semester)
PUT    /api/v1/teacher-evaluations/{id}                 — Update evaluasi (recommendation, dll)
POST   /api/v1/teacher-evaluations/{id}/approve         — Approve oleh kepala sekolah

# Scores
GET    /api/v1/teacher-evaluations/{id}/scores          — Detail skor per kompetensi
POST   /api/v1/teacher-evaluation-scores/bulk           — Bulk input skor (per assessor)
PUT    /api/v1/teacher-evaluation-scores/{id}           — Update skor individual

# Workflow
POST   /api/v1/teacher-evaluations/{id}/submit-self     — Submit self-assessment
POST   /api/v1/teacher-evaluations/{id}/submit-peer     — Submit peer assessment
POST   /api/v1/teacher-evaluations/{id}/submit-supervisor — Submit supervisor assessment
POST   /api/v1/teacher-evaluations/{id}/calculate       — Calculate final scores
```

### Score Calculation

```sql
-- Step 1: Rata-rata skor per area (weighted by assessor_type)
-- Bobot: self=0 (referensi), peer=0.25, supervisor=0.75
SELECT
    c.competency_area,
    AVG(
        CASE s.assessor_type
            WHEN 'peer' THEN s.score * 0.25
            WHEN 'supervisor' THEN s.score * 0.75
        END
    ) AS weighted_avg
FROM teacher_evaluation_scores s
JOIN teacher_evaluation_competencies c ON c.id = s.competency_id
WHERE s.evaluation_id = $1 AND s.assessor_type != 'self'
GROUP BY c.competency_area;

-- Step 2: Konversi ke skala 100
-- score_area = (weighted_avg / max_score) * 100
-- total_score = AVG(score_pedagogik, score_kepribadian, score_sosial, score_profesional)

-- Step 3: Tentukan grade
-- >= 91: amat_baik, >= 76: baik, >= 61: cukup, >= 51: sedang, < 51: kurang
```

## Consequences

### Positive

- **Regulasi compliant**: 4 kompetensi PKG sesuai Permenneg PAN & RB dan Permendiknas.
- **Multi-assessor**: Self + peer + supervisor memberikan penilaian komprehensif.
- **Workflow terstruktur**: Status progres evaluasi dari draft sampai approved.
- **BKN/Dapodik ready**: Predikat dan skor bisa diekspor untuk SKP dan pelaporan.
- **Improvement tracking**: `improvement_plan` menjadi input untuk PKB (S029).
- **Pesantren compatible**: Evaluasi ustadz/ustadzah menggunakan tabel yang sama.

### Negative / Trade-offs

- **3 tabel**: Complexity tinggi — master, evaluasi, dan detail skor.
- **Bobot assessor**: 25% peer + 75% supervisor hardcoded — sekolah mungkin ingin bobot berbeda.
- **Self-assessment non-weighted**: Skor self hanya sebagai referensi — bisa membingungkan jika tidak dikomunikasikan.
- **Observasi kelas**: Tidak ada scheduling observasi — hanya evidence text di skor.
- **Annual cycle**: PKG per semester — beberapa sekolah mungkin hanya per tahun.

## Alternatives Considered

### 1. PKG sebagai JSONB di teacher record
- Ditolak: per semester, multi-assessor — volume dan complexity terlalu tinggi untuk JSONB.

### 2. Single score per evaluasi (tanpa detail per kompetensi)
- Ditolak: BKN/Dapodik memerlukan breakdown per kompetensi. Kepsek perlu tahu area mana yang lemah.

### 3. Separate table per assessor type
- Ditolak: 3 tabel assessment (self, peer, supervisor) terlalu banyak. Satu tabel scores dengan `assessor_type` lebih efisien.

### 4. External survey tool (Google Form, dll)
- Ditolak: data perlu terintegrasi dengan SKP, payroll, dan PKB. External tool memerlukan import/export manual.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `EvaluationDescriptor.TableName()` | -- | `"teacher_evaluations"` |
| U02 | `ScoreDescriptor.TableName()` | -- | `"teacher_evaluation_scores"` |
| U03 | `CompetencyDescriptor.TableName()` | -- | `"teacher_evaluation_competencies"` |
| U04 | Validate rejects invalid `competency_area` | `"technical"` | Error |
| U05 | Validate rejects invalid `assessor_type` | `"parent"` | Error |
| U06 | Validate rejects score > 4 | `score = 5` | Error |
| U07 | Validate rejects score < 0 | `score = -1` | Error |
| U08 | Validate rejects invalid `grade` | `"excellent"` | Error |
| U09 | Calculate grade: amat_baik | total_score = 92 | `grade = 'amat_baik'` |
| U10 | Calculate grade: baik | total_score = 80 | `grade = 'baik'` |
| U11 | Calculate grade: cukup | total_score = 65 | `grade = 'cukup'` |
| U12 | Calculate grade: kurang | total_score = 45 | `grade = 'kurang'` |
| U13 | Weighted score: peer 25% + supervisor 75% | peer=3, supervisor=4 | weighted = 3.75 |

### Integration Tests — Evaluation CRUD

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create evaluation | POST for teacher + semester | 201, status=draft |
| I02 | Unique per teacher+year+semester | Create duplicate | 409/422 |
| I03 | Submit self-assessment | POST /submit-self | 200, status=self_assessment |
| I04 | Submit peer assessment | POST /submit-peer | 200, status=peer_review |
| I05 | Submit supervisor assessment | POST /submit-supervisor | 200, status=supervisor_review |
| I06 | Calculate final scores | POST /calculate | 200, all score_* and grade populated |
| I07 | Approve evaluation | POST /approve by kepsek | 200, status=approved, approved_by set |

### Integration Tests — Scores

| # | Test Case | Action | Expected |
|---|---|---|---|
| I08 | Bulk input self-assessment | POST /scores/bulk with assessor_type=self | 201, all competencies scored |
| I09 | Bulk input peer assessment | POST /scores/bulk with assessor_type=peer | 201 |
| I10 | Unique per eval+competency+assessor | Insert duplicate score | 409/422 |
| I11 | Score range CHECK | INSERT with score=5 | DB error |
| I12 | Get evaluation detail | GET /evaluations/{id}/scores | 200, grouped by competency_area |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | TeacherUpdated syncs to evaluations | Update teacher full_name | `_data.teacher.full_name` updated |
| I14 | CompetencyUpdated syncs to scores | Update indicator_name | `_data.competency.indicator_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I15 | Cannot access other tenant's evaluations | GET with wrong tenant | 404 |
| I16 | Cannot score other company's evaluation | POST score with wrong company | Error, scope violation |
