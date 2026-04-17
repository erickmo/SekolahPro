package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/user"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// UserRepository adalah concrete implementation dari user.WriteRepository + user.ReadRepository.
// Users bersifat global (tidak di-scope per tenant/company), karena email harus unik secara sistem.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository membuat instance baru UserRepository.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Compile-time interface checks.
var (
	_ user.WriteRepository = (*UserRepository)(nil)
	_ user.ReadRepository  = (*UserRepository)(nil)
)

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity User baru ke database.
func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	const q = `
		INSERT INTO users (id, email, phone, password_hash, full_name, is_active,
		                   is_superadmin, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, q,
		u.ID, u.Email, u.Phone, u.PasswordHash, u.FullName,
		u.IsActive, u.IsSuperadmin, u.EmailVerified,
		u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

// Update mengupdate User yang sudah ada berdasarkan ID.
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	const q = `
		UPDATE users
		SET email = $2, phone = $3, full_name = $4, is_active = $5, updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		u.ID, u.Email, u.Phone, u.FullName, u.IsActive, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return user.ErrNotFound
	}
	return nil
}

// UpdatePassword mengupdate password hash untuk user tertentu.
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	const q = `UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, id, passwordHash, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return user.ErrNotFound
	}
	return nil
}

// UpdateLastLogin mengupdate last_login_at untuk user tertentu.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET last_login_at = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, id, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return user.ErrNotFound
	}
	return nil
}

// Deactivate menonaktifkan user (set is_active = false).
func (r *UserRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET is_active = false, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, id, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("deactivate user: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return user.ErrNotFound
	}
	return nil
}

// SoftDelete melakukan soft delete (mengisi deleted_at = now()).
func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET deleted_at = $2, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`

	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, q, id, now)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return user.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu User berdasarkan ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	const q = `
		SELECT id, email, phone, password_hash, full_name, is_active,
		       is_superadmin, email_verified, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	var u user.User
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.FullName, &u.IsActive,
		&u.IsSuperadmin, &u.EmailVerified, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

// GetByEmail mengambil satu User berdasarkan email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	const q = `
		SELECT id, email, phone, password_hash, full_name, is_active,
		       is_superadmin, email_verified, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	var u user.User
	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.FullName, &u.IsActive,
		&u.IsSuperadmin, &u.EmailVerified, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

// List mengambil daftar User dengan pagination.
// Meskipun User bersifat global, List tetap menerima scope untuk konsistensi API.
// Filter tambahan bisa diterapkan berdasarkan user_roles jika dibutuhkan.
func (r *UserRepository) List(ctx context.Context, s scope.Scope, limit, offset int) ([]*user.User, int, error) {
	total, err := r.countUsers(ctx, s)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectUsers(ctx, s, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *UserRepository) countUsers(ctx context.Context, s scope.Scope) (int, error) {
	const q = `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, q).Scan(&total); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return total, nil
}

func (r *UserRepository) selectUsers(ctx context.Context, s scope.Scope, limit, offset int) ([]*user.User, error) {
	const q = `
		SELECT id, email, phone, password_hash, full_name, is_active,
		       is_superadmin, email_verified, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	return scanUsers(rows)
}

func scanUsers(rows *sql.Rows) ([]*user.User, error) {
	var results []*user.User
	for rows.Next() {
		var u user.User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.FullName, &u.IsActive,
			&u.IsSuperadmin, &u.EmailVerified, &u.LastLoginAt,
			&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &u)
	}
	return results, rows.Err()
}
