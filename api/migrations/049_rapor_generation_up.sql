-- Sprint 6: ADR-S018 — Rapor Generation
-- 2 tabel: rapor_templates, rapor_records

-- ============================================================================
-- 1. rapor_templates — Template layout rapor
-- ============================================================================

CREATE TABLE IF NOT EXISTS rapor_templates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

-- Expression indexes on _data FK fields (Vernon pattern)
CREATE INDEX IF NOT EXISTS idx_rapor_tpl_tenant_company ON rapor_templates (tenant_id, company_id);
CREATE INDEX IF NOT EXISTS idx_rapor_tpl_data ON rapor_templates USING GIN (_data);
CREATE INDEX IF NOT EXISTS idx_rapor_tpl_rels ON rapor_templates USING GIN (_rels);
CREATE INDEX IF NOT EXISTS idx_rapor_tpl_data_curriculum ON rapor_templates ((_data->>'curriculum_type')) WHERE _data->>'curriculum_type' IS NOT NULL;

-- ============================================================================
-- 2. rapor_records — Rapor per siswa per semester (snapshot)
-- ============================================================================

CREATE TABLE IF NOT EXISTS rapor_records (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

-- Expression indexes on _data FK fields (Vernon pattern — 4 BelongsTo relations)
CREATE INDEX IF NOT EXISTS idx_rapor_tenant_company ON rapor_records (tenant_id, company_id);
CREATE INDEX IF NOT EXISTS idx_rapor_data_student_id ON rapor_records ((_data->>'student_id')) WHERE _data->>'student_id' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_rapor_data_academic_year_id ON rapor_records ((_data->>'academic_year_id')) WHERE _data->>'academic_year_id' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_rapor_data_class_room_id ON rapor_records ((_data->>'class_room_id')) WHERE _data->>'class_room_id' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_rapor_data_template_id ON rapor_records ((_data->>'template_id')) WHERE _data->>'template_id' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_rapor_data_semester ON rapor_records ((_data->>'semester')) WHERE _data->>'semester' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_rapor_data_status ON rapor_records ((_data->>'status')) WHERE _data->>'status' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_rapor_rels ON rapor_records USING GIN (_rels);
CREATE INDEX IF NOT EXISTS idx_rapor_data ON rapor_records USING GIN (_data);
