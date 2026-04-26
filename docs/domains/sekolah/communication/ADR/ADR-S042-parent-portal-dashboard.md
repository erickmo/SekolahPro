# ADR-S042: Parent Portal & Dashboard (Portal Orang Tua)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Portal orang tua adalah **touchpoint utama** antara sekolah dan keluarga siswa. Di Indonesia, komunikasi sekolah-orang tua masih didominasi oleh grup WhatsApp yang tidak terstruktur, buku penghubung fisik, dan panggilan telepon. Portal digital memberikan akses terstruktur dan real-time ke data anak.

Kebutuhan portal orang tua:

1. **Dashboard ringkasan**: Overview akademik, kehadiran, keuangan per anak — informasi yang paling sering dicek orang tua.
2. **Multi-child support**: Satu orang tua bisa punya 2-5 anak di sekolah yang sama (kakak-adik). Portal harus bisa switch antar anak.
3. **Mobile-first**: 85%+ orang tua di Indonesia mengakses via smartphone — desain harus responsive/mobile-first.
4. **Read-only views**: Portal orang tua hanya membaca data dari domain lain (S001-S018), tidak menulis data akademik.
5. **Portal preferences**: Orang tua bisa set preferensi bahasa, notifikasi, child default view.
6. **Pesantren variant**: Portal pesantren menampilkan data tambahan — hafalan, diniyah grades, kitab progress (ADR-009).

Data yang ditampilkan berasal dari domain yang sudah ada:

| Sumber | Data di Portal |
|--------|----------------|
| S001 Student | Biodata anak |
| S003 Guardian | Relasi orang tua-anak |
| S004 Academic Record | Nilai semester, ranking |
| S008 Attendance | Kehadiran hari ini & rekap |
| S009 Finance/SPP | Status pembayaran, tunggakan |
| S011 Subject Grades | Nilai per mata pelajaran |
| S012 Discipline | Catatan perilaku |
| S015 Extracurricular | Aktivitas ekskul |
| S018 Rapor | Download rapor PDF |

### Mengapa Vernon Pattern?

- Portal preferences/settings = entity milik orang tua yang perlu di-persist.
- Read-heavy: orang tua cek dashboard setiap hari.
- Relasi ke user (ADR-013) dan multiple students.
- Dashboard widget config bersifat personalisasi per parent.
- Data portal sendiri ringan — berat di read aggregation dari domain lain.

## Decision

Menggunakan **Vernon Pattern** untuk `parent_portal_settings` (preferensi portal per orang tua). Data yang ditampilkan di portal **tidak disimpan ulang** — dibaca langsung dari domain source (S001-S018) via API aggregation.

### Table Schema

```sql
-- Preferensi dan pengaturan portal orang tua
CREATE TABLE parent_portal_settings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Parent identity
    user_id         UUID NOT NULL,
    guardian_id     UUID NOT NULL,

    -- Preferences
    language        VARCHAR(5) NOT NULL DEFAULT 'id',
    default_child_id UUID,
    dashboard_layout JSONB NOT NULL DEFAULT '{}',
    notification_prefs JSONB NOT NULL DEFAULT '{}',

    -- Theme
    theme           VARCHAR(20) NOT NULL DEFAULT 'light',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_portal_user UNIQUE (tenant_id, company_id, user_id),
    CONSTRAINT chk_portal_language CHECK (language IN ('id', 'en', 'ar')),
    CONSTRAINT chk_portal_theme CHECK (theme IN ('light', 'dark', 'auto'))
);

-- Indexes
CREATE INDEX idx_portal_settings_tenant_company ON parent_portal_settings (tenant_id, company_id);
CREATE INDEX idx_portal_settings_user ON parent_portal_settings (user_id);
CREATE INDEX idx_portal_settings_guardian ON parent_portal_settings (guardian_id);
CREATE INDEX idx_portal_settings_rels ON parent_portal_settings USING GIN (_rels);
CREATE INDEX idx_portal_settings_data ON parent_portal_settings USING GIN (_data);

-- Parent-child mapping view (convenience, source of truth is S003 guardians)
CREATE VIEW parent_children_view AS
SELECT
    g.id AS guardian_id,
    g.user_id,
    g.tenant_id,
    g.company_id,
    s.id AS student_id,
    s.full_name AS student_name,
    s.nis,
    s.status AS student_status,
    g.relationship_type
FROM student_guardians g
JOIN students s ON s.id = g.student_id AND s.deleted_at IS NULL
WHERE g.deleted_at IS NULL;
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `user_id` | UUID, NOT NULL, UNIQUE per tenant+company | Satu user = satu set preferensi portal |
| `guardian_id` | UUID, NOT NULL | Link ke S003 guardian — sumber relasi parent-child |
| `language` | VARCHAR(5), CHECK | Indonesia (id), English (en), Arabic (ar) untuk pesantren |
| `default_child_id` | UUID, nullable | Anak yang ditampilkan pertama saat buka portal (null = anak pertama) |
| `dashboard_layout` | JSONB | Konfigurasi widget mana yang tampil dan urutannya |
| `notification_prefs` | JSONB | Channel preference per event type (lihat S044) |
| `theme` | VARCHAR(20) | Light/dark/auto — mobile UX preference |

### Vernon Relationships

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `user` | belongs_to | **Ya** | Identitas login orang tua |
| `guardian` | belongs_to | **Ya** | Data guardian dari S003 |
| `default_child` | belongs_to | **Tidak** | Opsional, hanya jika di-set |

### _rels / _data Structure

```json
{
  "_rels": {
    "user_id": "018f...",
    "guardian_id": "018f...",
    "default_child_id": "018f..."
  },
  "_data": {
    "user": { "id": "018f...", "email": "budi@email.com", "display_name": "Budi Santoso" },
    "guardian": {
      "id": "018f...",
      "full_name": "Budi Santoso",
      "relationship_type": "father",
      "children": [
        { "student_id": "018f...", "full_name": "Ahmad Fadhil", "nis": "12345", "class": "VII-A" },
        { "student_id": "018f...", "full_name": "Siti Aisyah", "nis": "12346", "class": "V-B" }
      ]
    }
  }
}
```

### Dashboard Aggregation Endpoints

Portal dashboard **tidak menyimpan data sendiri** — ia membaca dari domain lain. Berikut endpoint aggregasi khusus portal:

```json
{
  "dashboard_layout": {
    "widgets": ["attendance_today", "grades_summary", "finance_status", "announcements"],
    "widget_order": ["attendance_today", "finance_status", "grades_summary", "announcements"],
    "collapsed": ["announcements"]
  },
  "notification_prefs": {
    "attendance_absent": ["push", "whatsapp"],
    "payment_due": ["push", "sms"],
    "grade_published": ["push"],
    "announcement": ["push"]
  }
}
```

### API Endpoints

```
# Portal Settings
GET    /api/v1/parent-portal/settings                — Get preferensi portal
PUT    /api/v1/parent-portal/settings                — Update preferensi
POST   /api/v1/parent-portal/settings/reset          — Reset ke default

# Dashboard (aggregation — read-only)
GET    /api/v1/parent-portal/dashboard               — Dashboard overview (all children summary)
GET    /api/v1/parent-portal/children                 — List semua anak
GET    /api/v1/parent-portal/children/{id}/summary    — Summary per anak

# Child Data (proxy read-only dari domain lain)
GET    /api/v1/parent-portal/children/{id}/attendance         — Kehadiran (dari S008)
GET    /api/v1/parent-portal/children/{id}/attendance/today   — Status hari ini
GET    /api/v1/parent-portal/children/{id}/grades             — Nilai (dari S011)
GET    /api/v1/parent-portal/children/{id}/grades/current     — Nilai semester berjalan
GET    /api/v1/parent-portal/children/{id}/finance            — Tagihan & pembayaran (dari S009)
GET    /api/v1/parent-portal/children/{id}/finance/outstanding — Tunggakan
GET    /api/v1/parent-portal/children/{id}/discipline         — Catatan perilaku (dari S012)
GET    /api/v1/parent-portal/children/{id}/extracurricular    — Ekskul (dari S015)
GET    /api/v1/parent-portal/children/{id}/rapor              — Rapor (dari S018)
GET    /api/v1/parent-portal/children/{id}/rapor/{rapor_id}/pdf — Download rapor PDF

# Pesantren variant (ADR-009 dual-mode)
GET    /api/v1/parent-portal/children/{id}/hafalan            — Progress hafalan
GET    /api/v1/parent-portal/children/{id}/diniyah-grades     — Nilai diniyah
```

### Authorization Model

Portal orang tua menggunakan role `parent` dari ADR-013. Akses dibatasi:

1. **Scope**: Orang tua hanya bisa akses data anak yang ter-link di S003 (guardian).
2. **Read-only**: Tidak bisa mengubah data akademik, keuangan, atau kehadiran.
3. **Write**: Hanya bisa ubah preferensi portal sendiri dan mengirim pesan (S043).
4. **Multi-child**: Semua anak yang ter-link otomatis muncul — tidak perlu setup manual.

```
Authorization Flow:
1. User login → role = 'parent'
2. System query S003: guardian WHERE user_id = current_user
3. System query students linked to guardian
4. All portal endpoints scoped to: student_ids IN (linked children)
5. Any attempt to access non-linked student → 403 Forbidden
```

## Consequences

### Positive

- **Single source of truth**: Portal tidak duplikasi data — membaca langsung dari domain source.
- **Multi-child native**: Desain mendukung banyak anak secara bawaan via S003 guardian linkage.
- **Mobile-first**: API dirancang untuk mobile consumption — summary endpoints, minimal payload.
- **Personalisasi**: Widget layout dan notification preferences disimpan per orang tua.
- **Pesantren ready**: Endpoint hafalan dan diniyah tersedia untuk dual-mode institution (ADR-009).
- **Security by design**: Akses di-scope ke linked children, tidak bisa akses anak orang lain.

### Negative / Trade-offs

- **N+1 aggregation risk**: Dashboard overview harus query 5+ domain per anak. Mitigasi: parallel query + caching.
- **Eventual consistency**: Data yang ditampilkan di portal bisa sedikit stale (detik) karena read from Vernon cache.
- **No offline**: Portal web tidak mendukung offline — butuh PWA atau native app untuk itu.
- **Guardian link dependency**: Jika S003 guardian belum di-link ke user, orang tua tidak bisa akses portal. Perlu onboarding flow.

## Alternatives Considered

### 1. Dedicated parent_dashboard table (materialized view)
- Ditolak: menduplikasi data dari 8+ domain — sync nightmare. Lebih baik aggregation saat read.

### 2. GraphQL untuk portal
- Ditolak untuk MVP: REST aggregation endpoints lebih sederhana. GraphQL bisa ditambahkan sebagai enhancement jika query patterns menjadi terlalu beragam.

### 3. Separate mobile app
- Deferred: MVP menggunakan responsive web. Native app bisa ditambahkan nanti dengan API yang sama.

### 4. Portal tanpa preferences table
- Ditolak: tanpa preferences, tidak bisa personalisasi layout, language, notification — UX buruk untuk daily usage.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` returns correct name | — | `"parent_portal_settings"` |
| U02 | `DefaultRels()` returns 2 autoloaded rels | — | `user`, `guardian` |
| U03 | Validate rejects invalid `language` | `"jp"` | Error: language must be id/en/ar |
| U04 | Validate rejects invalid `theme` | `"blue"` | Error: theme must be light/dark/auto |
| U05 | Validate accepts valid settings | All fields valid | No error |
| U06 | Dashboard layout parser handles empty config | `{}` | Returns default widget order |
| U07 | Notification prefs parser handles unknown event | `{"unknown_event": ["push"]}` | Ignores unknown, no error |

### Integration Tests — Portal Settings

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Get default settings | GET /parent-portal/settings (new user) | 200, default language=id, theme=light |
| I02 | Update preferences | PUT /parent-portal/settings with layout | 200, updated |
| I03 | Reset to default | POST /parent-portal/settings/reset | 200, reverted to defaults |
| I04 | Unique per user per tenant | Create 2 settings for same user | 409/422, unique constraint |

### Integration Tests — Dashboard Aggregation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I05 | Get children list | GET /parent-portal/children | 200, returns all linked children |
| I06 | Get child summary | GET /parent-portal/children/{id}/summary | 200, attendance + grades + finance summary |
| I07 | Get dashboard overview | GET /parent-portal/dashboard | 200, all children summaries |
| I08 | Attendance today | GET /parent-portal/children/{id}/attendance/today | 200, today's status or null if weekend |
| I09 | Finance outstanding | GET /parent-portal/children/{id}/finance/outstanding | 200, list of unpaid invoices |
| I10 | Download rapor PDF | GET /parent-portal/children/{id}/rapor/{rapor_id}/pdf | 200, PDF stream |

### Integration Tests — Authorization

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Cannot access unlinked child | GET /parent-portal/children/{other_child_id}/summary | 403, forbidden |
| I12 | Non-parent role rejected | GET /parent-portal/dashboard as teacher role | 403, forbidden |
| I13 | Parent with no linked children | GET /parent-portal/children (guardian not linked) | 200, empty array |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | UserUpdated syncs to portal settings | Update user display_name | `_data.user.display_name` updated |
| I15 | GuardianUpdated syncs children list | Add new child to guardian | `_data.guardian.children` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Cannot access other tenant's portal | GET with wrong tenant scope | 404 |
| I17 | Cannot see children from other company | GET /parent-portal/children with cross-company child | Empty, filtered out |
