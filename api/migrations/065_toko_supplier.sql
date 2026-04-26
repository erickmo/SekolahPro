-- 065: toko_supplier — Supplier toko/kantin (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_supplier (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    name            VARCHAR NOT NULL,
    contact_person  VARCHAR,
    phone           VARCHAR(30),
    email           VARCHAR(100),
    address         TEXT,
    city            VARCHAR(50),
    bank_name       VARCHAR(50),
    bank_account    VARCHAR(50),
    bank_holder     VARCHAR(100),
    is_active       BOOLEAN NOT NULL DEFAULT true,
    notes           TEXT,

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_toko_supplier_tenant_company ON toko_supplier (tenant_id, company_id);
CREATE INDEX idx_toko_supplier_active ON toko_supplier (is_active) WHERE is_active = true;
CREATE INDEX idx_toko_supplier_rels ON toko_supplier USING GIN (_rels);
CREATE INDEX idx_toko_supplier_data ON toko_supplier USING GIN (_data);
