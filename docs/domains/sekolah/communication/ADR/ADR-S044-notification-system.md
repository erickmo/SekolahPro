# ADR-S044: Notification System (Sistem Notifikasi Multi-Channel)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Sistem notifikasi adalah **infrastruktur pendukung** yang menghubungkan semua domain SekolahPro ke pengguna akhir. Berbeda dari messaging (S043) yang bersifat percakapan dua arah, notifikasi adalah **alert satu arah** yang dipicu oleh event dari domain lain.

Contoh event yang memicu notifikasi:

| Domain Source | Event | Penerima | Urgensi |
|---------------|-------|----------|---------|
| S008 Attendance | Siswa tidak hadir (absent) | Orang tua | Tinggi |
| S009 Finance | Tagihan jatuh tempo | Orang tua | Tinggi |
| S009 Finance | Pembayaran diterima | Orang tua | Medium |
| S011 Grades | Nilai dipublikasikan | Orang tua | Medium |
| S012 Discipline | Pelanggaran tercatat | Orang tua | Tinggi |
| S018 Rapor | Rapor tersedia | Orang tua | Medium |
| S043 Messaging | Pesan baru | Penerima | Medium |
| S045 Announcement | Pengumuman baru | Target audience | Varies |
| S046 Correspondence | Surat perlu disposisi | Penerima disposisi | Tinggi |
| S047 Approval | Persetujuan diminta | Approver | Tinggi |

Channel notifikasi yang perlu didukung:

1. **Push notification**: Firebase Cloud Messaging (FCM) untuk mobile/web.
2. **SMS**: Untuk orang tua yang tidak punya smartphone — via SMS gateway lokal.
3. **Email**: Untuk dokumen formal dan laporan.
4. **WhatsApp**: Channel paling populer di Indonesia — via WhatsApp Business API.

### Mengapa CQRS (bukan Vernon)?

- **Write-heavy**: Setiap event dari 10+ domain menghasilkan notifikasi. Ratusan ribu per hari.
- **Fire-and-forget**: Notifikasi dikirim dan dicatat, tidak perlu JOIN atau relasi kompleks.
- **No read-cache needed**: Notification log dibaca sequentially (timeline), bukan queried by relationships.
- **Eventual consistency native**: Notifikasi bisa delay beberapa detik tanpa masalah.
- **Idempotency required**: Event yang sama tidak boleh mengirim notifikasi duplikat.
- **Asynchronous**: Pengiriman ke external channel (SMS, WhatsApp, email) harus async dengan retry.

## Decision

Menggunakan **CQRS Pattern** untuk notification system. Write side menerima event dan mengirim notifikasi. Read side menyediakan log/riwayat notifikasi per user.

### Table Schema

```sql
-- Template notifikasi per event type
CREATE TABLE notification_templates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Template identity
    event_type      VARCHAR(50) NOT NULL,
    channel         VARCHAR(20) NOT NULL,
    language        VARCHAR(5) NOT NULL DEFAULT 'id',

    -- Template content
    title_template  TEXT NOT NULL,
    body_template   TEXT NOT NULL,
    action_url_template TEXT,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_notif_template UNIQUE (tenant_id, company_id, event_type, channel, language),
    CONSTRAINT chk_notif_channel CHECK (channel IN ('push', 'sms', 'email', 'whatsapp', 'in_app')),
    CONSTRAINT chk_notif_language CHECK (language IN ('id', 'en', 'ar'))
);

-- Indexes
CREATE INDEX idx_notif_tpl_tenant_company ON notification_templates (tenant_id, company_id);
CREATE INDEX idx_notif_tpl_event ON notification_templates (event_type, channel);

-- Preferensi notifikasi per user
CREATE TABLE notification_preferences (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- User
    user_id         UUID NOT NULL,

    -- Preferences per event type
    preferences     JSONB NOT NULL DEFAULT '{}',

    -- Global settings
    quiet_hours_start TIME,
    quiet_hours_end   TIME,
    timezone        VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_notif_pref_user UNIQUE (tenant_id, company_id, user_id)
);

-- Indexes
CREATE INDEX idx_notif_pref_tenant_company ON notification_preferences (tenant_id, company_id);
CREATE INDEX idx_notif_pref_user ON notification_preferences (user_id);

-- Log notifikasi terkirim (write-heavy, append-only)
CREATE TABLE notification_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Target
    user_id         UUID NOT NULL,
    channel         VARCHAR(20) NOT NULL,

    -- Event source
    event_type      VARCHAR(50) NOT NULL,
    event_id        UUID,
    source_type     VARCHAR(50),
    source_id       UUID,

    -- Content (rendered from template)
    title           TEXT NOT NULL,
    body            TEXT NOT NULL,
    action_url      TEXT,
    payload         JSONB NOT NULL DEFAULT '{}',

    -- Delivery status
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    sent_at         TIMESTAMPTZ,
    delivered_at    TIMESTAMPTZ,
    read_at         TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    failure_reason  TEXT,
    retry_count     INT NOT NULL DEFAULT 0,

    -- Idempotency
    idempotency_key VARCHAR(100) NOT NULL,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_notif_idempotency UNIQUE (idempotency_key),
    CONSTRAINT chk_notif_log_channel CHECK (channel IN ('push', 'sms', 'email', 'whatsapp', 'in_app')),
    CONSTRAINT chk_notif_log_status CHECK (status IN ('pending', 'sent', 'delivered', 'read', 'failed'))
);

-- Indexes (optimized for write-heavy + timeline read)
CREATE INDEX idx_notif_log_tenant_company ON notification_logs (tenant_id, company_id);
CREATE INDEX idx_notif_log_user_created ON notification_logs (user_id, created_at DESC);
CREATE INDEX idx_notif_log_user_unread ON notification_logs (user_id) WHERE status IN ('sent', 'delivered') AND read_at IS NULL;
CREATE INDEX idx_notif_log_status_pending ON notification_logs (status, created_at) WHERE status = 'pending';
CREATE INDEX idx_notif_log_status_failed ON notification_logs (status, retry_count) WHERE status = 'failed' AND retry_count < 3;
CREATE INDEX idx_notif_log_event ON notification_logs (event_type, event_id);
CREATE INDEX idx_notif_log_source ON notification_logs (source_type, source_id);
CREATE INDEX idx_notif_log_idempotency ON notification_logs (idempotency_key);

-- Device tokens untuk push notification
CREATE TABLE user_device_tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- User
    user_id         UUID NOT NULL,

    -- Device
    platform        VARCHAR(10) NOT NULL,
    device_token    TEXT NOT NULL,
    device_name     VARCHAR(100),

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,
    last_used_at    TIMESTAMPTZ,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_device_token UNIQUE (device_token),
    CONSTRAINT chk_device_platform CHECK (platform IN ('android', 'ios', 'web'))
);

-- Indexes
CREATE INDEX idx_device_tenant_company ON user_device_tokens (tenant_id, company_id);
CREATE INDEX idx_device_user ON user_device_tokens (user_id) WHERE is_active = true;
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `event_type` | VARCHAR(50) | Format: `domain.event` e.g. `attendance.absent`, `finance.payment_due` |
| `idempotency_key` | VARCHAR(100), UNIQUE | Format: `{event_type}:{event_id}:{channel}:{user_id}` — mencegah duplikat |
| `title_template` / `body_template` | TEXT dengan placeholder | Template: `"Anak Anda {{student_name}} tidak hadir hari ini"` |
| `status` | 5-state lifecycle | pending → sent → delivered → read, atau → failed |
| `retry_count` | INT | Max 3 retries untuk channel eksternal (SMS, WhatsApp, email) |
| `quiet_hours` | TIME range | Notifikasi non-urgent di-queue selama jam tidur (22:00-06:00) |
| `payload` | JSONB | Data konteks lengkap untuk rendering di client (deep link, etc.) |
| `source_type` + `source_id` | Polymorphic reference | Link ke entity sumber: `student_attendances` + UUID |
| No `deleted_at` on logs | Append-only | Notification logs tidak pernah dihapus — audit trail |

### Notification Preferences Structure

```json
{
  "preferences": {
    "attendance.absent": {
      "channels": ["push", "whatsapp"],
      "enabled": true
    },
    "finance.payment_due": {
      "channels": ["push", "sms"],
      "enabled": true
    },
    "finance.payment_received": {
      "channels": ["push"],
      "enabled": true
    },
    "grade.published": {
      "channels": ["push"],
      "enabled": true
    },
    "announcement.new": {
      "channels": ["push"],
      "enabled": true
    },
    "messaging.new_message": {
      "channels": ["push"],
      "enabled": true
    }
  },
  "quiet_hours_start": "22:00",
  "quiet_hours_end": "06:00",
  "timezone": "Asia/Jakarta"
}
```

### Template Rendering

Template menggunakan placeholder `{{variable}}`:

```
Event: attendance.absent
Channel: whatsapp
Template:
  title: "Pemberitahuan Ketidakhadiran"
  body: "Yth. {{parent_name}}, anak Anda {{student_name}} ({{class_name}}) tidak hadir pada {{date}}. Mohon konfirmasi alasan ketidakhadiran. Terima kasih."
  action_url: "{{portal_url}}/children/{{student_id}}/attendance/{{date}}"

Rendered:
  title: "Pemberitahuan Ketidakhadiran"
  body: "Yth. Budi Santoso, anak Anda Ahmad Fadhil (VII-A) tidak hadir pada 15 April 2026. Mohon konfirmasi alasan ketidakhadiran. Terima kasih."
  action_url: "https://sekolahpro.id/portal/children/018f.../attendance/2026-04-15"
```

### Event-Driven Architecture

```
Domain Event Flow:
1. Domain (e.g. S008) emits event: AttendanceMarkedAbsent { student_id, date }
2. Notification Service subscribes to event
3. Service:
   a. Resolve recipients: student → guardians → user_ids
   b. For each user: check notification_preferences
   c. For each enabled channel:
      - Generate idempotency_key
      - Check if already sent (idempotency)
      - Render template with event data
      - Create notification_log with status='pending'
      - Enqueue to channel-specific worker
4. Channel worker:
   a. push: send via FCM
   b. sms: send via SMS gateway (Twilio, Zenziva, etc.)
   c. email: send via SMTP/SendGrid
   d. whatsapp: send via WhatsApp Business API (Fonnte, Wablas, etc.)
   e. Update status: sent/failed
5. If failed: retry up to 3 times with exponential backoff
```

### API Endpoints

```
# Notification Log (read side)
GET    /api/v1/notifications                           — My notifications (timeline, paginated)
GET    /api/v1/notifications/unread-count              — Unread count
POST   /api/v1/notifications/{id}/read                 — Mark as read
POST   /api/v1/notifications/read-all                  — Mark all as read

# Preferences
GET    /api/v1/notification-preferences                — Get my preferences
PUT    /api/v1/notification-preferences                — Update preferences

# Device tokens
POST   /api/v1/device-tokens                           — Register device token
DELETE /api/v1/device-tokens/{id}                      — Unregister device

# Templates (admin only)
GET    /api/v1/notification-templates                  — List templates
POST   /api/v1/notification-templates                  — Create template
PUT    /api/v1/notification-templates/{id}             — Update template
POST   /api/v1/notification-templates/{id}/preview     — Preview rendered template

# Manual send (admin only)
POST   /api/v1/notifications/send                      — Send manual notification to user(s)
POST   /api/v1/notifications/broadcast                 — Broadcast to audience (class, all parents, etc.)
```

## Consequences

### Positive

- **Multi-channel**: Mendukung 5 channel (push, SMS, email, WhatsApp, in-app) — pasti sampai ke orang tua.
- **Event-driven**: Tidak ada coupling langsung antar domain — notifikasi otomatis terpicu oleh event.
- **Idempotent**: Idempotency key mencegah notifikasi duplikat meski event di-replay.
- **User control**: Orang tua bisa pilih channel mana yang aktif per event type.
- **Template-based**: Admin bisa customize wording tanpa ubah kode.
- **Quiet hours**: Menghormati waktu istirahat — notifikasi non-urgent di-queue.
- **Audit trail**: Semua notifikasi tercatat — bisa diaudit kapan, ke siapa, channel apa, status apa.

### Negative / Trade-offs

- **External dependency**: SMS dan WhatsApp tergantung third-party provider — biaya per message dan rate limit.
- **WhatsApp template approval**: WhatsApp Business API memerlukan template yang diapprove Meta — proses bisa lama.
- **SMS cost**: SMS di Indonesia ~Rp350-500 per pesan. Untuk 500 parents × 20 SMS/bulan = ~Rp5 juta/bulan.
- **Delivery guarantee varies**: Push bisa gagal jika user uninstall app, SMS bisa gagal jika nomor tidak aktif.
- **No real-time SSE/WebSocket (MVP)**: In-app notification dibaca via polling. Real-time push ke web via SSE bisa ditambahkan.
- **CQRS complexity**: Tanpa Vernon auto-sync, harus manage event subscription dan worker lifecycle sendiri.

## Alternatives Considered

### 1. Vernon Pattern untuk notification logs
- Ditolak: notification logs bersifat append-only, write-heavy, dan tidak butuh relationship autoloading. CQRS lebih tepat.

### 2. Notification tanpa template (hardcoded messages)
- Ditolak: admin perlu bisa customize wording (bahasa, tone) tanpa deploy ulang.

### 3. Hanya push notification (tanpa SMS/WhatsApp)
- Ditolak: banyak orang tua di daerah tidak punya smartphone memadai. SMS adalah fallback penting.

### 4. Third-party notification service (OneSignal, Novu)
- Deferred: untuk MVP, in-house notification service lebih controllable. Bisa migrasi ke managed service jika complexity meningkat.

### 5. WhatsApp sebagai messaging utama (bukan channel notifikasi)
- Ditolak: WhatsApp API mahal dan terbatas. Lebih baik sebagai delivery channel untuk alert, bukan platform komunikasi utama.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | Template renderer replaces all placeholders | Template + data | Fully rendered string, no `{{...}}` remaining |
| U02 | Template renderer handles missing placeholder | Template with `{{unknown}}` | Empty string for missing, no crash |
| U03 | Idempotency key generator | event_type + event_id + channel + user_id | Deterministic unique key |
| U04 | Quiet hours check — inside quiet hours | 23:00, quiet=22:00-06:00 | `true` (should queue) |
| U05 | Quiet hours check — outside quiet hours | 10:00, quiet=22:00-06:00 | `false` (send now) |
| U06 | Recipient resolver — student absent | student_id | Returns parent user_ids |
| U07 | Channel filter — user disabled SMS | preferences with SMS disabled | SMS excluded from channels |
| U08 | Validate rejects invalid channel | `"telegram"` | Error: invalid channel |
| U09 | Validate rejects invalid status | `"cancelled"` | Error: invalid status |

### Integration Tests — Notification Delivery

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Event triggers notification | Emit AttendanceAbsent event | notification_log created for each parent |
| I02 | Multi-channel delivery | User has push + whatsapp enabled | 2 notification_log records created |
| I03 | Idempotency — no duplicate | Emit same event twice | Only 1 notification_log per channel+user |
| I04 | Quiet hours queueing | Emit event at 23:00 | notification_log status=pending, not sent until 06:00 |
| I05 | Failed notification retry | SMS send fails | retry_count incremented, re-queued |
| I06 | Max retry exceeded | 3 failures | status=failed, no more retries |

### Integration Tests — Notification Log

| # | Test Case | Action | Expected |
|---|---|---|---|
| I07 | Get my notifications | GET /notifications | 200, sorted by created_at DESC |
| I08 | Get unread count | GET /notifications/unread-count | 200, correct count |
| I09 | Mark as read | POST /notifications/{id}/read | 200, read_at set |
| I10 | Mark all as read | POST /notifications/read-all | 200, all read_at set |

### Integration Tests — Preferences

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Get default preferences | GET /notification-preferences (new user) | 200, all events enabled with push |
| I12 | Update preferences | PUT with disabled SMS for finance | 200, updated |
| I13 | Disabled event not sent | Disable attendance.absent, emit event | No notification_log created |

### Integration Tests — Templates

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | Create template | POST with event_type + channel + template | 201 |
| I15 | Preview template | POST /templates/{id}/preview with data | 200, rendered content |
| I16 | Unique constraint | Create duplicate event+channel+language | 409/422 |

### Integration Tests — Device Tokens

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Register device token | POST /device-tokens | 201 |
| I18 | Duplicate token rejected | POST same token twice | 409, unique constraint |
| I19 | Unregister token | DELETE /device-tokens/{id} | 200, is_active=false |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I20 | Cannot see other tenant's notifications | GET with wrong tenant | 404/empty |
| I21 | Event scoped to tenant | Emit event in tenant A | Only tenant A users notified |
