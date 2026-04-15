# ADR-S041: Room & Facility Booking (Peminjaman Ruangan & Fasilitas)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Sekolah memiliki fasilitas bersama (shared spaces) yang digunakan oleh berbagai pihak — guru, ekskul, panitia acara, bahkan masyarakat sekitar. Tanpa sistem booking, sering terjadi:

1. **Bentrok jadwal**: 2 kegiatan booking aula di waktu yang sama.
2. **Booking lisan**: Permintaan lisan ke bagian TU yang sering terlewat.
3. **Tidak terdokumentasi**: Penggunaan fasilitas tidak tercatat — sulit untuk pelaporan.
4. **External booking**: Sekolah meminjamkan aula/lapangan ke pihak luar tanpa tracking.

Fasilitas yang bisa di-booking:
- **Aula**: Untuk rapat, seminar, pentas seni, wisuda.
- **Lapangan**: Olahraga, upacara, event outdoor.
- **Musholla/Masjid**: Kegiatan keagamaan tambahan.
- **Ruang meeting**: Rapat guru, rapat komite, pertemuan orang tua.
- **Lab** (di luar jadwal): Lab komputer untuk ANBK, lab IPA untuk lomba (jika tidak di jadwal S021).
- **Ruang serbaguna**: Multi-purpose room.

Integrasi kunci:
- **S021 (Timetable)**: Booking tidak boleh bentrok dengan jadwal reguler yang sudah ada.
- **S015 (Extracurricular)**: Ekskul mingguan bisa di-booking sebagai recurring booking.
- **S039 (Laboratory)**: Lab yang sudah dijadwalkan di S021 otomatis tidak available.

Workflow booking di sekolah Indonesia:
1. **Request**: Guru/panitia mengajukan peminjaman ruangan.
2. **Approval**: Admin TU atau Wakasek Sarana Prasarana menyetujui/menolak.
3. **Confirmed**: Booking dikonfirmasi dan muncul di kalender.
4. **In-use → Completed**: Saat hari H, fasilitas digunakan dan dicatat selesai.
5. **Cancelled**: Dibatalkan sebelum hari H.

### Mengapa Vernon Pattern?

- Booking data = moderate volume, read-heavy untuk calendar view.
- Relasi ke teacher (ADR-012), extracurricular (S015), laboratory (S039).
- Business logic utama: conflict detection terhadap S021 + booking lain.
- Eventually consistent acceptable — booking approval bukan real-time critical.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `facilities` (master fasilitas yang bisa di-booking) dan `facility_bookings` (transaksi peminjaman).

### Table Schema

```sql
-- Master fasilitas yang bisa di-booking
CREATE TABLE facilities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    facility_type   VARCHAR(20) NOT NULL,
    description     TEXT,

    -- Lokasi
    building        VARCHAR(100),
    floor           INT,
    room_number     VARCHAR(20),

    -- Kapasitas
    capacity        INT,

    -- Fasilitas tersedia
    amenities       JSONB NOT NULL DEFAULT '[]',

    -- Aturan booking
    requires_approval    BOOLEAN NOT NULL DEFAULT true,
    max_booking_days     INT NOT NULL DEFAULT 1,
    min_advance_hours    INT NOT NULL DEFAULT 24,
    max_advance_days     INT NOT NULL DEFAULT 30,
    allow_external       BOOLEAN NOT NULL DEFAULT false,
    allow_recurring      BOOLEAN NOT NULL DEFAULT true,

    -- Pengelola (approver default)
    managed_by      UUID,

    -- Referensi ke lab (jika fasilitas ini adalah lab)
    laboratory_id   UUID,

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

    CONSTRAINT uq_facility_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_facility_type CHECK (facility_type IN (
        'aula', 'lapangan', 'musholla', 'meeting_room',
        'lab', 'ruang_serbaguna', 'studio', 'other'
    )),
    CONSTRAINT chk_facility_capacity CHECK (capacity IS NULL OR capacity >= 1),
    CONSTRAINT chk_facility_max_booking CHECK (max_booking_days >= 1),
    CONSTRAINT chk_facility_min_advance CHECK (min_advance_hours >= 0),
    CONSTRAINT chk_facility_max_advance CHECK (max_advance_days >= 1)
);

-- Indexes
CREATE INDEX idx_facility_tenant_company ON facilities (tenant_id, company_id);
CREATE INDEX idx_facility_type ON facilities (facility_type);
CREATE INDEX idx_facility_active ON facilities (is_active) WHERE is_active = true;
CREATE INDEX idx_facility_lab ON facilities (laboratory_id) WHERE laboratory_id IS NOT NULL;
CREATE INDEX idx_facility_managed_by ON facilities (managed_by) WHERE managed_by IS NOT NULL;
CREATE INDEX idx_facility_rels ON facilities USING GIN (_rels);
CREATE INDEX idx_facility_data ON facilities USING GIN (_data);

-- Transaksi peminjaman fasilitas
CREATE TABLE facility_bookings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    facility_id     UUID NOT NULL,
    academic_year_id UUID,

    -- Pemohon
    requester_type  VARCHAR(20) NOT NULL,
    requester_id    UUID NOT NULL,
    requester_name  VARCHAR(200) NOT NULL,
    organization    VARCHAR(200),

    -- Waktu booking
    booking_date    DATE NOT NULL,
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,

    -- Detail kegiatan
    event_name      VARCHAR(300) NOT NULL,
    event_description TEXT,
    expected_attendees INT,

    -- Recurring
    is_recurring    BOOLEAN NOT NULL DEFAULT false,
    recurrence_pattern VARCHAR(20),
    recurrence_end_date DATE,
    parent_booking_id UUID,

    -- Approval
    booking_status  VARCHAR(20) NOT NULL DEFAULT 'pending',
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,
    rejection_reason TEXT,

    -- Referensi ekskul (jika booking untuk ekskul)
    extracurricular_id UUID,

    -- Catatan
    notes           TEXT,
    special_requirements TEXT,

    -- Realisasi
    actual_start_time TIME,
    actual_end_time   TIME,
    actual_attendees  INT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_booking_requester_type CHECK (requester_type IN (
        'teacher', 'staff', 'extracurricular', 'committee', 'external'
    )),
    CONSTRAINT chk_booking_time CHECK (start_time < end_time),
    CONSTRAINT chk_booking_status CHECK (booking_status IN (
        'pending', 'approved', 'confirmed', 'in_use', 'completed', 'cancelled', 'rejected'
    )),
    CONSTRAINT chk_booking_recurrence CHECK (recurrence_pattern IS NULL OR recurrence_pattern IN (
        'daily', 'weekly', 'biweekly', 'monthly'
    )),
    CONSTRAINT chk_booking_attendees CHECK (expected_attendees IS NULL OR expected_attendees >= 1),
    CONSTRAINT chk_booking_recurring_end CHECK (
        (is_recurring = false) OR
        (is_recurring = true AND recurrence_pattern IS NOT NULL AND recurrence_end_date IS NOT NULL)
    )
);

-- Conflict detection: satu fasilitas tidak boleh double-booked di waktu yang sama
CREATE UNIQUE INDEX uq_booking_facility_time
    ON facility_bookings (facility_id, booking_date, start_time)
    WHERE booking_status NOT IN ('cancelled', 'rejected') AND deleted_at IS NULL;

-- Indexes
CREATE INDEX idx_booking_tenant_company ON facility_bookings (tenant_id, company_id);
CREATE INDEX idx_booking_facility ON facility_bookings (facility_id);
CREATE INDEX idx_booking_date ON facility_bookings (booking_date);
CREATE INDEX idx_booking_date_range ON facility_bookings (facility_id, booking_date, start_time, end_time);
CREATE INDEX idx_booking_requester ON facility_bookings (requester_type, requester_id);
CREATE INDEX idx_booking_status ON facility_bookings (booking_status);
CREATE INDEX idx_booking_pending ON facility_bookings (booking_status) WHERE booking_status = 'pending';
CREATE INDEX idx_booking_approved ON facility_bookings (booking_status, booking_date) WHERE booking_status IN ('approved', 'confirmed');
CREATE INDEX idx_booking_recurring ON facility_bookings (parent_booking_id) WHERE is_recurring = true;
CREATE INDEX idx_booking_ekskul ON facility_bookings (extracurricular_id) WHERE extracurricular_id IS NOT NULL;
CREATE INDEX idx_booking_year ON facility_bookings (academic_year_id) WHERE academic_year_id IS NOT NULL;
CREATE INDEX idx_booking_rels ON facility_bookings USING GIN (_rels);
CREATE INDEX idx_booking_data ON facility_bookings USING GIN (_data);
```

### Field Design Rationale

**facilities:**

| Field | Keputusan | Alasan |
|---|---|---|
| `facility_type` | 8 tipe | aula, lapangan, musholla, meeting_room, lab, ruang_serbaguna, studio, other — mencakup fasilitas sekolah umum |
| `requires_approval` | BOOLEAN, default true | Sebagian besar fasilitas butuh approval Wakasek Sarana. Meeting room kecil bisa direct-book |
| `max_booking_days` | INT, default 1 | Maksimal berapa hari per booking — prevent monopoli fasilitas |
| `min_advance_hours` | INT, default 24 | Minimal booking H-1 (24 jam) — bisa diubah per fasilitas |
| `max_advance_days` | INT, default 30 | Maksimal booking 30 hari ke depan — prevent booking terlalu jauh |
| `allow_external` | BOOLEAN, default false | Apakah pihak luar boleh booking — default tidak |
| `allow_recurring` | BOOLEAN, default true | Apakah boleh recurring booking (untuk ekskul mingguan) |
| `amenities` | JSONB array | Fasilitas: `["proyektor", "sound_system", "ac", "whiteboard", "wifi"]` |
| `laboratory_id` | UUID, nullable | Link ke S039 jika fasilitas ini adalah lab — untuk check conflict dengan jadwal lab |
| `managed_by` | UUID, nullable | Staff yang bertanggung jawab dan menjadi default approver |

**facility_bookings:**

| Field | Keputusan | Alasan |
|---|---|---|
| `requester_type` | 5 tipe | teacher, staff, extracurricular, committee, external — mencakup semua pemohon |
| `requester_name` | VARCHAR(200), denormalisasi | Nama pemohon — denormalisasi untuk kemudahan display (terutama external yang tidak ada di DB) |
| `booking_status` | 7 status | pending → approved → confirmed → in_use → completed / cancelled / rejected |
| `is_recurring` | BOOLEAN | Flag untuk booking berulang (ekskul mingguan, rapat rutin) |
| `recurrence_pattern` | 4 pola | daily, weekly, biweekly, monthly |
| `parent_booking_id` | UUID, nullable | Self-reference — recurring child bookings point to parent |
| `extracurricular_id` | UUID, nullable | Link ke S015 jika booking untuk kegiatan ekskul |
| `actual_start_time/end_time` | TIME, nullable | Realisasi waktu — untuk audit apakah booking sesuai rencana |

### Conflict Detection Strategy

Konflik dideteksi di **3 level**:

1. **Database level** (partial unique index):
   - `uq_booking_facility_time`: Mencegah double-booking pada fasilitas + tanggal + waktu yang sama.

2. **Service level** (pre-insert validation):
   ```
   Sebelum INSERT booking:
   1. Check overlap: SELECT WHERE facility_id = $1
      AND booking_date = $2
      AND start_time < $end_time AND end_time > $start_time
      AND booking_status NOT IN ('cancelled', 'rejected')
   2. Check against S021 timetable (jika fasilitas = lab):
      SELECT FROM schedule_entries se
      JOIN time_slots ts ON se.time_slot_id = ts.id
      WHERE ts.day_of_week = $day_of_week
      AND ts.start_time < $end_time AND ts.end_time > $start_time
      AND se.room_name = $facility_name
   3. Jika conflict → return error: "Aula sudah dibooking oleh Panitia Wisuda pada 15 April 09:00-12:00"
   ```

3. **Timetable integration** (S021):
   - Untuk fasilitas bertipe `lab` yang memiliki `laboratory_id`, sistem check jadwal di `schedule_entries` sebelum menerima booking.
   - Jadwal reguler dari S021 otomatis "memblokir" slot waktu di calendar fasilitas.

### Recurring Booking Generation (Application Logic)

```
Ketika is_recurring = true:
1. Buat parent booking (booking pertama) dengan is_recurring = true.
2. Generate child bookings sampai recurrence_end_date:
   - weekly: setiap minggu di hari yang sama
   - biweekly: setiap 2 minggu
   - monthly: setiap bulan di tanggal yang sama
3. Setiap child booking:
   - parent_booking_id = parent.id
   - booking_status = parent.booking_status (inherit approval)
   - Booking date = next occurrence
4. Conflict check per child booking — skip jika conflict (misal tanggal merah).
5. Cancel parent → cancel all future children.

Contoh: Ekskul Pramuka setiap Rabu 14:00-16:00 di Lapangan
- Parent: 2026-04-15, Rabu, 14:00-16:00
- Children: 2026-04-22, 2026-04-29, 2026-05-06, ... sampai recurrence_end_date
```

### Vernon Relationships

**facilities:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `managed_by_staff` | belongs_to | **Ya** | Nama pengelola fasilitas |
| `laboratory` | belongs_to | **Tidak** | Hanya diload saat check conflict — tidak selalu relevan |

**facility_bookings:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `facility` | belongs_to | **Ya** | Nama dan tipe fasilitas selalu ditampilkan |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran (jika di-set) |
| `extracurricular` | belongs_to | **Tidak** | Hanya diload jika requester_type = extracurricular |

### _rels / _data Structure

**facilities:**
```json
{
  "_rels": {
    "managed_by": "018f...",
    "laboratory_id": "018f..."
  },
  "_data": {
    "managed_by_staff": { "id": "018f...", "full_name": "Pak Suryadi", "position": "Kepala TU" }
  }
}
```

**facility_bookings:**
```json
{
  "_rels": {
    "facility_id": "018f...",
    "academic_year_id": "018f...",
    "requester_id": "018f...",
    "extracurricular_id": "018f...",
    "approved_by": "018f...",
    "parent_booking_id": "018f..."
  },
  "_data": {
    "facility": { "id": "018f...", "name": "Aula Utama", "code": "AULA-01", "facility_type": "aula", "capacity": 300 },
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "approved_by_staff": { "id": "018f...", "full_name": "Pak Ahmad", "position": "Wakasek Sarana" }
  }
}
```

### Calendar View Response Structure

```json
{
  "facility": { "id": "018f...", "name": "Aula Utama", "code": "AULA-01" },
  "month": "2026-04",
  "bookings": [
    {
      "date": "2026-04-15",
      "slots": [
        {
          "id": "018f...",
          "start_time": "09:00",
          "end_time": "12:00",
          "event_name": "Wisuda Kelas XII",
          "requester_name": "Panitia Wisuda",
          "status": "confirmed",
          "type": "booking"
        },
        {
          "start_time": "13:00",
          "end_time": "14:30",
          "event_name": "Matematika - X-IPA-1",
          "teacher": "Bu Siti",
          "type": "timetable"
        }
      ]
    }
  ],
  "timetable_blocks": [
    { "day_of_week": 1, "start_time": "07:30", "end_time": "09:30", "subject": "Fisika", "class": "X-IPA-2" }
  ]
}
```

### API Endpoints

```
# Facilities (Master)
GET    /api/v1/facilities                                — List fasilitas
POST   /api/v1/facilities                                — Buat fasilitas baru
GET    /api/v1/facilities/{id}                           — Detail fasilitas
PUT    /api/v1/facilities/{id}                           — Update fasilitas
DELETE /api/v1/facilities/{id}                           — Soft delete

# Availability
GET    /api/v1/facilities/{id}/availability              — Calendar availability per bulan
  Query: month=2026-04
  → Returns: booked slots + timetable blocks + free slots
GET    /api/v1/facilities/check-availability             — Check availability (pre-booking)
  Query: facility_id, date, start_time, end_time
  → Returns: { available: true/false, conflicts: [...] }

# Bookings
GET    /api/v1/facility-bookings                         — List booking (filter: facility_id, status, date range)
POST   /api/v1/facility-bookings                         — Request booking (with conflict check)
GET    /api/v1/facility-bookings/{id}                    — Detail booking
PUT    /api/v1/facility-bookings/{id}                    — Update booking
DELETE /api/v1/facility-bookings/{id}                    — Cancel booking

# Recurring
POST   /api/v1/facility-bookings/recurring               — Create recurring booking
  Body: { facility_id, start_time, end_time, recurrence_pattern, recurrence_end_date, ... }
DELETE /api/v1/facility-bookings/{parent_id}/cancel-series — Cancel semua future recurring bookings

# Approval Workflow
GET    /api/v1/facility-bookings/pending                 — List booking pending approval
PUT    /api/v1/facility-bookings/{id}/approve            — Approve booking
PUT    /api/v1/facility-bookings/{id}/reject             — Reject booking
  Body: { rejection_reason }

# Calendar
GET    /api/v1/facilities/{id}/calendar                  — Monthly calendar view (bookings + timetable)
  Query: month=2026-04
GET    /api/v1/facilities/calendar                       — All facilities calendar (dashboard)
  Query: month=2026-04

# Completion
PUT    /api/v1/facility-bookings/{id}/start              — Mark booking as in_use
PUT    /api/v1/facility-bookings/{id}/complete            — Mark booking as completed
  Body: { actual_start_time, actual_end_time, actual_attendees }

# Reports
GET    /api/v1/facilities/statistics                     — Statistik penggunaan fasilitas
GET    /api/v1/facilities/{id}/statistics                 — Statistik per fasilitas (frekuensi, top requester)
```

## Consequences

### Positive

- **Conflict-free booking**: Database-level unique index + service-level overlap check + timetable integration menjamin tidak ada double-booking.
- **Timetable integration**: Calendar view menggabungkan booking dan jadwal S021 — satu dashboard untuk melihat availability.
- **Recurring support**: Ekskul mingguan (S015) bisa di-booking sebagai recurring — auto-generate bookings per minggu.
- **Approval workflow**: Request → approve → confirm — sesuai SOP sekolah dengan approval oleh admin TU/Wakasek.
- **External booking**: Mendukung peminjaman oleh pihak luar (optional, configurable per fasilitas).
- **Flexible facility types**: 8 tipe fasilitas mencakup semua kebutuhan sekolah umum dan pesantren.
- **Realisasi tracking**: Actual time vs planned time — untuk audit dan optimasi penggunaan.

### Negative / Trade-offs

- **Unique index hanya pada start_time**: Overlap detection di service layer — edge case overlap tanpa exact same start_time harus ditangani di application code, bukan database.
- **Recurring generation batch**: Child bookings di-generate sekaligus — jika recurrence_end_date jauh (6 bulan), bisa banyak records. Mitigasi: max_advance_days membatasi.
- **Timetable check memerlukan query ke schedule_entries**: Cross-table check menambah latency — mitigasi: denormalisasi availability di cache layer.
- **No payment integration**: External booking belum terintegrasi dengan pembayaran — enhancement di masa depan.
- **No waitlist**: Jika fasilitas sudah full, tidak ada mekanisme antrian — hanya reject.
- **Approval latency**: Booking pending menunggu manual approval — bisa terlambat jika approver tidak aktif.

## Alternatives Considered

### 1. Booking sebagai extension dari schedule_entries (S021)
- Ditolak: schedule_entries dirancang untuk jadwal akademik reguler. Booking fasilitas punya workflow berbeda (approval, external, recurring) yang tidak cocok di schema S021.

### 2. Tanpa tabel facilities (booking langsung ke room_name string)
- Ditolak: fasilitas butuh konfigurasi (approval rules, capacity, amenities) yang tidak bisa disimpan di string.

### 3. Real-time locking (optimistic/pessimistic)
- Ditolak untuk MVP: booking sekolah bukan high-concurrency scenario — unique index + service check cukup. Real-time locking menambah complexity tanpa benefit signifikan.

### 4. Calendar events sebagai tabel terpisah (facility_calendar_events)
- Ditolak: calendar adalah view yang menggabungkan bookings + timetable — bukan tabel terpisah. Calendar dirender dari query gabungan di API layer.

### 5. Gabung dengan S039 (Laboratory Management)
- Ditolak: lab management punya domain spesifik (equipment, safety, usage log) yang tidak ada di facility booking. Namun lab bisa menjadi facility (via `laboratory_id`) untuk booking di luar jadwal.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `FacilityDescriptor.TableName()` | — | `"facilities"` |
| U02 | `FacilityBookingDescriptor.TableName()` | — | `"facility_bookings"` |
| U03 | Validate rejects invalid `facility_type` | `"kolam_renang"` | Error: invalid type |
| U04 | Validate rejects invalid `booking_status` | `"expired"` | Error: invalid status |
| U05 | Validate rejects invalid `requester_type` | `"student"` | Error: must be teacher/staff/extracurricular/committee/external |
| U06 | Validate rejects invalid `recurrence_pattern` | `"yearly"` | Error: must be daily/weekly/biweekly/monthly |
| U07 | Validate rejects `start_time >= end_time` | 16:00 >= 14:00 | Error |
| U08 | Validate rejects recurring without end_date | is_recurring=true, end_date=null | Error |
| U09 | Validate accepts valid facility | All fields valid | No error |
| U10 | Validate accepts valid booking | All fields valid | No error |
| U11 | Recurring generation: weekly for 4 weeks | start=Apr 15, pattern=weekly, end=May 13 | 4 child bookings generated |
| U12 | Recurring generation: skip conflict dates | 1 of 4 weeks has conflict | 3 child bookings generated, 1 skipped |

### Integration Tests — Facilities

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create facility | POST with valid data | 201 |
| I02 | Unique code per company | Create 2 facilities same code | 409/422 |
| I03 | Facility type CHECK | INSERT with `facility_type = 'kolam_renang'` | DB error |
| I04 | Get facility with amenities | GET /facilities/{id} | 200, amenities JSONB parsed |

### Integration Tests — Bookings

| # | Test Case | Action | Expected |
|---|---|---|---|
| I05 | Create booking | POST with valid data | 201, status=pending |
| I06 | Conflict detection (same time) | POST booking same facility+date+time | 409, conflict detail |
| I07 | Overlap detection (partial overlap) | POST 09:00-12:00 when 10:00-11:00 exists | 409, overlap detected |
| I08 | No conflict (adjacent times) | POST 12:00-14:00 when 09:00-12:00 exists | 201, no conflict |
| I09 | Timetable conflict (lab) | POST booking for lab during S021 schedule | 409, "Lab Kimia sudah dijadwalkan..." |
| I10 | Check availability | GET /check-availability?date=2026-04-15 | 200, { available: true/false } |

### Integration Tests — Approval Workflow

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Approve booking | PUT /approve | 200, status=approved |
| I12 | Reject booking | PUT /reject with reason | 200, status=rejected |
| I13 | Cannot approve cancelled booking | PUT /approve on cancelled | 422, invalid transition |
| I14 | List pending bookings | GET /facility-bookings/pending | 200, only pending status |

### Integration Tests — Recurring

| # | Test Case | Action | Expected |
|---|---|---|---|
| I15 | Create recurring booking (weekly) | POST recurring for 4 weeks | 201, 1 parent + 3 children |
| I16 | Cancel recurring series | DELETE /cancel-series | 200, all future children cancelled |
| I17 | Recurring with partial conflicts | Some weeks have conflict | Created with skipped dates noted |

### Integration Tests — Calendar

| # | Test Case | Action | Expected |
|---|---|---|---|
| I18 | Get facility calendar | GET /calendar?month=2026-04 | 200, bookings + timetable blocks |
| I19 | Calendar shows timetable blocks | Lab with S021 schedule | Timetable blocks appear as unavailable |
| I20 | Calendar shows booking status colors | Mixed statuses | Each slot has correct status |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I21 | FacilityUpdated syncs to bookings | Update facility name | `_data.facility.name` updated |
| I22 | StaffUpdated syncs to facilities | Update managed_by name | `_data.managed_by_staff.full_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I23 | Cannot access other tenant's facilities | GET with wrong tenant | 404 |
| I24 | Cannot book other tenant's facility | POST booking cross-tenant | Error |
| I25 | Cannot approve other tenant's booking | PUT approve cross-tenant | Error |
