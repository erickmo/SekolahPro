-- 069: toko_penjualan_item — Item penjualan POS (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_penjualan_item (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    penjualan_id    UUID NOT NULL,
    produk_id       UUID NOT NULL,
    quantity        INT NOT NULL,
    unit_price      BIGINT NOT NULL,
    subtotal        BIGINT NOT NULL,
    discount_amount BIGINT NOT NULL DEFAULT 0,
    total_amount    BIGINT NOT NULL,
    notes           TEXT,

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_tpi_tenant_company ON toko_penjualan_item (tenant_id, company_id);
CREATE INDEX idx_tpi_penjualan ON toko_penjualan_item (penjualan_id);
CREATE INDEX idx_tpi_produk ON toko_penjualan_item (produk_id);
CREATE INDEX idx_tpi_rels ON toko_penjualan_item USING GIN (_rels);
CREATE INDEX idx_tpi_data ON toko_penjualan_item USING GIN (_data);
