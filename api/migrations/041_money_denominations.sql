-- 041: money_denominations
CREATE TABLE IF NOT EXISTS money_denominations (
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

CREATE INDEX idx_money_denominations_tenant ON money_denominations(tenant_id, company_id);
CREATE INDEX idx_money_denominations_teller_session_id ON money_denominations(tenant_id, company_id, (_data->>'teller_session_id'));
