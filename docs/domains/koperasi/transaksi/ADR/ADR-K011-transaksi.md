# ADR-K011: Transaksi Rekening & Non-Rekening

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Transaksi adalah **core engine** dari seluruh pergerakan keuangan di koperasi/BMT. Setiap aliran dana — masuk maupun keluar — harus tercatat sebagai transaksi yang immutable dan auditable. Sistem harus mengakomodasi:

- **Transaksi rekening**: mempengaruhi saldo rekening nasabah (setoran, penarikan, angsuran, pencairan, dll)
- **Transaksi non-rekening**: pergerakan kas operasional yang tidak terkait rekening nasabah (gaji, ATK, listrik)
- **Atomic balance update**: saldo rekening harus diperbarui bersamaan dengan pencatatan transaksi dalam satu database transaction
- **Immutability**: transaksi yang sudah committed tidak boleh diedit atau dihapus — koreksi dilakukan via transaksi reversal
- **Dual-mode terminologi**: Bunga vs Bagi Hasil, Pinjaman vs Pembiayaan — lihat [ADR-009](../core/ADR-009-dual-mode-institution-type.md)
- **Multi-tenant isolation**: transaksi terisolasi per tenant, transfer hanya dalam satu tenant

## Decision

### 1. Transaction Types Taxonomy

Seluruh transaksi dikelompokkan dalam dua kategori besar:

```
REKENING TRANSACTIONS (mempengaruhi saldo rekening):
├── SETORAN (deposit/credit ke rekening)
│   ├── Setoran tunai (cash)
│   ├── Setoran transfer (dari rekening lain)
│   ├── Setoran bagi hasil/bunga (system-generated)
│   └── Setoran pembukaan (initial deposit)
├── PENARIKAN (withdrawal/debit dari rekening)
│   ├── Penarikan tunai (cash)
│   ├── Penarikan transfer (ke rekening lain)
│   └── Penarikan penutupan (closing withdrawal)
├── ANGSURAN (pembayaran cicilan pinjaman)
├── PENCAIRAN (disbursement pinjaman ke nasabah)
├── BIAYA_ADMIN (potongan biaya administrasi)
├── DENDA (penalti keterlambatan/pelanggaran)
└── KOREKSI (correction — Manager+ only)

NON-REKENING TRANSACTIONS (tidak terkait rekening nasabah):
├── KAS_KELUAR (pembayaran operasional — gaji, ATK, listrik)
├── KAS_MASUK (penerimaan non-anggota — sewa, denda umum)
└── TRANSFER_KAS (perpindahan kas antar teller/vault/cabang)
```

**Aturan tipe transaksi:**

| Tipe | Direction | Affects Balance | Requires Rekening | Cash Movement |
|------|-----------|-----------------|-------------------|---------------|
| SETORAN | CREDIT | Ya, balance + | Ya | Masuk (tunai) / Internal (transfer) |
| PENARIKAN | DEBIT | Ya, balance - | Ya | Keluar (tunai) / Internal (transfer) |
| ANGSURAN | CREDIT (pinjaman) | Ya, outstanding - | Ya | Masuk |
| PENCAIRAN | DEBIT (pinjaman) | Ya, outstanding + | Ya | Keluar |
| BIAYA_ADMIN | DEBIT | Ya, balance - | Ya | Internal (auto-deduct) |
| DENDA | DEBIT | Ya, balance - | Ya | Internal (auto-deduct) |
| KOREKSI | CREDIT/DEBIT | Ya | Ya | Tidak ada |
| KAS_KELUAR | - | Tidak | Tidak | Keluar |
| KAS_MASUK | - | Tidak | Tidak | Masuk |
| TRANSFER_KAS | - | Tidak | Tidak | Internal |

### 2. Transaction Immutability

Transaksi yang sudah committed bersifat **immutable** — tidak bisa diedit atau dihapus dalam kondisi apapun.

```
┌─────────────────────────────────────┐
│         IMMUTABILITY RULES          │
├─────────────────────────────────────┤
│ 1. INSERT only — no UPDATE/DELETE   │
│ 2. Koreksi = transaksi baru         │
│ 3. Semua field immutable post-commit│
│ 4. Audit trail lengkap              │
│ 5. Database constraint: no UPDATE   │
│    trigger on transaksi table       │
└─────────────────────────────────────┘
```

**Enforcement:**
- Database trigger menolak UPDATE dan DELETE pada tabel transaksi
- Satu-satunya field yang boleh di-update: `is_reversed` (di-set `true` saat reversal dibuat)
- Application layer tidak menyediakan endpoint edit/delete transaksi
- Audit log mencatat semua attempt yang ditolak

### 3. Atomic Balance Update

Insert transaksi dan update saldo rekening **harus** dalam satu database transaction:

```
BEGIN TRANSACTION
    │
    ├── 1. Validate (lihat section 4)
    │
    ├── 2. SELECT rekening FOR UPDATE  ← Row-level lock
    │
    ├── 3. INSERT INTO transaksi (...)
    │
    ├── 4. UPDATE rekening
    │      SET balance = balance + amount,
    │          available_balance = balance + amount - hold_amount,
    │          last_transaction_at = NOW()
    │
    ├── 5. UPDATE teller_session totals (jika cash transaction)
    │
    └── 6. COMMIT
         │
         FAIL at any step → ROLLBACK semua
```

**Aturan:**
- `SELECT ... FOR UPDATE` mencegah race condition pada saldo rekening
- Amount positif untuk CREDIT, negatif untuk DEBIT di level database
- `available_balance` selalu dihitung ulang: `balance - hold_amount`
- Jika transaksi melibatkan teller session (cash), total teller juga di-update dalam transaksi yang sama

### 4. Transaction Validation Engine

Setiap transaksi melewati serangkaian validasi sebelum dieksekusi:

```
Transaction Validation Pipeline
        │
        v
┌───────────────────────────────────────┐
│ V1. Rekening Status Check             │
│     - ACTIVE: semua transaksi OK      │
│     - FROZEN: hanya CREDIT diizinkan  │
│     - CLOSED: semua ditolak           │
└────────────────┬──────────────────────┘
                 │ pass
                 v
┌───────────────────────────────────────┐
│ V2. Available Balance Check (DEBIT)   │
│     - available_balance >= amount     │
│     - Simpanan: balance - amount      │
│       >= minimum_balance              │
└────────────────┬──────────────────────┘
                 │ pass
                 v
┌───────────────────────────────────────┐
│ V3. KYC Limit Check (K001 §11)       │
│     - SUM transaksi bulan ini         │
│       + amount <= kyc_monthly_limit   │
│     - basic: configurable (def 10jt)  │
│     - full: unlimited                 │
└────────────────┬──────────────────────┘
                 │ pass
                 v
┌───────────────────────────────────────┐
│ V4. Product-Level Limit Check         │
│     - Daily transaction limit         │
│     - Monthly transaction limit       │
│     - Single transaction max amount   │
│     - Configurable per produk         │
└────────────────┬──────────────────────┘
                 │ pass
                 v
┌───────────────────────────────────────┐
│ V5. Minimum Balance Maintenance       │
│     - balance - amount >= min_balance │
│     - Hanya untuk DEBIT pada simpanan │
└────────────────┬──────────────────────┘
                 │ pass
                 v
┌───────────────────────────────────────┐
│ V6. Specimen Verification (K001 §10)  │
│     - Penarikan > threshold           │
│       → flag: specimen_verified       │
│     - Threshold configurable/tenant   │
│     - > threshold_manager → Manager+  │
└────────────────┬──────────────────────┘
                 │ pass
                 v
┌───────────────────────────────────────┐
│ V7. Teller Session Check              │
│     - Cash transaction → teller harus │
│       punya session aktif             │
│     - Session not suspended           │
└────────────────┬──────────────────────┘
                 │ all pass
                 v
            EXECUTE TRANSACTION
```

**Aturan:**
- Validasi berjalan secara sequential — fail di step manapun langsung menolak transaksi
- Error response mengembalikan **semua** validasi yang gagal (bukan hanya yang pertama) untuk UX yang baik
- V3 (KYC limit) menghitung SUM semua transaksi nasabah di bulan berjalan, bukan per rekening
- V6 (Specimen) hanya berlaku untuk penarikan tunai — transfer internal tidak perlu specimen

### 5. Transfer Antar Rekening

Transfer antara dua rekening dalam satu tenant dieksekusi sebagai **dua transaksi atomik**:

```
Transfer: Rekening A → Rekening B (Rp 500.000)
        │
        v
BEGIN TRANSACTION
    │
    ├── Validate Rekening A (DEBIT rules)
    ├── Validate Rekening B (CREDIT rules — bisa FROZEN)
    │
    ├── INSERT transaksi DEBIT  (rek A, -500.000, ref: TRF-xxx)
    ├── UPDATE rekening A balance
    │
    ├── INSERT transaksi CREDIT (rek B, +500.000, ref: TRF-xxx)
    ├── UPDATE rekening B balance
    │
    ├── Kedua transaksi share transfer_ref
    │
    └── COMMIT (atau ROLLBACK keduanya)
```

**Deadlock Prevention — Consistent Lock Ordering:**

Transfer antar rekening melibatkan locking dua row secara bersamaan. Untuk mencegah deadlock (A→B dan B→A bersamaan):

1. **WAJIB lock rekening dengan UUID lebih kecil terlebih dahulu**
2. Kemudian lock rekening dengan UUID lebih besar
3. Setelah kedua row terkunci, proses debit dan kredit

Pseudocode:
```
  if source_rekening_id < dest_rekening_id:
      LOCK source FIRST, then dest
  else:
      LOCK dest FIRST, then source
```

Aturan ini berlaku untuk SEMUA operasi yang melibatkan locking lebih dari satu rekening dalam satu database transaction.

**Aturan:**
- Kedua rekening **harus dalam tenant yang sama** — cross-tenant transfer tidak diizinkan
- Kedua rekening bisa di branch berbeda (intra-tenant)
- Rekening source harus ACTIVE, rekening destination bisa ACTIVE atau FROZEN
- Kedua transaksi memiliki `transfer_ref` yang sama untuk tracing
- `transfer_pair_id` menghubungkan kedua transaksi (debit → credit)
- Tidak ada pergerakan kas fisik — murni pembukuan internal
- KYC limit dihitung untuk **kedua** nasabah

### 6. Transaction Reference Number

Setiap transaksi mendapat nomor referensi unik yang auto-generated:

```
transaction_ref_config:
  prefix: "TRX"                      # Configurable per tenant
  separator: "-"                      # Configurable separator
  include_date: true                  # Sertakan tanggal (YYYYMMDD)
  include_branch: true                # Sertakan kode cabang
  sequence_digits: 8                  # Jumlah digit sequence
  sequence_scope: "daily"             # Reset harian / global
```

**Contoh hasil:**
```
full format:      TRX-20260415-JKT-00000001
minimal format:   TRX-00000001
transfer pair:    TRF-20260415-JKT-00000001 (khusus transfer)
```

**Aturan:**
- Referensi **immutable** setelah assigned
- Uniqueness di-enforce di database level (unique constraint per tenant)
- Sequence bisa di-reset harian (scope `daily`) atau tidak pernah reset (scope `global`)
- Transfer menggunakan prefix berbeda (`TRF`) untuk identifikasi visual
- Nomor referensi ditampilkan di receipt/bukti transaksi

### 7. Bukti Transaksi / Receipt

Setiap transaksi menghasilkan **bukti transaksi** yang bisa dicetak:

```
┌──────────────────────────────────────┐
│        KOPERASI SEKOLAH XYZ          │
│        Cabang Jakarta Pusat          │
│                                      │
│  BUKTI SETORAN TUNAI                 │
│                                      │
│  No. Ref    : TRX-20260415-JKT-001  │
│  Tanggal    : 15 April 2026, 10:30  │
│  Teller     : Ahmad (T-001)         │
│                                      │
│  Nasabah    : Budi Santoso          │
│  No. Anggota: KOP-2026-JKT-000042  │
│  Rekening   : TB-2026-JKT-00000001 │
│                                      │
│  Nominal    : Rp   500.000,00       │
│  Saldo Akhir: Rp 2.500.000,00      │
│                                      │
│  ──────────────────────────────────  │
│  Teller: ________  Nasabah: _______ │
└──────────────────────────────────────┘
```

**Aturan:**
- Receipt di-generate otomatis untuk **setiap** transaksi
- Format receipt **configurable per tenant** (header, footer, logo)
- Data receipt diambil dari transaksi + Vernon `_data` (no additional query)
- Receipt bisa dicetak ulang (reprint) — tidak generate transaksi baru
- Reprint ditandai dengan watermark "SALINAN" dan dicatat di audit log
- Receipt tersimpan sebagai template + data reference, bukan file statis

### 8. Reversal / Correction Flow

Koreksi transaksi dilakukan melalui **transaksi reversal** — bukan edit:

```
Transaksi Original (TRX-001, Setoran Rp 500.000)
        │
        │  Ditemukan kesalahan
        v
┌───────────────────────────────────────┐
│ STEP 1: Request Reversal              │
│ - Teller/Supervisor mengajukan        │
│ - Wajib isi alasan (reversal_reason)  │
│ - Lampirkan bukti jika ada            │
└────────────────┬──────────────────────┘
                 │
                 v
┌───────────────────────────────────────┐
│ STEP 2: Manager+ Approval             │
│ - Review original + alasan reversal   │
│ - Approve atau Reject                 │
└────────────────┬──────────────────────┘
                 │ approved
                 v
┌───────────────────────────────────────┐
│ STEP 3: Execute Reversal              │
│ - Original: is_reversed = true        │
│ - New TRX: tipe KOREKSI              │
│   amount = -500.000 (kebalikan)       │
│   reversal_of = TRX-001              │
│   reversal_reason = "..."            │
│ - Balance di-update atomik            │
└───────────────────────────────────────┘
```

**Aturan:**
- Reversal **selalu** membutuhkan approval Manager+ — tidak ada auto-reversal
- Original transaction ditandai `is_reversed = true` (satu-satunya field yang boleh di-update)
- Transaksi reversal baru dibuat dengan `reversal_of = original_id`
- Amount reversal = kebalikan dari original (jika original +500.000, reversal -500.000)
- Reversal mengikuti validasi yang sama (check balance, etc.) — jika saldo sudah habis, reversal bisa ditolak
- Satu transaksi hanya bisa di-reverse **sekali** — `is_reversed` mencegah double reversal
- Partial reversal tidak diizinkan — harus full amount. Jika perlu partial, buat transaksi koreksi manual
- Semua reversal tercatat di audit log dengan detail lengkap

### 9. Batch Transactions

Operasi massal diproses sebagai **batch** yang berisi transaksi individual:

```
Batch Operation (contoh: Simpanan Wajib bulanan)
        │
        v
┌───────────────────────────────────────┐
│ STEP 1: Generate Batch                │
│ - System/Supervisor buat batch        │
│ - Tipe: SIMPANAN_WAJIB_COLLECTION     │
│ - Target: semua nasabah aktif         │
│ - Generate list transaksi individual  │
└────────────────┬──────────────────────┘
                 │
                 v
┌───────────────────────────────────────┐
│ STEP 2: Review & Approve              │
│ - Manager review batch summary        │
│ - Total transaksi, total nominal      │
│ - Bisa exclude nasabah tertentu       │
│ - Approve untuk eksekusi              │
└────────────────┬──────────────────────┘
                 │ approved
                 v
┌───────────────────────────────────────┐
│ STEP 3: Execute                       │
│ - Proses satu per satu               │
│ - Setiap item = transaksi individual │
│ - Gagal 1 tidak menggagalkan semua   │
│ - Status per item: success/failed     │
│ - Summary report setelah selesai      │
└───────────────────────────────────────┘
```

**Jenis batch operations:**
- **Simpanan wajib collection**: debit rekening tabungan/potong gaji → credit simpanan wajib
- **Bagi hasil/bunga distribution**: credit ke semua rekening tabungan/deposito eligible
- **Admin fee deduction**: debit biaya admin dari rekening dormant atau bulanan
- **Denda batch**: apply denda ke semua pinjaman yang terlambat

**Aturan:**
- Setiap item batch diproses sebagai **transaksi individual** dengan referensi ke batch_id
- Item yang gagal validasi di-skip (tidak rollback seluruh batch)
- Batch summary report mencatat: total items, success count, failed count, total amount processed
- Batch membutuhkan approval **Manager+** sebelum eksekusi
- Batch bisa di-schedule (misal: simpanan wajib setiap tanggal 1)

### 10. Transaction Search & Filtering

Pencarian transaksi mendukung filter komprehensif:

```
Search Parameters:
├── date_from / date_to       DATE range
├── transaction_type          ENUM (single or multiple)
├── amount_min / amount_max   NUMERIC range
├── rekening_id               UUID (specific rekening)
├── nasabah_id                UUID (all rekening milik nasabah)
├── branch_id                 UUID (semua transaksi di cabang)
├── teller_session_id         UUID (transaksi dalam satu sesi)
├── batch_id                  UUID (transaksi dalam satu batch)
├── reference_number          VARCHAR (exact match)
├── is_reversed               BOOLEAN
├── is_cash                   BOOLEAN
└── keyword                   VARCHAR (search di description)

Sort Options:
├── created_at (default: DESC)
├── amount
└── reference_number

Output:
├── Paginated list (default: 20 per page)
├── Summary: total records, total credit, total debit, net
└── Export: CSV, PDF (configurable)
```

**Aturan:**
- Default filter: 30 hari terakhir (mencegah query terlalu besar)
- Teller hanya bisa search transaksi di **branch sendiri**
- Supervisor bisa search transaksi di branch sendiri
- Manager+ bisa search **cross-branch** dalam tenant
- Export membutuhkan minimal role **Supervisor**

### 11. Data Model — Transaksi Rekening

**Idempotency Key — Pencegahan Transaksi Duplikat:**

Setiap request pembuatan transaksi WAJIB menyertakan `idempotency_key`:

- `idempotency_key` UUID — di-generate oleh client (frontend/API caller)
- UNIQUE constraint: `(tenant_id, idempotency_key)` di tabel transaksi
- Jika key sudah ada: return transaksi existing (bukan error), HTTP 200
- Deduplication window: key valid selama 24 jam, setelah itu bisa di-reuse

Use case:
- Teller double-click tombol "Proses" — request kedua return transaksi pertama
- Network timeout + retry — retry idempotent, tidak membuat duplikat
- Batch processing retry — item yang sudah berhasil di-skip

```
transaksi
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant) NOT NULL
├── branch_id             UUID (FK → branch) NOT NULL
├── rekening_id           UUID (FK → rekening) NOT NULL
├── nasabah_id            UUID (FK → nasabah) NOT NULL
│
├── ── Identitas ──
├── idempotency_key       UUID NOT NULL
├── reference_number      VARCHAR UNIQUE per tenant
├── transaction_type      ENUM (setoran, penarikan, angsuran, pencairan,
│                               biaya_admin, denda, koreksi)
├── transaction_subtype   VARCHAR (tunai, transfer, bagi_hasil, bunga,
│                                  pembukaan, penutupan, system)
├── description           TEXT (keterangan transaksi)
│
├── ── Nominal ──
├── amount                NUMERIC(15,2) NOT NULL
│                         (positif = credit, negatif = debit)
├── balance_before        NUMERIC(15,2) NOT NULL (snapshot saldo sebelum)
├── balance_after         NUMERIC(15,2) NOT NULL (snapshot saldo sesudah)
│
├── ── Cash Info ──
├── is_cash               BOOLEAN NOT NULL DEFAULT false
├── teller_session_id     UUID (nullable, FK → teller_session)
│
├── ── Transfer Info ──
├── transfer_ref          VARCHAR (nullable, shared ref untuk pasangan transfer)
├── transfer_pair_id      UUID (nullable, FK → transaksi, pasangan transfer)
│
├── ── Reversal Info ──
├── is_reversed           BOOLEAN NOT NULL DEFAULT false
├── reversal_of           UUID (nullable, FK → transaksi, original yang di-reverse)
├── reversal_reason       TEXT (nullable)
│
├── ── Batch Info ──
├── batch_id              UUID (nullable, FK → transaction_batch)
│
├── ── Verification ──
├── specimen_verified     BOOLEAN NOT NULL DEFAULT false
├── specimen_verified_by  UUID (nullable, FK → user)
├── approval_status       ENUM (none, pending, approved, rejected) DEFAULT 'none'
├── approved_by           UUID (nullable, FK → user)
├── approved_at           TIMESTAMPTZ (nullable)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ NOT NULL
    ├── created_by        UUID NOT NULL
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Database constraints:**
- `CHECK (balance_after = balance_before + amount)` — integrity check
- `UNIQUE (tenant_id, reference_number)` — no duplicate reference
- `UNIQUE (tenant_id, idempotency_key)` — idempotency deduplication
- Trigger: reject UPDATE except on `is_reversed` field
- Trigger: reject DELETE
- Index: `(tenant_id, rekening_id, created_at DESC)` — query per rekening
- Index: `(tenant_id, nasabah_id, created_at DESC)` — query per nasabah
- Index: `(tenant_id, branch_id, created_at DESC)` — query per branch
- Index: `(tenant_id, teller_session_id)` — query per teller session
- Index: `(tenant_id, batch_id)` — query per batch

### 12. Data Model — Transaksi Non-Rekening

```
kas_transaksi
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant) NOT NULL
├── branch_id             UUID (FK → branch) NOT NULL
│
├── ── Identitas ──
├── reference_number      VARCHAR UNIQUE per tenant
├── transaction_type      ENUM (kas_masuk, kas_keluar, transfer_kas)
├── category              VARCHAR NOT NULL (gaji, atk, listrik, sewa, lain_lain)
├── description           TEXT NOT NULL
│
├── ── Nominal ──
├── amount                NUMERIC(15,2) NOT NULL (selalu positif)
├── direction             ENUM (in, out) NOT NULL
│
├── ── Cash Info ──
├── teller_session_id     UUID (nullable, FK → teller_session)
│
├── ── Transfer Kas Info ──
├── source_branch_id      UUID (nullable, FK → branch, untuk transfer_kas)
├── destination_branch_id UUID (nullable, FK → branch, untuk transfer_kas)
├── transfer_kas_pair_id  UUID (nullable, FK → kas_transaksi)
│
├── ── Approval ──
├── approval_status       ENUM (none, pending, approved, rejected) DEFAULT 'none'
├── approved_by           UUID (nullable, FK → user)
├── approved_at           TIMESTAMPTZ (nullable)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ NOT NULL
    ├── created_by        UUID NOT NULL
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

### 13. Data Model — Transaction Batch

```
transaction_batch
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant) NOT NULL
├── branch_id             UUID (FK → branch) NOT NULL
│
├── ── Identitas ──
├── batch_type            ENUM (simpanan_wajib, bagi_hasil, admin_fee, denda)
├── description           TEXT NOT NULL
├── period                VARCHAR (misal: "2026-04" untuk batch bulanan)
│
├── ── Summary ──
├── total_items           INTEGER NOT NULL DEFAULT 0
├── success_count         INTEGER NOT NULL DEFAULT 0
├── failed_count          INTEGER NOT NULL DEFAULT 0
├── total_amount          NUMERIC(15,2) NOT NULL DEFAULT 0
│
├── ── Workflow ──
├── status                ENUM (draft, pending, approved, processing,
│                               completed, cancelled)
├── approved_by           UUID (nullable, FK → user)
├── approved_at           TIMESTAMPTZ (nullable)
├── executed_at           TIMESTAMPTZ (nullable)
├── completed_at          TIMESTAMPTZ (nullable)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ NOT NULL
    ├── created_by        UUID NOT NULL
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

### 14. Vernon _rels dan _data Structure

Transaksi menggunakan Vernon pattern karena listing membutuhkan data nasabah, rekening, dan branch (≥ 3 JOIN).

**_rels (transaksi):**
```json
{
  "tenant_id":         "018f...",
  "branch_id":         "018f...",
  "rekening_id":       "018f...",
  "nasabah_id":        "018f...",
  "teller_session_id": "018f..."
}
```

**_data (transaksi):**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000042"
  },
  "rekening": {
    "id":             "018f...",
    "account_number": "TB-2026-JKT-00000001",
    "category":       "tabungan",
    "product_name":   "Tabungan Berkah"
  },
  "branch": {
    "id":   "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  },
  "teller": {
    "id":   "018f...",
    "name": "Ahmad Teller"
  }
}
```

**Catatan keamanan:**
- **Tidak ada data sensitif** di `_data` — no saldo snapshot, no identity_number
- `balance_before` dan `balance_after` ada di kolom asli transaksi, bukan di `_data`
- `_data` hanya berisi informasi untuk listing/display

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.nasabah` di semua transaksi nasabah tersebut
- `RekeningUpdatedEvent` → update `_data.rekening` di semua transaksi rekening tersebut
- `BranchUpdatedEvent` → update `_data.branch` di semua transaksi cabang tersebut

### 15. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| Setoran tunai | v | v | v | v |
| Penarikan tunai (≤ threshold) | v | v | v | v |
| Penarikan tunai (> threshold, perlu specimen) | v* | v | v | v |
| Transfer antar rekening | v | v | v | v |
| Angsuran pinjaman | v | v | v | v |
| Pencairan pinjaman | - | - | v | v |
| Biaya admin (manual) | - | v | v | v |
| Denda (manual) | - | v | v | v |
| Koreksi / Reversal — request | v | v | v | v |
| Koreksi / Reversal — approve | - | - | v | v |
| Batch — create | - | v | v | v |
| Batch — approve & execute | - | - | v | v |
| Kas keluar (operasional) | - | v | v | v |
| Kas masuk (non-anggota) | v | v | v | v |
| Transfer kas antar cabang | - | v** | v | v |
| Search cross-branch | - | - | v | v |
| Export transaksi | - | v | v | v |
| Reprint receipt | v | v | v | v |

`*` Teller bisa proses tapi harus verifikasi specimen terlebih dahulu
`**` Supervisor membutuhkan approval Manager untuk transfer kas antar cabang

**Catatan:**
- Teller bisa memproses transaksi harian standar tapi **tidak bisa** approve reversal/koreksi
- Pencairan pinjaman hanya Manager+ karena risiko finansial tinggi
- Batch operations selalu butuh Manager+ approval
- Kas keluar membutuhkan minimal Supervisor untuk mencegah pengeluaran tidak sah

### 16. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Transaction receipt title | Bukti Transaksi | Bukti Transaksi |
| Interest distribution label | Distribusi Bunga | Distribusi Bagi Hasil |
| Loan installment | Angsuran Pinjaman | Angsuran Pembiayaan |
| Loan disbursement | Pencairan Pinjaman | Pencairan Pembiayaan |
| Interest/profit credit | Bunga Tabungan/Deposito | Bagi Hasil Tabungan/Deposito |
| Penalty charge | Denda Keterlambatan | Ta'zir / Denda Keterlambatan |
| Admin fee | Biaya Administrasi | Biaya Administrasi |

BMT mode menggunakan terminologi syariah pada label transaksi dan receipt, tapi **tipe transaksi di database tetap sama** — perbedaan hanya di presentation layer.

### 17. Integration Points

Transaksi terintegrasi dengan ADR lain:

| ADR | Integrasi |
|-----|-----------|
| [K001 Nasabah](./ADR-K001-nasabah.md) | KYC limit check, specimen verification |
| [K002 Rekening](./ADR-K002-rekening.md) | Balance update, status check, available_balance |
| K003 Produk & Akad | Product-level limits, min balance |
| K004 Simpanan Pokok & Wajib | Setoran wajib, batch collection |
| K005 Tabungan | Setoran/penarikan, bagi hasil distribution |
| K006 Deposito | Pencairan, rollover, bunga/bagi hasil |
| K007 Pinjaman | Pencairan, angsuran |
| K008 Angsuran & Jadwal | Link angsuran ke jadwal |
| K009 Denda & Penalti | Charge denda via transaksi |
| [K012 Teller Session](./ADR-K012-teller-session.md) | Cash transaction linked to session |
| [K014 Kas & Cash Flow](./ADR-K014-kas-cashflow.md) | Daily cash position from transactions |
| K015 Jurnal & Akuntansi | Auto-generate journal entry per transaction |

## Consequences

### Positif

- **Immutable audit trail** — setiap pergerakan dana tercatat permanen, tidak bisa diedit/dihapus
- **Atomic consistency** — saldo selalu konsisten karena update dalam satu database transaction
- **Comprehensive validation** — 7-layer validation pipeline mencegah transaksi invalid
- **Batch capability** — operasi massal (simpanan wajib, bagi hasil) bisa diproses efisien
- **Transfer safety** — debit + credit atomik mencegah dana "hilang" di tengah proses
- **Fraud prevention** — reversal butuh Manager+ approval, specimen verification untuk nominal besar
- **Traceability** — reference number, transfer pair, batch id memudahkan tracing end-to-end
- **Performant listing** — Vernon pattern menghilangkan JOIN untuk listing transaksi

### Negatif

- **Write overhead** — setiap transaksi membutuhkan multiple INSERT + UPDATE dalam satu transaction
- **Validation latency** — 7-layer validation menambah latency per transaksi (mitigated by in-memory checks)
- **Immutability rigidity** — koreksi harus via reversal + transaksi baru, lebih lambat dari edit langsung
- **Vernon sync load** — update nama nasabah bisa trigger propagasi ke ribuan transaksi
- **Batch processing time** — batch besar (ribuan nasabah) bisa memakan waktu signifikan

### Mitigasi

- Row-level lock (bukan table lock) meminimasi contention pada transaksi concurrent
- Validasi dilakukan di application layer (in-memory) sebelum hit database — latency minimal
- Reversal flow didesain streamlined: request → approve → execute, bukan birokrasi berlapis
- SyncEngine menggunakan batching + background job untuk propagasi Vernon massal
- Batch processing menggunakan worker pool dengan configurable concurrency

## Alternatives Considered

### A. Mutable Transactions (Allow Edit/Delete)

Mengizinkan edit atau delete transaksi yang sudah committed.

**Ditolak** karena: melanggar prinsip dasar audit trail keuangan. Tidak ada cara untuk membuktikan integritas data jika transaksi bisa diubah. Regulasi koperasi dan OJK mewajibkan pencatatan yang tidak bisa dimanipulasi.

### B. Event Sourcing untuk Transaksi

Menyimpan semua perubahan sebagai events, rebuild state dari event stream.

**Ditolak** karena: over-engineering untuk skala koperasi sekolah. Event sourcing menambah complexity signifikan (event store, projection, eventual consistency) tanpa benefit yang sepadan. Immutable INSERT + cached balance sudah memberikan audit trail yang memadai dengan arsitektur yang lebih sederhana.

### C. Balance dari SUM Transaksi (tanpa Cache)

Menghitung saldo real-time dari SUM seluruh transaksi tanpa cache di rekening.

**Ditolak** karena: sudah diputuskan di [K002](./ADR-K002-rekening.md) bahwa cached balance + reconciliation lebih performant. Konsisten dengan keputusan tersebut.

### D. Batch sebagai Satu Transaksi Besar

Memperlakukan seluruh batch sebagai satu transaksi database tunggal.

**Ditolak** karena: batch dengan ribuan item bisa lock banyak row terlalu lama dan menyebabkan timeout. Item-per-item processing dengan summary lebih resilient — kegagalan satu item tidak menggagalkan seluruh batch.

### E. Approval untuk Semua Transaksi

Setiap transaksi membutuhkan approval sebelum dieksekusi.

**Ditolak** karena: operasional harian koperasi (setoran, penarikan kecil) membutuhkan throughput tinggi. Approval hanya diperlukan untuk operasi berisiko (reversal, pencairan pinjaman, batch). Transaksi standar cukup divalidasi oleh validation engine.
