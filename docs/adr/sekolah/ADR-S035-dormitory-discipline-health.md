# ADR-S035: Dormitory Discipline & Health / Disiplin & Kesehatan Asrama

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Disiplin asrama **berbeda dari disiplin sekolah** (ADR-S012). S012 mencakup pelanggaran akademik dan perilaku di lingkungan sekolah (terlambat masuk kelas, tidak mengerjakan PR, dll), sedangkan disiplin asrama mencakup:

1. **Pelanggaran asrama**: Tidak sholat berjamaah, keluar asrama tanpa izin, melanggar jam malam, kebersihan kamar.
2. **Inspeksi kamar**: Pemeriksaan rutin kebersihan dan kerapihan kamar oleh musyrif.
3. **Scoring kebersihan**: Penilaian kebersihan kamar per minggu/bulan.
4. **Monitoring kesehatan**: Pencatatan kesehatan santri yang tinggal 24 jam di asrama.

Konteks pesantren Indonesia:
- Pelanggaran asrama punya **skala dan sanksi berbeda** dari pelanggaran sekolah. Contoh: tidak sholat Subuh = pelanggaran berat di asrama, tapi bukan urusan disiplin sekolah.
- Inspeksi kamar adalah ritual mingguan di hampir semua pesantren.
- Santri yang sakit perlu penanganan di UKS asrama (bukan hanya UKS sekolah) — terutama di malam hari.
- Beberapa pesantren memiliki **kompetisi kamar terbersih** per bulan sebagai motivasi.
- Wabah penyakit menular (cacar, flu) perlu monitoring ketat karena santri tinggal bersama.

### Mengapa Vernon Pattern?

- Has_many dari student (banyak records per santri per semester).
- Read-heavy: dashboard musyrif, laporan bulanan ke orang tua, rekap kebersihan.
- Relasi ke student, dormitory_room (S033), academic_year.
- Business logic moderate: poin akumulasi, threshold eskalasi, scoring rata-rata.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `dormitory_violations` (pelanggaran asrama), `dormitory_inspections` (inspeksi kamar), dan `dormitory_health_records` (catatan kesehatan santri).

### Table Schema

```sql
-- Pelanggaran asrama
CREATE TABLE dormitory_violations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    room_id         UUID,

    -- Kejadian
    violation_date  DATE NOT NULL,
    semester        VARCHAR(10) NOT NULL,
    category        VARCHAR(30) NOT NULL,
    severity        VARCHAR(20) NOT NULL,
    description     TEXT NOT NULL,
    points          INT NOT NULL,

    -- Sanksi
    sanction        VARCHAR(30),
    sanction_note   TEXT,

    -- Pencatat
    reported_by     UUID NOT NULL,
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,

    -- Pemberitahuan
    parent_notified BOOLEAN NOT NULL DEFAULT false,
    notified_at     TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_dorm_viol_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_dorm_viol_category CHECK (category IN (
        'sholat', 'kebersihan', 'jam_malam', 'keluar_tanpa_izin',
        'kekerasan', 'pencurian', 'gadget', 'merokok', 'pacaran', 'other'
    )),
    CONSTRAINT chk_dorm_viol_severity CHECK (severity IN ('ringan', 'sedang', 'berat', 'sangat_berat')),
    CONSTRAINT chk_dorm_viol_sanction CHECK (sanction IS NULL OR sanction IN (
        'teguran_lisan', 'teguran_tertulis', 'hafalan_tambahan',
        'piket_tambahan', 'panggilan_ortu', 'skorsing_asrama',
        'dikeluarkan_asrama', 'dikeluarkan_pesantren'
    ))
);

-- Indexes
CREATE INDEX idx_dorm_viol_tenant_company ON dormitory_violations (tenant_id, company_id);
CREATE INDEX idx_dorm_viol_student ON dormitory_violations (student_id);
CREATE INDEX idx_dorm_viol_student_semester ON dormitory_violations (student_id, academic_year_id, semester);
CREATE INDEX idx_dorm_viol_date ON dormitory_violations (violation_date);
CREATE INDEX idx_dorm_viol_category ON dormitory_violations (category);
CREATE INDEX idx_dorm_viol_severity ON dormitory_violations (severity);
CREATE INDEX idx_dorm_viol_rels ON dormitory_violations USING GIN (_rels);
CREATE INDEX idx_dorm_viol_data ON dormitory_violations USING GIN (_data);

-- Inspeksi kamar
CREATE TABLE dormitory_inspections (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    room_id         UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Inspeksi
    inspection_date DATE NOT NULL,
    semester        VARCHAR(10) NOT NULL,

    -- Scoring (skala 1-100)
    cleanliness_score   INT NOT NULL,
    tidiness_score      INT NOT NULL,
    completeness_score  INT NOT NULL,
    overall_score       INT NOT NULL,
    grade               VARCHAR(2) NOT NULL,

    -- Detail
    notes           TEXT,
    issues_found    TEXT,

    -- Inspektor
    inspected_by    UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_dorm_inspection_room_date UNIQUE (room_id, inspection_date),
    CONSTRAINT chk_dorm_insp_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_dorm_insp_cleanliness CHECK (cleanliness_score BETWEEN 0 AND 100),
    CONSTRAINT chk_dorm_insp_tidiness CHECK (tidiness_score BETWEEN 0 AND 100),
    CONSTRAINT chk_dorm_insp_completeness CHECK (completeness_score BETWEEN 0 AND 100),
    CONSTRAINT chk_dorm_insp_overall CHECK (overall_score BETWEEN 0 AND 100),
    CONSTRAINT chk_dorm_insp_grade CHECK (grade IN ('A', 'B', 'C', 'D', 'E'))
);

-- Indexes
CREATE INDEX idx_dorm_insp_tenant_company ON dormitory_inspections (tenant_id, company_id);
CREATE INDEX idx_dorm_insp_room ON dormitory_inspections (room_id);
CREATE INDEX idx_dorm_insp_date ON dormitory_inspections (inspection_date);
CREATE INDEX idx_dorm_insp_grade ON dormitory_inspections (grade);
CREATE INDEX idx_dorm_insp_rels ON dormitory_inspections USING GIN (_rels);
CREATE INDEX idx_dorm_insp_data ON dormitory_inspections USING GIN (_data);

-- Catatan kesehatan santri di asrama
CREATE TABLE dormitory_health_records (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Kejadian
    record_date     DATE NOT NULL,
    record_type     VARCHAR(20) NOT NULL,
    complaint       TEXT NOT NULL,
    diagnosis       TEXT,
    treatment       TEXT,

    -- Severity
    severity        VARCHAR(20) NOT NULL DEFAULT 'minor',

    -- Tindak lanjut
    referred_to     VARCHAR(30),
    referred_at     TIMESTAMPTZ,
    recovery_date   DATE,

    -- Pencatat
    recorded_by     UUID NOT NULL,
    parent_notified BOOLEAN NOT NULL DEFAULT false,
    notified_at     TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_dorm_health_type CHECK (record_type IN (
        'illness', 'injury', 'allergy', 'infectious', 'mental_health', 'routine_check'
    )),
    CONSTRAINT chk_dorm_health_severity CHECK (severity IN ('minor', 'moderate', 'serious', 'emergency')),
    CONSTRAINT chk_dorm_health_referred CHECK (referred_to IS NULL OR referred_to IN (
        'uks', 'puskesmas', 'klinik', 'rumah_sakit', 'orang_tua'
    ))
);

-- Indexes
CREATE INDEX idx_dorm_health_tenant_company ON dormitory_health_records (tenant_id, company_id);
CREATE INDEX idx_dorm_health_student ON dormitory_health_records (student_id);
CREATE INDEX idx_dorm_health_date ON dormitory_health_records (record_date);
CREATE INDEX idx_dorm_health_type ON dormitory_health_records (record_type);
CREATE INDEX idx_dorm_health_severity ON dormitory_health_records (severity) WHERE severity IN ('serious', 'emergency');
CREATE INDEX idx_dorm_health_infectious ON dormitory_health_records (record_type, record_date) WHERE record_type = 'infectious';
CREATE INDEX idx_dorm_health_rels ON dormitory_health_records USING GIN (_rels);
CREATE INDEX idx_dorm_health_data ON dormitory_health_records USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `category` (violation) | 10 kategori | Covers pelanggaran khas pesantren: sholat, kebersihan, jam malam, gadget, merokok, pacaran |
| `sanction` | 8 jenis | Sanksi pesantren: hafalan tambahan dan piket tambahan adalah sanksi khas — berbeda dari S012 |
| `severity` (violation) | 4 level | Konsisten dengan S012 tapi independen — threshold berbeda |
| `cleanliness/tidiness/completeness_score` | 3 sub-score 0-100 | Tiga dimensi inspeksi: kebersihan, kerapihan, kelengkapan |
| `overall_score` | INT computed | Rata-rata atau weighted average dari 3 sub-score — stored untuk query cepat |
| `grade` | A-E | Mapping dari overall_score: A (80-100), B (60-79), C (40-59), D (20-39), E (0-19) |
| `record_type` (health) | 6 tipe | Termasuk `infectious` — kritis untuk asrama karena risiko penularan |
| `referred_to` | 5 opsi | Eskalasi kesehatan: UKS → Puskesmas → Klinik → RS → Pulang ke orang tua |

### Escalation Thresholds (Application Logic — Dormitory)

```
Pelanggaran ringan    → 5 poin   → teguran_lisan
Pelanggaran sedang    → 15 poin  → teguran_tertulis + hafalan_tambahan
Pelanggaran berat     → 30 poin  → panggilan_ortu + piket_tambahan
Akumulasi 50 poin     → skorsing_asrama (1 minggu di rumah)
Akumulasi 75 poin     → dikeluarkan_asrama (tetap sekolah, tidak boarding)
Akumulasi 100 poin    → dikeluarkan_pesantren (perlu approval pimpinan pondok)
```

Threshold ini dikonfigurasi per pesantren, independen dari threshold S012.

### Vernon Relationships

**dormitory_violations:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Nama santri |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |
| `room` | belongs_to | **Tidak** | Opsional — hanya jika pelanggaran terkait kamar |

**dormitory_inspections:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `room` | belongs_to | **Ya** | Nomor kamar + gedung selalu ditampilkan |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

**dormitory_health_records:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Nama santri |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

### _rels / _data Structure

```json
// dormitory_violations
{
  "_rels": {
    "student_id": "018f...",
    "academic_year_id": "018f...",
    "room_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad Fauzi", "nis": "12345" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}

// dormitory_inspections
{
  "_rels": {
    "room_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "room": {
      "id": "018f...",
      "room_number": "201",
      "building": { "id": "018f...", "name": "Gedung Al-Farabi" }
    },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}

// dormitory_health_records
{
  "_rels": {
    "student_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad Fauzi", "nis": "12345" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

### API Endpoints

```
# Violations
GET    /api/v1/students/{id}/dormitory-violations       — Riwayat pelanggaran asrama santri
GET    /api/v1/students/{id}/dormitory-violations/summary — Rekap poin per semester
POST   /api/v1/dormitory-violations                      — Catat pelanggaran
PUT    /api/v1/dormitory-violations/{id}                 — Update pelanggaran
POST   /api/v1/dormitory-violations/{id}/approve          — Approve sanksi
POST   /api/v1/dormitory-violations/{id}/notify-parent    — Tandai ortu sudah diberitahu

# Inspections
GET    /api/v1/dormitory-rooms/{id}/inspections          — Riwayat inspeksi kamar
POST   /api/v1/dormitory-inspections                     — Catat inspeksi
POST   /api/v1/dormitory-inspections/bulk                — Bulk inspeksi per gedung/lantai
GET    /api/v1/dormitory-inspections/ranking              — Ranking kamar terbersih
GET    /api/v1/dormitory-inspections/summary              — Rekap inspeksi per gedung

# Health Records
GET    /api/v1/students/{id}/dormitory-health-records    — Catatan kesehatan santri
POST   /api/v1/dormitory-health-records                  — Catat kejadian kesehatan
PUT    /api/v1/dormitory-health-records/{id}             — Update catatan
POST   /api/v1/dormitory-health-records/{id}/notify-parent — Notifikasi orang tua
GET    /api/v1/dormitory-health-records/infectious        — Monitoring penyakit menular
```

## Consequences

### Positive

- **Pesantren-specific**: Kategori pelanggaran dan sanksi khas pesantren (hafalan tambahan, piket).
- **Terpisah dari S012**: Domain disiplin asrama independen — threshold dan kategori berbeda.
- **Inspeksi terstruktur**: 3 sub-score memberikan feedback spesifik per dimensi kebersihan.
- **Health monitoring**: Tracking kesehatan kritis untuk boarding students yang tinggal 24 jam.
- **Infectious disease tracking**: Index khusus untuk monitoring penyakit menular — early warning system.
- **Ranking kompetisi**: Data inspeksi bisa diolah untuk kompetisi kamar terbersih.

### Negative / Trade-offs

- **3 tabel sekaligus**: Complexity lebih tinggi, tapi masing-masing domain cukup distinct.
- **Health bukan EMR**: Ini bukan electronic medical record — hanya pencatatan dasar untuk monitoring. Kasus serius tetap di-refer ke fasilitas kesehatan.
- **Score subjectivity**: Inspeksi scoring bersifat subjektif — perlu SOP inspeksi yang jelas di level kebijakan.
- **Threshold configurable**: Sama seperti S012, threshold eskalasi perlu config terpisah.

## Alternatives Considered

### 1. Gabungkan dengan S012 (student_disciplines)
- Ditolak: kategori, sanksi, dan threshold sangat berbeda. Satu santri bisa punya poin disiplin sekolah dan poin disiplin asrama secara terpisah.

### 2. Inspeksi sebagai JSONB array di dormitory_rooms
- Ditolak: perlu query histori, ranking, dan aggregate — row-based lebih efisien.

### 3. Health records di S005 (student health)
- Ditolak: S005 adalah data statis (riwayat penyakit, alergi, golongan darah). Dormitory health records adalah event-based daily monitoring.

### 4. Single score untuk inspeksi (tanpa sub-scores)
- Ditolak: sub-scores memberikan granularity untuk feedback dan perbaikan spesifik.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `ViolationDescriptor.TableName()` | — | `"dormitory_violations"` |
| U02 | `InspectionDescriptor.TableName()` | — | `"dormitory_inspections"` |
| U03 | `HealthRecordDescriptor.TableName()` | — | `"dormitory_health_records"` |
| U04 | Validate rejects invalid `category` | `"fighting"` | Error: invalid category |
| U05 | Validate rejects invalid `severity` | `"critical"` | Error: must be ringan/sedang/berat/sangat_berat |
| U06 | Validate rejects score out of range | `cleanliness_score = 105` | Error: must be 0-100 |
| U07 | Validate rejects invalid `grade` | `"F"` | Error: must be A/B/C/D/E |
| U08 | Validate rejects invalid health `record_type` | `"surgery"` | Error: invalid type |
| U09 | Validate accepts valid inspection | All fields valid | No error |

### Integration Tests — Violations

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Record violation | POST with valid data | 201, created with `_data` |
| I02 | Get student violation history | GET /students/{id}/dormitory-violations | 200, ordered by date |
| I03 | Get violation summary | GET /students/{id}/dormitory-violations/summary | 200, total points per semester |
| I04 | Approve sanction | POST /approve | 200, approved_by + approved_at set |
| I05 | Category CHECK enforced | INSERT with `category = 'fighting'` | DB error |
| I06 | Severity CHECK enforced | INSERT with `severity = 'critical'` | DB error |

### Integration Tests — Inspections

| # | Test Case | Action | Expected |
|---|---|---|---|
| I07 | Record inspection | POST with scores | 201, created |
| I08 | Unique per room per date | Inspect same room same date | 409/422 |
| I09 | Score range CHECK | INSERT with score 105 | DB error |
| I10 | Get room inspection history | GET /dormitory-rooms/{id}/inspections | 200, ordered by date |
| I11 | Get ranking | GET /inspections/ranking | 200, sorted by overall_score desc |
| I12 | Bulk inspection | POST /bulk for 10 rooms | 201, 10 records created |

### Integration Tests — Health Records

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | Record health event | POST with complaint + type | 201, created |
| I14 | Get student health records | GET /students/{id}/dormitory-health-records | 200, ordered by date |
| I15 | Get infectious monitoring | GET /infectious | 200, filtered infectious records |
| I16 | Record type CHECK | INSERT with `record_type = 'surgery'` | DB error |
| I17 | Severity CHECK | INSERT with `severity = 'critical'` | DB error |
| I18 | Notify parent | POST /notify-parent | 200, parent_notified = true |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I19 | StudentUpdated syncs to violations | Update student name | `_data.student.full_name` updated |
| I20 | RoomUpdated syncs to inspections | Update room number | `_data.room.room_number` updated |
| I21 | StudentUpdated syncs to health records | Update student name | `_data.student.full_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I22 | Cannot access other tenant's violations | GET with wrong tenant | 404 |
| I23 | Cannot inspect other tenant's rooms | POST inspection cross-tenant | Error |
