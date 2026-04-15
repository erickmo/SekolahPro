# 01 — Getting Started

## Prerequisites

| Tool | Minimum Version | Install |
|------|----------------|---------|
| Go | 1.25+ | https://go.dev/dl/ |
| Node.js | 18+ | https://nodejs.org/ |
| Docker Desktop | latest | https://www.docker.com/products/docker-desktop/ |
| Make | any | built-in di Linux/macOS; `choco install make` di Windows |
| git | 2.x+ | https://git-scm.com/ |

Verifikasi instalasi:

```bash
go version        # go version go1.25.x ...
node --version    # v18.x.x atau lebih baru
docker --version  # Docker version 24.x.x ...
make --version    # GNU Make 4.x.x ...
```

---

## Clone Repository

```bash
git clone <repo-url> boilerplate
cd boilerplate
```

Struktur root setelah clone:

```
boilerplate/
├── api/             # Go API
├── web-dashboard/   # React web app
└── docs/            # Dokumentasi
```

---

## Setup & Run Go API

### 1. Masuk ke direktori API

```bash
cd api
```

### 2. Install dev tools

```bash
make tools-install
```

Tools yang diinstall: `sqlc`, `mockery`, `air` (hot reload), `swag` (Swagger generator).

### 3. Start infrastruktur (Docker)

```bash
make infra-up
```

Perintah ini menjalankan:
- **PostgreSQL 17** di port `5432`
- **Redis 7** di port `6379`
- **NATS 2.10** di port `4222` (messaging, opsional di dev)
- **Jaeger** di port `16686` (tracing UI)
- **Prometheus** di port `9090` (metrics)

Tunggu hingga semua container healthy:

```bash
docker compose ps
```

### 4. Konfigurasi environment

```bash
cp .env.example .env
```

Edit `.env` dan sesuaikan nilai minimum berikut:

```env
# Database (sesuaikan dengan docker-compose.yml)
DATABASE_URL=postgres://postgres:postgres@localhost:5432/boilerplate_db?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379/0

# JWT — WAJIB diganti, minimum 32 karakter
JWT_SECRET=your-super-secret-key-minimum-32-chars

# Tenant mode
TENANT_MODE=single
TENANT_ID=00000000-0000-0000-0000-000000000001
COMPANY_ID=00000000-0000-0000-0000-000000000002
```

> Untuk mendapatkan UUID yang valid: `python3 -c "import uuid; print(uuid.uuid4())"`

### 5. Jalankan migrasi database

```bash
make migrate-up
```

Migrasi dijalankan secara berurutan dari direktori `migrations/`.

### 6. Jalankan API dengan hot reload

```bash
make air
```

Atau tanpa hot reload:

```bash
make dev
```

### 7. Verifikasi API berjalan

```bash
curl http://localhost:8080/health
# {"status":"ok","tenant_mode":"single"}
```

Swagger UI (setelah `make swagger`): http://localhost:8080/swagger/index.html

---

## Setup & Run Web Dashboard

### 1. Masuk ke direktori web-dashboard

```bash
cd ../web-dashboard
```

### 2. Install dependencies

```bash
npm install
```

### 3. Konfigurasi environment

File `.env.local` sudah ada dengan nilai default:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_NAME=Dashboard
VITE_MULTI_TENANT=false
VITE_MOCK_MODE=false
```

Untuk mode multi-tenant:
```env
VITE_MULTI_TENANT=true
```

### 4. Jalankan dev server

```bash
npm run dev
```

Dashboard tersedia di: http://localhost:5173

---

## Struktur Direktori Lengkap

### Go API (`api/`)

```
api/
├── cmd/api/
│   ├── main.go          # Entry point Uber FX — wiring semua dependency
│   └── server.go        # Chi router — middleware chain + route registration
├── infrastructure/
│   ├── config/          # Viper config loader (config.go)
│   ├── database/        # DB connection + repository implementations
│   ├── cache/           # Redis cache wrapper
│   └── telemetry/       # OpenTelemetry initialization
├── internal/
│   ├── domain/
│   │   ├── example/     # Entity, errors, events, repository interfaces
│   │   ├── product/     # Vernon domain descriptor
│   │   └── product_category/  # Vernon domain descriptor
│   ├── command/
│   │   ├── create_example/    # Command handler
│   │   ├── update_example/
│   │   └── delete_example/
│   ├── query/
│   │   ├── get_example_by_id/ # Query handler
│   │   └── list_examples/
│   ├── eventhandler/    # Cross-domain event handlers
│   └── delivery/http/   # HTTP handlers (decode, dispatch, encode)
├── pkg/
│   ├── commandbus/      # Generic type-safe CommandBus + OTel tracing
│   ├── querybus/        # Generic type-safe QueryBus + OTel tracing
│   ├── eventbus/        # EventBus interface (InMemory + NATS JetStream)
│   ├── jwt/             # JWT service (generate/validate)
│   ├── middleware/       # HTTP middleware (auth, logger, tracer, recoverer)
│   ├── pagination/      # Pagination helper
│   ├── scope/           # Scope struct + resolver + middleware
│   ├── tenant/          # TenantID-only helper (legacy)
│   ├── vernon/          # Vernon core (BaseDomain, BaseRepository, BaseService, BaseHandler, Registry)
│   └── vernonsync/      # SyncEngine (subscribe + propagate)
├── migrations/          # SQL migrations (golang-migrate format)
│   ├── 001_create_examples.sql
│   ├── 002_vernon_base.sql
│   ├── 003_vernon_product_categories.sql
│   └── 004_vernon_products.sql
├── sqlc/
│   ├── queries/         # SQL query files untuk sqlc generate
│   └── schema/          # Schema SQL untuk sqlc
├── tests/
│   └── integration/     # Integration tests (Testcontainers)
├── seeds/               # SQL seed data
├── .air.toml            # Konfigurasi air (hot reload)
├── .env.example         # Template environment variables
├── docker-compose.yml   # Docker services
├── Makefile             # Development commands
└── sqlc.yaml            # Konfigurasi sqlc
```

### Web Dashboard (`web-dashboard/src/`)

```
src/
├── config/
│   └── app.config.ts    # VITE_MULTI_TENANT, VITE_APP_NAME, dll
├── app/
│   ├── App.tsx          # Root component
│   ├── routes.tsx       # Route definitions (single + multi tenant)
│   ├── providers.tsx    # QueryClient, ToastProvider
│   └── ProtectedRoute.tsx  # Auth guard + role-based routing
├── hooks/               # Custom React hooks (useForm, useDataSource, dll)
├── layouts/
│   ├── AppShell/        # Layout utama (sidebar + content)
│   ├── AppNavbar/       # Navigation bar
│   └── PageHeader/      # Page title + breadcrumbs
├── pages/               # Page components (Login, Dashboard, dll)
├── services/
│   ├── api.client.ts    # HTTP client dengan JWT interceptor
│   ├── createEntityService.ts  # Factory untuk CRUD services
│   └── auth.service.ts  # Login, logout, refresh token
├── stores/
│   ├── auth.store.ts    # Zustand auth state (user, token, selectedCompany)
│   ├── ui.store.ts      # UI state (sidebar, theme)
│   └── notification.store.ts  # Toast notifications
├── theme/               # CSS variables, reset, typography, motion
├── types/               # TypeScript interfaces
├── utils/               # Helper functions (cn, format, export)
└── widgets/Toast/       # Toast notification system
```

---

## Makefile Commands Reference

```bash
# Di dalam api/

make dev              # Build + run (tanpa hot reload)
make air              # Run dengan hot reload (butuh air terinstall)
make test             # Unit tests (internal/... pkg/...)
make test-race        # Unit tests dengan race detector
make test-integration # Integration tests (butuh Docker)
make lint             # Jalankan golangci-lint
make sqlc             # Generate code dari SQL queries
make mock             # Generate mocks dari domain interfaces
make migrate-up       # Jalankan semua pending migrations
make migrate-down     # Rollback satu migration terakhir
make migrate-create name=nama_migration  # Buat file migration baru
make infra-up         # Start Docker services
make infra-down       # Stop Docker services
make swagger          # Generate Swagger docs ke docs/
make tools-install    # Install dev tools (sqlc, mockery, air, swag)
make tidy             # go mod tidy
make docker-build     # Build Docker image
make reset            # Hard reset (hapus .env + clean build)
```

```bash
# Di dalam web-dashboard/

npm run dev           # Dev server dengan hot reload
npm run build         # Production build
npm run preview       # Preview production build lokal
npm run test          # Vitest unit tests
npm run lint          # ESLint
```

---

## Health Check & Observability

Setelah API berjalan:

| Endpoint | URL | Keterangan |
|----------|-----|-----------|
| Health | http://localhost:8080/health | Status API + tenant mode |
| Metrics | http://localhost:8080/metrics | Prometheus metrics |
| Swagger UI | http://localhost:8080/swagger/index.html | API documentation (butuh `make swagger`) |
| Jaeger UI | http://localhost:16686 | Distributed traces |
| Prometheus | http://localhost:9090 | Metrics explorer |
| NATS Monitor | http://localhost:8222 | NATS server status |

---

## Troubleshooting

**Port 8080 sudah dipakai:**
```bash
lsof -ti :8080 | xargs kill -9
```

**Migration gagal — database belum siap:**
```bash
docker compose ps        # cek status postgres container
make migrate-up          # coba lagi setelah postgres healthy
```

**`go: command not found` setelah install:**
```bash
export PATH=$PATH:/usr/local/go/bin   # tambahkan ke ~/.zshrc atau ~/.bashrc
```

**`air: command not found`:**
```bash
make tools-install       # install ulang dev tools
export PATH=$PATH:$(go env GOPATH)/bin
```

**Web dashboard tidak terhubung ke API:**
Cek `VITE_API_BASE_URL` di `web-dashboard/.env.local` — harus menunjuk ke URL yang benar.
