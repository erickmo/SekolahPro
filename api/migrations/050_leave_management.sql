-- 050: Leave Management (ADR-S030)
-- leave_types, leave_balances, leave_requests, leave_approval_logs

-- ============================================================
-- leave_types: konfigurasi jenis cuti per sekolah
-- ============================================================
CREATE TABLE IF NOT EXISTS leave_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    _rels JSONB NOT NULL DEFAULT '{}',
    _data JSONB NOT NULL DEFAULT '{}',
    _sync_status TEXT NOT NULL DEFAULT 'synced',
    _sync_version BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_leave_type_tenant_company ON leave_types (tenant_id, company_id);
CREATE INDEX idx_leave_type_rels ON leave_types USING GIN (_rels);
CREATE INDEX idx_leave_type_data ON leave_types USING GIN (_data);

-- ============================================================
-- leave_balances: saldo cuti per guru per tahun
-- ============================================================
CREATE TABLE IF NOT EXISTS leave_balances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    _rels JSONB NOT NULL DEFAULT '{}',
    _data JSONB NOT NULL DEFAULT '{}',
    _sync_status TEXT NOT NULL DEFAULT 'synced',
    _sync_version BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_leave_bal_tenant_company ON leave_balances (tenant_id, company_id);
CREATE INDEX idx_leave_bal_teacher ON leave_balances (tenant_id, company_id, (_data->>'teacher_id'));
CREATE INDEX idx_leave_bal_year ON leave_balances (tenant_id, company_id, (_data->>'year'));
CREATE INDEX idx_leave_bal_teacher_year ON leave_balances (tenant_id, company_id, (_data->>'teacher_id'), (_data->>'year'));
CREATE INDEX idx_leave_bal_rels ON leave_balances USING GIN (_rels);
CREATE INDEX idx_leave_bal_data ON leave_balances USING GIN (_data);

-- ============================================================
-- leave_requests: pengajuan cuti
-- ============================================================
CREATE TABLE IF NOT EXISTS leave_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    _rels JSONB NOT NULL DEFAULT '{}',
    _data JSONB NOT NULL DEFAULT '{}',
    _sync_status TEXT NOT NULL DEFAULT 'synced',
    _sync_version BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_leave_req_tenant_company ON leave_requests (tenant_id, company_id);
CREATE INDEX idx_leave_req_teacher ON leave_requests (tenant_id, company_id, (_data->>'teacher_id'));
CREATE INDEX idx_leave_req_status ON leave_requests (tenant_id, company_id, (_data->>'status'));
CREATE INDEX idx_leave_req_dates ON leave_requests (tenant_id, company_id, (_data->>'start_date'), (_data->>'end_date'));
CREATE INDEX idx_leave_req_teacher_status ON leave_requests (tenant_id, company_id, (_data->>'teacher_id'), (_data->>'status'));
CREATE INDEX idx_leave_req_rels ON leave_requests USING GIN (_rels);
CREATE INDEX idx_leave_req_data ON leave_requests USING GIN (_data);

-- ============================================================
-- leave_approval_logs: riwayat approval per level
-- ============================================================
CREATE TABLE IF NOT EXISTS leave_approval_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    _rels JSONB NOT NULL DEFAULT '{}',
    _data JSONB NOT NULL DEFAULT '{}',
    _sync_status TEXT NOT NULL DEFAULT 'synced',
    _sync_version BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_leave_log_tenant_company ON leave_approval_logs (tenant_id, company_id);
CREATE INDEX idx_leave_log_request ON leave_approval_logs (tenant_id, company_id, (_data->>'leave_request_id'));
CREATE INDEX idx_leave_log_approver ON leave_approval_logs (tenant_id, company_id, (_data->>'approver_id'));
CREATE INDEX idx_leave_log_rels ON leave_approval_logs USING GIN (_rels);
CREATE INDEX idx_leave_log_data ON leave_approval_logs USING GIN (_data);
