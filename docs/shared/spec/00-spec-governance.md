# 00 — Spec Governance

**Status**: Draft  
**Owner**: Product + Architecture  
**Sources**: `docs/adr/README.md`, `docs/adr/SPRINT-ORDER.md`

## Scope

Dokumen ini mendefinisikan aturan untuk seluruh file di `docs/spec`.

## Roles

### C-Level
- Menetapkan prioritas portofolio
- Menyetujui packaging, sequencing, dan risk appetite
- Menentukan go/no-go per wave

### System Analyst
- Mendefinisikan aktor, proses bisnis, use case, acceptance criteria
- Menjaga traceability dari ADR ke delivery slice
- Memastikan domain boundary sesuai kebutuhan bisnis

### System Engineer
- Mendefinisikan service boundary, kontrak API/event, NFR, observability, deployment
- Memastikan implementability sesuai arsitektur core
- Menjaga keamanan, tenancy, dan operability

## Definition of Done for a Spec

Sebuah spec dianggap siap diimplementasikan bila memiliki:

- tujuan bisnis
- ruang lingkup dan out-of-scope
- ADR references
- actor list
- business process / capability breakdown
- dependency map
- acceptance criteria
- rollout slices
- risk dan open questions

## Document Rules

- Semua spec wajib mencantumkan ADR referensi
- Semua capability wajib diletakkan pada salah satu scope: `core`, `sekolah`, `koperasi`, `shared`
- Semua cross-cutting concern harus dirujuk ke `03-cross-cutting-foundation.md`
- Semua rollout wajib mengikuti urutan dependency pada `SPRINT-ORDER.md`
- Semua perubahan scope besar harus mengupdate `06-traceability-matrix.md`

## Acceptance Criteria Standard

Gunakan kategori berikut:

- `AC-FUNC` — hasil bisnis/fungsional
- `AC-DATA` — state, persistence, invariants
- `AC-AUTH` — actor, role, scope, permission
- `AC-VALID` — validation dan edge cases
- `AC-AUDIT` — logging, approval trail, immutability
- `AC-INT` — event, integration, external outcome
- `AC-NFR` — latency, availability, consistency, operability

## Rollout Model

Tiap capability dibagi menjadi salah satu tipe berikut:

- Foundation
- Transaction
- Orchestration
- Reporting
- Integration

## Review Gate

- **Draft -> Review**: analyst + engineer sepakat boundary dan acceptance cukup jelas
- **Review -> Approved**: C-level/owner menyetujui prioritas dan scope
- **Approved -> Implemented**: implementation merged
- **Implemented -> Verified**: lint, tests, UAT, dan operational checks lolos
