-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE backup_configs (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID          NOT NULL,
    company_id      UUID          NOT NULL,
    config_type     VARCHAR(20)   NOT NULL,
    schedule        VARCHAR(100)  NOT NULL,
    retention_days  INT           NOT NULL DEFAULT 30,
    is_enabled      BOOLEAN       NOT NULL DEFAULT TRUE,
    last_run_at     TIMESTAMPTZ,
    next_run_at     TIMESTAMPTZ,
    status          VARCHAR(10)   NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_backup_configs_tenant_company ON backup_configs (tenant_id, company_id);
CREATE INDEX idx_backup_configs_type ON backup_configs (tenant_id, company_id, config_type) WHERE deleted_at IS NULL;
