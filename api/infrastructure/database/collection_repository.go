package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// CollectionRepository mengimplementasikan semua repository interface untuk domain collection.
type CollectionRepository struct {
	db *sqlx.DB
}

// NewCollectionRepository membuat instance baru CollectionRepository.
func NewCollectionRepository(db *sqlx.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

// ── CollectionCase WriteRepository ───────────────────────────────────────────

// SaveCase menyimpan CollectionCase baru ke database.
func (r *CollectionRepository) SaveCase(ctx context.Context, s scope.Scope, c *collection.CollectionCase) error {
	const q = `
		INSERT INTO collection_cases (
			id, tenant_id, company_id, pinjaman_id, nasabah_id,
			case_number, current_dpd, aging_bucket, total_overdue_amount, total_overdue_installments,
			assigned_collector_id, assigned_at, escalation_level,
			status, resolution_type, opened_at, closed_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`

	_, err := r.db.ExecContext(ctx, q,
		c.ID, s.TenantID, s.CompanyID,
		c.PinjamanID, c.NasabahID,
		c.CaseNumber, c.CurrentDPD, string(c.AgingBucket), c.TotalOverdueAmount, c.TotalOverdueInstallments,
		c.AssignedCollectorID, c.AssignedAt, c.EscalationLevel,
		string(c.Status), resolutionTypeToNil(c.ResolutionType), c.OpenedAt, c.ClosedAt, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save collection case: %w", err)
	}
	return nil
}

// UpdateCase mengupdate CollectionCase yang sudah ada.
func (r *CollectionRepository) UpdateCase(ctx context.Context, s scope.Scope, c *collection.CollectionCase) error {
	const q = `
		UPDATE collection_cases
		SET current_dpd = $4, aging_bucket = $5, total_overdue_amount = $6,
		    total_overdue_installments = $7, assigned_collector_id = $8, assigned_at = $9,
		    escalation_level = $10, status = $11, resolution_type = $12, closed_at = $13,
		    updated_at = $14
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, c.ID,
		c.CurrentDPD, string(c.AgingBucket), c.TotalOverdueAmount,
		c.TotalOverdueInstallments, c.AssignedCollectorID, c.AssignedAt,
		c.EscalationLevel, string(c.Status), resolutionTypeToNil(c.ResolutionType), c.ClosedAt,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update collection case: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return collection.ErrCaseNotFound
	}
	return nil
}

// DeleteCase melakukan soft delete pada CollectionCase.
func (r *CollectionRepository) DeleteCase(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE collection_cases SET deleted_at = now(), updated_at = now()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete collection case: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return collection.ErrCaseNotFound
	}
	return nil
}

// ── CollectionCase ReadRepository ────────────────────────────────────────────

// GetCaseByID mengambil satu CollectionCase berdasarkan ID + scope.
func (r *CollectionRepository) GetCaseByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*collection.CollectionCase, error) {
	const q = `
		SELECT id, case_number, current_dpd, aging_bucket, total_overdue_amount,
		       total_overdue_installments, pinjaman_id, nasabah_id,
		       assigned_collector_id, assigned_at, escalation_level,
		       status, resolution_type, opened_at, closed_at, created_at, updated_at, deleted_at
		FROM collection_cases
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var c collection.CollectionCase
	var agingBucket, caseStatus string
	var resolutionType sql.NullString

	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&c.ID, &c.CaseNumber, &c.CurrentDPD, &agingBucket, &c.TotalOverdueAmount,
		&c.TotalOverdueInstallments, &c.PinjamanID, &c.NasabahID,
		&c.AssignedCollectorID, &c.AssignedAt, &c.EscalationLevel,
		&caseStatus, &resolutionType, &c.OpenedAt, &c.ClosedAt, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, collection.ErrCaseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get collection case by id: %w", err)
	}

	c.AgingBucket = collection.AgingBucket(agingBucket)
	c.Status = collection.CaseStatus(caseStatus)
	if resolutionType.Valid {
		rt := collection.ResolutionType(resolutionType.String)
		c.ResolutionType = &rt
	}
	return &c, nil
}

// ListCases mengambil daftar CollectionCase dengan filter, pagination, dan sorting.
func (r *CollectionRepository) ListCases(ctx context.Context, s scope.Scope, filter collection.CaseFilter, limit, offset int, sortBy, order string) ([]*collection.CollectionCase, int, error) {
	total, err := r.countCases(ctx, s, filter)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectCases(ctx, s, filter, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *CollectionRepository) countCases(ctx context.Context, s scope.Scope, f collection.CaseFilter) (int, error) {
	base := `SELECT COUNT(*) FROM collection_cases WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`
	args := []any{s.TenantID, s.CompanyID}
	n := 3

	base, args = appendCaseFilter(base, &n, args, f)

	var total int
	if err := r.db.QueryRowContext(ctx, base, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count collection cases: %w", err)
	}
	return total, nil
}

func (r *CollectionRepository) selectCases(ctx context.Context, s scope.Scope, f collection.CaseFilter, limit, offset int, sortBy, order string) ([]*collection.CollectionCase, error) {
	col := safeColumn(sortBy, map[string]string{
		"case_number": "case_number", "current_dpd": "current_dpd",
		"aging_bucket": "aging_bucket", "status": "status",
		"opened_at": "opened_at", "created_at": "created_at",
	}, "created_at")
	dir := safeOrder(order)

	base := `SELECT id, case_number, current_dpd, aging_bucket, total_overdue_amount,
		        total_overdue_installments, pinjaman_id, nasabah_id,
		        assigned_collector_id, assigned_at, escalation_level,
		        status, resolution_type, opened_at, closed_at, created_at, updated_at, deleted_at
		 FROM collection_cases
		 WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`

	args := []any{s.TenantID, s.CompanyID}
	n := 3

	base, args = appendCaseFilter(base, &n, args, f)

	base += fmt.Sprintf(` ORDER BY %s %s LIMIT $%d OFFSET $%d`, col, dir, n, n+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, fmt.Errorf("list collection cases: %w", err)
	}
	defer rows.Close()
	return scanCases(rows)
}

func appendCaseFilter(q string, n *int, args []any, f collection.CaseFilter) (string, []any) {
	if f.Status != nil {
		q += fmt.Sprintf(` AND status = $%d`, *n)
		args = append(args, string(*f.Status))
		*n++
	}
	if f.AgingBucket != nil {
		q += fmt.Sprintf(` AND aging_bucket = $%d`, *n)
		args = append(args, string(*f.AgingBucket))
		*n++
	}
	if f.CollectorID != nil {
		q += fmt.Sprintf(` AND assigned_collector_id = $%d`, *n)
		args = append(args, *f.CollectorID)
		*n++
	}
	if f.NasabahID != nil {
		q += fmt.Sprintf(` AND nasabah_id = $%d`, *n)
		args = append(args, *f.NasabahID)
		*n++
	}
	return q, args
}

func scanCases(rows *sql.Rows) ([]*collection.CollectionCase, error) {
	var results []*collection.CollectionCase
	for rows.Next() {
		var c collection.CollectionCase
		var agingBucket, caseStatus string
		var resolutionType sql.NullString

		if err := rows.Scan(
			&c.ID, &c.CaseNumber, &c.CurrentDPD, &agingBucket, &c.TotalOverdueAmount,
			&c.TotalOverdueInstallments, &c.PinjamanID, &c.NasabahID,
			&c.AssignedCollectorID, &c.AssignedAt, &c.EscalationLevel,
			&caseStatus, &resolutionType, &c.OpenedAt, &c.ClosedAt, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		); err != nil {
			return nil, err
		}

		c.AgingBucket = collection.AgingBucket(agingBucket)
		c.Status = collection.CaseStatus(caseStatus)
		if resolutionType.Valid {
			rt := collection.ResolutionType(resolutionType.String)
			c.ResolutionType = &rt
		}
		results = append(results, &c)
	}
	return results, rows.Err()
}

// ── Activity WriteRepository ─────────────────────────────────────────────────

// SaveActivity menyimpan Activity baru ke database.
func (r *CollectionRepository) SaveActivity(ctx context.Context, s scope.Scope, a *collection.Activity) error {
	const q = `
		INSERT INTO collection_activities (
			id, tenant_id, company_id, case_id,
			activity_type, activity_date, performed_by,
			contact_result, notes, nasabah_response,
			promise_amount, promise_date,
			document_ids,
			followup_required, followup_date, followup_type,
			created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`

	var docIDs interface{}
	if a.DocumentIDs == nil {
		docIDs = []uuid.UUID{}
	} else {
		docIDs = a.DocumentIDs
	}

	_, err := r.db.ExecContext(ctx, q,
		a.ID, s.TenantID, s.CompanyID, a.CaseID,
		string(a.ActivityType), a.ActivityDate, a.PerformedBy,
		string(a.ContactResult), a.Notes, nasabahResponseToNil(a.NasabahResponse),
		a.PromiseAmount, a.PromiseDate,
		docIDs,
		a.FollowupRequired, a.FollowupDate, a.FollowupType,
		a.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save collection activity: %w", err)
	}
	return nil
}

// ── Activity ReadRepository ──────────────────────────────────────────────────

// ListActivities mengambil daftar Activity berdasarkan case_id dengan filter.
func (r *CollectionRepository) ListActivities(ctx context.Context, s scope.Scope, caseID uuid.UUID, activityType *collection.ActivityType, limit, offset int) ([]*collection.Activity, int, error) {
	total, err := r.countActivities(ctx, s, caseID, activityType)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectActivities(ctx, s, caseID, activityType, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *CollectionRepository) countActivities(ctx context.Context, s scope.Scope, caseID uuid.UUID, actType *collection.ActivityType) (int, error) {
	base := `SELECT COUNT(*) FROM collection_activities WHERE tenant_id = $1 AND company_id = $2 AND case_id = $3`
	args := []any{s.TenantID, s.CompanyID, caseID}

	if actType != nil {
		base += ` AND activity_type = $4`
		args = append(args, string(*actType))
	}

	var total int
	if err := r.db.QueryRowContext(ctx, base, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count collection activities: %w", err)
	}
	return total, nil
}

func (r *CollectionRepository) selectActivities(ctx context.Context, s scope.Scope, caseID uuid.UUID, actType *collection.ActivityType, limit, offset int) ([]*collection.Activity, error) {
	base := `SELECT id, case_id, activity_type, activity_date, performed_by,
	                 contact_result, notes, nasabah_response, promise_amount, promise_date,
	                 document_ids, followup_required, followup_date, followup_type, created_at
	         FROM collection_activities
	         WHERE tenant_id = $1 AND company_id = $2 AND case_id = $3`
	args := []any{s.TenantID, s.CompanyID, caseID}
	n := 4

	if actType != nil {
		base += fmt.Sprintf(` AND activity_type = $%d`, n)
		args = append(args, string(*actType))
		n++
	}

	base += fmt.Sprintf(` ORDER BY activity_date DESC, created_at DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, fmt.Errorf("list collection activities: %w", err)
	}
	defer rows.Close()
	return scanActivities(rows)
}

func scanActivities(rows *sql.Rows) ([]*collection.Activity, error) {
	var results []*collection.Activity
	for rows.Next() {
		var a collection.Activity
		var actType, contactResult string
		var nasabahResp sql.NullString
		var followupType sql.NullString
		var promiseDate sql.NullTime
		var docIDsSlice []uuid.UUID

		if err := rows.Scan(
			&a.ID, &a.CaseID, &actType, &a.ActivityDate, &a.PerformedBy,
			&contactResult, &a.Notes, &nasabahResp, &a.PromiseAmount, &promiseDate,
			&docIDsSlice, &a.FollowupRequired, &a.FollowupDate, &followupType, &a.CreatedAt,
		); err != nil {
			return nil, err
		}

		a.ActivityType = collection.ActivityType(actType)
		a.ContactResult = collection.ContactResult(contactResult)
		if nasabahResp.Valid {
			nr := collection.NasabahResponse(nasabahResp.String)
			a.NasabahResponse = &nr
		}
		if promiseDate.Valid {
			a.PromiseDate = &promiseDate.Time
		}
		if followupType.Valid {
			a.FollowupType = &followupType.String
		}
		if len(docIDsSlice) > 0 {
			a.DocumentIDs = docIDsSlice
		}
		results = append(results, &a)
	}
	return results, rows.Err()
}

// ── Performance WriteRepository ──────────────────────────────────────────────

// SavePerformance menyimpan Performance baru ke database.
func (r *CollectionRepository) SavePerformance(ctx context.Context, s scope.Scope, p *collection.Performance) error {
	const q = `
		INSERT INTO collector_performances (
			id, tenant_id, company_id, collector_id,
			period_month,
			total_calls, successful_contacts, total_visits, successful_visits,
			cases_handled, cases_resolved, total_amount_collected,
			promise_to_pay_count, promise_kept_count,
			contact_rate_pct, resolution_rate_pct, collection_rate_pct, promise_kept_rate_pct,
			calculated_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`

	_, err := r.db.ExecContext(ctx, q,
		p.ID, s.TenantID, s.CompanyID, p.CollectorID,
		p.PeriodMonth,
		p.TotalCalls, p.SuccessfulContacts, p.TotalVisits, p.SuccessfulVisits,
		p.CasesHandled, p.CasesResolved, p.TotalAmountCollected,
		p.PromiseToPayCount, p.PromiseKeptCount,
		p.ContactRatePct, p.ResolutionRatePct, p.CollectionRatePct, p.PromiseKeptRatePct,
		p.CalculatedAt, p.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save collector performance: %w", err)
	}
	return nil
}

// ── Performance ReadRepository ───────────────────────────────────────────────

// ListPerformances mengambil daftar Performance dengan filter.
func (r *CollectionRepository) ListPerformances(ctx context.Context, s scope.Scope, collectorID *uuid.UUID, periodMonth string, limit, offset int) ([]*collection.Performance, int, error) {
	total, err := r.countPerformances(ctx, s, collectorID, periodMonth)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectPerformances(ctx, s, collectorID, periodMonth, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *CollectionRepository) countPerformances(ctx context.Context, s scope.Scope, collectorID *uuid.UUID, periodMonth string) (int, error) {
	base := `SELECT COUNT(*) FROM collector_performances WHERE tenant_id = $1 AND company_id = $2`
	args := []any{s.TenantID, s.CompanyID}
	n := 3

	base, args = appendPerfFilter(base, &n, args, collectorID, periodMonth)

	var total int
	if err := r.db.QueryRowContext(ctx, base, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count collector performances: %w", err)
	}
	return total, nil
}

func (r *CollectionRepository) selectPerformances(ctx context.Context, s scope.Scope, collectorID *uuid.UUID, periodMonth string, limit, offset int) ([]*collection.Performance, error) {
	base := `SELECT id, collector_id, period_month,
	                 total_calls, successful_contacts, total_visits, successful_visits,
	                 cases_handled, cases_resolved, total_amount_collected,
	                 promise_to_pay_count, promise_kept_count,
	                 contact_rate_pct, resolution_rate_pct, collection_rate_pct, promise_kept_rate_pct,
	                 calculated_at, created_at
	         FROM collector_performances
	         WHERE tenant_id = $1 AND company_id = $2`
	args := []any{s.TenantID, s.CompanyID}
	n := 3

	base, args = appendPerfFilter(base, &n, args, collectorID, periodMonth)

	base += fmt.Sprintf(` ORDER BY period_month DESC, created_at DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, fmt.Errorf("list collector performances: %w", err)
	}
	defer rows.Close()
	return scanPerformances(rows)
}

func appendPerfFilter(q string, n *int, args []any, collectorID *uuid.UUID, periodMonth string) (string, []any) {
	if collectorID != nil {
		q += fmt.Sprintf(` AND collector_id = $%d`, *n)
		args = append(args, *collectorID)
		*n++
	}
	if periodMonth != "" {
		q += fmt.Sprintf(` AND period_month = $%d`, *n)
		args = append(args, periodMonth)
		*n++
	}
	return q, args
}

func scanPerformances(rows *sql.Rows) ([]*collection.Performance, error) {
	var results []*collection.Performance
	for rows.Next() {
		var p collection.Performance
		if err := rows.Scan(
			&p.ID, &p.CollectorID, &p.PeriodMonth,
			&p.TotalCalls, &p.SuccessfulContacts, &p.TotalVisits, &p.SuccessfulVisits,
			&p.CasesHandled, &p.CasesResolved, &p.TotalAmountCollected,
			&p.PromiseToPayCount, &p.PromiseKeptCount,
			&p.ContactRatePct, &p.ResolutionRatePct, &p.CollectionRatePct, &p.PromiseKeptRatePct,
			&p.CalculatedAt, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &p)
	}
	return results, rows.Err()
}

// ── Helper functions ─────────────────────────────────────────────────────────

func resolutionTypeToNil(rt *collection.ResolutionType) interface{} {
	if rt == nil {
		return nil
	}
	return string(*rt)
}

func nasabahResponseToNil(nr *collection.NasabahResponse) interface{} {
	if nr == nil {
		return nil
	}
	return string(*nr)
}

// isValidPeriodMonth memvalidasi format period_month "YYYY-MM".
func isValidPeriodMonth(s string) bool {
	if len(s) != 7 {
		return false
	}
	return s[4] == '-' && isDigits(s[:4]) && isDigits(s[5:])
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
