# C-Suite Advisory Review: ADR K015-K018

**Review Date:** 2026-04-15
**Reviewers:** C-Suite Advisory Panel (CEO, CFO, CTO, COO, CMO)
**Scope:** Accounting, SHU, Regulatory Reporting, Zakat & Infaq

---

## K015: Jurnal & Akuntansi (COA)

**Verdict: APPROVED WITH ADJUSTMENTS**

### CEO Perspective
Strong foundation for regulatory compliance. The dual-mode approach (general/islamic) positions the product well for both conventional cooperatives and BMTs. The COA structure following PSAK/SAK ETAP and PSAK 101-110 is a regulatory necessity, not just a feature. **Strategic risk**: if COA structure is wrong at launch, every tenant's financial data is compromised and migration is extremely painful.

### CFO Perspective

**COA Structure Review:**
- 1xxx-5xxx structure follows standard Indonesian cooperative accounting -- CORRECT
- 6xxx (Zakat), 7xxx (Kebajikan), 8xxx (Ta'zir) separation for Islamic mode -- CORRECT per PSAK 109
- Simpanan Pokok (3100) and Simpanan Wajib (3200) correctly placed in Ekuitas, not Kewajiban -- CORRECT per UU Koperasi (these are member equity, not deposits)
- Contra accounts (1309 Cadangan Kerugian, 1509 Akumulasi Penyusutan) present -- CORRECT

**Auto-Journal Mapping Review:**
- Setoran tunai tabungan: Dr 1101 / Cr 2101 -- CORRECT
- Penarikan tunai tabungan: Dr 2101 / Cr 1101 -- CORRECT
- Setoran simpanan pokok: Dr 1101 / Cr 3100 -- CORRECT
- Pencairan pinjaman tunai: Dr 1301 / Cr 1101 -- CORRECT
- Angsuran pokok: Dr 1101 / Cr 1301 -- CORRECT
- Angsuran bunga/margin: Dr 1101 / Cr 4100 -- CORRECT
- Biaya admin bulanan: Dr 2101 / Cr 4200 -- CORRECT
- Bagi hasil tabungan: Dr 5100 / Cr 2101 -- CORRECT
- Denda (general): Dr 1101 / Cr 4300 -- CORRECT
- Denda (islamic/ta'zir): Dr 1101 / Cr 8101 -- CORRECT (ta'zir goes to social fund, NOT income)
- Penerimaan zakat: Dr 1101 / Cr 6101 -- CORRECT

**Off-balance sheet treatment (Section 10):**
- 6xxx-8xxx NOT entering P&L -- CORRECT per PSAK 109
- Dana zakat/kebajikan/ta'zir reported as separate statements -- CORRECT
- Year-end closing carries forward 6xxx-8xxx (not closed to SHU) -- CORRECT

### CTO Perspective

**Double-entry enforcement at DB level** -- CONFIRMED via:
1. `CHECK (total_debit = total_credit)` on jurnal header
2. `CHECK (debit >= 0 AND credit >= 0)` on jurnal_line
3. `CHECK (NOT (debit > 0 AND credit > 0))` per line -- each line is either debit OR credit
4. Scheduled reconciliation job as tertiary defense

**Period management** -- Sound design with OPEN -> CLOSED -> LOCKED lifecycle. Sequential closing requirement (can't close March before February) prevents gaps.

**Year-end closing** -- Automatable: clear pre-checks (all monthly periods CLOSED, no UNPOSTED journals), deterministic closing journal generation (zero out 4xxx/5xxx, net to 3500).

**Report generation feasibility** -- Trial balance as view/query (not materialized table) is correct for cooperative volumes. Financial statements derivable from posted journal data with standard SQL aggregation.

### COO Perspective
Operational workflow is sound. Auto-post for operational journals reduces daily workload. Clear separation between auto-journal (from transactions) and manual journals (for adjustments). Year-end closing pre-checks prevent premature execution.

### CMO Perspective
Transparent accounting builds member trust. The ability for members to trace their transactions to journal entries (via source_id linkage) supports accountability messaging.

### Issues Found

| # | Severity | Issue | Recommendation |
|---|----------|-------|----------------|
| 1 | **MEDIUM** | COA missing account 2104 or similar for "Simpanan Deposito Jatuh Tempo" -- when deposito matures but member hasn't withdrawn, the liability classification may change from time deposit to demand deposit. | Add COA 2104 "Deposito Jatuh Tempo" for matured but unclaimed deposits. |
| 2 | **MEDIUM** | No COA for inter-branch accounts visible in the template, but Section on multi-branch consolidation (K017) references 1901/2901. | Explicitly add 1901 "Piutang Antar Cabang" and 2901 "Hutang Antar Cabang" to the COA template. |
| 3 | **LOW** | Journal mapping table lacks `description_template` field -- auto-generated journal descriptions should be standardized (e.g., "Setoran tunai tabungan - {nasabah_name} - {account_number}"). | Add `description_template` VARCHAR to journal_mapping with placeholders. |
| 4 | **LOW** | No mention of opening balance journal mechanism for new tenants migrating from existing (manual) systems. | Add a section on "Opening Balance Entry" -- a special journal type for initial migration balances. |
| 5 | **MEDIUM** | NUMERIC(15,2) caps at 9,999,999,999,999.99 (roughly IDR 10 trillion). For consolidated reporting of larger cooperatives or holding-level aggregation, this may be tight. | Consider NUMERIC(18,2) for future-proofing, especially for aggregate/consolidated amounts. |

### Missing Items
- Opening balance mechanism for migrating tenants
- Inter-branch COA accounts (1901/2901) in the template
- Explicit handling of multi-currency (even if not needed now, a note on IDR-only assumption)
- Archival/purge strategy for journal data (10+ years retention)

---

## K016: SHU (Sisa Hasil Usaha)

**Verdict: APPROVED WITH ADJUSTMENTS**

### CEO Perspective
SHU is the single most politically sensitive operation in any koperasi. The configurable formula approach (per AD/ART) is strategically correct -- every koperasi has slightly different rules. The RAT simulation feature is a strong differentiator that helps pengurus make data-driven decisions.

### CFO Perspective

**SHU Formula Verification:**
- SHU Bruto = Total Pendapatan (4xxx) - Total Beban (5xxx) -- CORRECT
- Cadangan >= 25% enforced -- CORRECT per UU Koperasi No. 25/1992 Pasal 45
- Total distribution percentages = 100% enforced -- CORRECT
- Jasa Modal split + Jasa Usaha split = 100% enforced -- CORRECT

**Jasa Modal Calculation:**
- Uses **daily weighted average** (SUM saldo_harian / jumlah_hari) -- CORRECT, this is the most accurate method
- Includes Simpanan Pokok + Simpanan Wajib + Tabungan Sukarela -- CORRECT
- Deposito inclusion configurable -- ACCEPTABLE
- Pro-rata for mid-year join/exit based on active days -- CORRECT

**Jasa Usaha Calculation:**
- Based on absolute transaction volume -- CORRECT per cooperative practice
- Includes deposits, withdrawals, installments, disbursements -- CORRECT
- Configurable transaction types -- ACCEPTABLE

**Zakat Institusi deducted before SHU distribution (Islamic mode):**
- `shu_neto = shu_bruto - pajak - zakat_institusi` -- CORRECT, zakat comes off the top before member distribution

### CTO Perspective
Data model is sound. The `distribution_config` as JSONB snapshot in `shu_periode` is correct -- preserves the exact percentages used for that year even if config changes later. The denormalized `amount_*` fields allow quick reporting without recalculation.

Daily weighted average calculation will be computation-heavy for large cooperatives. The mitigation mentions pre-aggregated daily snapshots from K015 reconciliation, which is the right approach.

### COO Perspective
The four-stage approval flow (CALCULATED -> REVIEWED -> APPROVED -> DISTRIBUTED) maps well to real cooperative workflows: system calculates, pengurus reviews, RAT approves, admin executes. RAT override of percentages for a specific year (without changing default config) is operationally critical -- RAT decisions should not affect future years' defaults.

### CMO Perspective
SHU transparency is key to member retention. The individual SHU statement (slip per anggota showing avg_simpanan, total_transaksi, jasa_modal, jasa_usaha) is excellent for trust-building. The simulation feature helps pengurus present clear options at RAT, reducing conflict.

### Issues Found

| # | Severity | Issue | Recommendation |
|---|----------|-------|----------------|
| 1 | **HIGH** | Rounding handling is mentioned in Mitigasi but not formalized in the data model or calculation spec. When distributing proportional SHU to hundreds of members, the sum of individual amounts will NOT equal the pool due to rounding. The ADR says "selisih dialokasikan ke cadangan" but this needs to be explicit in the calculation algorithm. | Add explicit rounding rules: (a) round each member's SHU to nearest IDR 1 (floor), (b) calculate remainder = pool - SUM(individual), (c) add remainder to cadangan, (d) store `rounding_difference` in `shu_periode`. |
| 2 | **MEDIUM** | No `shu_periode.rounding_difference` field in data model to track the rounding residual. | Add `rounding_difference NUMERIC(15,2) DEFAULT 0` to `shu_periode`. |
| 3 | **MEDIUM** | The ADR does not specify how to handle **deceased members** (anggota meninggal) during the year. Their SHU entitlement needs to go to heirs (ahli waris from K001). | Add rules for deceased members: SHU calculated up to death date, distributed to heir's account or held as payable. Cross-reference K001 ahli waris data. |
| 4 | **MEDIUM** | `shu_anggota.distribution_method` has value `pending` but this is not a method -- it is a status. | Rename to `not_distributed` or make `distribution_method` nullable (null = not yet distributed) and track distribution status separately. |
| 5 | **LOW** | No explicit mention of **pajak penghasilan** treatment on SHU received by members. While this is typically the member's responsibility, some cooperatives withhold PPh. | Add a note/configurable field: `withhold_pph BOOLEAN DEFAULT false` and `pph_rate NUMERIC(5,2)` for cooperatives that withhold member income tax on SHU. |
| 6 | **LOW** | Missing index recommendations for `shu_anggota` table -- queries will be heavy during distribution (all members for a period). | Add explicit indexes: `(tenant_id, shu_periode_id)` and `(tenant_id, nasabah_id)`. |

### Missing Items
- Rounding algorithm specification
- Deceased member SHU handling
- Tax withholding option
- Performance benchmarks/estimates for daily average calculation at scale

---

## K017: Laporan Regulasi (Regulatory Reporting)

**Verdict: APPROVED WITH ADJUSTMENTS**

### CEO Perspective
Regulatory reporting compliance is non-negotiable. The multi-regulator approach (Dinas Koperasi + OJK + PSAK Syariah) covers the Indonesian regulatory landscape comprehensively. The scheduling and reminder system reduces regulatory risk from missed deadlines. The `is_lkm` flag for OJK reporting is smart -- not all cooperatives are registered as LKM.

### CFO Perspective

**NPL Ratio:**
- Formula: Total NPL (kol 3+4+5) / Total Outstanding Loans -- CORRECT per OJK definition
- Kolektibilitas 3 (Kurang Lancar) + 4 (Diragukan) + 5 (Macet) = NPL -- CORRECT
- Threshold > 5% -- CORRECT (OJK standard for microfinance)

**CAR Calculation:**
- Formula: Modal (3xxx) / ATMR -- CORRECT
- ATMR weights: Kas & SBI 0%, Antar bank 20%, Pinjaman anggota 100%, Aktiva tetap 100% -- CORRECT per simplified OJK weights for LKM/cooperative
- Threshold < 8% -- CORRECT

**BMPK:**
- Max Single Borrower Exposure / Modal, threshold > 20% -- CORRECT per OJK

**Other Ratios:**
- ROA = Net Income / Total Assets -- CORRECT
- ROE = Net Income / Equity -- CORRECT
- LDR/FDR = Total Financing / Total Deposits -- CORRECT
- FDR threshold at 95% for Islamic (vs 90% for general) -- ACCEPTABLE, reflects Islamic finance practice

**Dual-mode terminology:**
- NPL vs NPF (Islamic) -- CORRECT
- LDR vs FDR -- CORRECT
- "Neraca" vs "Laporan Posisi Keuangan" -- CORRECT per PSAK 101

**Islamic additional reports:**
- Laporan Sumber & Penyaluran Dana Zakat -- CORRECT per PSAK 109
- Laporan Dana Kebajikan -- CORRECT
- Laporan Perubahan Dana Investasi Terikat -- CORRECT per PSAK 105
- Laporan Rekonsiliasi Pendapatan Bagi Hasil -- CORRECT per PSAK 105

### CTO Perspective
The report engine architecture (template-based, snapshot approach) is sound. Storing report data as JSONB snapshots ensures point-in-time consistency -- critical for financial reports. The versioning system (draft regeneration creates new version) is correct.

Multi-branch consolidation with inter-branch elimination using 1901/2901 accounts is technically feasible given K015's journal structure.

Scheduled auto-generation via background jobs with retry mechanism is the right approach.

### COO Perspective
The scheduling with reminders (30d, 7d, 3d, 1d before deadline) is operationally excellent. The OVERDUE status auto-trigger ensures visibility. The holiday calendar awareness for deadline adjustment is a practical detail that many systems miss.

The four-stage approval (DRAFT -> REVIEWED -> FINAL -> SUBMITTED) with immutable FINAL status and revision mechanism is the correct audit approach.

### CMO Perspective
Regulatory compliance reporting builds institutional credibility. The ability to generate professional PDFs with tenant branding (logo, signatory) elevates the cooperative's professional image when submitting to regulators.

### Issues Found

| # | Severity | Issue | Recommendation |
|---|----------|-------|----------------|
| 1 | **HIGH** | CAR formula uses "Modal (3xxx)" but this is too simplistic. CAR should use **Tier 1 + Tier 2 capital**, which for cooperatives typically means: Simpanan Pokok + Simpanan Wajib + Cadangan Umum + Cadangan Risiko + SHU (retained). NOT all of 3xxx (e.g., SHU Tahun Berjalan that hasn't been through RAT should not count as capital for CAR). | Specify CAR numerator explicitly: Tier 1 = (3100 + 3200 + 3300 + 3400 + 3600), exclude 3500 (SHU Tahun Berjalan) until it has been approved by RAT and partially allocated to cadangan. |
| 2 | **HIGH** | ATMR calculation is missing weights for some asset categories in K015 COA: Piutang Bunga/Margin (1302), Penyertaan/Investasi (1400), Aktiva Lain-lain (1600). These all carry risk and need ATMR weights. | Add ATMR weights: 1302 (Piutang Bunga) = 100%, 1400 (Investasi) = 100%, 1600 (Aktiva Lain) = 100%. Make ATMR weight configurable per COA account for future flexibility. |
| 3 | **MEDIUM** | NPL ratio denominator says "Total Outstanding Loans" but should be clarified: is this gross outstanding (before CKPN/provisioning) or net? OJK uses **gross** outstanding as denominator. | Explicitly state: NPL Ratio = Total NPL (Kol 3+4+5 gross) / Total Outstanding Pembiayaan (gross, before CKPN). |
| 4 | **MEDIUM** | Laporan Perubahan Ekuitas is listed under Dinas Koperasi reports but has no formula/structure specified (unlike Neraca and Laba Rugi which have clear layouts in K015). | Add Laporan Perubahan Ekuitas structure showing: Saldo Awal, +/- SHU Tahun Berjalan, +/- Distribusi SHU (cadangan, dana-dana), Saldo Akhir per equity account. |
| 5 | **MEDIUM** | The `generated_by` field says "(user atau 'SYSTEM' untuk auto-generate)" -- storing string 'SYSTEM' in a UUID field is a type violation. | Use a well-known UUID constant for system-generated actions (e.g., UUID v7 of all zeros or a dedicated system user record), or make `generated_by` VARCHAR. |
| 6 | **LOW** | No mention of **Laporan Pengawas** (Supervisory Board report) which is a mandatory component of RAT documentation for Dinas Koperasi. | Add Laporan Pengawas as a report type under Dinas Koperasi category, even if it is partially manual (text-based assessment by Pengawas). |
| 7 | **LOW** | Liquidity Ratio formula: "Liquid Assets (1100+1200) / Short-term Liabilities (2100+2200)" -- should clarify that 1100 includes all sub-accounts (1101+1102+1103) and similarly for others. | Clarify that ratio calculations aggregate all sub-accounts under the specified parent code. |

### Missing Items
- Explicit ATMR weight table per COA account
- Laporan Perubahan Ekuitas structure
- Laporan Pengawas template
- Clarification on whether reports use accrual or cash basis (should be accrual for formal reports, but daily cash report is cash basis)
- XBRL or e-reporting format consideration for future OJK digital submission

---

## K018: Zakat & Infaq

**Verdict: APPROVED WITH ADJUSTMENTS**

### CEO Perspective
For a BMT product, this module is a core differentiator, not a nice-to-have. The Baitul Maal function is what distinguishes a BMT from a conventional cooperative. The school context features (bulk fitrah collection per class, beasiswa distribution) are strategically aligned with the sekolah target market.

Feature activation via `coop_type` derivation (not separate config) is the correct approach -- prevents misconfiguration.

### CFO Perspective

**Zakat Mal Calculation:**
- Nisab: 85 gram emas -- CORRECT per consensus of Islamic scholars
- Haul: 1 hijriah year (~355 days) -- CORRECT
- Rate: 2.5% -- CORRECT
- Basis: (total harta qualifying - hutang) x 2.5% -- CORRECT
- Haul resets if assets drop below nisab -- CORRECT per fiqh majority opinion

**Zakat Fitrah:**
- Fixed amount per head, configurable -- CORRECT
- Period: 1 Ramadan to 1 Syawal -- CORRECT (before Eid prayer is sunnah timing)

**Zakat Institusi:**
- 2.5% of net profit after tax -- CORRECT
- Deducted before SHU distribution (per K016) -- CORRECT
- Journal: Dr 3500 SHU / Cr 6103 Zakat Institusi -- CORRECT (reduces equity, increases zakat fund)

**Off-balance sheet treatment:**
- All zakat/infaq/ta'zir journals use 6xxx/8xxx -- CORRECT
- Not entering P&L -- CORRECT per PSAK 109
- Separate reporting statements -- CORRECT

**Ta'zir integration:**
- Ta'zir (denda) flows to COA 8xxx (NOT 4xxx income) -- CORRECT (Islamic compliance verified)
- Managed alongside infaq (not zakat) for distribution -- CORRECT (ta'zir is not zakat)
- Separate tracking in reports -- CORRECT for DPS audit

**Journal entries verification:**
- Zakat Mal: Dr 1101 / Cr 6101 -- CORRECT
- Zakat Fitrah: Dr 1101 / Cr 6102 -- CORRECT
- Zakat Institusi: Dr 3500 / Cr 6103 -- CORRECT
- Penyaluran Zakat: Dr 6201 / Cr 1101 -- CORRECT
- Infaq: Dr 1101 / Cr 6300 -- CORRECT
- Ta'zir: Dr 1101 / Cr 8101 -- CORRECT

### CTO Perspective
The data model is well-structured with clear separation between collection, distribution, mustahik, and ta'zir fund entities. The batch collection mechanism for zakat fitrah (school context) is a practical necessity.

Hijriah calendar tracking is a known technical challenge. The ADR correctly identifies the need for a Masehi-Hijriah conversion library. The daily nisab checking (for haul tracking) needs to be a lightweight background job, not a heavy computation.

### COO Perspective
The multi-level approval for distribution (amil proposes, manager reviews, admin approves above threshold) provides proper separation of duties. The proof-of-receipt requirement for every distribution ensures accountability.

The school-context features (per-class collection, wali kelas input, batch approval) map well to how pesantren actually operate during Ramadan.

### CMO Perspective
Zakat transparency is religiously significant for BMT members. The per-muzakki receipt, per-asnaf distribution reports, and annual RAT summary demonstrate responsible stewardship of religious funds. This builds deep trust with the Muslim community.

The education transparency feature (aggregate per class, not individual data) balances transparency with privacy -- good design.

### Issues Found

| # | Severity | Issue | Recommendation |
|---|----------|-------|----------------|
| 1 | **HIGH** | Zakat Mal calculation says "total harta qualifying - hutang" but does not specify WHICH hutang. In zakat fiqh, only hutang that is **due immediately** (hutang jatuh tempo) should be deducted, not total outstanding loans. A member with a 5-year loan should not deduct the entire principal from zakat calculation -- only the installments due within the haul period. | Clarify: deductible debts = outstanding installments due within the current haul period (e.g., 12 months of installments), NOT total remaining loan principal. Add `debts_calculation_method` config: `due_installments_only` (recommended) vs `total_outstanding` (alternative opinion). |
| 2 | **HIGH** | PSAK 109 requires the zakat fund balance to appear on the **Neraca (Balance Sheet)** as a separate liability line item ("Dana Zakat" and "Dana Infaq/Shadaqah"), not purely off-balance sheet. The ADR says "off-balance sheet notes" which is ambiguous. Per PSAK 109 paragraph 35-36, the balances of zakat and infaq/shadaqah funds are presented as part of the **entity's statement of financial position** under a separate section (neither liability nor equity, but a distinct "dana" section). | Clarify balance sheet presentation: Add a distinct section in Neraca for "DANA ZAKAT" and "DANA INFAQ/SHADAQAH" and "DANA TA'ZIR" -- separate from Kewajiban and Ekuitas, per PSAK 109. Update K015 Neraca structure to include this third section. |
| 3 | **MEDIUM** | Amil's share (asnaf #3) is typically capped at 12.5% (1/8) of total zakat collected per fiqh majority. The ADR does not enforce this cap. | Add validation: Amil distribution should not exceed 12.5% of total zakat collected for the period. Make this a configurable threshold with default 12.5%. |
| 4 | **MEDIUM** | No `nisab_history` table in the data model for tracking historical nisab values. The ADR mentions "riwayat perubahan nisab" but there is no table to store this. The current `nisab_config` only stores the latest value. | Add `nisab_history` table: (id, tenant_id, effective_date, gold_gram, nisab_amount_idr, source_reference, created_by, created_at). The zakat calculation should look up the nisab value effective at the haul completion date. |
| 5 | **MEDIUM** | Zakat Institusi journal (Dr 3500 SHU / Cr 6103) reduces SHU directly. However, this should only happen AFTER year-end closing when SHU is known. The timing relationship with K016 (SHU calculation) needs explicit sequencing: Year-end close -> Calculate SHU -> Calculate Zakat Institusi -> Deduct from SHU -> Distribute remaining SHU. | Add explicit sequencing rule: Zakat Institusi calculation happens AFTER year-end closing but BEFORE SHU distribution. The journal should be posted as part of the pre-RAT process, and `shu_neto` in K016 should be net of this zakat. |
| 6 | **LOW** | `zakat_collection` table has `journal_entry_id` as FK but the column name references `journal_entry` while K015 calls the table `jurnal`. | Align FK naming: use `jurnal_id` (matching K015 table name) instead of `journal_entry_id` to maintain naming consistency. |
| 7 | **LOW** | No explicit handling of **zakat refund** scenarios (e.g., member pays zakat but then requests cancellation before distribution, or overpayment). | Add `cancelled` status handling: if zakat_collection is cancelled before fund disbursement, create reversal journal (Dr 6101 / Cr 1101) and credit back to member's account. |

### Missing Items
- Nisab history table
- Amil percentage cap enforcement
- Explicit debt deduction methodology for zakat mal
- Balance sheet presentation clarification per PSAK 109
- DPS (Dewan Pengawas Syariah) audit trail / sign-off mechanism in the approval flow
- Zakat refund/cancellation reversal mechanism

---

## Cross-Cutting Concerns

### 1. K015 <-> K016 Dependency Chain
The year-end closing (K015) -> SHU calculation (K016) -> Zakat Institusi (K018) -> SHU distribution (K016) chain is the most critical annual process. The sequencing is implied across ADRs but not explicitly documented as a unified workflow. **Recommendation:** Create an explicit "Annual Closing Sequence Diagram" that shows the exact order and dependencies across K015, K016, K018.

### 2. K015 COA <-> K017 Ratio Formulas
K017 references COA account ranges (e.g., "Modal 3xxx" for CAR, "1100+1200" for liquidity) but K015's COA structure shows that not all 3xxx accounts should count as capital (e.g., 3500 SHU Tahun Berjalan is temporary). **Recommendation:** Create a mapping document that explicitly maps each ratio formula to specific COA account codes, not ranges.

### 3. K018 Balance Sheet Presentation <-> K015 Neraca
K015 Section 9 shows a standard Neraca (Assets = Liabilities + Equity) but PSAK 109 requires a third section for "Dana Zakat/Infaq/Shadaqah" that is neither liability nor equity. The K015 Neraca structure needs to be updated for Islamic mode to include this distinct section. **Recommendation:** Update K015 Section 9 to show the Islamic mode Neraca with three sections: Aset, Kewajiban + Ekuitas, Dana Zakat + Infaq + Ta'zir.

### 4. Ta'zir Flow: K009 -> K018 -> K015
The flow is: K009 generates denda -> K015 auto-journal to 8101 -> K018 manages distribution. This is correctly specified across all three ADRs. VERIFIED: ta'zir does NOT flow to income (4xxx). **No issue.**

### 5. NPL from K007 -> K017
K017 references K007 for NPL data (kolektibilitas 1-5). Ensure K007's collectibility classification aligns with OJK's definition (lancar/dalam perhatian khusus/kurang lancar/diragukan/macet) and that the day-past-due thresholds match OJK regulations. **Recommendation:** Cross-verify with K007 ADR.

### 6. Data Consistency Across Modules
All four ADRs use Vernon _rels/_data pattern consistently. SyncEngine triggers are defined for cross-entity updates. The pattern is coherent across modules.

---

## Recommended Adjustments (Priority Order)

### Must-Fix Before Implementation (HIGH)

1. **K017 Issue #1**: Fix CAR formula to use specific Tier 1 capital accounts, not all 3xxx
2. **K017 Issue #2**: Complete ATMR weight table for all asset COA accounts
3. **K018 Issue #1**: Clarify zakat mal debt deduction method (due installments only vs total outstanding)
4. **K018 Issue #2**: Align PSAK 109 balance sheet presentation -- update K015 Neraca for Islamic mode
5. **K016 Issue #1**: Formalize rounding algorithm with explicit remainder-to-cadangan rule

### Should-Fix Before Launch (MEDIUM)

6. **K015 Issue #1**: Add COA for matured deposits (2104)
7. **K015 Issue #2**: Add inter-branch COA accounts (1901/2901) to template
8. **K016 Issue #3**: Add deceased member SHU handling rules
9. **K017 Issue #3**: Clarify NPL uses gross outstanding denominator
10. **K017 Issue #4**: Add Laporan Perubahan Ekuitas structure
11. **K018 Issue #3**: Add amil percentage cap (12.5% default)
12. **K018 Issue #4**: Add nisab_history table
13. **K018 Issue #5**: Document explicit zakat institusi timing in annual closing sequence
14. **Cross-cutting #1**: Create Annual Closing Sequence Diagram
15. **Cross-cutting #2**: Create explicit ratio-to-COA mapping document

### Nice-to-Have (LOW)

16. **K015 Issue #3**: Add description_template to journal_mapping
17. **K015 Issue #4**: Add opening balance entry mechanism
18. **K016 Issue #5**: Add tax withholding option
19. **K017 Issue #6**: Add Laporan Pengawas template
20. **K018 Issue #6**: Align FK naming (jurnal_id vs journal_entry_id)
21. **K018 Issue #7**: Add zakat refund mechanism

---

## Overall Assessment

| ADR | Verdict | Accounting Compliance | Technical Soundness | Operational Feasibility |
|-----|---------|----------------------|--------------------|-----------------------|
| K015 | Approved with Adjustments | STRONG -- COA and mappings are correct | STRONG -- DB-level enforcement | STRONG -- auto-journal reduces workload |
| K016 | Approved with Adjustments | STRONG -- formula correct per UU Koperasi | GOOD -- needs rounding spec | STRONG -- simulation feature valuable |
| K017 | Approved with Adjustments | NEEDS FIX -- CAR formula oversimplified | GOOD -- report engine sound | STRONG -- scheduling excellent |
| K018 | Approved with Adjustments | NEEDS FIX -- PSAK 109 presentation gap | GOOD -- data model complete | STRONG -- school features practical |

**Summary:** All four ADRs demonstrate deep domain knowledge of Indonesian cooperative accounting and Islamic finance. The core accounting logic (double-entry, auto-journal, COA structure, SHU formula, zakat calculation) is largely correct. The critical fixes center on: (a) CAR formula precision in K017, (b) PSAK 109 balance sheet presentation in K018/K015, and (c) zakat mal debt deduction methodology in K018. These must be addressed before implementation to avoid regulatory compliance gaps.

**Signed:**
- CEO Advisory Panel
- CFO Advisory Panel
- CTO Advisory Panel
- COO Advisory Panel
- CMO Advisory Panel

Review Date: 2026-04-15
