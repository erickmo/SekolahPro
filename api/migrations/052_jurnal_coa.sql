-- +migrate Up
-- ADR-K015: Jurnal & COA / Accounting
-- Tables: coa, jurnal, jurnal_line, journal_mapping, accounting_period
--
-- Hybrid Vernon pattern for coa, jurnal, journal_mapping, accounting_period.
-- jurnal_line is a regular detail table (no Vernon _rels/_data).

--------------------------------------------------------------------------------
-- K015-1: coa (Chart of Accounts) — Vernon domain
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS coa (
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

    account_code        VARCHAR(10) NOT NULL,
    account_name        VARCHAR     NOT NULL,
    parent_id           UUID        REFERENCES coa(id),
    level               INT         NOT NULL DEFAULT 1,
    account_type        VARCHAR(20) NOT NULL,
    normal_balance      VARCHAR(6)  NOT NULL,
    description         TEXT,
    is_system           BOOLEAN     NOT NULL DEFAULT false,
    is_active           BOOLEAN     NOT NULL DEFAULT true,
    coop_type_required  VARCHAR(10) NOT NULL DEFAULT 'both',

    CONSTRAINT chk_coa_account_type
        CHECK (account_type IN ('asset','liability','equity','revenue','expense','zakat','kebajikan','tazir')),
    CONSTRAINT chk_coa_normal_balance
        CHECK (normal_balance IN ('debit','credit')),
    CONSTRAINT chk_coa_coop_type_required
        CHECK (coop_type_required IN ('general','islamic','both')),
    CONSTRAINT uq_coa_tenant_code UNIQUE (tenant_id, company_id, account_code)
);

CREATE INDEX IF NOT EXISTS idx_coa_scope
    ON coa (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_coa_data
    ON coa USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_coa_sync
    ON coa (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_coa_parent
    ON coa (parent_id) WHERE parent_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_coa_type
    ON coa (tenant_id, company_id, account_type) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- K015-2: accounting_period — Vernon domain
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS accounting_period (
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

    period_type         VARCHAR(10) NOT NULL,
    year                INT         NOT NULL,
    month               INT,
    period_name         VARCHAR,
    start_date          DATE        NOT NULL,
    end_date            DATE        NOT NULL,
    status              VARCHAR(10) NOT NULL DEFAULT 'open',

    CONSTRAINT chk_acct_period_type
        CHECK (period_type IN ('monthly','annual')),
    CONSTRAINT chk_acct_period_status
        CHECK (status IN ('open','closed','locked')),
    CONSTRAINT chk_acct_period_month
        CHECK (month IS NULL OR (month >= 1 AND month <= 12)),
    CONSTRAINT chk_acct_period_dates
        CHECK (end_date >= start_date),
    CONSTRAINT uq_acct_period UNIQUE (tenant_id, company_id, period_type, year, month)
);

CREATE INDEX IF NOT EXISTS idx_acct_period_scope
    ON accounting_period (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_acct_period_data
    ON accounting_period USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_acct_period_sync
    ON accounting_period (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_acct_period_dates
    ON accounting_period (tenant_id, company_id, start_date, end_date) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- K015-3: jurnal (header) — Vernon domain
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS jurnal (
    id                  UUID           PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id           UUID           NOT NULL,
    company_id          UUID           NOT NULL,
    _rels               JSONB          NOT NULL DEFAULT '{}',
    _data               JSONB          NOT NULL DEFAULT '{}',
    _sync_status        TEXT           NOT NULL DEFAULT 'synced',
    _sync_version       BIGINT         NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ,

    journal_number      VARCHAR        NOT NULL,
    journal_date        DATE           NOT NULL,
    description         TEXT           NOT NULL,
    source_type         VARCHAR(20)    NOT NULL DEFAULT 'manual',
    source_id           UUID,
    reference           VARCHAR,
    period_id           UUID           NOT NULL,
    branch_id           UUID,
    total_debit         NUMERIC(15,2)  NOT NULL DEFAULT 0,
    total_credit        NUMERIC(15,2)  NOT NULL DEFAULT 0,
    status              VARCHAR(10)    NOT NULL DEFAULT 'unposted',
    is_auto_post        BOOLEAN        NOT NULL DEFAULT false,

    CONSTRAINT chk_jurnal_source_type
        CHECK (source_type IN ('transaction','manual','closing','adjustment','shu_distribution')),
    CONSTRAINT chk_jurnal_status
        CHECK (status IN ('unposted','posted','reversed')),
    CONSTRAINT chk_jurnal_balanced
        CHECK (total_debit = total_credit),
    CONSTRAINT chk_jurnal_amounts_positive
        CHECK (total_debit >= 0 AND total_credit >= 0),
    CONSTRAINT uq_jurnal_number UNIQUE (tenant_id, company_id, journal_number)
);

CREATE INDEX IF NOT EXISTS idx_jurnal_scope
    ON jurnal (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_jurnal_data
    ON jurnal USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jurnal_sync
    ON jurnal (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jurnal_period
    ON jurnal (tenant_id, company_id, period_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jurnal_branch
    ON jurnal (branch_id) WHERE branch_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jurnal_date
    ON jurnal (tenant_id, company_id, journal_date DESC) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- K015-4: jurnal_line (detail) — regular table, NOT a Vernon domain
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS jurnal_line (
    id                  UUID           PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id           UUID           NOT NULL,
    company_id          UUID           NOT NULL,
    jurnal_id           UUID           NOT NULL REFERENCES jurnal(id) ON DELETE CASCADE,
    coa_id              UUID           NOT NULL REFERENCES coa(id),
    line_number         INT            NOT NULL,
    debit               NUMERIC(15,2)  NOT NULL DEFAULT 0,
    credit              NUMERIC(15,2)  NOT NULL DEFAULT 0,
    description         TEXT,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT now(),

    CONSTRAINT chk_jurnal_line_debit_credit
        CHECK (
            debit >= 0 AND credit >= 0
            AND (debit > 0 OR credit > 0)
            AND NOT (debit > 0 AND credit > 0)
        ),
    CONSTRAINT uq_jurnal_line_number UNIQUE (jurnal_id, line_number)
);

CREATE INDEX IF NOT EXISTS idx_jurnal_line_jurnal
    ON jurnal_line (jurnal_id);
CREATE INDEX IF NOT EXISTS idx_jurnal_line_coa
    ON jurnal_line (coa_id);
CREATE INDEX IF NOT EXISTS idx_jurnal_line_scope
    ON jurnal_line (tenant_id, company_id);

--------------------------------------------------------------------------------
-- K015-5: journal_mapping — Vernon domain
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS journal_mapping (
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

    transaction_type    VARCHAR     NOT NULL,
    coop_type           VARCHAR(10) NOT NULL DEFAULT 'both',
    rules               JSONB       NOT NULL DEFAULT '[]',
    description         TEXT,
    is_active           BOOLEAN     NOT NULL DEFAULT true,

    CONSTRAINT chk_jmap_coop_type
        CHECK (coop_type IN ('general','islamic','both')),
    CONSTRAINT uq_jmap_type UNIQUE (tenant_id, company_id, transaction_type, coop_type)
);

CREATE INDEX IF NOT EXISTS idx_jmap_scope
    ON journal_mapping (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_jmap_data
    ON journal_mapping USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jmap_sync
    ON journal_mapping (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS journal_mapping;
DROP TABLE IF EXISTS jurnal_line;
DROP TABLE IF EXISTS jurnal;
DROP TABLE IF EXISTS accounting_period;
DROP TABLE IF EXISTS coa;
