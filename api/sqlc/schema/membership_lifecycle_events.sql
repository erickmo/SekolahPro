-- Schema untuk sqlc code generation.
-- ADR-K030: Membership Lifecycle Management
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.

CREATE TYPE lifecycle_event_type AS ENUM (
    'approved', 'activated', 'suspended', 'reinstated',
    'dormant_detected', 'reactivated', 'resignation_submitted',
    'resignation_approved', 'resigned', 'expulsion_initiated',
    'expulsion_approved', 'expelled', 'deceased_reported',
    'settlement_initiated', 'settlement_completed', 'settled',
    'transfer_out', 'transfer_in'
);

CREATE TYPE membership_status AS ENUM (
    'applied', 'approved', 'active', 'suspended', 'dormant',
    'resigning', 'resigned', 'expelled', 'deceased', 'settling',
    'settled', 'transferred_out', 'transferred_in', 'rejected'
);

CREATE TYPE refund_status AS ENUM (
    'pending', 'processing', 'completed', 'hold'
);

CREATE TYPE loan_settlement_plan AS ENUM (
    'none', 'full_repayment', 'restructuring', 'write_off'
);

CREATE TABLE membership_lifecycle_events (
    id                       UUID                      PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                UUID                      NOT NULL,
    company_id               UUID                      NOT NULL,
    nasabah_id               UUID                      NOT NULL,

    -- Event Info
    event_type               lifecycle_event_type      NOT NULL,
    from_status              membership_status,
    to_status                membership_status         NOT NULL,

    -- Details
    reason                   TEXT                      NOT NULL DEFAULT '',
    initiated_by             UUID,
    approved_by              UUID,
    supporting_doc_ids      UUID[],

    -- Financial Impact
    simpanan_pokok_refund    BOOLEAN,
    simpanan_wajib_refund    BOOLEAN,
    refund_amount            BIGINT,
    refund_status            refund_status,
    outstanding_loans        BOOLEAN,
    loan_settlement_plan     loan_settlement_plan,

    -- Audit
    event_date               DATE                      NOT NULL,
    effective_date           DATE                      NOT NULL,
    created_at               TIMESTAMPTZ               NOT NULL DEFAULT NOW(),
    created_by               UUID                      NOT NULL,
    deleted_at               TIMESTAMPTZ
);

CREATE INDEX idx_lifecycle_events_nasabah
    ON membership_lifecycle_events (tenant_id, company_id, nasabah_id, event_date DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_lifecycle_events_type
    ON membership_lifecycle_events (tenant_id, company_id, event_type, event_date DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_lifecycle_events_date
    ON membership_lifecycle_events (tenant_id, company_id, event_date DESC)
    WHERE deleted_at IS NULL;
