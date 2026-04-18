-- 059: facility_bookings — Facilities & Facility Bookings
-- Vernon pattern: _rels/_data JSONB columns.

--------------------------------------------------------------------------------
-- facilities
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS facilities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Domain data
    name            VARCHAR(100) NOT NULL,
    type            VARCHAR(20) NOT NULL,
    building        VARCHAR(50),
    floor           INT,
    room_number     VARCHAR(20),
    capacity        INT,
    amenities       JSONB DEFAULT '{}',
    is_bookable     BOOLEAN NOT NULL DEFAULT true,
    requires_approval BOOLEAN NOT NULL DEFAULT false,
    approval_role_id UUID,
    operating_hours JSONB,
    booking_advance_min_days INT DEFAULT 1,
    booking_advance_max_days INT DEFAULT 30,
    status          VARCHAR(15) NOT NULL DEFAULT 'active',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_fac_type CHECK (type IN (
        'classroom', 'lab', 'library', 'hall',
        'meeting_room', 'sports_field', 'mosque', 'other'
    )),
    CONSTRAINT chk_fac_status CHECK (status IN (
        'active', 'inactive', 'maintenance'
    )),
    CONSTRAINT chk_fac_booking_advance CHECK (booking_advance_min_days >= 0 AND booking_advance_max_days >= booking_advance_min_days)
);

CREATE INDEX idx_facilities_tenant_company ON facilities (tenant_id, company_id);
CREATE INDEX idx_facilities_type ON facilities (type);
CREATE INDEX idx_facilities_status ON facilities (status);
CREATE INDEX idx_facilities_building ON facilities (building) WHERE building IS NOT NULL;
CREATE INDEX idx_facilities_bookable ON facilities (is_bookable) WHERE is_bookable = true;
CREATE INDEX idx_facilities_rels ON facilities USING GIN (_rels);
CREATE INDEX idx_facilities_data ON facilities USING GIN (_data);

--------------------------------------------------------------------------------
-- facility_bookings
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS facility_bookings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    facility_id     UUID NOT NULL,
    requester_type  VARCHAR(15) NOT NULL,
    requester_id    UUID NOT NULL,
    academic_year_id UUID,
    linked_lab_id   UUID,

    -- Domain data
    booking_date    DATE NOT NULL,
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,
    purpose         TEXT NOT NULL,
    participant_count INT,
    status          VARCHAR(15) NOT NULL DEFAULT 'pending',
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,
    rejection_reason TEXT,
    recurrence_pattern VARCHAR(15),
    recurrence_end_date DATE,
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_fb_requester_type CHECK (requester_type IN (
        'teacher', 'student', 'staff', 'external', 'admin'
    )),
    CONSTRAINT chk_fb_status CHECK (status IN (
        'pending', 'approved', 'rejected', 'cancelled', 'completed'
    )),
    CONSTRAINT chk_fb_recurrence CHECK (recurrence_pattern IS NULL OR recurrence_pattern IN (
        'daily', 'weekly', 'monthly', 'none'
    )),
    CONSTRAINT chk_fb_time_order CHECK (end_time > start_time)
);

CREATE INDEX idx_facility_bookings_tenant_company ON facility_bookings (tenant_id, company_id);
CREATE INDEX idx_fb_facility_date ON facility_bookings (facility_id, booking_date);
CREATE INDEX idx_fb_requester ON facility_bookings (requester_type, requester_id);
CREATE INDEX idx_facility_bookings_status ON facility_bookings (status);
CREATE INDEX idx_facility_bookings_academic_year ON facility_bookings (academic_year_id) WHERE academic_year_id IS NOT NULL;
CREATE INDEX idx_facility_bookings_linked_lab ON facility_bookings (linked_lab_id) WHERE linked_lab_id IS NOT NULL;
CREATE INDEX idx_facility_bookings_rels ON facility_bookings USING GIN (_rels);
CREATE INDEX idx_facility_bookings_data ON facility_bookings USING GIN (_data);
