package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/yourorg/boilerplate/internal/domain/incident"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// IncidentRepository adalah concrete implementation dari incident.WriteRepository + incident.ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type IncidentRepository struct {
	db *sqlx.DB
}

// NewIncidentRepository membuat instance baru IncidentRepository.
func NewIncidentRepository(db *sqlx.DB) *IncidentRepository {
	return &IncidentRepository{db: db}
}

// -- WriteRepository --

// Save menyimpan entity IncidentRecord baru ke database.
func (r *IncidentRepository) Save(ctx context.Context, s scope.Scope, e *incident.IncidentRecord) error {
	const q = `
		INSERT INTO incidents (id, tenant_id, company_id, severity, status,
		    incident_type, title, description, affected_systems,
		    root_cause, resolution, resolution_time, resolved_by, resolved_at,
		    created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.Severity, e.Status, e.IncidentType, e.Title, e.Description,
		pq.Array(e.AffectedSystems),
		e.RootCause, e.Resolution, e.ResolutionTime, e.ResolvedBy, e.ResolvedAt,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save incident: %w", err)
	}
	return nil
}

// Update mengupdate IncidentRecord yang sudah ada berdasarkan ID + scope.
func (r *IncidentRepository) Update(ctx context.Context, s scope.Scope, e *incident.IncidentRecord) error {
	const q = `
		UPDATE incidents
		SET severity = $4, status = $5, incident_type = $6, title = $7,
		    description = $8, affected_systems = $9,
		    root_cause = $10, resolution = $11, resolution_time = $12,
		    resolved_by = $13, resolved_at = $14, updated_at = $15
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.Severity, e.Status, e.IncidentType, e.Title, e.Description,
		pq.Array(e.AffectedSystems),
		e.RootCause, e.Resolution, e.ResolutionTime, e.ResolvedBy, e.ResolvedAt,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update incident: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return incident.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *IncidentRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE incidents SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete incident: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return incident.ErrNotFound
	}
	return nil
}

// -- ReadRepository --

// GetByID mengambil satu IncidentRecord berdasarkan ID + scope.
func (r *IncidentRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*incident.IncidentRecord, error) {
	const q = `
		SELECT id, severity, status, incident_type, title, description,
		       affected_systems, root_cause, resolution, resolution_time,
		       resolved_by, resolved_at, created_at, updated_at, deleted_at
		FROM incidents
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e incident.IncidentRecord
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.Severity, &e.Status, &e.IncidentType, &e.Title, &e.Description,
		pq.Array(&e.AffectedSystems),
		&e.RootCause, &e.Resolution, &e.ResolutionTime,
		&e.ResolvedBy, &e.ResolvedAt,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, incident.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get incident by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar IncidentRecord dengan pagination + sorting.
func (r *IncidentRepository) List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*incident.IncidentRecord, int, error) {
	total, err := r.countIncidents(ctx, s)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectIncidents(ctx, s, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// countIncidents menghitung total IncidentRecord aktif milik scope ini.
func (r *IncidentRepository) countIncidents(ctx context.Context, s scope.Scope) (int, error) {
	const q = `SELECT COUNT(*) FROM incidents WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count incidents: %w", err)
	}
	return total, nil
}

// selectIncidents mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *IncidentRepository) selectIncidents(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*incident.IncidentRecord, error) {
	col := safeColumn(sortBy, map[string]string{
		"severity":    "severity",
		"status":      "status",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
	}, "created_at")
	dir := safeOrder(order)
	q := fmt.Sprintf(
		`SELECT id, severity, status, incident_type, title, description,
		        affected_systems, root_cause, resolution, resolution_time,
		        resolved_by, resolved_at, created_at, updated_at, deleted_at
		 FROM incidents
		 WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
		 ORDER BY %s %s LIMIT $3 OFFSET $4`, col, dir,
	)
	rows, err := r.db.QueryContext(ctx, q, s.TenantID, s.CompanyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list incidents: %w", err)
	}
	defer rows.Close()
	return scanIncidents(rows)
}

func scanIncidents(rows *sql.Rows) ([]*incident.IncidentRecord, error) {
	var results []*incident.IncidentRecord
	for rows.Next() {
		var e incident.IncidentRecord
		if err := rows.Scan(
			&e.ID, &e.Severity, &e.Status, &e.IncidentType, &e.Title, &e.Description,
			pq.Array(&e.AffectedSystems),
			&e.RootCause, &e.Resolution, &e.ResolutionTime,
			&e.ResolvedBy, &e.ResolvedAt,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}
