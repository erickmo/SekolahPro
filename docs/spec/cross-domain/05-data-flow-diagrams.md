# 05 — Diagram Alur Data Kritis

Dokumen ini menjelaskan alur data end-to-end untuk skenario-skenario paling penting di SekolahPro. Setiap alur mencakup urutan proses, domain yang terlibat, mekanisme (SYNC/ASYNC), dan potensi risiko.

---

## 1. Pendaftaran Siswa Baru (End-to-End)

Alur lengkap: PPDB → penempatan kelas → setup SPP → Dapodik sync

```
FASE A: PENDAFTARAN ONLINE (S016)
──────────────────────────────────────────────────────────────────
Calon Siswa / Orang Tua
  │
  ▼
POST /api/v1/public/apply  (no auth required)
  │
  ├─ CREATE applicants {
  │    registration_no: "PPDB-2026-0001" (auto-generate)
  │    full_name, birth_date, gender, religion, ...
  │    address: {lat, lng untuk zonasi}
  │    guardian: {full_name, phone, nik}
  │    status: 'registered'
  │  }
  │
  └─ RESPONSE: registration_no + instruksi upload dokumen

Admin Sekolah verifikasi dokumen
  → PUT /api/v1/applicants/{id}/verify → status: 'verified'

System jalankan seleksi per jalur
  → POST /api/v1/applicants/select
    ├─ Zonasi: hitung jarak Haversine dari koordinat ke NPSN sekolah
    ├─ Prestasi: ranking berdasarkan nilai rapor
    └─ Afirmasi: validasi DTKS/SKTM

Admin umumkan hasil
  → applicant.selection_status: accepted / rejected / waitlisted


FASE B: DAFTAR ULANG → KONVERSI KE SISWA (ATOMIK)
──────────────────────────────────────────────────────────────────
Orang Tua konfirmasi daftar ulang
  │
  ▼
POST /api/v1/applicants/{id}/enroll  [DATABASE TRANSACTION]
  │
  ├─ 1. CREATE students {
  │       nis: auto-generate
  │       full_name, gender, birth_date, ...  (copy dari applicant)
  │       status: 'active'
  │       entry_date: today
  │       entry_type: 'new'
  │     }
  │
  ├─ 2. CREATE student_guardians {full_name, phone, ...}
  │      + CREATE student_guardian_map {student_id, guardian_id, is_primary: true}
  │
  ├─ 3. CREATE student_addresses {address, city, lat, lng, ...}
  │
  ├─ 4. CREATE student_class_placements {
  │       student_id, class_room_id (dari admin pilih kelas),
  │       placement_type: 'initial',
  │       is_current: true,
  │       effective_date: today
  │     }
  │
  ├─ 5. UPDATE class_rooms SET current_count = current_count + 1
  │
  ├─ 6. BUILD Vernon _rels + _data untuk student:
  │       _rels: {class_room_id, academic_year_id}
  │       _data: {class_room: {id, name, grade_level}, academic_year: {id, name, is_active}}
  │
  ├─ 7. UPDATE applicants SET student_id, enrolled_at, status='enrolled'
  │
  └─ 8. COMMIT transaction
  │
  ▼
POST-COMMIT (async):
  eventBus.Publish("events.students.student_activated", StudentActivatedEvent)


FASE C: SETUP SPP (S009) [ASYNC]
──────────────────────────────────────────────────────────────────
StudentActivatedHandler.Handle(event)
  │
  ├─ Lookup SPP template untuk grade_level siswa
  ├─ Generate student_invoices untuk bulan berjalan + bulan ke depan:
  │    {student_id, invoice_month, amount, due_date, status: 'unpaid'}
  └─ EMIT event: InvoiceCreated per invoice


FASE D: ONBOARDING KOPERASI [ASYNC — jika mode auto]
──────────────────────────────────────────────────────────────────
KoperasiEnrollmentHandler.Handle(StudentActivatedEvent)
  │
  ├─ CREATE nasabah_application {
  │    school_entity_id: student.id
  │    school_relation_type: 'student'
  │    full_name: student.full_name
  │    status: 'pending'  ← harus melalui approval
  │  }
  └─ Notify Supervisor Koperasi


FASE E: DAPODIK SYNC (S055) [MANUAL / PERIODIC]
──────────────────────────────────────────────────────────────────
Setiap akhir semester:
  Admin buat dapodik_sync_batch → VALIDATE → APPROVE → EXPORT → Upload ke Dapodik
  (lihat detail di 03-external-integrations.md)
```

**Risiko Kritis:**
- Transaksi DB di Fase B harus atomik — jika gagal di langkah 3-7, rollback total
- Event `StudentActivated` harus dikirim SETELAH commit DB, bukan di dalam transaksi
- Race condition: jika dua admin enroll applicant yang sama bersamaan → unique constraint pada `student.nis` akan menolak kedua operasi; hanya satu yang berhasil

---

## 2. Onboarding Nasabah Koperasi (End-to-End)

Alur: Nasabah baru → setup rekening → transaksi pertama

```
JALUR A: Dipicu oleh Event Sekolah (Auto Mode)
──────────────────────────────────────────────────────────────────
StudentActivatedEvent diterima Koperasi
  │
  ▼
CREATE nasabah_application (status: 'pending')

Supervisor Koperasi review aplikasi
  → PUT /api/v1/nasabah-applications/{id}/approve
  │
  ▼
[DATABASE TRANSACTION]
  ├─ 1. CREATE nasabah {
  │       member_number: "KOP-2026-JKT-000001" (auto-generate)
  │       school_entity_id: student.id
  │       status: 'active'
  │       kyc_level: 'basic'
  │       join_date: today
  │     }
  │
  ├─ 2. CREATE rekening (SIMPANAN POKOK) {
  │       nasabah_id, product_id (simpanan_pokok),
  │       account_number: "SP-2026-JKT-00000001"
  │       category: 'simpanan_pokok'
  │       balance: 0
  │       status: 'active'
  │     }
  │
  ├─ 3. CREATE rekening (SIMPANAN WAJIB) {
  │       account_number: "SW-2026-JKT-00000001"
  │       category: 'simpanan_wajib'
  │       balance: 0
  │       status: 'active'
  │     }
  │
  ├─ 4. UPDATE nasabah_application SET nasabah_id, status='approved'
  │
  └─ 5. COMMIT
  │
  ▼
POST-COMMIT (async):
  eventBus.Publish("events.koperasi.nasabah_approved", NasabahApprovedEvent)

Notify nasabah: "Selamat, keanggotaan Anda telah disetujui!"
Notify orang tua (jika siswa): "Anak Anda kini menjadi anggota koperasi"


TRANSAKSI PERTAMA: Setoran Simpanan Pokok
──────────────────────────────────────────────────────────────────
Teller buka sesi kerja (teller_session)
  │
  ▼
Nasabah datang ke counter koperasi
Teller: POST /api/v1/transaksi
  {
    "rekening_id": "018f-rekening-sp-uuid",
    "transaction_type": "setoran",
    "amount": 100000,  ← nominal simpanan pokok
    "teller_session_id": "018f-session-uuid"
  }
  │
  ▼
[DATABASE TRANSACTION]
  ├─ 1. INSERT transaksi {immutable record} + UPDATE rekening.balance
  ├─ 2. UPDATE teller_session.cash_in totals
  ├─ 3. Build jurnal double-entry (K015):
  │       DEBIT  : Kas Teller (101-001)
  │       KREDIT : Simpanan Pokok Nasabah (201-001)
  └─ 4. COMMIT
  │
  ▼
eventBus.Publish("events.koperasi.transaksi_created", TransaksiCreatedEvent)

Notify nasabah: "Setoran simpanan pokok Rp 100.000 berhasil. Saldo: Rp 100.000"
```

---

## 3. Alur Potongan Gaji Guru (Payroll Deduction)

```
SEKOLAH HR (setiap akhir bulan)
──────────────────────────────────────────────────────────────────
Bendahara Sekolah input/confirm data gaji bulan ini
  │
  ▼
System generate payroll batch record per guru
  │
  ▼
eventBus.Publish("events.teachers.payroll_data_ready", TeacherPayrollDataReadyEvent)
  {
    batch_id, period_month, period_year,
    items: [{teacher_id, gross_salary, take_home_pay}, ...]
  }


KOPERASI (async consumer)
──────────────────────────────────────────────────────────────────
PayrollDeductionHandler.Handle(event)
  │
  ├─ UNTUK SETIAP teacher dalam batch:
  │
  │   1. Lookup nasabah dengan school_entity_id = teacher.id
  │      → jika tidak ada: skip, log "guru belum terdaftar sebagai nasabah"
  │
  │   2. Cek payroll_authorization:
  │      → jika tidak ada: skip, log "tidak ada otorisasi"
  │      → jika revoked: skip
  │
  │   3. Hitung total potongan:
  │      simpanan_wajib.amount + SUM(angsuran yang jatuh tempo bulan ini)
  │      + tabungan_autorekt (jika dikonfigurasi)
  │
  │   4. Validasi: total_potongan <= take_home_pay
  │      → jika TIDAK: mark sebagai 'insufficient' + notify Supervisor
  │      → TIDAK otomatis potong parsial
  │
  │   5. [DATABASE TRANSACTION] per guru (atomik):
  │      ├─ INSERT transaksi (setoran simpanan wajib) + UPDATE rekening.balance
  │      ├─ UNTUK SETIAP angsuran: INSERT transaksi (pembayaran angsuran)
  │      │   + UPDATE angsuran.status = 'paid'
  │      ├─ UPDATE payroll_deduction.status = 'processed'
  │      └─ INSERT jurnal entry
  │
  └─ COMMIT per guru

eventBus.Publish("events.koperasi.payroll_processed", PayrollDeductionProcessedEvent)
  → Notify guru: "Potongan koperasi bulan ini: Rp X (simpanan wajib + angsuran)"


REKONSILIASI (jika ada 'insufficient' atau failure)
──────────────────────────────────────────────────────────────────
Supervisor menerima daftar potongan yang gagal
  │
  ├─ Tindak lanjut manual: hubungi guru
  ├─ Tandai sebagai 'tunggakan' (akan ditambahkan ke potongan bulan depan)
  └─ Atau: nasabah melunasi langsung ke counter
```

**Risiko:**
- Race condition jika payroll batch diproses dua kali (retry dari event bus)
  → Mitigasi: idempotency key = `batch_id + nasabah_id + period`

---

## 4. Tutup Buku Bulanan (Sekolah + Koperasi)

```
SEKOLAH — TUTUP BUKU KEUANGAN BULANAN
──────────────────────────────────────────────────────────────────
H-1 akhir bulan (cron job):
  ├─ Generate invoice SPP untuk bulan berikutnya (jika belum ada)
  ├─ Mark invoice yang melewati due_date sebagai 'overdue'
  └─ EMIT InvoiceOverdueEvent → notifikasi orang tua

H-0 akhir bulan:
  └─ Report keuangan SPP: total tagihan vs total terbayar per kelas/angkatan


KOPERASI — TUTUP BUKU BULANAN (K015)
──────────────────────────────────────────────────────────────────
H-0 akhir bulan (atau hari pertama bulan baru):

1. Hitung bunga/bagi hasil tabungan:
   ├─ General mode: saldo rata-rata × suku bunga harian × hari dalam bulan
   └─ Islamic mode: (profit riil bulan ini × nisbah) / total DPO

2. Hitung bagi hasil deposito:
   ├─ General: bunga fixed sesuai kontrak
   └─ Islamic: profit riil × nisbah indikatif (bisa berbeda dari estimasi)

3. Generate simpanan_wajib billing untuk semua nasabah aktif

4. Generate tagihan angsuran yang jatuh tempo bulan ini

5. Tutup accounting_period:
   ├─ Validasi jurnal: total debit = total kredit
   ├─ Generate trial balance
   ├─ Generate neraca bulanan
   └─ Update kas_harian.ending_balance

6. Mark rekening dormant (tidak transaksi > 12 bulan)

7. Update KPI metrics:
   ├─ NPL ratio (DPD > 90 hari / total outstanding)
   ├─ CAR (modal inti / ATMR)
   └─ Liquidity ratio


LAPORAN TERPADU (Sekolah + Koperasi — untuk Kepala Sekolah/Yayasan)
──────────────────────────────────────────────────────────────────
Dashboard Executive (read-only, materialized view):
  mv_school_dashboard: total siswa aktif, kehadiran %, SPP collection rate
  mv_koperasi_summary: total aset, total pinjaman, NPL %, kas
```

---

## 5. Generasi Rapor Semester

Rapor adalah output paling penting domain Sekolah — mengaggregasikan data dari 6+ sub-domain.

```
FASE 1: DATA COLLECTION (selama semester berlangsung)
──────────────────────────────────────────────────────────────────
Setiap hari sekolah:
  └─ Guru wali kelas input absensi (S008): present/sick/permitted/absent

Setiap assessment/ulangan:
  └─ Guru input nilai (S022 → S011): nilai per mapel per siswa

Setiap kejadian:
  └─ BK input prestasi (S013) dan pelanggaran (S012)
  └─ Guru ekskul input keaktifan (S015)


FASE 2: AKHIR SEMESTER — KALKULASI
──────────────────────────────────────────────────────────────────
Admin/Wali Kelas trigger "Hitung Nilai Akhir":
  │
  ▼
1. Agregasi absensi harian → student_academics.days_*
   SELECT status, COUNT(*) FROM student_attendances
   WHERE student_id = ? AND semester = ?
   GROUP BY status
   → UPDATE student_academics SET days_present, days_sick, ...

2. Hitung nilai akhir per mapel:
   (nilai_tugas × bobot_tugas) + (nilai_uts × bobot_uts) + (nilai_uas × bobot_uas)
   → UPDATE student_grades.final_score per mapel

3. Hitung rata-rata per siswa:
   AVG(student_grades.final_score) WHERE student_id = ? AND semester = ?
   → UPDATE student_academics.grade_average

4. Hitung peringkat kelas:
   RANK() OVER (PARTITION BY class_room_id ORDER BY grade_average DESC)
   → UPDATE student_academics.class_rank, total_students


FASE 3: GENERATE RAPOR (S018)
──────────────────────────────────────────────────────────────────
Admin trigger generate rapor per kelas/semua kelas:
  │
  ▼
[UNTUK SETIAP SISWA dalam kelas] (bisa di-parallelize):

  a. FETCH semua data yang dibutuhkan:
     ├─ students: biodata (NIS, nama, kelas)
     ├─ student_guardians: nama orang tua/wali
     ├─ student_academics: rekap kehadiran + rata-rata + ranking
     ├─ student_grades: nilai per mapel (JOIN ke subjects)
     ├─ student_disciplines: catatan pelanggaran semester ini
     ├─ student_achievements: prestasi semester ini
     └─ student_health (latest): data kesehatan terbaru

  b. CREATE student_rapor {
       snapshot_student: {...}    ← frozen copy data siswa saat ini
       snapshot_guardian: {...}   ← frozen copy data wali
       semester_data: {
         attendances: {...},
         grades: [{subject, score, predicate}, ...],
         grade_average, class_rank, total_students,
         disciplines: [...],
         achievements: [...],
         notes_homeroom_teacher: "..."
       }
       status: 'draft'
     }

  c. Wali kelas review + isi catatan → status: 'finalized'

  d. Kepala Sekolah approve → status: 'published' + signature_url

  e. Generate PDF (async):
     └─ Render template rapor dengan logo sekolah + TTD kepala sekolah
     └─ Upload PDF ke Object Storage
     └─ UPDATE rapor.pdf_url


FASE 4: DISTRIBUSI RAPOR
──────────────────────────────────────────────────────────────────
eventBus.Publish("events.rapor.rapor_generated", RaporGeneratedEvent)
  │
  ├─ Notify orang tua: "Rapor semester [Ganjil/Genap] sudah tersedia"
  │   (via WhatsApp + in-app notification)
  │
  └─ Parent portal S042: orang tua bisa view/download PDF
```

**Risiko:**
- Jika data absensi belum complete saat rapor digenerate → rapor tidak akurat
  → Mitigasi: validasi completeness sebelum generate (cek hari sekolah vs records yang ada)
- Snapshot data dalam rapor harus IMMUTABLE setelah published
  → Mitigasi: `published_at` berfungsi sebagai lock; tidak ada UPDATE setelah published
- Generate PDF concurrent untuk 300 siswa bisa menyebabkan memory spike
  → Mitigasi: queue dengan concurrency limit (max 10 parallel PDF generation)

---

## 6. Akhir Tahun Ajaran (Kenaikan Kelas)

Alur paling kompleks — mempengaruhi seluruh domain Sekolah dan berdampak ke Koperasi.

```
PRE-REQUISITES:
  ✓ Rapor semester genap sudah published (S018)
  ✓ Academic year baru sudah dibuat (status: planning)
  ✓ Kelas-kelas untuk tahun baru sudah dibuat


FASE 1: EVALUASI KENAIKAN KELAS
──────────────────────────────────────────────────────────────────
Admin/Wali Kelas buka halaman Kenaikan Kelas:
  └─ System tampilkan daftar siswa aktif beserta:
       grade_average, days_absent/sick/permitted, discipline_count

Admin tandai status per siswa:
  ├─ promoted   → naik ke kelas berikutnya
  ├─ retained   → tinggal kelas (tidak naik)
  └─ graduated  → lulus (kelas IX / XII)


FASE 2: EKSEKUSI KENAIKAN KELAS [PER SISWA, SEMI-ATOMIK]
──────────────────────────────────────────────────────────────────
FOR EACH siswa:
  [DATABASE TRANSACTION]
  │
  ├─ 1. UPDATE placement lama:
  │       is_current = false
  │       end_date = tanggal hari ini
  │
  ├─ 2. CREATE placement baru:
  │       academic_year_id = tahun_baru.id
  │       class_room_id = kelas_tujuan (dipilih admin)
  │       placement_type = 'promotion' / 'retention' / 'major_selection'
  │       is_current = true
  │       effective_date = start_date tahun baru
  │
  ├─ 3. UPDATE students._rels.class_room_id = kelas_baru.id
  │      UPDATE students._rels.academic_year_id = tahun_baru.id
  │      UPDATE students._data.class_room = {id, name, grade_level}
  │      UPDATE students._data.academic_year = {id, name, is_active}
  │
  ├─ 4. UPDATE class_rooms (lama) SET current_count = current_count - 1
  │      UPDATE class_rooms (baru) SET current_count = current_count + 1
  │
  ├─ 5. Jika graduated:
  │       UPDATE students.status = 'graduated'
  │       UPDATE students.exit_date = tanggal wisuda
  │
  └─ COMMIT
  │
  ▼
eventBus.Publish("events.students.student_promoted", StudentPromotedEvent)


FASE 3: AKTIVASI TAHUN AJARAN BARU
──────────────────────────────────────────────────────────────────
Admin trigger: POST /api/v1/academic-years/{id}/activate
  │
  ▼
[DATABASE TRANSACTION]
  ├─ UPDATE academic_years SET is_active = false WHERE company_id = ? AND is_active = true
  ├─ UPDATE academic_years SET is_active = true WHERE id = {baru}
  └─ COMMIT
  │
  ▼
eventBus.Publish("events.academic_years.academic_year_activated", AcademicYearActivatedEvent)
  │
  ▼
SyncEngine:
  ├─ Update _data.academic_year di SEMUA students (batch UPDATE)
  ├─ Update _data.academic_year di SEMUA class_rooms
  └─ Update _data.academic_year di SEMUA student_grades, student_invoices, dll


FASE 4: GENERATE SPP TAHUN BARU
──────────────────────────────────────────────────────────────────
AcademicYearActivatedHandler.Handle(event)
  │
  └─ Generate student_invoices untuk tahun ajaran baru:
       Untuk setiap siswa aktif: generate invoice per bulan × 12


FASE 5: ARCHIVING DATA HISTORIS
──────────────────────────────────────────────────────────────────
Tahun ajaran lama otomatis di-closed:
  └─ Semua data dengan academic_year_id = old_year → berstatus archive

Data tetap TIDAK DIHAPUS — diakses via "Riwayat Akademik" per siswa
Rapor yang sudah published tetap tersedia via parent portal
```

**Risiko:**
- Kenaikan kelas massal untuk sekolah besar (1000+ siswa) bisa timeout
  → Mitigasi: proses per kelas (batch 30-40 siswa), bukan semua sekaligus
- AcademicYearActivated memicu cascade SyncEngine yang besar
  → Mitigasi: SyncEngine batched (100 row per batch), rate limited
- Jika kelas tujuan sudah penuh saat kenaikan massal
  → Mitigasi: validasi kapasitas sebelum eksekusi; admin harus bagi siswa ke kelas lain
- Data siswa yang graduated masih bisa diakses oleh orang tua
  → Mitigasi: parent portal masih bisa akses rapor historis (read-only)

---

## 7. Tutup Buku Koperasi Tahunan (SHU)

```
PRE-REQUISITE:
  ✓ Accounting period 12 bulan sudah closed
  ✓ Audit laporan keuangan selesai (internal/eksternal)
  ✓ RAT (Rapat Anggota Tahunan) telah dilaksanakan


FASE 1: KALKULASI SHU
──────────────────────────────────────────────────────────────────
Admin Koperasi trigger hitung SHU:

1. Hitung total SHU tahun buku:
   SHU = Total Pendapatan - Total Beban (dari COA/jurnal)

2. Alokasi cadangan (minimum 25% — mandatory per UU Koperasi):
   Dana Cadangan = SHU × min 25%
   SHU yang dapat dibagi = SHU × max 75%

3. Hitung basis distribusi per anggota:
   Jasa Modal  = simpanan anggota / total simpanan (× alokasi jasa modal)
   Jasa Usaha  = transaksi anggota / total transaksi (× alokasi jasa usaha)
   Total SHU per anggota = Jasa Modal + Jasa Usaha

4. Review + approval pengurus koperasi


FASE 2: DISTRIBUSI SHU
──────────────────────────────────────────────────────────────────
Admin execute distribusi (setelah disetujui di RAT):
  │
  ▼
[UNTUK SETIAP NASABAH] (batch processing):

  [DATABASE TRANSACTION per nasabah]:
    ├─ CREATE shu_anggota {nasabah_id, amount_jasa_modal, amount_jasa_usaha}
    ├─ INSERT transaksi (setoran SHU ke rekening tabungan)
    ├─ UPDATE rekening.balance
    └─ INSERT jurnal entry
  COMMIT

  ▼
eventBus.Publish("events.koperasi.shu_distributed", SHUDistributedEvent)
  → Notify setiap nasabah: "SHU Anda tahun 2025: Rp X telah dikreditkan ke rekening"


FASE 3: PENUTUPAN TAHUN BUKU
──────────────────────────────────────────────────────────────────
1. Generate laporan keuangan tahunan (neraca, laba rugi)
2. Generate laporan untuk Dinas Koperasi (deadline 6 bulan setelah tutup buku)
3. Generate laporan OJK jika koperasi terdaftar LKM
4. Close accounting_period untuk tahun buku yang berakhir
5. Buka accounting_period untuk tahun buku baru
```

---

## Ringkasan Risiko Konsistensi Cross-Domain

| Skenario | Risiko | Mitigasi |
|----------|--------|---------|
| Siswa di-enroll tetapi Koperasi consumer down | Nasabah tidak ter-buat | NATS durable; consumer retry saat online kembali |
| Payment callback diterima dua kali | Double-update invoice | Idempotency check via external_id |
| Kenaikan kelas massal + SyncEngine overwhelmed | Vernon _data stale > 30 detik | Priority queue; critical fields (invoices) di-sync duluan |
| Rapor digenerate saat absensi belum complete | Rapor dengan data tidak akurat | Pre-generate validation; reject jika ada hari sekolah tanpa absensi |
| SHU distribusi gagal di tengah batch (500 dari 1000 nasabah) | Sebagian nasabah sudah terima SHU, sebagian belum | Idempotency per `shu_periode_id + nasabah_id`; resume dari failure point |
| Payroll deduction diproses dua kali bulan yang sama | Guru dipotong gaji dua kali | Idempotency key `batch_id + nasabah_id + period_month + period_year` |
| Academic year activation saat SyncEngine busy | _data tidak terupdate untuk tahun baru | Monitor `_sync_status = 'pending'`; alert jika > threshold |
| Parent portal menampilkan saldo stale koperasi | Orang tua melihat saldo lama | Cache TTL pendek (30 detik) untuk data keuangan real-time |
