-- 024: teaching_schedules
CREATE TABLE IF NOT EXISTS teaching_schedules (
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

CREATE INDEX idx_teaching_schedules_tenant ON teaching_schedules(tenant_id, company_id);
CREATE INDEX idx_teaching_schedules_teacher_id ON teaching_schedules(tenant_id, company_id, (_data->>'teacher_id'));
CREATE INDEX idx_teaching_schedules_class_room_id ON teaching_schedules(tenant_id, company_id, (_data->>'class_room_id'));
