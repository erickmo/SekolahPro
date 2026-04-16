-- Schema untuk Stress Test Scenarios.

CREATE TYPE scenario_type AS ENUM (
    'baseline', 'adverse', 'severely_adverse'
);

CREATE TABLE stress_test_scenarios (
    id                  UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID           NOT NULL,
    company_id          UUID           NOT NULL,
    scenario_name       VARCHAR(255)   NOT NULL,
    scenario_type       scenario_type  NOT NULL DEFAULT 'baseline',
    description         TEXT           NOT NULL DEFAULT '',
    parameters          JSONB          NOT NULL DEFAULT '{}',
    projected_car       DECIMAL(10,4),
    projected_npl       DECIMAL(10,4),
    projected_roa       DECIMAL(10,4),
    projected_roe       DECIMAL(10,4),
    capital_adequate    BOOLEAN,
    survives_scenario   BOOLEAN,
    test_date           DATE           NOT NULL,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_by          UUID           NOT NULL,
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_stress_tests_tenant ON stress_test_scenarios(tenant_id, company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_stress_tests_type ON stress_test_scenarios(tenant_id, company_id, scenario_type) WHERE deleted_at IS NULL;
