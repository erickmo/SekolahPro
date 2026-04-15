// Package user adalah domain CQRS untuk users dan authentication (ADR-013).
//
// Users bukan Vernon — ini domain auth/infrastructure yang dibutuhkan semua domain.
// Menggunakan typed columns karena:
//   - Password hash TIDAK boleh di-expose via generic _data response.
//   - Unique email constraint di database level.
//   - Auth logic lebih kompleks dari CRUD.
//
// Aturan layer:
//   - Package ini TIDAK boleh import infrastructure packages.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package user

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// User adalah entity utama untuk akun login.
type User struct {
	ID            uuid.UUID  `db:"id"             json:"id"`
	Email         string     `db:"email"          json:"email"`
	Phone         string     `db:"phone"          json:"phone,omitempty"`
	PasswordHash  string     `db:"password_hash"  json:"-"`
	FullName      string     `db:"full_name"      json:"full_name"`
	IsActive      bool       `db:"is_active"      json:"is_active"`
	IsSuperadmin  bool       `db:"is_superadmin"  json:"is_superadmin"`
	EmailVerified bool       `db:"email_verified" json:"email_verified"`
	LastLoginAt   *time.Time `db:"last_login_at"  json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"     json:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"     json:"-"`
}

// WriteRepository mendefinisikan operasi write untuk users.
type WriteRepository interface {
	Save(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk users.
type ReadRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, s scope.Scope, limit, offset int) ([]*User, int, error)
}
