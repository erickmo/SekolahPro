-- 027: pinjaman (loans)
CREATE TABLE IF NOT EXISTS pinjaman (
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

CREATE INDEX idx_pinjaman_tenant ON pinjaman(tenant_id, company_id);
CREATE INDEX idx_pinjaman_nasabah_id ON pinjaman(tenant_id, company_id, (_data->>'nasabah_id'));
CREATE INDEX idx_pinjaman_status ON pinjaman(tenant_id, company_id, (_data->>'status'));
