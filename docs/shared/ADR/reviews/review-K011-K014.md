# C-Suite Advisory Review: ADR K011-K014

**Review Date**: 2026-04-15
**Reviewers**: CEO, CFO, CTO, COO, CMO (Advisory Panel)
**Scope**: Transaction Engine, Teller Session, Money Denomination, Kas & Cash Flow
**Cross-reference**: K001 (Nasabah - KYC, Specimen), K002 (Rekening - Balance, Hold Amount)

---

## K011: Transaksi Rekening & Non-Rekening

**Verdict: APPROVED WITH ADJUSTMENTS**

### CEO Perspective
K011 is the operational backbone of the entire koperasi system. The taxonomy is comprehensive and covers both conventional and BMT modes cleanly. The separation of rekening vs non-rekening transactions is strategically sound — it keeps member-facing operations distinct from operational expenses. The immutability decision aligns with regulatory requirements and builds trust with school stakeholders (parents, teachers, administrators). Priority recommendation: ensure the validation pipeline (section 4) does not create perceptible latency for tellers during peak hours (school fee collection days).

### CFO Perspective
**Strengths:**
- KYC limit enforcement (V3) correctly references K001 section 11 and calculates SUM across all nasabah accounts (not per-rekening) — this is the correct interpretation
- Specimen verification (V6) correctly references K001 section 10 with configurable thresholds
- `balance_before` and `balance_after` snapshots provide point-in-time audit trail
- CHECK constraint `balance_after = balance_before + amount` is an excellent database-level integrity guard
- Non-rekening transactions properly segregated in `kas_transaksi` table — clean separation from member transactions

**Issues:**
1. **[MEDIUM] Reversal balance validation gap**: Section 8 states "reversal mengikuti validasi yang sama (check balance, etc.) — jika saldo sudah habis, reversal bisa ditolak." This creates a financial integrity problem. If a deposit of Rp 500,000 was made in error and the member has already withdrawn, the reversal cannot execute. The ADR needs to define what happens next — does the member incur a negative balance? Is a receivable created? This edge case must be explicitly resolved.
2. **[LOW] Batch partial failure accounting**: Section 9 handles partial failures (skip failed items), but the summary report needs to include the exact failed items with reasons for audit. Currently it mentions "success count, failed count, total amount processed" but should also require "failed amount, failed item details" for reconciliation.
3. **[LOW] Transfer KYC double-counting**: Section 5 states KYC limit is checked for both sender and receiver. Clarify: does the transfer amount count toward the monthly limit of both parties? If yes, a transfer is "double-taxing" the KYC limit. This seems correct for regulatory purposes but should be explicitly stated.

### CTO Perspective
**Strengths:**
- `SELECT ... FOR UPDATE` row-level lock is the correct approach for concurrent balance updates — prevents race conditions without table-level locking
- Atomic transaction (INSERT transaksi + UPDATE rekening + UPDATE teller_session in single DB transaction) is properly designed
- Database trigger to reject UPDATE/DELETE on transaksi table is defense-in-depth beyond application layer
- Single exception (`is_reversed` field) is well-scoped and prevents the need for a separate reversal tracking table
- Index strategy is comprehensive (per rekening, per nasabah, per branch, per session, per batch)

**Issues:**
1. **[HIGH] Deadlock risk on transfers**: Section 5 shows transfer locking Rekening A then Rekening B. If a concurrent transfer goes B-to-A, classic deadlock occurs. The ADR must mandate **consistent lock ordering** — always lock the rekening with the smaller UUID first, regardless of source/destination direction.
2. **[MEDIUM] Batch concurrency control**: Section 9 says batch items are processed individually. If a batch deduction hits the same rekening that a teller is processing simultaneously, the row-level lock will serialize them correctly, but latency may spike. Consider documenting that batch processing should be scheduled during non-peak hours or use advisory locks to prevent concurrent batch + teller operations on the same rekening.
3. **[MEDIUM] Vernon SyncEngine load on NasabahUpdatedEvent**: Section 14 notes that a name change triggers propagation to all transactions. For a long-running member with thousands of transactions, this is a heavy background job. The ADR acknowledges this in Mitigasi but should specify: what happens if sync fails midway? Is there a retry mechanism? Is there a staleness indicator on `_data`?
4. **[LOW] Sequence number collision under high concurrency**: Section 6 uses daily-scoped sequence. The ADR should specify the mechanism (database sequence, application-level counter, or UUID-based) to prevent collision under concurrent transaction creation.

### COO Perspective
**Strengths:**
- 7-layer validation pipeline is sequential but returns ALL failures — excellent UX for tellers who can fix multiple issues at once
- Receipt auto-generation with reprint capability (marked "SALINAN") is practical for school koperasi where parents may request copies
- RBAC is sensible: tellers handle daily operations, reversals require Manager+ approval

**Issues:**
1. **[MEDIUM] Specimen verification workflow unclear for tellers**: V6 says teller must verify specimen for withdrawals above threshold, but the ADR does not specify the UX — does the system show the specimen image side-by-side? Can the teller override without Supervisor? What if the specimen is missing (not uploaded yet)?
2. **[LOW] Offline/degraded mode not addressed**: School koperasi in remote areas may have intermittent connectivity. What happens if the system is down during peak transaction hours? Is there a manual fallback process defined? This is critical for school environments.

### CMO Perspective
- Receipt design (section 7) is well-thought-out with configurable header/footer/logo per tenant — good for branding
- Dual-mode terminology (section 16) ensures Islamic BMT members see appropriate labels without any code branching
- The validation error returning all failures (not just first) is excellent stakeholder experience

### Missing Items
1. **Transaction timeout**: No mention of what happens if a database transaction hangs — timeout configuration and cleanup
2. **Idempotency key**: No mention of idempotency for transaction creation — what if teller double-clicks submit? Need idempotency key or deduplication window
3. **Rate limiting**: No mention of rate limiting per teller/session to prevent abuse
4. **Transaction receipt delivery**: No mention of digital receipt (WhatsApp/SMS) for tech-savvy school stakeholders — only print

---

## K012: Teller Session

**Verdict: APPROVED WITH ADJUSTMENTS**

### CEO Perspective
Teller session is the daily operational foundation. The lifecycle (OPEN -> SUSPENDED -> CLOSED) is clean and practical. Dual-control on opening (Supervisor hands over + Teller confirms) builds accountability culture. This is well-suited for school koperasi where trust and transparency are paramount for parent confidence.

### CFO Perspective
**Strengths:**
- Expected cash formula is correct: `kas_awal + total_cash_in - total_cash_out`
- Closing section 4 adds `kas_masuk_netto` (operational cash in/out via teller) — comprehensive
- Variance handling with tiered thresholds (minor/major) and mandatory explanation is good financial control
- CHECK constraint on expected_cash ensures data integrity at database level
- Daily consolidation feeds directly into K014 — clean chain of custody

**Issues:**
1. **[MEDIUM] Expected cash formula inconsistency**: Section 3 shows `expected_cash = kas_awal + cash_in - cash_out`, but section 4 closing formula adds `+ kas_masuk_netto`. These must be consistent. The running total (section 3) should ALSO include non-rekening cash movements (kas_keluar for ATK, etc.) if the teller processes those. If non-rekening transactions go through the teller session, the running total must reflect them. Otherwise, closing will always show a variance for branches where tellers also handle operational cash.
2. **[LOW] Variance sign convention**: Section 5 defines `selisih = expected - actual`. Positive = cash short (expected more than found). This is correct and standard. Just ensure UI clearly labels "KURANG" (short) and "LEBIH" (over) rather than showing raw positive/negative numbers that confuse tellers.

### CTO Perspective
**Strengths:**
- One active session per teller constraint prevents accounting ambiguity
- SUSPENDED state with auto-lock timeout is practical and secure
- Force close with mandatory denomination counting maintains integrity even in exceptional cases
- UNIQUE constraint on (tenant_id, teller_user_id, session_date, session_number) is correct

**Issues:**
1. **[MEDIUM] Race condition on session status change**: When teller initiates close (step 2: "sistem block transaksi baru"), the ADR does not specify the locking mechanism. Between the teller clicking "close" and the system setting CLOSING status, a concurrent transaction could slip through. Need a status transition like `OPEN -> CLOSING -> CLOSED` where CLOSING blocks new transactions, or use `SELECT ... FOR UPDATE` on the session row before allowing any new transaction.
2. **[LOW] Suspended session timeout enforcement**: Section 7 mentions configurable max suspend duration (default 60 min), but does not specify what happens when exceeded. Does it auto-close? Does it alert Supervisor? Does it stay suspended indefinitely until someone acts? The ADR says "Supervisor harus intervene" but this is vague.

### COO Perspective
**Strengths:**
- Suspend/Resume is practical for school koperasi where the teller might be a teacher who steps away for class
- Template for denomination default speeds up repetitive opening (mentioned in Mitigasi)
- Force close handles the real-world scenario of a teller who forgets or is absent

**Issues:**
1. **[HIGH] Opening process too rigid for small koperasi**: In a school koperasi with 1 teller and 1 Supervisor (who might also be the kepala koperasi), requiring Supervisor to physically hand over cash and teller to confirm creates a bottleneck. If the Supervisor is absent, no session can be opened and no transactions can occur. Need a fallback: allow Manager to also hand over cash, or define a simplified process for single-teller branches.
2. **[MEDIUM] Closing time pressure**: Section 4 blocks new transactions when teller initiates close. In a school koperasi, the closing might happen during a busy period (end of school day). The ADR should recommend that closing should be initiated only after the last customer is served, and the blocking should be clearly communicated to other staff. Consider a "soft close" (stop accepting new queue numbers) vs "hard close" (block system).
3. **[LOW] Denomination counting burden**: For a small school koperasi handling Rp 5-10 million daily, counting 11 denominations twice daily (open + close) adds significant overhead. Consider allowing a "quick close" mode where only total is entered for sessions below a configurable threshold (with full counting required for larger amounts). This is acknowledged in Mitigasi but should be a formal option, not just a workaround.

### CMO Perspective
- Dashboard showing all active teller sessions per branch (section 6) gives Supervisor real-time visibility — builds confidence
- Variance tracking per teller creates accountability culture — but communicate it positively (accuracy metric) not punitively

### Missing Items
1. **Session handover**: No mechanism for mid-shift handover between tellers. If Teller A is sick mid-shift, the only option is force close + new session for Teller B. A formal handover (Teller A counts, Teller B verifies, session transfers) would be operationally smoother.
2. **Session history/audit view for teller**: Can the teller view their own historical sessions and variance patterns? This supports self-improvement.
3. **Concurrent session limit per branch**: No mention of maximum concurrent sessions per branch — relevant for licensing or hardware constraints (e.g., only 2 cash drawers available).

---

## K013: Money Denomination

**Verdict: APPROVED**

### CEO Perspective
Clean, focused ADR that serves its purpose as a supporting module for K012 and K014. The future-proofing for new BI denominations is appropriate. Per-tenant master is correctly justified for multi-tenant flexibility.

### CFO Perspective
**Strengths:**
- Denomination-level tracking enables fraud detection (same total, different composition)
- Vault management with periodic counting and variance investigation is standard banking practice
- Vault balance is derived (not stored) from movements — auditable and reconcilable
- Historical data preserved when denominations are deactivated — no data loss

**Issues:**
1. **[LOW] Vault count frequency not specified**: Section 7 says "periodik" but does not define minimum frequency. For regulatory compliance, recommend specifying: daily spot-check (optional), weekly full count (recommended), monthly mandatory count (required with Manager sign-off).
2. **[LOW] Vault count denomination table missing denomination_type**: `vault_count_denomination` (section 9) has `denomination_type` but `teller_session_denomination` (K012 section 11) does not. For the Rp 1,000 banknote vs coin distinction, both tables need `denomination_type`. This is an inconsistency.

### CTO Perspective
**Strengths:**
- Separate table for denominations (not JSONB) enables SQL aggregation — correct decision
- `subtotal = denomination_value * quantity` with CHECK constraint prevents calculation errors
- Denomination master per tenant with soft delete (is_active) is clean

**Issues:**
1. **[LOW] No FK to denomination master**: `teller_session_denomination` and `vault_count_denomination` use `denomination_value` (integer) as the denomination identifier, not a FK to `denominasi_master.id`. This means: (a) orphaned denomination values are possible, (b) denomination label/type changes do not propagate. Consider adding `denomination_id UUID FK` alongside `denomination_value` for referential integrity, while keeping `denomination_value` as a denormalized field for query convenience.

### COO Perspective
**Strengths:**
- UI optimization suggestions (keypad entry, auto-calculate, pre-fill) are practical
- Denomination template deferred to Phase 2 is a sensible prioritization

**Issues:**
1. **[MEDIUM] Coin counting is unrealistic for daily operations**: In practice, school koperasi rarely deal with coins below Rp 500. The default active denominations should exclude Rp 100 and Rp 200 coins (set `is_active = false` in seed data), with the option to activate if needed. This reduces counting from 11 entries to 8 entries — meaningful time savings for daily operations.

### CMO Perspective
- Denomination counting adds perceived professionalism and trust — parents seeing detailed denomination records on their receipt builds confidence in the koperasi

### Missing Items
1. **Denomination template for common scenarios**: Phase 2 item mentioned but not tracked in a backlog. Recommend creating a backlog entry.
2. **Barcode/QR on denomination summary**: For physical verification, a QR code linking to the digital denomination record could speed up audit.

---

## K014: Kas & Cash Flow

**Verdict: APPROVED WITH ADJUSTMENTS**

### CEO Perspective
This ADR completes the cash management chain from individual transactions (K011) through teller accountability (K012) to daily branch-level position. The multi-branch consolidation view (section 9) is essential for koperasi management oversight. The decision not to block transactions when cash is below minimum (section 6) is the correct member-first approach.

### CFO Perspective
**Strengths:**
- Daily cash position formula is correct and derived from transactions — not manual entry
- CHECK constraints on kas_harian ensure `closing = opening + inflow - outflow` at database level
- Sub-category CHECK constraints (`total_inflow = sum of sub-categories`) prevent data corruption
- Inter-branch transfer with dual-side recording and `transfer_kas_pair_id` linkage is proper
- Status tracking for inter-branch transfer (initiated -> in_transit -> received -> confirmed/disputed) covers the physical transfer lifecycle
- Reconciliation kas vs teller sessions is automated with clear escalation path
- Non-blocking on kas minimum is the right call — operational issue should not penalize members

**Issues:**
1. **[HIGH] Reconciliation formula incomplete**: Section 7 shows `closing_balance = total_actual_cash + non_cash_balance + selisih_teller`. But `non_cash_balance` is not defined anywhere in this ADR or K012. What is "non_cash_balance"? If it refers to bank account balances, that is a different domain. If it refers to non-cash transactions (transfers), those do not affect physical cash. This formula needs clarification or the term needs to be defined. The simpler reconciliation should be: `closing_balance (from kas_harian) = SUM(teller actual_cash at close) + vault_end_of_day_balance + operational_cash_movements_outside_teller_sessions`. Define each component explicitly.
2. **[MEDIUM] Inter-branch transfer denomination mismatch handling**: Section 4 says denomination must match between sender and receiver. But in reality, branch B may count and find a different denomination composition (e.g., some bills were swapped during transit, or counting error). The ADR handles total amount mismatch (investigation) but does not address denomination-level mismatch with same total. This should be flagged but not necessarily blocking.
3. **[MEDIUM] Kas harian opening balance chain integrity**: Section 1 says "saldo awal hari ini = saldo akhir kemarin." What happens if yesterday's kas_harian record does not exist (system was down, branch was closed, holiday)? The chain must handle gaps gracefully — either carry forward the last available closing balance or require Manager intervention.
4. **[LOW] Non-rekening transaction attachment**: Section 3 mentions `attachment_url` for receipt/nota but this field is not in the `kas_transaksi` data model (K011 section 12). Add `attachment_url VARCHAR (nullable)` to the kas_transaksi table.

### CTO Perspective
**Strengths:**
- Denormalized kas_harian with sub-category columns (not JSONB) enables efficient SQL aggregation
- UNIQUE constraint (tenant_id, branch_id, report_date) prevents duplicate daily records
- Auto-generation of daily report when all teller sessions are closed is event-driven — clean

**Issues:**
1. **[MEDIUM] Inter-branch transfer state machine not in data model**: Section 4 defines states (initiated, in_transit, received, confirmed, disputed) but the `kas_transaksi` table (K011 section 12) only has `approval_status (none, pending, approved, rejected)`. These are different state machines. The transfer lifecycle states need to be added to the data model — either as a new field on `kas_transaksi` or as a separate `kas_transfer` table.
2. **[LOW] Scheduled job for reconciliation**: Section 7 mentions "scheduled job" for daily reconciliation but no mechanism is defined. Specify: is it a cron job? Event-driven (triggered when last session closes)? What if it fails? Retry policy?

### COO Perspective
**Strengths:**
- Non-blocking kas minimum with alert is the right approach for school koperasi
- Kas minimum configurable per branch accommodates different branch sizes
- Cash flow report with auto-generate daily + manual on-demand is practical

**Issues:**
1. **[MEDIUM] Daily kas closing dependency on all teller sessions**: Section 5/section 7 state that daily report auto-generates when ALL teller sessions are closed. If one teller forgets to close (common in school koperasi), the entire branch daily report is delayed. Need a timeout mechanism: if a session is not closed by configurable time (e.g., 22:00), send escalation alert. If not closed by next morning, auto-flag and allow kas_harian to generate with that session marked as "pending_close".
2. **[LOW] Cash flow report export permissions**: Supervisor can export, which is appropriate. But consider also allowing a read-only summary view for the school principal (Kepala Sekolah) who may have oversight but no koperasi role — this may need a dedicated "Observer" role in the future.

### CMO Perspective
- Multi-branch consolidation dashboard with color-coded status (OK/LOW/CRITICAL) is excellent for management oversight
- Daily, weekly, monthly report cadence meets different stakeholder needs (daily for Supervisor, monthly for RAT/annual meeting)

### Missing Items
1. **End-of-year closing**: No mention of annual closing process, carry-forward to new fiscal year, or integration with annual audit (RAT - Rapat Anggota Tahunan)
2. **Cash insurance/security**: No mention of physical cash security considerations — vault insurance, maximum cash holding per branch, cash-in-transit insurance for inter-branch transfers
3. **Holiday/weekend handling**: What happens on non-operational days? Is there a kas_harian record with zero movement, or is it skipped? This affects the opening balance chain.

---

## Cross-Cutting Concerns

### 1. Immutability Consistency (CTO)
All four ADRs maintain consistent immutability principles. K011 enforces via database triggers. K012 teller sessions are mutable (status changes) but the critical financial fields (opening_amount, totals at close) become effectively immutable after closing. K013 denomination records are INSERT-only per phase. K014 kas_harian is mutable until reconciled. **Recommendation**: Add explicit immutability rules for K012 (no edit after CLOSED) and K014 (no edit after RECONCILED) with database-level enforcement similar to K011.

### 2. Offline/Degraded Mode (COO)
None of the four ADRs address what happens when the system is unavailable. For a school koperasi in Indonesia (potentially in areas with unreliable connectivity), this is a real operational risk. **Recommendation**: Define a manual fallback procedure — paper-based transaction log that can be batch-entered when the system is restored. This does not need to be in the ADR but should be in an operational manual (SOP) referenced by these ADRs.

### 3. End-to-End Data Flow Integrity (CFO)
The chain is: Transaction (K011) -> Teller Session Running Total (K012) -> Daily Kas (K014), with K013 providing denomination detail at K012 boundaries. The CHECK constraints at each level provide defense-in-depth. **Gap**: There is no mention of a periodic full reconciliation that verifies the entire chain (SUM of all K011 transactions = K014 closing balance). The daily reconciliation (K014 section 7) compares kas vs teller sessions but not kas vs sum-of-all-transactions. Add a weekly/monthly deep reconciliation.

### 4. Performance Under Load (CTO)
School fee collection days (awal bulan, awal semester) may see high transaction volume across multiple tellers. The row-level locking strategy (K011) is correct, but batch operations (simpanan wajib collection) running simultaneously with teller transactions could cause contention. **Recommendation**: Document that batch operations should be scheduled outside peak hours, or use read-committed isolation with retry logic for batch items that encounter lock contention.

### 5. Audit Trail Completeness (CFO)
K011 has comprehensive audit trail. K012 has audit fields but force_close and variance events should also generate dedicated audit log entries (not just field updates). K013 is audit-light (only created_at, created_by) which is sufficient for denomination records. K014 reconciliation events need audit log entries. **Recommendation**: Standardize audit event generation across all four ADRs — every state transition and approval should generate an audit event.

### 6. Idempotency (CTO)
K011 does not define idempotency keys for transaction creation. In a school koperasi environment (potentially slow network, impatient tellers double-clicking), duplicate transaction submission is a real risk. **Recommendation**: Add an idempotency_key field to the transaction creation API, with a deduplication window (e.g., same key within 5 minutes = reject duplicate). This is critical for financial integrity.

---

## Recommended Adjustments (Priority Order)

### Must Fix Before Implementation

| # | ADR | Severity | Issue | Fix |
|---|-----|----------|-------|-----|
| 1 | K011 | HIGH | Deadlock risk on transfers | Mandate consistent lock ordering (smaller UUID first) in section 5 |
| 2 | K014 | HIGH | Reconciliation formula uses undefined `non_cash_balance` | Rewrite section 7 with explicitly defined components |
| 3 | K011 | HIGH | No idempotency mechanism for transaction creation | Add idempotency_key to transaction creation API spec |

### Should Fix Before Implementation

| # | ADR | Severity | Issue | Fix |
|---|-----|----------|-------|-----|
| 4 | K012 | MEDIUM | Race condition on session close (transaction slip-through) | Add CLOSING intermediate status or use FOR UPDATE on session |
| 5 | K012 | MEDIUM | Opening requires Supervisor — blocks single-teller branches | Allow Manager as alternative for kas handover |
| 6 | K011 | MEDIUM | Reversal when balance insufficient — undefined behavior | Define explicit resolution path (negative balance, receivable, or block) |
| 7 | K014 | MEDIUM | Inter-branch transfer state machine not in data model | Add transfer_status field to kas_transaksi or create kas_transfer table |
| 8 | K014 | MEDIUM | Opening balance chain breaks on holidays/system downtime | Define gap-handling: carry forward last closing or require manual intervention |
| 9 | K012 | MEDIUM | Expected cash formula inconsistency between section 3 and section 4 | Unify formula to include non-rekening cash movements in running total |
| 10 | K014 | MEDIUM | Daily report blocked by single unclosed teller session | Add timeout + escalation mechanism |

### Nice to Have

| # | ADR | Severity | Issue | Fix |
|---|-----|----------|-------|-----|
| 11 | K013 | LOW | teller_session_denomination missing denomination_type | Add denomination_type ENUM to K012 section 11 |
| 12 | K013 | LOW | No FK from denomination records to denominasi_master | Add denomination_id FK for referential integrity |
| 13 | K014 | LOW | attachment_url mentioned but not in kas_transaksi model | Add field to K011 section 12 data model |
| 14 | K013 | MEDIUM | Default seed includes rarely-used coin denominations | Set Rp 100, Rp 200 coins as is_active=false in seed data |
| 15 | ALL | LOW | No offline/degraded mode SOP reference | Add reference to operational manual for manual fallback |
| 16 | K011 | LOW | No digital receipt channel (WhatsApp/SMS) | Add to future roadmap, not blocking for MVP |

---

## Final Summary

The K011-K014 ADR set forms a solid, well-integrated operational backbone for the koperasi/BMT system. The architecture correctly prioritizes immutability, atomicity, and auditability — the three pillars of financial system integrity. The separation of concerns (transactions, teller sessions, denominations, daily cash position) is clean and each ADR has clear boundaries and integration points.

The three HIGH-severity items (deadlock prevention, reconciliation formula clarity, and idempotency) must be addressed before implementation begins, as they directly impact financial data integrity. The MEDIUM items are important for operational robustness, particularly for the school koperasi context where staff are not full-time banking professionals.

Overall, this is production-grade financial architecture that respects both regulatory requirements and the practical realities of running a school cooperative.

**Panel Decision: APPROVED WITH ADJUSTMENTS** (address HIGH items before implementation, MEDIUM items before go-live)
