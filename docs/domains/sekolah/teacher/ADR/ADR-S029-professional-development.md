# ADR-S029: Professional Development (Pengembangan Profesi Berkelanjutan - PKB)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Pengembangan Keprofesian Berkelanjutan (PKB) adalah kewajiban guru berdasarkan **Permenneg PAN & RB No. 16 Tahun 2009** dan **Peraturan Bersama Mendiknas & BKN No. 14 Tahun 2010**. PKB terdiri dari 3 komponen:

1. **Pengembangan Diri**: Diklat fungsional (pelatihan, seminar, workshop, lokakarya) dan kegiatan kolektif guru (KKG/MGMP).
2. **Publikasi Ilmiah**: Penelitian, artikel, buku, modul, media pembelajaran.
3. **Karya Inovatif**: Teknologi tepat guna, alat peraga, karya seni.

PKB diperlukan untuk:

- **Angka Kredit**: Setiap kegiatan PKB bernilai angka kredit yang dibutuhkan untuk kenaikan pangkat PNS.
- **Sertifikasi Guru**: Guru harus memiliki sertifikat pendidik — tracking status sertifikasi dan perpanjangan.
- **SKP (Sasaran Kinerja Pegawai)**: PKB menjadi bagian dari target kinerja tahunan ASN.
- **Dapodik**: Pelaporan riwayat diklat dan sertifikasi ke Kemendikbud.
- **BKN**: Pelaporan angka kredit untuk kenaikan pangkat.
- **Hasil PKG (S028)**: Area "kurang" di PKG menjadi prioritas pengembangan PKB.

Jenjang karir guru PNS:

| Jabatan | Pangkat | Golongan | Angka Kredit |
|---------|---------|----------|--------------|
| Guru Pertama | Penata Muda III/a | III/a - III/b | 100 |
| Guru Muda | Penata III/c | III/c - III/d | 200 |
| Guru Madya | Pembina IV/a | IV/a - IV/c | 300 |
| Guru Utama | Pembina Utama Madya IV/d | IV/d - IV/e | 400 |

### Mengapa Vernon Pattern?

- has_many dari teacher: banyak kegiatan PKB per guru sepanjang karir.
- Read-heavy: laporan kredit, riwayat diklat, dashboard pengembangan.
- Relasi ke teacher dan academic_year.
- Business logic moderate: hitung total kredit, cek threshold kenaikan pangkat.
- Eventually consistent acceptable — pengembangan profesi bersifat ongoing.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `teacher_certifications` (sertifikasi guru), `teacher_development_activities` (kegiatan PKB), dan `teacher_credit_points` (ringkasan angka kredit).

### Table Schema

```sql
-- Sertifikasi guru
CREATE TABLE teacher_certifications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,

    -- Sertifikasi
    certification_type VARCHAR(30) NOT NULL,
    certificate_number VARCHAR(100),
    certificate_name VARCHAR(255) NOT NULL,
    issuing_institution VARCHAR(255) NOT NULL,
    issue_date      DATE NOT NULL,
    expiry_date     DATE,
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Dokumen
    document_url    TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_cert_type CHECK (certification_type IN (
        'sertifikasi_pendidik', 'sertifikasi_kompetensi',
        'sertifikasi_keahlian', 'ijazah', 'akta_mengajar',
        'pelatihan', 'other'
    ))
);

-- Indexes
CREATE INDEX idx_cert_tenant_company ON teacher_certifications (tenant_id, company_id);
CREATE INDEX idx_cert_teacher ON teacher_certifications (teacher_id);
CREATE INDEX idx_cert_type ON teacher_certifications (certification_type);
CREATE INDEX idx_cert_expiry ON teacher_certifications (expiry_date) WHERE expiry_date IS NOT NULL;
CREATE INDEX idx_cert_rels ON teacher_certifications USING GIN (_rels);
CREATE INDEX idx_cert_data ON teacher_certifications USING GIN (_data);

-- Kegiatan pengembangan profesi (PKB)
CREATE TABLE teacher_development_activities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    academic_year_id UUID,

    -- Kegiatan
    activity_type   VARCHAR(30) NOT NULL,
    pkb_component   VARCHAR(30) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    organizer       VARCHAR(255),
    location        VARCHAR(255),

    -- Waktu
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    total_hours     NUMERIC(6,1) NOT NULL DEFAULT 0,

    -- Angka kredit
    credit_points   NUMERIC(6,2) NOT NULL DEFAULT 0,
    credit_approved BOOLEAN NOT NULL DEFAULT false,
    credit_approved_by UUID,
    credit_approved_at TIMESTAMPTZ,

    -- Dokumen
    certificate_url TEXT,
    report_url      TEXT,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'planned',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_activity_type CHECK (activity_type IN (
        'diklat', 'seminar', 'workshop', 'lokakarya',
        'kkg_mgmp', 'penelitian', 'publikasi', 'buku',
        'modul', 'media_pembelajaran', 'karya_inovatif',
        'magang', 'studi_lanjut', 'other'
    )),
    CONSTRAINT chk_pkb_component CHECK (pkb_component IN (
        'pengembangan_diri', 'publikasi_ilmiah', 'karya_inovatif'
    )),
    CONSTRAINT chk_activity_status CHECK (status IN (
        'planned', 'in_progress', 'completed', 'cancelled'
    )),
    CONSTRAINT chk_credit_points CHECK (credit_points >= 0),
    CONSTRAINT chk_total_hours CHECK (total_hours >= 0),
    CONSTRAINT chk_date_range CHECK (end_date >= start_date)
);

-- Indexes
CREATE INDEX idx_dev_activity_tenant_company ON teacher_development_activities (tenant_id, company_id);
CREATE INDEX idx_dev_activity_teacher ON teacher_development_activities (teacher_id);
CREATE INDEX idx_dev_activity_year ON teacher_development_activities (academic_year_id);
CREATE INDEX idx_dev_activity_type ON teacher_development_activities (activity_type);
CREATE INDEX idx_dev_activity_component ON teacher_development_activities (pkb_component);
CREATE INDEX idx_dev_activity_status ON teacher_development_activities (status);
CREATE INDEX idx_dev_activity_rels ON teacher_development_activities USING GIN (_rels);
CREATE INDEX idx_dev_activity_data ON teacher_development_activities USING GIN (_data);

-- Ringkasan angka kredit per guru (snapshot)
CREATE TABLE teacher_credit_summaries (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,

    -- Jabatan fungsional saat ini
    current_rank    VARCHAR(30) NOT NULL,
    current_grade   VARCHAR(10) NOT NULL,

    -- Angka kredit
    credit_total    NUMERIC(8,2) NOT NULL DEFAULT 0,
    credit_pengembangan_diri NUMERIC(8,2) NOT NULL DEFAULT 0,
    credit_publikasi NUMERIC(8,2) NOT NULL DEFAULT 0,
    credit_karya_inovatif NUMERIC(8,2) NOT NULL DEFAULT 0,

    -- Target kenaikan pangkat
    target_rank     VARCHAR(30),
    target_credit   NUMERIC(8,2),
    credit_gap      NUMERIC(8,2),

    -- SKP
    skp_year        INT,
    skp_target      TEXT,
    skp_realization TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_credit_summary_teacher UNIQUE (teacher_id),
    CONSTRAINT chk_current_rank CHECK (current_rank IN (
        'guru_pertama', 'guru_muda', 'guru_madya', 'guru_utama', 'non_pns'
    )),
    CONSTRAINT chk_current_grade CHECK (current_grade IN (
        'III/a', 'III/b', 'III/c', 'III/d',
        'IV/a', 'IV/b', 'IV/c', 'IV/d', 'IV/e',
        'n/a'
    ))
);

-- Indexes
CREATE INDEX idx_credit_sum_tenant_company ON teacher_credit_summaries (tenant_id, company_id);
CREATE INDEX idx_credit_sum_teacher ON teacher_credit_summaries (teacher_id);
CREATE INDEX idx_credit_sum_rank ON teacher_credit_summaries (current_rank);
CREATE INDEX idx_credit_sum_rels ON teacher_credit_summaries USING GIN (_rels);
CREATE INDEX idx_credit_sum_data ON teacher_credit_summaries USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `certification_type` | 7 tipe | Mencakup sertifikasi pendidik (tunjangan), kompetensi, keahlian, ijazah, akta, pelatihan |
| `expiry_date` | DATE, nullable | Beberapa sertifikasi perlu diperpanjang — null = tidak kadaluarsa |
| `activity_type` | 14 tipe | Mencakup semua jenis kegiatan PKB sesuai regulasi |
| `pkb_component` | 3 komponen | Sesuai Permenneg PAN & RB: pengembangan_diri, publikasi_ilmiah, karya_inovatif |
| `credit_points` | NUMERIC(6,2) | Angka kredit per kegiatan — ditetapkan oleh tim penilai |
| `credit_approved` | BOOLEAN | Angka kredit harus disetujui oleh atasan/tim penilai |
| `total_hours` | NUMERIC(6,1) | Jumlah jam kegiatan — untuk laporan Dapodik |
| `current_rank` | VARCHAR(30) | Jabatan fungsional guru: pertama, muda, madya, utama, atau non_pns |
| `current_grade` | VARCHAR(10) | Golongan PNS: III/a s/d IV/e. Non-PNS = 'n/a' |
| `credit_gap` | NUMERIC(8,2) | Selisih kredit yang masih dibutuhkan untuk kenaikan pangkat |
| `skp_year` | INT | Tahun SKP — sasaran kinerja pegawai tahunan |

### Vernon Relationships

**teacher_certifications:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `teacher` | belongs_to | **Ya** | Pemilik sertifikasi |

**teacher_development_activities:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `teacher` | belongs_to | **Ya** | Peserta kegiatan |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran (jika applicable) |

**teacher_credit_summaries:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `teacher` | belongs_to | **Ya** | Pemilik ringkasan kredit |

### _rels / _data Structure

```json
// teacher_certifications
{
  "_rels": {
    "teacher_id": "018f..."
  },
  "_data": {
    "teacher": {
      "id": "018f...",
      "full_name": "Bu Siti Aminah",
      "nip": "198501012010012001",
      "employee_type": "pns"
    }
  }
}

// teacher_development_activities
{
  "_rels": {
    "teacher_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "teacher": {
      "id": "018f...",
      "full_name": "Bu Siti Aminah",
      "nip": "198501012010012001"
    },
    "academic_year": {
      "id": "018f...",
      "name": "2025/2026"
    }
  }
}

// teacher_credit_summaries
{
  "_rels": {
    "teacher_id": "018f..."
  },
  "_data": {
    "teacher": {
      "id": "018f...",
      "full_name": "Bu Siti Aminah",
      "nip": "198501012010012001",
      "employee_type": "pns",
      "role": "guru_mapel"
    }
  }
}
```

### API Endpoints

```
# Certifications
GET    /api/v1/teachers/{id}/certifications              — Sertifikasi guru
POST   /api/v1/teacher-certifications                    — Tambah sertifikasi
PUT    /api/v1/teacher-certifications/{id}               — Update sertifikasi
GET    /api/v1/teacher-certifications/expiring           — Sertifikasi yang akan kadaluarsa

# Development Activities (PKB)
GET    /api/v1/teachers/{id}/development-activities      — Riwayat PKB guru
POST   /api/v1/teacher-development-activities            — Tambah kegiatan PKB
PUT    /api/v1/teacher-development-activities/{id}       — Update kegiatan
GET    /api/v1/teacher-development-activities             — List semua kegiatan (admin view)
POST   /api/v1/teacher-development-activities/{id}/approve-credit — Approve angka kredit

# Credit Summary
GET    /api/v1/teachers/{id}/credit-summary              — Ringkasan angka kredit guru
PUT    /api/v1/teacher-credit-summaries/{id}             — Update ringkasan (recalculate)
GET    /api/v1/teacher-credit-summaries                  — Rekap kredit semua guru
POST   /api/v1/teacher-credit-summaries/recalculate      — Recalculate dari activities

# Reports
GET    /api/v1/teacher-development-activities/dapodik-report — Laporan PKB untuk Dapodik
GET    /api/v1/teacher-credit-summaries/rank-readiness     — Guru yang siap naik pangkat
```

### Credit Calculation

```sql
-- Recalculate credit summary dari approved activities
SELECT
    teacher_id,
    SUM(credit_points) FILTER (WHERE pkb_component = 'pengembangan_diri') AS credit_pengembangan_diri,
    SUM(credit_points) FILTER (WHERE pkb_component = 'publikasi_ilmiah') AS credit_publikasi,
    SUM(credit_points) FILTER (WHERE pkb_component = 'karya_inovatif') AS credit_karya_inovatif,
    SUM(credit_points) AS credit_total
FROM teacher_development_activities
WHERE credit_approved = true AND status = 'completed'
  AND tenant_id = $1 AND company_id = $2
GROUP BY teacher_id;
```

## Consequences

### Positive

- **Regulasi compliant**: 3 komponen PKB sesuai Permenneg PAN & RB No. 16/2009.
- **Angka kredit tracking**: Otomatis menghitung total kredit dari kegiatan yang diapprove.
- **Sertifikasi alert**: `expiry_date` memungkinkan notifikasi sertifikasi yang akan kadaluarsa.
- **SKP ready**: Ringkasan kredit bisa digunakan untuk penyusunan SKP tahunan.
- **PKG integration**: Hasil PKG (S028) dengan area "kurang" menjadi prioritas kegiatan PKB.
- **Dapodik ready**: Riwayat diklat dan sertifikasi bisa diekspor untuk pelaporan.

### Negative / Trade-offs

- **3 tabel**: Sertifikasi, kegiatan, dan ringkasan terpisah — complexity tinggi tapi separation of concerns baik.
- **Manual credit input**: Angka kredit per kegiatan diinput manual — belum ada auto-calculate berdasarkan jenis kegiatan.
- **PNS-centric**: Angka kredit dan golongan utamanya untuk PNS. Guru honorer/yayasan menggunakan `rank = 'non_pns'` dan `grade = 'n/a'`.
- **Credit approval**: Perlu tim penilai angka kredit — workflow belum detail di MVP.
- **Dokumen upload**: `certificate_url` dan `report_url` mengasumsikan ada storage service terpisah.

## Alternatives Considered

### 1. PKB sebagai extension dari PKG (S028)
- Ditolak: lifecycle berbeda — PKG per semester, PKB ongoing sepanjang karir. Domain yang berbeda.

### 2. Semua data dalam satu tabel
- Ditolak: sertifikasi, kegiatan, dan ringkasan kredit punya lifecycle dan granularity berbeda.

### 3. Tanpa credit summary (hitung real-time)
- Ditolak: query kredit total melibatkan aggregation banyak rows — snapshot summary lebih efisien untuk dashboard dan reporting.

### 4. External HR system integration
- Deferred: MVP standalone. Integrasi dengan SIMPEG/SAPK (sistem kepegawaian PNS) bisa ditambah via API gateway.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `CertificationDescriptor.TableName()` | -- | `"teacher_certifications"` |
| U02 | `ActivityDescriptor.TableName()` | -- | `"teacher_development_activities"` |
| U03 | `CreditSummaryDescriptor.TableName()` | -- | `"teacher_credit_summaries"` |
| U04 | Validate rejects invalid `certification_type` | `"diploma"` | Error |
| U05 | Validate rejects invalid `activity_type` | `"vacation"` | Error |
| U06 | Validate rejects invalid `pkb_component` | `"mengajar"` | Error |
| U07 | Validate rejects negative `credit_points` | `-5` | Error |
| U08 | Validate rejects end_date < start_date | start=2026-05-01, end=2026-04-01 | Error |
| U09 | Validate rejects invalid `current_rank` | `"professor"` | Error |
| U10 | Validate rejects invalid `current_grade` | `"V/a"` | Error |
| U11 | Calculate credit_gap | total=180, target=200 | gap=20 |
| U12 | Validate accepts valid activity | All fields valid | No error |

### Integration Tests — Certifications

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create certification | POST with valid data | 201 |
| I02 | Get teacher certifications | GET /teachers/{id}/certifications | 200, ordered by issue_date desc |
| I03 | Get expiring certifications | GET /certifications/expiring?days=90 | 200, only certs expiring within 90 days |
| I04 | Update certification status | PUT with is_active=false | 200 |

### Integration Tests — Development Activities

| # | Test Case | Action | Expected |
|---|---|---|---|
| I05 | Create PKB activity | POST with valid data | 201, status=planned |
| I06 | Complete activity | PUT with status=completed | 200 |
| I07 | Approve credit | POST /approve-credit | 200, credit_approved=true |
| I08 | Filter by pkb_component | GET ?pkb_component=pengembangan_diri | 200, filtered |
| I09 | Date range CHECK | INSERT with end < start | DB error |
| I10 | Credit points CHECK | INSERT with credit_points=-1 | DB error |

### Integration Tests — Credit Summary

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Recalculate credits | POST /recalculate | 200, totals match sum of approved activities |
| I12 | Credit gap calculated | target=200, total=180 | credit_gap=20 |
| I13 | Rank readiness report | GET /rank-readiness | 200, teachers where credit_gap <= 0 |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | TeacherUpdated syncs to certifications | Update teacher full_name | `_data.teacher.full_name` updated |
| I15 | TeacherUpdated syncs to activities | Update teacher full_name | `_data.teacher.full_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Cannot access other tenant's certifications | GET with wrong tenant | 404 |
| I17 | Cannot create activity for other company | POST with wrong company | Error, scope violation |
