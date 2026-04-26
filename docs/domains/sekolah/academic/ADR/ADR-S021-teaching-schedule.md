# ADR-S021: Teaching Schedule / Timetable (Jadwal Pelajaran)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Jadwal pelajaran adalah **operasi mingguan** yang mengatur kapan mata pelajaran diajarkan, oleh siapa, dan di mana. Jadwal harus bebas konflik — guru tidak boleh mengajar 2 kelas bersamaan, dan ruangan tidak boleh dipakai 2 kelas bersamaan.

Karakteristik jadwal sekolah di Indonesia:

1. **Weekly cycle**: Jadwal berulang setiap minggu (Senin-Sabtu/Senin-Jumat) per semester.
2. **Time slots**: Jam pelajaran terbagi dalam slot (jam ke-1, ke-2, dst.) — setiap slot ~40-45 menit.
3. **Per kelas per semester**: Setiap kelas punya jadwal berbeda.
4. **Istirahat**: Ada slot istirahat yang bukan jam pelajaran.
5. **Pesantren khusus**: Jadwal bisa mencakup pagi (diniyah) + siang (umum) + malam (tahfidz/kitab kuning).
6. **Perubahan insidental**: Guru berhalangan → perlu substitusi (guru pengganti).
7. **Input dari S020**: Jam pelajaran per minggu dari subject_configurations menentukan berapa slot yang harus dialokasikan per mapel.

Constraint utama:
- **Teacher conflict**: Guru tidak boleh dijadwalkan di 2 kelas pada waktu yang sama.
- **Room conflict**: Ruangan tidak boleh dijadwalkan untuk 2 kelas pada waktu yang sama.
- **Credit hours**: Total slot per mapel per minggu harus sesuai `credit_hours_per_week` dari S020.

### Mengapa Vernon Pattern?

- Read-heavy: jadwal dibaca oleh guru, siswa, dan admin setiap hari.
- Relasi ke teacher (ADR-012), subject (S020), class_room (ADR-011), academic_year (ADR-010).
- Volume moderate: ~30-40 slot per kelas per minggu × 20 kelas = ~600-800 records per semester.
- Business logic utama (conflict detection) di service layer, bukan di schema.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `time_slots` (definisi slot waktu) dan `schedule_entries` (jadwal mapel per slot per kelas).

### Table Schema

```sql
-- Definisi slot waktu (template jam pelajaran)
CREATE TABLE time_slots (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(30) NOT NULL,
    slot_number     INT NOT NULL,
    slot_type       VARCHAR(20) NOT NULL,

    -- Waktu
    day_of_week     INT NOT NULL,
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,

    -- Scope
    academic_year_id UUID NOT NULL,
    semester        VARCHAR(10) NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_time_slot UNIQUE (tenant_id, company_id, academic_year_id, semester, day_of_week, slot_number),
    CONSTRAINT chk_slot_type CHECK (slot_type IN ('lesson', 'break', 'assembly', 'prayer')),
    CONSTRAINT chk_day_of_week CHECK (day_of_week >= 1 AND day_of_week <= 7),
    CONSTRAINT chk_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_slot_time CHECK (start_time < end_time),
    CONSTRAINT chk_slot_number CHECK (slot_number >= 1 AND slot_number <= 15)
);

-- Indexes
CREATE INDEX idx_timeslot_tenant_company ON time_slots (tenant_id, company_id);
CREATE INDEX idx_timeslot_year_semester ON time_slots (academic_year_id, semester);
CREATE INDEX idx_timeslot_day ON time_slots (day_of_week);
CREATE INDEX idx_timeslot_rels ON time_slots USING GIN (_rels);
CREATE INDEX idx_timeslot_data ON time_slots USING GIN (_data);

-- Jadwal pelajaran per kelas per slot
CREATE TABLE schedule_entries (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    time_slot_id    UUID NOT NULL,
    class_room_id   UUID NOT NULL,
    subject_id      UUID NOT NULL,
    teacher_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Lokasi
    room_name       VARCHAR(50),

    -- Semester
    semester        VARCHAR(10) NOT NULL,

    -- Substitusi
    is_substitution     BOOLEAN NOT NULL DEFAULT false,
    original_teacher_id UUID,
    substitution_date   DATE,
    substitution_reason TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_schedule_slot_class UNIQUE (time_slot_id, class_room_id) WHERE is_substitution = false,
    CONSTRAINT chk_schedule_semester CHECK (semester IN ('ganjil', 'genap'))
);

-- Conflict detection indexes
CREATE UNIQUE INDEX uq_schedule_teacher_slot
    ON schedule_entries (teacher_id, time_slot_id)
    WHERE is_substitution = false AND deleted_at IS NULL;

CREATE UNIQUE INDEX uq_schedule_room_slot
    ON schedule_entries (room_name, time_slot_id)
    WHERE room_name IS NOT NULL AND is_substitution = false AND deleted_at IS NULL;

-- General indexes
CREATE INDEX idx_schedule_tenant_company ON schedule_entries (tenant_id, company_id);
CREATE INDEX idx_schedule_class ON schedule_entries (class_room_id);
CREATE INDEX idx_schedule_teacher ON schedule_entries (teacher_id);
CREATE INDEX idx_schedule_subject ON schedule_entries (subject_id);
CREATE INDEX idx_schedule_timeslot ON schedule_entries (time_slot_id);
CREATE INDEX idx_schedule_year_semester ON schedule_entries (academic_year_id, semester);
CREATE INDEX idx_schedule_substitution ON schedule_entries (is_substitution) WHERE is_substitution = true;
CREATE INDEX idx_schedule_rels ON schedule_entries USING GIN (_rels);
CREATE INDEX idx_schedule_data ON schedule_entries USING GIN (_data);
```

### Field Design Rationale

**time_slots:**

| Field | Keputusan | Alasan |
|---|---|---|
| `slot_number` | INT, 1-15 | Nomor urut jam (jam ke-1, ke-2, dst.) — max 15 untuk pesantren yang punya sesi pagi+siang+malam |
| `slot_type` | 4 tipe | lesson (jam pelajaran), break (istirahat), assembly (upacara), prayer (sholat — pesantren) |
| `day_of_week` | INT, 1-7 | 1=Senin, 7=Minggu. Sekolah umum: 1-5/1-6. Pesantren bisa 1-7 |
| `start_time` / `end_time` | TIME | Waktu mulai/selesai slot — configurable per sekolah |
| `semester` | VARCHAR(10) | Jadwal bisa berbeda antar semester |

**schedule_entries:**

| Field | Keputusan | Alasan |
|---|---|---|
| `room_name` | VARCHAR(50), nullable | Nama ruangan — nullable jika kelas tetap di ruangan sendiri (tidak pindah) |
| `is_substitution` | BOOLEAN | True jika ini entry substitusi guru (bukan jadwal reguler) |
| `original_teacher_id` | UUID, nullable | Guru asli yang digantikan — untuk audit trail substitusi |
| `substitution_date` | DATE, nullable | Tanggal substitusi — substitusi berlaku 1 hari saja |
| `substitution_reason` | TEXT, nullable | Alasan: sakit, cuti, dinas luar, dll |

### Conflict Detection Strategy

Konflik dideteksi di **2 level**:

1. **Database level** (partial unique indexes):
   - `uq_schedule_teacher_slot`: Satu guru hanya boleh di satu kelas per slot waktu.
   - `uq_schedule_room_slot`: Satu ruangan hanya boleh dipakai satu kelas per slot waktu.

2. **Service level** (pre-insert validation):
   ```
   Sebelum INSERT schedule_entry:
   1. Check teacher conflict: SELECT WHERE teacher_id = $1 AND time_slot_id = $2
   2. Check room conflict: SELECT WHERE room_name = $1 AND time_slot_id = $2
   3. Jika conflict → return error dengan detail: "Bu Siti sudah mengajar VII-B pada jam ke-3 Senin"
   ```

### Vernon Relationships

**time_slots:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

**schedule_entries:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `time_slot` | belongs_to | **Ya** | Info slot: hari, jam, waktu |
| `class_room` | belongs_to | **Ya** | Kelas |
| `subject` | belongs_to | **Ya** | Mata pelajaran |
| `teacher` | belongs_to | **Ya** | Guru pengajar |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |

### _rels / _data Structure

**time_slots:**
```json
{
  "_rels": {
    "academic_year_id": "018f..."
  },
  "_data": {
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

**schedule_entries:**
```json
{
  "_rels": {
    "time_slot_id": "018f...",
    "class_room_id": "018f...",
    "subject_id": "018f...",
    "teacher_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "time_slot": {
      "id": "018f...",
      "name": "Jam ke-1",
      "day_of_week": 1,
      "start_time": "07:30",
      "end_time": "08:10",
      "slot_type": "lesson"
    },
    "class_room": { "id": "018f...", "name": "VII-A" },
    "subject": { "id": "018f...", "name": "Matematika", "code": "MTK" },
    "teacher": { "id": "018f...", "full_name": "Bu Siti", "nip": "198501012010012001" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

### Pesantren Schedule (ADR-009)

Pesantren memiliki 3 sesi per hari:

| Sesi | Waktu | Slot | Mapel |
|------|-------|------|-------|
| Pagi (Diniyah) | 07:00-09:00 | 1-3 | Fiqh, Nahwu, Kitab Kuning |
| Siang (Umum) | 09:30-15:00 | 4-12 | Mapel Kemendikbud |
| Malam (Tahfidz) | 19:30-21:00 | 13-15 | Tahfidz, Muroja'ah |

Ini ditangani dengan `slot_type = 'prayer'` untuk sholat antar sesi dan `slot_number` yang cukup besar (max 15).

### API Endpoints

```
# Time Slots (Template)
GET    /api/v1/time-slots?year_id={id}&semester=ganjil           — List slot waktu
POST   /api/v1/time-slots/bulk                                    — Bulk create slots (per hari)
PUT    /api/v1/time-slots/{id}                                    — Update slot
POST   /api/v1/time-slots/carry-forward                           — Salin dari semester/tahun sebelumnya

# Schedule Entries
GET    /api/v1/schedule-entries?class_id={id}&semester=ganjil     — Jadwal per kelas
GET    /api/v1/schedule-entries?teacher_id={id}&semester=ganjil   — Jadwal per guru
POST   /api/v1/schedule-entries                                   — Tambah jadwal (with conflict check)
POST   /api/v1/schedule-entries/bulk                              — Bulk create (seluruh kelas sekaligus)
PUT    /api/v1/schedule-entries/{id}                              — Update jadwal
DELETE /api/v1/schedule-entries/{id}                              — Hapus jadwal

# Conflict Check
POST   /api/v1/schedule-entries/check-conflict                    — Check tanpa insert
  Body: { teacher_id, time_slot_id, room_name }
  → { has_conflict: true, conflicts: [...] }

# Substitusi
POST   /api/v1/schedule-entries/{id}/substitute                   — Buat substitusi guru
  Body: { substitute_teacher_id, substitution_date, reason }
GET    /api/v1/schedule-entries/substitutions?date=2026-04-15     — List substitusi per tanggal

# View
GET    /api/v1/schedules/weekly?class_id={id}&semester=ganjil     — View jadwal mingguan (matrix day × slot)
GET    /api/v1/schedules/teacher-weekly?teacher_id={id}           — View jadwal guru per minggu
```

### Weekly View Response Structure

```json
{
  "class_room": { "id": "018f...", "name": "VII-A" },
  "semester": "ganjil",
  "schedule": {
    "1": [
      { "slot": 1, "time": "07:30-08:10", "subject": "Matematika", "teacher": "Bu Siti", "room": null },
      { "slot": 2, "time": "08:10-08:50", "subject": "Matematika", "teacher": "Bu Siti", "room": null },
      { "slot": 3, "time": "08:50-09:30", "subject": "B. Indonesia", "teacher": "Pak Budi", "room": null },
      { "slot": 4, "type": "break", "time": "09:30-09:45" }
    ],
    "2": []
  }
}
```

## Consequences

### Positive

- **Conflict-free**: Database-level unique indexes menjamin tidak ada double-booking guru dan ruangan.
- **Flexible slots**: Time slot configurable — mendukung pola waktu berbeda antar sekolah dan pesantren.
- **Substitusi**: Mekanisme substitusi guru dengan audit trail lengkap.
- **Weekly view**: API endpoint khusus untuk render jadwal mingguan (matrix format).
- **Pesantren support**: Slot cukup untuk 3 sesi (pagi/siang/malam) dengan tipe prayer.
- **Carry-forward**: Copy jadwal dari semester/tahun sebelumnya.

### Negative / Trade-offs

- **Volume insert**: Setup jadwal awal semester butuh bulk insert yang besar — perlu transaksi yang baik.
- **Substitusi per hari**: Substitusi hanya berlaku per hari — jika guru sakit seminggu, perlu 5-6 substitusi. Bisa ditambah "bulk substitution" nanti.
- **Room sebagai string**: `room_name` bukan FK ke tabel rooms — trade-off untuk simplicity. Bisa dievolusi ke FK jika perlu room management.
- **No auto-scheduling**: Sistem tidak auto-generate jadwal optimal — admin manual setup. Auto-scheduling adalah problem NP-hard yang di-defer.
- **Partial unique constraint**: PostgreSQL-specific syntax — portability terbatas.

## Alternatives Considered

### 1. Jadwal sebagai JSONB per kelas per hari
- Ditolak: tidak bisa detect conflict antar kelas (guru double-booking), tidak bisa query jadwal per guru.

### 2. Auto-scheduling (constraint solver)
- Deferred: terlalu kompleks untuk MVP. Manual scheduling dengan conflict detection sudah memenuhi kebutuhan dasar.

### 3. Room sebagai FK ke tabel rooms
- Deferred: untuk MVP, string room_name cukup. Room management (kapasitas, fasilitas, booking) bisa jadi domain terpisah.

### 4. Time slot tanpa tabel terpisah (waktu langsung di schedule_entries)
- Ditolak: tanpa tabel time_slots, tidak ada standarisasi waktu — setiap entry bisa punya waktu berbeda untuk "jam ke-1". Time slots sebagai template memastikan konsistensi.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `TimeSlotDescriptor.TableName()` | — | `"time_slots"` |
| U02 | `ScheduleEntryDescriptor.TableName()` | — | `"schedule_entries"` |
| U03 | Validate rejects invalid `slot_type` | `"lunch"` | Error |
| U04 | Validate rejects `day_of_week = 0` | `0` | Error |
| U05 | Validate rejects `start_time >= end_time` | 09:00 >= 08:00 | Error |
| U06 | Validate rejects invalid `semester` | `"midterm"` | Error |
| U07 | Validate accepts valid time slot | All fields valid | No error |
| U08 | Validate accepts valid schedule entry | All fields valid | No error |

### Integration Tests — Time Slots

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Bulk create time slots | POST /time-slots/bulk for Monday | 201, all slots created |
| I02 | Unique per year+semester+day+slot | Create duplicate | 409/422 |
| I03 | Slot type CHECK | INSERT with `slot_type = 'lunch'` | DB error |
| I04 | Carry-forward | POST carry-forward from ganjil to genap | 201, all slots copied |

### Integration Tests — Schedule Entries

| # | Test Case | Action | Expected |
|---|---|---|---|
| I05 | Create schedule entry | POST with valid data | 201 |
| I06 | Teacher conflict detected (DB) | INSERT same teacher + same slot | DB unique violation |
| I07 | Room conflict detected (DB) | INSERT same room + same slot | DB unique violation |
| I08 | Teacher conflict detected (API) | POST same teacher + same slot | 409, "Bu Siti sudah mengajar..." |
| I09 | Room conflict detected (API) | POST same room + same slot | 409, "Lab IPA sudah dipakai..." |
| I10 | Check conflict endpoint | POST /check-conflict | 200, `{ has_conflict: true/false }` |
| I11 | Bulk create for class | POST /schedule-entries/bulk | 201, all entries created |
| I12 | Get class weekly schedule | GET /schedules/weekly?class_id={id} | 200, matrix format |
| I13 | Get teacher weekly schedule | GET /schedules/teacher-weekly?teacher_id={id} | 200, all classes |

### Integration Tests — Substitution

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | Create substitution | POST /schedule-entries/{id}/substitute | 201, new entry with is_substitution=true |
| I15 | Substitution teacher conflict | Substitute with busy teacher | 409, conflict |
| I16 | List substitutions by date | GET /substitutions?date=2026-04-15 | 200, filtered |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | TeacherUpdated syncs to schedules | Update teacher name | `_data.teacher.full_name` updated |
| I18 | SubjectUpdated syncs to schedules | Update subject name | `_data.subject.name` updated |
| I19 | ClassRoomUpdated syncs to schedules | Update class name | `_data.class_room.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I20 | Cannot access other tenant's schedules | GET with wrong tenant | 404 |
| I21 | Cannot create schedule in other tenant's class | POST with wrong tenant | Error |
