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

	"github.com/yourorg/boilerplate/internal/domain/aml_rule"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// AmlRuleRepository adalah concrete implementation dari aml_rule.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type AmlRuleRepository struct {
	db *sqlx.DB
}

// NewAmlRuleRepository membuat instance baru AmlRuleRepository.
func NewAmlRuleRepository(db *sqlx.DB) *AmlRuleRepository {
	return &AmlRuleRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity MonitoringRule baru ke database.
func (r *AmlRuleRepository) Save(ctx context.Context, s scope.Scope, e *aml_rule.MonitoringRule) error {
	const q = `
		INSERT INTO aml_monitoring_rules (
			id, tenant_id, company_id,
			rule_code, rule_name, rule_type, description,
			parameters, is_active, applies_to, auto_alert, auto_block,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.RuleCode, e.RuleName, e.RuleType, e.Description,
		e.Parameters, e.IsActive, e.AppliesTo, e.AutoAlert, e.AutoBlock,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save monitoring rule: %w", err)
	}
	return nil
}

// Update mengupdate MonitoringRule yang sudah ada berdasarkan ID + scope.
func (r *AmlRuleRepository) Update(ctx context.Context, s scope.Scope, e *aml_rule.MonitoringRule) error {
	const q = `
		UPDATE aml_monitoring_rules
		SET rule_code = $4, rule_name = $5, rule_type = $6, description = $7,
		    parameters = $8, is_active = $9, applies_to = $10,
		    auto_alert = $11, auto_block = $12, updated_at = $13
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.RuleCode, e.RuleName, e.RuleType, e.Description,
		e.Parameters, e.IsActive, e.AppliesTo,
		e.AutoAlert, e.AutoBlock, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update monitoring rule: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return aml_rule.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *AmlRuleRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE aml_monitoring_rules SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete monitoring rule: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return aml_rule.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu MonitoringRule berdasarkan ID + scope.
func (r *AmlRuleRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*aml_rule.MonitoringRule, error) {
	const q = `
		SELECT id, rule_code, rule_name, rule_type, description,
		       parameters, is_active, applies_to, auto_alert, auto_block,
		       created_at, updated_at, deleted_at
		FROM aml_monitoring_rules
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e aml_rule.MonitoringRule
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.RuleCode, &e.RuleName, &e.RuleType, &e.Description,
		&e.Parameters, &e.IsActive, &e.AppliesTo, &e.AutoAlert, &e.AutoBlock,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, aml_rule.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get monitoring rule by id: %w", err)
	}
	return &e, nil
}

// GetByRuleCode mengambil satu MonitoringRule berdasarkan rule_code + scope.
func (r *AmlRuleRepository) GetByRuleCode(ctx context.Context, s scope.Scope, ruleCode string) (*aml_rule.MonitoringRule, error) {
	const q = `
		SELECT id, rule_code, rule_name, rule_type, description,
		       parameters, is_active, applies_to, auto_alert, auto_block,
		       created_at, updated_at, deleted_at
		FROM aml_monitoring_rules
		WHERE tenant_id = $1 AND company_id = $2 AND rule_code = $3 AND deleted_at IS NULL`

	var e aml_rule.MonitoringRule
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, ruleCode).Scan(
		&e.ID, &e.RuleCode, &e.RuleName, &e.RuleType, &e.Description,
		&e.Parameters, &e.IsActive, &e.AppliesTo, &e.AutoAlert, &e.AutoBlock,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, aml_rule.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get monitoring rule by code: %w", err)
	}
	return &e, nil
}

// List mengambil daftar MonitoringRule dengan pagination, sorting, dan filter.
func (r *AmlRuleRepository) List(ctx context.Context, s scope.Scope, filter aml_rule.ListFilter, limit, offset int, sortBy, order string) ([]*aml_rule.MonitoringRule, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countRules(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectRules(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *AmlRuleRepository) buildWhereClause(s scope.Scope, f aml_rule.ListFilter) (string, []any) {
	conditions := []string{"tenant_id = $1", "company_id = $2", "deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	argIdx := 3

	if f.RuleType != nil {
		conditions = append(conditions, fmt.Sprintf("rule_type = $%d", argIdx))
		args = append(args, *f.RuleType)
		argIdx++
	}
	if f.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *f.IsActive)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")
	return where, args
}

// countRules menghitung total MonitoringRule aktif yang cocok dengan filter.
func (r *AmlRuleRepository) countRules(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM aml_monitoring_rules WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count monitoring rules: %w", err)
	}
	return total, nil
}

// selectRules mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *AmlRuleRepository) selectRules(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*aml_rule.MonitoringRule, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at": "created_at",
		"rule_code":  "rule_code",
		"rule_name":  "rule_name",
		"rule_type":  "rule_type",
		"is_active":  "is_active",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, rule_code, rule_name, rule_type, description,
		        parameters, is_active, applies_to, auto_alert, auto_block,
		        created_at, updated_at, deleted_at
		 FROM aml_monitoring_rules
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list monitoring rules: %w", err)
	}
	defer rows.Close()
	return scanRules(rows)
}

func scanRules(rows *sql.Rows) ([]*aml_rule.MonitoringRule, error) {
	var results []*aml_rule.MonitoringRule
	for rows.Next() {
		var e aml_rule.MonitoringRule
		if err := rows.Scan(
			&e.ID, &e.RuleCode, &e.RuleName, &e.RuleType, &e.Description,
			&e.Parameters, &e.IsActive, &e.AppliesTo, &e.AutoAlert, &e.AutoBlock,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// Compile-time interface check.
var (
	_ aml_rule.WriteRepository = (*AmlRuleRepository)(nil)
	_ aml_rule.ReadRepository  = (*AmlRuleRepository)(nil)
)

// suppressUnusedImport mencegah error unused import.
var (
	_ = time.Time{}
	_ = json.RawMessage{}
)
