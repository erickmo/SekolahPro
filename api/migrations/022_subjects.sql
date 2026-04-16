-- 022: subjects
CREATE TABLE IF NOT EXISTS subjects (
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

CREATE INDEX idx_subjects_tenant ON subjects(tenant_id, company_id);
CREATE INDEX idx_subjects_curriculum_id ON subjects(tenant_id, company_id, (_data->>'curriculum_id'));
CREATE UNIQUE INDEX idx_subjects_code ON subjects(tenant_id, company_id, (_data->>'code')) WHERE deleted_at IS NULL;
