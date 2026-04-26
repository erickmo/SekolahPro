# Student Core — Data Model, Relasi, Wali, & Dashboard

Dokumen ini menjelaskan fondasi entitas siswa di SekolahPro: model data inti, strategi autoload relasi, data orang tua/wali, alamat, sekolah asal, dan struktur dashboard siswa. Bersama-sama, komponen ini membentuk "kartu identitas digital" siswa yang digunakan di seluruh sistem.

---

## ADR References

| ADR | Judul |
|-----|-------|
| S001 | Student Core Data Model |
| S002 | Student Relationships & Autoload Strategy |
| S003 | Student Guardian (Orang Tua/Wali) |
| S006 | Student Dashboard Menu Structure |
| S007 | Student Address & Previous School |

---

## Domain Entities

### 1. `students` — Data Inti Siswa

Entitas utama dengan Vernon Pattern. Satu row = satu siswa aktif atau historis di satu sekolah.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `id` | UUID | PK, uuid_v7 | Primary key internal |
| `tenant_id` | UUID | NOT NULL | Isolasi multi-tenant |
| `company_id` | UUID | NOT NULL | Isolasi per sekolah |
| `nis` | VARCHAR(20) | NOT NULL, UNIQUE per company | Nomor Induk Siswa (lokal) |
| `nisn` | VARCHAR(10) | nullable, UNIQUE global | Nomor Induk Siswa Nasional (Kemendikbud) |
| `full_name` | VARCHAR(255) | NOT NULL | Nama lengkap |
| `nickname` | VARCHAR(100) | nullable | Nama panggilan |
| `gender` | VARCHAR(1) | CHECK: L/P | L = Laki-laki, P = Perempuan |
| `birth_place` | VARCHAR(100) | NOT NULL | Kota kelahiran |
| `birth_date` | DATE | NOT NULL | Tanggal lahir |
| `religion` | VARCHAR(20) | CHECK: 6 agama | islam / kristen / katolik / hindu / buddha / konghucu |
| `blood_type` | VARCHAR(2) | nullable | Golongan darah |
| `photo_url` | TEXT | nullable | URL ke object storage |
| `status` | VARCHAR(20) | CHECK | active / graduated / transferred / expelled / on_leave |
| `entry_date` | DATE | NOT NULL | Tanggal masuk sekolah |
| `entry_type` | VARCHAR(20) | CHECK | new / transfer / returning |
| `exit_date` | DATE | nullable | Diisi saat keluar |
| `exit_reason` | TEXT | nullable | Alasan keluar |
| `_rels` | JSONB | NOT NULL | FK IDs: `class_room_id`, `academic_year_id` |
| `_data` | JSONB | NOT NULL | Snapshot: `class_room`, `academic_year` |
| `deleted_at` | TIMESTAMPTZ | nullable | Soft delete |

**Catatan Desain:**
- NIS bersifat unik per sekolah (bukan global). NISN unik secara global menggunakan partial unique index (`WHERE nisn IS NOT NULL`) karena NISN boleh NULL.
- `photo_url` menyimpan URL ke object storage (S3/MinIO), bukan binary data di database.
- Status lifecycle tidak di-enforce di database level — validasi transisi dilakukan di application layer.

### 2. `student_guardians` — Orang Tua/Wali

Data orang tua atau wali siswa. Satu guardian bisa ter-link ke lebih dari satu siswa (kakak-adik).

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `id` | UUID | PK | |
| `full_name` | VARCHAR(255) | NOT NULL | Nama lengkap wali |
| `nik` | VARCHAR(16) | nullable | NIK KTP 16 digit |
| `gender` | VARCHAR(1) | CHECK: L/P | |
| `birth_place`, `birth_date` | — | nullable | |
| `religion` | VARCHAR(20) | nullable | |
| `phone` | VARCHAR(20) | nullable | Nomor HP untuk komunikasi |
| `email` | VARCHAR(255) | nullable | |
| `occupation` | VARCHAR(100) | nullable | Pekerjaan |
| `income_range` | VARCHAR(20) | CHECK | lt_1m / 1m_3m / 3m_5m / 5m_10m / gt_10m |
| `education` | VARCHAR(30) | CHECK | tidak_sekolah / sd / smp / sma / d1-d3 / s1-s3 |
| `is_alive` | BOOLEAN | NOT NULL, default true | Status — mempengaruhi data Dapodik |
| Kolom alamat | — | nullable | address, rt, rw, village, district, city, province, postal_code |

### 3. `student_guardian_map` — Junction Table Siswa ↔ Wali

Relasi many-to-many antara siswa dan guardian.

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `student_id` | UUID | NOT NULL, FK | |
| `guardian_id` | UUID | NOT NULL, FK | |
| `relationship` | VARCHAR(20) | CHECK: father/mother/guardian | |
| `is_primary` | BOOLEAN | NOT NULL, default false | Kontak prioritas utama |

**UNIQUE** pada `(student_id, guardian_id)` — satu guardian tidak bisa di-link dua kali ke siswa yang sama.

### 4. `student_addresses` — Alamat Siswa

Satu alamat aktif per siswa. Alamat siswa bisa berbeda dari alamat orang tua (siswa kos/pondok).

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `student_id` | UUID | UNIQUE | Satu siswa satu alamat |
| `address` | TEXT | NOT NULL | Jalan/dusun/kampung |
| `rt`, `rw` | VARCHAR(3) | nullable | RT/RW — tidak semua daerah pakai |
| `village`, `district`, `city`, `province` | VARCHAR | city & province NOT NULL | Level alamat Indonesia |
| `postal_code` | VARCHAR(5) | nullable | |
| `latitude`, `longitude` | NUMERIC(10,7) | nullable | Koordinat untuk zonasi PPDB |
| `living_with` | VARCHAR(20) | CHECK | parents / guardian / boarding / alone / other |
| `distance_km` | NUMERIC(5,1) | nullable | Jarak ke sekolah dalam km |

### 5. `student_previous_schools` — Sekolah Asal

Data sekolah sebelumnya. Hanya relevan untuk siswa pindahan atau masuk lewat jalur transfer. Satu record per siswa (unique pada `student_id`).

| Kolom | Tipe | Constraint | Keterangan |
|-------|------|-----------|------------|
| `student_id` | UUID | UNIQUE | |
| `school_name` | VARCHAR(255) | NOT NULL | Nama sekolah asal |
| `npsn` | VARCHAR(8) | nullable | Nomor Pokok Sekolah Nasional |
| `school_address`, `school_city`, `school_province` | — | nullable | |
| `last_class` | VARCHAR(20) | nullable | Kelas terakhir di sekolah asal |
| `exit_year` | INT | nullable | Tahun keluar |
| `certificate_no` | VARCHAR(50) | nullable | Nomor ijazah/SKHUN |

---

## Business Rules

1. **NIS wajib ada sebelum siswa aktif.** NIS diberikan oleh sekolah saat siswa pertama kali mendaftar.
2. **NISN tidak selalu ada.** Siswa baru, siswa sekolah informal, atau siswa yang belum terdaftar di Dapodik boleh tidak punya NISN. Setelah NISN diperoleh, harus unik secara global (partial unique index).
3. **Agama terbatas 6 agama resmi Indonesia.** Sesuai standar Kemendikbud dan Dapodik.
4. **Gender hanya L atau P.** Standar Kemendikbud.
5. **Status lifecycle siswa ada 5:** active → graduated (lulus), transferred (pindah), expelled (dikeluarkan), on_leave (cuti). Transisi harus mengisi `exit_date` dan `exit_reason`.
6. **Satu siswa bisa punya 2-3 guardian** (ayah, ibu, dan/atau wali). Minimal satu guardian harus ditandai `is_primary = true`.
7. **Guardian bisa shared antar siswa.** Untuk kakak-adik di sekolah yang sama, satu data guardian dipakai bersama via junction table.
8. **Income guardian disimpan sebagai range, bukan angka eksak.** Untuk privasi dan keperluan kategorisasi beasiswa.
9. **Alamat siswa independen dari alamat guardian.** Siswa yang tinggal di asrama/pondok punya alamat berbeda.
10. **Koordinat alamat untuk zonasi PPDB.** Latitude/longitude diisi untuk mendukung kalkulasi jarak otomatis di seleksi PPDB.
11. **Sekolah asal hanya untuk siswa pindahan.** Tidak wajib untuk siswa baru yang masuk dari jenjang bawah.
12. **Soft delete selalu digunakan.** Data siswa tidak dihapus permanen — `deleted_at` diisi, siswa tersembunyi dari listing tapi data bisa dipulihkan.

---

## Autoload Strategy (Vernon Pattern)

Hanya relasi `belongs_to` (sisi "few") yang di-autoload. Relasi `has_many` dimuat on-demand via sub-endpoint.

### Autoload pada `students`

| Relasi | Tipe | Autoload | Payload `_data` |
|--------|------|----------|----------------|
| `class_room` | belongs_to | **Ya** | `{id, name, grade_level}` |
| `academic_year` | belongs_to | **Ya** | `{id, name, is_active}` |
| `guardians` | has_many | Tidak | Via `GET /students/{id}/guardians` |
| `address` | has_one | Tidak | Via `GET /students/{id}/address` |
| `previous_school` | has_one | Tidak | Via `GET /students/{id}/previous-school` |
| `health_records` | has_many | Tidak | Via `GET /students/{id}/health` |
| `academic_records` | has_many | Tidak | Via `GET /students/{id}/academics` |

### Kenapa Hanya 2 Autoload?

Dashboard siswa dan listing **selalu** membutuhkan nama kelas dan tahun ajaran. Memuat relasi lain (guardian, kesehatan) secara default akan membengkakkan payload untuk listing 100+ siswa tanpa manfaat.

### Struktur `_rels` dan `_data`

```json
// _rels: FK IDs yang disimpan
{
  "class_room_id": "018f-...",
  "academic_year_id": "018f-..."
}

// _data: snapshot yang di-sync otomatis
{
  "class_room": {
    "id": "018f-...",
    "name": "VII-A",
    "grade_level": "7"
  },
  "academic_year": {
    "id": "018f-...",
    "name": "2025/2026",
    "is_active": true
  }
}
```

### SyncEngine Events untuk Student

| Event | Trigger | Action |
|-------|---------|--------|
| `ClassRoomUpdated` | Nama/level kelas berubah | Update `_data.class_room` di semua students matching `_rels.class_room_id` |
| `AcademicYearUpdated` | Nama/status tahun ajaran berubah | Update `_data.academic_year` di semua students matching `_rels.academic_year_id` |
| `StudentPromoted` | Kenaikan kelas | Update `_rels.class_room_id` + re-sync `_data.class_room` |

---

## Dashboard Siswa — Struktur Menu

Setiap siswa memiliki **dashboard individual** dengan layout sidebar menu.

### Struktur Layout

```
┌─ Header: [Nama Siswa] — [NIS] — [Kelas] — [Status] — [TA] ──────────┐
│                                                                        │
│  Sidebar Menu          │  Content Area                                 │
│  ─────────────         │  ───────────────────────────────              │
│  Overview              │  Default: Overview Cards                      │
│  Profil                │                                               │
│  Keluarga              │  Content berganti sesuai menu dipilih         │
│  Akademik              │                                               │
│  Kesehatan             │                                               │
│  Keuangan (future)     │                                               │
│  Dokumen (future)      │                                               │
│  Prestasi (future)     │                                               │
│  Tata Tertib (future)  │                                               │
└────────────────────────┴───────────────────────────────────────────────┘
```

### Spesifikasi Menu

| # | Menu | Domain | Endpoint | MVP |
|---|------|--------|----------|-----|
| 1 | **Overview** | student (S001) | `GET /students/{id}` | Ya |
| 2 | **Profil** | student (S001, S007) | `GET/PUT /students/{id}` | Ya |
| 3 | **Keluarga** | guardian (S003) | `GET /students/{id}/guardians` | Ya |
| 4 | **Akademik** | academic (S004) | `GET /students/{id}/academics` | Ya |
| 5 | **Kesehatan** | health (S005) | `GET /students/{id}/health` | Ya |
| 6 | **Keuangan** | finance (S009) | `GET /students/{id}/invoices` | Future |
| 7 | **Dokumen** | document (S010) | `GET /students/{id}/documents` | Future |
| 8 | **Prestasi** | achievement (S013) | `GET /students/{id}/achievements` | Future |
| 9 | **Tata Tertib** | discipline (S012) | `GET /students/{id}/disciplines` | Future |

### Strategi Loading Data

| Halaman | Strategi | Alasan |
|---------|----------|--------|
| **Listing Siswa** | `_data` langsung (zero JOIN) | Kelas & tahun ajaran sudah di snapshot |
| **Overview Dashboard** | Parallel fetch ke 4+ endpoint | Semua summary cards dimuat bersamaan |
| **Sub-page (Keluarga, Akademik, dll.)** | Lazy load saat navigasi | Hanya muat data yang dibutuhkan |

---

## API Endpoints

```
# Student
GET    /api/v1/students                          List siswa (paginated, filter status/search)
POST   /api/v1/students                          Buat siswa baru
GET    /api/v1/students/{id}                     Detail siswa (termasuk _data)
PUT    /api/v1/students/{id}                     Update biodata
DELETE /api/v1/students/{id}                     Soft delete

# Guardian
GET    /api/v1/students/{id}/guardians           List wali siswa
POST   /api/v1/students/{id}/guardians           Link guardian ke siswa
DELETE /api/v1/students/{id}/guardians/{gid}     Unlink guardian
GET    /api/v1/student-guardians                 List semua guardian (admin)
POST   /api/v1/student-guardians                 Buat guardian baru
PUT    /api/v1/student-guardians/{id}            Update guardian

# Address
GET    /api/v1/students/{id}/address             Alamat siswa
POST   /api/v1/student-addresses                 Set alamat siswa
PUT    /api/v1/student-addresses/{id}            Update alamat

# Previous School
GET    /api/v1/students/{id}/previous-school     Sekolah asal
POST   /api/v1/student-previous-schools          Set sekolah asal
PUT    /api/v1/student-previous-schools/{id}     Update sekolah asal
```

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Vernon Pattern dengan 2 autoload | Dashboard dan listing selalu butuh kelas + tahun ajaran; has_many tidak di-autoload agar payload tetap ringan |
| NIS dan NISN sebagai identifier ganda | NIS lokal (diberikan sekolah) + NISN nasional (Dapodik); UUID internal untuk stabilitas FK |
| Guardian many-to-many via junction table | Orang tua dengan beberapa anak di sekolah yang sama hanya perlu diinput sekali |
| Income guardian sebagai range | Privasi data sensitif; kategorisasi cukup untuk eligibilitas beasiswa |
| Alamat siswa terpisah dari guardian | Siswa kos/pondok punya alamat berbeda dari orang tua |
| Koordinat pada alamat | Mendukung seleksi PPDB jalur zonasi (jarak rumah–sekolah) |
| Soft delete | Data siswa tidak hilang permanen; bisa dipulihkan jika terjadi kesalahan input |

---

## Integration Points

| Titik Integrasi | Arah | Keterangan |
|----------------|------|------------|
| PPDB (S016) | S016 → S001 | Saat daftar ulang, data applicant di-copy ke `students`, `student_guardians`, `student_addresses`, `student_previous_schools` |
| Class Placement (S014) | S014 → S001 | Kenaikan kelas update `students._data.class_room` |
| Academic Record (S004) | S001 → S004 | `student_id` sebagai FK di setiap semester record |
| Rapor (S018) | S001 → S018 | Biodata siswa dan nama wali masuk ke `student_snapshot` dan `guardian_snapshot` |
| Dapodik Sync | S001 | `nisn`, `religion`, `gender`, data guardian di-sync ke Dapodik |
| Koperasi | S001 | `students.id` digunakan sebagai member reference di modul Koperasi |

---

## Buku Induk Siswa Digital

Buku Induk Siswa adalah dokumen wajib setiap sekolah berdasarkan regulasi Kemdikbud.
SekolahPro menyediakan fitur Buku Induk Digital yang di-generate dari data existing.

### Format Buku Induk
Buku Induk menggabungkan data dari beberapa tabel:
- `students` + `student_addresses` + `student_previous_schools` (data identitas)
- `student_guardians` (data orang tua/wali)
- `student_class_placements` (riwayat kelas per tahun ajaran)
- `student_academics` (nilai akhir per semester — ringkasan)

### Nomor Induk
- `students.nis` digunakan sebagai nomor induk
- Berurutan per sekolah, format: [kode sekolah]-[tahun masuk]-[urutan]
- Tidak berubah selama siswa aktif di sekolah tersebut

### Export
- Format: Excel (`.xlsx`) dan PDF
- Endpoint: `GET /api/v1/students/buku-induk?class_id={id}&academic_year_id={id}`
- PDF: satu halaman per siswa, dengan tanda tangan digital kepala sekolah
- Excel: satu baris per siswa untuk satu tahun ajaran
