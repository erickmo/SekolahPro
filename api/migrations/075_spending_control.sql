-- 075: spending_control — Kontrol belanja uang saku digital (ADR-K021)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS spending_control (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id               UUID NOT NULL,
    company_id              UUID NOT NULL,

    nasabah_id              UUID NOT NULL,
    controlled_by_id        UUID,
    rekening_id             UUID,
    daily_limit             BIGINT,
    per_transaction_limit   BIGINT,
    weekly_limit            BIGINT,
    monthly_limit           BIGINT,
    category_restrictions   JSONB,
    time_restrictions       JSONB,
    is_active               BOOLEAN NOT NULL DEFAULT true,
    pin_hash                VARCHAR(255),
    pin_threshold           BIGINT,

    _rels                   JSONB NOT NULL DEFAULT '{}',
    _data                   JSONB NOT NULL DEFAULT '{}',
    _sync_status            TEXT NOT NULL DEFAULT 'synced',
    _sync_version           BIGINT NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ
);

CREATE INDEX idx_sc_tenant_company ON spending_control (tenant_id, company_id);
CREATE INDEX idx_sc_nasabah ON spending_control (nasabah_id);
CREATE INDEX idx_sc_controlled_by ON spending_control (controlled_by_id);
CREATE INDEX idx_sc_rekening ON spending_control (rekening_id);
CREATE INDEX idx_sc_active ON spending_control (is_active) WHERE is_active = true;
CREATE INDEX idx_sc_rels ON spending_control USING GIN (_rels);
CREATE INDEX idx_sc_data ON spending_control USING GIN (_data);
