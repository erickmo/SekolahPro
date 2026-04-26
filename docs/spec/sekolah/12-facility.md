# 12 — Manajemen Fasilitas

Modul Fasilitas mencakup empat sub-domain yang mengelola aset dan infrastruktur sekolah: perpustakaan (S038), laboratorium (S039), manajemen aset dan inventaris (S040), serta peminjaman ruangan dan fasilitas (S041). Modul ini berorientasi pada kepatuhan terhadap Permendagri No. 19/2016 (untuk sekolah negeri) dan integrasi dengan jadwal pelajaran (S021) untuk mencegah konflik penggunaan fasilitas.

---

## ADR References

| ADR | Judul | Pattern |
|-----|-------|---------|
| ADR-S038 | Library Management (Perpustakaan) | Vernon |
| ADR-S039 | Laboratory Management (Laboratorium) | Vernon |
| ADR-S040 | Asset & Inventory Management | Vernon |
| ADR-S041 | Room & Facility Booking | Vernon |

---

## Domain Entities

### Library Management (S038)

**`library_books`** — Master judul buku:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `isbn` | VARCHAR(20) | Unik per tenant+company (nullable — tidak semua buku punya ISBN) |
| `title` | VARCHAR(500) | Full-text search index (tsvector) |
| `author` | VARCHAR(300) | Pengarang |
| `category` | VARCHAR(30) | `textbook`, `reference`, `fiction`, `non_fiction`, `science`, `social`, `religion`, `kitab_kuning`, `al_quran`, `magazine`, `journal`, `other` |
| `language` | VARCHAR(20) | `indonesian`, `english`, `arabic`, `javanese`, `other` |
| `ddc_code` | VARCHAR(20) | Dewey Decimal Classification (nullable) |
| `shelf_location` | VARCHAR(50) | Lokasi rak (nullable) |
| `total_copies` | INT | Denormalisasi jumlah salinan |
| `available_copies` | INT | Denormalisasi salinan tersedia |

**`library_copies`** — Salinan buku (exemplar):

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `book_id` | UUID | FK → library_books |
| `copy_number` | INT | Nomor salinan per judul (unik per buku) |
| `barcode` | VARCHAR(50) | Barcode scanner (nullable) |
| `condition` | VARCHAR(20) | `new`, `good`, `fair`, `damaged`, `lost` |
| `status` | VARCHAR(20) | `available`, `borrowed`, `reserved`, `maintenance`, `lost`, `retired` |
| `acquisition_date` | DATE | Tanggal perolehan |

**`library_borrows`** — Transaksi peminjaman:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `copy_id` | UUID | FK → library_copies |
| `book_id` | UUID | FK → library_books (denormalisasi) |
| `borrower_type` | VARCHAR(10) | `student` atau `teacher` (polymorphic) |
| `borrower_id` | UUID | ID siswa atau guru |
| `borrow_date` | DATE | Tanggal pinjam |
| `due_date` | DATE | Batas pengembalian |
| `return_date` | DATE | Tanggal kembali (nullable) |
| `is_overdue` | BOOLEAN | Apakah terlambat |
| `overdue_days` | INT | Jumlah hari terlambat |
| `fine_amount` | BIGINT | Denda dalam Rupiah |
| `fine_paid` | BOOLEAN | Denda sudah dibayar |
| `status` | VARCHAR(20) | `borrowed`, `returned`, `overdue`, `lost` |

**Konfigurasi perpustakaan (per tenant):**

| Parameter | Default | Keterangan |
|-----------|---------|-----------|
| `loan_period` (siswa) | 14 hari | Durasi pinjam |
| `loan_period` (guru) | 30 hari | Durasi pinjam guru |
| `max_borrows` (siswa) | 3 buku | Maksimal pinjam bersamaan |
| `max_borrows` (guru) | 5 buku | |
| `fine_per_day` | Rp 500 | Denda per hari |
| `max_fine` | Rp 50.000 | Maksimal denda per buku |

### Laboratory Management (S039)

**`laboratories`** — Master lab:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `code` | VARCHAR(20) | Kode lab |
| `lab_type` | VARCHAR(20) | `lab_ipa`, `lab_komputer`, `lab_bahasa`, `lab_multimedia` |
| `specialization` | VARCHAR(20) | `fisika`, `kimia`, `biologi`, `ipa_terpadu` (nullable — untuk lab IPA) |
| `capacity` | INT | Kapasitas meja praktikum (1–100, default 30) |
| `head_technician_id` | UUID | Laboran penanggung jawab (nullable) |
| `has_safety_shower`, `has_eye_wash`, `has_fire_extinguisher`, `has_first_aid_kit`, `has_fume_hood` | BOOLEAN | Fasilitas K3 |
| `last_safety_inspection` | DATE | Tanggal inspeksi keselamatan terakhir |

**`lab_equipment`** — Inventaris alat lab:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `laboratory_id` | UUID | FK → laboratories |
| `category` | VARCHAR(30) | `alat_ukur`, `alat_percobaan`, `bahan_kimia`, `glassware`, `komputer`, `peripheral`, `audio_visual`, `furniture`, `safety_equipment`, `consumable`, `other` |
| `condition` | VARCHAR(20) | `baik`, `rusak_ringan`, `rusak_berat`, `hilang` |
| `quantity` | INT | Jumlah (≥ 0) |
| `unit` | VARCHAR(20) | `unit`, `set`, `buah`, `lembar`, `botol`, `pack`, `roll`, `liter`, `kg` |
| `purchase_price` | BIGINT | Harga beli (Rupiah) |
| `requires_calibration` | BOOLEAN | Butuh kalibrasi berkala |
| `next_calibration_date` | DATE | Jadwal kalibrasi berikutnya |

**`lab_usage_logs`** — Log penggunaan per sesi:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `laboratory_id` | UUID | FK |
| `teacher_id` | UUID | Guru yang menggunakan |
| `class_room_id` | UUID | Kelas yang menggunakan |
| `schedule_entry_id` | UUID | Link ke jadwal S021 (nullable) |
| `equipment_used` | JSONB | Array `{equipment_id, name, qty_used, condition_after}` |
| `safety_checklist` | JSONB | Checklist K3 per sesi (flexible per tipe lab) |
| `has_incident` | BOOLEAN | Flag sesi yang ada insiden |

### Asset & Inventory Management (S040)

**`assets`** — Master aset sekolah:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `asset_code` | VARCHAR(50) | Format Permendagri: `GG.BB.KK.SK.RRRR` (unik) |
| `asset_category` | VARCHAR(30) | `tanah`, `peralatan_mesin`, `gedung_bangunan`, `jalan_irigasi_jaringan`, `aset_tetap_lainnya`, `konstruksi_dalam_pengerjaan` |
| `ownership_type` | VARCHAR(20) | `bmd`, `yayasan`, `sekolah`, `hibah`, `pinjam_pakai` |
| `acquisition_method` | VARCHAR(20) | `pembelian`, `hibah`, `dana_bos`, `apbd`, dll |
| `condition` | VARCHAR(20) | `baik`, `kurang_baik`, `rusak_berat` |
| `lifecycle_status` | VARCHAR(20) | `active`, `maintenance`, `disposed`, `transferred`, `lost` |
| `useful_life_years` | INT | Umur manfaat (nullable — tanah tidak disusutkan) |
| `accumulated_depreciation` | BIGINT | Total penyusutan |
| `book_value` | BIGINT | `purchase_price - accumulated_depreciation` |
| `room_id` | UUID | FK → class_rooms atau ruangan lain (nullable) |
| `fiscal_year` | INT | Tahun anggaran pengadaan |

**Format kode barang Permendagri:**

```
GG.BB.KK.SK.RRRR
01 = Tanah, 02 = Peralatan & Mesin, 03 = Gedung & Bangunan
04 = Jalan/Irigasi/Jaringan, 05 = Aset Tetap Lainnya, 06 = Konstruksi

Contoh: 02.06.01.04.0001 = Komputer Desktop PC Register #1
```

**`asset_maintenances`** — Log pemeliharaan dan mutasi:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `maintenance_type` | VARCHAR(20) | `perbaikan`, `perawatan_rutin`, `penggantian_komponen`, `kalibrasi`, `mutasi`, `opname`, `other` |
| `condition_before` / `condition_after` | VARCHAR(20) | Kondisi sebelum dan sesudah |
| `cost` | BIGINT | Biaya dalam Rupiah |
| `cost_source` | VARCHAR(20) | `bos`, `apbd`, `yayasan`, `komite`, `sumbangan`, `other` |

**Metode penyusutan (straight-line):**

```
annual_depreciation = (purchase_price - salvage_value) / useful_life_years
book_value = purchase_price - accumulated_depreciation
```

Umur manfaat default: Komputer = 4 tahun, Gedung = 20–50 tahun, Kendaraan = 5–8 tahun.

### Room & Facility Booking (S041)

**`facilities`** — Master fasilitas yang bisa di-booking:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `code` | VARCHAR(20) | Kode unik |
| `facility_type` | VARCHAR(20) | `aula`, `lapangan`, `musholla`, `meeting_room`, `lab`, `ruang_serbaguna`, `studio`, `other` |
| `requires_approval` | BOOLEAN | Default `true` — butuh persetujuan admin/wakasek |
| `max_booking_days` | INT | Maksimal hari per booking (default 1) |
| `min_advance_hours` | INT | Minimal booking H-N (default 24 jam) |
| `max_advance_days` | INT | Maksimal booking ke depan (default 30 hari) |
| `allow_external` | BOOLEAN | Izinkan pihak luar booking (default false) |
| `allow_recurring` | BOOLEAN | Izinkan booking berulang (default true) |
| `amenities` | JSONB | Fasilitas tersedia (proyektor, sound system, AC, dll) |
| `laboratory_id` | UUID | Link ke S039 jika fasilitas adalah lab (nullable) |

**`facility_bookings`** — Transaksi peminjaman:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `facility_id` | UUID | FK → facilities |
| `requester_type` | VARCHAR(20) | `teacher`, `staff`, `extracurricular`, `committee`, `external` |
| `booking_date` | DATE | Tanggal booking |
| `start_time` / `end_time` | TIME | Waktu penggunaan (start < end) |
| `event_name` | VARCHAR(300) | Nama kegiatan |
| `booking_status` | VARCHAR(20) | `pending` → `approved` → `confirmed` → `in_use` → `completed` / `cancelled` / `rejected` |
| `is_recurring` | BOOLEAN | Booking berulang |
| `recurrence_pattern` | VARCHAR(20) | `daily`, `weekly`, `biweekly`, `monthly` |
| `recurrence_end_date` | DATE | Akhir periode berulang |
| `parent_booking_id` | UUID | Self-referensi untuk child recurring bookings |

---

## Business Rules

### Perpustakaan

1. **Satu salinan tidak bisa dipinjam lebih dari satu peminjam secara bersamaan.** Constraint: copy status harus `available` sebelum dipinjam.
2. **Batas pinjam:** siswa 3 buku, guru 5 buku. Kelebihan ditolak (422).
3. **Denda dihitung otomatis:** `overdue_days × fine_per_day`, dengan cap `max_fine` per buku. Konfigurasi per sekolah.
4. **`total_copies` dan `available_copies` adalah denormalisasi** — diupdate setiap copy ditambah, dipinjam, atau dikembalikan.
5. **Polymorphic borrower:** siswa dan guru menggunakan tabel pinjam yang sama dengan `borrower_type` + `borrower_id`.
6. **Kategori `kitab_kuning` dan `al_quran`** tersedia untuk koleksi perpustakaan pesantren.
7. **Full-text search** pada `title` menggunakan indeks GIN tsvector.

### Laboratorium

8. **Kondisi alat setelah penggunaan** dicatat di JSONB `equipment_used.condition_after`. Jika berbeda dari kondisi sebelumnya, service layer memperbarui kondisi di `lab_equipment`.
9. **Lab IPA SMP umumnya `ipa_terpadu`**, SMA memiliki lab terpisah untuk fisika, kimia, dan biologi.
10. **Fasilitas K3 wajib ada di lab kimia** sesuai Permendikbud — `has_fume_hood`, `has_safety_shower`, `has_eye_wash` harus `true` untuk operasi yang aman.
11. **Kalibrasi** alat ukur dilacak via `next_calibration_date` — alert ketika mendekati jadwal.
12. **Log insiden** (`has_incident = true`) mendapat indeks terpisah untuk monitoring keselamatan.

### Aset & Inventaris

13. **Kode barang mengikuti format Permendagri** `GG.BB.KK.SK.RRRR` — unik per tenant+company.
14. **Tanah tidak disusutkan** (`useful_life_years = null`; `book_value = purchase_price`).
15. **Penyusutan dihitung secara batch bulanan** (cron job), bukan real-time. `book_value` bisa stale antar batch.
16. **Opname tahunan** adalah stock-taking fisik — `last_opname_date` dan `last_opname_condition` dicatat per aset.
17. **Aset dari dana BOS** (`acquisition_method = 'dana_bos'`) harus dapat difilter untuk pelaporan juknis BOS.
18. **Mutasi aset** membuat record `asset_maintenance` dengan `maintenance_type = 'mutasi'` sebagai audit trail perpindahan lokasi.

### Room Booking — Conflict Detection

19. **Tiga lapisan conflict detection:**
    - **Database:** partial unique index mencegah booking pada `(facility_id, booking_date, start_time)` yang sama.
    - **Service layer:** sebelum INSERT, cek overlap waktu (`start_time < new_end AND end_time > new_start`).
    - **Timetable (S021):** untuk fasilitas bertipe `lab`, cek jadwal reguler di `schedule_entries` — jadwal reguler memblokir slot.
20. **Adjacent times tidak konflik.** Booking 09:00–12:00 dan 12:00–14:00 dapat berlangsung bersamaan.
21. **Recurring booking** generate child bookings hingga `recurrence_end_date`. Setiap child di-cek konflik secara terpisah — conflict di tanggal tertentu di-skip.
22. **Cancel parent recurring** otomatis membatalkan semua future children.
23. **`allow_external = false` (default)** — fasilitas tidak bisa di-booking oleh pihak luar tanpa konfigurasi eksplisit.

---

## Siklus Hidup Aset

```
Pengadaan (registrasi aset baru)
  → Active (penggunaan normal)
  → Maintenance (sedang diperbaiki)
  → Active (setelah perbaikan selesai)
  → Transferred (mutasi ke lokasi lain)
  → Disposed (penghapusan: lelang, pemusnahan, hibah)
  → Lost (hilang — dicatat dalam opname)
```

## Siklus Hidup Booking Fasilitas

```
pending
  → approved     (Wakasek/admin setujui)
  → confirmed    (dikonfirmasi ke pemohon)
  → in_use       (hari H, sedang berlangsung)
  → completed    (selesai + catat actual_time dan actual_attendees)
  → rejected     (ditolak dengan alasan)
  → cancelled    (dibatalkan oleh pemohon atau admin)
```

---

## Key Decisions & Rationale

1. **Lab Equipment tidak digabung dengan Asset (S040)** karena lab equipment punya field spesifik (kalibrasi, kategori lab, safety) yang tidak ada di aset umum. Cross-reference via `asset_code` dimungkinkan.
2. **Lab booking menggunakan S041** untuk penggunaan di luar jadwal reguler — jadwal reguler sudah di S021. S041 mengisi gap "Lab Kimia perlu dipakai untuk lomba saat hari biasa tidak ada jadwal."
3. **`safety_checklist` sebagai JSONB** — flexible per tipe lab (checklist kimia berbeda dari checklist komputer) tanpa migration.
4. **Perpustakaan polymorphic borrower** — siswa dan guru dipinjam di tabel yang sama, menghindari duplikasi logic dan tabel.
5. **Denda perpustakaan tidak terintegrasi dengan S009** di MVP — collection denda masih manual.
6. **Penyusutan hanya straight-line** — cukup untuk kebutuhan laporan keuangan sekolah. Metode lain (declining balance) dapat ditambahkan sebagai enhancement.
7. **Calendar view booking menggabungkan** bookings + timetable blocks dari S021 — satu endpoint memberikan gambaran lengkap availability fasilitas.

---

## Integration Points

### Internal

| Domain | Arah | Deskripsi |
|--------|------|-----------|
| ADR-S021 Timetable | → S041 | Jadwal reguler memblokir slot booking fasilitas |
| ADR-S021 Timetable | → S039 | `schedule_entry_id` menghubungkan log lab dengan jadwal |
| ADR-S011 ClassRoom | → S040 | Aset dipetakan per ruangan via `room_id` |
| S015 Extracurricular | → S041 | Ekskul mingguan bisa di-booking sebagai recurring |
| S050 Budget/RKAS | ← S041 | Booking fasilitas eksternal bisa menghasilkan pemasukan |
| S044 Notification | ← S041 | Notifikasi approval/penolakan booking ke pemohon |
| S055 Dapodik | ← S040 | Inventaris aset dilaporkan ke Kemendikbud |

### Eksternal

| Sistem | Keterangan |
|--------|-----------|
| **Dapodik** | Kondisi sarana prasarana (`baik`, `kurang_baik`, `rusak_berat`) sesuai standar Permendagri dilaporkan via S055 |
| **Barcode Scanner Hardware** | Perpustakaan (S038): optional integration via `barcode` field di `library_copies` |
| **Google Maps API** | Tidak diintegrasikan di MVP — koordinat GPS disimpan, rendering map di frontend |
