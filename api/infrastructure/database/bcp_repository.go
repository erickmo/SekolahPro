package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/bcp"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// BCPRepository adalah concrete implementation dari bcp.WriteRepository + bcp.ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type BCPRepository struct {
	db *sqlx.DB
}

// NewBCPRepository membuat instance baru BCPRepository.
func NewBCPRepository(db *sqlx.DB) *BCPRepository {
	return &BCPRepository{db: db}
}

// -- WriteRepository --

// Save menyimpan entity BackupStrategy baru ke database.
func (r *BCPRepository) Save(ctx context.Context, s scope.Scope, e *bcp.BackupStrategy) error {
	const q = `
		INSERT INTO backup_strategies (id, tenant_id, company_id, strategy_type, frequency,
			retention_days, storage_location, encryption_enabled, is_active,
			last_backup_at, next_backup_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.StrategyType, e.Frequency, e.RetentionDays,
		e.StorageLocation, e.EncryptionEnabled, e.IsActive,
		e.LastBackupAt, e.NextBackupAt,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save backup strategy: %w", err)
	}
	return nil
}

// Update mengupdate BackupStrategy yang sudah ada berdasarkan ID + scope.
func (r *BCPRepository) Update(ctx context.Context, s scope.Scope, e *bcp.BackupStrategy) error {
	const q = `
		UPDATE backup_strategies
		SET strategy_type = $4, frequency = $5, retention_days = $6,
		    storage_location = $7, encryption_enabled = $8, is_active = $9,
		    updated_at = $10
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.StrategyType, e.Frequency, e.RetentionDays,
		e.StorageLocation, e.EncryptionEnabled, e.IsActive,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update backup strategy: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return bcp.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *BCPRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE backup_strategies SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete backup strategy: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return bcp.ErrNotFound
	}
	return nil
}

// -- ReadRepository --

// GetByID mengambil satu BackupStrategy berdasarkan ID + scope.
func (r *BCPRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*bcp.BackupStrategy, error) {
	const q = `
		SELECT id, strategy_type, frequency, retention_days, storage_location,
		       encryption_enabled, is_active, last_backup_at, next_backup_at,
		       created_at, updated_at, deleted_at
		FROM backup_strategies
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e bcp.BackupStrategy
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.StrategyType, &e.Frequency, &e.RetentionDays,
		&e.StorageLocation, &e.EncryptionEnabled, &e.IsActive,
		&e.LastBackupAt, &e.NextBackupAt,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, bcp.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get backup strategy by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar BackupStrategy dengan pagination + sorting.
func (r *BCPRepository) List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*bcp.BackupStrategy, int, error) {
	total, err := r.countBackupStrategies(ctx, s)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectBackupStrategies(ctx, s, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// countBackupStrategies menghitung total BackupStrategy aktif milik scope ini.
func (r *BCPRepository) countBackupStrategies(ctx context.Context, s scope.Scope) (int, error) {
	const q = `SELECT COUNT(*) FROM backup_strategies WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count backup strategies: %w", err)
	}
	return total, nil
}

// selectBackupStrategies mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *BCPRepository) selectBackupStrategies(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*bcp.BackupStrategy, error) {
	col := safeColumn(sortBy, map[string]string{
		"strategy_type": "strategy_type",
		"frequency":     "frequency",
		"created_at":    "created_at",
	}, "created_at")
	dir := safeOrder(order)
	q := fmt.Sprintf(
		`SELECT id, strategy_type, frequency, retention_days, storage_location,
		        encryption_enabled, is_active, last_backup_at, next_backup_at,
		        created_at, updated_at, deleted_at
		 FROM backup_strategies
		 WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
		 ORDER BY %s %s LIMIT $3 OFFSET $4`, col, dir,
	)
	rows, err := r.db.QueryContext(ctx, q, s.TenantID, s.CompanyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list backup strategies: %w", err)
	}
	defer rows.Close()
	return scanBackupStrategies(rows)
}

func scanBackupStrategies(rows *sql.Rows) ([]*bcp.BackupStrategy, error) {
	var results []*bcp.BackupStrategy
	for rows.Next() {
		var e bcp.BackupStrategy
		if err := rows.Scan(
			&e.ID, &e.StrategyType, &e.Frequency, &e.RetentionDays,
			&e.StorageLocation, &e.EncryptionEnabled, &e.IsActive,
			&e.LastBackupAt, &e.NextBackupAt,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}
