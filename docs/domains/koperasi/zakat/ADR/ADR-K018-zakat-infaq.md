# ADR-K018: Zakat & Infaq

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Zakat dan Infaq merupakan pilar keuangan Islam yang wajib dikelola oleh BMT (Baitul Maal wat Tamwil). Sebagai lembaga keuangan syariah, BMT memiliki **dua fungsi**:

- **Baitul Tamwil** — fungsi komersial (pembiayaan, tabungan, deposito)
- **Baitul Maal** — fungsi sosial (pengelolaan zakat, infaq, shadaqah, dan dana sosial)

Sistem harus mengakomodasi:

- **Zakat mal** (harta) dengan kalkulasi nisab, haul, dan persentase sesuai fiqh
- **Zakat fitrah** per kepala selama Ramadan
- **Zakat institusi** atas laba bersih BMT sendiri
- **Infaq dan shadaqah** (donasi sukarela) dengan mekanisme recurring atau one-time
- **Pengelolaan mustahik** (penerima) berdasarkan 8 asnaf
- **Integrasi ta'zir fund** dari [ADR-K009](./ADR-K009-denda-penalti.md) ke pengelolaan dana sosial
- **Akuntansi off-balance sheet** dengan COA 6xxx sesuai [ADR-K015](./ADR-K015-jurnal-coa.md)
- **Konteks sekolah** — zakat fitrah santri, distribusi beasiswa

Referensi terkait:
- [ADR-009](../core/ADR-009-dual-mode-institution-type.md) — Dual-mode institution type
- [ADR-K001](./ADR-K001-nasabah.md) — Nasabah / anggota
- [ADR-K009](./ADR-K009-denda-penalti.md) — Denda & penalti (ta'zir → dana sosial)
- [ADR-K015](./ADR-K015-jurnal-coa.md) — Jurnal & akuntansi (COA 6xxx zakat)

## Decision

### 1. Feature Activation — Islamic Only

Seluruh modul Zakat & Infaq **hanya aktif** ketika `coop_type = "islamic"`. Di mode general, modul ini **completely hidden** — tidak ada menu, tidak ada endpoint, tidak ada tabel yang diakses.

```
Feature flag derivation:

tenant_config.coop_type = "islamic"
    → enable: [zakat_module, infaq_module, mustahik_registry, baitul_maal_dashboard]

tenant_config.coop_type = "general"
    → disable: [zakat_module, infaq_module, mustahik_registry, baitul_maal_dashboard]
    → API endpoint return 404 (bukan 403)
    → Menu item tidak di-render sama sekali
```

**Aturan:**
- Feature flag **derived** dari `coop_type`, bukan konfigurasi terpisah — tidak bisa diaktifkan manual di mode general
- API endpoint untuk zakat/infaq mengembalikan **404** (bukan 403) di mode general — modul tidak eksis
- Middleware check `coop_type` di awal request pipeline — reject early, zero overhead
- Migrasi database tetap membuat tabel zakat/infaq di semua mode — hanya akses yang dibatasi

### 2. Jenis Zakat

Sistem mengelola **tiga jenis zakat** dengan kalkulasi dan periode berbeda:

```
ZAKAT MAL (Harta)
├── Subjek: Anggota/nasabah yang opt-in
├── Nisab: 85 gram emas setara IDR (configurable)
├── Haul: 1 tahun hijriah sejak harta melebihi nisab
├── Tarif: 2.5% dari (total harta qualifying - hutang)
├── Periode: Kapan saja, dihitung annual per anggota
└── Sifat: Wajib bagi yang memenuhi syarat

ZAKAT FITRAH
├── Subjek: Setiap jiwa Muslim (anggota + keluarga)
├── Besaran: Fixed amount per orang (configurable, default: ~2.5 kg beras equivalent)
├── Periode: 1 Ramadan - 1 Syawal (configurable per tahun hijriah)
├── Sifat: Wajib per kepala
└── Konteks sekolah: Koleksi bulk per kelas/halaqah untuk santri

ZAKAT INSTITUSI (Laba BMT)
├── Subjek: BMT/koperasi sebagai badan usaha
├── Basis: Laba bersih tahunan BMT (net profit after tax)
├── Tarif: 2.5% dari laba bersih
├── Periode: Setelah tutup buku tahunan, sebelum RAT
└── Sifat: Institutional — diputuskan dalam RAT
```

### 3. Nisab & Haul Configuration

Konfigurasi nisab dan haul dikelola oleh Admin dan bisa di-update berkala mengikuti harga emas terkini.

```
nisab_config:
  gold_gram_equivalent: 85            # 85 gram emas (standar fiqh)
  nisab_amount_idr: 127500000         # Setara IDR (85 x harga emas per gram)
  last_updated_at: "2026-04-15"       # Kapan terakhir diupdate
  updated_by: "admin-uuid"            # Siapa yang update
  source_reference: "Harga emas Antam 15 Apr 2026"  # Referensi harga

haul_config:
  calendar_type: "hijriah"            # Tahun hijriah (355 hari)
  haul_days: 355                      # 1 tahun hijriah (~355 hari)
  tracking_mode: "per_member"         # Tracking per anggota individual
```

**Aturan:**
- Nisab di-set dalam **IDR** berdasarkan harga emas terkini — Admin wajib update minimal per kuartal
- Sistem menyimpan **riwayat perubahan nisab** — kalkulasi menggunakan nisab yang berlaku saat haul tercapai
- Haul dihitung dalam **tahun hijriah** (~355 hari) — sistem track tanggal hijriah otomatis
- Satu anggota bisa memiliki **haul start date berbeda** tergantung kapan hartanya pertama kali melewati nisab
- Jika harta turun di bawah nisab sebelum haul tercapai → haul di-reset, mulai dari awal

### 4. Zakat Calculation untuk Anggota (Opt-in)

Anggota yang memilih layanan kalkulasi zakat otomatis akan dihitung berdasarkan aset qualifying di koperasi.

```
Opt-in Flow:
┌─────────────────────────┐
│ Anggota mengaktifkan     │  Via teller atau self-service
│ fitur zakat calculation  │
└────────┬────────────────┘
         │
         v
┌─────────────────────────┐
│ Sistem track qualifying  │  Simpanan pokok + wajib + tabungan + deposito
│ assets per anggota       │
└────────┬────────────────┘
         │ Setiap hari, cek nisab
         v
┌─────────────────────────┐
│ Harta >= nisab?          │
│ ├── Ya  → mulai/lanjut  │  Track haul start date
│ │        haul counter    │
│ └── Tidak → reset haul  │  Haul counter kembali ke 0
└────────┬────────────────┘
         │ Haul tercapai (355 hari)
         v
┌─────────────────────────┐
│ Auto-calculate zakat:    │
│ (total assets - hutang)  │
│ x 2.5%                   │
└────────┬────────────────┘
         │
         v
┌─────────────────────────┐
│ Notifikasi ke anggota:   │  "Zakat Anda tahun ini: Rp X"
│ Anggota confirm/adjust   │  Bisa adjust manual sebelum bayar
└────────┬────────────────┘
         │ Confirm
         v
┌─────────────────────────┐
│ Eksekusi pembayaran:     │  Auto-debit dari tabungan (jika authorized)
│ - Auto-debit tabungan    │  atau manual via teller
│ - Manual via teller      │
└─────────────────────────┘
```

**Qualifying assets (aset di koperasi):**
- Simpanan pokok
- Simpanan wajib
- Tabungan (semua jenis)
- Deposito (pokok, belum termasuk profit yang belum jatuh tempo)

**Bukan qualifying assets:**
- Pinjaman/pembiayaan outstanding (ini hutang, bukan aset)
- Bunga/profit yang belum dicairkan

**Pengurangan Hutang dalam Perhitungan Zakat Mal:**

Hutang yang boleh dikurangkan dari harta zakat-able:

```
harta_neto = total_harta_qualifying - hutang_jatuh_tempo
```

- `hutang_jatuh_tempo`: Hanya angsuran yang **jatuh tempo dalam periode haul** (12 bulan ke depan), BUKAN total sisa pokok pinjaman
- Contoh: Nasabah punya pinjaman sisa pokok Rp 50.000.000, angsuran Rp 1.000.000/bulan
  → Hutang yang dikurangkan = 12 × Rp 1.000.000 = Rp 12.000.000 (bukan Rp 50.000.000)
- Dasar: pendapat mayoritas ulama fiqh bahwa hutang jangka panjang hanya dikurangkan sebesar porsi yang jatuh tempo

Konfigurasi per tenant:
```
debts_calculation_method: "due_installments_only" | "total_outstanding"
Default: "due_installments_only" (recommended)
```

**Aturan:**
- Opt-in **bersifat sukarela** — anggota tidak dipaksa menggunakan layanan ini
- Kalkulasi hanya mencakup **aset di koperasi** — aset di luar koperasi menjadi tanggung jawab pribadi
- Anggota bisa **adjust nominal** sebelum konfirmasi pembayaran (misal: menambah karena aset di luar koperasi)
- Kalkulasi otomatis berjalan **annual** per anggota berdasarkan haul masing-masing
- Hasil kalkulasi bersifat **rekomendasi** — keputusan bayar tetap di tangan anggota

### 5. Infaq & Shadaqah — Donasi Sukarela

Infaq dan shadaqah adalah donasi sukarela yang tidak terikat nisab, haul, atau perhitungan fiqh tertentu.

```
infaq_config:
  minimum_amount: 0                   # Tidak ada minimum (even Rp 1.000 diterima)
  payment_modes:
    - one_time                        # Donasi sekali
    - recurring_monthly               # Auto-debit bulanan dari tabungan
    - recurring_weekly                 # Auto-debit mingguan
  designation_categories:             # Configurable per tenant
    - umum                            # Dana umum, dialokasikan oleh pengurus
    - pendidikan                      # Beasiswa, bantuan sekolah
    - masjid                          # Pembangunan/pemeliharaan masjid
    - yatim                           # Santunan anak yatim
    - kesehatan                       # Bantuan kesehatan
    - bencana                         # Tanggap bencana
    - custom_*                        # Tenant bisa tambah kategori sendiri
```

**Recurring Infaq:**

```
recurring_infaq_config:
  source_account: tabungan_id         # Rekening sumber (harus milik anggota)
  amount: 50000                       # Nominal per periode
  frequency: "monthly"                # monthly | weekly
  debit_day: 1                        # Tanggal debit (untuk monthly)
  designation: "pendidikan"           # Kategori tujuan
  start_date: "2026-05-01"
  end_date: null                      # null = berlaku terus sampai dibatalkan
  authorization_signed: true          # WAJIB ada otorisasi tertulis
  authorization_doc_url: "..."        # Scan dokumen otorisasi
```

**Aturan:**
- **Tidak ada minimum amount** — setiap nominal diterima
- Recurring infaq membutuhkan **otorisasi tertulis** (signed authorization) dari anggota
- Anggota bisa **membatalkan recurring** kapan saja — efektif mulai periode berikutnya
- Jika saldo tabungan tidak cukup saat auto-debit → skip periode tersebut, coba lagi periode berikutnya
- Setiap infaq tercatat dengan **designation** (tujuan) — memudahkan pelaporan dan penyaluran
- Designation categories **configurable per tenant** — Admin bisa tambah/hapus kategori

### 6. Collection Mechanism — Mekanisme Pengumpulan

```
PENGUMPULAN ZAKAT & INFAQ:

1. AUTO-DEBIT TABUNGAN
   ├── Anggota opt-in + signed authorization
   ├── Debit otomatis dari rekening tabungan
   ├── Untuk: zakat mal, zakat fitrah, recurring infaq
   └── Gagal jika saldo tidak cukup → notifikasi anggota

2. MANUAL VIA TELLER
   ├── Anggota datang ke kantor, bayar cash
   ├── Teller input: jenis (zakat mal/fitrah/infaq), nominal, muzakki
   ├── Untuk: semua jenis zakat & infaq
   └── Terima cash atau transfer dari rekening

3. BULK COLLECTION — ZAKAT FITRAH
   ├── Untuk konteks sekolah: pengumpulan per kelas/halaqah
   ├── Wali kelas/ustadz input daftar santri + jumlah per kepala
   ├── Sistem generate batch record
   ├── Batch approval oleh Supervisor+
   └── Periode: configurable (default: 1 Ramadan - 1 Syawal)
```

**Zakat Fitrah Collection Period:**

```
fitrah_period_config:
  hijri_start_month: 9               # Ramadan
  hijri_start_day: 1                  # 1 Ramadan
  hijri_end_month: 10                 # Syawal
  hijri_end_day: 1                    # 1 Syawal (sebelum shalat Eid)
  amount_per_person: 35000            # Configurable (setara ~2.5 kg beras)
  amount_unit: "IDR"
  year_hijri: 1448                    # Tahun hijriah aktif
```

**Aturan:**
- Auto-debit **hanya** dengan otorisasi tertulis — tidak ada auto-debit tanpa consent
- Bulk collection zakat fitrah membutuhkan **approval Supervisor+** sebelum diproses
- Periode zakat fitrah di-set per tahun hijriah oleh Admin — default 1 Ramadan sampai 1 Syawal
- Nominal zakat fitrah per kepala **configurable** — mengikuti harga beras setempat
- Teller bisa menerima zakat dari **non-anggota** (tamu) — dicatat sebagai muzakki eksternal

### 7. Muzakki & Mustahik Management

**Muzakki** (pembayar zakat) dan **mustahik** (penerima zakat) dikelola sebagai entity terpisah.

```
MUZAKKI (Pembayar):
├── Anggota internal → Link ke nasabah_id (FK)
├── Non-anggota (tamu) → Data tersimpan di zakat_collection langsung
└── Setiap pembayaran zakat tercatat: siapa bayar, berapa, kapan, jenis apa

MUSTAHIK (Penerima):
├── 8 Asnaf (kategori penerima sesuai Al-Quran At-Taubah: 60):
│   ├── 1. FAKIR     — Tidak memiliki harta & penghasilan sama sekali
│   ├── 2. MISKIN    — Penghasilan tidak mencukupi kebutuhan dasar
│   ├── 3. AMIL      — Pengelola zakat (staf baitul maal BMT)
│   ├── 4. MUALLAF   — Orang yang baru masuk Islam
│   ├── 5. RIQAB     — Memerdekakan budak (konteks modern: pembebasan hutang)
│   ├── 6. GHARIMIN  — Orang yang terlilit hutang untuk kebutuhan halal
│   ├── 7. FISABILILLAH — Di jalan Allah (dakwah, pendidikan Islam)
│   └── 8. IBNU SABIL   — Musafir yang kehabisan bekal
└── Mustahik BUKAN nasabah — entity terpisah, lebih sederhana
```

**Aturan:**
- Mustahik **bukan nasabah** — tidak perlu proses pendaftaran anggota, cukup data identitas dasar
- Mustahik bisa memiliki **multiple kategori** sekaligus (misal: fakir + gharimin)
- Registrasi mustahik oleh **Supervisor+** di divisi Baitul Maal
- Needs assessment dilakukan sebelum mustahik terdaftar — documented reason untuk setiap kategori
- Mustahik di-review secara **berkala** (minimal annual) — kondisi bisa berubah
- Jika mustahik kebetulan adalah anggota koperasi → `nasabah_ref_id` (optional, untuk referensi silang saja)

### 8. Distribution — Penyaluran Dana

Penyaluran zakat dan infaq menggunakan **multi-level approval** untuk memastikan akuntabilitas.

```
Distribution Flow:

Amil (staf baitul maal) membuat proposal
        │
        v
┌─────────────────────────────┐
│ PROPOSAL (Status: DRAFT)     │
│ - Mustahik: siapa penerima   │
│ - Amount: berapa              │
│ - Category: asnaf apa         │
│ - Purpose: untuk apa          │
│ - Method: cash/goods/transfer │
│ - Source fund: zakat/infaq    │
└────────┬────────────────────┘
         │ submit
         v
┌─────────────────────────────┐
│ REVIEW (Status: PENDING)     │
│ Manager review & approve     │
└────────┬────────────────────┘
         │ approve
         v
┌─────────────────────────────┐
│ FINAL APPROVAL               │
│ Admin final approve          │
│ (untuk nominal > threshold)  │
└────────┬────────────────────┘
         │ approve
         v
┌─────────────────────────────┐
│ EXECUTE (Status: APPROVED)   │
│ 1. Debit dana zakat/infaq    │
│ 2. Serahkan ke mustahik      │
│ 3. Catat bukti penerimaan    │
│ 4. Update saldo fund         │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│ COMPLETED                    │
│ - proof_of_receipt_url       │
│ - received_at                │
│ - witness (saksi, opsional)  │
└─────────────────────────────┘
```

**Distribution Methods:**
- **Cash** — penyerahan tunai dengan bukti tanda terima
- **Goods** — bantuan berupa barang (sembako, seragam, dll) — catat jenis & nilai barang
- **Transfer to account** — jika mustahik memiliki rekening di BMT → credit langsung

**Approval Threshold:**

```
distribution_approval_config:
  manager_max_amount: 5000000         # Manager bisa approve sampai Rp 5 juta
  above_manager_max: "admin"          # Di atas Rp 5 juta → Admin approval
  batch_distribution: "admin"         # Batch distribution selalu Admin approval
```

**Aturan:**
- Setiap distribusi **wajib memiliki bukti penerimaan** (proof of receipt) — foto/scan tanda terima
- Distribusi di bawah threshold → Manager cukup, di atas threshold → Admin wajib
- Batch distribution (ke banyak mustahik sekaligus) selalu membutuhkan **Admin approval**
- Amil **tidak bisa approve sendiri** — separation of duties
- Distribution priority per asnaf **configurable** — Admin bisa set urutan prioritas penyaluran
- Dana zakat **tidak boleh** dicampur dengan dana infaq dalam satu distribusi — harus terpisah

### 9. Ta'zir Fund Integration

Dana ta'zir dari denda keterlambatan ([ADR-K009](./ADR-K009-denda-penalti.md)) mengalir ke dana sosial dan dikelola bersama dengan zakat/infaq.

```
Alur Dana Ta'zir:

Denda keterlambatan (K009)
        │
        v
┌─────────────────────────────┐
│ Fund destination check:      │
│ coop_type = "islamic"        │
│ penalty_type = "tazir"       │
│ → destination: social_fund   │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│ Jurnal akuntansi:            │
│ Debit:  Kas / Rek. Nasabah   │
│ Credit: Dana Ta'zir (8100)   │
│                              │
│ COA: 8101 Denda Keterlambatan│
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│ Masuk ke saldo dana ta'zir   │
│ Dikelola BERSAMA infaq       │
│ (bukan bersama zakat)        │
│                              │
│ Penyaluran via approval flow │
│ yang sama (Section 8)        │
│ tapi dengan tracking terpisah│
└─────────────────────────────┘
```

**Aturan:**
- Dana ta'zir **bukan zakat** — tidak masuk ke COA 6xxx, melainkan COA 8xxx
- Ta'zir dikelola **bersama infaq/shadaqah** dari sisi penyaluran, tapi **tracking terpisah** di laporan
- Approval flow penyaluran ta'zir sama dengan zakat/infaq (Section 8)
- Penyaluran ta'zir tidak terikat 8 asnaf — bisa untuk kegiatan sosial umum
- Laporan ta'zir **wajib terpisah** — DPS (Dewan Pengawas Syariah) perlu audit spesifik

### 10. Accounting — Integrasi Akuntansi

Seluruh transaksi zakat, infaq, dan ta'zir menggunakan **COA 6xxx dan 8xxx** yang bersifat **off-balance sheet** terhadap P&L koperasi.

```
JURNAL PENERIMAAN ZAKAT MAL:
  Debit:  1101 Kas Teller           Rp 3.187.500
  Credit: 6101 Zakat Maal Anggota   Rp 3.187.500

JURNAL PENERIMAAN ZAKAT FITRAH:
  Debit:  1101 Kas Teller           Rp 35.000
  Credit: 6102 Zakat Fitrah         Rp 35.000

JURNAL PENERIMAAN ZAKAT INSTITUSI:
  Debit:  3500 SHU Tahun Berjalan   Rp 25.000.000
  Credit: 6103 Zakat Institusi       Rp 25.000.000

JURNAL PENYALURAN ZAKAT:
  Debit:  6201 Penyaluran ke Fakir   Rp 1.000.000
  Credit: 1101 Kas Teller            Rp 1.000.000

JURNAL PENERIMAAN INFAQ:
  Debit:  1101 Kas Teller            Rp 50.000
  Credit: 6300 Penerimaan Infaq      Rp 50.000

JURNAL PENYALURAN INFAQ:
  Debit:  6400 Penyaluran Infaq      Rp 500.000
  Credit: 1101 Kas Teller            Rp 500.000

JURNAL PENERIMAAN TA'ZIR (dari K009):
  Debit:  1101 Kas Teller            Rp 5.000
  Credit: 8101 Denda Keterlambatan   Rp 5.000
```

**Penyajian di Neraca per PSAK 109:**

Dana zakat, infaq, dan ta'zir BUKAN murni off-balance sheet. Per PSAK 109 paragraf 35-36, saldo dana ini disajikan sebagai **bagian terpisah di Neraca** (Laporan Posisi Keuangan):

```
Neraca BMT:
  ASET
  ├── 1xxx Aset Lancar & Tetap
  │
  KEWAJIBAN
  ├── 2xxx Simpanan Nasabah, Hutang
  │
  EKUITAS
  ├── 3xxx Simpanan Pokok, Cadangan, SHU
  │
  DANA ZAKAT, INFAQ & TA'ZIR (section terpisah — bukan kewajiban, bukan ekuitas)
  ├── 6xxx Dana Zakat (saldo belum disalurkan)
  ├── 6xxx Dana Infaq/Shadaqah (saldo belum disalurkan)
  └── 8xxx Dana Ta'zir (saldo belum disalurkan)

- Total Neraca = Aset = Kewajiban + Ekuitas + Dana ZIS
- Dana ZIS adalah amanah yang dititipkan, bukan milik koperasi
- General mode: section ini TIDAK ditampilkan (tidak ada dana ZIS)
```

**Aturan:**
- Dana zakat, infaq, ta'zir **tidak masuk** laporan laba rugi koperasi — terpisah dari P&L
- Dana ZIS disajikan sebagai **section terpisah di Neraca** per PSAK 109 (bukan kewajiban, bukan ekuitas)
- Laporan terpisah: **Laporan Sumber & Penyaluran Dana Zakat** (PSAK 109)
- Laporan terpisah: **Laporan Sumber & Penggunaan Dana Kebajikan** (termasuk infaq & ta'zir)
- Auto-journal: setiap collection/distribution otomatis generate jurnal — operator tidak perlu input manual
- Audit trail: setiap jurnal link ke `zakat_collection_id` atau `zakat_distribution_id` sebagai source document

### 11. School Context — Konteks Sekolah

Modul zakat memiliki fitur khusus untuk konteks sekolah/pesantren.

```
ZAKAT FITRAH SANTRI:
├── Pengumpulan per kelas/halaqah oleh wali kelas/ustadz
├── Input bulk: daftar santri + jumlah keluarga per santri
├── Nominal per kepala: sesuai fitrah_period_config
├── Wali santri bisa bayar via auto-debit tabungan (jika opt-in)
├── Batch record di-generate per kelas → approval Supervisor+
└── Laporan: per kelas, per angkatan, total sekolah

DISTRIBUSI KE SANTRI/KELUARGA:
├── Santri/keluarga tidak mampu sebagai mustahik (fakir/miskin)
├── Distribusi: beasiswa, bantuan SPP, seragam, buku
├── Needs assessment oleh wali kelas + verifikasi Supervisor
├── Distribusi dalam bentuk barang/voucher (bukan cash ke santri)
└── Laporan terpisah untuk yayasan dan orang tua

ASPEK EDUKASI:
├── Modul transparansi: santri bisa lihat ringkasan pengumpulan kelas
├── Bukan data individual — hanya total per kelas
├── Tujuan: mengajarkan kepedulian sosial & kewajiban zakat
└── Fitur opsional, diaktifkan per tenant
```

**Aturan:**
- Guru/ustadz yang juga anggota koperasi bisa **opt-in zakat mal auto-calculation** (Section 4)
- Pengumpulan zakat fitrah santri per kelas membutuhkan **approval Supervisor+**
- Distribusi ke santri dalam bentuk **barang/voucher** (bukan cash) — kontrol penggunaan
- Transparansi data hanya **agregat per kelas** — tidak ada data individual yang di-expose ke santri
- Laporan zakat fitrah sekolah tersedia untuk **yayasan** dan presentasi di RAT

### 12. Reporting — Laporan

```
LAPORAN PENERIMAAN ZAKAT:
├── Per jenis: zakat mal, zakat fitrah, zakat institusi
├── Per periode: bulanan, triwulan, tahunan
├── Per muzakki: detail siapa bayar berapa
├── Per branch: jika multi-cabang
└── Format: Laporan Sumber Dana Zakat (PSAK 109)

LAPORAN PENYALURAN ZAKAT:
├── Per kategori mustahik (8 asnaf)
├── Per mustahik: detail siapa terima berapa
├── Per periode: bulanan, triwulan, tahunan
├── Per tujuan: beasiswa, bantuan, dll
└── Format: Laporan Penyaluran Dana Zakat (PSAK 109)

LAPORAN SALDO DANA:
├── Saldo dana zakat (penerimaan - penyaluran)
├── Saldo dana infaq/shadaqah
├── Saldo dana ta'zir
├── Saldo gabungan dana sosial
└── Per branch dan konsolidasi

LAPORAN TAHUNAN UNTUK RAT:
├── Ringkasan sumber & penyaluran dana zakat setahun
├── Perbandingan antar tahun
├── Jumlah muzakki dan mustahik yang dilayani
├── Efektivitas penyaluran: % dana yang tersalurkan
└── Rekomendasi pengurus untuk tahun depan

LAPORAN KONTEKS SEKOLAH:
├── Zakat fitrah per kelas/angkatan
├── Distribusi beasiswa dari dana zakat
├── Statistik santri penerima bantuan
└── Laporan untuk yayasan
```

### 13. Data Model

```
zakat_collection
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
│
├── ── Muzakki ──
├── nasabah_id            UUID (nullable, FK → nasabah, jika anggota)
├── muzakki_name          VARCHAR NOT NULL (nama pembayar, diisi meskipun ada nasabah_id)
├── muzakki_phone         VARCHAR (nullable)
├── muzakki_address       TEXT (nullable)
│
├── ── Zakat Detail ──
├── zakat_type            ENUM (zakat_mal, zakat_fitrah, zakat_institusi)
├── amount                NUMERIC(15,2) NOT NULL
├── payment_method        ENUM (cash, auto_debit, transfer)
├── source_account_id     UUID (nullable, FK → rekening, jika auto-debit/transfer)
│
├── ── Zakat Fitrah Specific ──
├── fitrah_head_count     INTEGER (nullable, jumlah jiwa untuk fitrah)
├── fitrah_amount_per_head NUMERIC(15,2) (nullable)
├── fitrah_year_hijri     INTEGER (nullable, tahun hijriah)
│
├── ── Zakat Mal Specific ──
├── qualifying_assets     NUMERIC(15,2) (nullable, total aset qualifying)
├── debts_deducted        NUMERIC(15,2) (nullable, hutang yang dikurangi)
├── nisab_at_calculation  NUMERIC(15,2) (nullable, nisab saat kalkulasi)
├── haul_start_date       DATE (nullable, tanggal mulai haul)
├── haul_end_date         DATE (nullable, tanggal haul tercapai)
│
├── ── Batch (untuk bulk collection) ──
├── batch_id              UUID (nullable, FK → zakat_collection_batch)
│
├── ── Jurnal ──
├── journal_entry_id      UUID (nullable, FK → journal_entry)
│
├── ── Status ──
├── status                ENUM (pending, confirmed, cancelled)
├── confirmed_at          TIMESTAMPTZ (nullable)
├── confirmed_by          UUID (nullable, FK → user)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID


zakat_distribution
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
│
├── ── Mustahik ──
├── mustahik_id           UUID (FK → mustahik) NOT NULL
│
├── ── Distribution Detail ──
├── source_fund           ENUM (zakat, infaq, tazir)
├── asnaf_category        ENUM (fakir, miskin, amil, muallaf, riqab, gharimin, fisabilillah, ibnu_sabil)
│                         (nullable jika source_fund != zakat)
├── amount                NUMERIC(15,2) NOT NULL
├── purpose               TEXT NOT NULL (tujuan distribusi)
├── distribution_method   ENUM (cash, goods, account_transfer)
│
├── ── Goods Detail (jika method = goods) ──
├── goods_description     TEXT (nullable, deskripsi barang)
├── goods_estimated_value NUMERIC(15,2) (nullable, estimasi nilai barang)
│
├── ── Transfer Detail (jika method = account_transfer) ──
├── target_account_id     UUID (nullable, FK → rekening mustahik)
│
├── ── Proof ──
├── proof_of_receipt_url  VARCHAR (nullable, foto/scan bukti terima)
├── received_at           TIMESTAMPTZ (nullable)
├── witness_name          VARCHAR (nullable, nama saksi)
│
├── ── Approval ──
├── status                ENUM (draft, pending, approved, rejected, completed)
├── submitted_at          TIMESTAMPTZ (nullable)
├── approved_by           UUID (nullable, FK → user)
├── approved_at           TIMESTAMPTZ (nullable)
├── final_approved_by     UUID (nullable, FK → user, untuk di atas threshold)
├── final_approved_at     TIMESTAMPTZ (nullable)
├── rejection_reason      TEXT (nullable)
│
├── ── Jurnal ──
├── journal_entry_id      UUID (nullable, FK → journal_entry)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID


infaq
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
│
├── ── Donatur ──
├── nasabah_id            UUID (nullable, FK → nasabah)
├── donatur_name          VARCHAR NOT NULL
├── donatur_phone         VARCHAR (nullable)
│
├── ── Infaq Detail ──
├── infaq_type            ENUM (one_time, recurring)
├── amount                NUMERIC(15,2) NOT NULL
├── designation           VARCHAR NOT NULL (kategori tujuan: umum, pendidikan, dst)
├── payment_method        ENUM (cash, auto_debit, transfer)
├── source_account_id     UUID (nullable, FK → rekening)
│
├── ── Recurring Config ──
├── recurring_config_id   UUID (nullable, FK → infaq_recurring_config)
│
├── ── Jurnal ──
├── journal_entry_id      UUID (nullable, FK → journal_entry)
│
├── ── Status ──
├── status                ENUM (pending, confirmed, cancelled)
├── confirmed_at          TIMESTAMPTZ (nullable)
├── confirmed_by          UUID (nullable, FK → user)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID


infaq_recurring_config
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah) NOT NULL
├── source_account_id     UUID (FK → rekening) NOT NULL
├── amount                NUMERIC(15,2) NOT NULL
├── frequency             ENUM (weekly, monthly)
├── debit_day             INTEGER (1-31 untuk monthly, 1-7 untuk weekly)
├── designation           VARCHAR NOT NULL
├── start_date            DATE NOT NULL
├── end_date              DATE (nullable, null = berlaku selamanya)
├── authorization_doc_url VARCHAR NOT NULL (scan otorisasi tertulis)
├── is_active             BOOLEAN DEFAULT true
├── last_debit_date       DATE (nullable)
├── next_debit_date       DATE NOT NULL
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID


mustahik
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
│
├── ── Identitas ──
├── full_name             VARCHAR NOT NULL
├── identity_number       VARCHAR (nullable, KTP jika ada)
├── phone                 VARCHAR (nullable)
├── address               TEXT
├── photo_url             VARCHAR (nullable)
│
├── ── Kategori ──
├── asnaf_categories      VARCHAR[] NOT NULL (array, bisa multi-kategori)
│   # Values: fakir, miskin, amil, muallaf, riqab, gharimin, fisabilillah, ibnu_sabil
├── primary_category      VARCHAR NOT NULL (kategori utama)
│
├── ── Assessment ──
├── needs_assessment      TEXT NOT NULL (deskripsi kebutuhan)
├── assessment_date       DATE NOT NULL
├── assessed_by           UUID (FK → user)
├── next_review_date      DATE NOT NULL (jadwal review ulang)
│
├── ── Referensi Nasabah ──
├── nasabah_ref_id        UUID (nullable, FK → nasabah, jika kebetulan anggota)
│
├── ── Status ──
├── status                ENUM (active, inactive, graduated)
│   # active: masih menerima bantuan
│   # inactive: sementara tidak aktif
│   # graduated: sudah tidak memenuhi kriteria mustahik (mandiri)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID


tazir_fund
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
│
├── ── Source ──
├── denda_id              UUID (FK → denda) NOT NULL
├── pinjaman_id           UUID (FK → pinjaman) NOT NULL
├── nasabah_id            UUID (FK → nasabah) NOT NULL
│
├── ── Amount ──
├── amount                NUMERIC(15,2) NOT NULL
│
├── ── Jurnal ──
├── journal_entry_id      UUID (nullable, FK → journal_entry)
│
├── ── Status ──
├── status                ENUM (collected, distributed)
├── distribution_id       UUID (nullable, FK → zakat_distribution, jika sudah disalurkan)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID


zakat_collection_batch
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
│
├── ── Batch Info ──
├── batch_type            ENUM (fitrah_class, fitrah_bulk, general)
├── description           TEXT NOT NULL (misal: "Zakat Fitrah Kelas 7A - 1448H")
├── total_records         INTEGER NOT NULL
├── total_amount          NUMERIC(15,2) NOT NULL
│
├── ── Approval ──
├── status                ENUM (draft, pending, approved, rejected)
├── approved_by           UUID (nullable, FK → user)
├── approved_at           TIMESTAMPTZ (nullable)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Index penting:**
- `(tenant_id, zakat_type, status)` — query collection per jenis
- `(tenant_id, nasabah_id)` — riwayat zakat per anggota
- `(tenant_id, mustahik_id, status)` — distribusi per mustahik
- `(tenant_id, source_fund, status)` — laporan per sumber dana
- `(tenant_id, branch_id, created_at)` — laporan per cabang per periode
- `(batch_id)` — query semua record dalam satu batch

### 14. Vernon _rels dan _data Structure

**zakat_collection _rels:**
```json
{
  "tenant_id":  "018f...",
  "branch_id":  "018f...",
  "nasabah_id": "018f...",
  "batch_id":   "018f..."
}
```

**zakat_collection _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "branch": {
    "id":   "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  }
}
```

**zakat_distribution _data:**
```json
{
  "mustahik": {
    "id":               "018f...",
    "full_name":        "Siti Aminah",
    "primary_category": "fakir"
  },
  "branch": {
    "id":   "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.nasabah` di semua zakat_collection nasabah tersebut
- `BranchUpdatedEvent` → update `_data.branch` di semua zakat_collection & zakat_distribution cabang tersebut
- `MustahikUpdatedEvent` → update `_data.mustahik` di semua zakat_distribution mustahik tersebut

### 15. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| Terima zakat/infaq (collection) | v | v | v | v |
| View laporan zakat/infaq | v | v | v | v |
| Approve batch collection (fitrah) | - | v | v | v |
| Buat proposal distribusi | - | v | v | v |
| Approve distribusi (< threshold) | - | - | v | v |
| Approve distribusi (> threshold) | - | - | - | v |
| Approve batch distribution | - | - | - | v |
| Registrasi mustahik | - | v | v | v |
| Edit/deactivate mustahik | - | v | v | v |
| Configure nisab/fitrah amount | - | - | - | v |
| Configure designation categories | - | - | - | v |
| Manage ta'zir fund | - | - | v | v |
| View laporan dana sosial | - | v | v | v |
| Override fund allocation | - | - | - | v |
| Export laporan untuk RAT/DPS | - | - | v | v |

**Catatan:**
- **Teller** bisa terima pembayaran zakat/infaq — operasi rutin kasir
- **Supervisor** bisa registrasi mustahik dan buat proposal distribusi — divisi Baitul Maal
- **Manager** approve distribusi di bawah threshold dan manage ta'zir fund
- **Admin** approve distribusi besar, batch, dan konfigurasi global (nisab, fitrah amount)
- Amil (pengelola zakat) umumnya memiliki role **Supervisor** atau **Manager** di sistem

## Consequences

### Positif

- **Syariah compliant** — pengelolaan zakat sesuai fiqh (nisab, haul, 8 asnaf) dan PSAK 109
- **Transparan** — setiap collection dan distribution tercatat dengan bukti, approval, dan audit trail
- **Off-balance sheet** — dana zakat/infaq tidak mencemari laporan laba rugi koperasi
- **Flexible** — nisab, fitrah amount, designation categories, dan approval threshold configurable
- **School-ready** — bulk collection zakat fitrah per kelas, distribusi beasiswa ke santri
- **Integrated** — ta'zir fund dari K009 dikelola dalam satu dashboard dengan zakat/infaq
- **Auditable** — DPS (Dewan Pengawas Syariah) bisa audit seluruh alur dana sosial
- **Multi-approval** — separation of duties mencegah penyalahgunaan dana umat

### Negatif

- **Islamic-only module** — tidak memberikan nilai untuk tenant mode general
- **Hijriah calendar complexity** — tracking haul dan periode zakat fitrah menggunakan kalender hijriah
- **Mustahik management overhead** — registrasi, assessment, dan review berkala membutuhkan effort
- **Dual accounting** — laporan off-balance sheet terpisah menambah kompleksitas akuntansi
- **Auto-calculation limitation** — hanya aset di koperasi yang dihitung, aset luar tidak tercakup

### Mitigasi

- Module completely hidden di mode general — zero overhead untuk tenant konvensional
- Library konversi tanggal Masehi-Hijriah digunakan — tidak perlu kalkulasi manual
- Mustahik review di-schedule otomatis dengan reminder — Supervisor mendapat notifikasi
- Auto-journal memastikan entri akuntansi off-balance sheet otomatis — operator tidak perlu paham akuntansi syariah
- Kalkulasi zakat disajikan sebagai **rekomendasi** — anggota bisa adjust untuk aset di luar koperasi

## Alternatives Considered

### A. Zakat sebagai Modul Terpisah (bukan bagian Koperasi)

Memisahkan pengelolaan zakat sebagai aplikasi standalone yang tidak terintegrasi dengan koperasi.

**Ditolak** karena: BMT secara definisi memiliki fungsi Baitul Maal (pengelolaan dana sosial). Memisahkan modul ini berarti kehilangan integrasi data nasabah, kalkulasi aset otomatis, auto-debit dari tabungan, dan integrasi ta'zir fund. Justru integrasi inilah yang menjadi nilai tambah BMT dibandingkan lembaga zakat independen.

### B. Zakat Aktif di Semua Mode (General + Islamic)

Mengaktifkan modul zakat untuk semua tenant, termasuk mode general/konvensional.

**Ditolak** karena: koperasi konvensional tidak memiliki mandat pengelolaan zakat. Mengaktifkan modul ini di mode general akan membingungkan dan menambah kompleksitas tanpa memberikan nilai. Feature flag derived dari `coop_type` memastikan clean separation sesuai [ADR-009](../core/ADR-009-dual-mode-institution-type.md).

### C. Mustahik sebagai Nasabah

Menjadikan mustahik sebagai nasabah biasa dengan flag khusus, bukan entity terpisah.

**Ditolak** karena: mustahik dan nasabah memiliki lifecycle dan data requirement yang sangat berbeda. Nasabah membutuhkan KYC, ahli waris, approval flow pendaftaran (K001). Mustahik hanya butuh identitas dasar dan needs assessment. Menggabungkan keduanya akan over-complicate model nasabah dan memaksa mustahik melewati proses pendaftaran yang tidak relevan. Referensi silang via `nasabah_ref_id` cukup untuk kasus di mana mustahik kebetulan juga anggota.
