# 02 — Deployment Guide

## Prerequisites

Sebelum memulai, pastikan tools berikut sudah terinstall di server/mesin Anda:

| Tool | Versi Minimum | Cara Cek |
|------|--------------|----------|
| Docker | 24.x | `docker --version` |
| Docker Compose | v2.x | `docker compose version` |
| Git | 2.x | `git --version` |

> **Catatan:** Docker Compose v2 menggunakan perintah `docker compose` (dengan spasi),
> bukan `docker-compose` (dengan tanda hubung). Pastikan Anda menggunakan versi yang benar.

---

## Quick Start (Development / Lokal)

Langkah tercepat untuk menjalankan sistem di lokal:

### Langkah 1 — Clone repository

```bash
git clone <repository-url> boilerplate
cd boilerplate
```

### Langkah 2 — Jalankan infrastructure services

```bash
cd api
docker compose up -d postgres redis nats jaeger prometheus
```

Tunggu hingga semua service healthy. Cek status:

```bash
docker compose ps
```

Output yang diharapkan:

```
NAME        STATUS          PORTS
postgres    Up (healthy)    0.0.0.0:5432->5432/tcp
redis       Up (healthy)    0.0.0.0:6379->6379/tcp
nats        Up              0.0.0.0:4222->4222/tcp
jaeger      Up              0.0.0.0:16686->16686/tcp
prometheus  Up              0.0.0.0:9090->9090/tcp
```

### Langkah 3 — Setup environment API

```bash
# Masih di folder api/
cp .env.example .env
```

Edit file `.env` dan sesuaikan nilai berikut (minimal):

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/boilerplate_db?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=ganti-dengan-string-acak-minimal-32-karakter
TENANT_MODE=single
```

### Langkah 4 — Jalankan database migration

```bash
make migrate-up
```

### Langkah 5 — Jalankan API server

```bash
make dev
```

API server berjalan di `http://localhost:8080`

Verifikasi:
```bash
curl http://localhost:8080/health
```

Output yang diharapkan:
```json
{"status": "ok"}
```

### Langkah 6 — Setup dan jalankan Web Dashboard

```bash
cd ../web-dashboard
cp .env.local.example .env.local   # atau buat manual jika belum ada
```

Isi `.env.local`:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_NAME=Dashboard
VITE_MULTI_TENANT=false
```

Install dependencies dan jalankan:

```bash
npm install
npm run dev
```

Dashboard berjalan di `http://localhost:5173`

---

## Environment Setup Langkah Demi Langkah

### Konfigurasi API (`api/.env`)

Buat file `.env` dari template:

```bash
cp api/.env.example api/.env
```

Variabel yang **wajib** dikonfigurasi:

```env
# ─── Database ────────────────────────────────────────────────
DATABASE_URL=postgres://USER:PASSWORD@HOST:5432/DB_NAME?sslmode=require

# ─── Cache ───────────────────────────────────────────────────
REDIS_URL=redis://:PASSWORD@HOST:6379

# ─── Autentikasi ─────────────────────────────────────────────
JWT_SECRET=string-acak-sangat-panjang-minimal-32-karakter

# ─── Tenant Mode ─────────────────────────────────────────────
# Pilih salah satu: single atau multi
TENANT_MODE=single

# Wajib jika TENANT_MODE=single:
TENANT_ID=uuid-tenant-anda
COMPANY_ID=uuid-company-anda
```

### Konfigurasi Web Dashboard (`web-dashboard/.env.local`)

```env
VITE_API_BASE_URL=https://api.domain-anda.com
VITE_APP_NAME=Nama Aplikasi Anda
VITE_MULTI_TENANT=false
```

---

## Production Deployment

### Rekomendasi Arsitektur Production

```
Internet
   │
   ▼
┌─────────────────┐
│   Nginx / Load  │  ← SSL termination, reverse proxy
│   Balancer      │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
 API          Web Dashboard
(Go binary)   (static files / CDN)
    │
    ├── PostgreSQL (RDS / managed)
    ├── Redis (ElastiCache / managed)
    └── NATS (self-hosted / JetStream)
```

### Opsi A — Deploy dengan Docker Compose (Production)

Buat `docker-compose.prod.yml` di luar folder `api/`:

```yaml
version: '3.9'
services:
  api:
    image: boilerplate-api:latest
    build:
      context: ./api
      dockerfile: Dockerfile
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
      - TENANT_MODE=${TENANT_MODE}
      - HTTP_PORT=8080
    ports:
      - "8080:8080"
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:17-alpine
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

Deploy:

```bash
docker compose -f docker-compose.prod.yml up -d
```

### Opsi B — Build Binary Langsung

Build Go binary:

```bash
cd api
go build -o ./bin/api ./cmd/api
```

Jalankan sebagai system service (systemd):

```ini
# /etc/systemd/system/boilerplate-api.service
[Unit]
Description=Boilerplate API Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=api
WorkingDirectory=/opt/boilerplate/api
EnvironmentFile=/opt/boilerplate/api/.env
ExecStart=/opt/boilerplate/api/bin/api
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
systemctl enable boilerplate-api
systemctl start boilerplate-api
```

---

## Production Deployment Checklist

Sebelum go-live, pastikan semua item berikut sudah dicentang:

### Security

- [ ] `JWT_SECRET` minimal 64 karakter, dibuat secara acak (`openssl rand -hex 32`)
- [ ] Database tidak terekspos ke public internet (hanya accessible dari API)
- [ ] Redis dilindungi password (`requirepass`)
- [ ] HTTPS aktif (SSL certificate terpasang di Nginx/Load Balancer)
- [ ] File `.env` tidak di-commit ke Git
- [ ] API tidak berjalan sebagai root user

### Database

- [ ] Migration sudah dijalankan (`make migrate-up`)
- [ ] Backup database terjadwal (minimal harian)
- [ ] Connection pool dikonfigurasi sesuai kapasitas server
- [ ] PostgreSQL berjalan dengan `sslmode=require` di production

### Infrastructure

- [ ] Docker service restart policy diset ke `unless-stopped` atau `always`
- [ ] Volume data (postgres_data, redis_data) dipersist dan di-backup
- [ ] NATS JetStream diaktifkan jika menggunakan event streaming (`USE_NATS=true`)

### Monitoring

- [ ] Prometheus aktif dan scraping `/metrics`
- [ ] Jaeger aktif dan menerima traces
- [ ] Alert rules dikonfigurasi di Prometheus
- [ ] Log output diarahkan ke log aggregation system (Loki, ELK, dll)

### Application

- [ ] Health check berjalan: `curl https://api.domain.com/health`
- [ ] Swagger UI accessible: `https://api.domain.com/swagger/index.html`
- [ ] Web Dashboard dapat login dan mengakses data

---

## Health Check dan Verification

### 1. Cek API Health

```bash
curl -s http://localhost:8080/health | python3 -m json.tool
```

Response sukses:
```json
{
  "status": "ok"
}
```

### 2. Cek Metrics Endpoint

```bash
curl -s http://localhost:8080/metrics | head -30
```

### 3. Cek Swagger UI

Buka browser: `http://localhost:8080/swagger/index.html`

Halaman dokumentasi API interaktif akan tampil.

### 4. Cek Jaeger UI

Buka browser: `http://localhost:16686`

Anda akan melihat daftar service yang mengirim trace.

### 5. Cek Prometheus UI

Buka browser: `http://localhost:9090`

Query contoh: ketik `up` di search bar untuk melihat status semua target.

### 6. Verifikasi Koneksi Database

```bash
docker compose exec postgres psql -U postgres -d boilerplate_db -c "\dt"
```

Akan menampilkan daftar tabel yang sudah dibuat oleh migration.

---

## Common Deployment Issues dan Solusi

### Issue 1: Port sudah digunakan

**Gejala:**
```
Error response from daemon: driver failed programming external connectivity on endpoint postgres: 
Bind for 0.0.0.0:5432 failed: port is already allocated
```

**Solusi:**
```bash
# Cek proses yang menggunakan port
lsof -i :5432

# Hentikan service yang bertabrakan, atau ubah port di docker-compose.yml:
ports:
  - "15432:5432"   # Ubah port host ke 15432

# Update DATABASE_URL di .env:
DATABASE_URL=postgres://postgres:postgres@localhost:15432/boilerplate_db
```

---

### Issue 2: API gagal konek ke database

**Gejala:**
```
failed to connect to database: dial tcp: connect: connection refused
```

**Penyebab umum dan solusi:**

1. PostgreSQL belum healthy — tunggu 10-15 detik setelah `docker compose up`
2. `DATABASE_URL` salah — cek username, password, nama database
3. PostgreSQL belum menerima koneksi eksternal — pastikan host adalah `localhost`
   (bukan `postgres` kecuali API juga berjalan dalam Docker network yang sama)

```bash
# Cek status PostgreSQL
docker compose exec postgres pg_isready -U postgres

# Test koneksi manual
psql "postgres://postgres:postgres@localhost:5432/boilerplate_db"
```

---

### Issue 3: Migration gagal

**Gejala:**
```
error: migration failed: pq: relation "..." already exists
```

**Solusi:**
```bash
# Cek versi migration saat ini
make migrate-status

# Jika ada migration yang stuck, reset (HATI-HATI di production):
make migrate-down   # rollback satu step
make migrate-up     # apply ulang
```

---

### Issue 4: JWT secret terlalu pendek

**Gejala:**
```
jwt secret must be at least 32 characters
```

**Solusi:**
```bash
# Generate JWT secret yang aman
openssl rand -hex 32

# Salin hasilnya ke .env
JWT_SECRET=hasil-openssl-di-atas
```

---

### Issue 5: Web Dashboard tidak bisa konek ke API (CORS error)

**Gejala di browser console:**
```
Access to fetch at 'http://localhost:8080/api/...' from origin 'http://localhost:5173' 
has been blocked by CORS policy
```

**Solusi:**

Tambahkan `ALLOWED_ORIGINS` ke `.env` API:
```env
ALLOWED_ORIGINS=http://localhost:5173,https://dashboard.domain-anda.com
```

---

### Issue 6: Docker Compose versi lama

**Gejala:**
```
docker-compose: command not found
```

**Solusi:**

Gunakan `docker compose` (bukan `docker-compose`). Jika Docker Compose v2 belum
terinstall, install via:

```bash
# Ubuntu/Debian
apt-get install docker-compose-plugin

# macOS
brew install docker-compose
```

---

### Issue 7: NATS JetStream tidak berjalan

**Gejala:**
```
failed to subscribe to subject: nats: no response from server
```

**Solusi:**

Pastikan NATS dijalankan dengan flag `--js`:

```bash
# Cek perintah NATS di docker-compose.yml
command: ["--js", "--sd", "/data"]

# Restart NATS
docker compose restart nats

# Verifikasi JetStream aktif
curl http://localhost:8222/jsz
```

Atau jika NATS tidak dibutuhkan, nonaktifkan:
```env
USE_NATS=false
```
