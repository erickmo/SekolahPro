-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE incident_records (
    id                     UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id              UUID          NOT NULL,
    company_id             UUID          NOT NULL,
    incident_number        VARCHAR(20)   NOT NULL,
    severity               VARCHAR(2)    NOT NULL,
    incident_type          VARCHAR(30)   NOT NULL,
    title                  VARCHAR(255)  NOT NULL,
    description            TEXT          NOT NULL DEFAULT '',
    detected_at            TIMESTAMPTZ   NOT NULL,
    acknowledged_at        TIMESTAMPTZ,
    mitigated_at           TIMESTAMPTZ,
    resolved_at            TIMESTAMPTZ,
    post_mortem_at         TIMESTAMPTZ,
    affected_services      JSONB         NOT NULL DEFAULT '[]',
    affected_tenants       JSONB         DEFAULT NULL,
    affected_nasabah       INT           NOT NULL DEFAULT 0,
    data_loss              BOOLEAN       NOT NULL DEFAULT FALSE,
    data_loss_description  TEXT,
    responder_ids          JSONB         NOT NULL DEFAULT '[]',
    actions_taken          JSONB         NOT NULL DEFAULT '[]',
    root_cause             TEXT,
    remediation            TEXT,
    post_mortem_doc_id     UUID,
    status                 VARCHAR(20)   NOT NULL DEFAULT 'detected',
    created_at             TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at             TIMESTAMPTZ
);

CREATE INDEX idx_incident_records_tenant_company ON incident_records (tenant_id, company_id);
CREATE UNIQUE INDEX idx_incident_records_number ON incident_records (tenant_id, company_id, incident_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_incident_records_severity ON incident_records (tenant_id, company_id, severity) WHERE deleted_at IS NULL;
CREATE INDEX idx_incident_records_status ON incident_records (tenant_id, company_id, status) WHERE deleted_at IS NULL;
