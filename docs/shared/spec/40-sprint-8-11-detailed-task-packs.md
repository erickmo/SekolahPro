# 40 — Sprint 8-11 Detailed Task Packs

**Status**: Draft

## Sprint 8 — Boarding, Canteen, Cooperative Extensions

### C-Level Lane
- hanya aktifkan untuk segment pesantren / tenant yang relevan
- jaga extension koperasi tetap tunduk pada guardrail legal/ops

### System Analyst Lane
- definisikan boarding and canteen workflows
- definisikan payroll deduction / wallet extension assumptions

### System Engineer Lane
- siapkan extension boundaries agar tidak mencemari core flows
- siapkan guardrail checks before rollout

### Story Backlog
- story: manage dormitory occupancy and activities
- story: manage dormitory discipline and health linkage
- story: operate canteen menu, vendors, and billing linkage
- story: pilot cooperative store and payroll deduction extensions
- story: review wallet-like capabilities against guardrail constraints

## Sprint 9 — Communication, Portal, Stakeholder Experience

### C-Level Lane
- tetapkan stakeholder experience sebagai retention and differentiation layer
- tetapkan governance untuk broadcast and external comms

### System Analyst Lane
- definisikan parent/student/member journey utama
- definisikan messaging vs notification boundary and broadcast rules

### System Engineer Lane
- siapkan notification policy, channel fallback, and message governance controls
- siapkan portal-readiness metrics and complaint handling hooks

### Story Backlog
- story: deliver multi-channel notifications with preference handling
- story: support scoped messaging for school stakeholders
- story: publish announcements and governed broadcasts
- story: expose parent portal with key academic/finance views
- story: expose cooperative self-service portal and alerts

## Sprint 10 — Administration, Approval, Payments, Budgeting

### C-Level Lane
- prioritaskan payment modernization hanya jika school finance baseline stabil
- tetapkan approval engine sebagai reusable control-plane asset

### System Analyst Lane
- definisikan admin and approval use cases lintas domain
- definisikan payment reconciliation and exception closure requirements

### System Engineer Lane
- siapkan generic approval orchestration integration pattern
- siapkan payment controls, fee handling, and reconciliation hooks

### Story Backlog
- story: manage school correspondence and administrative routing
- story: orchestrate reusable approval requests across domains
- story: manage school profile, committee, and budgeting baseline
- story: enable payment gateway flows with fee policy awareness
- story: resolve finance exceptions and maintain audit-grade payment trail

## Sprint 11 — Dapodik, Integrations, Governance Hardening

### C-Level Lane
- gunakan evidence pilot untuk keputusan scale-out
- tetapkan go/no-go integrasi yang bernilai tinggi saja

### System Analyst Lane
- definisikan Dapodik operator operating rhythm dan governance review pack
- definisikan KPI review ritual pasca pilot

### System Engineer Lane
- siapkan Dapodik workflow hardening, reporting freshness alignment, and event catalog hygiene
- siapkan governance follow-up items from pilot and operations data

### Story Backlog
- story: validate and export Dapodik-ready data per semester workflow
- story: harden report/export and freshness governance for wider rollout
- story: review pilot rollout evidence and tenant ops KPI scorecards
- story: tighten cross-domain event catalog and governance alignment
