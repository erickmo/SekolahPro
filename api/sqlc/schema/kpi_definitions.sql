-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.

CREATE TYPE kpi_category AS ENUM (
    'capital', 'asset_quality', 'management', 'earnings', 'liquidity', 'sensitivity'
);

CREATE TYPE kpi_unit AS ENUM (
    'percentage', 'ratio', 'amount', 'count'
);

CREATE TYPE kpi_direction AS ENUM (
    'higher_is_better', 'lower_is_better', 'target_range'
);

CREATE TYPE calculation_frequency AS ENUM (
    'daily', 'weekly', 'monthly', 'quarterly', 'yearly'
);

CREATE TABLE kpi_definitions (
    id                   UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID              NOT NULL,
    company_id           UUID              NOT NULL,
    kpi_code             VARCHAR(50)       NOT NULL,
    kpi_name             VARCHAR(255)      NOT NULL,
    category             kpi_category      NOT NULL,
    description          TEXT              NOT NULL DEFAULT '',
    formula              TEXT              NOT NULL DEFAULT '',
    unit                 kpi_unit          NOT NULL DEFAULT 'percentage',
    direction            kpi_direction     NOT NULL DEFAULT 'higher_is_better',
    healthy_min          DECIMAL(10,4),
    healthy_max          DECIMAL(10,4),
    warning_min          DECIMAL(10,4),
    warning_max          DECIMAL(10,4),
    critical_min         DECIMAL(10,4),
    critical_max         DECIMAL(10,4),
    calculation_frequency calculation_frequency NOT NULL DEFAULT 'monthly',
    data_sources         TEXT[]            NOT NULL DEFAULT '{}',
    is_active            BOOLEAN           NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ,

    UNIQUE(tenant_id, company_id, kpi_code)
);

CREATE INDEX idx_kpi_definitions_tenant ON kpi_definitions(tenant_id, company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_kpi_definitions_category ON kpi_definitions(tenant_id, company_id, category) WHERE deleted_at IS NULL;
