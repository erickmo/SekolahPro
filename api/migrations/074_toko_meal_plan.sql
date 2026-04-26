-- 074: toko_meal_plan — Meal plan asrama (ADR-K019)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS toko_meal_plan (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    nasabah_id      UUID NOT NULL,
    rekening_id     UUID NOT NULL,
    plan_name       VARCHAR NOT NULL,
    meal_type       VARCHAR(20) NOT NULL,
    day_of_week     INT,
    start_date      DATE,
    end_date        DATE,
    menu_items      JSONB NOT NULL DEFAULT '[]',
    total_price     BIGINT NOT NULL DEFAULT 0,
    frequency       VARCHAR(20) NOT NULL DEFAULT 'daily',
    is_active       BOOLEAN NOT NULL DEFAULT true,
    notes           TEXT,

    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_tmp_meal_type CHECK (meal_type IN ('breakfast', 'lunch', 'dinner', 'snack')),
    CONSTRAINT chk_tmp_frequency CHECK (frequency IN ('daily', 'weekly', 'monthly')),
    CONSTRAINT chk_tmp_day CHECK (day_of_week IS NULL OR day_of_week BETWEEN 0 AND 6)
);

CREATE INDEX idx_tmp_tenant_company ON toko_meal_plan (tenant_id, company_id);
CREATE INDEX idx_tmp_nasabah ON toko_meal_plan (nasabah_id);
CREATE INDEX idx_tmp_rekening ON toko_meal_plan (rekening_id);
CREATE INDEX idx_tmp_active ON toko_meal_plan (is_active) WHERE is_active = true;
CREATE INDEX idx_tmp_rels ON toko_meal_plan USING GIN (_rels);
CREATE INDEX idx_tmp_data ON toko_meal_plan USING GIN (_data);
