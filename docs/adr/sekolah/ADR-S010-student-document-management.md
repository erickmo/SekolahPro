# ADR-S010: Student Document Management

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Sekolah menyimpan berbagai dokumen terkait siswa yang diperlukan untuk administrasi, pelaporan, dan arsip. Dokumen yang perlu dikelola:

1. **Dokumen pendaftaran**: Akta kelahiran, Kartu Keluarga, foto.
2. **Dokumen akademik**: Rapor sekolah sebelumnya, ijazah, SKHUN.
3. **Dokumen kesehatan**: Surat keterangan sehat, kartu vaksinasi.
4. **Dokumen beasiswa**: Surat keterangan tidak mampu, DTKS.
5. **Surat menyurat**: Surat pindah, surat keterangan lulus.

Karakteristik:
- File disimpan di **object storage** (S3/MinIO), bukan di database.
- Database menyimpan **metadata** saja: nama file, tipe, URL, siapa yang upload.
- Satu siswa bisa punya **banyak dokumen** dari berbagai kategori.
- Dokumen bersifat **immutable** — versi baru = upload baru, bukan overwrite.

### Mengapa Vernon Pattern?

- has_many dari student (banyak dokumen per siswa).
- Read-heavy: admin perlu lihat daftar dokumen saat verifikasi.
- Business logic sederhana: upload metadata, list, soft delete.
- Eventual consistency acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk domain `student_document` yang menyimpan metadata dokumen.

### Table Schema

```sql
CREATE TABLE student_documents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign key
    student_id      UUID NOT NULL,

    -- Document metadata
    document_type   VARCHAR(30) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    file_url        TEXT NOT NULL,
    file_size       BIGINT NOT NULL,
    mime_type       VARCHAR(100) NOT NULL,

    -- Audit
    uploaded_by     UUID NOT NULL,
    verified_by     UUID,
    verified_at     TIMESTAMPTZ,
    verification_note TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_document_type CHECK (document_type IN (
        'akta_lahir', 'kartu_keluarga', 'ktp_ortu', 'foto',
        'ijazah', 'skhun', 'rapor', 'surat_pindah',
        'surat_sehat', 'kartu_vaksin',
        'sktm', 'dtks',
        'other'
    )),
    CONSTRAINT chk_file_size CHECK (file_size > 0 AND file_size <= 10485760)
);

-- Indexes
CREATE INDEX idx_document_tenant_company ON student_documents (tenant_id, company_id);
CREATE INDEX idx_document_student ON student_documents (student_id);
CREATE INDEX idx_document_type ON student_documents (student_id, document_type);
CREATE INDEX idx_document_rels ON student_documents USING GIN (_rels);
CREATE INDEX idx_document_data ON student_documents USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `document_type` | VARCHAR(30), CHECK | Kategori terbatas untuk konsistensi dan filterability |
| `file_url` | TEXT, NOT NULL | URL ke object storage (S3 presigned URL atau MinIO path) |
| `file_size` | BIGINT, max 10MB | Batas ukuran file per dokumen |
| `mime_type` | VARCHAR(100) | Untuk validasi dan display (image/jpeg, application/pdf) |
| `uploaded_by` | UUID, NOT NULL | Audit trail: siapa yang upload |
| `verified_by` | UUID, nullable | Admin yang memverifikasi keaslian dokumen |

### Vernon Relationships

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Selalu perlu tahu pemilik dokumen |

### _rels / _data Structure

```json
{
  "_rels": { "student_id": "018f..." },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" }
  }
}
```

### API Endpoints

```
GET    /api/v1/students/{id}/documents              — List dokumen siswa
GET    /api/v1/students/{id}/documents?type=ijazah   — Filter by type
POST   /api/v1/student-documents                     — Upload dokumen (multipart)
DELETE /api/v1/student-documents/{id}                — Soft delete dokumen
POST   /api/v1/student-documents/{id}/verify         — Verifikasi dokumen
GET    /api/v1/student-documents/{id}/download       — Download file (presigned URL)
```

### Upload Flow

```
1. Client POST multipart: file + student_id + document_type + title
2. Server validates: file size <= 10MB, mime_type allowed
3. Server uploads file to object storage → gets file_url
4. Server creates student_documents record with metadata
5. Returns 201 with document metadata (NOT the file itself)
```

## Consequences

### Positive

- **Metadata only**: Database ringan — file di object storage.
- **Immutable files**: Tidak ada overwrite — setiap upload = record baru.
- **Verifikasi workflow**: Admin bisa verify dokumen dan beri catatan.
- **Kategori standar**: Document type terbatas memudahkan checklist "dokumen apa yang belum diupload".

### Negative / Trade-offs

- **Object storage dependency**: Memerlukan S3/MinIO yang belum di-define di ADR terpisah.
- **URL expiry**: Presigned URL punya expiry — perlu mekanisme re-generate.
- **Tidak ada versioning**: Jika dokumen perlu di-update, upload baru dan soft-delete lama.
- **Belum ada bulk upload**: MVP satu file per request.

## Alternatives Considered

### 1. Simpan file sebagai BLOB di database
- Ditolak: membengkakkan database, backup lambat, tidak scalable.

### 2. JSONB array di student table
- Ditolak: tidak bisa query per tipe dokumen, tidak bisa verifikasi individual.

### 3. Generic file management (bukan student-specific)
- Deferred: bisa di-abstract nanti jika domain lain (guru, karyawan) butuh document management. MVP tetap student-specific.

## Test Cases

### Unit Tests — Descriptor & Validation

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` | — | `"student_documents"` |
| U02 | Validate rejects invalid `document_type` | `"passport"` | Error: invalid document_type |
| U03 | Validate rejects empty `title` | `""` | Error: title required |
| U04 | Validate rejects empty `file_url` | `""` | Error: file_url required |
| U05 | Validate rejects file_size > 10MB | `10485761` | Error: file too large |
| U06 | Validate rejects file_size <= 0 | `0` | Error: invalid file_size |
| U07 | Validate accepts valid document | All fields valid | No error |

### Integration Tests — CRUD

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Upload document | POST multipart with valid file | 201, metadata created, file in storage |
| I02 | List student documents | GET /students/{id}/documents | 200, all documents |
| I03 | Filter by type | GET /students/{id}/documents?type=akta_lahir | 200, filtered |
| I04 | Soft delete document | DELETE /student-documents/{id} | 200, `deleted_at` set |
| I05 | Download document | GET /student-documents/{id}/download | 200/302, presigned URL |

### Integration Tests — Verification

| # | Test Case | Action | Expected |
|---|---|---|---|
| I06 | Verify document | POST /student-documents/{id}/verify | 200, `verified_by` + `verified_at` set |
| I07 | Verify with note | POST verify with `verification_note` | 200, note saved |

### Integration Tests — Constraints

| # | Test Case | Action | Expected |
|---|---|---|---|
| I08 | Document type CHECK | INSERT with `document_type = 'passport'` | DB error |
| I09 | File size CHECK | INSERT with `file_size = 20000000` | DB error |
| I10 | Multiple docs same type allowed | Upload 2 `foto` documents | 201, both created |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | StudentUpdated syncs to documents | Update student name | `_data.student.full_name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I12 | Cannot access other tenant's documents | GET with wrong tenant | 404 |
| I13 | Cannot download other tenant's file | GET download with wrong tenant | 404/403 |
