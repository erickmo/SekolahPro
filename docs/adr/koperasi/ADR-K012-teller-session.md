# ADR-K012: Teller Session

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Teller session adalah mekanisme pengelolaan kas fisik per teller per shift. Setiap transaksi tunai (cash) di koperasi/BMT harus terjadi dalam konteks session teller yang aktif. Sistem harus mengakomodasi:

- **Cash accountability**: setiap teller bertanggung jawab atas kas fisik di drawer-nya
- **Session lifecycle**: buka session (terima kas awal) → proses transaksi → tutup session (hitung kas akhir)
- **Variance handling**: selisih antara kas fisik dan kas sistem harus ditangani dengan approval
- **Multi-teller per branch**: setiap teller punya session dan drawer terpisah
- **Dual-mode**: terminologi sama di kedua mode (lihat [ADR-009](../core/ADR-009-dual-mode-institution-type.md))

Teller session menjadi penghubung antara transaksi individual ([ADR-K011](./ADR-K011-transaksi.md)) dan posisi kas harian ([ADR-K014](./ADR-K014-kas-cashflow.md)).

## Decision

### 1. Session Lifecycle

Teller session mengikuti lifecycle yang ketat — tidak ada transaksi tunai tanpa session aktif:

```
┌─────────────────────────────────────────────────────┐
│                SESSION LIFECYCLE                     │
└─────────────────────────────────────────────────────┘

Supervisor serahkan kas awal
        │
        v
┌─────────────────┐
│ Status: OPEN    │  Teller terima kas_awal, hitung + konfirmasi
│ kas_awal: X     │  Session aktif, bisa proses transaksi tunai
└────────┬────────┘
         │
    ┌────┴─────────────┐
    v                  v
 Proses             ┌──────────────┐
 transaksi          │  SUSPENDED   │  Teller istirahat / handover
 tunai              │  No trx      │  Bisa di-resume
    │               └──────┬───────┘
    │                      │ resume
    │                      v
    │               ┌─────────────┐
    │               │   OPEN      │  Kembali aktif
    │               └──────┬──────┘
    │                      │
    └──────────┬───────────┘
               │ close
               v
┌──────────────────────────────┐
│ STEP 1: Teller hitung kas    │
│ fisik per denominasi (K013)  │
└──────────────┬───────────────┘
               │
               v
┌──────────────────────────────┐
│ STEP 2: Sistem hitung        │
│ expected cash                │
│ = kas_awal                   │
│   + total_setoran_tunai      │
│   - total_penarikan_tunai    │
└──────────────┬───────────────┘
               │
               v
┌──────────────────────────────┐
│ STEP 3: Bandingkan           │
│ selisih = expected - actual  │
│                              │
│ selisih = 0 → auto-close    │
│ selisih ≠ 0 → perlu approval│
└──────────────┬───────────────┘
               │
               v
┌─────────────────┐
│ Status: CLOSED  │  Kas diserahkan ke vault
└─────────────────┘
```

**Aturan:**
- Satu teller hanya bisa punya **satu session aktif** pada satu waktu
- Transaksi tunai (cash) **wajib** dikaitkan dengan teller session aktif — sistem menolak transaksi tunai tanpa session
- Transaksi non-tunai (transfer internal, system-generated) **tidak** membutuhkan teller session
- Session dibuka per shift — bukan per hari. Satu teller bisa punya multiple sessions per hari jika ada shift berbeda

### 2. Opening Session — Kas Awal

Proses pembukaan session melibatkan penyerahan kas awal dari vault/supervisor:

```
Opening Flow:
        │
        v
┌───────────────────────────────────────┐
│ 1. Supervisor ambil kas dari vault    │
│ 2. Hitung per denominasi (K013)       │
│ 3. Serahkan ke Teller                 │
│ 4. Teller konfirmasi jumlah           │
│ 5. Session OPEN, kas_awal tercatat    │
└───────────────────────────────────────┘
```

**Aturan:**
- Kas awal **wajib dihitung per denominasi** — bukan hanya total nominal (lihat [ADR-K013](./ADR-K013-money-denomination.md))
- Supervisor yang menyerahkan tercatat sebagai `opened_by_supervisor`
- Teller yang menerima tercatat sebagai `teller_id`
- Kedua pihak harus konfirmasi jumlah — konfirmasi teller di-record sebagai `opening_confirmed_at`
- Kas awal **configurable** per branch — ada nominal default yang bisa diubah Supervisor
- Denominasi kas awal dicatat di tabel `teller_session_denomination` dengan `phase = 'opening'`

### 3. During Session — Running Total

Selama session aktif, sistem memaintain running total kas teller:

```
Running Total Calculation:
┌─────────────────────────────────────┐
│ kas_awal            = 5.000.000     │
│ + total_cash_in     = 12.500.000    │  (setoran tunai, angsuran tunai)
│ - total_cash_out    = 8.200.000     │  (penarikan tunai, pencairan tunai)
│ ─────────────────────────────────   │
│ expected_cash       = 9.300.000     │  (kas yang seharusnya ada di drawer)
└─────────────────────────────────────┘
```

**Fields yang di-update real-time:**

| Field | Deskripsi | Update Trigger |
|-------|-----------|----------------|
| `total_cash_in` | Total kas masuk tunai | Setiap setoran/angsuran tunai |
| `total_cash_out` | Total kas keluar tunai | Setiap penarikan/pencairan tunai |
| `transaction_count` | Jumlah transaksi tunai | Setiap transaksi tunai |
| `expected_cash` | Kas yang seharusnya ada | Dihitung: kas_awal + in - out |

**Aturan:**
- Running total di-update **atomik** bersamaan dengan insert transaksi (dalam satu database transaction — lihat [K011 §3](./ADR-K011-transaksi.md))
- Teller bisa melihat running total kapan saja selama session aktif
- Running total bersifat **derived** — bisa diverifikasi dari SUM transaksi di session tersebut
- Jika `expected_cash` mendekati 0, sistem memberikan **warning** ke teller (drawer hampir kosong)

### 4. Closing Session — Cash Count & Variance

Proses penutupan session adalah langkah paling kritis — memastikan kas fisik sesuai ekspektasi:

```
Closing Flow:
        │
        v
┌───────────────────────────────────────┐
│ 1. Teller inisiasi close              │
│ 2. Sistem block transaksi baru        │
│    untuk session ini                  │
│ 3. Teller hitung kas fisik per        │
│    denominasi (K013)                  │
│ 4. Input denominasi ke sistem         │
│ 5. Sistem hitung:                     │
│    expected = kas_awal + in - out     │
│    actual   = SUM(denominasi)         │
│    selisih  = expected - actual       │
│ 6. Proses berdasarkan selisih         │
└───────────────────────────────────────┘
```

**Expected cash formula:**
```
expected_cash = kas_awal
              + total_cash_in     (setoran tunai + angsuran tunai)
              - total_cash_out    (penarikan tunai + pencairan tunai)
              + kas_masuk_netto   (kas operasional masuk - keluar via teller)
```

**Aturan penutupan:**
- Saat teller inisiasi close, sistem **memblokir transaksi baru** untuk session tersebut
- Teller **wajib** input denominasi kas fisik (lihat [K013](./ADR-K013-money-denomination.md))
- Denominasi closing dicatat di `teller_session_denomination` dengan `phase = 'closing'`
- Sistem menampilkan side-by-side: expected vs actual vs selisih
- Kas fisik diserahkan ke vault setelah session closed

### 5. Variance (Selisih) Handling

Selisih antara kas yang diharapkan dan kas fisik ditangani berdasarkan threshold:

```
Selisih = Expected Cash - Actual Cash

┌─────────────────────────────────────────────────┐
│ selisih = 0                                     │
│ → Auto-close, tidak perlu approval              │
├─────────────────────────────────────────────────┤
│ |selisih| <= threshold_minor (configurable)     │
│ → Supervisor approve + wajib isi explanation    │
│ → Selisih dicatat sebagai kas_selisih           │
│ → Default threshold: Rp 10.000                  │
├─────────────────────────────────────────────────┤
│ |selisih| > threshold_minor                     │
│   AND |selisih| <= threshold_major              │
│ → Manager approve + investigation required      │
│ → Default threshold_major: Rp 100.000           │
├─────────────────────────────────────────────────┤
│ |selisih| > threshold_major                     │
│ → Manager approve + full investigation          │
│ → Flag teller for review                        │
│ → Notifikasi ke Admin                           │
└─────────────────────────────────────────────────┘
```

**Variance config (per tenant):**
```
variance_config:
  threshold_minor:  10000      # Selisih kecil, Supervisor cukup
  threshold_major:  100000     # Selisih besar, Manager + investigation
  auto_close_zero:  true       # Auto-close jika selisih = 0
  require_explanation: true    # Wajib isi penjelasan jika selisih ≠ 0
```

**Aturan:**
- Selisih **positif** (expected > actual): kas kurang → potensi kehilangan/pencurian
- Selisih **negatif** (expected < actual): kas lebih → potensi transaksi tidak tercatat
- Kedua arah selisih sama-sama membutuhkan perhatian
- Explanation **wajib** jika selisih ≠ 0 — tidak bisa close tanpa penjelasan
- Selisih dicatat sebagai `kas_selisih` di teller session — menjadi bagian dari reconciliation harian
- Riwayat selisih per teller dimonitor — pola selisih berulang bisa trigger review

### 6. Multiple Tellers per Branch

Setiap teller memiliki session dan cash drawer independen:

```
Branch: Cabang Jakarta Pusat
├── Teller 1 (Ahmad)
│   └── Session #101 (OPEN)
│       ├── kas_awal:      5.000.000
│       ├── cash_in:       8.000.000
│       ├── cash_out:      3.000.000
│       └── expected_cash: 10.000.000
│
├── Teller 2 (Budi)
│   └── Session #102 (OPEN)
│       ├── kas_awal:      5.000.000
│       ├── cash_in:       6.000.000
│       ├── cash_out:      4.000.000
│       └── expected_cash: 7.000.000
│
└── Teller 3 (Citra)
    └── Session #103 (SUSPENDED)
        └── (istirahat siang)
```

**Aturan:**
- Setiap teller punya **drawer fisik terpisah** — tidak ada sharing kas
- Transaksi tunai terikat pada **session teller yang memproses**
- Supervisor bisa melihat semua session aktif di branch-nya
- Dashboard branch menampilkan aggregate seluruh teller sessions
- Maximum teller per branch **configurable** (default: tidak dibatasi, tergantung jumlah user dengan role Teller)

### 7. Session Suspend & Resume

Teller bisa suspend session saat istirahat atau handover:

```
OPEN → SUSPENDED
├── Trigger: Teller klik suspend / auto-lock timeout
├── Effect: Tidak bisa proses transaksi baru
├── Kas tetap di drawer teller (tidak diserahkan)
└── Duration: max configurable (default: 60 menit)

SUSPENDED → OPEN
├── Trigger: Teller resume (login kembali)
├── Validasi: Teller yang sama yang buka session
└── Effect: Bisa proses transaksi lagi
```

**Aturan:**
- Session yang di-suspend **tidak bisa menerima transaksi** — teller harus resume dulu
- Hanya **teller yang sama** yang bisa resume session-nya — tidak bisa di-resume orang lain
- Auto-lock: jika teller idle selama configurable timeout (default: 15 menit), session otomatis suspend
- Maximum suspend duration: configurable (default: 60 menit) — jika melebihi, Supervisor harus intervene
- Kas **tetap di drawer** selama suspend — tidak perlu hitung ulang saat resume
- Tidak ada denominasi counting saat suspend/resume — hanya saat open dan close

### 8. Force Close

Manager bisa menutup paksa session yang ditinggalkan:

```
Force Close Flow:
        │
        v
┌───────────────────────────────────────┐
│ Kondisi:                              │
│ - Session masih OPEN atau SUSPENDED   │
│ - Teller tidak available              │
│   (sakit, emergency, lupa close)      │
│                                       │
│ Manager:                              │
│ 1. Initiate force close               │
│ 2. Wajib isi alasan force close       │
│ 3. Manager (atau Supervisor)          │
│    hitung kas fisik di drawer teller  │
│ 4. Input denominasi                   │
│ 5. Selisih ditangani seperti biasa    │
│ 6. Session closed dengan flag         │
│    is_force_closed = true             │
└───────────────────────────────────────┘
```

**Aturan:**
- Hanya **Manager+** yang bisa force close
- Flag `is_force_closed = true` + `force_close_reason` wajib diisi
- Kas fisik tetap dihitung per denominasi — force close bukan berarti skip counting
- Force close tercatat di audit log dan menjadi alert untuk review
- Teller yang session-nya di-force close harus memberikan penjelasan di shift berikutnya

### 9. Daily Consolidation

Semua teller sessions dalam satu hari dikonsolidasi menjadi branch daily report:

```
Daily Consolidation:
┌──────────────────────────────────────────┐
│ Branch: Cabang Jakarta Pusat             │
│ Tanggal: 15 April 2026                  │
│                                          │
│ Teller Sessions:                         │
│ ┌────────┬──────────┬──────────┬───────┐ │
│ │ Teller │ Kas Awal │ Expected │Selisih│ │
│ ├────────┼──────────┼──────────┼───────┤ │
│ │ Ahmad  │ 5.000.000│10.000.000│     0 │ │
│ │ Budi   │ 5.000.000│ 7.000.000│  -500 │ │
│ │ Citra  │ 3.000.000│ 5.500.000│     0 │ │
│ ├────────┼──────────┼──────────┼───────┤ │
│ │ TOTAL  │13.000.000│22.500.000│  -500 │ │
│ └────────┴──────────┴──────────┴───────┘ │
│                                          │
│ → Feed ke Kas Harian (K014)              │
└──────────────────────────────────────────┘
```

**Aturan:**
- Consolidation berjalan otomatis setelah **semua** teller sessions di branch closed untuk hari tersebut
- Jika ada session yang belum closed di akhir hari, sistem mengirim **alert** ke Supervisor/Manager
- Total dari consolidation menjadi input untuk kas harian ([K014](./ADR-K014-kas-cashflow.md))
- Total selisih seluruh teller menjadi selisih harian branch
- Report bisa di-generate ulang (recalculate) — bukan snapshot statis

### 10. Data Model — Teller Session

```
teller_session
├── id                        UUID v7 (PK)
├── tenant_id                 UUID (FK → tenant) NOT NULL
├── branch_id                 UUID (FK → branch) NOT NULL
├── teller_user_id            UUID (FK → user) NOT NULL
│
├── ── Session Info ──
├── session_date              DATE NOT NULL
├── session_number            INTEGER (sequence per branch per day)
│
├── ── Kas Awal ──
├── opening_amount            NUMERIC(15,2) NOT NULL
├── opened_by_supervisor      UUID (FK → user) NOT NULL
├── opening_confirmed_at      TIMESTAMPTZ
│
├── ── Running Total ──
├── total_cash_in             NUMERIC(15,2) NOT NULL DEFAULT 0
├── total_cash_out            NUMERIC(15,2) NOT NULL DEFAULT 0
├── transaction_count         INTEGER NOT NULL DEFAULT 0
├── expected_cash             NUMERIC(15,2) NOT NULL
│                             (= opening_amount + cash_in - cash_out)
│
├── ── Closing ──
├── actual_cash               NUMERIC(15,2) (nullable, diisi saat close)
├── variance                  NUMERIC(15,2) (nullable, = expected - actual)
├── variance_explanation      TEXT (nullable, wajib jika variance ≠ 0)
├── variance_approved_by      UUID (nullable, FK → user)
├── variance_approved_at      TIMESTAMPTZ (nullable)
│
├── ── Status ──
├── status                    ENUM (open, suspended, closed) NOT NULL
├── opened_at                 TIMESTAMPTZ NOT NULL
├── suspended_at              TIMESTAMPTZ (nullable)
├── resumed_at                TIMESTAMPTZ (nullable)
├── closed_at                 TIMESTAMPTZ (nullable)
│
├── ── Force Close ──
├── is_force_closed           BOOLEAN NOT NULL DEFAULT false
├── force_closed_by           UUID (nullable, FK → user)
├── force_close_reason        TEXT (nullable)
│
├── ── Vernon Fields ──
├── _rels                     JSONB NOT NULL DEFAULT '{}'
├── _data                     JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at            TIMESTAMPTZ
    ├── created_by            UUID
    ├── updated_at            TIMESTAMPTZ
    └── updated_by            UUID
```

**Database constraints:**
- `UNIQUE (tenant_id, teller_user_id, session_date, session_number)` — satu teller, satu nomor session per hari
- Index: `(tenant_id, branch_id, session_date)` — query per branch per hari
- Index: `(tenant_id, teller_user_id, status)` — cari session aktif teller
- `CHECK (expected_cash = opening_amount + total_cash_in - total_cash_out)`

### 11. Data Model — Teller Session Denomination

Denominasi kas dicatat per session per phase (opening dan closing):

```
teller_session_denomination
├── id                        UUID v7 (PK)
├── tenant_id                 UUID (FK → tenant) NOT NULL
├── teller_session_id         UUID (FK → teller_session) NOT NULL
│
├── ── Phase & Denomination ──
├── phase                     ENUM (opening, closing) NOT NULL
├── denomination_value        INTEGER NOT NULL (nilai pecahan: 100000, 50000, dll)
├── quantity                  INTEGER NOT NULL DEFAULT 0
├── subtotal                  NUMERIC(15,2) NOT NULL
│                             (= denomination_value * quantity)
│
└── ── Audit ──
    ├── created_at            TIMESTAMPTZ
    ├── created_by            UUID
```

**Aturan:**
- Satu record per denominasi per phase per session
- `SUM(subtotal) WHERE phase = 'opening'` = `opening_amount` (harus match)
- `SUM(subtotal) WHERE phase = 'closing'` = `actual_cash` (harus match)
- Denominasi mengacu ke master denominasi ([K013](./ADR-K013-money-denomination.md))

### 12. Vernon _rels dan _data Structure

**_rels:**
```json
{
  "tenant_id":           "018f...",
  "branch_id":           "018f...",
  "teller_user_id":      "018f...",
  "opened_by_supervisor":"018f..."
}
```

**_data:**
```json
{
  "teller": {
    "id":   "018f...",
    "name": "Ahmad Teller",
    "code": "T-001"
  },
  "branch": {
    "id":   "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  },
  "supervisor": {
    "id":   "018f...",
    "name": "Pak Budi Supervisor"
  }
}
```

**SyncEngine triggers:**
- `UserUpdatedEvent` (teller) → update `_data.teller` di semua session teller tersebut
- `UserUpdatedEvent` (supervisor) → update `_data.supervisor` di semua session yang dibuka supervisor tersebut
- `BranchUpdatedEvent` → update `_data.branch` di semua session cabang tersebut

### 13. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| Open session (terima kas awal) | v* | - | - | - |
| Serahkan kas awal ke teller | - | v | v | v |
| View own session | v | v | v | v |
| View all branch sessions | - | v | v | v |
| Suspend/Resume own session | v | v | - | - |
| Close own session (hitung kas) | v | v | - | - |
| Approve variance (minor) | - | v | v | v |
| Approve variance (major) | - | - | v | v |
| Force close session | - | - | v | v |
| View daily consolidation | - | v | v | v |
| View cross-branch consolidation | - | - | v | v |

`*` Teller membuka session setelah menerima kas awal dari Supervisor

**Catatan:**
- Teller hanya mengelola **session sendiri** — tidak bisa melihat session teller lain
- Supervisor bisa melihat semua session di branch-nya dan approve variance kecil
- Manager bisa force close dan approve variance besar
- Opening session membutuhkan **dua orang**: Supervisor (serahkan kas) + Teller (terima dan konfirmasi)

### 14. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Teller Session | Sesi Teller | Sesi Teller |
| Kas Awal | Kas Awal | Kas Awal |
| Kas Akhir | Kas Akhir | Kas Akhir |
| Selisih | Selisih Kas | Selisih Kas |
| Vault | Brankas | Brankas |

Terminologi teller session **identik** di kedua mode — tidak ada perbedaan syariah pada operasional kas fisik.

## Consequences

### Positif

- **Cash accountability** — setiap rupiah di drawer teller tercatat dan bisa ditelusuri
- **Fraud detection** — variance tracking per teller memudahkan deteksi anomali
- **Dual control** — opening membutuhkan Supervisor + Teller, closing diverifikasi sistem
- **Operational flexibility** — suspend/resume mengakomodasi istirahat tanpa close session
- **Consolidated view** — daily consolidation memberikan gambaran lengkap kas branch
- **Denomination tracking** — denominasi fisik tercatat untuk vault management
- **Force close safety net** — Manager bisa handle session yang ditinggalkan

### Negatif

- **Operational overhead** — setiap shift harus open/close session dengan hitung denominasi
- **Blocking on close** — transaksi baru diblokir saat teller menghitung kas akhir
- **Variance approval latency** — teller tidak bisa pulang sampai variance di-approve (jika ≠ 0)
- **Running total sync** — update atomik menambah overhead per transaksi tunai

### Mitigasi

- Denominasi counting bisa dipercepat dengan **template** jumlah default (misal: kas awal selalu sama)
- Blocking saat close hanya berlaku untuk **session teller tersebut** — teller lain tetap bisa proses
- Supervisor bisa approve variance langsung (untuk kasus minor) — tidak perlu menunggu Manager
- Running total update sangat ringan (single field increment/decrement) dalam transaksi yang sama

## Alternatives Considered

### A. Branch-Level Session (bukan per Teller)

Satu session per branch per hari, bukan per teller.

**Ditolak** karena: tidak bisa melacak kas per teller. Jika ada selisih, tidak tahu teller mana yang bertanggung jawab. Cash accountability individual adalah standar operasional koperasi/bank.

### B. Tanpa Denominasi pada Opening/Closing

Hanya input total nominal tanpa detail per denominasi.

**Ditolak** karena: denominasi detail diperlukan untuk vault management dan deteksi penipuan. Total nominal saja tidak cukup — teller bisa mengklaim jumlah yang sama dengan komposisi berbeda. Detail denominasi juga dibutuhkan untuk laporan kas ke pengawas.

### C. Auto-Close pada Akhir Hari

Session otomatis ditutup pada jam tertentu tanpa penghitungan kas.

**Ditolak** karena: melanggar prinsip cash accountability. Setiap penutupan session harus melibatkan penghitungan fisik kas. Auto-close tanpa counting sama saja dengan mengabaikan potensi selisih/kehilangan.

### D. No Suspend — Harus Close dan Open Baru

Jika teller istirahat, harus close session lalu open session baru setelah kembali.

**Ditolak** karena: setiap close + open membutuhkan hitung denominasi dan serah terima kas ke vault. Terlalu berat untuk istirahat 15-30 menit. Suspend lebih efisien — kas tetap di drawer, session di-lock dari transaksi.
