# ADR-S017: BK / Student Counseling Record

**Status**: Approved (Backlog — build on demand)
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Bimbingan Konseling (BK) adalah layanan pendampingan siswa yang dilakukan oleh guru BK. Catatan BK diperlukan untuk:

1. **Pembinaan individual**: Tracking masalah dan perkembangan siswa.
2. **Rujukan**: Eskalasi ke psikolog/pihak eksternal jika diperlukan.
3. **Dokumentasi**: Bukti layanan BK untuk akreditasi sekolah.
4. **Rapor**: Catatan BK bisa menjadi lampiran rapor (opsional).
5. **Handover**: Guru BK berganti — catatan harus terdokumentasi.

Karakteristik:
- **Privasi sangat tinggi**: Data BK bersifat rahasia — hanya guru BK dan kepsek yang boleh akses.
- **Event-based**: Setiap sesi konseling = satu record.
- **Frekuensi rendah**: Tidak semua siswa punya catatan BK — hanya yang dirujuk/berkonsultasi.
- **Unstructured**: Catatan sesi cenderung naratif, tidak terstruktur.

### Mengapa Vernon Pattern?

- has_many dari student (banyak sesi per siswa).
- Read-heavy saat review progress siswa.
- Business logic sederhana: CRUD + access control.
- Eventual consistency acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `counseling_cases` (kasus/masalah) dan `counseling_sessions` (sesi konseling per kasus).

### Table Schema

```sql
-- Kasus / masalah siswa
CREATE TABLE counseling_cases (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Detail kasus
    case_no         VARCHAR(20) NOT NULL,
    category        VARCHAR(20) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    severity        VARCHAR(10) NOT NULL DEFAULT 'low',

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'open',
    opened_date     DATE NOT NULL,
    closed_date     DATE,
    resolution      TEXT,

    -- Penanganan
    counselor_id    UUID NOT NULL,
    referred_to     TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_case_no UNIQUE (tenant_id, company_id, case_no),
    CONSTRAINT chk_case_category CHECK (category IN (
        'academic', 'social', 'personal', 'career', 'behavioral', 'family', 'other'
    )),
    CONSTRAINT chk_case_severity CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT chk_case_status CHECK (status IN ('open', 'in_progress', 'referred', 'resolved', 'closed'))
);

-- Indexes
CREATE INDEX idx_case_tenant_company ON counseling_cases (tenant_id, company_id);
CREATE INDEX idx_case_student ON counseling_cases (student_id);
CREATE INDEX idx_case_counselor ON counseling_cases (counselor_id);
CREATE INDEX idx_case_status ON counseling_cases (status) WHERE status NOT IN ('resolved', 'closed');
CREATE INDEX idx_case_rels ON counseling_cases USING GIN (_rels);
CREATE INDEX idx_case_data ON counseling_cases USING GIN (_data);

-- Sesi konseling per kasus
CREATE TABLE counseling_sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    case_id         UUID NOT NULL,
    student_id      UUID NOT NULL,

    -- Detail sesi
    session_date    DATE NOT NULL,
    session_type    VARCHAR(20) NOT NULL,
    duration_minutes INT,
    notes           TEXT NOT NULL,
    recommendation  TEXT,

    -- Peserta
    counselor_id    UUID NOT NULL,
    parent_present  BOOLEAN NOT NULL DEFAULT false,

    -- Follow-up
    follow_up_date  DATE,
    follow_up_note  TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_session_type CHECK (session_type IN (
        'individual', 'group', 'home_visit', 'parent_conference', 'referral'
    ))
);

-- Indexes
CREATE INDEX idx_session_tenant_company ON counseling_sessions (tenant_id, company_id);
CREATE INDEX idx_session_case ON counseling_sessions (case_id);
CREATE INDEX idx_session_student ON counseling_sessions (student_id);
CREATE INDEX idx_session_date ON counseling_sessions (session_date);
CREATE INDEX idx_session_rels ON counseling_sessions USING GIN (_rels);
CREATE INDEX idx_session_data ON counseling_sessions USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `category` | 7 kategori | Standar BK Indonesia: pribadi, sosial, belajar, karir + behavioral, family |
| `severity` | 4 level | low → critical untuk prioritas penanganan |
| `case_no` | VARCHAR(20), unique | Nomor kasus untuk referensi dan tracking |
| `counselor_id` | UUID, NOT NULL | Guru BK yang menangani — audit trail |
| `referred_to` | TEXT, nullable | Referral ke pihak eksternal (psikolog, dll) |
| `session_type` | 5 tipe | Individual, grup, kunjungan rumah, konferensi ortu, referral |
| `parent_present` | BOOLEAN | Apakah orang tua hadir dalam sesi |
| `notes` | TEXT, NOT NULL | Catatan sesi — inti dari BK record |

### Access Control (Critical)

```
Guru BK        → CRUD own cases + sessions
Kepala Sekolah → Read all cases (tidak bisa edit)
Wali Kelas     → Read cases untuk siswa di kelasnya (summary only, bukan detail sesi)
Admin          → Tidak ada akses ke BK data
Orang Tua      → Tidak ada akses ke BK data (diberitahu via sesi langsung)
```

Ini memerlukan **domain-level access control** yang lebih ketat dari RBAC standar.

### Vernon Relationships

**counseling_cases:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Pemilik kasus |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

**counseling_sessions:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `case` | belongs_to | **Ya** | Kasus induk |
| `student` | belongs_to | **Ya** | Pemilik sesi |

### API Endpoints

```
GET    /api/v1/counseling-cases                     — List kasus BK (guru BK only)
POST   /api/v1/counseling-cases                     — Buat kasus baru
PUT    /api/v1/counseling-cases/{id}                — Update kasus
PUT    /api/v1/counseling-cases/{id}/close           — Tutup kasus

GET    /api/v1/counseling-cases/{id}/sessions        — Sesi per kasus
POST   /api/v1/counseling-sessions                   — Catat sesi
PUT    /api/v1/counseling-sessions/{id}             — Update sesi

GET    /api/v1/students/{id}/counseling-summary      — Summary untuk wali kelas (limited view)
```

## Consequences

### Positive

- **Case-based**: Setiap masalah = satu case dengan banyak sesi — tracking progress jelas.
- **Privasi by design**: Access control ketat di API level, bukan hanya UI.
- **Rujukan terdokumentasi**: Referral ke pihak eksternal tercatat.
- **Follow-up tracking**: Setiap sesi bisa punya jadwal follow-up.

### Negative / Trade-offs

- **Access control complexity**: Domain-level ACL lebih kompleks dari RBAC standar — butuh implementasi khusus.
- **Sensitive data storage**: Data BK sangat sensitif — perlu encryption at rest dan audit log akses.
- **Low adoption**: Banyak sekolah belum siap untuk digitalisasi BK.
- **Naratif unstructured**: Notes sebagai TEXT tidak bisa di-aggregate untuk laporan otomatis.

## Alternatives Considered

### 1. BK sebagai sub-type di S012 (Discipline)
- Ditolak: BK bersifat pembinaan dan rahasia — disiplin bersifat poin dan transparan. Dua domain dengan access control berbeda.

### 2. Single table tanpa case/session separation
- Ditolak: satu masalah bisa punya 5-10 sesi — tanpa case entity, tidak bisa track progress per masalah.

### 3. External counseling system integration
- Deferred: mayoritas sekolah tidak punya system BK digital — built-in lebih practical.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `CaseDescriptor.TableName()` | — | `"counseling_cases"` |
| U02 | `SessionDescriptor.TableName()` | — | `"counseling_sessions"` |
| U03 | Validate rejects invalid `category` | `"health"` | Error |
| U04 | Validate rejects invalid `severity` | `"urgent"` | Error |
| U05 | Validate rejects invalid `session_type` | `"phone_call"` | Error |
| U06 | Validate rejects empty `notes` | `""` | Error: notes required |
| U07 | Validate accepts valid case | All fields valid | No error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create counseling case | POST with valid data | 201 |
| I02 | Add session to case | POST session with case_id | 201 |
| I03 | Close case | PUT /close with resolution | 200, status → closed |
| I04 | Guru BK access own cases | GET /counseling-cases | 200, filtered by counselor_id |
| I05 | Admin cannot access BK data | GET /counseling-cases as admin | 403 |
| I06 | Wali kelas gets summary only | GET /students/{id}/counseling-summary | 200, limited fields |
| I07 | Category CHECK | INSERT with `category = 'health'` | DB error |
| I08 | Severity CHECK | INSERT with `severity = 'urgent'` | DB error |
| I09 | Tenant isolation | Access other tenant's cases | 404 |
