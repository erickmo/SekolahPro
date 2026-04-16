-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE reserve_funds (
    id                      UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID             NOT NULL,
    company_id              UUID             NOT NULL,

    -- Fund Info
    fund_type               VARCHAR(20)      NOT NULL,
    fund_name               VARCHAR(255)     NOT NULL,
    description             TEXT             NOT NULL DEFAULT '',

    -- Balance
    current_balance         BIGINT           NOT NULL DEFAULT 0,
    target_balance          BIGINT,
    min_balance             BIGINT,

    -- Configuration
    shu_allocation_pct      DECIMAL(5,2)     NOT NULL DEFAULT 0,
    max_balance_pct         DECIMAL(5,2),
    auto_allocate           BOOLEAN          NOT NULL DEFAULT TRUE,

    -- Investment
    investment_instrument   VARCHAR(20),
    investment_maturity     DATE,
    investment_rate         DECIMAL(5,4),

    -- Status
    status                  VARCHAR(20)      NOT NULL DEFAULT 'active',
    frozen_reason           TEXT,

    -- Audit
    created_at              TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_by              UUID             NOT NULL,
    updated_by              UUID             NOT NULL,
    deleted_at              TIMESTAMPTZ
);

CREATE INDEX idx_reserve_funds_tenant_company ON reserve_funds (tenant_id, company_id);
CREATE INDEX idx_reserve_funds_type           ON reserve_funds (tenant_id, company_id, fund_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_reserve_funds_status         ON reserve_funds (tenant_id, company_id, status)   WHERE deleted_at IS NULL;
