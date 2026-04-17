package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/role"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// RoleRepository adalah concrete implementation dari role.RoleWriteRepository,
// role.RoleReadRepository, role.UserRoleWriteRepository, dan role.UserRoleReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type RoleRepository struct {
	db *sqlx.DB
}

// NewRoleRepository membuat instance baru RoleRepository.
func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Compile-time interface checks.
var (
	_ role.RoleWriteRepository  = (*RoleRepository)(nil)
	_ role.RoleReadRepository   = (*RoleRepository)(nil)
	_ role.UserRoleWriteRepository = (*RoleRepository)(nil)
	_ role.UserRoleReadRepository  = (*RoleRepository)(nil)
)

// ── RoleWriteRepository ───────────────────────────────────────────────────────

// Save menyimpan entity Role baru ke database.
func (r *RoleRepository) Save(ctx context.Context, s scope.Scope, rl *role.Role) error {
	const q = `
		INSERT INTO roles (id, tenant_id, company_id, name, code, description,
		                   role_type, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, q,
		rl.ID, s.TenantID, s.CompanyID,
		rl.Name, rl.Code, rl.Description,
		rl.RoleType, rl.IsSystem,
		rl.CreatedAt, rl.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save role: %w", err)
	}
	return nil
}

// Update mengupdate Role yang sudah ada berdasarkan ID + scope.
func (r *RoleRepository) Update(ctx context.Context, s scope.Scope, rl *role.Role) error {
	const q = `
		UPDATE roles
		SET name = $4, description = $5, role_type = $6, updated_at = $7
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, rl.ID,
		rl.Name, rl.Description, rl.RoleType, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return role.ErrNotFound
	}
	return nil
}

// SoftDelete melakukan soft delete pada Role (mengisi deleted_at).
func (r *RoleRepository) SoftDelete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE roles SET deleted_at = $4, updated_at = $4
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3
		  AND deleted_at IS NULL AND is_system = false`

	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id, now)
	if err != nil {
		return fmt.Errorf("soft delete role: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return role.ErrNotFound
	}
	return nil
}

// ── RoleReadRepository ────────────────────────────────────────────────────────

// GetByID mengambil satu Role berdasarkan ID + scope.
func (r *RoleRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*role.Role, error) {
	const q = `
		SELECT r.id, r.tenant_id, r.company_id, r.name, r.code, r.description,
		       r.role_type, r.is_system, r.created_at, r.updated_at, r.deleted_at
		FROM roles r
		WHERE r.tenant_id = $1 AND r.company_id = $2 AND r.id = $3 AND r.deleted_at IS NULL`

	var rl role.Role
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&rl.ID, &rl.TenantID, &rl.CompanyID, &rl.Name, &rl.Code, &rl.Description,
		&rl.RoleType, &rl.IsSystem, &rl.CreatedAt, &rl.UpdatedAt, &rl.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, role.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get role by id: %w", err)
	}

	perms, err := r.loadPermissions(ctx, rl.ID)
	if err != nil {
		return nil, err
	}
	rl.Permissions = perms

	return &rl, nil
}

// GetByCode mengambil satu Role berdasarkan code + scope.
func (r *RoleRepository) GetByCode(ctx context.Context, s scope.Scope, code string) (*role.Role, error) {
	const q = `
		SELECT r.id, r.tenant_id, r.company_id, r.name, r.code, r.description,
		       r.role_type, r.is_system, r.created_at, r.updated_at, r.deleted_at
		FROM roles r
		WHERE r.tenant_id = $1 AND r.company_id = $2 AND r.code = $3 AND r.deleted_at IS NULL`

	var rl role.Role
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, code).Scan(
		&rl.ID, &rl.TenantID, &rl.CompanyID, &rl.Name, &rl.Code, &rl.Description,
		&rl.RoleType, &rl.IsSystem, &rl.CreatedAt, &rl.UpdatedAt, &rl.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, role.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get role by code: %w", err)
	}

	perms, err := r.loadPermissions(ctx, rl.ID)
	if err != nil {
		return nil, err
	}
	rl.Permissions = perms

	return &rl, nil
}

// List mengambil semua Role aktif dalam scope tertentu.
func (r *RoleRepository) List(ctx context.Context, s scope.Scope) ([]*role.Role, error) {
	const q = `
		SELECT r.id, r.tenant_id, r.company_id, r.name, r.code, r.description,
		       r.role_type, r.is_system, r.created_at, r.updated_at, r.deleted_at
		FROM roles r
		WHERE r.tenant_id = $1 AND r.company_id = $2 AND r.deleted_at IS NULL
		ORDER BY r.created_at ASC`

	rows, err := r.db.QueryContext(ctx, q, s.TenantID, s.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	var results []*role.Role
	for rows.Next() {
		var rl role.Role
		if err := rows.Scan(
			&rl.ID, &rl.TenantID, &rl.CompanyID, &rl.Name, &rl.Code, &rl.Description,
			&rl.RoleType, &rl.IsSystem, &rl.CreatedAt, &rl.UpdatedAt, &rl.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &rl)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load permissions for each role.
	for _, rl := range results {
		perms, err := r.loadPermissions(ctx, rl.ID)
		if err != nil {
			return nil, err
		}
		rl.Permissions = perms
	}

	return results, nil
}

// loadPermissions mengambil daftar permission untuk role tertentu dari role_permissions.
func (r *RoleRepository) loadPermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	const q = `SELECT permissions FROM role_permissions WHERE role_id = $1`

	var raw json.RawMessage
	err := r.db.QueryRowContext(ctx, q, roleID).Scan(&raw)
	if err == sql.ErrNoRows {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load permissions for role %s: %w", roleID, err)
	}

	var perms []string
	if err := json.Unmarshal(raw, &perms); err != nil {
		return []string{}, nil
	}
	return perms, nil
}

// ── UserRoleWriteRepository ───────────────────────────────────────────────────

// Assign menetapkan user ke role.
func (r *RoleRepository) Assign(ctx context.Context, ur *role.UserRole) error {
	const q = `
		INSERT INTO user_roles (id, user_id, role_id, tenant_id, company_id,
		                        teacher_id, guardian_id, is_active, assigned_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, q,
		ur.ID, ur.UserID, ur.RoleID, ur.TenantID, ur.CompanyID,
		ur.TeacherID, ur.GuardianID, ur.IsActive, ur.AssignedAt,
	)
	if err != nil {
		return fmt.Errorf("assign user role: %w", err)
	}
	return nil
}

// Revoke menghapus assignment role dari user dalam scope tertentu.
func (r *RoleRepository) Revoke(ctx context.Context, s scope.Scope, userID, roleID uuid.UUID) error {
	const q = `
		DELETE FROM user_roles
		WHERE tenant_id = $1 AND company_id = $2 AND user_id = $3 AND role_id = $4`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, userID, roleID)
	if err != nil {
		return fmt.Errorf("revoke user role: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return role.ErrUserRoleNotFound
	}
	return nil
}

// ── UserRoleReadRepository ────────────────────────────────────────────────────

// GetByUserAndCompany mengambil semua UserRole untuk user tertentu di company tertentu.
func (r *RoleRepository) GetByUserAndCompany(ctx context.Context, userID, companyID uuid.UUID) ([]*role.UserRole, error) {
	const q = `
		SELECT id, user_id, role_id, tenant_id, company_id,
		       teacher_id, guardian_id, is_active, assigned_at
		FROM user_roles
		WHERE user_id = $1 AND company_id = $2 AND is_active = true`

	rows, err := r.db.QueryContext(ctx, q, userID, companyID)
	if err != nil {
		return nil, fmt.Errorf("get user roles by user and company: %w", err)
	}
	defer rows.Close()

	var results []*role.UserRole
	for rows.Next() {
		var ur role.UserRole
		if err := rows.Scan(
			&ur.ID, &ur.UserID, &ur.RoleID, &ur.TenantID, &ur.CompanyID,
			&ur.TeacherID, &ur.GuardianID, &ur.IsActive, &ur.AssignedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &ur)
	}
	return results, rows.Err()
}

// GetRolesForUser mengambil daftar Role yang dimiliki user di company tertentu.
func (r *RoleRepository) GetRolesForUser(ctx context.Context, userID, companyID uuid.UUID) ([]*role.Role, error) {
	const q = `
		SELECT r.id, r.tenant_id, r.company_id, r.name, r.code, r.description,
		       r.role_type, r.is_system, r.created_at, r.updated_at, r.deleted_at
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND ur.company_id = $2 AND ur.is_active = true
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at ASC`

	rows, err := r.db.QueryContext(ctx, q, userID, companyID)
	if err != nil {
		return nil, fmt.Errorf("get roles for user: %w", err)
	}
	defer rows.Close()

	var results []*role.Role
	for rows.Next() {
		var rl role.Role
		if err := rows.Scan(
			&rl.ID, &rl.TenantID, &rl.CompanyID, &rl.Name, &rl.Code, &rl.Description,
			&rl.RoleType, &rl.IsSystem, &rl.CreatedAt, &rl.UpdatedAt, &rl.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &rl)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load permissions for each role.
	for _, rl := range results {
		perms, err := r.loadPermissions(ctx, rl.ID)
		if err != nil {
			return nil, err
		}
		rl.Permissions = perms
	}

	return results, nil
}
