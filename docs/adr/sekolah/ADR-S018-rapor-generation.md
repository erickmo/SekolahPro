# ADR-S018: Rapor Generation

**Status**: Approved (BUILD NOW — highest priority)
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Rapor adalah **output paling penting** dari seluruh sistem manajemen sekolah. Ini adalah dokumen resmi yang diterima orang tua setiap akhir semester — dan sering menjadi **satu-satunya bukti tangible** bahwa software ini berfungsi.

Rapor bukan domain data baru — rapor adalah **orchestration layer** yang menggabungkan data dari domain yang sudah ada:

| Sumber | Data |
|--------|------|
| S001 Student | Biodata siswa (nama, NIS, NISN) |
| S003 Guardian | Nama orang tua/wali |
| S004 Academic Record | Rekap absensi, grade average, ranking, promotion status |
| S008 Attendance | Detail kehadiran (aggregated) |
| S011 Subject Grades | Nilai per mata pelajaran (knowledge, skill, attitude, deskripsi) |
| S012 Discipline | Catatan perilaku/kedisiplinan |
| S015 Extracurricular | Penilaian ekskul (grade + deskripsi) |
| S014 Class Placement | Kelas dan wali kelas |

Rapor yang perlu didukung:

1. **Kurikulum Merdeka**: Nilai kompetensi + deskripsi naratif per mapel + Profil Pelajar Pancasila.
2. **Kurikulum 2013 (K13)**: KI-1 (Spiritual), KI-2 (Sosial), KI-3 (Pengetahuan), KI-4 (Keterampilan).
3. **Pesantren**: Rapor diniyah (mata pelajaran keagamaan) — bisa sebagai rapor terpisah atau gabungan.

Karakteristik:
- **Batch generation**: 500 rapor harus bisa di-generate dalam satu operasi di akhir semester.
- **Template-based**: Setiap sekolah punya layout rapor yang sedikit berbeda.
- **PDF output**: Format final adalah PDF (cetak atau digital).
- **Digital signature**: Tanda tangan kepsek + wali kelas (image stamp).
- **Immutable setelah finalize**: Rapor yang sudah di-finalize tidak boleh berubah.

### Mengapa Vernon Pattern + Background Job?

- Rapor generation = heavy operation (query banyak tabel, render PDF).
- Result = immutable snapshot yang di-cache.
- Read-heavy setelah generate: cetak ulang, download.
- Butuh background job system untuk batch generation.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `rapor_templates` (template layout) dan `rapor_records` (rapor per siswa per semester). Generation dilakukan via **background job**.

### Table Schema

```sql
-- Template rapor per sekolah
CREATE TABLE rapor_templates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    curriculum_type VARCHAR(20) NOT NULL,
    grade_levels    JSONB NOT NULL DEFAULT '[]',
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Layout
    header_config   JSONB NOT NULL DEFAULT '{}',
    body_config     JSONB NOT NULL DEFAULT '{}',
    footer_config   JSONB NOT NULL DEFAULT '{}',
    signature_config JSONB NOT NULL DEFAULT '{}',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_curriculum_type CHECK (curriculum_type IN ('merdeka', 'k13', 'diniyah', 'custom'))
);

-- Indexes
CREATE INDEX idx_rapor_tpl_tenant_company ON rapor_templates (tenant_id, company_id);

-- Rapor per siswa per semester (snapshot)
CREATE TABLE rapor_records (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    class_room_id   UUID NOT NULL,
    template_id     UUID NOT NULL,

    -- Periode
    semester        VARCHAR(10) NOT NULL,

    -- Snapshot data (frozen at generation time)
    student_snapshot    JSONB NOT NULL,
    guardian_snapshot    JSONB NOT NULL DEFAULT '{}',
    grades_snapshot     JSONB NOT NULL DEFAULT '[]',
    attendance_snapshot JSONB NOT NULL DEFAULT '{}',
    extracurricular_snapshot JSONB NOT NULL DEFAULT '[]',
    discipline_snapshot JSONB NOT NULL DEFAULT '{}',
    academic_snapshot   JSONB NOT NULL DEFAULT '{}',

    -- Catatan
    teacher_note    TEXT,
    principal_note  TEXT,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',

    -- PDF
    pdf_url         TEXT,
    generated_at    TIMESTAMPTZ,

    -- Tanda tangan
    teacher_signature_url   TEXT,
    principal_signature_url TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_rapor_student_semester UNIQUE (student_id, academic_year_id, semester),
    CONSTRAINT chk_rapor_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_rapor_status CHECK (status IN ('draft', 'generated', 'reviewed', 'finalized'))
);

-- Indexes
CREATE INDEX idx_rapor_tenant_company ON rapor_records (tenant_id, company_id);
CREATE INDEX idx_rapor_student ON rapor_records (student_id);
CREATE INDEX idx_rapor_class_semester ON rapor_records (class_room_id, academic_year_id, semester);
CREATE INDEX idx_rapor_status ON rapor_records (status);
CREATE INDEX idx_rapor_rels ON rapor_records USING GIN (_rels);
CREATE INDEX idx_rapor_data ON rapor_records USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `curriculum_type` | 4 tipe template | Merdeka, K13, Diniyah (pesantren), Custom |
| `grade_levels` | JSONB array | Template berlaku untuk grade tertentu: `["7","8","9"]` |
| `*_config` | JSONB | Layout configurable: header (logo, nama sekolah), body (kolom nilai), footer, signatures |
| `*_snapshot` | JSONB | **Frozen data** — data di-snapshot saat generate, tidak berubah meski data asli berubah |
| `status` | 4 stage lifecycle | draft → generated → reviewed → finalized |
| `pdf_url` | TEXT, nullable | URL ke PDF di object storage — diisi setelah generate |
| `teacher_note` | TEXT, nullable | Catatan wali kelas (Ananda Ahmad perlu lebih fokus...) |
| `principal_note` | TEXT, nullable | Catatan kepala sekolah |

### Snapshot Strategy (Critical Design Decision)

Rapor adalah **dokumen hukum** — nilainya tidak boleh berubah setelah dicetak. Oleh karena itu, semua data di-snapshot ke JSONB saat generation:

```json
{
  "student_snapshot": {
    "full_name": "Ahmad Fadhil",
    "nis": "12345",
    "nisn": "0012345678",
    "gender": "L",
    "birth_place": "Jakarta",
    "birth_date": "2012-05-15",
    "class_room": "VII-A",
    "academic_year": "2025/2026"
  },
  "guardian_snapshot": {
    "father_name": "Budi Santoso",
    "mother_name": "Siti Aminah"
  },
  "grades_snapshot": [
    {
      "subject": "Matematika",
      "subject_group": "matematika",
      "score_knowledge": 85.50,
      "score_skill": 88.00,
      "score_attitude": "B",
      "final_score": 86.75,
      "grade_letter": "A",
      "description_knowledge": "Ahmad menunjukkan pemahaman yang baik...",
      "description_skill": "Ahmad mampu menyelesaikan soal..."
    }
  ],
  "attendance_snapshot": {
    "days_present": 95,
    "days_sick": 3,
    "days_permitted": 1,
    "days_absent": 1
  },
  "extracurricular_snapshot": [
    {
      "name": "Pramuka",
      "grade": "B",
      "description": "Aktif mengikuti kegiatan..."
    }
  ],
  "discipline_snapshot": {
    "violation_points": 5,
    "merit_points": 15,
    "note": "Perilaku baik, aktif membantu teman."
  },
  "academic_snapshot": {
    "grade_average": 86.75,
    "class_rank": 5,
    "total_students": 32,
    "promotion_status": "promoted"
  }
}
```

**Kenapa snapshot, bukan reference?**
1. Jika guru mengoreksi nilai setelah rapor dicetak → rapor lama tidak berubah.
2. Rapor bisa dicetak ulang kapan saja tanpa query 7+ tabel.
3. Rapor adalah arsip — harus stabil meski data source berubah.

### Generation Flow

```
1. Admin pilih: academic_year + semester + class_room (atau "semua kelas")
2. System validates: semua nilai per mapel sudah lengkap? Absensi sudah ter-aggregate?
   - Jika belum lengkap → warning: "15 siswa belum ada nilai Matematika"
   - Admin bisa force generate (draft) atau tunggu data lengkap
3. Background job per siswa:
   a. Query S001, S003 → student_snapshot, guardian_snapshot
   b. Query S011 → grades_snapshot (ordered by subject_group + sort_order)
   c. Query S004 or aggregate S008 → attendance_snapshot
   d. Query S015 → extracurricular_snapshot
   e. Query S012 → discipline_snapshot (aggregate points + summary)
   f. Query S004 → academic_snapshot (rank, average, promotion)
   g. Render PDF using template
   h. Upload PDF to object storage
   i. Update rapor_record with pdf_url + status = 'generated'
4. Wali kelas review → tambah teacher_note → status = 'reviewed'
5. Kepsek approve → status = 'finalized' → rapor locked
```

### Batch Generation Architecture

```
POST /api/v1/rapor/generate
  Body: { academic_year_id, semester, class_room_ids: [...] }

→ Creates batch job
→ Returns job_id for polling

GET /api/v1/rapor/jobs/{job_id}
  → { status: "processing", progress: 45, total: 120, completed: 54 }

→ Each student = 1 task in the batch
→ Parallelism: 10 concurrent workers
→ Estimated: 500 rapor in ~5 minutes
```

### Vernon Relationships

**rapor_records:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Pemilik rapor |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |
| `class_room` | belongs_to | **Ya** | Kelas |
| `template` | belongs_to | **Ya** | Template yang digunakan |

### API Endpoints

```
# Templates
GET    /api/v1/rapor-templates                       — List template
POST   /api/v1/rapor-templates                       — Buat template
PUT    /api/v1/rapor-templates/{id}                  — Update template

# Generation
POST   /api/v1/rapor/generate                        — Generate batch (background job)
GET    /api/v1/rapor/jobs/{id}                       — Cek progress batch
POST   /api/v1/rapor/preview/{student_id}            — Preview rapor 1 siswa (tanpa save)

# Records
GET    /api/v1/students/{id}/rapor                   — Rapor siswa (semua semester)
GET    /api/v1/rapor-records/{id}                    — Detail rapor
PUT    /api/v1/rapor-records/{id}                    — Update notes (sebelum finalize)
POST   /api/v1/rapor-records/{id}/review             — Wali kelas review
POST   /api/v1/rapor-records/{id}/finalize           — Kepsek finalize (lock)
GET    /api/v1/rapor-records/{id}/pdf                — Download PDF
POST   /api/v1/rapor/bulk-download                   — Download batch (ZIP)

# Completeness Check
GET    /api/v1/rapor/readiness?year_id={id}&semester=ganjil&class_id={id}
  → { ready: false, missing: [{ student: "Ahmad", missing: ["Matematika", "B.Inggris"] }] }
```

## Consequences

### Positive

- **Deal closer**: "Bisa cetak rapor?" → Yes → langsung close.
- **Snapshot immutable**: Rapor tidak berubah setelah finalize — integritas dokumen terjaga.
- **Multi-kurikulum**: Template system mendukung Merdeka, K13, dan Diniyah.
- **Batch efficient**: 500 rapor dalam ~5 menit dengan background job.
- **Readiness check**: Admin tahu persis data apa yang belum lengkap sebelum generate.
- **Cetak ulang aman**: PDF tersimpan di object storage, bisa diakses kapan saja.
- **Zero JOIN saat cetak**: Semua data di-snapshot ke JSONB — render PDF = baca 1 row.

### Negative / Trade-offs

- **Storage cost**: Snapshot JSONB + PDF per siswa per semester. 500 siswa × 2 semester = 1000 PDFs/tahun.
- **Snapshot staleness**: Jika data dikoreksi setelah generate, rapor harus di-regenerate (bukan auto-update).
- **Template complexity**: JSONB config untuk layout rapor memerlukan UI builder atau manual JSON editing.
- **PDF rendering**: Server-side PDF generation memerlukan library (wkhtmltopdf/Chromium headless) — setup awal non-trivial.
- **Background job**: Memerlukan job queue system (bisa in-memory untuk MVP, NATS untuk scale).

## Alternatives Considered

### 1. Client-side PDF generation (browser print)
- Ditolak: inkonsisten antar browser, tidak bisa batch, tidak bisa cetak ulang dari server.

### 2. Rapor tanpa snapshot (query langsung dari source tables)
- Ditolak: jika nilai dikoreksi setelah rapor dicetak, rapor lama ikut berubah — ini ilegal untuk dokumen resmi.

### 3. Rapor sebagai fitur e-Rapor Kemendikbud
- Ditolak: e-Rapor Kemendikbud terbatas dan tidak customizable. Sekolah swasta butuh rapor sendiri dengan branding sekolah.

### 4. Third-party report generator (JasperReports, etc.)
- Ditolak: menambah dependency besar, learning curve tinggi, sulit diintegrasikan dengan Vernon pattern.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `TemplateDescriptor.TableName()` | — | `"rapor_templates"` |
| U02 | `RecordDescriptor.TableName()` | — | `"rapor_records"` |
| U03 | Validate rejects invalid `curriculum_type` | `"cbsa"` | Error |
| U04 | Validate rejects invalid `status` | `"published"` | Error |
| U05 | Validate rejects invalid `semester` | `"midterm"` | Error |
| U06 | Snapshot builder builds correct structure | Student + grades data | Valid snapshot JSON |
| U07 | Snapshot includes all grade fields | Grade with all scores | All fields present in snapshot |

### Integration Tests — Templates

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create rapor template | POST with curriculum_type + configs | 201 |
| I02 | Curriculum type CHECK | INSERT with `curriculum_type = 'cbsa'` | DB error |
| I03 | List active templates | GET /rapor-templates?active=true | 200, only active |

### Integration Tests — Generation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I04 | Generate rapor for 1 student | POST /rapor/preview/{id} | 200, returns preview data |
| I05 | Generate batch for class | POST /rapor/generate with class_room_id | 202, job created |
| I06 | Check generation progress | GET /rapor/jobs/{id} | 200, progress updated |
| I07 | Generated rapor has snapshot | Check rapor_record after generation | All *_snapshot fields populated |
| I08 | Generated rapor has PDF | Check pdf_url | Not null, valid URL |
| I09 | Unique per student+year+semester | Generate same rapor twice | Overwrites draft, rejects if finalized |

### Integration Tests — Readiness Check

| # | Test Case | Action | Expected |
|---|---|---|---|
| I10 | All data complete | GET /rapor/readiness | `{ ready: true }` |
| I11 | Missing grades | Student has no Matematika grade | `{ ready: false, missing: [...] }` |
| I12 | Missing attendance | Student has no attendance aggregate | Warning in response |

### Integration Tests — Workflow

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | Wali kelas review | POST /rapor-records/{id}/review with note | 200, status → reviewed |
| I14 | Kepsek finalize | POST /rapor-records/{id}/finalize | 200, status → finalized |
| I15 | Cannot edit finalized rapor | PUT /rapor-records/{id} on finalized | 403, immutable |
| I16 | Can regenerate draft | POST /rapor/generate on draft rapor | 200, re-snapshot |
| I17 | Cannot regenerate finalized | POST /rapor/generate on finalized | 422, must create new draft |

### Integration Tests — PDF

| # | Test Case | Action | Expected |
|---|---|---|---|
| I18 | Download single PDF | GET /rapor-records/{id}/pdf | 200, PDF file |
| I19 | Bulk download ZIP | POST /rapor/bulk-download with IDs | 202, job → ZIP URL |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I20 | StudentUpdated does NOT change snapshot | Update student name | Snapshot unchanged (immutable) |
| I21 | StudentUpdated syncs _data (non-snapshot) | Update student name | `_data.student.full_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I22 | Cannot access other tenant's rapor | GET with wrong tenant | 404 |
| I23 | Cannot download other tenant's PDF | GET PDF with wrong tenant | 404/403 |
