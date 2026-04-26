-- 067: toko_stok_movement — Pergerakan stok toko/kantin (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_stok_movement (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    produk_id       UUID NOT NULL,
    lokasi_id       UUID NOT NULL,
    movement_type   VARCHAR(20) NOT NULL,
    quantity        INT NOT NULL,
    previous_stock  INT NOT NULL DEFAULT 0,
    new_stock       INT NOT NULL DEFAULT 0,
    reference_type  VARCHAR(30),
    reference_id    UUID,
    notes           TEXT,
    movement_date   TIMESTAMPTZ NOT NULL DEFAULT now(),

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_tsm_type CHECK (movement_type IN ('in', 'out', 'adjustment', 'transfer_in', 'transfer_out', 'opname'))
);

CREATE INDEX idx_tsm_tenant_company ON toko_stok_movement (tenant_id, company_id);
CREATE INDEX idx_tsm_produk ON toko_stok_movement (produk_id);
CREATE INDEX idx_tsm_lokasi ON toko_stok_movement (lokasi_id);
CREATE INDEX idx_tsm_date ON toko_stok_movement (movement_date);
CREATE INDEX idx_tsm_rels ON toko_stok_movement USING GIN (_rels);
CREATE INDEX idx_tsm_data ON toko_stok_movement USING GIN (_data);
