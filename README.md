# SekolahPro

Sistem manajemen sekolah yang mencakup pengelolaan stakeholder (siswa, guru, orang tua, staf) dan kerja sama antar sekolah. Dibangun dengan arsitektur hybrid CQRS + Vernon untuk performa optimal pada operasi read-heavy dan fleksibilitas domain yang kompleks.

## Arsitektur

Proyek ini menggunakan **monorepo** dengan dua komponen utama:

| Komponen | Stack | Deskripsi |
|---|---|---|
| `api/` | Go 1.25, Chi, PostgreSQL, Redis, NATS | Backend REST API dengan CQRS + Vernon pattern |
| `web-dashboard/` | React 18, TypeScript, Vite, Zustand | Dashboard admin untuk manajemen sekolah |

### Hybrid Architecture: CQRS + Vernon

Dua pola domain tersedia, dipilih berdasarkan karakteristik domain:

| Kriteria | CQRS | Vernon (`_rels`/`_data`) |
|---|---|---|
| Jumlah relasi/JOIN | Sedikit (<=2) | Banyak (3+) |
| Read:Write ratio | Seimbang | Read-heavy (10:1+) |
| Logic bisnis | Kompleks (workflow, saga) | Sederhana (CRUD) |
| Konsistensi | Strong consistency | Eventual consistency OK |

### Multi-Tenant Support

Mendukung dua mode operasi:

- **Single-tenant** — untuk deployment dedicated/self-hosted per sekolah
- **Multi-tenant** — untuk platform SaaS yang melayani banyak sekolah

Hierarki 4 level: `Tenant -> Company -> Branch -> Warehouse`

## Tech Stack

### Backend (Go)

- **Router**: Chi v5
- **DI**: Uber FX
- **Database**: PostgreSQL 17 + sqlx + sqlc
- **Cache**: Redis 7
- **Event Bus**: Watermill (InMemory dev / NATS JetStream prod)
- **Tracing**: OpenTelemetry + Jaeger
- **Metrics**: Prometheus
- **Auth**: JWT (golang-jwt)
- **Logging**: zerolog

### Frontend (React)

- **Build Tool**: Vite 8
- **State Management**: Zustand 5 (client) + TanStack React Query 5 (server)
- **Routing**: React Router 6
- **Styling**: CSS Modules (camelCase)
- **Testing**: Vitest + React Testing Library + MSW + Playwright

### Infrastructure

- PostgreSQL 17
- Redis 7
- NATS 2.10 (JetStream)
- Jaeger (distributed tracing)
- Prometheus (metrics)

## Prasyarat

- Go 1.25+
- Node.js 18+ dan npm
- Docker dan Docker Compose
- Make

## Quick Start

### 1. Clone dan Setup

```bash
git clone <repository-url>
cd boilerplate
```

### 2. Jalankan Infrastructure

```bash
cd api
docker compose up -d
```

Ini akan menjalankan PostgreSQL, Redis, NATS, Jaeger, dan Prometheus.

### 3. Jalankan Backend

```bash
cd api
cp .env.example .env       # Sesuaikan konfigurasi
make migrate-up            # Jalankan database migration
make air                   # Jalankan dengan hot reload
```

API akan tersedia di `http://localhost:8080`.

### 4. Jalankan Frontend

```bash
cd web-dashboard
cp .env.example .env.local  # Sesuaikan konfigurasi
npm install
npm run dev
```

Dashboard akan tersedia di `http://localhost:5173`.

## Perintah Umum

### Backend

```bash
make dev                    # Build + run (tanpa hot reload)
make air                    # Run dengan hot reload
make test                   # Unit tests
make test-integration       # Integration tests (butuh Docker)
make migrate-up             # Jalankan migrasi
make migrate-create name=X  # Buat migration baru
make infra-up               # Start Docker services
make sqlc                   # Generate code dari SQL
make swagger                # Generate Swagger docs
make tools-install          # Install semua dev tools
```

### Frontend

```bash
npm run dev                 # Development server
npm run build               # Production build
npm run test                # Unit tests
npm run test:e2e            # E2E tests (Playwright)
npm run lint                # Linting
```

## Struktur Proyek

```
boilerplate/
├── api/                          # Backend Go
│   ├── cmd/api/                  # Entry point
│   ├── infrastructure/           # Config, database, cache, telemetry
│   ├── internal/
│   │   ├── domain/               # Entity, errors, events, repository interfaces
│   │   ├── command/              # Write operations (CQRS)
│   │   ├── query/                # Read operations (CQRS)
│   │   ├── eventhandler/         # Cross-domain event handlers
│   │   └── delivery/http/        # HTTP handlers
│   ├── pkg/                      # Shared packages (vernon, eventbus, jwt, middleware)
│   ├── migrations/               # SQL migrations
│   ├── sqlc/                     # sqlc queries & schema
│   └── tests/                    # Integration tests
│
├── web-dashboard/                # Frontend React
│   └── src/
│       ├── app/                  # Router, providers, guards
│       ├── pages/                # Page components
│       ├── services/             # API client & entity services
│       ├── stores/               # Zustand stores
│       ├── hooks/                # Custom hooks
│       ├── layouts/              # AppShell, Navbar
│       ├── theme/                # CSS variables & reset
│       ├── types/                # TypeScript types
│       └── widgets/              # Reusable UI components
│
└── docs/                         # Dokumentasi
    ├── adr/                      # Architecture Decision Records
    ├── developer/                # Panduan developer
    └── user-manual/              # Panduan pengguna
```

## Dokumentasi

| Dokumen | Deskripsi |
|---|---|
| [Developer Guide](docs/developer/) | Setup, arsitektur, menambah domain, testing, konfigurasi |
| [User Manual](docs/user-manual/) | Panduan penggunaan, deployment, monitoring |
| [ADR](docs/adr/) | Architecture Decision Records |
| [API Swagger](http://localhost:8080/swagger/index.html) | API documentation (jalankan server terlebih dahulu) |

## Konfigurasi Environment

### Backend (`api/.env`)

| Variable | Deskripsi | Default |
|---|---|---|
| `DATABASE_URL` | Koneksi PostgreSQL | - |
| `REDIS_URL` | Koneksi Redis | - |
| `JWT_SECRET` | Secret key (min 32 karakter) | - |
| `TENANT_MODE` | `single` atau `multi` | `single` |
| `TENANT_ID` | UUID tenant (wajib jika single) | - |
| `COMPANY_ID` | UUID company (wajib jika single) | - |

### Frontend (`web-dashboard/.env.local`)

| Variable | Deskripsi | Default |
|---|---|---|
| `VITE_API_BASE_URL` | URL backend API | `http://localhost:8080` |
| `VITE_APP_NAME` | Nama aplikasi | `Dashboard` |
| `VITE_MULTI_TENANT` | Mode multi-tenant | `false` |

## Lisensi

Proprietary - Hak cipta dilindungi.
