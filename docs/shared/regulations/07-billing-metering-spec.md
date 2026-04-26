# 07 — Billing & Metering Spec

**Status**: Draft  
**Perspective**: C-Level + System Analyst + System Engineer  
**Primary Inputs**: `ADR-BIZ-001`, `ADR-004`, `ADR-S044`, `ADR-S051`, `ADR-K011`, `ADR-018`

## Purpose

Mendefinisikan model billing, metering, feature gating, invoice, tax, dan revenue control agar packaging Starter / Pro / Enterprise dapat dioperasionalkan.

## Commercial Model

### Tiers
- **Starter**: funnel akuisisi, batas siswa aktif, fitur dasar
- **Pro**: school management lengkap
- **Enterprise**: school + koperasi/BMT + pesantren premium capabilities

### Metering Drivers
- jumlah siswa aktif per tenant/company
- penggunaan channel notifikasi berbayar (SMS/WhatsApp)
- add-on custom domain / white-label / premium analytics
- layanan implementasi dan training sebagai non-recurring charge

## Capability Scope

### 1. Tenant Packaging
- setiap tenant memiliki tier aktif
- feature flag harus bisa dibaca lintas app dan service
- perubahan tier harus memiliki effective date dan audit trail

### 2. Usage Metering
- definisi `active student` harus eksplisit dan konsisten
- snapshot usage dihitung periodik per billing cycle
- usage outlier harus bisa direview dan dikoreksi secara auditable

### 3. Invoicing & Tax
- invoice bulanan per tenant/company
- PPN 11% sebagai komponen wajib saat applicable
- kredit/debit adjustment tercatat sebagai ledger billing

### 4. Collection & Payment State
- unpaid, partially_paid, paid, overdue, disputed
- payment method dapat berasal dari gateway eksternal atau collection manual
- finance ops harus bisa reconcile invoice terhadap payment event

### 5. Overage & Add-ons
- notifikasi berbayar dihitung per channel dan period
- add-on memiliki entitlement dan expiry yang jelas
- overage tidak boleh mengganggu fungsi core tanpa aturan throttle yang disetujui

## Data Objects

- billing_account
- subscription
- feature_entitlement
- usage_snapshot
- usage_adjustment
- invoice
- invoice_line
- payment_record
- tax_record
- billing_audit_log

## Business Rules

- tenant tidak boleh mengakses capability di luar entitlement aktif
- perubahan tier tidak menghapus histori invoice dan usage
- invoice generation harus idempotent per billing period
- semua angka billing harus bisa ditelusuri ke source usage
- write-off, refund, dan adjustment wajib lewat approval flow

## Acceptance Criteria

- `AC-FUNC`: tier, metering, invoice, dan collection berjalan end-to-end
- `AC-DATA`: satu period tidak menghasilkan invoice ganda untuk subscription yang sama
- `AC-AUTH`: hanya role billing/finance tertentu yang dapat melakukan override
- `AC-AUDIT`: setiap perubahan usage, invoice, atau entitlement tercatat
- `AC-INT`: payment callback dan notification usage sinkron dan dapat direkonsiliasi
- `AC-NFR`: invoice generation skala multi-tenant dapat berjalan batch tanpa duplikasi

## Open Design Decisions

- definisi final active student per tier
- apakah billing account berada di level tenant atau company
- kapan throttling berlaku untuk overage notification channels
- apakah koperasi modules di-bundle penuh atau per package enterprise sub-tier
