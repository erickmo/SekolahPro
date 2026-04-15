# User Manual — Go + React Boilerplate

Selamat datang di User Manual untuk **Go + React Production Boilerplate**.
Dokumen ini ditujukan untuk tim DevOps/operator yang melakukan deployment dan
operasional sistem, serta end-user yang menggunakan web dashboard.

---

## Daftar Isi

| No. | Dokumen | Deskripsi |
|-----|---------|-----------|
| 01 | [Introduction](./01-introduction.md) | Pengenalan sistem, komponen, diagram arsitektur, dan glossary |
| 02 | [Deployment Guide](./02-deployment-guide.md) | Cara deploy dengan Docker Compose, environment setup, checklist produksi |
| 03 | [Configuration Reference](./03-configuration-reference.md) | Referensi lengkap semua environment variable API dan Web Dashboard |
| 04 | [API Usage Guide](./04-api-usage-guide.md) | Panduan penggunaan API: autentikasi, pagination, filter, contoh curl |
| 05 | [Monitoring & Observability](./05-monitoring-observability.md) | Prometheus, Jaeger, health check, logging, dan rekomendasi alerting |
| 06 | [Web Dashboard Guide](./06-web-dashboard-guide.md) | Panduan pengguna web dashboard: login, navigasi, role, UI patterns |

---

## Cara Membaca Manual Ini

- **Operator / DevOps**: Mulai dari [02 - Deployment Guide](./02-deployment-guide.md),
  lalu [03 - Configuration Reference](./03-configuration-reference.md), dan
  [05 - Monitoring](./05-monitoring-observability.md).

- **Developer yang mengintegrasikan API**: Mulai dari [01 - Introduction](./01-introduction.md),
  lalu [04 - API Usage Guide](./04-api-usage-guide.md).

- **End-user dashboard**: Langsung ke [06 - Web Dashboard Guide](./06-web-dashboard-guide.md).

---

## Versi Komponen

| Komponen | Versi |
|----------|-------|
| Go | 1.25 |
| React | 18 |
| PostgreSQL | 17 |
| Redis | 7 |
| NATS | 2.10 |
| Jaeger | 1.57 |
| Prometheus | 2.51 |

---

## Kontak & Support

Untuk pertanyaan teknis, buka issue di repository atau hubungi tim DevOps internal.
