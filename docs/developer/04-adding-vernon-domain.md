# 04 — Adding a Vernon Domain

Vernon adalah pola denormalized read-cache: data disimpan dalam kolom `_data` JSONB sehingga GET tidak perlu JOIN. Konsistensi antar domain dijaga secara asinkron oleh `SyncEngine`.

## Kapan Gunakan Vernon

Gunakan checklist ini sebelum memilih Vernon:

```
[✓] Domain memiliki 3+ JOIN untuk membentuk satu response GET
[✓] Ratio read:write lebih dari 10:1
[✓] Business logic sederhana (validasi field, tidak ada workflow/saga)
[✓] Eventual consistency dapat diterima (data boleh terlambat beberapa milidetik)
[✓] Domain ini sering menjadi sumber embed untuk domain lain (belongs_to)

Jika semua di atas terpenuhi → pilih Vernon
Jika ada yang tidak terpenuhi → pertimbangkan CQRS
```

**Contoh domain yang cocok untuk Vernon:**
- Product categories (data referensi statis)
- Products (belongs_to categories, has_many variants)
- Locations / branches (data hierarki relatif statis)
- Currencies / price lists (lookup tables)
- SKU / inventory item specs

**Contoh domain yang TIDAK cocok untuk Vernon:**
- Orders (workflow multi-step, strong consistency)
- Payments (finansial, tidak boleh eventual)
- User authentication (security-critical)

---

## Struktur Tabel Vernon

Setiap tabel Vernon WAJIB memiliki tepat 10 kolom ini:

```
id              UUID          PK, DEFAULT uuid_generate_v7()
tenant_id       UUID          NOT NULL
company_id      UUID          NOT NULL
_rels           JSONB         NOT NULL DEFAULT '{}'  (definisi relasi)
_data           JSONB         NOT NULL DEFAULT '{}'  (data bisnis + embedded)
_sync_status    TEXT          NOT NULL DEFAULT 'synced'
_sync_version   BIGINT        NOT NULL DEFAULT 0
created_at      TIMESTAMPTZ   NOT NULL DEFAULT now()
updated_at      TIMESTAMPTZ   NOT NULL DEFAULT now()
deleted_at      TIMESTAMPTZ   (soft delete)
```

> `uuid_generate_v7()` menghasilkan UUID v7 yang time-sortable — lebih efisien untuk B-tree index daripada UUID v4 random.

---

## RelDef Types

| Type | Arah Relasi | Contoh | IsAutoload Default |
|------|------------|--------|-------------------|
| `belongs_to` | N ke 1 | products → categories | `true` (embed di sisi N) |
| `has_one` | 1 ke 1 | user → profile | `true` atau `false` |
| `has_many` | 1 ke N | order → order_items | `false` (on-demand) |
| `many_to_many` | N ke N | products ↔ tags | `false` (on-demand) |

```go
// pkg/vernon/domain.go — RelDef struct
type RelDef struct {
    Domain      string   // target table name, e.g. "product_categories"
    Type        string   // "belongs_to" | "has_one" | "has_many" | "many_to_many"
    FK          string   // foreign key field name di _data
    LocalKey    string   // key pada entitas ini
    ForeignKey  string   // key pada entitas target
    IsAutoload  bool     // true = embed di _data saat write
    Fields      []string // field yang di-denormalisasi ke _data

    // Many-to-many only
    Junction           string // junction table name
    JunctionLocalKey   string
    JunctionForeignKey string
}
```

---

## Autoload Rules

### Aturan 1: Sisi "Few" yang Autoload

```
product_categories  (sisi "one" / sumber)
        ↑
        autoload dari sini → embed ke _data["category"]
        ↓
products            (sisi "many" / consumer)
```

- **products** mendaftarkan `belongs_to` ke `product_categories` dengan `IsAutoload: true`
- **product_categories** TIDAK mendaftarkan `has_many` ke `products` dengan `IsAutoload: true`

Sederhananya: yang punya FK (`category_id` di `_data` products) yang memasang autoload.

### Aturan 2: Circular Autoload Dilarang

```
Domain A (autoload dari B) → Domain B (autoload dari A)  ← PANIC saat startup
```

Registry mendeteksi circular dependency via DFS saat startup. Jika terdeteksi, aplikasi langsung panic dengan pesan:

```
panic: vernon Registry: cycle detected: A → B — set is_autoload:false on one side
```

**Cara memperbaiki:** Set `IsAutoload: false` pada salah satu sisi.

### Aturan 3: Urutan Register di main.go

Domain sumber harus didaftarkan **sebelum** domain yang me-autoload dari mereka:

```go
// BENAR: sumber (product_categories) didaftarkan lebih dulu
registerVernonDomain(db, registry, eb, logger, &product_category.Descriptor{})
registerVernonDomain(db, registry, eb, logger, &product.Descriptor{})

// SALAH: akan panik karena reverseMap belum lengkap saat validasi
registerVernonDomain(db, registry, eb, logger, &product.Descriptor{})
registerVernonDomain(db, registry, eb, logger, &product_category.Descriptor{})
```

---

## Step-by-Step: Menambah Domain Vernon

Contoh: menambah domain `Warehouse` (belongs_to Branch).

### Step 1 — Buat Descriptor

```go
// internal/domain/warehouse/descriptor.go

// Package warehouse adalah domain Vernon untuk manajemen gudang.
// belongs_to branch — autoload nama cabang ke dalam _data["branch"].
package warehouse

import (
    "errors"
    "fmt"
    "regexp"

    "github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants — hindari magic string di seluruh codebase.
const (
    FieldName     = "name"
    FieldCode     = "code"
    FieldAddress  = "address"
    FieldCapacity = "capacity"
    FieldBranchID = "branch_id"
    FieldIsActive = "is_active"
)

// Relation name constants.
const RelBranch = "branch"

var codeRegex = regexp.MustCompile(`^[A-Z0-9-]+$`)

// Descriptor mengimplementasi vernon.DomainDescriptor untuk warehouses.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "warehouses" }

// DefaultRels mendefinisikan relasi domain ini.
// warehouses belongs_to branches — autoload nama dan kode cabang.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
    return map[string]vernon.RelDef{
        RelBranch: {
            Domain:     "branches",           // tabel sumber
            Type:       vernon.RelBelongsTo,
            FK:         FieldBranchID,        // field di _data warehouses
            LocalKey:   FieldBranchID,
            ForeignKey: "id",                 // PK di tabel branches
            IsAutoload: true,                 // embed branch data saat write
            Fields:     []string{FieldName, FieldCode},
        },
    }
}

// Validate memvalidasi invariant domain sebelum write.
// Hanya validasi struktural — tidak ada DB call.
func (d *Descriptor) Validate(data map[string]any) error {
    name, _ := data[FieldName].(string)
    if name == "" {
        return errors.New("name wajib diisi")
    }

    code, _ := data[FieldCode].(string)
    if code == "" {
        return errors.New("code wajib diisi")
    }
    if !codeRegex.MatchString(code) {
        return fmt.Errorf("code hanya boleh mengandung huruf kapital, angka, dan tanda hubung, got: %q", code)
    }

    if cap, ok := data[FieldCapacity]; ok && cap != nil {
        switch v := cap.(type) {
        case float64:
            if v < 0 {
                return errors.New("capacity tidak boleh negatif")
            }
        }
    }

    return nil
}
```

### Step 2 — Buat Migration SQL

```bash
make migrate-create name=create_warehouses
# Output: Created migrations/006_create_warehouses.sql
```

Edit file migrasi:

```sql
-- migrations/006_create_warehouses.sql

-- +migrate Up
CREATE TABLE IF NOT EXISTS warehouses (
    id            UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id     UUID        NOT NULL,
    company_id    UUID        NOT NULL,
    _rels         JSONB       NOT NULL DEFAULT '{}',
    _data         JSONB       NOT NULL DEFAULT '{}',
    _sync_status  TEXT        NOT NULL DEFAULT 'synced',
    _sync_version BIGINT      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

-- Index scope: filter utama semua query Vernon (tenant_id, company_id)
CREATE INDEX IF NOT EXISTS idx_warehouses_scope
    ON warehouses (tenant_id, company_id, created_at DESC);

-- Index GIN pada _data untuk filter field arbitrer (name, code, is_active, dll)
CREATE INDEX IF NOT EXISTS idx_warehouses_data
    ON warehouses USING GIN (_data)
    WHERE deleted_at IS NULL;

-- Index _sync_status untuk background sweep worker
CREATE INDEX IF NOT EXISTS idx_warehouses_sync
    ON warehouses (_sync_status)
    WHERE _sync_status != 'synced' AND deleted_at IS NULL;

-- Index tambahan: code sering dipakai sebagai query param
CREATE INDEX IF NOT EXISTS idx_warehouses_code
    ON warehouses (tenant_id, company_id, ((_data->>'code')))
    WHERE deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS warehouses;
```

### Step 3 — Register di main.go

Edit `cmd/api/main.go`:

```go
import (
    // ... existing imports ...
    "github.com/yourorg/boilerplate/internal/domain/warehouse"
)

// Di dalam registerVernonDomains():
func registerVernonDomains(
    db *sqlx.DB,
    registry *vernon.Registry,
    eb eventbus.EventBus,
    logger zerolog.Logger,
) {
    // Sumber harus didaftarkan lebih dulu
    registerVernonDomain(db, registry, eb, logger, &product_category.Descriptor{})
    registerVernonDomain(db, registry, eb, logger, &product.Descriptor{})

    // branch harus didaftarkan sebelum warehouse (warehouse belongs_to branch)
    // registerVernonDomain(db, registry, eb, logger, &branch.Descriptor{})
    registerVernonDomain(db, registry, eb, logger, &warehouse.Descriptor{})
}
```

### Step 4 — Jalankan Migrasi

```bash
make migrate-up
```

---

## API Routes yang Auto-Generated

Setelah domain Vernon terdaftar, `Registry.MountRoutes()` secara otomatis membuat route berikut:

```
GET    /{domain}/         List records (pagination + filter via query params)
POST   /{domain}/         Create record baru
GET    /{domain}/{id}     Get record by ID
PUT    /{domain}/{id}     Full replace (replace seluruh _data)
PATCH  /{domain}/{id}     Partial update (merge ke _data)
DELETE /{domain}/{id}     Soft delete
```

Untuk domain `warehouses`:

```
GET    /warehouses?limit=20&offset=0&sort=-created_at
POST   /warehouses
GET    /warehouses/{id}
PUT    /warehouses/{id}
PATCH  /warehouses/{id}
DELETE /warehouses/{id}
```

### Query Parameters untuk List

| Parameter | Contoh | Keterangan |
|-----------|--------|-----------|
| `limit` | `?limit=20` | Jumlah item per halaman |
| `offset` | `?offset=40` | Skip N item pertama |
| `sort` | `?sort=-created_at` | Prefix `-` = DESC, tanpa prefix = ASC |
| `page` | `?page=3` | Alias untuk offset (auto-computed) |
| `{field}` | `?code=WH-001` | Filter berdasarkan field di `_data` |
| `{field}_gte` | `?capacity_gte=100` | Greater than or equal |
| `{field}_lte` | `?capacity_lte=500` | Less than or equal |

### Request Body untuk Create/Update

```json
{
  "name": "Gudang Utama Jakarta",
  "code": "GU-JKT-001",
  "address": "Jl. Industri No. 1, Jakarta Utara",
  "capacity": 5000,
  "branch_id": "<uuid-branch>",
  "is_active": true
}
```

Response GET akan menyertakan data yang sudah di-embed:

```json
{
  "id": "<uuid>",
  "tenant_id": "<uuid>",
  "company_id": "<uuid>",
  "_data": {
    "name": "Gudang Utama Jakarta",
    "code": "GU-JKT-001",
    "address": "...",
    "capacity": 5000,
    "branch_id": "<uuid-branch>",
    "is_active": true,
    "branch": {
      "name": "Cabang Jakarta",
      "code": "CBG-JKT"
    }
  },
  "_sync_status": "synced",
  "created_at": "2026-04-14T10:00:00Z",
  "updated_at": "2026-04-14T10:00:00Z"
}
```

---

## SyncEvent Flow Diagram

```
╔══════════════════════════════════════════════════════════════════════╗
║                        WRITE OPERATION                               ║
║                                                                      ║
║  POST /branches/{id}  ─────────────────────────────────────────────  ║
║  BaseHandler.Update()                                                ║
║      │                                                               ║
║      ▼                                                               ║
║  BaseService.Update()                                                ║
║      ├─ 1. Validate() via Descriptor                                 ║
║      ├─ 2. BEGIN TX                                                  ║
║      ├─ 3. UPDATE branches SET _data = $newData, ...                 ║
║      ├─ 4. COMMIT TX                                                 ║
║      └─ 5. EventBus.Publish(SyncEvent{                               ║
║               domain:   "branches",                                  ║
║               entity_id: branchID,                                   ║
║               action:    "update",                                   ║
║           })                                                         ║
╚═════════════════════════════════════════════════════════════════════╝
                    │ publish AFTER commit (eventual)
                    ▼
╔══════════════════════════════════════════════════════════════════════╗
║                       SYNC ENGINE                                    ║
║                                                                      ║
║  SyncEngine.handleSyncEvent(event{domain: "branches"})               ║
║      │                                                               ║
║      ├─ Registry.GetConsumers("branches")                            ║
║      │       └─ returns: [ConsumerRef{                               ║
║      │               ConsumerDomain: "warehouses",                   ║
║      │               RelName:        "branch",                       ║
║      │               RelDef:         {FK:"branch_id", fields:[...]}, ║
║      │           }]                                                  ║
║      │                                                               ║
║      ├─ Fetch source: SELECT _data FROM branches WHERE id = branchID ║
║      │                                                               ║
║      ├─ Extract fields: {name, code} dari branch._data               ║
║      │                                                               ║
║      └─ propagateBelongsTo():                                        ║
║             SELECT * FROM warehouses                                 ║
║             WHERE _data->>'branch_id' = branchID                     ║
║             AND tenant_id = $1 AND company_id = $2                   ║
║                 │                                                    ║
║                 └─ For each warehouse:                               ║
║                     CASUpdate(                                       ║
║                       id, sync_version,                              ║
║                       key="branch",                                  ║
║                       value={name, code}                             ║
║                     )                                                ║
║                     → UPDATE warehouses                              ║
║                       SET _data = jsonb_set(_data, '{branch}', ...)  ║
║                       WHERE id=$1 AND _sync_version=$2               ║
╚══════════════════════════════════════════════════════════════════════╝
                    │ _sync_status = "synced" setelah CAS sukses
                    │ _sync_status = "error" jika CAS conflict (retry)
                    ▼
╔══════════════════════════════════════════════════════════════════════╗
║  GET /warehouses/{id}                                                ║
║  Response._data["branch"] sudah terupdate:                           ║
║  { "name": "Cabang Jakarta Baru", "code": "CBG-JKT-NEW" }           ║
╚══════════════════════════════════════════════════════════════════════╝
```

---

## Menambah Domain has_many (Contoh: Order → OrderItems)

Untuk relasi `has_many`, order menyimpan embed dari item-item-nya:

```go
// internal/domain/order_item/descriptor.go
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
    return map[string]vernon.RelDef{
        "order": {
            Domain:     "orders",
            Type:       vernon.RelHasMany,   // has_many dari sisi "one"
            FK:         "order_id",          // field di _data order_items
            LocalKey:   "order_id",
            ForeignKey: "id",
            IsAutoload: true,               // trigger sync ke parent order
            Fields:     []string{"name", "quantity", "unit_price"},
            MaxItems:   100,               // cap untuk array yang di-embed
        },
    }
}
```

Ketika `order_item` dibuat/diupdate, SyncEngine akan:
1. Ambil semua order_items yang punya `order_id` = parentID
2. Update `orders._data["items"]` dengan array items terbaru

---

## Menambah Domain many_to_many (Contoh: Products ↔ Tags)

Many-to-many via junction table — autoload **dinonaktifkan** (`IsAutoload: false`) karena propagasi array bisa mahal:

```go
// internal/domain/product/descriptor.go — tambah relasi tags
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
    return map[string]vernon.RelDef{
        "category": {
            // ... existing belongs_to ...
        },
        "tags": {
            Domain:             "tags",
            Type:               vernon.RelManyToMany,
            FK:                 "tag_ids",          // array of UUIDs di _data products
            LocalKey:           "id",
            ForeignKey:         "id",
            IsAutoload:         false,              // on-demand, bukan autoload
            Fields:             []string{"name", "color"},
            Junction:           "product_tags",     // junction table
            JunctionLocalKey:   "product_id",
            JunctionForeignKey: "tag_id",
        },
    }
}
```

Untuk many-to-many, client harus memanggil endpoint khusus untuk resolve relasi, atau relasi diload on-demand oleh service layer.

---

## Checklist Sebelum PR

```
[ ] Descriptor.TableName() mengembalikan snake_case plural (e.g. "warehouses")
[ ] DefaultRels() mengembalikan map literal baru setiap call (bukan var global)
[ ] Validate() tidak melakukan DB/HTTP call (hanya invariant struktural)
[ ] Migration memiliki tepat 10 kolom wajib Vernon
[ ] Index scope: (tenant_id, company_id, created_at DESC)
[ ] Index GIN pada _data
[ ] Index _sync_status untuk background worker
[ ] Domain sumber didaftarkan sebelum consumer di registerVernonDomains()
[ ] Tidak ada circular autoload (Registry akan panic jika ada)
[ ] Tidak ada kolom tenant_id/company_id di dalam _data (selalu kolom terpisah)
```
