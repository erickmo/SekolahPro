-- +migrate Up
-- Domain Vernon: teachers (ADR-012)
-- Guru dan tenaga kependidikan — referenced by 7+ student domains.
-- Satu tabel untuk semua jenis personel, dibedakan oleh field "role" di _data.

CREATE TABLE IF NOT EXISTS teachers (
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
CREATE INDEX IF NOT EXISTS idx_teachers_scope
    ON teachers (tenant_id, company_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_teachers_data
    ON teachers USING GIN (_data)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_teachers_sync
    ON teachers (_sync_status)
    WHERE _sync_status != 'synced' AND deleted_at IS NULL;

-- Domain-specific: NIP partial unique (hanya PNS/P3K yang punya)
CREATE UNIQUE INDEX IF NOT EXISTS uq_teacher_nip
    ON teachers (tenant_id, company_id, ((_data->>'nip')))
    WHERE (_data->>'nip') IS NOT NULL AND deleted_at IS NULL;

-- Domain-specific: NUPTK partial unique
CREATE UNIQUE INDEX IF NOT EXISTS uq_teacher_nuptk
    ON teachers (tenant_id, company_id, ((_data->>'nuptk')))
    WHERE (_data->>'nuptk') IS NOT NULL AND deleted_at IS NULL;

-- Domain-specific: query by role (guru_mapel, guru_bk, admin_tu, dll)
CREATE INDEX IF NOT EXISTS idx_teachers_role
    ON teachers (((_data->>'role')))
    WHERE deleted_at IS NULL;

-- Domain-specific: query by status
CREATE INDEX IF NOT EXISTS idx_teachers_status
    ON teachers (((_data->>'status')))
    WHERE deleted_at IS NULL;

-- Domain-specific: link ke user account
CREATE INDEX IF NOT EXISTS idx_teachers_user_id
    ON teachers (((_data->>'user_id')))
    WHERE (_data->>'user_id') IS NOT NULL AND deleted_at IS NULL;

-- Junction: Guru ↔ Mata Pelajaran (many-to-many, traditional table)
CREATE TABLE IF NOT EXISTS teacher_subject_map (
    id               UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id        UUID        NOT NULL,
    company_id       UUID        NOT NULL,
    teacher_id       UUID        NOT NULL,
    subject_id       UUID        NOT NULL,
    academic_year_id UUID        NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_teacher_subject_year UNIQUE (teacher_id, subject_id, academic_year_id)
);

CREATE INDEX IF NOT EXISTS idx_teacher_subject_teacher
    ON teacher_subject_map (teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_subject_subject
    ON teacher_subject_map (subject_id);
CREATE INDEX IF NOT EXISTS idx_teacher_subject_year
    ON teacher_subject_map (academic_year_id);
CREATE INDEX IF NOT EXISTS idx_teacher_subject_scope
    ON teacher_subject_map (tenant_id, company_id);

-- +migrate Down
DROP TABLE IF EXISTS teacher_subject_map;
DROP TABLE IF EXISTS teachers;
