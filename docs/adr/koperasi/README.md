# ADR — Management Koperasi Sekolah

Keputusan arsitektur yang spesifik untuk subproject **Management Koperasi Sekolah**.
Mendukung dual-mode: **Koperasi Konvensional** dan **BMT (Baitul Maal wat Tamwil)** — lihat [ADR-009](../core/ADR-009-dual-mode-institution-type.md).

> ADR di sini hanya berlaku untuk subproject ini. Untuk keputusan yang berlaku lintas subproject, lihat [core/](../core/).

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
