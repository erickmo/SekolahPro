-- 068: toko_penjualan — Penjualan POS toko/kantin (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_penjualan (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id               UUID NOT NULL,
    company_id              UUID NOT NULL,

    location_id             UUID NOT NULL,
    receipt_number          VARCHAR NOT NULL,
    sale_date               TIMESTAMPTZ NOT NULL DEFAULT now(),
    sale_type               VARCHAR(20) NOT NULL,
    nasabah_id              UUID,
    customer_name           VARCHAR,
    subtotal                BIGINT NOT NULL DEFAULT 0,
    discount_amount         BIGINT NOT NULL DEFAULT 0,
    total_amount            BIGINT NOT NULL DEFAULT 0,
    payment_method          VARCHAR(20) NOT NULL,
    cash_amount             BIGINT DEFAULT 0,
    tabungan_debit_amount   BIGINT DEFAULT 0,
    ewallet_debit_amount    BIGINT DEFAULT 0,
    change_amount           BIGINT DEFAULT 0,
    transaction_id          UUID,
    status                  VARCHAR(20) NOT NULL DEFAULT 'completed',
    voided_at               TIMESTAMPTZ,
    voided_by               UUID,
    void_reason             TEXT,

    _rels                   JSONB NOT NULL DEFAULT '{}',
    _data                   JSONB NOT NULL DEFAULT '{}',
    _sync_status            TEXT NOT NULL DEFAULT 'synced',
    _sync_version           BIGINT NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ,

    CONSTRAINT chk_tp_sale_type CHECK (sale_type IN ('store', 'canteen')),
    CONSTRAINT chk_tp_payment CHECK (payment_method IN ('cash', 'tabungan_debit', 'ewallet_debit', 'mixed')),
    CONSTRAINT chk_tp_status CHECK (status IN ('completed', 'voided')),
    CONSTRAINT uq_tp_receipt UNIQUE (tenant_id, company_id, receipt_number)
);

CREATE INDEX idx_tp_tenant_company ON toko_penjualan (tenant_id, company_id);
CREATE INDEX idx_tp_location ON toko_penjualan (location_id);
CREATE INDEX idx_tp_nasabah ON toko_penjualan (nasabah_id);
CREATE INDEX idx_tp_date ON toko_penjualan (sale_date);
CREATE INDEX idx_tp_status ON toko_penjualan (status);
CREATE INDEX idx_tp_rels ON toko_penjualan USING GIN (_rels);
CREATE INDEX idx_tp_data ON toko_penjualan USING GIN (_data);
