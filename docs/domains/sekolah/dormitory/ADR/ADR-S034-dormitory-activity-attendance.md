# ADR-S034: Dormitory Activity & Attendance / Aktivitas & Kehadiran Asrama

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Kehadiran asrama **berbeda dari kehadiran sekolah** (ADR-S008). S008 mencatat absensi kelas harian (present/sick/permitted/absent), sedangkan kehadiran asrama mencakup:

1. **Absensi aktivitas harian**: Sholat berjamaah (Subuh, Dzuhur, Ashar, Maghrib, Isya), makan (3x), belajar malam (mudzakaroh), apel pagi/malam.
2. **Check-in/curfew**: Tracking santri masuk asrama setelah jam malam.
3. **Izin pulang**: Weekend permission atau izin khusus (sakit, acara keluarga).
4. **Jadwal aktivitas**: Template aktivitas harian yang berlaku per gedung/asrama.

Konteks pesantren Indonesia:
- Santri tinggal 24 jam di pesantren — aktivitas terjadwal dari Subuh (04:30) hingga tidur (22:00).
- Kehadiran sholat berjamaah adalah **indikator utama** kedisiplinan santri.
- Izin pulang (boyongan) perlu persetujuan musyrif dan dikonfirmasi orang tua.
- Beberapa pesantren melarang pulang kecuali hari libur besar (Idul Fitri, Idul Adha).
- Musyrif/Musyrifah melakukan absensi per aktivitas, bukan hanya sekali per hari.

### Mengapa Vernon Pattern?

- Volume sangat tinggi: 5 sholat + 3 makan + 2 belajar = ~10 aktivitas/hari × 500 santri = 5.000 records/hari.
- Read-heavy untuk rekap dan laporan bulanan ke orang tua.
- Relasi ke student, dormitory_room (S033), academic_year.
- Eventually consistent acceptable — rekap bisa async.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `dormitory_activities` (master jadwal aktivitas), `dormitory_attendances` (absensi per aktivitas), dan `dormitory_permissions` (izin pulang/keluar).

### Table Schema

```sql
-- Master aktivitas asrama
CREATE TABLE dormitory_activities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    description     TEXT,

    -- Konfigurasi
    activity_type   VARCHAR(20) NOT NULL,
    time_start      TIME NOT NULL,
    time_end        TIME NOT NULL,
    is_mandatory    BOOLEAN NOT NULL DEFAULT true,
    applies_to      VARCHAR(10) NOT NULL DEFAULT 'all',

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_dorm_activity_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_dorm_activity_type CHECK (activity_type IN (
        'sholat', 'meal', 'study', 'roll_call', 'cleaning', 'sports', 'other'
    )),
    CONSTRAINT chk_dorm_applies_to CHECK (applies_to IN ('all', 'male', 'female'))
);

-- Indexes
CREATE INDEX idx_dorm_activity_tenant_company ON dormitory_activities (tenant_id, company_id);
CREATE INDEX idx_dorm_activity_type ON dormitory_activities (activity_type);

-- Absensi per aktivitas per santri
CREATE TABLE dormitory_attendances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    activity_id     UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Absensi
    attendance_date DATE NOT NULL,
    status          VARCHAR(20) NOT NULL,
    check_in_time   TIME,
    note            TEXT,

    -- Pencatat
    recorded_by     UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_dorm_attendance_unique UNIQUE (student_id, activity_id, attendance_date),
    CONSTRAINT chk_dorm_att_status CHECK (status IN ('present', 'late', 'absent', 'permitted', 'sick'))
);

-- Indexes
CREATE INDEX idx_dorm_att_tenant_company ON dormitory_attendances (tenant_id, company_id);
CREATE INDEX idx_dorm_att_student ON dormitory_attendances (student_id);
CREATE INDEX idx_dorm_att_activity ON dormitory_attendances (activity_id);
CREATE INDEX idx_dorm_att_date ON dormitory_attendances (attendance_date);
CREATE INDEX idx_dorm_att_student_date ON dormitory_attendances (student_id, attendance_date);
CREATE INDEX idx_dorm_att_status ON dormitory_attendances (status) WHERE status != 'present';
CREATE INDEX idx_dorm_att_rels ON dormitory_attendances USING GIN (_rels);
CREATE INDEX idx_dorm_att_data ON dormitory_attendances USING GIN (_data);

-- Izin pulang / keluar asrama
CREATE TABLE dormitory_permissions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Detail izin
    permission_type VARCHAR(20) NOT NULL,
    reason          TEXT NOT NULL,
    depart_date     DATE NOT NULL,
    depart_time     TIME,
    expected_return_date DATE NOT NULL,
    expected_return_time TIME,
    actual_return_date DATE,
    actual_return_time TIME,

    -- Status & approval
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,
    picked_up_by    VARCHAR(200),

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_dorm_perm_type CHECK (permission_type IN (
        'weekend', 'holiday', 'sick', 'family_event', 'emergency', 'other'
    )),
    CONSTRAINT chk_dorm_perm_status CHECK (status IN (
        'pending', 'approved', 'rejected', 'departed', 'returned', 'overdue'
    ))
);

-- Indexes
CREATE INDEX idx_dorm_perm_tenant_company ON dormitory_permissions (tenant_id, company_id);
CREATE INDEX idx_dorm_perm_student ON dormitory_permissions (student_id);
CREATE INDEX idx_dorm_perm_status ON dormitory_permissions (status) WHERE status IN ('pending', 'approved', 'departed', 'overdue');
CREATE INDEX idx_dorm_perm_depart ON dormitory_permissions (depart_date);
CREATE INDEX idx_dorm_perm_return ON dormitory_permissions (expected_return_date) WHERE actual_return_date IS NULL;
CREATE INDEX idx_dorm_perm_rels ON dormitory_permissions USING GIN (_rels);
CREATE INDEX idx_dorm_perm_data ON dormitory_permissions USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `activity_type` | 7 tipe | Covers semua aktivitas pesantren: sholat, makan, belajar, apel, kebersihan, olahraga |
| `applies_to` | all/male/female | Beberapa aktivitas gender-specific (misal: jadwal olahraga terpisah) |
| `is_mandatory` | BOOLEAN | Sholat wajib berjamaah vs aktivitas opsional |
| `status` (attendance) | 5 status | Tambah `late` dibanding S008 — keterlambatan sholat berjamaah sangat penting di pesantren |
| `check_in_time` | TIME, nullable | Waktu aktual check-in — untuk tracking keterlambatan |
| `permission_type` | 6 tipe | Covers izin pesantren: weekend, libur besar, sakit, acara keluarga, darurat |
| `picked_up_by` | VARCHAR(200) | Nama penjemput (orang tua/wali) — pesantren biasanya mencatat siapa yang menjemput |
| `actual_return_date/time` | nullable | Dicatat saat santri kembali — jika melewati expected = overdue |

### Vernon Relationships

**dormitory_attendances:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Nama santri selalu ditampilkan |
| `activity` | belongs_to | **Ya** | Nama dan tipe aktivitas |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

**dormitory_permissions:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Nama santri |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

### _rels / _data Structure

```json
// dormitory_attendances
{
  "_rels": {
    "student_id": "018f...",
    "activity_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad Fauzi", "nis": "12345" },
    "activity": { "id": "018f...", "name": "Sholat Subuh Berjamaah", "code": "SH-SUBUH", "activity_type": "sholat" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}

// dormitory_permissions
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
# Activities (Master)
GET    /api/v1/dormitory-activities                    — List aktivitas asrama
POST   /api/v1/dormitory-activities                    — Buat aktivitas
PUT    /api/v1/dormitory-activities/{id}               — Update aktivitas

# Attendances
POST   /api/v1/dormitory-attendances/bulk              — Bulk absensi per aktivitas (1 request = semua santri)
GET    /api/v1/students/{id}/dormitory-attendances      — Riwayat absensi asrama santri
GET    /api/v1/students/{id}/dormitory-attendances/summary — Rekap absensi per aktivitas per bulan
GET    /api/v1/dormitory-activities/{id}/attendances/{date} — Absensi per aktivitas per tanggal

# Permissions
GET    /api/v1/dormitory-permissions                    — List izin (filter by status)
POST   /api/v1/dormitory-permissions                    — Ajukan izin pulang
PUT    /api/v1/dormitory-permissions/{id}               — Update izin
POST   /api/v1/dormitory-permissions/{id}/approve       — Approve izin (musyrif)
POST   /api/v1/dormitory-permissions/{id}/reject        — Reject izin
POST   /api/v1/dormitory-permissions/{id}/depart        — Catat keberangkatan
POST   /api/v1/dormitory-permissions/{id}/return        — Catat kepulangan
GET    /api/v1/dormitory-permissions/overdue            — Daftar santri belum kembali
```

### Bulk Attendance Payload

```json
{
  "activity_id": "018f...",
  "academic_year_id": "018f...",
  "attendance_date": "2026-04-15",
  "attendances": [
    { "student_id": "018f...", "status": "present", "check_in_time": "04:35" },
    { "student_id": "018f...", "status": "late", "check_in_time": "04:52", "note": "Terlambat bangun" },
    { "student_id": "018f...", "status": "absent" },
    { "student_id": "018f...", "status": "permitted", "note": "Izin sakit di UKS" }
  ]
}
```

## Consequences

### Positive

- **Pesantren-native**: Dirancang untuk ritme harian pesantren (sholat, makan, belajar).
- **Granular tracking**: Absensi per aktivitas memberikan gambaran lengkap kedisiplinan santri.
- **Permission workflow**: Izin pulang dengan approval chain dan tracking kepulangan.
- **Overdue detection**: Santri yang belum kembali setelah tanggal diharapkan bisa dideteksi.
- **Terpisah dari S008**: Tidak mengotori domain absensi kelas sekolah.
- **Bulk-friendly**: Endpoint bulk untuk efisiensi input musyrif.

### Negative / Trade-offs

- **Volume sangat tinggi**: ~10 aktivitas/hari × 500 santri × 300 hari = 1.5 juta rows/tahun. Perlu partitioning strategy di masa depan.
- **Multiple musyrif input**: Belum ada conflict resolution jika 2 musyrif input absensi untuk aktivitas yang sama.
- **Overdue cron**: Deteksi overdue perlu background job — bukan real-time.
- **Offline support**: Musyrif sering input dari HP di area pesantren tanpa sinyal stabil — perlu offline-first strategy (future).

## Alternatives Considered

### 1. Gabungkan dengan S008 (student_attendances)
- Ditolak: S008 dirancang untuk 1 record/hari (kelas), dormitory attendance bisa 10+ records/hari. Struktur dan use case sangat berbeda.

### 2. Absensi per kamar (bukan per aktivitas)
- Ditolak: santri bisa hadir di kamar tapi absent saat sholat. Per-aktivitas lebih akurat.

### 3. Permission sebagai field di attendance
- Ditolak: izin pulang punya lifecycle sendiri (pending → approved → departed → returned). Tabel terpisah lebih clean.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `ActivityDescriptor.TableName()` | — | `"dormitory_activities"` |
| U02 | `AttendanceDescriptor.TableName()` | — | `"dormitory_attendances"` |
| U03 | `PermissionDescriptor.TableName()` | — | `"dormitory_permissions"` |
| U04 | Validate rejects invalid `activity_type` | `"prayer"` | Error: must be sholat/meal/study/... |
| U05 | Validate rejects invalid attendance `status` | `"excused"` | Error: must be present/late/absent/permitted/sick |
| U06 | Validate rejects invalid `permission_type` | `"vacation"` | Error: invalid type |
| U07 | Validate rejects `expected_return_date < depart_date` | return before depart | Error: invalid date range |
| U08 | Validate accepts valid attendance | All fields valid | No error |

### Integration Tests — Activities

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create activity | POST with name + code + type + time | 201 |
| I02 | Unique code per company | Create 2 activities same code | 409/422 |
| I03 | Activity type CHECK | INSERT with `activity_type = 'prayer'` | DB error |

### Integration Tests — Attendances

| # | Test Case | Action | Expected |
|---|---|---|---|
| I04 | Bulk insert attendance | POST bulk for 30 students | 201, 30 records created |
| I05 | Duplicate attendance rejected | Same student + activity + date twice | 409/422 |
| I06 | Get student dormitory attendance | GET /students/{id}/dormitory-attendances | 200, ordered by date desc |
| I07 | Get attendance summary | GET /students/{id}/dormitory-attendances/summary | 200, counts per activity per status |
| I08 | Status CHECK enforced | INSERT with `status = 'excused'` | DB error |

### Integration Tests — Permissions

| # | Test Case | Action | Expected |
|---|---|---|---|
| I09 | Create permission request | POST with student + dates + reason | 201, status = pending |
| I10 | Approve permission | POST /approve | 200, status → approved |
| I11 | Reject permission | POST /reject | 200, status → rejected |
| I12 | Record departure | POST /depart with picked_up_by | 200, status → departed |
| I13 | Record return | POST /return with actual dates | 200, status → returned |
| I14 | Overdue detection | GET /overdue after expected_return_date | 200, lists unreturned students |
| I15 | Permission type CHECK | INSERT with `permission_type = 'vacation'` | DB error |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | StudentUpdated syncs to attendances | Update student name | `_data.student.full_name` updated |
| I17 | ActivityUpdated syncs to attendances | Update activity name | `_data.activity.name` updated |
| I18 | StudentUpdated syncs to permissions | Update student name | `_data.student.full_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I19 | Cannot access other tenant's attendances | GET with wrong tenant | 404 |
| I20 | Cannot create permission cross-tenant | POST with wrong tenant student | Error |
