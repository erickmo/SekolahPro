# 09 — Manajemen Guru & Staff

Modul Guru & Staff mengelola seluruh siklus hidup tenaga pendidik dan kependidikan di SekolahPro, mulai dari absensi harian, perhitungan beban mengajar, penilaian kinerja (PKG), pengembangan profesi (PKB), manajemen cuti, penggajian, hingga penugasan penggantian guru. Modul ini sangat berorientasi pada regulasi Indonesia (PP 74/2008, Permendikbud 15/2018, Permenneg PAN & RB 16/2009) dan mendukung konteks pesantren (ustadz/ustadzah).

---

## ADR References

| ADR | Judul | Pattern |
|-----|-------|---------|
| ADR-S026 | Teacher Attendance (Absensi Guru & Staff) | Vernon |
| ADR-S027 | Teacher Workload (Beban Mengajar) | Vernon |
| ADR-S028 | Teacher Performance Evaluation (PKG) | Vernon |
| ADR-S029 | Professional Development (PKB) | Vernon |
| ADR-S030 | Leave Management (Cuti & Izin) | Vernon |
| ADR-S031 | Staff Payroll (Penggajian) | Vernon |
| ADR-S032 | Teacher Substitution (Penggantian Guru) | Vernon |

---

## Domain Entities

### Teacher Attendance (S026)

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `id` | UUID | PK |
| `teacher_id` | UUID | FK → teachers |
| `academic_year_id` | UUID | FK → academic_years |
| `attendance_date` | DATE | Satu record per hari per guru |
| `status` | VARCHAR(20) | `present`, `sick`, `permitted`, `absent`, `dinas_luar`, `cuti`, `libur` |
| `clock_in` | TIMESTAMPTZ | Waktu masuk (nullable untuk absent/cuti) |
| `clock_out` | TIMESTAMPTZ | Waktu keluar (nullable) |
| `late_minutes` | INT | Menit keterlambatan (0 jika tepat waktu) |
| `early_leave_minutes` | INT | Menit pulang lebih awal |
| `clock_in_method` | VARCHAR(20) | `fingerprint`, `face_recognition`, `gps`, `manual`, `qr_code` |
| `clock_in_location` | TEXT | Koordinat GPS atau ID mesin |
| `attachment_url` | TEXT | Surat dokter, surat tugas, dll |
| `validated_by` | UUID | User yang memvalidasi |

**Config Table (`teacher_attendance_configs`):**

| Field | Default | Keterangan |
|-------|---------|-----------|
| `work_start_time` | 07:00 | Jam mulai kerja |
| `work_end_time` | 14:00 | Jam selesai kerja |
| `late_tolerance_minutes` | 15 | Toleransi keterlambatan |
| `minimum_work_hours` | 7.0 | Jam kerja minimum |
| `work_days` | Sen–Jum | Hari kerja |

### Teacher Workload (S027)

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `id` | UUID | PK (`teacher_workloads`) |
| `teacher_id` | UUID | FK → teachers |
| `academic_year_id` | UUID | FK |
| `semester` | VARCHAR(10) | `ganjil` / `genap` |
| `teaching_hours` | NUMERIC(5,1) | JP mengajar tatap muka dari timetable |
| `additional_hours` | NUMERIC(5,1) | JP ekuivalen tugas tambahan |
| `total_hours` | NUMERIC(5,1) | `teaching_hours + additional_hours` |
| `minimum_required` | NUMERIC(5,1) | Default 24 JP; kepsek = 18 JP |
| `is_fulfilled` | BOOLEAN | `total_hours >= minimum_required` |
| `fulfillment_status` | VARCHAR(20) | `kurang`, `terpenuhi`, `lebih` |
| `dapodik_reported` | BOOLEAN | Sudah dilaporkan ke Dapodik |

**Workload Items (`teacher_workload_items`):**

| `item_type` | JP Ekuivalen |
|-------------|--------------|
| `mengajar` | 1 JP = 1 jam |
| `wali_kelas` | 2 JP |
| `pembina_ekskul` | 2 JP |
| `kepala_sekolah` | 18 JP |
| `wakil_kepsek` | 12 JP |
| `koordinator`, `panitia`, `tugas_tambahan` | Sesuai SK |

### Teacher Performance Evaluation / PKG (S028)

**`teacher_evaluations`:**

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `teacher_id` | UUID | Guru yang dievaluasi |
| `semester` | VARCHAR(10) | `ganjil` / `genap` |
| `evaluation_date` | DATE | Tanggal evaluasi |
| `score_pedagogik` | NUMERIC(5,2) | Skor kompetensi pedagogik (0–100) |
| `score_kepribadian` | NUMERIC(5,2) | Skor kompetensi kepribadian |
| `score_sosial` | NUMERIC(5,2) | Skor kompetensi sosial |
| `score_profesional` | NUMERIC(5,2) | Skor kompetensi profesional |
| `total_score` | NUMERIC(5,2) | Skor total (0–100) |
| `grade` | VARCHAR(20) | `amat_baik`, `baik`, `cukup`, `sedang`, `kurang` |
| `status` | VARCHAR(20) | `draft` → `self_assessment` → `peer_review` → `supervisor_review` → `completed` → `approved` |

**`teacher_evaluation_scores`** (detail per indikator per penilai):

| Field | Keterangan |
|-------|-----------|
| `competency_id` | FK → `teacher_evaluation_competencies` |
| `assessor_id` | FK → teachers (penilai) |
| `assessor_type` | `self`, `peer`, `supervisor` |
| `score` | 0–4 (skala PKG standar) |

**Bobot penilaian:** peer = 25%, supervisor = 75%; self = referensi saja.

### PKG — Rincian Kompetensi dan Indikator

Dokumen menggunakan "4 area kompetensi" sebagai struktur tingkat atas.
Dalam implementasi, masing-masing area memiliki sub-kompetensi dan indikator rinci:

**Guru Kelas/Mata Pelajaran (Permenneg PAN & RB No. 16/2009 jo. Permendiknas No. 35/2010):**
- Total: **14 kompetensi, 78 indikator**
- Area Pedagogik: 7 kompetensi
- Area Kepribadian: 3 kompetensi
- Area Sosial: 2 kompetensi
- Area Profesional: 2 kompetensi

**Guru BK/Konselor:**
- Total: **17 kompetensi, 76 indikator**

**Kepala Sekolah:**
- Menggunakan instrumen **PKKS** (Penilaian Kinerja Kepala Sekolah) — instrumen terpisah

**Implementasi:**
- Tabel `teacher_evaluation_competencies` menyimpan seluruh kompetensi dan indikator
- Seeded data default akan berisi template 14 kompetensi untuk guru kelas/mapel
- Admin sekolah dapat mengkustomisasi template jika ada instrumen tambahan dari dinas
- `evaluation_type` membedakan PKG reguler vs PKKS

**Grade PKG:**

| Grade | Skor |
|-------|------|
| Amat Baik | ≥ 91 |
| Baik | ≥ 76 |
| Cukup | ≥ 61 |
| Sedang | ≥ 51 |
| Kurang | < 51 |

### Professional Development / PKB (S029)

**`teacher_certifications`:** sertifikat pendidik, kompetensi, keahlian, ijazah, dll.

**`teacher_development_activities`:**

| Field | Keterangan |
|-------|-----------|
| `pkb_component` | `pengembangan_diri`, `publikasi_ilmiah`, `karya_inovatif` |
| `activity_type` | diklat, seminar, workshop, KKG/MGMP, penelitian, publikasi, buku, dll |
| `credit_points` | Angka kredit per kegiatan |
| `credit_approved` | Harus disetujui atasan |

**`teacher_credit_summaries`:**

| Field | Keterangan |
|-------|-----------|
| `current_rank` | `guru_pertama`, `guru_muda`, `guru_madya`, `guru_utama`, `non_pns` |
| `current_grade` | III/a s.d. IV/e atau `n/a` |
| `credit_total` | Total angka kredit |
| `credit_gap` | Selisih kredit menuju kenaikan pangkat |

### DUPAK (Daftar Usulan Penetapan Angka Kredit)

DUPAK adalah dokumen pengajuan angka kredit untuk kenaikan pangkat/jabatan fungsional guru.

**Komponen Angka Kredit PKB yang Dicatat di SekolahPro:**
- Sertifikat pelatihan (`professional_development_records.certificate_url`)
- Karya tulis/penelitian tindakan kelas (`professional_development_records.type = 'PTK'`)
- Pengembangan alat peraga/media (`professional_development_records.type = 'MEDIA'`)

**Proses:**
1. Guru upload bukti aktivitas PKB (sertifikat, laporan) via `professional_development_records`
2. Kepala Sekolah verifikasi dan berikan angka kredit per aktivitas
3. Sistem generate laporan DUPAK dalam format Excel yang bisa diisi dan diajukan ke dinas
4. DUPAK final ditandatangani kepala sekolah, diajukan ke dinas pendidikan kabupaten/kota

> **Catatan**: DUPAK diproses secara eksternal oleh Badan Kepegawaian daerah.
> SekolahPro hanya menyiapkan data dan format laporan — tidak terhubung langsung ke sistem BKN.

### Leave Management (S030)

**`leave_types`:** konfigurasi per sekolah untuk masing-masing jenis cuti.

| Kode | Nama | Max Hari PNS |
|------|------|-------------|
| `cuti_tahunan` | Cuti Tahunan | 12 hari/tahun |
| `cuti_sakit` | Cuti Sakit | Sesuai surat dokter |
| `cuti_melahirkan` | Cuti Melahirkan | 3 bulan |
| `cuti_besar` | Cuti Besar | 3 bulan (setelah 6 tahun) |
| `izin` | Izin | 1–3 hari |
| `tugas_belajar` | Tugas Belajar | 6 bln – 4 tahun |

**`leave_balances`:** saldo cuti per guru per tahun (bukan per tahun ajaran).

**`leave_requests`:**

| Field | Keterangan |
|-------|-----------|
| `status` | `draft` → `submitted` → `pending_approval` → `approved`/`rejected` → `completed` |
| `current_approval_level` | Level approval yang sedang pending (0–4) |
| `total_days` | Hari kerja (exclude weekend & libur) |

**`leave_approval_logs`:** audit trail setiap langkah approval.

### NUPTK dan Pelaporan Beban Mengajar

NUPTK (Nomor Unik Pendidik dan Tenaga Kependidikan) adalah identitas nasional guru.
- Field: `teachers.nuptk` (16 digit) — wajib untuk PNS dan guru yang memenuhi syarat
- NUPTK digunakan dalam pelaporan beban mengajar ke Dapodik
- Guru tanpa NUPTK tidak bisa mengklaim TPG (Tunjangan Profesi Guru)
- Verifikasi NUPTK via sync Dapodik (ADR-S055)

### Staff Payroll (S031)

**`payroll_configs`:** konfigurasi komponen gaji per sekolah (earning / deduction).

**`payroll_periods`:** periode bulanan dengan status `draft` → `calculating` → `calculated` → `reviewed` → `approved` → `disbursing` → `disbursed`.

**`payroll_entries`:** slip gaji per guru per bulan.

| Field | Keterangan |
|-------|-----------|
| `employee_type` | Snapshot: `pns`, `p3k`, `honorer`, `yayasan`, `kontrak` |
| `days_present`, `days_absent` | Dari S026 |
| `teaching_hours_per_week` | Dari S027 |
| `net_pay` | Gaji bersih setelah potongan |
| `pph21_amount` | Potongan pajak PPh 21 |

**`payroll_components`:** detail per komponen (15 kategori: gaji pokok, tunjangan sertifikasi, honor mengajar, BPJS, PPh, dll).

### Teacher Substitution (S032)

**`duty_schedules`:** jadwal piket harian (piket pagi, kelas, gerbang, upacara, siang, asrama, malam).

**`teacher_substitutions`:**

| Field | Keterangan |
|-------|-----------|
| `original_teacher_id` | Guru yang absen |
| `substitute_teacher_id` | Guru pengganti |
| `reason_type` | `sakit`, `cuti`, `izin`, `dinas_luar`, `tugas_belajar`, `terlambat`, `lain_lain` |
| `is_same_subject` | Apakah pengganti mengajar mapel yang sama |
| `status` | `pending` → `notified` → `accepted`/`declined` → `in_progress` → `completed` |

---

## Business Rules

1. **Clock-in hanya sekali per hari.** Satu record `teacher_attendances` per `(teacher_id, attendance_date)`. Duplicate clock-in ditolak dengan 409.
2. **Keterlambatan dihitung dari selisih `clock_in` dan `work_start_time`.** Jika masih dalam toleransi (`late_tolerance_minutes`), `late_minutes` = 0.
3. **Guru PNS wajib mengajar minimal 24 JP per minggu** (PP 74/2008). Kepala sekolah minimum 18 JP ekuivalen. P3K mengikuti aturan PNS.
4. **`total_hours` = `teaching_hours` (dari jadwal S021) + `additional_hours` (tugas tambahan).** Recalculation diperlukan setiap kali jadwal berubah.
5. **Tunjangan sertifikasi (TPG) hanya diberikan jika:** (a) guru bersertifikat, (b) beban mengajar ≥ 24 JP, (c) kehadiran memenuhi ambang batas minimum.
6. **PKG dievaluasi per semester.** Workflow: draft → self_assessment → peer_review → supervisor_review → completed → approved.
7. **Bobot PKG:** supervisor 75%, peer 25%. Self-assessment hanya referensi, tidak masuk perhitungan skor akhir.
8. **Predikat PKG ≥ "Baik" (skor ≥ 76) diperlukan** untuk pencairan tunjangan sertifikasi dan kenaikan pangkat PNS.
9. **Angka kredit PKB harus disetujui** (`credit_approved = true`) sebelum dihitung di `credit_total`.
10. **Cuti tahunan PNS:** maksimal 12 hari per tahun kalender. Carry-over maksimal 6 hari.

> **Koreksi Regulasi Cuti Tahunan:**
> Berdasarkan regulasi kepegawaian umum (PP No. 11/2017 untuk PNS), cuti tahunan
> yang tidak diambil **kadaluarsa** pada akhir tahun kalender — tidak ada carry-over
> otomatis.
>
> Kebijakan carry-over (jika ada) harus diatur dalam peraturan kepegawaian masing-masing
> sekolah/yayasan. Field `annual_leave_carryover_days` di `school_configurations` tersedia
> untuk sekolah yang memiliki kebijakan carry-over khusus (nilai default = 0).
11. **Cuti besar** hanya untuk PNS dengan masa kerja ≥ 6 tahun. Approval chain: wakil_kepsek → kepsek → yayasan.
12. **Approval cuti biasa** hanya memerlukan persetujuan kepala sekolah (1 level).
13. **Saat cuti diapprove**, sistem otomatis membuat record `teacher_attendances` dengan `status = 'cuti'` untuk setiap hari kerja dalam periode cuti.
14. **Pembatalan cuti yang sudah diapprove** akan me-refund saldo cuti dan menghapus (soft delete) record absensi yang dibuat otomatis.
15. **Gaji pokok PNS** tidak dikelola oleh SekolahPro — berasal dari APBN/APBD. SekolahPro mengelola tunjangan, honor, dan insentif dari dana sekolah.
16. **PPh 21** dihitung berdasarkan tarif progresif UU HPP No. 7/2021 dengan PTKP sesuai status kawin + tanggungan.
17. **MDR QRIS ditanggung merchant (sekolah)**, tidak boleh dibebankan ke orang tua per regulasi Bank Indonesia.
18. **Auto-suggest guru pengganti** menggunakan 4 level prioritas: (1) guru mapel sama + free period, (2) guru mapel serumpun + free period, (3) guru piket, (4) guru lain dengan beban mengajar terendah.
19. **Guru pengganti yang menolak** (decline) tidak dapat di-assign ulang ke slot yang sama tanpa override manual.
20. **Substitusi linked ke cuti (S030)** dapat dibuat lebih awal (proaktif) segera setelah cuti diapprove.

---

## State Machines

### Siklus Hidup Evaluasi PKG

```
draft
  → self_assessment   (guru submit self-assessment)
  → peer_review       (peer selesai menilai)
  → supervisor_review (kepsek/wakasek menilai)
  → completed         (semua skor dihitung, grade ditetapkan)
  → approved          (kepsek approve)
```

### Siklus Hidup Pengajuan Cuti

```
draft
  → submitted          (guru submit)
  → pending_approval   (menunggu approval berjenjang)
  → approved           (semua level approve — saldo dikurangi, absensi dibuat)
  → rejected           (ditolak di level manapun)
  → completed          (masa cuti selesai)

draft → cancelled      (guru batalkan sebelum submit)
approved → cancelled   (pembatalan setelah disetujui — saldo dikembalikan)
```

### Siklus Hidup Penggajian

```
draft
  → calculating  (hitung payroll)
  → calculated
  → reviewed     (bendahara/admin review)
  → approved     (kepsek/ketua yayasan approve)
  → disbursing
  → disbursed    (transfer selesai)
  → cancelled    (hanya dari draft)
```

### Siklus Hidup Substitusi

```
pending
  → notified    (notifikasi dikirim ke guru pengganti)
  → accepted    (guru pengganti terima)
  → declined    (guru pengganti tolak — re-assign)
  → in_progress (sedang mengajar)
  → completed   (selesai mengajar)
  → cancelled
```

### Jenjang Karir Guru PNS

```
Guru Pertama (III/a–III/b) → Guru Muda (III/c–III/d) →
Guru Madya (IV/a–IV/c) → Guru Utama (IV/d–IV/e)
```

Syarat naik: PKG ≥ "Baik" + angka kredit mencukupi.

---

## Key Decisions & Rationale

1. **Teacher Attendance terpisah dari Student Attendance (S008).** Guru melakukan self clock-in/clock-out dengan tracking keterlambatan per menit, berbeda dari siswa yang diabsen oleh wali kelas sekali per hari.
2. **Workload dihitung dari dua sumber:** teaching_hours dari timetable (S021) secara otomatis, additional_hours diinput manual dengan referensi SK. Recalculation diperlukan setiap perubahan jadwal.
3. **PKG menggunakan 3 tabel** (competencies, evaluations, scores) untuk mendukung multi-assessor dan detail per indikator yang diperlukan pelaporan BKN/Dapodik.
4. **Payroll menggunakan snapshot employee_type** saat perhitungan — perubahan status guru tidak mengubah payroll historis (immutable history).
5. **Gaji pokok PNS di luar scope SekolahPro** — SekolahPro hanya mengelola komponen yang dibayar langsung oleh sekolah (tunjangan, honor, insentif, potongan).
6. **Substitusi menggunakan tabel terpisah** dari schedule_entries (S021) karena memerlukan workflow accept/decline, task handover, dan audit trail yang tidak tersedia di schema jadwal.
7. **CQRS tidak digunakan untuk payroll** (berbeda dari payment gateway) — payroll adalah batch processing bulanan, bukan transaksi concurrent tinggi. Vernon cukup dengan trade-off eventual consistency yang acceptable.

---

## Integration Points

### Internal (Cross-Domain)

| Domain | Arah | Deskripsi |
|--------|------|-----------|
| S021 Timetable | → S027 | Jam mengajar tatap muka dihitung dari jadwal |
| S026 Attendance | → S031 | Data hadir/absen menjadi basis potongan/insentif payroll |
| S027 Workload | → S031 | Jam mengajar menjadi basis honor per jam untuk honorer |
| S028 PKG | → S029 | Area "kurang" di PKG menjadi prioritas kegiatan PKB |
| S030 Leave | → S026 | Approval cuti auto-create attendance status `cuti` |
| S030 Leave | → S032 | Leave approval auto-trigger substitution untuk slot yang terpengaruh |
| S031 Payroll | → S027 | Cek pemenuhan 24 JP sebelum TPG diberikan |
| S032 Substitution | → S044 | Notifikasi ke guru pengganti via notification system |
| ADR-012 Teachers | → semua | Master data guru sebagai referensi |

### Eksternal

| Sistem | Keterangan |
|--------|-----------|
| **Dapodik / BKN** | Pelaporan beban mengajar (S027), absensi ASN, PKG (S028), riwayat PKB (S029) per semester |
| **SIMPEG / SAPK** | Integrasi sistem kepegawaian PNS — deferred, bisa ditambah via API gateway |
| **Perangkat Fingerprint / Face Recognition** | Integrasi mesin absensi via middleware; MVP menggunakan REST API |
| **SMS/WhatsApp Gateway** | Notifikasi substitusi ke guru pengganti via S044 |
| **Bank / Payroll Service** | Disbursement gaji ke rekening guru — integrasi bank API di fase berikutnya |
| **BPJS Kesehatan & Ketenagakerjaan** | Laporan iuran BPJS dari data payroll |
