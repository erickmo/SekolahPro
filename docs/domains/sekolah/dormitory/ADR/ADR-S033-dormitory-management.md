# ADR-S033: Dormitory Management / Manajemen Asrama (Kamar & Penghuni)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Manajemen asrama adalah **kebutuhan inti** bagi pondok pesantren dan boarding school di Indonesia. Berdasarkan ADR-009 (dual-mode institution type), ketika `school_type = "islamic"`, modul asrama menjadi fitur utama — bukan opsional.

Data asrama diperlukan untuk:

1. **Penempatan santri**: Mapping santri ke gedung → lantai → kamar → tempat tidur.
2. **Kapasitas**: Monitoring kapasitas kamar untuk mencegah over-occupancy.
3. **Musyrif/Musyrifah**: Assignment pembina asrama (wali asrama) ke gedung/lantai.
4. **Rotasi tahunan**: Penempatan ulang santri setiap tahun ajaran baru (ADR-010).
5. **Perpindahan**: Transfer kamar mid-semester karena alasan tertentu.
6. **Pemisahan gender**: Asrama putra dan putri harus terpisah secara absolut.
7. **Laporan**: Daftar penghuni per kamar, statistik occupancy per gedung.

Konteks pesantren Indonesia:
- Satu pondok bisa memiliki **beberapa gedung** asrama (putra/putri terpisah).
- Kapasitas kamar bervariasi: 4-20 santri per kamar tergantung tipe pesantren.
- Musyrif (pembina putra) / Musyrifah (pembina putri) bertanggung jawab atas satu gedung atau satu lantai.
- Penempatan biasanya dikelompokkan per tingkat (kelas 7 di lantai 1, kelas 8 di lantai 2, dst).

### Mengapa Vernon Pattern?

- Master data (gedung, kamar) = read-heavy, jarang berubah.
- Assignment santri = has_many dari student, berubah per tahun ajaran.
- Relasi ke student, academic_year, teacher (musyrif).
- Business logic moderate: capacity check, gender validation, transfer workflow.
- Eventual consistency acceptable — penempatan bukan real-time operation.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `dormitory_buildings` (master gedung), `dormitory_rooms` (master kamar), dan `dormitory_assignments` (penempatan santri).

### Table Schema

```sql
-- Master gedung asrama
CREATE TABLE dormitory_buildings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    description     TEXT,

    -- Konfigurasi
    gender          VARCHAR(10) NOT NULL,
    total_floors    INT NOT NULL DEFAULT 1,
    total_capacity  INT NOT NULL DEFAULT 0,

    -- Penanggung jawab
    supervisor_id   UUID,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_dorm_building_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_dorm_gender CHECK (gender IN ('male', 'female')),
    CONSTRAINT chk_dorm_floors CHECK (total_floors >= 1)
);

-- Indexes
CREATE INDEX idx_dorm_building_tenant_company ON dormitory_buildings (tenant_id, company_id);
CREATE INDEX idx_dorm_building_rels ON dormitory_buildings USING GIN (_rels);
CREATE INDEX idx_dorm_building_data ON dormitory_buildings USING GIN (_data);

-- Master kamar asrama
CREATE TABLE dormitory_rooms (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    building_id     UUID NOT NULL,

    -- Identitas
    room_number     VARCHAR(20) NOT NULL,
    name            VARCHAR(100),
    floor           INT NOT NULL DEFAULT 1,

    -- Kapasitas
    capacity        INT NOT NULL,
    current_occupancy INT NOT NULL DEFAULT 0,

    -- Konfigurasi
    room_type       VARCHAR(20) NOT NULL DEFAULT 'regular',
    has_bathroom    BOOLEAN NOT NULL DEFAULT false,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_dorm_room_number UNIQUE (building_id, room_number),
    CONSTRAINT chk_dorm_room_type CHECK (room_type IN ('regular', 'vip', 'isolation', 'musyrif')),
    CONSTRAINT chk_dorm_capacity CHECK (capacity >= 1),
    CONSTRAINT chk_dorm_occupancy CHECK (current_occupancy >= 0 AND current_occupancy <= capacity),
    CONSTRAINT chk_dorm_floor CHECK (floor >= 1)
);

-- Indexes
CREATE INDEX idx_dorm_room_tenant_company ON dormitory_rooms (tenant_id, company_id);
CREATE INDEX idx_dorm_room_building ON dormitory_rooms (building_id);
CREATE INDEX idx_dorm_room_floor ON dormitory_rooms (building_id, floor);
CREATE INDEX idx_dorm_room_rels ON dormitory_rooms USING GIN (_rels);
CREATE INDEX idx_dorm_room_data ON dormitory_rooms USING GIN (_data);

-- Penempatan santri ke kamar per tahun ajaran
CREATE TABLE dormitory_assignments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    room_id         UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Penempatan
    bed_number      INT,
    assigned_date   DATE NOT NULL,
    end_date        DATE,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',

    -- Transfer tracking
    previous_room_id UUID,
    transfer_reason TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_dorm_assignment_active UNIQUE (student_id, academic_year_id, status),
    CONSTRAINT chk_dorm_assign_status CHECK (status IN ('active', 'transferred', 'ended')),
    CONSTRAINT chk_dorm_bed CHECK (bed_number IS NULL OR bed_number >= 1)
);

-- Indexes
CREATE INDEX idx_dorm_assign_tenant_company ON dormitory_assignments (tenant_id, company_id);
CREATE INDEX idx_dorm_assign_student ON dormitory_assignments (student_id);
CREATE INDEX idx_dorm_assign_room ON dormitory_assignments (room_id);
CREATE INDEX idx_dorm_assign_year ON dormitory_assignments (academic_year_id);
CREATE INDEX idx_dorm_assign_active ON dormitory_assignments (room_id, status) WHERE status = 'active';
CREATE INDEX idx_dorm_assign_rels ON dormitory_assignments USING GIN (_rels);
CREATE INDEX idx_dorm_assign_data ON dormitory_assignments USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `gender` | VARCHAR(10), CHECK male/female | Pemisahan absolut asrama putra/putri — validasi di application layer bahwa santri laki-laki hanya bisa di gedung `male` |
| `supervisor_id` | UUID, nullable | Musyrif/Musyrifah — reference ke teachers (ADR-012). Nullable karena bisa belum di-assign |
| `total_capacity` | INT, denormalisasi | Computed sum dari room capacities — denormalisasi untuk quick dashboard |
| `room_type` | VARCHAR(20), CHECK | 4 tipe: regular (santri), vip (santri khusus), isolation (sakit), musyrif (pembina) |
| `current_occupancy` | INT, denormalisasi | Counter aktif — diupdate setiap assignment/transfer. Trade-off: consistency vs query speed |
| `bed_number` | INT, nullable | Nullable karena tidak semua pesantren mengelola per-bed. Beberapa hanya per-kamar |
| `previous_room_id` | UUID, nullable | Tracking asal kamar saat transfer — audit trail perpindahan |
| `status` | active/transferred/ended | Lifecycle: active (saat ini), transferred (pindah kamar), ended (akhir tahun ajaran) |

### Vernon Relationships

**dormitory_rooms:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `building` | belongs_to | **Ya** | Selalu perlu nama gedung saat menampilkan kamar |

**dormitory_assignments:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Nama santri selalu ditampilkan |
| `room` | belongs_to | **Ya** | Nomor kamar + gedung |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

### _rels / _data Structure

```json
// dormitory_rooms
{
  "_rels": {
    "building_id": "018f..."
  },
  "_data": {
    "building": { "id": "018f...", "name": "Gedung Al-Farabi", "code": "GAF", "gender": "male" }
  }
}

// dormitory_assignments
{
  "_rels": {
    "student_id": "018f...",
    "room_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad Fauzi", "nis": "12345" },
    "room": {
      "id": "018f...",
      "room_number": "201",
      "floor": 2,
      "building": { "id": "018f...", "name": "Gedung Al-Farabi", "code": "GAF" }
    },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

### API Endpoints

```
# Buildings (Master)
GET    /api/v1/dormitory-buildings                    — List gedung asrama
POST   /api/v1/dormitory-buildings                    — Buat gedung
PUT    /api/v1/dormitory-buildings/{id}               — Update gedung
GET    /api/v1/dormitory-buildings/{id}/occupancy      — Statistik occupancy gedung

# Rooms (Master)
GET    /api/v1/dormitory-rooms                        — List kamar (filter by building, floor)
POST   /api/v1/dormitory-rooms                        — Buat kamar
PUT    /api/v1/dormitory-rooms/{id}                   — Update kamar
GET    /api/v1/dormitory-rooms/{id}/residents          — Daftar penghuni kamar

# Assignments
GET    /api/v1/students/{id}/dormitory-assignments     — Riwayat penempatan santri
POST   /api/v1/dormitory-assignments                   — Assign santri ke kamar
POST   /api/v1/dormitory-assignments/bulk              — Bulk assign (awal tahun ajaran)
POST   /api/v1/dormitory-assignments/{id}/transfer     — Transfer santri ke kamar lain
POST   /api/v1/dormitory-assignments/{id}/end          — Akhiri penempatan
```

## Consequences

### Positive

- **Pesantren-first**: Dirancang khusus untuk kebutuhan asrama pesantren Indonesia.
- **Gender safety**: Validasi pemisahan gender di level database (building) dan application.
- **Capacity control**: `current_occupancy` mencegah over-capacity secara real-time.
- **Transfer tracking**: Riwayat perpindahan kamar terdokumentasi dengan alasan.
- **Yearly rotation**: Assignment per academic_year mendukung rotasi tahunan.
- **Flexible bed management**: `bed_number` opsional — mendukung pesantren yang mengelola per-bed maupun per-kamar saja.

### Negative / Trade-offs

- **Occupancy denormalisasi**: `current_occupancy` dan `total_capacity` perlu dijaga konsisten via application logic — risiko drift jika ada bug.
- **Unique constraint limitation**: `uq_dorm_assignment_active` hanya efektif jika status di-manage dengan benar — perlu application-level enforcement.
- **Supervisor single**: Satu gedung satu supervisor — belum mendukung multiple supervisor (shift siang/malam). Enhancement di masa depan.
- **No bed mapping visual**: Belum ada support untuk layout visual kamar/bed. MVP hanya data-driven.

## Alternatives Considered

### 1. Single table dormitory (tanpa building hierarchy)
- Ditolak: pesantren besar punya multiple gedung — hierarchy building → room essential.

### 2. Room assignment sebagai JSONB array di student
- Ditolak: tidak bisa query "siapa saja di kamar 201" secara efisien.

### 3. Pisah transfer ke tabel terpisah
- Ditolak: transfer adalah assignment baru dengan `previous_room_id` — satu tabel cukup untuk MVP.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `BuildingDescriptor.TableName()` | — | `"dormitory_buildings"` |
| U02 | `RoomDescriptor.TableName()` | — | `"dormitory_rooms"` |
| U03 | `AssignmentDescriptor.TableName()` | — | `"dormitory_assignments"` |
| U04 | Validate rejects invalid `gender` | `"mixed"` | Error: must be male/female |
| U05 | Validate rejects invalid `room_type` | `"suite"` | Error: invalid room_type |
| U06 | Validate rejects `capacity < 1` | `capacity = 0` | Error: capacity must be >= 1 |
| U07 | Validate rejects invalid `status` | `"pending"` | Error: must be active/transferred/ended |
| U08 | Validate accepts valid assignment | All fields valid | No error |

### Integration Tests — Buildings & Rooms

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create building | POST with name + code + gender | 201, created |
| I02 | Unique building code | Create 2 buildings same code | 409/422 |
| I03 | Create room | POST with building_id + room_number + capacity | 201, created |
| I04 | Unique room number per building | Create 2 rooms same number in same building | 409/422 |
| I05 | Gender CHECK enforced | INSERT building with `gender = 'mixed'` | DB error |
| I06 | Room type CHECK enforced | INSERT room with `room_type = 'suite'` | DB error |

### Integration Tests — Assignments

| # | Test Case | Action | Expected |
|---|---|---|---|
| I07 | Assign student to room | POST assignment | 201, `current_occupancy` incremented |
| I08 | Capacity check | Assign when room full | 422, room at capacity |
| I09 | Gender mismatch | Assign female student to male building | 422, gender mismatch |
| I10 | Transfer student | POST /transfer with new room_id + reason | 200, old assignment → transferred, new assignment created |
| I11 | End assignment | POST /end | 200, status → ended, `current_occupancy` decremented |
| I12 | Bulk assign | POST /bulk for 20 students | 201, all assigned, occupancy updated |
| I13 | Get room residents | GET /dormitory-rooms/{id}/residents | 200, active assignments only |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | StudentUpdated syncs to assignments | Update student name | `_data.student.full_name` updated |
| I15 | RoomUpdated syncs to assignments | Update room number | `_data.room.room_number` updated |
| I16 | BuildingUpdated syncs to rooms | Update building name | `_data.building.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Cannot access other tenant's buildings | GET with wrong tenant | 404 |
| I18 | Cannot assign student cross-tenant | POST assignment with cross-tenant room | Error |
