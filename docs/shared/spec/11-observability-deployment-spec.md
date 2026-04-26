# 11 — Observability, Deployment & Operability

**Status**: Draft  
**Perspective**: System Engineer / SRE  
**Primary Inputs**: `ADR-016`, `ADR-005`, `ADR-014`, `ADR-017`

## Purpose

Mendefinisikan baseline operability untuk deployment, monitoring, alerting, release safety, dan failure handling.

## Environment Model

- **Local**: development dan integration cepat
- **Staging**: QA, contract check, rehearsal release
- **Production**: live tenant dengan approval release

## Runtime Topology

- Go API
- web apps / portal / admin / POS
- PostgreSQL
- Redis
- NATS JetStream
- observability stack

## Observability Requirements

### Logs
- structured logs
- tenant-aware tanpa membocorkan PII berlebihan
- korelasi request, command, event, integration callback

### Metrics
- request rate, error rate, latency
- queue depth, retry count, DLQ count
- import progress/failure rates
- notification delivery rates
- sync lag dan reconciliation backlog

### Traces
- trace end-to-end untuk HTTP -> domain -> event -> external provider
- correlation ID wajib lintas boundary

## Deployment Rules

- image yang sama dipromosikan antar environment
- migration selalu eksplisit
- healthz dan readyz wajib
- rollback release harus cepat dan terdokumentasi

## Operability Controls

- alert severity dan escalation path
- runbook untuk failure umum
- backup, restore, replay, resync drills
- release checklist untuk schema, event, dan contract changes

## Acceptance Criteria

- `AC-FUNC`: engineer dapat mendeteksi, mendiagnosis, dan merespons issue tanpa blind spot besar
- `AC-DATA`: backup/restore/replay menjaga integritas data dan audit trail
- `AC-AUDIT`: perubahan release, infra, dan incident action tercatat
- `AC-INT`: external provider failure dapat dideteksi dan diisolasi
- `AC-NFR`: SLI/SLO didefinisikan untuk capability kritikal

## Open Gaps

- target SLO per capability belum final
- runbook katalog belum diturunkan per modul
- kebijakan DR dan RTO/RPO detail masih perlu dipecah lanjut
