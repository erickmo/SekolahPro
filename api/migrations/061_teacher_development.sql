-- 061: teacher_development — Certifications, Activities, Credit Summaries
-- Vernon pattern: _rels/_data JSONB columns.

--------------------------------------------------------------------------------
-- teacher_certifications
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teacher_certifications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,

    -- Domain data
    certification_type VARCHAR(20) NOT NULL,
    certification_number VARCHAR(50),
    issue_date      DATE,
    expiry_date     DATE,
    issuing_body    VARCHAR(200),
    status          VARCHAR(15) NOT NULL DEFAULT 'active',
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_tc_type CHECK (certification_type IN (
        'profesi', 'penilaian', 'pengawas', 'teknisi'
    )),
    CONSTRAINT chk_tc_status CHECK (status IN (
        'active', 'expired', 'revoked', 'pending_renewal'
    )),
    CONSTRAINT chk_tc_dates CHECK (expiry_date IS NULL OR issue_date IS NULL OR expiry_date >= issue_date)
);

CREATE INDEX idx_teacher_certifications_tenant_company ON teacher_certifications (tenant_id, company_id);
CREATE INDEX idx_teacher_certifications_teacher ON teacher_certifications (teacher_id);
CREATE INDEX idx_teacher_certifications_type ON teacher_certifications (certification_type);
CREATE INDEX idx_teacher_certifications_status ON teacher_certifications (status);
CREATE INDEX idx_teacher_certifications_expiry ON teacher_certifications (expiry_date) WHERE status = 'active';
CREATE INDEX idx_teacher_certifications_rels ON teacher_certifications USING GIN (_rels);
CREATE INDEX idx_teacher_certifications_data ON teacher_certifications USING GIN (_data);

--------------------------------------------------------------------------------
-- teacher_development_activities
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teacher_development_activities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,

    -- Domain data
    activity_type   VARCHAR(30) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    organizer       VARCHAR(200),
    start_date      DATE NOT NULL,
    end_date        DATE,
    location        VARCHAR(200),
    credit_points   NUMERIC(5,2) NOT NULL DEFAULT 0,
    certificate_number VARCHAR(50),
    status          VARCHAR(15) NOT NULL DEFAULT 'registered',
    description     TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_tda_status CHECK (status IN (
        'registered', 'attended', 'completed', 'cancelled'
    )),
    CONSTRAINT chk_tda_credit_points CHECK (credit_points >= 0),
    CONSTRAINT chk_tda_dates CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_teacher_dev_activities_tenant_company ON teacher_development_activities (tenant_id, company_id);
CREATE INDEX idx_tda_teacher ON teacher_development_activities (teacher_id);
CREATE INDEX idx_tda_dates ON teacher_development_activities (start_date);
CREATE INDEX idx_teacher_dev_activities_status ON teacher_development_activities (status);
CREATE INDEX idx_teacher_dev_activities_type ON teacher_development_activities (activity_type);
CREATE INDEX idx_teacher_dev_activities_rels ON teacher_development_activities USING GIN (_rels);
CREATE INDEX idx_teacher_dev_activities_data ON teacher_development_activities USING GIN (_data);

--------------------------------------------------------------------------------
-- teacher_credit_summaries
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teacher_credit_summaries (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,

    -- Domain data
    current_rank    VARCHAR(10) NOT NULL,
    target_rank     VARCHAR(10),
    total_credits   NUMERIC(7,2) NOT NULL DEFAULT 0,
    required_credits NUMERIC(7,2) NOT NULL DEFAULT 0,
    credit_gap      NUMERIC(7,2),
    last_promotion_date DATE,
    next_eligible_date DATE,
    skp_score       NUMERIC(5,2),
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_tcs_current_rank CHECK (current_rank IN (
        'I/a', 'I/b', 'II/a', 'II/b', 'III/a', 'III/b', 'IV/a', 'IV/b', 'IV/c'
    )),
    CONSTRAINT chk_tcs_target_rank CHECK (target_rank IS NULL OR target_rank IN (
        'I/a', 'I/b', 'II/a', 'II/b', 'III/a', 'III/b', 'IV/a', 'IV/b', 'IV/c'
    )),
    CONSTRAINT chk_tcs_credits CHECK (total_credits >= 0 AND required_credits >= 0),
    CONSTRAINT chk_tcs_skp CHECK (skp_score IS NULL OR (skp_score >= 0 AND skp_score <= 100)),
    CONSTRAINT uq_tcs_teacher UNIQUE (tenant_id, company_id, teacher_id)
);

CREATE INDEX idx_teacher_credit_summaries_tenant_company ON teacher_credit_summaries (tenant_id, company_id);
CREATE INDEX idx_teacher_credit_summaries_teacher ON teacher_credit_summaries (teacher_id);
CREATE INDEX idx_teacher_credit_summaries_rank ON teacher_credit_summaries (current_rank);
CREATE INDEX idx_teacher_credit_summaries_next_eligible ON teacher_credit_summaries (next_eligible_date) WHERE next_eligible_date IS NOT NULL;
CREATE INDEX idx_teacher_credit_summaries_rels ON teacher_credit_summaries USING GIN (_rels);
CREATE INDEX idx_teacher_credit_summaries_data ON teacher_credit_summaries USING GIN (_data);
