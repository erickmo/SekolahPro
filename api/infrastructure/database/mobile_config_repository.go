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

	"github.com/yourorg/boilerplate/internal/domain/mobile_config"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// MobileConfigRepository adalah concrete implementation dari mobile_config.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type MobileConfigRepository struct {
	db *sqlx.DB
}

// NewMobileConfigRepository membuat instance baru MobileConfigRepository.
func NewMobileConfigRepository(db *sqlx.DB) *MobileConfigRepository {
	return &MobileConfigRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity MobileAppConfig baru ke database.
func (r *MobileConfigRepository) Save(ctx context.Context, s scope.Scope, e *mobile_config.MobileAppConfig) error {
	featureFlagsJSON, err := json.Marshal(e.FeatureFlags)
	if err != nil {
		return fmt.Errorf("marshal feature_flags: %w", err)
	}
	themeConfigJSON, err := json.Marshal(e.ThemeConfig)
	if err != nil {
		return fmt.Errorf("marshal theme_config: %w", err)
	}
	offlineConfigJSON, err := json.Marshal(e.OfflineConfig)
	if err != nil {
		return fmt.Errorf("marshal offline_config: %w", err)
	}

	const q = `
		INSERT INTO mobile_app_configs (
			id, tenant_id, company_id, app_variant, platform,
			min_version, current_version, force_update, maintenance_mode,
			feature_flags, api_base_url, theme_config, offline_config,
			is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err = r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.AppVariant, e.Platform,
		e.MinVersion, e.CurrentVersion, e.ForceUpdate, e.MaintenanceMode,
		featureFlagsJSON, e.APIBaseURL, themeConfigJSON, offlineConfigJSON,
		e.IsActive, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "uq_mobile_app_configs_tenant_variant") {
			return mobile_config.ErrDuplicateVariant
		}
		return fmt.Errorf("save mobile_config: %w", err)
	}
	return nil
}

// Update mengupdate MobileAppConfig yang sudah ada berdasarkan ID + scope.
func (r *MobileConfigRepository) Update(ctx context.Context, s scope.Scope, e *mobile_config.MobileAppConfig) error {
	featureFlagsJSON, err := json.Marshal(e.FeatureFlags)
	if err != nil {
		return fmt.Errorf("marshal feature_flags: %w", err)
	}
	themeConfigJSON, err := json.Marshal(e.ThemeConfig)
	if err != nil {
		return fmt.Errorf("marshal theme_config: %w", err)
	}
	offlineConfigJSON, err := json.Marshal(e.OfflineConfig)
	if err != nil {
		return fmt.Errorf("marshal offline_config: %w", err)
	}

	const q = `
		UPDATE mobile_app_configs
		SET app_variant = $4, platform = $5,
		    min_version = $6, current_version = $7,
		    force_update = $8, maintenance_mode = $9,
		    feature_flags = $10, api_base_url = $11,
		    theme_config = $12, offline_config = $13,
		    is_active = $14, updated_at = $15
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.AppVariant, e.Platform,
		e.MinVersion, e.CurrentVersion, e.ForceUpdate, e.MaintenanceMode,
		featureFlagsJSON, e.APIBaseURL, themeConfigJSON, offlineConfigJSON,
		e.IsActive, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update mobile_config: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return mobile_config.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *MobileConfigRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE mobile_app_configs SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete mobile_config: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return mobile_config.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu MobileAppConfig berdasarkan ID + scope.
func (r *MobileConfigRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*mobile_config.MobileAppConfig, error) {
	const q = `
		SELECT id, app_variant, platform, min_version, current_version,
		       force_update, maintenance_mode, feature_flags, api_base_url,
		       theme_config, offline_config, is_active, created_at, updated_at, deleted_at
		FROM mobile_app_configs
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e mobile_config.MobileAppConfig
	var featureFlagsJSON, themeConfigJSON, offlineConfigJSON []byte

	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.AppVariant, &e.Platform, &e.MinVersion, &e.CurrentVersion,
		&e.ForceUpdate, &e.MaintenanceMode, &featureFlagsJSON, &e.APIBaseURL,
		&themeConfigJSON, &offlineConfigJSON, &e.IsActive, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, mobile_config.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get mobile_config by id: %w", err)
	}

	if err := e.FeatureFlags.Scan(featureFlagsJSON); err != nil {
		return nil, fmt.Errorf("scan feature_flags: %w", err)
	}
	if err := e.ThemeConfig.Scan(themeConfigJSON); err != nil {
		return nil, fmt.Errorf("scan theme_config: %w", err)
	}
	if err := e.OfflineConfig.Scan(offlineConfigJSON); err != nil {
		return nil, fmt.Errorf("scan offline_config: %w", err)
	}

	return &e, nil
}

// List mengambil daftar MobileAppConfig dengan pagination, sorting, dan filter opsional.
func (r *MobileConfigRepository) List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string, filters mobile_config.ListFilters) ([]*mobile_config.MobileAppConfig, int, error) {
	total, err := r.countMobileConfigs(ctx, s, filters)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectMobileConfigs(ctx, s, limit, offset, sortBy, order, filters)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// countMobileConfigs menghitung total MobileAppConfig aktif milik scope ini.
func (r *MobileConfigRepository) countMobileConfigs(ctx context.Context, s scope.Scope, filters mobile_config.ListFilters) (int, error) {
	query := `SELECT COUNT(*) FROM mobile_app_configs WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`
	args := []any{s.TenantID, s.CompanyID}
	argIdx := 3

	query, args = appendFilter(query, args, &argIdx, filters)

	var total int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count mobile_configs: %w", err)
	}
	return total, nil
}

// selectMobileConfigs mengambil baris dengan ORDER BY + LIMIT/OFFSET + dynamic WHERE.
func (r *MobileConfigRepository) selectMobileConfigs(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string, filters mobile_config.ListFilters) ([]*mobile_config.MobileAppConfig, error) {
	col := safeColumn(sortBy, map[string]string{
		"app_variant":     "app_variant",
		"platform":        "platform",
		"created_at":      "created_at",
		"current_version": "current_version",
	}, "created_at")
	dir := safeOrder(order)

	query := fmt.Sprintf(
		`SELECT id, app_variant, platform, min_version, current_version,
		        force_update, maintenance_mode, feature_flags, api_base_url,
		        theme_config, offline_config, is_active, created_at, updated_at, deleted_at
		 FROM mobile_app_configs
		 WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
			 ORDER BY %s %s`, col, dir,
	)
	args := []any{s.TenantID, s.CompanyID}
	argIdx := 3

	query, args = appendFilter(query, args, &argIdx, filters)

	query += fmt.Sprintf(` ORDER BY %s %s LIMIT $%d OFFSET $%d`, col, dir, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list mobile_configs: %w", err)
	}
	defer rows.Close()
	return scanMobileConfigs(rows)
}

// appendFilter menambahkan WHERE clause dinamis berdasarkan filter yang diisi.
func appendFilter(query string, args []any, argIdx *int, filters mobile_config.ListFilters) (string, []any) {
	if filters.AppVariant != "" {
		query += fmt.Sprintf(` AND app_variant = $%d`, *argIdx)
		args = append(args, string(filters.AppVariant))
		*argIdx++
	}
	if filters.Platform != "" {
		query += fmt.Sprintf(` AND platform = $%d`, *argIdx)
		args = append(args, string(filters.Platform))
		*argIdx++
	}
	return query, args
}

func scanMobileConfigs(rows *sql.Rows) ([]*mobile_config.MobileAppConfig, error) {
	var results []*mobile_config.MobileAppConfig
	for rows.Next() {
		var e mobile_config.MobileAppConfig
		var featureFlagsJSON, themeConfigJSON, offlineConfigJSON []byte

		if err := rows.Scan(
			&e.ID, &e.AppVariant, &e.Platform, &e.MinVersion, &e.CurrentVersion,
			&e.ForceUpdate, &e.MaintenanceMode, &featureFlagsJSON, &e.APIBaseURL,
			&themeConfigJSON, &offlineConfigJSON, &e.IsActive, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}

		if err := e.FeatureFlags.Scan(featureFlagsJSON); err != nil {
			return nil, fmt.Errorf("scan feature_flags: %w", err)
		}
		if err := e.ThemeConfig.Scan(themeConfigJSON); err != nil {
			return nil, fmt.Errorf("scan theme_config: %w", err)
		}
		if err := e.OfflineConfig.Scan(offlineConfigJSON); err != nil {
			return nil, fmt.Errorf("scan offline_config: %w", err)
		}

		results = append(results, &e)
	}
	return results, rows.Err()
}
