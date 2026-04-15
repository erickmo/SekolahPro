# ADR-003: Hybrid CQRS + Vernon — Decision Criteria

**Status**: Accepted
**Date**: 2026-04-14
**Deciders**: Vernon Corp Engineering Team

## Context

Setelah mengadopsi Clean Architecture + CQRS (ADR-001) dan Vernon denormalized read-cache pattern (ADR-002), tim menghadapi pertanyaan praktis: **kapan menggunakan Vernon dan kapan cukup CQRS dengan sqlc biasa?**

Vernon pattern memiliki overhead yang nyata:
- Schema lebih kompleks (tambahan `_rels` dan `_data` columns)
- SyncEngine perlu dikembangkan dan di-maintain per domain
- Write operations lebih berat (harus populate `_rels`/`_data` saat insert/update)
- Debugging eventual consistency lebih sulit

Menerapkan Vernon ke semua domain adalah over-engineering. Sebaliknya, tidak menggunakan Vernon untuk domain yang tepat berarti performa query yang buruk.

Tim memerlukan **decision criteria yang jelas dan terukur** untuk memilih antara dua approach.

## Decision

Mengadopsi **Hybrid Model** dengan kriteria pemilihan berbasis metrik yang terukur:

### Decision Matrix

| Kriteria | CQRS + sqlc (Standard) | Vernon Pattern |
|----------|------------------------|----------------|
| Jumlah JOIN di query terumit | ≤ 2 JOIN | ≥ 3 JOIN |
| Read:Write ratio | < 10:1 | ≥ 10:1 |
| Strong consistency requirement | Ya | Tidak (eventual OK) |
| Data volatility relasi | Tinggi (sering berubah) | Rendah-sedang |
| Dataset size | < 100K rows | ≥ 100K rows |
| Query latency target | < 500ms OK | < 100ms required |

### Rule of Thumb

**Gunakan CQRS + sqlc (Standard) jika:**
- Domain memiliki ≤ 2 JOIN di query paling kompleks
- Data perlu immediate consistency (pembayaran, stok real-time, auth)
- Write ratio tinggi (order processing, event logging)
- Domain sederhana dengan < 5 use case total

**Gunakan Vernon Pattern jika:**
- Domain memiliki ≥ 3 JOIN di query listing/detail
- Read:Write ≥ 10:1 (catalog, product listing, report)
- Latency listing < 100ms adalah hard requirement
- Relasi parent (category, supplier, warehouse) jarang berubah

### Domain Classification di Boilerplate

```
Standard CQRS (sqlc):
├── auth/           — login, refresh token (strong consistency, low read)
├── users/          — 1-2 JOIN, write-balanced
├── permissions/    — small dataset, write-driven
└── audit_logs/     — write-heavy, append-only

Vernon Pattern:
├── products/       — 5+ JOIN (category, supplier, warehouse, branch, company)
├── orders/         — 4+ JOIN (customer, items, warehouse, company)
├── inventory/      — 3+ JOIN (product, warehouse, branch)
└── product_categories/ — read-heavy catalog browsing
```

### Implementation Guide

#### Standard CQRS Domain

```
internal/
└── usecase/
    ├── command/
    │   └── create_user/
    │       ├── command.go      # struct CreateUserCommand
    │       └── handler.go      # Handle() business logic
    └── query/
        └── get_user/
            ├── query.go        # struct GetUserQuery
            ├── result.go       # struct GetUserResult (DTO)
            └── handler.go      # Handle() — sqlc queries
```

```go
// usecase/query/get_user/handler.go
type GetUserHandler struct {
    repo UserRepository  // simple interface
}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*GetUserResult, error) {
    // Direct sqlc query, max 2 JOIN, fast enough
    user, err := h.repo.FindWithCompany(ctx, q.UserID)
    // map to result DTO
}
```

#### Vernon Domain

```
internal/
└── usecase/
    ├── command/
    │   └── create_product/
    │       ├── command.go
    │       └── handler.go      # Handles write + _rels/_data population
    ├── query/
    │   └── list_products/
    │       ├── query.go
    │       ├── result.go       # Reads from _data JSONB directly
    │       └── handler.go      # Zero-JOIN query via Vernon reader
    └── sync/
        └── product_sync_handler.go  # SyncEngine event handlers
```

```go
// usecase/command/create_product/handler.go
func (h *CreateProductHandler) Handle(ctx context.Context, cmd CreateProductCommand) (uuid.UUID, error) {
    // 1. Fetch relasi yang diperlukan untuk _data
    category, _ := h.categoryRepo.FindByID(ctx, cmd.CategoryID)
    supplier, _ := h.supplierRepo.FindByID(ctx, cmd.SupplierID)
    warehouse, _ := h.warehouseRepo.FindByID(ctx, cmd.WarehouseID)
    
    // 2. Build _rels dan _data snapshot
    rels := buildProductRels(cmd, warehouse)
    data := buildProductData(category, supplier, warehouse)
    
    // 3. Save dengan Vernon columns
    return h.repo.SaveWithVernon(ctx, product, rels, data)
}
```

### Migration dari Standard ke Vernon

Ketika sebuah domain awalnya Standard CQRS lalu tumbuh membutuhkan Vernon:

1. Tambahkan `_rels` dan `_data` columns via migration (nullable dulu)
2. Buat migration script untuk backfill existing rows
3. Implementasi SyncEngine handlers
4. Ganti query handler dari JOIN ke Vernon read
5. Hapus nullable constraint setelah backfill selesai

## Consequences

### Positive

- **Right tool for the right job**: Domain sederhana tidak dibebani Vernon overhead; domain kompleks mendapat performa optimal.
- **Incremental adoption**: Tim bisa mulai dengan Standard CQRS dan upgrade ke Vernon ketika domain terbukti membutuhkannya (data-driven decision).
- **Reduced cognitive load**: Decision matrix memberikan kriteria objektif — tidak perlu debat arsitektur per domain.
- **Maintenance burden terdistribusi**: Hanya domain yang benar-benar butuh Vernon yang menanggung complexity SyncEngine.

### Negative / Trade-offs

- **Dua mental model dalam satu codebase**: Developer harus memahami kapan mereka sedang bekerja di Standard vs Vernon domain.
- **Risk mis-classification**: Domain yang salah diklasifikasi akan menyebabkan performa buruk (seharusnya Vernon tapi pakai Standard) atau unnecessary complexity (seharusnya Standard tapi pakai Vernon).
- **SyncEngine boundary tidak selalu jelas**: Ketika domain A dan B saling mereferensikan, propagasi update bisa menjadi graph kompleks.
- **Testing lebih bervariasi**: Standard domain dan Vernon domain membutuhkan test strategy yang berbeda (Standard lebih mudah di-mock, Vernon butuh test untuk SyncEngine).

## Alternatives Considered

### 1. Vernon untuk semua domain
- Konsistensi arsitektur
- Ditolak karena overhead tidak justified untuk domain sederhana — setiap auth operation tidak perlu maintain `_data` JSONB.

### 2. Standard CQRS untuk semua domain + materialized views
- Lebih sederhana
- Ditolak karena materialized view tidak memberikan flexibility yang sama dengan Vernon untuk partial update dan per-tenant customization.

### 3. Flag per-query (bukan per-domain)
- Granularity tertinggi
- Ditolak karena terlalu fine-grained, sulit di-standardize, dan tidak memberikan clear boundary untuk SyncEngine.
