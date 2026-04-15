# ADR-006: Uber FX sebagai Dependency Injection Container

**Status**: Accepted
**Date**: 2026-04-14
**Deciders**: Vernon Corp Engineering Team

## Context

Go tidak memiliki built-in dependency injection framework seperti Spring (Java) atau .NET DI. Tim perlu memilih antara beberapa pendekatan untuk mengelola dependency:

1. **Manual wiring**: Semua dependency di-construct dan di-wire secara manual di `main.go`.
2. **DI library**: Menggunakan library seperti Wire (Google), FX (Uber), atau Dig (Uber).

### Masalah dengan Manual Wiring

Di aplikasi dengan 20+ service, manual wiring di `main.go` menjadi:

```go
// main.go tanpa DI framework — bisa menjadi 200+ baris
func main() {
    cfg := config.Load()
    
    db := database.NewPostgres(cfg.DatabaseURL)
    redis := cache.NewRedis(cfg.RedisURL)
    logger := logging.NewZerolog(cfg.LogLevel)
    
    categoryRepo := repository.NewCategoryPostgres(db)
    supplierRepo := repository.NewSupplierPostgres(db)
    warehouseRepo := repository.NewWarehousePostgres(db)
    
    createProductHandler := command.NewCreateProductHandler(
        repository.NewProductPostgres(db),
        categoryRepo,
        supplierRepo,
        warehouseRepo,
        eventBus,
    )
    
    // ... 50+ konstruktor lagi
    
    router := chi.NewRouter()
    // ... setup routes
    
    srv := &http.Server{Handler: router, Addr: cfg.Port}
    srv.ListenAndServe()
    
    // Graceful shutdown? Perlu manual handle setiap resource
}
```

Masalah:
- Urutan initialization manual — salah urutan = nil pointer panic
- Tidak ada compile-time validation dependency graph
- Graceful shutdown perlu di-manage manual per resource
- Sulit untuk swap implementasi (misal: swap InMemory bus ↔ NATS bus)

### Requirements DI Solution

- Lifecycle management (OnStart/OnStop hooks) untuk graceful startup dan shutdown
- Automatic dependency resolution — deklarasikan apa yang dibutuhkan, framework yang wire
- Modular — bisa grup dependencies per domain/layer ke dalam modul
- Compile-time safety — panic jika dependency graph invalid, sebelum server start
- Testable — bisa inject mock/test implementations

## Decision

Mengadopsi **Uber FX** sebagai DI container untuk seluruh Go service.

### Kenapa FX bukan Wire (Google)?

| Aspek | Wire (Google) | FX (Uber) |
|-------|---------------|-----------|
| Approach | Code generation | Runtime reflection |
| Lifecycle | Manual | Built-in (OnStart/OnStop) |
| Module system | Manual grouping | `fx.Module()` built-in |
| Error messages | Good | Excellent (visual graph) |
| Boilerplate | Lebih banyak (gen files) | Minimal |
| Runtime overhead | Minimal (generated) | Minimal (reflection hanya saat start) |

FX lebih cocok untuk proyek ini karena built-in lifecycle management dan module system yang align dengan clean architecture layering.

### Module Structure

```go
// infrastructure/fx/modules.go

// DatabaseModule menyediakan *sqlx.DB dan *sqlc.Queries
var DatabaseModule = fx.Module("database",
    fx.Provide(
        database.NewPostgres,      // *sqlx.DB
        database.NewSQLCQueries,   // *sqlc.Queries
        database.NewMigrator,      // untuk auto-migrate saat startup
    ),
)

// CacheModule menyediakan Redis client
var CacheModule = fx.Module("cache",
    fx.Provide(cache.NewRedis),
)

// MessagingModule menyediakan EventBus (InMemory atau NATS)
var MessagingModule = fx.Module("messaging",
    fx.Provide(newEventBus),  // switch berdasarkan config
)

// TelemetryModule menyediakan tracer dan meter
var TelemetryModule = fx.Module("telemetry",
    fx.Provide(
        telemetry.NewTracer,
        telemetry.NewMeter,
        telemetry.NewPrometheusExporter,
    ),
)

// ProductModule menyediakan semua dependencies untuk domain product
var ProductModule = fx.Module("product",
    fx.Provide(
        // Repositories
        repository.NewProductPostgres,
        repository.NewProductVernonReader,
        // Command handlers
        command.NewCreateProductHandler,
        command.NewUpdateProductHandler,
        command.NewDeleteProductHandler,
        // Query handlers
        query.NewGetProductHandler,
        query.NewListProductsHandler,
        // HTTP handlers
        handler.NewProductHandler,
        // SyncEngine handlers
        sync.NewProductSyncHandler,
    ),
)
```

### Application Bootstrap

```go
// cmd/api/main.go
func main() {
    app := fx.New(
        // Infrastructure
        fx.Provide(config.Load),
        fx.Provide(logging.NewZerolog),
        DatabaseModule,
        CacheModule,
        MessagingModule,
        TelemetryModule,
        
        // Domain modules
        AuthModule,
        UserModule,
        ProductModule,
        CategoryModule,
        SupplierModule,
        WarehouseModule,
        OrderModule,
        InventoryModule,
        
        // HTTP server
        fx.Provide(router.NewChi),
        fx.Invoke(router.RegisterRoutes),
        fx.Invoke(startHTTPServer),
    )
    
    app.Run()  // blocks, handles SIGTERM untuk graceful shutdown
}
```

### Lifecycle Management

```go
// infrastructure/database/postgres.go
func NewPostgres(lc fx.Lifecycle, cfg *Config, logger zerolog.Logger) (*sqlx.DB, error) {
    db, err := sqlx.Connect("pgx", cfg.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("connect postgres: %w", err)
    }
    
    db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
    db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)
    
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            logger.Info().Msg("postgres connection established")
            return db.PingContext(ctx)  // verify connection
        },
        OnStop: func(ctx context.Context) error {
            logger.Info().Msg("closing postgres connection")
            return db.Close()
        },
    })
    
    return db, nil
}

// infrastructure/messaging/nats/bus.go
func NewNATSBus(lc fx.Lifecycle, cfg *Config) (*NATSBus, error) {
    bus := &NATSBus{}
    
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            conn, err := nats.Connect(cfg.NATSUrl)
            if err != nil {
                return err
            }
            bus.nc = conn
            bus.js, _ = conn.JetStream()
            return SetupStreams(bus.js)
        },
        OnStop: func(ctx context.Context) error {
            bus.nc.Drain()  // flush pending messages
            return nil
        },
    })
    
    return bus, nil
}
```

### Dependency Declaration Pattern

```go
// usecase/command/create_product/handler.go
type CreateProductDeps struct {
    fx.In  // FX marker — field ini di-inject oleh FX
    
    Repo         ProductRepository
    CategoryRepo CategoryRepository
    SupplierRepo SupplierRepository
    EventBus     port.EventBus
    Logger       zerolog.Logger
    Tracer       trace.Tracer
}

func NewCreateProductHandler(deps CreateProductDeps) *CreateProductHandler {
    return &CreateProductHandler{
        repo:         deps.Repo,
        categoryRepo: deps.CategoryRepo,
        supplierRepo: deps.SupplierRepo,
        eventBus:     deps.EventBus,
        logger:       deps.Logger,
        tracer:       deps.Tracer,
    }
}
```

### FX Module Annotation (untuk multiple implementations)

```go
// Ketika ada dua implementasi untuk interface yang sama
fx.Provide(
    fx.Annotate(
        repository.NewProductPostgres,
        fx.As(new(port.ProductRepository)),
    ),
    fx.Annotate(
        repository.NewProductVernonReader,
        fx.As(new(port.ProductReader)),
    ),
)
```

## Consequences

### Positive

- **Graceful shutdown otomatis**: FX mengelola urutan shutdown — HTTP server stop dulu, lalu NATS drain, lalu DB close. Tidak ada resource leak.
- **Fail-fast startup**: Jika dependency graph invalid (misal: ada circular dependency atau missing provider), aplikasi panic dengan pesan error yang jelas sebelum menerima request pertama.
- **Module isolation**: Domain module dapat di-develop dan di-test secara independen dengan menyediakan mock implementations.
- **Zero runtime cost setelah startup**: FX menggunakan reflection hanya saat bootstrap — setelah `app.Run()`, tidak ada overhead reflection.
- **Testability**: Test bisa membuat FX app kecil dengan hanya subset module yang dibutuhkan.
- **Dependency graph visualization**: `fx.New(..., fx.WithLogger(...))` menampilkan dependency graph yang dapat diaudit.

### Negative / Trade-offs

- **Runtime reflection**: Error dependency missing baru diketahui saat aplikasi start, bukan saat compile. Berbeda dengan Wire yang 100% compile-time.
- **Magic injection**: Developer perlu memahami `fx.In`, `fx.Out`, dan `fx.Annotate` — ada learning curve.
- **Debugging complexity**: Stack trace saat FX error bisa panjang dan sulit dibaca bagi developer yang belum familiar.
- **Import cycles**: Jika dua module saling import, FX tidak bisa resolve — ini forcing function yang baik, tapi bisa frustrating saat refactor.
- **Overhead pada test**: Membuat FX app untuk setiap test suite ada overhead. Test murni unit test sebaiknya tidak menggunakan FX sama sekali.

## Alternatives Considered

### 1. Manual Wiring (no framework)
- Zero learning curve, full control
- Ditolak karena `main.go` akan menjadi 500+ baris saat project grow, tidak ada lifecycle management standar, dan error urutan initialization sulit dideteksi.

### 2. Wire (Google)
- Compile-time safety, zero runtime reflection
- Ditolak karena tidak ada built-in lifecycle management (harus manual implement Cleanup functions), lebih banyak boilerplate (generated `wire_gen.go`), dan setup lebih kompleks.

### 3. Dig (Uber, FX tanpa lifecycle)
- Lebih ringan dari FX, FX di-build di atas Dig
- Ditolak karena lifecycle management (OnStart/OnStop) adalah requirement utama — lebih baik langsung pakai FX yang sudah include ini.

### 4. Samber/do (modern Go DI)
- Simpler API dari FX
- Ditolak karena ekosistem lebih kecil, dokumentasi lebih terbatas, dan less battle-tested di production dibanding FX yang dipakai Uber di scale besar.
