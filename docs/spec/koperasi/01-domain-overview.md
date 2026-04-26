# Domain Koperasi — Gambaran Umum

SekolahPro Koperasi adalah modul pengelolaan koperasi simpan pinjam berbasis sekolah (Koperasi Sekolah) yang mendukung dua mode operasional dalam satu codebase: **konvensional (general)** dan **syariah (Islamic/BMT)**. Modul ini dirancang untuk koperasi yang beroperasi di lingkungan sekolah, pesantren, dan madrasah — melayani guru, santri/siswa, staf, orang tua, dan masyarakat sekitar.

---

## ADR References

| ADR | Topik |
|-----|-------|
| ADR-K001 | Nasabah (Anggota) |
| ADR-K002 | Rekening |
| ADR-K003 | Produk & Akad |
| ADR-K004 | Simpanan Pokok & Wajib |
| ADR-K005 | Tabungan |
| ADR-K006 | Deposito |
| ADR-K007 | Pinjaman |
| ADR-K008 | Angsuran & Jadwal |
| ADR-K009 | Denda & Penalti |
| ADR-K010 | Jaminan & Agunan |
| ADR-K011 | Transaksi |
| ADR-K012 | Teller Session |
| ADR-K013 | Money Denomination |
| ADR-K014 | Kas & Cash Flow |
| ADR-K015 | Jurnal & COA |
| ADR-K016 | SHU |
| ADR-K017 | Laporan Regulasi |
| ADR-K018 | Zakat & Infaq |
| ADR-K019 | Toko & Kantin |
| ADR-K020 | Payroll Integration |

---

## Scope dan Tujuan

Koperasi Sekolah diatur oleh:
- **UU No. 25 Tahun 1992** tentang Perkoperasian
- **Peraturan OJK** untuk Lembaga Keuangan Mikro (LKM) — berlaku jika koperasi terdaftar sebagai LKM
- **Fatwa DSN-MUI** — untuk operasional syariah (BMT mode)
- **PSAK / SAK ETAP** — akuntansi konvensional
- **PSAK 101-110 Syariah** — akuntansi syariah

Modul ini mengelola:
1. **Keanggotaan (Member)** — pendaftaran, KYC, ahli waris, lifecycle anggota
2. **Rekening (Account)** — simpanan pokok, simpanan wajib, tabungan, deposito, pinjaman
3. **Produk & Akad** — katalog produk keuangan dengan konfigurasi rate/nisbah
4. **Pinjaman/Pembiayaan** — aplikasi kredit, angsuran, jaminan, NPL
5. **Transaksi** — setoran, penarikan, angsuran, pencairan — immutable + atomic
6. **Teller & Kas** — sesi teller, denominasi uang, kas harian
7. **Akuntansi** — double-entry, COA, jurnal otomatis, laporan keuangan
8. **SHU** — bagi hasil tahunan sesuai UU Koperasi
9. **Kepatuhan & Regulasi** — laporan ke Dinas Koperasi dan OJK
10. **Zakat & Infaq** *(Islamic only)* — Baitul Maal, pengelolaan dana sosial
11. **Toko & Kantin** — unit usaha non-finansial koperasi
12. **Payroll Deduction** — integrasi potong gaji guru/staf

---

## Domain Entities Utama

| Entity | Deskripsi | ADR |
|--------|-----------|-----|
| `nasabah` | Anggota koperasi — siswa, guru, staf, orang tua, eksternal | K001 |
| `nasabah_application` | Formulir pendaftaran anggota dengan approval flow | K001 |
| `nasabah_beneficiary` | Ahli waris anggota | K001 |
| `rekening` | Akun keuangan milik nasabah per kategori produk | K002 |
| `rekening_application` | Formulir pembukaan rekening dengan approval flow | K002 |
| `produk` | Katalog produk keuangan dengan rate/akad/limit | K003 |
| `produk_version_history` | Riwayat perubahan konfigurasi produk | K003 |
| `simpanan_wajib_billing` | Tagihan simpanan wajib bulanan per nasabah | K004 |
| `auto_debit_config` | Konfigurasi auto-debit dari tabungan | K004 |
| `tabungan_statement` | Rekening koran tabungan bulanan | K005 |
| `deposito_interest_schedule` | Jadwal pembayaran bunga/bagi hasil deposito | K006 |
| `deposito_bilyet` | Sertifikat/bilyet deposito | K006 |
| `pinjaman` | Catatan pinjaman/pembiayaan aktif | K007 |
| `pinjaman_application` | Formulir pengajuan pinjaman | K007 |
| `pinjaman_penjamin` | Data penjamin per pinjaman | K007 |
| `angsuran` | Jadwal dan pencatatan pembayaran cicilan | K008 |
| `angsuran_pembayaran` | Detail alokasi per pembayaran angsuran | K008 |
| `denda` | Catatan denda/penalti/ta'zir | K009 |
| `jaminan` | Agunan/kolateral pinjaman | K010 |
| `jaminan_pinjaman` | Binding jaminan ke pinjaman (many-to-many) | K010 |
| `jaminan_dokumen` | Dokumen agunan (BPKB, sertifikat, dll) | K010 |
| `transaksi` | Setiap pergerakan saldo rekening (immutable) | K011 |
| `kas_transaksi` | Transaksi kas non-rekening (operasional) | K011 |
| `transaction_batch` | Batch operations (simpanan wajib, bagi hasil) | K011 |
| `teller_session` | Sesi kerja teller per shift | K012 |
| `denominasi_master` | Katalog pecahan uang per tenant | K013 |
| `vault_count` | Penghitungan fisik brankas | K013 |
| `kas_harian` | Posisi kas harian per cabang | K014 |
| `coa` | Chart of Accounts per tenant | K015 |
| `jurnal` | Jurnal akuntansi double-entry | K015 |
| `jurnal_line` | Baris debit/kredit per jurnal | K015 |
| `journal_mapping` | Aturan auto-journal per tipe transaksi | K015 |
| `accounting_period` | Periode akuntansi (bulanan/tahunan) | K015 |
| `shu_periode` | Data SHU per tahun buku | K016 |
| `shu_anggota` | Alokasi SHU per anggota | K016 |
| `laporan` | Laporan regulasi (Dinas Koperasi, OJK) | K017 |
| `laporan_config` | Konfigurasi jadwal dan template laporan | K017 |
| `zakat_collection` | Penerimaan zakat *(Islamic only)* | K018 |
| `zakat_distribution` | Penyaluran zakat ke mustahik *(Islamic only)* | K018 |
| `mustahik` | Penerima zakat/bantuan *(Islamic only)* | K018 |
| `infaq` | Penerimaan infaq/shadaqah *(Islamic only)* | K018 |
| `toko_produk` | Katalog barang toko/kantin | K019 |
| `toko_penjualan` | Transaksi POS toko/kantin | K019 |
| `payroll_batch` | Batch data payroll untuk potongan gaji | K020 |
| `payroll_deduction` | Detail potongan per karyawan per batch | K020 |
| `payroll_authorization` | Otorisasi tertulis potongan gaji nasabah | K020 |

---

## Aturan Bisnis dari UU Koperasi

1. Setiap anggota **wajib membayar Simpanan Pokok** saat masuk keanggotaan — satu kali, tidak bisa ditarik selama aktif
2. Setiap anggota **wajib membayar Simpanan Wajib** secara bulanan — tidak bisa ditarik selama aktif
3. **Cadangan SHU minimum 25%** dari Sisa Hasil Usaha — tidak bisa dikurangi
4. **SHU dibagikan** berdasarkan jasa modal (simpanan) dan jasa usaha (transaksi), bukan flat per kepala
5. Pembagian SHU **wajib disahkan di RAT** (Rapat Anggota Tahunan)
6. **Dokumentasi keanggotaan** wajib formal — formulir aplikasi, persetujuan pengurus, ahli waris
7. **Laporan berkala** wajib disampaikan ke Dinas Koperasi (RAT, neraca, laporan usaha)
8. Simpanan pokok dan wajib **wajib dikembalikan** saat anggota keluar (dikurangi kewajiban yang sah)
9. **BMPK** — batas kredit ke satu anggota maksimal 20% modal sendiri; grup 25%
10. **Klasifikasi kolektibilitas** mengikuti standar OJK (Kol 1-5 berbasis DPD)

---

## Prinsip Syariah yang Diterapkan (Islamic/BMT Mode)

Prinsip-prinsip syariah berikut hanya berlaku jika `coop_type = "islamic"`:

| Prinsip | Implementasi |
|---------|-------------|
| **Akad Wadiah** (titipan) | Simpanan pokok, simpanan wajib, tabungan wadiah — nasabah menitipkan dana, boleh diberi bonus (tidak wajib) |
| **Akad Mudharabah** (bagi hasil) | Tabungan dan deposito mudharabah — profit dibagi sesuai nisbah yang disepakati berdasarkan keuntungan riil |
| **Akad Murabahah** (jual beli) | Pembiayaan — koperasi beli barang, jual ke nasabah dengan margin tetap yang disepakati |
| **Akad Musyarakah** (kemitraan) | Pembiayaan modal usaha — modal bersama, bagi hasil sesuai nisbah |
| **Akad Ijarah** (sewa) | Pembiayaan aset — nasabah bayar ujrah (biaya sewa) berkala |
| **Akad Qardh** (pinjaman kebajikan) | Pembiayaan darurat — tanpa margin, hanya biaya admin |
| **Akad Rahn** (gadai) | Jaminan fisik — koperasi menyimpan Marhun, nasabah membayar ujrah penyimpanan |
| **Ta'zir** (denda) | Nominal tetap (bukan persentase), wajib masuk dana sosial — tidak boleh jadi pendapatan koperasi |
| **Ibra'** (diskon) | Pelunasan dini pembiayaan — sisa margin yang belum jatuh tempo wajib didiskon |
| **Off-balance sheet ZIS** | Dana Zakat, Infaq, Shadaqah disajikan sebagai section terpisah di Neraca per PSAK 109 |
| **Larangan Riba** | Tidak ada bunga — pendapatan dari margin, nisbah bagi hasil, atau ujrah |
| **Transparent Indicative Rate** | Rate bagi hasil deposito mudharabah hanya indikatif, bukan janji — disclaimer wajib ditampilkan |

---

## Dual-Mode Architecture

Satu tenant hanya memiliki satu `coop_type`:
- `general` — koperasi konvensional (bunga, pinjaman)
- `islamic` — BMT/koperasi syariah (akad, margin/nisbah, pembiayaan)

Perbedaan utama antar mode:

| Aspek | General | Islamic |
|-------|---------|---------|
| Terminologi produk | Pinjaman | Pembiayaan |
| Yield tabungan | Bunga (daily balance) | Bagi Hasil (nisbah × profit riil) |
| Yield deposito | Bunga tetap (fixed) | Nisbah indikatif (aktual per bulan) |
| Rate pinjaman | Suku bunga (flat/declining/annuity) | Margin Murabahah / Nisbah Musyarakah / Ujrah Ijarah |
| Denda | % dari tunggakan → pendapatan koperasi | Ta'zir nominal tetap → dana sosial |
| Pelunasan dini | Penalti opsional | Ibra' wajib (diskon margin) |
| Akuntansi | PSAK/SAK ETAP, COA 1xxx-5xxx | PSAK 101-110, COA 1xxx-8xxx |
| Laporan tambahan | - | Lap. Zakat, Dana Kebajikan, Ta'zir |
| Zakat & Infaq | Tidak ada | Ada (Baitul Maal function) |
| Formula SHU | SHU = pendapatan bunga - beban | SHU = pendapatan margin/bagi hasil - beban |

---

## Pola Arsitektur Teknis

### Vernon Denormalized Read-Cache
Semua entitas utama menggunakan pola Vernon dengan dua field JSONB:
- `_rels` — menyimpan UUID relasi (tenant_id, branch_id, nasabah_id, dll)
- `_data` — menyimpan snapshot data relasi untuk listing tanpa JOIN

**Keamanan**: `_data` **tidak pernah** menyimpan data sensitif (identity_number, saldo, data finansial personal).

### CQRS
Command (write) dan Query (read) dipisahkan. Read menggunakan `_data` cache; write selalu ke tabel asli.

### Immutability
- **Transaksi**: tidak bisa diedit/dihapus setelah commit — koreksi via reversal
- **Jurnal posted**: tidak bisa diedit — koreksi via jurnal reversal baru
- **Nomor anggota/rekening**: immutable setelah assigned

### Atomic Transactions
Setiap operasi yang mempengaruhi saldo rekening **wajib** dalam satu database transaction:
`INSERT transaksi + UPDATE rekening.balance + UPDATE teller_session.totals`

### Event-Driven Sync
`SyncEngine` memperbarui `_data` di semua entitas yang mereferensikan entitas yang berubah melalui events (NasabahUpdatedEvent, ProductUpdatedEvent, BranchUpdatedEvent, dll).

---

## Integration Points

| Integrasi | Arah | Keterangan |
|-----------|------|------------|
| **SekolahPro Core — Sekolah** | Koperasi ← Sekolah | `school_entity_id` di nasabah menghubungkan ke entitas siswa/guru/staf di modul sekolah |
| **SekolahPro HR/Payroll** | Koperasi ← HR | Payroll deduction (K020) mengonsumsi data gaji dari sistem HR sekolah |
| **SekolahPro Portal Orang Tua** | Koperasi → Portal | Spending limit, notifikasi saldo, laporan belanja siswa (K019, K023) |
| **Dinas Koperasi** | Koperasi → Regulator | Laporan RAT, neraca, laporan usaha (K017) |
| **OJK** | Koperasi → Regulator | Laporan kolektibilitas, CAR, BMPK *(jika LKM)* (K017) |
| **DSN-MUI** | BMT ← Fatwa | Compliance akad, ta'zir, ibra', zakat (K003, K009, K018) |
| **Bank Eksternal** | Koperasi ↔ Bank | Transfer dana nasabah, setoran deposito dari bank (via teller reconciliation) |

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Single codebase untuk dua mode (general + Islamic) | Mengurangi duplikasi kode; perbedaan hanya di terminologi, rate config, dan beberapa aturan bisnis khusus syariah |
| Vernon pattern untuk semua entitas utama | Listing rekening/transaksi/pinjaman membutuhkan 3-5 JOIN; cached _data menghilangkan JOIN dan menjaga performa < 100ms |
| Immutable transaksi | Integritas audit trail; compliance OJK tidak bisa ada manipulasi catatan keuangan |
| Approval flow bertingkat (Teller → Supervisor → Manager → Admin) | Separation of duties; mencegah satu orang melakukan seluruh siklus transaksi berisiko tinggi |
| Hard delete dihindari di semua entitas | Referential integrity; data historis diperlukan untuk laporan regulasi, SHU, dan audit |
| Soft deactivation untuk anggota | Data anggota tidak dihapus; dibutuhkan untuk audit dan laporan historis |
| Cadangan minimum 25% SHU di-enforce | Wajib per UU Koperasi No. 25/1992; constraint di application layer |
