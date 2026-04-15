-- +migrate Up
-- Domain Vernon: class_rooms (ADR-011)
-- Kelas — unit organisasi utama, di-scope per tahun ajaran.
-- Autoload: academic_year + homeroom_teacher.

CREATE TABLE IF NOT EXISTS class_rooms (
    id            UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id     UUID        NOT NULL,
    company_id    UUID        NOT NULL,
    _rels         JSONB       NOT NULL DEFAULT '{}',
    _data         JSONB       NOT NULL DEFAULT '{}',
    _sync_status  TEXT        NOT NULL DEFAULT 'synced',
    _sync_version BIGINT      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

-- Standard Vernon indexes
CREATE INDEX IF NOT EXISTS idx_class_rooms_scope
    ON class_rooms (tenant_id, company_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_class_rooms_data
    ON class_rooms USING GIN (_data)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_class_rooms_sync
    ON class_rooms (_sync_status)
    WHERE _sync_status != 'synced' AND deleted_at IS NULL;

-- Domain-specific: unique name per tahun ajaran per company
CREATE UNIQUE INDEX IF NOT EXISTS uq_class_room_name_year
    ON class_rooms (
        tenant_id, company_id,
        ((_data->>'academic_year_id')),
        ((_data->>'name'))
    )
    WHERE deleted_at IS NULL;

-- Domain-specific: query by academic year
CREATE INDEX IF NOT EXISTS idx_class_rooms_year
    ON class_rooms (((_data->>'academic_year_id')))
    WHERE deleted_at IS NULL;

-- Domain-specific: query by grade level
CREATE INDEX IF NOT EXISTS idx_class_rooms_grade
    ON class_rooms (((_data->>'grade_level')))
    WHERE deleted_at IS NULL;

-- Domain-specific: query by homeroom teacher
CREATE INDEX IF NOT EXISTS idx_class_rooms_teacher
    ON class_rooms (((_data->>'homeroom_teacher_id')))
    WHERE (_data->>'homeroom_teacher_id') IS NOT NULL AND deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS class_rooms;
