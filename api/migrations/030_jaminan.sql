-- 030: jaminan (collateral)
CREATE TABLE IF NOT EXISTS jaminan (
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

CREATE INDEX idx_jaminan_tenant ON jaminan(tenant_id, company_id);
CREATE INDEX idx_jaminan_pinjaman_id ON jaminan(tenant_id, company_id, (_data->>'pinjaman_id'));
