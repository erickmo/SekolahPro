-- Schema untuk sqlc code generation: AML Transaction Alerts (ADR-K027).
-- tenant_id dan company_id wajib ada di semua tabel.

CREATE TYPE aml_alert_type AS ENUM (
    'suspicious', 'threshold_exceeded', 'unusual_pattern',
    'structuring', 'rapid_movement', 'mismatch'
);
CREATE TYPE aml_severity AS ENUM ('low', 'medium', 'high', 'critical');
CREATE TYPE aml_alert_status AS ENUM ('new', 'investigating', 'resolved', 'escalated', 'reported');
CREATE TYPE aml_outcome AS ENUM ('false_positive', 'suspicious', 'confirmed', 'escalated');

CREATE TABLE aml_transaction_alerts (
    id                  UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID             NOT NULL,
    company_id          UUID             NOT NULL,
    nasabah_id          UUID             NOT NULL,
    transaksi_id        UUID,
    rule_id             UUID             NOT NULL,
    alert_type          aml_alert_type   NOT NULL,
    severity            aml_severity     NOT NULL DEFAULT 'medium',
    description         TEXT             NOT NULL DEFAULT '',
    transaction_details JSONB            NOT NULL DEFAULT '{}',
    status              aml_alert_status NOT NULL DEFAULT 'new',
    investigator_id     UUID,
    investigation_notes TEXT,
    outcome             aml_outcome,
    ltkm_filed          BOOLEAN          NOT NULL DEFAULT FALSE,
    ltkm_reference      VARCHAR(100),
    created_at          TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_aml_alerts_tenant_company ON aml_transaction_alerts (tenant_id, company_id);
CREATE INDEX idx_aml_alerts_status ON aml_transaction_alerts (tenant_id, company_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_aml_alerts_nasabah ON aml_transaction_alerts (tenant_id, company_id, nasabah_id) WHERE deleted_at IS NULL;
