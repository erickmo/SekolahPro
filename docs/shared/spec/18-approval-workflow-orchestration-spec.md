# 18 — Approval Workflow Orchestration Spec

**Status**: Draft  
**Perspective**: System Analyst + Workflow Architect + Engineering  
**Primary Inputs**: `ADR-S047`, `ADR-S046`, `ADR-S050`, `ADR-S030`, `16-release-readiness-gates.md`

## Purpose

Mendefinisikan orkestrasi universal untuk flow persetujuan lintas domain agar setiap request yang butuh sign-off mengikuti model yang seragam, auditable, dan dapat diintegrasikan ke domain asal tanpa duplicate logic.

## Workflow Principles

- engine approval bersifat generic, domain-specific decision tetap milik domain asal
- request dan step harus dapat ditelusuri end-to-end
- approval, rejection, revision, delegation, cancelation, dan expiration adalah state formal
- SLA dan overdue handling adalah bagian inti, bukan tambahan

## Core Concepts

### Approval Chain
- konfigurasi per entity type
- dapat berbeda berdasarkan kondisi tertentu
- menentukan urutan, role, SLA, dan aturan step

### Approval Request
- satu submission per entity instance pada satu state approval aktif
- membawa entity reference, requestor, priority, current step, dan completion summary

### Approval Step
- merepresentasikan satu decision point
- bisa assigned, delegated, approved, rejected, skipped, atau revision requested
- dapat dipantau per approver inbox

## Orchestration Rules

- domain membuat approval request, bukan menjalankan logic approval internal terpisah
- engine mempublikasikan hasil akhir/antara sebagai event ke domain asal
- domain asal memutuskan perubahan final pada entitasnya sendiri berdasarkan event approval
- duplicate active approval untuk entity yang sama harus dicegah sesuai rule domain

## Cross-Domain Use Cases

- lesson plan approval
- leave request
- correspondence/signature routing
- budget / RKAS
- facility booking
- student transfer
- procurement dan future custom workflows

## Delegation & SLA

- delegation harus menyimpan siapa mendelegasikan ke siapa dan kapan
- overdue item harus muncul pada inbox dan escalation lane
- urgent request dapat memakai SLA/priority berbeda, bukan bypass tanpa jejak

## Acceptance Criteria

- `AC-FUNC`: domain dapat menggunakan engine approval tanpa membangun state machine sendiri
- `AC-AUTH`: hanya approver sah atau delegate sah yang dapat bertindak pada step aktif
- `AC-DATA`: chain, request, step, dan hasil akhir konsisten serta tidak ambigu
- `AC-AUDIT`: seluruh keputusan, revisi, delegation, dan overdue trail tercatat
- `AC-INT`: event ke domain asal dikirim andal dan idempotent
- `AC-NFR`: inbox pending, audit lookup, dan escalation monitoring tetap responsif

## Open Questions

- kapan `approval_steps` tetap Vernon vs CQRS murni pada implementasi final
- rule override untuk emergency approval lintas role
