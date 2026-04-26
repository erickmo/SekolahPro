# Domain Sekolah — Gambaran Umum

Domain `sekolah` adalah inti dari sistem SekolahPro. Domain ini mengelola seluruh operasi akademik sekolah: siswa, kegiatan belajar-mengajar, kurikulum, penilaian, dan pelaporan. Semua modul bermuara pada satu output utama: rapor semester yang diterima orang tua dan siswa.

---

## ADR References

| ADR | Judul | Status |
|-----|-------|--------|
| S001–S018 | Student subdomain (core, relasi, lifecycle, akademik, kesejahteraan) | Approved |
| S019–S025 | Academic subdomain (kurikulum, mapel, jadwal, ujian, kalender, RPP, jurnal) | Proposed |

---

## Konteks Sistem Pendidikan Indonesia

### Jenjang Pendidikan

SekolahPro mendukung tiga jenjang utama:

| Jenjang | Kelas | Fase Merdeka | Durasi |
|---------|-------|--------------|--------|
| SD (Sekolah Dasar) | 1–6 | A (1-2), B (3-4), C (5-6) | 6 tahun |
| SMP (Sekolah Menengah Pertama) | 7–9 | D | 3 tahun |
| SMA/SMK (Sekolah Menengah Atas) | 10–12 | E (10), F (11-12) | 3 tahun |

Sistem juga mendukung **Madrasah** (MI/MTs/MA) dengan struktur setara dan **Pondok Pesantren** yang menggabungkan kurikulum diniyah di atas kurikulum nasional.

### Sistem Kurikulum

Indonesia saat ini berada dalam masa **transisi kurikulum**:

| Kurikulum | Tahun Berlaku | Karakteristik |
|-----------|--------------|---------------|
| **Kurikulum Merdeka** | 2022–sekarang | Capaian Pembelajaran (CP), Tujuan Pembelajaran (TP), Alur TP (ATP), Modul Ajar, Profil Pelajar Pancasila (P5). Fleksibel — guru bebas menyusun. |
| **Kurikulum 2013 (K13)** | 2013–2022 (masih dipakai) | Kompetensi Inti (KI-1 s/d KI-4), Kompetensi Dasar (KD), indikator. Format RPP baku (10-20 halaman). |
| **KTSP** | Legacy | Kurikulum Tingkat Satuan Pendidikan — varian K13 yang lebih tua. |
| **Kurikulum Diniyah** | Pesantren | Mapel keagamaan (Fiqh, Aqidah, Nahwu, Tahfidz, Kitab Kuning) di atas kurikulum nasional. |

### Struktur Tahun Ajaran

```
Tahun Ajaran (e.g. 2025/2026)
  ├─ Semester Ganjil: Juli – Desember
  │    ├─ PTS/STS (Penilaian Tengah Semester)
  │    └─ PAS/SAS (Penilaian Akhir Semester)
  └─ Semester Genap: Januari – Juni
       ├─ PTS/STS
       └─ PAS/SAS + PAT (Penilaian Akhir Tahun)
```

Setiap semester menghasilkan **satu rapor** per siswa.

### Integrasi Nasional: DAPODIK

DAPODIK (Data Pokok Pendidikan) adalah basis data nasional Kemendikbud. SekolahPro mengintegrasikan data siswa dengan DAPODIK melalui:

- **NISN** (Nomor Induk Siswa Nasional) — identifier siswa nasional
- **NPSN** (Nomor Pokok Sekolah Nasional) — identifier sekolah
- **NIP** guru — identifier ASN (Aparatur Sipil Negara)
- Field-field biodata yang mengikuti standar Dapodik (agama, jenjang pendidikan orang tua, dll.)

---

## Kelompok Entitas Utama

### 1. Student Subdomain (ADR S001–S018)

Mengelola seluruh data dan lifecycle siswa:

```
Siswa (Student)
  ├─ Data Diri → S001 (core), S007 (alamat, sekolah asal)
  ├─ Keluarga  → S003 (orang tua/wali)
  ├─ Lifecycle → S014 (penempatan kelas), S016 (PPDB)
  ├─ Akademik  → S004 (rekap semester), S008 (absensi harian),
  │              S011 (nilai per mapel), S018 (rapor)
  ├─ Kesehatan → S005
  ├─ Keuangan  → S009 (SPP)
  ├─ Dokumen   → S010
  ├─ Kegiatan  → S015 (ekskul)
  └─ Pembinaan → S012 (disiplin), S013 (prestasi), S017 (BK)
```

### 2. Academic Subdomain (ADR S019–S025)

Mengelola infrastruktur akademik:

```
Akademik
  ├─ Fondasi    → S019 (kurikulum), S020 (mata pelajaran)
  ├─ Jadwal     → S021 (jadwal pelajaran), S023 (kalender akademik)
  ├─ Penilaian  → S022 (ujian & assessment)
  └─ Pengajaran → S024 (RPP/Modul Ajar), S025 (jurnal mengajar)
```

---

## Arsitektur Domain

### Pola Vernon Denormalized Read-Cache

Seluruh entitas di domain `sekolah` menggunakan **Vernon Pattern** dengan alasan:

1. **Read:Write ratio tinggi** (estimasi 20:1) — dashboard, laporan, dan pencarian mendominasi
2. **Banyak relasi** — siswa terhubung ke 5+ entitas lain
3. **Business logic sederhana** — mayoritas CRUD + status transition
4. **Eventual consistency acceptable** — data profil tidak perlu real-time

Setiap tabel memiliki kolom standar Vernon:

```sql
_rels        JSONB NOT NULL DEFAULT '{}'   -- FK IDs (e.g. class_room_id)
_data        JSONB NOT NULL DEFAULT '{}'   -- Snapshot relasi (e.g. class_room name)
_sync_status TEXT NOT NULL DEFAULT 'synced'
_sync_version BIGINT NOT NULL DEFAULT 0
```

### Multi-Tenant Architecture

Setiap record di-scope dengan:
```sql
tenant_id  UUID NOT NULL  -- Tenant (group sekolah/yayasan)
company_id UUID NOT NULL  -- Company (sekolah individual)
```

Isolasi diterapkan di application layer — setiap query wajib include `tenant_id` dan `company_id`.

### Event-Driven Sync

Perubahan di entitas parent dipropagasi ke snapshot `_data` via SyncEngine:

| Event | Entitas yang ter-sync |
|-------|----------------------|
| `ClassRoomUpdated` | students, student_academics, student_attendances, schedules |
| `AcademicYearUpdated` | semua entitas yang reference academic_year |
| `StudentUpdated` | student_academics, student_grades, student_health, dll. |
| `SubjectUpdated` | student_grades, assessments, schedule_entries, lesson_plans |
| `TeacherUpdated` | schedule_entries, assessments, lesson_plans, teaching_journals |

---

## Alur Data Utama

### Alur Semester (End-to-End)

```
[Awal Semester]
  Setup Kurikulum (S019)
  → Setup Mata Pelajaran & Konfigurasi (S020)
  → Setup Jadwal Pelajaran (S021)
  → Setup Kalender Akademik (S023)

[Selama Semester]
  Guru mengajar
  → Isi Jurnal Mengajar (S025)
  → Input Nilai Assessment (S022)
  → Absensi Harian (S008)

[Akhir Semester]
  Hitung Nilai Akhir: S022 assessment scores → S011 student_grades
  Rekap Absensi: S008 daily → S004 academic record
  Generate Rapor: S001+S003+S004+S008+S011+S012+S015 → S018 rapor

[Akhir Tahun]
  Kenaikan Kelas (S014) → Update student._data.class_room
```

### Alur PPDB

```
Calon Siswa Daftar Online (S016 applicant)
→ Verifikasi Dokumen
→ Seleksi (zonasi/prestasi/afirmasi/pindahan/reguler)
→ Pengumuman
→ Daftar Ulang
→ Konversi: Applicant → Student (S001) + Guardian (S003) + Address (S007) + Placement (S014)
```

---

## Dependensi Antar Domain

| Domain | Bergantung Pada |
|--------|----------------|
| `student_grades` (S011) | `subjects` (S020), `academic_years`, `class_rooms` |
| `assessments` (S022) | `subjects`, `academic_years`, `class_rooms`, `teachers` |
| `rapor_records` (S018) | S001, S003, S004, S008, S011, S012, S015 |
| `lesson_plans` (S024) | `teachers`, `subjects`, `class_rooms`, `curricula` (S019) |
| `teaching_journals` (S025) | `teachers`, `subjects`, `class_rooms`, `schedule_entries` (S021), `lesson_plans` (S024) |
| `schedule_entries` (S021) | `time_slots`, `teachers`, `subjects`, `class_rooms` |
| `curricula` (S019) | `academic_years` |

---

## Integrasi Lintas Domain

| Titik Integrasi | Domain Sekolah | Domain Lain |
|----------------|---------------|-------------|
| Siswa → Member Koperasi | `students.id` | Koperasi module: member creation saat siswa aktif |
| SPP → Keuangan Sekolah | `student_payments` | Accounting/Finance module |
| Data Guru | `teachers` (ADR-012) | HR module |
| DAPODIK Sync | `students.nisn`, `student_guardians`, `student_health.disability_type` | Integration middleware |
| Notifikasi Orang Tua | `student_disciplines.parent_notified`, `student_invoices.status` | Notification module |

---

## MVP vs Roadmap

### MVP (Build Now)

| Prioritas | Modul | ADR |
|-----------|-------|-----|
| P0 | Student Core + Dashboard | S001, S002, S006 |
| P0 | Rapor Generation | S018 |
| P1 | Guardian, Health, Academic Record | S003, S004, S005 |
| P1 | Daily Attendance | S008 |
| P1 | Subject Grades | S011 |
| P2 | Kurikulum, Mata Pelajaran | S019, S020 |
| P2 | Finance (SPP) | S009 |
| P2 | Class Placement | S014 |

### Roadmap

| Fase | Modul | ADR |
|------|-------|-----|
| Q3 2026 | Teaching Schedule, Academic Calendar | S021, S023 |
| Q3 2026 | Exam & Assessment | S022 |
| Q4 2026 | Lesson Plan (RPP), Teaching Journal | S024, S025 |
| Q4 2026 | Document, Discipline, Achievement | S010, S012, S013 |
| Q4 2026 | Extracurricular, Counseling (BK) | S015, S017 |
| Q1-Q2 2027 | PPDB / Admission | S016 |
