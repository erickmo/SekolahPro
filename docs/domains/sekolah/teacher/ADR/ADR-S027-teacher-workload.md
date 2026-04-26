# ADR-S027: Teacher Workload (Beban Mengajar)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Beban mengajar guru adalah salah satu aspek paling diregulasi di sistem pendidikan Indonesia. Peraturan utama:

1. **PP No. 74 Tahun 2008**: Guru PNS wajib mengajar minimal **24 jam pelajaran per minggu** dan maksimal 40 jam pelajaran per minggu.
2. **Permendikbud No. 15 Tahun 2018**: Pemenuhan beban kerja guru meliputi: (a) merencanakan pembelajaran, (b) melaksanakan pembelajaran, (c) menilai hasil pembelajaran, (d) membimbing dan melatih peserta didik, (e) melaksanakan tugas tambahan.
3. **Tunjangan sertifikasi**: Guru bersertifikat hanya menerima tunjangan profesi jika memenuhi 24 jam/minggu.
4. **Dapodik**: Pelaporan beban mengajar per guru per semester ke Kemendikbud — menentukan tunjangan.

Beban mengajar terdiri dari:

| Komponen | Sumber | Jam Ekuivalen |
|----------|--------|---------------|
| Mengajar tatap muka | Jadwal pelajaran (S021) | 1 JP = 1 jam mengajar |
| Wali kelas | ADR-011 class_rooms | 2 JP ekuivalen |
| Pembina ekskul | S015 extracurricular | 2 JP ekuivalen |
| Tutor/pembimbing | S017 counseling (guru BK) | Sesuai SK |
| Panitia/koordinator | Surat tugas internal | Sesuai SK |
| Kepala sekolah | ADR-012 role | 18 JP (sebagai pengganti mengajar) |
| Wakil kepsek | ADR-012 role | 12 JP ekuivalen |

### Mengapa Vernon Pattern?

- Beban mengajar per guru per semester — has_many dari teacher.
- Calculated dari S021 (timetable) + tugas tambahan.
- Read-heavy: laporan Dapodik, dashboard kepsek, perhitungan tunjangan.
- Business logic moderate: hitung total JP, cek threshold 24 jam, generate laporan.
- Eventually consistent acceptable — dihitung periodically, bukan real-time.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `teacher_workloads` (ringkasan beban per guru per semester) dan `teacher_workload_items` (detail per komponen beban).

### Table Schema

```sql
-- Ringkasan beban mengajar per guru per semester
CREATE TABLE teacher_workloads (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Periode
    semester        VARCHAR(10) NOT NULL,

    -- Ringkasan (dalam jam pelajaran/JP)
    teaching_hours  NUMERIC(5,1) NOT NULL DEFAULT 0,
    additional_hours NUMERIC(5,1) NOT NULL DEFAULT 0,
    total_hours     NUMERIC(5,1) NOT NULL DEFAULT 0,

    -- Status pemenuhan
    minimum_required NUMERIC(5,1) NOT NULL DEFAULT 24,
    is_fulfilled    BOOLEAN NOT NULL DEFAULT false,
    fulfillment_status VARCHAR(20) NOT NULL DEFAULT 'kurang',

    -- Dapodik reporting
    dapodik_reported BOOLEAN NOT NULL DEFAULT false,
    dapodik_reported_at TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_workload_teacher_semester UNIQUE (teacher_id, academic_year_id, semester),
    CONSTRAINT chk_workload_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_fulfillment_status CHECK (fulfillment_status IN ('kurang', 'terpenuhi', 'lebih')),
    CONSTRAINT chk_teaching_hours CHECK (teaching_hours >= 0),
    CONSTRAINT chk_additional_hours CHECK (additional_hours >= 0),
    CONSTRAINT chk_total_hours CHECK (total_hours >= 0)
);

-- Indexes
CREATE INDEX idx_workload_tenant_company ON teacher_workloads (tenant_id, company_id);
CREATE INDEX idx_workload_teacher ON teacher_workloads (teacher_id);
CREATE INDEX idx_workload_year_semester ON teacher_workloads (academic_year_id, semester);
CREATE INDEX idx_workload_fulfilled ON teacher_workloads (is_fulfilled) WHERE is_fulfilled = false;
CREATE INDEX idx_workload_rels ON teacher_workloads USING GIN (_rels);
CREATE INDEX idx_workload_data ON teacher_workloads USING GIN (_data);

-- Detail komponen beban mengajar
CREATE TABLE teacher_workload_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    workload_id     UUID NOT NULL,
    teacher_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Komponen beban
    item_type       VARCHAR(30) NOT NULL,
    description     TEXT NOT NULL,
    hours_per_week  NUMERIC(5,1) NOT NULL,

    -- Reference (opsional — link ke sumber beban)
    reference_type  VARCHAR(30),
    reference_id    UUID,

    -- Dasar (surat tugas / SK)
    sk_number       VARCHAR(100),
    sk_date         DATE,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_item_type CHECK (item_type IN (
        'mengajar', 'wali_kelas', 'pembina_ekskul',
        'guru_bk', 'kepala_sekolah', 'wakil_kepsek',
        'koordinator', 'panitia', 'tugas_tambahan'
    )),
    CONSTRAINT chk_item_hours CHECK (hours_per_week >= 0 AND hours_per_week <= 40),
    CONSTRAINT chk_reference_type CHECK (
        reference_type IS NULL OR reference_type IN (
            'timetable_slot', 'class_room', 'extracurricular', 'sk_internal'
        )
    )
);

-- Indexes
CREATE INDEX idx_workload_item_tenant_company ON teacher_workload_items (tenant_id, company_id);
CREATE INDEX idx_workload_item_workload ON teacher_workload_items (workload_id);
CREATE INDEX idx_workload_item_teacher ON teacher_workload_items (teacher_id);
CREATE INDEX idx_workload_item_type ON teacher_workload_items (item_type);
CREATE INDEX idx_workload_item_rels ON teacher_workload_items USING GIN (_rels);
CREATE INDEX idx_workload_item_data ON teacher_workload_items USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `teaching_hours` | NUMERIC(5,1) | Jam mengajar tatap muka dari jadwal (S021) — bisa 0.5 JP |
| `additional_hours` | NUMERIC(5,1) | JP ekuivalen dari tugas tambahan (wali kelas, pembina, dll) |
| `total_hours` | NUMERIC(5,1) | `teaching_hours + additional_hours` — denormalisasi untuk query speed |
| `minimum_required` | NUMERIC(5,1), DEFAULT 24 | Standar 24 JP untuk guru PNS — configurable karena kepala sekolah berbeda (18 JP) |
| `is_fulfilled` | BOOLEAN | `total_hours >= minimum_required` — denormalisasi untuk filter cepat |
| `fulfillment_status` | VARCHAR(20) | kurang/terpenuhi/lebih — granularity lebih dari boolean |
| `dapodik_reported` | BOOLEAN | Flag apakah sudah dilaporkan ke Dapodik |
| `item_type` | VARCHAR(30), 9 tipe | Sesuai Permendikbud — setiap komponen punya JP ekuivalen berbeda |
| `reference_type` + `reference_id` | Polymorphic reference | Link ke sumber beban: slot jadwal, kelas, ekskul, atau SK internal |
| `sk_number` | VARCHAR(100), nullable | Nomor Surat Keputusan — bukti formal untuk tugas tambahan |

### Vernon Relationships

**teacher_workloads:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `teacher` | belongs_to | **Ya** | Nama guru, NIP, employee_type |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |

**teacher_workload_items:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `workload` | belongs_to | **Ya** | Parent workload |
| `teacher` | belongs_to | **Ya** | Nama guru |

### _rels / _data Structure

```json
// teacher_workloads
{
  "_rels": {
    "teacher_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "teacher": {
      "id": "018f...",
      "full_name": "Bu Siti Aminah",
      "nip": "198501012010012001",
      "employee_type": "pns",
      "role": "guru_mapel"
    },
    "academic_year": {
      "id": "018f...",
      "name": "2025/2026"
    }
  }
}

// teacher_workload_items
{
  "_rels": {
    "workload_id": "018f...",
    "teacher_id": "018f..."
  },
  "_data": {
    "workload": {
      "id": "018f...",
      "total_hours": 28,
      "is_fulfilled": true
    },
    "teacher": {
      "id": "018f...",
      "full_name": "Bu Siti Aminah"
    }
  }
}
```

### API Endpoints

```
# Workload summary
GET    /api/v1/teachers/{id}/workload                  — Beban mengajar guru per semester aktif
GET    /api/v1/teachers/{id}/workload/history          — Riwayat beban per semester
GET    /api/v1/teacher-workloads                       — Rekap beban seluruh guru
GET    /api/v1/teacher-workloads/unfulfilled           — Guru yang belum memenuhi 24 JP

# Workload items
GET    /api/v1/teacher-workloads/{id}/items            — Detail komponen beban guru
POST   /api/v1/teacher-workload-items                  — Tambah komponen beban (tugas tambahan)
PUT    /api/v1/teacher-workload-items/{id}             — Update komponen
DELETE /api/v1/teacher-workload-items/{id}             — Hapus komponen

# Calculation & Reporting
POST   /api/v1/teacher-workloads/calculate             — Recalculate dari jadwal + tugas tambahan
POST   /api/v1/teacher-workloads/dapodik-report        — Generate laporan Dapodik
```

### Calculation Logic

```sql
-- Step 1: Hitung jam mengajar dari timetable (S021)
SELECT
    teacher_id,
    SUM(hours_per_week) AS teaching_hours
FROM timetable_slots
WHERE academic_year_id = $1 AND semester = $2
GROUP BY teacher_id;

-- Step 2: Hitung tugas tambahan dari workload_items
SELECT
    teacher_id,
    SUM(hours_per_week) AS additional_hours
FROM teacher_workload_items
WHERE workload_id = $3 AND item_type != 'mengajar'
GROUP BY teacher_id;

-- Step 3: Update workload summary
UPDATE teacher_workloads SET
    teaching_hours = step1.teaching_hours,
    additional_hours = step2.additional_hours,
    total_hours = step1.teaching_hours + step2.additional_hours,
    is_fulfilled = (step1.teaching_hours + step2.additional_hours) >= minimum_required,
    fulfillment_status = CASE
        WHEN total < minimum_required THEN 'kurang'
        WHEN total = minimum_required THEN 'terpenuhi'
        ELSE 'lebih'
    END;
```

## Consequences

### Positive

- **Regulasi compliant**: Tracking 24 JP/minggu sesuai PP 74/2008 dan Permendikbud 15/2018.
- **Dapodik ready**: `dapodik_reported` flag dan laporan terstruktur.
- **Tunjangan sertifikasi**: Data workload menjadi dasar pencairan tunjangan profesi guru.
- **Transparan**: Guru dan kepsek bisa melihat breakdown beban per komponen.
- **Auto-calculate**: Jam mengajar dihitung otomatis dari jadwal (S021), tugas tambahan diinput manual.

### Negative / Trade-offs

- **Depends on S021**: Tanpa timetable, teaching_hours harus diinput manual.
- **Ekuivalen JP manual**: Konversi tugas tambahan ke JP (misal: wali kelas = 2 JP) bersifat configurable tapi perlu maintenance.
- **Recalculation**: Setiap perubahan jadwal harus trigger recalculation — bisa heavy jika jadwal sering berubah.
- **PNS-centric**: Aturan 24 JP utamanya untuk guru PNS/P3K — guru honorer/yayasan mungkin punya aturan berbeda.

## Alternatives Considered

### 1. Hitung real-time dari jadwal tanpa tabel workload
- Ditolak: terlalu banyak JOIN setiap query. Tugas tambahan tidak ada di jadwal. Perlu snapshot per semester untuk Dapodik.

### 2. Semua komponen dalam satu JSONB
- Ditolak: tidak bisa query "guru yang belum 24 JP", tidak bisa aggregate per item_type.

### 3. Workload tanpa item detail (summary only)
- Ditolak: kepsek dan Dapodik perlu breakdown — dari mana 24 JP itu terdiri.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `WorkloadDescriptor.TableName()` | -- | `"teacher_workloads"` |
| U02 | `WorkloadItemDescriptor.TableName()` | -- | `"teacher_workload_items"` |
| U03 | Validate rejects invalid `semester` | `"midterm"` | Error |
| U04 | Validate rejects invalid `item_type` | `"freelance"` | Error |
| U05 | Validate rejects negative `hours_per_week` | `-2` | Error |
| U06 | Validate rejects hours > 40 | `41` | Error |
| U07 | Calculate fulfillment: kurang | total=20, minimum=24 | `fulfillment_status = 'kurang'`, `is_fulfilled = false` |
| U08 | Calculate fulfillment: terpenuhi | total=24, minimum=24 | `fulfillment_status = 'terpenuhi'`, `is_fulfilled = true` |
| U09 | Calculate fulfillment: lebih | total=30, minimum=24 | `fulfillment_status = 'lebih'`, `is_fulfilled = true` |
| U10 | Kepsek minimum is 18 JP | teacher role=kepala_sekolah | `minimum_required = 18` |

### Integration Tests — Workload CRUD

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create workload for teacher | Auto-create on semester start | 201, workload with 0 hours |
| I02 | Unique per teacher+year+semester | Create duplicate | 409/422 |
| I03 | Add teaching item | POST item with type=mengajar | 201, teaching_hours updated |
| I04 | Add wali_kelas item | POST item with type=wali_kelas, hours=2 | 201, additional_hours += 2 |
| I05 | Delete item recalculates total | DELETE item | 200, total_hours decreased |
| I06 | Get teacher workload | GET /teachers/{id}/workload | 200, summary + items |
| I07 | Get unfulfilled teachers | GET /teacher-workloads/unfulfilled | 200, only teachers with is_fulfilled=false |

### Integration Tests — Calculation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I08 | Recalculate from timetable | POST /calculate after timetable change | teaching_hours matches timetable sum |
| I09 | Total = teaching + additional | teaching=20, additional=6 | total=26, is_fulfilled=true |
| I10 | Kepsek gets 18 JP minimum | role=kepala_sekolah | minimum_required=18 |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | TeacherUpdated syncs to workloads | Update teacher full_name | `_data.teacher.full_name` updated |
| I12 | AcademicYearUpdated syncs | Update year name | `_data.academic_year.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | Cannot access other tenant's workload | GET with wrong tenant | 404 |
| I14 | Cannot create item for other company | POST item with wrong company | Error, scope violation |
