-- 040: teller_sessions
CREATE TABLE IF NOT EXISTS teller_sessions (
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

CREATE INDEX idx_teller_sessions_tenant ON teller_sessions(tenant_id, company_id);
CREATE INDEX idx_teller_sessions_user_id ON teller_sessions(tenant_id, company_id, (_data->>'user_id'));
