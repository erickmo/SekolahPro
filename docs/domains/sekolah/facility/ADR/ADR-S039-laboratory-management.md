# ADR-S039: Laboratory Management (Manajemen Laboratorium)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Laboratorium adalah fasilitas penting di sekolah Indonesia, terutama untuk mata pelajaran IPA dan Informatika. Manajemen laboratorium yang baik diperlukan untuk:

1. **Inventaris alat**: Setiap lab memiliki alat-alat yang perlu ditracking kondisi, jumlah, dan lokasinya.
2. **Penjadwalan penggunaan**: Lab dipakai bergantian oleh beberapa kelas — harus terintegrasi dengan jadwal pelajaran (S021).
3. **Log penggunaan**: Setiap sesi penggunaan lab harus tercatat (guru, kelas, topik, alat yang dipakai).
4. **Keselamatan**: Lab IPA (terutama kimia) membutuhkan checklist keselamatan (K3).
5. **Pemeliharaan alat**: Tracking kondisi alat dari baru hingga rusak/hilang.
6. **Pelaporan Dapodik**: Inventaris lab dilaporkan ke Kemendikbud.
7. **Pesantren**: Beberapa pesantren modern memiliki lab komputer dan multimedia (ADR-009).

Tipe laboratorium di sekolah Indonesia:
- **Lab IPA**: Fisika, Kimia, Biologi — bisa terpisah (SMA) atau gabungan (SMP).
- **Lab Komputer**: Untuk mata pelajaran Informatika dan ANBK (asesmen nasional).
- **Lab Bahasa**: Audio-visual untuk pelajaran bahasa.
- **Lab Multimedia**: Editing video, desain grafis, broadcasting.

Volume data:
- Rata-rata sekolah memiliki 2-6 laboratorium.
- Setiap lab memiliki 20-200 alat/perangkat.
- Log penggunaan: ~5-15 sesi per lab per minggu = ~300-900 records per semester.

### Mengapa Vernon Pattern?

- Master data lab + equipment = read-heavy, jarang berubah.
- Usage log = has_many, moderate volume per semester.
- Relasi ke teacher (ADR-012), class_room (ADR-011), schedule_entries (S021).
- Business logic moderate: condition tracking, availability check, safety compliance.
- Eventually consistent acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `laboratories` (master lab), `lab_equipment` (inventaris alat per lab), dan `lab_usage_logs` (log penggunaan per sesi).

### Table Schema

```sql
-- Master laboratorium
CREATE TABLE laboratories (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20) NOT NULL,
    lab_type        VARCHAR(20) NOT NULL,

    -- Lokasi & kapasitas
    building        VARCHAR(100),
    floor           INT,
    room_number     VARCHAR(20),
    capacity        INT NOT NULL DEFAULT 30,

    -- Pengelola
    head_technician_id UUID,

    -- Spesialisasi (untuk lab IPA)
    specialization  VARCHAR(20),

    -- Keselamatan
    has_safety_shower    BOOLEAN NOT NULL DEFAULT false,
    has_eye_wash         BOOLEAN NOT NULL DEFAULT false,
    has_fire_extinguisher BOOLEAN NOT NULL DEFAULT false,
    has_first_aid_kit    BOOLEAN NOT NULL DEFAULT false,
    has_fume_hood        BOOLEAN NOT NULL DEFAULT false,
    last_safety_inspection DATE,

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

    CONSTRAINT uq_laboratory_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_lab_type CHECK (lab_type IN (
        'lab_ipa', 'lab_komputer', 'lab_bahasa', 'lab_multimedia'
    )),
    CONSTRAINT chk_lab_specialization CHECK (specialization IS NULL OR specialization IN (
        'fisika', 'kimia', 'biologi', 'ipa_terpadu'
    )),
    CONSTRAINT chk_lab_capacity CHECK (capacity >= 1 AND capacity <= 100)
);

-- Indexes
CREATE INDEX idx_lab_tenant_company ON laboratories (tenant_id, company_id);
CREATE INDEX idx_lab_type ON laboratories (lab_type);
CREATE INDEX idx_lab_active ON laboratories (is_active) WHERE is_active = true;
CREATE INDEX idx_lab_head_tech ON laboratories (head_technician_id) WHERE head_technician_id IS NOT NULL;
CREATE INDEX idx_lab_rels ON laboratories USING GIN (_rels);
CREATE INDEX idx_lab_data ON laboratories USING GIN (_data);

-- Inventaris alat laboratorium
CREATE TABLE lab_equipment (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    laboratory_id   UUID NOT NULL,

    -- Identitas alat
    name            VARCHAR(200) NOT NULL,
    code            VARCHAR(30),
    brand           VARCHAR(100),
    model           VARCHAR(100),
    serial_number   VARCHAR(100),
    category        VARCHAR(30) NOT NULL,

    -- Kuantitas
    quantity        INT NOT NULL DEFAULT 1,
    unit            VARCHAR(20) NOT NULL DEFAULT 'unit',

    -- Kondisi
    condition       VARCHAR(20) NOT NULL DEFAULT 'baik',

    -- Pengadaan
    purchase_date   DATE,
    purchase_price  BIGINT,
    supplier        VARCHAR(200),
    warranty_expiry DATE,

    -- Lokasi detail
    storage_location VARCHAR(100),

    -- Kalibrasi (untuk alat ukur)
    requires_calibration BOOLEAN NOT NULL DEFAULT false,
    last_calibration_date DATE,
    next_calibration_date DATE,

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

    CONSTRAINT uq_lab_equipment_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_equipment_category CHECK (category IN (
        'alat_ukur', 'alat_percobaan', 'bahan_kimia', 'glassware',
        'komputer', 'peripheral', 'audio_visual', 'furniture',
        'safety_equipment', 'consumable', 'other'
    )),
    CONSTRAINT chk_equipment_condition CHECK (condition IN (
        'baik', 'rusak_ringan', 'rusak_berat', 'hilang'
    )),
    CONSTRAINT chk_equipment_quantity CHECK (quantity >= 0),
    CONSTRAINT chk_equipment_unit CHECK (unit IN (
        'unit', 'set', 'buah', 'lembar', 'botol', 'pack', 'roll', 'liter', 'kg'
    ))
);

-- Indexes
CREATE INDEX idx_lab_equip_tenant_company ON lab_equipment (tenant_id, company_id);
CREATE INDEX idx_lab_equip_laboratory ON lab_equipment (laboratory_id);
CREATE INDEX idx_lab_equip_category ON lab_equipment (category);
CREATE INDEX idx_lab_equip_condition ON lab_equipment (condition);
CREATE INDEX idx_lab_equip_condition_bad ON lab_equipment (condition) WHERE condition IN ('rusak_ringan', 'rusak_berat', 'hilang');
CREATE INDEX idx_lab_equip_calibration ON lab_equipment (next_calibration_date) WHERE requires_calibration = true;
CREATE INDEX idx_lab_equip_rels ON lab_equipment USING GIN (_rels);
CREATE INDEX idx_lab_equip_data ON lab_equipment USING GIN (_data);

-- Log penggunaan laboratorium per sesi
CREATE TABLE lab_usage_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    laboratory_id   UUID NOT NULL,
    teacher_id      UUID NOT NULL,
    class_room_id   UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    schedule_entry_id UUID,

    -- Waktu penggunaan
    usage_date      DATE NOT NULL,
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,
    semester        VARCHAR(10) NOT NULL,

    -- Detail kegiatan
    subject_name    VARCHAR(100) NOT NULL,
    topic           VARCHAR(300) NOT NULL,
    activity_description TEXT,
    student_count   INT NOT NULL,

    -- Alat yang digunakan (JSONB array)
    equipment_used  JSONB NOT NULL DEFAULT '[]',

    -- Keselamatan
    safety_checklist JSONB NOT NULL DEFAULT '{}',
    incident_notes  TEXT,
    has_incident    BOOLEAN NOT NULL DEFAULT false,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'completed',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_usage_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_usage_time CHECK (start_time < end_time),
    CONSTRAINT chk_usage_student_count CHECK (student_count >= 0),
    CONSTRAINT chk_usage_status CHECK (status IN ('scheduled', 'in_progress', 'completed', 'cancelled'))
);

-- Indexes
CREATE INDEX idx_lab_usage_tenant_company ON lab_usage_logs (tenant_id, company_id);
CREATE INDEX idx_lab_usage_laboratory ON lab_usage_logs (laboratory_id);
CREATE INDEX idx_lab_usage_teacher ON lab_usage_logs (teacher_id);
CREATE INDEX idx_lab_usage_class ON lab_usage_logs (class_room_id);
CREATE INDEX idx_lab_usage_date ON lab_usage_logs (usage_date);
CREATE INDEX idx_lab_usage_schedule ON lab_usage_logs (schedule_entry_id) WHERE schedule_entry_id IS NOT NULL;
CREATE INDEX idx_lab_usage_year_semester ON lab_usage_logs (academic_year_id, semester);
CREATE INDEX idx_lab_usage_incident ON lab_usage_logs (has_incident) WHERE has_incident = true;
CREATE INDEX idx_lab_usage_equipment ON lab_usage_logs USING GIN (equipment_used);
CREATE INDEX idx_lab_usage_rels ON lab_usage_logs USING GIN (_rels);
CREATE INDEX idx_lab_usage_data ON lab_usage_logs USING GIN (_data);
```

### Field Design Rationale

**laboratories:**

| Field | Keputusan | Alasan |
|---|---|---|
| `lab_type` | 4 tipe | lab_ipa, lab_komputer, lab_bahasa, lab_multimedia — sesuai tipe lab sekolah Indonesia |
| `specialization` | 4 spesialisasi, nullable | Untuk lab IPA: fisika/kimia/biologi/ipa_terpadu. SMP biasanya `ipa_terpadu`, SMA terpisah. Null untuk non-IPA lab |
| `capacity` | INT, 1-100 | Kapasitas meja praktikum — default 30 (standar sekolah) |
| `head_technician_id` | UUID, nullable | Laboran/teknisi penanggung jawab — nullable karena belum tentu ada |
| `has_safety_*` | 5 BOOLEAN | Fasilitas keselamatan K3 — wajib untuk lab kimia sesuai Permendikbud |
| `last_safety_inspection` | DATE, nullable | Tanggal inspeksi terakhir — untuk compliance tracking |

**lab_equipment:**

| Field | Keputusan | Alasan |
|---|---|---|
| `category` | 11 kategori | Mencakup semua tipe alat lab: alat_ukur (termometer, multimeter), bahan_kimia, glassware, komputer, dll |
| `condition` | 4 status | `baik`, `rusak_ringan`, `rusak_berat`, `hilang` — sesuai standar BOS/Dapodik |
| `quantity` | INT, >= 0 | Jumlah per record — 0 jika semua hilang |
| `unit` | 9 satuan | unit, set, buah, botol, dll — sesuai jenis barang lab |
| `purchase_price` | BIGINT | Harga beli dalam Rupiah — untuk depreciation dan aset management |
| `requires_calibration` | BOOLEAN | Alat ukur presisi (timbangan, multimeter) butuh kalibrasi berkala |
| `serial_number` | VARCHAR, nullable | Untuk alat-alat mahal yang perlu tracking individu |

**lab_usage_logs:**

| Field | Keputusan | Alasan |
|---|---|---|
| `schedule_entry_id` | UUID, nullable | Link ke jadwal S021 — nullable karena ada penggunaan di luar jadwal reguler |
| `equipment_used` | JSONB array | Daftar alat yang dipakai per sesi `[{"equipment_id": "...", "name": "...", "qty_used": 5}]` |
| `safety_checklist` | JSONB object | Checklist keselamatan per sesi — flexible per tipe lab |
| `has_incident` | BOOLEAN | Flag cepat untuk filter sesi yang ada insiden (kecelakaan, kerusakan) |
| `subject_name` | VARCHAR, denormalisasi | Nama mapel — denormalisasi untuk kemudahan query tanpa join |

### Safety Checklist Structure (Application Logic)

```json
// Lab Kimia
{
  "apd_lengkap": true,
  "ventilasi_berfungsi": true,
  "fume_hood_aktif": true,
  "bahan_berbahaya_tersimpan": true,
  "p3k_tersedia": true,
  "jalur_evakuasi_bersih": true,
  "apar_tersedia": true,
  "siswa_briefing_keselamatan": true
}

// Lab Komputer
{
  "ac_berfungsi": true,
  "ups_aktif": true,
  "kabel_rapi": true,
  "internet_tersedia": true,
  "antivirus_update": true
}
```

Checklist di-generate oleh frontend berdasarkan `lab_type` — disimpan as-is di JSONB.

### Equipment Used Structure

```json
[
  {
    "equipment_id": "018f-abc...",
    "name": "Gelas Ukur 100ml",
    "qty_used": 15,
    "condition_after": "baik"
  },
  {
    "equipment_id": "018f-def...",
    "name": "Bunsen Burner",
    "qty_used": 8,
    "condition_after": "baik"
  }
]
```

`condition_after` memungkinkan update kondisi alat setelah digunakan — jika `condition_after != condition` saat ini, service layer akan update kondisi di `lab_equipment`.

### Vernon Relationships

**laboratories:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `head_technician` | belongs_to | **Ya** | Nama laboran ditampilkan di detail lab |

**lab_equipment:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `laboratory` | belongs_to | **Ya** | Nama dan tipe lab selalu ditampilkan bersama alat |

**lab_usage_logs:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `laboratory` | belongs_to | **Ya** | Nama dan tipe lab |
| `teacher` | belongs_to | **Ya** | Guru yang menggunakan lab |
| `class_room` | belongs_to | **Ya** | Kelas yang menggunakan lab |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

### _rels / _data Structure

**laboratories:**
```json
{
  "_rels": {
    "head_technician_id": "018f..."
  },
  "_data": {
    "head_technician": { "id": "018f...", "full_name": "Pak Darmawan", "nip": "198701012015011001" }
  }
}
```

**lab_equipment:**
```json
{
  "_rels": {
    "laboratory_id": "018f..."
  },
  "_data": {
    "laboratory": { "id": "018f...", "name": "Lab Kimia", "code": "LAB-KIM", "lab_type": "lab_ipa", "specialization": "kimia" }
  }
}
```

**lab_usage_logs:**
```json
{
  "_rels": {
    "laboratory_id": "018f...",
    "teacher_id": "018f...",
    "class_room_id": "018f...",
    "academic_year_id": "018f...",
    "schedule_entry_id": "018f..."
  },
  "_data": {
    "laboratory": { "id": "018f...", "name": "Lab Kimia", "code": "LAB-KIM", "lab_type": "lab_ipa" },
    "teacher": { "id": "018f...", "full_name": "Bu Ratna", "nip": "199001012018012001" },
    "class_room": { "id": "018f...", "name": "X-IPA-1" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

### API Endpoints

```
# Laboratories (Master)
GET    /api/v1/laboratories                             — List semua lab
POST   /api/v1/laboratories                             — Buat lab baru
GET    /api/v1/laboratories/{id}                        — Detail lab + summary equipment
PUT    /api/v1/laboratories/{id}                        — Update lab
DELETE /api/v1/laboratories/{id}                        — Soft delete lab

# Equipment (Inventaris)
GET    /api/v1/laboratories/{id}/equipment              — List alat per lab
GET    /api/v1/lab-equipment                            — List semua alat (filter: lab_id, category, condition)
POST   /api/v1/lab-equipment                            — Tambah alat baru
PUT    /api/v1/lab-equipment/{id}                       — Update alat (kondisi, lokasi)
PUT    /api/v1/lab-equipment/{id}/condition              — Update kondisi alat (dedicated endpoint)
GET    /api/v1/lab-equipment/condition-report            — Laporan kondisi alat (summary per condition)

# Usage Logs
GET    /api/v1/lab-usage-logs                           — List log (filter: lab_id, teacher_id, date range)
POST   /api/v1/lab-usage-logs                           — Catat penggunaan lab
PUT    /api/v1/lab-usage-logs/{id}                      — Update log
GET    /api/v1/laboratories/{id}/usage-logs             — Log penggunaan per lab
GET    /api/v1/teachers/{id}/lab-usage-logs             — Log penggunaan per guru

# Safety
PUT    /api/v1/laboratories/{id}/safety-inspection      — Catat inspeksi keselamatan
GET    /api/v1/laboratories/safety-report               — Laporan keselamatan semua lab

# Reports
GET    /api/v1/laboratories/statistics                  — Statistik penggunaan semua lab
GET    /api/v1/laboratories/{id}/statistics              — Statistik per lab (frekuensi, top teacher, top class)
```

## Consequences

### Positive

- **Inventaris lengkap**: Tracking per alat dengan kondisi, pembelian, dan kalibrasi.
- **Integrasi jadwal S021**: `schedule_entry_id` menghubungkan penggunaan lab dengan jadwal pelajaran.
- **Safety compliance**: Checklist keselamatan per sesi — audit trail untuk K3.
- **Condition lifecycle**: 4 status kondisi (`baik` → `rusak_ringan` → `rusak_berat` → `hilang`) sesuai standar Dapodik.
- **Flexible equipment tracking**: JSONB `equipment_used` memungkinkan pencatatan alat yang dipakai per sesi tanpa tabel junction tambahan.
- **Lab IPA specialization**: Mendukung lab fisika/kimia/biologi terpisah (SMA) atau terpadu (SMP).

### Negative / Trade-offs

- **Equipment used sebagai JSONB**: Tidak bisa di-JOIN langsung — trade-off untuk simplicity. Jika butuh analisis detail alat terpopuler, perlu parsing JSONB.
- **Safety checklist tidak terstruktur**: JSONB memungkinkan variasi checklist per tipe lab, tapi tidak enforce required fields di database level.
- **No real-time availability**: Belum ada mekanisme lock/booking lab secara real-time — dihandle oleh S041 (Room Booking) untuk booking di luar jadwal.
- **Condition update manual**: Kondisi alat diupdate manual — belum ada IoT integration.
- **Subject denormalisasi**: `subject_name` di usage log adalah string — bukan FK ke tabel subjects — trade-off untuk simplicity.

## Alternatives Considered

### 1. Equipment sebagai JSONB di laboratories (satu tabel saja)
- Ditolak: equipment butuh query per kondisi, per kategori, per lab — relational column lebih efisien dari nested JSONB.

### 2. Equipment used sebagai junction table (lab_usage_equipment)
- Ditolak untuk MVP: JSONB array cukup — junction table menambah complexity tanpa benefit signifikan untuk volume kecil (~10 alat per sesi).

### 3. Lab schedule sebagai tabel terpisah
- Ditolak: jadwal lab sudah terintegrasi di `schedule_entries` (S021) via `room_name`. Usage log cukup merecord realisasi penggunaan.

### 4. Gabung dengan S040 (Asset Management)
- Ditolak: lab equipment punya field spesifik (kalibrasi, safety, kategori lab) yang tidak ada di aset umum. Namun lab equipment bisa cross-reference ke `assets` via code.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `LaboratoryDescriptor.TableName()` | — | `"laboratories"` |
| U02 | `LabEquipmentDescriptor.TableName()` | — | `"lab_equipment"` |
| U03 | `LabUsageLogDescriptor.TableName()` | — | `"lab_usage_logs"` |
| U04 | Validate rejects invalid `lab_type` | `"lab_musik"` | Error: invalid lab_type |
| U05 | Validate rejects invalid `specialization` | `"geologi"` | Error: invalid specialization |
| U06 | Validate rejects invalid `condition` | `"hancur"` | Error: must be baik/rusak_ringan/rusak_berat/hilang |
| U07 | Validate rejects invalid `category` | `"senjata"` | Error: invalid category |
| U08 | Validate rejects `start_time >= end_time` | 14:00 >= 12:00 | Error: start must be before end |
| U09 | Validate rejects negative `quantity` | `-1` | Error: must be >= 0 |
| U10 | Validate accepts valid laboratory | All fields valid | No error |
| U11 | Validate accepts valid equipment | All fields valid | No error |
| U12 | Validate accepts valid usage log | All fields valid | No error |

### Integration Tests — Laboratories

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create laboratory | POST with valid data | 201 |
| I02 | Unique code per company | Create 2 labs same code | 409/422 |
| I03 | Lab type CHECK | INSERT with `lab_type = 'lab_musik'` | DB error |
| I04 | Specialization CHECK | INSERT lab_ipa with `specialization = 'geologi'` | DB error |
| I05 | Get lab detail with equipment summary | GET /laboratories/{id} | 200, includes equipment count by condition |

### Integration Tests — Equipment

| # | Test Case | Action | Expected |
|---|---|---|---|
| I06 | Add equipment to lab | POST with valid data | 201 |
| I07 | Unique code per company | Create 2 equipment same code | 409/422 |
| I08 | Condition CHECK | INSERT with `condition = 'hancur'` | DB error |
| I09 | Category CHECK | INSERT with `category = 'senjata'` | DB error |
| I10 | Update condition | PUT /lab-equipment/{id}/condition to `rusak_ringan` | 200 |
| I11 | Filter by condition | GET /lab-equipment?condition=rusak_berat | 200, filtered |
| I12 | Condition report | GET /lab-equipment/condition-report | 200, summary counts |

### Integration Tests — Usage Logs

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | Create usage log | POST with valid data + equipment_used | 201 |
| I14 | Semester CHECK | INSERT with `semester = 'midterm'` | DB error |
| I15 | Status CHECK | INSERT with `status = 'expired'` | DB error |
| I16 | Equipment condition auto-update | POST with `condition_after = 'rusak_ringan'` | Equipment condition updated |
| I17 | Filter by lab | GET /laboratories/{id}/usage-logs | 200, filtered per lab |
| I18 | Filter by teacher | GET /teachers/{id}/lab-usage-logs | 200, filtered per guru |
| I19 | Log with incident | POST with `has_incident = true` + incident_notes | 201 |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I20 | TeacherUpdated syncs to laboratories | Update technician name | `_data.head_technician.full_name` updated |
| I21 | LaboratoryUpdated syncs to equipment | Update lab name | `_data.laboratory.name` updated |
| I22 | LaboratoryUpdated syncs to usage logs | Update lab name | `_data.laboratory.name` updated |
| I23 | TeacherUpdated syncs to usage logs | Update teacher name | `_data.teacher.full_name` updated |
| I24 | ClassRoomUpdated syncs to usage logs | Update class name | `_data.class_room.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I25 | Cannot access other tenant's labs | GET with wrong tenant | 404 |
| I26 | Cannot add equipment to other tenant's lab | POST equipment cross-tenant | Error |
| I27 | Cannot create usage log in other tenant's lab | POST usage log cross-tenant | Error |
