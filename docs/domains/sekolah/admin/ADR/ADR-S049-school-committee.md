# ADR-S049: School Committee (Komite Sekolah)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Komite Sekolah adalah lembaga mandiri yang beranggotakan orang tua/wali murid, tokoh masyarakat, dan pihak yang peduli terhadap pendidikan. Berdasarkan **Permendikbud No. 75 Tahun 2016**, komite sekolah memiliki fungsi:

1. **Pemberi pertimbangan**: Memberikan pertimbangan dalam penentuan dan pelaksanaan kebijakan pendidikan.
2. **Pendukung**: Mendukung penyelenggaraan pendidikan (finansial, pemikiran, tenaga).
3. **Pengontrol**: Mengawasi transparansi dan akuntabilitas penyelenggaraan pendidikan.
4. **Mediator**: Menjadi penghubung antara masyarakat dan satuan pendidikan.

Kebutuhan pengelolaan komite sekolah:

1. **Kepengurusan**: Pengurus komite (ketua, sekretaris, bendahara, anggota) perlu tercatat resmi — termasuk masa jabatan.
2. **Masa jabatan**: Berdasarkan regulasi, masa jabatan pengurus komite adalah **3 tahun** dan dapat dipilih kembali untuk **maksimal 2 periode**.
3. **Rapat**: Komite mengadakan rapat rutin (minimal 2x per semester) — perlu agenda, notulen, daftar hadir, dan keputusan.
4. **Persetujuan RKAS**: Komite memiliki peran krusial dalam review dan persetujuan RKAS (S050) — salah satu step dalam approval chain (S047).
5. **Representasi orang tua**: Anggota komite adalah perwakilan orang tua — terhubung ke data wali murid (S003).
6. **Pesantren**: Untuk pesantren (ADR-009), komite bisa disebut "Dewan Pengurus Yayasan" atau "Majelis Wali Amanah" — fungsi serupa.

### Mengapa Vernon Pattern?

- Read-heavy: data pengurus dan riwayat rapat lebih sering dibaca daripada ditulis.
- Relasi ke guardians (S003), users (ADR-013), budget approval (S050/S047).
- Volume rendah: 1 komite per sekolah, ~10-15 pengurus, ~4-8 rapat per tahun.
- Ideal untuk denormalisasi — member data di-cache di _data untuk quick display.

## Decision

Menggunakan **Vernon Pattern** untuk 4 tabel: `committee_members` (pengurus komite), `committee_terms` (masa jabatan/periode), `committee_meetings` (rapat), dan `committee_meeting_attendees` (kehadiran rapat).

### Table Schema

```sql
-- Periode kepengurusan komite
CREATE TABLE committee_terms (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Periode
    term_name       VARCHAR(100) NOT NULL,
    term_number     INT NOT NULL,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,

    -- SK Pengangkatan
    decree_number   VARCHAR(100),
    decree_date     DATE,
    decree_url      VARCHAR(500),

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT false,
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_term_dates CHECK (end_date > start_date),
    CONSTRAINT chk_term_number CHECK (term_number BETWEEN 1 AND 2),
    CONSTRAINT uq_term_tenant_company_number UNIQUE (tenant_id, company_id, term_number, start_date)
);

-- Indexes
CREATE INDEX idx_term_tenant_company ON committee_terms (tenant_id, company_id);
CREATE INDEX idx_term_active ON committee_terms (is_active) WHERE is_active = true;
CREATE INDEX idx_term_dates ON committee_terms (start_date, end_date);
CREATE INDEX idx_term_rels ON committee_terms USING GIN (_rels);
CREATE INDEX idx_term_data ON committee_terms USING GIN (_data);


-- Anggota/pengurus komite
CREATE TABLE committee_members (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Periode
    term_id         UUID NOT NULL,

    -- Identitas anggota
    full_name       VARCHAR(255) NOT NULL,
    position        VARCHAR(30) NOT NULL,
    phone           VARCHAR(30),
    email           VARCHAR(255),
    address         TEXT,
    occupation      VARCHAR(100),

    -- Representasi
    representation  VARCHAR(30) NOT NULL,
    guardian_id     UUID,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,
    joined_at       DATE NOT NULL,
    resigned_at     DATE,
    resign_reason   TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_member_position CHECK (position IN (
        'ketua', 'wakil_ketua', 'sekretaris', 'bendahara', 'anggota'
    )),
    CONSTRAINT chk_member_representation CHECK (representation IN (
        'orang_tua', 'tokoh_masyarakat', 'pakar_pendidikan',
        'dunia_usaha', 'alumni', 'tokoh_agama'
    )),
    CONSTRAINT uq_member_term_position_ketua UNIQUE (term_id, position)
        -- Note: only enforced at application level for ketua/sekretaris/bendahara
        -- Multiple anggota allowed per term
);

-- Indexes
CREATE INDEX idx_member_tenant_company ON committee_members (tenant_id, company_id);
CREATE INDEX idx_member_term ON committee_members (term_id);
CREATE INDEX idx_member_guardian ON committee_members (guardian_id) WHERE guardian_id IS NOT NULL;
CREATE INDEX idx_member_position ON committee_members (position);
CREATE INDEX idx_member_active ON committee_members (is_active) WHERE is_active = true;
CREATE INDEX idx_member_rels ON committee_members USING GIN (_rels);
CREATE INDEX idx_member_data ON committee_members USING GIN (_data);


-- Rapat komite
CREATE TABLE committee_meetings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Periode
    term_id         UUID NOT NULL,

    -- Meeting info
    meeting_number  INT NOT NULL,
    meeting_type    VARCHAR(20) NOT NULL,
    title           VARCHAR(500) NOT NULL,
    description     TEXT,

    -- Waktu & tempat
    meeting_date    DATE NOT NULL,
    start_time      TIME NOT NULL,
    end_time        TIME,
    location        VARCHAR(255) NOT NULL,

    -- Agenda
    agenda          JSONB NOT NULL DEFAULT '[]',

    -- Notulen
    minutes         TEXT,
    resolutions     JSONB NOT NULL DEFAULT '[]',

    -- File
    attachment_urls JSONB NOT NULL DEFAULT '[]',

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'scheduled',

    -- Pencatat
    recorded_by     UUID,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_meeting_type CHECK (meeting_type IN (
        'regular', 'extraordinary', 'annual', 'rkas_review'
    )),
    CONSTRAINT chk_meeting_status CHECK (status IN (
        'scheduled', 'in_progress', 'completed', 'cancelled'
    )),
    CONSTRAINT uq_meeting_number UNIQUE (term_id, meeting_number)
);

-- Indexes
CREATE INDEX idx_meeting_tenant_company ON committee_meetings (tenant_id, company_id);
CREATE INDEX idx_meeting_term ON committee_meetings (term_id);
CREATE INDEX idx_meeting_date ON committee_meetings (meeting_date);
CREATE INDEX idx_meeting_type ON committee_meetings (meeting_type);
CREATE INDEX idx_meeting_status ON committee_meetings (status);
CREATE INDEX idx_meeting_rels ON committee_meetings USING GIN (_rels);
CREATE INDEX idx_meeting_data ON committee_meetings USING GIN (_data);


-- Kehadiran rapat
CREATE TABLE committee_meeting_attendees (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Parent
    meeting_id      UUID NOT NULL,
    member_id       UUID NOT NULL,

    -- Kehadiran
    attendance_status VARCHAR(20) NOT NULL DEFAULT 'present',
    arrival_time    TIME,
    notes           TEXT,

    -- Tanda tangan digital (bukti kehadiran)
    signature_url   VARCHAR(500),

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_attendee_status CHECK (attendance_status IN (
        'present', 'absent', 'excused', 'late'
    )),
    CONSTRAINT uq_attendee_meeting_member UNIQUE (meeting_id, member_id)
);

-- Indexes
CREATE INDEX idx_attendee_tenant_company ON committee_meeting_attendees (tenant_id, company_id);
CREATE INDEX idx_attendee_meeting ON committee_meeting_attendees (meeting_id);
CREATE INDEX idx_attendee_member ON committee_meeting_attendees (member_id);
CREATE INDEX idx_attendee_status ON committee_meeting_attendees (attendance_status);
CREATE INDEX idx_attendee_rels ON committee_meeting_attendees USING GIN (_rels);
CREATE INDEX idx_attendee_data ON committee_meeting_attendees USING GIN (_data);
```

### Field Design Rationale

**committee_terms:**

| Field | Keputusan | Alasan |
|---|---|---|
| `term_number` | INT, CHECK 1-2 | Berdasarkan Permendikbud — maksimal 2 periode per susunan pengurus |
| `start_date` / `end_date` | DATE, NOT NULL | Masa jabatan 3 tahun — end_date = start_date + 3 tahun |
| `decree_number` | VARCHAR(100), nullable | Nomor SK pengangkatan dari kepala sekolah — nullable untuk draft |
| `is_active` | BOOLEAN | Hanya 1 term yang aktif per sekolah — divalidasi di application layer |

**committee_members:**

| Field | Keputusan | Alasan |
|---|---|---|
| `position` | VARCHAR(30), CHECK | 5 posisi standar: ketua, wakil_ketua, sekretaris, bendahara, anggota |
| `representation` | VARCHAR(30), CHECK | Kategori perwakilan — sesuai Permendikbud: orang tua, tokoh masyarakat, dsb |
| `guardian_id` | UUID, nullable | Link ke S003 guardians — nullable karena tokoh masyarakat bukan orang tua siswa |
| `occupation` | VARCHAR(100) | Pekerjaan anggota — ditampilkan di daftar pengurus resmi |
| `resigned_at` | DATE, nullable | Tanggal pengunduran diri — jika anggota keluar sebelum masa jabatan berakhir |

**committee_meetings:**

| Field | Keputusan | Alasan |
|---|---|---|
| `meeting_type` | VARCHAR(20), CHECK | 4 jenis: regular (rutin), extraordinary (luar biasa), annual (tahunan), rkas_review (review anggaran) |
| `agenda` | JSONB array | List agenda: `[{"order":1,"title":"Pembukaan","duration_minutes":10},...]` |
| `resolutions` | JSONB array | Keputusan rapat: `[{"number":1,"resolution":"Menyetujui RKAS 2026/2027","vote":"unanimous"}]` |
| `minutes` | TEXT, nullable | Notulen lengkap — nullable karena diisi setelah rapat selesai |
| `recorded_by` | UUID, nullable | Notulis rapat — biasanya sekretaris komite atau staf TU |

**committee_meeting_attendees:**

| Field | Keputusan | Alasan |
|---|---|---|
| `attendance_status` | VARCHAR(20), CHECK | 4 status: present, absent, excused, late |
| `signature_url` | VARCHAR(500), nullable | Bukti kehadiran digital — pengganti tanda tangan di daftar hadir fisik |

### Vernon Relationships

**committee_terms:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `members` | has_many | Tidak | Dimuat terpisah di detail view |
| `meetings` | has_many | Tidak | Dimuat terpisah di tab meetings |

**committee_members:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `term` | belongs_to | **Ya** | Konteks periode kepengurusan |
| `guardian` | belongs_to | Tidak | Hanya dimuat jika perlu link ke data wali murid |

**committee_meetings:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `term` | belongs_to | **Ya** | Konteks periode |
| `attendees` | has_many | Tidak | Dimuat terpisah di detail meeting |
| `recorded_by_user` | belongs_to | **Ya** | Notulis rapat |

**committee_meeting_attendees:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `meeting` | belongs_to | **Ya** | Konteks rapat |
| `member` | belongs_to | **Ya** | Identitas peserta |

### _rels / _data Structure

**committee_terms:**

```json
{
  "_rels": {},
  "_data": {
    "members_summary": {
      "total": 11,
      "ketua": "Bapak Hadi Santoso",
      "sekretaris": "Ibu Rina Wati",
      "bendahara": "Bapak Joko Widodo",
      "anggota_count": 8
    },
    "meetings_count": 4,
    "is_expired": false
  }
}
```

**committee_members:**

```json
{
  "_rels": {
    "term_id": "018f...",
    "guardian_id": "018f..."
  },
  "_data": {
    "term": {
      "id": "018f...",
      "term_name": "Periode 2024-2027",
      "is_active": true
    },
    "guardian": {
      "id": "018f...",
      "full_name": "Bapak Hadi Santoso",
      "children": ["Ahmad Hadi (VII-A)", "Siti Hadi (IX-B)"]
    }
  }
}
```

**committee_meetings:**

```json
{
  "_rels": {
    "term_id": "018f...",
    "recorded_by": "018f..."
  },
  "_data": {
    "term": {
      "id": "018f...",
      "term_name": "Periode 2024-2027"
    },
    "recorded_by_user": {
      "id": "018f...",
      "full_name": "Ibu Siti (TU)"
    },
    "attendance_summary": {
      "total_invited": 11,
      "present": 9,
      "absent": 1,
      "excused": 1,
      "quorum_met": true
    }
  }
}
```

**committee_meeting_attendees:**

```json
{
  "_rels": {
    "meeting_id": "018f...",
    "member_id": "018f..."
  },
  "_data": {
    "meeting": {
      "id": "018f...",
      "title": "Rapat Review RKAS 2026/2027",
      "meeting_date": "2026-04-10"
    },
    "member": {
      "id": "018f...",
      "full_name": "Bapak Hadi Santoso",
      "position": "ketua"
    }
  }
}
```

### Integration with S050 (Budget/RKAS) via S047 (Approval Workflow)

Komite sekolah berperan sebagai salah satu approver dalam chain persetujuan RKAS:

```
S050 (RKAS) → S047 (Approval Engine):
  Chain: bendahara → kepsek → komite → dinas
  Step 3: role = "komite" → resolved to ketua komite user

Alur:
1. Bendahara menyusun RKAS → submit approval request
2. Kepsek review & approve → step 2 approved
3. Ketua komite mendapat notifikasi di inbox S047
4. Komite mengadakan rapat review RKAS (meeting_type = 'rkas_review')
5. Hasil rapat: resolusi "Menyetujui RKAS" → ketua komite approve di S047
6. Dinas review & approve → RKAS final
```

### API Endpoints

```
# Committee Terms (Periode)
POST   /api/v1/committee-terms                              — Buat periode baru
GET    /api/v1/committee-terms                              — List semua periode
GET    /api/v1/committee-terms/{id}                         — Detail periode + members
GET    /api/v1/committee-terms/active                       — Get periode aktif saat ini
PUT    /api/v1/committee-terms/{id}                         — Update periode
PUT    /api/v1/committee-terms/{id}/activate                — Set sebagai periode aktif

# Committee Members (Pengurus)
POST   /api/v1/committee-terms/{term_id}/members            — Tambah anggota
GET    /api/v1/committee-terms/{term_id}/members            — List anggota per periode
GET    /api/v1/committee-members/{id}                       — Detail anggota
PUT    /api/v1/committee-members/{id}                       — Update data anggota
PUT    /api/v1/committee-members/{id}/resign                — Catat pengunduran diri
DELETE /api/v1/committee-members/{id}                       — Soft delete

# Committee Meetings (Rapat)
POST   /api/v1/committee-meetings                           — Jadwalkan rapat baru
GET    /api/v1/committee-meetings                           — List rapat (filter: term_id, type, status)
GET    /api/v1/committee-meetings/{id}                      — Detail rapat + attendees + resolutions
PUT    /api/v1/committee-meetings/{id}                      — Update rapat (agenda, notulen)
PUT    /api/v1/committee-meetings/{id}/complete             — Selesaikan rapat (finalize minutes)
DELETE /api/v1/committee-meetings/{id}                      — Cancel/delete rapat

# Meeting Attendance
POST   /api/v1/committee-meetings/{id}/attendees/bulk       — Bulk insert kehadiran
GET    /api/v1/committee-meetings/{id}/attendees            — List kehadiran rapat
PUT    /api/v1/committee-meeting-attendees/{id}             — Update status kehadiran
```

## Consequences

### Positive

- **Tata kelola resmi**: Kepengurusan komite tercatat digital — lengkap dengan SK, masa jabatan, dan riwayat.
- **Rapat terdokumentasi**: Agenda, notulen, keputusan rapat tersimpan — tidak ada lagi "lupa apa keputusan rapat kemarin".
- **Integrasi RKAS**: Peran komite dalam persetujuan anggaran terhubung langsung ke S047/S050.
- **Link ke orang tua**: Anggota komite yang merupakan wali murid terhubung ke S003 — bisa lihat profil lengkap.
- **Compliance**: Tracking masa jabatan membantu memastikan kepatuhan terhadap Permendikbud (max 2 periode).
- **Transparansi**: Riwayat rapat dan keputusan bisa diakses oleh pihak berwenang untuk audit.

### Negative / Trade-offs

- **Low volume domain**: Hanya ~10-15 pengurus dan ~4-8 rapat per tahun — effort pengembangan mungkin tidak proporsional dengan frekuensi penggunaan.
- **Member position uniqueness**: Constraint unik untuk ketua/sekretaris/bendahara per term harus di-enforce di application layer — DB constraint terlalu rigid untuk posisi "anggota" yang bisa multiple.
- **Guardian link optional**: Tidak semua anggota komite adalah orang tua siswa — guardian_id bisa null, membuat link tidak selalu tersedia.
- **Quorum calculation**: Logika quorum (minimal 50%+1 hadir) perlu di-enforce di application layer.
- **Meeting scheduling**: Belum ada integrasi kalender — scheduling manual.

## Alternatives Considered

### 1. Komite sebagai sub-domain dari S003 (Guardians)
- Ditolak: tidak semua anggota komite adalah wali murid. Komite adalah entity mandiri dengan lifecycle sendiri (periode, rapat, keputusan).

### 2. Rapat komite digabung dengan meeting system general
- Deferred: general meeting system bisa dibangun nanti. Untuk MVP, rapat komite cukup spesifik (ada agenda RKAS, resolusi, quorum) sehingga tabel sendiri lebih tepat.

### 3. Keputusan rapat sebagai tabel terpisah (committee_resolutions)
- Ditolak: resolusi selalu terikat pada rapat — JSONB array di `committee_meetings.resolutions` cukup. Volume sangat rendah (~2-5 resolusi per rapat).

### 4. Komite multi-level (komite kelas + komite sekolah)
- Deferred: MVP fokus pada komite sekolah (level paling atas). Komite kelas bisa ditambahkan sebagai enhancement jika ada kebutuhan.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` returns correct name | — | `"committee_members"` |
| U02 | `DefaultRels()` returns autoloaded rels | — | `term` |
| U03 | Validate rejects invalid position | `{ "position": "direktur" }` | Error: position not in allowed list |
| U04 | Validate rejects invalid representation | `{ "representation": "pemerintah" }` | Error: representation not in allowed list |
| U05 | Validate term_number max 2 | `{ "term_number": 3 }` | Error: term_number must be 1 or 2 |
| U06 | Validate end_date > start_date | `{ "start_date": "2026-01-01", "end_date": "2025-01-01" }` | Error: end_date must be after start_date |
| U07 | Validate meeting_type | `{ "meeting_type": "informal" }` | Error: meeting_type not in allowed list |
| U08 | Validate attendance_status | `{ "attendance_status": "sick" }` | Error: must be present/absent/excused/late |
| U09 | Quorum calculation | 9 present out of 11 invited | quorum_met = true (>50%) |
| U10 | Quorum not met | 4 present out of 11 invited | quorum_met = false |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create committee term | POST with dates and term_number=1 | 201, term created |
| I02 | Add members to term | POST 11 members (ketua, sekretaris, bendahara, 8 anggota) | 201, all created with _data |
| I03 | Reject duplicate ketua | POST second member with position=ketua for same term | 409/422, only 1 ketua per term |
| I04 | Allow multiple anggota | POST multiple members with position=anggota | 201, all created |
| I05 | Schedule meeting | POST with agenda items | 201, status=scheduled |
| I06 | Record attendance (bulk) | POST bulk attendees for meeting | 201, all attendees recorded |
| I07 | Complete meeting with minutes | PUT complete with minutes + resolutions | 200, status=completed |
| I08 | Get active term with members | GET /committee-terms/active | 200, includes members_summary in _data |
| I09 | Member resignation | PUT resign with reason | 200, resigned_at set, is_active=false |
| I10 | Activate new term | PUT activate on term 2 | 200, term 1 deactivated, term 2 active |
| I11 | Link member to guardian | POST member with guardian_id | 201, _data includes guardian children |
| I12 | RKAS review meeting type | POST meeting with type=rkas_review | 201, tagged as RKAS review |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | GuardianUpdated syncs to member | Update guardian name (S003) | `_data.guardian.full_name` updated |
| I14 | MemberUpdated syncs to attendees | Update member name | `_data.member.full_name` updated |
| I15 | MeetingUpdated syncs to attendees | Update meeting title | `_data.meeting.title` updated |
| I16 | TermUpdated syncs to members | Update term name | `_data.term.term_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Cannot access other tenant's committee | GET with wrong tenant scope | 404 |
| I18 | Cannot add member to other tenant's term | POST member with cross-tenant term_id | 403/404, scope violation |
| I19 | Committee data per company | Company A and B each have own committee | Independent data, no cross-contamination |
