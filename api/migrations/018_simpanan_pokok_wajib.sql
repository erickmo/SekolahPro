-- 018: simpanan_pokok_wajib (mandatory/principal savings)
CREATE TABLE IF NOT EXISTS simpanan_pokok_wajib (
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

CREATE INDEX idx_simpanan_pokok_wajib_tenant ON simpanan_pokok_wajib(tenant_id, company_id);
CREATE INDEX idx_simpanan_pokok_wajib_nasabah_id ON simpanan_pokok_wajib(tenant_id, company_id, (_data->>'nasabah_id'));
