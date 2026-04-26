# 01 — Portfolio Thesis

**Status**: Draft  
**Perspective**: C-Level  
**Primary Inputs**: `ADR-BIZ-001`, `ADR-003`, `ADR-004`, `ADR-009`, `ADR-018`, `SPRINT-ORDER.md`

## Strategic Positioning

SekolahPro adalah platform SaaS hybrid untuk:

1. **Management Sekolah** sebagai acquisition engine
2. **Management Koperasi Sekolah / BMT** sebagai monetization expansion
3. **Pesantren / boarding school** sebagai diferensiasi utama

## Ideal Customer Profile

### Primary
- pesantren besar
- sekolah Islam swasta
- institusi dengan kebutuhan sekolah + koperasi + asrama

### Secondary
- sekolah swasta umum yang membutuhkan school management lengkap

## Commercial Packaging Direction

### Starter
- fokus pada kebutuhan dasar sekolah kecil
- dipakai sebagai funnel akuisisi

### Pro
- school management lengkap
- target monetisasi utama untuk sekolah non-koperasi

### Enterprise
- school + koperasi/BMT + kebutuhan pesantren premium
- target ARPU tertinggi

## Business Waves

### Wave 1 — School Core Monetization
Prioritas capability:
- tenant/auth/RBAC
- master data shared
- student core
- curriculum dan academic operations
- attendance, grades, rapor
- SPP dan payment readiness

**Success metrics**:
- tenant aktif berbayar
- aktivasi sekolah cepat
- retensi semester pertama

### Wave 2 — Cooperative Core Expansion
Prioritas capability:
- member/nasabah
- rekening
- produk/akad
- simpanan dan pembiayaan
- transaksi
- jurnal, SHU, regulatory reporting

**Success metrics**:
- attach rate Enterprise
- revenue per tenant naik
- operasional koperasi berjalan tanpa spreadsheet manual

### Wave 3 — Pesantren Premium Moat
Prioritas capability:
- dormitory
- canteen
- parent/member experience
- komunikasi, notifikasi, approval
- analytics dan eksekutif insight

**Success metrics**:
- diferensiasi kompetitif
- cross-module adoption tinggi
- churn rendah pada pesantren besar

## Executive Risks

### Portfolio sprawl
Mitigasi: semua module non-wave harus melewati governance gate.

### Regulatory exposure
Mitigasi: fitur keuangan dan wallet dibatasi oleh compliance dan legal readiness.

### Monetization gap
Mitigasi: packaging, feature gating, metering, billing, tax control harus dipersiapkan sebelum GTM scale.

### Onboarding failure
Mitigasi: migration, import, training, activation KPI menjadi bagian dari spec wajib.

## Executive Decisions Required

- finalisasi Starter / Pro / Enterprise boundaries
- persetujuan urutan wave delivery
- kebijakan launch untuk fitur koperasi/BMT regulated
- batas investasi sebelum product-market validation selesai
