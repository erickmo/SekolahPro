-- +migrate Up
-- ADR-K016: SHU (Sisa Hasil Usaha)
-- Tables: shu_config (regular), shu_periode (Vernon), shu_anggota (Vernon)
--
-- shu_config is a config-only table — no Vernon _rels/_data.
-- shu_periode and shu_anggota are Vernon domains.

--------------------------------------------------------------------------------
-- K016-1: shu_config — regular config table (NOT a Vernon domain)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shu_config (
    id                          UUID           PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id                   UUID           NOT NULL,
    company_id                  UUID           NOT NULL,
    cadangan_pct                NUMERIC(5,2)   NOT NULL,
    jasa_anggota_pct            NUMERIC(5,2)   NOT NULL,
    jasa_modal_split_pct        NUMERIC(5,2)   NOT NULL,
    jasa_usaha_split_pct        NUMERIC(5,2)   NOT NULL,
    dana_pengurus_pct           NUMERIC(5,2)   NOT NULL,
    dana_karyawan_pct           NUMERIC(5,2)   NOT NULL,
    dana_pendidikan_pct         NUMERIC(5,2)   NOT NULL,
    dana_sosial_pct             NUMERIC(5,2)   NOT NULL,
    dana_pembangunan_pct        NUMERIC(5,2)   NOT NULL,
    created_at                  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ    NOT NULL DEFAULT now(),

    CONSTRAINT chk_shu_cfg_cadangan
        CHECK (cadangan_pct >= 25.00),
    CONSTRAINT chk_shu_cfg_split
        CHECK (jasa_modal_split_pct + jasa_usaha_split_pct = 100.00),
    CONSTRAINT chk_shu_cfg_total
        CHECK (
            cadangan_pct + jasa_anggota_pct + dana_pengurus_pct
            + dana_karyawan_pct + dana_pendidikan_pct
            + dana_sosial_pct + dana_pembangunan_pct = 100.00
        ),
    CONSTRAINT chk_shu_cfg_pct_nonneg
        CHECK (
            cadangan_pct >= 0 AND jasa_anggota_pct >= 0
            AND jasa_modal_split_pct >= 0 AND jasa_usaha_split_pct >= 0
            AND dana_pengurus_pct >= 0 AND dana_karyawan_pct >= 0
            AND dana_pendidikan_pct >= 0 AND dana_sosial_pct >= 0
            AND dana_pembangunan_pct >= 0
        ),
    CONSTRAINT uq_shu_config UNIQUE (tenant_id, company_id)
);

--------------------------------------------------------------------------------
-- K016-2: shu_periode — Vernon domain
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shu_periode (
    id                          UUID           PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id                   UUID           NOT NULL,
    company_id                  UUID           NOT NULL,
    _rels                       JSONB          NOT NULL DEFAULT '{}',
    _data                       JSONB          NOT NULL DEFAULT '{}',
    _sync_status                TEXT           NOT NULL DEFAULT 'synced',
    _sync_version               BIGINT         NOT NULL DEFAULT 0,
    created_at                  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    deleted_at                  TIMESTAMPTZ,

    tahun_buku                  INT            NOT NULL,
    period_start                DATE           NOT NULL,
    period_end                  DATE           NOT NULL,
    total_pendapatan            NUMERIC(15,2)  NOT NULL DEFAULT 0,
    total_beban                 NUMERIC(15,2)  NOT NULL DEFAULT 0,
    shu_bruto                   NUMERIC(15,2)  NOT NULL DEFAULT 0,
    shu_neto                    NUMERIC(15,2)  NOT NULL DEFAULT 0,
    distribution_config         JSONB          NOT NULL DEFAULT '{}',
    status                      VARCHAR(15)    NOT NULL DEFAULT 'calculated',

    CONSTRAINT chk_shu_periode_status
        CHECK (status IN ('calculated','reviewed','approved','distributed')),
    CONSTRAINT chk_shu_periode_dates
        CHECK (period_end >= period_start),
    CONSTRAINT chk_shu_periode_amounts
        CHECK (
            total_pendapatan >= 0 AND total_beban >= 0
            AND shu_bruto >= 0 AND shu_neto >= 0
        ),
    CONSTRAINT uq_shu_periode_tahun UNIQUE (tenant_id, company_id, tahun_buku)
);

CREATE INDEX IF NOT EXISTS idx_shu_periode_scope
    ON shu_periode (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shu_periode_data
    ON shu_periode USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_shu_periode_sync
    ON shu_periode (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_shu_periode_status
    ON shu_periode (tenant_id, company_id, status) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- K016-3: shu_anggota — Vernon domain
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shu_anggota (
    id                          UUID           PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id                   UUID           NOT NULL,
    company_id                  UUID           NOT NULL,
    _rels                       JSONB          NOT NULL DEFAULT '{}',
    _data                       JSONB          NOT NULL DEFAULT '{}',
    _sync_status                TEXT           NOT NULL DEFAULT 'synced',
    _sync_version               BIGINT         NOT NULL DEFAULT 0,
    created_at                  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    deleted_at                  TIMESTAMPTZ,

    shu_periode_id              UUID           NOT NULL,
    nasabah_id                  UUID           NOT NULL,
    avg_simpanan                NUMERIC(15,2)  NOT NULL DEFAULT 0,
    total_transaksi             NUMERIC(15,2)  NOT NULL DEFAULT 0,
    active_days                 INT            NOT NULL DEFAULT 0,
    jasa_modal                  NUMERIC(15,2)  NOT NULL DEFAULT 0,
    jasa_usaha                  NUMERIC(15,2)  NOT NULL DEFAULT 0,
    total_shu                   NUMERIC(15,2)  NOT NULL DEFAULT 0,
    distribution_method         VARCHAR(20)    NOT NULL DEFAULT 'credit_tabungan',
    target_rekening_id          UUID,

    CONSTRAINT chk_shu_anggota_method
        CHECK (distribution_method IN ('credit_tabungan','separate_payout','pending')),
    CONSTRAINT chk_shu_anggota_amounts
        CHECK (
            avg_simpanan >= 0 AND total_transaksi >= 0
            AND active_days >= 0
            AND jasa_modal >= 0 AND jasa_usaha >= 0
            AND total_shu >= 0
        ),
    CONSTRAINT uq_shu_anggota UNIQUE (tenant_id, shu_periode_id, nasabah_id)
);

CREATE INDEX IF NOT EXISTS idx_shu_anggota_scope
    ON shu_anggota (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shu_anggota_data
    ON shu_anggota USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_shu_anggota_sync
    ON shu_anggota (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_shu_anggota_periode
    ON shu_anggota (shu_periode_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_shu_anggota_nasabah
    ON shu_anggota (nasabah_id) WHERE deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS shu_anggota;
DROP TABLE IF EXISTS shu_periode;
DROP TABLE IF EXISTS shu_config;
