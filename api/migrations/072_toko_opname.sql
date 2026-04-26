-- 072: toko_opname — Stock opname toko/kantin (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_opname (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    lokasi_id       UUID NOT NULL,
    opname_date     DATE NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    total_items     INT NOT NULL DEFAULT 0,
    matched_items   INT NOT NULL DEFAULT 0,
    mismatch_items  INT NOT NULL DEFAULT 0,
    notes           TEXT,

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_to_status CHECK (status IN ('draft', 'in_progress', 'completed', 'cancelled'))
);

CREATE INDEX idx_to_tenant_company ON toko_opname (tenant_id, company_id);
CREATE INDEX idx_to_lokasi ON toko_opname (lokasi_id);
CREATE INDEX idx_to_status ON toko_opname (status);
CREATE INDEX idx_to_date ON toko_opname (opname_date);
CREATE INDEX idx_to_rels ON toko_opname USING GIN (_rels);
CREATE INDEX idx_to_data ON toko_opname USING GIN (_data);
