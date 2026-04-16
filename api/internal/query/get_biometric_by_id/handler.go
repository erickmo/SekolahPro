// Package get_biometric_by_id menangani query untuk mengambil BiometricEnrollment berdasarkan ID.
package get_biometric_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/biometric"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_biometric_by_id"

// Query berisi parameter untuk mengambil BiometricEnrollment berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID             string  `json:"id"`
	NasabahID      string  `json:"nasabah_id"`
	BiometricType  string  `json:"biometric_type"`
	DeviceInfo     string  `json:"device_info"`
	TemplateHash   string  `json:"template_hash"`
	IsActive       bool    `json:"is_active"`
	VerifiedAt     *string `json:"verified_at"`
	VerifiedBy     *string `json:"verified_by"`
	FailedAttempts int     `json:"failed_attempts"`
	LastAttemptAt  *string `json:"last_attempt_at"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// Handler menangani Query get_biometric_by_id.
type Handler struct {
	repo biometric.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo biometric.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil BiometricEnrollment berdasarkan ID dengan filter scope organisasi.
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
		return nil, fmt.Errorf("get biometric by id: %w", err)
	}

	result := &Result{
		ID:             entity.ID.String(),
		NasabahID:      entity.NasabahID.String(),
		BiometricType:  string(entity.BiometricType),
		DeviceInfo:     entity.DeviceInfo,
		TemplateHash:   entity.TemplateHash,
		IsActive:       entity.IsActive,
		FailedAttempts: entity.FailedAttempts,
		CreatedAt:      entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if entity.VerifiedAt != nil {
		v := entity.VerifiedAt.Format("2006-01-02T15:04:05Z")
		result.VerifiedAt = &v
	}
	if entity.VerifiedBy != nil {
		v := entity.VerifiedBy.String()
		result.VerifiedBy = &v
	}
	if entity.LastAttemptAt != nil {
		v := entity.LastAttemptAt.Format("2006-01-02T15:04:05Z")
		result.LastAttemptAt = &v
	}

	return result, nil
}
