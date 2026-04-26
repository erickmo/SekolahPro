# 14 — Administrasi & Tata Kelola Sekolah

Modul Administrasi & Tata Kelola mencakup empat sub-domain yang mendukung operasional formal sekolah: Approval Workflow Engine (S047), Profil Sekolah & Akreditasi (S048), Komite Sekolah (S049), dan Anggaran Sekolah / RKAS (S050). Modul ini memastikan bahwa proses persetujuan, identitas sekolah, representasi orang tua dalam tata kelola, dan perencanaan keuangan terkelola secara digital dan terdokumentasi.

---

## ADR References

| ADR | Judul | Pattern |
|-----|-------|---------|
| ADR-S047 | Approval Workflow Engine | Vernon |
| ADR-S048 | School Profile & Accreditation | Vernon |
| ADR-S049 | School Committee (Komite Sekolah) | Vernon |
| ADR-S050 | School Budget Management (RKAS) | Vernon |

---

## Domain Entities

### Approval Workflow Engine (S047)

Engine generic yang melayani semua domain yang memerlukan approval bertingkat.

**`approval_chains`** — Konfigurasi chain per entity_type:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `entity_type` | VARCHAR(50) | `lesson_plan`, `leave_request`, `correspondence`, `budget_plan`, `facility_booking`, `student_transfer`, `procurement`, `custom` |
| `steps` | JSONB | Array step: `[{order, role, sla_hours, can_delegate}]` |
| `condition_rules` | JSONB | Rule engine untuk chain berbeda berdasarkan kondisi (field, operator, value) |
| `is_active` | BOOLEAN | Soft-disable tanpa delete |

**`approval_requests`** — Request persetujuan (1 per entity submission):

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `entity_type` + `entity_id` | VARCHAR + UUID | Polymorphic reference ke entity asal |
| `entity_title` | VARCHAR(500) | Denormalisasi judul — inbox dapat ditampilkan tanpa JOIN |
| `chain_id` | UUID | FK → approval_chains |
| `requestor_id` | UUID | FK → users |
| `status` | VARCHAR(25) | Lihat state machine di bawah |
| `current_step` | INT | Pointer ke step yang sedang aktif |
| `total_steps` | INT | Ditetapkan saat request dibuat |
| `sla_deadline` | TIMESTAMPTZ | Deadline keseluruhan request |
| `is_overdue` | BOOLEAN | Flag SLA terlewat |

**`approval_steps`** — Langkah per-level dalam request:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `request_id` | UUID | FK |
| `step_order` | INT | Urutan (1, 2, 3...) |
| `approver_role` | VARCHAR(50) | Role yang harus approve |
| `approver_id` | UUID | Di-resolve saat step menjadi active (nullable saat waiting) |
| `delegated_to` | UUID | UUID user jika di-delegate (nullable) |
| `action` | VARCHAR(25) | `approve`, `reject`, `revision_requested`, `delegate`, `skip` |
| `status` | VARCHAR(25) | `waiting` → `active` → `approved`/`rejected`/`revision_requested`/`delegated`/`skipped` |
| `sla_hours` | INT | SLA per step (dari chain config) |

### School Profile & Accreditation (S048)

**`school_profiles`** — Profil resmi sekolah (1 row per company_id):

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `npsn` | VARCHAR(20) | 8-digit kode unik Kemendikbud (nullable saat setup awal) |
| `nss` | VARCHAR(20) | Nomor Statistik Sekolah (legacy) |
| `school_level` | VARCHAR(20) | `tk`, `sd`, `smp`, `sma`, `smk`, `slb`, `mi`, `mts`, `ma`, `mak` |
| `school_status` | VARCHAR(20) | `negeri` atau `swasta` |
| `principal_id` | UUID | FK → teachers/users |
| `vision`, `mission`, `goals` | TEXT | Visi, misi, dan tujuan sekolah |
| `foundation_name` | VARCHAR(255) | Nama yayasan (untuk swasta) |
| `is_pesantren` | BOOLEAN | Flag dual-mode dari ADR-009 |
| `pesantren_nspp` | VARCHAR(30) | Nomor Statistik Pondok Pesantren (Kemenag) |
| `pesantren_type` | VARCHAR(30) | `salafiyah`, `modern`, `kombinasi` |
| `current_accreditation_grade` | VARCHAR(20) | Denormalisasi dari riwayat akreditasi |
| `total_students`, `total_teachers`, `total_staff`, `total_class_rooms`, `total_rombel` | INT | Statistik denormalisasi (event-driven update) |
| `total_santri_mukim`, `total_santri_kalong`, `total_asatidz` | INT | Statistik pesantren |
| `letter_code` | VARCHAR(20) | Kode untuk penomoran surat (dipakai S046) |
| `letter_number_format` | VARCHAR(100) | Format default nomor surat keluar |

**`school_accreditations`** — Riwayat akreditasi:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `accreditation_body` | VARCHAR(20) | `ban_s`, `ban_sm`, `ban_pnf`, `ban_pt`, `lam` |
| `grade` | VARCHAR(20) | `A` (Unggul), `B` (Baik Sekali), `C` (Baik), `not_accredited` |
| `score` | DECIMAL(5,2) | Nilai 0–100 (nullable untuk data historis) |
| `accreditation_date` | DATE | Tanggal akreditasi |
| `valid_until` | DATE | Berlaku hingga (5 tahun dari tanggal akreditasi) |
| `is_current` | BOOLEAN | Flag akreditasi yang sedang berlaku (hanya 1 per sekolah) |

### School Committee / Komite Sekolah (S049)

**`committee_terms`** — Periode kepengurusan:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `term_number` | INT | 1 atau 2 (maks 2 periode per Permendikbud 75/2016) |
| `start_date` / `end_date` | DATE | Masa jabatan 3 tahun |
| `decree_number` | VARCHAR(100) | Nomor SK pengangkatan (nullable) |
| `is_active` | BOOLEAN | Hanya 1 term aktif per sekolah |

**`committee_members`** — Anggota/pengurus:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `term_id` | UUID | FK → committee_terms |
| `position` | VARCHAR(30) | `ketua`, `wakil_ketua`, `sekretaris`, `bendahara`, `anggota` |
| `representation` | VARCHAR(30) | `orang_tua`, `tokoh_masyarakat`, `pakar_pendidikan`, `dunia_usaha`, `alumni`, `tokoh_agama` |
| `guardian_id` | UUID | FK → student_guardians (nullable — tokoh masyarakat bukan orang tua) |

**`committee_meetings`** — Rapat komite:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `meeting_type` | VARCHAR(20) | `regular`, `extraordinary`, `annual`, `rkas_review` |
| `agenda` | JSONB | List agenda `[{order, title, duration_minutes}]` |
| `resolutions` | JSONB | Keputusan rapat `[{number, resolution, vote}]` |
| `minutes` | TEXT | Notulen (nullable — diisi setelah rapat) |
| `status` | VARCHAR(20) | `scheduled` → `in_progress` → `completed` / `cancelled` |

**`committee_meeting_attendees`** — Kehadiran rapat:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `attendance_status` | VARCHAR(20) | `present`, `absent`, `excused`, `late` |
| `signature_url` | VARCHAR(500) | Bukti kehadiran digital |

### School Budget / RKAS (S050)

**`budget_sources`** — Master sumber dana:

| `source_type` | Keterangan |
|---------------|-----------|
| `bos_reguler`, `bos_kinerja`, `bos_afirmasi` | Dana BOS dari Kemendikbud |
| `apbd` | Dana Pemerintah Daerah |
| `yayasan` | Dana yayasan |
| `komite` | Dana komite sekolah |
| `dana_mandiri` | Dana sekolah sendiri |
| `infaq`, `wakaf`, `donatur` | Sumber dana pesantren |

**`budget_plans`** — RKAS per tahun ajaran:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `academic_year_id` | UUID | FK |
| `fiscal_year` | INT | Tahun anggaran |
| `total_planned` | BIGINT | Total anggaran direncanakan |
| `total_realized` | BIGINT | Total realisasi (denormalisasi) |
| `status` | VARCHAR(20) | `draft` → `submitted` → `approved` → `revised` → `closed` |

**`budget_items`** — Alokasi per kegiatan:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `budget_plan_id` | UUID | FK |
| `budget_source_id` | UUID | FK → budget_sources |
| `bos_component` | VARCHAR(50) | 8 Standar SNP (nullable untuk dana non-BOS) |
| `planned_amount` | BIGINT | Anggaran rencana (> 0) |
| `realized_amount` | BIGINT | Realisasi (default 0) |
| `remaining_amount` | BIGINT | `planned - realized` (denormalisasi) |
| `quarter` | INT | Triwulan 1–4 (nullable) |

**8 Standar Nasional Pendidikan (BOS component):**

| Kode | Standar |
|------|---------|
| `standar_kompetensi_lulusan` | SKL |
| `standar_isi` | SI |
| `standar_proses` | SP |
| `standar_penilaian` | SPn |
| `standar_ptk` | Pendidik & Tenaga Kependidikan |
| `standar_sarpras` | Sarana & Prasarana |
| `standar_pengelolaan` | Pengelolaan |
| `standar_pembiayaan` | Pembiayaan |

**`budget_realizations`** — Realisasi pengeluaran:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `budget_item_id` | UUID | FK |
| `transaction_date` | DATE | Tanggal transaksi |
| `amount` | BIGINT | Nominal (> 0) |
| `receipt_no` | VARCHAR(50) | Nomor kuitansi |
| `recorded_by` | UUID | Siapa yang mencatat (audit trail) |

---

## Business Rules

### Approval Workflow Engine

1. **Engine bersifat generic** — tidak mengandung business logic domain apapun. Domain menggunakan engine via event-driven integration.
2. **Step diproses secara sequential** — step N+1 hanya menjadi `active` setelah step N `approved`.
3. **Reject di level manapun** langsung mengubah request status menjadi `rejected` — step berikutnya tidak diaktifkan.
4. **`revision_requested`** mengembalikan request ke requestor — setelah revisi, requestor re-submit dan steps di-reset ke `waiting`.
5. **Delegasi dicatat secara eksplisit** (`delegated_to`, `delegated_by`, `delegated_at`). Audit trail menyimpan "Approved by Wakasek (delegated from Kepsek)".
6. **SLA tracking** menggunakan `sla_deadline` per step dan per request — cron job menandai `is_overdue = true`.
7. **`approver_id` di-resolve saat step menjadi active** berdasarkan `approver_role` + tenant context — bukan saat request dibuat.
8. **Chain dengan `condition_rules`** menggunakan rule engine sederhana: `{field, operator, value}` untuk memilih chain berbeda (misal: cuti >3 hari = chain 3 level, cuti ≤3 hari = chain 1 level).

### School Profile

9. **Satu profil per company.** Unique constraint `(tenant_id, company_id)`.
10. **Statistik di profil sekolah bersifat eventual consistent** — di-update via events dari domain lain (StudentCountChanged, TeacherCountChanged, dll).
11. **Jika `is_pesantren = true`, field `pesantren_name` wajib diisi.**
12. **Hanya satu akreditasi yang `is_current = true` per sekolah.** Saat akreditasi baru di-set sebagai current, akreditasi lama otomatis di-unset.
13. **Endpoint publik** (`/public/school-profile`) tersedia tanpa autentikasi untuk website sekolah — field sensitif di-filter.

### Komite Sekolah

> **⚠️ Larangan Pungutan Komite (Permendikbud No. 75/2016 Pasal 10)**
>
> Komite Sekolah **DILARANG** memungut iuran/dana dari orang tua/wali siswa.
>
> Implikasi sistem:
> - Tidak ada fitur "tagihan komite" ke orang tua
> - `dana_komite` sebagai `source_type` di anggaran hanya untuk **donasi sukarela**
> - Pencatatan donasi komite bersifat voluntary — orang tua tidak bisa di-invoice
> - Sistem akan menolak pembuatan invoice dengan `source_type = 'komite'`
>
> Dasar: Permendikbud No. 75/2016 Pasal 10 mengatur bahwa komite sekolah
> dilarang melakukan pungutan dari peserta didik atau orang tua/wali.

14. **Maks 2 periode kepengurusan** (`term_number` antara 1 dan 2) sesuai Permendikbud 75/2016.
15. **Hanya 1 term aktif per sekolah.** Aktivasi term baru otomatis menonaktifkan term sebelumnya.
16. **Posisi ketua, sekretaris, dan bendahara hanya satu per term** — anggota bisa multiple.
17. **Quorum rapat = 50%+1 anggota hadir** — dihitung di application layer, bukan constraint database.
18. **Komite adalah salah satu approver dalam chain RKAS** (step 3 dari chain: bendahara → kepsek → komite → dinas).

### Integrasi ARKAS (Aplikasi RKAS Kemdikbud)

ARKAS adalah sistem wajib Kemdikbud untuk sekolah penerima dana BOS. Sekolah harus
menginput RKAS ke ARKAS sebelum dana BOS cair.

**Hubungan SekolahPro RKAS ↔ ARKAS:**
- SekolahPro RKAS adalah sistem perencanaan internal (lebih detail dan fleksibel)
- ARKAS adalah sistem pelaporan ke Kemdikbud
- Sekolah harus menginput ke kedua sistem — **tidak ada API resmi ARKAS untuk integrasi otomatis**
- SekolahPro menyediakan **export format CSV** kompatibel ARKAS untuk meminimalkan double-entry

> **Catatan**: Komponen BOS dapat berubah setiap tahun berdasarkan Permendikbud terbaru.
> Admin sistem dapat memperbarui daftar komponen BOS dari menu konfigurasi — tidak hardcoded.

### LPJ BOS per Triwulan

Sekolah penerima BOS wajib membuat Laporan Pertanggungjawaban (LPJ) per triwulan:
- **Format**: rekapitulasi penerimaan + pengeluaran per komponen (8 SNP / 4 IASP)
- **Sumber data**: `budget_realizations` di SekolahPro
- **Generate laporan**: `GET /api/v1/reports/lpj-bos?period={Q1|Q2|Q3|Q4}&year={yyyy}`
- **Output**: PDF + Excel dengan breakdown per komponen BOS
- **Penerima laporan**: dinas pendidikan kabupaten/kota (upload ke portal BOS Kemdikbud)

### Budget / RKAS

19. **Satu RKAS per tahun ajaran per sekolah.** Unique constraint `(tenant_id, company_id, academic_year_id)`.
20. **Realisasi tidak boleh melebihi `remaining_amount`.** Transaksi yang melebihi sisa anggaran ditolak (422).
21. **`remaining_amount` adalah denormalisasi** yang harus dijaga konsisten setiap realisasi dicatat.
22. **Dana BOS harus dikategorikan ke 8 Standar SNP** (`bos_component`) untuk pelaporan ke Kemendikbud.
23. **RKAS memerlukan approval multi-level via S047:** bendahara → kepsek → komite → dinas pendidikan.
24. **RKAS bukan sistem akuntansi penuh** (double-entry) — hanya budget planning dan realisasi tracking. Jurnal umum dan neraca di luar scope.

---

## State Machine: Approval Request

```
draft (entitas di domain asal masih draft)
  ↓ submit
pending
  ↓ step 1 mulai diproses
in_review
  ↓ semua step approved          ↓ ada step reject       ↓ ada revision_requested
approved                         rejected                revision_requested
                                                            ↓ resubmit
                                                          pending (steps reset)

pending/in_review → cancelled  (requestor cancel)
pending/in_review → expired    (SLA terlewat, cron job)
```

## State Machine: RKAS

```
draft → submitted (bendahara submit) → approved (semua level approve) → closed (akhir tahun)
                                     ↓ revisi diminta
                                   revised → (ulangi proses approval)
```

---

### IASP 2020 (Instrumen Akreditasi Satuan Pendidikan)

BAN-S/M menggunakan IASP 2020 menggantikan 8 SNP murni. IASP 2020 memiliki
**4 komponen utama** yang masing-masing dinilai berdasarkan data di SekolahPro:

| Komponen IASP 2020 | Data SekolahPro yang Relevan |
|-------------------|------------------------------|
| 1. Mutu Lulusan | Nilai rata-rata, kelulusan, prestasi siswa (S013, S018) |
| 2. Proses Pembelajaran | Jurnal mengajar (S025), RPP (S024), absensi siswa (S008) |
| 3. Mutu Guru | PKG (S028), PKB (S029), kualifikasi guru (S012) |
| 4. Manajemen Sekolah | RKAS (S050), komite (S049), profil sekolah (S048) |

SekolahPro dapat menghasilkan **Laporan Persiapan Akreditasi** yang menampilkan
data per komponen IASP 2020 untuk membantu sekolah mempersiapkan visitasi.

> **Catatan**: Masa berlaku akreditasi 5 tahun (Permendikbud No. 13/2018).
> Sejak 2020, BAN-S/M menerapkan perpanjangan otomatis untuk sekolah Terakreditasi A
> yang memenuhi syarat. Field `valid_until` dapat di-override manual oleh admin
> saat ada kebijakan khusus dari BAN-S/M.

---

## Key Decisions & Rationale

1. **Engine approval generic (S047)** menggantikan state machine per-domain — mencegah duplikasi kode di 7+ domain yang memerlukan approval.
2. **Delegasi approval** diimplementasikan untuk mengatasi bottleneck kepsek yang sering dinas luar — wakasek bisa mewakili dengan audit trail yang jelas.
3. **School profile sangat read-heavy** (dibaca oleh hampir semua domain untuk header dokumen, profil publik, dll) — Vernon pattern adalah pilihan optimal dengan denormalisasi statistik.
4. **Statistik sekolah via event-driven** (bukan COUNT query real-time) — mencegah bottleneck query pada sekolah besar.
5. **RKAS bukan akuntansi penuh** — fokus pada budget tracking yang sekolah benar-benar butuhkan. Sistem akuntansi lengkap over-engineering untuk MVP.
6. **Komite sebagai domain mandiri** (bukan sub-domain dari guardians) karena anggota komite bisa bukan orang tua siswa, dan komite punya lifecycle sendiri (periode, rapat, keputusan).
7. **`approval_steps.approver_id` di-resolve saat active** (bukan saat request dibuat) agar perubahan jabatan tidak merusak requests yang sedang berjalan.

---

## Integration Points

### Internal

| Domain | Arah | Deskripsi |
|--------|------|-----------|
| S024 Lesson Plan | → S047 | RPP memerlukan approval wakasek kurikulum |
| S030 Leave Request | → S047 | Pengajuan cuti menggunakan approval engine |
| S041 Facility Booking | → S047 | Booking fasilitas memerlukan approval admin/wakasek |
| S046 Correspondence | → S047 | Surat keluar penting perlu persetujuan kepsek |
| S050 RKAS | → S047 | RKAS melalui chain: bendahara → kepsek → komite → dinas |
| S049 Committee | → S047 | Ketua komite menjadi approver di chain RKAS |
| S044 Notification | ← S047 | Notifikasi ke approver saat step active; notifikasi ke requestor saat selesai |
| S046 Correspondence | ← S048 | Header surat otomatis dari profil sekolah (letter_code, letter_number_format) |
| S018 Rapor | ← S048 | Nama sekolah, NPSN, alamat dari profil sekolah untuk cetak rapor |
| S042 Parent Portal | ← S048 | Profil sekolah untuk display di portal |
| S055 Dapodik | ← S048 | NPSN, data sekolah dikirim ke Dapodik |

### Eksternal

| Sistem | Keterangan |
|--------|-----------|
| **Dapodik** | NPSN, profil sekolah, data akreditasi dilaporkan via S055 |
| **BAN-S/BAN-SM** | Data akreditasi sekolah diinput manual; riwayat tersimpan di S048 |
| **Dinas Pendidikan** | Approval level 4 RKAS — akses via approval inbox S047 (future: integrasi sistem Dinas) |
| **Kemenag / EMIS** | NSPP pesantren dan data pondok untuk lembaga pesantren |
