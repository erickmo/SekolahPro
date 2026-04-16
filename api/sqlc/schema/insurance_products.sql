-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE insurance_products (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID          NOT NULL,
    company_id            UUID          NOT NULL,
    product_code          VARCHAR(50)   NOT NULL,
    product_name          VARCHAR(255)  NOT NULL,
    product_type          VARCHAR(50)   NOT NULL,
    provider_id           UUID,
    coverage_type         VARCHAR(30)   NOT NULL,
    coverage_percentage   NUMERIC(5,2)  NOT NULL DEFAULT 100.00,
    max_coverage_amount   BIGINT,
    premium_type          VARCHAR(30)   NOT NULL,
    premium_rate          NUMERIC(7,5)  NOT NULL,
    premium_paid_by       VARCHAR(20)   NOT NULL,
    min_age               INT,
    max_age               INT,
    health_check_required BOOLEAN       NOT NULL DEFAULT FALSE,
    max_plafon            BIGINT,
    applicable_mode       VARCHAR(20)   NOT NULL DEFAULT 'both',
    is_active             BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX idx_insurance_products_tenant ON insurance_products (tenant_id, company_id);
CREATE UNIQUE INDEX idx_insurance_products_code ON insurance_products (tenant_id, company_id, product_code) WHERE deleted_at IS NULL;
