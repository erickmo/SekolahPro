# 10 — Manajemen Asrama (Dormitory)

Modul Asrama adalah fitur inti untuk pondok pesantren dan boarding school Indonesia. Berdasarkan ADR-009 (dual-mode institution), ketika `school_type = "islamic"` modul ini menjadi wajib, bukan opsional. Modul mencakup tiga sub-domain: manajemen kamar dan penempatan santri, aktivitas harian dan absensi asrama, serta disiplin dan kesehatan penghuni asrama.

---

## ADR References

| ADR | Judul | Pattern |
|-----|-------|---------|
| ADR-S033 | Dormitory Management (Kamar & Penghuni) | Vernon |
| ADR-S034 | Dormitory Activity & Attendance | Vernon |
| ADR-S035 | Dormitory Discipline & Health | Vernon |

---

## Domain Entities

### Dormitory Management (S033)

**`dormitory_buildings`** — Master gedung asrama:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `name` | VARCHAR(100) | Nama gedung (misal: "Gedung Al-Farabi") |
| `code` | VARCHAR(20) | Kode unik per tenant+company |
| `gender` | VARCHAR(10) | `male` atau `female` — pemisahan absolut |
| `total_floors` | INT | Jumlah lantai |
| `total_capacity` | INT | Kapasitas total (denormalisasi dari kamar) |
| `supervisor_id` | UUID | Musyrif/Musyrifah (FK → teachers, nullable) |

**`dormitory_rooms`** — Master kamar:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `building_id` | UUID | FK → dormitory_buildings |
| `room_number` | VARCHAR(20) | Nomor kamar, unik per gedung |
| `floor` | INT | Lantai |
| `capacity` | INT | Kapasitas kamar (≥ 1) |
| `current_occupancy` | INT | Jumlah penghuni aktif (denormalisasi) |
| `room_type` | VARCHAR(20) | `regular`, `vip`, `isolation`, `musyrif` |
| `has_bathroom` | BOOLEAN | Memiliki kamar mandi |

**`dormitory_assignments`** — Penempatan santri per tahun ajaran:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `student_id` | UUID | FK → students |
| `room_id` | UUID | FK → dormitory_rooms |
| `academic_year_id` | UUID | FK → academic_years |
| `bed_number` | INT | Nomor tempat tidur (nullable — opsional) |
| `assigned_date` | DATE | Tanggal penempatan |
| `end_date` | DATE | Tanggal berakhir penempatan (nullable) |
| `status` | VARCHAR(20) | `active`, `transferred`, `ended` |
| `previous_room_id` | UUID | Asal kamar sebelum transfer (nullable) |
| `transfer_reason` | TEXT | Alasan transfer |

### Dormitory Activity & Attendance (S034)

**`dormitory_activities`** — Master jadwal aktivitas asrama:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `code` | VARCHAR(20) | Kode unik aktivitas (misal: `SH-SUBUH`) |
| `activity_type` | VARCHAR(20) | `sholat`, `meal`, `study`, `roll_call`, `cleaning`, `sports`, `other` |
| `time_start` / `time_end` | TIME | Waktu aktivitas |
| `is_mandatory` | BOOLEAN | Wajib (sholat) atau opsional |
| `applies_to` | VARCHAR(10) | `all`, `male`, `female` |

**`dormitory_attendances`** — Absensi per aktivitas per santri:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `student_id` | UUID | FK → students |
| `activity_id` | UUID | FK → dormitory_activities |
| `academic_year_id` | UUID | FK |
| `attendance_date` | DATE | Tanggal absensi |
| `status` | VARCHAR(20) | `present`, `late`, `absent`, `permitted`, `sick` |
| `check_in_time` | TIME | Waktu aktual check-in (untuk tracking keterlambatan) |
| `recorded_by` | UUID | Musyrif yang mencatat |

**`dormitory_permissions`** — Izin pulang/keluar asrama:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `permission_type` | VARCHAR(20) | `weekend`, `holiday`, `sick`, `family_event`, `emergency`, `other` |
| `depart_date` | DATE | Tanggal keberangkatan |
| `expected_return_date` | DATE | Tanggal kembali yang direncanakan |
| `actual_return_date` | DATE | Tanggal kembali aktual (nullable) |
| `status` | VARCHAR(20) | `pending` → `approved` → `departed` → `returned`; atau `rejected`; `overdue` jika belum kembali |
| `picked_up_by` | VARCHAR(200) | Nama penjemput |

### Dormitory Discipline & Health (S035)

**`dormitory_violations`** — Pelanggaran asrama:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `category` | VARCHAR(30) | `sholat`, `kebersihan`, `jam_malam`, `keluar_tanpa_izin`, `kekerasan`, `pencurian`, `gadget`, `merokok`, `pacaran`, `other` |
| `severity` | VARCHAR(20) | `ringan`, `sedang`, `berat`, `sangat_berat` |
| `points` | INT | Poin pelanggaran |
| `sanction` | VARCHAR(30) | Teguran lisan/tertulis, hafalan tambahan, piket tambahan, panggilan ortu, skorsing, dikeluarkan |
| `parent_notified` | BOOLEAN | Orang tua sudah diberitahu |

**`dormitory_inspections`** — Inspeksi kamar:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `room_id` | UUID | FK → dormitory_rooms |
| `cleanliness_score` | INT | Skor kebersihan (0–100) |
| `tidiness_score` | INT | Skor kerapihan (0–100) |
| `completeness_score` | INT | Skor kelengkapan (0–100) |
| `overall_score` | INT | Rata-rata tertimbang (0–100) |
| `grade` | VARCHAR(2) | A, B, C, D, atau E |

**Mapping grade inspeksi:** A (80–100), B (60–79), C (40–59), D (20–39), E (0–19).

**`dormitory_health_records`** — Catatan kesehatan santri:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `record_type` | VARCHAR(20) | `illness`, `injury`, `allergy`, `infectious`, `mental_health`, `routine_check` |
| `severity` | VARCHAR(20) | `minor`, `moderate`, `serious`, `emergency` |
| `complaint` | TEXT | Keluhan santri |
| `diagnosis` | TEXT | Diagnosis (nullable) |
| `treatment` | TEXT | Penanganan |
| `referred_to` | VARCHAR(30) | Eskalasi: `uks`, `puskesmas`, `klinik`, `rumah_sakit`, `orang_tua` |

---

## Business Rules

### Penempatan Santri

1. **Pemisahan gender absolut.** Santri laki-laki hanya bisa ditempatkan di gedung `gender = 'male'`, perempuan di `gender = 'female'`. Validasi di application layer dan constraint database.
2. **Kapasitas tidak boleh melebihi limit.** Setiap penugasan memeriksa `current_occupancy < capacity` sebelum dibuat.
3. **Satu santri hanya boleh punya satu assignment aktif** per tahun ajaran. Constraint unik `(student_id, academic_year_id, status='active')`.
4. **Transfer kamar** membuat assignment lama `status = 'transferred'` dan assignment baru dengan `previous_room_id` terisi sebagai audit trail.
5. **`current_occupancy` adalah denormalisasi** yang harus dijaga konsisten via application logic setiap kali assignment dibuat, ditransfer, atau diakhiri.
6. **Rotasi tahunan** dilakukan dengan membuat assignment baru di tahun ajaran berikutnya — assignment lama di-end.

### Absensi Aktivitas

7. **Volume sangat tinggi:** ~10 aktivitas/hari × 500 santri = 5.000 records/hari. Partitioning tabel mungkin diperlukan untuk pesantren besar.
8. **Satu record per santri per aktivitas per tanggal.** Duplicate ditolak.
9. **Bulk input** adalah metode utama — musyrif menginput seluruh santri dalam satu request per aktivitas.
10. **Status `late` ditambahkan** dibandingkan absensi kelas (S008) karena keterlambatan sholat berjamaah adalah indikator penting kedisiplinan di pesantren.
11. **Overdue detection** memerlukan background job/cron untuk menandai santri yang belum kembali melewati `expected_return_date`.

### Disiplin Asrama

12. **Threshold eskalasi pelanggaran** (konfigurabel per pesantren):
    - Akumulasi 50 poin → skorsing_asrama (1 minggu)
    - Akumulasi 75 poin → dikeluarkan_asrama (tetap sekolah)
    - Akumulasi 100 poin → dikeluarkan_pesantren (perlu approval pimpinan)
13. **Disiplin asrama terpisah dari disiplin sekolah (S012).** Kategori, sanksi, dan threshold berbeda. Satu santri bisa punya poin di kedua domain secara independen.
14. **Pemberitahuan orang tua** dilacak via `parent_notified` — pelanggaran berat wajib diberitahukan.

### Kesehatan

15. **Kasus `infectious` mendapat indeks khusus** untuk early warning system penularan di lingkungan asrama yang padat.
16. **Severity `serious` dan `emergency` mendapat indeks khusus** untuk monitoring dan eskalasi cepat.
17. **Catatan kesehatan asrama adalah event-based** (kejadian harian), berbeda dari data kesehatan statis siswa (S005: alergi, golongan darah, dll).

---

## Alur Workflow

### Izin Pulang Santri

```
Orang tua/santri ajukan izin
  → status: pending
  → Musyrif/Musyrifah review
  → approved / rejected
  → Saat santri berangkat: departed (catat penjemput)
  → Saat santri kembali: returned (catat actual_return_date/time)
  → Jika melewati expected_return_date: overdue (alert musyrif)
```

### Inspeksi Kamar

```
Setup jadwal inspeksi (mingguan/bulanan)
  → Musyrif lakukan inspeksi
  → Input skor (kebersihan, kerapihan, kelengkapan)
  → Grade otomatis dihitung
  → Notifikasi penghuni kamar
  → Ranking kamar terbersih (kompetisi motivasi)
```

---

## Key Decisions & Rationale

1. **Tiga sub-domain terpisah** (kamar, aktivitas, disiplin+kesehatan) karena masing-masing punya volume dan lifecycle yang berbeda dan tidak saling bergantung dalam penulisan.
2. **Absensi asrama tidak digabung dengan S008** karena fundamental berbeda: S008 adalah 1 record/hari (kelas), absensi asrama bisa 10+ records/hari per santri.
3. **Disiplin asrama tidak digabung dengan S012** karena kategori pelanggaran, sanksi (hafalan tambahan, piket), dan threshold eskalasi sangat berbeda dari pelanggaran sekolah.
4. **Catatan kesehatan asrama tidak digabung dengan S005** karena S005 adalah data statis (riwayat penyakit, alergi), sementara S035 adalah event-based daily monitoring.
5. **Inspeksi menggunakan 3 sub-skor** (bukan 1 skor) untuk memberikan feedback spesifik per dimensi kebersihan — mendukung program kompetisi kamar terbersih.
6. **`bed_number` opsional** — mendukung pesantren yang mengelola per-bed maupun hanya per-kamar.

---

## Integration Points

### Internal

| Domain | Arah | Deskripsi |
|--------|------|-----------|
| ADR-S001 Students | → S033 | Data siswa untuk assignment |
| ADR-S003 Guardians | → S034 | Orang tua dihubungi untuk konfirmasi izin |
| ADR-S012 Discipline | Terpisah | Disiplin asrama independen dari disiplin sekolah |
| ADR-S005 Student Health | Terpisah | S005 = data statis; S035 = event harian |
| S044 Notification | ← S034/S035 | Notifikasi izin, pelanggaran, kesehatan ke orang tua |
| S042 Parent Portal | ← S033/S034/S035 | Orang tua lihat penempatan, absensi aktivitas, dan pelanggaran anak |
| ADR-012 Teachers | → S033 | Musyrif/Musyrifah sebagai supervisor gedung |

### Eksternal

| Sistem | Keterangan |
|--------|-----------|
| **Kemenag / EMIS** | Beberapa pesantren melaporkan data santri ke EMIS (Education Management Information System Kemenag) — format berbeda dari Dapodik |
| **SMS/WhatsApp Gateway** | Notifikasi darurat (kesehatan `emergency`, izin overdue) ke orang tua via S044 |
