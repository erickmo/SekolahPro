# ADR-S014: Student Class Placement / Mutasi Kelas

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Riwayat penempatan kelas siswa diperlukan untuk:

1. **Tracking kelas**: Mengetahui kelas siswa saat ini dan riwayat kelas sebelumnya.
2. **Kenaikan kelas**: Proses kenaikan kelas di akhir tahun ajaran.
3. **Penjurusan**: Pemilihan jurusan (IPA/IPS/Bahasa) di SMA kelas XI.
4. **Mutasi**: Perpindahan kelas di tengah semester (kasus khusus).
5. **Rapor**: Setiap rapor memerlukan data kelas yang benar untuk semester tersebut.
6. **Historis**: Riwayat lengkap kelas dari masuk hingga lulus.

Hubungan student ↔ class_room berubah setiap tahun ajaran. ADR-S001 menyimpan class_room saat ini di `_data`, tapi **tidak menyimpan riwayat**. Riwayat ini perlu tabel sendiri karena:

- Satu siswa ada di satu kelas per tahun ajaran (bisa berubah mid-year untuk kasus mutasi).
- Kenaikan kelas adalah **event** — tanggalnya penting untuk audit.
- Penjurusan (IPA/IPS) juga terjadi di tahap class placement.

### Mengapa Vernon Pattern?

- has_many dari student (1 placement per tahun ajaran).
- Read-heavy: dashboard, rapor, dan laporan memerlukan kelas per periode.
- Relasi ke student, class_room, academic_year.
- Business logic moderate: kenaikan kelas massal, penjurusan.

## Decision

Menggunakan **Vernon Pattern** untuk domain `student_class_placement`.

### Table Schema

```sql
CREATE TABLE student_class_placements (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    class_room_id   UUID NOT NULL,

    -- Placement detail
    placement_type  VARCHAR(20) NOT NULL,
    effective_date  DATE NOT NULL,
    end_date        DATE,
    is_current      BOOLEAN NOT NULL DEFAULT true,

    -- Penjurusan (SMA)
    major           VARCHAR(30),

    -- Notes
    note            TEXT,
    placed_by       UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_placement_student_year UNIQUE (student_id, academic_year_id) WHERE (is_current = true),
    CONSTRAINT chk_placement_type CHECK (placement_type IN ('initial', 'promotion', 'retention', 'transfer', 'major_selection')),
    CONSTRAINT chk_placement_major CHECK (major IS NULL OR major IN ('ipa', 'ips', 'bahasa', 'agama', 'umum'))
);

-- Indexes
CREATE INDEX idx_placement_tenant_company ON student_class_placements (tenant_id, company_id);
CREATE INDEX idx_placement_student ON student_class_placements (student_id);
CREATE INDEX idx_placement_class ON student_class_placements (class_room_id);
CREATE INDEX idx_placement_year ON student_class_placements (academic_year_id);
CREATE INDEX idx_placement_current ON student_class_placements (student_id) WHERE is_current = true;
CREATE INDEX idx_placement_rels ON student_class_placements USING GIN (_rels);
CREATE INDEX idx_placement_data ON student_class_placements USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `placement_type` | VARCHAR(20), CHECK | 5 tipe: initial (masuk baru), promotion (naik kelas), retention (tinggal kelas), transfer (mutasi), major_selection (penjurusan) |
| `effective_date` | DATE, NOT NULL | Tanggal mulai penempatan — penting untuk audit |
| `end_date` | DATE, nullable | NULL jika masih aktif — diisi saat pindah kelas atau lulus |
| `is_current` | BOOLEAN | Flag untuk query cepat "kelas saat ini" |
| `major` | VARCHAR(30), nullable | Jurusan — hanya diisi untuk SMA/MA kelas XI+ |
| `placed_by` | UUID, NOT NULL | Admin yang melakukan penempatan — audit trail |

### Vernon Relationships

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Pemilik placement |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran penempatan |
| `class_room` | belongs_to | **Ya** | Kelas yang ditempati |

### _rels / _data Structure

```json
{
  "_rels": {
    "student_id": "018f...",
    "academic_year_id": "018f...",
    "class_room_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" },
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "class_room": { "id": "018f...", "name": "VIII-A", "grade_level": "8" }
  }
}
```

### API Endpoints

```
GET    /api/v1/students/{id}/class-placements        — Riwayat kelas siswa
GET    /api/v1/students/{id}/current-class            — Kelas saat ini
POST   /api/v1/student-class-placements               — Penempatan individual
POST   /api/v1/student-class-placements/promote       — Kenaikan kelas massal
POST   /api/v1/student-class-placements/transfer      — Mutasi kelas
PUT    /api/v1/student-class-placements/{id}          — Update placement
```

### Promotion Flow (Kenaikan Kelas Massal)

```
1. Admin pilih: academic_year (tahun baru) + source class (kelas lama)
2. System menampilkan daftar siswa kelas tersebut
3. Admin menandai: naik (promoted) / tinggal (retained) per siswa
4. Admin pilih target class untuk yang naik
5. System:
   a. Set end_date + is_current=false pada placement lama
   b. Create placement baru dengan placement_type='promotion' atau 'retention'
   c. Update student._data.class_room dengan kelas baru
   d. Emit event ClassPlacementCreated
```

### SyncEngine Integration

Ketika placement baru dibuat dengan `is_current = true`:
- Update `students._data.class_room` dengan kelas terbaru
- Update `student_academics._data.class_room` untuk semester aktif
- Update `student_attendances._data.class_room` untuk records tahun ajaran terkait

## Consequences

### Positive

- **Riwayat lengkap**: Setiap perpindahan kelas terdokumentasi dengan tanggal dan tipe.
- **Kenaikan massal**: Endpoint bulk promotion mendukung operasi akhir tahun ajaran.
- **Penjurusan**: Major field mendukung SMA/MA dengan peminatan.
- **Current class query**: `is_current` flag memungkinkan query O(1) untuk kelas saat ini.
- **Sync ke S001**: Placement baru otomatis update `_data.class_room` di student.

### Negative / Trade-offs

- **Partial unique index**: `UNIQUE WHERE is_current = true` tidak didukung di semua database — PostgreSQL specific.
- **Promotion complexity**: Kenaikan massal memerlukan batch transaction yang bisa gagal partial.
- **Mid-year transfer**: Mutasi tengah semester memerlukan update ke S004 (academic), S008 (attendance), dan S011 (grades).
- **Major immutability**: Belum ditentukan apakah jurusan bisa berubah setelah dipilih.

## Alternatives Considered

### 1. Class history di students._data sebagai JSONB array
- Ditolak: tidak bisa query "siswa mana saja yang di kelas VII-A tahun 2025/2026".

### 2. Foreign key class_room_id langsung di students table
- Ditolak: hanya menyimpan kelas saat ini, tidak ada riwayat. Dan students._data.class_room sudah serve this need.

### 3. Gabung dengan S004 (academic record)
- Ditolak: placement adalah event administrasi, academic record adalah data akademik. Satu placement bisa punya 2 academic records (ganjil + genap).

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` | — | `"student_class_placements"` |
| U02 | Validate rejects invalid `placement_type` | `"move"` | Error |
| U03 | Validate rejects invalid `major` | `"teknik"` | Error |
| U04 | Validate accepts null `major` | `major = null` | No error |
| U05 | Validate rejects missing `effective_date` | null | Error |

### Integration Tests — CRUD

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create initial placement | POST with `placement_type = 'initial'` | 201 |
| I02 | Get placement history | GET /students/{id}/class-placements | 200, ordered by effective_date |
| I03 | Get current class | GET /students/{id}/current-class | 200, returns `is_current = true` record |
| I04 | Placement type CHECK | INSERT with `placement_type = 'move'` | DB error |
| I05 | Major CHECK | INSERT with `major = 'teknik'` | DB error |

### Integration Tests — Promotion

| # | Test Case | Action | Expected |
|---|---|---|---|
| I06 | Bulk promote class | POST /promote with class of 30 students | 201, 30 new placements, 30 old set `is_current=false` |
| I07 | Retention (tinggal kelas) | Promote with 1 student marked retention | Retained student gets new placement with same grade_level |
| I08 | Unique current per student per year | Create 2 current placements same year | Constraint violation |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I09 | New placement syncs to student._data | Create placement | `students._data.class_room` updated |
| I10 | ClassRoomUpdated syncs to placements | Update class_room name | `_data.class_room.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Cannot access other tenant's placements | GET with wrong tenant | 404 |
