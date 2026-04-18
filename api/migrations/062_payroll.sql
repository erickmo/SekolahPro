-- 062: payroll — Configs, Periods, Entries, Components
-- Vernon pattern: _rels/_data JSONB columns.

--------------------------------------------------------------------------------
-- payroll_configs
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payroll_configs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,

    -- Domain data
    employee_type   VARCHAR(15) NOT NULL,
    base_salary     NUMERIC(14,2) NOT NULL,
    transport_allowance NUMERIC(14,2) DEFAULT 0,
    meal_allowance  NUMERIC(14,2) DEFAULT 0,
    position_allowance NUMERIC(14,2) DEFAULT 0,
    family_allowance NUMERIC(14,2) DEFAULT 0,
    rice_allowance  NUMERIC(14,2) DEFAULT 0,
    pph_status      VARCHAR(15) NOT NULL DEFAULT 'non_pkp',
    bank_name       VARCHAR(50),
    bank_account    VARCHAR(50),
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_pc_employee_type CHECK (employee_type IN (
        'pns', 'honorer', 'yayasan'
    )),
    CONSTRAINT chk_pc_pph_status CHECK (pph_status IN (
        'non_pkp', 'ptkp', 'pkp'
    )),
    CONSTRAINT chk_pc_base_salary CHECK (base_salary >= 0),
    CONSTRAINT uq_pc_teacher UNIQUE (tenant_id, company_id, teacher_id)
);

CREATE INDEX idx_payroll_configs_tenant_company ON payroll_configs (tenant_id, company_id);
CREATE INDEX idx_payroll_configs_teacher ON payroll_configs (teacher_id);
CREATE INDEX idx_payroll_configs_employee_type ON payroll_configs (employee_type);
CREATE INDEX idx_payroll_configs_active ON payroll_configs (is_active) WHERE is_active = true;
CREATE INDEX idx_payroll_configs_rels ON payroll_configs USING GIN (_rels);
CREATE INDEX idx_payroll_configs_data ON payroll_configs USING GIN (_data);

--------------------------------------------------------------------------------
-- payroll_periods
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payroll_periods (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    academic_year_id UUID NOT NULL,

    -- Domain data
    period_name     VARCHAR(50) NOT NULL,
    month           INT NOT NULL,
    year            INT NOT NULL,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    workflow_status VARCHAR(20) NOT NULL DEFAULT 'draft',
    total_employees INT DEFAULT 0,
    total_gross     NUMERIC(16,2) DEFAULT 0,
    total_deductions NUMERIC(16,2) DEFAULT 0,
    total_net       NUMERIC(16,2) DEFAULT 0,
    processed_by    UUID,
    processed_at    TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_pp_month CHECK (month BETWEEN 1 AND 12),
    CONSTRAINT chk_pp_workflow CHECK (workflow_status IN (
        'draft', 'calculating', 'calculated', 'approved',
        'processing', 'paid', 'cancelled', 'failed'
    )),
    CONSTRAINT chk_pp_dates CHECK (end_date >= start_date),
    CONSTRAINT chk_pp_totals CHECK (
        total_gross >= 0 AND total_deductions >= 0 AND total_net >= 0
    ),
    CONSTRAINT uq_pp_period UNIQUE (tenant_id, company_id, month, year)
);

CREATE INDEX idx_payroll_periods_tenant_company ON payroll_periods (tenant_id, company_id);
CREATE INDEX idx_payroll_periods_academic_year ON payroll_periods (academic_year_id);
CREATE INDEX idx_payroll_periods_workflow ON payroll_periods (workflow_status);
CREATE INDEX idx_payroll_periods_year_month ON payroll_periods (year, month);
CREATE INDEX idx_payroll_periods_rels ON payroll_periods USING GIN (_rels);
CREATE INDEX idx_payroll_periods_data ON payroll_periods USING GIN (_data);

--------------------------------------------------------------------------------
-- payroll_entries
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payroll_entries (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    period_id       UUID NOT NULL,
    teacher_id      UUID NOT NULL,

    -- Domain data
    employee_type   VARCHAR(15) NOT NULL,
    base_salary     NUMERIC(14,2) NOT NULL,
    total_allowances NUMERIC(14,2) DEFAULT 0,
    gross_salary    NUMERIC(14,2) NOT NULL,
    total_deductions NUMERIC(14,2) DEFAULT 0,
    net_salary      NUMERIC(14,2) NOT NULL,
    pph21_amount    NUMERIC(14,2) DEFAULT 0,
    working_days    INT NOT NULL DEFAULT 0,
    present_days    INT NOT NULL DEFAULT 0,
    absent_days     INT NOT NULL DEFAULT 0,
    leave_days      INT NOT NULL DEFAULT 0,
    payslip_number  VARCHAR(30) NOT NULL,
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_pe_employee_type CHECK (employee_type IN (
        'pns', 'honorer', 'yayasan'
    )),
    CONSTRAINT chk_pe_amounts CHECK (
        base_salary >= 0 AND gross_salary >= 0 AND
        total_deductions >= 0 AND net_salary >= 0 AND pph21_amount >= 0 AND
        total_allowances >= 0
    ),
    CONSTRAINT chk_pe_days CHECK (
        working_days >= 0 AND present_days >= 0 AND
        absent_days >= 0 AND leave_days >= 0
    ),
    CONSTRAINT uq_pe_period_teacher UNIQUE (tenant_id, company_id, period_id, teacher_id)
);

CREATE INDEX idx_payroll_entries_tenant_company ON payroll_entries (tenant_id, company_id);
CREATE INDEX idx_pe_period ON payroll_entries (period_id);
CREATE INDEX idx_pe_teacher ON payroll_entries (teacher_id);
CREATE INDEX idx_payroll_entries_employee_type ON payroll_entries (employee_type);
CREATE INDEX idx_payroll_entries_payslip ON payroll_entries (payslip_number);
CREATE INDEX idx_payroll_entries_rels ON payroll_entries USING GIN (_rels);
CREATE INDEX idx_payroll_entries_data ON payroll_entries USING GIN (_data);

--------------------------------------------------------------------------------
-- payroll_components
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payroll_components (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    entry_id        UUID NOT NULL,

    -- Domain data
    component_type  VARCHAR(10) NOT NULL,
    category        VARCHAR(30) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    amount          NUMERIC(14,2) NOT NULL,
    calculation_method VARCHAR(20) NOT NULL DEFAULT 'fixed',
    is_recurring    BOOLEAN NOT NULL DEFAULT true,
    reference_id    UUID,
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_pcom_type CHECK (component_type IN (
        'earning', 'deduction'
    )),
    CONSTRAINT chk_pcom_method CHECK (calculation_method IN (
        'fixed', 'percentage', 'formula', 'attendance_based'
    )),
    CONSTRAINT chk_pcom_amount CHECK (amount >= 0)
);

CREATE INDEX idx_payroll_components_tenant_company ON payroll_components (tenant_id, company_id);
CREATE INDEX idx_pc_entry ON payroll_components (entry_id);
CREATE INDEX idx_pc_type ON payroll_components (component_type);
CREATE INDEX idx_payroll_components_category ON payroll_components (category);
CREATE INDEX idx_payroll_components_rels ON payroll_components USING GIN (_rels);
CREATE INDEX idx_payroll_components_data ON payroll_components USING GIN (_data);
