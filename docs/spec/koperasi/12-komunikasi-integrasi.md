# 12 - Komunikasi, Dashboard, dan Integrasi API

Dokumen ini mendeskripsikan lapisan komunikasi dan integrasi Modul Koperasi SekolahPro: (1) Sistem Notifikasi multi-channel dengan WhatsApp sebagai primary (ADR-K022), (2) Dashboard & Self-Service Portal berbasis role (ADR-K023), dan (3) Integrasi API dengan sistem eksternal (ADR-K024). Ketiga modul ini merupakan "antarmuka keluar" koperasi — menghubungkan sistem inti dengan pengguna, orang tua, pemerintah, dan mitra.

---

## ADR References

| ADR | Judul | Status |
|-----|-------|--------|
| ADR-K022 | Notifikasi & Komunikasi | Accepted |
| ADR-K023 | Dashboard & Self-Service Portal | Accepted |
| ADR-K024 | Integrasi & API Gateway | Accepted |

---

## Domain Entities

### Notifikasi (K022)

| Entity | Deskripsi |
|--------|-----------|
| `notifikasi_preference` | Preferensi channel per user — primary/secondary channel, nomor WA, email, opt-out events, quiet hours. |
| `notifikasi_template` | Template pesan per event per channel per coop_type — body_template, variabel, status approval Meta (untuk WA). |
| `notifikasi_log` | Log setiap pengiriman notifikasi — status (queued/sending/sent/delivered/read/failed), retry count, external ID dari provider. |
| `in_app_notification` | Notifikasi di dalam dashboard — title, body, is_read, action_url. Selalu dibuat untuk semua event. |

### Dashboard & Portal (K023)

| Entity | Deskripsi |
|--------|-----------|
| `dashboard_widget` | Katalog widget sistem — code unik, data source, chart type, refresh interval, target roles. System-defined, tenant tidak bisa membuat widget baru. |
| `dashboard_layout` | Konfigurasi widget per role per tenant — posisi, ukuran, visibilitas. |
| `parent_child_config` | Link orang tua–anak beserta savings goal dan konfigurasi notifikasi. |
| `portal_session` | Tracking sesi portal nasabah/orang tua — device type, halaman yang dikunjungi, aksi yang dilakukan. |

### Integrasi API (K024)

| Entity | Deskripsi |
|--------|-----------|
| `integrasi_config` | Konfigurasi per integrasi per tenant — provider, config JSONB, credentials terenkripsi, health status. |
| `webhook_subscription` | Langganan sistem eksternal ke event koperasi — events yang disubscribe, URL tujuan, HMAC secret. |
| `api_key` | API key untuk integrasi eksternal — hash saja (key asli tidak disimpan), scope, IP whitelist, rate limit. |
| `integrasi_log` | Log setiap API call — direction, method, endpoint, status, duration, error. Dirotasi setelah 90 hari. |
| `integrasi_dead_letter` | Event/pesan yang gagal dikirim setelah semua retry — admin dapat replay atau discard. |

---

## Business Rules

### Sistem Notifikasi

1. **WhatsApp sebagai Primary Channel.** WhatsApp adalah channel komunikasi utama di Indonesia dengan open rate > 90%. Default preference nasabah adalah WhatsApp, dengan SMS sebagai fallback.

2. **Urutan Pengiriman Default.** IN_APP (selalu) → WhatsApp → SMS (fallback) → Email (untuk dokumen/attachment) → Push (jika app terinstall).

3. **IN_APP Tidak Dapat Di-Opt-Out.** Notifikasi di dalam dashboard selalu dibuat untuk semua event, tidak bisa dinonaktifkan. Bersifat silent dan non-intrusive (hanya bell notification).

4. **Priority dan Opt-Out.** CRITICAL dan HIGH (transaction): tidak dapat di-opt-out. HIGH (reminder) dan penting: tidak dapat di-opt-out. MEDIUM dan LOW: dapat di-opt-out oleh nasabah.

5. **Quiet Hours.** Notifikasi non-critical di-delay sampai quiet hours selesai. CRITICAL selalu dikirim terlepas dari quiet hours.

6. **Fallback Channel dengan Exponential Backoff.** Jika primary channel gagal: retry 3x dengan backoff 1 menit, 5 menit, 15 menit. Setelah 3x gagal → switch ke secondary channel. Jika semua channel gagal → log PERMANENTLY_FAILED + alert admin. Untuk CRITICAL: semua channel dikirim paralel (bukan sequential).

7. **WhatsApp Template Messages.** Notifikasi di luar 24-hour window menggunakan template message yang harus di-approve Meta. Proses approval 1-3 hari kerja. Status tracking di `wa_template_status`. Template default disiapkan untuk semua event saat tenant onboarding.

8. **Template Berbeda per coop_type.** Template `general` menggunakan "Halo" / "Terima kasih" / "Pinjaman". Template `islamic` menggunakan "Assalamu'alaikum" / "Jazakumullahu khairan" / "Pembiayaan". Template `both` berlaku jika tidak ada template spesifik.

9. **Provider Configurable per Tenant.** WhatsApp provider dapat dipilih: Meta (Official), Wablas, atau Fonnte. Payment gateway: Midtrans atau Xendit. Provider berbeda antar tenant diperbolehkan.

10. **Batch Notification Async.** Notifikasi massal (distribusi SHU, perubahan rate) di-queue dan diproses asynchronous. Rate limit: WA max 80 msg/detik, SMS max 30 msg/detik. Admin dapat memantau progress dan membatalkan batch yang masih QUEUED.

### Dashboard & Portal

11. **Dashboard Per Role.** Setiap role mendapat dashboard yang berbeda dengan widget yang relevan. Enam role utama: Admin/Manager (Executive Dashboard), Supervisor (Branch Operations), Teller (Transaction Workspace), Nasabah (Self-Service Portal), Orang Tua (Parent Portal), Kepala Sekolah (Oversight Dashboard).

12. **Widget-Based Architecture.** Setiap widget adalah unit independen dengan data source, refresh interval, dan fault isolation sendiri. Widget yang gagal fetch data menampilkan error state tanpa mempengaruhi widget lain.

13. **Widget Catalog System-Defined.** Tenant hanya dapat enable/disable widget, tidak dapat membuat widget baru. Konsistensi widget dijaga oleh tim development.

14. **Strict Row-Level Security.** Nasabah hanya melihat data milik sendiri. Orang tua hanya melihat data anak yang terhubung (via `controlled_by_id`). Teller/Supervisor dibatasi oleh branch_id. Kepala Sekolah hanya mendapat aggregate data tanpa PII.

15. **Self-Service Portal Capabilities.** Nasabah dapat VIEW data (saldo, riwayat, status pinjaman); DOWNLOAD statement PDF dan CSV; SUBMIT aplikasi (rekening baru, pinjaman baru, profile update request — semua memerlukan approval). Nasabah TIDAK DAPAT langsung edit saldo, transaksi, atau data pribadi.

16. **Profile Update via Request.** Edit data pribadi bukan direct edit — nasabah submit request yang harus diapprove oleh Teller/Supervisor. Ini menjaga integritas data.

17. **Parent Portal — Set Spending Limit Real-Time.** Perubahan `spending_control` oleh orang tua berlaku real-time. POS (K019) dan Kartu Belanja (K021) selalu query konfigurasi terbaru.

18. **Export Async untuk Data Besar.** Export > 1000 baris diproses asynchronous — user mendapat notifikasi (K022) saat file siap download. File memiliki expiry 7 hari, kemudian auto-delete. Semua export di-log ke audit trail.

19. **Statement PDF dengan Header Koperasi.** Statement menggunakan header yang dikonfigurasi per tenant (logo, nama, alamat, kontak). Format profesional siap cetak.

20. **Widget Pre-Computed Metrics.** Widget executive dashboard (NPL ratio, CAR, dll) menggunakan materialized view atau pre-computed metrics yang diupdate oleh background job, bukan real-time query saat request.

### Integrasi API

21. **API Design Standards.** URL pattern: `/api/v1/{resource}/{id}/{sub-resource}`. Cursor-based pagination (bukan offset) — performa konsisten di dataset besar. Response envelope konsisten: `data` + `meta` (request_id, timestamp). Error response dengan `errors[]` berisi code, message, field, detail.

22. **Versioning Policy.** Minor changes (tambah optional field): sama versi. Breaking changes: versi baru. Versi lama di-support minimal 6 bulan (internal) atau 12 bulan (public API) setelah versi baru.

23. **Tenant Context Wajib.** Setiap request wajib menyertakan tenant context via `X-Tenant-ID` header atau subdomain. Request tanpa tenant context ditolak dengan `400 Bad Request`.

24. **Authentication Tiga Metode.** JWT untuk dashboard/portal internal. API Key untuk integrasi eksternal (server-to-server). OAuth2 Authorization Code untuk self-service portal (nasabah/orang tua).

25. **API Key Hashing.** Key asli hanya ditampilkan SEKALI saat create. Server menyimpan hash saja (bcrypt/argon2). Tidak bisa di-retrieve ulang — harus buat key baru jika hilang.

26. **Webhook HMAC Signature.** Payload webhook ditandatangani dengan HMAC-SHA256. Subscriber WAJIB memverifikasi signature sebelum memproses. Header: `X-Webhook-Signature: sha256=<hash>` dan `X-Webhook-Timestamp`.

27. **Circuit Breaker per Integrasi.** CLOSED → (5 failures/menit) → OPEN (fast-fail 30 detik) → HALF-OPEN (1 test) → CLOSED/OPEN. Mencegah cascade failure saat external system down.

28. **Retry Exponential Backoff.** Attempt 1: immediate → 1 menit → 5 menit → 15 menit → 1 jam. Setelah 5x gagal: Dead Letter Queue + alert admin.

29. **Idempotency.** Callback payment yang sama (by external_payment_id) hanya diproses sekali. Upload inbound data dengan checksum yang sama tidak diproses ulang. Mencegah double-processing.

30. **IP Whitelist untuk Inbound.** Inbound API dari sistem eksternal (payroll, enrollment) memerlukan authentication API Key + IP whitelist. Double security untuk jalur yang menerima data dari luar.

31. **Sinkronisasi Sekolah → Koperasi Event-Driven.** Management Sekolah adalah source of truth untuk data person. Koperasi sebagai consumer. Mode enrollment: `auto` (buat application otomatis) atau `manual` (notifikasi saja). Bahkan mode `auto` tetap melalui approval flow (tidak langsung aktif).

32. **Dead Letter Queue Management.** Admin dapat melihat semua event yang gagal dikirim. Dapat replay manual setelah fix, atau discard. Jika subscription memiliki 10 consecutive failures: auto-disabled, admin harus re-enable manual.

33. **PII Masking di Log.** Data sensitif (nomor identitas, nomor HP, saldo) di-mask di semua integration logs. Credentials dan API key tidak pernah masuk log. Request body di-sanitize sebelum disimpan.

---

## Key Decisions & Rationale

### D1. WhatsApp-First, Bukan Email-First (K022)

**Keputusan:** WhatsApp sebagai primary notification channel.

**Alasan:** Penetrasi email di target user (orang tua siswa, guru sekolah di Indonesia) jauh lebih rendah dari WhatsApp. Open rate WhatsApp > 90% vs email < 30%. Ini mencerminkan realitas penggunaan komunikasi di Indonesia.

### D2. Tidak Build WhatsApp Gateway Sendiri (K022)

**Keputusan:** Menggunakan WhatsApp Business API resmi (Meta/Wablas/Fonnte), bukan unofficial web scraping.

**Alasan:** Unofficial API melanggar ToS, berisiko ban nomor, dan tidak reliable untuk template messages yang diperlukan notifikasi transaksional.

### D3. Widget-Based Configurable Dashboard, Bukan Hardcode per Role (K023)

**Keputusan:** Dashboard widget-based yang dapat dikonfigurasi per tenant, bukan layout yang di-hardcode per role.

**Alasan:** Setiap koperasi memiliki prioritas informasi yang berbeda. Koperasi kecil tidak butuh branch comparison. Widget-based memberikan fleksibilitas tanpa menambah kompleksitas development yang signifikan.

### D4. REST Bukan GraphQL (K024)

**Keputusan:** REST API dengan URL versioning.

**Alasan:** REST lebih familiar untuk developer target (koperasi/sekolah di Indonesia). GraphQL menambah complexity (schema management, N+1 protection, caching) yang belum justified untuk use case saat ini.

### D5. Strategy Pattern untuk Provider Integrasi (K024)

**Keputusan:** Semua provider diimplementasikan via interface/strategy pattern — Midtrans/Xendit, Meta/Wablas/Fonnte, Jurnal.id/Accurate.

**Alasan:** Mencegah vendor lock-in. Jika satu provider mengubah kebijakan atau harga, switch provider tidak memerlukan perubahan business logic. Beberapa tenant mungkin sudah memiliki kontrak dengan provider tertentu.

---

## Notification Events Komprehensif

### Event Nasabah

| Event | Channel | Priority | Opt-out |
|-------|---------|----------|---------|
| Transaksi receipt (setoran/penarikan/angsuran) | WA, IN_APP | HIGH | Tidak |
| Pinjaman disetujui/ditolak | WA, IN_APP | HIGH | Tidak |
| Pinjaman dicairkan | WA, IN_APP | HIGH | Tidak |
| Angsuran jatuh tempo (H-7, H-3) | WA, IN_APP | MEDIUM | Ya |
| Angsuran jatuh tempo (H-1) | WA, IN_APP | HIGH | Tidak |
| Angsuran overdue | WA, SMS, IN_APP | CRITICAL | Tidak |
| Deposito jatuh tempo (H-7) | WA, IN_APP | MEDIUM | Ya |
| Distribusi SHU | WA, EMAIL, IN_APP | HIGH | Tidak |
| Rekening dibekukan/dicairkan | WA, IN_APP | CRITICAL | Tidak |
| Password/PIN diubah | WA, IN_APP | CRITICAL | Tidak |

### Event Orang Tua

| Event | Channel | Priority | Opt-out |
|-------|---------|----------|---------|
| Ringkasan belanja harian anak | WA | MEDIUM | Ya |
| Notifikasi per transaksi anak | WA, IN_APP | LOW | Ya |
| Top-up berhasil | WA, IN_APP | MEDIUM | Tidak |
| Limit belanja hampir tercapai | WA, IN_APP | HIGH | Tidak |
| Kartu anak dibekukan/hilang | WA, IN_APP | CRITICAL | Tidak |
| Konfigurasi e-wallet diubah | WA, IN_APP | HIGH | Tidak |

### Event Staff Operasional

| Event | Channel | Priority | Opt-out |
|-------|---------|----------|---------|
| Queue approval baru | IN_APP, PUSH | MEDIUM | Ya |
| Posisi kas di bawah minimum | WA, IN_APP | CRITICAL | Tidak |
| NPL status berubah | WA, IN_APP | HIGH | Tidak |

### Event Management

| Event | Channel | Priority | Opt-out |
|-------|---------|----------|---------|
| Laporan regulasi jatuh tempo (H-30) | WA, EMAIL, IN_APP | HIGH | Tidak |
| Laporan regulasi jatuh tempo (H-7) | WA, EMAIL, IN_APP | CRITICAL | Tidak |
| Transaksi besar di atas threshold | WA, IN_APP | HIGH | Tidak |

---

## API Contracts Overview

### Endpoints Utama

```
GET    /api/v1/nasabah                       List nasabah (dengan filter)
GET    /api/v1/nasabah/{id}                  Detail nasabah
GET    /api/v1/nasabah/{id}/rekening         List rekening nasabah
GET    /api/v1/nasabah/{id}/transaksi        Riwayat transaksi
POST   /api/v1/transaksi                     Buat transaksi baru
GET    /api/v1/pinjaman/{id}/angsuran        Jadwal angsuran
GET    /api/v1/dashboard/executive           Data dashboard eksekutif
GET    /api/v1/portal/saldo                  Saldo nasabah (self-service)
GET    /api/v1/portal/statement              Statement PDF (self-service)
POST   /api/v1/portal/rekening/apply         Ajukan rekening baru
POST   /api/v1/portal/pinjaman/apply         Ajukan pinjaman baru
```

### Response Format

```json
// Success — single object
{
  "data": { "id": "018f...", "field": "value" },
  "meta": { "request_id": "req_018f...", "timestamp": "2026-04-15T10:30:00Z" }
}

// Success — list
{
  "data": [ {...}, {...} ],
  "meta": {
    "cursor": "eyJpZCI6IjAxOGYuLi4ifQ==",
    "has_more": true,
    "per_page": 25,
    "request_id": "req_018f...",
    "timestamp": "2026-04-15T10:30:00Z"
  }
}

// Error
{
  "errors": [
    {
      "code": "INSUFFICIENT_BALANCE",
      "message": "Saldo tidak mencukupi",
      "field": "amount",
      "detail": "Available: Rp 50.000, Required: Rp 100.000"
    }
  ],
  "meta": { "request_id": "req_018f...", "timestamp": "2026-04-15T10:30:00Z" }
}
```

### Webhook Events

| Event | Deskripsi |
|-------|-----------|
| `nasabah.approved` | Nasabah baru disetujui dan aktif |
| `rekening.opened` | Rekening baru dibuka |
| `transaksi.created` | Transaksi baru (setoran/penarikan/angsuran) |
| `pinjaman.approved` | Pinjaman disetujui |
| `pinjaman.overdue` | Angsuran melewati jatuh tempo |
| `shu.distributed` | SHU didistribusikan ke anggota |

---

## Dashboard Widgets per Role

### Executive Dashboard (Admin/Manager)

- Total aset koperasi (real-time)
- Total simpanan seluruh nasabah
- Total pinjaman outstanding
- NPL ratio, CAR, liquidity ratio (traffic light)
- Member growth chart
- Pending approvals count
- Cash position per branch (dari K014)
- SHU projection (dari K016)
- Regulatory report status
- Top 10 debtors
- Branch performance comparison

### Parent Portal

- Saldo rekening anak
- Belanja hari ini / minggu / bulan
- Tombol top-up
- Set/update spending limits (daily, per-transaction, category)
- Spending analytics (pie chart kategori, trend mingguan)
- Savings goal progress

---

## Integration Points

```
K022 (Notifikasi)
│
├── K011 (Transaksi)      — Trigger receipt notification per transaksi
├── K008 (Angsuran)       — Trigger reminder H-7, H-3, H-1, overdue
├── K016 (SHU)            — Trigger distribusi SHU notification
├── K021 (Uang Saku)      — Trigger spending alerts ke orang tua
├── K031 (Health)         — KpiAlertTriggeredEvent → notify management
└── K024 (API)            — WhatsApp Business API provider untuk pengiriman

K023 (Dashboard/Portal)
│
├── K021 (Uang Saku)      — Parent portal set spending_control
├── K002 (Rekening)       — View saldo dan riwayat
├── K007/K008 (Pinjaman)  — View loan status dan angsuran
├── K016 (SHU)            — View SHU history dan download sertifikat
├── K022 (Notifikasi)     — Download link file besar via notifikasi
└── K031 (Health)         — Executive dashboard widgets dari KPI measurements

K024 (API)
│
├── Management Sekolah    — Sinkronisasi data person (siswa, guru) event-driven
├── Payment Gateway       — Virtual Account top-up, QRIS (Midtrans/Xendit)
├── OJK/Dinas Koperasi    — Regulatory reporting (e-filing atau file download)
├── Accounting Software   — Ekspor jurnal (Jurnal.id, Accurate)
└── K022 (Notifikasi)     — WhatsApp Business API (Meta/Wablas/Fonnte)
```

---

## RBAC Summary

### Notifikasi

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View own notifications | v | v | v | v |
| View nasabah notification log | - | v | v | v |
| Manage templates | - | - | v | v |
| Configure WA integration | - | - | - | v |
| Send batch notification | - | - | v | v |

### Dashboard & Portal

| Permission | Nasabah | Orang Tua | Supervisor | Manager | Admin | Kepsek |
|------------|---------|-----------|------------|---------|-------|--------|
| View own portal | v | - | - | - | - | - |
| View parent portal | - | v | - | - | - | - |
| View executive dashboard | - | - | - | v | v | - |
| View oversight dashboard | - | - | - | - | - | v |
| Download own statement | v | - | - | - | - | - |
| Export data CSV | - | - | v | v | v | - |
| Configure widget layout | - | - | - | v | v | - |

### Integrasi API

| Permission | Supervisor | Manager | Admin |
|------------|------------|---------|-------|
| View integrasi log | v | v | v |
| View dead letter queue | - | v | v |
| Retry/discard dead letter | - | v | v |
| Trigger manual sync | - | v | v |
| Configure integrasi | - | - | v |
| Manage API keys | - | - | v |

---

## Dual-Mode Komunikasi

| Aspek | Koperasi Umum | BMT (Islamic) |
|-------|---------------|---------------|
| **Notifikasi** — Salam WA | "Halo" / "Terima kasih" | "Assalamu'alaikum" / "Jazakumullahu khairan" |
| **Notifikasi** — Terminologi | "Pinjaman", "Bunga" | "Pembiayaan", "Bagi Hasil/Margin" |
| **Dashboard** — Judul | "Dashboard Koperasi" | "Dashboard BMT" |
| **Dashboard** — Label | "Anggota", "Pinjaman" | "Nasabah", "Pembiayaan" |
| **Dashboard** — Greeting | "Selamat datang, Anggota" | "Assalamu'alaikum, Nasabah" |
| **API** — Response labels | Bunga, Pinjaman | Bagi Hasil, Pembiayaan |
| **API** — Webhook event | `loan.approved` | `pembiayaan.approved` |
| **Accounting Export** | Pendapatan Bunga | Pendapatan Bagi Hasil |
