-- 060: teacher_evaluations — Competencies, Evaluations, Scores
-- Vernon pattern: _rels/_data JSONB columns.

--------------------------------------------------------------------------------
-- teacher_evaluation_competencies
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teacher_evaluation_competencies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Domain data
    name            VARCHAR(100) NOT NULL,
    area            VARCHAR(20) NOT NULL,
    indicator_count INT NOT NULL DEFAULT 0,
    weight          NUMERIC(5,2) NOT NULL DEFAULT 1.0,
    description     TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_tec_area CHECK (area IN (
        'pedagogic', 'personality', 'social', 'professional'
    )),
    CONSTRAINT chk_tec_weight CHECK (weight > 0),
    CONSTRAINT chk_tec_indicator_count CHECK (indicator_count >= 0)
);

CREATE INDEX idx_teacher_eval_competencies_tenant_company ON teacher_evaluation_competencies (tenant_id, company_id);
CREATE INDEX idx_teacher_eval_competencies_area ON teacher_evaluation_competencies (area);
CREATE INDEX idx_teacher_eval_competencies_active ON teacher_evaluation_competencies (is_active) WHERE is_active = true;
CREATE INDEX idx_teacher_eval_competencies_rels ON teacher_evaluation_competencies USING GIN (_rels);
CREATE INDEX idx_teacher_eval_competencies_data ON teacher_evaluation_competencies USING GIN (_data);

--------------------------------------------------------------------------------
-- teacher_evaluations
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teacher_evaluations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    evaluator_id    UUID NOT NULL,

    -- Domain data
    semester        VARCHAR(10) NOT NULL,
    total_score     NUMERIC(5,2),
    grade           VARCHAR(5),
    workflow_status VARCHAR(20) NOT NULL DEFAULT 'draft',
    evaluation_date DATE,
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_te_semester CHECK (semester IN (
        'ganjil', 'genap'
    )),
    CONSTRAINT chk_te_grade CHECK (grade IS NULL OR grade IN (
        'A', 'B', 'C', 'D', 'E'
    )),
    CONSTRAINT chk_te_workflow CHECK (workflow_status IN (
        'draft', 'self_assessment', 'peer_review',
        'supervisor_review', 'final', 'approved'
    )),
    CONSTRAINT chk_te_score CHECK (total_score IS NULL OR (total_score >= 0 AND total_score <= 100)),
    CONSTRAINT uq_te_unique UNIQUE (tenant_id, company_id, teacher_id, academic_year_id, semester)
);

CREATE INDEX idx_teacher_evaluations_tenant_company ON teacher_evaluations (tenant_id, company_id);
CREATE INDEX idx_teacher_evaluations_teacher ON teacher_evaluations (teacher_id);
CREATE INDEX idx_teacher_evaluations_academic_year ON teacher_evaluations (academic_year_id);
CREATE INDEX idx_teacher_evaluations_evaluator ON teacher_evaluations (evaluator_id);
CREATE INDEX idx_teacher_evaluations_workflow ON teacher_evaluations (workflow_status);
CREATE INDEX idx_teacher_evaluations_grade ON teacher_evaluations (grade) WHERE grade IS NOT NULL;
CREATE INDEX idx_teacher_evaluations_rels ON teacher_evaluations USING GIN (_rels);
CREATE INDEX idx_teacher_evaluations_data ON teacher_evaluations USING GIN (_data);

--------------------------------------------------------------------------------
-- teacher_evaluation_scores
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teacher_evaluation_scores (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    evaluation_id   UUID NOT NULL,
    competency_id   UUID NOT NULL,
    assessor_id     UUID NOT NULL,

    -- Domain data
    assessor_type   VARCHAR(15) NOT NULL,
    score           INT NOT NULL,
    evidence        TEXT,
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_tes_assessor_type CHECK (assessor_type IN (
        'self', 'peer', 'supervisor'
    )),
    CONSTRAINT chk_tes_score CHECK (score BETWEEN 1 AND 100)
);

CREATE INDEX idx_teacher_evaluation_scores_tenant_company ON teacher_evaluation_scores (tenant_id, company_id);
CREATE INDEX idx_tes_eval ON teacher_evaluation_scores (evaluation_id);
CREATE INDEX idx_tes_comp ON teacher_evaluation_scores (competency_id);
CREATE INDEX idx_teacher_evaluation_scores_assessor ON teacher_evaluation_scores (assessor_id);
CREATE INDEX idx_teacher_evaluation_scores_assessor_type ON teacher_evaluation_scores (assessor_type);
CREATE INDEX idx_teacher_evaluation_scores_rels ON teacher_evaluation_scores USING GIN (_rels);
CREATE INDEX idx_teacher_evaluation_scores_data ON teacher_evaluation_scores USING GIN (_data);
