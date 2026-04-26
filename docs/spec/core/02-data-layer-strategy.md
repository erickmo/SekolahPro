# 02 — Strategi Data Layer SekolahPro

Dokumen ini menjabarkan seluruh keputusan desain lapisan data: pola Vernon denormalized read-cache, strategi SyncEngine, konvensi UUID v7, isolasi data multi-tenant, dan pendekatan migrasi data. Ini adalah referensi utama untuk engineer yang bekerja dengan skema database, query, atau data pipeline.

---

## ADR References

| ADR | Judul | Relevansi |
|-----|-------|-----------|
| ADR-002 | Vernon Denormalized Read-Cache | Pola _rels/_data JSONB, zero-JOIN read |
| ADR-003 | Hybrid CQRS + Vernon — Decision Criteria | Kapan pakai Vernon vs Standard CQRS |
| ADR-007 | UUID v7 sebagai Primary Key | Strategi primary key seluruh sistem |
| ADR-014 | Vernon Sync Engine Strategy | Mekanisme sinkronisasi _data secara async |
| ADR-017 | Strategi Migrasi Data & Import | Import awal, dry-run, batch processing |

---

## 1. UUID v7 sebagai Primary Key Universal

Semua tabel menggunakan UUID v7 sebagai primary key, di-generate via fungsi PostgreSQL kustom `uuid_generate_v7()`.

### Mengapa UUID v7?

| Kriteria | BIGSERIAL | UUID v4 | UUID v7 |
|----------|-----------|---------|---------|
| Globally unique (tanpa round-trip DB) | Tidak | Ya | Ya |
| Time-sortable | Ya | Tidak | Ya |
| B-tree friendly (monotonic) | Ya | Tidak | Ya |
| Privacy (tidak mengekspos data internal) | Tidak | Ya | Ya |
| RFC standard | — | RFC 4122 | RFC 9562 |

Benchmark internal pada 10 juta row: UUID v7 memiliki throughput INSERT ~73% lebih tinggi dan index size 58% lebih kecil dibandingkan UUID v4.

### Fungsi PostgreSQL

```sql
-- Harus ada di setiap environment sebelum migration apapun
CREATE OR REPLACE FUNCTION uuid_generate_v7() RETURNS UUID ...;
```

Fungsi ini di-setup via migration pertama (`00001_uuid_v7_function.sql`). Jika fungsi tidak ada, INSERT akan gagal.

### Konvensi Schema

```sql
-- Semua tabel menggunakan DEFAULT ini
id UUID PRIMARY KEY DEFAULT uuid_generate_v7()

-- DILARANG:
-- id SERIAL PRIMARY KEY
-- id UUID DEFAULT gen_random_uuid()   -- UUID v4
-- id UUID DEFAULT uuid_generate_v4()  -- UUID v4
```

### Application-Layer Generation

Untuk event sourcing dan pre-generated ID (ID diketahui sebelum INSERT ke DB):

```go
import "github.com/google/uuid"

func NewID() uuid.UUID {
    id, _ := uuid.NewV7()
    return id
}
```

ID yang di-generate di application layer tetap valid di PostgreSQL karena UUID v7 adalah standar RFC 9562.

---

## 2. Vernon Denormalized Read-Cache Pattern

### Masalah yang Diselesaikan

Domain-domain read-heavy seperti `students`, `class_rooms`, `academic_years` memerlukan data dari banyak relasi:

- Menampilkan profil siswa memerlukan data dari: `class_rooms`, `academic_years`, `teachers` (wali kelas)
- Listing 50 siswa dengan 5+ JOIN pada dataset besar (>100K rows) menghasilkan latency >500ms
- Menambah field baru ke read model memerlukan ALTER TABLE yang bisa lock tabel production

### Solusi: Dua Kolom JSONB Tambahan

Setiap tabel domain Vernon memiliki:

| Kolom | Tipe | Fungsi |
|-------|------|--------|
| `_rels` | JSONB | Menyimpan semua foreign key ID (untuk SyncEngine) |
| `_data` | JSONB | Menyimpan snapshot denormalized dari data relasi |
| `_sync_status` | VARCHAR(10) | Status sinkronisasi: `synced`, `pending`, `error` |
| `_sync_version` | BIGINT | Version number yang monotonically increasing |

### Struktur _rels

```json
{
  "academic_year_id": "018f4c2a-1234-7000-8000-000000000001",
  "class_room_id":    "018f4c2a-5678-7000-8000-000000000002",
  "homeroom_teacher_id": "018f4c2a-9abc-7000-8000-000000000003"
}
```

`_rels` digunakan oleh SyncEngine untuk menemukan semua entitas yang perlu di-update ketika suatu relasi berubah.

### Struktur _data

```json
{
  "academic_year": {
    "id":        "018f4c2a-1234-7000-8000-000000000001",
    "name":      "2025/2026",
    "is_active": true
  },
  "class_room": {
    "id":        "018f4c2a-5678-7000-8000-000000000002",
    "name":      "VII-A",
    "grade_level": "7"
  },
  "homeroom_teacher": {
    "id":       "018f4c2a-9abc-7000-8000-000000000003",
    "full_name": "Bu Siti",
    "nip":      "198501012010012001"
  }
}
```

`_data` adalah snapshot denormalized yang langsung bisa digunakan untuk response API tanpa JOIN.

### Index GIN untuk JSONB

```sql
CREATE INDEX idx_students_rels ON students USING GIN (_rels);
CREATE INDEX idx_students_data ON students USING GIN (_data);
```

Index GIN memungkinkan path queries yang efisien pada kolom JSONB.

### Query Zero-JOIN (Read Path)

```sql
-- Listing siswa — single table scan, zero JOIN
SELECT id, nis, full_name, status,
       _data->>'class_room'    AS class_room,
       _data->>'academic_year' AS academic_year,
       _data->>'homeroom_teacher' AS homeroom_teacher
FROM students
WHERE tenant_id = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT 50;
```

Benchmark internal: 450ms (7-way JOIN) → 12ms (Vernon zero-JOIN) pada dataset 500K row.

---

## 3. Vernon Sync Engine

### Konsep

Ketika parent entity berubah (misal: nama guru diupdate), SyncEngine secara async men-update semua entitas child yang meng-cache data tersebut di `_data`.

### Arsitektur Sync (Event-Driven)

```
Entity Updated
     │
     ▼
Publish EntityUpdatedEvent
  {entity_type, entity_id, fields_changed, tenant_id, version}
     │
     ▼
Event Bus (InMemory dev / NATS JetStream prod)
     │
     ▼
SyncWorker picks up event
     │
     ▼
Lookup SyncRegistry: tabel apa yang perlu di-sync?
     │
     ▼
Batch UPDATE _data
WHERE _rels->>'parent_id' = entity_id
     │
     ▼
UPDATE _sync_status = 'synced'
     │
     ▼
Publish SyncCompletedEvent
```

Prinsip kritis: **sync berjalan async, terpisah dari transaksi write utama**. Write response dikembalikan ke client tanpa menunggu sync selesai.

### SyncRegistry

Registry Go yang memetakan "ketika tabel X berubah, sync tabel [Y, Z, W]":

```go
var SyncRegistry = map[string][]SyncTarget{
    "teachers": {
        {Table: "class_rooms",              Priority: "normal"},
        {Table: "student_grades",           Priority: "normal"},
        {Table: "student_attendances",      Priority: "normal"},
    },
    "academic_years": {
        {Table: "class_rooms",              Priority: "normal"},
        {Table: "student_academic_records", Priority: "normal"},
        {Table: "student_grades",           Priority: "normal"},
        {Table: "student_finance_invoices", Priority: "critical"},
    },
    "students": {
        {Table: "student_academic_records", Priority: "normal"},
        {Table: "student_attendances",      Priority: "normal"},
        {Table: "student_finance_invoices", Priority: "critical"},
        {Table: "student_class_placements", Priority: "normal"},
    },
    // ... semua entitas Vernon
}
```

Setiap relasi baru **wajib** didaftarkan di registry ini.

### Priority Levels

| Priority | Contoh | SLA |
|----------|--------|-----|
| `critical` | Data keuangan (invoices, pembayaran) | < 1 detik |
| `normal` | Data display (nama, label, kelas) | < 30 detik |

Worker `critical` dan `normal` berjalan di goroutine pool terpisah.

### Batch Coalescing

Multiple update ke entitas yang sama dalam window 100ms di-coalesce menjadi satu sync operation:

```go
const CoalesceWindow = 100 * time.Millisecond
```

Mencegah N sync untuk N rapid updates (contoh: bulk import 1000 siswa).

### Selective Field Sync

Hanya field yang dideklarasikan di Vernon descriptor yang di-sync:

```go
var StudentDescriptor = vernon.Descriptor{
    Table:  "students",
    Fields: []string{"full_name", "nis", "status"},
}
```

Jika field yang berubah tidak ada di `Fields`, sync di-skip untuk efisiensi.

### Failure Handling

| Attempt | Delay |
|---------|-------|
| 1 | 1 detik |
| 2 | 5 detik |
| 3 | 30 detik |

Setelah max retries:
1. Set `_sync_status = 'error'` pada row yang gagal
2. Kirim event ke Dead Letter Queue (DLQ)
3. Alert via monitoring jika DLQ depth > threshold

Prinsip: **Stale > Down** — data stale lebih baik daripada system down.

### Manual Resync Endpoint

```
POST /api/v1/admin/sync/resync/{entity_type}/{id}
```

Untuk admin: trigger manual resync satu entitas beserta semua dependent-nya.

### Monitoring Metrics

| Metric | Tipe |
|--------|------|
| `sync_lag_seconds` | Histogram — waktu antara event published dan sync completed |
| `sync_error_total` | Counter — jumlah sync gagal per entity_type |
| `sync_queue_depth` | Gauge — jumlah event menunggu di queue |
| `sync_pending_rows` | Gauge — jumlah row dengan `_sync_status = 'pending'` |

---

## 4. Konvensi Schema Multi-Tenant

### Scope Columns per Tipe Tabel

```sql
-- Tabel master (data referensi) — hanya tenant_id
CREATE TABLE academic_years (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    company_id  UUID NOT NULL REFERENCES companies(id),
    ...
);

-- Tabel operasional — full hierarchy (sesuai kebutuhan domain)
CREATE TABLE student_finance_invoices (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id),
    company_id   UUID NOT NULL REFERENCES companies(id),
    ...
);
```

### Index Wajib

Setiap tabel yang memiliki `tenant_id` wajib memiliki index pada kolom scope:

```sql
CREATE INDEX idx_{table}_tenant    ON {table}(tenant_id);
CREATE INDEX idx_{table}_company   ON {table}(tenant_id, company_id);
```

### Vernon-Specific Columns

Semua tabel Vernon mengikuti pola ini:

```sql
-- Vernon read-cache
_rels           JSONB NOT NULL DEFAULT '{}',
_data           JSONB NOT NULL DEFAULT '{}',
_sync_status    TEXT NOT NULL DEFAULT 'synced',
_sync_version   BIGINT NOT NULL DEFAULT 0,

-- Standard timestamps
created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
deleted_at      TIMESTAMPTZ  -- soft delete
```

---

## 5. Strategi Migrasi Data & Import

### Tantangan Onboarding

Sekolah yang mengadopsi SekolahPro memiliki data existing di berbagai format: Excel, export Dapodik, sistem lama (Jibas, AdminSekolah). Tanpa tools import, onboarding satu sekolah bisa memakan 2-4 minggu.

### Jalur Import: Template-Based Excel/CSV

Admin sekolah mengunduh template Excel per entitas, mengisi data, lalu mengupload:

| Entitas | Template | Kolom Kunci |
|---------|----------|-------------|
| Siswa | `template_siswa.xlsx` | NIS, NISN, nama, kelas, jenis kelamin, tanggal lahir |
| Guru/Staff | `template_guru.xlsx` | NIP/NIK, nama, jabatan, mata pelajaran, email |
| Kelas | `template_kelas.xlsx` | kode_kelas, nama, tingkat, wali_kelas_NIP, kapasitas |
| Tahun Ajaran | `template_tahun_ajaran.xlsx` | kode, nama, tanggal_mulai, tanggal_selesai |
| Tagihan SPP | `template_tagihan.xlsx` | NIS_siswa, tipe_biaya, bulan, tahun, nominal |

### Alur Import

```
Download Template → Isi Data → Upload File
        │
        ▼
Dry-Run Validasi (WAJIB)
  - Parse seluruh baris
  - Validasi format, tipe data, constraint
  - Deteksi duplikat terhadap data existing
  - Return error report (tanpa menyimpan apapun)
        │
        ▼
Admin Review: total rows, valid, error, warning, duplikat
        │
        ▼
Konfirmasi Import (setelah review)
        │
        ▼
Batch Processing (async jika > 1000 rows)
        │
        ▼
Vernon Sync Post-Import
  - Emit StudentImported event (batched, max 100/event)
  - SyncEngine build _data JSONB untuk setiap record
        │
        ▼
Laporan Verifikasi + Rollback option (72 jam)
```

### Import Batch Table

```sql
CREATE TABLE import_batches (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,
    entity_type     VARCHAR(30) NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    uploaded_by     UUID NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    total_rows      INTEGER NOT NULL DEFAULT 0,
    processed_rows  INTEGER NOT NULL DEFAULT 0,
    success_rows    INTEGER NOT NULL DEFAULT 0,
    error_rows      INTEGER NOT NULL DEFAULT 0,
    error_report    JSONB,
    batch_record_ids JSONB,   -- array UUID record yang dibuat
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ,
    CONSTRAINT chk_status CHECK (status IN (
        'pending', 'validating', 'validated', 'importing',
        'completed', 'failed', 'rolled_back'
    ))
);
```

### Conflict Resolution

| Entitas | Duplicate Key | Opsi |
|---------|--------------|------|
| Siswa | NIS atau NISN | skip, overwrite, merge |
| Guru | NIP atau email | skip, overwrite, merge |
| Kelas | kode + tahun ajaran | skip, overwrite |

Strategi dipilih admin sebelum konfirmasi import.

### Rollback

- `batch_record_ids` menyimpan semua UUID record yang dibuat
- Rollback = soft-delete semua record dalam batch
- Batas waktu: 72 jam setelah import selesai
- Rollback juga trigger Vernon `_data` cleanup

---

## Key Decisions

1. **UUID v7 sebagai satu-satunya primary key** — konsistensi di seluruh sistem, tidak ada BIGSERIAL atau UUID v4

2. **Vernon hanya untuk domain dengan ≥3 JOIN dan read:write ≥10:1** — over-engineering untuk domain sederhana harus dihindari

3. **Sync SELALU async** — write performance tidak boleh terpengaruh oleh jumlah dependent domain

4. **Stale > Down** — aplikasi harus toleransi eventual consistency di read path

5. **Dry-run wajib sebelum import** — mencegah data corrupt masuk ke production

6. **SyncRegistry sebagai explicit mapping** — tidak ada implicit dependency; setiap relasi Vernon harus terdaftar

---

## Constraints & Implications

### Constraints

- Kolom `_rels` dan `_data` wajib ada di semua tabel Vernon, tidak boleh nullable
- Setiap update ke parent entity yang digunakan di Vernon wajib publish domain event
- `SyncRegistry` wajib di-update setiap kali relasi baru ditambahkan ke domain Vernon
- `uuid_generate_v7()` wajib di-setup di semua environment (dev, staging, prod) via migration pertama
- Import template tidak boleh memerlukan akses DB sistem lama — harus cukup dengan Excel/CSV

### Implications

- Perubahan schema pada tabel parent Vernon dapat memicu cascade sync ke ratusan ribu row — selalu evaluasi dampak SyncEngine sebelum ALTER TABLE
- Vernon `_data` bersifat eventually consistent — UI harus menampilkan indikasi jika `_sync_status = 'pending'` untuk data kritikal
- Bulk import (>1000 row) akan memicu burst event ke SyncEngine — pastikan rate limiting aktif
- Vernon read repository tidak boleh melakukan JOIN — semua data harus dari `_data` JSONB
- Saat menambah field baru ke `_data`, tidak perlu ALTER TABLE — cukup update SyncEngine dan jalankan backfill migration
