-- Schema untuk sqlc code generation: Customer Due Diligence (ADR-K027).
-- tenant_id dan company_id wajib ada di semua tabel.

CREATE TYPE aml_risk_level AS ENUM ('low', 'medium', 'high', 'prohibited');
CREATE TYPE aml_cdd_level AS ENUM ('sdd', 'cdd', 'edd');
CREATE TYPE aml_cdd_status AS ENUM ('pending', 'active', 'escalated', 'restricted', 'exited');
CREATE TYPE aml_pep_type AS ENUM ('none', 'domestic', 'foreign', 'international_organization');

CREATE TABLE aml_customer_due_diligences (
    id                UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID              NOT NULL,
    company_id        UUID              NOT NULL,
    nasabah_id        UUID              NOT NULL,
    risk_level        aml_risk_level    NOT NULL DEFAULT 'low',
    risk_score        INT               NOT NULL DEFAULT 1 CHECK (risk_score BETWEEN 1 AND 100),
    cdd_level         aml_cdd_level     NOT NULL DEFAULT 'sdd',
    is_pep            BOOLEAN           NOT NULL DEFAULT FALSE,
    pep_type          aml_pep_type      NOT NULL DEFAULT 'none',
    pep_position      VARCHAR(255),
    purpose           TEXT              NOT NULL DEFAULT '',
    source_of_funds   TEXT              NOT NULL DEFAULT '',
    source_of_wealth  TEXT              NOT NULL DEFAULT '',
    next_review_date  DATE,
    status            aml_cdd_status    NOT NULL DEFAULT 'pending',
    created_at        TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_aml_cdd_tenant_company ON aml_customer_due_diligences (tenant_id, company_id);
CREATE UNIQUE INDEX idx_aml_cdd_nasabah_unique ON aml_customer_due_diligences (tenant_id, company_id, nasabah_id) WHERE deleted_at IS NULL;
