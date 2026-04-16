// Package lifecycle mendefinisikan domain untuk Membership Lifecycle.
//
// Mengelola siklus keanggotaan koperasi: pendaftaran, verifikasi,
// pengaktifan, penangguhan, pengunduran diri, dan pencabutan.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package lifecycle

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// MembershipStatus mendefinisikan status keanggotaan.
type MembershipStatus string

const (
	MembershipStatusPending    MembershipStatus = "pending"
	MembershipStatusVerified   MembershipStatus = "verified"
	MembershipStatusActive     MembershipStatus = "active"
	MembershipStatusSuspended  MembershipStatus = "suspended"
	MembershipStatusResigned   MembershipStatus = "resigned"
	MembershipStatusRevoked    MembershipStatus = "revoked"
)

// TransitionType mendefinisikan jenis transisi status.
type TransitionType string

const (
	TransitionTypeApply        TransitionType = "apply"
	TransitionTypeVerify       TransitionType = "verify"
	TransitionTypeActivate     TransitionType = "activate"
	TransitionTypeSuspend      TransitionType = "suspend"
	TransitionTypeReactivate   TransitionType = "reactivate"
	TransitionTypeResign       TransitionType = "resign"
	TransitionTypeRevoke       TransitionType = "revoke"
)

// MembershipLifecycle adalah entity utama untuk siklus keanggotaan.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type MembershipLifecycle struct {
	ID              uuid.UUID       `db:"id"`
	NasabahID       uuid.UUID       `db:"nasabah_id"`
	CurrentStatus   MembershipStatus `db:"current_status"`
	PreviousStatus  *MembershipStatus `db:"previous_status"`
	LastTransition  TransitionType  `db:"last_transition"`
	TransitionReason string         `db:"transition_reason"`
	AppliedAt       time.Time       `db:"applied_at"`
	VerifiedAt      *time.Time      `db:"verified_at"`
	VerifiedBy      *uuid.UUID      `db:"verified_by"`
	ActivatedAt     *time.Time      `db:"activated_at"`
	SuspendedAt     *time.Time      `db:"suspended_at"`
	SuspendedBy     *uuid.UUID      `db:"suspended_by"`
	ResignedAt      *time.Time      `db:"resigned_at"`
	RevokedAt       *time.Time      `db:"revoked_at"`
	RevokedBy       *uuid.UUID      `db:"revoked_by"`
	MembershipNo    string          `db:"membership_no"`
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"`
	DeletedAt       *time.Time      `db:"deleted_at"`
}

// LifecycleFilter berisi parameter filter opsional untuk list query.
type LifecycleFilter struct {
	CurrentStatus  *MembershipStatus
	NasabahID      *uuid.UUID
	LastTransition *TransitionType
}

// WriteRepository mendefinisikan operasi write untuk domain lifecycle.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *MembershipLifecycle) error
	Update(ctx context.Context, s scope.Scope, e *MembershipLifecycle) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk domain lifecycle.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*MembershipLifecycle, error)
	List(ctx context.Context, s scope.Scope, filter LifecycleFilter, limit, offset int, sortBy, order string) ([]*MembershipLifecycle, int, error)
}
