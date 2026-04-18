-- 057: laboratory — Laboratories, Equipment, Usage Logs
-- Vernon pattern: _rels/_data JSONB columns.

--------------------------------------------------------------------------------
-- laboratories
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS laboratories (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Domain data
    name            VARCHAR(100) NOT NULL,
    type            VARCHAR(20) NOT NULL,
    capacity        INT,
    building        VARCHAR(50),
    floor           INT,
    room_number     VARCHAR(20),
    equipment_count INT NOT NULL DEFAULT 0,
    safety_rating   VARCHAR(10),
    last_inspection_date DATE,
    status          VARCHAR(15) NOT NULL DEFAULT 'active',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_lab_type CHECK (type IN (
        'ipa', 'komputer', 'bahasa', 'multimedia'
    )),
    CONSTRAINT chk_lab_safety_rating CHECK (safety_rating IS NULL OR safety_rating IN (
        'excellent', 'good', 'fair', 'poor'
    )),
    CONSTRAINT chk_lab_status CHECK (status IN (
        'active', 'inactive', 'maintenance'
    ))
);

CREATE INDEX idx_laboratories_tenant_company ON laboratories (tenant_id, company_id);
CREATE INDEX idx_laboratories_type ON laboratories (type);
CREATE INDEX idx_laboratories_status ON laboratories (status);
CREATE INDEX idx_laboratories_building ON laboratories (building) WHERE building IS NOT NULL;
CREATE INDEX idx_laboratories_rels ON laboratories USING GIN (_rels);
CREATE INDEX idx_laboratories_data ON laboratories USING GIN (_data);

--------------------------------------------------------------------------------
-- lab_equipment
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lab_equipment (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    lab_id          UUID NOT NULL,

    -- Domain data
    name            VARCHAR(200) NOT NULL,
    category        VARCHAR(30) NOT NULL,
    brand           VARCHAR(100),
    model           VARCHAR(100),
    serial_number   VARCHAR(100),
    condition       VARCHAR(15) NOT NULL DEFAULT 'good',
    quantity        INT NOT NULL DEFAULT 1,
    unit            VARCHAR(20) NOT NULL DEFAULT 'unit',
    purchase_date   DATE,
    purchase_price  NUMERIC(14,2),
    calibration_date DATE,
    next_calibration DATE,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_le_category CHECK (category IN (
        'optical', 'electronic', 'measuring', 'chemical',
        'biological', 'specimen', 'tool', 'safety',
        'furniture', 'computer', 'other'
    )),
    CONSTRAINT chk_le_condition CHECK (condition IN (
        'new', 'good', 'needs_repair', 'damaged'
    )),
    CONSTRAINT chk_le_quantity CHECK (quantity > 0)
);

CREATE INDEX idx_lab_equipment_tenant_company ON lab_equipment (tenant_id, company_id);
CREATE INDEX idx_lab_equipment_lab ON lab_equipment (lab_id);
CREATE INDEX idx_lab_equipment_category ON lab_equipment (category);
CREATE INDEX idx_lab_equipment_condition ON lab_equipment (condition);
CREATE INDEX idx_lab_equipment_next_calibration ON lab_equipment (next_calibration) WHERE next_calibration IS NOT NULL;
CREATE INDEX idx_lab_equipment_rels ON lab_equipment USING GIN (_rels);
CREATE INDEX idx_lab_equipment_data ON lab_equipment USING GIN (_data);

--------------------------------------------------------------------------------
-- lab_usage_logs
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lab_usage_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    lab_id          UUID NOT NULL,
    teacher_id      UUID NOT NULL,
    class_room_id   UUID,
    academic_year_id UUID NOT NULL,
    subject_id      UUID,

    -- Domain data
    usage_date      DATE NOT NULL,
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,
    topic           VARCHAR(255),
    participant_count INT,
    safety_checklist JSONB DEFAULT '{}',
    equipment_used  JSONB DEFAULT '[]',
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_lul_time_order CHECK (end_time > start_time)
);

CREATE INDEX idx_lab_usage_logs_tenant_company ON lab_usage_logs (tenant_id, company_id);
CREATE INDEX idx_lab_usage_logs_lab ON lab_usage_logs (lab_id);
CREATE INDEX idx_lul_teacher ON lab_usage_logs (teacher_id);
CREATE INDEX idx_lab_usage_logs_academic_year ON lab_usage_logs (academic_year_id);
CREATE INDEX idx_lab_usage_logs_subject ON lab_usage_logs (subject_id) WHERE subject_id IS NOT NULL;
CREATE INDEX idx_lul_date ON lab_usage_logs (usage_date);
CREATE INDEX idx_lab_usage_logs_rels ON lab_usage_logs USING GIN (_rels);
CREATE INDEX idx_lab_usage_logs_data ON lab_usage_logs USING GIN (_data);
