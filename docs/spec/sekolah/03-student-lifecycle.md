# Student Lifecycle — Penerimaan, Penempatan Kelas, Absensi, Kesehatan & Dokumen

Dokumen ini menjelaskan perjalanan hidup seorang siswa di sistem: dari pendaftaran (PPDB) hingga lulus atau keluar, termasuk penempatan kelas setiap tahun, pencatatan kehadiran harian, rekap akademik per semester, data kesehatan, dan manajemen dokumen.

---

## ADR References

| ADR | Judul |
|-----|-------|
| S008 | Daily Attendance Transaction |
| S004 | Student Academic Record |
| S005 | Student Health Record |
| S010 | Student Document Management |
| S014 | Student Class Placement / Mutasi Kelas |
| S016 | PPDB / Student Admission |

---

## State Machine Status Siswa

Status siswa merepresentasikan fase lifecycle di sistem:

```
                  ┌─────────┐
       Daftar →   │ active  │ ← Kembali dari cuti
                  └────┬────┘
          ┌────────────┼────────────┬──────────┐
          ▼            ▼            ▼          ▼
    ┌──────────┐  ┌──────────┐ ┌──────────┐ ┌─────────┐
    │graduated │  │transferred│ │ expelled │ │on_leave │
    │ (lulus)  │  │(pindah)  │ │(dikeluar)│ │ (cuti)  │
    └──────────┘  └──────────┘ └──────────┘ └─────────┘
```

| Status | Keterangan | `exit_date` | `exit_reason` |
|--------|-----------|------------|--------------|
| `active` | Siswa aktif bersekolah | NULL | NULL |
| `graduated` | Lulus dari sekolah | Tanggal wisuda | "Lulus" atau catatan |
| `transferred` | Pindah ke sekolah lain | Tanggal pindah | Alasan pindah |
| `expelled` | Dikeluarkan | Tanggal keputusan | Alasan dikeluarkan (wajib) |
| `on_leave` | Cuti sementara | Tanggal mulai cuti | Alasan cuti |

**Aturan:** Setiap transisi dari `active` ke status lain wajib mengisi `exit_date`. Transisi `on_leave` → `active` menghapus `exit_date`.

---

## Domain Entities

### 1. `admission_periods` — Periode PPDB

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `academic_year_id` | UUID | FK ke tahun ajaran |
| `name` | VARCHAR(100) | e.g. "PPDB 2025/2026" |
| `registration_start/end` | DATE | Jendela pendaftaran |
| `announcement_date` | DATE | Tanggal pengumuman hasil |
| `enrollment_deadline` | DATE | Deadline daftar ulang |
| `status` | VARCHAR(20) | draft / open / closed / announced / completed |

UNIQUE per `(tenant_id, company_id, academic_year_id)` — satu periode PPDB per sekolah per tahun.

### 2. `admission_tracks` — Jalur Seleksi

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `admission_period_id` | UUID | FK ke periode |
| `track_type` | VARCHAR(20) | zonasi / prestasi / afirmasi / pindahan / reguler |
| `quota` | INT | Daya tampung jalur ini |
| `registered_count` | INT | Counter denormalisasi (real-time) |
| `accepted_count` | INT | Counter denormalisasi |
| `requirements`, `selection_criteria` | TEXT | Persyaratan dan kriteria seleksi |

### 3. `applicants` — Data Calon Siswa

Data pendaftar yang **mirror** struktur Student + Guardian + Address. Setelah daftar ulang, data di-copy ke tabel-tabel tersebut.

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `registration_no` | VARCHAR(20) | UNIQUE — nomor pendaftaran auto-generate |
| Biodata (full_name, gender, birth_date, dll.) | — | Mirror S001 fields |
| Alamat (address, city, province, lat/lng) | — | Mirror S007 fields |
| Sekolah asal (previous_school_name, npsn) | — | Mirror S007 fields |
| Guardian utama (guardian_name, guardian_phone, dll.) | — | Mirror S003 fields |
| `selection_score` | NUMERIC(8,2) | Skor seleksi (jarak zonasi, nilai prestasi) |
| `selection_rank` | INT | Ranking dalam jalur |
| `selection_status` | VARCHAR(20) | registered / verified / accepted / rejected / waitlisted / enrolled / withdrawn |
| `enrolled_at` | TIMESTAMPTZ | Waktu daftar ulang |
| `student_id` | UUID | NULL sebelum enrolled; diisi setelah konversi ke siswa |

### 4. `student_class_placements` — Riwayat Penempatan Kelas

Satu record per penempatan kelas. Siswa SMP 3 tahun = minimal 3 records (kelas 7, 8, 9). Kasus mutasi bisa lebih.

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `student_id`, `academic_year_id`, `class_room_id` | UUID | FK |
| `placement_type` | VARCHAR(20) | initial / promotion / retention / transfer / major_selection |
| `effective_date` | DATE | NOT NULL — tanggal mulai (audit trail) |
| `end_date` | DATE | nullable — NULL selama masih aktif |
| `is_current` | BOOLEAN | Flag "kelas saat ini" |
| `major` | VARCHAR(30) | nullable — ipa / ips / bahasa / agama / umum (SMA kelas XI+) |
| `placed_by` | UUID | Admin yang melakukan penempatan |
| `note` | TEXT | nullable |

**UNIQUE** partial: `(student_id, academic_year_id) WHERE is_current = true` — satu siswa hanya boleh ada di satu kelas aktif per tahun ajaran.

### 5. `student_academics` — Rekap Akademik Per Semester

Agregat per semester. Satu record = satu siswa dalam satu semester.

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `student_id`, `academic_year_id`, `class_room_id` | UUID | FK |
| `semester` | VARCHAR(10) | ganjil / genap |
| `days_present` | INT | Total hari hadir |
| `days_sick` | INT | Total hari sakit |
| `days_permitted` | INT | Total hari izin |
| `days_absent` | INT | Total hari alpha |
| `grade_average` | NUMERIC(5,2) | Rata-rata nilai seluruh mapel |
| `class_rank` | INT | Peringkat di kelas |
| `total_students` | INT | Jumlah siswa di kelas (untuk konteks ranking) |
| `promotion_status` | VARCHAR(20) | promoted / retained / graduated / NULL (belum dievaluasi) |
| `notes` | TEXT | Catatan wali kelas |

**UNIQUE** pada `(student_id, academic_year_id, semester)` — tidak ada duplikasi data per semester.

### 6. `student_attendances` — Absensi Harian

Satu record per siswa per hari sekolah. Sumber data untuk perhitungan `student_academics.days_*`.

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `student_id`, `academic_year_id`, `class_room_id` | UUID | FK |
| `attendance_date` | DATE | NOT NULL |
| `semester` | VARCHAR(10) | Denormalisasi — ganjil/genap (mempercepat aggregation) |
| `status` | VARCHAR(20) | present / sick / permitted / absent |
| `note` | TEXT | Nomor surat izin, alasan sakit, dll. |
| `recorded_by` | UUID | NOT NULL — guru yang mengisi |
| `validated_by` | UUID | nullable — kepsek/wakasek |
| `validated_at` | TIMESTAMPTZ | nullable — timestamp validasi |

**UNIQUE** pada `(student_id, attendance_date)` — satu record per siswa per hari.

**Immutability Rule:** Setelah `validated_at` terisi, record tidak bisa diubah.

### 7. `student_health` — Rekap Kesehatan Berkala

Riwayat pemeriksaan fisik berkala. Setiap pengukuran = satu record baru (tidak overwrite).

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `student_id` | UUID | FK |
| `height_cm` | NUMERIC(5,1) | nullable |
| `weight_kg` | NUMERIC(5,1) | nullable |
| `eye_condition`, `hearing_condition`, `dental_condition` | VARCHAR(50) | nullable |
| `allergies` | TEXT[] | Array (bisa lebih dari satu) |
| `chronic_diseases` | TEXT[] | Array |
| `disability_type` | VARCHAR(50) | nullable — untuk Dapodik ABK |
| `disability_note` | TEXT | nullable |
| `insurance_type` | VARCHAR(20) | bpjs / swasta / none |
| `insurance_number` | VARCHAR(50) | nullable |
| `measured_at` | DATE | NOT NULL — tanggal pengukuran |
| `measured_by` | VARCHAR(255) | nullable — nama petugas UKS/dokter |

### 8. `student_documents` — Metadata Dokumen

Metadata dokumen siswa. File fisik disimpan di object storage (S3/MinIO).

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `student_id` | UUID | FK |
| `document_type` | VARCHAR(30) | CHECK: akta_lahir / kartu_keluarga / ktp_ortu / foto / ijazah / skhun / rapor / surat_pindah / surat_sehat / kartu_vaksin / sktm / dtks / other |
| `title` | VARCHAR(255) | NOT NULL |
| `file_name` | VARCHAR(255) | NOT NULL |
| `file_url` | TEXT | NOT NULL — URL ke storage |
| `file_size` | BIGINT | NOT NULL, max 10MB (10.485.760 bytes) |
| `mime_type` | VARCHAR(100) | NOT NULL |
| `uploaded_by` | UUID | NOT NULL — audit trail |
| `verified_by`, `verified_at` | UUID, TIMESTAMPTZ | nullable — verifikasi admin |
| `verification_note` | TEXT | nullable |

---

## Business Rules

### PPDB / Penerimaan

1. **Pendaftaran bersifat publik.** Endpoint `POST /public/apply` tidak memerlukan autentikasi — calon siswa/orang tua mendaftar sendiri.
2. **Nomor pendaftaran auto-generate** dan unik per sekolah per periode.
3. **5 jalur seleksi PPDB:** zonasi (jarak rumah), prestasi (nilai/rapor), afirmasi (ekonomi), pindahan (orang tua pindah), reguler (umum).
4. **Kuota per jalur wajib ditentukan.** Penerimaan tidak bisa melebihi kuota.
5. **Alur konversi applicant → student bersifat atomik.** Saat daftar ulang, semua record (student, guardian, address, previous_school, class_placement) dibuat dalam satu transaksi.
6. **Data biodata di-copy, bukan di-reference.** Setelah siswa dibuat, perubahan di `applicants` tidak mempengaruhi data siswa.

### Penempatan Kelas

7. **Satu siswa hanya boleh di satu kelas aktif per tahun ajaran.** Partial unique index pada `is_current = true` menegakkan aturan ini.
8. **Kenaikan kelas massal dilakukan akhir tahun ajaran.** Admin memilih kelas sumber, menandai promoted/retained per siswa, lalu sistem membuat placement baru.
9. **Siswa yang tinggal kelas (retained) mendapat placement baru di tingkat yang sama.** `placement_type = 'retention'`.
10. **Penjurusan SMA (kelas XI) menggunakan `placement_type = 'major_selection'`** dengan field `major` diisi.
11. **Mutasi tengah semester memerlukan update ke academic record dan absensi.**

### Absensi Harian

12. **Satu record per siswa per hari.** Guru wali kelas mengisi absensi untuk seluruh kelas via bulk endpoint.
13. **Bulk insert bersifat atomik.** Jika ada satu siswa invalid, seluruh batch ditolak.
14. **Status absensi mengikuti standar Dapodik:** present (hadir), sick (sakit), permitted (izin), absent (alpha).
15. **Absensi yang sudah divalidasi kepsek/wakasek tidak bisa diubah** — `validated_at` berfungsi sebagai lock.
16. **Agregasi ke academic record dilakukan async** via event `AttendanceAggregated` — eventual consistent.
17. **Denormalisasi `semester`** pada setiap record absensi untuk mempercepat query agregasi tanpa JOIN ke `academic_years`.

### Kesehatan

18. **Setiap pemeriksaan menghasilkan record baru** — tidak overwrite. Historis pertumbuhan tersedia.
19. **`measured_at` wajib diisi.** Setiap record kesehatan harus punya tanggal pengukuran.
20. **Alergi dan penyakit kronik disimpan sebagai PostgreSQL TEXT array** — bisa lebih dari satu, jarang di-query secara individual.
21. **Disability type untuk pelaporan Dapodik ABK (Anak Berkebutuhan Khusus).**
22. **BMI tidak disimpan** — dihitung di application layer dari height + weight.

### Dokumen

23. **File disimpan di object storage, bukan di database** — hanya metadata yang disimpan.
24. **Dokumen bersifat immutable.** Tidak ada overwrite; versi baru = upload baru.
25. **Batas ukuran file 10MB per dokumen.**
26. **Verifikasi dokumen dilakukan admin** — mengisi `verified_by` dan `verified_at` setelah memeriksa keaslian.

---

## Alur Penting

### PPDB: Applicant → Student

```
1. Orang tua POST /public/apply → Applicant dibuat (status: registered)
2. Admin verifikasi dokumen → status: verified
3. Admin/system run seleksi per jalur → selection_score, selection_rank dihitung
4. Pengumuman hasil → status: accepted / rejected / waitlisted
5. Orang tua daftar ulang → POST /applicants/{id}/enroll:
   a. CREATE students (S001)
   b. CREATE student_guardians + mapping (S003)
   c. CREATE student_addresses (S007)
   d. CREATE student_previous_schools (S007, jika ada)
   e. CREATE student_class_placements dengan placement_type='initial' (S014)
   f. UPDATE applicants.student_id, enrolled_at, status='enrolled'
   → SEMUA dalam satu database transaction
```

### Kenaikan Kelas Massal

```
1. Admin pilih academic_year baru + kelas sumber (e.g. VII-A 2024/2025)
2. System tampilkan daftar siswa → admin tandai promoted/retained per siswa
3. Admin pilih kelas tujuan untuk yang promoted (VIII-A 2025/2026)
4. System (atomik per siswa):
   a. UPDATE placement lama: end_date = hari ini, is_current = false
   b. CREATE placement baru: class_room_id baru, is_current = true, placement_type = 'promotion'/'retention'
   c. UPDATE students._data.class_room dengan data kelas baru
   d. UPDATE students._rels.class_room_id
   e. EMIT event ClassPlacementCreated
```

### Agregasi Absensi ke Academic Record

```
Setiap akhir semester:
SELECT
    student_id,
    COUNT(*) FILTER (WHERE status = 'present')    AS days_present,
    COUNT(*) FILTER (WHERE status = 'sick')        AS days_sick,
    COUNT(*) FILTER (WHERE status = 'permitted')   AS days_permitted,
    COUNT(*) FILTER (WHERE status = 'absent')      AS days_absent
FROM student_attendances
WHERE academic_year_id = $1 AND semester = $2
GROUP BY student_id
→ UPDATE student_academics (S004)
```

---

## API Endpoints

```
# PPDB
POST   /api/v1/public/apply                             Pendaftaran (no auth)
GET    /api/v1/public/check/{registration_no}           Cek status (no auth)
GET    /api/v1/applicants                               List pendaftar (admin)
PUT    /api/v1/applicants/{id}/verify                   Verifikasi dokumen
POST   /api/v1/applicants/select                        Jalankan seleksi
POST   /api/v1/applicants/{id}/enroll                   Daftar ulang → student

# Class Placement
GET    /api/v1/students/{id}/class-placements           Riwayat kelas
GET    /api/v1/students/{id}/current-class              Kelas saat ini
POST   /api/v1/student-class-placements/promote         Kenaikan kelas massal
POST   /api/v1/student-class-placements/transfer        Mutasi kelas

# Daily Attendance
POST   /api/v1/student-attendances/bulk                 Bulk input absensi kelas (1 request = 30-40 siswa)
GET    /api/v1/students/{id}/attendances                Riwayat absensi siswa
GET    /api/v1/students/{id}/attendances/summary        Rekap absensi per bulan/semester
GET    /api/v1/class-rooms/{id}/attendances/{date}      Absensi kelas per tanggal
PUT    /api/v1/student-attendances/{id}                 Update (sebelum validasi)
POST   /api/v1/student-attendances/validate             Validasi absensi

# Academic Record
GET    /api/v1/students/{id}/academics                  Riwayat per semester
GET    /api/v1/students/{id}/academics/{semester}       Detail satu semester
POST   /api/v1/student-academics                        Buat record
PUT    /api/v1/student-academics/{id}                   Update

# Health
GET    /api/v1/students/{id}/health                     Riwayat kesehatan (desc)
GET    /api/v1/students/{id}/health/latest              Data terbaru
POST   /api/v1/student-health                           Tambah record
PUT    /api/v1/student-health/{id}                      Update

# Documents
GET    /api/v1/students/{id}/documents                  List dokumen
GET    /api/v1/students/{id}/documents?type=ijazah      Filter by type
POST   /api/v1/student-documents                        Upload dokumen (multipart)
DELETE /api/v1/student-documents/{id}                   Soft delete
POST   /api/v1/student-documents/{id}/verify            Verifikasi dokumen
GET    /api/v1/student-documents/{id}/download          Download (presigned URL)
```

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Tabel `applicants` terpisah dari `students` | Tidak semua pendaftar diterima; perlu tahap seleksi sebelum jadi siswa |
| Data PPDB di-copy saat konversi, bukan di-reference | Data siswa harus stabil setelah terdaftar; perubahan data pendaftar tidak boleh mempengaruhi siswa |
| Partial unique index `is_current = true` untuk placement | Satu siswa hanya boleh satu kelas aktif; PostgreSQL partial index lebih efisien |
| Absensi per hari, bukan per sesi mapel | Mayoritas sekolah Indonesia absensi harian per kelas (wali kelas); absensi per sesi ada di jurnal mengajar (S025) |
| `semester` didenormalisasi di `student_attendances` | Mempercepat query agregasi akhir semester tanpa JOIN ke `academic_years` |
| Array `TEXT[]` untuk alergi dan penyakit | Bisa lebih dari satu; jarang di-query individual; menghindari over-normalization |
| Metadata-only untuk dokumen | File besar tidak boleh ada di database; presigned URL untuk akses |
| Absensi immutable setelah validasi | Integritas data laporan; mencegah manipulasi absensi setelah rekap |

---

## Integration Points

| Titik Integrasi | Keterangan |
|----------------|-----------|
| S001 (Student) | Class placement update `students._data.class_room` via SyncEngine |
| S004 (Academic) | Absensi harian di-aggregate ke `student_academics.days_*` setiap akhir semester |
| S011 (Grades) | `student_academics.grade_average` dihitung dari `student_grades` (S011) |
| S018 (Rapor) | `student_academics` + `student_attendances` menjadi `attendance_snapshot` dan `academic_snapshot` di rapor |
| S023 (Calendar) | Kalender akademik menentukan hari efektif untuk perhitungan persentase kehadiran |
| Dapodik | `disability_type` dari health record dilaporkan ke Dapodik untuk data ABK |
| Object Storage | Dokumen siswa disimpan di S3/MinIO; database hanya menyimpan `file_url` |
| Notification | Status pembayaran SPP dan kejadian disiplin memicu notifikasi ke orang tua |

---

## Proses Kelulusan dan Surat Keterangan Lulus (SKL)

### Alur Kelulusan
1. **Rapat Dewan Guru** — biasanya Mei/Juni, dipimpin Kepala Sekolah
   - Review nilai akhir + absensi + kondisi khusus per siswa
   - Putuskan: LULUS atau TIDAK LULUS per siswa
   - Status: `graduation_status` diupdate ke `graduation_decided`

2. **Penetapan Kelulusan**
   - Kepala Sekolah approve batch kelulusan
   - Status siswa diubah ke `graduated`
   - Trigger: `StudentGraduatedEvent`

3. **Penerbitan SKL Digital**
   - SKL di-generate otomatis saat status = `graduated`
   - Format PDF dengan QR code verifikasi
   - Ditandatangani digital oleh Kepala Sekolah
   - Berlaku sampai ijazah resmi keluar (Juli setelah ANBK selesai)

### Model Data SKL

```
student_graduation_records:
  id: UUID
  student_id: UUID
  academic_year_id: UUID
  graduation_date: DATE
  decided_by: UUID (kepala sekolah)
  decision_notes: TEXT (opsional, untuk kasus khusus)
  skl_issued_at: TIMESTAMP
  skl_pdf_url: TEXT
  ijazah_number: TEXT (diisi saat ijazah resmi keluar)
  status: ENUM(graduation_decided, graduated, skl_issued, ijazah_issued)
```

### Catatan
- Siswa tidak lulus: status kembali ke `active`, tetap di kelas yang sama tahun depan
- Sekolah swasta bisa menetapkan kriteria kelulusan sendiri
- Sekolah negeri mengikuti Permendikbud yang berlaku

---

## Mutasi Lintas Sekolah

### Pindah Keluar (Transfer Out)
- Status siswa diubah ke `transferred_out`
- Buat record `student_transfer_records` dengan field `surat_pindah_pdf_url`
- `exit_date` dan `exit_reason` wajib diisi

### Pindah Masuk (Transfer In)
- Status siswa = `active`, `placement_type = 'transfer'`
- Import data rapor sebelumnya jika tersedia
- Verifikasi NISN di Dapodik untuk siswa pindah masuk
- Data rapor sekolah asal: disimpan di `student_documents` tipe `rapor_sekolah_asal`
