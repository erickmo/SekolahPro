-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE reserve_fund_transactions (
    id                    UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID             NOT NULL,
    company_id            UUID             NOT NULL,
    fund_id               UUID             NOT NULL REFERENCES reserve_funds(id),

    -- Transaction Info
    transaction_type      VARCHAR(20)      NOT NULL,
    amount                BIGINT           NOT NULL,
    balance_after         BIGINT           NOT NULL,

    -- Reference
    shu_id                UUID,
    jurnal_id             UUID,
    reference_type        VARCHAR(50),
    reference_id          UUID,

    -- Approval
    requires_approval     BOOLEAN          NOT NULL DEFAULT TRUE,
    approved_by           UUID,
    approved_at           TIMESTAMPTZ,
    rejection_reason      TEXT,

    -- Audit
    transaction_date      DATE             NOT NULL,
    created_at            TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_by            UUID             NOT NULL,
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX idx_reserve_fund_txns_tenant_company ON reserve_fund_transactions (tenant_id, company_id);
CREATE INDEX idx_reserve_fund_txns_fund           ON reserve_fund_transactions (tenant_id, company_id, fund_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reserve_fund_txns_type            ON reserve_fund_transactions (tenant_id, company_id, transaction_type) WHERE deleted_at IS NULL;
