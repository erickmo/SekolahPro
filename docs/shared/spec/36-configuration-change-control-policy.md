# 36 — Configuration Change Control Policy

**Status**: Draft  
**Perspective**: Ops + Engineering + Security  
**Primary Inputs**: `31-tenant-configuration-baseline-spec.md`, `22-master-data-change-impact-matrix.md`, `16-release-readiness-gates.md`

## Purpose

Mendefinisikan kontrol perubahan konfigurasi tenant dan sistem agar perubahan penting tidak dilakukan secara ad-hoc tanpa impact review, approval, dan audit trail.

## Policy Principles

- perubahan konfigurasi bisa sama berbahayanya dengan code change
- high-impact config change perlu impact review dan rollback plan
- low-impact change tetap harus traceable
- emergency change boleh cepat, tapi tidak boleh undocumented

## Change Classes

### Standard
- perubahan rutin berisiko rendah

### Significant
- perubahan yang memengaruhi capability, entitlement, integration, atau user experience luas

### Critical
- perubahan yang memengaruhi security, compliance, financial operations, atau banyak tenant

## Control Requirements

- siapa mengubah apa, kapan, dari nilai apa ke nilai apa harus tercatat
- class change menentukan apakah approval, maintenance coordination, atau post-change verification dibutuhkan
- emergency change harus punya post-facto review
- rollback/restore path harus jelas untuk critical config

## Acceptance Criteria

- `AC-FUNC`: tim dapat melakukan perubahan konfigurasi dengan workflow kontrol yang sesuai risikonya
- `AC-DATA`: before/after state, owner, dan impact classification dapat ditelusuri
- `AC-AUDIT`: approval, emergency justification, dan post-change verification tercatat
- `AC-INT`: policy ini terhubung ke baseline config, activation, dan operational incident handling
- `AC-NFR`: kontrol cukup ringan untuk perubahan kecil namun ketat untuk perubahan kritikal

## Open Questions

- daftar config yang otomatis masuk class critical
- apakah approval chain berbeda per tenant tier atau per domain risiko
