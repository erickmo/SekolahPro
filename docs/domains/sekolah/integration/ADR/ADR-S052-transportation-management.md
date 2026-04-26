# ADR-S052: Transportation Management (Manajemen Transportasi / Antar Jemput)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Banyak sekolah di Indonesia menyediakan layanan antar jemput siswa, terutama sekolah swasta dan pesantren. Manajemen transportasi diperlukan untuk:

1. **Route management**: Rute antar jemput dengan titik-titik penjemputan/pengantaran.
2. **Master kendaraan & sopir**: Data kendaraan sekolah, sopir, dan kelayakan operasional.
3. **Assignment siswa ke rute**: Siswa mana naik kendaraan mana, di titik mana.
4. **Tracking pick-up/drop-off**: Pencatatan waktu aktual penjemputan dan pengantaran untuk keamanan dan akuntabilitas.
5. **Biaya transportasi bulanan**: Tagihan antar jemput per siswa per bulan — integrasi dengan keuangan (S009).
6. **Keamanan siswa**: Orang tua perlu tahu kapan anak dijemput dan diantar sampai sekolah/rumah.
7. **Pesantren**: Pesantren yang lokasinya di pedesaan sering menyediakan bus antar jemput dari kota terdekat.

### Mengapa Vernon Pattern?

- Read-heavy: dashboard tracking, laporan rute, daftar penumpang per kendaraan.
- Relasi ke student, academic_year — has_many passengers per route.
- Business logic moderate: assignment, scheduling, fee calculation.
- Write terjadi periodik (awal semester untuk assignment, harian untuk tracking).
- Eventual consistency acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk 4 tabel: `transport_vehicles` (kendaraan), `transport_routes` (rute), `transport_passengers` (assignment siswa), dan `transport_logs` (tracking pick-up/drop-off).

### Table Schema

```sql
-- Master kendaraan sekolah
CREATE TABLE transport_vehicles (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas kendaraan
    plate_number    VARCHAR(15) NOT NULL,
    vehicle_type    VARCHAR(20) NOT NULL,
    brand           VARCHAR(50),
    model           VARCHAR(50),
    year            INT,
    color           VARCHAR(30),
    capacity        INT NOT NULL,

    -- Kelayakan
    stnk_expiry     DATE,
    kir_expiry       DATE,
    insurance_expiry DATE,
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Sopir utama
    driver_name     VARCHAR(100) NOT NULL,
    driver_phone    VARCHAR(20) NOT NULL,
    driver_license_no VARCHAR(30),
    driver_license_expiry DATE,

    -- Photo
    photo_url       TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_plate_number UNIQUE (tenant_id, company_id, plate_number),
    CONSTRAINT chk_vehicle_type CHECK (vehicle_type IN ('bus', 'minibus', 'van', 'pickup', 'sedan', 'other')),
    CONSTRAINT chk_capacity CHECK (capacity > 0 AND capacity <= 60)
);

CREATE INDEX idx_vehicle_tenant ON transport_vehicles (tenant_id, company_id);
CREATE INDEX idx_vehicle_active ON transport_vehicles (is_active) WHERE is_active = true;

-- Rute antar jemput
CREATE TABLE transport_routes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    vehicle_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Identitas rute
    route_name      VARCHAR(100) NOT NULL,
    route_code      VARCHAR(20) NOT NULL,
    route_type      VARCHAR(10) NOT NULL,
    description     TEXT,

    -- Waypoints (ordered stops)
    waypoints       JSONB NOT NULL DEFAULT '[]',

    -- Jadwal
    departure_time  TIME NOT NULL,
    estimated_arrival TIME,

    -- Kapasitas
    passenger_count INT NOT NULL DEFAULT 0,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Biaya
    monthly_fee     BIGINT NOT NULL DEFAULT 0,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_route_code UNIQUE (tenant_id, company_id, academic_year_id, route_code),
    CONSTRAINT chk_route_type CHECK (route_type IN ('pickup', 'dropoff', 'both')),
    CONSTRAINT chk_monthly_fee CHECK (monthly_fee >= 0)
);

CREATE INDEX idx_route_tenant ON transport_routes (tenant_id, company_id);
CREATE INDEX idx_route_vehicle ON transport_routes (vehicle_id);
CREATE INDEX idx_route_year ON transport_routes (academic_year_id);
CREATE INDEX idx_route_rels ON transport_routes USING GIN (_rels);
CREATE INDEX idx_route_data ON transport_routes USING GIN (_data);

-- Assignment siswa ke rute
CREATE TABLE transport_passengers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    route_id        UUID NOT NULL,
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Detail
    pickup_point    VARCHAR(200) NOT NULL,
    pickup_latitude NUMERIC(10,7),
    pickup_longitude NUMERIC(10,7),
    pickup_order    INT NOT NULL,

    -- Guardian contact
    guardian_phone  VARCHAR(20) NOT NULL,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,
    start_date      DATE NOT NULL,
    end_date        DATE,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_passenger_route UNIQUE (route_id, student_id),
    CONSTRAINT chk_pickup_order CHECK (pickup_order > 0)
);

CREATE INDEX idx_passenger_route ON transport_passengers (route_id);
CREATE INDEX idx_passenger_student ON transport_passengers (student_id);
CREATE INDEX idx_passenger_year ON transport_passengers (academic_year_id);
CREATE INDEX idx_passenger_rels ON transport_passengers USING GIN (_rels);
CREATE INDEX idx_passenger_data ON transport_passengers USING GIN (_data);

-- Log harian pick-up / drop-off
CREATE TABLE transport_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    route_id        UUID NOT NULL,
    passenger_id    UUID NOT NULL,
    student_id      UUID NOT NULL,

    -- Detail log
    log_date        DATE NOT NULL,
    log_type        VARCHAR(10) NOT NULL,
    status          VARCHAR(20) NOT NULL,

    -- Waktu aktual
    scheduled_time  TIME NOT NULL,
    actual_time     TIME,

    -- Lokasi (opsional — jika GPS tracking)
    latitude        NUMERIC(10,7),
    longitude       NUMERIC(10,7),

    -- Note
    note            TEXT,
    recorded_by     UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_log_daily UNIQUE (passenger_id, log_date, log_type),
    CONSTRAINT chk_log_type CHECK (log_type IN ('pickup', 'dropoff')),
    CONSTRAINT chk_log_status CHECK (status IN ('picked_up', 'dropped_off', 'absent', 'cancelled', 'parent_pickup'))
);

CREATE INDEX idx_log_route ON transport_logs (route_id);
CREATE INDEX idx_log_student ON transport_logs (student_id);
CREATE INDEX idx_log_date ON transport_logs (log_date);
CREATE INDEX idx_log_type_date ON transport_logs (log_type, log_date);
CREATE INDEX idx_log_rels ON transport_logs USING GIN (_rels);
CREATE INDEX idx_log_data ON transport_logs USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `waypoints` | JSONB array di route | Ordered list of stops — `[{name, lat, lng, order}]`. Flexible, bisa berubah per semester |
| `pickup_order` | INT di passenger | Urutan penjemputan dalam rute — penting untuk efisiensi sopir |
| `stnk_expiry` / `kir_expiry` | DATE | Monitoring kelayakan kendaraan — bisa alert jika akan expired |
| `monthly_fee` | BIGINT di route | Biaya antar jemput per bulan — bisa beda per rute (tergantung jarak) |
| `passenger_count` | Denormalisasi di route | Counter untuk cek kapasitas tanpa COUNT query |
| `driver_name` / `driver_phone` | VARCHAR di vehicle | Sopir embedded di kendaraan (bukan tabel terpisah) — 1 kendaraan = 1 sopir utama cukup untuk MVP |
| `log_type` | pickup / dropoff | Setiap perjalanan punya 2 log: pickup dari rumah, dropoff ke sekolah (atau sebaliknya) |
| `status` 5 values | Mencakup skenario: hadir, absent, dibatalkan, dijemput orang tua sendiri |
| `guardian_phone` | VARCHAR di passenger | Kontak darurat per penumpang — sopir perlu akses cepat |

### Vernon Relationships

**transport_routes:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `vehicle` | belongs_to | **Ya** | Info kendaraan dan sopir selalu ditampilkan |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

**transport_passengers:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `route` | belongs_to | **Ya** | Info rute dan kendaraan |
| `student` | belongs_to | **Ya** | Identitas penumpang |

**transport_logs:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `route` | belongs_to | Tidak | Bisa di-load on demand |
| `passenger` | belongs_to | **Ya** | Info penumpang dan titik jemput |
| `student` | belongs_to | **Ya** | Nama siswa untuk display |

### _rels / _data Structure

```json
// transport_routes
{
  "_rels": {
    "vehicle_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "vehicle": {
      "id": "018f...",
      "plate_number": "B 1234 XYZ",
      "vehicle_type": "minibus",
      "capacity": 15,
      "driver_name": "Pak Slamet",
      "driver_phone": "081234567890"
    },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}

// transport_passengers
{
  "_rels": {
    "route_id": "018f...",
    "student_id": "018f..."
  },
  "_data": {
    "route": { "id": "018f...", "route_name": "Rute Utara", "route_code": "RU-01", "departure_time": "06:30" },
    "student": { "id": "018f...", "full_name": "Ahmad Rizky", "nis": "12345", "class_name": "VII-A" }
  }
}

// transport_logs
{
  "_rels": {
    "route_id": "018f...",
    "passenger_id": "018f...",
    "student_id": "018f..."
  },
  "_data": {
    "passenger": { "id": "018f...", "pickup_point": "Perumahan Griya Asri Blok C", "pickup_order": 3 },
    "student": { "id": "018f...", "full_name": "Ahmad Rizky", "nis": "12345" }
  }
}
```

### API Endpoints

```
# Vehicles (Master)
GET    /api/v1/transport-vehicles                      — List kendaraan
POST   /api/v1/transport-vehicles                      — Tambah kendaraan
PUT    /api/v1/transport-vehicles/{id}                 — Update kendaraan
GET    /api/v1/transport-vehicles/expiring              — Kendaraan dengan dokumen akan expired

# Routes
GET    /api/v1/transport-routes                        — List rute
POST   /api/v1/transport-routes                        — Buat rute
PUT    /api/v1/transport-routes/{id}                   — Update rute
GET    /api/v1/transport-routes/{id}/passengers         — List penumpang per rute

# Passengers (Assignment)
POST   /api/v1/transport-passengers                    — Assign siswa ke rute
PUT    /api/v1/transport-passengers/{id}               — Update assignment
DELETE /api/v1/transport-passengers/{id}               — Remove siswa dari rute
GET    /api/v1/students/{id}/transport                  — Info transportasi siswa

# Daily Logs (Tracking)
POST   /api/v1/transport-logs                          — Catat pick-up/drop-off
POST   /api/v1/transport-logs/bulk                     — Catat bulk per rute
GET    /api/v1/transport-routes/{id}/logs               — Log harian per rute
GET    /api/v1/students/{id}/transport-logs             — Log per siswa

# Reports
GET    /api/v1/transport/summary                       — Ringkasan per rute (occupancy, punctuality)
GET    /api/v1/transport/attendance                     — Laporan kehadiran transportasi
```

## Consequences

### Positive

- **Safety tracking**: Orang tua bisa dipastikan anak sudah dijemput/diantar melalui log harian.
- **Capacity management**: Denormalisasi passenger_count memudahkan monitoring kapasitas real-time.
- **Document monitoring**: Expiry tracking untuk STNK, KIR, SIM — mencegah operasi kendaraan ilegal.
- **Fee integration**: Monthly fee per rute bisa di-generate sebagai invoice via S009 (student finance).
- **GPS-ready**: Field latitude/longitude di log memungkinkan integrasi GPS tracking di masa depan.
- **Flexible waypoints**: JSONB waypoints mengakomodasi perubahan titik jemput tanpa migration.

### Negative / Trade-offs

- **Driver as embedded field**: Sopir tidak punya tabel terpisah — jika perlu rotasi sopir atau track riwayat sopir, perlu refactor.
- **No real-time tracking**: Log adalah pencatatan manual/semi-manual — bukan live GPS tracking. Real-time tracking perlu infrastruktur IoT terpisah.
- **Monthly fee sederhana**: Satu tarif per rute — belum mendukung tarif per zona/jarak. Enhancement bisa ditambahkan.
- **Waypoints as JSONB**: Tidak bisa di-query dengan SQL biasa — perlu JSONB operators untuk filter waypoints.

## Alternatives Considered

### 1. Driver sebagai tabel terpisah
- Ditolak untuk MVP: relasi 1 kendaraan = 1 sopir cukup untuk sekolah skala kecil-menengah. Tabel terpisah bisa ditambah saat perlu multi-sopir per kendaraan.

### 2. Real-time GPS tracking via WebSocket
- Ditolak: butuh infrastruktur IoT (GPS tracker di kendaraan) yang di luar scope SekolahPro. Manual log via app sopir lebih realistis untuk MVP.

### 3. Waypoints sebagai tabel terpisah (route_stops)
- Ditolak: waypoints jarang di-query individual — JSONB array cukup. Tabel terpisah over-engineering untuk data yang selalu di-read as a whole.

### 4. Integrasi langsung dengan Google Maps API
- Ditolak untuk MVP: Google Maps API mahal per request. Cukup simpan koordinat, rendering map di frontend saja.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `VehicleDescriptor.TableName()` | — | `"transport_vehicles"` |
| U02 | `RouteDescriptor.TableName()` | — | `"transport_routes"` |
| U03 | `PassengerDescriptor.TableName()` | — | `"transport_passengers"` |
| U04 | `LogDescriptor.TableName()` | — | `"transport_logs"` |
| U05 | Validate rejects invalid `vehicle_type` | `"helicopter"` | Error |
| U06 | Validate rejects capacity > 60 | `capacity = 100` | Error |
| U07 | Validate rejects invalid `route_type` | `"roundtrip"` | Error |
| U08 | Validate rejects invalid `log_status` | `"delayed"` | Error |
| U09 | Validate accepts valid vehicle | All fields valid | No error |
| U10 | Validate rejects pickup_order <= 0 | `pickup_order = 0` | Error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create vehicle | POST with plate, driver, capacity | 201 |
| I02 | Unique plate number per company | Create 2 with same plate | 409/422 |
| I03 | Create route | POST with vehicle + waypoints | 201 |
| I04 | Assign student to route | POST passenger | 201, route passenger_count updated |
| I05 | Capacity exceeded | Assign beyond vehicle capacity | 422, capacity full |
| I06 | Unique student per route | Assign same student twice | 409/422 |
| I07 | Record pickup log | POST log type=pickup | 201 |
| I08 | Record dropoff log | POST log type=dropoff | 201 |
| I09 | Unique log per day per type | Log same student same day same type | 409/422 |
| I10 | Bulk log per route | POST /bulk for all passengers | 201, one log per active passenger |
| I11 | Student transport info | GET /students/{id}/transport | 200, route + vehicle + schedule |
| I12 | Expiring documents | GET /vehicles/expiring | 200, vehicles with docs expiring in 30 days |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | VehicleUpdated syncs to routes | Update plate number | `_data.vehicle.plate_number` updated |
| I14 | StudentUpdated syncs to passengers | Update student name | `_data.student.full_name` updated |
| I15 | RouteUpdated syncs to passengers | Update route name | `_data.route.route_name` updated |
| I16 | PassengerUpdated syncs to logs | Update pickup point | `_data.passenger.pickup_point` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Cannot access other tenant's routes | GET with wrong tenant | 404 |
| I18 | Cannot assign student cross-tenant | POST passenger cross-tenant | Error |
