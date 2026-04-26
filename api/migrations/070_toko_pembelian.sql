-- 070: toko_pembelian — Purchase order toko/kantin (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_pembelian (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    supplier_id     UUID NOT NULL,
    po_number       VARCHAR NOT NULL,
    order_date      DATE NOT NULL,
    expected_date   DATE,
    subtotal        BIGINT NOT NULL DEFAULT 0,
    discount_amount BIGINT NOT NULL DEFAULT 0,
    total_amount    BIGINT NOT NULL DEFAULT 0,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    received_date   DATE,
    notes           TEXT,

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_tb_status CHECK (status IN ('draft', 'sent', 'partial', 'received', 'cancelled')),
    CONSTRAINT uq_tb_po_number UNIQUE (tenant_id, company_id, po_number)
);

CREATE INDEX idx_tb_tenant_company ON toko_pembelian (tenant_id, company_id);
CREATE INDEX idx_tb_supplier ON toko_pembelian (supplier_id);
CREATE INDEX idx_tb_status ON toko_pembelian (status);
CREATE INDEX idx_tb_date ON toko_pembelian (order_date);
CREATE INDEX idx_tb_rels ON toko_pembelian USING GIN (_rels);
CREATE INDEX idx_tb_data ON toko_pembelian USING GIN (_data);
