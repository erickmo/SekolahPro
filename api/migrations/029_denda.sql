-- 029: denda (penalties/fines)
CREATE TABLE IF NOT EXISTS denda (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    _rels JSONB NOT NULL DEFAULT '{}',
    _data JSONB NOT NULL DEFAULT '{}',
    _sync_status TEXT NOT NULL DEFAULT 'synced',
    _sync_version BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_denda_tenant ON denda(tenant_id, company_id);
CREATE INDEX idx_denda_pinjaman_id ON denda(tenant_id, company_id, (_data->>'pinjaman_id'));
