-- 054: laporan_regulasi — ADR-K017 Regulatory Reporting
-- Tables: laporan_config, laporan, laporan_versi

--------------------------------------------------------------------------------
-- laporan_config (konfigurasi per jenis laporan per tenant)
-- No BelongsTo — standalone config table
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS laporan_config (
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

    report_type         VARCHAR(50)   NOT NULL,
    report_name         VARCHAR(200)  NOT NULL,
    report_category     VARCHAR(30)   NOT NULL,
    description         TEXT,
    frequency           VARCHAR(20)   NOT NULL,
    template_config     JSONB         NOT NULL DEFAULT '{}',
    is_active           BOOLEAN       NOT NULL DEFAULT true,
    deadline_days_after_period INTEGER,

    CONSTRAINT chk_laporan_config_category
        CHECK (report_category IN ('dinas_koperasi','ojk','islamic','internal')),
    CONSTRAINT chk_laporan_config_frequency
        CHECK (frequency IN ('daily','weekly','monthly','quarterly','annually')),
    CONSTRAINT uq_laporan_config_type
        UNIQUE (tenant_id, company_id, report_type)
);

CREATE INDEX IF NOT EXISTS idx_laporan_config_scope
    ON laporan_config (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_laporan_config_data
    ON laporan_config USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_laporan_config_sync
    ON laporan_config (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_laporan_config_category
    ON laporan_config (tenant_id, company_id, report_category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_laporan_config_active
    ON laporan_config (tenant_id, company_id) WHERE is_active = true AND deleted_at IS NULL;

--------------------------------------------------------------------------------
-- laporan (record per laporan yang di-generate)
-- BelongsTo: branches, laporan_config
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS laporan (
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

    laporan_config_id   UUID          NOT NULL,
    report_type         VARCHAR(50)   NOT NULL,
    report_category     VARCHAR(30)   NOT NULL,
    period_type         VARCHAR(20)   NOT NULL,
    period_start        DATE          NOT NULL,
    period_end          DATE          NOT NULL,
    period_label        VARCHAR(100)  NOT NULL,
    report_data         JSONB         NOT NULL DEFAULT '{}',
    generated_at        TIMESTAMPTZ   NOT NULL DEFAULT now(),
    generated_by        UUID          NOT NULL,
    branch_id           UUID,
    revision_of         UUID,
    submitted_at        TIMESTAMPTZ,
    submitted_by        UUID,
    reviewed_at         TIMESTAMPTZ,
    reviewed_by         UUID,
    anomaly_flags       JSONB         DEFAULT '[]',
    status              VARCHAR(20)   NOT NULL DEFAULT 'draft',
    consolidation_level VARCHAR(20)   NOT NULL DEFAULT 'branch',

    CONSTRAINT chk_laporan_status
        CHECK (status IN ('draft','reviewed','final','submitted','overdue')),
    CONSTRAINT chk_laporan_period_type
        CHECK (period_type IN ('daily','weekly','monthly','quarterly','annually')),
    CONSTRAINT chk_laporan_report_category
        CHECK (report_category IN ('dinas_koperasi','ojk','islamic','internal')),
    CONSTRAINT chk_laporan_consolidation
        CHECK (consolidation_level IN ('branch','company','tenant')),
    CONSTRAINT uq_laporan_period_version
        UNIQUE (tenant_id, company_id, report_type, period_start, period_end, status)
);

CREATE INDEX IF NOT EXISTS idx_laporan_scope
    ON laporan (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_laporan_data
    ON laporan USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_laporan_sync
    ON laporan (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_laporan_config_id
    ON laporan (laporan_config_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_laporan_branch_id
    ON laporan (branch_id) WHERE branch_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_laporan_status
    ON laporan (tenant_id, company_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_laporan_period
    ON laporan (tenant_id, company_id, period_start, period_end) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_laporan_type_category
    ON laporan (tenant_id, company_id, report_type, report_category) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- laporan_versi (versioning — setiap regenerasi/revisi menyimpan versi)
-- BelongsTo: laporan
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS laporan_versi (
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

    laporan_id          UUID        NOT NULL,
    version_number      INTEGER     NOT NULL,
    report_data         JSONB       NOT NULL DEFAULT '{}',
    created_reason      VARCHAR(30) NOT NULL,
    change_summary      TEXT,

    CONSTRAINT chk_laporan_versi_reason
        CHECK (created_reason IN ('auto_generate','manual_trigger','regenerate','revision')),
    CONSTRAINT uq_laporan_versi_number
        UNIQUE (laporan_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_laporan_versi_scope
    ON laporan_versi (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_laporan_versi_data
    ON laporan_versi USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_laporan_versi_sync
    ON laporan_versi (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_laporan_versi_laporan_id
    ON laporan_versi (laporan_id) WHERE deleted_at IS NULL;
