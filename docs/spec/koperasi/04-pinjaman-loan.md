# Pinjaman / Pembiayaan

Pinjaman adalah domain paling kompleks dan berisiko tinggi dalam sistem koperasi/BMT. Setiap rupiah yang dicairkan adalah dana anggota yang harus dikelola dengan prudent. Domain ini mencakup aplikasi kredit, analisis, approval bertingkat, pencairan, jadwal angsuran, tracking DPD/NPL, jaminan, denda, restrukturisasi, hingga write-off.

---

## ADR References

- **ADR-K007** — Pinjaman / Pembiayaan
- **ADR-K008** — Angsuran & Jadwal
- **ADR-K009** — Denda & Penalti
- **ADR-K010** — Jaminan / Agunan (Collateral Management)

---

## State Machine — Lifecycle Pinjaman

```
                     ┌─────────────────┐
      Teller submit → │  Status: DRAFT  │ ← Teller bisa edit
                     └────────┬────────┘
                              │ submit
                              v
                     ┌─────────────────┐
                     │ Status: PENDING │ ← Menunggu analisis kredit
                     └────────┬────────┘
                              │ Supervisor analisis
                              v
                     ┌─────────────────────────┐
                     │ Status: ANALYZING       │ ← Credit analysis
                     └────────┬────────────────┘
                              │ Manager/Admin approve
                     ┌────────┴────────┐
                     v                 v
               ┌──────────┐    ┌──────────────┐
               │ APPROVED │    │   REJECTED   │
               └────┬─────┘    └──────────────┘
                    │ Pencairan
                    v
               ┌──────────────────┐
               │ Status: ACTIVE   │ ← Angsuran berjalan
               └────────┬─────────┘
                        │
          ┌─────────────┼──────────────┐
          v             v              v
    ┌──────────┐  ┌──────────────┐  ┌──────────────┐
    │COMPLETED │  │ ACTIVE       │  │ WRITTEN_OFF  │
    │ (lunas) │  │(is_restructured│  │ (hapus buku) │
    └──────────┘  │= true)       │  └──────────────┘
                  └──────────────┘
```

**Catatan**: RESTRUCTURED bukan status — pinjaman tetap ACTIVE dengan flag `is_restructured = true` dan jadwal angsuran baru.

---

## Domain Entities

### Tabel `pinjaman`

| Field | Tipe | Keterangan |
|-------|------|------------|
| id | UUID v7 (PK) | |
| nasabah_id | UUID (FK) | |
| rekening_id | UUID (FK) | Rekening pinjaman (1:1) |
| product_id | UUID (FK) | |
| application_id | UUID (FK) | |
| loan_number | VARCHAR UNIQUE/tenant | Auto-generate |
| loan_amount | NUMERIC(15,2) | Plafon yang disetujui |
| disbursed_amount | NUMERIC(15,2) | Nominal yang dicairkan |
| outstanding_principal | NUMERIC(15,2) | = rekening.balance |
| tenor_months | INTEGER | |
| installment_amount | NUMERIC(15,2) | Angsuran per bulan |
| interest_rate | NUMERIC(7,4) nullable | Rate bunga p.a. (general) |
| calculation_method | ENUM nullable | flat, declining, annuity |
| akad_type | ENUM nullable | murabahah, musyarakah, mudharabah, ijarah, qardh |
| margin_rate / margin_amount | NUMERIC nullable | Murabahah |
| nisbah_nasabah / nisbah_koperasi | NUMERIC nullable | Musyarakah/Mudharabah |
| ujrah_amount | NUMERIC nullable | Ijarah |
| loan_purpose / loan_purpose_category | TEXT / ENUM | konsumtif, produktif, pendidikan, darurat, lainnya |
| grace_period_months / grace_period_type | INTEGER / ENUM | none, principal_only, full |
| first_installment_date / last_installment_date | DATE | |
| installment_day | INTEGER | Tanggal jatuh tempo: 1-28 |
| disbursement_date | DATE nullable | |
| disbursement_method | ENUM | transfer_tabungan, wakalah, direct_supplier |
| collectibility | INTEGER NOT NULL DEFAULT 1 | Kol 1-5 (OJK classification) |
| dpd | INTEGER NOT NULL DEFAULT 0 | Days Past Due |
| status | ENUM | active, completed, written_off |
| is_restructured | BOOLEAN | |
| restructure_count | INTEGER | Jumlah restrukturisasi yang dilakukan |
| has_guarantor / has_collateral | BOOLEAN | |
| _rels / _data | JSONB | Vernon pattern |

### Tabel `pinjaman_application`

| Field | Tipe | Keterangan |
|-------|------|------------|
| nasabah_id | UUID (FK) | |
| product_id | UUID (FK) | |
| requested_amount | NUMERIC(15,2) | |
| requested_tenor_months | INTEGER | |
| loan_purpose | TEXT | |
| monthly_income | NUMERIC nullable | Deklarasi penghasilan |
| credit_score | NUMERIC nullable | Hasil scoring otomatis |
| dti_ratio | NUMERIC nullable | Debt-to-Income ratio |
| max_eligible_amount | NUMERIC nullable | |
| analysis_notes | TEXT nullable | Catatan Supervisor |
| analyzed_by / analyzed_at | UUID / TIMESTAMPTZ | |
| approved_amount | NUMERIC nullable | Bisa berbeda dari requested |
| approved_tenor_months / approved_rate | INTEGER / NUMERIC | |
| status | ENUM | draft, pending, analyzing, approved, rejected |
| pinjaman_id | UUID nullable | Diisi saat approved |

### Tabel `pinjaman_penjamin`

| Field | Tipe | Keterangan |
|-------|------|------------|
| pinjaman_id | UUID (FK) | |
| guarantor_type | ENUM | internal, external |
| nasabah_id | UUID nullable | FK → nasabah jika internal |
| full_name | VARCHAR | |
| identity_number | VARCHAR | |
| relationship | VARCHAR | Hubungan dengan peminjam |
| guarantee_amount | NUMERIC(15,2) | Nominal yang dijaminkan |
| guarantee_letter_url | VARCHAR nullable | Surat pernyataan penjaminan |
| status | ENUM | active, released |

### Tabel `angsuran`

| Field | Tipe | Keterangan |
|-------|------|------------|
| pinjaman_id | UUID (FK) | |
| rekening_id | UUID (FK) | |
| installment_number | INTEGER | 1, 2, 3, ..., N |
| due_date | DATE | |
| principal_amount | NUMERIC(15,2) | Porsi pokok jadwal |
| interest_amount | NUMERIC(15,2) | Porsi bunga (general) |
| margin_amount | NUMERIC(15,2) | Porsi margin (murabahah) |
| profit_share_amount | NUMERIC(15,2) | Bagi hasil (musyarakah) |
| ujrah_amount | NUMERIC(15,2) | Sewa (ijarah) |
| total_amount | NUMERIC(15,2) | Total angsuran jadwal |
| remaining_principal | NUMERIC(15,2) | Sisa pokok setelah angsuran ini |
| paid_principal | NUMERIC(15,2) | Aktual yang dibayar (pokok) |
| paid_interest / paid_margin | NUMERIC(15,2) | |
| paid_penalty | NUMERIC(15,2) | Denda yang dibayar |
| total_paid | NUMERIC(15,2) | |
| payment_status | ENUM | scheduled, partial, paid, overdue, waived, voided |
| paid_date | DATE nullable | |
| paid_via | ENUM | teller_cash, auto_debit, payroll |
| dpd | INTEGER | Days Past Due untuk angsuran ini |
| transaction_id | UUID nullable | Link ke transaksi pembayaran |

### Tabel `angsuran_pembayaran`

| Field | Tipe | Keterangan |
|-------|------|------------|
| angsuran_id | UUID (FK) | |
| pinjaman_id | UUID (FK) | |
| transaction_id | UUID (FK) | |
| payment_date | DATE | |
| payment_amount | NUMERIC(15,2) | Total yang dibayar |
| allocated_principal | NUMERIC(15,2) | |
| allocated_interest / allocated_margin | NUMERIC(15,2) | |
| allocated_penalty | NUMERIC(15,2) | |
| payment_channel | ENUM | teller_cash, auto_debit, payroll |
| overpayment_amount | NUMERIC(15,2) | |
| overpayment_handling | ENUM | next_installment, reduce_principal |

### Tabel `denda`

| Field | Tipe | Keterangan |
|-------|------|------------|
| pinjaman_id | UUID (FK) | |
| angsuran_id | UUID nullable | null untuk early_settlement/withdrawal |
| penalty_type | ENUM | late_payment, early_settlement, early_withdrawal, tazir, tawidh |
| calculation_basis | NUMERIC(15,2) | Nominal dasar perhitungan |
| penalty_rate | NUMERIC nullable | Rate yang digunakan |
| penalty_days | INTEGER | Jumlah hari keterlambatan |
| calculated_amount | NUMERIC(15,2) | Denda yang dihitung |
| final_amount | NUMERIC(15,2) | Setelah cap |
| cap_applied | BOOLEAN | |
| paid_amount | NUMERIC(15,2) | |
| waived_amount | NUMERIC(15,2) | |
| outstanding_amount | NUMERIC(15,2) | = final - paid - waived |
| status | ENUM | accruing, settled, waived, partial_waived |
| fund_destination | ENUM | koperasi_income, social_fund |
| waiver_reason | TEXT nullable | |
| waiver_approved_by | UUID nullable | |

### Tabel `jaminan`

| Field | Tipe | Keterangan |
|-------|------|------------|
| nasabah_id | UUID (FK) | Pemilik jaminan |
| collateral_number | VARCHAR UNIQUE/tenant | `JM-YYYY-BRANCH-NNNNNN` |
| collateral_type | ENUM | internal_balance, surat_berharga, barang_bergerak, personal_guarantee |
| description | TEXT | |
| item_category | VARCHAR nullable | Sub-kategori: kendaraan_roda2, tanah_shm, dll |
| item_serial_number | VARCHAR nullable | No rangka/mesin |
| ownership_document | VARCHAR nullable | No BPKB/sertifikat |
| rekening_id | UUID nullable | Rekening yang di-hold (internal_balance) |
| hold_amount | NUMERIC nullable | Nominal yang di-hold |
| appraised_value | NUMERIC(15,2) | Nilai taksasi |
| acceptance_rate | NUMERIC(5,4) | e.g., 0.7000 = 70% |
| collateral_value | NUMERIC(15,2) | = appraised_value × acceptance_rate |
| appraised_at / appraised_by | TIMESTAMPTZ / UUID | |
| next_revaluation_at | TIMESTAMPTZ nullable | |
| status | ENUM | registered, pledged, released, foreclosed |
| ujrah_monthly | NUMERIC nullable | Biaya pemeliharaan/bulan (Islamic only) |

### Tabel `jaminan_pinjaman` (binding)

| Field | Tipe | Keterangan |
|-------|------|------------|
| jaminan_id | UUID (FK) | |
| pinjaman_id | UUID (FK) | |
| pledged_value | NUMERIC(15,2) | Porsi nilai jaminan untuk pinjaman ini |
| status | ENUM | active, released |
| bound_at | TIMESTAMPTZ | |
| released_at | TIMESTAMPTZ nullable | |

---

## Business Rules

### Credit Analysis & Approval (K007)

1. Teller bisa buat pengajuan tapi **tidak bisa approve sendiri** — Supervisor untuk analisis, Manager/Admin untuk approval
2. Approval limit configurable per tenant:
   - Supervisor: hanya rekomendasi, tidak bisa final approve
   - Manager: bisa approve sampai limit konfigurasi (default: Rp 25 juta)
   - Admin: unlimited
3. Pengajuan di atas limit Manager **otomatis di-escalate** ke Admin
4. **Credit pre-check otomatis** saat submit: nasabah ACTIVE, KYC mencukupi, produk eligible, BMPK, kolektibilitas existing
5. BMPK: total eksposur nasabah ≤ 20% modal sendiri (single); grup ≤ 25%

   > **Catatan Limitasi BMPK — Definisi "Group Borrower" (MVP):**
   > Definisi "grup peminjam" dalam sistem SekolahPro saat ini = **keluarga inti** (pasangan, orang tua, anak kandung) + penjamin yang terdaftar di `pinjaman_penjamin`.
   > Ini adalah **penyederhanaan untuk MVP**. OJK mendefinisikan "pihak terkait" (related party) lebih luas — mencakup entitas dengan pemilik yang sama, direksi/komisaris yang berafiliasi, dll.
   > Akan diperluas di fase berikutnya sesuai pertumbuhan dan kompleksitas koperasi.
6. **Siswa/santri tidak boleh mengajukan pinjaman langsung** — orang tua/wali sebagai peminjam
7. Pinjaman darurat: fast-track approval oleh Manager langsung tanpa Supervisor analysis
8. Rekening pinjaman: `balance = outstanding principal` — berkurang setiap pembayaran angsuran (porsi pokok)

### Pencairan (Disbursement)

9. Pencairan hanya **satu kali** per pinjaman — tidak ada partial disbursement
10. General: dana ditransfer ke rekening tabungan nasabah
11. Murabahah: koperasi beli barang dari supplier (bisa dengan wakalah) — bukti pembelian wajib di-upload
12. Qardh: transfer ke rekening tabungan nasabah

### Angsuran & Jadwal (K008)

13. Jadwal angsuran di-generate **otomatis** saat pinjaman di-approve, sebelum disbursement
14. Jadwal **immutable** setelah disbursement — perubahan hanya via restrukturisasi
15. Pembayaran selalu di-match ke angsuran **paling lama yang belum lunas** (FIFO)
16. **Prioritas alokasi**: denda → bunga/margin tertunggak → pokok tertunggak → bunga/margin berjalan → pokok berjalan → overpayment
17. Concurrency control: `SELECT ... FOR UPDATE` pada row angsuran yang akan dibayar — timeout 5 detik
18. Lock ordering: selalu lock rekening dengan UUID lebih kecil terlebih dahulu (deadlock prevention)
19. Underpayment diterima — status `PARTIAL`, sisa carry forward ke bulan berikutnya
20. DPD dihitung dari **due_date**, bukan payment_date

### Kolektibilitas NPL (K007 §11)

> **Catatan Keselarasan ADR:** Tabel ini menggunakan klasifikasi **POJK 62/POJK.05/2014 untuk LKM (Lembaga Keuangan Mikro)**. Threshold ini lebih konservatif dari minimum regulasi untuk mendukung prudential lending di koperasi sekolah. Referensi ADR-K007 (yang sebelumnya menggunakan threshold berbeda) telah diselaraskan dengan ADR-K033.

| Kol | Nama | DPD | PPAP |
|-----|------|-----|------|
| 1 | Lancar | 0 hari | 0% |
| 2 | Dalam Perhatian Khusus | 1–90 hari | 1% |
| 3 | Kurang Lancar | 91–180 hari | 10% |
| 4 | Diragukan | 181–270 hari | 50% |
| 5 | Macet | > 270 hari | 100% |

**Catatan:** Aging buckets operasional dari ADR-K033 (`DPW_0_30`, `DPW_31_90`, `DPW_91_180`, dll.) adalah **sub-klasifikasi operasional dari Kol-2** untuk keperluan collection follow-up dan reminder otomatis. Mereka tidak menggantikan klasifikasi kolektibilitas di atas — hanya mempersempit granularitas untuk tim collection.

21. Klasifikasi kolektibilitas **otomatis** berdasarkan DPD — Supervisor+ bisa override dengan alasan
22. Downgrade kolektibilitas otomatis; upgrade setelah pembayaran tepat waktu N bulan (default: 3)
23. Notifikasi per perubahan kolektibilitas: Kol 1→2: Supervisor; 2→3: Manager; 3→4: Admin; 4→5: Admin + alert write-off

### Denda & Penalti (K009)

24. Grace period denda default: 3 hari setelah due_date — denda mulai dihitung hari ke-4
25. **Cap denda wajib** (default: 25% dari pokok) — denda berhenti dihitung setelah cap tercapai
26. **General mode**: denda = % dari jumlah tertunggak → masuk pendapatan koperasi
27. **Islamic mode (ta'zir)**: nominal tetap (BUKAN %) → **wajib masuk dana sosial** (bukan pendapatan koperasi)
28. Ta'widh (kompensasi kerugian nyata) berbeda dari ta'zir — membutuhkan bukti + approval Manager+; masuk P&L koperasi
29. Waiver partial: Manager+; waiver full: Admin
30. Saat restrukturisasi: denda bisa di-waive, di-freeze, atau di-carry forward (dicatat di dokumen restrukturisasi)

### Restrukturisasi (K007 §8)

31. Hanya untuk pinjaman dengan kolektibilitas **Kol-2 ke atas**
32. Approval minimal **Manager** — configurable per tenant
33. Jadwal lama di-void (status VOIDED), jadwal baru di-generate dari sisa pokok outstanding
34. Pinjaman tetap ACTIVE dengan flag `is_restructured = true`, `restructure_count++`
35. Maksimal **2x restrukturisasi** per pinjaman (configurable)
36. BMT: restrukturisasi bisa termasuk **konversi akad** (e.g., Murabahah → Qardh)

### Early Settlement (K007 §9)

37. **General mode**: penalti early settlement opsional (configurable per produk)
38. **Islamic mode**: **TIDAK ADA penalti** — BMT wajib memberikan **Ibra'** (diskon sisa margin belum jatuh tempo)
39. Ibra' = sisa margin belum earned × discount rate (default: 100%) — dicatat sebagai pengurang pendapatan margin di jurnal

### Write-Off (K007 §12)

40. Hanya pinjaman **Kol-5 (Macet)** yang bisa di-write-off
41. PPAP harus sudah **100% tercadangkan** sebelum write-off diizinkan
42. Write-off memerlukan approval **Admin** — proposal dari Manager
43. Data pinjaman tidak dihapus — dipindahkan ke tracking off-balance sheet
44. Recovery setelah write-off dicatat sebagai pendapatan lain-lain

### Jaminan Khas Koperasi Sekolah

Selain jenis jaminan umum (internal_balance, surat_berharga, barang_bergerak, personal_guarantee), koperasi sekolah mengenal jenis jaminan unik berbasis konteks pendidikan:

#### Ijazah sebagai Jaminan Moral

- **Nilai taksasi nominal:** Rp 1.000.000 (bukan nilai pasar — ijazah tidak memiliki nilai jual yang dapat dieksekusi secara hukum).
- **Tujuan:** Komitmen moral nasabah, bukan jaminan finansial yang bisa dieksekusi.
- Hanya berlaku untuk **pinjaman darurat kecil** dengan plafon maksimum = **3× simpanan wajib bulanan** nasabah.
- Ijazah disimpan secara fisik di **brankas koperasi** selama pinjaman aktif.
- Saat pinjaman lunas: ijazah dikembalikan ke nasabah dengan **tanda terima yang ditandatangani** kedua pihak.
- Hak koperasi atas ijazah **terbatas pada penyimpanan** — tidak dapat diperjualbelikan, dilelang, atau digunakan sebagai dasar klaim hukum.

**Kolom di tabel `jaminan`:**
| Field | Nilai |
|-------|-------|
| `collateral_type` | `'IJAZAH'` (tambahan ke ENUM existing) |
| `collateral_storage` | `'BRANKAS_KOPERASI'` |
| `appraised_value` | `1000000` (Rp 1.000.000 — nominal tetap) |
| `acceptance_rate` | `0.0000` (nilai moral, bukan finansial) |
| `collateral_value` | `0` (tidak masuk perhitungan coverage ratio) |
| `description` | Nomor ijazah, instansi penerbit, tahun lulus |

**Validasi di application layer:**
- Pinjaman dengan jaminan IJAZAH: plafon maksimum = `nasabah.simpanan_wajib_monthly × 3`.
- Jika nasabah belum punya simpanan wajib aktif: jaminan IJAZAH tidak diizinkan.

### Jaminan & Agunan (K010)

45. Collateral value = `appraised_value × acceptance_rate` (satu haircut — acceptance_rate sudah memperhitungkan LTV)
46. Default acceptance rates: INTERNAL_BALANCE 100%, SURAT_BERHARGA 70-80%, BARANG_BERGERAK 50-70%, PERSONAL_GUARANTEE 0%
47. Coverage ratio minimum = 1.00 (100%); warning threshold = 1.20 (120%)
48. Auto-release hold saat pinjaman **COMPLETED**
49. Internal collateral (saldo): `hold_amount` di rekening — tidak bisa ditarik saat di-pledge
50. Deposito yang di-pledge: early withdrawal diblokir; jika jatuh tempo → auto-rollover PRINCIPAL_AND_PROFIT
51. Revaluasi periodik wajib: SURAT_BERHARGA setiap 12 bulan; BARANG_BERGERAK setiap 6 bulan
52. Foreclosure: hanya untuk pinjaman WRITTEN_OFF; surplus hasil jual **wajib dikembalikan** ke nasabah
53. **Islamic (Rahn)**: koperasi tidak boleh menggunakan/mengambil manfaat dari marhun; ujrah = biaya riil penyimpanan

---

## Kalkulasi Jadwal Angsuran

### General Mode

**A. Flat (Bunga Tetap)**
```
Total Bunga = Pokok × Rate × (Tenor/12)
Angsuran/bln = (Pokok + Total Bunga) / Tenor
Porsi Pokok = Pokok / Tenor (tetap)
Porsi Bunga = Total Bunga / Tenor (tetap)
```

**B. Declining (Bunga Menurun)**
```
Porsi Pokok = Pokok / Tenor (tetap)
Bunga bln-N = Sisa Pokok × (Rate/12)
Total bln-N = Pokok_bln + Bunga_bln (menurun setiap bulan)
```

**C. Anuitas (Annuity)**
```
r = Rate/12
Angsuran = Pokok × r × (1+r)^n / ((1+r)^n - 1)   [tetap setiap bulan]
Bunga bln-N = Sisa Pokok × r
Pokok bln-N = Angsuran - Bunga bln-N
```

### Islamic Mode

**Murabahah** (sama dengan Flat konvensional secara kalkulasi):
```
Total Margin = Harga Pokok × Margin_rate
Harga Jual = Harga Pokok + Total Margin
Angsuran = Harga Jual / Tenor
```

**Musyarakah/Mudharabah**:
```
Porsi Pokok = Pokok / Tenor (tetap)
Bagi Hasil = Projected Profit × Nisbah_BMT
Angsuran bervariasi; review bagi hasil tiap 3 bulan
```

#### Review Triwulanan Bagi Hasil Musyarakah/Mudharabah

Bagi hasil tidak bersifat tetap — wajib di-review setiap 3 bulan berdasarkan **profit aktual** usaha nasabah.

**Alur Review Triwulanan:**
1. Sistem otomatis membuat tagihan `profit_report_submission` di akhir setiap kuartal angsuran.
2. Nasabah wajib menyerahkan **laporan keuangan/profit aktual** (minimal laporan sederhana penerimaan-pengeluaran) ke koperasi.
3. Supervisor memverifikasi laporan sebelum angsuran periode berikutnya dihitung.
4. Angsuran bagi hasil periode berikutnya = `laporan_profit_aktual × nisbah_BMT`.

**Jika laporan tidak disubmit:**
- Angsuran bagi hasil menggunakan **projected profit terakhir** yang tersimpan di `pinjaman._data.last_projected_profit`.
- Angsuran di-flag: `angsuran._data.is_estimated = true`.
- Setelah 2 kuartal berturut-turut tidak submit: pinjaman di-flag untuk review Supervisor (potensi restrukturisasi).

**Jika profit aktual < projected:** angsuran bagi hasil turun (nasabah diuntungkan).
**Jika profit aktual > projected:** angsuran bagi hasil naik (koperasi mendapat porsi sesuai akad).
**Jika profit negatif:** lihat aturan Negatif Profit 3 Bulan di dokumen 03-produk-financial.md.

**Ijarah**:
```
Ujrah/bln = Nilai Aset × ujrah_rate (tetap)
Tidak ada porsi pokok (nasabah bayar sewa)
```

**Qardh**:
```
Angsuran = Pokok / Tenor (tanpa bunga/margin)
```

---

## RBAC Summary

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| Buat pengajuan pinjaman | v | v | v | v |
| Credit analysis (rekomendasi) | - | v | v | v |
| Approve pinjaman (sesuai limit) | - | - | v | v |
| Approve pinjaman (di atas limit Manager) | - | - | - | v |
| Proses pencairan | - | - | v | v |
| Record pembayaran angsuran | v | v | v | v |
| Override payment allocation | - | v | v | v |
| Waive angsuran | - | - | v | v |
| Ajukan restrukturisasi | - | v | v | v |
| Approve restrukturisasi | - | - | v | v |
| Ajukan write-off | - | - | v | v |
| Approve write-off | - | - | - | v |
| Input data jaminan | v | v | v | v |
| Taksasi/valuasi jaminan | - | v | v | v |
| Bind jaminan ke pinjaman | - | v | v | v |
| Release jaminan fisik | - | - | v | v |
| Approve foreclosure | - | - | - | v |
| Waive denda (partial) | - | - | v | v |
| Waive denda (full) | - | - | - | v |

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Multi-level approval (Teller → Supervisor → Manager/Admin) | Separation of duties; fraud prevention; credit analysis butuh keahlian |
| Single disbursement | Cukup untuk koperasi sekolah; simplicity > feature completeness |
| Kolektibilitas otomatis berbasis DPD | Konsistensi standar OJK; manual override tersedia untuk kasus khusus |
| VOIDED jadwal lama saat restrukturisasi | Immutability data historis; jadwal baru dimulai dari sisa pokok |
| Ta'zir ke dana sosial (Islamic) | Wajib per fatwa DSN-MUI No. 17/DSN-MUI/IX/2000 |
| Ibra' wajib untuk early settlement (Islamic) | Fatwa DSN-MUI; tidak ada penalti pada pelunasan dini |
| Collateral value = appraised_value × acceptance_rate (satu haircut) | Menghindari double-haircut yang tidak realistis |
| Ujrah sebagai biaya riil penyimpanan (Rahn) | Prinsip syariah: koperasi tidak boleh mengambil manfaat dari marhun |

---

## Integration Points

| Integrasi | Arah | Keterangan |
|-----------|------|------------|
| `rekening` (K002) | Pinjaman ← Rekening | Rekening PINJAMAN adalah representasi saldo outstanding |
| `produk` (K003) | Pinjaman ← Produk | Rate, tenor, calculation method dari produk |
| `transaksi` (K011) | Pinjaman → Transaksi | Angsuran dan pencairan menghasilkan transaksi |
| `jurnal` (K015) | Pinjaman → Jurnal | Auto-journal untuk angsuran pokok, bunga/margin, pencairan |
| `jaminan` (K010) | Pinjaman ↔ Jaminan | Deposito tabungan bisa di-hold sebagai collateral |
| `payroll_deduction` (K020) | Pinjaman → Payroll | Angsuran bisa dipotong dari gaji guru/staf |
| `zakat` (K018) | Ta'zir → Dana Sosial | Dana ta'zir masuk ke baitul maal, bukan pendapatan koperasi |
| `laporan_regulasi` (K017) | NPL → Laporan | Kolektibilitas menjadi input laporan OJK (kolektibilitas, CAR) |
