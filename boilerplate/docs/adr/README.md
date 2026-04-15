# Architecture Decision Records (ADR)

Direktori ini berisi Architecture Decision Records untuk project **Vernon Corp Go + React Boilerplate**.

ADR adalah dokumen yang merekam keputusan arsitektur penting beserta konteks, alasan, dan konsekuensinya. Tujuannya adalah memberikan historical context bagi anggota tim yang bergabung di kemudian hari, dan mencegah re-debating keputusan yang sudah dibuat.

## Format

Setiap ADR menggunakan format standar:

```
# ADR-XXX: [Judul]

Status:     Accepted | Proposed | Deprecated | Superseded by ADR-YYY
Date:       YYYY-MM-DD
Deciders:   [Tim / individu yang membuat keputusan]

## Context
## Decision
## Consequences
## Alternatives Considered
```

## Status Definitions

| Status | Arti |
|--------|------|
| **Proposed** | Sedang dalam diskusi, belum final |
| **Accepted** | Keputusan sudah dibuat dan aktif diterapkan |
| **Deprecated** | Masih berlaku tapi tidak direkomendasikan untuk penggunaan baru |
| **Superseded** | Digantikan oleh ADR lain (selalu sertakan nomor ADR pengganti) |

---

## Index ADR

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

---

## Ringkasan Keputusan Utama

### Arsitektur Go

```
Clean Architecture (4 layers)
│
├── domain/         Pure business entities, no framework dependency
├── usecase/        Application logic, CQRS (command/ + query/)
├── adapter/        HTTP handlers, repositories, event handlers
└── infrastructure/ DB, cache, messaging, telemetry
```

### Hybrid Pattern Decision Tree

```
Domain baru → Analisis karakteristik
│
├── ≤ 2 JOIN && write-balanced && strong consistency?
│   └── CQRS + sqlc (Standard)
│       Example: auth, users, permissions
│
└── ≥ 3 JOIN || read:write ≥ 10:1 || latency < 100ms?
    └── Vernon Pattern (_rels + _data + SyncEngine)
        Example: products, orders, inventory
```

### Multi-Tenant Hierarchy

```
Tenant → Company → Branch → Warehouse
  │
  ├── SaaS mode:  Two-Phase JWT (Auth → Tenant Selection → Operational)
  └── Self-hosted: Single-Phase JWT (Auth → Operational)
```

### Event Bus

```
USE_NATS=false → InMemory (dev/test, no Docker required)
USE_NATS=true  → NATS JetStream (prod, durable, at-least-once)
```

### Primary Key

```
UUID v7 via uuid_generate_v7()
├── Time-sortable (48-bit timestamp prefix)
├── Monotonically increasing (B-tree friendly)
└── Globally unique (no DB round-trip needed)
```

### Frontend Stack

```
React 18 + TypeScript (strict)
├── Build:   Vite (native ESM, < 500ms cold start)
├── Style:   CSS Modules + CSS Variables (design tokens)
├── State:   Zustand (global UI) + TanStack Query (server state)
└── Test:    Vitest + MSW + Playwright
```

---

## Cara Membuat ADR Baru

1. Copy template dari ADR yang sudah ada
2. Beri nomor berikutnya (`ADR-009`, dst)
3. Isi semua section: Context, Decision, Consequences, Alternatives Considered
4. Update tabel index di README ini
5. Minta review dari minimal satu engineer lain
6. Set status ke `Accepted` setelah disetujui

## Panduan Kapan Membuat ADR

Buat ADR ketika memilih atau mengubah:
- Framework atau library utama
- Database schema pattern
- Authentication / authorization approach
- Communication pattern (sync vs async, REST vs gRPC)
- Deployment architecture
- Testing strategy
- Performance optimization technique yang berdampak pada desain

**Tidak perlu ADR untuk**: keputusan implementasi detail yang bisa diubah tanpa dampak arsitektur (nama variabel, utility function, minor refactor).

---

*ADR template ini mengikuti [Michael Nygard's ADR format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).*
