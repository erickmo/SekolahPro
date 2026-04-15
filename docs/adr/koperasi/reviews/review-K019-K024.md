# C-Suite Advisory Review: ADR K019-K024 (Extension Modules)

**Review Date**: 2026-04-15
**Panel**: CEO, CFO, CTO, COO, CMO
**Scope**: Extension modules beyond core koperasi operations
**Cross-Reference**: K001-K002 (core entities), K005 (tabungan), K011 (transactions)

---

## Individual ADR Reviews

---

### K019: Toko & Kantin

**Verdict**: APPROVED WITH ADJUSTMENTS

**CEO**: This is arguably the most strategically important extension. School koperasi in Indonesia historically operate toko and kantin as primary revenue sources. Excluding this would make the system feel incomplete to operators who see toko/kantin as *core* to their koperasi, not an extension. However, the scope is dangerously broad: POS + inventory + supplier management + meal plans + multi-location is essentially building a separate retail system. The risk of scope creep here is the highest across all six ADRs. **Recommendation**: Ship a minimal POS (cart, checkout, basic stock tracking) in MVP. Defer supplier management, purchase orders, meal plans, and stock opname to Phase 2.

**CFO**: Revenue impact is direct and measurable: toko/kantin profit flows into SHU (K016). The cashless payment integration (TABUNGAN_DEBIT) is a strong differentiator that ties the POS directly to the koperasi's financial core. However, the full inventory management system (supplier, PO workflow, stock opname) adds significant development cost relative to the immediate revenue it generates. Most small school koperasi manage suppliers informally. **ROI assessment**: POS + cashless = high ROI. Full inventory + supplier management = moderate ROI, better deferred.

**CTO**: This ADR is the largest single module in the entire system. The POS requirement (< 2 second response time) demands performance optimization that will consume engineering bandwidth. The data model is well-designed with proper stock movement tracking and atomic transactions, which is good. However, the meal plan auto-deduction and multi-location features add significant integration complexity. **Concern**: The `current_stock` field as a cache of SUM(movements) is a sound pattern but requires careful reconciliation logic. **Recommendation**: Treat POS as a bounded context with clear interfaces to the financial core. Consider it a separately deployable module in the future.

**COO**: Hardware dependency (barcode scanner, receipt printer, NFC reader) is a real barrier for schools with limited IT infrastructure. The ADR correctly notes hardware-agnostic fallback (manual input), but the UX gap between a barcode-scanned POS and manual input is enormous. Stock opname requires trained staff who understand the process. **Concern**: Who trains the kantin staff on POS operation? The training burden is underestimated.

**CMO**: Cashless kantin is a *killer feature* for parent adoption. "Your child doesn't need to carry cash to school" is an immediately compelling value proposition. Spending limit enforcement at POS checkout directly addresses parent anxiety about children's spending habits. The daily menu visibility (via K023 portal) adds engagement.

**Issues Found**:
1. **[HIGH]** Scope too large for single delivery. POS + inventory + supplier + meal plan + multi-location should be phased, not shipped together.
2. **[MEDIUM]** `spending_limit_config` in K019 Section 7 overlaps with `ewallet_config` in K021 Section 2. These are essentially the same domain concept (parent spending controls) defined in two places. Must consolidate into one source of truth.
3. **[MEDIUM]** No mention of offline POS capability. Kantin during lunch rush with 500 students cannot afford network failures. At minimum, document the offline strategy even if deferred.
4. **[LOW]** `margin_percentage` as a stored calculated field will drift if cost_price or sell_price is updated without recalculation trigger. Should be computed on read or have an explicit update trigger.

**Missing Items**:
- Offline POS strategy (even if Phase 2, document the approach)
- Receipt printing specification (thermal printer protocol, format)
- End-of-day POS session settlement workflow (similar to teller session K012 but for POS)
- Tax/pajak handling (PPN on goods sold, if applicable)
- Return/refund workflow (only void is documented, but partial returns are common in toko)

**MVP Recommendation**: **MVP (Minimal)** -- POS checkout + TABUNGAN_DEBIT payment + basic stock tracking (in/out only) + spending limit. Defer: supplier management, PO workflow, stock opname, meal plans, multi-location to Phase 2.

---

### K020: Payroll Integration & Auto-Deduction

**Verdict**: APPROVED

**CEO**: This is a clear efficiency play and directly relevant to koperasi operations. Simpanan wajib and angsuran collection via payroll deduction is standard practice in Indonesian school koperasi. Not having this would be a competitive disadvantage. The decision to be a *consumer* (not producer) of payroll data is strategically correct -- avoids building a payroll system while capturing the integration value.

**CFO**: Extremely high ROI. Payroll deduction ensures collection rates near 100% for simpanan wajib and angsuran, directly impacting koperasi cash flow predictability. The cost of implementation is low (CSV parsing + batch processing + approval workflow) relative to the revenue assurance it provides. The priority-based deduction logic is financially sound -- loan repayment (legal obligation) correctly prioritized over voluntary savings.

**CTO**: Clean architecture. The consumer-only approach is the right call. Data model is solid with proper audit trails. The idempotency constraint (one batch per period+branch) prevents double-deduction, which is critical. The matching strategy (NIK primary, NIP fallback) is pragmatic for Indonesian school systems. **Minor concern**: The batch execution for large koperasi (hundreds of employees) should use background jobs with progress tracking, which is correctly noted.

**COO**: The CSV upload path is essential. Most schools in Indonesia, especially those outside major cities, do not have HR systems with API capabilities. The configurable column mapping is a smart feature that accommodates the reality of every school having different spreadsheet formats. **Question**: What happens when the TU staff changes and the new person does not know the upload process? This is a training continuity risk.

**CMO**: Transparency wins here. Monthly payroll deduction notifications give nasabah confidence that the system is working correctly. The deduction report serves as proof of contribution, which matters for SHU distribution discussions.

**Issues Found**:
1. **[MEDIUM]** No mention of what happens when an employee resigns mid-month. The authorization revocation flow (Section 7) handles voluntary revocation, but involuntary separation (fired, contract ended) needs a clear trigger from HR data or manual process.
2. **[LOW]** The `raw_data_url` for storing uploaded CSV files needs a retention policy and encryption-at-rest policy explicitly stated (salary data is highly sensitive).
3. **[LOW]** The partial deduction option should default to OFF and require explicit tenant opt-in, which it does. Good.

**Missing Items**:
- Employee resignation/termination handling flow
- Data privacy impact assessment for salary data (even temporary storage of salary is PII under Indonesian data protection law)
- Rollback scenario documentation: what happens to already-credited tabungan/simpanan if batch is rolled back?

**MVP Recommendation**: **MVP** -- CSV upload + matching + priority deduction + approval workflow + execution. Defer: API push from HR system to Phase 2.

---

### K021: E-Wallet / Uang Saku Digital

**Verdict**: APPROVED WITH ADJUSTMENTS

**CEO**: The "spending interface over existing tabungan" architecture is a brilliant strategic decision. It avoids regulatory landmines (e-money license) while delivering the same user experience. This is NOT an e-wallet in the regulatory sense -- it is a controlled spending layer over a savings account. This distinction must be crystal clear in all marketing and documentation. **Risk**: If this is *marketed* as "e-wallet" to the public, regulators might take interest. Consider branding it as "Uang Saku Digital" or "Kartu Belanja Siswa" rather than "E-Wallet."

**CFO**: The regulatory risk assessment is the most critical item here. Under Bank Indonesia regulations (PBI 20/6/PBI/2018 on Electronic Money), an e-money product requires a license from BI. However, since K021 explicitly does NOT create a separate balance -- it is a debit interface to an existing tabungan -- this does NOT constitute e-money. The tabungan itself is the licensed product under the koperasi's OJK/Dinas Koperasi registration. **This is a safe architecture.** However, the ADR should explicitly document this regulatory analysis as a section. The QR/NFC card is merely an authentication mechanism, not a stored-value instrument. Payment gateway fees for top-up via transfer (K024) are the main ongoing cost: Midtrans/Xendit charges 0.7-1.5% per VA transaction + Rp 4,000-5,000 flat fee. For a typical top-up of Rp 100,000-500,000, this is 1-6% effective fee. This must be passed to the parent or absorbed by the koperasi.

**CTO**: The architecture is sound. Reusing tabungan infrastructure avoids the double-balance reconciliation problem. The card management system is well-designed with proper freeze/deactivate lifecycle. **Concerns**: (1) NFC hardware requirement is a barrier -- QR static should be the default recommendation for cost-sensitive schools. (2) The PIN management (bcrypt/argon2 for 6-digit PIN) is correct but the 30-minute auto-freeze on 3 failed attempts may be too aggressive for young students who forget their PIN. Consider a configurable lockout duration. (3) No offline mode documented -- this is the same concern as K019.

**COO**: Card issuance for hundreds of students at the start of each school year is a significant operational event. The ADR mentions "batch card issuance" in mitigations but does not detail the workflow. Schools will need a clear process: (1) generate card assignments from enrollment data, (2) print/order physical cards, (3) distribute and activate. The teacher/wali kelas proxy feature (Section 8) is a thoughtful addition for young students but adds operational complexity in managing proxy assignments.

**CMO**: This is the *differentiator* feature. "Cashless campus" is a compelling narrative for school administrators and parents. The parental control features (daily limit, category restrictions, real-time spending tracking) directly address the #1 parent concern: "What is my child spending money on?" Combined with K023's parent portal, this creates a sticky ecosystem that competitors without financial integration cannot replicate.

**Issues Found**:
1. **[CRITICAL]** The ADR title says "E-Wallet" but the architecture is explicitly NOT an e-wallet (no separate balance). The naming creates regulatory confusion. Rename to "Uang Saku Digital" or "Kartu Belanja Siswa" throughout all documentation and UI.
2. **[HIGH]** Spending limit configuration exists in THREE places: K019 Section 7 (`spending_limit_config`), K021 Section 2 (`ewallet_config`), and K023 Section 4 (`parent_child_config`). This is a serious design flaw. There must be ONE source of truth for spending limits.
3. **[MEDIUM]** No mention of regulatory compliance documentation. Add an explicit section stating: "This module does NOT constitute electronic money (uang elektronik) under PBI 20/6/PBI/2018 because no separate balance is created. The spending interface operates as a debit authorization mechanism against an existing tabungan account licensed under the koperasi's registration."
4. **[MEDIUM]** The `qr_dynamic` card type requires students to have a device (phone/tablet) to generate per-transaction QR codes. This is unrealistic for most Indonesian school students, especially at SD/SMP level. Remove or clearly mark as Phase 3.
5. **[LOW]** Card replacement fee (Rp 10,000 default) deducted from tabungan -- needs parent notification/consent before deduction.

**Missing Items**:
- Explicit regulatory compliance statement (BI e-money exclusion reasoning)
- Card batch issuance workflow for start-of-year enrollment
- Offline transaction strategy
- Card production/procurement process (NFC card vendor, QR card printing)
- Consolidation of spending limit configuration across K019/K021/K023

**MVP Recommendation**: **Phase 2** -- Depends on K019 POS being operational first. Ship K019 POS with TABUNGAN_DEBIT (which already provides cashless payment) in MVP. Add the card management, parent controls, and proxy features in Phase 2 as the "Uang Saku Digital" upgrade.

---

### K022: Notifikasi & Komunikasi

**Verdict**: APPROVED WITH ADJUSTMENTS

**CEO**: Notifications are infrastructure, not a feature. Every modern financial system needs reliable notifications. WhatsApp-first is the correct strategy for Indonesia. This is MVP-essential because transaction receipts and approval notifications are required for basic operations.

**CFO**: WhatsApp Business API costs are the main concern. Pricing varies by provider: Official API charges per conversation (Rp 300-700 per conversation depending on category), third-party providers (Wablas, Fonnte) charge Rp 50-150 per message. For a koperasi with 500 members, assuming 5 notifications per member per month = 2,500 messages/month = Rp 125,000-375,000/month with third-party, or Rp 750,000-1,750,000/month with official API. This is manageable but must be budgeted. **Recommendation**: Start with third-party provider (Wablas/Fonnte) for MVP; migrate to official API when volume justifies it.

**CTO**: The notification architecture is well-designed: queue-based, async, with retry and fallback. The template management system is solid. **Concern**: The comprehensive event list (Section 2) has ~30+ event types. For MVP, implement only the CRITICAL and HIGH-priority events that cannot be opted out. The template approval process for WhatsApp Business API (1-3 days per template) means all templates must be prepared and submitted BEFORE go-live. This is a deployment prerequisite that can delay launch if not planned.

**COO**: WhatsApp Business API approval process is non-trivial. Requirements include: verified business (Meta Business Verification), approved use case, and individual template approvals. For a school koperasi, the business verification might be challenging. The third-party provider route (Wablas/Fonnte) is operationally simpler -- they handle the Meta relationship. **Training concern**: Template management (customizing notification text) requires someone who understands the variable syntax.

**CMO**: Transaction notifications build trust. Parents receiving "Your child spent Rp 15,000 at Kantin Putra" creates engagement and confidence. The daily spending summary (opt-in) is a particularly strong retention driver for parent portal adoption.

**Issues Found**:
1. **[MEDIUM]** Too many event types for MVP. Implement only CRITICAL + non-opt-out HIGH events in Phase 1. Add MEDIUM and LOW events in Phase 2.
2. **[MEDIUM]** WhatsApp template approval process is a deployment dependency. Document as a prerequisite checklist for go-live.
3. **[LOW]** The `notifikasi_log` table will grow very fast. The 90-day rotation is mentioned but the archiving strategy (cold storage vs delete) needs to be explicit. For financial notifications, regulators may require longer retention.
4. **[LOW]** SMS fallback assumes the koperasi has an SMS gateway contract. This is an additional cost and setup step that should be documented as optional for MVP.

**Missing Items**:
- Go-live checklist for WhatsApp template preparation and approval
- Cost estimation model per tenant (messages/month x cost/message)
- SMS gateway provider options and setup process
- Notification volume projection and infrastructure sizing

**MVP Recommendation**: **MVP (Minimal)** -- IN_APP notifications for all events + WhatsApp for CRITICAL/HIGH events only (transaction receipts, approval notifications, overdue alerts). Defer: SMS fallback, batch notifications, quiet hours, full template management UI to Phase 2.

---

### K023: Dashboard & Self-Service Portal

**Verdict**: APPROVED WITH ADJUSTMENTS

**CEO**: The dashboard/portal is essential infrastructure. Without it, the system is an internal-only tool that requires staff to relay all information to members. The self-service portal reduces operational load (no more "teller, what's my balance?" visits) and the parent portal is the primary touchpoint for the parent market segment. However, six different dashboard types (Admin, Supervisor, Teller, Nasabah, Parent, Kepala Sekolah) is ambitious for MVP. **Recommendation**: MVP needs Teller workspace (operational necessity) + Nasabah portal (self-service) + Parent portal (differentiation). Admin/Supervisor dashboards can start as simpler pages. Kepala Sekolah dashboard can be Phase 2.

**CFO**: Self-service reduces operational cost: each "check my balance" visit that is handled via portal instead of teller saves approximately 5-10 minutes of staff time. At 50 such queries per day across a medium koperasi, that is 4-8 hours of staff time per day reclaimed. The parent portal drives top-up behavior (remote top-up = more deposits = more investable funds for the koperasi). ROI is strong.

**CTO**: The widget-based composable architecture is a sound design choice. System-defined widget catalog with tenant-level enable/disable gives flexibility without allowing tenants to create custom widgets (which would be a maintenance nightmare). The `portal_session` tracking for activity logging is good for analytics but must not block user experience (correctly noted as fire-and-forget async). **Concern**: The materialized view / pre-computed metrics approach for executive dashboard widgets is the right call but needs a clear background job architecture to keep metrics fresh.

**COO**: The self-service portal eliminates the #1 operational bottleneck: members coming to the office for information that could be self-served. The parent portal's remote top-up feature eliminates another bottleneck: parents queuing at the koperasi to deposit into their child's account. **Concern**: The profile update request flow (nasabah submits request, teller/supervisor approves) adds to the approval queue. If not managed well, this becomes a bottleneck.

**CMO**: The parent portal is the *killer feature for adoption*. When marketing to schools, "parents can monitor and control their child's finances from their phone" is the single most compelling pitch. Spending analytics (pie charts, trends) transform a utility into an engagement platform. Savings goal tracking gamifies the experience for children. This is the feature that sells the system.

**Issues Found**:
1. **[HIGH]** `parent_child_config` in K023 Section 8 includes spending limits that overlap with K019 and K021. This is the third place spending limits are defined. Consolidate immediately.
2. **[MEDIUM]** Six dashboard types for MVP is over-scoped. Prioritize: Teller workspace, Nasabah portal, Parent portal. Simplify Admin/Supervisor to basic pages. Defer Kepala Sekolah.
3. **[MEDIUM]** Statement PDF generation on-demand could be resource-intensive. Add rate limiting (max N statements per day per nasabah) and caching (same period = same PDF).
4. **[LOW]** The `pages_visited` and `actions_performed` JSONB arrays in `portal_session` will grow unbounded during long sessions. Consider max size or periodic flush.

**Missing Items**:
- Authentication flow for nasabah/parent portal (separate login from staff? Single app with role-based routing?)
- Password recovery / self-service password reset flow for portal users
- Session timeout configuration
- Accessibility compliance (WCAG) for public-facing portal

**MVP Recommendation**: **MVP (Minimal)** -- Teller workspace + Nasabah self-service portal (view balance, transaction history, statement download) + Parent portal (view child spending, set spending limits, remote top-up). Defer: Kepala Sekolah dashboard, widget configuration UI, bulk export, savings goal gamification to Phase 2.

---

### K024: Integrasi & API

**Verdict**: APPROVED WITH ADJUSTMENTS

**CEO**: API and integration architecture is foundational infrastructure. The decision to support multiple payment providers (Midtrans/Xendit) via strategy pattern is correct -- avoids vendor lock-in. The school management system sync is strategically important for the broader SekolahPro ecosystem. However, building ALL integrations (school sync, payment gateway, regulatory reporting, accounting export, WhatsApp, webhooks, inbound API) in Phase 1 creates too many external dependencies. Each integration is a potential failure point and maintenance burden. **Recommendation**: MVP only needs school sync (event-driven) + payment gateway (VA for top-up) + WhatsApp API (for K022). Defer: regulatory e-filing API, accounting export, webhook outbound, bank statement import.

**CFO**: Payment gateway integration has direct revenue impact (enables remote top-up, which increases deposits). Provider fees: Midtrans VA fee is Rp 4,000 flat per transaction; Xendit is Rp 4,500. QRIS fee is 0.7% MDR. For small transaction volumes (< 1,000/month), the flat fee structure matters more than percentage. Accounting export to Jurnal.id/Accurate adds cost (subscription to external software) but is operationally valuable for audit preparation. **Recommendation**: Negotiate volume-based pricing with payment providers once transaction volume is established.

**CTO**: The API design principles (Section 6) are solid: cursor-based pagination, consistent envelope, proper versioning, multi-auth (JWT + API Key + OAuth2). The circuit breaker pattern (Section 10) is essential for resilience. The dead letter queue design is thorough. **Concerns**: (1) The webhook outbound system (Section 7) is over-engineering for Phase 1 -- who are the external subscribers? Until there is a marketplace/ecosystem, this can wait. (2) The inbound API (Section 8) with staging tables is well-designed but adds complexity. For MVP, CSV upload via UI (K020 already supports this) is sufficient. (3) The `integrasi_config.credentials` JSONB storing encrypted credentials needs a proper secrets management strategy (Vault, KMS, or at minimum application-level encryption with key rotation).

**COO**: Every integration requires: (1) setup and configuration per tenant, (2) ongoing monitoring, (3) troubleshooting when things break. With six integration types, the support burden is significant. The integration health dashboard (Section 10) is a good mitigation, but someone needs to be trained to read and act on it. **Question**: For schools without IT staff, who monitors integration health? This likely falls on the SekolahPro support team, creating a support scaling challenge.

**CMO**: Payment gateway integration (VA, QRIS) is the enabler for remote top-up, which is a parent-facing feature. "Top up your child's account from any bank" is a strong convenience pitch. The school sync integration reduces onboarding friction: student data flows automatically, no manual re-entry.

**Issues Found**:
1. **[HIGH]** Too many integrations for Phase 1. Each one is a separate implementation, testing, and maintenance burden. Prioritize ruthlessly.
2. **[MEDIUM]** Webhook outbound (Section 7) has no clear consumer in MVP. Who subscribes to these webhooks? Defer until there is a concrete integration partner.
3. **[MEDIUM]** Secrets management for `integrasi_config.credentials` needs explicit strategy. JSONB with "encrypted" is too vague. Specify: encryption algorithm, key management (where is the encryption key?), rotation policy.
4. **[MEDIUM]** The dual-mode terminology in webhook event naming (`loan.approved` vs `pembiayaan.approved`) means external subscribers must handle two different event names for the same semantic event. Consider using canonical event names internally and adding `labels` in the payload for display.
5. **[LOW]** Bank statement import (MT940) is a niche feature. Most school koperasi bank with small local banks that do not provide MT940 exports. This is nice-to-have at best.

**Missing Items**:
- Secrets management architecture (encryption key storage, rotation)
- Payment provider onboarding checklist (merchant registration, sandbox testing, go-live approval)
- API documentation strategy (OpenAPI/Swagger spec, developer portal)
- Integration testing strategy (how to test payment callbacks in CI/CD)
- SLA expectations per integration (what uptime do we guarantee when dependent on third parties?)

**MVP Recommendation**: **MVP (Minimal)** -- School sync (event-driven, basic person data sync) + Payment gateway (VA only, one provider) + API design standards (apply to all internal APIs from day one). Defer: QRIS, accounting export, regulatory e-filing API, webhook outbound, inbound API endpoints, bank statement import to Phase 2/3.

---

## Phasing Recommendation

### Phase 1 (MVP) -- Core Operations + Minimal Extensions

**Goal**: A functional koperasi system that can operate daily and provide basic self-service.

| Module | Scope | Rationale |
|--------|-------|-----------|
| K019 (Toko/Kantin) | POS checkout + cash/TABUNGAN_DEBIT payment + basic stock in/out + spending limit | Primary revenue source; cashless payment is the hook |
| K020 (Payroll) | CSV upload + matching + priority deduction + approval + execution | Ensures collection rates; low implementation cost |
| K022 (Notifikasi) | IN_APP + WhatsApp (CRITICAL/HIGH events only) | Required for basic operational communication |
| K023 (Dashboard) | Teller workspace + Nasabah portal + Parent portal (basic) | Operational necessity + self-service + parent engagement |
| K024 (Integrasi) | School sync (basic) + VA payment gateway (one provider) + API standards | Foundation for future integrations; VA enables remote top-up |

**NOT in MVP**: K021 (E-Wallet/Uang Saku Digital). The cashless payment capability is already provided by K019's TABUNGAN_DEBIT payment method. K021 adds card management, advanced parent controls, and proxy features that are valuable but not essential for initial launch.

### Phase 2 -- Enhanced Experience + E-Wallet

**Goal**: Differentiated experience with advanced parent controls and operational maturity.

| Module | Scope | Rationale |
|--------|-------|-----------|
| K021 (Uang Saku Digital) | Card issuance + NFC/QR identification + parent controls + proxy | The differentiator; builds on proven K019 POS |
| K019 (Toko/Kantin) | Supplier management + PO workflow + stock opname + meal plans | Operational maturity for larger koperasi |
| K022 (Notifikasi) | SMS fallback + batch notifications + template management UI + quiet hours | Improved reliability and customization |
| K023 (Dashboard) | Admin/Supervisor dashboards + Kepala Sekolah dashboard + analytics + bulk export | Management visibility |
| K024 (Integrasi) | QRIS + accounting export + webhook outbound + multi-provider payment | Expanded payment and integration options |

### Phase 3 -- Ecosystem + Compliance Automation

| Module | Scope | Rationale |
|--------|-------|-----------|
| K019 (Toko/Kantin) | Multi-location + inter-location transfer + advanced reporting | Scale features |
| K022 (Notifikasi) | PUSH notifications (native app) + advanced analytics on delivery | Requires mobile app |
| K023 (Dashboard) | Native mobile app (Flutter) + offline capability | Phase 2 feedback informs design |
| K024 (Integrasi) | Regulatory e-filing API + bank statement import + inbound API + dead letter management UI | Compliance automation |

---

## Cross-Cutting Concerns (K019-K024)

### 1. Spending Limit Fragmentation (CRITICAL)

Spending limits are defined in three separate places:
- **K019 Section 7**: `spending_limit_config` (POS spending limit per nasabah)
- **K021 Section 2**: `ewallet_config` (e-wallet spending limits including daily/weekly/monthly/per-tx + category + time)
- **K023 Section 4 (data model 8)**: `parent_child_config` (parent-set spending limits)

**Resolution**: Create ONE canonical spending limit entity. Recommend using an enhanced version of `ewallet_config` (renamed to `spending_control_config`) that is referenced by K019 POS, K021 card system, and K023 parent portal. The `parent_child_config` should contain ONLY the parent-child link and notification preferences, not duplicate spending limits.

### 2. Offline Capability Gap

K019 (POS) and K021 (card payment) have no offline strategy. During lunch rush with 500+ students, network failures would halt all transactions. This is unacceptable for a POS system.

**Resolution**: Document an offline strategy even if deferred to Phase 2. Minimum viable: local transaction queue with sync-on-reconnect, risk-based offline limit (e.g., transactions under Rp 50,000 processed offline, reconciled when online).

### 3. Parent Identity and Authentication

K021 and K023 both reference "parent" access but the authentication mechanism for parents is not clearly defined. Are parents:
- Registered as nasabah first (with their own member_number)?
- Given portal-only access (no financial account)?
- Using a separate authentication system?

**Resolution**: Clarify in K001 (nasabah) how parent accounts work. If parents are nasabah, they should go through the standard registration flow. If parents are portal-only users, define a separate user entity.

### 4. Notification Cost at Scale

K022 + K023 (parent notifications) + K019 (spending alerts) + K020 (deduction receipts) combined could generate 10-20 notifications per active member per month. For a koperasi with 2,000 members, that is 20,000-40,000 messages/month. At Rp 100/message (third-party WA), that is Rp 2-4 million/month. This must be factored into the koperasi's operating budget and potentially passed to members as a service fee.

### 5. Integration Testing Complexity

K024 defines integrations with: school management system, payment gateway, WhatsApp API, SMS gateway, accounting software, and regulatory systems. Each requires mock/sandbox environments for testing. CI/CD pipeline must support all integration test modes without hitting production endpoints.

---

## Recommended Adjustments

### Immediate (Before Development Starts)

1. **Consolidate spending limits** into a single entity referenced across K019, K021, and K023. Update all three ADRs.
2. **Rename K021** from "E-Wallet" to "Uang Saku Digital" or "Kartu Belanja Siswa" to avoid regulatory confusion.
3. **Add regulatory compliance statement** to K021 explicitly stating this is NOT e-money under BI regulations.
4. **Define parent authentication model** clearly in K001 or a new ADR.
5. **Create a go-live checklist** for K022 WhatsApp template preparation.

### Architectural

6. **Design K019 POS as a bounded context** with clear interfaces to the financial core. This allows future separation into a standalone service if needed.
7. **Implement feature flags** for all extension modules (partially documented in K019, should be standardized across all).
8. **Add offline strategy section** to K019 and K021 even if implementation is deferred.
9. **Standardize webhook event naming** in K024 to use canonical names (not dual-mode variants).
10. **Specify secrets management architecture** in K024 for credential storage.

### Operational

11. **Create training material plans** alongside each module deployment. K019 POS requires the most staff training.
12. **Budget notification costs** per tenant. Provide cost calculator for koperasi administrators.
13. **Define SLA expectations** for third-party integrations (payment gateway uptime, WhatsApp delivery rate).
14. **Plan card procurement pipeline** for K021 Phase 2 launch (NFC card vendors, lead times, minimum order quantities).

---

## Summary Verdict

| ADR | Verdict | MVP Phase | Priority |
|-----|---------|-----------|----------|
| K019 Toko/Kantin | APPROVED WITH ADJUSTMENTS | MVP (minimal POS) | P0 |
| K020 Payroll | APPROVED | MVP | P0 |
| K021 E-Wallet | APPROVED WITH ADJUSTMENTS | Phase 2 | P1 |
| K022 Notifikasi | APPROVED WITH ADJUSTMENTS | MVP (minimal) | P0 |
| K023 Dashboard | APPROVED WITH ADJUSTMENTS | MVP (3 portals only) | P0 |
| K024 Integrasi | APPROVED WITH ADJUSTMENTS | MVP (minimal) | P0 |

**Overall Assessment**: The six ADRs are well-designed individually but collectively represent a scope that is 3-4x larger than what should be shipped in MVP. The phasing recommendation above reduces MVP scope to approximately 40% of the total documented functionality while preserving the core value proposition: a koperasi system with cashless POS, payroll deduction, self-service portal, and parent monitoring.

The single most important cross-cutting fix is **spending limit consolidation** -- this affects three ADRs and must be resolved before development starts to avoid technical debt that will be expensive to untangle later.

---

*Review conducted by C-Suite Advisory Panel*
*Date: 2026-04-15*
