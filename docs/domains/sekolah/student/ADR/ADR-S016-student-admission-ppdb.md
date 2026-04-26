# ADR-S016: PPDB / Student Admission

**Status**: Approved (Roadmap Q1-Q2 2027)
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

PPDB (Penerimaan Peserta Didik Baru) adalah proses tahunan penerimaan siswa baru yang menjadi pintu masuk utama data siswa ke dalam sistem. Di Indonesia, PPDB terjadi sekitar **Juni-Agustus** setiap tahun.

PPDB diperlukan untuk:

1. **Akuisisi sekolah**: Sekolah yang mulai pakai PPDB online kita di Juli, captive untuk pakai sistem seterusnya.
2. **Formulir pendaftaran**: Calon siswa/orang tua mengisi data online — mengurangi input manual admin.
3. **Seleksi multi-jalur**: Zonasi, prestasi, afirmasi, pindahan — masing-masing punya kuota dan kriteria.
4. **Pengumuman & konfirmasi**: Hasil seleksi, daftar ulang, dan konversi applicant → student.
5. **Branding**: Halaman PPDB online = touchpoint pertama orang tua dengan sekolah.
6. **Data seeding**: Data PPDB yang dikonfirmasi otomatis menjadi data S001 (student), S003 (guardian), S007 (address, previous school).

Karakteristik:
- **Seasonal**: Hanya aktif 2-3 bulan per tahun, tapi sangat intens selama periode tersebut.
- **Public-facing**: Halaman pendaftaran bisa diakses tanpa login (calon siswa/orang tua).
- **Multi-stage**: Pendaftaran → Verifikasi → Seleksi → Pengumuman → Daftar Ulang.
- **Kuota-based**: Setiap jalur punya kuota yang harus dimanage.

### Mengapa Vernon Pattern?

- has_many dari academic_year (banyak applicant per periode).
- Read-heavy selama pengumuman: ribuan orang tua cek hasil bersamaan.
- Business logic moderate: seleksi, ranking, kuota management.
- Eventual consistency acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `admission_periods` (periode PPDB), `admission_tracks` (jalur seleksi), dan `applicants` (pendaftar).

### Table Schema

```sql
-- Periode PPDB per tahun ajaran
CREATE TABLE admission_periods (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    academic_year_id UUID NOT NULL,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,

    -- Timeline
    registration_start DATE NOT NULL,
    registration_end   DATE NOT NULL,
    announcement_date  DATE NOT NULL,
    enrollment_deadline DATE NOT NULL,

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

    CONSTRAINT uq_admission_period_year UNIQUE (tenant_id, company_id, academic_year_id),
    CONSTRAINT chk_admission_status CHECK (status IN ('draft', 'open', 'closed', 'announced', 'completed'))
);

-- Jalur seleksi per periode
CREATE TABLE admission_tracks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    admission_period_id UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    track_type      VARCHAR(20) NOT NULL,
    quota           INT NOT NULL,
    registered_count INT NOT NULL DEFAULT 0,
    accepted_count  INT NOT NULL DEFAULT 0,

    -- Persyaratan
    requirements    TEXT,
    selection_criteria TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_track_type CHECK (track_type IN ('zonasi', 'prestasi', 'afirmasi', 'pindahan', 'reguler')),
    CONSTRAINT chk_quota CHECK (quota > 0)
);

-- Indexes
CREATE INDEX idx_admission_track_period ON admission_tracks (admission_period_id);

-- Pendaftar / Calon Siswa
CREATE TABLE applicants (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    admission_period_id UUID NOT NULL,
    track_id        UUID NOT NULL,

    -- Nomor pendaftaran
    registration_no VARCHAR(20) NOT NULL,

    -- Biodata calon siswa (mirror S001 fields)
    full_name       VARCHAR(255) NOT NULL,
    nickname        VARCHAR(100),
    gender          VARCHAR(1) NOT NULL,
    birth_place     VARCHAR(100) NOT NULL,
    birth_date      DATE NOT NULL,
    religion        VARCHAR(20) NOT NULL,
    nisn            VARCHAR(10),
    photo_url       TEXT,

    -- Alamat (mirror S007 fields)
    address         TEXT,
    city            VARCHAR(100),
    province        VARCHAR(100),
    latitude        NUMERIC(10,7),
    longitude       NUMERIC(10,7),

    -- Sekolah asal (mirror S007 fields)
    previous_school_name VARCHAR(255),
    previous_school_npsn VARCHAR(8),

    -- Data orang tua utama (mirror S003 fields)
    guardian_name   VARCHAR(255) NOT NULL,
    guardian_phone  VARCHAR(20) NOT NULL,
    guardian_email  VARCHAR(255),
    guardian_occupation VARCHAR(100),
    guardian_income_range VARCHAR(20),

    -- Seleksi
    selection_score NUMERIC(8,2),
    selection_rank  INT,
    selection_status VARCHAR(20) NOT NULL DEFAULT 'registered',
    selection_note  TEXT,

    -- Daftar ulang
    enrolled_at     TIMESTAMPTZ,
    student_id      UUID,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_registration_no UNIQUE (tenant_id, company_id, registration_no),
    CONSTRAINT chk_applicant_gender CHECK (gender IN ('L', 'P')),
    CONSTRAINT chk_applicant_religion CHECK (religion IN ('islam', 'kristen', 'katolik', 'hindu', 'buddha', 'konghucu')),
    CONSTRAINT chk_selection_status CHECK (selection_status IN ('registered', 'verified', 'accepted', 'rejected', 'waitlisted', 'enrolled', 'withdrawn'))
);

-- Indexes
CREATE INDEX idx_applicant_tenant_company ON applicants (tenant_id, company_id);
CREATE INDEX idx_applicant_period ON applicants (admission_period_id);
CREATE INDEX idx_applicant_track ON applicants (track_id);
CREATE INDEX idx_applicant_status ON applicants (selection_status);
CREATE INDEX idx_applicant_rels ON applicants USING GIN (_rels);
CREATE INDEX idx_applicant_data ON applicants USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `registration_no` | VARCHAR(20), unique | Nomor pendaftaran auto-generate per periode |
| Biodata fields | Mirror S001/S003/S007 | Data diinput oleh calon siswa/orang tua — setelah enrolled, di-copy ke tabel student |
| `selection_score` | NUMERIC(8,2), nullable | Skor gabungan untuk ranking (jarak zonasi, nilai prestasi, dll) |
| `selection_rank` | INT, nullable | Ranking dalam jalur — dihitung saat seleksi |
| `student_id` | UUID, nullable | Link ke students table setelah daftar ulang — NULL saat belum enrolled |
| `registered_count` / `accepted_count` | INT di track | Counter denormalisasi untuk dashboard kuota real-time |

### Applicant → Student Conversion Flow

```
1. Applicant diterima (selection_status = 'accepted')
2. Orang tua melakukan daftar ulang (pembayaran + konfirmasi)
3. System converts applicant → student:
   a. CREATE student (S001) dari biodata applicant
   b. CREATE student_guardian (S003) dari guardian data
   c. CREATE student_address (S007) dari alamat
   d. CREATE student_previous_school (S007) dari sekolah asal
   e. CREATE student_class_placement (S014) dengan placement_type = 'initial'
   f. UPDATE applicant.student_id = new student ID
   g. UPDATE applicant.selection_status = 'enrolled'
4. Student sudah ada di system — siap untuk operasional
```

### API Endpoints

```
# Admission Periods
GET    /api/v1/admission-periods                      — List periode PPDB
POST   /api/v1/admission-periods                      — Buat periode
PUT    /api/v1/admission-periods/{id}                 — Update (open/close)

# Tracks
GET    /api/v1/admission-periods/{id}/tracks           — Jalur seleksi per periode
POST   /api/v1/admission-tracks                        — Buat jalur
PUT    /api/v1/admission-tracks/{id}                   — Update (kuota)

# Applicants (Public + Admin)
POST   /api/v1/public/apply                            — Pendaftaran (public, no auth)
GET    /api/v1/public/check/{registration_no}          — Cek status (public)
GET    /api/v1/applicants                              — List pendaftar (admin)
PUT    /api/v1/applicants/{id}/verify                  — Verifikasi dokumen
POST   /api/v1/applicants/select                       — Jalankan seleksi per track
POST   /api/v1/applicants/{id}/enroll                  — Daftar ulang → convert to student
```

## Consequences

### Positive

- **Growth engine**: PPDB online = touchpoint pertama orang tua, multiplicatif untuk brand awareness.
- **Data seeding**: Applicant yang enrolled otomatis menjadi student + guardian + address — zero re-entry.
- **Multi-jalur**: Mendukung 5 jalur standar PPDB Indonesia.
- **Public API**: Orang tua bisa daftar dan cek status tanpa akun.
- **Kuota real-time**: Counter denormalisasi memungkinkan dashboard kuota tanpa COUNT query.

### Negative / Trade-offs

- **Seasonal complexity**: Modul besar yang hanya aktif 2-3 bulan/tahun.
- **Public-facing security**: Perlu rate limiting, CAPTCHA, DDoS protection.
- **Data duplication**: Biodata di applicants mirror S001/S003/S007 — saat convert, data di-copy bukan di-reference.
- **Selection algorithm**: Scoring dan ranking per jalur memerlukan business logic yang complex dan configurable per sekolah.

## Alternatives Considered

### 1. Langsung buat student tanpa applicant stage
- Ditolak: tidak semua pendaftar diterima — perlu stage seleksi sebelum jadi student.

### 2. Applicant sebagai JSONB di admission_periods
- Ditolak: ribuan applicant per periode — perlu tabel proper untuk query, filter, dan ranking.

### 3. External PPDB system (Dapodik PPDB)
- Ditolak: PPDB Dapodik hanya untuk sekolah negeri. Sekolah swasta mengelola PPDB sendiri.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `PeriodDescriptor.TableName()` | — | `"admission_periods"` |
| U02 | `ApplicantDescriptor.TableName()` | — | `"applicants"` |
| U03 | Validate rejects invalid `track_type` | `"umum"` | Error |
| U04 | Validate rejects invalid `selection_status` | `"pending"` | Error |
| U05 | Validate rejects empty `guardian_name` | `""` | Error |
| U06 | Validate rejects empty `guardian_phone` | `""` | Error |
| U07 | Validate accepts valid applicant | All fields valid | No error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create admission period | POST with valid data | 201 |
| I02 | Create track with quota | POST track with `quota = 30` | 201 |
| I03 | Public registration | POST /public/apply | 201, registration_no generated |
| I04 | Unique registration_no | Apply twice → different numbers | Both unique |
| I05 | Check status (public) | GET /public/check/{no} | 200, current status |
| I06 | Verify applicant | PUT /verify | 200, status → verified |
| I07 | Run selection | POST /select for track | 200, scores + ranks calculated |
| I08 | Quota exceeded | Accept beyond quota | 422, quota full |
| I09 | Enroll applicant | POST /enroll | 201, student created, student_id set |
| I10 | Enrolled applicant has student data | Check students table | S001 + S003 + S007 created |
| I11 | Track counters updated | After registration + acceptance | `registered_count` + `accepted_count` correct |
| I12 | Period status CHECK | INSERT with `status = 'cancelled'` | DB error |
| I13 | Tenant isolation | Access other tenant's applicants | 404 |
