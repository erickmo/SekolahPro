# ADR-S051: School Payment Gateway (Gateway Pembayaran)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Pembayaran online menjadi kebutuhan esensial bagi sekolah modern. Orang tua mengharapkan kemudahan membayar SPP, biaya kegiatan, dan top-up kantin secara digital tanpa harus datang ke sekolah.

Payment gateway diperlukan untuk:

1. **Pembayaran SPP online**: Integrasi dengan S009 (student finance) — orang tua bisa bayar tagihan via VA, QRIS, atau transfer bank.
2. **Top-up kantin**: Integrasi dengan S037 (canteen billing) — saldo kantin bisa di-top-up online.
3. **Multi-channel**: Virtual Account (per bank), QRIS (universal), transfer bank manual.
4. **Auto-generate VA**: Setiap siswa mendapat nomor VA unik yang tetap — orang tua cukup simpan nomor ini.
5. **Rekonsiliasi otomatis**: Callback dari payment provider otomatis matching ke invoice yang tepat.
6. **Settlement tracking**: Pelacakan dana masuk dari payment provider ke rekening sekolah.
7. **Pesantren**: Mendukung pembayaran syahriyyah (SPP pesantren), kitab, dan biaya pondok.

### Mengapa CQRS Pattern (bukan Vernon)?

- **Transaction-heavy**: Setiap pembayaran = write operation yang harus konsisten dan idempotent.
- **Concurrency tinggi**: Banyak orang tua bayar bersamaan, terutama awal bulan.
- **Financial accuracy**: Tidak boleh ada eventual consistency untuk transaksi keuangan — harus strong consistency.
- **Callback processing**: Payment provider mengirim webhook yang harus diproses secara atomic.
- **Reconciliation**: Proses matching otomatis memerlukan command/query separation yang jelas.

## Decision

Menggunakan **CQRS Pattern** untuk domain payment gateway. Write path (commands) menangani payment creation, callback processing, dan reconciliation. Read path (queries) menangani payment status, history, dan reporting.

### Table Schema

```sql
-- Payment channels yang aktif per sekolah
CREATE TABLE payment_channels (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    channel_type    VARCHAR(20) NOT NULL,
    channel_name    VARCHAR(100) NOT NULL,
    provider_code   VARCHAR(30) NOT NULL,

    -- Konfigurasi provider
    config          JSONB NOT NULL DEFAULT '{}',
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Fee
    admin_fee_type  VARCHAR(10) NOT NULL DEFAULT 'flat',
    admin_fee_amount BIGINT NOT NULL DEFAULT 0,
    admin_fee_percent NUMERIC(5,2) NOT NULL DEFAULT 0,
    fee_bearer      VARCHAR(10) NOT NULL DEFAULT 'parent',

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_channel_provider UNIQUE (tenant_id, company_id, provider_code),
    CONSTRAINT chk_channel_type CHECK (channel_type IN ('virtual_account', 'qris', 'bank_transfer', 'retail', 'ewallet')),
    CONSTRAINT chk_fee_type CHECK (admin_fee_type IN ('flat', 'percent', 'mixed')),
    CONSTRAINT chk_fee_bearer CHECK (fee_bearer IN ('parent', 'school', 'split'))
);

CREATE INDEX idx_channel_tenant ON payment_channels (tenant_id, company_id);
CREATE INDEX idx_channel_active ON payment_channels (is_active) WHERE is_active = true;

-- Virtual Account per siswa (persistent)
CREATE TABLE student_virtual_accounts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    channel_id      UUID NOT NULL,

    -- VA detail
    va_number       VARCHAR(30) NOT NULL,
    bank_code       VARCHAR(10) NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_va_number UNIQUE (va_number),
    CONSTRAINT uq_student_bank UNIQUE (tenant_id, company_id, student_id, bank_code)
);

CREATE INDEX idx_va_student ON student_virtual_accounts (student_id);
CREATE INDEX idx_va_number ON student_virtual_accounts (va_number);
CREATE INDEX idx_va_tenant ON student_virtual_accounts (tenant_id, company_id);

-- Transaksi pembayaran (command side — append-only)
CREATE TABLE payment_transactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Referensi
    transaction_no  VARCHAR(50) NOT NULL,
    student_id      UUID NOT NULL,
    channel_id      UUID NOT NULL,

    -- Link ke domain asal
    payment_type    VARCHAR(20) NOT NULL,
    reference_id    UUID NOT NULL,
    reference_type  VARCHAR(30) NOT NULL,

    -- Amount
    amount          BIGINT NOT NULL,
    admin_fee       BIGINT NOT NULL DEFAULT 0,
    total_amount    BIGINT NOT NULL,

    -- Payment detail
    payment_method  VARCHAR(20) NOT NULL,
    va_number       VARCHAR(30),
    external_id     VARCHAR(100),

    -- Status lifecycle
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    paid_at         TIMESTAMPTZ,
    expired_at      TIMESTAMPTZ,
    settled_at      TIMESTAMPTZ,

    -- Provider response
    provider_ref    VARCHAR(100),
    provider_data   JSONB DEFAULT '{}',

    -- Idempotency
    idempotency_key VARCHAR(100) NOT NULL,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_transaction_no UNIQUE (tenant_id, company_id, transaction_no),
    CONSTRAINT uq_idempotency UNIQUE (idempotency_key),
    CONSTRAINT chk_payment_type CHECK (payment_type IN ('spp', 'fee', 'canteen_topup', 'admission', 'other')),
    CONSTRAINT chk_reference_type CHECK (reference_type IN ('student_invoice', 'canteen_wallet', 'applicant')),
    CONSTRAINT chk_tx_status CHECK (status IN ('pending', 'paid', 'expired', 'failed', 'refunded', 'settled')),
    CONSTRAINT chk_tx_method CHECK (payment_method IN ('virtual_account', 'qris', 'bank_transfer', 'retail', 'ewallet')),
    CONSTRAINT chk_tx_amount CHECK (amount > 0),
    CONSTRAINT chk_tx_total CHECK (total_amount > 0)
);

CREATE INDEX idx_tx_tenant ON payment_transactions (tenant_id, company_id);
CREATE INDEX idx_tx_student ON payment_transactions (student_id);
CREATE INDEX idx_tx_status ON payment_transactions (status);
CREATE INDEX idx_tx_pending ON payment_transactions (expired_at) WHERE status = 'pending';
CREATE INDEX idx_tx_reference ON payment_transactions (reference_type, reference_id);
CREATE INDEX idx_tx_external ON payment_transactions (external_id) WHERE external_id IS NOT NULL;
CREATE INDEX idx_tx_provider_ref ON payment_transactions (provider_ref) WHERE provider_ref IS NOT NULL;
CREATE INDEX idx_tx_date ON payment_transactions (created_at);
CREATE INDEX idx_tx_paid ON payment_transactions (paid_at) WHERE paid_at IS NOT NULL;

-- Callback log dari payment provider (append-only, immutable)
CREATE TABLE payment_callbacks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Referensi
    transaction_id  UUID,
    external_id     VARCHAR(100),

    -- Callback data
    provider_code   VARCHAR(30) NOT NULL,
    callback_type   VARCHAR(20) NOT NULL,
    raw_payload     JSONB NOT NULL,
    signature       TEXT,

    -- Processing
    process_status  VARCHAR(20) NOT NULL DEFAULT 'received',
    process_error   TEXT,
    processed_at    TIMESTAMPTZ,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_callback_type CHECK (callback_type IN ('payment', 'settlement', 'refund', 'expiry')),
    CONSTRAINT chk_process_status CHECK (process_status IN ('received', 'processing', 'processed', 'failed', 'ignored'))
);

CREATE INDEX idx_callback_tx ON payment_callbacks (transaction_id);
CREATE INDEX idx_callback_external ON payment_callbacks (external_id);
CREATE INDEX idx_callback_status ON payment_callbacks (process_status) WHERE process_status IN ('received', 'failed');
CREATE INDEX idx_callback_date ON payment_callbacks (created_at);

-- Settlement / Pencairan dana (read model)
CREATE TABLE payment_settlements (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Batch settlement
    settlement_date DATE NOT NULL,
    provider_code   VARCHAR(30) NOT NULL,
    batch_ref       VARCHAR(100),

    -- Amount
    total_transactions INT NOT NULL,
    gross_amount    BIGINT NOT NULL,
    fee_amount      BIGINT NOT NULL,
    net_amount      BIGINT NOT NULL,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    bank_ref        VARCHAR(100),

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_settlement_status CHECK (status IN ('pending', 'settled', 'discrepancy'))
);

CREATE INDEX idx_settlement_tenant ON payment_settlements (tenant_id, company_id);
CREATE INDEX idx_settlement_date ON payment_settlements (settlement_date);
CREATE INDEX idx_settlement_status ON payment_settlements (status);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `va_number` | VARCHAR(30), globally unique | VA number harus unique di seluruh sistem (bukan per tenant) — bank requirement |
| `idempotency_key` | VARCHAR(100), unique | Mencegah duplicate transaction dari double-click atau retry |
| `external_id` | VARCHAR(100) | ID dari payment provider untuk cross-reference |
| `provider_data` | JSONB | Response mentah dari provider — berbeda format per provider |
| `raw_payload` | JSONB di callbacks | Immutable log dari webhook — untuk audit dan debugging |
| `reference_type` + `reference_id` | Polymorphic reference | Satu transaksi bisa untuk invoice (S009), canteen wallet (S037), atau applicant (S016) |
| `admin_fee` | BIGINT terpisah | Biaya admin harus transparan — bisa ditanggung orang tua, sekolah, atau split |
| `fee_bearer` | VARCHAR(10) | Konfigurasi per channel siapa yang menanggung biaya admin |
| `payment_callbacks` | Append-only, no update | Immutable audit log — setiap callback dari provider tercatat apa adanya |
| `settled_at` | TIMESTAMPTZ nullable | Tanggal dana benar-benar masuk ke rekening sekolah (beda dengan paid_at) |

### CQRS Command/Query Separation

**Commands (Write Path):**

| Command | Handler | Side Effect |
|---|---|---|
| `CreatePaymentTransaction` | Buat transaksi + request ke provider | Create VA/QRIS di provider |
| `ProcessPaymentCallback` | Terima webhook, validate, update status | Update invoice (S009) / wallet (S037) |
| `ExpireStaleTransactions` | Cron: expire transaksi yang lewat deadline | Cleanup pending transactions |
| `ReconcileSettlement` | Matching settlement batch dari provider | Update settlement status |
| `ProcessRefund` | Refund transaksi yang sudah paid | Reverse invoice payment |

**Queries (Read Path):**

| Query | Source | Cache |
|---|---|---|
| `GetTransactionStatus` | payment_transactions | No cache (real-time) |
| `ListStudentTransactions` | payment_transactions + JOIN | Short TTL cache |
| `GetPaymentSummary` | Aggregation | Redis cache 5 min |
| `GetSettlementReport` | payment_settlements | Long TTL cache |
| `GetReconciliationStatus` | transactions vs settlements | Computed on demand |

### Payment Flow

```
1. Orang tua pilih tagihan (invoice) yang akan dibayar
2. System:
   a. Lookup student VA number (atau generate jika belum ada)
   b. Create payment_transaction (status = 'pending')
   c. Call payment provider API (create VA / QRIS)
   d. Return payment instruction ke orang tua
3. Orang tua bayar via banking app
4. Payment provider kirim callback (webhook):
   a. Log ke payment_callbacks (immutable)
   b. Validate signature
   c. Match ke payment_transaction via external_id
   d. Update transaction status → 'paid'
   e. Side effect: Update student_invoice.paid_amount (S009)
   f. Side effect: Update canteen_wallet.balance (S037) jika top-up
5. Provider melakukan settlement (T+1 atau T+2):
   a. Batch settlement data masuk
   b. Create/update payment_settlement
   c. Match individual transactions
   d. Update transaction status → 'settled'
```

### API Endpoints

```
# Payment Channels (Admin)
GET    /api/v1/payment-channels                        — List payment channels
POST   /api/v1/payment-channels                        — Configure channel
PUT    /api/v1/payment-channels/{id}                   — Update channel config

# Virtual Accounts
GET    /api/v1/students/{id}/virtual-accounts           — VA numbers per siswa
POST   /api/v1/students/{id}/virtual-accounts           — Generate VA untuk siswa

# Transactions (Parent-facing)
POST   /api/v1/payments/create                          — Buat transaksi pembayaran
GET    /api/v1/payments/{id}/status                     — Cek status pembayaran
GET    /api/v1/students/{id}/payment-history             — Riwayat pembayaran

# Webhooks (Provider callback)
POST   /api/v1/webhooks/midtrans                        — Midtrans callback
POST   /api/v1/webhooks/xendit                          — Xendit callback
POST   /api/v1/webhooks/duitku                          — Duitku callback

# Reconciliation (Admin)
GET    /api/v1/payments/settlements                     — List settlement batches
GET    /api/v1/payments/reconciliation                  — Reconciliation report
POST   /api/v1/payments/reconcile                       — Trigger manual reconciliation

# Reports
GET    /api/v1/payments/summary                         — Summary per periode
GET    /api/v1/payments/by-channel                      — Breakdown per channel
GET    /api/v1/payments/by-type                          — Breakdown per payment type
```

## Fee Absorption Model

> **C-Suite CFO Review Note (2026-04-15):**
> ADR ini mendefinisikan `fee_bearer` (parent/school/split) per channel tapi belum menganalisis
> dampak finansial. Berikut analisis lengkap:

### Fee Structure per Provider (Estimasi 2026)

| Channel | Provider | Fee per Transaksi | Contoh (SPP Rp 500.000) |
|---------|----------|-------------------|--------------------------|
| Virtual Account | Midtrans | Rp 4.000 flat | Rp 4.000 |
| Virtual Account | Xendit | Rp 4.500 flat | Rp 4.500 |
| QRIS | Midtrans | 0.7% (MDR) | Rp 3.500 |
| QRIS | Xendit | 0.7% (MDR) | Rp 3.500 |
| Bank Transfer | Manual verify | Rp 0 (tapi butuh staff) | Rp 0 + labor cost |
| E-wallet (GoPay/OVO) | Midtrans | 2.0% | Rp 10.000 |
| Retail (Alfamart/Indomaret) | Xendit | Rp 5.000 flat | Rp 5.000 |

### Simulasi Biaya Bulanan

Asumsi: 500 siswa, SPP Rp 500.000/bulan, 80% bayar online:

| Skenario | Fee/bulan | Ditanggung |
|----------|-----------|------------|
| **100% VA (flat Rp 4.000)** | 400 tx × Rp 4.000 = **Rp 1.600.000** | Sekolah atau orang tua |
| **70% VA + 30% QRIS** | 280 × Rp 4.000 + 120 × Rp 3.500 = **Rp 1.540.000** | Campuran |
| **100% QRIS (0.7%)** | 400 × Rp 3.500 = **Rp 1.400.000** | Sekolah (MDR regulation) |

### Rekomendasi Model per Segmen Sekolah

| Segmen | Model | Alasan |
|--------|-------|--------|
| **Pesantren besar** (SPP tinggi) | `fee_bearer = school` | Absorb fee sebagai service cost, include di SPP |
| **Sekolah swasta menengah** | `fee_bearer = parent` + transparansi | Orang tua terbiasa bayar biaya admin ATM/transfer |
| **Sekolah dengan budget ketat** | `fee_bearer = split` (50:50) | Kompromi — orang tua bayar sebagian, sekolah subsidize |

### Konfigurasi Default

```json
{
  "fee_model": {
    "default_bearer": "parent",
    "show_fee_to_parent": true,
    "max_surcharge_percent": 2.0,
    "channels": {
      "virtual_account": { "bearer": "parent", "flat_fee": 4000 },
      "qris": { "bearer": "school", "note": "MDR regulation: merchant bears QRIS fee" },
      "bank_transfer": { "bearer": "school", "flat_fee": 0 }
    }
  }
}
```

> **Catatan regulasi QRIS:** Per regulasi BI, biaya MDR QRIS **ditanggung merchant** (sekolah),
> tidak boleh dibebankan ke konsumen (orang tua). Jadi untuk channel QRIS, `fee_bearer` harus selalu `school`.

## Consequences

### Positive

- **Multi-provider**: Arsitektur provider-agnostic — bisa switch atau pakai multiple provider (Midtrans, Xendit, Duitku).
- **Persistent VA**: Siswa punya VA tetap — orang tua cukup simpan 1 nomor, bayar kapan saja.
- **Idempotent**: Idempotency key mencegah duplicate payment dari retry atau double-click.
- **Immutable audit**: Callback log tidak bisa diubah — full audit trail untuk dispute.
- **Auto-reconciliation**: Settlement matching otomatis mengurangi kerja manual bendahara.
- **Integration-ready**: Polymorphic reference mendukung payment untuk berbagai domain (SPP, canteen, admission).

### Negative / Trade-offs

- **Provider dependency**: Tergantung pada uptime dan API stability payment provider.
- **Settlement delay**: Dana tidak real-time masuk ke sekolah (T+1 sampai T+7 tergantung provider).
- **Callback reliability**: Webhook bisa gagal / delay — perlu retry mechanism dan fallback polling.
- **Fee complexity**: Admin fee calculation berbeda per provider, per channel, per amount range.
- **PCI compliance**: Meskipun tidak menyimpan card data, tetap perlu security best practices untuk handling payment data.
- **No Vernon pattern**: Tidak ada `_rels/_data` denormalization — semua query via JOIN atau materialized view.

## Alternatives Considered

### 1. Vernon Pattern untuk payment
- Ditolak: transaksi keuangan membutuhkan strong consistency, bukan eventual consistency. CQRS lebih tepat.

### 2. Single payment provider only
- Ditolak: vendor lock-in. Sekolah di daerah berbeda mungkin butuh provider berbeda. Provider-agnostic abstraction layer lebih fleksibel.

### 3. Direct bank integration (tanpa payment gateway)
- Ditolak: setiap bank punya API berbeda, sertifikasi mahal. Payment aggregator lebih efisien.

### 4. Manual transfer + upload bukti
- Ditolak sebagai primary flow: terlalu manual dan rawan error. Tapi tetap bisa sebagai fallback (bank_transfer channel dengan manual verification).

### 5. Store payment data di S009 (student_payments)
- Ditolak: payment gateway perlu data tambahan (callback, settlement, provider ref) yang tidak relevan di domain student finance. Lebih baik terpisah dengan cross-reference via reference_id.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | Generate idempotency key | student_id + invoice_id + timestamp | Deterministic unique key |
| U02 | Calculate admin fee (flat) | amount=500000, flat=2500 | total=502500 |
| U03 | Calculate admin fee (percent) | amount=500000, percent=1.5% | total=507500 |
| U04 | Validate rejects invalid `payment_type` | `"donation"` | Error |
| U05 | Validate rejects invalid `status` | `"cancelled"` | Error |
| U06 | Validate rejects `amount <= 0` | amount=0 | Error |
| U07 | Validate callback signature (Midtrans) | Valid signature | OK |
| U08 | Validate callback signature (invalid) | Tampered signature | Error |
| U09 | Status transition valid: pending → paid | — | OK |
| U10 | Status transition invalid: paid → pending | — | Error |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Configure payment channel | POST with Midtrans config | 201, channel active |
| I02 | Generate VA for student | POST /students/{id}/virtual-accounts | 201, VA number assigned |
| I03 | Unique VA globally | Generate 2 VAs with same number | Unique constraint error |
| I04 | Create payment transaction | POST /payments/create for invoice | 201, status=pending, VA instructions returned |
| I05 | Idempotency prevents duplicate | POST same payment twice with same key | Second returns existing transaction |
| I06 | Process payment callback | POST webhook with valid payload | Transaction status → paid |
| I07 | Callback updates invoice | Process SPP payment callback | student_invoice.paid_amount updated (S009) |
| I08 | Callback updates canteen wallet | Process top-up callback | canteen_wallet.balance updated (S037) |
| I09 | Invalid callback signature | POST webhook with bad signature | 400, callback logged as failed |
| I10 | Expire stale transactions | Run expiry cron after deadline | Pending transactions → expired |
| I11 | Settlement reconciliation | Process settlement batch | Paid transactions → settled |
| I12 | Payment summary report | GET /summary | 200, aggregated by period |
| I13 | Concurrent payment handling | 2 payments for same invoice | Only one succeeds, other rejected |

### Integration Tests — Cross-Domain

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | SPP payment end-to-end | Create tx → callback → invoice update | Invoice status changes accordingly |
| I15 | Canteen top-up end-to-end | Create tx → callback → wallet update | Wallet balance increased |
| I16 | Partial SPP payment | Pay less than total invoice | Invoice status → partial |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I17 | Cannot access other tenant's transactions | GET with wrong tenant | 404 |
| I18 | Cannot process callback cross-tenant | Callback for other tenant's tx | Error |
| I19 | VA unique across tenants | Same VA cannot exist in 2 tenants | Unique constraint |
