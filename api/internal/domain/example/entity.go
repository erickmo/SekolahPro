// Package example adalah contoh domain module.
// Ganti "example" dengan nama domain yang sesuai.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package example

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// Example adalah entity utama untuk domain ini.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type Example struct {
	ID          uuid.UUID  `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	IsActive    bool       `db:"is_active"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

// WriteRepository mendefinisikan operasi write untuk domain ini.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *Example) error
	Update(ctx context.Context, s scope.Scope, e *Example) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk domain ini.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*Example, error)
	List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*Example, int, error)
}
