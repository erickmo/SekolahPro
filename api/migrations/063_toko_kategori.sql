-- 063: toko_kategori — Kategori produk toko/kantin (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_kategori (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    name            VARCHAR NOT NULL,
    parent_id       UUID,
    product_type    VARCHAR(20) NOT NULL,
    sort_order      INT NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_toko_kategori_type CHECK (product_type IN ('store_item', 'canteen_item'))
);

CREATE INDEX idx_toko_kategori_tenant_company ON toko_kategori (tenant_id, company_id);
CREATE INDEX idx_toko_kategori_parent ON toko_kategori (parent_id);
CREATE INDEX idx_toko_kategori_rels ON toko_kategori USING GIN (_rels);
CREATE INDEX idx_toko_kategori_data ON toko_kategori USING GIN (_data);
