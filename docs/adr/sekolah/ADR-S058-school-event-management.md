# ADR-S058: School Event Management (Manajemen Kegiatan Sekolah)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Sekolah menyelenggarakan berbagai kegiatan sepanjang tahun ajaran — dari upacara bendera mingguan hingga wisuda tahunan. Pengelolaan kegiatan sekolah mencakup perencanaan, persetujuan, pelaksanaan, dan dokumentasi. Tanpa sistem terpusat, informasi kegiatan tersebar di grup WhatsApp, spreadsheet, dan catatan manual.

Kebutuhan manajemen kegiatan:

1. **Beragam tipe kegiatan**: Upacara bendera, class meeting, study tour, field trip, pentas seni, wisuda, rapat orang tua, workshop guru.
2. **Planning workflow**: Proposal kegiatan → approval oleh kepsek (S047) → pelaksanaan → dokumentasi.
3. **Registration & attendance**: Peserta mendaftar dan kehadiran dicatat — terutama untuk kegiatan opsional.
4. **Budget linkage**: Setiap kegiatan punya anggaran yang terhubung dengan RKAS (S050).
5. **Calendar integration**: Kegiatan masuk ke kalender akademik (S023) — menghindari bentrok jadwal.
6. **Dokumentasi**: Foto, laporan, dan outcomes dicatat sebagai bukti pelaksanaan (penting untuk akreditasi).
7. **Pesantren**: Kegiatan pesantren meliputi haflah (perayaan), khataman (tamat Al-Quran), dan wisuda tahfidz.

### Mengapa Vernon Pattern?

- Read-heavy: guru, siswa, dan orang tua sering melihat jadwal kegiatan, detail, dan dokumentasi.
- Relasi ke academic_year, class, teacher, budget (S050), calendar (S023).
- Write periodik: kegiatan di-plan per semester, execution harian-mingguan.
- Business logic moderate: workflow approval, budget allocation, attendance tracking.
- Eventual consistency acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `school_events` (kegiatan), `school_event_participants` (peserta/kehadiran), dan `school_event_documents` (dokumentasi).

### Table Schema

```sql
-- Kegiatan sekolah
CREATE TABLE school_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    event_type      VARCHAR(30) NOT NULL,
    event_category  VARCHAR(20) NOT NULL,

    -- Jadwal
    academic_year_id UUID NOT NULL,
    semester        INT,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    start_time      TIME,
    end_time        TIME,

    -- Lokasi
    location        VARCHAR(255),
    is_outdoor      BOOLEAN NOT NULL DEFAULT false,
    venue_notes     TEXT,

    -- Peserta target
    target_audience VARCHAR(20) NOT NULL,
    target_classes  JSONB NOT NULL DEFAULT '[]',
    is_mandatory    BOOLEAN NOT NULL DEFAULT true,
    max_participants INT,

    -- Penyelenggara
    organizer_id    UUID NOT NULL,
    organizer_name  VARCHAR(255) NOT NULL,
    committee       JSONB NOT NULL DEFAULT '[]',

    -- Budget (link ke S050)
    budget_item_id  UUID,
    estimated_budget BIGINT,
    actual_budget   BIGINT,

    -- Approval workflow (link ke S047)
    approval_status VARCHAR(20) NOT NULL DEFAULT 'draft',
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,
    approval_note   TEXT,

    -- Execution
    execution_status VARCHAR(20) NOT NULL DEFAULT 'planned',
    participant_count INT NOT NULL DEFAULT 0,
    attendance_count INT NOT NULL DEFAULT 0,

    -- Calendar link (S023)
    calendar_event_id UUID,

    -- Outcomes
    outcomes        TEXT,
    lessons_learned TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_event_type CHECK (event_type IN (
        'upacara_bendera', 'class_meeting', 'study_tour', 'field_trip',
        'pentas_seni', 'graduation_ceremony', 'parent_meeting', 'teacher_workshop',
        'sports_day', 'science_fair', 'literacy_day', 'independence_day',
        'haflah', 'khataman', 'wisuda_tahfidz', 'pesantren_kilat',
        'isra_miraj', 'maulid_nabi', 'hari_santri',
        'other'
    )),
    CONSTRAINT chk_event_category CHECK (event_category IN ('academic', 'extracurricular', 'ceremonial', 'religious', 'social', 'administrative', 'other')),
    CONSTRAINT chk_target_audience CHECK (target_audience IN ('all_students', 'specific_classes', 'all_teachers', 'parents', 'all_school', 'external', 'santri')),
    CONSTRAINT chk_approval_status CHECK (approval_status IN ('draft', 'submitted', 'approved', 'rejected', 'revised')),
    CONSTRAINT chk_execution_status CHECK (execution_status IN ('planned', 'preparation', 'ongoing', 'completed', 'cancelled', 'postponed')),
    CONSTRAINT chk_event_dates CHECK (end_date >= start_date),
    CONSTRAINT chk_event_semester CHECK (semester IS NULL OR semester IN (1, 2))
);

CREATE INDEX idx_school_event_tenant ON school_events (tenant_id, company_id);
CREATE INDEX idx_school_event_year ON school_events (academic_year_id);
CREATE INDEX idx_school_event_type ON school_events (event_type);
CREATE INDEX idx_school_event_category ON school_events (event_category);
CREATE INDEX idx_school_event_date ON school_events (start_date, end_date);
CREATE INDEX idx_school_event_approval ON school_events (approval_status) WHERE approval_status IN ('submitted', 'revised');
CREATE INDEX idx_school_event_execution ON school_events (execution_status) WHERE execution_status NOT IN ('completed', 'cancelled');
CREATE INDEX idx_school_event_budget ON school_events (budget_item_id) WHERE budget_item_id IS NOT NULL;
CREATE INDEX idx_school_event_organizer ON school_events (organizer_id);
CREATE INDEX idx_school_event_rels ON school_events USING GIN (_rels);
CREATE INDEX idx_school_event_data ON school_events USING GIN (_data);

-- Peserta kegiatan & kehadiran
CREATE TABLE school_event_participants (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    event_id        UUID NOT NULL,

    -- Peserta (polymorphic: student, teacher, parent, external)
    participant_type VARCHAR(20) NOT NULL,
    participant_id  UUID,
    participant_name VARCHAR(255) NOT NULL,
    participant_role VARCHAR(30),

    -- Registrasi
    registration_status VARCHAR(20) NOT NULL DEFAULT 'registered',
    registered_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Kehadiran
    attendance_status VARCHAR(20),
    check_in_at     TIMESTAMPTZ,
    check_out_at    TIMESTAMPTZ,

    -- Feedback
    rating          INT,
    feedback        TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_event_participant UNIQUE (event_id, participant_type, participant_id),
    CONSTRAINT chk_participant_type CHECK (participant_type IN ('student', 'teacher', 'parent', 'staff', 'external')),
    CONSTRAINT chk_participant_role CHECK (participant_role IS NULL OR participant_role IN ('peserta', 'panitia', 'pembina', 'juri', 'narasumber', 'pendamping', 'tamu_undangan')),
    CONSTRAINT chk_reg_status CHECK (registration_status IN ('registered', 'confirmed', 'waitlisted', 'cancelled')),
    CONSTRAINT chk_attendance_status CHECK (attendance_status IS NULL OR attendance_status IN ('hadir', 'tidak_hadir', 'izin', 'terlambat')),
    CONSTRAINT chk_rating CHECK (rating IS NULL OR (rating >= 1 AND rating <= 5))
);

CREATE INDEX idx_event_participant_tenant ON school_event_participants (tenant_id, company_id);
CREATE INDEX idx_event_participant_event ON school_event_participants (event_id);
CREATE INDEX idx_event_participant_person ON school_event_participants (participant_type, participant_id);
CREATE INDEX idx_event_participant_attendance ON school_event_participants (attendance_status) WHERE attendance_status IS NOT NULL;
CREATE INDEX idx_event_participant_rels ON school_event_participants USING GIN (_rels);
CREATE INDEX idx_event_participant_data ON school_event_participants USING GIN (_data);

-- Dokumentasi kegiatan
CREATE TABLE school_event_documents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    event_id        UUID NOT NULL,

    -- Identitas dokumen
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    document_type   VARCHAR(20) NOT NULL,

    -- File
    file_url        TEXT NOT NULL,
    file_type       VARCHAR(20) NOT NULL,
    file_size_bytes BIGINT,
    thumbnail_url   TEXT,

    -- Metadata
    uploaded_by     UUID NOT NULL,
    is_public       BOOLEAN NOT NULL DEFAULT false,
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

    CONSTRAINT chk_document_type CHECK (document_type IN ('photo', 'video', 'report', 'proposal', 'budget_realization', 'attendance_list', 'certificate', 'other')),
    CONSTRAINT chk_file_type CHECK (file_type IN ('jpg', 'png', 'mp4', 'pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'other'))
);

CREATE INDEX idx_event_document_tenant ON school_event_documents (tenant_id, company_id);
CREATE INDEX idx_event_document_event ON school_event_documents (event_id);
CREATE INDEX idx_event_document_type ON school_event_documents (document_type);
CREATE INDEX idx_event_document_rels ON school_event_documents USING GIN (_rels);
CREATE INDEX idx_event_document_data ON school_event_documents USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `event_type` | VARCHAR(30), 19 types | Mencakup kegiatan umum sekolah + khusus pesantren (haflah, khataman, wisuda tahfidz) + hari besar Islam |
| `event_category` | VARCHAR(20) | Pengelompokan tingkat tinggi: academic, extracurricular, ceremonial, religious, social, administrative |
| `target_audience` | VARCHAR(20) | Siapa pesertanya — all_students, specific_classes, all_teachers, parents, santri |
| `target_classes` | JSONB array | Jika specific_classes — array class_id yang ditarget. Contoh: `["018f...", "018f..."]` |
| `committee` | JSONB array | Panitia kegiatan: `[{"user_id": "...", "name": "...", "role": "ketua"}, ...]` |
| `budget_item_id` | UUID, nullable | Link ke RKAS (S050) — kegiatan yang punya anggaran resmi |
| `estimated_budget` / `actual_budget` | BIGINT | Perbandingan rencana vs realisasi — penting untuk laporan keuangan |
| `calendar_event_id` | UUID, nullable | Link ke kalender akademik (S023) — otomatis masuk kalender sekolah |
| `participant_type` | VARCHAR(20), polymorphic | Peserta bisa student, teacher, parent, staff, atau external |
| `participant_role` | VARCHAR(30) | Peran di kegiatan: peserta, panitia, pembina, juri, narasumber, pendamping |
| `is_mandatory` | BOOLEAN | Upacara bendera mandatory, study tour optional — mempengaruhi attendance tracking |
| `outcomes` / `lessons_learned` | TEXT | Dokumentasi hasil kegiatan — penting untuk akreditasi dan evaluasi |

### Event Lifecycle Flow

```
┌─────────────────────────────────────────────────────────────┐
│ Event Lifecycle                                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ 1. DRAFT                                                    │
│    PIC buat proposal kegiatan di system                      │
│    Set: nama, tipe, tanggal, lokasi, target peserta          │
│    approval_status = 'draft'                                 │
│                                                             │
│ 2. SUBMIT                                                    │
│    PIC submit proposal → kepsek                              │
│    approval_status = 'submitted'                             │
│    (Integrates with S047 approval workflow)                  │
│                                                             │
│ 3. APPROVE / REJECT                                          │
│    Kepsek review → approve (budget allocated) atau reject    │
│    Jika approved: budget_item_id linked to S050 RKAS         │
│    calendar_event_id created in S023                         │
│                                                             │
│ 4. PREPARATION                                               │
│    execution_status = 'preparation'                          │
│    Panitia assign, peserta register                          │
│                                                             │
│ 5. EXECUTION                                                 │
│    execution_status = 'ongoing'                              │
│    Attendance tracking (check-in/check-out)                  │
│    Dokumentasi (foto, video)                                 │
│                                                             │
│ 6. COMPLETION                                                │
│    execution_status = 'completed'                            │
│    Upload laporan, budget realization                         │
│    Isi outcomes & lessons_learned                            │
│    actual_budget recorded                                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Vernon Relationships

**school_events:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |
| `organizer` (user) | belongs_to | **Ya** | PIC/penyelenggara kegiatan |
| `budget_item` (S050) | belongs_to | Tidak | Opsional — hanya perlu saat review budget |
| `approved_by` (user) | belongs_to | Tidak | Hanya perlu saat review detail approval |

**school_event_participants:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `event` | belongs_to | **Ya** | Selalu perlu konteks kegiatan |

**school_event_documents:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `event` | belongs_to | **Ya** | Selalu perlu konteks kegiatan |
| `uploaded_by` (user) | belongs_to | Tidak | Metadata — jarang ditampilkan |

### _rels / _data Structure

```json
// school_events
{
  "_rels": {
    "academic_year_id": "018f...",
    "organizer_id": "018f...",
    "budget_item_id": "018f..."
  },
  "_data": {
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "organizer": { "id": "018f...", "full_name": "Pak Budi", "role": "wakil_kepsek" }
  }
}

// school_event_participants
{
  "_rels": {
    "event_id": "018f...",
    "participant_id": "018f..."
  },
  "_data": {
    "event": { "id": "018f...", "name": "Study Tour Museum Nasional", "event_type": "study_tour", "start_date": "2026-05-15" }
  }
}

// school_event_documents
{
  "_rels": {
    "event_id": "018f...",
    "uploaded_by": "018f..."
  },
  "_data": {
    "event": { "id": "018f...", "name": "Pentas Seni Akhir Tahun", "event_type": "pentas_seni" }
  }
}
```

### API Endpoints

```
# School Events
GET    /api/v1/school-events                             — List kegiatan (filter by type, date, status)
POST   /api/v1/school-events                             — Buat proposal kegiatan
GET    /api/v1/school-events/{id}                        — Detail kegiatan
PUT    /api/v1/school-events/{id}                        — Update kegiatan
POST   /api/v1/school-events/{id}/submit                 — Submit untuk approval
PUT    /api/v1/school-events/{id}/approve                — Approve (kepsek)
PUT    /api/v1/school-events/{id}/reject                 — Reject (kepsek)
PUT    /api/v1/school-events/{id}/start                  — Mulai pelaksanaan
PUT    /api/v1/school-events/{id}/complete                — Selesaikan kegiatan

# Calendar
GET    /api/v1/school-events/calendar                    — Kegiatan dalam format kalender
GET    /api/v1/school-events/upcoming                    — Kegiatan mendatang

# Participants
GET    /api/v1/school-events/{id}/participants           — List peserta
POST   /api/v1/school-events/{id}/participants           — Daftarkan peserta (bulk)
PUT    /api/v1/school-event-participants/{id}/check-in    — Check-in kehadiran
PUT    /api/v1/school-event-participants/{id}/check-out   — Check-out
POST   /api/v1/school-event-participants/{id}/feedback    — Submit feedback/rating

# Documents
GET    /api/v1/school-events/{id}/documents              — List dokumen kegiatan
POST   /api/v1/school-events/{id}/documents              — Upload dokumen
DELETE /api/v1/school-event-documents/{id}               — Hapus dokumen

# Reports
GET    /api/v1/school-events/summary                     — Rekap kegiatan per semester
GET    /api/v1/school-events/{id}/attendance-report       — Laporan kehadiran kegiatan
GET    /api/v1/school-events/{id}/budget-report           — Laporan anggaran (estimasi vs realisasi)
```

## Consequences

### Positive

- **Centralized planning**: Semua kegiatan sekolah terkelola di satu tempat — tidak lagi tersebar di WhatsApp dan spreadsheet.
- **Approval workflow**: Proposal → approval memastikan governance yang baik dan anggaran terkontrol.
- **Budget accountability**: Link ke RKAS (S050) memastikan setiap kegiatan punya alokasi anggaran yang jelas. Perbandingan estimasi vs realisasi.
- **Calendar integration**: Kegiatan otomatis masuk kalender akademik (S023) — menghindari bentrok jadwal.
- **Documentation for accreditation**: Foto, laporan, outcomes tersimpan rapi — siap untuk visitasi akreditasi BAN-S/M.
- **Pesantren inclusive**: Event types mencakup haflah, khataman, wisuda tahfidz, pesantren kilat — relevan untuk pesantren.
- **Attendance tracking**: Kehadiran di kegiatan tercatat — bisa jadi input evaluasi siswa dan guru.

### Negative / Trade-offs

- **Overhead for simple events**: Upacara bendera yang rutin setiap Senin mungkin tidak perlu full workflow approval setiap minggu — perlu mekanisme recurring events.
- **Participant management**: Kegiatan besar (wisuda, upacara) bisa punya ratusan peserta — bulk registration perlu dioptimasi.
- **File storage cost**: Dokumentasi foto/video bisa consuming significant storage — perlu file size limits dan cleanup policy.
- **Cross-domain complexity**: Integrasi dengan S023 (calendar), S047 (approval), S050 (RKAS) menciptakan coupling antar domain.

## Alternatives Considered

### 1. Kegiatan hanya di kalender akademik (S023)
- Ditolak: kalender hanya menyimpan tanggal — tidak bisa track peserta, anggaran, approval, dan dokumentasi.

### 2. Menggunakan Google Calendar / external tool
- Ditolak: tidak terintegrasi dengan data internal sekolah (siswa, guru, anggaran). External tool juga tidak mendukung approval workflow.

### 3. Event sebagai JSONB di academic_years
- Ditolak: puluhan kegiatan per semester dengan participants dan documents — memerlukan tabel proper untuk query efisien.

### 4. Pisah tabel per event type
- Ditolak: event types terlalu banyak (19+) — satu tabel polymorphic dengan `event_type` column jauh lebih maintainable.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `EventDescriptor.TableName()` | — | `"school_events"` |
| U02 | `ParticipantDescriptor.TableName()` | — | `"school_event_participants"` |
| U03 | `DocumentDescriptor.TableName()` | — | `"school_event_documents"` |
| U04 | Validate rejects invalid `event_type` | `"concert"` | Error: invalid event_type |
| U05 | Validate rejects invalid `event_category` | `"entertainment"` | Error: invalid event_category |
| U06 | Validate rejects invalid `target_audience` | `"public"` | Error: invalid target_audience |
| U07 | Validate rejects end_date < start_date | start=2026-05-15, end=2026-05-14 | Error: end_date must >= start_date |
| U08 | Validate rejects invalid `participant_type` | `"alumni"` | Error: invalid participant_type |
| U09 | Validate rejects invalid `document_type` | `"audio"` | Error: invalid document_type |
| U10 | Validate rejects rating out of range | `rating = 6` | Error: rating must be 1-5 |
| U11 | Validate accepts valid event | All fields valid | No error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create event proposal | POST with type=study_tour | 201, approval_status = 'draft' |
| I02 | Submit for approval | POST /submit | 200, approval_status → 'submitted' |
| I03 | Approve event | PUT /approve by kepsek | 200, approval_status → 'approved' |
| I04 | Reject event | PUT /reject with note | 200, approval_status → 'rejected' |
| I05 | Start event execution | PUT /start | 200, execution_status → 'ongoing' |
| I06 | Cannot start unapproved event | PUT /start on draft event | 422, must be approved first |
| I07 | Register participants bulk | POST participants for class VII-A | 201, all students registered |
| I08 | Unique participant per event | Register same student twice | 409/422, unique constraint |
| I09 | Check-in participant | PUT /check-in | 200, check_in_at set, event attendance_count++ |
| I10 | Upload document | POST document with photo | 201, linked to event |
| I11 | Complete event | PUT /complete with outcomes | 200, execution_status → 'completed' |
| I12 | Budget report | GET /budget-report | estimated_budget vs actual_budget comparison |
| I13 | Attendance report | GET /attendance-report | Participant counts by attendance_status |
| I14 | Calendar view | GET /calendar?month=5&year=2026 | Events in May 2026 |
| I15 | Event type CHECK enforced | INSERT with `event_type = 'concert'` | DB error |
| I16 | Date range CHECK enforced | INSERT with end_date < start_date | DB error |
| I17 | Pesantren event types | Create haflah, khataman, wisuda_tahfidz | All accepted |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I18 | AcademicYearUpdated syncs to events | Update academic year name | `_data.academic_year.name` updated |
| I19 | EventUpdated syncs to participants | Update event name | `_data.event.name` updated in participants |
| I20 | EventUpdated syncs to documents | Update event name | `_data.event.name` updated in documents |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I21 | Cannot access other tenant's events | GET with wrong tenant | 404 |
| I22 | Cannot register participant cross-tenant | POST participant from other tenant | Error |
| I23 | Documents tenant-isolated | Access other tenant's documents | 404 |
