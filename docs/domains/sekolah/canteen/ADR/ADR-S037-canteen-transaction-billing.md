# ADR-S037: Canteen Transaction & Billing / Transaksi & Tagihan Kantin

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Setelah menu dan jadwal dikelola (ADR-S036), diperlukan domain untuk **transaksi makan dan penagihan** ke orang tua. Ini terutama penting untuk boarding school/pesantren di mana biaya makan terpisah dari SPP (ADR-S009).

Data transaksi kantin diperlukan untuk:

1. **Meal tracking**: Pencatatan siapa yang makan di meal mana — untuk monitoring dan billing.
2. **Prepaid balance**: Saldo prepaid/meal plan per siswa — top-up oleh orang tua.
3. **Billing**: Tagihan bulanan biaya makan untuk siswa non-prepaid.
4. **Top-up**: Mekanisme pengisian saldo via transfer/cash/QRIS.
5. **Laporan**: Rekap konsumsi per siswa, per bulan — untuk orang tua dan manajemen.
6. **Integrasi S009**: Biaya kantin bisa menjadi salah satu fee_type di student finance.

Konteks pesantren/boarding school Indonesia:
- Model billing bervariasi: ada yang **flat monthly** (biaya makan termasuk SPP), ada yang **per-meal** (dihitung per porsi).
- Beberapa pesantren menggunakan **kartu makan** (tap card) — integrasi reader di masa depan.
- Santri yang izin pulang (ADR-S034 permission) **tidak ditagih** untuk meal yang tidak dikonsumsi.
- Orang tua sering minta **rincian** makan anak (bukan hanya total tagihan).

### Mengapa Vernon Pattern?

- Volume tinggi: 3-4 meals/hari × 500 santri = 1.500-2.000 transaksi/hari.
- Read-heavy untuk laporan bulanan dan dashboard orang tua.
- Relasi ke student, meal_schedule (S036), academic_year.
- Business logic moderate: balance deduction, billing aggregation.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `canteen_student_plans` (meal plan/saldo per siswa), `canteen_transactions` (transaksi makan harian), dan `canteen_topups` (pengisian saldo).

### Table Schema

```sql
-- Meal plan / saldo per siswa per tahun ajaran
CREATE TABLE canteen_student_plans (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Plan type
    plan_type       VARCHAR(20) NOT NULL,
    monthly_fee     BIGINT,

    -- Saldo (untuk prepaid)
    balance         BIGINT NOT NULL DEFAULT 0,
    total_topup     BIGINT NOT NULL DEFAULT 0,
    total_consumed  BIGINT NOT NULL DEFAULT 0,

    -- Meals included
    includes_breakfast BOOLEAN NOT NULL DEFAULT true,
    includes_lunch     BOOLEAN NOT NULL DEFAULT true,
    includes_dinner    BOOLEAN NOT NULL DEFAULT true,
    includes_snack     BOOLEAN NOT NULL DEFAULT false,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'active',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_canteen_plan_student_year UNIQUE (student_id, academic_year_id),
    CONSTRAINT chk_canteen_plan_type CHECK (plan_type IN ('flat_monthly', 'prepaid', 'per_meal')),
    CONSTRAINT chk_canteen_plan_status CHECK (status IN ('active', 'suspended', 'ended')),
    CONSTRAINT chk_canteen_balance CHECK (balance >= 0),
    CONSTRAINT chk_canteen_monthly_fee CHECK (monthly_fee IS NULL OR monthly_fee >= 0)
);

-- Indexes
CREATE INDEX idx_canteen_plan_tenant_company ON canteen_student_plans (tenant_id, company_id);
CREATE INDEX idx_canteen_plan_student ON canteen_student_plans (student_id);
CREATE INDEX idx_canteen_plan_year ON canteen_student_plans (academic_year_id);
CREATE INDEX idx_canteen_plan_status ON canteen_student_plans (status) WHERE status = 'active';
CREATE INDEX idx_canteen_plan_rels ON canteen_student_plans USING GIN (_rels);
CREATE INDEX idx_canteen_plan_data ON canteen_student_plans USING GIN (_data);

-- Transaksi makan harian
CREATE TABLE canteen_transactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    plan_id         UUID NOT NULL,
    schedule_id     UUID,
    academic_year_id UUID NOT NULL,

    -- Transaksi
    transaction_date DATE NOT NULL,
    meal_type       VARCHAR(20) NOT NULL,
    amount          BIGINT NOT NULL DEFAULT 0,
    status          VARCHAR(20) NOT NULL DEFAULT 'consumed',

    -- Catatan
    note            TEXT,
    recorded_by     UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_canteen_txn_student_date_meal UNIQUE (student_id, transaction_date, meal_type),
    CONSTRAINT chk_canteen_txn_meal CHECK (meal_type IN ('breakfast', 'lunch', 'dinner', 'snack')),
    CONSTRAINT chk_canteen_txn_status CHECK (status IN ('consumed', 'skipped', 'absent')),
    CONSTRAINT chk_canteen_txn_amount CHECK (amount >= 0)
);

-- Indexes
CREATE INDEX idx_canteen_txn_tenant_company ON canteen_transactions (tenant_id, company_id);
CREATE INDEX idx_canteen_txn_student ON canteen_transactions (student_id);
CREATE INDEX idx_canteen_txn_plan ON canteen_transactions (plan_id);
CREATE INDEX idx_canteen_txn_date ON canteen_transactions (transaction_date);
CREATE INDEX idx_canteen_txn_student_date ON canteen_transactions (student_id, transaction_date);
CREATE INDEX idx_canteen_txn_status ON canteen_transactions (status);
CREATE INDEX idx_canteen_txn_rels ON canteen_transactions USING GIN (_rels);
CREATE INDEX idx_canteen_txn_data ON canteen_transactions USING GIN (_data);

-- Pengisian saldo (top-up)
CREATE TABLE canteen_topups (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    plan_id         UUID NOT NULL,
    student_id      UUID NOT NULL,

    -- Top-up detail
    receipt_no      VARCHAR(50) NOT NULL,
    amount          BIGINT NOT NULL,
    payment_method  VARCHAR(20) NOT NULL,
    topup_date      DATE NOT NULL,
    note            TEXT,

    -- Pencatat
    received_by     UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_canteen_topup_receipt UNIQUE (tenant_id, company_id, receipt_no),
    CONSTRAINT chk_canteen_topup_amount CHECK (amount > 0),
    CONSTRAINT chk_canteen_topup_method CHECK (payment_method IN ('cash', 'transfer', 'debit', 'qris', 'va'))
);

-- Indexes
CREATE INDEX idx_canteen_topup_tenant_company ON canteen_topups (tenant_id, company_id);
CREATE INDEX idx_canteen_topup_plan ON canteen_topups (plan_id);
CREATE INDEX idx_canteen_topup_student ON canteen_topups (student_id);
CREATE INDEX idx_canteen_topup_date ON canteen_topups (topup_date);
CREATE INDEX idx_canteen_topup_rels ON canteen_topups USING GIN (_rels);
CREATE INDEX idx_canteen_topup_data ON canteen_topups USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `plan_type` | flat_monthly/prepaid/per_meal | 3 model billing: flat bulanan (termasuk SPP), prepaid (saldo isi ulang), per-meal (dihitung per porsi) |
| `monthly_fee` | BIGINT, nullable | Hanya untuk `flat_monthly` — biaya tetap per bulan |
| `balance` | BIGINT, default 0 | Saldo prepaid — diupdate setiap topup (tambah) dan transaksi (kurang) |
| `includes_*` | 4 BOOLEAN | Konfigurasi meal mana yang termasuk dalam plan — beberapa siswa hanya lunch (day school) |
| `amount` (transaction) | BIGINT, default 0 | Biaya per meal — 0 untuk flat_monthly (sudah termasuk), >0 untuk prepaid/per_meal |
| `status` (transaction) | consumed/skipped/absent | consumed = makan, skipped = sengaja tidak makan, absent = tidak hadir (izin pulang S034) |
| `payment_method` (topup) | 5 metode | Konsisten dengan S009 (student_payments) — cash, transfer, debit, QRIS, VA |
| `receipt_no` | VARCHAR(50), unique | Nomor kwitansi top-up — untuk audit trail |

### Integration with S009 (Student Finance)

```
Untuk plan_type = 'flat_monthly':
1. Biaya makan bisa dijadikan fee_type di S009 (code: 'MEAL')
2. Invoice di-generate per bulan via S009 mechanism
3. Pembayaran via S009 student_payments

Untuk plan_type = 'prepaid':
1. Top-up dicatat di canteen_topups (domain ini)
2. Balance deducted per meal
3. Tidak masuk invoice S009

Untuk plan_type = 'per_meal':
1. Transaksi dihitung per bulan
2. Invoice di-generate via S009 berdasarkan aggregate canteen_transactions
```

### Vernon Relationships

**canteen_student_plans:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Nama siswa selalu ditampilkan |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

**canteen_transactions:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Nama siswa |
| `plan` | belongs_to | **Ya** | Tipe plan untuk kalkulasi |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

**canteen_topups:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `plan` | belongs_to | **Ya** | Plan terkait |
| `student` | belongs_to | **Ya** | Nama siswa |

### _rels / _data Structure

```json
// canteen_student_plans
{
  "_rels": {
    "student_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad Fauzi", "nis": "12345" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}

// canteen_transactions
{
  "_rels": {
    "student_id": "018f...",
    "plan_id": "018f...",
    "schedule_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad Fauzi", "nis": "12345" },
    "plan": { "id": "018f...", "plan_type": "prepaid", "balance": 750000 },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}

// canteen_topups
{
  "_rels": {
    "plan_id": "018f...",
    "student_id": "018f..."
  },
  "_data": {
    "plan": { "id": "018f...", "plan_type": "prepaid" },
    "student": { "id": "018f...", "full_name": "Ahmad Fauzi", "nis": "12345" }
  }
}
```

### API Endpoints

```
# Student Plans
GET    /api/v1/students/{id}/canteen-plan               — Plan makan siswa
GET    /api/v1/canteen-student-plans                     — List semua plan (admin)
POST   /api/v1/canteen-student-plans                     — Buat plan untuk siswa
POST   /api/v1/canteen-student-plans/bulk                — Bulk create plans (awal tahun)
PUT    /api/v1/canteen-student-plans/{id}                — Update plan

# Transactions
GET    /api/v1/students/{id}/canteen-transactions         — Riwayat makan siswa
GET    /api/v1/canteen-transactions                       — List transaksi (admin, filter by date)
POST   /api/v1/canteen-transactions/bulk                  — Bulk record makan per meal
GET    /api/v1/students/{id}/canteen-transactions/summary — Rekap per bulan

# Top-ups
GET    /api/v1/students/{id}/canteen-topups              — Riwayat top-up siswa
POST   /api/v1/canteen-topups                            — Record top-up
GET    /api/v1/canteen-topups/{id}/receipt                — Cetak kwitansi

# Reports
GET    /api/v1/canteen/billing-summary                   — Rekap tagihan bulanan (per_meal plan)
GET    /api/v1/canteen/consumption-report                 — Laporan konsumsi per siswa/kelas
```

### Bulk Transaction Payload

```json
{
  "academic_year_id": "018f...",
  "transaction_date": "2026-04-15",
  "meal_type": "lunch",
  "transactions": [
    { "student_id": "018f...", "status": "consumed" },
    { "student_id": "018f...", "status": "consumed" },
    { "student_id": "018f...", "status": "skipped", "note": "Tidak nafsu makan" },
    { "student_id": "018f...", "status": "absent", "note": "Izin pulang" }
  ]
}
```

## Consequences

### Positive

- **3 model billing**: Flat monthly, prepaid, dan per-meal mendukung berbagai model operasi pesantren/boarding school.
- **Balance tracking**: Saldo prepaid real-time — orang tua bisa monitor via dashboard.
- **Consumed/skipped/absent**: Granular tracking — santri yang izin pulang tidak ditagih.
- **Integrasi S009**: Model flat_monthly dan per_meal bisa di-bridge ke domain finance.
- **Audit trail**: Setiap top-up punya receipt_no, setiap transaksi punya recorded_by.
- **Bulk-friendly**: Endpoint bulk untuk efisiensi pencatatan per-meal.

### Negative / Trade-offs

- **Balance consistency**: `balance`, `total_topup`, `total_consumed` adalah denormalisasi — perlu dijaga konsisten via application logic. Drift bisa terjadi jika ada bug.
- **No refund**: Belum ada mekanisme refund saldo prepaid. Enhancement di masa depan.
- **No card integration**: Belum ada integrasi kartu tap/NFC untuk automatic meal tracking. Future hardware integration.
- **Cross-domain dependency**: Integrasi dengan S009 memerlukan event-based coordination — complexity tambahan.
- **No dietary matching**: Belum ada validasi bahwa meal yang dikonsumsi sesuai dietary restriction siswa.

## Alternatives Considered

### 1. Semua billing via S009 (tanpa domain kantin terpisah)
- Ditolak: S009 tidak bisa tracking per-meal consumption. Kantin butuh domain sendiri untuk granular reporting.

### 2. Balance di student table
- Ditolak: balance spesifik per tahun ajaran dan per plan — memerlukan tabel terpisah.

### 3. Transaction tanpa plan
- Ditolak: tanpa plan, tidak bisa menentukan model billing per siswa. Plan adalah konfigurasi essential.

### 4. Real-time balance deduction (sync)
- Ditolak: eventual consistency cukup — reconciliation bisa dilakukan end-of-day. Sync real-time menambah complexity tanpa manfaat signifikan untuk pesantren.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `PlanDescriptor.TableName()` | — | `"canteen_student_plans"` |
| U02 | `TransactionDescriptor.TableName()` | — | `"canteen_transactions"` |
| U03 | `TopupDescriptor.TableName()` | — | `"canteen_topups"` |
| U04 | Validate rejects invalid `plan_type` | `"subscription"` | Error: must be flat_monthly/prepaid/per_meal |
| U05 | Validate rejects invalid `meal_type` | `"brunch"` | Error: invalid meal_type |
| U06 | Validate rejects `topup amount <= 0` | `amount = 0` | Error: must be positive |
| U07 | Validate rejects invalid `payment_method` | `"bitcoin"` | Error: invalid method |
| U08 | Validate accepts valid transaction | All fields valid | No error |

### Integration Tests — Plans

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create plan | POST with student + year + plan_type | 201 |
| I02 | Unique per student per year | Create 2 plans same student same year | 409/422 |
| I03 | Plan type CHECK | INSERT with `plan_type = 'subscription'` | DB error |
| I04 | Bulk create plans | POST /bulk for 30 students | 201, 30 plans created |

### Integration Tests — Transactions

| # | Test Case | Action | Expected |
|---|---|---|---|
| I05 | Bulk record meal | POST bulk for lunch | 201, all transactions created |
| I06 | Unique per student per date per meal | Record same meal twice | 409/422 |
| I07 | Prepaid balance deducted | Record consumed for prepaid student | Balance decreased by amount |
| I08 | Skipped/absent not charged | Record skipped for prepaid student | Balance unchanged |
| I09 | Get student meal history | GET /students/{id}/canteen-transactions | 200, ordered by date |
| I10 | Get monthly summary | GET /students/{id}/canteen-transactions/summary | 200, counts per meal_type per status |

### Integration Tests — Top-ups

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Record top-up | POST with amount + method | 201, plan balance increased |
| I12 | Unique receipt number | Create 2 topups same receipt_no | 409/422 |
| I13 | Top-up amount CHECK | POST with amount = 0 | 422, rejected |
| I14 | Payment method CHECK | INSERT with `payment_method = 'bitcoin'` | DB error |
| I15 | Get topup receipt | GET /canteen-topups/{id}/receipt | 200, receipt data |

### Integration Tests — Balance Consistency

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Top-up updates total_topup | Record top-up 500000 | total_topup += 500000, balance += 500000 |
| I17 | Consumed updates total_consumed | Record consumed (amount 15000) | total_consumed += 15000, balance -= 15000 |
| I18 | Balance never negative | Consume more than balance | 422, insufficient balance |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I19 | StudentUpdated syncs to plans | Update student name | `_data.student.full_name` updated |
| I20 | StudentUpdated syncs to transactions | Update student name | `_data.student.full_name` updated |
| I21 | PlanUpdated syncs to transactions | Update plan_type | `_data.plan.plan_type` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I22 | Cannot access other tenant's plans | GET with wrong tenant | 404 |
| I23 | Cannot topup cross-tenant | POST topup for other tenant's plan | Error |
