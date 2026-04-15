# ADR-004: Multi-Tenant 4-Level Hierarchy + Two-Phase JWT

**Status**: Accepted
**Date**: 2026-04-14
**Deciders**: Vernon Corp Engineering Team

## Context

Sistem dirancang untuk mendukung dua deployment model:

1. **Self-hosted (Single)**: Satu instalasi untuk satu perusahaan — cocok untuk enterprise dengan kebutuhan data isolation penuh.
2. **SaaS (Multi-tenant)**: Satu instalasi melayani banyak tenant (perusahaan berbeda) — cocok untuk bisnis B2B SaaS.

Tantangan desain:
- **Organizational complexity**: Perusahaan enterprise sering memiliki struktur: Holding → Anak Perusahaan → Cabang → Gudang. Data perlu scope ke level yang tepat.
- **Data isolation**: Data tenant A tidak boleh terlihat oleh tenant B dalam mode multi-tenant.
- **Auth flexibility**: Login di SaaS harus bisa memilih "masuk sebagai tenant mana" setelah authentication.
- **Row-level security**: Setiap query harus secara implisit difilter oleh scope user.
- **Cross-tenant operation**: Superadmin perlu bisa mengakses data lintas tenant tanpa login ulang.

### Business Requirements

```
PT Vernon Indonesia (Tenant)
├── PT Vernon Retail (Company)
│   ├── Cabang Jakarta Pusat (Branch)
│   │   ├── Gudang Utama JKT (Warehouse)
│   │   └── Gudang Transit JKT (Warehouse)
│   └── Cabang Bandung (Branch)
│       └── Gudang Bandung (Warehouse)
└── PT Vernon Distribution (Company)
    └── Cabang Surabaya (Branch)
        └── Gudang Surabaya (Warehouse)
```

## Decision

### 1. Four-Level Scope Hierarchy

Mengadopsi hierarki 4 level sebagai scope standar untuk semua data dalam sistem:

```
Tenant → Company → Branch → Warehouse
```

**Tenant**: Unit paling atas. Dalam SaaS = satu pelanggan berbayar. Dalam self-hosted = satu perusahaan. Semua data memiliki `tenant_id`.

**Company**: Anak perusahaan atau business unit di bawah tenant. Beberapa domain (HR, Finance) di-scope ke company level.

**Branch**: Cabang fisik dari company. Point of Sale, inventory fisik, dan operasional harian di-scope ke branch.

**Warehouse**: Lokasi penyimpanan di dalam branch. Stok, picking, dan receiving di-scope ke warehouse.

### 2. Database Schema Convention

Setiap tabel mengikuti convention:

```sql
-- Master tables: hanya tenant_id
CREATE TABLE product_categories (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    name        VARCHAR(255) NOT NULL,
    -- ...
);

-- Operational tables: full hierarchy
CREATE TABLE inventory_transactions (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id),
    company_id   UUID NOT NULL REFERENCES companies(id),
    branch_id    UUID NOT NULL REFERENCES branches(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    -- ...
);

-- Indexes wajib untuk setiap scope level
CREATE INDEX idx_inventory_tx_tenant  ON inventory_transactions(tenant_id);
CREATE INDEX idx_inventory_tx_branch  ON inventory_transactions(branch_id);
CREATE INDEX idx_inventory_tx_wh      ON inventory_transactions(warehouse_id);
```

### 3. Two-Phase JWT untuk Mode Multi-Tenant

Mode SaaS membutuhkan alur auth dua fase:

**Phase 1 — Authentication**: User membuktikan identitas (username/password, OAuth).

```json
// Phase 1 Token — "Pre-auth token"
{
  "sub": "user-uuid",
  "phase": 1,
  "email": "user@company.com",
  "tenants": [
    {"id": "tenant-1-uuid", "name": "PT Vernon Indonesia", "role": "admin"},
    {"id": "tenant-2-uuid", "name": "PT Mitra Corp", "role": "viewer"}
  ],
  "exp": 1714000000  // short TTL: 5 menit
}
```

**Phase 2 — Tenant Selection**: User memilih tenant yang akan diakses. Server memvalidasi membership dan menerbitkan full-scope token.

```json
// Phase 2 Token — "Operational token"
{
  "sub":          "user-uuid",
  "phase":        2,
  "tenant_id":    "tenant-1-uuid",
  "company_id":   "company-uuid",    // null jika scope tenant-wide
  "branch_id":    "branch-uuid",     // null jika scope company-wide
  "warehouse_id": "warehouse-uuid",  // null jika scope branch-wide
  "role":         "branch_manager",
  "permissions":  ["inventory:read", "inventory:write", "order:read"],
  "exp":          1714003600  // TTL: 1 jam
}
```

### 4. Scope Resolution di Middleware

```go
// adapter/middleware/scope.go
type ScopeContext struct {
    TenantID    uuid.UUID
    CompanyID   *uuid.UUID  // nullable — bisa null untuk superadmin atau tenant-wide role
    BranchID    *uuid.UUID
    WarehouseID *uuid.UUID
    UserID      uuid.UUID
    Role        string
    Permissions []string
}

func ScopeMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := extractAndValidateJWT(r, jwtSecret)
            
            if claims.Phase != PhaseOperational {
                http.Error(w, "tenant selection required", http.StatusForbidden)
                return
            }
            
            scope := ScopeContext{
                TenantID:    claims.TenantID,
                CompanyID:   claims.CompanyID,
                BranchID:    claims.BranchID,
                WarehouseID: claims.WarehouseID,
                UserID:      claims.Subject,
                Permissions: claims.Permissions,
            }
            
            ctx := context.WithValue(r.Context(), ScopeContextKey, scope)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 5. Repository Scope Enforcement

Semua repository secara eksplisit menggunakan scope dari context:

```go
// adapter/repository/product_repository.go
func (r *productRepository) ListByScope(ctx context.Context, p Pagination) ([]domain.Product, error) {
    scope := middleware.ScopeFromContext(ctx)
    
    query := `
        SELECT * FROM products
        WHERE tenant_id = $1
          AND ($2::uuid IS NULL OR warehouse_id = $2)
        ORDER BY created_at DESC
        LIMIT $3 OFFSET $4
    `
    
    return r.db.QueryxContext(ctx, query,
        scope.TenantID,
        scope.WarehouseID,  // null = tampilkan semua warehouse dalam tenant
        p.Limit, p.Offset,
    )
}
```

### 6. Single-Tenant Mode

Untuk deployment self-hosted, variabel `DEPLOYMENT_MODE=single` mengaktifkan:
- Hanya Phase 2 token yang diterbitkan langsung (skip tenant selection)
- `tenant_id` di-inject otomatis dari config, tidak dari JWT
- UI tidak menampilkan tenant switcher

## Consequences

### Positive

- **One codebase, two deployment modes**: Tidak perlu maintain dua versi aplikasi.
- **Fine-grained access control**: Scope bisa di-set pada level mana pun (tenant-wide, company, branch, atau warehouse-specific).
- **Short-lived Phase 1 token**: Meminimalkan window eksploitasi jika token Phase 1 bocor.
- **Explicit scope dalam code**: Repository tidak pernah bisa "lupa" filter tenant — scope selalu diambil dari context.
- **Scalable hierarchy**: Jika di masa depan perlu level kelima (misal: aisle dalam warehouse), bisa ditambahkan.
- **Audit trail**: Setiap operasi terekam dengan full scope context (tenant, company, branch, warehouse).

### Negative / Trade-offs

- **UX friction**: Mode multi-tenant mengharuskan dua langkah login — bisa terasa lambat bagi user yang hanya punya satu tenant.
- **Token size**: Phase 2 token mengandung permissions list — bisa besar jika permission system granular.
- **Null nullable complexity**: `company_id`, `branch_id`, `warehouse_id` yang nullable membutuhkan pengecekan null di setiap query.
- **Migration complexity**: Menambah level baru ke hierarki di kemudian hari membutuhkan schema migration masif.
- **Cross-scope queries sulit**: Report yang agregat data lintas branch atau company membutuhkan query khusus dengan elevated scope.

## Alternatives Considered

### 1. 2-Level (Tenant → User)
- Paling sederhana
- Ditolak karena tidak mendukung organizational structure enterprise — tidak bisa restrict warehouse manager agar hanya melihat data warehousenya.

### 2. 3-Level (Tenant → Branch → User)
- Skip Company level
- Ditolak karena banyak holding company memiliki anak perusahaan dengan accounting terpisah — Company level penting untuk financial isolation.

### 3. Attribute-Based Access Control (ABAC) tanpa hierarki fixed
- Sangat fleksibel
- Ditolak karena kompleksitas implementasi sangat tinggi, sulit diaudit, dan tidak self-documenting untuk developer.

### 4. Single JWT (one-phase)
- Lebih sederhana
- Ditolak karena dalam SaaS mode, user dengan akses ke banyak tenant harus di-authenticate ulang setiap kali ganti tenant, atau token harus mengandung semua tenant (bloat ukuran token dan permission scope).

### 5. Separate database per tenant
- Isolasi paling kuat
- Ditolak karena operational overhead sangat tinggi (schema migration harus dijalankan N kali untuk N tenant), dan tidak scalable dalam SaaS model.
