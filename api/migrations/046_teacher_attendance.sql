-- ADR-S026: Teacher Attendance (Absensi Guru & Staff)
-- Vernon pattern: _rels/_data JSONB columns.

CREATE TABLE IF NOT EXISTS teacher_attendances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Attendance data
    attendance_date DATE NOT NULL,
    status          VARCHAR(20) NOT NULL,

    -- Clock-in/clock-out
    clock_in        TIMESTAMPTZ,
    clock_out       TIMESTAMPTZ,
    late_minutes    INT NOT NULL DEFAULT 0,
    early_leave_minutes INT NOT NULL DEFAULT 0,

    -- Metode clock
    clock_in_method  VARCHAR(20),
    clock_out_method VARCHAR(20),
    clock_in_location  TEXT,
    clock_out_location TEXT,

    -- Keterangan
    note            TEXT,
    attachment_url  TEXT,

    -- Validasi
    validated_by    UUID,
    validated_at    TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_teacher_attendance_date UNIQUE (teacher_id, attendance_date),
    CONSTRAINT chk_teacher_att_status CHECK (status IN (
        'present', 'sick', 'permitted', 'absent',
        'dinas_luar', 'cuti', 'libur'
    )),
    CONSTRAINT chk_clock_in_method CHECK (
        clock_in_method IS NULL OR clock_in_method IN ('fingerprint', 'face_recognition', 'gps', 'manual', 'qr_code')
    ),
    CONSTRAINT chk_clock_out_method CHECK (
        clock_out_method IS NULL OR clock_out_method IN ('fingerprint', 'face_recognition', 'gps', 'manual', 'qr_code')
    ),
    CONSTRAINT chk_late_minutes CHECK (late_minutes >= 0),
    CONSTRAINT chk_early_leave CHECK (early_leave_minutes >= 0)
);

-- Indexes
CREATE INDEX idx_teacher_att_tenant_company ON teacher_attendances (tenant_id, company_id);
CREATE INDEX idx_teacher_att_teacher ON teacher_attendances (teacher_id);
CREATE INDEX idx_teacher_att_date ON teacher_attendances (attendance_date);
CREATE INDEX idx_teacher_att_teacher_month ON teacher_attendances (teacher_id, attendance_date);
CREATE INDEX idx_teacher_att_status ON teacher_attendances (status) WHERE status != 'present';
CREATE INDEX idx_teacher_att_late ON teacher_attendances (late_minutes) WHERE late_minutes > 0;
CREATE INDEX idx_teacher_att_year ON teacher_attendances (academic_year_id);
CREATE INDEX idx_teacher_att_rels ON teacher_attendances USING GIN (_rels);
CREATE INDEX idx_teacher_att_data ON teacher_attendances USING GIN (_data);

-- Konfigurasi jam kerja per sekolah
CREATE TABLE IF NOT EXISTS teacher_attendance_configs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Jam kerja
    work_start_time TIME NOT NULL DEFAULT '07:00',
    work_end_time   TIME NOT NULL DEFAULT '14:00',
    late_tolerance_minutes INT NOT NULL DEFAULT 15,
    minimum_work_hours NUMERIC(4,2) NOT NULL DEFAULT 7.0,

    -- Hari kerja
    work_days       JSONB NOT NULL DEFAULT '["monday","tuesday","wednesday","thursday","friday"]',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_att_config_company UNIQUE (tenant_id, company_id)
);
