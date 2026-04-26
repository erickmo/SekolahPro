# 37 — Sprint Plan from Specs

**Status**: Draft  
**Perspective**: C-Level + PMO + System Analyst + Engineering  
**Primary Inputs**: `docs/adr/SPRINT-ORDER.md`, `docs/spec/04-school-platform-spec.md`, `docs/spec/05-cooperative-platform-spec.md`, `docs/spec/06-traceability-matrix.md`

## Purpose

Menerjemahkan ADR ordering dan seluruh spec operasional menjadi sprint plan delivery yang bisa dipakai untuk eksekusi roadmap secara terkontrol.

## Planning Principles

- sprint mengikuti dependency kritikal dari `SPRINT-ORDER.md`
- sprint scope dibatasi pada capability yang bisa selesai end-to-end
- shared governance, ops, dan compliance artifacts masuk sprint yang relevan, bukan ditunda ke akhir
- school core tetap menjadi jalur monetization utama; koperasi berjalan paralel secara selektif

## Sprint Structure

Setiap sprint wajib punya:
- tujuan bisnis
- scope capability
- spec references
- exit criteria
- operational readiness items
- open risks

---

## Sprint 0 — Platform & Delivery Foundation

**Objective**: memastikan landasan arsitektur, kontrak, release safety, dan operability siap untuk sprint bisnis.

### Scope
- platform foundation dari Fase 0
- guardrail shared spec awal

### Spec References
- `00-spec-governance.md`
- `03-cross-cutting-foundation.md`
- `11-observability-deployment-spec.md`
- `12-api-event-contract-spec.md`
- `16-release-readiness-gates.md`
- `29-cross-domain-event-catalog.md`

### Exit Criteria
- event/API contract baseline ada
- observability minimal aktif
- release gate dan config change baseline siap
- shared foundation dapat dipakai sprint domain berikutnya

---

## Sprint 1 — Auth, Tenant, Access, Baseline Ops

**Objective**: mengaktifkan tenant-safe operation dan baseline configuration untuk semua domain.

### Scope
- auth + multi-tenant + RBAC
- tenant baseline config
- activation and support operating baseline

### ADR Alignment
- Fase 1

### Spec References
- `01-portfolio-thesis.md`
- `09-onboarding-migration-spec.md`
- `20-tenant-activation-ops-playbook.md`
- `21-feature-gating-entitlement-model.md`
- `25-tenant-support-hypercare-playbook.md`
- `28-operational-sla-escalation-matrix.md`
- `30-pilot-tenant-rollout-governance.md`
- `31-tenant-configuration-baseline-spec.md`
- `34-tenant-ops-kpi-scorecard.md`
- `36-configuration-change-control-policy.md`

### Exit Criteria
- login, tenant scope, role enforcement stabil
- tenant baru bisa dibuat dari baseline repeatable
- activation + hypercare playbook siap dipakai pilot tenant

---

## Sprint 2 — Master Data Foundation + Cooperative Core Foundation

**Objective**: menyelesaikan master data sekolah dan core koperasi yang menjadi fondasi wave domain.

### Scope
- academic years, teachers/staff, class rooms
- member, product/akad, rekening koperasi
- master data governance dan blast-radius control

### ADR Alignment
- Fase 2 + Fase 3

### Spec References
- `02-capability-map.md`
- `05-cooperative-platform-spec.md`
- `17-master-data-governance-spec.md`
- `22-master-data-change-impact-matrix.md`

### Exit Criteria
- master data sekolah usable end-to-end
- core koperasi usable untuk pembukaan rekening dan setup produk
- governance untuk perubahan master data berjalan

---

## Sprint 3 — Student Core + Savings Core

**Objective**: membangun single source of truth siswa dan produk simpanan koperasi dasar.

### Scope
- student lifecycle lengkap
- placement, guardian, PPDB, document management
- simpanan pokok/wajib, tabungan, deposito

### ADR Alignment
- Fase 4A + Fase 4B

### Spec References
- `04-school-platform-spec.md`
- `05-cooperative-platform-spec.md`
- `09-onboarding-migration-spec.md`

### Exit Criteria
- sekolah dapat mengelola siswa dari intake sampai placement
- koperasi dapat membuka dan mengoperasikan simpanan dasar

---

## Sprint 4 — Curriculum, Scheduling, Financing Foundation

**Objective**: mengaktifkan operasi akademik inti sekolah dan pembiayaan koperasi.

### Scope
- curriculum, subject, calendar, schedule, lesson plan, teaching journal
- pinjaman/pembiayaan, angsuran, denda, agunan

### ADR Alignment
- Fase 5A + Fase 5B

### Spec References
- `04-school-platform-spec.md`
- `05-cooperative-platform-spec.md`
- `18-approval-workflow-orchestration-spec.md` (prep jika ada approval need)

### Exit Criteria
- jadwal belajar dan proses ajar siap dipakai operasional
- pengajuan dan pengelolaan pembiayaan koperasi siap dipakai terbatas

---

## Sprint 5 — Attendance, Grades, Student Services + Transaction Engine

**Objective**: mengaktifkan daily school operations dan engine transaksi koperasi.

### Scope
- attendance, academic record, grades, exams
- health, discipline, achievement, extracurricular
- rekening/non-rekening transaction, teller, denomination, kas

### ADR Alignment
- Fase 6A + Fase 6B

### Spec References
- `04-school-platform-spec.md`
- `05-cooperative-platform-spec.md`
- `23-financial-reconciliation-ops-spec.md`
- `33-financial-exception-resolution-playbook.md`

### Exit Criteria
- sekolah dapat menjalankan operasi akademik harian
- koperasi memiliki transaction engine yang dapat direkonsiliasi

---

## Sprint 6 — Rapor, SPP, HR School + Cooperative Accounting

**Objective**: menyelesaikan keluaran akademik utama dan accounting koperasi.

### Scope
- rapor generation, SPP baseline, BK
- teacher attendance, workload, leave, substitution
- jurnal/COA, SHU, laporan regulasi, zakat/infaq

### ADR Alignment
- Fase 7A + Fase 7B

### Spec References
- `04-school-platform-spec.md`
- `05-cooperative-platform-spec.md`
- `14-koperasi-regulated-finance-guardrails.md`
- `19-reporting-export-compliance-spec.md`
- `26-payment-fee-absorption-policy.md`

### Exit Criteria
- sekolah dapat menutup siklus akademik semester dasar
- koperasi memiliki accounting backbone dan report baseline

---

## Sprint 7 — Facilities, Payroll, Analytics, Reporting

**Objective**: memperluas operasional sekolah dan membangun reporting yang stabil.

### Scope
- library, lab, asset, booking, PKG, PKB, payroll
- analytics and reporting baseline
- report freshness and export governance

### ADR Alignment
- Fase 8 + sebagian growth/reporting readiness

### Spec References
- `15-school-analytics-reporting-spec.md`
- `19-reporting-export-compliance-spec.md`
- `32-report-refresh-data-freshness-policy.md`

### Exit Criteria
- reporting utama tersedia per role
- export dan freshness policy dapat dioperasionalkan
- fasilitas dan payroll school baseline siap

---

## Sprint 8 — Boarding, Canteen, Cooperative Extensions

**Objective**: membuka diferensiasi pesantren dan extension koperasi non-core.

### Scope
- dormitory management
- canteen and billing
- koperasi toko, payroll deduction, e-wallet guardrail review

### ADR Alignment
- Fase 9 + Fase 10

### Spec References
- `04-school-platform-spec.md`
- `05-cooperative-platform-spec.md`
- `14-koperasi-regulated-finance-guardrails.md`

### Exit Criteria
- fitur boarding/canteen siap pilot untuk tenant yang sesuai
- extension koperasi hanya berjalan bila guardrail terpenuhi

---

## Sprint 9 — Communication, Portal, Stakeholder Experience

**Objective**: membangun engagement layer untuk parent, student, dan member.

### Scope
- notification system
- messaging
- announcement
- parent portal
- koperasi portal and notification
- broadcast governance

### ADR Alignment
- Fase 11

### Spec References
- `10-parent-student-member-experience.md`
- `24-notification-messaging-operating-model.md`
- `35-broadcast-communication-governance.md`

### Exit Criteria
- external-facing experience siap untuk pilot tenant
- komunikasi high-reach dapat dikontrol dan diaudit

---

## Sprint 10 — Administration, Approval, Payments, Budgeting

**Objective**: menyelesaikan control-plane administrasi sekolah dan payment modernization.

### Scope
- correspondence
- approval workflow engine
- school profile/accreditation
- committee
- RKAS
- payment gateway

### ADR Alignment
- Fase 12

### Spec References
- `18-approval-workflow-orchestration-spec.md`
- `23-financial-reconciliation-ops-spec.md`
- `26-payment-fee-absorption-policy.md`
- `33-financial-exception-resolution-playbook.md`

### Exit Criteria
- approval reusable lintas domain berjalan
- payment gateway dapat direkonsiliasi dan diaudit
- administrative control plane sekolah siap scale

---

## Sprint 11 — Dapodik, Integrations, Governance Hardening

**Objective**: menutup gap integrasi dan hardening untuk rollout lebih luas.

### Scope
- Dapodik operator workflow
- reporting/export compliance hardening
- pilot rollout governance review
- support/ops KPI review
- cross-domain event governance tightening

### Spec References
- `27-dapodik-operator-workflow-spec.md`
- `30-pilot-tenant-rollout-governance.md`
- `34-tenant-ops-kpi-scorecard.md`
- `29-cross-domain-event-catalog.md`

### Exit Criteria
- operator Dapodik punya workflow stabil
- pilot evidence cukup untuk keputusan scale-out
- governance artifacts siap untuk multi-tenant growth

---

## Recommended Delivery Waves

### Wave 1 — Revenue-First School Core
- Sprint 0-6

### Wave 2 — Cooperative Enterprise Expansion
- Sprint 2-6 berjalan paralel selektif, lalu Sprint 8/10 untuk extension

### Wave 3 — Premium Experience & Institutional Control Plane
- Sprint 7-11

## Detailed Follow-Up Documents

- `38-sprint-0-3-detailed-task-packs.md`
- `39-sprint-4-7-detailed-task-packs.md`
- `40-sprint-8-11-detailed-task-packs.md`
- `41-sprint-backlog-summary-lane-view.md`

## Notes

- sprint numbering di atas adalah planning sprint, bukan pengganti `docs/adr/SPRINT-ORDER.md`
- beberapa sprint dapat dipecah menjadi sub-sprint jika kapasitas tim kecil
- spec governance, release gates, config control, dan ops KPI berlaku lintas semua sprint
