// Package create_verification_log menangani command untuk membuat log verifikasi biometric.
package create_verification_log

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/biometric_log"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_verification_log"

// Command berisi data yang dibutuhkan untuk membuat log verifikasi biometric.
type Command struct {
	EnrollmentID       uuid.UUID
	NasabahID          uuid.UUID
	VerificationResult biometric_log.VerificationResult
	FallbackMethod     *biometric_log.FallbackMethod
	DeviceInfo         string
	IPAddress          string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_verification_log.
type Handler struct {
	repo biometric_log.WriteRepository
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo biometric_log.WriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle memvalidasi command, membuat entity, dan menyimpan ke repository.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.EnrollmentID == uuid.Nil {
		return biometric_log.ErrEnrollmentEmpty
	}
	if cmd.NasabahID == uuid.Nil {
		return biometric_log.ErrNasabahEmpty
	}
	if cmd.VerificationResult == "" {
		return biometric_log.ErrResultEmpty
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &biometric_log.BiometricVerificationLog{
		ID:                 id,
		EnrollmentID:       cmd.EnrollmentID,
		NasabahID:          cmd.NasabahID,
		VerificationResult: cmd.VerificationResult,
		FallbackMethod:     cmd.FallbackMethod,
		DeviceInfo:         cmd.DeviceInfo,
		IPAddress:          cmd.IPAddress,
		AttemptedAt:        now,
		CreatedAt:          now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save verification log: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
