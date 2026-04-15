-- +migrate Up
-- Domain Vernon: academic_years (ADR-010)
-- Tahun ajaran — unit waktu fundamental, referenced by 10+ student domains.
-- Data field disimpan di _data JSONB, expression index untuk constraint.

CREATE TABLE IF NOT EXISTS academic_years (
    id            UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id     UUID        NOT NULL,
    company_id    UUID        NOT NULL,
    _rels         JSONB       NOT NULL DEFAULT '{}',
    _data         JSONB       NOT NULL DEFAULT '{}',
    _sync_status  TEXT        NOT NULL DEFAULT 'synced',
    _sync_version BIGINT      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

-- Standard Vernon indexes
CREATE INDEX IF NOT EXISTS idx_academic_years_scope
    ON academic_years (tenant_id, company_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_academic_years_data
    ON academic_years USING GIN (_data)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_academic_years_sync
    ON academic_years (_sync_status)
    WHERE _sync_status != 'synced' AND deleted_at IS NULL;

-- Domain-specific: unique name per company
CREATE UNIQUE INDEX IF NOT EXISTS uq_academic_year_name
    ON academic_years (tenant_id, company_id, ((_data->>'name')))
    WHERE deleted_at IS NULL;

-- Domain-specific: unique code per company
CREATE UNIQUE INDEX IF NOT EXISTS uq_academic_year_code
    ON academic_years (tenant_id, company_id, ((_data->>'code')))
    WHERE deleted_at IS NULL;

-- Domain-specific: hanya satu tahun ajaran aktif per company
CREATE UNIQUE INDEX IF NOT EXISTS uq_academic_year_active
    ON academic_years (tenant_id, company_id)
    WHERE ((_data->>'is_active')::boolean) = true AND deleted_at IS NULL;

-- Domain-specific: query by status
CREATE INDEX IF NOT EXISTS idx_academic_years_status
    ON academic_years (((_data->>'status')))
    WHERE deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS academic_years;
