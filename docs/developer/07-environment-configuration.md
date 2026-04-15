# 07 — Environment Configuration

## Go API — Semua Environment Variables

### Application

| Env Var | Default | Required | Deskripsi |
|---------|---------|----------|-----------|
| `APP_NAME` | `boilerplate` | No | Nama aplikasi (dipakai di logging, OTel service name) |
| `APP_ENV` | `development` | No | Environment: `development`, `staging`, `production` |
| `HTTP_PORT` | `8080` | No | Port HTTP server |
| `LOG_LEVEL` | `debug` | No | Level log: `debug`, `info`, `warn`, `error` |

### Database (PostgreSQL)

| Env Var | Default | Required | Deskripsi |
|---------|---------|----------|-----------|
| `DATABASE_URL` | - | **Yes** | PostgreSQL connection string format DSN |
| `DB_MAX_OPEN_CONNS` | `25` | No | Max koneksi aktif ke database |
| `DB_MAX_IDLE_CONNS` | `5` | No | Max koneksi idle di pool |
| `DB_CONN_MAX_LIFETIME` | `5m` | No | Lifetime maksimal satu koneksi |

**Format `DATABASE_URL`:**
```
postgres://user:password@host:port/dbname?sslmode=disable
postgres://postgres:postgres@localhost:5432/boilerplate_db?sslmode=disable
```

### Cache (Redis)

| Env Var | Default | Required | Deskripsi |
|---------|---------|----------|-----------|
| `REDIS_URL` | - | **Yes** | Redis connection URL |
| `REDIS_TTL_SECONDS` | `300` | No | TTL default untuk semua cache entry (5 menit) |

**Format `REDIS_URL`:**
```
redis://localhost:6379/0
redis://:password@host:6379/0
rediss://user:password@host:6380/0   (TLS)
```

### JWT Authentication

| Env Var | Default | Required | Deskripsi |
|---------|---------|----------|-----------|
| `JWT_SECRET` | - | **Yes** | Secret key untuk signing JWT. **Minimum 32 karakter.** |
| `JWT_EXPIRY_HOURS` | `24` | No | Expiry access token dalam jam |
| `JWT_ISSUER` | `boilerplate-api` | No | Issuer claim di JWT |
| `JWT_REFRESH_MULTIPLIER` | `7` | No | Refresh token = access token × multiplier (expiry) |

**Membuat JWT secret yang aman:**
```bash
openssl rand -base64 48
# Output: dY3h7z5K9...   ← gunakan ini sebagai JWT_SECRET
```

### Tenant & Scope

| Env Var | Default | Required | Deskripsi |
|---------|---------|----------|-----------|
| `TENANT_MODE` | `single` | No | Mode tenancy: `single` atau `multi` |
| `TENANT_ID` | - | **Yes (single)** | UUID top-level namespace. Wajib jika `TENANT_MODE=single` |
| `COMPANY_ID` | - | **Yes (single)** | UUID company/legal entity. Wajib jika `TENANT_MODE=single` |
| `BRANCH_ID` | - | No | UUID cabang operasional. Opsional |
| `WAREHOUSE_ID` | - | No | UUID gudang. Opsional; lebih baik kirim via request body |

**Membuat UUID:**
```bash
python3 -c "import uuid; print(uuid.uuid4())"
# atau
uuidgen
```

### HTTP Server

| Env Var | Default | Required | Deskripsi |
|---------|---------|----------|-----------|
| `HTTP_READ_TIMEOUT_SECONDS` | `15` | No | Read timeout per request |
| `HTTP_WRITE_TIMEOUT_SECONDS` | `15` | No | Write timeout per request |
| `HTTP_IDLE_TIMEOUT_SECONDS` | `60` | No | Idle connection timeout |
| `HTTP_MIDDLEWARE_TIMEOUT_SECONDS` | `30` | No | Timeout maksimal handler via `chimiddleware.Timeout` |
| `HTTP_MAX_PAGE_SIZE` | `1000` | No | Batas `limit` parameter untuk endpoint list |

### Event Bus (NATS JetStream)

| Env Var | Default | Required | Deskripsi |
|---------|---------|----------|-----------|
| `USE_NATS` | `false` | No | `true` = gunakan NATS JetStream; `false` = InMemory (dev) |
| `NATS_URL` | - | **Yes (jika USE_NATS=true)** | NATS server URL |
| `NATS_STREAM_NAME` | `APP_EVENTS` | No | Nama JetStream stream |

**Format `NATS_URL`:**
```
nats://localhost:4222
nats://user:password@host:4222
```

### Observability

| Env Var | Default | Required | Deskripsi |
|---------|---------|----------|-----------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | - | No | Endpoint OTLP exporter (Jaeger, Grafana Tempo, dll) |
| `PROMETHEUS_PORT` | `9090` | No | Port Prometheus metrics scrape (informasional) |

**Format `OTEL_EXPORTER_OTLP_ENDPOINT`:**
```
http://localhost:4318    (Jaeger via docker-compose)
http://tempo:4317        (Grafana Tempo)
```

Jika `OTEL_EXPORTER_OTLP_ENDPOINT` kosong atau tidak dapat dihubungi, aplikasi tetap berjalan tanpa tracing (non-fatal warning).

---

## .env.example Walkthrough

```env
# ═══════════════════════════════════════════════════════
# APPLICATION
# ═══════════════════════════════════════════════════════
APP_NAME=boilerplate          # Nama service di traces dan logs
APP_ENV=development           # Ganti ke "production" di prod
HTTP_PORT=8080
LOG_LEVEL=debug               # Di prod: "info" atau "warn"

# ═══════════════════════════════════════════════════════
# DATABASE (PostgreSQL)
# ═══════════════════════════════════════════════════════
DATABASE_URL=postgres://postgres:postgres@localhost:5432/boilerplate_db?sslmode=disable

# Tuning connection pool — sesuaikan dengan kapasitas DB server
DB_MAX_OPEN_CONNS=25          # Max aktif connections
DB_MAX_IDLE_CONNS=5           # Max idle connections
DB_CONN_MAX_LIFETIME=5m       # Lifetime koneksi (format Go duration)

# ═══════════════════════════════════════════════════════
# CACHE (Redis)
# ═══════════════════════════════════════════════════════
REDIS_URL=redis://localhost:6379/0
REDIS_TTL_SECONDS=300         # 5 menit default TTL

# ═══════════════════════════════════════════════════════
# JWT AUTHENTICATION
# ═══════════════════════════════════════════════════════
# WAJIB DIGANTI — minimum 32 karakter, gunakan: openssl rand -base64 48
JWT_SECRET=change-me-to-a-random-string-of-at-least-32-characters

JWT_EXPIRY_HOURS=24           # Access token valid 24 jam
JWT_ISSUER=boilerplate-api    # Issuer claim
JWT_REFRESH_MULTIPLIER=7      # Refresh token = 24 × 7 = 168 jam (7 hari)

# ═══════════════════════════════════════════════════════
# MULTI-TENANT SCOPE
# ═══════════════════════════════════════════════════════
# Mode: "single" (satu org) atau "multi" (banyak org via JWT)
TENANT_MODE=single

# Wajib diisi jika TENANT_MODE=single. Isi dengan UUID yang valid.
# Generate: python3 -c "import uuid; print(uuid.uuid4())"
TENANT_ID=00000000-0000-0000-0000-000000000001
COMPANY_ID=00000000-0000-0000-0000-000000000002

# Opsional — kosongkan jika tidak dipakai
BRANCH_ID=
WAREHOUSE_ID=

# ═══════════════════════════════════════════════════════
# HTTP SERVER TIMEOUTS
# ═══════════════════════════════════════════════════════
HTTP_READ_TIMEOUT_SECONDS=15
HTTP_WRITE_TIMEOUT_SECONDS=15
HTTP_IDLE_TIMEOUT_SECONDS=60
HTTP_MIDDLEWARE_TIMEOUT_SECONDS=30
HTTP_MAX_PAGE_SIZE=1000

# ═══════════════════════════════════════════════════════
# EVENT BUS (NATS JetStream)
# ═══════════════════════════════════════════════════════
# Di development: USE_NATS=false (InMemory event bus)
# Di production: USE_NATS=true + isi NATS_URL
USE_NATS=false
NATS_URL=nats://localhost:4222
NATS_STREAM_NAME=APP_EVENTS

# ═══════════════════════════════════════════════════════
# OBSERVABILITY
# ═══════════════════════════════════════════════════════
# Jaeger via docker-compose: http://localhost:4318
# Kosongkan untuk menonaktifkan tracing (non-fatal)
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

PROMETHEUS_PORT=9090          # Informasional — metrics di-serve di /metrics
```

---

## Web Dashboard — Environment Variables

File: `web-dashboard/.env.local`

| Env Var | Default | Required | Deskripsi |
|---------|---------|----------|-----------|
| `VITE_API_BASE_URL` | `http://localhost:8080` | No | Base URL Go API |
| `VITE_APP_NAME` | `Dashboard` | No | Nama app ditampilkan di navbar dan browser title |
| `VITE_MULTI_TENANT` | `false` | No | `true` = aktifkan multi-tenant routing dan UI |
| `VITE_MOCK_MODE` | `false` | No | `true` = gunakan MSW mock, tidak hit API real |

```env
# web-dashboard/.env.local

VITE_API_BASE_URL=http://localhost:8080
VITE_APP_NAME=My Dashboard
VITE_MULTI_TENANT=false
VITE_MOCK_MODE=false
```

**Perbedaan `.env.local` vs `.env`:**
- `.env.local` — tidak di-commit ke git (sensitive config, override lokal)
- `.env` — bisa di-commit (default values untuk semua developer)

---

## Docker Services

Semua services didefinisikan di `api/docker-compose.yml`.

### Menjalankan Services

```bash
cd api

# Start semua services
make infra-up
# Setara: docker compose up -d

# Stop semua services
make infra-down
# Setara: docker compose down

# Lihat status
docker compose ps

# Lihat logs
docker compose logs postgres -f
docker compose logs redis -f
docker compose logs nats -f
```

### Detail Services

#### PostgreSQL 17

```yaml
image: postgres:17-alpine
ports: ["5432:5432"]
environment:
  POSTGRES_DB: boilerplate_db
  POSTGRES_USER: postgres
  POSTGRES_PASSWORD: postgres
```

| Detail | Nilai |
|--------|-------|
| Host | `localhost` |
| Port | `5432` |
| Database | `boilerplate_db` |
| Username | `postgres` |
| Password | `postgres` |
| Connection string | `postgres://postgres:postgres@localhost:5432/boilerplate_db?sslmode=disable` |

**Akses psql:**
```bash
docker compose exec postgres psql -U postgres -d boilerplate_db
```

#### Redis 7

```yaml
image: redis:7-alpine
ports: ["6379:6379"]
```

| Detail | Nilai |
|--------|-------|
| Host | `localhost` |
| Port | `6379` |
| Connection string | `redis://localhost:6379/0` |

**Akses redis-cli:**
```bash
docker compose exec redis redis-cli
```

#### NATS 2.10 (dengan JetStream)

```yaml
image: nats:2.10-alpine
command: ["--js", "--sd", "/data", "--http_port", "8222"]
ports:
  - "4222:4222"  # client port
  - "8222:8222"  # monitoring UI
```

| Detail | Nilai |
|--------|-------|
| Client Port | `4222` |
| Monitoring UI | http://localhost:8222 |
| Connection string | `nats://localhost:4222` |

NATS hanya digunakan jika `USE_NATS=true`. Di development, InMemory event bus digunakan.

#### Jaeger (Tracing)

```yaml
image: jaegertracing/all-in-one:1.57
ports:
  - "16686:16686"  # UI
  - "4318:4318"    # OTLP HTTP receiver
```

| Detail | Nilai |
|--------|-------|
| UI | http://localhost:16686 |
| OTLP Endpoint | `http://localhost:4318` |

Set `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318` di `.env` untuk mengirim traces.

#### Prometheus (Metrics)

```yaml
image: prom/prometheus:v2.51.0
ports: ["9090:9090"]
volumes:
  - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

| Detail | Nilai |
|--------|-------|
| UI | http://localhost:9090 |
| Scrape config | `api/prometheus.yml` |
| API metrics endpoint | http://localhost:8080/metrics |

Prometheus dikonfigurasi untuk scrape `/metrics` dari API setiap 15 detik.

---

## Konfigurasi per Environment

### Development (local)

```env
APP_ENV=development
LOG_LEVEL=debug
USE_NATS=false           # InMemory event bus, lebih simpel
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318   # Jaeger lokal
JWT_SECRET=dev-secret-that-is-at-least-32-characters-long
TENANT_MODE=single
TENANT_ID=00000000-0000-0000-0000-000000000001
COMPANY_ID=00000000-0000-0000-0000-000000000002
```

### Staging

```env
APP_ENV=staging
LOG_LEVEL=info
USE_NATS=true
NATS_URL=nats://nats-service:4222
OTEL_EXPORTER_OTLP_ENDPOINT=http://tempo:4317
DATABASE_URL=postgres://app:secret@db-staging:5432/app?sslmode=require
REDIS_URL=redis://:secret@redis-staging:6379/0
JWT_SECRET=<secret-dari-secret-manager>
TENANT_MODE=single
TENANT_ID=<uuid-staging>
COMPANY_ID=<uuid-staging>
```

### Production (multi-tenant SaaS)

```env
APP_ENV=production
LOG_LEVEL=warn
USE_NATS=true
NATS_URL=nats://nats-cluster:4222
OTEL_EXPORTER_OTLP_ENDPOINT=https://tempo.yourorg.com:4317
DATABASE_URL=postgres://app:secret@db-prod:5432/app?sslmode=require&pool_max_conns=20
REDIS_URL=rediss://:secret@redis-prod:6380/0
JWT_SECRET=<secret-dari-vault>
JWT_EXPIRY_HOURS=8       # lebih ketat di prod
TENANT_MODE=multi        # baca scope dari JWT per request
# TENANT_ID dan COMPANY_ID tidak perlu diset di mode multi
DB_MAX_OPEN_CONNS=50     # scale up sesuai kapasitas DB
DB_MAX_IDLE_CONNS=10
```

---

## Secret Management

**JANGAN pernah commit `.env` ke git.**

Rekomendasi untuk production:
- **HashiCorp Vault** — inject secrets via agent sidecar
- **AWS Secrets Manager / Parameter Store** — baca saat startup
- **Kubernetes Secrets** — mount sebagai environment variables

`.gitignore` sudah mengecualikan `.env`:
```
.env
.env.local
.env.*.local
```
