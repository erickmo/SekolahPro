package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/aml_alert"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// AmlAlertRepository adalah concrete implementation dari aml_alert.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type AmlAlertRepository struct {
	db *sqlx.DB
}

// NewAmlAlertRepository membuat instance baru AmlAlertRepository.
func NewAmlAlertRepository(db *sqlx.DB) *AmlAlertRepository {
	return &AmlAlertRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity TransactionAlert baru ke database.
func (r *AmlAlertRepository) Save(ctx context.Context, s scope.Scope, e *aml_alert.TransactionAlert) error {
	const q = `
		INSERT INTO aml_transaction_alerts (
			id, tenant_id, company_id, nasabah_id, transaksi_id,
			rule_id, rule_type, alert_type, severity, description,
			transaction_details,
			investigated_by, investigation_notes, investigation_outcome,
			ltkm_filed, ltkm_reference,
			status, resolved_at, detected_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID, e.NasabahID, e.TransaksiID,
		e.RuleID, e.RuleType, e.AlertType, e.Severity, e.Description,
		e.TransactionDetails,
		e.InvestigatedBy, e.InvestigationNotes, e.InvestigationOutcome,
		e.LTKMFiled, e.LTKMReference,
		e.Status, e.ResolvedAt, e.DetectedAt, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save transaction alert: %w", err)
	}
	return nil
}

// Update mengupdate TransactionAlert yang sudah ada berdasarkan ID + scope.
func (r *AmlAlertRepository) Update(ctx context.Context, s scope.Scope, e *aml_alert.TransactionAlert) error {
	const q = `
		UPDATE aml_transaction_alerts
		SET nasabah_id = $4, transaksi_id = $5,
		    rule_id = $6, rule_type = $7, alert_type = $8, severity = $9, description = $10,
		    transaction_details = $11,
		    investigated_by = $12, investigation_notes = $13, investigation_outcome = $14,
		    ltkm_filed = $15, ltkm_reference = $16,
		    status = $17, resolved_at = $18
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.NasabahID, e.TransaksiID,
		e.RuleID, e.RuleType, e.AlertType, e.Severity, e.Description,
		e.TransactionDetails,
		e.InvestigatedBy, e.InvestigationNotes, e.InvestigationOutcome,
		e.LTKMFiled, e.LTKMReference,
		e.Status, e.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("update transaction alert: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return aml_alert.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *AmlAlertRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE aml_transaction_alerts SET deleted_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete transaction alert: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return aml_alert.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu TransactionAlert berdasarkan ID + scope.
func (r *AmlAlertRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*aml_alert.TransactionAlert, error) {
	const q = `
		SELECT id, nasabah_id, transaksi_id,
		       rule_id, rule_type, alert_type, severity, description,
		       transaction_details,
		       investigated_by, investigation_notes, investigation_outcome,
		       ltkm_filed, ltkm_reference,
		       status, resolved_at, detected_at, created_at, deleted_at
		FROM aml_transaction_alerts
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e aml_alert.TransactionAlert
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.NasabahID, &e.TransaksiID,
		&e.RuleID, &e.RuleType, &e.AlertType, &e.Severity, &e.Description,
		&e.TransactionDetails,
		&e.InvestigatedBy, &e.InvestigationNotes, &e.InvestigationOutcome,
		&e.LTKMFiled, &e.LTKMReference,
		&e.Status, &e.ResolvedAt, &e.DetectedAt, &e.CreatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, aml_alert.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get transaction alert by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar TransactionAlert dengan pagination, sorting, dan filter.
func (r *AmlAlertRepository) List(ctx context.Context, s scope.Scope, filter aml_alert.ListFilter, limit, offset int, sortBy, order string) ([]*aml_alert.TransactionAlert, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countAlerts(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectAlerts(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *AmlAlertRepository) buildWhereClause(s scope.Scope, f aml_alert.ListFilter) (string, []any) {
	conditions := []string{"tenant_id = $1", "company_id = $2", "deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	argIdx := 3

	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *f.Status)
		argIdx++
	}
	if f.Severity != nil {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, *f.Severity)
		argIdx++
	}
	if f.AlertType != nil {
		conditions = append(conditions, fmt.Sprintf("alert_type = $%d", argIdx))
		args = append(args, *f.AlertType)
		argIdx++
	}
	if f.RuleType != nil {
		conditions = append(conditions, fmt.Sprintf("rule_type = $%d", argIdx))
		args = append(args, *f.RuleType)
		argIdx++
	}
	if f.NasabahID != nil {
		conditions = append(conditions, fmt.Sprintf("nasabah_id = $%d", argIdx))
		args = append(args, *f.NasabahID)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")
	return where, args
}

// countAlerts menghitung total TransactionAlert aktif yang cocok dengan filter.
func (r *AmlAlertRepository) countAlerts(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM aml_transaction_alerts WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count transaction alerts: %w", err)
	}
	return total, nil
}

// selectAlerts mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *AmlAlertRepository) selectAlerts(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*aml_alert.TransactionAlert, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":  "created_at",
		"detected_at": "detected_at",
		"severity":    "severity",
		"alert_type":  "alert_type",
		"status":      "status",
		"rule_type":   "rule_type",
	}, "detected_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, nasabah_id, transaksi_id,
		        rule_id, rule_type, alert_type, severity, description,
		        transaction_details,
		        investigated_by, investigation_notes, investigation_outcome,
		        ltkm_filed, ltkm_reference,
		        status, resolved_at, detected_at, created_at, deleted_at
		 FROM aml_transaction_alerts
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list transaction alerts: %w", err)
	}
	defer rows.Close()
	return scanAlerts(rows)
}

func scanAlerts(rows *sql.Rows) ([]*aml_alert.TransactionAlert, error) {
	var results []*aml_alert.TransactionAlert
	for rows.Next() {
		var e aml_alert.TransactionAlert
		if err := rows.Scan(
			&e.ID, &e.NasabahID, &e.TransaksiID,
			&e.RuleID, &e.RuleType, &e.AlertType, &e.Severity, &e.Description,
			&e.TransactionDetails,
			&e.InvestigatedBy, &e.InvestigationNotes, &e.InvestigationOutcome,
			&e.LTKMFiled, &e.LTKMReference,
			&e.Status, &e.ResolvedAt, &e.DetectedAt, &e.CreatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// Compile-time interface check.
var (
	_ aml_alert.WriteRepository = (*AmlAlertRepository)(nil)
	_ aml_alert.ReadRepository  = (*AmlAlertRepository)(nil)
)

// suppressUnusedImport mencegah error unused import.
var (
	_ = time.Time{}
	_ = json.RawMessage{}
)
