-- +migrate Up
-- Branch Financial Summaries: laporan keuangan per cabang (null branch_id = konsolidasi)
CREATE TABLE IF NOT EXISTS branch_financial_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    branch_id UUID,  -- NULL berarti konsolidasi (agregasi semua cabang)

    -- Period
    period_type TEXT NOT NULL CHECK (period_type IN ('daily','weekly','monthly','quarterly','yearly')),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,

    -- P&L fields (satuan terkecil, BIGINT)
    pendapatan_operasional  BIGINT NOT NULL DEFAULT 0,
    pendapatan_bunga_margin BIGINT NOT NULL DEFAULT 0,
    pendapatan_lain         BIGINT NOT NULL DEFAULT 0,
    total_pendapatan        BIGINT NOT NULL DEFAULT 0,
    biaya_operasional       BIGINT NOT NULL DEFAULT 0,
    biaya_personel          BIGINT NOT NULL DEFAULT 0,
    biaya_administrasi      BIGINT NOT NULL DEFAULT 0,
    beban_ppap              BIGINT NOT NULL DEFAULT 0,
    total_biaya             BIGINT NOT NULL DEFAULT 0,
    laba_rugi_bersih        BIGINT NOT NULL DEFAULT 0,

    -- Balance Sheet fields
    total_aset        BIGINT NOT NULL DEFAULT 0,
    kas_dan_bank      BIGINT NOT NULL DEFAULT 0,
    pinjaman_diberikan BIGINT NOT NULL DEFAULT 0,
    simpanan_diterima BIGINT NOT NULL DEFAULT 0,
    total_kewajiban   BIGINT NOT NULL DEFAULT 0,
    modal_sendiri     BIGINT NOT NULL DEFAULT 0,

    -- KPI fields (nullable)
    npl_ratio  NUMERIC(10,4),
    bopo_ratio NUMERIC(10,4),
    roa        NUMERIC(10,4),
    car        NUMERIC(10,4),

    -- Audit
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_by   UUID,  -- NULL = belum disetujui
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,

    -- Unique: satu laporan per tenant+branch+period_type+period_start
    UNIQUE (tenant_id, branch_id, period_type, period_start)
);

-- Indexes untuk query umum
CREATE INDEX idx_bfs_tenant_company ON branch_financial_summaries (tenant_id, company_id);
CREATE INDEX idx_bfs_branch ON branch_financial_summaries (tenant_id, company_id, branch_id);
CREATE INDEX idx_bfs_period ON branch_financial_summaries (tenant_id, company_id, period_type, period_start);
CREATE INDEX idx_bfs_consolidated ON branch_financial_summaries (tenant_id, company_id, period_type, period_start) WHERE branch_id IS NOT NULL;

-- Elimination Rules: aturan eliminasi untuk consolidated reporting
CREATE TABLE IF NOT EXISTS elimination_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    rule_name TEXT NOT NULL,
    rule_type TEXT NOT NULL CHECK (rule_type IN ('inter_branch_transfer','loan_participation','profit_elimination')),
    from_branch_id UUID NOT NULL,
    to_branch_id UUID NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_er_tenant_company ON elimination_rules (tenant_id, company_id);

-- +migrate Down
DROP TABLE IF EXISTS elimination_rules;
DROP TABLE IF EXISTS branch_financial_summaries;
