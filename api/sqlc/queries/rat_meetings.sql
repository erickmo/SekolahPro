-- Queries untuk RAT Meeting Management (ADR-K026).
-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- ── RAT Meeting ──────────────────────────────────────────────────────────────

-- name: GetRATMeetingByID :one
SELECT id, tenant_id, company_id, meeting_type, meeting_number, title, description,
       meeting_date, meeting_time, meeting_location, meeting_mode, online_link,
       total_eligible_members, quorum_required, actual_attendees, quorum_met,
       status, cancelled_reason, rescheduled_to_id, convened_by, secretary_id,
       created_at, updated_at, deleted_at
FROM rat_meetings
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListRATMeetings :many
SELECT id, tenant_id, company_id, meeting_type, meeting_number, title, description,
       meeting_date, meeting_time, meeting_location, meeting_mode, online_link,
       total_eligible_members, quorum_required, actual_attendees, quorum_met,
       status, cancelled_reason, rescheduled_to_id, convened_by, secretary_id,
       created_at, updated_at
FROM rat_meetings
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListRATMeetingsByType :many
SELECT id, tenant_id, company_id, meeting_type, meeting_number, title, description,
       meeting_date, meeting_time, meeting_location, meeting_mode, online_link,
       total_eligible_members, quorum_required, actual_attendees, quorum_met,
       status, cancelled_reason, rescheduled_to_id, convened_by, secretary_id,
       created_at, updated_at
FROM rat_meetings
WHERE tenant_id = $1 AND company_id = $2 AND meeting_type = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListRATMeetingsByStatus :many
SELECT id, tenant_id, company_id, meeting_type, meeting_number, title, description,
       meeting_date, meeting_time, meeting_location, meeting_mode, online_link,
       total_eligible_members, quorum_required, actual_attendees, quorum_met,
       status, cancelled_reason, rescheduled_to_id, convened_by, secretary_id,
       created_at, updated_at
FROM rat_meetings
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListRATMeetingsByTypeAndStatus :many
SELECT id, tenant_id, company_id, meeting_type, meeting_number, title, description,
       meeting_date, meeting_time, meeting_location, meeting_mode, online_link,
       total_eligible_members, quorum_required, actual_attendees, quorum_met,
       status, cancelled_reason, rescheduled_to_id, convened_by, secretary_id,
       created_at, updated_at
FROM rat_meetings
WHERE tenant_id = $1 AND company_id = $2 AND meeting_type = $3 AND status = $4 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountRATMeetings :one
SELECT COUNT(*) FROM rat_meetings
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: CountRATMeetingsByType :one
SELECT COUNT(*) FROM rat_meetings
WHERE tenant_id = $1 AND company_id = $2 AND meeting_type = $3 AND deleted_at IS NULL;

-- name: CountRATMeetingsByStatus :one
SELECT COUNT(*) FROM rat_meetings
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL;

-- name: CountRATMeetingsByTypeAndStatus :one
SELECT COUNT(*) FROM rat_meetings
WHERE tenant_id = $1 AND company_id = $2 AND meeting_type = $3 AND status = $4 AND deleted_at IS NULL;

-- name: InsertRATMeeting :exec
INSERT INTO rat_meetings (
    id, tenant_id, company_id, meeting_type, meeting_number, title, description,
    meeting_date, meeting_time, meeting_location, meeting_mode, online_link,
    total_eligible_members, quorum_required, actual_attendees, quorum_met,
    status, cancelled_reason, rescheduled_to_id, convened_by, secretary_id,
    created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23);

-- name: UpdateRATMeeting :exec
UPDATE rat_meetings
SET meeting_type = $4, meeting_number = $5, title = $6, description = $7,
    meeting_date = $8, meeting_time = $9, meeting_location = $10, meeting_mode = $11, online_link = $12,
    total_eligible_members = $13, quorum_required = $14, actual_attendees = $15, quorum_met = $16,
    status = $17, cancelled_reason = $18, rescheduled_to_id = $19, convened_by = $20, secretary_id = $21,
    updated_at = $22
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: UpdateRATMeetingStatus :exec
UPDATE rat_meetings
SET status = $4, updated_at = $5
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteRATMeeting :exec
UPDATE rat_meetings SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- ── Agenda Item ──────────────────────────────────────────────────────────────

-- name: GetAgendaItemByID :one
SELECT id, tenant_id, company_id, meeting_id, agenda_number, category, title, description,
       proposal_type, proposer_id, proposal_document_id,
       discussion_notes, discussion_duration,
       result, result_notes, vote_for, vote_against, vote_abstain,
       followup_action, followup_deadline, followup_assignee_id, followup_status,
       created_at, updated_at, deleted_at
FROM rat_agenda_items
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListAgendaItemsByMeeting :many
SELECT id, tenant_id, company_id, meeting_id, agenda_number, category, title, description,
       proposal_type, proposer_id, proposal_document_id,
       discussion_notes, discussion_duration,
       result, result_notes, vote_for, vote_against, vote_abstain,
       followup_action, followup_deadline, followup_assignee_id, followup_status,
       created_at, updated_at
FROM rat_agenda_items
WHERE tenant_id = $1 AND company_id = $2 AND meeting_id = $3 AND deleted_at IS NULL
ORDER BY agenda_number ASC
LIMIT $4 OFFSET $5;

-- name: CountAgendaItemsByMeeting :one
SELECT COUNT(*) FROM rat_agenda_items
WHERE tenant_id = $1 AND company_id = $2 AND meeting_id = $3 AND deleted_at IS NULL;

-- name: InsertAgendaItem :exec
INSERT INTO rat_agenda_items (
    id, tenant_id, company_id, meeting_id, agenda_number, category, title, description,
    proposal_type, proposer_id, proposal_document_id,
    discussion_notes, discussion_duration,
    result, result_notes, vote_for, vote_against, vote_abstain,
    followup_action, followup_deadline, followup_assignee_id, followup_status,
    created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24);

-- name: UpdateAgendaItem :exec
UPDATE rat_agenda_items
SET agenda_number = $4, category = $5, title = $6, description = $7,
    proposal_type = $8, proposer_id = $9, proposal_document_id = $10,
    discussion_notes = $11, discussion_duration = $12,
    result = $13, result_notes = $14, vote_for = $15, vote_against = $16, vote_abstain = $17,
    followup_action = $18, followup_deadline = $19, followup_assignee_id = $20, followup_status = $21,
    updated_at = $22
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteAgendaItem :exec
UPDATE rat_agenda_items SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- ── Attendance ───────────────────────────────────────────────────────────────

-- name: GetAttendanceByID :one
SELECT id, tenant_id, company_id, meeting_id, nasabah_id, attendance_type,
       proxy_for_id, proxy_document_id, voting_right, checked_in_at, checked_in_by, created_at
FROM rat_attendances
WHERE tenant_id = $1 AND company_id = $2 AND id = $3;

-- name: ListAttendancesByMeeting :many
SELECT id, tenant_id, company_id, meeting_id, nasabah_id, attendance_type,
       proxy_for_id, proxy_document_id, voting_right, checked_in_at, checked_in_by, created_at
FROM rat_attendances
WHERE tenant_id = $1 AND company_id = $2 AND meeting_id = $3
ORDER BY created_at ASC
LIMIT $4 OFFSET $5;

-- name: CountAttendancesByMeeting :one
SELECT COUNT(*) FROM rat_attendances
WHERE tenant_id = $1 AND company_id = $2 AND meeting_id = $3;

-- name: InsertAttendance :exec
INSERT INTO rat_attendances (
    id, tenant_id, company_id, meeting_id, nasabah_id, attendance_type,
    proxy_for_id, proxy_document_id, voting_right, checked_in_at, checked_in_by, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: DeleteAttendance :exec
DELETE FROM rat_attendances
WHERE tenant_id = $1 AND company_id = $2 AND id = $3;

-- ── Vote ─────────────────────────────────────────────────────────────────────

-- name: GetVoteByID :one
SELECT id, tenant_id, company_id, meeting_id, agenda_item_id, voter_id, vote, vote_method, voted_at, created_at
FROM rat_votes
WHERE tenant_id = $1 AND company_id = $2 AND id = $3;

-- name: ListVotesByAgendaItem :many
SELECT id, tenant_id, company_id, meeting_id, agenda_item_id, voter_id, vote, vote_method, voted_at, created_at
FROM rat_votes
WHERE tenant_id = $1 AND company_id = $2 AND agenda_item_id = $3
ORDER BY voted_at ASC
LIMIT $4 OFFSET $5;

-- name: CountVotesByAgendaItem :one
SELECT COUNT(*) FROM rat_votes
WHERE tenant_id = $1 AND company_id = $2 AND agenda_item_id = $3;

-- name: InsertVote :exec
INSERT INTO rat_votes (
    id, tenant_id, company_id, meeting_id, agenda_item_id, voter_id, vote, vote_method, voted_at, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- ── Election ─────────────────────────────────────────────────────────────────

-- name: GetElectionByID :one
SELECT id, tenant_id, company_id, meeting_id, agenda_item_id, position_type,
       term_start, term_end, vacancies, status, election_method,
       created_at, updated_at, deleted_at
FROM rat_elections
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListElectionsByMeeting :many
SELECT id, tenant_id, company_id, meeting_id, agenda_item_id, position_type,
       term_start, term_end, vacancies, status, election_method,
       created_at, updated_at
FROM rat_elections
WHERE tenant_id = $1 AND company_id = $2 AND meeting_id = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountElectionsByMeeting :one
SELECT COUNT(*) FROM rat_elections
WHERE tenant_id = $1 AND company_id = $2 AND meeting_id = $3 AND deleted_at IS NULL;

-- name: InsertElection :exec
INSERT INTO rat_elections (
    id, tenant_id, company_id, meeting_id, agenda_item_id, position_type,
    term_start, term_end, vacancies, status, election_method, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: UpdateElection :exec
UPDATE rat_elections
SET position_type = $4, term_start = $5, term_end = $6, vacancies = $7,
    status = $8, election_method = $9, updated_at = $10
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteElection :exec
UPDATE rat_elections SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- ── Election Candidate ───────────────────────────────────────────────────────

-- name: GetElectionCandidateByID :one
SELECT id, tenant_id, company_id, election_id, nasabah_id,
       nominated_by_id, seconded_by_id, nomination_statement,
       status, vote_count, rank, created_at
FROM rat_election_candidates
WHERE tenant_id = $1 AND company_id = $2 AND id = $3;

-- name: ListElectionCandidatesByElection :many
SELECT id, tenant_id, company_id, election_id, nasabah_id,
       nominated_by_id, seconded_by_id, nomination_statement,
       status, vote_count, rank, created_at
FROM rat_election_candidates
WHERE tenant_id = $1 AND company_id = $2 AND election_id = $3
ORDER BY created_at ASC
LIMIT $4 OFFSET $5;

-- name: CountElectionCandidatesByElection :one
SELECT COUNT(*) FROM rat_election_candidates
WHERE tenant_id = $1 AND company_id = $2 AND election_id = $3;

-- name: InsertElectionCandidate :exec
INSERT INTO rat_election_candidates (
    id, tenant_id, company_id, election_id, nasabah_id,
    nominated_by_id, seconded_by_id, nomination_statement, status, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: UpdateElectionCandidate :exec
UPDATE rat_election_candidates
SET status = $4, vote_count = $5, rank = $6
WHERE tenant_id = $1 AND company_id = $2 AND id = $3;
