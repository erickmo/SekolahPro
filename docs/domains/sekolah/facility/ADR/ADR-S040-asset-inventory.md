# ADR-S040: Asset & Inventory Management (Manajemen Aset & Inventaris Sekolah)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Manajemen aset dan inventaris adalah kewajiban setiap sekolah di Indonesia. Sekolah negeri tunduk pada **Permendagri No. 19/2016** tentang Pengelolaan Barang Milik Daerah (BMD), sedangkan sekolah swasta mengikuti aturan yayasan masing-masing. Manajemen aset diperlukan untuk:

1. **Pencatatan aset**: Setiap barang milik sekolah harus tercatat dengan kode barang, lokasi, dan kondisi.
2. **Lifecycle management**: Dari pengadaan → registrasi → penggunaan → pemeliharaan → penghapusan.
3. **Kode barang**: Sistem penomoran aset sesuai Permendagri (kode golongan + bidang + kelompok + sub kelompok + register).
4. **Depreciation**: Perhitungan penyusutan aset (straight-line method) untuk laporan keuangan.
5. **Room-based mapping**: Pemetaan aset per ruangan — meja/kursi di ruang kelas, komputer di lab, dll.
6. **Opname tahunan**: Stock-taking tahunan untuk verifikasi keberadaan dan kondisi aset.
7. **Pelaporan**: Kartu Inventaris Barang (KIB), Laporan Mutasi Barang, Laporan Penyusutan.
8. **BOS compliance**: Pembelian aset dari dana BOS harus tercatat sesuai juknis.

Klasifikasi aset sekolah:
- **Tanah**: Lahan sekolah, lapangan.
- **Peralatan & Mesin**: Komputer, printer, AC, genset, kendaraan.
- **Gedung & Bangunan**: Gedung sekolah, aula, musholla, gudang.
- **Jalan, Irigasi & Jaringan**: Jalan lingkungan sekolah, instalasi listrik/air.
- **Aset Tetap Lainnya**: Buku perpustakaan (koleksi), alat peraga, karya seni.
- **Konstruksi dalam Pengerjaan**: Bangunan yang sedang dibangun.

Volume data:
- Sekolah kecil: 100-500 aset.
- Sekolah besar: 1.000-5.000 aset.
- Maintenance log: 50-200 records per tahun.

### Mengapa Vernon Pattern?

- Master data aset = read-heavy, jarang berubah (kecuali saat opname).
- Maintenance log = has_many, moderate volume.
- Relasi ke class_room (ADR-011) untuk room-based mapping.
- Business logic moderate: depreciation calculation, condition tracking, lifecycle management.
- Eventually consistent acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `assets` (master aset) dan `asset_maintenances` (log pemeliharaan/mutasi).

### Table Schema

```sql
-- Master aset sekolah
CREATE TABLE assets (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Kode barang (Permendagri format)
    asset_code      VARCHAR(50) NOT NULL,
    register_number VARCHAR(20) NOT NULL,

    -- Identitas
    name            VARCHAR(300) NOT NULL,
    description     TEXT,
    brand           VARCHAR(100),
    model           VARCHAR(100),
    serial_number   VARCHAR(100),

    -- Klasifikasi
    asset_category  VARCHAR(30) NOT NULL,
    asset_subcategory VARCHAR(50),
    ownership_type  VARCHAR(20) NOT NULL DEFAULT 'sekolah',

    -- Pengadaan
    acquisition_method VARCHAR(20) NOT NULL,
    acquisition_date DATE NOT NULL,
    acquisition_source VARCHAR(200),
    acquisition_document VARCHAR(100),
    purchase_price  BIGINT NOT NULL DEFAULT 0,

    -- Lokasi
    location_type   VARCHAR(20) NOT NULL DEFAULT 'ruangan',
    location_name   VARCHAR(100),
    room_id         UUID,
    building        VARCHAR(100),
    floor           INT,

    -- Kuantitas
    quantity        INT NOT NULL DEFAULT 1,
    unit            VARCHAR(20) NOT NULL DEFAULT 'unit',

    -- Kondisi
    condition       VARCHAR(20) NOT NULL DEFAULT 'baik',
    last_opname_date DATE,
    last_opname_condition VARCHAR(20),

    -- Penyusutan (straight-line)
    useful_life_years INT,
    salvage_value   BIGINT NOT NULL DEFAULT 0,
    accumulated_depreciation BIGINT NOT NULL DEFAULT 0,
    book_value      BIGINT NOT NULL DEFAULT 0,

    -- Status lifecycle
    lifecycle_status VARCHAR(20) NOT NULL DEFAULT 'active',
    disposal_date   DATE,
    disposal_method VARCHAR(20),
    disposal_document VARCHAR(100),
    disposal_reason TEXT,

    -- Foto
    photo_url       VARCHAR(500),

    -- Tahun anggaran pengadaan
    fiscal_year     INT NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_asset_code UNIQUE (tenant_id, company_id, asset_code),
    CONSTRAINT chk_asset_category CHECK (asset_category IN (
        'tanah', 'peralatan_mesin', 'gedung_bangunan',
        'jalan_irigasi_jaringan', 'aset_tetap_lainnya',
        'konstruksi_dalam_pengerjaan'
    )),
    CONSTRAINT chk_asset_ownership CHECK (ownership_type IN (
        'bmd', 'yayasan', 'sekolah', 'hibah', 'pinjam_pakai'
    )),
    CONSTRAINT chk_asset_acquisition CHECK (acquisition_method IN (
        'pembelian', 'hibah', 'tukar_menukar', 'pembangunan',
        'sumbangan', 'dana_bos', 'apbd', 'other'
    )),
    CONSTRAINT chk_asset_condition CHECK (condition IN (
        'baik', 'kurang_baik', 'rusak_berat'
    )),
    CONSTRAINT chk_asset_lifecycle CHECK (lifecycle_status IN (
        'active', 'maintenance', 'disposed', 'transferred', 'lost'
    )),
    CONSTRAINT chk_asset_disposal_method CHECK (disposal_method IS NULL OR disposal_method IN (
        'lelang', 'pemusnahan', 'hibah', 'tukar_menukar', 'penyertaan_modal'
    )),
    CONSTRAINT chk_asset_location_type CHECK (location_type IN (
        'ruangan', 'lapangan', 'gedung', 'gudang', 'outdoor', 'other'
    )),
    CONSTRAINT chk_asset_price CHECK (purchase_price >= 0),
    CONSTRAINT chk_asset_salvage CHECK (salvage_value >= 0),
    CONSTRAINT chk_asset_depreciation CHECK (accumulated_depreciation >= 0),
    CONSTRAINT chk_asset_book_value CHECK (book_value >= 0),
    CONSTRAINT chk_asset_quantity CHECK (quantity >= 1)
);

-- Indexes
CREATE INDEX idx_asset_tenant_company ON assets (tenant_id, company_id);
CREATE INDEX idx_asset_code ON assets (asset_code);
CREATE INDEX idx_asset_category ON assets (asset_category);
CREATE INDEX idx_asset_condition ON assets (condition);
CREATE INDEX idx_asset_lifecycle ON assets (lifecycle_status);
CREATE INDEX idx_asset_room ON assets (room_id) WHERE room_id IS NOT NULL;
CREATE INDEX idx_asset_location ON assets (location_name);
CREATE INDEX idx_asset_acquisition_date ON assets (acquisition_date);
CREATE INDEX idx_asset_fiscal_year ON assets (fiscal_year);
CREATE INDEX idx_asset_disposal ON assets (disposal_date) WHERE lifecycle_status = 'disposed';
CREATE INDEX idx_asset_opname ON assets (last_opname_date);
CREATE INDEX idx_asset_name ON assets USING gin (to_tsvector('indonesian', name));
CREATE INDEX idx_asset_rels ON assets USING GIN (_rels);
CREATE INDEX idx_asset_data ON assets USING GIN (_data);

-- Log pemeliharaan dan mutasi aset
CREATE TABLE asset_maintenances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    asset_id        UUID NOT NULL,

    -- Tipe kegiatan
    maintenance_type VARCHAR(20) NOT NULL,

    -- Detail
    description     TEXT NOT NULL,
    performed_by    VARCHAR(200),
    vendor          VARCHAR(200),

    -- Biaya
    cost            BIGINT NOT NULL DEFAULT 0,
    cost_source     VARCHAR(20),

    -- Waktu
    maintenance_date DATE NOT NULL,
    completion_date DATE,

    -- Kondisi
    condition_before VARCHAR(20) NOT NULL,
    condition_after VARCHAR(20),

    -- Dokumen
    document_number VARCHAR(100),
    notes           TEXT,

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

    CONSTRAINT chk_maintenance_type CHECK (maintenance_type IN (
        'perbaikan', 'perawatan_rutin', 'penggantian_komponen',
        'kalibrasi', 'mutasi', 'opname', 'other'
    )),
    CONSTRAINT chk_maintenance_condition_before CHECK (condition_before IN (
        'baik', 'kurang_baik', 'rusak_berat'
    )),
    CONSTRAINT chk_maintenance_condition_after CHECK (condition_after IS NULL OR condition_after IN (
        'baik', 'kurang_baik', 'rusak_berat'
    )),
    CONSTRAINT chk_maintenance_cost CHECK (cost >= 0),
    CONSTRAINT chk_maintenance_cost_source CHECK (cost_source IS NULL OR cost_source IN (
        'bos', 'apbd', 'yayasan', 'komite', 'sumbangan', 'other'
    )),
    CONSTRAINT chk_maintenance_status CHECK (status IN (
        'scheduled', 'in_progress', 'completed', 'cancelled'
    ))
);

-- Indexes
CREATE INDEX idx_asset_maint_tenant_company ON asset_maintenances (tenant_id, company_id);
CREATE INDEX idx_asset_maint_asset ON asset_maintenances (asset_id);
CREATE INDEX idx_asset_maint_type ON asset_maintenances (maintenance_type);
CREATE INDEX idx_asset_maint_date ON asset_maintenances (maintenance_date);
CREATE INDEX idx_asset_maint_status ON asset_maintenances (status) WHERE status IN ('scheduled', 'in_progress');
CREATE INDEX idx_asset_maint_cost ON asset_maintenances (cost) WHERE cost > 0;
CREATE INDEX idx_asset_maint_rels ON asset_maintenances USING GIN (_rels);
CREATE INDEX idx_asset_maint_data ON asset_maintenances USING GIN (_data);
```

### Field Design Rationale

**assets:**

| Field | Keputusan | Alasan |
|---|---|---|
| `asset_code` | VARCHAR(50), UNIQUE | Kode barang sesuai Permendagri 19/2016: format XX.XX.XX.XX.XXX (golongan.bidang.kelompok.sub_kelompok.register) |
| `register_number` | VARCHAR(20) | Nomor register internal sekolah — untuk identifikasi cepat |
| `asset_category` | 6 kategori | Sesuai klasifikasi BMD Permendagri: tanah, peralatan_mesin, gedung_bangunan, dll |
| `ownership_type` | 5 tipe | BMD (sekolah negeri), yayasan (swasta), hibah, pinjam_pakai, sekolah (milik sendiri) |
| `acquisition_method` | 8 metode | Pembelian, hibah, dana_bos, apbd, dll — penting untuk audit dan pelaporan BOS |
| `condition` | 3 status | `baik`, `kurang_baik`, `rusak_berat` — sesuai standar Permendagri opname |
| `lifecycle_status` | 5 status | active → maintenance → disposed/transferred/lost |
| `useful_life_years` | INT, nullable | Umur manfaat untuk perhitungan penyusutan — nullable untuk tanah (tidak disusutkan) |
| `salvage_value` | BIGINT | Nilai residu setelah umur manfaat habis |
| `book_value` | BIGINT | Nilai buku = purchase_price - accumulated_depreciation |
| `room_id` | UUID, nullable | FK ke class_room (ADR-011) atau ruangan lain — nullable karena aset outdoor tidak punya room |
| `fiscal_year` | INT | Tahun anggaran pengadaan — untuk pelaporan keuangan |

**asset_maintenances:**

| Field | Keputusan | Alasan |
|---|---|---|
| `maintenance_type` | 7 tipe | perbaikan, perawatan_rutin, penggantian_komponen, kalibrasi, mutasi, opname, other |
| `condition_before/after` | VARCHAR(20) | Kondisi sebelum dan sesudah maintenance — untuk audit trail |
| `cost` | BIGINT | Biaya pemeliharaan dalam Rupiah |
| `cost_source` | VARCHAR(20), nullable | Sumber dana: BOS, APBD, yayasan, komite, sumbangan |
| `vendor` | VARCHAR, nullable | Vendor/teknisi yang melakukan perbaikan |
| `document_number` | VARCHAR, nullable | Nomor SPK, kuitansi, atau dokumen terkait |

### Depreciation Calculation (Application Logic)

```
Metode: Straight-Line (Garis Lurus)

annual_depreciation = (purchase_price - salvage_value) / useful_life_years
monthly_depreciation = annual_depreciation / 12
accumulated_depreciation = annual_depreciation × years_since_acquisition
book_value = purchase_price - accumulated_depreciation

Contoh:
- Komputer: purchase_price = Rp 10.000.000, useful_life = 4 tahun, salvage = Rp 1.000.000
- Annual depreciation = (10.000.000 - 1.000.000) / 4 = Rp 2.250.000/tahun
- Setelah 2 tahun: book_value = 10.000.000 - 4.500.000 = Rp 5.500.000

Umur manfaat default (Permendagri):
- Tanah: tidak disusutkan
- Peralatan & Mesin: 4-8 tahun
- Gedung & Bangunan: 20-50 tahun
- Kendaraan: 5-8 tahun
- Furniture: 5-10 tahun
- Komputer: 4 tahun
```

Depreciation dihitung secara **batch monthly** (cron job) dan disimpan di `accumulated_depreciation` + `book_value`.

### Asset Code Numbering (Permendagri Format)

```
Format: GG.BB.KK.SK.RRRR

GG = Golongan (2 digit)
  01 = Tanah
  02 = Peralatan dan Mesin
  03 = Gedung dan Bangunan
  04 = Jalan, Irigasi, dan Jaringan
  05 = Aset Tetap Lainnya
  06 = Konstruksi dalam Pengerjaan

BB = Bidang (2 digit)
KK = Kelompok (2 digit)
SK = Sub Kelompok (2 digit)
RRRR = Nomor Register (4 digit, auto-increment per sub-kelompok)

Contoh:
- 02.06.01.04.0001 = Peralatan > Komputer > PC Desktop > Register #1
- 03.11.01.01.0001 = Gedung > Bangunan Pendidikan > Gedung Sekolah > Register #1
```

### Vernon Relationships

**assets:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `room` | belongs_to | **Ya** | Nama ruangan untuk lokasi aset (jika room_id di-set) |

**asset_maintenances:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `asset` | belongs_to | **Ya** | Nama dan kode aset selalu ditampilkan |

### _rels / _data Structure

**assets:**
```json
{
  "_rels": {
    "room_id": "018f..."
  },
  "_data": {
    "room": { "id": "018f...", "name": "Lab Komputer 1", "building": "Gedung A", "floor": 2 }
  }
}
```

**asset_maintenances:**
```json
{
  "_rels": {
    "asset_id": "018f..."
  },
  "_data": {
    "asset": {
      "id": "018f...",
      "name": "Komputer Desktop Dell OptiPlex",
      "asset_code": "02.06.01.04.0001",
      "asset_category": "peralatan_mesin",
      "condition": "kurang_baik"
    }
  }
}
```

### API Endpoints

```
# Assets (Master)
GET    /api/v1/assets                                    — List aset (filter: category, condition, location, lifecycle)
POST   /api/v1/assets                                    — Registrasi aset baru
GET    /api/v1/assets/{id}                               — Detail aset + maintenance history
PUT    /api/v1/assets/{id}                               — Update aset
DELETE /api/v1/assets/{id}                               — Soft delete

# Asset by Location
GET    /api/v1/rooms/{id}/assets                         — Aset per ruangan
GET    /api/v1/assets/by-location                        — Summary aset per lokasi

# Condition & Lifecycle
PUT    /api/v1/assets/{id}/condition                     — Update kondisi
PUT    /api/v1/assets/{id}/dispose                       — Penghapusan aset
  Body: { disposal_method, disposal_date, disposal_document, disposal_reason }
PUT    /api/v1/assets/{id}/transfer                      — Mutasi aset ke lokasi lain
  Body: { new_room_id, new_location_name, reason }

# Maintenance
GET    /api/v1/asset-maintenances                        — List maintenance (filter: asset_id, type, date range)
POST   /api/v1/asset-maintenances                        — Catat pemeliharaan
PUT    /api/v1/asset-maintenances/{id}                   — Update maintenance
GET    /api/v1/assets/{id}/maintenances                  — Riwayat maintenance per aset

# Opname (Stock-Taking)
POST   /api/v1/assets/opname/start                       — Mulai opname periode
  Body: { opname_date, location_filter }
PUT    /api/v1/assets/{id}/opname                        — Update hasil opname per aset
  Body: { condition, notes, is_found }
GET    /api/v1/assets/opname/summary                     — Summary hasil opname

# Depreciation
POST   /api/v1/assets/depreciation/calculate             — Hitung penyusutan batch (monthly cron)
GET    /api/v1/assets/depreciation/report                — Laporan penyusutan
  Query: fiscal_year, category

# Reports
GET    /api/v1/assets/statistics                         — Statistik aset keseluruhan
GET    /api/v1/assets/kib-report                         — Kartu Inventaris Barang (per golongan)
GET    /api/v1/assets/mutation-report                    — Laporan mutasi barang (periode)
```

## Consequences

### Positive

- **Permendagri compliant**: Kode barang, klasifikasi, dan kondisi sesuai standar BMD Permendagri 19/2016.
- **Full lifecycle**: Dari pengadaan hingga penghapusan dengan audit trail di `asset_maintenances`.
- **Room-based mapping**: Aset bisa dipetakan per ruangan — berguna untuk opname dan accountability.
- **Depreciation built-in**: Perhitungan penyusutan straight-line otomatis — untuk laporan keuangan sekolah.
- **Opname support**: Workflow stock-taking tahunan dengan tracking kondisi.
- **Multi-ownership**: Mendukung BMD (negeri), yayasan (swasta), hibah, dan pinjam pakai.
- **BOS tracking**: `acquisition_method = 'dana_bos'` memudahkan pelaporan penggunaan dana BOS.

### Negative / Trade-offs

- **Complex asset_code**: Format Permendagri rumit — butuh UI yang user-friendly untuk input kode barang atau auto-generate.
- **Manual depreciation**: Batch calculation per bulan — bukan real-time. Book value bisa stale sampai batch berikutnya.
- **No barcode/QR**: Belum ada integrasi barcode/QR code untuk scanning aset saat opname — enhancement di masa depan.
- **Room reference sederhana**: `room_id` nullable UUID — belum ada tabel `rooms` terpisah. Saat ini reference ke `class_rooms` (ADR-011) atau string `location_name`.
- **No photo management**: Hanya 1 `photo_url` per aset — belum mendukung multiple photos.
- **Depreciation hanya straight-line**: Belum mendukung metode lain (declining balance, sum-of-years).

## Alternatives Considered

### 1. JSONB untuk maintenance history (embedded di assets)
- Ditolak: maintenance perlu query per tipe, per tanggal, per biaya — tabel terpisah lebih queryable.

### 2. Tabel terpisah per kategori aset (tanah, peralatan, gedung)
- Ditolak: duplikasi struktur — single table dengan `asset_category` discriminator lebih DRY. Field spesifik per kategori disimpan di JSONB `_data` jika perlu.

### 3. Depreciation sebagai tabel terpisah (asset_depreciations per bulan)
- Ditolak untuk MVP: kolom `accumulated_depreciation` dan `book_value` di master record cukup. Tabel history bisa ditambah untuk audit trail detail.

### 4. Integration dengan S039 (Lab Equipment)
- Considered: lab equipment bisa jadi subset aset. Keputusan: tetap terpisah karena lab equipment punya field spesifik (kalibrasi, safety category). Cross-reference via `asset_code` jika perlu.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `AssetDescriptor.TableName()` | — | `"assets"` |
| U02 | `AssetMaintenanceDescriptor.TableName()` | — | `"asset_maintenances"` |
| U03 | Validate rejects invalid `asset_category` | `"kendaraan"` | Error: invalid category |
| U04 | Validate rejects invalid `condition` | `"hancur"` | Error: must be baik/kurang_baik/rusak_berat |
| U05 | Validate rejects invalid `lifecycle_status` | `"expired"` | Error: invalid status |
| U06 | Validate rejects invalid `acquisition_method` | `"curian"` | Error: invalid method |
| U07 | Validate rejects invalid `ownership_type` | `"pribadi"` | Error: invalid type |
| U08 | Depreciation calculation: straight-line | price=10M, salvage=1M, life=4yr, age=2yr | book_value=5.5M |
| U09 | Depreciation calculation: tanah (no depreciation) | category=tanah, useful_life=null | book_value=purchase_price |
| U10 | Asset code format validation | `"02.06.01.04.0001"` | Valid |
| U11 | Asset code format validation (invalid) | `"99.99.99"` | Error: invalid format |
| U12 | Validate accepts valid asset | All fields valid | No error |

### Integration Tests — Assets

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create asset | POST with valid data | 201 |
| I02 | Unique asset_code per company | Create 2 assets same code | 409/422 |
| I03 | Category CHECK | INSERT with `asset_category = 'kendaraan'` | DB error |
| I04 | Condition CHECK | INSERT with `condition = 'hancur'` | DB error |
| I05 | Lifecycle CHECK | INSERT with `lifecycle_status = 'expired'` | DB error |
| I06 | Get assets by room | GET /rooms/{id}/assets | 200, filtered per ruangan |
| I07 | Full-text search by name | GET /assets?q=komputer | 200, matching assets |

### Integration Tests — Lifecycle

| # | Test Case | Action | Expected |
|---|---|---|---|
| I08 | Update condition | PUT /assets/{id}/condition to `kurang_baik` | 200 |
| I09 | Dispose asset | PUT /assets/{id}/dispose | 200, lifecycle_status=disposed |
| I10 | Transfer asset | PUT /assets/{id}/transfer to new room | 200, room_id updated, maintenance log created |
| I11 | Disposal method CHECK | PUT dispose with `method = 'buang'` | Error |

### Integration Tests — Maintenance

| # | Test Case | Action | Expected |
|---|---|---|---|
| I12 | Create maintenance | POST with valid data | 201 |
| I13 | Maintenance type CHECK | INSERT with `type = 'upgrade'` | DB error |
| I14 | Condition auto-update | POST maintenance with condition_after != condition_before | Asset condition updated |
| I15 | Get maintenance history | GET /assets/{id}/maintenances | 200, ordered by date desc |

### Integration Tests — Opname & Depreciation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Start opname | POST /assets/opname/start | 200, returns assets to verify |
| I17 | Record opname result | PUT /assets/{id}/opname | 200, last_opname_date updated |
| I18 | Calculate depreciation | POST /assets/depreciation/calculate | 200, book_value updated |
| I19 | Depreciation report | GET /assets/depreciation/report?fiscal_year=2026 | 200, summary by category |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I20 | ClassRoomUpdated syncs to assets | Update room name | `_data.room.name` updated |
| I21 | AssetUpdated syncs to maintenances | Update asset name | `_data.asset.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I22 | Cannot access other tenant's assets | GET with wrong tenant | 404 |
| I23 | Cannot create maintenance for other tenant's asset | POST cross-tenant | Error |
