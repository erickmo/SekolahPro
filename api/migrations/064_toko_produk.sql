-- 064: toko_produk — Katalog produk toko/kantin (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_produk (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id           UUID NOT NULL,
    company_id          UUID NOT NULL,

    name                VARCHAR NOT NULL,
    sku                 VARCHAR NOT NULL,
    barcode             VARCHAR,
    description         TEXT,
    category_id         UUID NOT NULL,
    cost_price          BIGINT NOT NULL DEFAULT 0,
    sell_price          BIGINT NOT NULL DEFAULT 0,
    margin_percentage   DECIMAL(5,2) DEFAULT 0,
    current_stock       INT NOT NULL DEFAULT 0,
    minimum_stock       INT NOT NULL DEFAULT 0,
    unit                VARCHAR(20),
    product_type        VARCHAR(20) NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    is_daily_menu       BOOLEAN NOT NULL DEFAULT false,
    image_url           VARCHAR,

    _rels               JSONB NOT NULL DEFAULT '{}',
    _data               JSONB NOT NULL DEFAULT '{}',
    _sync_status        TEXT NOT NULL DEFAULT 'synced',
    _sync_version       BIGINT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT chk_toko_produk_type CHECK (product_type IN ('store_item', 'canteen_item')),
    CONSTRAINT uq_toko_produk_sku UNIQUE (tenant_id, company_id, sku)
);

CREATE INDEX idx_toko_produk_tenant_company ON toko_produk (tenant_id, company_id);
CREATE INDEX idx_toko_produk_category ON toko_produk (category_id);
CREATE INDEX idx_toko_produk_active ON toko_produk (is_active) WHERE is_active = true;
CREATE INDEX idx_toko_produk_rels ON toko_produk USING GIN (_rels);
CREATE INDEX idx_toko_produk_data ON toko_produk USING GIN (_data);
