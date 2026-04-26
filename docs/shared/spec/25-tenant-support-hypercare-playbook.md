# 25 — Tenant Support Hypercare Playbook

**Status**: Draft  
**Perspective**: COO + Support + Customer Success  
**Primary Inputs**: `20-tenant-activation-ops-playbook.md`, `11-observability-deployment-spec.md`, `16-release-readiness-gates.md`

## Purpose

Mendefinisikan playbook hypercare pasca go-live untuk tenant baru agar issue awal cepat tertangani, risiko churn turun, dan transisi dari onboarding ke BAU support berjalan mulus.

## Hypercare Principles

- fokus pada stabilitas minggu awal, bukan hanya ticket closure
- issue yang menghambat operasi inti diprioritaskan absolut
- owner, SLA, dan jalur eskalasi harus jelas sejak hari pertama live
- feedback hypercare harus dipakai untuk memperbaiki activation playbook

## Hypercare Window

- default 7-30 hari tergantung tier dan kompleksitas rollout
- Enterprise / hybrid school+koperasi dapat memakai window lebih panjang
- exit hypercare hanya setelah KPI minimum dan critical issue threshold tercapai

## Core Tracks

### Operational Monitoring
- first login, first import success, first transaction, first report/export
- issue trend per capability
- adoption gap per actor utama

### Support Handling
- war room / channel khusus tenant bila perlu
- incident triage: blocker, major, minor, how-to
- callback cadence ke owner tenant

### Stabilization Actions
- data correction / reconciliation
- permission misconfiguration fixes
- workflow tuning, template tuning, notification tuning
- retraining singkat bila root cause adalah process gap

## Exit Criteria

- tidak ada blocker terbuka
- proses inti berjalan stabil
- owner tenant menyetujui transisi ke support normal
- issue backlog tersisa sudah diprioritaskan dengan jelas

## Acceptance Criteria

- `AC-FUNC`: tenant baru mendapat jalur dukungan intensif yang terstruktur setelah go-live
- `AC-DATA`: issue, root cause, dan stabilization action tercatat dan dapat dianalisis
- `AC-AUDIT`: keputusan masuk/keluar hypercare dan exception penting terdokumentasi
- `AC-INT`: hypercare findings mengalir ke product/ops/engineering backlog yang tepat
- `AC-NFR`: model hypercare bisa diskalakan ke banyak tenant tanpa kehilangan kontrol

## Open Questions

- SLA hypercare final per tier
- kapan tenant otomatis kembali ke support normal walau masih ada issue minor
