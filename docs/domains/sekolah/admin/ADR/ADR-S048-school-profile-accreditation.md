# ADR-S048: School Profile & Accreditation (Profil Sekolah & Akreditasi)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Setiap sekolah di Indonesia memiliki data profil resmi yang digunakan untuk keperluan administrasi, pelaporan Dapodik, akreditasi, dan identitas di berbagai dokumen resmi (kop surat, rapor, ijazah). Data ini bersifat **master data** yang jarang berubah tetapi sering dibaca oleh hampir seluruh domain lain.

Kebutuhan profil sekolah:

1. **Identitas resmi**: NPSN (Nomor Pokok Sekolah Nasional), NSS (Nomor Statistik Sekolah) — wajib untuk Dapodik dan BAN.
2. **Kop surat**: S046 (Correspondence) memerlukan data sekolah untuk header surat otomatis.
3. **Rapor & Ijazah**: S018 (Rapor Generation) memerlukan nama sekolah, alamat, NPSN untuk cetak rapor.
4. **Visi Misi**: Ditampilkan di website sekolah, parent portal (S042), dan dokumen akreditasi.
5. **Akreditasi**: Status akreditasi sekolah harus ter-track — ini menentukan kredibilitas dan banyak kebijakan pendidikan.
6. **Pesantren dual-mode**: Untuk lembaga pesantren (ADR-009), profil mencakup data tambahan pondok pesantren (jumlah santri mukim, santri kalong, asatidz).
7. **Statistik sekolah**: Jumlah siswa, guru, rombel, ruang kelas — derived dari domain lain tetapi perlu di-cache untuk quick access.

Akreditasi sekolah di Indonesia:

| Lembaga Akreditasi | Jenjang | Peringkat |
|---------------------|---------|-----------|
| BAN-S (Badan Akreditasi Nasional Sekolah) | SD, SMP | Unggul (A), Baik Sekali (B), Baik (C), Tidak Terakreditasi |
| BAN-SM (Badan Akreditasi Nasional Sekolah Menengah) | SMA, SMK | Unggul (A), Baik Sekali (B), Baik (C), Tidak Terakreditasi |
| BAN-PNF | PKBM, Kursus | Sama |

Akreditasi berlaku 5 tahun dan harus diperpanjang. Sekolah menyimpan riwayat akreditasi untuk audit.

### Mengapa Vernon Pattern?

- **Sangat read-heavy**: profil sekolah dibaca oleh hampir semua domain (kop surat, rapor, dashboard, pelaporan).
- **Sangat jarang write**: data profil berubah mungkin 1-2x per tahun, akreditasi 1x per 5 tahun.
- **Ideal untuk _data cache**: data statistik (jumlah siswa, guru, rombel) di-cache di _data — menghindari COUNT query berulang.
- **Relasi ke banyak domain**: school profile menjadi sumber data untuk S046, S018, S042, S050, dll.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `school_profiles` (data profil sekolah) dan `school_accreditations` (riwayat akreditasi).

### Table Schema

```sql
-- Profil sekolah (1 row per company_id)
CREATE TABLE school_profiles (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas resmi
    school_name     VARCHAR(255) NOT NULL,
    npsn            VARCHAR(20),
    nss             VARCHAR(20),
    nis             VARCHAR(20),
    school_level    VARCHAR(20) NOT NULL,
    school_status   VARCHAR(20) NOT NULL,

    -- Alamat
    address         TEXT NOT NULL,
    village         VARCHAR(100),
    district        VARCHAR(100),
    city            VARCHAR(100) NOT NULL,
    province        VARCHAR(100) NOT NULL,
    postal_code     VARCHAR(10),
    latitude        DECIMAL(10, 8),
    longitude       DECIMAL(11, 8),

    -- Kontak
    phone           VARCHAR(30),
    fax             VARCHAR(30),
    email           VARCHAR(255),
    website         VARCHAR(255),

    -- Kepala sekolah
    principal_name  VARCHAR(255),
    principal_nip   VARCHAR(30),
    principal_id    UUID,

    -- Visi misi
    vision          TEXT,
    mission         TEXT,
    goals           TEXT,
    motto           VARCHAR(500),

    -- Yayasan (untuk swasta)
    foundation_name VARCHAR(255),
    foundation_deed_number VARCHAR(100),
    foundation_deed_date DATE,

    -- Pesantren (ADR-009 dual-mode)
    is_pesantren    BOOLEAN NOT NULL DEFAULT false,
    pesantren_name  VARCHAR(255),
    pesantren_nspp  VARCHAR(30),
    pesantren_type  VARCHAR(30),

    -- Branding
    logo_url        VARCHAR(500),
    header_image_url VARCHAR(500),

    -- Akreditasi terakhir (denormalized dari school_accreditations)
    current_accreditation_grade VARCHAR(20),
    current_accreditation_year  INT,
    current_accreditation_score DECIMAL(5, 2),

    -- Statistik (denormalized, updated via events)
    total_students      INT NOT NULL DEFAULT 0,
    total_teachers      INT NOT NULL DEFAULT 0,
    total_staff         INT NOT NULL DEFAULT 0,
    total_class_rooms   INT NOT NULL DEFAULT 0,
    total_rombel        INT NOT NULL DEFAULT 0,

    -- Pesantren statistics (ADR-009)
    total_santri_mukim  INT NOT NULL DEFAULT 0,
    total_santri_kalong INT NOT NULL DEFAULT 0,
    total_asatidz       INT NOT NULL DEFAULT 0,

    -- Surat numbering config (used by S046)
    letter_code     VARCHAR(20),
    letter_number_format VARCHAR(100) DEFAULT '{kode_sekolah}/{nomor_urut}/{bulan_romawi}/{tahun}',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_profile_school_level CHECK (school_level IN (
        'tk', 'sd', 'smp', 'sma', 'smk', 'slb', 'mi', 'mts', 'ma', 'mak'
    )),
    CONSTRAINT chk_profile_school_status CHECK (school_status IN (
        'negeri', 'swasta'
    )),
    CONSTRAINT chk_profile_pesantren_type CHECK (
        pesantren_type IS NULL OR pesantren_type IN (
            'salafiyah', 'modern', 'kombinasi'
        )
    ),
    CONSTRAINT chk_profile_accreditation_grade CHECK (
        current_accreditation_grade IS NULL OR current_accreditation_grade IN (
            'A', 'B', 'C', 'not_accredited'
        )
    ),
    CONSTRAINT uq_profile_company UNIQUE (tenant_id, company_id)
);

-- Indexes
CREATE INDEX idx_profile_tenant_company ON school_profiles (tenant_id, company_id);
CREATE INDEX idx_profile_npsn ON school_profiles (npsn) WHERE npsn IS NOT NULL;
CREATE INDEX idx_profile_nss ON school_profiles (nss) WHERE nss IS NOT NULL;
CREATE INDEX idx_profile_school_level ON school_profiles (school_level);
CREATE INDEX idx_profile_city ON school_profiles (city);
CREATE INDEX idx_profile_province ON school_profiles (province);
CREATE INDEX idx_profile_pesantren ON school_profiles (is_pesantren) WHERE is_pesantren = true;
CREATE INDEX idx_profile_rels ON school_profiles USING GIN (_rels);
CREATE INDEX idx_profile_data ON school_profiles USING GIN (_data);


-- Riwayat akreditasi
CREATE TABLE school_accreditations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Referensi profil
    school_profile_id UUID NOT NULL,

    -- Akreditasi data
    accreditation_body VARCHAR(20) NOT NULL,
    grade           VARCHAR(20) NOT NULL,
    score           DECIMAL(5, 2),
    certificate_number VARCHAR(100),

    -- Period
    accreditation_date DATE NOT NULL,
    valid_until     DATE,

    -- Dokumen
    certificate_url VARCHAR(500),
    report_url      VARCHAR(500),
    supporting_docs JSONB NOT NULL DEFAULT '[]',

    -- Status
    is_current      BOOLEAN NOT NULL DEFAULT false,
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_accreditation_body CHECK (accreditation_body IN (
        'ban_s', 'ban_sm', 'ban_pnf', 'ban_pt', 'lam'
    )),
    CONSTRAINT chk_accreditation_grade CHECK (grade IN (
        'A', 'B', 'C', 'not_accredited'
    )),
    CONSTRAINT chk_accreditation_score CHECK (
        score IS NULL OR (score >= 0 AND score <= 100)
    )
);

-- Indexes
CREATE INDEX idx_accreditation_tenant_company ON school_accreditations (tenant_id, company_id);
CREATE INDEX idx_accreditation_profile ON school_accreditations (school_profile_id);
CREATE INDEX idx_accreditation_current ON school_accreditations (is_current) WHERE is_current = true;
CREATE INDEX idx_accreditation_date ON school_accreditations (accreditation_date);
CREATE INDEX idx_accreditation_grade ON school_accreditations (grade);
CREATE INDEX idx_accreditation_body ON school_accreditations (accreditation_body);
CREATE INDEX idx_accreditation_rels ON school_accreditations USING GIN (_rels);
CREATE INDEX idx_accreditation_data ON school_accreditations USING GIN (_data);
```

### Field Design Rationale

**school_profiles:**

| Field | Keputusan | Alasan |
|---|---|---|
| `npsn` | VARCHAR(20), nullable | 8-digit kode unik dari Kemendikbud — nullable karena sekolah baru mungkin belum punya |
| `nss` | VARCHAR(20), nullable | Nomor Statistik Sekolah — legacy identifier, masih dipakai beberapa daerah |
| `school_level` | VARCHAR(20), CHECK | Jenjang sekolah — termasuk MI/MTs/MA untuk madrasah |
| `school_status` | VARCHAR(20), CHECK | Negeri vs swasta — menentukan banyak aturan (sumber dana, akreditasi) |
| `principal_id` | UUID, nullable | FK ke users/teachers — nullable karena bisa kosong saat setup awal |
| `is_pesantren` | BOOLEAN | Flag dual-mode dari ADR-009 — jika true, field pesantren_* wajib diisi |
| `pesantren_nspp` | VARCHAR(30), nullable | Nomor Statistik Pondok Pesantren — ID resmi dari Kemenag |
| `pesantren_type` | VARCHAR(30), nullable | salafiyah (tradisional), modern, kombinasi |
| `total_students` | INT, default 0 | Denormalisasi dari S001 — updated via event `StudentCountChanged` |
| `total_santri_mukim` / `total_santri_kalong` | INT, default 0 | Statistik pesantren: santri mukim = tinggal di asrama, santri kalong = pulang pergi |
| `letter_code` | VARCHAR(20) | Kode sekolah untuk penomoran surat (dipakai S046) |
| `letter_number_format` | VARCHAR(100) | Format default nomor surat — konfigurabel per sekolah |
| `current_accreditation_grade` | VARCHAR(20) | Denormalisasi dari `school_accreditations` — quick access tanpa JOIN |

**school_accreditations:**

| Field | Keputusan | Alasan |
|---|---|---|
| `accreditation_body` | VARCHAR(20), CHECK | BAN-S, BAN-SM, BAN-PNF — tergantung jenjang sekolah |
| `grade` | VARCHAR(20), CHECK | A/B/C/not_accredited — standar nasional |
| `score` | DECIMAL(5,2), nullable | Nilai akreditasi (0-100) — nullable karena sekolah lama mungkin tidak punya data skor |
| `valid_until` | DATE, nullable | Akreditasi berlaku 5 tahun — nullable untuk data historis |
| `is_current` | BOOLEAN | Flag akreditasi yang sedang berlaku — hanya 1 yang true per sekolah |
| `supporting_docs` | JSONB array | URL dokumen pendukung (8 standar nasional pendidikan) |

### Vernon Relationships

**school_profiles:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `principal` | belongs_to | **Ya** | Nama kepsek sering ditampilkan bersamaan dengan profil |
| `accreditations` | has_many | Tidak | History dimuat terpisah di detail view |

**school_accreditations:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `school_profile` | belongs_to | **Ya** | Selalu dalam konteks sekolah tertentu |

### _rels / _data Structure

**school_profiles:**

```json
{
  "_rels": {
    "principal_id": "018f..."
  },
  "_data": {
    "principal": {
      "id": "018f...",
      "full_name": "Drs. Ahmad Suryadi, M.Pd.",
      "nip": "198005152005011001"
    },
    "current_accreditation": {
      "grade": "A",
      "score": 93.5,
      "accreditation_date": "2024-11-15",
      "valid_until": "2029-11-15",
      "accreditation_body": "ban_sm"
    },
    "statistics": {
      "total_students": 487,
      "total_teachers": 32,
      "total_staff": 15,
      "total_class_rooms": 18,
      "total_rombel": 15,
      "student_teacher_ratio": 15.2
    },
    "pesantren": {
      "total_santri_mukim": 120,
      "total_santri_kalong": 367,
      "total_asatidz": 18
    }
  }
}
```

**school_accreditations:**

```json
{
  "_rels": {
    "school_profile_id": "018f..."
  },
  "_data": {
    "school_profile": {
      "id": "018f...",
      "school_name": "SMA Negeri 1 Bandung",
      "npsn": "20219432"
    }
  }
}
```

### Statistics Update via Events

Statistik di `school_profiles` di-update secara async melalui event dari domain lain:

```
S001 (Student) → StudentCountChanged → update total_students
S033 (Dormitory) → SantriCountChanged → update total_santri_mukim, total_santri_kalong
ADR-012 (Teachers) → TeacherCountChanged → update total_teachers, total_staff, total_asatidz
ADR-011 (ClassRooms) → ClassRoomCountChanged → update total_class_rooms, total_rombel
```

Ini menghindari COUNT query real-time yang bisa lambat pada sekolah besar.

### API Endpoints

```
# School Profile
GET    /api/v1/school-profile                              — Get profil sekolah (current tenant/company)
PUT    /api/v1/school-profile                              — Update profil sekolah
PATCH  /api/v1/school-profile/visi-misi                    — Update visi, misi, tujuan saja
PATCH  /api/v1/school-profile/pesantren                    — Update data pesantren
GET    /api/v1/school-profile/statistics                    — Get statistik terkini
POST   /api/v1/school-profile/logo                         — Upload logo sekolah

# Accreditation
POST   /api/v1/school-accreditations                       — Tambah riwayat akreditasi baru
GET    /api/v1/school-accreditations                       — List riwayat akreditasi
GET    /api/v1/school-accreditations/{id}                  — Detail akreditasi
PUT    /api/v1/school-accreditations/{id}                  — Update data akreditasi
PUT    /api/v1/school-accreditations/{id}/set-current      — Set sebagai akreditasi aktif
DELETE /api/v1/school-accreditations/{id}                  — Soft delete

# Public (untuk website sekolah)
GET    /api/v1/public/school-profile                       — Profil publik (tanpa auth — untuk website)
```

## Consequences

### Positive

- **Single source of truth**: Seluruh data profil sekolah terpusat — digunakan oleh kop surat (S046), rapor (S018), parent portal (S042).
- **Akreditasi ter-track**: Riwayat akreditasi lengkap — berguna saat visitasi BAN dan audit.
- **Statistik real-time**: Dashboard manajemen bisa melihat jumlah siswa, guru, rombel tanpa query berat.
- **Pesantren support**: Data pondok pesantren terintegrasi — tidak perlu tabel terpisah untuk lembaga dual-mode.
- **Vernon optimal**: Sangat read-heavy, sangat jarang write — Vernon pattern memberikan performance terbaik di sini.
- **Denormalisasi akreditasi**: `current_accreditation_grade` di profil menghindari JOIN ke tabel akreditasi untuk display sederhana.

### Negative / Trade-offs

- **God table risk**: `school_profiles` memiliki banyak kolom (~40) — tetapi ini legitimate karena memang 1 entity "profil sekolah" yang comprehensive.
- **Statistics staleness**: Statistik bersifat eventual consistent — bisa stale beberapa detik setelah operasi CRUD di domain lain.
- **Pesantren columns unused**: Untuk sekolah non-pesantren, kolom `pesantren_*` dan `total_santri_*` tidak terpakai — trade-off untuk menghindari tabel terpisah.
- **File storage**: Logo dan dokumen akreditasi memerlukan object storage terpisah.
- **Public API security**: Endpoint publik tanpa auth perlu rate limiting dan field filtering agar tidak expose data sensitif.

## Alternatives Considered

### 1. School profile dan pesantren profile di tabel terpisah
- Ditolak: memperumit query tanpa manfaat signifikan. Flag `is_pesantren` + kolom opsional lebih sederhana (ADR-009 pattern).

### 2. Statistik dihitung real-time (COUNT query)
- Ditolak: COUNT pada tabel besar (siswa, guru) bisa lambat. Denormalisasi + event-driven update jauh lebih performant untuk read-heavy use case.

### 3. Akreditasi sebagai JSONB array di school_profiles
- Ditolak: perlu di-query per tanggal, per grade, per body — tabel terpisah memungkinkan proper indexing.

### 4. Separate EAV (Entity-Attribute-Value) untuk extensible profile fields
- Ditolak: EAV anti-pattern — query lambat, validasi sulit. Kolom eksplisit lebih aman dan performant.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` returns correct name | — | `"school_profiles"` |
| U02 | `DefaultRels()` returns autoloaded rels | — | `principal` |
| U03 | Validate rejects invalid school_level | `{ "school_level": "universitas" }` | Error: school_level not in allowed list |
| U04 | Validate rejects invalid school_status | `{ "school_status": "hybrid" }` | Error: school_status must be negeri/swasta |
| U05 | Validate pesantren fields when is_pesantren=true | `{ "is_pesantren": true, "pesantren_name": null }` | Error: pesantren_name required when is_pesantren |
| U06 | Validate accreditation score range | `{ "score": 105 }` | Error: score must be 0-100 |
| U07 | Validate accreditation grade | `{ "grade": "D" }` | Error: grade must be A/B/C/not_accredited |
| U08 | Validate NPSN format | `{ "npsn": "123" }` | Error: NPSN must be 8 digits |
| U09 | Validate only one current accreditation | Set is_current=true on second record | Previous is_current set to false |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Get school profile | GET /school-profile | 200, complete profile with _data |
| I02 | Update school profile | PUT with updated address | 200, updated_at changed |
| I03 | Update visi misi only | PATCH /visi-misi with new vision | 200, only vision/mission/goals updated |
| I04 | Add accreditation record | POST accreditation with grade=A | 201, is_current=true, profile.current_accreditation_grade updated |
| I05 | Set new current accreditation | PUT set-current on older record | 200, previous current = false, new = true, profile denorm updated |
| I06 | Upload logo | POST /logo with image | 200, logo_url set |
| I07 | Pesantren profile update | PATCH /pesantren with santri data | 200, pesantren fields updated |
| I08 | Public profile endpoint | GET /public/school-profile (no auth) | 200, limited fields (no internal data) |
| I09 | Statistics reflect current data | After student enrollment | total_students incremented via event |
| I10 | Profile used by S046 for letter header | Generate surat from template | school_name, address, NPSN from school_profiles |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | PrincipalUpdated syncs to profile | Update principal user data | `_data.principal.full_name` updated |
| I12 | AccreditationAdded syncs to profile | Add new accreditation | `_data.current_accreditation` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | Cannot access other tenant's profile | GET with wrong tenant scope | 404 |
| I14 | One profile per company | POST second profile for same company | 409, unique constraint violation |
| I15 | Public endpoint is tenant-scoped | GET /public/school-profile requires tenant identification | Returns only that tenant's public data |
