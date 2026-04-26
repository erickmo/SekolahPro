# 16 — Alumni, Beasiswa & Kegiatan Sekolah

Modul ini mengelola tiga sub-domain pasca-siswa aktif: Alumni Management (S056) untuk direktori dan jaringan alumni, Scholarship Management (S057) untuk program beasiswa dengan workflow seleksi dan pencairan, serta School Event Management (S058) untuk perencanaan, pelaksanaan, dan dokumentasi kegiatan sekolah.

---

## ADR References

| ADR | Judul | Pattern |
|-----|-------|---------|
| ADR-S056 | Alumni Management | Vernon |
| ADR-S057 | Scholarship Management | Vernon |
| ADR-S058 | School Event Management | Vernon |

---

## Domain Entities

### Alumni Management (S056)

**`alumni`** — Profil alumni:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `student_id` | UUID | FK → students (1:1 per company) |
| `full_name` | VARCHAR(255) | Snapshot dari data siswa saat lulus |
| `graduation_year` | INT | Angkatan |
| `graduation_class` | VARCHAR(50) | Kelas terakhir (misal: "XII IPA 1") |
| `graduation_academic_year_id` | UUID | FK → academic_years |
| `certificate_no` | VARCHAR(50) | Nomor ijazah |
| `nisn` | VARCHAR(10) | NISN siswa saat lulus |
| `further_education_level` | VARCHAR(30) | `smp`, `sma`, `smk`, `d3`, `d4`, `s1`, `s2`, `s3`, `pesantren`, `kursus`, `kerja`, `other` |
| `further_education_name` | VARCHAR(255) | Nama institusi lanjutan |
| `current_occupation`, `current_company`, `current_position` | VARCHAR | Karir saat ini |
| `social_media` | JSONB | Flexible: Instagram, LinkedIn, Twitter, dll |
| `status` | VARCHAR(20) | `active`, `inactive`, `deceased`, `unreachable` |
| `last_updated_by` | VARCHAR(20) | `system` (auto), `admin`, atau `self` (alumni update sendiri) |

**`alumni_achievements`** — Prestasi alumni:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `achievement_type` | VARCHAR(30) | `academic`, `career`, `entrepreneurship`, `social`, `sports`, `arts`, `religious`, `other` |
| `level` | VARCHAR(20) | `local`, `regional`, `national`, `international` |
| `is_verified` | BOOLEAN | Verifikasi oleh admin sekolah |

**`alumni_events`** — Kegiatan alumni:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `event_type` | VARCHAR(30) | `reuni`, `homecoming`, `gathering`, `seminar`, `mentoring`, `fundraising`, `haflah`, `other` |
| `target_years` | JSONB | Array angkatan yang diundang `[2020, 2021]` |
| `is_all_alumni` | BOOLEAN | Untuk semua alumni |
| `registered_count` | INT | Jumlah terdaftar (denormalisasi) |
| `documentation` | JSONB | `{photos: [...], report_url, video_url}` |
| `status` | VARCHAR(20) | `draft` → `published` → `registration_open` → `ongoing` → `completed` / `cancelled` |

**`alumni_donations`** — Donasi/kontribusi alumni:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `donation_type` | VARCHAR(20) | `money`, `goods`, `service`, `scholarship_fund`, `infrastructure`, `other` |
| `amount` | BIGINT | Nominal (nullable untuk donasi barang/jasa) |
| `payment_method` | VARCHAR(20) | `cash`, `transfer`, `qris`, `va`, `other` |
| `status` | VARCHAR(20) | `pledged`, `received`, `acknowledged`, `returned` |

### Scholarship Management (S057)

**`scholarship_programs`** — Program beasiswa:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `code` | VARCHAR(30) | Kode unik per company |
| `scholarship_type` | VARCHAR(30) | `pip`, `kip`, `bos_siswa`, `yayasan`, `donatur_external`, `prestasi`, `hafidz`, `yatim`, `other` |
| `funding_source` | VARCHAR(30) | `pemerintah_pusat`, `pemerintah_daerah`, `bos`, `yayasan`, `alumni`, `perusahaan`, `lembaga`, `perorangan`, `other` |
| `amount_per_period` | BIGINT | Nominal per periode (> 0) |
| `period_type` | VARCHAR(20) | `monthly`, `semester`, `annual`, `one_time` |
| `max_recipients` | INT | Kuota penerima (nullable) |
| `min_gpa` | NUMERIC(4,2) | Minimal IPK/nilai rata-rata (nullable) |
| `min_attendance_rate` | NUMERIC(5,2) | Minimal persentase kehadiran (nullable) |
| `requires_financial_need` | BOOLEAN | Memerlukan bukti ketidakmampuan ekonomi |
| `is_renewable` | BOOLEAN | Dapat diperpanjang |

**`scholarship_applications`** — Pengajuan beasiswa:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `application_no` | VARCHAR(30) | Nomor pengajuan (unik) |
| `program_id` | UUID | FK → scholarship_programs |
| `student_id` | UUID | FK → students |
| `semester` | INT | 1 atau 2 |
| `gpa` | NUMERIC(4,2) | Nilai rata-rata saat pengajuan |
| `attendance_rate` | NUMERIC(5,2) | Persentase kehadiran |
| `kip_number` | VARCHAR(20) | Nomor KIP (nullable — khusus PIP/KIP) |
| `status` | VARCHAR(20) | `submitted` → `verified` → `approved` / `rejected` / `withdrawn` / `expired` / `renewed` |
| `is_renewal` | BOOLEAN | Perpanjangan dari semester sebelumnya |
| `previous_application_id` | UUID | Self-referensi ke pengajuan sebelumnya (nullable) |
| `approved_amount` | BIGINT | Nominal yang disetujui (nullable — bisa berbeda dari amount_per_period) |

**`scholarship_disbursements`** — Pencairan beasiswa:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `application_id` | UUID | FK |
| `invoice_id` | UUID | FK → student_invoices (nullable — untuk spp_offset) |
| `disbursement_no` | VARCHAR(30) | Nomor pencairan (unik) |
| `amount` | BIGINT | Nominal pencairan |
| `period_month` | INT | Bulan pencairan (nullable untuk one-time) |
| `period_year` | INT | Tahun pencairan |
| `disbursement_method` | VARCHAR(20) | `spp_offset`, `bank_transfer`, `cash`, `ewallet`, `other` |
| `status` | VARCHAR(20) | `pending` → `processing` → `disbursed` / `failed` / `returned` |

### School Event Management (S058)

**`school_events`** — Kegiatan sekolah:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `event_type` | VARCHAR(30) | 19 tipe: `upacara_bendera`, `class_meeting`, `study_tour`, `field_trip`, `pentas_seni`, `graduation_ceremony`, `parent_meeting`, `teacher_workshop`, `sports_day`, `science_fair`, `literacy_day`, `independence_day`, `haflah`, `khataman`, `wisuda_tahfidz`, `pesantren_kilat`, `isra_miraj`, `maulid_nabi`, `hari_santri`, `other` |
| `event_category` | VARCHAR(20) | `academic`, `extracurricular`, `ceremonial`, `religious`, `social`, `administrative`, `other` |
| `target_audience` | VARCHAR(20) | `all_students`, `specific_classes`, `all_teachers`, `parents`, `all_school`, `external`, `santri` |
| `target_classes` | JSONB | Array class_id jika `specific_classes` |
| `is_mandatory` | BOOLEAN | Wajib dihadiri atau opsional |
| `organizer_id` | UUID | FK → users (PIC/penyelenggara) |
| `committee` | JSONB | Panitia `[{user_id, name, role}]` |
| `budget_item_id` | UUID | FK → budget_items S050 (nullable) |
| `estimated_budget` | BIGINT | Anggaran rencana |
| `actual_budget` | BIGINT | Anggaran realisasi |
| `approval_status` | VARCHAR(20) | `draft` → `submitted` → `approved` / `rejected` / `revised` |
| `execution_status` | VARCHAR(20) | `planned` → `preparation` → `ongoing` → `completed` / `cancelled` / `postponed` |
| `calendar_event_id` | UUID | Link ke kalender akademik S023 (nullable) |

**`school_event_participants`** — Peserta dan kehadiran:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `participant_type` | VARCHAR(20) | `student`, `teacher`, `parent`, `staff`, `external` |
| `participant_id` | UUID | Nullable — eksternal tidak ada di DB |
| `participant_role` | VARCHAR(30) | `peserta`, `panitia`, `pembina`, `juri`, `narasumber`, `pendamping`, `tamu_undangan` |
| `registration_status` | VARCHAR(20) | `registered`, `confirmed`, `waitlisted`, `cancelled` |
| `attendance_status` | VARCHAR(20) | `hadir`, `tidak_hadir`, `izin`, `terlambat` |
| `check_in_at` / `check_out_at` | TIMESTAMPTZ | Waktu aktual |
| `rating` | INT | 1–5 (feedback peserta) |

**`school_event_documents`** — Dokumentasi kegiatan:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `document_type` | VARCHAR(20) | `photo`, `video`, `report`, `proposal`, `budget_realization`, `attendance_list`, `certificate`, `other` |
| `file_type` | VARCHAR(20) | jpg, png, mp4, pdf, doc, dll |
| `is_public` | BOOLEAN | Dapat dilihat publik (untuk website sekolah) |

---

## Business Rules

### Alumni

1. **Auto-konversi siswa lulus → alumni.** Saat student `status = 'graduated'`, sistem otomatis membuat record `alumni` dengan biodata yang di-snapshot dari data siswa saat itu.
2. **Data alumni adalah snapshot** — jika data siswa dikoreksi setelah lulus, alumni tidak auto-update. Ini disengaja untuk integritas data historis.
3. **Alumni dapat update profil sendiri** (`last_updated_by = 'self'`) — mekanisme self-service auth diperlukan (OTP atau link khusus).
4. **Satu alumni per siswa per company.** Unique constraint `(tenant_id, company_id, student_id)`.
5. **Prestasi alumni perlu diverifikasi** oleh admin sekolah sebelum tampil di public showcase.
6. **Data alumni mengandung privasi** — direktori alumni perlu consent management; tidak semua alumni mau data mereka publicly searchable.
7. **Event type `haflah`** tersedia khusus untuk pesantren.

### Beasiswa

8. **Satu pengajuan per (program, siswa, tahun ajaran, semester).** Unique constraint mencegah double application.
9. **Alur workflow:** submitted → verified (TU) → approved/rejected (kepsek).
10. **Eligibility check otomatis** dari data S011 (GPA) dan S008 (attendance_rate) dibandingkan `min_gpa` dan `min_attendance_rate` program.
11. **Kuota (`max_recipients`) dijaga** — program tidak bisa menerima penerima melebihi kuota.
12. **Renewal tracking via `previous_application_id`** membentuk chain pengajuan antar semester.
13. **`spp_offset` adalah metode pencairan utama** — beasiswa langsung mengurangi tagihan SPP di invoice S009.
14. **Beasiswa dapat dikombinasikan** (siswa punya multiple beasiswa aktif) — setiap beasiswa menghasilkan disbursement terpisah; total offset tidak boleh melebihi invoice.
15. **PIP/KIP tracking:** `kip_number` dan `pip_recipient_id` disimpan untuk reconciliation dengan data pemerintah.

### Kegiatan Sekolah

16. **Approval event menggunakan S047** — proposal kegiatan melalui workflow approval kepsek sebelum terlaksana.
17. **Event tidak bisa dimulai sebelum diapprove** — `execution_status = 'ongoing'` hanya bisa dari `approval_status = 'approved'`.
18. **Budget link ke RKAS (S050)** — setiap event berbudaya punya `budget_item_id` sebagai referensi alokasi anggaran.
19. **Calendar integration (S023)** — event yang diapprove otomatis masuk kalender akademik sekolah.
20. **Dokumentasi untuk akreditasi** — foto, laporan, dan outcomes tersimpan sebagai bukti pelaksanaan kegiatan (penting untuk visitasi BAN-S/M).
21. **Event type pesantren** (`haflah`, `khataman`, `wisuda_tahfidz`, `pesantren_kilat`, `isra_miraj`, `maulid_nabi`, `hari_santri`) tersedia khusus.
22. **Peserta polymorphic** — siswa, guru, orang tua, staf, dan tamu eksternal dapat didaftarkan dalam tabel yang sama.
23. **Unique participant per event** — `(event_id, participant_type, participant_id)` unik.

---

## Alur Konversi Siswa → Alumni

```
1. Siswa dinyatakan lulus (proses kelulusan/rapor)
   Student.status → 'graduated'

2. System auto-creates alumni record:
   - Snapshot: full_name, gender, birth_date, nisn dari S001
   - graduation_year dari academic_year
   - graduation_class dari student_class_placement (S014)
   - Link student_id

3. alumni.status = 'active'

4. Alumni dapat update profil sendiri (self-service):
   - Pendidikan lanjutan, karir, kontak, social media
   - last_updated_by = 'self'
```

## Alur Pencairan Beasiswa SPP Offset

```
1. Beasiswa approved (status = 'approved')
   approved_amount = Rp 500.000/bulan

2. Setiap bulan, generate disbursement:
   disbursement_method = 'spp_offset'

3. System find invoice SPP bulan ini (S009):
   student_invoices WHERE student_id + period_month

4. Apply offset:
   invoice.discount += disbursement.amount
   invoice.total_amount = invoice.amount - invoice.discount

5. Disbursement linked:
   disbursement.invoice_id = invoice.id
   disbursement.status = 'disbursed'

Contoh:
SPP = Rp 800.000
Beasiswa yayasan = Rp 500.000
Orang tua bayar = Rp 300.000 (remaining)
```

## Alur Kegiatan Sekolah (Event Lifecycle)

```
Draft (PIC buat proposal)
  ↓ submit
Submitted (menunggu approval kepsek via S047)
  ↓ kepsek approve
Approved (budget dialokasikan, calendar event dibuat di S023)
  ↓ persiapan
Preparation (panitia assign, peserta register)
  ↓ hari H
Ongoing (check-in peserta, dokumentasi foto/video)
  ↓ selesai
Completed (upload laporan, actual_budget, outcomes & lessons_learned)
```

---

## Key Decisions & Rationale

1. **Alumni sebagai domain terpisah** dari students (bukan hanya `status = 'graduated'`) karena alumni punya field berbeda (karir, pendidikan lanjutan, donasi, achievements) dan lifecycle berbeda dari siswa aktif.
2. **Biodata alumni adalah snapshot** saat lulus — tidak auto-update dengan perubahan student data. Ini memastikan integritas data historis. Koreksi data alumni dilakukan manual atau oleh alumni sendiri.
3. **Beasiswa punya workflow sendiri** (bukan hanya diskon di S009) karena butuh application workflow, eligibility check, approval, renewal tracking, dan audit trail yang tidak fit di domain keuangan murni.
4. **`spp_offset` sebagai metode pencairan default** — cara paling efisien karena langsung mengurangi tagihan tanpa transfer bank. Menghindari risiko dana tidak digunakan sesuai tujuan.
5. **School Events terpisah dari kalender (S023)** — kalender hanya menyimpan tanggal; S058 menyimpan workflow approval, peserta, anggaran, dan dokumentasi.
6. **19 event types** karena mencakup kebutuhan sekolah umum + pesantren secara komprehensif, lebih maintainable dari tabel terpisah per tipe.
7. **Dokumentasi event untuk akreditasi** — salah satu motivasi utama modul ini; bukti pelaksanaan kegiatan diperlukan saat visitasi BAN-S/M.

---

## Integration Points

### Internal

| Domain | Arah | Deskripsi |
|--------|------|-----------|
| ADR-S001 Students | → S056 | Auto-konversi siswa lulus → alumni |
| ADR-S014 Class Placement | → S056 | Graduation class dari placement terakhir |
| S008 Attendance | → S057 | Attendance rate untuk eligibility check beasiswa |
| S011 Subject Grades | → S057 | GPA untuk eligibility check beasiswa |
| S009 Student Finance | ← S057 | Disbursement spp_offset mengurangi invoice |
| S023 Academic Calendar | ← S058 | Event yang diapprove masuk ke kalender akademik |
| S047 Approval Workflow | ← S057, S058 | Approval beasiswa dan event menggunakan approval engine |
| S050 Budget/RKAS | → S058 | Budget item dialokasikan untuk event |
| S044 Notification | ← S056, S057, S058 | Notifikasi event, status beasiswa ke siswa/ortu |
| S056 Alumni | → S057 | Alumni bisa menjadi donatur untuk scholarship_fund |

### Eksternal

| Sistem | Keterangan |
|--------|-----------|
| **Dapodik** | Data alumni (pendidikan lanjutan) bisa menjadi input laporan akreditasi via S055 |
| **BAN-S/BAN-SM** | Summary pendidikan lanjutan dan prestasi alumni diperlukan saat visitasi akreditasi |
| **PIP/KIP Database (Kemendikbud)** | Reconciliation data penerima PIP/KIP — belum ada API resmi, masih manual upload |
| **CSR Platform / Donatur** | Alumni dan perusahaan eksternal bisa berkontribusi beasiswa — payment via S051 |
