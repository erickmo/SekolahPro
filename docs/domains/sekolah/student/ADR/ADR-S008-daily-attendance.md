# ADR-S008: Daily Attendance Transaction

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

ADR-S004 menyimpan **agregat kehadiran per semester** (days_present, days_sick, days_permitted, days_absent), namun belum ada mekanisme pencatatan **kehadiran harian** yang menjadi sumber data agregat tersebut.

Absensi harian adalah operasi paling sering di sekolah:

1. **Operasional harian**: Guru wali kelas mengisi absensi setiap pagi — ini aktivitas pertama setiap hari sekolah.
2. **Monitoring real-time**: Orang tua dan admin perlu melihat status kehadiran hari ini.
3. **Surat izin**: Dokumentasi izin sakit/keperluan dengan bukti surat.
4. **Peringatan**: Sistem alert jika siswa alpa berturut-turut (3+ hari).
5. **Agregasi ke S004**: Sumber data untuk `days_present`, `days_sick`, `days_permitted`, `days_absent` di academic record.
6. **Dapodik**: Pelaporan kehadiran ke Kemendikbud per semester.

Karakteristik operasional:
- **Bulk operation**: Guru mengisi absensi untuk 30-40 siswa sekaligus per kelas.
- **Write-heavy harian, read-heavy bulanan**: Write setiap pagi, read saat rekap dan laporan.
- **Immutable setelah validasi**: Absensi yang sudah divalidasi kepala sekolah tidak boleh diubah.

### Mengapa Vernon Pattern?

- has_many dari student (ratusan records per siswa per tahun).
- Relasi ke student, class_room, academic_year.
- Bulk read untuk rekap bulanan dan semester.
- Eventually consistent acceptable — agregat S004 bisa dihitung async.

## Decision

Menggunakan **Vernon Pattern** untuk domain `student_attendance` dengan optimasi untuk bulk insert.

### Table Schema

```sql
CREATE TABLE student_attendances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    class_room_id   UUID NOT NULL,

    -- Attendance data
    attendance_date DATE NOT NULL,
    semester        VARCHAR(10) NOT NULL,
    status          VARCHAR(20) NOT NULL,
    note            TEXT,

    -- Validasi
    recorded_by     UUID NOT NULL,
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

    CONSTRAINT uq_attendance_student_date UNIQUE (student_id, attendance_date),
    CONSTRAINT chk_attendance_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_attendance_status CHECK (status IN ('present', 'sick', 'permitted', 'absent'))
);

-- Indexes
CREATE INDEX idx_attendance_tenant_company ON student_attendances (tenant_id, company_id);
CREATE INDEX idx_attendance_student ON student_attendances (student_id);
CREATE INDEX idx_attendance_date ON student_attendances (attendance_date);
CREATE INDEX idx_attendance_class_date ON student_attendances (class_room_id, attendance_date);
CREATE INDEX idx_attendance_year_semester ON student_attendances (academic_year_id, semester);
CREATE INDEX idx_attendance_status ON student_attendances (status) WHERE status != 'present';
CREATE INDEX idx_attendance_rels ON student_attendances USING GIN (_rels);
CREATE INDEX idx_attendance_data ON student_attendances USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `attendance_date` | DATE, NOT NULL | Satu record per siswa per hari |
| `semester` | VARCHAR(10), CHECK | Denormalisasi dari academic_year — memudahkan aggregation query tanpa JOIN |
| `status` | VARCHAR(20), CHECK | 4 status standar Dapodik: hadir, sakit, izin, alpha |
| `note` | TEXT, nullable | Keterangan tambahan (nomor surat izin, alasan sakit) |
| `recorded_by` | UUID, NOT NULL | ID user yang memasukkan data — audit trail |
| `validated_by` | UUID, nullable | ID kepala sekolah/wakasek yang memvalidasi |
| `validated_at` | TIMESTAMPTZ, nullable | Timestamp validasi — setelah ini record immutable |

### Vernon Relationships

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Selalu perlu tahu pemilik absensi |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |
| `class_room` | belongs_to | **Ya** | Konteks kelas |

### _rels / _data Structure

```json
{
  "_rels": {
    "student_id": "018f...",
    "academic_year_id": "018f...",
    "class_room_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" },
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "class_room": { "id": "018f...", "name": "VII-A", "grade_level": "7" }
  }
}
```

### API Endpoints

```
POST   /api/v1/student-attendances/bulk            — Bulk insert absensi kelas (1 request = 30-40 siswa)
GET    /api/v1/students/{id}/attendances            — Riwayat absensi siswa
GET    /api/v1/students/{id}/attendances/summary    — Rekap absensi per bulan/semester
GET    /api/v1/class-rooms/{id}/attendances/{date}  — Absensi kelas per tanggal
PUT    /api/v1/student-attendances/{id}             — Update absensi (sebelum validasi)
POST   /api/v1/student-attendances/validate         — Validasi absensi oleh kepsek
```

### Bulk Insert Payload

```json
{
  "class_room_id": "018f...",
  "academic_year_id": "018f...",
  "attendance_date": "2026-04-15",
  "semester": "genap",
  "attendances": [
    { "student_id": "018f...", "status": "present" },
    { "student_id": "018f...", "status": "sick", "note": "Surat dokter #123" },
    { "student_id": "018f...", "status": "absent" }
  ]
}
```

### Aggregation to S004

Ketika rekap semester dilakukan, sistem menghitung dari `student_attendances`:

```sql
SELECT
    student_id,
    COUNT(*) FILTER (WHERE status = 'present')   AS days_present,
    COUNT(*) FILTER (WHERE status = 'sick')       AS days_sick,
    COUNT(*) FILTER (WHERE status = 'permitted')  AS days_permitted,
    COUNT(*) FILTER (WHERE status = 'absent')     AS days_absent
FROM student_attendances
WHERE academic_year_id = $1 AND semester = $2
GROUP BY student_id;
```

Hasil ini di-update ke `student_academics` (S004) via event `AttendanceAggregated`.

## Consequences

### Positive

- **Granular**: Data harian memungkinkan laporan per hari, minggu, bulan, semester.
- **Bulk friendly**: Endpoint bulk insert dirancang untuk use case guru mengisi 1 kelas sekaligus.
- **Audit trail**: `recorded_by` + `validated_by` memberikan akuntabilitas lengkap.
- **Immutable setelah validasi**: Data yang sudah divalidasi tidak bisa diubah — integritas terjaga.
- **Sumber kebenaran**: S004 aggregat dihitung dari data harian ini, bukan diinput manual.

### Negative / Trade-offs

- **Volume tinggi**: 30 siswa × 200 hari efektif = 6.000 rows per kelas per tahun. Untuk sekolah 500 siswa = 100.000 rows per tahun.
- **Semester denormalisasi**: `semester` disimpan per row (bukan lookup) — trade-off untuk query speed.
- **Aggregation async**: Rekap ke S004 bersifat eventual consistent, bukan real-time.
- **Validasi workflow**: Perlu business rule tambahan: siapa yang boleh validasi, deadline validasi.

## Alternatives Considered

### 1. Absensi per mata pelajaran (bukan per hari)
- Ditolak untuk MVP: terlalu granular — sekolah dasar dan menengah di Indonesia mayoritas menggunakan absensi per hari per kelas, bukan per sesi/mapel. Bisa ditambah sebagai enhancement.

### 2. Absensi langsung di S004 (aggregate only)
- Ditolak: tidak bisa drill-down ke hari tertentu, tidak bisa menampilkan "siapa yang tidak hadir hari ini".

### 3. JSONB array per bulan
- Ditolak: sulit di-query per hari, sulit di-aggregate, dan partial update JSONB array tidak efisien.

### 4. Separate table per bulan (partitioning)
- Deferred: bisa ditambahkan nanti jika volume data menjadi masalah performance. Untuk MVP, single table dengan proper indexing cukup.

## Test Cases

### Unit Tests — Descriptor & Validation

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` returns correct name | — | `"student_attendances"` |
| U02 | `DefaultRels()` returns 3 autoloaded rels | — | `student`, `academic_year`, `class_room` |
| U03 | Validate rejects invalid `status` | `{ ..., "status": "late" }` | Error: status must be present/sick/permitted/absent |
| U04 | Validate rejects invalid `semester` | `{ ..., "semester": "midterm" }` | Error: semester must be ganjil/genap |
| U05 | Validate rejects missing `attendance_date` | `{ ..., "attendance_date": null }` | Error: attendance_date required |
| U06 | Validate rejects missing `recorded_by` | `{ ..., "recorded_by": null }` | Error: recorded_by required |
| U07 | Validate accepts valid attendance | All required fields valid | No error |

### Integration Tests — Bulk Insert

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Bulk insert attendance for class | POST bulk with 30 students | 201, 30 records created with `_data` populated |
| I02 | Bulk insert — duplicate date rejected | Bulk insert same class + same date twice | 409/422, unique constraint on `(student_id, attendance_date)` |
| I03 | Bulk insert — partial failure rolls back | 29 valid + 1 invalid student_id | 422, no records created (atomic) |
| I04 | Bulk insert populates `_data` | Check created records | `_data` contains student, academic_year, class_room |

### Integration Tests — CRUD

| # | Test Case | Action | Expected |
|---|---|---|---|
| I05 | Get student attendance history | `GET /students/{id}/attendances` | 200, ordered by date desc |
| I06 | Get class attendance for date | `GET /class-rooms/{id}/attendances/2026-04-15` | 200, all students in class with status |
| I07 | Update attendance before validation | `PUT /student-attendances/{id}` change status | 200, updated |
| I08 | Update attendance after validation | `PUT /student-attendances/{id}` on validated record | 403, immutable after validation |
| I09 | Filter by date range | `GET /students/{id}/attendances?from=2026-04-01&to=2026-04-30` | 200, filtered results |
| I10 | Filter by status | `GET /students/{id}/attendances?status=absent` | 200, only absent records |

### Integration Tests — Validation Workflow

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Validate attendance | `POST /student-attendances/validate` with date + class | 200, `validated_by` + `validated_at` set |
| I12 | Non-authorized user cannot validate | POST validate by teacher (not kepsek) | 403, forbidden |

### Integration Tests — Summary & Aggregation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | Get monthly summary | `GET /students/{id}/attendances/summary?month=2026-04` | 200, counts per status for April |
| I14 | Get semester summary | `GET /students/{id}/attendances/summary?semester=genap&year_id={id}` | 200, total counts matching S004 format |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I15 | StudentUpdated syncs to attendances | Update student `full_name` | `_data.student.full_name` updated |
| I16 | ClassRoomUpdated syncs to attendances | Update class_room name | `_data.class_room.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Cannot access other tenant's attendance | GET with wrong tenant scope | 404 |
| I18 | Cannot bulk insert for other company's class | POST bulk with wrong company scope | Error, scope violation |
