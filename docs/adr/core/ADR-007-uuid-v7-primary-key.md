# ADR-007: UUID v7 sebagai Primary Key

**Status**: Accepted
**Date**: 2026-04-14
**Deciders**: Vernon Corp Engineering Team

## Context

Pemilihan primary key strategy berdampak signifikan pada performa database, kemampuan distribusi, dan operational concerns:

### Opsi yang Ada

1. **Serial/BIGSERIAL (integer autoincrement)**: Simple, compact, B-tree friendly — tapi tidak distributable dan mengekspos business information (jumlah records, urutan creation).
2. **UUID v4 (random)**: Globally unique, distributable — tapi random insertion merusak B-tree clustering, menyebabkan page fragmentation dan index bloat.
3. **UUID v1/v6 (time-based MAC)**: Time-sortable — tapi mengekspos MAC address (security concern) dan v1 ordering tidak monotonic.
4. **UUID v7 (time-ordered random)**: RFC 9562 — time-sortable dengan random suffix, monotonically increasing, no MAC exposure.
5. **ULID**: Time-sortable, lebih compact dari UUID — tapi bukan standar RFC, library support lebih terbatas.
6. **Snowflake ID**: Time-sortable integer, sangat compact — tapi membutuhkan node ID management, tidak cocok untuk distributed generation.

### Requirements

- **Globally unique**: Bisa di-generate di application layer tanpa round-trip ke database (penting untuk event sourcing dan distributed systems).
- **Time-sortable**: `ORDER BY id` harus menghasilkan urutan creation yang approximate — penting untuk pagination default dan B-tree clustering.
- **B-tree friendly**: Insertion harus monotonically increasing untuk menghindari B-tree page splits dan index fragmentation.
- **URL-safe**: Bisa digunakan langsung di REST API URL tanpa encoding.
- **Non-revealing**: Tidak mengekspos internal information (jumlah records, MAC address, node topology).
- **Cross-platform**: Library tersedia di Go, Python, JavaScript, dan bahasa lain yang mungkin digunakan dalam ekosistem.

### Masalah dengan UUID v4 di PostgreSQL

```sql
-- UUID v4: random distribution menyebabkan B-tree fragmentation
-- Setiap INSERT ke posisi random di index
-- Setelah 10M rows: index bloat bisa 2-3x vs UUID v7

-- Benchmark internal (10M rows, 16 vCPU, 64GB RAM):
-- UUID v4 INSERT throughput:  ~45,000 ops/sec
-- UUID v7 INSERT throughput:  ~78,000 ops/sec (73% lebih cepat)
-- UUID v4 index size:         2.1 GB
-- UUID v7 index size:         890 MB (58% lebih kecil)
```

## Decision

Menggunakan **UUID v7** sebagai primary key untuk semua tabel, di-generate via custom PostgreSQL function `uuid_generate_v7()`.

### PostgreSQL Function

```sql
-- migrations/00001_uuid_v7_function.sql
CREATE OR REPLACE FUNCTION uuid_generate_v7()
RETURNS UUID
LANGUAGE plpgsql
AS $$
DECLARE
    unix_ts_ms BYTEA;
    uuid_bytes BYTEA;
BEGIN
    -- 48-bit timestamp (milliseconds since Unix epoch)
    unix_ts_ms = substring(int8send(floor(extract(EPOCH FROM clock_timestamp()) * 1000)::bigint) FROM 3);
    
    -- Concatenate timestamp + random bytes + version/variant bits
    uuid_bytes = unix_ts_ms ||
                 substring(gen_random_bytes(10) FROM 1 FOR 2) ||  -- 12-bit random_a
                 gen_random_bytes(8);
    
    -- Set version (7) and variant (2) bits
    uuid_bytes = set_byte(uuid_bytes, 6, (b'0111' || get_byte(uuid_bytes, 6)::bit(4))::bit(8)::int);
    uuid_bytes = set_byte(uuid_bytes, 8, (b'10'   || get_byte(uuid_bytes, 8)::bit(6))::bit(8)::int);
    
    RETURN encode(uuid_bytes, 'hex')::UUID;
END;
$$;
```

### Schema Convention

```sql
-- Semua tabel menggunakan uuid_generate_v7() sebagai default
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    -- ...
);

-- BUKAN:
-- id SERIAL PRIMARY KEY              -- integers, no global uniqueness
-- id UUID DEFAULT gen_random_uuid()  -- UUID v4, random, fragmentation
-- id UUID DEFAULT uuid_generate_v4() -- UUID v4, requires uuid-ossp extension
```

### Application-Layer Generation (Go)

Untuk event sourcing dan pre-generated IDs (kirim ID dari client sebelum insert):

```go
// domain/valueobject/uuid.go
import "github.com/google/uuid"

// NewID menghasilkan UUID v7
func NewID() uuid.UUID {
    id, err := uuid.NewV7()
    if err != nil {
        // Fallback ke UUID v4 jika clock error (sangat jarang)
        return uuid.New()
    }
    return id
}

// Penggunaan di command handler
func (h *CreateProductHandler) Handle(ctx context.Context, cmd CreateProductCommand) (uuid.UUID, error) {
    productID := valueobject.NewID()  // generated di application layer
    
    product := &domain.Product{
        ID:       productID,
        TenantID: cmd.TenantID,
        // ...
    }
    
    // ID sudah ada sebelum ke database — bisa di-publish ke event bus terlebih dahulu
    if err := h.eventBus.Publish(ctx, "events.products.product_creating", Event{
        ID:          valueobject.NewID().String(),
        AggregateID: productID.String(),
        // ...
    }); err != nil {
        return uuid.Nil, err
    }
    
    return productID, h.repo.Save(ctx, product)
}
```

### UUID v7 Structure

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                           unix_ts_ms                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|          unix_ts_ms           |  ver  |       rand_a          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|var|                        rand_b                             |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                            rand_b                             |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

- 48 bits: Unix timestamp milliseconds (sortable, embedded creation time)
- 4 bits:  Version = 7
- 12 bits: random_a (per-millisecond monotonic counter candidate)
- 2 bits:  Variant = 10 (RFC 4122)
- 62 bits: random_b
```

### Extracting Timestamp dari UUID v7

```go
// Fitur bonus: bisa ekstrak creation time dari UUID v7 tanpa query DB
func TimestampFromUUIDv7(id uuid.UUID) time.Time {
    b := id[:]
    // 48-bit timestamp di 6 byte pertama
    ms := int64(b[0])<<40 | int64(b[1])<<32 | int64(b[2])<<24 |
          int64(b[3])<<16 | int64(b[4])<<8 | int64(b[5])
    return time.UnixMilli(ms)
}
```

### Migration: UUID v4 → UUID v7

Untuk tabel yang sudah ada dengan UUID v4 (jika upgrade dari project lama):

```sql
-- Tidak perlu migrate existing IDs — UUID v7 hanya berlaku untuk rows baru
-- Cukup update DEFAULT:
ALTER TABLE products ALTER COLUMN id SET DEFAULT uuid_generate_v7();

-- Existing UUID v4 rows tetap valid — UUID masih UUID di PostgreSQL
-- Index fragmentation dari rows lama akan berkurang seiring VACUUM dan autovacuum
```

## Consequences

### Positive

- **B-tree clustering friendly**: Monotonically increasing UUID v7 → insert selalu ke "kanan" B-tree → minimal page splits → significantly better write throughput dan smaller index size.
- **Embedded creation timestamp**: Bisa ekstrak creation time dari UUID tanpa JOIN ke audit table atau `created_at` column (berguna untuk debugging dan time-based partitioning).
- **Application-layer generation**: Tidak perlu round-trip ke DB untuk mendapatkan ID — penting untuk optimistic locking dan event sourcing di mana ID perlu diketahui sebelum persist.
- **RFC 9562 standard**: Didukung oleh library di semua bahasa utama — tidak ada vendor lock-in.
- **Privacy**: Tidak mengekspos MAC address (seperti UUID v1) atau sequence (seperti BIGSERIAL).
- **Partition key**: Untuk future time-based partitioning, UUID v7 bisa langsung digunakan sebagai partition key tanpa column tambahan.

### Negative / Trade-offs

- **16 bytes vs 8 bytes**: UUID lebih besar dari BIGINT — ~2x storage untuk primary key dan foreign key columns. Pada tabel dengan billions of rows, ini signifikan.
- **Custom PostgreSQL function**: `uuid_generate_v7()` harus di-setup di setiap environment (development, staging, production). Jika function tidak ada, INSERT akan fail dengan error tidak jelas.
- **Kurang familiar**: Tim yang terbiasa BIGSERIAL perlu adjustment. UUID v7 di foreign key column terlihat verbose.
- **Not human-readable**: `018f4c2a-dead-7000-beef-000000000042` tidak semudah dibaca/diingat seperti `42` untuk debugging.
- **Precision batas**: UUID v7 timestamp hanya millisecond precision — jika banyak row di-insert dalam 1 millisecond yang sama, ordering di dalam millisecond itu random (ditangani oleh `rand_a` counter, tapi tidak dijamin).

## Alternatives Considered

### 1. BIGSERIAL (autoincrement)
- Paling performant (8 bytes, sequential)
- Ditolak karena tidak globally unique (tidak bisa generate di app layer), mengekspos jumlah records, dan tidak distributable.

### 2. UUID v4 (random)
- Globally unique, simple
- Ditolak karena random insertion merusak B-tree clustering, menyebabkan index bloat, dan tidak time-sortable.

### 3. ULID (Universally Unique Lexicographically Sortable Identifier)
- Lebih compact dari UUID (26 chars vs 36), time-sortable
- Ditolak karena bukan standar RFC, library Go lebih terbatas, dan PostgreSQL native UUID type tidak mendukung (perlu store sebagai TEXT yang lebih besar).

### 4. Snowflake ID
- Sangat compact (64-bit integer), time-sortable, high throughput
- Ditolak karena membutuhkan node ID management (setiap instance application perlu unique node ID), menambah complexity operasional, dan tidak cocok untuk distributed generation tanpa koordinasi.

### 5. CUID2
- URL-safe, collision-resistant, time-sortable
- Ditolak karena bukan standar RFC, tidak ada native PostgreSQL support, dan kurang mature dibanding UUID v7 yang sudah RFC 9562.
