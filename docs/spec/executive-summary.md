# Ringkasan Eksekutif SekolahPro — Perspektif CEO

> Dokumen ini adalah sintesis strategis untuk pengambil keputusan tingkat C-Suite.
> Dibuat berdasarkan Wave 1 documentation (Core ADR, Koperasi K001–K040, Sekolah S001–S058).

---

## 1. Visi Platform dan Proposisi Nilai

SekolahPro adalah platform SaaS **dual-purpose** yang menggabungkan dua kebutuhan fundamental institusi pendidikan Indonesia dalam satu ekosistem digital:

1. **Manajemen Sekolah** — akademik, siswa, kurikulum, rapor, keuangan sekolah (SPP, BOS)
2. **Manajemen Koperasi Sekolah** — simpan pinjam, BMT, SHU, teller, akuntansi, kepatuhan OJK

Proposisi nilai utama adalah **eliminating the duplication problem**: sekolah yang saat ini menggunakan 3–5 sistem terpisah (Dapodik manual, spreadsheet SPP, aplikasi koperasi standalone, WhatsApp untuk komunikasi) dapat menggantikan semuanya dengan satu platform yang terintegrasi.

### Keunggulan Differensiasi

| Dimensi | SekolahPro | Kompetitor Umum |
|---------|-----------|-----------------|
| Dual-purpose (sekolah + koperasi) | Ya — satu platform | Tidak — produk terpisah |
| Mode Islami (pesantren/BMT) native | Ya — built-in | Tidak atau add-on |
| Dual institution type (Umum + Islam) | Ya — 4 kombinasi | Tidak |
| Koperasi syariah (BMT) terintegrasi | Ya | Tidak |
| Pelaporan regulasi otomatis (Dapodik, OJK, PPATK) | Ya | Parsial |
| Multi-tenant dengan hierarki yayasan | Ya | Jarang |

---

## 2. Analisis Pasar Target

### Pasar Primer: Pesantren dan Madrasah (Islamic Schools)

Indonesia memiliki lebih dari **26.000 pondok pesantren** (data Kemenag 2023) dan **80.000+ madrasah** — ini adalah pasar yang secara historis diabaikan oleh vendor SaaS pendidikan mainstream.

Karakteristik pasar pesantren yang menguntungkan SekolahPro:
- **Kebutuhan BMT/koperasi sangat tinggi** — hampir semua pesantren besar memiliki unit koperasi atau BMT
- **Terminologi khusus** (Santri, Ustadz, Halaqah, Marhalah) yang tidak bisa dilayani platform generik
- **Kurikulum Diniyah** di atas kurikulum nasional — tidak ada kompetitor yang mendukung ini
- **Loyalitas komunitas tinggi** — pesantren merekomendasikan antar pesantren via jaringan ulama

### Pasar Sekunder: Sekolah Umum dan Yayasan Pendidikan

- **214.000+ sekolah** di Indonesia (semua jenjang)
- Mayoritas masih menggunakan Excel + Dapodik manual + aplikasi fragmen
- Segmen yang paling mudah dikonversi: yayasan yang mengelola 2–10 sekolah (butuh hierarki multi-tenant)

### Segmentasi Berdasarkan Ukuran

| Segmen | Deskripsi | Fit dengan Tier |
|--------|-----------|----------------|
| Mikro (< 200 siswa) | Sekolah swasta kecil, madrasah desa | Starter (freemium → upgrade) |
| Menengah (200–1000 siswa) | Sekolah swasta/negeri medium, pesantren | Pro |
| Besar (> 1000 siswa) + Koperasi | Pesantren besar, yayasan multi-sekolah | Enterprise |

---

## 3. Posisi Kompetitif

### Lanskap Kompetitor

| Kompetitor | Kekuatan | Kelemahan vs SekolahPro |
|-----------|---------|------------------------|
| **Jibas** | Market share lama, instalasi lokal | Desktop/lokal, tidak ada koperasi, tidak ada Islamic mode |
| **Admin Sekolah** | Populer di Jawa | Tidak ada koperasi, tidak ada multi-tenant |
| **Emis (Kemenag)** | Gratis, Madrasah coverage | Bukan platform manajemen, hanya pelaporan |
| **EduManager / SchoolPro** | SaaS modern | Tidak ada koperasi, tidak ada Islamic mode |
| **Fintech koperasi (Koperasi.go)** | Fokus koperasi | Tidak ada modul sekolah/akademik |

**Kesimpulan:** Tidak ada kompetitor yang melayani kombinasi sekolah + koperasi syariah dalam satu platform. SekolahPro berada di **blue ocean** untuk segmen pesantren.

### Risiko Kompetitif
- **RISK HIGH**: Pemain besar (Google for Education, Microsoft Teams) bisa masuk ke segmen Indonesia jika pasar terbukti menarik
- **RISK MEDIUM**: Vendor lama (Jibas) bisa menambahkan modul koperasi
- **RISK LOW**: Kementerian membuat platform pemerintah yang gratis (sudah ada Dapodik, tapi tidak dikelola sebagai SaaS komersial)

---

## 4. Model Bisnis dan Analisis Tier

### Struktur Pricing

| Tier | Target | Harga | Modul |
|------|--------|-------|-------|
| **Starter (Freemium)** | Akuisisi & uji coba | Gratis (maks 100 siswa) | Siswa + absensi dasar |
| **Pro** | Sekolah menengah | Rp 5.000–8.000/siswa/bulan | Semua modul sekolah (S001–S058) |
| **Enterprise** | Pesantren besar + koperasi | Rp 10.000–15.000/siswa/bulan | Semua modul + Koperasi/BMT (K001–K040) + E-wallet |

### Unit Economics (Proyeksi Tahun ke-2)

| Tier | Sekolah | Rata-rata Siswa | Harga | MRR |
|------|---------|----------------|-------|-----|
| Starter | 200 | 80 | Rp 0 | Rp 0 |
| Pro | 50 | 500 | Rp 8.000 | Rp 200.000.000 |
| Enterprise | 20 | 800 | Rp 12.000 | Rp 192.000.000 |
| **Total** | **270** | | | **Rp 392.000.000/bulan** |

**ARR Potensial Tahun ke-2:** ~Rp 4,7 miliar

**Break-even Point:** 35–45 sekolah berbayar (~25.000–30.000 siswa aktif berbayar)

**Burn Rate Bulanan:** Rp 250–350 juta/bulan (tim + infrastruktur + support + compliance)

### Revenue Tambahan (15–25% dari MRR)

| Stream | Model Pendapatan |
|--------|----------------|
| Payment gateway margin | Markup 0,1–0,3% di atas biaya provider |
| SMS/WhatsApp notifikasi overage | Per-pesan di atas kuota gratis |
| Custom domain | Rp 50.000/bulan |
| White-label | Rp 500.000/bulan per institusi |
| Implementasi & training | One-time Rp 2–10 juta/sekolah |

---

## 5. Rekomendasi Go-to-Market

### Fase 1 — Penetrasi Pasar (Q1–Q3 2026)

**Prioritas #1: Pesantren Besar sebagai Anchor Customer**

Strategi "land dengan pesantren flagship":
- Target 3–5 pesantren besar (> 2.000 santri) dengan koperasi aktif sebagai referensi
- Berikan Enterprise tier dengan subsidi 50% selama 6 bulan pertama
- Dokumentasikan ROI dan jadikan case study
- Manfaatkan jaringan ulama/kyai untuk word-of-mouth

**Prioritas #2: Validasi Harga**
- Lakukan pricing discovery ke 20–30 sekolah target sebelum soft launch
- Validasi apakah Rp 5.000–8.000/siswa/bulan diterima pasar (ada risiko terlalu mahal untuk sekolah kecil)
- Pertimbangkan tier tambahan: Rp 3.000/siswa/bulan untuk sekolah < 200 siswa

**Prioritas #3: Channel Partnership**
- Partnership dengan Dinas Koperasi daerah (mereka butuh platform laporan digital)
- Partnership dengan asosiasi pesantren (RMI, PBNU, PP Muhammadiyah)
- Partnership dengan payment provider (Midtrans/Xendit bisa mereferensikan sekolah)

### Fase 2 — Akselerasi (Q4 2026–2027)

- Self-service onboarding setelah produk stabil
- Marketplace fitur untuk add-on (GPS tracking, e-library, dll)
- Program reseller via guru/operator sekolah
- Ekspansi ke Malaysia (komunitas pesantren juga besar)

---

## 6. Risiko Kunci

### Risiko Regulasi

| Risiko | Rating | Keterangan |
|--------|--------|-----------|
| Enterprise tier tidak bisa diluncurkan sebelum konsultasi OJK selesai | **HIGH** | Hard dependency — tidak ada exception |
| E-wallet harus tetap sebagai "spending interface" (bukan e-money) | **HIGH** | Fitur P2P/cash-out akan memicu kewajiban lisensi BI |
| UU PDP (UU 27/2022) consent flow belum terimplementasi | **HIGH** | Data siswa tidak bisa diproses legal tanpa consent wali |
| Format ekspor Dapodik berubah setiap tahun | **MEDIUM** | Kemendikbud sering update spesifikasi tanpa notifikasi awal |
| OJK menetapkan threshold aset koperasi lebih rendah dari Rp 10 miliar | **MEDIUM** | Bisa memaksa seluruh nasabah Enterprise masuk ke jurisdiksi OJK |

### Risiko Pasar

| Risiko | Rating | Keterangan |
|--------|--------|-----------|
| Harga tidak diterima pasar (terlalu mahal untuk sekolah kecil) | **HIGH** | Perlu validasi sebelum launch, bukan sesudahnya |
| Sales cycle panjang untuk institusi pendidikan | **MEDIUM** | Keputusan melibatkan yayasan, komite sekolah, dan kepsek |
| Resistensi perubahan dari operator sekolah | **MEDIUM** | Perlu program change management dan training intensif |
| Pesantren besar lebih memilih sistem custom | **MEDIUM** | Solusi: white-label dan kustomisasi modul |

### Risiko Teknis

| Risiko | Rating | Keterangan |
|--------|--------|-----------|
| Kompleksitas dual-mode (4 kombinasi institution type) meningkatkan bug risk | **MEDIUM** | Dimitigasi oleh strategy pattern, tapi butuh test coverage tinggi |
| Eventual consistency Vernon _data bisa membingungkan pengguna | **LOW** | Perlu UI indicator yang jelas saat data masih sinkronisasi |
| Vendor lock-in payment gateway (Midtrans/Xendit) | **LOW** | Arsitektur sudah menggunakan abstraction layer |

---

## 7. Metrik Keberhasilan (KPI Platform)

### KPI Bisnis

| Metrik | Target Tahun 1 | Target Tahun 2 |
|--------|---------------|---------------|
| Sekolah aktif berbayar | 20–30 | 70–100 |
| MRR | Rp 100 juta | Rp 400 juta |
| Churn rate bulanan | < 5% | < 3% |
| NPS (Net Promoter Score) | > 40 | > 55 |
| Free-to-paid conversion rate | 5% | 8% |
| Rata-rata waktu onboarding | < 2 minggu | < 1 minggu |

### KPI Produk

| Metrik | Target |
|--------|--------|
| Uptime platform | ≥ 99,5% (jam kerja) |
| Waktu respon API | < 200ms (p95) |
| Laporan regulasi tergenerate 100% otomatis | Dicapai sebelum Enterprise launch |
| Dapodik export validasi pass rate | > 95% |
| Transaksi keuangan zero manual error | 100% (enforced by system) |

### KPI Compliance

| Checkpoint | Deadline |
|-----------|---------|
| Consent flow UU PDP live | Q3 2026 |
| Dapodik export format validated | Q3 2026 |
| OJK review selesai | Q2 2026 |
| Legal review e-wallet vs PBI 20/6/2018 | Q2 2026 |
| ISO 27001 certification dimulai | Q4 2026 |

---

## 8. Rekomendasi Prioritas CEO

1. **Segera lakukan legal review e-wallet dan konsultasi OJK** — ini adalah hard blocker untuk Enterprise tier. Tanpa ini, ~50% dari proyeksi revenue tidak bisa dimonetisasi.

2. **Validasi harga ke 20–30 sekolah sebelum public launch** — asumsi Rp 5.000–8.000/siswa/bulan adalah proyeksi, bukan data. Jika asumsi ini salah, seluruh proyeksi revenue harus direvisi.

3. **Cari 3 pesantren flagship untuk early adopter** — credibility di komunitas pesantren adalah aset terpenting untuk go-to-market. Satu referensi dari kyai terkemuka bernilai lebih dari puluhan iklan digital.

4. **Pastikan compliance team dibentuk sebelum Q2 2026** — platform beroperasi di 6 jurisdiksi regulasi secara simultan. Ini bukan masalah teknis semata; butuh tim legal/compliance yang dedicated.

5. **Pertimbangkan pricing tier keempat** (Rp 2.000–3.000/siswa/bulan untuk sekolah mikro < 200 siswa) — ini memperlebar corong akuisisi tanpa mengurangi nilai di tier menengah-atas.
