# 03 — Integrasi Eksternal

Dokumen ini menjelaskan semua integrasi dengan sistem di luar SekolahPro: DAPODIK, Payment Gateway, E-Learning, BI-FAST/Perbankan, Pelaporan OJK, Takaful/Asuransi, dan WhatsApp/SMS.

---

## 1. DAPODIK (Kemendikbud)

### Konteks

DAPODIK (Data Pokok Pendidikan) adalah basis data nasional Kemendikbud. Sekolah **wajib** melaporkan data siswa, guru, dan sarana ke DAPODIK setiap semester untuk keperluan:
- Verifikasi BOS (Bantuan Operasional Sekolah)
- Data akreditasi BAN-S/BAN-SM
- Statistik pendidikan nasional
- Legitimasi NISN dan NUPTK

### Strategi Sinkronisasi

| Mode | Status | Keterangan |
|------|--------|------------|
| `manual_export` | **CURRENT (MVP)** | SekolahPro generate file, operator upload manual ke aplikasi Dapodik desktop |
| `api_sync` | **ROADMAP** | Integrasi langsung via API Dapodik (ditunda — API belum stabil) |

### Alur Manual Export (CURRENT)

```
1. PREPARE
   Operator buat sync batch → pilih:
   - Semester (ganjil/genap)
   - Entity types (siswa, guru, kelas, dll)
   System generate dapodik_sync_records dari data SekolahPro saat ini

2. VALIDATE
   Setiap record divalidasi terhadap aturan Dapodik:
   - NISN: 10 digit numerik wajib
   - NUPTK: 16 digit numerik untuk guru
   - NPSN: 8 digit numerik
   - Gender: hanya 'L' atau 'P'
   - Agama: kode 1-6 (Islam=1, Kristen=2, Katolik=3, Hindu=4, Buddha=5, Konghucu=6)
   - Tanggal lahir: format YYYY-MM-DD
   - disability_type: kode ABK Dapodik

   Result per record: valid | invalid (dengan detail error) | warning

3. REVIEW & FIX
   Operator review error list
   Fix data di SekolahPro (misal: tambahkan NISN yang belum ada)
   Re-validate records yang diperbaiki

4. APPROVE
   Kepala Sekolah approve batch → status: approved

5. EXPORT
   System generate file dengan format yang diterima Dapodik:
   - Format: JSON (preferred) atau CSV
   - Encoding: UTF-8
   - File tersimpan di Object Storage dengan URL berexpiry 7 hari

6. MANUAL UPLOAD
   Operator download file dari SekolahPro
   Upload ke aplikasi Dapodik desktop
   Konfirmasi hasil upload ke SekolahPro (manual entry)

7. TRACK
   Per record: pending → synced / error / retry / skipped
   Batch: completed / partial / error
```

### Field Mapping Utama (SekolahPro ↔ Dapodik)

| Entity | SP Field | Dapodik Field | Transform |
|--------|----------|---------------|-----------|
| Siswa | `students.nisn` | `no_peserta_didik` | Direct |
| Siswa | `students.full_name` | `nama` | Direct |
| Siswa | `students.gender` | `jenis_kelamin` | L→1, P→2 |
| Siswa | `students.religion` | `agama` | islam→1, kristen→2, ... |
| Siswa | `student_guardians.full_name` | `nama_orangtua` | Direct |
| Siswa | `student_guardians.education` | `pendidikan_orangtua` | Lookup code |
| Siswa | `student_health.disability_type` | `kebutuhan_khusus` | Kode ABK lookup |
| Guru | `teachers.nuptk` | `nuptk` | Direct |
| Guru | `teachers.nip` | `nip` | Direct |
| Guru | `teachers.employee_type` | `status_kepegawaian` | Lookup code |
| Sekolah | `company.npsn` | `npsn` | Direct |

### Limitasi Dapodik API

| Limitasi | Dampak | Penanganan |
|----------|--------|------------|
| Tidak ada public REST API yang stabil | Tidak bisa auto-sync real-time | Manual export sebagai solusi utama MVP |
| Format berubah setiap versi Dapodik | Template export bisa rusak saat Dapodik update | Gunakan `dapodik_field_mappings` tabel yang bisa diupdate tanpa deploy |
| Memerlukan login user ke aplikasi desktop | Tidak bisa fully automated | Operator sekolah tetap harus upload manual |
| NISN tidak selalu tersedia saat awal | Record invalid di Dapodik jika NISN kosong | Flag warning (bukan error) untuk siswa baru tanpa NISN |

---

## 2. Payment Gateway (Midtrans / Xendit / Duitku)

### Arsitektur Umum

```
SekolahPro (S051 Payment Gateway)
  │
  ├── Midtrans    — VA BNI/BRI/Mandiri, QRIS, GoPay, ShopeePay
  ├── Xendit      — VA multi-bank, QRIS, Retail (Alfamart/Indomaret), OVO
  └── Duitku      — Alternatif aggregator lokal
```

Provider dapat berbeda antar tenant (dikonfigurasi per `payment_channels`).

### Alur Pembayaran SPP (End-to-End)

```
SEKOLAH DOMAIN (S009)
─────────────────────
Sistem generate invoice SPP bulan ini per siswa
  │
  ▼
SEKOLAH DOMAIN (S051)
─────────────────────
1. Orang tua buka parent portal → pilih invoice yang mau dibayar

2. Request ke POST /api/v1/payment-transactions
   {
     "student_id": "...",
     "invoice_id": "...",
     "payment_type": "spp",
     "channel_type": "virtual_account",
     "idempotency_key": "{student_id}:{invoice_id}:{timestamp_second}"
   }

3. Payment service:
   a. Lookup/generate VA number (persistent per siswa per bank)
   b. Create payment_transaction (status = 'pending')
   c. Call provider API → return VA number + instructions

4. Response ke orang tua: VA number + nominal + expired_at

PROVIDER (Midtrans/Xendit)
──────────────────────────
5. Orang tua transfer via banking app/ATM

6. Provider konfirmasi pembayaran → kirim webhook callback:
   POST /api/v1/payment/callback/midtrans
   {
     "transaction_id": "...",
     "external_id": "ORDER-XXX",
     "status": "settlement",
     "amount": 500000,
     "payment_method": "bank_transfer",
     "bank": "BNI"
   }

SEKOLAH DOMAIN (S051)
─────────────────────
7. Callback handler:
   a. LOG ke payment_callbacks (immutable, raw payload)
   b. Validate signature (HMAC dari provider)
   c. Idempotency check via external_id — cegah double processing
   d. Update payment_transaction.status = 'paid'
   e. EMIT event: PaymentConfirmedEvent
      {payment_transaction_id, reference_type: 'student_invoice', reference_id: invoice_id}

SEKOLAH DOMAIN (S009)
─────────────────────
8. PaymentConfirmedHandler.Handle(event):
   UPDATE student_invoices SET
     paid_amount = paid_amount + payment.amount,
     status = CASE WHEN paid_amount >= total_amount THEN 'paid' ELSE 'partial' END
   WHERE id = invoice_id

9. EMIT event: InvoicePaidEvent → notifikasi ke orang tua (S044)
```

### Idempotency di Callback

```
Risiko: Provider kadang kirim callback yang sama 2x (retry dari sisi mereka)

Mitigasi:
  1. LOG semua callback ke payment_callbacks (immutable, no upsert)
  2. Sebelum proses: SELECT count(*) FROM payment_transactions
     WHERE external_id = $1 AND status IN ('paid', 'settled')
  3. Jika sudah ada record 'paid' → kembalikan 200 OK tanpa proses ulang
  4. Idempotency window: 24 jam (setelah itu, duplicate dianggap anomali)
```

### Virtual Account Persistent

```
student_virtual_accounts:
  student_id: 018f-student-uuid
  channel_id: channel-bni-uuid
  va_number: "8888012500010001"   ← Unik global, tidak berubah
  bank_code: "BNI"

Orang tua cukup simpan 1 nomor VA → digunakan untuk semua pembayaran SPP
mendatang (multi-invoice). Tidak perlu generate VA baru setiap bulan.
```

### Regulasi QRIS

```
MDR QRIS = 0.7% (per regulasi Bank Indonesia)
fee_bearer = 'school' selalu untuk QRIS
→ Sekolah menanggung biaya, bukan orang tua
→ fee_bearer tidak bisa di-set ke 'parent' untuk channel_type = 'qris'
```

### Rekonsiliasi Settlement

```
Provider settlement terjadi T+1 hingga T+7 setelah pembayaran

payment_settlements tracking:
  settlement_date: 2026-04-16
  gross_amount: 25,000,000
  fee_amount: 175,000
  net_amount: 24,825,000
  status: settled

Bendahara sekolah bisa download laporan settlement untuk rekonsiliasi
dengan rekening bank sekolah.
```

---

## 3. E-Learning (LMS)

### Strategi Integrasi

SekolahPro adalah **integration layer**, bukan LMS. File dan konten tetap di LMS — hanya metadata, statistik, dan nilai yang di-sync.

| Provider LMS | Status | Mode Sync |
|-------------|--------|-----------|
| Google Classroom | **CURRENT** | Polling + Webhook |
| Moodle (self-hosted) | **CURRENT** | REST API + Webhook |
| Microsoft Teams | **ROADMAP** | Graph API |
| Canvas | **ROADMAP** | REST API |
| Lainnya | Via `other` mode | Manual import |

### Alur Sinkronisasi LMS

```
OUTBOUND (SekolahPro → LMS)
─────────────────────────────
Guru assign tugas di SekolahPro (S022 assessment)
  │
  ▼
Jika LMS connected: system mirror assignment ke LMS
  POST ke LMS API: buat assignment di course yang sesuai
  Store external_id dari LMS di elearning_assignments.external_id

INBOUND (LMS → SekolahPro)
─────────────────────────────
Siswa submit tugas di LMS → Guru beri nilai di LMS
  │
  ▼
LMS kirim webhook event ke SekolahPro
  POST /api/v1/elearning/webhook/{integration_id}
  {
    "event": "submission.graded",
    "external_assignment_id": "LMS-ASSIGN-123",
    "student_external_id": "student-lms-456",
    "score": 88,
    "max_score": 100,
    "graded_at": "2026-04-15T10:00:00Z"
  }

Deduplikasi via external_id:
  UPSERT elearning_submissions ON CONFLICT (external_id) DO UPDATE

Jika integration_config.sync_grades_to_sp = true:
  Map LMS grade → student_grades (S011)
  dengan source_type = 'lms_assignment'
```

### Konfigurasi per Sekolah

```json
{
  "provider": "google_classroom",
  "sync_mode": "webhook",
  "sync_config": {
    "sync_grades_to_sp": true,
    "sync_attendance": false,
    "sync_materials": true,
    "grade_weight_percent": 30
  }
}
```

---

## 4. Perbankan / BI-FAST

### Status: ROADMAP (Belum Diimplementasi)

BI-FAST (Bank Indonesia Fast Payment) adalah infrastruktur pembayaran real-time antar bank di Indonesia. Relevansi untuk SekolahPro Koperasi:

| Use Case | Prioritas |
|----------|-----------|
| Transfer dana nasabah ke bank eksternal | Medium — nasabah penarikan via bank |
| Setoran dari rekening bank ke rekening koperasi | Medium — top-up via bank transfer |
| Kliring batch simpanan wajib dari bank | Low — jarang dibutuhkan skala kecil |

### Arsitektur Target (ROADMAP)

```
Koperasi SekolahPro
  │
  ▼
Bank Mitra (Koperasi memiliki rekening di bank mitra)
  │  BI-FAST API (via bank mitra sebagai sponsor)
  ▼
BI-FAST Infrastruktur Bank Indonesia
  │
  ▼
Bank Nasabah / Bank Tujuan
```

### Persyaratan Implementasi

- Koperasi harus memiliki rekening di bank mitra yang mendukung BI-FAST API
- API key dan sertifikat digital dari bank mitra
- Compliance KYC/AML untuk transaksi di atas threshold
- Tidak tersedia sebagai public API — harus bermitra dengan bank

---

## 5. Pelaporan OJK / Dinas Koperasi

### Kewajiban Regulasi

| Regulator | Laporan | Frekuensi | Format |
|-----------|---------|-----------|--------|
| **Dinas Koperasi Kabupaten/Kota** | Laporan tahunan (neraca + laporan usaha) | Tahunan (pasca RAT) | Form fisik / upload portal dinas |
| **Dinas Koperasi Provinsi** | Rekapitulasi data koperasi | Tahunan | Form online |
| **OJK** | Laporan kolektibilitas, CAR, BMPK | Bulanan (jika LKM) | e-reporting OJK |
| **Kemenkop-UKM** | Penilaian Koperasi Sehat | Tahunan | Online assessment |

### Pipeline Laporan Otomatis (K017)

```
PERSIAPAN (H-30 sebelum deadline)
─────────────────────────────────
System trigger: laporan_config.reminder_days_before
  │
  ▼
EMIT event: ReportDueReminderEvent
  → Notifikasi ke Manager + Admin (K022):
    "Laporan RAT jatuh tempo 30 hari lagi"

GENERASI LAPORAN
────────────────
Admin trigger generate laporan
  │
  ▼
System collect data dari:
  - Jurnal & COA (K015): neraca, laba rugi
  - SHU (K016): data bagi hasil anggota
  - Nasabah (K001): jumlah anggota aktif/keluar/baru
  - Pinjaman (K007): data NPL, kolektibilitas
  - Rekening (K002): total simpanan, tabungan, deposito

  ▼
Generate laporan per format yang dikonfigurasi:
  - Format Dinas Koperasi: sesuai template resmi (PDF/Excel)
  - Format OJK (jika LKM): sesuai format SIKP/e-reporting
  ⚠ Template form DAPAT BERUBAH setiap tahun — update manual diperlukan

REVIEW & APPROVAL
─────────────────
Laporan harus di-approve Kepala Koperasi sebelum disubmit
  │
  ▼
SUBMIT
──────
Opsi A: Download file → upload manual ke portal regulator
Opsi B (ROADMAP): Direct e-filing via API (OJK SIKP API — belum stabil)
```

### Laporan OJK untuk LKM (Jika Koperasi Terdaftar sebagai LKM)

| Metrik | Sumber Data | Frekuensi |
|--------|-------------|-----------|
| NPL Ratio | `pinjaman.collectibility` | Bulanan |
| CAR (Capital Adequacy Ratio) | `coa.modal_inti / ATMR` | Bulanan |
| BMPK | SUM pinjaman per nasabah / modal sendiri | Bulanan |
| Kolektibilitas (Kol 1-5) | DPD (Days Past Due) per pinjaman | Bulanan |
| Aset Produktif | SUM pinjaman + investasi | Bulanan |

---

## 6. Takaful / Asuransi Islam

### Status: ROADMAP (Target Pasar Pesantren)

Produk asuransi berbasis syariah (Takaful) relevan terutama untuk mode `coop_type = 'islamic'` dan `school_type = 'islamic'`.

### Use Cases yang Relevan

| Use Case | Domain | ADR |
|----------|--------|-----|
| Asuransi jiwa nasabah koperasi | Koperasi | Ekstensi K001 |
| Jaminan pembiayaan via Takaful | Koperasi | Ekstensi K010 |
| Asuransi kesehatan siswa (pesantren) | Sekolah | Ekstensi S005 |
| Infaq/zakat ke lembaga Takaful | Koperasi | K018 |

### Arsitektur Target (ROADMAP)

```
SekolahPro Koperasi
  │
  ├── Takaful Provider API (Takaful Indonesia / Takaful Keluarga / AIA Syariah)
  │     → Pendaftaran peserta
  │     → Pembayaran kontribusi (bukan premi) via tabungan nasabah
  │     → Klaim
  │
  └── OJK PPDPP (Pusat Data Pelaporan Perasuransian)
        → Pelaporan data peserta
```

---

## 7. WhatsApp / SMS Notification Providers

### Arsitektur Multi-Provider

```
Notification Service
  │  (Strategy Pattern — interface NotificationProvider)
  │
  ├── WhatsApp Business API (Primary Channel Indonesia)
  │     ├── Meta (Official API via Cloud API)
  │     ├── Wablas (Reseller Indonesia)
  │     └── Fonnte (Reseller Indonesia)
  │
  ├── SMS Gateway (Fallback)
  │     ├── Zenziva
  │     ├── Twilio
  │     └── Nexmo/Vonage
  │
  ├── Email (Dokumen Formal)
  │     ├── SMTP (self-hosted)
  │     └── SendGrid
  │
  └── Push Notification
        └── Firebase Cloud Messaging (FCM)
```

Provider dikonfigurasi per tenant di `integrasi_config`:
```json
{
  "whatsapp_provider": "fonnte",
  "wa_api_key_enc": "...",
  "sms_provider": "zenziva",
  "sms_api_key_enc": "...",
  "email_provider": "smtp",
  "smtp_host": "smtp.gmail.com"
}
```

### Alur Pengiriman Notifikasi

```
Domain event terjadi (contoh: siswa absen)
  │
  ▼
NotificationService.HandleEvent(event)
  │
  ├─ 1. Resolve recipients dari event payload
  │       (contoh: event.student_id → cari guardian_ids → cari user_ids)
  │
  ├─ 2. Load notification_preferences per user
  │       (primary channel, quiet hours, opt-out list)
  │
  ├─ 3. Render template per channel:
  │       "{{student_name}} tidak hadir hari ini, {{date}}"
  │
  ├─ 4. Generate idempotency_key:
  │       "attendance.absent:{attendance_id}:{channel}:{user_id}"
  │
  ├─ 5. Cek duplikat via idempotency_key
  │       → Jika sudah ada: skip
  │
  ├─ 6. Enqueue notification (async)
  │       INSERT notification_logs (status = 'pending')
  │
  └─ 7. Worker proses per channel:
         IN_APP: selalu dibuat, tidak bisa opt-out
         WhatsApp: call WA Business API
         SMS: call SMS Gateway (fallback jika WA gagal)
         Push: call FCM
```

### WhatsApp Template Approval

```
WhatsApp WAJIB menggunakan template yang di-approve Meta
untuk pesan di luar 24-hour window (pesan pertama ke user baru).

Template perlu di-approve per:
  - Bahasa (Indonesia / Arabic)
  - coop_type (general vs islamic — terminologi berbeda)
  - Event type (transaksi, angsuran, dll)

Lead time approval Meta: 1-3 hari kerja

Saat onboarding tenant baru:
  → System auto-submit semua template default untuk approval
  → Proses pembayaran: selama template belum approved,
    kirim via SMS sebagai fallback
```

### Rate Limiting

| Provider | Batas | Strategi |
|----------|-------|----------|
| WhatsApp (Meta Official) | 80 pesan/detik | Queue + batch dengan delay |
| WhatsApp (Wablas/Fonnte) | Sesuai paket | Rate limiter per tenant |
| SMS (Zenziva) | 30 pesan/detik | Queue + batch |
| FCM | 600 pesan/detik per project | Queue per project |

### Circuit Breaker

```go
// Per provider, per tenant
CLOSED → (5 failures dalam 1 menit) → OPEN (fast-fail 30 detik)
  → HALF-OPEN (1 test request) → kembali CLOSED atau OPEN
```

Jika semua channel gagal → log `PERMANENTLY_FAILED` + alert ke admin tenant.

### Biaya Estimasi

| Channel | Provider | Biaya/Pesan |
|---------|----------|-------------|
| Push (FCM) | Firebase | Gratis |
| In-App | Internal | Gratis |
| Email | SMTP / SendGrid | < Rp 1 |
| WhatsApp | Fonnte | Rp 100-500 |
| WhatsApp | Wablas | Rp 150-400 |
| SMS | Zenziva | Rp 350-500 |

Rekomendasi: gunakan Push + In-App untuk notifikasi volume tinggi (absensi, nilai). WhatsApp hanya untuk notifikasi kritikal (pembayaran, pelanggaran).
