# 05 — Multi-Tenant Guide

## Overview

Boilerplate ini mendukung dua mode tenancy yang dikontrol oleh satu env var:

```env
TENANT_MODE=single   # satu organisasi per deployment
TENANT_MODE=multi    # banyak organisasi, JWT-based isolation
```

---

## Single vs Multi — Perbandingan

| Aspek | Single Mode | Multi Mode |
|-------|-------------|-----------|
| `TENANT_MODE` | `single` | `multi` |
| Sumber TenantID | Config (env var) | JWT claims |
| Sumber CompanyID | Config (env var) | JWT claims (phase 2) |
| JWT wajib? | Tidak wajib (tapi auth tetap ada) | Wajib di semua request |
| Login flow | Satu phase (dapat token langsung) | Two-Phase JWT |
| Cocok untuk | Self-hosted, dedicated deployment | SaaS, shared infrastructure |
| Scope di code | Selalu sama dari config | Berbeda per request/per user |
| BranchID | Config opsional | JWT claims opsional |
| WarehouseID | Config opsional | Direkomendasikan via request body |

---

## Hierarki 4 Level

```
Tenant (L1, UUID wajib)
  └── Company (L2, UUID wajib)
        └── Branch (L3, UUID opsional)
              └── Warehouse (L4, UUID opsional)
```

| Level | Contoh | Wajib? | Sumber (single) | Sumber (multi) |
|-------|--------|--------|-----------------|----------------|
| Tenant | Holding company, SaaS customer | Ya | `TENANT_ID` env | JWT fase 1 |
| Company | PT subsidiary, legal entity | Ya | `COMPANY_ID` env | JWT fase 2 |
| Branch | Cabang operasional | Tidak | `BRANCH_ID` env | JWT fase 2 |
| Warehouse | Lokasi fisik gudang | Tidak | `WAREHOUSE_ID` env | Request body (direkomendasikan) |

---

## Mode Single

### Konfigurasi

```env
TENANT_MODE=single
TENANT_ID=00000000-0000-0000-0000-000000000001
COMPANY_ID=00000000-0000-0000-0000-000000000002

# Opsional
BRANCH_ID=00000000-0000-0000-0000-000000000003
WAREHOUSE_ID=                                    # kosongkan jika tidak dipakai
```

### Cara Kerja

Saat startup, `provideScopeResolver()` di `main.go` membaca config dan membuat `ConfigScopeResolver`:

```go
// pkg/scope/scope.go
type ConfigScopeResolver struct {
    fixed Scope
}

// Resolve selalu mengembalikan Scope yang sama (dari config).
func (r *ConfigScopeResolver) Resolve(_ *http.Request) (Scope, error) {
    return r.fixed, nil
}
```

Scope yang sama diinjeksikan ke setiap request — tidak peduli siapa yang login atau apa isi JWT-nya.

### Flow Request (Single Mode)

```
HTTP Request
    │
    ├─ RequireAuth: validasi JWT (opsional tapi direkomendasikan untuk auth)
    │
    ├─ ResolveScope: ConfigScopeResolver.Resolve()
    │   └─ inject fixed Scope{TenantID, CompanyID} dari config ke context
    │
    ├─ RequireScope(LevelTenant, LevelCompany): validasi Scope ada
    │
    └─ Handler: scope.ScopeFromContext(ctx) → selalu dapat Scope yang sama
```

---

## Mode Multi (Two-Phase JWT)

### Overview

Dalam mode multi, setiap user bisa menjadi member dari multiple company. Oleh karena itu, scope tidak bisa ditentukan saat login — user harus memilih company aktif terlebih dahulu.

```
Phase 1 Token: { user_id, tenant_id }
                          ↓
              User memilih company
                          ↓
Phase 2 Token: { user_id, tenant_id, company_id, branch_id? }
```

### Two-Phase JWT Flow

```
1. POST /auth/login
   ├─ Validasi credential
   └─ Response: {
          token: "<phase1-jwt>",    // berisi tenant_id saja
          user: { id, name, ... },
          companyGroups: [...]      // daftar company yang bisa dipilih
      }

2. User memilih company dari daftar companyGroups

3. POST /auth/select-scope
   ├─ Header: Authorization: Bearer <phase1-jwt>
   ├─ Body: { company_id: "<uuid>", branch_id: "<uuid>" }
   └─ Response: {
          token: "<phase2-jwt>",    // berisi tenant_id + company_id + branch_id
          refreshToken: "..."
      }

4. Semua request berikutnya menggunakan phase2-jwt:
   Authorization: Bearer <phase2-jwt>
   └─ JWTScopeResolver membaca semua level dari claims
```

### JWT Claims Structure

```go
// pkg/jwt/jwt.go — struktur claims

// Phase 1 token — setelah login
type Claims struct {
    UserID   string  `json:"user_id"`
    TenantID string  `json:"tenant_id"`
    // CompanyID, BranchID, WarehouseID belum ada di phase 1
}

// Phase 2 token — setelah select-scope
type ScopedClaims struct {
    UserID      string  `json:"user_id"`
    TenantID    string  `json:"tenant_id"`
    CompanyID   *string `json:"company_id"`
    BranchID    *string `json:"branch_id,omitempty"`
    WarehouseID *string `json:"warehouse_id,omitempty"`
}
```

### JWTScopeResolver

```go
// pkg/scope/scope.go

// JWTScopeResolver membaca Scope dari JWT claims di context.
// Dipasang SETELAH RequireAuth middleware yang menyimpan claims ke context.
type JWTScopeResolver struct {
    claimsKey any
}

func (r *JWTScopeResolver) Resolve(req *http.Request) (Scope, error) {
    v := req.Context().Value(r.claimsKey)
    if v == nil {
        return Scope{}, fmt.Errorf("JWT claims tidak ditemukan — RequireAuth belum dipasang")
    }

    cr, ok := v.(scopeClaimsReader)
    // ... parse TenantID, CompanyID, BranchID, WarehouseID dari claims
    return s, nil
}
```

### Konfigurasi Mode Multi

```env
TENANT_MODE=multi

# TENANT_ID dan COMPANY_ID TIDAK perlu diset — datang dari JWT
# Hanya NATS dan infrastruktur lain yang perlu dikonfigurasi
DATABASE_URL=...
REDIS_URL=...
JWT_SECRET=your-secret-min-32-chars
```

---

## Scope Struct

```go
// pkg/scope/scope.go

type Scope struct {
    TenantID    uuid.UUID  // selalu ada — L1
    CompanyID   uuid.UUID  // ada setelah select-scope — L2
    BranchID    *uuid.UUID // opsional — L3; nil jika tidak relevan
    WarehouseID *uuid.UUID // opsional — L4
}

// IsValid memeriksa bahwa TenantID dan CompanyID sudah di-set.
func (s Scope) IsValid() error { ... }

// HasBranch mengembalikan true jika BranchID ter-set.
func (s Scope) HasBranch() bool { ... }

// HasWarehouse mengembalikan true jika WarehouseID ter-set.
func (s Scope) HasWarehouse() bool { ... }
```

---

## Middleware Chain

### Urutan Middleware (server.go)

```go
r.Group(func(r chi.Router) {
    // 1. RequireAuth: validasi JWT signature + expiry
    //    Menyimpan claims ke context dengan key middleware.ContextKeyClaims
    r.Use(middleware.RequireAuth(jwtSvc))

    // 2. ResolveScope: fail-OPEN (tidak reject jika gagal)
    //    Mencoba extract Scope dari request, inject ke context
    //    Single mode: baca dari config
    //    Multi mode: baca dari JWT claims
    r.Use(scope.ResolveScope(scopeResolver))

    // 3. RequireScope: fail-CLOSED (reject jika level tidak tersedia)
    //    Memastikan Scope yang dibutuhkan tersedia di context
    r.Use(scope.RequireScope(scope.LevelTenant, scope.LevelCompany))

    // 4. Handler: Scope sudah pasti ada di context
    exampleHandler.RegisterRoutes(r)
    vernonRegistry.MountRoutes(r)
})
```

### ResolveScope vs RequireScope

| Middleware | Behavior jika gagal | Kapan dipakai |
|------------|--------------------|--------------| 
| `ResolveScope` | Fail-open: lanjutkan tanpa scope | Selalu — inject scope ke context |
| `RequireScope` | Fail-closed: 403 Forbidden | Setelah `ResolveScope` — enforce level |

### RequireScope per Route Group

Untuk route yang butuh level khusus (misal: branch-specific):

```go
r.Group(func(r chi.Router) {
    r.Use(middleware.RequireAuth(jwtSvc))
    r.Use(scope.ResolveScope(scopeResolver))

    // Route yang hanya butuh tenant + company (default)
    r.With(scope.RequireScope(scope.LevelTenant, scope.LevelCompany)).
        Mount("/api/v1/products", productRoutes)

    // Route yang butuh branch juga
    r.With(scope.RequireScope(scope.LevelTenant, scope.LevelCompany, scope.LevelBranch)).
        Mount("/api/v1/inventory", inventoryRoutes)

    // Route yang butuh warehouse (paling ketat)
    r.With(scope.RequireScope(scope.LevelTenant, scope.LevelCompany, scope.LevelBranch, scope.LevelWarehouse)).
        Get("/api/v1/stock-opname", stockOpnameHandler)
})
```

---

## Repository Pattern dengan Scope

### Cara yang Benar

```go
// internal/command/create_example/handler.go

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
    // Application layer membaca Scope dari context
    s, ok := scope.ScopeFromContext(ctx)
    if !ok {
        return scope.ErrMissingTenant
    }
    if err := s.IsValid(); err != nil {
        return err
    }

    // Scope diteruskan sebagai parameter EKSPLISIT ke repository
    if err := h.repo.Save(ctx, s, entity); err != nil {
        return fmt.Errorf("save: %w", err)
    }
    return nil
}

// infrastructure/database/example_repository.go

// Repo menerima Scope sebagai parameter — tidak baca dari context
func (r *ExampleRepository) Save(ctx context.Context, s scope.Scope, e *example.Example) error {
    const q = `
        INSERT INTO examples (id, tenant_id, company_id, ...)
        VALUES ($1, $2, $3, ...)`

    _, err := r.db.ExecContext(ctx, q,
        e.ID, s.TenantID, s.CompanyID, // inject dari parameter
        ...
    )
    return err
}
```

### Semua Query Harus Filter Scope

```go
// Benar: semua query filter tenant_id + company_id
const q = `
    SELECT id, name FROM examples
    WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`

r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id)
```

---

## Anti-Patterns yang Dilarang

### 1. Repository Membaca Scope dari Context

```go
// DILARANG — hidden dependency, sulit di-mock, sulit ditest
func (r *Repo) List(ctx context.Context) ([]*E, error) {
    s, _ := scope.ScopeFromContext(ctx)  // ← JANGAN
    q := `SELECT * FROM t WHERE tenant_id = $1`
    ...
}

// BENAR — scope sebagai parameter eksplisit
func (r *Repo) List(ctx context.Context, s scope.Scope, ...) ([]*E, error) {
    q := `SELECT * FROM t WHERE tenant_id = $1 AND company_id = $2`
    ...
}
```

### 2. Query Tanpa Filter Scope

```go
// DILARANG — cross-tenant data leak
const q = `SELECT * FROM examples WHERE id = $1`

// BENAR
const q = `SELECT * FROM examples WHERE tenant_id = $1 AND company_id = $2 AND id = $3`
```

### 3. TenantID/CompanyID di Entity Struct

```go
// DILARANG — tenancy adalah infrastruktur concern, bukan domain concern
type Example struct {
    ID        uuid.UUID
    TenantID  uuid.UUID  // ← JANGAN
    CompanyID uuid.UUID  // ← JANGAN
    Name      string
}

// BENAR — entity tidak punya pengetahuan tentang multi-tenancy
type Example struct {
    ID   uuid.UUID
    Name string
}
// TenantID/CompanyID hanya ada di SQL query, diinject dari Scope
```

### 4. Hardcode Scope di Handler

```go
// DILARANG — scope tidak boleh hardcode
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
    s := scope.Scope{
        TenantID:  uuid.MustParse("00000000-..."),  // ← JANGAN
        CompanyID: uuid.MustParse("00000000-..."),
    }
    ...
}

// BENAR — selalu baca dari context
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
    s, ok := scope.ScopeFromContext(ctx)
    if !ok {
        return scope.ErrMissingTenant
    }
    ...
}
```

### 5. Menaruh Scope dalam _data Vernon

```go
// DILARANG — tenant_id dan company_id harus kolom terpisah, BUKAN di _data
data := map[string]any{
    "name":       "Gudang A",
    "tenant_id":  tenantID.String(),  // ← JANGAN
    "company_id": companyID.String(), // ← JANGAN
}

// BENAR — tenant_id dan company_id adalah kolom terpisah di tabel Vernon
// BaseRepository.Create() otomatis inject dari Scope
```

---

## Testing dengan Scope

Untuk unit test, buat scope eksplisit:

```go
func TestMyHandler(t *testing.T) {
    testScope := scope.Scope{
        TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000001"),
        CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
    }

    ctx := scope.WithScope(context.Background(), testScope)

    // Jalankan handler atau repo dengan ctx yang sudah punya scope
    result, err := handler.Handle(ctx, cmd)
    ...
}
```

Untuk test isolasi scope (cross-tenant access harus gagal):

```go
func TestScopeIsolation(t *testing.T) {
    repo := database.NewExampleRepository(testDB)
    ctx := context.Background()

    // Simpan dengan scope A
    scopeA := scope.Scope{TenantID: uuidA, CompanyID: uuidAA}
    e := &example.Example{ID: uuid.New(), Name: "Test"}
    repo.Save(ctx, scopeA, e)

    // Coba baca dengan scope B — harus ErrNotFound
    scopeB := scope.Scope{TenantID: uuidB, CompanyID: uuidBB}
    _, err := repo.GetByID(ctx, scopeB, e.ID)
    if err != example.ErrNotFound {
        t.Errorf("cross-scope access harus ErrNotFound, got: %v", err)
    }
}
```
