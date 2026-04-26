# ADR-S024: Lesson Plan / RPP (Rencana Pelaksanaan Pembelajaran)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Rencana Pelaksanaan Pembelajaran (RPP) adalah **dokumen perencanaan mengajar** yang wajib dibuat oleh setiap guru sebelum mengajar. Dokumen ini menjadi acuan kegiatan belajar-mengajar dan merupakan salah satu indikator profesionalisme guru.

Sistem perencanaan pembelajaran di Indonesia saat ini berada dalam transisi:

1. **Kurikulum Merdeka — Modul Ajar**:
   - Format baru yang lebih ringkas dan fleksibel.
   - Komponen: Tujuan Pembelajaran (TP), Alur Tujuan Pembelajaran (ATP), Profil Pelajar Pancasila, kegiatan pembelajaran, asesmen formatif/sumatif, refleksi.
   - Guru diberi kebebasan menyusun sesuai kebutuhan siswa.

2. **K13 / KTSP — RPP Tradisional**:
   - Format standar: Kompetensi Inti (KI), Kompetensi Dasar (KD), Indikator, Tujuan, Materi, Metode, Langkah Pembelajaran (Pendahuluan/Inti/Penutup), Penilaian, Sumber/Media.
   - Bisa sangat panjang (10-20 halaman per pertemuan).

3. **Pesantren (ADR-009)**:
   - RPP untuk mata pelajaran diniyah: Fiqh, Nahwu, Kitab Kuning, Tahfidz.
   - Metode pengajaran khas: bandongan (guru baca, santri dengarkan), sorogan (santri baca, guru koreksi), halaqah (diskusi lingkaran).
   - Target hafalan per pertemuan untuk Tahfidz.

Kebutuhan:
- Satu guru bisa punya puluhan RPP per semester (per mapel × per kelas × per pertemuan/bab).
- Approval workflow: guru submit → wakasek kurikulum review → approve/revisi.
- File attachment: guru bisa upload RPP dalam format PDF sebagai lampiran.
- Linked ke curriculum (S019) untuk memastikan alignment.
- Linked ke subject (S020) dan class_room (ADR-011) untuk scope yang jelas.

### Mengapa Vernon Pattern?

- Relasi ke teacher (ADR-012), subject (S020), class_room (ADR-011), academic_year (ADR-010), curriculum (S019).
- Read-heavy: wakasek baca RPP untuk review, guru baca RPP saat mengajar, kepala sekolah baca untuk monitoring.
- Volume moderate: ~20-40 RPP per guru per semester.
- Data stabil setelah approved — jarang berubah mid-semester.
- SyncEngine: perubahan nama guru/mapel/kelas harus propagate ke `_data`.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `lesson_plans` (RPP/Modul Ajar utama) dan `lesson_plan_attachments` (file lampiran PDF).

### Table Schema

```sql
-- RPP / Modul Ajar
CREATE TABLE lesson_plans (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    subject_id      UUID NOT NULL,
    class_room_id   UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    curriculum_id   UUID,

    -- Identitas
    title           VARCHAR(300) NOT NULL,
    plan_type       VARCHAR(20) NOT NULL,
    semester        VARCHAR(10) NOT NULL,

    -- Scope pertemuan
    meeting_number  INT,
    topic           VARCHAR(300) NOT NULL,
    subtopic        TEXT,
    duration_minutes INT NOT NULL DEFAULT 90,

    -- Komponen RPP / Modul Ajar
    learning_objectives TEXT NOT NULL,
    learning_activities TEXT NOT NULL,
    assessment_plan TEXT,
    teaching_methods TEXT,
    media_and_resources TEXT,
    differentiation_notes TEXT,

    -- Komponen K13 spesifik (nullable untuk Merdeka)
    core_competency TEXT,
    basic_competency TEXT,
    indicators TEXT,

    -- Komponen Kurikulum Merdeka spesifik (nullable untuk K13)
    pancasila_profile TEXT,
    trigger_questions TEXT,
    reflection TEXT,

    -- Pesantren spesifik (ADR-009)
    kitab_reference VARCHAR(200),
    bab_fashl VARCHAR(200),
    teaching_method_pesantren VARCHAR(30),
    hafalan_target TEXT,

    -- Approval workflow
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    submitted_at    TIMESTAMPTZ,
    reviewed_by     UUID,
    reviewed_at     TIMESTAMPTZ,
    review_notes    TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_plan_type CHECK (plan_type IN (
        'modul_ajar', 'rpp_k13', 'rpp_ktsp', 'rpp_diniyah'
    )),
    CONSTRAINT chk_plan_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_plan_status CHECK (status IN (
        'draft', 'submitted', 'in_review', 'revision_needed', 'approved', 'archived'
    )),
    CONSTRAINT chk_duration CHECK (duration_minutes > 0 AND duration_minutes <= 480),
    CONSTRAINT chk_meeting_number CHECK (meeting_number IS NULL OR (meeting_number >= 1 AND meeting_number <= 100)),
    CONSTRAINT chk_pesantren_method CHECK (teaching_method_pesantren IS NULL OR teaching_method_pesantren IN (
        'bandongan', 'sorogan', 'halaqah', 'muhafadzah',
        'mudzakarah', 'ceramah', 'demonstrasi', 'tanya_jawab'
    ))
);

-- Indexes
CREATE INDEX idx_lesson_plan_tenant_company ON lesson_plans (tenant_id, company_id);
CREATE INDEX idx_lesson_plan_teacher ON lesson_plans (teacher_id);
CREATE INDEX idx_lesson_plan_subject ON lesson_plans (subject_id);
CREATE INDEX idx_lesson_plan_class ON lesson_plans (class_room_id);
CREATE INDEX idx_lesson_plan_year_semester ON lesson_plans (academic_year_id, semester);
CREATE INDEX idx_lesson_plan_curriculum ON lesson_plans (curriculum_id) WHERE curriculum_id IS NOT NULL;
CREATE INDEX idx_lesson_plan_type ON lesson_plans (plan_type);
CREATE INDEX idx_lesson_plan_status ON lesson_plans (status);
CREATE INDEX idx_lesson_plan_reviewed_by ON lesson_plans (reviewed_by) WHERE reviewed_by IS NOT NULL;
CREATE INDEX idx_lesson_plan_rels ON lesson_plans USING GIN (_rels);
CREATE INDEX idx_lesson_plan_data ON lesson_plans USING GIN (_data);

-- File lampiran RPP
CREATE TABLE lesson_plan_attachments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign key
    lesson_plan_id  UUID NOT NULL,

    -- File info
    file_name       VARCHAR(255) NOT NULL,
    file_path       TEXT NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    mime_type       VARCHAR(100) NOT NULL DEFAULT 'application/pdf',
    file_type       VARCHAR(20) NOT NULL DEFAULT 'document',

    -- Metadata
    description     VARCHAR(300),
    uploaded_by     UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_file_size CHECK (file_size_bytes > 0 AND file_size_bytes <= 52428800),
    CONSTRAINT chk_file_type CHECK (file_type IN ('document', 'image', 'presentation', 'spreadsheet', 'other')),
    CONSTRAINT chk_mime_type CHECK (mime_type IN (
        'application/pdf',
        'application/msword',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        'application/vnd.openxmlformats-officedocument.presentationml.presentation',
        'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'image/jpeg', 'image/png', 'image/webp'
    ))
);

-- Indexes
CREATE INDEX idx_lp_attach_tenant_company ON lesson_plan_attachments (tenant_id, company_id);
CREATE INDEX idx_lp_attach_plan ON lesson_plan_attachments (lesson_plan_id);
CREATE INDEX idx_lp_attach_uploaded_by ON lesson_plan_attachments (uploaded_by);
CREATE INDEX idx_lp_attach_rels ON lesson_plan_attachments USING GIN (_rels);
CREATE INDEX idx_lp_attach_data ON lesson_plan_attachments USING GIN (_data);
```

### Field Design Rationale

**lesson_plans:**

| Field | Keputusan | Alasan |
|---|---|---|
| `plan_type` | 4 tipe | modul_ajar (Merdeka), rpp_k13, rpp_ktsp, rpp_diniyah (pesantren) — cover semua kurikulum aktif |
| `meeting_number` | INT, nullable, 1-100 | Nomor pertemuan — nullable jika RPP per bab/unit (bukan per pertemuan) |
| `topic` / `subtopic` | VARCHAR + TEXT | Topik utama wajib, subtopik opsional — untuk index dan display |
| `learning_objectives` | TEXT, NOT NULL | Tujuan Pembelajaran (TP) — komponen wajib di semua format kurikulum |
| `learning_activities` | TEXT, NOT NULL | Langkah pembelajaran — komponen wajib: pendahuluan/inti/penutup (K13) atau kegiatan inti (Merdeka) |
| `assessment_plan` | TEXT, nullable | Rencana penilaian — formatif + sumatif. Nullable karena tidak semua pertemuan ada asesmen |
| `teaching_methods` | TEXT, nullable | Metode: diskusi, ceramah, praktikum, project-based, dll. |
| `core_competency` / `basic_competency` / `indicators` | TEXT, nullable | Komponen K13 spesifik — nullable untuk Kurikulum Merdeka yang tidak pakai KI/KD |
| `pancasila_profile` / `trigger_questions` / `reflection` | TEXT, nullable | Komponen Merdeka spesifik — nullable untuk K13 |
| `kitab_reference` / `bab_fashl` | VARCHAR, nullable | Referensi kitab untuk pesantren: nama kitab + bab/fashl yang diajarkan |
| `teaching_method_pesantren` | 8 metode | bandongan, sorogan, halaqah, muhafadzah (hafalan), mudzakarah (diskusi), ceramah, demonstrasi, tanya jawab |
| `hafalan_target` | TEXT, nullable | Target hafalan per pertemuan — surah/ayat/halaman untuk Tahfidz |
| `status` | 6 stage | draft → submitted → in_review → revision_needed/approved → archived |
| `reviewed_by` | UUID, nullable | Wakasek kurikulum yang mereview — for audit trail |
| `duration_minutes` | INT, max 480 | Durasi pertemuan dalam menit — default 90 (2 jam pelajaran × 45 menit) |

**lesson_plan_attachments:**

| Field | Keputusan | Alasan |
|---|---|---|
| `file_path` | TEXT | Path ke file storage (S3/MinIO) — bukan simpan file di DB |
| `file_size_bytes` | BIGINT, max 50MB | Batas ukuran file 50MB — cukup untuk RPP PDF/DOCX |
| `mime_type` | 7 tipe | PDF, Word (doc/docx), PowerPoint, Excel, JPEG, PNG, WebP — format umum RPP |
| `file_type` | 5 kategori | document (PDF/Word), image, presentation (PPT), spreadsheet (Excel), other |
| `uploaded_by` | UUID, NOT NULL | Guru yang upload — bisa beda dari teacher_id jika admin yang upload |

### Approval Workflow

```
Guru                    Wakasek Kurikulum       Status
─────                   ────────────────        ──────
Buat RPP draft          —                       draft
Submit RPP         →    Terima notifikasi       submitted
—                       Mulai review            in_review
—                       Butuh revisi       →    revision_needed (+ review_notes)
Revisi + re-submit →    —                       submitted (ulang)
—                       Approve            →    approved
—                       —                       archived (akhir tahun)
```

Approval bersifat **per lesson plan** — wakasek kurikulum bisa approve satu per satu atau batch.

### Vernon Relationships

**lesson_plans:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `teacher` | belongs_to | **Ya** | Nama guru penyusun RPP |
| `subject` | belongs_to | **Ya** | Mata pelajaran |
| `class_room` | belongs_to | **Ya** | Kelas target |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |
| `curriculum` | belongs_to | **Tidak** | Opsional — load on demand |
| `reviewer` (reviewed_by) | belongs_to | **Tidak** | Load on demand — hanya relevan saat review |

**lesson_plan_attachments:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `lesson_plan` | belongs_to | **Ya** | Konteks RPP induk |
| `uploader` (uploaded_by) | belongs_to | **Tidak** | Load on demand |

### _rels / _data Structure

**lesson_plans:**
```json
{
  "_rels": {
    "teacher_id": "018f...",
    "subject_id": "018f...",
    "class_room_id": "018f...",
    "academic_year_id": "018f...",
    "curriculum_id": "018f...",
    "reviewed_by": "018f..."
  },
  "_data": {
    "teacher": { "id": "018f...", "full_name": "Bu Siti", "nip": "198501012010012001" },
    "subject": { "id": "018f...", "name": "Matematika", "code": "MTK" },
    "class_room": { "id": "018f...", "name": "VII-A" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

**lesson_plan_attachments:**
```json
{
  "_rels": {
    "lesson_plan_id": "018f...",
    "uploaded_by": "018f..."
  },
  "_data": {
    "lesson_plan": {
      "id": "018f...",
      "title": "Modul Ajar Pecahan Kelas VII",
      "plan_type": "modul_ajar",
      "subject_name": "Matematika",
      "teacher_name": "Bu Siti"
    }
  }
}
```

### API Endpoints

```
# Lesson Plans
GET    /api/v1/lesson-plans?teacher_id={id}&semester=ganjil               — List RPP per guru
GET    /api/v1/lesson-plans?subject_id={id}&class_id={id}&semester=ganjil — List RPP per mapel per kelas
GET    /api/v1/lesson-plans?status=submitted                              — RPP menunggu review (wakasek)
GET    /api/v1/lesson-plans/{id}                                          — Detail RPP
POST   /api/v1/lesson-plans                                               — Buat RPP
PUT    /api/v1/lesson-plans/{id}                                          — Update RPP (hanya draft/revision_needed)
DELETE /api/v1/lesson-plans/{id}                                          — Hapus RPP (hanya draft)

# Approval Workflow
POST   /api/v1/lesson-plans/{id}/submit                                   — Guru submit RPP
POST   /api/v1/lesson-plans/{id}/review                                   — Wakasek mulai review
POST   /api/v1/lesson-plans/{id}/approve                                  — Wakasek approve
POST   /api/v1/lesson-plans/{id}/request-revision                         — Wakasek minta revisi
  Body: { "review_notes": "Tambahkan asesmen formatif di bagian penutup" }
POST   /api/v1/lesson-plans/bulk-approve                                  — Bulk approve (batch)
  Body: { "lesson_plan_ids": ["018f...", "018f...", "018f..."] }

# Attachments
GET    /api/v1/lesson-plans/{id}/attachments                              — List lampiran
POST   /api/v1/lesson-plans/{id}/attachments                              — Upload file (multipart/form-data)
DELETE /api/v1/lesson-plan-attachments/{id}                               — Hapus lampiran
GET    /api/v1/lesson-plan-attachments/{id}/download                      — Download file

# Statistics (Wakasek Dashboard)
GET    /api/v1/lesson-plans/statistics?year_id={id}&semester=ganjil
  → { "total": 120, "draft": 15, "submitted": 8, "approved": 90, "revision_needed": 7 }

# Carry-forward (Salin RPP ke semester/tahun baru)
POST   /api/v1/lesson-plans/{id}/duplicate
  Body: { "target_semester": "genap", "target_class_room_id": "018f..." }
  → Copy RPP content, reset status to draft
```

## Consequences

### Positive

- **Multi-kurikulum**: Mendukung Modul Ajar (Merdeka), RPP K13, RPP KTSP, dan RPP Diniyah (pesantren) dalam satu tabel.
- **Approval workflow**: Wakasek kurikulum bisa review dan approve/revisi — sesuai prosedur sekolah.
- **Pesantren support**: Metode pengajaran khas pesantren (bandongan, sorogan, halaqah), referensi kitab, dan target hafalan.
- **File attachment**: Guru bisa upload RPP PDF/DOCX yang sudah ada — tidak harus re-type.
- **Audit trail**: `reviewed_by`, `reviewed_at`, `review_notes` memberikan jejak review lengkap.
- **Linked to curriculum**: Opsional link ke S019 untuk memastikan RPP align dengan kurikulum.
- **Duplicate/carry-forward**: Guru bisa copy RPP ke semester/kelas lain — hemat waktu.

### Negative / Trade-offs

- **Banyak nullable fields**: Komponen K13 nullable untuk Merdeka dan sebaliknya — ~6 nullable TEXT fields. Trade-off untuk single table vs separate tables per kurikulum.
- **TEXT fields untuk konten**: `learning_objectives`, `learning_activities`, dll. disimpan sebagai plain TEXT — tidak structured. Trade-off: RPP format sangat bervariasi antar guru dan kurikulum.
- **File storage external**: `file_path` mengandalkan S3/MinIO — butuh infrastructure terpisah.
- **Approval bottleneck**: Wakasek harus review semua RPP — bisa jadi bottleneck jika banyak guru. Bulk approve membantu tapi tidak menghilangkan masalah.
- **No versioning**: Saat revisi, konten lama di-overwrite — tidak ada version history. Bisa ditambah jika dibutuhkan.
- **Pesantren method hardcoded**: 8 metode di CHECK constraint — mungkin ada metode lain yang belum tercakup.

## Alternatives Considered

### 1. RPP sebagai pure file upload tanpa structured fields
- Ditolak: tidak bisa search by topic, tidak bisa aggregate statistics, tidak bisa enforce komponen wajib.

### 2. Tabel terpisah per kurikulum (modul_ajar, rpp_k13, rpp_diniyah)
- Ditolak: 80% fields sama (teacher, subject, class, topic, objectives, activities) — duplikasi schema. Satu tabel dengan nullable kurikulum-specific fields lebih maintainable.

### 3. Konten RPP sebagai JSONB
- Deferred: JSONB memberikan flexibility (schema-less), tapi kehilangan NOT NULL constraint dan text search capability. Untuk MVP, TEXT fields cukup.

### 4. Git-like versioning untuk revisi
- Deferred: version control untuk RPP terlalu complex untuk MVP. Simple overwrite + review_notes audit trail sudah memenuhi kebutuhan.

### 5. Template RPP per kurikulum
- Deferred: template yang bisa di-generate (auto-fill KI/KD dari kurikulum) adalah enhancement yang bagus tapi complex. MVP fokus pada input manual.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `LessonPlanDescriptor.TableName()` | — | `"lesson_plans"` |
| U02 | `AttachmentDescriptor.TableName()` | — | `"lesson_plan_attachments"` |
| U03 | Validate rejects invalid `plan_type` | `"syllabus"` | Error |
| U04 | Validate rejects invalid `status` | `"rejected"` | Error |
| U05 | Validate rejects `duration_minutes = 0` | `0` | Error |
| U06 | Validate rejects `duration_minutes > 480` | `500` | Error |
| U07 | Validate rejects invalid `teaching_method_pesantren` | `"lecture"` | Error |
| U08 | Validate rejects invalid `semester` | `"summer"` | Error |
| U09 | Validate accepts valid modul ajar | All Merdeka fields + K13 nullable | No error |
| U10 | Validate accepts valid RPP K13 | All K13 fields + Merdeka nullable | No error |

### Integration Tests — Lesson Plans

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create lesson plan (Modul Ajar) | POST with plan_type=modul_ajar | 201 |
| I02 | Create lesson plan (RPP K13) | POST with plan_type=rpp_k13, KI/KD filled | 201 |
| I03 | Create RPP diniyah (pesantren) | POST with plan_type=rpp_diniyah, kitab_reference | 201 |
| I04 | Plan type CHECK | INSERT with `plan_type = 'syllabus'` | DB error |
| I05 | Status CHECK | INSERT with `status = 'rejected'` | DB error |
| I06 | Duration CHECK | INSERT with `duration_minutes = 500` | DB error |
| I07 | Pesantren method CHECK | INSERT with `teaching_method_pesantren = 'lecture'` | DB error |
| I08 | Update draft RPP | PUT on status=draft | 200 |
| I09 | Cannot update approved RPP | PUT on status=approved | 403 |

### Integration Tests — Approval Workflow

| # | Test Case | Action | Expected |
|---|---|---|---|
| I10 | Submit RPP | POST /lesson-plans/{id}/submit | 200, status=submitted, submitted_at set |
| I11 | Cannot submit non-draft | POST /submit on status=approved | 422 |
| I12 | Start review | POST /lesson-plans/{id}/review | 200, status=in_review |
| I13 | Approve RPP | POST /lesson-plans/{id}/approve | 200, status=approved, reviewed_by set |
| I14 | Request revision | POST /lesson-plans/{id}/request-revision | 200, status=revision_needed, review_notes set |
| I15 | Re-submit after revision | Update + POST /submit on revision_needed | 200, status=submitted |
| I16 | Bulk approve | POST /bulk-approve with 5 IDs | 200, all 5 approved |

### Integration Tests — Attachments

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Upload PDF attachment | POST multipart with PDF | 201, file_path set |
| I18 | File size CHECK | Upload > 50MB | 413/422 |
| I19 | MIME type CHECK | Upload .exe file | 422, invalid mime type |
| I20 | Download attachment | GET /attachments/{id}/download | 200, file stream |
| I21 | Delete attachment | DELETE /lesson-plan-attachments/{id} | 200 |

### Integration Tests — Pesantren (ADR-009)

| # | Test Case | Action | Expected |
|---|---|---|---|
| I22 | RPP with kitab reference | POST with kitab_reference="Fathul Qarib" | 201 |
| I23 | RPP with sorogan method | POST with teaching_method_pesantren=sorogan | 201 |
| I24 | RPP with hafalan target | POST with hafalan_target="Al-Baqarah ayat 1-10" | 201 |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I25 | TeacherUpdated syncs to plans | Update teacher name | `_data.teacher.full_name` updated |
| I26 | SubjectUpdated syncs to plans | Update subject name | `_data.subject.name` updated |
| I27 | ClassRoomUpdated syncs to plans | Update class name | `_data.class_room.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I28 | Cannot access other tenant's plans | GET with wrong tenant | 404 |
| I29 | Cannot approve other tenant's plan | POST /approve with wrong tenant | Error |
