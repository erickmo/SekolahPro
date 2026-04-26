# 34 — Tenant Ops KPI Scorecard

**Status**: Draft  
**Perspective**: COO + Customer Success + Product Ops  
**Primary Inputs**: `20-tenant-activation-ops-playbook.md`, `25-tenant-support-hypercare-playbook.md`, `28-operational-sla-escalation-matrix.md`

## Purpose

Mendefinisikan scorecard KPI operasional tenant untuk memantau aktivasi, adopsi, stabilitas, dan support load secara konsisten lintas tenant.

## KPI Families

### Activation
- time-to-first-value
- first import success
- first transaction/report completion

### Adoption
- active operator count
- capability usage depth
- critical workflow completion rate

### Stability
- incident count by severity
- open blocker duration
- hypercare burden

### Support & Satisfaction
- response SLA attainment
- repeat issue rate
- tenant owner satisfaction signal

## Acceptance Criteria

- `AC-FUNC`: tim ops dapat membandingkan health tenant secara konsisten
- `AC-DATA`: scorecard metric memiliki definisi operasional yang jelas
- `AC-AUDIT`: perubahan definisi KPI dan target dicatat
- `AC-INT`: scorecard terhubung ke activation, hypercare, dan escalation workflow
- `AC-NFR`: metric set cukup ringkas untuk operasional, tidak menjadi dashboard bloat

## Open Questions

- KPI minimum wajib per tier
- kapan scorecard dipakai sebagai trigger executive review
