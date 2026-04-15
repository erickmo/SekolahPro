# 01 — Introduction

## Apa Itu Boilerplate Ini?

**Go + React Production Boilerplate** adalah fondasi aplikasi web siap pakai yang
dirancang untuk membangun aplikasi bisnis skala enterprise. Boilerplate ini
menggabungkan backend Go dengan arsitektur bersih dan frontend React TypeScript
yang modern, sehingga tim dapat langsung fokus membangun fitur bisnis tanpa
harus membangun infrastruktur dari nol.

Boilerplate ini cocok digunakan untuk:

- **Aplikasi internal perusahaan** (dedicated per organisasi — single-tenant)
- **Produk SaaS** yang melayani banyak pelanggan sekaligus (multi-tenant)
- **ERP, dashboard operasional, sistem manajemen** berbasis web

---

## Komponen Sistem

Sistem terdiri dari tiga lapisan utama:

### 1. Go REST API (Backend)

Server API yang menangani semua logika bisnis dan akses data.

- Berjalan di port **8080** (default)
- Mendukung dua pola arsitektur domain: **CQRS** dan **Vernon**
- Menyediakan endpoint REST, dokumentasi Swagger, dan metrics Prometheus
- Autentikasi menggunakan **JWT Bearer Token**

### 2. Web Dashboard (Frontend)

Antarmuka pengguna berbasis browser yang terhubung ke API.

- Dibangun dengan **React 18 + TypeScript**
- Mendukung mode single-tenant dan multi-tenant
- Fitur: login, pilih company, dashboard, manajemen data

### 3. Infrastructure Services (Docker)

Layanan pendukung yang dijalankan via Docker Compose:

| Layanan | Fungsi | Port Default |
|---------|--------|--------------|
| **PostgreSQL** | Database utama — menyimpan semua data bisnis | 5432 |
| **Redis** | Cache layer — mempercepat query yang sering diakses | 6379 |
| **NATS** | Event streaming — komunikasi antar proses secara async | 4222 |
| **Jaeger** | Distributed tracing — melacak perjalanan setiap request | 16686 |
| **Prometheus** | Metrics collection — mengumpulkan data performa sistem | 9090 |

---

## Diagram Arsitektur

```
┌─────────────────────────────────────────────────────────────┐
│                        CLIENT LAYER                         │
│                                                             │
│   Browser / Mobile App / API Consumer                       │
│                                                             │
└────────────────────────────┬────────────────────────────────┘
                             │  HTTP/HTTPS
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                      GO REST API                            │
│                   (port 8080)                               │
│                                                             │
│   ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│   │  Middleware  │  │   Handlers   │  │  Swagger UI     │  │
│   │  (JWT Auth) │  │  (HTTP Layer)│  │  /swagger/*     │  │
│   └──────┬───────┘  └──────┬───────┘  └─────────────────┘  │
│          │                 │                                 │
│   ┌──────▼─────────────────▼──────────────────────────────┐ │
│   │              Business Logic Layer                      │ │
│   │   ┌─────────────────┐   ┌───────────────────────────┐ │ │
│   │   │  CQRS Pattern   │   │    Vernon Pattern          │ │ │
│   │   │  (Command/Query)│   │  (_rels/_data cache)       │ │ │
│   │   └─────────────────┘   └───────────────────────────┘ │ │
│   └───────────────────────────────────────────────────────┘ │
│                                                             │
└──────────┬──────────────┬───────────────────┬──────────────┘
           │              │                   │
           ▼              ▼                   ▼
    ┌────────────┐ ┌────────────┐    ┌─────────────────┐
    │ PostgreSQL │ │   Redis    │    │      NATS       │
    │ (port 5432)│ │ (port 6379)│    │   (port 4222)   │
    └────────────┘ └────────────┘    └─────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   OBSERVABILITY LAYER                        │
│                                                             │
│   ┌──────────────────┐         ┌───────────────────────┐   │
│   │     Jaeger       │         │      Prometheus        │   │
│   │  Distributed     │         │   Metrics Collection   │   │
│   │  Tracing UI      │         │   + Alerting           │   │
│   │  port: 16686     │         │   port: 9090           │   │
│   └──────────────────┘         └───────────────────────┘   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    WEB DASHBOARD                             │
│                (React 18 + TypeScript)                       │
│                                                             │
│   Single-tenant:  /login → /dashboard                       │
│   Multi-tenant:   /login → /choose-company → /c/:code/...   │
└─────────────────────────────────────────────────────────────┘
```

---

## Diagram Hierarki Tenant (Multi-Tenant)

```
Tenant (Level 1)
│  Contoh: PT Vernon Corp (pemilik langganan SaaS)
│
├── Company (Level 2)
│   │  Contoh: PT Vernon Corp Indonesia, PT Vernon Corp Singapore
│   │
│   ├── Branch (Level 3, opsional)
│   │   │  Contoh: Cabang Jakarta, Cabang Surabaya
│   │   │
│   │   └── Warehouse (Level 4, opsional)
│   │          Contoh: Gudang Utama, Gudang Transit
│   │
│   └── Branch lain...
│
└── Company lain...
```

---

## Glossary

| Istilah | Penjelasan |
|---------|-----------|
| **Tenant** | Pelanggan SaaS tingkat tertinggi. Dalam mode single, ada satu tenant yang dikonfigurasi di server. Dalam mode multi, setiap organisasi pelanggan adalah satu tenant. |
| **Company** | Entitas legal / perusahaan di dalam tenant. Satu tenant bisa memiliki beberapa company (misalnya untuk holding group dengan beberapa anak usaha). |
| **Branch** | Cabang atau lokasi fisik di dalam sebuah company. Opsional — tidak semua aplikasi membutuhkan level ini. |
| **Warehouse** | Gudang atau lokasi penyimpanan di dalam branch. Level paling dalam, opsional. |
| **CQRS** | Command Query Responsibility Segregation — pola arsitektur yang memisahkan operasi tulis (Command) dari operasi baca (Query). Digunakan untuk domain dengan logika bisnis kompleks. |
| **Vernon Pattern** | Pola arsitektur read-cache denormalized. Data disimpan dalam dua kolom JSON (`_rels` untuk relasi, `_data` untuk payload). Cocok untuk domain CRUD sederhana dengan banyak JOIN dan operasi baca tinggi. |
| **SyncEngine** | Komponen dalam Vernon Pattern yang mendengarkan event dari NATS/Watermill dan memperbarui cache `_rels`/`_data` secara asinkron setelah ada perubahan data. |
| **JWT** | JSON Web Token — standar token autentikasi yang digunakan API. Token dikirimkan di header `Authorization: Bearer {token}` setiap request. |
| **Single-Tenant Mode** | Mode operasi di mana satu instance API melayani satu organisasi. Konfigurasi tenant/company disimpan di environment variable server. |
| **Multi-Tenant Mode** | Mode operasi di mana satu instance API melayani banyak organisasi (SaaS). Setiap request difilter berdasarkan `tenant_id` dan `company_id` dari JWT. |
| **Two-Phase JWT** | Mekanisme autentikasi multi-tenant dua tahap: Fase 1 menghasilkan token dengan `tenant_id` saja (untuk memilih company), Fase 2 menghasilkan token lengkap dengan `tenant_id + company_id + branch_id`. |
| **Scoped Token** | JWT Fase 2 yang sudah berisi informasi konteks lengkap (tenant + company + opsional branch/warehouse). Digunakan untuk mengakses data operasional. |
| **NATS** | Message broker untuk event streaming antar service. Opsional — sistem bisa berjalan tanpa NATS menggunakan InMemory event bus (khusus development). |
| **Jaeger** | Platform distributed tracing open-source. Memungkinkan melacak perjalanan sebuah request dari handler hingga database query. |
| **Prometheus** | Sistem monitoring dan alerting open-source. Mengumpulkan metrics dari endpoint `/metrics` API secara berkala. |
| **zerolog** | Library logging Go yang menghasilkan output JSON terstruktur. Lebih efisien dari standard library `log`. |
| **sqlc** | Tool yang men-generate type-safe Go code dari SQL query. Mengurangi kemungkinan error runtime akibat ketidakcocokan tipe. |
| **Watermill** | Library Go untuk event-driven architecture. Mengabstraksi berbagai message broker (NATS, InMemory, Kafka, dll). |
