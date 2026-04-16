// Package biometric_log mendefinisikan domain model untuk log verifikasi biometric.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package biometric_log

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// VerificationResult mendefinisikan hasil verifikasi.
type VerificationResult string

const (
	ResultSuccess  VerificationResult = "success"
	ResultFailed   VerificationResult = "failed"
	ResultFallback VerificationResult = "fallback"
)

// FallbackMethod mendefinisikan metode fallback yang digunakan.
type FallbackMethod string

const (
	FallbackPIN      FallbackMethod = "pin"
	FallbackPassword FallbackMethod = "password"
	FallbackManual   FallbackMethod = "manual"
)

// BiometricVerificationLog adalah entity untuk log verifikasi biometric.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type BiometricVerificationLog struct {
	ID                 uuid.UUID          `db:"id"`
	EnrollmentID       uuid.UUID          `db:"enrollment_id"`
	NasabahID          uuid.UUID          `db:"nasabah_id"`
	VerificationResult VerificationResult `db:"verification_result"`
	FallbackMethod     *FallbackMethod    `db:"fallback_method"`
	DeviceInfo         string             `db:"device_info"`
	IPAddress          string             `db:"ip_address"`
	AttemptedAt        time.Time          `db:"attempted_at"`
	CreatedAt          time.Time          `db:"created_at"`
	DeletedAt          *time.Time         `db:"deleted_at"`
}

// WriteRepository mendefinisikan operasi write untuk domain biometric_log.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *BiometricVerificationLog) error
}

// ReadRepository mendefinisikan operasi read untuk domain biometric_log.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*BiometricVerificationLog, error)
	List(ctx context.Context, s scope.Scope, filter LogFilter, limit, offset int, sortBy, order string) ([]*BiometricVerificationLog, int, error)
}

// LogFilter berisi parameter filter opsional untuk list query.
type LogFilter struct {
	EnrollmentID       *uuid.UUID
	NasabahID          *uuid.UUID
	VerificationResult *VerificationResult
}
