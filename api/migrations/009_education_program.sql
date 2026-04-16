-- Education Courses
CREATE TABLE education_courses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    course_type VARCHAR(50) NOT NULL DEFAULT 'elective'
        CHECK (course_type IN ('mandatory', 'elective', 'certification')),
    category VARCHAR(255) NOT NULL DEFAULT '',
    duration_hours INT NOT NULL DEFAULT 0,
    content_url VARCHAR(500) NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    passing_score DECIMAL(5,2) NOT NULL DEFAULT 70.00
        CHECK (passing_score >= 0 AND passing_score <= 100),
    max_attempts INT NOT NULL DEFAULT 3
        CHECK (max_attempts > 0),
    mandatory_for TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_education_courses_tenant_company
    ON education_courses (tenant_id, company_id);
CREATE INDEX idx_education_courses_course_type
    ON education_courses (tenant_id, company_id, course_type)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_education_courses_category
    ON education_courses (tenant_id, company_id, category)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_education_courses_is_active
    ON education_courses (tenant_id, company_id, is_active)
    WHERE deleted_at IS NULL;

-- Education Enrollments
CREATE TABLE education_enrollments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    course_id UUID NOT NULL REFERENCES education_courses(id),
    nasabah_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'enrolled'
        CHECK (status IN ('enrolled', 'in_progress', 'completed', 'failed', 'expired')),
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    score DECIMAL(5,2) NOT NULL DEFAULT 0.00
        CHECK (score >= 0 AND score <= 100),
    attempts INT NOT NULL DEFAULT 1,
    certificate_url VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_education_enrollments_tenant_company
    ON education_enrollments (tenant_id, company_id);
CREATE INDEX idx_education_enrollments_course
    ON education_enrollments (tenant_id, company_id, course_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_education_enrollments_nasabah
    ON education_enrollments (tenant_id, company_id, nasabah_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_education_enrollments_status
    ON education_enrollments (tenant_id, company_id, status)
    WHERE deleted_at IS NULL;

-- Education KPIs
CREATE TABLE education_kpis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    metric_type VARCHAR(50) NOT NULL
        CHECK (metric_type IN ('completion_rate', 'avg_score', 'enrollment_count', 'pass_rate')),
    period_month VARCHAR(7) NOT NULL,
    value DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_education_kpis_tenant_company
    ON education_kpis (tenant_id, company_id);
CREATE INDEX idx_education_kpis_metric_type
    ON education_kpis (tenant_id, company_id, metric_type)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_education_kpis_period
    ON education_kpis (tenant_id, company_id, period_month)
    WHERE deleted_at IS NULL;
