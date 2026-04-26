# 09 — Onboarding & Migration Spec

**Status**: Draft  
**Perspective**: COO + System Analyst + Engineering  
**Primary Inputs**: `ADR-017`, `ADR-004`, `ADR-009`, `ADR-010`, `ADR-011`, `ADR-012`, `ADR-013`

## Purpose

Mendefinisikan alur aktivasi tenant dari setup awal sampai sekolah/koperasi dapat live dengan data yang tervalidasi.

## Onboarding Goals

- time-to-value singkat
- migrasi data tanpa input manual masif
- risiko data corrupt rendah
- aktivasi tenant dapat dipantau dan diukur

## Lifecycle

### 1. Tenant Setup
- buat tenant dan company
- pilih institution type dan mode institusi
- konfigurasi tahun ajaran aktif
- buat admin awal dan role dasar

### 2. Foundation Readiness
- teachers/staff tersedia
- class rooms tersedia
- actor mapping dan access model siap
- billing/package aktif bila tenant berbayar

### 3. Data Migration
- download template
- upload file
- dry-run validation
- error correction
- confirm import
- async import + verification
- rollback window bila diperlukan

### 4. Operational Activation
- sekolah menjalankan proses inti pertama
- verifikasi dashboard, laporan, dan role access
- training admin/operator selesai
- sign-off activation

## Supported Migration Domains

- students
- guardians
- teachers/staff
- academic years
- class rooms
- fee types / invoices
- cooperative members/accounts pada wave lanjutan

## Migration Rules

- semua import wajib dry-run sebelum write
- conflict resolution harus dipilih eksplisit
- import besar diproses async dan bisa dipantau
- rollback dibatasi waktu dan ruang lingkupnya
- post-import sync/rebuild tidak boleh silent-fail

## Activation KPIs

- waktu dari signup ke first successful import
- waktu dari first import ke first operational transaction
- error rate per template/entity
- activation completion rate per tenant cohort

## Acceptance Criteria

- `AC-FUNC`: tenant dapat onboard dari kosong ke siap operasi lewat flow terarah
- `AC-DATA`: hasil import tervalidasi, dapat diverifikasi, dan dapat di-rollback sesuai aturan
- `AC-AUTH`: hanya actor yang berwenang dapat upload, confirm, rollback
- `AC-AUDIT`: semua batch import dan keputusan conflict resolution tercatat
- `AC-INT`: post-import sync dan report verification menghasilkan status jelas
- `AC-NFR`: import ribuan baris dapat berjalan tanpa timeout user-facing

## Open Questions

- cakupan template MVP vs layanan migrasi premium
- kebutuhan mapping tool antar format sistem lama
- activation checklist untuk tenant school-only vs hybrid school+koperasi
