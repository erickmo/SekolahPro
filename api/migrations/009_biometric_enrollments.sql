-- Biometric Enrollments table
-- Stores biometric enrollment data for nasabah authentication.

CREATE TABLE biometric_enrollments (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,
    nasabah_id      UUID NOT NULL,
    biometric_type  VARCHAR(20) NOT NULL CHECK (biometric_type IN ('fingerprint', 'face', 'voice')),
    device_info     TEXT NOT NULL DEFAULT '',
    template_hash   VARCHAR(255) NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    verified_at     TIMESTAMPTZ,
    verified_by     UUID,
    failed_attempts INT NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Indexes for multi-tenant scoped queries
CREATE INDEX idx_biometric_enrollments_tenant_company ON biometric_enrollments (tenant_id, company_id);
CREATE INDEX idx_biometric_enrollments_nasabah ON biometric_enrollments (tenant_id, company_id, nasabah_id);
CREATE INDEX idx_biometric_enrollments_type ON biometric_enrollments (tenant_id, company_id, biometric_type);
CREATE INDEX idx_biometric_enrollments_active ON biometric_enrollments (tenant_id, company_id, is_active) WHERE deleted_at IS NULL;

-- Biometric Verification Logs table
-- Stores audit trail of all biometric verification attempts.

CREATE TABLE biometric_verification_logs (
    id                  UUID PRIMARY KEY,
    tenant_id           UUID NOT NULL,
    company_id          UUID NOT NULL,
    enrollment_id       UUID NOT NULL REFERENCES biometric_enrollments(id),
    nasabah_id          UUID NOT NULL,
    verification_result VARCHAR(20) NOT NULL CHECK (verification_result IN ('success', 'failed', 'fallback')),
    fallback_method     VARCHAR(20) CHECK (fallback_method IN ('pin', 'password', 'manual')),
    device_info         TEXT NOT NULL DEFAULT '',
    ip_address          VARCHAR(45) NOT NULL DEFAULT '',
    attempted_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

-- Indexes for multi-tenant scoped queries
CREATE INDEX idx_biometric_logs_tenant_company ON biometric_verification_logs (tenant_id, company_id);
CREATE INDEX idx_biometric_logs_enrollment ON biometric_verification_logs (tenant_id, company_id, enrollment_id);
CREATE INDEX idx_biometric_logs_nasabah ON biometric_verification_logs (tenant_id, company_id, nasabah_id);
CREATE INDEX idx_biometric_logs_result ON biometric_verification_logs (tenant_id, company_id, verification_result);
