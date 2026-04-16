-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE biometric_verification_logs (
    id                    UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID             NOT NULL,
    company_id            UUID             NOT NULL,

    -- Nasabah & Enrollment
    nasabah_id            UUID             NOT NULL,
    enrollment_id         UUID             NOT NULL,

    -- Verification
    verification_type     VARCHAR(20)      NOT NULL,
    purpose               VARCHAR(40)      NOT NULL,
    related_entity_type   VARCHAR(50),
    related_entity_id     UUID,

    -- Result
    result                VARCHAR(20)      NOT NULL,
    confidence_score      DECIMAL(5,2)     NOT NULL DEFAULT 0,
    match_threshold       DECIMAL(5,2)     NOT NULL DEFAULT 0,
    failure_reason        TEXT,

    -- Audit
    verified_at           TIMESTAMPTZ      NOT NULL,
    ip_address            VARCHAR(45),
    device_info           VARCHAR(500),
    created_at            TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_biometric_vlogs_tenant_company ON biometric_verification_logs (tenant_id, company_id);
CREATE INDEX idx_biometric_vlogs_nasabah       ON biometric_verification_logs (tenant_id, company_id, nasabah_id);
CREATE INDEX idx_biometric_vlogs_result         ON biometric_verification_logs (tenant_id, company_id, result);
CREATE INDEX idx_biometric_vlogs_purpose        ON biometric_verification_logs (tenant_id, company_id, purpose);
CREATE INDEX idx_biometric_vlogs_enrollment     ON biometric_verification_logs (tenant_id, company_id, enrollment_id);
