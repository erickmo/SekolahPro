# Ujian & Penilaian (Exam & Assessment)

Dokumen ini menjelaskan sistem penilaian dan ujian di SekolahPro: jenis-jenis penilaian sesuai kurikulum Indonesia, input nilai harian dan ujian, perhitungan nilai akhir, dan mekanisme remedial. Domain ini adalah jembatan antara proses belajar-mengajar dan nilai yang masuk ke rapor.

---

## ADR References

| ADR | Judul |
|-----|-------|
| S022 | Exam & Assessment Management (Ujian & Penilaian) |

Domain ini berinteraksi erat dengan:
- S011 (Student Grades) — tujuan akhir kalkulasi
- S020 (Subject Configurations) — bobot penilaian (KKM/KKTP)
- S019 (Curriculum) — tipe kurikulum menentukan tipe assessment
- S018 (Rapor) — nilai akhir masuk ke rapor

---

## Konteks Sistem Penilaian Indonesia

### Kurikulum Merdeka

| Jenis Penilaian | Penjelasan |
|----------------|-----------|
| **Asesmen Formatif** | Penilaian proses belajar — tugas, kuis, observasi, portofolio. Tidak harus berupa angka; bisa deskripsi. |
| **Asesmen Sumatif** | Penilaian akhir per unit/bab — menunjukkan ketercapaian TP. Berupa angka atau deskripsi. |
| **Sumatif Tengah Semester (STS)** | Menggantikan UTS/PTS. Penilaian akhir paruh semester. |
| **Sumatif Akhir Semester (SAS)** | Menggantikan UAS/PAS. Penilaian akhir semester. |

### Kurikulum 2013 (K13) / KTSP

| Jenis Penilaian | Penjelasan |
|----------------|-----------|
| **Ulangan Harian (UH)** | Per bab/KD — diselenggarakan oleh guru. |
| **Penilaian Tengah Semester (PTS/UTS)** | Tengah semester — mencakup beberapa KD. |
| **Penilaian Akhir Semester (PAS/UAS)** | Akhir semester. |
| **Penilaian Akhir Tahun (PAT)** | Akhir tahun (semester genap) — menentukan kenaikan kelas. |

### Pesantren (ADR-009)

| Jenis Penilaian | Penjelasan |
|----------------|-----------|
| **Imtihan** | Ujian resmi madrasah diniyah — setara UH/PAS untuk mapel keagamaan. |
| **Setoran Tahfidz** | Hafalan Al-Quran per surah/ayat — predikat Arabic (mumtaz/jayyid jiddan/jayyid/maqbul/rasib). |

### Kategori Penilaian

| Kategori | Penjelasan | Nilai |
|---------|-----------|-------|
| `knowledge` | Pengetahuan — KI-3 (K13) / asesmen sumatif (Merdeka) | Angka 0-100 |
| `skill` | Keterampilan — KI-4 (K13) / asesmen formatif (Merdeka) | Angka 0-100 |
| `attitude` | Sikap — KI-1 & KI-2 (K13) / observasi (Merdeka) | Predikat SB/B/C/K |
| `tahfidz` | Hafalan khusus pesantren | Predikat Arab |

---

## Domain Entities

### 1. `assessments` — Definisi Ujian/Penilaian

Satu record = satu penilaian yang dibuat guru untuk satu mapel, satu kelas, satu semester.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `subject_id`, `academic_year_id`, `class_room_id`, `teacher_id` | UUID | NOT NULL, FK | |
| `name` | VARCHAR(200) | NOT NULL | e.g. "UH-1 Matematika Bab Persamaan Linear" |
| `assessment_type` | VARCHAR(30) | CHECK (14 tipe) | Lihat tabel di bawah |
| `assessment_category` | VARCHAR(20) | CHECK | knowledge / skill / attitude / tahfidz |
| `semester` | VARCHAR(10) | ganjil/genap | |
| `scheduled_date` | DATE | nullable | Tanggal ujian |
| `scheduled_start`, `scheduled_end` | TIME | nullable, start < end | Waktu ujian |
| `max_score` | NUMERIC(5,2) | 0-100 | Nilai maksimal (default 100) |
| `weight` | NUMERIC(5,2) | 0-10 | Bobot relatif — UAS biasanya 3, UH biasanya 1 |
| `description` | TEXT | nullable | Keterangan tambahan |
| `status` | VARCHAR(20) | CHECK | draft / scheduled / in_progress / completed / finalized |

**Tipe Assessment yang Didukung:**

| Kategori | Tipe | Keterangan |
|---------|------|-----------|
| Kurikulum Merdeka | `formatif` | Penilaian proses (tugas, kuis harian) |
| | `sumatif` | Penilaian akhir unit/bab |
| | `sumatif_tengah_semester` | STS (pengganti PTS) |
| | `sumatif_akhir_semester` | SAS (pengganti PAS) |
| K13 / KTSP | `ulangan_harian` | UH per KD |
| | `pts` | Penilaian Tengah Semester |
| | `pas` | Penilaian Akhir Semester |
| | `pat` | Penilaian Akhir Tahun |
| Pesantren | `imtihan` | Ujian madrasah diniyah |
| | `setoran_tahfidz` | Setoran hafalan Al-Quran |
| Generic | `tugas` | Pekerjaan rumah / tugas terstruktur |
| | `praktik` | Penilaian praktik/demonstrasi |
| | `proyek` | Project-based assessment |
| | `portofolio` | Penilaian portofolio |

### 2. `assessment_scores` — Nilai Per Siswa Per Penilaian

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `assessment_id`, `student_id` | UUID | NOT NULL, FK | |
| `score` | NUMERIC(5,2) | nullable, 0-100 | Nilai angka — nullable untuk attitude (hanya predikat) |
| `score_attitude` | VARCHAR(2) | nullable, CHECK: SB/B/C/K | Untuk kategori attitude |
| `description` | TEXT | nullable | Deskripsi naratif (wajib untuk formatif Merdeka) |
| `is_remedial` | BOOLEAN | default false | TRUE jika nilai hasil remedial |
| `original_score` | NUMERIC(5,2) | nullable, 0-100 | Nilai sebelum remedial (audit trail) |
| `surah_name` | VARCHAR(50) | nullable | Untuk setoran tahfidz |
| `ayat_from`, `ayat_to` | INT | nullable | Range ayat yang disetorkan |
| `tahfidz_grade` | VARCHAR(20) | nullable, CHECK | Predikat hafalan: mumtaz / jayyid_jiddan / jayyid / maqbul / rasib |

**UNIQUE** pada `(assessment_id, student_id)` — satu nilai per siswa per penilaian.

---

## Business Rules

### Pembuatan Penilaian

1. **Assessment dibuat oleh guru** untuk mapel dan kelas yang mereka ajar.
2. **Tipe assessment harus sesuai kurikulum aktif.** Formatif/sumatif untuk Merdeka; UH/PTS/PAS untuk K13.
3. **Bobot (weight) bersifat relatif.** UH weight 1, PTS weight 2, PAS weight 3 → total weight = 7.
4. **Max score default 100** tapi bisa dikustom. Nilai siswa dinormalisasi ke 0-100 saat kalkulasi.
5. **Status assessment berjenjang:** draft → scheduled → in_progress → completed → finalized.
6. **Assessment yang sudah `finalized` tidak bisa diedit.** Nilai dikunci setelah finalize.
7. **Jadwal ujian opsional** — tugas/portofolio tidak punya jadwal; UAS punya jadwal spesifik.
8. **Penilaian attitude tidak menghasilkan angka** — hanya predikat SB/B/C/K.
9. **Penilaian P5 (Profil Pelajar Pancasila) tidak menggunakan `assessment_scores`** — menggunakan entitas terpisah di modul kurikulum (S019 p5_projects).

### Input Nilai

10. **Bulk input nilai per kelas** — guru tidak perlu input satu per satu. Satu request untuk seluruh kelas.
11. **Nilai NULL diperbolehkan** — siswa yang tidak mengikuti penilaian (sakit, dispensasi) boleh NULL; tidak diikutkan dalam weighted average.
12. **Deskripsi naratif wajib untuk formatif Kurikulum Merdeka** — validasi di application layer.
13. **Setoran tahfidz** memiliki field khusus: `surah_name`, `ayat_from`, `ayat_to`, dan `tahfidz_grade`.

### Remedial

14. **Siswa yang nilai di bawah KKM/KKTP berhak mendapat remedial.** KKM/KKTP dari `subject_configurations.passing_grade`.
15. **Nilai asli disimpan di `original_score`** saat remedial dilakukan — tidak dihapus.
16. **Setelah remedial, `score` diupdate ke nilai remedial** dan `is_remedial = true`.
17. **Nilai remedial menggantikan nilai asli** dalam kalkulasi weighted average ke S011.

### Konfigurasi Nilai Maksimal Remedial

Beberapa sekolah membatasi nilai remedial maksimum = nilai KKM.

Konfigurasi di `subject_configurations`:

```
remedial_max_score_policy:
  type: ENUM
  values:
    - unlimited: nilai remedial bisa melebihi KKM (default)
    - capped_at_kkm: nilai remedial maksimum = nilai KKM/KKTP
  default: unlimited
```

Contoh: KKM = 75. Siswa remedial mendapat nilai 85.
- `unlimited`: nilai yang dicatat = 85
- `capped_at_kkm`: nilai yang dicatat = 75

> Kebijakan ini bervariasi per sekolah dan kadang per juknis daerah.
> Kepala Sekolah mengkonfigurasi per mapel atau per sekolah.

### Kalkulasi Nilai Akhir → S011

18. **Weighted average formula:**
    ```
    score_knowledge = Σ(score × weight) / Σ(weight)
                      untuk semua assessments dengan category='knowledge'
                      dan score IS NOT NULL
    ```
19. **Bobot pengetahuan vs keterampilan** dari `subject_configurations.weight_knowledge` dan `weight_skill`.
20. **Final score formula:**
    ```
    final_score = (score_knowledge × weight_knowledge + score_skill × weight_skill) / 100
    ```
21. **Nilai sikap tidak masuk final_score** — ditampilkan sebagai kolom terpisah di rapor.
22. **Kalkulasi di-trigger manual** via `POST /assessments/calculate-final` atau otomatis saat semua assessment untuk mapel tersebut di-finalize.
23. **Hasil kalkulasi di-write ke `student_grades` (S011)** via event `AssessmentFinalized` → eventual consistent.

### Threshold Nilai (KKM/KKTP)

24. **KKM (Kurikulum 2013)** = Kriteria Ketuntasan Minimal. Nilai di bawah KKM = belum tuntas → remedial.
25. **KKTP (Kurikulum Merdeka)** = Kriteria Ketercapaian Tujuan Pembelajaran. Bisa berupa angka atau deskripsi.
26. **KKM/KKTP dikonfigurasi per mapel per jenjang** di `subject_configurations.passing_grade`. Default 70.00.

---

## Alur Kalkulasi Nilai Akhir

### Contoh K13 (Matematika, KKM 70):

```
Assessment (knowledge category):
  UH-1  weight=1  score=75
  UH-2  weight=1  score=80
  PTS   weight=2  score=85
  PAS   weight=3  score=90

Weighted average knowledge:
  = (75×1 + 80×1 + 85×2 + 90×3) / (1+1+2+3)
  = (75 + 80 + 170 + 270) / 7
  = 595 / 7 = 85.00 → score_knowledge

Assessment (skill category):
  Praktik-1  weight=1  score=82
  Praktik-2  weight=1  score=88
  Weighted avg = (82+88)/2 = 85.00 → score_skill

Assessment (attitude category):
  Observasi  → score_attitude = "B"

Subject Configuration: weight_knowledge=50, weight_skill=50

final_score = (85.00×50 + 85.00×50) / 100 = 85.00

Passing grade check: 85.00 ≥ 70.00 (KKM) → TUNTAS
grade_letter: 85 → "A" (jika threshold A ≥ 85)
```

### Contoh Kurikulum Merdeka (Formatif + Sumatif):

```
Formatif (tidak masuk final, hanya feedback):
  Tugas-1 weight=1 description="Baik dalam presentasi"
  Tugas-2 weight=1 score=78

Sumatif:
  STS    weight=2  score=82
  SAS    weight=3  score=88

Weighted avg sumatif knowledge:
  = (82×2 + 88×3) / (2+3) = (164+264) / 5 = 85.6 → score_knowledge

Deskripsi naratif wajib untuk Merdeka:
  description_knowledge = "Ahmad menunjukkan pemahaman yang baik terhadap operasi pecahan..."
```

### Alur Remedial:

```
1. Sistem deteksi nilai < KKM:
   Ahmad: Matematika UH-1 score=55, KKM=70 → BELUM TUNTAS
2. Guru buat program remedial (opsional — bisa dalam bentuk tugas/tes remedial)
3. Input nilai remedial:
   POST /assessment-scores/{id}/remedial
   Body: { "score": 75 }
   → original_score = 55 (tersimpan sebagai audit)
   → score = 75 (nilai baru)
   → is_remedial = true
4. Re-kalkulasi weighted average menggunakan nilai remedial
5. Ahmad: 75 ≥ 70 → TUNTAS setelah remedial
```

---

## Struktur Jadwal Ujian PTS/PAS

Saat `assessment_type` adalah `pts`, `pas`, atau `sumatif_tengah_semester`/`sumatif_akhir_semester`, jadwal ujian dibuat eksplisit:

```json
{
  "name": "PAS Matematika Semester Ganjil 2025/2026",
  "assessment_type": "pas",
  "assessment_category": "knowledge",
  "scheduled_date": "2025-12-01",
  "scheduled_start": "07:30",
  "scheduled_end": "09:30",
  "max_score": 100,
  "weight": 3
}
```

API untuk melihat jadwal ujian semua kelas:
```
GET /api/v1/assessments/schedule?year_id={id}&semester=ganjil
→ Menampilkan jadwal PTS/PAS per mapel per kelas
```

---

## API Endpoints

```
# Assessments
GET    /api/v1/assessments?class_id={id}&subject_id={id}&semester=ganjil  List per kelas/mapel
POST   /api/v1/assessments                                   Buat penilaian
PUT    /api/v1/assessments/{id}                              Update
POST   /api/v1/assessments/{id}/schedule                     Set jadwal ujian
POST   /api/v1/assessments/{id}/finalize                     Finalize (lock nilai)
GET    /api/v1/assessments/schedule?year_id={id}             Jadwal ujian PTS/PAS

# Scores (Input Nilai)
POST   /api/v1/assessments/{id}/scores/bulk                  Bulk input nilai (1 kelas)
PUT    /api/v1/assessment-scores/{id}                        Update nilai individual
GET    /api/v1/assessments/{id}/scores                       Rekap nilai per penilaian
POST   /api/v1/assessment-scores/{id}/remedial               Input nilai remedial

# Student View
GET    /api/v1/students/{id}/assessment-scores?subject_id={id}&semester=ganjil
                                                             Semua nilai per mapel
GET    /api/v1/students/{id}/assessment-summary?semester=ganjil
                                                             Ringkasan per mapel (avg, min, max)

# Calculation
POST   /api/v1/assessments/calculate-final?class_id={id}&subject_id={id}&semester=ganjil
                                                             Hitung weighted average → update S011
```

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| 14 tipe assessment dalam 1 tabel | Semua tipe punya struktur sama; menghindari banyak tabel per kurikulum |
| Assessment terpisah dari S011 | S011 menyimpan nilai akhir per semester; S022 menyimpan histori penilaian individual |
| Weight sebagai angka relatif (1-10) | Lebih fleksibel dari persentase; UH=1, PAS=3 lebih intuitif bagi guru |
| `original_score` untuk remedial | Audit trail; tidak bisa manipulasi nilai asli; nilai pre-remedial tetap tersimpan |
| Null score diperbolehkan | Siswa sakit/absent tidak bisa dinilai; null tidak dimasukkan dalam weighted average |
| Kategori tahfidz dengan predikat Arab | Standar penilaian tahfidz pesantren (mumtaz/jayyid/maqbul/rasib) bukan skala 0-100 |
| Field tahfidz di generic table (nullable) | Hindari tabel terpisah untuk pesantren; nullable fields trade-off untuk simplicity |
| Kalkulasi ke S011 via event (async) | Eventual consistency acceptable; tidak blocking operasi input nilai |
| Finalize lock | Nilai tidak bisa berubah setelah periode penilaian resmi ditutup |

---

## Asesmen Nasional (ANBK)

ANBK (Asesmen Nasional Berbasis Komputer) menggantikan Ujian Nasional sejak 2021.
Terdiri dari:
- **AKM** (Asesmen Kompetensi Minimum): literasi membaca + numerasi
- **Survei Karakter**: profil pelajar Pancasila
- **Survei Lingkungan Belajar**: kondisi sekolah

### Jadwal
- Pelaksanaan: Oktober-November (SD/MI kelas 5, SMP/MTs kelas 8, SMA/SMK kelas 11)
- Peserta: sampling, bukan seluruh siswa
- Proktor dan pengawas: guru yang ditugaskan oleh kepala sekolah

### Peran SekolahPro
SekolahPro mendukung persiapan ANBK:
1. **Data siswa peserta** (sampling oleh Kemdikbud): export data siswa kelas target ke format CSV Dapodik
2. **Penjadwalan internal**: buat jadwal ANBK dalam `academic_calendar` sebagai event khusus
3. **Penugasan proktor**: assign guru sebagai `proctor` via `school_events` atau `teacher_assignments`

> **Catatan**: Sistem ANBK itu sendiri dikelola oleh Kemdikbud (aplikasi terpisah).
> SekolahPro tidak menyediakan soal ANBK atau pengolahan hasil ANBK — hanya mendukung
> persiapan data dan penjadwalan internal.

---

## Integration Points

| Titik Integrasi | Keterangan |
|----------------|-----------|
| S011 (Student Grades) | `assessment_scores` di-aggregate menjadi `score_knowledge`, `score_skill`, `final_score` di `student_grades` via event `AssessmentFinalized` |
| S020 (Subject Config) | `weight_knowledge`, `weight_skill`, `passing_grade` dari `subject_configurations` digunakan dalam formula kalkulasi |
| S019 (Curriculum) | `curriculum_type` menentukan tipe assessment yang valid (formatif/sumatif vs UH/PAS) |
| S021 (Schedule) | `scheduled_date` penilaian harus cross-check dengan kalender akademik (S023) untuk memastikan bukan hari libur |
| S018 (Rapor) | `student_grades` (hasil kalkulasi S022) menjadi sumber `grades_snapshot` di rapor |
| S025 (Teaching Journal) | Journal mencatat assessment harian sebagai bagian dari aktivitas pembelajaran yang terdokumentasi |
| Notification | Siswa/orang tua bisa mendapat notifikasi saat nilai remedial tersedia (future) |
