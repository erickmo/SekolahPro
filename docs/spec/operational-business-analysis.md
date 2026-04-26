# Analisis Operasional dan Bisnis SekolahPro — Perspektif COO

> Dokumen ini adalah asesmen operasional untuk COO, VP Operations, dan Customer Success.
> Dibuat berdasarkan Wave 1 documentation: semua modul sekolah (S001–S058) dan koperasi (K001–K040).

---

## 1. Kompleksitas Implementasi per Modul

### Peta Kompleksitas Implementasi

Skala: 1 (sederhana) → 5 (sangat kompleks)

| Modul | Kompleksitas | Alasan Utama |
|-------|-------------|-------------|
| **Koperasi Konvensional Core** (K001–K015) | ★★★★★ | Double-entry accounting, transaksi immutable, teller session, denominasi fisik, kas harian |
| **Koperasi Syariah / BMT** (semua Islamic mode) | ★★★★★ | Di atas konvensional + 11 jenis akad, ta'zir, ibra', bagi hasil, DPS, PSAK 101-110 |
| **AML/CFT** (K027) | ★★★★☆ | CDD/EDD nasabah, 8 monitoring rules, PPATK reporting, EDD approval multi-level |
| **Payment Gateway** (S051) | ★★★★☆ | Provider multi-platform, webhook idempotency, settlement reconciliation, MDR compliance |
| **Rapor Generation** (S018) | ★★★★☆ | Agregasi data dari 6+ entitas berbeda per siswa, format Kurikulum Merdeka vs K13 |
| **RKAS / Anggaran Sekolah** (S050) | ★★★☆☆ | Approval 4 level, 8 Standar SNP, laporan ke Kemendikbud |
| **SHU** (K016) | ★★★☆☆ | Perhitungan berbasis jasa modal + jasa usaha, distribusi ke ribuan anggota, disahkan RAT |
| **Data Siswa Core** (S001–S008) | ★★★☆☆ | Banyak entitas terkait, Vernon sync, UU PDP consent flow, NISN validation |
| **Governance Koperasi** (K025–K026) | ★★★☆☆ | Authority matrix, maker-checker, RAT management, voting, pemilihan pengurus |
| **Absensi & Nilai** (S008, S011) | ★★☆☆☆ | CRUD standard, tapi volume tinggi (ratusan siswa per hari) |
| **Profil Sekolah & Dapodik** (S048, S055) | ★★☆☆☆ | Format mapping, validasi NISN/NUPTK, tapi tidak ada business logic kompleks |

### Urutan Implementasi yang Direkomendasikan (Risk-First)

**Fase 1 — Fondasi Wajib Ada:**
1. Multi-tenant core, auth, RBAC
2. Student core + guardian + academic year + class rooms
3. Absensi + nilai + rapor (S001–S018) → ini deliverable pertama yang dirasakan pengguna
4. SPP / keuangan sekolah dasar (S009)

**Fase 2 — Monetisasi Sekolah:**
5. Payment gateway (S051) → memungkinkan pembayaran online SPP
6. Dapodik export (S055) → menghilangkan pekerjaan manual terbesar operator
7. Kurikulum + jadwal (S019–S023)
8. Laporan & analytics (S054) → retention tool untuk kepala sekolah

**Fase 3 — Koperasi (Hanya Enterprise Tier):**
9. Koperasi konvensional core (K001–K015) — tanpa Islamic mode
10. Governance + RAT (K025–K026)
11. AML/CFT + UU PDP compliance (K027–K028)
12. BMT/Islamic mode (semua akad syariah) — **Terakhir, paling kompleks**

---

## 2. Kebutuhan Model Dukungan Pelanggan

### Profil Tim Support yang Dibutuhkan

| Level | Spesialisasi | Kompetensi Wajib |
|-------|-------------|-----------------|
| L1 (Frontline) | Onboarding, operasional dasar | Sistem sekolah Indonesia, Dapodik dasar, SPP |
| L2 (Teknis) | Troubleshooting sistem | SQL dasar, pemahaman CQRS, bug reporting |
| L3 (Domain Expert) | Koperasi, BMT, regulasi | Akuntansi koperasi, UU Koperasi, OJK, Fatwa DSN-MUI |
| L4 (Engineering) | Escalasi teknis | Akses codebase, database, log analysis |

### Topik Support yang Paling Sering Terjadi (Proyeksi)

| Kategori | Volume Perkiraan | Urgency |
|---------|-----------------|---------|
| Import data dari sistem lama (Excel/Jibas) | **Sangat Tinggi** — setiap onboarding baru | MEDIUM |
| Dapodik format tidak sesuai | **Tinggi** — setiap semester | HIGH |
| Sinkronisasi data stale (Vernon pending) | **Sedang** — setelah bulk update | MEDIUM |
| Reset password & akses user | **Tinggi** — harian | LOW |
| Kesalahan perhitungan SPP/SHU | **Rendah** — tapi dampak tinggi | CRITICAL |
| Setup akad syariah baru | **Rendah** — tapi butuh expertise DPS | HIGH |
| Transaksi koperasi yang salah (perlu reversal) | **Sedang** | HIGH |
| Laporan OJK/PPATK tidak sesuai format | **Rendah** | CRITICAL |

### SOP Support yang Harus Dibuat Sebelum Launch

- [ ] Prosedur onboarding sekolah (data import, konfigurasi awal, user setup)
- [ ] Prosedur rollback import data (72 jam window, per ADR-017)
- [ ] Prosedur reversal transaksi koperasi yang salah
- [ ] Prosedur reset tenant type (karena immutable — butuh alur khusus)
- [ ] Escalation matrix untuk insiden P1/P2 koperasi
- [ ] Panduan bahasa Arab/terminologi Islam untuk support agent (Ustadz, Marhalah, Akad, dll)

---

## 3. Kepatuhan Regulasi Operasional

### Peta Kewajiban Regulasi dan Owner Operasional

| Regulasi | Kewajiban Operasional | Frekuensi | Owner |
|---------|----------------------|-----------|-------|
| **UU PDP** | Consent collection per siswa baru | Setiap onboarding siswa | CS / Operator Sekolah |
| **UU PDP** | Respons Data Subject Request (DSR) | Dalam 72 jam (acknowledgment), 7–14 hari (penyelesaian) | Tim Compliance |
| **UU PDP** | Breach notification ke otoritas | Dalam 72 jam setelah insiden | CEO + Tim Legal |
| **Kemendikbud / Dapodik** | Export data per semester | 2x/tahun per sekolah | Operator Sekolah (dibantu CS) |
| **Dinas Koperasi** | Laporan RAT tahunan | 1x/tahun per koperasi | Admin Koperasi |
| **Dinas Koperasi** | Laporan keuangan neraca | Tahunan | Bendahara Koperasi |
| **OJK (jika LKM)** | Laporan kolektibilitas | Bulanan | Compliance Koperasi |
| **PPATK (AML)** | LTKM (Laporan Transaksi Keuangan Mencurigakan) | Dalam 3 hari kerja setelah deteksi | Compliance Officer Koperasi |
| **PPATK (AML)** | TKM (Transaksi > Rp 500 juta) | Bulanan, paling lambat tanggal 10 | Compliance Officer Koperasi |
| **DJP** | PPN 11% di invoice SaaS | Bulanan | Tim Finance SekolahPro |
| **DSN-MUI** | Review kesesuaian akad per produk baru (BMT) | Per produk baru | DPS + Tim Legal |

### Jadwal Compliance Tahunan (Kalender Operasional)

```
JANUARI
├── Laporan keuangan koperasi (akhir tahun buku)
├── Persiapan RAT (60 hari sebelum)
└── TKM bulan Desember (paling lambat tanggal 10)

FEBRUARI–MARET
├── Pelaksanaan RAT
├── Laporan RAT ke Dinas Koperasi
└── Distribusi SHU (jika disetujui RAT)

APRIL
└── TKM bulan Maret

JULI–AGUSTUS
├── Awal tahun ajaran baru
├── Dapodik export Semester 2 selesai
├── Setup academic year baru
└── Onboarding siswa baru (consent flow wajib)

NOVEMBER–DESEMBER
├── Dapodik export Semester 1
└── Persiapan laporan tahunan koperasi
```

---

## 4. Kompleksitas Migrasi Data dari Sistem Lama

### Sumber Data yang Harus Diservis

| Sistem Lama | Prevalensi | Kesulitan Migrasi |
|------------|-----------|------------------|
| Excel / Google Sheets | **Sangat Umum** | MEDIUM — Format tidak standar; perlu data cleaning |
| Dapodik (export CSV) | **Sangat Umum** | LOW — Sudah dalam format terstandarisasi |
| Jibas | **Umum (Jawa)** | MEDIUM — Export tersedia tapi mapping field manual |
| AdminSekolah | **Umum** | MEDIUM — Format proprietary |
| Sistem koperasi lama (manual ledger) | **Koperasi** | HIGH — Saldo historis perlu verifikasi manual |
| Sistem koperasi standalone | **Koperasi besar** | HIGH — Struktur data berbeda, rekonsiliasi saldo wajib |

### Alur Migrasi yang Sudah Dirancang (ADR-017)

```
Download Template Excel/CSV
    ↓
Isi Data (oleh admin sekolah)
    ↓
Upload ke SekolahPro
    ↓
DRY-RUN VALIDASI (WAJIB)
  → Validasi format, tipe data, constraint
  → Deteksi duplikat terhadap data existing
  → Return error report TANPA menyimpan apapun
    ↓
Admin Review (total rows, valid, error, duplikat)
    ↓
Konfirmasi Import
    ↓
Batch Processing (async jika > 1.000 rows)
    ↓
Vernon Sync Post-Import
    ↓
Laporan Verifikasi + Rollback option (72 jam)
```

### Tantangan Migrasi yang Perlu Diantisipasi

1. **Data koperasi historis** — Saldo rekening anggota per hari ini harus 100% akurat. Setiap discrepancy akan diketahui anggota segera. **Rekomendasi:** Selalu lakukan cut-off date migration + rekonsiliasi fisik sebelum go-live koperasi.

2. **Data siswa tanpa NISN** — Banyak sekolah kecil memiliki siswa yang belum punya NISN. Sistem harus mampu menyimpan siswa tanpa NISN (nullable) sambil mengingatkan admin untuk melengkapi.

3. **Data guru tanpa NIP/NUPTK** — Guru honorer tidak punya NIP. Field ini wajib partial-unique (bukan NOT NULL).

4. **Konflik duplicate siswa** — Siswa yang sudah pernah diinput manual dan kemudian diimport dari Dapodik akan menghasilkan duplikat. Aturan konfllik resolusi (skip/overwrite/merge) harus dipahami admin sebelum import.

5. **Nominal SPP historis** — Beberapa sekolah ingin memasukkan riwayat pembayaran SPP 2–3 tahun ke belakang. Ini memerlukan batch import transaksi dengan timestamp historis yang harus diizinkan di admin mode.

### Estimasi Waktu Onboarding per Ukuran Sekolah

| Ukuran Sekolah | Estimasi Waktu Onboarding | Bottleneck Utama |
|---------------|--------------------------|-----------------|
| Kecil (< 300 siswa, tanpa koperasi) | 3–5 hari kerja | Data cleaning Excel |
| Menengah (300–1.000 siswa, tanpa koperasi) | 1–2 minggu | Dapodik mapping + nilai historis |
| Besar (> 1.000 siswa + koperasi konvensional) | 3–4 minggu | Rekonsiliasi saldo koperasi |
| Besar (> 1.000 siswa + BMT syariah) | 5–8 minggu | Akad syariah setup + DPS review + PSAK mapping + rekonsiliasi |

---

## 5. Kebutuhan Training per Peran Pengguna

### Matriks Training

| Peran | Modul Kritis | Durasi Training | Format |
|-------|-------------|----------------|--------|
| **Admin Sistem / IT Sekolah** | Semua modul, user management, import data, konfigurasi | 2 hari | Tatap muka / video |
| **Kepala Sekolah (Kepsek)** | Dashboard, rapor, RKAS, approval workflows, laporan Dapodik | 4 jam | Video on-demand |
| **Wakasek Kurikulum** | Jadwal, RPP, assessment, nilai, rapor | 4 jam | Video on-demand |
| **Wali Kelas / Guru** | Absensi harian, input nilai, jurnal mengajar | 2 jam | Quick-start guide |
| **Bendahara Sekolah** | SPP, payment gateway, RKAS, laporan keuangan | 6 jam | Tatap muka + praktik |
| **Operator TU** | Data siswa, import Dapodik, dokumen, surat menyurat | 4 jam | Video on-demand |
| **Orang Tua / Wali Murid** | Portal orang tua, cek nilai, bayar SPP online | 30 menit | Video singkat + FAQ |
| **Ketua Koperasi** | Dashboard eksekutif, laporan, health indicators, RAT | 4 jam | Tatap muka |
| **Teller Koperasi** | Transaksi harian, kas, sesi teller, denominasi fisik | 1 hari penuh | Tatap muka + simulasi |
| **Admin Koperasi** | Setup produk, manajemen anggota, payroll deduction, laporan OJK | 2 hari | Tatap muka |
| **Compliance Officer Koperasi** | AML/CFT, CDD/EDD, PPATK reporting, data privacy | 1 hari | Tatap muka + sertifikasi |
| **DPS (Dewan Pengawas Syariah)** | Review akad, laporan syariah, audit syariah | 3 jam | Konsultasi langsung |

### Catatan Khusus: Training Koperasi Syariah

Training teller dan admin untuk BMT Islamic mode memerlukan pemahaman **konsep syariah** yang tidak dapat disingkat:
- Perbedaan wadiah, mudharabah, murabahah, musyarakah, ijarah, qardh
- Mekanisme ibra' (diskon pelunasan dini) yang tidak ada di koperasi konvensional
- Ta'zir (denda nominal) yang masuk dana sosial, bukan pendapatan koperasi
- Mekanisme bagi hasil yang berbeda dengan bunga flat/declining
- Peran DPS dalam approval produk baru dan audit syariah

**Rekomendasi:** Kembangkan modul sertifikasi online untuk teller koperasi syariah sebelum onboarding pesantren pertama.

---

## 6. SLA Requirements Berdasarkan Kekritisan Fitur

### Tier Kekritisan Fitur

| Tier | Fitur | SLA Uptime | SLA Response Time | Downtime Toleransi |
|------|-------|-----------|------------------|-------------------|
| **KRITIS** | Transaksi koperasi (setoran, penarikan, angsuran) | 99,9% | < 2 detik | Max 1 jam/bulan |
| **KRITIS** | Payment gateway SPP (callback webhook) | 99,9% | < 3 detik | Max 1 jam/bulan |
| **KRITIS** | Auth / login | 99,9% | < 1 detik | Max 1 jam/bulan |
| **TINGGI** | Absensi harian siswa | 99,5% | < 3 detik | Max 4 jam/bulan |
| **TINGGI** | AML monitoring alerts | 99,5% | Event-driven | Max 4 jam/bulan |
| **TINGGI** | Saldo & sesi teller | 99,5% | < 2 detik | Max 4 jam/bulan |
| **STANDAR** | Input nilai / rapor | 99,0% | < 5 detik | Max 7 jam/bulan |
| **STANDAR** | Dashboard & laporan | 99,0% | < 3 detik | Max 7 jam/bulan |
| **RENDAH** | Kalender akademik, dokumen sekolah | 98,5% | < 5 detik | Max 12 jam/bulan |
| **RENDAH** | Ekspor Dapodik (manual, 2x/tahun) | 98,0% | < 30 detik (async) | Max 1 hari |

### Prosedur Offline untuk Koperasi (Wajib Ada)

Per ADR-K029, koperasi **wajib** memiliki prosedur offline jika sistem down:
- Setoran: gunakan form manual triplicate
- Penarikan: hanya ≤ Rp 2 juta, dengan catatan manual
- Angsuran: kwitansi manual
- Semua transaksi manual di-input ulang saat sistem recovery (dengan verifikasi supervisor)

**Ini bukan fitur teknis — ini adalah SOP operasional** yang harus dilatih kepada setiap teller sebelum go-live.

---

## 7. Matriks Risiko Operasional

| Risiko | Probabilitas | Dampak | Rating | Mitigasi |
|--------|------------|--------|--------|---------|
| Data koperasi salah input sehingga saldo tidak sesuai fisik | MEDIUM | SANGAT TINGGI | **KRITIS** | Verifikasi saldo sebelum go-live; prosedur rekonsiliasi harian; maker-checker rule |
| Koperasi onboard sebelum OJK review selesai | LOW | SANGAT TINGGI | **KRITIS** | Hard gate: Enterprise tier tidak bisa diaktifkan sebelum checklist OJK selesai |
| Data siswa diproses tanpa consent wali | MEDIUM | TINGGI | **HIGH** | Consent flow wajib di onboarding; sistem block data processing jika consent belum ada |
| Dapodik export format berubah saat sekolah sedang submit | MEDIUM | TINGGI | **HIGH** | Monitoring perubahan spesifikasi Dapodik; dedicated engineer untuk maintain format mapping |
| Teller koperasi membuat kesalahan transaksi besar | MEDIUM | TINGGI | **HIGH** | Four-eyes principle untuk transaksi > threshold; training intensif; approval multi-level |
| Sekolah tidak bisa migrate data lama (berhenti berlangganan) | MEDIUM | MEDIUM | **MEDIUM** | Layanan migrasi data profesional; export tool yang mudah; onboarding success metric |
| Performance degradasi saat bulk import 10.000+ siswa | MEDIUM | MEDIUM | **MEDIUM** | Load testing sebelum launch; async processing; rate limiting import |
| Laporan PPATK/OJK tidak terkirim tepat waktu | LOW | TINGGI | **MEDIUM** | Alert otomatis H-3 sebelum deadline; SOP eskalasi untuk compliance officer |
| Operator sekolah salah konfigurasi institution type | LOW | TINGGI | **MEDIUM** | Institution type immutable setelah aktif; dokumentasi onboarding yang sangat jelas |
| Support ticket membludak saat tahun ajaran baru | HIGH | MEDIUM | **MEDIUM** | Pre-emptive training, self-service portal, FAQ yang lengkap; staffing plan untuk Juli–Agustus |
| Koperasi BMT diaudit DPS dan ditemukan inkonsistensi akad | LOW | MEDIUM | **MEDIUM** | Review pre-launch dengan DPS eksternal; testing komprehensif untuk semua akad syariah |

---

## 8. Rekomendasi Prioritas COO

### Sebelum Soft Launch (Pro Tier)

1. **Rekrut domain expert koperasi minimal 1 orang di tim support** — transaksi keuangan koperasi tidak bisa di-support oleh generalist IT. Perlu seseorang yang mengerti double-entry, SHU, dan AML.

2. **Buat SOP onboarding tertulis per segmen** (sekolah kecil, menengah, pesantren dengan koperasi) — tanpa SOP tertulis, waktu onboarding tidak bisa dikontrol dan menjadi bottleneck pertumbuhan.

3. **Load test dengan skenario tahun ajaran baru** — Juli–Agustus adalah peak load saat ribuan sekolah onboard, ratusan ribu siswa diimport, dan sistem pertama kali digunakan secara massal.

4. **Siapkan prosedur offline untuk koperasi** — setiap kantor koperasi yang menggunakan SekolahPro harus memiliki form manual. Ini bukan opsional per regulasi BCP/DRP.

### Sebelum Enterprise Launch (Koperasi Tier)

5. **Pastikan OJK review selesai dan terdokumentasi** — ini adalah hard blocker yang tidak bisa disiasati. Tanpa ini, platform berisiko hukum.

6. **Lakukan rekonsiliasi saldo sebelum setiap go-live koperasi baru** — gunakan independent reconciliation (data SekolahPro vs buku kas fisik) sebelum koperasi pertama kali "live" untuk transaksi riil.

7. **Bangun program sertifikasi teller BMT** sebelum onboarding koperasi syariah pertama — teller yang tidak paham akad syariah akan membuat kesalahan yang sulit dikoreksi karena immutable transactions.

8. **Tetapkan Service Level Agreement yang jelas** di perjanjian dengan pelanggan Enterprise — SLA 4 jam untuk support, 99,9% uptime untuk transaksi kritis, prosedur eskalasi yang jelas untuk insiden P1.
