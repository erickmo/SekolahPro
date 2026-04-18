-- 058: assets — Assets & Asset Maintenances
-- Vernon pattern: _rels/_data JSONB columns.

--------------------------------------------------------------------------------
-- assets
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS assets (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Domain data
    asset_code      VARCHAR(30) NOT NULL,
    name            VARCHAR(200) NOT NULL,
    category        VARCHAR(20) NOT NULL,
    description     TEXT,
    acquisition_date DATE,
    acquisition_cost NUMERIC(14,2),
    ownership_type  VARCHAR(15) NOT NULL DEFAULT 'owned',
    condition       VARCHAR(15) NOT NULL DEFAULT 'good',
    location_room_id UUID,
    responsible_person VARCHAR(100),
    vendor          VARCHAR(200),
    warranty_expiry DATE,
    depreciation_method VARCHAR(20) DEFAULT 'straight_line',
    useful_life_years INT,
    book_value      NUMERIC(14,2),
    lifecycle_status VARCHAR(20) NOT NULL DEFAULT 'in_use',
    disposal_date   DATE,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_assets_category CHECK (category IN (
        'tanah', 'bangunan', 'mesin', 'kendaraan', 'peralatan', 'lainnya'
    )),
    CONSTRAINT chk_assets_ownership CHECK (ownership_type IN (
        'owned', 'leased', 'donated'
    )),
    CONSTRAINT chk_assets_condition CHECK (condition IN (
        'new', 'good', 'fair', 'poor', 'damaged'
    )),
    CONSTRAINT chk_assets_lifecycle CHECK (lifecycle_status IN (
        'in_use', 'idle', 'maintenance', 'disposed', 'lost', 'transferred'
    )),
    CONSTRAINT uq_assets_code UNIQUE (tenant_id, company_id, asset_code)
);

CREATE INDEX idx_assets_tenant_company ON assets (tenant_id, company_id);
CREATE INDEX idx_assets_category ON assets (category);
CREATE INDEX idx_assets_condition ON assets (condition);
CREATE INDEX idx_assets_lifecycle ON assets (lifecycle_status);
CREATE INDEX idx_assets_location ON assets (location_room_id) WHERE location_room_id IS NOT NULL;
CREATE INDEX idx_assets_acquisition_date ON assets (acquisition_date) WHERE acquisition_date IS NOT NULL;
CREATE INDEX idx_assets_rels ON assets USING GIN (_rels);
CREATE INDEX idx_assets_data ON assets USING GIN (_data);

--------------------------------------------------------------------------------
-- asset_maintenances
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS asset_maintenances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    asset_id        UUID NOT NULL,

    -- Domain data
    type            VARCHAR(20) NOT NULL,
    description     TEXT NOT NULL,
    start_date      DATE NOT NULL,
    completion_date DATE,
    cost            NUMERIC(14,2),
    vendor          VARCHAR(200),
    technician      VARCHAR(100),
    status          VARCHAR(15) NOT NULL DEFAULT 'scheduled',
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_am_type CHECK (type IN (
        'preventive', 'corrective', 'emergency', 'calibration', 'inspection'
    )),
    CONSTRAINT chk_am_status CHECK (status IN (
        'scheduled', 'in_progress', 'completed', 'cancelled'
    )),
    CONSTRAINT chk_am_cost CHECK (cost IS NULL OR cost >= 0)
);

CREATE INDEX idx_asset_maintenances_tenant_company ON asset_maintenances (tenant_id, company_id);
CREATE INDEX idx_am_asset ON asset_maintenances (asset_id);
CREATE INDEX idx_am_dates ON asset_maintenances (start_date);
CREATE INDEX idx_asset_maintenances_status ON asset_maintenances (status);
CREATE INDEX idx_asset_maintenances_type ON asset_maintenances (type);
CREATE INDEX idx_asset_maintenances_rels ON asset_maintenances USING GIN (_rels);
CREATE INDEX idx_asset_maintenances_data ON asset_maintenances USING GIN (_data);
