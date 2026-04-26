# ADR-016: Deployment & CI/CD Pipeline

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: Erick Mo, COO + CTO Review Panel

## Context

SekolahPro telah mendefinisikan **95 ADR** yang mencakup seluruh arsitektur sistem — mulai dari backend (Go Clean Architecture + CQRS), frontend (3 React app), hingga domain bisnis sekolah. Namun **belum ada satu baris kode pun yang di-deploy**.

Sebelum coding dimulai, tim membutuhkan:

1. **Strategi deployment** yang jelas agar semua engineer tahu target runtime-nya.
2. **CI/CD pipeline** yang otomatis agar kualitas kode terjaga sejak PR pertama.
3. **Environment strategy** agar ada jalur yang aman dari development ke production.
4. **Observability** agar masalah terdeteksi sebelum user melaporkan.

Tanpa ADR ini, engineer akan membuat asumsi masing-masing tentang cara deploy — menghasilkan inkonsistensi dan technical debt sejak hari pertama.

## Decision

### 1. Container-Based Deployment

Semua service dikemas dalam **Docker image**:

| Service | Base Image | Ukuran Target |
|---------|-----------|---------------|
| Go API | `gcr.io/distroless/static` | < 30 MB |
| app-admin | `nginx:alpine` (static) | < 20 MB |
| app-portal | `nginx:alpine` (static) | < 20 MB |
| app-pos | `nginx:alpine` (static) | < 20 MB |

Multi-stage build digunakan untuk semua image agar ukuran minimal.

### 2. Orchestration

- **Production**: Kubernetes (GKE atau EKS) — autoscaling, rolling update, health check.
- **Local development**: Docker Compose — satu perintah `docker compose up` untuk seluruh stack.
- **Staging**: Kubernetes namespace terpisah di cluster yang sama dengan production.

### 3. CI/CD Pipeline (GitHub Actions)

```
┌─────────────────────────────────────────────────────────┐
│ On Pull Request                                         │
│  ├── lint (golangci-lint + eslint)     ─┐               │
│  ├── test (go test + vitest)            ├── parallel    │
│  └── build (docker build --target test)─┘               │
├─────────────────────────────────────────────────────────┤
│ On Merge to main                                        │
│  ├── build image → push to Container Registry           │
│  └── deploy ke staging (auto)                           │
├─────────────────────────────────────────────────────────┤
│ On Tag (vX.Y.Z)                                         │
│  └── promote staging image → production (manual approve)│
└─────────────────────────────────────────────────────────┘
```

Prinsip: **image yang sama** di-deploy ke staging dan production — hanya konfigurasi environment yang berbeda.

### 4. Database Migration

- Tool: **golang-migrate**
- Dijalankan sebagai **Kubernetes init container** sebelum app container start.
- Migration file disimpan di `migrations/` dengan format `NNNN_description.up.sql` / `.down.sql`.
- Tidak ada auto-migrate di application code — migration selalu eksplisit.

### 5. Environment Strategy

| Environment | Tujuan | Trigger Deploy | Database |
|-------------|--------|----------------|----------|
| Local | Development | Manual | Docker PostgreSQL |
| Staging | QA + integration test | Merge ke `main` | Managed (shared) |
| Production | Live user | Tag `vX.Y.Z` | Managed (dedicated) |

### 6. Infrastructure as Code

- **Terraform** untuk provisioning semua cloud resource (Kubernetes cluster, database, Redis, networking, IAM).
- State disimpan di remote backend (GCS bucket / S3).
- Perubahan infra melalui PR review — tidak ada manual console changes.

### 7. Secrets Management

- **Google Secret Manager** (GKE) atau **AWS Secrets Manager** (EKS).
- Secret di-mount sebagai environment variable via Kubernetes `ExternalSecret`.
- Tidak ada secret di Git, Docker image, atau ConfigMap.

### 8. Monitoring & Observability

| Layer | Tool | Fungsi |
|-------|------|--------|
| Traces | OpenTelemetry | Distributed tracing antar service |
| Metrics | Prometheus | CPU, memory, request latency, error rate |
| Dashboards | Grafana | Visualisasi metrics + alerting |
| Logs | Loki | Centralized log aggregation |

Setiap service wajib expose `/healthz` (liveness) dan `/readyz` (readiness) endpoint.

## Deployment Topology

```
Load Balancer (nginx-ingress / traefik)
├── Go API (2+ replicas, HPA autoscale 2-10)
├── app-admin (static files, CDN-backed)
├── app-portal (static files, CDN-backed)
├── app-pos (static files, CDN-backed)
├── PostgreSQL (managed: Cloud SQL / RDS)
├── Redis (managed: Memorystore / ElastiCache)
└── NATS JetStream (1-3 nodes, StatefulSet)
```

Frontend apps di-serve via CDN (Cloud CDN / CloudFront) dengan cache invalidation otomatis saat deploy.

## Consequences

### Positif

- **Konsistensi environment**: image yang sama jalan di staging dan production.
- **Rollback cepat**: revert ke image tag sebelumnya dalam hitungan detik.
- **Scalability**: HPA otomatis menambah replica saat load tinggi.
- **Observability sejak hari pertama**: masalah terdeteksi sebelum jadi eskalasi.
- **Infra as Code**: semua perubahan infra traceable dan reproducible.

### Negatif

- **Kubernetes learning curve**: tim perlu waktu untuk memahami K8s.
- **Biaya awal lebih tinggi**: managed K8s + managed DB + monitoring stack.
- **Kompleksitas operasional**: butuh engineer yang paham DevOps/SRE.

## Alternatives Considered

### 1. VM-Based (EC2/GCE + Ansible)

- **Pro**: Sederhana, familiar bagi kebanyakan engineer.
- **Kontra**: Manual scaling, deployment lambat, drift konfigurasi. Tidak cocok untuk SaaS multi-tenant yang perlu autoscale.

### 2. Serverless (Cloud Run / Lambda)

- **Pro**: Zero infrastructure management, pay-per-request.
- **Kontra**: Cold start latency, vendor lock-in kuat, NATS JetStream tidak bisa jalan di serverless. WebSocket/SSE untuk real-time fitur sulit diimplementasi.

### 3. PaaS (Fly.io / Railway)

- **Pro**: Developer experience sangat baik, deploy dalam menit.
- **Kontra**: Kurang kontrol untuk compliance (data residency Indonesia), biaya tidak predictable saat scale, limitasi networking untuk inter-service communication.

### 4. Hybrid (Cloud Run + GKE)

- **Pro**: Frontend di Cloud Run (simple), backend di GKE (full control).
- **Kontra**: Menambah kompleksitas operasional dengan dua platform berbeda. Ditunda sebagai opsi future optimization.

## References

- ADR-001: Go Clean Architecture + CQRS
- ADR-005: Event Bus (In-Memory + NATS)
- ADR-008: React + Vite + CSS Modules
