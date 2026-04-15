-- +migrate Up
-- Domain Vernon: product_categories
-- Tabel referensi untuk kategori produk. Digunakan sebagai sumber autoload oleh products.

CREATE TABLE IF NOT EXISTS product_categories (
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

-- Index scope: filter utama semua query Vernon (tenant_id, company_id)
CREATE INDEX IF NOT EXISTS idx_product_categories_scope
    ON product_categories (tenant_id, company_id, created_at DESC);

-- Index GIN pada _data untuk filter field arbitrer (name, slug, is_active, dll)
CREATE INDEX IF NOT EXISTS idx_product_categories_data
    ON product_categories USING GIN (_data)
    WHERE deleted_at IS NULL;

-- Index _sync_status untuk background sweep worker
CREATE INDEX IF NOT EXISTS idx_product_categories_sync
    ON product_categories (_sync_status)
    WHERE _sync_status != 'synced' AND deleted_at IS NULL;

-- Index tambahan: slug sering dipakai sebagai query param URL (e.g. ?slug=electronics)
CREATE INDEX IF NOT EXISTS idx_product_categories_slug
    ON product_categories (tenant_id, company_id, ((_data->>'slug')))
    WHERE deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS product_categories;
