-- +migrate Up
-- Mobile App Configuration: menyimpan konfigurasi per varian aplikasi mobile.
-- Unique constraint pada (tenant_id, company_id, app_variant) memastikan
-- hanya satu konfigurasi aktif per varian per organisasi.
CREATE TABLE IF NOT EXISTS mobile_app_configs (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID         NOT NULL,
    company_id       UUID         NOT NULL,
    app_variant      VARCHAR(20)  NOT NULL CHECK (app_variant IN ('student', 'parent', 'staff', 'full')),
    platform         VARCHAR(10)  NOT NULL DEFAULT 'both' CHECK (platform IN ('android', 'ios', 'both')),
    min_version      VARCHAR(20)  NOT NULL DEFAULT '',
    current_version  VARCHAR(20)  NOT NULL,
    force_update     BOOLEAN      NOT NULL DEFAULT FALSE,
    maintenance_mode BOOLEAN      NOT NULL DEFAULT FALSE,
    feature_flags    JSONB        NOT NULL DEFAULT '{}',
    api_base_url     VARCHAR(255) NOT NULL,
    theme_config     JSONB        NOT NULL DEFAULT '{}',
    offline_config   JSONB        NOT NULL DEFAULT '{"sync_interval":300,"max_cache_size":50,"offline_features":[]}',
    is_active        BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

-- Satu konfigurasi aktif per varian per organisasi (soft-delete dikecualikan).
CREATE UNIQUE INDEX IF NOT EXISTS uq_mobile_app_configs_tenant_variant
    ON mobile_app_configs(tenant_id, company_id, app_variant)
    WHERE deleted_at IS NULL;

-- Index untuk query list dengan filter app_variant dan platform.
CREATE INDEX IF NOT EXISTS idx_mobile_app_configs_tenant_company_created
    ON mobile_app_configs(tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_mobile_app_configs_tenant_company_variant
    ON mobile_app_configs(tenant_id, company_id, app_variant)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_mobile_app_configs_tenant_company_platform
    ON mobile_app_configs(tenant_id, company_id, platform)
    WHERE deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS mobile_app_configs;
