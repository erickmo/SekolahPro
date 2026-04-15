# C-Suite Advisory Review: ADR K007-K010

**Review Date**: 2026-04-15
**Reviewers**: CEO, CFO, CTO, COO, CMO (Advisory Panel)
**Scope**: Pinjaman, Angsuran & Jadwal, Denda & Penalti, Jaminan & Agunan
**Domain Risk**: CRITICAL (highest risk domain in koperasi/BMT)

---

## K007: Pinjaman / Pembiayaan

**Verdict**: APPROVED WITH ADJUSTMENTS

**CEO**: Strong strategic foundation. Multi-level approval and dual-mode support positions the platform well for both conventional koperasi and BMT markets. The school-context awareness (guru salary-based scoring, siswa restriction) is excellent differentiation. Concern: the 2x restructuring limit may alienate koperasi that deal with prolonged economic hardship among members.

**CFO**:
- NPL classification (Kol 1-5) aligns with POJK for LKM. DPD thresholds (0, 1-90, 91-120, 121-180, >180) are correct per OJK standards for micro-finance institutions. However, the current Kol-2 threshold of 1-90 DPD is very wide. OJK POJK No. 12/POJK.05/2014 for LKM uses 1-90 for Kol-2, so technically correct, but consider splitting internal monitoring buckets more granularly (1-30, 31-60, 61-90) for earlier intervention. The aging report in K008 already does this which is good.
- BMPK (Batas Maksimum Pemberian Kredit) is NOT explicitly addressed. For koperasi, POJK limits single borrower exposure to a percentage of modal/equity. This is a critical gap. The DTI check and salary multiplier are good but do not replace BMPK which looks at the koperasi's capital, not the borrower's income.
- Provisioning (PPAP) is mentioned but deferred to K015. This is acceptable but the ADR should explicitly state the PPAP rates per kolektibilitas level: Kol-1 = 0.5%, Kol-2 = 10%, Kol-3 = 50%, Kol-4 = 75%, Kol-5 = 100%. Without this, implementation may use wrong rates.
- Credit scoring formula (salary x multiplier + simpanan bonus) is simple but adequate for school koperasi context. DTI max 40% is standard and correct.
- Ibra (Islamic early settlement discount) is correctly handled: sisa margin yang belum earned di-diskon. This is DSN-MUI compliant.
- Write-off requires PPAP 100% coverage before approval, which is the correct conservative approach.
- Missing: no mention of BMPK group (aggregate exposure to related parties, e.g., multiple family members borrowing).

**CTO**:
- Loan state machine has a gap: RESTRUCTURED is described as both a status and a non-terminal state. The ADR says "status kembali ke ACTIVE setelah restrukturisasi" which is contradictory. If RESTRUCTURED is a transitional flag, use `is_restructured` boolean (which already exists) instead of a status enum value. This creates ambiguity in code: does the loan status column ever hold RESTRUCTURED? Or does it go ACTIVE -> (restructure process) -> ACTIVE with `is_restructured=true`? Clarify.
- Missing transition: there is no ACTIVE -> DEFAULTED or equivalent pre-write-off status. Going directly from ACTIVE to WRITTEN_OFF may miss intermediate steps where collection is intensified.
- Concurrent loan state changes: if a scheduled NPL job and a payment transaction happen simultaneously, there could be race conditions on `collectibility` and `dpd` fields. The ADR should specify optimistic locking or database-level serialization for these updates.
- Data model is comprehensive. The `pinjaman_application` with separate `analyzing` status in the workflow is good for tracking credit analysis state.
- Disbursement is single-shot only (no partial), which is correct simplification for school koperasi.

**COO**:
- Multi-level approval (Teller -> Supervisor analysis -> Manager approval) adds 2 handoffs. For a small koperasi with 1-2 staff, this is impractical. The fast-track for emergency loans (Manager direct) partially addresses this, but regular loans for guru wanting Rp 5 juta should not require 3 people.
- Recommendation: allow configurable approval flow where small koperasi can set Supervisor as final approver for loans below a threshold (not just credit analysis).
- Restructuring is well-designed with clear limits and approval chain.
- Payroll deduction integration with school system is operationally excellent for reducing NPL.

**CMO**:
- The application flow from nasabah perspective requires visiting the koperasi office (via Teller). No self-service or mobile pre-application mentioned. For adoption among busy guru/ustadz, a mobile pre-fill would reduce friction significantly.
- Transparency is good: nasabah can review schedule before disbursement.
- The siswa restriction (parent must be borrower) should be clearly communicated in product eligibility UI to avoid frustration.

**Issues Found**:

1. **[HIGH]** BMPK (single borrower limit as % of koperasi capital) is missing entirely. Without this, a koperasi could over-concentrate risk on a few large borrowers and become insolvent. **Fix**: Add BMPK check to credit pre-check. BMPK for koperasi is typically 20% of modal sendiri per single borrower, 25% for group. Make configurable per tenant.

2. **[HIGH]** RESTRUCTURED status ambiguity in state machine. The loan has both a status enum value `restructured` AND a boolean `is_restructured` + `restructure_count`. These overlap. **Fix**: Remove RESTRUCTURED from the status enum. A restructured loan stays ACTIVE with `is_restructured=true`. The old loan terms are preserved in audit/history. The current dual representation will cause bugs.

3. **[MEDIUM]** PPAP provisioning rates not specified. K015 is referenced but not yet written. **Fix**: Add a note in K007 specifying the standard PPAP rates (0.5%, 10%, 50%, 75%, 100%) to prevent incorrect implementation before K015 is finalized.

4. **[MEDIUM]** No group borrower exposure limit. Multiple family members (guru husband + guru wife + parent) could collectively over-borrow from the same koperasi. **Fix**: Add optional group exposure tracking configurable per tenant.

5. **[LOW]** `installment_day` constrained to 1-28 in data model but due date config allows fixed_day up to 31 with month-end fallback. The constraint should be documented consistently.

**Missing Items**:
- Disbursement reversal process (what if disbursement was done in error?)
- Cooling-off period for borrower (hak pembatalan sebelum pencairan)
- Insurance/takaful integration for credit protection (common in BMT)
- BMPK calculation formula and enforcement point

---

## K008: Angsuran & Jadwal

**Verdict**: APPROVED WITH ADJUSTMENTS

**CEO**: Comprehensive calculation engine supporting 7 methods (3 conventional + 4 Islamic) is a strong competitive advantage. Multi-channel payment (teller, auto-debit, payroll) reduces NPL risk which protects koperasi sustainability.

**CFO**:
- Flat calculation is correct: Total Bunga = P x r x (T/12), angsuran = (P + Total Bunga) / T. Verified.
- Declining calculation is correct: Bunga per bulan = Sisa Pokok x (r/12). Verified.
- Annuity formula is correct: A = P x r x (1+r)^n / ((1+r)^n - 1). The worked example shows 1,066,185 which is correct (verified: 12,000,000 x 0.01 x 1.12682503 / 0.12682503 = 1,066,185.15, rounds to 1,066,185). Verified.
- Murabahah = flat (same formula, different legal basis). Correct per DSN-MUI.
- Payment allocation priority (penalty -> interest/margin overdue -> principal overdue -> interest/margin current -> principal current) is the standard banking waterfall and is correct. Important: for Islamic mode, ta'zir must still route to social fund even when collected through this waterfall. Ensure fund routing in K009 is triggered post-allocation.
- Rounding strategy: "bulatkan ke bawah untuk pokok, sisa masuk ke angsuran terakhir" is correct. This prevents cumulative rounding from leaving residual balance.
- Underpayment carry-forward: the ADR says sisa 320,000 is "ditambahkan ke total_amount" of the next installment. This is problematic. The carried-forward amount should be tracked separately, not mutated into the scheduled amount, because the scheduled amount is supposed to be immutable post-disbursement. The jadwal should remain as-is; the unpaid portion simply remains as outstanding on the current installment.

**CTO**:
- Schedule generation edge cases need attention: (a) leap year February handling for fixed_day=29, (b) end-of-month for fixed_day=30/31 in February, (c) timezone handling for DPD calculations crossing midnight.
- The due_date_config with holiday_adjustment is well-designed. However, the holiday calendar itself is not defined (data model/table missing). Where is the holiday list stored? This needs a `tenant_holiday` table or reference to a national holiday API.
- DPD calculation uses calendar days (not business days) per K009 Section 3. Consistent. Good.
- Race condition concern: auto-debit scheduled job and teller manual payment could collide on the same installment. If the auto-debit runs at 08:00 and a teller processes payment at 08:01, the installment could be double-paid. **Fix**: Use database row-level locking on the angsuran record during payment processing. Add a `payment_lock` mechanism or rely on SELECT FOR UPDATE.
- The `angsuran_pembayaran` table for audit trail is well-designed, allowing multiple partial payments per installment.
- Missing: no `VOIDED` status in the payment_status enum for rescheduled installments. The ADR mentions VOIDED in rescheduling impact (Section 12) but the data model enum is `(scheduled, partial, paid, overdue, waived)`.

**COO**:
- Three payment channels (teller, auto-debit, payroll) cover all practical scenarios for school koperasi. Auto-debit with 3-day retry is sensible.
- Payroll deduction dependency on school system is a risk. If the school changes payroll software or process, the integration breaks. The file-exchange (CSV/API) mitigation is practical.
- Reconciliation monthly cadence is sufficient. The Rp 1 threshold for alerts is extremely tight -- this will generate many false alerts from rounding. Recommend Rp 100 as default threshold.

**CMO**:
- Payment receipt printing (bukti pembayaran) is mentioned for teller. Digital receipt via SMS/WhatsApp should also be supported for better member experience.
- Nasabah should have visibility into their payment schedule and history (read-only portal or printout).

**Issues Found**:

1. **[HIGH]** VOIDED status missing from `payment_status` enum in data model but referenced in rescheduling flow. **Fix**: Add `voided` to the ENUM: `(scheduled, partial, paid, overdue, waived, voided)`.

2. **[HIGH]** Concurrent payment race condition. Auto-debit job and teller payment can process the same installment simultaneously. **Fix**: Mandate SELECT FOR UPDATE on angsuran row before any payment processing. Document this in the ADR.

3. **[MEDIUM]** Underpayment carry-forward described as mutating next installment's total_amount. This breaks the immutability principle stated in Section 1 ("jadwal bersifat immutable setelah disbursement"). **Fix**: Unpaid portion stays as outstanding on the current installment (payment_status = PARTIAL). Do not modify the next installment's scheduled amounts. The carry-forward should be logical (FIFO payment matching handles it naturally), not physical mutation of data.

4. **[MEDIUM]** Holiday calendar data model not defined. **Fix**: Add `tenant_holiday` table or reference to holiday configuration.

5. **[LOW]** Reconciliation alert threshold of Rp 1 will generate noise from rounding. **Fix**: Increase default to Rp 100 (configurable).

**Missing Items**:
- Bulk payment processing (batch import for payroll deduction results)
- Payment reversal/refund process (erroneous payment correction)
- SMS/digital notification before due date (reminder mechanism)
- Ijarah end-of-term handling (buy option for IMBT not detailed in schedule generation)

---

## K009: Denda & Penalti

**Verdict**: APPROVED WITH ADJUSTMENTS

**CEO**: Excellent dual-mode separation. The fund routing for ta'zir to social fund vs. koperasi income is a strong compliance differentiator for BMT market. Penalty cap protects nasabah and builds trust.

**CFO**:
- Ta'zir must be fixed amount (not percentage) and must go to social fund (baitul maal), not koperasi income. This is correctly specified per Fatwa DSN-MUI No. 17/DSN-MUI/IX/2000. Verified and compliant.
- Ta'widh (compensation for actual loss) correctly requires proof and is correctly routed to koperasi income (as it compensates real cost, not punishment). Compliant.
- Ta'widh examples correctly exclude opportunity cost/bunga -- this would be riba. Verified.
- Penalty cap at 25% of principal is reasonable and protects nasabah. Some regulators suggest lower (e.g., OJK for fintech at 100% of principal, but koperasi best practice is lower). 25% is conservative and good.
- Early withdrawal penalty: 50% of earned interest/profit is standard. Verified.
- Early settlement: Islamic mode correctly has NO penalty (ibra instead). This is DSN-MUI compliant. Verified.
- Grace period denda (3 days default) is separate from grace period angsuran pertama (K007). This is correct and well-documented.
- The penalty calculation example is correct: 1,120,000 x 0.001 x 15 = 16,800. Verified.
- Concern: The `calculation_basis` option of `outstanding_principal` for late penalty is extremely harsh -- denda calculated from entire remaining principal, not just the overdue amount. This should be restricted or flagged as high-risk configuration. A nasabah who misses one Rp 1 juta installment but has Rp 50 juta outstanding would get penalized on the full Rp 50 juta.

**CTO**:
- Scheduled daily penalty calculation job is the right pattern (vs real-time). Idempotent re-runs are implied but should be explicitly stated.
- The data model is thorough with proper separation of calculated_amount, capped_amount, final_amount, and outstanding_amount.
- `fund_destination` enum (koperasi_income, social_fund) with automatic routing based on coop_type and penalty_type is clean.
- Missing: what happens to accrued denda when a loan is restructured? Is it waived? Carried forward? Frozen? The interaction between K009 penalties and K007 restructuring is not documented.
- Missing: penalty calculation for PARTIAL payment installments. If an installment is partially paid, what is the `overdue_amount` for penalty calculation? Is it the full scheduled amount or the remaining unpaid portion?

**COO**:
- Waiver process (partial by Manager, full by Admin) is practical for small koperasi where Manager and Admin may be the same person (Ketua Koperasi).
- Daily scheduled job for penalty calculation requires reliable infrastructure. For small koperasi with limited IT, this is a dependency risk. Should have manual trigger fallback.
- Ta'widh requiring proof and Manager approval is operationally manageable.

**CMO**:
- Grace period of 3 days is generous enough to accommodate weekend/holiday delays. Good for member satisfaction.
- Penalty cap display to nasabah (transparency) is excellent.
- The distinction between "mampu tapi lalai" vs "tidak mampu" for ta'zir enforcement needs clear operational guidelines -- who determines inability to pay, and based on what criteria?

**Issues Found**:

1. **[HIGH]** `calculation_basis: "outstanding_principal"` for late payment is disproportionate and potentially predatory. A late installment of Rp 1 juta should not trigger denda calculated from Rp 50 juta outstanding. **Fix**: Either remove `outstanding_principal` as an option, or add a strong warning/restriction that this should only be used with very low percentage rates (e.g., 0.01% daily max when using outstanding_principal).

2. **[HIGH]** Interaction with restructuring (K007) not defined. When a loan is restructured, are existing accrued penalties waived, frozen, or carried forward? **Fix**: Add explicit section on penalty handling during restructuring. Standard practice: outstanding penalties are either waived as part of restructuring agreement or frozen and added to the restructured loan balance.

3. **[MEDIUM]** Penalty calculation for PARTIAL installments not specified. **Fix**: Clarify that overdue_amount for penalty = (total_amount - total_paid) for the specific installment, i.e., only the unpaid portion.

4. **[MEDIUM]** No manual trigger fallback for daily penalty calculation job. If the scheduled job fails over a weekend, Monday's calculation should catch up (back-calculate). **Fix**: Ensure the job is idempotent and calculates from period_start to current date, not just "today's increment."

5. **[LOW]** Criteria for determining "nasabah tidak mampu" (to waive ta'zir in Islamic mode) are not defined. **Fix**: Add guidelines (e.g., based on declared income, employment status change, medical documentation) or reference to a tenant-configurable policy.

**Missing Items**:
- Penalty handling during restructuring and write-off
- Penalty calculation for partial payment installments
- Notification to nasabah when denda is accrued (daily/weekly summary)
- Denda on non-loan obligations (e.g., simpanan wajib arrears) -- or explicit statement that K009 only covers loan/financing penalties

---

## K010: Jaminan / Agunan (Collateral Management)

**Verdict**: APPROVED WITH ADJUSTMENTS

**CEO**: Comprehensive collateral management with school-context awareness (ijazah as moral commitment collateral, personal guarantee for small loans). The multi-type support and lifecycle management are enterprise-grade features that differentiate from simple spreadsheet-based koperasi operations.

**CFO**:
- LTV ratios are appropriate for school koperasi: 100% for internal balance, 70-80% for surat berharga, 50-70% for barang bergerak. These are conservative and industry-standard.
- Coverage ratio tracking (minimum 100%, warning at 120%) is correct risk management practice.
- Foreclosure surplus must be returned to nasabah -- correctly stated and legally required.
- Deposito as collateral with enforced auto-rollover is smart -- prevents maturity gap.
- Ijazah as collateral for emergency loans (nominal value Rp 1,000,000, moral commitment) is realistic for school context.
- Concern: The `acceptance_rate` (haircut) is stored at the jaminan level, but LTV is configured at the product level. The relationship between these two needs clarification. Currently: `max_loan = appraised_value x acceptance_rate x ltv_ratio`. Is this a double-haircut? Or does LTV replace acceptance_rate? This could result in overly conservative lending: a BPKB worth Rp 15M with 70% acceptance = Rp 10.5M, then 80% LTV = Rp 8.4M. That is a 44% total haircut which may be too aggressive for koperasi.
- Missing: insurance requirement for collateral (e.g., kendaraan yang dijaminkan wajib diasuransikan). Without insurance, if the vehicle is totaled, both the loan and collateral are lost.

**CTO**:
- Data model is well-normalized: separate tables for jaminan, jaminan_pinjaman (binding), jaminan_dokumen, jaminan_valuasi. This supports the many-to-many binding and valuation history requirements.
- The Vernon _data for jaminan includes `rekening.balance` for internal collateral. This is financial data in _data which violates the security principle stated in K007 ("tidak ada data finansial di _data"). However, for collateral monitoring, the current balance IS needed for display. **Fix**: Either accept this as an exception with documented justification, or remove balance from _data and fetch it on-demand.
- `pinjaman_list` in jaminan _data is an array, which means SyncEngine updates for PinjamanUpdatedEvent need to handle array element updates (find matching element by ID, update in-place). This is more complex than simple field replacement. Ensure the SyncEngine implementation handles this pattern.
- Missing: no `jaminan_application` or registration workflow. The ADR goes straight to REGISTERED status. How is the initial data collected? Is there an application form like K001/K002/K007? For consistency, there should be a lightweight registration process.

**COO**:
- Physical collateral release requires nasabah to visit the office (verify identity, sign berita acara, receive documents). This is standard and expected in Indonesian koperasi.
- Revaluation schedule (12 months for surat berharga, 6 months for barang bergerak) is operationally manageable. The automatic reminder/alert system reduces missed revaluations.
- Foreclosure process is appropriately heavy (Manager propose, Admin approve). For school koperasi, foreclosing a guru's BPKB is socially sensitive -- the process should be last resort. This is correctly positioned after write-off.
- Ujrah (storage fee) for BMT mode adds operational complexity of billing and tracking. For most school koperasi storing a few BPKB in a safe, this may be overkill. Recommend making ujrah disabled by default, only enabled for koperasi that actually have storage costs.

**CMO**:
- Nasabah visibility into their pledged collateral status is not mentioned. They should be able to see what is pledged and when it will be released.
- The release process after loan completion should be proactively communicated to nasabah, not waiting for them to ask. A notification saying "pinjaman lunas, silakan ambil dokumen jaminan" would improve experience significantly.

**Issues Found**:

1. **[HIGH]** Double-haircut ambiguity: acceptance_rate at jaminan level AND LTV at product level creates confusion. Is the effective coverage = appraised_value x acceptance_rate x ltv_ratio? Or does LTV already incorporate the haircut? **Fix**: Clarify the formula explicitly. Recommendation: `collateral_value = appraised_value x acceptance_rate` (this is the "recognized value"). Then loan eligibility check: `loan_amount <= SUM(collateral_value)` without additional LTV multiplication. LTV should be used to set the acceptance_rate, not as a separate multiplier.

2. **[MEDIUM]** Financial data (`rekening.balance`) in Vernon _data for internal collateral contradicts the no-financial-data-in-_data principle from K007. **Fix**: Document this as a justified exception for internal collateral monitoring, or remove and fetch on-demand.

3. **[MEDIUM]** Array handling in _data (`pinjaman_list`) adds SyncEngine complexity. No other ADR uses arrays in _data. **Fix**: Consider using `jaminan_pinjaman` table with its own _data instead of embedding pinjaman_list in jaminan._data. The binding table already has _data with pinjaman details.

4. **[MEDIUM]** No collateral insurance requirement for vehicle/property collateral. **Fix**: Add optional `insurance_required` flag per collateral type configuration. For BARANG_BERGERAK (vehicles), recommend insurance as mandatory when collateral_value > configurable threshold.

5. **[LOW]** No registration workflow/form for collateral input. Unlike K001/K002/K007, collateral goes directly to REGISTERED. **Fix**: For physical collateral, add a lightweight validation step (Supervisor verify documents before status = REGISTERED). Current ADR has teller input going straight to REGISTERED without verification.

**Missing Items**:
- Collateral insurance requirement and tracking
- Nasabah notification when collateral is released
- Collateral substitution process (swap one collateral for another on active loan)
- Depreciation schedule automation for barang bergerak (currently only `depreciation_rate` in config, but no automatic value reduction in scheduled jobs)

---

## Cross-Cutting Concerns (K007-K010)

### 1. Consistency of State Machine Patterns

K007 has RESTRUCTURED as both a status and a boolean flag. K008 uses VOIDED but does not include it in the enum. K009 has penalty status (accruing, settled, waived, partial_waived). K010 has a clean state machine (registered, pledged, released, foreclosed). These need harmonization. Each entity's state machine should be complete and self-consistent with no phantom states referenced but not defined.

### 2. Scheduled Job Dependency

K007 (NPL classification daily), K008 (DPD tracking daily, reconciliation monthly), K009 (penalty calculation daily), and K010 (revaluation reminders periodic) all depend on scheduled jobs. This is 4+ scheduled jobs that must run reliably. For small school koperasi with minimal IT infrastructure:
- All jobs should have idempotent retry logic
- A single dashboard to monitor job health
- Manual trigger capability for each job
- Clear documentation on what happens if a job misses a day

### 3. Concurrent Access / Race Conditions

Multiple scheduled jobs and real-time operations can conflict:
- Auto-debit (K008) + manual teller payment on same installment
- NPL classification job (K007) + payment that changes DPD
- Penalty calculation (K009) + payment that clears overdue

All payment-related operations MUST use database-level row locking (SELECT FOR UPDATE) on the angsuran and pinjaman records. This is not consistently specified across the ADRs.

### 4. Restructuring Impact Cascade

When a loan is restructured (K007), the impact cascades across:
- K008: installment schedule voided and regenerated
- K009: existing penalties -- unclear (MISSING)
- K010: collateral binding -- should remain, but collateral_value vs new loan terms may need re-evaluation

This cascade should be documented as a unified restructuring checklist.

### 5. Write-Off Impact Cascade

When a loan is written off (K007):
- K008: all remaining installments -- should be voided/written_off
- K009: all outstanding penalties -- should be written off as part of PPAP
- K010: collateral -- released or foreclosed

The write-off ADR section covers K010 (release jaminan) but does not explicitly mention K008 and K009 disposition.

### 6. BMPK Enforcement Gap

The most critical gap across K007-K010 is the absence of BMPK (Batas Maksimum Pemberian Kredit). The current system checks DTI (borrower's ability to pay) but not the koperasi's exposure limit (how much total the koperasi can lend to one borrower/group relative to its own capital). This is a regulatory requirement and prudential risk management essential.

### 7. Islamic Compliance Audit Trail

Ta'zir routing to social fund (K009), ibra calculation (K007/K008), Rahn/ujrah (K010), and akad enforcement (K003/K007) all need to be auditable by DPS (Dewan Pengawas Syariah). A unified Islamic compliance audit view/report is needed across these domains.

---

## Recommended Adjustments

### Priority 1 (Must Fix Before Implementation)

| # | ADR | Issue | Fix |
|---|-----|-------|-----|
| 1 | K007 | BMPK missing | Add BMPK check (20% modal sendiri default) to credit pre-check and approval validation |
| 2 | K007 | RESTRUCTURED status ambiguity | Remove from enum, use boolean `is_restructured` only |
| 3 | K008 | VOIDED missing from enum | Add `voided` to `payment_status` ENUM |
| 4 | K008 | Race condition on concurrent payments | Mandate SELECT FOR UPDATE on angsuran during payment, document in ADR |
| 5 | K009 | Restructuring impact on penalties undefined | Add section on penalty disposition during restructuring |
| 6 | K010 | Double-haircut formula ambiguity | Clarify formula: collateral_value = appraised x acceptance_rate, no additional LTV multiplication |

### Priority 2 (Should Fix Before Go-Live)

| # | ADR | Issue | Fix |
|---|-----|-------|-----|
| 7 | K007 | PPAP rates not specified | Add note with standard rates (0.5%, 10%, 50%, 75%, 100%) |
| 8 | K008 | Underpayment carry-forward breaks immutability | Keep unpaid portion on current installment, do not mutate next installment |
| 9 | K009 | outstanding_principal as penalty basis is predatory | Remove or restrict with mandatory low-rate cap |
| 10 | K009 | Idempotent penalty job not specified | Add idempotency requirement for scheduled jobs |
| 11 | K010 | Balance in Vernon _data | Document exception or remove |
| 12 | K010 | pinjaman_list array in _data | Move to jaminan_pinjaman._data |

### Priority 3 (Enhancement, Post-MVP Acceptable)

| # | ADR | Issue | Fix |
|---|-----|-------|-----|
| 13 | K007 | Group borrower exposure tracking | Add optional group BMPK |
| 14 | K007 | Disbursement reversal process | Add error correction workflow |
| 15 | K008 | Holiday calendar table | Add `tenant_holiday` data model |
| 16 | K008 | Payment reversal/refund | Add correction workflow |
| 17 | K009 | "Tidak mampu" criteria for ta'zir waiver | Add tenant-configurable guidelines |
| 18 | K010 | Collateral insurance tracking | Add optional insurance flag per type |
| 19 | K010 | Nasabah notification on release | Add to notification framework |
| 20 | All | Islamic compliance audit dashboard | Unified DPS audit report across K007-K010 |

---

## Overall Assessment

**Domain Quality**: 8/10 -- The four ADRs form a cohesive and comprehensive lending system with proper dual-mode support. The level of detail in calculation formulas, data models, and edge cases is impressive for an ADR.

**Risk Assessment**: The 6 Priority-1 issues must be addressed before any implementation begins. BMPK is a regulatory requirement that cannot be skipped. The state machine inconsistencies will cause bugs. The race condition risks will cause financial discrepancies.

**Recommendation**: Address Priority 1 items, then proceed with implementation. Priority 2 items should be resolved during development (before go-live). Priority 3 items can be deferred to post-MVP.

---

*Review conducted by C-Suite Advisory Panel*
*Reviewed by: CEO, CFO, CTO, COO, CMO*
*Date: 2026-04-15*
