# ADR-S053: E-Learning Integration (Integrasi Pembelajaran Daring)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Pandemi COVID-19 telah mengakselerasi adopsi pembelajaran daring di sekolah Indonesia. Meskipun tatap muka sudah kembali normal, banyak sekolah tetap menerapkan **blended learning** — kombinasi tatap muka dan daring. SekolahPro bukan LMS (Learning Management System), tetapi perlu **integrasi** dengan LMS yang sudah digunakan sekolah.

Kebutuhan integrasi e-learning:

1. **LMS connector**: Sekolah sudah pakai Google Classroom, Moodle, atau Microsoft Teams — kita integrasikan, bukan menggantikan.
2. **Assignment sync**: Tugas dan submission dari LMS di-sync ke SekolahPro untuk pencatatan nilai terpusat (S011).
3. **Online exam links**: Guru membuat link ujian online (Google Form, Quizizz, dll) — terhubung dengan jadwal (S021) dan assessment (S022).
4. **Blended attendance**: Kehadiran kelas daring dicatat terpisah namun terintegrasi dengan S008 (daily attendance).
5. **Material sharing**: Link materi/resource dari LMS atau cloud storage — bukan file hosting di SekolahPro.
6. **Sync monitoring**: Admin perlu melihat status sinkronisasi dan troubleshoot error.

### Mengapa Vernon Pattern?

- Read-heavy: guru dan admin sering melihat status sync, daftar assignment, dan material links.
- Relasi ke subject, teacher, academic_year, class.
- Write periodik: sync dari LMS terjadi via webhook atau scheduled polling.
- Business logic moderate: mapping fields, conflict resolution, retry logic.
- Eventual consistency acceptable — data dari LMS bisa delay beberapa menit.

## Decision

Menggunakan **Vernon Pattern** untuk 4 tabel: `elearning_integrations` (konfigurasi koneksi LMS), `elearning_assignments` (tugas yang di-sync), `elearning_materials` (materi/resource links), dan `elearning_sync_logs` (log sinkronisasi).

### Table Schema

```sql
-- Konfigurasi koneksi ke LMS
CREATE TABLE elearning_integrations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    provider        VARCHAR(30) NOT NULL,
    description     TEXT,

    -- Credentials (encrypted at app layer)
    api_base_url    TEXT NOT NULL,
    api_key_enc     TEXT,
    oauth_token_enc TEXT,
    webhook_secret  TEXT,

    -- Konfigurasi sync
    sync_mode       VARCHAR(20) NOT NULL DEFAULT 'webhook',
    poll_interval_min INT,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    last_sync_at    TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_elearning_integration UNIQUE (tenant_id, company_id, provider),
    CONSTRAINT chk_provider CHECK (provider IN ('google_classroom', 'moodle', 'ms_teams', 'canvas', 'other')),
    CONSTRAINT chk_sync_mode CHECK (sync_mode IN ('webhook', 'polling', 'manual'))
);

CREATE INDEX idx_elearning_integration_tenant ON elearning_integrations (tenant_id, company_id);
CREATE INDEX idx_elearning_integration_rels ON elearning_integrations USING GIN (_rels);
CREATE INDEX idx_elearning_integration_data ON elearning_integrations USING GIN (_data);

-- Tugas / Assignment dari LMS
CREATE TABLE elearning_assignments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    integration_id  UUID NOT NULL,
    subject_id      UUID,
    teacher_id      UUID,
    class_id        UUID,
    academic_year_id UUID,

    -- Identitas dari LMS
    external_id     VARCHAR(255) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    assignment_type VARCHAR(20) NOT NULL,

    -- Link & deadline
    external_url    TEXT NOT NULL,
    due_date        TIMESTAMPTZ,
    max_score       NUMERIC(8,2),

    -- Submission stats (synced from LMS)
    total_students  INT NOT NULL DEFAULT 0,
    submitted_count INT NOT NULL DEFAULT 0,
    graded_count    INT NOT NULL DEFAULT 0,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    last_synced_at  TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_elearning_assignment_ext UNIQUE (tenant_id, company_id, integration_id, external_id),
    CONSTRAINT chk_assignment_type CHECK (assignment_type IN ('tugas', 'kuis', 'ujian_online', 'diskusi', 'project')),
    CONSTRAINT chk_assignment_status CHECK (status IN ('active', 'closed', 'archived', 'error'))
);

CREATE INDEX idx_elearning_assignment_tenant ON elearning_assignments (tenant_id, company_id);
CREATE INDEX idx_elearning_assignment_integration ON elearning_assignments (integration_id);
CREATE INDEX idx_elearning_assignment_subject ON elearning_assignments (subject_id);
CREATE INDEX idx_elearning_assignment_teacher ON elearning_assignments (teacher_id);
CREATE INDEX idx_elearning_assignment_class ON elearning_assignments (class_id);
CREATE INDEX idx_elearning_assignment_status ON elearning_assignments (status) WHERE status = 'active';
CREATE INDEX idx_elearning_assignment_rels ON elearning_assignments USING GIN (_rels);
CREATE INDEX idx_elearning_assignment_data ON elearning_assignments USING GIN (_data);

-- Materi / Resource links dari LMS
CREATE TABLE elearning_materials (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    integration_id  UUID NOT NULL,
    subject_id      UUID,
    teacher_id      UUID,
    class_id        UUID,
    academic_year_id UUID,

    -- Identitas dari LMS
    external_id     VARCHAR(255),
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    material_type   VARCHAR(20) NOT NULL,

    -- Link
    external_url    TEXT NOT NULL,
    file_type       VARCHAR(20),
    file_size_bytes BIGINT,

    -- Status
    is_published    BOOLEAN NOT NULL DEFAULT true,
    last_synced_at  TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_material_type CHECK (material_type IN ('document', 'video', 'audio', 'presentation', 'link', 'other')),
    CONSTRAINT chk_file_type CHECK (file_type IS NULL OR file_type IN ('pdf', 'doc', 'docx', 'ppt', 'pptx', 'xls', 'xlsx', 'mp4', 'mp3', 'jpg', 'png', 'other'))
);

CREATE INDEX idx_elearning_material_tenant ON elearning_materials (tenant_id, company_id);
CREATE INDEX idx_elearning_material_integration ON elearning_materials (integration_id);
CREATE INDEX idx_elearning_material_subject ON elearning_materials (subject_id);
CREATE INDEX idx_elearning_material_teacher ON elearning_materials (teacher_id);
CREATE INDEX idx_elearning_material_rels ON elearning_materials USING GIN (_rels);
CREATE INDEX idx_elearning_material_data ON elearning_materials USING GIN (_data);

-- Log sinkronisasi
CREATE TABLE elearning_sync_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    integration_id  UUID NOT NULL,

    -- Sync details
    sync_type       VARCHAR(20) NOT NULL,
    sync_direction  VARCHAR(10) NOT NULL DEFAULT 'inbound',
    entity_type     VARCHAR(30) NOT NULL,
    entity_count    INT NOT NULL DEFAULT 0,

    -- Result
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message   TEXT,
    error_details   JSONB,

    -- Timing
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ,
    duration_ms     INT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_sync_type CHECK (sync_type IN ('webhook', 'polling', 'manual', 'full_sync')),
    CONSTRAINT chk_sync_direction CHECK (sync_direction IN ('inbound', 'outbound')),
    CONSTRAINT chk_sync_entity CHECK (entity_type IN ('assignment', 'submission', 'material', 'grade', 'attendance', 'roster')),
    CONSTRAINT chk_sync_log_status CHECK (status IN ('pending', 'running', 'success', 'partial', 'error', 'retry'))
);

CREATE INDEX idx_elearning_sync_tenant ON elearning_sync_logs (tenant_id, company_id);
CREATE INDEX idx_elearning_sync_integration ON elearning_sync_logs (integration_id);
CREATE INDEX idx_elearning_sync_status ON elearning_sync_logs (status) WHERE status IN ('pending', 'running', 'error', 'retry');
CREATE INDEX idx_elearning_sync_started ON elearning_sync_logs (started_at DESC);
CREATE INDEX idx_elearning_sync_rels ON elearning_sync_logs USING GIN (_rels);
CREATE INDEX idx_elearning_sync_data ON elearning_sync_logs USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `provider` | VARCHAR(30), CHECK constraint | Support 4 provider utama + 'other' untuk extensibility |
| `api_key_enc` / `oauth_token_enc` | TEXT, encrypted at app layer | Credentials TIDAK disimpan plain text — encryption di application layer |
| `sync_mode` | VARCHAR(20), 3 mode | Webhook (real-time), polling (scheduled), manual (admin trigger) |
| `external_id` | VARCHAR(255) per assignment | ID dari LMS — digunakan untuk deduplication dan update |
| `assignment_type` | VARCHAR(20) | Mapping ke tipe tugas Indonesia: tugas, kuis, ujian online, diskusi, project |
| `submitted_count` / `graded_count` | INT, denormalized | Stats dari LMS di-cache — menghindari query ke external API |
| `material_type` | VARCHAR(20) | Kategori resource: document, video, audio, presentation, link |
| `sync_direction` | VARCHAR(10) | Inbound (LMS → SekolahPro) atau outbound (SekolahPro → LMS) |
| `duration_ms` | INT | Performance monitoring — seberapa lama sync berlangsung |

### Vernon Relationships

**elearning_assignments:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `integration` | belongs_to | **Ya** | Perlu tahu dari LMS mana assignment ini |
| `subject` | belongs_to | **Ya** | Mapping ke mata pelajaran SekolahPro |
| `teacher` | belongs_to | **Ya** | Guru yang membuat assignment |
| `class` | belongs_to | Tidak | Opsional — assignment bisa lintas kelas |

**elearning_materials:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `integration` | belongs_to | **Ya** | Sumber materi |
| `subject` | belongs_to | **Ya** | Mapping ke mata pelajaran |
| `teacher` | belongs_to | **Ya** | Guru pemilik materi |

**elearning_sync_logs:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `integration` | belongs_to | **Ya** | Selalu perlu tahu koneksi mana yang di-sync |

### _rels / _data Structure

```json
// elearning_assignments
{
  "_rels": {
    "integration_id": "018f...",
    "subject_id": "018f...",
    "teacher_id": "018f...",
    "class_id": "018f..."
  },
  "_data": {
    "integration": { "id": "018f...", "name": "Google Classroom SMP", "provider": "google_classroom" },
    "subject": { "id": "018f...", "name": "Matematika", "code": "MTK" },
    "teacher": { "id": "018f...", "full_name": "Pak Budi", "nip": "198501012010011001" },
    "class": { "id": "018f...", "name": "VII-A" }
  }
}

// elearning_materials
{
  "_rels": {
    "integration_id": "018f...",
    "subject_id": "018f...",
    "teacher_id": "018f..."
  },
  "_data": {
    "integration": { "id": "018f...", "name": "Moodle SMA", "provider": "moodle" },
    "subject": { "id": "018f...", "name": "Bahasa Inggris", "code": "BIG" },
    "teacher": { "id": "018f...", "full_name": "Bu Ani", "nip": "199001012015012001" }
  }
}

// elearning_sync_logs
{
  "_rels": {
    "integration_id": "018f..."
  },
  "_data": {
    "integration": { "id": "018f...", "name": "Google Classroom SMP", "provider": "google_classroom" }
  }
}
```

### API Endpoints

```
# Integration Config
GET    /api/v1/elearning-integrations                    — List konfigurasi LMS
POST   /api/v1/elearning-integrations                    — Tambah koneksi LMS baru
PUT    /api/v1/elearning-integrations/{id}               — Update konfigurasi
POST   /api/v1/elearning-integrations/{id}/test          — Test koneksi ke LMS
POST   /api/v1/elearning-integrations/{id}/sync          — Trigger manual sync

# Assignments (synced from LMS)
GET    /api/v1/elearning-assignments                     — List semua assignment
GET    /api/v1/elearning-assignments/{id}                — Detail assignment + submission stats
GET    /api/v1/elearning-assignments/by-subject/{id}     — Assignment per mata pelajaran
GET    /api/v1/elearning-assignments/by-teacher/{id}     — Assignment per guru

# Materials
GET    /api/v1/elearning-materials                       — List materi
POST   /api/v1/elearning-materials                       — Tambah link materi (manual, tanpa LMS)
GET    /api/v1/elearning-materials/by-subject/{id}       — Materi per mata pelajaran

# Sync Logs
GET    /api/v1/elearning-sync-logs                       — List log sinkronisasi
GET    /api/v1/elearning-sync-logs/{integration_id}      — Log per integrasi
POST   /api/v1/elearning-sync-logs/{id}/retry            — Retry failed sync

# Webhook receiver
POST   /api/v1/webhooks/google-classroom                 — Webhook dari Google Classroom
POST   /api/v1/webhooks/moodle                           — Webhook dari Moodle
```

## Consequences

### Positive

- **Integration layer, bukan LMS**: SekolahPro tidak reinvent the wheel — leverage LMS yang sudah dipakai sekolah.
- **Centralized view**: Guru dan admin melihat semua assignment/materi dari berbagai LMS di satu tempat.
- **Grade pipeline**: Assignment sync menjadi input untuk S011 (subject grade) — nilai dari LMS bisa masuk ke rapor.
- **Flexible sync**: 3 mode (webhook, polling, manual) mengakomodasi berbagai kemampuan LMS.
- **Audit trail**: Sync logs memudahkan troubleshooting ketika sinkronisasi gagal.

### Negative / Trade-offs

- **External dependency**: Jika API LMS berubah atau down, sync terganggu — perlu circuit breaker dan retry logic.
- **Mapping complexity**: Setiap LMS punya struktur data berbeda — perlu adapter pattern per provider.
- **Credential management**: Menyimpan API key/OAuth token memerlukan encryption yang proper.
- **Eventual consistency**: Data dari LMS bisa delay — guru harus paham bahwa data tidak real-time.
- **Limited scope**: Hanya sync metadata dan stats — file/content tetap di LMS, kita hanya simpan link.

## Alternatives Considered

### 1. Build full LMS di dalam SekolahPro
- Ditolak: sekolah sudah invested di Google Classroom/Moodle — menggantikan tidak realistis dan scope terlalu besar.

### 2. Hanya simpan link tanpa sync
- Ditolak: tanpa sync, guru harus manual input data submission dan nilai ke SekolahPro — tidak efisien.

### 3. Direct database connection ke Moodle
- Ditolak: terlalu fragile, Moodle schema bisa berubah antar versi. API/webhook lebih stable.

### 4. LTI (Learning Tools Interoperability) standard
- Dipertimbangkan untuk future: LTI 1.3 adalah standar industri, tapi adopsi di sekolah Indonesia masih rendah. Bisa ditambahkan sebagai provider tambahan.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `IntegrationDescriptor.TableName()` | — | `"elearning_integrations"` |
| U02 | `AssignmentDescriptor.TableName()` | — | `"elearning_assignments"` |
| U03 | `MaterialDescriptor.TableName()` | — | `"elearning_materials"` |
| U04 | `SyncLogDescriptor.TableName()` | — | `"elearning_sync_logs"` |
| U05 | Validate rejects invalid `provider` | `"zoom"` | Error: invalid provider |
| U06 | Validate rejects invalid `sync_mode` | `"realtime"` | Error: invalid sync_mode |
| U07 | Validate rejects invalid `assignment_type` | `"homework"` | Error: invalid assignment_type |
| U08 | Validate rejects invalid `material_type` | `"image"` | Error: invalid material_type |
| U09 | Validate rejects invalid sync log `status` | `"cancelled"` | Error: invalid status |
| U10 | Validate accepts valid integration | All fields valid | No error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create integration config | POST with provider + credentials | 201, created |
| I02 | Unique provider per company | Create 2 Google Classroom configs | 409/422, unique constraint |
| I03 | Test connection | POST /test for valid config | 200, connection OK |
| I04 | Trigger manual sync | POST /sync | 200, sync log created |
| I05 | Webhook creates assignment | POST /webhooks/google-classroom | 201, assignment synced |
| I06 | Duplicate external_id ignored | Same webhook twice | Upsert, no duplicate |
| I07 | Assignment links to subject | Sync with mapped subject | `_data.subject` populated |
| I08 | Create material link | POST manual material | 201, created |
| I09 | Sync log tracks error | Sync fails with API error | Log status = 'error', error_message set |
| I10 | Retry failed sync | POST /retry | New sync attempt, new log entry |
| I11 | Provider CHECK enforced | INSERT with `provider = 'zoom'` | DB error |
| I12 | Sync status CHECK enforced | INSERT with `status = 'cancelled'` | DB error |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | SubjectUpdated syncs to assignments | Update subject name | `_data.subject.name` updated in assignments |
| I14 | TeacherUpdated syncs to assignments | Update teacher name | `_data.teacher.full_name` updated |
| I15 | IntegrationUpdated syncs to sync_logs | Update integration name | `_data.integration.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Cannot access other tenant's integrations | GET with wrong tenant | 404 |
| I17 | Webhook validates tenant | Webhook without valid tenant context | 401/403 |
| I18 | Sync logs isolated | List logs for other tenant | Empty result |
