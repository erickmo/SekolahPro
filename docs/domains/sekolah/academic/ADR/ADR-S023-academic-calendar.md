# ADR-S023: Academic Calendar (Kalender Akademik)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Kalender akademik adalah **jadwal tahunan** yang mendefinisikan seluruh event penting dalam satu tahun ajaran: hari libur, periode ujian, kegiatan sekolah, awal/akhir semester, dan kelulusan. Kalender ini menjadi **sumber kebenaran** untuk menentukan hari efektif belajar — data yang dibutuhkan oleh S008 (Daily Attendance) untuk menghitung persentase kehadiran.

Karakteristik kalender akademik di Indonesia:

1. **Ditetapkan per tahun ajaran** (ADR-010): Setiap tahun ajaran memiliki kalender sendiri.
2. **Hari libur nasional**: Ditetapkan pemerintah — termasuk libur keagamaan, kemerdekaan, dll.
3. **Hari libur sekolah**: Ditetapkan sekolah — libur semester, libur khusus.
4. **Periode ujian**: PTS/UTS, PAS/UAS, PAT — blok waktu yang mempengaruhi jadwal pelajaran (S021).
5. **Pesantren (ADR-009)**: Event kalender Islam — jadwal Ramadan (setengah hari/libur), Idul Fitri, Idul Adha, Maulid Nabi, Isra Mi'raj, 1 Muharram. Pesantren juga punya event khusus: Haul, Khataman, Wisuda Tahfidz.
6. **Hari efektif**: Jumlah hari sekolah aktif per semester (biasanya ~100 hari per semester) — dihitung dari total hari kerja minus libur/event non-teaching.
7. **Read-heavy**: Kalender dibaca setiap hari untuk attendance validation, dashboard, dan notifikasi orang tua.

### Mengapa Vernon Pattern?

- Read-heavy: kalender dibaca oleh semua role (guru, siswa, orang tua, admin) setiap hari.
- Relasi ke academic_year (ADR-010) — setiap event terikat ke tahun ajaran.
- Volume rendah: ~50-100 events per tahun ajaran — data sangat stabil.
- Jarang berubah setelah di-setup di awal tahun ajaran.
- SyncEngine: perubahan nama tahun ajaran harus propagate ke `_data`.

## Decision

Menggunakan **Vernon Pattern** untuk 1 tabel: `academic_calendar_events` (event kalender akademik per tahun ajaran).

### Table Schema

```sql
CREATE TABLE academic_calendar_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign key
    academic_year_id UUID NOT NULL,

    -- Identitas event
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    event_type      VARCHAR(30) NOT NULL,

    -- Periode
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    semester        VARCHAR(10),

    -- Konfigurasi
    is_school_day   BOOLEAN NOT NULL DEFAULT false,
    affects_attendance BOOLEAN NOT NULL DEFAULT true,
    is_recurring_yearly BOOLEAN NOT NULL DEFAULT false,

    -- Kalender Islam (ADR-009)
    hijri_date      VARCHAR(30),
    islamic_event_type VARCHAR(30),

    -- Metadata
    color_code      VARCHAR(7),
    sort_order      INT NOT NULL DEFAULT 0,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_event_type CHECK (event_type IN (
        'holiday', 'exam_period', 'school_event', 'semester_start', 'semester_end',
        'graduation', 'enrollment_period', 'teacher_training',
        -- Pesantren specific
        'islamic_holiday', 'pesantren_event', 'ramadan_schedule'
    )),
    CONSTRAINT chk_semester CHECK (semester IS NULL OR semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_date_range CHECK (start_date <= end_date),
    CONSTRAINT chk_islamic_event_type CHECK (islamic_event_type IS NULL OR islamic_event_type IN (
        'ramadan', 'idul_fitri', 'idul_adha', 'maulid_nabi',
        'isra_miraj', 'tahun_baru_hijriyah', 'nuzulul_quran',
        'haul', 'khataman', 'wisuda_tahfidz'
    )),
    CONSTRAINT chk_color_code CHECK (color_code IS NULL OR color_code ~ '^#[0-9A-Fa-f]{6}$')
);

-- Indexes
CREATE INDEX idx_calendar_tenant_company ON academic_calendar_events (tenant_id, company_id);
CREATE INDEX idx_calendar_year ON academic_calendar_events (academic_year_id);
CREATE INDEX idx_calendar_type ON academic_calendar_events (event_type);
CREATE INDEX idx_calendar_semester ON academic_calendar_events (semester);
CREATE INDEX idx_calendar_date_range ON academic_calendar_events (start_date, end_date);
CREATE INDEX idx_calendar_islamic ON academic_calendar_events (islamic_event_type) WHERE islamic_event_type IS NOT NULL;
CREATE INDEX idx_calendar_affects_attendance ON academic_calendar_events (affects_attendance) WHERE affects_attendance = true;
CREATE INDEX idx_calendar_rels ON academic_calendar_events USING GIN (_rels);
CREATE INDEX idx_calendar_data ON academic_calendar_events USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `event_type` | 9 tipe | 6 tipe umum (holiday, exam_period, school_event, semester_start/end, graduation) + 3 tipe khusus (enrollment_period, teacher_training, pesantren events) |
| `start_date` / `end_date` | DATE range | Banyak event multi-hari — libur semester (2 minggu), periode ujian (5-7 hari), Ramadan (~30 hari) |
| `semester` | VARCHAR(10), nullable | Nullable karena beberapa event lintas semester (e.g. libur akhir tahun, graduation) |
| `is_school_day` | BOOLEAN, default false | True hanya untuk event yang tetap hari sekolah (e.g. school_event tapi siswa tetap hadir). Default false = libur/tidak masuk |
| `affects_attendance` | BOOLEAN, default true | True = event ini dihitung dalam kalkulasi hari efektif. False = event informatif saja (e.g. reminder) |
| `is_recurring_yearly` | BOOLEAN, default false | True untuk event yang berulang setiap tahun (Kemerdekaan RI, Hari Guru) — bantu carry-forward |
| `hijri_date` | VARCHAR(30), nullable | Tanggal Hijriyah untuk event Islam — format "1 Syawal 1448 H". Nullable untuk sekolah umum |
| `islamic_event_type` | VARCHAR(30), nullable | Spesifik tipe event Islam — memudahkan filtering dan template pesantren |
| `color_code` | VARCHAR(7), nullable | Hex color (#FF0000) untuk display di kalender UI — merah libur, biru ujian, hijau kegiatan |

### Effective School Days Calculation (Integrasi S008)

```
Hari Efektif per Semester:
1. Hitung total hari kerja (Senin-Sabtu atau Senin-Jumat) dalam range semester
2. Kurangi: semua event WHERE affects_attendance = true AND is_school_day = false
3. Hasil = total hari efektif belajar

Contoh Semester Ganjil (Juli-Desember):
  Total hari kerja (Sen-Sab): 156 hari
  - Libur nasional:           12 hari
  - Libur semester:           14 hari
  - Periode ujian PTS/PAS:     0 hari (ujian tetap hari sekolah)
  = Hari efektif:            130 hari

Persentase kehadiran siswa (S008):
  = (days_present / effective_school_days) × 100%
```

Event dengan `is_school_day = true` (misal: upacara hari kemerdekaan) tetap dihitung sebagai hari efektif — siswa wajib hadir.

### Ramadan Schedule (ADR-009 Pesantren)

Pesantren memiliki jadwal khusus selama Ramadan:

| Event | Durasi | is_school_day | affects_attendance | islamic_event_type |
|-------|--------|---------------|--------------------|--------------------|
| Ramadan (jadwal pendek) | ~30 hari | true | true | ramadan |
| Libur Idul Fitri | 7-14 hari | false | true | idul_fitri |
| Nuzulul Quran (17 Ramadan) | 1 hari | false | true | nuzulul_quran |

Selama Ramadan, pesantren biasanya:
- Jam pelajaran diperpendek (30 menit per slot, bukan 40-45 menit)
- Tambah sesi tadarus/tarawih malam
- Event ini ditandai `islamic_event_type = 'ramadan'` agar sistem bisa adjust jadwal (S021)

### Vernon Relationships

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran — nama dan periode |

### _rels / _data Structure

```json
{
  "_rels": {
    "academic_year_id": "018f..."
  },
  "_data": {
    "academic_year": { "id": "018f...", "name": "2025/2026", "start_date": "2025-07-14", "end_date": "2026-06-20" }
  }
}
```

### API Endpoints

```
# Calendar Events
GET    /api/v1/calendar-events?year_id={id}                          — Semua event per tahun ajaran
GET    /api/v1/calendar-events?year_id={id}&semester=ganjil           — Event per semester
GET    /api/v1/calendar-events?year_id={id}&type=holiday              — Filter by event type
GET    /api/v1/calendar-events?year_id={id}&type=islamic_holiday      — Event Islam (pesantren)
GET    /api/v1/calendar-events/{id}                                   — Detail event
POST   /api/v1/calendar-events                                       — Buat event
POST   /api/v1/calendar-events/bulk                                  — Bulk create (e.g. import libur nasional)
PUT    /api/v1/calendar-events/{id}                                  — Update event
DELETE /api/v1/calendar-events/{id}                                  — Hapus event

# Carry-forward
POST   /api/v1/calendar-events/carry-forward                         — Salin event recurring dari tahun sebelumnya
  Body: { "source_academic_year_id": "018f...", "target_academic_year_id": "018f..." }
  → Copy events WHERE is_recurring_yearly = true, adjust dates +1 year

# Effective Days Calculation
GET    /api/v1/calendar-events/effective-days?year_id={id}&semester=ganjil
  → { "total_work_days": 156, "holiday_days": 26, "effective_days": 130 }

# Monthly View
GET    /api/v1/calendar-events/monthly?year_id={id}&month=2026-01
  → Events for January 2026 with daily breakdown

# Today's Status
GET    /api/v1/calendar-events/today?year_id={id}
  → { "date": "2026-04-15", "is_school_day": true, "events": [...] }
```

## Consequences

### Positive

- **Sumber kebenaran hari efektif**: S008 attendance calculation mengacu ke kalender — bukan hardcoded.
- **Multi-mode**: Mendukung sekolah umum dan pesantren (ADR-009) dengan event Islam lengkap.
- **Read-cache optimal**: Vernon pattern cocok karena data sangat stabil — setup sekali per tahun, baca setiap hari.
- **Carry-forward**: Event recurring (libur nasional, hari guru) bisa di-copy ke tahun ajaran baru.
- **Visual calendar**: `color_code` per event type memudahkan rendering kalender di frontend.
- **Ramadan-aware**: Sistem bisa mendeteksi periode Ramadan untuk adjust jadwal dan attendance.

### Negative / Trade-offs

- **Hijri date as string**: `hijri_date` disimpan sebagai string, bukan tipe data khusus — tidak bisa di-query range. Trade-off karena PostgreSQL tidak punya native Hijri type.
- **No time-of-day**: Event hanya DATE, bukan TIMESTAMP — tidak bisa specify jam mulai event (e.g. upacara jam 07:00). Cukup untuk kalender akademik yang beroperasi per hari.
- **Islamic calendar drift**: Tanggal Hijriyah bergeser ~11 hari per tahun Masehi — carry-forward untuk event Islam butuh adjustment manual oleh admin.
- **Single table**: Semua tipe event di satu tabel — jika ada tipe event baru, perlu ALTER TABLE untuk CHECK constraint.
- **Effective days calculation di API**: Perhitungan hari efektif dilakukan runtime, bukan pre-computed. Acceptable karena volume data rendah (~100 events per tahun).

## Alternatives Considered

### 1. Kalender sebagai JSONB per tahun ajaran
- Ditolak: tidak bisa query per event type, tidak bisa filter by date range, tidak bisa GIN index per field.

### 2. Tabel terpisah untuk holiday vs exam_period vs school_event
- Ditolak: struktur data sama (nama, tanggal, tipe) — satu tabel dengan `event_type` lebih sederhana dan queryable.

### 3. Integrasi Google Calendar / iCal
- Deferred: untuk MVP, internal calendar cukup. Export ke iCal (.ics) bisa ditambah sebagai enhancement.

### 4. Hijri date sebagai computed column
- Deferred: konversi Masehi-Hijriyah memerlukan library khusus. Untuk MVP, admin input manual `hijri_date` string sudah cukup.

### 5. Pre-computed effective days di tabel terpisah
- Deferred: volume data rendah (~100 events) — runtime calculation cukup cepat. Jika performa jadi masalah, bisa tambah materialized view.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `CalendarEventDescriptor.TableName()` | — | `"academic_calendar_events"` |
| U02 | Validate rejects invalid `event_type` | `"party"` | Error |
| U03 | Validate rejects invalid `islamic_event_type` | `"christmas"` | Error |
| U04 | Validate rejects `start_date > end_date` | 2026-01-15 > 2026-01-10 | Error |
| U05 | Validate rejects invalid `semester` | `"summer"` | Error |
| U06 | Validate rejects invalid `color_code` | `"red"` | Error |
| U07 | Validate accepts valid color code | `"#FF0000"` | No error |
| U08 | Effective days calculation | 156 work days - 26 holidays | 130 effective days |
| U09 | Validate accepts valid event | All fields valid | No error |
| U10 | Validate accepts null semester (cross-semester event) | semester = null | No error |

### Integration Tests — Calendar Events

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create calendar event | POST with valid data | 201 |
| I02 | Event type CHECK | INSERT with `event_type = 'party'` | DB error |
| I03 | Islamic event type CHECK | INSERT with `islamic_event_type = 'christmas'` | DB error |
| I04 | Date range CHECK | INSERT with `start_date > end_date` | DB error |
| I05 | Color code regex CHECK | INSERT with `color_code = 'red'` | DB error |
| I06 | Bulk create holidays | POST /calendar-events/bulk with 12 libur nasional | 201, 12 records |
| I07 | Filter by event type | GET /calendar-events?type=holiday | 200, only holidays |
| I08 | Filter by semester | GET /calendar-events?semester=ganjil | 200, semester 1 only |

### Integration Tests — Islamic Events (ADR-009)

| # | Test Case | Action | Expected |
|---|---|---|---|
| I09 | Create Ramadan event | POST with type=ramadan_schedule, islamic_event_type=ramadan | 201 |
| I10 | Create Idul Fitri | POST with type=islamic_holiday, 7 hari libur | 201 |
| I11 | Filter Islamic events | GET /calendar-events?type=islamic_holiday | 200, Islamic events only |
| I12 | Hijri date stored | POST with hijri_date="1 Syawal 1448 H" | 201, hijri_date persisted |

### Integration Tests — Effective Days

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | Calculate effective days | GET /effective-days?semester=ganjil | 200, correct count |
| I14 | Holiday reduces effective days | Add holiday, recalculate | Count decreases |
| I15 | School event (is_school_day=true) | Add school_event with is_school_day=true | Effective days unchanged |
| I16 | Non-affecting event | Add event with affects_attendance=false | Effective days unchanged |

### Integration Tests — Carry-forward

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Carry-forward recurring events | POST carry-forward | 201, recurring events copied |
| I18 | Non-recurring events skipped | POST carry-forward | Only is_recurring_yearly=true copied |
| I19 | Dates adjusted +1 year | Carry-forward 2025/2026 → 2026/2027 | Dates shifted by ~1 year |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I20 | AcademicYearUpdated syncs to events | Update year name | `_data.academic_year.name` updated |
| I21 | Sync version incremented | Update event, trigger sync | `_sync_version` incremented |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I22 | Cannot access other tenant's events | GET with wrong tenant | 404 |
| I23 | Cannot create event in other tenant's year | POST with wrong tenant | Error |
