# 21 — Feature Gating & Entitlement Model

**Status**: Draft  
**Perspective**: Product + Revenue Ops + Engineering  
**Primary Inputs**: `ADR-BIZ-001`, `07-billing-metering-spec.md`, `01-portfolio-thesis.md`

## Purpose

Mendefinisikan bagaimana tier produk, add-on, quota, dan feature access diterjemahkan menjadi entitlement teknis yang konsisten di seluruh platform.

## Entitlement Principles

- packaging bisnis harus punya representasi teknis eksplisit
- feature gate tidak boleh bergantung pada hardcode yang tersebar
- quota, add-on, dan temporary override harus dapat diaudit
- denial experience harus jelas dan tidak merusak trust user

## Entitlement Building Blocks

### Plan / Tier
- Starter
- Pro
- Enterprise

### Feature Access
- binary enabled/disabled
- module-level enablement
- action-level restriction bila diperlukan

### Quota / Limits
- active students
- notification overage
- custom domain count
- premium analytics seats / capability caps

### Add-ons & Overrides
- white-label
- custom domain
- premium support
- temporary commercial exception / grace period

## Resolution Rules

- effective entitlement dihitung dari plan + add-ons + overrides - suspensions
- regulated capability hanya aktif jika compliance prerequisites juga terpenuhi
- expired add-on atau unpaid state harus punya policy transisi yang jelas: grace, read-only, atau hard stop

## Enforcement Surfaces

- backend authorization / capability checks
- frontend navigation and UX hints
- billing and usage metering
- admin back-office entitlement management

## Acceptance Criteria

- `AC-FUNC`: sistem dapat memutuskan apakah tenant berhak atas capability tertentu secara konsisten
- `AC-DATA`: plan, add-on, quota, override, dan suspension tercatat dan dapat direkonstruksi
- `AC-AUTH`: actor internal yang boleh mengubah entitlement dibatasi dan diaudit
- `AC-AUDIT`: perubahan entitlement dan alasan komersialnya dapat ditelusuri
- `AC-INT`: metering, billing, dan gate enforcement selaras tanpa kontradiksi besar
- `AC-NFR`: entitlement lookup cukup cepat untuk dipakai di request path utama

## Open Questions

- model grace period final untuk unpaid tenant
- batas hard stop vs read-only per capability
- apakah entitlement dievaluasi per tenant saja atau juga per company/subscope tertentu
