# Go Boilerplate API

## Tech Stack
- **Language**: Go 1.25
- **Router**: Chi v5
- **DI**: Uber FX
- **Database**: PostgreSQL + sqlx + sqlc
- **Cache**: Redis
- **Events**: Watermill (InMemory dev / NATS prod)
- **Tracing**: OpenTelemetry + Jaeger
- **Metrics**: Prometheus
- **Auth**: JWT (golang-jwt)
- **Logging**: zerolog
- **Hot Reload**: air

## Struktur Proyek

```
api/
├── cmd/api/                 # Entry point (main.go, server.go)
├── infrastructure/
│   ├── config/              # Viper config loader
│   └── database/            # DB connection + sqlc generated code
├── internal/
│   ├── domain/{name}/       # Entity, errors, events, repository interfaces
│   ├── command/{cmd_name}/  # Command handler (1 folder = 1 command)
│   ├── query/{qry_name}/    # Query handler (1 folder = 1 query)
│   ├── eventhandler/        # Cross-domain event handlers
│   └── delivery/http/       # HTTP handlers
├── pkg/
│   ├── commandbus/          # Command dispatcher
│   ├── querybus/            # Query dispatcher
│   ├── eventbus/            # Event pub/sub
│   ├── jwt/                 # JWT service
│   ├── middleware/          # HTTP middleware
│   └── pagination/          # Pagination helper
├── migrations/              # SQL migrations
├── sqlc/                    # sqlc input (queries/ + schema/)
└── tests/integration/       # Integration tests (Testcontainers)
```

## Aturan Arsitektur

1. **Domain layer (internal/domain)**: ZERO dependency ke package lain
2. **Command handler**: hanya untuk write operations (create/update/delete)
3. **Query handler**: hanya untuk read operations
4. **HTTP handler**: tidak boleh ada business logic — hanya decode/dispatch/encode
5. **Repository**: interface di domain/, implementasi di infrastructure/database/

## Menambah Domain Baru

1. Buat `internal/domain/{name}/` (entity.go, errors.go, events.go)
2. Buat command handler di `internal/command/{cmd_name}/handler.go`
3. Buat query handler di `internal/query/{qry_name}/handler.go`
4. Buat HTTP handler di `internal/delivery/http/{name}_handler.go`
5. Buat repository implementation di `infrastructure/database/{name}_repository.go`
6. Daftarkan semua di `cmd/api/main.go`

## Commands

```bash
make dev              # Build + run (tanpa hot reload)
make air              # Run dengan hot reload
make test             # Unit tests
make test-integration # Integration tests (butuh Docker)
make migrate-up       # Jalankan migrasi
make migrate-create name=nama_migrasi  # Buat migration baru
make infra-up         # Start Docker services
make sqlc             # Generate code dari SQL
make swagger          # Generate Swagger docs ke docs/ (butuh swag)
make tools-install    # Install semua dev tools (termasuk swag)
```

## API Documentation (Swagger)

Jalankan `make tools-install` lalu `make swagger` untuk menggenerate docs.

Setelah generate, Swagger UI bisa diakses di `http://localhost:8080/swagger/index.html`
(file static di-serve dari `docs/` directory).

Tambahkan annotations ke handler baru:
```go
// Create godoc
// @Summary      Buat contoh baru
// @Tags         examples
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      201  {object}  map[string]string
// @Failure      422  {object}  map[string]string
// @Router       /api/v1/examples [post]
func (h *ExampleHandler) Create(w http.ResponseWriter, r *http.Request) {
```

## Tenant & Org Scope Architecture

### Hierarki 4 Level

```
Tenant (L1, wajib) → Company (L2, wajib) → Branch (L3, opsional) → Warehouse (L4, opsional)
```

### Mode

| Mode | Config | JWT | Use Case |
|---|---|---|---|
| `single` | `TENANT_MODE=single` + `TENANT_ID` + `COMPANY_ID` | Tidak wajib | Self-hosted / dedicated |
| `multi`  | `TENANT_MODE=multi` | Wajib ada semua level yang dipakai | SaaS / shared |

### Two-Phase JWT (mode multi)

- **Fase 1** (setelah login): token berisi `tenant_id` saja. Dipakai untuk memilih company.
- **Fase 2** (setelah select-scope): token berisi `tenant_id` + `company_id` + opsional branch/warehouse.

Generate via `jwtSvc.GenerateTokenPair()` (fase 1) dan `jwtSvc.GenerateScopedTokenPair()` (fase 2).

### Aturan Layer

```
pkg/scope/          ← Scope struct (pure value object), Resolver interface, middlewares
pkg/tenant/         ← TenantID-only resolver (legacy helper, masih tersedia)
cmd/api/main.go     ← Satu tempat: pilih resolver berdasarkan TENANT_MODE
cmd/api/server.go   ← Urutan middleware: RequireAuth → ResolveScope → RequireScope

internal/domain/    ← BOLEH import pkg/scope (pure value object, no framework dep)
internal/command/   ← Baca Scope dari context → pass ke repo sebagai parameter
internal/query/     ← Baca Scope dari context → pass ke repo sebagai parameter
infrastructure/     ← Repository impl inject tenant_id + company_id ke SQL query
```

### Middleware Flow

```
Request → RequireAuth → ResolveScope (fail-open) → RequireScope(L1,L2) → Handler
                              │
                 single: baca dari config (ConfigScopeResolver)
                  multi: baca dari JWT claims, semua 4 level (JWTScopeResolver)
```

### RequireScope per route

```go
// Contoh: route yang butuh branch
r.With(scope.RequireScope(scope.LevelTenant, scope.LevelCompany, scope.LevelBranch)).
    Get("/branch-specific", handler)
```

### Menambah Tabel Baru

1. Selalu tambah `tenant_id UUID NOT NULL` dan `company_id UUID NOT NULL` di CREATE TABLE
2. Selalu jadikan `tenant_id, company_id` sebagai leading columns di composite index
3. Selalu sertakan `WHERE tenant_id = $1 AND company_id = $2` di semua query
4. Selalu sertakan `TenantID` dan `CompanyID` di setiap domain event yang dipublish

### Anti-Pattern Yang Dilarang

```go
// DILARANG: repo membaca scope dari context (hidden dependency)
func (r *Repo) List(ctx context.Context) ([]*E, error) {
    s, _ := scope.ScopeFromContext(ctx)  // ← JANGAN
}

// BENAR: scope sebagai parameter eksplisit
func (r *Repo) List(ctx context.Context, s scope.Scope, ...) ([]*E, error) {
    // WHERE tenant_id = s.TenantID AND company_id = s.CompanyID
}
```

## Environment

Copy `.env.example` ke `.env` dan sesuaikan nilai:
- `DATABASE_URL` — koneksi PostgreSQL
- `REDIS_URL` — koneksi Redis
- `JWT_SECRET` — min 32 karakter
- `TENANT_MODE` — `single` atau `multi`
- `TENANT_ID` — UUID top-level namespace (wajib jika mode=single)
- `COMPANY_ID` — UUID company/legal entity (wajib jika mode=single)
- `BRANCH_ID` — UUID cabang (opsional)
- `WAREHOUSE_ID` — UUID gudang (opsional, lebih baik kirim via request body)

## Hybrid Architecture: CQRS vs Vernon

Dua pola domain tersedia. Pilih berdasarkan karakteristik domain:

| Kriteria | CQRS (existing) | Vernon (_rels/_data) |
|---|---|---|
| Jumlah relasi/JOIN | Sedikit (≤ 2) | Banyak (3+) |
| Read:Write ratio | Seimbang | Read-heavy (10:1+) |
| Logic bisnis | Kompleks (workflow, saga) | Sederhana (CRUD) |
| Konsistensi | Strong consistency | Eventual consistency OK |

### Vernon Pattern — Menambah Domain Baru

1. Buat descriptor (`internal/domain/{name}/descriptor.go`):
   ```go
   type Descriptor struct{}
   func (d *Descriptor) TableName() string { return "orders" }
   func (d *Descriptor) DefaultRels() map[string]vernon.RelDef { ... }
   func (d *Descriptor) Validate(data map[string]any) error { ... }
   ```

2. Buat migration dengan 10 kolom (`migrations/NNN_{name}.sql`):
   ```sql
   CREATE TABLE orders (
       id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
       tenant_id UUID NOT NULL, company_id UUID NOT NULL,
       _rels JSONB NOT NULL DEFAULT '{}',
       _data JSONB NOT NULL DEFAULT '{}',
       _sync_status TEXT NOT NULL DEFAULT 'synced',
       _sync_version BIGINT NOT NULL DEFAULT 0,
       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
       updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
       deleted_at TIMESTAMPTZ
   );
   ```

3. Daftarkan di `cmd/api/main.go` → `registerVernonDomains()`:
   ```go
   desc := &orders.Descriptor{}
   repo := vernon.NewBaseRepository(db, desc.TableName())
   svc  := vernon.NewBaseService(db, repo, desc, eb, logger)
   hdlr := vernon.NewBaseHandler(svc, logger)
   registry.Register(desc.TableName(), &vernon.RegisteredDomain{
       Descriptor: desc, Handler: hdlr, Service: svc,
   })
   ```

### Vernon Aturan

1. **Autoload belongs_to**: sisi "few" yang autoload, sisi "many" tidak.
2. **Circular autoload DILARANG**: Registry akan panic saat startup jika terdeteksi.
3. **SyncEvent setelah TX commit**: eventual consistency — `_sync_status` bisa `pending` sementara.
4. **tenant_id + company_id**: selalu kolom terpisah, BUKAN di dalam `_data`.
5. **Scope wajib**: semua query Vernon otomatis difilter tenant_id + company_id.

### Packages

```
pkg/vernon/      ← Core: BaseDomain, BaseRepository, BaseService, BaseHandler, Registry
pkg/vernonsync/  ← SyncEngine: subscribe Watermill + propagate ke consumers
```

## Stack Skills
~/.claude/skills/go-coding-standard/SKILL.md
