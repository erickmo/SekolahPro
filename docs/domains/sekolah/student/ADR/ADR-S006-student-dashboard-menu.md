# ADR-S006: Student Dashboard Menu Structure

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Setiap siswa di SekolahPro dikelola melalui **dashboard individual** — halaman yang menampilkan data lengkap siswa dan menyediakan menu navigasi ke berbagai aspek pengelolaan. Admin, guru, dan staf akan menghabiskan sebagian besar waktu di dashboard siswa ini.

Perlu didefinisikan:

1. **Menu apa saja** yang tersedia di dashboard siswa.
2. **Urutan dan pengelompokan** menu berdasarkan frekuensi penggunaan.
3. **Mapping ke API endpoints** dan domain yang sudah di-define di ADR sebelumnya.
4. **MVP scope** — menu mana yang dibangun dulu.

### Prinsip Desain

- **Data-first**: Dashboard menampilkan ringkasan data penting di overview, detail di sub-page.
- **Role-aware**: Menu yang tampil bisa berbeda berdasarkan role user (admin vs guru vs staf).
- **Consistent pattern**: Setiap menu mengikuti pola yang sama — list/detail/create/edit.

## Decision

### Dashboard Layout

```
┌─────────────────────────────────────────────────────┐
│  Student Dashboard: Ahmad Rizki (NIS: 12345)        │
│  Kelas: VII-A | Status: Aktif | TA: 2025/2026      │
├─────────────┬───────────────────────────────────────┤
│             │                                       │
│  [Menu]     │  [Content Area]                       │
│             │                                       │
│  Profil     │  Menampilkan konten sesuai             │
│  Keluarga   │  menu yang dipilih.                    │
│  Akademik   │                                       │
│  Kesehatan  │  Default: Overview (ringkasan          │
│  Keuangan   │  semua aspek siswa)                    │
│  Dokumen    │                                       │
│  Prestasi   │                                       │
│  Tata Tertib│                                       │
│             │                                       │
└─────────────┴───────────────────────────────────────┘
```

### Menu Specification

| # | Menu | Domain/ADR | API Endpoint | Fitur | MVP |
|---|---|---|---|---|---|
| 1 | **Overview** | student (ADR-S001) | `GET /students/{id}` | Ringkasan: foto, biodata singkat, kelas, absensi bulan ini, tagihan tertunggak | Ya |
| 2 | **Profil** | student (ADR-S001) | `GET/PUT /students/{id}` | View/edit biodata lengkap: identitas, alamat, foto | Ya |
| 3 | **Keluarga** | student_guardian (ADR-S003) | `GET /students/{id}/guardians` | CRUD orang tua/wali, tandai primary contact | Ya |
| 4 | **Akademik** | student_academic (ADR-S004) | `GET /students/{id}/academics` | Riwayat per semester: nilai, absensi, ranking, status kenaikan | Ya |
| 5 | **Kesehatan** | student_health (ADR-S005) | `GET /students/{id}/health` | Riwayat pemeriksaan: fisik, alergi, asuransi | Ya |
| 6 | **Keuangan** | *student_finance (future)* | `GET /students/{id}/finance` | Status SPP, tagihan, riwayat pembayaran | Tidak |
| 7 | **Dokumen** | *student_document (future)* | `GET /students/{id}/documents` | Upload/download: akta, KK, ijazah, rapor, foto | Tidak |
| 8 | **Prestasi** | *student_achievement (future)* | `GET /students/{id}/achievements` | Daftar prestasi akademik & non-akademik | Tidak |
| 9 | **Tata Tertib** | *student_discipline (future)* | `GET /students/{id}/disciplines` | Catatan pelanggaran, poin, sanksi | Tidak |

### Overview Card Layout

Overview menampilkan ringkasan dari setiap domain dalam bentuk cards:

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ 📋 Biodata    │  │ 👨‍👩‍👧 Keluarga  │  │ 📊 Akademik   │
│              │  │              │  │              │
│ VII-A        │  │ Ayah: Budi   │  │ Rata-rata:   │
│ TA 2025/2026 │  │ Ibu: Sari    │  │ 82.5         │
│ Status: Aktif│  │ HP: 08xx     │  │ Rank: 5/32   │
└──────────────┘  └──────────────┘  └──────────────┘

┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ 🏥 Kesehatan  │  │ 💰 Keuangan  │  │ 📄 Dokumen   │
│              │  │              │  │              │
│ TB: 155 cm   │  │ SPP: Lunas   │  │ Akta: ✓     │
│ BB: 45 kg    │  │ Tunggakan: 0 │  │ KK: ✓       │
│ Alergi: -    │  │              │  │ Foto: ✓     │
└──────────────┘  └──────────────┘  └──────────────┘
```

### Frontend Route Structure

```
/students                          → Student List Page
/students/{id}                     → Student Dashboard (Overview)
/students/{id}/profile             → Profil (edit biodata)
/students/{id}/guardians           → Keluarga
/students/{id}/academics           → Akademik
/students/{id}/health              → Kesehatan
/students/{id}/finance             → Keuangan (future)
/students/{id}/documents           → Dokumen (future)
/students/{id}/achievements        → Prestasi (future)
/students/{id}/disciplines         → Tata Tertib (future)
```

### Data Loading Strategy

| Halaman | Loading | Alasan |
|---|---|---|
| Overview | Parallel fetch: student + latest guardians + latest academic + latest health | Semua data ringkasan dimuat sekaligus untuk tampilan cepat |
| Sub-page | Lazy load saat navigasi | Hanya muat data yang dibutuhkan tab aktif |
| Student List | Paginated, data dari `_data` (Vernon) | Zero JOIN — kelas dan tahun ajaran sudah di `_data` |

## Consequences

### Positive

- **MVP jelas**: 5 menu pertama (Overview, Profil, Keluarga, Akademik, Kesehatan) sudah cukup untuk operasional dasar.
- **Extensible**: Menu baru (Keuangan, Dokumen, Prestasi, Tata Tertib) bisa ditambah tanpa mengubah arsitektur.
- **Consistent URL pattern**: `/students/{id}/{section}` — predictable dan SEO-friendly.
- **Parallel loading**: Overview memuat semua ringkasan secara paralel untuk responsivitas.

### Negative / Trade-offs

- **Overview N+1 risk**: Overview memerlukan fetch ke 4+ endpoints secara paralel — perlu dipastikan backend bisa handle concurrent requests.
- **Role-based visibility**: Belum di-define secara detail menu mana yang tersembunyi untuk role tertentu.
- **Future domains**: Keuangan, Dokumen, Prestasi, dan Tata Tertib belum punya ADR — perlu dibuat saat akan diimplementasi.

## Alternatives Considered

### 1. Single page dengan semua data
- Ditolak: terlalu berat — memuat semua data sekaligus lambat dan memboroskan bandwidth.

### 2. Tab-based (bukan menu sidebar)
- Ditolak: tab kurang scalable untuk 8+ section. Sidebar menu lebih mudah di-extend.

### 3. Modal-based editing
- Ditolak: data siswa terlalu banyak field untuk modal. Full-page form lebih user-friendly untuk data entry.

## Test Cases

### Unit Tests — Route Configuration

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `/students` route exists | Navigate to `/students` | Student List Page renders |
| U02 | `/students/{id}` route exists | Navigate to `/students/valid-uuid` | Student Dashboard (Overview) renders |
| U03 | `/students/{id}/profile` route exists | Navigate | Profil page renders |
| U04 | `/students/{id}/guardians` route exists | Navigate | Keluarga page renders |
| U05 | `/students/{id}/academics` route exists | Navigate | Akademik page renders |
| U06 | `/students/{id}/health` route exists | Navigate | Kesehatan page renders |
| U07 | Invalid student ID shows 404 | Navigate to `/students/nonexistent-uuid` | 404 or "Student not found" page |

### Integration Tests — Overview Page

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Overview loads student header | Navigate to `/students/{id}` | Header shows name, NIS, kelas, status, tahun ajaran |
| I02 | Overview loads all summary cards | Navigate to `/students/{id}` | Biodata, Keluarga, Akademik, Kesehatan cards visible |
| I03 | Overview parallel fetch performance | Navigate to `/students/{id}` | All cards loaded within acceptable time (parallel API calls) |
| I04 | Overview handles missing guardian data | Student with no guardians | Keluarga card shows empty state |
| I05 | Overview handles missing academic data | Student with no academic records | Akademik card shows empty state |
| I06 | Overview handles missing health data | Student with no health records | Kesehatan card shows empty state |

### Integration Tests — Menu Navigation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I07 | Sidebar menu renders all MVP items | Open student dashboard | Profil, Keluarga, Akademik, Kesehatan menus visible |
| I08 | Click Profil menu navigates | Click "Profil" in sidebar | URL changes to `/students/{id}/profile`, content updates |
| I09 | Click Keluarga menu navigates | Click "Keluarga" in sidebar | URL changes to `/students/{id}/guardians`, guardian list loads |
| I10 | Click Akademik menu navigates | Click "Akademik" in sidebar | URL changes to `/students/{id}/academics`, academic records load |
| I11 | Click Kesehatan menu navigates | Click "Kesehatan" in sidebar | URL changes to `/students/{id}/health`, health records load |
| I12 | Active menu highlighted | Navigate to `/students/{id}/guardians` | "Keluarga" menu item is visually active/highlighted |
| I13 | Future menus (Keuangan, Dokumen, Prestasi, Tata Tertib) | Check menu list | Either hidden or shown as disabled/coming soon |

### Integration Tests — Student List Page

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | Student list loads with pagination | `GET /students` | Paginated list with name, NIS, kelas, status from `_data` |
| I15 | Click student row opens dashboard | Click student row | Navigates to `/students/{id}` (Overview) |
| I16 | Search filters by name | Type name in search box | List filters to matching students |
| I17 | Filter by status | Select "active" filter | Only active students shown |

### Integration Tests — Data Loading Strategy

| # | Test Case | Action | Expected |
|---|---|---|---|
| I18 | Sub-page lazy loads on navigation | Navigate from Overview to Keluarga | API call for guardians fires only on navigation, not on initial load |
| I19 | Student list uses zero-JOIN query | Load student list | Network tab shows single API call, response includes `_data` (no separate calls for kelas/tahun ajaran) |
| I20 | Browser back/forward works | Navigate Overview → Keluarga → Back | Returns to Overview with cached data |
