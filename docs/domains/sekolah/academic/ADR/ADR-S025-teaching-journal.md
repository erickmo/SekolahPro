# ADR-S025: Teaching Journal (Jurnal Mengajar)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Jurnal mengajar adalah **log harian aktivitas pengajaran** yang dicatat oleh guru setelah setiap sesi mengajar. Jurnal ini berfungsi sebagai:

1. **Dokumentasi realisasi mengajar**: Bukti bahwa guru benar-benar mengajar sesuai RPP (S024).
2. **Catatan materi**: Materi apa yang diajarkan hari itu — berguna untuk guru pengganti dan tracking progress.
3. **Absensi per sesi**: Catatan kehadiran siswa **per mata pelajaran** (berbeda dari S008 yang absensi harian umum).
4. **Refleksi guru**: Catatan tentang respon siswa, kendala, dan penyesuaian yang diperlukan.
5. **Monitoring kepala sekolah**: Supervisi pengajaran — apakah guru mengajar sesuai jadwal dan RPP.

Jurnal mengajar di Indonesia:

1. **Wajib diisi**: Permendikbud mengharuskan guru mengisi jurnal mengajar — menjadi salah satu komponen penilaian kinerja guru.
2. **Per sesi per kelas**: Satu entry untuk setiap kali guru masuk kelas mengajar.
3. **Linked ke jadwal (S021)**: Jurnal terikat ke jadwal pelajaran — jam ke berapa, hari apa, kelas mana.
4. **Linked ke RPP (S024)**: Realisasi dari rencana pembelajaran yang sudah dibuat.
5. **Pesantren (ADR-009)**: Jurnal untuk mapel diniyah termasuk catatan metode (bandongan/sorogan), kitab yang dibaca, dan progress hafalan santri.

Karakteristik operasional:
- **Write-heavy harian**: Guru menulis jurnal setiap hari setelah mengajar (2-6 sesi per hari per guru).
- **Read-heavy bulanan**: Rekap jurnal untuk laporan bulanan/semester.
- **Volume tinggi**: 1 guru × 5 sesi/hari × 25 hari/bulan = 125 entries/bulan per guru.

### Mengapa Vernon Pattern?

- Relasi ke teacher (ADR-012), subject (S020), class_room (ADR-011), schedule_entry (S021), lesson_plan (S024), academic_year (ADR-010).
- Read-heavy: rekap bulanan, monitoring kepala sekolah, laporan.
- Volume tinggi tapi predictable — bisa partitioned by month jika perlu.
- SyncEngine: perubahan nama guru/mapel/kelas harus propagate ke `_data`.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `teaching_journals` (jurnal mengajar per sesi) dan `journal_session_attendances` (absensi siswa per sesi mengajar).

### Table Schema

```sql
-- Jurnal mengajar per sesi
CREATE TABLE teaching_journals (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    subject_id      UUID NOT NULL,
    class_room_id   UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    schedule_entry_id UUID,
    lesson_plan_id  UUID,

    -- Identitas sesi
    journal_date    DATE NOT NULL,
    semester        VARCHAR(10) NOT NULL,
    slot_start      INT NOT NULL,
    slot_end        INT NOT NULL,
    start_time      TIME,
    end_time        TIME,

    -- Konten jurnal
    topic_taught    VARCHAR(300) NOT NULL,
    material_detail TEXT,
    teaching_method VARCHAR(50),
    learning_activities_summary TEXT,
    student_responses TEXT,
    obstacles_notes TEXT,
    follow_up_plan  TEXT,

    -- Pesantren spesifik (ADR-009)
    kitab_reference VARCHAR(200),
    kitab_page_from INT,
    kitab_page_to   INT,
    teaching_method_pesantren VARCHAR(30),
    hafalan_progress TEXT,

    -- Statistik kehadiran sesi (denormalized)
    total_students  INT NOT NULL DEFAULT 0,
    present_count   INT NOT NULL DEFAULT 0,
    absent_count    INT NOT NULL DEFAULT 0,
    late_count      INT NOT NULL DEFAULT 0,
    permission_count INT NOT NULL DEFAULT 0,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_journal_teacher_class_date_slot UNIQUE (teacher_id, class_room_id, journal_date, slot_start),
    CONSTRAINT chk_journal_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_journal_slot CHECK (slot_start >= 1 AND slot_start <= 15 AND slot_end >= slot_start AND slot_end <= 15),
    CONSTRAINT chk_journal_time CHECK (start_time IS NULL OR end_time IS NULL OR start_time < end_time),
    CONSTRAINT chk_journal_status CHECK (status IN ('draft', 'submitted', 'verified')),
    CONSTRAINT chk_attendance_counts CHECK (
        present_count >= 0 AND absent_count >= 0 AND late_count >= 0 AND permission_count >= 0
        AND total_students >= 0
    ),
    CONSTRAINT chk_pesantren_method CHECK (teaching_method_pesantren IS NULL OR teaching_method_pesantren IN (
        'bandongan', 'sorogan', 'halaqah', 'muhafadzah',
        'mudzakarah', 'ceramah', 'demonstrasi', 'tanya_jawab'
    )),
    CONSTRAINT chk_kitab_pages CHECK (
        kitab_page_from IS NULL OR kitab_page_to IS NULL
        OR (kitab_page_from > 0 AND kitab_page_to >= kitab_page_from)
    )
);

-- Indexes
CREATE INDEX idx_journal_tenant_company ON teaching_journals (tenant_id, company_id);
CREATE INDEX idx_journal_teacher ON teaching_journals (teacher_id);
CREATE INDEX idx_journal_subject ON teaching_journals (subject_id);
CREATE INDEX idx_journal_class ON teaching_journals (class_room_id);
CREATE INDEX idx_journal_date ON teaching_journals (journal_date);
CREATE INDEX idx_journal_year_semester ON teaching_journals (academic_year_id, semester);
CREATE INDEX idx_journal_schedule ON teaching_journals (schedule_entry_id) WHERE schedule_entry_id IS NOT NULL;
CREATE INDEX idx_journal_lesson_plan ON teaching_journals (lesson_plan_id) WHERE lesson_plan_id IS NOT NULL;
CREATE INDEX idx_journal_status ON teaching_journals (status);
CREATE INDEX idx_journal_teacher_date ON teaching_journals (teacher_id, journal_date);
CREATE INDEX idx_journal_rels ON teaching_journals USING GIN (_rels);
CREATE INDEX idx_journal_data ON teaching_journals USING GIN (_data);

-- Absensi siswa per sesi mengajar
CREATE TABLE journal_session_attendances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    journal_id      UUID NOT NULL,
    student_id      UUID NOT NULL,

    -- Status kehadiran
    attendance_status VARCHAR(15) NOT NULL,
    late_minutes    INT,
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_session_attendance UNIQUE (journal_id, student_id),
    CONSTRAINT chk_attendance_status CHECK (attendance_status IN (
        'present', 'absent', 'late', 'sick', 'permitted', 'dispensasi'
    )),
    CONSTRAINT chk_late_minutes CHECK (late_minutes IS NULL OR (late_minutes > 0 AND late_minutes <= 120))
);

-- Indexes
CREATE INDEX idx_session_att_tenant_company ON journal_session_attendances (tenant_id, company_id);
CREATE INDEX idx_session_att_journal ON journal_session_attendances (journal_id);
CREATE INDEX idx_session_att_student ON journal_session_attendances (student_id);
CREATE INDEX idx_session_att_status ON journal_session_attendances (attendance_status);
CREATE INDEX idx_session_att_rels ON journal_session_attendances USING GIN (_rels);
CREATE INDEX idx_session_att_data ON journal_session_attendances USING GIN (_data);
```

### Field Design Rationale

**teaching_journals:**

| Field | Keputusan | Alasan |
|---|---|---|
| `schedule_entry_id` | UUID, nullable | Link ke jadwal (S021) — nullable jika guru mengajar di luar jadwal (pengganti, tambahan) |
| `lesson_plan_id` | UUID, nullable | Link ke RPP (S024) — nullable jika belum ada RPP atau mengajar spontan |
| `slot_start` / `slot_end` | INT, 1-15 | Jam ke berapa — range karena satu sesi bisa 2-3 slot (e.g. jam ke-1 sampai ke-3) |
| `start_time` / `end_time` | TIME, nullable | Waktu aktual — nullable karena bisa di-derive dari time_slots (S021). Berguna jika waktu aktual berbeda dari jadwal |
| `topic_taught` | VARCHAR(300), NOT NULL | Materi yang benar-benar diajarkan — bisa berbeda dari RPP |
| `material_detail` | TEXT, nullable | Detail materi — untuk catatan lebih lengkap |
| `teaching_method` | VARCHAR(50), nullable | Metode umum: diskusi, ceramah, praktikum, project, presentasi, dll. Free text karena sangat bervariasi |
| `student_responses` | TEXT, nullable | Catatan respon siswa — antusiasme, kesulitan, pertanyaan menarik |
| `obstacles_notes` | TEXT, nullable | Kendala yang dihadapi — LCD rusak, siswa ribut, materi terlalu sulit, dll. |
| `follow_up_plan` | TEXT, nullable | Rencana tindak lanjut — remedial, pengulangan materi, PR tambahan |
| `kitab_reference` + `kitab_page_*` | Pesantren fields | Kitab yang dibaca + halaman berapa sampai berapa — untuk tracking progress kitab kuning |
| `hafalan_progress` | TEXT, nullable | Catatan progress hafalan kelas — siapa yang sudah setor, ayat berapa |
| `total_students` / `*_count` | INT, denormalized | Statistik kehadiran sesi — denormalized dari journal_session_attendances untuk quick display |
| `status` | 3 stage | draft (belum lengkap), submitted (selesai diisi), verified (sudah diperiksa kepsek) |

**journal_session_attendances:**

| Field | Keputusan | Alasan |
|---|---|---|
| `attendance_status` | 6 status | present (hadir), absent (alpa), late (terlambat), sick (sakit), permitted (izin), dispensasi (tugas sekolah) |
| `late_minutes` | INT, nullable, max 120 | Menit keterlambatan — hanya relevan jika status = late |
| `notes` | TEXT, nullable | Catatan per siswa: "Izin ke UKS jam 09:30", "Sakit perut setelah istirahat", dll. |

### Perbedaan dengan S008 (Daily Attendance)

| Aspek | S008 Daily Attendance | S025 Session Attendance |
|-------|----------------------|------------------------|
| **Granularity** | Per hari per siswa | Per sesi mengajar per siswa |
| **Pencatat** | Wali kelas | Guru mata pelajaran |
| **Waktu** | Pagi hari (sekali) | Setiap sesi mengajar (2-6x/hari) |
| **Tujuan** | Rekap kehadiran semester | Monitoring kehadiran per mapel |
| **Contoh** | Ahmad: Hadir hari ini | Ahmad: Hadir di Matematika jam ke-1, Absent di IPA jam ke-5 |

Siswa bisa **hadir di S008** (datang ke sekolah) tapi **absent di session attendance** (bolos mata pelajaran tertentu).

### Vernon Relationships

**teaching_journals:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `teacher` | belongs_to | **Ya** | Nama guru |
| `subject` | belongs_to | **Ya** | Mata pelajaran |
| `class_room` | belongs_to | **Ya** | Kelas |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |
| `schedule_entry` | belongs_to | **Tidak** | Load on demand — detail jadwal |
| `lesson_plan` | belongs_to | **Tidak** | Load on demand — detail RPP |

**journal_session_attendances:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `journal` | belongs_to | **Ya** | Konteks jurnal (tanggal, mapel, kelas) |
| `student` | belongs_to | **Ya** | Nama dan NIS siswa |

### _rels / _data Structure

**teaching_journals:**
```json
{
  "_rels": {
    "teacher_id": "018f...",
    "subject_id": "018f...",
    "class_room_id": "018f...",
    "academic_year_id": "018f...",
    "schedule_entry_id": "018f...",
    "lesson_plan_id": "018f..."
  },
  "_data": {
    "teacher": { "id": "018f...", "full_name": "Bu Siti", "nip": "198501012010012001" },
    "subject": { "id": "018f...", "name": "Matematika", "code": "MTK" },
    "class_room": { "id": "018f...", "name": "VII-A" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

**journal_session_attendances:**
```json
{
  "_rels": {
    "journal_id": "018f...",
    "student_id": "018f..."
  },
  "_data": {
    "journal": {
      "id": "018f...",
      "journal_date": "2026-04-15",
      "subject_name": "Matematika",
      "class_name": "VII-A",
      "teacher_name": "Bu Siti",
      "slot_start": 1,
      "slot_end": 2
    },
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" }
  }
}
```

### API Endpoints

```
# Teaching Journals
GET    /api/v1/teaching-journals?teacher_id={id}&date=2026-04-15           — Jurnal guru per tanggal
GET    /api/v1/teaching-journals?teacher_id={id}&month=2026-04             — Jurnal guru per bulan (rekap)
GET    /api/v1/teaching-journals?class_id={id}&date=2026-04-15             — Jurnal per kelas per hari
GET    /api/v1/teaching-journals?subject_id={id}&class_id={id}&semester=ganjil — Jurnal per mapel per kelas
GET    /api/v1/teaching-journals/{id}                                       — Detail jurnal
POST   /api/v1/teaching-journals                                            — Buat jurnal
PUT    /api/v1/teaching-journals/{id}                                       — Update jurnal (draft/submitted)
DELETE /api/v1/teaching-journals/{id}                                       — Hapus jurnal (draft only)

# Status
POST   /api/v1/teaching-journals/{id}/submit                                — Submit jurnal
POST   /api/v1/teaching-journals/{id}/verify                                — Kepsek verify

# Quick Create from Schedule
POST   /api/v1/teaching-journals/from-schedule                              — Buat jurnal dari jadwal hari ini
  Body: { "schedule_entry_id": "018f...", "journal_date": "2026-04-15" }
  → Auto-fill teacher, subject, class, slot dari schedule_entry

# Session Attendance (Bulk)
POST   /api/v1/teaching-journals/{id}/attendances/bulk                     — Bulk input absensi sesi
  Body: {
    "attendances": [
      { "student_id": "018f...", "attendance_status": "present" },
      { "student_id": "018f...", "attendance_status": "late", "late_minutes": 15, "notes": "Dari UKS" },
      { "student_id": "018f...", "attendance_status": "absent" }
    ]
  }
GET    /api/v1/teaching-journals/{id}/attendances                          — List absensi sesi
PUT    /api/v1/journal-session-attendances/{id}                            — Update individual

# Student Session Attendance View
GET    /api/v1/students/{id}/session-attendances?subject_id={id}&semester=ganjil
  → Rekap kehadiran siswa per mapel per sesi

# Teacher Monthly Report
GET    /api/v1/teaching-journals/monthly-report?teacher_id={id}&month=2026-04
  → { "total_sessions": 22, "topics_covered": [...], "classes_taught": [...], "attendance_avg": 95.5 }

# Monitoring (Kepala Sekolah)
GET    /api/v1/teaching-journals/completion?date=2026-04-15
  → { "expected_sessions": 120, "submitted": 105, "missing": 15, "teachers_missing": [...] }
```

### Quick Create Flow

```
Guru buka app di pagi hari:
1. GET /schedule-entries?teacher_id={me}&day=today
   → Jadwal hari ini: Matematika VII-A jam 1-2, IPA VII-B jam 3-4, ...

2. Setelah mengajar, guru tap "Isi Jurnal":
   POST /teaching-journals/from-schedule
   Body: { "schedule_entry_id": "018f...", "journal_date": "2026-04-15" }
   → Auto-create jurnal dengan teacher, subject, class, slot pre-filled

3. Guru isi: topic_taught, teaching_method, student_responses, obstacles

4. Guru input absensi sesi:
   POST /teaching-journals/{id}/attendances/bulk
   → Pre-populated list siswa dari class_room, guru update status

5. POST /teaching-journals/{id}/submit
   → Status = submitted, attendance counts denormalized
```

## Consequences

### Positive

- **Realisasi RPP**: Jurnal terhubung ke RPP (S024) — bisa compare rencana vs realisasi.
- **Absensi per sesi**: Granularity lebih tinggi dari S008 — bisa detect siswa yang bolos mapel tertentu.
- **Quick create from schedule**: Guru tidak perlu re-input data jadwal — auto-fill dari S021.
- **Monitoring dashboard**: Kepala sekolah bisa lihat completion rate jurnal harian — guru mana yang belum isi.
- **Pesantren support**: Tracking progress kitab kuning (halaman) dan hafalan per sesi.
- **Denormalized counts**: Statistik kehadiran di jurnal — tidak perlu aggregate setiap kali display.

### Negative / Trade-offs

- **Volume tinggi**: ~125 entries/bulan/guru × 50 guru = ~6.250 jurnal/bulan. Session attendance: 6.250 × 30 siswa = ~187.500 records/bulan. Perlu monitoring disk dan consider archiving.
- **Duplikasi absensi**: Session attendance vs S008 daily attendance — siswa bisa "hadir" di S008 tapi "absent" di beberapa sesi. Perlu klarifikasi ke user bahwa keduanya complementary, bukan redundant.
- **Denormalized counts**: `total_students`, `*_count` harus di-recalculate setiap kali attendance diubah — eventual consistency.
- **Free text teaching_method**: Tidak di-constrain karena sangat bervariasi — tapi membuat aggregation/reporting lebih sulit.
- **No photo/evidence**: Tidak ada field untuk foto bukti mengajar — bisa ditambah sebagai enhancement (mirip lesson_plan_attachments).
- **Status sederhana**: Hanya 3 status (draft/submitted/verified) — jika perlu workflow lebih complex, harus extend.

## Alternatives Considered

### 1. Jurnal sebagai bagian dari S021 schedule_entries
- Ditolak: jadwal adalah template mingguan, jurnal adalah record harian. Satu schedule_entry bisa menghasilkan ~20 jurnal per semester. Separate concern.

### 2. Session attendance di tabel student_attendances (S008)
- Ditolak: S008 adalah absensi harian oleh wali kelas — scope dan pencatat berbeda. Menggabungkan akan membuat S008 terlalu complex dan query menjadi rumit.

### 3. Jurnal tanpa absensi per sesi
- Ditolak: absensi per sesi adalah salah satu kebutuhan utama — guru perlu tahu siapa yang hadir di kelasnya, bukan hanya siapa yang hadir di sekolah.

### 4. Pre-populated jurnal dari jadwal (auto-create)
- Deferred: auto-create jurnal setiap hari dari jadwal mungkin membantu tapi juga create empty records. MVP menggunakan "quick create from schedule" (on-demand) yang lebih bersih.

### 5. Attachment foto/video
- Deferred: untuk MVP, TEXT fields cukup untuk catatan. Photo evidence (foto papan tulis, foto aktivitas) bisa ditambah dengan pattern yang sama seperti lesson_plan_attachments.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `JournalDescriptor.TableName()` | — | `"teaching_journals"` |
| U02 | `SessionAttendanceDescriptor.TableName()` | — | `"journal_session_attendances"` |
| U03 | Validate rejects invalid `semester` | `"summer"` | Error |
| U04 | Validate rejects `slot_start > slot_end` | slot_start=5, slot_end=3 | Error |
| U05 | Validate rejects `slot_start = 0` | `0` | Error |
| U06 | Validate rejects invalid `status` | `"approved"` | Error |
| U07 | Validate rejects negative counts | `present_count = -1` | Error |
| U08 | Validate rejects invalid `attendance_status` | `"skipping"` | Error |
| U09 | Validate rejects invalid `teaching_method_pesantren` | `"lecture"` | Error |
| U10 | Validate rejects `late_minutes > 120` | `150` | Error |
| U11 | Validate accepts valid journal | All fields valid | No error |
| U12 | Validate kitab page range | page_from=10, page_to=5 | Error |

### Integration Tests — Journals

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create journal | POST with valid data | 201 |
| I02 | Create from schedule | POST /from-schedule with schedule_entry_id | 201, auto-filled |
| I03 | Unique per teacher+class+date+slot | Create duplicate | 409/422 |
| I04 | Status CHECK | INSERT with `status = 'approved'` | DB error |
| I05 | Slot range CHECK | INSERT with slot_start=5, slot_end=3 | DB error |
| I06 | Submit journal | POST /teaching-journals/{id}/submit | 200, status=submitted |
| I07 | Verify journal | POST /teaching-journals/{id}/verify | 200, status=verified |
| I08 | Cannot delete submitted | DELETE on status=submitted | 403 |
| I09 | Get teacher journals by date | GET ?teacher_id={id}&date=2026-04-15 | 200, day's entries |
| I10 | Get monthly report | GET /monthly-report?teacher_id={id}&month=2026-04 | 200, summary |

### Integration Tests — Session Attendance

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Bulk input attendance | POST /attendances/bulk for 30 students | 201, 30 records |
| I12 | Unique per journal+student | Insert same student twice | 409/422 |
| I13 | Attendance status CHECK | INSERT with `attendance_status = 'skipping'` | DB error |
| I14 | Late minutes CHECK | INSERT with late_minutes=150 | DB error |
| I15 | Attendance counts denormalized | After bulk input, check journal counts | Counts match actual records |
| I16 | Student session attendance view | GET /students/{id}/session-attendances | 200, per subject breakdown |

### Integration Tests — Pesantren (ADR-009)

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Journal with kitab reference | POST with kitab_reference, page range | 201 |
| I18 | Journal with sorogan method | POST with teaching_method_pesantren=sorogan | 201 |
| I19 | Journal with hafalan progress | POST with hafalan_progress="3 santri setor Al-Mulk" | 201 |
| I20 | Kitab page CHECK | INSERT with page_from=10, page_to=5 | DB error |
| I21 | Pesantren method CHECK | INSERT with method='lecture' | DB error |

### Integration Tests — Monitoring

| # | Test Case | Action | Expected |
|---|---|---|---|
| I22 | Completion check | GET /completion?date=2026-04-15 | 200, expected vs submitted |
| I23 | Missing teachers listed | Teachers without journal for today | Listed in missing |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I24 | TeacherUpdated syncs to journals | Update teacher name | `_data.teacher.full_name` updated |
| I25 | SubjectUpdated syncs to journals | Update subject name | `_data.subject.name` updated |
| I26 | ClassRoomUpdated syncs to journals | Update class name | `_data.class_room.name` updated |
| I27 | StudentUpdated syncs to session attendance | Update student name | `_data.student.full_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I28 | Cannot access other tenant's journals | GET with wrong tenant | 404 |
| I29 | Cannot input attendance to other tenant's journal | POST bulk with wrong tenant | Error |
