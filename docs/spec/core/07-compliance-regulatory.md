# 07 — Compliance, Regulasi, dan Model Bisnis SekolahPro

Dokumen ini menjabarkan kepatuhan regulasi yang berlaku untuk SekolahPro, implikasinya pada arsitektur sistem, serta keputusan model bisnis dan pricing. SekolahPro beroperasi di persimpangan dua sektor teregulasi: **pendidikan** dan **jasa keuangan** di Indonesia.

---

## ADR References

| ADR | Judul | Relevansi |
|-----|-------|-----------|
| ADR-018 | Regulatory Compliance Matrix | Peta regulasi dan dampak per modul |
| ADR-BIZ-001 | Business Model & Pricing Strategy | Model harga, tier, dan unit economics |

---

## 1. Lanskap Regulasi

SekolahPro wajib patuh terhadap minimal **6 regulator** secara simultan:

| Regulator | Ruang Lingkup | Modul yang Terpengaruh |
|-----------|--------------|------------------------|
| **Kemendikbud** | Dapodik, NISN, NUPTK, NPSN, pelaporan semester | Data siswa, guru, kelas, rapor |
| **Dinas Koperasi** | Registrasi koperasi, pelaporan RAT, SHU, modal minimum | Modul koperasi/BMT (K001-K024) |
| **OJK** | Berlaku jika aset/simpanan koperasi > Rp 10 miliar | Simpan-pinjam BMT |
| **Bank Indonesia** | Regulasi e-wallet (PBI 20/6/2018), QRIS, payment aggregator | E-wallet, payment gateway |
| **UU PDP (UU 27/2022)** | Perlindungan data pribadi, khususnya data anak (<18 tahun) | Semua data siswa |
| **DJP (Pajak)** | PPN 11% atas langganan SaaS, PPh withholding payroll | Billing SaaS, modul payroll |

**Compliance bukan fitur opsional** — ia adalah constraint arsitektur yang mempengaruhi desain setiap modul sejak awal.

---

## 2. Compliance Matrix per Modul

| Modul | Regulator | Kewajiban Utama | Prioritas |
|-------|-----------|-----------------|-----------|
| Data Siswa | UU PDP | Consent wali wajib, retensi max 5 tahun, right to erasure | **Critical** |
| Dapodik Export | Kemendikbud | Format ekspor sesuai spesifikasi Dapodik (NISN, NPSN) | High |
| Koperasi/BMT | Dinas Koperasi | Laporan RAT, perhitungan SHU, modal minimum | High |
| E-wallet | Bank Indonesia | Tetap sebagai "spending interface", bukan e-money | **Critical** |
| Payment Gateway | Bank Indonesia | Aturan MDR QRIS, payment aggregator | High |
| Payroll | BPJS | Potongan BPJS Kesehatan + Ketenagakerjaan | Medium |
| SaaS Billing | DJP | PPN 11% atas langganan | Medium |

---

## 3. Perlindungan Data Pribadi (UU PDP)

Mayoritas pengguna SekolahPro adalah **siswa di bawah 18 tahun** — UU No. 27/2022 memberikan perlindungan ekstra untuk data anak.

### Keputusan Implementasi

**1. Consent Wali Wajib**

Semua data siswa dikategorikan sebagai data anak. Proses onboarding wajib menyertakan persetujuan wali/orang tua sebelum data diproses. Arsitektur:

```
Consent Service:
  - Menyimpan record consent per siswa
  - Versioning (consent v1, v2, ...) — jika kebijakan berubah
  - Consent withdrawal = soft-disable akses data
```

**2. Retensi Data (Maksimal 5 Tahun)**

Setelah siswa lulus/keluar, data disimpan maksimal 5 tahun (konfigurabel per institusi). Setelah itu, data di-anonymize atau dihapus otomatis via scheduled job.

```
Data Lifecycle Service:
  - Retention policy per tenant (konfigurabel)
  - Auto-anonymize setelah N tahun
  - Hard delete on request
```

**3. Right to Erasure (Hak Penghapusan)**

Sistem wajib mendukung hard delete atas permintaan wali:

```
DELETE /api/v1/students/{id}/personal-data
  → Cascade delete ke semua tabel terkait (student_*_records, dll)
  → Audit log tetap disimpan (tanpa PII)
```

**4. Data Portability**

Ekspor seluruh data satu siswa dalam format standar (JSON/CSV) dalam < 72 jam setelah permintaan:

```
GET /api/v1/students/{id}/export
  → Returns: biodata, riwayat nilai, absensi, SPP, dll
  → Format: JSON atau CSV
```

**5. Breach Notification (72 Jam)**

Jika terjadi data breach, notifikasi wajib dikirim ke:
- Subjek data (wali/orang tua) — dalam 72 jam
- Otoritas PDP — dalam 72 jam

```
Breach Notification Service:
  - Alert pipeline
  - Template notifikasi (email + SMS/WhatsApp)
  - Audit trail setiap notifikasi yang dikirim
```

### Implikasi Teknis UU PDP pada Arsitektur

- Setiap field yang menyimpan data pribadi anak wajib teridentifikasi dan terdokumentasi
- Hard delete harus cascade ke **semua** tabel terkait, termasuk Vernon `_data` di tabel lain
- Anonymization tidak boleh merusak referential integrity — gunakan placeholder (`[DELETED]`, `XXXXXXXX`)
- Audit log harus memisahkan PII dari metadata (siapa yang mengakses, kapan) — log boleh disimpan lama, tapi bukan PII
- Data retention scheduler wajib ada sebelum production launch (Q4 2026 deadline)

---

## 4. E-Wallet: Batasan Regulasi Bank Indonesia

SekolahPro e-wallet **bukan** uang elektronik (e-money). Sistem dirancang sebagai **spending interface** — saldo hanya bisa digunakan untuk transaksi internal.

### Yang DILARANG (mencegah kewajiban lisensi e-money BI)

- Transfer antar pengguna (P2P transfer)
- Cash-out / pencairan saldo ke rekening bank
- Top-up dari sumber selain yang disetujui sekolah

### Yang DIIZINKAN

- Pembayaran SPP, kantin, koperasi internal
- Top-up oleh orang tua via payment gateway
- Riwayat transaksi per siswa

Dengan pembatasan ini, sistem tidak memerlukan lisensi e-money dari Bank Indonesia. Jika fitur P2P atau cash-out ditambahkan di masa depan, wajib konsultasi hukum terlebih dahulu.

**Catatan**: Review legal e-wallet vs PBI 20/6/2018 masih in-progress (deadline Q2 2026).

---

## 5. Dapodik Integration (Kemendikbud)

Kemendikbud mewajibkan pelaporan data pokok melalui sistem Dapodik. SekolahPro harus mendukung ekspor format Dapodik:

- **NISN** (Nomor Induk Siswa Nasional) — 10 digit, unik nasional
- **NPSN** (Nomor Pokok Sekolah Nasional) — identifier sekolah
- **NUPTK** (Nomor Unik Pendidik Tenaga Kependidikan) — identifier guru

Implementasi yang dibutuhkan:
```
GET /api/v1/dapodik/export/students
GET /api/v1/dapodik/export/teachers
→ Format CSV/Excel sesuai spesifikasi Dapodik terbaru
```

---

## 6. Koperasi & BMT (Dinas Koperasi / OJK)

Modul koperasi (K001-K024) mengikuti regulasi koperasi Indonesia:

- **Laporan RAT** (Rapat Anggota Tahunan) — format sesuai Dinas Koperasi setempat
- **Perhitungan SHU** (Sisa Hasil Usaha) — distribusi ke anggota berdasarkan kontribusi
- **Modal minimum** — sesuai ketentuan Dinas Koperasi
- **OJK threshold**: jika total aset/simpanan > Rp 10 miliar, koordinasi dengan OJK diperlukan

**Penting**: Fitur koperasi/BMT memerlukan **koordinasi OJK sebelum Enterprise tier diluncurkan**. Ini adalah prerequisite yang tidak bisa dilewati.

---

## 7. Model Bisnis & Pricing

### Model: Per-Siswa/Bulan dengan 3 Tier

| Tier | Harga | Maks Siswa | Modul |
|------|-------|-----------|-------|
| **Starter (Freemium)** | Gratis | 100 siswa | Data siswa + absensi saja |
| **Pro** | Rp 5.000-8.000/siswa/bulan | Unlimited | Semua modul sekolah (S001-S058) |
| **Enterprise** | Rp 10.000-15.000/siswa/bulan | Unlimited | Semua modul sekolah + koperasi/BMT (K001-K024) + E-wallet |

### Proyeksi Revenue (Tahun ke-2)

| Tier | Sekolah | Rata-rata Siswa | Harga | MRR |
|------|---------|----------------|-------|-----|
| Starter | 200 | 80 | Rp 0 | Rp 0 |
| Pro | 50 | 500 | Rp 8.000 | Rp 200.000.000 |
| Enterprise | 20 | 800 | Rp 12.000 | Rp 192.000.000 |
| **Total** | **270** | | | **Rp 392.000.000/bulan** |

ARR potensial: ~Rp 4,7 miliar/tahun

### Revenue Tambahan (15-25% dari MRR)

| Stream | Model |
|--------|-------|
| Payment gateway fee margin | Markup 0,1-0,3% di atas fee provider |
| SMS/WhatsApp notifikasi overage | Per-pesan di atas kuota gratis |
| Custom domain | Rp 50.000/bulan |
| White-label | Rp 500.000/bulan per institusi |
| Implementasi & training | One-time, Rp 2-10 juta/sekolah |

### Break-even Point

- **Total Burn Rate**: Rp 250.000.000-350.000.000/bulan (tim + infra + support + overhead)
- **Break-even**: 35-45 sekolah berbayar (Pro + Enterprise mix), setara ~25.000-30.000 siswa aktif berbayar

### Biaya Regulasi & Compliance

| Item | Estimasi |
|------|----------|
| Konsultasi OJK/BI | Rp 50.000.000-100.000.000 (one-time + annual review) |
| Legal compliance koperasi/BMT | Rp 30.000.000-50.000.000 (one-time) |
| Sertifikasi ISO 27001 | Rp 100.000.000-200.000.000 (annual) |
| PPN 11% atas SaaS | ~11% dari revenue (bulanan) |

---

## 8. Implikasi Pricing pada Arsitektur

### Feature Gating

Pricing tier memerlukan implementasi teknis berikut:

| Kebutuhan Teknis | Prioritas |
|-----------------|-----------|
| Tenant tier configuration di multi-tenant (ADR-004) | P0 |
| Feature flag per tier | P0 (perlu ADR baru) |
| Usage metering: jumlah siswa aktif | P0 (perlu ADR baru) |
| Billing & invoice system | P1 (perlu ADR baru) |
| SMS/WhatsApp usage tracking | P1 |

### Feature Availability per Tier

```
Starter (free):
  ✓ Data siswa (read-only setelah 100 siswa)
  ✓ Absensi dasar
  ✗ Nilai & rapor
  ✗ Keuangan (SPP)
  ✗ Semua modul Pro/Enterprise

Pro:
  ✓ Semua modul sekolah (S001-S058)
  ✗ Modul koperasi/BMT
  ✗ E-wallet

Enterprise:
  ✓ Semua modul sekolah
  ✓ Modul koperasi/BMT (K001-K024)
  ✓ E-wallet (spending interface)
  ✓ Custom domain, SLA 4 jam
```

---

## 9. Action Items Compliance (Timeline)

| # | Aksi | Owner | Deadline |
|---|------|-------|----------|
| 1 | Implementasi consent flow onboarding siswa | Engineering | Q3 2026 |
| 2 | Format ekspor Dapodik (NISN, NPSN) | Engineering | Q3 2026 |
| 3 | Review legal e-wallet vs PBI 20/6/2018 | CFO + Legal | Q2 2026 |
| 4 | Implementasi data retention scheduler | Engineering | Q4 2026 |
| 5 | Endpoint hard delete + data portability | Engineering | Q4 2026 |
| 6 | Breach notification pipeline | Engineering + Ops | Q4 2026 |
| 7 | PPN 11% di billing system | Engineering | Q3 2026 |
| 8 | Validasi SHU & RAT report vs template Dinas Koperasi | Engineering | Q3 2026 |
| 9 | Konsultasi OJK untuk Enterprise tier | CFO + Legal | Q2 2026 |
| 10 | Validasi harga ke 20-30 sekolah target | CEO + Sales | Q2 2026 |

---

## Key Decisions

1. **Compliance by design** — regulasi di-address sejak arsitektur awal, bukan retrofitted; terutama UU PDP yang mempengaruhi seluruh data model siswa

2. **E-wallet sebagai spending interface, bukan e-money** — membatasi fitur P2P dan cash-out untuk menghindari kewajiban lisensi BI yang mahal dan panjang prosesnya

3. **Per-siswa/bulan pricing** — paling adil dan scalable; sekolah kecil bayar lebih murah, revenue tumbuh seiring pertumbuhan sekolah

4. **Pesantren sebagai primary market** — Islamic mode adalah differentiator utama, bukan fitur tambahan; harus ada di Phase 1

5. **Freemium sebagai acquisition funnel** — barrier rendah untuk masuk, konversi natural saat sekolah tumbuh

6. **Enterprise tier gated oleh OJK compliance** — tidak bisa diluncurkan sebelum koordinasi OJK selesai; ini adalah hard dependency bisnis

---

## Constraints & Implications

### Constraints Hard

- Data siswa tidak boleh diproses tanpa consent wali yang valid — ini adalah kewajiban hukum, bukan preferensi
- E-wallet tidak boleh memiliki fitur P2P transfer atau cash-out tanpa lisensi BI
- Enterprise tier (koperasi/BMT + e-wallet) tidak boleh diluncurkan sebelum OJK review selesai
- PPN 11% wajib termasuk dalam harga yang ditampilkan ke pengguna
- Data retensi maksimal 5 tahun — sistem wajib mendukung auto-anonymize/delete

### Implications untuk Engineering

- Setiap tabel yang menyimpan data siswa harus mendukung hard delete cascade yang lengkap
- Audit log harus memisahkan PII dari metadata audit — log aktivitas boleh disimpan lama tapi tanpa nama/NISN
- Feature gating system wajib ada sebelum Pro tier diluncurkan — tidak ada workaround
- Usage metering (hitung siswa aktif per tenant) wajib ada sebelum billing system diimplementasi
- Import data massal via ADR-017 harus menyertakan consent batch (jika belum ada consent per siswa)
- Dapodik export format harus mengikuti spesifikasi terbaru — Kemendikbud sering mengupdate format

### Implications untuk Go-to-Market

- Demo dan marketing materials harus lead dengan use case pesantren (Islamic mode)
- Pricing harus divalidasi ke 20-30 sekolah target sebelum public launch
- Implementasi & training sebagai revenue stream memerlukan SOP dan tim CS yang terlatih
- ISO 27001 certification diperlukan untuk meyakinkan sekolah besar dan pemerintah terkait keamanan data
