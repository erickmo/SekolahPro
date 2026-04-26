# ADR-S036: Canteen Management / Manajemen Kantin

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Manajemen kantin adalah kebutuhan operasional penting terutama untuk **boarding school dan pesantren** (ADR-009 islamic mode) di mana santri makan 3 kali sehari di pesantren. Untuk sekolah umum (day school), kantin bersifat opsional tetapi tetap dibutuhkan untuk pengelolaan vendor dan menu.

Data kantin diperlukan untuk:

1. **Menu harian**: Perencanaan menu makan pagi, siang, malam, dan snack.
2. **Jadwal makan**: Pengaturan waktu dan giliran makan per gedung/kelas.
3. **Nutrisi**: Informasi gizi dasar untuk transparansi ke orang tua (opsional).
4. **Vendor/supplier**: Pengelolaan penyedia bahan makanan atau catering.
5. **Perencanaan**: Estimasi kebutuhan bahan berdasarkan jumlah santri aktif.
6. **Laporan**: Rekap menu bulanan untuk orang tua dan dinas pendidikan.

Konteks pesantren/boarding school Indonesia:
- Santri makan 3x sehari + snack sore — total 4 meal per hari.
- Menu direncanakan **mingguan** dan berulang per cycle (biasanya 2-4 minggu).
- Beberapa pesantren mengelola dapur sendiri, yang lain menggunakan catering eksternal.
- Orang tua sering bertanya "makan apa hari ini?" — transparansi menu penting.
- Pesantren besar bisa melayani **500-2000 santri** per meal — logistik signifikan.

### Mengapa Vernon Pattern?

- Master data (menu, vendor) = read-heavy, jarang berubah.
- Menu schedule = reference data, diakses sering oleh orang tua via dashboard.
- Relasi ke vendor, academic_year.
- Business logic ringan: rotation schedule, nutritional tracking.
- Eventually consistent acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `canteen_vendors` (master vendor/supplier), `canteen_menus` (master item menu), dan `canteen_meal_schedules` (jadwal menu harian).

### Table Schema

```sql
-- Master vendor/supplier kantin
CREATE TABLE canteen_vendors (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(200) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    contact_person  VARCHAR(100),
    phone           VARCHAR(20),
    email           VARCHAR(100),
    address         TEXT,

    -- Konfigurasi
    vendor_type     VARCHAR(20) NOT NULL,
    contract_start  DATE,
    contract_end    DATE,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_canteen_vendor_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_canteen_vendor_type CHECK (vendor_type IN ('catering', 'supplier', 'internal'))
);

-- Indexes
CREATE INDEX idx_canteen_vendor_tenant_company ON canteen_vendors (tenant_id, company_id);
CREATE INDEX idx_canteen_vendor_rels ON canteen_vendors USING GIN (_rels);
CREATE INDEX idx_canteen_vendor_data ON canteen_vendors USING GIN (_data);

-- Master item menu
CREATE TABLE canteen_menus (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(200) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    description     TEXT,
    category        VARCHAR(20) NOT NULL,

    -- Nutrisi (opsional)
    calories        INT,
    protein_gram    DECIMAL(6,1),
    carbs_gram      DECIMAL(6,1),
    fat_gram        DECIMAL(6,1),

    -- Konfigurasi
    is_halal        BOOLEAN NOT NULL DEFAULT true,
    allergens       TEXT,
    vendor_id       UUID,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_canteen_menu_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_canteen_menu_category CHECK (category IN (
        'main_course', 'side_dish', 'soup', 'dessert',
        'snack', 'beverage', 'fruit'
    )),
    CONSTRAINT chk_canteen_calories CHECK (calories IS NULL OR calories >= 0)
);

-- Indexes
CREATE INDEX idx_canteen_menu_tenant_company ON canteen_menus (tenant_id, company_id);
CREATE INDEX idx_canteen_menu_category ON canteen_menus (category);
CREATE INDEX idx_canteen_menu_vendor ON canteen_menus (vendor_id);
CREATE INDEX idx_canteen_menu_rels ON canteen_menus USING GIN (_rels);
CREATE INDEX idx_canteen_menu_data ON canteen_menus USING GIN (_data);

-- Jadwal menu harian
CREATE TABLE canteen_meal_schedules (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    academic_year_id UUID NOT NULL,

    -- Jadwal
    schedule_date   DATE NOT NULL,
    meal_type       VARCHAR(20) NOT NULL,
    serving_time    TIME,

    -- Menu items (denormalisasi — list menu_id)
    menu_items      JSONB NOT NULL DEFAULT '[]',

    -- Porsi
    estimated_portions INT NOT NULL,
    actual_portions   INT,

    -- Catatan
    note            TEXT,
    prepared_by     VARCHAR(200),

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_canteen_schedule_date_meal UNIQUE (tenant_id, company_id, schedule_date, meal_type),
    CONSTRAINT chk_canteen_meal_type CHECK (meal_type IN ('breakfast', 'lunch', 'dinner', 'snack')),
    CONSTRAINT chk_canteen_portions CHECK (estimated_portions > 0)
);

-- Indexes
CREATE INDEX idx_canteen_schedule_tenant_company ON canteen_meal_schedules (tenant_id, company_id);
CREATE INDEX idx_canteen_schedule_date ON canteen_meal_schedules (schedule_date);
CREATE INDEX idx_canteen_schedule_year ON canteen_meal_schedules (academic_year_id);
CREATE INDEX idx_canteen_schedule_meal ON canteen_meal_schedules (meal_type);
CREATE INDEX idx_canteen_schedule_rels ON canteen_meal_schedules USING GIN (_rels);
CREATE INDEX idx_canteen_schedule_data ON canteen_meal_schedules USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `vendor_type` | catering/supplier/internal | 3 model: catering full-service, supplier bahan baku, dapur internal |
| `category` (menu) | 7 kategori | Menu Indonesia: main course (nasi+lauk), side dish, sayur/sup, dessert, snack, minuman, buah |
| `is_halal` | BOOLEAN, default true | Wajib halal untuk pesantren (ADR-009). Default true karena mayoritas target market |
| `allergens` | TEXT, nullable | Free-text daftar alergen — format standar belum dibutuhkan di MVP |
| `calories/protein/carbs/fat` | Nullable | Informasi nutrisi opsional — tidak semua sekolah melacak ini |
| `menu_items` | JSONB array | Denormalisasi daftar menu per meal — satu meal bisa terdiri dari beberapa item (nasi + lauk + sayur + buah) |
| `estimated_portions` | INT | Estimasi porsi berdasarkan jumlah santri aktif — untuk perencanaan |
| `actual_portions` | INT, nullable | Porsi aktual yang disajikan — untuk monitoring food waste |
| `meal_type` | 4 tipe | Breakfast, lunch, dinner, snack — sesuai ritme pesantren |

### Vernon Relationships

**canteen_menus:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `vendor` | belongs_to | **Tidak** | Vendor hanya perlu ditampilkan di detail, bukan list |

**canteen_meal_schedules:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

### _rels / _data Structure

```json
// canteen_menus
{
  "_rels": {
    "vendor_id": "018f..."
  },
  "_data": {
    "vendor": { "id": "018f...", "name": "CV Berkah Catering", "code": "BKC" }
  }
}

// canteen_meal_schedules
{
  "_rels": {
    "academic_year_id": "018f..."
  },
  "_data": {
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "menu_items_detail": [
      { "id": "018f...", "name": "Nasi Putih", "category": "main_course" },
      { "id": "018f...", "name": "Ayam Goreng", "category": "main_course" },
      { "id": "018f...", "name": "Sayur Asem", "category": "soup" },
      { "id": "018f...", "name": "Pisang", "category": "fruit" }
    ]
  }
}
```

### API Endpoints

```
# Vendors (Master)
GET    /api/v1/canteen-vendors                         — List vendor
POST   /api/v1/canteen-vendors                         — Buat vendor
PUT    /api/v1/canteen-vendors/{id}                    — Update vendor

# Menus (Master)
GET    /api/v1/canteen-menus                           — List menu items
POST   /api/v1/canteen-menus                           — Buat menu item
PUT    /api/v1/canteen-menus/{id}                      — Update menu item

# Meal Schedules
GET    /api/v1/canteen-meal-schedules                  — List jadwal (filter by date range)
GET    /api/v1/canteen-meal-schedules/today             — Menu hari ini
GET    /api/v1/canteen-meal-schedules/week              — Menu minggu ini
POST   /api/v1/canteen-meal-schedules                  — Buat jadwal
POST   /api/v1/canteen-meal-schedules/bulk             — Bulk create jadwal mingguan
PUT    /api/v1/canteen-meal-schedules/{id}             — Update jadwal
```

## Consequences

### Positive

- **Boarding school-first**: Dirancang untuk kebutuhan makan 3x sehari + snack di pesantren.
- **Transparansi ke orang tua**: Endpoint `/today` dan `/week` bisa ditampilkan di parent dashboard.
- **Vendor management**: Mendukung 3 model operasi kantin (catering, supplier, internal).
- **Nutrisi opsional**: Sekolah yang peduli nutrisi bisa mengisi, yang tidak bisa skip.
- **Halal enforcement**: Default halal sesuai konteks pesantren.
- **Bulk schedule**: Menu mingguan bisa di-input sekaligus.

### Negative / Trade-offs

- **Menu items sebagai JSONB**: `menu_items` di schedule menggunakan JSONB array (bukan junction table) — trade-off simplicity vs normalized query. Untuk MVP, JSONB cukup karena query pattern selalu per-date.
- **No cost tracking**: Belum ada tracking biaya per meal/per porsi — ini domain keuangan (lihat S037).
- **No dietary restriction per student**: Belum ada mapping preferensi diet per santri (vegetarian, alergi). Enhancement di masa depan.
- **No recipe management**: Belum ada resep detail per menu — cukup nama dan kategori untuk MVP.

## Alternatives Considered

### 1. Menu schedule tanpa master menu
- Ditolak: tanpa master, nama menu harus diinput ulang setiap kali scheduling — inkonsisten.

### 2. Junction table untuk menu_items (normalized)
- Ditolak untuk MVP: menambah complexity query. JSONB array cukup karena query seldom joins across schedules.

### 3. Gabungkan vendor dengan asset management (S040)
- Ditolak: vendor kantin punya atribut spesifik (halal certification, contract) yang berbeda dari supplier aset umum.

### 4. Nutrisi sebagai JSONB (bukan kolom terpisah)
- Ditolak: kolom terpisah memudahkan query "menu dengan kalori < 500" dan aggregate nutritional reports.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `VendorDescriptor.TableName()` | — | `"canteen_vendors"` |
| U02 | `MenuDescriptor.TableName()` | — | `"canteen_menus"` |
| U03 | `ScheduleDescriptor.TableName()` | — | `"canteen_meal_schedules"` |
| U04 | Validate rejects invalid `vendor_type` | `"restaurant"` | Error: must be catering/supplier/internal |
| U05 | Validate rejects invalid `category` | `"appetizer"` | Error: invalid category |
| U06 | Validate rejects invalid `meal_type` | `"brunch"` | Error: must be breakfast/lunch/dinner/snack |
| U07 | Validate rejects `estimated_portions <= 0` | `0` | Error: must be positive |
| U08 | Validate accepts valid schedule | All fields valid | No error |

### Integration Tests — Vendors

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create vendor | POST with name + code + type | 201 |
| I02 | Unique vendor code | Create 2 vendors same code | 409/422 |
| I03 | Vendor type CHECK | INSERT with `vendor_type = 'restaurant'` | DB error |

### Integration Tests — Menus

| # | Test Case | Action | Expected |
|---|---|---|---|
| I04 | Create menu item | POST with name + code + category | 201 |
| I05 | Unique menu code | Create 2 menus same code | 409/422 |
| I06 | Category CHECK | INSERT with `category = 'appetizer'` | DB error |
| I07 | Optional nutrition | Create menu without nutrition fields | 201, nulls accepted |

### Integration Tests — Meal Schedules

| # | Test Case | Action | Expected |
|---|---|---|---|
| I08 | Create schedule | POST with date + meal_type + menu_items | 201 |
| I09 | Unique per date+meal_type | Create 2 schedules same date same meal | 409/422 |
| I10 | Get today's menu | GET /today | 200, all meals for today |
| I11 | Get week's menu | GET /week | 200, 7 days × up to 4 meals |
| I12 | Bulk create weekly | POST /bulk for 7 days × 4 meals | 201, 28 records |
| I13 | Meal type CHECK | INSERT with `meal_type = 'brunch'` | DB error |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | VendorUpdated syncs to menus | Update vendor name | `_data.vendor.name` updated |
| I15 | AcademicYearUpdated syncs to schedules | Update year name | `_data.academic_year.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Cannot access other tenant's menus | GET with wrong tenant | 404 |
| I17 | Cannot create schedule cross-tenant | POST with wrong tenant | Error |
