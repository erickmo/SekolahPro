# ADR — Core (Shared Foundation)

Keputusan arsitektur yang berlaku untuk **seluruh subproject** (Management Sekolah & Management Koperasi Sekolah). ADR di sini mencakup stack, pattern, dan infrastruktur dasar.

## Index

### Backend (Go)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-001](./ADR-001-go-clean-architecture-cqrs.md) | Go Clean Architecture + CQRS | Accepted | 2026-04-14 |
| [ADR-002](./ADR-002-vernon-denormalized-read-cache.md) | Vernon Denormalized Read-Cache (_rels/_data JSONB) | Accepted | 2026-04-14 |
| [ADR-003](./ADR-003-hybrid-cqrs-vernon.md) | Hybrid CQRS + Vernon — Decision Criteria | Accepted | 2026-04-14 |
| [ADR-004](./ADR-004-multi-tenant-4level-hierarchy.md) | Multi-Tenant 4-Level Hierarchy + Two-Phase JWT | Accepted | 2026-04-14 |
| [ADR-005](./ADR-005-event-bus-inmemory-nats.md) | Event Bus Abstraction — InMemory / NATS JetStream | Accepted | 2026-04-14 |
| [ADR-006](./ADR-006-uber-fx-dependency-injection.md) | Uber FX sebagai Dependency Injection Container | Accepted | 2026-04-14 |
| [ADR-007](./ADR-007-uuid-v7-primary-key.md) | UUID v7 sebagai Primary Key | Accepted | 2026-04-14 |

### Frontend (React)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-008](./ADR-008-react-vite-css-modules.md) | React 18 + Vite + CSS Modules | Accepted | 2026-04-14 |

### Foundation Domains (Shared Master Data)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-010](./ADR-010-academic-years.md) | Academic Years (Tahun Ajaran) | Approved | 2026-04-15 |
| [ADR-011](./ADR-011-class-rooms.md) | Class Rooms (Kelas & Rombel) | Approved | 2026-04-15 |
| [ADR-012](./ADR-012-teachers-staff.md) | Teachers & Staff (Guru & Tenaga Kependidikan) | Approved | 2026-04-15 |
| [ADR-013](./ADR-013-users-roles.md) | Users & Roles (Pengguna & Hak Akses) | Approved | 2026-04-15 |

### Cross-Cutting

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-009](./ADR-009-dual-mode-institution-type.md) | Dual-Mode Institution Type (General / Islamic) | Accepted | 2026-04-15 |
