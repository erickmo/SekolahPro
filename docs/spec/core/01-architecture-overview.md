# 01 — Ikhtisar Arsitektur Sistem SekolahPro

Dokumen ini mendeskripsikan arsitektur sistem SekolahPro secara menyeluruh: bagaimana Go Clean Architecture, CQRS, dan Vernon Denormalized Read-Cache Pattern digabungkan menjadi satu kesatuan yang kohesif. Dokumen ini menjadi referensi arsitektur utama bagi seluruh engineer yang bekerja pada proyek ini.

---

## ADR References

| ADR | Judul | Relevansi |
|-----|-------|-----------|
| ADR-001 | Go Clean Architecture + CQRS | Fondasi struktur layer dan pemisahan command/query |
| ADR-002 | Vernon Denormalized Read-Cache | Pola read-cache JSONB untuk query performa tinggi |
| ADR-003 | Hybrid CQRS + Vernon — Decision Criteria | Kriteria kapan pakai Standard CQRS vs Vernon |
| ADR-006 | Uber FX Dependency Injection | Wiring dependency seluruh aplikasi |

---

## Gambaran Umum

SekolahPro adalah platform SaaS dual-purpose: **Management Sekolah** (akademik) dan **Management Koperasi Sekolah** (keuangan). Keduanya berjalan di atas satu backend Go monolith yang bersifat multi-tenant dengan 4-level hierarki organisasi (Tenant → Company → Branch → Warehouse, lihat `03-multi-tenant-design.md`).

Arsitektur backend dibangun di atas tiga pilar:

1. **Go Clean Architecture** — memastikan separasi of concerns dan testabilitas tinggi
2. **CQRS** — memisahkan jalur tulis (command) dari jalur baca (query) secara eksplisit
3. **Vernon Pattern** — mengoptimalkan query baca domain-domain read-heavy melalui denormalized JSONB cache

---

## Struktur Layer (Go Clean Architecture)

```
internal/
├── domain/                  # Layer 1 — Enterprise Business Rules
│   ├── entity/              # Pure domain objects (no framework dependency)
│   ├── valueobject/         # Immutable value types (UUID, Money, Email)
│   └── event/               # Domain event definitions
│
├── usecase/                 # Layer 2 — Application Business Rules
│   ├── command/             # Write operations: CreateStudent, UpdateTeacher, ...
│   ├── query/               # Read operations: GetStudent, ListStudents, ...
│   ├── sync/                # Vernon SyncEngine event handlers
│   └── port/                # Interfaces (Repository, EventBus, Reader)
│
├── adapter/                 # Layer 3 — Interface Adapters
│   ├── handler/             # HTTP handlers (Chi router, NO business logic)
│   ├── repository/          # PostgreSQL implementations (sqlx/sqlc)
│   ├── reader/              # Vernon read repositories (zero-JOIN queries)
│   └── eventhandler/        # Event subscriber implementations
│
└── infrastructure/          # Layer 4 — Frameworks & Drivers
    ├── db/                  # PostgreSQL connection, golang-migrate
    ├── cache/               # Redis client
    ├── messaging/           # NATS JetStream + InMemory event bus
    ├── telemetry/           # OpenTelemetry, Prometheus
    └── fx/                  # Uber FX modules & application bootstrap
```

### Aturan Dependensi (Dependency Rule)

Dependensi hanya boleh mengarah ke dalam (inward):

```
infrastructure → adapter → usecase → domain
```

- `domain` tidak boleh mengimport package apapun di luar Go standard library
- Interface didefinisikan di `usecase/port/`, implementasi di `adapter/` atau `infrastructure/`
- `handler/` tidak boleh mengandung business logic — hanya parsing request, delegasi ke usecase, formatting response

---

## CQRS: Pemisahan Command dan Query

CQRS (Command Query Responsibility Segregation) adalah pola inti di SekolahPro. Setiap operasi diklasifikasikan sebagai salah satu dari dua jenis:

### Command Side (Write)

Menangani semua operasi yang memodifikasi state:

```
Command Struct → Command Handler → Repository (write) → Event Bus
```

Contoh: `CreateStudentCommand`, `ActivateAcademicYearCommand`, `UpdateTeacherCommand`

Karakteristik:
- Validasi business rule ada di handler, bukan di HTTP handler
- Setelah menyimpan ke DB, publish domain event ke event bus
- Return nilai minimal (biasanya ID entitas yang dibuat)

### Query Side (Read)

Menangani semua operasi yang membaca state tanpa memodifikasinya:

```
Query Struct → Query Handler → Reader (read) → Result DTO
```

Contoh: `GetStudentQuery`, `ListClassRoomsQuery`, `GetAcademicYearActiveQuery`

Karakteristik:
- Menggunakan interface `Reader` terpisah dari `Repository`
- Pada domain Standard CQRS: query via sqlc dengan maksimal 2 JOIN
- Pada domain Vernon: query zero-JOIN menggunakan `_data` JSONB column

---

## Vernon Pattern: Denormalized Read-Cache

Vernon Pattern adalah optimasi untuk domain yang read-heavy dengan banyak JOIN. Setiap tabel yang menggunakan Vernon memiliki dua kolom tambahan:

```sql
_rels  JSONB NOT NULL DEFAULT '{}'  -- Foreign key IDs semua relasi
_data  JSONB NOT NULL DEFAULT '{}'  -- Snapshot denormalized dari data relasi
```

### Alur Data Vernon

```
Write Path:
Client → HTTP Handler → Command Handler → Repository.SaveWithVernon()
         │                                     │
         │                               Build _rels & _data
         │                               dari relasi terkait
         └──────────────────────────────→ INSERT ke DB

Read Path:
Client → HTTP Handler → Query Handler → VernonReader.ListByTenant()
                                              │
                                        SELECT id, ..., _data
                                        FROM table
                                        WHERE tenant_id = $1
                                        (ZERO JOIN)
                                              │
                                        Map _data JSONB → Result DTO
                                              └──────────────→ Response

Sync Path (Async):
Parent Entity Updated → Publish Event → SyncEngine Handler
                                              │
                                        UPDATE child._data
                                        WHERE child._rels->>'parent_id' = $1
```

---

## Diagram Arsitektur Lengkap

```
┌─────────────────────────────────────────────────────────────────┐
│                         HTTP Layer                               │
│  Chi Router + JWT Middleware + Scope Middleware (ADR-004)        │
└────────────────────────────┬────────────────────────────────────┘
                             │
         ┌───────────────────┴───────────────────┐
         │ adapter/handler/                       │
         │ (HTTP parsing, response formatting)    │
         └───────────────────┬───────────────────┘
                             │
         ┌───────────────────┴───────────────────┐
         │ usecase/command/ + usecase/query/      │
         │ (BUSINESS LOGIC LIVES HERE)            │
         │ + usecase/port/ (interfaces)           │
         └──────┬────────────┬──────────┬────────┘
                │            │          │
         ┌──────┘    ┌───────┘   ┌──────┘
         ▼            ▼           ▼
  ┌──────────┐ ┌──────────┐ ┌──────────────┐
  │Repository│ │  Reader  │ │  EventBus    │
  │(Write)   │ │(Read)    │ │  Interface   │
  └────┬─────┘ └────┬─────┘ └──────┬───────┘
       │             │               │
  ┌────┘             │          ┌────┘
  ▼                  ▼          ▼
 PostgreSQL    PostgreSQL   InMemory (dev)
 (write)       _data cache  NATS JetStream
               (read, 0 JOIN) (prod)
```

---

## Hybrid Model: Standard CQRS vs Vernon

Tidak semua domain menggunakan Vernon. Keputusan didasarkan pada metrik terukur:

| Kriteria | Standard CQRS (sqlc) | Vernon Pattern |
|----------|----------------------|----------------|
| Jumlah JOIN terumit | ≤ 2 JOIN | ≥ 3 JOIN |
| Read:Write ratio | < 10:1 | ≥ 10:1 |
| Consistency requirement | Strong (immediate) | Eventual OK |
| Dataset size | < 100K rows | ≥ 100K rows |
| Latency target | < 500ms OK | < 100ms required |

### Domain Classification di SekolahPro

```
Standard CQRS (sqlc):
├── auth/           — login, refresh token (strong consistency)
├── users/          — 1-2 JOIN, kebutuhan immediate consistency
├── permissions/    — small dataset, write-driven
└── audit_logs/     — write-heavy, append-only

Vernon Pattern:
├── academic_years/ — referenced 10+ domain, read-heavy
├── class_rooms/    — referenced 6+ domain, per academic year
├── teachers/       — referenced 7+ domain, nama di rapor
├── students/       — core entity, 8+ dependent domain
└── [domain lain dengan ≥3 JOIN]
```

---

## Key Decisions

1. **Clean Architecture sebagai fondasi** — dependency rule yang ketat memastikan business logic bisa diuji tanpa database, HTTP server, atau external dependency apapun.

2. **CQRS bukan optional** — pemisahan command/query bukan sekadar gaya kode. Command path dan read path memiliki kebutuhan yang berbeda (consistency vs latency) dan dioptimasi secara independen.

3. **Hybrid Vernon** — Vernon diterapkan hanya pada domain yang memiliki ≥3 JOIN di query listing/detail dan read:write ratio ≥10:1. Domain sederhana tetap menggunakan Standard CQRS.

4. **Uber FX untuk lifecycle management** — semua dependency di-wire melalui FX module, memberikan graceful startup/shutdown otomatis dan fail-fast jika ada dependency graph yang invalid.

5. **No business logic di handler** — HTTP handler hanya bertanggung jawab: parse request → delegate ke usecase → format response. Ini adalah constraint yang tidak boleh dilanggar.

---

## Constraints & Implications

### Constraints

- `domain/` layer tidak boleh mengimport apapun di luar Go standard library
- Interface repository dan event bus harus didefinisikan di `usecase/port/`, BUKAN di `adapter/`
- Command handler tidak boleh memanggil query handler (no mixed CQRS)
- Setiap domain Vernon harus memiliki SyncEngine handler yang terdaftar di `SyncRegistry`
- Function maksimal 40 baris; tidak ada God Class (>300 baris / >5 tanggung jawab)

### Implications untuk Domain Baru

Ketika menambah domain baru, engineer harus:

1. **Tentukan klasifikasi**: Standard CQRS atau Vernon? (gunakan decision matrix ADR-003)
2. **Buat port interface** di `usecase/port/` sebelum implementasi repository
3. **Buat FX module** untuk domain tersebut
4. **Jika Vernon**: daftarkan di `SyncRegistry` dan implementasikan SyncEngine handler
5. **Test tanpa database**: unit test usecase harus bisa jalan dengan mock repository

### Implications untuk Query Performance

- Domain Standard CQRS: maksimal 2 JOIN — jika query butuh lebih, pertimbangkan upgrade ke Vernon
- Domain Vernon: query listing harus zero-JOIN (semua data dari `_data`)
- Read dari `_data` bersifat eventually consistent — aplikasi harus toleransi stale data < 30 detik

### Implications untuk Testing

- Standard CQRS domain: unit test usecase dengan mock `Repository`
- Vernon domain: unit test usecase dengan mock `Repository` + `Reader`; integration test SyncEngine memerlukan DB
- SyncEngine handler harus di-test secara terpisah dengan test event yang spesifik
