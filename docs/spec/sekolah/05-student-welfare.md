# Student Welfare — Keuangan SPP, Disiplin, Prestasi, Ekskul & Konseling BK

Dokumen ini menjelaskan domain kesejahteraan siswa di SekolahPro: pengelolaan keuangan (SPP dan tagihan), sistem tata tertib dan disiplin, pencatatan prestasi, kegiatan ekstrakurikuler, serta layanan bimbingan konseling (BK).

---

## ADR References

| ADR | Judul |
|-----|-------|
| S009 | Student Finance / SPP |
| S012 | Student Discipline / Tata Tertib |
| S013 | Student Achievement / Prestasi |
| S015 | Student Extracurricular |
| S017 | BK / Student Counseling Record |

---

## Domain Entities

### KEUANGAN (S009)

#### 1. `fee_types` — Master Jenis Tagihan

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `name`, `code` | VARCHAR | Nama dan kode unik per sekolah |
| `fee_category` | VARCHAR(20) | monthly / annual / one_time / incidental |
| `default_amount` | BIGINT | Template nominal (Rupiah, tanpa desimal) |
| `is_recurring` | BOOLEAN | TRUE untuk tagihan bulanan |
| `is_active` | BOOLEAN | |

#### 2. `student_invoices` — Tagihan Per Siswa

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `student_id`, `academic_year_id`, `fee_type_id` | UUID | FK |
| `invoice_no` | VARCHAR(50) | UNIQUE per sekolah — auto-generate |
| `period_month` | INT | 1-12 untuk tagihan bulanan; NULL untuk tahunan/one-time |
| `period_year` | INT | NOT NULL |
| `amount` | BIGINT | Nominal sebelum diskon |
| `discount` | BIGINT | Potongan (beasiswa, anak guru, dll.) |
| `total_amount` | BIGINT | `amount - discount` |
| `paid_amount` | BIGINT | Running total pembayaran |
| `status` | VARCHAR(20) | unpaid / partial / paid / overdue / waived |
| `due_date` | DATE | Tanggal jatuh tempo |

**UNIQUE** pada `(student_id, fee_type_id, period_year, period_month)` — tidak ada tagihan ganda per periode.

#### 3. `student_payments` — Record Pembayaran

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `invoice_id`, `student_id` | UUID | FK |
| `receipt_no` | VARCHAR(50) | UNIQUE — nomor kwitansi |
| `amount` | BIGINT | Jumlah yang dibayarkan (harus > 0) |
| `payment_method` | VARCHAR(20) | cash / transfer / debit / qris / va |
| `payment_date` | DATE | NOT NULL |
| `received_by` | UUID | NOT NULL — admin penerima |
| `note` | TEXT | nullable |

---

### DISIPLIN (S012)

#### 4. `discipline_rules` — Master Aturan Tata Tertib

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `code`, `name` | VARCHAR | Kode dan nama aturan |
| `rule_type` | VARCHAR(10) | violation / merit |
| `category` | VARCHAR(30) | Untuk violation: ringan / sedang / berat / sangat_berat. Untuk merit: akademik / sosial / kepemimpinan |
| `default_points` | INT | Poin default per kejadian |
| `is_active` | BOOLEAN | |

#### 5. `student_disciplines` — Catatan Per Kejadian

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `student_id`, `academic_year_id`, `rule_id` | UUID | FK |
| `incident_date` | DATE | NOT NULL |
| `semester` | VARCHAR(10) | ganjil / genap |
| `record_type` | VARCHAR(10) | violation / merit |
| `points` | INT | Poin aktual (negatif untuk violation, positif untuk merit) |
| `description` | TEXT | Deskripsi kejadian |
| `action_taken` | VARCHAR(30) | nullable — teguran_lisan / teguran_tertulis / panggilan_ortu / skorsing_1_hari / skorsing_3_hari / skorsing_1_minggu / dikeluarkan / pujian / sertifikat / hadiah |
| `action_note` | TEXT | nullable |
| `reported_by` | UUID | NOT NULL — guru/staff pelapor |
| `approved_by` | UUID | nullable — kepsek/wakasek approver |
| `approved_at` | TIMESTAMPTZ | nullable |
| `parent_notified` | BOOLEAN | FALSE default |
| `notified_at` | TIMESTAMPTZ | nullable |

---

### PRESTASI (S013)

#### 6. `student_achievements` — Data Prestasi Siswa

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `student_id`, `academic_year_id` | UUID | FK |
| `achievement_type` | VARCHAR(20) | academic / sports / arts / science / technology / religious / social / other |
| `title` | VARCHAR(255) | NOT NULL |
| `level` | VARCHAR(20) | school / district / city / province / national / international |
| `rank` | VARCHAR(30) | nullable — free text: "Juara 1", "Medali Emas", "Finalis" |
| `organizer` | VARCHAR(255) | nullable |
| `achievement_date` | DATE | NOT NULL |
| `certificate_url` | TEXT | nullable — link ke dokumen |
| `certificate_no` | VARCHAR(50) | nullable |

---

### EKSTRAKURIKULER (S015)

#### 7. `extracurriculars` — Master Kegiatan Ekskul

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `name`, `code` | VARCHAR | |
| `category` | VARCHAR(20) | sports / arts / science / technology / language / religious / social / scouts / other |
| `day_of_week` | VARCHAR(10) | nullable — monday/tuesday/.../saturday |
| `time_start`, `time_end` | TIME | nullable |
| `max_capacity` | INT | nullable — NULL = unlimited |
| `coach_id` | UUID | nullable — guru pembina |
| `is_mandatory` | BOOLEAN | TRUE untuk Pramuka (wajib di K13) |
| `is_active` | BOOLEAN | |

#### 8. `student_extracurricular_enrollments` — Pendaftaran Siswa

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `student_id`, `extracurricular_id`, `academic_year_id` | UUID | FK |
| `enrollment_status` | VARCHAR(20) | active / withdrawn / completed |
| `enrolled_date` | DATE | NOT NULL |
| `withdrawn_date` | DATE | nullable |

**UNIQUE** pada `(student_id, extracurricular_id, academic_year_id)`.

#### 9. `student_extracurricular_assessments` — Penilaian Per Semester

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `enrollment_id`, `student_id`, `extracurricular_id`, `academic_year_id` | UUID | FK |
| `semester` | VARCHAR(10) | ganjil / genap |
| `grade` | VARCHAR(2) | SB / B / C / K |
| `description` | TEXT | nullable — deskripsi naratif (wajib di Kurikulum Merdeka) |
| `assessed_by` | UUID | NOT NULL — guru pembina |

**UNIQUE** pada `(enrollment_id, semester)` — satu penilaian per enrollment per semester.

---

### KONSELING BK (S017)

#### 10. `counseling_cases` — Kasus Bimbingan

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `student_id`, `academic_year_id` | UUID | FK |
| `case_no` | VARCHAR(20) | UNIQUE — nomor kasus |
| `category` | VARCHAR(20) | academic / social / personal / career / behavioral / family / other |
| `title` | VARCHAR(255) | NOT NULL |
| `severity` | VARCHAR(10) | low / medium / high / critical |
| `status` | VARCHAR(20) | open / in_progress / referred / resolved / closed |
| `opened_date` | DATE | NOT NULL |
| `closed_date` | DATE | nullable |
| `resolution` | TEXT | nullable |
| `counselor_id` | UUID | NOT NULL — guru BK penanggungjawab |
| `referred_to` | TEXT | nullable — referral ke pihak eksternal |

#### 11. `counseling_sessions` — Sesi Per Kasus

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| `case_id`, `student_id` | UUID | FK |
| `session_date` | DATE | NOT NULL |
| `session_type` | VARCHAR(20) | individual / group / home_visit / parent_conference / referral |
| `duration_minutes` | INT | nullable |
| `notes` | TEXT | NOT NULL — inti catatan sesi |
| `recommendation` | TEXT | nullable |
| `counselor_id` | UUID | NOT NULL |
| `parent_present` | BOOLEAN | FALSE default |
| `follow_up_date` | DATE | nullable |
| `follow_up_note` | TEXT | nullable |

---

## Business Rules

### Keuangan SPP

1. **Arsitektur 3 layer:** Fee Type (template) → Invoice (tagihan) → Payment (bayar). Memisahkan concern dengan jelas.
2. **Nominal tagihan dalam Rupiah (BIGINT).** Tidak ada desimal — menghindari floating point issues.
3. **Tagihan di-generate bulk** per kelas, angkatan, atau semua siswa. Admin bisa adjust discount per siswa setelah generate.
4. **Cicilan didukung native** via `paid_amount`. Status otomatis berubah: unpaid → partial → paid.
5. **Payment amount tidak boleh melebihi sisa tagihan.** `paid_amount` tidak boleh melebihi `total_amount`.
6. **Overdue detection memerlukan cron job** yang mengubah status `unpaid` → `overdue` berdasarkan `due_date`.
7. **Nomor kwitansi unik** per sekolah. Auto-generate, tidak boleh sama.
8. **Fee type `period_month` NULL** untuk tagihan tahunan/one-time; 1-12 untuk bulanan.
9. **Waive tagihan** dimungkinkan (status `waived`) — untuk beasiswa penuh, keringanan khusus.
10. **Pembayaran tidak bisa dikembalikan (no refund)** di MVP — perlu enhancement terpisah.

### Disiplin & Tata Tertib

11. **Sistem poin kumulatif.** Poin pelanggaran dijumlah per semester atau tahun (dikonfigurasi per sekolah).
12. **Dua arah:** Violation (poin negatif) dan Merit (poin positif) ada dalam satu sistem.
13. **Eskalasi sanksi berbasis poin akumulasi:**

    | Akumulasi Poin | Sanksi Default |
    |---------------|----------------|
    | 5 poin | Teguran lisan |
    | 15 poin | Teguran tertulis |
    | 30 poin | Panggilan orang tua |
    | 50 poin | Skorsing 1 hari |
    | 75 poin | Skorsing 3 hari |
    | 100 poin | Skorsing 1 minggu |
    | 150 poin | Dikeluarkan (perlu approval kepsek) |

    *Threshold dapat dikonfigurasi per sekolah.*

14. **Sanksi `dikeluarkan` wajib di-approve kepsek** — field `approved_by` wajib diisi.
15. **Notifikasi orang tua didokumentasikan** via field `parent_notified` dan `notified_at`.
16. **Catatan disiplin masuk rapor** sebagai ringkasan poin dan catatan perilaku.
17. **Poin reset** — kebijakan reset per semester atau per tahun ajaran ditentukan per sekolah (belum di-define di ADR).

### Kebijakan Reset Poin Disiplin

Reset poin disiplin adalah kebijakan yang dapat dikonfigurasi per sekolah karena
setiap sekolah memiliki kebijakan berbeda:

| Opsi | Deskripsi |
|------|-----------|
| `per_semester` | Poin di-reset ke 0 setiap awal semester |
| `per_tahun_ajaran` | Poin di-reset ke 0 setiap awal tahun ajaran baru |
| `akumulatif` | Poin tidak pernah di-reset (akumulatif selama di sekolah) |
| `rolling_365_hari` | Hanya pelanggaran dalam 365 hari terakhir yang dihitung |

Konfigurasi di: `school_configurations.discipline_reset_policy`
Default: `per_tahun_ajaran`

> **Penting**: Kebijakan reset harus konsisten antara data disiplin (`student_disciplines`)
> dan rapor (nilai sikap). Saat reset terjadi, data historis tetap disimpan — yang berubah
> hanya `current_points` di `student_discipline_summary`.

### Prestasi

18. **Level prestasi mengikuti standar Dapodik:** school → district → city → province → national → international.
19. **Rank sebagai free text** (bukan enum) untuk fleksibilitas: "Juara 1", "Medali Emas", "Finalis", "Runner-up".
20. **Prestasi tim** (basket, paduan suara) — setiap anggota tim diinput sebagai record terpisah di MVP.
21. **Dokumen sertifikat disimpan di object storage** — `certificate_url` mereferensikan file yang mungkin juga ada di S010.

### Ekstrakurikuler

22. **Pramuka wajib** (`is_mandatory = true`) di Kurikulum 2013 untuk SD-SMP.
23. **Kapasitas ekskul** dikontrol via `max_capacity`. Enrollment ditolak jika kapasitas penuh.
24. **Penilaian ekskul per semester** menggunakan skala SB/B/C/K — sesuai format rapor Indonesia.
25. **Deskripsi naratif ekskul wajib** untuk Kurikulum Merdeka.
26. **Ekskul keagamaan** (Tahfidz, Kaligrafi, Tilawah) didukung via kategori `religious` — pesantren compliance.
27. **Withdrawal ekskul** dicatat dengan `withdrawn_date` — tidak dihapus (soft status).

### Konseling BK

28. **Privasi sangat ketat.** Access control domain-level:
    - Guru BK → CRUD kasus sendiri
    - Kepala Sekolah → Read-only semua kasus
    - Wali Kelas → Summary terbatas (jumlah kasus, tanpa detail)
    - Admin → Tidak ada akses
    - Orang Tua → Tidak ada akses digital
29. **Case-based:** Satu masalah = satu case dengan multiple sessions. Memungkinkan tracking progress.
30. **Referral terdokumentasi** — catatan rujukan ke psikolog atau pihak eksternal.
31. **Follow-up tracking** per sesi — setiap sesi bisa punya jadwal tindak lanjut.
32. **Status BK (backlog)** — modul ini dibangun on demand, bukan di MVP.

---

## Alur Penting

### Alur Pembayaran SPP

```
1. Admin setup fee_type "SPP Bulanan" dengan default_amount = 500.000
2. Admin generate tagihan:
   POST /student-invoices/generate
   Body: { fee_type_id, academic_year_id, target: "all_active", period_month: 7, period_year: 2025 }
   → System membuat 1 invoice per siswa aktif
3. Admin adjust discount per siswa (beasiswa):
   PUT /student-invoices/{id} → { discount: 200000 }
   → total_amount = 500000 - 200000 = 300000
4. Siswa/orang tua bayar:
   POST /student-payments
   Body: { invoice_id, amount: 300000, payment_method: "transfer", received_by: admin_id }
   → invoice.paid_amount = 300000
   → invoice.status = "paid" (karena paid_amount >= total_amount)
   → receipt_no auto-generated
5. Cetak kwitansi: GET /student-payments/{id}/receipt
```

### Alur Input Nilai Ekskul untuk Rapor

```
1. Guru pembina enroll siswa ke ekskul:
   POST /student-extracurricular-enrollments
   Body: { student_id, extracurricular_id, academic_year_id }
2. Akhir semester, guru beri penilaian:
   POST /student-extracurricular-assessments/bulk
   Body: { enrollments: [{ enrollment_id, grade: "B", description: "Aktif mengikuti..." }] }
3. Saat generate rapor (S018):
   Query assessments untuk semester ini → masuk extracurricular_snapshot
```

---

## API Endpoints

```
# SPP
GET    /api/v1/fee-types                            List jenis tagihan
POST   /api/v1/fee-types                            Buat jenis tagihan
POST   /api/v1/student-invoices/generate            Generate tagihan bulk
GET    /api/v1/students/{id}/invoices               Tagihan siswa
PUT    /api/v1/student-invoices/{id}                Update (discount/waive)
POST   /api/v1/student-payments                     Record pembayaran
GET    /api/v1/student-payments/{id}/receipt        Cetak kwitansi
GET    /api/v1/finance/summary                      Rekap keuangan per bulan
GET    /api/v1/finance/overdue                      Daftar tunggakan

# Disiplin
GET    /api/v1/discipline-rules                     List aturan tata tertib
POST   /api/v1/discipline-rules                     Buat aturan
GET    /api/v1/students/{id}/disciplines            Catatan disiplin siswa
GET    /api/v1/students/{id}/disciplines/summary    Rekap poin per semester
POST   /api/v1/student-disciplines                  Catat kejadian
POST   /api/v1/student-disciplines/{id}/approve     Approve sanksi (kepsek)
POST   /api/v1/student-disciplines/{id}/notify-parent  Tandai ortu sudah diberitahu

# Prestasi
GET    /api/v1/students/{id}/achievements           List prestasi siswa
POST   /api/v1/student-achievements                 Catat prestasi
PUT    /api/v1/student-achievements/{id}            Update
DELETE /api/v1/student-achievements/{id}            Soft delete

# Ekskul
GET    /api/v1/extracurriculars                     List ekskul
POST   /api/v1/student-extracurricular-enrollments  Enroll siswa
GET    /api/v1/students/{id}/extracurriculars       Ekskul siswa
POST   /api/v1/student-extracurricular-assessments/bulk  Bulk penilaian
GET    /api/v1/extracurriculars/{id}/members        Anggota ekskul

# BK
GET    /api/v1/counseling-cases                     List kasus (guru BK)
POST   /api/v1/counseling-cases                     Buat kasus
POST   /api/v1/counseling-sessions                  Catat sesi
PUT    /api/v1/counseling-cases/{id}/close          Tutup kasus
GET    /api/v1/students/{id}/counseling-summary     Summary (wali kelas, limited)
```

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| 3 layer keuangan (fee_type/invoice/payment) | Pemisahan template, tagihan, dan pembayaran membuat tracking tunggakan lebih mudah |
| BIGINT untuk Rupiah, bukan DECIMAL | Rupiah tidak punya desimal di konteks SPP; menghindari floating point errors |
| Paid_amount running total di invoice | Cicilan native tanpa tabel terpisah; query saldo cukup `total_amount - paid_amount` |
| Poin disiplin dua arah (violation + merit) | Gambaran perilaku siswa lebih lengkap; merit reward positif mendorong baik |
| Event-based disiplin (satu kejadian = satu record) | Historis per kejadian; audit trail lengkap; tidak ada aggregasi yang bisa dimanipulasi |
| Ekskul 3 layer (master/enrollment/assessment) | Granularity berbeda: enrollment per tahun, assessment per semester |
| BK case-based, bukan session-based | Satu masalah bisa 5-10 sesi; tanpa case entity tidak bisa track progress per masalah |
| Access control sangat ketat untuk BK | Data BK sensitif dan rahasia; tidak bisa diakses admin atau orang tua via sistem |

---

## Integration Points

| Titik Integrasi | Keterangan |
|----------------|-----------|
| S018 (Rapor) | `discipline_snapshot` (poin + catatan) dan `extracurricular_snapshot` (nilai + deskripsi) masuk ke rapor |
| S001 (Student) | Student status `active` menjadi syarat generate tagihan bulanan |
| Notification Module | Invoice overdue dan `parent_notified` trigger pengiriman notifikasi ke orang tua |
| Koperasi Module | Pembayaran SPP bisa terintegrasi dengan simpanan koperasi siswa |
| Accounting Module | `student_payments` di-feed ke jurnal keuangan sekolah (bukan double-entry di domain ini) |
| Dapodik | Prestasi level nasional dilaporkan ke Dapodik |
