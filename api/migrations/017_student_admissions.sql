-- 017: student_admissions
CREATE TABLE IF NOT EXISTS student_admissions (
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

CREATE INDEX idx_student_admissions_tenant ON student_admissions(tenant_id, company_id);
CREATE INDEX idx_student_admissions_student_id ON student_admissions(tenant_id, company_id, (_data->>'student_id'));
CREATE INDEX idx_student_admissions_registration_number ON student_admissions(tenant_id, company_id, (_data->>'registration_number'));
