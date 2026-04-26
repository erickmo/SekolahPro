# 23 — Financial Reconciliation Ops Spec

**Status**: Draft  
**Perspective**: Finance Ops + System Analyst + Engineering  
**Primary Inputs**: `ADR-S009`, `ADR-S051`, `ADR-K011`, `14-koperasi-regulated-finance-guardrails.md`

## Purpose

Mendefinisikan model operasional rekonsiliasi keuangan lintas invoice, payment gateway, settlement, kas, dan transaksi koperasi agar angka operasional, akuntansi, dan evidence audit tetap selaras.

## Reconciliation Scope

### School Finance
- invoice vs payment receipt
- payment transaction vs callback
- paid vs settled amounts
- outstanding vs waiver vs refund states

### Cooperative Finance
- rekening balance vs immutable transaction log
- teller cash movement vs transaction records
- reversal/correction trace
- daily close and exception handling

### Cross-System / External
- provider settlement vs internal paid transactions
- bank statement vs settlement batch
- manual transfer verification vs recorded payment

## Reconciliation Cadence

- near-real-time checks untuk callback/payment status
- daily reconciliation untuk settlement, teller, dan exception queues
- period-end reconciliation untuk finance reporting dan audit pack

## Exception Types

- callback received but transaction not matched
- paid but invoice not updated
- settled amount mismatch
- duplicate payment / duplicate callback
- teller imbalance
- ledger mismatch after correction/reversal

## Operational Rules

- semua mismatch harus masuk ke exception queue yang punya owner dan SLA
- reversal/koreksi tidak menghapus jejak transaksi asli
- reconciliation result harus punya status: matched, partially_matched, unmatched, disputed, resolved
- financial close tidak boleh dianggap final jika exception material masih terbuka tanpa waiver

## Acceptance Criteria

- `AC-FUNC`: tim ops dapat menjalankan rekonsiliasi rutin dan menyelesaikan mismatch dengan workflow jelas
- `AC-DATA`: setiap angka yang direkonsiliasi dapat ditelusuri ke source transaction/invoice/settlement
- `AC-AUTH`: correction, write-off, reversal, dan manual match dibatasi oleh approval yang tepat
- `AC-AUDIT`: semua keputusan exception dan resolution trail tercatat
- `AC-INT`: provider callback, settlement import, dan cooperative transaction log dapat digabung untuk evidence yang cukup
- `AC-NFR`: proses rekonsiliasi tidak mengganggu throughput transaksi utama secara signifikan

## Open Questions

- tingkat otomatisasi auto-match vs manual-review final
- policy cut-off dan materiality threshold per tenant/segmen
