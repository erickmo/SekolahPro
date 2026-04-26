# 22 — Master Data Change Impact Matrix

**Status**: Draft  
**Perspective**: Data Governance + System Engineer  
**Primary Inputs**: `ADR-014`, `ADR-010`, `ADR-011`, `ADR-012`, `17-master-data-governance-spec.md`

## Purpose

Mendefinisikan cara menilai dan mengelola dampak perubahan master data terhadap domain turunan, sync propagation, operasional, dan risiko stale-read.

## Impact Principles

- tidak semua perubahan master data setara
- perubahan label kecil, status, ownership, atau lifecycle bisa punya blast radius berbeda
- impact harus dinilai sebelum change kritikal dilakukan
- stale > down, tetapi stale kritikal harus punya SLA dan observability

## Change Categories

### Low Impact
- typo fix non-kritikal
- metadata tambahan yang tidak di-cache luas

### Medium Impact
- rename master data yang dipakai di display banyak modul
- reassignment owner seperti homeroom teacher atau class mapping tertentu

### High Impact
- activation / closure academic year
- status change yang memengaruhi akses atau validitas transaksi
- structural merge/split/carry-forward yang mengubah banyak relasi

## Impact Dimensions

- affected domains/tables
- sync priority dan SLA
- user-facing staleness risk
- reporting/export impact
- rollback/recovery complexity
- approval requirement

## Matrix Usage

Setiap master data change penting harus menjawab:
- master domain mana yang berubah
- field apa yang berubah
- domain turunan mana yang cache atau bergantung padanya
- apakah manual resync / reconciliation mungkin diperlukan
- apakah change boleh dilakukan jam kerja atau butuh maintenance window ringan

## Operational Controls

- dependency registry harus jadi sumber evaluasi blast radius
- pending/error sync setelah change harus dimonitor aktif
- high-impact change perlu post-change verification checklist
- change yang memengaruhi regulated/reporting output harus menandai downstream report risk

## Acceptance Criteria

- `AC-FUNC`: tim dapat mengklasifikasikan change dan tahu tindakan operasional yang dibutuhkan
- `AC-DATA`: affected domain dan sync dependency dapat diidentifikasi dengan cukup jelas
- `AC-AUDIT`: keputusan change-impact, approval, dan verification tercatat
- `AC-INT`: manual resync, reconciliation, atau report refresh dapat dipicu bila diperlukan
- `AC-NFR`: proses impact review cukup ringan untuk dipakai rutin namun cukup ketat untuk change kritikal

## Open Questions

- kapan matrix ini diubah menjadi template formal per change request
- domain mana yang perlu kategori `very_high` terpisah
