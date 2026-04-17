-- 051: Teacher Substitution (ADR-S032)
-- duty_schedules, teacher_substitutions, substitution_logs

-- ============================================================
-- duty_schedules: jadwal piket harian (rotasi per semester)
-- ============================================================
CREATE TABLE IF NOT EXISTS duty_schedules (
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

CREATE INDEX idx_duty_tenant_company ON duty_schedules (tenant_id, company_id);
CREATE INDEX idx_duty_teacher ON duty_schedules (tenant_id, company_id, (_data->>'teacher_id'));
CREATE INDEX idx_duty_year_semester ON duty_schedules (tenant_id, company_id, (_data->>'academic_year_id'), (_data->>'semester'));
CREATE INDEX idx_duty_day ON duty_schedules (tenant_id, company_id, (_data->>'day_of_week'));
CREATE INDEX idx_duty_type ON duty_schedules (tenant_id, company_id, (_data->>'duty_type'));
CREATE INDEX idx_duty_rels ON duty_schedules USING GIN (_rels);
CREATE INDEX idx_duty_data ON duty_schedules USING GIN (_data);

-- ============================================================
-- teacher_substitutions: penggantian guru mengajar
-- ============================================================
CREATE TABLE IF NOT EXISTS teacher_substitutions (
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

CREATE INDEX idx_sub_tenant_company ON teacher_substitutions (tenant_id, company_id);
CREATE INDEX idx_sub_original_teacher ON teacher_substitutions (tenant_id, company_id, (_data->>'original_teacher_id'));
CREATE INDEX idx_sub_substitute_teacher ON teacher_substitutions (tenant_id, company_id, (_data->>'substitute_teacher_id'));
CREATE INDEX idx_sub_date ON teacher_substitutions (tenant_id, company_id, (_data->>'substitution_date'));
CREATE INDEX idx_sub_status ON teacher_substitutions (tenant_id, company_id, (_data->>'status'));
CREATE INDEX idx_sub_class ON teacher_substitutions (tenant_id, company_id, (_data->>'class_room_id'));
CREATE INDEX idx_sub_year ON teacher_substitutions (tenant_id, company_id, (_data->>'academic_year_id'));
CREATE INDEX idx_sub_rels ON teacher_substitutions USING GIN (_rels);
CREATE INDEX idx_sub_data ON teacher_substitutions USING GIN (_data);

-- ============================================================
-- substitution_logs: log aktivitas substitusi (audit trail)
-- ============================================================
CREATE TABLE IF NOT EXISTS substitution_logs (
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

CREATE INDEX idx_sub_log_tenant_company ON substitution_logs (tenant_id, company_id);
CREATE INDEX idx_sub_log_substitution ON substitution_logs (tenant_id, company_id, (_data->>'substitution_id'));
CREATE INDEX idx_sub_log_actor ON substitution_logs (tenant_id, company_id, (_data->>'actor_id'));
CREATE INDEX idx_sub_log_action ON substitution_logs (tenant_id, company_id, (_data->>'action'));
CREATE INDEX idx_sub_log_rels ON substitution_logs USING GIN (_rels);
CREATE INDEX idx_sub_log_data ON substitution_logs USING GIN (_data);
