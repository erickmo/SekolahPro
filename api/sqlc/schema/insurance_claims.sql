-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE insurance_claims (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID          NOT NULL,
    company_id        UUID          NOT NULL,
    policy_id         UUID          NOT NULL,
    claim_number      VARCHAR(50)   NOT NULL,
    claim_type        VARCHAR(30)   NOT NULL,
    claim_date        DATE          NOT NULL,
    incident_date     DATE          NOT NULL,
    document_ids      UUID[]        NOT NULL DEFAULT '{}',
    description       TEXT          NOT NULL DEFAULT '',
    assessed_amount   BIGINT,
    approved_amount   BIGINT,
    assessor_id       UUID,
    assessment_notes  TEXT,
    status            VARCHAR(20)   NOT NULL DEFAULT 'submitted',
    settlement_date   DATE,
    settlement_type   VARCHAR(30),
    submitted_at      TIMESTAMPTZ   NOT NULL,
    resolved_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_insurance_claims_tenant ON insurance_claims (tenant_id, company_id);
CREATE UNIQUE INDEX idx_insurance_claims_number ON insurance_claims (tenant_id, company_id, claim_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_insurance_claims_policy ON insurance_claims (tenant_id, company_id, policy_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_insurance_claims_status ON insurance_claims (tenant_id, company_id, status) WHERE deleted_at IS NULL;
