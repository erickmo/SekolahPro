# 03 — Configuration Reference

Semua konfigurasi sistem menggunakan **environment variables**. Tidak ada file
konfigurasi khusus — semua dibaca dari `.env` (development) atau environment
container/server (production).

---

## Konfigurasi API (`api/.env`)

### Variabel Wajib

| Variable | Contoh Nilai | Deskripsi |
|----------|-------------|-----------|
| `DATABASE_URL` | `postgres://user:pass@host:5432/db?sslmode=disable` | Connection string PostgreSQL lengkap |
| `REDIS_URL` | `redis://localhost:6379` | Connection string Redis |
| `JWT_SECRET` | `a1b2c3d4...` (min 32 karakter) | Secret key untuk signing JWT token |
| `TENANT_MODE` | `single` atau `multi` | Mode operasi tenant sistem |

### HTTP Server

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `HTTP_PORT` | `8080` | Port yang didengarkan API server |
| `ALLOWED_ORIGINS` | `*` | Daftar origin yang diizinkan CORS, pisahkan dengan koma. Gunakan domain spesifik di production |
| `ENV` | `development` | Environment label (`development`, `staging`, `production`) |

### Tenant & Scope

| Variable | Wajib | Deskripsi |
|----------|-------|-----------|
| `TENANT_MODE` | Ya | `single` = satu organisasi per instance. `multi` = SaaS multi-organisasi |
| `TENANT_ID` | Ya (jika `single`) | UUID tenant. Dikonfigurasi sekali di server, tidak perlu di JWT |
| `COMPANY_ID` | Ya (jika `single`) | UUID company default untuk mode single-tenant |
| `BRANCH_ID` | Tidak | UUID branch default (opsional, untuk mode single dengan cabang) |
| `WAREHOUSE_ID` | Tidak | UUID warehouse default (opsional) |

### JWT Authentication

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `JWT_SECRET` | — | Secret key. **Wajib** minimal 32 karakter. Di production gunakan minimal 64 karakter |
| `JWT_ACCESS_TTL` | `15m` | Masa berlaku access token (format: `15m`, `1h`, `24h`) |
| `JWT_REFRESH_TTL` | `168h` | Masa berlaku refresh token (default 7 hari) |

> **Rekomendasi keamanan:** Generate JWT_SECRET dengan `openssl rand -hex 32`
> untuk menghasilkan 64 karakter hex acak.

### Database Connection Pool

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `DB_MAX_OPEN_CONNS` | `25` | Maksimum koneksi database yang dibuka secara bersamaan |
| `DB_MAX_IDLE_CONNS` | `5` | Maksimum koneksi idle yang dipertahankan di pool |
| `DB_CONN_MAX_LIFETIME` | `5m` | Maksimum usia sebuah koneksi sebelum ditutup dan dibuat ulang |

> **Tuning production:** Sesuaikan `DB_MAX_OPEN_CONNS` berdasarkan `max_connections`
> PostgreSQL Anda. Rumus umum: `max_connections_postgres / jumlah_instance_api - 5`.

### Cache (Redis)

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `REDIS_URL` | — | Connection string Redis. Format: `redis://:password@host:6379` |
| `CACHE_DEFAULT_TTL` | `5m` | TTL default untuk item cache |

### Event Streaming (NATS)

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `USE_NATS` | `false` | `true` untuk menggunakan NATS JetStream. `false` menggunakan InMemory bus (development only) |
| `NATS_URL` | `nats://localhost:4222` | URL koneksi NATS server |

### Observability

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `JAEGER_ENDPOINT` | `http://localhost:4318/v1/traces` | OTLP HTTP endpoint Jaeger untuk menerima traces |
| `LOG_LEVEL` | `info` | Level logging: `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | Format output log: `json` (production) atau `pretty` (development) |

---

## Konfigurasi Web Dashboard (`web-dashboard/.env.local`)

### Variabel Wajib

| Variable | Contoh Nilai | Deskripsi |
|----------|-------------|-----------|
| `VITE_API_BASE_URL` | `http://localhost:8080` | URL base API backend. Tanpa trailing slash |
| `VITE_MULTI_TENANT` | `false` atau `true` | Harus sesuai dengan `TENANT_MODE` di API |

### Variabel Opsional

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `VITE_APP_NAME` | `Dashboard` | Nama aplikasi yang ditampilkan di browser tab dan navbar |
| `VITE_MOCK_MODE` | `false` | `true` untuk menggunakan mock API (development tanpa backend) |

> **Penting:** Semua variabel dashboard harus diawali `VITE_`. Variable yang tidak
> diawali `VITE_` tidak akan tersedia di browser (Vite build system requirement).

---

## Single-Tenant vs Multi-Tenant Configuration Comparison

### Mode Single-Tenant

Gunakan ketika satu instance melayani satu organisasi (misalnya: self-hosted ERP internal).

**API `.env`:**
```env
TENANT_MODE=single
TENANT_ID=550e8400-e29b-41d4-a716-446655440001
COMPANY_ID=550e8400-e29b-41d4-a716-446655440002

# JWT tidak harus berisi tenant/company — scope dibaca dari config
JWT_SECRET=your-secret-here
JWT_ACCESS_TTL=8h
JWT_REFRESH_TTL=720h
```

**Web Dashboard `.env.local`:**
```env
VITE_API_BASE_URL=https://api.perusahaan.com
VITE_APP_NAME=ERP Perusahaan
VITE_MULTI_TENANT=false
```

**Alur login single-tenant:**
```
/login → [masukkan kredensial] → /dashboard
```

---

### Mode Multi-Tenant

Gunakan ketika satu instance melayani banyak pelanggan (SaaS).

**API `.env`:**
```env
TENANT_MODE=multi

# Tidak perlu TENANT_ID dan COMPANY_ID — dibaca dari JWT
JWT_SECRET=your-secret-here
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h
```

**Web Dashboard `.env.local`:**
```env
VITE_API_BASE_URL=https://api.saas-anda.com
VITE_APP_NAME=Platform SaaS
VITE_MULTI_TENANT=true
```

**Alur login multi-tenant:**
```
/login
  → [fase 1: token berisi tenant_id saja]
  → /choose-company
  → [pilih company]
  → [fase 2: token berisi tenant_id + company_id]
  → /c/:companyCode/dashboard
```

---

### Tabel Perbandingan

| Aspek | Single-Tenant | Multi-Tenant |
|-------|--------------|--------------|
| `TENANT_MODE` | `single` | `multi` |
| `TENANT_ID` di env | Wajib | Tidak diperlukan |
| `COMPANY_ID` di env | Wajib | Tidak diperlukan |
| JWT berisi scope | Opsional | Wajib |
| Login flow | 1 langkah | 2 langkah (+ pilih company) |
| Route dashboard | `/dashboard` | `/c/:code/dashboard` |
| `VITE_MULTI_TENANT` | `false` | `true` |

---

## Contoh File .env Lengkap

### Development (Single-Tenant)

```env
# Server
HTTP_PORT=8080
ENV=development
LOG_LEVEL=debug
LOG_FORMAT=pretty

# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/boilerplate_db?sslmode=disable
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# Cache
REDIS_URL=redis://localhost:6379
CACHE_DEFAULT_TTL=5m

# Event Streaming (development menggunakan InMemory)
USE_NATS=false

# JWT
JWT_SECRET=dev-secret-ganti-di-production-minimal32char
JWT_ACCESS_TTL=8h
JWT_REFRESH_TTL=720h

# Tenant
TENANT_MODE=single
TENANT_ID=550e8400-e29b-41d4-a716-446655440001
COMPANY_ID=550e8400-e29b-41d4-a716-446655440002

# Observability
JAEGER_ENDPOINT=http://localhost:4318/v1/traces
ALLOWED_ORIGINS=http://localhost:5173
```

### Production (Multi-Tenant)

```env
# Server
HTTP_PORT=8080
ENV=production
LOG_LEVEL=info
LOG_FORMAT=json

# Database
DATABASE_URL=postgres://api_user:STRONG_PASSWORD@db.internal:5432/saas_db?sslmode=require
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# Cache
REDIS_URL=redis://:REDIS_PASSWORD@cache.internal:6379
CACHE_DEFAULT_TTL=5m

# Event Streaming (production menggunakan NATS)
USE_NATS=true
NATS_URL=nats://nats.internal:4222

# JWT
JWT_SECRET=64-karakter-random-string-yang-sangat-panjang-dan-aman-dari-openssl
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# Tenant
TENANT_MODE=multi

# Observability
JAEGER_ENDPOINT=http://jaeger.internal:4318/v1/traces
ALLOWED_ORIGINS=https://app.saas-anda.com,https://dashboard.saas-anda.com
```
