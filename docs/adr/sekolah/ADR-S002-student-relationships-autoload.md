# ADR-S002: Student Relationships & Autoload Strategy

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Domain `student` (ADR-S001) menggunakan Vernon Pattern dengan `_rels`/`_data` JSONB. Perlu didefinisikan secara eksplisit:

1. **Relasi apa saja** yang dimiliki student ke domain lain.
2. **Mana yang autoload** (otomatis dimuat saat GET student) dan mana yang on-demand.
3. **Arah autoload** — mengikuti aturan Vernon: sisi "few" yang autoload, sisi "many" tidak.
4. **Circular autoload prevention** — Registry akan panic jika ada circular dependency.

Student adalah entitas paling terhubung di SekolahPro. Tanpa strategi autoload yang jelas, bisa terjadi:
- Over-fetching: semua relasi dimuat padahal tidak dibutuhkan.
- Circular reference: student → class_room → students → class_room (infinite loop).
- Inconsistent `_data`: relasi yang berubah tidak ter-sync ke snapshot.

## Decision

### Relationship Map

| Relasi | Domain Target | Tipe | Autoload | Alasan |
|---|---|---|---|---|
| `class_room` | `class_rooms` | belongs_to | **Ya** | Dashboard selalu menampilkan kelas siswa |
| `academic_year` | `academic_years` | belongs_to | **Ya** | Tahun ajaran aktif selalu relevan |
| `guardians` | `student_guardians` | has_many | **Tidak** | Bisa 1-3 records, dimuat on-demand di tab Keluarga |
| `address` | `student_addresses` | has_one | **Tidak** | Data detail, dimuat on-demand di profil lengkap |
| `previous_school` | `previous_schools` | has_one | **Tidak** | Hanya relevan untuk siswa pindahan |
| `health_records` | `student_health` | has_many | **Tidak** | Data berkala, dimuat on-demand di tab Kesehatan |
| `academic_records` | `student_academics` | has_many | **Tidak** | Data per semester, dimuat on-demand di tab Akademik |

### _rels Structure (stored in database)

```json
{
  "class_room_id": "018f...",
  "academic_year_id": "018f..."
}
```

### _data Structure (auto-populated by SyncEngine)

```json
{
  "class_room": {
    "id": "018f...",
    "name": "VII-A",
    "grade_level": "7"
  },
  "academic_year": {
    "id": "018f...",
    "name": "2025/2026",
    "is_active": true
  }
}
```

### Autoload Rules

1. **Hanya belongs_to yang autoload** — sesuai aturan Vernon (sisi "few" autoload, sisi "many" tidak).
2. **Maksimal 2 autoload** untuk student — `class_room` dan `academic_year`. Ini menjaga payload response tetap ringan.
3. **has_many dimuat via endpoint terpisah** — `/api/v1/students/{id}/guardians`, `/api/v1/students/{id}/health`, dll.
4. **Tidak ada circular autoload** — `class_rooms` TIDAK autoload balik ke `students` (has_many, arah sebaliknya).

### Circular Prevention

```
student → class_room (belongs_to, autoload: true)    OK
class_room → students (has_many, autoload: false)     OK — no circle

student → academic_year (belongs_to, autoload: true)  OK
academic_year → students (has_many, autoload: false)   OK — no circle
```

### SyncEngine Events

| Event | Trigger | Action |
|---|---|---|
| `ClassRoomUpdated` | Nama/level kelas berubah | Update `_data.class_room` di semua students dengan `_rels.class_room_id` matching |
| `AcademicYearUpdated` | Nama/status tahun ajaran berubah | Update `_data.academic_year` di semua students dengan `_rels.academic_year_id` matching |
| `StudentPromoted` | Kenaikan kelas | Update `_rels.class_room_id` + trigger re-sync `_data.class_room` |

## Consequences

### Positive

- **Response ringan**: GET student hanya membawa 2 relasi autoload — cukup untuk listing dan dashboard.
- **No circular risk**: has_many selalu non-autoload, eliminating circular dependency.
- **Scalable**: Menambah relasi baru cukup tambah entry di `DefaultRels()` tanpa mengubah schema.
- **Predictable sync**: Hanya 2 SyncEngine handler untuk student, mudah di-debug.

### Negative / Trade-offs

- **Extra API call**: Melihat data guardian/kesehatan membutuhkan request tambahan ke sub-endpoint.
- **Eventual consistency**: Jika nama kelas berubah, `_data.class_room` di student bisa stale selama window sync.
- **Kenaikan kelas massal**: Event `StudentPromoted` untuk ratusan siswa sekaligus membutuhkan batching.

## Alternatives Considered

### 1. Autoload semua relasi
- Ditolak: payload terlalu besar (guardian + alamat + kesehatan), response time lambat.

### 2. Tidak ada autoload (semua on-demand)
- Ditolak: dashboard dan listing SELALU butuh kelas + tahun ajaran, jadi setiap list page akan trigger N+1 queries.

### 3. GraphQL untuk flexible loading
- Ditolak: over-engineering untuk tahap awal. REST + sub-endpoints sudah cukup. Bisa dipertimbangkan di masa depan jika kebutuhan query flexibility meningkat.

## Test Cases

### Unit Tests — Autoload & Descriptor

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `DefaultRels()` returns exactly 2 autoloaded rels | — | `class_room` (autoload: true), `academic_year` (autoload: true) |
| U02 | `class_room` rel config correct | — | Table: `class_rooms`, Type: BelongsTo, Fields: `[id, name, grade_level]` |
| U03 | `academic_year` rel config correct | — | Table: `academic_years`, Type: BelongsTo, Fields: `[id, name, is_active]` |
| U04 | No circular autoload detected | Register student + class_room descriptors | Registry startup succeeds without panic |
| U05 | Circular autoload triggers panic | Register A→B (autoload) + B→A (autoload) | Registry panics with circular dependency error |

### Integration Tests — _rels / _data Population

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create student with class_room_id populates `_data` | POST student with valid `class_room_id` | `_data.class_room` contains `{id, name, grade_level}` |
| I02 | Create student with academic_year_id populates `_data` | POST student with valid `academic_year_id` | `_data.academic_year` contains `{id, name, is_active}` |
| I03 | `_rels` stores foreign key IDs | POST student with class_room_id + academic_year_id | `_rels` = `{class_room_id: "...", academic_year_id: "..."}` |
| I04 | GET student includes `_data` without JOIN | GET student by ID | Response contains `_data.class_room` + `_data.academic_year`, query uses zero JOINs |
| I05 | List students uses `_data` for display | GET students list | Each item has `_data.class_room.name` and `_data.academic_year.name` |

### Integration Tests — SyncEngine Events

| # | Test Case | Action | Expected |
|---|---|---|---|
| I06 | ClassRoomUpdated syncs to students | Update class_room name from "VII-A" to "VII-B" | All students with matching `_rels.class_room_id` have `_data.class_room.name` = "VII-B" |
| I07 | AcademicYearUpdated syncs to students | Update academic_year `is_active` from true to false | All students with matching `_rels.academic_year_id` have `_data.academic_year.is_active` = false |
| I08 | StudentPromoted updates class_room | Promote student to new class_room | `_rels.class_room_id` updated + `_data.class_room` re-synced with new class data |
| I09 | Sync does not affect unrelated students | Update class_room "VII-A" | Students in "VII-B" remain unchanged |
| I10 | `_sync_status` transitions during sync | Trigger ClassRoomUpdated | `_sync_status` goes `synced` → `pending` → `synced` |

### Integration Tests — Sub-endpoints (on-demand)

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Guardians loaded via sub-endpoint | `GET /students/{id}/guardians` | 200, returns guardian list (not from `_data`) |
| I12 | Health loaded via sub-endpoint | `GET /students/{id}/health` | 200, returns health records (not from `_data`) |
| I13 | Academics loaded via sub-endpoint | `GET /students/{id}/academics` | 200, returns academic records (not from `_data`) |
