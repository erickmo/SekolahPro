# ADR-001: Go Clean Architecture + CQRS

**Status**: Accepted
**Date**: 2026-04-14
**Deciders**: Vernon Corp Engineering Team

## Context

Ketika memulai Go backend service, tim menghadapi pilihan arsitektur yang menentukan bagaimana kode diorganisasi, diuji, dan di-scale. Beberapa masalah yang perlu diatasi:

1. **Separation of Concerns**: Handler HTTP tidak boleh mengandung business logic. Perlu pemisahan tegas antara transport layer, business logic, dan data access.
2. **Testability**: Unit test untuk business logic harus bisa berjalan tanpa database, HTTP server, atau external dependency apapun.
3. **Scalability**: Ketika fitur bertambah, codebase harus tetap mudah dinavigasi dan dimodifikasi tanpa risk regresi tinggi.
4. **Read/Write Asymmetry**: Banyak domain memiliki beban baca jauh lebih tinggi dari tulis (10:1 hingga 100:1). CQRS memungkinkan optimasi terpisah untuk query path vs command path.
5. **Team Onboarding**: Arsitektur harus self-documenting — developer baru harus paham di mana meletakkan kode baru tanpa bertanya.

Go sebagai bahasa tidak memiliki opinionated framework seperti Rails atau Django, sehingga tim perlu mendefinisikan sendiri struktur yang konsisten.

## Decision

Mengadopsi **Clean Architecture** (Robert C. Martin) dikombinasikan dengan **CQRS** (Command Query Responsibility Segregation) sebagai arsitektur utama Go service.

### Layer Structure

```
internal/
├── domain/           # Layer 1: Enterprise Business Rules
│   ├── entity/       # Pure domain objects, no framework dependency
│   ├── valueobject/  # Immutable value types
│   └── event/        # Domain events
├── usecase/          # Layer 2: Application Business Rules
│   ├── command/      # Write operations (CreateOrder, UpdateProduct, ...)
│   └── query/        # Read operations (GetOrder, ListProducts, ...)
├── adapter/          # Layer 3: Interface Adapters
│   ├── handler/      # HTTP handlers (Chi router)
│   ├── repository/   # DB implementations (sqlx/sqlc)
│   └── eventhandler/ # Event subscribers
└── infrastructure/   # Layer 4: Frameworks & Drivers
    ├── db/           # PostgreSQL connection, migrations
    ├── cache/        # Redis client
    ├── messaging/    # NATS JetStream / InMemory event bus
    └── telemetry/    # OpenTelemetry, Prometheus
```

### CQRS Implementation

**Command Side** (Write):
```go
// usecase/command/create_product.go
type CreateProductCommand struct {
    TenantID    uuid.UUID
    Name        string
    SKU         string
    Price       decimal.Decimal
}

type CreateProductHandler struct {
    repo    ProductRepository
    eventBus EventBus
}

func (h *CreateProductHandler) Handle(ctx context.Context, cmd CreateProductCommand) (uuid.UUID, error) {
    // Pure business logic — no HTTP, no DB driver details
}
```

**Query Side** (Read):
```go
// usecase/query/get_product.go
type GetProductQuery struct {
    TenantID  uuid.UUID
    ProductID uuid.UUID
}

type GetProductResult struct {
    ID       uuid.UUID
    Name     string
    // ... flat DTO, optimized for read
}

type GetProductHandler struct {
    reader ProductReader  // separate read interface
}
```

### Dependency Rule

Dependensi hanya boleh mengarah ke dalam (inward):
- `infrastructure` → `adapter` → `usecase` → `domain`
- `domain` tidak boleh mengimport package apapun di luar Go standard library
- Interface didefinisikan di `usecase`, implementasi di `adapter`/`infrastructure`

### Repository Interface Pattern

```go
// usecase/port/repository.go — interface defined in usecase layer
type ProductRepository interface {
    Save(ctx context.Context, p *domain.Product) error
    FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
}

// adapter/repository/product_pg.go — implementation in adapter layer
type productPgRepository struct {
    db *sqlx.DB
    q  *sqlc.Queries
}
```

## Consequences

### Positive

- **Testability tinggi**: Business logic di `usecase` layer dapat diuji dengan mock repository tanpa database nyata — test suite berjalan dalam < 1 detik.
- **Framework independence**: Mengganti Chi dengan Gin, atau PostgreSQL dengan MySQL, hanya menyentuh satu layer tanpa perubahan business logic.
- **Parallel read/write optimization**: Query handler dapat menggunakan read replica, caching layer, atau bahkan denormalized store (Vernon pattern) secara independen dari write path.
- **Clear ownership**: Setiap file memiliki satu alasan untuk berubah (Single Responsibility Principle).
- **Onboarding friction rendah**: Developer baru tinggal cari `usecase/command/` atau `usecase/query/` sesuai operasi yang dibutuhkan.
- **Audit trail alami**: Command objects merepresentasikan intent dengan eksplisit — mudah di-log untuk audit.

### Negative / Trade-offs

- **Boilerplate lebih banyak**: Untuk CRUD sederhana, perlu membuat command struct, handler, repository interface, dan implementasi secara terpisah. Estimasi 3-4x lebih banyak file vs layered MVC.
- **Indirection overhead**: Debugging memerlukan trace melalui lebih banyak layer. Stack trace bisa panjang.
- **Learning curve**: Tim yang terbiasa MVC perlu waktu 1-2 sprint untuk internalisasi dependency rule.
- **Over-engineering risk**: Domain sederhana dengan < 3 use case mungkin tidak membutuhkan kompleksitas ini.
- **Eventual consistency**: CQRS mendorong eventual consistency antara write dan read model — perlu desain yang hati-hati untuk UX yang membutuhkan immediate read-after-write.

## Alternatives Considered

### 1. Layered MVC (Controller → Service → Repository)
- Lebih familiar, lebih sedikit boilerplate
- Ditolak karena service layer sering menjadi God Object, sulit diuji tanpa database, dan tidak mendukung read/write split secara alami.

### 2. Hexagonal Architecture (Ports & Adapters) only, tanpa CQRS
- Sudah lebih baik dari MVC dalam testability
- Ditolak karena tidak secara eksplisit memisahkan read vs write path, yang penting untuk optimasi performa di domain read-heavy.

### 3. Event Sourcing penuh
- Maximum audit trail dan temporal queries
- Ditolak karena kompleksitas operasional sangat tinggi (snapshot management, event versioning), tidak justified untuk semua domain. Dipertimbangkan lagi jika ada domain audit-critical spesifik.

### 4. Flat package structure (satu package per domain)
- Lebih sederhana, lebih Go-idiomatic
- Ditolak karena tidak memberikan guardrail untuk dependency direction, dan cepat menjadi spaghetti saat domain berkembang.
