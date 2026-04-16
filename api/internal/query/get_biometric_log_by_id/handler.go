// Package get_biometric_log_by_id menangani query untuk mengambil BiometricVerificationLog berdasarkan ID.
package get_biometric_log_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/biometric_log"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_biometric_log_by_id"

// Query berisi parameter untuk mengambil BiometricVerificationLog berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                 string  `json:"id"`
	EnrollmentID       string  `json:"enrollment_id"`
	NasabahID          string  `json:"nasabah_id"`
	VerificationResult string  `json:"verification_result"`
	FallbackMethod     *string `json:"fallback_method"`
	DeviceInfo         string  `json:"device_info"`
	IPAddress          string  `json:"ip_address"`
	AttemptedAt        string  `json:"attempted_at"`
	CreatedAt          string  `json:"created_at"`
}

// Handler menangani Query get_biometric_log_by_id.
type Handler struct {
	repo biometric_log.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo biometric_log.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil BiometricVerificationLog berdasarkan ID dengan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	entity, err := h.repo.GetByID(ctx, s, qry.ID)
	if err != nil {
		return nil, fmt.Errorf("get biometric log by id: %w", err)
	}

	result := &Result{
		ID:                 entity.ID.String(),
		EnrollmentID:       entity.EnrollmentID.String(),
		NasabahID:          entity.NasabahID.String(),
		VerificationResult: string(entity.VerificationResult),
		DeviceInfo:         entity.DeviceInfo,
		IPAddress:          entity.IPAddress,
		AttemptedAt:        entity.AttemptedAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:          entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if entity.FallbackMethod != nil {
		v := string(*entity.FallbackMethod)
		result.FallbackMethod = &v
	}

	return result, nil
}
