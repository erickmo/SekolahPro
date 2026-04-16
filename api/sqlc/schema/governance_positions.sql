-- Schema untuk governance positions (ADR-K025: Cooperative Governance & Internal Controls).
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE governance_positions (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID          NOT NULL,
    company_id            UUID          NOT NULL,
    nasabah_id            UUID          NOT NULL,

    -- Position Info
    position_type         VARCHAR(30)   NOT NULL,
    position_level        VARCHAR(20)   NOT NULL,
    term_start            DATE          NOT NULL,
    term_end              DATE          NOT NULL,
    term_number           INT           NOT NULL DEFAULT 1,

    -- Status
    status                VARCHAR(20)   NOT NULL DEFAULT 'active',
    appointed_by          VARCHAR(30)   NOT NULL DEFAULT 'rat',

    -- Authority
    max_approval_amount   BIGINT        NOT NULL DEFAULT 0,
    can_disburse          BOOLEAN       NOT NULL DEFAULT FALSE,
    can_reverse           BOOLEAN       NOT NULL DEFAULT FALSE,
    can_waive_penalty     BOOLEAN       NOT NULL DEFAULT FALSE,
    can_write_off         BOOLEAN       NOT NULL DEFAULT FALSE,

    -- Audit
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

-- Index untuk pencarian berdasarkan tenant + company (selalu dipakai).
CREATE INDEX idx_governance_positions_tenant_company ON governance_positions (tenant_id, company_id);

-- Satu posisi hanya boleh diisi satu orang aktif per tenant.
CREATE UNIQUE INDEX idx_governance_positions_unique_active
    ON governance_positions (tenant_id, company_id, position_type)
    WHERE status = 'active' AND deleted_at IS NULL;

-- Index untuk filter berdasarkan nasabah.
CREATE INDEX idx_governance_positions_nasabah ON governance_positions (tenant_id, company_id, nasabah_id)
    WHERE deleted_at IS NULL;

-- Index untuk filter berdasarkan status.
CREATE INDEX idx_governance_positions_status ON governance_positions (tenant_id, company_id, status)
    WHERE deleted_at IS NULL;
