-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE education_enrollments (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID          NOT NULL,
    company_id            UUID          NOT NULL,
    course_id             UUID          NOT NULL REFERENCES education_courses(id),
    nasabah_id            UUID          NOT NULL,

    -- Progress
    status                VARCHAR(20)   NOT NULL DEFAULT 'enrolled',
    progress_pct          INT           NOT NULL DEFAULT 0,
    started_at            TIMESTAMPTZ,
    completed_at          TIMESTAMPTZ,

    -- Quiz Result
    quiz_score            INT,
    quiz_attempts         INT           NOT NULL DEFAULT 0,
    passed                BOOLEAN,

    -- Certificate
    certificate_id        UUID,
    certificate_issued_at TIMESTAMPTZ,

    -- Feedback
    rating                INT,
    feedback              TEXT,

    -- Audit
    enrolled_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deadline_at           TIMESTAMPTZ,
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX idx_education_enrollments_tenant_company ON education_enrollments (tenant_id, company_id);
CREATE INDEX idx_education_enrollments_course ON education_enrollments (tenant_id, company_id, course_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_education_enrollments_nasabah ON education_enrollments (tenant_id, company_id, nasabah_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_education_enrollments_status ON education_enrollments (tenant_id, company_id, status) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_education_enrollments_unique ON education_enrollments (tenant_id, company_id, course_id, nasabah_id) WHERE deleted_at IS NULL;
