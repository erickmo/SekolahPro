# SPECS

Direktori ini berisi dokumen spesifikasi operasional yang menerjemahkan ADR menjadi artefak delivery yang bisa dipakai oleh C-level, system analyst, dan system engineer.

## Tujuan

- Menjembatani keputusan arsitektur (`docs/adr`) dengan implementasi
- Menjadi sumber kebenaran untuk scope, aktor, proses, kontrak, acceptance criteria, dan sequencing
- Menyediakan traceability: `ADR -> spec -> rollout slice -> acceptance -> implementation`

## Prinsip

- **ADR tetap menjadi decision record**; spec mengoperasionalkan keputusan tersebut
- **Spec disusun per capability**, bukan satu file per ADR
- **Sprint-aligned** mengikuti `docs/adr/SPRINT-ORDER.md`
- **Traceable** ke ADR core, sekolah, dan koperasi
- **Business-first, actor-first, engineering-ready**

## Struktur

- `00-spec-governance.md` — aturan main dokumen spec
- `01-portfolio-thesis.md` — arah produk, ICP, packaging, dan prioritas eksekutif
- `02-capability-map.md` — capability map lintas core, sekolah, koperasi
- `03-cross-cutting-foundation.md` — tenant, RBAC, compliance, data, observability, deployment baseline
- `04-school-platform-spec.md` — spesifikasi capability Management Sekolah
- `05-cooperative-platform-spec.md` — spesifikasi capability Management Koperasi Sekolah
- `06-traceability-matrix.md` — pemetaan ADR ke spec dan rollout
- `07-billing-metering-spec.md` — pricing operations, metering, invoice, entitlement
- `08-compliance-risk-controls.md` — kontrol compliance, risk, dan governance
- `09-onboarding-migration-spec.md` — aktivasi tenant dan strategi migrasi data
- `10-parent-student-member-experience.md` — pengalaman parent, student, dan member
- `11-observability-deployment-spec.md` — deployment, monitoring, dan operability baseline
- `12-api-event-contract-spec.md` — aturan kontrak API dan event lintas sistem
- `13-security-privacy-data-lifecycle-spec.md` — keamanan, privasi, dan lifecycle data
- `14-koperasi-regulated-finance-guardrails.md` — guardrail untuk capability koperasi yang teregulasi
- `15-school-analytics-reporting-spec.md` — analytics, dashboard, dan reporting sekolah
- `16-release-readiness-gates.md` — gate kesiapan release dan evidence go-live
- `17-master-data-governance-spec.md` — tata kelola master data fondasional lintas domain
- `18-approval-workflow-orchestration-spec.md` — orkestrasi approval lintas domain
- `19-reporting-export-compliance-spec.md` — kontrol export, scheduled report, dan compliance output
- `20-tenant-activation-ops-playbook.md` — playbook operasional aktivasi tenant
- `21-feature-gating-entitlement-model.md` — model technical entitlement dari packaging bisnis
- `22-master-data-change-impact-matrix.md` — matrix dampak perubahan master data ke domain turunan
- `23-financial-reconciliation-ops-spec.md` — operasi rekonsiliasi keuangan sekolah dan koperasi
- `24-notification-messaging-operating-model.md` — model operasional komunikasi dan notifikasi
- `25-tenant-support-hypercare-playbook.md` — playbook support intensif pasca go-live tenant
- `26-payment-fee-absorption-policy.md` — kebijakan penanggung biaya pembayaran per channel
- `27-dapodik-operator-workflow-spec.md` — workflow operator Dapodik berbasis manual export
- `28-operational-sla-escalation-matrix.md` — matrix SLA insiden dan jalur eskalasi operasional
- `29-cross-domain-event-catalog.md` — katalog event lintas domain untuk producer dan consumer
- `30-pilot-tenant-rollout-governance.md` — governance rollout tenant pilot sebelum scale-out
- `31-tenant-configuration-baseline-spec.md` — baseline konfigurasi tenant yang repeatable
- `32-report-refresh-data-freshness-policy.md` — kebijakan freshness dan refresh report
- `33-financial-exception-resolution-playbook.md` — playbook penyelesaian exception finansial
- `34-tenant-ops-kpi-scorecard.md` — scorecard KPI operasional tenant
- `35-broadcast-communication-governance.md` — governance komunikasi broadcast institusional
- `36-configuration-change-control-policy.md` — kontrol perubahan konfigurasi tenant dan sistem
- `37-sprint-plan-from-specs.md` — rencana sprint eksekusi berdasarkan spec dan dependency ADR
- `38-sprint-0-3-detailed-task-packs.md` — task pack detail untuk sprint 0 sampai 3
- `39-sprint-4-7-detailed-task-packs.md` — task pack detail untuk sprint 4 sampai 7
- `40-sprint-8-11-detailed-task-packs.md` — task pack detail untuk sprint 8 sampai 11
- `41-sprint-backlog-summary-lane-view.md` — ringkasan backlog sprint dan pembagian lane per peran

## Status Lifecycle

- `Draft`
- `Review`
- `Approved`
- `Implemented`
- `Verified`

## Cara Pakai

1. Mulai dari ADR yang relevan
2. Cari capability di spec platform
3. Turunkan ke rollout slice dan acceptance criteria
4. Implementasi sesuai sequencing dan dependency
5. Update status serta traceability
