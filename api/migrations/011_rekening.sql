-- 011: rekening (cooperative account)
CREATE TABLE IF NOT EXISTS rekening (
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

CREATE INDEX idx_rekening_tenant ON rekening(tenant_id, company_id);
CREATE INDEX idx_rekening_no_rekening ON rekening(tenant_id, company_id, (_data->>'no_rekening'));
CREATE INDEX idx_rekening_nasabah_id ON rekening(tenant_id, company_id, (_data->>'nasabah_id'));
