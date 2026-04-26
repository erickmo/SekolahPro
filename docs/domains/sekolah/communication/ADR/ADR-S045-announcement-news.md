# ADR-S045: Announcement & News (Pengumuman & Berita Sekolah)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Pengumuman sekolah adalah komunikasi resmi satu arah dari sekolah ke stakeholder (orang tua, guru, siswa). Saat ini pengumuman disebarkan melalui:

1. **Mading fisik**: Papan pengumuman di sekolah — hanya bisa dibaca di lokasi.
2. **Grup WhatsApp**: Tercampur dengan chat pribadi, sulit dicari ulang.
3. **Surat edaran**: Dicetak dan dibagikan ke siswa untuk dibawa pulang — sering hilang.
4. **Website sekolah**: Jarang diupdate, tidak ada targeting.

Kebutuhan pengumuman digital:

1. **Target audience**: Pengumuman bisa ditargetkan ke semua, guru saja, orang tua saja, siswa tertentu, atau kelas tertentu.
2. **Priority levels**: Pengumuman darurat (penutupan sekolah mendadak) vs informasi biasa.
3. **Expiry date**: Pengumuman kegiatan harus hilang setelah acara selesai.
4. **Attachments**: Surat edaran, poster, jadwal kegiatan dalam bentuk file.
5. **Read tracking**: Sekolah perlu tahu berapa persen target yang sudah membaca.
6. **Pesantren variant**: Pengumuman kegiatan keagamaan, jadwal Ramadan, pengajian (ADR-009).

Pengumuman **berbeda** dari messaging (S043) — pengumuman bersifat satu arah dan resmi, sedangkan messaging adalah percakapan dua arah. Pengumuman juga **berbeda** dari notifikasi (S044) — pengumuman adalah konten itu sendiri, sedangkan notifikasi adalah alert bahwa ada konten baru.

### Mengapa Vernon Pattern?

- Announcement = entity yang perlu di-persist, di-query, dan di-filter.
- Read-heavy: pengumuman dibaca banyak orang berkali-kali.
- Relasi ke creator (user), target audience (class, role).
- Denormalized read count dan acknowledgement tracking.
- Business logic sederhana (CRUD + publish + expire).

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `announcements` (pengumuman) dan `announcement_reads` (tracking pembaca).

### Table Schema

```sql
-- Pengumuman
CREATE TABLE announcements (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Content
    title           VARCHAR(255) NOT NULL,
    content         TEXT NOT NULL,
    content_html    TEXT,
    excerpt         VARCHAR(500),

    -- Targeting
    target_audience VARCHAR(20) NOT NULL,
    target_roles    JSONB NOT NULL DEFAULT '[]',
    target_class_ids JSONB NOT NULL DEFAULT '[]',
    target_grade_levels JSONB NOT NULL DEFAULT '[]',

    -- Priority & scheduling
    priority        VARCHAR(10) NOT NULL DEFAULT 'normal',
    publish_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expire_at       TIMESTAMPTZ,
    is_pinned       BOOLEAN NOT NULL DEFAULT false,

    -- Attachments
    attachments     JSONB NOT NULL DEFAULT '[]',
    cover_image_url TEXT,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',

    -- Author
    created_by      UUID NOT NULL,
    published_by    UUID,

    -- Read tracking (denormalized counters)
    total_target    INT NOT NULL DEFAULT 0,
    total_read      INT NOT NULL DEFAULT 0,
    read_percentage NUMERIC(5,2) NOT NULL DEFAULT 0.00,

    -- Category
    category        VARCHAR(30) NOT NULL DEFAULT 'general',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_ann_audience CHECK (target_audience IN ('all', 'teachers', 'parents', 'students', 'specific_classes', 'specific_roles')),
    CONSTRAINT chk_ann_priority CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    CONSTRAINT chk_ann_status CHECK (status IN ('draft', 'published', 'expired', 'archived')),
    CONSTRAINT chk_ann_category CHECK (category IN ('general', 'academic', 'finance', 'event', 'health', 'emergency', 'religious'))
);

-- Indexes
CREATE INDEX idx_ann_tenant_company ON announcements (tenant_id, company_id);
CREATE INDEX idx_ann_status_publish ON announcements (status, publish_at DESC) WHERE status = 'published';
CREATE INDEX idx_ann_audience ON announcements (target_audience);
CREATE INDEX idx_ann_priority ON announcements (priority) WHERE priority IN ('high', 'urgent');
CREATE INDEX idx_ann_pinned ON announcements (is_pinned, publish_at DESC) WHERE is_pinned = true;
CREATE INDEX idx_ann_expire ON announcements (expire_at) WHERE expire_at IS NOT NULL AND status = 'published';
CREATE INDEX idx_ann_category ON announcements (category);
CREATE INDEX idx_ann_created_by ON announcements (created_by);
CREATE INDEX idx_ann_rels ON announcements USING GIN (_rels);
CREATE INDEX idx_ann_data ON announcements USING GIN (_data);

-- Tracking per pembaca
CREATE TABLE announcement_reads (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    announcement_id UUID NOT NULL,
    user_id         UUID NOT NULL,

    -- Read tracking
    read_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    acknowledged    BOOLEAN NOT NULL DEFAULT false,
    acknowledged_at TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_ann_read UNIQUE (announcement_id, user_id)
);

-- Indexes
CREATE INDEX idx_ann_read_tenant_company ON announcement_reads (tenant_id, company_id);
CREATE INDEX idx_ann_read_announcement ON announcement_reads (announcement_id);
CREATE INDEX idx_ann_read_user ON announcement_reads (user_id);
CREATE INDEX idx_ann_read_rels ON announcement_reads USING GIN (_rels);
CREATE INDEX idx_ann_read_data ON announcement_reads USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `target_audience` | VARCHAR(20), CHECK | 6 target audience types — dari semua sampai kelas spesifik |
| `target_roles` | JSONB array | `["teacher", "parent"]` — filter tambahan jika audience = specific_roles |
| `target_class_ids` | JSONB array | UUID array kelas target jika audience = specific_classes |
| `target_grade_levels` | JSONB array | `["7", "8"]` — target per jenjang kelas |
| `priority` | 4 level | urgent = sekolah tutup darurat, high = deadline penting, normal/low = info biasa |
| `publish_at` | TIMESTAMPTZ | Scheduled publishing — draft bisa di-set untuk publish otomatis |
| `expire_at` | TIMESTAMPTZ nullable | Pengumuman kegiatan hilang setelah acara selesai |
| `is_pinned` | BOOLEAN | Pengumuman pinned selalu muncul di atas |
| `content_html` | TEXT nullable | Rich content (HTML sanitized) — untuk pengumuman dengan formatting |
| `excerpt` | VARCHAR(500) | Preview text untuk list view |
| `total_target` | Denormalized INT | Dihitung saat publish — berapa user yang menjadi target |
| `read_percentage` | NUMERIC(5,2) | Dihitung async — `total_read / total_target * 100` |
| `category` | 7 kategori | Termasuk `religious` untuk konteks pesantren (ADR-009) |
| `acknowledged` | BOOLEAN | Untuk pengumuman urgent yang memerlukan konfirmasi baca |

### Vernon Relationships

**announcements:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `created_by_user` | belongs_to | **Ya** | Siapa yang membuat pengumuman |
| `published_by_user` | belongs_to | **Tidak** | Opsional — hanya jika sudah published |

**announcement_reads:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `announcement` | belongs_to | **Ya** | Induk pengumuman |
| `user` | belongs_to | **Ya** | Siapa yang membaca |

### _rels / _data Structure

```json
// announcements
{
  "_rels": {
    "created_by": "018f...",
    "published_by": "018f..."
  },
  "_data": {
    "created_by_user": { "id": "018f...", "display_name": "Admin Sekolah", "role": "admin" },
    "target_summary": "Semua orang tua kelas VII",
    "target_classes": [
      { "id": "018f...", "name": "VII-A" },
      { "id": "018f...", "name": "VII-B" }
    ]
  }
}

// announcement_reads
{
  "_rels": {
    "announcement_id": "018f...",
    "user_id": "018f..."
  },
  "_data": {
    "announcement": { "id": "018f...", "title": "Libur Maulid Nabi" },
    "user": { "id": "018f...", "display_name": "Budi Santoso", "role": "parent" }
  }
}
```

### API Endpoints

```
# Announcements (admin/teacher)
GET    /api/v1/announcements                           — List pengumuman (admin view, all statuses)
POST   /api/v1/announcements                           — Buat pengumuman (draft)
GET    /api/v1/announcements/{id}                      — Detail pengumuman
PUT    /api/v1/announcements/{id}                      — Update (draft only)
DELETE /api/v1/announcements/{id}                      — Soft delete
POST   /api/v1/announcements/{id}/publish              — Publish pengumuman
POST   /api/v1/announcements/{id}/archive              — Archive pengumuman

# Announcements (public — all roles)
GET    /api/v1/announcements/feed                      — Feed pengumuman untuk user (filtered by role/class)
GET    /api/v1/announcements/{id}/detail               — Detail pengumuman (public)

# Read tracking
POST   /api/v1/announcements/{id}/read                 — Mark as read
POST   /api/v1/announcements/{id}/acknowledge          — Acknowledge (untuk urgent)
GET    /api/v1/announcements/{id}/reads                — Read statistics (admin only)
GET    /api/v1/announcements/{id}/reads/unread         — List yang belum baca (admin only)

# Expiry cron (internal)
POST   /api/v1/announcements/expire                    — Expire past-due announcements (cron job)
```

### Publishing Flow

```
1. Admin/guru buat pengumuman (status=draft)
2. Set: target audience, priority, publish_at, expire_at
3. POST /announcements/{id}/publish:
   a. Validate: target audience not empty
   b. Calculate total_target (count users in audience)
   c. Set status = 'published'
   d. Emit event: AnnouncementPublished { id, target_audience }
   e. Notification system (S044) picks up event → send push/whatsapp to targets
4. Users open portal/app → GET /announcements/feed
5. User reads → POST /announcements/{id}/read → total_read++, read_percentage recalculated
6. If urgent: user must POST /announcements/{id}/acknowledge
7. expire_at reached → cron sets status = 'expired'
```

## Consequences

### Positive

- **Targeted**: Pengumuman bisa ditargetkan ke audience spesifik — tidak ada spam ke pihak yang tidak relevan.
- **Trackable**: Admin tahu berapa persen yang sudah membaca — follow-up bisa ditargetkan ke yang belum baca.
- **Scheduled**: Pengumuman bisa dijadwalkan untuk publish dan expire otomatis.
- **Prioritized**: Pengumuman urgent (darurat) dibedakan secara visual dan memerlukan acknowledgement.
- **Integrated**: Publish otomatis memicu notifikasi (S044) ke channel yang dipilih orang tua.
- **Pesantren ready**: Kategori `religious` untuk pengumuman kegiatan keagamaan.

### Negative / Trade-offs

- **Read tracking overhead**: Untuk pengumuman ke 500 orang tua = 500 rows di announcement_reads. Volume tinggi.
- **Target calculation**: Menghitung total_target memerlukan query users berdasarkan role/class — bisa lambat untuk sekolah besar.
- **Expiry cron**: Memerlukan cron job untuk update status expired — bisa delay beberapa menit.
- **Rich content**: HTML content memerlukan sanitization untuk mencegah XSS — perlu library khusus.
- **No comment/reaction**: Pengumuman bersifat satu arah — jika orang tua ingin bertanya, harus melalui messaging (S043).

## Alternatives Considered

### 1. Pengumuman sebagai bagian dari messaging (S043)
- Ditolak: pengumuman bersifat resmi dan satu arah — berbeda dari percakapan. Mencampurkan keduanya mengaburkan boundary.

### 2. Pengumuman tanpa read tracking
- Ditolak: read tracking adalah requirement utama — sekolah perlu tahu apakah informasi sudah sampai ke orang tua.

### 3. Email blast saja
- Ditolak: banyak orang tua tidak aktif cek email. In-app + push notification lebih efektif.

### 4. Tanpa expiry — manual archive saja
- Ditolak: pengumuman kegiatan yang sudah lewat membingungkan jika masih muncul. Auto-expire lebih clean.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `AnnouncementDescriptor.TableName()` | — | `"announcements"` |
| U02 | `ReadDescriptor.TableName()` | — | `"announcement_reads"` |
| U03 | Validate rejects invalid `target_audience` | `"everyone"` | Error: invalid audience |
| U04 | Validate rejects invalid `priority` | `"critical"` | Error: invalid priority |
| U05 | Validate rejects invalid `status` | `"deleted"` | Error: invalid status |
| U06 | Validate rejects invalid `category` | `"sports"` | Error: invalid category |
| U07 | Validate rejects empty title | `""` | Error: title required |
| U08 | Validate rejects expire_at before publish_at | `expire_at < publish_at` | Error: expire must be after publish |
| U09 | Read percentage calculation | 150 read / 300 target | 50.00% |

### Integration Tests — CRUD

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create draft announcement | POST with title + content + audience | 201, status=draft |
| I02 | Update draft | PUT with new content | 200, updated |
| I03 | Cannot update published | PUT on published announcement | 403, immutable after publish |
| I04 | List admin view | GET /announcements (as admin) | 200, all statuses |
| I05 | List feed view | GET /announcements/feed (as parent) | 200, only published, filtered by role |

### Integration Tests — Publishing

| # | Test Case | Action | Expected |
|---|---|---|---|
| I06 | Publish announcement | POST /announcements/{id}/publish | 200, status=published, total_target calculated |
| I07 | Publish with schedule | POST with future publish_at | 200, not visible until publish_at |
| I08 | Publish emits event | Publish announcement | AnnouncementPublished event emitted |
| I09 | Target calculation — all parents | audience=parents | total_target = count of parent users |
| I10 | Target calculation — specific classes | audience=specific_classes + class_ids | total_target = count of parents in those classes |

### Integration Tests — Read Tracking

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Mark as read | POST /announcements/{id}/read | 201, read record created |
| I12 | Duplicate read ignored | POST read twice | No duplicate, same read_at |
| I13 | Read count updated | Mark as read | total_read++, read_percentage recalculated |
| I14 | Acknowledge urgent | POST /announcements/{id}/acknowledge | 200, acknowledged=true |
| I15 | Read statistics | GET /announcements/{id}/reads | 200, list with read/unread breakdown |

### Integration Tests — Expiry

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Auto-expire | Run expire cron with past-due announcements | Status changed to expired |
| I17 | Expired not in feed | GET /announcements/feed | Expired announcements excluded |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I18 | UserUpdated syncs to announcements | Update user display_name | `_data.created_by_user.display_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I19 | Cannot see other tenant's announcements | GET /announcements/feed with wrong tenant | 404/empty |
| I20 | Cannot read other tenant's announcement | POST read on other tenant's announcement | 403 |
