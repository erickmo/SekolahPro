-- 010: produk_akad (cooperative product/contract type)
CREATE TABLE IF NOT EXISTS produk_akad (
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

CREATE INDEX idx_produk_akad_tenant ON produk_akad(tenant_id, company_id);
CREATE UNIQUE INDEX idx_produk_akad_code ON produk_akad(tenant_id, company_id, (_data->>'code')) WHERE deleted_at IS NULL;
