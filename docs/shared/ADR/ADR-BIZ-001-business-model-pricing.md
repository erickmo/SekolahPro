# ADR-BIZ-001: Business Model & Pricing Strategy

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: Erick Mo, C-Suite Advisory Panel

## Context

SekolahPro adalah platform SaaS manajemen sekolah + koperasi/BMT yang menargetkan pesantren (pondok pesantren Islam) sebagai pasar primer dan sekolah umum sebagai pasar sekunder. Dari 95 ADR yang sudah dibuat, seluruhnya bersifat teknis (arsitektur, domain model, integrasi). Belum ada ADR yang mendokumentasikan keputusan bisnis fundamental: model harga, unit economics, dan revenue streams.

C-Suite CFO mengidentifikasi ini sebagai critical gap. Tanpa kejelasan pricing model, tim engineering tidak bisa membuat keputusan arsitektur yang tepat terkait:

1. **Feature gating** — fitur mana yang gratis vs berbayar
2. **Metering infrastructure** — apa yang perlu di-track untuk billing
3. **Multi-tenant cost allocation** — bagaimana menghitung COGS per tenant
4. **Compliance budget** — investasi regulasi yang diperlukan untuk fitur finansial

### Profil Pasar Target

| Segmen | Jumlah Institusi | Rata-rata Siswa | Karakteristik |
|--------|----------------:|----------------:|---------------|
| Pesantren besar | ~3.000 | 500-2.000 | Butuh sekolah + koperasi/BMT + asrama |
| Sekolah swasta Islam | ~15.000 | 200-800 | Butuh manajemen sekolah lengkap |
| Sekolah swasta umum | ~10.000 | 150-500 | Butuh manajemen sekolah standar |

## Decision

### 1. Model Pricing: Per-Siswa/Bulan dengan 3 Tier

Mengadopsi model **per-siswa/bulan** karena paling adil (sekolah kecil bayar lebih murah) dan memberikan revenue yang tumbuh seiring pertumbuhan sekolah.

| Aspek | Starter (Freemium) | Pro | Enterprise |
|-------|-------------------|-----|------------|
| **Harga** | Gratis | Rp 5.000-8.000/siswa/bulan | Rp 10.000-15.000/siswa/bulan |
| **Maks Siswa** | 100 | Unlimited | Unlimited |
| **Modul Sekolah** | Data siswa + absensi saja | S001-S058 (lengkap) | S001-S058 (lengkap) |
| **Modul Koperasi/BMT** | - | - | K001-K024 (lengkap) |
| **E-Wallet** | - | - | Ya |
| **Support** | Community forum | Email (SLA 24 jam) | Priority (SLA 4 jam) + dedicated CSM |
| **Custom Domain** | - | Add-on | Termasuk |
| **White Label** | - | - | Add-on |
| **Target Pasar** | Sekolah kecil (acquisition funnel) | Sekolah swasta menengah-besar | Pesantren besar |

### 2. Proyeksi Revenue

#### Skenario Tahun ke-2 (At Scale)

| Segmen | Jumlah Sekolah | Rata-rata Siswa | Harga/Siswa | MRR |
|--------|---------------:|----------------:|------------:|----:|
| Starter (free) | 200 | 80 | Rp 0 | Rp 0 |
| Pro | 50 | 500 | Rp 8.000 | Rp 200.000.000 |
| Enterprise | 20 | 800 | Rp 12.000 | Rp 192.000.000 |
| **Total** | **270** | | | **Rp 392.000.000/bulan** |

**ARR potensial**: ~Rp 4,7 miliar/tahun

#### Revenue Tambahan (Estimasi 15-25% dari MRR)

| Stream | Model | Estimasi/Bulan |
|--------|-------|---------------:|
| Payment gateway fee margin | Markup 0,1-0,3% di atas fee provider | Rp 15.000.000-40.000.000 |
| SMS/WhatsApp notifikasi overage | Per-pesan di atas kuota gratis | Rp 10.000.000-25.000.000 |
| Custom domain | Rp 50.000/bulan per domain | Rp 3.500.000 |
| White-label | Rp 500.000/bulan per institusi | Rp 5.000.000-10.000.000 |
| Implementasi & training | One-time, Rp 2-10 juta/sekolah | Rp 20.000.000-50.000.000 |
| Data analytics premium | Rp 100.000-300.000/bulan per sekolah | Rp 5.000.000-15.000.000 |

### 3. Break-even Analysis

| Komponen Biaya | Estimasi/Bulan |
|----------------|---------------:|
| Engineering team (5-8 orang) | Rp 150.000.000-200.000.000 |
| Infrastructure (cloud, DB, NATS) | Rp 30.000.000-50.000.000 |
| Support & CS team | Rp 30.000.000-40.000.000 |
| Marketing & sales | Rp 20.000.000-30.000.000 |
| Overhead (legal, accounting, office) | Rp 20.000.000-30.000.000 |
| **Total Burn Rate** | **Rp 250.000.000-350.000.000/bulan** |

**Break-even point**: 35-45 sekolah berbayar (campuran Pro + Enterprise), atau setara ~25.000-30.000 siswa aktif berbayar.

### 4. Biaya Regulasi & Compliance

| Item | Estimasi Biaya | Frekuensi |
|------|---------------:|-----------|
| Konsultasi OJK/BI (e-wallet & simpan-pinjam) | Rp 50.000.000-100.000.000 | One-time + annual review |
| Legal compliance koperasi/BMT | Rp 30.000.000-50.000.000 | One-time |
| BPJS integration untuk payroll | Rp 10.000.000-20.000.000 | One-time development |
| PPN 11% atas layanan SaaS | ~11% dari revenue | Bulanan |
| Sertifikasi keamanan data (ISO 27001) | Rp 100.000.000-200.000.000 | Annual |

**Catatan penting**: Fitur e-wallet dan simpan-pinjam (K001-K024) memerlukan koordinasi dengan OJK karena termasuk aktivitas jasa keuangan. Ini harus diselesaikan sebelum Enterprise tier diluncurkan.

### 5. Feature Gating Architecture Impact

Keputusan pricing ini memerlukan implementasi teknis berikut:

| Kebutuhan Teknis | ADR Terkait | Prioritas |
|-------------------|-------------|-----------|
| Tenant tier configuration di multi-tenant | ADR-004 | P0 |
| Feature flag per tier | Baru (perlu ADR) | P0 |
| Usage metering (jumlah siswa aktif) | Baru (perlu ADR) | P0 |
| Billing & invoice system | Baru (perlu ADR) | P1 |
| SMS/WhatsApp usage tracking | ADR-S044 | P1 |
| Payment gateway abstraction | ADR-K011 | P1 |

## Consequences

### Positif

- **Adil dan scalable** — harga proporsional dengan ukuran sekolah
- **Freemium sebagai funnel** — barrier masuk rendah, sekolah bisa coba dulu
- **Revenue predictable** — subscription bulanan memudahkan forecasting
- **Upsell natural** — sekolah yang tumbuh otomatis membayar lebih banyak
- **Multiple revenue streams** — tidak bergantung pada satu sumber pendapatan
- **Alignment teknis-bisnis** — engineering tahu persis fitur mana di tier mana

### Negatif

- **Freemium cost** — 200 sekolah gratis tetap memakan infrastructure cost
- **Kompleksitas feature gating** — perlu investasi engineering untuk metering dan billing
- **Regulatory risk** — fitur koperasi/BMT memerlukan compliance OJK yang mahal dan lama
- **Price sensitivity** — pasar pesantren/sekolah Indonesia sangat price-sensitive, perlu validasi harga
- **Churn risk** — jika sekolah merasa fitur gratis sudah cukup, conversion rate bisa rendah
- **PPN burden** — 11% PPN mengurangi margin, dan banyak sekolah belum terbiasa dengan SaaS berbayar

## Alternatives Considered

### A. Per-Module Pricing

Setiap modul (absensi, SPP, koperasi, dll) dijual terpisah.

- **Pro**: Fleksibel, sekolah hanya bayar yang dipakai
- **Kontra**: Terlalu kompleks untuk pasar target, decision fatigue, average revenue per user rendah
- **Alasan ditolak**: Pasar pesantren lebih nyaman dengan paket all-in-one

### B. Flat Fee per Sekolah

Harga tetap per sekolah tanpa memperhitungkan jumlah siswa (misal: Rp 2 juta/bulan).

- **Pro**: Sederhana, mudah dipahami
- **Kontra**: Tidak adil — sekolah 100 siswa bayar sama dengan 2.000 siswa
- **Alasan ditolak**: Barrier terlalu tinggi untuk sekolah kecil, terlalu murah untuk sekolah besar

### C. Transaction-Based Only

Revenue hanya dari fee transaksi keuangan (SPP, koperasi).

- **Pro**: Sekolah tidak perlu bayar subscription
- **Kontra**: Revenue tidak predictable, hanya bekerja untuk sekolah dengan volume transaksi tinggi
- **Alasan ditolak**: Tidak semua sekolah punya volume transaksi cukup, terlalu bergantung pada satu stream

## Action Items

1. **Validasi harga** — survey 20-30 sekolah target untuk willingness-to-pay
2. **Buat ADR teknis** — feature flag system dan usage metering
3. **Konsultasi OJK** — timeline dan persyaratan untuk fitur koperasi/BMT
4. **Define Starter scope** — finalisasi fitur gratis yang cukup menarik tapi mendorong upgrade
5. **Financial model detail** — buat spreadsheet dengan sensitivity analysis per tier

## References

- ADR S001-S058: Modul manajemen sekolah
- ADR K001-K024: Modul koperasi/BMT
- ADR-004: Multi-tenant 4-level hierarchy
- ADR-S044: Notification system
- ADR-K011: Transaksi koperasi
