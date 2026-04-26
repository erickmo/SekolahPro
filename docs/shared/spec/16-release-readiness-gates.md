# 16 — Release Readiness Gates

**Status**: Draft  
**Perspective**: CTO + Product + Engineering + Ops  
**Primary Inputs**: `ADR-016`, `11-observability-deployment-spec.md`, `12-api-event-contract-spec.md`, `08-compliance-risk-controls.md`

## Purpose

Mendefinisikan gate sebelum sebuah capability atau release dapat dipromosikan dari draft implementation ke staging dan production.

## Gate Levels

### Gate 0 — Spec Ready
- scope jelas
- ADR references jelas
- acceptance criteria ada
- dependency dan out-of-scope terdokumentasi

### Gate 1 — Build Ready
- implementation selesai sesuai spec minimum
- lint/test/type/build relevan lolos
- migration/contract changes terdokumentasi

### Gate 2 — Integration Ready
- API/event contract diverifikasi
- observability minimum tersedia
- retry/failure mode utama diuji
- import/integration/payment callback behavior tervalidasi bila applicable

### Gate 3 — Compliance & Security Ready
- privacy/security checklist lolos
- regulated capability memiliki sign-off yang diperlukan
- audit trail dan access controls diverifikasi

### Gate 4 — Production Ready
- runbook tersedia
- rollback plan tersedia
- metrics/alerts hidup
- release owner dan on-call jelas
- stakeholder sign-off selesai

## Evidence Required

- spec link
- ADR link
- test/build evidence
- contract review note
- migration/release plan
- risk note / waiver jika ada
- operational dashboard or alert evidence untuk capability kritikal

## Waiver Policy

- waiver harus time-bound
- owner, risk, mitigation, dan expiry wajib ditulis
- regulated/security gate tidak boleh di-waive tanpa persetujuan eksplisit yang tepat

## Acceptance Criteria

- `AC-FUNC`: setiap release memiliki jalur evaluasi yang jelas
- `AC-AUDIT`: keputusan go/no-go dapat ditelusuri
- `AC-INT`: perubahan contract dan integration tidak lolos tanpa review
- `AC-NFR`: release safety dan rollback readiness menjadi syarat, bukan optional
