-- 078: payroll_batch — Batch payroll deduction (ADR-K020)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS payroll_batch (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id               UUID NOT NULL,
    company_id              UUID NOT NULL,

    batch_number            VARCHAR NOT NULL,
    period_month            INT NOT NULL,
    period_year             INT NOT NULL,
    upload_source           VARCHAR(10) NOT NULL DEFAULT 'api',
    total_employees         INT NOT NULL DEFAULT 0,
    matched_employees       INT NOT NULL DEFAULT 0,
    unmatched_employees     INT NOT NULL DEFAULT 0,
    total_deduction_amount  BIGINT NOT NULL DEFAULT 0,
    total_deductions        INT NOT NULL DEFAULT 0,
    skipped_deductions      INT NOT NULL DEFAULT 0,
    status                  VARCHAR(20) NOT NULL DEFAULT 'uploaded',
    calculated_at           TIMESTAMPTZ,
    approved_at             TIMESTAMPTZ,
    approved_by             UUID,
    executed_at             TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,
    error_message           TEXT,
    raw_data_url            VARCHAR,

    _rels                   JSONB NOT NULL DEFAULT '{}',
    _data                   JSONB NOT NULL DEFAULT '{}',
    _sync_status            TEXT NOT NULL DEFAULT 'synced',
    _sync_version           BIGINT NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ,

    CONSTRAINT chk_pb_source CHECK (upload_source IN ('csv', 'xlsx', 'api')),
    CONSTRAINT chk_pb_status CHECK (status IN (
        'uploaded', 'calculated', 'pending_approval', 'approved',
        'executing', 'executed', 'completed', 'failed'
    )),
    CONSTRAINT chk_pb_month CHECK (period_month BETWEEN 1 AND 12),
    CONSTRAINT uq_pb_batch_number UNIQUE (tenant_id, company_id, batch_number)
);

CREATE INDEX idx_pb_tenant_company ON payroll_batch (tenant_id, company_id);
CREATE INDEX idx_pb_status ON payroll_batch (status);
CREATE INDEX idx_pb_period ON payroll_batch (period_year, period_month);
CREATE INDEX idx_pb_rels ON payroll_batch USING GIN (_rels);
CREATE INDEX idx_pb_data ON payroll_batch USING GIN (_data);
