# ADR-S007: Student Address & Previous School

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

ADR-S002 mendefinisikan `address` dan `previous_school` sebagai on-demand relationship dari student, namun belum ada model data detailnya. Kedua data ini diperlukan untuk:

1. **Dapodik**: Pelaporan ke Kemendikbud mewajibkan alamat siswa dan asal sekolah.
2. **Rapor**: Alamat siswa tercantum di rapor.
3. **Zonasi**: Jarak rumah ke sekolah menentukan penerimaan PPDB jalur zonasi.
4. **Administrasi**: Surat menyurat memerlukan alamat lengkap siswa.
5. **Riwayat pendidikan**: Data sekolah asal diperlukan untuk siswa pindahan dan verifikasi ijazah.

Alamat siswa bisa **berbeda** dari alamat orang tua/wali (misal: siswa kos/pondok). Satu siswa memiliki tepat **satu alamat aktif** dan **nol atau satu sekolah asal**.

### Mengapa Vernon Pattern?

- belongs_to dari student (1:1 relationship).
- Read-heavy: profil siswa, rapor, dan laporan memerlukan akses cepat.
- Business logic sederhana: CRUD.
- Eventual consistency acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk dua tabel: `student_addresses` (alamat siswa) dan `student_previous_schools` (riwayat sekolah sebelumnya).

### Table Schema

```sql
-- Alamat siswa
CREATE TABLE student_addresses (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign key
    student_id      UUID NOT NULL,

    -- Alamat lengkap (format Indonesia)
    address         TEXT NOT NULL,
    rt              VARCHAR(3),
    rw              VARCHAR(3),
    village         VARCHAR(100),
    district        VARCHAR(100),
    city            VARCHAR(100) NOT NULL,
    province        VARCHAR(100) NOT NULL,
    postal_code     VARCHAR(5),

    -- Koordinat (untuk zonasi PPDB)
    latitude        NUMERIC(10,7),
    longitude       NUMERIC(10,7),

    -- Status tempat tinggal
    living_with     VARCHAR(20) NOT NULL DEFAULT 'parents',
    distance_km     NUMERIC(5,1),

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_student_address UNIQUE (student_id),
    CONSTRAINT chk_living_with CHECK (living_with IN ('parents', 'guardian', 'boarding', 'alone', 'other'))
);

-- Indexes
CREATE INDEX idx_student_addr_tenant_company ON student_addresses (tenant_id, company_id);
CREATE INDEX idx_student_addr_student ON student_addresses (student_id);
CREATE INDEX idx_student_addr_city ON student_addresses (city);
CREATE INDEX idx_student_addr_rels ON student_addresses USING GIN (_rels);
CREATE INDEX idx_student_addr_data ON student_addresses USING GIN (_data);

-- Sekolah asal
CREATE TABLE student_previous_schools (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign key
    student_id      UUID NOT NULL,

    -- Data sekolah asal
    school_name     VARCHAR(255) NOT NULL,
    npsn            VARCHAR(8),
    school_address  TEXT,
    school_city     VARCHAR(100),
    school_province VARCHAR(100),

    -- Data akademik asal
    last_class      VARCHAR(20),
    exit_year       INT,
    certificate_no  VARCHAR(50),

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_student_prev_school UNIQUE (student_id)
);

-- Indexes
CREATE INDEX idx_prev_school_tenant_company ON student_previous_schools (tenant_id, company_id);
CREATE INDEX idx_prev_school_student ON student_previous_schools (student_id);
CREATE INDEX idx_prev_school_rels ON student_previous_schools USING GIN (_rels);
CREATE INDEX idx_prev_school_data ON student_previous_schools USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `address` | TEXT, NOT NULL | Alamat jalan/dusun/kampung — wajib |
| `rt`/`rw` | VARCHAR(3), nullable | Tidak semua daerah menggunakan RT/RW |
| `latitude`/`longitude` | NUMERIC, nullable | Opsional — untuk zonasi PPDB |
| `living_with` | VARCHAR(20), CHECK | Konteks tempat tinggal: ortu, wali, asrama, sendiri |
| `distance_km` | NUMERIC(5,1), nullable | Jarak ke sekolah dalam km, presisi 1 desimal |
| `npsn` | VARCHAR(8), nullable | Nomor Pokok Sekolah Nasional, tidak semua sekolah terdaftar |
| `certificate_no` | VARCHAR(50), nullable | Nomor ijazah/SKHUN sekolah sebelumnya |

### Vernon Relationships

**student_addresses:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Selalu perlu tahu pemilik alamat |

**student_previous_schools:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Selalu perlu tahu pemilik riwayat |

### _rels / _data Structure

```json
// student_addresses
{
  "_rels": { "student_id": "018f..." },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" }
  }
}

// student_previous_schools
{
  "_rels": { "student_id": "018f..." },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" }
  }
}
```

### API Endpoints

```
GET    /api/v1/students/{id}/address             — Alamat siswa
POST   /api/v1/student-addresses                  — Buat/set alamat siswa
PUT    /api/v1/student-addresses/{id}             — Update alamat

GET    /api/v1/students/{id}/previous-school      — Sekolah asal siswa
POST   /api/v1/student-previous-schools            — Buat/set sekolah asal
PUT    /api/v1/student-previous-schools/{id}      — Update sekolah asal
```

## Consequences

### Positive

- **1:1 terjamin**: Unique constraint pada `student_id` mencegah duplikasi.
- **Zonasi siap**: Koordinat dan jarak mendukung PPDB zonasi.
- **Dapodik compliant**: Field set alamat mengikuti format Dapodik.
- **Terpisah dari guardian**: Alamat siswa independen dari alamat orang tua.

### Negative / Trade-offs

- **Dua tabel**: Address dan previous school bisa digabung jadi satu tabel "student details", tapi dipisah untuk clarity.
- **Koordinat manual**: Latitude/longitude harus diisi manual atau via geocoding service eksternal.
- **NPSN lookup**: Validasi NPSN memerlukan data referensi Kemendikbud yang belum ada.

## Alternatives Considered

### 1. Alamat sebagai kolom di Students Table
- Ditolak: students table sudah cukup besar, alamat lebih baik sebagai entity terpisah untuk maintainability.

### 2. Gabung address + previous_school dalam satu tabel
- Ditolak: dua domain yang berbeda — alamat bisa berubah berkali-kali, sekolah asal statis.

### 3. JSONB untuk alamat
- Ditolak: kolom terstruktur lebih mudah di-query untuk zonasi (filter by city, province).

## Test Cases

### Unit Tests — Descriptor & Validation

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `AddressDescriptor.TableName()` returns correct name | — | `"student_addresses"` |
| U02 | `PrevSchoolDescriptor.TableName()` returns correct name | — | `"student_previous_schools"` |
| U03 | Validate rejects empty `address` | `{ "address": "" }` | Error: address required |
| U04 | Validate rejects empty `city` | `{ ..., "city": "" }` | Error: city required |
| U05 | Validate rejects empty `province` | `{ ..., "province": "" }` | Error: province required |
| U06 | Validate rejects invalid `living_with` | `{ ..., "living_with": "dorm" }` | Error: invalid living_with |
| U07 | Validate accepts valid address with minimal fields | `address` + `city` + `province` + `student_id` | No error |
| U08 | Validate rejects empty `school_name` for previous school | `{ "school_name": "" }` | Error: school_name required |

### Integration Tests — Student Address

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create student address | `POST /api/v1/student-addresses` | 201, address created with `_data` |
| I02 | Get student address | `GET /api/v1/students/{id}/address` | 200, returns address |
| I03 | Update student address | `PUT /api/v1/student-addresses/{id}` | 200, fields updated |
| I04 | Unique per student | Create 2 addresses for same student | 409/422, unique constraint |
| I05 | Living_with CHECK enforced | INSERT with `living_with = 'dorm'` | DB error, CHECK violation |
| I06 | Nullable coordinates accepted | Create without latitude/longitude | 201, created with nulls |

### Integration Tests — Previous School

| # | Test Case | Action | Expected |
|---|---|---|---|
| I07 | Create previous school | `POST /api/v1/student-previous-schools` | 201, record created |
| I08 | Get previous school | `GET /api/v1/students/{id}/previous-school` | 200, returns record |
| I09 | Unique per student | Create 2 previous schools for same student | 409/422, unique constraint |
| I10 | NPSN nullable | Create without `npsn` | 201, allowed |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | StudentUpdated syncs to address | Update student `full_name` | `_data.student.full_name` updated in address |
| I12 | StudentUpdated syncs to previous school | Update student `full_name` | `_data.student.full_name` updated in previous school |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I13 | Cannot access other tenant's address | GET with wrong tenant scope | 404 |
| I14 | Cannot access other tenant's previous school | GET with wrong tenant scope | 404 |
