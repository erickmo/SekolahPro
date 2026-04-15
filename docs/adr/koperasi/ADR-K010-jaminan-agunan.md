# ADR-K010: Jaminan / Agunan (Collateral Management)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Jaminan/agunan adalah komponen penting dalam manajemen risiko pinjaman koperasi/BMT. Setiap pinjaman di atas threshold tertentu memerlukan jaminan sebagai mitigasi risiko gagal bayar. Sistem harus mengakomodasi:

- **Multi-type collateral**: saldo simpanan (internal), surat berharga (BPKB, sertifikat), barang bergerak, dan personal guarantee
- **Internal collateral**: saldo tabungan/deposito yang di-hold melalui mekanisme `hold_amount` di [ADR-K002](./ADR-K002-rekening.md)
- **Valuation (taksasi)**: penilaian nilai jaminan dengan LTV ratio yang configurable per produk
- **Lifecycle management**: dari pendaftaran hingga release/foreclosure
- **Dual-mode**: Jaminan/Agunan (konvensional) vs Rahn/Marhun (BMT) — lihat [ADR-009](../core/ADR-009-dual-mode-institution-type.md)
- **School context**: mayoritas pinjaman guru/staf berskala kecil (personal guarantee cukup), pinjaman besar memerlukan BPKB atau sertifikat
- **Penjamin (Guarantor)**: model penjamin sudah didefinisikan di [ADR-K007](./ADR-K007-pinjaman.md) Section 4 — ADR ini fokus pada **jaminan fisik dan saldo**, serta memperluas model penjamin

## Decision

### 1. Jenis Jaminan (Collateral Types)

```
collateral_type:
├── INTERNAL_BALANCE    Saldo simpanan (tabungan/deposito) sebagai jaminan
│                       Auto-hold via K002 hold_amount mechanism
├── SURAT_BERHARGA      Dokumen berharga: BPKB kendaraan, sertifikat tanah/bangunan,
│                       ijazah (khusus pinjaman darurat)
├── BARANG_BERGERAK     Aset fisik: kendaraan, elektronik, perhiasan
└── PERSONAL_GUARANTEE  Penjaminan oleh nasabah lain atau pihak eksternal
                        (detail model di K007 Section 4, diperluas di Section 10)
```

**Aturan per jenis:**

| Jenis | Valuation | Penyimpanan | LTV Default | Auto-release |
|-------|-----------|-------------|-------------|--------------|
| INTERNAL_BALANCE | 100% saldo (otomatis) | Di rekening nasabah | 100% | Ya, saat pinjaman COMPLETED |
| SURAT_BERHARGA | Manual oleh Supervisor+ | Brankas koperasi | 70-80% | Tidak, butuh release process |
| BARANG_BERGERAK | Manual oleh Supervisor+ | Gudang koperasi / di nasabah | 50-70% | Tidak, butuh release process |
| PERSONAL_GUARANTEE | Nominal yang disepakati | N/A (surat pernyataan) | 100% | Ya, saat pinjaman COMPLETED |

### 2. Valuation — Taksasi / Penilaian Jaminan

**Konfigurasi per produk pinjaman:**

```
collateral_valuation_config:
  ltv_ratio: 0.80                    # Loan-to-Value max 80% (default)
  # Artinya: plafon pinjaman max = 80% dari total nilai jaminan

  valuation_methods:
    INTERNAL_BALANCE:
      method: "auto"                 # Otomatis = 100% saldo saat pledging
      revaluation: "realtime"        # Selalu up-to-date dari saldo rekening

    SURAT_BERHARGA:
      method: "manual"               # Supervisor+ melakukan taksasi
      requires_role: "supervisor"    # Minimal Supervisor untuk taksasi
      revaluation_period_months: 12  # Wajib taksasi ulang setiap 12 bulan

    BARANG_BERGERAK:
      method: "manual"
      requires_role: "supervisor"
      depreciation_rate: 0.10        # 10% per tahun (configurable per item type)
      revaluation_period_months: 6   # Taksasi ulang setiap 6 bulan

    PERSONAL_GUARANTEE:
      method: "declared"             # Nominal yang disepakati bersama penjamin
      revaluation: "none"            # Tidak ada revaluasi
```

**Formula Nilai Jaminan (Satu Haircut, Bukan Dua):**

Nilai jaminan yang diakui untuk penjaminan pinjaman:

```
collateral_value = appraised_value × acceptance_rate
```

- `appraised_value`: Nilai taksasi/appraisal dari penilai
- `acceptance_rate`: Haircut berdasarkan jenis jaminan (sudah memperhitungkan LTV)
- `collateral_value`: Nilai yang diakui untuk coverage

LTV ratio di produk (K003) digunakan untuk MENENTUKAN acceptance_rate default per jenis jaminan, BUKAN sebagai pengali tambahan.

| Jenis Jaminan | Default Acceptance Rate | Penjelasan |
|--------------|------------------------|------------|
| INTERNAL_BALANCE (tabungan/deposito) | 100% | Likuid, tidak ada risiko penurunan nilai |
| SURAT_BERHARGA (BPKB, sertifikat) | 70-80% | LTV standar untuk surat berharga |
| BARANG_BERGERAK (kendaraan, elektronik) | 50-70% | Risiko depresiasi |
| PERSONAL_GUARANTEE | 0% (moral) | Tidak memiliki nilai moneter yang diakui |

```
Contoh: BPKB motor appraised Rp 15.000.000, acceptance_rate 70%
→ collateral_value = 15.000.000 × 70% = Rp 10.500.000
→ Pinjaman maksimal yang bisa dijamin: Rp 10.500.000 (TANPA pengali LTV tambahan)

Eligibility check: total_outstanding_loan ≤ SUM(collateral_value) dari semua jaminan yang di-pledge
```

**Revaluasi:**
- Revaluasi periodik otomatis di-schedule sesuai `revaluation_period_months`
- Revaluasi on-demand bisa diminta oleh Manager+ kapan saja
- Internal balance: revaluasi real-time dari saldo aktual rekening
- Jika nilai jaminan turun di bawah coverage ratio minimum → alert ke Manager
- Hasil revaluasi tersimpan di history (versi lama tidak dihapus)

### 3. Collateral Binding — Pengikatan Jaminan ke Pinjaman

**Konfigurasi:**

```
collateral_binding_config:
  allow_multi_loan: false            # Default: 1 jaminan = 1 pinjaman
  # Jika true: satu jaminan bisa menjamin beberapa pinjaman sekaligus

  coverage_ratio_config:
    minimum_coverage: 1.00           # Total collateral_value >= 100% outstanding loan
    warning_threshold: 1.20          # Alert jika coverage turun di bawah 120%
```

**Mekanisme binding:**

```
Satu Jaminan ──── menjaminkan ────> Satu Pinjaman (default)
                                    atau
Satu Jaminan ──── menjaminkan ────> Beberapa Pinjaman (jika allow_multi_loan = true)

Satu Pinjaman <── dijaminkan oleh ── Beberapa Jaminan (selalu diperbolehkan)
```

**Coverage ratio tracking:**

```
coverage_ratio = SUM(collateral_value semua jaminan terikat) / outstanding_loan_balance

Contoh:
  Pinjaman outstanding: 20.000.000
  Jaminan 1 (deposito): collateral_value = 10.000.000
  Jaminan 2 (BPKB):    collateral_value = 15.000.000
  Total collateral:     25.000.000

  Coverage ratio = 25.000.000 / 20.000.000 = 1.25 (125%)
  → Di atas minimum (100%) ✓
  → Di atas warning (120%) ✓
```

**Aturan:**
- Binding dilakukan saat pinjaman di-approve atau saat pencairan (configurable)
- Jaminan yang sudah di-bind **tidak bisa dilepas** selama pinjaman masih ACTIVE
- Jika `allow_multi_loan = true`: sisa nilai jaminan setelah dikurangi pinjaman pertama bisa menjamin pinjaman berikutnya
- Coverage ratio dihitung ulang setiap ada pembayaran angsuran (outstanding turun) atau revaluasi jaminan
- Alert otomatis ke Manager jika coverage ratio turun di bawah `warning_threshold`

### 4. Internal Collateral — Saldo Simpanan sebagai Jaminan

Jaminan internal menggunakan mekanisme `hold_amount` di rekening tabungan/deposito ([ADR-K002](./ADR-K002-rekening.md) Section 11).

**Flow pledging saldo:**

```
Pinjaman di-approve dengan jaminan internal
        |
        v
+-----------------------------------+
| 1. Identifikasi rekening nasabah  |
|    yang akan dijaminkan            |
|    (tabungan dan/atau deposito)    |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 2. Validasi:                      |
|    - Rekening status = ACTIVE     |
|    - available_balance >= hold    |
|    - Rekening milik nasabah ybs   |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 3. Update rekening:               |
|    hold_amount += collateral_hold |
|    available_balance -= collateral|
+-----------------------------------+
        |
        v
+-----------------------------------+
| 4. Create record jaminan          |
|    type = INTERNAL_BALANCE        |
|    status = PLEDGED               |
|    linked ke rekening_id          |
+-----------------------------------+
```

**Aturan khusus internal collateral:**
- Hold amount di rekening = porsi saldo yang dijaminkan untuk pinjaman tersebut
- Nasabah **tidak bisa menarik** saldo yang di-hold (available_balance berkurang, balance tetap)
- Auto-release hold saat pinjaman COMPLETED — `hold_amount -= collateral_hold`
- Jika deposito jatuh tempo sementara masih di-pledged: **auto-rollover dipaksakan** — nasabah tidak bisa mencairkan
- Jika saldo tabungan berkurang (karena admin fee dormant, dll) sampai di bawah hold_amount → alert ke Manager
- Satu rekening bisa menjaminkan **sebagian** saldonya — tidak harus seluruh saldo

**Deposito sebagai jaminan — aturan tambahan:**

```
deposito_collateral_rules:
  auto_rollover_enforced: true       # Deposito yang di-pledge WAJIB rollover
  maturity_alert_days: 30            # Alert 30 hari sebelum jatuh tempo
  early_withdrawal_blocked: true     # Pencairan dini diblokir selama di-pledge
```

### 5. Document Management — Dokumen Jaminan

Setiap jaminan fisik (SURAT_BERHARGA, BARANG_BERGERAK) memerlukan dokumentasi lengkap.

**Tipe dokumen:**

```
document_type:
├── FOTO_BARANG          Foto fisik barang jaminan (wajib untuk BARANG_BERGERAK)
├── FOTOKOPI_BPKB        Fotokopi BPKB kendaraan
├── SERTIFIKAT_TANAH     Fotokopi sertifikat tanah/bangunan
├── SURAT_KUASA          Surat kuasa penyerahan jaminan
├── BERITA_ACARA         Berita acara serah terima jaminan
├── FOTO_IDENTITAS       Foto KTP pemilik jaminan
├── SURAT_PERNYATAAN     Surat pernyataan kepemilikan
├── IJAZAH               Fotokopi ijazah (untuk pinjaman darurat)
└── LAINNYA              Dokumen pendukung lainnya
```

**Aturan dokumen:**
- Minimal 1 foto barang wajib untuk BARANG_BERGERAK
- Minimal 1 dokumen kepemilikan wajib untuk SURAT_BERHARGA
- Upload file: foto (JPG, PNG), scan dokumen (PDF) — max size configurable (default: 5MB per file)
- **Document versioning**: re-upload menghasilkan versi baru, versi lama **tetap tersimpan** (tidak dihapus)
- Setiap dokumen dicatat `uploaded_by` dan `uploaded_at` untuk audit trail
- Dokumen tidak bisa dihapus — hanya bisa ditandai `is_superseded = true` saat ada versi baru

### 6. Status Lifecycle — Siklus Hidup Jaminan

```
+──────────────+
│  REGISTERED  │  Data jaminan diinput, belum diikat ke pinjaman
+──────┬───────+
       │ bind to loan (saat approval/pencairan)
       v
+──────────────+
│   PLEDGED    │  Terikat ke pinjaman aktif, tidak bisa dilepas
+──────┬───────+
       │
  +────┴────────────+
  │                 │
  v                 v
+──────────+  +─────────────+
│ RELEASED │  │ FORECLOSED  │
+──────────+  +─────────────+
```

**Status definitions:**

| Status | Deskripsi | Trigger |
|--------|-----------|---------|
| REGISTERED | Jaminan terdaftar di sistem, data dan dokumen lengkap, belum diikat ke pinjaman | Teller/Supervisor input data jaminan |
| PLEDGED | Jaminan terikat ke pinjaman aktif, tidak bisa dilepas atau dipindahkan | Pinjaman di-approve/dicairkan |
| RELEASED | Jaminan dilepas, dikembalikan ke pemilik | Pinjaman COMPLETED atau Manager release manual |
| FORECLOSED | Jaminan dieksekusi/dilikuidasi karena pinjaman write-off | Admin approve foreclosure |

**Aturan transisi:**
- `REGISTERED → PLEDGED`: otomatis saat pinjaman di-approve/dicairkan, atau manual oleh Supervisor+
- `PLEDGED → RELEASED`: otomatis saat semua pinjaman terkait COMPLETED, atau manual oleh Manager+ (dengan validasi)
- `PLEDGED → FORECLOSED`: hanya jika pinjaman write-off (Kol 5/Macet), membutuhkan approval Admin
- `RELEASED → REGISTERED`: jaminan yang sudah dilepas bisa didaftarkan ulang untuk pinjaman baru
- `FORECLOSED → *`: terminal state — tidak bisa dikembalikan

### 7. Release Process — Pelepasan Jaminan

**Auto-release (pinjaman COMPLETED):**

```
PinjamanCompletedEvent
        |
        v
+-----------------------------------+
| 1. Query semua jaminan terikat    |
|    ke pinjaman tersebut           |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 2. Per jaminan:                   |
|    - Cek apakah jaminan juga      |
|      menjamin pinjaman lain       |
|      yang masih aktif             |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 3a. Jika TIDAK ada pinjaman lain: |
|    - status → RELEASED            |
|    - released_at = NOW()          |
|    - [INTERNAL] release hold      |
|                                   |
| 3b. Jika ADA pinjaman lain aktif: |
|    - status tetap PLEDGED         |
|    - hanya unbind dari pinjaman   |
|      yang completed               |
+-----------------------------------+
```

**Release untuk internal collateral (saldo):**

```
Jaminan internal di-release:
  1. rekening.hold_amount -= collateral_hold_amount
  2. rekening.available_balance += collateral_hold_amount
  3. jaminan.status = RELEASED
  4. Semua dalam satu database transaction
```

**Release untuk physical collateral:**

```
Pinjaman COMPLETED
        |
        v
+-----------------------------------+
| 1. Manager generate surat         |
|    pelepasan jaminan (release doc)|
+-----------------------------------+
        |
        v
+-----------------------------------+
| 2. Nasabah datang ke kantor       |
|    - Verifikasi identitas         |
|    - Tanda tangan berita acara    |
|      serah terima                 |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 3. Supervisor menyerahkan         |
|    dokumen/barang jaminan         |
|    - Upload foto serah terima     |
|    - Status → RELEASED            |
+-----------------------------------+
```

**Release checklist (validasi sebelum release):**

```
Release Pre-check:
├── Semua pinjaman terkait COMPLETED?          ✓
├── Tidak ada klaim/sengketa pending?          ✓
├── Tidak ada denda outstanding (K009)?        ✓
├── Berita acara serah terima di-upload?       ✓ (physical only)
└── Identitas penerima terverifikasi?          ✓ (physical only)
    │
    ALL PASS → Release allowed
    ANY FAIL → Release blocked, return blocking items
```

### 8. Foreclosure Process — Eksekusi Jaminan

Eksekusi jaminan hanya terjadi saat pinjaman sudah di-write-off (Kol 5/Macet, lihat [ADR-K007](./ADR-K007-pinjaman.md)).

**Flow:**

```
Pinjaman status = WRITTEN_OFF (Kol 5)
        |
        v
+-----------------------------------+
| 1. Manager mengajukan foreclosure |
|    - Alasan (wajib)               |
|    - Rencana likuidasi            |
|    - Estimasi nilai jual          |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 2. Admin review & approve         |
|    - Verifikasi kelengkapan       |
|    - Pertimbangan hukum           |
|    - Approval atau rejection      |
+-----------------------------------+
        |
        v (approved)
+-----------------------------------+
| 3. Proses lelang/penjualan        |
|    - Dicatat: tanggal jual,       |
|      pembeli, harga jual          |
|    - Upload bukti transaksi       |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 4. Alokasi hasil penjualan:       |
|                                   |
|    sale_price >= outstanding:      |
|      → Lunasi sisa pinjaman       |
|      → Surplus dikembalikan ke    |
|        nasabah                    |
|                                   |
|    sale_price < outstanding:       |
|      → Seluruh hasil ke pinjaman  |
|      → Selisih dicatat sebagai    |
|        kerugian (loss)            |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 5. Update status:                 |
|    - jaminan.status = FORECLOSED  |
|    - Record sale details          |
|    - Update pinjaman outstanding  |
+-----------------------------------+
```

**Aturan foreclosure:**
- **Hanya untuk pinjaman WRITTEN_OFF** — tidak bisa foreclosure pinjaman yang masih aktif
- Manager mengajukan, **Admin yang approve** — keputusan besar memerlukan otorisasi tertinggi
- Hasil penjualan di atas sisa hutang **wajib dikembalikan** ke nasabah (tidak boleh diambil koperasi)
- Selisih negatif (sale_price < outstanding) dicatat sebagai **realized loss**
- Internal collateral (saldo): tidak ada foreclosure — saldo langsung dipotong untuk melunasi pinjaman
- Audit trail lengkap: setiap langkah foreclosure tercatat dengan timestamp dan pelaku

### 9. School Context — Konteks Sekolah

Koperasi sekolah/BMT memiliki karakteristik unik terkait jaminan:

**Pola umum per segmen nasabah:**

| Segmen | Jenis Pinjaman | Jaminan Tipikal | Catatan |
|--------|---------------|-----------------|---------|
| Guru/Ustadz (gaji kecil-menengah) | Pinjaman reguler < 10 juta | Personal guarantee atau simpanan | Gaji tetap = risiko rendah |
| Guru/Ustadz (gaji menengah-besar) | Pinjaman reguler 10-50 juta | Deposito sebagai jaminan | Deposito di-hold via K002 |
| Guru/Ustadz (pinjaman besar) | Pinjaman > 50 juta (rumah, kendaraan) | BPKB atau sertifikat tanah | Membutuhkan taksasi fisik |
| TU/Staf | Pinjaman reguler < 10 juta | Personal guarantee atau simpanan | Sama dengan guru kecil |
| Orang tua/Wali (pinjaman pendidikan) | Pinjaman pendidikan 5-30 juta | Simpanan + personal guarantee | Tenor panjang, bunga kompetitif |
| Siswa/Santri | Tidak boleh pinjam langsung | N/A | Orang tua sebagai penjamin (K007) |

**Pinjaman darurat (Qardh untuk BMT):**
- Plafon kecil (configurable, default: max Rp 5.000.000)
- Jaminan **opsional** — bisa tanpa jaminan jika sesuai kebijakan
- Jika ada jaminan: ijazah atau surat berharga sederhana sudah cukup
- Fast-track approval: Manager bisa langsung approve tanpa credit analysis lengkap

**Ijazah sebagai jaminan:**
- Hanya berlaku untuk pinjaman darurat kecil — bukan jaminan untuk pinjaman besar
- Nilai taksasi: nominal (misal Rp 1.000.000) — bukan nilai pasar ijazah
- Tujuan: **komitmen moral** lebih dari nilai ekonomis
- Penyimpanan: di brankas koperasi, wajib dikembalikan saat pinjaman lunas

### 10. Penjamin (Guarantor) — Model Diperluas

Model penjamin dasar sudah didefinisikan di [ADR-K007](./ADR-K007-pinjaman.md) Section 4. ADR ini memperluas aturan terkait jaminan.

**Konfigurasi penjamin:**

```
guarantor_collateral_config:
  max_active_guarantees: 3           # Max penjaminan aktif per orang (default: 3)
  internal_only: false               # Jika true, hanya nasabah yang bisa jadi penjamin
  require_consent_document: true     # Wajib upload surat pernyataan penjaminan
  overdue_notification: true         # Notifikasi ke penjamin jika peminjam menunggak
  notification_after_dpd: 7          # Notifikasi setelah 7 hari keterlambatan
```

**Penjamin internal (nasabah existing):**
- Harus nasabah **ACTIVE** di tenant yang sama
- Tidak boleh memiliki pinjaman dengan Kol > 2 (lancar/dalam perhatian khusus saja)
- Max active guarantees configurable per tenant (default: 3)
- Penjamin di-notify saat peminjam overdue > N hari (configurable)
- Penjamin **tidak bisa di-deactivate** (K001) selama masih menjamin pinjaman aktif

**Penjamin eksternal:**
- Data dasar disimpan: nama lengkap, nomor identitas (KTP), alamat, telepon
- **Bukan nasabah** — tidak memiliki rekening di koperasi
- Surat pernyataan penjaminan wajib di-upload dan di-tanda-tangani
- Hubungan dengan peminjam dicatat (`relationship`: orang tua, saudara, rekan kerja, dll)

**Surat pernyataan penjaminan (guarantee letter):**
- Template disediakan oleh koperasi
- Wajib ditandatangani penjamin dan peminjam
- Wajib di-upload sebagai dokumen jaminan (document_type = SURAT_PERNYATAAN)
- Menyebutkan: nominal yang dijaminkan, jangka waktu, konsekuensi gagal bayar

### 11. Islamic Mode — Rahn (Gadai Syariah)

Dalam mode BMT, jaminan mengikuti konsep **Rahn** sesuai fatwa DSN-MUI:

**Terminologi:**

```
Rahn         = Akad gadai / penjaminan syariah
Marhun       = Barang yang digadaikan (jaminan)
Rahin        = Pihak yang menggadaikan (nasabah/peminjam)
Murtahin     = Pihak yang menerima gadai (koperasi/BMT)
Ujrah        = Biaya penyimpanan/pemeliharaan barang gadai
```

**Prinsip syariah Rahn:**
- Koperasi/BMT **tidak boleh menggunakan atau mengambil manfaat** dari barang yang digadaikan (Marhun)
- Marhun tetap milik Rahin (nasabah) — koperasi hanya menyimpan
- Jika Rahin gagal bayar: Marhun dijual, **surplus wajib dikembalikan** ke Rahin
- Biaya penyimpanan/pemeliharaan (**Ujrah**) diperbolehkan — bukan bunga, tapi biaya riil

**Ujrah — biaya pemeliharaan (physical collateral):**

```
ujrah_config:
  enabled: true                      # Hanya untuk BMT mode
  calculation_method: "fixed_monthly" | "percentage_of_value"

  fixed_monthly_amount: 50000        # Rp 50.000/bulan per item
  percentage_rate: 0.005             # 0.5% per bulan dari appraised_value

  billing_frequency: "monthly"       # Ditagihkan bulanan
  included_in_installment: true      # Bisa dimasukkan ke angsuran
```

**Aturan ujrah:**
- Hanya berlaku untuk jaminan **fisik** yang disimpan di koperasi (SURAT_BERHARGA di brankas, BARANG_BERGERAK di gudang)
- **Tidak berlaku** untuk INTERNAL_BALANCE dan PERSONAL_GUARANTEE
- Ujrah adalah **biaya riil** penyimpanan/pemeliharaan — bukan keuntungan dari gadai
- Ujrah bisa dimasukkan ke angsuran bulanan atau ditagihkan terpisah (configurable)
- Jurnal: Debit Kas/Rekening Nasabah → Credit Pendapatan Ujrah (halal, bukan riba)

### 12. Data Model

**Tabel `jaminan` (Collateral):**

```
jaminan
+-- id                    UUID v7 (PK)
+-- tenant_id             UUID (FK -> tenant)
+-- nasabah_id            UUID (FK -> nasabah) NOT NULL (pemilik jaminan)
+-- branch_id             UUID (FK -> branch)
|
+-- -- Identitas Jaminan --
+-- collateral_number     VARCHAR UNIQUE per tenant (auto-generate: JM-2026-JKT-00000001)
+-- collateral_type       ENUM (internal_balance, surat_berharga, barang_bergerak, personal_guarantee)
+-- description           TEXT NOT NULL (deskripsi lengkap jaminan)
|
+-- -- Detail per Jenis --
+-- item_category         VARCHAR (nullable, sub-kategori: "kendaraan_roda2", "tanah_shm", dll)
+-- item_brand            VARCHAR (nullable, merek/tipe: "Honda Vario 150")
+-- item_year             INTEGER (nullable, tahun pembuatan)
+-- item_serial_number    VARCHAR (nullable, no rangka/mesin/seri)
+-- ownership_name        VARCHAR NOT NULL (nama pemilik di dokumen)
+-- ownership_document    VARCHAR (nullable, no BPKB/no sertifikat)
|
+-- -- Internal Balance (khusus INTERNAL_BALANCE) --
+-- rekening_id           UUID (nullable, FK -> rekening, rekening yang di-hold)
+-- hold_amount           NUMERIC(15,2) (nullable, nominal yang di-hold di rekening)
|
+-- -- Valuasi --
+-- appraised_value       NUMERIC(15,2) NOT NULL (nilai taksasi/pasar)
+-- acceptance_rate       NUMERIC(5,4) NOT NULL DEFAULT 1.0000 (misal: 0.7000 = 70%)
+-- collateral_value      NUMERIC(15,2) NOT NULL (= appraised_value x acceptance_rate)
+-- appraised_at          TIMESTAMPTZ NOT NULL
+-- appraised_by          UUID NOT NULL (FK -> user, penilai)
+-- next_revaluation_at   TIMESTAMPTZ (nullable, jadwal taksasi ulang)
|
+-- -- Status --
+-- status                ENUM (registered, pledged, released, foreclosed)
|
+-- -- Release --
+-- released_at           TIMESTAMPTZ (nullable)
+-- released_by           UUID (nullable, FK -> user)
+-- release_notes         TEXT (nullable)
|
+-- -- Foreclosure --
+-- foreclosed_at         TIMESTAMPTZ (nullable)
+-- foreclosed_by         UUID (nullable, FK -> user)
+-- sale_price            NUMERIC(15,2) (nullable, harga jual saat foreclosure)
+-- sale_date             DATE (nullable)
+-- sale_buyer            VARCHAR (nullable, pembeli)
+-- sale_notes            TEXT (nullable)
|
+-- -- Islamic (Rahn) --
+-- ujrah_monthly         NUMERIC(15,2) (nullable, biaya pemeliharaan bulanan)
+-- ujrah_total_charged   NUMERIC(15,2) DEFAULT 0 (total ujrah yang sudah ditagihkan)
+-- ujrah_total_paid      NUMERIC(15,2) DEFAULT 0 (total ujrah yang sudah dibayar)
|
+-- -- Vernon Fields --
+-- _rels                 JSONB NOT NULL DEFAULT '{}'
+-- _data                 JSONB NOT NULL DEFAULT '{}'
|
+-- -- Audit --
    +-- created_at        TIMESTAMPTZ
    +-- created_by        UUID
    +-- updated_at        TIMESTAMPTZ
    +-- updated_by        UUID
```

**Tabel `jaminan_pinjaman` (Binding jaminan ke pinjaman — many-to-many):**

```
jaminan_pinjaman
+-- id                    UUID v7 (PK)
+-- tenant_id             UUID (FK -> tenant)
+-- jaminan_id            UUID (FK -> jaminan) NOT NULL
+-- pinjaman_id           UUID (FK -> pinjaman) NOT NULL
|
+-- -- Nilai yang Diikat --
+-- pledged_value         NUMERIC(15,2) NOT NULL (porsi nilai jaminan untuk pinjaman ini)
|
+-- -- Status --
+-- status                ENUM (active, released)
+-- bound_at              TIMESTAMPTZ NOT NULL
+-- released_at           TIMESTAMPTZ (nullable)
|
+-- -- Vernon Fields --
+-- _rels                 JSONB NOT NULL DEFAULT '{}'
+-- _data                 JSONB NOT NULL DEFAULT '{}'
|
+-- -- Audit --
    +-- created_at        TIMESTAMPTZ
    +-- created_by        UUID
    +-- updated_at        TIMESTAMPTZ
    +-- updated_by        UUID
```

**Tabel `jaminan_dokumen` (Dokumen per jaminan):**

```
jaminan_dokumen
+-- id                    UUID v7 (PK)
+-- tenant_id             UUID (FK -> tenant)
+-- jaminan_id            UUID (FK -> jaminan) NOT NULL
|
+-- -- Dokumen --
+-- document_type         ENUM (foto_barang, fotokopi_bpkb, sertifikat_tanah,
|                               surat_kuasa, berita_acara, foto_identitas,
|                               surat_pernyataan, ijazah, lainnya)
+-- file_url              VARCHAR NOT NULL (path/URL file yang di-upload)
+-- file_name             VARCHAR NOT NULL (nama file asli)
+-- file_size_bytes       BIGINT NOT NULL
+-- mime_type             VARCHAR NOT NULL (image/jpeg, application/pdf, dll)
+-- description           TEXT (nullable, keterangan dokumen)
|
+-- -- Versioning --
+-- version               INTEGER NOT NULL DEFAULT 1
+-- is_superseded         BOOLEAN DEFAULT false (true jika ada versi baru)
+-- superseded_by_id      UUID (nullable, FK -> jaminan_dokumen, versi pengganti)
|
+-- -- Vernon Fields --
+-- _rels                 JSONB NOT NULL DEFAULT '{}'
+-- _data                 JSONB NOT NULL DEFAULT '{}'
|
+-- -- Audit --
    +-- created_at        TIMESTAMPTZ
    +-- created_by        UUID (uploader)
    +-- updated_at        TIMESTAMPTZ
    +-- updated_by        UUID
```

**Tabel `jaminan_valuasi` (History revaluasi):**

```
jaminan_valuasi
+-- id                    UUID v7 (PK)
+-- tenant_id             UUID (FK -> tenant)
+-- jaminan_id            UUID (FK -> jaminan) NOT NULL
|
+-- -- Valuasi --
+-- appraised_value       NUMERIC(15,2) NOT NULL
+-- acceptance_rate       NUMERIC(5,4) NOT NULL
+-- collateral_value      NUMERIC(15,2) NOT NULL
+-- valuation_type        ENUM (initial, periodic, on_demand)
+-- valuation_notes       TEXT (nullable, catatan penilai)
|
+-- -- Penilai --
+-- appraised_by          UUID NOT NULL (FK -> user)
+-- appraised_at          TIMESTAMPTZ NOT NULL
|
+-- -- Vernon Fields --
+-- _rels                 JSONB NOT NULL DEFAULT '{}'
+-- _data                 JSONB NOT NULL DEFAULT '{}'
|
+-- -- Audit --
    +-- created_at        TIMESTAMPTZ
    +-- created_by        UUID
    +-- updated_at        TIMESTAMPTZ
    +-- updated_by        UUID
```

**Index penting:**
- `jaminan(tenant_id, nasabah_id, status)` — query jaminan per nasabah
- `jaminan(tenant_id, collateral_type, status)` — laporan per jenis
- `jaminan(tenant_id, status, next_revaluation_at)` — scheduled revaluation
- `jaminan_pinjaman(jaminan_id, status)` — binding aktif per jaminan
- `jaminan_pinjaman(pinjaman_id, status)` — jaminan per pinjaman
- `jaminan_dokumen(jaminan_id, is_superseded)` — dokumen aktif per jaminan

### 13. Vernon _rels dan _data Structure

**Jaminan _rels:**
```json
{
  "nasabah_id":   "018f...",
  "branch_id":    "018f...",
  "tenant_id":    "018f...",
  "rekening_id":  "018f..."
}
```

**Jaminan _data:**
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
  },
  "rekening": {
    "id":             "018f...",
    "account_number": "TB-2026-JKT-00000005",
    "category":       "tabungan",
    "balance":        25000000
  },
  "pinjaman_list": [
    {
      "id":          "018f...",
      "loan_number": "PJ-2026-JKT-00000001",
      "status":      "active",
      "outstanding": 15000000
    }
  ]
}
```

**Jaminan_pinjaman _rels:**
```json
{
  "jaminan_id":  "018f...",
  "pinjaman_id": "018f...",
  "tenant_id":   "018f..."
}
```

**Jaminan_pinjaman _data:**
```json
{
  "jaminan": {
    "id":                "018f...",
    "collateral_number": "JM-2026-JKT-00000001",
    "collateral_type":   "surat_berharga",
    "description":       "BPKB Motor Honda Vario 150 2024",
    "collateral_value":  10500000
  },
  "pinjaman": {
    "id":          "018f...",
    "loan_number": "PJ-2026-JKT-00000001",
    "status":      "active",
    "outstanding": 15000000
  },
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.nasabah` di semua jaminan nasabah tersebut
- `BranchUpdatedEvent` → update `_data.branch` di semua jaminan cabang tersebut
- `RekeningUpdatedEvent` → update `_data.rekening` di jaminan INTERNAL_BALANCE terkait
- `PinjamanUpdatedEvent` → update `_data.pinjaman_list` di jaminan terkait
- `PinjamanUpdatedEvent` → update `_data.pinjaman` di jaminan_pinjaman terkait
- `JaminanUpdatedEvent` → update `_data.jaminan` di jaminan_pinjaman terkait
- `JaminanRevaluatedEvent` → update `collateral_value` di jaminan_pinjaman terkait

### 14. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View daftar jaminan | v | v | v | v |
| View detail jaminan + dokumen | v | v | v | v |
| Input data jaminan baru (REGISTERED) | v | v | v | v |
| Upload dokumen jaminan | v | v | v | v |
| Melakukan taksasi/valuasi | - | v | v | v |
| Revaluasi jaminan | - | v | v | v |
| Bind jaminan ke pinjaman (PLEDGED) | - | v | v | v |
| Release jaminan (physical) | - | - | v | v |
| Release jaminan (internal/auto) | - | - | auto | auto |
| Ajukan foreclosure | - | - | v | v |
| Approve foreclosure | - | - | - | v |
| Configure collateral policy | - | - | - | v |
| Configure LTV ratio per produk | - | - | - | v |
| Override acceptance_rate per item | - | - | v | v |

**Catatan:**
- Teller bisa input data dan upload dokumen — operasi rutin harian
- Taksasi/valuasi membutuhkan **Supervisor+** — penilaian butuh keahlian dan otorisasi
- Release physical collateral membutuhkan **Manager+** — keputusan bisnis yang signifikan
- Release internal collateral otomatis oleh sistem saat pinjaman COMPLETED — tidak butuh approval manual
- Foreclosure: Manager mengajukan, **Admin yang approve** — keputusan paling berat
- Konfigurasi LTV dan policy hanya **Admin** — mempengaruhi seluruh tenant

### 15. Dual-Mode Terminology

| Konsep | `coop_type = "general"` | `coop_type = "islamic"` |
|--------|------------------------|------------------------|
| Jaminan/agunan | Jaminan / Agunan | Rahn / Marhun |
| Pemilik jaminan | Pemberi jaminan | Rahin |
| Penerima jaminan | Koperasi | Murtahin |
| Barang jaminan | Agunan / Barang jaminan | Marhun / Marhun Bih |
| Pengikatan | Pengikatan jaminan | Akad Rahn |
| Pelepasan | Pelepasan jaminan | Fakku Rahn |
| Eksekusi jaminan | Eksekusi / Penjualan | Bai' al-Marhun |
| Biaya penyimpanan | Biaya penyimpanan | Ujrah |
| Nilai taksasi | Nilai taksasi | Nilai Marhun |
| Personal guarantee | Penjamin / Borgtocht | Kafiil / Kafalah |
| Surat penjaminan | Surat penjaminan | Akad Kafalah |
| Pinjaman terkait | Pinjaman | Pembiayaan |
| Rasio jaminan | Loan-to-Value (LTV) | Loan-to-Value (LTV) |

## Consequences

### Positif

- **Multi-type support** — mengakomodasi semua jenis jaminan yang umum di koperasi sekolah: saldo simpanan, surat berharga, barang bergerak, dan personal guarantee
- **Auto-hold integration** — jaminan internal terhubung langsung ke mekanisme `hold_amount` K002, sehingga saldo yang dijaminkan otomatis terlindungi
- **Traceable lifecycle** — setiap jaminan tercatat dari REGISTERED hingga RELEASED/FORECLOSED dengan audit trail lengkap
- **Syariah compliant** — Rahn mechanics dengan pemisahan ujrah yang jelas, koperasi tidak mengambil manfaat dari marhun
- **Coverage monitoring** — coverage ratio tracking otomatis dengan alert saat turun di bawah threshold

### Negatif

- **Manual valuation overhead** — taksasi fisik membutuhkan waktu dan keahlian, bisa menjadi bottleneck untuk pinjaman yang butuh approval cepat
  - *Mitigasi*: pinjaman darurat bisa tanpa jaminan atau dengan jaminan minimal (ijazah), fast-track approval tetap bisa berjalan
- **Document storage cost** — penyimpanan foto dan scan dokumen membutuhkan storage yang signifikan seiring pertumbuhan data
  - *Mitigasi*: implementasi max file size (5MB default), archival policy untuk dokumen jaminan yang sudah RELEASED > 2 tahun
- **Revaluation discipline** — taksasi ulang periodik membutuhkan disiplin operasional yang konsisten
  - *Mitigasi*: scheduled job reminder untuk revaluasi yang akan jatuh tempo, alert ke Supervisor/Manager
- **Multi-loan binding complexity** — jika `allow_multi_loan = true`, tracking porsi nilai jaminan per pinjaman menjadi lebih kompleks
  - *Mitigasi*: default `allow_multi_loan = false`, hanya diaktifkan jika tenant benar-benar membutuhkan dengan pelatihan operator

## Alternatives Considered

### A. Jaminan sebagai Atribut Pinjaman (bukan Entitas Terpisah)

Data jaminan disimpan sebagai field/JSONB di tabel pinjaman, bukan entitas terpisah.

**Ditolak** karena: satu jaminan bisa digunakan untuk beberapa pinjaman (jika dikonfigurasi), jaminan memiliki lifecycle sendiri (REGISTERED bisa ada sebelum pinjaman diajukan), dan dokumen jaminan memerlukan tabel relasi terpisah. Menyimpan di pinjaman menyulitkan tracking jaminan yang masih tersedia untuk pinjaman baru.

### B. Tanpa Revaluasi Periodik

Nilai taksasi ditetapkan sekali saat pendaftaran dan tidak pernah diubah.

**Ditolak** karena: nilai aset fisik berubah seiring waktu (depresiasi kendaraan, apresiasi tanah). Tanpa revaluasi, coverage ratio bisa menjadi tidak akurat dan memberikan false sense of security. Regulasi juga mengisyaratkan penilaian berkala untuk aset yang dijaminkan.

### C. Ujrah Dimasukkan ke Margin Pembiayaan (bukan Terpisah)

Biaya penyimpanan/pemeliharaan barang gadai dimasukkan ke dalam margin pembiayaan, bukan ditagihkan terpisah sebagai ujrah.

**Ditolak** karena: mencampurkan ujrah dengan margin pembiayaan mengaburkan transparansi biaya. Nasabah berhak tahu berapa biaya penyimpanan dan berapa margin pembiayaan. Secara syariah, ujrah harus berdasarkan biaya riil pemeliharaan, bukan ditambahkan ke margin yang bersifat profit.
