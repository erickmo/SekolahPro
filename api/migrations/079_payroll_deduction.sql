-- 079: payroll_deduction — Potongan per employee (ADR-K020)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS payroll_deduction (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id           UUID NOT NULL,
    company_id          UUID NOT NULL,

    batch_id            UUID NOT NULL,
    nasabah_id          UUID NOT NULL,
    rekening_id         UUID NOT NULL,
    authorization_id    UUID,
    employee_name       VARCHAR,
    employee_number     VARCHAR,
    deduction_type      VARCHAR(30) NOT NULL,
    deduction_amount    BIGINT NOT NULL DEFAULT 0,
    match_status        VARCHAR(20) NOT NULL DEFAULT 'matched',
    deduction_status    VARCHAR(20) NOT NULL DEFAULT 'pending',
    executed_at         TIMESTAMPTZ,
    error_message       TEXT,
    notes               TEXT,

    _rels               JSONB NOT NULL DEFAULT '{}',
    _data               JSONB NOT NULL DEFAULT '{}',
    _sync_status        TEXT NOT NULL DEFAULT 'synced',
    _sync_version       BIGINT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT chk_pd_deduction CHECK (deduction_type IN ('simpanan', 'pinjaman', 'spp', 'kantin', 'other')),
    CONSTRAINT chk_pd_match CHECK (match_status IN ('matched', 'unmatched', 'skipped')),
    CONSTRAINT chk_pd_status CHECK (deduction_status IN ('pending', 'executed', 'failed', 'skipped'))
);

CREATE INDEX idx_pd_tenant_company ON payroll_deduction (tenant_id, company_id);
CREATE INDEX idx_pd_batch ON payroll_deduction (batch_id);
CREATE INDEX idx_pd_nasabah ON payroll_deduction (nasabah_id);
CREATE INDEX idx_pd_rekening ON payroll_deduction (rekening_id);
CREATE INDEX idx_pd_authorization ON payroll_deduction (authorization_id);
CREATE INDEX idx_pd_status ON payroll_deduction (deduction_status);
CREATE INDEX idx_pd_rels ON payroll_deduction USING GIN (_rels);
CREATE INDEX idx_pd_data ON payroll_deduction USING GIN (_data);
