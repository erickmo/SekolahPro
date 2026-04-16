-- Schema untuk KPI Measurements.

CREATE TYPE health_status AS ENUM (
    'healthy', 'warning', 'critical'
);

CREATE TYPE kpi_trend AS ENUM (
    'improving', 'stable', 'deteriorating'
);

CREATE TABLE kpi_measurements (
    id                  UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID           NOT NULL,
    company_id          UUID           NOT NULL,
    kpi_definition_id   UUID           NOT NULL REFERENCES kpi_definitions(id),
    measurement_date    DATE           NOT NULL,
    measurement_period  VARCHAR(20)    NOT NULL,
    value               DECIMAL(20,4)  NOT NULL,
    previous_value      DECIMAL(20,4),
    change_pct          DECIMAL(10,4),
    health_status       health_status  NOT NULL DEFAULT 'healthy',
    trend               kpi_trend      NOT NULL DEFAULT 'stable',
    alert_triggered     BOOLEAN        NOT NULL DEFAULT FALSE,
    calculated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, company_id, kpi_definition_id, measurement_date)
);

CREATE INDEX idx_kpi_measurements_tenant ON kpi_measurements(tenant_id, company_id) ;
CREATE INDEX idx_kpi_measurements_kpi ON kpi_measurements(tenant_id, company_id, kpi_definition_id);
CREATE INDEX idx_kpi_measurements_status ON kpi_measurements(tenant_id, company_id, health_status);
CREATE INDEX idx_kpi_measurements_date ON kpi_measurements(tenant_id, company_id, measurement_date DESC);
