-- 025: lesson_plans
CREATE TABLE IF NOT EXISTS lesson_plans (
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

CREATE INDEX idx_lesson_plans_tenant ON lesson_plans(tenant_id, company_id);
CREATE INDEX idx_lesson_plans_teacher_id ON lesson_plans(tenant_id, company_id, (_data->>'teacher_id'));
CREATE INDEX idx_lesson_plans_subject_id ON lesson_plans(tenant_id, company_id, (_data->>'subject_id'));
