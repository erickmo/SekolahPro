# Student Academic — Nilai Mapel, Rapor, & Rekap Akademik

Dokumen ini menjelaskan sistem penilaian akademik siswa di SekolahPro: nilai per mata pelajaran, rekap akademik per semester, dan proses pembuatan rapor (report card). Ini adalah domain dengan prioritas tertinggi karena rapor adalah output paling tangible dari keseluruhan sistem.

---

## ADR References

| ADR | Judul |
|-----|-------|
| S004 | Student Academic Record |
| S011 | Subject Grade Detail |
| S018 | Rapor Generation |

---

## Posisi SekolahPro Rapor vs e-Rapor Kemdikbud

SekolahPro menyediakan sistem rapor internal yang lengkap. Untuk sekolah yang juga
wajib menggunakan e-Rapor Kemdikbud (sistem resmi Kemendikbud), clarification:

- **SekolahPro Rapor**: digunakan untuk rapor digital internal, cetak rapor fisik,
  arsip digital, dan akses orang tua via portal
- **e-Rapor Kemdikbud**: sistem terpisah Kemendikbud yang terintegrasi dengan Dapodik
- **Integrasi**: SekolahPro menyediakan export data nilai dalam format kompatibel e-Rapor
  (format CSV/Excel sesuai template Kemdikbud) untuk menghindari double-entry manual
- **Legal Standing**: Rapor yang dicetak dari SekolahPro adalah dokumen resmi sekolah.
  Validitas legal tergantung tanda tangan/stempel kepala sekolah dan kebijakan dinas setempat.

> **Rekomendasi Implementasi**: Operator melakukan input nilai SATU KALI di SekolahPro,
> kemudian export ke e-Rapor Kemdikbud jika diperlukan.

---

## Domain Entities

### 1. `subjects` — Master Mata Pelajaran

Lihat detail di [06-curriculum-subject.md](06-curriculum-subject.md). Ringkasan:

| Kolom | Keterangan |
|-------|-----------|
| `name`, `code` | Nama dan kode mapel |
| `subject_group` | agama / pkn / bahasa / matematika / ipa / ips / seni_budaya / pjok / prakarya / informatika / muatan_lokal / keagamaan / tahfidz / lainnya |
| `is_national` | TRUE jika ditetapkan Kemendikbud |
| `sort_order` | Urutan tampil di rapor |

### 2. `student_grades` — Nilai Per Siswa Per Mapel Per Semester

Sumber data utama untuk rapor. Satu record = satu siswa, satu mapel, satu semester.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `student_id` | UUID | NOT NULL, FK | |
| `subject_id` | UUID | NOT NULL, FK | |
| `academic_year_id` | UUID | NOT NULL, FK | |
| `class_room_id` | UUID | NOT NULL, FK | |
| `semester` | VARCHAR(10) | CHECK: ganjil/genap | |
| `score_knowledge` | NUMERIC(5,2) | nullable, 0-100 | Nilai Pengetahuan (KI-3 / Asesmen Sumatif) |
| `score_skill` | NUMERIC(5,2) | nullable, 0-100 | Nilai Keterampilan (KI-4 / Asesmen Formatif) |
| `score_attitude` | VARCHAR(2) | nullable, CHECK: SB/B/C/K | Nilai Sikap (non-numerik) |
| `final_score` | NUMERIC(5,2) | nullable, 0-100 | Nilai akhir yang masuk rapor |
| `description_knowledge` | TEXT | nullable | Deskripsi naratif Pengetahuan (Kurikulum Merdeka) |
| `description_skill` | TEXT | nullable | Deskripsi naratif Keterampilan (Kurikulum Merdeka) |
| `grade_letter` | VARCHAR(2) | nullable, CHECK: A/B/C/D/E | Predikat huruf |
| `teacher_id` | UUID | nullable | Guru pengajar (untuk atribusi) |

**UNIQUE** pada `(student_id, subject_id, academic_year_id, semester)`.

**Nilai Sikap (score_attitude):**
- `SB` = Sangat Baik
- `B` = Baik
- `C` = Cukup
- `K` = Kurang

Nilai sikap tidak bisa di-average karena bukan angka — ditampilkan sebagai predikat tersendiri di rapor.

> **Catatan Kurikulum:**
> - **K13**: nilai sikap (SB/B/C/K) wajib diisi per mapel oleh setiap guru
> - **Kurikulum Merdeka**: penilaian profil pelajar dilakukan melalui P5 projects
>   (`p5_projects` table, ADR-S019), BUKAN per mapel. Field `score_attitude` di
>   `subject_grade_details` tidak wajib diisi untuk siswa dengan kurikulum = 'MERDEKA'.
> - `SB/B/C/K` adalah kode representasi internal — kode resmi Permendikbud No. 23/2016
>   menggunakan A/B/C/D, tapi keduanya merujuk hal yang sama.

### 3. `student_academics` — Rekap Per Semester

Agregat dari `student_grades` dan `student_attendances`. Satu record per siswa per semester.

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `grade_average` | NUMERIC(5,2) | Rata-rata `final_score` semua mapel |
| `class_rank` | INT | Peringkat di kelas (dihitung manual, disimpan) |
| `total_students` | INT | Jumlah siswa di kelas (untuk konteks ranking) |
| `promotion_status` | VARCHAR(20) | promoted / retained / graduated / NULL |
| `days_present/sick/permitted/absent` | INT | Rekap absensi harian (dari S008) |

### 4. `rapor_templates` — Template Rapor Per Sekolah

Konfigurasi layout rapor per kurikulum dan jenjang.

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `name` | VARCHAR(100) | Nama template |
| `curriculum_type` | VARCHAR(20) | merdeka / k13 / diniyah / custom |
| `grade_levels` | JSONB array | Jenjang yang menggunakan template ini, e.g. `["7","8","9"]` |
| `header_config` | JSONB | Logo sekolah, nama sekolah, alamat |
| `body_config` | JSONB | Konfigurasi kolom nilai (knowledge/skill/attitude, deskripsi) |
| `footer_config` | JSONB | Catatan wali kelas, tanda tangan |
| `signature_config` | JSONB | Posisi dan ukuran stempel/tanda tangan digital |
| `is_active` | BOOLEAN | Template aktif/tidak |

### 5. `rapor_records` — Rapor Per Siswa Per Semester

Snapshot immutable dari seluruh data akademik siswa. Ini adalah dokumen resmi.

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `student_id`, `academic_year_id`, `class_room_id`, `template_id` | UUID | FK |
| `semester` | VARCHAR(10) | ganjil / genap |
| `student_snapshot` | JSONB | Biodata siswa saat generate |
| `guardian_snapshot` | JSONB | Nama ayah, ibu/wali |
| `grades_snapshot` | JSONB array | Nilai per mapel (semua field) |
| `attendance_snapshot` | JSONB | Rekap hadir/sakit/izin/alpha |
| `extracurricular_snapshot` | JSONB array | Ekskul + nilai |
| `discipline_snapshot` | JSONB | Poin pelanggaran + poin merit + catatan |
| `academic_snapshot` | JSONB | Rata-rata, ranking, promotion_status |
| `teacher_note` | TEXT | nullable — catatan wali kelas |
| `principal_note` | TEXT | nullable — catatan kepala sekolah |
| `status` | VARCHAR(20) | draft / generated / reviewed / finalized |
| `pdf_url` | TEXT | nullable — URL PDF di storage setelah generate |
| `generated_at` | TIMESTAMPTZ | nullable |
| `teacher_signature_url` | TEXT | nullable |
| `principal_signature_url` | TEXT | nullable |

**UNIQUE** pada `(student_id, academic_year_id, semester)` — satu rapor per siswa per semester.

---

## Business Rules

### Nilai Per Mapel

1. **Nilai terdiri dari 3 komponen:** Pengetahuan (score_knowledge), Keterampilan (score_skill), dan Sikap (score_attitude). Tidak semua komponen wajib diisi — bergantung pada kebijakan kurikulum sekolah.
2. **Nilai Sikap tidak numerik.** Menggunakan skala SB/B/C/K sesuai standar Kemendikbud — tidak bisa di-average dengan nilai angka.
3. **Final score** = weighted average dari score_knowledge dan score_skill, sesuai bobot di `subject_configurations.weight_knowledge` dan `weight_skill` (total 100%). Contoh: 50% pengetahuan + 50% keterampilan.
4. **Predikat huruf (grade_letter)** dihitung dari final_score:
   - A ≥ 90
   - B ≥ 80
   - C ≥ 70 (biasanya setara KKM)
   - D ≥ 60
   - E < 60
   (Threshold dikonfigurasi per sekolah — bukan hardcoded.)
5. **Deskripsi naratif** (description_knowledge, description_skill) wajib diisi untuk Kurikulum Merdeka — rapor naratif menggantikan angka.
6. **Guru bisa bulk input nilai** seluruh kelas per mapel sekaligus via `POST /student-grades/bulk`.
7. **Nilai unik per siswa per mapel per semester** — tidak ada duplikasi. Jika nilai perlu diubah, gunakan PUT.

### Rekap Akademik (S004)

8. **`grade_average`** = rata-rata `final_score` semua mapel yang tidak NULL. Nilai NULL (mapel tanpa nilai) tidak diikutkan dalam perhitungan.
9. **`class_rank`** dihitung manual oleh sistem saat semua nilai kelas sudah masuk — harus di-trigger secara eksplisit, bukan auto-computed.
10. **`promotion_status`** diisi oleh wali kelas atau admin saat proses kenaikan kelas:
    - `promoted` = naik kelas
    - `retained` = tinggal kelas
    - `graduated` = lulus (kelas terakhir)
11. **Rekap absensi di S004** bersumber dari S008 (daily attendance) yang di-aggregate setiap akhir semester.

### Rapor Generation

12. **Rapor wajib menggunakan snapshot data.** Semua data di-freeze saat generate — perubahan data sumber setelah generate tidak mempengaruhi rapor yang sudah dibuat.
13. **Alasan snapshot:** Rapor adalah dokumen resmi/hukum. Jika guru mengoreksi nilai setelah rapor dicetak, rapor lama tidak boleh berubah.
14. **Status rapor berjenjang:** draft → generated → reviewed → finalized.
15. **Rapor yang sudah `finalized` tidak bisa diedit atau di-regenerate.** Hanya bisa buat rapor baru (regenerate as draft).
16. **Wali kelas harus review** (`status = 'reviewed'`) sebelum kepsek finalize.
17. **Kepala sekolah yang mem-finalize** rapor — ini adalah kontrol kualitas akhir.
18. **Readiness check wajib dilakukan** sebelum generate batch — sistem menampilkan daftar siswa/mapel yang datanya masih belum lengkap.
19. **Batch generation menggunakan background job** — 500 rapor ~5 menit dengan 10 concurrent workers.
20. **Cetak ulang aman** — PDF tersimpan di object storage, bisa diakses kapan saja.

---

## Alur Kalkulasi Nilai (Assessment → Grade → Rapor)

```
[Assessment Scores dari S022]
  UH-1 (weight 1): 75     ← assessment_scores
  UH-2 (weight 1): 80
  PTS  (weight 2): 85
  PAS  (weight 3): 90

[Weighted Average]
  score_knowledge = (75×1 + 80×1 + 85×2 + 90×3) / (1+1+2+3) = 85.00
  score_skill     = (weighted avg skill assessments)

[Final Score via S020 Subject Configuration]
  final_score = (score_knowledge × weight_k + score_skill × weight_s) / 100
  Contoh 50/50: final_score = (85 × 50 + 88 × 50) / 100 = 86.50

[Grade Letter]
  86.50 → grade_letter = 'A' (jika threshold A ≥ 85)

[Average ke S004]
  grade_average = AVG(final_score) semua mapel yang ada nilainya

[Rapor Snapshot]
  grades_snapshot = [{
    subject: "Matematika",
    score_knowledge: 85.00,
    score_skill: 88.00,
    score_attitude: "B",
    final_score: 86.50,
    grade_letter: "A",
    description_knowledge: "Ahmad menunjukkan...",
    description_skill: "Ahmad mampu..."
  }, ...]
```

## Alur Generate Rapor (Batch)

```
1. Admin pilih: academic_year + semester + class_room_ids
2. GET /rapor/readiness → cek kelengkapan data
   → { missing: [{ student: "Ahmad", missing: ["Matematika", "B.Inggris"] }] }
3. Jika siap: POST /rapor/generate
   → Response: { job_id: "...", status: "queued", total: 32 }
4. Background job per siswa:
   a. Query S001 + S003 → student_snapshot, guardian_snapshot
   b. Query S011 (ordered by subject sort_order) → grades_snapshot
   c. Query S004 → attendance_snapshot + academic_snapshot
   d. Query S015 → extracurricular_snapshot
   e. Query S012 → discipline_snapshot
   f. Render PDF via template (wkhtmltopdf / Chromium headless)
   g. Upload PDF ke object storage → dapat file_url
   h. UPDATE rapor_record: pdf_url, generated_at, status='generated'
5. GET /rapor/jobs/{job_id} untuk polling progress
   → { status: "processing", progress: 54, total: 120 }
6. Wali kelas review + tambah teacher_note → status='reviewed'
7. Kepsek finalize → status='finalized' → LOCK
```

---

## Struktur Snapshot Rapor

```json
{
  "student_snapshot": {
    "full_name": "Ahmad Fadhil",
    "nis": "12345",
    "nisn": "0012345678",
    "gender": "L",
    "birth_place": "Jakarta",
    "birth_date": "2012-05-15",
    "class_room": "VII-A",
    "academic_year": "2025/2026"
  },
  "guardian_snapshot": {
    "father_name": "Budi Santoso",
    "mother_name": "Siti Aminah"
  },
  "grades_snapshot": [
    {
      "subject": "Matematika",
      "subject_group": "matematika",
      "score_knowledge": 85.50,
      "score_skill": 88.00,
      "score_attitude": "B",
      "final_score": 86.75,
      "grade_letter": "A",
      "description_knowledge": "Ahmad menunjukkan pemahaman yang baik...",
      "description_skill": "Ahmad mampu menyelesaikan soal..."
    }
  ],
  "attendance_snapshot": {
    "days_present": 95,
    "days_sick": 3,
    "days_permitted": 1,
    "days_absent": 1
  },
  "extracurricular_snapshot": [
    { "name": "Pramuka", "grade": "B", "description": "Aktif mengikuti..." }
  ],
  "discipline_snapshot": {
    "violation_points": 5,
    "merit_points": 15,
    "note": "Perilaku baik..."
  },
  "academic_snapshot": {
    "grade_average": 86.75,
    "class_rank": 5,
    "total_students": 32,
    "promotion_status": "promoted"
  }
}
```

---

## API Endpoints

```
# Subjects (Master)
GET    /api/v1/subjects                             List mata pelajaran
POST   /api/v1/subjects                             Buat mata pelajaran
PUT    /api/v1/subjects/{id}                        Update

# Student Grades
GET    /api/v1/students/{id}/grades                 Semua nilai (semua semester)
GET    /api/v1/students/{id}/grades?semester=ganjil&year_id={id}  Per semester
POST   /api/v1/student-grades/bulk                  Bulk input nilai (per mapel per kelas)
PUT    /api/v1/student-grades/{id}                  Update nilai individual
GET    /api/v1/class-rooms/{id}/grades/{semester}   Nilai kelas per semester (matriks guru)

# Student Academics
GET    /api/v1/students/{id}/academics              Riwayat per semester
GET    /api/v1/students/{id}/academics/{semester}   Detail satu semester
POST   /api/v1/student-academics                    Buat record
PUT    /api/v1/student-academics/{id}               Update

# Rapor Templates
GET    /api/v1/rapor-templates                      List template
POST   /api/v1/rapor-templates                      Buat template
PUT    /api/v1/rapor-templates/{id}                 Update

# Rapor Generation
POST   /api/v1/rapor/generate                       Generate batch (background job)
GET    /api/v1/rapor/jobs/{id}                      Cek progress batch
POST   /api/v1/rapor/preview/{student_id}           Preview 1 siswa (tanpa simpan)
GET    /api/v1/rapor/readiness                      Cek kelengkapan data sebelum generate

# Rapor Records
GET    /api/v1/students/{id}/rapor                  Rapor semua semester siswa
GET    /api/v1/rapor-records/{id}                   Detail rapor
PUT    /api/v1/rapor-records/{id}                   Update notes (sebelum finalize)
POST   /api/v1/rapor-records/{id}/review            Wali kelas review
POST   /api/v1/rapor-records/{id}/finalize          Kepsek finalize (lock)
GET    /api/v1/rapor-records/{id}/pdf               Download PDF
POST   /api/v1/rapor/bulk-download                  Download batch (ZIP)
```

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Snapshot strategi untuk rapor | Rapor adalah dokumen legal — tidak boleh berubah setelah dicetak meski data source berubah |
| Snapshot JSONB, bukan JOIN saat render | Render PDF = baca 1 row; tidak perlu query 7+ tabel setiap cetak ulang |
| Nilai Sikap sebagai predikat (SB/B/C/K) | Standar Kemendikbud; sikap tidak bisa di-rata-rata |
| 3 komponen nilai terpisah (knowledge/skill/attitude) | Kurikulum Indonesia membedakan pengetahuan, keterampilan, dan sikap sejak K13 |
| Deskripsi naratif | Wajib di Kurikulum Merdeka; rapor naratif menggantikan/melengkapi angka |
| Background job untuk batch rapor | 500 rapor tidak bisa diproses synchronous; job system dengan polling lebih UX-friendly |
| Finalize lock | Kontrol kualitas dua tahap: wali kelas → kepsek; mencegah perubahan tidak terotorisasi |
| Template-based rapor | Setiap sekolah punya branding dan layout berbeda; JSONB config lebih fleksibel dari hardcode |

---

## Integration Points

| Titik Integrasi | Keterangan |
|----------------|-----------|
| S022 (Assessment) | `assessment_scores` di-aggregate menjadi `student_grades.final_score` via event `AssessmentFinalized` |
| S004 (Academic Record) | `student_grades` di-aggregate menjadi `student_academics.grade_average` via event `GradeAggregated` |
| S008 (Attendance) | Absensi harian di-aggregate ke `student_academics.days_*` setiap akhir semester |
| S001 (Student) | Biodata siswa masuk ke `student_snapshot` rapor |
| S003 (Guardian) | Nama ayah/ibu masuk ke `guardian_snapshot` rapor |
| S012 (Discipline) | Poin disiplin masuk ke `discipline_snapshot` rapor |
| S015 (Extracurricular) | Penilaian ekskul masuk ke `extracurricular_snapshot` rapor |
| S019 (Curriculum) | Tipe kurikulum menentukan template rapor yang digunakan |
| S020 (Subject Config) | Bobot nilai (weight_knowledge/weight_skill) dan KKM/KKTP dari konfigurasi mapel |
| Object Storage | PDF rapor disimpan di S3/MinIO setelah generate |
| Dapodik | Nilai per mapel dan data rekap semester dilaporkan ke e-Rapor Kemendikbud |
