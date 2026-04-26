-- 066: toko_lokasi — Lokasi toko/kantin (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_lokasi (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    name            VARCHAR NOT NULL,
    location_type   VARCHAR(20) NOT NULL,
    building        VARCHAR(50),
    floor           VARCHAR(10),
    room_number     VARCHAR(20),
    is_active       BOOLEAN NOT NULL DEFAULT true,
    notes           TEXT,

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_toko_lokasi_type CHECK (location_type IN ('toko', 'kantin', 'kantin_asrama', 'kiosk'))
);

CREATE INDEX idx_toko_lokasi_tenant_company ON toko_lokasi (tenant_id, company_id);
CREATE INDEX idx_toko_lokasi_active ON toko_lokasi (is_active) WHERE is_active = true;
CREATE INDEX idx_toko_lokasi_rels ON toko_lokasi USING GIN (_rels);
CREATE INDEX idx_toko_lokasi_data ON toko_lokasi USING GIN (_data);
