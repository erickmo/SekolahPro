// Package role adalah domain CQRS untuk roles dan permissions (ADR-013).
//
// Roles di-scope per company — setiap company punya set roles sendiri.
// System roles di-seed saat company dibuat dan tidak bisa dihapus.
//
// Aturan layer:
//   - Package ini TIDAK boleh import infrastructure packages.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package role

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// Role adalah definisi role per company.
type Role struct {
	ID          uuid.UUID  `db:"id"          json:"id"`
	TenantID    uuid.UUID  `db:"tenant_id"   json:"tenant_id"`
	CompanyID   uuid.UUID  `db:"company_id"  json:"company_id"`
	Name        string     `db:"name"        json:"name"`
	Code        string     `db:"code"        json:"code"`
	Description string     `db:"description" json:"description,omitempty"`
	RoleType    string     `db:"role_type"   json:"role_type"`
	IsSystem    bool       `db:"is_system"   json:"is_system"`
	Permissions []string   `db:"-"           json:"permissions"`
	CreatedAt   time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"  json:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"  json:"-"`
}

// UserRole adalah assignment user ke role per company.
type UserRole struct {
	ID         uuid.UUID  `db:"id"          json:"id"`
	UserID     uuid.UUID  `db:"user_id"     json:"user_id"`
	RoleID     uuid.UUID  `db:"role_id"     json:"role_id"`
	TenantID   uuid.UUID  `db:"tenant_id"   json:"tenant_id"`
	CompanyID  uuid.UUID  `db:"company_id"  json:"company_id"`
	TeacherID  *uuid.UUID `db:"teacher_id"  json:"teacher_id,omitempty"`
	GuardianID *uuid.UUID `db:"guardian_id"  json:"guardian_id,omitempty"`
	IsActive   bool       `db:"is_active"   json:"is_active"`
	AssignedAt time.Time  `db:"assigned_at" json:"assigned_at"`
}

// Role type constants.
const (
	TypeAdmin          = "admin"
	TypePrincipal      = "principal"
	TypeVicePrincipal  = "vice_principal"
	TypeTeacher        = "teacher"
	TypeHomeroomTeacher = "homeroom_teacher"
	TypeCounselor      = "counselor"
	TypeFinance        = "finance"
	TypeStaff          = "staff"
	TypeParent         = "parent"
)

// System role codes — di-seed saat company dibuat.
const (
	CodeAdmin          = "admin"
	CodePrincipal      = "principal"
	CodeVicePrincipal  = "vice_principal"
	CodeHomeroomTeacher = "homeroom_teacher"
	CodeTeacher        = "teacher"
	CodeCounselor      = "counselor"
	CodeFinance        = "finance"
	CodeStaff          = "staff"
	CodeParent         = "parent"
)

// HasPermission memeriksa apakah role memiliki permission tertentu.
// Wildcard "*" berarti semua permission.
func (r *Role) HasPermission(perm string) bool {
	for _, p := range r.Permissions {
		if p == "*" || p == perm {
			return true
		}
	}
	return false
}

// RoleWriteRepository mendefinisikan operasi write untuk roles.
type RoleWriteRepository interface {
	Save(ctx context.Context, s scope.Scope, r *Role) error
	Update(ctx context.Context, s scope.Scope, r *Role) error
	SoftDelete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// RoleReadRepository mendefinisikan operasi read untuk roles.
type RoleReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*Role, error)
	GetByCode(ctx context.Context, s scope.Scope, code string) (*Role, error)
	List(ctx context.Context, s scope.Scope) ([]*Role, error)
}

// UserRoleWriteRepository mendefinisikan operasi write untuk user_roles.
type UserRoleWriteRepository interface {
	Assign(ctx context.Context, ur *UserRole) error
	Revoke(ctx context.Context, s scope.Scope, userID, roleID uuid.UUID) error
}

// UserRoleReadRepository mendefinisikan operasi read untuk user_roles.
type UserRoleReadRepository interface {
	GetByUserAndCompany(ctx context.Context, userID, companyID uuid.UUID) ([]*UserRole, error)
	GetRolesForUser(ctx context.Context, userID, companyID uuid.UUID) ([]*Role, error)
}
