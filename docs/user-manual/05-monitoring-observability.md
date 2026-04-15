# 05 — Monitoring & Observability

Sistem ini dilengkapi tiga pilar observability: **Metrics** (Prometheus),
**Tracing** (Jaeger), dan **Logging** (zerolog JSON). Ketiganya bekerja bersama
untuk memberikan visibilitas penuh terhadap kesehatan dan performa sistem.

---

## Health Check

### Endpoint

```
GET /health
```

Endpoint ini **tidak memerlukan autentikasi** dan dapat diakses oleh load balancer,
container orchestrator (Kubernetes, ECS), atau monitoring tool.

### Response

**Sistem berjalan normal:**
```json
{
  "status": "ok"
}
```

**HTTP Status:** `200 OK`

### Cara Menggunakan

```bash
# Manual check
curl -s http://localhost:8080/health

# Dalam script monitoring (bash)
if curl -sf http://localhost:8080/health > /dev/null; then
  echo "API: UP"
else
  echo "API: DOWN — alert!"
fi
```

### Konfigurasi Health Check di Docker Compose

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 15s
```

---

## Prometheus Metrics

### Endpoint

```
GET /metrics
```

Endpoint ini mengembalikan data dalam format teks Prometheus (OpenMetrics).
**Tidak memerlukan autentikasi** — pastikan diproteksi di level jaringan
(jangan terekspos ke internet publik).

### Cara Mengakses

```bash
# Lihat semua metrics
curl -s http://localhost:8080/metrics

# Filter metrics tertentu
curl -s http://localhost:8080/metrics | grep "http_"

# Lihat request rate
curl -s http://localhost:8080/metrics | grep "http_requests_total"
```

### Prometheus UI

Buka: `http://localhost:9090`

**Query berguna untuk mulai:**

| Query | Deskripsi |
|-------|-----------|
| `up` | Status semua scrape target (1 = up, 0 = down) |
| `http_requests_total` | Total request HTTP yang sudah diproses |
| `http_request_duration_seconds` | Histogram durasi request |
| `go_goroutines` | Jumlah goroutine aktif |
| `go_memstats_alloc_bytes` | Memory yang sedang digunakan |
| `process_cpu_seconds_total` | Total CPU time yang digunakan |

### Konfigurasi Prometheus Scraping

File `api/prometheus.yml` sudah dikonfigurasi untuk scrape API secara otomatis:

```yaml
scrape_configs:
  - job_name: 'boilerplate-api'
    scrape_interval: 15s
    static_configs:
      - targets: ['host.docker.internal:8080']
```

> Di production dengan banyak instance, gunakan service discovery
> (Consul, Kubernetes SD, dll) sebagai pengganti `static_configs`.

---

## Key Metrics yang Perlu Dimonitor

### 1. Availability

| Metric | Query Prometheus | Threshold Alert |
|--------|-----------------|-----------------|
| API uptime | `up{job="boilerplate-api"}` | Alert jika `0` selama > 1 menit |
| Error rate | `rate(http_requests_total{status=~"5.."}[5m])` | Alert jika > 1% dari total request |

### 2. Performance (RED Method)

| Metric | Query Prometheus | Threshold Alert |
|--------|-----------------|-----------------|
| Request rate | `rate(http_requests_total[5m])` | Informational |
| Error rate (5xx) | `rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])` | Alert jika > 1% |
| Duration (p99) | `histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))` | Alert jika > 2 detik |
| Duration (p50) | `histogram_quantile(0.5, rate(http_request_duration_seconds_bucket[5m]))` | Alert jika > 500ms |

### 3. Resources

| Metric | Query Prometheus | Threshold Alert |
|--------|-----------------|-----------------|
| Goroutine count | `go_goroutines` | Alert jika > 1000 (bisa goroutine leak) |
| Memory usage | `go_memstats_alloc_bytes / 1024 / 1024` (MB) | Sesuai kapasitas server |
| GC pause | `go_gc_duration_seconds` | Alert jika p99 > 100ms |

### 4. Database

| Metric | Deskripsi | Threshold Alert |
|--------|-----------|-----------------|
| `db_connection_pool_open` | Koneksi DB yang terbuka | Alert jika mendekati `DB_MAX_OPEN_CONNS` |
| `db_query_duration_seconds` | Durasi eksekusi query | Alert jika p99 > 1 detik |

---

## Jaeger Distributed Tracing

### Mengakses Jaeger UI

Buka: `http://localhost:16686`

### Cara Menggunakan Jaeger UI

**Mencari trace untuk request tertentu:**

1. Di dropdown **Service**, pilih `boilerplate-api`
2. Di **Operation**, pilih endpoint yang ingin dilihat (misalnya `GET /api/v1/products/`)
3. Sesuaikan **Lookback** (misal: `Last 1 hour`)
4. Klik **Find Traces**
5. Klik salah satu trace untuk melihat detail span

**Contoh tampilan trace:**

```
[Trace ID: abc123...]
│
├── GET /api/v1/products/         (total: 45ms)
│   ├── middleware.RequireAuth     (2ms)
│   ├── middleware.ResolveScope    (1ms)
│   ├── handler.List               (42ms)
│   │   ├── querybus.Dispatch      (1ms)
│   │   ├── query.ListProducts     (38ms)
│   │   │   ├── db.query (SELECT)  (35ms)
│   │   │   └── cache.set          (2ms)
│   │   └── json.encode            (1ms)
```

### Trace Context di Log

Setiap log line yang berhubungan dengan sebuah request akan menyertakan
`trace_id` yang sama dengan yang ada di Jaeger. Ini memudahkan korelasi
antara log dan trace:

```json
{"level":"info","trace_id":"abc123def456","span_id":"789xyz","msg":"request completed","duration_ms":45}
```

---

## Logging (zerolog JSON)

### Format Log

API menggunakan zerolog yang menghasilkan log berformat JSON. Setiap baris log
adalah satu objek JSON yang dapat di-parse oleh log aggregation tools.

**Contoh log request:**
```json
{
  "level": "info",
  "time": "2026-04-14T08:30:00Z",
  "service": "boilerplate-api",
  "trace_id": "abc123def456789",
  "span_id": "123abc",
  "method": "GET",
  "path": "/api/v1/products/",
  "status": 200,
  "duration_ms": 45,
  "user_id": "uuid-user",
  "tenant_id": "uuid-tenant",
  "company_id": "uuid-company",
  "msg": "request completed"
}
```

**Contoh log error:**
```json
{
  "level": "error",
  "time": "2026-04-14T08:30:01Z",
  "service": "boilerplate-api",
  "trace_id": "xyz789",
  "error": "record not found",
  "method": "GET",
  "path": "/api/v1/products/nonexistent-id",
  "status": 404,
  "msg": "handler error"
}
```

### Level Log

| Level | Kapan Digunakan |
|-------|-----------------|
| `debug` | Detail eksekusi, query SQL, cache hit/miss. Hanya untuk development |
| `info` | Request masuk/selesai, startup, shutdown, milestone event |
| `warn` | Situasi yang tidak ideal tapi sistem masih berjalan (retry, degraded) |
| `error` | Error yang mempengaruhi request pengguna (5xx, panic recovery) |

Konfigurasi level via env:
```env
LOG_LEVEL=info    # Rekomendasi production
LOG_LEVEL=debug   # Development saja
```

### Membaca Log di Docker

```bash
# Tail log API secara realtime
docker compose logs -f api

# Log dengan timestamp
docker compose logs --timestamps api

# Filter hanya error
docker compose logs api 2>&1 | grep '"level":"error"'

# Parse JSON log dengan jq
docker compose logs api 2>&1 | jq 'select(.level == "error")'
```

### Integrasi dengan Log Aggregation

Untuk production, arahkan log ke sistem aggregasi:

**Loki + Grafana (rekomendasi):**
```yaml
# docker-compose.yml tambahan
logging:
  driver: loki
  options:
    loki-url: "http://loki:3100/loki/api/v1/push"
    loki-labels: "service=boilerplate-api,env=production"
```

**ELK Stack (filebeat):**
```yaml
logging:
  driver: json-file
  options:
    max-size: "100m"
    max-file: "5"
# Konfigurasikan filebeat untuk membaca log Docker JSON
```

---

## Alerting Recommendations

### Alert Rules Prometheus (contoh)

Buat file `alerts.yml` dan load di Prometheus:

```yaml
groups:
  - name: boilerplate-api
    rules:
      # API Down
      - alert: APIDown
        expr: up{job="boilerplate-api"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "API server down"
          description: "boilerplate-api has been down for more than 1 minute"

      # High Error Rate
      - alert: HighErrorRate
        expr: |
          rate(http_requests_total{status=~"5.."}[5m])
          /
          rate(http_requests_total[5m]) > 0.01
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High HTTP error rate"
          description: "Error rate is above 1% for the last 5 minutes"

      # Slow Response Time
      - alert: SlowResponseTime
        expr: |
          histogram_quantile(0.99,
            rate(http_request_duration_seconds_bucket[5m])
          ) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow API response time"
          description: "99th percentile response time is above 2 seconds"

      # High Memory Usage
      - alert: HighMemoryUsage
        expr: go_memstats_alloc_bytes > 500 * 1024 * 1024
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "API memory usage is above 500MB"

      # Goroutine Leak
      - alert: GoroutineLeak
        expr: go_goroutines > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Possible goroutine leak"
          description: "Goroutine count is above 1000"
```

### Notifikasi Alert

Konfigurasikan Alertmanager untuk mengirim notifikasi ke:
- **Slack/Teams:** Webhook integration untuk alert real-time
- **Email:** Untuk alert severity tinggi
- **PagerDuty/OpsGenie:** Untuk on-call rotation pada alert critical

---

## Dashboard Grafana (Opsional)

Jika menggunakan Grafana untuk visualisasi:

1. Tambahkan Prometheus sebagai data source: `http://prometheus:9090`
2. Import dashboard community untuk Go apps: [Grafana Dashboard #13240](https://grafana.com/grafana/dashboards/13240)
3. Buat custom dashboard dengan query-query dari seksi "Key Metrics" di atas

### Metrik Penting untuk Dashboard

- **Panel 1:** Request rate (req/s) — `rate(http_requests_total[1m])`
- **Panel 2:** Error rate (%) — persentase 5xx
- **Panel 3:** Latency percentiles (p50, p95, p99) — histogram
- **Panel 4:** Active goroutines — `go_goroutines`
- **Panel 5:** Memory usage (MB) — `go_memstats_alloc_bytes`
- **Panel 6:** DB connection pool utilization — aktif vs maksimum
