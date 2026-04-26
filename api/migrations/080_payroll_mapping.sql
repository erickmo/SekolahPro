-- 080: payroll_mapping — Mapping employee-nasabah (ADR-K020)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS payroll_mapping (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    employee_id     UUID NOT NULL,
    employee_number VARCHAR NOT NULL,
    employee_name   VARCHAR NOT NULL,
    nasabah_id      UUID NOT NULL,
    rekening_id     UUID NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    mapped_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    notes           TEXT,

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_pm_employee UNIQUE (tenant_id, company_id, employee_id)
);

CREATE INDEX idx_pm_tenant_company ON payroll_mapping (tenant_id, company_id);
CREATE INDEX idx_pm_employee ON payroll_mapping (employee_id);
CREATE INDEX idx_pm_nasabah ON payroll_mapping (nasabah_id);
CREATE INDEX idx_pm_rekening ON payroll_mapping (rekening_id);
CREATE INDEX idx_pm_active ON payroll_mapping (is_active) WHERE is_active = true;
CREATE INDEX idx_pm_rels ON payroll_mapping USING GIN (_rels);
CREATE INDEX idx_pm_data ON payroll_mapping USING GIN (_data);
