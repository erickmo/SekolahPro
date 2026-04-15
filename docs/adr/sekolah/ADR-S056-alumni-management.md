# ADR-S056: Alumni Management (Manajemen Alumni)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Alumni adalah aset jangka panjang sekolah. Setelah siswa lulus (graduated dari S016 PPDB atau proses kelulusan), data mereka perlu tetap dikelola untuk berbagai keperluan:

1. **Direktori alumni**: Database alumni yang bisa dicari berdasarkan angkatan, nama, lokasi — untuk networking dan silaturahmi.
2. **Tracking karir & pendidikan**: Kemana alumni melanjutkan (SMP → SMA, SMA → PT) dan karirnya — untuk laporan akreditasi dan marketing sekolah.
3. **Achievement tracking**: Prestasi alumni di luar sekolah — untuk membangun reputasi dan branding sekolah.
4. **Event management**: Reuni, homecoming, alumni gathering — mempererat ikatan alumni.
5. **Donasi & kontribusi**: Alumni yang sudah mapan sering ingin berkontribusi ke sekolah (terutama yayasan) — perlu tracking donasi.
6. **Akreditasi**: BAN-S/M menanyakan data alumni sebagai indikator kualitas sekolah — kemana mereka melanjutkan, berapa yang diterima di PTN, dll.
7. **Pesantren**: Alumni pesantren sering menjadi ustadz/ustadzah di pesantren lain — tracking ini penting untuk jaringan pesantren.

### Mengapa Vernon Pattern?

- Read-heavy: direktori alumni, dashboard prestasi, laporan akreditasi — lebih banyak dibaca daripada ditulis.
- Relasi ke student (data asli), academic_year (angkatan), events.
- Write infrequent: data alumni di-update sporadis (update karir, tambah prestasi, catat donasi).
- Business logic moderate: auto-convert dari graduated student, search, aggregasi per angkatan.

## Decision

Menggunakan **Vernon Pattern** untuk 4 tabel: `alumni` (profil alumni), `alumni_achievements` (prestasi alumni), `alumni_events` (kegiatan alumni), dan `alumni_donations` (donasi/kontribusi alumni).

### Table Schema

```sql
-- Profil alumni
CREATE TABLE alumni (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Link ke data siswa asli
    student_id      UUID NOT NULL,

    -- Identitas (snapshot dari student saat lulus)
    full_name       VARCHAR(255) NOT NULL,
    nickname        VARCHAR(100),
    gender          VARCHAR(1) NOT NULL,
    birth_date      DATE,
    photo_url       TEXT,
    nisn            VARCHAR(10),

    -- Data kelulusan
    graduation_year INT NOT NULL,
    graduation_class VARCHAR(50) NOT NULL,
    graduation_academic_year_id UUID NOT NULL,
    certificate_no  VARCHAR(50),

    -- Pendidikan lanjutan
    further_education_level VARCHAR(30),
    further_education_name  VARCHAR(255),
    further_education_major VARCHAR(255),
    further_education_city  VARCHAR(100),

    -- Karir
    current_occupation VARCHAR(255),
    current_company    VARCHAR(255),
    current_position   VARCHAR(255),
    current_city       VARCHAR(100),

    -- Kontak
    email           VARCHAR(255),
    phone           VARCHAR(20),
    whatsapp        VARCHAR(20),
    address         TEXT,
    social_media    JSONB NOT NULL DEFAULT '{}',

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    is_contactable  BOOLEAN NOT NULL DEFAULT true,
    last_updated_by VARCHAR(20) NOT NULL DEFAULT 'system',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_alumni_student UNIQUE (tenant_id, company_id, student_id),
    CONSTRAINT chk_alumni_gender CHECK (gender IN ('L', 'P')),
    CONSTRAINT chk_alumni_status CHECK (status IN ('active', 'inactive', 'deceased', 'unreachable')),
    CONSTRAINT chk_alumni_updated_by CHECK (last_updated_by IN ('system', 'admin', 'self')),
    CONSTRAINT chk_further_edu_level CHECK (further_education_level IS NULL OR further_education_level IN ('smp', 'sma', 'smk', 'd3', 'd4', 's1', 's2', 's3', 'pesantren', 'kursus', 'kerja', 'other'))
);

CREATE INDEX idx_alumni_tenant ON alumni (tenant_id, company_id);
CREATE INDEX idx_alumni_student ON alumni (student_id);
CREATE INDEX idx_alumni_year ON alumni (graduation_year);
CREATE INDEX idx_alumni_name ON alumni (full_name);
CREATE INDEX idx_alumni_status ON alumni (status) WHERE status = 'active';
CREATE INDEX idx_alumni_rels ON alumni USING GIN (_rels);
CREATE INDEX idx_alumni_data ON alumni USING GIN (_data);

-- Prestasi alumni
CREATE TABLE alumni_achievements (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    alumni_id       UUID NOT NULL,

    -- Detail prestasi
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    achievement_type VARCHAR(30) NOT NULL,
    achievement_date DATE,
    institution     VARCHAR(255),
    level           VARCHAR(20),

    -- Media
    proof_url       TEXT,
    is_verified     BOOLEAN NOT NULL DEFAULT false,
    verified_by     UUID,
    verified_at     TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_achievement_type CHECK (achievement_type IN ('academic', 'career', 'entrepreneurship', 'social', 'sports', 'arts', 'religious', 'other')),
    CONSTRAINT chk_achievement_level CHECK (level IS NULL OR level IN ('local', 'regional', 'national', 'international'))
);

CREATE INDEX idx_alumni_achievement_tenant ON alumni_achievements (tenant_id, company_id);
CREATE INDEX idx_alumni_achievement_alumni ON alumni_achievements (alumni_id);
CREATE INDEX idx_alumni_achievement_type ON alumni_achievements (achievement_type);
CREATE INDEX idx_alumni_achievement_rels ON alumni_achievements USING GIN (_rels);
CREATE INDEX idx_alumni_achievement_data ON alumni_achievements USING GIN (_data);

-- Kegiatan alumni (reuni, homecoming, gathering)
CREATE TABLE alumni_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    event_type      VARCHAR(30) NOT NULL,

    -- Jadwal & lokasi
    event_date      DATE NOT NULL,
    start_time      TIME,
    end_time        TIME,
    location        VARCHAR(255),
    is_online       BOOLEAN NOT NULL DEFAULT false,
    online_link     TEXT,

    -- Target angkatan
    target_years    JSONB NOT NULL DEFAULT '[]',
    is_all_alumni   BOOLEAN NOT NULL DEFAULT false,

    -- Stats
    registered_count INT NOT NULL DEFAULT 0,
    attended_count  INT NOT NULL DEFAULT 0,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    organized_by    UUID NOT NULL,

    -- Dokumentasi
    documentation   JSONB NOT NULL DEFAULT '{}',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_event_type CHECK (event_type IN ('reuni', 'homecoming', 'gathering', 'seminar', 'mentoring', 'fundraising', 'haflah', 'other')),
    CONSTRAINT chk_event_status CHECK (status IN ('draft', 'published', 'registration_open', 'ongoing', 'completed', 'cancelled'))
);

CREATE INDEX idx_alumni_event_tenant ON alumni_events (tenant_id, company_id);
CREATE INDEX idx_alumni_event_date ON alumni_events (event_date);
CREATE INDEX idx_alumni_event_status ON alumni_events (status) WHERE status NOT IN ('completed', 'cancelled');
CREATE INDEX idx_alumni_event_rels ON alumni_events USING GIN (_rels);
CREATE INDEX idx_alumni_event_data ON alumni_events USING GIN (_data);

-- Donasi / kontribusi alumni
CREATE TABLE alumni_donations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    alumni_id       UUID NOT NULL,
    event_id        UUID,

    -- Detail donasi
    donation_type   VARCHAR(20) NOT NULL,
    amount          BIGINT,
    item_description TEXT,
    purpose         VARCHAR(255),

    -- Pembayaran
    payment_method  VARCHAR(20),
    payment_date    DATE NOT NULL,
    receipt_no      VARCHAR(50),

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'received',
    received_by     UUID NOT NULL,
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

    CONSTRAINT chk_donation_type CHECK (donation_type IN ('money', 'goods', 'service', 'scholarship_fund', 'infrastructure', 'other')),
    CONSTRAINT chk_donation_status CHECK (status IN ('pledged', 'received', 'acknowledged', 'returned')),
    CONSTRAINT chk_donation_payment CHECK (payment_method IS NULL OR payment_method IN ('cash', 'transfer', 'qris', 'va', 'other'))
);

CREATE INDEX idx_alumni_donation_tenant ON alumni_donations (tenant_id, company_id);
CREATE INDEX idx_alumni_donation_alumni ON alumni_donations (alumni_id);
CREATE INDEX idx_alumni_donation_event ON alumni_donations (event_id) WHERE event_id IS NOT NULL;
CREATE INDEX idx_alumni_donation_date ON alumni_donations (payment_date);
CREATE INDEX idx_alumni_donation_rels ON alumni_donations USING GIN (_rels);
CREATE INDEX idx_alumni_donation_data ON alumni_donations USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `student_id` | UUID, NOT NULL, unique per company | Setiap alumni pasti berasal dari student — link 1:1 |
| Biodata fields | Snapshot dari student | Data di-copy saat lulus — jika student di-update, alumni tetap memiliki data saat kelulusan |
| `graduation_year` | INT, NOT NULL | Angkatan — field utama untuk pengelompokan dan pencarian alumni |
| `graduation_class` | VARCHAR(50) | Kelas terakhir saat lulus — "XII IPA 1", "IX-A", dll |
| `further_education_level` | VARCHAR(30), nullable | Jenjang lanjutan — dari SMP ke SMA, SMA ke PT. Include 'pesantren' dan 'kerja' |
| `social_media` | JSONB | Flexible — bisa Instagram, LinkedIn, Twitter, Facebook |
| `last_updated_by` | VARCHAR(20) | Track apakah data di-update oleh system (auto), admin, atau alumni sendiri (self-service) |
| `target_years` | JSONB array di events | Event bisa target angkatan tertentu [2020, 2021] atau semua (is_all_alumni) |
| `documentation` | JSONB di events | Flexible: { "photos": [...], "report_url": "...", "video_url": "..." } |
| `donation_type` | VARCHAR(20) | Tidak hanya uang — bisa barang, jasa, dana beasiswa, infrastruktur |

### Student → Alumni Conversion Flow

```
┌─────────────────────────────────────────────────────────────┐
│ Student Graduation → Alumni Conversion                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ 1. Siswa dinyatakan LULUS (dari proses kelulusan/rapor)     │
│    Student status → 'graduated'                              │
│                                                             │
│ 2. System auto-creates alumni record:                        │
│    a. Copy biodata dari student (S001)                       │
│    b. Set graduation_year dari academic_year                 │
│    c. Set graduation_class dari student_class_placement      │
│    d. Set NISN dari student                                  │
│    e. Link student_id                                        │
│                                                             │
│ 3. Alumni record status = 'active'                           │
│                                                             │
│ 4. Alumni bisa update profil sendiri (self-service):         │
│    - Pendidikan lanjutan                                     │
│    - Karir                                                   │
│    - Kontak                                                  │
│    - Social media                                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Vernon Relationships

**alumni:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | Tidak | Data sudah di-snapshot — hanya perlu untuk deep lookup |
| `graduation_academic_year` | belongs_to | **Ya** | Konteks tahun ajaran kelulusan |

**alumni_achievements:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `alumni` | belongs_to | **Ya** | Selalu perlu nama dan angkatan alumni |

**alumni_events:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `organized_by` (user) | belongs_to | **Ya** | Siapa yang mengorganisir event |

**alumni_donations:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `alumni` | belongs_to | **Ya** | Selalu perlu nama donatur |
| `event` | belongs_to | Tidak | Opsional — donasi bisa tanpa event |

### _rels / _data Structure

```json
// alumni
{
  "_rels": {
    "student_id": "018f...",
    "graduation_academic_year_id": "018f..."
  },
  "_data": {
    "graduation_academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}

// alumni_achievements
{
  "_rels": {
    "alumni_id": "018f..."
  },
  "_data": {
    "alumni": { "id": "018f...", "full_name": "Ahmad Fauzi", "graduation_year": 2025, "graduation_class": "XII IPA 1" }
  }
}

// alumni_donations
{
  "_rels": {
    "alumni_id": "018f...",
    "event_id": "018f..."
  },
  "_data": {
    "alumni": { "id": "018f...", "full_name": "Siti Rahayu", "graduation_year": 2010 },
    "event": { "id": "018f...", "name": "Reuni Akbar 2026", "event_type": "reuni" }
  }
}
```

### API Endpoints

```
# Alumni Directory
GET    /api/v1/alumni                                    — List alumni (search, filter by year)
GET    /api/v1/alumni/{id}                               — Detail profil alumni
POST   /api/v1/alumni                                    — Buat alumni manual (import, edge case)
PUT    /api/v1/alumni/{id}                               — Update profil alumni
PUT    /api/v1/alumni/{id}/self-update                   — Self-service update oleh alumni

# Auto-conversion
POST   /api/v1/alumni/convert-graduates                  — Batch convert graduated students ke alumni

# Achievements
GET    /api/v1/alumni/{id}/achievements                  — List prestasi per alumni
POST   /api/v1/alumni-achievements                       — Tambah prestasi
PUT    /api/v1/alumni-achievements/{id}                  — Update prestasi
PUT    /api/v1/alumni-achievements/{id}/verify           — Verifikasi prestasi (admin)

# Events
GET    /api/v1/alumni-events                             — List event alumni
POST   /api/v1/alumni-events                             — Buat event
PUT    /api/v1/alumni-events/{id}                        — Update event
POST   /api/v1/alumni-events/{id}/register               — Daftar ke event
GET    /api/v1/alumni-events/{id}/attendees              — List peserta

# Donations
GET    /api/v1/alumni-donations                          — List donasi
POST   /api/v1/alumni-donations                          — Catat donasi
GET    /api/v1/alumni/{id}/donations                     — Donasi per alumni
GET    /api/v1/alumni-donations/summary                  — Rekap donasi per tahun

# Reports
GET    /api/v1/alumni/stats                              — Statistik alumni (per angkatan, pendidikan lanjutan)
GET    /api/v1/alumni/further-education-summary           — Rekap pendidikan lanjutan (untuk akreditasi)
```

## Consequences

### Positive

- **Auto-conversion**: Graduated students otomatis menjadi alumni — zero manual entry.
- **Self-service update**: Alumni bisa update profil sendiri — data tetap fresh tanpa beban admin.
- **Akreditasi ready**: Data alumni (pendidikan lanjutan, prestasi) langsung tersedia untuk laporan akreditasi BAN-S/M.
- **Fundraising support**: Tracking donasi memudahkan yayasan mengelola kontribusi alumni.
- **Networking**: Direktori alumni memfasilitasi koneksi antar alumni dan sekolah.
- **Pesantren friendly**: Event types include haflah — relevan untuk pesantren.

### Negative / Trade-offs

- **Data decay**: Kontak alumni cepat berubah — email/phone bisa expired. Perlu mekanisme periodik update reminder.
- **Privacy concern**: Direktori alumni perlu consent management — tidak semua alumni mau data mereka publicly searchable.
- **Self-service authentication**: Alumni perlu akun terpisah atau OTP-based auth — tambah complexity auth system.
- **Snapshot vs reference**: Data alumni di-snapshot saat lulus — jika ada koreksi data student, alumni tidak auto-update. Trade-off yang disengaja untuk data integrity historical.

## Alternatives Considered

### 1. Tetap di tabel students dengan status 'graduated'
- Ditolak: alumni punya field berbeda (karir, pendidikan lanjutan, donasi) yang tidak relevan untuk siswa aktif. Memisahkan domain lebih clean.

### 2. External alumni platform (LinkedIn-like)
- Ditolak: sekolah butuh kontrol penuh atas data alumni. External platform tidak terintegrasi dengan data internal sekolah.

### 3. Alumni sebagai extension di students table (JSONB)
- Ditolak: alumni data yang bertumbuh (achievements, donations, events) memerlukan tabel terpisah untuk query yang efisien.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `AlumniDescriptor.TableName()` | — | `"alumni"` |
| U02 | `AchievementDescriptor.TableName()` | — | `"alumni_achievements"` |
| U03 | `EventDescriptor.TableName()` | — | `"alumni_events"` |
| U04 | `DonationDescriptor.TableName()` | — | `"alumni_donations"` |
| U05 | Validate rejects invalid `status` | `"deleted"` | Error: invalid status |
| U06 | Validate rejects invalid `gender` | `"M"` | Error: must be L or P |
| U07 | Validate rejects invalid `achievement_type` | `"hobby"` | Error: invalid achievement_type |
| U08 | Validate rejects invalid `event_type` | `"party"` | Error: invalid event_type |
| U09 | Validate rejects invalid `donation_type` | `"crypto"` | Error: invalid donation_type |
| U10 | Validate accepts valid alumni | All fields valid | No error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Auto-convert graduated student | POST /convert-graduates for angkatan 2025 | Alumni records created for all graduated students |
| I02 | Unique student_id per company | Convert same student twice | 409/422, unique constraint |
| I03 | Alumni snapshot has correct data | Check alumni after conversion | full_name, NISN, graduation_class match student data |
| I04 | Search alumni by name | GET /alumni?search=Ahmad | Results filtered by name |
| I05 | Filter alumni by year | GET /alumni?graduation_year=2025 | Only 2025 graduates |
| I06 | Self-service update | PUT /self-update with karir data | 200, last_updated_by = 'self' |
| I07 | Add achievement | POST alumni-achievements | 201, linked to alumni |
| I08 | Verify achievement | PUT /verify by admin | is_verified = true, verified_by set |
| I09 | Create event | POST alumni-events | 201, status = 'draft' |
| I10 | Register to event | POST /register | 200, registered_count incremented |
| I11 | Record donation | POST alumni-donations | 201, linked to alumni |
| I12 | Donation with event link | POST with event_id | 201, linked to both alumni and event |
| I13 | Further education summary | GET /further-education-summary | Correct counts per level (SMA, S1, S2, etc.) |
| I14 | Event type CHECK enforced | INSERT with `event_type = 'party'` | DB error |
| I15 | Alumni status CHECK enforced | INSERT with `status = 'deleted'` | DB error |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | AcademicYearUpdated syncs to alumni | Update academic year name | `_data.graduation_academic_year.name` updated |
| I17 | AlumniUpdated syncs to achievements | Update alumni name | `_data.alumni.full_name` updated |
| I18 | AlumniUpdated syncs to donations | Update alumni name | `_data.alumni.full_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I19 | Cannot access other tenant's alumni | GET with wrong tenant | 404 |
| I20 | Cannot donate to other tenant's event | POST donation cross-tenant | Error |
