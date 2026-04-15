-- +migrate Up
-- Domain Vernon: products
-- Menyimpan data produk. Autoload name+slug dari product_categories ke dalam _data["category"].
-- Dengan autoload, GET /products tidak perlu JOIN ke product_categories.

CREATE TABLE IF NOT EXISTS products (
    id            UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id     UUID        NOT NULL,
    company_id    UUID        NOT NULL,
    _rels         JSONB       NOT NULL DEFAULT '{}',
    _data         JSONB       NOT NULL DEFAULT '{}',
    _sync_status  TEXT        NOT NULL DEFAULT 'synced',
    _sync_version BIGINT      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

-- Index scope
CREATE INDEX IF NOT EXISTS idx_products_scope
    ON products (tenant_id, company_id, created_at DESC);

-- Index GIN pada _data
CREATE INDEX IF NOT EXISTS idx_products_data
    ON products USING GIN (_data)
    WHERE deleted_at IS NULL;

-- Index _sync_status untuk background sweep
CREATE INDEX IF NOT EXISTS idx_products_sync
    ON products (_sync_status)
    WHERE _sync_status != 'synced' AND deleted_at IS NULL;

-- Index tambahan: SKU adalah identifier bisnis utama, sering di-lookup
CREATE INDEX IF NOT EXISTS idx_products_sku
    ON products (tenant_id, company_id, ((_data->>'sku')))
    WHERE deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS products;
