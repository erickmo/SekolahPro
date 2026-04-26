# Jadwal Akademik — Jadwal Pelajaran, Kalender, RPP & Jurnal Mengajar

Dokumen ini menjelaskan infrastruktur perencanaan dan pelaksanaan pengajaran di SekolahPro: jadwal pelajaran mingguan, kalender akademik tahunan, rencana pembelajaran (RPP/Modul Ajar), dan jurnal mengajar harian.

---

## ADR References

| ADR | Judul |
|-----|-------|
| S021 | Teaching Schedule / Timetable (Jadwal Pelajaran) |
| S023 | Academic Calendar (Kalender Akademik) |
| S024 | Lesson Plan / RPP (Rencana Pelaksanaan Pembelajaran) |
| S025 | Teaching Journal (Jurnal Mengajar) |

---

## Domain Entities

### JADWAL PELAJARAN (S021)

#### 1. `time_slots` — Definisi Slot Waktu

Template jam pelajaran yang berlaku per semester per tahun ajaran. Mendefinisikan kapan setiap "jam ke-N" berlangsung.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `name` | VARCHAR(30) | NOT NULL | e.g. "Jam ke-1", "Istirahat", "Sholat Dzuhur" |
| `slot_number` | INT | 1-15 | Nomor urut jam (pesantren bisa sampai 15: pagi+siang+malam) |
| `slot_type` | VARCHAR(20) | CHECK | lesson / break / assembly / prayer |
| `day_of_week` | INT | 1-7 | 1=Senin, 7=Minggu |
| `start_time`, `end_time` | TIME | NOT NULL, start < end | Waktu mulai dan selesai |
| `academic_year_id` | UUID | NOT NULL, FK | |
| `semester` | VARCHAR(10) | CHECK: ganjil/genap | Jadwal bisa berbeda antar semester |

**UNIQUE** pada `(tenant_id, company_id, academic_year_id, semester, day_of_week, slot_number)`.

**Pesantren Schedule (3 sesi):**

| Sesi | Waktu | Slot | Tipe | Mapel |
|------|-------|------|------|-------|
| Pagi (Diniyah) | 07:00-09:00 | 1-3 | lesson | Fiqh, Nahwu, Kitab Kuning |
| Sholat Dhuha | 09:00-09:15 | — | prayer | — |
| Siang (Umum) | 09:15-15:00 | 4-12 | lesson | Mapel Kemendikbud |
| Sholat Dzuhur | 12:00-13:00 | — | prayer | — |
| Malam (Tahfidz) | 19:30-21:00 | 13-15 | lesson | Tahfidz, Muroja'ah |

#### 2. `schedule_entries` — Jadwal Per Kelas Per Slot

Satu record = satu mapel di satu slot waktu untuk satu kelas.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `time_slot_id`, `class_room_id`, `subject_id`, `teacher_id`, `academic_year_id` | UUID | NOT NULL, FK | |
| `room_name` | VARCHAR(50) | nullable | Ruangan — nullable jika kelas menetap |
| `semester` | VARCHAR(10) | CHECK: ganjil/genap | |
| `is_substitution` | BOOLEAN | default false | TRUE = jadwal pengganti sementara |
| `original_teacher_id` | UUID | nullable | Guru asli yang digantikan |
| `substitution_date` | DATE | nullable | Berlaku hanya pada tanggal ini |
| `substitution_reason` | TEXT | nullable | Sakit, cuti, dinas luar, dll. |

**Conflict Detection Indexes (PostgreSQL partial unique):**
```sql
-- Guru tidak boleh 2 kelas bersamaan
UNIQUE INDEX uq_schedule_teacher_slot ON schedule_entries (teacher_id, time_slot_id)
WHERE is_substitution = false AND deleted_at IS NULL

-- Ruangan tidak boleh 2 kelas bersamaan
UNIQUE INDEX uq_schedule_room_slot ON schedule_entries (room_name, time_slot_id)
WHERE room_name IS NOT NULL AND is_substitution = false AND deleted_at IS NULL
```

---

### KALENDER AKADEMIK (S023)

#### 3. `academic_calendar_events` — Event Kalender Per Tahun Ajaran

Satu tabel untuk semua event penting dalam satu tahun ajaran.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `academic_year_id` | UUID | NOT NULL, FK | |
| `name` | VARCHAR(200) | NOT NULL | e.g. "Libur Hari Raya Idul Fitri" |
| `event_type` | VARCHAR(30) | CHECK | holiday / exam_period / school_event / semester_start / semester_end / graduation / enrollment_period / teacher_training / islamic_holiday / pesantren_event / ramadan_schedule |
| `start_date`, `end_date` | DATE | start <= end | Range tanggal event |
| `semester` | VARCHAR(10) | nullable | NULL = lintas semester |
| `is_school_day` | BOOLEAN | default false | TRUE = event tapi siswa tetap masuk |
| `affects_attendance` | BOOLEAN | default true | FALSE = event informatif saja |
| `is_recurring_yearly` | BOOLEAN | default false | TRUE = berulang setiap tahun (untuk carry-forward) |
| `hijri_date` | VARCHAR(30) | nullable | Tanggal Hijriyah — e.g. "1 Syawal 1448 H" |
| `islamic_event_type` | VARCHAR(30) | nullable, CHECK | ramadan / idul_fitri / idul_adha / maulid_nabi / isra_miraj / tahun_baru_hijriyah / nuzulul_quran / haul / khataman / wisuda_tahfidz |
| `color_code` | VARCHAR(7) | nullable, regex `^#[0-9A-Fa-f]{6}$` | Hex color untuk kalender UI |

---

### RPP / MODUL AJAR (S024)

#### 4. `lesson_plans` — Rencana Pelaksanaan Pembelajaran

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `teacher_id`, `subject_id`, `class_room_id`, `academic_year_id` | UUID | NOT NULL, FK | |
| `curriculum_id` | UUID | nullable, FK | Link ke kurikulum (S019) |
| `title` | VARCHAR(300) | NOT NULL | Judul RPP/Modul Ajar |
| `plan_type` | VARCHAR(20) | CHECK | modul_ajar / rpp_k13 / rpp_ktsp / rpp_diniyah |
| `semester` | VARCHAR(10) | ganjil/genap | |
| `meeting_number` | INT | nullable, 1-100 | Pertemuan ke berapa |
| `topic` | VARCHAR(300) | NOT NULL | Topik utama |
| `subtopic` | TEXT | nullable | |
| `duration_minutes` | INT | 1-480 | Default 90 (2 × 45 menit) |
| `learning_objectives` | TEXT | NOT NULL | Tujuan pembelajaran (TP) — wajib semua kurikulum |
| `learning_activities` | TEXT | NOT NULL | Langkah pembelajaran — pendahuluan/inti/penutup |
| `assessment_plan` | TEXT | nullable | Rencana penilaian (formatif + sumatif) |
| `teaching_methods` | TEXT | nullable | Diskusi, ceramah, praktikum, PBL, dll. |
| `media_and_resources` | TEXT | nullable | Media dan sumber belajar |
| `differentiation_notes` | TEXT | nullable | Catatan diferensiasi pembelajaran |
| Kolom K13 spesifik | TEXT | nullable | `core_competency`, `basic_competency`, `indicators` |
| Kolom Merdeka spesifik | TEXT | nullable | `pancasila_profile`, `trigger_questions`, `reflection` |
| Kolom Pesantren | — | nullable | `kitab_reference`, `bab_fashl`, `teaching_method_pesantren`, `hafalan_target` |
| `status` | VARCHAR(20) | CHECK | draft / submitted / in_review / revision_needed / approved / archived |
| `submitted_at` | TIMESTAMPTZ | nullable | |
| `reviewed_by` | UUID | nullable | Wakasek kurikulum |
| `reviewed_at` | TIMESTAMPTZ | nullable | |
| `review_notes` | TEXT | nullable | Catatan revisi dari wakasek |

**Metode Pesantren (`teaching_method_pesantren`):**
bandongan / sorogan / halaqah / muhafadzah / mudzakarah / ceramah / demonstrasi / tanya_jawab

#### 5. `lesson_plan_attachments` — File Lampiran RPP

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `lesson_plan_id` | UUID | FK |
| `file_name` | VARCHAR(255) | NOT NULL |
| `file_path` | TEXT | NOT NULL — path ke object storage |
| `file_size_bytes` | BIGINT | max 50MB |
| `mime_type` | VARCHAR(100) | PDF / Word / PowerPoint / Excel / JPEG / PNG / WebP |

---

### JURNAL MENGAJAR (S025)

#### 6. `teaching_journals` — Log Harian Per Sesi Mengajar

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `teacher_id`, `subject_id`, `class_room_id`, `academic_year_id` | UUID | NOT NULL, FK | |
| `schedule_entry_id` | UUID | nullable, FK | Link ke jadwal pelajaran |
| `lesson_plan_id` | UUID | nullable, FK | Link ke RPP |
| `journal_date` | DATE | NOT NULL | |
| `semester` | VARCHAR(10) | ganjil/genap | |
| `slot_start`, `slot_end` | INT | 1-15 | Jam ke berapa (range untuk sesi ganda) |
| `start_time`, `end_time` | TIME | nullable | Waktu aktual — bisa berbeda dari jadwal |
| `topic_taught` | VARCHAR(300) | NOT NULL | Materi yang diajarkan |
| `material_detail` | TEXT | nullable | Detail materi |
| `teaching_method` | VARCHAR(50) | nullable | Free text: diskusi, ceramah, PBL, dll. |
| `student_responses` | TEXT | nullable | Catatan respon siswa |
| `obstacles_notes` | TEXT | nullable | Kendala: LCD rusak, materi sulit, dll. |
| `follow_up_plan` | TEXT | nullable | Rencana tindak lanjut |
| Kolom Pesantren | — | nullable | `kitab_reference`, `kitab_page_from/to`, `teaching_method_pesantren`, `hafalan_progress` |
| Statistik denormalisasi | INT | | `total_students`, `present_count`, `absent_count`, `late_count`, `permission_count` |
| `status` | VARCHAR(20) | CHECK | draft / submitted / verified |

**UNIQUE** pada `(teacher_id, class_room_id, journal_date, slot_start)`.

#### 7. `journal_session_attendances` — Absensi Per Sesi Mengajar

Berbeda dari `student_attendances` (S008) yang absensi harian oleh wali kelas — ini adalah absensi per mata pelajaran oleh guru pengajar.

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `journal_id`, `student_id` | UUID | FK |
| `attendance_status` | VARCHAR(15) | present / absent / late / sick / permitted / dispensasi |
| `late_minutes` | INT | nullable, max 120 — menit keterlambatan |
| `notes` | TEXT | nullable — "Izin ke UKS jam 09:30" |

**UNIQUE** pada `(journal_id, student_id)`.

---

## Business Rules

### Jadwal Pelajaran

1. **Guru tidak boleh dijadwalkan di 2 kelas pada waktu yang sama** — ditegakkan via partial unique index di database.
2. **Ruangan tidak boleh dipakai 2 kelas pada waktu yang sama** — partial unique index pada `room_name` + `time_slot_id`.
3. **Conflict detection 2 level:** database (unique index) + service layer (pesan error yang deskriptif sebelum insert).
4. **Substitusi guru berlaku satu hari saja** — `substitution_date` menentukan tanggal berlaku.
5. **Slot waktu berbeda antar semester** — time_slots di-scope per semester; memungkinkan perubahan jadwal semester genap vs ganjil.
6. **Pesantren support hingga 15 slot per hari** (pagi/siang/malam) dengan `slot_type = 'prayer'` untuk istirahat sholat.
7. **Slot `break` dan `prayer` tidak bisa diisi schedule_entry** (hanya `lesson` dan `assembly`).
8. **Carry-forward jadwal** dari semester/tahun sebelumnya tersedia untuk mengurangi setup ulang.
9. **Room management minimal di MVP** — `room_name` sebagai string, bukan FK ke tabel rooms. Room FK bisa dievolusi nanti.
10. **Total slot per mapel per minggu** harus sesuai `credit_hours_per_week` dari konfigurasi mapel (S020) — validasi di application layer, bukan constraint database.

### Kalender Akademik

11. **Kalender menjadi sumber kebenaran hari efektif** yang digunakan absensi (S008) untuk menghitung persentase kehadiran.
12. **Kalkulasi hari efektif:**
    ```
    Hari Efektif = Total hari kerja - Jumlah hari event WHERE affects_attendance = true AND is_school_day = false
    ```
13. **Event dengan `is_school_day = true`** (misal: upacara kemerdekaan) tetap dihitung sebagai hari efektif — siswa wajib hadir.
14. **Event dengan `affects_attendance = false`** bersifat informatif — tidak mempengaruhi hitungan hari efektif.
15. **Ramadan pesantren** ditandai `islamic_event_type = 'ramadan'` — sistem bisa menyesuaikan durasi slot jadwal.
16. **Carry-forward event recurring** — event `is_recurring_yearly = true` bisa di-copy ke tahun ajaran baru dengan penyesuaian tanggal.
17. **Tanggal Hijriyah disimpan sebagai string** — PostgreSQL tidak punya native Hijri type; admin input manual.

### RPP / Modul Ajar

18. **RPP wajib punya `learning_objectives` dan `learning_activities`** — komponen minimal yang berlaku di semua kurikulum.
19. **Alur approval RPP:** draft → submitted → in_review → approved (atau revision_needed → submitted ulang).
20. **Hanya `draft` dan `revision_needed` yang bisa di-edit.** RPP yang sudah `approved` tidak bisa diubah.
21. **Wakasek kurikulum yang mereview dan approve.** Bulk approve tersedia untuk efisiensi.
22. **Komponen K13 (KI/KD/Indikator) nullable** untuk Kurikulum Merdeka.
23. **Komponen Merdeka (pancasila_profile, reflection) nullable** untuk K13.
24. **Pesantren RPP** mendukung referensi kitab (kitab_reference, bab_fashl) dan metode bandongan/sorogan.
25. **Duplicate/carry-forward RPP** ke semester atau kelas lain tersedia — reset status ke `draft`.
26. **File attachment RPP** maksimal 50MB per file; format: PDF, Word, PowerPoint, Excel, gambar.

### Jurnal Mengajar

27. **Guru wajib isi jurnal setiap sesi mengajar** — ini adalah komponen penilaian kinerja guru (Permendikbud).
28. **Quick create from schedule** — guru tidak perlu re-input data jadwal; sistem auto-fill dari `schedule_entry`.
29. **Jurnal terhubung ke RPP** untuk tracking rencana vs realisasi.
30. **Absensi per sesi berbeda dari absensi harian (S008):** S008 = wali kelas, satu kali per hari; S025 = guru mapel, per sesi mengajar.
31. **Siswa bisa "hadir" di S008 tapi "absent" di session attendance** — bolos mapel tertentu terdeteksi.
32. **Statistik kehadiran denormalisasi** di jurnal — `present_count`, `absent_count`, dll. diupdate otomatis saat bulk input attendance.
33. **Monitoring kepsek:** endpoint completion check menampilkan guru yang belum isi jurnal hari ini.
34. **Jurnal yang sudah submitted tidak bisa dihapus** — hanya bisa di-edit sebelum submit.

---

## Alur Conflict Detection Jadwal

```
1. Admin POST /schedule-entries
2. Service layer:
   a. Cek teacher conflict: SELECT WHERE teacher_id = $1 AND time_slot_id = $2 AND is_substitution = false
   b. Cek room conflict: SELECT WHERE room_name = $1 AND time_slot_id = $2 AND is_substitution = false
3. Jika ada conflict → 409 Conflict dengan pesan deskriptif:
   "Bu Siti sudah mengajar VII-B pada Senin Jam ke-3"
4. Jika tidak ada conflict → INSERT schedule_entry
5. Database-level backup: partial unique indexes akan block jika service layer terlewat
```

## Alur Pengisian Jurnal Mengajar

```
1. Guru buka jadwal hari ini:
   GET /schedule-entries?teacher_id=me&day=senin
2. Setelah mengajar, guru buat jurnal (quick create):
   POST /teaching-journals/from-schedule
   Body: { schedule_entry_id: "...", journal_date: "2026-04-15" }
   → Auto-fill: teacher, subject, class, slot_start/end dari schedule_entry
3. Guru mengisi konten jurnal:
   PUT /teaching-journals/{id}
   Body: { topic_taught: "Persamaan Linear", teaching_method: "Diskusi kelompok", ... }
4. Guru input absensi sesi:
   POST /teaching-journals/{id}/attendances/bulk
   Body: { attendances: [{ student_id, attendance_status: "present" }, ...] }
   → Auto-update: present_count, absent_count di jurnal
5. Guru submit:
   POST /teaching-journals/{id}/submit
   → status = submitted
6. Kepsek verify (opsional):
   POST /teaching-journals/{id}/verify
   → status = verified
```

## Alur Kalkulasi Hari Efektif

```
GET /calendar-events/effective-days?year_id={id}&semester=ganjil

Algoritma:
1. Hitung total hari kerja (Senin-Sabtu atau Senin-Jumat, sesuai config sekolah)
   dalam range semester (e.g. 14 Juli 2025 - 20 Desember 2025)
   → 156 hari
2. Kurangi hari event WHERE:
   - affects_attendance = true
   - is_school_day = false
   (Libur nasional 12 hari + libur semester 14 hari = 26 hari)
3. Hari efektif = 156 - 26 = 130 hari

Response: { total_work_days: 156, holiday_days: 26, effective_days: 130 }
```

---

## API Endpoints

```
# Time Slots
GET    /api/v1/time-slots?year_id={id}&semester=ganjil    List slot waktu
POST   /api/v1/time-slots/bulk                            Bulk create (per hari)
PUT    /api/v1/time-slots/{id}                            Update slot
POST   /api/v1/time-slots/carry-forward                   Copy dari semester/tahun lalu

# Schedule Entries
GET    /api/v1/schedule-entries?class_id={id}&semester=ganjil    Jadwal per kelas
GET    /api/v1/schedule-entries?teacher_id={id}&semester=ganjil  Jadwal per guru
POST   /api/v1/schedule-entries                           Tambah jadwal (dengan conflict check)
POST   /api/v1/schedule-entries/bulk                     Bulk create
PUT    /api/v1/schedule-entries/{id}                     Update
DELETE /api/v1/schedule-entries/{id}                     Hapus
POST   /api/v1/schedule-entries/check-conflict           Check tanpa insert
POST   /api/v1/schedule-entries/{id}/substitute          Buat substitusi guru
GET    /api/v1/schedules/weekly?class_id={id}            View jadwal mingguan (matrix)
GET    /api/v1/schedules/teacher-weekly?teacher_id={id}  Jadwal guru per minggu

# Calendar Events
GET    /api/v1/calendar-events?year_id={id}              Semua event tahun ini
POST   /api/v1/calendar-events                           Buat event
POST   /api/v1/calendar-events/bulk                     Bulk create (import libur nasional)
PUT    /api/v1/calendar-events/{id}                     Update
POST   /api/v1/calendar-events/carry-forward            Copy event recurring ke tahun baru
GET    /api/v1/calendar-events/effective-days            Hitung hari efektif
GET    /api/v1/calendar-events/today                    Status hari ini (libur/masuk)

# Lesson Plans
GET    /api/v1/lesson-plans?teacher_id={id}&semester=ganjil        List RPP per guru
GET    /api/v1/lesson-plans?status=submitted                       Menunggu review
POST   /api/v1/lesson-plans                                        Buat RPP
PUT    /api/v1/lesson-plans/{id}                                   Update (hanya draft)
POST   /api/v1/lesson-plans/{id}/submit                            Submit
POST   /api/v1/lesson-plans/{id}/approve                           Approve (wakasek)
POST   /api/v1/lesson-plans/{id}/request-revision                  Minta revisi
POST   /api/v1/lesson-plans/bulk-approve                           Bulk approve
POST   /api/v1/lesson-plans/{id}/attachments                       Upload file
GET    /api/v1/lesson-plans/statistics?year_id={id}                Statistik (wakasek dashboard)
POST   /api/v1/lesson-plans/{id}/duplicate                         Duplikat ke semester/kelas lain

# Teaching Journals
GET    /api/v1/teaching-journals?teacher_id={id}&date=2026-04-15   Per guru per hari
GET    /api/v1/teaching-journals?teacher_id={id}&month=2026-04     Per bulan
POST   /api/v1/teaching-journals                                   Buat jurnal
POST   /api/v1/teaching-journals/from-schedule                     Quick create dari jadwal
PUT    /api/v1/teaching-journals/{id}                              Update
POST   /api/v1/teaching-journals/{id}/submit                       Submit
POST   /api/v1/teaching-journals/{id}/verify                       Kepsek verify
POST   /api/v1/teaching-journals/{id}/attendances/bulk             Bulk input absensi sesi
GET    /api/v1/teaching-journals/completion?date=2026-04-15        Monitor kelengkapan hari ini
GET    /api/v1/teaching-journals/monthly-report?teacher_id={id}    Laporan bulanan guru
GET    /api/v1/students/{id}/session-attendances                   Kehadiran per mapel (siswa view)
```

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Conflict detection di 2 level (DB + service) | DB sebagai safety net; service memberikan error message yang informatif |
| `time_slots` sebagai tabel terpisah | Template jam — konsistensi waktu antar semua entri jadwal; sekolah yang berbeda bisa punya jam berbeda |
| Room sebagai string, bukan FK | MVP simplicity; room management (kapasitas, fasilitas) bisa dikembangkan nanti |
| Kalender sebagai sumber hari efektif | Kalkulasi persentase kehadiran yang akurat; sumber kebenaran terpusat |
| Event tunggal untuk semua jenis kalender | Struktur sama (nama, tanggal, tipe) — satu tabel lebih mudah query dan maintain |
| RPP 3 type di satu tabel (bukan per kurikulum) | 80% field sama antar kurikulum; nullable kurikulum-specific fields trade-off untuk simplicity |
| Approval workflow RPP | Kontrol kualitas; sekolah terakreditasi membutuhkan bukti review RPP |
| Jurnal terpisah dari jadwal | Jadwal = template mingguan; jurnal = record harian — satu schedule bisa menghasilkan 20+ jurnal per semester |
| Session attendance terpisah dari S008 | Granularity berbeda (per sesi vs per hari) dan pencatat berbeda (guru mapel vs wali kelas) |
| Denormalized attendance counts di jurnal | Quick display tanpa aggregate query — eventually consistent via update |
| Pesantren support hingga slot 15 | 3 sesi per hari (pagi diniyah + siang umum + malam tahfidz) membutuhkan lebih dari 10 slot |

---

## Integration Points

| Titik Integrasi | Keterangan |
|----------------|-----------|
| S020 (Subject Config) | `credit_hours_per_week` menentukan jumlah slot yang harus dialokasikan per mapel di jadwal |
| S022 (Assessment) | `scheduled_date` assessment mengacu kalender — tidak boleh di hari libur kalender akademik |
| S008 (Attendance) | Kalender menentukan hari efektif untuk kalkulasi persentase kehadiran |
| S025 → S021 (Journal → Schedule) | Quick create jurnal dari schedule entry — auto-fill data |
| S025 → S024 (Journal → Lesson Plan) | Jurnal bisa di-link ke RPP untuk tracking rencana vs realisasi |
| S019 (Curriculum) | RPP/Modul Ajar di-link ke kurikulum untuk memastikan alignment dengan CP/TP/ATP |
| Teacher Module (ADR-012) | `teacher_id` di `schedule_entries`, `lesson_plans`, `teaching_journals` — SyncEngine update `_data.teacher` |
