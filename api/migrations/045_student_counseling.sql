-- ADR-S017: Student Counseling / BK
-- Kasus konseling dan sesi konseling per kasus.
-- Vernon pattern: _rels/_data JSONB columns.

-- Kasus / masalah siswa
CREATE TABLE IF NOT EXISTS counseling_cases (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Detail kasus
    case_no         VARCHAR(20) NOT NULL,
    category        VARCHAR(20) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    severity        VARCHAR(10) NOT NULL DEFAULT 'low',

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'open',
    opened_date     DATE NOT NULL,
    closed_date     DATE,
    resolution      TEXT,

    -- Penanganan
    counselor_id    UUID NOT NULL,
    referred_to     TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_case_no UNIQUE (tenant_id, company_id, case_no),
    CONSTRAINT chk_case_category CHECK (category IN (
        'academic', 'social', 'personal', 'career', 'behavioral', 'family', 'other'
    )),
    CONSTRAINT chk_case_severity CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT chk_case_status CHECK (status IN ('open', 'in_progress', 'referred', 'resolved', 'closed'))
);

-- Indexes
CREATE INDEX idx_case_tenant_company ON counseling_cases (tenant_id, company_id);
CREATE INDEX idx_case_student ON counseling_cases (student_id);
CREATE INDEX idx_case_counselor ON counseling_cases (counselor_id);
CREATE INDEX idx_case_status ON counseling_cases (status) WHERE status NOT IN ('resolved', 'closed');
CREATE INDEX idx_case_rels ON counseling_cases USING GIN (_rels);
CREATE INDEX idx_case_data ON counseling_cases USING GIN (_data);

-- Sesi konseling per kasus
CREATE TABLE IF NOT EXISTS counseling_sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    case_id         UUID NOT NULL,
    student_id      UUID NOT NULL,

    -- Detail sesi
    session_date    DATE NOT NULL,
    session_type    VARCHAR(20) NOT NULL,
    duration_minutes INT,
    notes           TEXT NOT NULL,
    recommendation  TEXT,

    -- Peserta
    counselor_id    UUID NOT NULL,
    parent_present  BOOLEAN NOT NULL DEFAULT false,

    -- Follow-up
    follow_up_date  DATE,
    follow_up_note  TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_session_type CHECK (session_type IN (
        'individual', 'group', 'home_visit', 'parent_conference', 'referral'
    ))
);

-- Indexes
CREATE INDEX idx_session_tenant_company ON counseling_sessions (tenant_id, company_id);
CREATE INDEX idx_session_case ON counseling_sessions (case_id);
CREATE INDEX idx_session_student ON counseling_sessions (student_id);
CREATE INDEX idx_session_date ON counseling_sessions (session_date);
CREATE INDEX idx_session_rels ON counseling_sessions USING GIN (_rels);
CREATE INDEX idx_session_data ON counseling_sessions USING GIN (_data);
