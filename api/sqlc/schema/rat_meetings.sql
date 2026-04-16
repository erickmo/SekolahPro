-- Schema untuk RAT Meeting Management (ADR-K026: Rapat Anggota Tahunan).
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.

-- Meeting Type ENUM
CREATE TYPE rat_meeting_type AS ENUM ('annual', 'special', 'joint', 'dissolution');

-- Meeting Mode ENUM
CREATE TYPE rat_meeting_mode AS ENUM ('offline', 'online', 'hybrid');

-- Meeting Status ENUM
CREATE TYPE rat_meeting_status AS ENUM ('draft', 'invitation_sent', 'in_progress', 'completed', 'cancelled', 'rescheduled');

-- Agenda Category ENUM
CREATE TYPE rat_agenda_category AS ENUM ('opening', 'report', 'election', 'shu', 'ad_art_change', 'budget', 'other', 'closing');

-- Proposal Type ENUM
CREATE TYPE rat_proposal_type AS ENUM ('information', 'discussion', 'voting', 'election');

-- Agenda Result ENUM
CREATE TYPE rat_agenda_result AS ENUM ('approved', 'rejected', 'deferred', 'no_vote');

-- Follow-up Status ENUM
CREATE TYPE rat_followup_status AS ENUM ('pending', 'in_progress', 'completed');

-- Attendance Type ENUM
CREATE TYPE rat_attendance_type AS ENUM ('present', 'proxy', 'absent_apology', 'absent_no_info');

-- Voting Right ENUM
CREATE TYPE rat_voting_right AS ENUM ('full', 'limited', 'none');

-- Vote Option ENUM
CREATE TYPE rat_vote_option AS ENUM ('for', 'against', 'abstain');

-- Vote Method ENUM
CREATE TYPE rat_vote_method AS ENUM ('show_of_hands', 'secret_ballot', 'electronic');

-- Election Status ENUM
CREATE TYPE rat_election_status AS ENUM ('nomination', 'voting', 'completed', 'failed');

-- Election Method ENUM
CREATE TYPE rat_election_method AS ENUM ('open', 'secret_ballot', 'acclamation');

-- Candidate Status ENUM
CREATE TYPE rat_candidate_status AS ENUM ('proposed', 'accepted', 'withdrawn', 'elected', 'not_elected');

-- ── RAT Meeting ──────────────────────────────────────────────────────────────
CREATE TABLE rat_meetings (
    id                       UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                UUID              NOT NULL,
    company_id               UUID              NOT NULL,

    -- Meeting Info
    meeting_type             rat_meeting_type  NOT NULL,
    meeting_number           VARCHAR(50)       NOT NULL,
    title                    VARCHAR(255)      NOT NULL,
    description              TEXT              DEFAULT '',

    -- Schedule
    meeting_date             DATE              NOT NULL,
    meeting_time             TIME              NOT NULL,
    meeting_location         VARCHAR(500)      NOT NULL DEFAULT '',
    meeting_mode             rat_meeting_mode  NOT NULL DEFAULT 'offline',
    online_link              VARCHAR(500),

    -- Quorum
    total_eligible_members   INT               NOT NULL DEFAULT 0,
    quorum_required          INT               NOT NULL DEFAULT 0,
    actual_attendees         INT               NOT NULL DEFAULT 0,
    quorum_met               BOOLEAN           NOT NULL DEFAULT FALSE,

    -- Status
    status                   rat_meeting_status NOT NULL DEFAULT 'draft',
    cancelled_reason         TEXT,
    rescheduled_to_id        UUID REFERENCES rat_meetings(id),

    -- Conveners
    convened_by              UUID              NOT NULL,
    secretary_id             UUID              NOT NULL,

    -- Audit
    created_at               TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ
);

CREATE INDEX idx_rat_meetings_tenant_company ON rat_meetings (tenant_id, company_id);
CREATE INDEX idx_rat_meetings_status ON rat_meetings (tenant_id, company_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_rat_meetings_type ON rat_meetings (tenant_id, company_id, meeting_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_rat_meetings_date ON rat_meetings (tenant_id, company_id, meeting_date) WHERE deleted_at IS NULL;

-- ── Agenda Item ──────────────────────────────────────────────────────────────
CREATE TABLE rat_agenda_items (
    id                       UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                UUID                NOT NULL,
    company_id               UUID                NOT NULL,
    meeting_id               UUID                NOT NULL REFERENCES rat_meetings(id),

    -- Agenda
    agenda_number            INT                 NOT NULL DEFAULT 1,
    category                 rat_agenda_category NOT NULL,
    title                    VARCHAR(255)        NOT NULL,
    description              TEXT                DEFAULT '',

    -- Proposal
    proposal_type            rat_proposal_type   NOT NULL,
    proposer_id              UUID,
    proposal_document_id     UUID,

    -- Discussion
    discussion_notes         TEXT,
    discussion_duration      INT,

    -- Result
    result                   rat_agenda_result,
    result_notes             TEXT,
    vote_for                 INT,
    vote_against             INT,
    vote_abstain             INT,

    -- Follow-up
    followup_action          TEXT,
    followup_deadline        DATE,
    followup_assignee_id     UUID,
    followup_status          rat_followup_status,

    -- Audit
    created_at               TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ
);

CREATE INDEX idx_rat_agenda_items_tenant_company ON rat_agenda_items (tenant_id, company_id);
CREATE INDEX idx_rat_agenda_items_meeting ON rat_agenda_items (tenant_id, company_id, meeting_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_rat_agenda_items_category ON rat_agenda_items (tenant_id, company_id, category) WHERE deleted_at IS NULL;

-- ── Attendance ───────────────────────────────────────────────────────────────
CREATE TABLE rat_attendances (
    id                       UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                UUID                NOT NULL,
    company_id               UUID                NOT NULL,
    meeting_id               UUID                NOT NULL REFERENCES rat_meetings(id),
    nasabah_id               UUID                NOT NULL,

    -- Attendance
    attendance_type          rat_attendance_type NOT NULL DEFAULT 'present',
    proxy_for_id             UUID,
    proxy_document_id        UUID,

    -- Voting Rights
    voting_right             rat_voting_right    NOT NULL DEFAULT 'full',

    -- Check-in
    checked_in_at            TIMESTAMPTZ,
    checked_in_by            UUID,

    -- Audit
    created_at               TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rat_attendances_tenant_company ON rat_attendances (tenant_id, company_id);
CREATE INDEX idx_rat_attendances_meeting ON rat_attendances (tenant_id, company_id, meeting_id);
CREATE UNIQUE INDEX idx_rat_attendances_unique ON rat_attendances (tenant_id, company_id, meeting_id, nasabah_id);

-- ── Vote ─────────────────────────────────────────────────────────────────────
CREATE TABLE rat_votes (
    id                       UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                UUID              NOT NULL,
    company_id               UUID              NOT NULL,
    meeting_id               UUID              NOT NULL REFERENCES rat_meetings(id),
    agenda_item_id           UUID              NOT NULL REFERENCES rat_agenda_items(id),
    voter_id                 UUID              NOT NULL,

    -- Vote
    vote                     rat_vote_option   NOT NULL,
    vote_method              rat_vote_method   NOT NULL DEFAULT 'show_of_hands',

    -- Audit
    voted_at                 TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    created_at               TIMESTAMPTZ       NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rat_votes_tenant_company ON rat_votes (tenant_id, company_id);
CREATE INDEX idx_rat_votes_agenda ON rat_votes (tenant_id, company_id, agenda_item_id);
CREATE UNIQUE INDEX idx_rat_votes_unique ON rat_votes (tenant_id, company_id, agenda_item_id, voter_id);

-- ── Election ─────────────────────────────────────────────────────────────────
CREATE TABLE rat_elections (
    id                       UUID                 PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                UUID                 NOT NULL,
    company_id               UUID                 NOT NULL,
    meeting_id               UUID                 NOT NULL REFERENCES rat_meetings(id),
    agenda_item_id           UUID                 NOT NULL REFERENCES rat_agenda_items(id),

    -- Election Info
    position_type            VARCHAR(30)          NOT NULL,
    term_start               DATE                 NOT NULL,
    term_end                 DATE                 NOT NULL,
    vacancies                INT                  NOT NULL DEFAULT 1,

    -- Status
    status                   rat_election_status  NOT NULL DEFAULT 'nomination',
    election_method          rat_election_method  NOT NULL DEFAULT 'open',

    -- Audit
    created_at               TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ
);

CREATE INDEX idx_rat_elections_tenant_company ON rat_elections (tenant_id, company_id);
CREATE INDEX idx_rat_elections_meeting ON rat_elections (tenant_id, company_id, meeting_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_rat_elections_status ON rat_elections (tenant_id, company_id, status) WHERE deleted_at IS NULL;

-- ── Election Candidate ───────────────────────────────────────────────────────
CREATE TABLE rat_election_candidates (
    id                       UUID                  PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                UUID                  NOT NULL,
    company_id               UUID                  NOT NULL,
    election_id              UUID                  NOT NULL REFERENCES rat_elections(id),
    nasabah_id               UUID                  NOT NULL,

    -- Nomination
    nominated_by_id          UUID                  NOT NULL,
    seconded_by_id           UUID                  NOT NULL,
    nomination_statement     TEXT,

    -- Status
    status                   rat_candidate_status  NOT NULL DEFAULT 'proposed',

    -- Results
    vote_count               INT,
    rank                     INT,

    -- Audit
    created_at               TIMESTAMPTZ           NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rat_election_candidates_tenant_company ON rat_election_candidates (tenant_id, company_id);
CREATE INDEX idx_rat_election_candidates_election ON rat_election_candidates (tenant_id, company_id, election_id);
CREATE UNIQUE INDEX idx_rat_election_candidates_unique ON rat_election_candidates (tenant_id, company_id, election_id, nasabah_id);
