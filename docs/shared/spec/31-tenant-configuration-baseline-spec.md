# 31 — Tenant Configuration Baseline Spec

**Status**: Draft  
**Perspective**: System Analyst + Ops + Engineering  
**Primary Inputs**: `ADR-004`, `20-tenant-activation-ops-playbook.md`, `21-feature-gating-entitlement-model.md`

## Purpose

Mendefinisikan baseline konfigurasi tenant yang harus tersedia agar tenant baru dapat diaktifkan secara konsisten tanpa konfigurasi liar dan tanpa ketergantungan tribal knowledge.

## Baseline Configuration Areas

### Tenant Identity & Scope
- tenant/company identity
- deployment mode / scope model
- institution type / school mode
- operational timezone / locale baseline

### Access & Roles
- admin awal
- role baseline
- permission scope default
- escalation/admin contact

### Product & Entitlement
- active plan/tier
- enabled modules
- quotas, add-ons, grace settings
- regulated capability switches

### Operational Defaults
- academic year baseline
- notification defaults
- payment channel defaults jika applicable
- import/report/export baseline settings

## Configuration Principles

- baseline harus minimal namun cukup untuk activation
- setiap override dari baseline harus tercatat
- tenant-specific customization tidak boleh mem-bypass core compliance/security guardrail
- config ownership antara ops, finance, dan engineering harus jelas

## Acceptance Criteria

- `AC-FUNC`: tenant baru dapat dibuat dari baseline yang repeatable
- `AC-DATA`: baseline dan override terdokumentasi dan dapat direkonstruksi
- `AC-AUTH`: hanya role yang tepat dapat mengubah baseline dan override kritikal
- `AC-AUDIT`: perubahan konfigurasi penting tercatat lengkap
- `AC-INT`: baseline sinkron dengan activation, entitlement, dan support workflow
- `AC-NFR`: baseline config cukup stabil untuk mengurangi setup drift antar tenant

## Open Questions

- format sumber kebenaran baseline final: config-as-data, admin UI, atau template hybrid
- baseline mana yang wajib locked per tier vs bisa diubah tenant sendiri
