# ADR — Management Koperasi Sekolah

Keputusan arsitektur yang spesifik untuk subproject **Management Koperasi Sekolah**.
Mendukung dual-mode: **Koperasi Konvensional** dan **BMT (Baitul Maal wat Tamwil)** — lihat [ADR-009](../core/ADR-009-dual-mode-institution-type.md).

> ADR di sini hanya berlaku untuk subproject ini. Untuk keputusan yang berlaku lintas subproject, lihat [core/](../core/).

> **Sprint Order**: Koperasi ADR tercakup di **Fase 3, 4B, 5B, 6B, 7B, 10** — lihat [SPRINT-ORDER.md](../SPRINT-ORDER.md)

## Index & Implementation Status

### Legend

| Symbol | Arti |
|--------|------|
| `-` | Belum mulai |
| `~` | In progress |
| `v` | Selesai |
| `x` | Tidak diperlukan |

---

### Core Domain

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K001 | [Nasabah (Anggota/Member)](./ADR-K001-nasabah.md) | v | - | - | - | Application + approval, RBAC, ahli waris, KYC, transfer cabang, merge duplikat |
| K002 | [Rekening](./ADR-K002-rekening.md) | v | - | - | - | Form pengajuan + approval, auto-create wajib, balance management, dormant handling |
| K003 | [Produk & Akad](./ADR-K003-produk-akad.md) | v | - | - | - | Katalog produk, akad syariah, rate/nisbah, product lifecycle, eligibility rules |

### Simpanan

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K004 | [Simpanan Pokok & Wajib](./ADR-K004-simpanan-pokok-wajib.md) | v | - | - | - | Collection mechanisms, payroll deduction, student collection, arrears tracking |
| K005 | [Tabungan](./ADR-K005-tabungan.md) | v | - | - | - | Multi-produk, bagi hasil/bunga, tabungan berencana, parent top-up, student savings |
| K006 | [Deposito / Simpanan Berjangka](./ADR-K006-deposito.md) | v | - | - | - | Multi-tenor, rate tiers, maturity handling, rollover, early withdrawal, bilyet |

### Pembiayaan / Pinjaman

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K007 | [Pinjaman / Pembiayaan](./ADR-K007-pinjaman.md) | v | - | - | - | Multi-level approval, credit scoring, NPL classification, restructuring, penjamin |
| K008 | [Angsuran & Jadwal](./ADR-K008-angsuran-jadwal.md) | v | - | - | - | Schedule generation, flat/declining/anuitas, payment channels, DPD tracking |
| K009 | [Denda & Penalti](./ADR-K009-denda-penalti.md) | v | - | - | - | Denda bunga, ta'zir (social fund), penalty cap, waiver, grace period |
| K010 | [Jaminan / Agunan](./ADR-K010-jaminan-agunan.md) | v | - | - | - | Internal/physical collateral, LTV, penjamin, rahn, foreclosure, document mgmt |

### Transaksi & Operasional

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K011 | [Transaksi Rekening & Non Rekening](./ADR-K011-transaksi.md) | v | - | - | - | Transaction engine, immutability, atomic balance, validation, reversal, batch |
| K012 | [Teller Session](./ADR-K012-teller-session.md) | v | - | - | - | Open/close session, cash balancing, variance handling, daily consolidation |
| K013 | [Money Denomination](./ADR-K013-money-denomination.md) | v | - | - | - | IDR denominations, cash count, vault management |
| K014 | [Kas & Cash Flow](./ADR-K014-kas-cashflow.md) | v | - | - | - | Daily cash position, inflow/outflow, inter-branch transfer, reconciliation |

### Akuntansi & Pelaporan

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K015 | [Jurnal & Akuntansi (COA)](./ADR-K015-jurnal-coa.md) | v | - | - | - | Double-entry, COA per PSAK/SAK ETAP + PSAK Syariah, auto-journal, period mgmt |
| K016 | [SHU (Sisa Hasil Usaha)](./ADR-K016-shu.md) | v | - | - | - | Annual calculation, jasa modal/usaha, distribution formula, RAT approval |
| K017 | [Laporan Regulasi](./ADR-K017-laporan-regulasi.md) | v | - | - | - | Dinas Koperasi, OJK/LKM, PSAK Syariah reports, financial ratios, scheduling |

### Khusus BMT (Islamic Mode)

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K018 | [Zakat & Infaq](./ADR-K018-zakat-infaq.md) | v | - | - | - | Zakat mal/fitrah, nisab/haul, mustahik 8 asnaf, ta'zir fund, off-balance sheet |

### Stakeholder Extensions

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K019 | [Koperasi Toko & Kantin](./ADR-K019-toko-kantin.md) | v | - | - | - | POS, inventory, supplier, canteen menu, student spending limit, asrama meal plan |
| K020 | [Payroll Integration & Auto-Deduction](./ADR-K020-payroll-deduction.md) | v | - | - | - | Salary deduction guru/staff, priority rules, batch processing, authorization |
| K021 | [E-Wallet / Uang Saku Digital](./ADR-K021-ewallet.md) | v | - | - | - | Student digital wallet, parent controls, spending limits, QR/NFC payment |
| K022 | [Notifikasi & Komunikasi](./ADR-K022-notifikasi.md) | v | - | - | - | WhatsApp, SMS, email, push, event-driven, template management |
| K023 | [Dashboard & Self-Service Portal](./ADR-K023-dashboard-portal.md) | v | - | - | - | Role-based dashboards, nasabah portal, parent portal, kepala sekolah oversight |
| K024 | [Integrasi & API](./ADR-K024-integrasi-api.md) | v | - | - | - | School system sync, payment gateway, OJK reporting, webhooks, API design |

### Governance & Compliance (CRITICAL)

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K025 | [Cooperative Governance & Internal Controls](./ADR-K025-governance-internal-controls.md) | v | - | - | - | Struktur pengurus/pengawas, authority matrix, segregation of duties, internal audit, CoI |
| K026 | [RAT (Rapat Anggota Tahunan) Management](./ADR-K026-rat-management.md) | v | - | - | - | Meeting management, quorum, voting, elections, notulensi, follow-up actions |
| K027 | [AML/CFT Compliance](./ADR-K027-aml-cft-compliance.md) | v | - | - | - | CDD/EDD, PEP screening, transaction monitoring, PPATK reporting, record retention |
| K028 | [Data Privacy (UU PDP)](./ADR-K028-data-privacy-uupdp.md) | v | - | - | - | Consent management, data subject rights, breach notification, DPA, privacy by design |
| K029 | [Business Continuity & Disaster Recovery](./ADR-K029-business-continuity-drp.md) | v | - | - | - | RTO/RPO, backup strategy, HA architecture, incident response, DR tiers |

### Operational Excellence (HIGH)

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K030 | [Membership Lifecycle Management](./ADR-K030-membership-lifecycle.md) | v | - | - | - | Resignation, expulsion, death settlement, dormant management, transfer |
| K031 | [Cooperative Health Indicators & Risk Management](./ADR-K031-cooperative-health-indicators.md) | v | - | - | - | KPI dashboard, early warning system, stress testing, composite health score |
| K032 | [Biometric Authentication](./ADR-K032-biometric-authentication.md) | v | - | - | - | Fingerprint/face recognition, liveness detection, fallback chain |
| K033 | [Loan Collection Management](./ADR-K033-loan-collection-management.md) | v | - | - | - | Aging buckets, collection workflow, restructuring, write-off, collector tracking |
| K034 | [Reserve Fund Management](./ADR-K034-reserve-fund-management.md) | v | - | - | - | Statutory/general/investment reserve, auto SHU allocation, investment instruments |

### Growth & Innovation (MEDIUM)

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K035 | [Insurance / Takaful Integration](./ADR-K035-insurance-takaful-integration.md) | v | - | - | - | Credit life insurance, takaful, claim processing, premium collection |
| K036 | [Mobile App Strategy](./ADR-K036-mobile-app-strategy.md) | v | - | - | - | Flutter hybrid, offline-first, app variants, mobile security |
| K037 | [Member Education Program](./ADR-K037-member-education-program.md) | v | - | - | - | Course management, mandatory training, quiz, certificates, KPI tracking |

### Strategic (LOW)

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K038 | [Cooperative Dissolution Process](./ADR-K038-cooperative-dissolution.md) | v | - | - | - | Voluntary/government dissolution, liquidation, asset distribution |
| K039 | [Multi-Branch Consolidation Reporting](./ADR-K039-multi-branch-consolidation.md) | v | - | - | - | Inter-branch elimination, consolidated financials, branch comparison |
| K040 | [Digital Transformation Roadmap](./ADR-K040-digital-transformation-roadmap.md) | v | - | - | - | Technology evolution, AI/ML opportunities, API versioning, cloud strategy |

---

## Items per Domain

Ringkasan entity/item utama yang terlibat di setiap ADR:

| ADR | Items |
|-----|-------|
| K001 | nasabah, anggota, identitas, kontak, status_keanggotaan, ahli_waris, specimen, kyc_level, sponsor, transfer_cabang, merge |
| K002 | rekening, rekening_application, form_template, nomor_rekening, saldo, status_rekening, produk_ref, dormant |
| K003 | produk, jenis_produk, akad, nisbah, margin, suku_bunga, product_version, eligibility_rule, fee_config |
| K004 | simpanan_pokok, simpanan_wajib, collection_batch, arrears_tracking, payroll_deduction, refund |
| K005 | tabungan, setoran, penarikan, bagi_hasil, bunga, tabungan_berencana, statement, parent_topup |
| K006 | deposito, tenor, jatuh_tempo, rollover, pencairan_dini, bilyet, rate_tier, maturity_tracking |
| K007 | pinjaman, pembiayaan, plafon, tenor, credit_scoring, penjamin, disbursement, NPL, restructuring, write_off |
| K008 | angsuran, jadwal_angsuran, pokok, margin_bunga, sisa_pokok, DPD, payment_channel, rescheduling |
| K009 | denda, ta_zir, ta_widh, penalty_cap, grace_period, waiver, social_fund |
| K010 | jaminan, agunan, rahn, penjamin, taksasi, LTV, dokumen_jaminan, foreclosure, release |
| K011 | transaksi, jurnal_entry, debit, kredit, referensi, reversal, batch, receipt, validation |
| K012 | sesi_teller, kas_awal, kas_akhir, selisih, denomination_count, handover, force_close |
| K013 | denominasi, pecahan, jumlah_lembar, total_nominal, vault, cash_count |
| K014 | kas_harian, pemasukan, pengeluaran, saldo_kas, inter_branch_transfer, consolidation |
| K015 | coa, akun, jurnal, jurnal_line, neraca, laba_rugi, period, posting, auto_journal |
| K016 | shu, jasa_modal, jasa_usaha, cadangan, distribusi, RAT, simulasi |
| K017 | laporan_ojk, laporan_dinas, NPL_ratio, CAR, BMPK, financial_ratios, scheduling |
| K018 | zakat, infaq, shadaqah, muzakki, mustahik, nisab, haul, ta_zir_fund, distribution |
| K019 | toko_produk, inventory, POS, penjualan, supplier, canteen_menu, meal_plan, spending_limit |
| K020 | payroll_batch, deduction, authorization, priority_rules, matching, opt_in |
| K021 | ewallet_config, ewallet_card, spending_limit, parent_control, QR_payment, NFC |
| K022 | notifikasi_template, notifikasi_log, preference, channel, WhatsApp_API, escalation |
| K023 | dashboard_widget, dashboard_layout, portal_session, parent_child_config, export |
| K024 | integrasi_config, integrasi_log, webhook_subscription, api_key, dead_letter, sync |
| K025 | governance_position, authority_matrix, internal_audit, conflict_of_interest, segregation_of_duties |
| K026 | rat_meeting, rat_agenda_item, rat_attendance, rat_vote, rat_election, rat_election_candidate |
| K027 | aml_customer_due_diligence, aml_transaction_monitoring, aml_monitoring_rule, aml_report |
| K028 | consent_record, data_subject_request, data_breach, data_processing_agreement |
| K029 | backup_strategy, dr_tier, incident_record, offline_procedure |
| K030 | membership_lifecycle_event, resignation_clearing, death_settlement, dormant_management |
| K031 | kpi_definition, kpi_measurement, early_warning_alert, stress_test_scenario, health_score |
| K032 | biometric_enrollment, biometric_verification_log |
| K033 | collection_case, collection_activity, collector_performance |
| K034 | reserve_fund, reserve_fund_transaction |
| K035 | insurance_product, insurance_policy, insurance_claim |
| K036 | mobile_app_config, app_variant, offline_queue |
| K037 | education_course, education_enrollment, education_kpi |
| K038 | dissolution_process, liquidation_stage |
| K039 | branch_financial_summary, elimination_rule |
| K040 | (strategic ADR — no specific data model) |
