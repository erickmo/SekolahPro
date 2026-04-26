-- 073: toko_opname_item — Item stock opname (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_opname_item (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    opname_id       UUID NOT NULL,
    produk_id       UUID NOT NULL,
    system_stock    INT NOT NULL DEFAULT 0,
    physical_stock  INT NOT NULL DEFAULT 0,
    difference      INT NOT NULL DEFAULT 0,
    notes           TEXT,

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_toi_tenant_company ON toko_opname_item (tenant_id, company_id);
CREATE INDEX idx_toi_opname ON toko_opname_item (opname_id);
CREATE INDEX idx_toi_produk ON toko_opname_item (produk_id);
CREATE INDEX idx_toi_rels ON toko_opname_item USING GIN (_rels);
CREATE INDEX idx_toi_data ON toko_opname_item USING GIN (_data);
