-- 009: nasabah (cooperative customer/depositor)
CREATE TABLE IF NOT EXISTS nasabah (
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

CREATE INDEX idx_nasabah_tenant ON nasabah(tenant_id, company_id);
CREATE UNIQUE INDEX idx_nasabah_nik ON nasabah(tenant_id, company_id, (_data->>'nik')) WHERE deleted_at IS NULL;
CREATE INDEX idx_nasabah_no_nasabah ON nasabah(tenant_id, company_id, (_data->>'no_nasabah'));
