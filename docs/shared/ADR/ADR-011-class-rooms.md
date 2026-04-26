# ADR-011: Class Rooms (Kelas)

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Kelas (class room) adalah **unit organisasi utama** dalam sekolah. Hampir semua operasional sekolah diorganisir per kelas:

- Absensi diambil per kelas
- Nilai diinput per kelas per mapel
- SPP bisa di-generate per kelas
- Rapor dicetak per kelas
- Kenaikan kelas dari kelas X ke kelas Y

Domain yang memiliki `class_room_id` sebagai foreign key:

- S001 Student (autoloaded `_data.class_room`)
- S004 Academic Record
- S008 Daily Attendance
- S011 Subject Grades
- S014 Class Placement
- S018 Rapor

Kelas di Indonesia memiliki karakteristik:
- **Grade level**: Tingkat kelas (1-6 SD, 7-9 SMP, 10-12 SMA).
- **Parallel class**: Kelas paralel (VII-A, VII-B, VII-C).
- **Wali kelas**: Setiap kelas punya satu guru wali kelas.
- **Per tahun ajaran**: Kelas yang sama (VII-A) bisa punya siswa berbeda setiap tahun.
- **Kapasitas**: Batas siswa per kelas (biasanya 28-36).

### Mengapa Vernon Pattern?

- Referenced by 6+ domain sebagai belongs_to.
- Read-heavy: setiap operasional diorganisir per kelas.
- Relasi ke academic_year dan teacher (wali kelas).
- SyncEngine: perubahan nama kelas harus propagate ke semua `_data`.

## Decision

Menggunakan **Vernon Pattern** untuk domain `class_room`. Kelas di-scope per **tahun ajaran** — setiap tahun ajaran punya set kelas sendiri.

### Table Schema

```sql
CREATE TABLE class_rooms (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    academic_year_id UUID NOT NULL,
    homeroom_teacher_id UUID,

    -- Identitas
    name            VARCHAR(20) NOT NULL,
    grade_level     VARCHAR(5) NOT NULL,
    parallel_id     VARCHAR(5),

    -- Konfigurasi
    capacity        INT NOT NULL DEFAULT 36,
    current_count   INT NOT NULL DEFAULT 0,

    -- Penjurusan (SMA)
    major           VARCHAR(30),

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

    CONSTRAINT uq_class_room_name_year UNIQUE (tenant_id, company_id, academic_year_id, name),
    CONSTRAINT chk_class_grade_level CHECK (grade_level IN (
        '1', '2', '3', '4', '5', '6',
        '7', '8', '9',
        '10', '11', '12'
    )),
    CONSTRAINT chk_class_capacity CHECK (capacity > 0),
    CONSTRAINT chk_class_count CHECK (current_count >= 0 AND current_count <= capacity),
    CONSTRAINT chk_class_major CHECK (major IS NULL OR major IN ('ipa', 'ips', 'bahasa', 'agama', 'umum'))
);

-- Indexes
CREATE INDEX idx_class_room_tenant_company ON class_rooms (tenant_id, company_id);
CREATE INDEX idx_class_room_year ON class_rooms (academic_year_id);
CREATE INDEX idx_class_room_grade ON class_rooms (grade_level);
CREATE INDEX idx_class_room_teacher ON class_rooms (homeroom_teacher_id);
CREATE INDEX idx_class_room_rels ON class_rooms USING GIN (_rels);
CREATE INDEX idx_class_room_data ON class_rooms USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `name` | VARCHAR(20), e.g. "VII-A" | Nama kelas yang ditampilkan — unik per tahun ajaran |
| `grade_level` | VARCHAR(5), CHECK | Tingkat: 1-12 (SD-SMA). Sebagai string karena bisa "10" |
| `parallel_id` | VARCHAR(5), nullable | Identitas paralel: "A", "B", "C" — nullable jika hanya 1 kelas per grade |
| `homeroom_teacher_id` | UUID, nullable | Wali kelas — nullable karena mungkin belum di-assign |
| `capacity` | INT, default 36 | Batas siswa per kelas — standar Kemendikbud 28-36 |
| `current_count` | INT, denormalisasi | Jumlah siswa saat ini — diupdate saat placement CRUD. Menghindari COUNT query |
| `major` | VARCHAR(30), nullable | Penjurusan — hanya untuk SMA kelas 11-12 |
| `academic_year_id` | UUID, NOT NULL | Kelas di-scope per tahun ajaran |

### Pesantren Terminology (ADR-009)

Untuk `school_type = 'islamic'`, terminologi berubah tapi data structure sama:

| General | Islamic | Mapping |
|---------|---------|---------|
| Kelas | Halaqah | `name`: "Halaqah VII-A" |
| Grade Level | Marhalah | `grade_level`: sama (1-12) |
| Wali Kelas | Musyrif | `homeroom_teacher_id` |

Terminologi di-resolve di presentation layer berdasarkan `school_type`, bukan di database.

### Vernon Relationships

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `academic_year` | belongs_to | **Ya** | Tahun ajaran kelas |
| `homeroom_teacher` | belongs_to | **Ya** | Wali kelas — nama ditampilkan di rapor |

### _rels / _data Structure

```json
{
  "_rels": {
    "academic_year_id": "018f...",
    "homeroom_teacher_id": "018f..."
  },
  "_data": {
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "homeroom_teacher": { "id": "018f...", "full_name": "Bu Siti", "nip": "198501012010012001" }
  }
}
```

### SyncEngine Impact

Ketika `class_rooms` diupdate (nama kelas, wali kelas), SyncEngine propagate ke:
- `students._data.class_room`
- `student_academics._data.class_room`
- `student_attendances._data.class_room`
- `student_grades._data.class_room`
- `student_class_placements._data.class_room`
- `rapor_records._data.class_room`

### Year Transition: Carry-Forward Kelas

Di awal tahun ajaran baru, admin perlu membuat set kelas baru. Fitur **carry-forward**:

```
1. Admin pilih: "Copy kelas dari tahun ajaran 2025/2026 ke 2026/2027"
2. System:
   a. Duplikasi semua class_rooms dari tahun lama ke tahun baru
   b. Reset: current_count = 0, homeroom_teacher_id = NULL
   c. Admin adjust: tambah/hapus kelas, assign wali kelas baru
3. Siswa belum ada di kelas baru — S014 (Class Placement) yang assign siswa
```

### API Endpoints

```
GET    /api/v1/class-rooms                           — List kelas (default: tahun ajaran aktif)
GET    /api/v1/class-rooms?year_id={id}              — Kelas per tahun ajaran
GET    /api/v1/class-rooms/{id}                      — Detail kelas + wali kelas
GET    /api/v1/class-rooms/{id}/students             — Siswa di kelas ini
POST   /api/v1/class-rooms                           — Buat kelas
PUT    /api/v1/class-rooms/{id}                      — Update kelas
POST   /api/v1/class-rooms/carry-forward             — Copy kelas ke tahun baru
```

## Consequences

### Positive

- **Per tahun ajaran**: Kelas VII-A tahun 2025 ≠ VII-A tahun 2026 — historical accuracy.
- **Capacity enforced**: Database-level CHECK mencegah kelas overload.
- **Count denormalisasi**: `current_count` menghindari COUNT query — dashboard cepat.
- **Carry-forward**: Admin tidak perlu buat kelas dari nol setiap tahun.
- **Pesantren compatible**: Data structure sama, terminologi di-resolve di presentation.

### Negative / Trade-offs

- **Duplicate per tahun**: Setiap tahun ajaran = set kelas baru. 10 tahun × 20 kelas = 200 rows — masih sangat kecil.
- **Teacher dependency**: `homeroom_teacher_id` reference tabel `teachers` yang belum di-define (ADR-012).
- **Count consistency**: `current_count` harus selalu sinkron dengan actual count di S014. Jika mismatch, perlu reconciliation job.
- **SyncEngine heavy**: Update nama kelas memicu sync ke 6+ domain.

## Alternatives Considered

### 1. Kelas tanpa scope tahun ajaran (reuse tiap tahun)
- Ditolak: kelas VII-A tahun 2025 punya siswa dan wali kelas berbeda dari 2026. Tanpa scope tahun ajaran, historical query tidak mungkin.

### 2. Kelas sebagai JSONB di academic_years
- Ditolak: kelas punya relationship sendiri (wali kelas, siswa) dan di-reference oleh banyak domain.

### 3. Grade level sebagai tabel terpisah
- Deferred: untuk MVP, `grade_level` sebagai field cukup. Bisa dijadikan tabel jika perlu config per grade (kurikulum, jam pelajaran).

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` | — | `"class_rooms"` |
| U02 | `DefaultRels()` returns 2 rels | — | `academic_year` + `homeroom_teacher` |
| U03 | Validate rejects invalid `grade_level` | `"13"` | Error |
| U04 | Validate rejects `capacity <= 0` | `0` | Error |
| U05 | Validate rejects invalid `major` | `"teknik"` | Error |
| U06 | Validate accepts valid class room | All fields valid | No error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create class room | POST with valid data | 201, `_data.academic_year` populated |
| I02 | Unique name per year per company | Create 2 "VII-A" same year | 409/422 |
| I03 | Same name different year OK | Create "VII-A" in 2025 and 2026 | 201, both created |
| I04 | Get students in class | GET /class-rooms/{id}/students | 200, list from S014 placements |
| I05 | Grade level CHECK | INSERT with `grade_level = '13'` | DB error |
| I06 | Capacity CHECK | INSERT with `capacity = 0` | DB error |
| I07 | Count CHECK | UPDATE `current_count` > `capacity` | DB error |
| I08 | Carry-forward | POST /carry-forward from year A to B | 201, all classes copied with reset counts |
| I09 | SyncEngine: name change propagates | Update class name | All referencing `_data.class_room` updated |
| I10 | Tenant isolation | Access other tenant's class | 404 |
