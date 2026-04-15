# ADR-S026: Teacher Attendance (Absensi Guru & Staff)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Absensi guru dan staff berbeda fundamental dari absensi siswa (ADR-S008). Siswa diabsen oleh wali kelas per hari, sedangkan guru/staff melakukan **clock-in/clock-out** sendiri dan kehadiran mereka berdampak pada:

1. **Operasional sekolah**: Guru yang tidak hadir perlu digantikan — langsung berdampak pada jadwal pelajaran (S021) dan penggantian guru (S032).
2. **Penggajian**: Potongan gaji untuk ketidakhadiran tanpa izin, insentif kehadiran (S031).
3. **Beban mengajar**: Guru PNS wajib 24 jam/minggu — kehadiran menjadi bukti pemenuhan (S027).
4. **Penilaian kinerja**: Kehadiran adalah salah satu indikator PKG (S028).
5. **Dapodik / BKN**: Pelaporan kehadiran ASN ke Badan Kepegawaian Negara.
6. **Pesantren**: Absensi ustadz/ustadzah mengikuti pola yang sama, termasuk jadwal mengajar diniyah.

Perbedaan dengan S008 (Student Attendance):

| Aspek | S008 (Siswa) | S026 (Guru/Staff) |
|-------|-------------|-------------------|
| Cara input | Guru mengabsen siswa | Self clock-in/clock-out |
| Timing | Sekali per hari (pagi) | Clock-in + clock-out (jam kerja) |
| Keterlambatan | Tidak ditrack | Track menit keterlambatan |
| Pulang awal | Tidak ditrack | Track jam pulang awal |
| Dampak | Rapor, Dapodik | Payroll, PKG, penggantian |
| Volume | Ratusan siswa/hari | Puluhan guru/hari |

### Mengapa Vernon Pattern?

- has_many dari teacher (satu record per hari per guru).
- Relasi ke teacher, academic_year.
- Read-heavy: laporan bulanan, rekap tahunan, dashboard kepala sekolah.
- Write moderate: clock-in/clock-out setiap hari kerja.
- Eventually consistent acceptable — aggregation ke payroll bersifat bulanan.

## Decision

Menggunakan **Vernon Pattern** untuk domain `teacher_attendance` dengan mekanisme clock-in/clock-out.

### Table Schema

```sql
CREATE TABLE teacher_attendances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Attendance data
    attendance_date DATE NOT NULL,
    status          VARCHAR(20) NOT NULL,

    -- Clock-in/clock-out
    clock_in        TIMESTAMPTZ,
    clock_out       TIMESTAMPTZ,
    late_minutes    INT NOT NULL DEFAULT 0,
    early_leave_minutes INT NOT NULL DEFAULT 0,

    -- Lokasi (opsional, untuk fingerprint/GPS)
    clock_in_method  VARCHAR(20),
    clock_out_method VARCHAR(20),
    clock_in_location  TEXT,
    clock_out_location TEXT,

    -- Keterangan
    note            TEXT,
    attachment_url  TEXT,

    -- Validasi
    validated_by    UUID,
    validated_at    TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_teacher_attendance_date UNIQUE (teacher_id, attendance_date),
    CONSTRAINT chk_teacher_att_status CHECK (status IN (
        'present', 'sick', 'permitted', 'absent',
        'dinas_luar', 'cuti', 'libur'
    )),
    CONSTRAINT chk_clock_method CHECK (
        clock_in_method IS NULL OR clock_in_method IN ('fingerprint', 'face_recognition', 'gps', 'manual', 'qr_code')
    ),
    CONSTRAINT chk_clock_out_method CHECK (
        clock_out_method IS NULL OR clock_out_method IN ('fingerprint', 'face_recognition', 'gps', 'manual', 'qr_code')
    ),
    CONSTRAINT chk_late_minutes CHECK (late_minutes >= 0),
    CONSTRAINT chk_early_leave CHECK (early_leave_minutes >= 0)
);

-- Indexes
CREATE INDEX idx_teacher_att_tenant_company ON teacher_attendances (tenant_id, company_id);
CREATE INDEX idx_teacher_att_teacher ON teacher_attendances (teacher_id);
CREATE INDEX idx_teacher_att_date ON teacher_attendances (attendance_date);
CREATE INDEX idx_teacher_att_teacher_month ON teacher_attendances (teacher_id, attendance_date);
CREATE INDEX idx_teacher_att_status ON teacher_attendances (status) WHERE status != 'present';
CREATE INDEX idx_teacher_att_late ON teacher_attendances (late_minutes) WHERE late_minutes > 0;
CREATE INDEX idx_teacher_att_year ON teacher_attendances (academic_year_id);
CREATE INDEX idx_teacher_att_rels ON teacher_attendances USING GIN (_rels);
CREATE INDEX idx_teacher_att_data ON teacher_attendances USING GIN (_data);

-- Konfigurasi jam kerja per sekolah
CREATE TABLE teacher_attendance_configs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Jam kerja
    work_start_time TIME NOT NULL DEFAULT '07:00',
    work_end_time   TIME NOT NULL DEFAULT '14:00',
    late_tolerance_minutes INT NOT NULL DEFAULT 15,
    minimum_work_hours NUMERIC(4,2) NOT NULL DEFAULT 7.0,

    -- Hari kerja
    work_days       JSONB NOT NULL DEFAULT '["monday","tuesday","wednesday","thursday","friday"]',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_att_config_company UNIQUE (tenant_id, company_id)
);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `attendance_date` | DATE, NOT NULL | Satu record per guru per hari — sama seperti S008 |
| `status` | VARCHAR(20), 7 status | Lebih banyak dari S008 karena guru punya `dinas_luar`, `cuti`, `libur` |
| `clock_in` / `clock_out` | TIMESTAMPTZ, nullable | Nullable karena status `absent`/`cuti`/`libur` tidak punya clock |
| `late_minutes` | INT, DEFAULT 0 | Selisih menit dari `work_start_time` — 0 jika tepat waktu atau lebih awal |
| `early_leave_minutes` | INT, DEFAULT 0 | Selisih menit pulang sebelum `work_end_time` |
| `clock_in_method` | VARCHAR(20), 5 metode | Mendukung berbagai perangkat: fingerprint, face recognition, GPS, QR, manual |
| `clock_in_location` | TEXT, nullable | Koordinat GPS atau ID mesin fingerprint — audit trail |
| `attachment_url` | TEXT, nullable | Foto surat dokter, surat tugas dinas luar, dll |
| `work_start_time` | TIME, config table | Konfigurasi per sekolah — sekolah negeri biasanya 07:00, swasta bisa berbeda |
| `late_tolerance_minutes` | INT, DEFAULT 15 | Toleransi keterlambatan sebelum dianggap terlambat |

### Vernon Relationships

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `teacher` | belongs_to | **Ya** | Selalu perlu tahu pemilik absensi — nama, NIP, role |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran untuk pelaporan |

### _rels / _data Structure

```json
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
```

### API Endpoints

```
# Clock-in / Clock-out
POST   /api/v1/teacher-attendances/clock-in           — Clock-in guru (self-service)
POST   /api/v1/teacher-attendances/clock-out          — Clock-out guru (self-service)
POST   /api/v1/teacher-attendances/manual             — Input manual oleh admin (untuk absen sakit/izin/cuti)

# Query
GET    /api/v1/teachers/{id}/attendances              — Riwayat absensi guru
GET    /api/v1/teachers/{id}/attendances/summary      — Rekap bulanan/semesteran
GET    /api/v1/teacher-attendances/today               — Absensi semua guru hari ini
GET    /api/v1/teacher-attendances/late                — Daftar guru terlambat hari ini

# Validasi & Rekap
POST   /api/v1/teacher-attendances/validate           — Validasi oleh kepala sekolah
GET    /api/v1/teacher-attendances/monthly-report     — Rekap bulanan seluruh guru

# Config
GET    /api/v1/teacher-attendance-configs              — Get konfigurasi jam kerja
PUT    /api/v1/teacher-attendance-configs              — Update konfigurasi
```

### Monthly Summary Query

```sql
SELECT
    teacher_id,
    COUNT(*) FILTER (WHERE status = 'present')     AS days_present,
    COUNT(*) FILTER (WHERE status = 'sick')         AS days_sick,
    COUNT(*) FILTER (WHERE status = 'permitted')    AS days_permitted,
    COUNT(*) FILTER (WHERE status = 'absent')       AS days_absent,
    COUNT(*) FILTER (WHERE status = 'dinas_luar')   AS days_dinas_luar,
    COUNT(*) FILTER (WHERE status = 'cuti')         AS days_cuti,
    SUM(late_minutes)                                AS total_late_minutes,
    SUM(early_leave_minutes)                         AS total_early_leave_minutes
FROM teacher_attendances
WHERE attendance_date >= $1 AND attendance_date < $2
  AND tenant_id = $3 AND company_id = $4
GROUP BY teacher_id;
```

Hasil ini digunakan oleh S031 (Payroll) untuk menghitung potongan/insentif.

## Consequences

### Positive

- **Clock-in/clock-out**: Mendukung jam kerja aktual, bukan hanya status hadir/tidak.
- **Keterlambatan terukur**: `late_minutes` memungkinkan kebijakan potongan per menit.
- **Multi-metode**: Fingerprint, face recognition, GPS, QR code — fleksibel sesuai infrastruktur sekolah.
- **Payroll ready**: Data bulanan langsung bisa dikonsumsi S031 untuk penggajian.
- **Pesantren compatible**: Ustadz/ustadzah menggunakan tabel yang sama — terminologi di presentation layer.

### Negative / Trade-offs

- **Satu record per hari**: Tidak mendukung multiple clock-in/out (misal: keluar siang, masuk lagi). Bisa ditambah sebagai enhancement.
- **Config per company**: Semua guru di satu sekolah punya jam kerja sama. Jika ada shift berbeda, perlu enhancement.
- **Manual override**: Admin bisa input manual — perlu audit trail ketat.
- **Perangkat hardware**: Integrasi fingerprint/face recognition memerlukan middleware terpisah.

## Alternatives Considered

### 1. Extend S008 (student attendance) untuk guru
- Ditolak: fundamental berbeda — S008 adalah guru mengabsen siswa, S026 adalah self clock-in/out. Field berbeda (clock time, late tracking, method).

### 2. Multiple clock-in/out per hari (shift model)
- Deferred: sekolah Indonesia umumnya single shift. Multi-shift bisa ditambah nanti jika ada demand dari sekolah boarding/pesantren.

### 3. Real-time streaming dari perangkat
- Deferred: MVP menggunakan REST API. WebSocket/streaming bisa ditambah untuk integrasi mesin absensi real-time.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` returns correct name | -- | `"teacher_attendances"` |
| U02 | `DefaultRels()` returns 2 autoloaded rels | -- | `teacher`, `academic_year` |
| U03 | Validate rejects invalid `status` | `{ ..., "status": "late" }` | Error: invalid status |
| U04 | Validate rejects invalid `clock_in_method` | `{ ..., "clock_in_method": "sms" }` | Error: invalid method |
| U05 | Validate rejects negative `late_minutes` | `{ ..., "late_minutes": -5 }` | Error: must be >= 0 |
| U06 | Validate rejects missing `teacher_id` | `{ ..., "teacher_id": null }` | Error: teacher_id required |
| U07 | Calculate late_minutes from clock_in and config | clock_in=07:20, start=07:00, tolerance=15 | late_minutes=5 |
| U08 | No late if within tolerance | clock_in=07:10, start=07:00, tolerance=15 | late_minutes=0 |
| U09 | Calculate early_leave_minutes | clock_out=13:00, end=14:00 | early_leave_minutes=60 |
| U10 | Validate accepts valid attendance | All required fields valid | No error |

### Integration Tests — Clock-in / Clock-out

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Clock-in creates attendance record | POST /clock-in at 07:00 | 201, record with status=present, clock_in set |
| I02 | Clock-out updates existing record | POST /clock-out at 14:00 | 200, clock_out set |
| I03 | Duplicate clock-in same day rejected | POST /clock-in twice same day | 409, unique constraint |
| I04 | Clock-in calculates late_minutes | POST /clock-in at 07:25 (tolerance=15) | late_minutes=10 |
| I05 | Clock-out calculates early_leave | POST /clock-out at 12:30 (end=14:00) | early_leave_minutes=90 |
| I06 | Manual input by admin | POST /manual with status=sick + attachment | 201, status=sick, attachment_url set |

### Integration Tests — Query & Report

| # | Test Case | Action | Expected |
|---|---|---|---|
| I07 | Get today's attendance | GET /teacher-attendances/today | 200, all teachers with status |
| I08 | Get late teachers | GET /teacher-attendances/late | 200, only teachers with late_minutes > 0 |
| I09 | Get teacher history | GET /teachers/{id}/attendances?month=2026-04 | 200, filtered by month |
| I10 | Monthly summary | GET /teachers/{id}/attendances/summary?month=2026-04 | 200, aggregated counts |
| I11 | Monthly report all teachers | GET /teacher-attendances/monthly-report?month=2026-04 | 200, all teachers summarized |

### Integration Tests — Validation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I12 | Validate attendance by kepala sekolah | POST /validate with date range | 200, validated_by + validated_at set |
| I13 | Non-kepsek cannot validate | POST /validate by guru_mapel | 403, forbidden |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | TeacherUpdated syncs to attendances | Update teacher full_name | `_data.teacher.full_name` updated |
| I15 | AcademicYearUpdated syncs | Update academic_year name | `_data.academic_year.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Cannot access other tenant's attendance | GET with wrong tenant scope | 404 |
| I17 | Cannot clock-in for other company's teacher | POST /clock-in with wrong company | Error, scope violation |
