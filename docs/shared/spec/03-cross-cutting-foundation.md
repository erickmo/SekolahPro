# 03 — Cross-Cutting Foundation

**Status**: Draft  
**Perspective**: System Engineer

## ADR References

- ADR-001 Go Clean Architecture + CQRS
- ADR-002 Vernon Denormalized Read-Cache
- ADR-003 Hybrid CQRS + Vernon
- ADR-004 Multi-Tenant 4-Level Hierarchy + JWT
- ADR-005 Event Bus
- ADR-006 Uber FX DI
- ADR-007 UUID v7
- ADR-009 Dual-Mode Institution Type
- ADR-013 Users & Roles
- ADR-014 Sync Engine Strategy
- ADR-016 Deployment & CI/CD
- ADR-017 Data Migration & Import
- ADR-018 Regulatory Compliance

## Architecture Baseline

### Write/Read Pattern
- gunakan CQRS untuk domain dengan consistency kuat dan kompleksitas moderat
- gunakan Vernon untuk domain read-heavy dengan autoload relationship dan query fleksibel
- pemilihan domain mengikuti decision criteria ADR-003

### Service / Module Boundary
- satu aggregate memiliki satu write owner
- tidak boleh ada direct cross-domain table write
- delivery layer hanya orchestration ringan
- business rules berada di domain/application layer

## Tenancy and Scope

- semua object harus memiliki definisi scope yang eksplisit
- enforce tenant isolation sebagai default
- JWT harus membawa context autentikasi dan scope yang dibutuhkan
- support access lintas tenant harus exceptional dan auditable

## Security and Privacy

- data siswa termasuk high-sensitivity data
- RBAC wajib untuk seluruh capability operasional
- audit trail wajib untuk approval, finance, payroll, dan perubahan data kritikal
- retention, consent, dan breach handling harus mengacu ke compliance spec lanjutan

## Data and Consistency

- UUID v7 untuk primary key
- soft delete dan audit columns menjadi standar
- eventual consistency harus eksplisit; tidak boleh implisit
- domain financial dan identity harus menuliskan consistency class secara tegas

## Integration Rules

- semua external integration lewat anti-corruption boundary
- provider schema tidak boleh bocor ke core domain
- webhook/event inbound harus bisa diverifikasi, direplay, dan diaudit
- mapping external ID dan internal ID harus eksplisit

## Observability

- structured logs, traces, metrics
- correlation ID harus menembus HTTP, event bus, dan callback external
- monitor sync lag, queue depth, failed jobs, dan tenant-scoped anomalies

## Deployment and Operability

- environment strategy harus konsisten antar local, staging, production
- migration harus safe dan bisa dijalankan bertahap
- backup/restore, replay, dan DR flow harus terdokumentasi sebelum production scale

## Shared Acceptance Baseline

- `AC-AUTH`: tenant scope tidak bisa dibypass
- `AC-DATA`: seluruh write path menjaga ownership dan invariants
- `AC-AUDIT`: perubahan data sensitif tercatat
- `AC-INT`: event dan integration memiliki retry/idempotency policy
- `AC-NFR`: latency, availability, dan sync SLA didefinisikan per capability
