# ADR-002: Vernon Denormalized Read-Cache Pattern (_rels/_data JSONB)

**Status**: Accepted
**Date**: 2026-04-14
**Deciders**: Vernon Corp Engineering Team

## Context

Domain seperti `products`, `orders`, dan `inventory` dalam sistem multi-tenant memiliki karakteristik query yang kompleks:

1. **Banyak JOIN**: Menampilkan satu produk membutuhkan JOIN ke `product_categories`, `suppliers`, `warehouses`, `branches`, `companies`, dan `tenants` — bisa 5-7 JOIN dalam satu query.
2. **Read-heavy ratio**: Read:Write = 10:1 hingga 100:1 untuk domain seperti product catalog dan order history.
3. **Latency requirement**: Dashboard dan listing page harus respond < 100ms. Query dengan 7 JOIN pada tabel besar (> 1M rows) sering > 500ms bahkan dengan index optimal.
4. **Denormalization risk**: Denormalisasi manual ke flat columns rawan inkonsistensi saat data berubah.
5. **Schema rigidity**: Menambah field baru ke read model membutuhkan ALTER TABLE yang bisa lock tabel production.

### Masalah Konkret

```sql
-- Query ini terlalu lambat untuk high-traffic listing
SELECT 
    p.*, 
    c.name as category_name, c.slug as category_slug,
    s.name as supplier_name, s.code as supplier_code,
    w.name as warehouse_name, w.code as warehouse_code,
    b.name as branch_name,
    co.name as company_name,
    t.name as tenant_name
FROM products p
JOIN product_categories c ON p.category_id = c.id
JOIN suppliers s ON p.supplier_id = s.id  
JOIN warehouses w ON p.warehouse_id = w.id
JOIN branches b ON w.branch_id = b.id
JOIN companies co ON b.company_id = co.id
JOIN tenants t ON co.tenant_id = t.id
WHERE p.tenant_id = $1 AND p.is_active = true
ORDER BY p.created_at DESC
LIMIT 50;
```

## Decision

Mengadopsi **Vernon Pattern**: setiap domain entity yang read-heavy memiliki dua JSONB columns tambahan di tabel utamanya:

- `_rels`: menyimpan foreign key IDs dari semua relasi (untuk SyncEngine re-populate)
- `_data`: menyimpan denormalized snapshot dari semua data relasi yang dibutuhkan untuk read

### Schema Pattern

```sql
CREATE TABLE products (
    -- Normal columns (source of truth untuk writes)
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    category_id     UUID NOT NULL REFERENCES product_categories(id),
    supplier_id     UUID REFERENCES suppliers(id),
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
    name            VARCHAR(255) NOT NULL,
    sku             VARCHAR(100) NOT NULL,
    price           NUMERIC(15,2) NOT NULL,
    
    -- Vernon read-cache columns
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk JSON path queries
CREATE INDEX idx_products_rels ON products USING GIN (_rels);
CREATE INDEX idx_products_data ON products USING GIN (_data);
```

### _rels Structure

`_rels` menyimpan semua foreign key IDs yang diperlukan SyncEngine untuk mengetahui relasi mana yang perlu di-update:

```json
{
  "category_id":  "018f4c2a-1234-7000-8000-000000000001",
  "supplier_id":  "018f4c2a-5678-7000-8000-000000000002",
  "warehouse_id": "018f4c2a-9abc-7000-8000-000000000003",
  "branch_id":    "018f4c2a-def0-7000-8000-000000000004",
  "company_id":   "018f4c2a-1111-7000-8000-000000000005",
  "tenant_id":    "018f4c2a-2222-7000-8000-000000000006"
}
```

### _data Structure

`_data` menyimpan denormalized snapshot yang langsung bisa digunakan untuk response tanpa JOIN:

```json
{
  "category": {
    "id":   "018f4c2a-1234-7000-8000-000000000001",
    "name": "Electronics",
    "slug": "electronics"
  },
  "supplier": {
    "id":   "018f4c2a-5678-7000-8000-000000000002",
    "name": "PT Supplier Utama",
    "code": "SUP-001"
  },
  "warehouse": {
    "id":   "018f4c2a-9abc-7000-8000-000000000003",
    "name": "Gudang Utama Jakarta",
    "code": "WH-JKT-001"
  },
  "branch": {
    "id":   "018f4c2a-def0-7000-8000-000000000004",
    "name": "Cabang Jakarta Pusat"
  },
  "company": {
    "id":   "018f4c2a-1111-7000-8000-000000000005",
    "name": "PT Vernon Indonesia"
  }
}
```

### Read Query (Zero JOIN)

```sql
-- Dengan Vernon pattern, listing page hanya butuh single table scan
SELECT 
    id, name, sku, price,
    _data->>'category' as category,
    _data->>'supplier' as supplier,
    _data->>'warehouse' as warehouse
FROM products
WHERE tenant_id = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT 50;
```

### SyncEngine: Maintaining Consistency

Ketika data relasi berubah (misal: nama category di-update), `SyncEngine` secara async men-update semua entities yang mereferensikan relasi tersebut:

```go
// infrastructure/sync/sync_engine.go
type SyncEngine struct {
    db       *sqlx.DB
    eventBus EventBus
}

// Triggered by CategoryUpdatedEvent
func (s *SyncEngine) OnCategoryUpdated(ctx context.Context, evt CategoryUpdatedEvent) error {
    // Find all products with this category_id in _rels
    // Batch update their _data->>'category' field
    _, err := s.db.ExecContext(ctx, `
        UPDATE products
        SET _data = jsonb_set(_data, '{category}', $1::jsonb),
            updated_at = NOW()
        WHERE _rels->>'category_id' = $2
    `, evt.CategorySnapshot, evt.CategoryID)
    return err
}
```

### Vernon Read Repository

```go
// adapter/repository/product_vernon_reader.go
func (r *productVernonReader) ListByTenant(ctx context.Context, tenantID uuid.UUID, p Pagination) ([]ProductReadModel, error) {
    rows, err := r.db.QueryxContext(ctx, `
        SELECT id, name, sku, price, is_active,
               _data, created_at, updated_at
        FROM products
        WHERE tenant_id = $1 AND is_active = true
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `, tenantID, p.Limit, p.Offset)
    // ...
}
```

## Consequences

### Positive

- **Query latency drastis turun**: Single table scan vs 7-way JOIN. Benchmark internal: 450ms → 12ms untuk listing 50 products dengan dataset 500K rows.
- **Schema flexibility untuk read**: Menambah field baru ke `_data` tidak perlu ALTER TABLE — cukup update SyncEngine dan re-populate via migration script.
- **PostgreSQL GIN index**: JSONB dengan GIN index memungkinkan path queries yang efisien.
- **Isolation antara read dan write**: Write path tidak terpengaruh oleh kebutuhan read model. Bisa di-optimize secara independen.
- **Self-contained rows**: Setiap row mengandung semua informasi yang dibutuhkan untuk render, cocok untuk pagination dan export.

### Negative / Trade-offs

- **Eventual consistency**: Ada jeda waktu antara update relasi dan propagasi ke entities yang mereferensikannya. Perlu toleransi stale data pada read.
- **Storage overhead**: Setiap product row menyimpan duplikat data category, supplier, warehouse. Estimasi overhead 2-5KB per row.
- **SyncEngine complexity**: Harus maintain mapping: "entitas apa yang perlu di-update ketika X berubah?" — ini bisa menjadi kompleks dengan banyak relasi.
- **Write amplification**: Mengubah nama satu category dengan 100K produk = 100K UPDATE di background. Perlu batching dan rate limiting.
- **Debugging lebih sulit**: Data bisa tidak konsisten sementara. Perlu tooling untuk mendeteksi dan memperbaiki stale `_data`.
- **Tidak cocok untuk semua domain**: Domain dengan write-heavy pattern atau strong consistency requirement (pembayaran, stok real-time) sebaiknya tidak menggunakan pattern ini.

## Alternatives Considered

### 1. Materialized Views
- Database-native, automatic refresh tersedia
- Ditolak karena refresh blocking (shared lock), tidak fleksibel untuk partial update, dan sulit dikustomisasi per-tenant.

### 2. Separate Read Model Table
- Clear separation antara write dan read tables
- Ditolak karena doubles storage, memerlukan sinkronisasi table-level yang lebih kompleks, dan sulit di-manage schema evolution.

### 3. Redis Cache Layer
- Sangat cepat, sub-millisecond
- Ditolak sebagai primary read solution karena: cold start perlu warm up, cache invalidation kompleks, tidak queryable (tidak bisa filter/sort), dan menambah operational dependency.

### 4. Elasticsearch
- Full-text search + aggregasi sangat kuat
- Ditolak karena: operational complexity tinggi (cluster management), eventual consistency yang lebih kompleks dari JSONB, dan over-engineering untuk kebutuhan saat ini.

### 5. Denormalized Flat Columns (manual)
- Performance sama bagusnya
- Ditolak karena schema migration untuk setiap field baru, ALTER TABLE di production berisiko, dan tidak fleksibel untuk variasi per-tenant.
