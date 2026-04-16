-- 028: angsuran (loan installments)
CREATE TABLE IF NOT EXISTS angsuran (
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

CREATE INDEX idx_angsuran_tenant ON angsuran(tenant_id, company_id);
CREATE INDEX idx_angsuran_pinjaman_id ON angsuran(tenant_id, company_id, (_data->>'pinjaman_id'));
CREATE INDEX idx_angsuran_status ON angsuran(tenant_id, company_id, (_data->>'status'));
