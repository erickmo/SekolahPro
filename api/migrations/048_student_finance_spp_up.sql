-- Sprint 6: ADR-S009 — Student Finance / SPP
-- 3 tabel: fee_types, student_invoices, student_payments

-- ============================================================================
-- 1. fee_types — Master jenis tagihan
-- ============================================================================

CREATE TABLE IF NOT EXISTS fee_types (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

-- Expression indexes on _data FK fields (Vernon pattern)
CREATE INDEX IF NOT EXISTS idx_fee_type_tenant_company ON fee_types (tenant_id, company_id);
CREATE INDEX IF NOT EXISTS idx_fee_type_data ON fee_types USING GIN (_data);
CREATE INDEX IF NOT EXISTS idx_fee_type_rels ON fee_types USING GIN (_rels);

-- ============================================================================
-- 2. student_invoices — Tagihan per siswa
-- ============================================================================

CREATE TABLE IF NOT EXISTS student_invoices (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

-- Expression indexes on _data FK fields (Vernon pattern)
CREATE INDEX IF NOT EXISTS idx_invoice_tenant_company ON student_invoices (tenant_id, company_id);
CREATE INDEX IF NOT EXISTS idx_invoice_data_student_id ON student_invoices ((_data->>'student_id')) WHERE _data->>'student_id' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_invoice_data_academic_year_id ON student_invoices ((_data->>'academic_year_id')) WHERE _data->>'academic_year_id' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_invoice_data_fee_type_id ON student_invoices ((_data->>'fee_type_id')) WHERE _data->>'fee_type_id' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_invoice_data_status ON student_invoices ((_data->>'status')) WHERE _data->>'status' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_invoice_data_due_date ON student_invoices ((_data->>'due_date')) WHERE _data->>'due_date' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_invoice_data_invoice_no ON student_invoices ((_data->>'invoice_no')) WHERE _data->>'invoice_no' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_invoice_data_period_year ON student_invoices ((_data->>'period_year')) WHERE _data->>'period_year' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_invoice_rels ON student_invoices USING GIN (_rels);
CREATE INDEX IF NOT EXISTS idx_invoice_data ON student_invoices USING GIN (_data);

-- ============================================================================
-- 3. student_payments — Pembayaran
-- ============================================================================

CREATE TABLE IF NOT EXISTS student_payments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

-- Expression indexes on _data FK fields (Vernon pattern)
CREATE INDEX IF NOT EXISTS idx_payment_tenant_company ON student_payments (tenant_id, company_id);
CREATE INDEX IF NOT EXISTS idx_payment_data_invoice_id ON student_payments ((_data->>'invoice_id')) WHERE _data->>'invoice_id' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payment_data_student_id ON student_payments ((_data->>'student_id')) WHERE _data->>'student_id' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payment_data_payment_date ON student_payments ((_data->>'payment_date')) WHERE _data->>'payment_date' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payment_data_receipt_no ON student_payments ((_data->>'receipt_no')) WHERE _data->>'receipt_no' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payment_rels ON student_payments USING GIN (_rels);
CREATE INDEX IF NOT EXISTS idx_payment_data ON student_payments USING GIN (_data);
