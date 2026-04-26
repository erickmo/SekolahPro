# ADR-S013: Student Achievement / Prestasi

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Pencatatan prestasi siswa diperlukan untuk:

1. **Portofolio siswa**: Dokumentasi lomba, sertifikasi, dan penghargaan.
2. **Rapor**: Catatan prestasi dilampirkan di rapor.
3. **Beasiswa**: Prestasi menjadi salah satu pertimbangan beasiswa.
4. **Promosi sekolah**: Data prestasi untuk branding sekolah.
5. **Dapodik**: Pelaporan prestasi ke Kemendikbud.
6. **PPDB**: Jalur prestasi memerlukan data terstruktur.

Setiap prestasi = satu event spesifik (lomba tertentu, sertifikasi tertentu). Satu siswa bisa punya banyak prestasi.

### Mengapa Vernon Pattern?

- has_many dari student.
- Read-heavy: portofolio, rapor, laporan.
- Business logic sederhana: CRUD.
- Eventual consistency acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk domain `student_achievement`.

### Table Schema

```sql
CREATE TABLE student_achievements (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Detail prestasi
    achievement_type VARCHAR(20) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    level           VARCHAR(20) NOT NULL,
    rank            VARCHAR(30),
    organizer       VARCHAR(255),
    achievement_date DATE NOT NULL,

    -- Bukti
    certificate_url TEXT,
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

    CONSTRAINT chk_achievement_type CHECK (achievement_type IN (
        'academic', 'sports', 'arts', 'science', 'technology',
        'religious', 'social', 'other'
    )),
    CONSTRAINT chk_achievement_level CHECK (level IN (
        'school', 'district', 'city', 'province', 'national', 'international'
    ))
);

-- Indexes
CREATE INDEX idx_achievement_tenant_company ON student_achievements (tenant_id, company_id);
CREATE INDEX idx_achievement_student ON student_achievements (student_id);
CREATE INDEX idx_achievement_year ON student_achievements (academic_year_id);
CREATE INDEX idx_achievement_type ON student_achievements (achievement_type);
CREATE INDEX idx_achievement_level ON student_achievements (level);
CREATE INDEX idx_achievement_rels ON student_achievements USING GIN (_rels);
CREATE INDEX idx_achievement_data ON student_achievements USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `achievement_type` | VARCHAR(20), CHECK | 8 kategori standar prestasi sekolah |
| `level` | VARCHAR(20), CHECK | 6 level: sekolah → internasional (standar Dapodik) |
| `rank` | VARCHAR(30), nullable | Peringkat fleksibel: "Juara 1", "Finalis", "Medali Emas", dll |
| `organizer` | VARCHAR(255), nullable | Penyelenggara lomba/kegiatan |
| `certificate_url` | TEXT, nullable | Link ke dokumen sertifikat (bisa reference S010) |
| `certificate_no` | VARCHAR(50), nullable | Nomor sertifikat/piagam |

### Vernon Relationships

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Pemilik prestasi |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

### _rels / _data Structure

```json
{
  "_rels": {
    "student_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

### API Endpoints

```
GET    /api/v1/students/{id}/achievements           — List prestasi siswa
POST   /api/v1/student-achievements                  — Catat prestasi
PUT    /api/v1/student-achievements/{id}            — Update prestasi
DELETE /api/v1/student-achievements/{id}            — Soft delete
GET    /api/v1/student-achievements?level=national   — Filter by level (admin)
```

## Consequences

### Positive

- **Portofolio lengkap**: Semua prestasi terdokumentasi dengan bukti.
- **Dapodik aligned**: Level prestasi sesuai standar Kemendikbud.
- **Fleksibel**: `rank` sebagai free text mengakomodasi berbagai format peringkat.
- **Searchable**: Index per type dan level memudahkan laporan.

### Negative / Trade-offs

- **Rank tidak standar**: Free text rank tidak bisa di-sort atau di-compare secara otomatis.
- **Certificate storage**: Bergantung pada object storage (sama seperti S010).
- **Tidak ada tim**: Prestasi grup (misalnya tim basket) belum di-handle — setiap anggota tim harus diinput terpisah.

## Alternatives Considered

### 1. Gabung dengan S012 (discipline)
- Ditolak: prestasi dan disiplin punya nature berbeda — prestasi adalah portfolio, disiplin adalah pembinaan.

### 2. JSONB array di student
- Ditolak: tidak bisa filter per level/type, tidak bisa paginate.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` | — | `"student_achievements"` |
| U02 | Validate rejects invalid `achievement_type` | `"music"` | Error |
| U03 | Validate rejects invalid `level` | `"regional"` | Error |
| U04 | Validate rejects empty `title` | `""` | Error |
| U05 | Validate accepts valid achievement | All fields valid | No error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create achievement | POST with valid data | 201 |
| I02 | List student achievements | GET /students/{id}/achievements | 200 |
| I03 | Filter by type | GET ?achievement_type=academic | 200, filtered |
| I04 | Filter by level | GET ?level=national | 200, filtered |
| I05 | Type CHECK enforced | INSERT with `achievement_type = 'music'` | DB error |
| I06 | Level CHECK enforced | INSERT with `level = 'regional'` | DB error |
| I07 | StudentUpdated syncs | Update student name | `_data.student` updated |
| I08 | Tenant isolation | GET with wrong tenant | 404 |
