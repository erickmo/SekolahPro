-- +migrate Up
-- Hierarki organisasi: tenant_id (L1) → company_id (L2) → branch_id (L3, opsional)
-- Semua tabel wajib punya tenant_id dan company_id agar isolasi data bekerja
-- pada mode single maupun multi tenant.
CREATE TABLE IF NOT EXISTS examples (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID         NOT NULL,
    company_id  UUID         NOT NULL,
    branch_id   UUID,                    -- NULL berarti tidak terikat ke cabang tertentu
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- tenant_id sebagai leading column di semua composite index
-- agar query dengan WHERE tenant_id = ? AND company_id = ? memanfaatkan index secara optimal.
CREATE INDEX IF NOT EXISTS idx_examples_tenant_company_created
    ON examples(tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_examples_tenant_company_name
    ON examples(tenant_id, company_id, name);
CREATE INDEX IF NOT EXISTS idx_examples_tenant_company_active
    ON examples(tenant_id, company_id, is_active)
    WHERE deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS examples;
