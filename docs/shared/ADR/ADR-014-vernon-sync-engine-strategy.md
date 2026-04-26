# ADR-014: Vernon Sync Engine Strategy

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: Erick Mo, CTO Review Panel

## Context

SekolahPro menggunakan Vernon Pattern (ADR-002) di **175+ tabel** dengan kolom JSONB `_rels` dan `_data` untuk denormalized read-cache. Setiap tabel juga memiliki:

| Kolom | Tipe | Fungsi |
|-------|------|--------|
| `_rels` | JSONB | Foreign key IDs dari semua relasi |
| `_data` | JSONB | Denormalized snapshot data relasi |
| `_sync_status` | VARCHAR(10) | Status sinkronisasi: `synced`, `pending`, `error` |
| `_sync_version` | BIGINT | Monotonically increasing version number |

### Masalah

Ketika parent entity berubah (contoh: nama siswa di-update), **semua tabel child yang meng-cache data tersebut di `_data` harus di-update**. Dengan 175+ tabel dan cross-domain relations:

- **Tanpa strategi**: Sync menjadi O(N) cascading updates dalam satu transaksi, menyebabkan lock contention dan timeout.
- **DB trigger approach**: Sulit di-debug, tidak bisa di-test secara unit, dan menyembunyikan side-effect.
- **Synchronous update**: Satu write ke `students` bisa memicu 15+ UPDATE ke tabel lain, memperlambat response time.

Contoh dependency chain:

```
students → student_academic_records → student_grades → student_rapor
         → student_attendance
         → student_finance_invoices
         → student_discipline
         → student_class_placements
```

Satu perubahan nama siswa bisa cascade ke **8+ tabel**.

## Decision

### 1. Event-Driven Sync (BUKAN DB Triggers)

Sync dipicu oleh domain event melalui event bus (ADR-005), **bukan** DB trigger atau direct call.

```go
// Ketika entity di-update, publish event
type EntityUpdatedEvent struct {
    EntityType string    `json:"entity_type"` // "student"
    EntityID   string    `json:"entity_id"`
    TenantID   string    `json:"tenant_id"`
    Fields     []string  `json:"fields"`      // field yang berubah
    Version    int64     `json:"version"`
    OccurredAt time.Time `json:"occurred_at"`
}
```

### 2. Async Batch Processing

Sync berjalan **asinkron** via event bus, terpisah dari transaksi write utama.

```
Entity Updated → Event Published → SyncWorker picks up →
Query dependents dari registry → Batch UPDATE _data →
Update _sync_status = 'synced' → Publish SyncCompleted
```

Write response langsung dikembalikan ke client tanpa menunggu sync selesai.

### 3. Selective Field Sync

Hanya field yang dideklarasikan di Vernon descriptor `Fields` array yang di-sync, **bukan** seluruh entity.

```go
// Descriptor mendefinisikan field mana yang di-cache
var StudentDescriptor = vernon.Descriptor{
    Table:  "students",
    Fields: []string{"name", "nis", "status"}, // hanya ini yang di-sync
}
```

Jika field yang berubah tidak ada di `Fields` array tabel dependent, **sync di-skip**.

### 4. Batch Coalescing

Multiple updates ke entity yang sama dalam window **100ms** di-coalesce menjadi satu sync operation.

```go
const CoalesceWindow = 100 * time.Millisecond
```

Menghindari N sync untuk N rapid updates (contoh: bulk import siswa).

### 5. Priority Levels

| Priority | Contoh | SLA |
|----------|--------|-----|
| `critical` | Data keuangan (invoices, pembayaran) | < 1 detik |
| `normal` | Data display (nama, label) | < 30 detik |

Worker `critical` dan `normal` berjalan di goroutine pool terpisah.

### 6. Dependency Registry

Registry Go yang memetakan "ketika tabel X berubah, sync tabel [Y, Z, W]".

```go
var SyncRegistry = map[string][]SyncTarget{
    "students": {
        {Table: "student_academic_records", Priority: "normal"},
        {Table: "student_grades",          Priority: "normal"},
        {Table: "student_attendance",      Priority: "normal"},
        {Table: "student_finance_invoices", Priority: "critical"},
        {Table: "student_class_placements", Priority: "normal"},
    },
    "academic_years": {
        {Table: "student_academic_records", Priority: "normal"},
        {Table: "student_grades",          Priority: "normal"},
        {Table: "student_rapor",           Priority: "normal"},
    },
    // ... 175+ entries
}

type SyncTarget struct {
    Table    string
    Priority string // "critical" | "normal"
}
```

### 7. Sync Status Tracking

```sql
-- Setelah entity berubah, dependents ditandai pending
UPDATE student_academic_records
SET _sync_status = 'pending', _sync_version = _sync_version + 1
WHERE _rels->>'student_id' = $1;

-- Setelah sync selesai
UPDATE student_academic_records
SET _data = jsonb_set(_data, '{student}', $2::jsonb),
    _sync_status = 'synced',
    _sync_version = _sync_version + 1
WHERE _rels->>'student_id' = $1 AND _sync_status = 'pending';
```

### 8. Monitoring

Prometheus metrics yang di-expose:

| Metric | Tipe | Deskripsi |
|--------|------|-----------|
| `sync_lag_seconds` | Histogram | Waktu antara event published dan sync completed |
| `sync_error_total` | Counter | Jumlah sync yang gagal, per entity_type |
| `sync_queue_depth` | Gauge | Jumlah event menunggu di queue |
| `sync_pending_rows` | Gauge | Jumlah row dengan `_sync_status = 'pending'` |

## Failure Handling

### Retry Strategy

| Attempt | Delay | Keterangan |
|---------|-------|------------|
| 1 | 1 detik | Retry pertama |
| 2 | 5 detik | Retry kedua |
| 3 | 30 detik | Retry terakhir |

### Setelah Max Retries

1. Set `_sync_status = 'error'` pada row yang gagal.
2. Kirim event ke **Dead Letter Queue (DLQ)**.
3. Log error dengan context lengkap (entity_type, entity_id, error detail).
4. Alert via monitoring jika DLQ depth > threshold.

### Prinsip: Stale > Down

Data stale di `_data` **lebih baik daripada** system down. Read-cache boleh stale sementara — user tetap bisa membaca data (meskipun belum terbaru).

### Manual Resync Endpoint

```
POST /api/v1/admin/sync/resync/{entity_type}/{id}
```

Untuk admin: trigger manual resync satu entity beserta semua dependent-nya. Berguna untuk recovery dari `_sync_status = 'error'`.

## Consequences

### Positif

- **Write tetap cepat**: Response time tidak terdampak oleh jumlah dependents.
- **Fault-tolerant**: Kegagalan sync tidak menggagalkan operasi utama.
- **Observable**: Semua metric dan status bisa dimonitor secara real-time.
- **Scalable**: Worker pool bisa di-scale horizontal per priority level.
- **Testable**: Event-driven approach memungkinkan unit test tanpa DB trigger.

### Negatif

- **Eventual consistency**: Read setelah write mungkin menampilkan data lama (window < 30 detik).
- **Kompleksitas operasional**: Perlu monitoring DLQ dan alerting untuk sync errors.
- **Registry maintenance**: Setiap relasi baru harus didaftarkan di `SyncRegistry`.

## Alternatives Considered

| Alternatif | Alasan Ditolak |
|------------|----------------|
| **DB Triggers** | Sulit di-debug, tidak bisa di-unit-test, menyembunyikan side-effect, tidak bisa prioritas. |
| **Synchronous cascading UPDATE** | O(N) updates dalam satu transaksi = lock contention dan slow writes. |
| **Materialized Views** | Tidak mendukung selective field sync, sulit di-customize per domain, refresh mahal. |
| **Application-level JOIN on read** | Mengembalikan masalah awal yang diselesaikan Vernon Pattern (ADR-002). |
| **Change Data Capture (Debezium)** | Infrastruktur tambahan terlalu berat untuk fase awal. Bisa dipertimbangkan di masa depan. |

## References

- ADR-002: Vernon Denormalized Read-Cache Pattern
- ADR-005: Event Bus Abstraction (InMemory / NATS JetStream)
