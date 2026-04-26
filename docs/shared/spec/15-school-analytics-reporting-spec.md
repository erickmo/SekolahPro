# 15 — School Analytics & Reporting Spec

**Status**: Draft  
**Perspective**: Headmaster + System Analyst + Data/Engineering  
**Primary Inputs**: `ADR-S054`, `ADR-S055`, `04-school-platform-spec.md`, `11-observability-deployment-spec.md`

## Purpose

Mendefinisikan model reporting dan analytics sekolah untuk kebutuhan operasional, eksekutif, guru, bendahara, wali kelas, dan kebutuhan ekspor/pelaporan eksternal.

## Reporting Domains

### 1. Academic Reporting
- rata-rata nilai per kelas/mapel
- ranking dan trend performa siswa
- analisis kelulusan / mastery
- rapor support dan semester comparison

### 2. Attendance Reporting
- rekap hadir/sakit/izin/alpha
- trend attendance per kelas/periode
- alert untuk chronic absenteeism

### 3. Finance Reporting
- SPP billed vs paid vs outstanding
- aging tunggakan
- pembayaran per periode/kelas/fee type
- readiness untuk rekonsiliasi payment gateway/manual collection

### 4. Executive Dashboard
- jumlah siswa aktif, guru aktif, kelas, tunggakan
- KPI operasional sekolah
- alert widgets untuk area bermasalah

### 5. Compliance & External Reporting
- Dapodik pre-validation views
- export-ready datasets
- report pack untuk akreditasi dan kebutuhan dinas tertentu

## Consumer Roles

- kepala sekolah / yayasan
- wakasek
- guru / wali kelas
- bendahara
- tata usaha
- operator Dapodik
- orang tua / siswa untuk subset personal views

## Data Strategy

- gunakan materialized views / precomputed summary untuk query berat
- ad-hoc report builder dibatasi dengan governance agar tidak menjadi backdoor query liar
- export history wajib tercatat
- role-based access wajib diterapkan sampai level report template dan data slice

## Dapodik Alignment

- reporting internal harus membantu operator memperbaiki data sebelum export
- validation mismatch, missing NISN/NUPTK/NPSN, dan error status harus terlihat jelas
- `manual_export` adalah baseline operasional phase awal

## Acceptance Criteria

- `AC-FUNC`: stakeholder utama dapat menghasilkan insight dan laporan inti tanpa olah manual besar
- `AC-AUTH`: user hanya melihat laporan sesuai scope dan kebutuhan peran
- `AC-DATA`: angka dashboard dan export dapat ditelusuri ke source summary/source domain
- `AC-AUDIT`: export, scheduled report, dan report template changes tercatat
- `AC-INT`: Dapodik-oriented report dan export dapat dipakai untuk proses operator dengan error handling jelas
- `AC-NFR`: dashboard dan laporan yang sering dipakai tetap responsif pada data sekolah besar

## Open Questions

- batas custom report builder self-service vs curated templates
- cadence refresh materialized views per kategori report
- laporan standar minimum untuk setiap role di wave awal
