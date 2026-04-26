# 20 — Tenant Activation Ops Playbook

**Status**: Draft  
**Perspective**: COO + Customer Success + System Analyst  
**Primary Inputs**: `ADR-017`, `01-portfolio-thesis.md`, `09-onboarding-migration-spec.md`

## Purpose

Mendefinisikan playbook operasional untuk mengaktifkan tenant dari signed deal / approved onboarding sampai tenant benar-benar live dan menjalankan proses inti pertamanya.

## Activation Goals

- time-to-value cepat
- onboarding dapat diprediksi dan diulang
- risiko aktivasi gagal turun
- handoff antar sales, onboarding, ops, dan support jelas

## Activation Stages

### Stage 1 — Commercial Readiness
- tenant/package/tier dikonfirmasi
- scope implementasi dan success criteria disepakati
- owner tenant dan owner internal ditetapkan

### Stage 2 — Foundation Setup
- tenant/company dibuat
- admin awal dan role dasar aktif
- institution mode, academic year, dan baseline config disiapkan

### Stage 3 — Data Readiness
- template import dibagikan
- data mapping workshop bila perlu
- dry-run dan correction loop diselesaikan
- critical master data tervalidasi

### Stage 4 — Operational Readiness
- proses inti diuji: student ops, finance ops, atau cooperative ops sesuai paket
- reporting/dashboard dasar diverifikasi
- training operator selesai

### Stage 5 — Go-Live & Early Support
- first live transaction / first live process dicatat
- hypercare window aktif
- activation KPI dan issue log dipantau

## Mandatory Checklists

- access & admin readiness
- master data readiness
- migration/import readiness
- billing/subscription readiness
- support escalation path
- go-live rollback/fallback plan bila dibutuhkan

## Acceptance Criteria

- `AC-FUNC`: tenant dapat berpindah dari setup ke live dengan checklist yang jelas
- `AC-DATA`: data inti tervalidasi sebelum dipakai operasional
- `AC-AUTH`: admin/operator memiliki akses yang tepat sebelum go-live
- `AC-AUDIT`: keputusan activation, import, dan sign-off tercatat
- `AC-INT`: issue onboarding dan dependency eksternal memiliki owner/follow-up
- `AC-NFR`: activation process dapat diskalakan untuk banyak tenant tanpa chaos operasional

## Open Questions

- definisi final activation success metric per tier
- kapan onboarding premium / assisted migration menjadi paket terpisah
