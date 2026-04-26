-- 076: ewallet_card — Kartu belanja uang saku digital (ADR-K021)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS ewallet_card (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id           UUID NOT NULL,
    company_id          UUID NOT NULL,

    nasabah_id          UUID NOT NULL,
    spending_control_id UUID,
    card_type           VARCHAR(20) NOT NULL,
    card_number         VARCHAR NOT NULL,
    card_uid            VARCHAR,
    status              VARCHAR(20) NOT NULL DEFAULT 'active',
    activated_at        TIMESTAMPTZ,
    frozen_at           TIMESTAMPTZ,
    frozen_by           UUID,
    deactivated_at      TIMESTAMPTZ,
    expires_at          DATE,

    _rels               JSONB NOT NULL DEFAULT '{}',
    _data               JSONB NOT NULL DEFAULT '{}',
    _sync_status        TEXT NOT NULL DEFAULT 'synced',
    _sync_version       BIGINT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT chk_ec_type CHECK (card_type IN ('nfc', 'qr_static', 'qr_dynamic', 'barcode')),
    CONSTRAINT chk_ec_status CHECK (status IN ('active', 'frozen', 'deactivated', 'lost')),
    CONSTRAINT uq_ec_card_number UNIQUE (card_number)
);

CREATE INDEX idx_ec_tenant_company ON ewallet_card (tenant_id, company_id);
CREATE INDEX idx_ec_nasabah ON ewallet_card (nasabah_id);
CREATE INDEX idx_ec_spending_control ON ewallet_card (spending_control_id);
CREATE INDEX idx_ec_status ON ewallet_card (status);
CREATE INDEX idx_ec_card_uid ON ewallet_card (card_uid);
CREATE INDEX idx_ec_rels ON ewallet_card USING GIN (_rels);
CREATE INDEX idx_ec_data ON ewallet_card USING GIN (_data);
