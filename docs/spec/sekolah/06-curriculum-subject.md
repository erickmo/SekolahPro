# Kurikulum & Mata Pelajaran

Dokumen ini menjelaskan manajemen kurikulum dan mata pelajaran di SekolahPro. Kurikulum adalah fondasi dari seluruh aktivitas akademik — tanpa kurikulum yang terdefinisi, tidak ada dasar untuk menentukan mata pelajaran, jadwal, penilaian, RPP, atau rapor.

---

## ADR References

| ADR | Judul |
|-----|-------|
| S019 | Curriculum Management (Manajemen Kurikulum) |
| S020 | Subject Management (Manajemen Mata Pelajaran) |

---

## Konteks Kurikulum Indonesia

### Jenis Kurikulum yang Didukung

| Tipe | Nama | Karakteristik |
|------|------|---------------|
| `merdeka` | Kurikulum Merdeka | CP → TP → ATP; Modul Ajar; Profil Pelajar Pancasila (P5); berlaku 2022+ |
| `k13` | Kurikulum 2013 | KI (1-4) → KD → Indikator; RPP baku; format rapor terstruktur |
| `ktsp` | KTSP | Legacy; mirip K13 tapi versi lebih tua |
| `diniyah` | Kurikulum Diniyah | Pesantren; mapel keagamaan di atas kurikulum nasional |
| `custom` | Kustom | Kurikulum khusus sekolah (bilingual, internasional, dll.) |

### Hierarki Kurikulum Merdeka

```
Kurikulum (e.g. Merdeka 2025/2026 Fase D Kelas 7)
  └─ Capaian Pembelajaran (CP)  ← per fase, per mapel
       └─ Tujuan Pembelajaran (TP)  ← spesifik per unit/bab
            └─ Alur Tujuan Pembelajaran (ATP)  ← urutan TP per semester
```

### Fase Kurikulum Merdeka

| Fase | Kelas | Jenjang |
|------|-------|---------|
| A | 1-2 | SD Awal |
| B | 3-4 | SD Tengah |
| C | 5-6 | SD Akhir |
| D | 7-9 | SMP |
| E | 10 | SMA Kelas 10 |
| F | 11-12 | SMA Kelas 11-12 |

### Profil Pelajar Pancasila (P5)

6 dimensi P5 yang diukur melalui proyek lintas mapel:
1. Beriman, Bertakwa kepada Tuhan YME, dan Berakhlak Mulia
2. Berkebinekaan Global
3. Bergotong Royong
4. Mandiri
5. Bernalar Kritis
6. Kreatif

---

## Domain Entities

### 1. `curricula` — Master Kurikulum

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `name` | VARCHAR(100) | NOT NULL | e.g. "Kurikulum Merdeka 2025/2026 Kelas 7" |
| `code` | VARCHAR(20) | UNIQUE per company | e.g. "MRD-2526-7" |
| `curriculum_type` | VARCHAR(20) | CHECK | merdeka / k13 / ktsp / diniyah / custom |
| `academic_year_id` | UUID | NOT NULL, FK | Kurikulum berlaku per tahun ajaran |
| `grade_level` | VARCHAR(5) | CHECK: 1-12 | Berlaku untuk kelas berapa |
| `phase` | VARCHAR(10) | nullable, CHECK: A-F | Fase Kurikulum Merdeka |
| `is_active` | BOOLEAN | NOT NULL | |
| `description` | TEXT | nullable | |

**UNIQUE** pada `(tenant_id, company_id, academic_year_id, grade_level, curriculum_type)` — tidak ada duplikasi kurikulum yang sama untuk tahun/kelas/tipe yang sama.

**UNIQUE** pada `code` per company.

### 2. `learning_outcomes` — CP / TP / ATP

Hierarki kompetensi menggunakan self-referencing `parent_id`.

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `curriculum_id` | UUID | NOT NULL, FK — kurikulum induk |
| `subject_id` | UUID | nullable, FK — mapel terkait (null untuk CP lintas mapel) |
| `parent_id` | UUID | nullable, FK self-reference — CP tidak punya parent; TP parent = CP; ATP parent = TP |
| `outcome_type` | VARCHAR(10) | cp / tp / atp |
| `code` | VARCHAR(30) | UNIQUE per curriculum |
| `title` | VARCHAR(255) | NOT NULL |
| `description` | TEXT | nullable |
| `semester` | VARCHAR(10) | nullable — ATP per semester; CP/TP bisa lintas semester |
| `sequence_order` | INT | Urutan tampil |
| `ki_number` | INT | nullable — K13: 1/2/3/4 (KI Spiritual/Sosial/Pengetahuan/Keterampilan) |
| `kd_code` | VARCHAR(20) | nullable — K13: kode KD seperti "3.1", "4.2" |
| `is_active` | BOOLEAN | |

### 3. `p5_projects` — Proyek Profil Pelajar Pancasila

Proyek lintas mapel untuk pengukuran dimensi P5. Khusus Kurikulum Merdeka.

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `curriculum_id`, `academic_year_id` | UUID | FK |
| `name` | VARCHAR(200) | NOT NULL |
| `theme` | VARCHAR(50) | CHECK — 7 tema: gaya_hidup_berkelanjutan / kearifan_lokal / bhinneka_tunggal_ika / bangunlah_jiwa_raganya / suara_demokrasi / berekayasa_dan_berteknologi / kewirausahaan |
| `grade_level` | VARCHAR(5) | CHECK: 1-12 |
| `semester` | VARCHAR(10) | ganjil / genap |
| `p5_dimensions` | JSONB array | Dimensi P5 yang diukur — multi-select dari 6 dimensi |
| `start_date`, `end_date` | DATE | nullable — jadwal proyek |
| `status` | VARCHAR(20) | draft / active / completed / archived |

---

### 4. `subjects` — Master Mata Pelajaran

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `name` | VARCHAR(100) | NOT NULL | Nama mapel |
| `code` | VARCHAR(20) | UNIQUE per company | Kode singkat: MTK, BIN, IPA, dll. |
| `name_en` | VARCHAR(100) | nullable | Nama Inggris untuk rapor bilingual |
| `subject_group` | VARCHAR(30) | CHECK (16 group) | Kelompok mapel — lihat tabel di bawah |
| `category` | VARCHAR(20) | CHECK | wajib_nasional / wajib_lokal / pilihan / diniyah / ekstra_kurikuler |
| `is_national` | BOOLEAN | | TRUE = ditetapkan Kemendikbud |
| `is_scored` | BOOLEAN | | FALSE untuk mapel P5 (hanya deskripsi) |
| `sort_order` | INT | | Urutan tampil di rapor |

**Daftar Subject Group:**

| Group | Contoh Mapel |
|-------|-------------|
| `agama` | Pendidikan Agama Islam/Kristen/dll. |
| `pkn` | Pendidikan Pancasila & Kewarganegaraan |
| `bahasa` | Bahasa Indonesia, Bahasa Inggris |
| `matematika` | Matematika |
| `ipa` | Ilmu Pengetahuan Alam |
| `ips` | Ilmu Pengetahuan Sosial |
| `seni_budaya` | Seni Budaya |
| `pjok` | Pendidikan Jasmani, Olahraga & Kesehatan |
| `prakarya` | Prakarya & Kewirausahaan |
| `informatika` | Informatika (mapel baru Kurikulum Merdeka) |
| `muatan_lokal` | Bahasa Jawa, Bahasa Sunda, dll. |
| `keagamaan` | Fiqh, Aqidah, Nahwu, dll. (pesantren) |
| `tahfidz` | Tahfidz Al-Quran (pesantren) |
| `lintas_minat` | Mata pelajaran pilihan lintas jurusan (SMA) |
| `peminatan` | Mata pelajaran jurusan (IPA/IPS/Bahasa) |
| `other` | Lain-lain |

### 5. `subject_configurations` — Konfigurasi Per Tahun Per Jenjang

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `subject_id`, `academic_year_id` | UUID | NOT NULL, FK | |
| `curriculum_id` | UUID | nullable, FK | Link ke kurikulum jika ada |
| `grade_level` | VARCHAR(5) | CHECK: 1-12 | |
| `credit_hours_per_week` | INT | 1-12 | Jam pelajaran per minggu — input utama jadwal |
| `weight_knowledge` | INT | CHECK | Bobot pengetahuan (%) — total dengan skill = 100 |
| `weight_skill` | INT | CHECK | Bobot keterampilan (%) |
| `passing_grade` | NUMERIC(5,2) | 0-100 | KKM (K13) atau KKTP (Merdeka) |
| `is_active` | BOOLEAN | | |

**UNIQUE** pada `(subject_id, academic_year_id, grade_level)` — satu konfigurasi per mapel per tahun per kelas.

**Constraint:** `weight_knowledge + weight_skill = 100` (selalu).

---

## Daftar Mapel Pesantren (ADR-009)

Sekolah berbasis pesantren dapat menambahkan mata pelajaran diniyah:

| Kode | Nama Mapel | Group | Category |
|------|-----------|-------|----------|
| FQH | Fiqh | keagamaan | diniyah |
| AQD | Aqidah Akhlak | keagamaan | diniyah |
| QHD | Quran Hadits | keagamaan | diniyah |
| SKI | Sejarah Kebudayaan Islam | keagamaan | diniyah |
| ARB | Bahasa Arab | keagamaan | diniyah |
| THF | Tahfidz Al-Quran | tahfidz | diniyah |
| KTB | Kitab Kuning | keagamaan | diniyah |
| NHW | Nahwu Shorof | keagamaan | diniyah |

---

## Business Rules

### Kurikulum

1. **Kurikulum terikat tahun ajaran.** Setiap tahun ajaran baru, kurikulum bisa berevolusi (ditambah/diubah CP/TP).
2. **Satu kurikulum per tipe per kelas per tahun.** Tidak ada duplikasi `(academic_year_id, grade_level, curriculum_type)`.
3. **Pesantren bisa punya 2 kurikulum aktif bersamaan:** kurikulum nasional + kurikulum diniyah, untuk kelas dan tahun ajaran yang sama.
4. **Fase (A-F) hanya untuk Kurikulum Merdeka.** Field phase NULL untuk K13/KTSP/Diniyah.
5. **Hierarki CP > TP > ATP harus konsisten.** ATP harus punya parent TP; TP harus punya parent CP. Validasi di application layer.
6. **Self-referencing tree di `learning_outcomes`** menggunakan recursive query (`WITH RECURSIVE`) untuk render pohon lengkap.
7. **Kode learning outcome unik per kurikulum.** Format bebas tapi harus konsisten (e.g. "MTK.D.CP.1", "MTK.D.TP.1.1").
8. **P5 tema ditetapkan Kemendikbud** — 7 tema tidak bisa ditambah tanpa alter schema. Jika Kemendikbud menambah tema, perlu migrasi.
9. **P5 dimensi disimpan sebagai JSONB array** — satu proyek bisa mengukur beberapa dimensi sekaligus.
10. **KI/KD field (ki_number, kd_code)** hanya relevan untuk K13/KTSP. Nullable untuk Kurikulum Merdeka.

### Mata Pelajaran

11. **Mapel national (`is_national = true`)** ditetapkan Kemendikbud — tidak bisa dihapus, hanya dinonaktifkan.
12. **Mapel diniyah pesantren** memiliki `category = 'diniyah'` dan `is_national = false`.
13. **Mapel P5** (`is_scored = false`) tidak menghasilkan nilai angka — hanya deskripsi naratif.
14. **Informatika** adalah mapel baru Kurikulum Merdeka — `subject_group = 'informatika'`.
15. **Carry-forward konfigurasi mapel** dari tahun sebelumnya ke tahun baru: `POST /subject-configurations/carry-forward`. Admin bisa edit setelah carry-forward.
16. **Bobot nilai total harus 100%.** Constraint database: `weight_knowledge + weight_skill = 100`.
17. **KKM/KKTP dikonfigurasi per mapel per jenjang.** Default 70.00. Sekolah bisa menetapkan berbeda (e.g. Matematika KKM 65, B. Indonesia KKM 75).
18. **Jam pelajaran per minggu** (`credit_hours_per_week`) menjadi input utama untuk modul jadwal (S021).

---

## Alur Setup Kurikulum Awal Tahun Ajaran

```
1. Admin buat academic_year baru (e.g. "2025/2026")
2. Admin buat curricula:
   POST /curricula
   Body: { curriculum_type: "merdeka", grade_level: "7", phase: "D", academic_year_id }
3. Admin input CP per mapel:
   POST /learning-outcomes
   Body: { curriculum_id, subject_id, outcome_type: "cp", code: "MTK.D.CP.1", title: "..." }
4. Admin input TP di bawah CP:
   POST /learning-outcomes
   Body: { parent_id: cp_id, outcome_type: "tp", ... }
5. Admin input ATP per semester:
   POST /learning-outcomes
   Body: { parent_id: tp_id, outcome_type: "atp", semester: "ganjil", sequence_order: 1, ... }
6. Admin setup subject_configurations:
   POST /subject-configurations
   Body: { subject_id, academic_year_id, grade_level: "7",
           credit_hours_per_week: 5, weight_knowledge: 50, weight_skill: 50,
           passing_grade: 70.00 }
   ATAU: POST /subject-configurations/carry-forward (salin dari tahun sebelumnya)
```

---

## API Endpoints

```
# Curricula
GET    /api/v1/curricula                              List kurikulum
GET    /api/v1/curricula?year_id={id}&grade_level=7   Filter per tahun + kelas
POST   /api/v1/curricula                              Buat kurikulum
PUT    /api/v1/curricula/{id}                         Update
GET    /api/v1/curricula/{id}                         Detail

# Learning Outcomes (CP/TP/ATP)
GET    /api/v1/curricula/{id}/learning-outcomes        Tree CP > TP > ATP
GET    /api/v1/learning-outcomes?type=cp&subject_id={id}  Filter
POST   /api/v1/learning-outcomes                      Buat entry
PUT    /api/v1/learning-outcomes/{id}                 Update
DELETE /api/v1/learning-outcomes/{id}                 Soft delete

# P5 Projects
GET    /api/v1/p5-projects                            List proyek P5
POST   /api/v1/p5-projects                            Buat proyek
PUT    /api/v1/p5-projects/{id}                       Update
POST   /api/v1/p5-projects/{id}/activate              Aktivasi
POST   /api/v1/p5-projects/{id}/complete              Tandai selesai

# Subjects (Master)
GET    /api/v1/subjects                               List mata pelajaran
GET    /api/v1/subjects?group=keagamaan&category=diniyah  Filter pesantren
GET    /api/v1/subjects/{id}                          Detail
POST   /api/v1/subjects                               Buat mapel
PUT    /api/v1/subjects/{id}                          Update

# Subject Configurations
GET    /api/v1/subject-configurations?year_id={id}&grade_level=7  Filter
POST   /api/v1/subject-configurations                 Buat konfigurasi
PUT    /api/v1/subject-configurations/{id}            Update
POST   /api/v1/subject-configurations/carry-forward   Copy dari tahun lalu
GET    /api/v1/subject-configurations/{id}/teachers   Guru yang mengajar
```

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Kurikulum per tahun ajaran, bukan global | Kurikulum berevolusi tiap tahun; memungkinkan perubahan tanpa mengganggu data historis |
| Self-referencing `parent_id` untuk CP/TP/ATP | Hierarki dinamis tanpa tabel terpisah; efisien untuk tree query |
| P5 sebagai tabel terpisah dari learning_outcomes | P5 punya struktur berbeda (tema, dimensi, timeline proyek) — bukan kompetensi mapel |
| KI/KD sebagai nullable fields | Backward compatible dengan K13 tanpa memaksa Kurikulum Merdeka menggunakan KI/KD |
| `subject_configurations` sebagai tabel terpisah | Jam pelajaran dan KKM berbeda per jenjang per tahun — tanpa tabel konfigurasi, harus duplikasi seluruh record |
| `weight_knowledge + weight_skill = 100` constraint | Mencegah konfigurasi bobot yang tidak valid di database level |
| Subject group CHECK dengan 16 pilihan | Standarisasi kelompok mapel; GIN index pada `_rels`/`_data` lebih efisien dengan bounded values |
| Carry-forward di awal tahun | Mengurangi beban admin yang harus setup ulang konfigurasi mapel setiap tahun |

---

## Integration Points

| Titik Integrasi | Keterangan |
|----------------|-----------|
| S011 (Student Grades) | `subject_id` FK di `student_grades`; SyncEngine update `_data.subject` saat nama mapel berubah |
| S021 (Schedule) | `credit_hours_per_week` dari konfigurasi mapel menjadi input alokasi slot jadwal |
| S022 (Assessment) | `assessment_type` (formatif/sumatif/UH/PTS/PAS) bergantung pada `curriculum_type` kurikulum aktif |
| S024 (Lesson Plan) | RPP/Modul Ajar ber-link ke `curriculum_id` dan `subject_id`; wajib align dengan ATP |
| S018 (Rapor) | `subject_group` dan `sort_order` menentukan urutan mapel di rapor; `curriculum_type` menentukan template |
| S020 → S019 | `subject_configurations.curriculum_id` (opsional) link ke kurikulum untuk memastikan mapel sesuai kurikulum |
| Dapodik | Struktur CP/KD mengikuti format yang bisa di-sync ke e-Rapor Kemendikbud |
