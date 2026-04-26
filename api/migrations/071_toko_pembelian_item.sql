-- 071: toko_pembelian_item — Item purchase order (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_pembelian_item (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    pembelian_id    UUID NOT NULL,
    produk_id       UUID NOT NULL,
    quantity        INT NOT NULL,
    unit_price      BIGINT NOT NULL,
    subtotal        BIGINT NOT NULL,
    received_qty    INT NOT NULL DEFAULT 0,
    notes           TEXT,

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_tpbi_tenant_company ON toko_pembelian_item (tenant_id, company_id);
CREATE INDEX idx_tpbi_pembelian ON toko_pembelian_item (pembelian_id);
CREATE INDEX idx_tpbi_produk ON toko_pembelian_item (produk_id);
CREATE INDEX idx_tpbi_rels ON toko_pembelian_item USING GIN (_rels);
CREATE INDEX idx_tpbi_data ON toko_pembelian_item USING GIN (_data);
