-- 026: teaching_journals
CREATE TABLE IF NOT EXISTS teaching_journals (
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

CREATE INDEX idx_teaching_journals_tenant ON teaching_journals(tenant_id, company_id);
CREATE INDEX idx_teaching_journals_teacher_id ON teaching_journals(tenant_id, company_id, (_data->>'teacher_id'));
CREATE INDEX idx_teaching_journals_date ON teaching_journals(tenant_id, company_id, (_data->>'date'));
