# ADR-K014: Kas & Cash Flow

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Kas adalah posisi uang tunai dan dana yang dimiliki koperasi/BMT pada suatu waktu tertentu. Cash flow mencatat seluruh pergerakan kas masuk dan keluar. Sistem harus mengakomodasi:

- **Daily cash position**: saldo kas harian per branch yang akurat dan reconcilable
- **Cash flow tracking**: pencatatan inflow dan outflow berdasarkan kategori
- **Non-rekening transactions**: pengeluaran operasional (gaji, ATK, listrik) yang tidak terkait rekening nasabah
- **Inter-branch cash transfer**: perpindahan kas fisik antar cabang
- **Reconciliation**: kas harian harus cocok dengan total teller sessions ([ADR-K012](./ADR-K012-teller-session.md))
- **Dual-mode terminologi**: terminologi kas **identik** di kedua mode — lihat [ADR-009](../core/ADR-009-dual-mode-institution-type.md)
- **Link to accounting**: setiap pergerakan kas menghasilkan jurnal (K015)

## Decision

### 1. Daily Cash Position

Posisi kas harian dihitung per branch per hari:

```
Daily Cash Position Formula:
┌─────────────────────────────────────────────────┐
│ Saldo Awal (Opening Balance)                    │
│   = Saldo akhir hari sebelumnya                 │
│                                                 │
│ + Total Inflow                                  │
│   = Setoran tunai nasabah                       │
│   + Angsuran tunai pinjaman                     │
│   + Kas masuk operasional                       │
│   + Transfer kas masuk dari cabang lain          │
│   + Penerimaan lain-lain                        │
│                                                 │
│ - Total Outflow                                 │
│   = Penarikan tunai nasabah                     │
│   + Pencairan pinjaman tunai                    │
│   + Kas keluar operasional                      │
│   + Transfer kas keluar ke cabang lain           │
│   + Pengeluaran lain-lain                       │
│                                                 │
│ = Saldo Akhir (Closing Balance)                 │
└─────────────────────────────────────────────────┘
```

**Aturan:**
- Saldo awal hari ini = saldo akhir kemarin — chain yang tidak boleh putus
- Saldo awal hari pertama di-set manual saat branch onboarding
- Posisi kas dihitung **otomatis** dari transaksi hari tersebut — bukan input manual
- Posisi kas bersifat **derived** — bisa dihitung ulang dari data transaksi
- Saldo akhir **tidak boleh negatif** — jika kas mendekati 0, alert dikirim ke Supervisor/Manager

### 2. Cash Flow Categories

Seluruh pergerakan kas dikelompokkan dalam kategori terstruktur:

```
INFLOW (Kas Masuk):
├── SETORAN_NASABAH
│   ├── Setoran tunai tabungan
│   ├── Setoran tunai simpanan pokok/wajib
│   └── Setoran tunai deposito
├── ANGSURAN_PINJAMAN
│   └── Pembayaran angsuran tunai
├── PENERIMAAN_OPERASIONAL
│   ├── Pendapatan administrasi (tunai)
│   ├── Pendapatan denda (tunai)
│   └── Pendapatan lain-lain
├── TRANSFER_KAS_MASUK
│   ├── Dari cabang lain
│   └── Dari kantor pusat
└── LAIN_LAIN_MASUK
    └── Penerimaan tidak terkategorikan

OUTFLOW (Kas Keluar):
├── PENARIKAN_NASABAH
│   ├── Penarikan tunai tabungan
│   ├── Penarikan tunai simpanan (saat keluar)
│   └── Pencairan deposito tunai
├── PENCAIRAN_PINJAMAN
│   └── Disbursement pinjaman tunai
├── BIAYA_OPERASIONAL
│   ├── Gaji karyawan / honor teller
│   ├── ATK (alat tulis kantor)
│   ├── Listrik, air, internet
│   ├── Sewa tempat
│   ├── Transport / perjalanan dinas
│   └── Biaya operasional lain
├── TRANSFER_KAS_KELUAR
│   ├── Ke cabang lain
│   └── Ke kantor pusat
└── LAIN_LAIN_KELUAR
    └── Pengeluaran tidak terkategorikan
```

**Aturan:**
- Kategori inflow dari transaksi rekening (setoran, angsuran) **otomatis** terklasifikasi dari tipe transaksi ([K011](./ADR-K011-transaksi.md))
- Kategori biaya operasional diinput **manual** oleh Supervisor+ sebagai transaksi non-rekening
- Kategori `LAIN_LAIN` membutuhkan deskripsi wajib — tidak boleh kosong
- Kategori bisa di-extend oleh Admin melalui konfigurasi — menambah sub-kategori baru di bawah kategori utama
- Internal movement (transfer antar rekening, bagi hasil) **tidak** mempengaruhi kas fisik — tidak masuk cash flow

### 3. Non-Rekening Transactions

Pengeluaran dan penerimaan operasional yang tidak terkait rekening nasabah:

```
Non-Rekening Transaction Flow:
        │
        v
┌───────────────────────────────────────┐
│ 1. Supervisor/Manager buat entry      │
│ 2. Pilih kategori (gaji, ATK, dll)    │
│ 3. Input nominal + keterangan         │
│ 4. Lampirkan bukti (nota, kwitansi)   │
│ 5. Approval (jika > threshold)        │
│ 6. Dieksekusi → update kas harian     │
│ 7. Auto-generate jurnal entry (K015)  │
└───────────────────────────────────────┘
```

**Contoh transaksi non-rekening:**

| Jenis | Kategori | Contoh | Approval |
|-------|----------|--------|----------|
| Kas Keluar | Gaji | Honor teller bulan April | Manager+ |
| Kas Keluar | ATK | Beli kertas dan tinta printer | Supervisor+ |
| Kas Keluar | Listrik | Bayar tagihan listrik | Supervisor+ |
| Kas Masuk | Sewa | Sewa ruangan ke pihak ketiga | Supervisor+ |
| Kas Masuk | Lain-lain | Pengembalian kelebihan bayar | Supervisor+ |

**Aturan:**
- Non-rekening transactions dicatat di tabel `kas_transaksi` (lihat [K011 §12](./ADR-K011-transaksi.md))
- Setiap pengeluaran operasional **wajib** punya bukti/nota (attachment_url)
- Pengeluaran di bawah threshold minor bisa approved oleh **Supervisor**
- Pengeluaran di atas threshold minor membutuhkan **Manager+** approval
- Threshold configurable per tenant (default minor: Rp 500.000, default major: Rp 5.000.000)
- Non-rekening transaction **tetap** dikaitkan dengan teller session jika melibatkan kas fisik

### 4. Inter-Branch Cash Transfer

Transfer kas fisik antara cabang dicatat di kedua sisi:

```
Inter-Branch Cash Transfer Flow:

Branch A (Pengirim)              Branch B (Penerima)
        │                               │
        v                               │
┌─────────────────┐                     │
│ Supervisor/Mgr  │                     │
│ buat request    │                     │
│ transfer        │                     │
└────────┬────────┘                     │
         │                              │
         v                              │
┌─────────────────┐                     │
│ Manager approve │                     │
│ (branch A)      │                     │
└────────┬────────┘                     │
         │                              │
         v                              │
┌─────────────────┐                     │
│ Catat KAS_KELUAR│                     │
│ di Branch A     │                     │
│ + detail        │                     │
│   denominasi    │                     │
└────────┬────────┘                     │
         │ fisik transfer               │
         └──────────────────┐           │
                            v           v
                    ┌─────────────────────┐
                    │ Branch B terima     │
                    │ Hitung + konfirmasi │
                    │ denominasi          │
                    │ Catat KAS_MASUK     │
                    └─────────────────────┘
```

**Aturan:**
- Transfer membutuhkan **Manager+ approval** di branch pengirim
- Branch penerima **konfirmasi** penerimaan + hitung denominasi
- Kedua sisi mencatat transaksi masing-masing (kas_keluar di pengirim, kas_masuk di penerima)
- Kedua transaksi dihubungkan oleh `transfer_kas_pair_id` ([K011 §12](./ADR-K011-transaksi.md))
- **Selisih** antara nominal kirim dan terima membutuhkan investigation
- Denominasi dicatat di kedua sisi — harus match
- Transfer tidak bisa dilakukan jika branch pengirim akan memiliki saldo di bawah `kas_minimum`
- Status transfer: `initiated → in_transit → received → confirmed` (atau `disputed`)

### 5. Cash Flow Report

Laporan cash flow tersedia dalam beberapa periode:

```
Cash Flow Report Structure:
┌──────────────────────────────────────────────────┐
│ LAPORAN KAS HARIAN                               │
│ Cabang: Jakarta Pusat                            │
│ Tanggal: 15 April 2026                           │
│                                                  │
│ Saldo Awal:                    Rp  35.000.000    │
│                                                  │
│ KAS MASUK:                                       │
│ ├── Setoran Nasabah            Rp  25.500.000    │
│ ├── Angsuran Pinjaman          Rp   8.200.000    │
│ ├── Penerimaan Operasional     Rp     500.000    │
│ └── Transfer Masuk             Rp  10.000.000    │
│     TOTAL MASUK                Rp  44.200.000    │
│                                                  │
│ KAS KELUAR:                                      │
│ ├── Penarikan Nasabah          Rp  18.000.000    │
│ ├── Pencairan Pinjaman         Rp  15.000.000    │
│ ├── Biaya Operasional          Rp   2.500.000    │
│ │   ├── Gaji/Honor             Rp   1.500.000    │
│ │   ├── ATK                    Rp     300.000    │
│ │   ├── Listrik                Rp     500.000    │
│ │   └── Lain-lain              Rp     200.000    │
│ └── Transfer Keluar            Rp   5.000.000    │
│     TOTAL KELUAR               Rp  40.500.000    │
│                                                  │
│ Saldo Akhir:                   Rp  38.700.000    │
│                                                  │
│ Selisih Teller:                Rp       -500     │
│ (dari konsolidasi K012)                          │
└──────────────────────────────────────────────────┘
```

**Periode laporan:**

| Periode | Deskripsi | Auto-generate |
|---------|-----------|---------------|
| Harian | Per branch per hari | Ya, saat semua teller session closed |
| Mingguan | Agregasi 7 hari per branch | On-demand |
| Bulanan | Agregasi 1 bulan per branch | Ya, akhir bulan |
| Custom | Range tanggal custom | On-demand |

**Aturan:**
- Laporan harian di-generate **otomatis** setelah semua teller session di branch closed
- Laporan mingguan dan custom di-generate **on-demand** (query dari data harian)
- Laporan bulanan di-generate **otomatis** di akhir bulan sebagai checkpoint
- Semua laporan bisa di-export ke **CSV** dan **PDF**
- Laporan menampilkan breakdown per kategori dan per sub-kategori

### 6. Kas Minimum (Minimum Cash)

Setiap branch memiliki batas minimum kas yang harus dijaga:

```
kas_minimum_config:
  enabled: true
  minimum_amount: 10000000     # Rp 10.000.000 per branch (configurable)
  warning_threshold: 15000000  # Alert saat mendekati minimum
  alert_recipients:            # Siapa yang menerima alert
    - supervisor
    - manager
```

**Aturan:**
- Minimum kas **configurable per branch** — cabang besar bisa set lebih tinggi
- **Warning alert** dikirim saat kas mendekati threshold warning (sebelum minimum)
- **Critical alert** dikirim saat kas di bawah minimum
- Sistem **tidak memblokir** transaksi saat kas di bawah minimum — hanya alert
- Alasan: memblokir penarikan nasabah karena kas minimum tidak fair bagi nasabah
- Manager harus segera arrange top-up kas (transfer dari cabang lain/pusat) saat kas rendah
- Monitoring kas minimum masuk ke dashboard Manager dan Admin

### 7. Reconciliation — Kas vs Teller Sessions

Kas harian harus reconcilable dengan total dari teller sessions:

```
Reconciliation Check:
┌─────────────────────────────────────────────────┐
│ Source 1: Kas Harian (K014)                     │
│ closing_balance = opening_balance               │
│                 + total_inflow                   │
│                 - total_outflow                  │
│                                                 │
│ Source 2: Teller Sessions (K012)                │
│ total_actual_cash = SUM(actual_cash)            │
│                     semua teller sessions        │
│                     + vault_end_of_day           │
│                                                 │
│ Reconciliation:                                 │
│ closing_balance = total_actual_cash             │
│                 + non_cash_balance (bank, dll)   │
│                 + selisih_teller                 │
│                                                 │
│ Match     → reconciled                          │
│ Not match → investigation required              │
└─────────────────────────────────────────────────┘
```

**Reconciliation schedule:**
- **Harian**: otomatis setelah semua teller session closed — bandingkan kas harian vs sum teller sessions
- **Mingguan**: Supervisor review — spot check random hari
- **Bulanan**: Manager sign-off — laporan kas bulanan di-approve formal

**Aturan:**
- Reconciliation harian berjalan **otomatis** sebagai scheduled job
- Hasil reconciliation disimpan di `kas_harian` record
- Selisih reconciliation membutuhkan **investigation dan penjelasan**
- Selisih berulang pada cabang yang sama membutuhkan **audit oleh Admin**
- Reconciliation menjadi input untuk laporan bulanan ke pengawas koperasi

### 8. Link to Accounting

Setiap pergerakan kas menghasilkan journal entry untuk pembukuan:

```
Cash Movement → Journal Entry (K015):

Setoran Tunai Rp 500.000:
├── Debit:  Kas (1-1100)           500.000
└── Credit: Tabungan Nasabah       500.000

Penarikan Tunai Rp 200.000:
├── Debit:  Tabungan Nasabah       200.000
└── Credit: Kas (1-1100)           200.000

Bayar Gaji Rp 1.500.000:
├── Debit:  Beban Gaji (5-1100)  1.500.000
└── Credit: Kas (1-1100)         1.500.000

Transfer Kas Cabang A → B Rp 10.000.000:
Branch A:
├── Debit:  Kas Dalam Transit    10.000.000
└── Credit: Kas (1-1100)        10.000.000
Branch B (saat terima):
├── Debit:  Kas (1-1100)        10.000.000
└── Credit: Kas Dalam Transit   10.000.000
```

**Aturan:**
- Journal entry di-generate **otomatis** saat transaksi committed — tidak perlu input manual
- Mapping tipe transaksi ke akun COA dikonfigurasi per tenant (di K015)
- Setiap transaksi kas memiliki `journal_entry_id` sebagai referensi
- Journal entry mengikuti prinsip **double-entry accounting** — debit selalu = credit
- Detail implementasi jurnal di [ADR-K015](./ADR-K015-jurnal-coa.md)

### 9. Multi-Branch Consolidation

HQ/Admin bisa melihat posisi kas seluruh cabang:

```
Multi-Branch Cash Position:
┌──────────────────────────────────────────────────┐
│ POSISI KAS SELURUH CABANG                        │
│ Tanggal: 15 April 2026                           │
│                                                  │
│ ┌──────────────┬───────────┬───────────┬───────┐ │
│ │ Cabang       │ Saldo Awal│ Saldo Akhir│Status│ │
│ ├──────────────┼───────────┼───────────┼───────┤ │
│ │ Jakarta Pusat│35.000.000 │38.700.000 │  OK   │ │
│ │ Bandung      │20.000.000 │22.300.000 │  OK   │ │
│ │ Surabaya     │15.000.000 │ 9.500.000 │ LOW!  │ │
│ │ Yogyakarta   │25.000.000 │27.100.000 │  OK   │ │
│ ├──────────────┼───────────┼───────────┼───────┤ │
│ │ TOTAL        │95.000.000 │97.600.000 │       │ │
│ └──────────────┴───────────┴───────────┴───────┘ │
│                                                  │
│ Alert: Surabaya di bawah kas minimum!            │
└──────────────────────────────────────────────────┘
```

**Aturan:**
- Consolidation view tersedia untuk **Manager+ dan Admin**
- Status per branch: `OK` (di atas warning threshold), `LOW` (di bawah warning), `CRITICAL` (di bawah minimum)
- Drill-down ke detail per branch tersedia
- Data real-time (dari kas_harian terbaru masing-masing branch)
- Inter-branch transfer tidak menambah total consolidated — hanya redistribusi antar branch

### 10. Data Model — Kas Harian

```
kas_harian
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant) NOT NULL
├── branch_id             UUID (FK → branch) NOT NULL
│
├── ── Identitas ──
├── report_date           DATE NOT NULL
│
├── ── Saldo ──
├── opening_balance       NUMERIC(15,2) NOT NULL
├── closing_balance       NUMERIC(15,2) NOT NULL
│
├── ── Inflow ──
├── total_inflow          NUMERIC(15,2) NOT NULL DEFAULT 0
├── inflow_setoran        NUMERIC(15,2) NOT NULL DEFAULT 0
├── inflow_angsuran       NUMERIC(15,2) NOT NULL DEFAULT 0
├── inflow_operasional    NUMERIC(15,2) NOT NULL DEFAULT 0
├── inflow_transfer       NUMERIC(15,2) NOT NULL DEFAULT 0
├── inflow_lainnya        NUMERIC(15,2) NOT NULL DEFAULT 0
│
├── ── Outflow ──
├── total_outflow         NUMERIC(15,2) NOT NULL DEFAULT 0
├── outflow_penarikan     NUMERIC(15,2) NOT NULL DEFAULT 0
├── outflow_pencairan     NUMERIC(15,2) NOT NULL DEFAULT 0
├── outflow_operasional   NUMERIC(15,2) NOT NULL DEFAULT 0
├── outflow_transfer      NUMERIC(15,2) NOT NULL DEFAULT 0
├── outflow_lainnya       NUMERIC(15,2) NOT NULL DEFAULT 0
│
├── ── Teller Summary ──
├── total_teller_sessions INTEGER NOT NULL DEFAULT 0
├── total_teller_variance NUMERIC(15,2) NOT NULL DEFAULT 0
│
├── ── Reconciliation ──
├── is_reconciled         BOOLEAN NOT NULL DEFAULT false
├── reconciled_at         TIMESTAMPTZ (nullable)
├── reconciled_by         UUID (nullable, FK → user)
├── reconciliation_note   TEXT (nullable)
│
├── ── Status ──
├── status                ENUM (open, closed, reconciled) NOT NULL
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

**Database constraints:**
- `UNIQUE (tenant_id, branch_id, report_date)` — satu record per branch per hari
- `CHECK (closing_balance = opening_balance + total_inflow - total_outflow)`
- `CHECK (total_inflow = inflow_setoran + inflow_angsuran + inflow_operasional + inflow_transfer + inflow_lainnya)`
- `CHECK (total_outflow = outflow_penarikan + outflow_pencairan + outflow_operasional + outflow_transfer + outflow_lainnya)`
- Index: `(tenant_id, branch_id, report_date DESC)`
- Index: `(tenant_id, report_date)` — untuk consolidation query

### 11. Vernon _rels dan _data Structure

**_rels:**
```json
{
  "tenant_id": "018f...",
  "branch_id": "018f..."
}
```

**_data:**
```json
{
  "branch": {
    "id":   "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  }
}
```

**SyncEngine triggers:**
- `BranchUpdatedEvent` → update `_data.branch` di semua kas_harian cabang tersebut

### 12. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View kas harian (own branch) | v* | v | v | v |
| View kas harian (all branches) | - | - | v | v |
| Create non-rekening transaction (minor) | - | v | v | v |
| Create non-rekening transaction (major) | - | - | v | v |
| Approve non-rekening transaction | - | - | v | v |
| Initiate inter-branch transfer | - | v | v | v |
| Approve inter-branch transfer | - | - | v | v |
| Confirm inter-branch receipt | - | v | v | v |
| View cash flow report | - | v | v | v |
| Export cash flow report | - | v | v | v |
| Reconcile kas harian | - | - | v | v |
| View multi-branch consolidation | - | - | v | v |
| Configure kas minimum | - | - | - | v |

`*` Teller hanya melihat summary kas harian — tidak detail pengeluaran operasional

**Catatan:**
- Teller **tidak bisa** membuat transaksi non-rekening — mencegah pengeluaran tidak sah
- Non-rekening transaction minor (≤ threshold) bisa dibuat oleh Supervisor, di atasnya butuh Manager
- Inter-branch transfer membutuhkan Manager approval di sisi pengirim
- Reconciliation formal hanya Manager+ — Supervisor bisa view tapi tidak bisa sign-off
- Konfigurasi kas minimum hanya Admin — setting yang jarang berubah

### 13. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Kas Harian | Kas Harian | Kas Harian |
| Cash Flow | Arus Kas | Arus Kas |
| Biaya Operasional | Biaya Operasional | Biaya Operasional |
| Laporan Kas | Laporan Kas | Laporan Kas |
| Transfer Kas | Transfer Kas | Transfer Kas |
| Saldo Kas | Saldo Kas | Saldo Kas |

Terminologi kas dan cash flow **identik** di kedua mode — tidak ada perbedaan syariah pada pengelolaan kas operasional.

## Consequences

### Positif

- **Complete cash visibility** — posisi kas harian terlacak per branch dengan breakdown kategori
- **Reconcilable** — kas harian selalu bisa di-cross-check dengan teller sessions
- **Operational expense tracking** — pengeluaran operasional tercatat dan ter-approve
- **Inter-branch control** — transfer kas antar cabang dengan approval dan konfirmasi dua sisi
- **Auto-journaling** — setiap pergerakan kas otomatis generate journal entry
- **Multi-branch oversight** — HQ bisa monitor posisi kas seluruh cabang real-time
- **Kas minimum alert** — early warning mencegah branch kehabisan kas

### Negatif

- **Daily maintenance** — kas harian harus di-close dan reconcile setiap hari
- **Non-rekening entry overhead** — pengeluaran operasional harus diinput manual dengan bukti
- **Inter-branch latency** — transfer kas fisik membutuhkan konfirmasi di kedua sisi (bisa delay)
- **Storage** — kas_harian dengan breakdown kategori menghasilkan record yang relatif besar per branch per hari

### Mitigasi

- Kas harian closing bisa **otomatis** saat semua teller session closed — tidak perlu manual
- Template pengeluaran rutin (gaji, listrik) bisa di-preset untuk mengurangi input manual
- Inter-branch transfer menggunakan status tracking (initiated → in_transit → received → confirmed) untuk visibility
- Breakdown kategori sebagai field terpisah (bukan JSONB) memungkinkan aggregation SQL yang efisien

## Alternatives Considered

### A. Manual Kas Entry (tanpa derivasi dari transaksi)

Operator input kas harian secara manual setiap hari.

**Ditolak** karena: rawan human error dan manipulasi. Kas harian harus **derived** dari data transaksi aktual — bukan input manual. Manual entry memutus chain of custody antara transaksi individual dan posisi kas.

### B. Real-Time Kas Balance (tanpa daily snapshot)

Menghitung posisi kas real-time dari SUM semua transaksi historis.

**Ditolak** karena: query SUM pada seluruh riwayat transaksi terlalu berat. Daily snapshot (`kas_harian`) berfungsi sebagai checkpoint — query hanya perlu menjumlahkan transaksi hari ini, bukan seluruh history. Juga, daily snapshot memungkinkan reconciliation formal.

### C. Single Account untuk Semua Branch Cash

Satu akun kas untuk seluruh tenant (tidak per branch).

**Ditolak** karena: setiap branch memiliki kas fisik terpisah yang harus dilacak independen. Kas branch Jakarta tidak bisa diklaim tersedia di branch Surabaya. Per-branch tracking diperlukan untuk operational reality dan accountability.

### D. Blocking Transaction saat Kas Minimum

Memblokir semua transaksi penarikan/pengeluaran saat kas di bawah minimum.

**Ditolak** karena: nasabah memiliki hak menarik dananya dari rekening yang sah. Memblokir penarikan karena kas rendah adalah masalah operasional koperasi, bukan masalah nasabah. Solusi yang tepat: alert ke Manager untuk arrange top-up kas, bukan memblokir transaksi nasabah.

### E. Non-Rekening Transaction di Tabel yang Sama dengan Transaksi Rekening

Menggabungkan transaksi rekening dan non-rekening dalam satu tabel.

**Ditolak** karena: transaksi rekening dan non-rekening memiliki field set yang sangat berbeda. Transaksi rekening terikat nasabah/rekening/saldo, non-rekening tidak. Penggabungan menghasilkan banyak nullable columns dan mempersulit query. Tabel terpisah (`kas_transaksi`) lebih clean dan sesuai domain masing-masing.
