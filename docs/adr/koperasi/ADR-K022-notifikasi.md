# ADR-K022: Notifikasi & Komunikasi

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Sistem koperasi sekolah memiliki banyak event yang memerlukan komunikasi ke berbagai stakeholder:

- **Nasabah** perlu tahu saat transaksi terjadi, pinjaman disetujui, atau angsuran jatuh tempo
- **Orang tua** perlu dipantau soal belanja dan saldo anak
- **Staff operasional** perlu alert untuk approval queue, posisi kas, dan overdue
- **Management** perlu summary dan reminder regulasi

Di Indonesia, **WhatsApp** adalah channel komunikasi dominan — digunakan oleh hampir semua segmen, termasuk orang tua siswa di pedesaan. Sistem harus mendukung multi-channel dengan WhatsApp sebagai primary.

## Decision

### 1. Notification Channels

Sistem mendukung beberapa channel notifikasi:

```
notification_channel:
├── WHATSAPP          Primary channel (via WhatsApp Business API)
├── SMS               Fallback jika WhatsApp gagal/tidak tersedia
├── EMAIL             Untuk dokumen formal (statement, laporan)
├── PUSH              In-app push notification (jika mobile/web app)
└── IN_APP            Notification bell di dashboard
```

**Prioritas channel:**
```
Default delivery order:
1. IN_APP           ← Selalu dikirim (bell notification)
2. WHATSAPP         ← Primary external channel
3. SMS              ← Fallback jika WA gagal
4. EMAIL            ← Untuk dokumen/attachment
5. PUSH             ← Jika app terinstall
```

### 2. Notification Events — Comprehensive List

**NASABAH events:**

| Event | Channel | Priority | Opt-out |
|-------|---------|----------|---------|
| Application approved/rejected | WA, IN_APP | HIGH | Tidak |
| Transaction receipt (setoran/penarikan/angsuran) | WA, IN_APP | HIGH | Tidak |
| Balance below minimum | WA, IN_APP | MEDIUM | Ya |
| Balance approaching limit (KYC) | WA, IN_APP | MEDIUM | Ya |
| Loan application approved/rejected | WA, IN_APP | HIGH | Tidak |
| Loan disbursed | WA, IN_APP | HIGH | Tidak |
| Installment due reminder (7d before) | WA, IN_APP | MEDIUM | Ya |
| Installment due reminder (3d before) | WA, IN_APP | MEDIUM | Ya |
| Installment due reminder (1d before) | WA, IN_APP | HIGH | Tidak |
| Installment overdue alert | WA, SMS, IN_APP | CRITICAL | Tidak |
| Deposito maturity reminder (30d) | WA, IN_APP | LOW | Ya |
| Deposito maturity reminder (7d) | WA, IN_APP | MEDIUM | Ya |
| Deposito maturity on date | WA, IN_APP | HIGH | Tidak |
| KYC upgrade notification | WA, IN_APP | MEDIUM | Tidak |
| SHU distribution notification | WA, EMAIL, IN_APP | HIGH | Tidak |
| Payroll deduction summary | WA, IN_APP | HIGH | Tidak |
| Account frozen/unfrozen | WA, IN_APP | CRITICAL | Tidak |
| Password/PIN changed | WA, IN_APP | CRITICAL | Tidak |

**PARENT events (untuk akun anak):**

| Event | Channel | Priority | Opt-out |
|-------|---------|----------|---------|
| Child daily spending summary | WA | MEDIUM | Ya |
| Child single transaction alert | WA, IN_APP | LOW | Ya |
| Top-up confirmation | WA, IN_APP | MEDIUM | Tidak |
| Spending limit exceeded attempt | WA, IN_APP | HIGH | Tidak |
| Child low balance alert | WA, IN_APP | MEDIUM | Ya |
| Card frozen/lost report | WA, IN_APP | CRITICAL | Tidak |
| E-wallet config changed | WA, IN_APP | HIGH | Tidak |

**STAFF/OPERATIONAL events:**

| Event | Channel | Priority | Opt-out |
|-------|---------|----------|---------|
| Pending approval queue (new item) | IN_APP, PUSH | MEDIUM | Ya |
| Pending approval reminder (aging) | WA, IN_APP | HIGH | Tidak |
| Teller session reminder (close session) | IN_APP, PUSH | MEDIUM | Tidak |
| Cash position below minimum | WA, IN_APP | CRITICAL | Tidak |
| Cash position above maximum | WA, IN_APP | HIGH | Tidak |
| Dormant account alerts (batch) | EMAIL, IN_APP | LOW | Ya |
| NPL status change | WA, IN_APP | HIGH | Tidak |
| Stock reorder point alert (K019) | IN_APP, PUSH | MEDIUM | Ya |
| Payroll batch uploaded | IN_APP | MEDIUM | Tidak |

**MANAGEMENT events:**

| Event | Channel | Priority | Opt-out |
|-------|---------|----------|---------|
| Daily summary report | EMAIL, IN_APP | MEDIUM | Ya |
| Weekly summary report | EMAIL, IN_APP | MEDIUM | Ya |
| Monthly KPI summary | EMAIL, IN_APP | HIGH | Tidak |
| Regulatory report due reminder (30d) | WA, EMAIL, IN_APP | HIGH | Tidak |
| Regulatory report due reminder (7d) | WA, EMAIL, IN_APP | CRITICAL | Tidak |
| SHU calculation ready for review | WA, IN_APP | HIGH | Tidak |
| Large transaction alert (above threshold) | WA, IN_APP | HIGH | Tidak |

### 3. Channel Preference

Setiap nasabah/user bisa mengatur preferensi channel:

```
notifikasi_preference
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── user_id               UUID (FK → user, bisa nasabah atau staff)
│
├── ── Channel Preferences ──
├── primary_channel       ENUM (whatsapp, sms, email, push) DEFAULT 'whatsapp'
├── secondary_channel     ENUM (whatsapp, sms, email, push) DEFAULT 'sms'
├── whatsapp_number       VARCHAR (nullable, nomor WA — default: phone dari nasabah)
├── sms_number            VARCHAR (nullable, default: phone dari nasabah)
├── email_address         VARCHAR (nullable, default: email dari nasabah)
│
├── ── Opt-out ──
├── opted_out_events      JSONB DEFAULT '[]' (list event_type yang di-opt-out)
│
├── ── Quiet Hours ──
├── quiet_hours_enabled   BOOLEAN DEFAULT false
├── quiet_start           TIME (nullable, misal: 22:00)
├── quiet_end             TIME (nullable, misal: 06:00)
│
├── ── Language ──
├── language              ENUM (id, en) DEFAULT 'id'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Aturan:**
- Default channel: WhatsApp — karena penetrasi tertinggi di Indonesia
- Nasabah bisa opt-out dari notifikasi non-critical — tapi **tidak bisa** opt-out dari CRITICAL dan HIGH priority yang tidak bisa di-opt-out (lihat tabel di Section 2)
- Quiet hours: notifikasi non-critical di-delay sampai quiet hours selesai. CRITICAL tetap dikirim
- IN_APP selalu dikirim untuk semua event — tidak bisa di-opt-out (silent, non-intrusive)

### 4. Template Management

Admin mengelola template notifikasi per event:

```
notifikasi_template
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Identitas ──
├── event_type            VARCHAR NOT NULL (misal: TRANSACTION_RECEIPT)
├── channel               ENUM (whatsapp, sms, email, push, in_app)
├── language              ENUM (id, en) DEFAULT 'id'
├── coop_type             ENUM (general, islamic, both) DEFAULT 'both'
│
├── ── Template ──
├── subject               VARCHAR (nullable, untuk email)
├── body_template         TEXT NOT NULL (template dengan placeholder)
├── variables             JSONB NOT NULL (daftar variable yang tersedia)
│
├── ── WhatsApp Specific ──
├── wa_template_name      VARCHAR (nullable, nama template di WA Business API)
├── wa_template_status    ENUM (pending, approved, rejected) (nullable)
│
├── ── Status ──
├── is_active             BOOLEAN DEFAULT true
├── version               INTEGER DEFAULT 1
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

**Contoh template:**

```
Event: TRANSACTION_RECEIPT
Channel: WhatsApp
Language: id
coop_type: general

Body:
"Halo {{nasabah_name}},
Transaksi berhasil:
- Jenis: {{transaction_type}}
- Rekening: {{account_number}}
- Nominal: Rp {{amount}}
- Saldo akhir: Rp {{ending_balance}}
- Waktu: {{transaction_date}}

Terima kasih,
Koperasi {{tenant_name}}"

Variables: ["nasabah_name", "transaction_type", "account_number", "amount", "ending_balance", "transaction_date", "tenant_name"]
```

**Contoh template BMT (islamic):**

```
Event: TRANSACTION_RECEIPT
Channel: WhatsApp
Language: id
coop_type: islamic

Body:
"Assalamu'alaikum {{nasabah_name}},
Transaksi berhasil:
- Jenis: {{transaction_type}}
- Rekening: {{account_number}}
- Nominal: Rp {{amount}}
- Saldo akhir: Rp {{ending_balance}}
- Waktu: {{transaction_date}}

Jazakumullahu khairan,
BMT {{tenant_name}}"
```

**Aturan template:**
- Template bisa **berbeda per coop_type** — salam pembuka/penutup berbeda
- Template `both` berlaku jika tidak ada template spesifik per coop_type
- Admin bisa customize body — tapi **tidak bisa menghapus variable wajib** (misal: amount di transaction receipt)
- Setiap edit template menghasilkan **version baru** (audit trail)
- Template default disediakan untuk semua event saat tenant onboarding

### 5. WhatsApp Business API Integration

Integrasi dengan WhatsApp Business API untuk pengiriman notifikasi transaksional:

```
wa_config (per tenant):
  provider: "official" | "twilio" | "wablas" | "fonnte"
  api_key: "encrypted_..."
  phone_number_id: "..."
  business_account_id: "..."
  webhook_verify_token: "..."
```

**Aturan WhatsApp:**
- Menggunakan **template messages** untuk notifikasi yang dikirim di luar 24-hour window (sebagian besar notifikasi kita)
- Template harus di-approve oleh Meta sebelum bisa digunakan — status tracking di `wa_template_status`
- Provider configurable per tenant — mendukung official API atau third-party (Twilio, Wablas, Fonnte)
- Rate limiting sesuai kebijakan WhatsApp Business API
- Fallback ke SMS jika pengiriman WA gagal 3x

### 6. Delivery Tracking

Melacak status pengiriman setiap notifikasi:

```
notifikasi_log
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Identitas ──
├── event_type            VARCHAR NOT NULL
├── template_id           UUID (FK → notifikasi_template)
├── recipient_user_id     UUID (FK → user)
├── recipient_contact     VARCHAR NOT NULL (nomor/email tujuan)
│
├── ── Channel ──
├── channel               ENUM (whatsapp, sms, email, push, in_app)
├── priority              ENUM (low, medium, high, critical)
│
├── ── Content ──
├── subject               VARCHAR (nullable)
├── body                  TEXT NOT NULL (rendered template)
│
├── ── Delivery Status ──
├── status                ENUM (queued, sending, sent, delivered, read, failed, bounced)
├── sent_at               TIMESTAMPTZ (nullable)
├── delivered_at          TIMESTAMPTZ (nullable)
├── read_at               TIMESTAMPTZ (nullable)
├── failed_at             TIMESTAMPTZ (nullable)
├── failure_reason        TEXT (nullable)
│
├── ── Retry ──
├── retry_count           INTEGER DEFAULT 0
├── max_retries           INTEGER DEFAULT 3
├── next_retry_at         TIMESTAMPTZ (nullable)
│
├── ── External Reference ──
├── external_id           VARCHAR (nullable, message ID dari provider)
│
├── ── Context ──
├── context_type          VARCHAR (nullable, misal: "transaction")
├── context_id            UUID (nullable, misal: transaction_id)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    └── created_by        UUID
```

**Status flow:**
```
QUEUED → SENDING → SENT → DELIVERED → READ
                     │
                     └→ FAILED (retry) → QUEUED (jika retry < max)
                                       → FAILED (permanent, fallback to secondary channel)
```

### 7. Opt-in/Opt-out Rules

| Category | Opt-out Allowed | Contoh |
|----------|----------------|--------|
| CRITICAL | Tidak | Account frozen, security alert, overdue |
| HIGH (transaction) | Tidak | Transaction receipt, loan approved |
| HIGH (reminder) | Tidak | Installment due 1d before, regulatory due |
| MEDIUM | Ya | Balance alert, daily summary, due 7d before |
| LOW | Ya | Dormant alert, deposito maturity 30d |

**Aturan:**
- Default: semua notifikasi AKTIF — nasabah harus explicitly opt-out
- Opt-out disimpan per event_type di `notifikasi_preference.opted_out_events`
- Admin bisa set **tenant-level defaults** untuk event mana yang aktif/tidak

### 8. Batch Notifications

Untuk event massal yang mempengaruhi banyak nasabah:

```
Batch notification use cases:
├── SHU distribution (semua anggota)
├── Simpanan wajib monthly reminder (semua guru/staff)
├── Rate change announcement (semua nasabah tabungan/deposito)
├── System maintenance notice (semua user)
└── Payroll deduction summary (semua yang dipotong gaji)
```

**Aturan batch:**
- Batch notification di-queue dan diproses secara **asynchronous** — tidak blocking
- Rate limiting per channel: WA max 80 messages/second (sesuai tier), SMS max 30/second
- Progress tracking: admin bisa lihat berapa yang sudah dikirim / gagal / pending
- Batch bisa di-cancel selama masih ada yang QUEUED

### 9. Escalation — Channel Fallback

Jika primary channel gagal, sistem otomatis fallback:

```
Delivery attempt:
  1. Send via primary channel (misal: WhatsApp)
     └→ Gagal? Retry 3x dengan exponential backoff (1min, 5min, 15min)
         └→ Masih gagal?
             2. Send via secondary channel (misal: SMS)
                └→ Gagal? Retry 3x
                    └→ Masih gagal?
                        3. Log as PERMANENTLY_FAILED
                           Alert admin untuk follow-up manual
```

**Aturan:**
- Retry menggunakan **exponential backoff** — 1 menit, 5 menit, 15 menit
- Max retry per channel: 3 (configurable)
- Setelah semua channel gagal, admin mendapat alert
- CRITICAL priority: attempt ALL channels sekaligus (parallel), bukan sequential

### 10. In-App Notification Bell

Notifikasi di dalam dashboard aplikasi:

```
in_app_notification
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── user_id               UUID (FK → user)
│
├── ── Content ──
├── title                 VARCHAR NOT NULL
├── body                  TEXT NOT NULL
├── event_type            VARCHAR NOT NULL
├── priority              ENUM (low, medium, high, critical)
│
├── ── Status ──
├── is_read               BOOLEAN DEFAULT false
├── read_at               TIMESTAMPTZ (nullable)
│
├── ── Action ──
├── action_url            VARCHAR (nullable, deep link ke halaman terkait)
├── action_label          VARCHAR (nullable, misal: "Lihat Detail")
│
├── ── Context ──
├── context_type          VARCHAR (nullable)
├── context_id            UUID (nullable)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    └── created_by        UUID
```

**Aturan:**
- Semua event menghasilkan in_app notification — **selalu**, tidak bisa di-opt-out
- Badge count (unread count) ditampilkan di header dashboard
- Auto-expire setelah 90 hari (configurable) — soft delete
- Nasabah bisa mark as read individual atau bulk "mark all as read"

### 11. Vernon _rels dan _data Structure

**notifikasi_template _rels/_data:**

```json
// _rels
{
  "tenant_id": "018f..."
}

// _data — minimal, template jarang di-list dengan JOIN
{}
```

**notifikasi_log — tidak menggunakan Vernon pattern** karena:
- Log write-heavy, read by query (bukan listing)
- Tidak perlu denormalized read-cache
- Query by: tenant_id, user_id, event_type, status, date range

### 12. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View own notifications | v | v | v | v |
| View nasabah notification log | - | v | v | v |
| Manage templates | - | - | v | v |
| Configure WA integration | - | - | - | v |
| Send batch notification | - | - | v | v |
| Cancel batch notification | - | - | v | v |
| View delivery report | - | v | v | v |
| Configure tenant defaults | - | - | - | v |

### 13. Dual-Mode Terminology

Template content bervariasi per mode:

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|-------|------------------------|------------------------|
| Salam pembuka | "Halo" / "Yth." | "Assalamu'alaikum" |
| Salam penutup | "Terima kasih" | "Jazakumullahu khairan" |
| Loan reference | "Pinjaman" | "Pembiayaan" |
| Interest reference | "Bunga" | "Bagi Hasil / Margin" |
| Org name prefix | "Koperasi" | "BMT" |

Perbedaan dihandle di **template level** — ada template terpisah per coop_type, atau template `both` yang menggunakan mode-aware variable.

## Consequences

### Positif

- **Multi-channel** — reach semua stakeholder via channel yang mereka gunakan
- **WhatsApp-first** — sesuai realita penggunaan di Indonesia
- **Customizable** — template bisa disesuaikan per tenant dan per mode
- **Traceable** — setiap notifikasi dilacak: queued, sent, delivered, read, failed
- **Resilient** — escalation dan fallback mencegah notifikasi hilang
- **Respectful** — opt-out, quiet hours, dan priority levels menghormati preferensi user
- **Batch capable** — bisa kirim notifikasi massal untuk event yang mempengaruhi banyak nasabah

### Negatif

- **Cost** — WhatsApp Business API dan SMS berbayar per message
- **WA template approval** — proses approval Meta bisa memakan waktu 1-3 hari
- **Provider dependency** — bergantung pada third-party (WA API, SMS gateway)
- **Volume** — koperasi aktif bisa menghasilkan ribuan notifikasi per hari
- **Privacy** — data nomor HP dan email harus dilindungi

### Mitigasi

- Cost optimization: batching, deduplication, dan priority-based sending (tidak semua event perlu WA)
- Template pre-approval: siapkan semua template saat onboarding, sebelum go-live
- Multi-provider support: jika satu provider down, bisa switch ke provider lain
- Queue-based architecture: notifikasi di-queue, processed asynchronously — tidak ada spike
- Data encryption at rest untuk nomor HP dan email di database

## Alternatives Considered

### A. Email-First Approach

Menggunakan email sebagai primary channel.

**Ditolak** karena: penetrasi email di target user (orang tua siswa, guru sekolah) jauh lebih rendah dari WhatsApp di Indonesia. Banyak yang tidak rutin cek email. WhatsApp memiliki open rate > 90% vs email < 30%.

### B. Build WhatsApp Gateway Sendiri

Menggunakan WhatsApp Web API (unofficial) untuk menghindari biaya Business API.

**Ditolak** karena: melanggar Terms of Service WhatsApp, berisiko ban nomor, tidak reliable, dan tidak mendukung template messages. Official Business API lebih mahal tapi legal dan stable.

### C. Single Channel Only

Hanya mendukung satu channel (misal: WhatsApp saja).

**Ditolak** karena: tidak semua nasabah punya WhatsApp (terutama nasabah tua atau daerah dengan koneksi internet terbatas). SMS sebagai fallback penting untuk reliability. Email penting untuk dokumen formal (statement, laporan).

### D. Notification tanpa Template Management

Hardcode semua pesan notifikasi di code.

**Ditolak** karena: setiap tenant mungkin ingin customize pesan (nama koperasi, tone, bahasa). Dual-mode juga memerlukan variasi template. Hardcode membuat perubahan memerlukan deploy ulang.
