# 06 — Traceability Matrix

**Status**: Draft

## Mapping Summary

| Spec | Scope | Primary ADR Sources | Delivery Intent |
|---|---|---|---|
| `00-spec-governance.md` | shared | ADR README, Sprint Order | aturan spec dan gate |
| `01-portfolio-thesis.md` | shared | ADR-BIZ-001, ADR-003, ADR-004, ADR-009, ADR-018 | prioritas bisnis dan packaging |
| `02-capability-map.md` | shared | core + sekolah + koperasi README | capability decomposition |
| `03-cross-cutting-foundation.md` | core/shared | ADR-001..018 | kontrak arsitektur dan operability |
| `04-school-platform-spec.md` | sekolah | ADR-S001..S058 | rollout school platform |
| `05-cooperative-platform-spec.md` | koperasi | ADR-K001..K038 | rollout cooperative platform |
| `07-billing-metering-spec.md` | shared | ADR-BIZ-001, ADR-004, ADR-S044, ADR-S051, ADR-K011, ADR-018 | monetization operations |
| `08-compliance-risk-controls.md` | shared | ADR-018, ADR-004, ADR-013, ADR-S031, ADR-S051, koperasi compliance ADRs | compliance operating model |
| `09-onboarding-migration-spec.md` | shared | ADR-017, ADR-004, ADR-009, ADR-010, ADR-011, ADR-012, ADR-013 | activation and migration readiness |
| `10-parent-student-member-experience.md` | shared | ADR-S042, ADR-S043, ADR-S044, ADR-S045, ADR-S009, ADR-S018 | external user experience |
| `11-observability-deployment-spec.md` | shared | ADR-016, ADR-005, ADR-014, ADR-017 | operability and release safety |
| `12-api-event-contract-spec.md` | shared | ADR-004, ADR-005, ADR-014, ADR-015 | contract and compatibility rules |
| `13-security-privacy-data-lifecycle-spec.md` | shared | ADR-018, ADR-004, ADR-013 | security, privacy, and data lifecycle |
| `14-koperasi-regulated-finance-guardrails.md` | koperasi/shared | ADR-K027, ADR-K028, ADR-K031, ADR-K038, ADR-018 | regulated finance control boundaries |
| `15-school-analytics-reporting-spec.md` | sekolah/shared | ADR-S054, ADR-S055 | analytics, dashboard, and reporting model |
| `16-release-readiness-gates.md` | shared | ADR-016, spec 08, spec 11, spec 12 | go-live control gates |
| `17-master-data-governance-spec.md` | shared/core | ADR-010, ADR-011, ADR-012, ADR-013 | governed master data operating model |
| `18-approval-workflow-orchestration-spec.md` | shared/sekolah | ADR-S047, ADR-S046, ADR-S050, ADR-S030 | universal approval orchestration |
| `19-reporting-export-compliance-spec.md` | shared | ADR-S054, ADR-S055, ADR-K017, ADR-K027, ADR-018 | controlled reporting and export evidence |
| `20-tenant-activation-ops-playbook.md` | shared/ops | ADR-017, ADR-BIZ-001 | repeatable tenant activation operations |
| `21-feature-gating-entitlement-model.md` | shared | ADR-BIZ-001, spec 07 | technical packaging and entitlement control |
| `22-master-data-change-impact-matrix.md` | shared/core | ADR-014, ADR-010, ADR-011, ADR-012 | blast-radius and sync impact review |
| `23-financial-reconciliation-ops-spec.md` | shared/sekolah/koperasi | ADR-S009, ADR-S051, ADR-K011 | reconciliation and exception handling model |
| `24-notification-messaging-operating-model.md` | shared/sekolah | ADR-S043, ADR-S044 | communications and notification operations |
| `25-tenant-support-hypercare-playbook.md` | shared/ops | spec 20, spec 11, spec 16 | post-go-live stabilization playbook |
| `26-payment-fee-absorption-policy.md` | shared/finance | ADR-S051, spec 07, spec 23 | fee bearer and surcharge policy |
| `27-dapodik-operator-workflow-spec.md` | sekolah/ops | ADR-S055, ADR-S054 | operator-facing Dapodik execution workflow |
| `28-operational-sla-escalation-matrix.md` | shared/ops | spec 11, spec 16, spec 25 | incident SLA and escalation control |
| `29-cross-domain-event-catalog.md` | shared | ADR-005, spec 12, ADR-S044, ADR-014 | event ecosystem reference map |
| `30-pilot-tenant-rollout-governance.md` | shared/ops | spec 16, spec 20, spec 25 | controlled pilot rollout governance |
| `31-tenant-configuration-baseline-spec.md` | shared/ops/core | ADR-004, spec 20, spec 21 | repeatable tenant baseline configuration |
| `32-report-refresh-data-freshness-policy.md` | shared/reporting | spec 15, spec 19, spec 11 | report freshness and refresh governance |
| `33-financial-exception-resolution-playbook.md` | shared/finance | spec 23, spec 26, spec 14 | financial anomaly resolution workflow |
| `34-tenant-ops-kpi-scorecard.md` | shared/ops | spec 20, spec 25, spec 28 | tenant operations health metrics |
| `35-broadcast-communication-governance.md` | shared/sekolah | spec 24, spec 10, spec 19 | high-reach communication governance |
| `36-configuration-change-control-policy.md` | shared/ops/core | spec 31, spec 22, spec 16 | controlled configuration change management |
| `37-sprint-plan-from-specs.md` | shared/portfolio | sprint-order + spec 04 + spec 05 + spec 06 | executable sprint roadmap |
| `38-sprint-0-3-detailed-task-packs.md` | shared/portfolio | spec 37 + activation/ops specs | early sprint execution packs |
| `39-sprint-4-7-detailed-task-packs.md` | shared/portfolio | spec 37 + finance/reporting specs | mid sprint execution packs |
| `40-sprint-8-11-detailed-task-packs.md` | shared/portfolio | spec 37 + experience/integration specs | late sprint execution packs |
| `41-sprint-backlog-summary-lane-view.md` | shared/portfolio | spec 37, spec 38, spec 39, spec 40 | sprint lane and backlog operating view |

## Rollout Alignment

| Wave | Focus | Dependency Basis |
|---|---|---|
| Wave 1 | Platform + School Core | Fase 0-6A yang critical path untuk monetization awal |
| Wave 2 | Cooperative Core + Transaction Engine | Fase 3, 4B, 5B, 6B |
| Wave 3 | Premium Pesantren + Compliance/Analytics/Integrations | Fase 7+ dan ADR proposed lintas domain |

## Current Coverage Notes

- billing, metering, compliance, onboarding, experience, dan operability kini sudah memiliki spec awal
- kebutuhan berikutnya adalah pemecahan lebih detail per domain atau per contract area
- billing/payment dan compliance kemungkinan perlu dipecah lagi saat implementation dimulai

## Next Spec Candidates

Tidak ada gap prioritas langsung yang teridentifikasi dari review ADR saat ini. Spec lanjutan berikutnya sebaiknya hanya ditambahkan bila:

- implementasi menemukan ambiguity nyata
- ada ADR baru
- ada kebutuhan audit/compliance baru
- hasil pilot tenant memunculkan gap operasional baru
