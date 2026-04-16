package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// KPIRepository adalah concrete implementation dari semua repository interfaces
// domain kpi: KPIDefinition, KPIMeasurement, dan StressTestScenario.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type KPIRepository struct {
	db *sqlx.DB
}

// NewKPIRepository membuat instance baru KPIRepository.
func NewKPIRepository(db *sqlx.DB) *KPIRepository {
	return &KPIRepository{db: db}
}

// ── KPIDefinition WriteRepository ─────────────────────────────────────────────

// Save menyimpan entity KPIDefinition baru ke database.
func (r *KPIRepository) Save(ctx context.Context, s scope.Scope, e *kpi.KPIDefinition) error {
	const q = `
		INSERT INTO kpi_definitions (
			id, tenant_id, company_id, name, code, category, frequency,
			unit, target_value, warning_threshold, critical_threshold,
			is_active, description, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.Name, e.Code, e.Category, e.Frequency,
		e.Unit, e.TargetValue, e.WarningThreshold, e.CriticalThreshold,
		e.IsActive, e.Description,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save kpi definition: %w", err)
	}
	return nil
}

// Update mengupdate KPIDefinition yang sudah ada berdasarkan ID + scope.
func (r *KPIRepository) Update(ctx context.Context, s scope.Scope, e *kpi.KPIDefinition) error {
	const q = `
		UPDATE kpi_definitions
		SET name = $4, code = $5, category = $6, frequency = $7,
		    unit = $8, target_value = $9, warning_threshold = $10, critical_threshold = $11,
		    is_active = $12, description = $13, updated_at = $14
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.Name, e.Code, e.Category, e.Frequency,
		e.Unit, e.TargetValue, e.WarningThreshold, e.CriticalThreshold,
		e.IsActive, e.Description,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update kpi definition: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kpi.ErrDefinitionNotFound
	}
	return nil
}

// Delete melakukan soft delete pada KPIDefinition.
func (r *KPIRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE kpi_definitions SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete kpi definition: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kpi.ErrDefinitionNotFound
	}
	return nil
}

// ── KPIDefinition ReadRepository ──────────────────────────────────────────────

// GetByID mengambil satu KPIDefinition berdasarkan ID + scope.
func (r *KPIRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*kpi.KPIDefinition, error) {
	const q = `
		SELECT id, name, code, category, frequency, unit,
		       target_value, warning_threshold, critical_threshold,
		       is_active, description, created_at, updated_at, deleted_at
		FROM kpi_definitions
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e kpi.KPIDefinition
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.Name, &e.Code, &e.Category, &e.Frequency, &e.Unit,
		&e.TargetValue, &e.WarningThreshold, &e.CriticalThreshold,
		&e.IsActive, &e.Description,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, kpi.ErrDefinitionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get kpi definition by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar KPIDefinition dengan filter, pagination, dan sorting.
func (r *KPIRepository) List(ctx context.Context, s scope.Scope, filter kpi.KPIDefinitionFilter, limit, offset int, sortBy, order string) ([]*kpi.KPIDefinition, int, error) {
	where, args := r.buildDefinitionWhereClause(s, filter)

	total, err := r.countDefinitions(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectDefinitions(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *KPIRepository) buildDefinitionWhereClause(s scope.Scope, f kpi.KPIDefinitionFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if f.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", paramIdx))
		args = append(args, string(*f.Category))
		paramIdx++
	}
	if f.Frequency != nil {
		conditions = append(conditions, fmt.Sprintf("frequency = $%d", paramIdx))
		args = append(args, string(*f.Frequency))
		paramIdx++
	}
	if f.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", paramIdx))
		args = append(args, *f.IsActive)
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

func (r *KPIRepository) countDefinitions(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM kpi_definitions WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count kpi definitions: %w", err)
	}
	return total, nil
}

func (r *KPIRepository) selectDefinitions(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*kpi.KPIDefinition, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at": "created_at",
		"name":       "name",
		"category":   "category",
		"frequency":  "frequency",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, name, code, category, frequency, unit,
		        target_value, warning_threshold, critical_threshold,
		        is_active, description, created_at, updated_at, deleted_at
		 FROM kpi_definitions
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list kpi definitions: %w", err)
	}
	defer rows.Close()
	return scanKPIDefinitions(rows)
}

func scanKPIDefinitions(rows *sql.Rows) ([]*kpi.KPIDefinition, error) {
	var results []*kpi.KPIDefinition
	for rows.Next() {
		var e kpi.KPIDefinition
		if err := rows.Scan(
			&e.ID, &e.Name, &e.Code, &e.Category, &e.Frequency, &e.Unit,
			&e.TargetValue, &e.WarningThreshold, &e.CriticalThreshold,
			&e.IsActive, &e.Description,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// ── KPIMeasurement WriteRepository ────────────────────────────────────────────

// SaveMeasurement menyimpan entity KPIMeasurement baru ke database.
func (r *KPIRepository) SaveMeasurement(ctx context.Context, s scope.Scope, e *kpi.KPIMeasurement) error {
	const q = `
		INSERT INTO kpi_measurements (
			id, tenant_id, company_id, definition_id, measured_value, status,
			measured_at, measured_by, notes, period_start, period_end,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.DefinitionID, e.MeasuredValue, e.Status,
		e.MeasuredAt, e.MeasuredBy, e.Notes, e.PeriodStart, e.PeriodEnd,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save kpi measurement: %w", err)
	}
	return nil
}

// UpdateMeasurement mengupdate KPIMeasurement yang sudah ada.
func (r *KPIRepository) UpdateMeasurement(ctx context.Context, s scope.Scope, e *kpi.KPIMeasurement) error {
	const q = `
		UPDATE kpi_measurements
		SET definition_id = $4, measured_value = $5, status = $6,
		    measured_at = $7, measured_by = $8, notes = $9,
		    period_start = $10, period_end = $11, updated_at = $12
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.DefinitionID, e.MeasuredValue, e.Status,
		e.MeasuredAt, e.MeasuredBy, e.Notes, e.PeriodStart, e.PeriodEnd,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update kpi measurement: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kpi.ErrMeasurementNotFound
	}
	return nil
}

// DeleteMeasurement melakukan soft delete pada KPIMeasurement.
func (r *KPIRepository) DeleteMeasurement(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE kpi_measurements SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete kpi measurement: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kpi.ErrMeasurementNotFound
	}
	return nil
}

// ── KPIMeasurement ReadRepository ─────────────────────────────────────────────

// GetMeasurementByID mengambil satu KPIMeasurement berdasarkan ID + scope.
func (r *KPIRepository) GetMeasurementByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*kpi.KPIMeasurement, error) {
	const q = `
		SELECT id, definition_id, measured_value, status,
		       measured_at, measured_by, notes, period_start, period_end,
		       created_at, updated_at, deleted_at
		FROM kpi_measurements
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e kpi.KPIMeasurement
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.DefinitionID, &e.MeasuredValue, &e.Status,
		&e.MeasuredAt, &e.MeasuredBy, &e.Notes, &e.PeriodStart, &e.PeriodEnd,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, kpi.ErrMeasurementNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get kpi measurement by id: %w", err)
	}
	return &e, nil
}

// ListMeasurements mengambil daftar KPIMeasurement dengan filter, pagination, dan sorting.
func (r *KPIRepository) ListMeasurements(ctx context.Context, s scope.Scope, filter kpi.KPIMeasurementFilter, limit, offset int, sortBy, order string) ([]*kpi.KPIMeasurement, int, error) {
	where, args := r.buildMeasurementWhereClause(s, filter)

	total, err := r.countMeasurements(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectMeasurements(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *KPIRepository) buildMeasurementWhereClause(s scope.Scope, f kpi.KPIMeasurementFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if f.DefinitionID != nil {
		conditions = append(conditions, fmt.Sprintf("definition_id = $%d", paramIdx))
		args = append(args, *f.DefinitionID)
		paramIdx++
	}
	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", paramIdx))
		args = append(args, string(*f.Status))
		paramIdx++
	}
	if f.MeasuredAtFrom != nil {
		conditions = append(conditions, fmt.Sprintf("measured_at >= $%d", paramIdx))
		args = append(args, *f.MeasuredAtFrom)
		paramIdx++
	}
	if f.MeasuredAtTo != nil {
		conditions = append(conditions, fmt.Sprintf("measured_at <= $%d", paramIdx))
		args = append(args, *f.MeasuredAtTo)
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

func (r *KPIRepository) countMeasurements(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM kpi_measurements WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count kpi measurements: %w", err)
	}
	return total, nil
}

func (r *KPIRepository) selectMeasurements(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*kpi.KPIMeasurement, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":   "created_at",
		"measured_at":  "measured_at",
		"definition_id": "definition_id",
		"status":       "status",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, definition_id, measured_value, status,
		        measured_at, measured_by, notes, period_start, period_end,
		        created_at, updated_at, deleted_at
		 FROM kpi_measurements
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list kpi measurements: %w", err)
	}
	defer rows.Close()
	return scanKPIMeasurements(rows)
}

func scanKPIMeasurements(rows *sql.Rows) ([]*kpi.KPIMeasurement, error) {
	var results []*kpi.KPIMeasurement
	for rows.Next() {
		var e kpi.KPIMeasurement
		if err := rows.Scan(
			&e.ID, &e.DefinitionID, &e.MeasuredValue, &e.Status,
			&e.MeasuredAt, &e.MeasuredBy, &e.Notes, &e.PeriodStart, &e.PeriodEnd,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// ── StressTestScenario WriteRepository ────────────────────────────────────────

// SaveScenario menyimpan entity StressTestScenario baru ke database.
func (r *KPIRepository) SaveScenario(ctx context.Context, s scope.Scope, e *kpi.StressTestScenario) error {
	const q = `
		INSERT INTO stress_test_scenarios (
			id, tenant_id, company_id, name, description, scenario_status,
			parameters, results, simulated_at, simulated_by, completed_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.Name, e.Description, e.ScenarioStatus,
		e.Parameters, e.Results, e.SimulatedAt, e.SimulatedBy, e.CompletedAt,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save stress test scenario: %w", err)
	}
	return nil
}

// UpdateScenario mengupdate StressTestScenario yang sudah ada.
func (r *KPIRepository) UpdateScenario(ctx context.Context, s scope.Scope, e *kpi.StressTestScenario) error {
	const q = `
		UPDATE stress_test_scenarios
		SET name = $4, description = $5, scenario_status = $6,
		    parameters = $7, results = $8, simulated_at = $9, simulated_by = $10,
		    completed_at = $11, updated_at = $12
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.Name, e.Description, e.ScenarioStatus,
		e.Parameters, e.Results, e.SimulatedAt, e.SimulatedBy, e.CompletedAt,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update stress test scenario: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kpi.ErrScenarioNotFound
	}
	return nil
}

// DeleteScenario melakukan soft delete pada StressTestScenario.
func (r *KPIRepository) DeleteScenario(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE stress_test_scenarios SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete stress test scenario: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kpi.ErrScenarioNotFound
	}
	return nil
}

// ── StressTestScenario ReadRepository ─────────────────────────────────────────

// GetScenarioByID mengambil satu StressTestScenario berdasarkan ID + scope.
func (r *KPIRepository) GetScenarioByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*kpi.StressTestScenario, error) {
	const q = `
		SELECT id, name, description, scenario_status,
		       parameters, results, simulated_at, simulated_by, completed_at,
		       created_at, updated_at, deleted_at
		FROM stress_test_scenarios
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e kpi.StressTestScenario
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.Name, &e.Description, &e.ScenarioStatus,
		&e.Parameters, &e.Results, &e.SimulatedAt, &e.SimulatedBy, &e.CompletedAt,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, kpi.ErrScenarioNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get stress test scenario by id: %w", err)
	}
	return &e, nil
}

// ListScenarios mengambil daftar StressTestScenario dengan filter, pagination, dan sorting.
func (r *KPIRepository) ListScenarios(ctx context.Context, s scope.Scope, filter kpi.StressTestFilter, limit, offset int, sortBy, order string) ([]*kpi.StressTestScenario, int, error) {
	where, args := r.buildScenarioWhereClause(s, filter)

	total, err := r.countScenarios(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectScenarios(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *KPIRepository) buildScenarioWhereClause(s scope.Scope, f kpi.StressTestFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if f.ScenarioStatus != nil {
		conditions = append(conditions, fmt.Sprintf("scenario_status = $%d", paramIdx))
		args = append(args, string(*f.ScenarioStatus))
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

func (r *KPIRepository) countScenarios(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM stress_test_scenarios WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count stress test scenarios: %w", err)
	}
	return total, nil
}

func (r *KPIRepository) selectScenarios(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*kpi.StressTestScenario, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":      "created_at",
		"scenario_status": "scenario_status",
		"name":            "name",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, name, description, scenario_status,
		        parameters, results, simulated_at, simulated_by, completed_at,
		        created_at, updated_at, deleted_at
		 FROM stress_test_scenarios
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list stress test scenarios: %w", err)
	}
	defer rows.Close()
	return scanStressTestScenarios(rows)
}

func scanStressTestScenarios(rows *sql.Rows) ([]*kpi.StressTestScenario, error) {
	var results []*kpi.StressTestScenario
	for rows.Next() {
		var e kpi.StressTestScenario
		if err := rows.Scan(
			&e.ID, &e.Name, &e.Description, &e.ScenarioStatus,
			&e.Parameters, &e.Results, &e.SimulatedAt, &e.SimulatedBy, &e.CompletedAt,
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
	_ kpi.KPIDefinitionWriteRepository           = (*KPIRepository)(nil)
	_ kpi.KPIDefinitionReadRepository            = (*KPIRepository)(nil)
	_ kpi.KPIMeasurementWriteRepository          = (*KPIRepository)(nil)
	_ kpi.KPIMeasurementReadRepository           = (*KPIRepository)(nil)
	_ kpi.StressTestWriteRepository              = (*KPIRepository)(nil)
	_ kpi.StressTestReadRepository               = (*KPIRepository)(nil)
)
