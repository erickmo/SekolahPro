-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE education_courses (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID          NOT NULL,
    company_id           UUID          NOT NULL,

    -- Course Info
    course_code          VARCHAR(50)   NOT NULL,
    title                VARCHAR(255)  NOT NULL,
    description          TEXT          NOT NULL DEFAULT '',
    program_type         VARCHAR(30)   NOT NULL,

    -- Content
    content_type         VARCHAR(20)   NOT NULL,
    content_url          VARCHAR(500),
    content_doc_id       UUID,
    duration_minutes     INT           NOT NULL DEFAULT 0,

    -- Prerequisites
    prerequisite_ids     UUID[]        DEFAULT '{}',
    mandatory            BOOLEAN       NOT NULL DEFAULT FALSE,

    -- Assessment
    has_quiz             BOOLEAN       NOT NULL DEFAULT FALSE,
    passing_score        INT,
    certificate_template VARCHAR(255),

    -- Targeting
    target_audience      VARCHAR(20)   NOT NULL DEFAULT 'all_members',
    applicable_mode      VARCHAR(20)   NOT NULL DEFAULT 'both',

    -- Status
    is_active            BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ
);

CREATE INDEX idx_education_courses_tenant_company ON education_courses (tenant_id, company_id);
CREATE INDEX idx_education_courses_program_type ON education_courses (tenant_id, company_id, program_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_education_courses_target_audience ON education_courses (tenant_id, company_id, target_audience) WHERE deleted_at IS NULL;
CREATE INDEX idx_education_courses_active ON education_courses (tenant_id, company_id, is_active) WHERE deleted_at IS NULL;
