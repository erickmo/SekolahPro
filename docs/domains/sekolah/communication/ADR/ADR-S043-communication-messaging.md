# ADR-S043: Communication & Messaging (Komunikasi & Pesan)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Komunikasi antara sekolah dan orang tua di Indonesia saat ini bergantung pada grup WhatsApp yang **tidak terstruktur**, buku penghubung fisik, dan panggilan telepon. Masalah utama:

1. **Tidak terdokumentasi**: Pesan di WhatsApp hilang, tidak bisa di-audit, tidak tersimpan di sistem sekolah.
2. **Tidak ter-scope**: Grup kelas campur antara pengumuman resmi, diskusi pribadi, dan spam.
3. **Tidak bisa dilacak**: Sekolah tidak tahu apakah orang tua sudah membaca informasi penting.
4. **Privasi**: Guru terpaksa share nomor pribadi ke orang tua — tidak ada batas profesional.

Domain messaging mencakup 3 pola komunikasi:

| Pola | Contoh | Volume |
|------|--------|--------|
| **1-to-1** | Guru ↔ Orang tua (diskusi akademik anak) | Medium |
| **1-to-many (broadcast)** | Sekolah → Semua orang tua (pengumuman) | Low frequency, high reach |
| **Group (class-level)** | Wali kelas → Semua orang tua di kelas | Medium |

Messaging ini **berbeda** dari notification (S044) — messaging adalah percakapan dua arah, sementara notification adalah alert satu arah yang dipicu oleh event sistem.

### Mengapa Vernon Pattern?

- Conversations dan messages = entity yang perlu di-persist dan di-query.
- Read-heavy: orang tua dan guru membaca pesan lebih sering dari menulis.
- Relasi ke users (ADR-013), students (S001), class_rooms (ADR-011).
- Read receipt tracking memerlukan efficient read dari banyak participants.
- Thread/conversation model cocok untuk denormalisasi — preview pesan terakhir di list conversations.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `conversations` (thread percakapan), `conversation_participants` (peserta), dan `messages` (pesan individual).

### Table Schema

```sql
-- Thread percakapan
CREATE TABLE conversations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Conversation metadata
    conversation_type VARCHAR(20) NOT NULL,
    title           VARCHAR(255),
    context_type    VARCHAR(30),
    context_id      UUID,

    -- Last message preview (denormalized for list view)
    last_message_at     TIMESTAMPTZ,
    last_message_preview TEXT,
    last_sender_id      UUID,
    message_count       INT NOT NULL DEFAULT 0,

    -- Settings
    is_archived     BOOLEAN NOT NULL DEFAULT false,
    is_locked       BOOLEAN NOT NULL DEFAULT false,

    -- Created by
    created_by      UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_conv_type CHECK (conversation_type IN ('direct', 'class_group', 'broadcast', 'custom_group')),
    CONSTRAINT chk_context_type CHECK (context_type IS NULL OR context_type IN ('class_room', 'student', 'academic_year'))
);

-- Indexes
CREATE INDEX idx_conv_tenant_company ON conversations (tenant_id, company_id);
CREATE INDEX idx_conv_type ON conversations (conversation_type);
CREATE INDEX idx_conv_context ON conversations (context_type, context_id) WHERE context_type IS NOT NULL;
CREATE INDEX idx_conv_last_message ON conversations (last_message_at DESC);
CREATE INDEX idx_conv_created_by ON conversations (created_by);
CREATE INDEX idx_conv_rels ON conversations USING GIN (_rels);
CREATE INDEX idx_conv_data ON conversations USING GIN (_data);

-- Peserta percakapan
CREATE TABLE conversation_participants (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    conversation_id UUID NOT NULL,
    user_id         UUID NOT NULL,

    -- Participant metadata
    role            VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    left_at         TIMESTAMPTZ,

    -- Read tracking
    last_read_at    TIMESTAMPTZ,
    last_read_message_id UUID,
    unread_count    INT NOT NULL DEFAULT 0,

    -- Notification
    is_muted        BOOLEAN NOT NULL DEFAULT false,
    muted_until     TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_participant UNIQUE (conversation_id, user_id),
    CONSTRAINT chk_participant_role CHECK (role IN ('admin', 'member', 'readonly'))
);

-- Indexes
CREATE INDEX idx_participant_tenant_company ON conversation_participants (tenant_id, company_id);
CREATE INDEX idx_participant_conv ON conversation_participants (conversation_id);
CREATE INDEX idx_participant_user ON conversation_participants (user_id);
CREATE INDEX idx_participant_unread ON conversation_participants (user_id, unread_count) WHERE unread_count > 0;
CREATE INDEX idx_participant_rels ON conversation_participants USING GIN (_rels);
CREATE INDEX idx_participant_data ON conversation_participants USING GIN (_data);

-- Pesan individual
CREATE TABLE messages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    conversation_id UUID NOT NULL,
    sender_id       UUID NOT NULL,

    -- Message content
    message_type    VARCHAR(20) NOT NULL DEFAULT 'text',
    content         TEXT NOT NULL,
    metadata        JSONB NOT NULL DEFAULT '{}',

    -- Reply
    reply_to_id     UUID,

    -- Attachments
    attachments     JSONB NOT NULL DEFAULT '[]',

    -- Status
    is_edited       BOOLEAN NOT NULL DEFAULT false,
    edited_at       TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_message_type CHECK (message_type IN ('text', 'image', 'file', 'system'))
);

-- Indexes
CREATE INDEX idx_message_tenant_company ON messages (tenant_id, company_id);
CREATE INDEX idx_message_conv ON messages (conversation_id, created_at DESC);
CREATE INDEX idx_message_sender ON messages (sender_id);
CREATE INDEX idx_message_reply ON messages (reply_to_id) WHERE reply_to_id IS NOT NULL;
CREATE INDEX idx_message_type ON messages (message_type) WHERE message_type != 'text';
CREATE INDEX idx_message_rels ON messages USING GIN (_rels);
CREATE INDEX idx_message_data ON messages USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `conversation_type` | 4 tipe | direct (1-to-1), class_group (wali kelas → parents), broadcast (sekolah → all), custom_group |
| `context_type` + `context_id` | Polymorphic context | Conversation bisa di-scope ke class_room, student, atau academic_year |
| `last_message_preview` | Denormalized TEXT | Agar list conversations tidak perlu JOIN ke messages — mobile performance |
| `message_count` | Denormalized INT | Counter untuk display, diupdate via trigger/application |
| `unread_count` | Per participant | Setiap peserta punya hitungan unread sendiri |
| `last_read_message_id` | UUID per participant | Marker posisi baca terakhir — untuk read receipt |
| `attachments` | JSONB array | Array of `{filename, url, mime_type, size_bytes}` — max 5 files per message |
| `reply_to_id` | UUID nullable | Threading — memungkinkan reply ke pesan spesifik |
| `is_locked` | BOOLEAN | Admin bisa lock conversation (broadcast) agar hanya admin yang bisa kirim |
| `is_muted` | Per participant | Orang tua bisa mute conversation tanpa keluar |

### Vernon Relationships

**conversations:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `created_by_user` | belongs_to | **Ya** | Siapa yang memulai percakapan |

**conversation_participants:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `conversation` | belongs_to | **Ya** | Thread induk |
| `user` | belongs_to | **Ya** | Identitas peserta |

**messages:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `conversation` | belongs_to | **Tidak** | Sudah implicit dari endpoint URL |
| `sender` | belongs_to | **Ya** | Selalu perlu tahu pengirim |

### _rels / _data Structure

```json
// conversations
{
  "_rels": {
    "created_by": "018f..."
  },
  "_data": {
    "created_by_user": { "id": "018f...", "display_name": "Ibu Sari", "role": "teacher" },
    "participant_count": 32,
    "context": { "type": "class_room", "id": "018f...", "name": "VII-A" }
  }
}

// messages
{
  "_rels": {
    "conversation_id": "018f...",
    "sender_id": "018f..."
  },
  "_data": {
    "sender": { "id": "018f...", "display_name": "Budi Santoso", "role": "parent", "avatar_url": null }
  }
}
```

### Attachments Structure

```json
{
  "attachments": [
    {
      "filename": "surat-izin.pdf",
      "url": "https://storage.sekolahpro.id/files/018f.../surat-izin.pdf",
      "mime_type": "application/pdf",
      "size_bytes": 245760
    },
    {
      "filename": "foto-sakit.jpg",
      "url": "https://storage.sekolahpro.id/files/018f.../foto-sakit.jpg",
      "mime_type": "image/jpeg",
      "size_bytes": 1048576
    }
  ]
}
```

### API Endpoints

```
# Conversations
GET    /api/v1/conversations                           — List percakapan user (paginated, sorted by last_message_at)
POST   /api/v1/conversations                           — Buat percakapan baru
GET    /api/v1/conversations/{id}                      — Detail percakapan
PUT    /api/v1/conversations/{id}                      — Update (title, lock, archive)
DELETE /api/v1/conversations/{id}                      — Soft delete (admin only)

# Participants
GET    /api/v1/conversations/{id}/participants         — List peserta
POST   /api/v1/conversations/{id}/participants         — Tambah peserta (group only)
DELETE /api/v1/conversations/{id}/participants/{uid}   — Remove peserta
PUT    /api/v1/conversations/{id}/participants/me      — Update my settings (mute, etc.)

# Messages
GET    /api/v1/conversations/{id}/messages             — List pesan (paginated, cursor-based)
POST   /api/v1/conversations/{id}/messages             — Kirim pesan
PUT    /api/v1/messages/{id}                           — Edit pesan (within 15 min)
DELETE /api/v1/messages/{id}                           — Soft delete pesan

# Read receipts
POST   /api/v1/conversations/{id}/read                — Mark as read (up to message_id)
GET    /api/v1/conversations/unread-count              — Total unread across all conversations

# Class group auto-create
POST   /api/v1/conversations/class-group               — Auto-create group for class (wali kelas)

# File upload
POST   /api/v1/messages/upload                         — Upload attachment (returns URL)
```

### Class Group Auto-Creation Flow

```
1. Wali kelas pilih: "Buat grup kelas VII-A"
2. System:
   a. Query S014 class_placement → get all students in class
   b. Query S003 guardians → get all parent user_ids for those students
   c. Create conversation with type='class_group', context_type='class_room'
   d. Add wali kelas as participant (role='admin')
   e. Add all parent users as participants (role='member')
3. Ketika siswa baru masuk kelas → auto-add parent ke group
4. Ketika siswa pindah → auto-remove parent dari group
```

## Consequences

### Positive

- **Terstruktur**: Percakapan di-scope ke konteks (kelas, siswa) — tidak campur aduk seperti grup WhatsApp.
- **Terdokumentasi**: Semua komunikasi tersimpan dan bisa di-audit oleh admin sekolah.
- **Read receipt**: Sekolah tahu apakah orang tua sudah membaca informasi penting.
- **Privacy**: Guru dan orang tua berkomunikasi melalui sistem — nomor pribadi tidak perlu dishare.
- **Auto-group**: Grup kelas otomatis dibuat berdasarkan data placement — tidak perlu manual invite.
- **Mobile-optimized**: Cursor-based pagination dan denormalized preview untuk performa mobile.

### Negative / Trade-offs

- **Bukan real-time (MVP)**: MVP menggunakan polling atau long-polling. WebSocket bisa ditambahkan untuk real-time.
- **Storage attachment**: File attachments memerlukan object storage (S3/MinIO) — biaya tambahan.
- **Moderation**: Butuh mekanisme moderasi untuk mencegah spam/konten tidak pantas — deferred.
- **Not replacing WhatsApp**: Banyak orang tua tetap prefer WhatsApp — adoption bisa lambat. Mitigasi: kirim notification ke WhatsApp (S044) yang link ke portal.
- **Message volume**: 30 parents × 10 messages/day × 200 days = 60.000 messages per class per year.

## Alternatives Considered

### 1. Integrasi langsung ke WhatsApp Business API
- Ditolak untuk core: WhatsApp API mahal per message dan tergantung Meta. Messaging in-app lebih sustainable. WhatsApp digunakan sebagai notification channel (S044), bukan messaging channel.

### 2. Tanpa conversation model (flat messages)
- Ditolak: tanpa grouping, messages sulit di-organize. Conversation model memberikan context dan scope.

### 3. Forum/bulletin board model
- Ditolak: tidak cocok untuk komunikasi cepat guru-orang tua. Forum lebih cocok untuk diskusi panjang.

### 4. Email-only communication
- Ditolak: banyak orang tua di Indonesia tidak aktif cek email. In-app messaging + push notification lebih efektif.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `ConversationDescriptor.TableName()` | — | `"conversations"` |
| U02 | `MessageDescriptor.TableName()` | — | `"messages"` |
| U03 | `ParticipantDescriptor.TableName()` | — | `"conversation_participants"` |
| U04 | Validate rejects invalid `conversation_type` | `"channel"` | Error: invalid type |
| U05 | Validate rejects invalid `message_type` | `"video"` | Error: invalid type |
| U06 | Validate rejects invalid participant `role` | `"moderator"` | Error: invalid role |
| U07 | Validate rejects empty message content | `""` | Error: content required |
| U08 | Attachment validator rejects >5 files | 6 attachments | Error: max 5 attachments |
| U09 | Attachment validator rejects >10MB file | `size_bytes: 11000000` | Error: file too large |

### Integration Tests — Conversations

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create direct conversation | POST with type=direct, 2 participants | 201 |
| I02 | Create class group | POST /conversations/class-group with class_room_id | 201, all parents added |
| I03 | List conversations | GET /conversations | 200, sorted by last_message_at |
| I04 | Duplicate direct conversation | Create 2 directs between same users | Returns existing conversation |
| I05 | Archive conversation | PUT with is_archived=true | 200, hidden from default list |

### Integration Tests — Messages

| # | Test Case | Action | Expected |
|---|---|---|---|
| I06 | Send text message | POST message with content | 201, conversation.last_message_* updated |
| I07 | Send with attachment | POST message with file | 201, attachments array populated |
| I08 | Reply to message | POST with reply_to_id | 201, reply_to linked |
| I09 | Edit message within 15min | PUT message within window | 200, is_edited=true |
| I10 | Edit message after 15min | PUT message after window | 403, edit window expired |
| I11 | Cursor-based pagination | GET messages with cursor | 200, correct page |
| I12 | Send to locked conversation (non-admin) | POST message as member to locked conv | 403, conversation locked |

### Integration Tests — Read Receipts

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | Mark as read | POST /conversations/{id}/read | 200, unread_count = 0, last_read updated |
| I14 | Unread count | GET /conversations/unread-count | 200, total across conversations |
| I15 | New message increments unread | Send message to conversation | All other participants unread_count += 1 |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | UserUpdated syncs to messages | Update user display_name | `_data.sender.display_name` updated |
| I17 | UserUpdated syncs to participants | Update user display_name | `_data.user.display_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I18 | Cannot access other tenant's conversations | GET with wrong tenant | 404 |
| I19 | Cannot send message cross-tenant | POST message to other tenant's conv | 403 |
| I20 | Cannot add participant cross-company | Add user from other company | 422 |
