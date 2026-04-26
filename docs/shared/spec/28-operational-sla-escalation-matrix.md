# 28 — Operational SLA & Escalation Matrix

**Status**: Draft  
**Perspective**: Ops + Support + Engineering + Leadership  
**Primary Inputs**: `11-observability-deployment-spec.md`, `16-release-readiness-gates.md`, `25-tenant-support-hypercare-playbook.md`

## Purpose

Mendefinisikan klasifikasi insiden/issue operasional, target respons dan resolusi, serta jalur eskalasi agar operasi tenant, support, dan engineering memiliki bahasa prioritas yang sama.

## Severity Model

### Sev 1 — Critical Business Stop
- tenant tidak bisa menjalankan proses inti
- payment/finance/critical auth failure luas
- major data integrity risk aktif

### Sev 2 — Major Degradation
- capability utama terganggu tapi ada workaround terbatas
- tenant penting / banyak user terdampak

### Sev 3 — Moderate Issue
- gangguan terbatas, non-critical path, atau workaround jelas

### Sev 4 — Minor / How-to / Cosmetic
- tidak menghambat operasi inti

## SLA Dimensions

- acknowledge time
- owner assignment time
- escalation trigger time
- target restore / target workaround / target resolution
- communication cadence ke stakeholder

## Escalation Paths

- Support L1 -> Ops / Product Ops
- Ops -> Engineering on-call
- Engineering -> domain owner / infra / security / leadership sesuai jenis insiden
- compliance/regulatory issue harus punya jalur terpisah dan cepat

## Special Cases

- hypercare tenant memakai threshold eskalasi lebih agresif
- regulated finance / privacy / security incident tidak mengikuti jalur minor biasa
- repeated issue pada capability sama harus memicu problem-management style review

## Acceptance Criteria

- `AC-FUNC`: setiap issue operasional dapat dipetakan ke severity dan jalur eskalasi yang jelas
- `AC-DATA`: SLA timestamps, owner, dan escalation history tercatat
- `AC-AUDIT`: keputusan severity override dan escalation terdokumentasi
- `AC-INT`: observability alerts dan support tickets dapat dipetakan ke matrix ini
- `AC-NFR`: matrix cukup sederhana untuk dipakai cepat saat incident nyata

## Open Questions

- target numerik final per severity dan per tier
- kapan executive escalation menjadi mandatory
