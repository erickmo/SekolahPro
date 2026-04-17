-- ADR-S027: Teacher Workload (Beban Mengajar)
-- Vernon pattern: _rels/_data JSONB columns.

-- Ringkasan beban mengajar per guru per semester
CREATE TABLE IF NOT EXISTS teacher_workloads (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Periode
    semester        VARCHAR(10) NOT NULL,

    -- Ringkasan (dalam jam pelajaran/JP)
    teaching_hours  NUMERIC(5,1) NOT NULL DEFAULT 0,
    additional_hours NUMERIC(5,1) NOT NULL DEFAULT 0,
    total_hours     NUMERIC(5,1) NOT NULL DEFAULT 0,

    -- Status pemenuhan
    minimum_required NUMERIC(5,1) NOT NULL DEFAULT 24,
    is_fulfilled    BOOLEAN NOT NULL DEFAULT false,
    fulfillment_status VARCHAR(20) NOT NULL DEFAULT 'kurang',

    -- Dapodik reporting
    dapodik_reported BOOLEAN NOT NULL DEFAULT false,
    dapodik_reported_at TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_workload_teacher_semester UNIQUE (teacher_id, academic_year_id, semester),
    CONSTRAINT chk_workload_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_fulfillment_status CHECK (fulfillment_status IN ('kurang', 'terpenuhi', 'lebih')),
    CONSTRAINT chk_teaching_hours CHECK (teaching_hours >= 0),
    CONSTRAINT chk_additional_hours CHECK (additional_hours >= 0),
    CONSTRAINT chk_total_hours CHECK (total_hours >= 0)
);

-- Indexes
CREATE INDEX idx_workload_tenant_company ON teacher_workloads (tenant_id, company_id);
CREATE INDEX idx_workload_teacher ON teacher_workloads (teacher_id);
CREATE INDEX idx_workload_year_semester ON teacher_workloads (academic_year_id, semester);
CREATE INDEX idx_workload_fulfilled ON teacher_workloads (is_fulfilled) WHERE is_fulfilled = false;
CREATE INDEX idx_workload_rels ON teacher_workloads USING GIN (_rels);
CREATE INDEX idx_workload_data ON teacher_workloads USING GIN (_data);

-- Detail komponen beban mengajar
CREATE TABLE IF NOT EXISTS teacher_workload_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    workload_id     UUID NOT NULL,
    teacher_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Komponen beban
    item_type       VARCHAR(30) NOT NULL,
    description     TEXT NOT NULL,
    hours_per_week  NUMERIC(5,1) NOT NULL,

    -- Reference (opsional)
    reference_type  VARCHAR(30),
    reference_id    UUID,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_item_type CHECK (item_type IN (
        'mengajar', 'wali_kelas', 'pembina_ekskul',
        'guru_bk', 'kepala_sekolah', 'wakil_kepsek',
        'koordinator', 'panitia', 'tugas_tambahan'
    )),
    CONSTRAINT chk_item_hours CHECK (hours_per_week >= 0 AND hours_per_week <= 40),
    CONSTRAINT chk_reference_type CHECK (
        reference_type IS NULL OR reference_type IN (
            'timetable_slot', 'class_room', 'extracurricular', 'sk_internal'
        )
    )
);

-- Indexes
CREATE INDEX idx_workload_item_tenant_company ON teacher_workload_items (tenant_id, company_id);
CREATE INDEX idx_workload_item_workload ON teacher_workload_items (workload_id);
CREATE INDEX idx_workload_item_teacher ON teacher_workload_items (teacher_id);
CREATE INDEX idx_workload_item_type ON teacher_workload_items (item_type);
CREATE INDEX idx_workload_item_rels ON teacher_workload_items USING GIN (_rels);
CREATE INDEX idx_workload_item_data ON teacher_workload_items USING GIN (_data);
