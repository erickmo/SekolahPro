# 06 — Deployment & Operasional SekolahPro

Dokumen ini menjabarkan strategi deployment, CI/CD pipeline, environment strategy, infrastructure as code, secrets management, dan observability untuk SekolahPro. Dokumen ini menjadi referensi untuk semua engineer, DevOps, dan SRE yang mengelola infrastruktur SekolahPro.

---

## ADR References

| ADR | Judul | Relevansi |
|-----|-------|-----------|
| ADR-016 | Deployment & CI/CD Pipeline | Strategi deployment, CI/CD, dan observability |
| ADR-017 | Strategi Migrasi Data & Import | Import awal, batch processing, rollback |

---

## 1. Container-Based Deployment

Semua service dikemas dalam Docker image dengan multi-stage build:

| Service | Base Image | Ukuran Target |
|---------|-----------|---------------|
| Go API | `gcr.io/distroless/static` | < 30 MB |
| app-admin | `nginx:alpine` (static files) | < 20 MB |
| app-portal | `nginx:alpine` (static files) | < 20 MB |
| app-pos | `nginx:alpine` (static files) | < 20 MB |

Multi-stage build memastikan image sekecil mungkin — binary Go dicompile di stage builder, hanya binary yang masuk ke distroless image final.

### Prinsip: Immutable Image

Image yang sama di-deploy ke staging dan production — hanya konfigurasi environment yang berbeda (via secrets/config maps). Tidak ada perubahan kode setelah image di-build.

---

## 2. Topologi Deployment

```
                    Internet
                       │
              ┌────────┴────────┐
              │  Load Balancer  │
              │ (nginx-ingress) │
              └────────┬────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
   ┌────┴────┐   ┌─────┴────┐  ┌─────┴────┐
   │ Go API  │   │CDN Admin │  │CDN Portal│
   │ 2+ reps │   │  /admin  │  │ /portal  │
   │ HPA 2-10│   └──────────┘  └──────────┘
   └────┬────┘
        │
   ┌────┴────────────────────┐
   │  Backend Infrastructure │
   │ ┌──────────────────────┐│
   │ │ PostgreSQL (managed) ││
   │ │ Cloud SQL / RDS      ││
   │ └──────────────────────┘│
   │ ┌──────────────────────┐│
   │ │ Redis (managed)      ││
   │ │ Memorystore/ElastiCache│
   │ └──────────────────────┘│
   │ ┌──────────────────────┐│
   │ │ NATS JetStream       ││
   │ │ 1-3 nodes StatefulSet││
   │ └──────────────────────┘│
   └─────────────────────────┘
```

### Orchestration

| Environment | Platform | Tujuan |
|-------------|----------|--------|
| Local | Docker Compose | Development — `docker compose up` untuk seluruh stack |
| Staging | Kubernetes namespace | QA + integration test |
| Production | Kubernetes (GKE atau EKS) | Live users — autoscaling, rolling update |

---

## 3. CI/CD Pipeline (GitHub Actions)

```
┌──────────────────────────────────────────────────────────────┐
│  On Pull Request (setiap PR ke main)                         │
│                                                              │
│  ├── lint (golangci-lint + eslint)  ─┐                       │
│  ├── test (go test + vitest)         ├── parallel            │
│  └── build (docker build test stage)─┘                       │
│                                                              │
│  Semua harus PASS sebelum PR bisa di-merge                   │
├──────────────────────────────────────────────────────────────┤
│  On Merge to main                                            │
│                                                              │
│  ├── build Docker image                                      │
│  ├── push ke Container Registry (GCR / ECR)                  │
│  │     tag: main-{git_sha}                                   │
│  └── deploy ke staging (otomatis)                            │
│        via helm upgrade atau kubectl apply                   │
├──────────────────────────────────────────────────────────────┤
│  On Tag (format: vX.Y.Z)                                     │
│                                                              │
│  ├── promote staging image ke production                     │
│  └── manual approval gate (via GitHub Environment Protection)│
│        image yang sama (staging image) → production          │
└──────────────────────────────────────────────────────────────┘
```

### Frontend CI (Turborepo-aware)

Turborepo dependency graph memastikan hanya apps yang terpengaruh yang di-rebuild:

- Perubahan di `apps/app-admin/` → rebuild + redeploy hanya `app-admin`
- Perubahan di `packages/ui` → rebuild + redeploy semua 3 apps

---

## 4. Database Migration Strategy

Tool: **golang-migrate**

Setiap migration memiliki dua file:
```
migrations/
  00001_uuid_v7_function.up.sql
  00001_uuid_v7_function.down.sql
  00002_create_tenants.up.sql
  00002_create_tenants.down.sql
  ...
```

### Aturan Migration

- Migration dijalankan sebagai **Kubernetes init container** sebelum app container start — memastikan schema selalu up-to-date sebelum kode baru berjalan
- **Tidak ada auto-migrate** di application code — migration selalu eksplisit dan terdokumentasi
- Setiap migration wajib memiliki `.down.sql` untuk rollback
- Migration pertama (`00001`) selalu harus `uuid_generate_v7_function` — ini dependency semua tabel

### Zero-Downtime Migration

Untuk tabel production besar (>1M rows):
1. Tambahkan kolom baru sebagai nullable
2. Backfill data secara bertahap (batch, off-peak hours)
3. Tambahkan NOT NULL constraint setelah backfill selesai
4. Hapus kolom lama di migration terpisah (deployment berikutnya)

---

## 5. Environment Strategy

| Environment | Tujuan | Trigger Deploy | Database |
|-------------|--------|----------------|----------|
| **Local** | Development | Manual | Docker PostgreSQL |
| **Staging** | QA + integration test | Merge ke `main` | Managed (shared) |
| **Production** | Live users | Tag `vX.Y.Z` (manual approval) | Managed (dedicated) |

### Environment Variables

Konfigurasi per environment via environment variables:

```bash
# Konfigurasi utama
DATABASE_URL=postgres://...
REDIS_URL=redis://...
NATS_URL=nats://...
JWT_SECRET=...
DEPLOYMENT_MODE=saas  # atau: single

# Feature flags
USE_NATS=true         # false di development

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=...
```

---

## 6. Infrastructure as Code (Terraform)

Semua cloud resources dikelola via Terraform:

```
terraform/
  modules/
    kubernetes/   # GKE / EKS cluster
    database/     # Cloud SQL / RDS
    redis/        # Memorystore / ElastiCache
    networking/   # VPC, subnets, firewall
    iam/          # Service accounts, roles
  environments/
    staging/
    production/
```

Aturan:
- State disimpan di remote backend (GCS bucket / S3)
- Semua perubahan infra melalui PR review — tidak ada manual console changes
- `terraform plan` dijalankan otomatis di PR, `terraform apply` dijalankan manual setelah approval

---

## 7. Secrets Management

- **GKE**: Google Secret Manager, di-mount ke pod via `ExternalSecret` (External Secrets Operator)
- **EKS**: AWS Secrets Manager, di-mount via `SecretProviderClass` (Secrets Store CSI Driver)

Aturan ketat:
- **Tidak ada secret di Git** — tidak ada `.env` file yang di-commit
- **Tidak ada secret di Docker image** — tidak ada `ARG SECRET` di Dockerfile
- **Tidak ada secret di Kubernetes ConfigMap** — hanya Kubernetes Secret atau External Secret

---

## 8. Monitoring & Observability

| Layer | Tool | Fungsi |
|-------|------|--------|
| Traces | OpenTelemetry | Distributed tracing: request path dari HTTP ke DB |
| Metrics | Prometheus | CPU, memory, request latency, error rate, sync metrics |
| Dashboards | Grafana | Visualisasi + alerting |
| Logs | Loki | Centralized log aggregation via Grafana stack |

### Health Endpoints (Wajib per Service)

```
GET /healthz   — Liveness probe: apakah proses berjalan?
GET /readyz    — Readiness probe: apakah siap menerima traffic?
GET /metrics   — Prometheus metrics exposition
```

### Key Metrics yang Dimonitor

```
# Aplikasi
http_request_duration_seconds{path, method, status}
http_requests_total{path, method, status}

# Vernon Sync Engine
sync_lag_seconds{entity_type}           — harus < 30 detik untuk "normal"
sync_error_total{entity_type}           — alert jika > 0 sustained
sync_queue_depth                         — alert jika menumpuk
sync_pending_rows                        — alert jika > threshold

# Database
db_connection_pool_size
db_query_duration_seconds

# NATS
nats_consumer_lag{stream, consumer}     — DLQ monitoring
```

### Alert Conditions

| Alert | Kondisi | Severity |
|-------|---------|----------|
| High error rate | 5xx > 1% per 5 menit | Critical |
| High sync lag | sync_lag > 60 detik | Warning |
| DLQ accumulation | DLQ depth > 100 events | Warning |
| Sync errors | sync_error_total increase > 10/menit | Critical |
| DB connection exhaustion | pool usage > 90% | Warning |

---

## 9. Autoscaling

Go API menggunakan Horizontal Pod Autoscaler (HPA):

```yaml
# Kubernetes HPA
minReplicas: 2
maxReplicas: 10
scaleUpPolicy:
  targetCPUUtilizationPercentage: 70
scaleDownPolicy:
  stabilizationWindowSeconds: 300  # 5 menit sebelum scale down
```

Frontend apps (static files) tidak memerlukan autoscaling — di-serve via CDN.

---

## 10. Rollback Strategy

### Application Rollback

```bash
# Rollback ke image sebelumnya dalam hitungan detik
kubectl rollout undo deployment/sekolahpro-api
```

### Database Rollback

```bash
# Jalankan migration .down.sql
migrate -path ./migrations -database $DATABASE_URL down 1
```

### Import Batch Rollback

Admin dapat rollback import batch dalam 72 jam:

```
POST /api/v1/imports/{batch_id}/rollback
→ Soft-delete semua record dalam batch
→ Trigger Vernon _data cleanup
→ Update batch status = 'rolled_back'
```

---

## Key Decisions

1. **Container-based, immutable image** — image yang sama dari staging ke production; environment hanya dari config

2. **Kubernetes untuk production** — autoscaling, rolling update, health check built-in

3. **GitHub Actions sebagai CI/CD** — standardisasi pipeline; tidak ada CI vendor lock-in yang kuat

4. **Manual approval untuk production deploy** — setiap rilis ke production memerlukan tag `vX.Y.Z` dan explicit approval

5. **golang-migrate, bukan GORM auto-migrate** — migration selalu eksplisit, terdokumentasi, dan reversible

6. **Init container untuk migration** — memastikan schema selalu up-to-date sebelum app baru menerima traffic

7. **Terraform untuk semua infra** — tidak ada manual console changes; perubahan infra melalui PR review

8. **Observability sejak hari pertama** — OpenTelemetry, Prometheus, Grafana, Loki disetup sebelum fitur pertama di-deploy

---

## Constraints & Implications

### Constraints

- Tidak ada migration auto-run di application code — selalu via golang-migrate
- Migration pertama (`00001_uuid_v7_function`) HARUS ada di setiap environment sebelum migration lain dijalankan
- Secret tidak boleh ada di Git, Docker image, atau ConfigMap
- Setiap service wajib memiliki `/healthz` dan `/readyz` endpoint
- Production deploy wajib melalui tag dan manual approval — tidak ada deploy langsung dari `main`

### Implications untuk Engineer

- Setiap perubahan schema database harus disertai migration file yang lengkap (up + down)
- Vernon sync metrics (`sync_lag_seconds`, `sync_error_total`) harus dimonitor aktif — lag yang tinggi mengindikasikan masalah di event bus atau SyncEngine
- Jika menambah service baru, wajib expose `/healthz`, `/readyz`, `/metrics` sebelum merge
- Bulk import (via ADR-017) akan menciptakan burst traffic ke SyncEngine — koordinasikan dengan deployment window yang tepat
- Import rollback hanya tersedia 72 jam — admin harus diedukasi tentang batas waktu ini
