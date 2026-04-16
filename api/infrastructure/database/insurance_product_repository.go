package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// InsuranceProductRepository adalah concrete implementation dari insurance.ProductWriteRepository + ProductReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type InsuranceProductRepository struct {
	db *sqlx.DB
}

// NewInsuranceProductRepository membuat instance baru InsuranceProductRepository.
func NewInsuranceProductRepository(db *sqlx.DB) *InsuranceProductRepository {
	return &InsuranceProductRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity Product baru ke database.
func (r *InsuranceProductRepository) Save(ctx context.Context, s scope.Scope, e *insurance.Product) error {
	const q = `
		INSERT INTO insurance_products (
			id, tenant_id, company_id, name, code, product_type, status,
			provider_name, premium_amount, coverage_amount, premium_frequency,
			term_months, description, terms_conditions, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.Name, e.Code, e.ProductType, e.Status,
		e.ProviderName, e.PremiumAmount, e.CoverageAmount, e.PremiumFrequency,
		e.TermMonths, e.Description, e.TermsConditions,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save insurance product: %w", err)
	}
	return nil
}

// Update mengupdate Product yang sudah ada berdasarkan ID + scope.
func (r *InsuranceProductRepository) Update(ctx context.Context, s scope.Scope, e *insurance.Product) error {
	const q = `
		UPDATE insurance_products
		SET name = $4, code = $5, product_type = $6, status = $7,
		    provider_name = $8, premium_amount = $9, coverage_amount = $10,
		    premium_frequency = $11, term_months = $12,
		    description = $13, terms_conditions = $14, updated_at = $15
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.Name, e.Code, e.ProductType, e.Status,
		e.ProviderName, e.PremiumAmount, e.CoverageAmount, e.PremiumFrequency,
		e.TermMonths, e.Description, e.TermsConditions,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update insurance product: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return insurance.ErrProductNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *InsuranceProductRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE insurance_products SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete insurance product: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return insurance.ErrProductNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu Product berdasarkan ID + scope.
func (r *InsuranceProductRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*insurance.Product, error) {
	const q = `
		SELECT id, name, code, product_type, status,
		       provider_name, premium_amount, coverage_amount, premium_frequency,
		       term_months, description, terms_conditions,
		       created_at, updated_at, deleted_at
		FROM insurance_products
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e insurance.Product
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.Name, &e.Code, &e.ProductType, &e.Status,
		&e.ProviderName, &e.PremiumAmount, &e.CoverageAmount, &e.PremiumFrequency,
		&e.TermMonths, &e.Description, &e.TermsConditions,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, insurance.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get insurance product by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar Product dengan filter, pagination, dan sorting.
func (r *InsuranceProductRepository) List(ctx context.Context, s scope.Scope, filter insurance.ProductFilter, limit, offset int, sortBy, order string) ([]*insurance.Product, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countProducts(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectProducts(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *InsuranceProductRepository) buildWhereClause(s scope.Scope, f insurance.ProductFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if f.ProductType != nil {
		conditions = append(conditions, fmt.Sprintf("product_type = $%d", paramIdx))
		args = append(args, string(*f.ProductType))
		paramIdx++
	}
	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", paramIdx))
		args = append(args, string(*f.Status))
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

// countProducts menghitung total Product yang cocok dengan filter.
func (r *InsuranceProductRepository) countProducts(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM insurance_products WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count insurance products: %w", err)
	}
	return total, nil
}

// selectProducts mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *InsuranceProductRepository) selectProducts(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*insurance.Product, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":   "created_at",
		"name":         "name",
		"product_type": "product_type",
		"status":       "status",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, name, code, product_type, status,
		        provider_name, premium_amount, coverage_amount, premium_frequency,
		        term_months, description, terms_conditions,
		        created_at, updated_at, deleted_at
		 FROM insurance_products
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list insurance products: %w", err)
	}
	defer rows.Close()
	return scanInsuranceProducts(rows)
}

func scanInsuranceProducts(rows *sql.Rows) ([]*insurance.Product, error) {
	var results []*insurance.Product
	for rows.Next() {
		var e insurance.Product
		if err := rows.Scan(
			&e.ID, &e.Name, &e.Code, &e.ProductType, &e.Status,
			&e.ProviderName, &e.PremiumAmount, &e.CoverageAmount, &e.PremiumFrequency,
			&e.TermMonths, &e.Description, &e.TermsConditions,
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
	_ insurance.ProductWriteRepository = (*InsuranceProductRepository)(nil)
	_ insurance.ProductReadRepository  = (*InsuranceProductRepository)(nil)
)
