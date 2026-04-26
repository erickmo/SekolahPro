# Checklist Traceability ADR → Dokumentasi Koperasi

**Tanggal Validasi:** 2026-04-26  
**Validator:** Domain Expert — Koperasi/BMT Indonesia  
**Jumlah ADR:** 40 (K001–K040)  
**Jumlah Doc Files:** 14 (01-domain-overview.md s/d 14-digital-transformation-roadmap.md)

---

## Legenda

| Coverage Level | Deskripsi |
|----------------|-----------|
| **FULL** | Semua keputusan kunci ADR tercakup di dokumentasi dengan akurasi tinggi |
| **PARTIAL** | Sebagian besar konten ADR ada, tapi ada celah atau inakurasi |
| **MISSING** | Konten ADR tidak ada atau sangat minim di dokumentasi |

| Prioritas Gap | Deskripsi |
|---------------|-----------|
| **P0** | Kritis — wajib ada untuk compliance regulasi atau menghindari bug sistem |
| **P1** | Penting — diperlukan untuk implementasi yang benar |
| **P2** | Nice to have — melengkapi pemahaman, bukan blocker |

---

## Tabel Traceability

| ADR ID | ADR Title | Covered In Doc | Coverage Level | Missing/Gaps | Priority |
|--------|-----------|----------------|----------------|--------------|----------|
| **K001** | Nasabah (Anggota/Member) | 02-member-rekening.md, 01-domain-overview.md | **FULL** | — | — |
| **K002** | Rekening (Akun Nasabah) | 02-member-rekening.md | **FULL** | Tidak ada `transfer_pair_id` di doc rekening_application (minor) | P2 |
| **K003** | Produk & Akad | 03-produk-financial.md, 01-domain-overview.md | **FULL** | — | — |
| **K004** | Simpanan Pokok & Wajib | 03-produk-financial.md | **PARTIAL** | Grace period simpanan wajib di doc = 7 hari; ADR tidak menyebut angka default. Aturan kenaikan nominal (batch top-up) tidak ada di doc. Wali kelas collection flow hanya disebut singkat. | P1 |
| **K005** | Tabungan (Simpanan Sukarela) | 03-produk-financial.md | **PARTIAL** | Negative profit handling Mudharabah (3-months consecutive loss rule) tidak ada di doc. `max_transactions_daily` field tidak disebutkan. Early withdrawal tabungan berencana rule (Supervisor+) tidak dinyatakan eksplisit. | P1 |
| **K006** | Deposito / Simpanan Berjangka | 03-produk-financial.md | **PARTIAL** | Maturity date edge cases (akhir bulan) ada di doc secara ringkas, tapi teacher/staff bonus rate config tidak ada di doc. Payroll auto-placement untuk deposito tidak ada. Interest payment method MONTHLY vs AT_MATURITY vs CAPITALIZE tidak dijelaskan di doc. | P1 |
| **K007** | Pinjaman / Pembiayaan | 04-pinjaman-loan.md | **FULL** | Proses pencairan detail (wakalah Murabahah, bukti pembelian) hanya singkat di doc. Pinjaman darurat fast-track tidak disebutkan. | P2 |
| **K008** | Angsuran & Jadwal | 04-pinjaman-loan.md | **FULL** | Concurrency control (SELECT FOR UPDATE) ada di doc. Musyarakah/Mudharabah projected profit review per-3-bulan tidak disebutkan di doc. | P2 |
| **K009** | Denda & Penalti | 04-pinjaman-loan.md | **FULL** | Ta'widh explained but `max_admin_fee_pct` untuk Qardh tidak disebutkan. | P2 |
| **K010** | Jaminan / Agunan | 04-pinjaman-loan.md | **PARTIAL** | Ujrah untuk Rahn (biaya penyimpanan) ada di doc tapi tanpa formula detail. Revaluasi periodik wajib (12 bulan SURAT_BERHARGA, 6 bulan BARANG_BERGERAK) ada di doc. Foreclosure surplus ke nasabah ada. Tapi ijazah sebagai jaminan moral (konteks sekolah) tidak ada di doc sama sekali. | P1 |
| **K011** | Transaksi | 05-transaksi-teller.md | **FULL** | — |  — |
| **K012** | Teller Session | 05-transaksi-teller.md | **FULL** | — | — |
| **K013** | Money Denomination | 05-transaksi-teller.md | **FULL** | — | — |
| **K014** | Kas & Cash Flow | 06-keuangan-accounting.md | **PARTIAL** | Kas minimum per branch configurable ada di doc. Petty cash management tidak disebutkan di doc sama sekali. Reconciliation formula (expected_kas_closing = teller + vault + petty cash) disebut tapi petty cash dikesampingkan. Inter-branch transfer denominasi detail tidak ada. | P1 |
| **K015** | Jurnal & COA (Akuntansi) | 06-keuangan-accounting.md | **FULL** | Journal mapping auto-post option tidak disebutkan eksplisit. PPAP entries (terkait K007 NPL) tidak ada di doc mapping table. | P1 |
| **K016** | SHU (Sisa Hasil Usaha) | 06-keuangan-accounting.md | **FULL** | — | — |
| **K017** | Laporan Regulasi | 07-kepatuhan-regulasi.md | **FULL** | — | — |
| **K018** | Zakat & Infaq | 06-keuangan-accounting.md | **PARTIAL** | Haul dalam tahun hijriah (355 hari) ada di doc. Pengurangan hutang: "angsuran 12 bulan ke depan" ada di doc. Tapi: koleksi zakat fitrah bulk per kelas/halaqah (konteks pesantren) tidak ada. Penyaluran ke 8 asnaf disebutkan tapi tidak dirinci mana yang diutamakan. Zakat institusi (2.5% laba BMT) tidak ada di doc. | P1 |
| **K019** | Toko & Kantin | 08-toko-hr.md | **FULL** | Kantin meal plan dan boarding integration ada di doc. — | — |
| **K020** | Payroll Integration & Auto-Deduction | 08-toko-hr.md | **FULL** | — | — |
| **K021** | Uang Saku Digital (E-Wallet) | 09-digital-features.md | **FULL** | — | — |
| **K022** | Notifikasi & Komunikasi | 12-komunikasi-integrasi.md | **FULL** | — | — |
| **K023** | Dashboard & Self-Service Portal | 12-komunikasi-integrasi.md | **FULL** | — | — |
| **K024** | Integrasi & API Gateway | 12-komunikasi-integrasi.md | **FULL** | — | — |
| **K025** | Governance & Internal Controls | 10-governance.md | **PARTIAL** | Four-eyes principle ada di doc. Rotasi teller 6 bulan ada di doc. Internal audit schedule (min 1x/semester) ada. Tapi: authority matrix detail (approval limit per jabatan struktural) tidak ada di doc. Konflik kepentingan > Rp 500.000 wajib lapor tidak ada di doc. Selisih kas threshold (Rp 50rb→Supervisor, Rp 500rb→Manager) ada di ADR K025 tapi di doc K025 berbeda — doc menggunakan threshold Rp 10.000/100.000 dari K012, bukan angka K025. | P0 |
| **K026** | RAT Management | 10-governance.md | **FULL** | — | — |
| **K027** | AML/CFT Compliance | 11-kepatuhan-advanced.md | **FULL** | — | — |
| **K028** | Data Privacy & UU PDP | 11-kepatuhan-advanced.md | **FULL** | — | — |
| **K029** | Business Continuity & DRP | 11-kepatuhan-advanced.md | **FULL** | — | — |
| **K030** | Membership Lifecycle Management | 10-governance.md | **PARTIAL** | Refund formula ada di doc. Dormant trigger (12 bulan + 2 RAT + 6 bulan wajib) ada di doc. TAPI: proses ekskumul (due process 14 hari) tidak detail di doc. Prioritas ahli waris faraidh vs KUHPerdata hanya disebut singkat. Clearing period pengunduran diri 30 hari ada. TRANSFERRED_OUT dan TRANSFERRED_IN status tidak ada di state machine doc. | P1 |
| **K031** | Cooperative Health Indicators | 10-governance.md | **PARTIAL** | KPI CAMELS ada di doc. Rating AA-C ada di doc. Stress testing wajib 1x/tahun ada. TAPI: formula komposit health score (Capital 20%, Asset Quality 25%, Management 15%, Earnings 20%, Liquidity 20%) tidak ada di doc. Early warning escalation (assignment ke PIC + deadline) tidak ada di doc. | P1 |
| **K032** | Biometric Authentication | 09-digital-features.md | **FULL** | — | — |
| **K033** | Loan Collection Management | 13-keuangan-advanced.md | **FULL** | — | — |
| **K034** | Reserve Fund Management | 13-keuangan-advanced.md | **FULL** | — | — |
| **K035** | Insurance / Takaful Integration | 13-keuangan-advanced.md | **FULL** | — | — |
| **K036** | Mobile App Strategy | 09-digital-features.md | **FULL** | — | — |
| **K037** | Member Education Program | 10-governance.md | **PARTIAL** | COOP_BASIC wajib 30 hari ada di doc. PRODUCT_TRAINING sebelum pinjaman ada. ISLAMIC_FINANCE wajib untuk anggota baru BMT ada. TAPI: impact tracking (NPL rate vs trained/untrained) tidak ada di doc. Kursus completion requirement sebelum bisa ajukan pinjaman tidak disebutkan eksplisit. | P2 |
| **K038** | Cooperative Dissolution | 10-governance.md | **PARTIAL** | Urutan pembayaran saat likuidasi ada di doc. Claim period 60 hari ada. Arsip 5 tahun ada. TAPI: urutan lengkap pembubaran (RAT → otoritas → tim likuidasi → pengumuman → settlement) tidak ada di doc secara step-by-step. Perlakuan khusus BMT saat pembubaran (ZIS disalurkan dulu) hanya disebut singkat. | P1 |
| **K039** | Multi-Branch Consolidation | 10-governance.md | **PARTIAL** | Eliminasi inter-branch ada di doc. Empat level pelaporan ada di doc. TAPI: mekanisme teknis eliminasi (akun inter-branch 1901/2901) tidak ada di doc (hanya ada di 07-kepatuhan doc). Alokasi biaya pusat ke cabang tidak disebutkan sama sekali di doc. | P1 |
| **K040** | Digital Transformation Roadmap | 14-digital-transformation-roadmap.md | **FULL** | — | — |

---

## Ringkasan Coverage

| Status | Jumlah ADR | Persentase |
|--------|-----------|------------|
| **FULL** | 25 | 62.5% |
| **PARTIAL** | 15 | 37.5% |
| **MISSING** | 0 | 0% |
| **TOTAL** | 40 | 100% |

### ADR dengan Gap P0 (Kritis)

| ADR | Gap |
|-----|-----|
| K025 | Threshold selisih kas di doc (10k/100k dari K012) konflik dengan ADR K025 (50k/500k). Ini adalah dua sumber berbeda yang akan menyebabkan bug jika implementor mengikuti doc kepatuhan. |

### ADR dengan Gap P1 (Penting)

K004, K005, K006, K010, K014, K015, K018, K030, K031, K038, K039 — lihat kolom Missing/Gaps di atas untuk detail.

### ADR dengan Gap P2 (Nice to Have)

K002, K007, K008, K009, K037 — gap minor, tidak blocking.

---

*Dokumen ini dihasilkan dari proses validasi sistematis ADR-K001 s/d ADR-K040 terhadap 14 file dokumentasi di /spec/koperasi/. Diperbarui: 2026-04-26.*
