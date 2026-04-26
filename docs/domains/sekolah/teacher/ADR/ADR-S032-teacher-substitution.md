# ADR-S032: Teacher Schedule & Substitution (Jadwal Piket & Penggantian Guru)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Ketika guru tidak hadir (sakit, cuti, dinas luar), kelas tidak boleh kosong tanpa pengawasan. Di sekolah Indonesia, mekanisme penggantian melibatkan:

1. **Penggantian Guru Mengajar (Substitusi)**: Guru lain menggantikan mengajar di kelas yang ditinggalkan. Idealnya guru pengganti mengajar mata pelajaran yang sama, tapi seringkali guru piket hanya mengawasi tugas yang diberikan guru asli.
2. **Jadwal Piket Harian**: Rotasi guru yang bertugas jaga setiap hari — guru piket bertanggung jawab untuk:
   - Mengawasi kelas yang kosong (jika tidak ada guru pengganti mapel)
   - Menyambut siswa di gerbang (piket pagi)
   - Mengurus administrasi harian (buku piket)
   - Koordinasi jika ada keadaan darurat
3. **Pesantren**: Piket pengasuhan (musyrif/musyrifah) untuk asrama — berbeda dari piket mengajar.

Hubungan antar domain:
- **S021 (Timetable)**: Menentukan jadwal reguler — dari sini diketahui kelas mana yang kosong jika guru absen.
- **S026 (Teacher Attendance)**: Mendeteksi guru yang tidak hadir hari ini.
- **S030 (Leave Management)**: Cuti yang sudah diapprove — bisa dijadwalkan substitusi lebih awal.
- **S027 (Teacher Workload)**: Beban mengajar guru — substitusi harus mempertimbangkan agar tidak overload guru pengganti.

Auto-suggest logic untuk guru pengganti:
1. **Prioritas 1**: Guru mapel yang sama dan sedang free period (dari S021).
2. **Prioritas 2**: Guru mapel serumpun yang sedang free period.
3. **Prioritas 3**: Guru piket hari itu.
4. **Prioritas 4**: Guru manapun yang sedang free period dan beban mengajar terendah (dari S027).

### Mengapa Vernon Pattern?

- has_many dari teacher — banyak record piket/substitusi per guru per semester.
- Relasi ke teacher (ADR-012), timetable (S021), attendance (S026), leave (S030).
- Read-heavy: dashboard piket harian, laporan substitusi bulanan.
- Write moderate: piket di-setup per semester, substitusi terjadi beberapa kali per minggu.
- Eventually consistent acceptable — piket dan substitusi bisa di-assign beberapa jam sebelumnya.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `duty_schedules` (jadwal piket harian), `teacher_substitutions` (penggantian guru mengajar), dan `substitution_logs` (log aktivitas substitusi).

### Table Schema

```sql
-- Jadwal piket harian (rotasi per semester)
CREATE TABLE duty_schedules (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Jadwal
    semester        VARCHAR(10) NOT NULL,
    day_of_week     INT NOT NULL,
    duty_type       VARCHAR(30) NOT NULL,

    -- Detail
    start_time      TIME,
    end_time        TIME,
    location        VARCHAR(100),
    note            TEXT,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,
    effective_date  DATE,
    end_date        DATE,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_duty_teacher_day_type UNIQUE (teacher_id, academic_year_id, semester, day_of_week, duty_type),
    CONSTRAINT chk_duty_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_duty_day CHECK (day_of_week >= 1 AND day_of_week <= 7),
    CONSTRAINT chk_duty_type CHECK (duty_type IN (
        'piket_pagi', 'piket_kelas', 'piket_gerbang',
        'piket_upacara', 'piket_siang', 'piket_asrama',
        'piket_malam'
    )),
    CONSTRAINT chk_duty_dates CHECK (end_date IS NULL OR end_date >= effective_date),
    CONSTRAINT chk_duty_time CHECK (end_time IS NULL OR end_time > start_time)
);

-- Indexes
CREATE INDEX idx_duty_tenant_company ON duty_schedules (tenant_id, company_id);
CREATE INDEX idx_duty_teacher ON duty_schedules (teacher_id);
CREATE INDEX idx_duty_year_semester ON duty_schedules (academic_year_id, semester);
CREATE INDEX idx_duty_day ON duty_schedules (day_of_week);
CREATE INDEX idx_duty_type ON duty_schedules (duty_type);
CREATE INDEX idx_duty_active ON duty_schedules (is_active) WHERE is_active = true;
CREATE INDEX idx_duty_teacher_day ON duty_schedules (teacher_id, day_of_week);
CREATE INDEX idx_duty_rels ON duty_schedules USING GIN (_rels);
CREATE INDEX idx_duty_data ON duty_schedules USING GIN (_data);

-- Penggantian guru mengajar
CREATE TABLE teacher_substitutions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    original_teacher_id UUID NOT NULL,
    substitute_teacher_id UUID NOT NULL,
    schedule_entry_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Tanggal dan waktu substitusi
    substitution_date DATE NOT NULL,
    time_slot_id    UUID NOT NULL,

    -- Alasan
    reason_type     VARCHAR(30) NOT NULL,
    reason_detail   TEXT,
    leave_request_id UUID,

    -- Subject (bisa berbeda jika pengganti bukan guru mapel yang sama)
    original_subject_id UUID NOT NULL,
    substitute_subject_id UUID,
    is_same_subject BOOLEAN NOT NULL DEFAULT false,

    -- Tugas
    task_description TEXT,
    task_attachment_url TEXT,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    notified_at     TIMESTAMPTZ,
    accepted_at     TIMESTAMPTZ,
    declined_at     TIMESTAMPTZ,
    decline_reason  TEXT,
    completed_at    TIMESTAMPTZ,

    -- Kelas
    class_room_id   UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_substitution_slot_class UNIQUE (substitution_date, time_slot_id, class_room_id),
    CONSTRAINT chk_sub_reason CHECK (reason_type IN (
        'sakit', 'cuti', 'izin', 'dinas_luar',
        'tugas_belajar', 'terlambat', 'lain_lain'
    )),
    CONSTRAINT chk_sub_status CHECK (status IN (
        'pending', 'notified', 'accepted', 'declined',
        'in_progress', 'completed', 'cancelled'
    )),
    CONSTRAINT chk_different_teacher CHECK (original_teacher_id != substitute_teacher_id)
);

-- Indexes
CREATE INDEX idx_sub_tenant_company ON teacher_substitutions (tenant_id, company_id);
CREATE INDEX idx_sub_original_teacher ON teacher_substitutions (original_teacher_id);
CREATE INDEX idx_sub_substitute_teacher ON teacher_substitutions (substitute_teacher_id);
CREATE INDEX idx_sub_date ON teacher_substitutions (substitution_date);
CREATE INDEX idx_sub_schedule ON teacher_substitutions (schedule_entry_id);
CREATE INDEX idx_sub_timeslot ON teacher_substitutions (time_slot_id);
CREATE INDEX idx_sub_status ON teacher_substitutions (status);
CREATE INDEX idx_sub_pending ON teacher_substitutions (status) WHERE status IN ('pending', 'notified');
CREATE INDEX idx_sub_leave ON teacher_substitutions (leave_request_id) WHERE leave_request_id IS NOT NULL;
CREATE INDEX idx_sub_class ON teacher_substitutions (class_room_id);
CREATE INDEX idx_sub_date_class ON teacher_substitutions (substitution_date, class_room_id);
CREATE INDEX idx_sub_year ON teacher_substitutions (academic_year_id);
CREATE INDEX idx_sub_rels ON teacher_substitutions USING GIN (_rels);
CREATE INDEX idx_sub_data ON teacher_substitutions USING GIN (_data);

-- Log aktivitas substitusi (audit trail)
CREATE TABLE substitution_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    substitution_id UUID NOT NULL,
    actor_id        UUID NOT NULL,

    -- Action
    action          VARCHAR(30) NOT NULL,
    detail          TEXT,
    previous_status VARCHAR(20),
    new_status      VARCHAR(20),

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_log_action CHECK (action IN (
        'created', 'notified', 'accepted', 'declined',
        'reassigned', 'started', 'completed', 'cancelled'
    ))
);

-- Indexes
CREATE INDEX idx_sub_log_tenant_company ON substitution_logs (tenant_id, company_id);
CREATE INDEX idx_sub_log_substitution ON substitution_logs (substitution_id);
CREATE INDEX idx_sub_log_actor ON substitution_logs (actor_id);
CREATE INDEX idx_sub_log_action ON substitution_logs (action);
CREATE INDEX idx_sub_log_rels ON substitution_logs USING GIN (_rels);
CREATE INDEX idx_sub_log_data ON substitution_logs USING GIN (_data);
```

### Field Design Rationale

**duty_schedules:**

| Field | Keputusan | Alasan |
|---|---|---|
| `duty_type` | 7 tipe piket | Mencakup piket umum (pagi, kelas, gerbang, upacara, siang) + piket pesantren (asrama, malam) |
| `day_of_week` | INT, 1-7 | Rotasi per hari — setiap hari ada guru piket berbeda |
| `start_time` / `end_time` | TIME, nullable | Piket bisa full-day (null) atau waktu tertentu (piket pagi: 06:30-07:30) |
| `location` | VARCHAR(100), nullable | Lokasi piket: "Gerbang Utama", "Lobby Lantai 2", "Asrama Putra" |
| `effective_date` / `end_date` | DATE, nullable | Jadwal piket bisa per periode — tidak harus satu semester penuh |

**teacher_substitutions:**

| Field | Keputusan | Alasan |
|---|---|---|
| `original_teacher_id` | UUID, NOT NULL | Guru yang tidak hadir — dari S026 attendance |
| `substitute_teacher_id` | UUID, NOT NULL | Guru pengganti — dari auto-suggest atau manual assignment |
| `schedule_entry_id` | UUID, NOT NULL | Jadwal yang perlu digantikan — dari S021 timetable |
| `reason_type` | 7 alasan | Alasan ketidakhadiran: sakit, cuti (link ke S030), izin, dinas_luar, dll |
| `leave_request_id` | UUID, nullable | Link ke S030 jika alasan = cuti — untuk traceability |
| `original_subject_id` | UUID | Mapel yang seharusnya diajarkan |
| `substitute_subject_id` | UUID, nullable | Mapel yang benar-benar diajarkan — bisa sama atau berbeda |
| `is_same_subject` | BOOLEAN | Apakah pengganti mengajar mapel yang sama — untuk laporan kualitas substitusi |
| `task_description` | TEXT, nullable | Tugas yang diberikan guru asli: "Kerjakan LKS halaman 45-47" |
| `task_attachment_url` | TEXT, nullable | File tugas (PDF/gambar) yang di-upload guru asli |
| `status` | 7 status | Lifecycle: pending -> notified -> accepted/declined -> in_progress -> completed |
| `class_room_id` | UUID, NOT NULL | Kelas yang butuh pengganti — denormalisasi dari schedule_entry |

### Auto-Suggest Algorithm

```sql
-- Step 1: Dapatkan slot waktu yang perlu pengganti
WITH absent_slots AS (
    SELECT
        se.id AS schedule_entry_id,
        se.time_slot_id,
        se.teacher_id AS original_teacher_id,
        se.subject_id AS original_subject_id,
        se.class_room_id,
        ts.day_of_week,
        ts.start_time,
        ts.end_time,
        s.subject_group
    FROM schedule_entries se
    JOIN time_slots ts ON ts.id = se.time_slot_id
    JOIN subjects s ON s.id = se.subject_id
    WHERE se.teacher_id = $absent_teacher_id
      AND ts.day_of_week = EXTRACT(ISODOW FROM $substitution_date::date)
      AND se.is_substitution = false
      AND se.deleted_at IS NULL
),

-- Step 2: Cari guru yang FREE di slot tersebut
free_teachers AS (
    SELECT DISTINCT t.id AS teacher_id, t.full_name, t.employee_type
    FROM teachers t
    WHERE t.tenant_id = $tenant_id AND t.company_id = $company_id
      AND t.id != $absent_teacher_id
      AND t.is_active = true
      AND t.deleted_at IS NULL
      -- Guru hadir hari ini (dari S026)
      AND EXISTS (
          SELECT 1 FROM teacher_attendances ta
          WHERE ta.teacher_id = t.id
            AND ta.attendance_date = $substitution_date
            AND ta.status = 'present'
      )
      -- Tidak ada jadwal di slot yang sama
      AND NOT EXISTS (
          SELECT 1 FROM schedule_entries se2
          JOIN absent_slots abs ON abs.time_slot_id = se2.time_slot_id
          WHERE se2.teacher_id = t.id
            AND se2.deleted_at IS NULL
            AND se2.is_substitution = false
      )
      -- Belum dijadwalkan substitusi di slot yang sama
      AND NOT EXISTS (
          SELECT 1 FROM teacher_substitutions tsub
          JOIN absent_slots abs ON abs.time_slot_id = tsub.time_slot_id
          WHERE tsub.substitute_teacher_id = t.id
            AND tsub.substitution_date = $substitution_date
            AND tsub.status NOT IN ('declined', 'cancelled')
      )
),

-- Step 3: Ranking berdasarkan prioritas
ranked_substitutes AS (
    SELECT
        ft.teacher_id,
        ft.full_name,
        -- Prioritas 1: Guru mapel yang sama
        CASE WHEN EXISTS (
            SELECT 1 FROM teacher_subjects ts2
            JOIN absent_slots abs ON abs.original_subject_id = ts2.subject_id
            WHERE ts2.teacher_id = ft.teacher_id
        ) THEN 1
        -- Prioritas 2: Guru mapel serumpun
        WHEN EXISTS (
            SELECT 1 FROM teacher_subjects ts3
            JOIN subjects s2 ON s2.id = ts3.subject_id
            JOIN absent_slots abs ON abs.subject_group = s2.subject_group
            WHERE ts3.teacher_id = ft.teacher_id
        ) THEN 2
        -- Prioritas 3: Guru piket hari ini
        WHEN EXISTS (
            SELECT 1 FROM duty_schedules ds
            WHERE ds.teacher_id = ft.teacher_id
              AND ds.day_of_week = EXTRACT(ISODOW FROM $substitution_date::date)
              AND ds.is_active = true
        ) THEN 3
        -- Prioritas 4: Guru lain (sort by workload terendah)
        ELSE 4
        END AS priority,
        -- Sub-sort by workload (dari S027)
        COALESCE(tw.total_hours, 0) AS current_workload
    FROM free_teachers ft
    LEFT JOIN teacher_workloads tw ON tw.teacher_id = ft.teacher_id
        AND tw.academic_year_id = $academic_year_id
)
SELECT * FROM ranked_substitutes
ORDER BY priority ASC, current_workload ASC
LIMIT 5;
```

### Notification Flow

```
1. Guru absen terdeteksi (dari S026 atau S030 leave approval)
2. Sistem menjalankan auto-suggest untuk setiap slot jadwal yang kosong
3. Admin/wakakurikulum memilih guru pengganti (atau accept auto-suggestion)
4. Substitusi dibuat dengan status = 'pending'
5. Notifikasi dikirim ke guru pengganti (via S044 Notification System):
   - "Anda diminta menggantikan Bu Siti mengajar Matematika di kelas VII-A jam ke-3 (08:50-09:30)"
6. Guru pengganti accept/decline
7. Jika decline → re-assign ke guru lain (suggest ulang)
8. Jika accept → status = 'accepted'
9. Setelah selesai mengajar → status = 'completed'
```

### Vernon Relationships

**duty_schedules:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `teacher` | belongs_to | **Ya** | Guru yang piket — nama, role |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |

**teacher_substitutions:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `original_teacher` | belongs_to | **Ya** | Guru yang absen |
| `substitute_teacher` | belongs_to | **Ya** | Guru pengganti |
| `schedule_entry` | belongs_to | **Ya** | Jadwal yang digantikan |
| `time_slot` | belongs_to | **Ya** | Slot waktu |
| `class_room` | belongs_to | **Ya** | Kelas |
| `original_subject` | belongs_to | **Ya** | Mapel asli |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |

**substitution_logs:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `substitution` | belongs_to | **Ya** | Parent substitution |
| `actor` | belongs_to | **Ya** | Siapa yang melakukan action |

### _rels / _data Structure

```json
// duty_schedules
{
  "_rels": {
    "teacher_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "teacher": {
      "id": "018f...",
      "full_name": "Pak Budi Santoso",
      "nip": "198703152011011002",
      "employee_type": "pns",
      "role": "guru_mapel"
    },
    "academic_year": {
      "id": "018f...",
      "name": "2025/2026"
    }
  }
}

// teacher_substitutions
{
  "_rels": {
    "original_teacher_id": "018f...",
    "substitute_teacher_id": "018f...",
    "schedule_entry_id": "018f...",
    "time_slot_id": "018f...",
    "class_room_id": "018f...",
    "original_subject_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "original_teacher": {
      "id": "018f...",
      "full_name": "Bu Siti Aminah",
      "nip": "198501012010012001"
    },
    "substitute_teacher": {
      "id": "018f...",
      "full_name": "Pak Budi Santoso",
      "nip": "198703152011011002"
    },
    "schedule_entry": {
      "id": "018f...",
      "room_name": null
    },
    "time_slot": {
      "id": "018f...",
      "name": "Jam ke-3",
      "day_of_week": 1,
      "start_time": "08:50",
      "end_time": "09:30"
    },
    "class_room": {
      "id": "018f...",
      "name": "VII-A"
    },
    "original_subject": {
      "id": "018f...",
      "name": "Matematika",
      "code": "MTK"
    },
    "academic_year": {
      "id": "018f...",
      "name": "2025/2026"
    }
  }
}

// substitution_logs
{
  "_rels": {
    "substitution_id": "018f...",
    "actor_id": "018f..."
  },
  "_data": {
    "substitution": {
      "id": "018f...",
      "original_teacher_name": "Bu Siti Aminah",
      "substitute_teacher_name": "Pak Budi Santoso",
      "class_name": "VII-A",
      "subject_name": "Matematika"
    },
    "actor": {
      "id": "018f...",
      "full_name": "Pak Ahmad",
      "role": "wakil_kepsek"
    }
  }
}
```

### API Endpoints

```
# Duty Schedules (Jadwal Piket)
GET    /api/v1/duty-schedules                              -- List jadwal piket
GET    /api/v1/duty-schedules/today                        -- Guru piket hari ini
GET    /api/v1/teachers/{id}/duty-schedules                -- Jadwal piket per guru
POST   /api/v1/duty-schedules                              -- Tambah jadwal piket
POST   /api/v1/duty-schedules/bulk                         -- Bulk create (setup semester)
PUT    /api/v1/duty-schedules/{id}                         -- Update jadwal piket
DELETE /api/v1/duty-schedules/{id}                         -- Hapus jadwal piket
POST   /api/v1/duty-schedules/carry-forward                -- Copy dari semester sebelumnya

# Teacher Substitutions
POST   /api/v1/teacher-substitutions                       -- Buat substitusi (manual assign)
GET    /api/v1/teacher-substitutions/{id}                  -- Detail substitusi
GET    /api/v1/teacher-substitutions/today                 -- Daftar substitusi hari ini
GET    /api/v1/teacher-substitutions?date=2026-04-15       -- Substitusi per tanggal
GET    /api/v1/teachers/{id}/substitutions                 -- Riwayat substitusi guru (sebagai pengganti)
DELETE /api/v1/teacher-substitutions/{id}                  -- Cancel substitusi

# Auto-Suggest
POST   /api/v1/teacher-substitutions/suggest               -- Auto-suggest guru pengganti
  Body: { absent_teacher_id, substitution_date }
  Response: { suggestions: [{ teacher_id, full_name, priority, reason, current_workload }] }

# Bulk Substitution (guru sakit >1 hari)
POST   /api/v1/teacher-substitutions/bulk                  -- Buat substitusi multi-hari
  Body: { absent_teacher_id, start_date, end_date, reason_type, reason_detail }

# Accept / Decline
POST   /api/v1/teacher-substitutions/{id}/accept           -- Guru pengganti terima
POST   /api/v1/teacher-substitutions/{id}/decline          -- Guru pengganti tolak
  Body: { decline_reason }
POST   /api/v1/teacher-substitutions/{id}/complete         -- Mark selesai mengajar

# Reports
GET    /api/v1/teacher-substitutions/report?month=2026-04  -- Rekap substitusi bulanan
GET    /api/v1/teacher-substitutions/statistics             -- Statistik: guru paling sering absen, paling sering jadi pengganti
```

## Consequences

### Positive

- **Auto-suggest**: Sistem merekomendasikan guru pengganti berdasarkan 4 level prioritas — hemat waktu admin.
- **Free period aware**: Integrasi dengan S021 timetable memastikan guru pengganti tidak double-booked.
- **Workload balanced**: Substitusi mempertimbangkan beban mengajar (S027) — distribusi merata.
- **Leave-linked**: Cuti yang sudah diapprove (S030) bisa langsung generate substitusi di depan — proaktif, bukan reaktif.
- **Notification**: Guru pengganti dinotifikasi dan bisa accept/decline — transparan dan consent-based.
- **Audit trail**: Substitution logs mencatat seluruh lifecycle — siapa assign, siapa accept, kapan selesai.
- **Piket + Substitusi**: Dua mekanisme saling melengkapi — guru piket menjadi fallback jika tidak ada guru mapel yang free.
- **Pesantren support**: Duty type mencakup piket asrama dan malam — cocok untuk pondok pesantren.

### Negative / Trade-offs

- **Auto-suggest imperfect**: Algoritma ranking bisa dipertanyakan — admin tetap bisa override.
- **Decline chain**: Jika guru pengganti decline berulang, bisa delay penugasan — perlu timeout.
- **Task handover manual**: Guru asli harus upload tugas sendiri — jika tidak, guru pengganti mengajar tanpa arahan.
- **Subject mismatch**: Jika tidak ada guru mapel yang sama yang free, kualitas substitusi menurun — ini realita di sekolah.
- **Depends on S021 + S026**: Tanpa jadwal dan attendance, auto-suggest tidak berfungsi — perlu fallback manual.
- **Notification dependency**: Tanpa S044 Notification System, notifikasi harus manual (WA/telepon).

## Alternatives Considered

### 1. Substitusi sebagai field di schedule_entries (S021)
- Ditolak: schedule_entries sudah ada field `is_substitution`, tapi tidak cukup untuk workflow accept/decline, task handover, dan audit trail.

### 2. Tanpa auto-suggest (full manual)
- Ditolak: admin harus cek jadwal satu-satu untuk cari guru free — time-consuming, error-prone.

### 3. Piket tanpa tabel terpisah (JSONB di config)
- Ditolak: perlu query "siapa piket hari ini" — JSONB tidak efisien untuk ini. Tabel relasional lebih baik.

### 4. Auto-assign tanpa accept/decline
- Ditolak: guru pengganti perlu consent — bisa jadi ada alasan mereka tidak bisa (kondisi darurat, dll).

### 5. Substitusi hanya per hari (bukan per slot)
- Ditolak: guru bisa absen setengah hari, atau hanya 1-2 jam. Per-slot lebih granular dan akurat.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `DutyScheduleDescriptor.TableName()` | -- | `"duty_schedules"` |
| U02 | `TeacherSubstitutionDescriptor.TableName()` | -- | `"teacher_substitutions"` |
| U03 | `SubstitutionLogDescriptor.TableName()` | -- | `"substitution_logs"` |
| U04 | Validate rejects invalid `duty_type` | `"piket_lab"` | Error |
| U05 | Validate rejects invalid `reason_type` | `"malas"` | Error |
| U06 | Validate rejects invalid `status` | `"waiting"` | Error |
| U07 | Validate rejects same original and substitute teacher | original=A, substitute=A | Error: must be different |
| U08 | Validate rejects invalid `day_of_week` | `0` | Error |
| U09 | Auto-suggest: prioritas 1 (same subject) | Guru mapel sama, free period | Priority = 1 |
| U10 | Auto-suggest: prioritas 2 (serumpun) | Guru mapel serumpun, free period | Priority = 2 |
| U11 | Auto-suggest: prioritas 3 (guru piket) | Guru piket hari ini | Priority = 3 |
| U12 | Auto-suggest: prioritas 4 (workload terendah) | Guru lain, free period | Priority = 4, sorted by workload |
| U13 | Auto-suggest: excludes busy teachers | Guru ada jadwal di slot yang sama | Not in suggestions |
| U14 | Auto-suggest: excludes absent teachers | Guru tidak hadir hari ini | Not in suggestions |
| U15 | Validate accepts valid duty schedule | All required fields valid | No error |

### Integration Tests -- Duty Schedules

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create duty schedule | POST /duty-schedules | 201, piket created |
| I02 | Duplicate duty rejected | Same teacher+day+type+semester | 409, unique constraint |
| I03 | Bulk create semester piket | POST /duty-schedules/bulk | 201, all schedules created |
| I04 | Get today's duty | GET /duty-schedules/today | 200, filtered by today's day_of_week |
| I05 | Carry-forward from last semester | POST /carry-forward | 201, copied to new semester |

### Integration Tests -- Substitution Lifecycle

| # | Test Case | Action | Expected |
|---|---|---|---|
| I06 | Create substitution (manual) | POST /teacher-substitutions | 201, status=pending |
| I07 | Auto-suggest returns ranked list | POST /suggest | 200, sorted by priority + workload |
| I08 | Notify substitute teacher | After create | Notification sent, notified_at set |
| I09 | Accept substitution | POST /accept | 200, status=accepted |
| I10 | Decline substitution | POST /decline with reason | 200, status=declined |
| I11 | Complete substitution | POST /complete | 200, status=completed, completed_at set |
| I12 | Cancel substitution | DELETE /teacher-substitutions/{id} | 200, status=cancelled |
| I13 | Unique per date+slot+class | Create duplicate substitution | 409, unique constraint |

### Integration Tests -- Bulk Substitution

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | Bulk create for 3-day absence | POST /bulk with start-end date | 201, substitutions for all affected slots |
| I15 | Bulk skips weekend slots | Leave covers Mon-Sun | Only Mon-Fri substitutions created |
| I16 | Bulk auto-suggests per slot | POST /bulk | Each slot gets best available substitute |

### Integration Tests -- Auto-Suggest Logic

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Suggest same-subject teacher first | Math teacher absent, another math teacher free | Priority 1 suggestion |
| I18 | Suggest serumpun if no same subject | Math teacher absent, physics teacher free | Priority 2 suggestion |
| I19 | Suggest piket if no serumpun | No subject match, but piket teacher exists | Priority 3 suggestion |
| I20 | Sort by workload for same priority | Two teachers same priority | Lower workload first |
| I21 | Empty suggestions if all busy | No teachers available | Empty list with message |
| I22 | Exclude already-substituting teachers | Teacher already substituting at same slot | Not in suggestions |

### Integration Tests -- S026/S030 Integration

| # | Test Case | Action | Expected |
|---|---|---|---|
| I23 | Leave approval triggers substitution creation | S030 leave approved | Substitutions auto-created for affected slots |
| I24 | Absent teacher detected creates alert | S026 attendance status=absent | Admin notified, suggest ready |

### Integration Tests -- SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I25 | TeacherUpdated syncs to duty_schedules | Update teacher full_name | `_data.teacher.full_name` updated |
| I26 | TeacherUpdated syncs to substitutions | Update teacher full_name | `_data.original_teacher.full_name` updated |
| I27 | SubjectUpdated syncs to substitutions | Update subject name | `_data.original_subject.name` updated |

### Integration Tests -- Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I28 | Cannot access other tenant's duty | GET with wrong tenant | 404 |
| I29 | Cannot create substitution in other tenant | POST with wrong tenant | Error, scope violation |
| I30 | Cannot suggest teachers from other company | POST /suggest with wrong company | Only same-company teachers |
