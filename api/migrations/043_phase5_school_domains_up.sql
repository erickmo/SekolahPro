-- +migrate Up
-- Phase 5: School Domain Enhancements (S019-S025)
-- Curricula, Learning Outcomes, P5 Projects, Subjects, Subject Configurations,
-- Academic Calendar Events, Time Slots, Schedule Entries, Lesson Plans,
-- Lesson Plan Attachments, Teaching Journals, Journal Session Attendances.
--
-- Hybrid Vernon pattern: 10 standard columns + explicit domain-specific columns.
-- All CHECK constraints and UNIQUE constraints are enforced at DB level.
-- Expression indexes on _data for backward compatibility with existing queries.

--------------------------------------------------------------------------------
-- S019: curricula (replaces stub 021)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS curricula (
    id                UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id         UUID        NOT NULL,
    company_id        UUID        NOT NULL,
    _rels             JSONB       NOT NULL DEFAULT '{}',
    _data             JSONB       NOT NULL DEFAULT '{}',
    _sync_status      TEXT        NOT NULL DEFAULT 'synced',
    _sync_version     BIGINT      NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    name              VARCHAR(100) NOT NULL,
    code              VARCHAR(20)  NOT NULL,
    curriculum_type   VARCHAR(20)  NOT NULL,
    academic_year_id  UUID         NOT NULL,
    grade_level       VARCHAR(5)   NOT NULL,
    phase             VARCHAR(10),
    is_active         BOOLEAN      NOT NULL DEFAULT true,
    description       TEXT,

    CONSTRAINT chk_curricula_curriculum_type
        CHECK (curriculum_type IN ('merdeka','k13','ktsp','diniyah','custom')),
    CONSTRAINT chk_curricula_grade_level
        CHECK (grade_level ~ '^[1-9]|1[0-2]$'),
    CONSTRAINT chk_curricula_phase
        CHECK (phase IS NULL OR phase IN ('A','B','C','D','E','F'))
);

CREATE INDEX IF NOT EXISTS idx_curricula_scope
    ON curricula (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_curricula_data
    ON curricula USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_curricula_sync
    ON curricula (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_curricula_code
    ON curricula (tenant_id, company_id, code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_curricula_year_grade_type
    ON curricula (tenant_id, company_id, academic_year_id, grade_level, curriculum_type)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_curricula_academic_year
    ON curricula (academic_year_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_curricula_active
    ON curricula (tenant_id, company_id) WHERE is_active = true AND deleted_at IS NULL;

--------------------------------------------------------------------------------
-- S019: learning_outcomes
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS learning_outcomes (
    id                UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id         UUID        NOT NULL,
    company_id        UUID        NOT NULL,
    _rels             JSONB       NOT NULL DEFAULT '{}',
    _data             JSONB       NOT NULL DEFAULT '{}',
    _sync_status      TEXT        NOT NULL DEFAULT 'synced',
    _sync_version     BIGINT      NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    curriculum_id     UUID         NOT NULL,
    subject_id        UUID,
    parent_id         UUID,
    outcome_type      VARCHAR(10)  NOT NULL,
    code              VARCHAR(30)  NOT NULL,
    title             VARCHAR(255) NOT NULL,
    description       TEXT,
    semester          VARCHAR(10),
    sequence_order    INT          NOT NULL DEFAULT 0,
    ki_number         INT,
    kd_code           VARCHAR(20),
    is_active         BOOLEAN      NOT NULL DEFAULT true,

    CONSTRAINT chk_lo_outcome_type
        CHECK (outcome_type IN ('cp','tp','atp')),
    CONSTRAINT chk_lo_semester
        CHECK (semester IS NULL OR semester IN ('ganjil','genap')),
    CONSTRAINT chk_lo_ki_number
        CHECK (ki_number IS NULL OR ki_number BETWEEN 1 AND 4)
);

CREATE INDEX IF NOT EXISTS idx_learning_outcomes_scope
    ON learning_outcomes (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_learning_outcomes_data
    ON learning_outcomes USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_learning_outcomes_sync
    ON learning_outcomes (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_lo_curriculum_code
    ON learning_outcomes (tenant_id, company_id, curriculum_id, code)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lo_curriculum
    ON learning_outcomes (curriculum_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lo_subject
    ON learning_outcomes (subject_id) WHERE subject_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lo_parent
    ON learning_outcomes (parent_id) WHERE parent_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lo_type
    ON learning_outcomes (outcome_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lo_active
    ON learning_outcomes (tenant_id, company_id) WHERE is_active = true AND deleted_at IS NULL;

--------------------------------------------------------------------------------
-- S019: p5_projects (Projek Penguatan Profil Pelajar Pancasila)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS p5_projects (
    id                UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id         UUID        NOT NULL,
    company_id        UUID        NOT NULL,
    _rels             JSONB       NOT NULL DEFAULT '{}',
    _data             JSONB       NOT NULL DEFAULT '{}',
    _sync_status      TEXT        NOT NULL DEFAULT 'synced',
    _sync_version     BIGINT      NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    curriculum_id     UUID         NOT NULL,
    academic_year_id  UUID         NOT NULL,
    name              VARCHAR(200) NOT NULL,
    theme             VARCHAR(50)  NOT NULL,
    description       TEXT,
    grade_level       VARCHAR(5)   NOT NULL,
    semester          VARCHAR(10)  NOT NULL,
    p5_dimensions     JSONB        NOT NULL DEFAULT '[]',
    start_date        DATE,
    end_date          DATE,
    status            VARCHAR(20)  NOT NULL DEFAULT 'draft',

    CONSTRAINT chk_p5_theme
        CHECK (theme IN (
            'gaya_hidup_berkelanjutan','kearifan_lokal','suara_demokratis',
            'rekayasa_dan_teknologi','kewirausahaan','bhinneka_tunggal_ika',
            'peradaban_indonesia'
        )),
    CONSTRAINT chk_p5_grade_level
        CHECK (grade_level ~ '^[1-9]|1[0-2]$'),
    CONSTRAINT chk_p5_semester
        CHECK (semester IN ('ganjil','genap')),
    CONSTRAINT chk_p5_status
        CHECK (status IN ('draft','active','completed','archived')),
    CONSTRAINT chk_p5_dates
        CHECK (end_date IS NULL OR start_date IS NULL OR start_date <= end_date)
);

CREATE INDEX IF NOT EXISTS idx_p5_projects_scope
    ON p5_projects (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_p5_projects_data
    ON p5_projects USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_p5_projects_sync
    ON p5_projects (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_p5_curriculum
    ON p5_projects (curriculum_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_p5_academic_year
    ON p5_projects (academic_year_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_p5_status
    ON p5_projects (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_p5_grade_semester
    ON p5_projects (tenant_id, company_id, grade_level, semester) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_p5_dimensions
    ON p5_projects USING GIN (p5_dimensions) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- S020: subjects (replaces stub 022)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS subjects (
    id                UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id         UUID        NOT NULL,
    company_id        UUID        NOT NULL,
    _rels             JSONB       NOT NULL DEFAULT '{}',
    _data             JSONB       NOT NULL DEFAULT '{}',
    _sync_status      TEXT        NOT NULL DEFAULT 'synced',
    _sync_version     BIGINT      NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    name              VARCHAR(100) NOT NULL,
    code              VARCHAR(20)  NOT NULL,
    name_en           VARCHAR(100),
    subject_group     VARCHAR(30)  NOT NULL,
    category          VARCHAR(20)  NOT NULL,
    is_national       BOOLEAN      NOT NULL DEFAULT false,
    is_scored         BOOLEAN      NOT NULL DEFAULT true,
    is_active         BOOLEAN      NOT NULL DEFAULT true,
    sort_order        INT          NOT NULL DEFAULT 0,
    description       TEXT,

    CONSTRAINT chk_subjects_group
        CHECK (subject_group IN (
            'agama','pkn','bahasa_indonesia','bahasa_inggris','matematika',
            'ipa','ips','seni_budaya','penjas','pramuka','tik',
            'bahasa_daerah','muatan_lokal','pendidikan_lingkungan',
            'kewirausahaan','lainnya'
        )),
    CONSTRAINT chk_subjects_category
        CHECK (category IN ('normatif','adaptif','produktif','mulok','ekstrakurikuler'))
);

CREATE INDEX IF NOT EXISTS idx_subjects_scope
    ON subjects (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_subjects_data
    ON subjects USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subjects_sync
    ON subjects (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_subjects_code
    ON subjects (tenant_id, company_id, code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subjects_group
    ON subjects (subject_group) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subjects_category
    ON subjects (category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subjects_active
    ON subjects (tenant_id, company_id) WHERE is_active = true AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subjects_sort
    ON subjects (tenant_id, company_id, sort_order) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- S020: subject_configurations
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS subject_configurations (
    id                     UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id              UUID        NOT NULL,
    company_id             UUID        NOT NULL,
    _rels                  JSONB       NOT NULL DEFAULT '{}',
    _data                  JSONB       NOT NULL DEFAULT '{}',
    _sync_status           TEXT        NOT NULL DEFAULT 'synced',
    _sync_version          BIGINT      NOT NULL DEFAULT 0,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ,

    subject_id             UUID         NOT NULL,
    academic_year_id       UUID         NOT NULL,
    curriculum_id          UUID,
    grade_level            VARCHAR(5)   NOT NULL,
    credit_hours_per_week  INT          NOT NULL DEFAULT 2,
    weight_knowledge       INT          NOT NULL DEFAULT 50,
    weight_skill           INT          NOT NULL DEFAULT 50,
    passing_grade          NUMERIC(5,2) NOT NULL DEFAULT 70.00,
    is_active              BOOLEAN      NOT NULL DEFAULT true,

    CONSTRAINT chk_subj_config_grade
        CHECK (grade_level ~ '^[1-9]|1[0-2]$'),
    CONSTRAINT chk_subj_config_credit_hours
        CHECK (credit_hours_per_week BETWEEN 1 AND 12),
    CONSTRAINT chk_subj_config_weights
        CHECK (weight_knowledge + weight_skill = 100)
);

CREATE INDEX IF NOT EXISTS idx_subject_configurations_scope
    ON subject_configurations (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_subject_configurations_data
    ON subject_configurations USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subject_configurations_sync
    ON subject_configurations (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_subj_config_unique
    ON subject_configurations (subject_id, academic_year_id, grade_level)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subj_config_subject
    ON subject_configurations (subject_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subj_config_academic_year
    ON subject_configurations (academic_year_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subj_config_curriculum
    ON subject_configurations (curriculum_id) WHERE curriculum_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subj_config_grade
    ON subject_configurations (tenant_id, company_id, grade_level) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- S023: academic_calendar_events (replaces stub 023)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS academic_calendar_events (
    id                     UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id              UUID        NOT NULL,
    company_id             UUID        NOT NULL,
    _rels                  JSONB       NOT NULL DEFAULT '{}',
    _data                  JSONB       NOT NULL DEFAULT '{}',
    _sync_status           TEXT        NOT NULL DEFAULT 'synced',
    _sync_version          BIGINT      NOT NULL DEFAULT 0,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ,

    academic_year_id       UUID         NOT NULL,
    name                   VARCHAR(200) NOT NULL,
    description            TEXT,
    event_type             VARCHAR(30)  NOT NULL,
    start_date             DATE         NOT NULL,
    end_date               DATE         NOT NULL,
    semester               VARCHAR(10),
    is_school_day          BOOLEAN      NOT NULL DEFAULT false,
    affects_attendance     BOOLEAN      NOT NULL DEFAULT true,
    is_recurring_yearly    BOOLEAN      NOT NULL DEFAULT false,
    hijri_date             VARCHAR(30),
    islamic_event_type     VARCHAR(30),
    color_code             VARCHAR(7),
    sort_order             INT          NOT NULL DEFAULT 0,

    CONSTRAINT chk_cal_event_type
        CHECK (event_type IN (
            'semester_start','semester_end','mid_exam','final_exam',
            'national_holiday','religious_holiday','school_event',
            'teacher_day','student_day','parent_day','other'
        )),
    CONSTRAINT chk_cal_dates
        CHECK (start_date <= end_date),
    CONSTRAINT chk_cal_semester
        CHECK (semester IS NULL OR semester IN ('ganjil','genap')),
    CONSTRAINT chk_cal_islamic_event
        CHECK (islamic_event_type IS NULL OR islamic_event_type IN (
            'ramadhan','idul_fitri','idul_adha','isra_miraj',
            'maulid_nabi','tahun_baru_islam','nuzulul_quran',
            'hari_suci','tarawih','other_islamic'
        ))
);

CREATE INDEX IF NOT EXISTS idx_academic_cal_events_scope
    ON academic_calendar_events (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_academic_cal_events_data
    ON academic_calendar_events USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_academic_cal_events_sync
    ON academic_calendar_events (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_cal_academic_year
    ON academic_calendar_events (academic_year_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cal_event_type
    ON academic_calendar_events (event_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cal_dates
    ON academic_calendar_events (tenant_id, company_id, start_date, end_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cal_semester
    ON academic_calendar_events (semester) WHERE semester IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cal_school_day
    ON academic_calendar_events (tenant_id, company_id) WHERE is_school_day = true AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cal_recurring
    ON academic_calendar_events (tenant_id, company_id) WHERE is_recurring_yearly = true AND deleted_at IS NULL;

--------------------------------------------------------------------------------
-- S021: time_slots
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS time_slots (
    id                UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id         UUID        NOT NULL,
    company_id        UUID        NOT NULL,
    _rels             JSONB       NOT NULL DEFAULT '{}',
    _data             JSONB       NOT NULL DEFAULT '{}',
    _sync_status      TEXT        NOT NULL DEFAULT 'synced',
    _sync_version     BIGINT      NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    name              VARCHAR(30) NOT NULL,
    slot_number       INT         NOT NULL,
    slot_type         VARCHAR(20) NOT NULL,
    day_of_week       INT         NOT NULL,
    start_time        TIME        NOT NULL,
    end_time          TIME        NOT NULL,
    academic_year_id  UUID        NOT NULL,
    semester          VARCHAR(10) NOT NULL,

    CONSTRAINT chk_ts_slot_number
        CHECK (slot_number BETWEEN 1 AND 15),
    CONSTRAINT chk_ts_slot_type
        CHECK (slot_type IN ('lesson','break','assembly','prayer')),
    CONSTRAINT chk_ts_day_of_week
        CHECK (day_of_week BETWEEN 1 AND 7),
    CONSTRAINT chk_ts_times
        CHECK (start_time < end_time),
    CONSTRAINT chk_ts_semester
        CHECK (semester IN ('ganjil','genap'))
);

CREATE INDEX IF NOT EXISTS idx_time_slots_scope
    ON time_slots (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_time_slots_data
    ON time_slots USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_time_slots_sync
    ON time_slots (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_time_slots_unique
    ON time_slots (tenant_id, company_id, academic_year_id, semester, day_of_week, slot_number)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ts_academic_year
    ON time_slots (academic_year_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ts_day_slot
    ON time_slots (tenant_id, company_id, day_of_week, slot_number) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- S021: schedule_entries (replaces stub 024)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS schedule_entries (
    id                    UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id             UUID        NOT NULL,
    company_id            UUID        NOT NULL,
    _rels                 JSONB       NOT NULL DEFAULT '{}',
    _data                 JSONB       NOT NULL DEFAULT '{}',
    _sync_status          TEXT        NOT NULL DEFAULT 'synced',
    _sync_version         BIGINT      NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at            TIMESTAMPTZ,

    time_slot_id          UUID        NOT NULL,
    class_room_id         UUID        NOT NULL,
    subject_id            UUID        NOT NULL,
    teacher_id            UUID        NOT NULL,
    academic_year_id      UUID        NOT NULL,
    room_name             VARCHAR(50),
    semester              VARCHAR(10) NOT NULL,
    is_substitution       BOOLEAN     NOT NULL DEFAULT false,
    original_teacher_id   UUID,
    substitution_date     DATE,
    substitution_reason   TEXT,

    CONSTRAINT chk_se_semester
        CHECK (semester IN ('ganjil','genap'))
);

CREATE INDEX IF NOT EXISTS idx_schedule_entries_scope
    ON schedule_entries (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_schedule_entries_data
    ON schedule_entries USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_schedule_entries_sync
    ON schedule_entries (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_se_time_slot
    ON schedule_entries (time_slot_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_se_class_room
    ON schedule_entries (class_room_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_se_subject
    ON schedule_entries (subject_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_se_teacher
    ON schedule_entries (teacher_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_se_academic_year
    ON schedule_entries (academic_year_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_se_semester
    ON schedule_entries (semester) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_se_substitution
    ON schedule_entries (tenant_id, company_id) WHERE is_substitution = true AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_se_substitution_date
    ON schedule_entries (substitution_date) WHERE substitution_date IS NOT NULL AND deleted_at IS NULL;

--------------------------------------------------------------------------------
-- S024: lesson_plans (replaces stub 025)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lesson_plans (
    id                       UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id                UUID        NOT NULL,
    company_id               UUID        NOT NULL,
    _rels                    JSONB       NOT NULL DEFAULT '{}',
    _data                    JSONB       NOT NULL DEFAULT '{}',
    _sync_status             TEXT        NOT NULL DEFAULT 'synced',
    _sync_version            BIGINT      NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at               TIMESTAMPTZ,

    teacher_id               UUID         NOT NULL,
    subject_id               UUID         NOT NULL,
    class_room_id            UUID         NOT NULL,
    academic_year_id         UUID         NOT NULL,
    curriculum_id            UUID,
    title                    VARCHAR(300) NOT NULL,
    plan_type                VARCHAR(20)  NOT NULL,
    semester                 VARCHAR(10)  NOT NULL,
    meeting_number           INT,
    topic                    VARCHAR(300) NOT NULL,
    subtopic                 TEXT,
    duration_minutes         INT          NOT NULL DEFAULT 90,
    learning_objectives      TEXT         NOT NULL,
    learning_activities      TEXT         NOT NULL,
    assessment_plan          TEXT,
    teaching_methods         TEXT,
    media_and_resources      TEXT,
    differentiation_notes    TEXT,
    core_competency          TEXT,
    basic_competency         TEXT,
    indicators               TEXT,
    pancasila_profile        TEXT,
    trigger_questions        TEXT,
    reflection               TEXT,
    kitab_reference          VARCHAR(200),
    bab_fashl                VARCHAR(200),
    teaching_method_pesantren VARCHAR(30),
    hafalan_target           TEXT,
    status                   VARCHAR(20)  NOT NULL DEFAULT 'draft',
    submitted_at             TIMESTAMPTZ,
    reviewed_by              UUID,
    reviewed_at              TIMESTAMPTZ,
    review_notes             TEXT,

    CONSTRAINT chk_lp_plan_type
        CHECK (plan_type IN ('daily','weekly','semester','annual')),
    CONSTRAINT chk_lp_semester
        CHECK (semester IN ('ganjil','genap')),
    CONSTRAINT chk_lp_meeting_number
        CHECK (meeting_number IS NULL OR meeting_number BETWEEN 1 AND 100),
    CONSTRAINT chk_lp_duration
        CHECK (duration_minutes BETWEEN 1 AND 480),
    CONSTRAINT chk_lp_status
        CHECK (status IN ('draft','submitted','reviewed','approved','revision','archived')),
    CONSTRAINT chk_lp_method_pesantren
        CHECK (teaching_method_pesantren IS NULL OR teaching_method_pesantren IN (
            'sorogan','bandongan','halaqah','mudzakarah','murasalah',
            'tahfidz','tadabbur','talaqqi'
        ))
);

CREATE INDEX IF NOT EXISTS idx_lesson_plans_scope
    ON lesson_plans (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_lesson_plans_data
    ON lesson_plans USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lesson_plans_sync
    ON lesson_plans (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lp_teacher
    ON lesson_plans (teacher_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lp_subject
    ON lesson_plans (subject_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lp_class_room
    ON lesson_plans (class_room_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lp_academic_year
    ON lesson_plans (academic_year_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lp_curriculum
    ON lesson_plans (curriculum_id) WHERE curriculum_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lp_status
    ON lesson_plans (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lp_plan_type
    ON lesson_plans (plan_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lp_semester
    ON lesson_plans (semester) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lp_teacher_subject_class
    ON lesson_plans (teacher_id, subject_id, class_room_id) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- S024: lesson_plan_attachments
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lesson_plan_attachments (
    id                UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id         UUID        NOT NULL,
    company_id        UUID        NOT NULL,
    _rels             JSONB       NOT NULL DEFAULT '{}',
    _data             JSONB       NOT NULL DEFAULT '{}',
    _sync_status      TEXT        NOT NULL DEFAULT 'synced',
    _sync_version     BIGINT      NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    lesson_plan_id    UUID         NOT NULL,
    file_name         VARCHAR(255) NOT NULL,
    file_path         TEXT         NOT NULL,
    file_size_bytes   BIGINT       NOT NULL,
    mime_type         VARCHAR(100) NOT NULL,
    file_type         VARCHAR(20)  NOT NULL DEFAULT 'document',
    description       VARCHAR(300),
    uploaded_by       UUID         NOT NULL,

    CONSTRAINT chk_lpa_file_size
        CHECK (file_size_bytes BETWEEN 1 AND 52428800),
    CONSTRAINT chk_lpa_file_type
        CHECK (file_type IN ('document','image','video','audio','other'))
);

CREATE INDEX IF NOT EXISTS idx_lesson_plan_attachments_scope
    ON lesson_plan_attachments (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_lesson_plan_attachments_data
    ON lesson_plan_attachments USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lesson_plan_attachments_sync
    ON lesson_plan_attachments (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lpa_lesson_plan
    ON lesson_plan_attachments (lesson_plan_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lpa_uploaded_by
    ON lesson_plan_attachments (uploaded_by) WHERE deleted_at IS NULL;

--------------------------------------------------------------------------------
-- S025: teaching_journals (replaces stub 026)
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teaching_journals (
    id                         UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id                  UUID        NOT NULL,
    company_id                 UUID        NOT NULL,
    _rels                      JSONB       NOT NULL DEFAULT '{}',
    _data                      JSONB       NOT NULL DEFAULT '{}',
    _sync_status               TEXT        NOT NULL DEFAULT 'synced',
    _sync_version              BIGINT      NOT NULL DEFAULT 0,
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                 TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                 TIMESTAMPTZ,

    teacher_id                 UUID         NOT NULL,
    subject_id                 UUID         NOT NULL,
    class_room_id              UUID         NOT NULL,
    academic_year_id           UUID         NOT NULL,
    schedule_entry_id          UUID,
    lesson_plan_id             UUID,
    journal_date               DATE         NOT NULL,
    semester                   VARCHAR(10)  NOT NULL,
    slot_start                 INT          NOT NULL,
    slot_end                   INT          NOT NULL,
    start_time                 TIME,
    end_time                   TIME,
    topic_taught               VARCHAR(300) NOT NULL,
    material_detail            TEXT,
    teaching_method            VARCHAR(50),
    learning_activities_summary TEXT,
    student_responses          TEXT,
    obstacles_notes            TEXT,
    follow_up_plan             TEXT,
    kitab_reference            VARCHAR(200),
    kitab_page_from            INT,
    kitab_page_to              INT,
    teaching_method_pesantren  VARCHAR(30),
    hafalan_progress           TEXT,
    total_students             INT          NOT NULL DEFAULT 0,
    present_count              INT          NOT NULL DEFAULT 0,
    absent_count               INT          NOT NULL DEFAULT 0,
    late_count                 INT          NOT NULL DEFAULT 0,
    permission_count           INT          NOT NULL DEFAULT 0,
    status                     VARCHAR(20)  NOT NULL DEFAULT 'draft',

    CONSTRAINT chk_tj_semester
        CHECK (semester IN ('ganjil','genap')),
    CONSTRAINT chk_tj_slot_start
        CHECK (slot_start BETWEEN 1 AND 15),
    CONSTRAINT chk_tj_slot_end
        CHECK (slot_end BETWEEN 1 AND 15),
    CONSTRAINT chk_tj_slot_range
        CHECK (slot_start <= slot_end),
    CONSTRAINT chk_tj_status
        CHECK (status IN ('draft','submitted','verified')),
    CONSTRAINT chk_tj_method_pesantren
        CHECK (teaching_method_pesantren IS NULL OR teaching_method_pesantren IN (
            'sorogan','bandongan','halaqah','mudzakarah','murasalah',
            'tahfidz','tadabbur','talaqqi'
        ))
);

CREATE INDEX IF NOT EXISTS idx_teaching_journals_scope
    ON teaching_journals (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_teaching_journals_data
    ON teaching_journals USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_teaching_journals_sync
    ON teaching_journals (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_tj_unique
    ON teaching_journals (teacher_id, class_room_id, journal_date, slot_start)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tj_teacher
    ON teaching_journals (teacher_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tj_subject
    ON teaching_journals (subject_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tj_class_room
    ON teaching_journals (class_room_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tj_academic_year
    ON teaching_journals (academic_year_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tj_date
    ON teaching_journals (journal_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tj_status
    ON teaching_journals (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tj_schedule_entry
    ON teaching_journals (schedule_entry_id) WHERE schedule_entry_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tj_lesson_plan
    ON teaching_journals (lesson_plan_id) WHERE lesson_plan_id IS NOT NULL AND deleted_at IS NULL;

--------------------------------------------------------------------------------
-- S025: journal_session_attendances
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS journal_session_attendances (
    id                 UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id          UUID        NOT NULL,
    company_id         UUID        NOT NULL,
    _rels              JSONB       NOT NULL DEFAULT '{}',
    _data              JSONB       NOT NULL DEFAULT '{}',
    _sync_status       TEXT        NOT NULL DEFAULT 'synced',
    _sync_version      BIGINT      NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ,

    journal_id         UUID         NOT NULL,
    student_id         UUID         NOT NULL,
    attendance_status  VARCHAR(15)  NOT NULL,
    late_minutes       INT,
    notes              TEXT,

    CONSTRAINT chk_jsa_status
        CHECK (attendance_status IN ('present','absent','sick','permission','late','skip')),
    CONSTRAINT chk_jsa_late_minutes
        CHECK (late_minutes IS NULL OR late_minutes BETWEEN 1 AND 120)
);

CREATE INDEX IF NOT EXISTS idx_journal_session_attendances_scope
    ON journal_session_attendances (tenant_id, company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_journal_session_attendances_data
    ON journal_session_attendances USING GIN (_data) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_journal_session_attendances_sync
    ON journal_session_attendances (_sync_status) WHERE _sync_status != 'synced' AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_jsa_journal_student
    ON journal_session_attendances (journal_id, student_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jsa_journal
    ON journal_session_attendances (journal_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jsa_student
    ON journal_session_attendances (student_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jsa_status
    ON journal_session_attendances (attendance_status) WHERE deleted_at IS NULL;
