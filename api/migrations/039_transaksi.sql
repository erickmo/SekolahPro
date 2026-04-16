-- 039: transaksi (transactions)
CREATE TABLE IF NOT EXISTS transaksi (
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

CREATE INDEX idx_transaksi_tenant ON transaksi(tenant_id, company_id);
CREATE INDEX idx_transaksi_rekening_id ON transaksi(tenant_id, company_id, (_data->>'rekening_id'));
CREATE INDEX idx_transaksi_reference_no ON transaksi(tenant_id, company_id, (_data->>'reference_no'));
