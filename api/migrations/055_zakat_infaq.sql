-- 055: zakat_infaq — ADR-K018 Zakat & Infaq
-- Tables: zakat_collection, zakat_distribution, infaq,
--         infaq_recurring_config (non-Vernon), mustahik, tazir_fund,
--         zakat_collection_batch (non-Vernon)

--------------------------------------------------------------------------------
-- zakat_collection (Vernon — BelongsTo nasabah, branches)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS zakat_collection (
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

    nasabah_id          UUID,
    branch_id           UUID,
    batch_id            UUID,
    muzakki_name        VARCHAR(200)  NOT NULL,
    zakat_type          VARCHAR(20)   NOT NULL,
    amount              NUMERIC(15,2) NOT NULL,
    payment_method      VARCHAR(20)   NOT NULL,
    fitrah_head_count   INTEGER,
    collection_date     DATE,
    confirmed_at        TIMESTAMPTZ,
    confirmed_by        UUID,
    notes               TEXT,
    status              VARCHAR(20)   NOT NULL DEFAULT 'pending',

    CONSTRAINT chk_zakat_collection_type
        CHECK (zakat_type IN ('zakat_mal','zakat_fitrah','zakat_institusi')),
    CONSTRAINT chk_zakat_collection_payment
        CHECK (payment_method IN ('cash','auto_debit','transfer')),
    CONSTRAINT chk_zakat_collection_status
        CHECK (status IN ('pending','confirmed','cancelled')),
    CONSTRAINT chk_zakat_collection_amount
        CHECK (amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_zakat_collection_scope
    ON zakat_collection (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_zakat_collection_data
    ON zakat_collection USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_zakat_collection_sync
    ON zakat_collection (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_zakat_collection_nasabah
    ON zakat_collection (nasabah_id) WHERE nasabah_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_zakat_collection_branch
    ON zakat_collection (branch_id) WHERE branch_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_zakat_collection_type_status
    ON zakat_collection (tenant_id, company_id, zakat_type, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_zakat_collection_batch
    ON zakat_collection (batch_id) WHERE batch_id IS NOT NULL AND deleted_at IS NULL;

--------------------------------------------------------------------------------
-- zakat_distribution (Vernon — BelongsTo mustahik, branches)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS zakat_distribution (
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

    mustahik_id         UUID          NOT NULL,
    branch_id           UUID,
    source_fund         VARCHAR(20)   NOT NULL,
    amount              NUMERIC(15,2) NOT NULL,
    purpose             TEXT          NOT NULL,
    asnaf_category      VARCHAR(20),
    distribution_date   DATE,
    approved_at         TIMESTAMPTZ,
    approved_by         UUID,
    proof_url           VARCHAR(500),
    notes               TEXT,
    status              VARCHAR(20)   NOT NULL DEFAULT 'draft',
    distribution_method VARCHAR(20)   NOT NULL DEFAULT 'cash',

    CONSTRAINT chk_zakat_distribution_source
        CHECK (source_fund IN ('zakat','infaq','tazir')),
    CONSTRAINT chk_zakat_distribution_status
        CHECK (status IN ('draft','pending','approved','rejected','completed')),
    CONSTRAINT chk_zakat_distribution_method
        CHECK (distribution_method IN ('cash','goods','account_transfer')),
    CONSTRAINT chk_zakat_distribution_amount
        CHECK (amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_zakat_distribution_scope
    ON zakat_distribution (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_zakat_distribution_data
    ON zakat_distribution USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_zakat_distribution_sync
    ON zakat_distribution (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_zakat_distribution_mustahik
    ON zakat_distribution (mustahik_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_zakat_distribution_branch
    ON zakat_distribution (branch_id) WHERE branch_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_zakat_distribution_source_status
    ON zakat_distribution (tenant_id, company_id, source_fund, status) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- infaq (Vernon — BelongsTo nasabah, branches)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS infaq (
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

    nasabah_id          UUID,
    branch_id           UUID,
    donatur_name        VARCHAR(200)  NOT NULL,
    infaq_type          VARCHAR(20)   NOT NULL,
    amount              NUMERIC(15,2) NOT NULL,
    designation         VARCHAR(100)  NOT NULL,
    payment_method      VARCHAR(20)   NOT NULL,
    infaq_date          DATE,
    recurring_config_id UUID,
    notes               TEXT,

    CONSTRAINT chk_infaq_type
        CHECK (infaq_type IN ('one_time','recurring')),
    CONSTRAINT chk_infaq_payment
        CHECK (payment_method IN ('cash','transfer','auto_debit')),
    CONSTRAINT chk_infaq_amount
        CHECK (amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_infaq_scope
    ON infaq (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_infaq_data
    ON infaq USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_infaq_sync
    ON infaq (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_infaq_nasabah
    ON infaq (nasabah_id) WHERE nasabah_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_infaq_branch
    ON infaq (branch_id) WHERE branch_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_infaq_type_designation
    ON infaq (tenant_id, company_id, infaq_type, designation) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- infaq_recurring_config (NON-Vernon — config table, no descriptor)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS infaq_recurring_config (
    id                      UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id               UUID        NOT NULL,
    company_id              UUID        NOT NULL,
    nasabah_id              UUID        NOT NULL,
    source_account_id       UUID        NOT NULL,
    amount                  NUMERIC(15,2) NOT NULL,
    frequency               VARCHAR(20)   NOT NULL,
    debit_day               INTEGER       NOT NULL,
    designation             VARCHAR(100)  NOT NULL,
    start_date              DATE          NOT NULL,
    end_date                DATE,
    authorization_doc_url   VARCHAR(500)  NOT NULL,
    next_debit_date         DATE          NOT NULL,
    is_active               BOOLEAN       NOT NULL DEFAULT true,
    created_at              TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT chk_infaq_recurring_freq
        CHECK (frequency IN ('weekly','monthly')),
    CONSTRAINT chk_infaq_recurring_amount
        CHECK (amount > 0),
    CONSTRAINT chk_infaq_recurring_debit_day
        CHECK (debit_day >= 1 AND debit_day <= 31)
);

CREATE INDEX IF NOT EXISTS idx_infaq_recurring_scope
    ON infaq_recurring_config (tenant_id, company_id);
CREATE INDEX IF NOT EXISTS idx_infaq_recurring_nasabah
    ON infaq_recurring_config (nasabah_id);
CREATE INDEX IF NOT EXISTS idx_infaq_recurring_active
    ON infaq_recurring_config (tenant_id, company_id) WHERE is_active = true;

--------------------------------------------------------------------------------
-- mustahik (Vernon — optional BelongsTo branch only)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mustahik (
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

    branch_id           UUID,
    full_name           VARCHAR(200)  NOT NULL,
    asnaf_categories    JSONB         NOT NULL DEFAULT '[]',
    primary_category    VARCHAR(30)   NOT NULL,
    needs_assessment    TEXT          NOT NULL,
    assessment_date     DATE          NOT NULL,
    assessed_by         UUID          NOT NULL,
    phone               VARCHAR(30),
    address             TEXT,
    next_review_date    DATE,
    status              VARCHAR(20)   NOT NULL DEFAULT 'active',

    CONSTRAINT chk_mustahik_status
        CHECK (status IN ('active','inactive','graduated'))
);

CREATE INDEX IF NOT EXISTS idx_mustahik_scope
    ON mustahik (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_mustahik_data
    ON mustahik USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_mustahik_sync
    ON mustahik (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_mustahik_branch
    ON mustahik (branch_id) WHERE branch_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_mustahik_status
    ON mustahik (tenant_id, company_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_mustahik_category
    ON mustahik (primary_category) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- tazir_fund (Vernon — BelongsTo denda, pinjaman, nasabah)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tazir_fund (
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

    denda_id            UUID          NOT NULL,
    pinjaman_id         UUID          NOT NULL,
    nasabah_id          UUID          NOT NULL,
    branch_id           UUID,
    amount              NUMERIC(15,2) NOT NULL,
    collection_date     DATE,
    distribution_id     UUID,
    notes               TEXT,
    status              VARCHAR(20)   NOT NULL DEFAULT 'collected',

    CONSTRAINT chk_tazir_fund_status
        CHECK (status IN ('collected','distributed')),
    CONSTRAINT chk_tazir_fund_amount
        CHECK (amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_tazir_fund_scope
    ON tazir_fund (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tazir_fund_data
    ON tazir_fund USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tazir_fund_sync
    ON tazir_fund (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tazir_fund_denda
    ON tazir_fund (denda_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tazir_fund_pinjaman
    ON tazir_fund (pinjaman_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tazir_fund_nasabah
    ON tazir_fund (nasabah_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tazir_fund_status
    ON tazir_fund (tenant_id, company_id, status) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- zakat_collection_batch (NON-Vernon — operational table, no descriptor)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS zakat_collection_batch (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id           UUID        NOT NULL,
    company_id          UUID        NOT NULL,
    batch_type          VARCHAR(30)   NOT NULL,
    description         TEXT          NOT NULL,
    total_records       INTEGER       NOT NULL DEFAULT 0,
    total_amount        NUMERIC(15,2) NOT NULL DEFAULT 0,
    status              VARCHAR(20)   NOT NULL DEFAULT 'draft',
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT chk_zakat_batch_type
        CHECK (batch_type IN ('fitrah_class','fitrah_bulk','general')),
    CONSTRAINT chk_zakat_batch_status
        CHECK (status IN ('draft','pending','approved','rejected'))
);

CREATE INDEX IF NOT EXISTS idx_zakat_batch_scope
    ON zakat_collection_batch (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_zakat_batch_status
    ON zakat_collection_batch (tenant_id, company_id, status);
