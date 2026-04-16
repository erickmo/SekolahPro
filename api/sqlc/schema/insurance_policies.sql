-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE insurance_policies (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID          NOT NULL,
    company_id          UUID          NOT NULL,
    pinjaman_id         UUID          NOT NULL,
    nasabah_id          UUID          NOT NULL,
    product_id          UUID          NOT NULL,
    policy_number       VARCHAR(50)   NOT NULL,
    coverage_amount     BIGINT        NOT NULL,
    premium_amount      BIGINT        NOT NULL,
    premium_type        VARCHAR(30)   NOT NULL,
    effective_date      DATE          NOT NULL,
    expiry_date         DATE          NOT NULL,
    status              VARCHAR(20)   NOT NULL DEFAULT 'active',
    cancellation_reason TEXT,
    issued_at           TIMESTAMPTZ   NOT NULL,
    issued_by           UUID          NOT NULL,
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_insurance_policies_tenant ON insurance_policies (tenant_id, company_id);
CREATE UNIQUE INDEX idx_insurance_policies_number ON insurance_policies (tenant_id, company_id, policy_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_insurance_policies_pinjaman ON insurance_policies (tenant_id, company_id, pinjaman_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_insurance_policies_nasabah ON insurance_policies (tenant_id, company_id, nasabah_id) WHERE deleted_at IS NULL;
