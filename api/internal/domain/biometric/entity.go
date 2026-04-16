// Package biometric mendefinisikan domain model untuk biometric enrollment.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package biometric

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// BiometricType mendefinisikan tipe biometric yang didukung.
type BiometricType string

const (
	BiometricTypeFingerprint BiometricType = "fingerprint"
	BiometricTypeFace        BiometricType = "face"
	BiometricTypeVoice       BiometricType = "voice"
)

// BiometricEnrollment adalah entity utama untuk biometric enrollment.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type BiometricEnrollment struct {
	ID              uuid.UUID      `db:"id"`
	NasabahID       uuid.UUID      `db:"nasabah_id"`
	BiometricType   BiometricType  `db:"biometric_type"`
	DeviceInfo      string         `db:"device_info"`
	TemplateHash    string         `db:"template_hash"`
	IsActive        bool           `db:"is_active"`
	VerifiedAt      *time.Time     `db:"verified_at"`
	VerifiedBy      *uuid.UUID     `db:"verified_by"`
	FailedAttempts  int            `db:"failed_attempts"`
	LastAttemptAt   *time.Time     `db:"last_attempt_at"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
	DeletedAt       *time.Time     `db:"deleted_at"`
}

// WriteRepository mendefinisikan operasi write untuk domain biometric.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *BiometricEnrollment) error
	Update(ctx context.Context, s scope.Scope, e *BiometricEnrollment) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk domain biometric.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*BiometricEnrollment, error)
	List(ctx context.Context, s scope.Scope, filter BiometricFilter, limit, offset int, sortBy, order string) ([]*BiometricEnrollment, int, error)
}

// BiometricFilter berisi parameter filter opsional untuk list query.
type BiometricFilter struct {
	NasabahID     *uuid.UUID
	BiometricType *BiometricType
	IsActive      *bool
}
