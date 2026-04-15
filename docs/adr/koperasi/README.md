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
| K003 | [Produk & Akad](./ADR-K003-produk-akad.md) | - | - | - | - | Katalog produk + akad syariah (BMT mode) |

### Simpanan

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K004 | [Simpanan Pokok & Wajib](./ADR-K004-simpanan-pokok-wajib.md) | - | - | - | - | Mandatory, syarat keanggotaan |
| K005 | [Tabungan](./ADR-K005-tabungan.md) | - | - | - | - | Simpanan sukarela, multiple jenis |
| K006 | [Deposito / Simpanan Berjangka](./ADR-K006-deposito.md) | - | - | - | - | Time deposit, jatuh tempo |

### Pembiayaan / Pinjaman

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K007 | [Pinjaman / Pembiayaan](./ADR-K007-pinjaman.md) | - | - | - | - | Konvensional: bunga · BMT: margin/bagi hasil |
| K008 | [Angsuran & Jadwal](./ADR-K008-angsuran-jadwal.md) | - | - | - | - | Jadwal cicilan, kalkulasi sisa pokok |
| K009 | [Denda & Penalti](./ADR-K009-denda-penalti.md) | - | - | - | - | Konvensional: denda bunga · BMT: ta'zir |
| K010 | [Jaminan / Agunan](./ADR-K010-jaminan-agunan.md) | - | - | - | - | Collateral management |

### Transaksi & Operasional

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K011 | [Transaksi Rekening & Non Rekening](./ADR-K011-transaksi.md) | - | - | - | - | Core transaction engine |
| K012 | [Teller Session](./ADR-K012-teller-session.md) | - | - | - | - | Buka/tutup kas, cash balancing |
| K013 | [Money Denomination](./ADR-K013-money-denomination.md) | - | - | - | - | Pecahan uang fisik untuk teller |
| K014 | [Kas & Cash Flow](./ADR-K014-kas-cashflow.md) | - | - | - | - | Posisi kas harian |

### Akuntansi & Pelaporan

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K015 | [Jurnal & Akuntansi (COA)](./ADR-K015-jurnal-coa.md) | - | - | - | - | Chart of Accounts, double-entry |
| K016 | [SHU (Sisa Hasil Usaha)](./ADR-K016-shu.md) | - | - | - | - | Perhitungan & distribusi tahunan |
| K017 | [Laporan Regulasi](./ADR-K017-laporan-regulasi.md) | - | - | - | - | Format laporan OJK / Dinas Koperasi |

### Khusus BMT (Islamic Mode)

| No. | Judul | ADR | Code | Test API | Test UI | Notes |
|-----|-------|-----|------|----------|---------|-------|
| K018 | [Zakat & Infaq](./ADR-K018-zakat-infaq.md) | - | - | - | - | Feature flag: `coop_type = "islamic"` |

---

## Items per Domain

Ringkasan entity/item utama yang terlibat di setiap ADR:

| ADR | Items |
|-----|-------|
| K001 | nasabah, anggota, identitas, kontak, status_keanggotaan, ahli_waris, specimen, kyc_level, sponsor, transfer_cabang, merge |
| K002 | rekening, rekening_application, form_template, nomor_rekening, saldo, status_rekening, produk_ref, dormant |
| K003 | produk, jenis_produk, akad, nisbah, margin, suku_bunga |
| K004 | simpanan_pokok, simpanan_wajib, saldo_keanggotaan |
| K005 | tabungan, setoran, penarikan, saldo_tabungan |
| K006 | deposito, jatuh_tempo, perpanjangan, pencairan |
| K007 | pinjaman, pembiayaan, plafon, tenor, status_pinjaman |
| K008 | angsuran, jadwal_angsuran, pokok, margin_bunga, sisa_pokok |
| K009 | denda, ta'zir, keterlambatan, penalti_pelunasan_dini |
| K010 | jaminan, agunan, jenis_jaminan, nilai_taksasi, status_jaminan |
| K011 | transaksi, jurnal_entry, debit, kredit, referensi |
| K012 | sesi_teller, kas_awal, kas_akhir, selisih, approval |
| K013 | denominasi, pecahan, jumlah_lembar, total_nominal |
| K014 | kas_harian, pemasukan, pengeluaran, saldo_kas |
| K015 | coa, akun, jurnal, neraca, laba_rugi |
| K016 | shu, kontribusi_anggota, persentase_bagi, distribusi |
| K017 | laporan_ojk, laporan_dinas, periode, format |
| K018 | zakat, infaq, muzakki, mustahik, nisab, haul |
