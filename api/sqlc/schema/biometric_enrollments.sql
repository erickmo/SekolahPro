-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE biometric_enrollments (
    id                    UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID             NOT NULL,
    company_id            UUID             NOT NULL,

    -- Nasabah
    nasabah_id            UUID             NOT NULL,

    -- Enrollment Info
    biometric_type        VARCHAR(20)      NOT NULL,
    device_type           VARCHAR(30)      NOT NULL,
    device_id             VARCHAR(255),

    -- Template
    storage_type          VARCHAR(20)      NOT NULL,
    template_hash         VARCHAR(255)     NOT NULL,
    template_encrypted    BYTEA,
    encryption_key_ref    VARCHAR(255),

    -- Quality
    quality_score         INT              NOT NULL DEFAULT 0,
    enrollment_attempts   INT              NOT NULL DEFAULT 0,
    liveness_verified     BOOLEAN          NOT NULL DEFAULT FALSE,

    -- Status
    status                VARCHAR(20)      NOT NULL DEFAULT 'active',
    disabled_reason       TEXT,
    expires_at            TIMESTAMPTZ,

    -- Consent
    consent_id            UUID             NOT NULL,
    consent_given_at      TIMESTAMPTZ      NOT NULL,

    -- Audit
    enrolled_at           TIMESTAMPTZ      NOT NULL,
    enrolled_by           UUID             NOT NULL,
    created_at            TIMESTAMPTZ      NOT NULL DEFAULT NOW(),

    -- Soft delete
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX idx_biometric_enrollments_tenant_company ON biometric_enrollments (tenant_id, company_id);
CREATE INDEX idx_biometric_enrollments_nasabah       ON biometric_enrollments (tenant_id, company_id, nasabah_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_biometric_enrollments_type           ON biometric_enrollments (tenant_id, company_id, biometric_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_biometric_enrollments_status         ON biometric_enrollments (tenant_id, company_id, status) WHERE deleted_at IS NULL;
