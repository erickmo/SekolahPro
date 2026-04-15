# ADR-K009: Denda & Penalti

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Denda dan penalti adalah mekanisme untuk mendisiplinkan nasabah yang terlambat membayar angsuran atau melanggar ketentuan produk. Ini juga merupakan salah satu area dengan **perbedaan paling signifikan** antara koperasi konvensional dan BMT:

- **Konvensional**: Denda = persentase dari jumlah tertunggak, masuk ke pendapatan koperasi
- **Islamic/BMT**: Ta'zir = nominal tetap (bukan persentase), **wajib masuk ke dana sosial** (baitul maal), bukan pendapatan koperasi — sesuai fatwa DSN-MUI No. 17/DSN-MUI/IX/2000
- **Ta'widh**: Kompensasi atas kerugian nyata yang diderita koperasi — membutuhkan bukti dan hanya berlaku di BMT
- **Transparansi**: Nasabah harus tahu sejak awal berapa denda yang akan dikenakan

Sistem harus bisa menghitung denda otomatis, track terpisah dari angsuran, dan memastikan compliance syariah untuk BMT mode.

## Decision

### 1. Late Payment Penalty — Denda Keterlambatan

**General Mode (Konvensional):**

```
late_penalty_config_general:
  calculation_method: "percentage_daily" | "percentage_monthly" | "fixed_daily" | "fixed_monthly"

  # percentage_daily: denda = % per hari dari jumlah tertunggak
  percentage_daily_rate: 0.001      # 0.1% per hari (default)

  # percentage_monthly: denda = % per bulan dari jumlah tertunggak
  percentage_monthly_rate: 0.02     # 2% per bulan

  # fixed_daily: denda = nominal tetap per hari keterlambatan
  fixed_daily_amount: 5000          # Rp 5.000 per hari

  # fixed_monthly: denda = nominal tetap per bulan keterlambatan
  fixed_monthly_amount: 50000       # Rp 50.000 per bulan

  # Basis perhitungan
  calculation_basis: "overdue_amount" | "installment_amount" | "outstanding_principal"
  # overdue_amount: denda dihitung dari jumlah yang terlambat saja
  # installment_amount: denda dihitung dari total angsuran bulanan
  # outstanding_principal: denda dihitung dari sisa pokok keseluruhan
```

**Contoh perhitungan (percentage_daily, overdue_amount):**

```
Angsuran jatuh tempo: 1.120.000 (pokok 1.000.000 + bunga 120.000)
Terlambat: 15 hari
Rate denda: 0.1% per hari

Denda = 1.120.000 x 0.001 x 15 = 16.800
```

**Islamic Mode (BMT) — Ta'zir:**

```
tazir_config_islamic:
  calculation_method: "fixed_daily" | "fixed_monthly"
  # WAJIB fixed amount — TIDAK BOLEH persentase (fatwa DSN-MUI)

  fixed_daily_amount: 5000          # Rp 5.000 per hari
  fixed_monthly_amount: 50000       # Rp 50.000 per bulan

  # Dana ta'zir wajib masuk ke:
  destination: "social_fund"        # Baitul maal / dana sosial
  # TIDAK BOLEH masuk ke pendapatan koperasi
```

**Aturan ta'zir (BMT):**
- **WAJIB nominal tetap** — tidak boleh persentase dari jumlah tertunggak
- **WAJIB masuk ke dana sosial** (baitul maal) — bukan pendapatan operasional BMT
- Hanya dikenakan pada nasabah yang **mampu tapi lalai** (bukan yang tidak mampu)
- Jika nasabah terbukti **tidak mampu** → ta'zir di-waive dan diberikan keringanan
- Jurnal akuntansi: Debit Kas → Credit Dana Sosial (bukan Credit Pendapatan Denda)

### 2. Ta'widh — Kompensasi Kerugian Nyata (Islamic Only)

```
tawidh_config:
  enabled: true                     # Hanya untuk BMT mode
  requires_proof: true              # WAJIB ada bukti kerugian nyata
  max_amount: "actual_loss"         # Tidak boleh melebihi kerugian nyata
  approval_required: true           # Manager+ harus approve
```

**Ta'widh vs Ta'zir:**

| Aspek | Ta'zir | Ta'widh |
|-------|--------|---------|
| Tujuan | Sanksi/disiplin | Kompensasi kerugian |
| Dasar | Keterlambatan | Kerugian nyata yang bisa dibuktikan |
| Perhitungan | Fixed amount | Sesuai kerugian aktual |
| Bukti | Tidak perlu | **WAJIB** ada bukti (biaya penagihan, dll) |
| Penerima | Dana sosial (baitul maal) | Koperasi/BMT (sebagai kompensasi) |
| Masuk P&L | **TIDAK** | Ya, sebagai pendapatan lain-lain |
| Approval | Auto (sesuai config) | Manager+ (case-by-case) |

**Contoh ta'widh:**
- Biaya penagihan (transportasi ke rumah nasabah): Rp 50.000 → bisa dikenakan ta'widh
- Biaya surat peringatan (cetak, kirim): Rp 25.000 → bisa dikenakan ta'widh
- Bunga/opportunity cost dari dana yang terlambat: **TIDAK BOLEH** — ini riba

### 3. Penalty Calculation Engine

```
Penalty Calculation Job (harian, scheduled):
+-----------------------------------+
| 1. Query semua angsuran yang      |
|    is_overdue = true              |
|    DAN belum lunas                |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 2. Per angsuran:                  |
|    - Cek apakah masih dalam grace |
|      period denda                 |
|    - Hitung DPD (hari keterlambatan)|
|    - Hitung denda sesuai config   |
|    - Cek penalty cap              |
+-----------------------------------+
        |
        v
+-----------------------------------+
| 3. Simpan/update ke tabel denda   |
|    - denda_calculated (akumulasi) |
|    - denda_paid (sudah dibayar)   |
|    - denda_outstanding (belum bayar)|
+-----------------------------------+
        |
        v
+-----------------------------------+
| 4. Update pinjaman:               |
|    outstanding_penalty =           |
|    SUM(denda_outstanding) all      |
|    angsuran                        |
+-----------------------------------+
```

**Aturan:**
- Perhitungan denda berjalan **harian** via scheduled job
- Denda dihitung secara **akumulatif** — setiap hari keterlambatan menambah denda
- Denda berhenti bertambah saat angsuran dilunasi
- Denda yang sudah dihitung dan dibayar **tidak di-refund** meskipun nasabah melunasi
- Perhitungan menggunakan **calendar days** (bukan business days)
- Denda di-calculate ulang jika ada perubahan konfigurasi rate → hanya berlaku untuk hari setelah perubahan (tidak retroaktif)

### 4. Penalty Waiver — Pengurangan/Penghapusan Denda

```
Supervisor/Manager mengajukan waiver
        |
        v
+-----------------------------------+
| Waiver Application:               |
| - pinjaman_id                     |
| - waiver_type: full | partial     |
| - waiver_amount (jika partial)    |
| - waiver_reason (WAJIB)           |
| - supporting_document (opsional)  |
+-----------------------------------+
        |
        v
+-----------------------------------+
| Approval:                         |
| - Partial waiver: Manager+        |
| - Full waiver: Admin              |
+-----------------------------------+
        |
        v
+-----------------------------------+
| Execute:                          |
| 1. Update denda_outstanding       |
| 2. Record waiver di audit trail   |
| 3. Update pinjaman.outstanding_   |
|    penalty                        |
+-----------------------------------+
```

**Aturan:**
- Waiver hanya oleh **Manager+** (partial) atau **Admin** (full)
- Alasan waiver **wajib** dicantumkan — documented decision
- Waiver di BMT mode lebih umum — untuk nasabah yang terbukti tidak mampu (ta'zir di-waive)
- Waiver tidak mengubah riwayat keterlambatan — DPD dan aging tetap tercatat
- Max total waiver per pinjaman configurable (default: unlimited, tapi setiap waiver di-audit)

### 5. Early Withdrawal Penalty — Deposito

Penalti untuk penarikan deposito sebelum jatuh tempo:

```
early_withdrawal_penalty_config:
  general_mode:
    penalty_type: "percentage_of_interest"
    penalty_rate: 0.50              # 50% dari bunga yang sudah diperoleh
    minimum_holding_period_days: 30 # Tidak boleh tarik sebelum 30 hari

  islamic_mode:
    penalty_type: "percentage_of_profit"
    penalty_rate: 0.50              # 50% dari bagi hasil yang sudah diterima
    minimum_holding_period_days: 30
```

**Contoh:**

```
Deposito: 10.000.000, Tenor: 12 bulan, Rate: 6%/tahun
Ditarik setelah 6 bulan

Bunga yang seharusnya (6 bulan):
  10.000.000 x 6% x (6/12) = 300.000

Penalti: 300.000 x 50% = 150.000

Nasabah terima: 10.000.000 + 300.000 - 150.000 = 10.150.000
```

**Aturan:**
- Penalti early withdrawal **configurable per produk deposito**
- Minimum holding period: deposito tidak boleh ditarik sebelum periode minimum (default: 30 hari)
- Jika ditarik sebelum minimum holding period → seluruh bunga/profit hangus (penalti 100%)
- Penarikan dini membutuhkan approval **Manager+**
- Record penalti di tabel denda sebagai jenis `EARLY_WITHDRAWAL`

### 6. Early Settlement Penalty — Pinjaman

Penalti untuk pelunasan pinjaman sebelum tenor berakhir:

```
early_settlement_penalty_config:
  general_mode:
    enabled: true                   # Configurable — bisa dimatikan
    penalty_type: "percentage_of_remaining"
    penalty_rate: 0.01              # 1% dari sisa pokok
    min_tenor_before_settlement: 6  # Minimal sudah 6 bulan baru boleh early settlement

  islamic_mode:
    enabled: false                  # Default: TIDAK ADA penalti (ibra)
    # BMT memberikan ibra (pembebasan margin), bukan mengenakan penalti
```

**Aturan:**
- General mode: penalti early settlement **opsional** — configurable per produk
- Islamic mode: **tidak ada penalti** pelunasan dini — BMT memberikan **ibra** (pembebasan sisa margin)
- Minimum tenor: nasabah harus sudah membayar minimal N bulan sebelum boleh early settlement (configurable)
- Penalti dihitung dari **sisa pokok**, bukan sisa total kewajiban
- Record di tabel denda sebagai jenis `EARLY_SETTLEMENT`

### 7. Grace Period Denda

Periode setelah due date di mana **denda belum dihitung**:

```
penalty_grace_period_config:
  enabled: true
  grace_days: 3                    # Default: 3 hari setelah due date
  # Denda mulai dihitung dari hari ke-4 setelah due date
```

**Contoh:**

```
Due date: 10 Januari
Grace period: 3 hari

Bayar tgl 10-13 Januari → TIDAK KENA denda
Bayar tgl 14 Januari    → Denda 1 hari (bukan 4 hari)
Bayar tgl 20 Januari    → Denda 7 hari (20 - 13)
```

**Aturan:**
- Grace period denda **berbeda** dari grace period angsuran pertama (K007 Section 6)
- Grace period denda berlaku untuk **setiap angsuran**, bukan hanya angsuran pertama
- Selama grace period: DPD tetap dihitung (untuk NPL), tapi **denda belum dikenakan**
- Grace period **configurable per produk** — bisa berbeda antar produk
- Default: 3 hari (umum di koperasi Indonesia)

### 8. Maximum Penalty Cap

Batas maksimal total denda yang bisa dikenakan per pinjaman:

```
penalty_cap_config:
  enabled: true
  cap_type: "percentage_of_principal" | "fixed_amount" | "percentage_of_installment"

  # percentage_of_principal: max denda = X% dari pokok pinjaman
  cap_percentage: 0.25              # Max 25% dari pokok (default)

  # fixed_amount: max denda = nominal tetap
  cap_amount: 5000000               # Max Rp 5.000.000

  # percentage_of_installment: max denda per angsuran = X% dari angsuran
  cap_per_installment_rate: 0.50    # Max 50% dari angsuran
```

**Aturan:**
- Cap **wajib diaktifkan** — mencegah denda yang tidak wajar dan melindungi nasabah
- Default: 25% dari pokok pinjaman — sesuai praktik umum koperasi
- Setelah denda mencapai cap: **denda berhenti dihitung** — sistem otomatis stop
- Cap berlaku untuk **total kumulatif** denda per pinjaman, bukan per angsuran
- Per-angsuran cap opsional — sebagai tambahan dari total cap
- Cap di-display ke nasabah di informasi pinjaman (transparansi)

### 9. Penalty Tracking — Data Model

```
denda
+-- id                    UUID v7 (PK)
+-- tenant_id             UUID (FK -> tenant)
+-- pinjaman_id           UUID (FK -> pinjaman) NOT NULL
+-- angsuran_id           UUID (nullable, FK -> angsuran, null untuk early_settlement/withdrawal)
+-- rekening_id           UUID (FK -> rekening pinjaman) NOT NULL
|
+-- -- Jenis Denda --
+-- penalty_type          ENUM (late_payment, early_settlement, early_withdrawal, tazir, tawidh)
|
+-- -- Perhitungan --
+-- calculation_basis     NUMERIC(15,2) NOT NULL (nominal yang jadi basis perhitungan)
+-- penalty_rate          NUMERIC(8,6) (nullable, rate yang digunakan)
+-- penalty_days          INTEGER DEFAULT 0 (jumlah hari keterlambatan)
+-- calculated_amount     NUMERIC(15,2) NOT NULL (denda yang dihitung)
|
+-- -- Cap --
+-- cap_applied           BOOLEAN DEFAULT false (apakah sudah kena cap)
+-- capped_amount         NUMERIC(15,2) (nullable, jika di-cap, berapa yang dipotong)
+-- final_amount          NUMERIC(15,2) NOT NULL (denda final setelah cap)
|
+-- -- Pembayaran --
+-- paid_amount           NUMERIC(15,2) DEFAULT 0
+-- waived_amount         NUMERIC(15,2) DEFAULT 0
+-- outstanding_amount    NUMERIC(15,2) NOT NULL (= final_amount - paid_amount - waived_amount)
|
+-- -- Status --
+-- status                ENUM (accruing, settled, waived, partial_waived)
|   # accruing: masih berjalan (denda masih bertambah)
|   # settled: sudah lunas
|   # waived: dihapuskan seluruhnya
|   # partial_waived: sebagian dihapuskan
|
+-- -- Waiver --
+-- waiver_reason         TEXT (nullable)
+-- waiver_approved_by    UUID (nullable, FK -> user)
+-- waiver_approved_at    TIMESTAMPTZ (nullable)
|
+-- -- Islamic Compliance --
+-- fund_destination      ENUM (koperasi_income, social_fund) DEFAULT 'koperasi_income'
|   # koperasi_income: masuk P&L koperasi (general mode denda)
|   # social_fund: masuk baitul maal (BMT ta'zir)
|
+-- -- Periode --
+-- period_start          DATE NOT NULL (mulai dihitung dari)
+-- period_end            DATE (nullable, sampai kapan — null jika masih accruing)
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
- `(pinjaman_id, status)` — query denda outstanding per pinjaman
- `(angsuran_id)` — link denda ke angsuran spesifik
- `(tenant_id, fund_destination, status)` — laporan dana sosial BMT
- `(tenant_id, penalty_type)` — laporan per jenis denda

### 10. Islamic Compliance — Fund Routing

```
Pembayaran denda masuk
        |
        v
+-----------------------------------+
| Cek coop_type:                    |
|                                   |
| GENERAL:                          |
|   Debit: Kas/Rekening Nasabah     |
|   Credit: Pendapatan Denda        |
|   -> Masuk P&L koperasi           |
|                                   |
| ISLAMIC (ta'zir):                 |
|   Debit: Kas/Rekening Nasabah     |
|   Credit: Dana Sosial/Baitul Maal |
|   -> TIDAK masuk P&L koperasi     |
|   -> Disalurkan untuk kegiatan    |
|      sosial (beasiswa, bantuan)   |
|                                   |
| ISLAMIC (ta'widh):                |
|   Debit: Kas/Rekening Nasabah     |
|   Credit: Pendapatan Lain-lain    |
|   -> Masuk P&L (kompensasi)       |
+-----------------------------------+
```

**Aturan:**
- Sistem **otomatis menentukan** `fund_destination` berdasarkan `coop_type` dan `penalty_type`
- General mode: semua denda → `koperasi_income`
- Islamic mode ta'zir: → `social_fund` (baitul maal)
- Islamic mode ta'widh: → `koperasi_income` (kompensasi kerugian nyata)
- Laporan dana sosial harus **terpisah** dari laporan keuangan operasional BMT
- Penyaluran dana sosial dicatat terpisah (akan diatur di K018 Zakat & Infaq)
- Audit trail lengkap untuk setiap alur dana denda — penting untuk DPS (Dewan Pengawas Syariah)

### 11. Vernon _rels dan _data Structure

**Denda _rels:**
```json
{
  "pinjaman_id":  "018f...",
  "angsuran_id":  "018f...",
  "rekening_id":  "018f...",
  "tenant_id":    "018f..."
}
```

**Denda _data:**
```json
{
  "pinjaman": {
    "id":          "018f...",
    "loan_number": "PJ-2026-JKT-00000001",
    "akad_type":   "murabahah"
  },
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "angsuran": {
    "id":                 "018f...",
    "installment_number": 5,
    "due_date":           "2026-09-10"
  }
}
```

**SyncEngine triggers:**
- `PinjamanUpdatedEvent` → update `_data.pinjaman` di semua denda pinjaman tersebut
- `NasabahUpdatedEvent` → update `_data.nasabah` di semua denda nasabah tersebut
- `AngsuranUpdatedEvent` → update `_data.angsuran` di denda terkait

### 12. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View denda per pinjaman | v | v | v | v |
| View laporan denda aggregat | v | v | v | v |
| Terima pembayaran denda | v | v | v | v |
| Waive denda (partial) | - | - | v | v |
| Waive denda (full) | - | - | - | v |
| Configure penalty rate | - | - | - | v |
| Configure penalty cap | - | - | - | v |
| Apply ta'widh (BMT) | - | - | v | v |
| View laporan dana sosial (BMT) | - | v | v | v |
| Override fund destination | - | - | - | v |

**Catatan:**
- Teller bisa lihat dan terima pembayaran denda — operasi rutin
- Waiver partial oleh Manager, full oleh Admin — keputusan bisnis yang perlu otorisasi tinggi
- Configuration penalty rate/cap hanya Admin — mempengaruhi seluruh tenant
- Ta'widh membutuhkan Manager+ karena harus ada bukti kerugian nyata
- Override fund destination hanya Admin — mencegah ta'zir dialihkan ke pendapatan koperasi

### 13. Dual-Mode Terminology

| Konsep | `coop_type = "general"` | `coop_type = "islamic"` |
|--------|------------------------|------------------------|
| Denda keterlambatan | Denda | Ta'zir |
| Kompensasi kerugian | Ganti rugi | Ta'widh |
| Basis perhitungan | % dari jumlah tertunggak | Nominal tetap (tidak boleh %) |
| Penerima dana | Koperasi (pendapatan) | Baitul maal / Dana sosial |
| Jurnal akuntansi | Credit: Pendapatan Denda | Credit: Dana Sosial (ta'zir) / Pendapatan Lain (ta'widh) |
| Penalti pelunasan dini | Denda pelunasan dini | Tidak ada (ibra — pembebasan margin) |
| Penalti deposito dini | Denda pencairan dini | Penalti pencairan dini bagi hasil |
| Grace period denda | Grace period | Grace period |
| Cap denda | Batas maksimal denda | Batas maksimal ta'zir |
| Pengurangan denda | Pengurangan denda | Pembebasan ta'zir |
| Nasabah mampu tapi lalai | N/A (denda berlaku untuk semua) | Wajib ta'zir |
| Nasabah tidak mampu | N/A (denda tetap berlaku) | Ta'zir di-waive, beri keringanan |

## Consequences

### Positif

- **Syariah compliant** — ta'zir dan ta'widh terpisah, fund routing otomatis ke dana sosial
- **Transparan** — cap denda dan grace period melindungi nasabah dari denda berlebihan
- **Flexible** — semua parameter configurable per tenant dan per produk
- **Auditable** — setiap perhitungan, pembayaran, dan waiver denda tercatat lengkap
- **Fair** — BMT membedakan nasabah mampu vs tidak mampu untuk pengenaan ta'zir
- **Automated** — perhitungan harian otomatis, cap otomatis, fund routing otomatis
- **Regulatory compliant** — mengikuti fatwa DSN-MUI dan praktik OJK/Dinas Koperasi

### Negatif

- **Dual-mode complexity** — logika berbeda untuk general vs Islamic, terutama di jurnal akuntansi
- **Fund routing** — perlu tracking terpisah untuk dana sosial BMT (rekening dan laporan terpisah)
- **Waiver governance** — risiko moral hazard jika terlalu mudah waive denda
- **Scheduled job dependency** — perhitungan denda bergantung pada job harian yang harus reliable
- **Ta'widh subjectivity** — penentuan "kerugian nyata" bisa subjektif, butuh kebijakan yang jelas

### Mitigasi

- Dual-mode logic di-implement dengan **strategy pattern** — satu engine per mode
- Dana sosial menggunakan rekening internal terpisah — tracking otomatis via jurnal
- Waiver membutuhkan approval + alasan + audit trail — governance ketat
- Scheduled job dengan **retry mechanism** dan monitoring — alert jika gagal
- Ta'widh membutuhkan **upload bukti** dan approval Manager+ — objectivity dijaga

## Alternatives Considered

### A. Denda sebagai Bagian dari Angsuran (bukan terpisah)

Denda langsung ditambahkan ke total angsuran, bukan di-track terpisah.

**Ditolak** karena: menyulitkan tracking denda per angsuran, waiver, dan laporan. Denda terpisah memudahkan audit, waiver partial, dan reporting ke regulator. Untuk BMT, ta'zir wajib terpisah karena fund destination berbeda.

### B. Real-time Calculation (tanpa scheduled job)

Denda dihitung real-time setiap kali nasabah buka informasi pinjaman.

**Ditolak** karena: menghasilkan angka yang tidak konsisten antar akses. Scheduled job memastikan semua perhitungan di-snapshot dan konsisten. Juga diperlukan untuk laporan batch dan reconciliation.

### C. Persentase untuk Ta'zir BMT

Menggunakan persentase (bukan nominal tetap) untuk perhitungan ta'zir di BMT mode.

**Ditolak** karena: **melanggar fatwa DSN-MUI**. Ta'zir harus berupa nominal tetap, bukan persentase dari jumlah tertunggak. Persentase membuat ta'zir mirip bunga (riba), yang bertentangan dengan prinsip syariah.

### D. Tanpa Cap Denda

Denda terus bertambah tanpa batas maksimal.

**Ditolak** karena: denda tanpa batas bisa melebihi pokok pinjaman — tidak adil bagi nasabah dan bertentangan dengan prinsip perlindungan konsumen. Regulasi OJK juga mengisyaratkan pembatasan denda yang wajar.

### E. Ta'zir Masuk Pendapatan Koperasi

Dana ta'zir dari BMT mode masuk ke pendapatan operasional seperti denda konvensional.

**Ditolak** karena: **melanggar fatwa DSN-MUI** secara fundamental. Ta'zir **wajib** disalurkan untuk kepentingan sosial melalui baitul maal. Mencampurkan ta'zir dengan pendapatan operasional merusak legitimasi syariah BMT.
