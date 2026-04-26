# 33 — Financial Exception Resolution Playbook

**Status**: Draft  
**Perspective**: Finance Ops + Risk + Engineering  
**Primary Inputs**: `23-financial-reconciliation-ops-spec.md`, `26-payment-fee-absorption-policy.md`, `14-koperasi-regulated-finance-guardrails.md`

## Purpose

Mendefinisikan playbook penyelesaian exception keuangan agar mismatch, duplicate, failed callback, teller imbalance, dan anomaly finansial lain ditangani dengan workflow yang seragam dan auditable.

## Exception Principles

- tidak ada exception finansial yang hilang di chat informal
- resolution harus preserve trail, bukan overwrite sejarah
- tindakan koreksi mengikuti approval dan materiality yang jelas
- close case hanya bila root cause dan financial effect dipahami

## Exception Families

### Payment Exceptions
- callback mismatch
- duplicate payment
- paid not posted
- settlement discrepancy
- refund/reversal anomaly

### Cooperative Exceptions
- teller imbalance
- balance mismatch
- reversal chain issue
- transaction validation bypass suspicion

## Resolution Flow

- detect / register exception
- assign owner and severity
- freeze or continue decision jika perlu
- investigate source evidence
- perform correction / reversal / waiver / dispute handling
- verify post-resolution state
- close with root cause note

## Acceptance Criteria

- `AC-FUNC`: tim dapat menangani exception finansial dengan langkah yang repeatable
- `AC-DATA`: evidence, correction, dan resulting balances/status dapat ditelusuri end-to-end
- `AC-AUTH`: tindakan korektif penting dibatasi oleh role dan approval tier
- `AC-AUDIT`: exception history dan root-cause closure tercatat lengkap
- `AC-INT`: hasil resolution mengalir ke reconciliation status dan audit/report pack yang relevan
- `AC-NFR`: playbook cukup cepat dipakai saat insiden nyata tanpa mengorbankan kontrol

## Open Questions

- materiality threshold final per exception family
- kapan auto-resolution boleh dilakukan tanpa review manual
