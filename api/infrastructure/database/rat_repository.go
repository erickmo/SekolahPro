package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/rat"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// RatRepository adalah concrete implementation dari rat.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type RatRepository struct {
	db *sqlx.DB
}

// NewRatRepository membuat instance baru RatRepository.
func NewRatRepository(db *sqlx.DB) *RatRepository {
	return &RatRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity Meeting baru ke database.
func (r *RatRepository) Save(ctx context.Context, s scope.Scope, e *rat.Meeting) error {
	const q = `
		INSERT INTO rat_meetings (
			id, tenant_id, company_id,
			meeting_type, meeting_number, title, description,
			meeting_date, meeting_time, meeting_location, meeting_mode, online_link,
			total_eligible_members, quorum_required, actual_attendees, quorum_met,
			status, cancelled_reason, rescheduled_to_id,
			agenda_doc_id, minutes_doc_id, financial_report_id, shu_proposal_id,
			convened_by, secretary_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.MeetingType, e.MeetingNumber, e.Title, e.Description,
		e.MeetingDate, e.MeetingTime, e.MeetingLocation, e.MeetingMode, e.OnlineLink,
		e.TotalEligibleMembers, e.QuorumRequired, e.ActualAttendees, e.QuorumMet,
		e.Status, e.CancelledReason, e.RescheduledToID,
		e.AgendaDocID, e.MinutesDocID, e.FinancialReportID, e.ShuProposalID,
		e.ConvenedBy, e.SecretaryID, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save rat meeting: %w", err)
	}
	return nil
}

// Update mengupdate Meeting yang sudah ada berdasarkan ID + scope.
func (r *RatRepository) Update(ctx context.Context, s scope.Scope, e *rat.Meeting) error {
	const q = `
		UPDATE rat_meetings
		SET meeting_type = $4, meeting_number = $5, title = $6, description = $7,
		    meeting_date = $8, meeting_time = $9, meeting_location = $10,
		    meeting_mode = $11, online_link = $12,
		    total_eligible_members = $13, quorum_required = $14,
		    actual_attendees = $15, quorum_met = $16,
		    status = $17, cancelled_reason = $18, rescheduled_to_id = $19,
		    agenda_doc_id = $20, minutes_doc_id = $21,
		    financial_report_id = $22, shu_proposal_id = $23,
		    convened_by = $24, secretary_id = $25, updated_at = $26
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.MeetingType, e.MeetingNumber, e.Title, e.Description,
		e.MeetingDate, e.MeetingTime, e.MeetingLocation, e.MeetingMode, e.OnlineLink,
		e.TotalEligibleMembers, e.QuorumRequired,
		e.ActualAttendees, e.QuorumMet,
		e.Status, e.CancelledReason, e.RescheduledToID,
		e.AgendaDocID, e.MinutesDocID,
		e.FinancialReportID, e.ShuProposalID,
		e.ConvenedBy, e.SecretaryID, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update rat meeting: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return rat.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *RatRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE rat_meetings SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete rat meeting: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return rat.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu Meeting berdasarkan ID + scope.
func (r *RatRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*rat.Meeting, error) {
	const q = `
		SELECT id, meeting_type, meeting_number, title, description,
		       meeting_date, meeting_time, meeting_location, meeting_mode, online_link,
		       total_eligible_members, quorum_required, actual_attendees, quorum_met,
		       status, cancelled_reason, rescheduled_to_id,
		       agenda_doc_id, minutes_doc_id, financial_report_id, shu_proposal_id,
		       convened_by, secretary_id, created_at, updated_at, deleted_at
		FROM rat_meetings
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e rat.Meeting
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.MeetingType, &e.MeetingNumber, &e.Title, &e.Description,
		&e.MeetingDate, &e.MeetingTime, &e.MeetingLocation, &e.MeetingMode, &e.OnlineLink,
		&e.TotalEligibleMembers, &e.QuorumRequired, &e.ActualAttendees, &e.QuorumMet,
		&e.Status, &e.CancelledReason, &e.RescheduledToID,
		&e.AgendaDocID, &e.MinutesDocID, &e.FinancialReportID, &e.ShuProposalID,
		&e.ConvenedBy, &e.SecretaryID, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, rat.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get rat meeting by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar Meeting dengan pagination, sorting, dan filter.
func (r *RatRepository) List(ctx context.Context, s scope.Scope, filter rat.ListFilter, limit, offset int, sortBy, order string) ([]*rat.Meeting, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countMeetings(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectMeetings(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *RatRepository) buildWhereClause(s scope.Scope, f rat.ListFilter) (string, []any) {
	conditions := []string{"tenant_id = $1", "company_id = $2", "deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	argIdx := 3

	if f.MeetingType != nil {
		conditions = append(conditions, fmt.Sprintf("meeting_type = $%d", argIdx))
		args = append(args, *f.MeetingType)
		argIdx++
	}
	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *f.Status)
		argIdx++
	}
	if f.MeetingDate != nil {
		conditions = append(conditions, fmt.Sprintf("meeting_date = $%d", argIdx))
		args = append(args, *f.MeetingDate)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")
	return where, args
}

// countMeetings menghitung total Meeting aktif yang cocok dengan filter.
func (r *RatRepository) countMeetings(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM rat_meetings WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count rat meetings: %w", err)
	}
	return total, nil
}

// selectMeetings mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *RatRepository) selectMeetings(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*rat.Meeting, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":    "created_at",
		"meeting_date":  "meeting_date",
		"meeting_type":  "meeting_type",
		"status":        "status",
		"meeting_number": "meeting_number",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, meeting_type, meeting_number, title, description,
		        meeting_date, meeting_time, meeting_location, meeting_mode, online_link,
		        total_eligible_members, quorum_required, actual_attendees, quorum_met,
		        status, cancelled_reason, rescheduled_to_id,
		        agenda_doc_id, minutes_doc_id, financial_report_id, shu_proposal_id,
		        convened_by, secretary_id, created_at, updated_at, deleted_at
		 FROM rat_meetings
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list rat meetings: %w", err)
	}
	defer rows.Close()
	return scanRatMeetings(rows)
}

func scanRatMeetings(rows *sql.Rows) ([]*rat.Meeting, error) {
	var results []*rat.Meeting
	for rows.Next() {
		var e rat.Meeting
		if err := rows.Scan(
			&e.ID, &e.MeetingType, &e.MeetingNumber, &e.Title, &e.Description,
			&e.MeetingDate, &e.MeetingTime, &e.MeetingLocation, &e.MeetingMode, &e.OnlineLink,
			&e.TotalEligibleMembers, &e.QuorumRequired, &e.ActualAttendees, &e.QuorumMet,
			&e.Status, &e.CancelledReason, &e.RescheduledToID,
			&e.AgendaDocID, &e.MinutesDocID, &e.FinancialReportID, &e.ShuProposalID,
			&e.ConvenedBy, &e.SecretaryID, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// Compile-time interface check.
var (
	_ rat.WriteRepository = (*RatRepository)(nil)
	_ rat.ReadRepository  = (*RatRepository)(nil)
)

// suppressUnusedImport mencegah error unused import untuk time package.
var _ = time.Time{}
