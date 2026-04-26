# 13 - Keuangan Lanjutan: Dana Cadangan, Penagihan Pinjaman, dan Asuransi/Takaful

Dokumen ini mendeskripsikan tiga domain keuangan lanjutan Modul Koperasi SekolahPro: (1) Manajemen Dana Cadangan (Reserve Fund) sesuai UU Koperasi No. 25/1992 Pasal 45 (ADR-K034), (2) Manajemen Penagihan Pinjaman dengan aging bucket dan escalation workflow (ADR-K033), dan (3) Integrasi Asuransi/Takaful sebagai proteksi kredit (ADR-K035). Ketiga domain ini saling terkait dalam siklus pinjaman — dari proteksi awal (asuransi), pemantauan berkala (collection), hingga penyangga kerugian akhir (dana cadangan).

---

## ADR References

| ADR | Judul | Status |
|-----|-------|--------|
| ADR-K033 | Loan Collection Management | Accepted |
| ADR-K034 | Reserve Fund Management (Dana Cadangan) | Accepted |
| ADR-K035 | Insurance / Takaful Integration | Accepted |

---

## Domain Entities

### Dana Cadangan (K034)

| Entity | Deskripsi |
|--------|-----------|
| `reserve_fund` | Master data dana cadangan — tipe (statutory/general/investment/special), saldo saat ini, target, konfigurasi alokasi SHU, instrumen investasi. |
| `reserve_fund_transaction` | Riwayat transaksi masuk/keluar dana cadangan — allocation, return investasi, withdrawal, loss coverage. Memerlukan approval untuk sebagian besar jenis. |

### Penagihan (K033)

| Entity | Deskripsi |
|--------|-----------|
| `collection_case` | Kasus penagihan per pinjaman — aging bucket, total overdue, assignment collector, escalation level, status penyelesaian. |
| `collection_activity` | Log setiap aksi penagihan — telepon, kunjungan, SP letter, negosiasi, janji bayar. Termasuk foto/dokumen bukti. |
| `collector_performance` | Metrik performa per collector per bulan — contact rate, resolution rate, collection rate, promise kept rate. |

### Asuransi/Takaful (K035)

| Entity | Deskripsi |
|--------|-----------|
| `insurance_product` | Produk asuransi/takaful — tipe coverage, jenis premi, eligibility, applicable mode (konvensional/syariah). |
| `insurance_policy` | Polis per pinjaman per nasabah — coverage amount, premium, tanggal efektif, status. |
| `insurance_claim` | Klaim asuransi — tipe klaim, dokumen, assessed amount, status, tanggal settlement. |

---

## Business Rules

### Dana Cadangan

1. **Empat Jenis Dana Cadangan.**
   - **Statutory Reserve (Dana Cadangan Wajib):** Sumber minimal 25% dari SHU per tahun. Tujuan: menutupi kerugian dan jaminan simpanan. Tidak bisa dibagi ke anggota. Target: 25% dari total simpanan anggota.
   - **General Reserve (Dana Cadangan Umum):** Sumber dari sisa SHU di atas statutory minimum. Untuk pengembangan usaha dan darurat.
   - **Investment Reserve (Dana Cadangan Investasi):** Disimpan di instrumen aman (deposito, SBN, reksa dana pasar uang). Return otomatis di-reinvest.
   - **Special Reserve (Dana Cadangan Khusus):** Dibentuk dan dicairkan hanya atas keputusan RAT. Tujuan spesifik (gedung, IT system, dll).

2. **Alokasi SHU Otomatis.** Saat SHU disetujui di RAT (K016), sistem otomatis mengalokasikan ke dana cadangan sesuai persentase yang dikonfigurasi, sebelum sisa dibagikan ke anggota.

3. **Target Dana Cadangan Wajib.** Jika balance statutory reserve sudah mencapai target (25% dari total simpanan), alokasi SHU dapat dialihkan ke tujuan lain. Jika balance turun di bawah target, wajib top-up dari SHU berikutnya.

4. **Instrumen Investasi Diizinkan.** Deposito bank (rating ≥ BBB+, tenor max 12 bulan, max 25% per bank); SBN (SUN/SPN/SRI, tenor max 5 tahun); Reksa Dana Pasar Uang/Fixed Income (rating AAA, max 20% dari investment reserve). DILARANG: saham, crypto, properti langsung, spekulasi forex.

5. **Aturan Withdrawal per Tipe.**
   - **Statutory:** Hanya untuk tutup kerugian. Approval: Ketua + Bendahara + Pengawas. Tidak untuk operasional.
   - **General:** Untuk pengembangan usaha atau darurat. Approval: Ketua + Bendahara. Withdrawal > 25% balance tambah Pengawas.
   - **Investment:** Sesuai maturity instrumen. Premature withdrawal butuh approval Bendahara + Ketua.
   - **Special:** Hanya sesuai tujuan yang disetujui RAT. Jika tujuan berubah: harus kembali ke RAT.

6. **Journal Entries Dana Cadangan.**
   - Alokasi dari SHU: Debit SHU Belum Distribusi → Credit Dana Cadangan (ekuitas)
   - Withdrawal untuk tutup kerugian: Debit Dana Cadangan → Credit Akumulasi Kerugian
   - Investment placement: Debit Investasi-Deposito/SBN (aset) → Credit Kas/Bank
   - Investment return: Debit Kas/Bank → Credit Pendapatan Investasi (income)

### Penagihan Pinjaman

7. **Delapan Aging Bucket.**

| Bucket | DPD | Kualitas OJK | PPAP |
|--------|-----|-------------|------|
| CURRENT | 0 hari | Lancar | 1% |
| DPW_1_7 | 1-7 hari | Dalam Perhatian Khusus | 5% |
| DPW_8_30 | 8-30 hari | Dalam Perhatian Khusus | 5% |
| DPW_31_60 | 31-60 hari | Dalam Perhatian Khusus | 5% |
| DPW_61_90 | 61-90 hari | Dalam Perhatian Khusus | 5% |
| DPW_91_180 | 91-180 hari | Kurang Lancar (Substandard) | 15% |
| DPW_181_270 | 181-270 hari | Diragukan (Doubtful) | 50% |
| DPW_271_360 | 271-360 hari | Macet (Loss) | 100% |

8. **Strategi Penagihan Bertingkat.**
   - **0 hari:** Auto reminder H-3 sebelum jatuh tempo.
   - **1-7 hari:** Auto WA/SMS reminder. Flag di dashboard teller.
   - **8-30 hari:** Telepon oleh collector Level 1. Notifikasi ke penjamin.
   - **31-60 hari:** Surat Peringatan 1 (SP1). Home visit. Tawaran restructuring. Escalation ke supervisor collection.
   - **61-90 hari:** SP2. Multiple home visits. Meeting nasabah + penjamin. Mandatory restructuring offer. Escalation ke manager.
   - **91-180 hari:** SP3 (final). Penjamin diminta bertanggung jawab. Proses jaminan/agunan mulai. Persiapan legal action.
   - **181-270 hari:** Legal notice dari notaris. Eksekusi jaminan. Mediasi pihak ketiga.
   - **271-360 hari:** Write-off proposal ke management. Laporan ke OJK. Opsi penjualan NPL.
   - **> 360 hari:** Write-off execution. Blacklist nasabah.

9. **Restructuring Eligibility.** DPD ≥ 30 hari + nasabah menunjukkan willingness to pay + tidak dalam proses hukum + approval supervisor collection. Opsi: perpanjangan tenor (max 2x tenor awal), pengurangan angsuran, grace period 3-6 bulan (bayar pokok saja), atau kombinasi.

10. **Write-Off Criteria.** DPD > 360 hari + semua collection actions terdokumentasi + collateral sudah dieksekusi + penjamin sudah diklaim + legal action tidak feasible atau sudah gagal. Approval chain: Manager → Bendahara → Ketua.

11. **Collection Performance Tracking.** Empat metrik utama per collector per bulan: Contact Rate (%), Resolution Rate (%), Collection Rate (%), dan Promise Kept Rate (%). Digunakan untuk evaluasi dan insentif.

12. **Activity Logging Wajib.** Setiap aksi penagihan — telepon, kunjungan, surat, negosiasi — wajib dicatat di `collection_activity` beserta waktu, hasil kontak, respons nasabah, dan janji bayar jika ada. Foto kunjungan dapat dilampirkan.

### Asuransi/Takaful

13. **Lima Produk Asuransi.** Credit Life (jiwa kredit — meninggal/cacat total), Credit Disability (kecacatan sementara), Credit Unemployment (kehilangan pekerjaan — untuk guru/staff), Collateral Protection (asuransi jaminan — kebakaran, bencana), Personal Accident (kecelakaan diri — opsional).

14. **Tiga Coverage Types.** Decreasing term (coverage menurun sesuai outstanding — paling umum), Level term (coverage tetap selama tenor), Full outstanding (coverage = sisa outstanding saat klaim — paling protektif).

15. **Tiga Metode Premium Collection.** Single upfront (dibayar sekali saat disbursement), Monthly deducted (dipotong dari angsuran bulanan — komponen angsuran = Pokok + Bunga/Margin + Premi), Annual (dibayar tahunan oleh koperasi).

16. **Alur Klaim — Empat Langkah.** (1) Submission hari ke-1 + upload dokumen. (2) Document verification H+1 s/d H+7. (3) Assessment H+7 s/d H+14 (hitung coverage, approval Manager). (4) Settlement H+14 s/d H+30 (offset pinjaman atau bayar ke ahli waris).

17. **Settlement Type.** Klaim dapat diselesaikan dengan: pay_to_koperasi (langsung lunasi pinjaman), pay_to_beneficiary (bayar ke ahli waris), atau offset_loan (kombinasi).

18. **Integrasi dengan Lifecycle Kematian.** Saat anggota meninggal (K030), sistem cek apakah ada polis asuransi jiwa/takaful aktif. Jika ada, proses klaim asuransi sebelum membebankan sisa pinjaman ke ahli waris.

---

## Key Decisions & Rationale

### D1. Dana Cadangan sebagai Ekuitas (K034)

**Keputusan:** Dana cadangan dicatat sebagai ekuitas koperasi, bukan kewajiban.

**Alasan:** UU Koperasi Pasal 45 menetapkan dana cadangan sebagai bagian dari modal sendiri koperasi. Ini berbeda dari provisi (PPAP) yang merupakan beban/counter-asset.

### D2. Investment Reserve dengan Instrumen Terbatas (K034)

**Keputusan:** Investment reserve hanya boleh diinvestasikan di deposito, SBN, dan reksa dana pasar uang/fixed income.

**Alasan:** Saham dan instrumen berisiko tinggi tidak sesuai dengan prinsip kehati-hatian koperasi simpan pinjam. Dana cadangan adalah jaring pengaman — bukan portofolio investasi agresif.

### D3. Aging Bucket 8 Level, Bukan 5 (K033)

**Keputusan:** Delapan aging bucket, lebih granular dari klasifikasi OJK (5 level).

**Alasan:** Granularitas lebih tinggi memungkinkan strategi penagihan yang lebih tepat sasaran. DPW 1-7 berbeda treatment dari DPW 8-30 meskipun keduanya masuk kategori "Dalam Perhatian Khusus" di OJK.

### D4. Takaful untuk Mode BMT (K035)

**Keputusan:** Untuk `coop_type = "islamic"`, asuransi konvensional digantikan takaful (asuransi syariah).

**Alasan:** Prinsip syariah melarang gharar (ketidakpastian berlebih) yang ada dalam asuransi konvensional. Takaful menggunakan mekanisme tabarru' (donasi) ke pool bersama, bukan premi kepada perusahaan asuransi.

---

## Alur Penagihan Bertingkat (Collection Workflow)

```
Angsuran Jatuh Tempo → Tidak Dibayar
         │
         v
DPD 1-7: Auto reminder WA/SMS
         │ Masih tidak bayar
         v
DPD 8-30: Telepon Collector L1 → Notifikasi penjamin
         │
         ├──→ Nasabah bayar → RESOLVED
         │
         v
DPD 31-60: SP1 + Home Visit → Tawaran Restructuring
         │ Supervisor Collection assigned
         │
         ├──→ Nasabah setuju restructuring → RESTRUCTURED
         │
         v
DPD 61-90: SP2 + Multiple Visits + Meeting penjamin
         │ Manager assigned
         │
         v
DPD 91-180: SP3 (Final) + Proses agunan + Persiapan legal
         │
         ├──→ Jaminan dieksekusi → COLLATERAL_LIQUIDATION
         │
         v
DPD 181-270: Legal Notice + Mediasi + Eksekusi jaminan
         │
         v
DPD 271-360: Write-Off Proposal → Manager → Bendahara → Ketua
         │
         v
DPD > 360: Write-Off → Blacklist → Laporan OJK (NPL)
```

---

## Alur Klaim Asuransi/Takaful

```
Event: Nasabah meninggal / cacat / kehilangan pekerjaan
         │
         v
┌────────────────────────────────────────┐
│ Step 1: Cek Polis Aktif                │
│ Apakah ada insurance_policy aktif      │
│ untuk pinjaman ini?                    │
└───────────────┬────────────────────────┘
                │ Ada polis
                v
┌────────────────────────────────────────┐
│ Step 2: Submission (Hari ke-1)         │
│ Upload dokumen (surat kematian/dokter) │
│ status → submitted                     │
└───────────────┬────────────────────────┘
                │
                v
┌────────────────────────────────────────┐
│ Step 3: Verifikasi Dokumen (H+1-H+7)  │
│ Staff cek kelengkapan                  │
│ Request dokumen tambahan jika kurang   │
│ status → under_review                  │
└───────────────┬────────────────────────┘
                │
                v
┌────────────────────────────────────────┐
│ Step 4: Assessment (H+7-H+14)         │
│ Hitung coverage amount                 │
│ Tentukan settlement type               │
│ Approval Manager                       │
└───────────────┬────────────────────────┘
                │
                v
┌────────────────────────────────────────┐
│ Step 5: Settlement (H+14-H+30)        │
│ Offset pinjaman (pay_to_koperasi) ATAU │
│ Bayar ke ahli waris (pay_to_beneficiary│
│ Jurnal penyelesaian klaim              │
│ status → paid                          │
└────────────────────────────────────────┘
```

---

## Integration Points

### Dana Cadangan

```
K034 (Reserve Fund)
│
├── K016 (SHU)          — ShuApprovedEvent → trigger alokasi otomatis ke dana cadangan
├── K015 (Jurnal/COA)   — Setiap transaksi dana cadangan menghasilkan jurnal entry
├── K017 (Laporan Reg)  — Saldo dan transaksi dana cadangan masuk laporan regulasi
├── K026 (RAT)          — RAT menyetujui SHU dan distribusi; Special reserve dibentuk/dicairkan via RAT
└── K031 (Health)       — Dana cadangan masuk perhitungan CAR (Capital Adequacy Ratio)
```

### Penagihan

```
K033 (Collection)
│
├── K007 (Pinjaman)     — Collection case dibuat dari pinjaman yang DPD > 0
├── K008 (Angsuran)     — DPD dihitung dari jadwal angsuran
├── K009 (Denda)        — Denda terus berjalan selama collection (konvensional)
├── K010 (Agunan)       — Agunan diproses saat DPD 91+ hari
├── K022 (Notifikasi)   — Auto reminder, SP letter notification, escalation alert
├── K031 (Health)       — Collection efficiency, NPL rate masuk KPI kesehatan
├── K035 (Asuransi)     — Cek polis saat write-off proposal (ada coverage yang bisa diklaim?)
└── K036 (Mobile App)   — Collector App untuk field collection dengan GPS + foto
```

### Asuransi

```
K035 (Asuransi/Takaful)
│
├── K007 (Pinjaman)     — Polis terbit saat disbursement; klaim terhubung ke pinjaman
├── K030 (Lifecycle)    — Death settlement cek polis asuransi jiwa sebelum bebankan ke ahli waris
├── K008 (Angsuran)     — Komponen premi masuk jadwal angsuran (monthly deducted)
├── K033 (Collection)   — Saat write-off: cek apakah ada polis yang bisa diklaim
└── K015 (Jurnal)       — Premium payment dan claim settlement menghasilkan jurnal entry
```

---

## RBAC Summary

### Dana Cadangan

| Permission | Bendahara | Ketua | Pengawas | Admin |
|------------|-----------|-------|----------|-------|
| View reserve fund balance | v | v | v | v |
| Request withdrawal | v | v | - | v |
| Approve withdrawal (statutory) | v | v | v (witness) | v |
| Approve withdrawal (general) | v | v | - | v |
| Investment placement | v | v | - | v |
| Configure allocation percentage | - | - | - | v |

### Penagihan

| Permission | Collector L1 | Collector L2 | Supervisor | Manager | Admin |
|------------|-------------|-------------|------------|---------|-------|
| View own cases | v | v | v | v | v |
| View all cases | - | - | v | v | v |
| Log collection activity | v | v | v | v | v |
| Issue SP letter | - | v | v | v | v |
| Offer restructuring | - | - | v | v | v |
| Approve restructuring | - | - | - | v | v |
| Initiate legal action | - | - | - | v | v |
| Propose write-off | - | - | - | v | v |
| Approve write-off | - | - | - | - | v |

### Asuransi

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View policy details | v | v | v | v |
| Issue policy | - | v | v | v |
| Submit claim | v | v | v | v |
| Process claim | - | v | v | v |
| Approve claim | - | - | v | v |
| Configure products | - | - | - | v |

---

## Dual-Mode Perbedaan

| Aspek | Koperasi Umum | BMT (Islamic) |
|-------|---------------|---------------|
| **Dana Cadangan** — Instrumen investasi | Deposito, SBN, Reksa Dana | Deposito Syariah, Sukuk |
| **Dana Cadangan** — Return | Pendapatan bunga | Bagi hasil → income |
| **Dana Cadangan** — Loss coverage | Standard | + DPS harus verifikasi |
| **Dana Cadangan** — Withdrawal approval | Ketua + Bendahara + Pengawas | + DPS untuk withdrawal besar |
| **Dana Cadangan** — Social fund | Tidak ada | Dana sosial (baitul maal) terpisah |
| **Penagihan** — Pendekatan | Standard | + pendekatan kekeluargaan/nasihat |
| **Penagihan** — Denda selama collection | Bunga terus berjalan | Ta'zir nominal tetap (bukan compound) |
| **Penagihan** — Restructuring | Semua jenis | Harus sesuai prinsip syariah, DPS approve |
| **Penagihan** — Write-off | Standard | + izin DPS |
| **Penagihan** — Collection fee | Dibebankan ke nasabah | Ta'widh (harus ada bukti biaya nyata) |
| **Asuransi** — Produk | Asuransi konvensional | Takaful (asuransi syariah) |
| **Asuransi** — Premium | Premi fixed | Kontribusi tabarru' (donasi ke pool) |
| **Asuransi** — Provider | Asuransi umum | Perusahaan asuransi syariah |
| **Asuransi** — DPS oversight | Tidak perlu | DPS harus approve provider dan produk |
| **Asuransi** — Surplus | Tidak ada | Dapat dibagikan kembali ke peserta |
