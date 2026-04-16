-- Schema untuk sqlc code generation: AML Monitoring Rules (ADR-K027).
-- tenant_id dan company_id wajib ada di semua tabel.

CREATE TYPE aml_rule_type AS ENUM ('threshold', 'pattern', 'velocity', 'anomaly');
CREATE TYPE aml_rule_applies_to AS ENUM ('all', 'high_risk_only', 'pep_only');

CREATE TABLE aml_monitoring_rules (
    id          UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID                NOT NULL,
    company_id  UUID                NOT NULL,
    rule_code   VARCHAR(50)         NOT NULL,
    rule_name   VARCHAR(255)        NOT NULL,
    rule_type   aml_rule_type       NOT NULL,
    description TEXT                NOT NULL DEFAULT '',
    parameters  JSONB               NOT NULL DEFAULT '{}',
    is_active   BOOLEAN             NOT NULL DEFAULT TRUE,
    applies_to  aml_rule_applies_to NOT NULL DEFAULT 'all',
    auto_alert  BOOLEAN             NOT NULL DEFAULT TRUE,
    auto_block  BOOLEAN             NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_aml_rules_tenant_company ON aml_monitoring_rules (tenant_id, company_id);
CREATE UNIQUE INDEX idx_aml_rules_code_unique ON aml_monitoring_rules (tenant_id, company_id, rule_code) WHERE deleted_at IS NULL;
