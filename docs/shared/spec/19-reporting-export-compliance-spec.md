# 19 — Reporting Export Compliance Spec

**Status**: Draft  
**Perspective**: Compliance + Operations + Engineering  
**Primary Inputs**: `ADR-S054`, `ADR-S055`, `ADR-K017`, `ADR-K027`, `ADR-018`, `15-school-analytics-reporting-spec.md`

## Purpose

Mendefinisikan aturan untuk export, scheduled report, regulator-ready output, dan evidence handling agar reporting tidak hanya informatif, tetapi juga aman, traceable, dan sesuai kewajiban operasional/regulasi.

## Export Principles

- setiap export adalah tindakan yang dapat diaudit
- report output harus punya source traceability
- sensitive exports mengikuti access policy dan minimization rule
- scheduled export dan manual export mengikuti kontrol yang setara

## Export Categories

### Operational Exports
- PDF, XLSX, CSV untuk kebutuhan sekolah/koperasi internal
- bukti bayar, laporan kelas, laporan keuangan internal, rekap absensi

### Compliance / Regulator Exports
- Dapodik-oriented export
- koperasi regulatory / PPATK / dinas-related pack
- tax/payroll supporting output bila berlaku

### Executive Packs
- board summary, KPI deck, health indicator summary, exception report

## Control Requirements

- siapa meminta export, kapan, parameter apa, format apa, dan hasilnya ke mana dikirim harus tercatat
- file expiration/retention policy harus jelas
- watermarking, masking, atau role restriction diterapkan untuk data sensitif bila perlu
- export retry tidak boleh membuat ambiguity antara output lama dan baru

## Data Integrity Rules

- exported numbers harus dapat ditelusuri ke template/query/source snapshot
- jika report berasal dari materialized view atau precomputed source, timestamp freshness harus diketahui
- correction atau re-issue harus menandai versi/output sebelumnya bila relevan

## Acceptance Criteria

- `AC-FUNC`: stakeholder dapat menghasilkan export yang diperlukan tanpa proses manual berisiko tinggi
- `AC-AUTH`: hanya role yang sah dapat menghasilkan dan mengakses export tertentu
- `AC-DATA`: isi export traceable, versioned bila perlu, dan konsisten dengan source reporting
- `AC-AUDIT`: export history lengkap tersedia untuk review/forensik
- `AC-INT`: regulator-oriented export dapat disiapkan, divalidasi, dan diulang bila ada koreksi
- `AC-NFR`: proses export berat tidak mengganggu operasional utama secara signifikan

## Open Questions

- masa retensi default per jenis export
- kapan signed export / digitally sealed document menjadi wajib
