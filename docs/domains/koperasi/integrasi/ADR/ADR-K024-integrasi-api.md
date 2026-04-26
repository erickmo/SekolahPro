# ADR-K024: Integrasi & API (Integration & API Gateway)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Sistem koperasi sekolah tidak berdiri sendiri — ia berinteraksi dengan banyak sistem eksternal dan internal:

- **Management Sekolah** (sibling subproject) — sumber data siswa, guru, staf, dan kalender akademik
- **Payment Gateway** — untuk pembayaran via Virtual Account dan QRIS
- **Regulator** (OJK/Dinas Koperasi) — pelaporan berkala
- **Software Akuntansi** — ekspor jurnal ke sistem akuntansi eksternal
- **WhatsApp Business API** — notifikasi transaksional ([ADR-K022](./ADR-K022-notifikasi.md))
- **Sistem eksternal lainnya** — payroll, bank statement, dsb.

Sistem membutuhkan **API Gateway** yang terstandar, aman, dan scalable untuk menangani semua integrasi ini. Desain harus mengakomodasi:

- Multi-tenant: setiap tenant bisa memiliki konfigurasi integrasi berbeda
- Dual-mode: terminologi dan label menyesuaikan `coop_type` ([ADR-009](../core/ADR-009-dual-mode-institution-type.md))
- Resilience: integrasi eksternal bisa gagal — sistem harus tetap berjalan
- Security: data keuangan sensitif — semua komunikasi harus terenkripsi dan terautentikasi

## Decision

### 1. Integrasi dengan Management Sekolah (Sibling Subproject)

Management Sekolah dan Management Koperasi adalah dua subproject independen yang bisa digunakan bersamaan. Jika keduanya aktif, sinkronisasi data dilakukan secara **event-driven**.

```
Management Sekolah (Source of Truth: data person)
        │
        │  Event Bus / REST Webhook
        v
Management Koperasi (Consumer)
```

**Skenario sinkronisasi:**

| Event di Sekolah | Aksi di Koperasi | Mode |
|------------------|------------------|------|
| Siswa/santri enrolled | Auto-create nasabah application atau notifikasi untuk trigger manual | Configurable |
| Guru/staf onboarding | Sync data personal → nasabah application | Configurable |
| Data personal updated (nama, alamat, status) | Update data nasabah yang ter-link via `school_entity_id` | Auto |
| Siswa keluar/lulus/mutasi | Flag nasabah untuk review deactivation ([ADR-K001](./ADR-K001-nasabah.md) Section 7) | Auto |
| Kalender akademik (libur, cuti) | Update holiday calendar untuk penyesuaian jatuh tempo angsuran ([ADR-K008](./ADR-K008-angsuran-jadwal.md)) | Auto |
| Roster kelas/halaqah berubah | Update workflow koleksi simpanan wajib via wali kelas ([ADR-K004](./ADR-K004-simpanan-pokok-wajib.md)) | Auto |

**Aturan sinkronisasi:**
- **Arah utama**: Sekolah → Koperasi (sekolah adalah source of truth untuk data person)
- **Mode enrollment**: configurable per tenant — `auto` (otomatis buat application) atau `manual` (hanya notifikasi, operator buat application manual)
- Jika mode `auto`: application tetap melalui approval flow ([ADR-K001](./ADR-K001-nasabah.md) Section 1) — tidak langsung jadi nasabah aktif
- Link antara entitas sekolah dan nasabah via field `school_entity_id` di tabel nasabah
- Jika Management Sekolah tidak aktif (tenant hanya pakai koperasi), sinkronisasi dinonaktifkan — tidak ada error

**Konfigurasi per tenant:**
```
school_sync_config:
  enabled: true | false
  enrollment_mode: "auto" | "manual"
  sync_personal_data: true              # Auto-update nama, alamat, dll
  sync_exit_status: true                # Auto-flag saat siswa keluar
  sync_calendar: true                   # Sync kalender libur
  sync_roster: true                     # Sync roster kelas/halaqah
  webhook_url: "https://..."            # Jika via REST webhook
  event_bus_topic: "school.events"      # Jika via shared event bus
```

### 2. Payment Gateway Integration

Integrasi dengan payment gateway untuk menerima pembayaran dari luar sistem (top-up tabungan, pembayaran angsuran, pembelian di toko/kantin via QRIS).

**Channel pembayaran:**

| Channel | Use Case | Provider |
|---------|----------|----------|
| Virtual Account (VA) | Top-up tabungan via transfer bank (BCA, Mandiri, BNI, BRI) | Midtrans / Xendit |
| QRIS | Pembayaran di kantin/toko ([ADR-K019](./ADR-K019-toko-kantin.md)) | Midtrans / Xendit |

**Flow Virtual Account:**
```
Orang tua request VA
        │
        v
┌─────────────────────────┐
│ Sistem generate VA      │
│ via Payment Provider    │
│ VA: 8001-1234-5678-9012 │
│ Amount: Rp 500.000      │
│ Expiry: 24 jam          │
└────────┬────────────────┘
         │ Transfer via bank
         v
┌─────────────────────────┐
│ Bank proses transfer    │
│ Provider terima payment │
└────────┬────────────────┘
         │ Callback/webhook
         v
┌─────────────────────────┐
│ Sistem terima notifikasi│
│ Validasi signature      │
│ Create transaksi (K011) │
│ Update saldo tabungan   │
│ Kirim notifikasi (K022) │
└─────────────────────────┘
```

**Flow QRIS:**
```
Kasir tampilkan QRIS
        │
        v
┌─────────────────────────┐
│ Customer scan via app   │
│ bank/e-wallet           │
└────────┬────────────────┘
         │ Payment processed
         v
┌─────────────────────────┐
│ Provider callback       │
│ → Create transaksi POS  │
│ → Update stok           │
│ → Kirim receipt         │
└─────────────────────────┘
```

**Aturan payment gateway:**
- Provider **configurable per tenant** — Midtrans atau Xendit (atau keduanya untuk channel berbeda)
- Callback/webhook dari provider harus divalidasi menggunakan **signature verification** (HMAC)
- Setiap payment yang masuk otomatis membuat transaksi di [ADR-K011](./ADR-K011-transaksi.md)
- **Reconciliation harian**: auto-match VA payments dengan expected amounts — selisih di-flag untuk review manual
- **Settlement**: provider menyetor ke rekening bank koperasi (T+1 atau T+2 tergantung provider)
- Idempotent: callback yang sama diproses hanya sekali (deduplicated by `external_payment_id`)

**Konfigurasi per tenant:**
```
payment_gateway_config:
  provider: "midtrans" | "xendit"
  environment: "sandbox" | "production"
  server_key: "encrypted_..."
  client_key: "encrypted_..."
  va_enabled: true
  va_banks: ["bca", "mandiri", "bni", "bri"]
  qris_enabled: true
  callback_url: "https://{tenant}.api.sekolahpro.id/webhooks/payment"
  settlement_account: "1234567890"        # Rekening bank koperasi
  auto_reconcile: true
  reconcile_time: "02:00"                 # Jadwal reconciliation harian
```

### 3. OJK / Dinas Koperasi Reporting Integration

Koperasi wajib melaporkan data keuangan secara berkala ke regulator. Data laporan bersumber dari [ADR-K015](./ADR-K015-jurnal-coa.md) (akuntansi) dan ADR-K017 (laporan regulasi).

**Mekanisme pelaporan:**

```
Data keuangan (K015, K017)
        │
        v
┌─────────────────────────┐
│ Generate laporan dalam  │
│ format yang diminta      │
│ (XBRL, XML, CSV, XLSX)  │
└────────┬────────────────┘
         │
    ┌────┴────┐
    v         v
┌────────┐ ┌──────────────────────┐
│E-filing│ │ Download file untuk  │
│via API │ │ upload manual         │
└────────┘ └──────────────────────┘
```

**Aturan:**
- Jika regulator menyediakan **e-filing API** (XBRL, XML, atau provider-specific): sistem submit langsung via API
- Jika tidak ada API: sistem auto-generate file dalam format yang diminta, siap untuk **upload manual** oleh admin
- Format didukung: XBRL, XML, CSV, XLSX — configurable per jenis laporan
- Laporan di-generate dari data yang sudah tersedia di modul akuntansi — **tidak ada input manual tambahan**
- Jadwal pelaporan sesuai regulasi — sistem kirim reminder ke Manager sebelum deadline ([ADR-K022](./ADR-K022-notifikasi.md))

**Konfigurasi:**
```
regulatory_reporting_config:
  ojk_enabled: false                       # Aktifkan jika LKM terdaftar OJK
  dinas_koperasi_enabled: true             # Wajib untuk semua koperasi
  e_filing_mode: "auto" | "manual"         # Auto = submit via API, Manual = generate file
  e_filing_api_url: "https://..."          # Jika auto
  e_filing_credentials: "encrypted_..."    # Jika auto
  report_format: "xlsx"                    # Default format untuk manual upload
```

### 4. Accounting Software Export

Ekspor jurnal dan data akuntansi ke software akuntansi eksternal untuk kebutuhan audit atau konsolidasi.

**Software yang didukung:**

| Software | Metode Ekspor | Format |
|----------|--------------|--------|
| Jurnal.id | API (REST) | JSON |
| Accurate | File export | CSV |
| Custom | File export | CSV / XLSX / JSON |

**Aturan:**
- **COA mapping**: internal Chart of Account ([ADR-K015](./ADR-K015-jurnal-coa.md)) di-mapping ke COA di software tujuan — configurable per tenant
- Ekspor bisa **scheduled** (harian/mingguan/bulanan) atau **on-demand** (manual trigger)
- Setiap ekspor menghasilkan **batch** — semua jurnal dalam periode yang dipilih
- Jurnal yang sudah diekspor di-flag `exported_at` — mencegah duplikasi pada ekspor berikutnya
- Jika via API (Jurnal.id): auto-push, dengan retry jika gagal
- Jika via file: generate file, simpan di storage, admin download

**Konfigurasi:**
```
accounting_export_config:
  enabled: false
  target: "jurnal_id" | "accurate" | "custom"
  format: "json" | "csv" | "xlsx"
  schedule: "daily" | "weekly" | "monthly" | "manual"
  schedule_time: "23:00"
  api_url: "https://..."                   # Untuk Jurnal.id
  api_key: "encrypted_..."                 # Untuk Jurnal.id
  coa_mapping: {                           # Internal COA → External COA
    "1-1001": "10100",
    "1-1002": "10200"
  }
```

### 5. WhatsApp Business API (untuk K022 Notifikasi)

Detail integrasi WhatsApp untuk mendukung notifikasi transaksional di [ADR-K022](./ADR-K022-notifikasi.md).

**Provider yang didukung:**

| Provider | Tipe | Kelebihan |
|----------|------|-----------|
| Meta (Official) | Direct API | Paling reliable, rate limit tertinggi |
| Wablas | Third-party | Mudah setup, harga kompetitif |
| Fonnte | Third-party | Populer di Indonesia, integrasi mudah |

**Jenis pesan:**

| Tipe | Kapan Digunakan | Persetujuan Meta |
|------|-----------------|------------------|
| Template message | Di luar 24-hour window (mayoritas notifikasi kita) | Wajib pre-approved |
| Session message | Dalam 24-hour window setelah user reply | Tidak perlu approval |

**Aturan:**
- Provider **configurable per tenant** — bisa berbeda antara tenant satu dan lainnya
- **Template messages** wajib di-approve Meta sebelum bisa digunakan — proses approval 1-3 hari
- Status template dilacak di `wa_template_status` (pending, approved, rejected) — lihat [ADR-K022](./ADR-K022-notifikasi.md) Section 4
- **Session messages** digunakan untuk flow interaktif (misal: konfirmasi OTP, reply-based approval)
- **Delivery tracking**: sent → delivered → read — status di-update via webhook dari provider
- Jika pengiriman gagal 3x, fallback ke SMS ([ADR-K022](./ADR-K022-notifikasi.md) Section 9)
- Rate limiting sesuai tier WhatsApp Business API (1K, 10K, 100K messages/day)

### 6. API Design Principles

Semua API (internal dan eksternal) mengikuti prinsip desain yang konsisten:

**Struktur URL:**
```
/api/v1/{resource}
/api/v1/{resource}/{id}
/api/v1/{resource}/{id}/{sub-resource}
```

**Contoh:**
```
GET    /api/v1/nasabah                    # List nasabah
GET    /api/v1/nasabah/{id}               # Detail nasabah
POST   /api/v1/nasabah                    # Create nasabah
PUT    /api/v1/nasabah/{id}               # Update nasabah
GET    /api/v1/nasabah/{id}/rekening      # List rekening nasabah
POST   /api/v1/transaksi                  # Create transaksi
```

**Versioning:**
- URL-based versioning: `/api/v1/...`, `/api/v2/...`
- Breaking changes = versi baru; backward-compatible changes = versi yang sama
- Versi lama di-support minimal 6 bulan setelah versi baru rilis

**Tenant context:**
- Setiap request **wajib** menyertakan tenant context
- Mekanisme: `X-Tenant-ID` header (internal) atau subdomain (external)
- Request tanpa tenant context ditolak dengan `400 Bad Request`

**Authentication:**

| Metode | Digunakan Untuk | Mekanisme |
|--------|-----------------|-----------|
| JWT | Internal users (dashboard, portal) | Bearer token di `Authorization` header |
| API Key | External integrations (webhook, third-party) | `X-API-Key` header |
| OAuth2 | Self-service portal ([ADR-K023](./ADR-K023-dashboard-portal.md)) | Authorization code flow |

**Authorization:**
- RBAC di-enforce di API layer — setiap endpoint memiliki permission yang diperlukan
- Permission di-check berdasarkan role user (dari JWT) atau scope API key
- Unauthorized request ditolak dengan `403 Forbidden`

**Rate limiting:**
```
rate_limit_config:
  default_per_minute: 60                  # Per API key / per user
  default_per_tenant_minute: 1000         # Per tenant aggregate
  burst_allowance: 10                     # Burst di atas limit (short-term)
  custom_limits:                          # Override per endpoint
    "POST /api/v1/transaksi": 30          # Transaksi lebih ketat
    "GET /api/v1/nasabah": 120            # Read lebih longgar
```

**Pagination — cursor-based:**
```json
{
  "data": [...],
  "meta": {
    "cursor": "eyJpZCI6IjAxOGYuLi4ifQ==",
    "has_more": true,
    "per_page": 25
  }
}
```

- Cursor-based (bukan offset) — performa konsisten di dataset besar
- Default `per_page`: 25, max: 100
- Cursor di-encode sebagai opaque string (base64 dari ID terakhir)

**Response format — consistent envelope:**
```json
// Success
{
  "data": { ... },
  "meta": {
    "request_id": "req_018f...",
    "timestamp": "2026-04-15T10:30:00Z"
  }
}

// Success (list)
{
  "data": [ ... ],
  "meta": {
    "cursor": "...",
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
  "meta": {
    "request_id": "req_018f...",
    "timestamp": "2026-04-15T10:30:00Z"
  }
}
```

### 7. Webhook System (Outbound)

Sistem eksternal bisa berlangganan event dari koperasi melalui webhook:

**Event yang tersedia:**

| Event | Deskripsi |
|-------|-----------|
| `nasabah.approved` | Nasabah baru disetujui dan aktif |
| `rekening.opened` | Rekening baru dibuka |
| `transaksi.created` | Transaksi baru dibuat (setoran, penarikan, angsuran) |
| `pinjaman.approved` | Pinjaman/pembiayaan disetujui |
| `pinjaman.overdue` | Angsuran melewati jatuh tempo |
| `shu.distributed` | SHU didistribusikan ke anggota |

**Webhook delivery flow:**
```
Event terjadi di koperasi
        │
        v
┌─────────────────────────┐
│ Lookup subscribers      │
│ untuk event ini         │
└────────┬────────────────┘
         │ per subscriber
         v
┌─────────────────────────┐
│ Build payload + sign    │
│ POST ke subscriber URL  │
└────────┬────────────────┘
         │
    ┌────┴────────┐
    v             v
┌────────┐  ┌──────────────────┐
│ 2xx OK │  │ Gagal (timeout/  │
│ Done   │  │ 4xx/5xx)         │
└────────┘  └────────┬─────────┘
                     │ retry
                     v
            ┌──────────────────┐
            │ Exponential      │
            │ backoff retry    │
            │ 1m, 5m, 15m,    │
            │ 1h, 4h           │
            │ Max: 5 retries   │
            └────────┬─────────┘
                     │ semua gagal
                     v
            ┌──────────────────┐
            │ Dead Letter Queue│
            │ Admin notifikasi │
            └──────────────────┘
```

**Payload format:**
```json
{
  "id": "evt_018f...",
  "event": "transaksi.created",
  "timestamp": "2026-04-15T10:30:00Z",
  "tenant_id": "018f...",
  "data": {
    "transaction_id": "018f...",
    "type": "setoran",
    "amount": 500000,
    "account_id": "018f..."
  }
}
```

**Signature verification:**
```
X-Webhook-Signature: sha256=<HMAC-SHA256(payload, secret)>
X-Webhook-Timestamp: 2026-04-15T10:30:00Z
```

Subscriber memverifikasi signature menggunakan shared secret untuk memastikan payload berasal dari sistem koperasi dan tidak di-tamper.

**Data model webhook subscription:**
```
webhook_subscription
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Konfigurasi ──
├── name                  VARCHAR NOT NULL (label deskriptif)
├── url                   VARCHAR NOT NULL (endpoint tujuan)
├── events                JSONB NOT NULL (list event yang disubscribe)
├── secret                VARCHAR NOT NULL (untuk HMAC signature)
├── is_active             BOOLEAN DEFAULT true
│
├── ── Headers ──
├── custom_headers        JSONB DEFAULT '{}' (header tambahan per request)
│
├── ── Status ──
├── last_triggered_at     TIMESTAMPTZ (nullable)
├── last_status           ENUM (success, failed) (nullable)
├── consecutive_failures  INTEGER DEFAULT 0
├── disabled_at           TIMESTAMPTZ (nullable, auto-disable jika terlalu banyak gagal)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Aturan webhook:**
- Delivery **async** — tidak blocking proses utama
- Retry dengan **exponential backoff**: 1 menit, 5 menit, 15 menit, 1 jam, 4 jam (max 5 retries)
- Jika 10 consecutive failures: subscription **auto-disabled** — admin harus re-enable manual
- **Dead letter queue** untuk event yang gagal dikirim setelah semua retry
- Admin bisa lihat delivery status per subscription di dashboard
- Admin bisa **replay** event dari dead letter queue setelah fix

### 8. Inbound API (Sistem Eksternal Push Data)

Endpoint untuk menerima data dari sistem eksternal:

| Endpoint | Sumber Data | Format |
|----------|-------------|--------|
| `POST /api/v1/inbound/payroll` | Sistem payroll sekolah | CSV / JSON |
| `POST /api/v1/inbound/enrollment` | Management Sekolah | JSON |
| `POST /api/v1/inbound/bank-statement` | Bank / Admin upload | CSV (MT940) / XLSX |

**Flow inbound data:**
```
Sistem eksternal POST data
        │
        v
┌─────────────────────────┐
│ Autentikasi: API Key    │
│ + IP whitelist check    │
└────────┬────────────────┘
         │ valid
         v
┌─────────────────────────┐
│ Validasi format & data  │
│ Parse → staging table   │
└────────┬────────────────┘
         │
    ┌────┴────┐
    v         v
┌────────┐ ┌──────────┐
│ Valid  │ │ Invalid  │
│ rows   │ │ rows     │
└───┬────┘ └────┬─────┘
    │            │
    v            v
┌────────┐ ┌──────────────┐
│Process │ │ Error report │
│& apply │ │ (downloadable│
└────────┘ │  CSV)        │
           └──────────────┘
```

**Aturan inbound API:**
- Authentication: **API Key** + **IP whitelist** (double security)
- Data masuk ke **staging table** dulu — tidak langsung ke tabel utama
- Validasi dilakukan per row — row valid diproses, row invalid dikumpulkan di error report
- Error report bisa di-download oleh admin untuk koreksi dan re-upload
- **Idempotent**: upload yang sama (by checksum) tidak diproses ulang
- Rate limit: max 10 upload per jam per API key (configurable)

**Payroll inbound ([ADR-K020](./ADR-K020-payroll-deduction.md)):**
- Format: CSV dengan kolom `employee_id, deduction_type, amount, period`
- Mapping `employee_id` → `nasabah.school_entity_id` → `nasabah_id`
- Deduction types: simpanan_wajib, angsuran_pinjaman, tabungan, custom

**Bank statement inbound:**
- Format: CSV (MT940 standard) atau XLSX
- Digunakan untuk **reconciliation** — matching transaksi di sistem dengan mutasi bank
- Unmatched entries di-flag untuk review manual

### 9. Data Sync Strategy

Strategi sinkronisasi data antar sistem:

```
┌─────────────────────────────────────────────┐
│              Sync Strategy Matrix            │
├─────────────────┬───────────┬───────────────┤
│ Skenario        │ Primary   │ Fallback      │
├─────────────────┼───────────┼───────────────┤
│ School → Koop   │ Event-driven (near RT)    │ Batch (scheduled)   │
│ Payment Gateway │ Webhook callback          │ Polling (5 min)     │
│ Accounting Exp  │ Scheduled batch           │ Manual trigger      │
│ Regulatory Rep  │ Scheduled batch           │ Manual trigger      │
│ Payroll Upload  │ Inbound API               │ Manual CSV upload   │
│ Bank Statement  │ Inbound API               │ Manual CSV upload   │
└─────────────────┴───────────┴───────────────┘
```

**Prinsip sinkronisasi:**

- **Event-driven preferred** — near real-time, minimal latency
- **Batch sync sebagai fallback** — jika event-driven gagal atau tidak tersedia
- **Idempotent operations** — event/data yang sama diproses multiple kali menghasilkan hasil yang sama
- Setiap record yang di-sync memiliki `external_id` + `sync_version` untuk deduplication
- **Conflict resolution**: last-write-wins dengan audit trail lengkap

```
Conflict resolution flow:
  1. Incoming data memiliki timestamp lebih baru?
     ├── Ya → Apply update, simpan versi lama di audit
     └── Tidak → Skip, log sebagai "stale update"
  
  2. Jika data yang conflict adalah data kritis (saldo, status pinjaman)?
     └── Masukkan ke manual_resolution_queue → admin review
```

**Konfigurasi sync per integrasi:**
```
sync_config:
  mode: "event" | "batch" | "both"
  batch_interval: "1h" | "6h" | "24h"     # Interval batch sync
  batch_time: "02:00"                       # Waktu eksekusi batch
  retry_failed: true                        # Auto-retry failed syncs
  max_retry: 3
  conflict_mode: "last_write_wins" | "manual_queue"
```

### 10. Error Handling & Resilience

Semua integrasi harus fault-tolerant — kegagalan satu integrasi tidak boleh mengganggu sistem utama.

**Retry strategy:**
```
Retry dengan exponential backoff:
  Attempt 1: immediate
  Attempt 2: setelah 1 menit
  Attempt 3: setelah 5 menit
  Attempt 4: setelah 15 menit
  Attempt 5: setelah 1 jam
  Setelah 5x gagal → Dead Letter Queue + admin alert
```

**Circuit breaker:**
```
Circuit breaker per integrasi:
  CLOSED  → Normal operation, semua request diteruskan
              │ 5 failures dalam 1 menit
              v
  OPEN    → Semua request langsung gagal (fast-fail), tidak hit external system
              │ setelah 30 detik
              v
  HALF-OPEN → 1 test request diteruskan
              │ sukses → CLOSED
              │ gagal  → OPEN (reset timer)
```

**Dead letter queue:**
```
integrasi_dead_letter
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Context ──
├── integration_type      VARCHAR NOT NULL (payment, webhook, sync, export)
├── operation             VARCHAR NOT NULL (misal: "payment_callback")
├── payload               JSONB NOT NULL (data asli yang gagal diproses)
│
├── ── Error ──
├── error_message         TEXT NOT NULL
├── error_code            VARCHAR (nullable)
├── retry_count           INTEGER NOT NULL
├── last_retry_at         TIMESTAMPTZ
│
├── ── Resolution ──
├── status                ENUM (pending, retrying, resolved, discarded)
├── resolved_at           TIMESTAMPTZ (nullable)
├── resolved_by           UUID (nullable)
├── resolution_notes      TEXT (nullable)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    └── created_by        UUID
```

**Integration health dashboard:**
- Status per integrasi: healthy / degraded / down
- Metrik: success rate, avg response time, error count (last 24h)
- Alert jika success rate < 95% atau response time > 5 detik
- Admin bisa lihat dead letter queue dan retry/discard manual

### 11. Security

Keamanan berlapis untuk semua integrasi:

**API Key management:**
```
api_key
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Identitas ──
├── name                  VARCHAR NOT NULL (label deskriptif)
├── key_prefix            VARCHAR NOT NULL (misal: "sk_live_")
├── key_hash              VARCHAR NOT NULL (bcrypt/argon2 hash — key asli tidak disimpan)
├── key_hint              VARCHAR NOT NULL (4 karakter terakhir untuk identifikasi)
│
├── ── Scope ──
├── scopes                JSONB NOT NULL (list permission yang diizinkan)
├── ip_whitelist          JSONB DEFAULT '[]' (list IP yang diizinkan, kosong = semua)
├── allowed_endpoints     JSONB DEFAULT '[]' (list endpoint yang diizinkan, kosong = semua dalam scope)
│
├── ── Rate Limit ──
├── rate_limit_per_minute INTEGER DEFAULT 60
│
├── ── Status ──
├── is_active             BOOLEAN DEFAULT true
├── expires_at            TIMESTAMPTZ (nullable, null = tidak expire)
├── last_used_at          TIMESTAMPTZ (nullable)
├── revoked_at            TIMESTAMPTZ (nullable)
├── revoked_by            UUID (nullable)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Aturan keamanan:**

| Aspek | Implementasi |
|-------|-------------|
| API Key storage | Hash only — key asli ditampilkan **sekali** saat create, tidak bisa di-retrieve |
| API Key rotation | Generate key baru → aktifkan → revoke key lama (grace period configurable) |
| Webhook signature | HMAC-SHA256 — subscriber wajib verify sebelum proses |
| OAuth2 | Authorization code flow untuk portal nasabah/orang tua |
| IP whitelist | Configurable per API key — untuk server-to-server integrations |
| TLS | **Wajib** untuk semua komunikasi eksternal — reject HTTP (non-TLS) |
| PII in logs | Mask data sensitif (nomor identitas, nomor HP, saldo) di semua log |
| Encryption at rest | API keys, credentials, dan secrets di-encrypt menggunakan AES-256-GCM |
| Request signing | Inbound webhook dari payment provider harus di-verify signature-nya |
| Audit trail | Setiap API call tercatat: who, when, what, from where (IP) |

### 12. Data Model — Integrasi Config & Log

**Konfigurasi integrasi per tenant:**
```
integrasi_config
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Identitas ──
├── integration_type      VARCHAR NOT NULL (school_sync, payment_gateway, accounting_export,
│                                          regulatory_report, whatsapp, webhook_outbound)
├── name                  VARCHAR NOT NULL (label deskriptif)
├── provider              VARCHAR (nullable, misal: "midtrans", "xendit", "jurnal_id")
│
├── ── Konfigurasi ──
├── config                JSONB NOT NULL (konfigurasi spesifik per type — lihat Section 1-5)
├── credentials           JSONB NOT NULL (encrypted — API keys, secrets)
│
├── ── Status ──
├── is_active             BOOLEAN DEFAULT true
├── health_status         ENUM (healthy, degraded, down) DEFAULT 'healthy'
├── last_health_check     TIMESTAMPTZ (nullable)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Log integrasi (sync/API call):**
```
integrasi_log
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── config_id             UUID (FK → integrasi_config)
│
├── ── Request ──
├── direction             ENUM (inbound, outbound)
├── method                VARCHAR (GET, POST, PUT, CALLBACK)
├── endpoint              VARCHAR NOT NULL
├── request_headers       JSONB (nullable, sanitized — tanpa credentials)
├── request_body          JSONB (nullable, sanitized — tanpa PII)
│
├── ── Response ──
├── response_status       INTEGER (nullable, HTTP status code)
├── response_body         JSONB (nullable, truncated jika terlalu besar)
├── duration_ms           INTEGER (nullable, response time in ms)
│
├── ── Status ──
├── status                ENUM (success, failed, timeout, rejected)
├── error_message         TEXT (nullable)
│
├── ── Context ──
├── context_type          VARCHAR (nullable, misal: "transaction", "nasabah")
├── context_id            UUID (nullable)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    └── created_by        UUID
```

**Catatan integrasi_log:**
- **Tidak menggunakan Vernon pattern** — write-heavy, query by filter (bukan listing dengan JOIN)
- Log di-**rotate** setelah 90 hari (configurable) — pindah ke cold storage atau delete
- Data sensitif di-sanitize sebelum disimpan (mask API keys, PII, credentials)
- Index: `(tenant_id, config_id, created_at)`, `(tenant_id, status, created_at)`

### 13. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin | System |
|------------|--------|------------|---------|-------|--------|
| View integrasi config | - | - | v | v | - |
| Configure integrasi | - | - | - | v | - |
| Enable/disable integrasi | - | - | - | v | - |
| Manage API keys | - | - | - | v | - |
| Manage webhook subscriptions | - | - | - | v | - |
| View integrasi log | - | v | v | v | - |
| View dead letter queue | - | - | v | v | - |
| Retry/discard dead letter | - | - | v | v | - |
| View integration health | - | v | v | v | - |
| Trigger manual sync | - | - | v | v | - |
| Trigger manual export | - | - | v | v | - |
| Upload inbound data (CSV) | - | v | v | v | - |
| Execute sync/webhook/callback | - | - | - | - | v |

**Catatan:**
- **Admin** adalah satu-satunya role yang bisa konfigurasi integrasi — mengubah credentials, URL, dan settings
- **Manager** bisa monitoring dan trigger manual — tapi tidak bisa ubah konfigurasi
- **Supervisor** hanya bisa view log dan upload data inbound
- **System** adalah role internal untuk proses otomatis (scheduler, event handler) — bukan user role
- Teller **tidak memiliki akses** ke modul integrasi — bukan tanggung jawabnya

### 14. Dual-Mode Terminology

API responses menyesuaikan terminologi berdasarkan `coop_type` tenant:

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|-------|------------------------|------------------------|
| API response labels | Bunga, Pinjaman, Anggota | Bagi Hasil, Pembiayaan, Nasabah |
| Webhook event naming | `loan.approved` | `pembiayaan.approved` |
| Notification templates | Salam umum | Salam islami |
| Export labels (akuntansi) | Pendapatan Bunga | Pendapatan Bagi Hasil |
| Regulatory report header | Koperasi [nama] | BMT [nama] |

**Implementasi:**
- API response menyertakan field `labels` di `meta` jika diminta (`?include_labels=true`)
- Webhook payload menyertakan `coop_type` agar subscriber bisa handle sesuai mode
- Export akuntansi menggunakan COA label yang sesuai mode
- Perbedaan **hanya di level presentasi/label** — data model dan business logic tetap sama

### 15. Vernon _rels dan _data Structure

**integrasi_config _rels/_data:**
```json
// _rels
{
  "tenant_id": "018f..."
}

// _data
{
  "tenant": {
    "id": "018f...",
    "name": "Koperasi Al-Hikmah",
    "coop_type": "islamic"
  }
}
```

**webhook_subscription _rels/_data:**
```json
// _rels
{
  "tenant_id": "018f..."
}

// _data
{
  "tenant": {
    "id": "018f...",
    "name": "Koperasi Al-Hikmah"
  }
}
```

**api_key _rels/_data:**
```json
// _rels
{
  "tenant_id": "018f..."
}

// _data
{
  "tenant": {
    "id": "018f...",
    "name": "Koperasi Al-Hikmah"
  }
}
```

**SyncEngine triggers:**
- `TenantUpdatedEvent` → update `_data.tenant` di semua integrasi_config, webhook_subscription, dan api_key tenant tersebut

**Catatan:**
- integrasi_log dan integrasi_dead_letter **tidak menggunakan Vernon pattern** — write-heavy, read by query
- Vernon hanya diterapkan pada entity yang sering di-list (config, subscription, api_key)

## Consequences

### Positif

- **Extensible** — arsitektur plugin-based memudahkan penambahan integrasi baru tanpa mengubah core
- **Resilient** — circuit breaker, retry, dan dead letter queue mencegah kegagalan integrasi mengganggu sistem utama
- **Secure** — API key hashing, HMAC signature, IP whitelist, dan TLS memberikan keamanan berlapis
- **Traceable** — setiap API call dan sync event tercatat di integrasi_log dengan full context
- **Multi-tenant** — setiap tenant bisa memiliki konfigurasi integrasi yang berbeda
- **School integration** — sinkronisasi event-driven dengan Management Sekolah mengurangi input manual
- **Payment ready** — VA dan QRIS siap digunakan untuk cashless payment
- **Regulatory compliant** — auto-generate laporan dalam format yang diminta regulator
- **Webhook ecosystem** — sistem eksternal bisa berlangganan event tanpa polling

### Negatif

- **Complexity** — banyak integrasi berarti banyak failure points yang harus di-monitor
- **Cost** — payment gateway, WhatsApp API, dan SMS gateway berbayar per transaksi
- **Provider dependency** — bergantung pada third-party (Midtrans/Xendit, Meta, dll)
- **Security surface** — setiap integrasi membuka attack surface baru yang harus diamankan
- **Testing complexity** — integrasi membutuhkan mock/sandbox untuk testing — tidak bisa test dengan provider real di CI/CD

### Mitigasi

- **Monitoring dashboard** — health status per integrasi, real-time alert, dead letter queue visibility
- **Provider abstraction** — interface/strategy pattern per provider — switch provider tanpa ubah business logic
- **Sandbox mode** — semua provider didukung sandbox/test mode untuk development dan CI/CD
- **Cost optimization** — batching, deduplication, dan priority-based sending mengurangi cost per transaction
- **Security audit** — regular rotation API keys, IP whitelist review, dan log monitoring
- **Graceful degradation** — jika integrasi down, core system tetap berjalan — fitur yang bergantung pada integrasi di-disable sementara dengan pesan yang jelas ke user

## Alternatives Considered

### A. Direct Point-to-Point Integration (Tanpa API Gateway)

Setiap modul terhubung langsung ke sistem eksternal tanpa layer integrasi terpusat.

**Ditolak** karena: duplikasi logika (retry, auth, logging) di setiap modul. Tidak ada single point untuk monitoring dan security enforcement. Penambahan integrasi baru membutuhkan perubahan di banyak tempat. Sulit di-maintain seiring bertambahnya jumlah integrasi.

### B. GraphQL Instead of REST

Menggunakan GraphQL sebagai API layer utama.

**Ditolak** karena: REST lebih familiar untuk target developer (pengembang koperasi/sekolah di Indonesia). GraphQL menambah complexity (schema management, N+1 query protection, caching complexity) yang belum justified untuk use case koperasi. REST + cursor pagination sudah cukup untuk kebutuhan saat ini. GraphQL bisa dipertimbangkan di fase 2 jika ada kebutuhan query flexibility yang tinggi dari client.

### C. Single Payment Provider (Hardcode)

Memilih satu payment provider (misal: Midtrans saja) dan hardcode integrasinya.

**Ditolak** karena: lock-in ke satu provider berisiko jika provider mengubah kebijakan, menaikkan harga, atau mengalami downtime. Beberapa tenant mungkin sudah memiliki hubungan dengan provider tertentu. Arsitektur pluggable (strategy pattern) memungkinkan switch provider dengan effort minimal.
