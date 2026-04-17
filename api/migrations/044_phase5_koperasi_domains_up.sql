-- +migrate Up
-- Phase 5: Koperasi Domain Enhancements (K009-K010)
-- Denda (penalties), Jaminan (collateral) + sub-tables.
--
-- Hybrid Vernon pattern: 10 standard columns + explicit domain-specific columns.
-- All CHECK constraints and UNIQUE constraints are enforced at DB level.

--------------------------------------------------------------------------------
-- K009: denda (replaces stub 029 — enhanced with full penalty management)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS denda (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id           UUID        NOT NULL,
    company_id          UUID        NOT NULL,
    _rels               JSONB       NOT NULL DEFAULT '{}',
    _data               JSONB       NOT NULL DEFAULT '{}',
    _sync_status        TEXT        NOT NULL DEFAULT 'synced',
    _sync_version       BIGINT      NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ,

    pinjaman_id         UUID          NOT NULL,
    angsuran_id         UUID,
    rekening_id         UUID          NOT NULL,
    penalty_type        VARCHAR(30)   NOT NULL,
    calculation_basis   NUMERIC(15,2) NOT NULL,
    penalty_rate        NUMERIC(8,6),
    penalty_days        INT           NOT NULL DEFAULT 0,
    calculated_amount   NUMERIC(15,2) NOT NULL,
    cap_applied         BOOLEAN       NOT NULL DEFAULT false,
    capped_amount       NUMERIC(15,2),
    final_amount        NUMERIC(15,2) NOT NULL,
    paid_amount         NUMERIC(15,2) NOT NULL DEFAULT 0,
    waived_amount       NUMERIC(15,2) NOT NULL DEFAULT 0,
    outstanding_amount  NUMERIC(15,2) NOT NULL,
    status              VARCHAR(20)   NOT NULL DEFAULT 'accruing',
    waiver_reason       TEXT,
    waiver_approved_by  UUID,
    waiver_approved_at  TIMESTAMPTZ,
    fund_destination    VARCHAR(20)   NOT NULL DEFAULT 'koperasi_income',
    period_start        DATE          NOT NULL,
    period_end          DATE,
    created_by          UUID,
    updated_by          UUID,

    CONSTRAINT chk_denda_penalty_type
        CHECK (penalty_type IN ('late_payment','early_settlement','administrative','legal_fee','other')),
    CONSTRAINT chk_denda_status
        CHECK (status IN ('accruing','billed','paid','waived')),
    CONSTRAINT chk_denda_fund_destination
        CHECK (fund_destination IN ('koperasi_income','benevolent_fund')),
    CONSTRAINT chk_denda_amounts_positive
        CHECK (
            calculation_basis >= 0 AND
            calculated_amount >= 0 AND
            final_amount >= 0 AND
            paid_amount >= 0 AND
            waived_amount >= 0 AND
            outstanding_amount >= 0
        ),
    CONSTRAINT chk_denda_capped_amount
        CHECK (capped_amount IS NULL OR capped_amount >= 0),
    CONSTRAINT chk_denda_penalty_rate
        CHECK (penalty_rate IS NULL OR penalty_rate >= 0)
);

CREATE INDEX IF NOT EXISTS idx_denda_scope
    ON denda (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_denda_data
    ON denda USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_denda_sync
    ON denda (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_denda_pinjaman
    ON denda (pinjaman_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_denda_angsuran
    ON denda (angsuran_id) WHERE angsuran_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_denda_rekening
    ON denda (rekening_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_denda_status
    ON denda (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_denda_penalty_type
    ON denda (penalty_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_denda_period
    ON denda (period_start, period_end) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_denda_outstanding
    ON denda (tenant_id, company_id) WHERE outstanding_amount > 0 AND deleted_at IS NULL;

--------------------------------------------------------------------------------
-- K010: jaminan (replaces stub 030 — enhanced with full collateral management)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS jaminan (
    id                     UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id              UUID        NOT NULL,
    company_id             UUID        NOT NULL,
    _rels                  JSONB       NOT NULL DEFAULT '{}',
    _data                  JSONB       NOT NULL DEFAULT '{}',
    _sync_status           TEXT        NOT NULL DEFAULT 'synced',
    _sync_version          BIGINT      NOT NULL DEFAULT 0,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ,

    nasabah_id             UUID          NOT NULL,
    collateral_number      VARCHAR(50)   NOT NULL,
    collateral_type        VARCHAR(30)   NOT NULL,
    description            TEXT          NOT NULL,
    item_category          VARCHAR(50),
    item_brand             VARCHAR(100),
    item_year              INT,
    item_serial_number     VARCHAR(100),
    ownership_name         VARCHAR(200)  NOT NULL,
    ownership_document     VARCHAR(100),
    rekening_id            UUID,
    hold_amount            NUMERIC(15,2),
    appraised_value        NUMERIC(15,2) NOT NULL,
    acceptance_rate        NUMERIC(5,4)  NOT NULL DEFAULT 1.0000,
    collateral_value       NUMERIC(15,2) NOT NULL,
    appraised_at           TIMESTAMPTZ   NOT NULL,
    appraised_by           UUID          NOT NULL,
    next_revaluation_at    TIMESTAMPTZ,
    status                 VARCHAR(20)   NOT NULL DEFAULT 'registered',
    released_at            TIMESTAMPTZ,
    released_by            UUID,
    release_notes          TEXT,
    foreclosed_at          TIMESTAMPTZ,
    foreclosed_by          UUID,
    sale_price             NUMERIC(15,2),
    sale_date              DATE,
    sale_buyer             VARCHAR(200),
    sale_notes             TEXT,
    ujrah_monthly          NUMERIC(15,2),
    ujrah_total_charged    NUMERIC(15,2) NOT NULL DEFAULT 0,
    ujrah_total_paid       NUMERIC(15,2) NOT NULL DEFAULT 0,
    created_by             UUID,
    updated_by             UUID,

    CONSTRAINT chk_jaminan_collateral_type
        CHECK (collateral_type IN ('blandongan','motor_vehicle','land_certificate','other')),
    CONSTRAINT chk_jaminan_status
        CHECK (status IN ('registered','active','released','foreclosed')),
    CONSTRAINT chk_jaminan_values_positive
        CHECK (
            appraised_value >= 0 AND
            acceptance_rate >= 0 AND
            collateral_value >= 0 AND
            ujrah_total_charged >= 0 AND
            ujrah_total_paid >= 0
        ),
    CONSTRAINT chk_jaminan_hold_amount
        CHECK (hold_amount IS NULL OR hold_amount >= 0),
    CONSTRAINT chk_jaminan_sale_price
        CHECK (sale_price IS NULL OR sale_price >= 0),
    CONSTRAINT chk_jaminan_ujrah_monthly
        CHECK (ujrah_monthly IS NULL OR ujrah_monthly >= 0),
    CONSTRAINT chk_jaminan_item_year
        CHECK (item_year IS NULL OR item_year BETWEEN 1900 AND 2100)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_jaminan_collateral_number
    ON jaminan (collateral_number);
CREATE INDEX IF NOT EXISTS idx_jaminan_scope
    ON jaminan (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_jaminan_data
    ON jaminan USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jaminan_sync
    ON jaminan (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_jaminan_nasabah
    ON jaminan (nasabah_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jaminan_rekening
    ON jaminan (rekening_id) WHERE rekening_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jaminan_status
    ON jaminan (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jaminan_collateral_type
    ON jaminan (collateral_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jaminan_appraiser
    ON jaminan (appraised_by) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jaminan_next_revaluation
    ON jaminan (next_revaluation_at) WHERE next_revaluation_at IS NOT NULL AND deleted_at IS NULL;

--------------------------------------------------------------------------------
-- K010: jaminan_pinjaman (junction: collateral <-> loan binding)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS jaminan_pinjaman (
    id                UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id         UUID        NOT NULL,
    company_id        UUID        NOT NULL,
    _rels             JSONB       NOT NULL DEFAULT '{}',
    _data             JSONB       NOT NULL DEFAULT '{}',
    _sync_status      TEXT        NOT NULL DEFAULT 'synced',
    _sync_version     BIGINT      NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    jaminan_id        UUID         NOT NULL,
    pinjaman_id       UUID         NOT NULL,
    pledged_value     NUMERIC(15,2) NOT NULL,
    status            VARCHAR(20)  NOT NULL DEFAULT 'active',
    bound_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    released_at       TIMESTAMPTZ,

    CONSTRAINT chk_jp_status
        CHECK (status IN ('active','released')),
    CONSTRAINT chk_jp_pledged_value
        CHECK (pledged_value >= 0)
);

CREATE INDEX IF NOT EXISTS idx_jaminan_pinjaman_scope
    ON jaminan_pinjaman (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_jaminan_pinjaman_data
    ON jaminan_pinjaman USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jaminan_pinjaman_sync
    ON jaminan_pinjaman (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_jp_jaminan
    ON jaminan_pinjaman (jaminan_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jp_pinjaman
    ON jaminan_pinjaman (pinjaman_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jp_status
    ON jaminan_pinjaman (status) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- K010: jaminan_dokumen (collateral document attachments)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS jaminan_dokumen (
    id                 UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id          UUID        NOT NULL,
    company_id         UUID        NOT NULL,
    _rels              JSONB       NOT NULL DEFAULT '{}',
    _data              JSONB       NOT NULL DEFAULT '{}',
    _sync_status       TEXT        NOT NULL DEFAULT 'synced',
    _sync_version      BIGINT      NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ,

    jaminan_id         UUID         NOT NULL,
    document_type      VARCHAR(30)  NOT NULL,
    file_url           VARCHAR(500) NOT NULL,
    file_name          VARCHAR(255) NOT NULL,
    file_size_bytes    BIGINT       NOT NULL,
    mime_type          VARCHAR(100) NOT NULL,
    description        TEXT,
    version            INT          NOT NULL DEFAULT 1,
    is_superseded      BOOLEAN      NOT NULL DEFAULT false,
    superseded_by_id   UUID,
    created_by         UUID,

    CONSTRAINT chk_jd_document_type
        CHECK (document_type IN (
            'ownership_deed','vehicle_registration','land_certificate',
            'photo_front','photo_back','appraisal_report',
            'insurance_document','agreement_letter','other'
        )),
    CONSTRAINT chk_jd_file_size
        CHECK (file_size_bytes > 0),
    CONSTRAINT chk_jd_version
        CHECK (version >= 1)
);

CREATE INDEX IF NOT EXISTS idx_jaminan_dokumen_scope
    ON jaminan_dokumen (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_jaminan_dokumen_data
    ON jaminan_dokumen USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jaminan_dokumen_sync
    ON jaminan_dokumen (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_jd_jaminan
    ON jaminan_dokumen (jaminan_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jd_document_type
    ON jaminan_dokumen (document_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jd_superseded
    ON jaminan_dokumen (jaminan_id, version) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jd_created_by
    ON jaminan_dokumen (created_by) WHERE created_by IS NOT NULL AND deleted_at IS NULL;

--------------------------------------------------------------------------------
-- K010: jaminan_valuasi (collateral revaluation history)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS jaminan_valuasi (
    id                 UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id          UUID        NOT NULL,
    company_id         UUID        NOT NULL,
    _rels              JSONB       NOT NULL DEFAULT '{}',
    _data              JSONB       NOT NULL DEFAULT '{}',
    _sync_status       TEXT        NOT NULL DEFAULT 'synced',
    _sync_version      BIGINT      NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ,

    jaminan_id         UUID          NOT NULL,
    appraised_value    NUMERIC(15,2) NOT NULL,
    acceptance_rate    NUMERIC(5,4)  NOT NULL,
    collateral_value   NUMERIC(15,2) NOT NULL,
    valuation_type     VARCHAR(20)   NOT NULL,
    valuation_notes    TEXT,
    appraised_by       UUID          NOT NULL,
    appraised_at       TIMESTAMPTZ   NOT NULL,
    created_by         UUID,

    CONSTRAINT chk_jv_valuation_type
        CHECK (valuation_type IN ('initial','periodic','on_demand')),
    CONSTRAINT chk_jv_values_positive
        CHECK (
            appraised_value >= 0 AND
            acceptance_rate >= 0 AND
            collateral_value >= 0
        )
);

CREATE INDEX IF NOT EXISTS idx_jaminan_valuasi_scope
    ON jaminan_valuasi (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_jaminan_valuasi_data
    ON jaminan_valuasi USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jaminan_valuasi_sync
    ON jaminan_valuasi (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_jv_jaminan
    ON jaminan_valuasi (jaminan_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jv_valuation_type
    ON jaminan_valuasi (valuation_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jv_appraiser
    ON jaminan_valuasi (appraised_by) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jv_appraised_at
    ON jaminan_valuasi (appraised_at DESC) WHERE deleted_at IS NULL;
