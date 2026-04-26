# 41 — Sprint Backlog Summary & Lane View

**Status**: Draft

## Lane Definitions

### C-Level
- portfolio priority
- pricing / rollout / risk appetite
- go/no-go and pilot decisions

### System Analyst
- actor/use case modeling
- acceptance criteria and process design
- dependency and rollout slice clarity

### System Engineer
- service/integration/operability design
- NFR, controls, observability, deployment readiness
- implementation feasibility and risk reduction

## Sprint-to-Lane Summary

| Sprint | C-Level Focus | Analyst Focus | Engineer Focus |
|---|---|---|---|
| 0 | foundation approval | shared acceptance baseline | contracts, observability, release gates |
| 1 | pilot/tier direction | onboarding + tenant config | auth, RBAC, baseline config |
| 2 | school-first, coop-core as expansion | master data + coop core use cases | governance + scope controls |
| 3 | student proof of value | intake-to-placement | import/data quality/auditability |
| 4 | academic core priority | curriculum + financing flows | orchestration + dependency controls |
| 5 | daily ops and transaction readiness | attendance/grades + exception handling | reconciliation + throughput stability |
| 6 | rapor/SPP retention + accounting readiness | orchestration + HR/accounting outputs | guardrails + export/report control |
| 7 | reporting as scale enabler | role-based reports | refresh/freshness/export operability |
| 8 | segment-specific premium rollout | boarding/canteen/extension flows | extension boundaries + guardrails |
| 9 | retention & engagement | portal/comms journey | multi-channel delivery + governance |
| 10 | control plane maturity | approval/payment/admin flows | approval engine + payment controls |
| 11 | scale-out decision | Dapodik + KPI review | governance hardening + integration hygiene |

## Recommended Backlog Management

- backlog dipecah per sprint lalu per lane
- setiap sprint memiliki `must-have`, `should-have`, `pilot-only`
- semua story harus merujuk minimal satu spec
- sprint review harus memperbarui pilot/governance docs bila ada pembelajaran baru
