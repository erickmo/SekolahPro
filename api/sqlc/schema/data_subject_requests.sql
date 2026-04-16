-- Schema untuk data subject requests (ADR-K028: Data Privacy / UU PDP).
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE data_subject_requests (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID          NOT NULL,
    company_id            UUID          NOT NULL,
    nasabah_id            UUID          NOT NULL,

    -- Request Info
    request_type          VARCHAR(30)   NOT NULL,
    description           TEXT          NOT NULL,
    data_categories       JSONB         NOT NULL DEFAULT '[]',

    -- Processing
    status                VARCHAR(30)   NOT NULL DEFAULT 'received',
    assigned_to           UUID,
    rejection_reason      TEXT,

    -- Response
    response_data         JSONB,
    response_summary      TEXT,

    -- SLA
    received_at           TIMESTAMPTZ   NOT NULL,
    deadline_at           TIMESTAMPTZ   NOT NULL,
    completed_at          TIMESTAMPTZ,

    -- Audit
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

-- Index untuk pencarian berdasarkan tenant + company (selalu dipakai).
CREATE INDEX idx_dsr_tenant_company ON data_subject_requests (tenant_id, company_id);

-- Index untuk filter berdasarkan nasabah.
CREATE INDEX idx_dsr_nasabah ON data_subject_requests (tenant_id, company_id, nasabah_id)
    WHERE deleted_at IS NULL;

-- Index untuk filter berdasarkan status.
CREATE INDEX idx_dsr_status ON data_subject_requests (tenant_id, company_id, status)
    WHERE deleted_at IS NULL;

-- Index untuk filter berdasarkan request type.
CREATE INDEX idx_dsr_request_type ON data_subject_requests (tenant_id, company_id, request_type)
    WHERE deleted_at IS NULL;
