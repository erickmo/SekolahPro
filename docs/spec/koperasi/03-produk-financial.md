# Produk Keuangan (Akad, Simpanan Pokok/Wajib, Tabungan, Deposito)

Produk adalah **fondasi** sistem koperasi — mendefinisikan aturan bisnis (rate, tenor, biaya, limit) yang berlaku untuk setiap rekening. Tanpa katalog produk yang terstruktur, konfigurasi bisnis tersebar dan sulit di-maintain. Semua ADR pinjaman (K007) dan rekening (K002) bergantung pada entitas produk ini.

---

## ADR References

- **ADR-K003** — Produk & Akad
- **ADR-K004** — Simpanan Pokok & Wajib
- **ADR-K005** — Tabungan (Simpanan Sukarela)
- **ADR-K006** — Deposito / Simpanan Berjangka

---

## Domain Entities

### Tabel `produk`

| Field | Tipe | Keterangan |
|-------|------|------------|
| id | UUID v7 (PK) | |
| tenant_id | UUID (FK) | |
| code | VARCHAR UNIQUE/tenant | Kode produk, e.g., `TB-BERKAH` |
| name | VARCHAR NOT NULL | |
| category | ENUM | simpanan_pokok, simpanan_wajib, tabungan, deposito, pinjaman |
| coop_type | ENUM NOT NULL | general, islamic |
| interest_config | JSONB nullable | Konfigurasi bunga (general mode) |
| akad_type | ENUM nullable | wadiah, mudharabah, murabahah, musyarakah, ijarah, qardh |
| islamic_config | JSONB nullable | Konfigurasi syariah (nisbah, margin, ujrah) |
| rate_type | ENUM | fixed, floating |
| min_balance | NUMERIC(15,2) | |
| max_balance | NUMERIC(15,2) nullable | null = unlimited |
| min_deposit | NUMERIC(15,2) | |
| max_deposit_per_trx | NUMERIC(15,2) nullable | |
| min_withdrawal | NUMERIC(15,2) | |
| max_withdrawal_daily | NUMERIC(15,2) nullable | |
| max_withdrawal_monthly | NUMERIC(15,2) nullable | |
| min_loan_amount / max_loan_amount | NUMERIC nullable | Khusus pinjaman |
| min_tenor_months / max_tenor_months | INTEGER nullable | Khusus pinjaman/deposito |
| min_placement | NUMERIC(15,2) nullable | Khusus deposito |
| tenor_options | JSONB nullable | Array tenor: `[1,3,6,12,24]` |
| fee_config | JSONB nullable | Opening fee, admin fee, closing fee, penalti |
| eligible_relations | JSONB NOT NULL | Array school_relation_type yang boleh akses |
| status | ENUM | draft, active, discontinued |
| version | INTEGER NOT NULL | Increment setiap perubahan konfigurasi |
| effective_since | DATE NOT NULL | |
| is_default | BOOLEAN | true = produk default tenant untuk kategori ini |
| _rels / _data | JSONB | Vernon pattern |

### Tabel `produk_version_history`

| Field | Tipe | Keterangan |
|-------|------|------------|
| produk_id | UUID (FK) | |
| version | INTEGER NOT NULL | |
| interest_config | JSONB | Snapshot konfigurasi bunga |
| islamic_config | JSONB | Snapshot konfigurasi syariah |
| fee_config | JSONB | Snapshot biaya |
| effective_since | DATE NOT NULL | |
| superseded_at | DATE nullable | null = versi terkini |

### Tabel `simpanan_wajib_billing`

| Field | Tipe | Keterangan |
|-------|------|------------|
| nasabah_id | UUID (FK) | |
| rekening_id | UUID (FK) | Rekening simpanan_wajib |
| period | VARCHAR | Format: `YYYY-MM` |
| amount | NUMERIC(15,2) NOT NULL | |
| due_date | DATE NOT NULL | |
| status | ENUM | unpaid, paid, partial, waived |
| paid_amount | NUMERIC(15,2) | |
| paid_via | ENUM nullable | teller, auto_debit, payroll, batch, transfer |
| paid_by_nasabah_id | UUID nullable | Jika dibayarkan orang lain (orang tua bayar untuk anak) |
| is_overdue | BOOLEAN | |
| overdue_since | DATE nullable | |

### Tabel `auto_debit_config`

| Field | Tipe | Keterangan |
|-------|------|------------|
| nasabah_id | UUID (FK) | |
| target_rekening_id | UUID (FK) | Rekening tujuan (simpanan_wajib) |
| source_rekening_id | UUID (FK) | Rekening tabungan sumber |
| enabled | BOOLEAN | |
| priority | INTEGER | 1 = tertinggi |
| retry_on_failure | BOOLEAN | |
| max_retry | INTEGER | Default: 3 |

### Tabel `tabungan_statement`

| Field | Tipe | Keterangan |
|-------|------|------------|
| rekening_id | UUID (FK) | |
| period_start / period_end | DATE | |
| opening_balance | NUMERIC(15,2) | |
| closing_balance | NUMERIC(15,2) | |
| total_deposits | NUMERIC(15,2) | |
| total_withdrawals | NUMERIC(15,2) | |
| total_interest | NUMERIC(15,2) | Bunga/bagi hasil |
| total_fees | NUMERIC(15,2) | |
| transaction_count | INTEGER | |
| document_url | VARCHAR | URL ke PDF |

### Tabel `deposito_interest_schedule`

| Field | Tipe | Keterangan |
|-------|------|------------|
| rekening_id | UUID (FK) | Rekening deposito |
| period | VARCHAR | Format: `YYYY-MM` |
| posting_date | DATE NOT NULL | |
| status | ENUM | scheduled, posted, skipped |
| principal_amount | NUMERIC(15,2) | |
| rate_applied | NUMERIC(7,4) | Rate yang digunakan |
| gross_interest | NUMERIC(15,2) | |
| tax_amount | NUMERIC(15,2) | PPh |
| net_interest | NUMERIC(15,2) | |
| effective_rate | NUMERIC(7,4) nullable | EAR untuk metode capitalize |
| paid_to_rekening_id | UUID nullable | Rekening tabungan tujuan (jika monthly) |

### Tabel `deposito_bilyet`

| Field | Tipe | Keterangan |
|-------|------|------------|
| rekening_id | UUID (FK) | |
| certificate_number | VARCHAR UNIQUE/tenant | Format: `BYT-YYYY-BRANCH-NNNNNN` |
| status | ENUM | active, matured, cancelled, replaced |
| placement_amount | NUMERIC(15,2) | |
| tenor_months | INTEGER | |
| placement_date / maturity_date | DATE | |
| rate_or_nisbah | VARCHAR | Display: `8% p.a.` atau `Nisbah 45:55` |
| rollover_instruction | VARCHAR | |
| document_url | VARCHAR | URL ke PDF bilyet |
| signed_by | UUID (FK) | Manager penandatangan |
| replaced_by_id | UUID nullable | FK ke bilyet pengganti saat rollover |

---

## Business Rules

### Produk & Akad (K003)

1. Satu produk **tepat satu** kategori rekening
2. Satu kategori bisa memiliki **banyak produk** (kecuali SIMPANAN_POKOK dan SIMPANAN_WAJIB biasanya 1)
3. Produk BMT **wajib** memiliki `akad_type` — tidak boleh null
4. Produk general **tidak boleh** memiliki `akad_type`
5. Mapping akad ke kategori di-enforce di application layer — Murabahah tidak bisa di-assign ke produk tabungan
6. Satu tenant hanya bisa memiliki satu `coop_type` — produk harus sesuai mode tenant
7. Produk yang di-discontinue **tidak bisa** dipakai untuk rekening baru — rekening existing tetap aktif
8. Produk DRAFT bisa dihapus hard; produk ACTIVE/DISCONTINUED tidak bisa dihapus
9. Edit produk ACTIVE memicu **version baru** — rekening existing tetap pakai rate versi saat buka
10. `rate_type = FIXED`: rate dikunci saat buka rekening; `FLOATING`: mengikuti versi terbaru
11. Setiap perubahan rate/config menghasilkan **version baru** yang disimpan di `produk_version_history`
12. Deposito: rate **locked** saat placement; tabungan: rate floating (ikut produk terbaru)
13. Eligibility produk di-cek saat pengajuan rekening — jika tidak eligible, pengajuan ditolak otomatis
14. Saat tenant onboarding, 2 produk default otomatis di-create: Simpanan Pokok dan Simpanan Wajib
15. Produk default tidak bisa dihapus — hanya bisa di-discontinue dengan menentukan pengganti dulu

### Simpanan Pokok (K004)

16. Simpanan Pokok adalah setoran **satu kali** saat join — nominal tetap per tenant sesuai AD/ART
17. **Tidak bisa ditarik** selama nasabah ACTIVE — hanya ditarik saat keluar
18. Rekening otomatis dibuat saat nasabah di-approve; setoran adalah transaksi terpisah
19. Jika AD/ART mengubah nominal, anggota existing bayar selisih (batch top-up oleh Manager+)

#### Simpanan Pokok — Batch Top-Up saat AD/ART Berubah

**Skenario:** AD/ART koperasi diubah, nominal simpanan pokok berubah dari Rp X menjadi Rp Y (Y > X).

```
Langkah:
1. Admin mencatat perubahan AD/ART di sistem (dokumen + tanggal efektif)
2. Manager membuat batch transaksi top-up untuk semua anggota existing
   → Nominal per anggota = Y - X (selisih)
   → Tipe transaksi: SIMPANAN_POKOK_TOPUP
3. Batch memerlukan approval Admin sebelum dieksekusi
4. Setelah disetujui: sistem generate tagihan individual per anggota
5. Anggota membayar via channel normal (teller, auto-debit, payroll)
6. Anggota yang bergabung setelah tanggal efektif langsung bayar nominal Y
```

**Ketentuan:**
- Top-up bersifat **one-time** — hanya satu kali per perubahan AD/ART.
- Anggota **tidak bisa ditolak keanggotaannya** karena belum top-up, tapi statusnya di-flag sebagai `simpanan_pokok_incomplete` sampai lunas.
- Anggota dengan status `simpanan_pokok_incomplete` tidak boleh mengajukan pinjaman baru.
- Transaksi top-up masuk ke jurnal: Debit 1101 Kas Teller → Kredit 3100 Simpanan Pokok.
20. Refund simpanan pokok saat keluar membutuhkan approval **Manager+**
21. Simpanan pokok menjadi basis **Jasa Modal** dalam perhitungan SHU

### Simpanan Wajib (K004)

22. Simpanan Wajib adalah setoran **bulanan** wajib selama keanggotaan
23. Tagihan di-generate otomatis oleh scheduled job di awal bulan untuk semua nasabah aktif
24. Nominal bisa berbeda per `school_relation_type` — dikonfigurasi per tenant
25. **Tidak ada denda finansial** atas keterlambatan simpanan wajib — hanya sanksi administratif
26. Grace period default: 7 hari — keterlambatan di atas grace period masuk kategori tunggakan
27. Channel pembayaran: teller cash, auto-debit tabungan, payroll deduction, wali kelas batch, transfer bank
28. **Wali kelas** bisa kumpulkan dari siswa dan setor batch — Teller proses sebagai 1 transaksi per siswa

#### Simpanan Wajib — Alur Koleksi Wali Kelas (Detail)

```
Minggu 1-4: Wali kelas mengumpulkan simpanan wajib dari setiap siswa di kelasnya
            ↓
Akhir bulan: Wali kelas menyetor batch ke koperasi (kas teller)
            ↓
Teller memvalidasi: total_batch_disetor = SUM(nominal per siswa dalam daftar)
            ↓
       ┌────────────────────┬─────────────────────────────┐
       │ Total cocok        │ Total selisih               │
       ↓                    ↓                             │
Batch diterima         Batch DITOLAK                      │
       ↓                    ↓                             │
Generate transaksi     Wali kelas wajib                   │
per siswa dari         rekonsiliasi ulang                 │
batch yang valid       (cek daftar vs uang fisik)         │
                                                          │
Jika ada siswa tidak setor minggu tertentu:               │
→ Tetap tercatat sebagai tunggakan individual             │
  di simpanan_wajib_billing (status = 'unpaid')           │
→ Tidak memblokir batch bulan berikutnya                  │
```

**Aturan Validasi Batch Wali Kelas:**
- Teller membandingkan total setoran fisik dengan daftar siswa yang dikirimkan wali kelas.
- Jika total tidak cocok: teller **menolak seluruh batch** — tidak ada transaksi parsial.
- Wali kelas harus rekonsiliasi dan menyetor ulang di hari kerja berikutnya.
- Siswa yang tidak menyetor pada minggu tertentu: dicatat sebagai `simpanan_wajib_billing.status = 'unpaid'` untuk periode minggu tersebut — ini adalah tunggakan individual, bukan kesalahan wali kelas.
- Setelah batch valid diterima: sistem generate **1 transaksi per siswa** dari batch, dengan referensi ke batch header (`batch_transaction_id`).
29. Status `waived` hanya oleh **Manager+** dengan alasan — untuk kasus darurat
30. Simpanan wajib menjadi basis **Jasa Modal** dalam perhitungan SHU

### Tabungan (K005)

31. Satu nasabah bisa punya **multiple rekening tabungan** — bahkan beberapa dari produk yang sama
32. Setoran ke rekening FROZEN **diizinkan** — dana masuk tapi tidak bisa ditarik
33. Penarikan ditolak jika saldo setelah penarikan < `min_balance` produk
34. Daily/monthly limit dihitung dari **akumulasi** penarikan hari/bulan ini
35. Tabungan berencana (goal-based): penalty early withdrawal jika tarik sebelum target
36. Auto-debit simpanan wajib diprioritaskan di atas auto-debit tabungan berencana
37. Bunga/bagi hasil tabungan diposting bulanan sebagai transaksi tipe `INTEREST_CREDIT` / `PROFIT_SHARE_CREDIT`

### Kalkulasi Bunga Tabungan — General Mode (Daily Balance)

```
Bunga harian = Saldo harian × (rate p.a. / 100) × (1/365)
Bunga periode = SUM(bunga harian) selama bulan tersebut
PPh = Bunga × tax_rate (default 10%, threshold bebas PPh: Rp 240.000/tahun)
Bunga bersih = Bunga bruto - PPh
```

**Contoh:**
```
Saldo harian bulan Januari: Rp 5.000.000
Rate: 3% p.a.
Bunga (31 hari) = 5.000.000 × (3/100) × (31/365) = Rp 12.740
PPh 10% = Rp 1.274
Bunga bersih = Rp 11.466
```

### Kalkulasi Bagi Hasil Tabungan — Islamic Mode (Mudharabah)

```
1. Rata-rata saldo harian nasabah bulan ini
2. Porsi nasabah = avg_saldo_nasabah / total_avg_saldo_semua_mudharabah
3. Profit pool bulan ini (dari penyaluran pembiayaan)
4. Bagian nasabah = profit_pool × nisbah_nasabah%
5. Bagi hasil nasabah = bagian_nasabah × porsi_nasabah
6. PPh = bagi_hasil × tax_rate
```

**Contoh:**
```
Avg saldo nasabah A = Rp 5.500.000
Total avg saldo = Rp 500.000.000
Porsi A = 5.500.000 / 500.000.000 = 1.1%
Profit pool = Rp 25.000.000
Nisbah nasabah = 40%
Bagi hasil A = 25.000.000 × 40% × 1.1% = Rp 110.000
```

**Catatan**: Jika profit pool negatif, bagi hasil = Rp 0 (tidak mengurangi saldo pokok)

**Aturan Mudharabah — Negatif Profit 3 Bulan Berturut-turut:**

- Jika laporan profit nasabah (untuk produk pembiayaan Mudharabah) menunjukkan **negatif selama 3 bulan berturut-turut**, sistem otomatis men-flag pinjaman tersebut untuk review Supervisor.
- Nasabah **tidak dikenakan bagi hasil** pada periode rugi (prinsip Mudharabah: koperasi menanggung kerugian selama bukan akibat kelalaian/kecurangan nasabah).
- Supervisor wajib memverifikasi laporan profit dan menentukan apakah kerugian disebabkan oleh kondisi bisnis (ditanggung koperasi) atau kelalaian nasabah (nasabah tetap wajib membayar pokok).
- Flag dicatat di `pinjaman._data.negative_profit_flag = true` dan `pinjaman._data.consecutive_negative_months = N`.

### Deposito (K006)

38. Nasabah memilih tenor saat pengajuan: 1, 3, 6, 12, atau 24 bulan
39. Maturity date dihitung: `placement_date + tenor_months` — edge case akhir bulan: gunakan hari terakhir bulan target
40. Rate **locked** saat placement berdasarkan tier (tenor × nominal)
41. Bunga/bagi hasil bisa dibayar bulanan, saat jatuh tempo, atau di-capitalize (compound)
42. Rollover instruction (NONE/PRINCIPAL_ONLY/PRINCIPAL_AND_PROFIT/TO_TABUNGAN) dipilih saat pengajuan, bisa diubah sebelum jatuh tempo
43. Auto-rollover menggunakan **rate terbaru** dari produk (bukan rate deposito lama)
44. Deposito yang di-hold sebagai collateral **tidak bisa** early withdrawal — release hold dulu
45. Deposito yang jatuh tempo saat masih di-hold → auto-rollover PRINCIPAL_AND_PROFIT (override instruction)
46. **Partial withdrawal tidak diizinkan** — harus break seluruh deposito
47. Early withdrawal membutuhkan approval **Manager+** — ada cooling-off period (default 3 hari kerja)

### Kalkulasi Bunga Deposito — General Mode

```
Metode AT_MATURITY:
Bunga = principal × rate × (tenor_months / 12)
PPh = Bunga × tax_rate
Bunga bersih = Bunga - PPh

Metode CAPITALIZE (compound):
EAR = (1 + r/n)^n - 1  (n = 12 untuk compounding bulanan)
Total pokok + bunga = principal × (1 + EAR)^tenor_years
```

**Contoh (at maturity):**
```
Deposito: Rp 50.000.000, Tenor: 12 bulan, Rate: 8% p.a.
Bunga = 50.000.000 × 8% × (12/12) = Rp 4.000.000
PPh 10% = Rp 400.000
Bunga bersih = Rp 3.600.000
```

### Rate Tiers Deposito (Tiered Structure)

| Tenor | < 10 jt | 10 - 50 jt | > 50 jt |
|-------|---------|------------|---------|
| 1 bulan | 4.0% | 4.5% | 5.0% |
| 3 bulan | 5.0% | 5.5% | 6.0% |
| 6 bulan | 6.0% | 6.5% | 7.0% |
| 12 bulan | 7.0% | 7.5% | 8.0% |
| 24 bulan | 8.0% | 8.5% | 9.0% |

*Nilai di atas adalah contoh — dikonfigurasi per tenant via JSONB `rate_tiers`*

### Penalti Early Withdrawal Deposito

**General mode:**
```
Penalti = earned_interest × penalty_rate (default 50%)
Net payout = pokok + bunga earned - penalti
```

**Islamic mode:**
Tidak ada penalti; bagi hasil bulan berjalan **di-forfeit** (hangus)

### Deposito — Tiga Metode Pembayaran Bunga/Bagi Hasil

Nasabah memilih metode pembayaran saat pembukaan deposito. Metode ini tersimpan di `produk.interest_config.payment_method` (ENUM).

| Metode (`payment_method`) | Deskripsi | Keterangan |
|---------------------------|-----------|------------|
| `monthly` | Bunga/bagi hasil dikreditkan ke rekening tabungan nasabah setiap bulan | Memerlukan rekening tabungan aktif di koperasi yang sama |
| `at_maturity` | Bunga/bagi hasil dibayar saat jatuh tempo, digabung dengan pokok | Default jika nasabah tidak memilih |
| `capitalize` | Bunga/bagi hasil ditambahkan ke pokok deposito setiap bulan (compound interest) | EAR formula berlaku; pokok efektif bertambah setiap bulan |

**Aturan Bisnis — Fallback Metode `monthly`:**

- Saat pencairan bunga bulanan, sistem memeriksa apakah nasabah memiliki rekening tabungan aktif di tenant yang sama.
- Jika rekening tabungan **tidak ditemukan** atau **berstatus FROZEN/CLOSED**: sistem otomatis fallback ke metode `capitalize` untuk periode tersebut.
- Fallback dicatat di `deposito_interest_schedule.status = 'fallback_capitalize'` dan notifikasi dikirim ke nasabah.
- Nasabah tidak bisa memilih `monthly` jika tidak punya rekening tabungan saat pengajuan — validasi dilakukan di application layer.

**Kalkulasi EAR untuk Metode `capitalize`:**

```
EAR = (1 + r/12)^12 - 1   (compounding bulanan)
Bulan ke-N: Pokok efektif = placement_amount × (1 + r/12)^N
Bunga bulan ke-N = Pokok_efektif_(N-1) × (r/12)
```

Field `deposito_interest_schedule.effective_rate` diisi dengan EAR yang dihitung saat pokok berubah setiap bulan.

### Deposito Guru & Staf — Bonus Rate

Nasabah dengan `school_relation_type IN ('guru', 'staf', 'karyawan')` berhak mendapat bonus rate di atas rate tier standar, berdasarkan masa kerja.

**Tabel Bonus Rate Berdasarkan Masa Kerja:**

| Masa Kerja | Bonus Rate (p.a.) |
|------------|------------------|
| < 5 tahun | +0.00% (tidak ada bonus) |
| > 5 tahun | +0.25% |
| > 10 tahun | +0.50% |
| > 20 tahun | +0.75% |

Bonus rate bersifat **kumulatif dengan rate tier** (bukan menggantikan). Contoh: tier 12 bulan Rp 10-50 jt = 7.5%, masa kerja 12 tahun → rate efektif = 7.5% + 0.50% = 8.00%.

**Sumber Data:**
- Field `employment_tenure_years` dibaca dari modul HR (`hr_employee.tenure_years`) saat pengajuan deposito.
- Jika data HR tidak tersedia, bonus rate = 0% (tidak diblokir, hanya tidak ada bonus).
- Masa kerja di-snapshot ke `deposito_bilyet._data.employment_tenure_years` saat placement — tidak berubah meski masa kerja bertambah selama deposito berjalan.

**Payroll Auto-Placement:**
- Guru/staf bisa mengaktifkan fitur `payroll_auto_placement`: potongan gaji bulanan otomatis membuka deposito baru atau menambah nominal deposito yang sedang berjalan.
- Konfigurasi di tabel `payroll_deduction_config` (K020): `deduction_type = 'DEPOSITO_AUTO_PLACEMENT'`, `target_product_id`, `amount`.
- Setiap potongan gaji yang diarahkan ke deposito menghasilkan transaksi `DEPOSITO_PLACEMENT` dari rekening virtual payroll ke rekening deposito baru.

---

## Compliance Notes Syariah

| Akad | Persyaratan Syariah |
|------|---------------------|
| Wadiah Yad Dhamanah | Bonus bersifat sukarela (athaya), bukan bunga wajib |
| Mudharabah | Bagi hasil dari profit riil; disclaimer indikatif wajib di formulir; jika rugi berturut-turut 3 bulan, pinjaman otomatis di-flag untuk review Supervisor |
| Murabahah | Total harga jual ditetapkan di awal dan tidak berubah; bukti pembelian barang wajib |
| Musyarakah | Profit dan loss sharing sesuai nisbah; laporan profit nasabah wajib di-submit berkala |
| Ijarah | Ujrah ditetapkan dari biaya riil sewa/penyimpanan; bukan tambahan margin |
| Qardh | Hanya biaya admin (max configurable); tidak ada margin |
| Ta'zir | Nominal tetap (BUKAN persentase); wajib masuk dana sosial bukan pendapatan koperasi |

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Katalog produk dinamis per tenant | Setiap koperasi punya produk berbeda; dinamis tanpa deploy kode baru |
| Versioning produk | Perubahan rate tidak merusak rekening existing; deposito fixed rate dijaga sesuai kontrak |
| JSONB untuk interest_config dan islamic_config | Fleksibel untuk berbagai struktur akad; schema di-validate di application layer |
| Rate tiers untuk deposito | Mendorong penempatan nominal besar dan tenor panjang; standar industri |
| Dormant fee dari produk, bukan tenant config | Fee berbeda per jenis tabungan (reguler vs pendidikan vs qurban) |
| Bilyet deposito PDF | Regulasi koperasi dan kebiasaan nasabah; bukti legal jika ada sengketa |
| Tidak ada partial withdrawal deposito | Memperumit perhitungan rate tier dan manajemen bilyet; break + open new lebih clean |

---

## Integration Points

| Integrasi | Arah | Keterangan |
|-----------|------|------------|
| `rekening` (K002) | Produk → Rekening | Setiap rekening terikat pada satu produk |
| `transaksi` (K011) | Produk → Transaksi | Product-level limits di-enforce di transaction layer |
| `jurnal` (K015) | Produk → Jurnal | Tipe transaksi menentukan mapping jurnal (auto-journal) |
| `shu_anggota` (K016) | Simpanan → SHU | Saldo simpanan menjadi basis Jasa Modal SHU |
| `denda` (K009) | Produk → Denda | Fee config produk (ta'zir/penalty) menentukan denda |
| `angsuran` (K008) | Produk → Angsuran | Rate/akad produk menentukan metode kalkulasi jadwal |
