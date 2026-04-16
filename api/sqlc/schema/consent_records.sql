-- Schema untuk consent records (ADR-K028: Data Privacy / UU PDP).
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE consent_records (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID          NOT NULL,
    company_id            UUID          NOT NULL,
    nasabah_id            UUID          NOT NULL,

    -- Consent Info
    consent_type          VARCHAR(30)   NOT NULL,
    purpose               TEXT          NOT NULL,
    legal_basis           VARCHAR(30)   NOT NULL,

    -- Consent Detail
    consent_text          TEXT          NOT NULL,
    consent_version       VARCHAR(20)   NOT NULL,
    consent_given         BOOLEAN       NOT NULL DEFAULT FALSE,
    consent_method        VARCHAR(30)   NOT NULL,

    -- Withdrawal
    withdrawn             BOOLEAN       NOT NULL DEFAULT FALSE,
    withdrawn_at          TIMESTAMPTZ,
    withdrawal_reason     TEXT,

    -- Parental Consent (untuk minor)
    parent_id             UUID,
    parent_relationship   VARCHAR(20),
    parent_consent_given  BOOLEAN,

    -- Audit
    given_at              TIMESTAMPTZ   NOT NULL,
    ip_address            VARCHAR(45),
    user_agent            VARCHAR(500),
    witness_id            UUID,
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

-- Index untuk pencarian berdasarkan tenant + company (selalu dipakai).
CREATE INDEX idx_consent_records_tenant_company ON consent_records (tenant_id, company_id);

-- Index untuk filter berdasarkan nasabah.
CREATE INDEX idx_consent_records_nasabah ON consent_records (tenant_id, company_id, nasabah_id)
    WHERE deleted_at IS NULL;

-- Index untuk filter berdasarkan consent type.
CREATE INDEX idx_consent_records_type ON consent_records (tenant_id, company_id, consent_type)
    WHERE deleted_at IS NULL;

-- Index untuk filter berdasarkan status withdrawn.
CREATE INDEX idx_consent_records_withdrawn ON consent_records (tenant_id, company_id, withdrawn)
    WHERE deleted_at IS NULL;
