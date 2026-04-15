package vernon

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// BaseRepository menyediakan generic CRUD untuk semua domain Vernon.
// Semua query otomatis difilter oleh tenant_id + company_id dari Scope.
type BaseRepository struct {
	db        *sqlx.DB
	tableName string
}

// NewBaseRepository membuat instance baru BaseRepository.
func NewBaseRepository(db *sqlx.DB, tableName string) *BaseRepository {
	return &BaseRepository{db: db, tableName: tableName}
}

// FindByID mengambil satu entitas berdasarkan ID dengan scope filtering.
func (r *BaseRepository) FindByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*BaseDomain, error) {
	q := fmt.Sprintf(
		`SELECT id, tenant_id, company_id, _rels, _data, _sync_status, _sync_version,
		        created_at, updated_at, deleted_at
		 FROM %s WHERE id = $1 AND tenant_id = $2 AND company_id = $3 AND deleted_at IS NULL`,
		r.tableName,
	)
	row := r.db.QueryRowContext(ctx, q, id, s.TenantID, s.CompanyID)
	return scanRow(row)
}

// FindAll mengambil daftar entitas dengan filter, sort, dan paginasi.
// Dibagi menjadi 3 fungsi helper agar setiap fungsi ≤ 40 baris.
func (r *BaseRepository) FindAll(ctx context.Context, s scope.Scope, p FindParams) ([]BaseDomain, int64, error) {
	where, args := BuildWhereClause(s.TenantID.String(), s.CompanyID.String(), p.Filters)
	total, err := r.countRows(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectRows(ctx, where, args, p)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// countRows menghitung total baris yang cocok dengan WHERE clause.
func (r *BaseRepository) countRows(ctx context.Context, where string, args []any) (int64, error) {
	q := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", r.tableName, where)
	var total int64
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count %s: %w", r.tableName, err)
	}
	return total, nil
}

// selectRows mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *BaseRepository) selectRows(ctx context.Context, where string, args []any, p FindParams) ([]BaseDomain, error) {
	order := BuildOrderClause(p.Sort)
	limit, offset := paginationCalc(p)
	argN := len(args) + 1
	q := fmt.Sprintf(
		`SELECT id, tenant_id, company_id, _rels, _data, _sync_status, _sync_version,
		        created_at, updated_at, deleted_at
		 FROM %s WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		r.tableName, where, order, argN, argN+1,
	)
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("select %s: %w", r.tableName, err)
	}
	defer rows.Close()
	return scanRows(rows)
}

// Create menyisipkan entitas baru dalam transaksi yang diberikan.
func (r *BaseRepository) Create(ctx context.Context, tx *sqlx.Tx, s scope.Scope, rels map[string]RelDef, data map[string]any) (*BaseDomain, error) {
	id := uuid.New()
	now := time.Now().UTC()
	relsBytes, _ := json.Marshal(rels)
	dataBytes, _ := json.Marshal(data)

	q := fmt.Sprintf(
		`INSERT INTO %s (id, tenant_id, company_id, _rels, _data, _sync_status, _sync_version, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 0, $7, $8)
		 RETURNING id, tenant_id, company_id, _rels, _data, _sync_status, _sync_version, created_at, updated_at, deleted_at`,
		r.tableName,
	)
	row := tx.QueryRowContext(ctx, q, id, s.TenantID, s.CompanyID, relsBytes, dataBytes, SyncStatusSynced, now, now)
	return scanRow(row)
}

// Update mengganti seluruh _data entitas dalam transaksi.
func (r *BaseRepository) Update(ctx context.Context, tx *sqlx.Tx, s scope.Scope, id uuid.UUID, data map[string]any) (*BaseDomain, error) {
	dataBytes, _ := json.Marshal(data)
	q := fmt.Sprintf(
		`UPDATE %s SET _data = $1, _sync_version = _sync_version + 1, updated_at = now()
		 WHERE id = $2 AND tenant_id = $3 AND company_id = $4 AND deleted_at IS NULL
		 RETURNING id, tenant_id, company_id, _rels, _data, _sync_status, _sync_version, created_at, updated_at, deleted_at`,
		r.tableName,
	)
	row := tx.QueryRowContext(ctx, q, dataBytes, id, s.TenantID, s.CompanyID)
	return scanRow(row)
}

// Patch melakukan merge partial update pada _data menggunakan operator PostgreSQL ||.
func (r *BaseRepository) Patch(ctx context.Context, tx *sqlx.Tx, s scope.Scope, id uuid.UUID, partial map[string]any) (*BaseDomain, error) {
	partialBytes, _ := json.Marshal(partial)
	q := fmt.Sprintf(
		`UPDATE %s SET _data = _data || $1::jsonb, _sync_version = _sync_version + 1, updated_at = now()
		 WHERE id = $2 AND tenant_id = $3 AND company_id = $4 AND deleted_at IS NULL
		 RETURNING id, tenant_id, company_id, _rels, _data, _sync_status, _sync_version, created_at, updated_at, deleted_at`,
		r.tableName,
	)
	row := tx.QueryRowContext(ctx, q, partialBytes, id, s.TenantID, s.CompanyID)
	return scanRow(row)
}

// SoftDelete menandai entitas sebagai deleted (deleted_at = now()).
func (r *BaseRepository) SoftDelete(ctx context.Context, tx *sqlx.Tx, s scope.Scope, id uuid.UUID) error {
	q := fmt.Sprintf(
		`UPDATE %s SET deleted_at = now(), updated_at = now()
		 WHERE id = $1 AND tenant_id = $2 AND company_id = $3 AND deleted_at IS NULL`,
		r.tableName,
	)
	res, err := tx.ExecContext(ctx, q, id, s.TenantID, s.CompanyID)
	if err != nil {
		return fmt.Errorf("soft delete %s %s: %w", r.tableName, id, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// FindByDataField mencari entitas berdasarkan satu field di dalam _data.
func (r *BaseRepository) FindByDataField(ctx context.Context, s scope.Scope, field, value string) ([]BaseDomain, error) {
	where := fmt.Sprintf(
		"deleted_at IS NULL AND tenant_id = $1 AND company_id = $2 AND _data->>'%s' = $3", field,
	)
	args := []any{s.TenantID, s.CompanyID, value}
	return r.selectRows(ctx, where, args, FindParams{Limit: 500, Page: 1})
}

// CASUpdate update field relasi di dalam _data menggunakan Compare-And-Swap.
// Hanya update jika _sync_version masih sama (cegah overwrite concurrent).
func (r *BaseRepository) CASUpdate(ctx context.Context, s scope.Scope, id uuid.UUID, expectedVer int64, relName string, newValue any) error {
	valueBytes, _ := json.Marshal(newValue)
	q := fmt.Sprintf(
		`UPDATE %s
		 SET _data = jsonb_set(_data, '{%s}', $1::jsonb),
		     _sync_version = _sync_version + 1,
		     _sync_status = '%s',
		     updated_at = now()
		 WHERE id = $2 AND tenant_id = $3 AND company_id = $4
		   AND _sync_version = $5 AND deleted_at IS NULL`,
		r.tableName, relName, SyncStatusSynced,
	)
	res, err := r.db.ExecContext(ctx, q, valueBytes, id, s.TenantID, s.CompanyID, expectedVer)
	if err != nil {
		return fmt.Errorf("cas update %s %s: %w", r.tableName, id, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrVersionConflict
	}
	return nil
}

// ── scan helpers ─────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanRow(row scanner) (*BaseDomain, error) {
	var e BaseDomain
	var relsBytes, dataBytes []byte
	err := row.Scan(
		&e.ID, &e.TenantID, &e.CompanyID,
		&relsBytes, &dataBytes,
		&e.SyncStatus, &e.SyncVersion,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(relsBytes, &e.Rels)  //nolint:errcheck
	json.Unmarshal(dataBytes, &e.Data)  //nolint:errcheck
	return &e, nil
}

func scanRows(rows *sql.Rows) ([]BaseDomain, error) {
	var results []BaseDomain
	for rows.Next() {
		var e BaseDomain
		var relsBytes, dataBytes []byte
		err := rows.Scan(
			&e.ID, &e.TenantID, &e.CompanyID,
			&relsBytes, &dataBytes,
			&e.SyncStatus, &e.SyncVersion,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(relsBytes, &e.Rels) //nolint:errcheck
		json.Unmarshal(dataBytes, &e.Data) //nolint:errcheck
		results = append(results, e)
	}
	return results, rows.Err()
}

func paginationCalc(p FindParams) (limit, offset int) {
	limit = p.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	page := p.Page
	if page <= 0 {
		page = 1
	}
	return limit, (page - 1) * limit
}
