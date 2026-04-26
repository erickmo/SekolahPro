# 30 — Pilot Tenant Rollout Governance

**Status**: Draft  
**Perspective**: CEO/COO/CTO + Product + Ops  
**Primary Inputs**: `16-release-readiness-gates.md`, `20-tenant-activation-ops-playbook.md`, `25-tenant-support-hypercare-playbook.md`

## Purpose

Mendefinisikan governance untuk rollout tenant pilot agar validasi produk, risiko operasional, dan pembelajaran lapangan berjalan terkendali sebelum scale-out lebih luas.

## Governance Principles

- pilot bukan hanya sales milestone, tapi controlled learning phase
- jumlah tenant pilot dibatasi sesuai kapasitas hypercare
- setiap pilot harus punya hypothesis, success metric, dan kill criteria
- temuan pilot harus kembali ke roadmap dan spec, bukan hilang di chat operasional

## Pilot Controls

### Entry Criteria
- capability relevan lolos release gate minimum
- owner internal dan owner tenant jelas
- support/hypercare capacity tersedia

### During Pilot
- issue log dan change log aktif
- adoption, failure mode, dan feedback dipantau mingguan
- scope creep dibatasi ketat

### Exit / Expansion Decision
- continue as-is
- continue with remediation backlog
- pause rollout
- rollback/disable capability tertentu

## Decision Inputs

- activation success
- hypercare burden
- incident severity profile
- feature adoption vs intended value
- compliance or finance readiness signals

## Acceptance Criteria

- `AC-FUNC`: tim dapat menjalankan pilot tenant dengan kontrol, review cadence, dan decision gate yang jelas
- `AC-DATA`: hypothesis, KPI, incidents, dan decision outcome terdokumentasi
- `AC-AUDIT`: siapa yang menyetujui masuk/keluar pilot dan alasan keputusannya tercatat
- `AC-INT`: temuan pilot mengalir ke backlog product/ops/engineering secara sistematis
- `AC-NFR`: governance cukup ketat untuk mengurangi risiko, tapi tidak membunuh kecepatan belajar

## Open Questions

- maksimum tenant pilot paralel per wave
- syarat formal untuk pindah dari pilot ke GA per capability
