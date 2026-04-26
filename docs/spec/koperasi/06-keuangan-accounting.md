# Keuangan & Akuntansi (Kas, Jurnal/COA, SHU, Zakat)

Modul keuangan adalah tulang punggung pelaporan koperasi. Mencakup pengelolaan kas harian, mesin akuntansi double-entry dengan Chart of Accounts lengkap, perhitungan dan distribusi SHU tahunan, serta pengelolaan zakat dan infaq untuk BMT.

---

## ADR References

- **ADR-K014** — Kas & Cash Flow
- **ADR-K015** — Jurnal & Akuntansi (COA)
- **ADR-K016** — SHU (Sisa Hasil Usaha)
- **ADR-K018** — Zakat & Infaq *(Islamic only)*

---

## Domain Entities

### Tabel `kas_harian`

| Field | Tipe | Keterangan |
|-------|------|------------|
| branch_id | UUID (FK) | |
| report_date | DATE | |
| opening_balance | NUMERIC(15,2) | = closing_balance hari sebelumnya |
| closing_balance | NUMERIC(15,2) | |
| total_inflow / total_outflow | NUMERIC(15,2) | |
| inflow_setoran / inflow_angsuran | NUMERIC(15,2) | Breakdown per kategori |
| inflow_operasional / inflow_transfer / inflow_lainnya | NUMERIC(15,2) | |
| outflow_penarikan / outflow_pencairan | NUMERIC(15,2) | |
| outflow_operasional / outflow_transfer / outflow_lainnya | NUMERIC(15,2) | |
| total_teller_sessions | INTEGER | |
| total_teller_variance | NUMERIC(15,2) | |
| is_reconciled | BOOLEAN | |
| reconciled_at / reconciled_by | TIMESTAMPTZ / UUID | |
| status | ENUM | open, closed, reconciled |

**Constraint:** `CHECK (closing_balance = opening_balance + total_inflow - total_outflow)`
**Unique:** `(tenant_id, branch_id, report_date)`

### Tabel `coa` (Chart of Accounts)

| Field | Tipe | Keterangan |
|-------|------|------------|
| account_code | VARCHAR NOT NULL | e.g., `1101` |
| account_name | VARCHAR NOT NULL | |
| parent_id | UUID nullable | Self-reference untuk hirarki |
| level | INTEGER | 1 = top-level |
| account_type | ENUM | asset, liability, equity, revenue, expense, zakat, kebajikan, tazir |
| normal_balance | ENUM | debit, credit |
| is_header | BOOLEAN | true = akun induk, tidak bisa di-posting |
| is_system | BOOLEAN | true = tidak bisa hapus/ubah kode |
| coop_type_required | ENUM | general, islamic, both |

**Unique:** `(tenant_id, account_code)`

### Tabel `jurnal` (Header)

| Field | Tipe | Keterangan |
|-------|------|------------|
| journal_number | VARCHAR UNIQUE/tenant | Format: `JRN-YYYY-MM-NNNNNNNN` |
| journal_date | DATE | |
| description | TEXT | |
| source_type | ENUM | transaction, manual, closing, adjustment, shu_distribution |
| source_id | UUID nullable | FK → sumber jurnal |
| status | ENUM | unposted, posted, reversed |
| posted_at / posted_by | TIMESTAMPTZ / UUID | |
| period_id | UUID (FK) | FK → accounting_period |
| total_debit | NUMERIC(15,2) | |
| total_credit | NUMERIC(15,2) | CHECK (total_debit = total_credit) |

### Tabel `jurnal_line` (Detail)

| Field | Tipe | Keterangan |
|-------|------|------------|
| jurnal_id | UUID (FK) | |
| coa_id | UUID (FK) | |
| line_number | INTEGER | Urutan baris |
| debit | NUMERIC(15,2) NOT NULL DEFAULT 0 | |
| credit | NUMERIC(15,2) NOT NULL DEFAULT 0 | |
| description | TEXT nullable | |
| rekening_id | UUID nullable | Opsional, jika terkait rekening spesifik |

**Constraints:** `CHECK (debit >= 0 AND credit >= 0)`, `CHECK (NOT (debit > 0 AND credit > 0))`, `CHECK (debit > 0 OR credit > 0)`

### Tabel `journal_mapping`

| Field | Tipe | Keterangan |
|-------|------|------------|
| transaction_type | VARCHAR | Trigger: `deposit_cash`, `withdrawal_cash`, dll |
| coop_type | ENUM | general, islamic, both |
| rules | JSONB | Array: `[{line, coa_code, side, amount_field}]` |
| is_system | BOOLEAN | System mapping tidak bisa dihapus |

### Tabel `accounting_period`

| Field | Tipe | Keterangan |
|-------|------|------------|
| period_type | ENUM | monthly, annual |
| year | INTEGER | |
| month | INTEGER | 1-12, nullable untuk annual |
| period_name | VARCHAR | `Januari 2026`, `Tahun Buku 2026` |
| status | ENUM | open, closed, locked |
| start_date / end_date | DATE | |

### Tabel `shu_periode`

| Field | Tipe | Keterangan |
|-------|------|------------|
| tahun_buku | INTEGER UNIQUE/tenant | |
| total_pendapatan / total_beban | NUMERIC(15,2) | |
| shu_bruto | NUMERIC(15,2) | = total_pendapatan - total_beban |
| pajak / zakat_institusi | NUMERIC(15,2) | |
| shu_neto | NUMERIC(15,2) | = shu_bruto - pajak - zakat_institusi |
| distribution_config | JSONB | Snapshot persentase distribusi |
| amount_cadangan | NUMERIC(15,2) | |
| amount_jasa_modal / amount_jasa_usaha | NUMERIC(15,2) | |
| amount_dana_pengurus / amount_dana_karyawan | NUMERIC(15,2) | |
| amount_dana_pendidikan / amount_dana_sosial | NUMERIC(15,2) | |
| rounding_difference | NUMERIC(15,2) | Selisih pembulatan → masuk cadangan |
| status | ENUM | calculated, reviewed, approved, distributed |
| rat_date / rat_minutes_url | DATE / VARCHAR | Tanggal dan notulen RAT |

### Tabel `shu_anggota`

| Field | Tipe | Keterangan |
|-------|------|------------|
| shu_periode_id | UUID (FK) | |
| nasabah_id | UUID (FK) | |
| avg_simpanan | NUMERIC(15,2) | Rata-rata harian simpanan selama tahun buku |
| total_transaksi | NUMERIC(15,2) | Total volume transaksi selama tahun buku |
| active_days | INTEGER | Jumlah hari aktif sebagai anggota |
| jasa_modal | NUMERIC(15,2) | |
| jasa_usaha | NUMERIC(15,2) | |
| total_shu | NUMERIC(15,2) | = jasa_modal + jasa_usaha |
| distribution_method | ENUM | credit_tabungan, separate_payout, pending |
| target_rekening_id | UUID nullable | |
| distributed_at / transaction_id | TIMESTAMPTZ / UUID | |

**Unique:** `(tenant_id, shu_periode_id, nasabah_id)`

---

## Chart of Accounts — Struktur Standar

### Koperasi Konvensional (PSAK / SAK ETAP): COA 1xxx-5xxx

```
1xxx — ASET
├── 1100  Kas (1101 Kas Teller, 1102 Kas Besar, 1103 Kas Kecil)
├── 1200  Bank
├── 1300  Piutang (1301 Piutang Pinjaman, 1302 Piutang Bunga, 1309 Cadangan Kerugian)
├── 1400  Penyertaan / Investasi
└── 1500  Aktiva Tetap

2xxx — KEWAJIBAN
├── 2100  Simpanan Nasabah (2101 Tabungan, 2102 Deposito, 2103 Bunga YHD Bayar)
└── 2200  Hutang

3xxx — EKUITAS
├── 3100  Simpanan Pokok Anggota
├── 3200  Simpanan Wajib Anggota
├── 3300  Cadangan Umum
├── 3400  Cadangan Risiko
├── 3500  SHU Tahun Berjalan
└── 3600  SHU Tahun Lalu

4xxx — PENDAPATAN
├── 4100  Pendapatan Bunga Pinjaman
├── 4200  Pendapatan Administrasi
├── 4300  Pendapatan Denda
└── 4900  Pendapatan Lain-lain

5xxx — BEBAN
├── 5100  Beban Bunga Simpanan
├── 5200  Beban Bunga Deposito
├── 5300  Beban Operasional (5301 Gaji, 5302 Sewa, 5303 ATK, 5304 Penyusutan)
└── 5400  Beban Cadangan Kerugian Piutang
```

### BMT/Koperasi Syariah — Tambahan (COA 6xxx-8xxx, Islamic only)

```
6xxx — DANA ZAKAT & INFAQ
├── 6100  Penerimaan Zakat (6101 Zakat Maal, 6102 Zakat Fitrah, 6103 Zakat Institusi)
├── 6200  Penyaluran Zakat (6201-6208 per asnaf)
├── 6300  Penerimaan Infaq/Shadaqah
└── 6400  Penyaluran Infaq/Shadaqah

7xxx — DANA KEBAJIKAN (QARDHUL HASAN)
├── 7100  Penerimaan Dana Kebajikan
├── 7200  Penyaluran Dana Kebajikan
└── 7300  Saldo Dana Kebajikan

8xxx — DANA TA'ZIR (SOCIAL FUND)
├── 8100  Penerimaan Ta'zir (8101 Denda Keterlambatan Anggota)
├── 8200  Penyaluran Dana Sosial
└── 8300  Saldo Dana Ta'zir
```

**Penting**: Akun 6xxx, 7xxx, 8xxx **hanya tersedia** di mode Islamic — hidden entirely di mode general.

---

## Auto-Journal Mapping (Contoh)

| Transaksi | Debit | Kredit |
|-----------|-------|--------|
| Setoran tunai tabungan | 1101 Kas Teller | 2101 Simpanan Tabungan |
| Penarikan tunai tabungan | 2101 Simpanan Tabungan | 1101 Kas Teller |
| Setoran simpanan pokok | 1101 Kas Teller | 3100 Simpanan Pokok |
| Setoran simpanan wajib | 1101 Kas Teller | 3200 Simpanan Wajib |
| Pencairan pinjaman (tunai) | 1301 Piutang Pinjaman | 1101 Kas Teller |
| Angsuran pokok (tunai) | 1101 Kas Teller | 1301 Piutang Pinjaman |
| Angsuran bunga/margin | 1101 Kas Teller | 4100 Pendapatan Bunga/Margin |
| Biaya admin bulanan | 2101 Simpanan Tabungan | 4200 Pendapatan Admin |
| Distribusi bunga/bagi hasil | 5100 Beban Bunga/Bagi Hasil | 2101 Simpanan Tabungan |
| Denda keterlambatan (general) | 1101 Kas Teller | 4300 Pendapatan Denda |
| Ta'zir (Islamic) | 1101 Kas Teller | 8101 Penerimaan Ta'zir |
| Penerimaan zakat (Islamic) | 1101 Kas Teller | 6101 Zakat Maal |
| Penyaluran zakat (Islamic) | 6201 Penyaluran Fakir | 1101 Kas Teller |
| Bayar gaji | 5301 Beban Gaji | 1101 Kas Teller |
| **PPAP Pembentukan** (bulanan, berdasarkan kolektibilitas) | 5400 Beban Cadangan Kerugian Piutang | 1309 Cadangan Kerugian Piutang |
| **PPAP Pembalikan** (pinjaman kembali Kol-1) | 1309 Cadangan Kerugian Piutang | 5400 Beban Cadangan Kerugian Piutang |
| **PPAP Hapus Buku / Write-off** (Kol-5, setelah Manager approval) | 1309 Cadangan Kerugian Piutang | 1301 Piutang Pinjaman |
| **Pemulihan Pinjaman yang Dihapusbukukan** | 1301 Piutang Pinjaman | 7100 Pendapatan Pemulihan Piutang |

**Catatan PPAP:**
- **Pembentukan**: Dijalankan oleh batch end-of-month berdasarkan kolektibilitas aktif seluruh pinjaman. Besaran PPAP sesuai persentase di tabel kolektibilitas (Kol-1: 0%, Kol-2: 1%, Kol-3: 10%, Kol-4: 50%, Kol-5: 100% dari outstanding principal).
- **Pembalikan**: Otomatis saat pembayaran angsuran masuk dan DPD kembali ke 0 (kolektibilitas upgrade ke Kol-1). Nominal pembalikan = selisih PPAP sebelum dan sesudah.
- **Hapus Buku**: Hanya dieksekusi setelah approval Admin; PPAP harus sudah 100% sebelum write-off diizinkan.
- **Pemulihan**: Saat nasabah membayar setelah pinjaman di-write-off — masuk ke akun 7100 (Pendapatan Pemulihan Piutang), bukan 4100.

---

## Business Rules

### Kas & Cash Flow (K014)

1. Saldo awal hari ini = saldo akhir kemarin — chain tidak boleh putus
2. Posisi kas dihitung **otomatis** dari transaksi hari tersebut — bukan input manual
3. Saldo akhir **tidak boleh negatif** — alert ke Supervisor/Manager jika mendekati nol
4. Inflow dari transaksi rekening (setoran, angsuran) **otomatis** terklasifikasi dari tipe transaksi K011
5. Biaya operasional diinput **manual** oleh Supervisor+ — wajib punya bukti/nota
6. Inter-branch cash transfer membutuhkan **Manager+ approval** di sisi pengirim; penerima konfirmasi + hitung denominasi
7. Kas minimum per branch configurable — sistem **tidak memblokir** transaksi saat kas rendah, hanya alert
8. Reconciliation harian otomatis:
   ```
   expected_kas_closing = SUM(teller_actual_cash) + vault_end_of_day + petty_cash
   Selisih = kas_harian.closing_balance - expected_kas_closing
   ```
9. Rekonsiliasi kas hanya untuk uang tunai fisik — transaksi non-tunai tidak masuk rekonsiliasi kas

### Jurnal & COA (K015)

10. Setiap transaksi (K011) **otomatis** menghasilkan jurnal entry — tidak perlu input manual
11. Auto-journal dilakukan dalam **satu database transaction** dengan transaksi sumber
12. Double-entry di-enforce di 3 level: application layer (JournalEngine), database CHECK constraint, scheduled reconciliation harian
13. Total debit = total kredit — zero tolerance, tidak ada selisih yang diizinkan
14. Jurnal dibuat dalam status **UNPOSTED** — hanya jurnal POSTED yang masuk ke laporan keuangan
15. **Auto-post option**: tenant bisa aktifkan auto-post untuk jurnal dari transaksi operasional
16. Jurnal manual (adjustment) **selalu UNPOSTED** — harus melalui review dan posting manual
17. Posting bersifat **irreversible** — koreksi via jurnal reversal baru (bukan edit)
18. Posting ke periode CLOSED atau LOCKED **diblokir**
19. Periode ditutup **berurutan** — tidak bisa close Maret jika Februari masih OPEN
20. System accounts (semua akun 4-digit dari template) tidak bisa dihapus atau diubah kodenya
21. Tenant bisa tambah **sub-akun** (5+ digit) dan **edit nama** akun; tidak bisa ubah kode atau hirarki 4-digit
22. Akun bisa dinonaktifkan (`is_active = false`) hanya jika saldo = 0 dan tidak ada jurnal unposted

### Year-End Closing

23. Semua 12 periode bulanan harus **CLOSED** sebelum year-end closing
24. Semua jurnal di tahun tersebut harus sudah **POSTED**
25. Closing menghasilkan jurnal otomatis yang memindahkan saldo 4xxx (Pendapatan) dan 5xxx (Beban) ke akun 3500 (SHU Tahun Berjalan)
26. Setelah closing, semua periode bulanan di-**LOCK** (tidak bisa reopen)
27. Saldo neraca (1xxx-3xxx) dibawa sebagai **opening balance** tahun buku baru
28. **Irreversible** — kesalahan dikoreksi via jurnal adjustment di tahun buku baru
29. Hanya **Admin** yang bisa menjalankan year-end closing

### SHU (K016)

30. SHU **hanya bisa dihitung** setelah year-end closing selesai
31. Cadangan minimum **25% dari SHU neto** — di-enforce di application layer (tidak bisa di-set < 25%)
32. Total semua persentase distribusi harus **= 100%**
33. Jasa Modal = proporsional terhadap **rata-rata harian simpanan** selama tahun buku (daily weighted average)
34. Jasa Usaha = proporsional terhadap **total volume transaksi** nasabah selama tahun buku
35. Anggota mid-year dihitung **prorated** (hari aktif / 365)
36. Siswa yang menjadi anggota berhak mendapat SHU sama seperti anggota lain
37. SHU siswa masuk ke rekening tabungan siswa (credit_tabungan) atau ditahan sebagai kewajiban
38. Distribusi bersifat **irreversible** — kesalahan dikoreksi via jurnal adjustment

### Formula Distribusi SHU

```
SHU Neto = SHU Bruto - Pajak - Zakat Institusi (Islamic)

Cadangan = SHU Neto × cadangan_pct         (min 25%)
Jasa Anggota Pool = SHU Neto × jasa_anggota_pct
  Jasa Modal = Jasa Anggota Pool × jasa_modal_split_pct
  Jasa Usaha = Jasa Anggota Pool × jasa_usaha_split_pct

Per nasabah (Jasa Modal):
  share_modal = avg_simpanan / total_avg_simpanan_semua × Jasa_Modal

Per nasabah (Jasa Usaha):
  share_usaha = total_transaksi / total_transaksi_semua × Jasa_Usaha

Total SHU nasabah = share_modal + share_usaha
```

**Algoritma Pembulatan SHU:**
1. Hitung SHU per anggota dengan presisi penuh (NUMERIC 15,2)
2. Floor ke Rp 1 per anggota
3. Remainder = pool - SUM(floored) → dialokasikan ke Cadangan
4. `rounding_difference` disimpan di `shu_periode` untuk audit

### Penyajian Neraca Islamic Mode (PSAK 109)

```
ASET = KEWAJIBAN + EKUITAS + DANA ZIS
         └── Dana Zakat (6xxx)
         └── Dana Infaq/Shadaqah (6300+)
         └── Dana Ta'zir (8xxx)
```

Dana ZIS disajikan sebagai **section terpisah di Neraca** — bukan kewajiban, bukan ekuitas. Ini adalah amanah yang dititipkan, bukan milik koperasi.

> **Referensi PSAK 109:** Sesuai **PSAK 109 paragraf 35-36**, dana ZIS disajikan sebagai pos terpisah dalam Neraca BMT (off-balance sheet section), bukan dicampur dengan modal/ekuitas koperasi. Akun 6xxx, 7xxx, dan 8xxx dalam COA SekolahPro mengimplementasikan penyajian ini secara langsung — sistem tidak akan memasukkan akun-akun tersebut ke dalam kalkulasi total ekuitas pada laporan neraca.

### Zakat & Infaq (K018) — Islamic Only

39. Modul Zakat & Infaq **hanya aktif** jika `coop_type = "islamic"` — API 404 di mode general
40. Tiga jenis zakat: Zakat Mal (2.5% dari harta qualifying - hutang, setelah haul 355 hari), Zakat Fitrah (fixed per kepala), Zakat Institusi (2.5% dari laba BMT)

#### Zakat Institusi BMT (2.5% Laba Bersih)

Berbeda dari zakat mal anggota individual (K018), **Zakat Institusi** adalah kewajiban zakat atas laba bersih BMT itu sendiri sebagai entitas.

- **Besaran:** 2.5% dari laba bersih BMT setelah tutup buku tahunan.
- **Dasar hukum:** Fatwa DSN-MUI dan prinsip bahwa BMT sebagai entitas yang mengumpulkan dan mengelola harta wajib mengeluarkan zakat atas keuntungan usahanya.
- **Pengambilan keputusan:** Diputuskan dan disahkan dalam **RAT (Rapat Anggota Tahunan)**.
- **Jurnal pembentukan (saat RAT menyetujui):**
  - Debit: 3100 SHU Tahun Berjalan (atau akun SHU yang relevan)
  - Kredit: 2600 Dana Zakat Institusi (akun kewajiban sementara menunggu distribusi)
- **Jurnal penyaluran (setelah distribusi via baitul maal):**
  - Debit: 2600 Dana Zakat Institusi
  - Kredit: 6103 Zakat Institusi (akun dana zakat)
- **Penyaluran** dilakukan via baitul maal ke asnaf yang berhak, sesuai prinsip yang sama dengan zakat mal.
- **Perbedaan dari Zakat Mal Anggota (K018):** Zakat mal anggota adalah kewajiban pribadi setiap anggota — dihitung dari harta pribadi mereka di koperasi. Zakat Institusi adalah kewajiban BMT sebagai badan usaha.
41. Nisab = 85 gram emas setara IDR — Admin wajib update minimal per kuartal
42. Haul dihitung dalam **tahun hijriah** (~355 hari)
43. Hutang yang dikurangi dari harta: hanya angsuran yang **jatuh tempo 12 bulan ke depan**, bukan total sisa pokok
44. Opt-in zakat mal: kalkulasi hanya mencakup **aset di koperasi** — aset di luar menjadi tanggung jawab pribadi
45. Dana ta'zir (K009) dikelola bersama dana infaq — tracking terpisah di laporan
46. Dana zakat **tidak boleh dicampur** dengan dana infaq dalam satu distribusi
47. Distribusi ke mustahik (8 asnaf per QS At-Taubah: 60) membutuhkan approval Manager+ (< threshold) atau Admin (≥ threshold)
48. **Bukti penerimaan wajib** untuk setiap distribusi — foto/scan tanda terima
49. Mustahik **bukan nasabah** — entitas terpisah dengan data lebih sederhana
50. Laporan terpisah untuk DPS (Dewan Pengawas Syariah): Lap. Zakat, Lap. Dana Kebajikan, Lap. Ta'zir

---

## Laporan Keuangan yang Di-Generate

| Laporan | Deskripsi | Frekuensi |
|---------|-----------|-----------|
| **Neraca** | Total Aset = Kewajiban + Ekuitas (+Dana ZIS untuk BMT) | Bulanan/Tahunan |
| **Laporan Laba Rugi** | SHU = Total Pendapatan - Total Beban | Bulanan/Tahunan |
| **Laporan Arus Kas** | Operasi + Investasi + Pendanaan | Bulanan/Tahunan |
| **Trial Balance** | SUM debit/kredit per akun | On-demand |
| **Laporan SHU per Anggota** | Detail jasa modal + jasa usaha per nasabah | Tahunan (saat distribusi) |
| **Lap. Sumber & Penyaluran Zakat** | Per asnaf *(Islamic only)* | Bulanan/Tahunan |
| **Lap. Dana Kebajikan** | Infaq + ta'zir *(Islamic only)* | Bulanan/Tahunan |

---

## RBAC Summary

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View kas harian | v* | v | v | v |
| Create non-rekening transaction | - | v† | v | v |
| Approve inter-branch transfer | - | - | v | v |
| Reconcile kas harian | - | - | v | v |
| View COA | v | v | v | v |
| Add sub-akun | - | - | v | v |
| Create jurnal manual | - | v | v | v |
| Post jurnal | - | v | v | v |
| Reverse jurnal | - | - | v | v |
| Close periode bulanan | - | - | v | v |
| Lock periode / Year-end closing | - | - | - | v |
| Trigger SHU calculation | - | - | - | v |
| Review SHU | - | - | v | v |
| Approve SHU (RAT) & Distribute | - | - | - | v |
| Run SHU simulation | - | - | v | v |
| Terima zakat/infaq | v | v | v | v |
| Approve distribusi zakat (< threshold) | - | - | v | v |
| Approve distribusi zakat (≥ threshold) | - | - | - | v |
| Configure nisab/fitrah | - | - | - | v |

`*` Teller hanya lihat summary, tidak detail operasional
`†` Supervisor bisa buat non-rekening transaction minor (≤ threshold)

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Auto-journal dari setiap transaksi | Konsistensi; mengurangi human error; operator tidak perlu paham akuntansi double-entry |
| Double-entry di 3 level enforcement | Safety net berlapis; database constraint adalah fallback jika bug di application layer |
| UNPOSTED → POSTED (irreversible) | Review sebelum masuk laporan; kesalahan tidak mudah disebarkan |
| Period management (open/closed/locked) | Melindungi data historis; laporan yang sudah diterbitkan tidak berubah retroaktif |
| COA template berbeda per coop_type | Konvensional (PSAK SAK ETAP) dan syariah (PSAK 101-110) memiliki COA berbeda |
| Dana ZIS sebagai section terpisah di Neraca | Compliance PSAK 109 paragraf 35-36 |
| Cadangan SHU min 25% di-enforce | UU Koperasi No. 25/1992 |
| Distribusi SHU via RAT | UU Koperasi mengharuskan distribusi disahkan di RAT |
| Pembulatan SHU ke Cadangan | Transparansi; `rounding_difference` tersimpan untuk audit |
| Zakat API 404 di general mode (bukan 403) | Clean separation; modul tidak eksis, bukan forbidden |
| Hutang dikurangi = angsuran 12 bulan ke depan (bukan total sisa pokok) | Pendapat mayoritas ulama fiqh; implementasi lebih proporsional |

---

## Integration Points

| Integrasi | Arah | Keterangan |
|-----------|------|------------|
| `transaksi` (K011) | Jurnal ← Transaksi | Auto-journal triggered oleh TransactionCreatedEvent |
| `teller_session` (K012) | Kas ← Teller | Daily consolidation dari teller session ke kas harian |
| `pinjaman` (K007) | Jurnal ← Pinjaman | PPAP dihitung berdasarkan kolektibilitas |
| `shu_periode` (K016) | Jurnal → SHU | Year-end closing menghasilkan saldo SHU di akun 3500 |
| `laporan_regulasi` (K017) | Laporan ← Jurnal | Trial balance dan laporan keuangan dari data jurnal POSTED |
| `denda` (K009) | Jurnal ← Ta'zir | Ta'zir → akun 8101 (bukan 4300); ta'widh → akun 4900 |
| `zakat_collection` (K018) | Jurnal ← Zakat | Penerimaan zakat → akun 6xxx |
