# 15 — Keuangan & Integrasi Eksternal

Modul ini mencakup Payment Gateway (S051) untuk pembayaran online, serta lima integrasi eksternal dan infrastruktur: Transportasi (S052), E-Learning (S053), Reporting & Analytics (S054), Integrasi Dapodik (S055). Payment Gateway menggunakan CQRS Pattern karena sifat transaksional dengan konsistensi kuat; Reporting & Analytics juga menggunakan CQRS dengan materialized views karena 100% read operation; sisanya menggunakan Vernon Pattern.

---

## ADR References

| ADR | Judul | Pattern |
|-----|-------|---------|
| ADR-S051 | School Payment Gateway | CQRS |
| ADR-S052 | Transportation Management | Vernon |
| ADR-S053 | E-Learning Integration | Vernon |
| ADR-S054 | Reporting & Analytics Dashboard | CQRS + Materialized Views |
| ADR-S055 | Dapodik Integration | Vernon |

---

## Domain Entities

### Payment Gateway (S051)

**`payment_channels`** — Konfigurasi channel per sekolah:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `channel_type` | VARCHAR(20) | `virtual_account`, `qris`, `bank_transfer`, `retail`, `ewallet` |
| `provider_code` | VARCHAR(30) | Kode provider (Midtrans, Xendit, Duitku) |
| `config` | JSONB | Credentials dan konfigurasi per provider (encrypted at app layer) |
| `admin_fee_type` | VARCHAR(10) | `flat`, `percent`, `mixed` |
| `admin_fee_amount` | BIGINT | Fee flat dalam Rupiah |
| `admin_fee_percent` | NUMERIC(5,2) | Fee dalam persen |
| `fee_bearer` | VARCHAR(10) | `parent`, `school`, `split` |

**`student_virtual_accounts`** — VA persistent per siswa:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `student_id` | UUID | FK |
| `channel_id` | UUID | FK → payment_channels |
| `va_number` | VARCHAR(30) | Unik secara global (bukan per tenant) |
| `bank_code` | VARCHAR(10) | Kode bank |

**`payment_transactions`** — Transaksi pembayaran (append-only):

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `transaction_no` | VARCHAR(50) | Nomor transaksi (unik) |
| `student_id` | UUID | FK |
| `payment_type` | VARCHAR(20) | `spp`, `fee`, `canteen_topup`, `admission`, `other` |
| `reference_type` / `reference_id` | VARCHAR + UUID | Polymorphic: `student_invoice`, `canteen_wallet`, `applicant` |
| `amount` | BIGINT | Nominal (> 0) |
| `admin_fee` | BIGINT | Biaya admin |
| `total_amount` | BIGINT | `amount + admin_fee` |
| `status` | VARCHAR(20) | `pending` → `paid` → `settled` / `expired` / `failed` / `refunded` |
| `idempotency_key` | VARCHAR(100) | Unik — mencegah duplikasi dari retry/double-click |
| `external_id` | VARCHAR(100) | ID dari payment provider |
| `paid_at` | TIMESTAMPTZ | Waktu pembayaran dikonfirmasi |
| `settled_at` | TIMESTAMPTZ | Waktu dana masuk ke rekening sekolah (T+1 s/d T+7) |

**`payment_callbacks`** — Log webhook (immutable):

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `provider_code` | VARCHAR(30) | Provider yang mengirim callback |
| `raw_payload` | JSONB | Payload mentah dari provider |
| `signature` | TEXT | Signature untuk verifikasi |
| `process_status` | VARCHAR(20) | `received` → `processed` / `failed` / `ignored` |

**`payment_settlements`** — Batch settlement dari provider:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `settlement_date` | DATE | Tanggal settlement |
| `gross_amount` | BIGINT | Total bruto |
| `fee_amount` | BIGINT | Fee provider |
| `net_amount` | BIGINT | Dana yang diterima sekolah |
| `status` | VARCHAR(20) | `pending`, `settled`, `discrepancy` |

### Transportation Management (S052)

**`transport_vehicles`** — Kendaraan sekolah:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `plate_number` | VARCHAR(15) | Plat nomor (unik per company) |
| `vehicle_type` | VARCHAR(20) | `bus`, `minibus`, `van`, `pickup`, `sedan`, `other` |
| `capacity` | INT | Kapasitas (1–60) |
| `stnk_expiry`, `kir_expiry`, `insurance_expiry` | DATE | Monitoring dokumen kendaraan |
| `driver_name`, `driver_phone`, `driver_license_no` | VARCHAR | Data sopir (embedded, cukup untuk 1 sopir utama) |

**`transport_routes`** — Rute antar jemput:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `vehicle_id` | UUID | FK |
| `academic_year_id` | UUID | FK |
| `route_type` | VARCHAR(10) | `pickup`, `dropoff`, `both` |
| `waypoints` | JSONB | Ordered stops: `[{name, lat, lng, order}]` |
| `departure_time` | TIME | Jam keberangkatan |
| `monthly_fee` | BIGINT | Biaya per bulan per siswa |
| `passenger_count` | INT | Denormalisasi jumlah penumpang aktif |

**`transport_passengers`** — Assignment siswa ke rute:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `route_id` | UUID | FK |
| `student_id` | UUID | FK |
| `pickup_point` | VARCHAR(200) | Nama lokasi jemput |
| `pickup_latitude` / `pickup_longitude` | NUMERIC | Koordinat GPS (nullable) |
| `pickup_order` | INT | Urutan penjemputan (> 0) |
| `guardian_phone` | VARCHAR(20) | Kontak darurat sopir |

**`transport_logs`** — Log harian pick-up/drop-off:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `log_type` | VARCHAR(10) | `pickup` atau `dropoff` |
| `status` | VARCHAR(20) | `picked_up`, `dropped_off`, `absent`, `cancelled`, `parent_pickup` |
| `scheduled_time` | TIME | Waktu dijadwalkan |
| `actual_time` | TIME | Waktu aktual (nullable) |

### E-Learning Integration (S053)

**`elearning_integrations`** — Konfigurasi koneksi LMS:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `provider` | VARCHAR(30) | `google_classroom`, `moodle`, `ms_teams`, `canvas`, `other` |
| `sync_mode` | VARCHAR(20) | `webhook`, `polling`, `manual` |
| `api_key_enc`, `oauth_token_enc` | TEXT | Credentials (encrypted at app layer) |
| `last_sync_at` | TIMESTAMPTZ | Waktu sinkronisasi terakhir |

**`elearning_assignments`** — Tugas dari LMS:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `external_id` | VARCHAR(255) | ID dari LMS (unik per integration) |
| `assignment_type` | VARCHAR(20) | `tugas`, `kuis`, `ujian_online`, `diskusi`, `project` |
| `external_url` | TEXT | Link ke LMS |
| `submitted_count`, `graded_count` | INT | Stats dari LMS (di-cache) |
| `status` | VARCHAR(20) | `active`, `closed`, `archived`, `error` |

**`elearning_materials`** — Materi/resource links:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `material_type` | VARCHAR(20) | `document`, `video`, `audio`, `presentation`, `link`, `other` |
| `external_url` | TEXT | Link ke LMS atau cloud storage |

**`elearning_sync_logs`** — Log sinkronisasi:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `sync_type` | VARCHAR(20) | `webhook`, `polling`, `manual`, `full_sync` |
| `sync_direction` | VARCHAR(10) | `inbound` (LMS → SekolahPro) atau `outbound` |
| `entity_type` | VARCHAR(30) | `assignment`, `submission`, `material`, `grade`, `attendance`, `roster` |
| `status` | VARCHAR(20) | `pending` → `success` / `partial` / `error` / `retry` |
| `duration_ms` | INT | Durasi sinkronisasi dalam milidetik |

### Reporting & Analytics (S054)

**Menggunakan CQRS + Materialized Views** (tidak ada _rels/_data Vernon).

**Materialized Views:**

| View | Sumber | Refresh |
|------|--------|---------|
| `mv_school_dashboard` | students, teachers, classes, invoices | Setiap 5 menit |
| `mv_attendance_summary` | daily_attendances + class_placements | Setiap 15 menit |
| `mv_finance_summary` | student_invoices | Setiap 15 menit |
| `mv_grade_summary` | subject_grade_details + class_placements | Setiap 30 menit |

**`report_templates`** — Template laporan (pre-built + custom):

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `category` | VARCHAR(30) | `akademik`, `kehadiran`, `keuangan`, `dapodik`, `kepegawaian`, `sarana`, `custom` |
| `source_type` | VARCHAR(30) | Salah satu materialized view atau `custom_query` |
| `query_config`, `column_config`, `filter_config` | JSONB | Konfigurasi laporan |
| `is_system` | BOOLEAN | Laporan bawaan tidak bisa dihapus user |
| `allowed_roles` | JSONB | Role yang bisa mengakses |

**`dashboard_widgets`** — Konfigurasi widget per role:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `widget_type` | VARCHAR(20) | `counter`, `bar_chart`, `line_chart`, `pie_chart`, `area_chart`, `table`, `trend`, `alert` |
| `target_role` | VARCHAR(30) | Role yang melihat widget ini |
| `position_x`, `position_y`, `width`, `height` | INT | Grid layout (drag-and-drop) |

**`report_exports`** — Log export file:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `format` | VARCHAR(10) | `pdf`, `xlsx`, `csv` |
| `status` | VARCHAR(20) | `pending` → `generating` → `completed` / `error` / `expired` |
| `expires_at` | TIMESTAMPTZ | File dihapus setelah expired |

## EMIS Kemenag (untuk Madrasah)

Madrasah (MI, MTs, MA, MAK) berada di bawah Kementerian Agama — **bukan** Kemdikbud.
Mereka menggunakan **EMIS** (Education Management Information System Kemenag), bukan Dapodik.

**Perbedaan EMIS vs Dapodik:**
| Aspek | Dapodik (Kemdikbud) | EMIS (Kemenag) |
|-------|---------------------|-----------------|
| Digunakan oleh | SD, SMP, SMA, SMK | MI, MTs, MA, MAK |
| BOS | Dana BOS Kemdikbud | Dana BOS Kemenag (berbeda rekening) |
| Data guru | NUPTK via GTK | NUPTK via Simpatika Kemenag |
| Rapor | e-Rapor Kemdikbud | e-Rapor Kemenag (ARD) |

**Status Integrasi di SekolahPro:**
- ADR-S055 saat ini hanya membahas integrasi Dapodik (Kemdikbud)
- **EMIS integration adalah fitur roadmap** — belum ada di ADR
- Untuk madrasah yang menggunakan SekolahPro: data harus diinput manual ke EMIS
- SekolahPro menyediakan export format CSV yang disesuaikan dengan kebutuhan EMIS
  (format berbeda dari Dapodik — konfigurasi per sekolah)

> **Risiko untuk Segmen Pesantren**: Sebagian besar pesantren mengelola madrasah
> (MTs/MA) di bawah Kemenag. Gap EMIS ini adalah kebutuhan mendesak untuk segmen
> target utama SekolahPro.

---

### Dapodik Integration (S055)

### Verval PD dan Alokasi NISN

**Verval PD** (Verifikasi dan Validasi Peserta Didik) adalah proses wajib untuk
memastikan setiap siswa memiliki NISN (Nomor Induk Siswa Nasional) yang valid.

**Alur NISN untuk Siswa Baru:**
1. Operator input data siswa baru di SekolahPro (`students` tabel)
2. Cek NISN: apakah siswa sudah punya NISN dari sekolah sebelumnya?
   - **Ya**: input NISN existing → validasi via sync Dapodik (batch/manual)
   - **Tidak**: kasus umum untuk siswa SD kelas 1 atau siswa pindahan tanpa NISN
3. Untuk siswa tanpa NISN: operator upload data ke **Verval PD Kemdikbud** (portal terpisah)
4. Kemdikbud mengalokasikan NISN baru dalam 1-7 hari kerja
5. Operator update `students.nisn` di SekolahPro setelah NISN diterima
6. Sync Dapodik selanjutnya akan memvalidasi NISN

**Field terkait:**
- `students.nisn`: wajib untuk semua siswa SD ke atas
- `students.nisn_verified`: boolean, true setelah dikonfirmasi valid di Dapodik
- Siswa tanpa NISN yang terverifikasi tidak bisa di-include dalam laporan BOS

> **Catatan**: NISN dialokasikan oleh Kemdikbud via portal Verval PD
> (vervalpdun.data.kemdikbud.go.id). SekolahPro tidak dapat mengalokasikan NISN —
> hanya menyimpan dan memvalidasi NISN yang sudah ada.

**`dapodik_field_mappings`** — Pemetaan field SekolahPro ↔ Dapodik:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `entity_type` | VARCHAR(30) | `student`, `teacher`, `school_profile`, `class`, `subject`, `attendance`, `grade`, `infrastructure` |
| `sp_field` | VARCHAR(100) | Nama field di SekolahPro |
| `dapodik_field` | VARCHAR(100) | Nama field di Dapodik |
| `transform_type` | VARCHAR(20) | `direct`, `lookup`, `format`, `concat`, `split`, `custom` |
| `dapodik_lookup` | JSONB | Tabel referensi kode Dapodik (misal agama: 1=Islam, dll) |

**`dapodik_sync_batches`** — Batch sync per semester:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `academic_year_id` | UUID | FK |
| `semester` | INT | 1 atau 2 |
| `npsn` | VARCHAR(8) | NPSN sekolah (8 digit) |
| `sync_mode` | VARCHAR(20) | `manual_export` (default MVP) atau `api_sync` |
| `status` | VARCHAR(20) | `draft` → `validating` → `validated` → `approved` → `syncing` → `completed` / `partial` / `error` |
| `export_file_url` | TEXT | URL file export (untuk manual_export mode) |

**`dapodik_sync_records`** — Status per entity:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `entity_type` + `entity_id` | VARCHAR + UUID | Polymorphic reference |
| `nisn` | VARCHAR(10) | NISN 10 digit (untuk student) |
| `nuptk` | VARCHAR(16) | NUPTK 16 digit (untuk teacher) |
| `sp_data` | JSONB | Snapshot data SekolahPro saat sync |
| `dapodik_data` | JSONB | Data dari Dapodik (untuk perbandingan) |
| `diff_fields` | JSONB | Field yang berbeda antara SP dan Dapodik |
| `validation_status` | VARCHAR(20) | `pending`, `valid`, `invalid`, `warning` |
| `validation_errors` | JSONB | List error validasi Dapodik per record |
| `sync_status` | VARCHAR(20) | `pending` → `synced` / `error` / `retry` / `skipped` |

---

## Business Rules

### Payment Gateway

1. **Idempotency key wajib** pada setiap transaksi — format: `{student_id}:{invoice_id}:{timestamp_second}`. Duplicate request dengan key yang sama mengembalikan transaksi yang sudah ada.
2. **MDR QRIS (0.7%) ditanggung merchant (sekolah)** per regulasi Bank Indonesia — `fee_bearer` untuk QRIS harus selalu `school`.
3. **Virtual Account bersifat persistent** per siswa per bank — sekali di-generate, VA dapat digunakan untuk semua pembayaran mendatang.
4. **`payment_callbacks` adalah append-only dan immutable** — setiap webhook dari provider tercatat apa adanya untuk audit dan debugging.
5. **Timeout transaksi** — transaksi yang melewati `expired_at` otomatis di-expire via cron job.
6. **Settlement delay:** dana tidak real-time masuk ke rekening sekolah — T+1 hingga T+7 tergantung provider.
7. **Pembayaran invoice S009** di-update via side effect callback processing — payment domain tidak langsung menulis ke domain keuangan.

### Transportation

8. **Kapasitas kendaraan tidak boleh melebihi batas.** Penambahan penumpang dicek terhadap `vehicle.capacity`.
9. **`passenger_count` adalah denormalisasi** yang diperbarui setiap assignment ditambah atau dihapus.
10. **Monitoring dokumen kendaraan** (`stnk_expiry`, `kir_expiry`, `insurance_expiry`) — endpoint `/expiring` memberikan daftar kendaraan dengan dokumen yang akan habis dalam 30 hari.
11. **Log harian bersifat manual** (input sopir via app) — belum ada integrasi GPS tracker. Field koordinat tersedia untuk future integrasi IoT.
12. **Biaya transportasi bulanan** dapat di-generate sebagai invoice via S009.

### E-Learning Integration

13. **SekolahPro adalah integration layer, bukan LMS.** File/konten tetap di LMS — hanya link dan metadata yang di-sync.
14. **Deduplikasi via `external_id`.** Sync yang sama dua kali menghasilkan upsert, bukan duplikasi.
15. **Credentials LMS dienkripsi di application layer** — tidak disimpan plain text di database.
16. **Circuit breaker dan retry** diperlukan — API LMS bisa down; sync tidak boleh membuat cascade failure.

### Reporting & Analytics

17. **Materialized views menggunakan `REFRESH CONCURRENTLY`** — memerlukan unique index agar view dapat di-refresh tanpa locking read queries.
18. **Custom query di report builder harus melewati whitelist fields dan operators** — mencegah SQL injection.
19. **Role-based access:**
    - Kepsek: akses semua laporan.
    - Guru/Wali Kelas: hanya kelas sendiri.
    - Orang Tua: hanya data anak sendiri.
    - Bendahara: laporan keuangan penuh; tidak bisa akses akademik.
20. **Export file bersifat async** — file di-generate di background, user polling status via `/report-exports/{id}`.
21. **Data di materialized views bisa stale 5–30 menit** — design goal adalah performa dashboard sub-1-detik, bukan real-time.

### Dapodik Integration

22. **Mode utama MVP adalah `manual_export`** — SekolahPro generate file (JSON/CSV/XLSX) yang diupload manual oleh operator ke aplikasi Dapodik desktop.
23. **`api_sync` bersifat deferred** — tidak diimplementasikan di Phase 1 karena Dapodik tidak memiliki public REST API yang stabil.
24. **Pre-validation sebelum sync** — setiap record di-validasi terhadap aturan Dapodik (NISN 10 digit, NUPTK 16 digit, gender L/P, kode agama 1–6, dll) sebelum export.
25. **NISN selalu 10 digit, NUPTK selalu 16 digit, NPSN selalu 8 digit** — validasi format wajib.
26. **Batch sync memerlukan approval kepsek** sebelum export/sync dijalankan.
27. **Auto-retry max 3x** untuk records dengan `sync_status = 'error'` (hanya untuk mode `api_sync`).
28. **Data Dapodik adalah PII sensitif** — NISN, NUPTK, biodata lengkap memerlukan enkripsi dan access control ketat.

---

## Alur Sinkronisasi Dapodik

```
1. PREPARE: Operator buat batch → pilih semester + entity types
   System generate dapodik_sync_records dari data SekolahPro

2. VALIDATE: Setiap record di-validasi terhadap aturan Dapodik
   Result: valid / invalid (dengan detail error) / warning per record

3. REVIEW & FIX: Operator review errors → fix data di SekolahPro
   Re-validate records yang diperbaiki

4. APPROVE: Kepsek approve batch → status: approved

5. SYNC:
   (A) Manual Export: Generate file → operator upload ke Dapodik desktop
   (B) API Sync (future): Kirim langsung via API Dapodik

6. TRACK: Per record: pending → synced / error / retry / skipped
   Batch: syncing → completed / partial / error
```

## Alur Pembayaran Online

```
1. Orang tua pilih invoice yang dibayar
2. System lookup/generate VA number siswa
3. Create payment_transaction (status = 'pending')
4. Call payment provider API → return payment instructions
5. Orang tua bayar via banking app
6. Provider kirim callback webhook:
   a. Log ke payment_callbacks (immutable)
   b. Validate signature provider
   c. Match ke transaction via external_id
   d. Update status → 'paid'
   e. Side effect: update invoice S009 atau canteen_wallet S037
7. Provider settlement (T+1 s/d T+7):
   a. Batch settlement data masuk
   b. Update transaction status → 'settled'
   c. Update payment_settlement record
```

---

## Key Decisions & Rationale

1. **Payment Gateway menggunakan CQRS** (bukan Vernon) karena transaksi keuangan memerlukan strong consistency, bukan eventual consistency. Tidak ada `_rels/_data` denormalization.
2. **Reporting menggunakan Materialized Views** (bukan real-time aggregation) karena query GROUP BY + SUM pada jutaan baris terlalu lambat untuk dashboard (target < 1 detik). Stale data 5–30 menit acceptable.
3. **Dapodik mode utama adalah manual_export** karena Dapodik tidak expose public API yang stabil. Engineering time tidak diinvestasikan ke integrasi API yang rapuh.
4. **SekolahPro bukan LMS** — integrasi e-learning hanya sync metadata dan stats dari LMS yang sudah ada (Google Classroom, Moodle, dll) untuk view terpusat.
5. **VA bersifat persistent** — orang tua cukup simpan 1 nomor VA untuk semua pembayaran, mengurangi friction.
6. **Tabel transport_logs bukan real-time GPS** — tracking manual via app sopir lebih realistis untuk MVP dibandingkan infrastruktur IoT.

---

## Integration Points

### Payment Gateway (S051)

| Domain | Keterangan |
|--------|-----------|
| S009 Student Finance | Callback payment memperbarui invoice (paid_amount, status) |
| S037 Canteen Billing | Callback top-up memperbarui canteen_wallet.balance |
| S016 PPDB | Pembayaran pendaftaran via payment gateway |

### Transportation (S052)

| Domain | Keterangan |
|--------|-----------|
| S009 Student Finance | Biaya transportasi bulanan di-generate sebagai invoice |
| ADR-S001 Students | Data siswa untuk assignment ke rute |
| ADR-S003 Guardians | Kontak darurat orang tua di setiap penumpang |

### E-Learning (S053)

| Domain | Keterangan |
|--------|-----------|
| S011 Subject Grades | Assignment scores dari LMS bisa menjadi input nilai |
| S021 Timetable | Assignment/materi di-link ke jadwal pelajaran |
| ADR-012 Teachers | Mapping guru SekolahPro ke teacher di LMS |

### Reporting (S054)

| Domain | Sumber Data |
|--------|-----------|
| S008 Attendance | `mv_attendance_summary` |
| S009 Finance | `mv_finance_summary` |
| S011 Grades | `mv_grade_summary` |
| Semua domain | `mv_school_dashboard` (counter umum) |

### Dapodik (S055)

| Sistem | Keterangan |
|--------|-----------|
| **Kemendikbud Dapodik** | Export field-mapped data per semester; verifikasi BOS |
| **BAN-S/BAN-SM** | Data akreditasi (S048) dari profil sekolah |
| **BKN/SIASN** | Data kepegawaian guru PNS (deferred — API belum stabil) |

### Eksternal (Payment)

| Provider | Keterangan |
|----------|-----------|
| **Midtrans** | VA BNI/BRI/Mandiri, QRIS, GoPay |
| **Xendit** | VA multi-bank, QRIS, retail (Alfamart/Indomaret) |
| **Duitku** | Alternatif payment aggregator lokal |
