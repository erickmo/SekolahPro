-- 077: payroll_authorization — Otorisasi potongan payroll (ADR-K020)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS payroll_authorization (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    nasabah_id      UUID NOT NULL,
    rekening_id     UUID NOT NULL,
    authorization_type VARCHAR(30) NOT NULL,
    deduction_type  VARCHAR(30) NOT NULL,
    max_amount      BIGINT,
    max_percentage  DECIMAL(5,2),
    effective_date  DATE NOT NULL,
    expiry_date     DATE,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    authorized_by   UUID,
    notes           TEXT,

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_pa_type CHECK (authorization_type IN ('payroll_deduction', 'auto_debit', 'standing_order')),
    CONSTRAINT chk_pa_deduction CHECK (deduction_type IN ('simpanan', 'pinjaman', 'spp', 'kantin', 'other')),
    CONSTRAINT chk_pa_status CHECK (status IN ('active', 'revoked', 'expired'))
);

CREATE INDEX idx_pa_tenant_company ON payroll_authorization (tenant_id, company_id);
CREATE INDEX idx_pa_nasabah ON payroll_authorization (nasabah_id);
CREATE INDEX idx_pa_rekening ON payroll_authorization (rekening_id);
CREATE INDEX idx_pa_status ON payroll_authorization (status);
CREATE INDEX idx_pa_rels ON payroll_authorization USING GIN (_rels);
CREATE INDEX idx_pa_data ON payroll_authorization USING GIN (_data);
