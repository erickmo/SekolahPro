# ADR-018: Regulatory Compliance Matrix

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: Erick Mo, CFO Review Panel

## Context

SekolahPro beroperasi di **persimpangan dua sektor teregulasi**: pendidikan dan jasa keuangan di Indonesia. Sebagai platform SaaS yang mengelola data siswa sekaligus operasional koperasi/BMT sekolah, sistem harus patuh terhadap **minimal 6 regulator** secara simultan:

1. **Kemendikbud** — Dapodik, NISN, NUPTK, NPSN, pelaporan semester.
2. **Dinas Koperasi** — Registrasi koperasi, pelaporan RAT, perhitungan SHU, modal minimum.
3. **OJK** — Berlaku jika koperasi mencapai threshold aset/simpanan tertentu (>10 miliar).
4. **Bank Indonesia** — Regulasi e-wallet (PBI 20/6/2018), QRIS, payment aggregator.
5. **UU PDP (UU No. 27/2022)** — Perlindungan data pribadi, khususnya data anak di bawah umur.
6. **DJP (Pajak)** — PPN 11% pada langganan SaaS, PPh withholding untuk payroll.

Ketidakpatuhan terhadap salah satu regulator dapat mengakibatkan **denda, pencabutan izin, atau tuntutan hukum**. Oleh karena itu, compliance bukan fitur opsional — ia adalah **constraint arsitektur** yang mempengaruhi desain setiap modul.

## Decision

### Compliance Matrix per Modul

| Modul | Regulator | Kewajiban | Dampak pada Sistem | Prioritas |
|-------|-----------|-----------|---------------------|-----------|
| Student Data | UU PDP | Consent data anak, persetujuan wali | Consent flow saat onboarding | **Critical** |
| Dapodik | Kemendikbud | NPSN, NISN, pelaporan semester | Format ekspor sesuai spesifikasi Dapodik | High |
| Koperasi | Dinas Koperasi | Laporan RAT, perhitungan SHU, modal minimum | Lihat ADR K016, K017 | High |
| E-wallet | Bank Indonesia | Analisis PBI 20/6/2018 | Harus tetap "spending interface", bukan uang elektronik | **Critical** |
| Payment | Bank Indonesia | Aturan MDR QRIS, aturan payment aggregator | Model fee S051 | High |
| Payroll | BPJS | Potongan BPJS Kesehatan + Ketenagakerjaan | Kalkulasi payroll S031 | Medium |
| SaaS Billing | DJP | PPN 11% pada langganan SaaS | Harga harus sudah termasuk PPN | Medium |

### E-wallet: Batasan Regulasi Bank Indonesia

SekolahPro e-wallet **bukan** uang elektronik (e-money). Sistem dirancang sebagai **spending interface** — saldo hanya bisa digunakan untuk transaksi internal (kantin, koperasi, pembayaran SPP). Tidak ada:
- Transfer antar pengguna (P2P transfer).
- Cash-out / pencairan saldo ke rekening bank.
- Top-up dari sumber selain yang disetujui sekolah.

Dengan pembatasan ini, sistem tidak memerlukan lisensi e-money dari Bank Indonesia.

## Data Protection (UU PDP)

Karena mayoritas pengguna adalah **siswa di bawah 18 tahun**, UU PDP memberikan perlindungan ekstra:

### Keputusan Spesifik

1. **Consent wali wajib** — Semua data siswa dikategorikan sebagai data anak. Proses onboarding harus menyertakan persetujuan wali/orang tua sebelum data diproses.

2. **Retensi data** — Maksimal **5 tahun** setelah siswa lulus/keluar (konfigurabel per institusi). Setelah itu, data di-anonymize atau dihapus secara otomatis via scheduled job.

3. **Hak penghapusan (right to erasure)** — Sistem harus mendukung **hard delete** atas permintaan wali. Implementasi:
   - Endpoint `DELETE /api/v1/students/{id}/personal-data`
   - Cascade delete ke semua tabel terkait
   - Audit log tetap disimpan (tanpa PII)

4. **Portabilitas data** — Ekspor seluruh data siswa dalam format standar (JSON/CSV) melalui endpoint khusus. Target: selesai dalam < 72 jam setelah permintaan.

5. **Notifikasi pelanggaran data** — Jika terjadi data breach, notifikasi wajib dikirim ke:
   - Subjek data (wali/orang tua) — dalam **72 jam**
   - Otoritas PDP — dalam **72 jam**
   - Implementasi: alert system + template notifikasi bawaan

### Implementasi Teknis UU PDP

```
┌─────────────────────────────────────────┐
│           Consent Service               │
│  - Simpan record consent per siswa      │
│  - Versioning (consent v1, v2, ...)     │
│  - Consent withdrawal = soft-disable    │
├─────────────────────────────────────────┤
│         Data Lifecycle Service          │
│  - Retention policy per tenant          │
│  - Auto-anonymize setelah N tahun       │
│  - Hard delete on request               │
├─────────────────────────────────────────┤
│         Breach Notification Service     │
│  - Alert pipeline                       │
│  - Template notifikasi (email/SMS)      │
│  - Audit trail setiap notifikasi        │
└─────────────────────────────────────────┘
```

## Action Items

| # | Aksi | Owner | Deadline | Status |
|---|------|-------|----------|--------|
| 1 | Implementasi consent flow di onboarding siswa | Engineering | Q3 2026 | Planned |
| 2 | Buat format ekspor Dapodik (NISN, NPSN) | Engineering | Q3 2026 | Planned |
| 3 | Review legal: e-wallet vs PBI 20/6/2018 | CFO Panel + Legal | Q2 2026 | In Progress |
| 4 | Implementasi data retention scheduler | Engineering | Q4 2026 | Planned |
| 5 | Endpoint hard delete + data portability | Engineering | Q4 2026 | Planned |
| 6 | Breach notification pipeline | Engineering + Ops | Q4 2026 | Planned |
| 7 | Integrasi kalkulasi PPN 11% di billing | Engineering | Q3 2026 | Planned |
| 8 | Validasi SHU & RAT report vs template Dinas Koperasi | Engineering | Q3 2026 | Planned |

## Consequences

### Positif

- **Kepastian hukum** — Setiap modul memiliki peta regulasi yang jelas, mengurangi risiko pelanggaran.
- **Trust dari sekolah** — Kepatuhan UU PDP menjadi nilai jual, terutama untuk sekolah yang mengelola data ribuan siswa.
- **Skalabilitas bisnis** — E-wallet yang dirancang sebagai spending interface menghindari kebutuhan lisensi e-money yang mahal dan kompleks.
- **Audit-ready** — Sistem siap diaudit oleh regulator manapun karena compliance sudah by-design.

### Negatif

- **Kompleksitas onboarding** — Consent flow menambah langkah bagi orang tua/wali saat pendaftaran.
- **Biaya development** — Fitur hard delete, data portability, dan breach notification memerlukan effort signifikan.
- **Fitur e-wallet terbatas** — Pembatasan regulasi membuat e-wallet kurang fleksibel dibanding kompetitor fintech.

## Alternatives Considered

### 1. Abaikan Compliance, Perbaiki Nanti

Ditolak. Risiko hukum terlalu tinggi, terutama untuk data anak (UU PDP) dan regulasi keuangan (BI). Biaya remediasi pasca-pelanggaran jauh lebih besar daripada compliance by-design.

### 2. Gunakan Third-Party Payment Provider Sepenuhnya

Dipertimbangkan untuk payment gateway (QRIS, VA). Namun e-wallet internal tetap diperlukan untuk transaksi mikro di lingkungan sekolah (kantin, fotokopi). Solusi: **hybrid** — third-party untuk payment gateway, internal untuk spending interface.

### 3. Pisahkan Koperasi Menjadi Aplikasi Terpisah

Ditolak. Integrasi koperasi-sekolah adalah **value proposition utama** SekolahPro. Memisahkan keduanya menghilangkan sinergi data (siswa = anggota koperasi) dan menambah friction bagi pengguna.

## References

- UU No. 27 Tahun 2022 tentang Perlindungan Data Pribadi
- PBI No. 20/6/PBI/2018 tentang Uang Elektronik
- Permendikbud tentang Dapodik
- UU No. 25 Tahun 1992 tentang Perkoperasian
- ADR K016, K017 (Koperasi compliance)
- ADR S031 (Payroll), S051 (Payment fee model)
