# 08 — Compliance & Risk Controls

**Status**: Draft  
**Perspective**: C-Level + Risk + Engineering  
**Primary Inputs**: `ADR-018`, `ADR-004`, `ADR-013`, `ADR-S031`, `ADR-S051`, koperasi compliance ADRs

## Purpose

Menyatukan kewajiban compliance lintas pendidikan, data privacy, financial operations, tax, dan audit ke dalam satu operating spec yang dapat dieksekusi.

## Compliance Domains

### 1. Data Privacy (UU PDP)
- consent wali untuk data anak
- retention policy per tenant
- right to erasure dan data portability
- breach notification dalam SLA regulasi

### 2. Education Reporting
- kesiapan data untuk Dapodik
- validitas identifier seperti NISN, NPSN, NUPTK
- audit trail perubahan data master pendidikan

### 3. Cooperative & Financial Governance
- SHU, RAT, laporan koperasi
- threshold awareness untuk exposure OJK
- segregation of duties untuk transaksi dan approval

### 4. Payments & Wallet Guardrails
- spending interface, bukan e-money terbuka
- tidak ada P2P transfer
- tidak ada cash-out ke rekening bank
- top-up dan penggunaan dibatasi rule institusi

### 5. Tax & Payroll Controls
- PPN atas SaaS billing
- PPh/BPJS untuk payroll saat berlaku
- bukti potong dan dokumen pendukung dapat ditelusuri

## Control Types

### Preventive
- RBAC dan scope enforcement
- feature gate untuk capability regulated
- validation rules pada onboarding, payment, payroll, dan koperasi

### Detective
- anomaly alerts
- audit log review
- suspicious access dan suspicious transaction checks
- reconciliation reports

### Corrective
- freeze tenant/module access bila terjadi pelanggaran berat
- rollback / correction workflow yang auditable
- incident response dan breach playbook

## Governance Rules

- semua capability regulated wajib memiliki owner bisnis dan owner teknis
- tidak ada go-live financial module tanpa compliance readiness review
- semua exception harus memiliki expiry date dan approver
- legal interpretation yang mempengaruhi arsitektur wajib dicatat sebagai decision reference

## Acceptance Criteria

- `AC-FUNC`: consent, retention, deletion, portability, breach response punya flow operasional
- `AC-AUTH`: akses PII dan data finansial dibatasi dan diaudit
- `AC-DATA`: retention scheduler, anonymization, dan audit trail konsisten
- `AC-AUDIT`: bukti kepatuhan dapat dihasilkan per tenant dan periode
- `AC-INT`: regulator/reporting export dapat diverifikasi format dan completeness-nya
- `AC-NFR`: control checks tidak merusak availability modul inti secara berlebihan

## High Risks

- child data misuse
- regulated finance scope creep
- undocumented operator overrides
- missing breach process
- tax mismatch antara invoice dan pembayaran
