package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ReserveRepository adalah concrete implementation dari semua repository interfaces
// domain reserve: ReserveFund dan ReserveFundTransaction.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type ReserveRepository struct {
	db *sqlx.DB
}

// NewReserveRepository membuat instance baru ReserveRepository.
func NewReserveRepository(db *sqlx.DB) *ReserveRepository {
	return &ReserveRepository{db: db}
}

// ── ReserveFund WriteRepository ───────────────────────────────────────────────

// Save menyimpan entity ReserveFund baru ke database.
func (r *ReserveRepository) Save(ctx context.Context, s scope.Scope, e *reserve.ReserveFund) error {
	const q = `
		INSERT INTO reserve_funds (
			id, tenant_id, company_id, name, fund_type, status,
			target_amount, current_balance, minimum_balance, contribution_pct,
			description, last_calculated_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.Name, e.FundType, e.Status,
		e.TargetAmount, e.CurrentBalance, e.MinimumBalance, e.ContributionPct,
		e.Description, e.LastCalculatedAt,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save reserve fund: %w", err)
	}
	return nil
}

// Update mengupdate ReserveFund yang sudah ada berdasarkan ID + scope.
func (r *ReserveRepository) Update(ctx context.Context, s scope.Scope, e *reserve.ReserveFund) error {
	const q = `
		UPDATE reserve_funds
		SET name = $4, fund_type = $5, status = $6,
		    target_amount = $7, current_balance = $8, minimum_balance = $9,
		    contribution_pct = $10, description = $11, last_calculated_at = $12,
		    updated_at = $13
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.Name, e.FundType, e.Status,
		e.TargetAmount, e.CurrentBalance, e.MinimumBalance, e.ContributionPct,
		e.Description, e.LastCalculatedAt,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update reserve fund: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return reserve.ErrFundNotFound
	}
	return nil
}

// Delete melakukan soft delete pada ReserveFund.
func (r *ReserveRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE reserve_funds SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete reserve fund: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return reserve.ErrFundNotFound
	}
	return nil
}

// ── ReserveFund ReadRepository ────────────────────────────────────────────────

// GetByID mengambil satu ReserveFund berdasarkan ID + scope.
func (r *ReserveRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*reserve.ReserveFund, error) {
	const q = `
		SELECT id, name, fund_type, status,
		       target_amount, current_balance, minimum_balance, contribution_pct,
		       description, last_calculated_at,
		       created_at, updated_at, deleted_at
		FROM reserve_funds
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e reserve.ReserveFund
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.Name, &e.FundType, &e.Status,
		&e.TargetAmount, &e.CurrentBalance, &e.MinimumBalance, &e.ContributionPct,
		&e.Description, &e.LastCalculatedAt,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, reserve.ErrFundNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get reserve fund by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar ReserveFund dengan filter, pagination, dan sorting.
func (r *ReserveRepository) List(ctx context.Context, s scope.Scope, filter reserve.ReserveFundFilter, limit, offset int, sortBy, order string) ([]*reserve.ReserveFund, int, error) {
	where, args := r.buildFundWhereClause(s, filter)

	total, err := r.countFunds(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectFunds(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *ReserveRepository) buildFundWhereClause(s scope.Scope, f reserve.ReserveFundFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if f.FundType != nil {
		conditions = append(conditions, fmt.Sprintf("fund_type = $%d", paramIdx))
		args = append(args, *f.FundType)
		paramIdx++
	}
	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", paramIdx))
		args = append(args, string(*f.Status))
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

func (r *ReserveRepository) countFunds(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM reserve_funds WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count reserve funds: %w", err)
	}
	return total, nil
}

func (r *ReserveRepository) selectFunds(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*reserve.ReserveFund, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":      "created_at",
		"name":            "name",
		"current_balance": "current_balance",
		"fund_type":       "fund_type",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, name, fund_type, status,
		        target_amount, current_balance, minimum_balance, contribution_pct,
		        description, last_calculated_at,
		        created_at, updated_at, deleted_at
		 FROM reserve_funds
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list reserve funds: %w", err)
	}
	defer rows.Close()
	return scanReserveFunds(rows)
}

func scanReserveFunds(rows *sql.Rows) ([]*reserve.ReserveFund, error) {
	var results []*reserve.ReserveFund
	for rows.Next() {
		var e reserve.ReserveFund
		if err := rows.Scan(
			&e.ID, &e.Name, &e.FundType, &e.Status,
			&e.TargetAmount, &e.CurrentBalance, &e.MinimumBalance, &e.ContributionPct,
			&e.Description, &e.LastCalculatedAt,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// ── ReserveFundTransaction WriteRepository ────────────────────────────────────

// SaveTransaction menyimpan entity ReserveFundTransaction baru ke database.
func (r *ReserveRepository) SaveTransaction(ctx context.Context, s scope.Scope, e *reserve.ReserveFundTransaction) error {
	const q = `
		INSERT INTO reserve_fund_transactions (
			id, tenant_id, company_id, reserve_fund_id, transaction_type,
			amount, balance_before, balance_after, reference_no, description,
			processed_by, processed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.ReserveFundID, e.TransactionType,
		e.Amount, e.BalanceBefore, e.BalanceAfter,
		e.ReferenceNo, e.Description,
		e.ProcessedBy, e.ProcessedAt,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save reserve fund transaction: %w", err)
	}
	return nil
}

// UpdateTransaction mengupdate ReserveFundTransaction yang sudah ada.
func (r *ReserveRepository) UpdateTransaction(ctx context.Context, s scope.Scope, e *reserve.ReserveFundTransaction) error {
	const q = `
		UPDATE reserve_fund_transactions
		SET reserve_fund_id = $4, transaction_type = $5,
		    amount = $6, balance_before = $7, balance_after = $8,
		    reference_no = $9, description = $10,
		    processed_by = $11, processed_at = $12, updated_at = $13
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.ReserveFundID, e.TransactionType,
		e.Amount, e.BalanceBefore, e.BalanceAfter,
		e.ReferenceNo, e.Description,
		e.ProcessedBy, e.ProcessedAt,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update reserve fund transaction: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return reserve.ErrTransactionNotFound
	}
	return nil
}

// DeleteTransaction melakukan soft delete pada ReserveFundTransaction.
func (r *ReserveRepository) DeleteTransaction(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE reserve_fund_transactions SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete reserve fund transaction: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return reserve.ErrTransactionNotFound
	}
	return nil
}

// ── ReserveFundTransaction ReadRepository ─────────────────────────────────────

// GetTransactionByID mengambil satu ReserveFundTransaction berdasarkan ID + scope.
func (r *ReserveRepository) GetTransactionByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*reserve.ReserveFundTransaction, error) {
	const q = `
		SELECT id, reserve_fund_id, transaction_type,
		       amount, balance_before, balance_after, reference_no, description,
		       processed_by, processed_at,
		       created_at, updated_at, deleted_at
		FROM reserve_fund_transactions
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e reserve.ReserveFundTransaction
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.ReserveFundID, &e.TransactionType,
		&e.Amount, &e.BalanceBefore, &e.BalanceAfter,
		&e.ReferenceNo, &e.Description,
		&e.ProcessedBy, &e.ProcessedAt,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, reserve.ErrTransactionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get reserve fund transaction by id: %w", err)
	}
	return &e, nil
}

// ListTransactions mengambil daftar ReserveFundTransaction dengan filter, pagination, dan sorting.
func (r *ReserveRepository) ListTransactions(ctx context.Context, s scope.Scope, filter reserve.ReserveFundTransactionFilter, limit, offset int, sortBy, order string) ([]*reserve.ReserveFundTransaction, int, error) {
	where, args := r.buildTransactionWhereClause(s, filter)

	total, err := r.countTransactions(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectTransactions(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *ReserveRepository) buildTransactionWhereClause(s scope.Scope, f reserve.ReserveFundTransactionFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if f.ReserveFundID != nil {
		conditions = append(conditions, fmt.Sprintf("reserve_fund_id = $%d", paramIdx))
		args = append(args, *f.ReserveFundID)
		paramIdx++
	}
	if f.TransactionType != nil {
		conditions = append(conditions, fmt.Sprintf("transaction_type = $%d", paramIdx))
		args = append(args, string(*f.TransactionType))
		paramIdx++
	}
	if f.ProcessedAtFrom != nil {
		conditions = append(conditions, fmt.Sprintf("processed_at >= $%d", paramIdx))
		args = append(args, *f.ProcessedAtFrom)
		paramIdx++
	}
	if f.ProcessedAtTo != nil {
		conditions = append(conditions, fmt.Sprintf("processed_at <= $%d", paramIdx))
		args = append(args, *f.ProcessedAtTo)
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

func (r *ReserveRepository) countTransactions(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM reserve_fund_transactions WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count reserve fund transactions: %w", err)
	}
	return total, nil
}

func (r *ReserveRepository) selectTransactions(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*reserve.ReserveFundTransaction, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":      "created_at",
		"processed_at":    "processed_at",
		"transaction_type": "transaction_type",
		"amount":          "amount",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, reserve_fund_id, transaction_type,
		        amount, balance_before, balance_after, reference_no, description,
		        processed_by, processed_at,
		        created_at, updated_at, deleted_at
		 FROM reserve_fund_transactions
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list reserve fund transactions: %w", err)
	}
	defer rows.Close()
	return scanReserveFundTransactions(rows)
}

func scanReserveFundTransactions(rows *sql.Rows) ([]*reserve.ReserveFundTransaction, error) {
	var results []*reserve.ReserveFundTransaction
	for rows.Next() {
		var e reserve.ReserveFundTransaction
		if err := rows.Scan(
			&e.ID, &e.ReserveFundID, &e.TransactionType,
			&e.Amount, &e.BalanceBefore, &e.BalanceAfter,
			&e.ReferenceNo, &e.Description,
			&e.ProcessedBy, &e.ProcessedAt,
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
	_ reserve.ReserveFundWriteRepository           = (*ReserveRepository)(nil)
	_ reserve.ReserveFundReadRepository            = (*ReserveRepository)(nil)
	_ reserve.ReserveFundTransactionWriteRepository = (*ReserveRepository)(nil)
	_ reserve.ReserveFundTransactionReadRepository  = (*ReserveRepository)(nil)
)
