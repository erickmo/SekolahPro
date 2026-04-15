package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/example"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ExampleRepository adalah concrete implementation dari example.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type ExampleRepository struct {
	db *sqlx.DB
}

// NewExampleRepository membuat instance baru ExampleRepository.
func NewExampleRepository(db *sqlx.DB) *ExampleRepository {
	return &ExampleRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity Example baru ke database.
func (r *ExampleRepository) Save(ctx context.Context, s scope.Scope, e *example.Example) error {
	const q = `
		INSERT INTO examples (id, tenant_id, company_id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.Name, e.Description, e.IsActive,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save example: %w", err)
	}
	return nil
}

// Update mengupdate Example yang sudah ada berdasarkan ID + scope.
func (r *ExampleRepository) Update(ctx context.Context, s scope.Scope, e *example.Example) error {
	const q = `
		UPDATE examples
		SET name = $4, description = $5, is_active = $6, updated_at = $7
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.Name, e.Description, e.IsActive, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update example: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return example.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *ExampleRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE examples SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete example: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return example.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu Example berdasarkan ID + scope.
func (r *ExampleRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*example.Example, error) {
	const q = `
		SELECT id, name, description, is_active, created_at, updated_at, deleted_at
		FROM examples
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e example.Example
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.Name, &e.Description, &e.IsActive, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, example.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get example by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar Example dengan pagination + sorting.
// sortBy + order sudah divalidasi oleh HTTP handler layer.
func (r *ExampleRepository) List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*example.Example, int, error) {
	total, err := r.countExamples(ctx, s)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectExamples(ctx, s, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// countExamples menghitung total Example aktif milik scope ini.
func (r *ExampleRepository) countExamples(ctx context.Context, s scope.Scope) (int, error) {
	const q = `SELECT COUNT(*) FROM examples WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count examples: %w", err)
	}
	return total, nil
}

// selectExamples mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *ExampleRepository) selectExamples(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*example.Example, error) {
	col := safeColumn(sortBy, map[string]string{"name": "name", "created_at": "created_at"}, "created_at")
	dir := safeOrder(order)
	q := fmt.Sprintf(
		`SELECT id, name, description, is_active, created_at, updated_at, deleted_at
		 FROM examples
		 WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
		 ORDER BY %s %s LIMIT $3 OFFSET $4`, col, dir,
	)
	rows, err := r.db.QueryContext(ctx, q, s.TenantID, s.CompanyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list examples: %w", err)
	}
	defer rows.Close()
	return scanExamples(rows)
}

func scanExamples(rows *sql.Rows) ([]*example.Example, error) {
	var results []*example.Example
	for rows.Next() {
		var e example.Example
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.IsActive, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// safeColumn memastikan nama kolom ada dalam whitelist untuk mencegah SQL injection.
func safeColumn(col string, allowed map[string]string, fallback string) string {
	if v, ok := allowed[col]; ok {
		return v
	}
	return fallback
}

// safeOrder memastikan hanya "asc" atau "desc" yang diterima.
func safeOrder(order string) string {
	if order == "asc" {
		return "ASC"
	}
	return "DESC"
}
