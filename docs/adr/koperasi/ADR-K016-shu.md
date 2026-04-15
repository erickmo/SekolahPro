# ADR-K016: SHU (Sisa Hasil Usaha)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

SHU (Sisa Hasil Usaha) adalah **mekanisme bagi hasil tahunan** yang unik untuk koperasi Indonesia. Berbeda dari dividen di perusahaan biasa, distribusi SHU didasarkan pada **kontribusi anggota** (besaran simpanan dan volume transaksi), bukan hanya kepemilikan saham. Regulasi di UU Koperasi No. 25 Tahun 1992 dan AD/ART masing-masing koperasi menentukan formula pembagian.

Sistem harus mengakomodasi:

- **Perhitungan otomatis**: SHU dihitung dari data akuntansi setelah year-end closing (K015)
- **Formula distribusi configurable**: persentase untuk cadangan, jasa anggota, dana pengurus, dll — bisa berbeda per tenant sesuai AD/ART
- **Dual-mode**: koperasi konvensional menggunakan istilah SHU, BMT bisa menggunakan istilah berbeda tapi konsepnya sama
- **Jasa Modal + Jasa Usaha**: dua komponen distribusi ke anggota — berdasarkan simpanan dan berdasarkan transaksi
- **Approval flow**: SHU dihitung sistem → disetujui pengurus → disahkan RAT → didistribusikan
- **Siswa sebagai anggota**: siswa/santri yang menjadi anggota koperasi juga berhak mendapat SHU

Referensi terkait:
- [ADR-K015](./ADR-K015-jurnal-coa.md) — Year-end closing menghasilkan angka SHU
- [ADR-K001](./ADR-K001-nasabah.md) — Data keanggotaan untuk perhitungan per anggota
- [ADR-K002](./ADR-K002-rekening.md) — Rekening simpanan dan saldo untuk jasa modal

## Decision

### 1. SHU Calculation — Dari Year-End Closing

SHU dihitung **setelah year-end closing** (K015) berhasil dijalankan. Angka SHU adalah selisih total pendapatan dikurangi total beban.

```
Year-End Closing (K015)
        │
        v
┌──────────────────────────────────────┐
│ SHU = Total Pendapatan (4xxx)        │
│      - Total Beban (5xxx)            │
│                                      │
│ Jika positif → Surplus (keuntungan)  │
│ Jika negatif → Defisit (kerugian)    │
└────────┬─────────────────────────────┘
         │
         v
┌──────────────────────────────────────┐
│ Create shu_periode record            │
│ - tahun_buku: 2026                   │
│ - total_pendapatan: xxx              │
│ - total_beban: xxx                   │
│ - shu_bruto: xxx                     │
│ - status: CALCULATED                 │
└──────────────────────────────────────┘
```

**Aturan:**
- SHU **hanya bisa dihitung** jika year-end closing untuk tahun tersebut sudah selesai
- Angka SHU diambil langsung dari saldo akun 3500 (SHU Tahun Berjalan) setelah closing
- Jika SHU negatif (defisit), **tidak ada distribusi** — defisit ditanggung cadangan atau dibawa ke tahun berikutnya sesuai keputusan RAT
- Satu tahun buku hanya memiliki **satu record** `shu_periode` — tidak bisa dihitung ulang setelah di-lock

### 2. Distribution Formula — Configurable per Tenant

Formula distribusi SHU mengikuti UU Koperasi dengan persentase yang **configurable per tenant** sesuai AD/ART masing-masing.

```
SHU Bruto
├── Pajak (jika ada)
│
└── SHU Neto (setelah pajak)
    │
    ├── Cadangan (min 25%)                  ← Wajib, tidak bisa < 25%
    │   Ditahan di koperasi, tidak dibagikan
    │
    ├── Jasa Anggota (configurable %)
    │   ├── Jasa Modal (sub-configurable %)
    │   │   Proporsional terhadap simpanan anggota
    │   │
    │   └── Jasa Usaha (sub-configurable %)
    │       Proporsional terhadap volume transaksi anggota
    │
    ├── Dana Pengurus (configurable %)
    │   Untuk pengurus koperasi aktif di tahun tersebut
    │
    ├── Dana Karyawan (configurable %)
    │   Untuk staf/karyawan koperasi
    │
    ├── Dana Pendidikan (configurable %)
    │   Untuk kegiatan pendidikan koperasi
    │
    ├── Dana Sosial (configurable %)
    │   Untuk kegiatan sosial / bantuan anggota
    │
    └── Dana Pembangunan Daerah (configurable %)
        Untuk kontribusi pembangunan daerah
```

**Aturan:**
- **Cadangan minimum 25%** — di-enforce di application layer. Tenant tidak bisa set < 25%
- Total semua persentase harus **= 100%** — di-enforce saat konfigurasi
- Persentase disimpan di **tenant config** dan bisa diubah setiap tahun (berlaku untuk SHU tahun berikutnya)
- Jasa Anggota dibagi lagi menjadi **Jasa Modal** dan **Jasa Usaha** — sub-persentase configurable
- Contoh konfigurasi default:

```
shu_distribution_config:
  cadangan_pct:               25.00   # min 25%
  jasa_anggota_pct:           40.00
  jasa_anggota_split:
    jasa_modal_pct:           50.00   # 50% dari jasa_anggota
    jasa_usaha_pct:           50.00   # 50% dari jasa_anggota
  dana_pengurus_pct:          10.00
  dana_karyawan_pct:           5.00
  dana_pendidikan_pct:        10.00
  dana_sosial_pct:             5.00
  dana_pembangunan_daerah_pct: 5.00
  # Total: 100%
```

### 3. Jasa Modal Calculation — Berdasarkan Simpanan

Jasa Modal dibagikan **proporsional terhadap rata-rata simpanan** setiap anggota selama tahun buku.

```
Jasa Modal Pool = SHU Neto × jasa_anggota_pct × jasa_modal_pct

Per anggota:
┌─────────────────────────────────────────────┐
│ rata_rata_simpanan_anggota                   │
│ = SUM(saldo_harian_simpanan) / jumlah_hari  │
│                                              │
│ Simpanan yang dihitung:                      │
│ ├── Simpanan Pokok                           │
│ ├── Simpanan Wajib                           │
│ └── Tabungan Sukarela                        │
│     (Deposito opsional, configurable)        │
└──────────────────────────────────────────────┘

share_anggota = rata_rata_simpanan_anggota
              / total_rata_rata_simpanan_semua_anggota
              × jasa_modal_pool
```

**Aturan:**
- Rata-rata simpanan dihitung **harian** (daily weighted average) selama tahun buku — lebih akurat daripada rata-rata bulanan
- Simpanan yang dihitung: **Simpanan Pokok + Simpanan Wajib + Tabungan Sukarela** — configurable per tenant
- Deposito bisa di-include atau exclude dari perhitungan — configurable via `include_deposito_in_jasa_modal`
- Anggota yang baru bergabung di pertengahan tahun mendapat proporsi **sesuai hari aktifnya** — simpanan dihitung sejak tanggal join
- Anggota yang keluar di pertengahan tahun mendapat proporsi **sampai tanggal keluar** — simpanan dihitung sampai exit_date
- Jasa Modal = 0 jika anggota tidak memiliki simpanan apapun

### 4. Jasa Usaha Calculation — Berdasarkan Volume Transaksi

Jasa Usaha dibagikan **proporsional terhadap total volume transaksi** setiap anggota selama tahun buku.

```
Jasa Usaha Pool = SHU Neto × jasa_anggota_pct × jasa_usaha_pct

Per anggota:
┌─────────────────────────────────────────────┐
│ total_transaksi_anggota                      │
│ = SUM(|amount|) dari semua transaksi         │
│   rekening anggota selama tahun buku         │
│                                              │
│ Transaksi yang dihitung:                     │
│ ├── Setoran tabungan                         │
│ ├── Penarikan tabungan                       │
│ ├── Angsuran pinjaman (pokok + bunga/margin) │
│ ├── Pencairan pinjaman                       │
│ └── Setoran/penarikan deposito               │
│     (configurable)                           │
└──────────────────────────────────────────────┘

share_anggota = total_transaksi_anggota
              / total_transaksi_semua_anggota
              × jasa_usaha_pool
```

**Aturan:**
- Volume transaksi menggunakan **nilai absolut** — baik setoran maupun penarikan dihitung
- Jenis transaksi yang dihitung **configurable per tenant** via `jasa_usaha_transaction_types[]`
- Transaksi otomatis (biaya admin, bunga/bagi hasil bulanan) bisa di-include atau exclude — configurable
- Anggota yang baru bergabung/keluar di pertengahan tahun hanya dihitung transaksi **selama masa keanggotaan aktif**
- Jasa Usaha = 0 jika anggota tidak memiliki transaksi apapun

### 5. Distribution Percentages — Konfigurasi & Validasi

Persentase distribusi disimpan sebagai konfigurasi tenant dan divalidasi ketat.

```
shu_config
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant) UNIQUE
│
├── ── Persentase Distribusi ──
├── cadangan_pct          NUMERIC(5,2) NOT NULL
│                         CHECK (cadangan_pct >= 25.00)
├── jasa_anggota_pct      NUMERIC(5,2) NOT NULL
├── jasa_modal_split_pct  NUMERIC(5,2) NOT NULL
│                         (% dari jasa_anggota untuk jasa modal)
├── jasa_usaha_split_pct  NUMERIC(5,2) NOT NULL
│                         (% dari jasa_anggota untuk jasa usaha)
│                         CHECK (jasa_modal_split_pct + jasa_usaha_split_pct = 100.00)
├── dana_pengurus_pct     NUMERIC(5,2) NOT NULL
├── dana_karyawan_pct     NUMERIC(5,2) NOT NULL
├── dana_pendidikan_pct   NUMERIC(5,2) NOT NULL
├── dana_sosial_pct       NUMERIC(5,2) NOT NULL
├── dana_pembangunan_pct  NUMERIC(5,2) NOT NULL
│   CHECK (cadangan_pct + jasa_anggota_pct + dana_pengurus_pct
│         + dana_karyawan_pct + dana_pendidikan_pct
│         + dana_sosial_pct + dana_pembangunan_pct = 100.00)
│
├── ── Opsi Perhitungan ──
├── include_deposito_in_jasa_modal  BOOLEAN DEFAULT false
├── jasa_usaha_transaction_types    JSONB NOT NULL DEFAULT '["deposit","withdrawal","installment","disbursement"]'
├── include_auto_transactions       BOOLEAN DEFAULT false
│
├── ── Distribusi ──
├── distribution_method    ENUM (credit_tabungan, separate_payout) DEFAULT 'credit_tabungan'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Validasi saat save:**

```
Save Pre-check:
├── cadangan_pct >= 25.00                                    ✓
├── SUM semua *_pct = 100.00                                 ✓
├── jasa_modal_split_pct + jasa_usaha_split_pct = 100.00     ✓
├── Semua persentase >= 0                                    ✓
└── Tidak ada persentase > 100.00                            ✓
    │
    ALL PASS → Save allowed
    ANY FAIL → Save blocked, return error list
```

### 6. SHU untuk Siswa/Santri

Siswa dan santri yang menjadi **anggota koperasi** (memiliki record nasabah active) berhak mendapatkan SHU.

**Aturan:**
- Hak SHU berlaku **sama** untuk semua anggota — tidak ada diskriminasi berdasarkan `school_relation_type`
- Siswa yang memiliki simpanan (pokok + wajib + tabungan) mendapat **Jasa Modal** proporsional
- Siswa yang aktif bertransaksi mendapat **Jasa Usaha** proporsional
- Distribusi SHU siswa masuk ke **rekening tabungan siswa** (credit_tabungan) atau ditahan sampai diklaim via orang tua/wali (configurable)
- Jika siswa sudah lulus/keluar sebelum distribusi, SHU tetap dihitung berdasarkan masa aktif dan masuk ke **rekening terakhir** atau ditahan sebagai kewajiban
- Notifikasi SHU dikirim ke orang tua/wali jika nasabah adalah siswa minor

### 7. Approval Flow — Calculation → RAT → Distribution

SHU melalui proses approval bertahap sebelum didistribusikan.

```
Year-End Closing selesai (K015)
        │
        v
┌──────────────────────────┐
│ Status: CALCULATED       │  Sistem menghitung otomatis
│ - SHU bruto              │
│ - Breakdown per komponen  │
│ - Simulasi per anggota   │
└────────┬─────────────────┘
         │ review by Pengurus
         v
┌──────────────────────────┐
│ Status: REVIEWED         │  Pengurus sudah review & setuju
│ Manager+ approve         │
└────────┬─────────────────┘
         │ RAT approval
         v
┌──────────────────────────┐
│ Status: APPROVED         │  Disahkan di RAT
│ Admin input RAT decision │
│ (bisa ubah persentase    │
│  jika RAT memutuskan)    │
└────────┬─────────────────┘
         │ execute distribution
         v
┌──────────────────────────┐
│ Status: DISTRIBUTED      │  SHU sudah masuk ke rekening anggota
│ Irreversible             │  Jurnal distribusi tercatat
└──────────────────────────┘
```

**Aturan:**
- **CALCULATED** → otomatis setelah year-end closing. Status awal.
- **REVIEWED** → Manager+ meng-approve setelah review angka SHU dan breakdown
- **APPROVED** → Admin mencatat keputusan RAT. RAT bisa **mengubah persentase** dari konfigurasi default — perubahan dicatat sebagai override untuk tahun tersebut
- **DISTRIBUTED** → Admin menjalankan distribusi. Proses ini menghasilkan:
  - Jurnal distribusi SHU (K015) — debit 3500 SHU, credit akun tujuan per komponen
  - Credit ke rekening tabungan anggota (untuk Jasa Modal + Jasa Usaha)
  - Record per anggota di `shu_anggota`
- Distribusi bersifat **irreversible** — jika ada kesalahan, harus dikoreksi via jurnal adjustment
- RAT override persentase **hanya berlaku untuk tahun tersebut** — tidak mengubah konfigurasi default tenant

### 8. Distribution Method — Cara Pembayaran

SHU anggota bisa dibayarkan melalui dua metode:

| Metode | Deskripsi | Default |
|--------|-----------|---------|
| `credit_tabungan` | SHU di-credit ke rekening tabungan utama anggota | Ya |
| `separate_payout` | SHU dibayarkan terpisah (cash atau transfer bank) | Tidak |

**Aturan credit_tabungan:**
- Sistem memilih rekening tabungan **pertama** (oldest active) milik anggota
- Jika anggota tidak punya rekening tabungan aktif, SHU ditahan sebagai **kewajiban** (`shu_payable`) sampai anggota membuka rekening atau mengklaim cash
- Setiap credit ke tabungan menghasilkan **transaksi** (K011) dengan type `shu_distribution` — sehingga auto-journal juga tercipta

**Aturan separate_payout:**
- Membutuhkan proses manual — teller mencatat penerimaan SHU oleh anggota
- Setiap payout menghasilkan transaksi kas keluar — auto-journal tercipta

### 9. Islamic Mode — Terminologi & Penyesuaian

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|-------|------------------------|------------------------|
| Nama | SHU (Sisa Hasil Usaha) | SHU / Surplus Usaha |
| Sumber | Pendapatan bunga - Beban | Pendapatan margin/bagi hasil - Beban |
| Cadangan | Cadangan Umum | Cadangan Umum |
| Dana Sosial | Dana Sosial | Dana Sosial + Dana Ta'zir (jika ada saldo ta'zir) |
| Distribusi | Sama | Sama — distribusi SHU terpisah dari distribusi dana zakat (K018) |

**Aturan:**
- Konsep SHU **sama** di kedua mode — yang berbeda hanya sumber pendapatan (bunga vs margin/bagi hasil)
- Di Islamic mode, dana zakat (6xxx) dan dana kebajikan (7xxx) **tidak termasuk** dalam perhitungan SHU — ini dana titipan, bukan pendapatan koperasi
- Dana ta'zir (8xxx) juga **tidak termasuk** SHU — didistribusikan terpisah sebagai dana sosial (K018)
- Jika BMT juga mengeluarkan **zakat institusi** dari laba, ini dihitung sebelum distribusi SHU (mengurangi SHU neto)

### 10. SHU Simulation — Sebelum RAT

Sebelum RAT, sistem menyediakan fitur **simulasi** untuk membantu pengurus melihat dampak dari berbagai skenario persentase.

```
SHU Simulation
        │
        v
┌──────────────────────────────────────┐
│ Input: shu_bruto + skenario %        │
│                                      │
│ Output per skenario:                 │
│ ├── Breakdown per komponen           │
│ │   (cadangan, jasa anggota, dll)    │
│ ├── Top 10 anggota penerima terbesar │
│ ├── Bottom 10 anggota penerima       │
│ ├── Rata-rata SHU per anggota        │
│ ├── Median SHU per anggota           │
│ └── SHU per anggota (detail list)    │
└──────────────────────────────────────┘
```

**Aturan:**
- Simulasi **tidak menyimpan data** — hanya kalkulasi on-the-fly
- Bisa dijalankan **berulang kali** dengan persentase berbeda
- Simulasi membutuhkan status SHU minimal **CALCULATED**
- Output dalam format **Excel** untuk presentasi di RAT
- Minimal role: **Manager+** untuk menjalankan simulasi

### 11. Data Model

```
shu_periode
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Identitas ──
├── tahun_buku            INTEGER NOT NULL UNIQUE per tenant
├── period_start          DATE NOT NULL
├── period_end            DATE NOT NULL
│
├── ── Angka SHU ──
├── total_pendapatan      NUMERIC(15,2) NOT NULL
├── total_beban           NUMERIC(15,2) NOT NULL
├── shu_bruto             NUMERIC(15,2) NOT NULL
│                         (= total_pendapatan - total_beban)
├── pajak                 NUMERIC(15,2) DEFAULT 0
├── zakat_institusi       NUMERIC(15,2) DEFAULT 0
│                         (Islamic mode: zakat dari laba BMT)
├── shu_neto              NUMERIC(15,2) NOT NULL
│                         (= shu_bruto - pajak - zakat_institusi)
│
├── ── Distribusi ──
├── distribution_config   JSONB NOT NULL
│   {
│     "cadangan_pct": 25.00,
│     "jasa_anggota_pct": 40.00,
│     "jasa_modal_split_pct": 50.00,
│     "jasa_usaha_split_pct": 50.00,
│     "dana_pengurus_pct": 10.00,
│     "dana_karyawan_pct": 5.00,
│     "dana_pendidikan_pct": 10.00,
│     "dana_sosial_pct": 5.00,
│     "dana_pembangunan_pct": 5.00
│   }
│
├── ── Hasil Distribusi (denormalized) ──
├── amount_cadangan       NUMERIC(15,2) DEFAULT 0
├── amount_jasa_modal     NUMERIC(15,2) DEFAULT 0
├── amount_jasa_usaha     NUMERIC(15,2) DEFAULT 0
├── amount_dana_pengurus  NUMERIC(15,2) DEFAULT 0
├── amount_dana_karyawan  NUMERIC(15,2) DEFAULT 0
├── amount_dana_pendidikan NUMERIC(15,2) DEFAULT 0
├── amount_dana_sosial    NUMERIC(15,2) DEFAULT 0
├── amount_dana_pembangunan NUMERIC(15,2) DEFAULT 0
├── rounding_difference    NUMERIC(15,2) NOT NULL DEFAULT 0
│                         (selisih pembulatan, dialokasikan ke cadangan)
│
├── ── Status & Workflow ──
├── status                ENUM (calculated, reviewed, approved, distributed)
├── reviewed_at           TIMESTAMPTZ (nullable)
├── reviewed_by           UUID (nullable, FK → user)
├── rat_date              DATE (nullable, tanggal pelaksanaan RAT)
├── rat_minutes_url       VARCHAR (nullable, URL file notulen RAT)
├── approved_at           TIMESTAMPTZ (nullable)
├── approved_by           UUID (nullable, FK → user)
├── distributed_at        TIMESTAMPTZ (nullable)
├── distributed_by        UUID (nullable, FK → user)
│
├── ── Statistik ──
├── total_eligible_members INTEGER DEFAULT 0
├── total_distributed_members INTEGER DEFAULT 0
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
```

```
shu_anggota (detail per anggota)
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── shu_periode_id        UUID (FK → shu_periode) NOT NULL
├── nasabah_id            UUID (FK → nasabah) NOT NULL
│
├── ── Basis Perhitungan ──
├── avg_simpanan          NUMERIC(15,2) NOT NULL
│                         (rata-rata harian simpanan selama tahun buku)
├── total_transaksi       NUMERIC(15,2) NOT NULL
│                         (total volume transaksi selama tahun buku)
├── active_days           INTEGER NOT NULL
│                         (jumlah hari aktif sebagai anggota di tahun tersebut)
│
├── ── Hasil Perhitungan ──
├── jasa_modal            NUMERIC(15,2) NOT NULL DEFAULT 0
├── jasa_usaha            NUMERIC(15,2) NOT NULL DEFAULT 0
├── total_shu             NUMERIC(15,2) NOT NULL DEFAULT 0
│                         (= jasa_modal + jasa_usaha)
│
├── ── Distribusi ──
├── distribution_method   ENUM (credit_tabungan, separate_payout, pending)
├── target_rekening_id    UUID (nullable, FK → rekening)
├── distributed_at        TIMESTAMPTZ (nullable)
├── transaction_id        UUID (nullable, FK → transaksi, ref ke transaksi distribusi)
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
```

**Unique constraint:** `(tenant_id, shu_periode_id, nasabah_id)` — satu anggota satu record per periode SHU.

### 12. Vernon _rels dan _data Structure

**_rels (shu_periode):**
```json
{
  "tenant_id": "018f..."
}
```

**_data (shu_periode):**
```json
{
  "tenant": {
    "id":   "018f...",
    "name": "Koperasi Sekolah Harapan"
  }
}
```

**_rels (shu_anggota):**
```json
{
  "tenant_id":      "018f...",
  "shu_periode_id": "018f...",
  "nasabah_id":     "018f...",
  "target_rekening_id": "018f..."
}
```

**_data (shu_anggota):**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "shu_periode": {
    "id":          "018f...",
    "tahun_buku":  2026
  },
  "target_rekening": {
    "id":             "018f...",
    "account_number": "TB-2026-JKT-00000001",
    "account_name":   "Tabungan Berkah - Ahmad Fauzi"
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.nasabah` di semua `shu_anggota` nasabah tersebut
- `RekeningUpdatedEvent` → update `_data.target_rekening` di semua `shu_anggota` yang mereferensikan rekening tersebut

### 13. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View SHU periode | - | v | v | v |
| View SHU per anggota (detail) | - | v | v | v |
| Trigger SHU calculation | - | - | - | v |
| Review SHU (CALCULATED → REVIEWED) | - | - | v | v |
| Approve SHU (REVIEWED → APPROVED, input RAT) | - | - | - | v |
| Execute distribution (APPROVED → DISTRIBUTED) | - | - | - | v |
| Run SHU simulation | - | - | v | v |
| Configure SHU percentages | - | - | - | v |
| View SHU statement (per anggota, self) | v | v | v | v |

**Catatan:**
- SHU calculation, approval, dan distribution hanya **Admin** — operasi setahun sekali yang berdampak besar
- Review oleh **Manager+** — memastikan angka sudah benar sebelum dibawa ke RAT
- Teller bisa melihat **SHU statement** per anggota — untuk informasi ke nasabah yang bertanya
- Simulation oleh **Manager+** — untuk persiapan materi RAT

### 14. Reporting — Laporan SHU

Sistem menghasilkan beberapa laporan terkait SHU:

```
Laporan SHU:
├── Laporan SHU Ringkasan
│   ├── Total SHU bruto & neto
│   ├── Breakdown per komponen distribusi
│   └── Persentase yang diterapkan
│
├── Laporan SHU per Anggota (detail)
│   ├── Nama, nomor anggota
│   ├── Rata-rata simpanan
│   ├── Total transaksi
│   ├── Jasa Modal yang diterima
│   ├── Jasa Usaha yang diterima
│   └── Total SHU yang diterima
│
├── Statement SHU Individual
│   ├── Format: slip/kwitansi per anggota
│   ├── Detail perhitungan pribadi
│   └── Metode distribusi
│
└── Laporan Perbandingan SHU Tahunan
    ├── SHU tahun ini vs tahun lalu
    ├── Tren 5 tahun terakhir
    └── Jumlah anggota eligible vs distributed
```

**Aturan:**
- Laporan SHU Ringkasan wajib ada untuk **materi RAT**
- Statement individual bisa dicetak per anggota — untuk transparansi
- Format: PDF (formal), Excel (analisis)
- Laporan perbandingan tahunan membantu pengurus melihat tren kinerja koperasi

## Consequences

### Positif

- **Otomatis** — perhitungan SHU dari data akuntansi, bukan input manual
- **Transparan** — setiap anggota bisa melihat basis perhitungan (simpanan, transaksi) dan hasilnya
- **Configurable** — persentase distribusi bisa disesuaikan per AD/ART, RAT bisa override
- **Inklusif** — siswa/santri sebagai anggota juga mendapat SHU proporsional
- **Auditable** — approval flow bertahap (calculate → review → RAT → distribute) dengan audit trail
- **Simulasi** — pengurus bisa menyiapkan skenario sebelum RAT, keputusan lebih informatif

### Negatif

- **Perhitungan berat** — daily weighted average simpanan untuk semua anggota butuh kalkulasi intensif
- **Year-end dependency** — SHU tidak bisa dihitung sebelum year-end closing selesai (K015)
- **RAT bottleneck** — distribusi tergantung pelaksanaan RAT yang bisa tertunda
- **Pembulatan** — distribusi proporsional bisa menghasilkan selisih pembulatan (SUM per anggota != total pool)
- **Complexity siswa** — siswa yang lulus/keluar perlu penanganan khusus untuk SHU yang belum didistribusikan

### Mitigasi

- Perhitungan daily weighted average menggunakan **pre-aggregated data** — snapshot saldo harian sudah tersedia dari reconciliation K015
- Year-end closing memiliki checklist dan pre-check yang jelas — meminimalkan delay
- SHU yang belum didistribusikan (siswa keluar, rekening tutup) dicatat sebagai **kewajiban** — tidak hilang, bisa diklaim kapan saja
- **Pembulatan**: selisih pembulatan dialokasikan ke **cadangan** — documented, auditable (lihat algoritma pembulatan di bawah)

**Algoritma Pembulatan SHU:**

Distribusi proporsional ke ratusan anggota menghasilkan selisih pembulatan. Aturan:

1. Hitung SHU per anggota dengan presisi penuh (NUMERIC(15,2))
2. **Floor** (bulatkan ke bawah) ke Rp 1 terdekat per anggota
3. Hitung remainder: `pool_amount - SUM(individual_floored)`
4. Remainder dialokasikan ke **Cadangan** (bukan ke anggota tertentu)
5. Simpan `rounding_difference` di `shu_periode` untuk audit

Contoh:
```
Pool Jasa Modal = Rp 10.000.000
Anggota A: proporsi 33.333...% → Rp 3.333.333 (floor)
Anggota B: proporsi 33.333...% → Rp 3.333.333 (floor)  
Anggota C: proporsi 33.333...% → Rp 3.333.333 (floor)
SUM = Rp 9.999.999
Remainder = Rp 1 → masuk Cadangan
rounding_difference = Rp 1
```

Tambahkan field di data model `shu_periode`:
```
rounding_difference   NUMERIC(15,2) NOT NULL DEFAULT 0
```
- Simulasi membantu mempercepat keputusan RAT — pengurus datang dengan data, bukan diskusi tanpa angka

## Alternatives Considered

### A. SHU Manual (tanpa perhitungan otomatis)

Pengurus menghitung SHU di Excel, input hasil ke sistem untuk distribusi.

**Ditolak** karena: rawan human error, tidak scalable untuk koperasi dengan ratusan anggota, dan tidak bisa audit trail dari proses perhitungan. Otomatisasi memastikan konsistensi formula dan transparansi.

### B. SHU Bulanan (bukan tahunan)

Menghitung dan mendistribusikan SHU setiap bulan.

**Ditolak** karena: tidak sesuai UU Koperasi yang mengatur SHU sebagai surplus tahunan yang disahkan di RAT. Distribusi bulanan juga berisiko — jika di akhir tahun terjadi kerugian, SHU yang sudah dibagikan harus di-clawback.

### C. Flat Distribution (sama rata)

Setiap anggota mendapat SHU yang sama, tanpa memperhitungkan simpanan dan transaksi.

**Ditolak** karena: tidak adil dan tidak sesuai prinsip koperasi. UU Koperasi secara eksplisit mengatur bahwa SHU dibagikan berdasarkan **jasa modal** (kontribusi simpanan) dan **jasa usaha** (kontribusi transaksi). Flat distribution tidak memberikan insentif bagi anggota untuk aktif.

### D. SHU Hanya untuk Jasa Modal (tanpa Jasa Usaha)

Distribusi SHU hanya berdasarkan besaran simpanan, mengabaikan volume transaksi.

**Ditolak** karena: menghilangkan insentif bagi anggota untuk aktif bertransaksi. Jasa Usaha mendorong anggota menggunakan layanan koperasi (menabung, meminjam) yang pada akhirnya meningkatkan pendapatan koperasi. Kedua komponen penting untuk keadilan distribusi.
