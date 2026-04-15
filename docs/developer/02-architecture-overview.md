# 02 — Architecture Overview

## System Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                                    │
│                                                                         │
│   ┌────────────────────────┐    ┌──────────────────────────────────┐   │
│   │  React Web Dashboard   │    │   Mobile App / Third-Party Client │   │
│   │  (Vite + TypeScript)   │    │                                  │   │
│   └──────────┬─────────────┘    └──────────────┬───────────────────┘   │
└──────────────┼───────────────────────────────────┼─────────────────────┘
               │  HTTPS + JWT Bearer               │
               ▼                                   ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           Go API                                        │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                    HTTP Layer (Chi Router)                        │   │
│  │  RequestID → Logger → Tracer → Recoverer → Timeout              │   │
│  │  RequireAuth → ResolveScope → RequireScope → Handler            │   │
│  └────────────────────────────┬─────────────────────────────────────┘   │
│                               │                                         │
│          ┌────────────────────┴────────────────────┐                    │
│          │                                         │                    │
│   ┌──────▼──────────┐                  ┌───────────▼─────────────┐     │
│   │   CQRS Pattern  │                  │    Vernon Pattern        │     │
│   │                 │                  │                          │     │
│   │  CommandBus     │                  │  Registry               │     │
│   │  ├─ create_*    │                  │  ├─ product_categories   │     │
│   │  ├─ update_*    │                  │  ├─ products             │     │
│   │  └─ delete_*    │                  │  └─ ...                 │     │
│   │                 │                  │                          │     │
│   │  QueryBus       │                  │  BaseService            │     │
│   │  ├─ get_*_by_id │                  │  BaseRepository         │     │
│   │  └─ list_*      │                  │  BaseHandler            │     │
│   └──────┬──────────┘                  └───────────┬─────────────┘     │
│          │                                         │                    │
│          └────────────────────┬────────────────────┘                    │
│                               │                                         │
│  ┌────────────────────────────▼─────────────────────────────────────┐   │
│  │                      EventBus (Watermill)                         │   │
│  │      InMemory (dev/test) ←→ NATS JetStream (production)          │   │
│  │      Topics: example.created, example.updated, sync.*            │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                    Infrastructure Layer                           │   │
│  │  PostgreSQL (sqlx)   Redis (cache)   OpenTelemetry (OTel)        │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
               │               │              │
               ▼               ▼              ▼
          PostgreSQL         Redis          NATS
          (port 5432)     (port 6379)   (port 4222)

                    Observability
                    ┌──────────────────────┐
                    │  Jaeger (traces)     │ :16686
                    │  Prometheus (metrics)│ :9090
                    └──────────────────────┘
```

---

## Go API Layer Diagram

```
┌──────────────────────────────────────────────────────┐
│                  cmd/api/ (Entry Point)               │
│  main.go — Uber FX wiring                            │
│  server.go — Chi router + middleware chain            │
└──────────────────────────┬───────────────────────────┘
                           │ DI via Uber FX
           ┌───────────────┼───────────────┐
           │               │               │
           ▼               ▼               ▼
┌──────────────────┐  ┌──────────────┐  ┌─────────────────┐
│  DELIVERY LAYER  │  │ CQRS BUSES   │  │  VERNON REGISTRY │
│  internal/       │  │  pkg/        │  │  pkg/vernon/     │
│  delivery/http/  │  │  commandbus/ │  │  pkg/vernonsync/ │
│                  │  │  querybus/   │  │                  │
│  ExampleHandler  │  │  eventbus/   │  │  Registry        │
│  (decode, route, │  │              │  │  BaseService     │
│   encode only)   │  │  CommandBus  │  │  BaseRepository  │
│                  │  │  QueryBus    │  │  BaseHandler     │
└─────────┬────────┘  └──────┬───────┘  └────────┬────────┘
          │                  │                   │
          │ dispatch         │ dispatch          │ CRUD ops
          ▼                  ▼                   ▼
┌──────────────────────────────────────────────────────┐
│              APPLICATION LAYER                       │
│  internal/command/     internal/query/               │
│  ├─ create_example/    ├─ get_example_by_id/         │
│  ├─ update_example/    └─ list_examples/             │
│  └─ delete_example/                                  │
│                                                      │
│  Reads Scope from ctx, passes as explicit param      │
└─────────────────────────┬────────────────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────────────┐
│                   DOMAIN LAYER                       │
│  internal/domain/{name}/                             │
│  ├─ entity.go     — pure Go structs                  │
│  ├─ errors.go     — domain errors (sentinel)         │
│  ├─ events.go     — domain events (immutable)        │
│  └─ (repository interfaces declared here)            │
│                                                      │
│  ZERO external dependencies — no framework imports   │
└─────────────────────────┬────────────────────────────┘
                          │ implements interface
                          ▼
┌──────────────────────────────────────────────────────┐
│               INFRASTRUCTURE LAYER                   │
│  infrastructure/database/                            │
│  ├─ db.go                  — sqlx connection         │
│  └─ example_repository.go  — concrete implementation │
│                                                      │
│  infrastructure/config/    — Viper config loader     │
│  infrastructure/cache/     — Redis cache             │
│  infrastructure/telemetry/ — OTel initialization     │
└──────────────────────────────────────────────────────┘
```

---

## Layer Responsibility Table

| Layer | Package | Boleh Import | Dilarang Import |
|-------|---------|-------------|-----------------|
| Entry Point | `cmd/api/` | semua | - |
| Delivery | `internal/delivery/http/` | `pkg/*`, `internal/command/`, `internal/query/` | `infrastructure/`, `internal/domain/` langsung |
| Application (Command) | `internal/command/*/` | `internal/domain/`, `pkg/scope`, `pkg/eventbus` | `infrastructure/`, `delivery/` |
| Application (Query) | `internal/query/*/` | `internal/domain/`, `pkg/scope` | `infrastructure/`, `delivery/`, `pkg/eventbus` |
| Domain | `internal/domain/*/` | `pkg/scope` (pure value object) | **semua package lain** |
| Infrastructure | `infrastructure/database/` | `internal/domain/`, `pkg/scope`, `sqlx` | `internal/command/`, `internal/query/`, `delivery/` |
| Shared Pkg | `pkg/*/` | stdlib, third-party | `internal/*` |

### Aturan Penting

1. **Domain layer adalah yang paling murni** — hanya stdlib + `pkg/scope` (pure value object).
2. **Handler layer hanya decode/dispatch/encode** — tidak ada business logic, validasi domain, atau akses DB langsung.
3. **Repository interface di domain layer** — implementasi di infrastructure layer.
4. **Scope dibaca application layer, diteruskan sebagai parameter** — bukan dibaca di repository.

---

## Hybrid CQRS vs Vernon — Decision Matrix

Gunakan tabel ini untuk memutuskan pola yang tepat saat menambah domain baru:

| Kriteria | CQRS (standard) | Vernon (_rels/_data) |
|----------|-----------------|---------------------|
| **Jumlah JOIN per read** | Sedikit (0–2) | Banyak (3+) |
| **Read:Write ratio** | Seimbang (~1:1) | Read-heavy (10:1+) |
| **Business logic complexity** | Kompleks (workflow, validasi multi-step, saga) | Sederhana (CRUD + validasi field) |
| **Konsistensi data** | Strong consistency dibutuhkan | Eventual consistency dapat diterima |
| **Relasi antar domain** | Independen atau minimal | Banyak belongs_to + autoload |
| **Skema tabel** | Kolom eksplisit per field | 10 kolom wajib + `_data` JSONB |
| **Query flexibility** | SQL penuh via sqlc | Query hanya via `_data` JSONB filter |
| **Auto-generated CRUD** | Tidak — developer tulis sendiri | Ya — `BaseHandler` memberikan CRUD + list |

### Kapan Pilih CQRS

- Domain dengan workflow multi-step (order → payment → fulfillment)
- Domain yang butuh strong consistency (inventory, finansial)
- Query yang butuh aggregasi atau JOIN yang kompleks
- Domain dengan validasi bisnis yang melibatkan state dari banyak tabel

### Kapan Pilih Vernon

- Data referensi (categories, products, locations, currencies)
- Domain yang sering di-embed oleh domain lain via `belongs_to`
- Domain dengan CRUD sederhana dan banyak filter parameter
- Domain di mana performance read lebih penting dari consistency

### Contoh di Boilerplate Ini

```
Example domain   → CQRS     (business logic, events, validasi)
ProductCategory  → Vernon   (data referensi, read-heavy, no complex logic)
Product          → Vernon   (belongs_to category, autoload embedded data)
```

---

## Dependency Injection — Uber FX

```
main.go
│
├── fx.Provide (providers)
│   ├── Infrastructure: Config, Logger, DB, Redis, Cache, EventBus, Telemetry
│   ├── Buses: CommandBus, QueryBus
│   ├── Auth: JWT Service
│   ├── Scope: Resolver (pilih berdasarkan TENANT_MODE)
│   ├── HTTP Handlers: ExampleHandler
│   ├── Vernon: Registry, SyncEngine
│   └── Router: Chi router
│
└── fx.Invoke (lifecycle hooks)
    ├── initTelemetry          → start OTel, register shutdown
    ├── registerCommandHandlers → register semua command handlers ke CommandBus
    ├── registerQueryHandlers   → register semua query handlers ke QueryBus
    ├── registerEventHandlers   → register event handlers + start Watermill router
    ├── registerVernonDomains   → daftarkan descriptor Vernon ke Registry
    ├── startVernonSync         → subscribe SyncEngine ke EventBus topics
    └── startServer             → start HTTP server, register graceful shutdown
```

---

## Vernon Architecture Detail

```
Write Flow:
  HTTP POST /products
  → BaseHandler.Create()
  → BaseService.Create()
       ├── Validate() via Descriptor
       ├── Autoload: fetch linked category data
       ├── Write record (id, tenant_id, company_id, _rels, _data, ...)
       ├── TX COMMIT
       └── Publish SyncEvent "sync.products" via EventBus (AFTER commit)

Sync Flow:
  SyncEvent "sync.product_categories" diterima SyncEngine
  → Cari consumers dari Registry.reverseMap["product_categories"]
     → consumer: products (via belongs_to category, is_autoload=true)
  → Fetch source data (category._data)
  → Update semua products yang _data->>'category_id' = changedCategoryID
     → CAS update _data["category"] = {name, sku}
     → _sync_version incremented

Read Flow:
  HTTP GET /products
  → BaseHandler.List()
  → BaseRepository.FindAll()
  → SELECT id, tenant_id, _data FROM products WHERE tenant_id=$1 AND company_id=$2
  → Return _data langsung — NO JOIN diperlukan
```

---

## Event Flow Diagram

```
Write Operation (command handler)
─────────────────────────────────
│
├── Validate input
├── BEGIN TX
├── INSERT / UPDATE / DELETE row
├── COMMIT TX
└── EventBus.Publish(DomainEvent)
         │
         ├── InMemory mode (dev):
         │   └── goroutine → EventHandler.Handle(event)
         │
         └── NATS mode (prod):
             └── NATS JetStream → at-least-once delivery
                  └── EventHandler.Handle(event)

Vernon Sync (after write)
──────────────────────────
BaseService.Create/Update
  └── Publish SyncEvent{domain, entity_id, action}
            │
            └── SyncEngine.handleSyncEvent()
                     │
                     └── Registry.GetConsumers(domain)
                              │
                              └── For each consumer:
                                   ├── Fetch source _data
                                   ├── Extract fields
                                   └── CASUpdate consumer._data[relName]
```
